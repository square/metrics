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
	"errors"
	"testing"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/api/backend"
	"github.com/square/metrics/assert"
	"github.com/square/metrics/function"
	"github.com/square/metrics/mocks"
)

var emptyGraphiteName = api.GraphiteMetric("")

type fakeApiBackend struct{}

func (f fakeApiBackend) FetchSingleSeries(request api.FetchSeriesRequest) (api.Timeseries, error) {
	metricMap := map[api.MetricKey][]api.Timeseries{
		"series_1": {{[]float64{1, 2, 3, 4, 5}, api.ParseTagSet("dc=west")}},
		"series_2": {{[]float64{1, 2, 3, 4, 5}, api.ParseTagSet("dc=west")}, {[]float64{3, 0, 3, 6, 2}, api.ParseTagSet("dc=east")}},
	}
	if string(request.Metric.MetricKey) == "series_timeout" {
		<-make(chan struct{}) // block forever
	}
	list, ok := metricMap[request.Metric.MetricKey]
	if !ok {
		return api.Timeseries{}, errors.New("internal error")
	}
	for _, series := range list {
		if request.Metric.TagSet.Serialize() == series.TagSet.Serialize() {
			// Cut the values based on the Timerange.
			values := make([]float64, request.Timerange.Slots())
			for i := range values {
				values[i] = series.Values[i+int(request.Timerange.Start())/30]
			}
			return api.Timeseries{values, series.TagSet}, nil
		}
	}
	return api.Timeseries{}, errors.New("internal error")
}

func TestCommand_Describe(t *testing.T) {
	fakeApi := mocks.NewFakeApi()
	fakeApi.AddPair(api.TaggedMetric{"series_0", api.ParseTagSet("dc=west,env=production,host=a")}, emptyGraphiteName)
	fakeApi.AddPair(api.TaggedMetric{"series_0", api.ParseTagSet("dc=west,env=staging,host=b")}, emptyGraphiteName)
	fakeApi.AddPair(api.TaggedMetric{"series_0", api.ParseTagSet("dc=east,env=production,host=c")}, emptyGraphiteName)
	fakeApi.AddPair(api.TaggedMetric{"series_0", api.ParseTagSet("dc=east,env=staging,host=d")}, emptyGraphiteName)

	for _, test := range []struct {
		query   string
		backend api.API
		length  int // expected length of the result.
	}{
		{"describe series_0", fakeApi, 4},
		{"describe`series_0`", fakeApi, 4},
		{"describe series_0 where dc='west'", fakeApi, 2},
		{"describe`series_0`where(dc='west')", fakeApi, 2},
		{"describe series_0 where dc='west' or env = 'production'", fakeApi, 3},
		{"describe series_0 where`dc`='west'or`env`='production'", fakeApi, 3},
		{"describe series_0 where dc='west' or env = 'production' and doesnotexist = ''", fakeApi, 2},
		{"describe series_0 where env = 'production' and doesnotexist = '' or dc = 'west'", fakeApi, 2},
		{"describe series_0 where (dc='west' or env = 'production') and doesnotexist = ''", fakeApi, 0},
		{"describe series_0 where(dc='west' or env = 'production')and`doesnotexist` = ''", fakeApi, 0},
	} {
		a := assert.New(t).Contextf("query=%s", test.query)
		command, err := Parse(test.query)
		if err != nil {
			a.Errorf("Unexpected error while parsing")
			continue
		}
		a.EqString(command.Name(), "describe")
		rawResult, _ := command.Execute(ExecutionContext{Backend: nil, API: test.backend, FetchLimit: 1000, Timeout: 0})
		parsedResult := rawResult.([]string)
		a.EqInt(len(parsedResult), test.length)
	}
}

