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

package query

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

// If the given group will accept this given series (since it belongs to this group)
// then groupAccept will return true.
func groupAccepts(left api.TagSet, right api.TagSet, tags []string) bool {
	for _, tag := range tags {
		if left[tag] != right[tag] {
			return false
		}
	}
	return true
}

// Adds the series to the corresponding bucket, possibly modifying the input `rows` and returning a new list.
func addToGroup(rows []group, series api.Timeseries, tags []string) []group {
	// First we delete all tags with names other than those found in 'tags'
	newTags := api.NewTagSet()
	for _, tag := range tags {
		newTags[tag] = series.TagSet[tag]
	}
	// replace series' TagSet with newTags
	series.TagSet = newTags

	// Next, find the best bucket for this series:
	for i, row := range rows {
		if groupAccepts(row.TagSet, series.TagSet, tags) {
			rows[i].List = append(rows[i].List, series)
			return rows
		}
	}
	// Otherwise, no bucket yet exists
	return append(rows, group{
		[]api.Timeseries{series},
		newTags,
	})
}

// Groups the given SeriesList by tags, producing a list of lists (of type groupResult)
func groupBy(list api.SeriesList, tags []string) []group {
	result := []group{}
	for _, series := range list.Series {
		result = addToGroup(result, series, tags)
	}
	return result
}

// The aggregator interface is the public-facing way in which values are aggregated.
// Aggregator objects are required to perform aggregation (max, min, range, mean, sum, etc.)
// Their interface is the `aggregate` method, which takes an array of floats.
type aggregator interface {
	aggregate([]float64) float64
}

type sumAggregator struct {
}

var _ aggregator = sumAggregator{}

// The `aggregate()` for sum finds the sum of the given array.
func (aggregator sumAggregator) aggregate(array []float64) float64 {
	sum := 0.0
	for _, v := range array {
		sum += v
	}
	return sum
}

// A mean aggregator is highly similar to a sum aggregator.
// It computes the aggregate mean for a seriesList
type meanAggregator struct {
}

var _ aggregator = meanAggregator{}

// The mean aggregator returns the mean of the given array
func (aggregator meanAggregator) aggregate(array []float64) float64 {
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

// The min aggregator is an aggregator that computes the aggregate minimum for a seriesList
type minAggregator struct {
}

var _ aggregator = minAggregator{}

func (aggregator minAggregator) aggregate(array []float64) float64 {
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

// The maxAggregator is an aggregator that computes the aggregate maximum for a seriesList
type maxAggregator struct {
}

var _ aggregator = maxAggregator{}

func (aggregator maxAggregator) aggregate(array []float64) float64 {
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

// applyAggregation takes an aggregation function ( [float64] => float64 ) and applies it to a given list of Timeseries
// the list must be non-empty, or an error is returned
func applyAggregation(group group, aggregator aggregator) api.Timeseries {
	list := group.List
	tagSet := group.TagSet

	if len(list) == 0 {
		// This case cannot actually occur, provided the rest of the code has been implemented correctly.
		// However, if it *does* occur, then we need to exit early to prevent a panic when we access list[0].
		return api.Timeseries{
			Values: []float64{},
			TagSet: tagSet,
		}
	}

	series := api.Timeseries{
		Values: make([]float64, len(list[0].Values)), // The first Series in the given list is used to determine this length
		TagSet: tagSet,                               // The tagset is supplied by an argument (it will be the values grouped on)
	}

	// Make a slice of time to reuse.
	// Each entry corresponds to a particular Series, all having the same index within their corresponding Series.
	timeSlice := make([]float64, len(list))

	for i := range series.Values {
		// We need to determine each value in turn.
		for j := range timeSlice {
			timeSlice[j] = list[j].Values[i]
		}
		// Find the aggregated value:
		series.Values[i] = aggregator.aggregate(timeSlice)
	}

	return series
}

// This function is the culmination of all others.
// `aggregateBy` takes a series list, an aggregator, and a set of tags.
// It produces a SeriesList which is the result of grouping by the tags and then aggregating each group
// into a single Series.
func aggregateBy(list api.SeriesList, aggregator aggregator, tags []string) api.SeriesList {
	// Begin by grouping the input:
	groups := groupBy(list, tags)

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
