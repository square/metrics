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

// Package api holds common data types and public interface exposed by the indexer library.
// Refer to the doc
// https://docs.google.com/a/squareup.com/document/d/1k0Wgi2wnJPQoyDyReb9dyIqRrD8-v0u8hz37S282ii4/edit
// for the terminology.

package api

import (
	"math"
	"testing"

	"github.com/square/metrics/testing_support/assert"
)

func max(array []float64) float64 {
	max := array[0]
	for _, v := range array {
		max = math.Max(max, v)
	}
	return max
}

func TestTimeseries_Downsample(t *testing.T) {
	a := assert.New(t)
	for _, suite := range []struct {
		input      []float64
		inputRange Timerange
		newRange   Timerange
		sampler    func([]float64) float64
		expected   []float64
	}{
		{[]float64{1, 2, 3, 4, 5}, Timerange{0, 4, 1}, Timerange{0, 4, 2}, max, []float64{2, 4, 5}},
	} {
		tagset := ParseTagSet("key=value")
		ts := Timeseries{Values: suite.input, TagSet: tagset}
		sampled, err := ts.Downsample(suite.inputRange, suite.newRange, suite.sampler)
		a.CheckError(err)
		a.Eq(sampled.Values, suite.expected)
	}
}
