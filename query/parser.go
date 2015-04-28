// Package query contains all the logic to parse
// and execute queries against the underlying metric system.
package query

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/square/metrics-indexer/api"
)

// Parse is the entry point of the parser.
// It does the following:
// * Parses the given query string.
// * Checks for any syntax error (detected by peg during Parse())
// or logic error (detected while traversing the parse tree in Execute()).
// * Returns the final Command resulting from the parsing.
//
// The parsing is done in the following ways.
// 1. Parse() method constructs the abstract syntax tree.
// 2. Execute() method visits each node in-order, executing the
// snippet of code embedded in the grammar definition.
//
// Details on Execute():
//   Execute traverses the AST to generate more refined output
// from the AST. Our final output is Command, representing the
// procedural operation of the query.
// To generate this, we maintain a stack of processed AST nodes
// in our parser object. (AST nodes are abstracted away by PEG.
// Processed nodes are represented by the go interface Node).
//
// * The stack starts empty.
// * Nodes are repeatedly pushed and popped during the traversal.
// * At the end of the run, stack should be empty and a single
// Command object is produced. (Technically Command could've been
// pushed to the stack also)
// * Each AST node can make intelligent assumptions on the current
// state of the node stack. This information is a bit implicit but
// enforced via type assertions throughout the code.
func Parse(query string) (Command, error) {
	p := Parser{Buffer: query}
	p.Init()
	if err := p.Parse(); err != nil {
		// Parsing error - invalid syntax.
		return nil, err
	}
	p.Execute()
	if len(p.errors) > 0 {
		// Logic error - doing some invalid operation.
		return nil, p.errors[0]
	}
	if len(p.nodeStack) > 0 {
		return nil, errors.New("Node stack is not empty")
	}
	if p.command == nil {
		// after parsing has finished, there should be a command available.
		return nil, errors.New("Invalid Command")
	}
	return p.command, nil
}

// Error functions
// ===============
// these functions

func (p *Parser) flagError(err error) {
	p.errors = append(p.errors, err)
}

func (p *Parser) flagTypeError(typeString string) {
	p.flagError(fmt.Errorf("expected %s", typeString))
}

// Generic Stack Operation
// =======================
func (p *Parser) popNode() Node {
	l := len(p.nodeStack)
	if l == 0 {
		p.flagError(errors.New("popNode() on an empty stack"))
		return nil
	}
	node := p.nodeStack[l-1]
	p.nodeStack = p.nodeStack[:l-1]
	return node
}

func (p *Parser) peekNode() Node {
	l := len(p.nodeStack)
	if l == 0 {
		p.flagError(errors.New("peekNode() on an empty stack"))
		return nil
	}
	node := p.nodeStack[l-1]
	return node
}

func (p *Parser) pushNode(node Node) {
	p.nodeStack = append(p.nodeStack, node)
}

// Modification Operations
// =======================
// These operations are used by the embedded code snippets in language.peg
func (p *Parser) makeDescribe() {
	predicateNode, ok := p.popNode().(Predicate)
	if !ok {
		p.flagTypeError("Predicate")
		return
	}
	literalNode, ok := p.popNode().(*literalNode)
	if !ok {
		p.flagTypeError("literalNode")
		return
	}
	p.command = &DescribeCommand{
		metricName: api.MetricKey(literalNode.literal),
		predicate:  predicateNode,
	}
}

func (p *Parser) makeDescribeAll() {
	p.command = &DescribeAllCommand{}
}

func (p *Parser) addLiteralMatcher() {
	literalNode, ok := p.popNode().(*literalNode)
	if !ok {
		p.flagTypeError("literalNode")
		return
	}
	tagRefNode, ok := p.popNode().(*tagRefNode)
	if !ok {
		p.flagTypeError("tagRefNode")
		return
	}
	p.pushNode(&listMatcher{
		alias:   tagRefNode.alias,
		tag:     tagRefNode.tag,
		matches: []string{literalNode.literal},
	})
}

