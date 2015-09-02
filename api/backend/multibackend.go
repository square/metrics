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

type ParallelTimeseriesStorageWrapper struct {
	limit                int
	tickets              chan struct{}
	TimeseriesStorageAPI api.TimeseriesStorageAPI
}

func NewParallelMultiBackend(backend api.TimeseriesStorageAPI, limit int) *ParallelTimeseriesStorageWrapper {
	tickets := make(chan struct{}, limit)
	for i := 0; i < limit; i++ {
		tickets <- struct{}{}
	}
	return &ParallelTimeseriesStorageWrapper{
		limit:                limit,
		tickets:              tickets,
		TimeseriesStorageAPI: backend,
	}
}

// fetchLazy issues a goroutine to compute the timeseries once a fetchticket becomes available.
// It returns a channel to wait for the response to finish (the error).
// It stores the result of the function invokation in the series pointer it is given.
func (m *ParallelTimeseriesStorageWrapper) fetchLazy(cancellable api.Cancellable, result *api.Timeseries, work func() (api.Timeseries, error), channel chan error) {
	go func() {
		select {
		case ticket := <-m.tickets:
			series, err := work()
			// Put the ticket back (regardless of whether caller drops)
			m.tickets <- ticket
			// Store the result
			*result = series
			// Return the error (and sync up with the caller).
			channel <- err
		case <-cancellable.Done():
			channel <- api.TimeseriesStorageError{
				api.TaggedMetric{},
				api.FetchTimeoutError,
				"",
			}
		}
	}()
}

// fetchManyLazy abstracts upon fetchLazy so that looping over the resulting channels is not needed.
// It returns any overall error, as well as a slice of the resulting timeseries.
func (m *ParallelTimeseriesStorageWrapper) fetchManyLazy(cancellable api.Cancellable, works []func() (api.Timeseries, error)) ([]api.Timeseries, error) {
	results := make([]api.Timeseries, len(works))
	channel := make(chan error, len(works)) // Buffering the channel means the goroutines won't need to wait.
	for i := range results {
		m.fetchLazy(cancellable, &results[i], works[i], channel)
	}

	var err error = nil
	for _ = range works {
		select {
		case thisErr := <-channel:
			if thisErr != nil {
				err = thisErr
			}
		case <-cancellable.Done():
			return nil, api.TimeseriesStorageError{
				api.TaggedMetric{},
				api.FetchTimeoutError,
				"",
			}
		}
	}
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (m *ParallelTimeseriesStorageWrapper) FetchMultipleTimeseries(request api.FetchMultipleTimeseriesRequest) (api.SeriesList, error) {
	if request.Cancellable == nil {
		panic("The cancellable component of a FetchMultipleTimeseriesRequest cannot be nil")
	}
	works := make([]func() (api.Timeseries, error), len(request.Metrics))
	for i, metric := range request.Metrics {
		// Since we want to create a closure, we want to close over this particular metric,
		// rather than the variable itself (which is the same between iterations).
		// We accomplish this here:
		metric := metric
		works[i] = func() (api.Timeseries, error) {
			return m.TimeseriesStorageAPI.FetchSingleTimeseries(request.ToSingle(metric))
		}
	}

	resultSeries, err := m.fetchManyLazy(request.Cancellable, works)
	if err != nil {
		return api.SeriesList{}, err
	}

	return api.SeriesList{
		Series:    resultSeries,
		Timerange: request.Timerange,
	}, nil
}

type ProfilingMultiTimeseriesStorageWrapper struct {
	MultiBackend ParallelTimeseriesStorageWrapper
}

func (b ProfilingMultiTimeseriesStorageWrapper) FetchMultipleTimeseries(request api.FetchMultipleTimeseriesRequest) (api.SeriesList, error) {
	defer request.Profiler.Record("fetchMultipleSeries")()
	return b.MultiBackend.FetchMultipleTimeseries(request)
}
