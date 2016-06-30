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

package join

import (
	"testing"

	"github.com/square/metrics/api"
)

var (
	seriesDCOfAHost1 = api.Timeseries{Values: []float64{1, 2, 3}, TagSet: map[string]string{"dc": "A", "host": "#1"}}
	seriesDCOfAHost2 = api.Timeseries{Values: []float64{4, 5, 6}, TagSet: map[string]string{"dc": "A", "host": "#2"}}
	seriesDCOfBHost3 = api.Timeseries{Values: []float64{0, 1, 1}, TagSet: map[string]string{"dc": "B", "host": "#3"}}
	seriesDCOfBHost4 = api.Timeseries{Values: []float64{1, 3, 2}, TagSet: map[string]string{"dc": "B", "host": "#4"}}
	seriesDCOfCHost5 = api.Timeseries{Values: []float64{2, 2, 3}, TagSet: map[string]string{"dc": "C", "host": "#5"}}

	seriesDCOfA = api.Timeseries{Values: []float64{2, 0, 1}, TagSet: map[string]string{"dc": "A"}}
	seriesDCOfB = api.Timeseries{Values: []float64{2, 0, 1}, TagSet: map[string]string{"dc": "B"}}
	seriesDCOfC = api.Timeseries{Values: []float64{2, 0, 1}, TagSet: map[string]string{"dc": "C"}}

	seriesEnvOfProd  = api.Timeseries{Values: []float64{2, 0, 1}, TagSet: map[string]string{"env": "production"}}
	seriesEnvOfStage = api.Timeseries{Values: []float64{2, 0, 1}, TagSet: map[string]string{"env": "staging"}}

	voidSeries = api.Timeseries{Values: []float64{0, 0, 0}, TagSet: map[string]string{}}

	emptyList = api.SeriesList{Series: []api.Timeseries{}}
	basicList = api.SeriesList{Series: []api.Timeseries{seriesDCOfAHost1, seriesDCOfAHost2, seriesDCOfBHost3, seriesDCOfBHost4, seriesDCOfCHost5}}
	dcList    = api.SeriesList{Series: []api.Timeseries{seriesDCOfA, seriesDCOfB, seriesDCOfC}}
	envList   = api.SeriesList{Series: []api.Timeseries{seriesEnvOfProd, seriesEnvOfStage}}

	voidList = api.SeriesList{Series: []api.Timeseries{voidSeries}}
)

var testCases = []struct {
	joinArgument   []api.SeriesList
	expectedLength int
}{
	// Cases with empty results:
	{joinArgument: []api.SeriesList{emptyList}, expectedLength: 0},
	{joinArgument: []api.SeriesList{emptyList, emptyList}, expectedLength: 0},
	{joinArgument: []api.SeriesList{emptyList, basicList}, expectedLength: 0},
	{joinArgument: []api.SeriesList{basicList, emptyList}, expectedLength: 0},
	{joinArgument: []api.SeriesList{basicList, basicList, basicList, emptyList, basicList}, expectedLength: 0},
	// Cases where the resulting length is the same as the input(s)
	{joinArgument: []api.SeriesList{basicList}, expectedLength: len(basicList.Series)},
	{joinArgument: []api.SeriesList{basicList, basicList}, expectedLength: len(basicList.Series)},
	{joinArgument: []api.SeriesList{dcList}, expectedLength: len(dcList.Series)},
	{joinArgument: []api.SeriesList{dcList, dcList}, expectedLength: len(dcList.Series)},
	{joinArgument: []api.SeriesList{envList}, expectedLength: len(envList.Series)},
	{joinArgument: []api.SeriesList{envList, envList}, expectedLength: len(envList.Series)},
	// Cases where the resulting length is the maximum of the inputs'
	{joinArgument: []api.SeriesList{basicList, dcList}, expectedLength: max(len(basicList.Series), len(dcList.Series))},
	{joinArgument: []api.SeriesList{dcList, basicList}, expectedLength: max(len(basicList.Series), len(dcList.Series))},
	{joinArgument: []api.SeriesList{basicList, voidList}, expectedLength: len(basicList.Series)},
	{joinArgument: []api.SeriesList{voidList, basicList}, expectedLength: len(basicList.Series)},
	{joinArgument: []api.SeriesList{basicList, dcList}, expectedLength: len(basicList.Series)},
	// Cases where the resulting length is the product of the inputs'
	{joinArgument: []api.SeriesList{basicList, envList}, expectedLength: len(basicList.Series) * len(envList.Series)},
	{joinArgument: []api.SeriesList{envList, dcList}, expectedLength: len(envList.Series) * len(dcList.Series)},
}

func Test_join_ResultSizes(t *testing.T) {
	for i, testCase := range testCases {
		result := Join(testCase.joinArgument)
		if len(result.Rows) != testCase.expectedLength {
			t.Errorf("join testcase %d results in %d; expected %d", i, len(result.Rows), testCase.expectedLength)
			t.Errorf("testcase: %+v", testCase.joinArgument)
		}
	}
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}
