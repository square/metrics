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

	"testing"
)

var (
	listA = api.SeriesList{
		Series: []api.Timeseries{
			api.Timeseries{
				Values: []float64{0, 0, 0},
				TagSet: map[string]string{
					"dc":   "A",
					"env":  "production",
					"host": "#1",
				},
			},
			api.Timeseries{
				Values: []float64{1, 1, 1},
				TagSet: map[string]string{
					"dc":   "B",
					"env":  "staging",
					"host": "#1",
				},
			},
			api.Timeseries{
				Values: []float64{2, 2, 2},
				TagSet: map[string]string{
					"dc":   "C",
					"env":  "staging",
					"host": "#1",
				},
			},
			api.Timeseries{
				Values: []float64{3, 3, 3},
				TagSet: map[string]string{
					"dc":   "B",
					"env":  "production",
					"host": "#2",
				},
			},
			api.Timeseries{
				Values: []float64{4, 4, 4},
				TagSet: map[string]string{
					"dc":   "C",
					"env":  "staging",
					"host": "#2",
				},
			},
		},
		Timerange: api.Timerange{},
		Name:      "",
	}
)

var aggregateTestCases = []struct {
	Tags           []string
	ExpectedGroups int
}{
	{
		[]string{"dc"},
		3,
	},
	{
		[]string{"host"},
		2,
	},
	{
		[]string{"env"},
		2,
	},
	{
		[]string{"dc", "host"},
		5,
	},
	{
		[]string{"dc", "env"},
		4,
	},
	{
		[]string{"dc", "env"},
		4,
	},
	{
		[]string{},
		1,
	},
}

// Checks that groupBy() behaves as expected
func Test_groupBy(t *testing.T) {
	for i, testCase := range aggregateTestCases {
		result := groupBy(listA, testCase.Tags)
		if len(result.Results) != testCase.ExpectedGroups {
			t.Errorf("Testcase %d results in %d groups when %d are expected (tags %+v)", i, len(result.Results), testCase.ExpectedGroups, testCase.Tags)
			continue
		}
		for _, row := range result.Results {
			// Further consistency checks are needed
			for _, series := range row.List {
				for _, tag := range testCase.Tags {
					if series.TagSet[tag] != row.TagSet[tag] {
						t.Errorf("Series %+v in row %+v has inconsistent tag %s", series, row, tag)
						continue
					}
					if len(series.Values) != 3 {
						t.Errorf("groupBy changed the number of elements in Values: %+v", series)
						continue
					}
					originalIndex := int(series.Values[0])
					if originalIndex < 0 || originalIndex >= len(listA.Series) {
						t.Errorf("groupBy has changed the values in Values: %+v", series)
						continue
					}
					original := listA.Series[originalIndex]
					if original.TagSet[tag] != series.TagSet[tag] {
						t.Errorf("groupBy changed a series' tagset[%s]: original %+v; result %+v", tag, original, series)
						continue
					}
				}
			}
		}
	}

}
