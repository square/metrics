// Copyright 2015 - 2016 Square Inc.
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
	"testing"

	"github.com/square/metrics/api"
	"github.com/square/metrics/testing_support/assert"
)

func checkRuleErrorCode(a assert.Assert, err error, expected RuleErrorCode) {
	a = a.Stack(1)
	if err == nil {
		a.Errorf("No error provided.")
		return
	}
	casted, ok := err.(RuleError)
	if !ok {
		a.Errorf("Invalid Error type: %s", err.Error())
		return
	}
	a.EqInt(int(casted.Code()), int(expected))
}

func checkConversionErrorCode(t *testing.T, err error, expected ConversionErrorCode) {
	casted, ok := err.(ConversionError)
	if !ok {
		t.Errorf("Invalid Error type")
		return
	}
	a := assert.New(t)
	a.EqInt(int(casted.Code()), int(expected))
}

func TestCompile_Good(t *testing.T) {
	a := assert.New(t)
	_, err := Compile(RawRule{
		Pattern:          "prefix.%foo%",
		MetricKeyPattern: "test-metric",
	})
	a.CheckError(err)
}

func TestCompile_Error(t *testing.T) {
	for _, test := range []struct {
		rawRule      RawRule
		expectedCode RuleErrorCode
	}{
		{RawRule{Pattern: "prefix.%foo%", MetricKeyPattern: ""}, InvalidMetricKey},
		{RawRule{Pattern: "prefix.%foo%abc%", MetricKeyPattern: "test-metric"}, InvalidPattern},
		{RawRule{Pattern: "", MetricKeyPattern: "test-metric"}, InvalidPattern},
		{RawRule{Pattern: "prefix.%foo%.%foo%", MetricKeyPattern: "test-metric"}, InvalidPattern},
		{RawRule{Pattern: "prefix.%foo%.abc.%%", MetricKeyPattern: "test-metric"}, InvalidPattern},
		{RawRule{Pattern: "prefix.%foo%", MetricKeyPattern: "test-metric", Regex: map[string]string{"foo": "(bar)"}}, InvalidCustomRegex},
	} {
		_, err := Compile(test.rawRule)
		a := assert.New(t).Contextf("%s", test.rawRule.Pattern)
		checkRuleErrorCode(a, err, test.expectedCode)
	}
}

