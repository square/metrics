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

// Package api holds common data types and public interface exposed by the indexer library.
// Refer to the doc
// https://docs.google.com/a/squareup.com/document/d/1k0Wgi2wnJPQoyDyReb9dyIqRrD8-v0u8hz37S282ii4/edit
// for the terminology.

package api

// SeriesType is a different aspect of data.
// For example, Blueflood may stores (min / max / average / count) during rollups,
// and these data are exposed via columns
type SeriesType string

// SampleMethod determines how the given time series should be sampled.
type SampleMethod int

const (
	// SamplingMax chooses the maximum value.
	SampleMax SampleMethod = iota + 1
	// SamplingMin chooses the minimum value.
	SampleMin
	// SamplingMean chooses the average value.
	SampleMean
)

func (sm SampleMethod) String() string {
	switch sm {
	case SampleMax:
		return "SampleMax"
	case SampleMin:
		return "SampleMin"
	case SampleMean:
		return "SampleMean"
	}

	return "unknown"
}
