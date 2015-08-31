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

package api

import (
	"time"
)

type Cancellable interface {
	Done() chan struct{}         // returns a channel that is closed when the work is done.
	Deadline() (time.Time, bool) // deadline for the request.
}

type DefaultCancellable struct {
	done     chan struct{}
	deadline *time.Time
}

func (c DefaultCancellable) Done() chan struct{} {
	return c.done
}

func (c DefaultCancellable) Deadline() (time.Time, bool) {
	if c.deadline == nil {
		return time.Time{}, false
	} else {
		return *c.deadline, true
	}
}

func NewTimeoutCancellable(t time.Time) Cancellable {
	return DefaultCancellable{make(chan struct{}), &t}
}

func NewCancellable() Cancellable {
	return DefaultCancellable{make(chan struct{}), nil}
}
