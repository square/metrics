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

package filter

import (
	"math"
	"testing"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/function/aggregate"
	"github.com/square/metrics/testing_support/assert"
)

func TestFilter(t *testing.T) {
	a := assert.New(t)
	timerange, err := api.NewTimerange(1300, 1700, 100)
	if err != nil {
		t.Fatalf("invalid timerange used in testcase")
	}

	series := map[string]api.Timeseries{
		"A": {
			Values: []float64{3, 3, 3, 3, 3},
			TagSet: api.TagSet{
				"name": "A",
			},
		},
		"B": {
			Values: []float64{1, 2, 2, 1, 0},
			TagSet: api.TagSet{
				"name": "B",
			},
		},
		"C": {
			Values: []float64{1, 2, 3, 4, 5.1},
			TagSet: api.TagSet{
				"name": "C",
			},
		},
		"D": {
			Values: []float64{4, 4, 3, 4, 3},
			TagSet: api.TagSet{
				"name": "D",
			},
		},
		"NaN": {
			Values: []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()},
			TagSet: api.TagSet{
				"name": "NaN",
			},
		},
	}

	list := api.SeriesList{
		Series:    []api.Timeseries{series["NaN"], series["A"], series["B"], series["NaN"], series["C"], series["D"], series["NaN"]},
		Timerange: timerange,
		Name:      "test_series",
	}
	tests := []struct {
		summary func([]float64) float64
		lowest  bool
		count   int
		expect  []string
	}{
		{
			summary: aggregate.Sum,
			lowest:  true,
			count:   6,
			expect:  []string{"B", "A", "C", "D", "NaN", "NaN"},
		},
		{
			summary: aggregate.Mean,
			lowest:  true,
			count:   6,
			expect:  []string{"B", "A", "C", "D", "NaN", "NaN"},
		},
		{
			summary: aggregate.Sum,
			lowest:  false,
			count:   6,
			expect:  []string{"A", "B", "C", "D", "NaN", "NaN"},
		},
		{
			summary: aggregate.Mean,
			lowest:  false,
			count:   6,
			expect:  []string{"A", "B", "C", "D", "NaN", "NaN"},
		},
		{
			summary: aggregate.Sum,
			lowest:  true,
			count:   4,
			expect:  []string{"A", "B", "C", "D"},
		},
		{
			summary: aggregate.Mean,
			lowest:  true,
			count:   4,
			expect:  []string{"A", "B", "C", "D"},
		},
		{
			summary: aggregate.Sum,
			lowest:  true,
			count:   3,
			expect:  []string{"A", "B", "C"},
		},
		{
			summary: aggregate.Mean,
			lowest:  true,
			count:   3,
			expect:  []string{"A", "B", "C"},
		},
		{
			summary: aggregate.Sum,
			lowest:  true,
			count:   2,
			expect:  []string{"A", "B"},
		},
		{
			summary: aggregate.Mean,
			lowest:  true,
			count:   2,
			expect:  []string{"A", "B"},
		},
		{
			summary: aggregate.Sum,
			lowest:  true,
			count:   1,
			expect:  []string{"B"},
		},
		{
			summary: aggregate.Mean,
			lowest:  true,
			count:   1,
			expect:  []string{"B"},
		},
		{
			summary: aggregate.Sum,
			lowest:  false,
			count:   4,
			expect:  []string{"A", "B", "C", "D"},
		},
		{
			summary: aggregate.Mean,
			lowest:  false,
			count:   4,
			expect:  []string{"A", "B", "C", "D"},
		},
		{
			summary: aggregate.Sum,
			lowest:  false,
			count:   3,
			expect:  []string{"A", "C", "D"},
		},
		{
			summary: aggregate.Mean,
			lowest:  false,
			count:   3,
			expect:  []string{"A", "C", "D"},
		},
		{
			summary: aggregate.Sum,
			lowest:  false,
			count:   2,
			expect:  []string{"C", "D"},
		},
		{
			summary: aggregate.Mean,
			lowest:  false,
			count:   2,
			expect:  []string{"C", "D"},
		},
		{
			summary: aggregate.Sum,
			lowest:  false,
			count:   1,
			expect:  []string{"D"},
		},
		{
			summary: aggregate.Mean,
			lowest:  false,
			count:   1,
			expect:  []string{"D"},
		},
		{
			summary: aggregate.Max,
			lowest:  false,
			count:   1,
			expect:  []string{"C"},
		},
		{
			summary: aggregate.Max,
			lowest:  false,
			count:   2,
			expect:  []string{"C", "D"},
		},
		{
			summary: aggregate.Min,
			lowest:  false,
			count:   2,
			expect:  []string{"A", "D"},
		},
		{
			summary: aggregate.Min,
			lowest:  false,
			count:   3,
			expect:  []string{"A", "C", "D"},
		},
	}
	for _, test := range tests {
		filtered := FilterBy(list, test.count, test.summary, test.lowest)
		// Verify that every series in the result is from the original.
		// Also verify that we only get the ones we expect.
		if len(filtered.Series) != len(test.expect) {
			t.Errorf("Expected only %d in results but got %d", len(test.expect), len(filtered.Series))
			continue
		}
		for _, s := range filtered.Series {
			original, ok := series[s.TagSet["name"]]
			if !ok {
				t.Fatalf("Result tagset called '%s' is not an original", s.TagSet["name"])
			}
			a.EqFloatArray(original.Values, s.Values, 1e-7)
		}
		names := map[string]int{}
		for _, name := range test.expect {
			names[name]++
		}
		for _, s := range filtered.Series {
			if names[s.TagSet["name"]] == 0 {
				t.Fatalf("TagSets %+v aren't expected; %+v are", filtered.Series, test.expect)
			}
			names[s.TagSet["name"]]-- // Use up the name so that a seocnd Series can't also use it.
		}
	}
}

