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

import "time"

// Interval is an interval of time.
type Interval struct {
	Start time.Time // Start of the interval (including this instant in time)
	End   time.Time // End of the interval (excluding this instant time)
}

// Contains tells whether the given point is contained in the interval.
func (i Interval) Contains(t time.Time) bool {
	return t.Before(i.End) && !t.Before(i.Start)
}

// Duration is the duration of the interval.
func (i Interval) Duration() time.Duration {
	return i.End.Sub(i.Start)
}

// CoveringTimerange returns the smallest timerange of the given resolution
// that covers this interval.
func (i Interval) CoveringTimerange(resolution time.Duration) Timerange {
	res := int64(resolution.Seconds() * 1000)
	return Timerange{
		start:      (i.Start.UnixNano() / 1e6) / res * res,
		end:        (i.End.UnixNano() / 1e6) / res * res,
		resolution: res,
	}
}
