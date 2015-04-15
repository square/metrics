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
	t *testing.T
}

func caller() (string, int) {
	_, file, line, _ := runtime.Caller(2)
	return file, line
}

// New creates a new Assert struct.
func New(t *testing.T) Assert {
	return Assert{t}
}

func (assert Assert) withCaller(format string, a ...interface{}) {
	file, line := caller()
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
