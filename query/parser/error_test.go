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

package parser

import "testing"

func TestErrorMessages(t *testing.T) {
	type sample struct {
		query   string
		message string
	}
	tests := []sample{
		{
			query:   "select foo from",
			message: "line 1, column 16: expected value to follow key 'from' in property clause of select statement",
		},
		{
			query:   "select foo\nfrom -30m to now\nwhere app = 'mqe'",
			message: `line 3, column 6: encountered "where" after property clause; "where" blocks must go BEFORE 'from' and 'to' specifiers in property clause of select statement`,
		},
		{
			query:   "select foo\nfrom -30m to mow",
			message: "line 2, column 13: expected value to follow key 'to' in property clause of select statement",
		},
		{
			query:   "select foo + bar[blah='22']\nwhere tag != 'value' and qux = 'qux'\nfrom -30m to mow",
			message: "line 3, column 13: expected value to follow key 'to' in property clause of select statement",
		},
		{
			query:   "select foo + bar[\nwhere tag != 'value' and qux = 'qux'\nfrom -30m to mow",
			message: `line 1, column 18: expected predicate to follow "[" after metric`,
		},
		{
			query:   "select crazy#2dinvalid.metric + bar\nwhere tag != 'value' and qux = 'qux'\nfrom -30m to now",
			message: `line 1, column 13: expected key (one of 'from', 'to', 'resolution', or 'sample by') or end of input but got "#2dinvalid.metric + bar\nwhere tag != 'value' and qux = 'qux'\nfrom -30m to now" following a completed expression`,
		},
		{
			query:   "serlect foo from -30m to now",
			message: `line 1, column 9: expected key (one of 'from', 'to', 'resolution', or 'sample by') or end of input but got "foo from -30m to now" following a completed expression`,
		},
		{
			query:   "describe all where host = 'foo'",
			message: `line 1, column 14: expected end of input after 'describe all' and optional match clause but got "where host = 'foo'"`,
		},
		{
			query:   "select foo, bar,\nfrom -30m to now",
			message: `line 1, column 17: expected expression to follow ","`,
		},
		{
			query:   "select foo, bar[host = 'x' and]\nfrom -30m to now",
			message: `line 1, column 31: expected predicate to follow "and" operator`,
		},
		{
			query:   "select foo, bar[host = 'x' and '2']\nfrom -30m to now",
			message: `line 1, column 31: expected predicate to follow "and" operator`,
		},
	}
	for _, test := range tests {
		_, err := Parse(test.query)
		if err == nil {
			t.Errorf("Expected error to happen in query\n\t%s\nbut not error happened", test.query)
			continue
		}
		actual := err.Error()
		if test.message == "" {
			t.Errorf("In query\n\t%s\ngot out error message\n\t%s\n(no expected message was set)", test.query, actual)
			continue
		}
		if test.message != actual {
			t.Errorf("In query\n\t%s\ngot out error message\n\t%s\nbut expected\n\t%s", test.query, actual, test.message)
		}
	}
}
