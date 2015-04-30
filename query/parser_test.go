package query

import (
	"github.com/square/metrics/assert"
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

func testFunction1() (string, string) {
	return functionName(0), functionName(1)

}
func TestFunctionName(t *testing.T) {
	a := assert.New(t)
	a.EqString(functionName(0), "TestFunctionName")
	first, second := testFunction1()
	a.EqString(first, "testFunction1")
	a.EqString(second, "TestFunctionName")
}
