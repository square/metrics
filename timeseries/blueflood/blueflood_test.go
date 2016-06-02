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
	"testing"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/testing_support/assert"
)

const day = 24 * time.Hour

var resolutionFull = Resolution{
	Name:           "FULL",
	Resolution:     30 * time.Second,
	FirstAvailable: 0,
	TimeToLive:     1 * day,
}
var resolution5Min = Resolution{
	Name:           "5MIN",
	Resolution:     5 * time.Minute,
	FirstAvailable: 1 * day,
	TimeToLive:     30 * day,
}
var resolution60Min = Resolution{
	Name:           "60MIN",
	Resolution:     time.Hour,
	FirstAvailable: 15 * day,
	TimeToLive:     90 * day,
}
var resolution1440Min = Resolution{
	Name:           "1440MIN",
	Resolution:     day,
	FirstAvailable: 80 * day,
	TimeToLive:     900 * day,
}

var testResolutions = []Resolution{resolutionFull, resolution5Min, resolution60Min, resolution1440Min}

// TestPlanFetchIntervals tests whether the planning intervals behave correctly.
func TestPlanFetchIntervals(t *testing.T) {
	// Note: this constant is not completely arbitrary. It has lots of factors,
	// which means that it lies on a lot of resolution boundaries,
	// so most resolutions will be able to work without rounding (e.g., 31ms).
	nowMillis := int64(12331800) * 60000
	nowFunc := func() time.Time {
		timeValue := time.Unix(nowMillis/1000, nowMillis%1000*1e6)
		return timeValue
	}
	makeInterval := func(beforeStart time.Duration, beforeEnd time.Duration) api.Interval {
		return api.Interval{Start: nowFunc().Add(-beforeStart), End: nowFunc().Add(-beforeEnd)}
	}
	a := assert.New(t).Contextf("Blueflood planFetchIntervals")
	type test struct {
		resolutions []Resolution
		requested   api.Interval
		lowerBound  time.Duration
		expected    map[Resolution]api.Interval
		error       bool
	}
	testcases := []test{
		{
			// -1hr to now (30s only)
			resolutions: testResolutions[:1],
			requested:   makeInterval(1*time.Hour, 0),
			lowerBound:  0,
			expected: map[Resolution]api.Interval{
				// only use full resolution
				resolutionFull: makeInterval(1*time.Hour, 0),
			},
		},
		{
			// -1hr to now (30s and 5m)
			resolutions: testResolutions[:2],
			requested:   makeInterval(1*time.Hour, 0),
			lowerBound:  0,
			expected: map[Resolution]api.Interval{
				// only use full resolution
				resolutionFull: makeInterval(1*time.Hour, 0),
			},
		},
		{
			// -1hr27s to now (30s and 5m)
			resolutions: testResolutions[:2],
			requested:   makeInterval(1*time.Hour+27*time.Second, 0),
			lowerBound:  0,
			expected: map[Resolution]api.Interval{
				// only use full resolution
				resolutionFull: makeInterval(1*time.Hour+5*time.Minute, 0),
			},
		},
		{
			// -37h to now (30s, 5m, 60m)
			resolutions: testResolutions[:3],
			requested:   makeInterval(37*time.Hour, 0),
			lowerBound:  0,
			expected: map[Resolution]api.Interval{
				// use 5m until 24hr, then 30s for the rest
				resolution5Min: makeInterval(37*time.Hour, 24*time.Hour),
				resolutionFull: makeInterval(24*time.Hour, 0),
			},
		},
		{
			// -20d to -11d (30s, 5m, 60m)
			resolutions: testResolutions[:3],
			requested:   makeInterval(20*day, 11*day),
			lowerBound:  0,
			expected: map[Resolution]api.Interval{
				// use 60m and 5m resolution
				resolution60Min: makeInterval(20*day, 15*day),
				resolution5Min:  makeInterval(15*day, 11*day),
			},
		},
		{
			// -20d to now (30s, 5m, 60m)
			resolutions: testResolutions[:3],
			requested:   makeInterval(20*day, 0),
			lowerBound:  0,
			expected: map[Resolution]api.Interval{
				// use all resolutions
				resolution60Min: makeInterval(20*day, 15*day),
				resolution5Min:  makeInterval(15*day, 1*day),
				resolutionFull:  makeInterval(1*day, 0),
			},
		},
		{
			// -20d to -6hr (30s, 5m, 60m)
			resolutions: testResolutions[:3],
			requested:   makeInterval(20*day, 6*time.Hour),
			lowerBound:  0,
			expected: map[Resolution]api.Interval{
				// use all resolutions
				resolution60Min: makeInterval(20*day, 15*day),
				resolution5Min:  makeInterval(15*day, 1*day),
				resolutionFull:  makeInterval(1*day, 6*time.Hour),
			},
		},
		{
			// -1d to now (30s, 5m, 60m)
			resolutions: testResolutions[:3],
			requested:   makeInterval(1*day, 0),
			lowerBound:  0,
			expected: map[Resolution]api.Interval{
				// use only full resolution
				resolutionFull: makeInterval(1*day, 0), // the rest
			},
		},
		{
			// -901d to now (30s, 5m, 60m)
			// should error
			resolutions: testResolutions[:3],
			requested:   makeInterval(91*day, 0),
			lowerBound:  0,
			error:       true,
		},
	}
	for i, test := range testcases {
		a := a.Contextf("test #%d (input %+v)", i+1, test.requested)
		actual, err := planFetchIntervals(test.resolutions, nowFunc(), test.requested)
		if test.error {
			if err == nil {
				a.Errorf("Expected error but got: %+v", actual)
			}
		} else {
			if err != nil {
				a.Errorf("Unexpected error: %s", err.Error())
				continue
			}
			a.Eq(actual, test.expected)
		}
	}
}

