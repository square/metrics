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

package query

import (
	"fmt"
	"time"

	"github.com/square/metrics/api"
)

// A value is the result of evaluating an expression.
// They can be floating point values, strings, or series lists.
type value interface {
	toSeriesList(api.Timerange) (api.SeriesList, error)
	toString() (string, error)
	toScalar() (float64, error)
}

type conversionError struct {
	from string
	to   string
}

func (e conversionError) Error() string {
	return fmt.Sprintf("cannot convert from type %s to type %s", e.from, e.to)
}

// A seriesListValue is a value which holds a SeriesList
type seriesListValue api.SeriesList

func (value seriesListValue) toSeriesList(time api.Timerange) (api.SeriesList, error) {
	return api.SeriesList(value), nil
}
func (value seriesListValue) toString() (string, error) {
	return "", conversionError{"SeriesList", "string"}
}
func (value seriesListValue) toScalar() (float64, error) {
	return 0, conversionError{"SeriesList", "scalar"}
}

// A stringValue holds a string
type stringValue string

func (value stringValue) toSeriesList(time api.Timerange) (api.SeriesList, error) {
	return api.SeriesList{}, conversionError{"string", "SeriesList"}
}
func (value stringValue) toString() (string, error) {
	return string(value), nil
}
func (value stringValue) toScalar() (float64, error) {
	return 0, conversionError{"string", "scalar"}
}

// A scalarValue holds a float and can be converted to a serieslist
type scalarValue float64

func (value scalarValue) toSeriesList(timerange api.Timerange) (api.SeriesList, error) {

	series := make([]float64, timerange.Slots())
	for i := range series {
		series[i] = float64(value)
	}

	return api.SeriesList{
		Series:    []api.Timeseries{api.Timeseries{series, api.NewTagSet()}},
		Timerange: timerange,
	}, nil
}

func (value scalarValue) toString() (string, error) {
	return "", conversionError{"scalar", "string"}
}

func (value scalarValue) toScalar() (float64, error) {
	return float64(value), nil
}

// toDuration will take a value, convert it to a string, and then parse it.
// the valid suffixes are: ns, us (µs), ms, s, m, h
// It converts the return value to milliseconds.
func toDuration(value value) (int64, error) {
	timeString, err := value.toString()
	if err != nil {
		return 0, err
	}
	duration, err := time.ParseDuration(timeString)
	if err != nil {
		return 0, err
	}
	return int64(duration / 1000000), nil
}
