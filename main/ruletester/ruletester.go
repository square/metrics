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
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"
	"sync"

	"github.com/square/metrics/api"
	"github.com/square/metrics/internal"
	"github.com/square/metrics/main/common"
)

var (
	metricsFile      = flag.String("metrics-file", "", "Location of YAML configuration file.")
	unmatchedFile    = flag.String("unmatched-file", "", "location of metrics list to output unmatched transformations.")
	insertToDatabase = flag.Bool("insert-to-db", false, "If true, insert rows to database.")
	reverse          = flag.Bool("reverse", false, "If true, then attempt the reverse-rule lookup also.")
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

func main() {
	flag.Parse()
	common.SetupLogger()

	config := common.LoadConfig()

	ruleset, err := internal.LoadRules(config.API.ConversionRulesPath)
	if err != nil {
		common.ExitWithMessage(fmt.Sprintf("Error while reading rules: %s", err.Error()))
	}
	metricFile, err := os.Open(*metricsFile)
	if err != nil {
		common.ExitWithMessage("No metric file.")
	}
	scanner := bufio.NewScanner(metricFile)
	apiInstance := common.NewAPI(config.API)
	var output *os.File
	if *unmatchedFile != "" {
		output, err = os.Create(*unmatchedFile)
		if err != nil {
			common.ExitWithMessage(fmt.Sprintf("Error creating the output file: %s", err.Error()))
		}
	}
	stat := run(ruleset, scanner, apiInstance, output)
	report(stat)
}

func run(ruleset internal.RuleSet, scanner *bufio.Scanner, apiInstance api.API, unmatched *os.File) Statistics {
	var wg sync.WaitGroup
	stat := Statistics{
		perMetric: make(map[api.MetricKey]PerMetricStatistics),
	}
	type result struct {
		input   string
		result  api.TaggedMetric
		success bool
	}
	inputBuffer := make(chan string, 10)
	outputBuffer := make(chan result, 10)
	done := make(chan struct{})
	for id := 0; id < runtime.NumCPU(); id++ {
		go func() {
			for {
				select {
				case <-done:
					return
				case input := <-inputBuffer:
					metric, matched := ruleset.MatchRule(input)
					outputBuffer <- result{
						input,
						metric,
						matched,
					}
				}
			}
		}()
	}
	go func() {
		// aggregate function.
		for {
			select {
			case <-done:
				return
			case output := <-outputBuffer:
				converted, matched := output.result, output.success
				if matched {
					stat.matched++
					perMetric := stat.perMetric[converted.MetricKey]
					perMetric.matched++
					if *insertToDatabase {
						apiInstance.AddMetric(converted)
					}
					if *reverse {
						reversed, err := ruleset.ToGraphiteName(converted)
						if err != nil {
							perMetric.reverseError++
						} else if string(reversed) != output.input {
							perMetric.reverseIncorrect++
						} else {
							perMetric.reverseSuccess++
						}
					}
					stat.perMetric[converted.MetricKey] = perMetric
				} else {
					stat.unmatched++
					if unmatched != nil {
						unmatched.WriteString(output.input)
						unmatched.WriteString("\n")
					}
				}
				wg.Done()
			}
		}
	}()

	for scanner.Scan() {
		wg.Add(1)
		input := scanner.Text()
		inputBuffer <- input
	}
	wg.Wait()
	close(done) // broadcast to shutdown all goroutines.
	return stat
}

func report(stat Statistics) {
	total := stat.matched + stat.unmatched
	fmt.Printf("Processed %d entries\n", total)
	fmt.Printf("Matched:   %d\n", stat.matched)
	fmt.Printf("Unmatched: %d\n", stat.unmatched)
	fmt.Printf("Per-rule statistics\n")
	rowformat := "%-60s %7d %7d %7d %7d\n"
	headformat := "%-60s %7s %7s %7s %7s\n"
	fmt.Printf(headformat, "name", "match", "rev-suc", "rev-err", "rev-fail")
	sortedKeys := make([]string, len(stat.perMetric))
	index := 0
	for key := range stat.perMetric {
		sortedKeys[index] = string(key)
		index++
	}
	sort.Strings(sortedKeys)
	for _, key := range sortedKeys {
		perMetric := stat.perMetric[api.MetricKey(key)]
		fmt.Printf(rowformat,
			string(key),
			perMetric.matched,
			perMetric.reverseSuccess,
			perMetric.reverseError,
			perMetric.reverseIncorrect,
		)
	}
}
