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

package graphite

import (
	"testing"

	"github.com/square/metrics/api"
)

func TestApplyPattern(t *testing.T) {
	tests := []struct {
		pieces  []string
		metric  string
		success bool
		expect  api.TagSet
	}{
		{
			pieces:  []string{"this.is.a.graphite.metric"},
			metric:  "this.is.a.graphite.metric",
			success: true,
			expect:  api.TagSet{"$graphite": "this.is.a.graphite.metric"},
		},
		{
			pieces:  []string{"this.is.a.graphite.metric"},
			metric:  "this.is.a.different_graphite.metric",
			success: false,
		},
		{
			pieces:  []string{"this.is.a.graphite.metric"},
			metric:  "this.is.a.different.graphite.metric",
			success: false,
		},
		{
			pieces:  []string{"this.is.a.graphite.metric"},
			metric:  "this.is.a.graphite.metric.too",
			success: false,
		},
		{
			pieces:  []string{"this.is.a.graphite.metric"},
			metric:  "this.is.a.graphite.metricQ",
			success: false,
		},
		{
			pieces:  []string{"this.is.a.", "something", ".metric"},
			metric:  "this.is.a.graphite.metric",
			success: true,
			expect: api.TagSet{
				"$graphite": "this.is.a.graphite.metric",
				"something": "graphite",
			},
		},
		{
			pieces:  []string{"this.is.a.", "something", ".", "type", ""},
			metric:  "this.is.a.graphite.metric",
			success: true,
			expect: api.TagSet{
				"$graphite": "this.is.a.graphite.metric",
				"something": "graphite",
				"type":      "metric",
			},
		},
		{
			pieces:  []string{"", "word1", ".", "word2", ".", "word3", ".", "word4", ".", "word5", ""},
			metric:  "this.is.a.graphite.metric",
			success: true,
			expect: api.TagSet{
				"$graphite": "this.is.a.graphite.metric",
				"word1":     "this",
				"word2":     "is",
				"word3":     "a",
				"word4":     "graphite",
				"word5":     "metric",
			},
		},
		{
			pieces:  []string{"", "app", ".", "datacenter", ".cpu.", "quantity", ""},
			metric:  "metrics-query-engine.north.cpu.total",
			success: true,
			expect: api.TagSet{
				"$graphite":  "metrics-query-engine.north.cpu.total",
				"app":        "metrics-query-engine",
				"datacenter": "north",
				"quantity":   "total",
			},
		},
	}
	for i, test := range tests {
		result, ok := applyPattern(test.pieces, test.metric)
		if ok != test.success {
			t.Errorf("Test i = %d, test = %+v: didn't expect ok = %t", i, test, ok)
			continue
		}
		if !test.expect.Equals(result.TagSet) {
			t.Errorf("Expected %+v but got %+v", test.expect, result)
		}
	}
}

type testStore struct {
}

func (t testStore) AddMetric(api.TaggedMetric) error {
	return nil
}
func (t testStore) GetAllMetrics() ([]api.MetricKey, error) {
	return nil, nil
}
func (t testStore) GetAllTags(api.MetricKey) ([]api.TagSet, error) {
	return nil, nil
}
func (t testStore) GetMetricsForTag(string, string) ([]api.MetricKey, error) {
	return nil, nil
}
func (t testStore) RemoveMetric(api.TaggedMetric) error {
	return nil
}
func (t testStore) ToGraphiteName(api.TaggedMetric) (api.GraphiteMetric, error) {
	return "", nil
}
func (t testStore) ToTaggedName(api.GraphiteMetric) (api.TaggedMetric, error) {
	return api.TaggedMetric{}, nil
}

