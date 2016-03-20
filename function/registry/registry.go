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

package registry

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
	"github.com/square/metrics/function/aggregate"
	"github.com/square/metrics/function/filter"
	"github.com/square/metrics/function/forecast"
	"github.com/square/metrics/function/join"
	"github.com/square/metrics/function/tag"
	"github.com/square/metrics/function/transform"
)

func init() {
	// Arithmetic operators
	MustRegister(NewOperator("+", func(x float64, y float64) float64 { return x + y }))
	MustRegister(NewOperator("-", func(x float64, y float64) float64 { return x - y }))
	MustRegister(NewOperator("*", func(x float64, y float64) float64 { return x * y }))
	MustRegister(NewOperator("/", func(x float64, y float64) float64 { return x / y }))
	// Aggregates
	MustRegister(NewAggregate("aggregate.max", aggregate.Max))
	MustRegister(NewAggregate("aggregate.min", aggregate.Min))
	MustRegister(NewAggregate("aggregate.mean", aggregate.Mean))
	MustRegister(NewAggregate("aggregate.sum", aggregate.Sum))
	MustRegister(NewAggregate("aggregate.total", aggregate.Total))
	MustRegister(NewAggregate("aggregate.count", aggregate.Count))
	// Transformations
	MustRegister(NewTransform("transform.integral", 0, transform.Integral))
	MustRegister(NewTransform("transform.cumulative", 0, transform.Cumulative))
	MustRegister(NewTransform("transform.nan_fill", 1, transform.Default))
	MustRegister(NewTransform("transform.abs", 0, transform.MapMaker(math.Abs)))
	MustRegister(NewTransform("transform.log", 0, transform.MapMaker(math.Log10)))
	MustRegister(NewTransform("transform.nan_keep_last", 0, transform.NaNKeepLast))
	MustRegister(NewTransform("transform.bound", 2, transform.Bound))
	MustRegister(NewTransform("transform.lower_bound", 1, transform.LowerBound))
	MustRegister(NewTransform("transform.upper_bound", 1, transform.UpperBound))
	// Filter
	MustRegister(NewFilter("filter.highest_mean", aggregate.Mean, false))
	MustRegister(NewFilter("filter.lowest_mean", aggregate.Mean, true))
	MustRegister(NewFilter("filter.highest_max", aggregate.Max, false))
	MustRegister(NewFilter("filter.lowest_max", aggregate.Max, true))
	MustRegister(NewFilter("filter.highest_min", aggregate.Min, false))
	MustRegister(NewFilter("filter.lowest_min", aggregate.Min, true))
	// Filter Recent
	MustRegister(NewFilterRecent("filter.recent_highest_mean", aggregate.Mean, false))
	MustRegister(NewFilterRecent("filter.recent_lowest_mean", aggregate.Mean, true))
	MustRegister(NewFilterRecent("filter.recent_highest_max", aggregate.Max, false))
	MustRegister(NewFilterRecent("filter.recent_lowest_max", aggregate.Max, true))
	MustRegister(NewFilterRecent("filter.recent_highest_min", aggregate.Min, false))
	MustRegister(NewFilterRecent("filter.recent_lowest_min", aggregate.Min, true))
	// Weird ones
	MustRegister(transform.Alias)
	MustRegister(transform.Derivative)
	MustRegister(transform.MovingAverage)
	MustRegister(transform.ExponentialMovingAverage)
	MustRegister(transform.Rate)
	MustRegister(transform.Timeshift)
	// Tags
	MustRegister(tag.DropFunction)
	MustRegister(tag.SetFunction)
	// Forecasting
	MustRegister(forecast.FunctionRollingMultiplicativeHoltWinters)
	MustRegister(forecast.FunctionAnomalyRollingMultiplicativeHoltWinters)
	MustRegister(forecast.FunctionRollingSeasonal)
	MustRegister(forecast.FunctionAnomalyRollingSeasonal)
	MustRegister(forecast.FunctionForecastLinear)

	MustRegister(forecast.FunctionDrop)
}

