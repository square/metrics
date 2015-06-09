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
	"math"
	"testing"

	"github.com/square/metrics/api"
	"github.com/square/metrics/assert"
	"github.com/square/metrics/mocks"
)

func Test_Blueflood(t *testing.T) {
	for _, test := range []struct {
		metricMap           map[api.GraphiteMetric]api.TaggedMetric
		queryMetric         api.TaggedMetric
		predicate           api.Predicate
		sampleMethod        api.SampleMethod
		timerangeStart      int64
		timerangeEnd        int64
		timerangeResolution int64
		baseUrl             string
		tenantId            string
		queryUrl            string
		queryResponse       string
		expectedSeriesList  api.SeriesList
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
			predicate:           nil,
			sampleMethod:        api.SampleMean,
			timerangeStart:      1000,
			timerangeEnd:        6000,
			timerangeResolution: 1000,
			baseUrl:             "https://blueflood.url",
			tenantId:            "square",
			queryUrl:            "https://blueflood.url/v2.0/square/views/some.key.graphite?from=1000&resolution=FULL&select=numPoints%2Caverage&to=7000",
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
						Values: []float64{5, math.NaN(), math.NaN(), math.NaN(), math.NaN(), 3},
						TagSet: api.TagSet(map[string]string{
							"tag": "value",
						}),
					},
				},
				// Set this from the built timerange
				Timerange: api.Timerange{},
				Name:      "",
			},
		},
	} {
		// Setup
		a := assert.New(t)

		fakeApi := mocks.NewFakeApi()
		for k, v := range test.metricMap {
			fakeApi.AddPair(v, k)
		}

		fakeHttpClient := mocks.NewFakeHttpClient()
		fakeHttpClient.SetResponse(test.queryUrl, test.queryResponse)

		b := NewBlueflood(test.baseUrl, test.tenantId)
		b.client = fakeHttpClient

		timerange, err := api.NewTimerange(test.timerangeStart, test.timerangeEnd, test.timerangeResolution)
		if err != nil {
			a.CheckError(err)
			continue
		}

		test.expectedSeriesList.Timerange = *timerange

		// Do the fetch
		seriesList, err := b.FetchSeries(api.FetchSeriesRequest{
			test.queryMetric, test.predicate, test.sampleMethod, *timerange,
			fakeApi,
		})
		if err != nil {
			a.CheckError(err)
			continue
		}

		// Do deep comparisons of the SeriesList ourselves because NaN != NaN
		a.Eq(seriesList.Timerange, test.expectedSeriesList.Timerange)
		a.Eq(seriesList.Name, test.expectedSeriesList.Name)
		a.EqInt(len(seriesList.Series), len(test.expectedSeriesList.Series))
		for i := 0; i < len(seriesList.Series); i++ {
			actualSeries := seriesList.Series[i]
			expectedSeries := test.expectedSeriesList.Series[i]

			a.Eq(actualSeries.TagSet, expectedSeries.TagSet)
			a.EqInt(len(actualSeries.Values), len(expectedSeries.Values))
			for j := 0; j < len(actualSeries.Values); j++ {
				a.EqFloat(actualSeries.Values[j], expectedSeries.Values[j])
			}
		}
	}
}
