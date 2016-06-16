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
	ToSeriesList(api.Timerange) (api.SeriesList, *ConversionFailure)
	ToString() (string, *ConversionFailure)
	ToScalar() (float64, *ConversionFailure)
	ToScalarSet() (ScalarSet, *ConversionFailure)
	ToDuration() (time.Duration, *ConversionFailure)
}

type ConversionFailure struct {
	From string // the original data type
	To   string // the type that it attempted to convert to
}

// WithContext adds enough context to make the ConversionFailure into an error.
func (c *ConversionFailure) WithContext(context string) ConversionError {
	// @@ leaking param: c to result ~r1 level=1
	// @@ leaking param: context to result ~r1 level=0
	return ConversionError{
		// @@ can inline (*ConversionFailure).WithContext
		From:    c.From,
		To:      c.To,
		Context: context,
	}
}

// ConversionError represents an error converting between two items of different types.
type ConversionError struct {
	From    string // the original data type
	To      string // the type that attempted to convert to
	Context string // a short string representation of the value
}

// Error gives a readable description of the error.
func (e ConversionError) Error() string {
	// @@ leaking param: e
	return fmt.Sprintf("cannot convert %s (type %s) to type %s", e.Context, e.From, e.To)
}

// @@ e.Context escapes to heap
// @@ e.From escapes to heap
// @@ e.To escapes to heap

// A SeriesListValue holds a SeriesList.
type SeriesListValue api.SeriesList

// ToSeriesList is an identity function that allows SeriesList to implement the Value interface.
func (list SeriesListValue) ToSeriesList(time api.Timerange) (api.SeriesList, *ConversionFailure) {
	// @@ leaking param: list to result ~r1 level=0
	return api.SeriesList(list), nil
	// @@ can inline SeriesListValue.ToSeriesList
}

// ToString is a conversion function to implement the Value interface.
func (list SeriesListValue) ToString() (string, *ConversionFailure) {
	return "", &ConversionFailure{"series list", "string"}
	// @@ can inline SeriesListValue.ToString
}

// @@ &ConversionFailure literal escapes to heap

// ToScalar is a conversion function to implement the Value interface.
func (list SeriesListValue) ToScalar() (float64, *ConversionFailure) {
	return 0, &ConversionFailure{"series list", "scalar"}
	// @@ can inline SeriesListValue.ToScalar
}

// @@ &ConversionFailure literal escapes to heap

// ToScalarSet is a conversion function to implement the Value interface.
func (list SeriesListValue) ToScalarSet() (ScalarSet, *ConversionFailure) {
	return nil, &ConversionFailure{"series list", "scalar set"}
	// @@ can inline SeriesListValue.ToScalarSet
}

// @@ &ConversionFailure literal escapes to heap

// ToDuration is a conversion function to implement the Value interface.
func (list SeriesListValue) ToDuration() (time.Duration, *ConversionFailure) {
	return 0, &ConversionFailure{"series list", "duration"}
	// @@ can inline SeriesListValue.ToDuration
}

// @@ &ConversionFailure literal escapes to heap

// A StringValue holds a string
type StringValue string

// ToSeriesList is a conversion function.
func (value StringValue) ToSeriesList(time api.Timerange) (api.SeriesList, *ConversionFailure) {
	return api.SeriesList{}, &ConversionFailure{"string", "SeriesList"}
	// @@ can inline StringValue.ToSeriesList
}

// @@ &ConversionFailure literal escapes to heap

// ToString is a conversion function.
func (value StringValue) ToString() (string, *ConversionFailure) {
	// @@ leaking param: value to result ~r0 level=0
	return string(value), nil
	// @@ can inline StringValue.ToString
}

// ToScalar is a conversion function.
func (value StringValue) ToScalar() (float64, *ConversionFailure) {
	return 0, &ConversionFailure{"string", "scalar"}
	// @@ can inline StringValue.ToScalar
}

// @@ &ConversionFailure literal escapes to heap

// ToScalarSet is a conversion function.
func (value StringValue) ToScalarSet() (ScalarSet, *ConversionFailure) {
	return nil, &ConversionFailure{"string", "scalar set"}
	// @@ can inline StringValue.ToScalarSet
}

// @@ &ConversionFailure literal escapes to heap

// ToDuration is a conversion function.
func (value StringValue) ToDuration() (time.Duration, *ConversionFailure) {
	return 0, &ConversionFailure{"string", "duration"}
	// @@ can inline StringValue.ToDuration
}

// @@ &ConversionFailure literal escapes to heap

