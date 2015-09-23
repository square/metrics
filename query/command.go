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
	"sort"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
	"github.com/square/metrics/function/registry"
	"github.com/square/metrics/inspect"
	"github.com/square/metrics/optimize"
	"github.com/square/metrics/query/natural_sort"
)

// ExecutionContext is the context supplied when invoking a command.
type ExecutionContext struct {
	TimeseriesStorageAPI      api.TimeseriesStorageAPI            // the backend
	MetricMetadataAPI         api.MetricMetadataAPI               // the api
	FetchLimit                int                                 // the maximum number of fetches
	Timeout                   time.Duration                       // optional
	Registry                  function.Registry                   // optional
	SlotLimit                 int                                 // optional (0 => default 1000)
	Profiler                  *inspect.Profiler                   // optional
	OptimizationConfiguration *optimize.OptimizationConfiguration // optional
}

// Command is the final result of the parsing.
// A command contains all the information to execute the
// given query against the API.
type Command interface {
	// Execute the given command. Returns JSON-encodable result or an error.
	Execute(ExecutionContext) (interface{}, error)
	Name() string
}

// DescribeCommand describes the tag set managed by the given metric indexer.
type DescribeCommand struct {
	metricName api.MetricKey
	predicate  api.Predicate
}

// DescribeAllCommand returns all the metrics available in the system.
type DescribeAllCommand struct {
	matcher *matcherClause
}

// DescribeMetricsCommand returns all metrics that use a particular key-value pair.
type DescribeMetricsCommand struct {
	tagKey   string
	tagValue string
}

// SelectCommand is the bread and butter of the metrics query engine.
// It actually performs the query against the underlying metrics system.
type SelectCommand struct {
	predicate   api.Predicate
	expressions []function.Expression
	context     *evaluationContextNode
}

// Execute returns the list of tags satisfying the provided predicate.
func (cmd *DescribeCommand) Execute(context ExecutionContext) (interface{}, error) {

	// We generate a simple update function that closes around the profiler
	// so if we do have a cache miss it's correctly reported on this request.
	updateFunction := func() ([]api.TagSet, error) {
		tagsets, err := context.MetricMetadataAPI.GetAllTags(cmd.metricName, api.MetricMetadataAPIContext{
			Profiler: context.Profiler,
		})
		return tagsets, err
	}
	tagsets, _ := context.OptimizationConfiguration.AllTagsCacheHitOrExecute(cmd.metricName, updateFunction)

	// Splitting each tag key into its own set of values is helpful for discovering actual metrics.
	keyValueSets := map[string]map[string]bool{} // a map of tag_key => Set{tag_value}.
	for _, tagset := range tagsets {
		if cmd.predicate.Apply(tagset) {
			// Add each key as needed
			for key, value := range tagset {
				if keyValueSets[key] == nil {
					keyValueSets[key] = map[string]bool{}
				}
				keyValueSets[key][value] = true // add `value` to the set for `key`
			}
		}
	}
	keyValueLists := map[string][]string{} // a map of tag_key => list[tag_value]
	for key, set := range keyValueSets {
		list := make([]string, 0, len(set))
		for value := range set {
			list = append(list, value)
		}
		// sort the result
		natural_sort.Sort(list)
		keyValueLists[key] = list
	}
	return keyValueLists, nil
}

func (cmd *DescribeCommand) Name() string {
	return "describe"
}

// Execute of a DescribeAllCommand returns the list of all metrics.
func (cmd *DescribeAllCommand) Execute(context ExecutionContext) (interface{}, error) {
	result, err := context.MetricMetadataAPI.GetAllMetrics(api.MetricMetadataAPIContext{
		Profiler: context.Profiler,
	})
	if err == nil {
		filtered := make([]api.MetricKey, 0, len(result))
		for _, row := range result {
			if cmd.matcher.regex.MatchString(string(row)) {
				filtered = append(filtered, row)
			}
		}
		sort.Sort(api.MetricKeys(filtered))
		return filtered, nil
	}
	return nil, err
}

