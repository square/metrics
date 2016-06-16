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
	// @@ can inline init.1.func1
	// @@ func literal escapes to heap
	// @@ func literal escapes to heap
	MustRegister(NewOperator("*", func(x float64, y float64) float64 { return x * y }))
	// @@ can inline init.1.func2
	// @@ func literal escapes to heap
	// @@ func literal escapes to heap
	MustRegister(NewOperator("/", func(x float64, y float64) float64 { return x / y }))
	// @@ can inline init.1.func3
	// @@ func literal escapes to heap
	// @@ func literal escapes to heap
	// Aggregates
	// @@ can inline init.1.func4
	// @@ func literal escapes to heap
	// @@ func literal escapes to heap
	MustRegister(NewAggregate("aggregate.max", aggregate.Max))
	MustRegister(NewAggregate("aggregate.min", aggregate.Min))
	// @@ NewAggregate("aggregate.max", aggregate.Max) escapes to heap
	MustRegister(NewAggregate("aggregate.mean", aggregate.Mean))
	// @@ NewAggregate("aggregate.min", aggregate.Min) escapes to heap
	MustRegister(NewAggregate("aggregate.sum", aggregate.Sum))
	// @@ NewAggregate("aggregate.mean", aggregate.Mean) escapes to heap
	MustRegister(NewAggregate("aggregate.total", aggregate.Total))
	// @@ NewAggregate("aggregate.sum", aggregate.Sum) escapes to heap
	MustRegister(NewAggregate("aggregate.count", aggregate.Count))
	// @@ NewAggregate("aggregate.total", aggregate.Total) escapes to heap
	// Transformations
	// @@ NewAggregate("aggregate.count", aggregate.Count) escapes to heap
	MustRegister(transform.Integral)
	MustRegister(transform.Cumulative)
	// @@ transform.Integral escapes to heap
	MustRegister(transform.NaNFill)
	// @@ transform.Cumulative escapes to heap
	MustRegister(transform.MapMaker("transform.abs", math.Abs))
	// @@ transform.NaNFill escapes to heap
	MustRegister(transform.MapMaker("transform.log", math.Log10))
	MustRegister(transform.NaNKeepLast)
	MustRegister(transform.Bound)
	// @@ transform.NaNKeepLast escapes to heap
	MustRegister(transform.LowerBound)
	// @@ transform.Bound escapes to heap
	MustRegister(transform.UpperBound)
	// @@ transform.LowerBound escapes to heap

	// @@ transform.UpperBound escapes to heap
	// Filter
	MustRegister(NewFilterCount("filter.highest_mean", aggregate.Mean, false))
	MustRegister(NewFilterCount("filter.highest_max", aggregate.Max, false))
	// @@ NewFilterCount("filter.highest_mean", aggregate.Mean, false) escapes to heap
	MustRegister(NewFilterCount("filter.highest_min", aggregate.Min, false))
	// @@ NewFilterCount("filter.highest_max", aggregate.Max, false) escapes to heap

	// @@ NewFilterCount("filter.highest_min", aggregate.Min, false) escapes to heap
	MustRegister(NewFilterCount("filter.lowest_mean", aggregate.Mean, true))
	MustRegister(NewFilterCount("filter.lowest_max", aggregate.Max, true))
	// @@ NewFilterCount("filter.lowest_mean", aggregate.Mean, true) escapes to heap
	MustRegister(NewFilterCount("filter.lowest_min", aggregate.Min, true))
	// @@ NewFilterCount("filter.lowest_max", aggregate.Max, true) escapes to heap

	// @@ NewFilterCount("filter.lowest_min", aggregate.Min, true) escapes to heap
	MustRegister(NewFilterThreshold("filter.mean_above", aggregate.Mean, false))
	MustRegister(NewFilterThreshold("filter.max_above", aggregate.Max, false))
	// @@ NewFilterThreshold("filter.mean_above", aggregate.Mean, false) escapes to heap
	MustRegister(NewFilterThreshold("filter.min_above", aggregate.Min, false))
	// @@ NewFilterThreshold("filter.max_above", aggregate.Max, false) escapes to heap

	// @@ NewFilterThreshold("filter.min_above", aggregate.Min, false) escapes to heap
	MustRegister(NewFilterThreshold("filter.mean_below", aggregate.Mean, true))
	MustRegister(NewFilterThreshold("filter.max_below", aggregate.Max, true))
	// @@ NewFilterThreshold("filter.mean_below", aggregate.Mean, true) escapes to heap
	MustRegister(NewFilterThreshold("filter.min_below", aggregate.Min, true))
	// @@ NewFilterThreshold("filter.max_below", aggregate.Max, true) escapes to heap

	// @@ NewFilterThreshold("filter.min_below", aggregate.Min, true) escapes to heap
	// Weird ones
	MustRegister(transform.Alias)
	MustRegister(transform.Derivative)
	// @@ transform.Alias escapes to heap
	MustRegister(transform.MovingAverage)
	// @@ transform.Derivative escapes to heap
	MustRegister(transform.ExponentialMovingAverage)
	// @@ transform.MovingAverage escapes to heap
	MustRegister(transform.Rate)
	// @@ transform.ExponentialMovingAverage escapes to heap
	MustRegister(transform.Timeshift)
	// @@ transform.Rate escapes to heap

	// @@ transform.Timeshift escapes to heap
	// Tags
	MustRegister(tag.DropFunction)
	MustRegister(tag.SetFunction)
	// @@ tag.DropFunction escapes to heap

	// @@ tag.SetFunction escapes to heap
	// Forecasting
	MustRegister(forecast.FunctionRollingMultiplicativeHoltWinters)
	MustRegister(forecast.FunctionAnomalyRollingMultiplicativeHoltWinters)
	// @@ forecast.FunctionRollingMultiplicativeHoltWinters escapes to heap
	MustRegister(forecast.FunctionRollingSeasonal)
	// @@ forecast.FunctionAnomalyRollingMultiplicativeHoltWinters escapes to heap
	MustRegister(forecast.FunctionAnomalyRollingSeasonal)
	// @@ forecast.FunctionRollingSeasonal escapes to heap
	MustRegister(forecast.FunctionForecastLinear)
	// @@ forecast.FunctionAnomalyRollingSeasonal escapes to heap

	// @@ forecast.FunctionForecastLinear escapes to heap
	MustRegister(forecast.FunctionDrop)

	// @@ forecast.FunctionDrop escapes to heap
	// Summary
	MustRegister(summary.Current)
	MustRegister(summary.Mean)
	// @@ summary.Current escapes to heap
	MustRegister(summary.Min)
	// @@ summary.Mean escapes to heap
	MustRegister(summary.Max)
	// @@ summary.Min escapes to heap
	MustRegister(summary.LastNotNaN)
	// @@ summary.Max escapes to heap
}

