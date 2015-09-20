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

// program which takes
// - a rule file
// - a sample list of metrics
// and sees how well the rule performs against the metrics.
package main

import (
	"bytes"
	"compress/zlib"
	"encoding/gob"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"sync"

	"github.com/square/metrics/api"
	"github.com/square/metrics/main/common"
	"github.com/square/metrics/util"
)

var (
	metricsFile = flag.String("metrics-file", "", "Location of zlib compressed gob string file.")
	rulePath    = flag.String("rule-path", "", "Path to directory containing conversion rules.")
)

// Statistics represents the aggregated result of rules
// after running through the test file.
type Statistics struct {
	perMetric map[api.MetricKey]PerMetricStatistics
	matched   int // number of matched rows
	unmatched int // number of unmatched rows
}

// PerMetricStatistics represents per-metric result of rules
// after running through the test file.
type PerMetricStatistics struct {
	matched          int // number of matched rows
	reverseSuccess   int // number of reversed entries
	reverseError     int // number of incorrectly reversed entries.
	reverseIncorrect int // number of incorrectly reversed entries.
}

func ReadMetricsFile(file string) ([]string, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	// Read the data with a zlib reader
	r, err := zlib.NewReader(bytes.NewBuffer(data))
	defer r.Close()
	if err != nil {
		return nil, fmt.Errorf("Problem with zlib compressed data: %s", err.Error())
	}

	// Store the result of the decode in this map:
	result := map[string]int{}
	err = gob.NewDecoder(r).Decode(&result)
	if err != nil {
		return nil, err
	}

	strings := make([]string, 0, len(result))
	for k := range result {
		strings = append(strings, k)
	}
	return strings, nil
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	common.SetupLogger()

	if *metricsFile == "" {
		common.ExitWithMessage("No metric file specified. Use '-metrics-file'")
	}

	if *rulePath == "" {
		common.ExitWithMessage("No rule path specified. Use '-rule-path'")
	}

	//TODO(cchandler): Make a constructor for a graphite converter so we don't
	//have to stich everything together outside of the package.

	ruleset, err := util.LoadRules(*rulePath)
	if err != nil {
		common.ExitWithMessage(fmt.Sprintf("Error while reading rules: %s", err.Error()))
	}

	graphiteConverter := util.RuleBasedGraphiteConverter{Ruleset: ruleset}

	metrics, err := ReadMetricsFile(*metricsFile)
	if err != nil {
		fmt.Printf("Fatal error reading metrics file %s\n", err)
		os.Exit(1)
	}

	classifiedMetrics := DoAnalysis(metrics, graphiteConverter)
	fmt.Printf("Generating report files...\n")
	GenerateReport(classifiedMetrics[Unmatched], graphiteConverter)

	fmt.Printf("Total metric count %d\n", len(metrics))
}

type ConversionStatus int

const (
	Matched ConversionStatus = iota
	Unmatched
	ReverseFailed
	ReverseChanged
)

func ClassifyMetric(metric string, graphiteConverter util.RuleBasedGraphiteConverter) ConversionStatus {
	graphiteMetric := util.GraphiteMetric(metric)
	taggedMetric, err := graphiteConverter.ToTaggedName(graphiteMetric)
	if err != nil {
		return Unmatched
	}
	reversedMetric, err := graphiteConverter.ToGraphiteName(taggedMetric)
	if err != nil {
		return ReverseFailed
	}
	if reversedMetric != graphiteMetric {
		return ReverseChanged
	}
	return Matched
}

func DoAnalysis(metrics []string, graphiteConverter util.RuleBasedGraphiteConverter) map[ConversionStatus][]string {
	graphiteConverter.EnableStats()

	workQueue := make(chan string, 100)

	go func() {
		// Add the metrics to the work queue
		for _, metric := range metrics {
			workQueue <- metric
		}
		close(workQueue)
	}()

	classifiedMetricResults := map[ConversionStatus]chan string{
		Matched:        make(chan string, 100),
		Unmatched:      make(chan string, 100),
		ReverseFailed:  make(chan string, 100),
		ReverseChanged: make(chan string, 100),
	}

	classifiedMetrics := map[ConversionStatus][]string{}

	var wgClassifyAppend sync.WaitGroup

	for status := range classifiedMetricResults {
		status := status
		// These goroutines move things from the `classifiedMetricResults` map (ConversionStatus => chan string)
		// into the `classifiedMetrics` map (ConversionStatus => []string)
		wgClassifyAppend.Add(1)
		go func() {
			for metric := range classifiedMetricResults[status] {
				classifiedMetrics[status] = append(classifiedMetrics[status], metric)
			}
			wgClassifyAppend.Done()
		}()
	}

	var wgWorkQueue sync.WaitGroup

	fmt.Printf("Starting work...\n")
	for i := 0; i < runtime.NumCPU(); i++ {
		// Launch 1 goroutine per CPU to process the work queue
		// This task is CPU-bound, so adding more goroutines beyond this probably won't help.
		// Benchmarking seems to confirm this suspicion- although even NumCPU() seems high
		wgWorkQueue.Add(1)
		go func() {
			counter := 0
			defer wgWorkQueue.Done()
			for metric := range workQueue {
				counter++
				if counter%1000 == 0 && counter != 0 {
					fmt.Printf(".")
				}
				// Classify the metric, then send it to the corresponding channel.
				classifiedMetricResults[ClassifyMetric(metric, graphiteConverter)] <- metric
			}
		}()
	}
	wgWorkQueue.Wait()

	for _, channel := range classifiedMetricResults {
		close(channel)
	}
	// Wait for the results to be moved from the channels into the slices.
	wgClassifyAppend.Wait()

	fmt.Printf("\n")
	fmt.Printf("Matched: %d\n", len(classifiedMetrics[Matched]))
	fmt.Printf("Unmatched: %d\n", len(classifiedMetrics[Unmatched]))
	fmt.Printf("Reverse convert failed: %d\n", len(classifiedMetrics[ReverseFailed]))
	// Since these indicate broken rules, printing out the particular metrics is very helpful.
	for _, metric := range classifiedMetrics[ReverseFailed] {
		fmt.Printf("\t%s\n", metric)
	}
	fmt.Printf("Reverse convert changed metric: %d\n", len(classifiedMetrics[ReverseChanged]))
	// Since these indicate broken rules, printing out the particular metrics is very helpful.
	for _, metric := range classifiedMetrics[ReverseChanged] {
		fmt.Printf("\t%s\n", metric)
	}
	return classifiedMetrics
}

func GenerateReport(unmatched []string, graphiteConverter util.RuleBasedGraphiteConverter) {
	err := os.RemoveAll("report")
	if err != nil {
		panic("Can't delete the report directory")
	}
	if err := os.Mkdir("report", 0744); err != nil {
		panic("Can't create report directory")
	}

	f, err := os.Create("report/unmatched.txt")
	if err != nil {
		panic("Can't create report/unmatched.txt")
	}
	defer f.Close()

	for _, metric := range unmatched {
		f.WriteString(fmt.Sprintf("%s\n", metric))
	}

	for i, rule := range graphiteConverter.Ruleset.Rules {
		f, err := os.Create(fmt.Sprintf("report/%d.txt", i))
		defer f.Close()
		if err != nil {
			panic("Unable to create report file!")
		}
		f.WriteString(fmt.Sprintf("Rule: %s\n", rule.MetricKeyRegex))

		for _, match := range rule.Statistics.SuccessfulMatches {
			f.WriteString(fmt.Sprintf("%s\n", match))
		}

	}
}
