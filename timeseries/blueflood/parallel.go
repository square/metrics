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

	"github.com/square/metrics/api"
	"github.com/square/metrics/tasks"
)

// ChooseResolution will use the finest-grained permissible resolution at least as coarse as the requested,
// unless no such timerange exists.
// Thus, if you request too many points, it will automatically reduce the resolution.
func (b *Blueflood) ChooseResolution(requested api.Timerange, lowerBound time.Duration) (time.Duration, error) {
	oldestAge := b.config.TimeSource().Sub(requested.Start())
	youngestAge := b.config.TimeSource().Sub(requested.End())
	for _, resolution := range b.config.Resolutions {
		if resolution.TimeToLive < oldestAge {
			// Doesn't live long enough.
			continue
		}
		if resolution.FirstAvailable > youngestAge {
			// Doesn't exist soon enough.
			continue
		}
		if resolution.Resolution < lowerBound {
			// Resolution is too high (which would produce too many points).
			continue
		}
		return resolution.Resolution, nil
	}
	return 0, fmt.Errorf("")
}

type ticket struct{}

type ParallelQueue struct {
	limit     int
	tickets   chan ticket
	timeout   *tasks.Timeout
	waitgroup sync.WaitGroup
}

func (b *ParallelQueue) Do(f func()) {
	b.waitgroup.Add(1)
	go func() {
		defer b.waitgroup.Done()
		select {
		case <-b.tickets:
			defer func() { b.tickets <- ticket{} }()
			f()
		case <-b.timeout.Done():
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
		return nil
	}
}