func TestCommand_Select(t *testing.T) {
	epsilon := 1e-10
	fakeApi := mocks.NewFakeApi()
	fakeApi.AddPair(api.TaggedMetric{"series_1", api.ParseTagSet("dc=west")}, emptyGraphiteName)
	fakeApi.AddPair(api.TaggedMetric{"series_2", api.ParseTagSet("dc=east")}, emptyGraphiteName)
	fakeApi.AddPair(api.TaggedMetric{"series_2", api.ParseTagSet("dc=west")}, emptyGraphiteName)
	fakeApi.AddPair(api.TaggedMetric{"series_timeout", api.ParseTagSet("dc=west")}, emptyGraphiteName)
	var fakeBackend fakeApiBackend
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
		expected    api.SeriesList
	}{
		{"select does_not_exist from 0 to 120 resolution '30ms'", true, api.SeriesList{}},
		{"select series_1 from 0 to 120 resolution '30ms'", false, api.SeriesList{
			Series: []api.Timeseries{{
				[]float64{1, 2, 3, 4, 5},
				api.ParseTagSet("dc=west"),
			}},
			Timerange: testTimerange,
			Name:      "series_1",
		}},
		{"select series_timeout from 0 to 120 resolution '30ms'", true, api.SeriesList{}},
		{"select series_1 + 1 from 0 to 120 resolution 30ms", false, api.SeriesList{
			Series: []api.Timeseries{{
				[]float64{2, 3, 4, 5, 6},
				api.ParseTagSet("dc=west"),
			}},
			Timerange: testTimerange,
			Name:      "",
		}},
		{"select series_1 * 2 from 0 to 120 resolution '30ms'", false, api.SeriesList{
			Series: []api.Timeseries{{
				[]float64{2, 4, 6, 8, 10},
				api.ParseTagSet("dc=west"),
			}},
			Timerange: testTimerange,
			Name:      "",
		}},
		{"select aggregate.max(series_2) from 0 to 120 resolution 30ms", false, api.SeriesList{
			Series: []api.Timeseries{{
				[]float64{3, 2, 3, 6, 5},
				api.NewTagSet(),
			}},
			Timerange: testTimerange,
			Name:      "series_2",
		}},
		{"select (1 + series_2) | aggregate.max from 0 to 120 resolution 30ms", false, api.SeriesList{
			Series: []api.Timeseries{{
				[]float64{4, 3, 4, 7, 6},
				api.NewTagSet(),
			}},
			Timerange: testTimerange,
			Name:      "series_2",
		}},
		{"select series_1 from 0 to 60 resolution '30ms'", false, api.SeriesList{
			Series: []api.Timeseries{{
				[]float64{1, 2, 3},
				api.ParseTagSet("dc=west"),
			}},
			Timerange: earlyTimerange,
			Name:      "series_1",
		}},
		{"select transform.timeshift(series_1,'31ms') from 0 to 60 resolution 30ms", false, api.SeriesList{
			Series: []api.Timeseries{{
				[]float64{2, 3, 4},
				api.ParseTagSet("dc=west"),
			}},
			Timerange: earlyTimerange,
			Name:      "series_1",
		}},
		{"select transform.timeshift(series_1,'62ms') from 0 to 60 resolution '30ms'", false, api.SeriesList{
			Series: []api.Timeseries{{
				[]float64{3, 4, 5},
				api.ParseTagSet("dc=west"),
			}},
			Timerange: earlyTimerange,
			Name:      "series_1",
		}},
		{"select transform.timeshift(series_1,'29ms') from 0 to 60 resolution '30ms'", false, api.SeriesList{
			Series: []api.Timeseries{{
				[]float64{2, 3, 4},
				api.ParseTagSet("dc=west"),
			}},
			Timerange: earlyTimerange,
			Name:      "series_1",
		}},
		{"select transform.timeshift(series_1,'-31ms') from 60 to 120 resolution '30ms'", false, api.SeriesList{
			Series: []api.Timeseries{{
				[]float64{2, 3, 4},
				api.ParseTagSet("dc=west"),
			}},
			Timerange: lateTimerange,
			Name:      "series_1",
		}},
		{"select transform.timeshift(series_1,'-29ms') from 60 to 120 resolution '30ms'", false, api.SeriesList{
			Series: []api.Timeseries{{
				[]float64{2, 3, 4},
				api.ParseTagSet("dc=west"),
			}},
			Timerange: lateTimerange,
			Name:      "series_1",
		}},
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
			Backend:    backend.NewSequentialMultiBackend(fakeBackend),
			API:        fakeApi,
			FetchLimit: 1000,
			Timeout:    10 * time.Millisecond,
		})
		if err != nil {
			if !test.expectError {
				a.Errorf("Unexpected error while executing: %s", err.Error())
			}
		} else {
			casted := rawResult.([]function.Value)
			actual, _ := casted[0].ToSeriesList(api.Timerange{})
			a.EqInt(len(actual.Series), len(expected.Series))
			if len(actual.Series) == len(expected.Series) {
				for i := 0; i < len(expected.Series); i++ {
					a.Eq(actual.Series[i].TagSet, expected.Series[i].TagSet)
					actualLength := len(actual.Series[i].Values)
					expectedLength := len(actual.Series[i].Values)
					a.Eq(actualLength, expectedLength)
					if actualLength == expectedLength {
						for j := 0; j < actualLength; j++ {
							a.EqFloat(actual.Series[i].Values[j], expected.Series[i].Values[j], epsilon)
						}
					}
				}
			}
		}
	}

	// Test that the limit is correct
	command, err := Parse("select series_1, series_2 from 0 to 120 resolution '30ms'")
	if err != nil {
		t.Fatalf("Unexpected error while parsing")
		return
	}
	context := ExecutionContext{Backend: backend.NewSequentialMultiBackend(fakeBackend), API: fakeApi, FetchLimit: 3, Timeout: 0}
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
	command, err = Parse("select series2 from 0 to 120 resolution '30ms'")
	if err != nil {
		t.Fatalf("Unexpected error while parsing")
		return
	}
	_, err = command.Execute(context)
	if err != nil {
		t.Fatalf("expected success with limit = 2 but got %s", err.Error())
	}
}

