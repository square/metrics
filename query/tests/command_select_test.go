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
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/query/command"
	"github.com/square/metrics/query/parser"
	"github.com/square/metrics/query/predicate"
	"github.com/square/metrics/testing_support/assert"
	"github.com/square/metrics/testing_support/mocks"

	"golang.org/x/net/context"
)

func TestCommand_Select(t *testing.T) {
	testTimerange, err := api.NewSnappedTimerange(0, 120, 30)
	if err != nil {
		t.Fatalf("Error creating timerange for test: %s", err.Error())
	}

	comboAPI := mocks.NewComboAPI(
		// timerange
		testTimerange,
		// series_1
		api.Timeseries{Values: []float64{1, 2, 3, 4, 5}, TagSet: api.TagSet{"metric": "series_1", "dc": "west"}},
		// series_2
		api.Timeseries{Values: []float64{1, 2, 3, 4, 5}, TagSet: api.TagSet{"metric": "series_2", "dc": "west"}},
		api.Timeseries{Values: []float64{3, 0, 3, 6, 2}, TagSet: api.TagSet{"metric": "series_2", "dc": "east"}},
		// series_3
		api.Timeseries{Values: []float64{1, 1, 1, 4, 4}, TagSet: api.TagSet{"metric": "series_3", "dc": "west"}},
		api.Timeseries{Values: []float64{5, 5, 5, 2, 2}, TagSet: api.TagSet{"metric": "series_3", "dc": "east"}},
		api.Timeseries{Values: []float64{3, 3, 3, 3, 3}, TagSet: api.TagSet{"metric": "series_3", "dc": "north"}},
	)
	for _, test := range []struct {
		query       string
		expectError bool
		expected    []api.SeriesList
	}{
		{"select does_not_exist from 0 to 120 resolution 30ms", true, []api.SeriesList{}},
		{"select series_1 from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{1, 2, 3, 4, 5},
				TagSet: api.TagSet{"dc": "west"},
			}},
		}}},
		{"select series_timeout from 0 to 120 resolution 30ms", true, []api.SeriesList{}},
		{"select series_1 + 1 from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{2, 3, 4, 5, 6},
				TagSet: api.TagSet{"dc": "west"},
			}},
		}}},
		{"select series_1 * 2 from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{2, 4, 6, 8, 10},
				TagSet: api.TagSet{"dc": "west"},
			}},
		}}},
		{"select aggregate.max(series_2) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{3, 2, 3, 6, 5},
				TagSet: api.NewTagSet(),
			}},
		}}},
		{"select (1 + series_2) | aggregate.max from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{4, 3, 4, 7, 6},
				TagSet: api.NewTagSet(),
			}},
		}}},
		{"select series_1 from 0 to 60 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{1, 2, 3},
				TagSet: api.TagSet{"dc": "west"},
			}},
		}}},
		{"select transform.timeshift(series_1,31ms) from 0 to 60 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{2, 3, 4},
				TagSet: api.TagSet{"dc": "west"},
			}},
		}}},
		{"select transform.timeshift(series_1,62ms) from 0 to 60 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{3, 4, 5},
				TagSet: api.TagSet{"dc": "west"},
			}},
		}}},
		{"select transform.timeshift(series_1,29ms) from 0 to 60 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{2, 3, 4},
				TagSet: api.TagSet{"dc": "west"},
			}},
		}}},
		{"select transform.timeshift(series_1,-31ms) from 60 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{2, 3, 4},
				TagSet: api.TagSet{"dc": "west"},
			}},
		}}},
		{"select transform.timeshift(series_1,-29ms) from 60 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{2, 3, 4},
				TagSet: api.TagSet{"dc": "west"},
			}},
		}}},
		{"select series_3 from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.TagSet{"dc": "west"},
				},
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.TagSet{"dc": "east"},
				},
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.TagSet{"dc": "north"},
				},
			},
		}}},
		{"select series_3 | filter.highest_max(3, 30ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.TagSet{"dc": "west"},
				},
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.TagSet{"dc": "north"},
				},
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.TagSet{"dc": "east"},
				},
			},
		}}},
		{"select series_3 | filter.highest_max(2, 30ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.TagSet{"dc": "west"},
				},
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.TagSet{"dc": "north"},
				},
			},
		}}},
		{"select series_3 | filter.highest_max(1, 30ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.TagSet{"dc": "west"},
				},
			},
		}}},
		{"select series_3 | filter.lowest_max(3, 30ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.TagSet{"dc": "east"},
				},
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.TagSet{"dc": "north"},
				},
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.TagSet{"dc": "west"},
				},
			},
		}}},
		{"select series_3 | filter.lowest_max(4, 30ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.TagSet{"dc": "east"},
				},
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.TagSet{"dc": "north"},
				},
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.TagSet{"dc": "west"},
				},
			},
		}}},
		{"select series_3 | filter.highest_max(70, 30ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.TagSet{"dc": "west"},
				},
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.TagSet{"dc": "north"},
				},
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.TagSet{"dc": "east"},
				},
			},
		}}},
		{"select series_3 | filter.lowest_max(2, 30ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.TagSet{"dc": "east"},
				},
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.TagSet{"dc": "north"},
				},
			},
		}}},
		{"select series_3 | filter.lowest_max(1, 30ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.TagSet{"dc": "east"},
				},
			},
		}}},
		{"select series_3 | filter.highest_max(3, 3000ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.TagSet{"dc": "east"},
				},
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.TagSet{"dc": "west"},
				},
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.TagSet{"dc": "north"},
				},
			},
		}}},
		{"select series_3 | filter.highest_max(2, 3000ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.TagSet{"dc": "east"},
				},
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.TagSet{"dc": "west"},
				},
			},
		}}},
		{"select series_3 | filter.highest_max(1, 3000ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.TagSet{"dc": "east"},
				},
			},
		}}},
		{"select series_3 | filter.lowest_max(3, 3000ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.TagSet{"dc": "north"},
				},
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.TagSet{"dc": "west"},
				},
				{
					Values: []float64{5, 5, 5, 2, 2},
					TagSet: api.TagSet{"dc": "east"},
				},
			},
		}}},
		{"select series_3 | filter.lowest_max(2, 3000ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.TagSet{"dc": "north"},
				},
				{
					Values: []float64{1, 1, 1, 4, 4},
					TagSet: api.TagSet{"dc": "west"},
				},
			},
		}}},
		{"select series_3 | filter.lowest_max(1, 3000ms) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{
				{
					Values: []float64{3, 3, 3, 3, 3},
					TagSet: api.TagSet{"dc": "north"},
				},
			},
		}}},
		{"select series_1 from -1000d to now resolution 30ms", true, []api.SeriesList{}},
	} {
		a := assert.New(t).Contextf("query=%s", test.query)
		expected := test.expected
		testCommand, err := parser.Parse(test.query)
		if err != nil {
			a.Errorf("Unexpected error while parsing")
			continue
		}
		a.EqString(testCommand.Name(), "select")
		rawResult, err := testCommand.Execute(command.ExecutionContext{
			TimeseriesStorageAPI: comboAPI,
			MetricMetadataAPI:    comboAPI,
			FetchLimit:           1000,
			Timeout:              100 * time.Millisecond,
			Ctx:                  context.Background(),
		})
		if test.expectError {
			if err == nil {
				t.Errorf("Expected error on %s but got no error; got value: %+v", test.query, rawResult.Body)
			}
		} else {
			a.CheckError(err)
			actual := rawResult.Body.([]command.QueryResult)
			a.EqInt(len(actual), len(expected))
			if len(actual) == len(expected) {
				for i := range actual {
					list := actual[i]
					if list.Type != "series" {
						t.Errorf("Should be given series, but was not.")
						continue
					}
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
	testCommand, err := parser.Parse("select series_1, series_2 from 0 to 120 resolution 30ms")
	if err != nil {
		t.Fatalf("Unexpected error while parsing")
		return
	}
	context := command.ExecutionContext{
		TimeseriesStorageAPI: comboAPI,
		MetricMetadataAPI:    comboAPI,
		FetchLimit:           3,
		Timeout:              0,
		Ctx:                  context.Background(),
	}
	_, err = testCommand.Execute(context)
	if err != nil {
		t.Fatalf("expected success with limit 3 but got err = %s", err.Error())
		return
	}
	context.FetchLimit = 2
	_, err = testCommand.Execute(context)
	if err == nil {
		t.Fatalf("expected failure with limit = 2")
		return
	}
	testCommand, err = parser.Parse("select series_2 from 0 to 120 resolution 30ms")
	if err != nil {
		t.Fatalf("Unexpected error while parsing")
		return
	}
	_, err = testCommand.Execute(context)
	if err != nil {
		t.Fatalf("expected success with limit = 2 but got '%s'", err.Error())
	}

	// Add an additional constraint and select.

	testCommand, err = parser.Parse("select series_2 from 0 to 120 resolution 30ms")
	if err != nil {
		t.Fatalf("Unexpected error while parsing")
	}
	context.AdditionalConstraints = predicate.ListMatcher{"dc", []string{"east"}}
	result, err := testCommand.Execute(context)
	if err != nil {
		t.Fatalf("expected success but got %s", err.Error())
	}
	queries := result.Body.([]command.QueryResult)[0].Series

	assert.New(t).Eq(queries, []api.Timeseries{
		{
			Values: []float64{3, 0, 3, 6, 2},
			TagSet: api.TagSet{"dc": "east"},
		},
	})
}

func TestTag(t *testing.T) {
	fakeAPI := mocks.NewFakeMetricMetadataAPI()
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_1", api.TagSet{"dc": "west", "env": "production"}})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_1", api.TagSet{"dc": "east", "env": "staging"}})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_2", api.TagSet{"dc": "west", "env": "production"}})
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series_2", api.TagSet{"dc": "east", "env": "staging"}})

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
		testCommand, err := parser.Parse(test.query)
		if err != nil {
			t.Fatalf("Unexpected error while parsing")
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
			Ctx:                  context.Background(),
		})
		if err != nil {
			t.Errorf("Unexpected error while execution: %s", err.Error())
			continue
		}
		seriesListList, ok := rawResult.Body.([]command.QueryResult)
		if !ok || len(seriesListList) != 1 || seriesListList[0].Type != "series" {
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
