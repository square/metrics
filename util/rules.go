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

package util

import (
	"bytes"
	"regexp"
	"strings"
	"sync"

	"github.com/square/metrics/api"
	"github.com/square/metrics/log"
	"gopkg.in/yaml.v2"
)

var defaultRegex = "[^.]+"

// RawRule is the input provided by the YAML file to specify the rul.
type RawRule struct {
	Pattern          string            `yaml:"pattern"`
	MetricKeyPattern string            `yaml:"metric_key"`
	Regex            map[string]string `yaml:"regex,omitempty"`
}

// RawRules is list of RawRule
type RawRules struct {
	RawRules []RawRule `yaml:"rules"`
}

// Rule is a sanitized version of RawRule. Only valid rules
// can be converted to Rule.
type Rule struct {
	raw                  RawRule
	graphitePatternRegex *regexp.Regexp
	MetricKeyRegex       *regexp.Regexp
	graphitePatternTags  []string // tags extracted from the raw graphite string, in the order of appearance.
	metricKeyTags        []string // tags extracted from MetricKey, in the order of appearance.
	Statistics           RuleStatistics
}

type RuleStatistics struct {
	mutex             *sync.Mutex
	statisticsEnabled bool
	Matches           int
	SuccessfulMatches []string
}

// RuleSet is a sanitized version of RawRules.
// Rules are matched sequentially until a correct one is matched.
type RuleSet struct {
	Rules             []Rule
	statisticsEnabled bool
}

// Compile a given RawRule into a regex and exposed tagset.
// error is an instance of RuleError.
func Compile(rule RawRule) (Rule, error) {
	if len(rule.MetricKeyPattern) == 0 {
		return Rule{}, newInvalidMetricKey(rule.MetricKeyPattern)
	}
	graphitePatternTags := extractTags(rule.Pattern)
	if graphitePatternTags == nil {
		return Rule{}, newInvalidPattern(rule.MetricKeyPattern)
	}
	metricKeyTags := extractTags(string(rule.MetricKeyPattern))
	if !isSubset(metricKeyTags, graphitePatternTags) {
		return Rule{}, newInvalidMetricKey(rule.MetricKeyPattern)
	}
	if metricKeyTags == nil {
		return Rule{}, newInvalidMetricKey(rule.MetricKeyPattern)
	}
	if !rule.checkTagRegexes() {
		return Rule{}, newInvalidCustomRegex(rule.MetricKeyPattern)
	}
	regex := rule.toRegexp(rule.Pattern)
	if regex == nil {
		return Rule{}, newInvalidPattern(rule.MetricKeyPattern)
	}
	if regex.NumSubexp() != len(graphitePatternTags) {
		return Rule{}, newInvalidCustomRegex(rule.MetricKeyPattern)
	}
	metricKeyRegex := rule.toRegexp(rule.MetricKeyPattern)
	if metricKeyRegex == nil {
		return Rule{}, newInvalidPattern(rule.MetricKeyPattern)
	}
	if metricKeyRegex.NumSubexp() != len(metricKeyTags) {
		return Rule{}, newInvalidCustomRegex(rule.MetricKeyPattern)
	}

	stats := RuleStatistics{}
	stats.mutex = &sync.Mutex{}

	return Rule{
		raw:                  rule,
		graphitePatternRegex: regex,
		MetricKeyRegex:       metricKeyRegex,
		graphitePatternTags:  graphitePatternTags,
		metricKeyTags:        metricKeyTags,
		Statistics:           stats,
	}, nil
}

func (rule *Rule) AddMatch(matchedResult string) {
	rule.Statistics.AddMatch(matchedResult)
}

func (ruleStat *RuleStatistics) AddMatch(matchedResult string) {
	if !ruleStat.statisticsEnabled {
		return
	}
	ruleStat.mutex.Lock()
	defer ruleStat.mutex.Unlock()
	ruleStat.Matches++
	ruleStat.SuccessfulMatches = append(ruleStat.SuccessfulMatches, matchedResult)
}

