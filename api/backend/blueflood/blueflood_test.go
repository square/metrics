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
	for _, test := range []struct {
		metricMap          map[api.GraphiteMetric]api.TaggedMetric
		queryMetric        api.TaggedMetric
		predicate          api.Predicate
		sampleMethod       api.SampleMethod
		timerange          api.Timerange
		baseUrl            string
		tenantId           string
		queryUrl           string
		queryResponse      string
		expectedSeriesList api.SeriesList
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
			predicate:    nil,
			sampleMethod: api.SampleMean,
			timerange:    api.Timerange{1, 6, 5},
			baseUrl:      "https://blueflood.url",
			tenantId:     "square",
			queryUrl:     "https://blueflood.url/v2.0/square/views/some.key.graphite?from=1000&resolution=FULL&select=numPoints%2Caverage&to=6000",
			queryResponse: `{
        "unit": "unknown", 
        "values": [
          {
            "numPoints": 1,
            "timestamp": 1000,
            "average": 5
          },
          {
            "numPoints": 1,
            "timestamp": 6000,
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
			expectedSeriesList: api.SeriesList{
				Series: []api.Timeseries{
					api.Timeseries{
						Values: []float64{5, 3},
						TagSet: api.TagSet(map[string]string{
							"tag": "value",
						}),
					},
				},
				Timerange: api.Timerange{1, 6, 5},
				Name:      "",
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

		b := NewBlueflood(fakeApi, test.baseUrl, test.tenantId)
		b.client = fakeHttpClient

		seriesList, err := b.FetchSeries(test.queryMetric, test.predicate, test.sampleMethod, test.timerange)
		if err != nil {
			t.Errorf(err.Error())
		}

		a.Eq(seriesList, &test.expectedSeriesList)
	}
}
