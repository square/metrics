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

package timeseries_map

import (
	"math"
	"time"

	"github.com/square/metrics/api"
)

type Point struct {
	Time  time.Time
	Value float64
}

type PointMap struct {
	Points []Point
}

type PointDatabase struct {
	Map map[string]PointMap
}

func (p PointDatabase) ChooseResolution(requested api.Timerange, smallestResolution time.Duration) time.Duration {
	return requested.Resolution()
}
func (p PointDatabase) fetch(request api.FetchTimeseriesRequest) []float64 {
	result := make([]float64, request.Timerange.Slots())
	for i := range result {
		result[i] = math.NaN()
	}
	for _, point := range p.Map[request.Metric].Points {
		i := request.Timerange.Index(point.Time)
		if 0 <= i && i < len(result) {
			result[i] = point.Value
		}
	}
	return result
}
func (p PointDatabase) FetchSingleTimeseries(request api.FetchTimeseriesRequest) ([]float64, error) {
	return p.fetch(request), nil
}
func (p PointDatabase) FetchMultipleTimeseries(requests api.FetchMultipleTimeseriesRequest) ([][]float64, error) {
	list := requests.ToSingle()
	result := make([][]float64, len(list))
	for i, request := range list {
		result[i] = p.fetch(request)
	}
	return result, nil
}
