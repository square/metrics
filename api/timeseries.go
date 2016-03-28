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

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"math"
	"strconv"
)

// Timeseries is a single time series, identified with the associated tagset.
type Timeseries struct {
	Values []float64
	TagSet TagSet
	Raw    [][]byte
}

// MarshalJSON exists to manually encode floats.
func (ts Timeseries) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	var scratch [64]byte
	buffer.WriteByte('{')
	buffer.WriteString("\"tagset\":")
	tagset, err := json.Marshal(ts.TagSet)
	if err != nil {
		return []byte{}, err
	}
	buffer.Write(tagset)
	buffer.WriteByte(',')

	if ts.Raw != nil {
		buffer.WriteString("\"raw\":")
		buffer.WriteByte('[')
		first := true
		for _, raw := range ts.Raw {
			if !first {
				buffer.WriteByte(',')
			}
			buffer.WriteByte('[')
			buffer.WriteByte('"')
			base64Wrapped := base64.StdEncoding.EncodeToString(raw)
			buffer.WriteString(base64Wrapped)
			buffer.WriteByte('"')
			buffer.WriteByte(']')
			first = false
		}
		// raw, _ := json.Marshal(ts.Raw)
		buffer.WriteByte(']')
		buffer.WriteByte(',')
	}

	// buffer.WriteByte(',')
	buffer.WriteString("\"values\":")
	buffer.WriteByte('[')
	n := len(ts.Values)
	for i := 0; i < n; i++ {
		if i > 0 {
			buffer.WriteByte(',')
		}
		f := ts.Values[i]
		if math.IsInf(f, 1) {
			buffer.WriteString("null") // TODO - positive infinity
		} else if math.IsInf(f, -1) {
			buffer.WriteString("null") // TODO - negative infinity
		} else if math.IsNaN(f) {
			buffer.WriteString("null")
		} else {
			b := strconv.AppendFloat(scratch[:0], f, 'g', -1, 64)
			buffer.Write(b)
		}
	}
	buffer.WriteByte(']')
	buffer.WriteByte('}')
	return buffer.Bytes(), err
}