// @@ summary.LastNotNaN escapes to heap

// StandardRegistry of a functions available in MQE.
type StandardRegistry struct {
	mapping map[string]function.Function
}

var defaultRegistry = StandardRegistry{mapping: make(map[string]function.Function)}

func Default() StandardRegistry {
	return defaultRegistry
	// @@ can inline Default
}

// GetFunction returns a function associated with the given name, if it exists.
func (r StandardRegistry) GetFunction(name string) (function.Function, bool) {
	fun, ok := r.mapping[name]
	// @@ can inline StandardRegistry.GetFunction
	return fun, ok
}

func (r StandardRegistry) All() []string {
	result := make([]string, len(r.mapping))
	counter := 0
	// @@ make([]string, len(r.mapping)) escapes to heap
	// @@ make([]string, len(r.mapping)) escapes to heap
	// @@ make([]string, len(r.mapping)) escapes to heap
	for key := range r.mapping {
		result[counter] = key
		counter++
	}
	sort.Strings(result)
	return result
}

// Register a new function into the registry.
func (r StandardRegistry) Register(fun function.Function) error {
	// @@ leaking param: fun
	_, ok := r.mapping[fun.Name()]
	if ok {
		return fmt.Errorf("function %s has already been registered", fun.Name)
	}
	// @@ fun.Name escapes to heap
	// @@ fun.Name escapes to heap
	if fun.Name() == "" {
		return fmt.Errorf("empty function name")
	}
	r.mapping[fun.Name()] = fun
	return nil
}

