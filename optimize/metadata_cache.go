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

// TODO: if this kind of repeated "embedded-but-with-different-behavior-on-a-few-methods" struct is defined in more places,
// then combine them via actual embedding.
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
