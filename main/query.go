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
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/square/metrics/api/backend/blueflood"
	"github.com/square/metrics/main/common"
	"github.com/square/metrics/query"
	"os"
)

var (
	BluefloodUrl      = flag.String("blueflood-url", "", "Blueflood url")
	BluefloodTenantId = flag.String("blueflood-tenant-id", "", "Blueflood tenant id")
)

func main() {
	flag.Parse()
	if *BluefloodUrl == "" {
		common.ExitWithRequired("blueflood-url")
	}
	if *BluefloodTenantId == "" {
		common.ExitWithRequired("blueflood-tenant-id")
	}

	apiInstance := common.NewAPI()
	backend := blueflood.NewBlueflood(apiInstance, *BluefloodUrl, *BluefloodTenantId)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		cmd, err := query.Parse(input)
		if err != nil {
			fmt.Println("parsing error", err.Error())
			continue
		}

		n, ok := cmd.(query.Node)
		if !ok {
			fmt.Println("error: %+v doesn't implement Node", cmd)
			continue
		}
		fmt.Println(query.PrintNode(n))

		result, err := cmd.Execute(backend)
		if err != nil {
			fmt.Println("execution error:", err.Error())
			continue
		}
		encoded, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Println("encoding error:", err.Error())
			return
		}
		fmt.Println("success:")
		fmt.Println(string(encoded))
	}
}
