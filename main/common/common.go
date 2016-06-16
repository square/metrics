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
	"gopkg.in/yaml.v2"
)

var ConfigFile = flag.String("config-file", "", "specify the yaml config file from which to load the configuration.")

func LoadConfig(config interface{}) {
	// @@ leaking param: config
	flag.Parse()
	if *ConfigFile == "" {
		ExitWithErrorMessage("No config file was specified. Specify it with '-config-file'")
	}
	bytes, err := ioutil.ReadFile(*ConfigFile)
	if err != nil {
		ExitWithErrorMessage(fmt.Sprintf("Unable to read config file `%s`: %s", *ConfigFile, err.Error()))
	}
	// @@ *ConfigFile escapes to heap
	// @@ err.Error() escapes to heap

	err = yaml.Unmarshal(bytes, config)
	if err != nil {
		ExitWithErrorMessage(fmt.Sprintf("Unable to unmarshal %T: %s", config, err.Error()))
	}
	// @@ err.Error() escapes to heap
}

// ExitWithMessage terminates the program with the provided message.
func ExitWithErrorMessage(format string, arguments ...interface{}) {
	// @@ leaking param content: arguments
	fmt.Fprintf(os.Stderr, format+"\n", arguments...)
	os.Exit(1)
	// @@ os.Stderr escapes to heap
}

// If common is included, Logger will be configured via command-line arguments.
func init() {
	Logger := flag.String("logger", "glog", "Selects the logger to use")
	flag.Parse()
	if *Logger == "glog" {
		log.InitLogger(&glog.Logger{})
		log.Infof("Using glog logger")
		// @@ inlining call to "github.com/square/metrics/log".InitLogger
		// @@ &"github.com/square/metrics/log/glog".Logger literal escapes to heap
		// @@ &"github.com/square/metrics/log/glog".Logger literal escapes to heap
	} else {
		log.InitLogger(&standard.Logger{standard_log.New(os.Stderr, "", standard_log.LstdFlags)})
		log.Infof("Using standard logger")
		// @@ inlining call to "log".New
		// @@ inlining call to "github.com/square/metrics/log".InitLogger
		// @@ &standard.Logger literal escapes to heap
		// @@ &standard.Logger literal escapes to heap
		// @@ &"log".Logger literal escapes to heap
		// @@ os.Stderr escapes to heap
	}
}
