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
	"github.com/square/metrics/api"
)

// A transformation takes the list of values, and a "scale" factor:
// It is the unitless quotient (sample duration / 1s)
// So if the resolution is "once every 5 minutes" them we get a scale factor of 5*60 = 300
type transformation func([]float64, transformationParameter) []float64

type transformationParameter struct {
	scale     float64
	parameter float64
}

func transformTimeseries(series api.Timeseries, transformation transformation, parameter transformationParameter) api.Timeseries {
	return api.Timeseries{
		Values: transformation(series.Values, parameter),
		TagSet: series.TagSet,
	}
}

// applyTransformation applies the given transformation to the entire list of series.
func applyTransformation(list api.SeriesList, transformation transformation, parameter float64) api.SeriesList {
	result := api.SeriesList{
		Series:    make([]api.Timeseries, len(list.Series)),
		Timerange: list.Timerange,
		Name:      list.Name,
	}
	scale := float64(list.Timerange.Resolution)
	for i, series := range list.Series {
		result.Series[i] = transformTimeseries(series, transformation, transformationParameter{
			scale:     scale,
			parameter: parameter,
		})
	}
	return result
}

// transformDerivative estimates the "change per second" between the two samples (scaled consecutive difference)
func transformDerivative(values []float64, parameter transformationParameter) []float64 {
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
	return result
}

// transformIntegral integrates a series whose values are "X per second" to estimate "total X so far"
func transformIntegral(values []float64, parameter transformationParameter) []float64 {
	result := make([]float64, len(values))
	integral := 0.0
	for i := range values {
		integral += values[i]
		result[i] = integral * parameter.scale
	}
	return result
}

// transformRate functions exactly like transformDerivative but bounds the result to be positive and does not normalize
func transformRate(values []float64, parameter transformationParameter) []float64 {
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
	return result
}

// transformMovingAverage finds the average over the time period given in the parameter.parameter value
func transformMovingAverage(values []float64, parameter transformationParameter) []float64 {
	result := make([]float64, len(values))
	limit := int(parameter.parameter/parameter.scale + 0.5) // Limit is the number of items to include in the average
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
	return result
}

// transformMapMaker can be used to use a function as a transformation, such as 'math.Abs' (or similar):
//  `transformMapMaker(math.Abs)` is a transformation function which can be used, e.g. with applyTransformation
func transformMapMaker(fun func(float64) float64) func([]float64, transformationParameter) []float64 {
	return func(values []float64, parameter transformationParameter) []float64 {
		result := make([]float64, len(values))
		for i := range values {
			result[i] = fun(values[i])
		}
		return result
	}
}

func transformTimeOffset(values []float64, parameter transformationParameter) []float64 {
	result := make([]float64, len(values))
	// Shifting the time series by the given number of samples (by dividing the parameter by scale)
	shift := int(parameter.parameter/parameter.scale + 0.5) // The (+ 0.5) causes it to round.
	// Positive shift means forward in time (so result[shift] = values[0])
	// Negative shift means backwards in time.
	// Values outside this range are assigned 0
	for i := range values {
		if i-shift < 0 || i-shift >= len(values) {

			result[i] = 0
			continue
		}
		result[i] = values[i-shift]
	}
	return result
}
