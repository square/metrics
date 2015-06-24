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

// Package api holds common data types and public interface exposed by the indexer library.
// Refer to the doc
// https://docs.google.com/a/squareup.com/document/d/1k0Wgi2wnJPQoyDyReb9dyIqRrD8-v0u8hz37S282ii4/edit
// for the terminology.
package api

import "github.com/square/metrics/inspect"

// API is the set of public methods exposed by the indexer library.
type API interface {
	// AddMetrics adds the metrics to the system.
	AddMetrics(metric []TaggedMetric) error

	// RemoveMetric removes the metric from the system.
	RemoveMetric(metric TaggedMetric) error

	// Convert the given tag-based metric name to graphite metric name,
	// using the configured rules. May error out.
	ToGraphiteName(metric TaggedMetric) (GraphiteMetric, error)

	// Converts the given graphite metric to the tag-based meric,
	// using the configured rules. May error out.
	ToTaggedName(metric GraphiteMetric) (TaggedMetric, error)

	// For a given MetricKey, retrieve all the tagsets associated with it.
	GetAllTags(metricKey MetricKey) ([]TagSet, error)

	// GetAllMetrics returns all metrics managed by the system.
	GetAllMetrics() ([]MetricKey, error)

	// For a given tag key-value pair, obtain the list of all the MetricKeys
	// associated with them.
	GetMetricsForTag(tagKey, tagValue string) ([]MetricKey, error)
}

// Configuration is the struct that tells how to instantiate a new copy of an API.
type Config struct {
	ConversionRulesPath string `yaml:"conversion_rules_path"` // Location of the rule yaml file.

	// Database configurations
	// mostly cassandra configurations from
	// https://github.com/gocql/gocql/blob/master/cluster.go
	Hosts    []string `yaml:"hosts"`
	Keyspace string   `yaml:"keyspace"`
}

// ProfilingAPI wraps an ordinary API and also records profiling metrics to a given Profiler object.
type ProfilingAPI struct {
	Profiler *inspect.Profiler
	API      API
}

func (api ProfilingAPI) AddMetrics(metrics []TaggedMetric) error {
	defer api.Profiler.Record("api.AddMetrics")()
	return api.API.AddMetrics(metrics)
}
func (api ProfilingAPI) RemoveMetric(metric TaggedMetric) error {
	defer api.Profiler.Record("api.RemoveMetric")()
	return api.API.RemoveMetric(metric)
}
func (api ProfilingAPI) ToGraphiteName(metric TaggedMetric) (GraphiteMetric, error) {
	defer api.Profiler.Record("api.ToGraphiteName")()
	return api.API.ToGraphiteName(metric)
}
func (api ProfilingAPI) ToTaggedName(metric GraphiteMetric) (TaggedMetric, error) {
	defer api.Profiler.Record("api.ToTaggedName")()
	return api.API.ToTaggedName(metric)
}
func (api ProfilingAPI) GetAllTags(metricKey MetricKey) ([]TagSet, error) {
	defer api.Profiler.Record("api.GetAllTags")()
	return api.API.GetAllTags(metricKey)
}
func (api ProfilingAPI) GetAllMetrics() ([]MetricKey, error) {
	defer api.Profiler.Record("api.GetAllMetrics")()
	return api.API.GetAllMetrics()
}
func (api ProfilingAPI) GetMetricsForTag(tagKey, tagValue string) ([]MetricKey, error) {
	defer api.Profiler.Record("api.GetMetricsForTag")()
	return api.API.GetMetricsForTag(tagKey, tagValue)
}
