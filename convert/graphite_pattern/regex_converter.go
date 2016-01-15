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

package graphite_pattern

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/square/metrics/api"
	"github.com/square/metrics/log"
)

func NewConverter(conversionRulesPath string) (api.MetricConverter, error) {
	ruleset, err := LoadRules(conversionRulesPath)
	if err != nil {
		return nil, err
	}
	return &RuleBasedGraphiteConverter{ruleset}, nil
}

// GraphiteMetric is a flat, dot-separated identifier to a series of metric.
type GraphiteMetric string

type RuleBasedGraphiteConverter struct {
	Ruleset RuleSet
}

func (g *RuleBasedGraphiteConverter) EnableStats() {
	g.Ruleset.EnableStats()
}

func (g *RuleBasedGraphiteConverter) ToUntagged(metric api.TaggedMetric) (string, error) {
	return g.Ruleset.ToGraphiteName(metric)
}

func (g *RuleBasedGraphiteConverter) ToTagged(metric string) (api.TaggedMetric, error) {
	match, matched := g.Ruleset.MatchRule(metric)
	if matched {
		return match, nil
	}
	return api.TaggedMetric{}, newNoMatch()
}

func LoadRules(conversionRulesPath string) (RuleSet, error) {
	ruleSet := RuleSet{
		Rules: []Rule{},
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
			return RuleSet{}, fmt.Errorf("error opening file %s: %s", filename, err.Error())
		}
		defer file.Close()

		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			return RuleSet{}, fmt.Errorf("error reading file %s: %s", filename, err.Error())
		}

		rs, err := LoadYAML(bytes)
		if err != nil {
			return RuleSet{}, fmt.Errorf("error loading YAML from file %s: %s", filename, err.Error())
		}

		for i := range rs.Rules {
			rs.Rules[i].file = filename
		}

		ruleSet.Rules = append(ruleSet.Rules, rs.Rules...)
	}

	return ruleSet, nil
}
