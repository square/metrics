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

package blueflood

import (
	"math"

	"github.com/square/metrics/api"
	"github.com/square/metrics/timeseries"
)

type sampler struct {
	fieldName    string                          // Name of field in Blueflood JSON response
	selectField  func(point metricPoint) float64 // Function for extracting field from metricPoint
	sampleBucket func([]float64) float64         // Function to sample from the bucket (e.g., min, mean, max)
}

// sampleResult samples the points into a uniform slice of float64s.
func samplePoints(points []metricPoint, timerange api.Timerange, sampler sampler) []float64 {
	// A bucket holds a set of points corresponding to one interval in the result.
	buckets := make([][]float64, timerange.Slots())
	for _, point := range points {
		pointValue := sampler.selectField(point)
		index := (point.Timestamp - timerange.StartMillis()) / timerange.ResolutionMillis()
		if index < 0 || int(index) >= len(buckets) {
			continue
		}
		buckets[index] = append(buckets[index], pointValue)
	}

	// values will hold the final values to be returned as the series.
	values := make([]float64, timerange.Slots())

	for i, bucket := range buckets {
		if len(bucket) == 0 {
			values[i] = math.NaN()
			continue
		}
		values[i] = sampler.sampleBucket(bucket)
	}
	return values
}

var samplerMap = map[timeseries.SampleMethod]sampler{
	timeseries.SampleMean: {
		fieldName:   "average",
		selectField: func(point metricPoint) float64 { return point.Average },
		sampleBucket: func(bucket []float64) float64 {
			value := 0.0
			count := 0
			for _, v := range bucket {
				if !math.IsNaN(v) {
					value += v
					count++
				}
			}
			return value / float64(count)
		},
	},
	timeseries.SampleMin: {
		fieldName:   "min",
		selectField: func(point metricPoint) float64 { return point.Min },
		sampleBucket: func(bucket []float64) float64 {
			smallest := math.NaN()
			for _, v := range bucket {
				if math.IsNaN(v) {
					continue
				}
				if math.IsNaN(smallest) {
					smallest = v
				} else {
					smallest = math.Min(smallest, v)
				}
			}
			return smallest
		},
	},
	timeseries.SampleMax: {
		fieldName:   "max",
		selectField: func(point metricPoint) float64 { return point.Max },
		sampleBucket: func(bucket []float64) float64 {
			largest := math.NaN()
			for _, v := range bucket {
				if math.IsNaN(v) {
					continue
				}
				if math.IsNaN(largest) {
					largest = v
				} else {
					largest = math.Max(largest, v)
				}
			}
			return largest
		},
	},
}
