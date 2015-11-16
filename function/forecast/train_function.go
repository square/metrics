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

func functionTrainModel(name string, trainer func(ys []float64, period int) (Model, error)) function.MetricFunction {
	return function.MetricFunction{
		Name:         name,
		MinArguments: 4, // Series, period, start time of training, end time of training
		MaxArguments: 4,
		Compute: func(context *function.EvaluationContext, arguments []function.Expression, groups function.Groups) (function.Value, error) {
			period, err := function.EvaluateToDuration(arguments[1], context)
			if err != nil {
				return nil, err
			}
			periodSamples := int(period / context.Timerange.Resolution())
			if periodSamples <= 0 {
				return nil, fmt.Errorf("%s expected the period to exceed the resolution", name) // TODO: use structured error
			}
			start, err := function.EvaluateToDuration(arguments[2], context)
			if err != nil {
				return nil, err
			}
			end, err := function.EvaluateToDuration(arguments[3], context)
			if err != nil {
				return nil, err
			}
			if end < start {
				return nil, fmt.Errorf("%s expected the end time to come after the start time", name) // TODO: use a structured error
			}
			newContext := context.Copy()
			newTimerange, err := api.NewSnappedTimerange(context.Timerange.End()+start.Nanoseconds()/1e6, context.Timerange.End()+end.Nanoseconds()/1e6, context.Timerange.ResolutionMillis())
			if err != nil {
				return nil, err
			}
			newContext.Timerange = newTimerange
			trainingSeries, err := function.EvaluateToSeriesList(arguments[0], &newContext)
			context.CopyNotesFrom(&newContext)
			newContext.Invalidate()

			// Run the series through the model estimator, and then use this model to estimate the current timerange.

			result := api.SeriesList{
				Name:      trainingSeries.Name,
				Query:     fmt.Sprintf("%s(%s, %s, %s, %s)", name, trainingSeries.Query, period.String(), start.String(), end.String()),
				Timerange: context.Timerange,
				Series:    make([]api.Timeseries, len(trainingSeries.Series)),
			}

			// How far in the future the fetch time is than the training time.
			// In general, 'start' will be negative. Therefore, we negate it here,
			// so that data will be fetched in the model's future (the present)
			// since it was trained in the past.
			timeOffset := int(-start / context.Timerange.Resolution())

			for i := range result.Series {
				trainingData := trainingSeries.Series[i]
				model, err := trainer(trainingData.Values, periodSamples)
				if err != nil {
					return nil, err // TODO: add further explanatory message
				}
				estimate := model.EstimateRange(timeOffset, len(trainingData.Values))
				result.Series[i] = api.Timeseries{
					Values: estimate,
					TagSet: trainingData.TagSet,
				}
			}

			return result, nil
		},
	}
}

// FunctionTrainGeneralizedHoltWinters is a MetricFunction that trains a Generalized Holt-Winters model on some time interval,
// then uses the predicted model to estimate the data for the currently requested interval.
var FunctionTrainGeneralizedHoltWinters = functionTrainModel("forecast.train_generalized_holt_winters_model", TrainGeneralizedHoltWintersModel)

// FunctionTrainMultiplicativeHoltWinters is a MetricFunction that trains a multiplicative Holt-Winters model on some time interval,
// then uses the predicted model to estimate the data for the currently requested interval.
var FunctionTrainMultiplicativeHoltWinters = functionTrainModel("forecast.train_multiplicative_holt_winters_model", TrainMultiplicativeHoltWintersModel)
