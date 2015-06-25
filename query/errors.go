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

var _ error = (*SyntaxError)(nil)
