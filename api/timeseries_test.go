package api

import (
	"math"
	"testing"

	"github.com/square/metrics/assert"
)

func max(array []float64) float64 {
	max := array[0]
	for _, v := range array {
		max = math.Max(max, v)
	}
	return max
}

func TestTimeseries_Downsample(t *testing.T) {
	a := assert.New(t)
	for _, suite := range []struct {
		input      []float64
		inputRange Timerange
		newRange   Timerange
		sampler    func([]float64) float64
		expected   []float64
	}{
		{[]float64{1, 2, 3, 4, 5}, Timerange{0, 4, 1}, Timerange{0, 4, 2}, max, []float64{2, 4, 5}},
	} {
		tagset := ParseTagSet("key=value")
		ts := Timeseries{suite.input, tagset}
		sampled, err := ts.downsample(suite.inputRange, suite.newRange, suite.sampler)
		a.CheckError(err)
		a.Eq(sampled.Values, suite.expected)
	}
}
