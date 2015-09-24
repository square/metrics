package optimize

import (
	_ "fmt"
	"sync"
	"time"

	"github.com/square/metrics/api"
)

// The optimization package is designed to be an externalized
// set of behaviors that can be shared across query/function to
// enable selective behaviors that will improve performance.
// The goal is to keep them as implementation-agnostic as possible.

// Optimization configuration has the tunable nobs for
// improving performance.
type OptimizationConfiguration struct {
	EnableMetricMetadataCaching bool
	metricMetadataAPI           api.MetricMetadataAPI
	metricKeyToTagCache         map[api.MetricKey]TagSetCacheEntry
	mutex                       *sync.Mutex
	TimeSourceForNow            TimeSource
	CacheTTL                    time.Duration
}

type TimeSource func() time.Time

type TagSetCacheEntry struct {
	tags      []api.TagSet
	expiresAt time.Time
}

func NewOptimizationConfiguration() *OptimizationConfiguration {
	optimize := OptimizationConfiguration{
		metricKeyToTagCache: make(map[api.MetricKey]TagSetCacheEntry, 3000),
		mutex:               &sync.Mutex{},
		CacheTTL:            time.Hour * 2,
		TimeSourceForNow:    time.Now,
	}
	return &optimize
}

// If we have a cached result for this particular metric then use it,
// otherwise call the supplied update function and cache the result
// transparently.
func (optimize *OptimizationConfiguration) AllTagsCacheHitOrExecute(metric api.MetricKey, update func() ([]api.TagSet, error)) ([]api.TagSet, error) {
	// Just in case we were never initialized.
	if optimize == nil {
		return update()
	}
	// If caching is disabled, always run the provided update function
	if !optimize.EnableMetricMetadataCaching {
		return update()
	}

	tags, cacheHit := optimize.cacheGet(metric)
	if !cacheHit {
		var err error
		if tags, err = update(); err != nil {
			return nil, err
		}

		optimize.cacheUpdate(metric, tags)
		return tags, nil
	}
	return tags, nil
}

func (optimize *OptimizationConfiguration) cacheUpdate(metric api.MetricKey, tags []api.TagSet) {
	optimize.mutex.Lock()
	defer optimize.mutex.Unlock()
	optimize.metricKeyToTagCache[metric] = TagSetCacheEntry{
		tags:      tags,
		expiresAt: time.Now().Add(optimize.CacheTTL),
	}
}

func (optimize *OptimizationConfiguration) cacheGet(metric api.MetricKey) ([]api.TagSet, bool) {
	if val, ok := optimize.metricKeyToTagCache[metric]; ok {
		if optimize.TimeSourceForNow().After(val.expiresAt) {
			return nil, false
		}
		return val.tags, true
	}
	return nil, false
}
