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
	// @@ leaking param: r
	return fmt.Sprintf("Name: %s Duration: %+v", r.Name, r.Resolution)
}

// @@ r.Name escapes to heap
// @@ r.Resolution escapes to heap

type Config struct {
	BaseURL                 string        `yaml:"base_url"`
	TenantID                string        `yaml:"tenant_id"`
	Resolutions             []Resolution  `yaml:"resolutions"`           // Resolutions are ordered by priority: best (typically finest) first.
	Timeout                 time.Duration `yaml:"timeout"`               // Timeout is the amount of time a single fetch request is allowed.
	MaxSimultaneousRequests int           `yaml:"simultaneous_requests"` // simultaneous requests limits the number of concurrent single-fetches for each multi-fetch

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
	// @@ leaking param: c to result ~r1 level=-1
	if c.HTTPClient == nil {
		// @@ can inline NewBlueflood
		c.HTTPClient = http.DefaultClient
	}
	// @@ http.DefaultClient escapes to heap
	if c.TimeSource == nil {
		c.TimeSource = time.Now
	}
	if c.MaxSimultaneousRequests == 0 {
		c.MaxSimultaneousRequests = 5
	}

	b := &Blueflood{
		config: c,
	}
	// @@ &Blueflood literal escapes to heap
	// TODO: copy internal config structures to prevent modification?
	return b
}

// @@ b escapes to heap

// ChooseResolution will choose the finest-grained resolution for which an interval fetch plan exists that
// is at least as coarse as the lower bound.
func (b *Blueflood) ChooseResolution(requested api.Timerange, lowerBound time.Duration) (time.Duration, error) {
	// @@ leaking param content: b
	now := b.config.TimeSource()
	for i, current := range b.config.Resolutions {
		if current.Resolution < lowerBound || current.Resolution < requested.Resolution() {
			continue
			// @@ inlining call to api.Timerange.Resolution
		}
		_, err := planFetchIntervals(b.config.Resolutions[:i+1], now, requested.Interval())
		if err == nil {
			// @@ inlining call to api.Timerange.Interval
			// @@ inlining call to api.Timerange.Start
			// @@ inlining call to time.Unix
			// @@ inlining call to api.Timerange.End
			// @@ inlining call to time.Unix
			return current.Resolution, nil
		}
	}
	return 0, fmt.Errorf("cannot choose resolution for timerange %+v; available resolutions do not live long enough or are not available soon enough.", requested)
}

// @@ requested escapes to heap

// planFetchIntervals will plan the (point-count minimal) request intervals needed to cover the given timerange.
// the resolutions slice should be sorted, with the finest-grained resolution first.
func planFetchIntervals(resolutions []Resolution, now time.Time, requestInterval api.Interval) (map[Resolution]api.Interval, error) {
	// @@ leaking param content: resolutions
	answer := map[Resolution]api.Interval{}
	// Note: for anything other than FULL, a Blueflood returned point corresponds to the period FOLLOWING that point.
	// @@ map[Resolution]api.Interval literal escapes to heap
	// e.g. at 1hr resolution, a 4pm point summarizes all points in [4pm, 5pm], exclusive of 5pm.
	requestTimerange := requestInterval.CoveringTimerange(resolutions[len(resolutions)-1].Resolution)
	here := requestTimerange.Start()
	// @@ inlining call to api.Interval.CoveringTimerange
	// @@ inlining call to time.Duration.Seconds
	// @@ inlining call to time.Time.UnixNano
	// @@ inlining call to time.Time.UnixNano
	end := requestTimerange.End()
	// @@ inlining call to api.Timerange.Start
	// @@ inlining call to time.Unix
	for i := len(resolutions) - 1; i >= 0; i-- {
		// @@ inlining call to api.Timerange.End
		// @@ inlining call to time.Unix
		resolution := resolutions[i]
		if !here.Before(end) {
			break
			// @@ inlining call to time.Time.Before
		}
		if here.Before(now.Add(-resolution.TimeToLive)) {
			// Expired
			// @@ inlining call to time.Time.Add
			// @@ inlining call to time.Time.Before
			return nil, fmt.Errorf("resolutions up to %+v only live for %+v, but request needs data that's at least %+v old", resolution.Resolution, resolution.TimeToLive, now.Sub(here))
		}
		// @@ resolution.Resolution escapes to heap
		// @@ resolution.TimeToLive escapes to heap
		// @@ now.Sub(here) escapes to heap

		// clipEnd is the end of requested interval,
		// or where the data is not yet available,
		// whichever is earlier.
		clipEnd := now.Add(-resolution.FirstAvailable)
		if end.Before(clipEnd) {
			// @@ inlining call to time.Time.Add
			clipEnd = end
			// @@ inlining call to time.Time.Before
		}

		// count how many resolution intervals pass from now until then.
		count := clipEnd.Sub(here) / resolution.Resolution
		if count < 0 {
			count = 0
		}

		// advance that number of intervals
		newHere := here.Add(count * resolution.Resolution)

		// @@ inlining call to time.Time.Add
		if newHere != here {
			// At least one point is included, so:
			answer[resolution] = api.Interval{Start: here, End: newHere}
			here = newHere
		}
	}
	return answer, nil
}

