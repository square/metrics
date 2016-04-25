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

	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
)

func transformEach(list api.SeriesList, transformation func([]float64) []float64) api.SeriesList {
	resultList := api.SeriesList{
		Series: make([]api.Timeseries, len(list.Series)),
	}
	for seriesIndex, series := range list.Series {
		resultList.Series[seriesIndex] = api.Timeseries{
			Values: transformation(series.Values),
			TagSet: series.TagSet, // TODO: verify that these are immutable
		}
	}
	return resultList
}

func mapper(list api.SeriesList, mapFunc func(float64) float64) api.SeriesList {
	return transformEach(list, func(values []float64) []float64 {
		result := make([]float64, len(values))
		for i := range result {
			result[i] = mapFunc(values[i])
		}
		return result
	})
}

// Integral integrates a series whose values are "X per millisecond" to estimate "total X so far"
// if the series represents "X in this sampling interval" instead, then you should use transformCumulative.
var Integral = function.MakeFunction(
	"transform.integral",
	func(list api.SeriesList, timerange api.Timerange) api.SeriesList {
		return transformEach(list, func(values []float64) []float64 {
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
				result[i] = integral * timerange.Resolution().Seconds()
			}
			return result
		})
	},
)

// Cumulative computes the cumulative sum of the given values.
var Cumulative = function.MakeFunction(
	"transform.cumulative",
	func(list api.SeriesList, timerange api.Timerange) api.SeriesList {
		return transformEach(list, func(values []float64) []float64 {
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
			return result
		})
	},
)

// MapMaker can be used to use a function as a transform, such as 'math.Abs' (or similar):
//  `MapMaker(math.Abs)` is a transform function which can be used, e.g. with ApplyTransform
// The name is used for error-checking purposes.

func MapMaker(name string, fun func(float64) float64) function.Function {
	return function.MakeFunction(
		name,
		func(list api.SeriesList, timerange api.Timerange) api.SeriesList {
			return transformEach(list, func(values []float64) []float64 {
				result := make([]float64, len(values))
				for i := range values {
					result[i] = fun(values[i])
				}
				return result
			})
		},
	)
}

// Default will replacing missing data (NaN) with the `default` value supplied as a parameter.
var Default = function.MakeFunction(
	"transform.default",
	func(list api.SeriesList, defaultValue float64) api.SeriesList {
		return mapper(list, func(value float64) float64 {
			if math.IsNaN(value) {
				return defaultValue
			}
			return value
		})
	},
)

// NaNKeepLast will replace missing NaN data with the data before it
var NaNKeepLast = function.MakeFunction(
	"transform.nan_keep_last",
	func(list api.SeriesList) api.SeriesList {
		return transformEach(list, func(values []float64) []float64 {
			result := make([]float64, len(values))
			for i := range result {
				result[i] = values[i]
				if math.IsNaN(values[i]) && i > 0 {
					result[i] = result[i-1]
				}
			}
			return result
		})
	},
)

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
var Bound = function.MakeFunction(
	"transform.bound",
	func(list api.SeriesList, lowerBound float64, upperBound float64) (api.SeriesList, error) {
		if lowerBound > upperBound {
			return api.SeriesList{}, boundError{lowerBound, upperBound}
		}
		return mapper(list, func(value float64) float64 {
			if value < lowerBound {
				return lowerBound
			}
			if value > upperBound {
				return upperBound
			}
			return value
		}), nil
	},
)

// LowerBound replaces values that fall below the given bound with the lower bound.
var LowerBound = function.MakeFunction(
	"transform.lower_bound",
	func(list api.SeriesList, lowerBound float64) (api.SeriesList, error) {
		return mapper(list, func(value float64) float64 {
			if value < lowerBound {
				return lowerBound
			}
			return value
		}), nil
	},
)

// UpperBound replaces values that fall below the given bound with the lower bound.
var UpperBound = function.MakeFunction(
	"transform.upper_bound",
	func(list api.SeriesList, upperBound float64) (api.SeriesList, error) {
		return mapper(list, func(value float64) float64 {
			if value > upperBound {
				return upperBound
			}
			return value
		}), nil
	},
)
