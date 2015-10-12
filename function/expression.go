package function

import (
	"fmt"
	"sync/atomic"

	"github.com/square/metrics/api"
	"github.com/square/metrics/inspect"
	"github.com/square/metrics/optimize"
)

// EvaluationContext is the central piece of logic, providing
// helper funcions & varaibles to evaluate a given piece of
// metrics query.
// * Contains a TimeseriesStorageAPI object, which can be used to fetch data
// from the backend systems.
// * Contains current timerange being queried for - this can be
// changed by say, application of time shift function.
type EvaluationContext struct {
	TimeseriesStorageAPI      api.TimeseriesStorageAPI // Backend to fetch data from
	MetricMetadataAPI         api.MetricMetadataAPI    // Api to obtain metadata from
	Timerange                 api.Timerange            // Timerange to fetch data from
	SampleMethod              api.SampleMethod         // SampleMethod to use when up/downsampling to match the requested resolution
	Predicate                 api.Predicate            // Predicate to apply to TagSets prior to fetching
	FetchLimit                FetchCounter             // A limit on the number of fetches which may be performed
	Cancellable               api.Cancellable
	Registry                  Registry
	Profiler                  *inspect.Profiler // A profiler pointer
	OptimizationConfiguration *optimize.OptimizationConfiguration
	EvaluationNotes           []string //Debug + numerical notes that can be added during evaluation
	invalid                   bool     // Because these can be copied, it's best to mark a no-longer used context as dead
	UserSpecifiableConfig     api.UserSpecifiableConfig
}

type Registry interface {
	GetFunction(string) (MetricFunction, bool) // returns an instance of MetricFunction
	All() []string                             // all the registered functions
}

// Groups holds grouping information - which tags to group by (if any), and whether to `collapse` (Collapses = true) or `group` (Collapses = false)
type Groups struct {
	List      []string
	Collapses bool
}

// MetricFunction defines a common logic to dispatch a function in MQE.
type MetricFunction struct {
	Name          string
	MinArguments  int
	MaxArguments  int
	AllowsGroupBy bool // Whether the function allows a 'group by' clause.
	Compute       func(*EvaluationContext, []Expression, Groups) (Value, error)
}

func (e *EvaluationContext) Copy() EvaluationContext {
	return EvaluationContext{
		MetricMetadataAPI:         e.MetricMetadataAPI,
		FetchLimit:                e.FetchLimit,
		TimeseriesStorageAPI:      e.TimeseriesStorageAPI,
		Predicate:                 e.Predicate,
		SampleMethod:              e.SampleMethod,
		Timerange:                 e.Timerange,
		Cancellable:               e.Cancellable,
		Registry:                  e.Registry,
		Profiler:                  e.Profiler,
		OptimizationConfiguration: e.OptimizationConfiguration,
		EvaluationNotes:           []string{},
		invalid:                   false,
		UserSpecifiableConfig:     e.UserSpecifiableConfig,
	}
}

func (e *EvaluationContext) AddNote(note string) {
	if e.EvaluationNotes == nil {
		e.EvaluationNotes = []string{}
	}
	e.EvaluationNotes = append(e.EvaluationNotes, note)
}

func (e *EvaluationContext) CopyNotesFrom(other *EvaluationContext) {
	if e.EvaluationNotes == nil {
		e.EvaluationNotes = []string{}
	}
	if len(other.EvaluationNotes) > 0 {
		e.EvaluationNotes = append(e.EvaluationNotes, other.EvaluationNotes...)
	}
}

func (e *EvaluationContext) Invalidate() {
	e.invalid = true
}

// Evaluate the given metric function.
func (f MetricFunction) Evaluate(context *EvaluationContext,
	arguments []Expression, groupBy []string, collapses bool) (Value, error) {
	if context.invalid {
		panic("Attempted to evaluate a function on an EvaluationContext that's been explicitly invalidated.")
	}
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

func (c FetchCounter) Current() int {
	return c.limit - int(atomic.LoadInt32(c.count))
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
	Evaluate(context *EvaluationContext) (Value, error)
}

// EvaluateMany evaluates a list of expressions using a single EvaluationContext.
// If any evaluation errors, EvaluateMany will propagate that error. The resulting values
// will be in the order corresponding to the provided expressions.
func EvaluateMany(context *EvaluationContext, expressions []Expression) ([]Value, error) {
	if context.invalid {
		panic("Attempted to evaluate a function on an EvaluationContext that's been explicitly invalidated.")
	}

	type result struct {
		index int
		err   error
		value Value
	}
	length := len(expressions)
	if length == 0 {
		return []Value{}, nil
	} else if length == 1 {
		result, err := expressions[0].Evaluate(context)
		if err != nil {
			return nil, err
		}
		return []Value{result}, nil
	} else {
		// concurrent evaluations
		results := make(chan result, length)
		for i, expr := range expressions {
			go func(i int, expr Expression) {
				value, err := expr.Evaluate(context)
				results <- result{i, err, value}
			}(i, expr)
		}
		array := make([]Value, length)
		for result := range results {
			if result.err != nil {
				return nil, result.err
			}
			array[result.index] = result.value
		}
		return array, nil
	}
}
