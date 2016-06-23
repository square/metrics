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
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/inspect"
	"github.com/square/metrics/metric_metadata"
	"github.com/square/metrics/query/predicate"
	"github.com/square/metrics/timeseries"

	"golang.org/x/net/context"
)

// An EvaluationContextBuilder is used to create an EvaluationContext because
// the EvaluationContext's fields are private to prevent accidental modification.
type EvaluationContextBuilder struct {
	TimeseriesStorageAPI timeseries.StorageAPI   // Backend to fetch data from
	MetricMetadataAPI    metadata.MetricAPI      // Api to obtain metadata from
	Timerange            api.Timerange           // Timerange to fetch data from
	SampleMethod         timeseries.SampleMethod // SampleMethod to use when up/downsampling to match the requested resolution
	Predicate            predicate.Predicate     // Predicate to apply to TagSets prior to fetching
	FetchLimit           FetchCounter            // A limit on the number of fetches which may be performed
	Registry             Registry
	Profiler             *inspect.Profiler // A profiler pointer
	EvaluationNotes      *EvaluationNotes  // Debug + numerical notes that can be added during evaluation
	Ctx                  context.Context
}

// Build creates an evaluation context from the provided builder.
func (builder EvaluationContextBuilder) Build() EvaluationContext {
	return EvaluationContext{
		private:     builder,
		memoization: newMemo(),
	}
}

// EvaluationContext holds all information relevant to executing a single query.
type EvaluationContext struct {
	private     EvaluationContextBuilder // So that it can't be easily modified from outside this package.
	memoization *memoization
}

// TimeseriesStorageAPI returns the underlying timeseries.StorageAPI.
func (context EvaluationContext) TimeseriesStorageAPI() timeseries.StorageAPI {
	return context.private.TimeseriesStorageAPI
}

// MetricMetadataAPI returns the underlying metadata.MetricAPI.
func (context EvaluationContext) MetricMetadataAPI() metadata.MetricAPI {
	return context.private.MetricMetadataAPI
}

// Timerange returns the underlying api.Timerange.
func (context EvaluationContext) Timerange() api.Timerange {
	return context.private.Timerange
}

// SampleMethod returns the underlying timeseries.SampleMethod.
func (context EvaluationContext) SampleMethod() timeseries.SampleMethod {
	return context.private.SampleMethod
}

// Predicate returns the underlying predicate.Predicate.
func (context EvaluationContext) Predicate() predicate.Predicate {
	return context.private.Predicate
}

// FetchLimitConsume tries to consume the amount of resources from the limit,
// returning a non-nil error if this would overdraw the alloted limit.
func (context EvaluationContext) FetchLimitConsume(n int) error {
	return context.private.FetchLimit.Consume(n)
}

// Ctx returns the underlying Context instance for the evaluation.
func (context EvaluationContext) Ctx() context.Context {
	return context.private.Ctx
}

// RegistryGetFunction gets the function of the provided name, return true if
// the function could be found and false otherwise.
func (context EvaluationContext) RegistryGetFunction(name string) (Function, bool) {
	return context.private.Registry.GetFunction(name)
}

// Profiler returns the underlying inspect.Profiler instance.
func (context EvaluationContext) Profiler() *inspect.Profiler {
	return context.private.Profiler
}

// AddNote adds a note to the evaluation context.
func (context EvaluationContext) AddNote(note string) {
	context.private.EvaluationNotes.AddNote(note)
}

// Notes returns all notes added to the evaluation context.
func (context EvaluationContext) Notes() []string {
	return context.private.EvaluationNotes.Notes()
}

// EvaluationNotes holds notes that were recorded during evaluation.
type EvaluationNotes struct {
	mutex sync.Mutex
	notes []string
}

// AddNote adds a new note to the collection in a threadsafe manner.
func (notes *EvaluationNotes) AddNote(note string) {
	if notes == nil {
		return
	}
	notes.mutex.Lock()
	defer notes.mutex.Unlock()
	notes.notes = append(notes.notes, note)
}

// Notes returns the current collection of notes in a threadsafe manner.
func (notes *EvaluationNotes) Notes() []string {
	if notes == nil {
		return nil
	}
	notes.mutex.Lock()
	defer notes.mutex.Unlock()
	return notes.notes
}

// WithTimerange duplicates the EvaluationContext but with a new timerange.
func (context EvaluationContext) WithTimerange(t api.Timerange) EvaluationContext {
	context.private.Timerange = t
	context.memoization = newMemo()
	return context
}

func (context EvaluationContext) EvaluateMemoized(expression ActualExpression) (Value, error) {
	return context.memoization.evaluate(expression, context)
}

// FetchCounter is used to count the number of fetches remaining in a thread-safe manner.
type FetchCounter struct {
	count *int32
	limit int
}

// NewFetchCounter creates a FetchCounter with n as the limit.
func NewFetchCounter(n int) FetchCounter {
	n32 := int32(n)
	return FetchCounter{
		count: &n32,
		limit: n,
	}
}

// Limit returns the max # of fetches allowed by this counter.
func (c FetchCounter) Limit() int {
	return c.limit
}

// Current returns the current number of fetches remaining for the counter.
func (c FetchCounter) Current() int {
	return c.limit - int(atomic.LoadInt32(c.count))
}

// Consume decrements the internal counter and returns whether the result is at least 0.
// It does so in a threadsafe manner.
func (c FetchCounter) Consume(n int) error {
	remaining := atomic.AddInt32(c.count, -int32(n))
	if remaining < 0 {
		return fmt.Errorf("performing fetch of %d additional series brings the total to %d, which exceeds the specified limit %d", n, c.limit-int(remaining), c.limit)
	}
	return nil
}

// Expression is a piece of code, which can be evaluated in a given
// EvaluationContext. EvaluationContext must never be changed in an Evaluate().
//
// If an Expression returns a SeriesList, its timerange must match the context's
// timerange exactly.
type Expression interface {
	Evaluate(context EvaluationContext) (Value, error)
	ExpressionString(DescriptionMode) string
}

// An ActualExpression is how expressions are internally implemented by the
// library, but not how they should be consumed. ActualExpressions don't
// benefit from memoization, but can
type ActualExpression interface {
	// Evaluate the given expression.
	ActualEvaluate(context EvaluationContext) (Value, error)
	ExpressionString(DescriptionMode) string
}

// DescriptionMode indicates how the expression should be evaluated.
type DescriptionMode int

const (
	StringName        DescriptionMode = iota // StringName is for human readability and respects aliases
	StringQuery                              // StringQuery is for humans but ignores aliases, presenting the query as written
	StringMemoization                        // StringMemoization is not for humans and intended to give a unique name to every expression
)

// EvaluateToScalar is a helper function that takes an Expression and makes it a scalar.
func EvaluateToScalar(e Expression, context EvaluationContext) (float64, error) {
	scalarValue, err := e.Evaluate(context)
	if err != nil {
		return 0, err
	}
	value, convErr := scalarValue.ToScalar()
	if convErr != nil {
		return 0, convErr.WithContext(e.ExpressionString(StringQuery))
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
		return nil, convErr.WithContext(e.ExpressionString(StringQuery))
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
		return 0, convErr.WithContext(e.ExpressionString(StringQuery))
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
		return api.SeriesList{}, convErr.WithContext(e.ExpressionString(StringQuery))
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
		return "", convErr.WithContext(e.ExpressionString(StringQuery))
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
