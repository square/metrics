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
	"github.com/square/metrics/assert"
)

type FakeBackend struct{}

func (b FakeBackend) Api() api.API {
	return nil
}

func (b FakeBackend) FetchSeries(metric api.TaggedMetric, tagConstraints api.Predicate, sampleMethod api.SampleMethod, timerange api.Timerange) (*api.SeriesList, error) {
	return &api.SeriesList{}, nil
}

type LiteralExpression struct {
	Values []float64
}

func (expr *LiteralExpression) Evaluate(context EvaluationContext) (value, error) {
	return seriesListValue(api.SeriesList{
		Series:    []api.Timeseries{api.Timeseries{expr.Values, api.NewTagSet()}},
		Timerange: api.Timerange{},
	}), nil
}

type LiteralSeriesExpression struct {
	Values []api.Timeseries
}

func (expr *LiteralSeriesExpression) Evaluate(context EvaluationContext) (value, error) {
	result := api.SeriesList{
		Series:    make([]api.Timeseries, len(expr.Values)),
		Timerange: api.Timerange{},
	}
	for i, values := range expr.Values {
		result.Series[i] = values
	}
	return seriesListValue(result), nil
}

func Test_ScalarExpression(t *testing.T) {
	timerangeA, _ := api.NewTimerange(0, 10, 2)
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
					[]float64{5.0, 5.0, 5.0, 5.0, 5.0, 5.0},
					api.NewTagSet(),
				},
			},
		},
	} {
		a := assert.New(t).Contextf("%+v", test)

		result, err := evaluateToSeriesList(test.expr, EvaluationContext{
			Backend:      FakeBackend{},
			Timerange:    test.timerange,
			SampleMethod: api.SampleMean,
		})

		if err != nil {
			t.Fatalf("failed to convert number into serieslist")
		}

		a.EqInt(len(result.Series), len(test.expectedSeries))

		for i := 0; i < len(result.Series); i += 1 {
			a.Eq(result.Series[i].Values, test.expectedSeries[i].Values)
		}
	}
}

func Test_evaluateBinaryOperation(t *testing.T) {
	emptyContext := EvaluationContext{FakeBackend{}, api.Timerange{}, api.SampleMean, nil}
	for _, test := range []struct {
		context              EvaluationContext
		functionName         string
		left                 value
		right                value
		evalFunction         func(float64, float64) float64
		expectSuccess        bool
		expectedResultValues [][]float64
	}{
		{
			emptyContext,
			"add",
			seriesListValue(api.SeriesList{
				[]api.Timeseries{
					{
						Values: []float64{1, 2, 3},
						TagSet: api.TagSet{},
					},
				},
				api.Timerange{},
				"",
			}),
			seriesListValue(api.SeriesList{
				[]api.Timeseries{
					{
						Values: []float64{4, 5, 1},
						TagSet: api.TagSet{},
					},
				},
				api.Timerange{},
				"",
			}),
			func(left, right float64) float64 { return left + right },
			true,
			[][]float64{{5, 7, 4}},
		},
		{
			emptyContext,
			"subtract",
			seriesListValue(api.SeriesList{
				[]api.Timeseries{
					{
						Values: []float64{1, 2, 3},
					},
				},
				api.Timerange{},
				"",
			}),
			seriesListValue(api.SeriesList{
				[]api.Timeseries{
					{
						Values: []float64{4, 5, 1},
					},
				},
				api.Timerange{},
				"",
			}),
			func(left, right float64) float64 { return left - right },
			true,
			[][]float64{{-3, -3, 2}},
		},
		{
			emptyContext,
			"add",
			seriesListValue(api.SeriesList{
				[]api.Timeseries{
					api.Timeseries{
						[]float64{1, 2, 3},
						api.TagSet{
							"env":  "production",
							"host": "#1",
						},
					},
					api.Timeseries{
						[]float64{7, 7, 7},
						api.TagSet{
							"env":  "staging",
							"host": "#2",
						},
					},
					api.Timeseries{
						[]float64{1, 0, 2},
						api.TagSet{
							"env":  "staging",
							"host": "#3",
						},
					},
				},
				api.Timerange{},
				"",
			}),
			seriesListValue(api.SeriesList{
				[]api.Timeseries{
					api.Timeseries{
						[]float64{5, 5, 5},
						api.TagSet{
							"env": "staging",
						},
					},
					api.Timeseries{
						[]float64{10, 100, 1000},
						api.TagSet{
							"env": "production",
						},
					},
				},
				api.Timerange{},
				"",
			}),
			func(left, right float64) float64 { return left + right },
			true,
			[][]float64{{11, 102, 1003}, {12, 12, 12}, {6, 5, 7}},
		},
		{
			emptyContext,
			"add",
			seriesListValue(api.SeriesList{
				[]api.Timeseries{
					api.Timeseries{
						[]float64{1, 2, 3},
						api.TagSet{
							"env":  "production",
							"host": "#1",
						},
					},
					api.Timeseries{
						[]float64{4, 5, 6},
						api.TagSet{
							"env":  "staging",
							"host": "#2",
						},
					},
					api.Timeseries{
						[]float64{7, 8, 9},
						api.TagSet{
							"env":  "staging",
							"host": "#3",
						},
					},
				},
				api.Timerange{},
				"",
			}),
			seriesListValue(api.SeriesList{
				[]api.Timeseries{
					api.Timeseries{
						[]float64{2, 2, 2},
						api.TagSet{
							"env": "staging",
						},
					},
					api.Timeseries{
						[]float64{3, 3, 3},
						api.TagSet{
							"env": "staging",
						},
					},
				},
				api.Timerange{},
				"",
			}),
			func(left, right float64) float64 { return left * right },
			true,
			[][]float64{{8, 10, 12}, {14, 16, 18}, {12, 15, 18}, {21, 24, 27}},
		},
		{
			emptyContext,
			"add",
			seriesListValue(api.SeriesList{
				[]api.Timeseries{
					api.Timeseries{
						[]float64{103, 103, 103},
						api.TagSet{
							"env":  "production",
							"host": "#1",
						},
					},
					api.Timeseries{
						[]float64{203, 203, 203},
						api.TagSet{
							"env":  "staging",
							"host": "#2",
						},
					},
					api.Timeseries{
						[]float64{303, 303, 303},
						api.TagSet{
							"env":  "staging",
							"host": "#3",
						},
					},
				},
				api.Timerange{},
				"",
			}),
			seriesListValue(api.SeriesList{
				[]api.Timeseries{
					api.Timeseries{
						[]float64{1, 2, 3},
						api.TagSet{
							"env": "staging",
						},
					},
					api.Timeseries{
						[]float64{3, 0, 3},
						api.TagSet{
							"env": "production",
						},
					},
				},
				api.Timerange{},
				"",
			}),
			func(left, right float64) float64 { return left - right },
			true,
			[][]float64{{100, 103, 100}, {202, 201, 200}, {302, 301, 300}},
		},
	} {
		a := assert.New(t).Contextf("%+v", test)

		value, err := evaluateBinaryOperation(
			test.context,
			test.functionName,
			test.left,
			test.right,
			test.evalFunction,
		)
		if err != nil {
			a.EqBool(err == nil, test.expectSuccess)
			continue
		}

		result, err := value.toSeriesList(test.context.Timerange)
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

var _ api.Backend = (*FakeBackend)(nil)
var _ Expression = (*LiteralExpression)(nil)
var _ Expression = (*LiteralSeriesExpression)(nil)
