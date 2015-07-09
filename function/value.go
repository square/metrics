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
	"time"

	"github.com/square/metrics/api"
)

// A value is the result of evaluating an expression.
// They can be floating point values, strings, or series lists.
type Value interface {
	ToSeriesList(api.Timerange) (api.SeriesList, error)
	ToString() (string, error)
	ToScalar() (float64, error)
	ToDuration() (time.Duration, error)
	GetName() string
}

type conversionError struct {
	from  string
	to    string
	value interface{}
}

func (e conversionError) Error() string {
	return fmt.Sprintf("cannot convert %+v (type %s) to type %s", e.value, e.from, e.to)
}

func (e conversionError) TokenName() string {
	return fmt.Sprintf("%+v (type %s)", e.value, e.from)
}

// A seriesListValue is a value which holds a SeriesList
type SeriesListValue api.SeriesList

func (value SeriesListValue) ToSeriesList(time api.Timerange) (api.SeriesList, error) {
	return api.SeriesList(value), nil
}

func (value SeriesListValue) ToString() (string, error) {
	return "", conversionError{"SeriesList", "string", fmt.Sprintf("serieslist[%s]", value.Name)}
}

func (value SeriesListValue) ToScalar() (float64, error) {
	return 0, conversionError{"SeriesList", "scalar", fmt.Sprintf("serieslist[%s]", value.Name)}
}

func (value SeriesListValue) ToDuration() (time.Duration, error) {
	return 0, conversionError{"SeriesList", "duration", fmt.Sprintf("serieslist[%s]", value.Name)}
}

func (value SeriesListValue) GetName() string {
	return api.SeriesList(value).Name
}

// A stringValue holds a string
type StringValue string

func (value StringValue) ToSeriesList(time api.Timerange) (api.SeriesList, error) {
	return api.SeriesList{}, conversionError{"string", "SeriesList", fmt.Sprintf("%q", value)}
}

func (value StringValue) ToString() (string, error) {
	return string(value), nil
}

func (value StringValue) ToScalar() (float64, error) {
	return 0, conversionError{"string", "scalar", fmt.Sprintf("%q", value)}
}

func (value StringValue) ToDuration() (time.Duration, error) {
	return 0, conversionError{"string", "duration", fmt.Sprintf("%q", value)}
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
	return "", conversionError{"scalar", "string", value}
}

func (value ScalarValue) ToScalar() (float64, error) {
	return float64(value), nil
}

func (value ScalarValue) ToDuration() (time.Duration, error) {
	return 0, conversionError{"scalar", "duration", value}
}

func (value ScalarValue) GetName() string {
	return fmt.Sprintf("%g", value)
}

type DurationValue struct {
	name     string
	duration time.Duration
}

func NewDurationValue(name string, duration time.Duration) DurationValue {
	return DurationValue{name, duration}
}

func (value DurationValue) ToSeriesList(timerange api.Timerange) (api.SeriesList, error) {
	return api.SeriesList{}, conversionError{"duration", "SeriesList", fmt.Sprintf("%dms", value.duration)}
}

func (value DurationValue) ToString() (string, error) {
	return "", conversionError{"duration", "string", fmt.Sprintf("%dms", value.name)}
}

func (value DurationValue) ToScalar() (float64, error) {
	return 0, conversionError{"duration", "scalar", fmt.Sprintf("%dms", value.name)}
}

func (value DurationValue) ToDuration() (time.Duration, error) {
	return time.Duration(value.duration), nil
}

func (value DurationValue) GetName() string {
	return value.name
}

var durationRegexp = regexp.MustCompile(`^([+-]?[0-9]+)([smhdwMy]|ms|hr|mo|yr)$`)

func StringToDuration(timeString string) (time.Duration, error) {
	matches := durationRegexp.FindStringSubmatch(timeString)
	if matches == nil {
		return -1, fmt.Errorf("expected duration to be of the form `%s`", durationRegexp.String())
	}
	duration, err := strconv.ParseInt(matches[1], 10, 0)
	if err != nil {
		return -1, err
	}
	scale := time.Millisecond
	switch matches[2] {
	case "ms":
		// no change in scale
	case "s":
		scale *= 1000
	case "m":
		scale *= 1000 * 60
	case "h", "hr":
		scale *= 1000 * 60 * 60
	case "d":
		scale *= 1000 * 60 * 60 * 24
	case "w":
		scale *= 1000 * 60 * 60 * 24 * 7
	case "M", "mo":
		scale *= 1000 * 60 * 60 * 24 * 30
	case "y", "yr":
		scale *= 1000 * 60 * 60 * 24 * 365
	}
	return time.Duration(duration) * scale, nil
}
