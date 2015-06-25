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
	"testing"

	"github.com/square/metrics/api"
	"github.com/square/metrics/assert"
	"github.com/square/metrics/function/aggregate"
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
	}

	list := api.SeriesList{
		Series:    []api.Timeseries{series["A"], series["B"], series["C"], series["D"]},
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
			expect:  []string{"A", "B", "C", "D"},
		},

		{
			summary: aggregate.Sum,
			lowest:  false,
			count:   6,
			expect:  []string{"A", "B", "C", "D"},
		},

		{
			summary: aggregate.Sum,
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
			summary: aggregate.Sum,
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
			summary: aggregate.Sum,
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
			summary: aggregate.Sum,
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
		names := map[string]bool{}
		for _, name := range test.expect {
			names[name] = true
		}
		for _, s := range filtered.Series {
			if !names[s.TagSet["name"]] {
				t.Fatalf("TagSets %+v aren't expected; %+v are", filtered.Series, test.expect)
			}
			names[s.TagSet["name"]] = false // Use up the name so that a seocnd Series can't also use it.
		}
	}
}
