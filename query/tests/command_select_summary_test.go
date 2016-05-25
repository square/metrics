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
)

func TestSelectSummary(t *testing.T) {
	a := assert.New(t)
	testTimerange, err := api.NewSnappedTimerange(0, 4*30000, 30000)
	if err != nil {
		t.Fatalf("Error creating timerange for test: %s", err.Error())
	}

	n := math.NaN()

	comboAPI := mocks.NewComboAPI(
		// timerange
		testTimerange,
		// series_a
		api.Timeseries{Values: []float64{0, 2, 3, 4, 6}, TagSet: api.TagSet{"metric": "series_a", "app": "web", "dc": "west"}},
		api.Timeseries{Values: []float64{1, 1, 1, 0, 2}, TagSet: api.TagSet{"metric": "series_a", "app": "web", "dc": "east"}},
		api.Timeseries{Values: []float64{5, 5, 5, 6, 4}, TagSet: api.TagSet{"metric": "series_a", "app": "fun", "dc": "north"}},
		// series_b
		api.Timeseries{Values: []float64{3, n, 7, n, n}, TagSet: api.TagSet{"metric": "series_b", "dc": "west"}},
		api.Timeseries{Values: []float64{n, n, 5, 2, 2}, TagSet: api.TagSet{"metric": "series_b", "dc": "east"}},
		api.Timeseries{Values: []float64{n, n, n, n, n}, TagSet: api.TagSet{"metric": "series_b", "dc": "miss"}},
	)

	type test struct {
		query    string
		expected map[string]float64
	}

	tests := []test{
		{
			query: "select series_a | summarize.mean from 0 to 120000",
			expected: map[string]float64{
				api.TagSet{"app": "web", "dc": "west"}.Serialize():  3,
				api.TagSet{"app": "web", "dc": "east"}.Serialize():  1,
				api.TagSet{"app": "fun", "dc": "north"}.Serialize(): 5,
			},
		},
		{
			query: "select series_a | summarize.min from 0 to 120000",
			expected: map[string]float64{
				api.TagSet{"app": "web", "dc": "west"}.Serialize():  0,
				api.TagSet{"app": "web", "dc": "east"}.Serialize():  0,
				api.TagSet{"app": "fun", "dc": "north"}.Serialize(): 4,
			},
		},

		{
			query: "select series_a | summarize.max from 0 to 120000",
			expected: map[string]float64{
				api.TagSet{"app": "web", "dc": "west"}.Serialize():  6,
				api.TagSet{"app": "web", "dc": "east"}.Serialize():  2,
				api.TagSet{"app": "fun", "dc": "north"}.Serialize(): 6,
			},
		},
		{
			query: "select series_a | summarize.current from 0 to 120000",
			expected: map[string]float64{
				api.TagSet{"app": "web", "dc": "west"}.Serialize():  6,
				api.TagSet{"app": "web", "dc": "east"}.Serialize():  2,
				api.TagSet{"app": "fun", "dc": "north"}.Serialize(): 4,
			},
		},
		// recent
		{
			query: "select series_a | summarize.mean(60s) from 0 to 120000",
			expected: map[string]float64{
				api.TagSet{"app": "web", "dc": "west"}.Serialize():  13.0 / 3,
				api.TagSet{"app": "web", "dc": "east"}.Serialize():  1,
				api.TagSet{"app": "fun", "dc": "north"}.Serialize(): 5,
			},
		},
		{
			query: "select series_a | summarize.min(60s) from 0 to 120000",
			expected: map[string]float64{
				api.TagSet{"app": "web", "dc": "west"}.Serialize():  3,
				api.TagSet{"app": "web", "dc": "east"}.Serialize():  0,
				api.TagSet{"app": "fun", "dc": "north"}.Serialize(): 4,
			},
		},
		{
			query: "select series_a | summarize.max(60s) from 0 to 120000",
			expected: map[string]float64{
				api.TagSet{"app": "web", "dc": "west"}.Serialize():  6,
				api.TagSet{"app": "web", "dc": "east"}.Serialize():  2,
				api.TagSet{"app": "fun", "dc": "north"}.Serialize(): 6,
			},
		},
		// with NaNs
		{
			query: "select series_b | summarize.mean from 0 to 120000",
			expected: map[string]float64{
				api.TagSet{"dc": "west"}.Serialize(): 5,
				api.TagSet{"dc": "east"}.Serialize(): 3,
				api.TagSet{"dc": "miss"}.Serialize(): n,
			},
		},
		{
			query: "select series_b | summarize.min from 0 to 120000",
			expected: map[string]float64{
				api.TagSet{"dc": "west"}.Serialize(): 3,
				api.TagSet{"dc": "east"}.Serialize(): 2,
				api.TagSet{"dc": "miss"}.Serialize(): n,
			},
		},
		{
			query: "select series_b | summarize.max from 0 to 120000",
			expected: map[string]float64{
				api.TagSet{"dc": "west"}.Serialize(): 7,
				api.TagSet{"dc": "east"}.Serialize(): 5,
				api.TagSet{"dc": "miss"}.Serialize(): n,
			},
		},
	}

	for _, test := range tests {
		a := a.Contextf("Query %s", test.query)
		context := command.ExecutionContext{
			TimeseriesStorageAPI: comboAPI,
			MetricMetadataAPI:    comboAPI,
			FetchLimit:           100,
		}
		commandObject, err := parser.Parse(test.query)
		if err != nil {
			t.Fatalf("Error parsing command %s: %s", test.query, err.Error())
		}
		result, err := commandObject.Execute(context)
		if err != nil {
			t.Errorf("Error evaluating %s: %s", test.query, err.Error())
		}
		value := result.Body.([]command.QueryResult)[0]
		a.Eq(value.Type, "scalars") // Confirm that it's a scalar set.
		a.Contextf("number of results").Eq(len(value.Scalars), len(test.expected))
		for i := range value.Scalars {
			scalar := value.Scalars[i]
			if correct, ok := test.expected[scalar.TagSet.Serialize()]; ok {
				a.Contextf("value for %+v", scalar.TagSet).EqFloat(scalar.Value, correct, 1e-10)
			} else {
				a.Errorf("Unexpected tag set in result: %+v", scalar)
			}
		}
	}

}