func TestNaming(t *testing.T) {
	fakeApi := mocks.NewFakeApi()
	fakeBackend := backend.NewSequentialMultiBackend(fakeApiBackend{})
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
			query:    "select transform.moving_average(series_2, '2h') from 0 to 0",
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
		rawResult, err := command.Execute(ExecutionContext{Backend: fakeBackend, API: fakeApi, FetchLimit: 1000, Timeout: 0})
		if err != nil {
			t.Errorf("Unexpected error while execution: %s", err.Error())
			continue
		}
		seriesListList, ok := rawResult.([]function.Value)
		if !ok || len(seriesListList) != 1 {
			t.Errorf("expected query `%s` to produce []value; got %+v :: %T", test.query, rawResult, rawResult)
			continue
		}
		actual := api.SeriesList(seriesListList[0].(function.SeriesListValue)).Name
		if actual != test.expected {
			t.Errorf("Expected `%s` but got `%s` for query `%s`", test.expected, actual, test.query)
			continue
		}
	}

}

func TestTagDrop(t *testing.T) {
	fakeApi := mocks.NewFakeApi()
	fakeBackend := backend.NewSequentialMultiBackend(fakeApiBackend{})
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
		rawResult, err := command.Execute(ExecutionContext{Backend: fakeBackend, API: fakeApi, FetchLimit: 1000, Timeout: 0})
		if err != nil {
			t.Errorf("Unexpected error while execution: %s", err.Error())
			continue
		}
		seriesListList, ok := rawResult.([]function.Value)
		if !ok || len(seriesListList) != 1 {
			t.Errorf("expected query `%s` to produce []value; got %+v :: %T", test.query, rawResult, rawResult)
			continue
		}
		list, err := seriesListList[0].ToSeriesList(api.Timerange{})
		if err != nil {
			t.Fatal(err)
		}
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
