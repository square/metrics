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

package parser

import (
	"testing"
	"time"

	"github.com/square/metrics/testing_support/assert"
)

func Test_parseRelativeTime(t *testing.T) {
	now := time.Unix(1413321866, 0).UTC()

	// Valid relative timestamps
	timestampTests := []struct {
		timeString        string
		expectedTimestamp int64
		expectSuccess     bool
	}{
		// Valid relative timestamps
		{"-2s", 1413321864000, true},
		{"-3m", 1413321686000, true},
		{"-4h", 1413307466000, true},
		{"-5d", 1412889866000, true},
		{"-3w", 1411507466000, true},
		{"-1M", 1410729866000, true},
		{"-1y", 1381785866000, true},
		{"1s", 1413321867000, true},
		{"+1s", 1413321867000, true},
		{"5d", 1413753866000, true},
		// Bad relative timestamps
		{"5dd", -1, false},
		{"-5dd", -1, false},
		{"-5z", -1, false},
	}

	for _, c := range timestampTests {
		ts, err := parseDate(c.timeString, now)
		if err != nil && c.expectSuccess {
			t.Fatal("Received unexpected error from parseRelativeTime: ", err)
		}

		if ts != c.expectedTimestamp {
			t.Fatalf("Expected %d but received %d", c.expectedTimestamp, ts)
		}
	}
}

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
