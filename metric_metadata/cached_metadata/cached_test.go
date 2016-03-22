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
	finished chan string
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
	defer func() { c.finished <- string(metricKey) }()
	// Wait for permission to proceed before returning.

	c.count++

	return []api.TagSet{
		{"foo": c.data[metricKey]},
	}, nil
}

func TestCached(t *testing.T) {
	a := assert.New(t)

	underlying := &testAPI{
		count:    0,
		finished: make(chan string, 10),
		data: map[api.MetricKey]string{
			"metric_one": "one",
			"metric_two": "two",
		},
	}
	cached := NewCachedMetricMetadataAPI(underlying, Config{time.Second, 1000})

	tags, err := cached.GetAllTags("metric_one", api.MetricMetadataAPIContext{})

	a.CheckError(err)
	a.Eq(tags, []api.TagSet{{"foo": "one"}})

	underlying.data["metric_one"] = "new one"

	tags, err = cached.GetAllTags("metric_one", api.MetricMetadataAPIContext{})
	a.CheckError(err)
	a.Eq(tags, []api.TagSet{{"foo": "one"}}) // read from cache

	tags, err = cached.GetAllTags("metric_one", api.MetricMetadataAPIContext{})
	a.CheckError(err)
	a.Eq(tags, []api.TagSet{{"foo": "one"}}) // still read from cache

	a.CheckError(cached.GetBackgroundAction()(api.MetricMetadataAPIContext{})) // updates cache

	tags, err = cached.GetAllTags("metric_one", api.MetricMetadataAPIContext{})
	a.CheckError(err)
	a.Eq(tags, []api.TagSet{{"foo": "new one"}}) // still read from cache

	a.EqInt(cached.CurrentLiveRequests(), 2)

	a.CheckError(cached.GetBackgroundAction()(api.MetricMetadataAPIContext{})) // updates cache

	a.EqInt(cached.CurrentLiveRequests(), 1)

	a.CheckError(cached.GetBackgroundAction()(api.MetricMetadataAPIContext{})) // updates cache

	a.EqInt(cached.CurrentLiveRequests(), 0)

	underlying.data["metric_one"] = "ignore"

	tags, err = cached.GetAllTags("metric_one", api.MetricMetadataAPIContext{})
	a.CheckError(err)
	a.Eq(tags, []api.TagSet{{"foo": "new one"}})
}

func TestQueueSize(t *testing.T) {
	a := assert.New(t)

	underlying := &testAPI{
		count:    0,
		finished: make(chan string, 10),
		data: map[api.MetricKey]string{
			"metric_one": "one",
			"metric_two": "two",
		},
	}
	cached := NewCachedMetricMetadataAPI(underlying, Config{time.Second, 3})

	_, err := cached.GetAllTags("metric_one", api.MetricMetadataAPIContext{})
	a.CheckError(err)

	_, err = cached.GetAllTags("metric_one", api.MetricMetadataAPIContext{})
	a.CheckError(err)

	_, err = cached.GetAllTags("metric_one", api.MetricMetadataAPIContext{})
	a.CheckError(err)
	a.EqInt(cached.CurrentLiveRequests(), 2)

	_, err = cached.GetAllTags("metric_one", api.MetricMetadataAPIContext{})
	a.CheckError(err)
	a.EqInt(cached.CurrentLiveRequests(), 3)

	// Adding another one should not increase the number of requests,
	// and it shouldn't cause this call to block.

	for i := 0; i < 100; i++ {
		_, err = cached.GetAllTags("metric_one", api.MetricMetadataAPIContext{})
		a.CheckError(err)
		a.EqInt(cached.CurrentLiveRequests(), 3)
	}

	a.CheckError(cached.GetBackgroundAction()(api.MetricMetadataAPIContext{}))
	a.CheckError(cached.GetBackgroundAction()(api.MetricMetadataAPIContext{}))
	a.CheckError(cached.GetBackgroundAction()(api.MetricMetadataAPIContext{}))

	a.EqInt(cached.CurrentLiveRequests(), 0)

}
