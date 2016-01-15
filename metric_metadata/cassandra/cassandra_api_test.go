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

package cassandra

import (
	"fmt"
	"sort"
	"testing"

	"github.com/square/metrics/api"
	"github.com/square/metrics/testing_support/assert"
)

var cassandraClean = true

func newCassandraAPI(t *testing.T) (*CassandraMetricMetadataAPI, api.MetricMetadataAPIContext) {
	if !cassandraClean {
		t.Fatalf("Attempted to create new database without cleaning up the old one.")
	}
	cassandraClean = false
	cassandraInterface, err := NewMetricMetadataInterface(Config{
		Hosts:    []string{"localhost"},
		Keyspace: "metrics_indexer_test",
	})
	if err != nil {
		t.Fatalf("Cannot instantiate Cassandra API: %s", err.Error())
	}
	cassandra := cassandraInterface.(*CassandraMetricMetadataAPI)

	tables := []string{"metric_names", "tag_index", "metric_name_set"}
	for _, table := range tables {
		// Truncate the tables
		if err := cassandra.db.session.Query(fmt.Sprintf("TRUNCATE %s", table)).Exec(); err != nil {
			t.Fatalf("Cannot truncate %s: %s", table, err.Error())
		}
	}
	return cassandra, api.MetricMetadataAPIContext{}
}

func cleanAPI(t *testing.T, c *CassandraMetricMetadataAPI) {
	cleanDatabase(t, &c.db)
	cassandraClean = true
}

func Test_MetricName_GetTagSet_API(t *testing.T) {
	a := assert.New(t)
	cassandra, context := newCassandraAPI(t)
	defer cleanAPI(t, cassandra)

	if _, err := cassandra.GetAllTags("sample", context); err == nil {
		t.Errorf("Cassandra API should error on fetching nonexistent metric")
	}

	metricNamesTests := []struct {
		addTest      bool
		metricName   string
		tagString    string
		expectedTags map[string][]string // { metricName: [ tags ] }
	}{
		{true, "sample", "foo=bar1", map[string][]string{
			"sample": []string{"foo=bar1"},
		}},
		{true, "sample", "foo=bar2", map[string][]string{
			"sample": []string{"foo=bar1", "foo=bar2"},
		}},
		{true, "sample2", "foo=bar2", map[string][]string{
			"sample":  []string{"foo=bar1", "foo=bar2"},
			"sample2": []string{"foo=bar2"},
		}},
		{false, "sample2", "foo=bar2", map[string][]string{
			"sample": []string{"foo=bar1", "foo=bar2"},
		}},
		{false, "sample", "foo=bar1", map[string][]string{
			"sample": []string{"foo=bar2"},
		}},
	}

	for _, c := range metricNamesTests {
		if c.addTest {
			a.CheckError(cassandra.AddMetric(api.TaggedMetric{
				api.MetricKey(c.metricName),
				api.ParseTagSet(c.tagString),
			}, context))
		} else {
			a.CheckError(cassandra.RemoveMetric(api.TaggedMetric{

				api.MetricKey(c.metricName),
				api.ParseTagSet(c.tagString),
			}, context))
		}

		for k, v := range c.expectedTags {
			if tags, err := cassandra.GetAllTags(api.MetricKey(k), context); err != nil {
				t.Errorf("Error fetching tags")
			} else {
				stringTags := make([]string, len(tags))
				for i, tag := range tags {
					stringTags[i] = tag.Serialize()
				}

				a.EqInt(len(stringTags), len(v))
				sort.Sort(sort.StringSlice(stringTags))
				sort.Sort(sort.StringSlice(v))
				a.Eq(stringTags, v)
			}
		}
	}
}

func Test_GetAllMetrics_API(t *testing.T) {
	a := assert.New(t)
	cassandra, context := newCassandraAPI(t)
	defer cleanAPI(t, cassandra)
	a.CheckError(cassandra.AddMetric(api.TaggedMetric{
		"metric.a",
		api.ParseTagSet("foo=a"),
	}, context))
	a.CheckError(cassandra.AddMetric(api.TaggedMetric{
		"metric.a",
		api.ParseTagSet("foo=b"),
	}, context))
	a.CheckError(cassandra.AddMetrics([]api.TaggedMetric{
		{
			"metric.c",
			api.TagSet{
				"bar": "cat",
			},
		},
		{
			"metric.d",
			api.TagSet{
				"bar": "dog",
			},
		},
		{
			"metric.e",
			api.TagSet{
				"bar": "cat",
			},
		},
	}, context))
	keys, err := cassandra.GetAllMetrics(context)
	a.CheckError(err)
	sort.Sort(api.MetricKeys(keys))
	a.Eq(keys, []api.MetricKey{"metric.a", "metric.c", "metric.d", "metric.e"})
	a.CheckError(cassandra.AddMetric(api.TaggedMetric{
		"metric.b",
		api.ParseTagSet("foo=c"),
	}, context))
	a.CheckError(cassandra.AddMetric(api.TaggedMetric{
		"metric.b",
		api.ParseTagSet("foo=c"),
	}, context))
	keys, err = cassandra.GetAllMetrics(context)
	a.CheckError(err)
	sort.Sort(api.MetricKeys(keys))
	a.Eq(keys, []api.MetricKey{"metric.a", "metric.b", "metric.c", "metric.d", "metric.e"})
}

func Test_TagIndex_API(t *testing.T) {
	a := assert.New(t)
	cassandra, context := newCassandraAPI(t)
	defer cleanAPI(t, cassandra)

	if rows, err := cassandra.GetMetricsForTag("environment", "production", context); err != nil {
		a.CheckError(err)
	} else {
		a.EqInt(len(rows), 0)
	}
	a.CheckError(cassandra.AddMetric(api.TaggedMetric{
		"a.b.c",
		api.TagSet{
			"environment": "production",
		},
	}, context))
	a.CheckError(cassandra.AddMetric(api.TaggedMetric{
		"d.e.f",
		api.TagSet{
			"environment": "production",
		},
	}, context))

	if rows, err := cassandra.GetMetricsForTag("environment", "production", context); err != nil {
		a.CheckError(err)
	} else {
		a.EqInt(len(rows), 2)
	}

	a.CheckError(cassandra.RemoveMetric(api.TaggedMetric{
		"a.b.c",
		api.TagSet{
			"environment": "production",
		},
	}, context))
	if rows, err := cassandra.GetMetricsForTag("environment", "production", context); err != nil {
		a.CheckError(err)
	} else {
		a.EqInt(len(rows), 1)
		a.EqString(string(rows[0]), "d.e.f")
	}
}
