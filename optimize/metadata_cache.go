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
	"sync"
	"time"

	"github.com/square/metrics/api"
)

func NewMetadataAPICache(metadataAPI api.MetricMetadataAPI, ttl time.Duration) *MetadataAPICache {
	return &MetadataAPICache{
		MetricMetadataAPI: metadataAPI,
		cacheAllTags:      map[api.MetricKey]tagSetExpiry{},
		ttl:               ttl,
		now:               time.Now,
	}
}

type MetadataAPICache struct {
	api.MetricMetadataAPI
	mutex        sync.Mutex
	cacheAllTags map[api.MetricKey]tagSetExpiry
	ttl          time.Duration
	now          func() time.Time
}

type tagSetExpiry struct {
	TagSet  []api.TagSet
	Expires time.Time
}

func (mc *MetadataAPICache) getAllTagsCache(metricKey api.MetricKey) ([]api.TagSet, bool) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	value, ok := mc.cacheAllTags[metricKey]
	if !ok || mc.now().After(value.Expires) {
		return nil, false
	}
	return value.TagSet, true
}
func (mc *MetadataAPICache) storeAllTagsCache(metricKey api.MetricKey, tagSet []api.TagSet) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.cacheAllTags[metricKey] = tagSetExpiry{
		TagSet:  tagSet,
		Expires: mc.now().Add(mc.ttl),
	}
}
func (mc *MetadataAPICache) GetAllTags(metricKey api.MetricKey, context api.MetricMetadataAPIContext) ([]api.TagSet, error) {
	tagset, ok := mc.getAllTagsCache(metricKey)
	if ok {
		return tagset, nil
	}
	tagSet, err := mc.MetricMetadataAPI.GetAllTags(metricKey, context)
	if err != nil {
		return nil, err
	}
	mc.storeAllTagsCache(metricKey, tagSet)
	return tagSet, nil
}
