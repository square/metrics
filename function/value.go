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

package function

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/square/metrics/api"
)

// Value is the result of evaluating an expression.
// They can be floating point values, strings, or series lists.
type Value interface {
	ToSeriesList(api.Timerange) (api.SeriesList, error)
	ToString() (string, error)
	ToScalar() (float64, error)
	ToDuration() (time.Duration, error)
}

// ConversionError represents an error converting between two items of different types.
type ConversionError struct {
	From  string // the original data type
	To    string // the type that attempted to convert to
	Value string // a short string representation of the value
}

// Error gives a readable description of the error.
func (e ConversionError) Error() string {
	return fmt.Sprintf("cannot convert %s (type %s) to type %s", e.Value, e.From, e.To)
}

// TokenName gives the token name where the error occurred.
func (e ConversionError) TokenName() string {
	return fmt.Sprintf("%+v (type %s)", e.Value, e.From)
}

// A StringValue holds a string
type StringValue string

// ToSeriesList is a conversion function.
func (value StringValue) ToSeriesList(time api.Timerange) (api.SeriesList, error) {
	return api.SeriesList{}, ConversionError{"string", "SeriesList", "// TODO //"}
}

// ToString is a conversion function.
func (value StringValue) ToString() (string, error) {
	return string(value), nil
}

// ToScalar is a conversion function.
func (value StringValue) ToScalar() (float64, error) {
	return 0, ConversionError{"string", "scalar", ""}
}

// ToDuration is a conversion function.
func (value StringValue) ToDuration() (time.Duration, error) {
	return 0, ConversionError{"string", "duration", ""}
}

// A ScalarValue holds a float and can be converted to a serieslist
type ScalarValue float64

// ToSeriesList is a conversion function.
// The scalar becomes a constant value for the timerange.
func (value ScalarValue) ToSeriesList(timerange api.Timerange) (api.SeriesList, error) {
	series := make([]float64, timerange.Slots())
	for i := range series {
		series[i] = float64(value)
	}

	return api.SeriesList{
		Series: []api.Timeseries{{Values: series, TagSet: api.NewTagSet()}},
	}, nil
}

// ToString is a conversion function. Numbers become formatted.
func (value ScalarValue) ToString() (string, error) {
	return "", ConversionError{"scalar", "string", fmt.Sprintf("%f", value)}
}

// ToScalar is a conversion function.
func (value ScalarValue) ToScalar() (float64, error) {
	return float64(value), nil
}

// ToDuration is a conversion function.
// Scalars cannot be converted to durations.
func (value ScalarValue) ToDuration() (time.Duration, error) {
	return 0, ConversionError{"scalar", "duration", ""}
}

// DurationValue is a duration with a (usually) human-written name.
type DurationValue struct {
	name     string
	duration time.Duration
}

// NewDurationValue creates a duration value with the given name and duration.
func NewDurationValue(name string, duration time.Duration) DurationValue {
	return DurationValue{name, duration}
}

// ToSeriesList is a conversion function.
func (value DurationValue) ToSeriesList(timerange api.Timerange) (api.SeriesList, error) {
	return api.SeriesList{}, ConversionError{"duration", "SeriesList", "// TODO //"}
}

// ToString is a conversion function.
func (value DurationValue) ToString() (string, error) {
	return "", ConversionError{"duration", "string", ""}
}

// ToScalar is a conversion function.
func (value DurationValue) ToScalar() (float64, error) {
	return 0, ConversionError{"duration", "scalar", ""}
}

// ToDuration is a conversion function.
func (value DurationValue) ToDuration() (time.Duration, error) {
	return time.Duration(value.duration), nil
}

var durationRegexp = regexp.MustCompile(`^([+-]?[0-9]+)([smhdwMy]|ms|hr|mo|yr)$`)

// StringToDuration parses strings into timesdurations by examining their suffixes.
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
