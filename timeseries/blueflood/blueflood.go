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
	"sync"
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
// is at least as coarse as the requested timerange and lower bound.
func (b *Blueflood) ChooseResolution(requested api.Timerange, lowerBound time.Duration) (time.Duration, error) {
	now := b.config.TimeSource()
	lastErr := fmt.Errorf("no available resolutions to choose from")
	for i, current := range b.config.Resolutions {
		if current.Resolution < lowerBound || current.Resolution < requested.Resolution() {
			continue
		}
		_, err := planFetchIntervals(b.config.Resolutions[:i+1], now, requested)
		if err == nil {
			return current.Resolution, nil
		}
		lastErr = fmt.Errorf("cannot choose resolution for timerange %+v: %s", requested, err.Error())
	}
	return 0, lastErr
}

// planFetchIntervals will plan the (point-count minimal) request intervals needed to cover the given timerange.
// the resolutions slice should be sorted, with the finest-grained resolution first.
func planFetchIntervals(resolutions []Resolution, now time.Time, requestRange api.Timerange) (map[Resolution]api.Timerange, error) {
	answer := map[Resolution]api.Timerange{}

	cutTime := requestRange.Start().Add(-1 * time.Millisecond)
	for i := len(resolutions) - 1; i >= 0; i-- {
		if cutTime.After(requestRange.End()) || cutTime == requestRange.End() {
			// Don't need to fetch any more points.
			break
		}
		resolution := resolutions[i]
		if cutTime.Before(now.Add(-resolution.TimeToLive)) {
			// This resolution data doesn't live long enough to fetch the beginning of the timerange.
			continue
		}
		clippedAfter, ok := requestRange.Resample(resolution.Resolution).OnlyAfterExclusive(cutTime)
		if !ok {
			// This shouldn't ever be able to happen (provided that resolutions are in a sensible order).
			// It would mean that a coarser resolution first appears before some finer resolution.
			// However, this ordering is contingent upon resolutions being ordered by coarseness
			// (with finest first).
			continue
		}
		// Cut the timerange to the point where it's valid.
		nextCutTime := now.Add(-resolution.FirstAvailable)
		clippedBefore, ok := clippedAfter.Resample(resolution.Resolution).OnlyBeforeInclusive(nextCutTime)
		if !ok {
			// This resolution data expires much, much sooner than is useful.
			// If this occurs then (provided that the resolutions have a sane ordering),
			// it's very likely that the rest of the evaluation will fail.
			continue
		}
		cutTime = nextCutTime

		answer[resolution] = clippedBefore
	}
	if cutTime.After(requestRange.End()) || cutTime == requestRange.End() {
		return answer, nil
	}
	return nil, fmt.Errorf("Cannot cover timerange %+v with available resolution data", requestRange)
}

// planFetchIntervalsRestricted assumes that the requested range is as coarse as desired.
// Hence, it will trim all coarser resolutions before doing planning.
func planFetchIntervalsRestricted(resolutions []Resolution, now time.Time, requestRange api.Timerange) (map[Resolution]api.Timerange, error) {
	for i := range resolutions {
		if resolutions[i].Resolution > requestRange.Resolution() {
			return planFetchIntervals(resolutions[:i], now, requestRange)
		}
	}
	return planFetchIntervals(resolutions, now, requestRange)
}

// FetchSingleTimeseries fetches a timeseries with the given tagged metric.
// It requires that the resolution is supported.
func (b *Blueflood) FetchSingleTimeseries(request timeseries.FetchRequest) (api.Timeseries, error) {
	defer request.Profiler.RecordWithDescription("Blueflood FetchSingleTimeseries", request.Metric.String())()
	sampler, ok := samplerMap[request.SampleMethod]
	if !ok {
		return api.Timeseries{}, fmt.Errorf("unsupported SampleMethod %s", request.SampleMethod.String())
	}
	intervals, err := planFetchIntervalsRestricted(b.config.Resolutions, b.config.TimeSource(), request.Timerange)
	if err != nil {
		return api.Timeseries{}, err
	}

	allPoints := []metricPoint{}
	mutex := sync.Mutex{}
	wait := sync.WaitGroup{}
	someErr := error(nil)

	for resolution, timerange := range intervals {
		wait.Add(1)
		go func() {
			defer wait.Done()
			points, err := b.requestPoints(request.Metric, timerange, sampler, resolution)
			if err != nil {
				mutex.Lock()
				defer mutex.Unlock()
				someErr = err
				return
			}
			mutex.Lock()
			defer mutex.Unlock()
			allPoints = append(allPoints, points...)
		}()
	}
	wait.Wait()

	if someErr != nil {
		return api.Timeseries{}, someErr
	}

	values := samplePoints(allPoints, request.Timerange, sampler)

	return api.Timeseries{
		Values: values,
		TagSet: request.Metric.TagSet,
	}, nil
}

func (b *Blueflood) requestPoints(metric api.TaggedMetric, timerange api.Timerange, sampler sampler, resolution Resolution) ([]metricPoint, error) {
	queryURL, err := b.constructURL(metric, timerange, sampler, resolution)
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
func (b *Blueflood) constructURL(metric api.TaggedMetric, timerange api.Timerange, sampler sampler, resolution Resolution) (*url.URL, error) {
	graphiteName, err := b.config.GraphiteMetricConverter.ToGraphiteName(metric)
	if err != nil {
		return nil, timeseries.Error{metric, timeseries.InvalidSeriesError, "cannot convert to graphite name"}
	}

	result, err := url.Parse(fmt.Sprintf("%s/v2.0/%s/views/%s", b.config.BaseURL, b.config.TenantID, graphiteName))
	if err != nil {
		return nil, timeseries.Error{metric, timeseries.InvalidSeriesError, fmt.Sprintf("cannot generate URL for tagged metric with graphite name %s", graphiteName)}
	}

	result.RawQuery = url.Values{
		"from":       {strconv.FormatInt(timerange.StartMillis(), 10)},
		"to":         {strconv.FormatInt(timerange.EndMillis()+timerange.ResolutionMillis(), 10)},
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
	queue := &ParallelQueue{
		timeout: tasks.NewTimeout(b.config.Timeout).Timeout(),
	}
	for i := range singleRequests {
		singleRequest := singleRequests[i]
		queue.Do(func() {
			results[i], errs[i] = b.FetchSingleTimeseries(singleRequest)
		})
	}

	queue.Wait()

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
