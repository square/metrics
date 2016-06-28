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

package server

import (
	"regexp"
	"testing"

	"github.com/square/metrics/query/predicate"
	"github.com/square/metrics/testing_support/assert"
)

func TestPredicateFromConstraint(t *testing.T) {
	a := assert.New(t)
	tests := []struct {
		constraint Constraint
		err        string
		result     predicate.Predicate
	}{
		{
			constraint: Constraint{
				All: []Constraint{
					{
						KeyIs: &KeyIs{
							Key:   "dc",
							Value: "west",
						},
					},
					{
						KeyMatch: &KeyMatch{
							Key:   "host",
							Regex: `server_a[0-9]+`,
						},
					},
				},
			},
			result: predicate.All(
				predicate.ListMatcher{
					Tag:    "dc",
					Values: []string{"west"},
				},
				predicate.RegexMatcher{
					Tag:   "host",
					Regex: regexp.MustCompile(`server_a[0-9]+`),
				},
			),
		},
		{
			constraint: Constraint{
				Not: &Constraint{
					Any: []Constraint{
						{
							Not: &Constraint{
								KeyIn: &KeyIn{
									Key:    "host",
									Values: []string{"host14", "host15", "host16"},
								},
							},
						},
						{
							KeyIs: &KeyIs{
								Key:   "app",
								Value: "mqe",
							},
						},
						{
							Not: &Constraint{
								All: []Constraint{
									{
										KeyMatch: &KeyMatch{
											Key:   "test1",
											Regex: `blah\+`,
										},
									},
									{
										KeyIs: &KeyIs{
											Key:   "test2",
											Value: "blah",
										},
									},
								},
							},
						},
					},
				},
			},
			result: predicate.NotPredicate{
				Predicate: predicate.Any(

					predicate.NotPredicate{
						Predicate: predicate.ListMatcher{
							Tag:    "host",
							Values: []string{"host14", "host15", "host16"},
						},
					},

					predicate.ListMatcher{
						Tag:    "app",
						Values: []string{"mqe"},
					},

					predicate.NotPredicate{
						Predicate: predicate.All(
							predicate.RegexMatcher{
								Tag:   "test1",
								Regex: regexp.MustCompile(`blah\+`),
							},
							predicate.ListMatcher{
								Tag:    "test2",
								Values: []string{"blah"},
							},
						),
					},
				),
			},
		},
		// Now for invalid inputs
		{
			err: "Multiple assignments of sub-expressions",
			constraint: Constraint{
				Not: &Constraint{
					Any: []Constraint{
						{
							Not: &Constraint{
								KeyIn: &KeyIn{
									Key:    "host",
									Values: []string{"host14", "host15", "host16"},
								},
							},
						},
						{
							KeyIs: &KeyIs{
								Key:   "app",
								Value: "mqe",
							},
							KeyIn: &KeyIn{ // Illegal - can't set both KeyIn and KeyIs
								Key:    "app",
								Values: []string{"mqe"},
							},
						},
						{
							Not: &Constraint{
								All: []Constraint{
									{
										KeyMatch: &KeyMatch{
											Key:   "test1",
											Regex: `blah\+`,
										},
									},
									{
										KeyIs: &KeyIs{
											Key:   "test2",
											Value: "blah",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			err:        "zero value is not a legal Constraint",
			constraint: Constraint{},
		},
		{
			err: "invalid regexp",
			constraint: Constraint{
				KeyMatch: &KeyMatch{
					Key:   "test",
					Regex: `[open`,
				},
			},
		},
	}

	for i, test := range tests {
		result, err := predicateFromConstraint(test.constraint)
		if test.err != "" {
			if err == nil {
				t.Errorf("Expected test %d to cause error (expected %s), but didn't.", i, test.err)
			}
			continue
		}
		if err != nil {
			t.Errorf("Unexpected error in test %d: %s", i, err.Error())
			continue
		}
		a.Contextf("test %d", i).Eq(result, test.result)
	}
}