// planFetchIntervalsWithOnlyFiner assumes that the requested range is as coarse as desired.
// Hence, it will trim all coarser resolutions before doing planning.
func planFetchIntervalsWithOnlyFiner(resolutions []Resolution, now time.Time, requestRange api.Timerange) (map[Resolution]api.Interval, error) {
	// @@ leaking param content: resolutions
	// @@ leaking param content: resolutions
	for i := range resolutions {
		if resolutions[i].Resolution > requestRange.Resolution() {
			if i == 0 {
				// @@ inlining call to api.Timerange.Resolution
				return nil, fmt.Errorf("No resolutions are available at least as fine as the chosen %+v", requestRange.Resolution())
			}
			// @@ inlining call to api.Timerange.Resolution
			// @@ requestRange.Resolution() escapes to heap
			return planFetchIntervals(resolutions[:i], now, requestRange.Interval())
		}
		// @@ inlining call to api.Timerange.Interval
		// @@ inlining call to api.Timerange.Start
		// @@ inlining call to time.Unix
		// @@ inlining call to api.Timerange.End
		// @@ inlining call to time.Unix
	}
	return planFetchIntervals(resolutions, now, requestRange.Interval())
}

// @@ inlining call to api.Timerange.Interval
// @@ inlining call to api.Timerange.Start
// @@ inlining call to time.Unix
// @@ inlining call to api.Timerange.End
// @@ inlining call to time.Unix

