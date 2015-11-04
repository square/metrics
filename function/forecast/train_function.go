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

var FunctionTrainGeneralizedHoltWinters = function.MetricFunction{
	Name:         "forecast.train_generalized_holt_winters",
	MinArguments: 4, // Series, period, start time of training, end time of training
	MaxArguments: 4,
	Compute: func(context *function.EvaluationContext, arguments []function.Expression, groups function.Groups) (function.Value, error) {
		period, err := function.EvaluateToDuration(arguments[1], context)
		if err != nil {
			return nil, err
		}
		periodSamples := int(period / context.Timerange.Resolution())
		if periodSamples <= 0 {
			return nil, fmt.Errorf("forecast.train_generalized_holt_winters expected the period to exceed the resolution") // TODO: use structured error
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
			return nil, fmt.Errorf("forecast.train_generalized_holt_winters expected the end time to come after the start time") // TODO: use a structured error
		}
		newContext := context.Copy()
		newTimerange, err := api.NewSnappedTimerange(context.Timerange.End()-start.Nanoseconds()/1e6, context.Timerange.End()-end.Nanoseconds()/1e6, context.Timerange.ResolutionMillis())
		if err != nil {
			return nil, err
		}
		newContext.Timerange = newTimerange
		trainingSeries, err := function.EvaluateToSeriesList(arguments[0], &newContext)
		context.CopyNotesFrom(&newContext)
		newContext.Invalidate()

		// Run the series through the generalized Holt-Winters model estimator, and then use this model to estimate the current timerange.

		result := api.SeriesList{
			Name:      trainingSeries.Name,
			Query:     fmt.Sprintf("forecast.train_generalized_holt_winters(%s, %s, %s, %s)", trainingSeries.Query, period.String(), start.String(), end.String()),
			Timerange: context.Timerange,
			Series:    make([]api.Timeseries, len(trainingSeries.Series)),
		}

		// How far in the future the fetch time is than the training time.
		timeOffset := int(-start / context.Timerange.Resolution())

		for i := range result.Series {
			trainingData := trainingSeries.Series[i]
			model, err := TrainGeneralizedHoltWintersModel(trainingData.Values, periodSamples)
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
