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
	"sort"

	"github.com/square/metrics/api"
)

type filterList struct {
	index     []int
	value     []float64
	ascending bool
}

func (list filterList) Len() int {
	return len(list.index)
}
func (list filterList) Less(i, j int) bool {
	return (list.value[i] < list.value[j]) == list.ascending
}
func (list filterList) Swap(i, j int) {
	list.index[i], list.index[j] = list.index[j], list.index[i]
	list.value[i], list.value[j] = list.value[j], list.value[i]
}

func newFilterList(size int, bool ascending) {
	return filterList{
		index:     make([]int, size),
		value:     make([]float64, size),
		ascending: ascending,
	}
}

// FilteryBy reduces the number of things in the series `list` to at most the given `count`.
// They're chosen by sorting by `summary` in `ascending` or descending order.
func FilterBy(list api.SeriesList, count int, summary func([]float64) float64, ascending bool) {
	if len(list.Series) < count {
		// No need to change if there's already fewer.
		return list
	}
	array := newFilterList(len(list.Series), ascending)
	for i := range array {
		array[i].index = i
		array[i].value = summary(list.Series[i].Values)
	}
	sort.Sort(array)

	series := make([]api.Timeseries, count)
	for i := range series {
		series[i] = list.Series[array.index[i]]
	}

	return api.SeriesList{
		Series:    series,
		Timerange: list.Timerange,
	}
}
