// Copyright 2016 Square Inc.
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

package cached_metadata

import (
	"testing"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/testing_support/assert"
)

type testAPI struct {
	count    int
	proceed  chan struct{}
	finished chan struct{}
	data     map[api.MetricKey]string
}

// AddMetric waits for a slot to be open, then queries the underlying API.
func (c *testAPI) AddMetric(metric api.TaggedMetric, context api.MetricMetadataAPIContext) error {
	panic("unimplemented")
}

// AddMetrics waits for a slot to be open, then queries the underlying API.
func (c *testAPI) AddMetrics(metrics []api.TaggedMetric, context api.MetricMetadataAPIContext) error {
	panic("unimplemented")
}

// GetAllMetrics waits for a slot to be open, then queries the underlying API.
func (c *testAPI) GetAllMetrics(context api.MetricMetadataAPIContext) ([]api.MetricKey, error) {
	panic("unimplemented")
}

// GetMetricsForTag wwaits for a slot to be open, then queries the underlying API.
func (c *testAPI) GetMetricsForTag(tagKey, tagValue string, context api.MetricMetadataAPIContext) ([]api.MetricKey, error) {
	panic("unimplemented")
}

func (c *testAPI) GetAllTags(metricKey api.MetricKey, context api.MetricMetadataAPIContext) ([]api.TagSet, error) {
	// Wait for permission to proceed before returning.
	<-c.proceed
	defer func() { c.finished <- struct{}{} }()

	c.count++

	return []api.TagSet{
		{"foo": c.data[metricKey]},
	}, nil
}

func TestCached(t *testing.T) {
	a := assert.New(t)

	underlying := &testAPI{
		count:    0,
		proceed:  make(chan struct{}, 10),
		finished: make(chan struct{}, 10),
		data: map[api.MetricKey]string{
			"metric_one": "one",
			"metric_two": "two",
		},
	}
	cached := NewCachedMetricMetadataAPI(underlying, Config{time.Second, 1000})

	// Ask for metric_one
	{
		underlying.proceed <- struct{}{} // give it permission to go
		answer, err := cached.GetAllTags("metric_one", api.MetricMetadataAPIContext{})
		a.CheckError(err)
		a.Eq(answer[0]["foo"], "one")
		a.Contextf("underlying count").Eq(underlying.count, 1)
		<-underlying.finished
	}

	// Ask for metric_two
	{
		underlying.proceed <- struct{}{} // give it permission to go
		answer, err := cached.GetAllTags("metric_two", api.MetricMetadataAPIContext{})
		a.CheckError(err)
		a.Eq(answer[0]["foo"], "two")
		a.Contextf("underlying count").Eq(underlying.count, 2)
		<-underlying.finished
	}

	// Ask for metric_one again, but do not give permission to go.
	// Trying to verify that it's returning from the cache and not by querying the database.
	{
		answer, err := cached.GetAllTags("metric_one", api.MetricMetadataAPIContext{})
		a.CheckError(err)
		a.Eq(answer[0]["foo"], "one")
		a.Contextf("underlying count").Eq(underlying.count, 2)
	}

	// Now give the underlying permission to go:
	underlying.proceed <- struct{}{}
	// Now wait for it to finish
	<-underlying.finished

	a.Contextf("underlying count").Eq(underlying.count, 3)

	// Again on metric_two, but this time we'll update the original.
	underlying.data["metric_two"] = "new_two"
	{
		// It's in the cache- so it will return immediately
		answer, err := cached.GetAllTags("metric_two", api.MetricMetadataAPIContext{})
		a.CheckError(err)
		a.Eq(answer[0]["foo"], "two")

		a.Contextf("underlying count").Eq(underlying.count, 3)

		// Let the background update run.
		underlying.proceed <- struct{}{}
		<-underlying.finished

		a.Contextf("underlying count").Eq(underlying.count, 4)
	}

	// And again, but it should now be updated- even though the cache has yet to expire.
	// Again on metric_two, but this time we'll update the original.
	{
		// It's in the cache- so it will return immediately
		answer, err := cached.GetAllTags("metric_two", api.MetricMetadataAPIContext{})
		a.CheckError(err)
		a.Eq(answer[0]["foo"], "new_two")
		a.Contextf("underlying count").Eq(underlying.count, 4)

		// Let the background update run.
		underlying.proceed <- struct{}{}
		<-underlying.finished

		a.Contextf("underlying count").Eq(underlying.count, 5)
	}

}
