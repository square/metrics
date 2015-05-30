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

	"math"
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
		if len(result) != testCase.ExpectedGroups {
			t.Errorf("Testcase %d results in %d groups when %d are expected (tags %+v)", i, len(result), testCase.ExpectedGroups, testCase.Tags)
			continue
		}
		for _, row := range result {
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

var testGroup = group{
	List: []api.Timeseries{
		api.Timeseries{
			Values: []float64{0, 1, 2, 3},
			TagSet: api.TagSet{
				"env": "production",
				"dc":  "A",
			},
		},
		api.Timeseries{
			Values: []float64{4, 4, 4, 4},
			TagSet: api.TagSet{
				"env": "production",
				"dc":  "A",
			},
		},
		api.Timeseries{
			Values: []float64{-1, -1, 2, 2},
			TagSet: api.TagSet{
				"env": "production",
				"dc":  "A",
			},
		},
		api.Timeseries{
			Values: []float64{0, 2, 0, 2},
			TagSet: api.TagSet{
				"env": "production",
				"dc":  "A",
			},
		},
	},
	TagSet: api.TagSet{
		"env": "production",
		"dc":  "A",
	},
}

var aggregationTestCases = []struct {
	Aggregator aggregator
	Expected   []float64
}{
	{
		sumAggregator{},
		[]float64{3, 6, 8, 11},
	},
	{
		meanAggregator{},
		[]float64{3.0 / 4.0, 6.0 / 4.0, 8.0 / 4.0, 11.0 / 4.0},
	},
	{
		maxAggregator{},
		[]float64{4, 4, 4, 4},
	},
	{
		minAggregator{},
		[]float64{-1, -1, 0, 2},
	},
}

const epsilon = 1e-10 // epsilon is a constant for the maximum allowable error between correct test case answers and actual results

func Test_applyAggregation(t *testing.T) {
	for _, testCase := range aggregationTestCases {
		result, err := applyAggregation(testGroup, testCase.Aggregator)
		if err != nil {
			t.Error(err)
		}
		if result.TagSet["env"] != "production" {
			t.Fatalf("applyAggregation() produces tagset with env=%s but expected env=production", result.TagSet["env"])
		}
		if result.TagSet["dc"] != "A" {
			t.Fatalf("applyAggregation() produces tagset with dc=%s but expected dc=A", result.TagSet["dc"])
		}
		// Next, compare the aggregated values:
		for i, correct := range testCase.Expected {
			if math.Abs(result.Values[i]-correct) > epsilon {
				t.Fatalf("applyAggregation() produces incorrect values on aggregation %+v; should be %+v but is %+v", testCase.Aggregator, testCase.Expected, result.Values)
			}
		}
	}
}

var testList = api.SeriesList{
	[]api.Timeseries{
		api.Timeseries{
			Values: []float64{0, 1, 2},
			TagSet: api.TagSet{
				"env":  "staging",
				"dc":   "A",
				"host": "q77",
			},
		},
		api.Timeseries{
			Values: []float64{4, 4, 4},
			TagSet: api.TagSet{
				"env":  "staging",
				"dc":   "B",
				"host": "r53",
			},
		},
		api.Timeseries{
			Values: []float64{-1, -1, 2},
			TagSet: api.TagSet{
				"env":  "production",
				"dc":   "A",
				"host": "y1",
			},
		},
		api.Timeseries{
			Values: []float64{0, 2, 0},
			TagSet: api.TagSet{
				"env":  "production",
				"dc":   "A",
				"host": "w20",
			},
		},
		api.Timeseries{
			Values: []float64{2, 0, 0},
			TagSet: api.TagSet{
				"env":  "production",
				"dc":   "B",
				"host": "t8",
			},
		},
		api.Timeseries{
			Values: []float64{0, 0, 1},
			TagSet: api.TagSet{
				"env":  "production",
				"dc":   "C",
				"host": "b38",
			},
		},
	},
	api.Timerange{
		40,
		280,
		6,
	},
	"Test.List",
}

var aggregatedTests = []struct {
	Tags       []string
	Aggregator aggregator
	Results    []api.Timeseries
}{
	{
		[]string{"env"},
		sumAggregator{},
		[]api.Timeseries{
			api.Timeseries{
				Values: []float64{1, 1, 3},
				TagSet: map[string]string{
					"env": "production",
				},
			},
			api.Timeseries{
				Values: []float64{4, 5, 6},
				TagSet: map[string]string{
					"env": "staging",
				},
			},
		},
	},
	{
		[]string{"dc"},
		maxAggregator{},
		[]api.Timeseries{
			api.Timeseries{
				Values: []float64{0, 2, 2},
				TagSet: map[string]string{
					"dc": "A",
				},
			},
			api.Timeseries{
				Values: []float64{4, 4, 4},
				TagSet: map[string]string{
					"dc": "B",
				},
			},
			api.Timeseries{
				Values: []float64{0, 0, 1},
				TagSet: map[string]string{
					"dc": "C",
				},
			},
		},
	},
	{
		[]string{"dc", "env"},
		meanAggregator{},
		[]api.Timeseries{
			api.Timeseries{
				Values: []float64{0, 1, 2},
				TagSet: map[string]string{
					"dc":  "A",
					"env": "staging",
				},
			},
			api.Timeseries{
				Values: []float64{-1.0 / 2.0, 1.0 / 2.0, 1.0},
				TagSet: map[string]string{
					"dc":  "A",
					"env": "production",
				},
			},
			api.Timeseries{
				Values: []float64{4, 4, 4},
				TagSet: map[string]string{
					"dc":  "B",
					"env": "staging",
				},
			},
			api.Timeseries{
				Values: []float64{2, 0, 0},
				TagSet: map[string]string{
					"dc":  "B",
					"env": "production",
				},
			},
			api.Timeseries{
				Values: []float64{0, 0, 1},
				TagSet: map[string]string{
					"dc":  "C",
					"env": "production",
				},
			},
		},
	},
}

func tagSetsEqual(leftSet api.TagSet, rightSet api.TagSet) bool {
	for key, left := range leftSet {
		right, ok := rightSet[key]
		if !ok || left != right {
			return false
		}
	}
	return true
}

func Test_aggregateBy(t *testing.T) {

	for _, testCase := range aggregatedTests {
		aggregated, err := aggregateBy(testList, testCase.Aggregator, testCase.Tags)
		if err != nil {
			t.Error(err)
		}
		// Check that aggregated looks correct.
		// There should be two series
		if aggregated.Timerange != testList.Timerange {
			t.Errorf("Expected aggregate's Timerange to be %+v but is %+v", testList.Timerange, aggregated.Timerange)
			continue
		}
		if aggregated.Name != testList.Name {
			t.Errorf("Expected aggregate's Name to be %s but is %s", testList.Name, aggregated.Name)
			continue
		}
		if len(aggregated.Series) != len(testCase.Results) {
			t.Errorf("Expected %d series in aggregation result but found %d", len(testCase.Results), len(aggregated.Series))
			continue
		}
		// Lastly, we have to check that the values are correct.
		// First, check that an aggregated series corresponding to each correct tagset:
		for _, series := range testCase.Results {
			found := false
			for _, aggregate := range aggregated.Series {
				if tagSetsEqual(series.TagSet, aggregate.TagSet) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("Expected to find series corresponding to %+v but could not in %+v", series.TagSet, aggregated)
			}
		}
		// Next, each series will do the reverse-lookup and check that its values match the expected results.
		// (It is neccesary to check both ways [see above] to ensure that the result doesn't contain just one of the series repeatedly)
		for _, aggregate := range aggregated.Series {
			// Any of the testCase results which it matches are candidates
			for _, correct := range testCase.Results {
				if tagSetsEqual(aggregate.TagSet, correct.TagSet) {
					// Compare their values
					for i := range aggregate.Values {
						if math.Abs(aggregate.Values[i]-correct.Values[i]) > epsilon {
							t.Errorf("For tagset %+v, result %+v does not match expected %+v", correct.TagSet, aggregate.Values, correct.Values)
							break
						}
					}
				}
			}
		}
	}
}
