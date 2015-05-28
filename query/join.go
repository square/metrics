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

import (
	"github.com/square/metrics/api"
)

type joinRow struct {
	TagSet api.TagSet       // The tagSet is used to improve performance, or, possibly in the future, for later queries
	Row    []api.Timeseries // The Row consists of all Timeseries which got collected into this joinRow
}

type joinResult struct {
	Rows []joinRow
}

// This method takes a partial joinrow, and evaluates the validity of appending `series` to it.
// If this is possible, return the new series and true; otherwise return false for "ok"
func extendRow(row joinRow, series api.Timeseries) (joinRow, bool) {
	for key, newValue := range series.Metric.TagSet {
		oldValue, ok := map[string]string(row.TagSet)[key]
		if ok && newValue != oldValue {
			// If this occurs, then the candidate member (series) and the rest of the row are in
			// conflict about `key`, since they assign it different values. If this occurs, then
			// it is not possible to assign any key here.
			return joinRow{}, false
		}
	}
	// if this point has been reached, then it is possible to extend the row without conflict
	newTagSet := api.NewTagSet()
	result := joinRow{newTagSet, append(row.Row, series)}
	for key, newValue := range series.Metric.TagSet {
		newTagSet[key] = newValue
	}
	for key, oldValue := range row.TagSet {
		newTagSet[key] = oldValue
	}
	return result, true
}

// join generates a cartesian product of the given series lists, and then returns rows where the tags are matching.
func join(lists []api.SeriesList) joinResult {
	// place an empty row inside the results list first
	// this row will be used to build up all others
	emptyRow := joinRow{api.NewTagSet(), []api.Timeseries{}}
	results := []joinRow{emptyRow}

	// The `results` list is given an inductive definition:
	// at the end of the `i`th iteration of the outer loop,
	// `results` corresponds to the join of the first `i` seriesLists given as input

	for _, list := range lists {
		next := []joinRow{}
		// `next` is gradually accumulated into the final join of the first `i` (iteration) seriesLists
		// results already contains the join of the the first (i-1)th series
		for _, series := range list.Series {
			// here we have our series
			// iterator over the results of the previous iteration:
			for _, previous := range results {
				// consider adding this series to each row from the joins of all previous series lists
				// if this is successful, the newly extended list is added to the `next` slice
				extension, ok := extendRow(previous, series)
				if ok {
					next = append(next, extension)
				}
			}
		}
		// `next` now contains the join of the first `i` iterations,
		// while `results` contains the join of the first `i-1` iterations.
		results = next
		// thus we update `results`
	}
	// at this stage, iteration has continued over the entire set of lists,
	// so `results` contains the join of all of the lists.

	return joinResult{Rows: results}
}
