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

package cached

import (
	"errors"
	standard_log "log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/inspect/log"
	"github.com/square/metrics/inspect/log/standard"
	"github.com/square/metrics/interface/metric_metadata"
	"github.com/square/metrics/testing_support/assert"
	"github.com/square/metrics/testing_support/mocks"
)

type testAPI struct {
	count    int
	finished chan string
	data     map[api.MetricKey]string

	getAllTagsError error

	synchronize bool
	calledWG    sync.WaitGroup
	returnWG    sync.WaitGroup
}

// GetAllMetrics waits for a slot to be open, then queries the underlying API.
func (c *testAPI) GetAllMetrics(context metadata.Context) ([]api.MetricKey, error) {
	panic("unimplemented")
}

// GetMetricsForTag wwaits for a slot to be open, then queries the underlying API.
func (c *testAPI) GetMetricsForTag(tagKey, tagValue string, context metadata.Context) ([]api.MetricKey, error) {
	panic("unimplemented")
}

// CheckHealthy checks if the underlying MetricMetadataAPI is healthy
func (c *testAPI) CheckHealthy() error {
	panic("unimplemented")
}

func (c *testAPI) GetAllTags(metricKey api.MetricKey, context metadata.Context) ([]api.TagSet, error) {
	defer func() { c.finished <- string(metricKey) }()

	// Signal we've been called and wait for permission to continue
	if c.synchronize {
		c.calledWG.Done()
		c.returnWG.Wait()
	}

	c.count++

	if c.getAllTagsError != nil {
		return nil, c.getAllTagsError
	}

	return []api.TagSet{
		{"foo": c.data[metricKey]},
	}, nil
}

func TestCached(t *testing.T) {
	log.InitLogger(&standard.Logger{
		Logger: standard_log.New(os.Stderr, "", standard_log.LstdFlags),
	})
	log.Infof("Starting TestCached")
	defer log.Infof("Finished TestCached")

	a := assert.New(t)

	underlying := &testAPI{
		count:    0,
		finished: make(chan string, 10),
		data: map[api.MetricKey]string{
			"metric_one": "one",
			"metric_two": "two",
		},
	}
	cached := NewMetricMetadataAPI(underlying, Config{
		Freshness:    5 * time.Second,
		RequestLimit: 1000,
		TimeToLive:   10 * time.Second,
	}).(*metricMetadataAPI)
	clock := mocks.NewTestClock(time.Now())
	cached.clock = clock

	tags, err := cached.GetAllTags("metric_one", metadata.Context{})
	a.CheckError(err)
	a.Eq(tags, []api.TagSet{{"foo": "one"}})

	underlying.data["metric_one"] = "new one"

	tags, err = cached.GetAllTags("metric_one", metadata.Context{})
	a.CheckError(err)
	a.Eq(tags, []api.TagSet{{"foo": "one"}}) // read from cache

	// Advance the clock so the next call is stale
	clock.Move(6 * time.Second)

	tags, err = cached.GetAllTags("metric_one", metadata.Context{})
	a.CheckError(err)
	a.Eq(tags, []api.TagSet{{"foo": "one"}}) // still read from cache

	a.MustEqInt(cached.CurrentLiveRequests(), 1)
	a.CheckError(cached.GetBackgroundAction()(metadata.Context{})) // updates cache

	tags, err = cached.GetAllTags("metric_one", metadata.Context{})
	a.CheckError(err)
	// still read from cache, doesn't enqueue background since it's fresh
	a.Eq(tags, []api.TagSet{{"foo": "new one"}})

	// Advance the clock so the next call is expired
	clock.Move(11 * time.Second)

	tags, err = cached.GetAllTags("metric_one", metadata.Context{})
	a.CheckError(err)
	a.Eq(tags, []api.TagSet{{"foo": "new one"}})

	a.MustEqInt(cached.CurrentLiveRequests(), 0)

	underlying.data["metric_one"] = "ignore"

	// Advance the clock so the next call isn't stale yet
	clock.Move(3 * time.Second)

	tags, err = cached.GetAllTags("metric_one", metadata.Context{})
	a.CheckError(err)
	a.Eq(tags, []api.TagSet{{"foo": "new one"}})

	a.MustEqInt(cached.CurrentLiveRequests(), 0)

	// Advance the clock so the next call is stale
	clock.Move(3 * time.Second)

	tags, err = cached.GetAllTags("metric_one", metadata.Context{})
	a.CheckError(err)
	a.Eq(tags, []api.TagSet{{"foo": "new one"}})

	// Send another one to make sure we don't dupe the backend requests
	tags, err = cached.GetAllTags("metric_one", metadata.Context{})
	a.CheckError(err)
	a.Eq(tags, []api.TagSet{{"foo": "new one"}})

	a.MustEqInt(cached.CurrentLiveRequests(), 1)

	a.CheckError(cached.GetBackgroundAction()(metadata.Context{})) // cleanout the channel

	a.MustEqInt(cached.CurrentLiveRequests(), 0)
}

