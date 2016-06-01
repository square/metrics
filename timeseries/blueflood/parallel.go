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

package blueflood

import (
	"fmt"
	"sync"
	"time"

	"github.com/square/metrics/tasks"
)

type ticket struct{}

type ParallelQueue struct {
	limit     int
	tickets   chan ticket
	timeout   *tasks.Timeout
	waitgroup sync.WaitGroup
	sync.Mutex
	errorResult error
}

func (b *ParallelQueue) FlagError(err error) {
	if err == nil {
		return
	}
	b.Lock()
	defer b.Unlock()
	b.errorResult = err
}

func (b *ParallelQueue) Do(f func() error) {
	b.waitgroup.Add(1)
	go func() {
		defer b.waitgroup.Done()
		select {
		case <-b.tickets:
			defer func() { b.tickets <- ticket{} }()
			b.FlagError(f())
		case <-b.timeout.Done():
			b.FlagError(fmt.Errorf("timeout"))
		}
	}()
}

func (b *ParallelQueue) Wait() error {
	done := make(chan ticket, 1)
	go func() {
		b.waitgroup.Wait()
		done <- ticket{}
	}()
	select {
	case <-b.timeout.Done():
		return fmt.Errorf("timeout")
	case <-done:
		b.Lock()
		defer b.Unlock()
		return b.errorResult
	}
}

func NewParallelQueue(tickets int, timeout time.Duration) *ParallelQueue {
	ticketChannel := make(chan ticket, tickets)
	for i := 0; i < tickets; i++ {
		ticketChannel <- ticket{}
	}
	return &ParallelQueue{
		tickets: ticketChannel,
		timeout: tasks.NewTimeout(timeout).Timeout(),
	}
}
