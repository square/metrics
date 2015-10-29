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

// HoltWintersModel computes a Holt-Winters model for the given time series.
var ModelHoltWinters = function.MetricFunction{
	Name:         "forecast.model_holt_winters",
	MinArguments: 4, // series, period, time offset to train, length of training period
	MaxArguments: 4,
	Compute: func(context *function.EvaluationContext, arguments []function.Expression, groups function.Groups) (function.Value, error) {
		period, err := function.EvaluateToDuration(arguments[1], context)
		if err != nil {
			return nil, err
		}
		if period <= 0 {
			return nil, fmt.Errorf("forecast.holt_winters expected period to be positive") // TODO: use a structured error
		}
		when, err := function.EvaluateToDuration(arguments[3], context)
		if err != nil {
			return nil, err
		}
		length, err := function.EvaluateToDuration(arguments[3], context)
		if err != nil {
			return nil, err
		}
		// We need to perform a fetch of length 'length' offset 'when' for all this data.
		// Then we apply the Holt-Winters model to each of the resulting series.
		newContext := *context
		newContext.Timerange = newContext.Timerange.Shift(when).SelectLength(length)

		original, err := function.EvaluateToSeriesList(arguments[0], &newContext)
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

type weighted struct {
	value  float64
	weight float64
	rate   float64
}

func (w *weighted) get() float64 {
	return w.value
}
func (w *weighted) observe(y float64) {
	if math.IsNaN(y) {
		w.skip()
		return
	}
	w.weight *= 1 - w.rate
	w.value = (w.value*w.weight + y*w.rate) / (w.weight + w.rate)
	w.weight += w.rate
}
func (w *weighted) boostAdd(dy float64) {
	if math.IsNaN(dy) {
		return
	}
	w.value += dy
}
func (w *weighted) skip() {
	w.weight *= 1 - w.rate
}
func newWeighted(rate float64) *weighted {
	return &weighted{
		value:  0,
		weight: 0,
		rate:   rate,
	}
}

type cycle struct {
	season []*weighted
}

func (c cycle) index(i int) int {
	return (i%len(c.season) + len(c.season)) % len(c.season)
}
func (c cycle) get(i int) float64 {
	return c.season[c.index(i)].get()
}
func (c cycle) observe(i int, v float64) {
	c.season[c.index(i)].observe(v)
}
func (c cycle) skip(i int) {
	c.season[c.index(i)].skip()
}
func newCycle(rate float64, n int) cycle {
	c := make([]*weighted, n)
	for i := range c {
		c[i] = newWeighted(rate)
	}
	return cycle{
		season: c,
	}
}

var SmoothHoltWinters = function.MetricFunction{
	Name:         "forecast.smooth_holt_winters",
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
			return nil, fmt.Errorf("forecast.holt_winters_adaptive expects the period parameter to mean at least one slot") // TODO: use a structured error
		}

		seriesList, err := function.EvaluateToSeriesList(arguments[0], context)
		if err != nil {
			return nil, err
		}

		result := api.SeriesList{
			Series:    make([]api.Timeseries, len(seriesList.Series)),
			Timerange: context.Timerange,
			Name:      seriesList.Name,
			Query:     fmt.Sprintf("forecast.holt_winters_adaptive(%s, %s, %f, %f)", seriesList.Query, period.String(), seasonalLearningRate, trendLearningRate),
		}

		for seriesIndex := range result.Series {
			// This will be the result.
			estimate := make([]float64, len(seriesList.Series[seriesIndex].Values))

			level := newWeighted(levelLearningRate)
			trend := newWeighted(trendLearningRate)
			season := newCycle(seasonalLearningRate, samples)

			for i, y := range seriesList.Series[seriesIndex].Values {
				// Remember the old values.
				oldLevel := level.get()
				oldTrend := trend.get()
				oldSeason := season.get(i)

				// Update the level, by increasing it by the estimate slope
				level.boostAdd(oldTrend)
				// Then observing the new y [if y is NaN, this skips, as desired]
				level.observe(y / oldSeason) // observe the y/s non-seasonal value

				// Next, observe the trend- difference between this level and last.
				// If y is NaN, we want to skip instead of updating.
				if math.IsNaN(y) {
					trend.skip()
				} else {
					// Compare the new level against the old.
					trend.observe(level.get() - oldLevel)
				}

				// Lastly, the seasonal value is just y / (l+b) the non-seasonal component.
				// If y is NaN, this will be NaN too, causing it to skip (as desired).
				season.observe(i, y/(oldLevel+oldTrend))

				// Our estimate is the level times the seasonal component.
				estimate[i] = level.get() * season.get(i)
			}

			result.Series[seriesIndex] = api.Timeseries{
				TagSet: seriesList.Series[seriesIndex].TagSet,
				Raw:    seriesList.Series[seriesIndex].Raw,
				Values: estimate,
			}
		}

		return result, nil
	},
}
