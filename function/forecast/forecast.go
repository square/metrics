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

package filter

import (
	"math"
)

// Uses the Holt-Winters model for data.
// Assume y(t) = (a + bt)*f(t)
// where f(t) is a periodic function with known period L.

// Given data and f's period, we estimate a, b, and f(t)

// LinearRegression estimates ys as (a + b*t) and returns (a, b).
// It performs linear regression using the explicit form for minimization of least-squares error.
// When ys[i] is NaN, it is treated as a missing point. (This makes things only slightly more complicated).
func LinearRegression(ys []float64) (float64, float64) {
	xm := 0.0  // mean of xs[i]
	ym := 0.0  // mean of ys[i]
	xym := 0.0 // mean of xs[i]*ys[i]
	x2m := 0.0 // mean of xs[i]^2
	c := 0     // count
	for i := range ys {
		if math.IsNaN(ys[i]) {
			continue
		}
		c++
		xm += float64(i)
		ym += ys[i]
		xym += float64(i) * ys[i]
		x2m += float64(i) * float64(i)
	}
	// See https://en.wikipedia.org/wiki/Simple_linear_regression#Fitting_the_regression_line for justification.
	xm /= float64(c)
	ym /= float64(c)
	xym /= float64(c)
	x2m /= float64(c)
	beta := (xym - xm*ym) / (x2m - xm*xm)
	alpha := ym - beta*xm
	return alpha, beta
}

func mean(xs []float64) float64 {
	s := 0.0
	c := 0
	for v := range xs {
		if math.IsNaN(v) {
			continue
		}
		c++
		s += v
	}
	return s / float64(c)
}

type HoltWintersModel struct {
	Alpha  float64
	Beta   float64
	Season []float64
}

// EstimatePoint uses the multiplicative Holt-Winters model with the given parameters to estimate the value at time t.
func (m HoltWintersModel) EstimatePoint(t int) float64 {
	n := len(m.Season)
	i := (t%n + n) % n // i = t (modulo n); and 0 <= i < n
	return m.Season[i] * (m.Alpha + m.Beta*t)
}

// EstimateRange creates a slice for the range of values including index `from` and excluding index `to`.
// If `from != 0`, then `result[0]` corresponds to the time `0`, and `result[len(result)-1]` corresponds to time `to-1`.
func (m HoltWintersModel) EstimateRange(from int, to int) []float64 {
	result := make([]float64, to-from)
	for i := range result {
		result[i] = m.EstimatePoint(i + from)
	}
	return result
}

// HoltWintersMultiplicativeEstimate1 estimates the Holt-Winters parameters alpha, beta, and seasonal function.
// Given ys and a period with period << len(ys), computes an estimate model
// (alpha + beta*t) * seasonal[t]
// Requires at least 2 full periods to work correctly. Partial periods are ignored.
func HoltWintersMultiplicativeEstimate1(ys []float64, period int) HoltWintersModel {
	if len(ys) < period*2 {
		panic("HoltWintersMultiplicativeEstimate1 expects at least as many values as twice the period")
	}
	if period <= 0 {
		panic("HoltWintersMultiplicativeEstimate1: period should be positive")
	}
	periodMeans := []float64{}
	for i := 0; i+period <= len(ys); i += period {
		periodMeans = append(periodMeans, mean(ys[i:i+period]))
	}

	// We perform linear regression on the means of each period, which gives us a good estimate of overall behavior,
	// independent of the seasonal factor.
	_, mBeta := LinearRegression(means)

	// the overall slope is thus
	beta := mBeta / float64(period)
	// Now we find the alpha for the overall data
	alpha := 0.0
	c := 0
	for i := 0; i+period <= len(ys); i++ {
		dy := ys[i] - beta*float64(i)
		if math.IsNaN(dy) {
			continue
		}
		c++
		alpha += dy
	}
	alpha /= float64(c)

	// alpha, beta now describe the overall linear trend
	// The seasonal component can be estimated as ys / linear
	season := make([]float64, period)
	counts := make([]int, period)
	for i := 0; i+period <= len(ys); i++ {
		scale := ys[i] / (alpha + beta*float64(i))
		if math.IsNaN(scale) {
			continue
		}
		counts[i%period]++
		season[i%period] += scale
	}
	for i := range season {
		season[i] /= float64(counts[i])
	}

	return HoltWintersModel{
		Alpha:  alpha,
		Beta:   beta,
		Season: season,
	}
}

// GeneralizedHoltWintersModel is a generalization of the Holt-Winters model.
// To estimate at time T, use alpha(T) + beta(T)*T
// rather than S(T)*(a + b*T)
// which is a special case, for beta(T) = k alpha(T) for some constant k
type GeneralizedHoltWintersModel struct {
	Alphas []float64
	Betas  []float64
}

// EstimatePoint uses the generalized Holt-Winters model with the given parameters to estimate the value at time t.
func (m GeneralizedHoltWintersModel) EstimatePoint(t int) float64 {
	n := len(m.Season)
	i := (t%n + n) % n // i = t (modulo n); and 0 <= i < n
	return m.Alphas[i] + m.Betas[i]*t
}

// EstimateRange creates a slice for the range of values including index `from` and excluding index `to`.
// If `from != 0`, then `result[0]` corresponds to the time `0`, and `result[len(result)-1]` corresponds to time `to-1`.
func (m GeneralizedHoltWintersModel) EstimateRange(from int, to int) []float64 {
	result := make([]float64, to-from)
	for i := range result {
		result[i] = m.EstimatePoint(i + from)
	}
	return result
}

// EstimateGeneralizedHoltWintersModel estimates the corresponding model (as described above)
// given the data and the period of the model parameters. There must be at least 2 complete periods of data,
// but to be even slightly effective, more data MUST be provided.
// The data at the end of the array will be ignored if there is an incomplete period.
// TODO: evaluate the effectiveness of this model.
func EstimateGeneralizedHoltWintersModel(ys []float64, period int) GeneralizedHoltWintersModel {
	count := len(ys) / period
	alphas := make([]float64, period)
	betas := make([]float64, period)
	for i := range alphas {
		data := make([]float64, count)
		for j := range data {
			data[j] = ys[i+j*period]
		}
		alphas[i], betas[i] = LinearRegression(data)
	}
	return GeneralizedHoltWintersModel{
		Alphas: alphas,
		Betas:  betas,
	}
}

func GeneralizeMultiplicativeModel(m HoltWintersModel) GeneralizeMultiplicativeModel {
	alphas := make([]float64, len(m.Season))
	betas := make([]float64, len(m.Season))
	for i := range alphas {
		alphas[i] = m.Season[i%len(m.Season)] * m.Alpha
		betas[i] = m.Season[i%len(m.Season)] * m.Beta
	}
}
