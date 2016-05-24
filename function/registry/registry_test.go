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

package registry

import (
	"errors"
	"testing"

	"github.com/square/metrics/function"
	"github.com/square/metrics/testing_support/assert"
)

var dummyCompute = func(function.EvaluationContext, []function.Expression, function.Groups) (function.Value, error) {
	return nil, errors.New("Not implemented")
}

func Test_Registry_Default(t *testing.T) {
	a := assert.New(t)
	sr := StandardRegistry{mapping: make(map[string]function.Function)}
	a.Eq(sr.All(), []string{})
	if err := sr.Register(function.MetricFunction{FunctionName: "foo", Compute: dummyCompute}); err != nil {
		a.CheckError(err)
	}
	if err := sr.Register(function.MetricFunction{FunctionName: "bar", Compute: dummyCompute}); err != nil {
		a.CheckError(err)
	}
	a.Eq(sr.All(), []string{"bar", "foo"})
}

func Test_Registry_Error(t *testing.T) {
	for _, suite := range []struct {
		Name     string
		Function function.MetricFunction
	}{
		{"empty name", function.MetricFunction{FunctionName: "", Compute: dummyCompute}},
		{"duplicate name", function.MetricFunction{FunctionName: "existing", Compute: dummyCompute}},
	} {
		a := assert.New(t).Contextf("%s", suite.Name)
		// set up the standard registry
		sr := StandardRegistry{mapping: make(map[string]function.Function)}
		if err := sr.Register(function.MetricFunction{FunctionName: "existing", Compute: dummyCompute}); err != nil {
			a.CheckError(err)
			return
		}
		if err := sr.Register(suite.Function); err == nil {
			a.Errorf("Expected error, but got none.")
		}
	}
}
