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

// Integration test for the query execution.
package tests

import (
	"testing"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/interface/query/command"
	"github.com/square/metrics/interface/query/parser"
	"github.com/square/metrics/testing_support/assert"
	"github.com/square/metrics/testing_support/mocks"
)

func TestCommandSelectFilterRange(t *testing.T) {
	a := assert.New(t)
	timerange, err := api.NewSnappedTimerange(3000000, 3270000, 30000) // 10 slots
	if err != nil {
		t.Fatalf("Error constructing test timerange: %s", err.Error())
	}
	comboAPI := mocks.NewComboAPI(
		timerange,
		// Metric A
		api.Timeseries{
			Values: []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			TagSet: api.TagSet{"metric": "A", "foo": "rising"},
		},
		api.Timeseries{
			Values: []float64{7, 5, 6, 4, 8, 4, 5, 4, 6, 5},
			TagSet: api.TagSet{"metric": "A", "foo": "medium"},
		},
		api.Timeseries{
			Values: []float64{6, 8, 8, 8, 8, 8, 7, 8, 6, 7},
			TagSet: api.TagSet{"metric": "A", "foo": "high"},
		},
		api.Timeseries{
			Values: []float64{9, 7, 7, 5, 5, 4, 2, 1, 1, -1},
			TagSet: api.TagSet{"metric": "A", "foo": "falling"},
		},
		// Metric B
		api.Timeseries{
			Values: []float64{0, 1, 1, 0, 1, 4, 7, 8, 6, 9},
			TagSet: api.TagSet{"metric": "B", "foo": "low-high"},
		},
		api.Timeseries{
			Values: []float64{5, 8, 9, 7, 9, 0, 1, 1, 0, 1},
			TagSet: api.TagSet{"metric": "B", "foo": "high-low"},
		},
		api.Timeseries{
			Values: []float64{6, 8, 8, 8, 8, 8, 7, 8, 6, 7},
			TagSet: api.TagSet{"metric": "B", "foo": "high"},
		},
		api.Timeseries{
			Values: []float64{9, 9, 7, 7, 6, 6, 4, 4, 2, -1},
			TagSet: api.TagSet{"metric": "B", "foo": "falling"},
		},
	)
	type Test struct {
		Query    string
		Expected []string
	}
	tests := []Test{
		// Highest mean (A)
		{
			Query:    `select A | filter.highest_mean(4) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"high", "medium", "rising", "falling"},
		},
		{
			Query:    `select A | filter.highest_mean(3) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"high", "medium", "rising"},
		},
		{
			Query:    `select A | filter.highest_mean(2) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"high", "medium"},
		},
		{
			Query:    `select A | filter.highest_mean(1) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"high"},
		},
		// Lowest mean (A)
		{
			Query:    `select A | filter.lowest_mean(4) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"falling", "rising", "medium", "high"},
		},
		{
			Query:    `select A | filter.lowest_mean(3) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"falling", "rising", "medium"},
		},
		{
			Query:    `select A | filter.lowest_mean(2) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"falling", "rising"},
		},
		{
			Query:    `select A | filter.lowest_mean(1) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"falling"},
		},
		// Highest recent mean vs. highest mean (B)
		{
			Query:    `select B | filter.highest_mean(4) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"high", "falling", "high-low", "low-high"},
		},
		{ // 3000s is more than the request interval, so it will not change the answer.
			Query:    `select B | filter.highest_mean(4, 3000s) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"high", "falling", "high-low", "low-high"},
		},
		{ // 150s is only the second half.
			Query:    `select B | filter.highest_mean(4, 150s) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"high", "low-high", "falling", "high-low"},
		},
		// Now use "above" and "below" instead of "count"
		// Mean above (A)
		{
			Query:    `select A | filter.mean_above(0) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"high", "medium", "rising", "falling"},
		},
		{
			Query:    `select A | filter.mean_above(4.45) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"high", "medium", "rising"},
		},
		{
			Query:    `select A | filter.mean_above(4.55) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"high", "medium"},
		},
		{
			Query:    `select A | filter.mean_above(7) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"high"},
		},
		{
			Query:    `select A | filter.mean_above(12) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{},
		},
		// Mean below (A)
		{
			Query:    `select A | filter.mean_below(0) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{},
		},
		{
			Query:    `select A | filter.mean_below(4.45) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"falling"},
		},
		{
			Query:    `select A | filter.mean_below(4.55) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"falling", "rising"},
		},
		{
			Query:    `select A | filter.mean_below(7) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"falling", "rising", "medium"},
		},
		{
			Query:    `select A | filter.mean_below(12) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"falling", "rising", "medium", "high"},
		},
		// Mean above recent (B)
		{
			Query:    `select B | filter.mean_above(0, 150s) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"high", "low-high", "falling", "high-low"},
		},
		{
			Query:    `select B | filter.mean_above(1, 150s) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"high", "low-high", "falling", "high-low"},
		},
		{
			Query:    `select B | filter.mean_above(1, 120s) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"high", "low-high", "falling"},
		},
		{
			Query:    `select B | filter.mean_above(4.45, 150s) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"high", "low-high"},
		},
		{
			Query:    `select B | filter.mean_above(7, 150s) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{"high"},
		},
		{
			Query:    `select B | filter.mean_above(12, 150s) from 3000000 to 3270000 resolution 30s`,
			Expected: []string{},
		},
	}
	for _, test := range tests {
		testCommand, err := parser.Parse(test.Query)
		if err != nil {
			t.Errorf("Error parsing test query %q: %s", test.Query, err.Error())
			continue
		}
		rawResult, err := testCommand.Execute(command.ExecutionContext{
			TimeseriesStorageAPI: comboAPI,
			MetricMetadataAPI:    comboAPI,
			FetchLimit:           1000,
			Timeout:              100 * time.Millisecond,
		})
		if err != nil {
			t.Errorf("Error evaluating query %q: %s", test.Query, err.Error())
			continue
		}
		list := rawResult.Body.([]command.QueryResult)[0]
		tags := make([]string, len(list.Series))
		for i, series := range list.Series {
			tags[i] = series.TagSet["foo"]
		}
		a.Contextf("Query %q", test.Query).Eq(tags, test.Expected)
	}
}
