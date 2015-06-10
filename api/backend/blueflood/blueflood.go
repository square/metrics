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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/square/metrics/api"
)

type httpClient interface {
	// our own client to mock out the standard golang HTTP Client.
	Get(url string) (resp *http.Response, err error)
}

type Blueflood struct {
	baseUrl  string
	tenantId string
	client   httpClient
}

type QueryResponse struct {
	Values []MetricPoint `json:"values"`
}

type MetricPoint struct {
	Points    int     `json:"numPoints"`
	Timestamp int64   `json:"timestamp"`
	Average   float64 `json:"average"`
	Max       float64 `json:"max"`
	Min       float64 `json:"min"`
	Variance  float64 `json:"variance"`
}

const (
	ResolutionFull    = "FULL"
	Resolution5Min    = "MIN5"
	Resolution20Min   = "MIN20"
	Resolution60Min   = "MIN60"
	Resolution240Min  = "MIN240"
	Resolution1440Min = "MIN1440"
)

func NewBlueflood(baseUrl string, tenantId string) *Blueflood {
	return &Blueflood{baseUrl: baseUrl, tenantId: tenantId, client: http.DefaultClient}
}

func (b *Blueflood) FetchSeries(request api.FetchSeriesRequest) (*api.SeriesList, error) {
	metric := request.Metric
	predicate := request.Predicate
	sampleMethod := request.SampleMethod
	timerange := request.Timerange
	metricTagSets, err := request.Api.GetAllTags(metric.MetricKey)
	if err != nil {
		return nil, err
	}

	// TODO: Be a little smarter about this?
	resultSeries := []api.Timeseries{}
	for _, ts := range metricTagSets {
		if predicate == nil || predicate.Apply(ts) {
			graphiteName, err := request.Api.ToGraphiteName(api.TaggedMetric{
				MetricKey: metric.MetricKey,
				TagSet:    ts,
			})
			if err != nil {
				return nil, err
			}

			queryResult, err := b.fetchSingleSeries(
				graphiteName,
				sampleMethod,
				timerange,
			)
			// TODO: Be more tolerant of errors fetching a single metric?
			// Though I guess this behavior is fine since skipping fetches
			// that fail would end up in a result set that you don't quite
			// expect.
			if err != nil {
				return nil, err
			}

			resultSeries = append(resultSeries, api.Timeseries{
				Values: queryResult,
				TagSet: ts,
			})
		}
	}

	return &api.SeriesList{
		Series:    resultSeries,
		Timerange: timerange,
	}, nil
}

func addMetricPoint(metricPoint MetricPoint, field func(MetricPoint) float64, timerange api.Timerange, buckets [][]float64) bool {
	value := field(metricPoint)
	// The index to assign within the array is computed using the timestamp.
	// It floors to the nearest index.
	index := (metricPoint.Timestamp - timerange.Start()) / timerange.Resolution()
	if index < 0 || index >= int64(timerange.Slots()) {
		return false
	}
	buckets[index] = append(buckets[index], value)
	return true
}

func bucketsFromMetricPoints(metricPoints []MetricPoint, resultField func(MetricPoint) float64, timerange api.Timerange) [][]float64 {
	buckets := make([][]float64, timerange.Slots())
	// Make the buckets:
	for i := range buckets {
		buckets[i] = []float64{}
	}
	for _, point := range metricPoints {
		addMetricPoint(point, resultField, timerange, buckets)
	}
	return buckets
}

var samplerMap map[api.SampleMethod]struct {
	fieldName     string
	fieldSelector func(point MetricPoint) float64
	bucketSampler func([]float64) float64
} = map[api.SampleMethod]struct {
	fieldName     string
	fieldSelector func(point MetricPoint) float64
	bucketSampler func([]float64) float64
}{
	api.SampleMean: {
		fieldName:     "average",
		fieldSelector: func(point MetricPoint) float64 { return point.Average },
		bucketSampler: func(bucket []float64) float64 {
			value := 0.0
			for _, v := range bucket {
				value += v
			}
			return value / float64(len(bucket))
		},
	},
	api.SampleMin: {
		fieldName:     "min",
		fieldSelector: func(point MetricPoint) float64 { return point.Min },
		bucketSampler: func(bucket []float64) float64 {
			value := bucket[0]
			for _, v := range bucket {
				value = math.Min(value, v)
			}
			return value
		},
	},
	api.SampleMax: {
		fieldName:     "max",
		fieldSelector: func(point MetricPoint) float64 { return point.Max },
		bucketSampler: func(bucket []float64) float64 {
			value := bucket[0]
			for _, v := range bucket {
				value = math.Max(value, v)
			}
			return value
		},
	},
}

func (b *Blueflood) fetchSingleSeries(metric api.GraphiteMetric, sampleMethod api.SampleMethod, timerange api.Timerange) ([]float64, error) {
	sampler, ok := samplerMap[sampleMethod]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Unsupported SampleMethod %d", sampleMethod))
	}

	// Issue GET to fetch metrics
	queryUrl, err := url.Parse(fmt.Sprintf("%s/v2.0/%s/views/%s",
		b.baseUrl,
		b.tenantId,
		metric))
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("from", strconv.FormatInt(timerange.Start(), 10))
	// Pull a bit outside of the requested range from blueflood so we
	// have enough data to generate all snapped values
	params.Set("to", strconv.FormatInt(timerange.End()+timerange.Resolution(), 10))
	params.Set("resolution", bluefloodResolution(timerange.Resolution()))
	params.Set("select", fmt.Sprintf("numPoints,%s", strings.ToLower(sampler.fieldName)))

	queryUrl.RawQuery = params.Encode()

	log.Printf("Blueflood fetch: %s", queryUrl.String())
	resp, err := b.client.Get(queryUrl.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("Fetch result: %s", string(body))

	var result QueryResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	// Construct a Timeseries from the result:

	// buckets are each filled with from the points stored in result.Values, according to their timestamps.
	buckets := bucketsFromMetricPoints(result.Values, sampler.fieldSelector, timerange)

	// values will hold the final values to be returned as the series.
	values := make([]float64, timerange.Slots())

	for i, bucket := range buckets {
		if len(bucket) == 0 {
			values[i] = math.NaN()
			continue
		}
		values[i] = sampler.bucketSampler(bucket)
	}

	log.Printf("Constructed timeseries from result: %v", values)

	// TODO: Resample to the requested resolution

	return values, nil
}

// Blueflood keys the resolution param to a java enum, so we have to convert
// between them.
func bluefloodResolution(r int64) string {
	switch {
	case r < 5*60*1000:
		return ResolutionFull
	case r < 20*60*1000:
		return Resolution5Min
	case r < 60*60*1000:
		return Resolution20Min
	case r < 240*60*1000:
		return Resolution60Min
	case r < 1440*60*1000:
		return Resolution240Min
	}
	return Resolution1440Min
}

func resample(points []float64, currentResolution int64, expectedTimerange api.Timerange, sampleMethod api.SampleMethod) ([]float64, error) {
	return nil, errors.New("Not implemented")
}

var _ api.Backend = (*Blueflood)(nil)
