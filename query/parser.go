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
	"time"

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

// The dateFormats are tried in sequence until one of them succeeds.
// This is the best way that I can see to allow multiple formats (which is reasonable for human input).
// Keep in mind that the format is YEAR-MONTH-DAY.
// Time zones are ommitted where only the date is given (perhaps this should be changed)
var dateFormats = []string{
	"2006-1-2 15:04:05 MST",
	"2006-1-2 15:04 MST",
	"2006-1-2 15 MST",
	"2006-1-2 MST",
	"2006-1-2",
	"2006-1",
	"2006/1/2 15:04:05 MST",
	"2006/1/2 15:04 MST",
	"2006/1/2 15 MST",
	"2006/1/2 MST",
	"2006/1/2",
	"2006/1",
	"Jan 2 2006 15:04:05 MST",
	"Jan 2 2006 15:04 MST",
	"Jan 2 2006 15 MST",
	"Jan 2 2006 MST",
	"Jan 2 2006",
	"Jan 2006",
	"2 Jan 2006 15:04:05 MST",
	"2 Jan 2006 15:04 MST",
	"2 Jan 2006 15 MST",
	"2 Jan 2006 MST",
	"2 Jan 2006",
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
}

var relativeTimeRegexp = regexp.MustCompile(`^now\s*([+-])\s*([0-9]+)([smhdwMy])$`)

// parseDate converts the given datestring (from one of the allowable formats) into a millisecond offset from the Unix epoch.
func parseDate(date string) (int64, error) {

	if date == "now" {
		return time.Now().Unix() * 1000, nil
	}

	matches := relativeTimeRegexp.FindStringSubmatch(date)
	if matches != nil {
		// This match can be used to determine the duration of time.
		now := time.Now().Unix()
		sign := int64(1)
		if matches[1] == "-" {
			sign = -1
		}
		offset, err := strconv.ParseInt(matches[2], 10, 0)
		if err != nil {
			// This case could occur when the number is too large, for example.
			return 0, err
		}
		switch matches[3] {
		case "s":
			offset *= 1
		case "m":
			offset *= 60
		case "h":
			offset *= 60 * 60
		case "d":
			offset *= 60 * 60 * 24
		case "w":
			offset *= 60 * 60 * 24 * 7
		case "M":
			offset *= 60 * 60 * 24 * 30
		case "y":
			offset *= 60 * 60 * 24 * 365
		default:
			// Won't happen, regex can't produce any other value.
			panic(matches[3])
		}
		return (now + sign*offset) * 1000, nil
	}

	intValue, err := strconv.ParseInt(date, 10, 64)
	if err == nil {
		return intValue, nil
	}
	errorMessage := "Expected formatted date or relative time (string of the form 'now - 5h' for example)"
	for _, format := range dateFormats {
		t, err := time.Parse(format, date)
		if err == nil {
			return t.Unix()*1000 + int64(t.Nanosecond()/1000000), nil
		}
		errorMessage += "\nfailed: " + err.Error()
	}
	return 0, errors.New(errorMessage)
}

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
	predicateNode, ok := p.popNode(predicateType).(api.Predicate)
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
	contextNode, ok := p.popNode(evaluationContextNodePointer).(*evaluationContextNode)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	predicateNode, ok := p.popNode(predicateType).(api.Predicate)
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
		context:     contextNode,
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

func (p *Parser) addPropertyKey(key string) {
	p.pushNode(&evaluationContextKey{key})
}

func (p *Parser) addPropertyValue(value string) {
	p.pushNode(&evaluationContextValue{value})
}

func (p *Parser) addEvaluationContext() {
	p.pushNode(&evaluationContextNode{
		0, 0, 30000,
		api.SampleMean,
		make(map[string]bool),
	})
}

func (p *Parser) insertPropertyKeyValue() {
	valueNode, ok := p.popNode(evaluationContextValuePointer).(*evaluationContextValue)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	keyNode, ok := p.popNode(evaluationContextKeyPointer).(*evaluationContextKey)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	contextNode, ok := p.popNode(evaluationContextNodePointer).(*evaluationContextNode)
	if !ok {
		p.flagTypeAssertion()
		return
	}

	key := keyNode.key
	value := valueNode.value
	// Authenticate the validity of the given key and value...
	// The key must be one of "sample"(by), "from", "to", "resolution"

	// First check that the key has been assigned only once:
	if contextNode.assigned[key] {
		p.flagSyntaxError(SyntaxError{
			token:   key,
			message: fmt.Sprintf("Key %s has already been assigned", key),
		})
	}
	contextNode.assigned[key] = true

	switch key {
	case "sample":
		// If the key is "sample", it means we're in a "sample by" declaration.
		// Only three possible sample methods are defined: min, max, or mean.
		switch value {
		case "max":
			contextNode.SampleMethod = api.SampleMax
		case "min":
			contextNode.SampleMethod = api.SampleMin
		case "mean":
			contextNode.SampleMethod = api.SampleMean
		default:
			p.flagSyntaxError(SyntaxError{
				token:   value,
				message: fmt.Sprintf("Expected sampling method 'max', 'min', or 'mean' but got %s", value),
			})
		}
	case "from":
		fallthrough
	case "to":
		var unix int64
		var err error
		if unix, err = parseDate(value); err == nil {
			// Valid, so do nothing.
		} else {
			p.flagSyntaxError(SyntaxError{
				token:   value,
				message: err.Error(),
			})
		}
		if key == "from" {
			contextNode.Start = unix
		} else {
			contextNode.End = unix
		}
	case "resolution":
		// The value must be determined to be an int if the key is "resolution".
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			contextNode.Resolution = intValue
			break
		}
		duration, err := toDuration(stringValue(value))
		if err == nil {
			contextNode.Resolution = duration
			break
		}
		p.flagSyntaxError(SyntaxError{
			token:   value,
			message: fmt.Sprintf("Expected number but parse failed; %s", err.Error()),
		})
	default:
		p.flagSyntaxError(SyntaxError{
			token:   key,
			message: fmt.Sprintf("Unknown property key %s", key),
		})
	}
	p.pushNode(contextNode)
}

