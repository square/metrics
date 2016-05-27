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

package api

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/square/metrics/testing_support/assert"
)

func TestTagSet_Serialize(t *testing.T) {
	a := assert.New(t)
	a.EqString(NewTagSet().Serialize(), "")
	ts := NewTagSet()
	ts["dc"] = "sjc1b"
	ts["env"] = "production"
	a.EqString(ts.Serialize(), "dc=sjc1b,env=production")
}

func TestTagSetEquals(t *testing.T) {
	sets := []TagSet{
		{ // Case 0
			"A": "x",
			"B": "y",
			"C": "z",
		},
		{ // Case 1
			"A": "x",
			"B": "y",
			"C": "z",
		},
		{ // Case 2
			"A": "q",
			"B": "y",
			"C": "z",
		},
		{ // Case 3
			"A": "x",
			"C": "z",
		},
		{ // Case 4
			"A": "x",
			"C": "z",
		},
	}
	tests := []struct {
		left     int
		right    int
		expected bool
	}{
		{0, 0, true}, // Compare to self
		{1, 1, true},
		{2, 2, true},
		{3, 3, true},
		{4, 4, true},
		{0, 1, true}, // Compare to identical
		{3, 4, true},
		{0, 2, false}, // Compare to different
		{1, 2, false},
		{0, 3, false}, // Compare to missing
		{1, 3, false},
		{0, 4, false},
		{1, 4, false},
	}
	for i, test := range tests {
		if sets[test.left].Equals(sets[test.right]) != test.expected {
			t.Errorf("Test %d on sets %d and %d fails (expected %t)", i, test.left, test.right, test.expected)
			continue
		}
		if sets[test.right].Equals(sets[test.left]) != test.expected {
			t.Errorf("Test %d on sets %d and %d fails (expected %t)", i, test.right, test.left, test.expected)
			continue
		}
	}
}

func TestTagSet_Serialize_Escape(t *testing.T) {
	a := assert.New(t)
	ts := NewTagSet()
	ts["weird=key=1"] = "weird,value"
	ts["weird=key=2"] = "weird\\value"
	a.EqString(ts.Serialize(), "weird\\=key\\=1=weird\\,value,weird\\=key\\=2=weird\\\\value")
	parsed := ParseTagSet(ts.Serialize())
	a.EqInt(len(parsed), 2)
	a.EqString(parsed["weird=key=1"], "weird,value")
	a.EqString(parsed["weird=key=2"], "weird\\value")
}

func TestTagSet_ParseTagSet(t *testing.T) {
	a := assert.New(t)
	a.EqString(ParseTagSet("foo=bar").Serialize(), "foo=bar")
	a.EqString(ParseTagSet("a=1,b=2").Serialize(), "a=1,b=2")
	a.EqString(ParseTagSet("a\\,b=1").Serialize(), "a\\,b=1")
	a.EqString(ParseTagSet("a\\=b=1").Serialize(), "a\\=b=1")
}

func TestTimeseries_MarshalJSON(t *testing.T) {
	for _, suite := range []struct {
		input    Timeseries
		expected string
	}{
		{
			Timeseries{
				TagSet: ParseTagSet("foo=bar"),
				Values: []float64{0, 1, -1, math.NaN()},
			},
			`{"tagset":{"foo":"bar"},"values":[0,1,-1,null]}`,
		},
		{
			Timeseries{
				TagSet: NewTagSet(),
				Values: []float64{0, 1, -1, math.NaN()},
			},
			`{"tagset":{},"values":[0,1,-1,null]}`,
		},
	} {
		a := assert.New(t).Contextf("expected=%s", suite.expected)
		encoded, err := json.Marshal(suite.input)
		a.CheckError(err)
		a.Eq(string(encoded), suite.expected)
	}
}
