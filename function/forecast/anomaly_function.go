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

var FunctionAnomalyRollingMultiplicativeHoltWinters = function.MetricFunction{
	Name:         "forecast.anomaly_rolling_multiplicative_holt_winters",
	MinArguments: 5, // Series, period, level learning rate,  trend learning rate, seasonal learning rate,
	MaxArguments: 5,
	Compute: func(context *function.EvaluationContext, arguments []function.Expression, groups function.Groups) (function.Value, error) {
		period, err := function.EvaluateToDuration(arguments[1], context)
		if err != nil {
			return nil, err
		}
		levelLearningRate, err := function.EvaluateToScalar(arguments[2], context)
		if err != nil {
			return nil, err
		}
		trendLearningRate, err := function.EvaluateToScalar(arguments[3], context)
		if err != nil {
			return nil, err
		}
		seasonalLearningRate, err := function.EvaluateToScalar(arguments[4], context)
		if err != nil {
			return nil, err
		}

		samples := int(period / context.Timerange.Resolution())
		if samples <= 0 {
			return nil, fmt.Errorf("forecast.rolling_multiplicative_holt_winters expects the period parameter to mean at least one slot") // TODO: use a structured error
		}

		seriesList, err := function.EvaluateToSeriesList(arguments[0], context)
		if err != nil {
			return nil, err
		}

		result := api.SeriesList{
			Series:    make([]api.Timeseries, len(seriesList.Series)),
			Timerange: context.Timerange,
			Name:      seriesList.Name,
			Query:     fmt.Sprintf("forecast.rolling_multiplicative_holt_winters(%s, %s, %f, %f)", seriesList.Query, period.String(), seasonalLearningRate, trendLearningRate),
		}

		for seriesIndex, series := range seriesList.Series {
			result.Series[seriesIndex] = api.Timeseries{
				TagSet: series.TagSet,
				Raw:    series.Raw,
				Values: RollingMultiplicativeHoltWinters(series.Values, samples, levelLearningRate, trendLearningRate, seasonalLearningRate),
			}
		}

		return result, nil
	},
}
