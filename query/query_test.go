package query

// show all metrics
// show tags WHERE predicate

import (
	"github.com/square/metrics-indexer/assert"
	"testing"
)

var inputs = []string{
	// describes
	"describe all",
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

var parseOnly = []string{
	// selects
	"select x",
	"select x-y-z",
	"select 0",
	"select x, y",
	"select 1 + 2 * 3 + 4",
	"select x * (y + 123), z",
	"select scalar.max(x)",
	"select aggregate.max(x, y)",
	"select aggregate.max(x group by foo) + 3",
	"select x where y = 'z'",
	"select x + z[y = 'z']",
	"select x[y = 'z']",
	"select aggregate.max(x[y = 'z'] group by foo)",
}

// TODO - add test for "does not parse"

func TestParse_success(t *testing.T) {
	for _, row := range inputs {
		if err := testParser(t, row); err != nil {
			t.Errorf("[%s] failed to parse: %s", row, err.Error())
		}
	}
	for _, row := range parseOnly {
		if err := testParser(t, row); err != nil {
			t.Errorf("[%s] failed to parse: %s", row, err.Error())
		}
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
