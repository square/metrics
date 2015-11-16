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

// Model represents an abstract model for estimating data.
// The forecast package describes several models based on the Holt-Winters model.
type Model interface {
	EstimatePoint(i int) float64
	EstimateRange(start int, length int) []float64
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
