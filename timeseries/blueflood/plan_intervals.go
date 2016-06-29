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

package blueflood

import (
	"fmt"
	"time"

	"github.com/square/metrics/api"
)

// planFetchIntervals will plan the (point-count minimal) request intervals needed to cover the given timerange.
// the resolutions slice should be sorted, with the finest-grained resolution first.
func planFetchIntervals(resolutions []Resolution, now time.Time, requestInterval api.Interval) (map[Resolution]api.Interval, error) {
	answer := map[Resolution]api.Interval{}
	// Note: for anything other than FULL, a Blueflood returned point corresponds to the period FOLLOWING that point.
	// e.g. at 1hr resolution, a 4pm point summarizes all points in [4pm, 5pm], exclusive of 5pm.
	requestTimerange := requestInterval.CoveringTimerange(resolutions[len(resolutions)-1].Resolution)
	here := requestTimerange.Start()
	end := requestTimerange.End()
	for i := len(resolutions) - 1; i >= 0; i-- {
		resolution := resolutions[i]
		if !here.Before(end) {
			break
		}
		if here.Before(now.Add(-resolution.TimeToLive)) {
			// Expired
			return nil, fmt.Errorf("resolutions up to %+v only live for %+v, but request needs data that's at least %+v old", resolution.Resolution, resolution.TimeToLive, now.Sub(here))
		}

		// clipEnd is the end of requested interval,
		// or where the data is not yet available,
		// whichever is earlier.
		clipEnd := now.Add(-resolution.FirstAvailable)
		if end.Before(clipEnd) {
			clipEnd = end
		}

		// count how many resolution intervals pass from now until then.
		count := clipEnd.Sub(here) / resolution.Resolution
		if count < 0 {
			count = 0
		}

		// advance that number of intervals
		newHere := here.Add(count * resolution.Resolution)

		if newHere != here {
			// At least one point is included, so:
			answer[resolution] = api.Interval{Start: here, End: newHere}
			here = newHere
		}
	}
	return answer, nil
}

// planFetchIntervalsWithOnlyFiner assumes that the requested range is as coarse as desired.
// Hence, it will trim all coarser resolutions before doing planning.
func planFetchIntervalsWithOnlyFiner(resolutions []Resolution, now time.Time, requestRange api.Timerange) (map[Resolution]api.Interval, error) {
	for i := range resolutions {
		if resolutions[i].Resolution > requestRange.Resolution() {
			if i == 0 {
				return nil, fmt.Errorf("No resolutions are available at least as fine as the chosen %+v", requestRange.Resolution())
			}
			return planFetchIntervals(resolutions[:i], now, requestRange.Interval())
		}
	}
	return planFetchIntervals(resolutions, now, requestRange.Interval())
}
