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

package query

import (
	"fmt"
	"strings"

	"github.com/square/metrics/api"
)

func (matcher *andPredicate) Apply(tagSet api.TagSet) bool {
	for _, subPredicate := range matcher.predicates {
		if !subPredicate.Apply(tagSet) {
			return false
		}
	}
	return true
}
func (matcher *andPredicate) Query() string {
	substrings := []string{}
	for _, predicate := range matcher.predicates {
		query := predicate.Query()
		if query == "" {
			continue
		}
		substrings = append(substrings, query)
	}
	if len(substrings) == 0 {
		return ""
	}
	if len(substrings) == 1 {
		return substrings[0]
	}
	return fmt.Sprintf("(%s)", strings.Join(substrings, " and "))
}

func (matcher *orPredicate) Apply(tagSet api.TagSet) bool {
	for _, subPredicate := range matcher.predicates {
		if subPredicate.Apply(tagSet) {
			return true
		}
	}
	return false
}
func (matcher *orPredicate) Query() string {
	substrings := []string{}
	for _, predicate := range matcher.predicates {
		query := predicate.Query()
		if query == "" {
			continue
		}
		substrings = append(substrings, query)
	}
	if len(substrings) == 0 {
		return ""
	}
	if len(substrings) == 1 {
		return substrings[0]
	}
	return fmt.Sprintf("(%s)", strings.Join(substrings, " or "))
}

func (matcher *notPredicate) Apply(tagSet api.TagSet) bool {
	return !matcher.predicate.Apply(tagSet)
}
func (matcher *notPredicate) Query() string {
	return fmt.Sprintf("not %s", matcher.predicate.Query())
}

func (matcher *listMatcher) Apply(tagSet api.TagSet) bool {
	if !matchPrecondition(matcher.tag, tagSet) {
		return false
	}
	tagValue := tagSet[matcher.tag]
	for _, match := range matcher.values {
		if match == tagValue {
			return true
		}
	}
	return false
}
func (matcher *listMatcher) Query() string {
	if len(matcher.values) == 1 {
		return fmt.Sprintf("%s = %q", EscapeIdentifier(matcher.tag), matcher.values[0])
	}
	quotedValues := make([]string, len(matcher.values))
	for i, value := range matcher.values {
		quotedValues[i] = fmt.Sprintf("%q", value)
	}
	return fmt.Sprintf("%s in (%s)", EscapeIdentifier(matcher.tag), strings.Join(quotedValues, ", "))
}

func (matcher *regexMatcher) Apply(tagSet api.TagSet) bool {
	if !matchPrecondition(matcher.tag, tagSet) {
		return false
	}
	return matcher.regex.MatchString(tagSet[matcher.tag])
}
func (matcher *regexMatcher) Query() string {
	return fmt.Sprintf("%s match %q", EscapeIdentifier(matcher.tag), matcher.regex.String())
}

func matchPrecondition(matcherTag string, tagSet api.TagSet) bool {
	return tagSet.HasKey(matcherTag)
}
