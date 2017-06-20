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
	"context"
	"testing"

	"github.com/square/metrics/api"
	"github.com/square/metrics/metric_metadata"
	"github.com/square/metrics/query/command"
	"github.com/square/metrics/query/parser"
	"github.com/square/metrics/query/predicate"
	"github.com/square/metrics/testing_support/assert"
	"github.com/square/metrics/testing_support/mocks"
)

func TestCommand_Describe(t *testing.T) {
	fakeAPI := mocks.NewFakeMetricMetadataAPI()
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{MetricKey: "series_0", TagSet: api.TagSet{"dc": "west", "env": "production", "host": "a"}})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{MetricKey: "series_0", TagSet: api.TagSet{"dc": "west", "env": "staging", "host": "b"}})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{MetricKey: "series_0", TagSet: api.TagSet{"dc": "east", "env": "production", "host": "c"}})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{MetricKey: "series_0", TagSet: api.TagSet{"dc": "east", "env": "staging", "host": "d"}})

	for _, test := range []struct {
		query          string
		metricmetadata metadata.MetricAPI
		expected       map[string][]string
	}{
		{"describe series_0", fakeAPI, map[string][]string{"dc": {"east", "west"}, "env": {"production", "staging"}, "host": {"a", "b", "c", "d"}}},
		{"describe`series_0`", fakeAPI, map[string][]string{"dc": {"east", "west"}, "env": {"production", "staging"}, "host": {"a", "b", "c", "d"}}},
		{"describe series_0 where dc='west'", fakeAPI, map[string][]string{"dc": {"west"}, "env": {"production", "staging"}, "host": {"a", "b"}}},
		{"describe`series_0`where(dc='west')", fakeAPI, map[string][]string{"dc": {"west"}, "env": {"production", "staging"}, "host": {"a", "b"}}},
		{"describe series_0 where dc='west' or env = 'production'", fakeAPI, map[string][]string{"dc": {"east", "west"}, "env": {"production", "staging"}, "host": {"a", "b", "c"}}},
		{"describe series_0 where`dc`='west'or`env`='production'", fakeAPI, map[string][]string{"dc": {"east", "west"}, "env": {"production", "staging"}, "host": {"a", "b", "c"}}},
		{"describe series_0 where dc='west' or env = 'production' and doesnotexist = ''", fakeAPI, map[string][]string{"dc": {"west"}, "env": {"production", "staging"}, "host": {"a", "b"}}},
		{"describe series_0 where env = 'production' and doesnotexist = '' or dc = 'west'", fakeAPI, map[string][]string{"dc": {"west"}, "env": {"production", "staging"}, "host": {"a", "b"}}},
		{"describe series_0 where (dc='west' or env = 'production') and doesnotexist = ''", fakeAPI, map[string][]string{}},
		{"describe series_0 where(dc='west' or env = 'production')and`doesnotexist` = ''", fakeAPI, map[string][]string{}},
	} {
		a := assert.New(t).Contextf("query=%s", test.query)
		testCommand, err := parser.Parse(test.query)
		a.CheckError(err)

		a.EqString(testCommand.Name(), "describe")
		fakeTimeseriesStorage := mocks.FakeTimeseriesStorageAPI{}
		rawResult, err := testCommand.Execute(command.ExecutionContext{
			TimeseriesStorageAPI: fakeTimeseriesStorage,
			MetricMetadataAPI:    test.metricmetadata,
			FetchLimit:           1000,
			Timeout:              0,
			Ctx:                  context.Background(),
		})
		a.CheckError(err)
		a.Eq(rawResult.Body, test.expected)
	}

	// Test AdditionalConstraints with describe commands.
	a := assert.New(t).Contextf("Checking AdditionalConstraints")
	testCommand, err := parser.Parse(`describe series_0`)
	a.CheckError(err)
	rawResult, err := testCommand.Execute(command.ExecutionContext{
		TimeseriesStorageAPI: mocks.FakeTimeseriesStorageAPI{},
		MetricMetadataAPI:    fakeAPI,
		FetchLimit:           1000,
		Timeout:              0,
		Ctx:                  context.Background(),
		AdditionalConstraints: predicate.ListMatcher{Tag: "dc", Values: []string{"west"}},
	})
	a.CheckError(err)
	a.Eq(rawResult.Body, map[string][]string{"dc": {"west"}, "env": {"production", "staging"}, "host": {"a", "b"}})
}

func TestCommand_DescribeAll(t *testing.T) {
	fakeAPI := mocks.NewFakeMetricMetadataAPI()
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{MetricKey: "series_0", TagSet: api.TagSet{}})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{MetricKey: "series_1", TagSet: api.TagSet{}})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{MetricKey: "series_2", TagSet: api.TagSet{}})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{MetricKey: "series_3", TagSet: api.TagSet{}})

	for _, test := range []struct {
		query          string
		metricmetadata metadata.MetricAPI
		expected       []api.MetricKey
	}{
		{"describe all", fakeAPI, []api.MetricKey{"series_0", "series_1", "series_2", "series_3"}},
		{"describe all match '_0'", fakeAPI, []api.MetricKey{"series_0"}},
		{"describe all match '_5'", fakeAPI, []api.MetricKey{}},
	} {
		a := assert.New(t).Contextf("query=%s", test.query)
		testCommand, err := parser.Parse(test.query)
		a.CheckError(err)

		a.EqString(testCommand.Name(), "describe all")
		fakeMulti := mocks.FakeTimeseriesStorageAPI{}
		rawResult, err := testCommand.Execute(command.ExecutionContext{
			TimeseriesStorageAPI: fakeMulti,
			MetricMetadataAPI:    test.metricmetadata,
			FetchLimit:           1000,
			Timeout:              0,
			Ctx:                  context.Background(),
		})
		a.CheckError(err)
		a.Eq(rawResult.Body, test.expected)
	}
}
