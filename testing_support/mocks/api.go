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

func NewFakeMetricMetadataAPI() FakeMetricMetadataAPI {
	return FakeMetricMetadataAPI{
		metricTagSets: make(map[api.MetricKey][]api.TagSet),
		metricsForTags: make(map[struct {
			key   string
			value string
		}][]api.MetricKey),
	}
}

func (fa *FakeMetricMetadataAPI) AddPair(tm api.TaggedMetric, gm util.GraphiteMetric, converter *FakeGraphiteConverter) {
	converter.MetricMap[gm] = tm

	if metricTagSets, ok := fa.metricTagSets[tm.MetricKey]; !ok {
		fa.metricTagSets[tm.MetricKey] = []api.TagSet{tm.TagSet}
	} else {
		fa.metricTagSets[tm.MetricKey] = append(metricTagSets, tm.TagSet)
	}
}

func (fa *FakeMetricMetadataAPI) AddPairWithoutGraphite(tm api.TaggedMetric, gm util.GraphiteMetric) {
	if metricTagSets, ok := fa.metricTagSets[tm.MetricKey]; !ok {
		fa.metricTagSets[tm.MetricKey] = []api.TagSet{tm.TagSet}
	} else {
		fa.metricTagSets[tm.MetricKey] = append(metricTagSets, tm.TagSet)
	}
}

func (fa *FakeMetricMetadataAPI) AddMetric(metric api.TaggedMetric) error {
	return nil
}

func (fa *FakeMetricMetadataAPI) AddMetrics(metric []api.TaggedMetric) error {
	return nil
}

func (fa *FakeMetricMetadataAPI) RemoveMetric(metric api.TaggedMetric) error {
	return nil
}

func (fa *FakeMetricMetadataAPI) GetAllTags(metricKey api.MetricKey) ([]api.TagSet, error) {
	return fa.metricTagSets[metricKey], nil
}

func (fa *FakeMetricMetadataAPI) GetAllMetrics() ([]api.MetricKey, error) {
	array := []api.MetricKey{}
	for key := range fa.metricTagSets {
		array = append(array, key)
	}
	return array, nil
}

// Adds a metric to the Key/Value set list.
func (fa *FakeMetricMetadataAPI) AddMetricsForTag(key string, value string, metric string) {
	pair := struct {
		key   string
		value string
	}{key, value}
	// If the slice was previously nil, it will be expanded.
	fa.metricsForTags[pair] = append(fa.metricsForTags[pair], api.MetricKey(metric))
}

func (fa *FakeMetricMetadataAPI) GetMetricsForTag(tagKey, tagValue string) ([]api.MetricKey, error) {
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
	return "", errors.New(fmt.Sprintf("No mapping for tagged metric %+v to tagged metric", metric))
}

func (fa *FakeGraphiteConverter) ToTaggedName(metric util.GraphiteMetric) (api.TaggedMetric, error) {
	tm, exists := fa.MetricMap[metric]
	if !exists {
		return api.TaggedMetric{}, errors.New(fmt.Sprintf("No mapping for graphite metric %+s to graphite metric", string(metric)))
	}

	return tm, nil
}
