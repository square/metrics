package internal

import (
	"bytes"
	"errors"
	"regexp"
	"strings"

	"github.com/square/metrics-indexer/api"
	"gopkg.in/yaml.v2"
)

var defaultRegex = "[^.]+"

var (
	// ErrInvalidPattern is returned when an invalid rule pattern is provided.
	ErrInvalidPattern = errors.New("Invalid Pattern")
	// ErrInvalidMetricKey is retruned when an invalid metric key is provided.
	ErrInvalidMetricKey = errors.New("Invalid Metric Key")
	// ErrInvalidCustomRegex is retruned when the custom regex is invalid.
	ErrInvalidCustomRegex = errors.New("Invalid Custom Regex")
	// ErrMissingTag is returned during the reverse mapping, when a tag is not provided.
	ErrMissingTag = errors.New("Missing Tag")
	// ErrCannotInterpolate is returned when the tag interpolation fails.
	ErrCannotInterpolate = errors.New("Cannot Interpolate")
	// ErrNoMatch is returned when the conversion to tagged metric fails.
	ErrNoMatch = errors.New("No Match")
)

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
	raw                 RawRule
	regex               *regexp.Regexp
	graphitePatternTags []string // tags extracted from the raw graphite string, in the order of appearance.
	metricKeyTags       []string // tags extracted from MetricKey, in the order of appearance.
}

// RuleSet is a sanitized version of RawRules.
// Rules are matched sequentially until a correct one is matched.
type RuleSet struct {
	rules []Rule
}

// Compile a given RawRule into a regex and exposed tagset.
func Compile(rule RawRule) (Rule, error) {
	if len(rule.MetricKeyPattern) == 0 {
		return Rule{}, ErrInvalidMetricKey
	}
	graphitePatternTags := extractTags(rule.Pattern)
	if graphitePatternTags == nil {
		return Rule{}, ErrInvalidPattern
	}
	metricKeyTags := extractTags(string(rule.MetricKeyPattern))
	if !isSubset(metricKeyTags, graphitePatternTags) {
		return Rule{}, ErrInvalidMetricKey
	}
	if metricKeyTags == nil {
		return Rule{}, ErrInvalidMetricKey
	}
	if !rule.checkTagRegexes() {
		return Rule{}, ErrInvalidCustomRegex
	}
	regex := rule.toRegexp()
	if regex == nil {
		return Rule{}, ErrInvalidPattern
	}
	if regex.NumSubexp() != len(graphitePatternTags) {
		return Rule{}, ErrInvalidCustomRegex
	}
	return Rule{
		rule,
		regex,
		graphitePatternTags,
		metricKeyTags,
	}, nil
}

// MatchRule sees if a given graphite string matches the rule, and if so, returns the generated tag.
func (rule Rule) MatchRule(input string) (api.TaggedMetric, bool) {
	matches := rule.regex.FindStringSubmatch(input)
	if matches == nil {
		return api.TaggedMetric{}, false
	}
	tagSet := api.NewTagSet()
	for index, tagValue := range matches {
		if index == 0 {
			continue
		}
		tagKey := rule.graphitePatternTags[index-1]
		tagSet[tagKey] = tagValue
	}
	interpolatedKey, err := interpolateTags(rule.raw.MetricKeyPattern, tagSet)
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
	return api.TaggedMetric{
		api.MetricKey(interpolatedKey),
		tagSet,
	}, true
}

// ToGraphiteName transforms the given tagged metric back to its graphite metric.
func (rule Rule) ToGraphiteName(taggedMetric api.TaggedMetric) (api.GraphiteMetric, error) {
	interpolatedKey, err := interpolateTags(rule.raw.MetricKeyPattern, taggedMetric.TagSet)
	if err != nil {
		return "", err
	}
	if interpolatedKey != string(taggedMetric.MetricKey) {
		return "", ErrCannotInterpolate
	}
	interpolated, err := interpolateTags(rule.raw.Pattern, taggedMetric.TagSet)
	if err != nil {
		return "", err
	}
	return api.GraphiteMetric(interpolated), nil
}

// MatchRule sees if a given graphite string matches
// any of the specified rules.
func (ruleSet RuleSet) MatchRule(input string) (api.TaggedMetric, bool) {
	for _, rule := range ruleSet.rules {
		value, matched := rule.MatchRule(input)
		if matched {
			return value, matched
		}
	}
	return api.TaggedMetric{}, false
}

// ToGraphiteName transforms the given tagged metric back to its graphite name,
// checking against all the rules.
func (ruleSet RuleSet) ToGraphiteName(taggedMetric api.TaggedMetric) (api.GraphiteMetric, error) {
	for _, rule := range ruleSet.rules {
		reversed, err := rule.ToGraphiteName(taggedMetric)
		if err == nil {
			return reversed, nil
		}
	}
	return "", ErrCannotInterpolate
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

func (rule RawRule) toRegexp() *regexp.Regexp {
	splitted := strings.Split(rule.Pattern, "%")
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
func LoadYAML(input []byte) (RuleSet, error) {
	rawRules := RawRules{}
	if err := yaml.Unmarshal(input, &rawRules); err != nil {
		return RuleSet{}, err
	}
	rules := make([]Rule, len(rawRules.RawRules))
	for index, rawRule := range rawRules.RawRules {
		rule, err := Compile(rawRule)
		if err != nil {
			return RuleSet{}, err
		}
		rules[index] = rule
	}
	return RuleSet{rules}, nil
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

func interpolateTags(pattern string, tagSet api.TagSet) (string, error) {
	if !strings.Contains(pattern, "%") {
		return pattern, nil // short circuit when there are no tags to interpolate.
	}
	splitted := strings.Split(pattern, "%")
	buffer := new(bytes.Buffer)
	for index, token := range splitted {
		if isTagPortion(index) {
			tagValue, hasTag := tagSet[token]
			if hasTag {
				buffer.WriteString(tagValue)
			} else {
				return "", ErrMissingTag
			}
		} else {
			buffer.WriteString(token)
		}
	}
	return buffer.String(), nil
}
