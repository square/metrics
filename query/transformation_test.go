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

// TODO - remove this file in future.
// this tests moving average which is a very special logic.
package query

import (
	"math"
	"testing"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
	"github.com/square/metrics/function/registry"
	"github.com/square/metrics/optimize"
	"github.com/square/metrics/testing_support/mocks"
)

type movingAverageBackend struct{ mocks.FakeTimeseriesStorageAPI }

func (b movingAverageBackend) FetchSingleTimeseries(r api.FetchTimeseriesRequest) (api.Timeseries, error) {
	t := r.Timerange
	values := []float64{9, 2, 1, 6, 4, 5}
	startIndex := t.Start()/100 - 10
	result := make([]float64, t.Slots())
	for i := range result {
		result[i] = values[i+int(startIndex)]
	}
	return api.Timeseries{Values: values, TagSet: api.NewTagSet()}, nil
}

func (b movingAverageBackend) FetchMultipleTimeseries(r api.FetchMultipleTimeseriesRequest) (api.SeriesList, error) {
	timeseries := make([]api.Timeseries, 0)
	singleRequests := r.ToSingle()
	for _, request := range singleRequests {
		series, _ := b.FetchSingleTimeseries(request)
		timeseries = append(timeseries, series)
	}
	return api.SeriesList{
		Series: timeseries,
	}, nil
}

func TestMovingAverage(t *testing.T) {
	fakeAPI := mocks.NewFakeMetricMetadataAPI()
	fakeAPI.AddPairWithoutGraphite(api.TaggedMetric{"series", api.NewTagSet()})

	fakeBackend := movingAverageBackend{}
	timerange, err := api.NewTimerange(1200, 1500, 100)
	if err != nil {
		t.Fatalf(err.Error())
	}

	expression := &functionExpression{
		functionName: "transform.moving_average",
		groupBy:      []string{},
		arguments: []function.Expression{
			&metricFetchExpression{"series", api.TruePredicate},
			durationExpression{"300ms", 300 * time.Millisecond},
		},
	}

	backend := fakeBackend
	result, err := evaluateToSeriesList(expression,
		&function.EvaluationContext{
			MetricMetadataAPI:         fakeAPI,
			TimeseriesStorageAPI:      backend,
			Timerange:                 timerange,
			SampleMethod:              api.SampleMean,
			FetchLimit:                function.NewFetchCounter(1000),
			Registry:                  registry.Default(),
			Cancellable:               api.NewCancellable(),
			OptimizationConfiguration: optimize.NewOptimizationConfiguration(),
		})
	if err != nil {
		t.Errorf(err.Error())
	}

	expected := []float64{4, 3, 11.0 / 3, 5}

	if len(result.Series) != 1 {
		t.Fatalf("expected exactly 1 returned series")
	}
	if len(result.Series[0].Values) != len(expected) {
		t.Fatalf("expected exactly %d values in returned series, but got %d", len(expected), len(result.Series[0].Values))
	}
	const eps = 1e-7
	for i := range expected {
		if math.Abs(result.Series[0].Values[i]-expected[i]) > eps {
			t.Fatalf("expected %+v but got %+v", expected, result.Series[0].Values)
		}
	}
}
