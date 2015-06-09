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

// Integration test for the query execution.
package query

import (
	"testing"

	"github.com/square/metrics/api"
	"github.com/square/metrics/assert"
)

type fakeApiBackend struct {
	api.Backend
}

type simpleFakeApi struct {
	api.API
}

func (f simpleFakeApi) GetAllTags(metricKey api.MetricKey) ([]api.TagSet, error) {
	return []api.TagSet{
		api.ParseTagSet("dc=west,env=production,host=a"),
		api.ParseTagSet("dc=west,env=staging,host=b"),
		api.ParseTagSet("dc=east,env=production,host=c"),
		api.ParseTagSet("dc=east,env=staging,host=d"),
	}, nil
}

func (f fakeApiBackend) Api() api.API {
	return simpleFakeApi{}
}

func TestCommand_Describe(t *testing.T) {
	var fakeBackend fakeApiBackend

	for _, test := range []struct {
		query   string
		backend api.Backend
		length  int // expected length of the result.
	}{
		{"describe m", fakeBackend, 4},
		{"describe m where dc='west'", fakeBackend, 2},
		{"describe m where dc='west' or env = 'production'", fakeBackend, 3},
		{"describe m where dc='west' or env = 'production' and doesnotexist = ''", fakeBackend, 2},
		{"describe m where env = 'production' and doesnotexist = '' or dc = 'west'", fakeBackend, 2},
		// {"describe m where (dc='west' or env = 'production') and doesnotexist = ''", fakeBackend, 0}, // PARSER ERROR, currently.
	} {
		a := assert.New(t).Contextf("query=%s", test.query)
		rawCommand, err := Parse(test.query)
		if err != nil {
			a.Errorf("Unexpected error while parsing")
			continue
		}
		command := rawCommand.(*DescribeCommand)
		rawResult, _ := command.Execute(test.backend)
		parsedResult := rawResult.([]string)
		a.EqInt(len(parsedResult), test.length)
	}
}
