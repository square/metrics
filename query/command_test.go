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
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/optimize"
	"github.com/square/metrics/testing_support/assert"
	"github.com/square/metrics/testing_support/mocks"
	"github.com/square/metrics/util"
)

var emptyGraphiteName = util.GraphiteMetric("")

func TestCommand_Describe(t *testing.T) {
	fakeAPI := mocks.NewFakeMetricMetadataAPI()
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_0", api.ParseTagSet("dc=west,env=production,host=a")})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_0", api.ParseTagSet("dc=west,env=staging,host=b")})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_0", api.ParseTagSet("dc=east,env=production,host=c")})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_0", api.ParseTagSet("dc=east,env=staging,host=d")})

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
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_0", api.ParseTagSet("")})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_1", api.ParseTagSet("")})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_2", api.ParseTagSet("")})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_3", api.ParseTagSet("")})

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

func TestCommand_Select(t *testing.T) {
	fakeAPI := mocks.NewFakeMetricMetadataAPI()
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_1", api.ParseTagSet("dc=west")})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_2", api.ParseTagSet("dc=east")})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_2", api.ParseTagSet("dc=west")})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_3", api.ParseTagSet("dc=west")})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_3", api.ParseTagSet("dc=east")})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_3", api.ParseTagSet("dc=north")})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_timeout", api.ParseTagSet("dc=west")})
	var fakeBackend mocks.FakeTimeseriesStorageAPI
	testTimerange, err := api.NewTimerange(0, 120, 30)
	if err != nil {
		t.Errorf("Invalid test timerange")
		return
	}
	earlyTimerange, err := api.NewTimerange(0, 60, 30)
	if err != nil {
		t.Errorf("Invalid test timerange")
	}
	lateTimerange, err := api.NewTimerange(60, 120, 30)
	if err != nil {
		t.Errorf("Invalid test timerange")
	}
	for _, test := range []struct {
		query       string
		expectError bool
		expected    []api.SeriesList
	}{
		{"select does_not_exist from 0 to 120 resolution 30ms", true, []api.SeriesList{}},
		{"select series_1 from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{1, 2, 3, 4, 5},
				TagSet: api.ParseTagSet("dc=west"),
			}},
			Timerange: testTimerange,
			Name:      "series_1",
		}}},
		{"select series_timeout from 0 to 120 resolution 30ms", true, []api.SeriesList{}},
		{"select series_1 + 1 from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{2, 3, 4, 5, 6},
				TagSet: api.ParseTagSet("dc=west"),
			}},
			Timerange: testTimerange,
			Name:      "",
		}}},
		{"select series_1 * 2 from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{2, 4, 6, 8, 10},
				TagSet: api.ParseTagSet("dc=west"),
			}},
			Timerange: testTimerange,
			Name:      "",
		}}},
		{"select aggregate.max(series_2) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{3, 2, 3, 6, 5},
				TagSet: api.NewTagSet(),
			}},
			Timerange: testTimerange,
			Name:      "series_2",
		}}},
		{"select (1 + series_2) | aggregate.max from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{4, 3, 4, 7, 6},
				TagSet: api.NewTagSet(),
			}},
			Timerange: testTimerange,
			Name:      "series_2",
		}}},
		{"select series_1 from 0 to 60 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{1, 2, 3},
				TagSet: api.ParseTagSet("dc=west"),
			}},
			Timerange: earlyTimerange,
			Name:      "series_1",
		}}},
		{"select transform.timeshift(series_1,31ms) from 0 to 60 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{2, 3, 4},
				TagSet: api.ParseTagSet("dc=west"),
			}},
			Timerange: earlyTimerange,
			Name:      "series_1",
		}}},
		{"select transform.timeshift(series_1,62ms) from 0 to 60 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{3, 4, 5},
				TagSet: api.ParseTagSet("dc=west"),
			}},
			Timerange: earlyTimerange,
			Name:      "series_1",
		}}},
		{"select transform.timeshift(series_1,29ms) from 0 to 60 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{2, 3, 4},
				TagSet: api.ParseTagSet("dc=west"),
			}},
			Timerange: earlyTimerange,
			Name:      "series_1",
		}}},
		{"select transform.timeshift(series_1,-31ms) from 60 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{2, 3, 4},
				TagSet: api.ParseTagSet("dc=west"),
			}},
			Timerange: lateTimerange,
			Name:      "series_1",
		}}},
		{"select transform.timeshift(series_1,-29ms) from 60 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{2, 3, 4},
				TagSet: api.ParseTagSet("dc=west"),
			}},
			Timerange: lateTimerange,
			Name:      "series_1",
		}}},
		{"select series_3 from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.ParseTagSet("dc=west"),
				},
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.ParseTagSet("dc=east"),
				},
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.ParseTagSet("dc=north"),
				},
			},
		}}},
		{"select series_3 | filter.recent_highest_max(3, 30ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.ParseTagSet("dc=west"),
				},
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.ParseTagSet("dc=north"),
				},
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.ParseTagSet("dc=east"),
				},
			},
		}}},
		{"select series_3 | filter.recent_highest_max(2, 30ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.ParseTagSet("dc=west"),
				},
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.ParseTagSet("dc=north"),
				},
			},
		}}},
		{"select series_3 | filter.recent_highest_max(1, 30ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.ParseTagSet("dc=west"),
				},
			},
		}}},
		{"select series_3 | filter.recent_lowest_max(3, 30ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.ParseTagSet("dc=east"),
				},
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.ParseTagSet("dc=north"),
				},
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.ParseTagSet("dc=west"),
				},
			},
		}}},
		{"select series_3 | filter.recent_lowest_max(4, 30ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.ParseTagSet("dc=east"),
				},
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.ParseTagSet("dc=north"),
				},
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.ParseTagSet("dc=west"),
				},
			},
		}}},
		{"select series_3 | filter.recent_highest_max(70, 30ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.ParseTagSet("dc=west"),
				},
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.ParseTagSet("dc=north"),
				},
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.ParseTagSet("dc=east"),
				},
			},
		}}},
		{"select series_3 | filter.recent_lowest_max(2, 30ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.ParseTagSet("dc=east"),
				},
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.ParseTagSet("dc=north"),
				},
			},
		}}},
		{"select series_3 | filter.recent_lowest_max(1, 30ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.ParseTagSet("dc=east"),
				},
			},
		}}},
		{"select series_3 | filter.recent_highest_max(3, 3000ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.ParseTagSet("dc=east"),
				},
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.ParseTagSet("dc=west"),
				},
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.ParseTagSet("dc=north"),
				},
			},
		}}},
		{"select series_3 | filter.recent_highest_max(2, 3000ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.ParseTagSet("dc=east"),
				},
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.ParseTagSet("dc=west"),
				},
			},
		}}},
		{"select series_3 | filter.recent_highest_max(1, 3000ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.ParseTagSet("dc=east"),
				},
			},
		}}},
		{"select series_3 | filter.recent_lowest_max(3, 3000ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.ParseTagSet("dc=north"),
				},
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.ParseTagSet("dc=west"),
				},
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.ParseTagSet("dc=east"),
				},
			},
		}}},
		{"select series_3 | filter.recent_lowest_max(2, 3000ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.ParseTagSet("dc=north"),
				},
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.ParseTagSet("dc=west"),
				},
			},
		}}},
		{"select series_3 | filter.recent_lowest_max(1, 3000ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.ParseTagSet("dc=north"),
				},
			},
		}}},
		{"select series_1 from -1000d to now resolution 30s", true, []api.SeriesList{}},
	} {
		a := assert.New(t).Contextf("query=%s", test.query)
		expected := test.expected
		command, err := Parse(test.query)
		if err != nil {
			a.Errorf("Unexpected error while parsing")
			continue
		}
		a.EqString(command.Name(), "select")
		rawResult, err := command.Execute(ExecutionContext{
			TimeseriesStorageAPI:      fakeBackend,
			MetricMetadataAPI:         fakeAPI,
			FetchLimit:                1000,
			Timeout:                   100 * time.Millisecond,
			OptimizationConfiguration: optimize.NewOptimizationConfiguration(),
		})
		if test.expectError {
			if err == nil {
				t.Errorf("Expected error on %s but got no error; got value: %+v", test.query, rawResult.Body)
			}
		} else {
			a.CheckError(err)
			actual := rawResult.Body.([]QuerySeriesList)
			a.EqInt(len(actual), len(expected))
			if len(actual) == len(expected) {
				for i := range actual {
					list := actual[i]
					a.EqInt(len(list.Series), len(expected[i].Series))
					for j := range list.Series {
						a.Eq(list.Series[j].TagSet, expected[i].Series[j].TagSet)
						a.EqFloatArray(list.Series[j].Values, expected[i].Series[j].Values, 1e-4)
					}
				}
			}
		}
	}

	// Test that the limit is correct
	command, err := Parse("select series_1, series_2 from 0 to 120 resolution 30ms")
	if err != nil {
		t.Fatalf("Unexpected error while parsing")
		return
	}
	context := ExecutionContext{
		TimeseriesStorageAPI:      fakeBackend,
		MetricMetadataAPI:         fakeAPI,
		FetchLimit:                3,
		Timeout:                   0,
		OptimizationConfiguration: optimize.NewOptimizationConfiguration(),
	}
	_, err = command.Execute(context)
	if err != nil {
		t.Fatalf("expected success with limit 3 but got err = %s", err.Error())
		return
	}
	context.FetchLimit = 2
	_, err = command.Execute(context)
	if err == nil {
		t.Fatalf("expected failure with limit = 2")
		return
	}
	command, err = Parse("select series_2 from 0 to 120 resolution 30ms")
	if err != nil {
		t.Fatalf("Unexpected error while parsing")
		return
	}
	_, err = command.Execute(context)
	if err != nil {
		t.Fatalf("expected success with limit = 2 but got '%s'", err.Error())
	}
}

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
			t.Fatalf("Unexpected error while parsing")
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
	}
	for _, test := range tests {
		command, err := Parse(test.query)
		if err != nil {
			t.Fatalf("Unexpected error while parsing")
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
			t.Errorf("Expected `%s` but got `%s` for query `%s`", test.expected, actual, test.query)
			continue
		}
	}
}

