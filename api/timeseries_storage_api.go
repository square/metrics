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

package api

import (
	"fmt"
	"time"

	"github.com/square/metrics/inspect"
)

type TimeseriesStorageAPI interface {
	ChooseResolution(requested Timerange, smallestResolution time.Duration) time.Duration
	FetchSingleTimeseries(request FetchTimeseriesRequest) ([]float64, error)
	FetchMultipleTimeseries(request FetchMultipleTimeseriesRequest) ([][]float64, error)
}

type FetchTimeseriesRequest struct {
	Metric                []byte       // metric to fetch.
	SampleMethod          SampleMethod // up/downsampling behavior.
	Timerange             Timerange    // time range to fetch data from.
	Cancellable           Cancellable
	Profiler              *inspect.Profiler
	UserSpecifiableConfig UserSpecifiableConfig
}

type FetchMultipleTimeseriesRequest struct {
	Metrics               [][]byte
	SampleMethod          SampleMethod
	Timerange             Timerange
	Cancellable           Cancellable
	Profiler              *inspect.Profiler
	UserSpecifiableConfig UserSpecifiableConfig
}

type UserSpecifiableConfig struct {
	IncludeRawData bool
}

type TimeseriesStorageErrorCode int

const (
	FetchTimeoutError  TimeseriesStorageErrorCode = iota + 1 // error while fetching - timeout.
	FetchIOError                                             // error while fetching - general IO.
	InvalidSeriesError                                       // the given series is not well-defined.
	LimitError                                               // the fetch limit is reached.
	Unsupported                                              // the given fetch operation is unsupported by the backend.
)

type TimeseriesStorageError struct {
	Metric  []byte
	Code    TimeseriesStorageErrorCode
	Message string
}

func (err TimeseriesStorageError) Error() string {
	message := "[%s %+v] unknown error"
	switch err.Code {
	case FetchTimeoutError:
		message = "[%s %+v] timeout"
	case InvalidSeriesError:
		message = "[%s %+v] invalid series"
	case LimitError:
		message = "[%s %+v] limit reached"
	case Unsupported:
		message = "[%s %+v] unsupported operation"
	}
	formatted := fmt.Sprintf(message, string(err.Metric))
	if err.Message != "" {
		formatted = formatted + " - " + err.Message
	}
	return formatted
}

func (err TimeseriesStorageError) TokenName() string {
	return string(err.Metric)
}

// ToSingle very simply decompose the FetchMultipleTimeseriesRequest into single
// fetch requests (for now).
func (r FetchMultipleTimeseriesRequest) ToSingle() []FetchTimeseriesRequest {
	fetchSingleRequests := make([]FetchTimeseriesRequest, len(r.Metrics))
	for i, metric := range r.Metrics {
		fetchSingleRequests[i] = FetchTimeseriesRequest{
			Metric:                metric,
			Cancellable:           r.Cancellable,
			SampleMethod:          r.SampleMethod,
			Timerange:             r.Timerange,
			Profiler:              r.Profiler,
			UserSpecifiableConfig: r.UserSpecifiableConfig,
		}
	}
	return fetchSingleRequests
}
