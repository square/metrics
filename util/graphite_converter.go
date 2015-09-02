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

// Package api holds common data types and public interface exposed by the indexer library.
// Refer to the doc
// https://docs.google.com/a/squareup.com/document/d/1k0Wgi2wnJPQoyDyReb9dyIqRrD8-v0u8hz37S282ii4/edit
// for the terminology.

package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/square/metrics/api"
	"github.com/square/metrics/log"
)

// GraphiteMetric is a flat, dot-separated identifier to a series of metric.
type GraphiteMetric string

type GraphiteConverterConfig struct {
	ConversionRulesPath string `yaml:"conversion_rules_path"`
}

type RuleBasedGraphiteConverter struct {
	Ruleset RuleSet
}

func (g *RuleBasedGraphiteConverter) ToGraphiteName(metric api.TaggedMetric) (GraphiteMetric, error) {
	return g.Ruleset.ToGraphiteName(metric)
}

func (g *RuleBasedGraphiteConverter) ToTaggedName(metric GraphiteMetric) (api.TaggedMetric, error) {
	match, matched := g.Ruleset.MatchRule(string(metric))
	if matched {
		return match, nil
	}
	return api.TaggedMetric{}, newNoMatch()
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

type GraphiteConverter interface {
	// Convert the given tag-based metric name to graphite metric name,
	// using the configured rules. May error out.
	ToGraphiteName(metric api.TaggedMetric) (GraphiteMetric, error)
	// Converts the given graphite metric to the tag-based meric,
	// using the configured rules. May error out.
	ToTaggedName(metric GraphiteMetric) (api.TaggedMetric, error)
}

var _ GraphiteConverter = (*RuleBasedGraphiteConverter)(nil)
