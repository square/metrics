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
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/inspect"
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

// TimeSource represents a source of time values.
// Its zero value will give the current time.
type TimeSource struct {
	GetTime func() time.Time
}

func (t TimeSource) Now() time.Time {
	if t.GetTime == nil {
		return time.Now()
	}
	return t.GetTime()
}

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
	BaseURL                 string       `yaml:"base_url"`
	TenantID                string       `yaml:"tenant_id"`
	Resolutions             []Resolution `yaml:"resolutions"`           // Resolutions are ordered by priority: best (typically finest) first.
	MaxSimultaneousRequests int          `yaml:"simultaneous_requests"` // simultaneous requests limits the number of concurrent single-fetches for each multi-fetch

	GraphiteMetricConverter util.GraphiteConverter

	HTTPClient httpClient
	TimeSource TimeSource
}

// NewBlueflood uses the Config to create an instance of Blueflood.
func NewBlueflood(c Config) timeseries.StorageAPI {
	if c.HTTPClient == nil {
		c.HTTPClient = http.DefaultClient
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

// ChooseResolution will choose the finest-grained resolution for which an
// interval fetch plan exists that is at least as coarse as the lower bound.
func (b *Blueflood) ChooseResolution(requested api.Timerange, lowerBound time.Duration) (time.Duration, error) {
	now := b.config.TimeSource.Now()
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

// FetchSingleTimeseries fetches a timeseries with the given tagged metric.
// The resolution is required to be supported (as ensured by ChooseResolution).
func (b *Blueflood) FetchSingleTimeseries(request timeseries.FetchRequest) (api.Timeseries, error) {
	defer request.Profiler.RecordWithDescription("Blueflood FetchSingleTimeseries", request.Metric.String())()
	plan, err := b.createPlan(request.RequestDetails)
	if err != nil {
		return api.Timeseries{}, err
	}
	return b.fetchTimeseries(request.Metric, plan, request.Profiler, request.Timeout)
}

// FetchMultipleRequest fetches multiple timeseries. It requires that the
// The resolution is required to be supported (as ensured by ChooseResolution).
func (b *Blueflood) FetchMultipleTimeseries(request timeseries.FetchMultipleRequest) (api.SeriesList, error) {
	defer request.Profiler.Record("Blueflood FetchMultipleTimeseries")()
	plan, err := b.createPlan(request.RequestDetails)
	if err != nil {
		return api.SeriesList{}, err
	}

	singleRequests := request.ToSingle()
	results := make([]api.Timeseries, len(singleRequests))
	queue := tasks.NewParallelQueue(b.config.MaxSimultaneousRequests, request.Timeout)
	for i := range singleRequests {
		i := i // Captures it in a new local for the closure.
		queue.Do(func() error {
			result, err := b.fetchTimeseries(singleRequests[i].Metric, plan, request.Profiler, request.Timeout)
			if err != nil {
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

// fetchPlan contains data required to fetch a timeseries by stitching together
// multi-resolution data.
type fetchPlan struct {
	intervals map[Resolution]api.Interval
	sampler   sampler
	timerange api.Timerange
}

// createPlan uses the specified request details (which don't depend on the
// metric itself) to create a plan for fetching it with multi-resolution data.
func (b *Blueflood) createPlan(request timeseries.RequestDetails) (fetchPlan, error) {
	samplerFunc, ok := samplerMap[request.SampleMethod]
	if !ok {
		return fetchPlan{}, fmt.Errorf("unsupported SampleMethod %s", request.SampleMethod.String())
	}
	// Extend it one point forward, unless that would fetch past the current time.
	modifiedRange := request.Timerange
	if modifiedRange.End().Add(modifiedRange.Resolution()).Before(b.config.TimeSource.Now()) {
		modifiedRange = modifiedRange.ExtendAfter(modifiedRange.Resolution())
	}
	intervals, err := planFetchIntervalsWithOnlyFiner(b.config.Resolutions, b.config.TimeSource.Now(), modifiedRange)
	if err != nil {
		return fetchPlan{}, err
	}
	return fetchPlan{
		intervals: intervals,
		sampler:   samplerFunc,
		timerange: request.Timerange,
	}, nil
}

// fetchTimeseries uses the provided plan to fetch the timeseries from Blueflood
// using several HTTP queries. FetchMultipleTimeseries defers to this method,
// rather than FetchSingleTimeseries, in order to prevent duplicating work on a
// per-timeseries basis.
func (b *Blueflood) fetchTimeseries(metric api.TaggedMetric, plan fetchPlan, profiler *inspect.Profiler, timeout *tasks.Timeout) (api.Timeseries, error) {
	queue := tasks.NewParallelQueue(len(plan.intervals), timeout)
	allPoints := []metricPoint{}
	for resolution, interval := range plan.intervals {
		resolution, interval := resolution, interval
		queue.Do(func() error {
			defer profiler.RecordWithDescription("Blueflood FetchSingleTimeseries Resolution", fmt.Sprintf("%s at %+v", metric.String(), resolution.Resolution))()
			// Construct the URL
			queryURL, err := b.constructURL(metric, interval, plan.sampler, resolution)
			if err != nil {
				return err
			}
			// Then query it.
			points, err := b.fetchTimeseriesHTTP(queryURL, timeout)
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

	values := samplePoints(allPoints, plan.timerange, plan.sampler)

	return api.Timeseries{
		Values: values,
		TagSet: metric.TagSet,
	}, nil
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

type httpClient interface {
	// our own client to mock out the standard golang HTTP Client.
	Get(string) (*http.Response, error)
	Do(*http.Request) (*http.Response, error)
}

// fetch fetches from the backend, asynchronously calling performFetch and cancelling on timeout.
func (b *Blueflood) fetchTimeseriesHTTP(queryURL *url.URL, timeout *tasks.Timeout) ([]metricPoint, error) {
	request, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}
	request.Cancel = timeout.Done()
	response, err := b.config.HTTPClient.Do(request)
	if err != nil {
		return nil, timeseries.FetchError{Code: 500, Message: fmt.Sprintf("error fetching from Blueflood at URL %q: %s", queryURL.String(), err.Error())}
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, timeseries.FetchError{Code: 500, Message: fmt.Sprintf("error reading from Blueflood response body at URL %q: %s", queryURL.String(), err.Error())}
	}
	err = response.Body.Close()
	if err != nil {
		return nil, timeseries.FetchError{Code: 500, Message: fmt.Sprintf("error finishing response from Blueflood at URL %q: %s", queryURL.String(), err.Error())}
	}
	var parsedJSON queryResponse
	err = json.Unmarshal(body, &parsedJSON)
	if err != nil {
		return nil, timeseries.FetchError{Code: 500, Message: fmt.Sprintf("error unmarshaling JSON from Blueflood at URL %q: %s;\nBody:%s", queryURL.String(), err.Error(), body)}
	}
	return parsedJSON.Values, nil
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