func (cmd *DescribeAllCommand) Name() string {
	return "describe all"
}

// Execute asks for all metrics with the given name.
func (cmd *DescribeMetricsCommand) Execute(context ExecutionContext) (interface{}, error) {
	return context.MetricMetadataAPI.GetMetricsForTag(cmd.tagKey, cmd.tagValue, api.MetricMetadataAPIContext{
		Profiler: context.Profiler,
	})
}

func (cmd *DescribeMetricsCommand) Name() string {
	return "describe metrics"
}

// Execute performs the query represented by the given query string, and returs the result.
func (cmd *SelectCommand) Execute(context ExecutionContext) (interface{}, error) {
	timerange, err := api.NewSnappedTimerange(cmd.context.Start, cmd.context.End, cmd.context.Resolution)
	if err != nil {
		return nil, err
	}
	slotLimit := context.SlotLimit
	defaultLimit := 1000
	if slotLimit == 0 {
		slotLimit = defaultLimit // the default limit
	}
	if timerange.Slots() > slotLimit {
		return nil, function.NewLimitError(
			"Requested number of data points exceeds the configured limit",
			timerange.Slots(), slotLimit)
	}
	hasTimeout := context.Timeout != 0
	var cancellable api.Cancellable
	if hasTimeout {
		cancellable = api.NewTimeoutCancellable(time.Now().Add(context.Timeout))
	} else {
		cancellable = api.NewCancellable()
	}
	r := context.Registry
	if r == nil {
		r = registry.Default()
	}

	defer close(cancellable.Done()) // broadcast the finish - this ensures that the future work is cancelled.
	evaluationContext := function.EvaluationContext{
		MetricMetadataAPI:         context.MetricMetadataAPI,
		FetchLimit:                function.NewFetchCounter(context.FetchLimit),
		TimeseriesStorageAPI:      context.TimeseriesStorageAPI,
		Predicate:                 cmd.predicate,
		SampleMethod:              cmd.context.SampleMethod,
		Timerange:                 timerange,
		Cancellable:               cancellable,
		Registry:                  r,
		Profiler:                  context.Profiler,
		OptimizationConfiguration: context.OptimizationConfiguration,
	}

	if hasTimeout {
		timeout := time.After(context.Timeout)
		results := make(chan interface{})
		errors := make(chan error)
		go func() {
			result, err := function.EvaluateMany(evaluationContext, cmd.expressions)
			if err != nil {
				errors <- err
			} else {
				results <- result
			}
		}()
		select {
		case <-timeout:
			return nil, function.NewLimitError("Timeout while executing the query.",
				context.Timeout, context.Timeout)
		case result := <-results:
			return result, nil
		case err := <-errors:
			return nil, err
		}
	} else {
		values, err := function.EvaluateMany(evaluationContext, cmd.expressions)
		if err != nil {
			return nil, err
		}
		lists := make([]api.SeriesList, len(values))
		for i := range values {
			lists[i], err = values[i].ToSeriesList(evaluationContext.Timerange)
			if err != nil {
				return nil, err
			}
		}
		return lists, nil
	}
}

func (cmd *SelectCommand) Name() string {
	return "select"
}

//ProfilingCommand is a Command that also performs profiling actions.
type ProfilingCommand struct {
	Profiler *inspect.Profiler
	Command  Command
}

func NewProfilingCommand(command Command) (Command, *inspect.Profiler) {
	profiler := inspect.New()
	return ProfilingCommand{
		Profiler: profiler,
		Command:  command,
	}, profiler
}

func (cmd ProfilingCommand) Name() string {
	return cmd.Command.Name()
}

func (cmd ProfilingCommand) Execute(context ExecutionContext) (interface{}, error) {
	defer cmd.Profiler.Record(fmt.Sprintf("%s.Execute", cmd.Name()))()
	context.Profiler = cmd.Profiler
	return cmd.Command.Execute(context)
}
