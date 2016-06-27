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

import "math"

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
	if w.weight == 0 || math.IsNaN(w.value) || math.IsInf(w.value, 0) { // Special case to prevent 'NaN'
		w.value = y
		w.weight = w.rate
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

// RollingMultiplicativeHoltWinters approximate the given input using the Holt-Winters model by performing exponential averaging on the HW parameters.
// It scales 'levelLearningRate' and 'trendLearningRate' by the 'period'.
// That is, if you double the period, it will take twice as long as before for the level and trend parameters to update.
// This makes it easier to use with varying period values.
func RollingMultiplicativeHoltWinters(ys []float64, period int, levelLearningRate float64, trendLearningRate float64, seasonalLearningRate float64) []float64 {
	// We'll interpret the rates as "the effective change per whole period" (so the seasonal learning rate is unchanged).
	// The intensity of the old value after n iterations is (1-rate)^n. We want to find rate' such that
	// 1 - rate = (1 - rate')^n
	// so
	// 1 - (1 - rate)^(1/n) = rate'
	levelLearningRate = 1 - math.Pow(1-levelLearningRate, 1/float64(period))
	trendLearningRate = 1 - math.Pow(1-trendLearningRate, 1/float64(period))
	estimate := make([]float64, len(ys))

	level := newWeighted(levelLearningRate)
	trend := newWeighted(trendLearningRate)
	season := newCycle(seasonalLearningRate, period)

	// we need to initialize the season to '1':
	for i := 0; i < period; i++ {
		season.observe(i, 1)
	}

	for i, y := range ys {
		// Remember the old values.
		oldLevel := level.get()
		oldTrend := trend.get()
		oldSeason := season.get(i)

		// Update the level, by increasing it by the estimate slope
		level.boostAdd(oldTrend)
		// Then observing the new y [if y is NaN, this skips, as desired]
		level.observe(y / oldSeason) // observe the y's non-seasonal value

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
	return estimate
}

// RollingSeasonal estimates purely seasonal data without a trend or level component.
// For data which shows no long- or short-term trends, this model is more likely to recognize
// deviant behavior. However, it will perform worse than Holt-Winters on data which does
// have any significant trends.
func RollingSeasonal(ys []float64, period int, seasonalLearningRate float64) []float64 {
	season := newCycle(seasonalLearningRate, period)
	estimate := make([]float64, len(ys))
	for i := range ys {
		season.observe(i, ys[i])
		estimate[i] = season.get(i)
	}
	return estimate
}

// ForecastLinear estimates a purely linear trend from the data.
func ForecastLinear(ys []float64) []float64 {
	estimate := make([]float64, len(ys))
	a, b := LinearRegression(ys)
	for i := range ys {
		estimate[i] = a + b*float64(i)
	}
	return estimate
}
