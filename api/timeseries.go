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
	"bytes"
	"encoding/json"
	"math"
	"strconv"
)

// Timeseries is a single time series, identified with the associated tagset.
type Timeseries struct {
	Values []float64 `json:"values"`
	TagSet TagSet    `json:"tagset"`
}

// MarshalJSON exists to manually encode floats.
func (ts Timeseries) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{"tagset":`)
	tagset, err := json.Marshal(ts.TagSet)
	if err != nil {
		return nil, err
	}
	buffer.Write(tagset)
	buffer.WriteString(`,"values":[`)
	for i, y := range ts.Values {
		if i > 0 {
			buffer.WriteByte(',')
		}
		if math.IsInf(y, 0) || math.IsNaN(y) {
			buffer.WriteString(`null`)
			continue
		}
		buffer.WriteString(strconv.FormatFloat(y, 'g', -1, 64))
	}
	buffer.WriteString("]}")
	return buffer.Bytes(), nil
}
