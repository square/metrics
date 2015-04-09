// Package internal holds classes used in internal implementation of metric-indexer
package internal

import (
	"github.com/gocql/gocql"
	"square/vis/metrics-indexer/api"
)

// Database represents internal connection to Cassandra.
type Database interface {
	// Insertion Methods
	// -----------------
	AddMetricName(metricKey api.MetricKey, metric api.TagSet) error
	AddTagIndex(tagKey string, tagValue string, metricKey api.MetricKey) error

	// Query methods
	// -------------
	GetTagSet(metricKey api.MetricKey) ([]api.TagSet, error)
	GetMetricKeys(tagKey string, tagValue string) ([]api.MetricKey, error)
}

type defaultDatabase struct {
	session *gocql.Session
}

// AddMetricName inserts to metric to Cassandra.
func (db *defaultDatabase) AddMetricName(metricKey api.MetricKey, tagSet api.TagSet) error {
	return db.session.Query(
		"INSERT INTO metric_names (metric_key, tag_set) VALUES (?, ?)",
		metricKey,
		tagSet.Serialize(),
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
