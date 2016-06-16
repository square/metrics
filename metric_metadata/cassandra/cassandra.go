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

package cassandra

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/square/metrics/api"
	"github.com/square/metrics/metric_metadata"
)

type MetricMetadataAPI struct {
	db cassandraDatabase
}

var _ metadata.MetricAPI = (*MetricMetadataAPI)(nil)
var _ metadata.MetricUpdateAPI = (*MetricMetadataAPI)(nil)

type Config struct {
	Hosts    []string `yaml:"hosts"`
	Keyspace string   `yaml:"keyspace"`
}

// NewCassandraMetricMetadataAPI creates a new instance of API from the given configuration.
func NewMetricMetadataAPI(config Config) (*MetricMetadataAPI, error) {
	// @@ leaking param: config
	clusterConfig := gocql.NewCluster()
	clusterConfig.Consistency = gocql.One
	clusterConfig.Hosts = config.Hosts
	clusterConfig.Keyspace = config.Keyspace
	clusterConfig.Timeout = time.Second * 30
	db, err := newCassandraDatabase(clusterConfig)
	if err != nil {
		return nil, err
	}
	return &MetricMetadataAPI{
		db: db,
	}, nil
	// @@ &MetricMetadataAPI literal escapes to heap
}

func (a *MetricMetadataAPI) AddMetric(metric api.TaggedMetric, context metadata.Context) error {
	// @@ leaking param: context
	// @@ leaking param content: a
	// @@ leaking param: metric
	// @@ leaking param content: a
	defer context.Profiler.Record("Cassandra AddMetric")()
	if err := a.db.AddMetricName(metric.MetricKey, metric.TagSet); err != nil {
		return err
	}
	return a.AddMetricTagsToTagIndex(metric, context)
}
func (a *MetricMetadataAPI) AddMetricTagsToTagIndex(metric api.TaggedMetric, context metadata.Context) error {
	// @@ leaking param: context
	// @@ leaking param content: a
	// @@ leaking param content: metric
	// @@ leaking param: metric
	defer context.Profiler.Record("Cassandra AddMetricTagsToTagIndex")()
	for tagKey, tagValue := range metric.TagSet {
		if err := a.db.AddToTagIndex(tagKey, tagValue, metric.MetricKey); err != nil {
			return err
		}
	}
	return nil
}

func (a *MetricMetadataAPI) AddMetrics(metrics []api.TaggedMetric, context metadata.Context) error {
	// @@ leaking param: context
	// @@ leaking param content: a
	// @@ leaking param content: metrics
	// @@ leaking param content: a
	// @@ leaking param: metrics
	defer context.Profiler.Record("Cassandra AddMetrics")()
	// Add each of the metrics to the tag index
	for _, metric := range metrics {
		err := a.AddMetricTagsToTagIndex(metric, context)
		if err != nil {
			return err
		}
	}
	return a.db.AddMetricNames(metrics)
}

func (a *MetricMetadataAPI) GetAllTags(metricKey api.MetricKey, context metadata.Context) ([]api.TagSet, error) {
	// @@ leaking param: context
	// @@ leaking param content: a
	// @@ leaking param: metricKey
	defer context.Profiler.Record("Cassandra GetAllTags")()
	return a.db.GetTagSet(metricKey)
}

func (a *MetricMetadataAPI) GetMetricsForTag(tagKey, tagValue string, context metadata.Context) ([]api.MetricKey, error) {
	// @@ leaking param: context
	// @@ leaking param content: a
	// @@ leaking param: tagKey
	// @@ leaking param: tagValue
	defer context.Profiler.Record("Cassandra GetMetricsForTag")()
	return a.db.GetMetricKeys(tagKey, tagValue)
}

func (a *MetricMetadataAPI) GetAllMetrics(context metadata.Context) ([]api.MetricKey, error) {
	// @@ leaking param: context
	// @@ leaking param content: a
	defer context.Profiler.Record("Cassandra GetAllMetrics")()
	return a.db.GetAllMetrics()
}

