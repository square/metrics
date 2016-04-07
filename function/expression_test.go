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

	"github.com/square/metrics/testing_support/assert"
)

func Test_FetchCounter(t *testing.T) {
	c := NewFetchCounter(10)
	a := assert.New(t)
	a.EqInt(c.Current(), 0)
	a.EqInt(c.Limit(), 10)
	a.CheckError(c.Consume(5))
	a.EqInt(c.Current(), 5)
	a.CheckError(c.Consume(4))
	a.EqInt(c.Current(), 9)
	a.CheckError(c.Consume(1))
	a.EqInt(c.Current(), 10)
	if c.Consume(1) == nil {
		a.Errorf("Expected there to be an error when consuming; but none was found")
	}
	a.EqInt(c.Current(), 11)
}
