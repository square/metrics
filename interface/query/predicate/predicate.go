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

package predicate

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/square/metrics/api"
	"github.com/square/metrics/util"
)

// Predicate is a boolean function applied against the given
// metric alias and tagset. It determines whether the given metric
// should be included in the query.
type Predicate interface {
	// checks the matcher.
	Apply(tagSet api.TagSet) bool
	Query() string
}

// TruePredicate is always true
type TruePredicate struct{}

func (_ TruePredicate) Apply(_ api.TagSet) bool {
	return true
}
func (_ TruePredicate) Query() string {
	return "true"
}

// FalsePredicate is always false
type FalsePredicate struct{}

func (_ FalsePredicate) Apply(_ api.TagSet) bool {
	return false
}

func (_ FalsePredicate) Query() string {
	return "false"
}

// All takes a slice of predicates, removes the nil values, and returns the result in an AndPredicate.
// Note that if all values passed are nil, it's the True constant predicate.
func All(predicates ...Predicate) Predicate {
	result := []Predicate{}
	for _, p := range predicates {
		if p != nil {
			result = append(result, p)
		}
	}
	if len(result) == 1 {
		return result[0]
	}
	if len(result) == 0 {
		return TruePredicate{}
	}
	return AndPredicate{
		Predicates: result,
	}
}

type AndPredicate struct {
	Predicates []Predicate
}

func (p AndPredicate) Apply(tagset api.TagSet) bool {
	for _, predicate := range p.Predicates {
		if !predicate.Apply(tagset) {
			return false
		}
	}
	return true
}
func (p AndPredicate) Query() string {
	list := []string{}
	for _, predicate := range p.Predicates {
		list = append(list, predicate.Query())
	}
	return fmt.Sprintf("(%s)", strings.Join(list, " and "))
}

// Any filters out nil predicates and then returns the result in an OrPredicate.
// Note that if all values passed are nil, it's the False constant predicate.
func Any(predicates ...Predicate) Predicate {
	result := []Predicate{}
	for _, p := range predicates {
		if p != nil {
			result = append(result, p)
		}
	}
	if len(result) == 1 {
		return result[0]
	}
	if len(result) == 0 {
		return FalsePredicate{}
	}
	return OrPredicate{
		Predicates: result,
	}
}

type OrPredicate struct {
	Predicates []Predicate
}

func (p OrPredicate) Apply(tagset api.TagSet) bool {
	for _, predicate := range p.Predicates {
		if predicate.Apply(tagset) {
			return true
		}
	}
	return false
}
func (p OrPredicate) Query() string {
	list := []string{}
	for _, predicate := range p.Predicates {
		list = append(list, predicate.Query())
	}
	return fmt.Sprintf("(%s)", strings.Join(list, " or "))
}

type NotPredicate struct {
	Predicate Predicate
}

func (p NotPredicate) Apply(tagset api.TagSet) bool {
	return !p.Predicate.Apply(tagset)
}
func (p NotPredicate) Query() string {
	return fmt.Sprintf("not %s", p.Predicate.Query())
}

type ListMatcher struct {
	Tag    string
	Values []string
}

func (p ListMatcher) Apply(tagset api.TagSet) bool {
	value, ok := tagset[p.Tag]
	if !ok {
		return false
	}
	for _, accept := range p.Values {
		if accept == value {
			return true
		}
	}
	return false
}
func (p ListMatcher) Query() string {
	if len(p.Values) == 1 {
		return fmt.Sprintf("%s = %q", util.EscapeIdentifier(p.Tag), p.Values[0])
	}
	quotedValues := make([]string, len(p.Values))
	for i, value := range p.Values {
		quotedValues[i] = fmt.Sprintf("%q", value)
	}
	return fmt.Sprintf("%s in (%s)", util.EscapeIdentifier(p.Tag), strings.Join(quotedValues, ", "))
}

type RegexMatcher struct {
	Tag   string
	Regex *regexp.Regexp
}

func (p RegexMatcher) Apply(tagset api.TagSet) bool {
	return tagset.HasKey(p.Tag) && p.Regex.MatchString(tagset[p.Tag])
}
func (p RegexMatcher) Query() string {
	return fmt.Sprintf("%s match %q", util.EscapeIdentifier(p.Tag), p.Regex.String())
}
