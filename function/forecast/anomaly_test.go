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

import "testing"
import "github.com/square/metrics/testing_support/assert"

func TestStandardDeviationsFromExpected(t *testing.T) {
	tests := []struct {
		correct  []float64
		estimate []float64
		expected []float64
	}{
		{
			correct:  []float64{3, 3, -3, -3, 3, -3},
			estimate: []float64{0, 0, 0, 0, 0, 0},
			expected: []float64{-0.913, -0.913, 0.913, 0.913, -0.913, 0.913},
		},
		{
			correct:  []float64{25, 25, 25, 25},
			estimate: []float64{30.1, 30, 29.9, 30},
			expected: []float64{1.22, 0, -1.22, 0},
		},
		{
			correct:  []float64{1, 2, 3, 4, 10, 6, 7, 8},
			estimate: []float64{1.01, 1.99, 3.02, 3.99, 4.05, 5.97, 7.03, 8.1},
			expected: []float64{0.35, 0.34, 0.35, 0.34, -2.47, 0.33, 0.36, 0.39},
		},
	}
	for _, test := range tests {
		a := assert.New(t).Contextf("Standardized Deviations from Expected\nGround truth: %+v\nGiven estimate: %+v\n", test.correct, test.estimate)
		actual, err := standardDeviationsFromExpected(test.correct, test.estimate)
		a.CheckError(err)
		a.EqFloatArray(actual, test.expected, 0.01)
	}
}

func TestPeriodicStandardDeviationsFromExpected(t *testing.T) {
	tests := []struct {
		period   int
		correct  []float64
		estimate []float64
		expected []float64
	}{
		{
			period:   3,
			correct:  []float64{3, 3, -3, -3, 3, -3},
			estimate: []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6},

			expected: []float64{-0.707, -0.707, -0.707, 0.707, 0.707, 0.707},
		},
		{
			period:   3,
			correct:  []float64{25, 25, 25, 25, 27, 27, 27, 27},
			estimate: []float64{30.1, 30, 29.9, 30, 30.1, 30.3, 29, 31},
			expected: []float64{0.606, 1.017, 0.707, 0.549, -0.982, -0.707, -1.154, -0.035},
		},
		{
			period:   3,
			correct:  []float64{1, 2, 3, 4, 10, 6, 7, 8},
			estimate: []float64{1.01, 1.99, 3.02, 3.99, 4.05, 5.97, 7.03, 8.1},
			expected: []float64{0.000, 0.561, 0.707, -1.000, -1.155, -0.707, 1.000, 0.593},
		},
	}
	for _, test := range tests {
		a := assert.New(t).Contextf("Standardized Deviations from Expected\nPeriod: %d\nGround truth: %+v\nGiven estimate: %+v\n", test.period, test.correct, test.estimate)
		actual, err := periodicStandardDeviationsFromExpected(test.correct, test.estimate, test.period)
		a.CheckError(err)
		a.EqFloatArray(actual, test.expected, 0.01)
	}
}
