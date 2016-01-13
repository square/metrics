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

	"github.com/square/metrics/api"
	"github.com/square/metrics/inspect"
	"github.com/square/metrics/optimize"
	"github.com/square/metrics/testing_support/mocks"
)

func TestProfilerIntegration(t *testing.T) {
	fakeConverter, fakeAPI := mocks.NewFakeGraphiteConverter([]api.TaggedMetric{
		{"A", api.TagSet{"x": "1", "y": "2"}},
		{"A", api.TagSet{"x": "2", "y": "2"}},
		{"A", api.TagSet{"x": "3", "y": "1"}},

		{"B", api.TagSet{"q": "foo"}},
		{"B", api.TagSet{"q": "bar"}},

		{"C", api.TagSet{"c": "1"}},
		{"C", api.TagSet{"c": "2"}},
		{"C", api.TagSet{"c": "3"}},
		{"C", api.TagSet{"c": "4"}},
		{"C", api.TagSet{"c": "5"}},
		{"C", api.TagSet{"c": "6"}},
	})
	fakeTimeStorage := mocks.FakeTimeseriesStorageAPI{AlwaysReturnData: true}

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
	}

	for _, test := range testCases {
		cmd, err := Parse(test.query)
		if err != nil {
			t.Error(err.Error())
			continue
		}
		profiler := inspect.New()
		profilingCommand := NewProfilingCommandWithProfiler(cmd, profiler)

		_, err = profilingCommand.Execute(ExecutionContext{
			MetricConverter:           fakeConverter,
			TimeseriesStorageAPI:      fakeTimeStorage,
			MetricMetadataAPI:         fakeAPI,
			FetchLimit:                10000,
			Timeout:                   time.Second * 4,
			OptimizationConfiguration: optimize.NewOptimizationConfiguration(),
		})

		if err != nil {
			t.Fatal(err.Error())
		}
		list := profiler.All()
		counts := map[string]int{}
		for _, node := range list {
			counts[node.Name()]++
		}

		if len(test.expected) != len(counts) {
			t.Errorf("The number of calls doesn't match the expected amount.")
			t.Errorf("Expected %+v, but got %+v", test.expected, counts)
		}

		for name, count := range test.expected {
			if counts[name] != count {
				t.Errorf("Expected `%s` to have %d occurrences, but had %d\n", name, count, counts[name])
				t.Errorf("Expected: %+v\nBut got: %+v\n", test.expected, counts)
				break
			}
		}

	}

}
