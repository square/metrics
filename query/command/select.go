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

package command

import (
	"fmt"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
	"github.com/square/metrics/function/registry"
	"github.com/square/metrics/query/natural_sort"
	"github.com/square/metrics/query/predicate"
	"github.com/square/metrics/tasks"
	"github.com/square/metrics/timeseries"
)

type SelectContext struct {
	Start        int64                   // Start of data timerange
	End          int64                   // End of data timerange
	Resolution   int64                   // Resolution of data timerange
	SampleMethod timeseries.SampleMethod // to use when up/downsampling to match requested resolution
}

// SelectCommand is the bread and butter of the metrics query engine.
// It actually performs the query against the underlying metrics system.
type SelectCommand struct {
	Predicate   predicate.Predicate
	Expressions []function.Expression
	Context     SelectContext
}

// Execute performs the query represented by the given query string, and returs the result.
func (cmd *SelectCommand) Execute(context ExecutionContext) (CommandResult, error) {
	userTimerange, err := api.NewSnappedTimerange(cmd.Context.Start, cmd.Context.End, cmd.Context.Resolution)
	if err != nil {
		return CommandResult{}, err
	}
	slotLimit := context.SlotLimit
	defaultLimit := 1000
	if slotLimit == 0 {
		slotLimit = defaultLimit // the default limit
	}

	smallestResolution := userTimerange.Duration() / time.Duration(slotLimit-2)
	// ((end + res/2) - (start - res/2)) / res + 1 <= slots // make adjustments for a snap that moves the endpoints
	// (do some algebra)
	// (end - start + res) + res <= slots * res
	// end - start <= res * (slots - 2)
	// so
	// res >= (end - start) / (slots - 2)

	// Update the timerange by applying the insights of the storage API:
	chosenResolution, err := context.TimeseriesStorageAPI.ChooseResolution(userTimerange, smallestResolution)
	if err != nil {
		return CommandResult{}, err
	}

	chosenTimerange, err := api.NewSnappedTimerange(userTimerange.StartMillis(), userTimerange.EndMillis(), int64(chosenResolution/time.Millisecond))
	if err != nil {
		return CommandResult{}, err
	}

	if chosenTimerange.Slots() > slotLimit {
		return CommandResult{}, function.NewLimitError(
			"Requested number of data points exceeds the configured limit",
			chosenTimerange.Slots(), slotLimit)
	}
	hasTimeout := context.Timeout != 0
	var timeoutOwner tasks.TimeoutOwner
	if hasTimeout {
		timeoutOwner = tasks.NewTimeout(context.Timeout)
	}
	r := context.Registry
	if r == nil {
		r = registry.Default()
	}

	defer timeoutOwner.Finish() // broadcast the finish - this ensures that the future work is cancelled.
	evaluationContext := function.EvaluationContext{
		MetricMetadataAPI:     context.MetricMetadataAPI,
		FetchLimit:            function.NewFetchCounter(context.FetchLimit),
		TimeseriesStorageAPI:  context.TimeseriesStorageAPI,
		Predicate:             predicate.All(cmd.Predicate, context.AdditionalConstraints),
		SampleMethod:          cmd.Context.SampleMethod,
		Timerange:             chosenTimerange,
		Timeout:               timeoutOwner.Timeout(),
		Registry:              r,
		Profiler:              context.Profiler,
		EvaluationNotes:       new(function.EvaluationNotes),
		UserSpecifiableConfig: context.UserSpecifiableConfig,
	}

	timeout := (<-chan time.Time)(nil)
	if hasTimeout {
		// A nil channel will just block forever
		timeout = time.After(context.Timeout)
	}

	results := make(chan []function.Value, 1)
	errors := make(chan error, 1)
	// Goroutines are never garbage collected, so we need to provide capacity so that the send always succeeds.
	go func() {
		// Evaluate the result, and send it along the goroutines.
		result, err := function.EvaluateMany(evaluationContext, cmd.Expressions)
		if err != nil {
			errors <- err
			return
		}
		results <- result
	}()
	select {
	case <-timeout:
		return CommandResult{}, function.NewLimitError("Timeout while executing the query.",
			context.Timeout, context.Timeout)
	case err := <-errors:
		return CommandResult{}, err
	case result := <-results:
		description := map[string][]string{}
		for _, value := range result {
			listValue, err := value.ToSeriesList(evaluationContext.Timerange)
			if err != nil {
				continue
			}
			list := api.SeriesList(listValue)
			for _, series := range list.Series {
				for key, value := range series.TagSet {
					description[key] = append(description[key], value)
				}
			}
		}
		for key, values := range description {
			natural_sort.Sort(values)
			filtered := []string{}
			for i := range values {
				if i == 0 || values[i-1] != values[i] {
					filtered = append(filtered, values[i])
				}
			}
			description[key] = filtered
		}

		// Body adds the Query as an annotation.
		// It's a slice of interfaces; it will be cast to an interface
		// when returned from this function in a CommandResult.
		body := make([]QueryResult, len(result))
		for i := range body {
			if list, ok := result[i].(function.SeriesListValue); ok {
				body[i] = QueryResult{
					Query:     cmd.Expressions[i].QueryString(),
					Name:      cmd.Expressions[i].Name(),
					Type:      "series",
					Series:    list.Series,
					Timerange: chosenTimerange,
				}
				continue
			}
			if scalars, err := result[i].ToScalarSet(); err == nil {
				body[i] = QueryResult{
					Query:   cmd.Expressions[i].QueryString(),
					Name:    cmd.Expressions[i].Name(),
					Type:    "scalars",
					Scalars: scalars,
				}
				continue
			}
			return CommandResult{}, fmt.Errorf("Query %s does not result in a timeseries or scalar.", cmd.Expressions[i].QueryString())
		}

		return CommandResult{
			Body: body,
			Metadata: map[string]interface{}{
				"description": description,
				"notes":       evaluationContext.EvaluationNotes,
			},
		}, nil
	}
}

func (cmd *SelectCommand) Name() string {
	return "select"
}
