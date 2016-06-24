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

// A Clock can give you the current time.Now(), or you can mock it out.
// Its zero value will
type Clock struct {
	Offset  time.Duration
	NowFunc func() time.Time
}

// Now returns either time.Now(), or the configured overriden NowFunc.
func (c *Clock) Now() time.Time {
	if c == nil {
		return time.Now()
	}
	if c.NowFunc == nil {
		return time.Now().Add(c.Offset)
	}
	return c.NowFunc().Add(c.Offset)
}
func (c *Clock) Move(offset time.Duration) {
	c.Offset += offset
}
