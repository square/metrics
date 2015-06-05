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
	"github.com/square/metrics/api"
)

// Command is the final result of the parsing.
// A command contains all the information to execute the
// given query against the API.
type Command interface {
	// Execute the given command. Returns JSON-encodable result or an error.
	Execute(b api.Backend) (interface{}, error)
	// Name is the human-readable identifier for the command.
	Name() string
	String() string
}

// DescribeCommand describes the tag set managed by the given metric indexer.
type DescribeCommand struct {
	metricName api.MetricKey
	predicate  api.Predicate
}

func (cmd *DescribeCommand) String() string {
	return fmt.Sprintf("describe %s: %s", string(cmd.metricName), PrintNode(cmd.predicate))
}

// DescribeAllCommand returns all the metrics available in the system.
type DescribeAllCommand struct {
}

func (cmd *DescribeAllCommand) String() string {
	return "describe all"
}

// SelectCommand is the bread and butter of the metrics query engine.
// It actually performs the query against the underlying metrics system.
type SelectCommand struct {
	predicate   api.Predicate
	expressions []Expression
	context     *evaluationContextNode
}

func (cmd *SelectCommand) String() string {
	return fmt.Sprintf("select{context: %+v, expressions: %+v, predicate: %s}", cmd.context, cmd.expressions, PrintNode(cmd.predicate))
}

// Execute returns the list of tags satisfying the provided predicate.
func (cmd *DescribeCommand) Execute(b api.Backend) (interface{}, error) {
	tags, _ := b.Api().GetAllTags(cmd.metricName)
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

// Execute of a DescribeAllCommand returns the list of all metrics.
func (cmd *DescribeAllCommand) Execute(b api.Backend) (interface{}, error) {
	return b.Api().GetAllMetrics()
}

// Name of the command
func (cmd *DescribeAllCommand) Name() string {
	return "describe all"
}

// Name of the command
func (cmd *SelectCommand) Name() string {
	return "select"
}

// Execute performs the query represented by the given query string, and returs the result.
func (cmd *SelectCommand) Execute(b api.Backend) (interface{}, error) {
	return evaluateExpressions(EvaluationContext{
		Backend:      b,
		Timerange:    cmd.context.Timerange,
		SampleMethod: cmd.context.SampleMethod,
		Predicate:    cmd.predicate,
	}, cmd.expressions)
}
