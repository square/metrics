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

package tag

import (
	"math"
	"testing"

	"github.com/square/metrics/api"
	"github.com/square/metrics/testing_support/assert"
)

func TestDropSeries(t *testing.T) {
	a := assert.New(t)
	testCases := []struct {
		Series   api.Timeseries
		Expected map[string]api.Timeseries
	}{
		{
			Series: api.Timeseries{
				Values: []float64{3, 5, 7, 6},
				TagSet: api.TagSet{
					"app": "metrics-indexer",
					"dc":  "north",
					"env": "production",
				},
			},
			Expected: map[string]api.Timeseries{
				"app": {
					Values: []float64{3, 5, 7, 6},
					TagSet: api.TagSet{
						"dc":  "north",
						"env": "production",
					},
				},
				"dc": {
					Values: []float64{3, 5, 7, 6},
					TagSet: api.TagSet{
						"app": "metrics-indexer",
						"env": "production",
					},
				},
				"env": {
					Values: []float64{3, 5, 7, 6},
					TagSet: api.TagSet{
						"app": "metrics-indexer",
						"dc":  "north",
					},
				},
				"fake": {
					Values: []float64{3, 5, 7, 6},
					TagSet: api.TagSet{
						"app": "metrics-indexer",
						"dc":  "north",
						"env": "production",
					},
				},
			},
		},
		{
			Series: api.Timeseries{
				Values: []float64{1, 2, 3},
				TagSet: api.TagSet{
					"host": "q33",
					"dc":   "south",
				},
			},
			Expected: map[string]api.Timeseries{
				"host": {
					Values: []float64{1, 2, 3},
					TagSet: api.TagSet{"dc": "south"},
				},
				"dc": {
					Values: []float64{1, 2, 3},
					TagSet: api.TagSet{"host": "q33"},
				},
				"env": {
					Values: []float64{1, 2, 3},
					TagSet: api.TagSet{"host": "q33", "dc": "south"},
				},
			},
		},
	}
	for _, test := range testCases {
		for dropTag, expected := range test.Expected {
			result := dropTagSeries(test.Series, dropTag)
			a.EqFloatArray(result.Values, expected.Values, 1e-100)
			if !result.TagSet.Equals(expected.TagSet) {
				t.Errorf("failed to drop tag correctly: expected tagset %+v when dropping %s, but got %+v", expected.TagSet, dropTag, result.TagSet)
			}
		}
	}
}

func TestDrop(t *testing.T) {
	timerange, err := api.NewTimerange(1300, 1600, 100)
	if err != nil {
		t.Fatal("invalid timerange used in testcase")
	}
	list := api.SeriesList{
		Timerange: timerange,
		Series: []api.Timeseries{
			{
				Values: []float64{1, 2, 3, 4},
				TagSet: api.TagSet{
					"name": "A",
					"host": "q12",
					"dc":   "east",
				},
			},
			{
				Values: []float64{6, 7, 3, 1},
				TagSet: api.TagSet{
					"name": "B",
					"host": "r2",
					"dc":   "north",
				},
			},
			{
				Values: []float64{2, 4, 6, 8},
				TagSet: api.TagSet{
					"name": "C",
					"host": "q12",
					"dc":   "south",
				},
			},
			{
				Values: []float64{5, math.NaN(), 2, math.NaN()},
				TagSet: api.TagSet{
					"name": "D",
					"host": "q12",
					"dc":   "south",
				},
			},
		},
	}
	result := DropTag(list, "host")
	expect := api.SeriesList{
		Timerange: timerange,
		Series: []api.Timeseries{
			{
				Values: []float64{1, 2, 3, 4},
				TagSet: api.TagSet{
					"name": "A",
					"dc":   "east",
				},
			},
			{
				Values: []float64{6, 7, 3, 1},
				TagSet: api.TagSet{
					"name": "B",
					"dc":   "north",
				},
			},
			{
				Values: []float64{2, 4, 6, 8},
				TagSet: api.TagSet{
					"name": "C",
					"dc":   "south",
				},
			},
			{
				Values: []float64{5, math.NaN(), 2, math.NaN()},
				TagSet: api.TagSet{
					"name": "D",
					"dc":   "south",
				},
			},
		},
	}
	// Verify that result == expect
	a := assert.New(t)
	a.Eq(result.Timerange, expect.Timerange)
	a.EqInt(len(result.Series), len(expect.Series))
	for i := range result.Series {
		// Verify that the two are equal
		seriesResult := result.Series[i]
		seriesExpect := expect.Series[i]
		a.EqFloatArray(seriesResult.Values, seriesExpect.Values, 1e-7)
		if !seriesResult.TagSet.Equals(seriesExpect.TagSet) {
			t.Errorf("Expected series %+v, but got %+v", seriesExpect, seriesResult)
		}
	}
}

