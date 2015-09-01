package backend

import (
	"sync"
	"testing"

	"github.com/square/metrics/api"
	"github.com/square/metrics/assert"
)

// fake backend to control the rate of queries being done.
type fakeBackend struct {
	tickets chan struct{}
}

func (b fakeBackend) FetchSingleSeries(request api.FetchSeriesRequest) (api.Timeseries, error) {
	<-b.tickets
	return api.Timeseries{}, nil
}

func (b fakeBackend) DecideTimerange(start int64, end int64, resolution int64) (api.Timerange, error) {
	return api.NewSnappedTimerange(start, end, resolution)
}

type Suite struct {
	backend      fakeBackend
	waitGroup    sync.WaitGroup
	multiBackend api.MultiBackend
	cancellable  api.Cancellable
}

func newSuite() Suite {
	b := fakeBackend{make(chan struct{}, 10)}
	suite := Suite{
		backend:      b,
		multiBackend: NewParallelMultiBackend(b, 4),
		cancellable:  api.NewCancellable(),
	}
	suite.waitGroup.Add(1)
	return suite
}

func (s Suite) cleanup() {
	close(s.backend.tickets)
}

func Test_ParallelMultiBackend_Success(t *testing.T) {
	a := assert.New(t)
	suite := newSuite()
	defer suite.cleanup()
	go func() {
		_, err := suite.multiBackend.FetchMultipleSeries(api.FetchMultipleRequest{
			Metrics:     []api.TaggedMetric{api.TaggedMetric{"a", api.NewTagSet()}},
			Cancellable: suite.cancellable,
		})
		a.CheckError(err)
		suite.waitGroup.Done()
	}()
	suite.backend.tickets <- struct{}{}
	suite.waitGroup.Wait()
}

func Test_ParallelMultiBackend_Timeout(t *testing.T) {
	a := assert.New(t)
	suite := newSuite()
	defer suite.cleanup()
	go func() {
		_, err := suite.multiBackend.FetchMultipleSeries(api.FetchMultipleRequest{
			Metrics:     []api.TaggedMetric{api.TaggedMetric{"a", api.NewTagSet()}},
			Cancellable: suite.cancellable,
		})
		if err == nil {
			t.Errorf("Error expected, but got nil")
		} else {
			casted, ok := err.(api.BackendError)
			if !ok {
				t.Errorf("Invalid error type")
			} else {
				a.Eq(casted.Code, api.FetchTimeoutError)
			}
		}
		suite.waitGroup.Done()
	}()
	close(suite.cancellable.Done())
	suite.waitGroup.Wait()
}
