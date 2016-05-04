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

package cached

import (
	"errors"
	"sync"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/log"
	"github.com/square/metrics/metric_metadata"
	"github.com/square/metrics/util"
)

// CachedMetricMetadataAPI caches some of the metadata associated with the API to reduce latency.
// However, it does not reduce total QPS: whenever it reads from the cache, it performs an update
// in the background by launching a new goroutine.
type CachedMetricMetadataAPI struct {
	metricMetadataAPI metadata.MetricAPI // The internal MetricAPI that performs the actual queries.
	clock             util.Clock         // Here so we can mock out in tests

	// Cached items
	getAllTagsCache      map[api.MetricKey]*CachedTagSetList // The cache of metric -> tags
	getAllTagsCacheMutex sync.RWMutex                        // Mutex for getAllTagsCache

	// Cache Config
	freshness  time.Duration // How long until cache entries become stale
	timeToLive time.Duration // How long until cache entries become expired

	// Queue
	backgroundQueue chan func(metadata.Context) error // A channel that holds background requests.
	queueMutex      sync.Mutex                        // Synchronizing mutex for the queue
}

// Config stores data needed to instantiate a CachedMetricMetadataAPI.
type Config struct {
	Freshness    time.Duration
	RequestLimit int
	TimeToLive   time.Duration
}

// CachedTagSetList is an item in the cache.
type CachedTagSetList struct {
	TagSets []api.TagSet // The tagsets for this metric
	Expiry  time.Time    // The time at which the cache entry expires
	Stale   time.Time    // The time at which the cache entry becomes stale

	sync.Mutex // Synchronizing mutex

	inflight bool           // Indicates a request is already in flight
	enqueued bool           // Indicates a request has been enqueued
	wg       sync.WaitGroup // Synchronizing wait group

	fetchError error // Fetch error from the last attempt
}

// NewMetricMetadataAPI creates a cached API given configuration and an underlying API object.
func NewMetricMetadataAPI(apiInstance metadata.MetricAPI, config Config) *CachedMetricMetadataAPI {
	requests := make(chan func(metadata.Context) error, config.RequestLimit)
	if config.Freshness == 0 {
		config.Freshness = config.TimeToLive
	}
	return &CachedMetricMetadataAPI{
		metricMetadataAPI: apiInstance,
		clock:             util.RealClock{},
		getAllTagsCache:   map[api.MetricKey]*CachedTagSetList{},
		freshness:         config.Freshness,
		timeToLive:        config.TimeToLive,
		backgroundQueue:   requests,
	}
}

// addBackgroundGetAllTagsRequest adds a job to update the lag list for the given
// metric. Requires the caller hold the lock for the item in the cache.
func (c *CachedMetricMetadataAPI) addBackgroundGetAllTagsRequest(item *CachedTagSetList, metricKey api.MetricKey) {
	if item == nil {
		log.Errorf("Asked to perform a background GetAllTags lookup for %s but missing entry", metricKey)
		return
	}

	c.queueMutex.Lock()
	defer c.queueMutex.Unlock()

	if cap(c.backgroundQueue) <= len(c.backgroundQueue) {
		log.Warningf("Unable to enqueue a background GetAllTags lookup for %s due to a full queue", metricKey)
		return
	}

	if item.enqueued {
		log.Infof("Unable to perform a background GetAllTags lookup for %s as one is already enqueued", metricKey)
		return
	}

	if item.inflight {
		log.Infof("Unable to perform a background GetAllTags lookup for %s as one is already in flight", metricKey)
		return
	}

	log.Infof("Enqueuing a background GetAllTags lookup for %s", metricKey)
	item.enqueued = true

	c.backgroundQueue <- func(context metadata.Context) error {
		log.Infof("Executing the background GetAllTags lookup for %s", metricKey)
		defer log.Infof("Finished the background GetAllTags lookup for %s", metricKey)

		item.Lock()
		defer item.Unlock()
		item.enqueued = false

		defer context.Profiler.Record("CachedMetricMetadataAPI_BackgroundAction_GetAllTags")()

		_, err := c.fetchAndUpdateCachedTagSet(item, metricKey, context)
		return err
	}
}

// GetBackgroundAction is a blocking method that runs one queued cache update.
// It will block until an update is available.
func (c *CachedMetricMetadataAPI) GetBackgroundAction() func(metadata.Context) error {
	return <-c.backgroundQueue
}

// GetAllMetrics waits for a slot to be open, then queries the underlying API.
func (c *CachedMetricMetadataAPI) GetAllMetrics(context metadata.Context) ([]api.MetricKey, error) {
	return c.metricMetadataAPI.GetAllMetrics(context)
}

