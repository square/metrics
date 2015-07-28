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
	BaseUrl  string           `yaml:"base_url"`
	TenantId string           `yaml:"tenant_id"`
	Ttls     map[string]int64 `yaml:"ttls"` // Ttl in days
	Timeout  time.Duration    `yaml:"timeout"`
}

func (c Config) getTTL(r Resolution) time.Duration {
	var ttl int64
	if v, ok := c.Ttls[r.bluefloodEnum]; ok {
		ttl = v
	} else {
		// Use blueflood defaults
		switch r {
		case ResolutionFull:
			ttl = 1
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

	return time.Duration(ttl) * 24 * time.Hour
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

type Resolution struct {
	bluefloodEnum string
	duration      time.Duration
}

var (
	ResolutionFull    Resolution = Resolution{"FULL", time.Second * 30}
	Resolution5Min               = Resolution{"MIN5", time.Minute * 5}
	Resolution20Min              = Resolution{"MIN20", time.Minute * 20}
	Resolution60Min              = Resolution{"MIN60", time.Minute * 60}
	Resolution240Min             = Resolution{"MIN240", time.Minute * 240}
	Resolution1440Min            = Resolution{"MIN1440", time.Minute * 1440}
)
var Resolutions []Resolution = []Resolution{
	ResolutionFull,
	Resolution5Min,
	Resolution20Min,
	Resolution60Min,
	Resolution240Min,
	Resolution1440Min,
}

func NewBlueflood(c Config) api.Backend {
	b := blueflood{config: c, client: http.DefaultClient}
	b.config.Ttls = map[string]int64{}
	for k, v := range c.Ttls {
		b.config.Ttls[k] = v
	}
	return &b
}

type sampler struct {
	fieldName     string
	fieldSelector func(point metricPoint) float64
	bucketSampler func([]float64) float64
}

func (b *blueflood) FetchSingleSeries(request api.FetchSeriesRequest) (api.Timeseries, error) {
	sampler, ok := samplerMap[request.SampleMethod]
	if !ok {
		return api.Timeseries{}, fmt.Errorf("unsupported SampleMethod %s", request.SampleMethod.String())
	}
	queryResolution := b.config.bluefloodResolution(
		request.Timerange.Resolution(),
		request.Timerange.Start())

	queryUrl, err := b.constructURL(request, sampler, queryResolution)
	if err != nil {
		return api.Timeseries{}, err
	}

	// Issue GET to fetch metrics
	parsedResult, err := b.fetch(request, queryUrl)
	if err != nil {
		return api.Timeseries{}, err
	}

	values := processResult(parsedResult, request.Timerange, sampler, queryResolution)
	log.Debugf("Constructed timeseries from result: %v", values)

	return api.Timeseries{
		Values: values,
		TagSet: request.Metric.TagSet,
	}, nil
}

// Helper functions
// ----------------

// constructURL creates the URL to the blueflood's backend to fetch the data from.
func (b *blueflood) constructURL(request api.FetchSeriesRequest,
	sampler sampler,
	queryResolution Resolution) (*url.URL, error) {
	graphiteName, err := request.API.ToGraphiteName(request.Metric)
	if err != nil {
		return nil, api.BackendError{request.Metric, api.InvalidSeriesError, "cannot convert to graphite name"}
	}

	result, err := url.Parse(fmt.Sprintf("%s/v2.0/%s/views/%s", b.config.BaseUrl, b.config.TenantId, graphiteName))
	if err != nil {
		return nil, api.BackendError{request.Metric, api.InvalidSeriesError, "cannot generate URL"}
	}

	params := url.Values{}
	params.Set("from", strconv.FormatInt(request.Timerange.Start(), 10))
	// Pull a bit outside of the requested range from blueflood so we
	// have enough data to generate all snapped values
	params.Set("to", strconv.FormatInt(request.Timerange.End()+request.Timerange.ResolutionMillis(), 10))
	params.Set("resolution", queryResolution.bluefloodEnum)
	params.Set("select", fmt.Sprintf("numPoints,%s", strings.ToLower(sampler.fieldName)))
	result.RawQuery = params.Encode()
	return result, nil
}

// fetches from the backend. on error, it returns an instance of api.BackendError
func (b *blueflood) fetch(request api.FetchSeriesRequest, queryUrl *url.URL) (queryResponse, error) {
	log.Debugf("Blueflood fetch: %s", queryUrl.String())
	success := make(chan queryResponse)
	failure := make(chan error)
	timeout := time.After(b.config.Timeout)
	go func() {
		resp, err := b.client.Get(queryUrl.String())
		if err != nil {
			failure <- api.BackendError{request.Metric, api.FetchIOError, "error while fetching - http connection"}
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			failure <- api.BackendError{request.Metric, api.FetchIOError, "error while fetching - reading"}
			return
		}

		log.Debugf("Fetch result: %s", string(body))

		var parsedJson queryResponse
		err = json.Unmarshal(body, &parsedJson)
		// Construct a Timeseries from the result:
		if err != nil {
			failure <- api.BackendError{request.Metric, api.FetchIOError, "error while fetching - json decoding"}
			return
		}
		success <- parsedJson
	}()
	select {
	case response := <-success:
		return response, nil
	case err := <-failure:
		return queryResponse{}, err
	case <-timeout:
		return queryResponse{}, api.BackendError{request.Metric, api.FetchTimeoutError, ""}
	}
}

func processResult(parsedResult queryResponse,
	timerange api.Timerange,
	sampler sampler,
	queryResolution Resolution) []float64 {
	// buckets are each filled with from the points stored in result.Values, according to their timestamps.
	buckets := bucketsFromMetricPoints(parsedResult.Values, sampler.fieldSelector, timerange)

	// values will hold the final values to be returned as the series.
	values := make([]float64, timerange.Slots())

	for i, bucket := range buckets {
		if len(bucket) == 0 {
			values[i] = math.NaN()
			continue
		}
		values[i] = sampler.bucketSampler(bucket)
	}

	// interpolate.
	return values
}

func addMetricPoint(metricPoint metricPoint, field func(metricPoint) float64, timerange api.Timerange, buckets [][]float64) bool {
	value := field(metricPoint)
	// The index to assign within the array is computed using the timestamp.
	// It floors to the nearest index.
	index := (metricPoint.Timestamp - timerange.Start()) / timerange.ResolutionMillis()
	if index < 0 || index >= int64(timerange.Slots()) {
		return false
	}
	buckets[index] = append(buckets[index], value)
	return true
}

func bucketsFromMetricPoints(metricPoints []metricPoint, resultField func(metricPoint) float64, timerange api.Timerange) [][]float64 {
	buckets := make([][]float64, timerange.Slots())
	for _, point := range metricPoints {
		addMetricPoint(point, resultField, timerange, buckets)
	}
	return buckets
}

var samplerMap map[api.SampleMethod]sampler = map[api.SampleMethod]sampler{
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
func (c Config) bluefloodResolution(
	desiredResolution time.Duration,
	startMs int64) Resolution {
	now := time.Now().Unix() * 1000
	// Choose the appropriate resolution based on TTL, fetching the highest resolution data we can
	for _, current := range Resolutions {
		age := time.Duration(now-startMs) * time.Millisecond
		maxAge := c.getTTL(current)
		log.Debugf("Desired (s): %d\n", desiredResolution/time.Second)
		log.Debugf("Current (s): %d\n", current.duration/time.Second)
		log.Debugf("age (s): %d\n", age/time.Second)
		log.Debugf("ttl (s): %d\n", maxAge/time.Second)
		if desiredResolution <= current.duration &&
			age < c.getTTL(current) {
			return current
		}
	}
	// return the coarsest resolution.
	return Resolutions[len(Resolutions)-1]
}

func (b blueflood) DecideTimerange(start int64, end int64, userResolutionMillis int64) (api.Timerange, error) {
	slotLimit := 3000
	resolutions := []int64{
		userResolutionMillis,
		// Try each of these resolutions in order
		// Use the first that's fine-grained enough, if the user-specified value fails
		30 * 1000,
		5 * 60 * 1000,
		60 * 60 * 1000,
		1440 * 60 * 1000,
	}
	for _, resolution := range resolutions {
		timerange, err := api.NewSnappedTimerange(start, end, resolution)
		if err != nil {
			continue
		}
		if timerange.Slots() <= slotLimit {
			return timerange, nil
		}
	}
	return api.Timerange{}, fmt.Errorf("no resolution produced a valid timerange")
}
