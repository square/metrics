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
	"fmt"
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
				{
					fun:      transformNaNKeepLast,
					expected: []float64{0, 1, 2, 3, 4, 5},
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
			result, err := transformTimeseries(series, transform.fun, params, test.scale)
			if err != nil {
				t.Error(err)
				continue
			}
			if !result.TagSet.Equals(test.tagSet) {
				t.Errorf("Expected tagset to be unchanged by transform, changed %+v into %+v", test.tagSet, result.TagSet)
				continue
			}
			if len(result.Values) != len(transform.expected) {
				t.Errorf("Expected result to have length %d but has length %d", len(transform.expected), len(result.Values))
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
	var testTimerange, err = api.NewTimerange(758400, 758400+30*5, 30)
	if err != nil {
		t.Fatalf("invalid timerange used for testcase")
		return
	}
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
		Timerange: testTimerange,
		Name:      "test",
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

func TestApplyTransformNaN(t *testing.T) {
	var testTimerange, err = api.NewTimerange(758400, 758400+30*5, 30)
	if err != nil {
		t.Fatalf("invalid timerange used for testcase")
		return
	}
	nan := math.NaN()
	list := api.SeriesList{
		Series: []api.Timeseries{
			{
				Values: []float64{0, 1, nan, 3, 4, 5},
				TagSet: api.TagSet{
					"series": "A",
				},
			},
			{
				Values: []float64{2, nan, nan, nan, 3, 3},
				TagSet: api.TagSet{
					"series": "B",
				},
			},
			{
				Values: []float64{0, 1, 2, nan, 2, 1},
				TagSet: api.TagSet{
					"series": "C",
				},
			},
		},
		Timerange: testTimerange,
		Name:      "test",
	}
	tests := []struct {
		transform  transform
		parameters []value
		expected   map[string][]float64
	}{
		{
			transform:  transformDerivative,
			parameters: []value{},
			expected: map[string][]float64{
				"A": {0, 1.0 / 30, nan, nan, 1.0 / 30, 1.0 / 30},
				"B": {0, nan, nan, nan, nan, 0.0},
				"C": {0, 1.0 / 30, 1.0 / 30, nan, nan, -1.0 / 30},
			},
		},
		{
			transform:  transformIntegral,
			parameters: []value{},
			expected: map[string][]float64{
				"A": {0, 1 * 30, 1 * 30, 4 * 30, 8 * 30, 13 * 30},
				"B": {2 * 30, 2 * 30, 2 * 30, 2 * 30, 5 * 30, 8 * 30},
				"C": {0, 1 * 30, 3 * 30, 3 * 30, 5 * 30, 6 * 30},
			},
		},
		{
			transform:  transformRate,
			parameters: []value{},
			expected: map[string][]float64{
				"A": {0, 1 / 30.0, nan, nan, 1 / 30.0, 1 / 30.0},
				"B": {0, nan, nan, nan, nan, 0},
				"C": {0, 1 / 30.0, 1 / 30.0, nan, nan, 0},
			},
		},
		{
			transform:  transformCumulative,
			parameters: []value{},
			expected: map[string][]float64{
				"A": {0, 1, 1, 4, 8, 13},
				"B": {2, 2, 2, 2, 5, 8},
				"C": {0, 1, 3, 3, 5, 6},
			},
		},
		{
			transform:  transformMovingAverage,
			parameters: []value{scalarValue(100)},
			expected: map[string][]float64{
				"A": {0, 0.5, 0.5, 2, 3.5, 4},
				"B": {2, 2, 2, nan, 3, 3},
				"C": {0, 0.5, 1, 1.5, 2, 1.5},
			},
		},
		{
			transform:  transformDefault,
			parameters: []value{scalarValue(17)},
			expected: map[string][]float64{
				"A": {0, 1, 17, 3, 4, 5},
				"B": {2, 17, 17, 17, 3, 3},
				"C": {0, 1, 2, 17, 2, 1},
			},
		},
		{
			transform:  transformNaNKeepLast,
			parameters: []value{},
			expected: map[string][]float64{
				"A": {0, 1, 1, 3, 4, 5},
				"B": {2, 2, 2, 2, 3, 3},
				"C": {0, 1, 2, 2, 2, 1},
			},
		},
	}
	for _, test := range tests {
		result, err := ApplyTransform(list, test.transform, test.parameters)
		if err != nil {
			t.Fatalf(fmt.Sprintf("error applying transformation %s", err))
			return
		}
		for _, series := range result.Series {
			values := series.Values
			expected := test.expected[series.TagSet["series"]]
			if len(values) != len(expected) {
				t.Errorf("values != expected; %+v != %+v", values, expected)
				continue
			}
			for i := range values {
				v := values[i]
				e := expected[i]
				if (math.IsNaN(e) != math.IsNaN(v)) || (!math.IsNaN(e) && math.Abs(v-e) > 1e-7) {
					t.Errorf("(actual) %+v != %+v (expected)", values, expected)
					break
				}
			}
		}
	}
}

func TestApplyTransformFailure(t *testing.T) {
	var testTimerange, err = api.NewTimerange(758400, 758400+30*5, 30)
	if err != nil {
		t.Fatalf("invalid timerange used for testcase")
		return
	}
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
		Timerange: testTimerange,
		Name:      "test",
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