func TestSetSeries(t *testing.T) {
	a := assert.New(t)
	newValues := []string{"new", "north", "production"}
	for _, newValue := range newValues {
		testCases := []struct {
			Series   api.Timeseries
			Expected map[string]api.Timeseries
		}{
			{
				Series: api.Timeseries{
					Values: []float64{3, 5, 7, 6},
					TagSet: api.TagSet{
						"app": "metrics-indexer",
						"dc":  "north",
						"env": "production",
					},
				},
				Expected: map[string]api.Timeseries{
					"app": {
						Values: []float64{3, 5, 7, 6},
						TagSet: api.TagSet{
							"app": newValue,
							"dc":  "north",
							"env": "production",
						},
					},
					"dc": {
						Values: []float64{3, 5, 7, 6},
						TagSet: api.TagSet{
							"app": "metrics-indexer",
							"dc":  newValue,
							"env": "production",
						},
					},
					"env": {
						Values: []float64{3, 5, 7, 6},
						TagSet: api.TagSet{
							"app": "metrics-indexer",
							"dc":  "north",
							"env": newValue,
						},
					},
					"fake": {
						Values: []float64{3, 5, 7, 6},
						TagSet: api.TagSet{
							"app":  "metrics-indexer",
							"dc":   "north",
							"env":  "production",
							"fake": newValue,
						},
					},
				},
			},
			{
				Series: api.Timeseries{
					Values: []float64{1, 2, 3},
					TagSet: api.TagSet{
						"host": "q33",
						"dc":   "south",
					},
				},
				Expected: map[string]api.Timeseries{
					"host": {
						Values: []float64{1, 2, 3},
						TagSet: api.TagSet{
							"host": newValue,
							"dc":   "south",
						},
					},
					"dc": {
						Values: []float64{1, 2, 3},
						TagSet: api.TagSet{
							"host": "q33",
							"dc":   newValue,
						},
					},
					"env": {
						Values: []float64{1, 2, 3},
						TagSet: api.TagSet{
							"host": "q33",
							"dc":   "south",
							"env":  newValue,
						},
					},
				},
			},
		}
		for _, test := range testCases {
			for newTag, expected := range test.Expected {
				result := setTagSeries(test.Series, newTag, newValue)
				a.EqFloatArray(result.Values, expected.Values, 1e-100)
				if !result.TagSet.Equals(expected.TagSet) {
					t.Errorf("failed to drop tag correctly: expected tagset %+v when set %s to %s, but got %+v", expected.TagSet, newTag, newValue, result.TagSet)
				}
			}
		}
	}
}

func TestSet(t *testing.T) {
	timerange, err := api.NewTimerange(1300, 1600, 100)
	if err != nil {
		t.Fatal("invalid timerange used in testcase")
	}
	newValue := "east"
	list := api.SeriesList{
		Timerange: timerange,
		Series: []api.Timeseries{
			{
				Values: []float64{1, 2, 3, 4},
				TagSet: api.TagSet{
					"name": "A",
					"host": "q12",
				},
			},
			{
				Values: []float64{6, 7, 3, 1},
				TagSet: api.TagSet{
					"name": "B",
					"host": "r2",
				},
			},
			{
				Values: []float64{2, 4, 6, 8},
				TagSet: api.TagSet{
					"name": "C",
					"host": "q12",
					"dc":   "south",
				},
			},
			{
				Values: []float64{5, math.NaN(), 2, math.NaN()},
				TagSet: api.TagSet{
					"name": "D",
					"host": "q12",
					"dc":   "south",
				},
			},
		},
	}
	result := SetTag(list, "dc", newValue)
	expect := api.SeriesList{
		Timerange: timerange,
		Series: []api.Timeseries{
			{
				Values: []float64{1, 2, 3, 4},
				TagSet: api.TagSet{
					"name": "A",
					"host": "q12",
					"dc":   "east",
				},
			},
			{
				Values: []float64{6, 7, 3, 1},
				TagSet: api.TagSet{
					"name": "B",
					"host": "r2",
					"dc":   "east",
				},
			},
			{
				Values: []float64{2, 4, 6, 8},
				TagSet: api.TagSet{
					"name": "C",
					"host": "q12",
					"dc":   "east",
				},
			},
			{
				Values: []float64{5, math.NaN(), 2, math.NaN()},
				TagSet: api.TagSet{
					"name": "D",
					"host": "q12",
					"dc":   "east",
				},
			},
		},
	}
	// Verify that result == expect
	a := assert.New(t)
	a.Eq(result.Timerange, expect.Timerange)
	a.EqInt(len(result.Series), len(expect.Series))
	for i := range result.Series {
		// Verify that the two are equal
		seriesResult := result.Series[i]
		seriesExpect := expect.Series[i]
		a.EqFloatArray(seriesResult.Values, seriesExpect.Values, 1e-7)
		if !seriesResult.TagSet.Equals(seriesExpect.TagSet) {
			t.Errorf("Expected series %+v, but got %+v", seriesExpect, seriesResult)
		}
	}
}
