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

// Integration test for the query execution.
package query

import (
	"testing"

	"github.com/square/metrics/api"
	"github.com/square/metrics/optimize"
	"github.com/square/metrics/testing_support/mocks"
)

func TestNaming(t *testing.T) {
	fakeAPI := mocks.NewFakeMetricMetadataAPI()
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_1", api.ParseTagSet("dc=west,env=production")})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_1", api.ParseTagSet("dc=east,env=staging")})

	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_2", api.ParseTagSet("dc=west,env=production")})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_2", api.ParseTagSet("dc=east,env=staging")})
	fakeBackend := mocks.FakeTimeseriesStorageAPI{}
	tests := []struct {
		query    string
		expected string
	}{
		{
			query:    "select series_1 from 0 to 0",
			expected: "series_1",
		},
		{
			query:    "select series_1 + 17 from 0 to 0",
			expected: "(series_1 + 17)",
		},
		{
			query:    "select series_1 + 2342.32 from 0 to 0",
			expected: "(series_1 + 2342.32)",
		},
		{
			query:    "select series_1*17 from 0 to 0",
			expected: "(series_1 * 17)",
		},
		{
			query:    "select aggregate.sum(series_1) from 0 to 0",
			expected: "aggregate.sum(series_1)",
		},
		{
			query:    "select aggregate.sum(series_1 group by dc) from 0 to 0",
			expected: "aggregate.sum(series_1 group by dc)",
		},
		{
			query:    "select aggregate.sum(series_1 group by dc, env) from 0 to 0",
			expected: "aggregate.sum(series_1 group by dc, env)",
		},
		{
			query:    "select aggregate.sum(series_1 collapse by dc) from 0 to 0",
			expected: "aggregate.sum(series_1 collapse by dc)",
		},
		{
			query:    "select aggregate.sum(series_1 collapse by dc,env) from 0 to 0",
			expected: "aggregate.sum(series_1 collapse by dc, env)",
		},
		{
			query:    "select transform.alias(aggregate.sum(series_1 group by dc, env), 'hello') from 0 to 0",
			expected: "hello",
		},
		{
			query:    "select transform.moving_average(series_2, 2h) from 0 to 0",
			expected: "transform.moving_average(series_2, 2h)",
		},
		{
			query:    "select filter.lowest_max(series_2, 6) from 0 to 0",
			expected: "filter.lowest_max(series_2, 6)",
		},
	}
	for _, test := range tests {
		command, err := Parse(test.query)
		if err != nil {
			t.Fatalf("Unexpected error while parsing: %s", err.Error())
			return
		}
		if command.Name() != "select" {
			t.Errorf("Expected select command but got %s", command.Name())
			continue
		}
		rawResult, err := command.Execute(ExecutionContext{
			TimeseriesStorageAPI:      fakeBackend,
			MetricMetadataAPI:         fakeAPI,
			FetchLimit:                1000,
			Timeout:                   0,
			OptimizationConfiguration: optimize.NewOptimizationConfiguration(),
		})
		if err != nil {
			t.Errorf("Unexpected error while execution: %s", err.Error())
			continue
		}
		seriesListList, ok := rawResult.Body.([]QuerySeriesList)
		if !ok || len(seriesListList) != 1 {
			t.Errorf("expected query `%s` to produce []QuerySeriesList; got %+v :: %T", test.query, rawResult.Body, rawResult.Body)
			continue
		}
		actual := seriesListList[0].Name
		if actual != test.expected {
			t.Errorf("Expected `%s` but got `%s` for query `%s`", test.expected, actual, test.query)
			continue
		}
	}
}

