// Copyright 2015 - 2016 Square Inc.
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
var FunctionRollingMultiplicativeHoltWinters = function.MakeFunction(
	"forecast.rolling_multiplicative_holt_winters",
	func(context function.EvaluationContext, seriesExpression function.Expression, period time.Duration, levelLearningRate float64, trendLearningRate float64, seasonalLearningRate float64, optionalExtraTrainingTime *time.Duration) (api.SeriesList, error) {
		extraTrainingTime := time.Duration(0)
		if optionalExtraTrainingTime != nil {
			extraTrainingTime = *optionalExtraTrainingTime
		}
		if extraTrainingTime < 0 {
			return api.SeriesList{}, fmt.Errorf("Extra training time must be non-negative, but got %s", extraTrainingTime.String()) // TODO: use structured error
		}

		samples := int(period / context.Timerange.Resolution())
		if samples <= 0 {
			return api.SeriesList{}, fmt.Errorf("forecast.rolling_multiplicative_holt_winters expects the period parameter to mean at least one slot") // TODO: use a structured error
		}

		newContext := context.WithTimerange(context.Timerange.ExtendBefore(extraTrainingTime))
		extraSlots := newContext.Timerange.Slots() - context.Timerange.Slots()
		seriesList, err := function.EvaluateToSeriesList(seriesExpression, newContext)
		if err != nil {
			return api.SeriesList{}, err
		}

		result := api.SeriesList{
			Series: make([]api.Timeseries, len(seriesList.Series)),
		}

		for seriesIndex, series := range seriesList.Series {
			result.Series[seriesIndex] = api.Timeseries{
				TagSet: series.TagSet,
				Values: RollingMultiplicativeHoltWinters(series.Values, samples, levelLearningRate, trendLearningRate, seasonalLearningRate)[extraSlots:], // Slice to drop the first few extra slots from the result
			}
		}

		return result, nil
	},
)

// FunctionRollingSeasonal is a forecasting MetricFunction that performs the rolling seasonal estimation.
// It is designed for data which shows seasonality without trends, although which a high learning rate it can
// perform tolerably well on data with trends as well.
var FunctionRollingSeasonal = function.MakeFunction(
	"forecast.rolling_seasonal",
	func(context function.EvaluationContext, seriesExpression function.Expression, period time.Duration, seasonalLearningRate float64, optionalExtraTrainingTime *time.Duration) (api.SeriesList, error) {
		extraTrainingTime := time.Duration(0)
		if optionalExtraTrainingTime != nil {
			extraTrainingTime = *optionalExtraTrainingTime
		}
		if extraTrainingTime < 0 {
			return api.SeriesList{}, fmt.Errorf("Extra training time must be non-negative, but got %s", extraTrainingTime.String()) // TODO: use structured error
		}

		samples := int(period / context.Timerange.Resolution())
		if samples <= 0 {
			return api.SeriesList{}, fmt.Errorf("forecast.rolling_seasonal expects the period parameter to mean at least one slot") // TODO: use a structured error
		}

		newContext := context.WithTimerange(context.Timerange.ExtendBefore(extraTrainingTime))
		extraSlots := newContext.Timerange.Slots() - context.Timerange.Slots()
		seriesList, err := function.EvaluateToSeriesList(seriesExpression, newContext)
		if err != nil {
			return api.SeriesList{}, err
		}

		result := api.SeriesList{
			Series: make([]api.Timeseries, len(seriesList.Series)),
		}

		for seriesIndex, series := range seriesList.Series {
			result.Series[seriesIndex] = api.Timeseries{
				TagSet: series.TagSet,
				Values: RollingSeasonal(series.Values, samples, seasonalLearningRate)[extraSlots:], // Slice to drop the first few extra slots from the result
			}
		}

		return result, nil
	},
)

// FunctionLinear forecasts with a simple linear regression.
// For data which is mostly just a linear trend up or down, this will provide a good model of current behavior,
// as well as a good estimate of near-future behavior.
var FunctionLinear = function.MakeFunction(
	"forecast.linear",
	func(context function.EvaluationContext, seriesExpression function.Expression, optionalTrainingTime *time.Duration) (api.SeriesList, error) {
		extraTrainingTime := time.Duration(0)
		if optionalTrainingTime != nil {
			extraTrainingTime = *optionalTrainingTime
		}
		if extraTrainingTime < 0 {
			return api.SeriesList{}, fmt.Errorf("Extra training time must be non-negative, but got %s", extraTrainingTime.String()) // TODO: use structured error
		}

		newContext := context.WithTimerange(context.Timerange.ExtendBefore(extraTrainingTime))
		extraSlots := newContext.Timerange.Slots() - context.Timerange.Slots()
		seriesList, err := function.EvaluateToSeriesList(seriesExpression, newContext)
		if err != nil {
			return api.SeriesList{}, err
		}

		result := api.SeriesList{
			Series: make([]api.Timeseries, len(seriesList.Series)),
		}

		for seriesIndex, series := range seriesList.Series {
			result.Series[seriesIndex] = api.Timeseries{
				TagSet: series.TagSet,
				Values: Linear(series.Values)[extraSlots:], // Slice to drop the first few extra slots from the result
			}
		}

		return result, nil
	},
)