func TestCachedNoStale(t *testing.T) {
	log.InitLogger(&standard.Logger{
		Logger: standard_log.New(os.Stderr, "", standard_log.LstdFlags),
	})
	log.Infof("Starting TestCachedNoStale")
	defer log.Infof("Finished TestCachedNoStale")

	a := assert.New(t)

	underlying := &testAPI{
		count:    0,
		finished: make(chan string, 10),
		data: map[api.MetricKey]string{
			"metric_one": "one",
			"metric_two": "two",
		},
	}
	cached := NewMetricMetadataAPI(underlying, Config{
		RequestLimit: 1000,
		TimeToLive:   10 * time.Second,
	}).(*metricMetadataAPI)
	clock := mocks.NewTestClock(time.Now())
	cached.clock = clock

	tags, err := cached.GetAllTags("metric_one", metadata.Context{})
	a.CheckError(err)
	a.Eq(tags, []api.TagSet{{"foo": "one"}})

	underlying.data["metric_one"] = "new one"

	tags, err = cached.GetAllTags("metric_one", metadata.Context{})
	a.CheckError(err)
	a.Eq(tags, []api.TagSet{{"foo": "one"}}) // read from cache

	// Advance the clock so the next call is still fresh
	clock.Move(6 * time.Second)

	tags, err = cached.GetAllTags("metric_one", metadata.Context{})
	a.CheckError(err)
	a.Eq(tags, []api.TagSet{{"foo": "one"}}) // still read from cache

	a.MustEqInt(cached.CurrentLiveRequests(), 0)

	// Advance the clock so the next call is expired
	clock.Move(5 * time.Second)

	tags, err = cached.GetAllTags("metric_one", metadata.Context{})
	a.CheckError(err)
	a.Eq(tags, []api.TagSet{{"foo": "new one"}})

	a.MustEqInt(cached.CurrentLiveRequests(), 0)
}

// Specific testing around when a request is already inflight
func TestInflight(t *testing.T) {
	log.InitLogger(&standard.Logger{
		Logger: standard_log.New(os.Stderr, "", standard_log.LstdFlags),
	})
	log.Infof("Starting TestInflight")
	defer log.Infof("Finished TestInflight")

	// WaitGroups for making all the bg goroutines work
	var goWgMain, goWgTwo, goWgThree sync.WaitGroup

	a := assert.New(t)

	underlying := &testAPI{
		count:    0,
		finished: make(chan string, 10),
		data: map[api.MetricKey]string{
			"metric_one": "one",
			"metric_two": "two",
		},
		synchronize: true,
	}
	cached := NewMetricMetadataAPI(underlying, Config{
		RequestLimit: 1000,
		TimeToLive:   10 * time.Second,
	}).(*metricMetadataAPI)
	clock := mocks.NewTestClock(time.Now())
	cached.clock = clock

	a.MustEqInt(underlying.count, 0)

	// Going to spin up two goroutines
	goWgMain.Add(2)

	// Signal that we expect a call to happen and block the return
	underlying.calledWG.Add(1)
	underlying.returnWG.Add(1)

	// Routine One
	go func() {
		tags, err := cached.GetAllTags("metric_one", metadata.Context{})
		a.CheckError(err)
		a.Eq(tags, []api.TagSet{{"foo": "one"}})

		goWgMain.Done()
	}()

	// Wait here for Routine One to be blocked on the call to the underlying MetricMetadataAPI
	underlying.calledWG.Wait()

	// Routine Two
	goWgTwo.Add(1)
	go func() {
		goWgTwo.Done()
		tags, err := cached.GetAllTags("metric_one", metadata.Context{})
		a.CheckError(err)
		a.Eq(tags, []api.TagSet{{"foo": "one"}})
		goWgMain.Done()
	}()

	goWgTwo.Wait()

	a.MustEqInt(underlying.count, 0)

	// Allow the call in Routine One to return
	underlying.returnWG.Done()

	goWgMain.Wait()

	a.MustEqInt(underlying.count, 1)

	// Now check that we can send another request
	//////////////////////////////////////////////////////////////////////////////

	// Advance the clock so the next call is a cache miss
	clock.Move(11 * time.Second)

	// Signal that we expect a call to happen and block the return
	underlying.calledWG.Add(1)
	underlying.returnWG.Add(1)

	// Routine Three
	goWgMain.Add(1)
	goWgThree.Add(1)
	go func() {
		goWgThree.Done()
		tags, err := cached.GetAllTags("metric_one", metadata.Context{})
		a.CheckError(err)
		a.Eq(tags, []api.TagSet{{"foo": "one"}})
		goWgMain.Done()
	}()

	// Wait here for Routine Three to be blocked on the call to the underlying MetricMetadataAPI
	underlying.calledWG.Wait()

	// Then allow it to return
	underlying.returnWG.Done()

	goWgMain.Wait()

	a.MustEqInt(underlying.count, 2)
}

