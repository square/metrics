package internal

import (
	"github.com/gocql/gocql"
	"square/vis/metrics-indexer/api"
	"square/vis/metrics-indexer/assert"
	"testing"
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

func Test_AddMetricName_GetTagSet(t *testing.T) {
	a := assert.New(t)
	db := newDatabase(t)
	defer cleanDatabase(t, db)
	if db == nil {
		return
	}
	if tags, err := db.GetTagSet("sample"); err != nil {
		t.Errorf("Error while accessing cassandra.")
	} else {
		a.EqInt(len(tags), 0)
	}
	db.AddMetricName("sample", api.ParseTagSet("foo=bar1"))
	if tags, err := db.GetTagSet("sample"); err != nil {
		t.Errorf("Error while accessing cassandra.")
	} else {
		a.EqInt(len(tags), 1)
		a.EqString(tags[0].Serialize(), "foo=bar1")
	}
	db.AddMetricName("sample", api.ParseTagSet("foo=bar2"))
	if tags, err := db.GetTagSet("sample"); err != nil {
		t.Errorf("Error while accessing cassandra.")
	} else {
		a.EqInt(len(tags), 2)
	}
	db.AddMetricName("sample2", api.ParseTagSet("foo=bar2"))
	if tags, err := db.GetTagSet("sample"); err != nil {
		t.Errorf("Error while accessing cassandra.")
	} else {
		a.EqInt(len(tags), 2)
	}
}
