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
	"encoding/json"
	"fmt"
	"time"
)

// Timerange represents a range of time a given time series is defined in:
// it is 3-tuple of (start, end, resolution) with the following constraints:
// start <= end
// start = 0 mod resolution
// end =   0 mod resolution
//
// This range is inclusive of Start and End (i.e. [Start, End]). Start and End
// are Unix milliseconds timestamps. Resolution is in milliseconds.
// (Millisecond precision allows an effective range of 290 million years in each direction)
type Timerange struct {
	start      int64
	end        int64
	resolution int64
}

// StartMillis returns the number of milliseconds between the epoch and the start of the timerange.
// The start is inclusive.
// StartMillis() is always divisible by ResolutionMillis()
func (tr Timerange) StartMillis() int64 {
	return tr.start
}

// Start returns the time.Time value corresponding to the start of the timerange (inclusive).
// Start always divides evenly into Duration().
func (tr Timerange) Start() time.Time {
	seconds := tr.start / 1000
	nanoseconds := (tr.start % 1000) * 1000000
	return time.Unix(seconds, nanoseconds)
}

// EndMillis returns the number of milliseconds between the epoch and the end of the timerange.
// The end is inclusive.
func (tr Timerange) EndMillis() int64 {
	return tr.end
}

// End returns the time.Time value corresponding to the end of the timerange (inclusive).
func (tr Timerange) End() time.Time {
	seconds := tr.end / 1000
	milliseconds := tr.start % 1000
	nanoseconds := milliseconds * 1000000
	return time.Unix(seconds, nanoseconds)
}

// DurationMillis returns the number of milliseconds in a timerange.
func (tr Timerange) DurationMillis() int64 {
	return tr.end - tr.start
}

// Duration returns a time.Duration value corresponding to the length of the timerange.
func (tr Timerange) Duration() time.Duration {
	return tr.End().Sub(tr.Start())
}

// MarshalJSON marshals the Timerange into a byte error
func (tr Timerange) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Start      int64 `json:"start"`
		End        int64 `json:"end"`
		Resolution int64 `json:"resolution"`
	}{tr.start, tr.end, tr.resolution})
}

// ResolutionMillis returns the .resolution field
func (tr Timerange) ResolutionMillis() int64 {
	return tr.resolution
}

// Resolution returns the resolution in time.Duration.
func (tr Timerange) Resolution() time.Duration {
	return time.Duration(tr.resolution) * time.Millisecond
}

// NewTimerange creates a timerange which is validated, providing error otherwise.
func NewTimerange(start, end, resolution int64) (Timerange, error) {
	if resolution <= 0 {
		return Timerange{}, fmt.Errorf("invalid resolution %d", resolution)
	}

	if start%resolution != 0 {
		return Timerange{}, fmt.Errorf("start %% resolution (mod) must be 0 (start=%d, resolution=%d)", start, resolution)
	}
	if end%resolution != 0 {
		return Timerange{}, fmt.Errorf("end %% resolution (mod) must be 0 (end=%d, resolution=%d)", end, resolution)
	}
	if start > end {
		return Timerange{}, fmt.Errorf("start must be <= end (start=%d, end=%d)", start, end)
	}
	return Timerange{start: start, end: end, resolution: resolution}, nil
}

// NewSnappedTimerange creates a new timerange and properly snaps it
func NewSnappedTimerange(start, end, resolution int64) (Timerange, error) {
	if resolution <= 0 {
		return Timerange{}, fmt.Errorf("invalid resolution %d", resolution)
	}
	if start > end {
		return Timerange{}, fmt.Errorf("start must be <= end (start=%d, end=%d)", start, end)
	}
	return Timerange{start: start, end: end, resolution: resolution}.Snap(), nil
}

func snap(n, boundary int64) int64 {
	if n < 0 {
		return -snap(-n, boundary)
	}
	// This performs a round.
	// Dividing by `boundary` truncates towards zero.
	// The resulting integer is then multiplied by `boundary` again.
	// Thus the result is a multiple of `boundary`.
	// For integer division, x/r*r = (x/r)*r in general rounds to a multiple of r towards 0.
	// Adding `boundary/2` changes this instead to a "round to nearest" rather than "round towards 0".
	// (Where "up" is the round for values exactly halfway between).
	// These halfway points round "away from zero" (rather than "towards -infinity").
	return (n + boundary/2) / boundary * boundary
}

// Snap will fix some invalid timeranges by rounding their starts and ends.
func (tr Timerange) Snap() Timerange {
	if tr.resolution == 0 {
		panic("Unable to snap with resolution of 0")
	}
	tr.start = snap(tr.start, tr.resolution)
	tr.end = snap(tr.end, tr.resolution)
	if tr.end < tr.start {
		tr.end = tr.start // This better preserves the invariants without having to return an error.
	}
	return tr
}

// Shift returns a timerange which is shifted in time by the amount given
func (tr Timerange) Shift(shift time.Duration) Timerange {
	tr.start += int64(shift / time.Millisecond)
	tr.end += int64(shift / time.Millisecond)
	return tr.Snap()
}

// SelectLength returns a timerange which whose length is set to the given amount
func (tr Timerange) SelectLength(length time.Duration) Timerange {
	tr.end = tr.start + int64(length/time.Millisecond)
	return tr.Snap()
}

// ExtendBefore increases the length of the timerange by the given duration.
func (tr Timerange) ExtendBefore(length time.Duration) Timerange {
	tr.start -= int64(length / time.Millisecond)
	return tr.Snap()
}

// Slots represent the total # of data points
// Behavior is undefined when operating on an invalid Timerange. There's a
// circular dependency here, but it all works out.
func (tr Timerange) Slots() int {
	return int((tr.end-tr.start)/tr.resolution) + 1
}
