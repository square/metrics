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

package cached_metadata

import (
	"fmt"
	"sync"
	"time"

	"github.com/square/metrics/api"
)

// CachedMetricMetadataAPI caches some of the metadata associated with the API to reduce latency.
// However, it does not reduce total QPS: whenever it reads from the cache, it performs an update
// in the background by launching a new goroutine.
type CachedMetricMetadataAPI struct {
	metricMetadataAPI api.MetricMetadataAPI              // The internal MetricMetadataAPI that performs the actual queries.
	getAllTagsCache   map[api.MetricKey]CachedTagSetList // The cache of "tags from metric name"
	timeToLive        time.Duration                      // Time for cache entries to survive.
	mutex             sync.Mutex                         // Synchronizing mutex

	requestQueue chan struct{} // A channel that synchronizes the in-flight requests.
}

// Config stores data needed to instantiate a CachedMetricMetadataAPI.
type Config struct {
	TimeToLive   time.Duration
	RequestLimit int
}

// NewCachedMetricMetadataAPI creates a cached API given configuration and an underlying API object.
func NewCachedMetricMetadataAPI(metadata api.MetricMetadataAPI, config Config) *CachedMetricMetadataAPI {
	requests := make(chan struct{}, config.RequestLimit)
	// Fill the requests
	for i := 0; i < config.RequestLimit; i++ {
		requests <- struct{}{}
	}
	return &CachedMetricMetadataAPI{
		metricMetadataAPI: metadata,
		getAllTagsCache:   map[api.MetricKey]CachedTagSetList{},
		timeToLive:        config.TimeToLive,
		requestQueue:      requests,
	}
}

// An item in the cache.
type CachedTagSetList struct {
	TagSets []api.TagSet // The tagsets for this metric
	Expiry  time.Time    // The time at which the cache entry expires
}

// Expired tells whether the entry has zero-time (meaning absent) or is out-of-date.
func (c CachedTagSetList) Expired() bool {
	return c.Expiry.IsZero() || c.Expiry.Before(time.Now())
}

// startRequest waits until a slot is open to make a request to the underlying API.
func (c *CachedMetricMetadataAPI) startRequest() {
	<-c.requestQueue // Wait for a request to open.
}

// finishRequest signals that the request to the underlying API is done,
// which allows another request to be performed.
func (c *CachedMetricMetadataAPI) finishRequest() {
	c.requestQueue <- struct{}{}
}

// AddMetric waits for a slot to be open, then queries the underlying API.
func (c *CachedMetricMetadataAPI) AddMetric(metric api.TaggedMetric, context api.MetricMetadataAPIContext) error {
	c.startRequest()
	defer c.finishRequest()

	return c.metricMetadataAPI.AddMetric(metric, context)
}

// AddMetrics waits for a slot to be open, then queries the underlying API.
func (c *CachedMetricMetadataAPI) AddMetrics(metrics []api.TaggedMetric, context api.MetricMetadataAPIContext) error {
	c.startRequest()
	defer c.finishRequest()

	return c.metricMetadataAPI.AddMetrics(metrics, context)
}

// GetAllMetrics waits for a slot to be open, then queries the underlying API.
func (c *CachedMetricMetadataAPI) GetAllMetrics(context api.MetricMetadataAPIContext) ([]api.MetricKey, error) {
	c.startRequest()
	defer c.finishRequest()

	return c.metricMetadataAPI.GetAllMetrics(context)
}

// GetMetricsForTag wwaits for a slot to be open, then queries the underlying API.
func (c *CachedMetricMetadataAPI) GetMetricsForTag(tagKey, tagValue string, context api.MetricMetadataAPIContext) ([]api.MetricKey, error) {
	c.startRequest()
	defer c.finishRequest()

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
func (c *CachedMetricMetadataAPI) replaceCachedTagSet(metricKey api.MetricKey, value CachedTagSetList) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.getAllTagsCache[metricKey].Expiry.Before(value.Expiry) {
		c.getAllTagsCache[metricKey] = value
	}
}

// getAllTagsRaw performs a request to the underlying API and updates the cached values with whatever is returned.
// Then it returns the result to the caller.
// Note that getAllTagsRaw does not respect the request limit behavior- this must be enforced by the caller.
func (c *CachedMetricMetadataAPI) getAllTagsRaw(metricKey api.MetricKey, context api.MetricMetadataAPIContext) ([]api.TagSet, error) {
	tagsets, err := c.metricMetadataAPI.GetAllTags(metricKey, context)
	if err != nil {
		return nil, err
	}
	c.replaceCachedTagSet(metricKey, CachedTagSetList{
		TagSets: tagsets,
		Expiry:  time.Now().Add(c.timeToLive),
	})
	return tagsets, nil
}

// getAllTagsAndUpdateCache performs a fetch and updates the cache correspondingly.
// If `mandatory` is false and the request limit has been reached, getAllTagsAndUpdateCache will return an error
// without performing a query on the underlying API.
// This gives priority to updating out-of-date entries at the expense of sometimes discarding background updates to the cache.
func (c *CachedMetricMetadataAPI) getAllTagsAndUpdateCache(metricKey api.MetricKey, context api.MetricMetadataAPIContext, mandatory bool) ([]api.TagSet, error) {
	if mandatory {
		c.startRequest()
		defer c.finishRequest()
	} else {
		select {
		case <-c.requestQueue:
			// A request is available, so run but make sure to put it back.
			defer c.finishRequest() // Once this function finishes, restore request state.
		default:
			return nil, fmt.Errorf("Pressure on API through cache is too high (%d simultaneous requests allowed)- dropping non-mandatory request.", cap(c.requestQueue))
		}
	}

	return c.getAllTagsRaw(metricKey, context)
}

// GetAllTags uses the cache to serve tag data for the given metric.
// If the cache entry is missing or out of date, it uses the results of a query to the underlying API to return to the caller.
// Even if the cache entry is up-to-date, this method performs a background request to the underlying API to keep the cache fresh.
func (c *CachedMetricMetadataAPI) GetAllTags(metricKey api.MetricKey, context api.MetricMetadataAPIContext) ([]api.TagSet, error) {
	defer context.Profiler.Record("CachedMetricMetadataAPI_GetAllTags")()
	// Get the cached result for this metric.
	item := c.getCachedTagSet(metricKey)

	if item.Expired() {
		defer context.Profiler.Record("CachedMetricMetadataAPI_Miss")()
		// The item was expired, so query the metadata API.
		return c.getAllTagsAndUpdateCache(metricKey, context, true)
	}

	defer context.Profiler.Record("CachedMetricMetadataAPI_Hit")()

	// Update the cache in the background,
	go c.getAllTagsAndUpdateCache(metricKey, context, false)

	// but return the cached result immediately.
	return item.TagSets, nil
}

func (c *CachedMetricMetadataAPI) CurrentLiveRequests() int {
	return cap(c.requestQueue) - len(c.requestQueue)
}
func (c *CachedMetricMetadataAPI) MaximumLiveRequests() int {
	return cap(c.requestQueue)
}
