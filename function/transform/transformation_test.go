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

package transform

import (
	"fmt"
	"math"
	"testing"

	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
	"github.com/square/metrics/testing_support/assert"
)

func TestTransformTimeseries(t *testing.T) {
	//This is to make sure that the scale of all the data
	//is interpreted as 30 seconds (30000 milliseconds)
	timerange, _ := api.NewTimerange(0, int64(30000*5), int64(30000))

	testCases := []struct {
		series     api.Timeseries
		values     []float64
		tagSet     api.TagSet
		parameters []function.Value
		timerange  api.Timerange
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
			timerange:  timerange,
			parameters: []function.Value{function.ScalarValue(100)},
			tests: []struct {
				fun      transform
				expected []float64
				useParam bool
			}{
				{
					fun:      derivative,
					expected: []float64{1.0 / 30.0, 1.0 / 30.0, 1.0 / 30.0, 1.0 / 30.0, 1.0 / 30.0},
					useParam: false,
				},
				{
					fun:      Integral,
					expected: []float64{0.0, 1.0 * 30.0, 3.0 * 30.0, 6.0 * 30.0, 10.0 * 30.0, 15.0 * 30.0},
					useParam: false,
				},
				{
					fun:      MapMaker(func(x float64) float64 { return -x }),
					expected: []float64{0, -1, -2, -3, -4, -5},
					useParam: false,
				},
				{
					fun:      NaNKeepLast,
					expected: []float64{0, 1, 2, 3, 4, 5},
					useParam: false,
				},
				{
					fun:      rate,
					expected: []float64{1.0 / 30.0, 1.0 / 30.0, 1.0 / 30.0, 1.0 / 30.0, 1.0 / 30.0},
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
				params = []function.Value{}
			}
			ctx := function.EvaluationContext{}
			seriesList := api.SeriesList{
				Series:    []api.Timeseries{series},
				Timerange: timerange,
			}

			a, err := ApplyTransform(ctx, seriesList, transform.fun, params)
			result := a.Series[0]
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
	var testTimerange, err = api.NewTimerange(758400000, 758400000+30000*5, 30000)
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
	}
	testCases := []struct {
		transform transform
		parameter []function.Value
		expected  map[string][]float64
	}{
		{
			transform: Cumulative,
			parameter: []function.Value{},
			expected: map[string][]float64{
				"A": {0, 1, 3, 6, 10, 15},
				"B": {0, 2, 3, 4, 7, 10},
				"C": {0, 1, 3, 6, 8, 9},
			},
		},
		{
			transform: derivative,
			parameter: []function.Value{},
			expected: map[string][]float64{
				"A": {1.0 / 30, 1.0 / 30, 1.0 / 30, 1.0 / 30, 1.0 / 30},
				"B": {0, -1.0 / 30, 0, 2.0 / 30, 0},
				"C": {1.0 / 30, 1.0 / 30, 1.0 / 30, -1.0 / 30, -1.0 / 30},
			},
		},
		{
			transform: Integral,
			parameter: []function.Value{},
			expected: map[string][]float64{
				"A": {0, 1 * 30, 3 * 30, 6 * 30, 10 * 30, 15 * 30},
				"B": {0, 2 * 30, 3 * 30, 4 * 30, 7 * 30, 10 * 30},
				"C": {0, 1 * 30, 3 * 30, 6 * 30, 8 * 30, 9 * 30},
			},
		},
		{
			transform: rate,
			parameter: []function.Value{},
			expected: map[string][]float64{
				"A": {1.0 / 30, 1.0 / 30, 1.0 / 30, 1.0 / 30, 1.0 / 30},
				"B": {0, 1.0 / 30, 0, 2.0 / 30, 0},
				"C": {1.0 / 30, 1.0 / 30, 1.0 / 30, 0.0, 0.0},
			},
		},
	}
	for _, test := range testCases {
		ctx := function.EvaluationContext{}
		result, err := ApplyTransform(ctx, list, test.transform, test.parameter)
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

func TestApplyNotes(t *testing.T) {
	var testTimerange, err = api.NewTimerange(758400000, 758400000+30000*5, 30000)
	if err != nil {
		t.Fatalf("invalid timerange used for testcase")
		return
	}
	// epsilon := 1e-10
	list := api.SeriesList{
		Series: []api.Timeseries{
			{
				Values: []float64{1, 2, 3, 2, 1, 2},
				TagSet: api.TagSet{
					"series": "C",
				},
			},
		},
		Timerange: testTimerange,
	}

	testCases := []struct {
		transform transform
		parameter []function.Value
		expected  []string
	}{
		{
			transform: rate,
			parameter: []function.Value{},
			expected: []string{
				"Rate(map[series:C]): The underlying counter reset between 2.000000, 1.000000\n",
			},
		},
	}

	for _, test := range testCases {
		ctx := function.CreateEvaluationContext(api.Timerange{}, api.UserSpecifiableConfig{}, function.EvaluationContextInternals{EvaluationNotes: new(function.EvaluationNotes)})
		_, err := ApplyTransform(ctx, list, test.transform, test.parameter)
		if err != nil {
			t.Error(err)
			continue
		}
		if len(test.expected) != len(ctx.Notes()) {
			t.Errorf("Expected there to be %d notes but there were %d of them", len(test.expected), len(ctx.Notes()))
		}
		for i, note := range test.expected {
			if i >= len(ctx.Notes()) {
				break
			}
			if ctx.Notes()[i] != note {
				t.Errorf("The context notes didn't include the evaluation message. Expected: %s Actually found: %s\n", note, ctx.Notes()[i])
			}
		}

	}
}

func TestApplyBound(t *testing.T) {
	a := assert.New(t)
	testTimerange, err := api.NewTimerange(758400000, 758400000+30000*5, 30000)
	//{2, nan, nan, nan, 3, 3},
	if err != nil {
		t.Fatal("invalid timerange used for testcase")
		return
	}
	list := api.SeriesList{
		Series: []api.Timeseries{
			{
				Values: []float64{1, 2, 3, 4, 5, 6},
				TagSet: api.TagSet{
					"name": "A",
				},
			},
			{
				Values: []float64{5, 5, 3, -7, math.NaN(), -20},
				TagSet: api.TagSet{
					"name": "B",
				},
			},
			{
				Values: []float64{math.NaN(), 100, 90, 0, 0, 3},
				TagSet: api.TagSet{
					"name": "C",
				},
			},
		},
		Timerange: testTimerange,
	}
	tests := []struct {
		lower       float64
		upper       float64
		expectBound map[string][]float64
		expectLower map[string][]float64
		expectUpper map[string][]float64
	}{
		{
			lower: 2,
			upper: 5,
			expectBound: map[string][]float64{
				"A": {2, 2, 3, 4, 5, 5},
				"B": {5, 5, 3, 2, math.NaN(), 2},
				"C": {math.NaN(), 5, 5, 2, 2, 3},
			},
			expectLower: map[string][]float64{
				"A": {2, 2, 3, 4, 5, 6},
				"B": {5, 5, 3, 2, math.NaN(), 2},
				"C": {math.NaN(), 100, 90, 2, 2, 3},
			},
			expectUpper: map[string][]float64{
				"A": {1, 2, 3, 4, 5, 5},
				"B": {5, 5, 3, -7, math.NaN(), -20},
				"C": {math.NaN(), 5, 5, 0, 0, 3},
			},
		},
		{
			lower: -10,
			upper: 40,
			expectBound: map[string][]float64{
				"A": {1, 2, 3, 4, 5, 6},
				"B": {5, 5, 3, -7, math.NaN(), -10},
				"C": {math.NaN(), 40, 40, 0, 0, 3},
			},
			expectLower: map[string][]float64{
				"A": {1, 2, 3, 4, 5, 6},
				"B": {5, 5, 3, -7, math.NaN(), -10},
				"C": {math.NaN(), 100, 90, 0, 0, 3},
			},
			expectUpper: map[string][]float64{
				"A": {1, 2, 3, 4, 5, 6},
				"B": {5, 5, 3, -7, math.NaN(), -20},
				"C": {math.NaN(), 40, 40, 0, 0, 3},
			},
		},
	}
	for _, test := range tests {
		bounders := []struct {
			bounder  func(ctx function.EvaluationContext, series api.Timeseries, parameters []function.Value, scale float64) ([]float64, error)
			params   []function.Value
			expected map[string][]float64
			name     string
		}{
			{bounder: Bound, params: []function.Value{function.ScalarValue(test.lower), function.ScalarValue(test.upper)}, expected: test.expectBound, name: "bound"},
			{bounder: LowerBound, params: []function.Value{function.ScalarValue(test.lower)}, expected: test.expectLower, name: "lower"},
			{bounder: UpperBound, params: []function.Value{function.ScalarValue(test.upper)}, expected: test.expectUpper, name: "upper"},
		}

		for _, bounder := range bounders {
			ctx := function.EvaluationContext{}
			bounded, err := ApplyTransform(ctx, list, bounder.bounder, bounder.params)
			if err != nil {
				t.Errorf(err.Error())
				continue
			}
			if len(bounded.Series) != len(list.Series) {
				t.Errorf("Expected to get %d results but got %d in %+v", len(list.Series), len(bounded.Series), bounded)
				continue
			}
			// Next, check they're all unique and such
			alreadyUsed := map[string]bool{}
			for _, series := range bounded.Series {
				if alreadyUsed[series.TagSet["name"]] {
					t.Fatalf("Repeating name `%s`", series.TagSet["name"])
				}
				alreadyUsed[series.TagSet["name"]] = true
				// Next, verify that it's what we expect
				a.EqFloatArray(series.Values, bounder.expected[series.TagSet["name"]], 3e-7)
			}
		}
	}
	ctx := function.EvaluationContext{}
	if _, err = ApplyTransform(ctx, list, Bound, []function.Value{function.ScalarValue(18), function.ScalarValue(17)}); err == nil {
		t.Fatalf("Expected error on invalid bounds")
	}
	if _, err = ApplyTransform(ctx, list, Bound, []function.Value{function.ScalarValue(-17), function.ScalarValue(-18)}); err == nil {
		t.Fatalf("Expected error on invalid bounds")
	}
}

func TestApplyTransformNaN(t *testing.T) {
	var testTimerange, err = api.NewTimerange(758400000, 758400000+30000*5, 30000)
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
	}
	tests := []struct {
		transform  transform
		parameters []function.Value
		expected   map[string][]float64
	}{
		{
			transform:  derivative,
			parameters: []function.Value{},
			expected: map[string][]float64{
				"A": {1.0 / 30, nan, nan, 1.0 / 30, 1.0 / 30},
				"B": {nan, nan, nan, nan, 0.0},
				"C": {1.0 / 30, 1.0 / 30, nan, nan, -1.0 / 30},
			},
		},
		{
			transform:  Integral,
			parameters: []function.Value{},
			expected: map[string][]float64{
				"A": {0, 1 * 30, 1 * 30, 4 * 30, 8 * 30, 13 * 30},
				"B": {0, 0, 0, 0, 3 * 30, 6 * 30},
				"C": {0, 1 * 30, 3 * 30, 3 * 30, 5 * 30, 6 * 30},
			},
		},
		{
			transform:  rate,
			parameters: []function.Value{},
			expected: map[string][]float64{
				"A": {1 / 30.0, nan, nan, 1 / 30.0, 1 / 30.0},
				"B": {nan, nan, nan, nan, 0},
				"C": {1 / 30.0, 1 / 30.0, nan, nan, 0.0},
			},
		},
		{
			transform:  Cumulative,
			parameters: []function.Value{},
			expected: map[string][]float64{
				"A": {0, 1, 1, 4, 8, 13},
				"B": {0, 0, 0, 0, 3, 6},
				"C": {0, 1, 3, 3, 5, 6},
			},
		},
		{
			transform:  Default,
			parameters: []function.Value{function.ScalarValue(17)},
			expected: map[string][]float64{
				"A": {0, 1, 17, 3, 4, 5},
				"B": {2, 17, 17, 17, 3, 3},
				"C": {0, 1, 2, 17, 2, 1},
			},
		},
		{
			transform:  NaNKeepLast,
			parameters: []function.Value{},
			expected: map[string][]float64{
				"A": {0, 1, 1, 3, 4, 5},
				"B": {2, 2, 2, 2, 3, 3},
				"C": {0, 1, 2, 2, 2, 1},
			},
		},
	}
	for _, test := range tests {
		ctx := function.EvaluationContext{}
		result, err := ApplyTransform(ctx, list, test.transform, test.parameters)
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
					t.Errorf("(actual) %+v != %+v (expected) for series %s", values, expected, series.TagSet["series"])
					break
				}
			}
		}
	}
}

