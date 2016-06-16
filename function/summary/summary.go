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

package summary

import (
	"math"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
)

// This file culminates in the definition of `aggregateBy`, which takes a SeriesList and an Aggregator and a list of tags,
// and produces an aggregated SeriesList with one list per group, each group having been aggregated into it.

var recent = func(name string, summarizer func([]float64) float64) function.MetricFunction {
	// @@ leaking param: summarizer
	// @@ leaking param: name to result ~r2 level=0
	return function.MakeFunction(
		name,
		func(list api.SeriesList, optionalDuration *time.Duration, timerange api.Timerange) function.ScalarSet {
			// @@ leaking param content: list
			// @@ leaking param content: list
			duration := timerange.Duration()
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			if optionalDuration != nil {
				duration = *optionalDuration
			}
			start := timerange.Slots() - 1 - int(duration/timerange.Resolution())
			if start < 0 {
				// @@ inlining call to api.Timerange.Slots
				// @@ inlining call to api.Timerange.Resolution
				start = 0
				// TODO: warn or error?
			}
			result := function.ScalarSet{}
			for i := range list.Series {
				// @@ function.ScalarSet literal escapes to heap
				slice := list.Series[i].Values[start:]
				result = append(result, function.TaggedScalar{
					TagSet: list.Series[i].TagSet,
					Value:  summarizer(slice),
				})
			}
			return result
		},
	)
}

var Mean = recent(
	"summarize.mean",
	func(slice []float64) float64 {
		sum := 0.0
		count := 0
		for i := range slice {
			if math.IsNaN(slice[i]) {
				continue
				// @@ inlining call to math.IsNaN
			}
			sum += slice[i]
			count++
		}
		return sum / float64(count)
	},
)

var Min = recent(
	"summarize.min",
	func(slice []float64) float64 {
		min := math.NaN()
		for i := range slice {
			// @@ inlining call to math.NaN
			// @@ inlining call to math.Float64frombits
			if math.IsNaN(min) {
				min = slice[i]
				// @@ inlining call to math.IsNaN
			}
			if math.IsNaN(slice[i]) {
				continue
				// @@ inlining call to math.IsNaN
			}
			min = math.Min(min, slice[i])
		}
		return min
	},
)

var Max = recent(
	"summarize.max",
	func(slice []float64) float64 {
		max := math.NaN()
		for i := range slice {
			// @@ inlining call to math.NaN
			// @@ inlining call to math.Float64frombits
			if math.IsNaN(max) {
				max = slice[i]
				// @@ inlining call to math.IsNaN
			}
			if math.IsNaN(slice[i]) {
				continue
				// @@ inlining call to math.IsNaN
			}
			max = math.Max(max, slice[i])
		}
		return max
	},
)

var LastNotNaN = recent(
	"summarize.last_not_nan",
	func(slice []float64) float64 {
		for i := range slice {
			if !math.IsNaN(slice[len(slice)-1-i]) {
				return slice[len(slice)-1-i]
				// @@ inlining call to math.IsNaN
			}
		}
		return math.NaN()
	},
	// @@ inlining call to math.NaN
	// @@ inlining call to math.Float64frombits
)

var Current = function.MakeFunction(
	"summarize.current",
	func(list api.SeriesList) function.ScalarSet {
		// @@ leaking param content: list
		result := function.ScalarSet{}
		for i := range list.Series {
			// @@ function.ScalarSet literal escapes to heap
			values := list.Series[i].Values
			result = append(result, function.TaggedScalar{
				TagSet: list.Series[i].TagSet,
				Value:  values[len(values)-1],
			})
		}
		return result
	},
)
