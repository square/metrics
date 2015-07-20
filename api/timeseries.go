// timeseries related functions.
package api

import (
	"fmt"
	"time"
)

func (ts Timeseries) truncate(original Timerange, timerange Timerange) (Timeseries, error) {
	return Timeseries{}, fmt.Errorf("Implement Truncate()")
}

func (ts Timeseries) upsample(original Timerange, resolution time.Duration, interpolator func(start, end, position float64) float64) (Timeseries, error) {
	/*
		newValue := make([]float64, timerange.Slots())
		return Timeseries{
			Values: newValue,
			TagSet: ts.TagSet,
		}
	*/
	return Timeseries{}, nil
}

func (ts Timeseries) downsample(original Timerange, newRange Timerange, bucketSampler func([]float64) float64) (Timeseries, error) {
	newValue := make([]float64, newRange.Slots())
	width := float64(newRange.Resolution()) / float64(original.Resolution())
	for i := 0; i < newRange.Slots(); i++ {
		start := int(float64(i) * width)
		end := min(int(float64(i+1)*width), original.Slots())
		fmt.Printf("start=%d,end=%d\n", start, end)
		slice := ts.Values[start:end]
		newValue[i] = bucketSampler(slice)
	}
	return Timeseries{
		Values: newValue,
		TagSet: ts.TagSet,
	}, nil
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
