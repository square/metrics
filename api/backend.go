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

// MetricMetadata is a metadata for metrics defined in a backend,
// describing the capabilities of a metric exposed by the given backend.
type MetricMetadata struct {
	Meta map[SeriesType]SeriesMetadata
}

// SeriesMetadata is a metadata about a single time series.
type SeriesMetadata struct {
	Resolutions []Timerange // list of available resolutions for the list of time ranges.
}

// Backend describes how to fetch time-series data
// from a given backend supporting the metrics query engine.
// This a MQE level driver to interact with different backends.
type Backend interface {
	FetchMetadata(metric TaggedMetric) MetricMetadata
	FetchSeries(query Query) SeriesList
}
