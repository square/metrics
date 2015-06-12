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

package backend

import (
	"github.com/square/metrics/api"
)

type sequentialMultiBackend struct {
	Backend api.Backend
}

func NewSequentialMultiBackend(backend api.Backend) api.MultiBackend {
	return &sequentialMultiBackend{Backend: backend}
}

func (m *sequentialMultiBackend) FetchMultipleSeries(metrics []api.TaggedMetric, sampleMethod api.SampleMethod, timerange api.Timerange, myAPI api.API) (api.SeriesList, error) {

	series := make([]api.Timeseries, len(metrics))
	var err error = nil
	for i, metric := range metrics {
		series[i], err = m.Backend.FetchSingleSeries(api.FetchSeriesRequest{
			Metric:       metric,
			SampleMethod: sampleMethod,
			Timerange:    timerange,
			Api:          myAPI,
		})
		if err != nil {
			return api.SeriesList{}, err
		}
	}

	return api.SeriesList{
		Series:    series,
		Timerange: timerange,
	}, nil
}

type parallelMultiBackend struct {
	limit   int
	tickets chan struct{}
	Backend api.Backend
}

func NewParallelMultiBackend(backend api.Backend, limit int) api.MultiBackend {
	tickets := make(chan struct{}, limit)
	for i := 0; i < limit; i++ {
		tickets <- struct{}{}
	}
	return &parallelMultiBackend{
		limit:   limit,
		tickets: tickets,
		Backend: backend,
	}
}

// fetchLazy issues a goroutine to compute the timeseries once a fetchticket becomes available.
// It returns a channel to wait for the response to finish (the error).
// It stores the result of the function invokation in the series pointer it is given.
func (m *parallelMultiBackend) fetchLazy(result *api.Timeseries, fun func() (api.Timeseries, error), channel chan error) {
	go func() {
		ticket := <-m.tickets
		series, err := fun()
		// Put the ticket back (regardless of whether caller drops)
		m.tickets <- ticket
		// Store the result
		*result = series
		// Return the error (and sync up with the caller).
		channel <- err
	}()
}

// fetchManyLazy abstracts upon fetchLazy so that looping over the resulting channels is not needed.
// It returns any overall error, as well as a slice of the resulting timeseries.
func (m *parallelMultiBackend) fetchManyLazy(funs []func() (api.Timeseries, error)) ([]api.Timeseries, error) {
	results := make([]api.Timeseries, len(funs))
	channel := make(chan error, len(funs)) // Buffering the channel means the goroutines won't need to wait.
	for i := range results {
		m.fetchLazy(&results[i], funs[i], channel)
	}
	var err error = nil
	for _ = range funs {
		thisErr := <-channel
		if thisErr != nil {
			err = thisErr
		}
	}
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (m *parallelMultiBackend) FetchMultipleSeries(metrics []api.TaggedMetric, sampleMethod api.SampleMethod, timerange api.Timerange, myAPI api.API) (api.SeriesList, error) {

	funs := make([]func() (api.Timeseries, error), len(metrics))
	for i, metric := range metrics {
		// Since we want to create a closure, we want to close over this particular metric,
		// rather than the variable itself (which is the same between iterations).
		// We accomplish this here:
		metric := metric
		funs[i] = func() (api.Timeseries, error) {
			return m.Backend.FetchSingleSeries(api.FetchSeriesRequest{
				Metric:       metric,
				SampleMethod: sampleMethod,
				Timerange:    timerange,
				Api:          myAPI,
			})
		}
	}

	resultSeries, err := m.fetchManyLazy(funs)
	if err != nil {
		return api.SeriesList{}, err
	}

	return api.SeriesList{
		Series:    resultSeries,
		Timerange: timerange,
	}, nil
}
