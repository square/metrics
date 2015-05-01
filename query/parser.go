// Copyright 2015 Square Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package query contains all the logic to parse
// and execute queries against the underlying metric system.
package query

import (
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/square/metrics/api"
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
//
// returns either:
// * command
// * SyntaxError (user error - query is invalid)
// * AssertionError (programming error, and is a sign of a bug).
func Parse(query string) (Command, error) {
	p := Parser{Buffer: query}
	p.Init()
	if err := p.Parse(); err != nil {
		// Parsing error - invalid syntax.
		// TODO - return the token where the error is occurring.
		return nil, SyntaxErrors([]SyntaxError{{
			token:   "",
			message: err.Error(),
		}})
	}
	p.Execute()
	if len(p.assertions) > 0 {
		// logic error - an internal constraint is violated.
		// TODO - log this error internally.
		return nil, AssertionError{"Programming error"}
	}
	if len(p.nodeStack) > 0 {
		return nil, AssertionError{"Node stack is not empty"}
	}
	if len(p.errors) > 0 {
		// user error - an invalid query is provided.
		return nil, SyntaxErrors(p.errors)
	}
	if p.command == nil {
		// after parsing has finished, there should be a command available.
		return nil, AssertionError{"No command"}
	}
	return p.command, nil
}

// Error functions
// ===============
// these functions are called to mark that an error has occurred
// while parsing or constructing command.

func (p *Parser) flagSyntaxError(err SyntaxError) {
	p.errors = append(p.errors, err)
}

func (p *Parser) flagAssert(err error) {
	p.assertions = append(p.assertions, err)
}

func (p *Parser) flagTypeAssertion(typeString string) {
	p.flagAssert(fmt.Errorf("[%s] expected %s", functionName(1), typeString))
}

// Generic Stack Operation
// =======================
func (p *Parser) popNode() Node {
	l := len(p.nodeStack)
	if l == 0 {
		p.flagAssert(errors.New("popNode() on an empty stack"))
		return nil
	}
	node := p.nodeStack[l-1]
	p.nodeStack = p.nodeStack[:l-1]
	return node
}

func (p *Parser) peekNode() Node {
	l := len(p.nodeStack)
	if l == 0 {
		p.flagAssert(errors.New("peekNode() on an empty stack"))
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
		p.flagTypeAssertion("Predicate")
		return
	}
	literalNode, ok := p.popNode().(*literalNode)
	if !ok {
		p.flagTypeAssertion("literalNode")
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
		p.flagTypeAssertion("literalNode")
		return
	}
	tagNode, ok := p.popNode().(*tagNode)
	if !ok {
		p.flagTypeAssertion("tagNode")
		return
	}
	p.pushNode(&listMatcher{
		tag:     tagNode.tag,
		matches: []string{literalNode.literal},
	})
}

func (p *Parser) addListMatcher() {
	literalNode, ok := p.popNode().(*literalListNode)
	if !ok {
		p.flagTypeAssertion("literalNode")
		return
	}
	tagNode, ok := p.popNode().(*tagNode)
	if !ok {
		p.flagTypeAssertion("tagNode")
		return
	}
	p.pushNode(&listMatcher{
		tag:     tagNode.tag,
		matches: literalNode.literals,
	})
}

func (p *Parser) addRegexMatcher() {
	literalNode, ok := p.popNode().(*literalNode)
	if !ok {
		p.flagTypeAssertion("literalNode")
		return
	}
	tagNode, ok := p.popNode().(*tagNode)
	if !ok {
		p.flagTypeAssertion("tagNode")
		return
	}
	compiled, err := regexp.Compile(literalNode.literal)
	if err != nil {
		// TODO - return more user-friendly error.
		p.flagSyntaxError(SyntaxError{
			token:   literalNode.literal,
			message: fmt.Sprintf("Cannot parse the regex: %s", err.Error()),
		})
	}
	p.pushNode(&regexMatcher{
		tag:   tagNode.tag,
		regex: compiled,
	})
}

func (p *Parser) addTag(tag string) {
	p.pushNode(&tagNode{tag: tag})
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
		p.flagTypeAssertion("literalNode")
	}
}

func (p *Parser) addNotPredicate() {
	predicate, ok := p.popNode().(Predicate)
	if ok {
		p.pushNode(&notPredicate{predicate})
	} else {
		p.flagTypeAssertion("Predicate")
	}
}
func (p *Parser) addOrPredicate() {
	rightPredicate, ok := p.popNode().(Predicate)
	if !ok {
		p.flagTypeAssertion("Predicate")
	}
	leftPredicate, ok := p.popNode().(Predicate)
	if !ok {
		p.flagTypeAssertion("Predicate")
	}
	p.pushNode(&orPredicate{
		predicates: []Predicate{
			leftPredicate,
			rightPredicate,
		},
	})
}

func (p *Parser) addNullPredicate() {
	p.pushNode(&andPredicate{predicates: []Predicate{}})
}

func (p *Parser) addAndPredicate() {
	rightPredicate, ok := p.popNode().(Predicate)
	if !ok {
		p.flagTypeAssertion("Predicate")
	}
	leftPredicate, ok := p.popNode().(Predicate)
	if !ok {
		p.flagTypeAssertion("Predicate")
	}
	p.pushNode(&andPredicate{
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

var functionNameRegex = regexp.MustCompile(`[^./]+$`)

// name of the function on the stack.
// depth(0) - name of the function calling functionName(0)
// each additional depth traverses the stack frame further towards the caller.
func functionName(depth int) string {
	pc := make([]uintptr, 1)
	runtime.Callers(depth+2, pc)
	f := runtime.FuncForPC(pc[0])
	return functionNameRegex.FindString(f.Name())
}
