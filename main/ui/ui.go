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

package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/square/metrics/log"

	"github.com/square/metrics/api"
	"github.com/square/metrics/api/backend"
	"github.com/square/metrics/api/backend/blueflood"
	"github.com/square/metrics/function/registry"
	"github.com/square/metrics/main/common"
	"github.com/square/metrics/query"
	"github.com/square/metrics/ui"
)

func startServer(config common.UIConfig, context query.ExecutionContext) {
	httpMux := ui.NewMux(config.Config, context, ui.Hook{})

	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", config.Port),
		Handler:        httpMux,
		ReadTimeout:    time.Duration(config.Timeout) * time.Second,
		WriteTimeout:   time.Duration(config.Timeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Infof(err.Error())
	}
}

func main() {
	flag.Parse()
	common.SetupLogger()

	config := common.LoadConfig()

	apiInstance := common.NewAPI(config.API)

	blueflood := api.ProfilingBackend{
		Backend: blueflood.NewBlueflood(config.Blueflood),
	}
	backend := api.ProfilingMultiBackend{
		MultiBackend: backend.NewParallelMultiBackend(blueflood, 20),
	}

	startServer(config.UIConfig, query.ExecutionContext{
		API: apiInstance, Backend: backend, FetchLimit: 1000,
		Registry: registry.Default(),
	})
}
