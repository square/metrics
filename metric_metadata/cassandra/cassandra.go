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

package cassandra

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/square/metrics/api"
)

type CassandraMetricMetadataAPI struct {
	db cassandraDatabase
}

var _ api.MetricMetadataAPI = (*CassandraMetricMetadataAPI)(nil)

type CassandraMetricMetadataConfig struct {
	Hosts    []string `yaml:"hosts"`
	Keyspace string   `yaml:"keyspace"`
}

// NewCassandraMetricMetadataAPI creates a new instance of API from the given configuration.
func NewCassandraMetricMetadataAPI(config CassandraMetricMetadataConfig) (api.MetricMetadataAPI, error) {
	clusterConfig := gocql.NewCluster()
	clusterConfig.Consistency = gocql.One
	clusterConfig.Hosts = config.Hosts
	clusterConfig.Keyspace = config.Keyspace
	clusterConfig.Timeout = time.Second * 30
	db, err := NewCassandraDatabase(clusterConfig)
	if err != nil {
		return nil, err
	}
	return &CassandraMetricMetadataAPI{
		db: db,
	}, nil
}

func (a *CassandraMetricMetadataAPI) AddMetric(metric api.TaggedMetric) error {
	if err := a.db.AddMetricName(metric.MetricKey, metric.TagSet); err != nil {
		return err
	}
	for tagKey, tagValue := range metric.TagSet {
		if err := a.db.AddToTagIndex(tagKey, tagValue, metric.MetricKey); err != nil {
			return err
		}
	}
	return nil
}

func (a *CassandraMetricMetadataAPI) AddMetrics(metrics []api.TaggedMetric) error {
	a.db.AddMetricNames(metrics)
	return nil
}

func (a *CassandraMetricMetadataAPI) GetAllTags(metricKey api.MetricKey) ([]api.TagSet, error) {
	return a.db.GetTagSet(metricKey)
}

func (a *CassandraMetricMetadataAPI) GetMetricsForTag(tagKey, tagValue string) ([]api.MetricKey, error) {
	return a.db.GetMetricKeys(tagKey, tagValue)
}

func (a *CassandraMetricMetadataAPI) GetAllMetrics() ([]api.MetricKey, error) {
	return a.db.GetAllMetrics()
}

func (a *CassandraMetricMetadataAPI) RemoveMetric(metric api.TaggedMetric) error {
	if err := a.db.RemoveMetricName(metric.MetricKey, metric.TagSet); err != nil {
		return err
	}
	for tagKey, tagValue := range metric.TagSet {
		if err := a.db.RemoveFromTagIndex(tagKey, tagValue, metric.MetricKey); err != nil {
			return err
		}
	}
	return nil
}

// ensure interface
var _ api.MetricMetadataAPI = (*CassandraMetricMetadataAPI)(nil)

type cassandraDatabase struct {
	session *gocql.Session
}

// NewCassandraDatabase creates an instance of database, backed by Cassandra.
func NewCassandraDatabase(clusterConfig *gocql.ClusterConfig) (cassandraDatabase, error) {
	session, err := clusterConfig.CreateSession()
	if err != nil {
		return cassandraDatabase{}, err
	}
	return cassandraDatabase{
		session: session,
	}, nil
}

// AddMetricName inserts to metric to Cassandra.
func (db *cassandraDatabase) AddMetricName(metricKey api.MetricKey, tagSet api.TagSet) error {

	if err := db.session.Query("INSERT INTO metric_names (metric_key, tag_set) VALUES (?, ?)", metricKey, tagSet.Serialize()).Exec(); err != nil {
		return err
	}
	if err := db.session.Query("UPDATE metric_name_set SET metric_names = metric_names + ? WHERE shard = ?", []string{string(metricKey)}, 0).Exec(); err != nil {
		return err
	}
	return nil

}

func (db *cassandraDatabase) AddMetricNames(metrics []api.TaggedMetric) error {
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

func (db *cassandraDatabase) AddToTagIndex(tagKey string, tagValue string, metricKey api.MetricKey) error {
	err := db.session.Query(
		"UPDATE tag_index SET metric_keys = metric_keys + ? WHERE tag_key = ? AND tag_value = ?",
		[]string{string(metricKey)},
		tagKey,
		tagValue,
	).Exec()
	return err
}

func (db *cassandraDatabase) GetTagSet(metricKey api.MetricKey) ([]api.TagSet, error) {
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
		return nil, api.NewNoSuchMetricError(string(metricKey))
	}
	return tags, nil
}

func (db *cassandraDatabase) GetMetricKeys(tagKey string, tagValue string) ([]api.MetricKey, error) {
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

func (db *cassandraDatabase) GetAllMetrics() ([]api.MetricKey, error) {
	var keys []api.MetricKey
	err := db.session.Query("SELECT metric_names FROM metric_name_set WHERE shard = ?", 0).Scan(&keys)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (db *cassandraDatabase) RemoveMetricName(metricKey api.MetricKey, tagSet api.TagSet) error {
	return db.session.Query(
		"DELETE FROM metric_names WHERE metric_key = ? AND tag_set = ?",
		metricKey,
		tagSet.Serialize(),
	).Exec()
}

func (db *cassandraDatabase) RemoveFromTagIndex(tagKey string, tagValue string, metricKey api.MetricKey) error {
	return db.session.Query(
		"UPDATE tag_index SET metric_keys = metric_keys - ? WHERE tag_key = ? AND tag_value = ?",
		[]string{string(metricKey)},
		tagKey,
		tagValue,
	).Exec()
}
