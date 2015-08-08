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

// Integration test for the query execution.
package query

import (
	"testing"

	"github.com/square/metrics/api"
	"github.com/square/metrics/api/backend"
	"github.com/square/metrics/assert"
)

type fakeGraphiteAPI struct {
	api.API
}

func (a fakeGraphiteAPI) ToGraphiteName(metric api.TaggedMetric) (api.GraphiteMetric, error) {
	if graphiteMetric, ok := metric.TagSet["$graphite"]; ok {
		return api.GraphiteMetric(graphiteMetric), nil
	}
	return a.API.ToGraphiteName(metric)
}

func (a fakeGraphiteAPI) AddGraphiteMetric(api.GraphiteMetric) error {
	return nil
}
func (a fakeGraphiteAPI) GetAllGraphiteMetrics() ([]api.GraphiteMetric, error) {
	metrics := make([]api.GraphiteMetric, 0, len(graphiteMap))
	for metric := range graphiteMap {
		metrics = append(metrics, metric)
	}
	return metrics, nil
}

type fakeGraphiteBackend struct {
}

var graphiteMap = map[api.GraphiteMetric][]float64{
	"server.north.cpu.mean":   {1, 2, 3},
	"server.north.cpu.median": {4, 5, 6},
	"server.south.cpu.mean":   {7, 8, 9},
	"server.south.cpu.median": {2, 4, 8},
	"proxy.south.cpu.mean":    {1, 1, 1},
	"proxy.south.cpu.median":  {2, 2, 2},
	"latency.host45.http":     {3, 3, 3},
	"latency.host68.http":     {4, 4, 4},
	"latency.host68.rpc":      {5, 5, 5},
	"latency.host22.rpc":      {6, 6, 6},
}

func (b fakeGraphiteBackend) FetchSingleSeries(f api.FetchSeriesRequest) (api.Timeseries, error) {
	graphiteName, err := f.API.ToGraphiteName(f.Metric)
	if err != nil {
		return api.Timeseries{}, err
	}
	return api.Timeseries{TagSet: f.Metric.TagSet, Values: graphiteMap[graphiteName]}, nil
}

func TestGraphite(t *testing.T) {
	fakeAPI := fakeGraphiteAPI{}
	fakeBackend := backend.NewSequentialMultiBackend(fakeGraphiteBackend{})
	tests := []struct {
		query    string
		expected []api.Timeseries
	}{
		{
			query: "graphite('server.north.cpu.mean') from -10m to now",
			expected: []api.Timeseries{
				{
					Values: []float64{1, 2, 3},
					TagSet: api.TagSet{},
				},
			},
		},
		{
			query: "graphite('server.north.cpu.%quantity%') from -10m to now",
			expected: []api.Timeseries{
				{
					Values: []float64{1, 2, 3},
					TagSet: api.TagSet{
						"quantity": "mean",
					},
				},
				{
					Values: []float64{4, 5, 6},
					TagSet: api.TagSet{
						"quantity": "median",
					},
				},
			},
		},
		{
			query: "graphite('%app%.%dc%.cpu.%quantity%') from -10m to now",
			expected: []api.Timeseries{
				{
					Values: []float64{1, 2, 3},
					TagSet: api.TagSet{
						"app":      "server",
						"dc":       "north",
						"quantity": "mean",
					},
				},
				{
					Values: []float64{4, 5, 6},
					TagSet: api.TagSet{
						"app":      "server",
						"dc":       "north",
						"quantity": "median",
					},
				},
				{
					Values: []float64{7, 8, 9},
					TagSet: api.TagSet{
						"app":      "server",
						"dc":       "south",
						"quantity": "mean",
					},
				},
				{
					Values: []float64{2, 4, 8},
					TagSet: api.TagSet{
						"app":      "server",
						"dc":       "south",
						"quantity": "median",
					},
				},
				{
					Values: []float64{1, 1, 1},
					TagSet: api.TagSet{
						"app":      "proxy",
						"dc":       "south",
						"quantity": "mean",
					},
				},
				{
					Values: []float64{2, 2, 2},
					TagSet: api.TagSet{
						"app":      "proxy",
						"dc":       "south",
						"quantity": "median",
					},
				},
			},
		},
		{
			query: "graphite('latency.%host%.%method%') from -10m to now",
			expected: []api.Timeseries{
				{
					Values: []float64{3, 3, 3},
					TagSet: api.TagSet{
						"host":   "host45",
						"method": "http",
					},
				},
				{
					Values: []float64{4, 4, 4},
					TagSet: api.TagSet{
						"host":   "host68",
						"method": "http",
					},
				},
				{
					Values: []float64{5, 5, 5},
					TagSet: api.TagSet{
						"host":   "host68",
						"method": "rpc",
					},
				},
				{
					Values: []float64{6, 6, 6},
					TagSet: api.TagSet{
						"host":   "host22",
						"method": "rpc",
					},
				},
			},
		},
		{
			query: "graphite('latency.%host%.%method%') where method = 'rpc' from -10m to now",
			expected: []api.Timeseries{
				{
					Values: []float64{5, 5, 5},
					TagSet: api.TagSet{
						"host":   "host68",
						"method": "rpc",
					},
				},
				{
					Values: []float64{6, 6, 6},
					TagSet: api.TagSet{
						"host":   "host22",
						"method": "rpc",
					},
				},
			},
		},
	}
	for _, test := range tests {
		command, err := Parse(test.query)
		if err != nil {
			t.Fatalf("Unexpected error while parsing")
			return
		}
		if command.Name() != "select" {
			t.Errorf("Expected select command but got %s", command.Name())
			continue
		}
		rawResult, err := command.Execute(ExecutionContext{Backend: fakeBackend, API: fakeAPI, FetchLimit: 1000, Timeout: 0})
		if err != nil {
			t.Errorf("Unexpected error while execution: %s", err.Error())
			continue
		}
		seriesListList, ok := rawResult.([]api.SeriesList)
		if !ok || len(seriesListList) != 1 {
			t.Errorf("expected query `%s` to produce []value; got %+v :: %T", test.query, rawResult, rawResult)
			continue
		}
		list := seriesListList[0]
		if err != nil {
			t.Fatal(err)
		}
		a := assert.New(t)
		expectedSeries := test.expected
		a.EqInt(len(list.Series), len(expectedSeries))
		lookup := map[string][]float64{}
		for _, expect := range expectedSeries {
			lookup[expect.TagSet.Serialize()] = expect.Values
		}
		a.EqString(list.Name, "$graphite")
		for _, series := range list.Series {
			a.EqFloatArray(series.Values, lookup[series.TagSet.Serialize()], 1e-100)
		}
	}
}

