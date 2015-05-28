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

// Expression is a piece of code, which can be evaluated in a given
// EvaluationContext. EvaluationContext must never be changed in an Evalute().
//
// The contract of Expressions is that leaf nodes must sample a resulting
// timeseries according to the resolution specified in its EvaluationContext's
// Timerange. Internal nodes may assume that results from evaluating child
// Expressions correspond to the timerange in the current EvaluationContext.
type Expression interface {
	// Evaluate the given expression.
	Evaluate(context EvaluationContext) (*api.SeriesList, error)
}

// Implementations
// ===============

// Generates a Timeseries from the encapsulated scalar.
func (expr *scalarExpression) Evaluate(context EvaluationContext) (*api.SeriesList, error) {
	if !context.Timerange.IsValid() {
		return nil, errors.New("Invalid context.Timerange")
	}

	series := []float64{}
	for i := 0; i < context.Timerange.Slots(); i += 1 {
		series = append(series, expr.value)
	}

	return &api.SeriesList{
		Series:    []api.Timeseries{api.Timeseries{series, api.NewTagSet()}},
		Timerange: context.Timerange,
	}, nil
}

func (expr *metricFetchExpression) Evaluate(context EvaluationContext) (*api.SeriesList, error) {
	return context.Backend.FetchSeries(api.TaggedMetric{api.MetricKey(expr.metricName), nil}, expr.predicate, context.SampleMethod, context.Timerange)
}

func (expr *functionExpression) Evaluate(context EvaluationContext) (*api.SeriesList, error) {
	switch expr.functionName {
	case "+":
		return evaluateBinaryOperation(context, expr.functionName, expr.arguments,
			func(left, right float64) float64 { return left + right })
	case "-":
		return evaluateBinaryOperation(context, expr.functionName, expr.arguments,
			func(left, right float64) float64 { return left - right })
	case "/":
		return evaluateBinaryOperation(context, expr.functionName, expr.arguments,
			func(left, right float64) float64 { return left / right })
	default:
		return nil, errors.New(fmt.Sprintf("Invalid function: %s", functionName))
	}

	return nil, errors.New("I'm not sure how you got here...")
}

//
// Auxiliary functions
//

// evaluateExpression wraps expr.Evaluate() to provide common messaging
// for errors. This can get pretty messy if the Expression we evaluate
// isn't a leaf node, but a leaf fails.
func evaluateExpression(context EvaluationContext, expr Expression) (*api.SeriesList, error) {
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
	operands []Expression,
	evaluate func(float64, float64) float64,
) (*api.SeriesList, error) {
	if len(operands) != 2 {
		return nil, errors.New(fmt.Sprintf("Function `%s` expects 2 operands but received %d (%+v)", functionName, len(operands), operands))
	}

	results, err := evaluateExpressions(context, operands)
	if err != nil {
		return nil, err
	}

	for _, seriesList := range results {
		if len(seriesList.Series) != 1 {
			return nil, errors.New(fmt.Sprintf("Operand %+v must contain only one Timeseries", seriesList.Series))
		}
	}

	left := results[0].Series[0]
	right := results[1].Series[0]
	result := make([]float64, len(left.Values))

	for i := 0; i < len(left.Values); i += 1 {
		result[i] = evaluate(left.Values[i], right.Values[i])
	}

	return &api.SeriesList{
		Series:    []api.Timeseries{api.Timeseries{result, api.NewTagSet()}},
		Timerange: context.Timerange,
	}, nil
}

// evaluateExpressions evaluates all provided Expressions in the
// EvaluationContext. If any evaluations error, evaluateExpressions will
// propagate that error. The resulting SeriesLists will be in an order
// corresponding to the provided Expressions.
func evaluateExpressions(context EvaluationContext, expressions []Expression) ([]*api.SeriesList, error) {
	if len(expressions) == 0 {
		return []*api.SeriesList{}, nil
	}

	results := []*api.SeriesList{}
	for _, expr := range expressions {
		result, err := expr.Evaluate(context)
		if err != nil {
			return []*api.SeriesList{}, err
		}

		results = append(results, result)
	}

	return results, nil
}
