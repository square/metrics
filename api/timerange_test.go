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
	"testing"
	"time"

	"github.com/square/metrics/testing_support/assert"
)

func TestTimerange(t *testing.T) {
	for _, suite := range []struct {
		Start         int64
		End           int64
		Resolution    int64
		ExpectedValid bool
		ExpectedSlots int
	}{
		// valid cases
		{0, 0, 1, true, 1},
		{0, 1, 1, true, 2},
		{0, 100, 1, true, 101},
		{0, 100, 5, true, 21},
		// invalid cases
		{100, 0, 1, false, 0},
		{0, 100, 6, false, 0},
		{0, 100, 200, false, 0},
	} {
		a := assert.New(t).Contextf("input=%d:%d:%d",
			suite.Start,
			suite.End,
			suite.Resolution,
		)
		timerange, err := NewTimerange(suite.Start, suite.End, suite.Resolution)
		a.EqBool(err == nil, suite.ExpectedValid)
		if !suite.ExpectedValid {
			continue
		}

		a.EqInt(timerange.Slots(), suite.ExpectedSlots)
	}
}

func TestTimerange_MarshalJSON(t *testing.T) {
	for _, suite := range []struct {
		input    Timerange
		expected string
	}{
		{Timerange{0, 100, 10}, `{"start":0,"end":100,"resolution":10}`},
		{Timerange{100, 10000, 50}, `{"start":100,"end":10000,"resolution":50}`},
	} {
		a := assert.New(t).Contextf("expected=%s", suite.expected)
		encoded, err := json.Marshal(suite.input)
		a.CheckError(err)
		a.Eq(string(encoded), suite.expected)
	}
}

func TestTimerange_Later(t *testing.T) {
	// Check that when moving forward, when moving backward, etc., time ranges work as expected.
	ranges := []Timerange{
		{
			start:      400,
			end:        900,
			resolution: 100,
		},
		{
			start:      400,
			end:        900,
			resolution: 1,
		},
		{
			start:      120,
			end:        150,
			resolution: 30,
		},
		{
			start:      400,
			end:        520,
			resolution: 40,
		},
	}
	for _, time := range ranges {
		// A sanity check for the above calculations.
		if _, err := NewTimerange(time.start, time.end, time.resolution); err != nil {
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
		for _, timerange := range ranges {
			later := timerange.Shift(time.Duration(offset) * time.Millisecond)
			if later.EndMillis()-later.StartMillis() != timerange.EndMillis()-timerange.StartMillis() || later.ResolutionMillis() != timerange.ResolutionMillis() {
				t.Errorf("Range %+v on offset %d fails; produces %+v", timerange, offset, later)
				continue
			}
			later = timerange.Shift(-time.Duration(offset) * time.Millisecond)
			if later.EndMillis()-later.StartMillis() != timerange.EndMillis()-timerange.StartMillis() || later.ResolutionMillis() != timerange.ResolutionMillis() {
				t.Errorf("Range %+v on offset %d fails; produces %+v", timerange, -offset, later)
				continue
			}
		}
	}
}

func TestTimerangeResample(t *testing.T) {
	type test struct {
		timerange  Timerange
		resolution time.Duration
		expected   Timerange
	}
	tests := []test{
		{
			timerange:  Timerange{200, 300, 10},
			resolution: time.Millisecond * 10,
			expected:   Timerange{200, 300, 10},
		},
		{
			timerange:  Timerange{200, 300, 100},
			resolution: time.Millisecond * 10,
			expected:   Timerange{200, 300, 10},
		},
		{
			timerange:  Timerange{200, 300, 100},
			resolution: time.Millisecond * 200,
			expected:   Timerange{200, 400, 200},
		},
		{
			timerange:  Timerange{201, 399, 1},
			resolution: time.Millisecond * 200,
			expected:   Timerange{200, 400, 200},
		},
		{
			timerange:  Timerange{1199, 1399, 1},
			resolution: time.Millisecond * 200,
			expected:   Timerange{1000, 1400, 200},
		},
		{
			timerange:  Timerange{1199, 1401, 1},
			resolution: time.Millisecond * 200,
			expected:   Timerange{1000, 1600, 200},
		},
		{
			timerange:  Timerange{2839, 4556, 17},
			resolution: time.Millisecond * 100,
			expected:   Timerange{2800, 4600, 100},
		},
	}
	for i, test := range tests {
		assert.New(t).Contextf("Test #%d: %+v.Resample(%+v) => %+v", i+1, test.timerange, test.resolution, test.expected).Eq(test.timerange.Resample(test.resolution), test.expected)
	}
}

func TestTimerangeOnlyAfterExclusive(t *testing.T) {
	type test struct {
		timerange Timerange
		cut       time.Time
		expected  Timerange
		empty     bool
	}
	now := int64(1) * 291720
	makeRange := func(start int64, end int64, resolution int64) Timerange {
		timerange, err := NewTimerange(now+start, now+end, resolution)
		if err != nil {
			t.Fatalf("Problem creating timerange (%d, %d, %d): %s", start, end, resolution, err.Error())
		}
		return timerange
	}
	makeTime := func(after int64) time.Time {
		return time.Unix((now+after)/1000, (now+after)%1000*1e6)
	}
	tests := []test{
		// Near beginnging
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(120),
			expected:  makeRange(200, 300, 10),
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(199),
			expected:  makeRange(200, 300, 10),
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(200),
			expected:  makeRange(210, 300, 10),
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(201),
			expected:  makeRange(210, 300, 10),
		},
		// Middle
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(241),
			expected:  makeRange(250, 300, 10),
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(247),
			expected:  makeRange(250, 300, 10),
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(249),
			expected:  makeRange(250, 300, 10),
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(250),
			expected:  makeRange(260, 300, 10),
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(251),
			expected:  makeRange(260, 300, 10),
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(258),
			expected:  makeRange(260, 300, 10),
		},
		// Near end
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(290),
			expected:  makeRange(300, 300, 10),
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(299),
			expected:  makeRange(300, 300, 10),
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(300),
			empty:     true,
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(301),
			empty:     true,
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(309),
			empty:     true,
		},

		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(310),
			empty:     true,
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(370),
			empty:     true,
		},
	}
	for i, test := range tests {
		a := assert.New(t).Contextf("Test #%d: %+v.OnlyAfterExclusive(%+v)", i+1, test.timerange, test.cut)
		value, nonEmpty := test.timerange.OnlyAfterExclusive(test.cut)
		if test.empty {
			a.Contextf("empty").Eq(nonEmpty, !test.empty)
		} else {
			a.Contextf("nonempty").Eq(nonEmpty, !test.empty)
			a.Contextf("result").Eq(value, test.expected)
		}

	}
}

