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
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
)

// FunctionRollingMultiplicativeHoltWinters computes a rolling multiplicative Holt-Winters model for the data.
// It takes in several learning rates, as well as the period that describes the periodicity of the seasonal term.
// The learning rates are interpreted as being "per period." For example, a value of 0.5 means that values in
// this period are effectively weighted twice as much as those in the previous. A value of 0.9 means that values in
// this period are weighted 1.0/(1.0 - 0.9) = 10 times as much as the previous.
var FunctionRollingMultiplicativeHoltWinters = function.MetricFunction{
	Name:         "forecast.rolling_multiplicative_holt_winters",
	MinArguments: 5, // Series, period, level learning rate,  trend learning rate, seasonal learning rate
	MaxArguments: 6, // Series, period, level learning rate,  trend learning rate, seasonal learning rate, extra training time
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
		extraTrainingTime := time.Duration(0)
		if len(arguments) == 6 {
			extraTrainingTime, err = function.EvaluateToDuration(arguments[5], context)
			if err != nil {
				return nil, err
			}
		}

		samples := int(period / context.Timerange.Resolution())
		if samples <= 0 {
			return nil, fmt.Errorf("forecast.rolling_multiplicative_holt_winters expects the period parameter to mean at least one slot") // TODO: use a structured error
		}

		newContext := context.Copy()
		newContext.Timerange = newContext.Timerange.ExtendBefore(extraTrainingTime)
		seriesList, err := function.EvaluateToSeriesList(arguments[0], &newContext)
		context.CopyNotesFrom(&newContext)
		newContext.Invalidate()
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
