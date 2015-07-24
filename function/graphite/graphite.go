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
	"github.com/square/metrics/api"
	"github.com/square/metrics/internal"
)

func applyPattern(rule internal.Rule, metric string) (api.TaggedMetric, bool) {
	tagged, ok := rule.MatchRule(metric)
	if !ok {
		return api.TaggedMetric{}, ok
	}
	tagged.TagSet[api.SpecialGraphiteName] = metric
	return tagged, true
}

func GetGraphiteMetrics(pattern string, API api.API) ([]api.TaggedMetric, error) {
	rule, err := internal.CompileGraphiteRule(pattern)
	if err != nil {
		return nil, err
	}
	metrics, err := API.GetAllGraphiteMetrics()
	if err != nil {
		// There was some issue with data
		return nil, err
	}
	results := []api.TaggedMetric{}
	for _, metric := range metrics {
		result, ok := applyPattern(rule, string(metric))
		if ok {
			results = append(results, result)
		}
	}
	return results, nil
}
