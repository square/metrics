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
	"github.com/square/metrics/tasks"
	"github.com/square/metrics/timeseries"
	"github.com/square/metrics/util"
)

// Blueflood is a timeseries storage API instance.
type Blueflood struct {
	config Config
}

//Blueflood implements TimeseriesStorageAPI
var _ timeseries.StorageAPI = (*Blueflood)(nil)

type TimeSource func() time.Time

// A Resolution stores information about supported resolutions and the timeranges in which they're available.
type Resolution struct {
	Name           string        `yaml:"name"`
	Resolution     time.Duration `yaml:"resolution"`
	FirstAvailable time.Duration `yaml:"first_available"`
	TimeToLive     time.Duration `yaml:"ttl"` // TimeToLive excludes time before availablility
}

func (r Resolution) String() string {
	return fmt.Sprintf("Name: %s Duration: %+v", r.Name, r.Resolution)
}

type Config struct {
	BaseURL                 string        `yaml:"base_url"`
	TenantID                string        `yaml:"tenant_id"`
	Resolutions             []Resolution  `yaml:"resolutions"`           // Resolutions are ordered by priority: best (typically finest) first.
	Timeout                 time.Duration `yaml:"timeout"`               // Timeout is the amount of time a single fetch request is allowed.
	MaxSimultaneousRequests int           `yaml:"simultaneous_requests"` // simultaneous requests limits the number of concurrent single-fetches for each multi-fetcj

	GraphiteMetricConverter util.GraphiteConverter

	HTTPClient httpClient
	TimeSource TimeSource
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

	b := &Blueflood{
		config: c,
	}
	// TODO: copy internal config structures to prevent modification?
	return b
}

// ChooseResolution will choose the finest-grained resolution for which an interval fetch plan exists that
// is at least as coarse as the lower bound.
func (b *Blueflood) ChooseResolution(requested api.Timerange, lowerBound time.Duration) (time.Duration, error) {
	now := b.config.TimeSource()
	for i, current := range b.config.Resolutions {
		if current.Resolution < lowerBound || current.Resolution < requested.Resolution() {
			continue
		}
		_, err := planFetchIntervals(b.config.Resolutions[:i+1], now, requested.Interval())
		if err == nil {
			return current.Resolution, nil
		}
	}
	return 0, fmt.Errorf("cannot choose resolution for timerange %+v; available resolutions do not live long enough or are not available soon enough.", requested)
}

// planFetchIntervals will plan the (point-count minimal) request intervals needed to cover the given timerange.
// the resolutions slice should be sorted, with the finest-grained resolution first.
func planFetchIntervals(resolutions []Resolution, now time.Time, requestInterval api.Interval) (map[Resolution]api.Interval, error) {
	answer := map[Resolution]api.Interval{}
	// Note: for anything other than FULL, a Blueflood returned point corresponds to the period FOLLOWING that point.
	// e.g. at 1hr resolution, a 4pm point summarizes all points in [4pm, 5pm], exclusive of 5pm.
	requestTimerange := requestInterval.CoveringTimerange(resolutions[len(resolutions)-1].Resolution)
	here := requestTimerange.Start()
	end := requestTimerange.End()
	for i := len(resolutions) - 1; i >= 0; i-- {
		resolution := resolutions[i]
		if here.Before(now.Add(-resolution.TimeToLive)) {
			// Expired
			continue
		}
		originalHere := here
		for {
			// TODO: optimize this into a division.
			if !here.Before(end) {
				break // We'll covered the timerange.
			}
			if here.Add(resolution.Resolution).After(now.Add(-resolution.FirstAvailable)) {
				break // Can't add this point- it's not available yet.
			}
			here = here.Add(resolution.Resolution)
		}
		if here != originalHere {
			// At least one point is included, so:
			answer[resolution] = api.Interval{Start: originalHere, End: here}
		}
	}
	if here.Before(end) {
		return answer, fmt.Errorf("can't reach end of timerange using available resolutions up to %+v: it expires after only %+v try using a coarser resolution", resolutions[len(resolutions)-1].Resolution, end.Sub(here))
	}
	return answer, nil
}

