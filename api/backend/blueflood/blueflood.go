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
	"net/http"
	"reflect"
	"strings"

	"github.com/golang/glog"
	"github.com/square/metrics/api"
)

type Blueflood struct {
	api      api.API
	baseUrl  string
	tenantId string
}

type QueryResponse struct {
	Values []MetricPoint `json:"values"`
}

type MetricPoint struct {
	Points    int     `json:"numPoints"`
	Timestamp int64   `json:"timestamp`
	Average   float64 `json:"average"`
	Max       float64 `json:"max"`
	Min       float64 `json:"min"`
	Variance  float64 `json:"variance"`
}

func NewBlueflood(api api.API, baseUrl string, tenantId string) *Blueflood {
	return &Blueflood{api: api, baseUrl: baseUrl, tenantId: tenantId}
}

func (b *Blueflood) Api() api.API {
	return b.api
}

func (b *Blueflood) FetchSeries(metric api.TaggedMetric, predicate api.Predicate, sampleMethod api.SampleMethod, timerange api.Timerange) (*api.SeriesList, error) {
	graphiteMetric, err := b.api.ToGraphiteName(metric)
	if err != nil {
		return nil, err
	}

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
		return nil, errors.New(fmt.Sprintf("Unsupported SampleMethod %s", sampleMethod))
	}

	// Issue GET to fetch metrics
	queryUrl := fmt.Sprintf(
		"%s/v2.0/%s/views/%s?from=%d&to=%d&resolution=%s&select=numPoints,%s",
		b.baseUrl,
		b.tenantId,
		graphiteMetric,
		timerange.Start*1000,
		timerange.End*1000,
		bluefloodResolution(timerange.Resolution),
		strings.ToLower(selectResultField))
	glog.V(2).Infof("Blueflood fetch: %s", queryUrl)
	resp, err := http.Get(queryUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	glog.V(2).Infof("Fetch result: %s", string(body))

	var result QueryResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	// Construct a Timeseries from the result
	series := make([]float64, len(result.Values))
	for i, metricPoint := range result.Values {
		series[i] = reflect.ValueOf(metricPoint).FieldByName(selectResultField).Float()
	}

	glog.V(2).Infof("Constructed timeseries from result: %v", series)

	// TODO: Resample to the requested resolution

	return &api.SeriesList{
		Series: []api.Timeseries{
			api.Timeseries{
				Values: series,
				Metric: metric,
			},
		},
		Timerange: timerange,
	}, nil
}

// Blueflood keys the resolution param to a java enum, so we have to convert
// between them.
func bluefloodResolution(r int64) string {
	switch {
	case r < 5*60:
		return "full"
	case r < 20*60:
		return "min5"
	case r < 60*60:
		return "min20"
	case r < 240*60:
		return "min60"
	case r < 1440*60:
		return "min240"
	}
	return "min1440"
}

func resample(points []float64, currentResolution int64, expectedTimerange api.Timerange, sampleMethod api.SampleMethod) ([]float64, error) {
	return nil, errors.New("Not implemented")
}

var _ api.Backend = (*Blueflood)(nil)
