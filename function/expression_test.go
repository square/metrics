package function

import (
	"testing"

	"github.com/square/metrics/testing_support/assert"
)

func Test_FetchCounter(t *testing.T) {
	c := NewFetchCounter(10)
	a := assert.New(t)
	a.EqInt(c.Current(), 0)
	a.EqInt(c.Limit(), 10)
	a.CheckError(c.Consume(5))
	a.EqInt(c.Current(), 5)
	a.CheckError(c.Consume(4))
	a.EqInt(c.Current(), 9)
	a.CheckError(c.Consume(1))
	a.EqInt(c.Current(), 10)
	if c.Consume(1) == nil {
		a.Errorf("Expected there to be an error when consuming; but none was found")
	}
	a.EqInt(c.Current(), 11)
}
