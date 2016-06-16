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
	"bytes"
	"encoding/json"
	"math"
	"strconv"
	"time"

	"github.com/square/metrics/api"
)

type TaggedScalar struct {
	TagSet api.TagSet
	Value  float64
}

// MarshalJSON for TaggedScalar marshals NaN or infinity to null.
func (ts TaggedScalar) MarshalJSON() ([]byte, error) {
	// @@ leaking param: ts
	var buffer bytes.Buffer
	buffer.WriteString(`{"tagset":`)
	// @@ moved to heap: buffer
	tagset, err := json.Marshal(ts.TagSet)
	// @@ buffer escapes to heap
	if err != nil {
		// @@ ts.TagSet escapes to heap
		return nil, err
	}
	buffer.Write(tagset)
	buffer.WriteString(`,"value":`)
	// @@ buffer escapes to heap
	if math.IsInf(ts.Value, 0) || math.IsNaN(ts.Value) {
		// @@ buffer escapes to heap
		buffer.WriteString(`null`)
		// @@ inlining call to math.IsInf
		// @@ inlining call to math.IsNaN
	} else {
		// @@ buffer escapes to heap
		buffer.WriteString(strconv.FormatFloat(ts.Value, 'g', -1, 64))
	}
	// @@ buffer escapes to heap
	buffer.WriteString("}")
	return buffer.Bytes(), nil
	// @@ buffer escapes to heap
}

// @@ inlining call to Bytes

type ScalarSet []TaggedScalar

func (set ScalarSet) ToSeriesList(timerange api.Timerange) (api.SeriesList, *ConversionFailure) {
	// @@ leaking param content: set
	list := api.SeriesList{
		Series: make([]api.Timeseries, len(set)),
	}
	// @@ make([]api.Timeseries, len(set)) escapes to heap
	// @@ make([]api.Timeseries, len(set)) escapes to heap
	for i := range list.Series {
		list.Series[i] = api.Timeseries{
			TagSet: set[i].TagSet,
			Values: make([]float64, timerange.Slots()),
		}
		// @@ inlining call to api.Timerange.Slots
		// @@ make([]float64, int(~r0)) escapes to heap
		// @@ make([]float64, int(~r0)) escapes to heap
		for j := range list.Series[i].Values {
			list.Series[i].Values[j] = set[i].Value
		}
	}
	return list, nil
}
func (set ScalarSet) ToString() (string, *ConversionFailure) {
	return "", &ConversionFailure{
		// @@ can inline ScalarSet.ToString
		From: "scalar set",
		To:   "string",
	}
	// @@ &ConversionFailure literal escapes to heap
}
func (set ScalarSet) ToScalar() (float64, *ConversionFailure) {
	if len(set) == 1 && set[0].TagSet.Equals(api.TagSet{}) {
		return set[0].Value, nil
	}
	return 0, &ConversionFailure{
		From: "scalar set",
		To:   "scalar",
	}
	// @@ &ConversionFailure literal escapes to heap
}
func (set ScalarSet) ToScalarSet() (ScalarSet, *ConversionFailure) {
	// @@ leaking param: set to result ~r0 level=0
	return set, nil
	// @@ can inline ScalarSet.ToScalarSet
}
func (set ScalarSet) ToDuration() (time.Duration, *ConversionFailure) {
	return 0, &ConversionFailure{
		// @@ can inline ScalarSet.ToDuration
		From: "scalar set",
		To:   "duration",
	}
	// @@ &ConversionFailure literal escapes to heap
}
