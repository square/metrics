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

package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/square/metrics/function/registry"
	"github.com/square/metrics/inspect/log"
	common "github.com/square/metrics/main"
	"github.com/square/metrics/main/web/server"
	"github.com/square/metrics/metric_metadata"
	"github.com/square/metrics/metric_metadata/cached"
	"github.com/square/metrics/metric_metadata/cassandra"
	"github.com/square/metrics/query/command"
	"github.com/square/metrics/timeseries"
	"github.com/square/metrics/timeseries/blueflood"
	"github.com/square/metrics/util"
)

func startServer(config server.Config, context command.ExecutionContext) error {
	httpMux, err := server.NewMux(config, context, server.Hook{})
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", config.Port),
		Handler:        httpMux,
		ReadTimeout:    time.Duration(config.Timeout) * time.Second,
		WriteTimeout:   time.Duration(config.Timeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Printf("Listening on port %d.\n", config.Port)
	return server.ListenAndServe()
}

func main() {
	//Adding a signal handler to dump goroutines
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR2)

	go func() {
		for range sigs {
			pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
		}
	}()

	config := struct {
		ConversionRulesPath string           `yaml:"conversion_rules_path"`
		Cassandra           cassandra.Config `yaml:"cassandra"`
		Blueflood           blueflood.Config `yaml:"blueflood"`
		Web                 server.Config    `yaml:"web"`
	}{}

	common.LoadConfig(&config)

	metadataAPI, err := cassandra.NewMetricMetadataAPI(config.Cassandra)
	if err != nil {
		common.ExitWithErrorMessage("Error loading Cassandra API: %s", err.Error())
		return
	}

	ruleset, err := util.LoadRules(config.ConversionRulesPath)
	if err != nil {
		common.ExitWithErrorMessage("Error loading conversion rules: %s", err.Error())
		return
	}

	config.Blueflood.GraphiteMetricConverter = &util.RuleBasedGraphiteConverter{Ruleset: ruleset}

	blueflood := blueflood.NewBlueflood(config.Blueflood)

	optimizedMetadataAPI := cached.NewMetricMetadataAPI(metadataAPI, cached.Config{
		TimeToLive:   time.Minute * 5, // Cache items invalidated after 5 minutes.
		RequestLimit: 500,
	})
	for i := 0; i < 10; i++ {
		// Start goroutines to update the metadata cache in the background.
		go func() {
			for {
				err := optimizedMetadataAPI.GetBackgroundAction()(metadata.Context{})
				if err != nil {
					log.Errorf("Error performing background cache-update: %s", err.Error())
				}
			}
		}()
	}

	//Defaults
	userConfig := timeseries.UserSpecifiableConfig{
		IncludeRawData: false,
	}

	err = startServer(config.Web, command.ExecutionContext{
		MetricMetadataAPI:     optimizedMetadataAPI,
		TimeseriesStorageAPI:  blueflood,
		FetchLimit:            1500,
		SlotLimit:             5000,
		Registry:              registry.Default(),
		UserSpecifiableConfig: userConfig,
	})
	if err != nil {
		log.Infof(err.Error())
	}
}
