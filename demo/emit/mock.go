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

var bluefloodAddress = flag.String("blueflood-address", "", "Specify the URL for the HTTP  Blueflood server. ex: [http://localhost:19000/v2.0/tenant-id/ingest]")
var mqeIngestionAddress = flag.String("mqe-ingestion-address", "", "Specify the URL (including port) for the MQE Ingestion server.")

type SourceMetric struct {
	Metric string
	Source Source
}

func main() {
	flag.Parse()

	if *bluefloodAddress == "" {
		fmt.Printf("You must specify the blueflood address with `-blueflood-address`\n")
		return
	}
	if *mqeIngestionAddress == "" {
		fmt.Printf("You must specify the MQE ingestion address address with `-mqe-ingestion-address`\n")
		return
	}

	//

	metrics := []SourceMetric{}

	{
		sources := make([]Source, 5)
		for i := range sources {
			sources[i] = Generate(func(x float64) float64 {
				return math.Max(0, math.Min(300, x+10*rand.NormFloat64()))
			})
		}
		for i := 0; i < 10; i++ {
			metrics = append(metrics, SourceMetric{
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
			metrics = append(metrics, SourceMetric{
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
			metrics = append(metrics, SourceMetric{
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

	//
	/*
		serverCount := 200

		cpuPercentage := make([]float64, serverCount)
		cpuVelocities := make([]float64, serverCount)
		cpuQuota := make([]float64, serverCount)

		apps := []string{"mqe", "example", "webserver", "blueflood"}

		for i := 0; i < serverCount; i++ {
			cpuQuota[i] = float64(rand.Intn(6)*10 + 50)
			cpuPercentage[i] = cpuQuota[i] * rand.Float64()
			cpuVelocities[i] = rand.Float64() * 4
		}

		client := &http.Client{}

		for {
			// Report data to both Blueflood and MQE ingestion.
			for i := range cpuPercentage {
				err := reportMetric(client, cpuPercentage[i], "%s.server%d.cpu.percentage", apps[i%len(apps)], i/len(apps))
				if err != nil {
					fmt.Printf("Error\n", err.Error())
				}
			}
			for i := range cpuQuota {
				err := reportMetric(client, cpuQuota[i], "%s.server%d.cpu.quota", apps[i%len(apps)], i/len(apps))
				if err != nil {
					fmt.Printf("Error\n", err.Error())
				}
			}
			// Simulate changing CPU usage (with randomness + sinusoidal trends)
			appBias := make([]float64, len(apps))
			for i := range appBias {
				appBias[i] = rand.NormFloat64()
			}
			for i := range cpuPercentage {
				cpuVelocities[i] *= 0.99
				cpuVelocities[i] += appBias[i%len(apps)]
				cpuVelocities[i] += rand.NormFloat64() * 0.25

				cpuPercentage[i] *= 0.99

				cpuPercentage[i] += cpuVelocities[i]
				if cpuPercentage[i] < 0 {
					cpuPercentage[i] = 0
				}
				if cpuPercentage[i] > cpuQuota[i]*1.05 {
					cpuPercentage[i] = cpuQuota[i] * 1.05
				}
			}
			<-time.After(15 * time.Second)
		}
	*/
}

func reportMetric(client *http.Client, value float64, name string, options ...interface{}) error {
	metricName := fmt.Sprintf(name, options...)
	json := createPointJSON(value, metricName)

	request, err := http.NewRequest("POST", *bluefloodAddress, bytes.NewBuffer([]byte(json)))
	if err != nil {
		return fmt.Errorf("Error creating Blueflood request: %s\n", err.Error())
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("Error performing Blueflood POST: %s\n", err.Error())
	}
	fmt.Println("Blueflood: ", response.Status)
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(body))

	// Next, we need to inform the ingestion service

	request, err = http.NewRequest("POST", *mqeIngestionAddress, bytes.NewBuffer([]byte(metricName)))
	if err != nil {
		return fmt.Errorf("Error creating MQE Ingestion request: %s\n", err.Error())
	}
	response, err = client.Do(request)
	if err != nil {
		return fmt.Errorf("Error performing MQE Ingestion POST: %s\n", err.Error())
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
