package internal

import (
	"bytes"
	"errors"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
	"square/vis/metrics-indexer/api"
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
	Regex     map[string]string `yaml:"regex,omitempty"`
}

// RawRules is list of RawRule
type RawRules struct {
	RawRules []RawRule `yaml:"rules"`
}

// Rule is a sanitized version of RawRule. Only valid rules
// can be converted to Rule.
type Rule struct {
	raw   RawRule
	regex *regexp.Regexp
	tags  []string
}

// RuleSet is a sanitized version of RawRules.
// Rules are matched sequentially until a correct one is matched.
type RuleSet struct {
	rules []Rule
}

// Compile a given RawRule into a regex and exposed tagset.
func Compile(rule RawRule) (Rule, error) {
	if len(rule.MetricKey) == 0 {
		return Rule{}, ErrInvalidMetricKey
	}
	tags := rule.extractTags()
	if tags == nil {
		return Rule{}, ErrInvalidPattern
	}
	regex := rule.toRegexp()
	if regex == nil {
		return Rule{}, ErrInvalidPattern
	}
	return Rule{
		rule,
		regex,
		tags,
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
		tagKey := rule.tags[index-1]
		tagSet[tagKey] = tagValue
	}
	return api.TaggedMetric{
		rule.raw.MetricKey,
		tagSet,
	}, true
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

// LoadYAML loads a RuleSet from the byte array of the YAML file.
func LoadYAML(input []byte) (RuleSet, error) {
	rawRules := RawRules{}
	if err := yaml.Unmarshal(input, &rawRules); err != nil {
		return RuleSet{}, err
	}
	rules := make([]Rule, len(rawRules.RawRules))
	for index, rawRule := range rawRules.RawRules {
		if rule, err := Compile(rawRule); err != nil {
			return RuleSet{}, err
		} else {
			rules[index] = rule
		}
	}
	return RuleSet{rules}, nil
}

func isTagPortion(index int) bool {
	return index%2 == 1
}
