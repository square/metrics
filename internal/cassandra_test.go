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

package internal

import (
	"sort"
	"testing"

	"github.com/gocql/gocql"
	"github.com/square/metrics/api"
	"github.com/square/metrics/assert"
)

func newDatabase(t *testing.T) *defaultDatabase {
	cluster := gocql.NewCluster("localhost")
	cluster.Keyspace = "metrics_indexer_test"
	session, err := cluster.CreateSession()
	if err != nil {
		t.Errorf("Cannot connect to Cassandra")
		return nil
	}
	if session.Query("TRUNCATE metric_names").Exec() != nil {
		t.Errorf("Cannot truncate")
		return nil
	}
	if session.Query("TRUNCATE tag_index").Exec() != nil {
		t.Errorf("Cannot truncate")
		return nil
	}
	return &defaultDatabase{session}
}

func cleanDatabase(t *testing.T, db *defaultDatabase) {
	db.session.Close()
}

func Test_MetricName_GetTagSet(t *testing.T) {
	a := assert.New(t)
	db := newDatabase(t)
	if db == nil {
		return
	}
	defer cleanDatabase(t, db)
	if db == nil {
		return
	}
	if tags, err := db.GetTagSet("sample"); err != nil {
		t.Errorf("Error fetching tags from Cassandra")
	} else {
		a.EqInt(len(tags), 0)
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
			a.CheckError(db.AddMetricNames([]api.TaggedMetric{{api.MetricKey(c.metricName), api.ParseTagSet(c.tagString)}}))
		} else {
			a.CheckError(db.RemoveMetricName(api.MetricKey(c.metricName), api.ParseTagSet(c.tagString)))
		}

		for k, v := range c.expectedTags {
			if tags, err := db.GetTagSet(api.MetricKey(k)); err != nil {
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

func Test_GetAllMetrics(t *testing.T) {
	a := assert.New(t)
	db := newDatabase(t)
	if db == nil {
		return
	}
	defer cleanDatabase(t, db)
	a.CheckError(db.AddMetricNames([]api.TaggedMetric{{"metric.a", api.ParseTagSet("foo=a")}, {"metric.a", api.ParseTagSet("foo=b")}}))
	keys, err := db.GetAllMetrics()
	a.CheckError(err)
	sort.Sort(api.MetricKeys(keys))
	a.Eq(keys, []api.MetricKey{"metric.a"})
	a.CheckError(db.AddMetricNames([]api.TaggedMetric{{"metric.b", api.ParseTagSet("foo=c")}, {"metric.b", api.ParseTagSet("foo=c")}}))
	keys, err = db.GetAllMetrics()
	a.CheckError(err)
	sort.Sort(api.MetricKeys(keys))
	a.Eq(keys, []api.MetricKey{"metric.a", "metric.b"})
}

func Test_TagIndex(t *testing.T) {
	a := assert.New(t)
	db := newDatabase(t)
	if db == nil {
		return
	}
	defer cleanDatabase(t, db)
	if db == nil {
		return
	}
	if rows, err := db.GetMetricKeys("environment", "production"); err != nil {
		a.CheckError(err)
	} else {
		a.EqInt(len(rows), 0)
	}
	db.AddToTagIndex(api.TaggedMetric{"a.b.c", api.TagSet{"environment": "production"}})
	db.AddToTagIndex(api.TaggedMetric{"d.e.f", api.TagSet{"environment": "production"}})
	if rows, err := db.GetMetricKeys("environment", "production"); err != nil {
		a.CheckError(err)
	} else {
		a.EqInt(len(rows), 2)
	}

	a.CheckError(db.RemoveFromTagIndex("environment", "production", "a.b.c"))
	if rows, err := db.GetMetricKeys("environment", "production"); err != nil {
		a.CheckError(err)
	} else {
		a.EqInt(len(rows), 1)
		a.EqString(string(rows[0]), "d.e.f")
	}
}
