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

import (
	"github.com/square/metrics/assert"
	"testing"
)

func TestUnescapeLiteral(t *testing.T) {
	a := assert.New(t)
	a.EqString(unescapeLiteral("'foo'"), "foo")
	a.EqString(unescapeLiteral("foo"), "foo")
	a.EqString(unescapeLiteral("nodes.cpu.io"), "nodes.cpu.io")
	a.EqString(unescapeLiteral(`"hello"`), `hello`)
	a.EqString(unescapeLiteral(`"\"hello\""`), `"hello"`)
	a.EqString(unescapeLiteral(`'\"hello\"'`), `"hello"`)
	a.EqString(unescapeLiteral("\"\\`\""), "`")
}

func testFunction1() (string, string) {
	return functionName(0), functionName(1)
}

func TestFunctionName(t *testing.T) {
	a := assert.New(t)
	a.EqString(functionName(0), "TestFunctionName")
	first, second := testFunction1()
	a.EqString(first, "testFunction1")
	a.EqString(second, "TestFunctionName")
}
