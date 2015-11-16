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
	"math"
	"math/rand"
	"testing"
)

func gaussianNoise(data []float64) []float64 {
	result := make([]float64, len(data))
	for i := range data {
		result[i] = data[i] + rand.ExpFloat64()
	}
	return result
}

func spikeNoise(data []float64) []float64 {
	result := make([]float64, len(data))
	min := data[0]
	max := data[0]
	for i := range data {
		result[i] = data[i]
		min = math.Min(min, data[i])
		max = math.Max(max, data[i])
	}
	// expand the range:
	size := max - min
	for i := 0; i < len(data)/100+3; i++ {
		result[rand.Intn(len(result))] = rand.Float64()*size*3 + min - size
	}
	return result
}

// computeRMSEPercentHoles computes the percent-root-mean-square-error for the given input on the given roller,
// and inserting a hole into the last quarter
func computeRMSEPercentHoles(correct []float64, period int, roller func([]float64, int) []float64, noiser func([]float64) []float64) float64 {
	// We feed noisy data into the roller, then check its result against the non-noisy data.
	noisyData := correct
	if noiser != nil {
		noisyData = noiser(correct)
	}
	// We'll have to put holes in the correct data.
	// We'll split it into 4 quadrants. The second and fourth will be missing, and must be inferred.
	training := make([]float64, len(correct))
	for i := range training {
		if i < 3*len(training)/4 {
			training[i] = noisyData[i]
		} else {
			training[i] = math.NaN()
		}
	}
	guess := roller(training, period)
	// Evaluate the RMSE for the holes
	count := 0
	rmse := 0.0      // root mean squared error
	magnitude := 0.0 // magnitude of correct values
	for i := range training {
		if !math.IsNaN(training[i]) {
			continue
		}
		count++
		rmse += (correct[i] - guess[i]) * (correct[i] - guess[i])
		magnitude += math.Abs(correct[i])
	}
	rmse /= float64(count)
	magnitude /= float64(count)
	rmse = math.Sqrt(rmse)
	return rmse / magnitude * 100
}

func computeRMSEStatistics(t *testing.T, test rollingTest) {
	n := 10000
	results := make([]float64, n)
	for i := range results {
		correct, period := test.source()
		results[i] = computeRMSEPercentHoles(correct, period, test.roller, test.noiser)
	}
	stats := summarizeSlice(results)
	improvement := stats.improvementOver(test.maximumError)
	if math.IsNaN(improvement) {
		t.Errorf("Roller model `%s` produces unexpected NaNs on input of type `%s` with %s noise", test.rollerName, test.sourceName, test.noiserName)
		return
	}
	if stats.FirstQuartile > test.maximumError.FirstQuartile+1 || stats.Median > test.maximumError.Median+1 || stats.ThirdQuartile > test.maximumError.ThirdQuartile+1 {
		t.Errorf("Model `%s` fails on input `%s` with %s noise\n\terror: %s\n\ttolerance: %s", test.rollerName, test.sourceName, test.noiserName, stats.String(), test.maximumError.String())
		return
	}
	if stats.FirstQuartile+1 < test.maximumError.FirstQuartile || stats.Median+1 < test.maximumError.Median || stats.ThirdQuartile+1 < test.maximumError.ThirdQuartile {
		t.Errorf("You can improve the error bounds for model `%s` on input `%s` with %s noise\n\tError: %s\n\tTolerance: %s", test.rollerName, test.sourceName, test.noiserName, stats.String(), test.maximumError.String())
		return
	}
}

type rollingTest struct {
	roller       func([]float64, int) []float64
	rollerName   string
	source       func() ([]float64, int)
	sourceName   string
	noiser       func([]float64) []float64
	noiserName   string
	maximumError statisticalSummary
}

func parameters(fun func([]float64, int, float64, float64, float64) []float64, a float64, b float64, c float64) func([]float64, int) []float64 {
	return func(xs []float64, p int) []float64 {
		return fun(xs, p, a, b, c)
	}
}

