// Copyright 2015 - 2016 Square Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package blueflood

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/log"
	"github.com/square/metrics/tasks"
	"github.com/square/metrics/timeseries"
	"github.com/square/metrics/util"
)

// Blueflood is a timeseries storage API instance.
type Blueflood struct {
	config Config
}

//Implements TimeseriesStorageAPI
var _ timeseries.StorageAPI = (*Blueflood)(nil)

type TimeSource func() time.Time

type Config struct {
	BaseURL                 string           `yaml:"base_url"`
	TenantID                string           `yaml:"tenant_id"`
	Ttls                    map[string]int64 `yaml:"ttls"` // Ttl in days
	Timeout                 time.Duration    `yaml:"timeout"`
	FullResolutionOverlap   int64            `yaml:"full_resolution_overlap"` // overlap to draw full resolution in seconds
	GraphiteMetricConverter util.GraphiteConverter
	HTTPClient              httpClient
	TimeSource              TimeSource
	MaxSimultaneousRequests int `yaml:"simultaneous_requests"`
}

type BluefloodParallelRequest struct {
	limit   int
	tickets chan struct{}
}

type httpClient interface {
	// our own client to mock out the standard golang HTTP Client.
	Get(url string) (resp *http.Response, err error)
}

type queryResponse struct {
	Values []metricPoint `json:"values"`
}

type metricPoint struct {
	Points    int     `json:"numPoints"`
	Timestamp int64   `json:"timestamp"`
	Average   float64 `json:"average"`
	Max       float64 `json:"max"`
	Min       float64 `json:"min"`
	Variance  float64 `json:"variance"`
}

type Resolution struct {
	bluefloodEnum string
	duration      time.Duration
}

func (r Resolution) String() string {
	return fmt.Sprintf("Name: %s Duration: %d", r.bluefloodEnum, r.duration/time.Minute)
}

var (
	ResolutionFull    = Resolution{"FULL", time.Second * 30}
	Resolution5Min    = Resolution{"MIN5", time.Minute * 5}
	Resolution20Min   = Resolution{"MIN20", time.Minute * 20}
	Resolution60Min   = Resolution{"MIN60", time.Minute * 60}
	Resolution240Min  = Resolution{"MIN240", time.Minute * 240}
	Resolution1440Min = Resolution{"MIN1440", time.Minute * 1440}
)
var Resolutions = []Resolution{
	ResolutionFull,
	Resolution5Min,
	Resolution20Min,
	Resolution60Min,
	Resolution240Min,
	Resolution1440Min,
}

func NewBlueflood(c Config) timeseries.StorageAPI {
	if c.HTTPClient == nil {
		c.HTTPClient = http.DefaultClient
	}
	if c.TimeSource == nil {
		c.TimeSource = time.Now
	}
	if c.MaxSimultaneousRequests == 0 {
		c.MaxSimultaneousRequests = 5
	}

	b := Blueflood{
		config: c,
	}
	b.config.Ttls = map[string]int64{}
	for k, v := range c.Ttls {
		b.config.Ttls[k] = v
	}
	return &b
}

type sampler struct {
	fieldName     string
	fieldSelector func(point metricPoint) float64
	bucketSampler func([]float64) float64
}

func (b *Blueflood) fetchLazy(timeout *tasks.Timeout, result *api.Timeseries, work func() (api.Timeseries, error), channel chan error, ctx BluefloodParallelRequest) {
	go func() {
		select {
		case ticket := <-ctx.tickets:
			series, err := work()
			// Put the ticket back (regardless of whether caller drops)
			ctx.tickets <- ticket
			// Store the result
			*result = series
			// Return the error (and sync up with the caller).
			channel <- err
		case <-timeout.Done():
			channel <- timeseries.Error{Code: timeseries.FetchTimeoutError}
		}
	}()
}

