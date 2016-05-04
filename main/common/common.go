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

package common

import (
	"flag"
	"fmt"
	"io/ioutil"
	standard_log "log"
	"os"

	"github.com/square/metrics/log"
	"github.com/square/metrics/log/glog"
	"github.com/square/metrics/log/standard"
	"github.com/square/metrics/metric_metadata"
	"github.com/square/metrics/metric_metadata/cassandra"
	"github.com/square/metrics/timeseries/blueflood"
	"github.com/square/metrics/ui"
	"gopkg.in/yaml.v2"
)

var (
	// ConfigFile is the location of the rule YAML file.
	ConfigFile = flag.String("config-file", "", "Location of YAML config file")
	Logger     = flag.String("logger", "glog", "Selects the logger to use")
)

type Config struct {
	ConversionRulesPath string           `yaml:"conversion_rules_path"`
	Blueflood           blueflood.Config `yaml:"blueflood"`
	Cassandra           cassandra.Config `yaml:"cassandra"`
	UI                  ui.Config        `yaml:"ui"`
}

func LoadConfig() Config {
	var config Config
	if *ConfigFile == "" {
		ExitWithMessage("No config file was specified. Specify it with '-config-file'")
	}
	f, err := os.Open(*ConfigFile)
	if err != nil {
		ExitWithMessage(fmt.Sprintf("unable to open config file `%s`", *ConfigFile))
	}

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		ExitWithMessage(fmt.Sprintf("unable to read config file `%s`", *ConfigFile))
	}

	if err := yaml.Unmarshal(bytes, &config); err != nil {
		ExitWithMessage(fmt.Sprintf("unable to load config file `%s`: %s", *ConfigFile, err.Error()))
	}

	fmt.Printf("parsed config: %#v\n", config)

	return config
}

// ExitWithMessage terminates the program with the provided message.
func ExitWithMessage(message string) {
	fmt.Fprintf(os.Stderr, "%s\n", message)
	os.Exit(1)
}

// NewMetricMetadataAPI creates a new instance of the API.
func NewMetricMetadataAPI(config cassandra.Config) metadata.MetricAPI {
	apiInstance, err := cassandra.NewMetricMetadataAPI(config)
	if err != nil {
		ExitWithMessage(fmt.Sprintf("Cannot instantiate a new API from %#v: %s\n", config, err.Error()))
	}
	return apiInstance
}

func SetupLogger() {
	if *Logger == "glog" {
		log.InitLogger(&glog.Logger{})
		log.Infof("Using glog logger")
	} else {
		log.InitLogger(&standard.Logger{standard_log.New(os.Stderr, "", standard_log.LstdFlags)})
		log.Infof("Using standard logger")
	}
}
