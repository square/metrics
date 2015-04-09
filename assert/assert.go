// Package assert contains wrapper on top of go's testing library
// to make tests easier.
package assert

import (
	"testing"
)

// Assert is a helper struct for testing methods.
type Assert struct {
	t *testing.T
}

// New creates a new Assert struct.
func New(t *testing.T) Assert {
	return Assert{t}
}

// EqString fails the test if two strings aren't equal.
func (assert Assert) EqString(actual string, expected string) {
	if actual != expected {
		assert.t.Errorf("Expected=[%s], actual=[%s]", expected, actual)
	}
}

// EqInt fails the test if two ints aren't equal.
func (assert Assert) EqInt(actual int, expected int) {
	if actual != expected {
		assert.t.Errorf("Expected=[%d], actual=[%d]", expected, actual)
	}
}