func TestTimerangeOnlyBeforeInclusive(t *testing.T) {
	type test struct {
		timerange Timerange
		cut       time.Time
		expected  Timerange
		empty     bool
	}
	now := int64(1) * 291720
	makeRange := func(start int64, end int64, resolution int64) Timerange {
		timerange, err := NewTimerange(now+start, now+end, resolution)
		if err != nil {
			t.Fatalf("Problem creating timerange (%d, %d, %d): %s", start, end, resolution, err.Error())
		}
		return timerange
	}
	makeTime := func(after int64) time.Time {
		return time.Unix((now+after)/1000, (now+after)%1000*1e6)
	}
	tests := []test{
		// Near beginnging
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(120),
			empty:     true,
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(199),
			empty:     true,
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(200),
			expected:  makeRange(200, 200, 10),
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(201),
			expected:  makeRange(200, 210, 10),
		},
		// Middle
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(241),
			expected:  makeRange(200, 250, 10),
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(247),
			expected:  makeRange(200, 250, 10),
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(249),
			expected:  makeRange(200, 250, 10),
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(250),
			expected:  makeRange(200, 250, 10),
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(251),
			expected:  makeRange(200, 260, 10),
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(258),
			expected:  makeRange(200, 260, 10),
		},
		// Near end
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(290),
			expected:  makeRange(200, 290, 10),
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(299),
			expected:  makeRange(200, 300, 10),
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(300),
			expected:  makeRange(200, 300, 10),
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(301),
			expected:  makeRange(200, 300, 10),
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(309),
			expected:  makeRange(200, 300, 10),
		},

		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(310),
			expected:  makeRange(200, 300, 10),
		},
		{
			timerange: makeRange(200, 300, 10),
			cut:       makeTime(370),
			expected:  makeRange(200, 300, 10),
		},
	}
	for i, test := range tests {
		a := assert.New(t).Contextf("Test #%d: %+v.OnlyBeforeInclusive(%+v)", i+1, test.timerange, test.cut)
		value, nonEmpty := test.timerange.OnlyBeforeInclusive(test.cut)
		if test.empty {
			a.Contextf("empty").Eq(nonEmpty, !test.empty)
		} else {
			a.Contextf("nonempty").Eq(nonEmpty, !test.empty)
			a.Contextf("result").Eq(value, test.expected)
		}

	}
}
