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

func (p ParallelTimeseriesStorageAPI) FetchMultipleTimeseries(request api.FetchMultipleTimeseriesRequest) (api.SeriesList, error) {
	defer request.Profiler.Record("parallel FetchMultipleTimeseries")()
	if request.Cancellable == nil {
		panic("request.Cancellable cannot be nil")
	}

	count := len(request.Metrics)

	total := make(chan struct{}, count) // synchronizes the individual tasks

	errChan := make(chan error, count) // possible collection of individual errors

	singleRequests := request.ToSingle()

	answers := make([]api.Timeseries, count) // Where the resulting Timeseries are placed.

	for i := range singleRequests {
		go func(i int) {
			select {
			// TODO: add a way to stop early (e.g. in event of error)
			case <-p.tickets.Wait():
				// Congestion is minimal, so proceed,
				// but free up our spot when we're done:
				defer p.tickets.Done()
			case <-request.Cancellable.Done():
				// Don't bother doing anything at all, since the timeout has occurred,
				// and our answer would be ignored anyway.
				return
			}
			// Ask the TimeseriesStorageAPI about my request, #i.
			answer, err := p.TimeseriesStorageAPI.FetchSingleTimeseries(singleRequests[i])
			if err != nil {
				// Errors go in the error channel.
				errChan <- err
				return
			}
			answers[i] = answer // Store my answer
			total <- struct{}{} // Let the synchronizer know that I'm done
		}(i)
	}
	waiting := count
	for {
		select {
		case <-total:
			waiting--
			if waiting != 0 {
				// All done!
				return api.SeriesList{
					Series:    answers,
					Timerange: request.Timerange,
				}, nil
			}
			// Still waiting on work
		case err := <-errChan: // One of the fetches produced an error:
			return api.SeriesList{}, err
		case <-request.Cancellable.Done(): // Timeout
			return api.SeriesList{}, api.TimeseriesStorageError{Code: api.FetchTimeoutError}
		}
	}
}
