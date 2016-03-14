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

package tasks

import (
	"time"
)

type TimeoutOwner struct {
	Timeout
}

func (to TimeoutOwner) Finish() {
	if to.Timeout.done == nil {
		return
	}
	close(to.Timeout.done)
}

type Timeout struct {
	done     chan struct{}
	deadline time.Time
}

func (t Timeout) Done() <-chan struct{} {
	return t.done
}

func (t Timeout) Deadline() (time.Time, bool) {
	if t.done == nil {
		return time.Time{}, false
	}
	return t.deadline, true
}

type timeout struct {
	done     chan struct{}
	deadline *time.Time
}

func NewTimeout(deadline time.Time) TimeoutOwner {
	return TimeoutOwner{
		Timeout{
			done:     make(chan struct{}),
			deadline: deadline,
		},
	}
}
