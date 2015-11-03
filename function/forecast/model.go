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
	xm /= float64(c)
	ym /= float64(c)
	xym /= float64(c)
	x2m /= float64(c)
	// See https://en.wikipedia.org/wiki/Simple_linear_regression#Fitting_the_regression_line for justification.
	beta := (xym - xm*ym) / (x2m - xm*xm)
	alpha := ym - beta*xm
	return alpha, beta
}

// GeneralizedHoltWintersModel is a generalization of the Holt-Winters model.
// To estimate at time t, where t = kp + r, with 0 <= r < p: alpha[r] + beta[r] * k
// The Holt-Winters model is a special case, where Alphas[i] = k*Betas[i], for some constant k the same for all i.
type GeneralizedHoltWintersModel struct {
	Alphas []float64
	Betas  []float64
}

// EstimatePoint uses the generalized Holt-Winters model with the given parameters to estimate the value at time t.
func (m GeneralizedHoltWintersModel) EstimatePoint(t int) float64 {
	period := len(m.Alphas)
	n := len(m.Alphas)
	i := (t%n + n) % n // i = t (modulo n); and 0 <= i < n
	return m.Alphas[i] + m.Betas[i]*(float64(t)-float64(i))/float64(period)
}

// EstimateRange creates a slice for the range of values including index `start` having `length` values.
func (m GeneralizedHoltWintersModel) EstimateRange(start int, length int) []float64 {
	result := make([]float64, length)
	for i := range result {
		result[i] = m.EstimatePoint(i + start)
	}
	return result
}

// EstimateGeneralizedHoltWintersModel estimates the corresponding model (as described above)
// given the data and the period of the model parameters. There must be at least 2 complete periods of data,
// but to be even slightly effective, more data MUST be provided.
// The data at the end of the array will be ignored if there is an incomplete period.
func EstimateGeneralizedHoltWintersModel(ys []float64, period int) (GeneralizedHoltWintersModel, error) {
	if period <= 0 {
		return GeneralizedHoltWintersModel{}, fmt.Errorf("Generalized Holt-Winters model expects a positive period")
	}
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
	}, nil
}
