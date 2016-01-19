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

package metadata_map

import (
	"sync"

	"github.com/square/metrics/api"
)

type MetadataMap struct {
	Mutex         sync.Mutex
	TagsOfMetric  map[api.MetricKey][]api.TagSet
	AllMetrics    []api.MetricKey
	MetricsForTag map[string]map[string][]api.MetricKey
}

func NewMetadataMap() api.MetricMetadataInterface {
	return &MetadataMap{
		TagsOfMetric:  map[api.MetricKey][]api.TagSet{},
		MetricsForTag: map[string]map[string][]api.MetricKey{},
	}
}

func addUniqueMetricKey(keys []api.MetricKey, newKey api.MetricKey) []api.MetricKey {
	for _, key := range keys {
		if key == newKey {
			return keys
		}
	}
	return append(keys, newKey)
}
func addUniqueTagSet(tagsets []api.TagSet, newTagset api.TagSet) []api.TagSet {
	for _, tagset := range tagsets {
		if tagset.Equals(newTagset) {
			return tagsets
		}
	}
	return append(tagsets, newTagset)
}

func (m *MetadataMap) AddMetric(metric api.TaggedMetric, context api.MetricMetadataAPIContext) error {
	defer context.Profiler.Record("MetadataMap AddMetric")()
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.TagsOfMetric[metric.MetricKey] = addUniqueTagSet(m.TagsOfMetric[metric.MetricKey], metric.TagSet)
	m.AllMetrics = addUniqueMetricKey(m.AllMetrics, metric.MetricKey)
	for key, value := range metric.TagSet {
		if m.MetricsForTag[key] == nil {
			m.MetricsForTag[key] = map[string][]api.MetricKey{}
		}
		m.MetricsForTag[key][value] = addUniqueMetricKey(m.MetricsForTag[key][value], metric.MetricKey)
	}
	return nil
}

func (m *MetadataMap) AddMetrics(metrics []api.TaggedMetric, context api.MetricMetadataAPIContext) error {
	for _, metric := range metrics {
		err := m.AddMetric(metric, context)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MetadataMap) RemoveMetric(metric api.TaggedMetric, context api.MetricMetadataAPIContext) error {
	// TODO: implement this
	return nil
}

func (m *MetadataMap) GetAllTags(metricKey api.MetricKey, context api.MetricMetadataAPIContext) ([]api.TagSet, error) {
	return append([]api.TagSet{}, m.TagsOfMetric[metricKey]...), nil
}

func (m *MetadataMap) GetAllMetrics(context api.MetricMetadataAPIContext) ([]api.MetricKey, error) {
	return append([]api.MetricKey{}, m.AllMetrics...), nil
}

func (m *MetadataMap) GetMetricsForTag(tagKey, tagValue string, context api.MetricMetadataAPIContext) ([]api.MetricKey, error) {
	if lookup := m.MetricsForTag[tagKey]; lookup != nil {
		return append([]api.MetricKey{}, lookup[tagValue]...), nil
	}
	return nil, nil
}
