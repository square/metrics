// Copyright 2015 - 2016 Square Inc.
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

package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/square/metrics/api"
	"github.com/square/metrics/metric_metadata"
)

// tokenHandler exposes all the tokens available in the system for the autocomplete.
type ingestHandler struct {
	metricMetadataAPI metadata.MetricUpdateAPI
}

type IngestRequest struct {
	Name string            `json:"name"`
	Tags map[string]string `json:"tags"`
}

func (h ingestHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	if request.Header.Get("Content-Type") != "application/json" {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(encodeError(fmt.Errorf("index endpoint expects Content-Type: application/json")))
		return
	}
	metrics := []IngestRequest{}
	if err := json.NewDecoder(request.Body).Decode(&metrics); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(encodeError(err))
		return
	}
	taggedMetrics := []api.TaggedMetric{}
	for i := range metrics {
              //Append will create taggedMetrics and add to metrics range
              taggedMetrics = append(taggedMetrics,api.TaggedMetric{MetricKey: api.MetricKey(metrics[i].Name),TagSet: metrics[i].Tags})
	}
	err := h.metricMetadataAPI.AddMetrics(taggedMetrics, metadata.Context{})
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(encodeError(err))
		return
	}
	writer.Write([]byte(`{"success": true}`))
}
