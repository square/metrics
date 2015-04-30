package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/square/metrics-indexer/main/common"
	"github.com/square/metrics-indexer/query"
	"os"
)

func main() {
	flag.Parse()
	apiInstance := common.NewAPI()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		cmd, err := query.Parse(input)
		if err != nil {
			fmt.Println("parsing error", err.Error())
			continue
		}
		result, err := cmd.Execute(apiInstance)
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
