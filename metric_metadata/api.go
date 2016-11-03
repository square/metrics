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

// Package metadata holds the interface for accessing metadata for indexing metrics.
package metadata

import (
	"github.com/square/metrics/api"
	"github.com/square/metrics/inspect"
)

// Context holds contextual information for performing MetricAPI queries.
type Context struct {
	// Profiler is used to record execution time for metadata queries.
	Profiler *inspect.Profiler
}

// MetricAPI is an interface for obtaining metric metadata for indexing in MQE.
type MetricAPI interface {
	// GetAllTags takes a MetricKey and retrieves all the tagsets associated with it.
	GetAllTags(metricKey api.MetricKey, context Context) ([]api.TagSet, error)
	// GetAllAvailableTags returns a list of tags indexed in the database
	GetAllAvailableTags(context Context) (map[string][]string, error)
	// GetAllMetrics returns all metrics managed by the system.
	GetAllMetrics(context Context) ([]api.MetricKey, error)
	//GetAllTagSets geta all the tagsets
	GetAllTagSets(context Context) ([]api.TagSetInfo, error)
	// GetMetricsForTag takes a tag key-value pair and returnsthe list of all the
	// MetricKeys associated with them.
	GetMetricsForTag(tagKey, tagValue string, context Context) ([]api.MetricKey, error)
	// CheckHealthy checks if this MetricAPI is healthy, returning a possible error
	CheckHealthy() error
}
