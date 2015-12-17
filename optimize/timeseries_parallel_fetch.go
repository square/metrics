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

package optimize

import "github.com/square/metrics/api"

type Tickets struct {
	tickets chan struct{}
	size    int
}

func NewTickets(n int) Tickets {
	t := Tickets{
		tickets: make(chan struct{}, n),
		size:    n,
	}
	for i := 0; i < n; i++ {
		t.tickets <- struct{}{}
	}
	return t
}
func (t Tickets) Wait() <-chan struct{} {
	return t.tickets
}
func (t Tickets) Done() {
	t.tickets <- struct{}{}
}
func (t Tickets) Size() int {
	return t.size
}

type ParallelTimeseriesStorageAPI struct {
	api.TimeseriesStorageAPI
	tickets Tickets
}

func NewParallelTimeseriesStorageAPI(concurrentRequests int, storage api.TimeseriesStorageAPI) api.TimeseriesStorageAPI {
	return ParallelTimeseriesStorageAPI{
		TimeseriesStorageAPI: storage,
		tickets:              NewTickets(concurrentRequests),
	}
}

func (p ParallelTimeseriesStorageAPI) spawnSingleRequest(request api.FetchTimeseriesRequest) (chan api.Timeseries, chan error) {
	errChan := make(chan error, 1) // Capacity prevents deadlock and then garbage
	resultChan := make(chan api.Timeseries, 1)
	go func() {
		select {
		case <-p.tickets.Wait():
			// Congestion is minimal, so proceed,
			// but free up our spot when we're done:
			defer p.tickets.Done()
		case <-request.Cancellable.Done():
			// Don't bother doing anything at all, since the timeout has occurred,
			// and our answer would be ignored anyway.
			return
		}
		// Ask the TimeseriesStorageAPI about my request.
		if result, err := p.TimeseriesStorageAPI.FetchSingleTimeseries(request); err != nil {
			errChan <- err
		} else {
			resultChan <- result
		}
	}()
	return resultChan, errChan
}

func (p ParallelTimeseriesStorageAPI) FetchMultipleTimeseries(request api.FetchMultipleTimeseriesRequest) (api.SeriesList, error) {
	defer request.Profiler.Record("parallel FetchMultipleTimeseries")()
	if request.Cancellable == nil {
		panic("request.Cancellable cannot be nil")
	}

	singleRequests := request.ToSingle()
	count := len(singleRequests)

	answerChannels := make([]chan api.Timeseries, count)
	errChannels := make([]chan error, count)
	for i := range singleRequests {
		answerChannels[i], errChannels[i] = p.spawnSingleRequest(singleRequests[i])
	}

	answers := make([]api.Timeseries, count)

	for i := range answers {
		select {
		case <-request.Cancellable.Done():
			// Runs out of time
			return api.SeriesList{}, api.TimeseriesStorageError{Code: api.FetchTimeoutError}
		case err := <-errChannels[i]:
			// Got an error
			return api.SeriesList{}, err
		case answers[i] = <-answerChannels[i]:
			// Fill this answer
		}
	}

	return api.SeriesList{
		Series:    answers,
		Timerange: request.Timerange,
	}, nil
}
