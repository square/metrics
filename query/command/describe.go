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
	"github.com/square/metrics/api"
	"github.com/square/metrics/metric_metadata"
	"github.com/square/metrics/query/natural_sort"
	"github.com/square/metrics/query/predicate"
)

// DescribeCommand describes the tag set managed by the given metric indexer.
type DescribeCommand struct {
	MetricName api.MetricKey
	Predicate  predicate.Predicate
}

// Execute returns the list of tags satisfying the provided predicate.
func (cmd *DescribeCommand) Execute(context ExecutionContext) (CommandResult, error) {
	// We generate a simple update function that closes around the profiler
	// so if we do have a cache miss it's correctly reported on this request.

	tagsets, err := context.MetricMetadataAPI.GetAllTags(cmd.MetricName, metadata.Context{
		Profiler: context.Profiler,
	})
	if err != nil {
		return CommandResult{}, err
	}

	// Splitting each tag key into its own set of values is helpful for discovering actual metrics.
	predicate := predicate.All(cmd.Predicate, context.AdditionalConstraints)
	keyValueSets := map[string]map[string]bool{} // a map of tag_key => Set{tag_value}.
	for _, tagset := range tagsets {
		if predicate.Apply(tagset) {
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
	return CommandResult{Body: keyValueLists}, nil
}

func (cmd *DescribeCommand) Name() string {
	return "describe"
}
