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
	"math"
	"sync/atomic"

	"github.com/square/metrics/api"
	"github.com/square/metrics/query/aggregate"
)

func init() {
	// Arithmetic operators
	MustRegister(MakeOperatorMetricFunction("+", func(x float64, y float64) float64 { return x + y }))
	MustRegister(MakeOperatorMetricFunction("-", func(x float64, y float64) float64 { return x - y }))
	MustRegister(MakeOperatorMetricFunction("*", func(x float64, y float64) float64 { return x * y }))
	MustRegister(MakeOperatorMetricFunction("/", func(x float64, y float64) float64 { return x / y }))
	// Aggregates
	MustRegister(MakeAggregateMetricFunction("aggregate.max", aggregate.AggregateMax))
	MustRegister(MakeAggregateMetricFunction("aggregate.min", aggregate.AggregateMin))
	MustRegister(MakeAggregateMetricFunction("aggregate.mean", aggregate.AggregateMean))
	MustRegister(MakeAggregateMetricFunction("aggregate.sum", aggregate.AggregateSum))
	// Transformations
	MustRegister(MakeTransformMetricFunction("transform.derivative", 0, transformDerivative))
	MustRegister(MakeTransformMetricFunction("transform.integral", 0, transformIntegral))
	MustRegister(MakeTransformMetricFunction("transform.rate", 0, transformRate))
	MustRegister(MakeTransformMetricFunction("transform.cumulative", 0, transformCumulative))
	MustRegister(MakeTransformMetricFunction("transform.moving_average", 1, transformMovingAverage))
	MustRegister(MakeTransformMetricFunction("transform.default", 1, transformDefault))
	MustRegister(MakeTransformMetricFunction("transform.abs", 0, transformMapMaker("abs", math.Abs)))
	// Timeshift
	MustRegister(TimeshiftFunction)
	// Filter
	MustRegister(MakeFilterMetricFunction("filter.highest_mean", aggregate.AggregateMean, false))
	MustRegister(MakeFilterMetricFunction("filter.lowest_mean", aggregate.AggregateMean, true))
	MustRegister(MakeFilterMetricFunction("filter.highest_max", aggregate.AggregateMax, false))
	MustRegister(MakeFilterMetricFunction("filter.lowest_max", aggregate.AggregateMax, true))
	MustRegister(MakeFilterMetricFunction("filter.highest_min", aggregate.AggregateMin, false))
	MustRegister(MakeFilterMetricFunction("filter.lowest_min", aggregate.AggregateMin, true))
}

// EvaluationContext is the central piece of logic, providing
// helper funcions & varaibles to evaluate a given piece of
// metrics query.
// * Contains Backend object, which can be used to fetch data
// from the backend system.s
// * Contains current timerange being queried for - this can be
// changed by say, application of time shift function.
type EvaluationContext struct {
	MultiBackend api.MultiBackend // Backend to fetch data from
	API          api.API          // Api to obtain metadata from
	Timerange    api.Timerange    // Timerange to fetch data from
	SampleMethod api.SampleMethod // SampleMethod to use when up/downsampling to match the requested resolution
	Predicate    api.Predicate    // Predicate to apply to TagSets prior to fetching
	FetchLimit   fetchCounter     // A limit on the number of fetches which may be performed
}

// fetchCounter is used to count the number of fetches remaining in a thread-safe manner.
type fetchCounter struct {
	count *int32
}

func NewFetchCounter(n int) fetchCounter {
	n32 := int32(n)
	return fetchCounter{
		count: &n32,
	}
}

// Consume decrements the internal counter and returns whether the result is at least 0.
// It does so in a threadsafe manner.
func (c fetchCounter) Consume(n int) bool {
	return atomic.AddInt32(c.count, -int32(n)) >= 0
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

// Implementations
// ===============

// Generates a Timeseries from the encapsulated scalar.
func (expr scalarExpression) Evaluate(context EvaluationContext) (value, error) {
	return scalarValue(expr.value), nil
}

func (expr stringExpression) Evaluate(context EvaluationContext) (value, error) {
	return stringValue(expr.value), nil
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

	ok := context.FetchLimit.Consume(len(filtered))

	if !ok {
		return nil, errors.New("fetch limit exceeded: too many series to fetch")
	}

	metrics := make([]api.TaggedMetric, len(filtered))
	for i := range metrics {
		metrics[i] = api.TaggedMetric{api.MetricKey(expr.metricName), filtered[i]}
	}

	serieslist, err := context.MultiBackend.FetchMultipleSeries(metrics, context.SampleMethod, context.Timerange, context.API)

	if err != nil {
		return nil, err
	}

	return seriesListValue(serieslist), nil
}

func (expr *functionExpression) Evaluate(context EvaluationContext) (value, error) {
	fun, ok := GetFunction(expr.functionName)
	if !ok {
		return nil, SyntaxError{expr.functionName, fmt.Sprintf("no such function %s", expr.functionName)}
	}

	return fun.Evaluate(context, expr.arguments, expr.groupBy)
}

// Auxiliary functions
// ===================

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
