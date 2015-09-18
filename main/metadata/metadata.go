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
	"os"
	// "net/http"
	// "time"

	// "github.com/square/metrics/log"

	"github.com/square/metrics/api"

	// "github.com/square/metrics/main/common"
	"github.com/square/metrics/metric_metadata/cassandra"
	// "github.com/square/metrics/query"
	// "github.com/square/metrics/ui"
)

func main() {
	flag.Parse()
	// common.SetupLogger()

	// Hosts    []string `yaml:"hosts"`
	// Keyspace string   `yaml:"keyspace"`
	config := cassandra.CassandraMetricMetadataConfig{Hosts: []string{"aws1.medium-trigger.universe.square"}, Keyspace: "metrics_indexer"}
	var metadataAPI api.MetricMetadataAPI
	metadataAPI, err := cassandra.NewCassandraMetricMetadataAPI(config)
	if err != nil {
		fmt.Printf("ERROR %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Success\n")
	stuff, err := metadataAPI.GetAllTags("jvm.thread-states", api.MetricMetadataAPIContext{})
	if err != nil {
		fmt.Printf("ERROR %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Stuff %+v\n", stuff)
}
