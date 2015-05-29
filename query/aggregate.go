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

type groupResult struct {
	Results []group
}

// If the given group will accept this given series (since it belongs to this group)
// then bucketValid will return true.
func bucketValid(row group, series api.Timeseries, tags []string) bool {
	for _, tag := range tags {
		if row.TagSet[tag] != series.TagSet[tag] {
			return false
		}
	}
	return true
}

// Adds the series to the corresponding bucket, possibly modifying the input `rows` and returning a new list.
func addToBucket(rows []group, series api.Timeseries, tags []string) []group {
	// First we delete all tags with names other than those found in 'tags'
	newTags := api.NewTagSet()
	for _, tag := range tags {
		newTags[tag] = series.TagSet[tag]
	}
	// replace series' TagSet with newTags
	series.TagSet = newTags

	// Next, find the best bucket for this series:
	for i, row := range rows {
		if bucketValid(row, series, tags) {
			rows[i].List = append(rows[i].List, series)
			return rows
		}
	}
	tagSet := api.NewTagSet()
	for _, tag := range tags {
		tagSet[tag] = series.TagSet[tag]
	}
	return append(rows, group{
		[]api.Timeseries{series},
		tagSet,
	})
}

// Groups the given SeriesList by tags, producing a list of lists (of type groupResult)
func groupBy(list api.SeriesList, tags []string) []group {
	result := []group{}
	for _, series := range list.Series {
		result = addToBucket(result, series, tags)
	}
	return result
}

// The aggregator interface is the public-facing way in which values are aggregated.
// Aggregator objects are required to perform aggregation (max, min, range, mean, sum, etc.)
// Their only interface method is the `beginAggregation()` method which returns an "aggregation".
type aggregator interface {
	beginAggregation() aggregation
}

// An aggregation is a private interface which aggregates values for a particular SeriesList.
// Because this interface is private, there would be no way to create aggregator outside this package.
// If extension is desirable, this may change.
// They can `accumulate(...)` value, and they can compute their `result()`.
// During aggregation, `accumulate(...)` is called with each value corresponding to a given index,
// with one call for each Timeseries inside the SeriesList.
// Then, to compute the resulting value, `result()` is invoked exactly once.
type aggregation interface {
	accumulate(float64)
	result() float64
}

// Here are example aggregators and their aggregation types.
// Aggregators will generally be empty structs (or an equivalent),
// although they could alternatively store parameters which they can use
// to which which aggregation to return, or to supply parameters to their aggregations.
type sumAggregator struct {
}

// The `beginAggregation()` for sum returns a pointer to a new sumAggregation,
// which has a sum set to 0.
func (aggregator sumAggregator) beginAggregation() aggregation {
	return &sumAggregation{
		sum: 0,
	}
}

// The sumAggregation struct just contains the sum, which it will accumulate over time.
type sumAggregation struct {
	sum float64
}

// The accumulation for the sumAggregation consists of adding the value to the struct's sum.
// Note that the interface is on the pointer `*sumAggregation` rather than `sumAggregation`
// because it must be mutable.
func (aggregation *sumAggregation) accumulate(value float64) {
	aggregation.sum += value
}

// The result just returns this value.
func (aggregation *sumAggregation) result() float64 {
	return aggregation.sum
}

// A mean aggregator is highly similar to a sum aggregator.
// It computes the aggregate mean for a seriesList
type meanAggregator struct {
}

// The mean aggregator returns a meanAggregation pointer with a `sum` and `count` both 0.
func (aggregator meanAggregator) beginAggregation() aggregation {
	return &meanAggregation{
		sum:   0,
		count: 0,
	}
}

// The `sum` and `count` fields totally define the meanAggregation's state.
type meanAggregation struct {
	sum   float64
	count int
}

// The accumulaton function adds the value to the mean's running `sum`, and increments its `count`.
func (aggregation *meanAggregation) accumulate(value float64) {
	aggregation.sum += value
	aggregation.count++
}

// The result returns the quotient of the running `sum` and `count`, computed through `accumulate()`
func (aggregation *meanAggregation) result() float64 {
	return aggregation.sum / float64(aggregation.count)
}

// The min aggregator is an aggregator that computes the aggregate minimum for a seriesList
type minAggregator struct {
}

func (aggregator minAggregator) beginAggregation() aggregation {
	return &minAggregation{
		min: math.Inf(1),
	}
}

// The min aggregation is the aggregation for the min aggregator
type minAggregation struct {
	min float64
}

func (aggregation *minAggregation) accumulate(value float64) {
	aggregation.min = math.Min(aggregation.min, value)
}
func (aggregation *minAggregation) result() float64 {
	return aggregation.min
}

// The maxAggregator is an aggregator that computes the aggregate maximum for a seriesList
type maxAggregator struct {
}

func (aggregator maxAggregator) beginAggregation() aggregation {
	return &maxAggregation{
		max: math.Inf(-1),
	}
}

// The maxAggregation is an aggregation that computes the aggregate minimum for a seriesList
type maxAggregation struct {
	max float64
}

func (aggregation *maxAggregation) accumulate(value float64) {
	aggregation.max = math.Max(aggregation.max, value)
}
func (aggregation *maxAggregation) result() float64 {
	return aggregation.max
}

func useAggregator(aggregator aggregator, values []float64) float64 {
	aggregation := aggregator.beginAggregation()
	for _, v := range values {
		aggregation.accumulate(v)
	}
	return aggregation.result()
}

// applyAggregation takes an aggregation function ( [float64] => float64 ) and applies it to a given list of Timeseries
// the list must be non-empty, or an error is returned
func applyAggregation(group group, aggregator aggregator) (api.Timeseries, error) {
	list := group.List
	tagSet := group.TagSet

	if len(list) == 0 {
		return api.Timeseries{}, EmptyAggregateError{}
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
		series.Values[i] = useAggregator(aggregator, timeSlice)
	}

	return series, nil
}

// This function is the culmination of all others.
// `aggregateBy` takes a series list, an aggregator, and a set of tags.
// It produces a SeriesList which is the result of grouping by the tags and then aggregating each group
// into a single Series.
func aggregateBy(list api.SeriesList, aggregator aggregator, tags []string) (api.SeriesList, error) {
	// Begin by grouping the input:
	groups := groupBy(list, tags)

	result := api.SeriesList{
		Series:    make([]api.Timeseries, len(groups)),
		Timerange: list.Timerange,
		Name:      list.Name,
	}

	for i, group := range groups {
		// The group contains a list of Series and a TagSet.
		aggregated, err := applyAggregation(group, aggregator)
		if err != nil {
			return api.SeriesList{}, err
		}
		result.Series[i] = aggregated
	}
	return result, nil
}
