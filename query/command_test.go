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
	"github.com/square/metrics/testing_support/assert"
	"github.com/square/metrics/testing_support/mocks"
)

func TestCommand_Describe(t *testing.T) {
	fakeAPI := mocks.NewFakeMetricMetadataAPI()
	fakeAPI.MockMetric(api.TaggedMetric{"series_0", api.ParseTagSet("dc=west,env=production,host=a")})
	fakeAPI.MockMetric(api.TaggedMetric{"series_0", api.ParseTagSet("dc=west,env=staging,host=b")})
	fakeAPI.MockMetric(api.TaggedMetric{"series_0", api.ParseTagSet("dc=east,env=production,host=c")})
	fakeAPI.MockMetric(api.TaggedMetric{"series_0", api.ParseTagSet("dc=east,env=staging,host=d")})

	for _, test := range []struct {
		query          string
		metricmetadata api.MetricMetadataAPI
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
		command, err := Parse(test.query)
		if err != nil {
			a.Errorf("Unexpected error while parsing")
			continue
		}

		a.EqString(command.Name(), "describe")
		fakeTimeseriesStorage := mocks.FakeTimeseriesStorageAPI{}
		rawResult, err := command.Execute(ExecutionContext{
			MetricConverter:           &mocks.FakeGraphiteConverter{},
			TimeseriesStorageAPI:      fakeTimeseriesStorage,
			MetricMetadataAPI:         test.metricmetadata,
			FetchLimit:                1000,
			Timeout:                   0,
			OptimizationConfiguration: optimize.NewOptimizationConfiguration(),
		})
		a.CheckError(err)
		a.Eq(rawResult.Body, test.expected)
	}
}

func TestCommand_DescribeAll(t *testing.T) {
	fakeAPI := mocks.NewFakeMetricMetadataAPI()
	fakeAPI.MockMetric(api.TaggedMetric{"series_0", api.ParseTagSet("")})
	fakeAPI.MockMetric(api.TaggedMetric{"series_1", api.ParseTagSet("")})
	fakeAPI.MockMetric(api.TaggedMetric{"series_2", api.ParseTagSet("")})
	fakeAPI.MockMetric(api.TaggedMetric{"series_3", api.ParseTagSet("")})

	for _, test := range []struct {
		query          string
		metricmetadata api.MetricMetadataAPI
		expected       []api.MetricKey
	}{
		{"describe all", fakeAPI, []api.MetricKey{"series_0", "series_1", "series_2", "series_3"}},
		{"describe all match '_0'", fakeAPI, []api.MetricKey{"series_0"}},
		{"describe all match '_5'", fakeAPI, []api.MetricKey{}},
	} {
		a := assert.New(t).Contextf("query=%s", test.query)
		command, err := Parse(test.query)
		if err != nil {
			a.Errorf("Unexpected error while parsing")
			continue
		}

		a.EqString(command.Name(), "describe all")
		fakeMulti := mocks.FakeTimeseriesStorageAPI{}
		rawResult, err := command.Execute(ExecutionContext{
			TimeseriesStorageAPI:      fakeMulti,
			MetricMetadataAPI:         test.metricmetadata,
			FetchLimit:                1000,
			Timeout:                   0,
			OptimizationConfiguration: optimize.NewOptimizationConfiguration(),
		})
		a.CheckError(err)
		a.Eq(rawResult.Body, test.expected)
	}
}
