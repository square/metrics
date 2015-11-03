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

var modelHoltWinters = function.MetricFunction{
	Name:         "forecast.model_generalized_holt_winters",
	MinArguments: 4, // Series, period, start time of training, end time of training
	MaxArguments: 4,
	Compute: func(context *function.EvaluationContext, arguments []function.Expression, groups function.Groups) (function.Value, error) {
		period, err := function.EvaluateToDuration(arguments[1], context)
		if err != nil {
			return nil, err
		}
		periodSamples := int(period / context.Timerange.Resolution())
		if periodSamples <= 0 {
			return nil, fmt.Errorf("forecast.model_generalized_holt_winters expected the period to exceed the resolution") // TODO: use structured error
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
			return nil, fmt.Errorf("forecast.model_generalized_holt_winters expected the end time to come after the start time") // TODO: use a structured error
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
			Query:     fmt.Sprintf("forecast.model_generalized_holt_winters(%s, %s, %s, %s)", trainingSeries.Query, period.String(), start.String(), end.String()),
			Timerange: context.Timerange,
			Series:    make([]api.Timeseries, len(trainingSeries.Series)),
		}

		// How far in the future the fetch time is than the training time.
		timeOffset := int(-start / context.Timerange.Resolution())

		for i := range result.Series {
			trainingData := trainingSeries.Series[i]
			model, err := EstimateGeneralizedHoltWintersModel(trainingData.Values, periodSamples)
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
			return nil, fmt.Errorf("forecast.smooth_holt_winters expects the period parameter to mean at least one slot") // TODO: use a structured error
		}

		seriesList, err := function.EvaluateToSeriesList(arguments[0], context)
		if err != nil {
			return nil, err
		}

		result := api.SeriesList{
			Series:    make([]api.Timeseries, len(seriesList.Series)),
			Timerange: context.Timerange,
			Name:      seriesList.Name,
			Query:     fmt.Sprintf("forecast.smooth_holt_winters(%s, %s, %f, %f)", seriesList.Query, period.String(), seasonalLearningRate, trendLearningRate),
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
