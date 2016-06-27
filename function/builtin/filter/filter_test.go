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

package filter

import (
	"math"
	"sort"
	"testing"

	"github.com/square/metrics/api"
	"github.com/square/metrics/function/builtin/aggregate"
	"github.com/square/metrics/testing_support/assert"
)

func TestFilter(t *testing.T) {
	a := assert.New(t)

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
			Values: []float64{4, 4, 3.01, 4, 3.01},
			TagSet: api.TagSet{
				"name": "D",
			},
		},
	}

	list := api.SeriesList{
		Series: []api.Timeseries{series["A"], series["B"], series["C"], series["D"]},
	}
	tests := []struct {
		summary     func([]float64) float64
		lowest      bool
		count       int
		expect      []string
		description string
	}{
		{
			summary:     aggregate.Sum,
			lowest:      true,
			count:       6,
			expect:      []string{"B", "A", "C", "D"},
			description: "sum",
		},

		{
			summary:     aggregate.Sum,
			lowest:      false,
			count:       6,
			expect:      []string{"D", "C", "A", "B"},
			description: "sum",
		},

		{
			summary:     aggregate.Sum,
			lowest:      true,
			count:       4,
			expect:      []string{"B", "A", "C", "D"},
			description: "sum",
		},
		{
			summary:     aggregate.Sum,
			lowest:      true,
			count:       3,
			expect:      []string{"B", "A", "C"},
			description: "sum",
		},
		{
			summary:     aggregate.Sum,
			lowest:      true,
			count:       2,
			expect:      []string{"B", "A"},
			description: "sum",
		},
		{
			summary:     aggregate.Sum,
			lowest:      true,
			count:       1,
			expect:      []string{"B"},
			description: "sum",
		},
		{
			summary:     aggregate.Sum,
			lowest:      false,
			count:       4,
			expect:      []string{"D", "C", "A", "B"},
			description: "sum",
		},
		{
			summary:     aggregate.Sum,
			lowest:      false,
			count:       3,
			expect:      []string{"D", "C", "A"},
			description: "sum",
		},
		{
			summary:     aggregate.Sum,
			lowest:      false,
			count:       2,
			expect:      []string{"D", "C"},
			description: "sum",
		},
		{
			summary:     aggregate.Sum,
			lowest:      false,
			count:       1,
			expect:      []string{"D"},
			description: "sum",
		},
		{
			summary:     aggregate.Max,
			lowest:      false,
			count:       1,
			expect:      []string{"C"},
			description: "max",
		},
		{
			summary:     aggregate.Max,
			lowest:      false,
			count:       2,
			expect:      []string{"C", "D"},
			description: "max",
		},
		{
			summary:     aggregate.Min,
			lowest:      false,
			count:       2,
			expect:      []string{"D", "A"},
			description: "min",
		},
		{
			summary:     aggregate.Min,
			lowest:      false,
			count:       3,
			expect:      []string{"D", "A", "C"},
			description: "min",
		},
	}
	for _, test := range tests {
		filtered := FilterByRecent(list, test.count, test.summary, test.lowest, 100)
		// Verify that every series in the result is from the original.
		// Also verify that we only get the ones we expect.
		if len(filtered.Series) != len(test.expect) {
			t.Errorf("Expected only %d in results but got %d", len(test.expect), len(filtered.Series))
			continue
		}
		for i, s := range filtered.Series {
			original, ok := series[s.TagSet["name"]]
			if !ok {
				t.Fatalf("Result tagset called '%s' is not an original", s.TagSet["name"])
			}
			if s.TagSet["name"] != test.expect[i] {
				testOrder := "highest"
				if test.lowest {
					testOrder = "lowest"
				}
				t.Errorf("((%s %d %s)) Expected filtered sets to be %+v but were:\n%+v", testOrder, test.count, test.description, test.expect, filtered.Series)
				break
			}
			a.EqFloatArray(original.Values, s.Values, 1e-7)
		}
	}
}

func TestFilterRecent(t *testing.T) {
	a := assert.New(t)
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
		Series: series,
	}
	seriesMap := map[string]api.Timeseries{"A": series[0], "B": series[1], "C": series[2], "D": series[3]}
	tests := []struct {
		summary func([]float64) float64
		lowest  bool
		count   int
		slots   int
		expect  []string
	}{
		{
			summary: aggregate.Max,
			lowest:  false,
			count:   50,
			slots:   4,
			expect:  []string{"A", "B", "C", "D"},
		},
		{
			summary: aggregate.Min,
			lowest:  true,
			count:   5,
			slots:   4,
			expect:  []string{"A", "B", "C", "D"},
		},
		{
			summary: aggregate.Mean,
			lowest:  false,
			count:   4,
			slots:   4,
			expect:  []string{"A", "B", "C", "D"},
		},
		{
			summary: aggregate.Max,
			lowest:  false,
			count:   2,
			slots:   4,
			expect:  []string{"A", "B"},
		},
		{
			summary: aggregate.Max,
			lowest:  true,
			count:   2,
			slots:   4,
			expect:  []string{"C", "D"},
		},
		{
			summary: aggregate.Sum,
			lowest:  true,
			count:   1,
			slots:   19, // all points
			expect:  []string{"B"},
		},
		{
			summary: aggregate.Sum,
			lowest:  false,
			count:   1,
			slots:   19,
			expect:  []string{"A"},
		},
	}
	for _, test := range tests {
		filtered := FilterByRecent(list, test.count, test.summary, test.lowest, test.slots)
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

func TestFilterListNaN(t *testing.T) {
	a := assert.New(t)
	array := filterList{nil, nil, true}

	array.index = []int{11, 2, 11, 4, 5, 6}
	array.value = []float64{math.NaN(), 2, math.NaN(), 8, 1, 0}
	sort.Sort(array)
	a.Eq(array.index, []int{6, 5, 2, 4, 11, 11})

	array.index = []int{6, 5, 4, 11, 2, 11}
	array.value = []float64{0, 1, 8, math.NaN(), 2, math.NaN()}
	sort.Sort(array)
	a.Eq(array.index, []int{6, 5, 2, 4, 11, 11})

	array.ascending = false

	array.index = []int{11, 2, 11, 4, 5, 6}
	array.value = []float64{math.NaN(), 2, math.NaN(), 8, 1, 0}
	sort.Sort(array)
	a.Eq(array.index, []int{4, 2, 5, 6, 11, 11})

	array.index = []int{6, 5, 4, 11, 2, 11}
	array.value = []float64{0, 1, 8, math.NaN(), 2, math.NaN()}
	sort.Sort(array)
	a.Eq(array.index, []int{4, 2, 5, 6, 11, 11})
}
