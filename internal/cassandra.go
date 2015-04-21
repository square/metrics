// Package internal holds classes used in internal implementation of metric-indexer
package internal

import (
	"github.com/gocql/gocql"
	"github.com/square/metrics-indexer/api"
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

type defaultDatabase struct {
	session *gocql.Session
}

// NewCassandraDatabase creates an instance of database, backed by Cassandra.
func NewCassandraDatabase(clusterConfig *gocql.ClusterConfig) (Database, error) {
	session, err := clusterConfig.CreateSession()
	if err != nil {
		return nil, err
	}
	return &defaultDatabase{session: session}, nil
}

// AddMetricName inserts to metric to Cassandra.
func (db *defaultDatabase) AddMetricName(metricKey api.MetricKey, tagSet api.TagSet) error {
	return db.session.Query(
		"INSERT INTO metric_names (metric_key, tag_set) VALUES (?, ?)",
		metricKey,
		tagSet.Serialize(),
	).Exec()
}

func (db *defaultDatabase) AddToTagIndex(tagKey string, tagValue string, metricKey api.MetricKey) error {
	return db.session.Query(
		"UPDATE tag_index SET metric_keys = metric_keys + ? WHERE tag_key = ? AND tag_value = ?",
		[]string{string(metricKey)},
		tagKey,
		tagValue,
	).Exec()
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
	metricKey := ""
	iterator := db.session.Query("SELECT distinct metric_key FROM metric_names").Iter()
	for iterator.Scan(&metricKey) {
		keys = append(keys, api.MetricKey(metricKey))
	}
	if err := iterator.Close(); err != nil {
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
