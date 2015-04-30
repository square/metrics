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
	"reflect"
	"runtime"
	"testing"
)

// Assert is a helper struct for testing methods.
type Assert struct {
	t     *testing.T
	stack int // number of stack frames to traverse to generate error.
}

func caller(depth int) (string, int) {
	// determines how many stack frames to traverse.
	// we need to traverse 3 for the original caller:
	// 0: caller()
	// 1: Assert.withCaller()
	// 2: Assert.Eq...()
	// 3: <- original caller
	_, file, line, _ := runtime.Caller(depth + 3)
	return file, line
}

// New creates a new Assert struct.
func New(t *testing.T) Assert {
	return Assert{t, 0}
}

func NewWithStack(t *testing.T, stack int) Assert {
	return Assert{t, stack}
}

func (assert Assert) withCaller(format string, a ...interface{}) {
	file, line := caller(assert.stack)
	assert.t.Errorf(fmt.Sprintf("%s:%d>", file, line)+format, a...)
}

// EqStringSlices checks whether given two string slices are equal.
func (assert Assert) EqStringSlices(actual []string, expected []string) {
	if !reflect.DeepEqual(actual, expected) {
		assert.withCaller("Expected \"%s\", but got \"%s\"", actual, expected)
	}
}

// EqString fails the test if two strings aren't equal.
func (assert Assert) EqString(actual, expected string) {
	if actual != expected {
		assert.withCaller("Expected=[%s], actual=[%s]", expected, actual)
	}
}

// EqInt fails the test if two ints aren't equal.
func (assert Assert) EqInt(actual, expected int) {
	if actual != expected {
		assert.withCaller("Expected=[%d], actual=[%d]", expected, actual)
	}
}

// Eq fails the test if two arguments are not equal.
func (assert Assert) Eq(actual, expected interface{}) {
	if !reflect.DeepEqual(actual, expected) {
		assert.withCaller("Expected=%s, actual=%s", expected, actual)
	}
}

// CheckError fails the test if a non-nil error is passed.
func (assert Assert) CheckError(err error) {
	if err != nil {
		assert.withCaller("Unexpected error: %s", err.Error())
	}
}
