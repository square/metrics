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

func rollingMultiplicativeHoltWinters(ys []float64, period int, levelLearningRate float64, trendLearningRate float64, seasonalLearningRate float64) []float64 {
	estimate := make([]float64, len(ys))

	level := newWeighted(levelLearningRate)
	trend := newWeighted(trendLearningRate)
	season := newCycle(seasonalLearningRate, period)

	for i, y := range ys {
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
	return estimate
}
