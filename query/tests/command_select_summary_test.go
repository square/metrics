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

/*
func TestSelectSummary(t *testing.T) {
	a := assert.New(t)
	testTimerange, err := api.NewSnappedTimerange(0, 5*30000, 30000)
	if err != nil {
		t.Fatalf("Error creating timerange for test: %s", err.Error())
	}

	n := math.NaN()

	comboAPI := mocks.NewComboAPI(
		// timerange
		testTimerange,
		// series_a
		api.Timeseries{Values: []float64{1, 2, 3, 4, 5}, TagSet: api.TagSet{"metric": "series_a", "app": "web", "dc": "west"}},
		api.Timeseries{Values: []float64{1, 1, 1, 1, 1}, TagSet: api.TagSet{"metric": "series_a", "app": "web", "dc": "east"}},
		api.Timeseries{Values: []float64{5, 5, 5, 5, 5}, TagSet: api.TagSet{"metric": "series_a", "app": "fun", "dc": "north"}},
		// series_b
		api.Timeseries{Values: []float64{3, n, 7, n, n}, TagSet: api.TagSet{"metric": "series_b", "dc": "west"}},
		api.Timeseries{Values: []float64{n, n, 5, 2, 2}, TagSet: api.TagSet{"metric": "series_b", "dc": "east"}},
	)

	type test struct {
		query    string
		expected map[string]float64
	}

	tests := []test{
		{
			query: "select series_a | summary.mean from 0 to 150000",
			expected: map[string]float64{
				api.TagSet{"app": "web", "dc": "west"}.Serialize():  3,
				api.TagSet{"app": "web", "dc": "east"}.Serialize():  1,
				api.TagSet{"app": "fun", "dc": "north"}.Serialize(): 5,
			},
		},
	}

	for _, test := range tests {
		a := a.Contextf("Query %s", test.query)
		context := command.ExecutionContext{
			TimeseriesStorageAPI: comboAPI,
			MetricMetadataAPI:    comboAPI,
			Timerange:            testTimerange,
		}
		command, err := parser.Parse(test.query)
		if err != nil {
			t.Fatalf("Error parsing command %s: %s", test.query, err.Error())
		}
		result, err := command.Execute(context)
		if err != nil {
			t.Errorf("Error evaluating %s: %s", test.query, err.Error())
		}
		result := result.Body.([]command.QuerySeriesList)[0].Series
	}

}*/
