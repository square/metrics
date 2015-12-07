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

	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
)

// FunctionAnomalyMaker makes anomaly-measurement functions that return simple p-values for deviations from the predicted model.
// In order to make this procedure mostly automatic, it performs a join on the original tagsets to match them up with their predictions.
func FunctionPeriodicAnomalyMaker(name string, model function.MetricFunction) function.MetricFunction {
	if model.MinArguments < 2 {
		panic("FunctionAnomalyMaker requires that the model argument take at least two parameters; series and period.")
	}
	return function.MetricFunction{
		Name:         name,
		MinArguments: model.MinArguments,
		MaxArguments: model.MaxArguments,
		Compute: func(context function.EvaluationContext, arguments []function.Expression, groups function.Groups) (function.Value, error) {
			original, err := function.EvaluateToSeriesList(arguments[0], context)
			if err != nil {
				return nil, err
			}
			// TODO: improve sharing by using the `original` value as the first argument to the arguments,
			// since the context is known to be the same and therefore it should evaluate identically.
			// There is currently no standard "literal series list node" or similar that we could use for this purepose.
			predictionValue, err := model.Compute(context, arguments, groups)
			if err != nil {
				return nil, err // TODO: add decoration to describe it's coming from the anomaly function
			}
			prediction, err := predictionValue.ToSeriesList(context.Timerange)
			if err != nil {
				return nil, err
			}
			period, err := function.EvaluateToDuration(arguments[1], context)
			if err != nil {
				return nil, err
			}
			periodSlots := int(period / context.Timerange.Resolution())
			// Now we need to match up 'original' and 'prediction'
			// We'll use a hashmap for now.
			// TODO: clean this up to hog less memory
			lookup := map[string][]float64{}
			for _, series := range original.Series {
				lookup[series.TagSet.Serialize()] = series.Values
			}

			result := make([]api.Timeseries, len(prediction.Series))
			for i, series := range prediction.Series {
				result[i] = series
				result[i].Values, err = periodicStandardDeviationsFromExpected(lookup[series.TagSet.Serialize()], series.Values, periodSlots)
				if err != nil {
					return nil, err
				}
			}
			prediction.Series = result
			return prediction, nil
		},
	}
}

var FunctionAnomalyRollingMultiplicativeHoltWinters = FunctionPeriodicAnomalyMaker("forecast.anomaly_rolling_multiplicative_holt_winters", FunctionRollingMultiplicativeHoltWinters)
var FunctionAnomalyRollingSeasonal = FunctionPeriodicAnomalyMaker("forecast.anomaly_rolling_seasonal", FunctionRollingSeasonal)

func standardDeviationsFromExpected(correct []float64, estimate []float64) ([]float64, error) {
	if len(correct) != len(estimate) {
		return nil, fmt.Errorf("p-value calculation requires two lists of equal size")
	}
	differences := []float64{}
	for i := range correct {
		if math.IsInf(correct[i], 0) || math.IsNaN(correct[i]) {
			continue
		}
		if math.IsInf(estimate[i], 0) || math.IsNaN(estimate[i]) {
			continue
		}
		differences = append(differences, estimate[i]-correct[i])
	}
	meanDifference := 0.0
	for _, difference := range differences {
		meanDifference += difference
	}
	meanDifference /= float64(len(differences))

	//
	stddevDifference := 0.0
	for _, difference := range differences {
		stddevDifference += math.Pow(difference-meanDifference, 2)
	}
	stddevDifference /= float64(len(differences)) - 1
	stddevDifference = math.Sqrt(stddevDifference)
	// stddevDifference estimates the true population standard deviation of the differences between the estimate and the correct values.
	// We now use this value to standardize our differences.
	standardDifferences := make([]float64, len(estimate))
	for i := range standardDifferences {
		difference := (estimate[i] - correct[i])
		standardDifferences[i] = (difference - meanDifference) / stddevDifference
	}
	return standardDifferences, nil
}
func periodicStandardDeviationsFromExpected(correct []float64, estimate []float64, period int) ([]float64, error) {
	if period <= 0 {
		return nil, fmt.Errorf("Period must be strictly positive")
	}
	if len(correct) != len(estimate) {
		return nil, fmt.Errorf("to estimate anomaly values, the ground truth and estimate slices must be the same length")
	}
	slices := make([][]float64, period)
	for r := range slices {
		// Consider len(correct) = 42
		// If our period is 10, what are the indices for each group?
		// 0: [0,10,20,30,40]
		// 1: [1,11,21,31,41]
		// 2: [2,12,22,32]
		// 3: [3,13,23,33]
		// ...
		// 9: [9,19,29,39]

		// each slot r, (0 <= r < p) gets (n/p + (n%p < r ? 1 : 0)) indices.
		length := len(correct) / period
		if r < len(correct)%period {
			length++
		}
		correctSlice := make([]float64, length)
		estimateSlice := make([]float64, length)
		for i := range correctSlice {
			correctSlice[i] = correct[r+i*period]
			estimateSlice[i] = estimate[r+i*period]
		}
		var err error
		slices[r], err = standardDeviationsFromExpected(correctSlice, estimateSlice)
		if err != nil {
			return nil, err
		}
	}
	// un-interleave the slices
	answer := make([]float64, len(correct))
	for i := range answer {
		answer[i] = slices[i%period][i/period]
	}
	return answer, nil
}
