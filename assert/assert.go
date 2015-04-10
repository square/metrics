// Package assert contains wrapper on top of go's testing library
// to make tests easier.
package assert

import (
	"runtime"
	"testing"
  "reflect"
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

// EqString fails the test if two strings aren't equal.
func (assert Assert) EqString(actual, expected string) {
	file, line := caller()
	if actual != expected {
		assert.t.Errorf("%s:%d>Expected=[%s], actual=[%s]", file, line, expected, actual)
	}
}

// EqInt fails the test if two ints aren't equal.
func (assert Assert) EqInt(actual, expected int) {
	file, line := caller()
	if actual != expected {
		assert.t.Errorf("%s:%d>Expected=[%d], actual=[%d]", file, line, expected, actual)
	}
}

// Eq fails the test if two arguments are not equal.
func (assert Assert) Eq(actual, expected interface{}) {
	file, line := caller()
	if !reflect.DeepEqual(actual, expected) {
		assert.t.Errorf("%s:%d>Expected=%s, actual=%s", file, line, expected, actual)
	}
}

// CheckError fails the test if a non-nil error is passed.
func (assert Assert) CheckError(err error) {
	file, line := caller()
	if err != nil {
		assert.t.Errorf("%s:%d>Unexpected error: %s", file, line, err.Error())
	}
}
