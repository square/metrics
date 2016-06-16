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
	Literal  string
	Duration time.Duration
}

func (expr Duration) Evaluate(context function.EvaluationContext) (function.Value, error) {
	// @@ leaking param: expr to result ~r1 level=0
	return function.NewDurationValue(expr.Literal, expr.Duration), nil
	// @@ can inline Duration.Evaluate
}

// @@ inlining call to function.NewDurationValue
// @@ function.NewDurationValue(expr.Literal, expr.Duration) escapes to heap

func (expr Duration) Name() string {
	// @@ leaking param: expr to result ~r0 level=0
	return expr.Literal
	// @@ can inline Duration.Name
}
func (expr Duration) QueryString() string {
	// @@ leaking param: expr to result ~r0 level=0
	return expr.Literal
	// @@ can inline Duration.QueryString
}

type Scalar struct {
	Value float64
}

func (expr Scalar) Evaluate(context function.EvaluationContext) (function.Value, error) {
	return function.ScalarValue(expr.Value), nil
	// @@ can inline Scalar.Evaluate
}

// @@ function.ScalarValue(expr.Value) escapes to heap

func (expr Scalar) Name() string {
	return fmt.Sprintf("%+v", expr.Value)
}

// @@ expr.Value escapes to heap

func (expr Scalar) QueryString() string {
	return fmt.Sprintf("%+v", expr.Value)
}

// @@ expr.Value escapes to heap

type String struct {
	Value string
}

func (expr String) Evaluate(context function.EvaluationContext) (function.Value, error) {
	// @@ leaking param: expr to result ~r1 level=0
	return function.StringValue(expr.Value), nil
	// @@ can inline String.Evaluate
}

// @@ function.StringValue(expr.Value) escapes to heap

func (expr String) Name() string {
	// @@ leaking param: expr
	return fmt.Sprintf("%q", expr.Value)
}

// @@ expr.Value escapes to heap

func (expr String) QueryString() string {
	// @@ leaking param: expr
	return fmt.Sprintf("%q", expr.Value)
}

// @@ expr.Value escapes to heap

type MetricFetchExpression struct {
	MetricName string
	Predicate  predicate.Predicate
}

