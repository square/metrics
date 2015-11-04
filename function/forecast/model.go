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

// Model represents an abstract model for estimating data.
// The forecast package describes several models based on the Holt-Winters model.
type Model interface {
	EstimatePoint(i int) float64
	EstimateRange(start int, length int) []float64
}

// Forecast uses variants of the Holt-Winters model for data.
// Multiplicative model: Assume y(t) = (a + bt)*f(t)  where f(t) is a periodic function with known period L and mean 1.
// Additive model:       Assume y(t) = a + bt + f(t)  where f(t) is a periodic function with known period L and mean 0.
// "Generalized" model:  Assume y(t) = a(t) + b(t)x where a(t) and b(t) are periodic functions with known period L.

// Returns the unique integer r such that x == r (mod m) and 0 <= r < m
func mod(x int, m int) int {
	return ((x % m) + m) % m
}

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
// The input slice is not modified.
func EstimateGeneralizedHoltWintersModel(ys []float64, period int) (Model, error) {
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

// MultiplicativeHoltWintersModel represents a model of data covered by the Multiplicative Holt-Winters model,
// f(t) = S(t) (a + bt), where S(t) is periodic with known period and has mean 1.
type MultiplicativeHoltWintersModel struct {
	season []float64
	alpha  float64
	beta   float64
}

// EstimatePoint provides an estimate for a particular point in time from the given model.
func (m MultiplicativeHoltWintersModel) EstimatePoint(t int) float64 {
	period := len(m.season)
	index := mod(t, period)
	return m.season[index] * (m.alpha + m.beta*float64(t))
}

// EstimateRange provides an estimate for each point in time for the given range specified as a start and a length.
func (m MultiplicativeHoltWintersModel) EstimateRange(start int, length int) []float64 {
	result := make([]float64, length)
	for i := range result {
		result[i] = m.EstimatePoint(i + start)
	}
	return result
}

// Trains a multiplicative Holt-Winters model on the given data (using the given period).
// The input slice is not modified.
func TrainMultiplicativeHoltWintersModel(ys []float64, period int) (Model, error) {
	if period <= 0 {
		return MultiplicativeHoltWintersModel{}, fmt.Errorf("Training the multiplicative Holt-Winters model requires a positive period.") // TODO: structured error
	}
	if len(ys) < period*3 {
		return MultiplicativeHoltWintersModel{}, fmt.Errorf("Good results with the Multiplicative Holt-Winters model training require at least 3 periods of data.") // TODO: structured error
	}
	// First we will find the "beta" parameter (the average trend).
	// To do this, we require
	periodMeans := make([]float64, len(ys)/period)
	for i := range periodMeans {
		sum := 0.0
		for t := 0; t < period; t++ {
			sum += ys[i*period+t]
		}
		periodMeans[i] = sum / float64(period)
	}
	// periodMeans now contains the mean of each period of the data.
	_, beta := LinearRegression(periodMeans)
	// This beta is the overall trend of the data, but it needs to be rescaled:
	beta /= float64(period)

	// Next we calculate the "untrended" data, by subtracting beta*t from each point:

	zs := make([]float64, len(ys))
	for i := range ys {
		zs[i] = ys[i] - beta*float64(i)
	}

	// Now we make the following observation. Consider g(t) = f(t) - bt. Then we have
	// g(t) = f(t) - bt = S(t)(a + bt) - bt = a S(t) + b S(t) t - b t = a S(t) + b (S(t) - 1) t
	// So now compute g(np + t) - g(mp + t), where 0 <= t < p, and n, m are two integer. So,
	// g(np+t) - g(mp+t) = aS(np+t) + b(S(np+t)-1)(np+t) - aS(mp+t) + b(S(mp+t)-1)(mp+t).
	// But S is a periodic function, so we can simplify:
	// g(np+t) - g(mp+t) = aS(t) + b(S(t)-1)(np+t) - aS(t) + b(S(t)-1)(mp+t)
	// and a bit more expansion and factoring gives us
	// g(np+t) - g(mp+t) = b(S(t) - 1)(np+t - mp - t) = bp(S(t)-1)(n-m).
	// Thus, by solving for S(t), we can see that
	// (g(np+t) - g(mp+t)) / (bp (n-m)) = S(t) - 1, so
	// S(t) = 1 + (g(np+t) - g(mp+t)) / (bp (n-m))

	// However, this gives us n^2 equations where we have n periods of data. Therefore, we'll use the average across all of these.

	season := make([]float64, period)

	for t := 0; t < period; t++ {
		gs := make([]float64, len(ys)/period)
		for n := range gs {
			gs[n] = zs[n*period+t] // For convenience
		}
		sumS := 0.0
		countS := 0
		for n := range gs {
			for m := range gs {
				if n == m {
					continue
				}
				value := 1 + (gs[n]-gs[m])/beta/float64(period)/float64(n-m)
				sumS += value
				countS++
			}
		}
		season[t] = sumS / float64(countS)
	}

	//log.Printf("Calculated season: %+v", season)

	// Lastly, we'll need to compute 'alpha'. We do this be "deseasonalizing" zs.

	// g(t) = a S(t) + b (S(t) - 1) t
	// So we have to subtract out b(S(t)-1)t
	// and then divide by S(t):
	// a = (g(t) - b(S(t)-1)t) / S(t)

	ds := make([]float64, len(zs))
	for i := range zs {
		s := season[mod(i, period)]
		ds[i] = (zs[i] - beta*(s-1)*float64(i)) / s
	}

	alpha := 0.0
	for i := range ds {
		alpha += ds[i]
	}
	alpha /= float64(len(ds))

	//log.Printf("Calculated alpha: %f", alpha)

	return MultiplicativeHoltWintersModel{
		season: season,
		alpha:  alpha,
		beta:   beta,
	}, nil
}
