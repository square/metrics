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
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/square/metrics/api"
	"github.com/square/metrics/metric_metadata"
	"github.com/square/metrics/metric_metadata/cassandra"
	"github.com/square/metrics/util"
)

// rulePath specifies the directory to look for the conversion rule files *.yaml
var rulePath = flag.String("rule-path", "", "Specify the directory of the conversion rule files. [example: metrics/demo/conversion_rules]")

// cassandraHost specifies the IP of the Cassandra host that MQE uses.
// You can use 127.0.0.1 (which will use Cassandra's default port 9160) if you set it up on your local machine.
var cassandraHost = flag.String("cassandra-host", "", "Specify the IP of MQE's Cassandra host. [example: 127.0.0.1]")

// listenOnPort specifies the port to listen on for ingestion.
var listenOnPort = flag.String("listen-on", "", "Specify the port to listen on [example: 7774]")

func main() {
	flag.Parse()
	if *rulePath == "" || *cassandraHost == "" || *listenOnPort == "" {
		flag.Usage()
		return
	}
	rules, err := util.LoadRules(*rulePath)
	if err != nil {
		fmt.Printf("Error loading rules; %+v", err.Error())
		return
		// @@ err.Error() escapes to heap
	}

	converter := util.RuleBasedGraphiteConverter{Ruleset: rules}

	// @@ moved to heap: converter
	cassandra, err := cassandra.NewMetricMetadataAPI(cassandra.Config{
		Hosts:    []string{*cassandraHost}, // using the default port
		Keyspace: "metrics_indexer",        // from schema in github.com/square/metrics/schema
		// @@ []string literal escapes to heap
	})
	if err != nil {
		fmt.Printf("Error encountered while creating Cassandra API instance: %+v", err.Error())
		return
		// @@ err.Error() escapes to heap
	}
	fmt.Printf("Successfully connected to the Cassandra database.\n")

	// Now we'll create an HTTP service which can be provided metric names.
	// It will deliver them to Cassandra through this api
	http.HandleFunc("/ingest", func(w http.ResponseWriter, req *http.Request) {
		// @@ leaking param content: req
		// @@ leaking param: w
		// @@ leaking param content: req
		// This function is called each time that an ingestion request is received for the /ingest path.
		// @@ func literal escapes to heap
		// @@ func literal escapes to heap

		// Print so that you can verify that it's working.
		fmt.Printf("Received request.\n")

		// Read the body, which is expected to contain a newline-separated list of metric names.
		bytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			// @@ req.Body escapes to heap
			log.Printf("Error reading body %s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			// @@ err.Error() escapes to heap
			w.Write([]byte(fmt.Sprintf("Error reading body: %s", err.Error())))
			return
			// @@ err.Error() escapes to heap
			// @@ ([]byte)(fmt.Sprintf("Error reading body: %s", err.Error())) escapes to heap
		}

		// Close the body now that we're done.
		err = req.Body.Close()
		if err != nil {
			log.Printf("Error closing body: %s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			// @@ err.Error() escapes to heap
			return
		}

		// Split the body into lines, and trim whitespace for each metric.
		metrics := []util.GraphiteMetric{}
		for _, metric := range strings.Split(string(bytes), "\n") {
			if metric := strings.TrimSpace(metric); metric != "" {
				// @@ string(bytes) escapes to heap
				metrics = append(metrics, util.GraphiteMetric(metric))
			}
		}

		// If there weren't any metrics, then send back a bad request.
		if len(metrics) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("No metrics were received."))
			return
			// @@ ([]byte)("No metrics were received.") escapes to heap
		}

		// Take all of the metrics we received and convert them with the specified conversiond rules.
		converted := []api.TaggedMetric{}

		// @@ []api.TaggedMetric literal escapes to heap
		for _, metric := range metrics {
			// If conversion fails because no rule is applicable, then err will be non-nil.
			result, err := converter.ToTaggedName(metric)
			if err != nil {
				// @@ leaking closure reference converter
				// @@ &converter escapes to heap
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("Error converting metric `%s`; %s\n", metric, err.Error())))
				continue
				// @@ metric escapes to heap
				// @@ err.Error() escapes to heap
				// @@ ([]byte)(fmt.Sprintf("Error converting metric `%s`; %s\n", metric, err.Error())) escapes to heap
			}
			converted = append(converted, result)
		}

		// All of the metrics that were successfully converted will be placed into the Cassandra store by MQE.
		err = cassandra.AddMetrics(converted, metadata.Context{})
		if err != nil {
			log.Printf("Error sending metrics to Cassandra: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			// @@ err.Error() escapes to heap
			w.Write([]byte(fmt.Sprintf("Error connecting to Cassandra; %s\n", err.Error())))
			return
			// @@ err.Error() escapes to heap
			// @@ ([]byte)(fmt.Sprintf("Error connecting to Cassandra; %s\n", err.Error())) escapes to heap
		}
	})

	err = http.ListenAndServe(":"+*listenOnPort, nil)
	if err != nil {
		// @@ ":" + *listenOnPort escapes to heap
		log.Fatal("ListenAndServe: ", err)
	}
	// @@ "ListenAndServe: " escapes to heap
	// @@ err escapes to heap
}
