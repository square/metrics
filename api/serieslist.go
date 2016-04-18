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

package api

import (
	"fmt"
	"time"
)

// SeriesList is a list of time series sharing the same time range.
// this struct must satisfy the `function.Value` interface. However, a type assertion
// cannot be held here due to a circular import.
type SeriesList struct {
	Series    []Timeseries `json:"series"`
	Timerange Timerange    `json:"timerange"`
}

// ToSeriesList is an identity function that allows SeriesList to implement the expression.Value interface.
func (list SeriesList) ToSeriesList(time Timerange) (SeriesList, error) {
	return list, nil
}

// ToString is a conversion function to implement the expression.Value interface.
func (list SeriesList) ToString(description string) (string, error) {
	return "", fmt.Errorf("cannot convert %s (type SeriesList) to type string", description)
}

// ToScalar is a conversion function to implement the expression.Value interface.
func (list SeriesList) ToScalar(description string) (float64, error) {
	return 0, fmt.Errorf("cannot convert %s (type SeriesList) to type scalar", description)
}

// ToDuration is a conversion function to implement the expression.Value interface.
func (list SeriesList) ToDuration(description string) (time.Duration, error) {
	return 0, fmt.Errorf("cannot convert %s (type SeriesList) to type duration", description)
}
