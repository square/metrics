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
package transform

import (
	"fmt"
	"math"

	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
)

// A transform takes the list of values, other parameters, and the resolution (as a float64) of the query.
type transform func([]float64, []function.Value, float64) ([]float64, error)

// transformTimeseries transforms an individual series (rather than an entire serieslist) taking the same parameters as a transform,
// but with the serieslist standing in for the simplified []float64 argument.
func transformTimeseries(series api.Timeseries, transform transform, parameters []function.Value, scale float64) (api.Timeseries, error) {
	values, err := transform(series.Values, parameters, scale)
	if err != nil {
		return api.Timeseries{}, err
	}
	return api.Timeseries{
		Values: values,
		TagSet: series.TagSet,
	}, nil
}

// applyTransform applies the given transform to the entire list of series.
func ApplyTransform(list api.SeriesList, transform transform, parameters []function.Value) (api.SeriesList, error) {
	result := api.SeriesList{
		Series:    make([]api.Timeseries, len(list.Series)),
		Timerange: list.Timerange,
		Name:      list.Name,
		Query:     list.Query,
	}
	var err error
	for i, series := range list.Series {
		result.Series[i], err = transformTimeseries(series, transform, parameters, float64(list.Timerange.ResolutionMillis())/1000)
		if err != nil {
			return api.SeriesList{}, err
		}
	}
	return result, nil
}

// transformDerivative estimates the "change per second" between the two samples (scaled consecutive difference)
func Derivative(values []float64, parameters []function.Value, scale float64) ([]float64, error) {
	result := make([]float64, len(values))
	for i := range values {
		if i == 0 {
			// The first element has 0
			result[i] = 0
			continue
		}
		// Otherwise, it's the scaled difference
		result[i] = (values[i] - values[i-1]) / scale
	}
	return result, nil
}

// Integral integrates a series whose values are "X per millisecond" to estimate "total X so far"
// if the series represents "X in this sampling interval" instead, then you should use transformCumulative.
func Integral(values []float64, parameters []function.Value, scale float64) ([]float64, error) {
	result := make([]float64, len(values))
	integral := 0.0
	for i := range values {
		if !math.IsNaN(values[i]) {
			integral += values[i]
		}
		result[i] = integral * scale
	}
	return result, nil
}

// Rate functions exactly like transformDerivative but bounds the result to be positive.
// That is, it returns consecutive differences which are at least 0.
func Rate(values []float64, parameters []function.Value, scale float64) ([]float64, error) {
	result := make([]float64, len(values))
	for i := range values {
		if i == 0 {
			result[i] = 0
			continue
		}
		result[i] = (values[i] - values[i-1]) / scale
		if result[i] < 0 {
			result[i] = 0
		}
	}
	return result, nil
}

// Cumulative computes the cumulative sum of the given values.
func Cumulative(values []float64, parameters []function.Value, scale float64) ([]float64, error) {
	result := make([]float64, len(values))
	sum := 0.0
	for i := range values {
		if !math.IsNaN(values[i]) {
			sum += values[i]
		}
		result[i] = sum
	}
	return result, nil
}

// MapMaker can be used to use a function as a transform, such as 'math.Abs' (or similar):
//  `MapMaker(math.Abs)` is a transform function which can be used, e.g. with ApplyTransform
// The name is used for error-checking purposes.
func MapMaker(fun func(float64) float64) func([]float64, []function.Value, float64) ([]float64, error) {
	return func(values []float64, parameters []function.Value, scale float64) ([]float64, error) {
		result := make([]float64, len(values))
		for i := range values {
			result[i] = fun(values[i])
		}
		return result, nil
	}
}

// Default will replacing missing data (NaN) with the `default` value supplied as a parameter.
func Default(values []float64, parameters []function.Value, scale float64) ([]float64, error) {
	defaultValue, err := parameters[0].ToScalar()
	if err != nil {
		return nil, err
	}
	result := make([]float64, len(values))
	for i := range values {
		if math.IsNaN(values[i]) {
			result[i] = defaultValue
		} else {
			result[i] = values[i]
		}
	}
	return result, nil
}

// NaNKeepLast will replace missing NaN data with the data before it
func NaNKeepLast(values []float64, parameters []function.Value, scale float64) ([]float64, error) {
	result := make([]float64, len(values))
	for i := range result {
		result[i] = values[i]
		if math.IsNaN(result[i]) && i > 0 {
			result[i] = result[i-1]
		}
	}
	return result, nil
}

// boundError represents an error in bounds, when (lower > upper) so the interval is empty.
type boundError struct {
	lower float64
	upper float64
}

func (b boundError) Error() string {
	return fmt.Sprintf("the lower bound (%f) should be no more than the upper bound (%f) in the parameters to transform.bound( ..., %f, %f)", b.lower, b.upper, b.lower, b.upper)
}

func (b boundError) TokenName() string {
	return fmt.Sprintf("transform.bound(..., %f, %f)", b.lower, b.upper)
}

// Bound replaces values which fall outside the given limits with the limits themselves. If the lowest bound exceeds the upper bound, an error is returned.
func Bound(values []float64, parameters []function.Value, scale float64) ([]float64, error) {
	lowerBound, err := parameters[0].ToScalar()
	if err != nil {
		return nil, err
	}
	upperBound, err := parameters[1].ToScalar()
	if err != nil {
		return nil, err
	}
	if lowerBound > upperBound {
		return nil, boundError{lowerBound, upperBound}
	}
	result := make([]float64, len(values))
	for i := range result {
		result[i] = values[i]
		if result[i] < lowerBound {
			result[i] = lowerBound
		}
		if result[i] > upperBound {
			result[i] = upperBound
		}
	}
	return result, nil
}

// LowerBound replaces values that fall below the given bound with the lower bound.
func LowerBound(values []float64, parameters []function.Value, scale float64) ([]float64, error) {
	lowerBound, err := parameters[0].ToScalar()
	if err != nil {
		return nil, err
	}
	result := make([]float64, len(values))
	for i := range result {
		result[i] = values[i]
		if result[i] < lowerBound {
			result[i] = lowerBound
		}
	}
	return result, nil
}

// UpperBound replaces values that fall below the given bound with the lower bound.
func UpperBound(values []float64, parameters []function.Value, scale float64) ([]float64, error) {
	upperBound, err := parameters[0].ToScalar()
	if err != nil {
		return nil, err
	}
	result := make([]float64, len(values))
	for i := range result {
		result[i] = values[i]
		if result[i] > upperBound {
			result[i] = upperBound
		}
	}
	return result, nil
}
