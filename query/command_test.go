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
	"github.com/square/metrics/api"
	"github.com/square/metrics/assert"
	"testing"
)

type fakeApiBackend struct {
}

type simpleFakeApi struct {
	api.API
}

func (f simpleFakeApi) GetAllTags(metricKey api.MetricKey) ([]api.TagSet, error) {
	return []api.TagSet{
		api.ParseTagSet("dc=west,env=production,host=a"),
		api.ParseTagSet("dc=west,env=staging,host=b"),
		api.ParseTagSet("dc=east,env=production,host=c"),
		api.ParseTagSet("dc=east,env=staging,host=d"),
	}, nil
}

func (f fakeApiBackend) FetchSeries(request api.FetchSeriesRequest) (*api.SeriesList, error) {
	metric := request.Metric
	timerange := request.Timerange
	if metric.MetricKey == "error_series" {
		return nil, errors.New("backend error")
	} else if metric.MetricKey == "series_1" {
		return &api.SeriesList{
			Series: []api.Timeseries{{
				[]float64{1, 2, 3, 4, 5},
				api.ParseTagSet("dc=west"),
			}},
			Timerange: timerange,
			Name:      string(metric.MetricKey),
		}, nil
	} else if metric.MetricKey == "series_2" {
		return &api.SeriesList{
			Series: []api.Timeseries{
				{
					[]float64{1, 2, 3, 4, 5},
					api.ParseTagSet("dc=west"),
				},
				{
					[]float64{3, 0, 3, 6, 2},
					api.ParseTagSet("dc=east"),
				},
			},
			Timerange: timerange,
			Name:      string(metric.MetricKey),
		}, nil
	}
	return nil, errors.New("internal error")
}

func TestCommand_Describe(t *testing.T) {
	var fakeBackend fakeApiBackend

	for _, test := range []struct {
		query   string
		backend api.Backend
		length  int // expected length of the result.
	}{
		{"describe m", fakeBackend, 4},
		{"describe m where dc='west'", fakeBackend, 2},
		{"describe m where dc='west' or env = 'production'", fakeBackend, 3},
		{"describe m where dc='west' or env = 'production' and doesnotexist = ''", fakeBackend, 2},
		{"describe m where env = 'production' and doesnotexist = '' or dc = 'west'", fakeBackend, 2},
		{"describe m where (dc='west' or env = 'production') and doesnotexist = ''", fakeBackend, 0}, // PARSER ERROR, currently.
	} {
		a := assert.New(t).Contextf("query=%s", test.query)
		rawCommand, err := Parse(test.query)
		if err != nil {
			a.Errorf("Unexpected error while parsing")
			continue
		}
		command := rawCommand.(*DescribeCommand)
		rawResult, _ := command.Execute(test.backend, simpleFakeApi{})
		parsedResult := rawResult.([]string)
		a.EqInt(len(parsedResult), test.length)
	}
}

func TestCommand_Select(t *testing.T) {
	epsilon := 1e-10
	var fakeBackend fakeApiBackend
	testTimerange, err := api.NewTimerange(0, 120, 30)
	if err != nil {
		t.Errorf("Invalid test timerange")
		return
	}

	for _, test := range []struct {
		query       string
		expectError bool
		expected    api.SeriesList
	}{
		{"select error_series from 0 to 120 resolution 30", true, api.SeriesList{}},
		{"select series_1 from 0 to 120 resolution 30", false, api.SeriesList{
			Series: []api.Timeseries{{
				[]float64{1, 2, 3, 4, 5},
				api.ParseTagSet("dc=west"),
			}},
			Timerange: testTimerange,
			Name:      "series_1",
		}},
		{"select series_1 + 1 from 0 to 120 resolution 30", false, api.SeriesList{
			Series: []api.Timeseries{{
				[]float64{2, 3, 4, 5, 6},
				api.ParseTagSet("dc=west"),
			}},
			Timerange: testTimerange,
			Name:      "",
		}},
		{"select series_1 * 2 from 0 to 120 resolution 30", false, api.SeriesList{
			Series: []api.Timeseries{{
				[]float64{2, 4, 6, 8, 10},
				api.ParseTagSet("dc=west"),
			}},
			Timerange: testTimerange,
			Name:      "",
		}},
		{"select aggregate.max(series_2) from 0 to 120 resolution 30", false, api.SeriesList{
			Series: []api.Timeseries{{
				[]float64{3, 2, 3, 6, 5},
				api.NewTagSet(),
			}},
			Timerange: testTimerange,
			Name:      "series_2",
		}},
	} {
		a := assert.New(t).Contextf("query=%s", test.query)
		expected := test.expected
		rawCommand, err := Parse(test.query)
		if err != nil {
			a.Errorf("Unexpected error while parsing")
			continue
		}
		command := rawCommand.(*SelectCommand)
		rawResult, err := command.Execute(fakeBackend, simpleFakeApi{})
		if err != nil {
			if !test.expectError {
				a.Errorf("Unexpected error while executing: %s", err.Error())
			}
		} else {
			casted := rawResult.([]value)
			actual, _ := casted[0].toSeriesList(api.Timerange{})
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
}
