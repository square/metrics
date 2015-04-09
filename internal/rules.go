package internal

import (
	"bytes"
	"errors"
	"regexp"
	"square/vis/metrics-indexer/api"
	"strings"
)

var defaultRegex = "[^.]+"

var (
	// ErrInvalidPattern is returned when an invalid rule pattern is provided.
	ErrInvalidPattern = errors.New("Invalid Pattern")
	// ErrInvalidMetricKey  is retruned when an invalid metric key is provided.
	ErrInvalidMetricKey = errors.New("Invalid Metric Key")
)

// RawRule is the input provided by the YAML file to specify the rul.
type RawRule struct {
	Pattern   string            `yaml:"pattern"`
	MetricKey api.MetricKey     `yaml:"metric_key"`
	Regex     map[string]string `yaml:"regex"`
}

// CompiledRule is sanitized version of RawRule. Only valid rules
// can be converted to CompiledRule.
type CompiledRule struct {
	rule  RawRule
	regex *regexp.Regexp
	tags  []string
}

// Compile a given RawRule into a regex and exposed tagset.
func Compile(rule RawRule) (CompiledRule, error) {
	if len(rule.MetricKey) == 0 {
		return CompiledRule{}, ErrInvalidMetricKey
	}
	tags := rule.extractTags()
	if tags == nil {
		return CompiledRule{}, ErrInvalidPattern
	}
	regex := rule.toRegexp()
	if regex == nil {
		return CompiledRule{}, ErrInvalidPattern
	}
	return CompiledRule{
		rule,
		regex,
		tags,
	}, nil
}

// MatchRule sees if a given graphite string matches the rule, and if so, returns the generated tag.
func (compiledRule CompiledRule) MatchRule(input string) (api.TaggedMetric, bool) {
	matches := compiledRule.regex.FindStringSubmatch(input)
	if matches == nil {
		return api.TaggedMetric{}, false
	}
	tagSet := api.NewTagSet()
	for index, tagValue := range matches {
		if index == 0 {
			continue
		}
		tagKey := compiledRule.tags[index-1]
		tagSet[tagKey] = tagValue
	}
	return api.TaggedMetric{
		compiledRule.rule.MetricKey,
		tagSet,
	}, true
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

func (rule RawRule) extractTags() []string {
	if len(rule.Pattern) == 0 {
		return nil // empty pattern is not allowed.
	}
	splitted := strings.Split(rule.Pattern, "%")
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

func isTagPortion(index int) bool {
	return index%2 == 1
}
