package function

import (
	"sync/atomic"

	"github.com/square/metrics/api"
	"github.com/square/metrics/inspect"
)

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
	FetchLimit   FetchCounter     // A limit on the number of fetches which may be performed
	Cancellable  api.Cancellable
	Profiler     *inspect.Profiler
}

// fetchCounter is used to count the number of fetches remaining in a thread-safe manner.
type FetchCounter struct {
	count *int32
}

func NewFetchCounter(n int) FetchCounter {
	n32 := int32(n)
	return FetchCounter{
		count: &n32,
	}
}

// Consume decrements the internal counter and returns whether the result is at least 0.
// It does so in a threadsafe manner.
func (c FetchCounter) Consume(n int) bool {
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
	Evaluate(context EvaluationContext) (Value, error)
}
