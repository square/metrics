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

package api

import "time"

// SeriesList is a list of time series sharing the same time range.
// this struct must satisfy the `function.Value` interface. However, a type assertion
// cannot be held here due to a circular import.
type SeriesList struct {
	Series    []Timeseries `json:"series"`
	Timerange Timerange    `json:"timerange"`
}

// IsValid determines whether the given time series is valid.
func (list SeriesList) isValid() bool {
	for _, series := range list.Series {
		// # of slots per series must be valid.
		if len(series.Values) != list.Timerange.Slots() {
			return false
		}
	}
	return true // validation is now successful.
}

func (list SeriesList) ToSeriesList(time Timerange) (SeriesList, error) {
	return list, nil
}

func (list SeriesList) ToString(description string) (string, error) {
	return "", ConversionError{"SeriesList", "string", description}
}

func (list SeriesList) ToScalar(description string) (float64, error) {
	return 0, ConversionError{"SeriesList", "scalar", description}
}

func (list SeriesList) ToDuration(description string) (time.Duration, error) {
	return 0, ConversionError{"SeriesList", "duration", description}
}
