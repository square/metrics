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
	"github.com/gocql/gocql"
	"github.com/square/metrics/api"
)

// Database represents internal connection to Cassandra.
type Database interface {
	// Insertion Methods
	// -----------------
	AddMetricName(metricKey api.MetricKey, metric api.TagSet) error
	AddMetricNames(metrics []api.TaggedMetric) error
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
	session *gocql.Session
}

// NewCassandraDatabase creates an instance of database, backed by Cassandra.
func NewCassandraDatabase(clusterConfig *gocql.ClusterConfig) (Database, error) {
	session, err := clusterConfig.CreateSession()
	if err != nil {
		return nil, err
	}
	return &defaultDatabase{
		session: session,
	}, nil
}

// AddMetricName inserts to metric to Cassandra.
func (db *defaultDatabase) AddMetricName(metricKey api.MetricKey, tagSet api.TagSet) error {

	if err := db.session.Query("INSERT INTO metric_names (metric_key, tag_set) VALUES (?, ?)", metricKey, tagSet.Serialize()).Exec(); err != nil {
		return err
	}
	if err := db.session.Query("UPDATE metric_name_set SET metric_names = metric_names + ? WHERE shard = ?", []string{string(metricKey)}, 0).Exec(); err != nil {
		return err
	}
	return nil

}

func (db *defaultDatabase) AddMetricNames(metrics []api.TaggedMetric) error {
	queryInsert := "INSERT INTO metric_names (metric_key, tag_set) VALUES (?, ?)"
	queryUpdate := "UPDATE metric_name_set SET metric_names = metric_names + ? WHERE shard = ?"

	c := make(chan *gocql.Query, 10)
	done := make(chan bool)
	go func() {
		for {
			boundQuery, more := <-c
			if !more {
				done <- true
				return
			}
			_ = boundQuery.Exec()
		}
	}()

	//For every query queue up an insert and a shard update and start streaming them.
	for _, m := range metrics {
		boundQuery := db.session.Bind(queryInsert, func(q *gocql.QueryInfo) ([]interface{}, error) {
			data := make([]interface{}, 2)
			data[0] = m.MetricKey
			data[1] = m.TagSet.Serialize()
			return data, nil
		})
		boundQuery.Consistency(gocql.One)
		c <- boundQuery

		boundQuery = db.session.Bind(queryUpdate, func(q *gocql.QueryInfo) ([]interface{}, error) {
			data := make([]interface{}, 2)
			data[0] = []string{string(m.MetricKey)}
			data[1] = 0
			return data, nil
		})
		boundQuery.Consistency(gocql.One)
		c <- boundQuery
	}
	close(c)

	<-done
	return nil
}

func (db *defaultDatabase) AddToTagIndex(tagKey string, tagValue string, metricKey api.MetricKey) error {
	err := db.session.Query(
		"UPDATE tag_index SET metric_keys = metric_keys + ? WHERE tag_key = ? AND tag_value = ?",
		[]string{string(metricKey)},
		tagKey,
		tagValue,
	).Exec()
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
	if len(tags) == 0 {
		//
		return nil, newNoSuchMetricError(string(metricKey))
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
	return keys, nil
}

func (db *defaultDatabase) RemoveMetricName(metricKey api.MetricKey, tagSet api.TagSet) error {
	return db.session.Query(
		"DELETE FROM metric_names WHERE metric_key = ? AND tag_set = ?",
		metricKey,
		tagSet.Serialize(),
	).Exec()
}

func (db *defaultDatabase) RemoveFromTagIndex(tagKey string, tagValue string, metricKey api.MetricKey) error {
	return db.session.Query(
		"UPDATE tag_index SET metric_keys = metric_keys - ? WHERE tag_key = ? AND tag_value = ?",
		[]string{string(metricKey)},
		tagKey,
		tagValue,
	).Exec()
}
