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
	"strings"

	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
	"github.com/square/metrics/function/aggregate"
	"github.com/square/metrics/function/filter"
	"github.com/square/metrics/function/join"
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
	Compute     func(function.EvaluationContext, []function.Expression, []string) (function.Value, error)
}

func (f MetricFunction) Evaluate(
	context function.EvaluationContext, arguments []function.Expression, groupBy []string,
) (function.Value, error) {
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
		Compute: func(context function.EvaluationContext, args []function.Expression, groups []string) (function.Value, error) {
			leftValue, err := args[0].Evaluate(context)
			if err != nil {
				return nil, err
			}
			rightValue, err := args[1].Evaluate(context)
			if err != nil {
				return nil, err
			}
			leftList, err := leftValue.ToSeriesList(context.Timerange)
			if err != nil {
				return nil, err
			}
			rightList, err := rightValue.ToSeriesList(context.Timerange)
			if err != nil {
				return nil, err
			}

			joined := join.Join([]api.SeriesList{leftList, rightList})

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

			return function.SeriesListValue(api.SeriesList{
				Series:    result,
				Timerange: context.Timerange,
				Name:      fmt.Sprintf("(%s %s %s)", leftValue.GetName(), op, rightValue.GetName()),
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
		Groups:      true,
		Compute: func(context function.EvaluationContext, args []function.Expression, groups []string) (function.Value, error) {
			argument := args[0]
			value, err := argument.Evaluate(context)
			if err != nil {
				return nil, err
			}
			seriesList, err := value.ToSeriesList(context.Timerange)
			if err != nil {
				return nil, err
			}
			result := aggregate.AggregateBy(seriesList, aggregator, groups)
			groupNames := make([]string, len(groups))
			for i, group := range groups {
				groupNames[i] += group
			}
			if len(groups) == 0 {
				result.Name = fmt.Sprintf("%s(%s)", name, value.GetName())
			} else {
				result.Name = fmt.Sprintf("%s(%s group by %s)", name, value.GetName(), strings.Join(groupNames, ", "))
			}
			return function.SeriesListValue(result), nil
		},
	}
}

// MakeTransformMetircFunction takes a named transforming function `[float64], [value] => [float64]` and makes it into a MetricFunction.
func MakeTransformMetricFunction(name string, parameterCount int, transformer func([]float64, []function.Value, float64) ([]float64, error)) MetricFunction {
	return MetricFunction{
		Name:        name,
		MinArgument: parameterCount + 1,
		MaxArgument: parameterCount + 1,
		Compute: func(context function.EvaluationContext, args []function.Expression, groups []string) (function.Value, error) {
			// ApplyTransform(list api.SeriesList, transform transform, parameters []value) (api.SeriesList, error)
			listValue, err := args[0].Evaluate(context)
			if err != nil {
				return nil, err
			}
			list, err := listValue.ToSeriesList(context.Timerange)
			if err != nil {
				return nil, err
			}
			parameters := make([]function.Value, parameterCount)
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
			parameterNames := make([]string, len(parameters))
			for i, param := range parameters {
				parameterNames[i] = param.GetName()
			}
			if len(parameters) != 0 {
				result.Name = fmt.Sprintf("%s(%s, %s)", name, listValue.GetName(), strings.Join(parameterNames, ", "))
			} else {
				result.Name = fmt.Sprintf("%s(%s)", name, listValue.GetName())
			}
			return function.SeriesListValue(result), nil
		},
	}
}

var TimeshiftFunction = MetricFunction{
	Name:        "transform.timeshift",
	MinArgument: 2,
	MaxArgument: 2,
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

var MovingAverageFunction = MetricFunction{
	Name:        "transform.moving_average",
	MinArgument: 2,
	MaxArgument: 2,
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

var AliasFunction = MetricFunction{
	Name:        "transform.alias",
	MinArgument: 2,
	MaxArgument: 2,
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

func MakeFilterMetricFunction(name string, summary func([]float64) float64, ascending bool) MetricFunction {
	return MetricFunction{
		Name:        name,
		MinArgument: 2,
		MaxArgument: 2,
		Compute: func(context function.EvaluationContext, arguments []function.Expression, groups []string) (function.Value, error) {
			value, err := arguments[0].Evaluate(context)
			if err != nil {
				return nil, err
			}
			// The value must be a SeriesList.
			list, err := value.ToSeriesList(context.Timerange)
			if err != nil {
				return nil, err
			}
			countValue, err := arguments[1].Evaluate(context)
			if err != nil {
				return nil, err
			}
			countFloat, err := countValue.ToScalar()
			if err != nil {
				return nil, err
			}
			// Round to the nearest integer.
			count := int(countFloat + 0.5)
			if count < 0 {
				return nil, fmt.Errorf("expected positive count but got %d", count)
			}
			result := filter.FilterBy(list, count, summary, ascending)
			result.Name = fmt.Sprintf("%s(%s, %d)", name, value.GetName(), count)
			return function.SeriesListValue(result), nil
		},
	}
}
