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

package query

import (
	"fmt"
	"math"

	"github.com/square/metrics/api"
	"github.com/square/metrics/query/aggregate"
)

var functionRegistry = map[string]MetricFunction{}

func GetFunction(name string) (MetricFunction, bool) {
	fun, ok := functionRegistry[name]
	return fun, ok
}

func RegisterFunction(fun MetricFunction) error {
	_, ok := functionRegistry[fun.Name]
	if ok {
		return fmt.Errorf("function %s has already been registered", fun.Name)
	}
	functionRegistry[fun.Name] = fun
	return nil
}

func MustRegister(fun MetricFunction) {
	err := RegisterFunction(fun)
	if err != nil {
		panic(fmt.Sprintf("function %s in failed to register", fun.Name))
	}
}

type MetricFunction struct {
	Name        string
	MinArgument int
	MaxArgument int
	Groups      bool // Whether the function allows a 'group by' clause.
	Compute     func(EvaluationContext, []Expression, []string) (value, error)
}

func (f MetricFunction) Evaluate(
	context EvaluationContext,
	arguments []Expression,
	groupBy []string,
) (value, error) {
	// preprocessing
	length := len(arguments)
	if length < f.MinArgument || (f.MaxArgument != -1 && f.MaxArgument < length) {
		return nil, ArgumentLengthError{f.Name, f.MinArgument, f.MaxArgument, length}
	}
	if len(groupBy) > 0 && !f.Groups {
		return nil, fmt.Errorf("function %s doesn't allow a group-by clause", f.Name)
	}
	return f.Compute(context, arguments, groupBy)
}

func MakeOperatorMetricFunction(op string, operator func(float64, float64) float64) MetricFunction {
	return MetricFunction{
		Name:        op,
		MinArgument: 2,
		MaxArgument: 2,
		Compute: func(context EvaluationContext, args []Expression, groups []string) (value, error) {
			leftValue, err := args[0].Evaluate(context)
			if err != nil {
				return nil, err
			}
			rightValue, err := args[1].Evaluate(context)
			if err != nil {
				return nil, err
			}
			leftList, err := leftValue.toSeriesList(context.Timerange)
			if err != nil {
				return nil, err
			}
			rightList, err := rightValue.toSeriesList(context.Timerange)
			if err != nil {
				return nil, err
			}

			joined := join([]api.SeriesList{leftList, rightList})

			result := make([]api.Timeseries, len(joined.Rows))

			for i, row := range joined.Rows {
				left := row.Row[0]
				right := row.Row[1]
				array := make([]float64, len(left.Values))
				for j := 0; j < len(left.Values); j++ {
					array[j] = operator(left.Values[j], right.Values[j])
				}
				result[i] = api.Timeseries{array, row.TagSet}
			}

			return seriesListValue(api.SeriesList{
				Series:    result,
				Timerange: context.Timerange,
			}), nil
		},
	}
}

// MakeAggregateMetricFunction takes a named aggregating function `[float64] => float64` and makes it into a MetricFunction.
func MakeAggregateMetricFunction(name string, aggregator func([]float64) float64) MetricFunction {
	return MetricFunction{
		Name:        name,
		MinArgument: 1,
		MaxArgument: 1,
		Compute: func(context EvaluationContext, args []Expression, groups []string) (value, error) {
			argument := args[0]
			value, err := argument.Evaluate(context)
			if err != nil {
				return nil, err
			}
			seriesList, err := value.toSeriesList(context.Timerange)
			if err != nil {
				return nil, err
			}
			result := aggregate.AggregateBy(seriesList, aggregator, groups)
			return seriesListValue(result), nil
		},
	}
}

// MakeTransformMetircFunction takes a named transforming function `[float64], [value] => [float64]` and makes it into a MetricFunction.
func MakeTransformMetricFunction(name string, parameterCount int, transformer func([]float64, []value, float64) ([]float64, error)) MetricFunction {
	return MetricFunction{
		Name:        name,
		MinArgument: parameterCount + 1,
		MaxArgument: parameterCount + 1,
		Groups:      true,
		Compute: func(context EvaluationContext, args []Expression, groups []string) (value, error) {
			// ApplyTransform(list api.SeriesList, transform transform, parameters []value) (api.SeriesList, error)
			listValue, err := args[0].Evaluate(context)
			if err != nil {
				return nil, err
			}
			list, err := listValue.toSeriesList(context.Timerange)
			if err != nil {
				return nil, err
			}
			parameters := make([]value, parameterCount)
			for i := range parameters {
				parameters[i], err = args[i+1].Evaluate(context)
				if err != nil {
					return nil, err
				}
			}
			result, err := ApplyTransform(list, transformer, parameters)
			if err != nil {
				return nil, err
			}
			return seriesListValue(result), nil
		},
	}
}

var TimeshiftFunction = MetricFunction{
	Name:        "timeshift",
	MinArgument: 2,
	MaxArgument: 2,
	Compute: func(context EvaluationContext, arguments []Expression, groups []string) (value, error) {
		value, err := arguments[1].Evaluate(context)
		if err != nil {
			return nil, err
		}
		millis, err := toDuration(value)
		if err != nil {
			return nil, err
		}
		newContext := context
		newContext.Timerange = newContext.Timerange.Shift(millis)

		result, err := arguments[0].Evaluate(newContext)
		if err != nil {
			return nil, err
		}

		if seriesValue, ok := result.(seriesListValue); ok {
			seriesValue.Timerange = context.Timerange
			return seriesValue, nil
		}
		return result, nil
	},
}

var MovingAverageFunction = MetricFunction{
	Name:        "transform.moving_average",
	MinArgument: 2,
	MaxArgument: 2,
	Compute: func(context EvaluationContext, arguments []Expression, groups []string) (value, error) {
		// Applying a similar trick as did TimeshiftFunction. It fetches data prior to the start of the timerange.

		sizeValue, err := arguments[1].Evaluate(context)
		if err != nil {
			return nil, err
		}
		size, err := toDuration(sizeValue)
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
		list, err := listValue.toSeriesList(newContext.Timerange)
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
				if count == 0 && i-limit >= 0 {
					results[i-limit] = math.NaN()
				} else {
					results[i-limit] = sum / float64(count)
				}
			}
			list.Series[index].Values = results
		}
		return seriesListValue(list), nil
	},
}

func MakeFilterMetricFunction(name string, summary func([]float64) float64, ascending bool) MetricFunction {
	return MetricFunction{
		Name:        name,
		MinArgument: 2,
		MaxArgument: 2,
		Compute: func(context EvaluationContext, arguments []Expression, groups []string) (value, error) {
			value, err := arguments[0].Evaluate(context)
			if err != nil {
				return nil, err
			}
			// The value must be a SeriesList.
			list, err := value.toSeriesList(context.Timerange)
			if err != nil {
				return nil, err
			}
			countValue, err := arguments[1].Evaluate(context)
			if err != nil {
				return nil, err
			}
			countFloat, err := countValue.toScalar()
			if err != nil {
				return nil, err
			}
			// Round to the nearest integer.
			count := int(countFloat + 0.5)
			if count < 0 {
				return nil, fmt.Errorf("expected positive count but got %d", count)
			}
			result := FilterBy(list, count, summary, ascending)
			return seriesListValue(result), nil
		},
	}
}
