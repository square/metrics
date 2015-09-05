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
	a.EqBool(c.Consume(5), true)
	a.EqInt(c.Current(), 5)
	a.EqBool(c.Consume(4), true)
	a.EqInt(c.Current(), 9)
	a.EqBool(c.Consume(1), true)
	a.EqInt(c.Current(), 10)
	a.EqBool(c.Consume(1), false)
	a.EqInt(c.Current(), 11)
}
