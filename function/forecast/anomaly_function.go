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
	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
)

// FunctionAnomalyMaker makes anomaly-measurement functions that return simple p-values for deviations from the predicted model.
// In order to make this procedure mostly automatic, it performs a join on the original tagsets to match them up with their predictions.
func FunctionAnomalyMaker(name string, model function.MetricFunction) function.MetricFunction {
	if model.MinArguments < 1 {
		panic("FunctionAnomalyMaker requires that the model argument take at least one parameter.")
	}
	return function.MetricFunction{
		Name:         name,
		MinArguments: model.MinArguments,
		MaxArguments: model.MaxArguments,
		Compute: func(context *function.EvaluationContext, arguments []function.Expression, groups function.Groups) (function.Value, error) {
			original, err := function.EvaluateToSeriesList(arguments[0], context)
			if err != nil {
				return nil, err
			}
			predictionValue, err := model.Compute(context, arguments, groups)
			if err != nil {
				return nil, err // TODO: add decoration to describe it's coming from the anomaly function
			}
			prediction, err := predictionValue.ToSeriesList(context.Timerange)
			if err != nil {
				return nil, err
			}
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
				result[i].Values, err = pValueFromNormalDifferences(lookup[series.TagSet.Serialize()], series.Values)
				if err != nil {
					return nil, err
				}
			}
			prediction.Series = result
			return prediction, nil
		},
	}
}

var FunctionAnomalyRollingMultiplicativeHoltWinters = FunctionAnomalyMaker("forecast.anomaly_rolling_multiplicative_holt_winters", FunctionRollingMultiplicativeHoltWinters)
var FunctionAnomalyTrainGeneralizedHoltWinters = FunctionAnomalyMaker("forecast.anomaly_train_generalized_holt_winters_model", FunctionTrainGeneralizedHoltWinters)
var FunctionAnomalyTrainMultiplicativeHoltWinters = FunctionAnomalyMaker("forecast.anomaly_train_multiplicative_holt_winters_model", FunctionTrainMultiplicativeHoltWinters)
