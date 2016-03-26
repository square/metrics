// Copyright 2016 Square Inc.
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
package metadata

import (
	"github.com/square/metrics/api"
	"github.com/square/metrics/inspect"
)

type Context struct {
	Profiler *inspect.Profiler
}

type MetricAPI interface {
	// AddMetric adds the metric to the system.
	AddMetric(metric api.TaggedMetric, context Context) error
	// Bulk metrics addition
	AddMetrics(metric []api.TaggedMetric, context Context) error
	// For a given MetricKey, retrieve all the tagsets associated with it.
	GetAllTags(metricKey api.MetricKey, context Context) ([]api.TagSet, error)
	// GetAllMetrics returns all metrics managed by the system.
	GetAllMetrics(context Context) ([]api.MetricKey, error)
	// For a given tag key-value pair, obtain the list of all the MetricKeys
	// associated with them.
	GetMetricsForTag(tagKey, tagValue string, context Context) ([]api.MetricKey, error)
}