func TestPlanChooseResolution(t *testing.T) {
	// Note: this constant is not completely arbitrary. It has lots of factors,
	// which means that it lies on a lot of resolution boundaries,
	// so most resolutions will be able to work without rounding (e.g., 31ms).
	nowMillis := int64(12331800) * 60000
	nowFunc := func() time.Time {
		timeValue := time.Unix(nowMillis/1000, nowMillis%1000*1e6)
		return timeValue
	}
	makeRange := func(beforeStart time.Duration, beforeEnd time.Duration, resolution time.Duration) api.Timerange {
		if beforeStart < beforeEnd {
			t.Fatalf("Before start must be at least as large as before end.")
		}
		// Note: it's not snapped so that we don't accidentally alter the ends of the timerange via a snap.
		timerange, err := api.NewTimerange(nowMillis-int64(beforeStart.Seconds()*1000), nowMillis-int64(beforeEnd.Seconds()*1000), int64(resolution.Seconds()*1000))
		if err != nil {
			panic(fmt.Sprintf("Problem creating timerange for test: %s", err.Error()))
		}
		return timerange
	}
	a := assert.New(t).Contextf("Blueflood ChooseResolution")
	type test struct {
		requested  api.Timerange
		lowerBound time.Duration
		expected   time.Duration
		error      bool
	}
	testcases := []test{
		{
			requested:  makeRange(1*time.Hour, 0, 12*time.Second),
			lowerBound: 0,
			expected:   30 * time.Second,
		},
		{
			requested:  makeRange(1*time.Hour+27*time.Second, 0, 31*time.Second),
			lowerBound: 0,
			expected:   5 * time.Minute, // because of hint
		},
		{
			requested:  makeRange(37*time.Hour, 0, 12*time.Second),
			lowerBound: 0,
			expected:   5 * time.Minute,
		},
		{
			requested:  makeRange(20*day, 11*day, 12*time.Second),
			lowerBound: 0,
			expected:   5 * time.Minute,
		},
		{
			requested:  makeRange(20*day, 0, 12*time.Second),
			lowerBound: 0,
			expected:   5 * time.Minute,
		},
		{
			requested:  makeRange(20*day, 0, 30*time.Second),
			lowerBound: 0,
			expected:   5 * time.Minute,
		},
		{
			requested:  makeRange(1*day, 0, 12*time.Second),
			lowerBound: 0,
			expected:   30 * time.Second,
		},
		{
			requested:  makeRange(901*day, 0, 30*time.Second),
			lowerBound: 0,
			error:      true,
		},
	}
	for i, test := range testcases {
		a := a.Contextf("test #%d (input %+v)", i+1, test.requested)
		actual, err := (&Blueflood{config: Config{TimeSource: nowFunc, Resolutions: testResolutions}}).ChooseResolution(test.requested, test.lowerBound)
		if test.error {
			if err == nil {
				a.Errorf("Expected error but got: %+v", actual)
			}
		} else {
			if err != nil {
				a.Errorf("Unexpected error: %s", err.Error())
				continue
			}
			a.Eq(actual, test.expected)
		}
	}
}
