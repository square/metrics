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

package function

import (
	"fmt"
)

// ExecutionError is returned if an error is occurred during
// the execution of the query.
type ExecutionError interface {
	error
	TokenName() string // name of the token / expression which have caused it.
}

// LimitError is returned if an error occurs where limits are surpassed.
type LimitError interface {
	Actual() interface{} // actual from the system which triggered this error.
	Limit() interface{}  // configured limit
	error
}

// NewLimitError uses its parameters to create a LimitError.
func NewLimitError(message string, actual interface{}, limit interface{}) LimitError {
	// @@ leaking param: message to result ~r3 level=0
	// @@ leaking param: limit to result ~r3 level=0
	// @@ leaking param: actual to result ~r3 level=0
	return defaultLimitError{
		// @@ can inline NewLimitError
		message: message,
		// @@ composite literal escapes to heap
		limit:  limit,
		actual: actual,
	}
}

type defaultLimitError struct {
	message string
	actual  interface{}
	limit   interface{}
}

// Error returns a nicely-formatted error message for the default limit error.
func (err defaultLimitError) Error() string {
	// @@ leaking param: err
	return fmt.Sprintf("%s (actual=%v limit=%v)", err.message, err.actual, err.limit)
}

// @@ err.message escapes to heap

// Actual returns the actual value in the limit comparison.
func (err defaultLimitError) Actual() interface{} {
	// @@ leaking param: err to result ~r0 level=0
	return err.actual
	// @@ can inline defaultLimitError.Actual
}

// Limit returns the limit value in the comparison (maximum or minimum).
func (err defaultLimitError) Limit() interface{} {
	// @@ leaking param: err to result ~r0 level=0
	return err.limit
	// @@ can inline defaultLimitError.Limit
}

// ArgumentLengthError is a kind of error that describes when a function is given too many or too few arguments.
type ArgumentLengthError struct {
	Name        string
	ExpectedMin int
	ExpectedMax int
	Actual      int
}

// TokenName decribes where the error occurs.
func (err ArgumentLengthError) TokenName() string {
	// @@ leaking param: err to result ~r0 level=0
	return err.Name
	// @@ can inline ArgumentLengthError.TokenName
}

// Error gives a detailed description of the error.
func (err ArgumentLengthError) Error() string {
	// @@ leaking param: err
	if err.ExpectedMin == err.ExpectedMax {
		return fmt.Sprintf(
			"Function `%s` expected %d arguments but received %d.",
			err.Name,
			err.ExpectedMin,
			// @@ err.Name escapes to heap
			err.Actual,
			// @@ err.ExpectedMin escapes to heap
		)
		// @@ err.Actual escapes to heap
	} else if err.ExpectedMax == -1 {
		return fmt.Sprintf(
			"Function `%s` expected at least %d arguments but received %d.",
			err.Name,
			err.ExpectedMin,
			// @@ err.Name escapes to heap
			err.Actual,
			// @@ err.ExpectedMin escapes to heap
		)
		// @@ err.Actual escapes to heap
	} else {
		return fmt.Sprintf(
			"Function `%s` expected between %d and %d arguments but received %d.",
			err.Name,
			err.ExpectedMin,
			// @@ err.Name escapes to heap
			err.ExpectedMax,
			// @@ err.ExpectedMin escapes to heap
			err.Actual,
			// @@ err.ExpectedMax escapes to heap
		)
		// @@ err.Actual escapes to heap
	}
}
