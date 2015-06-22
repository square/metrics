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
	"io/ioutil"
	"os"
  standard_log "log"

	"github.com/square/metrics/api"
	"github.com/square/metrics/api/backend/blueflood"
	"github.com/square/metrics/internal"
	"github.com/square/metrics/log"
	"github.com/square/metrics/log/standard"
	"github.com/square/metrics/ui"
	"gopkg.in/yaml.v2"
)

var (
	// YamlFile is the location of the rule YAML file.
	ConfigFile = flag.String("config-file", "", "Location of YAML config file")
	Verbose    = flag.Bool("verbose", false, "Set to true to enable logging")
)

type Config struct {
	Blueflood blueflood.Config `yaml:"blueflood"`
	API       api.Config       `yaml:"api"` // TODO: Probably rethink how we name this
	UIConfig  ui.Config        `yaml:"ui"`
}

func LoadConfig() Config {
	var config Config

	f, err := os.Open(*ConfigFile)
	if err != nil {
		ExitWithMessage(fmt.Sprintf("unable to open config file `%s`", *ConfigFile))
	}

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		ExitWithMessage(fmt.Sprintf("unable to read config file `%s`", *ConfigFile))
	}

	if err := yaml.Unmarshal(bytes, &config); err != nil {
		ExitWithMessage(fmt.Sprintf("unable to load config file `%s`", *ConfigFile))
	}

	fmt.Printf("parsed config: %#v\n", config)

	return config
}

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
func NewAPI(config api.Config) api.API {
	apiInstance, err := internal.NewAPI(config)
	if err != nil {
		ExitWithMessage(fmt.Sprintf("Cannot instantiate a new API from %#v: %s\n", config, err.Error()))
	}
	return apiInstance
}

func SetupLogger() {
	if *Verbose {
		log.InitLogger(&standard.StandardLogger{standard_log.New(os.Stderr, "", standard_log.LstdFlags)})
	}
}
