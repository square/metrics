package internal

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gocql/gocql"
	"github.com/square/metrics/api"
	"github.com/square/metrics/log"
)

// API implementations.
type defaultAPI struct {
	db      Database
	ruleset RuleSet
}

// NewAPI creates a new instance of API from the given configuration.
func NewAPI(config api.Config) (api.API, error) {
	ruleset, err := LoadRules(config.ConversionRulesPath)
	if err != nil {
		return nil, err
	}

	clusterConfig := gocql.NewCluster()
	clusterConfig.Hosts = config.Hosts
	clusterConfig.Keyspace = config.Keyspace
	clusterConfig.Timeout = time.Second * 30
	db, err := NewCassandraDatabase(clusterConfig, config.Database)
	if err != nil {
		return nil, err
	}
	return &defaultAPI{
		db:      db,
		ruleset: ruleset,
	}, nil
}

func LoadRules(conversionRulesPath string) (RuleSet, error) {
	ruleSet := RuleSet{
		rules: []Rule{},
	}

	filenames, err := filepath.Glob(filepath.Join(conversionRulesPath, "*.yaml"))
	if err != nil {
		return RuleSet{}, err
	}

	sort.Strings(filenames)

	for _, filename := range filenames {
		log.Infof("Loading rules from %s", filename)
		file, err := os.Open(filename)
		if err != nil {
			return RuleSet{}, err
		}
		defer file.Close()

		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			return RuleSet{}, err
		}

		rs, err := LoadYAML(bytes)
		if err != nil {
			return RuleSet{}, err
		}

		ruleSet.rules = append(ruleSet.rules, rs.rules...)
	}

	return ruleSet, nil
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
	if metric.MetricKey == api.SpecialGraphiteName {
		return api.GraphiteMetric(metric.TagSet[api.SpecialGraphiteName]), nil
	}
	return a.ruleset.ToGraphiteName(metric)
}

func (a *defaultAPI) ToTaggedName(metric api.GraphiteMetric) (api.TaggedMetric, error) {
	match, matched := a.ruleset.MatchRule(string(metric))
	if matched {
		return match, nil
	}
	return api.TaggedMetric{}, newNoMatch()
}

func (a *defaultAPI) AddGraphiteMetric(metric api.GraphiteMetric) error {
	if graphiteDB, ok := a.db.(DatabaseGraphiteStore); ok {
		return graphiteDB.AddGraphiteMetric(metric)
	}
	return fmt.Errorf("database does not support the AddGraphiteMetric operator")
}
func (a *defaultAPI) GetAllGraphiteMetrics() ([]api.GraphiteMetric, error) {
	if graphiteDB, ok := a.db.(DatabaseGraphiteStore); ok {
		return graphiteDB.GetAllGraphiteMetrics()
	}
	return nil, fmt.Errorf("database does not support the GetAllGraphiteMetrics operator")
}

// ensure interface
var _ api.API = (*defaultAPI)(nil)
