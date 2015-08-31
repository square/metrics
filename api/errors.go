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
)

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

type ConversionError struct {
	From  string // the original data type
	To    string // the type that attempted to convert to
	Value string // a short string representation of the value
}

func (e ConversionError) Error() string {
	return fmt.Sprintf("cannot convert %s (type %s) to type %s", e.Value, e.From, e.To)
}

func (e ConversionError) TokenName() string {
	return fmt.Sprintf("%+v (type %s)", e.Value, e.From)
}