// MustRegister adds a new metric function to the global function registry.
func MustRegister(fun function.Function) {
	// @@ leaking param: fun
	err := defaultRegistry.Register(fun)
	if err != nil {
		panic(fmt.Sprintf("function %s has failed to register", fun.Name))
	}
	// @@ fun.Name escapes to heap
	// @@ fun.Name escapes to heap
}

// Constructor Functions

// NewFilterCount creates a new instance of a filtering function with count limit.
func NewFilterCount(name string, summary func([]float64) float64, ascending bool) function.MetricFunction {
	// @@ leaking param: summary
	// @@ leaking param: name to result ~r3 level=0
	return function.MakeFunction(
		name,
		func(list api.SeriesList, countFloat float64, optionalDuration *time.Duration, timerange api.Timerange) (api.SeriesList, error) {
			// @@ leaking param content: list
			count := int(countFloat + 0.5)
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			if count < 0 {
				return api.SeriesList{}, fmt.Errorf("expected positive count but got %d", count)
			}
			// @@ count escapes to heap
			duration := timerange.Duration()
			if optionalDuration != nil {
				duration = *optionalDuration
			}
			if duration < 0 {
				return api.SeriesList{}, fmt.Errorf("expected positive recent duration but got %+v", duration)
			}
			// @@ duration escapes to heap
			return filter.FilterByRecent(list, count, summary, ascending, 1+int(duration/timerange.Resolution())), nil
		},
		// @@ inlining call to api.Timerange.Resolution
	)
}

// NewFilterThreshold creates a new instance of a filtering function.
func NewFilterThreshold(name string, summary func([]float64) float64, below bool) function.MetricFunction {
	// @@ leaking param: summary
	// @@ leaking param: name to result ~r3 level=0
	return function.MakeFunction(
		name,
		func(list api.SeriesList, threshold float64, optionalDuration *time.Duration, timerange api.Timerange) (api.SeriesList, error) {
			// @@ leaking param content: list
			duration := timerange.Duration()
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			if optionalDuration != nil {
				duration = *optionalDuration
			}
			if duration < 0 {
				return api.SeriesList{}, fmt.Errorf("expected positive recent duration but got %+v", duration)
			}
			// @@ duration escapes to heap
			return filter.FilterThresholdByRecent(list, threshold, summary, below, 1+int(duration/timerange.Resolution())), nil
		},
		// @@ inlining call to api.Timerange.Resolution
	)
}

// NewAggregate takes a named aggregating function `[float64] => float64` and makes it into a MetricFunction.
func NewAggregate(name string, aggregator func([]float64) float64) function.MetricFunction {
	// @@ leaking param: aggregator
	// @@ leaking param: name to result ~r2 level=0
	return function.MakeFunction(
		name,
		func(seriesList api.SeriesList, groups function.Groups) api.SeriesList {
			// @@ leaking param content: seriesList
			// @@ leaking param content: groups
			return aggregate.By(seriesList, aggregator, groups.List, groups.Collapses)
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
		},
	)
}

// NewOperator creates a new binary operator function.
// the binary operators display a natural join semantic.
func NewOperator(op string, operator func(float64, float64) float64) function.Function {
	// @@ leaking param: operator
	// @@ leaking param: op to result ~r2 level=0
	return function.MakeFunction(
		op,
		func(leftList api.SeriesList, rightList api.SeriesList, timerange api.Timerange) (api.SeriesList, error) {
			// @@ leaking param: leftList
			// @@ leaking param: rightList
			joined := join.Join([]api.SeriesList{leftList, rightList})
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap

			result := make([]api.Timeseries, len(joined.Rows))

			// @@ make([]api.Timeseries, len(joined.Rows)) escapes to heap
			// @@ make([]api.Timeseries, len(joined.Rows)) escapes to heap
			for i, row := range joined.Rows {
				left := row.Row[0]
				right := row.Row[1]
				array := make([]float64, len(left.Values))
				for j := 0; j < len(left.Values); j++ {
					// @@ make([]float64, len(left.Values)) escapes to heap
					// @@ make([]float64, len(left.Values)) escapes to heap
					array[j] = operator(left.Values[j], right.Values[j])
				}
				result[i] = api.Timeseries{Values: array, TagSet: row.TagSet}
			}

			return api.SeriesList{
				Series: result,
			}, nil
		},
	)
}

// @@ function.MakeFunction(op, func literal) escapes to heap
