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

// Package internal holds classes used in internal implementation of metric-indexer
package internal

import (
	"sync"

	"github.com/gocql/gocql"
	"github.com/square/metrics/api"
)

// Database represents internal connection to Cassandra.
type Database interface {
	// Insertion Methods
	// -----------------
	AddMetricName(metricKey api.MetricKey, metric api.TagSet) error
	AddToTagIndex(tagKey, tagValue string, metricKey api.MetricKey) error

	// Query methods
	// -------------
	GetTagSet(metricKey api.MetricKey) ([]api.TagSet, error)
	GetMetricKeys(tagKey, tagValue string) ([]api.MetricKey, error)
	GetAllMetrics() ([]api.MetricKey, error)

	// Deletion Method
	// ---------------
	RemoveMetricName(metricKey api.MetricKey, tagSet api.TagSet) error
	RemoveFromTagIndex(tagKey, tagValue string, metricKey api.MetricKey) error
}

type tagIndexCacheKey struct {
	key    string
	value  string
	metric api.MetricKey
}

type defaultDatabase struct {
	session         *gocql.Session
	allMetricsCache map[api.MetricKey]bool
	allMetricsMutex *sync.RWMutex
	tagIndexCache   map[tagIndexCacheKey]bool
	tagIndexMutex   *sync.RWMutex
}

// NewCassandraDatabase creates an instance of database, backed by Cassandra.
func NewCassandraDatabase(clusterConfig *gocql.ClusterConfig) (Database, error) {
	session, err := clusterConfig.CreateSession()
	if err != nil {
		return nil, err
	}
	return &defaultDatabase{
		session:         session,
		allMetricsCache: make(map[api.MetricKey]bool),
		allMetricsMutex: &sync.RWMutex{},
		tagIndexCache:   make(map[tagIndexCacheKey]bool),
		tagIndexMutex:   &sync.RWMutex{},
	}, nil
}

func eitherError(a error, b error) error {
	if a != nil {
		return a
	}
	return b
}

// AddMetricName inserts to metric to Cassandra.
func (db *defaultDatabase) AddMetricName(metricKey api.MetricKey, tagSet api.TagSet) error {

	if err := db.session.Query("INSERT INTO metric_names (metric_key, tag_set) VALUES (?, ?)", metricKey, tagSet.Serialize()).Exec(); err != nil {
		return err
	}
	db.allMetricsMutex.RLock()
	if db.allMetricsCache[metricKey] {
		db.allMetricsMutex.RUnlock()
		// If the key is found in the cache, exit early.
		return nil
	}
	db.allMetricsMutex.RUnlock()
	if err := db.session.Query("UPDATE metric_name_set SET metric_names = metric_names + ? WHERE shard = ?", []string{string(metricKey)}, 0).Exec(); err != nil {
		return err
	}
	db.allMetricsMutex.Lock()
	// Remember the cached value so that it won't be written again in the absence of reads.
	db.allMetricsCache[metricKey] = true
	db.allMetricsMutex.Unlock()
	return nil

}

func (db *defaultDatabase) AddToTagIndex(tagKey string, tagValue string, metricKey api.MetricKey) error {
	indexKey := tagIndexCacheKey{tagKey, tagValue, metricKey}
	db.tagIndexMutex.RLock()
	indexValue := db.tagIndexCache[indexKey]
	db.tagIndexMutex.RUnlock()
	if indexValue {
		return nil // Found in the cache so already in the table, so no need to perform a write.
	}
	err := db.session.Query(
		"UPDATE tag_index SET metric_keys = metric_keys + ? WHERE tag_key = ? AND tag_value = ?",
		[]string{string(metricKey)},
		tagKey,
		tagValue,
	).Exec()
	if err == nil {
		db.tagIndexMutex.Lock()
		// Remember this write in the cache.
		db.tagIndexCache[indexKey] = true
		db.tagIndexMutex.Unlock()
	}
	return err
}

func (db *defaultDatabase) GetTagSet(metricKey api.MetricKey) ([]api.TagSet, error) {
	var tags []api.TagSet
	rawTag := ""
	iterator := db.session.Query(
		"SELECT tag_set FROM metric_names WHERE metric_key = ?",
		metricKey,
	).Iter()
	for iterator.Scan(&rawTag) {
		parsedTagSet := api.ParseTagSet(rawTag)
		if parsedTagSet != nil {
			tags = append(tags, parsedTagSet)
		}
	}
	if err := iterator.Close(); err != nil {
		return nil, err
	}
	return tags, nil
}

func (db *defaultDatabase) GetMetricKeys(tagKey string, tagValue string) ([]api.MetricKey, error) {
	var keys []api.MetricKey
	err := db.session.Query(
		"SELECT metric_keys FROM tag_index WHERE tag_key = ? AND tag_value = ?",
		tagKey,
		tagValue,
	).Scan(&keys)
	if err == gocql.ErrNotFound {
		return keys, nil
	}
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (db *defaultDatabase) GetAllMetrics() ([]api.MetricKey, error) {
	var keys []api.MetricKey
	err := db.session.Query("SELECT metric_names FROM metric_name_set WHERE shard = ?", 0).Scan(&keys)
	if err != nil {
		return nil, err
	}
	db.allMetricsMutex.Lock()
	for _, key := range keys {
		db.allMetricsCache[key] = true
	}
	db.allMetricsMutex.Unlock()
	return keys, nil
}

func (db *defaultDatabase) RemoveMetricName(metricKey api.MetricKey, tagSet api.TagSet) error {
	db.allMetricsMutex.Lock()
	// Forget the metric in the cache.
	// (If this delete fails, there will be an extraneous write the next time the metric is consumed).
	db.allMetricsCache[metricKey] = false
	db.allMetricsMutex.Unlock()
	return db.session.Query(
		"DELETE FROM metric_names WHERE metric_key = ? AND tag_set = ?",
		metricKey,
		tagSet.Serialize(),
	).Exec()
}

func (db *defaultDatabase) RemoveFromTagIndex(tagKey string, tagValue string, metricKey api.MetricKey) error {
	// Forget the tag key/value/metric triplet in the cache.
	// (If this delete fails, there will be an extraneous write the next time they are consumed).
	db.tagIndexMutex.Lock()
	db.tagIndexCache[tagIndexCacheKey{tagKey, tagValue, metricKey}] = false
	db.tagIndexMutex.Unlock()
	return db.session.Query(
		"UPDATE tag_index SET metric_keys = metric_keys - ? WHERE tag_key = ? AND tag_value = ?",
		[]string{string(metricKey)},
		tagKey,
		tagValue,
	).Exec()
}
