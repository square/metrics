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
	// AddMetric adds the metric to the system.
	AddMetric(metric TaggedMetric) error

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

	// Allow the API to store and retrieve graphite metrics:
	// Add a Graphite metric to the complete list (as needed)
	AddGraphiteMetric(metric GraphiteMetric) error

	// Get all the Graphite metrics
	GetAllGraphiteMetrics() ([]GraphiteMetric, error)
}

// Configuration is the struct that tells how to instantiate a new copy of an API.
type Config struct {
	// Location of conversion rules. All *.yaml files in here will be loaded.
	ConversionRulesPath string `yaml:"conversion_rules_path"`

	// Database configurations
	// mostly cassandra configurations from
	// https://github.com/gocql/gocql/blob/master/cluster.go
	Hosts    []string       `yaml:"hosts"`
	Keyspace string         `yaml:"keyspace"`
	Database DatabaseConfig `yaml:"database"`
}

// Configuration for the Database attached to the default API
type DatabaseConfig struct {
	GraphiteMetricTTL int `yaml:graphite_metric_ttl` // in seconds
}

// ProfilingAPI wraps an ordinary API and also records profiling metrics to a given Profiler object.
type ProfilingAPI struct {
	Profiler *inspect.Profiler
	API      API
}

func (api ProfilingAPI) AddMetric(metric TaggedMetric) error {
	defer api.Profiler.Record("api.AddMetric")()
	return api.API.AddMetric(metric)
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

const SpecialGraphiteName = "$graphite"

func (api ProfilingAPI) AddGraphiteMetric(metric GraphiteMetric) error {
	defer api.Profiler.Record("api.AddGraphiteMetric")()
	return api.API.AddGraphiteMetric(metric)
}
func (api ProfilingAPI) GetAllGraphiteMetrics() ([]GraphiteMetric, error) {
	defer api.Profiler.Record("api.GetAllGraphiteMetrics")()
	return api.API.GetAllGraphiteMetrics()
}
