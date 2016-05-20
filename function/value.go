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
	return ConversionError{
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
	return fmt.Sprintf("cannot convert %s (type %s) to type %s", e.Context, e.From, e.To)
}

// A SeriesListValue holds a SeriesList.
type SeriesListValue api.SeriesList

// ToSeriesList is an identity function that allows SeriesList to implement the expression.Value interface.
func (list SeriesListValue) ToSeriesList(time api.Timerange) (api.SeriesList, *ConversionFailure) {
	return api.SeriesList(list), nil
}

// ToString is a conversion function to implement the expression.Value interface.
func (list SeriesListValue) ToString() (string, *ConversionFailure) {
	return "", &ConversionFailure{"series list", "string"}
}

// ToScalar is a conversion function to implement the expression.Value interface.
func (list SeriesListValue) ToScalar() (float64, *ConversionFailure) {
	return 0, &ConversionFailure{"series list", "scalar"}
}

// ToScalarSet is a conversion function to implement the expression.Value interface.
func (list SeriesListValue) ToScalarSet() (ScalarSet, *ConversionFailure) {
	return nil, &ConversionFailure{"series list", "scalar set"}
}

// ToDuration is a conversion function to implement the expression.Value interface.
func (list SeriesListValue) ToDuration() (time.Duration, *ConversionFailure) {
	return 0, &ConversionFailure{"series list", "duration"}
}

// A StringValue holds a string
type StringValue string

// ToSeriesList is a conversion function.
func (value StringValue) ToSeriesList(time api.Timerange) (api.SeriesList, *ConversionFailure) {
	return api.SeriesList{}, &ConversionFailure{"string", "SeriesList"}
}

// ToString is a conversion function.
func (value StringValue) ToString() (string, *ConversionFailure) {
	return string(value), nil
}

// ToScalar is a conversion function.
func (value StringValue) ToScalar() (float64, *ConversionFailure) {
	return 0, &ConversionFailure{"string", "scalar"}
}

// ToScalarSet is a conversion function.
func (value StringValue) ToScalarSet() (ScalarSet, *ConversionFailure) {
	return nil, &ConversionFailure{"string", "scalar set"}
}

// ToDuration is a conversion function.
func (value StringValue) ToDuration() (time.Duration, *ConversionFailure) {
	return 0, &ConversionFailure{"string", "duration"}
}

// A ScalarValue holds a float and can be converted to a serieslist
type ScalarValue float64

// ToSeriesList is a conversion function.
// The scalar becomes a constant value for the timerange.
func (value ScalarValue) ToSeriesList(timerange api.Timerange) (api.SeriesList, *ConversionFailure) {
	series := make([]float64, timerange.Slots())
	for i := range series {
		series[i] = float64(value)
	}

	return api.SeriesList{
		Series:    []api.Timeseries{{Values: series, TagSet: api.NewTagSet()}},
		Timerange: timerange,
	}, nil
}

// ToString is a conversion function. Numbers become formatted.
func (value ScalarValue) ToString() (string, *ConversionFailure) {
	return "", &ConversionFailure{"scalar", "string"}
}

// ToScalar is a conversion function.
func (value ScalarValue) ToScalar() (float64, *ConversionFailure) {
	return float64(value), nil
}

// ToScalarSet is a conversion function.
func (value ScalarValue) ToScalarSet() (ScalarSet, *ConversionFailure) {
	// Return a singleton set.
	return ScalarSet{{nil, float64(value)}}, nil
}

// ToDuration is a conversion function.
// Scalars cannot be converted to durations.
func (value ScalarValue) ToDuration() (time.Duration, *ConversionFailure) {
	return 0, &ConversionFailure{"scalar", "duration"}
}

// DurationValue is a duration with a (usually) human-written name.
type DurationValue time.Duration

// ToSeriesList is a conversion function.
func (value DurationValue) ToSeriesList(timerange api.Timerange) (api.SeriesList, *ConversionFailure) {
	return api.SeriesList{}, &ConversionFailure{"duration", "SeriesList"}
}

// ToString is a conversion function.
func (value DurationValue) ToString() (string, *ConversionFailure) {
	return "", &ConversionFailure{"duration", "string"}
}

// ToScalar is a conversion function.
func (value DurationValue) ToScalar() (float64, *ConversionFailure) {
	return 0, &ConversionFailure{"duration", "scalar"}
}

// ToScalarSet is a conversion function.
func (value DurationValue) ToScalarSet() (ScalarSet, *ConversionFailure) {
	return nil, &ConversionFailure{"duration", "scalar set"}
}

// ToDuration is a conversion function.
func (value DurationValue) ToDuration() (time.Duration, *ConversionFailure) {
	return time.Duration(value), nil
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

type TaggedScalar struct {
	TagSet api.TagSet
	Value  float64
}

type ScalarSet []TaggedScalar

// ToSeriesList is a conversion function.
// The scalar becomes a constant value for the timerange.
func (set ScalarSet) ToSeriesList(timerange api.Timerange) (api.SeriesList, *ConversionFailure) {
	result := api.SeriesList{
		Series:    make([]api.Timeseries, len(set)),
		Timerange: timerange,
	}
	for i := range result.Series {
		result.Series[i] = api.Timeseries{
			Values: make([]float64, timerange.Slots()),
			TagSet: set[i].TagSet,
		}
		for j := range result.Series[i].Values {
			result.Series[i].Values[j] = set[i].Value
		}
	}
	return result, nil
}

// ToString is a conversion function. Numbers become formatted.
func (set ScalarSet) ToString() (string, *ConversionFailure) {
	return "", &ConversionFailure{"scalar set", "string"}
}

// ToScalar is a conversion function.
func (set ScalarSet) ToScalar() (float64, *ConversionFailure) {
	if len(set) == 1 && set[0].TagSet.Equals(api.TagSet{}) {
		return set[0].Value, nil
	}
	return 0, &ConversionFailure{"scalar set", "scalar"}
}

// ToScalarSet is a conversion function.
func (set ScalarSet) ToScalarSet() (ScalarSet, *ConversionFailure) {
	return set, nil
}

// ToDuration is a conversion function.
func (set ScalarSet) ToDuration() (time.Duration, *ConversionFailure) {
	return 0, &ConversionFailure{"scalar set", "duration"}
}