// Specific testing around when a request is already inflight and it errors
func TestInflightError(t *testing.T) {
	log.InitLogger(&standard.Logger{
		Logger: standard_log.New(os.Stderr, "", standard_log.LstdFlags),
	})
	log.Infof("Starting TestInflightError")
	defer log.Infof("Finished TestInflightError")

	// WaitGroups for making all the bg goroutines work
	var goWgMain, goWgTwo, goWgThree sync.WaitGroup

	var routineOneError, routineTwoError error

	a := assert.New(t)

	underlying := &testAPI{
		count:    0,
		finished: make(chan string, 10),
		data: map[api.MetricKey]string{
			"metric_one": "one",
			"metric_two": "two",
		},
		getAllTagsError: errors.New("uh oh"),
		synchronize:     true,
	}
	cached := NewMetricMetadataAPI(underlying, Config{
		RequestLimit: 1000,
		TimeToLive:   10 * time.Second,
	}).(*metricMetadataAPI)
	clock := mocks.NewTestClock(time.Now())
	cached.clock = clock

	a.MustEqInt(underlying.count, 0)

	// Going to spin up two goroutines
	goWgMain.Add(2)

	// Signal that we expect a call to happen and block the return
	underlying.calledWG.Add(1)
	underlying.returnWG.Add(1)

	// Routine One
	go func() {
		_, err := cached.GetAllTags("metric_one", metadata.Context{})
		routineOneError = err
		goWgMain.Done()
	}()

	// Wait here for Routine One to be blocked on the call to the underlying MetricMetadataAPI
	underlying.calledWG.Wait()

	// Routine Two
	goWgTwo.Add(1)
	go func() {
		goWgTwo.Done()
		_, err := cached.GetAllTags("metric_one", metadata.Context{})
		routineTwoError = err
		goWgMain.Done()
	}()

	goWgTwo.Wait()

	a.MustEqInt(underlying.count, 0)

	// Allow the call in Routine One to return
	underlying.returnWG.Done()

	goWgMain.Wait()

	a.MustEqInt(underlying.count, 1)

	if routineOneError == nil {
		t.Fatal("Expected error from routine one")
	}

	if routineTwoError == nil {
		t.Fatal("Expected error from routine one")
	}

	if routineOneError != routineTwoError {
		t.Fatal("Expected the same error from both routines")
	}

	// Now check that we can send another request
	//////////////////////////////////////////////////////////////////////////////

	// Advance the clock so the next call is a cache miss
	clock.Move(11 * time.Second)

	// Let's not error this time
	underlying.getAllTagsError = nil

	// Signal that we expect a call to happen and block the return
	underlying.calledWG.Add(1)
	underlying.returnWG.Add(1)

	// Routine Three
	goWgMain.Add(1)
	goWgThree.Add(1)
	go func() {
		goWgThree.Done()
		tags, err := cached.GetAllTags("metric_one", metadata.Context{})
		a.CheckError(err)
		a.Eq(tags, []api.TagSet{{"foo": "one"}})
		goWgMain.Done()
	}()

	// Wait here for Routine Three to be blocked on the call to the underlying MetricMetadataAPI
	underlying.calledWG.Wait()

	// Then allow it to return
	underlying.returnWG.Done()

	goWgMain.Wait()

	a.MustEqInt(underlying.count, 2)
}

