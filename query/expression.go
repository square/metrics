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
	"errors"
	"fmt"

	"github.com/square/metrics/api"
	"github.com/square/metrics/query/aggregate"
)

// Expression is a piece of code, which can be evaluated in a given
// EvaluationContext. EvaluationContext must never be changed in an Evalute().
//
// The contract of Expressions is that leaf nodes must sample a resulting
// timeseries according to the resolution specified in its EvaluationContext's
// Timerange. Internal nodes may assume that results from evaluating child
// Expressions correspond to the timerange in the current EvaluationContext.
type Expression interface {
	// Evaluate the given expression.
	Evaluate(context EvaluationContext) (value, error)
}

func evaluateToSeriesList(e Expression, context EvaluationContext) (api.SeriesList, error) {
	value, err := e.Evaluate(context)
	if err != nil {
		return api.SeriesList{}, err
	}
	return value.toSeriesList(context.Timerange)
}

// EvaluationContext is the central piece of logic, providing
// helper funcions & varaibles to evaluate a given piece of
// metrics query.
// * Contains Backend object, which can be used to fetch data
// from the backend system.s
// * Contains current timerange being queried for - this can be
// changed by say, application of time shift function.
type EvaluationContext struct {
	Backend      api.Backend      // Backend to fetch data from
	API          api.API          // Api to obtain metadata from
	Timerange    api.Timerange    // Timerange to fetch data from
	SampleMethod api.SampleMethod // SampleMethod to use when up/downsampling to match the requested resolution
	Predicate    api.Predicate
}

// Implementations
// ===============

// Generates a Timeseries from the encapsulated scalar.
func (expr scalarExpression) Evaluate(context EvaluationContext) (value, error) {
	return scalarValue(expr.value), nil
}

func (expr *metricFetchExpression) Evaluate(context EvaluationContext) (value, error) {
	// Merge predicates appropriately
	var predicate api.Predicate
	if context.Predicate == nil && expr.predicate == nil {
		predicate = api.TruePredicate
	} else if context.Predicate == nil {
		predicate = expr.predicate
	} else if expr.predicate == nil {
		predicate = context.Predicate
	} else {
		predicate = &andPredicate{[]api.Predicate{expr.predicate, context.Predicate}}
	}

	metricTagSets, err := context.API.GetAllTags(api.MetricKey(expr.metricName))
	if err != nil {
		return nil, err
	}
	filtered := applyPredicates(metricTagSets, predicate)

	resultSeries := []api.Timeseries{}
	for _, ts := range filtered {
		result, err := context.Backend.FetchSingleSeries(api.FetchSeriesRequest{
			api.TaggedMetric{api.MetricKey(expr.metricName), ts}, context.SampleMethod, context.Timerange,
			context.API,
		})
		if err != nil {
			return nil, err
		}
		resultSeries = append(resultSeries, result)
	}

	return seriesListValue(api.SeriesList{
		Series:    resultSeries,
		Timerange: context.Timerange,
	}), nil
}

func (expr *functionExpression) Evaluate(context EvaluationContext) (value, error) {
	name := expr.functionName
	length := len(expr.arguments)
	values, err := evaluateExpressions(context, expr.arguments)

	if err != nil {
		return nil, err
	}

	operatorMap := map[string]func(float64, float64) float64{
		"+": func(x, y float64) float64 { return x + y },
		"-": func(x, y float64) float64 { return x - y },
		"*": func(x, y float64) float64 { return x * y },
		"/": func(x, y float64) float64 { return x / y },
	}

	if operator, ok := operatorMap[name]; ok {
		// Evaluation of a binary operator:
		// Verify that exactly 2 arguments are given.
		if length != 2 {
			return nil, errors.New(fmt.Sprintf("Function `%s` expects 2 operands but received %d (%+v)", name, len(expr.arguments), expr.arguments))
		}
		return evaluateBinaryOperation(context, name, values[0], values[1], operator)
	}

	if aggregator, ok := aggregate.GetAggregate(name); ok {
		// Verify that exactly 1 argument is given.
		if length != 1 {
			return nil, errors.New(fmt.Sprintf("Function `%s` expects 1 argument but received %d (%+v)", name, len(expr.arguments), expr.arguments))
		}
		value := values[0]
		list, err := value.toSeriesList(context.Timerange)
		if err != nil {
			return nil, err
		}
		series := aggregate.AggregateBy(list, aggregator, expr.groupBy)
		return seriesListValue(series), nil
	}

	if transform, ok := GetTransformation(name); ok {
		//Verify that at least one argument is given.
		if length == 0 {
			return nil, errors.New(fmt.Sprintf("Function `%s` expects at least 1 argument but was given 0", name))
		}
		first, err := expr.arguments[0].Evaluate(context)
		if err != nil {
			return nil, err
		}
		list, err := first.toSeriesList(context.Timerange)
		if err != nil {
			return nil, err
		}
		// Evaluate all the other parameters:
		rest := values[1:]
		series, err := ApplyTransform(list, transform, rest)
		if err != nil {
			return nil, err
		}
		return seriesListValue(series), nil
	}

	if name == "timeshift" {
		// A timeshift performs a modification to the evaluation context.
		// In the future, it may be one of a class of functions which performs a similar modification.
		// A timeshift has two parameters: its first (which it evaluates), and its second (the time offset).
		if len(expr.arguments) != 2 {
			return nil, errors.New(fmt.Sprintf("Function `timeshift` expects 2 parameters but is given %d (%+v)", len(expr.arguments), expr.arguments))
		}
		shift := values[1]
		duration, err := toDuration(shift)
		if err != nil {
			return nil, err
		}
		newContext := context
		newContext.Timerange = newContext.Timerange.Shift(int64(duration))
		value := values[0]
		if series, ok := value.(seriesListValue); ok {
			// If it's a series, then we need to reset its timerange to the original.
			// Although it's questionably useful to use timeshifting for a non-series,
			// it seems sensible to allow it anyway.
			series.Timerange = context.Timerange
		}
		return value, nil
	}

	return nil, errors.New(fmt.Sprintf("unknown function name `%s`", name))
}

// Auxiliary functions
// ===================

// evaluateBinaryOperation applies an arbirary binary operation to two
// Expressions.
func evaluateBinaryOperation(
	context EvaluationContext,
	functionName string,
	leftValue value,
	rightValue value,
	evaluate func(float64, float64) float64,
) (value, error) {

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
			array[j] = evaluate(left.Values[j], right.Values[j])
		}
		result[i] = api.Timeseries{array, row.TagSet}
	}

	return seriesListValue(api.SeriesList{
		Series:    result,
		Timerange: context.Timerange,
	}), nil
}

func applyPredicates(tagSets []api.TagSet, predicate api.Predicate) []api.TagSet {
	output := []api.TagSet{}
	for _, ts := range tagSets {
		if predicate.Apply(ts) {
			output = append(output, ts)
		}
	}
	return output
}

// evaluateExpressions evaluates all provided Expressions in the
// EvaluationContext. If any evaluations error, evaluateExpressions will
// propagate that error. The resulting SeriesLists will be in an order
// corresponding to the provided Expresesions.
func evaluateExpressions(context EvaluationContext, expressions []Expression) ([]value, error) {
	if len(expressions) == 0 {
		return []value{}, nil
	}
	results := make([]value, len(expressions))
	for i, expr := range expressions {
		result, err := expr.Evaluate(context)
		if err != nil {
			return nil, err
		}
		results[i] = result
	}
	return results, nil
}

