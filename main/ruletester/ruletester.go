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
	"io"
	"io/ioutil"
	"os"
	// "runtime"
	// "sort"
	"sync"

	"github.com/square/metrics/api"
	"github.com/square/metrics/main/common"
	"github.com/square/metrics/metric_metadata/cassandra"
	"github.com/square/metrics/util"
)

var (
	metricsFile   = flag.String("metrics-file", "", "Location of zlib compressed gob string file.")
	unmatchedFile = flag.String("unmatched-file", "", "location of metrics list to output unmatched transformations.")
	reverse       = flag.Bool("reverse", false, "If true, then attempt the reverse-rule lookup also.")
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

	result := make(map[string]int)
	r, err := zlib.NewReader(bytes.NewBuffer(data))
	b := new(bytes.Buffer)
	io.Copy(b, r)
	d := gob.NewDecoder(b)

	err = d.Decode(&result)
	if err != nil {
		return nil, err
	}

	strings := []string{}
	for k, _ := range result {
		strings = append(strings, k)
	}
	return strings, nil
}

func main() {
	flag.Parse()
	common.SetupLogger()

	if *metricsFile == "" {
		common.ExitWithMessage("No metric file.")
		fmt.Printf("You must specify a metrics file.\n")
		os.Exit(1)
	}

	config := common.LoadConfig()

	// graphiteConfig := util.GraphiteConverterConfig{ConversionRulesPath: config.MetricMetadataAPI.ConversionRulesPath}
	//TODO(cchandler): Make a constructor for a graphite converter so we don't
	//have to stich everything together outside of the package.

	ruleset, err := util.LoadRules(config.MetricMetadataConfig.ConversionRulesPath)

	if err != nil {
		common.ExitWithMessage(fmt.Sprintf("Error while reading rules: %s", err.Error()))
	}
	graphiteConverter := util.RuleBasedGraphiteConverter{Ruleset: ruleset}

	// metricFile, err := os.Open(*metricsFile)
	metrics, err := ReadMetricsFile(*metricsFile)
	if err != nil {
		fmt.Printf("Fatal error reading metrics file %s\n", err)
		os.Exit(1)
	}

	DoAnalysis(metrics, graphiteConverter)

	fmt.Printf("Total metric count %d\n", len(metrics))

	// matched := 0
	// unmatched := 0
	// reverse_convert_failed := 0

	// fmt.Printf("Matched: %d\n", matched)
	// fmt.Printf("Unmatched: %d\n", unmatched)
	// fmt.Printf("Reverse convert failed: %d\n", reverse_convert_failed)

	if err != nil {
		common.ExitWithMessage("No metric file.")
	}
	// scanner := bufio.NewScanner(metricFile)
	cassandraConfig := cassandra.CassandraMetricMetadataConfig{
		Hosts:    config.MetricMetadataConfig.Hosts,
		Keyspace: config.MetricMetadataConfig.Keyspace,
	}
	_ = common.NewMetricMetadataAPI(cassandraConfig)

	// var output *os.File
	// if *unmatchedFile != "" {
	// 	output, err = os.Create(*unmatchedFile)
	// 	if err != nil {
	// 		common.ExitWithMessage(fmt.Sprintf("Error creating the output file: %s", err.Error()))
	// 	}
	// }
	// stat := run(ruleset, scanner, apiInstance, output)
	// report(stat)
}

type ChunkResult struct {
	matched                int
	unmatched              int
	reverse_convert_failed int
}

func DoAnalysis(metrics []string, graphiteConverter util.RuleBasedGraphiteConverter) {
	graphiteConverter.EnableStats()

	workQueue := make(chan []string, len(metrics)/25+1)
	resultQueue := make(chan ChunkResult, len(metrics))
	unmatchedQueue := make(chan string, len(metrics))
	var wg sync.WaitGroup

	i := 0
	for i = 0; i < len(metrics); i = i + 25 {
		workQueue <- metrics[i : i+25]
	}
	if i <= len(metrics) {
		workQueue <- metrics[i:]
	}

	close(workQueue)

	fmt.Printf("Starting work...\n")
	for j := 0; j < 10; j++ {
		wg.Add(1)
		go func() {
			counter := 0
			defer wg.Done()
			for {
				select {
				case metrics, more := <-workQueue:
					counter++
					if counter%100 == 0 && counter != 0 {
						fmt.Printf(".")
					}
					chunk_result := ChunkResult{}
					for _, metric := range metrics {
						graphiteMetric := util.GraphiteMetric(metric)
						taggedMetric, err := graphiteConverter.ToTaggedName(graphiteMetric)
						if err != nil {
							_, err := graphiteConverter.ToGraphiteName(taggedMetric)
							if err != nil {
								chunk_result.reverse_convert_failed++
							} else {

							}
							chunk_result.matched++
						} else {
							unmatchedQueue <- metric
							chunk_result.unmatched++
						}
					}
					resultQueue <- chunk_result

					if !more {
						return
					}
				default:
					fmt.Printf("Nothing more!\n")
				}
			}

		}()
	}
	wg.Wait()
	close(resultQueue)
	close(unmatchedQueue)

	fmt.Printf("\n")
	fmt.Printf("Processing results!")
	totalResults := ChunkResult{}
	//Merge chunks
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case result, more := <-resultQueue:
				totalResults.matched += result.matched
				totalResults.unmatched += result.unmatched
				totalResults.reverse_convert_failed += result.reverse_convert_failed

				if !more {
					return
				}
			}
		}
	}()
	wg.Add(1)
	unmatchedResults := []string{}
	go func() {
		defer wg.Done()
		for {
			select {
			case unmatched, more := <-unmatchedQueue:
				unmatchedResults = append(unmatchedResults, unmatched)
				if !more {
					return
				}
			}
		}
	}()

	wg.Wait()
	fmt.Printf("\n")
	fmt.Printf("Matched: %d\n", totalResults.matched)
	fmt.Printf("Unmatched: %d\n", totalResults.unmatched)
	fmt.Printf("Reverse convert failed: %d\n", totalResults.reverse_convert_failed)

	GenerateReport(unmatchedResults, graphiteConverter)
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

