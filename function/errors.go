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

type LimitError interface {
	Actual() interface{} // actual from the system which triggered this error.
	Limit() interface{}  // configured limit
	error
}

func NewLimitError(message string, actual interface{}, limit interface{}) LimitError {
	return defaultLimitError{
		message: message,
		limit:   limit,
		actual:  actual,
	}
}

type defaultLimitError struct {
	message string
	actual  interface{}
	limit   interface{}
}

func (err defaultLimitError) Error() string {
	return fmt.Sprintf("%s (actual=%v limit=%v)", err.message, err.actual, err.limit)
}

func (err defaultLimitError) Actual() interface{} {
	return err.actual
}

func (err defaultLimitError) Limit() interface{} {
	return err.limit
}

type ArgumentLengthError struct {
	Name        string
	ExpectedMin int
	ExpectedMax int
	Actual      int
}

func (err ArgumentLengthError) TokenName() string {
	return err.Name
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
