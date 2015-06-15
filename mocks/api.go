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
)

type FakeApi struct {
	metricMap      map[api.GraphiteMetric]api.TaggedMetric
	metricTagSets  map[api.MetricKey][]api.TagSet
	metricsForTags map[struct {
		key   string
		value string
	}][]api.MetricKey
}

func NewFakeApi() *FakeApi {
	return &FakeApi{
		metricMap:     make(map[api.GraphiteMetric]api.TaggedMetric),
		metricTagSets: make(map[api.MetricKey][]api.TagSet),
		metricsForTags: make(map[struct {
			key   string
			value string
		}][]api.MetricKey),
	}
}

func (fa *FakeApi) AddPair(tm api.TaggedMetric, gm api.GraphiteMetric) {
	fa.metricMap[gm] = tm

	if metricTagSets, ok := fa.metricTagSets[tm.MetricKey]; !ok {
		fa.metricTagSets[tm.MetricKey] = []api.TagSet{tm.TagSet}
	} else {
		fa.metricTagSets[tm.MetricKey] = append(metricTagSets, tm.TagSet)
	}
}

func (fa *FakeApi) AddMetric(metric api.TaggedMetric) error {
	return nil
}

func (fa *FakeApi) RemoveMetric(metric api.TaggedMetric) error {
	return nil
}

func (fa *FakeApi) ToGraphiteName(metric api.TaggedMetric) (api.GraphiteMetric, error) {
	for k, v := range fa.metricMap {
		if reflect.DeepEqual(v, metric) {
			return k, nil
		}
	}
	return "", errors.New(fmt.Sprintf("No mapping for tagged metric %+v to tagged metric", metric))
}

func (fa *FakeApi) ToTaggedName(metric api.GraphiteMetric) (api.TaggedMetric, error) {
	tm, exists := fa.metricMap[metric]
	if !exists {
		return api.TaggedMetric{}, errors.New(fmt.Sprintf("No mapping for graphite metric %+s to graphite metric", string(metric)))
	}

	return tm, nil
}

func (fa *FakeApi) GetAllTags(metricKey api.MetricKey) ([]api.TagSet, error) {
	return fa.metricTagSets[metricKey], nil
}

func (fa *FakeApi) GetAllMetrics() ([]api.MetricKey, error) {
	return nil, errors.New("Implement me")
}

// Adds a metric to the Key/Value set list.
func (fa *FakeApi) AddMetricsForTag(key string, value string, metric string) {
	pair := struct {
		key   string
		value string
	}{key, value}
	// If the slice was previously nil, it will be expanded.
	fa.metricsForTags[pair] = append(fa.metricsForTags[pair], api.MetricKey(metric))
}

func (fa *FakeApi) GetMetricsForTag(tagKey, tagValue string) ([]api.MetricKey, error) {
	return nil, errors.New("Implement me")
}
