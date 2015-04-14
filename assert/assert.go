// Package assert contains wrapper on top of go's testing library
// to make tests easier.
package assert

import (
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

// EqStringSlices checks whether given two string slices are equal.
func (assert Assert) EqStringSlices(actual []string, expected []string) {
	if !reflect.DeepEqual(actual, expected) {
		file, line := caller()
		assert.t.Errorf("%s:%d> Expected \"%s\", but got \"%s\"", file, line, actual, expected)
	}
}

// EqString fails the test if two strings aren't equal.
func (assert Assert) EqString(actual, expected string) {
	if actual != expected {
		file, line := caller()
		assert.t.Errorf("%s:%d>Expected=[%s], actual=[%s]", file, line, expected, actual)
	}
}

// EqInt fails the test if two ints aren't equal.
func (assert Assert) EqInt(actual, expected int) {
	if actual != expected {
		file, line := caller()
		assert.t.Errorf("%s:%d>Expected=[%d], actual=[%d]", file, line, expected, actual)
	}
}

// Eq fails the test if two arguments are not equal.
func (assert Assert) Eq(actual, expected interface{}) {
	if !reflect.DeepEqual(actual, expected) {
		file, line := caller()
		assert.t.Errorf("%s:%d>Expected=%s, actual=%s", file, line, expected, actual)
	}
}

// CheckError fails the test if a non-nil error is passed.
func (assert Assert) CheckError(err error) {
	if err != nil {
		file, line := caller()
		assert.t.Errorf("%s:%d>Unexpected error: %s", file, line, err.Error())
	}
}
