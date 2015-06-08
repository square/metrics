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

package api

import (
	"testing"

	"github.com/square/metrics/assert"
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
		TagSet{ // Case 0
			"A": "x",
			"B": "y",
			"C": "z",
		},
		TagSet{ // Case 1
			"A": "x",
			"B": "y",
			"C": "z",
		},
		TagSet{ // Case 2
			"A": "q",
			"B": "y",
			"C": "z",
		},
		TagSet{ // Case 3
			"A": "x",
			"C": "z",
		},
		TagSet{ // Case 4
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

func TestTimerange(t *testing.T) {
	for _, suite := range []struct {
		Timerange     Timerange
		ExpectedValid bool
		ExpectedSlots int
	}{
		// valid cases
		{Timerange{0, 0, 1}, true, 1},
		{Timerange{0, 1, 1}, true, 2},
		{Timerange{0, 100, 1}, true, 101},
		{Timerange{0, 100, 5}, true, 21},
		// invalid cases
		{Timerange{100, 0, 1}, false, 0},
		{Timerange{0, 100, 6}, false, 0},
		{Timerange{0, 100, 200}, false, 0},
	} {
		a := assert.New(t).Contextf("input=%d:%d:%d",
			suite.Timerange.start,
			suite.Timerange.end,
			suite.Timerange.resolution,
		)
		a.EqBool(suite.Timerange.isValid(), suite.ExpectedValid)
		// If invalid, nothing else to check
		if !suite.ExpectedValid {
			continue
		}

		a.EqInt(suite.Timerange.Slots(), suite.ExpectedSlots)
	}
}

func TestTimerangeLater(t *testing.T) {
	// Check that when moving forward, when moving backward, etc., time ranges work as expected.
	ranges := []Timerange{
		{
			Start:      400,
			End:        900,
			Resolution: 100,
		},
		{
			Start:      400,
			End:        900,
			Resolution: 1,
		},
		{
			Start:      120,
			End:        150,
			Resolution: 30,
		},
		{
			Start:      400,
			End:        520,
			Resolution: 40,
		},
	}
	for _, time := range ranges {
		// A sanity check for the above calculations.
		if !time.IsValid() {
			panic("Invalid timerange used as test case")
		}
	}
	offsets := []int64{
		0,
		1,
		10,
		100,
		28,
		30,
		40,
		50,
		51,
		56,
		76,
		99,
		100,
		101,
		110,
		140,
		149,
		150,
		151,
		199,
		200,
		201,
	}
	for _, offset := range offsets {
		for _, time := range ranges {
			later := time.Shift(offset)
			if later.End-later.Start != time.End-time.Start || later.Resolution != time.Resolution || !later.IsValid() {
				t.Errorf("Range %+v on offset %d fails; produces %+v", time, offset, later)
				continue
			}
			later = time.Shift(-offset)
			if later.End-later.Start != time.End-time.Start || later.Resolution != time.Resolution || !later.IsValid() {
				t.Errorf("Range %+v on offset %d fails; produces %+v", time, -offset, later)
				continue
			}
		}
	}
}
