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
	"regexp"
	"sort"

	"github.com/square/metrics/api"
	"github.com/square/metrics/metric_metadata"
)

// DescribeAllCommand returns all the metrics available in the system.
type DescribeAllCommand struct {
	Matcher *regexp.Regexp
}

// Execute of a DescribeAllCommand returns the list of all metrics.
func (cmd *DescribeAllCommand) Execute(context ExecutionContext) (CommandResult, error) {
	result, err := context.MetricMetadataAPI.GetAllMetrics(metadata.Context{
		Profiler: context.Profiler,
	})
	if err == nil {
		filtered := make([]api.MetricKey, 0, len(result))
		for _, row := range result {
			if cmd.Matcher.MatchString(string(row)) {
				filtered = append(filtered, row)
			}
		}
		sort.Sort(api.MetricKeys(filtered))
		return CommandResult{
			Body: filtered,
			Metadata: map[string]interface{}{
				"count": len(filtered),
			},
		}, nil
	}
	return CommandResult{}, err
}

func (cmd *DescribeAllCommand) Name() string {
	return "describe all"
}