// CheckHealthy checks if the underlying connection to Cassandra is healthy
func (a *MetricMetadataAPI) CheckHealthy() error {
	// @@ leaking param content: a
	return a.db.CheckHealthy()
}

type cassandraDatabase struct {
	session *gocql.Session
}

// NewCassandraDatabase creates an instance of database, backed by Cassandra.
func newCassandraDatabase(clusterConfig *gocql.ClusterConfig) (cassandraDatabase, error) {
	// @@ leaking param content: clusterConfig
	session, err := clusterConfig.CreateSession()
	if err != nil {
		return cassandraDatabase{}, err
	}
	return cassandraDatabase{
		session: session,
	}, nil
}

// AddMetricName inserts the metric to Cassandra.
func (db *cassandraDatabase) AddMetricName(metricKey api.MetricKey, tagSet api.TagSet) error {
	// @@ leaking param content: db
	// @@ leaking param: metricKey
	// @@ leaking param content: db
	if err := db.session.Query("INSERT INTO metric_names (metric_key, tag_set) VALUES (?, ?)", metricKey, tagSet.Serialize()).Exec(); err != nil {
		return err
		// @@ ... argument escapes to heap
		// @@ metricKey escapes to heap
		// @@ tagSet.Serialize() escapes to heap
	}
	if err := db.session.Query("UPDATE metric_name_set SET metric_names = metric_names + ? WHERE shard = ?", []string{string(metricKey)}, 0).Exec(); err != nil {
		return err
		// @@ ... argument escapes to heap
		// @@ []string literal escapes to heap
		// @@ []string literal escapes to heap
		// @@ 0 escapes to heap
	}
	return nil

}

