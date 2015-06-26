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

package transform

import (
	"fmt"
	"math"

	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
)

var Timeshift = function.MetricFunction{
	Name:         "transform.timeshift",
	MinArguments: 2,
	MaxArguments: 2,
	Compute: func(context function.EvaluationContext, arguments []function.Expression, groups []string) (function.Value, error) {
		value, err := arguments[1].Evaluate(context)
		if err != nil {
			return nil, err
		}
		millis, err := function.ToDuration(value)
		if err != nil {
			return nil, err
		}
		newContext := context
		newContext.Timerange = newContext.Timerange.Shift(millis)

		result, err := arguments[0].Evaluate(newContext)
		if err != nil {
			return nil, err
		}

		if seriesValue, ok := result.(function.SeriesListValue); ok {
			seriesValue.Timerange = context.Timerange
			seriesValue.Name = fmt.Sprintf("transform.timeshift(%s,%s)", result.GetName(), value.GetName())
			return seriesValue, nil
		}
		return result, nil
	},
}

var MovingAverage = function.MetricFunction{
	Name:         "transform.moving_average",
	MinArguments: 2,
	MaxArguments: 2,
	Compute: func(context function.EvaluationContext, arguments []function.Expression, groups []string) (function.Value, error) {
		// Applying a similar trick as did TimeshiftFunction. It fetches data prior to the start of the timerange.

		sizeValue, err := arguments[1].Evaluate(context)
		if err != nil {
			return nil, err
		}
		size, err := function.ToDuration(sizeValue)
		if err != nil {
			return nil, err
		}
		limit := int(float64(size)/float64(context.Timerange.Resolution()) + 0.5) // Limit is the number of items to include in the average
		if limit < 1 {
			// At least one value must be included at all times
			limit = 1
		}

		newContext := context
		timerange := context.Timerange
		newContext.Timerange, err = api.NewTimerange(timerange.Start()-int64(limit-1)*timerange.Resolution(), timerange.End(), timerange.Resolution())
		if err != nil {
			return nil, err
		}
		// The new context has a timerange which is extended beyond the query's.
		listValue, err := arguments[0].Evaluate(newContext)
		if err != nil {
			return nil, err
		}

		// This value must be a SeriesList.
		list, err := listValue.ToSeriesList(newContext.Timerange)
		if err != nil {
			return nil, err
		}

		// The timerange must be reverted.
		list.Timerange = context.Timerange

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
		list.Name = fmt.Sprintf("transform.moving_average(%s, %s)", listValue.GetName(), sizeValue.GetName())
		return function.SeriesListValue(list), nil
	},
}

var Alias = function.MetricFunction{
	Name:         "transform.alias",
	MinArguments: 2,
	MaxArguments: 2,
	Compute: func(context function.EvaluationContext, arguments []function.Expression, groups []string) (function.Value, error) {
		value, err := arguments[0].Evaluate(context)
		if err != nil {
			return nil, err
		}
		list, err := value.ToSeriesList(context.Timerange)
		if err != nil {
			return nil, err
		}
		nameValue, err := arguments[1].Evaluate(context)
		if err != nil {
			return nil, err
		}
		name, err := nameValue.ToString()
		if err != nil {
			return nil, err
		}
		list.Name = name
		return function.SeriesListValue(list), nil
	},
}
