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
	"encoding/json"
	"flag"
	"fmt"

	"github.com/peterh/liner"
	"github.com/square/metrics/api/backend"
	"github.com/square/metrics/api/backend/blueflood"
	"github.com/square/metrics/main/common"
	"github.com/square/metrics/query"
)

func main() {
	flag.Parse()
	common.SetupLogger()

	config := common.LoadConfig()

	apiInstance := common.NewAPI(config.API)
	myBackend := blueflood.NewBlueflood(config.Blueflood)

	l := liner.NewLiner()
	defer l.Close()
	for {
		input, err := l.Prompt("> ")
		if err != nil {
			return
		}

		l.AppendHistory(input)

		cmd, err := query.Parse(input)
		if err != nil {
			fmt.Println("parsing error", err.Error())
			continue
		}

		n, ok := cmd.(query.Node)
		if !ok {
			fmt.Printf("error: %#v doesn't implement Node\n", cmd)
			continue
		}
		fmt.Println(query.PrintNode(n))

		result, err := cmd.Execute(query.ExecutionContext{Backend: backend.NewSequentialMultiBackend(myBackend), API: apiInstance, FetchLimit: 1000})
		if err != nil {
			fmt.Println("execution error:", err.Error())
			continue
		}
		encoded, err := json.Marshal(result)
		if err != nil {
			fmt.Println("encoding error:", err.Error())
			return
		}
		fmt.Println("success:")
		fmt.Println(string(encoded))
	}
}
