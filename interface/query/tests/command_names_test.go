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
	"testing"

	"github.com/square/metrics/api"
	"github.com/square/metrics/interface/query/command"
	"github.com/square/metrics/interface/query/parser"
	"github.com/square/metrics/testing_support/mocks"
)

func TestQueryNaming(t *testing.T) {
	fakeAPI := mocks.NewFakeMetricMetadataAPI()
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_1", api.TagSet{"dc": "west", "env": "production"}})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_1", api.TagSet{"dc": "east", "env": "staging"}})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_2", api.TagSet{"dc": "west", "env": "production"}})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_2", api.TagSet{"dc": "east", "env": "staging"}})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series-special#characters", api.TagSet{"dc": "east", "env": "staging"}})

	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"foo.bar.", api.TagSet{"qaz": "foo1"}})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{".foo.bar", api.TagSet{"qaz": "foo1"}})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"foo.2bar", api.TagSet{"qaz": "foo1"}})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"_names423.with_.dots_and_und3rsc0r3s", api.TagSet{"qaz": "foo1"}})

	fakeBackend := mocks.FakeTimeseriesStorageAPI{}
	tests := []struct {
		query        string
		expected     string
		expectedName string
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
			query:    "select transform.moving_average(series_2, 2h) from 0 to 0",
			expected: "transform.moving_average(series_2, 2h)",
		},
		{
			query:    "select filter.lowest_max(series_2, 6) from 0 to 0",
			expected: "filter.lowest_max(series_2, 6)",
		},
		// Tests for {annotations} in queries
		{
			query:        "select series_1{annotation override} from 0 to 0",
			expected:     "series_1 {annotation override}",
			expectedName: "annotation override",
		},
		{
			query:        "select aggregate.sum(series_1 group by dc,env)   {hello} from 0 to 0",
			expected:     "aggregate.sum(series_1 group by dc, env) {hello}",
			expectedName: "hello",
		},
		{
			query:        "select series_1 + series_2 {hello} from 0 to 0",
			expected:     "(series_1 + series_2 {hello})",
			expectedName: "(series_1 + hello)",
		},
		{
			query:        "select series_1 {goodbye} + series_2 {hello} from 0 to 0",
			expected:     "(series_1 {goodbye} + series_2 {hello})",
			expectedName: "(goodbye + hello)",
		},
		{
			query:        `select (series_1 + series_2) {varied whitespace and 0987654321~!@#$%^&*()||\\,,[]:<>.?/"''" punctuation is allowed} from 0 to 0`,
			expected:     `(series_1 + series_2) {varied whitespace and 0987654321~!@#$%^&*()||\\,,[]:<>.?/"''" punctuation is allowed}`,
			expectedName: `varied whitespace and 0987654321~!@#$%^&*()||\\,,[]:<>.?/"''" punctuation is allowed`,
		},
		{
			query:        "select series_1 | aggregate.sum {it's a sum} from 0 to 0",
			expected:     "aggregate.sum(series_1) {it's a sum}",
			expectedName: "it's a sum",
		},
		{
			query:        "select series_1 | aggregate.sum {it's a sum} | transform.derivative from 0 to 0",
			expected:     "transform.derivative(aggregate.sum(series_1) {it's a sum})",
			expectedName: "transform.derivative(it's a sum)",
		},
		{
			query:    "`series-special#characters`[app in ('test', \"test\") and not host match 'qaz'] from 0 to 0",
			expected: "`series-special#characters`[(app in (\"test\", \"test\") and not host match \"qaz\")]",
		},
		{
			query:    "series_1[foo = 'bar' or bar = 'foo' or qux != 'baz'] from 0 to 0",
			expected: `series_1[(foo = "bar" or (bar = "foo" or not qux = "baz"))]`,
		},
		{
			query:    "series_1[`foo-bar` = 'qaz' and `foo-bar` match 'x' and `foo-bar` in ('a', 'b')] from 0 to 0",
			expected: "series_1[(`foo-bar` = \"qaz\" and (`foo-bar` match \"x\" and `foo-bar` in (\"a\", \"b\")))]",
		},
		{
			query:    "_names423.with_.dots_and_und3rsc0r3s from 0 to 0",
			expected: "_names423.with_.dots_and_und3rsc0r3s",
		},
		{
			query:    "`foo.bar.` from 0 to 0",
			expected: "`foo.bar.`",
		},
		{
			query:    "`.foo.bar` from 0 to 0",
			expected: "`.foo.bar`",
		},
		{
			query:    "`foo.2bar` from 0 to 0",
			expected: "`foo.2bar`",
		},
	}
	for _, test := range tests {
		testCommand, err := parser.Parse(test.query)
		if err != nil {
			t.Fatalf("Unexpected error while parsing: %s", err.Error())
			return
		}
		if testCommand.Name() != "select" {
			t.Errorf("Expected select command but got %s", testCommand.Name())
			continue
		}
		rawResult, err := testCommand.Execute(command.ExecutionContext{
			TimeseriesStorageAPI: fakeBackend,
			MetricMetadataAPI:    fakeAPI,
			FetchLimit:           1000,
			Timeout:              0,
		})
		if err != nil {
			t.Errorf("Unexpected error while execution: %s", err.Error())
			continue
		}
		seriesListList, ok := rawResult.Body.([]command.QueryResult)
		if !ok || len(seriesListList) != 1 || seriesListList[0].Type != "series" {
			t.Errorf("expected query `%s` to produce []QueryResult of series list; got %+v :: %T", test.query, rawResult.Body, rawResult.Body)
			continue
		}
		actualQuery := seriesListList[0].Query
		if actualQuery != test.expected {
			t.Errorf("Expected Query:\n\t%s\n but got \n\t%s\n for query\n\t%s", test.expected, actualQuery, test.query)
		}
		actualName := seriesListList[0].Name
		expectedName := test.expectedName
		if test.expectedName == "" {
			expectedName = test.expected
		}
		if actualName != expectedName {
			t.Errorf("Expected Name:\n\t%s\n but got \n\t%s\n for query\n\t%s", expectedName, actualName, test.query)
		}
	}
}
