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
	TimeToLive:     1*day + 1*time.Minute, // If the TTL is exactly 1 day, it will be excluded unnecesarily.
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
	FirstAvailable: 180 * day,
	TimeToLive:     900 * day,
}

var testResolutions = []Resolution{resolutionFull, resolution5Min, resolution60Min, resolution1440Min}

func TestPlanFetchIntervals(t *testing.T) {
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
	a := assert.New(t).Contextf("Blueflood planFetchIntervals")
	type test struct {
		requested  api.Timerange
		lowerBound time.Duration
		expected   map[Resolution]api.Timerange
		error      bool
	}
	testcases := []test{
		{
			requested:  makeRange(1*time.Hour, 0, 12*time.Second),
			lowerBound: 0,
			expected: map[Resolution]api.Timerange{
				resolutionFull: makeRange(1*time.Hour, 0, 30*time.Second),
			},
		},
		{
			requested:  makeRange(1*time.Hour+27*time.Second, 0, 31*time.Second),
			lowerBound: 0,
			expected: map[Resolution]api.Timerange{
				resolutionFull: makeRange(1*time.Hour, 0, 30*time.Second),
			},
		},
		{
			requested:  makeRange(37*time.Hour, 0, 12*time.Second),
			lowerBound: 0,
			expected: map[Resolution]api.Timerange{
				resolution5Min: makeRange(37*time.Hour, 24*time.Hour, 5*time.Minute),
				resolutionFull: makeRange(24*time.Hour-30*time.Second, 0, 30*time.Second),
			},
		},
		{
			requested:  makeRange(20*day, 11*day, 12*time.Second),
			lowerBound: 0,
			expected: map[Resolution]api.Timerange{
				resolution60Min: makeRange(20*day, 15*day, 60*time.Minute),
				resolution5Min:  makeRange(15*day-5*time.Minute, 11*day, 5*time.Minute),
			},
		},
		{
			requested:  makeRange(20*day, 0, 12*time.Second),
			lowerBound: 0,
			expected: map[Resolution]api.Timerange{
				resolution60Min: makeRange(20*day, 15*day, 60*time.Minute),
				resolution5Min:  makeRange(15*day-5*time.Minute, 1*day, 5*time.Minute),
				resolutionFull:  makeRange(1*day-30*time.Second, 0, 30*time.Second),
			},
		},
		{
			requested:  makeRange(20*day, 0, 30*time.Second),
			lowerBound: 0,
			expected: map[Resolution]api.Timerange{
				resolution60Min: makeRange(20*day, 15*day, 60*time.Minute),
				resolution5Min:  makeRange(15*day-5*time.Minute, 1*day, 5*time.Minute),
				resolutionFull:  makeRange(1*day-30*time.Second, 0, 30*time.Second),
			},
		},
		{
			requested:  makeRange(1*day, 0, 12*time.Second),
			lowerBound: 0,
			expected: map[Resolution]api.Timerange{
				// These seems very wasteful, but it seems consistent.
				// In practice, this should be fixed by staggering TTL and first available
				resolution5Min: makeRange(1*day, 1*day, 5*time.Minute),             // one point
				resolutionFull: makeRange(1*day-30*time.Second, 0, 30*time.Second), // the rest
			},
		},
		{
			requested:  makeRange(901*day, 0, 30*time.Second),
			lowerBound: 0,
			error:      true,
		},
	}
	for i, test := range testcases {
		a := a.Contextf("test #%d (input %+v)", i+1, test.requested)
		actual, err := planFetchIntervals(testResolutions, nowFunc(), test.requested)
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

func TestPlanFetchIntervalsRestricted(t *testing.T) {
	// planFetchIntervalsRestricted uses the computed resolution to intelligently
	// plan the fetch: it won't downsample if the high-resolution data is good enough.

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
	a := assert.New(t).Contextf("Blueflood planFetchIntervalsRestricted")
	type test struct {
		requested  api.Timerange
		lowerBound time.Duration
		expected   map[Resolution]api.Timerange
		error      bool
	}
	testcases := []test{
		{
			requested:  makeRange(1*time.Hour, 0, 30*time.Second),
			lowerBound: 0,
			expected: map[Resolution]api.Timerange{
				resolutionFull: makeRange(1*time.Hour, 0, 30*time.Second),
			},
		},
		{
			requested:  makeRange(1*time.Hour+30*time.Second, 0, 30*time.Second),
			lowerBound: 0,
			expected: map[Resolution]api.Timerange{
				resolutionFull: makeRange(1*time.Hour+30*time.Second, 0, 30*time.Second),
			},
		},
		{
			requested:  makeRange(37*time.Hour, 0, 5*time.Minute),
			lowerBound: 0,
			expected: map[Resolution]api.Timerange{
				resolution5Min: makeRange(37*time.Hour, 24*time.Hour, 5*time.Minute),
				resolutionFull: makeRange(24*time.Hour-30*time.Second, 0, 30*time.Second),
			},
		},
		{
			requested:  makeRange(20*day, 11*day, 60*time.Minute),
			lowerBound: 0,
			expected: map[Resolution]api.Timerange{
				resolution60Min: makeRange(20*day, 15*day, 60*time.Minute),
				resolution5Min:  makeRange(15*day-5*time.Minute, 11*day, 5*time.Minute),
			},
		},
		{
			requested:  makeRange(20*day, 0, 60*time.Minute),
			lowerBound: 0,
			expected: map[Resolution]api.Timerange{
				resolution60Min: makeRange(20*day, 15*day, 60*time.Minute),
				resolution5Min:  makeRange(15*day-5*time.Minute, 1*day, 5*time.Minute),
				resolutionFull:  makeRange(1*day-30*time.Second, 0, 30*time.Second),
			},
		},
		{
			requested:  makeRange(20*day, 0, 60*time.Minute),
			lowerBound: 0,
			expected: map[Resolution]api.Timerange{
				resolution60Min: makeRange(20*day, 15*day, 60*time.Minute),
				resolution5Min:  makeRange(15*day-5*time.Minute, 1*day, 5*time.Minute),
				resolutionFull:  makeRange(1*day-30*time.Second, 0, 30*time.Second),
			},
		},
		{
			requested:  makeRange(1*day, 0, 30*time.Second),
			lowerBound: 0,
			expected: map[Resolution]api.Timerange{
				resolutionFull: makeRange(1*day, 0, 30*time.Second), // the rest
			},
		},
		{
			requested:  makeRange(901*day, 0, 30*time.Second),
			lowerBound: 0,
			error:      true,
		},
	}
	for i, test := range testcases {
		a := a.Contextf("test #%d (input %+v)", i+1, test.requested)
		actual, err := planFetchIntervalsRestricted(testResolutions, nowFunc(), test.requested)
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

/*
func Test_Blueflood(t *testing.T) {
	timerange, err := api.NewTimerange(12000, 13000, 1000)
	if err != nil {
		t.Fatalf("invalid testcase timerange")
		return
	}

	graphite := mocks.FakeGraphiteConverter{
		MetricMap: map[util.GraphiteMetric]api.TaggedMetric{
			util.GraphiteMetric("some.key.graphite"): {
				MetricKey: api.MetricKey("some.key"),
				TagSet:    api.TagSet{"tag": "value"},
			},
		},
	}

	defaultClientConfig := Config{
		BaseURL:                 "https://blueflood.url",
		TenantID:                "square",
		Timeout:                 time.Millisecond,
		GraphiteMetricConverter: &graphite,
	}
	// Not really MIN1440, but that's what default TTLs will get with the Timerange we use
	defaultQueryURL := "https://blueflood.url/v2.0/square/views/some.key.graphite?from=12000&resolution=MIN1440&select=numPoints%2Caverage&to=14000"

	for _, test := range []struct {
		name               string
		metricMap          map[util.GraphiteMetric]api.TaggedMetric
		queryMetric        api.TaggedMetric
		sampleMethod       timeseries.SampleMethod
		timerange          api.Timerange
		clientConfig       Config
		queryURL           string
		queryResponse      string
		queryResponseCode  int
		queryDelay         time.Duration
		expectedErrorCode  timeseries.ErrorCode
		expectedSeriesList api.Timeseries
	}{
		{
			name: "Success case",
			queryMetric: api.TaggedMetric{
				MetricKey: api.MetricKey("some.key"),
				TagSet:    api.TagSet{"tag": "value"},
			},
			sampleMethod: timeseries.SampleMean,
			timerange:    timerange,
			queryURL:     defaultQueryURL,
			clientConfig: defaultClientConfig,
			queryResponse: `{
        "unit": "unknown",
        "values": [
          {
            "numPoints": 1,
            "timestamp": 12000,
            "average": 5
          },
          {
            "numPoints": 1,
            "timestamp": 13000,
            "average": 3
          }
        ],
        "metadata": {
          "limit": null,
          "next_href": null,
          "count": 2,
          "marker": null
        }
      }`,
			expectedSeriesList: api.Timeseries{
				Values: []float64{5, 3},
				TagSet: api.TagSet{"tag": "value"},
			},
		},
		{
			name: "Failure case - invalid JSON",
			queryMetric: api.TaggedMetric{
				MetricKey: api.MetricKey("some.key"),
				TagSet:    api.TagSet{"tag": "value"},
			},
			sampleMethod:      timeseries.SampleMean,
			timerange:         timerange,
			clientConfig:      defaultClientConfig,
			queryURL:          defaultQueryURL,
			queryResponse:     `{invalid}`,
			expectedErrorCode: timeseries.FetchIOError,
		},
		{
			name: "Failure case - HTTP error",
			queryMetric: api.TaggedMetric{
				MetricKey: api.MetricKey("some.key"),
				TagSet:    api.TagSet{"tag": "value"},
			},
			sampleMethod:      timeseries.SampleMean,
			timerange:         timerange,
			clientConfig:      defaultClientConfig,
			queryURL:          defaultQueryURL,
			queryResponse:     `{}`,
			queryResponseCode: 400,
			expectedErrorCode: timeseries.FetchIOError,
		},
		{
			name: "Failure case - timeout",
			queryMetric: api.TaggedMetric{
				MetricKey: api.MetricKey("some.key"),
				TagSet:    api.TagSet{"tag": "value"},
			},
			sampleMethod:      timeseries.SampleMean,
			timerange:         timerange,
			clientConfig:      defaultClientConfig,
			queryURL:          defaultQueryURL,
			queryResponse:     `{}`,
			queryDelay:        1 * time.Second,
			expectedErrorCode: timeseries.FetchTimeoutError,
		},
	} {
		a := assert.New(t).Contextf("%s", test.name)

		fakeHTTPClient := mocks.NewFakeHTTPClient()
		code := test.queryResponseCode
		if code == 0 {
			code = http.StatusOK
		}
		fakeHTTPClient.SetResponse(test.queryURL, mocks.Response{test.queryResponse, test.queryDelay, code})

		test.clientConfig.HTTPClient = fakeHTTPClient

		b := NewBlueflood(test.clientConfig).(*Blueflood)

		seriesList, err := b.FetchSingleTimeseries(timeseries.FetchRequest{
			Metric: test.queryMetric,
			RequestDetails: timeseries.RequestDetails{
				SampleMethod: test.sampleMethod,
				Timerange:    test.timerange,
			},
		})

		if test.expectedErrorCode != 0 {
			if err == nil {
				a.Errorf("Expected error, but was successful.")
				continue
			}
			berr, ok := err.(timeseries.Error)
			if !ok {
				a.Errorf("Failed to cast error to TimeseriesStorageError")
				continue
			}
			a.Eq(berr.Code, test.expectedErrorCode)
		} else {
			if err != nil {
				a.CheckError(err)
				continue
			}
			a.Eq(seriesList, test.expectedSeriesList)
		}
	}
}

func TestFullResolutionDataFilling(t *testing.T) {
	graphite := mocks.FakeGraphiteConverter{
		MetricMap: map[util.GraphiteMetric]api.TaggedMetric{
			util.GraphiteMetric("some.key.value"): {
				MetricKey: api.MetricKey("some.key"),
				TagSet:    api.TagSet{"tag": "value"},
			},
		},
	}

	now := time.Unix(1438734300000, 0)

	baseTime := now.Unix() * 1000
	timeSource := func() time.Time { return now }

	queryTimerange, err := api.NewSnappedTimerange(
		int64(baseTime)-300*1000*10, // 50 minutes ago
		int64(baseTime)-300*1000*4,  // 20 minutes ago
		300*1000,                    // 5 minute resolution
	)

	// The queries have to be relative to "now"
	defaultClientConfig := Config{
		BaseURL:                 "https://blueflood.url",
		TenantID:                "square",
		Timeout:                 time.Millisecond,
		GraphiteMetricConverter: &graphite,
		TimeSource:              timeSource,
	}

	regularQueryURL := fmt.Sprintf(
		"https://blueflood.url/v2.0/square/views/some.key.value?from=%d&resolution=MIN5&select=numPoints%%2Caverage&to=%d",
		queryTimerange.StartMillis(),
		queryTimerange.EndMillis()+queryTimerange.ResolutionMillis(),
	)

	regularResponse := fmt.Sprintf(`{
	  "unit": "unknown",
	  "values": [
	    {
	      "numPoints": 28,
	      "timestamp": %d,
	      "average": 100
	    },
	    {
	      "numPoints": 29,
	      "timestamp": %d,
	      "average": 142
	    },
	    {
	      "numPoints": 27,
	      "timestamp": %d,
	      "average": 138
	    },
	    {
	      "numPoints": 28,
	      "timestamp": %d,
	      "average": 182
	    }
	  ],
	  "metadata": {
	    "limit": null,
	    "next_href": null,
	    "count": 4,
	    "marker": null
	  }
	}`,
		baseTime-300*1000*10, // 50 minutes ago
		baseTime-300*1000*9,  // 45 minutes ago
		baseTime-300*1000*8,  // 40 minutes ago
		baseTime-300*1000*7,  // 35 minutes ago
	)

	fullResolutionQueryURL := fmt.Sprintf(
		"https://blueflood.url/v2.0/square/views/some.key.value?from=%d&resolution=FULL&select=numPoints%%2Caverage&to=%d",
		queryTimerange.StartMillis(),
		queryTimerange.EndMillis()+queryTimerange.ResolutionMillis(),
	)
	fullResolutionResponse := fmt.Sprintf(`{
	  "unit": "unknown",
	  "values": [
	    {
	      "numPoints": 28,
	      "timestamp": %d,
	      "average": 13
	    },
	    {
	      "numPoints": 29,
	      "timestamp": %d,
	      "average": 16
	    },
	    {
	      "numPoints": 27,
	      "timestamp": %d,
	      "average": 19
	    },
	    {
	      "numPoints": 28,
	      "timestamp": %d,
	      "average": 27
	    }
	  ],
	  "metadata": {
	    "limit": null,
	    "next_href": null,
	    "count": 4,
	    "marker": null
	  }
	}`,
		baseTime-300*1000*6,      // 30m ago
		baseTime-300*1000*5+17,   // 25m ago with random shuffling
		baseTime-300*1000*4+2821, // 20m ago with random shuffling
		baseTime-300*1000*3,      // 15m ago
	)

	fakeHTTPClient := mocks.NewFakeHTTPClient()
	fakeHTTPClient.SetResponse(regularQueryURL, mocks.Response{regularResponse, 0, http.StatusOK})
	fakeHTTPClient.SetResponse(fullResolutionQueryURL, mocks.Response{fullResolutionResponse, 0, http.StatusOK})
	defaultClientConfig.HTTPClient = fakeHTTPClient
	defaultClientConfig.TimeSource = timeSource

	b := NewBlueflood(defaultClientConfig)

	if err != nil {
		t.Fatalf("timerange error: %s", err.Error())
	}

	seriesList, err := b.FetchSingleTimeseries(timeseries.FetchRequest{
		Metric: api.TaggedMetric{
			MetricKey: api.MetricKey("some.key"),
			TagSet:    api.TagSet{"tag": "value"},
		},
		RequestDetails: timeseries.RequestDetails{
			SampleMethod: timeseries.SampleMean,
			Timerange:    queryTimerange,
		},
	})
	if err != nil {
		t.Fatalf("Expected success, but got error: %s", err.Error())
	}
	expected := []float64{100, 142, 138, 182, 13, 16, 19}
	if len(seriesList.Values) != len(expected) {
		t.Fatalf("Expected %+v but got %+v", expected, seriesList)
	}
	for i, expect := range expected {
		if seriesList.Values[i] != expect {
			t.Fatalf("Expected %+v but got %+v", expected, seriesList)
		}
	}
}

*/
