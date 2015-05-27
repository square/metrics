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

type JoinRow struct {
	// The tagSet is used to improve performance, or, possibly in the future, for later queries
	TagSet api.TagSet
	// The Row consists of all Timeseries which got collected into this JoinRow
	Row []api.Timeseries
}

type JoinResult struct {
	Rows []JoinRow
}

// This method takes a partial joinrow, and evaluates the validity of appending `series` to it.
// If this is possible, return the new series and true; otherwise return false for "ok"
func extendRow(row JoinRow, series api.Timeseries) (JoinRow, bool) {
	for key, newValue := range series.Metric.TagSet {
		oldValue, ok := map[string]string(row.TagSet)[key]
		if ok && newValue != oldValue {
			// If this occurs, then the candidate member (series) and the rest of the row are in
			// conflict about `key`, since they assign it different values. If this occurs, then
			// it is not possible to assign any key here.
			return JoinRow{}, false
		}
	}
	// if this point has been reached, then it is possible to extend the row without conflict
	newTagSet := make(map[string]string)
	result := JoinRow{newTagSet, append(row.Row, series)}
	for key, newValue := range series.Metric.TagSet {
		newTagSet[key] = newValue
	}
	for key, oldValue := range row.TagSet {
		newTagSet[key] = oldValue
	}
	return result, true
}

func Join(lists []api.SeriesList) JoinResult {
	// place an empty row inside the results list first
	// this row will be used to build up all others
	emptyRow := JoinRow{make(map[string]string), []api.Timeseries{}}
	results := []JoinRow{emptyRow}

	// The `results` list is given an inductive definition:
	// at the end of the `i`th iteration of the outer loop,
	// `results` corresponds to the Join of the first `i` seriesLists given as input

	for _, list := range lists {
		next := []JoinRow{}
		// `next` is gradually accumulated into the final Join of the first `i` seriesLists
		// so that at the end of the loop, we can assign:
		//     results = next
		// satisfying the specification of results given above (that it will be the join of the first `i` seriesLists)
		for _, series := range list.Series {
			// here we have our series
			// iterator over the results of the previous iteration:
			for _, previous := range results {
				extension, ok := extendRow(previous, series)
				if ok {
					next = append(next, extension)
				}
			}
		}
		results = next
	}

	return JoinResult{Rows: results}
}
