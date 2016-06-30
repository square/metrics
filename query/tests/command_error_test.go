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

// Integration test for the query execution.
package tests

import (
	"strings"
	"testing"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/query/command"
	"github.com/square/metrics/query/parser"
	"github.com/square/metrics/testing_support/mocks"

	"golang.org/x/net/context"
)

func TestCommandError(t *testing.T) {
	testTimerange, err := api.NewTimerange(0, 120, 30)
	if err != nil {
		t.Fatalf("Error creating timerange for test: %s", err.Error())
	}
	comboAPI := mocks.NewComboAPI(testTimerange,
		api.Timeseries{Values: []float64{1, 2, 3, 4, 5}, TagSet: api.TagSet{"metric": "testmetric", "host": "h1"}},
		api.Timeseries{Values: []float64{5, 4, 3, 4, 5}, TagSet: api.TagSet{"metric": "testmetric", "host": "h2"}},
		api.Timeseries{Values: []float64{1, 7, 7, 7, 5}, TagSet: api.TagSet{"metric": "testmetric", "host": "h3"}},
		api.Timeseries{Values: []float64{3, 3, 3, 3, 3}, TagSet: api.TagSet{"metric": "testmetric", "host": "h4"}},
		api.Timeseries{Values: []float64{1, 2, 0, 0, 5}, TagSet: api.TagSet{"metric": "testmetric", "host": "h5"}},
		api.Timeseries{Values: []float64{0, 2, 0, 0, 0}, TagSet: api.TagSet{"metric": "testmetric", "host": "h6"}},
	)

	context := command.ExecutionContext{
		TimeseriesStorageAPI: comboAPI,
		MetricMetadataAPI:    comboAPI,
		FetchLimit:           13,
		Timeout:              100 * time.Millisecond,
		Ctx:                  context.Background(),
	}
	command, err := parser.Parse(`select testmetric + testmetric + testmetric from 0 to 120 resolution 30ms`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	_, err = command.Execute(context)
	if err == nil {
		t.Fatalf("expected error")
	}
	t.Logf("Message :: %s", err.Error())
	if !strings.Contains(err.Error(), "brings the total to 18") {
		t.Errorf(`"brings the total to 18" expected in error message %s`, err.Error())
	}
	if !strings.Contains(err.Error(), "specified limit 13") {
		t.Errorf(`"specified limit 13" expected in error message %s`, err.Error())
	}
	if !strings.Contains(err.Error(), "6 additional series") {
		t.Errorf(`"6 additional series" expected in error message %s`, err.Error())
	}
}
