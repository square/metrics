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

package query

// show all metrics
// show tags WHERE predicate

import (
	"testing"

	"github.com/square/metrics/assert"
)

// these queries should successfully parse,
// with a corresponding command.
var inputs = []string{
	// describes
	"describe all",
	"describe x",
	"describe cpu_usage",
	"describe inspect",
	"describe in2",
	"describe cpu_usage where key = 'value'",
	"describe cpu_usage where key = 'value\\''",
	"describe cpu_usage where key != 'value'",
	"describe cpu_usage where (key = 'value')",
	"describe cpu_usage where not (key = 'value')",
	"describe cpu_usage where not key = 'value'",
	"describe cpu_usage where (key = 'value')",
	"describe cpu_usage where key = 'value' or key = 'value'",
	"describe cpu_usage where key in ('value', 'value')",
	"describe cpu_usage where key matches 'abc'",
	"describe nodes.cpu.usage where datacenter='sjc1b' and type='idle' and host matches 'fwd'",
	// predicate parenthesis test
	"describe cpu_usage where key = 'value' and (key = 'value')",
	"describe cpu_usage where (key = 'value') and key = 'value'",
	"describe cpu_usage where (key = 'value') and (key = 'value')",
	"describe cpu_usage where (key = 'value' and key = 'value')",
	// selects - parenthesis
	"select 0 from 0 to 0",
	"select (0) from 0 to 0",
	"select (0) where foo = 'bar' from 0 to 0",
	// selects - numbers
	"select 0, 1, 2, 3, 4, 5, 6, 7, 8, 9 from 0 to 0",
	"select 10, 100, 1000 from 0 to 0",
	"select 10.1, 10.01, 10.001 from 0 to 0",
	"select -10.1, -10.01, -10.001 from 0 to 0",
	"select 1.0e1, 1.0e2, 1.0e10, 1.0e0 from 0 to 0",
	"select 1.0e-5, 1.0e+5 from 0 to 0",
	// selects - trying out arithmetic
	"select x from 0 to 0",
	"select x-y-z from 0 to 0",
	"select (x)-(y)-(z) from 0 to 0",
	"select 0 from 0 to 0",
	"select x, y from 0 to 0",
	"select 1 + 2 * 3 + 4 from 0 to 0",
	"select x * (y + 123), z from 0 to 0",
	// testing escaping
	"select `x` from 0 to 0",
	// selects - timestamps
	"select x * (y + 123), z from '2011-2-4 PTZ' to '2015-6-1 PTZ'",
	"select x * (y + 123), z from 0 to 10000",
	"select 1 from -10m to now",
	"select 1 from -10M to -10m",
	// selects - function calls
	"select foo(x) from 0 to 0",
	"select bar(x, y) from 0 to 0",
	"select baz(x, y, z+1+foo(1)) from 0 to 0",
	// selects - testing out property values
	"select x from 0 to 0",
	"select x from 0 to 0",
	"select x from 0 to 0 resolution 10",
	"select x from 0 to 0 resolution '10h'",
	"select x from 0 to 0 resolution '300s'",
	"select x from 0 to 0 resolution '17m'",
	"select x from 0 to 0 sample by 'max'",
	"select x from 0 to 0 sample   by 'max'",
	// Leading/trailing whitespace
	" describe all ",
	" describe x ",
	" select 0, 1, 2, 3, 4, 5, 6, 7, 8, 9 from 0 to 0 ",
}

var parseOnly = []string{
	// selects - aggregate functions
	"select scalar.max(x) from 0 to 0",
	"select aggregate.max(x, y) from 0 to 0",
	"select aggregate.max(x group by foo) + 3 from 0 to 0",
	// selects - where clause
	"select x where y = 'z' from 0 to 0",
	// selects - per-identifier where clause
	"select x + z[y = 'z'] from 0 to 0",
	"select x[y = 'z'] from 0 to 0",
	// selects - complicated queries
	"select aggregate.max(x[y = 'z'] group by foo) from 0 to 0",
	"select cpu.user + cpu.kernel where host = 'apa3.sjc2b' from 0 to 0",
	"select 'string literal' where host = 'apa3.sjc2b' from 0 to 0",
	"select timeshift( metric, '5h') where host = 'apa3.sjc2b' from 0 to 0",
}

// these queries should fail with a syntax error.
var syntaxErrorQuery = []string{
	"select ( from 0 to 0",
	"select ) from 0 to 0",
	"describe ( from 0 to 0",
	"describe in from 0 to 0",
	"describe invalid_regex where key matches 'ab[' from 0 to 0",
	"select x invalid_property 0 from 0 to 0",
	"select x sampleby 0 from 0 to 0",
	"select x sample 0 from 0 to 0",
	"select x by 0 from 0 to 0",
	"select x",
	"select x from 0",
	"select x to 0",
	"select x from 0 from 1 to 0",
	"select x from 0 to 1 to 0",
	"select x from 0 resolution 30 resolution 25 to 0",
	"select x from 0 from 1 sample by 'min' sample by 'min' to 0",
}

func TestParse_success(t *testing.T) {
	for _, row := range inputs {
		_, err := Parse(row)
		if err != nil {
			t.Errorf("[%s] failed to parse: %s", row, err.Error())
		}
	}
	for _, row := range parseOnly {
		_, err := Parse(row)
		if err != nil {
			t.Errorf("[%s] failed to parse: %s", row, err.Error())
		}
	}
}

func TestParse_syntaxError(t *testing.T) {
	for _, row := range syntaxErrorQuery {
		_, err := Parse(row)
		if err == nil {
			t.Errorf("[%s] should have failed to parse", row)
		} else if _, ok := err.(SyntaxErrors); !ok {
			t.Logf("[%s] Expected SyntaxErrors, got: %s", row, err.Error())
		}
	}
}

func TestCompile(t *testing.T) {
	for _, row := range inputs {
		a := assert.New(t).Contextf(row)
		p := Parser{Buffer: row}
		p.Init()
		a.CheckError(p.Parse())
		p.Execute()
		testParserResult(a, p)
	}
}

// Helper functions
// ================

func testParserResult(a assert.Assert, p Parser) {
	a.EqInt(len(p.assertions), 0)
	if len(p.assertions) != 0 {
		for _, err := range p.assertions {
			a.Errorf("assertion error: %s", err.Error())
		}
	}
	if len(p.nodeStack) != 0 {
		for _, node := range p.nodeStack {
			a.Errorf("node error:\n%s", PrintNode(node))
		}
	}
}
