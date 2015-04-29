package query

import (
	"github.com/square/metrics-indexer/assert"
	"testing"
)

func TestUnescapeLiteral(t *testing.T) {
	a := assert.New(t)
	a.EqString(unescapeLiteral("'foo'"), "foo")
	a.EqString(unescapeLiteral("foo"), "foo")
	a.EqString(unescapeLiteral("nodes.cpu.io"), "nodes.cpu.io")
	a.EqString(unescapeLiteral(`"hello"`), `hello`)
	a.EqString(unescapeLiteral(`"\"hello\""`), `"hello"`)
	a.EqString(unescapeLiteral(`'\"hello\"'`), `"hello"`)
	a.EqString(unescapeLiteral("\"\\`\""), "`")
}
