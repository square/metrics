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

var Timeshift = function.MetricFunction{
	Name:         "transform.timeshift",
	MinArguments: 2,
	MaxArguments: 2,
	Compute: func(context function.EvaluationContext, arguments []function.Expression, groups function.Groups) (function.Value, error) {
		duration, err := function.EvaluateToDuration(arguments[1], context)
		if err != nil {
			return nil, err
		}

		newContext := context.WithTimerange(context.Timerange.Shift(duration))
		return arguments[0].Evaluate(newContext)
	},
}

var MovingAverage = function.MetricFunction{
	Name:         "transform.moving_average",
	MinArguments: 2,
	MaxArguments: 2,
	Compute: func(context function.EvaluationContext, arguments []function.Expression, groups function.Groups) (function.Value, error) {
		// Applying a similar trick as did TimeshiftFunction. It fetches data prior to the start of the timerange.

		size, err := function.EvaluateToDuration(arguments[1], context)
		if err != nil {
			return nil, err
		}
		limit := int(float64(size)/float64(context.Timerange.Resolution()) + 0.5) // Limit is the number of items to include in the average
		if limit < 1 {
			// At least one value must be included at all times
			limit = 1
		}

		timerange := context.Timerange
		newTimerange, err := api.NewSnappedTimerange(timerange.StartMillis()-int64(limit-1)*timerange.ResolutionMillis(), timerange.EndMillis(), timerange.ResolutionMillis())
		if err != nil {
			return nil, err
		}
		newContext := context.WithTimerange(newTimerange)
		// The new context has a timerange which is extended beyond the query's.
		list, err := function.EvaluateToSeriesList(arguments[0], newContext)
		if err != nil {
			return nil, err
		}

		// Update each series in the list.
		for index, series := range list.Series {
			// The series will be given a (shorter) replaced list of values.
			results := make([]float64, context.Timerange.Slots())
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
}

var ExponentialMovingAverage = function.MetricFunction{
	Name:         "transform.exponential_moving_average",
	MinArguments: 2,
	MaxArguments: 2,
	Compute: func(context function.EvaluationContext, arguments []function.Expression, groups function.Groups) (function.Value, error) {
		// Applying a similar trick as did TimeshiftFunction. It fetches data prior to the start of the timerange.

		size, err := function.EvaluateToDuration(arguments[1], context)
		if err != nil {
			return nil, err
		}
		limit := int(float64(size)/float64(context.Timerange.Resolution()) + 0.5) // Limit is the number of items to include in the average
		if limit < 1 {
			// At least one value must be included at all times
			limit = 1
		}

		timerange := context.Timerange
		newTimerange, err := api.NewSnappedTimerange(timerange.StartMillis()-int64(limit-1)*timerange.ResolutionMillis(), timerange.EndMillis(), timerange.ResolutionMillis())
		if err != nil {
			return nil, err
		}

		newContext := context.WithTimerange(newTimerange)

		// The new context has a timerange which is extended beyond the query's.
		list, err := function.EvaluateToSeriesList(arguments[0], newContext)
		if err != nil {
			return nil, err
		}

		// How many "ticks" are there in "size"?
		// size / resolution
		// alpha is a parameter such that
		// alpha^ticks = 1/2
		// so, alpha = exp(log(1/2) / ticks)
		alpha := math.Exp(math.Log(0.5) * float64(context.Timerange.Resolution()) / float64(size))

		// Update each series in the list.
		for index, series := range list.Series {
			// The series will be given a (shorter) replaced list of values.
			results := make([]float64, context.Timerange.Slots())
			weight := 0.0
			sum := 0.0
			for i := range series.Values {
				weight *= alpha
				sum *= alpha
				if !math.IsNaN(series.Values[i]) {
					weight++
					sum += series.Values[i]
				}
				results[i-limit+1] = sum / weight
			}
			list.Series[index].Values = results
		}
		return list, nil
	},
}

var Alias = function.MetricFunction{
	Name:         "transform.alias",
	MinArguments: 2,
	MaxArguments: 2,
	Compute: func(context function.EvaluationContext, arguments []function.Expression, groups function.Groups) (function.Value, error) {
		// TODO: delete this function
		context.EvaluationNotes.AddNote("transform.alias is deprecated")
		return arguments[0].Evaluate(context)
	},
}

// Derivative is special because it needs to get one extra data point to the left
// This transform estimates the "change per second" between the two samples (scaled consecutive difference)
var Derivative = newDerivativeBasedTransform("derivative", derivative)

func derivative(ctx function.EvaluationContext, series api.Timeseries, parameters []function.Value, resolution time.Duration) ([]float64, error) {
	values := series.Values
	result := make([]float64, len(values)-1)
	for i := range values {
		if i == 0 {
			continue
		}
		// Scaled difference
		result[i-1] = (values[i] - values[i-1]) / resolution.Seconds()
	}
	return result, nil
}

// Rate is special because it needs to get one extra data point to the left.
// This transform functions mostly like Derivative but bounds the result to be positive.
// Specifically this function is designed for strictly increasing counters that
// only decrease when reset to zero. That is, thie function returns consecutive
// differences which are at least 0, or math.Max of the newly reported value and 0
var Rate = newDerivativeBasedTransform("rate", rate)

func rate(ctx function.EvaluationContext, series api.Timeseries, parameters []function.Value, resolution time.Duration) ([]float64, error) {
	values := series.Values
	result := make([]float64, len(values)-1)
	for i := range values {
		if i == 0 {
			continue
		}
		// Scaled difference
		result[i-1] = (values[i] - values[i-1]) / resolution.Seconds()
		if result[i-1] < 0 {
			result[i-1] = 0
		}
		if i+1 < len(values) && values[i-1] > values[i] && values[i] <= values[i+1] {
			// Downsampling may cause a drop from 1000 to 0 to look like [1000, 500, 0] instead of [1000, 1001, 0].
			// So we check the next, in addition to the previous.
			ctx.EvaluationNotes.AddNote(fmt.Sprintf("Rate(%v): The underlying counter reset between %f, %f\n", series.TagSet, values[i-1], values[i]))
			// values[i] is our best approximatation of the delta between i-1 and i
			// Why? This should only be used on counters, so if v[i] - v[i-1] < 0 then
			// the counter has reset, and we know *at least* v[i] increments have happened
			result[i-1] = math.Max(values[i], 0) / resolution.Seconds()
		}
	}
	return result, nil
}

// newDerivativeBasedTransform returns a function.MetricFunction that performs
// a delta between two data points. The transform parameter is a function of type
// transform is expected to return an array of values whose length is 1 less
// than the given series
func newDerivativeBasedTransform(name string, transformer transform) function.MetricFunction {
	return function.MetricFunction{
		Name:         "transform." + name,
		MinArguments: 1,
		MaxArguments: 1,
		Compute: func(context function.EvaluationContext, arguments []function.Expression, groups function.Groups) (function.Value, error) {

			// Calcuate the new timerange to include one extra point to the left.
			newContext := context.WithTimerange(context.Timerange.ExtendBefore(context.Timerange.Resolution()))

			// The new context has a timerange which is extended beyond the query's.
			list, err := function.EvaluateToSeriesList(arguments[0], newContext)
			if err != nil {
				return nil, err
			}

			//Apply the original context to the transform even though the list
			//will include one additional data point.
			result, err := ApplyTransform(context, list, transformer, []function.Value{}, context.Timerange.Resolution())
			if err != nil {
				return nil, err
			}

			// Validate our series are the correct length
			for i := range result.Series {
				if len(result.Series[i].Values) != len(list.Series[i].Values)-1 {
					panic(fmt.Sprintf("Expected transform to return %d values, received %d", len(list.Series[i].Values)-1, len(result.Series[i].Values)))
				}
			}
			return result, nil
		},
	}
}
