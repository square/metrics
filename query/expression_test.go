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
	"testing"

	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
	"github.com/square/metrics/function/registry"
	"github.com/square/metrics/testing_support/assert"
)

type FakeBackend struct {
	api.TimeseriesStorageAPI
}

type LiteralExpression struct {
	Values []float64
}

func (le LiteralExpression) QueryString() string {
	return "<literal expression>"
}
func (le LiteralExpression) Name() string {
	return "<literal expression>"
}

func (expr *LiteralExpression) Evaluate(context function.EvaluationContext) (function.Value, error) {
	return api.SeriesList{
		Series:    []api.Timeseries{api.Timeseries{Values: expr.Values, TagSet: api.NewTagSet()}},
		Timerange: api.Timerange{},
	}, nil
}

type LiteralSeriesExpression struct {
	list api.SeriesList
}

func (lse LiteralSeriesExpression) QueryString() string {
	return "<literal series expression>"
}
func (lse LiteralSeriesExpression) Name() string {
	return "<literal series expression>"
}
func (expr *LiteralSeriesExpression) Evaluate(context function.EvaluationContext) (function.Value, error) {
	return expr.list, nil
}

func Test_ScalarExpression(t *testing.T) {
	timerangeA, err := api.NewTimerange(0, 10, 2)
	if err != nil {
		t.Fatalf("invalid timerange used for testcase")
		return
	}
	for _, test := range []struct {
		expr           scalarExpression
		timerange      api.Timerange
		expectedSeries []api.Timeseries
	}{
		{
			scalarExpression{5},
			timerangeA,
			[]api.Timeseries{
				api.Timeseries{
					Values: []float64{5.0, 5.0, 5.0, 5.0, 5.0, 5.0},
					TagSet: api.NewTagSet(),
				},
			},
		},
	} {
		a := assert.New(t).Contextf("%+v", test)
		result, err := function.EvaluateToSeriesList(test.expr, function.EvaluationContext{
			TimeseriesStorageAPI: FakeBackend{},
			Timerange:            test.timerange,
			SampleMethod:         api.SampleMean,
			FetchLimit:           function.NewFetchCounter(1000),
			SlotLimit:            28800,
			Registry:             registry.Default(),
		})

		if err != nil {
			t.Fatalf("failed to convert number into serieslist")
		}

		a.EqInt(len(result.Series), len(test.expectedSeries))

		for i := 0; i < len(result.Series); i++ {
			a.Eq(result.Series[i].Values, test.expectedSeries[i].Values)
		}
	}
}

