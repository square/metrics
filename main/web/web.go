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
	"github.com/square/metrics/log"
	"github.com/square/metrics/main/common"
	"github.com/square/metrics/metric_metadata"
	"github.com/square/metrics/metric_metadata/cached"
	"github.com/square/metrics/metric_metadata/cassandra"
	"github.com/square/metrics/query/command"
	"github.com/square/metrics/timeseries"
	"github.com/square/metrics/timeseries/blueflood"
	"github.com/square/metrics/util"
	"github.com/square/metrics/web"
)

func startServer(config web.Config, context command.ExecutionContext) error {
	// @@ leaking param: config
	// @@ leaking param: context
	httpMux, err := web.NewMux(config, context, web.Hook{})
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: httpMux,
		// @@ config.Port escapes to heap
		ReadTimeout: time.Duration(config.Timeout) * time.Second,
		// @@ httpMux escapes to heap
		WriteTimeout:   time.Duration(config.Timeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	// @@ &http.Server literal escapes to heap
	fmt.Printf("Listening on port %d.\n", config.Port)
	return server.ListenAndServe()
	// @@ config.Port escapes to heap
}

func main() {
	//Adding a signal handler to dump goroutines
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR2)
	// @@ make(chan os.Signal, 1) escapes to heap
	// @@ make(chan os.Signal, 1) escapes to heap

	// @@ syscall.SIGUSR2 escapes to heap
	go func() {
		for range sigs {
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
		}
		// @@ os.Stdout escapes to heap
	}()

	config := struct {
		ConversionRulesPath string           `yaml:"conversion_rules_path"`
		Cassandra           cassandra.Config `yaml:"cassandra"`
		Blueflood           blueflood.Config `yaml:"blueflood"`
		Web                 web.Config       `yaml:"web"`
	}{}

	// @@ moved to heap: config
	common.LoadConfig(&config)

	// @@ &config escapes to heap
	// @@ &config escapes to heap
	metadataAPI, err := cassandra.NewMetricMetadataAPI(config.Cassandra)
	if err != nil {
		common.ExitWithErrorMessage("Error loading Cassandra API: %s", err.Error())
		return
		// @@ err.Error() escapes to heap
	}

	ruleset, err := util.LoadRules(config.ConversionRulesPath)
	if err != nil {
		common.ExitWithErrorMessage("Error loading conversion rules: %s", err.Error())
		return
		// @@ err.Error() escapes to heap
	}

	config.Blueflood.GraphiteMetricConverter = &util.RuleBasedGraphiteConverter{Ruleset: ruleset}

	// @@ &util.RuleBasedGraphiteConverter literal escapes to heap
	// @@ &util.RuleBasedGraphiteConverter literal escapes to heap
	// @@ &util.RuleBasedGraphiteConverter literal escapes to heap
	// @@ &util.RuleBasedGraphiteConverter literal escapes to heap
	// @@ &util.RuleBasedGraphiteConverter literal escapes to heap
	// @@ &util.RuleBasedGraphiteConverter literal escapes to heap
	// @@ &util.RuleBasedGraphiteConverter literal escapes to heap
	// @@ &util.RuleBasedGraphiteConverter literal escapes to heap
	// @@ &util.RuleBasedGraphiteConverter literal escapes to heap
	// @@ &util.RuleBasedGraphiteConverter literal escapes to heap
	blueflood := blueflood.NewBlueflood(config.Blueflood)

	// @@ inlining call to blueflood.NewBlueflood
	// @@ blueflood.b·3 escapes to heap
	// @@ &blueflood.Blueflood literal escapes to heap
	// @@ http.DefaultClient escapes to heap
	optimizedMetadataAPI := cached.NewMetricMetadataAPI(metadataAPI, cached.Config{
		TimeToLive:   time.Minute * 5, // Cache items invalidated after 5 minutes.
		RequestLimit: 500,
	})
	for i := 0; i < 10; i++ {
		// @@ inlining call to cached.NewMetricMetadataAPI
		// @@ &cached.metricUpdateAPI literal escapes to heap
		// @@ &cached.metricUpdateAPI literal escapes to heap
		// @@ metadataAPI escapes to heap
		// @@ composite literal escapes to heap
		// @@ map[api.MetricKey]*cached.CachedTagSetList literal escapes to heap
		// @@ make(chan func(metadata.Context) error, cached.config·3.RequestLimit) escapes to heap
		// @@ &cached.result·5 escapes to heap
		// @@ moved to heap: cached.result·5
		// @@ &cached.result·5 escapes to heap
		// @@ &cached.metricUpdateAPI literal escapes to heap
		// @@ &cached.metricUpdateAPI literal escapes to heap
		// @@ metadataAPI escapes to heap
		// @@ composite literal escapes to heap
		// @@ map[api.MetricKey]*cached.CachedTagSetList literal escapes to heap
		// @@ make(chan func(metadata.Context) error, cached.config·3.RequestLimit) escapes to heap
		// @@ &cached.result·5 escapes to heap
		// @@ &cached.result·5 escapes to heap
		// @@ &cached.metricUpdateAPI literal escapes to heap
		// @@ &cached.metricUpdateAPI literal escapes to heap
		// @@ metadataAPI escapes to heap
		// @@ composite literal escapes to heap
		// @@ map[api.MetricKey]*cached.CachedTagSetList literal escapes to heap
		// @@ make(chan func(metadata.Context) error, cached.config·3.RequestLimit) escapes to heap
		// @@ &cached.result·5 escapes to heap
		// @@ &cached.result·5 escapes to heap
		// Start goroutines to update the metadata cache in the background.
		go func() {
			for {
				// @@ func literal escapes to heap
				// @@ func literal escapes to heap
				err := optimizedMetadataAPI.GetBackgroundAction()(metadata.Context{})
				if err != nil {
					// @@ leaking closure reference optimizedMetadataAPI
					log.Errorf("Error performing background cache-update: %s", err.Error())
				}
				// @@ ... argument escapes to heap
				// @@ err.Error() escapes to heap
			}
		}()
	}

	//Defaults
	userConfig := timeseries.UserSpecifiableConfig{
		IncludeRawData: false,
	}

	err = startServer(config.Web, command.ExecutionContext{
		MetricMetadataAPI:    optimizedMetadataAPI,
		TimeseriesStorageAPI: blueflood,
		// @@ optimizedMetadataAPI escapes to heap
		FetchLimit:            1500,
		SlotLimit:             5000,
		Registry:              registry.Default(),
		UserSpecifiableConfig: userConfig,
		// @@ inlining call to registry.Default
		// @@ registry.Default() escapes to heap
	})
	if err != nil {
		log.Infof(err.Error())
	}
}
