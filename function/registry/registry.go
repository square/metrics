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
	"github.com/square/metrics/function/summary"
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
	MustRegister(transform.Integral)
	MustRegister(transform.Cumulative)
	MustRegister(transform.Default)
	MustRegister(transform.MapMaker("transform.abs", math.Abs))
	MustRegister(transform.MapMaker("transform.log", math.Log10))
	MustRegister(transform.NaNKeepLast)
	MustRegister(transform.Bound)
	MustRegister(transform.LowerBound)
	MustRegister(transform.UpperBound)

	// Filter
	MustRegister(NewFilterCount("filter.highest_mean", aggregate.Mean, false))
	MustRegister(NewFilterCount("filter.highest_max", aggregate.Max, false))
	MustRegister(NewFilterCount("filter.highest_min", aggregate.Min, false))

	MustRegister(NewFilterCount("filter.lowest_mean", aggregate.Mean, true))
	MustRegister(NewFilterCount("filter.lowest_max", aggregate.Max, true))
	MustRegister(NewFilterCount("filter.lowest_min", aggregate.Min, true))

	MustRegister(NewFilterThreshold("filter.mean_above", aggregate.Mean, false))
	MustRegister(NewFilterThreshold("filter.max_above", aggregate.Max, false))
	MustRegister(NewFilterThreshold("filter.min_above", aggregate.Min, false))

	MustRegister(NewFilterThreshold("filter.mean_below", aggregate.Mean, true))
	MustRegister(NewFilterThreshold("filter.max_below", aggregate.Max, true))
	MustRegister(NewFilterThreshold("filter.min_below", aggregate.Min, true))

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

	// Summary
	MustRegister(summary.Mean)
	MustRegister(summary.Min)
	MustRegister(summary.Max)
	MustRegister(summary.Current)
	MustRegister(summary.LastNotNaN)

}

// StandardRegistry of a functions available in MQE.
type StandardRegistry struct {
	mapping map[string]function.Function
}

var defaultRegistry = StandardRegistry{mapping: make(map[string]function.Function)}

func Default() StandardRegistry {
	return defaultRegistry
}

// GetFunction returns a function associated with the given name, if it exists.
func (r StandardRegistry) GetFunction(name string) (function.Function, bool) {
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
func (r StandardRegistry) Register(fun function.Function) error {
	_, ok := r.mapping[fun.Name()]
	if ok {
		return fmt.Errorf("function %s has already been registered", fun.Name())
	}
	if fun.Name() == "" {
		return fmt.Errorf("empty function name")
	}
	r.mapping[fun.Name()] = fun
	return nil
}

// MustRegister adds a new metric function to the global function registry.
func MustRegister(fun function.Function) {
	err := defaultRegistry.Register(fun)
	if err != nil {
		panic(fmt.Sprintf("function %s has failed to register: %s", fun.Name(), err.Error()))
	}
}

// Constructor Functions

// NewFilter creates a new instance of a filtering function.
func NewFilterCount(name string, summary func([]float64) float64, ascending bool) function.MetricFunction {
	return function.MakeFunction(
		name,
		func(list api.SeriesList, countFloat float64, optionalDuration *time.Duration, timerange api.Timerange) (api.SeriesList, error) {
			count := int(countFloat + 0.5)
			if count < 0 {
				return api.SeriesList{}, fmt.Errorf("expected positive count but got %d", count)
			}
			duration := timerange.Duration()
			if optionalDuration != nil {
				if *optionalDuration < 0 {
					return api.SeriesList{}, fmt.Errorf("expected a positive duration but got %+v", *optionalDuration)
				}
				duration = *optionalDuration
			}
			return filter.FilterByRecent(list, count, summary, ascending, duration), nil
		},
	)
}

// NewFilterThreshold creates a new instance of a filtering function.
func NewFilterThreshold(name string, summary func([]float64) float64, ascending bool) function.MetricFunction {
	return function.MakeFunction(
		name,
		func(list api.SeriesList, threshold float64, optionalDuration *time.Duration, timerange api.Timerange) (api.SeriesList, error) {
			duration := timerange.Duration()
			if optionalDuration != nil {
				if *optionalDuration < 0 {
					return api.SeriesList{}, fmt.Errorf("expected a positive duration but got %+v", *optionalDuration)
				}
				duration = *optionalDuration
			}
			return filter.FilterThresholdByRecent(list, threshold, summary, ascending, duration), nil
		},
	)
}

// NewAggregate takes a named aggregating function `[float64] => float64` and makes it into a MetricFunction.
func NewAggregate(name string, aggregator func([]float64) float64) function.MetricFunction {
	return function.MakeFunction(
		name,
		func(seriesList api.SeriesList, groups function.Groups) api.SeriesList {
			return aggregate.By(seriesList, aggregator, groups.List, groups.Collapses)
		},
	)
}

// NewOperator creates a new binary operator function.
// the binary operators display a natural join semantic.
func NewOperator(op string, operator func(float64, float64) float64) function.Function {
	return function.MakeFunction(
		op,
		func(leftList api.SeriesList, rightList api.SeriesList, timerange api.Timerange) (api.SeriesList, error) {
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
				Timerange: timerange,
			}, nil
		},
	)
}