// planFetchIntervalsWithOnlyFiner assumes that the requested range is as coarse as desired.
// Hence, it will trim all coarser resolutions before doing planning.
func planFetchIntervalsWithOnlyFiner(resolutions []Resolution, now time.Time, requestRange api.Timerange) (map[Resolution]api.Interval, error) {
	for i := range resolutions {
		if resolutions[i].Resolution > requestRange.Resolution() {
			if i == 0 {
				return nil, fmt.Errorf("No resolutions are available at least as fine as the chosen %+v", requestRange.Resolution())
			}
			return planFetchIntervals(resolutions[:i], now, requestRange.Interval())
		}
	}
	return planFetchIntervals(resolutions, now, requestRange.Interval())
}

// FetchSingleTimeseries fetches a timeseries with the given tagged metric.
// It requires that the resolution is supported.
func (b *Blueflood) FetchSingleTimeseries(request timeseries.FetchRequest) (api.Timeseries, error) {
	defer request.Profiler.RecordWithDescription("Blueflood FetchSingleTimeseries", request.Metric.String())()
	sampler, ok := samplerMap[request.SampleMethod]
	if !ok {
		return api.Timeseries{}, fmt.Errorf("unsupported SampleMethod %s", request.SampleMethod.String())
	}
	// Extend it one point forward, unless that would fetch past the current time.
	modifiedRange := request.Timerange
	if modifiedRange.End().Add(modifiedRange.Resolution()).Before(b.config.TimeSource()) {
		modifiedRange = modifiedRange.ExtendAfter(modifiedRange.Resolution())
	}
	intervals, err := planFetchIntervalsWithOnlyFiner(b.config.Resolutions, b.config.TimeSource(), modifiedRange)
	if err != nil {
		return api.Timeseries{}, err
	}

	queue := tasks.NewParallelQueue(len(intervals), b.config.Timeout)

	allPoints := []metricPoint{}

	for resolution, interval := range intervals {
		resolution, interval := resolution, interval
		var points []metricPoint
		queue.Do(func() error {
			defer request.Profiler.RecordWithDescription("Blueflood FetchSingleTimeseries Resolution", fmt.Sprintf("%s at %+v", request.Metric.String(), resolution.Resolution))
			points, err = b.requestPoints(request.Metric, interval, sampler, resolution)
			if err != nil {
				return err
			}
			queue.Lock()
			defer queue.Unlock()
			allPoints = append(allPoints, points...)
			return nil
		})
	}

	if err := queue.Wait(); err != nil {
		return api.Timeseries{}, err
	}

	values := samplePoints(allPoints, request.Timerange, sampler)

	return api.Timeseries{
		Values: values,
		TagSet: request.Metric.TagSet,
	}, nil
}

func (b *Blueflood) requestPoints(metric api.TaggedMetric, interval api.Interval, sampler sampler, resolution Resolution) ([]metricPoint, error) {
	queryURL, err := b.constructURL(metric, interval, sampler, resolution)
	if err != nil {
		return nil, err
	}
	parsedResult, err := b.fetch(queryURL)
	if err != nil {
		return nil, err
	}
	return parsedResult.Values, nil
}

// Helper functions
// ----------------

// constructURL creates the URL to the blueflood's backend to fetch the data from.
func (b *Blueflood) constructURL(metric api.TaggedMetric, interval api.Interval, sampler sampler, resolution Resolution) (*url.URL, error) {
	graphiteName, err := b.config.GraphiteMetricConverter.ToGraphiteName(metric)
	if err != nil {
		return nil, timeseries.Error{metric, timeseries.InvalidSeriesError, "cannot convert to graphite name"}
	}

	result, err := url.Parse(fmt.Sprintf("%s/v2.0/%s/views/%s", b.config.BaseURL, b.config.TenantID, graphiteName))
	if err != nil {
		return nil, timeseries.Error{metric, timeseries.InvalidSeriesError, fmt.Sprintf("cannot generate URL for tagged metric with graphite name %s", graphiteName)}
	}

	result.RawQuery = url.Values{
		"from":       {strconv.FormatInt(int64(interval.Start.UnixNano()/1e6), 10)},
		"to":         {strconv.FormatInt(int64(interval.End.UnixNano()/1e6-1), 10)},
		"resolution": {resolution.Name},
		"select":     {fmt.Sprintf("numPoints,%s", strings.ToLower(sampler.fieldName))},
	}.Encode()

	return result, nil
}

