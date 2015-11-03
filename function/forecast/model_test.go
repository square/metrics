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

import "testing"
import "math"
import "math/rand"

func randomModel() ([]float64, int) {
	period := rand.Intn(10) + 10
	length := 10 * period
	season := make([]float64, period)
	sum := 0.0
	for i := range season {
		season[i] = math.Abs(rand.ExpFloat64()) + 1
		sum += season[i]
	}
	for i := range season {
		// Normalize them so that their mean is 1
		season[i] /= sum / float64(len(season))
	}
	trend := rand.ExpFloat64()
	start := rand.ExpFloat64() * 20
	result := make([]float64, length)
	for i := range result {
		result[i] = (start + trend*float64(i)) * season[i%period]
	}
	//t.Logf("Season: %+v", season)
	//t.Logf("Trend: %f", trend)
	//t.Logf("Start: %f", start)
	return result, period
}

func noisyRandomModel() ([]float64, int) {
	data, period := randomModel()
	for i := range data {
		data[i] += rand.ExpFloat64()
	}
	return data, period
}

func randomEvaluateModel(t *testing.T, epsilon float64, source func() ([]float64, int), model func([]float64, int) (Model, error)) {
	data, period := source()
	if len(data) < period*3 {
		t.Fatalf("TEST CASE ERROR: must be sufficient data; we require len(data) >= period*3")
	}
	partialLength := rand.Intn(len(data)-period*2) + period*2

	start := rand.Intn(len(data))
	length := rand.Intn(len(data) - start)

	trainedModel, _ := model(data[:partialLength], period)
	guess := trainedModel.EstimateRange(start, length)
	if len(guess) != length {
		t.Errorf("Expected length %d but got length %d", length, len(guess))
		return
	}

	// root mean square error
	rmse := 0.0

	for i := range guess {
		if math.IsNaN(guess[i]) {
			t.Errorf("Missing data in result: %+v", guess)
			return
		}
		correct := data[i+start]
		rmse += (guess[i] - correct) * (guess[i] - correct)
	}
	rmse /= float64(len(guess))
	rmse = math.Sqrt(rmse)
	if rmse > epsilon {
		t.Errorf("Root-mean-square error is %f which exceeds %f; \nexpected: %+v\nestimate: %+v", rmse, epsilon, data[start:start+length], guess)
	}
}

func TestModel(t *testing.T) {
	// The model's accuracy varies, depending on how exactly the noise affects it.
	for i := 0; i < 1000; i++ {
		randomEvaluateModel(t, 0.001, randomModel, EstimateGeneralizedHoltWintersModel)
		randomEvaluateModel(t, 20, noisyRandomModel, EstimateGeneralizedHoltWintersModel)
	}
}
