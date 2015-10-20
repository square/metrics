package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/square/metrics/api"
	"github.com/square/metrics/metric_metadata/cassandra"
	"github.com/square/metrics/util"
)

const rulePath = "conversion_rules" // relative to execution

func main() {
	rules, err := util.LoadRules(rulePath)
	if err != nil {
		fmt.Printf("Error loading rules; %+v", err.Error())
		return
	}

	converter := util.RuleBasedGraphiteConverter{Ruleset: rules}

	cassandra, err := cassandra.NewCassandraMetricMetadataAPI(cassandra.CassandraMetricMetadataConfig{
		Hosts:    []string{"127.0.0.1"}, // using the default port
		Keyspace: "metrics_indexer",     // from schema in github.com/square/metrics/schema
	})
	if err != nil {
		fmt.Printf("Error encountered while creating Cassandra API instance: %+v", err.Error())
		return
	}
	fmt.Printf("Successfully connected to the Cassandra database.\n")

	// Now we'll create an HTTP service which can be provided metric names.
	// It will deliver them to Cassandra through this api

	http.HandleFunc("/ingest", func(w http.ResponseWriter, req *http.Request) {
		fmt.Printf("Received request.\n")
		body := req.Body

		bytes, err := ioutil.ReadAll(body)
		if err != nil {
			log.Printf("Error reading body %s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Error reading body: %s", err.Error())))
		}

		metrics := []util.GraphiteMetric{} // TODO: apply conversion rules

		for _, metric := range strings.Split(string(bytes), "\n") {
			metric = strings.TrimSpace(metric)
			if metric != "" {
				metrics = append(metrics, util.GraphiteMetric(metric))
			}
		}

		if len(metrics) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("received no metrics"))
			return
		}

		// Now, convert them

		converted := []api.TaggedMetric{}

		messages := []string{}

		for _, metric := range metrics {
			result, err := converter.ToTaggedName(metric)
			if err != nil {
				messages = append(messages, err.Error())
				continue
			}
			converted = append(converted, result)
		}

		err = cassandra.AddMetrics(converted, api.MetricMetadataAPIContext{})

		if err != nil {
			log.Printf("Error sending metrics to Cassandra: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("error connecting to cassandra %s", err.Error())))
			return
		}
		err = body.Close()
		if err != nil {
			log.Printf("Error closing body: %s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		for _, message := range messages {
			w.Write([]byte("Error: "))
			w.Write([]byte(message))
			w.Write([]byte("\n"))
		}
	})

	err = http.ListenAndServe(":7774", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
