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

// Query represents the list of required parameters to perform the
// time-series query to the backend.
type Query struct {
	metric     TaggedMetric
	seriesType SeriesType
	timerange  Timerange
}

// Backend describes how to fetch time-series data
// from a given backend supporting the metrics query engine.
type Backend interface {
	FetchMetadata(metric TaggedMetric) MetricMetadata
	FetchSeries(query Query) SeriesList
}
