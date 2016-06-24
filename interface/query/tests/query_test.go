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

package tests

// show all metrics
// show tags WHERE predicate

import (
	"testing"

	"github.com/square/metrics/interface/query/parser"
)

// these queries should successfully parse,
// with a corresponding command.
var inputs = []string{
	// describe all
	"describe all",
	"describe all match 'abc'",
	"describe all match \"abc\"",
	// describes
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
	"describe cpu_usage where key match 'abc'",
	"describe nodes.cpu.usage where datacenter='sjc1b' and type='idle' and host match 'fwd'",
	// predicate parenthesis test
	"describe cpu_usage where key = 'value' and (key = 'value')",
	"describe cpu_usage where (key = 'value') and key = 'value'",
	"describe cpu_usage where (key = 'value') and (key = 'value')",
	"describe cpu_usage where (key = 'value' and key = 'value')",

	// Leading/trailing whitespace
	" describe all ",
	" describe x ",
	" select 0, 1, 2, 3, 4, 5, 6, 7, 8, 9 from 0 to 0 ",

	// selects - spaces and keywords
	"select f( g(5) group by a,w,q) from 0 to 0",
	"select f( g(5) group by a, w, q )from 0 to 0 ",
	"select f( g(5) group by a,w, q)from 0 to 0",
	"select f( g(5) group by a,w,q)to 0 from 0",
	"select f(g(5)group by a,w,q)to 0 from 0",
	"   select f(g(5)group by a,w,q)  from  	  0 	   to     0",
	"   select( f(g(5)group by a,w,q)  )from 0 to 0",
	"select(f(g(5)group by`a`,w,q)) from 0 to 0",
	"select(f(g(5)group by`a`,w,q)) from 0 to 0",
	"select(fromx+tox+groupx+byx+selectx+describex+allx+wherex) from 0 to 0",
	// comments
	"describe all-- trailing comment",
	"describe-- trailing comment\n all",
	"describe/* inline */all",
	"describe/**/all",
	"select/**/(/**/f/**/(/**/g/**/(/**/5/**/)group/**/by/**/`a`/**/,/**/w/**/,/**/q/**/)/**/)/**/from/**/0/**/to/**/0",
}

var selects = []string{
	// All these queries are tested with and without the prefix "select"
	// selects - parenthesis
	"0 from 0 to 0",
	"(0) from 0 to 0",
	"(0) where foo = 'bar' from 0 to 0",
	// selects - numbers
	"0, 1, 2, 3, 4, 5, 6, 7, 8, 9 from 0 to 0",
	"10, 100, 1000 from 0 to 0",
	"10.1, 10.01, 10.001 from 0 to 0",
	"-10.1, -10.01, -10.001 from 0 to 0",
	"1.0e1, 1.0e2, 1.0e10, 1.0e0 from 0 to 0",
	"1.0e-5, 1.0e+5 from 0 to 0",
	// selects - trying out arithmetic
	"x from 0 to 0",
	"x-y-z from 0 to 0",
	"(x)-(y)-(z) from 0 to 0",
	"0 from 0 to 0",
	"x, y from 0 to 0",
	"1 + 2 * 3 + 4 from 0 to 0",
	"x * (y + 123), z from 0 to 0",
	// testing escaping
	"`x` from 0 to 0",
	// selects - timestamps
	"x * (y + 123), z from '2011-2-4 PTZ' to '2015-6-1 PTZ'",
	"x * (y + 123), z from 0 to 10000",
	"1 from -10m to now",
	"1 from -10M to -10m",
	// selects - function calls
	"foo(x) from 0 to 0",
	"bar(x, y) from 0 to 0",
	"baz(x, y, z+1+foo(1)) from 0 to 0",
	// selects - testing out property values
	"x from 0 to 0",
	"x from 0 to 0",
	"x from 0 to 0 resolution '10s'",
	"x from 0 to 0 resolution '10h'",
	"x from 0 to 0 resolution '300s'",
	"x from 0 to 0 resolution '17m'",
	"x from 0 to 0 sample by 'max'",
	"x from 0 to 0 sample   by 'max'",
	// selects - aggregate functions
	"scalar.max(x) from 0 to 0",
	"aggregate.max(x, y) from 0 to 0",
	"aggregate.max(x group by foo) + 3 from 0 to 0",
	// selects - where clause
	"x where y = 'z' from 0 to 0",
	// selects - per-identifier where clause
	"x + z[y = 'z'] from 0 to 0",
	"x[y = 'z'] from 0 to 0",
	// selects - complicated queries
	"aggregate.max(x[y = 'z'] group by foo) from 0 to 0",
	"aggregate.max(x[y = 'z'] collapse by foo) from 0 to 0",
	"aggregate.max(x[y = 'z'] group by foo, bar) from 0 to 0",
	"aggregate.max(x[y = 'z'] collapse by foo, bar) from 0 to 0",
	"cpu.user + cpu.kernel where host = 'apa3.sjc2b' from 0 to 0",
	"'string literal' where host = 'apa3.sjc2b' from 0 to 0",
	"timeshift( metric, '5h') where host = 'apa3.sjc2b' from 0 to 0",
	// pipe expressions
	"x | y from 0 to 0",
	"x | y + 1 from 0 to 0",
	"x | y - 1 from 0 to 0",
	"x | y * 1 from 0 to 0",
	"x | y / 1 from 0 to 0",
	"x | y(group by a) from 0 to 0",
	"x + 1 | y(group by a) from 0 to 0",
	"x | y | z + 1 from 0 to 0",
	"x|y from 0 to 0",
	"x|f + y*z from 0 to 0",
	"x|f + y|g from 0 to 0",
	"x|f + y|g(4) from 0 to 0",
	"x|f(1,2,3) + y|g(4) from 0 to 0",
	"x|f(1s,2,3y) + y|g(4mo) from 0 to 0",
	"x|f(1s,'r3r2',3y) + y|g(4mo) from 0 to 0",
	"1 + 2 | f from 0 to 0",
}

