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
	"fmt"
	"net/http"

	"github.com/square/metrics/metric_metadata"
	"github.com/square/metrics/query/command"
)

func NewMux(config Config, context command.ExecutionContext, hook Hook) (*http.ServeMux, error) {
	// Wrap the given API and Backend in their Profiling counterparts.
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "/ui", http.StatusTemporaryRedirect)
	})
	httpMux.Handle("/ui", singleStaticHandler{config.StaticDir, "index.html"})
	httpMux.Handle("/embed", singleStaticHandler{config.StaticDir, "embed.html"})
	httpMux.Handle("/query", queryHandler{
		context: context,
		hook:    hook,
	})
	httpMux.Handle("/token", tokenHandler{
		context: context,
	})
	if config.HTTPIngestion {
		if updateAPI, ok := context.MetricMetadataAPI.(metadata.MetricUpdateAPI); ok {
			httpMux.Handle("/ingest", ingestHandler{
				metricMetadataAPI: updateAPI,
			})
		} else {
			return nil, fmt.Errorf("HTTP Ingestion is on, but the metadata API does not implement updates.")
		}
	}
	httpMux.Handle(
		"/static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(http.Dir(config.StaticDir)),
		),
	)
	return httpMux, nil
}
