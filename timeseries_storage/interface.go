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

package timeseries_storage

import (
	"fmt"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/inspect"
)

type TimeseriesStorageAPI interface {
	ChooseResolution(requested api.Timerange, smallestResolution time.Duration) time.Duration
	FetchSingleTimeseries(request FetchTimeseriesRequest) (api.Timeseries, error)
	FetchMultipleTimeseries(request FetchMultipleTimeseriesRequest) (api.SeriesList, error)
}

type RequestDetails struct {
	SampleMethod          SampleMethod  // up/downsampling behavior.
	Timerange             api.Timerange // time range to fetch data from.
	Cancellable           api.Cancellable
	Profiler              *inspect.Profiler
	UserSpecifiableConfig UserSpecifiableConfig
}

type FetchTimeseriesRequest struct {
	Metric api.TaggedMetric // metric to fetch.
	RequestDetails
}

type FetchMultipleTimeseriesRequest struct {
	Metrics []api.TaggedMetric
	RequestDetails
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
	Metric  api.TaggedMetric
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
	formatted := fmt.Sprintf(message, string(err.Metric.MetricKey), err.Metric.TagSet)
	if err.Message != "" {
		formatted = formatted + " - " + err.Message
	}
	return formatted
}

func (err TimeseriesStorageError) TokenName() string {
	return string(err.Metric.MetricKey)
}

// ToSingle very simply decompose the FetchMultipleTimeseriesRequest into single
// fetch requests (for now).
func (r FetchMultipleTimeseriesRequest) ToSingle() []FetchTimeseriesRequest {
	fetchSingleRequests := make([]FetchTimeseriesRequest, len(r.Metrics))
	for i, metric := range r.Metrics {
		fetchSingleRequests[i] = FetchTimeseriesRequest{
			Metric:         metric,
			RequestDetails: r.RequestDetails,
		}
	}
	return fetchSingleRequests
}
