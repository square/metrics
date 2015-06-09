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

package common

import (
	"flag"
	"fmt"
	"os"

	"github.com/square/metrics/api"
	"github.com/square/metrics/internal"
)

var (
	// YamlFile is the location of the rule YAML file.
	YamlFile      = flag.String("yaml-file", "", "Location of YAML configuration file.")
	CassandraHost = flag.String("cassandra-host", "localhost", "Cassandra host")
)

// ExitWithRequired terminates the program when a required flag is missing.
func ExitWithRequired(flagName string) {
	fmt.Fprintf(os.Stderr, "%s is required\n", flagName)
	os.Exit(1)
}

// ExitWithMessage terminates the program with the provided message.
func ExitWithMessage(message string) {
	fmt.Fprint(os.Stderr, message)
	os.Exit(1)
}

// NewAPI creates a new instance of the API.
func NewAPI() api.API {
	apiInstance, err := internal.NewAPI(api.Configuration{
		RuleYamlFilePath: *YamlFile,
		Hosts:            []string{*CassandraHost},
		Keyspace:         "metrics_indexer",
	})
	if err != nil {
		ExitWithMessage(fmt.Sprintf("Cannot instantiate a new API: %s\n", err.Error()))
	}
	return apiInstance
}
