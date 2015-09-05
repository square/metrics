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

// Package assert contains wrapper on top of go's testing library
// to make tests easier.
package assert

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"runtime"
	"testing"
)

// used to strip out the long filename.
var fileRegex = regexp.MustCompile("([^/]*/){0,2}[^/]*$")

// Assert is a helper struct for testing methods.
type Assert struct {
	t       *testing.T
	stack   int // number of stack frames to traverse to generate error.
	context string
}

// New creates a new Assert struct.
func New(t *testing.T) Assert {
	return Assert{t, 0, ""}
}

// Stack shifts how many stack frames to traverse to print the error message.
// this may be useful if you're creating a helper testing method.
// returns a new instances of Assert.
func (assert Assert) Stack(stack int) Assert {
	assert.stack += stack
	return assert
}

// Contextf sets the human-readable context of the test. This is useful
// when the line number is not sufficient locator for the test failure:
// i.e. testing in a loop.
// returns a new instances of Assert.
func (assert Assert) Contextf(format string, a ...interface{}) Assert {
	assert.context = fmt.Sprintf(format, a...)
	return assert
}

// Errorf marks the test as failed.
func (assert Assert) Errorf(format string, a ...interface{}) {
	assert.withCaller(format, a...)
}

// EqString fails the test if two strings aren't equal.
func (assert Assert) EqString(actual, expected string) {
	if actual != expected {
		assert.withCaller("Expected=[%s], actual=[%s]", expected, actual)
	}
}

// EqBool fails the test if two booleans aren't equal.
func (assert Assert) EqBool(actual, expected bool) {
	if actual != expected {
		assert.withCaller("Expected=[%t], actual=[%t]", expected, actual)
	}
}

// EqInt fails the test if two ints aren't equal.
func (assert Assert) EqInt(actual, expected int) {
	if actual != expected {
		assert.withCaller("Expected=[%d], actual=[%d]", expected, actual)
	}
}

func (assert Assert) EqFloatArray(actual, expected []float64, epsilon float64) {
	if len(actual) != len(expected) {
		assert.withCaller("Expected=%+v, actual=%+v", expected, actual)
		return
	}
	for i := range actual {
		if math.IsNaN(expected[i]) {
			if !math.IsNaN(actual[i]) {
				assert.withCaller("Expected=%+v, actual=%+v", expected, actual)
				return
			}
		} else {
			delta := actual[i] - expected[i]
			if math.IsNaN(delta) || math.Abs(delta) > epsilon {
				assert.withCaller("Expected=%+v, actual=%+v", expected, actual)
				return
			}
		}
	}
}

// EqFloat fails the test if two floats aren't equal. NaNs are considered equal.
func (assert Assert) EqFloat(actual, expected, epsilon float64) {
	delta := math.Abs(actual - expected)
	if (delta > epsilon && actual != expected) && !(math.IsNaN(actual) && math.IsNaN(expected)) {
		assert.withCaller("Expected=[%f], actual=[%f]", expected, actual)
	}
}

// EqFloat fails the test if two floats aren't equal. NaNs are considered equal.
func (assert Assert) EqApproximate(actual, expected, epsilon float64) {
	delta := actual - expected
	if !(-epsilon < delta && delta < epsilon) {
		assert.withCaller("Expected=[%f], actual=[%f]", expected, actual)
	}
}

// Eq fails the test if two arguments are not equal.
func (assert Assert) Eq(actual, expected interface{}) {
	if !reflect.DeepEqual(actual, expected) {
		assert.withCaller("\nExpected=%+v\nActual  =%+v", expected, actual)
	}
}

// CheckError fails the test if a non-nil error is passed.
func (assert Assert) CheckError(err error) {
	if err != nil {
		assert.withCaller("Unexpected error: %s", err.Error())
	}
}

// Utility Functions
// =================

func (assert Assert) withCaller(format string, a ...interface{}) {
	file, line := caller(assert.stack)
	if assert.context != "" {
		assert.t.Errorf("%s:%d> [%s] %s", file, line, assert.context, fmt.Sprintf(format, a...))
	} else {
		assert.t.Errorf("%s:%d>%s", file, line, fmt.Sprintf(format, a...))
	}
}

func caller(depth int) (string, int) {
	// determines how many stack frames to traverse.
	// we need to traverse 3 for the original caller:
	// 0: caller()
	// 1: Assert.withCaller()
	// 2: Assert.Eq...()
	// 3: <- original caller
	_, file, line, _ := runtime.Caller(depth + 3)
	match := fileRegex.FindString(file)
	if match != "" {
		return match, line
	}
	return file, line
}