func (b *Blueflood) fetchManyLazy(timeout *tasks.Timeout, works []func() (api.Timeseries, error)) ([]api.Timeseries, error) {
	results := make([]api.Timeseries, len(works))
	channel := make(chan error, len(works)) // Buffering the channel means the goroutines won't need to wait.

	limit := b.config.MaxSimultaneousRequests
	tickets := make(chan struct{}, limit)
	for i := 0; i < limit; i++ {
		tickets <- struct{}{}
	}
	ctx := BluefloodParallelRequest{
		tickets: tickets,
	}
	for i := range results {
		b.fetchLazy(timeout, &results[i], works[i], channel, ctx)
	}

	var err error
	for range works {
		select {
		case thisErr := <-channel:
			if thisErr != nil {
				err = thisErr
			}
		case <-timeout.Done():
			return nil, timeseries.Error{Code: timeseries.FetchTimeoutError}
		}
	}
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (b *Blueflood) FetchMultipleTimeseries(request timeseries.FetchMultipleRequest) (api.SeriesList, error) {
	defer request.Profiler.Record("Blueflood FetchMultipleTimeseries")()
	works := make([]func() (api.Timeseries, error), len(request.Metrics))

	singleRequests := request.ToSingle()
	for i := range singleRequests {
		//Make sure we close around the specific individual request
		//and not the ranged/copied request.
		singleRequest := singleRequests[i]
		works[i] = func() (api.Timeseries, error) {
			return b.FetchSingleTimeseries(singleRequest)
		}
	}

	resultSeries, err := b.fetchManyLazy(request.Timeout, works)
	if err != nil {
		return api.SeriesList{}, err
	}

	return api.SeriesList{
		Series: resultSeries,
	}, nil
}

func (b *Blueflood) FetchSingleTimeseries(request timeseries.FetchRequest) (api.Timeseries, error) {
	defer request.Profiler.RecordWithDescription("Blueflood FetchSingleTimeseries", request.Metric.String())()
	sampler, ok := samplerMap[request.SampleMethod]
	if !ok {
		return api.Timeseries{}, fmt.Errorf("unsupported SampleMethod %s", request.SampleMethod.String())
	}
	queryResolution := b.config.bluefloodResolution(
		request.Timerange.Resolution(),
		request.Timerange.StartMillis(),
	)
	log.Debugf("Blueflood resolution: %s\n", queryResolution.String())

	// Sample the data at the given `queryResolution`
	queryURL, err := b.constructURL(request, sampler, queryResolution)
	if err != nil {
		return api.Timeseries{}, err
	}

	rawResults := make([][]byte, 1)
	parsedResult, rawResult, err := b.fetch(request, queryURL)
	rawResults[0] = rawResult
	if err != nil {
		return api.Timeseries{}, err
	}

	// combinedResult contains the requested data, along with higher-resolution data intended to fill in gaps.
	combinedResult := parsedResult.Values

	// Sample the data at the FULL resolution.
	// We clip the timerange so that it's only #{config.FullResolutionOverlap} seconds long.
	// This limits the amount of data to be fetched.
	fullResolutionParsedResult := func() []metricPoint {
		// If an error occurs, we just return nothing. We don't return the error.
		// This is so that errors while fetching the FULL-resolution data don't impact the requested data.
		fullResolutionRequest := request // Copy the request
		if request.Timerange.EndMillis()-request.Timerange.StartMillis() > b.config.FullResolutionOverlap*1000 {
			// Clip the timerange
			newTimerange, err := api.NewSnappedTimerange(request.Timerange.EndMillis()-b.config.FullResolutionOverlap*1000, request.Timerange.EndMillis(), request.Timerange.ResolutionMillis())
			if err != nil {
				log.Infof("FULL resolution data errored while building timerange: %s", err.Error())
				return nil
			}
			fullResolutionRequest.Timerange = newTimerange
		}
		fullResolutionQueryURL, err := b.constructURL(fullResolutionRequest, sampler, ResolutionFull)
		if err != nil {
			log.Infof("FULL resolution data errored while building url: %s", err.Error())
			return nil
		}
		fullResolutionParsedResult, rawResult, err := b.fetch(request, fullResolutionQueryURL)
		rawResults = append(rawResults, rawResult)
		if err != nil {
			log.Infof("FULL resolution data errored while parsing result: %s", err.Error())
			return nil
		}
		// The higher-resolution data will likely overlap with the requested data.
		// This isn't a problem - the requested, higher-resolution data will be downsampled by this code.
		// This downsampling should arrive at the same answer as Blueflood's built-in rollups.
		return fullResolutionParsedResult.Values
	}()

	combinedResult = append(combinedResult, fullResolutionParsedResult...)

	values := processResult(combinedResult, request.Timerange, sampler, queryResolution)
	log.Debugf("Constructed timeseries from result: %v", values)

	if request.UserSpecifiableConfig.IncludeRawData {
		return api.Timeseries{
			Values: values,
			TagSet: request.Metric.TagSet,
			Raw:    rawResults,
		}, nil
	} else {
		return api.Timeseries{
			Values: values,
			TagSet: request.Metric.TagSet,
		}, nil
	}
}

// CheckHealthy checks if the blueflood server is available by querying /v2.0
func (b *Blueflood) CheckHealthy() error {
	resp, err := b.config.HTTPClient.Get(fmt.Sprintf("%s/v2.0", b.config.BaseURL))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("Blueflood returned an unhealthy status of %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Helper functions
// ----------------

// constructURL creates the URL to the blueflood's backend to fetch the data from.
func (b *Blueflood) constructURL(
	request timeseries.FetchRequest,
	sampler sampler,
	queryResolution Resolution,
) (*url.URL, error) {
	graphiteName, err := b.config.GraphiteMetricConverter.ToGraphiteName(request.Metric)
	if err != nil {
		return nil, timeseries.Error{request.Metric, timeseries.InvalidSeriesError, "cannot convert to graphite name"}
	}

	result, err := url.Parse(fmt.Sprintf("%s/v2.0/%s/views/%s", b.config.BaseURL, b.config.TenantID, graphiteName))
	if err != nil {
		return nil, timeseries.Error{request.Metric, timeseries.InvalidSeriesError, "cannot generate URL"}
	}

	params := url.Values{}
	params.Set("from", strconv.FormatInt(request.Timerange.StartMillis(), 10))
	// Pull a bit outside of the requested range from blueflood so we
	// have enough data to generate all snapped values
	params.Set("to", strconv.FormatInt(request.Timerange.EndMillis()+request.Timerange.ResolutionMillis(), 10))
	params.Set("resolution", queryResolution.bluefloodEnum)
	params.Set("select", fmt.Sprintf("numPoints,%s", strings.ToLower(sampler.fieldName)))
	result.RawQuery = params.Encode()
	return result, nil
}

// fetches from the backend. on error, it returns an instance of timeseries.Error
func (b *Blueflood) fetch(request timeseries.FetchRequest, queryURL *url.URL) (queryResponse, []byte, error) {
	log.Debugf("Blueflood fetch: %s", queryURL.String())
	success := make(chan queryResponse, 1)
	failure := make(chan error, 1)
	timeout := time.After(b.config.Timeout)
	var rawResponse []byte
	go func() {
		resp, err := b.config.HTTPClient.Get(queryURL.String())
		if err != nil {
			failure <- timeseries.Error{request.Metric, timeseries.FetchIOError, "error while fetching - http connection"}
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		rawResponse = body
		if err != nil {
			failure <- timeseries.Error{request.Metric, timeseries.FetchIOError, "error while fetching - reading"}
			return
		}

		log.Debugf("Fetch result: %s", string(body))

		var parsedJSON queryResponse
		err = json.Unmarshal(body, &parsedJSON)
		// Construct a Timeseries from the result:
		if err != nil {
			failure <- timeseries.Error{request.Metric, timeseries.FetchIOError, "error while fetching - json decoding\nBody: " + string(body) + "\nError: " + err.Error() + "\nURL: " + queryURL.String()}
			return
		}
		success <- parsedJSON
	}()
	select {
	case response := <-success:
		return response, rawResponse, nil
	case err := <-failure:
		return queryResponse{}, rawResponse, err
	case <-timeout:
		return queryResponse{}, rawResponse, timeseries.Error{request.Metric, timeseries.FetchTimeoutError, ""}
	}
}

func processResult(
	points []metricPoint,
	timerange api.Timerange,
	sampler sampler,
	queryResolution Resolution) []float64 {
	// buckets are each filled with from the points stored in `points`, according to their timestamps.
	buckets := bucketsFromMetricPoints(points, sampler.fieldSelector, timerange)

	// values will hold the final values to be returned as the series.
	values := make([]float64, timerange.Slots())

	for i, bucket := range buckets {
		if len(bucket) == 0 {
			values[i] = math.NaN()
			continue
		}
		values[i] = sampler.bucketSampler(bucket)
	}

	// interpolate.
	return values
}

func addMetricPoint(metricPoint metricPoint, field func(metricPoint) float64, timerange api.Timerange, buckets [][]float64) bool {
	value := field(metricPoint)
	// The index to assign within the array is computed using the timestamp.
	// It floors to the nearest index.
	index := (metricPoint.Timestamp - timerange.StartMillis()) / timerange.ResolutionMillis()
	if index < 0 || index >= int64(timerange.Slots()) {
		return false
	}
	buckets[index] = append(buckets[index], value)
	return true
}

func bucketsFromMetricPoints(metricPoints []metricPoint, resultField func(metricPoint) float64, timerange api.Timerange) [][]float64 {
	buckets := make([][]float64, timerange.Slots())
	for _, point := range metricPoints {
		addMetricPoint(point, resultField, timerange, buckets)
	}
	return buckets
}

var samplerMap = map[timeseries.SampleMethod]sampler{
	timeseries.SampleMean: {
		fieldName:     "average",
		fieldSelector: func(point metricPoint) float64 { return point.Average },
		bucketSampler: func(bucket []float64) float64 {
			value := 0.0
			count := 0
			for _, v := range bucket {
				if !math.IsNaN(v) {
					value += v
					count++
				}
			}
			return value / float64(count)
		},
	},
	timeseries.SampleMin: {
		fieldName:     "min",
		fieldSelector: func(point metricPoint) float64 { return point.Min },
		bucketSampler: func(bucket []float64) float64 {
			smallest := math.NaN()
			for _, v := range bucket {
				if math.IsNaN(v) {
					continue
				}
				if math.IsNaN(smallest) {
					smallest = v
				} else {
					smallest = math.Min(smallest, v)
				}
			}
			return smallest
		},
	},
	timeseries.SampleMax: {
		fieldName:     "max",
		fieldSelector: func(point metricPoint) float64 { return point.Max },
		bucketSampler: func(bucket []float64) float64 {
			largest := math.NaN()
			for _, v := range bucket {
				if math.IsNaN(v) {
					continue
				}
				if math.IsNaN(largest) {
					largest = v
				} else {
					largest = math.Max(largest, v)
				}
			}
			return largest
		},
	},
}

// ChooseResolution will use the finest-grained resolution which doesn't exceed the slot limit.
// Thus, if you request too many points, it will automatically reduce the resolution.
func (b *Blueflood) ChooseResolution(requested api.Timerange, smallestResolution time.Duration) time.Duration {
	// In some cases, coarser-resolution data may have a shorter TTL.
	// To accommodate these cases, it must be verified that the requested timerange will
	// actually be present for the chosen resolution.
	// TODO: figure out how to make this work with moving averages and timeshifts

	requiredAge := b.config.TimeSource().Sub(requested.Start())

	for _, resolution := range Resolutions {
		survivesFor := b.config.oldestViableDataForResolution(resolution)
		if survivesFor < requiredAge {
			// The data probably won't be around for the earliest part of the timerange,
			// so don't use this resolution
			continue
		}
		if resolution.duration < requested.Resolution() {
			// Skip this timerange, it is finer than the one requested.
			continue
		}
		// Check that the timerange is large enough
		if resolution.duration >= smallestResolution {
			return resolution.duration
		}
	}
	// Leave it alone, since a better one can't be found
	return requested.Resolution()
}

// Blueflood keys the resolution param to a java enum, so we have to convert
// between them.
//
func (c Config) bluefloodResolution(
	desiredResolution time.Duration,
	startMs int64) Resolution {
	log.Debugf("Desired resolution in minutes: %d\n", desiredResolution/time.Minute)
	now := c.TimeSource().Unix() * 1000 //Milliseconds
	// Choose the appropriate resolution based on TTL, fetching the highest resolution data we can
	//
	age := time.Duration(now-startMs) * time.Millisecond //Age in milliseconds
	log.Debugf("The age in minutes of the start time %d\n", age/time.Minute)

	for _, current := range Resolutions {
		maxAge := c.oldestViableDataForResolution(current)
		// log.Debugf("Oldest age? %+v\n", maxAge)

		log.Debugf("Considering resolution %v\n", current)
		log.Debugf("Oldest conceivable data is %v\n", maxAge)
		log.Debugf("Is the desired resolution less than or equal to the current? %b\n", desiredResolution <= current.duration)
		log.Debugf("Is the start time within the TTL window? %b\n", age < maxAge)

		// If the desired resolution is less than or equal to the
		// current resolution and is the distance from now to the oldest
		// viable data still available?
		if desiredResolution <= current.duration &&
			age < maxAge {
			log.Debugf("Choosing resolution: %v\n", current)
			return current
		}
	}
	// If none of the above matched, we choose the coarsest
	return Resolutions[len(Resolutions)-1]
}

// Given a particular resolution, what's the duration of the oldest
// data that could still be available. For instance, if the resolution
// is Resolution20Min and the ttl is 60, then the data is available for
// 60 days.
func (c Config) oldestViableDataForResolution(r Resolution) time.Duration {
	var ttl int64
	if v, ok := c.Ttls[r.bluefloodEnum]; ok {
		ttl = v
	} else {
		// log.Debugf("Using blueflood default TTLs\n")
		// Use blueflood defaults
		switch r {
		case ResolutionFull:
			ttl = 1
		case Resolution5Min:
			ttl = 30
		case Resolution20Min:
			ttl = 60
		case Resolution60Min:
			ttl = 90
		case Resolution240Min:
			ttl = 180
		case Resolution1440Min:
			ttl = 365
		default:
			// Not a supported resolution by blueflood. No real way to recover if
			// someone's trying to fetch ttl for an invalid resolution.
			panic(fmt.Sprintf("invalid resolution `%s`", r))
		}
	}

	return time.Duration(ttl) * 24 * time.Hour
}
