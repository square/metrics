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
	return function.MakeFunction(
		name,
		func(list api.SeriesList, optionalDuration *time.Duration, timerange api.Timerange) function.ScalarSet {
			duration := timerange.Duration()
			if optionalDuration != nil {
				duration = *optionalDuration
			}
			start := timerange.Slots() - 1 - int(duration/timerange.Resolution())
			if start < 0 {
				start = 0
				// TODO: warn or error?
			}
			result := function.ScalarSet{}
			for i := range list.Series {
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
			if math.IsNaN(min) {
				min = slice[i]
			}
			if math.IsNaN(slice[i]) {
				continue
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
			if math.IsNaN(max) {
				max = slice[i]
			}
			if math.IsNaN(slice[i]) {
				continue
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
			}
		}
		return math.NaN()
	},
)

var Current = function.MakeFunction(
	"summarize.current",
	func(list api.SeriesList) function.ScalarSet {
		result := function.ScalarSet{}
		for i := range list.Series {
			values := list.Series[i].Values
			result = append(result, function.TaggedScalar{
				TagSet: list.Series[i].TagSet,
				Value:  values[len(values)-1],
			})
		}
		return result
	},
)