func TestTag(t *testing.T) {
	fakeAPI := mocks.NewFakeMetricMetadataAPI()
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_1", api.ParseTagSet("dc=west,env=production")})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_1", api.ParseTagSet("dc=east,env=staging")})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_2", api.ParseTagSet("dc=west,env=production")})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_2", api.ParseTagSet("dc=east,env=staging")})

	fakeBackend := mocks.FakeTimeseriesStorageAPI{}
	tests := []struct {
		query    string
		expected api.SeriesList
	}{
		{
			query: "select series_2 | tag.drop('dc') from 0  to 120",
			expected: api.SeriesList{
				Series: []api.Timeseries{
					{
						Values: []float64{1, 2, 3, 4, 5},
						TagSet: api.TagSet{},
					},
					{
						Values: []float64{3, 0, 3, 6, 2},
						TagSet: api.TagSet{},
					},
				},
			},
		},
		{
			query: "select series_2 | tag.drop('none') from 0  to 120",
			expected: api.SeriesList{
				Series: []api.Timeseries{
					{
						Values: []float64{1, 2, 3, 4, 5},
						TagSet: api.TagSet{"dc": "west"},
					},
					{
						Values: []float64{3, 0, 3, 6, 2},
						TagSet: api.TagSet{"dc": "east"},
					},
				},
			},
		},
		{
			query: "select series_2 | tag.set('dc', 'north') from 0  to 120",
			expected: api.SeriesList{
				Series: []api.Timeseries{
					{
						Values: []float64{1, 2, 3, 4, 5},
						TagSet: api.TagSet{"dc": "north"},
					},
					{
						Values: []float64{3, 0, 3, 6, 2},
						TagSet: api.TagSet{"dc": "north"},
					},
				},
			},
		},
		{
			query: "select series_2 | tag.set('none', 'north') from 0  to 120",
			expected: api.SeriesList{
				Series: []api.Timeseries{
					{
						Values: []float64{1, 2, 3, 4, 5},
						TagSet: api.TagSet{"dc": "west", "none": "north"},
					},
					{
						Values: []float64{3, 0, 3, 6, 2},
						TagSet: api.TagSet{"dc": "east", "none": "north"},
					},
				},
			},
		},
	}
	for _, test := range tests {
		command, err := Parse(test.query)
		if err != nil {
			t.Fatalf("Unexpected error while parsing")
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
		list := seriesListList[0]
		a := assert.New(t)
		expectedSeries := test.expected.Series
		for i, series := range list.Series {
			a.EqFloatArray(series.Values, expectedSeries[i].Values, 1e-100)
			if !series.TagSet.Equals(expectedSeries[i].TagSet) {
				t.Errorf("expected tagset %+v but got %+v for series %d of query %s", expectedSeries[i].TagSet, series.TagSet, i, test.query)
			}
		}
	}
}
