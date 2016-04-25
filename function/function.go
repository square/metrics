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

package function

import "fmt"

// The Function interface defines a metric function.
// It is given several (unevaluated) expressions as input, and evaluates to a Value.
type Function interface {
	Run(EvaluationContext, []Expression, Groups) (Value, error)
	Name() string
}

// The Registry interface defines a mapping from names to Functions
// and provides a way to get the full list of functions defined.
type Registry interface {
	GetFunction(string) (Function, bool) // returns an instance of a Function
	All() []string                       // all the registered functions
}

// Groups holds grouping information - which tags to group by (if any), and whether to `collapse` (Collapses = true) or `group` (Collapses = false)
type Groups struct {
	List      []string // the tags to group by
	Collapses bool     // whether to "collapse by" instead of "group by"
}

// MetricFunction holds a generic function object with information about its parameters.
type MetricFunction struct {
	FunctionName  string // Name is the name of the function, used in its registration.
	MinArguments  int    // MinArguments is the minimum number of arguments the function allows.
	MaxArguments  int    // MaxArguments is the maximum number of arguments the function allows. -1 indicates an unlimited number.
	AllowsGroupBy bool   // Whether the function allows a 'group by' clause.
	Compute       func(EvaluationContext, []Expression, Groups) (Value, error)
}

// Name returns the MetricFunction's name.
func (metricFunc MetricFunction) Name() string {
	return metricFunc.FunctionName
}

// Run evaluates the given MetricFunction on its arguments.
// It performs error-checking against the supplies number of arguments and/or group-by clause.
func (f MetricFunction) Run(context EvaluationContext, arguments []Expression, groups Groups) (Value, error) {
	// preprocessing
	length := len(arguments)
	if length < f.MinArguments || (f.MaxArguments != -1 && f.MaxArguments < length) {
		return nil, ArgumentLengthError{f.FunctionName, f.MinArguments, f.MaxArguments, length}
	}
	if len(groups.List) > 0 && !f.AllowsGroupBy {
		// TODO(jee) - use typed errors
		return nil, fmt.Errorf("function %s doesn't allow a group-by clause", f.FunctionName)
	}
	return f.Compute(context, arguments, groups)
}