// TODO: Use this test with the NEW query system
func TestQuery(t *testing.T) {
	fakeAPI := mocks.NewFakeMetricMetadataAPI()
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_1", api.ParseTagSet("dc=west,env=production")})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_1", api.ParseTagSet("dc=east,env=staging")})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_2", api.ParseTagSet("dc=west,env=production")})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_2", api.ParseTagSet("dc=east,env=staging")})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series-special#characters", api.ParseTagSet("dc=east,env=staging")})

	fakeBackend := mocks.FakeTimeseriesStorageAPI{}
	tests := []struct {
		query    string
		expected string
	}{
		{
			query:    "select series_1 from 0 to 0",
			expected: "series_1",
		},
		{
			query:    "select series_1 + 17 from 0 to 0",
			expected: "(series_1 + 17)",
		},
		{
			query:    "select series_1 + 2342.32 from 0 to 0",
			expected: "(series_1 + 2342.32)",
		},
		{
			query:    "select series_1*17 from 0 to 0",
			expected: "(series_1 * 17)",
		},
		{
			query:    "select aggregate.sum(series_1) from 0 to 0",
			expected: "aggregate.sum(series_1)",
		},
		{
			query:    "select aggregate.sum(series_1 group by dc) from 0 to 0",
			expected: "aggregate.sum(series_1 group by dc)",
		},
		{
			query:    "select aggregate.sum(series_1 group by dc,env) from 0 to 0",
			expected: "aggregate.sum(series_1 group by dc, env)",
		},
		{
			query:    "select aggregate.sum(series_1 collapse by dc) from 0 to 0",
			expected: "aggregate.sum(series_1 collapse by dc)",
		},
		{
			query:    "select aggregate.sum(series_1 collapse by dc,env) from 0 to 0",
			expected: "aggregate.sum(series_1 collapse by dc, env)",
		},
		{
			query:    "select transform.alias(aggregate.sum(series_1 group by dc,env), 'hello') from 0 to 0",
			expected: "transform.alias(aggregate.sum(series_1 group by dc, env), \"hello\")",
		},
		{
			query:    "select transform.moving_average(series_2, 2h) from 0 to 0",
			expected: "transform.moving_average(series_2, 2h)",
		},
		{
			query:    "select filter.lowest_max(series_2, 6) from 0 to 0",
			expected: "filter.lowest_max(series_2, 6)",
		},
		// Tests for {annotations} in queries
		{
			query:    "select series_1{annotation override} from 0 to 0",
			expected: "series_1 {annotation override}",
		},
		{
			query:    "select aggregate.sum(series_1 group by dc,env)   {hello} from 0 to 0",
			expected: "aggregate.sum(series_1 group by dc, env) {hello}",
		},
		{
			query:    "select series_1 + series_2 {hello} from 0 to 0",
			expected: "(series_1 + series_2 {hello})",
		},
		{
			query:    "select series_1 {goodbye} + series_2 {hello} from 0 to 0",
			expected: "(series_1 {goodbye} + series_2 {hello})",
		},
		{
			query:    `select (series_1 + series_2) {varied whitespace and 0987654321~!@#$%^&*()||\\,,[]:<>.?/"''" punctuation is allowed} from 0 to 0`,
			expected: `(series_1 + series_2) {varied whitespace and 0987654321~!@#$%^&*()||\\,,[]:<>.?/"''" punctuation is allowed}`,
		},
		{
			query:    "select series_1 | aggregate.sum {it's a sum} from 0 to 0",
			expected: "aggregate.sum(series_1) {it's a sum}",
		},
		{
			query:    "select series_1 | aggregate.sum {it's a sum} | transform.derivative from 0 to 0",
			expected: "transform.derivative(aggregate.sum(series_1) {it's a sum})",
		},
		{
			query:    "`series-special#characters`[app in ('test', \"test\") and not host match 'qaz'] from 0 to 0",
			expected: "`series-special#characters`[(app in (\"test\", \"test\") and not host match \"qaz\")]",
		},
		{
			query:    "series_1[foo = 'bar' or bar = 'foo' or qux != 'baz'] from 0 to 0",
			expected: `series_1[(foo = "bar" or (bar = "foo" or not qux = "baz"))]`,
		},
	}
	for _, test := range tests {
		command, err := Parse(test.query)
		if err != nil {
			t.Fatalf("Unexpected error while parsing: %s", err.Error())
			return
		}
		if command.Name() != "select" {
			t.Errorf("Expected select command but got %s", command.Name())
			continue
		}
		rawResult, err := command.Execute(ExecutionContext{
			TimeseriesStorageAPI:      fakeBackend,
			MetricMetadataAPI:         fakeAPI,
			FetchLimit:                1000,
			Timeout:                   0,
			OptimizationConfiguration: optimize.NewOptimizationConfiguration(),
		})
		if err != nil {
			t.Errorf("Unexpected error while execution: %s", err.Error())
			continue
		}
		seriesListList, ok := rawResult.Body.([]QuerySeriesList)
		if !ok || len(seriesListList) != 1 {
			t.Errorf("expected query `%s` to produce []QuerySeriesList; got %+v :: %T", test.query, rawResult.Body, rawResult.Body)
			continue
		}
		actual := seriesListList[0].Query
		if actual != test.expected {
			t.Errorf("Expected:\n\t%s\n but got \n\t%s\n for query `%s`", test.expected, actual, test.query)
			continue
		}
	}
}
