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
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"github.com/square/metrics/function/registry"
	"github.com/square/metrics/main/common"
	"github.com/square/metrics/metric_metadata/cassandra"
	"github.com/square/metrics/query/command"
	"github.com/square/metrics/query/parser"
	"github.com/square/metrics/timeseries"
	"github.com/square/metrics/timeseries/blueflood"
	"github.com/square/metrics/util"
)

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
	}{}

	common.LoadConfigs(&config)

	cassandraAPI, err := cassandra.NewMetricMetadataAPI(config.Cassandra)
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

	//Defaults
	userConfig := timeseries.UserSpecifiableConfig{
		IncludeRawData: false,
	}

	executionContext := command.ExecutionContext{
		MetricMetadataAPI:     cassandraAPI,
		TimeseriesStorageAPI:  blueflood,
		FetchLimit:            1500,
		SlotLimit:             5000,
		Registry:              registry.Default(),
		UserSpecifiableConfig: userConfig,
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("> ")
		query, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Fatal error reading input: %s\n", err.Error())
			break
		}
		command, err := parser.Parse(query)
		if err != nil {
			fmt.Printf("Parse error: %s\n", err.Error())
			continue
		}
		result, err := command.Execute(executionContext)
		if err != nil {
			fmt.Printf("Execution error: %s\n", err.Error())
			continue
		}
		encoded, err := json.MarshalIndent(result.Body, "", "  ")
		if err != nil {
			fmt.Printf("Error encoding json: %s\n", err.Error())
			continue
		}
		fmt.Println(string(encoded))
	}
}
