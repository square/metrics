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
	"github.com/square/metrics/inspect"
)

// ExecutionContext is the context supplied when invoking a command.
type ExecutionContext struct {
	Backend    api.MultiBackend // the backend
	API        api.API          // the api
	FetchLimit int              // the maximum number of fetches
	Timeout    time.Duration
	Profiler   *inspect.Profiler
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
	expressions []Expression
	context     *evaluationContextNode
}

// Execute returns the list of tags satisfying the provided predicate.
func (cmd *DescribeCommand) Execute(context ExecutionContext) (interface{}, error) {
	tags, _ := context.API.GetAllTags(cmd.metricName)
	output := make([]string, 0, len(tags))
	for _, tag := range tags {
		if cmd.predicate.Apply(tag) {
			output = append(output, tag.Serialize())
		}
	}
	sort.Strings(output)
	return output, nil
}
func (cmd *DescribeCommand) Name() string {
	return "describe"
}

// Execute of a DescribeAllCommand returns the list of all metrics.
func (cmd *DescribeAllCommand) Execute(context ExecutionContext) (interface{}, error) {
	result, err := context.API.GetAllMetrics()
	if err == nil {
		sort.Sort(api.MetricKeys(result))
	}
	return result, err
}

func (cmd *DescribeAllCommand) Name() string {
	return "describe all"
}

// Execute asks for all metrics with the given name.
func (cmd *DescribeMetricsCommand) Execute(context ExecutionContext) (interface{}, error) {
	return context.API.GetMetricsForTag(cmd.tagKey, cmd.tagValue)
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
	hasTimeout := context.Timeout != 0
	var cancellable api.Cancellable
	if hasTimeout {
		cancellable = api.NewTimeoutCancellable(time.Now().Add(context.Timeout))
	} else {
		cancellable = api.NewCancellable()
	}

	defer close(cancellable.Done()) // broadcast the finish - this ensures that the future work is cancelled.
	evaluationContext := EvaluationContext{
		API:          context.API,
		FetchLimit:   NewFetchCounter(context.FetchLimit),
		MultiBackend: context.Backend,
		Predicate:    cmd.predicate,
		SampleMethod: cmd.context.SampleMethod,
		Timerange:    timerange,
		Cancellable:  cancellable,
		Profiler:     context.Profiler,
	}
	if hasTimeout {
		timeout := time.After(context.Timeout)
		results := make(chan interface{})
		errors := make(chan error)
		go func() {
			result, err := evaluateExpressions(evaluationContext, cmd.expressions)
			if err != nil {
				errors <- err
			} else {
				results <- result
			}
		}()
		select {
		case <-timeout:
			return nil, fmt.Errorf("Timeout while executing the query.") // timeout.
		case result := <-results:
			return result, nil
		case err := <-errors:
			return nil, err
		}
	} else {
		return evaluateExpressions(evaluationContext, cmd.expressions)
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
	defer cmd.Profiler.Record(fmt.Sprintf("%s.Execute", cmd.Name()))
	context.API = api.ProfilingAPI{
		Profiler: cmd.Profiler,
		API:      context.API,
	}
	context.Profiler = cmd.Profiler
	return cmd.Command.Execute(context)
}