// these queries should fail with a syntax error.
var syntaxErrorQuery = []string{
	"",
	"\n",
	" \n",
	"\t",
	" ",
	"// comment only",
	// invalid regex
	"describe all match 'ab['",
	"describe invalid_regex where key match 'ab['",
	// invalid syntax
	"describe (",
	"describe ( from 0 to 0",
	"describe in",
	"describe in from 0 to 0",
	"describe invalid_property \nwhere key match 'ab' from 0 to 0",
	"describe all matches 'abc'", // matches is not a keyword.
	"select 'a\nac\nabc",
	"select ( from 0 to 0",
	"select ) from 0 to 0",
	"select x invalid_property 0 from 0 to 0",
	"select x sampleby 0 from 0 to 0",
	"select x sample 0 from 0 to 0",
	"select x by 0 from 0 to 0",
	"select x",
	"select x from 0",
	"select x to 0",
	"select x from 0 from 1 to 0",
	"select x from 0 to 1 to 0",
	"select x from 0 resolution '30s' resolution '25s' to 0",
	"select x from 0 from 1 sample by 'min' sample by 'min' to 0",
	"select f(3 groupby x) from 0 to 0",
	"select c group by a from 0 to 0",
	"select x[] from 0 to 0",
	"select cpu | transform.moving_average(10qq) from 0 to 0",
}

func TestParse_success(t *testing.T) {
	for _, row := range inputs {
		_, err := parser.Parse(row)
		if err != nil {
			t.Errorf("[%s] failed to parse: %s", row, err.Error())
		}
	}

	for _, row := range selects {
		for _, prefix := range []string{"", "select "} {
			query := prefix + row
			_, err := parser.Parse(query)
			if err != nil {
				t.Errorf("[%s] failed to parse: %s", query, err.Error())
			}
		}
	}
}

func TestParse_syntaxError(t *testing.T) {
	for _, row := range syntaxErrorQuery {
		_, err := parser.Parse(row)
		if err == nil {
			t.Errorf("[%s] should have failed to parse", row)
		} else if _, ok := err.(parser.SyntaxErrors); !ok {
			t.Logf("[%s] Expected SyntaxErrors, got: %s", row, err.Error())
			err.Error() // test that it does not panic.
		}
	}
}