func (expr *MetricFetchExpression) Evaluate(context function.EvaluationContext) (function.Value, error) {
	// @@ leaking param content: expr
	// @@ leaking param: context
	// @@ leaking param content: expr
	// @@ leaking param content: expr
	// Merge predicates appropriately
	p := predicate.All(expr.Predicate, context.Predicate)

	metricTagSets, err := context.MetricMetadataAPI.GetAllTags(api.MetricKey(expr.MetricName), metadata.Context{
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
		// @@ make([]api.TaggedMetric, len(filtered)) escapes to heap
		// @@ make([]api.TaggedMetric, len(filtered)) escapes to heap
		metrics[i] = api.TaggedMetric{api.MetricKey(expr.MetricName), filtered[i]}
	}

	seriesList, err := context.TimeseriesStorageAPI.FetchMultipleTimeseries(
		timeseries.FetchMultipleRequest{
			metrics,
			timeseries.RequestDetails{
				context.SampleMethod,
				context.Timerange,
				context.Timeout,
				context.Profiler,
				context.UserSpecifiableConfig,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return function.SeriesListValue(seriesList), nil
}

// @@ function.SeriesListValue(seriesList) escapes to heap

func (expr *MetricFetchExpression) QueryString() string {
	// @@ leaking param content: expr
	// @@ leaking param content: expr
	// @@ leaking param content: expr
	// @@ leaking param content: expr
	if expr.Predicate.Query() == "true" {
		return util.EscapeIdentifier(expr.MetricName)
	}
	return fmt.Sprintf("%s[%s]", util.EscapeIdentifier(expr.MetricName), expr.Predicate.Query())
}

// @@ util.EscapeIdentifier(expr.MetricName) escapes to heap
// @@ expr.Predicate.Query() escapes to heap
func (expr *MetricFetchExpression) Name() string {
	// @@ leaking param content: expr
	return expr.QueryString()
}

type FunctionExpression struct {
	FunctionName     string
	Arguments        []function.Expression
	GroupBy          []string
	GroupByCollapses bool
}

func (expr *FunctionExpression) Evaluate(context function.EvaluationContext) (function.Value, error) {
	// @@ leaking param content: expr
	// @@ leaking param: context
	// @@ leaking param content: expr
	// @@ leaking param content: expr
	// @@ leaking param content: expr
	fun, ok := context.Registry.GetFunction(expr.FunctionName)
	if !ok {
		return nil, SyntaxError{fmt.Sprintf("no such function %s", expr.FunctionName)}
	}
	// @@ expr.FunctionName escapes to heap
	// @@ SyntaxError literal escapes to heap

	return fun.Run(context, expr.Arguments, function.Groups{expr.GroupBy, expr.GroupByCollapses})
}

func functionFormatString(argumentStrings []string, f FunctionExpression) string {
	// @@ leaking param content: argumentStrings
	// @@ leaking param: f
	// @@ leaking param content: argumentStrings
	// @@ leaking param content: argumentStrings
	// @@ leaking param content: argumentStrings
	switch f.FunctionName {
	case "+", "-", "*", "/":
		if len(f.Arguments) != 2 {
			// Then it's not actually an operator.
			break
		}
		return fmt.Sprintf("(%s %s %s)", argumentStrings[0], f.FunctionName, argumentStrings[1])
	}
	// @@ argumentStrings[0] escapes to heap
	// @@ f.FunctionName escapes to heap
	// @@ argumentStrings[1] escapes to heap
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
	// @@ groupKeyword escapes to heap
	// @@ strings.Join(escapedGroupBy, ", ") escapes to heap
	return fmt.Sprintf("%s(%s%s)", f.FunctionName, argumentString, groupString)
}

// @@ f.FunctionName escapes to heap
// @@ argumentString escapes to heap
// @@ groupString escapes to heap

func (expr *FunctionExpression) QueryString() string {
	// @@ leaking param content: expr
	// @@ leaking param content: expr
	argumentStrings := []string{}
	for i := range expr.Arguments {
		argumentStrings = append(argumentStrings, expr.Arguments[i].QueryString())
	}
	return functionFormatString(argumentStrings, *expr)
}

func (expr *FunctionExpression) Name() string {
	// @@ leaking param: expr to result ~r0 level=2
	// @@ leaking param content: expr
	// @@ leaking param content: expr
	// TODO: deprecate (and remove) this behavior before it becomes permanent
	if expr.FunctionName == "transform.alias" && len(expr.Arguments) == 2 {
		if alias, ok := expr.Arguments[1].(String); ok {
			return alias.Value
		}
	}
	argumentStrings := []string{}
	for i := range expr.Arguments {
		argumentStrings = append(argumentStrings, expr.Arguments[i].Name())
	}
	return functionFormatString(argumentStrings, *expr)
}

type AnnotationExpression struct {
	Expression function.Expression
	Annotation string
}

func (expr *AnnotationExpression) Evaluate(context function.EvaluationContext) (function.Value, error) {
	// @@ leaking param: context
	// @@ leaking param content: expr
	return expr.Expression.Evaluate(context)
}

func (expr *AnnotationExpression) QueryString() string {
	// @@ leaking param content: expr
	// @@ leaking param content: expr
	return fmt.Sprintf("%s {%s}", expr.Expression.QueryString(), expr.Annotation)
}

// @@ expr.Expression.QueryString() escapes to heap
// @@ expr.Annotation escapes to heap

func (expr *AnnotationExpression) Name() string {
	// @@ leaking param: expr to result ~r0 level=1
	return expr.Annotation
	// @@ can inline (*AnnotationExpression).Name
}

// Auxiliary functions
// ===================

func applyPredicates(tagSets []api.TagSet, predicate predicate.Predicate) []api.TagSet {
	// @@ leaking param content: tagSets
	// @@ leaking param: predicate
	// @@ leaking param content: tagSets
	output := []api.TagSet{}
	for _, ts := range tagSets {
		// @@ []api.TagSet literal escapes to heap
		if predicate.Apply(ts) {
			output = append(output, ts)
		}
	}
	return output
}
