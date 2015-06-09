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
	"reflect"
	"strconv"
	"strings"

	"github.com/square/metrics/api"
)

type httpClient interface {
	// our own client to mock out the standard golang HTTP Client.
	Get(url string) (resp *http.Response, err error)
}

type Blueflood struct {
	api      api.API
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

func NewBlueflood(api api.API, baseUrl string, tenantId string) *Blueflood {
	return &Blueflood{api: api, baseUrl: baseUrl, tenantId: tenantId, client: http.DefaultClient}
}

func (b *Blueflood) Api() api.API {
	return b.api
}

func (b *Blueflood) FetchSeries(metric api.TaggedMetric, predicate api.Predicate, sampleMethod api.SampleMethod, timerange api.Timerange) (*api.SeriesList, error) {
	metricTagSets, err := b.Api().GetAllTags(metric.MetricKey)
	if err != nil {
		return nil, err
	}

	// TODO: Be a little smarter about this?
	resultSeries := []api.Timeseries{}
	for _, ts := range metricTagSets {
		if predicate == nil || predicate.Apply(ts) {
			graphiteName, err := b.Api().ToGraphiteName(api.TaggedMetric{
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

func (b *Blueflood) fetchSingleSeries(metric api.GraphiteMetric, sampleMethod api.SampleMethod, timerange api.Timerange) ([]float64, error) {
	// Use this lowercase of this as the select query param. Use the actual value
	// to reflect into result MetricPoints to fetch the correct field.
	var selectResultField string
	switch sampleMethod {
	case api.SampleMean:
		selectResultField = "Average"
	case api.SampleMin:
		selectResultField = "Min"
	case api.SampleMax:
		selectResultField = "Max"
	default:
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
	params.Set("to", strconv.FormatInt(timerange.End(), 10))
	params.Set("resolution", bluefloodResolution(timerange.Resolution()))
	params.Set("select", fmt.Sprintf("numPoints,%s", strings.ToLower(selectResultField)))

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

	// Construct a Timeseries from the result

	values := make([]float64, timerange.Slots())
	// The array is initialized to NaN.
	for i := range values {
		values[i] = math.NaN()
	}

	for _, metricPoint := range result.Values {
		// TODO: check the result of .FieldByName against nil (or a panic will occur)
		value := reflect.ValueOf(metricPoint).FieldByName(selectResultField).Float()
		// TODO: check the result of .FieldByName against nil (or a panic will occur)
		timestamp := reflect.ValueOf(metricPoint).FieldByName("Timestamp").Int()
		// The index to assign within the array is computed using the timestamp.
		// It rounds to the nearest index.
		index := (timestamp - timerange.Start() + timerange.Resolution()/2) / timerange.Resolution()
		if index >= 0 && index < int64(timerange.Slots()) {
			values[index] = value
		}
		// TODO: error otherwise?
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