// MatchRule sees if a given graphite string matches the rule, and if so, returns the generated tag.
func (rule *Rule) MatchRule(input string) (api.TaggedMetric, bool) {
	if strings.Contains(input, "\x00") {
		log.Errorf("MatchRule (graphite string => metric name) has been given bad metric: `%s`", input)
	}
	tagSet := extractTagValues(rule.graphitePatternRegex, rule.graphitePatternTags, input)
	if tagSet == nil {
		return api.TaggedMetric{}, false
	}
	interpolatedKey, err := interpolateTags(rule.raw.MetricKeyPattern, tagSet, false)
	if err != nil {
		return api.TaggedMetric{}, false
	}
	// Do not output tags appearing in both graphite metric & metric key.
	// for exmaple, if graphite metric is
	//   `foo.%a%.%b%`
	// and metric key is
	//   `bar.%b%`
	// the resulting tag set should only contain {a} after the matching
	// because %b% is already encoded.
	for _, metricKeyTag := range rule.metricKeyTags {
		if _, containsKey := tagSet[metricKeyTag]; containsKey {
			delete(tagSet, metricKeyTag)
		}
	}
	rule.AddMatch(input)
	if strings.Contains(interpolatedKey, "\x00") {
		log.Errorf("MatchRule (graphite string => metric name) is returning bad metric: `%s` from input `%s`", interpolatedKey, input)
	}
	return api.TaggedMetric{
		api.MetricKey(interpolatedKey),
		tagSet,
	}, true
}

// ToGraphiteName transforms the given tagged metric back to its graphite metric.
func (rule Rule) ToGraphiteName(taggedMetric api.TaggedMetric) (GraphiteMetric, error) {
	extractedTagSet := extractTagValues(rule.MetricKeyRegex, rule.metricKeyTags, string(taggedMetric.MetricKey))
	if extractedTagSet == nil {
		// no match found. not a correct rule to interpolate.
		return "", newCannotInterpolate(taggedMetric)
	}
	// Merge the tags in the provided tag set, and tags extracted from the metric.
	// This is necessary because tags embedded in the metric are not
	// exported to the tagset.
	mergedTagSet := taggedMetric.TagSet.Merge(extractedTagSet)
	interpolated, err := interpolateTags(rule.raw.Pattern, mergedTagSet, true)
	if err != nil {
		return "", err
	}
	return GraphiteMetric(interpolated), nil
}

// MatchRule sees if a given graphite string matches
// any of the specified rules.
func (ruleSet *RuleSet) MatchRule(input string) (api.TaggedMetric, bool) {
	for i := 0; i < len(ruleSet.Rules); i++ {
		value, matched := ruleSet.Rules[i].MatchRule(input)
		if matched {
			return value, matched
		}
	}
	return api.TaggedMetric{}, false
}

func (rule *Rule) EnableStats() {
	rule.Statistics.statisticsEnabled = true
}

func (rule *Rule) DisableStats() {
	rule.Statistics.statisticsEnabled = false
}

func (ruleSet *RuleSet) EnableStats() {
	for i := 0; i < len(ruleSet.Rules); i++ {
		ruleSet.Rules[i].EnableStats()
	}
	ruleSet.statisticsEnabled = true
}

func (ruleSet *RuleSet) DisableStats() {
	for i := 0; i < len(ruleSet.Rules); i++ {
		ruleSet.Rules[i].DisableStats()
	}
	ruleSet.statisticsEnabled = false
}

// GraphitePatternTags return a list of tags available in the original metric.
func (rule Rule) GraphitePatternTags() []string {
	return rule.graphitePatternTags
}

// ToGraphiteName transforms the given tagged metric back to its graphite name,
// checking against all the rules.
func (ruleSet RuleSet) ToGraphiteName(taggedMetric api.TaggedMetric) (GraphiteMetric, error) {
	for _, rule := range ruleSet.Rules {
		reversed, err := rule.ToGraphiteName(taggedMetric)
		if err == nil {
			return reversed, nil
		}
	}
	return "", newCannotInterpolate(taggedMetric)
}

