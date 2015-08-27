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
	"github.com/square/metrics/api/backend"
)

type fakeAPI struct {
	tagSets map[string][]api.TagSet
}

func (a fakeAPI) AddMetric(metric api.TaggedMetric) error {
	// NOTHING
	return nil
}

func (a fakeAPI) AddMetrics(metrics []api.TaggedMetric) error {
	// NOTHING
	return nil
}

func (a fakeAPI) RemoveMetric(metric api.TaggedMetric) error {
	// NOTHING
	return nil
}

func (a fakeAPI) ToGraphiteName(metric api.TaggedMetric) (api.GraphiteMetric, error) {
	return api.GraphiteMetric(metric.MetricKey), nil
}

func (a fakeAPI) ToTaggedName(metric api.GraphiteMetric) (api.TaggedMetric, error) {
	return api.TaggedMetric{
		MetricKey: api.MetricKey(metric),
		TagSet:    api.NewTagSet(),
	}, nil
}

func (a fakeAPI) GetAllTags(metricKey api.MetricKey) ([]api.TagSet, error) {
	return a.tagSets[string(metricKey)], nil
}

func (a fakeAPI) GetAllMetrics() ([]api.MetricKey, error) {
	list := []api.MetricKey{}
	for metric := range a.tagSets {
		list = append(list, api.MetricKey(metric))
	}
	return list, nil
}

func (a fakeAPI) GetMetricsForTag(tagKey, tagValue string) ([]api.MetricKey, error) {
	list := []api.MetricKey{}
MetricLoop:
	for metric, tagsets := range a.tagSets {
		for _, tagset := range tagsets {
			for key, val := range tagset {
				if key == tagKey && val == tagValue {
					list = append(list, api.MetricKey(metric))
					continue MetricLoop
				}
			}
		}
	}
	return list, nil
}

type fakeBackend struct {
}

func (f fakeBackend) FetchSingleSeries(request api.FetchSeriesRequest) (api.Timeseries, error) {
	return api.Timeseries{}, nil
}

func TestProfilerIntegration(t *testing.T) {
	myAPI := fakeAPI{
		tagSets: map[string][]api.TagSet{"A": []api.TagSet{
			{"x": "1", "y": "2"},
			{"x": "2", "y": "2"},
			{"x": "3", "y": "1"},
		},
			"B": []api.TagSet{
				{"q": "foo"},
				{"q": "bar"},
			},
			"C": []api.TagSet{
				{"c": "1"},
				{"c": "2"},
				{"c": "3"},
				{"c": "4"},
				{"c": "5"},
				{"c": "6"},
			},
		},
	}
	myBackend := api.ProfilingBackend{fakeBackend{}}
	multiBackend := api.ProfilingMultiBackend{backend.NewSequentialMultiBackend(myBackend)}

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
			Backend:    multiBackend,
			API:        myAPI,
			FetchLimit: 10000,
			Timeout:    time.Second * 4,
		})

		if err != nil {
			t.Fatal(err.Error())
		}
		list := profiler.All()
		counts := map[string]int{}
		for _, node := range list {
			counts[node.Name()]++
		}
		for name, count := range counts {
			if test.expected[name] != count {
				t.Errorf("Expected %+v but got %+v (from %+v)", test.expected, counts, list)
				break
			}
		}

	}

}
