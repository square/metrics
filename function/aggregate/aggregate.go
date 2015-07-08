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

package aggregate

// This file culminates in the definition of `aggregateBy`, which takes a SeriesList and an Aggregator and a list of tags,
// and produces an aggregated SeriesList with one list per group, each group having been aggregated into it.

import (
	"math"

	"github.com/square/metrics/api"
)

type group struct {
	List   []api.Timeseries
	TagSet api.TagSet
}

// groupAccepts determines whether the given `group` tagset will accept the `next` candidate tagset.
// in particular, they must have the same values for any keys that they share.
func groupAccepts(group api.TagSet, next api.TagSet) bool {
	for tag, value := range group {
		if nextValue, ok := next[tag]; ok {
			if nextValue != value {
				return false
			}
		}
	}
	return true
}

// addToGroup adds the `series` to the corresponding bucket, possibly modifying the input `rows` and returning a new list.
func addToGroup(rows []group, series api.Timeseries) []group {
	// Find the best bucket for this series:
	for i, row := range rows {
		if groupAccepts(row.TagSet, series.TagSet) {
			rows[i].List = append(rows[i].List, series)
			return rows
		}
	}
	// Otherwise, no bucket yet exists
	return append(rows, group{
		[]api.Timeseries{series},
		series.TagSet,
	})
}

// startingGroup returns a tagset that only has tags from `original` that are found in `tags`.
func startingGroup(original api.TagSet, tags []string) api.TagSet {
	result := api.NewTagSet()
	for _, tag := range tags {
		result[tag] = original[tag]
	}
	return result
}

// startingCollapse returns a tagset copy of `original` but with all tags in `tags` deleted.
func startingCollase(original api.TagSet, tags []string) api.TagSet {
	result := api.NewTagSet()
	for tag, value := range original {
		result[tag] = value
	}
	for _, tagToDelete := range tags {
		delete(result, tagToDelete)
	}
	return result
}

// filterTagSet takes a `series` and filters its `tags` based on whether it needs to `collapse` or group on these tags.
// if `collapses` is false, then tags not found in the `tags` list will be deleted. If `collapses` is true, then tags found in `tags` will be deleted.
func filterTagSet(series api.Timeseries, tags []string, collapses bool) api.Timeseries {
	if collapses {
		series.TagSet = startingCollase(series.TagSet, tags)
	} else {
		series.TagSet = startingGroup(series.TagSet, tags)
	}
	return series
}

// groupBy breaks the given `list` into `groups` that all agree on each tag they have in `tags`.
// if `collapses` is true, then it groups on all other tags instead.
func groupBy(list api.SeriesList, tags []string, collapses bool) []group {
	result := []group{}
	for _, series := range list.Series {
		result = addToGroup(result, filterTagSet(series, tags, collapses))
	}
	return result
}

// filterNaN removes NaN elements from the given slice (producing a copy)
func filterNaN(array []float64) []float64 {
	result := []float64{}
	for _, v := range array {
		if !math.IsNaN(v) {
			result = append(result, v)
		}
	}
	return result
}

// Sum returns the mean of the given slice
func Sum(array []float64) float64 {
	array = filterNaN(array)
	sum := 0.0
	for _, v := range array {
		sum += v
	}
	return sum
}

// Mean aggregator returns the mean of the given slice
func Mean(array []float64) float64 {
	array = filterNaN(array)
	if len(array) == 0 {
		// The mean of an empty list is not well-defined
		return math.NaN()
	}
	sum := 0.0
	for _, v := range array {
		sum += v
	}
	return sum / float64(len(array))
}

// Min returns the minimum of the given slice
func Min(array []float64) float64 {
	array = filterNaN(array)
	if len(array) == 0 {
		// The minimum of an empty list is not well-defined
		return math.NaN()
	}
	min := array[0]
	for _, v := range array {
		min = math.Min(min, v)
	}
	return min
}

// Max returns the maximum of the given slice
func Max(array []float64) float64 {
	array = filterNaN(array)
	if len(array) == 0 {
		// The maximum of an empty list is not well-defined
		return math.NaN()
	}
	max := array[0]
	for _, v := range array {
		max = math.Max(max, v)
	}
	return max
}

// Total returns the number of values in the given list.
func Total(array []float64) float64 {
	return float64(len(array))
}

// Count returns the number of non-NaN values in the givne list.
func Count(array []float64) float64 {
	return float64(len(filterNaN(array)))
}

// applyAggregation takes an aggregation function ( [float64] => float64 ) and applies it to a given list of Timeseries
// the list must be non-empty, or an error is returned
func applyAggregation(group group, aggregator func([]float64) float64) api.Timeseries {
	list := group.List
	tagSet := group.TagSet

	if len(list) == 0 {
		// This case should not actually occur, provided the rest of the code has been implemented correctly.
		// So when it does, issue a panic:
		panic("applyAggregation given empty group for tagset")
	}

	result := api.Timeseries{
		Values: make([]float64, len(list[0].Values)), // The first Series in the given list is used to determine this length
		TagSet: tagSet,                               // The tagset is supplied by an argument (it will be the values grouped on)
	}

	for i := range result.Values {
		timeSlice := make([]float64, len(list))
		for j := range list {
			timeSlice[j] = list[j].Values[i]
		}
		result.Values[i] = aggregator(timeSlice)
	}

	return result
}

// AggregateBy takes a series list, an aggregator, and a set of tags.
// It produces a SeriesList which is the result of grouping by the tags and then aggregating each group
// into a single Series.
func AggregateBy(list api.SeriesList, aggregator func([]float64) float64, tags []string, collapses bool) api.SeriesList {
	// Begin by grouping the input:
	groups := groupBy(list, tags, collapses)

	result := api.SeriesList{
		Series:    make([]api.Timeseries, len(groups)),
		Timerange: list.Timerange,
		Name:      list.Name,
	}

	for i, group := range groups {
		// The group contains a list of Series and a TagSet.
		result.Series[i] = applyAggregation(group, aggregator)
	}
	return result
}
