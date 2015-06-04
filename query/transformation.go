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

// Package query contains all the logic to parse
// and execute queries against the underlying metric system.
package query

import (
	"errors"
	"fmt"

	"github.com/square/metrics/api"
)

// A transform takes the list of values, and a "scale" factor:
// It is the unitless quotient (sample duration / 1s)
// So if the resolution is "once every 5 minutes" them we get a scale factor of 5*60 = 300
type transform func([]float64, transformParameter) ([]float64, error)

type transformParameter struct {
	scale      float64
	parameters []value
}

func transformTimeseries(series api.Timeseries, transform transform, parameter transformParameter) (api.Timeseries, error) {
	values, err := transform(series.Values, parameter)
	if err != nil {
		return api.Timeseries{}, err
	}
	return api.Timeseries{
		Values: values,
		TagSet: series.TagSet,
	}, nil
}

// applyTransform applies the given transform to the entire list of series.
func ApplyTransform(list api.SeriesList, transform transform, parameters []value) (api.SeriesList, error) {
	result := api.SeriesList{
		Series:    make([]api.Timeseries, len(list.Series)),
		Timerange: list.Timerange,
		Name:      list.Name,
	}
	scale := float64(list.Timerange.Resolution)
	var err error
	for i, series := range list.Series {
		result.Series[i], err = transformTimeseries(series, transform, transformParameter{
			scale:      scale,
			parameters: parameters,
		})
		if err != nil {
			return api.SeriesList{}, err
		}
	}
	return result, nil
}

func checkParameters(name string, expected int, parameter transformParameter) error {
	args := parameter.parameters
	if len(args) != expected {
		printArgs := append([]value{stringValue("(SeriesList)")}, args...)
		return errors.New(fmt.Sprintf("expected %s to be given %d parameters but was given %d: %+v", name, expected+1, len(args)+1, printArgs))
	}
	return nil
}

// transformDerivative estimates the "change per second" between the two samples (scaled consecutive difference)
func transformDerivative(values []float64, parameter transformParameter) ([]float64, error) {
	if err := checkParameters("transform.derivative", 0, parameter); err != nil {
		return nil, err
	}
	result := make([]float64, len(values))
	for i := range values {
		if i == 0 {
			// The first element has 0
			result[i] = 0
			continue
		}
		// Otherwise, it's the scaled difference
		result[i] = (values[i] - values[i-1]) / parameter.scale
	}
	return result, nil
}

// transformIntegral integrates a series whose values are "X per second" to estimate "total X so far"
func transformIntegral(values []float64, parameter transformParameter) ([]float64, error) {
	if err := checkParameters("transform.integral", 0, parameter); err != nil {
		return nil, err
	}
	result := make([]float64, len(values))
	integral := 0.0
	for i := range values {
		integral += values[i]
		result[i] = integral * parameter.scale
	}
	return result, nil
}

// transformRate functions exactly like transformDerivative but bounds the result to be positive and does not normalize
func transformRate(values []float64, parameter transformParameter) ([]float64, error) {
	if err := checkParameters("transform.rate", 0, parameter); err != nil {
		return nil, err
	}
	result := make([]float64, len(values))
	for i := range values {
		if i == 0 {
			result[i] = 0
			continue
		}
		result[i] = (values[i] - values[i-1]) / parameter.scale
		if result[i] < 0 {
			result[i] = 0
		}
	}
	return result, nil
}

// transformMovingAverage finds the average over the time period given in the parameter.parameter value
func transformMovingAverage(values []float64, parameter transformParameter) ([]float64, error) {
	if err := checkParameters("transform.moving_average", 1, parameter); err != nil {
		return nil, err
	}
	result := make([]float64, len(values))
	if len(parameter.parameters) != 1 {

	}
	size, err := parameter.parameters[0].toScalar()
	if err != nil {
		return nil, err
	}
	limit := int(size/parameter.scale + 0.5) // Limit is the number of items to include in the average
	if limit < 1 {
		// At least one value must be included at all times
		limit = 1
	}
	count := 0
	sum := 0.0
	for i := range values {
		sum += values[i]
		if count < limit {
			// Increment the number of participants.
			count++
		} else {
			// Remove the earliest participant
			// (This is an optimization)
			sum -= values[i-limit]
			// For example, if limit = 4, we are suppose to include 4 items (including oneself).
			// If i = 10, then we want to include 10, 9, 8, 7.
			// So exclude 6 = 10 - 4 = i - limit
		}
		result[i] = sum / float64(count)
	}
	return result, nil
}

// transformMapMaker can be used to use a function as a transform, such as 'math.Abs' (or similar):
//  `transformMapMaker(math.Abs)` is a transform function which can be used, e.g. with ApplyTransform
// The name is used for error-checking purposes.
func transformMapMaker(name string, fun func(float64) float64) func([]float64, transformParameter) ([]float64, error) {
	return func(values []float64, parameter transformParameter) ([]float64, error) {
		if err := checkParameters(fmt.Sprintf("transform.%s", name), 0, parameter); err != nil {
			return nil, err
		}
		result := make([]float64, len(values))
		for i := range values {
			result[i] = fun(values[i])
		}
		return result, nil
	}
}
