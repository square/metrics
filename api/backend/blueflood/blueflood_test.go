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

package blueflood

import (
	"net/http"
	"testing"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/assert"
	"github.com/square/metrics/mocks"
)

func Test_Blueflood(t *testing.T) {
	timerange, err := api.NewTimerange(12000, 13000, 1000)
	if err != nil {
		t.Fatalf("invalid testcase timerange")
		return
	}
	defaultClientConfig := Config{
		BaseUrl:  "https://blueflood.url",
		TenantId: "square",
		Ttls:     make(map[string]int64),
		Timeout:  time.Millisecond,
	}
	// Not really MIN1440, but that's what default TTLs will get with the Timerange we use
	defaultQueryUrl := "https://blueflood.url/v2.0/square/views/some.key.graphite?from=12000&resolution=MIN1440&select=numPoints%2Caverage&to=14000"

	for _, test := range []struct {
		name               string
		metricMap          map[api.GraphiteMetric]api.TaggedMetric
		queryMetric        api.TaggedMetric
		sampleMethod       api.SampleMethod
		timerange          api.Timerange
		clientConfig       Config
		queryUrl           string
		queryResponse      string
		queryResponseCode  int
		queryDelay         time.Duration
		expectedErrorCode  api.BackendErrorCode
		expectedSeriesList api.Timeseries
	}{
		{
			name: "Success case",
			metricMap: map[api.GraphiteMetric]api.TaggedMetric{
				api.GraphiteMetric("some.key.graphite"): api.TaggedMetric{
					MetricKey: api.MetricKey("some.key"),
					TagSet:    api.ParseTagSet("tag=value"),
				},
			},
			queryMetric: api.TaggedMetric{
				MetricKey: api.MetricKey("some.key"),
				TagSet:    api.ParseTagSet("tag=value"),
			},
			sampleMethod: api.SampleMean,
			timerange:    timerange,
			queryUrl:     defaultQueryUrl,
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
				TagSet: api.ParseTagSet("tag=value"),
			},
		},
		{
			name: "Failure case - invalid JSON",
			metricMap: map[api.GraphiteMetric]api.TaggedMetric{
				api.GraphiteMetric("some.key.graphite"): api.TaggedMetric{
					MetricKey: api.MetricKey("some.key"),
					TagSet:    api.ParseTagSet("tag=value"),
				},
			},
			queryMetric: api.TaggedMetric{
				MetricKey: api.MetricKey("some.key"),
				TagSet:    api.ParseTagSet("tag=value"),
			},
			sampleMethod:      api.SampleMean,
			timerange:         timerange,
			clientConfig:      defaultClientConfig,
			queryUrl:          defaultQueryUrl,
			queryResponse:     `{invalid}`,
			expectedErrorCode: api.FetchIOError,
		},
		{
			name: "Failure case - HTTP error",
			metricMap: map[api.GraphiteMetric]api.TaggedMetric{
				api.GraphiteMetric("some.key.graphite"): api.TaggedMetric{
					MetricKey: api.MetricKey("some.key"),
					TagSet:    api.ParseTagSet("tag=value"),
				},
			},
			queryMetric: api.TaggedMetric{
				MetricKey: api.MetricKey("some.key"),
				TagSet:    api.ParseTagSet("tag=value"),
			},
			sampleMethod:      api.SampleMean,
			timerange:         timerange,
			clientConfig:      defaultClientConfig,
			queryUrl:          defaultQueryUrl,
			queryResponse:     `{}`,
			queryResponseCode: 400,
			expectedErrorCode: api.FetchIOError,
		},
		{
			name: "Failure case - timeout",
			metricMap: map[api.GraphiteMetric]api.TaggedMetric{
				api.GraphiteMetric("some.key.graphite"): api.TaggedMetric{
					MetricKey: api.MetricKey("some.key"),
					TagSet:    api.ParseTagSet("tag=value"),
				},
			},
			queryMetric: api.TaggedMetric{
				MetricKey: api.MetricKey("some.key"),
				TagSet:    api.ParseTagSet("tag=value"),
			},
			sampleMethod:      api.SampleMean,
			timerange:         timerange,
			clientConfig:      defaultClientConfig,
			queryUrl:          defaultQueryUrl,
			queryResponse:     `{}`,
			queryDelay:        1 * time.Second,
			expectedErrorCode: api.FetchTimeoutError,
		},
	} {
		a := assert.New(t).Contextf("%s", test.name)

		fakeApi := mocks.NewFakeApi()
		for k, v := range test.metricMap {
			fakeApi.AddPair(v, k)
		}

		fakeHttpClient := mocks.NewFakeHttpClient()
		code := test.queryResponseCode
		if code == 0 {
			code = http.StatusOK
		}
		fakeHttpClient.SetResponse(test.queryUrl, mocks.Response{test.queryResponse, test.queryDelay, code})

		b := NewBlueflood(test.clientConfig).(*blueflood)
		b.client = fakeHttpClient

		seriesList, err := b.FetchSingleSeries(api.FetchSeriesRequest{
			Metric:       test.queryMetric,
			SampleMethod: test.sampleMethod,
			Timerange:    test.timerange,
			API:          fakeApi,
			Cancellable:  api.NewCancellable(),
		})

		if test.expectedErrorCode != 0 {
			if err == nil {
				a.Errorf("Expected error, but was successful.")
				continue
			}
			berr, ok := err.(api.BackendError)
			if !ok {
				a.Errorf("Failed to cast error to BackendError")
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

func TestSeriesFromMetricPoints(t *testing.T) {
	timerange, err := api.NewTimerange(4000, 4800, 100)
	if err != nil {
		t.Fatalf("testcase timerange is invalid")
		return
	}
	points := []metricPoint{
		{
			Timestamp: 4100,
			Average:   1,
		},
		{
			Timestamp: 4299, // Test flooring behavior
			Average:   2,
		},
		{
			Timestamp: 4403, // Test flooring behavior
			Average:   3,
		},
		{
			Timestamp: 4500,
			Average:   4,
		},
		{
			Timestamp: 4700,
			Average:   5,
		},
		{
			Timestamp: 4749,
			Average:   6,
		},
	}
	expected := [][]float64{{}, {1}, {2}, {}, {3}, {4}, {}, {5, 6}, {}}
	result := bucketsFromMetricPoints(points, func(point metricPoint) float64 { return point.Average }, timerange)
	if len(result) != len(expected) {
		t.Fatalf("Expected %+v but got %+v", expected, result)
		return
	}
	for i, expect := range expected {
		if len(result[i]) != len(expect) {
			t.Fatalf("Exected %+v but got %+v", expected, result)
			return
		}
		for j := range expect {
			if result[i][j] != expect[j] {
				t.Fatalf("Expected %+v but got %+v", expected, result)
				return
			}
		}
	}
}
