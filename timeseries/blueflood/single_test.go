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
	"github.com/square/metrics/testing_support/mocks"
	"github.com/square/metrics/timeseries"
	"github.com/square/metrics/util"
)

func TestBluefloodHTTPQueriesSingle(t *testing.T) {
	// from -2m to now
	nowMillis := int64(739908000000)
	nowFunc := TimeSource{GetTime: func() time.Time {
		timeValue := time.Unix(nowMillis/1000, nowMillis%1000*1e6)
		return timeValue
	}}
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

	type test struct {
		request            timeseries.FetchRequest
		expectedURLs       []string
		expectedTimeseries api.Timeseries
		error              bool
	}
	testClient := mocks.NewFakeHTTPClient()
	testClient.SetResponse("https://blueflood.url/v2.0/square/views/some.key.graphite?from=739907880000&resolution=FULL&select=numPoints%2Caverage&to=739907999999", mocks.Response{
		Body: fmt.Sprintf( // Note: some of the timestamps are off slightly in order to test its behavior.
			`{
			"unit": "unknown",
			"values": [
				{
					"numPoints": 1,
					"timestamp": 739907880026,
					"average": 5
				},
				{
					"numPoints": 1,
					"timestamp": 739907910078,
					"average": 8
				},
				{
					"numPoints": 1,
					"timestamp": 739907910379,
					"note": "shares bucket with previous",
					"average": 10
				},
				{
					"numPoints": 1,
					"timestamp": 739907940000,
					"average": -72.13
				},
				{
					"numPoints": 1,
					"timestamp": 739907970088,
					"average": 6
				},
				{
					"numPoints": 1,
					"timestamp": 739908000000,
					"average": 4.5
				}
			],
			"metadata": {
				"limit": null,
				"next_href": null,
				"count": 5,
				"marker": null
			}
		}`),
		Delay:      0,
		StatusCode: 200,
	})
	config := Config{
		BaseURL:                 "https://blueflood.url",
		TenantID:                "square",
		Resolutions:             []Resolution{resolutionFull, resolution5Min, resolution60Min, resolution1440Min},
		MaxSimultaneousRequests: 5,

		GraphiteMetricConverter: &mocks.FakeGraphiteConverter{
			MetricMap: map[util.GraphiteMetric]api.TaggedMetric{
				"some.key.graphite": {
					MetricKey: "some.key",
					TagSet:    api.TagSet{"tag": "value"},
				},
			},
		},

		HTTPClient: testClient,
		TimeSource: nowFunc,
	}
	blueflood := NewBlueflood(config)
	request := timeseries.FetchRequest{
		Metric: api.TaggedMetric{MetricKey: "some.key", TagSet: api.TagSet{"tag": "value"}},
		RequestDetails: timeseries.RequestDetails{
			SampleMethod: timeseries.SampleMean,
			Timerange:    makeRange(2*time.Minute, 0, 30*time.Second),
		},
	}
	expected := api.Timeseries{
		Values: []float64{5, 9, -72.13, 6, 4.5},
		TagSet: api.TagSet{"tag": "value"},
	}
	result, err := blueflood.FetchSingleTimeseries(request)
	if err != nil {
		t.Fatalf("Blueflood returns unexpected error: %s", err.Error())
	}
	assert.New(t).Contextf("request for timerange").Eq(result, expected)
}

