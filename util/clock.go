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

package util

import "time"

// Clock provides the functions from the time package. Exists so we can mock
// out time.Now
type Clock interface {
	// Now returns the current local time.
	Now() time.Time
}

// RealClock is a wrapper over the time package
type RealClock struct{}

// Now returns the current time.Time
func (r RealClock) Now() time.Time {
	return time.Now()
}
