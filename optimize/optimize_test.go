package optimize

import (
	"testing"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/testing_support/assert"
	"github.com/square/metrics/testing_support/mocks"
	"github.com/square/metrics/util"
)

type pingAPI struct {
	ping map[api.MetricKey]int
	api.MetricMetadataAPI
}

func (pingAPI *pingAPI) GetAllTags(key api.MetricKey, context api.MetricMetadataAPIContext) ([]api.TagSet, error) {
	pingAPI.ping[key]++
	return pingAPI.MetricMetadataAPI.GetAllTags(key, context)
}

func TestCachingAPI(t *testing.T) {
	a := assert.New(t)

	now := time.Now()
	nowFunc := func() time.Time { return now }
	fakeGraphiteConverter := &mocks.FakeGraphiteConverter{
		MetricMap: map[util.GraphiteMetric]api.TaggedMetric{},
	}
	mock := mocks.NewFakeMetricMetadataAPI()
	pingAPI := &pingAPI{
		ping:              map[api.MetricKey]int{},
		MetricMetadataAPI: mock,
	}
	cached := &MetadataAPICache{
		MetricMetadataAPI: pingAPI,
		cacheAllTags:      map[api.MetricKey]tagSetExpiry{},
		ttl:               time.Minute * 7,
		now:               nowFunc,
	}
	mock.AddPair(api.TaggedMetric{"blah", api.TagSet{"app": "mqe"}}, "ignore1", fakeGraphiteConverter)
	mock.AddPair(api.TaggedMetric{"blah", api.TagSet{"app": "server"}}, "ignore2", fakeGraphiteConverter)
	mock.AddPair(api.TaggedMetric{"second", api.TagSet{"host": "hostA"}}, "ignore3", fakeGraphiteConverter)
	mock.AddPair(api.TaggedMetric{"second", api.TagSet{"host": "hostB"}}, "ignore4", fakeGraphiteConverter)
	mock.AddPair(api.TaggedMetric{"second", api.TagSet{"host": "hostC"}}, "ignore5", fakeGraphiteConverter)
	mock.AddPair(api.TaggedMetric{"third", api.TagSet{"foo": "bar"}}, "ignore6", fakeGraphiteConverter)
	mock.AddPair(api.TaggedMetric{"third", api.TagSet{"foo": "qux"}}, "ignore7", fakeGraphiteConverter)

	// This will increase ping[`blah`]

	if result, err := cached.GetAllTags(api.MetricKey("blah"), api.MetricMetadataAPIContext{}); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	} else {
		a.Eq(result, []api.TagSet{{"app": "mqe"}, {"app": "server"}})
	}
	if pingAPI.ping["blah"] != 1 {
		t.Errorf("Expected ping[`blah`] to be 1, but got %d", pingAPI.ping["blah"])
	}

	// This won't increase ping[`blah`]
	now = now.Add(3 * time.Minute)
	if result, err := cached.GetAllTags(api.MetricKey("blah"), api.MetricMetadataAPIContext{}); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	} else {
		a.Eq(result, []api.TagSet{{"app": "mqe"}, {"app": "server"}})
	}
	a.Contextf("ping[`blah`]").EqInt(pingAPI.ping["blah"], 1)

	// This will increase ping[`blah`] since our TTL is only 7 minutes (13 have passed)
	now = now.Add(10 * time.Minute)
	if result, err := cached.GetAllTags(api.MetricKey("blah"), api.MetricMetadataAPIContext{}); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	} else {
		a.Eq(result, []api.TagSet{{"app": "mqe"}, {"app": "server"}})
	}
	a.Contextf("ping[`blah`]").EqInt(pingAPI.ping["blah"], 2)

	// Now hit "second" and "blah" again
	if result, err := cached.GetAllTags(api.MetricKey("second"), api.MetricMetadataAPIContext{}); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	} else {
		a.Eq(result, []api.TagSet{{"host": "hostA"}, {"host": "hostB"}, {"host": "hostC"}})
	}
	a.Contextf("ping[`second`]").EqInt(pingAPI.ping["second"], 1)

	if result, err := cached.GetAllTags(api.MetricKey("second"), api.MetricMetadataAPIContext{}); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	} else {
		a.Eq(result, []api.TagSet{{"host": "hostA"}, {"host": "hostB"}, {"host": "hostC"}})
	}
	a.Contextf("ping[`second`]").EqInt(pingAPI.ping["second"], 1)

	if result, err := cached.GetAllTags(api.MetricKey("blah"), api.MetricMetadataAPIContext{}); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	} else {
		a.Eq(result, []api.TagSet{{"app": "mqe"}, {"app": "server"}})
	}
	a.Contextf("ping[`blah`]").EqInt(pingAPI.ping["blah"], 2)
}