// TestBluefloodHTTPQueriesMultiResolution tests that multiresolution fetching works.
func TestBluefloodHTTPQueriesMultiResolutionSingle(t *testing.T) {
	// from -30d5h to -14d17h
	nowMillis := int64(739908000000)
	nowFunc := TimeSource{GetTime: func() time.Time {
		timeValue := time.Unix(nowMillis/1000, nowMillis%1000*1e6)
		return timeValue
	}}
	offset30 := int64(30 * 24 * 60 * 60 * 1000) // 30 days, TTL on 5m resolution
	offset15 := int64(15 * 24 * 60 * 60 * 1000) // 15 days, first available on 60
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

	type test struct {
		request            timeseries.FetchRequest
		expectedURLs       []string
		expectedTimeseries api.Timeseries
		error              bool
	}
	testClient := mocks.NewFakeHTTPClient()
	// 5 MINUTE RESOLUTION
	testClient.SetResponse("https://blueflood.url/v2.0/square/views/some.key.graphite?from=738612000000&resolution=5MIN&select=numPoints%2Caverage&to=738640799999", mocks.Response{
		Body: fmt.Sprintf(
			`{
			"unit": "unknown",
			"values": [` + func() string {
				result := ""
				for i := 0; i < 8*12; i++ {
					if i != 0 {
						result += ","
					}
					result += fmt.Sprintf(`{
						"numPoints": 1,
						"average": %d,
						"timestamp": %d
					}`, i*i, nowMillis-offset15+int64(i)*5*60*1000)
				}
				return result
			}() + `],
			"metadata": {
				"limit": null,
				"next_href": null,
				"count": 5,
				"marker": null
			}
		}`),
		Delay:      0,
		StatusCode: 200,
	})
	// 60 MINUTE RESOLUTION
	testClient.SetResponse("https://blueflood.url/v2.0/square/views/some.key.graphite?from=737298000000&resolution=60MIN&select=numPoints%2Caverage&to=738611999999", mocks.Response{
		Body: fmt.Sprintf(
			`{
			"unit": "unknown",
			"values": [` + func() string {
				result := ""
				for i := -5; i < 24*15; i++ {
					if i != -5 {
						result += ","
					}
					result += fmt.Sprintf(`{
						"numPoints": 1,
						"average": %d,
						"timestamp": %d
					}`, i*i, nowMillis-offset30+int64(i)*60*60*1000)
				}
				return result
			}() + `],
			"metadata": {
				"limit": null,
				"next_href": null,
				"count": 5,
				"marker": null
			}
		}`),
		Delay:      0,
		StatusCode: 200,
	})
	config := Config{
		BaseURL:                 "https://blueflood.url",
		TenantID:                "square",
		Resolutions:             []Resolution{resolutionFull, resolution5Min, resolution60Min, resolution1440Min},
		MaxSimultaneousRequests: 5,

		GraphiteMetricConverter: &mocks.FakeGraphiteConverter{
			MetricMap: map[util.GraphiteMetric]api.TaggedMetric{
				"some.key.graphite": {
					MetricKey: "some.key",
					TagSet:    api.TagSet{"tag": "value"},
				},
			},
		},

		HTTPClient: testClient,
		TimeSource: nowFunc,
	}
	blueflood := NewBlueflood(config)
	request := timeseries.FetchRequest{
		Metric: api.TaggedMetric{MetricKey: "some.key", TagSet: api.TagSet{"tag": "value"}},
		RequestDetails: timeseries.RequestDetails{
			SampleMethod: timeseries.SampleMean,
			Timerange:    makeRange(30*day+5*time.Hour, 15*day-7*time.Hour, 60*time.Minute),
		},
	}

	values := make([]float64, 5+24*15+7+1)
	for i := 0; i < 5+24*15; i++ {
		values[i] = float64(i-5) * float64(i-5)
	}
	for i := 0; i < 7+1; i++ {
		average := 0.0
		for j := 0; j < 12; j++ {
			average += float64(i*12+j) * float64(i*12+j)
		}
		values[i+5+24*15] = average / 12.0
	}

	expected := api.Timeseries{
		Values: values,
		TagSet: api.TagSet{"tag": "value"},
	}
	result, err := blueflood.FetchSingleTimeseries(request)
	if err != nil {
		t.Fatalf("Blueflood returns unexpected error: %s", err.Error())
	}
	assert.New(t).Contextf("request for timerange").Eq(result, expected)
}
