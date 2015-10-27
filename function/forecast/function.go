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

// ForecastFunction computes a Holt-Winters model for the given time series.
var ForecastFunction = function.MetricFunction{
	Name:         "forecast.holt_winters",
	MinArguments: 4, // series, period, time offset to train, length of training period
	MaxArguments: 4,
	Compute: func(context *function.EvaluationContext, arguments []function.Expression, groups function.Groups) (function.Value, error) {
		periodValue, err := arguments[1].Evaluate(context)
		if err != nil {
			return nil, err
		}
		period, err := periodValue.ToDuration()
		if err != nil {
			return nil, err
		}
		if period <= 0 {
			return nil, fmt.Errorf("forecast.holt_winters expected period to be positive") // TODO: use a structured error
		}
		whenValue, err := arguments[2].Evaluate(context)
		if err != nil {
			return nil, err
		}
		when, err := whenValue.ToDuration()
		if err != nil {
			return nil, err
		}
		lengthValue, err := arguments[3].Evaluate(context)
		if err != nil {
			return nil, err
		}
		length, err := lengthValue.ToDuration()
		if err != nil {
			return nil, err
		}
		// We need to perform a fetch of length 'length' offset 'when' for all this data.
		// Then we apply the Holt-Winters model to each of the resulting series.
		newContext := *context
		newContext.Timerange = newContext.Timerange.Shift(when).SelectLength(length)
		data, err := arguments[0].Evaluate(&newContext)
		newContext.Invalidate()
		if err != nil {
			return nil, err
		}

		original, err := data.ToSeriesList(context.Timerange)
		if err != nil {
			return nil, err
		}

		result := api.SeriesList{
			Series:    make([]api.Timeseries, len(original.Series)),
			Timerange: context.Timerange,
			Name:      original.Name,
			Query:     fmt.Sprintf("forecast.holt_winters(%s, %s, %s, %s)", original.Query, period.String(), when.String(), length.String()),
		}
		slotTrainingStart := int(when / context.Timerange.Resolution())
		slotQueryStart := int(context.Timerange.Start() / context.Timerange.ResolutionMillis())
		for s := range result.Series {
			training := original.Series[s].Values
			model, err := HoltWintersMultiplicativeEstimate(training, int(period/context.Timerange.Resolution()))
			if err != nil {
				return nil, err // TODO: determine if there's a more graceful way to indicate the error - probably not
			}
			result.Series[s] = api.Timeseries{
				TagSet: original.Series[s].TagSet,
				Raw:    original.Series[s].Raw,
				Values: model.EstimateRange(slotQueryStart-slotTrainingStart, context.Timerange.Slots()),
			}
		}
		return result, nil
	},
}
