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
	"math"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
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

func (p *Parser) flagTypeAssertion() {
	p.flagAssert(fmt.Errorf("[%s] type assertion failure", functionName(1)))
}

// Generic Stack Operation
// =======================
func (p *Parser) popNode(expected reflect.Type) Node {
	l := len(p.nodeStack)
	if l == 0 {
		p.flagAssert(fmt.Errorf("[%s] popNode() on an empty stack", functionName(1)))
		return nil
	}
	node := p.nodeStack[l-1]
	p.nodeStack = p.nodeStack[:l-1]
	actualType := reflect.ValueOf(node).Type()
	if !actualType.ConvertibleTo(expected) {
		p.flagAssert(fmt.Errorf("[%s] popNode() - expected %s, got %s",
			functionName(1),
			expected.String(),
			reflect.ValueOf(node).Type().String()),
		)
		return nil
	}
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
	predicateNode, ok := p.popNode(predicateType).(Predicate)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	stringLiteral, ok := p.popNode(stringLiteralPointer).(*stringLiteral)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	p.command = &DescribeCommand{
		metricName: api.MetricKey(stringLiteral.literal),
		predicate:  predicateNode,
	}
}

func (p *Parser) makeSelect() {
	predicateNode, ok := p.popNode(predicateType).(Predicate)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	expressionList, ok := p.popNode(expressionListPointer).(*expressionList)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	p.command = &SelectCommand{
		predicate:   predicateNode,
		expressions: expressionList.expressions,
	}
}

func (p *Parser) makeDescribeAll() {
	p.command = &DescribeAllCommand{}
}

func (p *Parser) addOperatorLiteral(operator string) {
	p.pushNode(&operatorLiteral{operator})
}

func (p *Parser) addOperatorFunction() {
	right, ok := p.popNode(expressionType).(Expression)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	operatorNode, ok := p.popNode(operatorLiteralPointer).(*operatorLiteral)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	left, ok := p.popNode(expressionType).(Expression)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	p.pushNode(&functionExpression{
		functionName: operatorNode.operator,
		arguments:    []Expression{left, right},
	})
}

func (p *Parser) addFunctionInvocation() {
	expressionList, ok := p.popNode(expressionListPointer).(*expressionList)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	stringLiteral, ok := p.popNode(stringLiteralPointer).(*stringLiteral)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	// user-level error generation here.
	p.pushNode(&functionExpression{
		functionName: stringLiteral.literal,
		arguments:    expressionList.expressions,
	})
}

func (p *Parser) addMetricExpression() {
	predicateNode, ok := p.popNode(predicateType).(Predicate)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	stringLiteral, ok := p.popNode(stringLiteralPointer).(*stringLiteral)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	p.pushNode(&metricFetchExpression{
		metricName: stringLiteral.literal,
		predicate:  predicateNode,
	})
}

func (p *Parser) addExpressionList() {
	p.pushNode(&expressionList{
		make([]Expression, 0),
	})
}

func (p *Parser) appendExpression() {
	expressionNode, ok := p.popNode(expressionType).(Expression)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	listNode, ok := p.peekNode().(*expressionList)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	listNode.expressions = append(listNode.expressions, expressionNode)
}

func (p *Parser) addLiteralMatcher() {
	stringLiteral, ok := p.popNode(stringLiteralPointer).(*stringLiteral)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	tagLiteral, ok := p.popNode(tagLiteralPointer).(*tagLiteral)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	p.pushNode(&listMatcher{
		tag:     tagLiteral.tag,
		matches: []string{stringLiteral.literal},
	})
}

func (p *Parser) addListMatcher() {
	stringLiteral, ok := p.popNode(stringLiteralListPointer).(*stringLiteralList)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	tagLiteral, ok := p.popNode(tagLiteralPointer).(*tagLiteral)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	p.pushNode(&listMatcher{
		tag:     tagLiteral.tag,
		matches: stringLiteral.literals,
	})
}

func (p *Parser) addRegexMatcher() {
	stringLiteral, ok := p.popNode(stringLiteralPointer).(*stringLiteral)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	tagLiteral, ok := p.popNode(tagLiteralPointer).(*tagLiteral)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	compiled, err := regexp.Compile(stringLiteral.literal)
	if err != nil {
		p.flagSyntaxError(SyntaxError{
			token:   stringLiteral.literal,
			message: fmt.Sprintf("Cannot parse the regex: %s", err.Error()),
		})
	}
	p.pushNode(&regexMatcher{
		tag:   tagLiteral.tag,
		regex: compiled,
	})
}

func (p *Parser) addTagLiteral(tag string) {
	p.pushNode(&tagLiteral{tag: tag})
}

func (p *Parser) addLiteralListNode() {
	p.pushNode(&stringLiteralList{make([]string, 0)})
}

func (p *Parser) addStringLiteral(literal string) {
	p.pushNode(&stringLiteral{literal})
}

func (p *Parser) appendLiteral(literal string) {
	listNode, ok := p.peekNode().(*stringLiteralList)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	listNode.literals = append(listNode.literals, literal)
}

func (p *Parser) addNotPredicate() {
	predicate, ok := p.popNode(predicateType).(Predicate)
	if ok {
		p.pushNode(&notPredicate{predicate})
	} else {
		p.flagTypeAssertion()
		return
	}
}

func (p *Parser) addOrPredicate() {
	rightPredicate, ok := p.popNode(predicateType).(Predicate)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	leftPredicate, ok := p.popNode(predicateType).(Predicate)
	if !ok {
		p.flagTypeAssertion()
		return
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
	rightPredicate, ok := p.popNode(predicateType).(Predicate)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	leftPredicate, ok := p.popNode(predicateType).(Predicate)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	p.pushNode(&andPredicate{
		predicates: []Predicate{
			leftPredicate,
			rightPredicate,
		},
	})
}

func (p *Parser) addNumberNode(value string) {
	parsedValue, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(parsedValue) {
		p.flagSyntaxError(SyntaxError{
			token:   value,
			message: fmt.Sprintf("Cannot parse the number: %s", value),
		})
		return
	}
	p.pushNode(&numberExpression{parsedValue})
}

// Utility Functions
// =================

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

// utility type variables
var (
	predicateType            = reflect.TypeOf((*Predicate)(nil)).Elem()
	expressionType           = reflect.TypeOf((*Expression)(nil)).Elem()
	stringLiteralListPointer = reflect.TypeOf((*stringLiteralList)(nil))
	stringLiteralPointer     = reflect.TypeOf((*stringLiteral)(nil))
	operatorLiteralPointer   = reflect.TypeOf((*operatorLiteral)(nil))
	expressionListPointer    = reflect.TypeOf((*expressionList)(nil))
	tagLiteralPointer        = reflect.TypeOf((*tagLiteral)(nil))
)
