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

package tests

import (
	"fmt"
	"math"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/square/metrics/api"
	"github.com/square/metrics/query/command"
	"github.com/square/metrics/query/parser"
	"github.com/square/metrics/testing_support/mocks"
	"github.com/square/metrics/timeseries"
)

var fixedNow = time.Now().Round(30 * time.Second)
var fullResolutionCutoff = fixedNow.Add(-24 * time.Hour)

type testResolutionStorage struct{}

func (t testResolutionStorage) ChooseResolution(requested api.Timerange, lowerBound time.Duration) (time.Duration, error) {
	if requested.Start().Before(fullResolutionCutoff) {
		return 5 * time.Minute, nil
	}
	return 30 * time.Second, nil
}
func (t testResolutionStorage) FetchSingleTimeseries(request timeseries.FetchRequest) (api.Timeseries, error) {
	if request.Timerange.Resolution() == 30*time.Second && request.Timerange.Start().Before(fullResolutionCutoff) {
		return api.Timeseries{}, fmt.Errorf("querying 30s resolution data prior to 24h ago (over by %v)", fullResolutionCutoff.Sub(request.Timerange.Start()))
	}
	return api.Timeseries{
		Values: make([]float64, request.Timerange.Slots()),
	}, nil
}
func (t testResolutionStorage) FetchMultipleTimeseries(request timeseries.FetchMultipleRequest) (api.SeriesList, error) {
	requests := request.ToSingle()
	series := make([]api.Timeseries, len(requests))
	for i := range requests {
		result, err := t.FetchSingleTimeseries(requests[i])
		if err != nil {
			return api.SeriesList{}, err
		}
		series[i] = result
	}
	return api.SeriesList{
		Series: series,
	}, nil
}

func (t testResolutionStorage) CheckHealthy() error {
	return nil
}

func relative(format string, durations ...time.Duration) string {
	args := make([]interface{}, len(durations))
	for i := range durations {
		args[i] = fixedNow.Add(durations[i]).Format(time.UnixDate)
	}
	return fmt.Sprintf(format, args...)
}

func TestResolutionEdge(t *testing.T) {
	queries := map[string]time.Duration{
		relative(`select foo from '%s' to '%s'`, -24*time.Hour, 0):                            30 * time.Second,
		relative(`select foo from '%s' to '%s'`, -24*time.Hour-time.Minute, 0):                5 * time.Minute,
		`select foo | transform.timeshift(-5m) from -1d to now`:                               5 * time.Minute,
		`select bar + transform.timeshift(foo, -5m) from -1d to now`:                          5 * time.Minute,
		`select foo | transform.moving_average(5m) from -1d to now`:                           5 * time.Minute,
		`select foo | transform.moving_average(5m) | transform.timeshift(5m) from -1d to now`: 30 * time.Second,
		`select foo | forecast.linear(5m) from -1d to now`:                                    5 * time.Minute,
		`select bar, foo | forecast.linear(5m) from -1d to now`:                               5 * time.Minute,
		`select bar + foo, foo | forecast.linear(5m) from -1d to now`:                         5 * time.Minute,
		`select foo | forecast.linear(5m) from -23h to now`:                                   30 * time.Second,
	}
	timerange, err := api.NewSnappedTimerange(300000000, 300000000, 30000)
	if err != nil {
		t.Fatalf("Error creating test timerange: %s", err.Error())
	}
	combo := mocks.NewComboAPI(
		timerange,
		api.Timeseries{TagSet: api.TagSet{"metric": "foo"}, Values: []float64{math.NaN()}},
		api.Timeseries{TagSet: api.TagSet{"metric": "bar"}, Values: []float64{math.NaN()}},
	)
	for query, expectedResolution := range queries {
		parsed, err := parser.Parse(query)
		if err != nil {
			t.Errorf("parsing error for query %q: %s", query, err.Error())
			continue
		}
		result, err := parsed.Execute(command.ExecutionContext{
			TimeseriesStorageAPI: testResolutionStorage{},
			MetricMetadataAPI:    combo,
			FetchLimit:           1000,
			SlotLimit:            1000000, // want this to not be a concern
			Ctx:                  context.Background(),
		})
		if err != nil {
			t.Errorf("unexpected error executing query %q: %s", query, err.Error())
			continue
		}
		if chosenResolution := result.Metadata["resolution"].(time.Duration); chosenResolution != expectedResolution {
			t.Errorf("expected query %q to use resolution %v but got %v", query, expectedResolution, chosenResolution)
		}
	}
}
