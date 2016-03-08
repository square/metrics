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

package inspect

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/square/metrics/testing_support/assert"
)

func TestProfilerSimple(t *testing.T) {
	a := assert.New(t)

	utc := time.UTC
	now := time.Date(2015, 2, 17, 4, 35, 0, 0, utc)

	profiler := &Profiler{
		now: func() time.Time {
			return now
		},
		mutex:    sync.Mutex{},
		profiles: []Profile{},
	}
	// It's now possible to manipulate the internal now of the Profiler

	a.EqInt(len(profiler.All()), 0)

	finisher := profiler.Record("metricA")
	start, now := now, time.Date(2015, 2, 17, 6, 35, 0, 0, utc)
	finisher()

	list := profiler.All()
	a.EqInt(len(list), 1)
	a.EqString(list[0].Name, "metricA")
	if list[0].Start != start {
		t.Fatalf("expected to start at %+v but started at %+v", start, list[0].Start)
	}
	if list[0].Finish != now {
		t.Fatalf("expected to finish at %+v but finished at %+v", now, list[0].Finish)
	}

	// Changing the value of `now` shouldn't change the list.
	old, now := now, time.Date(2015, 3, 4, 6, 35, 0, 0, utc)

	list = profiler.All()
	a.EqInt(len(list), 1)
	a.EqString(list[0].Name, "metricA")
	if list[0].Start != start {
		t.Fatalf("expected to start at %+v but started at %+v", start, list[0].Start)
	}
	if list[0].Finish != old {
		t.Fatalf("expected to finish at %+v but finished at %+v", old, list[0].Finish)
	}

	count := 2000

	var wait sync.WaitGroup
	wait.Add(count)
	for i := 0; i < count; i++ {
		go func(i int) {
			profiler.Record(fmt.Sprintf("metric_%d", i))()
			wait.Done()
		}(i)
	}
	wait.Wait()
	list = profiler.All()
	a.EqInt(len(list), count+1)
	flushed := profiler.Flush()
	a.EqInt(len(flushed), count+1)
	flushed = profiler.Flush()
	a.EqInt(len(flushed), 0)
}
