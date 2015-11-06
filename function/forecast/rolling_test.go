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
		t.Errorf("Roller model `%s` produces unexpected NaNs on input of type `%s` with %s noise", test.rollerName, test.sourceName, test.noiseName)
		return
	}
	if stats.FirstQuartile > test.maximumError.FirstQuartile || stats.Median > test.maximumError.Median || stats.ThirdQuartile > test.maximumError.ThirdQuartile {
		t.Errorf("Model `%s` fails on input `%s` with %s noise\n\terror: %s\n\ttolerance: %s", test.rollerName, test.sourceName, test.noiseName, stats.String(), test.maximumError.String())
		return
	}
	if stats.FirstQuartile+0.1 < test.maximumError.FirstQuartile || stats.Median+0.1 < test.maximumError.Median || stats.ThirdQuartile+0.1 < test.maximumError.ThirdQuartile {
		t.Errorf("You can improve the error bounds for model `%s` on input `%s` with %s noise\n\tError: %s\n\tTolerance: %s", test.rollerName, test.sourceName, test.noiseName, stats.String(), test.maximumError.String())
	}
}

type rollingTest struct {
	roller       func([]float64, int) []float64
	rollerName   string
	source       func() ([]float64, int)
	sourceName   string
	noiser       func([]float64) []float64
	noiseName    string
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
	tests := []rollingTest{
		{
			roller:     parameters(RollingMultiplicativeHoltWinters, 0.5, 0.5, 0.6),
			rollerName: "Rolling Multiplicative Holt-Winters",
			source:     pureMultiplicativeHoltWintersSource,
			sourceName: "pure random Holt-Winters model instance",
			noiseName:  "no",
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
			noiseName:  "gaussian (strength 1)",
			maximumError: statisticalSummary{
				FirstQuartile: 1.2,
				Median:        2.6,
				ThirdQuartile: 6.7,
			},
		},
		{
			roller:     parameters(RollingMultiplicativeHoltWinters, 0.36, 0.36, 0.88),
			rollerName: "Rolling Multiplicative Holt-Winters",
			source:     pureInterpolatingMultiplicativeHoltWintersSource,
			sourceName: "time-interpolation of two pure random Holt-Winters model instances",
			noiseName:  "no",
			maximumError: statisticalSummary{
				FirstQuartile: 10.6,
				Median:        17.8,
				ThirdQuartile: 42.3,
			},
		},

		{
			roller:     parameters(RollingMultiplicativeHoltWinters, 0.36, 0.36, 0.88),
			rollerName: "Rolling Multiplicative Holt-Winters",
			source:     pureInterpolatingMultiplicativeHoltWintersSource,
			sourceName: "time-interpolation of two pure random Holt-Winters model instances",
			noiser:     gaussianNoise,
			noiseName:  "gaussian (strength 1)",
			maximumError: statisticalSummary{
				FirstQuartile: 11.0,
				Median:        18.4,
				ThirdQuartile: 42.95,
			},
		},
	}
	for _, test := range tests {
		computeRMSEStatistics(t, test)
	}
}
