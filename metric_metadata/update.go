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

import "github.com/square/metrics/api"

// UpdateAPI is an interface for updating metric metadata for indexing in MQE.
type MetricUpdateAPI interface {
	// AddMetric adds the metric to the system.
	AddMetric(metric api.TaggedMetric, context Context) error
	// AddMetrics adds several metrics (possibly more efficiently than one at a time)
	AddMetrics(metric []api.TaggedMetric, context Context) error
	// CheckHealthy checks if this MetricAPI is healthy, returning a possible error
	CheckHealthy() error
}