// Verify that we filter before fetching
func TestGraphiteLimits(t *testing.T) {
	fakeAPI := fakeGraphiteAPI{}
	fakeBackend := backend.NewSequentialMultiBackend(fakeGraphiteBackend{})
	tests := []struct {
		query    string
		expected []api.Timeseries
	}{
		{
			query: "graphite('%app%.%dc%.cpu.%quantity%') where app = 'server' and dc = 'north' from -10m to now",
			expected: []api.Timeseries{
				{
					Values: []float64{1, 2, 3},
					TagSet: api.TagSet{
						"app":      "server",
						"dc":       "north",
						"quantity": "mean",
					},
				},
				{
					Values: []float64{4, 5, 6},
					TagSet: api.TagSet{
						"app":      "server",
						"dc":       "north",
						"quantity": "median",
					},
				},
			},
		},
	}
	for _, test := range tests {
		command, err := Parse(test.query)
		if err != nil {
			t.Fatalf("Unexpected error while parsing")
			return
		}
		if command.Name() != "select" {
			t.Errorf("Expected select command but got %s", command.Name())
			continue
		}
		rawResult, err := command.Execute(ExecutionContext{Backend: fakeBackend, API: fakeAPI, FetchLimit: 2, Timeout: 0})
		if err != nil {
			t.Errorf("Unexpected error while execution: %s", err.Error())
			continue
		}
		seriesListList, ok := rawResult.([]api.SeriesList)
		if !ok || len(seriesListList) != 1 {
			t.Errorf("expected query `%s` to produce []value; got %+v :: %T", test.query, rawResult, rawResult)
			continue
		}
		list := seriesListList[0]
		a := assert.New(t)
		expectedSeries := test.expected
		lookup := map[string][]float64{}
		for _, expect := range expectedSeries {
			lookup[expect.TagSet.Serialize()] = expect.Values
		}
		a.EqString(list.Name, "$graphite")
		for _, series := range list.Series {
			a.EqFloatArray(series.Values, lookup[series.TagSet.Serialize()], 1e-100)
		}
	}
}

func TestGraphiteFailure(t *testing.T) {
	fakeAPI := fakeGraphiteAPI{}
	fakeBackend := backend.NewSequentialMultiBackend(fakeGraphiteBackend{})
	tests := []string{
		"graphite('') from -10m to now",
		"graphite('does.not.exist') from -10m to now",
		"graphite('latency.%host%.%method%s') where host = 'hostDNE' from -10m to now",
		"graphite('latency.%host%.%invalid') where host = 'hostDNE' from -10m to now",
		"graphite('latency.%host%.invalid%') where host = 'hostDNE' from -10m to now",
	}
	for _, test := range tests {
		command, err := Parse(test)
		if err != nil {
			t.Fatalf("Unexpected error while parsing: `%s`", test)
		}
		_, err = command.Execute(ExecutionContext{Backend: fakeBackend, API: fakeAPI, FetchLimit: 1000, Timeout: 0})
		if err == nil {
			t.Fatalf("Expected query `%s` to fail", test)
		}
	}
	// These tests make sure that the Fetch limit is being exercized
	smallTests := []string{
		"graphite('%app%.%dc%.cpu.%quantity%') from -10m to now",
		"graphite('latency.%host%.%method%') from -10m to now",
	}
	for _, test := range smallTests {
		command, err := Parse(test)
		if err != nil {
			t.Fatalf("Unexpected error while parsing: `%s`", test)
		}
		_, err = command.Execute(ExecutionContext{Backend: fakeBackend, API: fakeAPI, FetchLimit: 2, Timeout: 0})
		if err == nil {
			t.Fatalf("Expected query `%s` to fail", test)
		}
	}
}
