package common

import (
	"flag"
	"fmt"
	"github.com/square/metrics-indexer/api"
	"github.com/square/metrics-indexer/internal"
	"os"
)

var (
	// YamlFile is the location of the rule YAML file.
	YamlFile = flag.String("yaml-file", "", "Location of YAML configuration file.")
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
		Hosts:            []string{"localhost"},
		Keyspace:         "metrics_indexer",
	})
	if err != nil {
		ExitWithMessage(fmt.Sprintf("Cannot instantiate a new API: %s\n", err.Error()))
	}
	return apiInstance
}