// StandardRegistry of a functions available in MQE.
type StandardRegistry struct {
	mapping map[string]function.MetricFunction
}

var defaultRegistry = StandardRegistry{mapping: make(map[string]function.MetricFunction)}

func Default() StandardRegistry {
	return defaultRegistry
}

// GetFunction returns a function associated with the given name, if it exists.
func (r StandardRegistry) GetFunction(name string) (function.MetricFunction, bool) {
	fun, ok := r.mapping[name]
	return fun, ok
}

func (r StandardRegistry) All() []string {
	result := make([]string, len(r.mapping))
	counter := 0
	for key := range r.mapping {
		result[counter] = key
		counter++
	}
	sort.Strings(result)
	return result
}

// Register a new function into the registry.
func (r StandardRegistry) Register(fun function.MetricFunction) error {
	_, ok := r.mapping[fun.Name]
	if ok {
		return fmt.Errorf("function %s has already been registered", fun.Name)
	}
	if fun.Compute == nil {
		return fmt.Errorf("function %s has no Compute() field", fun.Name)
	}
	if fun.Name == "" {
		return fmt.Errorf("empty function name")
	}
	r.mapping[fun.Name] = fun
	return nil
}

// MustRegister adds a new metric function to the global function registry.
func MustRegister(fun function.MetricFunction) {
	err := defaultRegistry.Register(fun)
	if err != nil {
		panic(fmt.Sprintf("function %s has failed to register", fun.Name))
	}
}

// Constructor Functions

// NewFilter creates a new instance of a filtering function.
func NewFilter(name string, summary func([]float64) float64, ascending bool) function.MetricFunction {
	return function.MetricFunction{
		Name:         name,
		MinArguments: 2,
		MaxArguments: 2,
		Compute: func(context function.EvaluationContext, arguments []function.Expression, groups function.Groups) (function.Value, error) {
			list, err := function.EvaluateToSeriesList(arguments[0], context)
			if err != nil {
				return nil, err
			}
			countFloat, err := function.EvaluateToScalar(arguments[1], context)
			if err != nil {
				return nil, err
			}
			// Round to the nearest integer.
			count := int(countFloat + 0.5)
			if count < 0 {
				return nil, fmt.Errorf("expected positive count but got %d", count)
			}
			result := filter.FilterBy(list, count, summary, ascending)
			return result, nil
		},
	}
}

// NewFilterRecent creates a new instance of a recent-filtering function.
func NewFilterRecent(name string, summary func([]float64) float64, ascending bool) function.MetricFunction {
	return function.MetricFunction{
		Name:         name,
		MinArguments: 3,
		MaxArguments: 3,
		Compute: func(context function.EvaluationContext, arguments []function.Expression, groups function.Groups) (function.Value, error) {
			list, err := function.EvaluateToSeriesList(arguments[0], context)
			if err != nil {
				return nil, err
			}
			countFloat, err := function.EvaluateToScalar(arguments[1], context)
			if err != nil {
				return nil, err
			}
			// Round to the nearest integer.
			count := int(countFloat + 0.5)
			if count < 0 {
				return nil, fmt.Errorf("expected positive count but got %d", count)
			}
			duration, err := function.EvaluateToDuration(arguments[2], context)
			if err != nil {
				return nil, err
			}
			result := filter.FilterRecentBy(list, count, summary, ascending, duration)
			return result, nil
		},
	}
}

// NewAggregate takes a named aggregating function `[float64] => float64` and makes it into a MetricFunction.
func NewAggregate(name string, aggregator func([]float64) float64) function.MetricFunction {
	return function.MetricFunction{
		Name:          name,
		MinArguments:  1,
		MaxArguments:  1,
		AllowsGroupBy: true,
		Compute: func(context function.EvaluationContext, args []function.Expression, groups function.Groups) (function.Value, error) {
			argument := args[0]
			seriesList, err := function.EvaluateToSeriesList(argument, context)
			if err != nil {
				return nil, err
			}
			return aggregate.AggregateBy(seriesList, aggregator, groups.List, groups.Collapses), nil
		},
	}
}