// performFetch is a synchronous method that fetches the given URL.
func (b *Blueflood) performFetch(queryURL *url.URL) (queryResponse, error) {
	resp, err := b.config.HTTPClient.Get(queryURL.String())
	if err != nil {
		// TODO: report the right metric
		return queryResponse{}, timeseries.Error{api.TaggedMetric{}, timeseries.FetchIOError, "error while fetching - http connection"}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// TODO: report the right metric
		return queryResponse{}, timeseries.Error{api.TaggedMetric{}, timeseries.FetchIOError, "error while fetching - reading"}
	}
	var parsedJSON queryResponse
	err = json.Unmarshal(body, &parsedJSON)
	if err != nil {
		// TODO: report the right metric
		return queryResponse{}, timeseries.Error{api.TaggedMetric{}, timeseries.FetchIOError, "error while fetching - json decoding\nBody: " + string(body) + "\nError: " + err.Error() + "\nURL: " + queryURL.String()}
	}
	return parsedJSON, nil
}

// fetch fetches from the backend, asynchronously calling performFetch and cancelling on timeout.
func (b *Blueflood) fetch(queryURL *url.URL) (queryResponse, error) {
	type Answer struct {
		response queryResponse
		err      error
	}
	answer := make(chan Answer, 1)
	go func() {
		response, err := b.performFetch(queryURL)
		answer <- Answer{response, err}
	}()
	select {
	case result := <-answer:
		return result.response, result.err
	case <-time.After(b.config.Timeout):
		// TODO: report the right metric
		return queryResponse{}, timeseries.Error{api.TaggedMetric{}, timeseries.FetchTimeoutError, ""}
	}
}

type sampler struct {
	fieldName    string
	selectField  func(point metricPoint) float64
	sampleBucket func([]float64) float64
}

// sampleResult samples the points into a uniform slice of float64s.
func samplePoints(points []metricPoint, timerange api.Timerange, sampler sampler) []float64 {
	// A bucket holds a set of points corresponding to one interval in the result.
	buckets := make([][]float64, timerange.Slots())
	for _, point := range points {
		pointValue := sampler.selectField(point)
		index := (point.Timestamp - timerange.StartMillis()) / timerange.ResolutionMillis()
		if index < 0 || int(index) >= len(buckets) {
			continue
		}
		buckets[index] = append(buckets[index], pointValue)
	}

	// values will hold the final values to be returned as the series.
	values := make([]float64, timerange.Slots())

	for i, bucket := range buckets {
		if len(bucket) == 0 {
			values[i] = math.NaN()
			continue
		}
		values[i] = sampler.sampleBucket(bucket)
	}
	return values
}

var samplerMap = map[timeseries.SampleMethod]sampler{
	timeseries.SampleMean: {
		fieldName:   "average",
		selectField: func(point metricPoint) float64 { return point.Average },
		sampleBucket: func(bucket []float64) float64 {
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
		fieldName:   "min",
		selectField: func(point metricPoint) float64 { return point.Min },
		sampleBucket: func(bucket []float64) float64 {
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
		fieldName:   "max",
		selectField: func(point metricPoint) float64 { return point.Max },
		sampleBucket: func(bucket []float64) float64 {
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

func (b *Blueflood) FetchMultipleTimeseries(request timeseries.FetchMultipleRequest) (api.SeriesList, error) {
	defer request.Profiler.Record("Blueflood FetchMultipleTimeseries")()

	singleRequests := request.ToSingle()
	results := make([]api.Timeseries, len(singleRequests))
	errs := make([]error, len(singleRequests))
	queue := tasks.NewParallelQueue(b.config.MaxSimultaneousRequests, b.config.Timeout)
	for i := range singleRequests {
		i := i // Captures it in a new local for the closure.
		queue.Do(func() error {
			result, err := b.FetchSingleTimeseries(singleRequests[i])
			if err != nil {
				return err
			}
			results[i] = result
			return nil
		})
	}

	err := queue.Wait()
	if err != nil {
		return api.SeriesList{}, err
	}

	for _, err := range errs {
		if err != nil {
			return api.SeriesList{}, err
		}
	}

	return api.SeriesList{
		Series: results,
	}, nil
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
