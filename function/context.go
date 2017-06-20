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
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/square/metrics/api"
	"github.com/square/metrics/inspect"
	"github.com/square/metrics/metric_metadata"
	"github.com/square/metrics/query/predicate"
	"github.com/square/metrics/timeseries"
)

// An EvaluationContextBuilder is used to create an EvaluationContext because
// the EvaluationContext's fields are private to prevent accidental modification.
type EvaluationContextBuilder struct {
	TimeseriesStorageAPI timeseries.StorageAPI   // Backend to fetch data from
	MetricMetadataAPI    metadata.MetricAPI      // Api to obtain metadata from
	Registry             Registry                // Registry stores functions
	SampleMethod         timeseries.SampleMethod // SampleMethod to use when up/downsampling to match the requested resolution
	FetchLimit           FetchCounter            // A limit on the number of fetches which may be performed
	Profiler             *inspect.Profiler       // A profiler pointer
	EvaluationNotes      *EvaluationNotes        // Debug + numerical notes that can be added during evaluation
	Ctx                  context.Context

	// These may be changed in sub-contexts while evaluating the query.
	Timerange api.Timerange       // Timerange to fetch data from
	Predicate predicate.Predicate // Predicate to apply to TagSets prior to fetching
}

// Build creates an evaluation context from the provided builder.
func (builder EvaluationContextBuilder) Build() EvaluationContext {
	memoMap := newMemoMap()
	memo := memoMap.get(builder.memoizationIdentity())
	return EvaluationContext{
		private:        builder,
		memoizationMap: memoMap,
		memoization:    memo,
	}
}

// EvaluationContext holds all information relevant to executing a single query.
type EvaluationContext struct {
	private        EvaluationContextBuilder // So that it can't be easily modified from outside this package.
	memoizationMap *memoizationMap          // This map stores results of expression evaluations
	memoization    *memoization             // This map stores memoizations for better sharing between contexts
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
	if context.private.Timerange == t {
		// don't reduce sharing if the timerange hasn't changed
		return context
	}
	context.private.Timerange = t
	context.memoization = context.memoizationMap.get(context.private.memoizationIdentity())
	return context
}

// WithAdditionalConstraint return a new copy of the evaluation context with a
// distinct memoization map.
func (context EvaluationContext) WithAdditionalConstraint(p predicate.Predicate) EvaluationContext {
	context.private.Predicate = predicate.All(context.private.Predicate, p)
	context.memoization = context.memoizationMap.get(context.private.memoizationIdentity())
	return context
}

// EvaluateMemoized evaluates the given ActualExpression using the memoization
// map internal to the context.
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

type contextIdentity struct {
	Timerange      api.Timerange
	PredicateQuery string
}

// memoizationIdentity is used to improve sharing between contexts
func (builder EvaluationContextBuilder) memoizationIdentity() contextIdentity {
	timerange := builder.Timerange
	predicate := "<nil>"
	if builder.Predicate != nil {
		predicate = builder.Predicate.Query()
	}
	return contextIdentity{
		Timerange:      timerange,
		PredicateQuery: predicate,
	}
}
