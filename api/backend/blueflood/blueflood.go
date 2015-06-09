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
	// Pull a bit outside of the requested range from blueflood so we
	// have enough data to generate all snapped values.
	params.Set("from", strconv.FormatInt(timerange.Start(), 10))
	params.Set("to", strconv.FormatInt(timerange.End()+timerange.Resolution(), 10))
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

	series, err := resample(result.Values, timerange, sampleMethod)
	if err == nil {
		log.Printf("Constructed timeseries from result: %v", series)
	}

	return series, err
}

// Blueflood keys the resolution param to a java enum, so we have to convert
// between them.
// TODO: This is currently only based on the requested resolution to minimize
// the amount of data we work on. We'll need to be aware of blueflood's
// retention policies for each resolution as well.
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

// Resamples the original data returned from Blueflood. Returns raw values
// corresponding to the `timerange`.
//
// Snaps data left to the `timerange` buckets. Resampling buckets using data
// to the right of the timestamp (i.e. the data point at time 123000 with
// resolution 60 is generated from [123000, 123060). If there is no data, that
// point is filled with NaN (note: this means upsampling doesn't work).
//
// Assumes `points` is sorted chronologically.
func resample(points []MetricPoint, timerange api.Timerange, sampleMethod api.SampleMethod) ([]float64, error) {
	series := make([]float64, timerange.Slots())

	// Cursor into `series`
	idx := 0
	// Cursors into `points` for working data windows
	windowStart, windowEnd := 0, 0
	// Generate points at timestamps corresponding to the timerange
	for t := timerange.Start(); t <= timerange.End(); t, idx = t+timerange.Resolution(), idx+1 {
		// Set up our series window
		for ; windowStart < len(points)-1 && points[windowStart].Timestamp < t; windowStart++ {
		}
		// No data points in [t, t + resolution) becomes NaN
		if points[windowStart].Timestamp >= (t + timerange.Resolution()) {
			series[idx] = math.NaN()
			continue
		}
		for windowEnd = windowStart; windowEnd < len(points)-1 && points[windowEnd+1].Timestamp < (t+timerange.Resolution()); windowEnd++ {
		}

		switch sampleMethod {
		case api.SampleMean:
			// Uses weighted average over [t, t + resolution)
			totalPoints, runningTotal := 0, 0.0
			for i := windowStart; i <= windowEnd; i++ {
				totalPoints += points[i].Points
				runningTotal += points[i].Average * float64(points[i].Points)
			}

			series[idx] = runningTotal / float64(totalPoints)

		case api.SampleMax:
			return nil, errors.New("Not implemented")
		case api.SampleMin:
			return nil, errors.New("Not implemented")
		default:
			return nil, errors.New(fmt.Sprintf("Unsupported SampleMethod %d", sampleMethod))
		}
	}

	return series, nil
}

var _ api.Backend = (*Blueflood)(nil)
