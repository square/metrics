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

package cached

import (
	"sync"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/metric_metadata"
)

// CachedMetricMetadataAPI caches some of the metadata associated with the API to reduce latency.
// However, it does not reduce total QPS: whenever it reads from the cache, it performs an update
// in the background by launching a new goroutine.
type CachedMetricMetadataAPI struct {
	metricMetadataAPI metadata.MetricAPI                 // The internal MetricMetadataAPI that performs the actual queries.
	getAllTagsCache   map[api.MetricKey]CachedTagSetList // The cache of "tags from metric name"
	timeToLive        time.Duration                      // Time for cache entries to survive.
	mutex             sync.Mutex                         // Synchronizing mutex

	backgroundQueue chan func(metadata.Context) error // A channel that holds background requests.
}

// Config stores data needed to instantiate a CachedMetricMetadataAPI.
type Config struct {
	TimeToLive   time.Duration
	RequestLimit int
}

// NewMetricMetadataAPI creates a cached API given configuration and an underlying API object.
func NewMetricMetadataAPI(apiInstance metadata.MetricAPI, config Config) *CachedMetricMetadataAPI {
	requests := make(chan func(metadata.Context) error, config.RequestLimit)
	return &CachedMetricMetadataAPI{
		metricMetadataAPI: apiInstance,
		getAllTagsCache:   map[api.MetricKey]CachedTagSetList{},
		timeToLive:        config.TimeToLive,
		backgroundQueue:   requests,
	}
}

// CachedTagSetList is an item in the cache.
type CachedTagSetList struct {
	TagSets []api.TagSet // The tagsets for this metric
	Expiry  time.Time    // The time at which the cache entry expires
}

// Expired tells whether the entry has zero-time (meaning absent) or is out-of-date.
func (c CachedTagSetList) Expired() bool {
	return c.Expiry.IsZero() || c.Expiry.Before(time.Now())
}

func (c *CachedMetricMetadataAPI) addBackgroundGetAllTagsRequest(metric api.MetricKey) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if len(c.backgroundQueue) < cap(c.backgroundQueue) {
		// Add this function to the background queue.
		c.backgroundQueue <- func(context metadata.Context) error {
			defer context.Profiler.Record("CachedMetricMetadataAPI_BackgroundAction_GetAllTags")()
			startTime := time.Now()
			tagsets, err := c.metricMetadataAPI.GetAllTags(metric, context)
			if err != nil {
				return err
			}
			c.replaceCachedTagSet(metric, tagsets, startTime)
			return nil
		}
	}
}

// GetBackgroundAction is a blocking method that runs one queued cache update.
// It will block until an update is available.
func (c *CachedMetricMetadataAPI) GetBackgroundAction() func(metadata.Context) error {
	return <-c.backgroundQueue
}

// AddMetric waits for a slot to be open, then queries the underlying API.
func (c *CachedMetricMetadataAPI) AddMetric(metric api.TaggedMetric, context metadata.Context) error {
	return c.metricMetadataAPI.AddMetric(metric, context)
}

// AddMetrics waits for a slot to be open, then queries the underlying API.
func (c *CachedMetricMetadataAPI) AddMetrics(metrics []api.TaggedMetric, context metadata.Context) error {
	return c.metricMetadataAPI.AddMetrics(metrics, context)
}

// GetAllMetrics waits for a slot to be open, then queries the underlying API.
func (c *CachedMetricMetadataAPI) GetAllMetrics(context metadata.Context) ([]api.MetricKey, error) {
	return c.metricMetadataAPI.GetAllMetrics(context)
}

// GetMetricsForTag wwaits for a slot to be open, then queries the underlying API.
func (c *CachedMetricMetadataAPI) GetMetricsForTag(tagKey, tagValue string, context metadata.Context) ([]api.MetricKey, error) {
	return c.metricMetadataAPI.GetMetricsForTag(tagKey, tagValue, context)
}

// getCachedTagSet is a thread-safe way to get the cached data for a metric (protected by a mutex)
func (c *CachedMetricMetadataAPI) getCachedTagSet(metricKey api.MetricKey) CachedTagSetList {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.getAllTagsCache[metricKey]
}

// setCachedTagSet is a thread-safe way to assign cached data (protected by a mutex).
// if the given value is less recent than the stored value, nothing happens.
func (c *CachedMetricMetadataAPI) replaceCachedTagSet(metricKey api.MetricKey, tagsets []api.TagSet, startTime time.Time) {
	value := CachedTagSetList{
		TagSets: tagsets,
		Expiry:  startTime.Add(c.timeToLive),
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.getAllTagsCache[metricKey].Expiry.Before(value.Expiry) {
		c.getAllTagsCache[metricKey] = value
	}
}

// getAllTagsRaw performs a request to the underlying API and updates the cached values with whatever is returned.
// Then it returns the result to the caller.
// Note that getAllTagsRaw does not respect the request limit behavior- this must be enforced by the caller.
func (c *CachedMetricMetadataAPI) getAllTagsRaw(metricKey api.MetricKey, context metadata.Context) ([]api.TagSet, error) {
	startTime := time.Now()
	tagsets, err := c.metricMetadataAPI.GetAllTags(metricKey, context)
	if err != nil {
		return nil, err
	}
	c.replaceCachedTagSet(metricKey, tagsets, startTime)
	return tagsets, nil
}

// GetAllTags uses the cache to serve tag data for the given metric.
// If the cache entry is missing or out of date, it uses the results of a query to the underlying API to return to the caller.
// Even if the cache entry is up-to-date, this method performs a background request to the underlying API to keep the cache fresh.
func (c *CachedMetricMetadataAPI) GetAllTags(metricKey api.MetricKey, context metadata.Context) ([]api.TagSet, error) {
	defer context.Profiler.Record("CachedMetricMetadataAPI_GetAllTags")()
	// Get the cached result for this metric.
	item := c.getCachedTagSet(metricKey)

	if item.Expired() {
		defer context.Profiler.Record("CachedMetricMetadataAPI_GetAllTags_Expired")()
		startTime := time.Now()
		tagsets, err := c.metricMetadataAPI.GetAllTags(metricKey, context)
		if err != nil {
			return nil, err
		}
		c.replaceCachedTagSet(metricKey, tagsets, startTime)
		return tagsets, nil
	}

	defer context.Profiler.Record("CachedMetricMetadataAPI_Hit")()

	c.addBackgroundGetAllTagsRequest(metricKey)

	// but return the cached result immediately.
	return item.TagSets, nil
}

func (c *CachedMetricMetadataAPI) CurrentLiveRequests() int {
	return len(c.backgroundQueue)
}
func (c *CachedMetricMetadataAPI) MaximumLiveRequests() int {
	return cap(c.backgroundQueue)
}
