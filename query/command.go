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
	"github.com/square/metrics/api"
)

// Command is the final result of the parsing.
// A command contains all the information to execute the
// given query against the API.
type Command interface {
	// Execute the given command. Returns JSON-encodable result or an error.
	Execute(a api.API) (interface{}, error)
	// Name is the human-readable identifier for the command.
	Name() string
}

// DescribeCommand describes the tag set managed by the given metric indexer.
type DescribeCommand struct {
	metricName api.MetricKey
	predicate  Predicate
}

// Execute returns the list of tags satisfying the provided predicate.
func (cmd *DescribeCommand) Execute(a api.API) (interface{}, error) {
	tags, _ := a.GetAllTags(cmd.metricName)
	output := make([]string, 0, len(tags))
	for _, tag := range tags {
		if cmd.predicate.Apply(tag) {
			output = append(output, tag.Serialize())
		}
	}
	return output, nil
}

// Name of the command
func (cmd *DescribeCommand) Name() string {
	return "describe"
}

// DescribeAllCommand returns all the metrics available in the system.
type DescribeAllCommand struct {
}

// Execute of a DescribeAllCommand returns the list of all metrics.
func (cmd *DescribeAllCommand) Execute(a api.API) (interface{}, error) {
	return a.GetAllMetrics()
}

// Name of the command
func (cmd *DescribeAllCommand) Name() string {
	return "describe all"
}