// checkTagRegexes sees if any of the custom regular expressions are invalid.
func (rule RawRule) checkTagRegexes() bool {
	for _, regex := range rule.Regex {
		compiled, err := regexp.Compile(regex)
		if err != nil {
			return false
		}
		if compiled.NumSubexp() > 0 {
			return false // do not allow subexpressions.
		}
	}
	return true
}

func (rule RawRule) toRegexp(pattern string) *regexp.Regexp {
	splitted := strings.Split(pattern, "%")
	buffer := new(bytes.Buffer)
	if len(splitted)%2 == 0 {
		// invalid pattern - even number of parts mean odd number of %.
		return nil
	}
	buffer.WriteString("^")
	for index, token := range splitted {
		if isTagPortion(index) {
			regex, contains := rule.Regex[token]
			if !contains {
				// use the defuault regex
				regex = defaultRegex
			}
			// wrap the regex in parenthesis, so that it can be matched.
			// if the regex contains matching groups, bad things will happen.
			buffer.WriteString("(" + regex + ")")
		} else {
			buffer.WriteString(regexp.QuoteMeta(token))
		}
	}
	buffer.WriteString("$")
	compiled, err := regexp.Compile(buffer.String())
	if err != nil {
		return nil
	}
	return compiled
}

// extractTagValues extracts the tagset using the given regex and the list of tags.
func extractTagValues(regex *regexp.Regexp, tagList []string, input string) api.TagSet {
	matches := regex.FindStringSubmatch(input)
	if matches == nil {
		return nil
	}
	tagSet := api.NewTagSet()
	for index, tagValue := range matches {
		if index == 0 {
			continue
		}
		tagKey := tagList[index-1]
		tagSet[tagKey] = tagValue
	}
	return tagSet
}

// extractTags extracts list of tags in the given pattern string.
func extractTags(pattern string) []string {
	if len(pattern) == 0 {
		return nil // empty pattern is not allowed.
	}
	splitted := strings.Split(pattern, "%")
	if len(splitted)%2 == 0 {
		return nil // invalid tags.
	}

	result := make([]string, len(splitted)/2, len(splitted)/2)
	for index, token := range splitted {
		if isTagPortion(index) {
			if len(token) == 0 {
				return nil // no empty tag
			}
			result[index/2] = token
		}
	}

	// check for duplicates.
	exists := make(map[string]bool)
	for _, token := range result {
		if exists[token] {
			return nil // no duplicate
		}
		exists[token] = true
	}
	return result
}

// LoadYAML loads a RuleSet from the byte array of the YAML file.
// error is an interface of RuleError.
func LoadYAML(input []byte) (RuleSet, error) {
	rawRules := RawRules{}
	if err := yaml.Unmarshal(input, &rawRules); err != nil {
		return RuleSet{}, ruleError{
			code:    InvalidYaml,
			message: err.Error(),
		}
	}
	rules := make([]Rule, len(rawRules.RawRules))
	for index, rawRule := range rawRules.RawRules {
		rule, err := Compile(rawRule)
		if err != nil {
			return RuleSet{}, err
		}
		rules[index] = rule
	}
	return RuleSet{
		Rules:             rules,
		statisticsEnabled: false,
	}, nil
}

// check if setA is subset of setB.
func isSubset(setA, setB []string) bool {
	set := make(map[string]bool)
	for _, v := range setB {
		set[v] = true
	}
	for _, v := range setA {
		if !set[v] {
			return false
		}
	}
	return true
}

func isTagPortion(index int) bool {
	return index%2 == 1
}

func interpolateTags(pattern string, tagSet api.TagSet, enforceAllTagsUsed bool) (string, error) {

	usedTags := make(map[string]bool)

	splitted := strings.Split(pattern, "%")
	buffer := new(bytes.Buffer)
	for index, token := range splitted {
		if isTagPortion(index) {
			usedTags[token] = true
			tagValue, hasTag := tagSet[token]
			if hasTag {
				buffer.WriteString(tagValue)
			} else {
				return "", newMissingTag(token)
			}
		} else {
			buffer.WriteString(token)
		}
	}

	if enforceAllTagsUsed {
		for key := range tagSet {
			if !usedTags[key] {
				return "", newUnusedTag(key)
			}
		}
	}

	return buffer.String(), nil
}
