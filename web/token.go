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

package web

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/square/metrics/metric_metadata"
	"github.com/square/metrics/query/command"
)

// tokenHandler function and metric name tokens available in the system for the autocomplete.
type tokenHandler struct {
	context command.ExecutionContext
}

func (h tokenHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	metrics, err := h.context.MetricMetadataAPI.GetAllMetrics(metadata.Context{}) // no profiling used
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(encodeError(err))
		return
	}

	response := Response{
		Success: true,
		QueryResponse: QueryResponse{
			Body: map[string]interface{}{ // map to array-like types.
				"functions": h.context.Registry.All(),
				"metrics":   metrics,
			},
		},
	}

	// Make sure the query params have been parsed
	if err := request.ParseForm(); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(encodeError(err))
		return
	}

	pretty, _ := strconv.ParseBool(request.Form.Get("pretty"))
	var encoded []byte
	if pretty {
		encoded, err = json.MarshalIndent(response, "", "  ")
	} else {
		encoded, err = json.Marshal(response)
	}
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(`{"success": false, "message": "Failed to encode the result message."}`))
		return
	}
	writer.Write(encoded)
}
