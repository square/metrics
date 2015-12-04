package function

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/inspect"
	"github.com/square/metrics/optimize"
)

// EvaluationContextInternals contains fields which cannot be changed or configured during evaluation.
// In order to reduce the size of the struct, they're stored outside it, as a pointer.
type EvaluationContextInternals struct {
	TimeseriesStorageAPI      api.TimeseriesStorageAPI // Backend to fetch data from
	MetricMetadataAPI         api.MetricMetadataAPI    // Api to obtain metadata from
	SampleMethod              api.SampleMethod         // SampleMethod to use when up/downsampling to match the requested resolution
	Predicate                 api.Predicate            // Predicate to apply to TagSets prior to fetching
	FetchLimit                FetchCounter             // A limit on the number of fetches which may be performed
	Cancellable               api.Cancellable
	Registry                  Registry
	Profiler                  *inspect.Profiler // A profiler pointer
	OptimizationConfiguration *optimize.OptimizationConfiguration
	EvaluationNotes           *EvaluationNotes //Debug + numerical notes that can be added during evaluation
}

// EvaluationContext is the central piece of logic, providing
// Only the Timerange and UserSpecifiableConfig are public. They can be modified during execution,
// but the other internal fields cannot. (For example, the sampling resolution or the timeseries API cannot be changed).
// All of the internal values are exposed as methods, which effectively makes them read-only.
type EvaluationContext struct {
	Timerange             api.Timerange // Timerange to fetch data from
	UserSpecifiableConfig api.UserSpecifiableConfig
	internal              *EvaluationContextInternals
}

func CreateEvaluationContext(timerange api.Timerange, config api.UserSpecifiableConfig, internal EvaluationContextInternals) EvaluationContext {
	return EvaluationContext{
		Timerange:             timerange,
		UserSpecifiableConfig: config,
		internal:              &internal,
	}
}

func (e EvaluationContext) Predicate() api.Predicate {
	return e.internal.Predicate
}
func (e EvaluationContext) FetchLimitConsume(n int) error {
	ok := e.internal.FetchLimit.Consume(n)
	if ok {
		return nil
	}
	return NewLimitError("fetch limit exceeded: too many series to fetch", e.internal.FetchLimit.Current(), e.internal.FetchLimit.Limit())
}
func (e EvaluationContext) GetFunction(name string) (MetricFunction, bool) {
	return e.internal.Registry.GetFunction(name)
}
func (e EvaluationContext) OptimizationConfiguration() *optimize.OptimizationConfiguration {
	return e.internal.OptimizationConfiguration
}
func (e EvaluationContext) GetAllTags(metricKey api.MetricKey) ([]api.TagSet, error) {
	context := api.MetricMetadataAPIContext{
		Profiler: e.internal.Profiler,
	}
	return e.internal.MetricMetadataAPI.GetAllTags(metricKey, context)
}
func (e EvaluationContext) FetchMultipleTimeseries(metrics []api.TaggedMetric) (api.SeriesList, error) {
	request := api.FetchMultipleTimeseriesRequest{
		metrics,
		e.internal.SampleMethod,
		e.Timerange,
		e.internal.MetricMetadataAPI,
		e.internal.Cancellable,
		e.internal.Profiler,
		e.UserSpecifiableConfig,
	}
	return e.internal.TimeseriesStorageAPI.FetchMultipleTimeseries(request)
}

type EvaluationNotes struct {
	mutex sync.Mutex
	notes []string
}

func (notes *EvaluationNotes) AddNote(note string) {
	if notes == nil {
		return
	}
	notes.mutex.Lock()
	defer notes.mutex.Unlock()
	notes.notes = append(notes.notes, note)
}
func (notes *EvaluationNotes) Notes() []string {
	if notes == nil {
		return nil
	}
	notes.mutex.Lock()
	defer notes.mutex.Unlock()
	return notes.notes
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
	Compute       func(EvaluationContext, []Expression, Groups) (Value, error)
}

func (e EvaluationContext) AddNote(note string) {
	if e.internal == nil {
		return // drop the note
	}
	e.internal.EvaluationNotes.AddNote(note)
}

func (e EvaluationContext) Notes() []string {
	return e.internal.EvaluationNotes.Notes()
}

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
	Evaluate(context EvaluationContext) (Value, error)
}

func EvaluateToScalar(e Expression, context EvaluationContext) (float64, error) {
	scalarValue, err := e.Evaluate(context)
	if err != nil {
		return 0, err
	}
	return scalarValue.ToScalar()
}
func EvaluateToDuration(e Expression, context EvaluationContext) (time.Duration, error) {
	scalarValue, err := e.Evaluate(context)
	if err != nil {
		return 0, err
	}
	return scalarValue.ToDuration()
}
func EvaluateToSeriesList(e Expression, context EvaluationContext) (api.SeriesList, error) {
	seriesValue, err := e.Evaluate(context)
	if err != nil {
		return api.SeriesList{}, err
	}
	return seriesValue.ToSeriesList(context.Timerange)
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
}