// TestRollingAccuracy tests how accurate the rolling forecast functions are.
// For example, those that use exponential smoothing to estimate the parameters of the Multiplicative Holt-Winters model.
// They must be tested differently than others, due to the fact that they don't receive separate training data and prediction intervals.
func TestRollingAccuracy(t *testing.T) {
	// Note: the sample size is not large enough for the tolerances below to be precise.
	// If the random seed is changed, they will likely need to be changed.
	// Increasing the sample size in computeRMSEStatistics will reduce this effect.
	tests := []rollingTest{
		{
			roller:     parameters(RollingMultiplicativeHoltWinters, 0.5, 0.5, 0.6),
			rollerName: "Rolling Multiplicative Holt-Winters",
			source:     pureMultiplicativeHoltWintersSource,
			sourceName: "pure random Holt-Winters model instance",
			noiserName: "no",
			maximumError: statisticalSummary{
				FirstQuartile: 1.0,
				Median:        2.5,
				ThirdQuartile: 6.6,
			},
		},
		{
			roller:     parameters(RollingMultiplicativeHoltWinters, 0.5, 0.5, 0.6),
			rollerName: "Rolling Multiplicative Holt-Winters",
			source:     pureMultiplicativeHoltWintersSource,
			sourceName: "pure random Holt-Winters model instance",
			noiser:     gaussianNoise,
			noiserName: "gaussian (strength 1)",
			maximumError: statisticalSummary{
				FirstQuartile: 1.2,
				Median:        2.6,
				ThirdQuartile: 6.7,
			},
		},
		{
			roller:     parameters(RollingMultiplicativeHoltWinters, 0.5, 0.4, 0.4),
			rollerName: "Rolling Multiplicative Holt-Winters",
			source:     pureMultiplicativeHoltWintersSource,
			sourceName: "pure random Holt-Winters model instance",
			noiser:     spikeNoise,
			noiserName: "spiking",
			maximumError: statisticalSummary{
				FirstQuartile: 20.1,
				Median:        47.6,
				ThirdQuartile: 137.1,
			},
		},
		{
			roller:     parameters(RollingMultiplicativeHoltWinters, 0.36, 0.36, 0.88),
			rollerName: "Rolling Multiplicative Holt-Winters",
			source:     pureInterpolatingMultiplicativeHoltWintersSource,
			sourceName: "time-interpolation of two pure random Holt-Winters model instances",
			noiserName: "no",
			maximumError: statisticalSummary{
				FirstQuartile: 10.6,
				Median:        17.9,
				ThirdQuartile: 40.8,
			},
		},
		{
			roller:     parameters(RollingMultiplicativeHoltWinters, 0.36, 0.36, 0.88),
			rollerName: "Rolling Multiplicative Holt-Winters",
			source:     pureInterpolatingMultiplicativeHoltWintersSource,
			sourceName: "time-interpolation of two pure random Holt-Winters model instances",
			noiser:     gaussianNoise,
			noiserName: "gaussian (strength 1)",
			maximumError: statisticalSummary{
				FirstQuartile: 10.9,
				Median:        18.4,
				ThirdQuartile: 42.4,
			},
		},
		{
			roller:     parameters(RollingMultiplicativeHoltWinters, 0.36, 0.36, 0.88),
			rollerName: "Rolling Multiplicative Holt-Winters",
			source:     pureInterpolatingMultiplicativeHoltWintersSource,
			sourceName: "time-interpolation of two pure random Holt-Winters model instances",
			noiser:     spikeNoise,
			noiserName: "spiking",
			maximumError: statisticalSummary{
				FirstQuartile: 17.8,
				Median:        42.3,
				ThirdQuartile: 124.6,
			},
		},
	}
	for _, test := range tests {
		computeRMSEStatistics(t, test)
	}
}