// Test that the transforms of the following work as expected:
// - transform.derivative | transform.integral
func TestTransformIdentity(t *testing.T) {
	//This is to make sure that the scale of all the data
	//is interpreted as 30 seconds (30000 milliseconds)
	timerange, _ := api.NewTimerange(0, int64(30000*5), int64(30000))

	testCases := []struct {
		values    []float64
		timerange api.Timerange
		tests     []struct {
			expected   []float64
			transforms []transform
		}
	}{
		{
			values:    []float64{0, 1, 2, 3, 4, 5},
			timerange: timerange,
			tests: []struct {
				expected   []float64
				transforms []transform
			}{
				{
					expected: []float64{0, 1, 2, 3, 4},
					transforms: []transform{
						derivative,
						Integral,
					},
				},
				{
					expected: []float64{0, 1, 2, 3, 4},
					transforms: []transform{
						rate,
						Integral,
					},
				},
			},
		},
		{
			values:    []float64{12, 15, 20, 3, 18, 30},
			timerange: timerange,
			tests: []struct {
				expected   []float64
				transforms []transform
			}{
				{
					expected: []float64{0, 5, -12, 3, 15},
					transforms: []transform{
						derivative,
						Integral,
					},
				},
				{
					// While this is odd, think about it this way:
					// We saw 5 increments (15 - 20), then we saw thirty total increments
					// (3, 18, 30) over the rest of the time period
					expected: []float64{0, 5, 8, 23, 35},
					transforms: []transform{
						rate,
						Integral,
					},
				},
			},
		},
	}
	epsilon := 1e-10
	var err error
	for _, test := range testCases {
		series := api.Timeseries{
			Values: test.values,
			TagSet: api.TagSet{},
		}
		for _, transform := range test.tests {
			result := series
			for _, fun := range transform.transforms {
				ctx := function.EvaluationContext{}

				seriesList := api.SeriesList{
					Series:    []api.Timeseries{result},
					Timerange: timerange,
				}
				params := []function.Value{}
				a, err := ApplyTransform(ctx, seriesList, fun, params)
				result = a.Series[0]
				if err != nil {
					t.Error(err)
					break
				}
			}
			if err != nil {
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
