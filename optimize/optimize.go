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

package optimize

import (
	"time"

	"github.com/square/metrics/query"
)

// The optimization package is designed to be an externalized
// set of behaviors that can be shared across query/function to
// enable selective behaviors that will improve performance.
// The goal is to keep them as implementation-agnostic as possible.

// Optimization configuration has the tunable nobs for
// improving performance.
type OptimizationConfiguration struct {
	MetadataCacheTTL                   time.Duration
	FetchTimeseriesStorageConcurrently bool
	MaximumConcurrentFetches           int
}

func NewOptimizationConfiguration() OptimizationConfiguration {
	return OptimizationConfiguration{
		MetadataCacheTTL:                   2 * time.Hour,
		FetchTimeseriesStorageConcurrently: true,
		MaximumConcurrentFetches:           10,
	}
}

func (config OptimizationConfiguration) OptimizeExecutionContext(e query.ExecutionContext) query.ExecutionContext {
	if config.MetadataCacheTTL > 0 {
		e.MetricMetadataAPI = NewMetadataAPICache(e.MetricMetadataAPI, config.MetadataCacheTTL)
	}
	if config.FetchTimeseriesStorageConcurrently {
		e.TimeseriesStorageAPI = NewParallelTimeseriesStorageAPI(config.MaximumConcurrentFetches, e.TimeseriesStorageAPI)
	}
	return e
}
