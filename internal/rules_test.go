package internal

import (
	"square/vis/metrics-indexer/assert"
	"testing"
)

func TestCompile_Good(t *testing.T) {
	a := assert.New(t)
	_, err := Compile(RawRule{
		Pattern:   "prefix.%foo%",
		MetricKey: "test-metric",
	})
	a.CheckError(err)
}

func TestCompile_InvalidMetric(t *testing.T) {
	_, err := Compile(RawRule{
		Pattern:   "prefix.%foo%",
		MetricKey: "",
	})
	if err != ErrInvalidMetricKey {
		t.Errorf("Expected error, but something else happened.")
	}
}

func TestCompile_InvalidPattern(t *testing.T) {
	_, err := Compile(RawRule{
		Pattern:   "prefix.%foo%abc%",
		MetricKey: "test-metric",
	})
	if err != ErrInvalidPattern {
		t.Errorf("Expected error, but something else happened.")
	}
	_, err = Compile(RawRule{
		Pattern:   "",
		MetricKey: "test-metric",
	})
	if err != ErrInvalidPattern {
		t.Errorf("Expected error, but something else happened.")
	}
	_, err = Compile(RawRule{
		Pattern:   "prefix.%foo%.%foo%",
		MetricKey: "test-metric",
	})
	if err != ErrInvalidPattern {
		t.Errorf("Expected error, but something else happened.")
	}
	_, err = Compile(RawRule{
		Pattern:   "prefix.%foo%.abc.%%",
		MetricKey: "test-metric",
	})
	if err != ErrInvalidPattern {
		t.Errorf("Expected error, but something else happened.")
	}
}

func TestCompile_InvalidCustomRegex(t *testing.T) {
	regex := make(map[string]string)
	regex["foo"] = "(bar)"
	_, err := Compile(RawRule{
		Pattern:   "prefix.%foo%",
		MetricKey: "test-metric",
		Regex:     regex,
	})
	if err != ErrInvalidCustomRegex {
		t.Errorf("Expected error, but something else happened.")
	}
}

func TestMatchRule_Simple(t *testing.T) {
	a := assert.New(t)
	rule, err := Compile(RawRule{
		Pattern:   "prefix.%foo%",
		MetricKey: "test-metric",
	})
	a.CheckError(err)

	_, matches := rule.MatchRule("")
	if matches {
		t.Errorf("Unexpected matching")
	}
	matcher, matches := rule.MatchRule("prefix.abc")
	if !matches {
		t.Errorf("Expected matching but didn't occur")
	}
	a.EqString(string(matcher.MetricKey), "test-metric")
	a.EqString(matcher.TagSet["foo"], "abc")

	_, matches = rule.MatchRule("prefix.abc.def")
	if matches {
		t.Errorf("Unexpected matching")
	}
}

func TestMatchRule_CustomRegex(t *testing.T) {
	a := assert.New(t)
	regex := make(map[string]string)
	regex["name"] = "[a-z]+"
	regex["shard"] = "[0-9]+"
	rule, err := Compile(RawRule{
		Pattern:   "feed.%name%-shard-%shard%",
		MetricKey: "test-feed-metric",
		Regex:     regex,
	})
	a.CheckError(err)

	_, matches := rule.MatchRule("")
	if matches {
		t.Errorf("Unexpected matching")
	}
	matcher, matches := rule.MatchRule("feed.feedname-shard-12")
	if !matches {
		t.Errorf("Expected matching but didn't occur")
	}
	a.EqString(string(matcher.MetricKey), "test-feed-metric")
	a.EqString(matcher.TagSet["name"], "feedname")
	a.EqString(matcher.TagSet["shard"], "12")
}

func TestLoadYAML(t *testing.T) {
	a := assert.New(t)
	rawYAML := `
rules:
  -
    pattern: foo.bar.baz.%tag%
    metric_key: abc
    regex: {}
  `
	ruleSet, err := LoadYAML([]byte(rawYAML))
	a.CheckError(err)
	a.EqInt(len(ruleSet.rules), 1)
	a.EqString(string(ruleSet.rules[0].raw.MetricKey), "abc")
	a.Eq(ruleSet.rules[0].tags, []string{"tag"})
}