// fetchSingleTimeseriesPrepped uses info prepped by FetchSingleTimeseries and
// FetchMultipleTimeseries to fetch data. FetchMultipleTimeseries defers to this
// method, instead of FetchSingleTimeseries, to avoid duplication of work across
// each of these calls.
func (b *Blueflood) fetchSingleTimeseriesPrepped(request timeseries.FetchRequest, intervals map[Resolution]api.Interval, sampler sampler) (api.Timeseries, error) {
	// @@ leaking param: request
	// @@ leaking param: b
	// @@ leaking param: sampler
	// @@ leaking param content: intervals

	// @@ moved to heap: request
	// @@ mark escaped content: intervals
	queue := tasks.NewParallelQueue(len(intervals), b.config.Timeout)
	allPoints := []metricPoint{}

	// @@ moved to heap: allPoints
	// @@ []metricPoint literal escapes to heap
	for resolution, interval := range intervals {
		resolution, interval := resolution, interval
		queue.Do(func() error {
			defer request.Profiler.RecordWithDescription("Blueflood FetchSingleTimeseries Resolution", fmt.Sprintf("%s at %+v", request.Metric.String(), resolution.Resolution))()
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			points, err := b.requestPoints(request.Metric, interval, sampler, resolution)
			// @@ leaking closure reference request
			// @@ request.Metric.String() escapes to heap
			// @@ resolution.Resolution escapes to heap
			// @@ leaking closure reference request
			// @@ leaking closure reference request
			// @@ leaking closure reference resolution
			// @@ &request escapes to heap
			if err != nil {
				// @@ leaking closure reference b
				// @@ leaking closure reference sampler
				return err
			}
			queue.Lock()
			defer queue.Unlock()
			// @@ queue.Mutex escapes to heap
			// @@ leaking closure reference queue
			// @@ leaking closure reference queue
			allPoints = append(allPoints, points...)
			// @@ queue.Mutex escapes to heap
			return nil
			// @@ &allPoints escapes to heap
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
	return api.Timeseries{}, nil
}

func (b *Blueflood) prepWork(request timeseries.RequestDetails) (map[Resolution]api.Interval, sampler, error) {
	// @@ leaking param content: b
	samplerFunc, ok := samplerMap[request.SampleMethod]
	if !ok {
		return nil, sampler{}, fmt.Errorf("unsupported SampleMethod %s", request.SampleMethod.String())
	}
	// @@ request.SampleMethod.String() escapes to heap
	// Extend it one point forward, unless that would fetch past the current time.
	modifiedRange := request.Timerange
	if modifiedRange.End().Add(modifiedRange.Resolution()).Before(b.config.TimeSource()) {
		modifiedRange = modifiedRange.ExtendAfter(modifiedRange.Resolution())
		// @@ inlining call to api.Timerange.End
		// @@ inlining call to time.Unix
		// @@ inlining call to api.Timerange.Resolution
		// @@ inlining call to time.Time.Add
		// @@ inlining call to time.Time.Before
	}
	// @@ inlining call to api.Timerange.Resolution
	intervals, err := planFetchIntervalsWithOnlyFiner(b.config.Resolutions, b.config.TimeSource(), modifiedRange)
	if err != nil {
		return nil, sampler{}, err
	}
	return intervals, samplerFunc, nil
}

// FetchSingleTimeseries fetches a timeseries with the given tagged metric.
// It requires that the resolution is supported.
func (b *Blueflood) FetchSingleTimeseries(request timeseries.FetchRequest) (api.Timeseries, error) {
	// @@ leaking param: request
	// @@ leaking param content: b
	// @@ leaking param: b
	defer request.Profiler.RecordWithDescription("Blueflood FetchSingleTimeseries", request.Metric.String())()
	intervals, samplerFunc, err := b.prepWork(request.RequestDetails)
	if err != nil {
		return api.Timeseries{}, err
	}
	return b.fetchSingleTimeseriesPrepped(request, intervals, samplerFunc)
}

func (b *Blueflood) requestPoints(metric api.TaggedMetric, interval api.Interval, sampler sampler, resolution Resolution) ([]metricPoint, error) {
	// @@ leaking param content: b
	// @@ leaking param: metric
	// @@ leaking param: sampler
	// @@ leaking param: resolution
	// @@ leaking param: b
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
	// @@ leaking param: metric
	// @@ leaking param content: b
	// @@ leaking param content: b
	// @@ leaking param content: b
	// @@ leaking param content: sampler
	// @@ leaking param: sampler
	// @@ leaking param: resolution
	graphiteName, err := b.config.GraphiteMetricConverter.ToGraphiteName(metric)
	if err != nil {
		return nil, timeseries.Error{metric, timeseries.InvalidSeriesError, "cannot convert to graphite name"}
	}
	// @@ timeseries.Error literal escapes to heap

	result, err := url.Parse(fmt.Sprintf("%s/v2.0/%s/views/%s", b.config.BaseURL, b.config.TenantID, graphiteName))
	if err != nil {
		// @@ b.config.BaseURL escapes to heap
		// @@ b.config.TenantID escapes to heap
		// @@ graphiteName escapes to heap
		return nil, timeseries.Error{metric, timeseries.InvalidSeriesError, fmt.Sprintf("cannot generate URL for tagged metric with graphite name %s", graphiteName)}
	}
	// @@ graphiteName escapes to heap
	// @@ timeseries.Error literal escapes to heap

	result.RawQuery = url.Values{
		"from": {strconv.FormatInt(int64(interval.Start.UnixNano()/1e6), 10)},
		"to":   {strconv.FormatInt(int64(interval.End.UnixNano()/1e6-1), 10)},
		// @@ inlining call to time.Time.UnixNano
		// @@ composite literal escapes to heap
		"resolution": {resolution.Name},
		// @@ inlining call to time.Time.UnixNano
		// @@ composite literal escapes to heap
		"select": {fmt.Sprintf("numPoints,%s", strings.ToLower(sampler.fieldName))},
		// @@ composite literal escapes to heap
	}.Encode()
	// @@ strings.ToLower(sampler.fieldName) escapes to heap
	// @@ composite literal escapes to heap

	return result, nil
}

// performFetch is a synchronous method that fetches the given URL.
func (b *Blueflood) performFetch(queryURL *url.URL) (queryResponse, error) {
	// @@ leaking param content: queryURL
	// @@ leaking param content: b
	// @@ leaking param content: queryURL
	// @@ leaking param content: queryURL
	// @@ leaking param content: queryURL
	// @@ leaking param content: queryURL
	resp, err := b.config.HTTPClient.Get(queryURL.String())
	if err != nil {
		// TODO: report the right metric
		return queryResponse{}, timeseries.FetchError{Code: 500, Message: fmt.Sprintf("error fetching Blueflood at URL %q: %s", queryURL.String(), err.Error())}
	}
	// @@ queryURL.String() escapes to heap
	// @@ err.Error() escapes to heap
	// @@ timeseries.FetchError literal escapes to heap
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// @@ resp.Body escapes to heap
		// TODO: report the right metric
		return queryResponse{}, timeseries.FetchError{Code: 500, Message: fmt.Sprintf("error reading response from Blueflood at URL %q: %s", queryURL.String(), err.Error())}
	}
	// @@ queryURL.String() escapes to heap
	// @@ err.Error() escapes to heap
	// @@ timeseries.FetchError literal escapes to heap

	// Don't try and JSON parse a non-200 response
	if resp.StatusCode != http.StatusOK {
		// TODO: report the right metric
		return queryResponse{}, timeseries.FetchError{Code: 500, Message: fmt.Sprintf("error fetching Blueflood at URL %q, got %d: %s", queryURL.String(), resp.StatusCode, string(body))}
	}
	// @@ queryURL.String() escapes to heap
	// @@ resp.StatusCode escapes to heap
	// @@ string(body) escapes to heap
	// @@ string(body) escapes to heap
	// @@ timeseries.FetchError literal escapes to heap

	var parsedJSON queryResponse
	err = json.Unmarshal(body, &parsedJSON)
	// @@ moved to heap: parsedJSON
	if err != nil {
		// @@ &parsedJSON escapes to heap
		// @@ &parsedJSON escapes to heap
		// TODO: report the right metric
		return queryResponse{}, timeseries.Error{api.TaggedMetric{}, timeseries.FetchIOError, "error while fetching - json decoding\nBody: " + string(body) + "\nError: " + err.Error() + "\nURL: " + queryURL.String()}
	}
	// @@ timeseries.Error literal escapes to heap
	// @@ "error while fetching - json decoding\nBody: " + string(body) + "\nError: " + err.Error() + "\nURL: " + queryURL.String() escapes to heap
	return parsedJSON, nil
}

// fetch fetches from the backend, asynchronously calling performFetch and cancelling on timeout.
func (b *Blueflood) fetch(queryURL *url.URL) (queryResponse, error) {
	// @@ leaking param content: b
	// @@ leaking param content: queryURL
	// @@ leaking param: b
	// @@ leaking param: queryURL
	type Answer struct {
		response queryResponse
		err      error
	}
	answer := make(chan Answer, 1)
	go func() {
		// @@ make(chan Answer, 1) escapes to heap
		response, err := b.performFetch(queryURL)
		// @@ func literal escapes to heap
		// @@ func literal escapes to heap
		answer <- Answer{response, err}
	}()
	select {
	case result := <-answer:
		return result.response, result.err
	case <-time.After(b.config.Timeout):
		// TODO: report the right metric
		return queryResponse{}, timeseries.Error{api.TaggedMetric{}, timeseries.FetchTimeoutError, ""}
	}
	// @@ timeseries.Error literal escapes to heap
}

type sampler struct {
	fieldName    string                          // Name of field in Blueflood JSON response
	selectField  func(point metricPoint) float64 // Function for extracting field from metricPoint
	sampleBucket func([]float64) float64         // Function to sample from the bucket (e.g., min, mean, max)
}

// sampleResult samples the points into a uniform slice of float64s.
func samplePoints(points []metricPoint, timerange api.Timerange, sampler sampler) []float64 {
	// A bucket holds a set of points corresponding to one interval in the result.
	buckets := make([][]float64, timerange.Slots())
	for _, point := range points {
		// @@ inlining call to api.Timerange.Slots
		// @@ make([][]float64, int(~r0)) escapes to heap
		pointValue := sampler.selectField(point)
		index := (point.Timestamp - timerange.StartMillis()) / timerange.ResolutionMillis()
		if index < 0 || int(index) >= len(buckets) {
			// @@ inlining call to api.Timerange.StartMillis
			// @@ inlining call to api.Timerange.ResolutionMillis
			continue
		}
		buckets[index] = append(buckets[index], pointValue)
	}

	// values will hold the final values to be returned as the series.
	values := make([]float64, timerange.Slots())

	// @@ inlining call to api.Timerange.Slots
	// @@ make([]float64, int(~r0)) escapes to heap
	// @@ make([]float64, int(~r0)) escapes to heap
	for i, bucket := range buckets {
		if len(bucket) == 0 {
			values[i] = math.NaN()
			continue
			// @@ inlining call to math.NaN
			// @@ inlining call to math.Float64frombits
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
			// @@ can inline glob.func1
			value := 0.0
			count := 0
			for _, v := range bucket {
				if !math.IsNaN(v) {
					value += v
					// @@ inlining call to math.IsNaN
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
			// @@ can inline glob.func3
			smallest := math.NaN()
			for _, v := range bucket {
				// @@ inlining call to math.NaN
				// @@ inlining call to math.Float64frombits
				if math.IsNaN(v) {
					continue
					// @@ inlining call to math.IsNaN
				}
				if math.IsNaN(smallest) {
					smallest = v
					// @@ inlining call to math.IsNaN
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
			// @@ can inline glob.func5
			largest := math.NaN()
			for _, v := range bucket {
				// @@ inlining call to math.NaN
				// @@ inlining call to math.Float64frombits
				if math.IsNaN(v) {
					continue
					// @@ inlining call to math.IsNaN
				}
				if math.IsNaN(largest) {
					largest = v
					// @@ inlining call to math.IsNaN
				} else {
					largest = math.Max(largest, v)
				}
			}
			return largest
		},
	},
}

func (b *Blueflood) FetchMultipleTimeseries(request timeseries.FetchMultipleRequest) (api.SeriesList, error) {
	// @@ leaking param: b
	// @@ leaking param: request
	defer request.Profiler.Record("Blueflood FetchMultipleTimeseries")()
	intervals, samplerFunc, err := b.prepWork(request.RequestDetails)
	if err != nil {
		return api.SeriesList{}, err
	}

	singleRequests := request.ToSingle()
	results := make([]api.Timeseries, len(singleRequests))
	queue := tasks.NewParallelQueue(b.config.MaxSimultaneousRequests, b.config.Timeout)
	// @@ make([]api.Timeseries, len(singleRequests)) escapes to heap
	// @@ make([]api.Timeseries, len(singleRequests)) escapes to heap
	// @@ make([]api.Timeseries, len(singleRequests)) escapes to heap
	for i := range singleRequests {
		i := i // Captures it in a new local for the closure.
		queue.Do(func() error {
			result, err := b.fetchSingleTimeseriesPrepped(singleRequests[i], intervals, samplerFunc)
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			if err != nil {
				// @@ leaking closure reference b
				// @@ leaking closure reference samplerFunc
				return err
			}
			results[i] = result
			return nil
		})
	}

	if err := queue.Wait(); err != nil {
		return api.SeriesList{}, err
	}

	return api.SeriesList{
		Series: results,
	}, nil
}

// CheckHealthy checks if the blueflood server is available by querying /v2.0
func (b *Blueflood) CheckHealthy() error {
	// @@ leaking param content: b
	// @@ leaking param content: b
	resp, err := b.config.HTTPClient.Get(fmt.Sprintf("%s/v2.0", b.config.BaseURL))
	if err != nil {
		// @@ b.config.BaseURL escapes to heap
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// @@ resp.Body escapes to heap
			return err
		}
		return fmt.Errorf("Blueflood returned an unhealthy status of %d: %s", resp.StatusCode, string(body))
	}
	// @@ resp.StatusCode escapes to heap
	// @@ string(body) escapes to heap
	// @@ string(body) escapes to heap

	return nil
}
