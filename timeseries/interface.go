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

package timeseries

import (
	"fmt"
	"net/http"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/inspect"

	"golang.org/x/net/context"
)

type StorageAPI interface {
	ChooseResolution(requested api.Timerange, lowerBound time.Duration) (time.Duration, error)
	FetchSingleTimeseries(request FetchRequest) (api.Timeseries, error)
	FetchMultipleTimeseries(request FetchMultipleRequest) (api.SeriesList, error)
	// CheckHealthy checks if this StorageAPI is healthy, returning a possible error
	CheckHealthy() error
}

type RequestDetails struct {
	SampleMethod SampleMethod    // up/downsampling behavior.
	Timerange    api.Timerange   // time range to fetch data from.
	Ctx          context.Context // context includes timeout details
	Profiler     *inspect.Profiler
}

type FetchRequest struct {
	Metric api.TaggedMetric // metric to fetch.
	RequestDetails
}

type FetchMultipleRequest struct {
	Metrics []api.TaggedMetric
	RequestDetails
}

type ErrorCode int

// FetchError can return a custom error code
type FetchError struct {
	Message string
	Code    int
}

// Error returns the message associated with the FetchError.
func (e FetchError) Error() string {
	return e.Message
}

// ErrorCode returns the error code of the fetch error.
func (e FetchError) ErrorCode() int {
	if e.Code == 0 {
		return http.StatusBadRequest
	}
	return e.Code
}

const (
	FetchTimeoutError  ErrorCode = iota + 1 // FetchTimeoutError indicates a timeout happened
	FetchIOError                            // FetchIOError indicates an IO error occurred
	InvalidSeriesError                      // InvalidSeriesError indicates the requested series was ill-formed
	LimitError                              // LimitError indicates a resource limit was reached
	Unsupported                             // Unsupported indicates an operation was attempted which is not supported
)

type Error struct {
	Metric  api.TaggedMetric
	Code    ErrorCode
	Message string
}

func (err Error) Error() string {
	message := "unknown error"
	switch err.Code {
	case FetchTimeoutError:
		message = "timeout"
	case FetchIOError:
		message = "IO error"
	case InvalidSeriesError:
		message = "invalid series"
	case LimitError:
		message = "limit reached"
	case Unsupported:
		message = "unsupported operation"
	}
	formatted := fmt.Sprintf("[%s %+v] %s", string(err.Metric.MetricKey), err.Metric.TagSet, message)
	if err.Message != "" {
		formatted = formatted + " - " + err.Message
	}
	return formatted
}

// ToSingle very simply decompose the FetchMultipleTimeseriesRequest into single
// fetch requests (for now).
func (r FetchMultipleRequest) ToSingle() []FetchRequest {
	fetchSingleRequests := make([]FetchRequest, len(r.Metrics))
	for i, metric := range r.Metrics {
		fetchSingleRequests[i] = FetchRequest{
			Metric:         metric,
			RequestDetails: r.RequestDetails,
		}
	}
	return fetchSingleRequests
}
