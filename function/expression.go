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
	"github.com/square/metrics/tasks"
	"github.com/square/metrics/timeseries"
)

// EvaluationContext is the central piece of logic, providing
// helper funcions & varaibles to evaluate a given piece of
// metrics query.
// * Contains a TimeseriesStorageAPI object, which can be used to fetch data
// from the backend systems.
// * Contains current timerange being queried for - this can be
// changed by say, application of time shift function.
type EvaluationContext struct {
	TimeseriesStorageAPI  timeseries.StorageAPI   // Backend to fetch data from
	MetricMetadataAPI     metadata.MetricAPI      // Api to obtain metadata from
	Timerange             api.Timerange           // Timerange to fetch data from
	SampleMethod          timeseries.SampleMethod // SampleMethod to use when up/downsampling to match the requested resolution
	Predicate             predicate.Predicate     // Predicate to apply to TagSets prior to fetching
	FetchLimit            FetchCounter            // A limit on the number of fetches which may be performed
	Timeout               *tasks.Timeout
	Registry              Registry
	Profiler              *inspect.Profiler // A profiler pointer
	EvaluationNotes       *EvaluationNotes  // Debug + numerical notes that can be added during evaluation
	UserSpecifiableConfig timeseries.UserSpecifiableConfig
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

// The Registry interface defines a mapping from names to MetricFunctions
// and provides a way to get the full list of functions defined.
type Registry interface {
	GetFunction(string) (MetricFunction, bool) // returns an instance of MetricFunction
	All() []string                             // all the registered functions
}

// Groups holds grouping information - which tags to group by (if any), and whether to `collapse` (Collapses = true) or `group` (Collapses = false)
type Groups struct {
	List      []string // the tags to group by
	Collapses bool     // whether to "collapse by" instead of "group by"
}

// MetricFunction defines a common logic to dispatch a function in MQE.
type MetricFunction struct {
	Name          string
	MinArguments  int
	MaxArguments  int
	AllowsGroupBy bool // Whether the function allows a 'group by' clause.
	Compute       func(EvaluationContext, []Expression, Groups) (Value, error)
}

// WithTimerange duplicates the EvaluationContext but with a new timerange.
func (e EvaluationContext) WithTimerange(t api.Timerange) EvaluationContext {
	e.Timerange = t
	return e
}

// Evaluate the given metric function.
func (f MetricFunction) Evaluate(context EvaluationContext,
	arguments []Expression, groupBy []string, collapses bool) (Value, error) {
	// preprocessing
	length := len(arguments)
	if length < f.MinArguments || (f.MaxArguments != -1 && f.MaxArguments < length) {
		return nil, ArgumentLengthError{f.Name, f.MinArguments, f.MaxArguments, length}
	}
	if len(groupBy) > 0 && !f.AllowsGroupBy {
		// TODO(jee) - use typed errors
		return nil, fmt.Errorf("function %s doesn't allow a group-by clause", f.Name)
	}
	return f.Compute(context, arguments, Groups{groupBy, collapses})
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
	if atomic.AddInt32(c.count, -int32(n)) < 0 {
		return NewLimitError("fetch limit exceeded: too many series to fetch", n, c.limit)
	}
	return nil
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
	Evaluate(context EvaluationContext) (Value, error)
	Name() string
	QueryString() string
}

// EvaluateToScalar is a helper function that takes an Expression and makes it a scalar.
func EvaluateToScalar(e Expression, context EvaluationContext) (float64, error) {
	scalarValue, err := e.Evaluate(context)
	if err != nil {
		return 0, err
	}
	return scalarValue.ToScalar(e.QueryString())
}

// EvaluateToDuration is a helper function that takes an Expression and makes it a duration.
func EvaluateToDuration(e Expression, context EvaluationContext) (time.Duration, error) {
	durationValue, err := e.Evaluate(context)
	if err != nil {
		return 0, err
	}
	return durationValue.ToDuration(e.QueryString())
}

// EvaluateToDuration is a helper function that takes an Expression and makes it a series list.
func EvaluateToSeriesList(e Expression, context EvaluationContext) (api.SeriesList, error) {
	seriesValue, err := e.Evaluate(context)
	if err != nil {
		return api.SeriesList{}, err
	}
	return seriesValue.ToSeriesList(context.Timerange, e.QueryString())
}

// EvaluateToDuration is a helper function that takes an Expression and makes it a string.
func EvaluateToString(e Expression, context EvaluationContext) (string, error) {
	stringValue, err := e.Evaluate(context)
	if err != nil {
		return "", err
	}
	return stringValue.ToString(e.QueryString())
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
		} else {
			array[result.index] = result.value
		}
	}
	return array, nil

}
