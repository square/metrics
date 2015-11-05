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

package forecast

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

func randomSlice(n int) []float64 {
	slice := make([]float64, n)
	for i := range slice {
		slice[i] = rand.ExpFloat64()
	}
	return slice
}

func pureMultiplicativeHoltWintersSource() ([]float64, int) {
	period := rand.Intn(10) + 10
	length := 10*period + rand.Intn(period)
	season := randomSlice(period)
	result := make([]float64, length)

	trend := rand.Float64()*4 - 2
	level := rand.Float64()*100 + 200

	for i := range result {
		result[i] = season[i%period] * (trend*float64(i) + level)
	}
	return result, period
}

func addNoiseToSource(source func() ([]float64, int), strength float64) func() ([]float64, int) {
	return func() ([]float64, int) {
		data, period := source()
		for i := range data {
			data[i] += rand.ExpFloat64() * strength
		}
		return data, period
	}
}

type statisticalSummary struct {
	FirstQuartile float64
	Median        float64
	ThirdQuartile float64
}

func (s statisticalSummary) String() string {
	return fmt.Sprintf("First quartile: %f  Median: %f  Third quartile: %f", s.FirstQuartile, s.Median, s.ThirdQuartile)
}

func (s statisticalSummary) improvementOver(other statisticalSummary) float64 {
	return math.Min(other.FirstQuartile-s.FirstQuartile, math.Min(other.Median-s.Median, other.ThirdQuartile-s.ThirdQuartile))
}
func summarizeSlice(slice []float64) statisticalSummary {
	sort.Float64s(slice)
	return statisticalSummary{
		FirstQuartile: slice[len(slice)/4],
		Median:        slice[len(slice)/2],
		ThirdQuartile: slice[len(slice)/4*3],
	}
}
