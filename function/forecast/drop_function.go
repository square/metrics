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
	"math"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
)

var FunctionDrop = function.MakeFunction("forecast.drop", func(context function.EvaluationContext, original api.SeriesList, dropTime time.Duration) function.Value {
	lastValue := float64(context.Timerange.Slots()) - dropTime.Seconds()/context.Timerange.Resolution().Seconds()
	result := make([]api.Timeseries, len(original.Series))
	for i, series := range original.Series {
		values := make([]float64, len(series.Values))
		result[i] = series
		for j := range values {
			if float64(j) < lastValue {
				values[j] = series.Values[j]
			} else {
				values[j] = math.NaN()
			}
		}
		result[i].Values = values
	}

	return api.SeriesList{
		Series:    result,
		Timerange: original.Timerange,
	}
})
