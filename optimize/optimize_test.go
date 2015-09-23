package optimize

import (
	"testing"
	"time"

	"github.com/square/metrics/api"
)

func TestCaching(t *testing.T) {
	optimizer := NewOptimizationConfiguration()
	optimizer.EnableMetricMetadataCaching = true

	updateFunc := func() ([]api.TagSet, error) {
		// map[string]string
		result := []api.TagSet{api.NewTagSet()}
		return result, nil
	}
	someMetric := api.MetricKey("blah")
	optimizer.AllTagsCacheHitOrExecute(someMetric, updateFunc)

	updateFunc = func() ([]api.TagSet, error) {
		t.Errorf("Should not be called")
		return nil, nil
	}

	optimizer.AllTagsCacheHitOrExecute(someMetric, updateFunc)
}

func TestCacheExpiration(t *testing.T) {
	optimizer := NewOptimizationConfiguration()
	optimizer.EnableMetricMetadataCaching = true

	latch := false
	updateFunc := func() ([]api.TagSet, error) {
		// map[string]string
		latch = true
		result := []api.TagSet{api.NewTagSet()}
		return result, nil
	}
	someMetric := api.MetricKey("blah")
	optimizer.AllTagsCacheHitOrExecute(someMetric, updateFunc)
	if !latch {
		t.Errorf("We expected the update function to be called, but it wasn't")
	}
	optimizer.TimeSourceForNow = func() time.Time { return time.Now().Add(5 * time.Hour) }
	latch = false // Reset the latch

	optimizer.AllTagsCacheHitOrExecute(someMetric, updateFunc)
	if !latch {
		t.Errorf("We expected the update function to be called, but it wasn't")
	}
}