func Test_evaluateBinaryOperation(t *testing.T) {
	emptyContext := function.EvaluationContext{
		TimeseriesStorageAPI: FakeBackend{},
		MetricMetadataAPI:    nil,
		Timerange:            api.Timerange{},
		SampleMethod:         api.SampleMean,
		Predicate:            nil,
		FetchLimit:           function.NewFetchCounter(1000),
		SlotLimit:            28800,
		Cancellable:          api.NewCancellable(),
	}
	for _, test := range []struct {
		context              function.EvaluationContext
		functionName         string
		left                 api.SeriesList
		right                api.SeriesList
		evalFunction         func(float64, float64) float64
		expectSuccess        bool
		expectedResultValues [][]float64
	}{
		{
			emptyContext,
			"add",
			api.SeriesList{
				[]api.Timeseries{
					{
						Values: []float64{1, 2, 3},
						TagSet: api.TagSet{},
					},
				},
				api.Timerange{},
			},
			api.SeriesList{
				[]api.Timeseries{
					{
						Values: []float64{4, 5, 1},
						TagSet: api.TagSet{},
					},
				},
				api.Timerange{},
			},
			func(left, right float64) float64 { return left + right },
			true,
			[][]float64{{5, 7, 4}},
		},
		{
			emptyContext,
			"subtract",
			api.SeriesList{
				[]api.Timeseries{
					{
						Values: []float64{1, 2, 3},
					},
				},
				api.Timerange{},
			},
			api.SeriesList{
				[]api.Timeseries{
					{
						Values: []float64{4, 5, 1},
					},
				},
				api.Timerange{},
			},
			func(left, right float64) float64 { return left - right },
			true,
			[][]float64{{-3, -3, 2}},
		},
		{
			emptyContext,
			"add",
			api.SeriesList{
				[]api.Timeseries{
					api.Timeseries{
						Values: []float64{1, 2, 3},
						TagSet: api.TagSet{
							"env":  "production",
							"host": "#1",
						},
					},
					api.Timeseries{
						Values: []float64{7, 7, 7},
						TagSet: api.TagSet{
							"env":  "staging",
							"host": "#2",
						},
					},
					api.Timeseries{
						Values: []float64{1, 0, 2},
						TagSet: api.TagSet{
							"env":  "staging",
							"host": "#3",
						},
					},
				},
				api.Timerange{},
			},
			api.SeriesList{
				[]api.Timeseries{
					api.Timeseries{
						Values: []float64{5, 5, 5},
						TagSet: api.TagSet{
							"env": "staging",
						},
					},
					api.Timeseries{
						Values: []float64{10, 100, 1000},
						TagSet: api.TagSet{
							"env": "production",
						},
					},
				},
				api.Timerange{},
			},
			func(left, right float64) float64 { return left + right },
			true,
			[][]float64{{11, 102, 1003}, {12, 12, 12}, {6, 5, 7}},
		},
		{
			emptyContext,
			"add",
			api.SeriesList{
				[]api.Timeseries{
					api.Timeseries{
						Values: []float64{1, 2, 3},
						TagSet: api.TagSet{
							"env":  "production",
							"host": "#1",
						},
					},
					api.Timeseries{
						Values: []float64{4, 5, 6},
						TagSet: api.TagSet{
							"env":  "staging",
							"host": "#2",
						},
					},
					api.Timeseries{
						Values: []float64{7, 8, 9},
						TagSet: api.TagSet{
							"env":  "staging",
							"host": "#3",
						},
					},
				},
				api.Timerange{},
			},
			api.SeriesList{
				[]api.Timeseries{
					api.Timeseries{
						Values: []float64{2, 2, 2},
						TagSet: api.TagSet{
							"env": "staging",
						},
					},
					api.Timeseries{
						Values: []float64{3, 3, 3},
						TagSet: api.TagSet{
							"env": "staging",
						},
					},
				},
				api.Timerange{},
			},
			func(left, right float64) float64 { return left * right },
			true,
			[][]float64{{8, 10, 12}, {14, 16, 18}, {12, 15, 18}, {21, 24, 27}},
		},
		{
			emptyContext,
			"add",
			api.SeriesList{
				[]api.Timeseries{
					api.Timeseries{
						Values: []float64{103, 103, 103},
						TagSet: api.TagSet{
							"env":  "production",
							"host": "#1",
						},
					},
					api.Timeseries{
						Values: []float64{203, 203, 203},
						TagSet: api.TagSet{
							"env":  "staging",
							"host": "#2",
						},
					},
					api.Timeseries{
						Values: []float64{303, 303, 303},
						TagSet: api.TagSet{
							"env":  "staging",
							"host": "#3",
						},
					},
				},
				api.Timerange{},
			},
			api.SeriesList{
				[]api.Timeseries{
					api.Timeseries{
						Values: []float64{1, 2, 3},
						TagSet: api.TagSet{
							"env": "staging",
						},
					},
					api.Timeseries{
						Values: []float64{3, 0, 3},
						TagSet: api.TagSet{
							"env": "production",
						},
					},
				},
				api.Timerange{},
			},
			func(left, right float64) float64 { return left - right },
			true,
			[][]float64{{100, 103, 100}, {202, 201, 200}, {302, 301, 300}},
		},
	} {
		a := assert.New(t).Contextf("%+v", test)

		metricFun := registry.NewOperator(test.functionName, test.evalFunction)

		value, err := metricFun.Evaluate(test.context, []function.Expression{&LiteralSeriesExpression{test.left}, &LiteralSeriesExpression{test.right}}, []string{}, false)
		if err != nil {
			a.EqBool(err == nil, test.expectSuccess)
			continue
		}

		result, err := value.ToSeriesList(test.context.Timerange, "-test-")
		if err != nil {
			a.EqBool(err == nil, test.expectSuccess)
			continue
		}

		// Our expected list should be the same length as the actual one:
		a.EqInt(len(result.Series), len(test.expectedResultValues))

		// The "expected" results are only true up to permutation (since guessing the order they'll come out of `join()` is hard)
		// Provided that they're all unique then we just need to check that every member that's expected can be found
		// This is a bit more annoying:

		equal := func(left, right []float64) bool {
			if len(left) != len(right) {
				return false
			}
			for i := range left {
				if left[i] != right[i] {
					return false
				}
			}
			return true
		}

		for _, expectedMember := range test.expectedResultValues {
			found := false
			// check that expectedMember is inside our result list
			// look for it inside result.Series
			for _, resultMember := range result.Series {
				if equal(resultMember.Values, expectedMember) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("got %+v for test %+v", result, test)
			}
		}

	}
}

var _ api.TimeseriesStorageAPI = (*FakeBackend)(nil)
var _ function.Expression = (*LiteralExpression)(nil)
var _ function.Expression = (*LiteralSeriesExpression)(nil)
