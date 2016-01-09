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

type Ticket struct{}

type Tickets struct {
	tickets chan Ticket
}

func NewTickets(n int) Tickets {
	t := Tickets{
		tickets: make(chan Ticket, n),
	}
	for i := 0; i < n; i++ {
		t.tickets <- Ticket{}
	}
	return t
}
func (t Tickets) Wait() <-chan Ticket {
	return t.tickets
}
func (t Tickets) Done() {
	t.tickets <- Ticket{}
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
		result, err := p.FetchSingleTimeseries(request)
		resultChan <- result
		errChan <- err
	}()
	return resultChan, errChan
}

func (p ParallelTimeseriesStorageAPI) FetchSingleTimeseries(request api.FetchTimeseriesRequest) (api.Timeseries, error) {
	// We check that we're not performing too many concurrent requests, by waiting for a ticket.
	// Note: this could block for a while, depending on congestion.
	select {
	case <-p.tickets.Wait():
		// We have our ticket, so we can proceed.
		defer p.tickets.Done() // Put the ticket back when we're done.
	case <-request.Cancellable.Done():
		// We ran out of time, so don't bother performing the query at all.
		return api.Timeseries{}, api.TimeseriesStorageError{Code: api.FetchTimeoutError}
	}
	return p.FetchSingleTimeseries(request)
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
