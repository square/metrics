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
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
	"github.com/square/metrics/inspect"
	"github.com/square/metrics/metric_metadata"
	"github.com/square/metrics/query/predicate"
	"github.com/square/metrics/timeseries"
)

// ExecutionContext is the context supplied when invoking a command.
type ExecutionContext struct {
	TimeseriesStorageAPI  timeseries.StorageAPI            // the backend
	MetricMetadataAPI     metadata.MetricAPI               // the api
	FetchLimit            int                              // the maximum number of fetches
	Timeout               time.Duration                    // optional
	Registry              function.Registry                // optional
	SlotLimit             int                              // optional (0 => default 1000)
	Profiler              *inspect.Profiler                // optional
	UserSpecifiableConfig timeseries.UserSpecifiableConfig // optional. User tunable parameters for execution.
	AdditionalConstraints predicate.Predicate              // optional. Additional contrains for describe and select commands
}

type CommandResult struct {
	Body     interface{}
	Metadata map[string]interface{}
}

// Command is the final result of the parsing.
// A command contains all the information to execute the
// given query against the API.
type Command interface {
	// Execute the given command. Returns JSON-encodable result or an error.
	Execute(ExecutionContext) (CommandResult, error)
	Name() string
}

type QueryResult struct {
	Query string `json:"query"`
	Name  string `json:"name"`
	Type  string `json:"type"` // one of "series" or "scalars"
	// for "series" type
	Series    []api.Timeseries `json:"series,omitempty"`
	Timerange api.Timerange    `json:"timerange,omitempty"`
	// for "scalar" type
	Scalars []function.TaggedScalar `json:"scalars,omitempty"`
}

type profileJSON struct {
	Name   string `json:"name"`
	Start  int64  `json:"start"`  // ms since Unix epoch
	Finish int64  `json:"finish"` // ms since Unix epoch
}
