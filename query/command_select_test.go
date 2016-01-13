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
)

func TestCommand_Select(t *testing.T) {
	fakeConverter, fakeAPI := mocks.NewFakeGraphiteConverter([]api.TaggedMetric{
		{"series_1", api.TagSet{"dc": "west"}},

		{"series_2", api.TagSet{"dc": "east"}},
		{"series_2", api.TagSet{"dc": "west"}},

		{"series_3", api.TagSet{"dc": "west"}},
		{"series_3", api.TagSet{"dc": "east"}},
		{"series_3", api.TagSet{"dc": "north"}},

		{"series_timeout", api.TagSet{"dc": "west"}},
	})

	fakeBackend := &mocks.FakeTimeseriesStorageAPI{
		MetricMap: map[string][]float64{
			"series_1.dc.west": []float64{1, 2, 3, 4, 5},

			"series_2.dc.west": []float64{1, 2, 3, 4, 5},
			"series_2.dc.east": []float64{30, 0, 3, 6, 2},

			"series_3.dc.west":  []float64{1, 1, 1, 4, 4},
			"series_3.dc.east":  []float64{5, 5, 5, 2, 2},
			"series_3.dc.north": []float64{3, 3, 3, 3, 3},
		},
	}
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
		}}},
		{"select series_timeout from 0 to 120 resolution 30ms", true, []api.SeriesList{}},
		{"select series_1 + 1 from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{2, 3, 4, 5, 6},
				TagSet: api.ParseTagSet("dc=west"),
			}},
			Timerange: testTimerange,
		}}},
		{"select series_1 * 2 from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{2, 4, 6, 8, 10},
				TagSet: api.ParseTagSet("dc=west"),
			}},
			Timerange: testTimerange,
		}}},
		{"select aggregate.max(series_2) from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{30, 2, 3, 6, 5},
				TagSet: api.NewTagSet(),
			}},
			Timerange: testTimerange,
		}}},
		{"select (1 + series_2) | aggregate.max from 0 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{31, 3, 4, 7, 6},
				TagSet: api.NewTagSet(),
			}},
			Timerange: testTimerange,
		}}},
		{"select series_1 from 0 to 60 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{1, 2, 3},
				TagSet: api.ParseTagSet("dc=west"),
			}},
			Timerange: earlyTimerange,
		}}},
		{"select transform.timeshift(series_1,31ms) from 0 to 60 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{2, 3, 4},
				TagSet: api.ParseTagSet("dc=west"),
			}},
			Timerange: earlyTimerange,
		}}},
		{"select transform.timeshift(series_1,62ms) from 0 to 60 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{3, 4, 5},
				TagSet: api.ParseTagSet("dc=west"),
			}},
			Timerange: earlyTimerange,
		}}},
		{"select transform.timeshift(series_1,29ms) from 0 to 60 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{2, 3, 4},
				TagSet: api.ParseTagSet("dc=west"),
			}},
			Timerange: earlyTimerange,
		}}},
		{"select transform.timeshift(series_1,-31ms) from 60 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{2, 3, 4},
				TagSet: api.ParseTagSet("dc=west"),
			}},
			Timerange: lateTimerange,
		}}},
		{"select transform.timeshift(series_1,-29ms) from 60 to 120 resolution 30ms", false, []api.SeriesList{{
			Series: []api.Timeseries{{
				Values: []float64{2, 3, 4},
				TagSet: api.ParseTagSet("dc=west"),
			}},
			Timerange: lateTimerange,
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
			MetricConverter:           fakeConverter,
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
						a.Contextf("query: %s", test.query).Eq(list.Series[j].TagSet, expected[i].Series[j].TagSet)
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
		MetricConverter:           fakeConverter,
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

func TestTag(t *testing.T) {

	fakeConverter, fakeAPI := mocks.NewFakeGraphiteConverter([]api.TaggedMetric{
		{"series_2", api.ParseTagSet("dc=west,env=production")},
		{"series_2", api.ParseTagSet("dc=east,env=staging")},
	})

	fakeBackend := &mocks.FakeTimeseriesStorageAPI{
		MetricMap: map[string][]float64{
			"series_2.dc.west.env.production": []float64{1, 2, 3, 4, 5},
			"series_2.dc.east.env.staging":    []float64{3, 0, 3, 6, 2},
		},
	}

	tests := []struct {
		query    string
		expected api.SeriesList
	}{
		{
			query: "select series_2 | tag.drop('dc') from 0  to 120 resolution 30ms",
			expected: api.SeriesList{
				Series: []api.Timeseries{
					{
						Values: []float64{1, 2, 3, 4, 5},
						TagSet: api.TagSet{"env": "production"},
					},
					{
						Values: []float64{3, 0, 3, 6, 2},
						TagSet: api.TagSet{"env": "staging"},
					},
				},
			},
		},
		{
			query: "select series_2 | tag.drop('none') from 0  to 120 resolution 30ms",
			expected: api.SeriesList{
				Series: []api.Timeseries{
					{
						Values: []float64{1, 2, 3, 4, 5},
						TagSet: api.TagSet{"dc": "west", "env": "production"},
					},
					{
						Values: []float64{3, 0, 3, 6, 2},
						TagSet: api.TagSet{"dc": "east", "env": "staging"},
					},
				},
			},
		},
		{
			query: "select series_2 | tag.set('dc', 'north') from 0  to 120 resolution 30ms",
			expected: api.SeriesList{
				Series: []api.Timeseries{
					{
						Values: []float64{1, 2, 3, 4, 5},
						TagSet: api.TagSet{"dc": "north", "env": "production"},
					},
					{
						Values: []float64{3, 0, 3, 6, 2},
						TagSet: api.TagSet{"dc": "north", "env": "staging"},
					},
				},
			},
		},
		{
			query: "select series_2 | tag.set('none', 'north') from 0  to 120 resolution 30ms",
			expected: api.SeriesList{
				Series: []api.Timeseries{
					{
						Values: []float64{1, 2, 3, 4, 5},
						TagSet: api.TagSet{"dc": "west", "none": "north", "env": "production"},
					},
					{
						Values: []float64{3, 0, 3, 6, 2},
						TagSet: api.TagSet{"dc": "east", "none": "north", "env": "staging"},
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
			MetricConverter:           fakeConverter,
			TimeseriesStorageAPI:      fakeBackend,
			MetricMetadataAPI:         fakeAPI,
			FetchLimit:                1000,
			Timeout:                   0,
			OptimizationConfiguration: optimize.NewOptimizationConfiguration(),
		})
		if err != nil {
			t.Errorf("Unexpected error while exucting query %q: %s", test.query, err.Error())
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
