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
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/log"
)

type httpClient interface {
	// our own client to mock out the standard golang HTTP Client.
	Get(url string) (resp *http.Response, err error)
}

type Config struct {
	BaseUrl  string               `yaml:"base_url"`
	TenantId string               `yaml:"tenant_id"`
	Ttls     map[Resolution]int64 `yaml:"ttls"` // Ttl in days
}

func (c Config) getTtlInMillis(r Resolution) int64 {
	var ttl int64
	if v, ok := c.Ttls[r]; ok {
		ttl = v
	} else {
		// Use blueflood defaults
		switch r {
		case ResolutionFull:
			ttl = 7
		case Resolution5Min:
			ttl = 30
		case Resolution20Min:
			ttl = 60
		case Resolution60Min:
			ttl = 90
		case Resolution240Min:
			ttl = 180
		case Resolution1440Min:
			ttl = 365
		default:
			// Not a supported resolution by blueflood. No real way to recover if
			// someone's trying to fetch ttl for an invalid resolution.
			panic(fmt.Sprintf("invalid resolution `%s`", r))
		}
	}

	return ttl * 24 * 60 * 60 * 1000
}

type blueflood struct {
	config Config
	client httpClient
}

type queryResponse struct {
	Values []metricPoint `json:"values"`
}

type metricPoint struct {
	Points    int     `json:"numPoints"`
	Timestamp int64   `json:"timestamp"`
	Average   float64 `json:"average"`
	Max       float64 `json:"max"`
	Min       float64 `json:"min"`
	Variance  float64 `json:"variance"`
}

type Resolution string

const (
	ResolutionFull    Resolution = "FULL"
	Resolution5Min               = "MIN5"
	Resolution20Min              = "MIN20"
	Resolution60Min              = "MIN60"
	Resolution240Min             = "MIN240"
	Resolution1440Min            = "MIN1440"
)

func NewBlueflood(c Config) *blueflood {
	b := blueflood{config: c, client: http.DefaultClient}

	b.config.Ttls = map[Resolution]int64{}
	for k, v := range c.Ttls {
		b.config.Ttls[k] = v
	}

	return &b
}

func (b *blueflood) FetchSingleSeries(request api.FetchSeriesRequest) (api.Timeseries, error) {
	sampler, ok := samplerMap[request.SampleMethod]
	if !ok {
		return api.Timeseries{}, fmt.Errorf("unsupported SampleMethod %s", request.SampleMethod.String())
	}

	graphiteName, err := request.Api.ToGraphiteName(request.Metric)

	if err != nil {
		return api.Timeseries{}, api.BackendError{request.Metric, api.InvalidSeriesError, "cannot convert to graphite name"}
	}

	// Issue GET to fetch metrics
	queryUrl, err := url.Parse(fmt.Sprintf("%s/v2.0/%s/views/%s",
		b.config.BaseUrl,
		b.config.TenantId,
		graphiteName))

	if err != nil {
		return api.Timeseries{}, api.BackendError{request.Metric, api.InvalidSeriesError, "cannot generate URL"}
	}

	params := url.Values{}
	params.Set("from", strconv.FormatInt(request.Timerange.Start(), 10))
	// Pull a bit outside of the requested range from blueflood so we
	// have enough data to generate all snapped values
	params.Set("to", strconv.FormatInt(request.Timerange.End()+request.Timerange.Resolution(), 10))
	params.Set("resolution", b.config.bluefloodResolution(request.Timerange.Resolution(), request.Timerange.Start()))
	params.Set("select", fmt.Sprintf("numPoints,%s", strings.ToLower(sampler.fieldName)))

	queryUrl.RawQuery = params.Encode()

	log.Infof("Blueflood fetch: %s", queryUrl.String())
	resp, err := b.client.Get(queryUrl.String())
	if err != nil {
		return api.Timeseries{}, api.BackendError{request.Metric, api.FetchIOError, "error while fetching - http connection"}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return api.Timeseries{}, api.BackendError{request.Metric, api.FetchIOError, "error while fetching - reading"}
	}

	log.Infof("Fetch result: %s", string(body))

	var result queryResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return api.Timeseries{}, api.BackendError{request.Metric, api.FetchIOError, "error while fetching - json decoding"}
	}

	// Construct a Timeseries from the result:

	// buckets are each filled with from the points stored in result.Values, according to their timestamps.
	buckets := bucketsFromMetricPoints(result.Values, sampler.fieldSelector, request.Timerange)

	// values will hold the final values to be returned as the series.
	values := make([]float64, request.Timerange.Slots())

	for i, bucket := range buckets {
		if len(bucket) == 0 {
			values[i] = math.NaN()
			continue
		}
		values[i] = sampler.bucketSampler(bucket)
	}

	log.Infof("Constructed timeseries from result: %v", values)

	return api.Timeseries{
		Values: values,
		TagSet: request.Metric.TagSet,
	}, nil
}

func addMetricPoint(metricPoint metricPoint, field func(metricPoint) float64, timerange api.Timerange, buckets [][]float64) bool {
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

func bucketsFromMetricPoints(metricPoints []metricPoint, resultField func(metricPoint) float64, timerange api.Timerange) [][]float64 {
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
	fieldSelector func(point metricPoint) float64
	bucketSampler func([]float64) float64
} = map[api.SampleMethod]struct {
	fieldName     string
	fieldSelector func(point metricPoint) float64
	bucketSampler func([]float64) float64
}{
	api.SampleMean: {
		fieldName:     "average",
		fieldSelector: func(point metricPoint) float64 { return point.Average },
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
		fieldSelector: func(point metricPoint) float64 { return point.Min },
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
		fieldSelector: func(point metricPoint) float64 { return point.Max },
		bucketSampler: func(bucket []float64) float64 {
			value := bucket[0]
			for _, v := range bucket {
				value = math.Max(value, v)
			}
			return value
		},
	},
}

// Blueflood keys the resolution param to a java enum, so we have to convert
// between them.
func (c Config) bluefloodResolution(resolution int64, timestamp int64) string {
	now := time.Now().UTC().Unix() * 1000

	// Choose the appropriate resolution based on TTL, fetching the highest resolution data we can
	switch {
	case resolution < 5*60*1000 && now-timestamp < c.getTtlInMillis(ResolutionFull):
		return string(ResolutionFull)
	case resolution < 20*60*1000 && now-timestamp < c.getTtlInMillis(Resolution5Min):
		return string(Resolution5Min)
	case resolution < 60*60*1000 && now-timestamp < c.getTtlInMillis(Resolution20Min):
		return string(Resolution20Min)
	case resolution < 240*60*1000 && now-timestamp < c.getTtlInMillis(Resolution60Min):
		return string(Resolution60Min)
	case resolution < 1440*60*1000 && now-timestamp < c.getTtlInMillis(Resolution240Min):
		return string(Resolution240Min)
	}
	return string(Resolution1440Min)
}
