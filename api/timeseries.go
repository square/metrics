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

// Package api holds common data types and public interface exposed by the indexer library.
// Refer to the doc
// https://docs.google.com/a/squareup.com/document/d/1k0Wgi2wnJPQoyDyReb9dyIqRrD8-v0u8hz37S282ii4/edit
// for the terminology.

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"
)

// Timeseries is a single time series, identified with the associated tagset.
type Timeseries struct {
	Values []float64
	TagSet TagSet
	Raw    []byte
}

// MarshalJSON exists to manually encode floats.
func (ts Timeseries) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	var scratch [64]byte
	buffer.WriteByte('{')
	buffer.WriteString("\"tagset\":")
	tagset, err := json.Marshal(ts.TagSet)
	if err != nil {
		return []byte{}, err
	}
	buffer.Write(tagset)
	buffer.WriteByte(',')
	buffer.WriteString("\"values\":")
	buffer.WriteByte('[')
	n := len(ts.Values)
	for i := 0; i < n; i++ {
		if i > 0 {
			buffer.WriteByte(',')
		}
		f := ts.Values[i]
		if math.IsInf(f, 1) {
			buffer.WriteString("null") // TODO - positive infinity
		} else if math.IsInf(f, -1) {
			buffer.WriteString("null") // TODO - negative infinity
		} else if math.IsNaN(f) {
			buffer.WriteString("null")
		} else {
			b := strconv.AppendFloat(scratch[:0], f, 'g', -1, 64)
			buffer.Write(b)
		}
	}
	buffer.WriteByte(']')
	buffer.WriteByte('}')
	return buffer.Bytes(), err
}

func (ts Timeseries) Truncate(original Timerange, timerange Timerange) (Timeseries, error) {
	return Timeseries{}, fmt.Errorf("Implement Truncate()")
}

func (ts Timeseries) Upsample(original Timerange, resolution time.Duration, interpolator func(start, end, position float64) float64) (Timeseries, error) {
	/*
		newValue := make([]float64, timerange.Slots())
		return Timeseries{
			Values: newValue,
			TagSet: ts.TagSet,
		}
	*/
	return Timeseries{}, nil
}

func (ts Timeseries) Downsample(original Timerange, newRange Timerange, bucketSampler func([]float64) float64) (Timeseries, error) {
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