// AddMetricNames adds many metric names to Cassandra (equivalent to calling AddMetricName many times, but more performant)
func (db *cassandraDatabase) AddMetricNames(metrics []api.TaggedMetric) error {
	// @@ leaking param: metrics
	// @@ leaking param content: db
	// @@ leaking param content: db
	queryInsert := "INSERT INTO metric_names (metric_key, tag_set) VALUES (?, ?)"
	queryUpdate := "UPDATE metric_name_set SET metric_names = metric_names + ? WHERE shard = ?"

	//For every query queue up an insert and a shard update and start streaming them.
	for _, m := range metrics {
		boundQuery := db.session.Bind(queryInsert, func(q *gocql.QueryInfo) ([]interface{}, error) {
			// @@ moved to heap: m
			return []interface{}{
				// @@ func literal escapes to heap
				// @@ func literal escapes to heap
				m.MetricKey,
				// @@ []interface {} literal escapes to heap
				m.TagSet.Serialize(),
				// @@ m.MetricKey escapes to heap
				// @@ &m escapes to heap
			}, nil
			// @@ m.TagSet.Serialize() escapes to heap
		})
		boundQuery.Consistency(gocql.One)
		err := boundQuery.Exec()
		// @@ inlining call to Consistency
		if err != nil {
			return err
		}

		boundQuery = db.session.Bind(queryUpdate, func(q *gocql.QueryInfo) ([]interface{}, error) {
			return []interface{}{
				// @@ can inline (*cassandraDatabase).AddMetricNames.func2
				// @@ func literal escapes to heap
				// @@ func literal escapes to heap
				[]string{string(m.MetricKey)},
				// @@ []interface {} literal escapes to heap
				0,
				// @@ []string literal escapes to heap
				// @@ []string literal escapes to heap
				// @@ &m escapes to heap
			}, nil
			// @@ 0 escapes to heap
		})
		boundQuery.Consistency(gocql.One)
		err = boundQuery.Exec()
		// @@ inlining call to Consistency
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *cassandraDatabase) AddToTagIndex(tagKey string, tagValue string, metricKey api.MetricKey) error {
	// @@ leaking param content: db
	// @@ leaking param: metricKey
	// @@ leaking param: tagKey
	// @@ leaking param: tagValue
	err := db.session.Query(
		"UPDATE tag_index SET metric_keys = metric_keys + ? WHERE tag_key = ? AND tag_value = ?",
		[]string{string(metricKey)},
		tagKey,
		// @@ []string literal escapes to heap
		// @@ []string literal escapes to heap
		// @@ tagKey escapes to heap
		// @@ tagValue escapes to heap
		tagValue,
	).Exec()
	return err
	// @@ ... argument escapes to heap
}

func (db *cassandraDatabase) GetTagSet(metricKey api.MetricKey) ([]api.TagSet, error) {
	// @@ leaking param content: db
	// @@ leaking param: metricKey
	var tags []api.TagSet
	rawTag := ""
	iterator := db.session.Query(
		// @@ moved to heap: rawTag
		"SELECT tag_set FROM metric_names WHERE metric_key = ?",
		metricKey,
		// @@ metricKey escapes to heap
	).Iter()
	for iterator.Scan(&rawTag) {
		// @@ ... argument escapes to heap
		parsedTagSet := api.ParseTagSet(rawTag)
		// @@ ... argument escapes to heap
		// @@ &rawTag escapes to heap
		// @@ &rawTag escapes to heap
		if parsedTagSet != nil {
			tags = append(tags, parsedTagSet)
		}
	}
	if err := iterator.Close(); err != nil {
		return nil, err
		// @@ inlining call to Close
	}
	if len(tags) == 0 {
		//
		return nil, metadata.NewNoSuchMetricError(string(metricKey))
	}
	// @@ inlining call to metadata.NewNoSuchMetricError
	// @@ metadata.NewNoSuchMetricError(string(metricKey)) escapes to heap
	return tags, nil
}

func (db *cassandraDatabase) GetMetricKeys(tagKey string, tagValue string) ([]api.MetricKey, error) {
	// @@ leaking param content: db
	// @@ leaking param: tagKey
	// @@ leaking param: tagValue
	var keys []api.MetricKey
	err := db.session.Query(
		// @@ moved to heap: keys
		"SELECT metric_keys FROM tag_index WHERE tag_key = ? AND tag_value = ?",
		tagKey,
		// @@ tagKey escapes to heap
		// @@ tagValue escapes to heap
		tagValue,
	).Scan(&keys)
	if err == gocql.ErrNotFound {
		// @@ ... argument escapes to heap
		// @@ ... argument escapes to heap
		// @@ &keys escapes to heap
		// @@ &keys escapes to heap
		return keys, nil
	}
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (db *cassandraDatabase) GetAllMetrics() ([]api.MetricKey, error) {
	// @@ leaking param content: db
	var keys []api.MetricKey
	err := db.session.Query("SELECT metric_names FROM metric_name_set WHERE shard = ?", 0).Scan(&keys)
	// @@ moved to heap: keys
	if err != nil {
		// @@ ... argument escapes to heap
		// @@ 0 escapes to heap
		// @@ ... argument escapes to heap
		// @@ &keys escapes to heap
		// @@ &keys escapes to heap
		return nil, err
	}
	return keys, nil
}

func (db *cassandraDatabase) RemoveFromTagIndex(tagKey string, tagValue string, metricKey api.MetricKey) error {
	// @@ leaking param content: db
	// @@ leaking param: metricKey
	// @@ leaking param: tagKey
	// @@ leaking param: tagValue
	return db.session.Query(
		"UPDATE tag_index SET metric_keys = metric_keys - ? WHERE tag_key = ? AND tag_value = ?",
		[]string{string(metricKey)},
		tagKey,
		// @@ []string literal escapes to heap
		// @@ []string literal escapes to heap
		// @@ tagKey escapes to heap
		// @@ tagValue escapes to heap
		tagValue,
	).Exec()
}

// @@ ... argument escapes to heap

// CheckHealthy checks if the connection to Cassandra is healthy
func (db *cassandraDatabase) CheckHealthy() error {
	// @@ leaking param content: db
	return db.session.Query("SELECT now() FROM system.local").Exec()
}
