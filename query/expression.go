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

	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
	"github.com/square/metrics/query/predicate"
)

// Implementations
// ===============

func (expr durationExpression) Evaluate(context function.EvaluationContext) (function.Value, error) {
	return function.DurationValue(expr.duration), nil
}

func (expr scalarExpression) Evaluate(context function.EvaluationContext) (function.Value, error) {
	return function.ScalarValue(expr.value), nil
}

func (expr stringExpression) Evaluate(context function.EvaluationContext) (function.Value, error) {
	return function.StringValue(expr.value), nil
}

func (expr *metricFetchExpression) Evaluate(context function.EvaluationContext) (function.Value, error) {
	// Merge predicates appropriately
	p := predicate.All(expr.predicate, context.Predicate)

	metricTagSets, err := context.MetricMetadataAPI.GetAllTags(api.MetricKey(expr.metricName), api.MetricMetadataAPIContext{
		Profiler: context.Profiler,
	})

	if err != nil {
		return nil, err
	}
	filtered := applyPredicates(metricTagSets, p)

	if err := context.FetchLimit.Consume(len(filtered)); err != nil {
		return nil, err
	}

	metrics := make([]api.TaggedMetric, len(filtered))
	for i := range metrics {
		metrics[i] = api.TaggedMetric{api.MetricKey(expr.metricName), filtered[i]}
	}

	return context.TimeseriesStorageAPI.FetchMultipleTimeseries(
		api.FetchMultipleTimeseriesRequest{
			metrics,
			context.SampleMethod,
			context.Timerange,
			context.Cancellable,
			context.Profiler,
			context.UserSpecifiableConfig,
		},
	)
}

func (expr *functionExpression) Evaluate(context function.EvaluationContext) (function.Value, error) {
	fun, ok := context.Registry.GetFunction(expr.functionName)
	if !ok {
		return nil, SyntaxError{expr.functionName, fmt.Sprintf("no such function %s", expr.functionName)}
	}

	return fun.Evaluate(context, expr.arguments, expr.groupBy, expr.groupByCollapses)
}

// Auxiliary functions
// ===================

func applyPredicates(tagSets []api.TagSet, predicate predicate.Predicate) []api.TagSet {
	output := []api.TagSet{}
	for _, ts := range tagSets {
		if predicate.Apply(ts) {
			output = append(output, ts)
		}
	}
	return output
}
