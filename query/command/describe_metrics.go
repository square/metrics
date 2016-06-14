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

import "github.com/square/metrics/metric_metadata"

// DescribeMetricsCommand returns all metrics that use a particular key-value pair.
type DescribeMetricsCommand struct {
	TagKey   string
	TagValue string
}

// Execute asks for all metrics with the given name.
func (cmd *DescribeMetricsCommand) Execute(context ExecutionContext) (CommandResult, error) {
	data, err := context.MetricMetadataAPI.GetMetricsForTag(cmd.TagKey, cmd.TagValue, metadata.Context{
		Profiler: context.Profiler,
	})
	if err != nil {
		return CommandResult{}, err
	}
	return CommandResult{
		Body: data,
		Metadata: map[string]interface{}{
			"count": len(data),
		},
	}, nil
}

func (cmd *DescribeMetricsCommand) Name() string {
	return "describe metrics"
}
