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