// makePropertyClause verifies that all mandatory fields have been assigned in the evaluation context.
func (p *Parser) checkPropertyClause() {
	contextNode, ok := p.popNode(evaluationContextNodePointer).(*evaluationContextNode)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	mandatoryFields := []string{"from", "to"} // Sample, resolution is optional (default to mean, 30s)
	for _, field := range mandatoryFields {
		if !contextNode.assigned[field] {
			p.flagSyntaxError(SyntaxError{
				token:   field,
				message: fmt.Sprintf("Field %s is never assigned in property clause", field),
			})
		}
	}
	p.pushNode(contextNode)
}

func (p *Parser) addFunctionInvocation() {
	groupBy, ok := p.popNode(groupByListPointer).(*groupByList)
	if !ok {
		p.flagTypeAssertion()
		return
	}
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
		groupBy:      groupBy.list,
	})
}

func (p *Parser) addMetricExpression() {
	predicateNode, ok := p.popNode(predicateType).(api.Predicate)
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
		tag:    tagLiteral.tag,
		values: []string{stringLiteral.literal},
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
		tag:    tagLiteral.tag,
		values: stringLiteral.literals,
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

func (p *Parser) addStringLiteral(literal string) {
	p.pushNode(&stringLiteral{literal})
}

func (p *Parser) addLiteralList() {
	p.pushNode(&stringLiteralList{make([]string, 0)})
}

func (p *Parser) appendLiteral(literal string) {
	listNode, ok := p.peekNode().(*stringLiteralList)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	listNode.literals = append(listNode.literals, literal)
}

func (p *Parser) addGroupBy() {
	p.pushNode(&groupByList{make([]string, 0)})
}

func (p *Parser) appendGroupBy(literal string) {
	listNode, ok := p.peekNode().(*groupByList)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	listNode.list = append(listNode.list, literal)
}

func (p *Parser) addNotPredicate() {
	predicate, ok := p.popNode(predicateType).(api.Predicate)
	if ok {
		p.pushNode(&notPredicate{predicate})
	} else {
		p.flagTypeAssertion()
		return
	}
}

func (p *Parser) addOrPredicate() {
	rightPredicate, ok := p.popNode(predicateType).(api.Predicate)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	leftPredicate, ok := p.popNode(predicateType).(api.Predicate)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	p.pushNode(&orPredicate{
		predicates: []api.Predicate{
			leftPredicate,
			rightPredicate,
		},
	})
}

func (p *Parser) addNullPredicate() {
	p.pushNode(&andPredicate{predicates: []api.Predicate{}})
}

func (p *Parser) addAndPredicate() {
	rightPredicate, ok := p.popNode(predicateType).(api.Predicate)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	leftPredicate, ok := p.popNode(predicateType).(api.Predicate)
	if !ok {
		p.flagTypeAssertion()
		return
	}
	p.pushNode(&andPredicate{
		predicates: []api.Predicate{
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
	p.pushNode(&scalarExpression{parsedValue})
}

func (p *Parser) addStringNode(value string) {
	p.pushNode(&stringExpression{value})
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
	predicateType                 = reflect.TypeOf((*api.Predicate)(nil)).Elem()
	expressionType                = reflect.TypeOf((*Expression)(nil)).Elem()
	expressionListPointer         = reflect.TypeOf((*expressionList)(nil))
	groupByListPointer            = reflect.TypeOf((*groupByList)(nil))
	operatorLiteralPointer        = reflect.TypeOf((*operatorLiteral)(nil))
	stringLiteralListPointer      = reflect.TypeOf((*stringLiteralList)(nil))
	stringLiteralPointer          = reflect.TypeOf((*stringLiteral)(nil))
	tagLiteralPointer             = reflect.TypeOf((*tagLiteral)(nil))
	evaluationContextValuePointer = reflect.TypeOf((*evaluationContextValue)(nil))
	evaluationContextKeyPointer   = reflect.TypeOf((*evaluationContextKey)(nil))
	evaluationContextNodePointer  = reflect.TypeOf((*evaluationContextNode)(nil))
)