func TestFilterRecent(t *testing.T) {
	timerange, err := api.NewTimerange(1300, 2000, 100)
	a := assert.New(t)
	a.CheckError(err)
	series := []api.Timeseries{
		{
			Values: []float64{0, 1, 1, 0, 8, 8, 9, 8},
			TagSet: api.TagSet{"name": "A"},
		},
		{
			Values: []float64{-5, -6, -3, -4, 5, 6, 7, 8},
			TagSet: api.TagSet{"name": "B"},
		},
		{
			Values: []float64{7, 7, 6, 7, 3, 2, 1, 1},
			TagSet: api.TagSet{"name": "C"},
		},
		{
			Values: []float64{6, 5, 5, 5, 2, 2, 3, 3},
			TagSet: api.TagSet{"name": "D"},
		},
	}
	list := api.SeriesList{
		Series:    series,
		Timerange: timerange,
	}
	seriesMap := map[string]api.Timeseries{"A": series[0], "B": series[1], "C": series[2], "D": series[3]}
	tests := []struct {
		summary  func([]float64) float64
		lowest   bool
		count    int
		duration time.Duration
		expect   []string
	}{
		{
			summary:  aggregate.Max,
			lowest:   false,
			count:    50,
			duration: time.Millisecond * 450, // Four points
			expect:   []string{"A", "B", "C", "D"},
		},
		{
			summary:  aggregate.Min,
			lowest:   true,
			count:    5,
			duration: time.Millisecond * 450, // Four points
			expect:   []string{"A", "B", "C", "D"},
		},
		{
			summary:  aggregate.Mean,
			lowest:   false,
			count:    4,
			duration: time.Millisecond * 450, // Four points
			expect:   []string{"A", "B", "C", "D"},
		},
		{
			summary:  aggregate.Max,
			lowest:   false,
			count:    2,
			duration: time.Millisecond * 450, // Four points
			expect:   []string{"A", "B"},
		},
		{
			summary:  aggregate.Max,
			lowest:   true,
			count:    2,
			duration: time.Millisecond * 450, // Four points
			expect:   []string{"C", "D"},
		},
		{
			summary:  aggregate.Sum,
			lowest:   true,
			count:    1,
			duration: time.Millisecond * 9000, // All points
			expect:   []string{"B"},
		},
		{
			summary:  aggregate.Sum,
			lowest:   false,
			count:    1,
			duration: time.Millisecond * 9000, // All points
			expect:   []string{"A"},
		},
	}
	for _, test := range tests {
		filtered := FilterRecentBy(list, test.count, test.summary, test.lowest, test.duration)
		// Verify that they're all unique and expected and unchanged
		a.EqInt(len(filtered.Series), len(test.expect))
		// Next, verify that the names are the same.
		correct := map[string]bool{}
		for _, name := range test.expect {
			correct[name] = true
		}
		for _, series := range filtered.Series {
			name := series.TagSet["name"]
			if !correct[name] {
				t.Errorf("Expected %+v but got %+v", test.expect, filtered.Series)
				break
			}
			correct[name] = false // Delete it so that there can be no repeats.
			a.EqFloatArray(series.Values, seriesMap[name].Values, 1e-7)
		}
	}
}
