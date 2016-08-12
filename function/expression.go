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

package function

import (
	"sync"
	"time"

	"github.com/square/metrics/api"
)

// Expression is a piece of code, which can be evaluated in a given
// EvaluationContext. EvaluationContext must never be changed in an Evaluate().
//
// If an Expression returns a SeriesList, its timerange must match the context's
// timerange exactly.
type Expression interface {
	Evaluate(context EvaluationContext) (Value, error)
	ExpressionDescription(DescriptionMode) string
}

// A LiteralExpression is an expression that holds a literal value.
// Returning `nil` indicates that your particular instance doesn't actually
// hold any literal value.
type LiteralExpression interface {
	Literal() interface{}
}

// An ActualExpression is how expressions are internally implemented by the
// library, but not how they should be consumed.
type ActualExpression interface {
	// Evaluate the given expression.
	ActualEvaluate(context EvaluationContext) (Value, error)
	ExpressionDescription(DescriptionMode) string
}

// DescriptionMode indicates how the expression should be evaluated.
type DescriptionMode interface{}

type StringNameMode struct{}

type StringQueryMode struct{}

type StringMemoizationMode struct{}

type WidestMode struct {
	Registry   Registry
	Current    time.Time
	Earliest   *time.Time
	Resolution time.Duration
	mutex      sync.Mutex
}

func (w *WidestMode) AddTime(t time.Time) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.Earliest.After(t) {
		*w.Earliest = t
	}
}

// StringName is for human readability and respects aliases
func StringName() DescriptionMode {
	return StringNameMode{}
}

// StringQuery is for humans but ignores aliases, presenting the query as written
func StringQuery() DescriptionMode {
	return StringQueryMode{}
}

// StringMemoization is not for humans and intended to give a unique name to every expression
func StringMemoization() DescriptionMode {
	return StringMemoizationMode{}
}

// EvaluateToScalar is a helper function that takes an Expression and makes it a scalar.
func EvaluateToScalar(e Expression, context EvaluationContext) (float64, error) {
	scalarValue, err := e.Evaluate(context)
	if err != nil {
		return 0, err
	}
	value, convErr := scalarValue.ToScalar()
	if convErr != nil {
		return 0, convErr.WithContext(e.ExpressionDescription(StringQuery))
	}
	return value, nil
}

// EvaluateToScalarSet is a helper function that takes an Expression and makes it a scalar set.
func EvaluateToScalarSet(e Expression, context EvaluationContext) (ScalarSet, error) {
	scalarValue, err := e.Evaluate(context)
	if err != nil {
		return nil, err
	}
	value, convErr := scalarValue.ToScalarSet()
	if convErr != nil {
		return nil, convErr.WithContext(e.ExpressionDescription(StringQuery))
	}
	return value, nil
}

// EvaluateToDuration is a helper function that takes an Expression and makes it a duration.
func EvaluateToDuration(e Expression, context EvaluationContext) (time.Duration, error) {
	durationValue, err := e.Evaluate(context)
	if err != nil {
		return 0, err
	}
	value, convErr := durationValue.ToDuration()
	if convErr != nil {
		return 0, convErr.WithContext(e.ExpressionDescription(StringQuery))
	}
	return value, nil
}

// EvaluateToSeriesList is a helper function that takes an Expression and makes it a series list.
func EvaluateToSeriesList(e Expression, context EvaluationContext) (api.SeriesList, error) {
	seriesValue, err := e.Evaluate(context)
	if err != nil {
		return api.SeriesList{}, err
	}
	value, convErr := seriesValue.ToSeriesList(context.private.Timerange)
	if convErr != nil {
		return api.SeriesList{}, convErr.WithContext(e.ExpressionDescription(StringQuery))
	}
	return value, nil
}

// EvaluateToString is a helper function that takes an Expression and makes it a string.
func EvaluateToString(e Expression, context EvaluationContext) (string, error) {
	stringValue, err := e.Evaluate(context)
	if err != nil {
		return "", err
	}
	value, convErr := stringValue.ToString()
	if convErr != nil {
		return "", convErr.WithContext(e.ExpressionDescription(StringQuery))
	}
	return value, nil
}

// EvaluateMany evaluates a list of expressions using a single EvaluationContext.
// If any evaluation errors, EvaluateMany will propagate that error. The resulting values
// will be in the order corresponding to the provided expressions.
func EvaluateMany(context EvaluationContext, expressions []Expression) ([]Value, error) {
	type result struct {
		index int
		err   error
		value Value
	}
	length := len(expressions)
	if length == 0 {
		return []Value{}, nil
	}
	if length == 1 {
		result, err := expressions[0].Evaluate(context)
		if err != nil {
			return nil, err
		}
		return []Value{result}, nil
	}
	// concurrent evaluations
	results := make(chan result, length)
	for i, expr := range expressions {
		go func(i int, expr Expression) {
			value, err := expr.Evaluate(context)
			results <- result{i, err, value}
		}(i, expr)
	}
	array := make([]Value, length)
	for i := 0; i < length; i++ {
		result := <-results
		if result.err != nil {
			return nil, result.err
		}
		array[result.index] = result.value
	}
	return array, nil
}
