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
	// ErrInvalidMetricKey is retruned when an invalid metric key is provided.
	ErrInvalidMetricKey = errors.New("Invalid Metric Key")
	// ErrInvalidCustomRegex is retruned when the custom regex is invalid.
	ErrInvalidCustomRegex = errors.New("Invalid Custom Regex")
	// ErrMissingTag is returned during the reverse mapping, when a tag is not provided.
	ErrMissingTag = errors.New("Missing Tag")
	// ErrCannotReverse is returned during the reverse mapping, when a given mapping cannot be reversed.
	ErrCannotReverse = errors.New("Cannot Reverse")
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
	if !rule.checkRegex() {
		return Rule{}, ErrInvalidCustomRegex
	}
	regex := rule.toRegexp()
	if regex == nil {
		return Rule{}, ErrInvalidPattern
	}
	if regex.NumSubexp() != len(tags) {
		return Rule{}, ErrInvalidCustomRegex
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

// Reverse transforms the given tagged metric back to its graphite metric.
func (rule Rule) Reverse(taggedMetric api.TaggedMetric) (api.GraphiteMetric, error) {
	if rule.raw.MetricKey != taggedMetric.MetricKey {
		return "", ErrCannotReverse
	}
	splitted := strings.Split(rule.raw.Pattern, "%")
	buffer := new(bytes.Buffer)
	for index, token := range splitted {
		if isTagPortion(index) {
			tagValue, hasTag := taggedMetric.TagSet[token]
			if hasTag {
				buffer.WriteString(tagValue)
			} else {
				return "", ErrMissingTag
			}
		} else {
			buffer.WriteString(token)
		}
	}
	return api.GraphiteMetric(buffer.String()), nil
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

// Reverse transforms the given tagged metric back to its graphite name,
// checking against all the rules.
func (ruleSet RuleSet) Reverse(taggedMetric api.TaggedMetric) (api.GraphiteMetric, error) {
	for _, rule := range ruleSet.rules {
		reversed, err := rule.Reverse(taggedMetric)
		if err == nil {
			return reversed, nil
		}
	}
	return "", ErrCannotReverse
}

// AllKeys returns list of all metric keys defined in the system.
func (ruleSet RuleSet) AllKeys() []api.MetricKey {
	metrics := make([]api.MetricKey, 0, len(ruleSet.rules))
	for _, rule := range ruleSet.rules {
		metrics = append(metrics, rule.raw.MetricKey)
	}
	return metrics
}

func (rule RawRule) checkRegex() bool {
	for _, regex := range rule.Regex {
		compiled, err := regexp.Compile(regex)
		if err != nil {
			return false
		}
		if compiled.NumSubexp() > 0 {
			return false
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
