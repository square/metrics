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
)

// Implementations
// ===============

func (expr durationExpression) Evaluate(context function.EvaluationContext) (function.Value, error) {
	return function.NewDurationValue(expr.name, expr.duration), nil
}

func (expr scalarExpression) Evaluate(context function.EvaluationContext) (function.Value, error) {
	return function.ScalarValue(expr.value), nil
}

func (expr stringExpression) Evaluate(context function.EvaluationContext) (function.Value, error) {
	return function.StringValue(expr.value), nil
}

func (expr *metricFetchExpression) Evaluate(context function.EvaluationContext) (function.Value, error) {
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

	updateFunction := func() ([]api.TagSet, error) {
		metricTagSets, err := context.MetricMetadataAPI.GetAllTags(api.MetricKey(expr.metricName), api.MetricMetadataAPIContext{
			Profiler: context.Profiler,
		})
		if err != nil {
			return nil, err
		}
		return metricTagSets, nil
	}
	metricTagSets, err := context.OptimizationConfiguration.AllTagsCacheHitOrExecute(api.MetricKey(expr.metricName), updateFunction)
	if err != nil {
		return nil, err
	}
	filtered := applyPredicates(metricTagSets, predicate)

	if err := context.FetchLimit.Consume(len(filtered)); err != nil {
		return nil, err
	}

	metrics := make([][]byte, len(filtered))
	for i := range metrics {
		metric, err := context.MetricConverter.ToUntagged(api.TaggedMetric{api.MetricKey(expr.metricName), filtered[i]})
		if err != nil {
			// TODO: evaluate if this is a good idea- otherwise persisted entries that cannot be converted will be an error
			return nil, err
		}
		metrics[i] = metric
	}

	valuelist, err := context.TimeseriesStorageAPI.FetchMultipleTimeseries(
		api.FetchMultipleTimeseriesRequest{
			metrics,
			context.SampleMethod,
			context.Timerange,
			context.Cancellable,
			context.Profiler,
			context.UserSpecifiableConfig,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(valuelist) != len(filtered) {
		return nil, fmt.Errorf("Internal Server Error - Attempted to fetch %d but received only %d (without any indicated error).", len(filtered), len(valuelist))
	}

	serieslist := api.SeriesList{
		Series:    make([]api.Timeseries, len(filtered)),
		Timerange: context.Timerange,
	}

	for i := range serieslist.Series {
		serieslist.Series[i] = api.Timeseries{
			Values: valuelist[i],
			TagSet: filtered[i],
		}
	}

	return serieslist, nil
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

func applyPredicates(tagSets []api.TagSet, predicate api.Predicate) []api.TagSet {
	output := []api.TagSet{}
	for _, ts := range tagSets {
		if predicate.Apply(ts) {
			output = append(output, ts)
		}
	}
	return output
}
