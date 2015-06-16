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
	"testing"

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
	for _, test := range []struct {
		metricMap          map[api.GraphiteMetric]api.TaggedMetric
		queryMetric        api.TaggedMetric
		sampleMethod       api.SampleMethod
		timerange          api.Timerange
		baseUrl            string
		tenantId           string
		queryUrl           string
		queryResponse      string
		expectedSeriesList api.Timeseries
	}{
		{
			metricMap: map[api.GraphiteMetric]api.TaggedMetric{
				api.GraphiteMetric("some.key.graphite"): api.TaggedMetric{
					MetricKey: api.MetricKey("some.key"),
					TagSet: api.TagSet(map[string]string{
						"tag": "value",
					}),
				},
			},
			queryMetric: api.TaggedMetric{
				MetricKey: api.MetricKey("some.key"),
				TagSet: api.TagSet(map[string]string{
					"tag": "value",
				}),
			},
			sampleMethod: api.SampleMean,
			timerange:    timerange,
			baseUrl:      "https://blueflood.url",
			tenantId:     "square",
			// Not really MIN1440, but that's what default TTLs will get with the Timerange we use
			queryUrl: "https://blueflood.url/v2.0/square/views/some.key.graphite?from=12000&resolution=MIN1440&select=numPoints%2Caverage&to=14000",
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
				TagSet: api.TagSet(map[string]string{
					"tag": "value",
				}),
			},
		},
	} {
		a := assert.New(t)

		fakeApi := mocks.NewFakeApi()
		for k, v := range test.metricMap {
			fakeApi.AddPair(v, k)
		}

		fakeHttpClient := mocks.NewFakeHttpClient()
		fakeHttpClient.SetResponse(test.queryUrl, test.queryResponse)

		b := NewBlueflood(Config{
			BaseUrl:  test.baseUrl,
			TenantId: test.tenantId,
			Ttls:     make(map[Resolution]int64),
		})
		b.client = fakeHttpClient

		seriesList, err := b.FetchSingleSeries(api.FetchSeriesRequest{
			test.queryMetric, test.sampleMethod, test.timerange,
			fakeApi,
		})
		if err != nil {
			a.CheckError(err)
			continue
		}

		a.Eq(seriesList, test.expectedSeriesList)
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