// func run(ruleset util.RuleSet, scanner *bufio.Scanner, apiInstance api.MetricMetadataAPI, unmatched *os.File) Statistics {
// 	var wg sync.WaitGroup
// 	stat := Statistics{
// 		perMetric: make(map[api.MetricKey]PerMetricStatistics),
// 	}
// 	type result struct {
// 		input   string
// 		result  api.TaggedMetric
// 		success bool
// 	}
// 	inputBuffer := make(chan string, 10)
// 	outputBuffer := make(chan result, 10)
// 	done := make(chan struct{})
// 	for id := 0; id < runtime.NumCPU(); id++ {
// 		go func() {
// 			for {
// 				select {
// 				case <-done:
// 					return
// 				case input := <-inputBuffer:
// 					metric, matched := ruleset.MatchRule(input)
// 					outputBuffer <- result{
// 						input,
// 						metric,
// 						matched,
// 					}
// 				}
// 			}
// 		}()
// 	}
// 	go func() {
// 		// aggregate function.
// 		for {
// 			select {
// 			case <-done:
// 				return
// 			case output := <-outputBuffer:
// 				converted, matched := output.result, output.success
// 				if matched {
// 					stat.matched++
// 					perMetric := stat.perMetric[converted.MetricKey]
// 					perMetric.matched++
// 					if *insertToDatabase {
// 						apiInstance.AddMetric(converted)
// 					}
// 					if *reverse {
// 						reversed, err := ruleset.ToGraphiteName(converted)
// 						if err != nil {
// 							perMetric.reverseError++
// 						} else if string(reversed) != output.input {
// 							perMetric.reverseIncorrect++
// 						} else {
// 							perMetric.reverseSuccess++
// 						}
// 					}
// 					stat.perMetric[converted.MetricKey] = perMetric
// 				} else {
// 					stat.unmatched++
// 					if unmatched != nil {
// 						unmatched.WriteString(output.input)
// 						unmatched.WriteString("\n")
// 					}
// 				}
// 				wg.Done()
// 			}
// 		}
// 	}()

// 	for scanner.Scan() {
// 		wg.Add(1)
// 		input := scanner.Text()
// 		inputBuffer <- input
// 	}
// 	wg.Wait()
// 	close(done) // broadcast to shutdown all goroutines.
// 	return stat
// }

// func report(stat Statistics) {
// 	total := stat.matched + stat.unmatched
// 	fmt.Printf("Processed %d entries\n", total)
// 	fmt.Printf("Matched:   %d\n", stat.matched)
// 	fmt.Printf("Unmatched: %d\n", stat.unmatched)
// 	fmt.Printf("Per-rule statistics\n")
// 	rowformat := "%-60s %7d %7d %7d %7d\n"
// 	headformat := "%-60s %7s %7s %7s %7s\n"
// 	fmt.Printf(headformat, "name", "match", "rev-suc", "rev-err", "rev-fail")
// 	sortedKeys := make([]string, len(stat.perMetric))
// 	index := 0
// 	for key := range stat.perMetric {
// 		sortedKeys[index] = string(key)
// 		index++
// 	}
// 	sort.Strings(sortedKeys)
// 	for _, key := range sortedKeys {
// 		perMetric := stat.perMetric[api.MetricKey(key)]
// 		fmt.Printf(rowformat,
// 			string(key),
// 			perMetric.matched,
// 			perMetric.reverseSuccess,
// 			perMetric.reverseError,
// 			perMetric.reverseIncorrect,
// 		)
// 	}
// }
