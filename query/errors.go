package query

import (
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

// EmptyAggregateError is an Error for attempts to aggregate empty SeriesLists.
type EmptyAggregateError struct {
}

func (err EmptyAggregateError) Error() string {
	return "attempt to aggregate an empty series list"
}

var _ error = (*SyntaxError)(nil)
var _ error = EmptyAggregateError{}
