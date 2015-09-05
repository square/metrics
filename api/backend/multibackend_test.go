package backend

import (
	"sync"
	"testing"

	"github.com/square/metrics/api"
	"github.com/square/metrics/testing_support/assert"
)

// fake backend to control the rate of queries being done.
type fakeBackend struct {
	tickets chan struct{}
}

func (b fakeBackend) FetchSingleTimeseries(request api.FetchTimeseriesRequest) (api.Timeseries, error) {
	<-b.tickets
	return api.Timeseries{}, nil
}

type Suite struct {
	backend         fakeBackend
	waitGroup       sync.WaitGroup
	parallelWrapper api.ParallelTimeseriesStorageAPI
	cancellable     api.Cancellable
}

func newSuite() Suite {
	b := fakeBackend{make(chan struct{}, 10)}
	suite := Suite{
		backend:         b,
		parallelWrapper: NewParallelMultiBackend(b, 4),
		cancellable:     api.NewCancellable(),
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
		_, err := suite.parallelWrapper.FetchMultipleTimeseries(api.FetchMultipleTimeseriesRequest{
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
		_, err := suite.parallelWrapper.FetchMultipleTimeseries(api.FetchMultipleTimeseriesRequest{
			Metrics:     []api.TaggedMetric{api.TaggedMetric{"a", api.NewTagSet()}},
			Cancellable: suite.cancellable,
		})
		if err == nil {
			t.Errorf("Error expected, but got nil")
		} else {
			casted, ok := err.(api.TimeseriesStorageError)
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
