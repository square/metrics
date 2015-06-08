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

package query

import (
	"math"
	"testing"

	"github.com/square/metrics/api"
)

func TestTransformTimeseries(t *testing.T) {
	testCases := []struct {
		values     []float64
		tagSet     api.TagSet
		parameters []value
		scale      float64
		tests      []struct {
			fun      transform
			expected []float64
			useParam bool
		}
	}{
		{
			values: []float64{0, 1, 2, 3, 4, 5},
			tagSet: api.TagSet{
				"dc":   "A",
				"host": "B",
				"env":  "C",
			},
			scale:      30,
			parameters: []value{scalarValue(100)},
			tests: []struct {
				fun      transform
				expected []float64
				useParam bool
			}{
				{
					fun:      transformDerivative,
					expected: []float64{0.0, 1.0 / 30.0, 1.0 / 30.0, 1.0 / 30.0, 1.0 / 30.0, 1.0 / 30.0},
					useParam: false,
				},
				{
					fun:      transformIntegral,
					expected: []float64{0.0, 1.0 * 30.0, 3.0 * 30.0, 6.0 * 30.0, 10.0 * 30.0, 15.0 * 30.0},
					useParam: false,
				},
				{
					fun:      transformMovingAverage,
					expected: []float64{0.0, 0.5, 1.0, 2.0, 3.0, 4.0},
					useParam: true,
				},
				{
					fun:      transformMapMaker("negate", func(x float64) float64 { return -x }),
					expected: []float64{0, -1, -2, -3, -4, -5},
					useParam: false,
				},
			},
		},
	}
	epsilon := 1e-10
	for _, test := range testCases {
		series := api.Timeseries{
			Values: test.values,
			TagSet: test.tagSet,
		}
		for _, transform := range test.tests {
			params := test.parameters
			if !transform.useParam {
				params = []value{}
			}
			result, err := transformTimeseries(series, transform.fun, params, api.Timerange{0, int64(test.scale), int64(test.scale)})
			if err != nil {
				t.Error(err)
				continue
			}
			if !result.TagSet.Equals(test.tagSet) {
				t.Errorf("Expected tagset to be unchanged by transform, changed %+v into %+v", test.tagSet, result.TagSet)
				continue
			}
			if len(result.Values) != len(transform.expected) {
				t.Errorf("Expected result to have length %d but has length %d", transform.expected, result.Values)
				continue
			}
			// Now check that the values are approximately equal
			for i := range result.Values {
				if math.Abs(result.Values[i]-transform.expected[i]) > epsilon {
					t.Errorf("Expected %+v but got %+v", transform.expected, result.Values)
					break
				}
			}
		}
	}
}

func TestApplyTransform(t *testing.T) {
	epsilon := 1e-10
	list := api.SeriesList{
		Series: []api.Timeseries{
			{
				Values: []float64{0, 1, 2, 3, 4, 5},
				TagSet: api.TagSet{
					"series": "A",
				},
			},
			{
				Values: []float64{2, 2, 1, 1, 3, 3},
				TagSet: api.TagSet{
					"series": "B",
				},
			},
			{
				Values: []float64{0, 1, 2, 3, 2, 1},
				TagSet: api.TagSet{
					"series": "C",
				},
			},
		},
		Timerange: api.Timerange{
			Start:      758300,
			End:        758300 + 30*5,
			Resolution: 30,
		},
		Name: "test",
	}
	testCases := []struct {
		transform transform
		parameter []value
		expected  map[string][]float64
	}{
		{
			transform: transformDerivative,
			parameter: []value{},
			expected: map[string][]float64{
				"A": {0, 1.0 / 30, 1.0 / 30, 1.0 / 30, 1.0 / 30, 1.0 / 30},
				"B": {0, 0, -1.0 / 30, 0, 2.0 / 30, 0},
				"C": {0, 1.0 / 30, 1.0 / 30, 1.0 / 30, -1.0 / 30, -1.0 / 30},
			},
		},
		{
			transform: transformIntegral,
			parameter: []value{},
			expected: map[string][]float64{
				"A": {0, 1 * 30, 3 * 30, 6 * 30, 10 * 30, 15 * 30},
				"B": {2 * 30, 4 * 30, 5 * 30, 6 * 30, 9 * 30, 12 * 30},
				"C": {0, 1 * 30, 3 * 30, 6 * 30, 8 * 30, 9 * 30},
			},
		},
		{
			transform: transformCumulative,
			parameter: []value{},
			expected: map[string][]float64{
				"A": {0, 1, 3, 6, 10, 15},
				"B": {2, 4, 5, 6, 9, 12},
				"C": {0, 1, 3, 6, 8, 9},
			},
		},
		{
			transform: transformMovingAverage,
			parameter: []value{scalarValue(100)}, // 100 seconds corresponds to roughly 3 samples
			expected: map[string][]float64{
				"A": {0, 0.5, 1, 2, 3, 4},
				"B": {2.0, 2.0, 5.0 / 3, 4.0 / 3, 5.0 / 3, 7.0 / 3},
				"C": {0, 0.5, 1, 2, 7.0 / 3, 2},
			},
		},
	}
	for _, test := range testCases {
		result, err := ApplyTransform(list, test.transform, test.parameter)
		if err != nil {
			t.Error(err)
			continue
		}
		alreadyUsed := make(map[string]bool)
		for _, series := range result.Series {
			name := series.TagSet["series"]
			expected, ok := test.expected[name]
			if !ok {
				t.Errorf("Series not present in testcase (A, B, or C). Is instead [%s]", name)
				continue
			}
			if alreadyUsed[name] {
				t.Errorf("Multiple series posing as %s", name)
				continue
			}
			alreadyUsed[name] = true
			// Lastly, compare the actual values
			if len(series.Values) != len(expected) {
				t.Errorf("Expected result to have %d entries but has %d entries; for series %s", len(expected), len(series.Values), name)
				continue
			}
			// Check that elements are within epsilon
			for i := range series.Values {
				if math.Abs(series.Values[i]-expected[i]) > epsilon {
					t.Errorf("Expected values for series %s to be %+v but are %+v", name, expected, series.Values)
					break
				}
			}
		}
	}
}

func TestApplyTransformFailure(t *testing.T) {
	list := api.SeriesList{
		Series: []api.Timeseries{
			{
				Values: []float64{0, 1, 2, 3, 4, 5},
				TagSet: api.TagSet{
					"series": "A",
				},
			},
			{
				Values: []float64{2, 2, 1, 1, 3, 3},
				TagSet: api.TagSet{
					"series": "B",
				},
			},
			{
				Values: []float64{0, 1, 2, 3, 2, 1},
				TagSet: api.TagSet{
					"series": "C",
				},
			},
		},
		Timerange: api.Timerange{
			Start:      758300,
			End:        758300 + 30*5,
			Resolution: 30,
		},
		Name: "test",
	}
	testCases := []struct {
		transform transform
		parameter []value
	}{
		{
			transform: transformDerivative,
			parameter: []value{scalarValue(3)},
		},
		{
			transform: transformMapMaker("abs", math.Abs),
			parameter: []value{scalarValue(3)},
		},
		{
			transform: transformMovingAverage,
			parameter: []value{},
		},
	}
	for _, test := range testCases {
		_, err := ApplyTransform(list, test.transform, test.parameter)
		if err == nil {
			t.Errorf("expected failure for testcase %+v", test)
			continue
		}
	}
}
