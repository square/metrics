package internal

import (
	"io/ioutil"
	"os"

	"github.com/gocql/gocql"
	"github.com/square/metrics/api"
)

// API implementations.
type defaultAPI struct {
	db      Database
	ruleset RuleSet
}

// NewAPI creates a new instance of API from the given configuration.
func NewAPI(config api.Configuration) (api.API, error) {
	file, err := os.Open(config.RuleYamlFilePath)
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	ruleset, err := LoadYAML(bytes)
	if err != nil {
		return nil, err
	}
	clusterConfig := gocql.NewCluster()
	clusterConfig.Hosts = config.Hosts
	clusterConfig.Keyspace = config.Keyspace
	db, err := NewCassandraDatabase(clusterConfig)
	if err != nil {
		return nil, err
	}
	return &defaultAPI{
		db:      db,
		ruleset: ruleset,
	}, nil
}

func (a *defaultAPI) AddMetric(metric api.TaggedMetric) error {
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

func (a *defaultAPI) GetAllTags(metricKey api.MetricKey) ([]api.TagSet, error) {
	return a.db.GetTagSet(metricKey)
}

func (a *defaultAPI) GetMetricsForTag(tagKey, tagValue string) ([]api.MetricKey, error) {
	return a.db.GetMetricKeys(tagKey, tagValue)
}

func (a *defaultAPI) GetAllMetrics() ([]api.MetricKey, error) {
	return a.db.GetAllMetrics()
}

func (a *defaultAPI) RemoveMetric(metric api.TaggedMetric) error {
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

func (a *defaultAPI) ToGraphiteName(metric api.TaggedMetric) (api.GraphiteMetric, error) {
	return a.ruleset.ToGraphiteName(metric)
}

func (a *defaultAPI) ToTaggedName(metric api.GraphiteMetric) (api.TaggedMetric, error) {
	match, matched := a.ruleset.MatchRule(string(metric))
	if matched {
		return match, nil
	}
	return api.TaggedMetric{}, ErrNoMatch
}

// ensure interface
var _ api.API = (*defaultAPI)(nil)