// GetMetricsForTag wwaits for a slot to be open, then queries the underlying API.
func (c *CachedMetricMetadataAPI) GetMetricsForTag(tagKey, tagValue string, context metadata.Context) ([]api.MetricKey, error) {
	return c.metricMetadataAPI.GetMetricsForTag(tagKey, tagValue, context)
}

// CheckHealthy checks if the underlying MetricAPI is healthy
func (c *CachedMetricMetadataAPI) CheckHealthy() error {
	return c.metricMetadataAPI.CheckHealthy()
}

// fetchAndUpdateCachedTagSet updates the in-memory cache (asusming the update
// is newer than what is in the cache). Requires the caller hold the lock for the
// item in the cache.
func (c *CachedMetricMetadataAPI) fetchAndUpdateCachedTagSet(item *CachedTagSetList, metricKey api.MetricKey, context metadata.Context) ([]api.TagSet, error) {
	if item == nil {
		return nil, errors.New("Missing cache list entry")
	}

	item.wg.Add(1)
	item.fetchError = nil
	item.inflight = true
	item.Unlock()

	startTime := c.clock.Now()
	tagsets, err := c.metricMetadataAPI.GetAllTags(metricKey, context)

	item.Lock()

	if err != nil {
		item.fetchError = err
		item.wg.Done()
		item.inflight = false

		return nil, err
	}

	// Only update the cache if the update expires later than the current
	// entry in the cache
	newExpiry := startTime.Add(c.timeToLive)
	if item.Expiry.Before(newExpiry) {
		item.TagSets = tagsets
		item.Expiry = newExpiry
		item.Stale = startTime.Add(c.freshness)
	} else {
		log.Warningf("Asked to update the tag set for %s but new expiry is earlier than current (%s vs %s)",
			metricKey, newExpiry.String(), item.Expiry.String())
	}

	item.wg.Done()
	item.inflight = false

	return tagsets, nil
}

// GetAllTags uses the cache to serve tag data for the given metric.
// If the cache entry is missing or out of date, it uses the results of a query
// to the underlying API to return to the caller. Even if the cache entry is
// up-to-date, this method may enqueue a background request to the underlying API
// to keep the cache fresh.
func (c *CachedMetricMetadataAPI) GetAllTags(metricKey api.MetricKey, context metadata.Context) ([]api.TagSet, error) {
	defer context.Profiler.Record("CachedMetricMetadataAPI_GetAllTags")()

	// Get the cached result for this metric.
	c.getAllTagsCacheMutex.RLock()
	item, ok := c.getAllTagsCache[metricKey]
	c.getAllTagsCacheMutex.RUnlock()

	if !ok {
		c.getAllTagsCacheMutex.Lock()

		// Now that we have the mutex for getAllTagsCache, make sure another goroutine
		// hasn't already updated the cache
		item, ok = c.getAllTagsCache[metricKey]
		if !ok {
			item = &CachedTagSetList{}
			c.getAllTagsCache[metricKey] = item
		}

		c.getAllTagsCacheMutex.Unlock()
	}

	item.Lock()

	if item.Expiry.IsZero() || item.Expiry.Before(c.clock.Now()) {
		if item.inflight {
			item.Unlock()
			item.wg.Wait()

			// Make sure we have the lock to re-read
			item.Lock()
			defer item.Unlock()

			// If the request we were waiting on errored, we also errored
			return item.TagSets, item.fetchError
		}

		defer item.Unlock()

		// We're going to execute this fetch now
		defer context.Profiler.Record("CachedMetricMetadataAPI_GetAllTags_Expired")()

		tagsets, err := c.fetchAndUpdateCachedTagSet(item, metricKey, context)
		if err != nil {
			defer context.Profiler.Record("CachedMetricMetadataAPI_GetAllTags_Errored")()
			return nil, err
		}

		return tagsets, nil
	}

	defer context.Profiler.Record("CachedMetricMetadataAPI_Hit")()
	defer item.Unlock()

	// Otherwise, we could be stale
	if item.Stale.Before(c.clock.Now()) {
		// Enqueue a background request
		c.addBackgroundGetAllTagsRequest(item, metricKey)
	}

	// but return the cached result immediately.
	return item.TagSets, nil
}

// CurrentLiveRequests returns the number of requests currently in the queue
func (c *CachedMetricMetadataAPI) CurrentLiveRequests() int {
	return len(c.backgroundQueue)
}

// MaximumLiveRequests returns the maximum number of requests that can be in the queue
func (c *CachedMetricMetadataAPI) MaximumLiveRequests() int {
	return cap(c.backgroundQueue)
}
