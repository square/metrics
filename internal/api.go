package internal

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/gocql/gocql"
	"github.com/square/metrics/api"
)

// API implementations.
type defaultAPI struct {
	db      Database
	ruleset RuleSet
	cached  map[api.GraphiteMetric]api.TaggedMetric
	mutex   *sync.Mutex
}

// NewAPI creates a new instance of API from the given configuration.
func NewAPI(config api.Config) (api.API, error) {
	ruleset, err := loadRules(config.ConversionRulesPath)
	if err != nil {
		return nil, err
	}

	clusterConfig := gocql.NewCluster()
	clusterConfig.Hosts = config.Hosts
	clusterConfig.Keyspace = config.Keyspace
	clusterConfig.Timeout = time.Second * 30
	db, err := NewCassandraDatabase(clusterConfig)
	if err != nil {
		return nil, err
	}
	return &defaultAPI{
		db:      db,
		ruleset: ruleset,
		cached:  map[api.GraphiteMetric]api.TaggedMetric{},
		mutex:   &sync.Mutex{},
	}, nil
}

func loadRules(conversionRulesPath string) (RuleSet, error) {
	ruleSet := RuleSet{
		rules: []Rule{},
	}

	filenames, err := filepath.Glob(filepath.Join(conversionRulesPath, "*.yaml"))
	if err != nil {
		return RuleSet{}, err
	}

	sort.Strings(filenames)

	for _, filename := range filenames {
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
	return a.ruleset.ToGraphiteName(metric)
}

func (a *defaultAPI) ToTaggedName(metric api.GraphiteMetric) (api.TaggedMetric, error) {
	a.mutex.Lock()
	if result, ok := a.cached[metric]; ok {
		a.mutex.Unlock()
		return result, nil
	}
	a.mutex.Unlock()
	match, matched := a.ruleset.MatchRule(string(metric))
	if matched {
		a.mutex.Lock()
		a.cached[metric] = match
		a.mutex.Unlock()
		return match, nil
	}
	return api.TaggedMetric{}, newNoMatch()
}

// ensure interface
var _ api.API = (*defaultAPI)(nil)
