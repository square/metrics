package query

import (
	"fmt"
	"strings"
)

// SyntaxError is raised when the user query is invalid.
// This can happen for two reasons:
// * The query does not generate a valid AST.
// * Invalid input is provided
type SyntaxError struct {
	token   string
	message string
}

type ArgumentLengthError struct {
	Name        string
	ExpectedMin int
	ExpectedMax int
	Actual      int
}

func (err ArgumentLengthError) Error() string {
	if err.ExpectedMin == err.ExpectedMax {
		return fmt.Sprintf(
			"Function `%s` expected %d arguments but received %d.",
			err.Name,
			err.ExpectedMin,
			err.Actual,
		)
	} else if err.ExpectedMax == -1 {
		return fmt.Sprintf(
			"Function `%s` expected at least %d arguments but received %d.",
			err.Name,
			err.ExpectedMin,
			err.Actual,
		)
	} else {
		return fmt.Sprintf(
			"Function `%s` expected between %d and %d arguments but received %d.",
			err.Name,
			err.ExpectedMin,
			err.ExpectedMax,
			err.Actual,
		)
	}
}

// AssertionError is raised when an internal invariant is violated,
// indicating a programming bug.
type AssertionError struct {
	message string
}

func (err AssertionError) Error() string {
	return err.message
}

// Token returns the token of the AST related to the parsing error.
func (err SyntaxError) Token() string {
	return err.token
}

func (err SyntaxError) Error() string {
	return err.message
}

// SyntaxErrors is a slice of SyntaxErrors implementing Error() method.
type SyntaxErrors []SyntaxError

func (errors SyntaxErrors) Error() string {
	errorStrings := make([]string, len(errors))
	for i := 0; i < len(errorStrings); i++ {
		errorStrings[i] = errors[i].Error()
	}
	return strings.Join(errorStrings, "\n")
}

var _ error = (*SyntaxError)(nil)