// NewTransform takes a named transforming function `[float64], [value] => [float64]` and makes it into a MetricFunction.
func NewTransform(name string, parameterCount int, transformer func(function.EvaluationContext, api.Timeseries, []function.Value, float64) ([]float64, error)) function.MetricFunction {
	return function.MetricFunction{
		Name:         name,
		MinArguments: parameterCount + 1,
		MaxArguments: parameterCount + 1,
		Compute: func(context function.EvaluationContext, arguments []function.Expression, groups function.Groups) (function.Value, error) {
			list, err := function.EvaluateToSeriesList(arguments[0], context)
			if err != nil {
				return nil, err
			}
			parameters := make([]function.Value, parameterCount)
			for i := range parameters {
				parameters[i], err = arguments[i+1].Evaluate(context)
				if err != nil {
					return nil, err
				}
			}
			return transform.ApplyTransform(context, list, transformer, parameters)
		},
	}
}

// NewOperator creates a new binary operator function.
// the binary operators display a natural join semantic.
func NewOperator(op string, operator func(float64, float64) float64) function.MetricFunction {
	return function.MetricFunction{
		Name:         op,
		MinArguments: 2,
		MaxArguments: 2,
		Compute: func(context function.EvaluationContext, args []function.Expression, groups function.Groups) (function.Value, error) {
			evaluated, err := function.EvaluateMany(context, args)
			if err != nil {
				return nil, err
			}
			leftValue := evaluated[0]
			rightValue := evaluated[1]

			leftScalar, errLeftScalar := leftValue.ToScalar("left operand scalar")
			rightScalar, errRightScalar := rightValue.ToScalar("right operand scalar")

			leftDuration, errLeftDuration := leftValue.ToDuration("left operand duration")
			rightDuration, errRightDuration := rightValue.ToDuration("right operand duration")

			// Handles the two-scalar case.
			if errLeftScalar == nil && errRightScalar == nil {
				return function.ScalarValue(operator(leftScalar, rightScalar)), nil
			}

			if (errLeftScalar == nil || errLeftDuration == nil) && (errRightScalar == nil || errRightDuration == nil) {
				// both are number-types
				switch op {
				case "+", "-":
					// Input units must match, output has the same units.
					if errLeftDuration == nil && errRightDuration == nil {
						// Convert to floats, then convert back.
						return function.DurationValue(time.Duration(operator(float64(leftDuration), float64(rightDuration)))), nil
					}
				case "*":
					// One of the input units must be a scalar, output has units of the other.
					if errLeftScalar == nil && errRightDuration == nil {
						return function.DurationValue(time.Duration(operator(leftScalar, float64(rightDuration)))), nil
					}
					if errLeftDuration == nil && errRightScalar == nil {
						return function.DurationValue(time.Duration(operator(float64(leftDuration), rightScalar))), nil
					}
				case "/":
					// Input units must match, output is scalar.
					if errLeftDuration == nil && errRightDuration == nil {
						return function.ScalarValue(operator(float64(leftDuration), float64(rightDuration))), nil
					}
					// Or, the right is a scalar
					if errLeftDuration == nil && errRightScalar == nil {
						return function.DurationValue(time.Duration(operator(float64(leftDuration), rightScalar))), nil
					}
				}
			}

			leftList, err := leftValue.ToSeriesList(context.Timerange, args[0].QueryString())
			if err != nil {
				return nil, err
			}
			rightList, err := rightValue.ToSeriesList(context.Timerange, args[1].QueryString())
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
				result[i] = api.Timeseries{Values: array, TagSet: row.TagSet}
			}

			return api.SeriesList{
				Series:    result,
				Timerange: context.Timerange,
			}, nil
		},
	}
}
