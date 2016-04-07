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

package function

import (
	"testing"
	"time"
)

func TestToDuration(t *testing.T) {
	helper := func(given string, expected int64) {
		actual, err := StringToDuration(given)
		if err != nil || actual != time.Duration(expected)*time.Millisecond {
			t.Fatalf("Expected %s to produce %d but got %d", given, expected, actual)
		}
	}
	// Verify that all of the following produce the expected result:
	helper("7ms", 7)
	helper("7s", 7000)
	helper("7m", 7000*60)
	helper("7h", 7000*60*60)
	helper("7hr", 7000*60*60)
	helper("7d", 7000*60*60*24)
	helper("7w", 7000*60*60*24*7)
	helper("7M", 7000*60*60*24*30)
	helper("7mo", 7000*60*60*24*30)
	helper("7y", 7000*60*60*24*365)
	helper("7yr", 7000*60*60*24*365)

	helper("-7ms", -7)
	helper("-7s", -7000)
	helper("-7m", -7000*60)
	helper("-7h", -7000*60*60)
	helper("-7hr", -7000*60*60)
	helper("-7d", -7000*60*60*24)
	helper("-7w", -7000*60*60*24*7)
	helper("-7M", -7000*60*60*24*30)
	helper("-7mo", -7000*60*60*24*30)
	helper("-7y", -7000*60*60*24*365)
	helper("-7yr", -7000*60*60*24*365)
}
