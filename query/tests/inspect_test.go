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

package tests

import (
	"testing"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/inspect"
	"github.com/square/metrics/query/command"
	"github.com/square/metrics/query/parser"
	"github.com/square/metrics/testing_support/mocks"

	"golang.org/x/net/context"
)

func TestProfilerIntegration(t *testing.T) {
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

	myAPI.AddPairWithoutGraphite(api.TaggedMetric{MetricKey: "A", TagSet: api.TagSet{"x": "1", "y": "2"}})
	myAPI.AddPairWithoutGraphite(api.TaggedMetric{MetricKey: "A", TagSet: api.TagSet{"x": "2", "y": "2"}})
	myAPI.AddPairWithoutGraphite(api.TaggedMetric{MetricKey: "A", TagSet: api.TagSet{"x": "3", "y": "1"}})

	myAPI.AddPairWithoutGraphite(api.TaggedMetric{MetricKey: "B", TagSet: api.TagSet{"q": "foo"}})
	myAPI.AddPairWithoutGraphite(api.TaggedMetric{MetricKey: "B", TagSet: api.TagSet{"q": "bar"}})

	myAPI.AddPairWithoutGraphite(api.TaggedMetric{MetricKey: "C", TagSet: api.TagSet{"c": "1"}})
	myAPI.AddPairWithoutGraphite(api.TaggedMetric{MetricKey: "C", TagSet: api.TagSet{"c": "2"}})
	myAPI.AddPairWithoutGraphite(api.TaggedMetric{MetricKey: "C", TagSet: api.TagSet{"c": "3"}})
	myAPI.AddPairWithoutGraphite(api.TaggedMetric{MetricKey: "C", TagSet: api.TagSet{"c": "4"}})
	myAPI.AddPairWithoutGraphite(api.TaggedMetric{MetricKey: "C", TagSet: api.TagSet{"c": "5"}})
	myAPI.AddPairWithoutGraphite(api.TaggedMetric{MetricKey: "C", TagSet: api.TagSet{"c": "6"}})

	testCases := []struct {
		query    string
		expected map[string]int
	}{
		{
			query: "describe all",
			expected: map[string]int{
				"describe all.Execute": 1,
				"Mock GetAllMetrics":   1,
			},
		},
		{
			query: "select A from 0 to 0",
			expected: map[string]int{
				"select.Execute":               1,
				"Mock FetchMultipleTimeseries": 1,
				"Mock GetAllTags":              1,
				"Mock FetchSingleTimeseries":   3,
			},
		},
		{
			query: "select A+A from 0 to 0",
			expected: map[string]int{
				"select.Execute":               1,
				"Mock FetchMultipleTimeseries": 1,
				"Mock GetAllTags":              1,
				"Mock FetchSingleTimeseries":   3,
			},
		},
		{
			query: `select A+A[foo != "blah"] from 0 to 0`,
			expected: map[string]int{
				"select.Execute":               1,
				"Mock FetchMultipleTimeseries": 2,
				"Mock GetAllTags":              2,
				"Mock FetchSingleTimeseries":   6,
			},
		},
		{
			query: "select A+2 from 0 to 0",
			expected: map[string]int{
				"select.Execute":               1,
				"Mock FetchMultipleTimeseries": 1,
				"Mock GetAllTags":              1,
				"Mock FetchSingleTimeseries":   3,
			},
		},
		{
			query: "select A where y = '2' from 0 to 0",
			expected: map[string]int{
				"select.Execute":               1,
				"Mock FetchMultipleTimeseries": 1,
				"Mock GetAllTags":              1,
				"Mock FetchSingleTimeseries":   2,
			},
		},
		{
			query: "describe A",
			expected: map[string]int{
				"describe.Execute": 1,
				"Mock GetAllTags":  1,
			},
		},
		{
			query: "describe metrics where y='2'",
			expected: map[string]int{
				"describe metrics.Execute": 1,
				"Mock GetMetricsForTag":    1,
			},
		},
		{
			query: "describe all",
			expected: map[string]int{
				"describe all.Execute": 1,
				"Mock GetAllMetrics":   1,
			},
		},
		{
			query: "select transform.timeshift(A, -5m) + transform.timeshift(A, -5m) from 0 to 0",
			expected: map[string]int{
				"select.Execute":               1,
				"Mock FetchMultipleTimeseries": 1,
				"Mock GetAllTags":              1,
				"Mock FetchSingleTimeseries":   3,
			},
		},
		{
			query: "select transform.timeshift(A | transform.timeshift(-3m), -2m) + transform.timeshift(A, -5m) from 0 to 0",
			expected: map[string]int{
				"select.Execute":               1,
				"Mock FetchMultipleTimeseries": 1,
				"Mock GetAllTags":              1,
				"Mock FetchSingleTimeseries":   3,
			},
		},
	}

	for _, test := range testCases {
		cmd, err := parser.Parse(test.query)
		if err != nil {
			t.Error(err.Error())
			continue
		}
		profiler := inspect.New()
		profilingCommand := command.NewProfilingCommandWithProfiler(cmd, profiler)

		_, err = profilingCommand.Execute(command.ExecutionContext{
			TimeseriesStorageAPI: fakeTimeStorage,
			MetricMetadataAPI:    myAPI,
			FetchLimit:           10000,
			Timeout:              time.Second * 4,
			Ctx:                  context.Background(),
		})

		if err != nil {
			t.Fatal(err.Error())
		}
		list := profiler.All()
		counts := map[string]int{}
		for _, node := range list {
			counts[node.Name]++
		}

		if len(test.expected) != len(counts) {
			t.Errorf("The number of calls doesn't match the expected amount.")
			t.Errorf("Expected %+v, but got %+v", test.expected, counts)
		}

		for name, count := range test.expected {
			if counts[name] != count {
				t.Errorf("Unexpected problem in query '%s'", test.query)
				t.Errorf("Expected `%s` to have %d occurrences, but had %d\n", name, count, counts[name])
				t.Errorf("Expected: %+v\nBut got: %+v\n", test.expected, counts)
				break
			}
		}

	}

}
