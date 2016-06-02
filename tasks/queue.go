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
	"fmt"
	"sync"
	"time"
)

// A TimeoutError is an error associated with a timeout.
// It can be queried for the time in question.
type TimeoutError interface {
	error
	Timeout() time.Duration
}

type timeoutError struct {
	duration time.Duration
}

func (t timeoutError) Error() string {
	return fmt.Sprintf("Timeout after %+v", t.duration)
}

func (t timeoutError) Timeout() time.Duration {
	return t.duration
}

func NewTimeoutError(duration time.Duration) TimeoutError {
	return timeoutError{duration: duration}
}

type ticket struct{}

// ParallelQueue is a queue of actions which are rate-limited by a specified
// maximum amount of concurrency.
type ParallelQueue struct {
	sync.Mutex // ParallelQueue uses itself as a mutex, and clients can also use it as a mutex.

	timeoutTime time.Duration // how long before it times out

	tickets   chan ticket    // tickets internally limit the number of simultaneous actions
	timeout   *Timeout       // timeout is used to stop the queue
	waitgroup sync.WaitGroup // the waitgroup synchronizes the actions

	errorResult       error       // errorResult holds an execution error.
	errorNotification chan ticket // errorNotification receives a ticket after errorResult has been set.
}

// FlagError sets the ParallelQueue's error if it hasn't already been set.
// If err is nil, it has no effect.
// This method is safe to call concurrently; it's synchronized by the queue's mutex.
func (q *ParallelQueue) FlagError(err error) {
	if err == nil {
		return
	}
	q.Lock()
	defer q.Unlock()
	if q.errorResult != nil {
		return
	}
	q.errorResult = err
	q.errorNotification <- ticket{}
}

// Do executes the given func. If it returns an error, the queue will flag the error on its.
// It immediately spawns a goroutine that will wait until a ticket is available or the timeout is reached.
func (q *ParallelQueue) Do(f func() error) {
	q.waitgroup.Add(1)
	go func() {
		defer q.waitgroup.Done()
		select {
		case <-q.tickets:
			defer func() { q.tickets <- ticket{} }()
			q.FlagError(f())
		case <-q.timeout.Done():
			q.FlagError(NewTimeoutError(q.timeoutTime))
		}
	}()
}

// Wait blocks until all actions spawned with "Do" have run to completion OR the timeout has been reached,
// OR one of the actions returned an error.
// It returns an error if a timeout occurred or if the actions return an error.
func (q *ParallelQueue) Wait() error {
	done := make(chan ticket, 1)
	go func() {
		q.waitgroup.Wait()
		done <- ticket{}
	}()
	select {
	case <-q.timeout.Done():
		return NewTimeoutError(q.timeoutTime)
	case <-q.errorNotification:
		q.Lock()
		defer q.Unlock()
		return q.errorResult
	case <-done:
		q.Lock()
		defer q.Unlock()
		return q.errorResult
	}
}

// NewParallelQueue creates a ParallelQueue with the given number of tickets whose timeout is the specified timeout.
func NewParallelQueue(tickets int, timeout time.Duration) *ParallelQueue {
	ticketChannel := make(chan ticket, tickets)
	for i := 0; i < tickets; i++ {
		ticketChannel <- ticket{}
	}
	return &ParallelQueue{
		timeoutTime:       timeout,
		tickets:           ticketChannel,
		timeout:           NewTimeout(timeout).Timeout(),
		errorNotification: make(chan ticket, 1),
	}
}
