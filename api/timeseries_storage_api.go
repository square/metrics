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

package api

type TimeseriesStorageAPI interface {
	// FetchSingleSeries should return an instance of BackendError
	FetchSingleSeries(request FetchTimeseriesRequest) (Timeseries, error)
	FetchMultipleSeries(request FetchMultipleTimeseriesRequest) (SeriesList, error)
}

type FetchTimeseriesRequest struct {
	Metric       TaggedMetric // metric to fetch.
	SampleMethod SampleMethod // up/downsampling behavior.
	Timerange    Timerange    // time range to fetch data from.
	API          API          // an API instance.
	Cancellable  Cancellable
	// Profiler     *inspect.Profiler
}

type FetchMultipleTimeseriesRequest struct {
	Metrics      []TaggedMetric
	SampleMethod SampleMethod
	Timerange    Timerange
	API          API
	Cancellable  Cancellable
	// Profiler     *inspect.Profiler
}

// func (r FetchMultipleTimeseriesRequest) ToSingle(metric TaggedMetric) FetchSeriesRequest {
// 	return FetchTimeseriesRequest{
// 		Metric:       metric,
// 		API:          r.API,
// 		Cancellable:  r.Cancellable,
// 		SampleMethod: r.SampleMethod,
// 		Timerange:    r.Timerange,
// 		Profiler:     r.Profiler,
// 	}
// }
