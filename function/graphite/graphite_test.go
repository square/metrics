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
	"github.com/square/metrics/internal"
	"github.com/square/metrics/mocks"
)

func TestApplyPattern(t *testing.T) {
	tests := []struct {
		pattern string
		metric  string
		success bool
		expect  api.TagSet
	}{
		{
			pattern: "this.is.a.graphite.metric",
			metric:  "this.is.a.graphite.metric",
			success: true,
			expect:  api.TagSet{"$graphite": "this.is.a.graphite.metric"},
		},
		{
			pattern: "this.is.a.graphite.metric",
			metric:  "this.is.a.different_graphite.metric",
			success: false,
		},
		{
			pattern: "this.is.a.graphite.metric",
			metric:  "this.is.a.different.graphite.metric",
			success: false,
		},
		{
			pattern: "this.is.a.graphite.metric",
			metric:  "this.is.a.graphite.metric.too",
			success: false,
		},
		{
			pattern: "this.is.a.graphite.metric",
			metric:  "this.is.a.graphite.metricQ",
			success: false,
		},
		{
			pattern: "this.is.a.%something%.metric",
			metric:  "this.is.a.graphite.metric",
			success: true,
			expect: api.TagSet{
				"$graphite": "this.is.a.graphite.metric",
				"something": "graphite",
			},
		},
		{
			pattern: "this.is.a.%something%.%type%",
			metric:  "this.is.a.graphite.metric",
			success: true,
			expect: api.TagSet{
				"$graphite": "this.is.a.graphite.metric",
				"something": "graphite",
				"type":      "metric",
			},
		},
		{
			pattern: "%word1%.%word2%.%word3%.%word4%.%word5%",
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
			pattern: "%app%.%datacenter%.cpu.%quantity%",
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
		rule, err := internal.Compile(internal.RawRule{Pattern: test.pattern, MetricKeyPattern: "$graphite"})
		if (err != nil) && test.success {
			t.Errorf("Expected success, but rule compilation failed for test %+v", test)
		}
		if err != nil {
			continue
		}
		result, ok := applyPattern(rule, test.metric)
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
	api.API
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
	store := testStore{&mocks.FakeApi{}}
	tests := []struct {
		pattern string
		expect  []api.TagSet
	}{
		{
			pattern: "%app%.%dc%.cpu.%quantity%",
			expect: []api.TagSet{
				{
					"$graphite": "server.north.cpu.mean",
					"app":       "server",
					"dc":        "north",
					"quantity":  "mean",
				},
				{
					"$graphite": "server.north.cpu.median",
					"app":       "server",
					"dc":        "north",
					"quantity":  "median",
				},
				{
					"$graphite": "server.south.cpu.mean",
					"app":       "server",
					"dc":        "south",
					"quantity":  "mean",
				},
				{
					"$graphite": "server.south.cpu.median",
					"app":       "server",
					"dc":        "south",
					"quantity":  "median",
				},
				{
					"$graphite": "proxy.south.cpu.mean",
					"app":       "proxy",
					"dc":        "south",
					"quantity":  "mean",
				},
				{
					"$graphite": "proxy.south.cpu.median",
					"app":       "proxy",
					"dc":        "south",
					"quantity":  "median",
				},
			},
		},
		{
			pattern: "host45.latency.http",
			expect: []api.TagSet{
				{
					"$graphite": "host45.latency.http",
				},
			},
		},
		{
			pattern: "%host%.latency.%method%",
			expect: []api.TagSet{
				{
					"$graphite": "host45.latency.http",
					"host":      "host45",
					"method":    "http",
				},
				{
					"$graphite": "host12.latency.http",
					"host":      "host12",
					"method":    "http",
				},
				{
					"$graphite": "host12.latency.rpc",
					"host":      "host12",
					"method":    "rpc",
				},
			},
		},
		{
			pattern: "does.not.exist",
			expect:  []api.TagSet{},
		},
		{
			pattern: "%does%.not.exist",
			expect:  []api.TagSet{},
		},
	}
	for testNumber, test := range tests {
		result, err := GetGraphiteMetrics(test.pattern, store)
		if err != nil {
			t.Fatalf("Unexpected error occurred: %s", err.Error())
		}
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
			if !expected.Equals(tagged.TagSet) {
				t.Errorf("Test #%d: Expected tagset %+v but got %+v", testNumber, expected, tagged.TagSet)
			}
		}
	}
}

func TestFailures(t *testing.T) {
	store := testStore{}
	tests := []struct {
		pattern string
	}{
		{
			pattern: "invalid.metric%",
		},
	}
	for testNumber, test := range tests {
		result, err := GetGraphiteMetrics(test.pattern, store)
		if err == nil {
			t.Fatalf("Expected error on input `%s`, test %d; got %+v", test, testNumber, result)
		}
	}
}
