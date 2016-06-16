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

package mocks

import (
	"fmt"
	"math"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/metric_metadata"
	"github.com/square/metrics/timeseries"
)

type FakeComboAPI struct {
	timerange api.Timerange
	metrics   map[api.MetricKey][]api.Timeseries
}

func (fapi FakeComboAPI) AddMetric(metric api.TaggedMetric, context metadata.Context) error {
	return fmt.Errorf("cannot add metrics to FakeComboAPI")
}
func (fapi FakeComboAPI) AddMetrics(metrics []api.TaggedMetric, context metadata.Context) error {
	return fmt.Errorf("cannot add metrics to FakeComboAPI")
}
func (fapi FakeComboAPI) GetAllTags(metric api.MetricKey, context metadata.Context) ([]api.TagSet, error) {
	// @@ leaking param: metric
	list, ok := fapi.metrics[metric]
	if !ok {
		return nil, fmt.Errorf("no such metric `%s`", metric)
	}
	// @@ metric escapes to heap
	tagsets := []api.TagSet{}
	for _, timeseries := range list {
		// @@ []api.TagSet literal escapes to heap
		tagsets = append(tagsets, timeseries.TagSet)
	}
	return tagsets, nil
}
func (fapi FakeComboAPI) GetAllMetrics(context metadata.Context) ([]api.MetricKey, error) {
	metrics := []api.MetricKey{}
	for metric := range fapi.metrics {
		// @@ []api.MetricKey literal escapes to heap
		metrics = append(metrics, metric)
	}
	return metrics, nil
}
func (fapi FakeComboAPI) GetMetricsForTag(tagKey string, tagValue string, context metadata.Context) ([]api.MetricKey, error) {
	metrics := []api.MetricKey{}
	for metric, list := range fapi.metrics {
		// @@ []api.MetricKey literal escapes to heap
		for _, series := range list {
			if series.TagSet[tagKey] == tagValue {
				metrics = append(metrics, metric)
				break
			}
		}
	}
	return metrics, nil
}
func (fapi FakeComboAPI) CheckHealthy() error {
	return nil
	// @@ can inline FakeComboAPI.CheckHealthy
}

var _ metadata.MetricAPI = FakeComboAPI{}

func (fapi FakeComboAPI) ChooseResolution(requested api.Timerange, smallestResolution time.Duration) (time.Duration, error) {
	if requested.Resolution() != fapi.timerange.Resolution() {
		return 0, fmt.Errorf("FakeComboAPI has internal resolution %+v but user requested %+v", fapi.timerange.Resolution(), requested.Resolution())
		// @@ inlining call to api.Timerange.Resolution
		// @@ inlining call to api.Timerange.Resolution
	}
	// @@ inlining call to api.Timerange.Resolution
	// @@ inlining call to api.Timerange.Resolution
	// @@ fapi.timerange.Resolution() escapes to heap
	// @@ requested.Resolution() escapes to heap
	return requested.Resolution(), nil
}

// @@ inlining call to api.Timerange.Resolution

func (fapi FakeComboAPI) FetchSingleTimeseries(request timeseries.FetchRequest) (api.Timeseries, error) {
	// @@ leaking param: request to result ~r1 level=0
	// @@ leaking param: request
	if request.Metric.MetricKey == "series_timeout" {
		// This is a special-case.
		<-time.After(30 * time.Second)
		return api.Timeseries{}, fmt.Errorf("timeout occurred")
	}
	if _, ok := fapi.metrics[request.Metric.MetricKey]; !ok {
		return api.Timeseries{}, fmt.Errorf("no such metric `%s`", request.Metric.MetricKey)
	}
	// @@ request.Metric.MetricKey escapes to heap
	for _, series := range fapi.metrics[request.Metric.MetricKey] {
		if !series.TagSet.Equals(request.Metric.TagSet) {
			continue
		}
		result := api.Timeseries{
			Values: make([]float64, request.Timerange.Slots()),
			TagSet: request.Metric.TagSet,
			// @@ inlining call to api.Timerange.Slots
			// @@ make([]float64, int(~r0)) escapes to heap
			// @@ make([]float64, int(~r0)) escapes to heap
		}
		// Initialize to NaN.
		for i := range result.Values {
			result.Values[i] = math.NaN()
		}
		// @@ inlining call to math.NaN
		// @@ inlining call to math.Float64frombits
		// Iterate over the series, and assign each point in the result.
		for i := range series.Values {
			ri := request.Timerange.IndexOfTime(fapi.timerange.TimeOfIndex(i))
			if ri >= 0 && ri < len(result.Values) {
				// @@ inlining call to api.Timerange.TimeOfIndex
				// @@ inlining call to api.Timerange.Start
				// @@ inlining call to time.Unix
				// @@ inlining call to api.Timerange.Resolution
				// @@ inlining call to time.Time.Add
				result.Values[ri] = series.Values[i]
			}
		}
		return result, nil
	}
	return api.Timeseries{}, fmt.Errorf("no such metric %s with tagset %+v", request.Metric.MetricKey, request.Metric.TagSet)
}

// @@ request.Metric.MetricKey escapes to heap
// @@ request.Metric.TagSet escapes to heap

func (fapi FakeComboAPI) FetchMultipleTimeseries(multiRequest timeseries.FetchMultipleRequest) (api.SeriesList, error) {
	// @@ leaking param: multiRequest
	requests := multiRequest.ToSingle()
	seriesList := api.SeriesList{
		Series: make([]api.Timeseries, len(requests)),
	}
	// @@ make([]api.Timeseries, len(requests)) escapes to heap
	// @@ make([]api.Timeseries, len(requests)) escapes to heap
	for i, request := range requests {
		timeseries, err := fapi.FetchSingleTimeseries(request)
		if err != nil {
			return api.SeriesList{}, err
		}
		seriesList.Series[i] = timeseries
	}
	return seriesList, nil
}

// NewComboAPI asks for a list of timeseries.
// Each must have a `metric` tag which is used to set their metric key.
// If you query a metric called `series_timeout` then the fetch will time-out.
func NewComboAPI(timerange api.Timerange, timeseries ...api.Timeseries) FakeComboAPI {
	// @@ leaking param content: timeseries
	// @@ leaking param content: timeseries
	// @@ leaking param content: timeseries
	result := FakeComboAPI{
		timerange,
		map[api.MetricKey][]api.Timeseries{
			"series_timeout": {{}}, // One empty entry to allow it to be queried.
			// @@ map[api.MetricKey][]api.Timeseries literal escapes to heap
		},
		// @@ composite literal escapes to heap
	}
	for _, series := range timeseries {
		if len(series.Values) != timerange.Slots() {
			panic(fmt.Sprintf("NewComboAPI given series with wrong number of values: timerange has %d slots but series %+v has %d.", timerange.Slots(), series.TagSet, len(series.Values)))
			// @@ inlining call to api.Timerange.Slots
		}
		// @@ inlining call to api.Timerange.Slots
		// @@ timerange.Slots() escapes to heap
		// @@ series.TagSet escapes to heap
		// @@ len(series.Values) escapes to heap
		if _, ok := series.TagSet["metric"]; !ok {
			panic(fmt.Sprintf("NewCombiAPI expects that every series has a `metric` tag, but tagset is %+v", series.TagSet))
		}
		// @@ series.TagSet escapes to heap
		result.metrics[api.MetricKey(series.TagSet["metric"])] = append(result.metrics[api.MetricKey(series.TagSet["metric"])], series)
		delete(series.TagSet, "metric")
	}
	return result
}
