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
)

// EvaluationContext is the central piece of logic, providing
// helper funcions & varaibles to evaluate a given piece of
// metrics query.
// * Contains Backend object, which can be used to fetch data
// from the backend system.s
// * Contains current timerange being queried for - this can be
// changed by say, application of time shift function.
type EvaluationContext struct {
	// Backend to fetch data from
	Backend api.Backend
	// Timerange to fetch data from
	Timerange api.Timerange
	// SampleMethod to use when up/downsampling to match the requested resolution
	SampleMethod api.SampleMethod
}

// A value is the result of evaluating an expression.
// They can be floating point values, strings, or series lists.
type value interface {
	toSeriesList(api.Timerange) (api.SeriesList, error)
	toString() (string, error)
	toScalar() (float64, error)
}

type conversionError struct {
	from string
	to   string
}

func (e conversionError) Error() string {
	return fmt.Sprintf("cannot convert from type %s to type %s", e.from, e.to)
}

// A seriesListValue is a value which holds a SeriesList
type seriesListValue api.SeriesList

func (value seriesListValue) toSeriesList(time api.Timerange) (api.SeriesList, error) {
	return api.SeriesList(value), nil
}
func (value seriesListValue) toString() (string, error) {
	return "", conversionError{"SeriesList", "string"}
}
func (value seriesListValue) toScalar() (float64, error) {
	return 0, conversionError{"SeriesList", "scalar"}
}

// A stringValue holds a string
type stringValue string

func (value stringValue) toSeriesList(time api.Timerange) (api.SeriesList, error) {
	return api.SeriesList{}, conversionError{"string", "SeriesList"}
}
func (value stringValue) toString() (string, error) {
	return string(value), nil
}
func (value stringValue) toScalar() (float64, error) {
	return 0, conversionError{"string", "scalar"}
}

// A scalarValue holds a float and can be converted to a serieslist
type scalarValue float64

func (value scalarValue) toSeriesList(timerange api.Timerange) (api.SeriesList, error) {
	if !timerange.IsValid() {
		return api.SeriesList{}, errors.New("Invalid context.Timerange")
	}

	series := make([]float64, timerange.Slots())
	for i := range series {
		series[i] = float64(value)
	}

	return api.SeriesList{
		Series:    []api.Timeseries{api.Timeseries{series, api.NewTagSet()}},
		Timerange: timerange,
	}, nil
}

func (value scalarValue) toString() (string, error) {
	return "", conversionError{"scalar", "string"}
}

func (value scalarValue) toScalar() (float64, error) {
	return float64(value), nil
}

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

// Implementations
// ===============

// Generates a Timeseries from the encapsulated scalar.
func (expr scalarExpression) Evaluate(context EvaluationContext) (value, error) {
	return scalarValue(expr.value), nil
}

func (expr *metricFetchExpression) Evaluate(context EvaluationContext) (value, error) {
	series, err := context.Backend.FetchSeries(api.TaggedMetric{api.MetricKey(expr.metricName), nil}, expr.predicate, context.SampleMethod, context.Timerange)
	if err != nil {
		return seriesListValue{}, err
	}
	return seriesListValue(*series), err
}

func (expr *functionExpression) Evaluate(context EvaluationContext) (value, error) {
	arguments := make([]value, len(expr.arguments))
	var err error
	for i := range arguments {
		arguments[i], err = expr.arguments[i].Evaluate(context)
		if err != nil {
			return nil, err
		}
	}
	name := expr.functionName
	switch name {
	case "+":
		fallthrough
	case "-":
		fallthrough
	case "*":
		fallthrough
	case "/":
		// Evaluation of a binary operator:
		if len(arguments) != 2 {
			return nil, errors.New(fmt.Sprintf("Function `%s` expects 2 operands but received %d (%+v)", name, len(arguments), arguments))
		}
		left := arguments[0]
		right := arguments[1]
		operatorMap := map[string]func(float64, float64) float64{
			"+": func(x, y float64) float64 { return x + y },
			"-": func(x, y float64) float64 { return x - y },
			"*": func(x, y float64) float64 { return x * y },
			"/": func(x, y float64) float64 { return x / y },
		}
		return evaluateBinaryOperation(context, name, left, right, operatorMap[name])
	default:
		return nil, errors.New(fmt.Sprintf("Invalid function: %s", functionName))
	}
}

//
// Auxiliary functions
//

// evaluateExpression wraps expr.Evaluate() to provide common messaging
// for errors. This can get pretty messy if the Expression we evaluate
// isn't a leaf node, but a leaf fails.
func evaluateExpression(context EvaluationContext, expr Expression) (value, error) {
	result, err := expr.Evaluate(context)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Evaluation of expression %+v failed:\n%s\n", expr, err.Error()))
	}
	return result, err
}

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