func (t testStore) GetAllGraphiteMetrics() ([]api.GraphiteMetric, error) {
	return []api.GraphiteMetric{
		"server.north.cpu.mean",
		"server.north.cpu.median",
		"server.south.cpu.mean",
		"server.south.cpu.median",
		"proxy.south.cpu.mean",
		"proxy.south.cpu.median",
		"host45.latency.http",
		"host12.latency.http",
		"host12.latency.rpc",
	}, nil
}
func (t testStore) AddGraphiteMetric(metric api.GraphiteMetric) error {
	// A no-op for testing
	return nil
}

func TestGetGraphiteMetrics(t *testing.T) {
	store := testStore{}
	tests := []struct {
		pattern string
		expect  []api.TaggedMetric
	}{
		{
			pattern: "%app%.%dc%.cpu.%quantity%",
			expect: []api.TaggedMetric{
				{
					MetricKey: "$graphite",
					TagSet: api.TagSet{
						"$graphite": "server.north.cpu.mean",
						"app":       "server",
						"dc":        "north",
						"quantity":  "mean",
					},
				},
				{
					MetricKey: "$graphite",
					TagSet: api.TagSet{
						"$graphite": "server.north.cpu.median",
						"app":       "server",
						"dc":        "north",
						"quantity":  "median",
					},
				},
				{
					MetricKey: "$graphite",
					TagSet: api.TagSet{
						"$graphite": "server.south.cpu.mean",
						"app":       "server",
						"dc":        "south",
						"quantity":  "mean",
					},
				},
				{
					MetricKey: "$graphite",
					TagSet: api.TagSet{
						"$graphite": "server.south.cpu.median",
						"app":       "server",
						"dc":        "south",
						"quantity":  "median",
					},
				},
				{
					MetricKey: "$graphite",
					TagSet: api.TagSet{
						"$graphite": "proxy.south.cpu.mean",
						"app":       "proxy",
						"dc":        "south",
						"quantity":  "mean",
					},
				},
				{
					MetricKey: "$graphite",
					TagSet: api.TagSet{
						"$graphite": "proxy.south.cpu.median",
						"app":       "proxy",
						"dc":        "south",
						"quantity":  "median",
					},
				},
			},
		},
		{
			pattern: "does.not.exist",
			expect:  []api.TaggedMetric{},
		},
		{
			pattern: "invalid.metric%",
			expect:  []api.TaggedMetric{},
		},
		{
			pattern: "host45.latency.http",
			expect: []api.TaggedMetric{
				{
					MetricKey: "$graphite",
					TagSet: api.TagSet{
						"$graphite": "host45.latency.http",
					},
				},
			},
		},
		{
			pattern: "%host%.latency.%method%",
			expect: []api.TaggedMetric{
				{
					MetricKey: "$graphite",
					TagSet: api.TagSet{
						"$graphite": "host45.latency.http",
						"host":      "host45",
						"method":    "http",
					},
				},
				{
					MetricKey: "$graphite",
					TagSet: api.TagSet{
						"$graphite": "host12.latency.http",
						"host":      "host12",
						"method":    "http",
					},
				},
				{
					MetricKey: "$graphite",
					TagSet: api.TagSet{
						"$graphite": "host12.latency.rpc",
						"host":      "host12",
						"method":    "rpc",
					},
				},
			},
		},
	}
	for testNumber, test := range tests {
		result := GetGraphiteMetrics(test.pattern, store)
		if len(test.expect) != len(result) {
			t.Errorf("Test #%d: Expected %d but got %d results: %+v, %+v",
				testNumber,
				len(test.expect),
				len(result),
				test.expect,
				result,
			)
			continue
		}
		for i, tagged := range result {
			expected := test.expect[i]
			if expected.MetricKey != tagged.MetricKey {
				t.Errorf("Test #%d: Expected metric key %s but got %s", testNumber, expected.MetricKey, tagged.MetricKey)
				break
			}
			if !expected.TagSet.Equals(tagged.TagSet) {
				t.Errorf("Test #%d: Expected tagset %+v but got %+v", testNumber, expected.TagSet, tagged.TagSet)
			}
		}
	}
}
