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

// Package transform contains all the logic to parse
// and execute queries against the underlying metric system.
package transform

import (
	"fmt"
	"math"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
)

// A transform takes the list of values, other parameters, and the resolution of the query.
type transform func(function.EvaluationContext, api.Timeseries, []function.Value, time.Duration) ([]float64, error)

// ApplyTransform applies the given transform to the entire list of series.
func ApplyTransform(ctx function.EvaluationContext, list api.SeriesList, transformFunc transform, parameters []function.Value, resolution time.Duration) (api.SeriesList, error) {
	result := api.SeriesList{
		Series: make([]api.Timeseries, len(list.Series)),
	}
	var numResult []float64
	var err error
	for i, series := range list.Series {
		//TODO(cchandler): Modify the last parameter of this type to be an actual Resolution
		numResult, err = transformFunc(ctx, series, parameters, resolution)
		if err != nil {
			return api.SeriesList{}, err
		}
		result.Series[i] = api.Timeseries{
			Values: numResult,
			TagSet: series.TagSet,
		}
	}
	return result, nil
}

// Integral integrates a series whose values are "X per millisecond" to estimate "total X so far"
// if the series represents "X in this sampling interval" instead, then you should use transformCumulative.
func Integral(ctx function.EvaluationContext, series api.Timeseries, parameters []function.Value, resolution time.Duration) ([]float64, error) {
	values := series.Values
	result := make([]float64, len(values))
	integral := 0.0
	for i := range values {
		// Skip the 0th element since thats not technically part of our timerange
		if i == 0 {
			continue
		}

		if !math.IsNaN(values[i]) {
			integral += values[i]
		}
		result[i] = integral * resolution.Seconds()
	}
	return result, nil
}

// Cumulative computes the cumulative sum of the given values.
func Cumulative(ctx function.EvaluationContext, series api.Timeseries, parameters []function.Value, resolution time.Duration) ([]float64, error) {
	values := series.Values
	result := make([]float64, len(values))
	sum := 0.0
	for i := range values {
		// Skip the 0th element since thats not technically part of our timerange
		if i == 0 {
			continue
		}

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
func MapMaker(fun func(float64) float64) func(function.EvaluationContext, api.Timeseries, []function.Value, time.Duration) ([]float64, error) {
	return func(ctx function.EvaluationContext, series api.Timeseries, parameters []function.Value, resolution time.Duration) ([]float64, error) {
		values := series.Values
		result := make([]float64, len(values))
		for i := range values {
			result[i] = fun(values[i])
		}
		return result, nil
	}
}

// Default will replacing missing data (NaN) with the `default` value supplied as a parameter.
func Default(ctx function.EvaluationContext, series api.Timeseries, parameters []function.Value, resolution time.Duration) ([]float64, error) {
	values := series.Values
	defaultValue, err := parameters[0].ToScalar("default value")
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
func NaNKeepLast(ctx function.EvaluationContext, series api.Timeseries, parameters []function.Value, resolution time.Duration) ([]float64, error) {
	values := series.Values
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
func Bound(ctx function.EvaluationContext, series api.Timeseries, parameters []function.Value, resolution time.Duration) ([]float64, error) {
	values := series.Values
	lowerBound, err := parameters[0].ToScalar("lower bound")
	if err != nil {
		return nil, err
	}
	upperBound, err := parameters[1].ToScalar("upper bound")
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
func LowerBound(ctx function.EvaluationContext, series api.Timeseries, parameters []function.Value, resolution time.Duration) ([]float64, error) {
	values := series.Values
	lowerBound, err := parameters[0].ToScalar("lower bound")
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
func UpperBound(ctx function.EvaluationContext, series api.Timeseries, parameters []function.Value, resolution time.Duration) ([]float64, error) {
	values := series.Values
	upperBound, err := parameters[0].ToScalar("upper bound")
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
