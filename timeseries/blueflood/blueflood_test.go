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

var testResolutions = []Resolution{
	{
		Name:           "FULL",
		Resolution:     30 * time.Second,
		FirstAvailable: 0,
		TimeToLive:     1 * day,
	},
	{
		Name:           "5MIN",
		Resolution:     5 * time.Minute,
		FirstAvailable: 1 * day,
		TimeToLive:     30 * day,
	},
	{
		Name:           "60MIN",
		Resolution:     time.Hour,
		FirstAvailable: 15 * day,
		TimeToLive:     90 * day,
	},
	{
		Name:           "1440MIN",
		Resolution:     day,
		FirstAvailable: 180 * day,
		TimeToLive:     900 * day,
	},
}

func TestBluefloodChooseResolution(t *testing.T) {
	nowMillis := int64((time.Hour*24*365*19 + time.Minute*1734 + time.Second*17) / time.Millisecond) // Arbitrary magic constant.
	makeRange := func(beforeStart time.Duration, beforeEnd time.Duration, resolution time.Duration) api.Timerange {
		if beforeStart < beforeEnd {
			t.Fatalf("Before start must be at least as large as before end.")
		}
		timerange, err := api.NewSnappedTimerange(nowMillis-int64(beforeStart.Seconds()*1000), nowMillis-int64(beforeEnd.Seconds()*1000), int64(resolution.Seconds()*1000))
		if err != nil {
			t.Fatalf("Problem creating timerange for test.")
		}
		return timerange
	}
	a := assert.New(t)
	type test struct {
		requested  api.Timerange
		lowerBound time.Duration
		expected   time.Duration
		problem    bool
	}
	testcases := []test{
		{
			requested:  makeRange(1*time.Hour, 0, 12*time.Second),
			lowerBound: 0,
			expected:   30 * time.Second,
		},
		{
			requested:  makeRange(1*time.Hour, 0, 31*time.Second),
			lowerBound: 0,
			expected:   5 * time.Minute,
		},
		{
			requested:  makeRange(1*time.Hour, 0, 12*time.Second),
			lowerBound: 31 * time.Second,
			expected:   5 * time.Minute,
		},
		{
			requested:  makeRange(2*day, 0, 12*time.Second),
			lowerBound: 0,
			expected:   5 * time.Minute,
		},
	}
	instance := &Blueflood{
		config: Config{
			Resolutions: testResolutions,
			TimeSource: func() time.Time {
				timeValue := time.Unix(nowMillis/1000, nowMillis%1000*1e6)
				return timeValue
			},
		},
	}
	for i, test := range testcases {
		fmt.Printf("\n\n\nTEST #%d\n===========================\n\n", i+1)
		a := a.Contextf("test #%d (( %+v ))", i+1, test)
		actual, err := instance.ChooseResolution(test.requested, test.lowerBound)
		if test.problem {
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

func TestIncludeRawPayload(t *testing.T) {
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

	fakeHTTPClient := mocks.NewFakeHTTPClient()
	fakeHTTPClient.SetResponse(regularQueryURL, mocks.Response{regularResponse, 0, http.StatusOK})
	// fakeHTTPClient.SetResponse(fullResolutionQueryURL, mocks.Response{fullResolutionResponse, 0, http.StatusOK})
	defaultClientConfig.HTTPClient = fakeHTTPClient
	defaultClientConfig.TimeSource = timeSource

	b := NewBlueflood(defaultClientConfig)
	if err != nil {
		t.Fatalf("timerange error: %s", err.Error())
	}

	userConfig := timeseries.UserSpecifiableConfig{
		IncludeRawData: true,
	}

	timeSeries, err := b.FetchSingleTimeseries(timeseries.FetchRequest{
		Metric: api.TaggedMetric{
			MetricKey: api.MetricKey("some.key"),
			TagSet:    api.TagSet{"tag": "value"},
		},
		RequestDetails: timeseries.RequestDetails{
			SampleMethod:          timeseries.SampleMean,
			Timerange:             queryTimerange,
			UserSpecifiableConfig: userConfig,
		},
	})
	if err != nil {
		t.Fatalf("Expected success, but got error: %s", err.Error())
	}

	if timeSeries.Raw == nil || string(timeSeries.Raw[0]) != regularResponse {
		t.Fatalf("Didn't fill in the raw result correctly, got: %s\n", string(timeSeries.Raw[0]))
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

func TestBlueflood_UserSuppliedTTLs(t *testing.T) {
	ttls := make(map[string]int64)
	myRes := Resolution{"MIN7", time.Minute * 7}
	ttlInDays := int64(5)
	ttls["MIN7"] = ttlInDays // 7 minute resolution is available for 5 days
	conf := Config{
		TTLs: ttls,
	}
	result := conf.oldestViableDataForResolution(myRes)
	if time.Duration(ttlInDays)*24*time.Hour != result {
		t.Errorf("The custom TTL didn't make it back %v %v\n", myRes.duration, result)
	}
}

func TestBlueflood_DefaultTTLs(t *testing.T) {
	conf := Config{}
	resolution := Resolution20Min
	duration := conf.oldestViableDataForResolution(resolution)
	// 20 minutes should be available for 60 days
	if duration != 60*24*time.Hour {
		t.Fail()
	}
}

func TestBlueflood_ChooseResolution(t *testing.T) {
	makeTimerange := func(start, end, resolution int64) api.Timerange {
		timerange, err := api.NewSnappedTimerange(start, end, resolution)
		if err != nil {
			t.Fatalf("error creating testcase timerange: %s", err.Error())
		}
		return timerange
	}

	// The millisecond epoch for Sep 1, 2001.
	start := int64(999316800000)

	second := int64(1000)
	minute := 60 * second
	hour := 60 * minute
	day := 24 * hour

	tests := []struct {
		input     api.Timerange
		slotLimit int
		expected  time.Duration
	}{
		{
			input:     makeTimerange(start, start+4*hour, 30*second),
			slotLimit: 5000,
			expected:  30 * time.Second,
		},
		{
			input:     makeTimerange(start, start+4*hour, 30*second),
			slotLimit: 50,
			expected:  5 * time.Minute,
		},
		{
			input:     makeTimerange(start, start+4*hour, 30*second),
			slotLimit: 470,
			expected:  5 * time.Minute,
		},
		{
			input:     makeTimerange(start, start+40*hour, 30*second),
			slotLimit: 500,
			expected:  5 * time.Minute,
		},
		{
			input:     makeTimerange(start, start+40*hour, 30*second),
			slotLimit: 4700,
			expected:  5 * time.Minute,
		},
		{
			input:     makeTimerange(start, start+40*hour, 30*second),
			slotLimit: 110,
			expected:  1 * time.Hour,
		},
		{
			input:     makeTimerange(start, start+70*day, 30*second),
			slotLimit: 200,
			expected:  24 * time.Hour,
		},
		{
			input:     makeTimerange(start-25*day, start, 30*second),
			slotLimit: 200,
			expected:  24 * time.Hour,
		},
	}

	b := &Blueflood{
		config: Config{
			TTLs: map[string]int64{
				"FULL":    1,
				"MIN5":    30,
				"MIN20":   60,
				"MIN60":   90,
				"MIN240":  20,
				"MIN1440": 365,
			},
			TimeSource: func() time.Time {
				return time.Unix(start/1000, 0)
			},
		},
	}

	for i, test := range tests {
		smallestResolution := test.input.Duration() / time.Duration(test.slotLimit-2)
		result, err := b.ChooseResolution(test.input, smallestResolution)
		if err != nil {
			t.Errorf("Unexpected error choosing resolution: %s", err.Error())
			continue
		}
		// This is mostly a sanity check:
		_, err = api.NewSnappedTimerange(test.input.StartMillis(), test.input.EndMillis(), int64(result/time.Millisecond))
		if err != nil {
			t.Errorf("Test %+v:\nEncountered error when building timerange: %s", test, err.Error())
		}
		if result != test.expected {
			t.Errorf("Testcase %d failed: expected %+v but got %+v; slot limit %d", i, test.expected, result, test.slotLimit)
		}
	}
}
*/
