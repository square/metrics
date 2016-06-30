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
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"time"
)

var bluefloodAddress = flag.String("blueflood-address", "", "Specify the URL for the HTTP  Blueflood server. [example: http://localhost:19000/v2.0/tenant-id/ingest]")
var mqeIngestionAddress = flag.String("mqe-ingestion-address", "", "Specify the URL (including port) for the MQE Ingestion server [example: localhost:7774]")

type sourceMetric struct {
	Metric string
	Source Source
}

func main() {
	flag.Parse()

	if *bluefloodAddress == "" || *mqeIngestionAddress == "" {
		flag.Usage()
		return
	}

	//

	metrics := []sourceMetric{}

	{
		sources := make([]Source, 5)
		for i := range sources {
			sources[i] = Generate(func(x float64) float64 {
				return math.Max(0, math.Min(300, x+10*rand.NormFloat64()))
			})
		}
		for i := 0; i < 10; i++ {
			metrics = append(metrics, sourceMetric{
				fmt.Sprintf("webserver.host%d.cpu.percentage", i),
				NewLinear(sources),
			})
		}
	}

	{
		sources := make([]Source, 5)
		for i := range sources {
			sources[i] = Generate(func(x float64) float64 {
				return x + math.Floor(math.Max(0, math.Min(300, 20+30*rand.NormFloat64())))
			})
		}
		for i := 0; i < 10; i++ {
			scale := rand.Float64()*90 + 10
			metrics = append(metrics, sourceMetric{
				fmt.Sprintf("webserver.host%d.connection.http.count", i),
				&Mapper{NewLinear(sources), func(x float64) float64 { return scale * math.Floor(x) }}, // MAP
			})
		}
	}

	{
		sources := make([]Source, 5)
		for i := range sources {
			sources[i] = Generate(func(x float64) float64 {
				return math.Max(0, math.Min(1000, 10*rand.NormFloat64()))
			})
		}
		for i := range sources {
			metrics = append(metrics, sourceMetric{
				fmt.Sprintf("webserver.host%d.connection.http.latency", i),
				&Capper{sources[i], 0, 1000},
			})
		}
	}

	client := &http.Client{}
	iteration := 0
	for range time.Tick(15 * time.Second) {
		iteration++
		// every 15 seconds...
		for _, metric := range metrics {
			metric.Source.Advance(iteration)
			err := reportMetric(client, metric.Source.Value(), metric.Metric)
			if err != nil {
				fmt.Printf("Error for metric %s:\n%s\n", metric.Metric, err.Error())
			}
		}
	}
}

func reportMetric(client *http.Client, value float64, name string, options ...interface{}) error {
	metricName := fmt.Sprintf(name, options...)
	json := createPointJSON(value, metricName)

	request, err := http.NewRequest("POST", *bluefloodAddress, bytes.NewBuffer([]byte(json)))
	if err != nil {
		return fmt.Errorf("error creating Blueflood request: %s\n", err.Error())
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("error performing Blueflood POST: %s\n", err.Error())
	}
	fmt.Println("Blueflood: ", response.Status)
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(body))

	// Next, we need to inform the ingestion service

	request, err = http.NewRequest("POST", *mqeIngestionAddress, bytes.NewBuffer([]byte(metricName)))
	if err != nil {
		return fmt.Errorf("error creating MQE Ingestion request: %s\n", err.Error())
	}
	response, err = client.Do(request)
	if err != nil {
		return fmt.Errorf("error performing MQE Ingestion POST: %s\n", err.Error())
	}
	fmt.Println("MQE: ", response.Status)
	body, _ = ioutil.ReadAll(response.Body)
	fmt.Println(string(body))
	return nil
}

func createPointJSON(value float64, metricName string) string {
	// 10 day TTL
	return fmt.Sprintf(`[{
		"collectionTime": %d,
		"ttlInSeconds": 864000,
		"metricValue": %f,
		"metricName": "%s"
	}]`, time.Now().Unix()*1000, value, metricName)
}