func TestMatchRule_Simple(t *testing.T) {
	a := assert.New(t)
	rule, err := Compile(RawRule{
		Pattern:          "prefix.%foo%",
		MetricKeyPattern: "test-metric",
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

func TestMatchRule_FilterTag(t *testing.T) {
	a := assert.New(t)
	rule, err := Compile(RawRule{
		Pattern:          "prefix.%foo%.%bar%",
		MetricKeyPattern: "test-metric.%bar%",
	})
	a.CheckError(err)
	originalName := "prefix.fooValue.barValue"
	matcher, matched := rule.MatchRule(originalName)
	if !matched {
		t.Errorf("Expected matching but didn't occur")
		return
	}
	a.EqString(string(matcher.MetricKey), "test-metric.barValue")
	a.Eq(matcher.TagSet, api.TagSet(map[string]string{"foo": "fooValue"}))
	// perform the reverse.
	reversed, err := rule.ToGraphiteName(matcher)
	a.CheckError(err)
	a.EqString(string(reversed), originalName)
}

func TestMatchRule_CustomRegex(t *testing.T) {
	a := assert.New(t)
	regex := make(map[string]string)
	regex["name"] = "[a-z]+"
	regex["shard"] = "[0-9]+"
	rule, err := Compile(RawRule{
		Pattern:          "feed.%name%-shard-%shard%",
		MetricKeyPattern: "test-feed-metric",
		Regex:            regex,
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
	a.EqInt(len(ruleSet.Rules), 1)
	a.EqString(string(ruleSet.Rules[0].raw.MetricKeyPattern), "abc")
	a.Eq(ruleSet.Rules[0].graphitePatternTags, []string{"tag"})
}

func TestLoadYAML_Invalid(t *testing.T) {
	a := assert.New(t)
	rawYAML := `
rules
  -
    pattern: foo.bar.baz.%tag%
    metric_key: abc
    regex: {}
  `
	ruleSet, err := LoadYAML([]byte(rawYAML))
	checkRuleErrorCode(a, err, InvalidYaml)
	a.EqInt(len(ruleSet.Rules), 0)
}

func TestToGraphiteName(t *testing.T) {
	a := assert.New(t)
	rule, err := Compile(RawRule{
		Pattern:          "prefix.%foo%",
		MetricKeyPattern: "test-metric",
	})
	a.CheckError(err)
	tm := api.TaggedMetric{
		MetricKey: "test-metric",
		TagSet:    api.TagSet{"foo": "fooValue"},
	}
	reversed, err := rule.ToGraphiteName(tm)
	a.CheckError(err)
	a.EqString(string(reversed), "prefix.fooValue")
}

func TestToGraphiteName_Error(t *testing.T) {
	a := assert.New(t)
	rule, err := Compile(RawRule{
		Pattern:          "prefix.%foo%",
		MetricKeyPattern: "test-metric",
	})
	a.CheckError(err)
	reversed, err := rule.ToGraphiteName(api.TaggedMetric{
		MetricKey: "test-metric",
		TagSet:    api.TagSet{},
	})
	checkConversionErrorCode(t, err, MissingTag)
	a.EqString(string(reversed), "")

	reversed, err = rule.ToGraphiteName(api.TaggedMetric{
		MetricKey: "test-metric-foo",
		TagSet:    api.TagSet{"foo": "fooValue"},
	})
	checkConversionErrorCode(t, err, CannotInterpolate)
	a.EqString(string(reversed), "")
}

func Test_interpolateTags(t *testing.T) {

	for _, testCase := range []struct {
		pattern  string
		tagSet   api.TagSet
		enforce  bool
		result   string
		succeeds bool
	}{
		// note that the result <fail> indicates that the test case should fail to parse
		{"%A%.%B%.foo.bar.%C%", map[string]string{"A": "cat", "B": "dog", "C": "box"}, false, "cat.dog.foo.bar.box", true},
		{"%A%.%B%.foo.bar.%C%", map[string]string{"A": "cat", "B": "dog", "C": "box"}, true, "cat.dog.foo.bar.box", true},
		{"%A%.%B%.foo.bar.%C%", map[string]string{"A": "cat", "B": "dog", "C": "box", "D": "other"}, false, "cat.dog.foo.bar.box", true},
		{"%A%.%B%.foo.bar.%C%", map[string]string{"A": "cat", "B": "dog", "C": "box", "D": "other"}, true, "", false},
		{"no.variable.test", map[string]string{"A": "cat", "B": "dog", "C": "box"}, false, "no.variable.test", true},
		{"no.variable.test", map[string]string{"A": "cat", "B": "dog", "C": "box"}, true, "", false},
		{"test.for.%extra%", map[string]string{"A": "cat", "B": "dog", "C": "box"}, false, "", false},
		{"test.for.%extra%", map[string]string{"A": "cat", "B": "dog", "C": "box"}, true, "", false},
	} {
		pattern := testCase.pattern
		tagSet := testCase.tagSet
		result := testCase.result
		enforce := testCase.enforce
		succeeds := testCase.succeeds
		testResult, err := interpolateTags(pattern, tagSet, enforce)
		if succeeds {
			if err != nil {
				t.Errorf("pattern %s fails for tagset %+v", pattern, tagSet)
				continue
			}
			if testResult != result {
				t.Errorf("pattern %s for tagset %+v produces incorrect pattern %s instead of %s (enforce=%v)", pattern, tagSet, testResult, result, enforce)
				continue
			}
			// otherwise, everything is okay since no error occurred and the results match
		} else {
			if err == nil {
				t.Errorf("pattern %s succeeds for tagset %+v producing output %s when it should not succeed (enforce=%v)", pattern, tagSet, testResult, enforce)
				continue
			}
			// otherwise, everything is okay since the match failed
		}
	}

}

func TestDoNotMatchRegex(t *testing.T) {
	rule, err := Compile(RawRule{
		Pattern:          `%foo%.%animal%.%color%`,
		MetricKeyPattern: `%foo%.%color%`,
		DoNotMatch: map[string]string{
			`animal`: `stuffed|teddy`,
			`color`:  `z{4}|qy+`,
		},
	})
	if err != nil {
		t.Fatalf("Unexpected error %s", err.Error())
	}
	tests := []struct {
		input   string
		success bool
	}{
		{"bar.dog.green", true},
		{"qux.cat.yellow", true},
		{"foo.stuffed-tiger.blue", false},
		{"foo.striped-tiger.blue", true},
		{"foo.bar.zzzz", false},
		{"foo.bar.zzz", true},
		{"foo.bar.abcdqefgh", true},
		{"foo.bar.abcdqyyyefgh", false},
		{"foo.teddy-bear.qqyyzzzz", false},
	}
	for _, test := range tests {
		if _, success := rule.MatchRule(test.input); success != test.success {
			t.Errorf("Expected success=%t on input `%s`", test.success, test.input)
		}
	}
}

func TestDoNotMatchRegexReverse(t *testing.T) {
	rule, err := Compile(RawRule{
		Pattern:          `%foo%.%animal%.%color%`,
		MetricKeyPattern: `%foo%.%color%`,
		DoNotMatch: map[string]string{
			`animal`: `stuffed|teddy`,
			`color`:  `z{4}|qy+`,
		},
	})
	if err != nil {
		t.Fatalf("Unexpected error %s", err.Error())
	}
	tests := []struct {
		input    api.TaggedMetric
		graphite string
		success  bool
	}{
		{
			api.TaggedMetric{
				MetricKey: "qux.red",
				TagSet: map[string]string{
					`animal`: `elephant`,
				},
			},
			"qux.elephant.red",
			true,
		},
		{
			api.TaggedMetric{
				MetricKey: "qux.qyyy",
				TagSet: map[string]string{
					`animal`: `elephant`,
				},
			},
			"",
			false,
		},
		{
			api.TaggedMetric{
				MetricKey: "bar.red",
				TagSet: map[string]string{
					`animal`: `teddy-bear`,
				},
			},
			"",
			false,
		},
	}
	for _, test := range tests {
		result, err := rule.ToGraphiteName(test.input)
		if test.success {
			if err != nil {
				t.Errorf("Expected success but failed: test %+v; result %s", test, err.Error())
			}
		} else {
			if err == nil {
				t.Errorf("Expected failure to convert for test %+v but got %s", test, result)
			}
		}
	}
}