// Specific testing around when a request is already inflight and the requests
// are stale
func TestStaleInflight(t *testing.T) {
	log.InitLogger(&standard.Logger{
		Logger: standard_log.New(os.Stderr, "", standard_log.LstdFlags),
	})
	log.Infof("Starting TestStaleInflight")
	defer log.Infof("Finished TestStaleInflight")

	a := assert.New(t)

	underlying := &testAPI{
		count:    0,
		finished: make(chan string, 10),
		data: map[api.MetricKey]string{
			"metric_one": "one",
			"metric_two": "two",
		},
	}
	cached := NewMetricMetadataAPI(underlying, Config{
		Freshness:    5 * time.Second,
		RequestLimit: 1000,
		TimeToLive:   10 * time.Second,
	}).(*metricMetadataAPI)
	clock := mocks.NewTestClock(time.Now())
	cached.clock = clock

	tags, err := cached.GetAllTags("metric_one", metadata.Context{})
	a.CheckError(err)
	a.Eq(tags, []api.TagSet{{"foo": "one"}})

	a.MustEqInt(underlying.count, 1)
	a.MustEqInt(cached.CurrentLiveRequests(), 0)

	// Now move the clock forward so the next requests are stale
	clock.Move(6 * time.Second)

	// We now want to control the specific execution flow
	underlying.synchronize = true

	// WaitGroups for making all the bg goroutines work
	var goWgMain sync.WaitGroup // goWgTwo, goWgThree

	// Going to spin up two goroutines
	goWgMain.Add(1)

	// This pulls from the cache but should enqueue a bg lookup
	tags, err = cached.GetAllTags("metric_one", metadata.Context{})
	a.CheckError(err)
	a.Eq(tags, []api.TagSet{{"foo": "one"}})

	a.MustEqInt(cached.CurrentLiveRequests(), 1)

	// Change the data
	underlying.data["metric_one"] = "new one"

	// Signal that we expect a call to happen (in the background) and block the return
	underlying.calledWG.Add(1)
	underlying.returnWG.Add(1)

	// Routine One
	go func() {
		a.CheckError(cached.GetBackgroundAction()(metadata.Context{}))
		goWgMain.Done()
	}()

	// Wait here for Routine One to be blocked on the call to the underlying MetricMetadataAPI
	underlying.calledWG.Wait()
	a.MustEqInt(cached.CurrentLiveRequests(), 0)

	// This pulls from the cache and should not enqueue a bg lookup
	tags, err = cached.GetAllTags("metric_one", metadata.Context{})
	a.CheckError(err)
	a.Eq(tags, []api.TagSet{{"foo": "one"}})

	a.MustEqInt(cached.CurrentLiveRequests(), 0)

	// Allow Routine One to finish
	underlying.returnWG.Done()

	goWgMain.Wait()

	// Make another call, the cache is now fresh
	tags, err = cached.GetAllTags("metric_one", metadata.Context{})
	a.CheckError(err)
	a.Eq(tags, []api.TagSet{{"foo": "new one"}})
}

func TestQueueSize(t *testing.T) {
	log.InitLogger(&standard.Logger{
		Logger: standard_log.New(os.Stderr, "", standard_log.LstdFlags),
	})
	log.Infof("Starting TestQueueSize")
	defer log.Infof("Finished TestQueueSize")

	a := assert.New(t)

	underlying := &testAPI{
		count:    0,
		finished: make(chan string, 10),
		data: map[api.MetricKey]string{
			"metric_one":   "one",
			"metric_two":   "two",
			"metric_three": "three",
			"metric_four":  "four",
		},
	}
	cached := NewMetricMetadataAPI(underlying, Config{
		Freshness:    5 * time.Second,
		RequestLimit: 3,
		TimeToLive:   10 * time.Second,
	}).(*metricMetadataAPI)
	clock := mocks.NewTestClock(time.Now())
	cached.clock = clock

	// Prime the cache
	_, err := cached.GetAllTags("metric_one", metadata.Context{})
	a.CheckError(err)

	_, err = cached.GetAllTags("metric_two", metadata.Context{})
	a.CheckError(err)

	// Advance the clock so that metric_one and metric_two are stale
	clock.Move(6 * time.Second)

	// Stale entries
	_, err = cached.GetAllTags("metric_one", metadata.Context{})
	a.CheckError(err)

	_, err = cached.GetAllTags("metric_two", metadata.Context{})
	a.CheckError(err)

	a.MustEqInt(cached.CurrentLiveRequests(), 2)

	_, err = cached.GetAllTags("metric_three", metadata.Context{})
	a.CheckError(err)

	// Advance the clock so that metric_three is stale
	clock.Move(6 * time.Second)

	_, err = cached.GetAllTags("metric_three", metadata.Context{})
	a.CheckError(err)

	a.MustEqInt(cached.CurrentLiveRequests(), 3)

	// Adding another one should not increase the number of requests,
	// and it shouldn't cause this call to block.

	_, err = cached.GetAllTags("metric_four", metadata.Context{})
	a.CheckError(err)

	// Advance the clock so that metric_four is stale
	clock.Move(6 * time.Second)

	for i := 0; i < 100; i++ {
		_, err = cached.GetAllTags("metric_four", metadata.Context{})
		a.CheckError(err)
		a.MustEqInt(cached.CurrentLiveRequests(), 3)
	}

	a.CheckError(cached.GetBackgroundAction()(metadata.Context{}))
	a.CheckError(cached.GetBackgroundAction()(metadata.Context{}))
	a.CheckError(cached.GetBackgroundAction()(metadata.Context{}))

	a.MustEqInt(cached.CurrentLiveRequests(), 0)
}
