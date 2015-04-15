package query

// show all metrics
// show tags WHERE predicate

import (
	"github.com/square/metrics-indexer/assert"
	"testing"
)

var inputs = []string{
	"describe all metrics",
	"describe x",
	"describe cpu_usage",
	"describe cpu_usage where key = 'value'",
	"describe cpu_usage where key = 'value\\''",
	"describe cpu_usage where key != 'value'",
	"describe cpu_usage where (key = 'value')",
	"describe cpu_usage where not (key = 'value')",
	"describe cpu_usage where not key = 'value'",
	"describe cpu_usage where (key = 'value' and key = 'value')",
	"describe cpu_usage where key = 'value' or key = 'value'",
	"describe cpu_usage where name:key in ('value')",
	"describe cpu_usage where key in ('value', 'value')",
	"describe cpu_usage where key matches 'abc'",
	"describe nodes.cpu.usage where datacenter='sjc1b' and type='idle' and host matches 'fwd'",
}

func TestParse_success(t *testing.T) {
	a := assert.New(t)
	for _, row := range inputs {
		a.CheckError(testParser(t, row))
	}
}

func TestCompile(t *testing.T) {
	a := assert.New(t)
	for _, row := range inputs {
		p := Parser{Buffer: row}
		p.Init()
		a.CheckError(p.Parse())
		p.Execute()
		testParserResult(t, p)
	}
}

func TestPredicate_parse(t *testing.T) {
	a := assert.New(t)
	p := Parser{Buffer: "describe x where key in ('a', 'b', 'c')"}
	p.Init()
	a.CheckError(p.Parse())
	p.Execute()
	testParserResult(t, p)
}

func testParser(t *testing.T, input string) error {
	p := Parser{Buffer: input}
	p.Init()
	return p.Parse()
}

func testParserResult(t *testing.T, p Parser) {
	a := assert.New(t)
	a.EqInt(len(p.nodeStack), 0)
	a.EqInt(len(p.errors), 0)
}

func getNode(input string) {
	p := Parser{Buffer: input}
	p.Init()
	p.Parse()
	p.Execute()
}