func (p *Parser) addListMatcher() {
	literalNode, ok := p.popNode().(*literalListNode)
	if !ok {
		p.flagTypeError("literalNode")
		return
	}
	tagRefNode, ok := p.popNode().(*tagRefNode)
	if !ok {
		p.flagTypeError("tagRefNode")
		return
	}
	p.pushNode(&listMatcher{
		alias:   tagRefNode.alias,
		tag:     tagRefNode.tag,
		matches: literalNode.literals,
	})
}

func (p *Parser) addRegexMatcher() {
	literalNode, ok := p.popNode().(*literalNode)
	if !ok {
		p.flagTypeError("literalNode")
		return
	}
	tagRefNode, ok := p.popNode().(*tagRefNode)
	if !ok {
		p.flagTypeError("tagRefNode")
		return
	}
	compiled, err := regexp.Compile(literalNode.literal)
	if err != nil {
		// TODO - return more user-friendly error.
		p.flagError(errors.New("Cannot parse regex."))
	}
	p.pushNode(&regexMatcher{
		alias: tagRefNode.alias,
		tag:   tagRefNode.tag,
		regex: compiled,
	})
}

func (p *Parser) addTagRefNode() {
	p.pushNode(&tagRefNode{})
}

func (p *Parser) setTag(tag string) {
	tagRefNode, ok := p.peekNode().(*tagRefNode)
	if ok {
		tagRefNode.tag = tag
	} else {
		p.flagTypeError("tagRefNode")
	}
}

func (p *Parser) setAlias(alias string) {
	tagRefNode, ok := p.peekNode().(*tagRefNode)
	if ok {
		tagRefNode.alias = alias
	} else {
		p.flagTypeError("tagRefNode")
	}
}

func (p *Parser) addLiteralListNode() {
	p.pushNode(&literalListNode{make([]string, 0)})
}

func (p *Parser) addLiteralNode(literal string) {
	p.pushNode(&literalNode{literal})
}

func (p *Parser) appendLiteral(literal string) {
	literalNode, ok := p.peekNode().(*literalListNode)
	if ok {
		literalNode.literals = append(literalNode.literals, literal)
	} else {
		p.flagTypeError("literalNode")
	}
}

func (p *Parser) addNotMatcher() {
	predicate, ok := p.popNode().(Predicate)
	if ok {
		p.pushNode(&notPred{predicate})
	} else {
		p.flagTypeError("Predicate")
	}
}
func (p *Parser) addOrMatcher() {
	rightPredicate, ok := p.popNode().(Predicate)
	if !ok {
		p.flagTypeError("Predicate")
	}
	leftPredicate, ok := p.popNode().(Predicate)
	if !ok {
		p.flagTypeError("Predicate")
	}
	p.pushNode(&orPred{
		predicates: []Predicate{
			leftPredicate,
			rightPredicate,
		},
	})
}

func (p *Parser) addNullPredicate() {
	p.pushNode(&andPred{predicates: []Predicate{}})
}

func (p *Parser) addAndMatcher() {
	rightPredicate, ok := p.popNode().(Predicate)
	if !ok {
		p.flagTypeError("Predicate")
	}
	leftPredicate, ok := p.popNode().(Predicate)
	if !ok {
		p.flagTypeError("Predicate")
	}
	p.pushNode(&andPred{
		predicates: []Predicate{
			leftPredicate,
			rightPredicate,
		},
	})
}

// used to unescape:
// - identifiers (no unescaping required).
// - quoted strings.
func unescapeLiteral(escaped string) string {
	if len(escaped) <= 1 {
		return escaped
	}
	escapedCharacters := []string{
		"'", "`", "\"", "\\",
	}
	processed := escaped
	first := processed[0]
	if first == '\'' || first == '"' || first == '`' {
		processed = processed[1 : len(processed)-1]
		for _, char := range escapedCharacters {
			processed = strings.Replace(processed, `\`+char, char, -1)
		}
	}
	return processed
}
