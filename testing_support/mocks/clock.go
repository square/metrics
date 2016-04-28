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

package mocks

import "time"

type testClock struct {
	now time.Time
}

func NewTestClock(startTime time.Time) *testClock {
	return &testClock{
		now: startTime,
	}
}

func (t *testClock) Now() time.Time {
	return t.now
}

func (t *testClock) Move(diff time.Duration) {
	t.now = t.now.Add(diff)
}
