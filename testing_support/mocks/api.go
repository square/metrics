// Copyright 2015 Square Inc.
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
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/util"
)

type FakeMetricMetadataAPI struct {
	metricTagSets  map[api.MetricKey][]api.TagSet
	metricsForTags map[struct {
		key   string
		value string
	}][]api.MetricKey
}

var _ api.MetricMetadataAPI = (*FakeMetricMetadataAPI)(nil)

func NewFakeMetricMetadataAPI() *FakeMetricMetadataAPI {
	return &FakeMetricMetadataAPI{
		metricTagSets: make(map[api.MetricKey][]api.TagSet),
		metricsForTags: make(map[struct {
			key   string
			value string
		}][]api.MetricKey),
	}
}

func (fa *FakeMetricMetadataAPI) AddPair(tm api.TaggedMetric, gm util.GraphiteMetric, converter *FakeGraphiteConverter) {
	converter.MetricMap[gm] = tm
	fa.AddPairWithoutGraphite(tm)
}

func (fa *FakeMetricMetadataAPI) AddPairWithoutGraphite(tm api.TaggedMetric) {
	fa.metricTagSets[tm.MetricKey] = append(fa.metricTagSets[tm.MetricKey], tm.TagSet)
}

func (fa *FakeMetricMetadataAPI) AddMetric(metric api.TaggedMetric, context api.MetricMetadataAPIContext) error {
	defer context.Profiler.Record("Mock AddMetric")()
	return nil
}

func (fa *FakeMetricMetadataAPI) AddMetrics(metric []api.TaggedMetric, context api.MetricMetadataAPIContext) error {
	defer context.Profiler.Record("Mock AddMetrics")()
	return nil
}

func (fa *FakeMetricMetadataAPI) RemoveMetric(metric api.TaggedMetric, context api.MetricMetadataAPIContext) error {
	defer context.Profiler.Record("Mock RemoveMetric")()
	return nil
}

func (fa *FakeMetricMetadataAPI) GetAllTags(metricKey api.MetricKey, context api.MetricMetadataAPIContext) ([]api.TagSet, error) {
	defer context.Profiler.Record("Mock GetAllTags")()
	if len(fa.metricTagSets[metricKey]) == 0 {
		// This matches the behavior of the Cassandra API
		return nil, fmt.Errorf("metric %s does not exist", metricKey)
	}
	return fa.metricTagSets[metricKey], nil
}

func (fa *FakeMetricMetadataAPI) GetAllMetrics(context api.MetricMetadataAPIContext) ([]api.MetricKey, error) {
	defer context.Profiler.Record("Mock GetAllMetrics")()
	array := []api.MetricKey{}
	for key := range fa.metricTagSets {
		array = append(array, key)
	}
	return array, nil
}

// AddMetricsForTag adds a metric to the Key/Value set list.
func (fa *FakeMetricMetadataAPI) AddMetricsForTag(key string, value string, metric string) {
	pair := struct {
		key   string
		value string
	}{key, value}
	// If the slice was previously nil, it will be expanded.
	fa.metricsForTags[pair] = append(fa.metricsForTags[pair], api.MetricKey(metric))
}

func (fa *FakeMetricMetadataAPI) GetMetricsForTag(tagKey, tagValue string, context api.MetricMetadataAPIContext) ([]api.MetricKey, error) {
	defer context.Profiler.Record("Mock GetMetricsForTag")()
	list := []api.MetricKey{}
MetricLoop:
	for metric, tagsets := range fa.metricTagSets {
		for _, tagset := range tagsets {
			for key, val := range tagset {
				if key == tagKey && val == tagValue {
					list = append(list, api.MetricKey(metric))
					continue MetricLoop
				}
			}
		}
	}
	return list, nil
}

type FakeGraphiteConverter struct {
	MetricMap map[util.GraphiteMetric]api.TaggedMetric
}

var _ util.GraphiteConverter = (*FakeGraphiteConverter)(nil)

func (fa *FakeGraphiteConverter) ToGraphiteName(metric api.TaggedMetric) (util.GraphiteMetric, error) {
	for k, v := range fa.MetricMap {
		if reflect.DeepEqual(v, metric) {
			return k, nil
		}
	}
	return "", fmt.Errorf("No mapping for tagged metric %+v to tagged metric", metric)
}

func (fa *FakeGraphiteConverter) ToTaggedName(metric util.GraphiteMetric) (api.TaggedMetric, error) {
	tm, exists := fa.MetricMap[metric]
	if !exists {
		return api.TaggedMetric{}, fmt.Errorf("No mapping for graphite metric %+s to graphite metric", string(metric))
	}

	return tm, nil
}

type FakeTimeseriesStorageAPI struct{}

func (f FakeTimeseriesStorageAPI) ChooseResolution(requested api.Timerange, smallestResolution time.Duration) time.Duration {
	return requested.Resolution()
}

func (f FakeTimeseriesStorageAPI) FetchSingleTimeseries(request api.FetchTimeseriesRequest) (api.Timeseries, error) {
	defer request.Profiler.Record("Mock FetchSingleTimeseries")()
	metricMap := map[api.MetricKey][]api.Timeseries{
		"series_1": {{[]float64{1, 2, 3, 4, 5}, api.ParseTagSet("dc=west")}},
		"series_2": {{[]float64{1, 2, 3, 4, 5}, api.ParseTagSet("dc=west")}, {[]float64{3, 0, 3, 6, 2}, api.ParseTagSet("dc=east")}},
		"series_3": {{[]float64{1, 1, 1, 4, 4}, api.ParseTagSet("dc=west")}, {[]float64{5, 5, 5, 2, 2}, api.ParseTagSet("dc=east")}, {[]float64{3, 3, 3, 3, 3}, api.ParseTagSet("dc=north")}},
	}
	if string(request.Metric.MetricKey) == "series_timeout" {
		<-make(chan struct{}) // block forever
	}
	list, ok := metricMap[request.Metric.MetricKey]
	if !ok {
		return api.Timeseries{}, errors.New("internal error")
	}
	for _, series := range list {
		if request.Metric.TagSet.Serialize() == series.TagSet.Serialize() {
			// Cut the values based on the Timerange.
			values := make([]float64, request.Timerange.Slots())
			for i := range values {
				values[i] = series.Values[i+int(request.Timerange.Start())/30]
			}
			return api.Timeseries{values, series.TagSet}, nil
		}
	}
	return api.Timeseries{}, errors.New("internal error")
}

func (f FakeTimeseriesStorageAPI) FetchMultipleTimeseries(request api.FetchMultipleTimeseriesRequest) (api.SeriesList, error) {
	defer request.Profiler.Record("Mock FetchMultipleTimeseries")()
	timeseries := make([]api.Timeseries, 0)

	singleRequests := request.ToSingle()
	for _, singleRequest := range singleRequests {
		series, err := f.FetchSingleTimeseries(singleRequest)
		if err != nil {
			continue
		}
		timeseries = append(timeseries, series)
	}

	return api.SeriesList{
		Series:    timeseries,
		Timerange: request.Timerange,
	}, nil
}
