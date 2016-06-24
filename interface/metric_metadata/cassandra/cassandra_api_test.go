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

package cassandra

import (
	"fmt"
	"sort"
	"testing"

	"github.com/square/metrics/api"
	"github.com/square/metrics/interface/metric_metadata"
	"github.com/square/metrics/testing_support/assert"
)

func clearCassandraInstance(t *testing.T, db *cassandraDatabase, metricName api.MetricKey, tagString string) {
	assert.New(t).Contextf("clearing DB").CheckError(db.session.Query(
		"DELETE FROM metric_names WHERE metric_key = ? AND tag_set = ?",
		metricName,
		tagString,
	).Exec())
}

var cassandraClean = true

func newCassandraAPI(t *testing.T) (*MetricMetadataAPI, metadata.Context) {
	if !cassandraClean {
		t.Fatalf("Attempted to create new database without cleaning up the old one.")
	}
	cassandraClean = false
	cassandra, err := NewMetricMetadataAPI(Config{
		Hosts:    []string{"localhost"},
		Keyspace: "metrics_indexer_test",
	})
	if err != nil {
		t.Fatalf("Cannot instantiate Cassandra API: %s", err.Error())
	}

	tables := []string{"metric_names", "tag_index", "metric_name_set"}
	for _, table := range tables {
		// Truncate the tables
		if err := cassandra.db.session.Query(fmt.Sprintf("TRUNCATE %s", table)).Exec(); err != nil {
			t.Fatalf("Cannot truncate %s: %s", table, err.Error())
		}
	}
	return cassandra, metadata.Context{}
}

func cleanAPI(t *testing.T, c *MetricMetadataAPI) {
	cleanDatabase(t, &c.db)
	cassandraClean = true
}

func TestMetricNameGetTagSetAPI(t *testing.T) {
	a := assert.New(t)
	cassandra, context := newCassandraAPI(t)
	defer cleanAPI(t, cassandra)

	if _, err := cassandra.GetAllTags("sample", context); err == nil {
		t.Errorf("Cassandra API should error on fetching nonexistent metric")
	}

	metricNamesTests := []struct {
		addTest      bool
		metricName   api.MetricKey
		tagSet       api.TagSet
		expectedTags map[string][]api.TagSet // { metricName: [ tags ] }
	}{
		{true, "sample", api.TagSet{"foo": "bar1"}, map[string][]api.TagSet{
			"sample": {{"foo": "bar1"}},
		}},
		{true, "sample", api.TagSet{"foo": "bar2"}, map[string][]api.TagSet{
			"sample": {{"foo": "bar1"}, {"foo": "bar2"}},
		}},
		{true, "sample2", api.TagSet{"foo": "bar2"}, map[string][]api.TagSet{
			"sample":  {{"foo": "bar1"}, {"foo": "bar2"}},
			"sample2": {{"foo": "bar2"}},
		}},
		{false, "sample2", api.TagSet{"foo": "bar2"}, map[string][]api.TagSet{
			"sample": {{"foo": "bar1"}, {"foo": "bar2"}},
		}},
		{false, "sample", api.TagSet{"foo": "bar1"}, map[string][]api.TagSet{
			"sample": {{"foo": "bar2"}},
		}},
	}

	for _, c := range metricNamesTests {
		if c.addTest {
			a.CheckError(cassandra.AddMetric(api.TaggedMetric{
				api.MetricKey(c.metricName),
				c.tagSet,
			}, context))
		} else {
			clearCassandraInstance(t, &cassandra.db, c.metricName, c.tagSet.Serialize())
		}

		for metric, expected := range c.expectedTags {
			tags, err := cassandra.GetAllTags(api.MetricKey(metric), context)
			if err != nil {
				t.Errorf("Error fetching tags")
				continue
			}
			api.SortTagSets(tags)
			api.SortTagSets(expected)
			a.Contextf("GetAllTags(%q)", metric).Eq(tags, expected)
		}
	}
}

func TestGetAllMetricsAPI(t *testing.T) {
	a := assert.New(t)
	cassandra, context := newCassandraAPI(t)
	defer cleanAPI(t, cassandra)
	a.CheckError(cassandra.AddMetric(api.TaggedMetric{
		"metric.a",
		api.TagSet{"foo": "a"},
	}, context))
	a.CheckError(cassandra.AddMetric(api.TaggedMetric{
		"metric.a",
		api.TagSet{"foo": "b"},
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
		api.TagSet{"foo": "c"},
	}, context))
	a.CheckError(cassandra.AddMetric(api.TaggedMetric{
		"metric.b",
		api.TagSet{"foo": "c"},
	}, context))
	keys, err = cassandra.GetAllMetrics(context)
	a.CheckError(err)
	sort.Sort(api.MetricKeys(keys))
	a.Eq(keys, []api.MetricKey{"metric.a", "metric.b", "metric.c", "metric.d", "metric.e"})
}

func TestTagIndexAPI(t *testing.T) {
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
}
