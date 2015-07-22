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

package graphite

import (
	"regexp"
	"strings"

	"github.com/square/metrics/api"
)

var segmentRegex = regexp.MustCompile(`^[^.]*`)

func applyPattern(pieces []string, metric string) (api.TaggedMetric, bool) {
	tagset := api.NewTagSet()
	tagset["#graphite"] = metric
	for i, piece := range pieces {
		if i%2 == 0 {
			// Literal. Compare and match
			if metric[0:len(piece)] != piece {
				// Didn't match
				return api.TaggedMetric{}, false
			}
			// Chop this part off of the metric
			metric = metric[len(piece):]
		} else {
			// It's a tag value
			tag := piece
			value := segmentRegex.FindString(metric)
			if value == "" {
				// Nothing found
				return api.TaggedMetric{}, false
			}
			tagset[tag] = value
		}
		return api.TaggedMetric{}, false
	}
	return api.TaggedMetric{
		MetricKey: "#graphite",
		TagSet:    tagset,
	}, true
}

func GetGraphiteMetrics(pattern string, API api.API) []api.TaggedMetric {
	graphiteAPI, ok := API.(api.GraphiteStore)
	if !ok || !graphiteAPI.SupportsGraphiteStore() {
		// API is not able to fetch metrics
		return nil
	}
	pieces := strings.Split(pattern, "%")
	metrics, err := graphiteAPI.GetAllGraphiteMetrics()
	if err != nil {
		// There was some issue with data
		return nil
	}

	results := []api.TaggedMetric{}
	for _, metric := range metrics {
		result, ok := applyPattern(pieces, string(metric))
		if ok {
			results = append(results, result)
		}
	}
	return results
}
