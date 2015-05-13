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

package query

import (
	"testing"

	"github.com/square/metrics/api"
	"github.com/square/metrics/assert"
)

type FakeBackend struct{}

func (b FakeBackend) FetchMetadata(metric api.TaggedMetric) api.MetricMetadata {
	return api.MetricMetadata{}
}

func (b FakeBackend) FetchSeries(query api.Query) api.SeriesList {
	return api.SeriesList{}
}

func Test_ScalarExpression(t *testing.T) {
	for _, test := range []struct {
		expectSuccess  bool
		expr           scalarExpression
		timerange      api.Timerange
		expectedSeries []api.Timeseries
	}{
		{
			true,
			scalarExpression{5},
			api.Timerange{0, 10, 2},
			[]api.Timeseries{
				api.Timeseries{
					[]float64{5.0, 5.0, 5.0, 5.0, 5.0, 5.0},
					api.TaggedMetric{},
				},
			},
		},
		{
			false,
			scalarExpression{5},
			api.Timerange{0, 10, 3},
			[]api.Timeseries{},
		},
	} {
		a := assert.New(t).Contextf("%+v", test)

		result, err := test.expr.Evaluate(EvaluationContext{FakeBackend{}, test.timerange})

		a.EqBool(err == nil, test.expectSuccess)
		// Nothing else to validate if we expect failure
		if !test.expectSuccess {
			continue
		}

		a.EqInt(len(result.Series), len(test.expectedSeries))

		for i := 0; i < len(result.Series); i += 1 {
			a.Eq(result.Series[i].Values, test.expectedSeries[i].Values)
		}
	}
}

func Test_evaluateBinaryOperation(t *testing.T) {
	for _, test := range []struct {
		expectSuccess bool
		context       EvaluationContext
		functionName  string
		operands      []Expression
		evalFunction  func(float64, float64) float64
	}{} {
		// XXX Add tests
		_ = test
	}
}

var _ api.Backend = (*FakeBackend)(nil)
