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
	"github.com/square/metrics/api"
)

// EvaluationContext is the central piece of logic, providing
// helper funcions & varaibles to evaluate a given piece of
// metrics query.
// * Contains Backend object, which can be used to fetch data
// from the backend system.s
// * Contains current timerange being queried for - this can be
// changed by say, application of time shift function.
type EvaluationContext struct {
	Backend   api.Backend   // backend to fetch data from.
	Timerange api.Timerange // current time range to fetch data from.
}

// Expression is a piece of code, which can be evaluated by a given EvaluationContext.
// Internally, expressions form a tree of subexpressions, delegating work between them.
type Expression interface {
	// Evaluate the given expression.
	Evaluate(context EvaluationContext) api.SeriesResult
}

// Implementations
// ===============
func (expr *numberExpression) Evaluate(context EvaluationContext) api.SeriesResult {
	return nil // TODO - implement this.
}

func (expr *metricFetchExpression) Evaluate(context EvaluationContext) api.SeriesResult {
	return nil // TODO - implement this.
}

func (expr *functionExpression) Evaluate(context EvaluationContext) api.SeriesResult {
	return nil // TODO - implement this.
}
