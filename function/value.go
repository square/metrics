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

package function

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/square/metrics/api"
)

// A value is the result of evaluating an expression.
// They can be floating point values, strings, or series lists.
type Value interface {
	ToSeriesList(api.Timerange) (api.SeriesList, error)
	ToString() (string, error)
	ToScalar() (float64, error)
	GetName() string
}

type conversionError struct {
	from string
	to   string
}

func (e conversionError) Error() string {
	return fmt.Sprintf("cannot convert from type %s to type %s", e.from, e.to)
}

// A seriesListValue is a value which holds a SeriesList
type SeriesListValue api.SeriesList

func (value SeriesListValue) ToSeriesList(time api.Timerange) (api.SeriesList, error) {
	return api.SeriesList(value), nil
}
func (value SeriesListValue) ToString() (string, error) {
	return "", conversionError{"SeriesList", "string"}
}
func (value SeriesListValue) ToScalar() (float64, error) {
	return 0, conversionError{"SeriesList", "scalar"}
}
func (value SeriesListValue) GetName() string {
	return api.SeriesList(value).Name
}

// A stringValue holds a string
type StringValue string

func (value StringValue) ToSeriesList(time api.Timerange) (api.SeriesList, error) {
	return api.SeriesList{}, conversionError{"string", "SeriesList"}
}
func (value StringValue) ToString() (string, error) {
	return string(value), nil
}
func (value StringValue) ToScalar() (float64, error) {
	return 0, conversionError{"string", "scalar"}
}
func (value StringValue) GetName() string {
	return string(value)
}

// A scalarValue holds a float and can be converted to a serieslist
type ScalarValue float64

func (value ScalarValue) ToSeriesList(timerange api.Timerange) (api.SeriesList, error) {

	series := make([]float64, timerange.Slots())
	for i := range series {
		series[i] = float64(value)
	}

	return api.SeriesList{
		Series:    []api.Timeseries{api.Timeseries{series, api.NewTagSet()}},
		Timerange: timerange,
	}, nil
}
func (value ScalarValue) ToString() (string, error) {
	return "", conversionError{"scalar", "string"}
}
func (value ScalarValue) ToScalar() (float64, error) {
	return float64(value), nil
}
func (value ScalarValue) GetName() string {
	return fmt.Sprintf("%g", value)
}

var durationRegexp = regexp.MustCompile(`^([+-]?[0-9]+)([smhdwMy]|ms|hr|mo|yr)$`)

// toDuration will take a value, convert it to a string, and then parse it.
// the valid suffixes are: ms, s, m, min, h, hr, d, w, M, mo, y, yr.
// It converts the return value to milliseconds.
func ToDuration(value Value) (int64, error) {
	timeString, err := value.ToString()
	if err != nil {
		return -1, err
	}
	matches := durationRegexp.FindStringSubmatch(timeString)
	if matches == nil {
		return -1, fmt.Errorf("expected duration to be of the form `%s`", durationRegexp.String())
	}
	duration, err := strconv.ParseInt(matches[1], 10, 0)
	if err != nil {
		return -1, err
	}
	scale := int64(1)
	switch matches[2] {
	case "ms":
		// no change in scale
	case "s":
		scale = 1000
	case "m":
		scale = 1000 * 60
	case "h", "hr":
		scale = 1000 * 60 * 60
	case "d":
		scale = 1000 * 60 * 60 * 24
	case "w":
		scale = 1000 * 60 * 60 * 24 * 7
	case "M", "mo":
		scale = 1000 * 60 * 60 * 24 * 30
	case "y", "yr":
		scale = 1000 * 60 * 60 * 24 * 365
	}
	return int64(duration) * scale, nil
}