// A ScalarValue holds a float and can be converted to a serieslist
type ScalarValue float64

// ToSeriesList is a conversion function.
// The scalar becomes a constant value for the timerange.
func (value ScalarValue) ToSeriesList(timerange api.Timerange) (api.SeriesList, *ConversionFailure) {
	series := make([]float64, timerange.Slots())
	for i := range series {
		// @@ inlining call to api.Timerange.Slots
		// @@ make([]float64, int(~r0)) escapes to heap
		// @@ make([]float64, int(~r0)) escapes to heap
		series[i] = float64(value)
	}

	return api.SeriesList{
		Series: []api.Timeseries{{Values: series, TagSet: api.NewTagSet()}},
	}, nil
	// @@ inlining call to api.NewTagSet
	// @@ make(map[string]string) escapes to heap
	// @@ []api.Timeseries literal escapes to heap
}

// ToString is a conversion function. Numbers become formatted.
func (value ScalarValue) ToString() (string, *ConversionFailure) {
	return "", &ConversionFailure{"scalar", "string"}
	// @@ can inline ScalarValue.ToString
}

// @@ &ConversionFailure literal escapes to heap

// ToScalar is a conversion function.
func (value ScalarValue) ToScalar() (float64, *ConversionFailure) {
	return float64(value), nil
	// @@ can inline ScalarValue.ToScalar
}

// ToScalarSet is a conversion function.
func (value ScalarValue) ToScalarSet() (ScalarSet, *ConversionFailure) {
	return ScalarSet{
		// @@ can inline ScalarValue.ToScalarSet
		TaggedScalar{
			// @@ ScalarSet literal escapes to heap
			Value:  float64(value),
			TagSet: api.TagSet{},
		},
		// @@ api.TagSet literal escapes to heap
	}, nil
}

// ToDuration is a conversion function.
// Scalars cannot be converted to durations.
func (value ScalarValue) ToDuration() (time.Duration, *ConversionFailure) {
	return 0, &ConversionFailure{"scalar", "duration"}
	// @@ can inline ScalarValue.ToDuration
}

// @@ &ConversionFailure literal escapes to heap

// DurationValue is a duration with a (usually) human-written name.
type DurationValue struct {
	name     string
	duration time.Duration
}

// NewDurationValue creates a duration value with the given name and duration.
func NewDurationValue(name string, duration time.Duration) DurationValue {
	// @@ leaking param: name to result ~r2 level=0
	return DurationValue{name, duration}
	// @@ can inline NewDurationValue
}

// ToSeriesList is a conversion function.
func (value DurationValue) ToSeriesList(timerange api.Timerange) (api.SeriesList, *ConversionFailure) {
	return api.SeriesList{}, &ConversionFailure{"duration", "SeriesList"}
	// @@ can inline DurationValue.ToSeriesList
}

// @@ &ConversionFailure literal escapes to heap

// ToString is a conversion function.
func (value DurationValue) ToString() (string, *ConversionFailure) {
	return "", &ConversionFailure{"duration", "string"}
	// @@ can inline DurationValue.ToString
}

// @@ &ConversionFailure literal escapes to heap

// ToScalar is a conversion function.
func (value DurationValue) ToScalar() (float64, *ConversionFailure) {
	return 0, &ConversionFailure{"duration", "scalar"}
	// @@ can inline DurationValue.ToScalar
}

// @@ &ConversionFailure literal escapes to heap

func (value DurationValue) ToScalarSet() (ScalarSet, *ConversionFailure) {
	return nil, &ConversionFailure{"duration", "scalar set"}
	// @@ can inline DurationValue.ToScalarSet
}

// @@ &ConversionFailure literal escapes to heap

// ToDuration is a conversion function.
func (value DurationValue) ToDuration() (time.Duration, *ConversionFailure) {
	return time.Duration(value.duration), nil
	// @@ can inline DurationValue.ToDuration
}

var durationRegexp = regexp.MustCompile(`^([+-]?[0-9]+)([smhdwMy]|ms|hr|mo|yr)$`)

// StringToDuration parses strings into timesdurations by examining their suffixes.
func StringToDuration(timeString string) (time.Duration, error) {
	// @@ leaking param: timeString
	matches := durationRegexp.FindStringSubmatch(timeString)
	if matches == nil {
		return -1, fmt.Errorf("expected duration to be of the form `%s`", durationRegexp.String())
	}
	// @@ inlining call to String
	// @@ durationRegexp.String() escapes to heap
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
