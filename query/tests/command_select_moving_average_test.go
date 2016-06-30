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

// Integration test for the query execution.
package tests

import (
	"math"
	"testing"

	"github.com/square/metrics/api"
	"github.com/square/metrics/query/command"
	"github.com/square/metrics/query/parser"
	"github.com/square/metrics/testing_support/assert"
	"github.com/square/metrics/testing_support/mocks"

	"golang.org/x/net/context"
)

func TestSelectMovingAverage(t *testing.T) {
	a := assert.New(t)
	testTimerange, err := api.NewSnappedTimerange(0, 70, 10) // inclusive: 8 slots
	if err != nil {
		t.Fatalf("Error creating timerange for test: %s", err.Error())
	}

	n := math.NaN()

	comboAPI := mocks.NewComboAPI(
		// timerange
		testTimerange,
		// series_a
		api.Timeseries{Values: []float64{0, 2, 3, 4, 6, 7, 8, 9}, TagSet: api.TagSet{"metric": "series_a", "line": "a"}},
		api.Timeseries{Values: []float64{1, 1, 1, 0, 2, 4, 3, 1}, TagSet: api.TagSet{"metric": "series_a", "line": "b"}},
		api.Timeseries{Values: []float64{5, 5, 5, 6, 4, 2, 6, 4}, TagSet: api.TagSet{"metric": "series_a", "line": "c"}},

		api.Timeseries{Values: []float64{n, n, 5, 6, 4, n, n, 1}, TagSet: api.TagSet{"metric": "series_a", "line": "na"}},
		api.Timeseries{Values: []float64{5, n, 5, 6, n, n, n, 4}, TagSet: api.TagSet{"metric": "series_a", "line": "nb"}},
		api.Timeseries{Values: []float64{n, n, n, n, n, n, n, n}, TagSet: api.TagSet{"metric": "series_a", "line": "nc"}},
	)

	type test struct {
		query    string
		expected map[string][]float64
	}

	nnnnn := n

	tests := []test{
		{
			query: "select series_a | transform.moving_average(30ms) from 40 to 70 resolution 10ms",
			expected: map[string][]float64{
				"a":  {4.333, 5.666, 7.000, 8.000},
				"b":  {1.000, 2.000, 3.000, 2.666},
				"c":  {5.000, 4.000, 4.000, 4.000},
				"na": {5.000, 5.000, 4.000, 1.000},
				"nb": {5.500, 6.000, nnnnn, 4.000},
				"nc": {nnnnn, nnnnn, nnnnn, nnnnn},
			},
		},
		{
			query: "select series_a | transform.moving_average(40ms) from 50 to 70 resolution 10ms",
			expected: map[string][]float64{
				"a":  {5.000, 6.250, 7.500},
				"b":  {1.750, 2.250, 2.500},
				"c":  {4.250, 4.500, 4.000},
				"na": {5.000, 5.000, 2.500},
				"nb": {5.500, 6.000, 4.000},
				"nc": {nnnnn, nnnnn, nnnnn},
			},
		},
		// exponential
		{
			query: "select series_a | transform.exponential_moving_average(30ms) from 40 to 70 resolution 10ms",
			expected: map[string][]float64{
				"a":  {4.126, 4.991, 5.819, 6.637},
				"b":  {1.070, 1.952, 2.240, 1.921},
				"c":  {4.929, 4.047, 4.584, 4.433},
				"na": {4.914, 4.914, 4.914, 3.144},
				"nb": {5.557, 5.557, 5.557, 4.647},
				"nc": {nnnnn, nnnnn, nnnnn, nnnnn},
			},
		},
		{
			query: "select series_a | transform.exponential_moving_average(40ms) from 50 to 70 resolution 10ms",
			expected: map[string][]float64{
				"a":  {4.847, 5.623, 6.387},
				"b":  {1.860, 2.140, 1.882},
				"c":  {4.139, 4.597, 4.462},
				"na": {4.937, 4.937, 3.371},
				"nb": {5.543, 5.543, 4.739},
				"nc": {nnnnn, nnnnn, nnnnn},
			},
		},
	}

	for _, test := range tests {
		a := a.Contextf("Query %s", test.query)
		context := command.ExecutionContext{
			TimeseriesStorageAPI: comboAPI,
			MetricMetadataAPI:    comboAPI,
			FetchLimit:           100,
			Ctx:                  context.Background(),
		}
		commandObject, err := parser.Parse(test.query)
		if err != nil {
			t.Fatalf("Error parsing command %s: %s", test.query, err.Error())
		}
		result, err := commandObject.Execute(context)
		if err != nil {
			t.Fatalf("Error evaluating %s: %s", test.query, err.Error())
		}
		value := result.Body.([]command.QueryResult)[0]
		a.Eq(value.Type, "series") // Confirm that it's a scalar set.
		a.Contextf("number of results").Eq(len(value.Series), len(test.expected))
		for i := range value.Series {
			series := value.Series[i]
			if correct, ok := test.expected[series.TagSet["line"]]; ok {
				a.Contextf("value for line %s", series.TagSet["line"]).EqFloatArray(series.Values, correct, 1e-3)
			} else {
				a.Errorf("Unexpected tag set in result: %+v", series)
			}
		}
	}

}
