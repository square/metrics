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

// Backend describes how to fetch time-series data from a given backend.
type FetchSeriesRequest struct {
	Metric       TaggedMetric
	SampleMethod SampleMethod
	Timerange    Timerange
	Api          API
}

type Backend interface {
	// FetchSingleSeries fetches the series described by the provided TaggedMetric
	// corresponding to the Timerange, down/upsampling if necessary using
	// SampleMethod
	FetchSingleSeries(request FetchSeriesRequest) (Timeseries, error)
}
