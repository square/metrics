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
	"sort"
	"testing"
)
import "math"
import "math/rand"

// Returns the root-mean-square-error percentage of total (so that it is scale independent)
func testModelRMSEPercent(t *testing.T, source func() ([]float64, int), model func([]float64, int) (Model, error), noiser func([]float64) []float64) float64 {
	data, period := source()
	if len(data) < period*3 {
		t.Fatalf("TEST CASE ERROR: must be sufficient data; we require len(data) >= period*3")
	}
	trainingLength := rand.Intn(len(data)-period*3) + period*3 // We need at least 3 periods of data.

	testStart := rand.Intn(len(data))
	testLength := rand.Intn(len(data) - testStart)

	// We add noise to the training data, but compare the result to the original,
	// since we want to model to ignore noisy effects.
	trainingData := data
	if noiser != nil {
		trainingData = noiser(trainingData)
	}

	trainedModel, err := model(trainingData[:trainingLength], period)
	if err != nil {
		t.Errorf("Got error when evaluating model: %s", err.Error())
		return 0
	}
	guess := trainedModel.EstimateRange(testStart, testLength)
	if len(guess) != testLength {
		t.Errorf("Expected length %d but got length %d", testLength, len(guess))
		return 0
	}

	rmse := 0.0      // root mean square error
	magnitude := 0.0 // the size of the data (mean of absolute value of all correct data)
	for i := range guess {
		if math.IsNaN(guess[i]) {
			t.Errorf("Missing data in result: %+v", guess)
			return 0
		}
		correct := data[i+testStart]
		magnitude += math.Abs(correct)
		rmse += (guess[i] - correct) * (guess[i] - correct)
	}
	rmse /= float64(len(guess))
	magnitude /= float64(len(guess))

	rmse = math.Sqrt(rmse)
	return rmse / magnitude * 100
}

func testModelRMSEs(t *testing.T, source func() ([]float64, int), model func([]float64, int) (Model, error), noiser func([]float64) []float64) statisticalSummary {
	n := 2000
	result := make([]float64, n)
	for i := range result {
		result[i] = testModelRMSEPercent(t, source, model, noiser)
	}
	sort.Float64s(result)

	return summarizeSlice(result)
}

type modelTest struct {
	model        func([]float64, int) (Model, error)
	modelName    string
	source       func() ([]float64, int)
	sourceName   string
	noiser       func([]float64) []float64
	noiserName   string
	maximumError statisticalSummary
}

func applyTestForModel(t *testing.T, test modelTest) {
	modelError := testModelRMSEs(t, test.source, test.model, test.noiser)
	improvement := modelError.improvementOver(test.maximumError)
	if math.IsNaN(improvement) {
		t.Errorf("Trained model `%s` produces unexpected NaNs on input of type `%s`", test.modelName, test.sourceName)
		return
	}
	if modelError.FirstQuartile > test.maximumError.FirstQuartile+1 || modelError.Median > test.maximumError.Median+1 || modelError.ThirdQuartile > test.maximumError.ThirdQuartile+1 {
		t.Errorf("Model `%s` fails on input `%s` with %s noise\n\tError: %s\n\tTolerance: %s", test.modelName, test.sourceName, test.noiserName, modelError.String(), test.maximumError.String())
		return
	}
	if modelError.FirstQuartile+1 < test.maximumError.FirstQuartile || modelError.Median+1 < test.maximumError.Median || modelError.ThirdQuartile+1 < test.maximumError.ThirdQuartile {
		t.Errorf("You can improve the error bounds by %f for model `%s` on input `%s` with %s noise\n\tError: %s\n\tTolerance: %s", improvement, test.modelName, test.sourceName, test.noiserName, modelError.String(), test.maximumError.String())
		return
	}
}

// TestModelAccuracy acts primarily as a sanity check and a regression test.
// It calculates the root-mean-square-error as a percentage of the mean data magnitude (so that it is scale-independent)
// as a means for evaluating the accuracy of various models.
// The models are each tried on many inputs, and the 1st (25), 2nd (50), and 3rd (75) quartiles of error are recorded.
// These quartiles are compared to the limits established by the test.
func TestModelAccuracy(t *testing.T) {
	// The model's accuracy varies, depending on how exactly the noise affects it.

	tests := []modelTest{
		{
			model:      TrainGeneralizedHoltWintersModel,
			modelName:  "Generalized Holt Winters Model",
			source:     pureMultiplicativeHoltWintersSource,
			sourceName: "Random Holt-Winters model instance",
			noiserName: "no",
			maximumError: statisticalSummary{ // Should be perfect, up to FP error.
				FirstQuartile: 0.00001,
				Median:        0.00001,
				ThirdQuartile: 0.00001,
			},
		},
		{
			model:      TrainGeneralizedHoltWintersModel,
			modelName:  "Generalized Holt Winters Model",
			source:     pureMultiplicativeHoltWintersSource,
			sourceName: "Random Holt-Winters model instance",
			noiser:     gaussianNoise,
			noiserName: "guassian (strength 1)",
			maximumError: statisticalSummary{ // Do not expect it to do perfectly, since there's error
				FirstQuartile: 0.4,
				Median:        0.65,
				ThirdQuartile: 1.2,
			},
		},

		{
			model:      TrainMultiplicativeHoltWintersModel,
			modelName:  "Multiplicative Holt Winters Model",
			source:     pureMultiplicativeHoltWintersSource,
			sourceName: "Random Holt-Winters model instance",
			noiserName: "no",
			maximumError: statisticalSummary{ // Should be perfect, up to FP error.
				FirstQuartile: 0.00001,
				Median:        0.00001,
				ThirdQuartile: 0.00001,
			},
		},
		{
			model:      TrainMultiplicativeHoltWintersModel,
			modelName:  "Multiplicative Holt Winters Model",
			source:     pureMultiplicativeHoltWintersSource,
			sourceName: "Random Holt-Winters model instance",
			noiser:     gaussianNoise,
			noiserName: "gaussian (strength 1)",
			maximumError: statisticalSummary{ // Do not expect it to do perfectly, since there's error
				FirstQuartile: 2.3,
				Median:        5.9,
				ThirdQuartile: 17.9,
			},
		},
	}
	for _, test := range tests {
		applyTestForModel(t, test)
	}
}
