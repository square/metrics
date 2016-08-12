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

package expression

import (
	"fmt"
	"strings"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
	"github.com/square/metrics/metric_metadata"
	"github.com/square/metrics/query/predicate"
	"github.com/square/metrics/timeseries"
	"github.com/square/metrics/util"
)

// Implementations
// ===============

type Duration struct {
	Source   string
	Duration time.Duration
}

func (expr Duration) Literal() interface{} {
	return expr.Duration
}

func (expr Duration) ActualEvaluate(context function.EvaluationContext) (function.Value, error) {
	return function.NewDurationValue(expr.Source, expr.Duration), nil
}

func (expr Duration) ExpressionDescription(mode function.DescriptionMode) string {
	if mode == function.StringMemoization {
		return fmt.Sprintf("%#v", expr)
	}
	return expr.Source
}

type Scalar struct {
	Value float64
}

func (expr Scalar) Literal() interface{} {
	return expr.Value
}

func (expr Scalar) ActualEvaluate(context function.EvaluationContext) (function.Value, error) {
	return function.ScalarValue(expr.Value), nil
}

func (expr Scalar) ExpressionDescription(mode function.DescriptionMode) string {
	if mode == function.StringMemoization {
		return fmt.Sprintf("%#v", expr)
	}
	return fmt.Sprintf("%+v", expr.Value)
}

type String struct {
	Value string
}

func (expr String) Literal() interface{} {
	return expr.Value
}

func (expr String) ActualEvaluate(context function.EvaluationContext) (function.Value, error) {
	return function.StringValue(expr.Value), nil
}

func (expr String) ExpressionDescription(mode function.DescriptionMode) string {
	if mode == function.StringMemoization {
		return fmt.Sprintf("%#v", expr)
	}
	return fmt.Sprintf("%q", expr.Value)
}

type MetricFetchExpression struct {
	MetricName string
	Predicate  predicate.Predicate
}

func (expr *MetricFetchExpression) ActualEvaluate(context function.EvaluationContext) (function.Value, error) {
	// Merge predicates appropriately
	p := predicate.All(expr.Predicate, context.Predicate())

	metricTagSets, err := context.MetricMetadataAPI().GetAllTags(api.MetricKey(expr.MetricName), metadata.Context{
		Profiler: context.Profiler(),
	})

	if err != nil {
		return nil, err
	}
	filtered := applyPredicates(metricTagSets, p)

	if err := context.FetchLimitConsume(len(filtered)); err != nil {
		return nil, err
	}

	metrics := make([]api.TaggedMetric, len(filtered))
	for i := range metrics {
		metrics[i] = api.TaggedMetric{MetricKey: api.MetricKey(expr.MetricName), TagSet: filtered[i]}
	}

	seriesList, err := context.TimeseriesStorageAPI().FetchMultipleTimeseries(
		timeseries.FetchMultipleRequest{
			Metrics: metrics,
			RequestDetails: timeseries.RequestDetails{
				SampleMethod: context.SampleMethod(),
				Timerange:    context.Timerange(),
				Ctx:          context.Ctx(),
				Profiler:     context.Profiler(),
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return function.SeriesListValue(seriesList), nil
}

func (expr *MetricFetchExpression) ExpressionDescription(mode function.DescriptionMode) string {
	if mode == function.StringMemoization {
		return fmt.Sprintf("fetch[%q][%s]", expr.MetricName, expr.Predicate.Query())
	}
	if expr.Predicate.Query() == "true" {
		return util.EscapeIdentifier(expr.MetricName)
	}
	return fmt.Sprintf("%s[%s]", util.EscapeIdentifier(expr.MetricName), expr.Predicate.Query())
}

type FunctionExpression struct {
	FunctionName     string
	Arguments        []function.Expression
	GroupBy          []string
	GroupByCollapses bool
}

func (expr *FunctionExpression) ActualEvaluate(context function.EvaluationContext) (function.Value, error) {
	fun, ok := context.RegistryGetFunction(expr.FunctionName)
	if !ok {
		return nil, SyntaxError{fmt.Sprintf("no such function %s", expr.FunctionName)}
	}

	return fun.Run(context, expr.Arguments, function.Groups{List: expr.GroupBy, Collapses: expr.GroupByCollapses})
}

func functionFormatString(argumentStrings []string, f FunctionExpression) string {
	switch f.FunctionName {
	case "+", "-", "*", "/":
		if len(f.Arguments) != 2 {
			// Then it's not actually an operator.
			break
		}
		return fmt.Sprintf("(%s %s %s)", argumentStrings[0], f.FunctionName, argumentStrings[1])
	}
	argumentString := strings.Join(argumentStrings, ", ")
	groupString := ""
	if len(f.GroupBy) != 0 {
		groupKeyword := "group by"
		if f.GroupByCollapses {
			groupKeyword = "collapse by"
		}
		escapedGroupBy := []string{}
		for _, group := range f.GroupBy {
			escapedGroupBy = append(escapedGroupBy, util.EscapeIdentifier(group))
		}
		groupString = fmt.Sprintf(" %s %s", groupKeyword, strings.Join(escapedGroupBy, ", "))
	}
	return fmt.Sprintf("%s(%s%s)", f.FunctionName, argumentString, groupString)
}

func (expr *FunctionExpression) ExpressionDescription(mode function.DescriptionMode) string {
	if widest, ok := mode.(function.WidestMode); ok {
		// Handle "widening" here
		if registered, ok := widest.Registry.GetFunction(expr.FunctionName); ok {
			if metricFunction, ok := registered.(function.MetricFunction); ok {
				if metricFunction.Widen != nil {
					widest.Current = metricFunction.Widen(widest, expr.Arguments)
				}
			}
		}
		for _, argument := range expr.Arguments {
			argument.ExpressionDescription(widest)
		}
		return ""
	}
	argumentStrings := []string{}
	for i := range expr.Arguments {
		argumentStrings = append(argumentStrings, expr.Arguments[i].ExpressionDescription(mode))
	}
	return functionFormatString(argumentStrings, *expr)
}

type AnnotationExpression struct {
	Expression function.Expression
	Annotation string
}

func (expr *AnnotationExpression) Literal() interface{} {
	literalExpression, ok := expr.Expression.(function.LiteralExpression)
	if !ok {
		return nil
	}
	return literalExpression.Literal()
}

// Evaluate evalutes the underlying expression without memoization, since its
// child expression should handle memoization itself.
func (expr *AnnotationExpression) Evaluate(context function.EvaluationContext) (function.Value, error) {
	return expr.Expression.Evaluate(context)
}

func (expr *AnnotationExpression) ExpressionDescription(mode function.DescriptionMode) string {
	if mode == function.StringName {
		return expr.Annotation
	}
	if mode == function.StringMemoization {
		return expr.Expression.ExpressionDescription(mode) // annotations can be ignored for memoization purposes since they don't modify their input
	}
	return fmt.Sprintf("%s {%s}", expr.Expression.ExpressionDescription(mode), expr.Annotation)
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
