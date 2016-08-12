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

package transform

import (
	"fmt"
	"math"
	"testing"

	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
	"github.com/square/metrics/testing_support/assert"

	"golang.org/x/net/context"
)

type literal struct {
	value function.Value
}

func (lit literal) ExpressionDescription(mode function.DescriptionMode) string {
	if mode == function.StringMemoization {
		return fmt.Sprintf("%#v", lit)
	}
	return "<literal>"
}
func (lit literal) Evaluate(context function.EvaluationContext) (function.Value, error) {
	return lit.value, nil
}

func TestTransformTimeseries(t *testing.T) {
	timerange, err := api.NewSnappedTimerange(0, 4*30000, 30000)
	if err != nil {
		t.Fatalf("Error creating test timerange: %s", err.Error())
	}
	testCases := []struct {
		series     api.Timeseries
		values     []float64
		tagSet     api.TagSet
		parameters []function.Value
		tests      []struct {
			fun      function.Function
			expected []float64
		}
	}{
		{
			values: []float64{0, 1, 2, 3, 4, 5},
			tagSet: api.TagSet{
				"dc":   "A",
				"host": "B",
				"env":  "C",
			},
			parameters: []function.Value{function.ScalarValue(100)},
			tests: []struct {
				fun      function.Function
				expected []float64
			}{
				{
					fun:      Derivative,
					expected: []float64{1.0 / 30.0, 1.0 / 30.0, 1.0 / 30.0, 1.0 / 30.0, 1.0 / 30.0},
				},
				{
					fun:      Integral,
					expected: []float64{0.0, 1.0 * 30.0, 3.0 * 30.0, 6.0 * 30.0, 10.0 * 30.0, 15.0 * 30.0},
				},
				{
					fun:      MapMaker("negative", func(x float64) float64 { return -x }),
					expected: []float64{0, -1, -2, -3, -4, -5},
				},
				{
					fun:      NaNKeepLast,
					expected: []float64{0, 1, 2, 3, 4, 5},
				},
				{
					fun:      Rate,
					expected: []float64{1.0 / 30.0, 1.0 / 30.0, 1.0 / 30.0, 1.0 / 30.0, 1.0 / 30.0},
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
			ctx := function.EvaluationContextBuilder{Timerange: timerange, Ctx: context.Background()}.Build()
			seriesList := api.SeriesList{
				Series: []api.Timeseries{series},
			}
			resultValue, err := transform.fun.Run(ctx, []function.Expression{&literal{function.SeriesListValue(seriesList)}}, function.Groups{})
			if err != nil {
				t.Error(err)
				continue
			}
			resultList, convErr := resultValue.ToSeriesList(ctx.Timerange())
			if convErr != nil {
				t.Errorf("Conversion to series list failed: %s", convErr.WithContext("???").Error())
				continue
			}
			result := resultList.Series[0]
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
	timerange, err := api.NewSnappedTimerange(0, 5*30000, 30000)
	if err != nil {
		t.Fatalf("Error creating timerange: %s", err.Error())
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
	}
	listExpression := literal{function.SeriesListValue(list)}
	testCases := []struct {
		transform function.Function
		// Each function is given just the list as an argument.
		expected map[string][]float64
	}{
		{
			transform: Cumulative,
			expected: map[string][]float64{
				"A": {0, 1, 3, 6, 10, 15},
				"B": {0, 2, 3, 4, 7, 10},
				"C": {0, 1, 3, 6, 8, 9},
			},
		},
		{
			transform: Derivative,
			expected: map[string][]float64{
				"A": {1.0 / 30, 1.0 / 30, 1.0 / 30, 1.0 / 30, 1.0 / 30},
				"B": {0, -1.0 / 30, 0, 2.0 / 30, 0},
				"C": {1.0 / 30, 1.0 / 30, 1.0 / 30, -1.0 / 30, -1.0 / 30},
			},
		},
		{
			transform: Integral,
			expected: map[string][]float64{
				"A": {0, 1 * 30, 3 * 30, 6 * 30, 10 * 30, 15 * 30},
				"B": {0, 2 * 30, 3 * 30, 4 * 30, 7 * 30, 10 * 30},
				"C": {0, 1 * 30, 3 * 30, 6 * 30, 8 * 30, 9 * 30},
			},
		},
		{
			transform: Rate,
			expected: map[string][]float64{
				"A": {1.0 / 30, 1.0 / 30, 1.0 / 30, 1.0 / 30, 1.0 / 30},
				"B": {0, 1.0 / 30, 0, 2.0 / 30, 0},
				"C": {1.0 / 30, 1.0 / 30, 1.0 / 30, 0.0, 0.0},
			},
		},
	}
	for _, test := range testCases {
		ctx := function.EvaluationContextBuilder{Timerange: timerange, Ctx: context.Background()}.Build()
		resultValue, err := test.transform.Run(ctx, []function.Expression{listExpression}, function.Groups{})
		if err != nil {
			t.Error(err)
			continue
		}
		result, convErr := resultValue.ToSeriesList(ctx.Timerange())
		if convErr != nil {
			t.Errorf("Error converting to series list: %s", convErr.WithContext("test case").Error())
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
	timerange, err := api.NewSnappedTimerange(0, 5*30000, 30000)
	if err != nil {
		t.Fatalf("Error creating timerange for test case: %s", err.Error())
	}
	list := api.SeriesList{
		Series: []api.Timeseries{
			{
				Values: []float64{1, 2, 3, 2, 1, 2},
				TagSet: api.TagSet{
					"series": "C",
				},
			},
		},
	}
	listExpression := literal{function.SeriesListValue(list)}

	testCases := []struct {
		transform  function.Function
		parameters []function.Expression
		expected   []string
	}{
		{
			transform:  Rate,
			parameters: []function.Expression{listExpression},
			expected: []string{
				"Rate(map[series:C]): The underlying counter reset between 2.000000, 1.000000\n",
			},
		},
	}

	for _, test := range testCases {
		ctx := function.EvaluationContextBuilder{EvaluationNotes: &function.EvaluationNotes{}, Timerange: timerange, Ctx: context.Background()}.Build()
		_, err := test.transform.Run(ctx, test.parameters, function.Groups{})
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
	}
	listExpression := literal{function.SeriesListValue(list)}
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
			bounder    function.Function
			parameters []function.Expression
			expected   map[string][]float64
			name       string
		}{
			{
				bounder:    Bound,
				parameters: []function.Expression{listExpression, literal{function.ScalarValue(test.lower)}, literal{function.ScalarValue(test.upper)}},
				expected:   test.expectBound,
				name:       "bound",
			},
			{
				bounder:    LowerBound,
				parameters: []function.Expression{listExpression, literal{function.ScalarValue(test.lower)}},
				expected:   test.expectLower,
				name:       "lower",
			},
			{
				bounder:    UpperBound,
				parameters: []function.Expression{listExpression, literal{function.ScalarValue(test.upper)}},
				expected:   test.expectUpper,
				name:       "upper",
			},
		}

		for _, bounderDetails := range bounders {
			ctx := function.EvaluationContextBuilder{Ctx: context.Background()}.Build()
			boundedValue, err := bounderDetails.bounder.Run(ctx, bounderDetails.parameters, function.Groups{})
			if err != nil {
				t.Errorf(err.Error())
				continue
			}
			bounded, convErr := boundedValue.ToSeriesList(ctx.Timerange())
			if convErr != nil {
				t.Errorf("Error converting to series list: %s", convErr.WithContext("test case"))
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
				a.EqFloatArray(series.Values, bounderDetails.expected[series.TagSet["name"]], 3e-7)
			}
		}
	}
	ctx := function.EvaluationContextBuilder{Ctx: context.Background()}.Build()
	if _, err := Bound.Run(ctx, []function.Expression{listExpression, literal{function.ScalarValue(18)}, literal{function.ScalarValue(17)}}, function.Groups{}); err == nil {
		t.Fatalf("Expected error on invalid bounds")
	}
	if _, err := Bound.Run(ctx, []function.Expression{listExpression, literal{function.ScalarValue(-17)}, literal{function.ScalarValue(-18)}}, function.Groups{}); err == nil {
		t.Fatalf("Expected error on invalid bounds")
	}
}

func TestApplyTransformNaN(t *testing.T) {
	timerange, err := api.NewSnappedTimerange(0, 5*30000, 30000)
	if err != nil {
		t.Fatalf("Error constructing timerange for testcase; %s", err.Error())
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
	}
	listExpression := literal{function.SeriesListValue(list)}
	tests := []struct {
		transform  function.Function
		parameters []function.Expression
		expected   map[string][]float64
	}{
		{
			transform:  Derivative,
			parameters: []function.Expression{listExpression},
			expected: map[string][]float64{
				"A": {1.0 / 30, nan, nan, 1.0 / 30, 1.0 / 30},
				"B": {nan, nan, nan, nan, 0.0},
				"C": {1.0 / 30, 1.0 / 30, nan, nan, -1.0 / 30},
			},
		},
		{
			transform:  Integral,
			parameters: []function.Expression{listExpression},
			expected: map[string][]float64{
				"A": {0, 1 * 30, 1 * 30, 4 * 30, 8 * 30, 13 * 30},
				"B": {0, 0, 0, 0, 3 * 30, 6 * 30},
				"C": {0, 1 * 30, 3 * 30, 3 * 30, 5 * 30, 6 * 30},
			},
		},
		{
			transform:  Rate,
			parameters: []function.Expression{listExpression},
			expected: map[string][]float64{
				"A": {1 / 30.0, nan, nan, 1 / 30.0, 1 / 30.0},
				"B": {nan, nan, nan, nan, 0},
				"C": {1 / 30.0, 1 / 30.0, nan, nan, 0.0},
			},
		},
		{
			transform:  Cumulative,
			parameters: []function.Expression{listExpression},
			expected: map[string][]float64{
				"A": {0, 1, 1, 4, 8, 13},
				"B": {0, 0, 0, 0, 3, 6},
				"C": {0, 1, 3, 3, 5, 6},
			},
		},
		{
			transform:  NaNFill,
			parameters: []function.Expression{listExpression, literal{function.ScalarValue(17)}},
			expected: map[string][]float64{
				"A": {0, 1, 17, 3, 4, 5},
				"B": {2, 17, 17, 17, 3, 3},
				"C": {0, 1, 2, 17, 2, 1},
			},
		},
		{
			transform:  NaNKeepLast,
			parameters: []function.Expression{listExpression},
			expected: map[string][]float64{
				"A": {0, 1, 1, 3, 4, 5},
				"B": {2, 2, 2, 2, 3, 3},
				"C": {0, 1, 2, 2, 2, 1},
			},
		},
	}
	for _, test := range tests {
		ctx := function.EvaluationContextBuilder{Timerange: timerange, Ctx: context.Background()}.Build()
		resultValue, err := test.transform.Run(ctx, test.parameters, function.Groups{})
		if err != nil {
			t.Fatalf(fmt.Sprintf("error applying transformation %s", err))
			return
		}
		result, convErr := resultValue.ToSeriesList(ctx.Timerange())
		if convErr != nil {
			t.Fatalf("error converting to series list: %s", convErr.WithContext("test case"))
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
	timerange, err := api.NewSnappedTimerange(0, 5*30000, 30000)
	if err != nil {
		t.Fatalf("Error constructing timerange for testcase: %s", err.Error())
	}

	testCases := []struct {
		values []float64
		tests  []struct {
			expected   []float64
			transforms []function.Function
		}
	}{
		{
			values: []float64{0, 1, 2, 3, 4, 5},
			tests: []struct {
				expected   []float64
				transforms []function.Function
			}{
				{
					expected: []float64{0, 1, 2, 3, 4},
					transforms: []function.Function{
						Derivative,
						Integral,
					},
				},
				{
					expected: []float64{0, 1, 2, 3, 4},
					transforms: []function.Function{
						Rate,
						Integral,
					},
				},
			},
		},
		{
			values: []float64{12, 15, 20, 3, 18, 30},
			tests: []struct {
				expected   []float64
				transforms []function.Function
			}{
				{
					expected: []float64{0, 5, -12, 3, 15},
					transforms: []function.Function{
						Derivative,
						Integral,
					},
				},
				{
					// While this is odd, think about it this way:
					// We saw 5 increments (15 - 20), then we saw thirty total increments
					// (3, 18, 30) over the rest of the time period
					expected: []float64{0, 5, 8, 23, 35},
					transforms: []function.Function{
						Rate,
						Integral,
					},
				},
			},
		},
	}
	epsilon := 1e-10
	for _, test := range testCases {
		series := api.Timeseries{
			Values: test.values,
			TagSet: api.TagSet{},
		}
		for _, transform := range test.tests {
			result := series
			for _, fun := range transform.transforms {
				ctx := function.EvaluationContextBuilder{Timerange: timerange, Ctx: context.Background()}.Build()

				seriesList := api.SeriesList{
					Series: []api.Timeseries{result},
				}
				params := []function.Expression{literal{function.SeriesListValue(seriesList)}}
				aValue, runErr := fun.Run(ctx, params, function.Groups{})
				if runErr != nil {
					t.Error(runErr)
					break
				}
				a, convErr := aValue.ToSeriesList(ctx.Timerange())
				if convErr != nil {
					t.Error(convErr)
					break
				}
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
