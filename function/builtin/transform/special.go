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

package transform

import (
	"fmt"
	"math"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
)

var Timeshift = function.MakeFunction(
	"transform.timeshift",
	func(expression function.Expression, duration time.Duration, context function.EvaluationContext) (function.Value, error) {
		newContext := context.WithTimerange(context.Timerange().Shift(duration))
		return expression.Evaluate(newContext)
	},
)

var MovingAverage = function.MakeFunction(
	"transform.moving_average",
	func(context function.EvaluationContext, listExpression function.Expression, size time.Duration) (api.SeriesList, error) {
		// Applying a similar trick as did TimeshiftFunction. It fetches data prior to the start of the timerange.
		limit := int(float64(size)/float64(context.Timerange().Resolution()) + 0.5) // Limit is the number of items to include in the average
		if limit < 1 {
			// At least one value must be included at all times
			limit = 1
		}

		timerange := context.Timerange()
		newTimerange := timerange.ExtendBefore(time.Duration(limit-1) * timerange.Resolution())
		newContext := context.WithTimerange(newTimerange)
		// The new context has a timerange which is extended beyond the query's.
		list, err := function.EvaluateToSeriesList(listExpression, newContext)
		if err != nil {
			return api.SeriesList{}, err
		}

		// Update each series in the list.
		for index, series := range list.Series {
			// The series will be given a (shorter) replaced list of values.
			results := make([]float64, context.Timerange().Slots())
			count := 0
			sum := 0.0
			for i := range series.Values {
				// Add the new element, if it isn't NaN.
				if !math.IsNaN(series.Values[i]) {
					sum += series.Values[i]
					count++
				}
				// Remove the oldest element, if it isn't NaN, and it's in range.
				// (e.g., if limit = 1, then this removes the previous element from the sum).
				if i >= limit && !math.IsNaN(series.Values[i-limit]) {
					sum -= series.Values[i-limit]
					count--
				}
				// Numerical error could (possibly) cause count == 0 but sum != 0.
				if i-limit+1 >= 0 {
					if count == 0 {
						results[i-limit+1] = math.NaN()
					} else {
						results[i-limit+1] = sum / float64(count)
					}
				}
			}
			list.Series[index].Values = results
		}
		return list, nil
	},
)

var ExponentialMovingAverage = function.MakeFunction(
	"transform.exponential_moving_average",
	func(context function.EvaluationContext, listExpression function.Expression, size time.Duration) (api.SeriesList, error) {
		// Applying a similar trick as did TimeshiftFunction. It fetches data prior to the start of the timerange.
		limit := int(float64(size)/float64(context.Timerange().Resolution()) + 0.5) // Limit is the number of items to include in the average
		if limit < 1 {
			// At least one value must be included at all times
			limit = 1
		}
		timerange := context.Timerange()
		newTimerange := timerange.ExtendBefore(time.Duration(limit) * timerange.Resolution())
		newContext := context.WithTimerange(newTimerange)

		// The new context has a timerange which is extended beyond the query's.
		list, err := function.EvaluateToSeriesList(listExpression, newContext)
		if err != nil {
			return api.SeriesList{}, err
		}

		alpha := math.Exp(math.Log(0.5) * context.Timerange().Resolution().Seconds() / size.Seconds())

		resultList := api.SeriesList{
			Series: make([]api.Timeseries, len(list.Series)),
		}
		for i := range resultList.Series {
			weight := 0.0
			sum := 0.0
			values := make([]float64, newTimerange.Slots())
			for t, y := range list.Series[i].Values {
				sum *= alpha
				weight *= alpha
				if !math.IsNaN(y) {
					sum += y
					weight++
				}
				values[t] = sum / weight
			}
			resultList.Series[i] = api.Timeseries{
				Values: values[newTimerange.Slots()-timerange.Slots():],
				TagSet: list.Series[i].TagSet,
			}
		}
		return resultList, nil
	},
)

// Derivative is special because it needs to get one extra data point to the left
// This transform estimates the "change per second" between the two samples (scaled consecutive difference)
var Derivative = function.MakeFunction(
	"transform.derivative",
	func(listExpression function.Expression, context function.EvaluationContext) (api.SeriesList, error) {
		newContext := context.WithTimerange(context.Timerange().ExtendBefore(context.Timerange().Resolution()))
		list, err := function.EvaluateToSeriesList(listExpression, newContext)
		if err != nil {
			return api.SeriesList{}, err
		}
		resultList := api.SeriesList{
			Series: make([]api.Timeseries, len(list.Series)),
		}
		for seriesIndex, series := range list.Series {
			newValues := make([]float64, len(series.Values)-1)
			for i := range series.Values {
				if i == 0 {
					continue
				}
				// Scaled difference
				newValues[i-1] = (series.Values[i] - series.Values[i-1]) / context.Timerange().Resolution().Seconds()
			}
			resultList.Series[seriesIndex] = api.Timeseries{
				Values: newValues,
				TagSet: series.TagSet, // TODO: verify that these are immutable
			}
		}
		return resultList, nil
	},
)

// Rate is special because it needs to get one extra data point to the left.
// This transform functions mostly like Derivative but bounds the result to be positive.
// Specifically this function is designed for strictly increasing counters that
// only decrease when reset to zero. That is, thie function returns consecutive
// differences which are at least 0, or math.Max of the newly reported value and 0
var Rate = function.MakeFunction(
	"transform.rate",
	func(listExpression function.Expression, context function.EvaluationContext) (api.SeriesList, error) {
		newContext := context.WithTimerange(context.Timerange().ExtendBefore(context.Timerange().Resolution()))
		list, err := function.EvaluateToSeriesList(listExpression, newContext)
		if err != nil {
			return api.SeriesList{}, err
		}
		resultList := api.SeriesList{
			Series: make([]api.Timeseries, len(list.Series)),
		}
		for seriesIndex, series := range list.Series {
			newValues := make([]float64, len(series.Values)-1)
			for i := range series.Values {
				if i == 0 {
					continue
				}
				// Scaled difference
				newValues[i-1] = (series.Values[i] - series.Values[i-1]) / context.Timerange().Resolution().Seconds()
				if newValues[i-1] < 0 {
					newValues[i-1] = 0
				}
				if i+1 < len(series.Values) && series.Values[i-1] > series.Values[i] && series.Values[i] <= series.Values[i+1] {
					// Downsampling may cause a drop from 1000 to 0 to look like [1000, 500, 0] instead of [1000, 1001, 0].
					// So we check the next, in addition to the previous.
					context.AddNote(fmt.Sprintf("Rate(%v): The underlying counter reset between %f, %f\n", series.TagSet, series.Values[i-1], series.Values[i]))
					// values[i] is our best approximatation of the delta between i-1 and i
					// Why? This should only be used on counters, so if v[i] - v[i-1] < 0 then
					// the counter has reset, and we know *at least* v[i] increments have happened
					newValues[i-1] = math.Max(series.Values[i], 0) / context.Timerange().Resolution().Seconds()
				}
			}
			resultList.Series[seriesIndex] = api.Timeseries{
				Values: newValues,
				TagSet: series.TagSet, // TODO: verify that these are immutable
			}
		}
		return resultList, nil
	},
)
