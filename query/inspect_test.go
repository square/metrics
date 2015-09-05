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
	"time"

	// "github.com/square/metrics/api/backend"
	"github.com/square/metrics/testing_support/mocks"
)

func TestProfilerIntegration(t *testing.T) {
	t.Skip("This test is entirely broken. Postponing until proper cleanup can be done. Notes in-line")
	myAPI := mocks.NewFakeMetricMetadataAPI()
	fakeTimeStorage := mocks.FakeTimeseriesStorageAPI{}
	// 	myAPI := fakeAPI{
	// 	tagSets: map[string][]api.TagSet{"A": []api.TagSet{
	// 		{"x": "1", "y": "2"},
	// 		{"x": "2", "y": "2"},
	// 		{"x": "3", "y": "1"},
	// 	},
	// 		"B": []api.TagSet{
	// 			{"q": "foo"},
	// 			{"q": "bar"},
	// 		},
	// 		"C": []api.TagSet{
	// 			{"c": "1"},
	// 			{"c": "2"},
	// 			{"c": "3"},
	// 			{"c": "4"},
	// 			{"c": "5"},
	// 			{"c": "6"},
	// 		},
	// 	},
	// }

	// emptyGraphiteName := util.GraphiteMetric("")
	// myAPI.AddPairWithoutGraphite(api.TaggedMetric{"A", api.ParseTagSet("x=1,y=2")}, emptyGraphiteName)

	multiBackend := fakeTimeStorage

	testCases := []struct {
		query    string
		expected map[string]int
	}{
		{
			query: "describe all",
			expected: map[string]int{
				"describe all.Execute": 1,
				"api.GetAllMetrics":    1,
			},
		},
		{
			query: "select A from 0 to 0",
			expected: map[string]int{
				"select.Execute":      1,
				"fetchMultipleSeries": 1,
				"api.GetAllTags":      1,
				"fetchSingleSeries":   3,
			},
		},
		{
			query: "select A+A from 0 to 0",
			expected: map[string]int{
				"select.Execute":      1,
				"fetchMultipleSeries": 2,
				"api.GetAllTags":      2,
				"fetchSingleSeries":   6,
			},
		},
		{
			query: "select A+2 from 0 to 0",
			expected: map[string]int{
				"select.Execute":      1,
				"fetchMultipleSeries": 1,
				"api.GetAllTags":      1,
				"fetchSingleSeries":   3,
			},
		},
		{
			query: "select A where y = '2' from 0 to 0",
			expected: map[string]int{
				"select.Execute":      1,
				"fetchMultipleSeries": 1,
				"api.GetAllTags":      1,
				"fetchSingleSeries":   2,
			},
		},
		{
			query: "describe A",
			expected: map[string]int{
				"describe.Execute": 1,
				"api.GetAllTags":   1,
			},
		},
		{
			query: "describe metrics where y='2'",
			expected: map[string]int{
				"describe metrics.Execute": 1,
				"api.GetMetricsForTag":     1,
			},
		},
		{
			query: "describe all",
			expected: map[string]int{
				"describe all.Execute": 1,
				"api.GetAllMetrics":    1,
			},
		},
	}

	for _, test := range testCases {
		cmd, err := Parse(test.query)
		if err != nil {
			t.Error(err.Error())
			continue
		}
		profilingCommand, profiler := NewProfilingCommand(cmd)

		_, err = profilingCommand.Execute(ExecutionContext{
			TimeseriesStorageAPI: multiBackend,
			MetricMetadataAPI:    &myAPI,
			FetchLimit:           10000,
			Timeout:              time.Second * 4,
		})

		if err != nil {
			t.Fatal(err.Error())
		}
		list := profiler.All()
		counts := map[string]int{}
		for _, node := range list {
			counts[node.Name()]++
		}

		//TODO(cchandler): This added expectation demonstrates that this test has always
		//been broken. We'll have to clean this up later when we can address the time series
		//storage API.
		if len(test.expected) != len(list) {
			t.Errorf("The number of calls doesn't match the expected amount: %+v %+v", test.expected, list)
		}

		for name, count := range counts {
			if test.expected[name] != count {
				t.Errorf("Expected %+v but got %+v (from %+v)", test.expected, counts, list)
				break
			}
		}

	}

}
