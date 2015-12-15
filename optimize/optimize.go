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
	MetadataCacheTTL time.Duration
}

func NewOptimizationConfiguration() OptimizationConfiguration {
	return OptimizationConfiguration{
		MetadataCacheTTL: 2 * time.Hour,
	}
}

func (config OptimizationConfiguration) OptimizeExecutionContext(e query.ExecutionContext) query.ExecutionContext {
	if config.MetadataCacheTTL > 0 {
		e.MetricMetadataAPI = NewMetadataAPICache(e.MetricMetadataAPI, config.MetadataCacheTTL)
	}
	return e
}
