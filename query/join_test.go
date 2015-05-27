package query

import (
	"testing"

	"github.com/square/metrics/api"
)

var (
	seriesA1 = api.Timeseries{[]float64{1, 2, 3}, api.TaggedMetric{"cpu", map[string]string{"dc": "A", "host": "#1"}}}
	seriesA2 = api.Timeseries{[]float64{4, 5, 6}, api.TaggedMetric{"cpu", map[string]string{"dc": "A", "host": "#2"}}}
	seriesB3 = api.Timeseries{[]float64{0, 1, 1}, api.TaggedMetric{"cpu", map[string]string{"dc": "B", "host": "#3"}}}
	seriesB4 = api.Timeseries{[]float64{1, 3, 2}, api.TaggedMetric{"cpu", map[string]string{"dc": "B", "host": "#4"}}}
	seriesC5 = api.Timeseries{[]float64{2, 2, 3}, api.TaggedMetric{"cpu", map[string]string{"dc": "C", "host": "#5"}}}

	seriesDC_A = api.Timeseries{[]float64{2, 0, 1}, api.TaggedMetric{"cpu", map[string]string{"dc": "A"}}}
	seriesDC_B = api.Timeseries{[]float64{2, 0, 1}, api.TaggedMetric{"cpu", map[string]string{"dc": "B"}}}
	seriesDC_C = api.Timeseries{[]float64{2, 0, 1}, api.TaggedMetric{"cpu", map[string]string{"dc": "C"}}}

	seriesENV_PROD  = api.Timeseries{[]float64{2, 0, 1}, api.TaggedMetric{"cpu", map[string]string{"env": "production"}}}
	seriesENV_STAGE = api.Timeseries{[]float64{2, 0, 1}, api.TaggedMetric{"cpu", map[string]string{"env": "staging"}}}

	voidSeries = api.Timeseries{[]float64{0, 0, 0}, api.TaggedMetric{"cpu", map[string]string{}}}

	emptyList = api.SeriesList{[]api.Timeseries{}, api.Timerange{}}
	basicList = api.SeriesList{[]api.Timeseries{seriesA1, seriesA2, seriesB3, seriesB4, seriesC5}, api.Timerange{}}
	dcList    = api.SeriesList{[]api.Timeseries{seriesDC_A, seriesDC_B, seriesDC_C}, api.Timerange{}}
	envList   = api.SeriesList{[]api.Timeseries{seriesENV_PROD, seriesENV_STAGE}, api.Timerange{}}

	voidList = api.SeriesList{[]api.Timeseries{voidSeries}, api.Timerange{}}
)

func Test_Join_Zero(t *testing.T) {
	// attempt to join empty with empty, empty with many, and many with empty, etc.
	// all should produce empty results
	result := Join([]api.SeriesList{emptyList})
	if len(result.Rows) != 0 {
		t.Fatalf("Join of single empty row produces non-empty result")
	}
	result = Join([]api.SeriesList{emptyList, emptyList})
	if len(result.Rows) != 0 {
		t.Fatalf("Join of two empty rows produces non-empty result")
	}
	result = Join([]api.SeriesList{emptyList, basicList})
	if len(result.Rows) != 0 {
		t.Fatalf("Join of (empty row, nonempty row) produces non-empty result")
	}
	result = Join([]api.SeriesList{basicList, emptyList})
	if len(result.Rows) != 0 {
		t.Fatalf("Join of single (nonempty row, empty row) produces non-empty result")
	}
	result = Join([]api.SeriesList{basicList, basicList, basicList, emptyList, basicList})
	if len(result.Rows) != 0 {
		t.Fatalf("Join of many nonempty rows and one empty row produces non-empty result")
	}
}

func Test_Join_Self(t *testing.T) {
	// attempt to join (well-behaved) sets with themselves
	// or with nothing
	result := Join([]api.SeriesList{basicList})
	if len(result.Rows) != len(basicList.Series) {
		t.Fatalf("Join of a row with nothing else changes its length")
	}

	result = Join([]api.SeriesList{basicList, basicList})
	if len(result.Rows) != len(basicList.Series) {
		t.Fatalf("Join of a row with itself changes its length")
	}

	result = Join([]api.SeriesList{dcList, dcList})
	if len(result.Rows) != len(dcList.Series) {
		t.Fatalf("Join of a row with itself changes its length")
	}

	result = Join([]api.SeriesList{envList, envList})
	if len(result.Rows) != len(envList.Series) {
		t.Fatalf("Join of a row with itself changes its length")
	}
}

func Test_Join_CartesianProduct(t *testing.T) {
	// attempt to join two different sets with nothing in common
	// and check that they multiply
	result := Join([]api.SeriesList{basicList, envList})
	if len(result.Rows) != len(basicList.Series)*len(envList.Series) {
		t.Fatalf("Join with serieslists having disjoint key sets results in unexpected number of entries (should be %v but is %v)", len(basicList.Series)*len(envList.Series), len(result.Rows))
	}
}

// A simple max function used in the Test_Join_KeySubsets which makes it less brittle
// to modifications to the series lists used as example testcases
func max(x, y int) int {
	if x < y {
		return y
	} else {
		return x
	}
}

func Test_Join_KeySubsets(t *testing.T) {
	// Attempt to join two different sets with partial keys in common
	// but where in the one with fewer keys, the common keys are uniquely identifying.
	// This is sufficient to obtain the desired behavior: len(result) = max(len(left), len(right))
	result := Join([]api.SeriesList{basicList, dcList})
	if len(result.Rows) != max(len(basicList.Series), len(dcList.Series)) {
		t.Fatalf("Join with serieslists having keysets which are a subset of one another resulted in unexpected number of entries")
	}
	// switch the order:
	result = Join([]api.SeriesList{dcList, basicList})
	if len(result.Rows) != max(len(basicList.Series), len(dcList.Series)) {
		t.Fatalf("Join with serieslists having keysets which are a subset of one another resulted in unexpected number of entries")
	}
}

func Test_Join_Single(t *testing.T) {
	// Attempt to join various series lists with a list containing a single, tagless list
	// and verify that the lists go together
	result := Join([]api.SeriesList{basicList, voidList})
	if len(result.Rows) != len(basicList.Series) {
		t.Fatalf("Join with tagless series list changed the length of the result")
	}
	result = Join([]api.SeriesList{voidList, basicList})
	if len(result.Rows) != len(basicList.Series) {
		t.Fatalf("Join with tagless series list changed the length of the result")
	}
}
