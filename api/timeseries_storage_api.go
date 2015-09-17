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
	ChooseResolution(requested Timerange, slotLimit int) time.Duration
	FetchSingleTimeseries(request FetchTimeseriesRequest) (Timeseries, error)
	FetchMultipleTimeseries(request FetchMultipleTimeseriesRequest) (SeriesList, error)
}

type ProfilingTimeseriesStorageAPI struct {
	Profiler             *inspect.Profiler
	TimeseriesStorageAPI TimeseriesStorageAPI
}

var _ TimeseriesStorageAPI = (*ProfilingTimeseriesStorageAPI)(nil)

func (a ProfilingTimeseriesStorageAPI) ChooseResolution(requested Timerange, slotLimit int) time.Duration {
	defer a.Profiler.Record("timeseriesStorage.ChooseResolution")()
	return a.TimeseriesStorageAPI.ChooseResolution(requested, slotLimit)
}

func (a ProfilingTimeseriesStorageAPI) FetchSingleTimeseries(request FetchTimeseriesRequest) (Timeseries, error) {
	defer a.Profiler.Record("timeseriesStorage.FetchSingleTimeseries")()
	return a.TimeseriesStorageAPI.FetchSingleTimeseries(request)
}

func (a ProfilingTimeseriesStorageAPI) FetchMultipleTimeseries(request FetchMultipleTimeseriesRequest) (SeriesList, error) {
	defer a.Profiler.Record("timeseriesStorage.FetchMultipleTimeseries")()
	return a.TimeseriesStorageAPI.FetchMultipleTimeseries(request)
}

type FetchTimeseriesRequest struct {
	Metric         TaggedMetric // metric to fetch.
	SampleMethod   SampleMethod // up/downsampling behavior.
	Timerange      Timerange    // time range to fetch data from.
	MetricMetadata MetricMetadataAPI
	Cancellable    Cancellable
	Profiler       *inspect.Profiler
}

type FetchMultipleTimeseriesRequest struct {
	Metrics        []TaggedMetric
	SampleMethod   SampleMethod
	Timerange      Timerange
	MetricMetadata MetricMetadataAPI
	Cancellable    Cancellable
	Profiler       *inspect.Profiler
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
	Metric  TaggedMetric
	Code    TimeseriesStorageErrorCode
	Message string
}

func (err TimeseriesStorageError) Error() string {
	message := "[%s] unknown error"
	switch err.Code {
	case FetchTimeoutError:
		message = "[%s] timeout"
	case InvalidSeriesError:
		message = "[%s] invalid series"
	case LimitError:
		message = "[%s] limit reached"
	case Unsupported:
		message = "[%s] unsupported operation"
	}
	formatted := fmt.Sprintf(message, string(err.Metric.MetricKey))
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
	fetchSingleRequests := make([]FetchTimeseriesRequest, 0)
	for _, metric := range r.Metrics {
		request := FetchTimeseriesRequest{
			Metric:         metric,
			MetricMetadata: r.MetricMetadata,
			Cancellable:    r.Cancellable,
			SampleMethod:   r.SampleMethod,
			Timerange:      r.Timerange,
			Profiler:       r.Profiler,
		}
		fetchSingleRequests = append(fetchSingleRequests, request)
	}
	return fetchSingleRequests
}
