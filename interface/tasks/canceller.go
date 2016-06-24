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

package tasks

import (
	"sync"
	"time"
)

// TimeoutOwner holds a Timeout object which it can cancel.
type TimeoutOwner struct {
	timeout *Timeout
}

// Timeout returns the Timeout object from the TimeoutOwner.
func (t TimeoutOwner) Timeout() *Timeout {
	return t.timeout
}

// Finish tells a TimeoutOwner to stop.
// If the TimeoutOwner is the zero value, this has no effect.
func (to TimeoutOwner) Finish() {
	if to.timeout == nil {
		return // No effect
	}
	to.timeout.mutex.Lock()
	defer to.timeout.mutex.Unlock()
	if to.timeout.done == nil {
		return
	}
	close(to.timeout.done)
	to.timeout.done = nil
}

// A Timeout is an object which can alert that an event should be cancelled.
type Timeout struct {
	mutex sync.Mutex
	done  chan struct{}
}

// Done returns a receive-only channel, which will close when the timeout ends.
func (t *Timeout) Done() <-chan struct{} {
	if t == nil {
		return nil
	}

	return t.done
}

// NewTimeout returns a TimeoutOwner and spawns a goroutine to cancel the timeout.
func NewTimeout(duration time.Duration) TimeoutOwner {
	owner := TimeoutOwner{
		&Timeout{
			mutex: sync.Mutex{},
			done:  make(chan struct{}),
		},
	}
	go func() {
		<-time.After(duration)
		owner.Finish()
	}()
	return owner
}
