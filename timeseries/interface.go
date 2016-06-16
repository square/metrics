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
	"github.com/square/metrics/tasks"
)

type StorageAPI interface {
	ChooseResolution(requested api.Timerange, lowerBound time.Duration) (time.Duration, error)
	FetchSingleTimeseries(request FetchRequest) (api.Timeseries, error)
	FetchMultipleTimeseries(request FetchMultipleRequest) (api.SeriesList, error)
	// CheckHealthy checks if this StorageAPI is healthy, returning a possible error
	CheckHealthy() error
}

type RequestDetails struct {
	SampleMethod          SampleMethod  // up/downsampling behavior.
	Timerange             api.Timerange // time range to fetch data from.
	Timeout               *tasks.Timeout
	Profiler              *inspect.Profiler
	UserSpecifiableConfig UserSpecifiableConfig
}

type FetchRequest struct {
	Metric api.TaggedMetric // metric to fetch.
	RequestDetails
}

type FetchMultipleRequest struct {
	Metrics []api.TaggedMetric
	RequestDetails
}

type UserSpecifiableConfig struct {
	IncludeRawData bool
}

type ErrorCode int

// FetchError can return a custom error code
type FetchError struct {
	Message string
	Code    int
}

// Error returns the message associated with the FetchError.
func (e FetchError) Error() string {
	// @@ leaking param: e to result ~r0 level=0
	return e.Message
	// @@ can inline FetchError.Error
}

// Error500 indicates that it's a 500-level error.
func (e FetchError) ErrorCode() int {
	if e.Code == 0 {
		// @@ can inline FetchError.ErrorCode
		return http.StatusBadRequest
	}
	return e.Code
}

const (
	FetchTimeoutError  ErrorCode = iota + 1 // error while fetching - timeout.
	FetchIOError                            // error while fetching - general IO.
	InvalidSeriesError                      // the given series is not well-defined.
	LimitError                              // the fetch limit is reached.
	Unsupported                             // the given fetch operation is unsupported by the backend.
)

type Error struct {
	Metric  api.TaggedMetric
	Code    ErrorCode
	Message string
}

func (err Error) Error() string {
	// @@ leaking param: err
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
		// @@ string(err.Metric.MetricKey) escapes to heap
		// @@ err.Metric.TagSet escapes to heap
		// @@ message escapes to heap
		formatted = formatted + " - " + err.Message
	}
	// @@ formatted + " - " + err.Message escapes to heap
	return formatted
}

// ToSingle very simply decompose the FetchMultipleTimeseriesRequest into single
// fetch requests (for now).
func (r FetchMultipleRequest) ToSingle() []FetchRequest {
	// @@ leaking param content: r
	// @@ leaking param: r
	fetchSingleRequests := make([]FetchRequest, len(r.Metrics))
	for i, metric := range r.Metrics {
		// @@ make([]FetchRequest, len(r.Metrics)) escapes to heap
		// @@ make([]FetchRequest, len(r.Metrics)) escapes to heap
		fetchSingleRequests[i] = FetchRequest{
			Metric:         metric,
			RequestDetails: r.RequestDetails,
		}
	}
	return fetchSingleRequests
}
