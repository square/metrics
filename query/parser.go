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
	"github.com/square/metrics/function"
	"github.com/square/metrics/query/predicate"
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

// parseDate converts the given datestring (from one of the allowable formats) into a millisecond offset from the Unix epoch.
func parseDate(date string, now time.Time) (int64, error) {

	if date == "now" {
		return now.Unix() * 1000, nil
	}

	// Millisecond epoch timestamp.
	if epoch, err := strconv.ParseInt(date, 10, 0); err == nil {
		return epoch, nil
	}

	relativeTime, err := function.StringToDuration(date)
	if err == nil {
		// A relative date.
		return now.Add(relativeTime).Unix() * 1000, nil
	}

	errorMessage := fmt.Sprintf("Expected formatted date or relative time but got '%s'", date)
	for _, format := range dateFormats {
		t, err := time.Parse(format, date)
		if err == nil {
			return t.Unix()*1000 + int64(t.Nanosecond()/1000000), nil
		}
	}
	return -1, errors.New(errorMessage)
}

type ParserAssert struct {
	error
}

func Parse(query string) (commandResult Command, finalErr error) {
	p := Parser{Buffer: query}
	p.Init()
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if message, ok := r.(ParserAssert); ok {
			finalErr = message
		}
	}()
	if err := p.Parse(); err != nil {
		// Parsing error - invalid syntax.
		// TODO - return the token where the error is occurring.
		if _, ok := err.(*parseError); ok {
			return nil, SyntaxErrors([]SyntaxError{{
				token:   "",
				message: customParseError(&p),
			}})
		} else {
			// generic error (should not occur).
			return nil, AssertionError{"Non-parse error raised"}
		}
	}
	p.Execute()
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

// Generic Stack Operation
// =======================
func (p *Parser) popNodeInto(target interface{}) {
	targetValue := reflect.ValueOf(target)
	if targetValue.Type().Kind() != reflect.Ptr {
		panic(ParserAssert{ // Will unwind until it comes to "p.Parse()" which has a recover.
			fmt.Errorf("[%s] popNodeInto() given a non-pointer target", functionName(1)),
		})
	}
	l := len(p.nodeStack)
	if l == 0 {
		panic(ParserAssert{fmt.Errorf("[%s] popNodeInto() on an empty stack", functionName(1))})
	}
	node := p.nodeStack[l-1]
	p.nodeStack = p.nodeStack[:l-1]

	nodeValue := reflect.ValueOf(node)
	actualType := nodeValue.Type()
	expectedType := targetValue.Elem().Type()
	if !actualType.ConvertibleTo(expectedType) {
		panic(ParserAssert{fmt.Errorf("[%s] popNodeInto() - expected %s, got off the stack %s",
			functionName(1),
			expectedType.String(),
			actualType.String()),
		})
	}
	targetValue.Elem().Set(nodeValue)
}

func (p *Parser) peekNodeInto(target interface{}) {
	targetValue := reflect.ValueOf(target)
	if targetValue.Type().Kind() != reflect.Ptr {
		panic(ParserAssert{
			fmt.Errorf("[%s] peekNodeInto() given a non-pointer target", functionName(1)),
		})
	}
	l := len(p.nodeStack)
	if l == 0 {
		panic(ParserAssert{fmt.Errorf("[%s] peekNodeInto() on an empty stack", functionName(1))})
	}

	nodeValue := reflect.ValueOf(p.nodeStack[l-1])

	expectedType := targetValue.Elem().Type()
	actualType := nodeValue.Type()
	if !actualType.ConvertibleTo(expectedType) {
		panic(ParserAssert{fmt.Errorf("[%s] peekNodeInto() - expected %s, got off the stack %s",
			functionName(1),
			expectedType.String(),
			actualType.String()),
		})
	}
	targetValue.Elem().Set(nodeValue)
}

func (p *Parser) pushNode(node Node) {
	p.nodeStack = append(p.nodeStack, node)
}

// Modification Operations
// =======================
// These operations are used by the embedded code snippets in language.peg
func (p *Parser) makeDescribe() {
	var node *predicateNode
	p.popNodeInto(&node)
	literal := p.popStringLiteral()
	p.command = &DescribeCommand{
		metricName: api.MetricKey(literal),
		predicate:  node.Predicate,
	}
}

func (p *Parser) makeSelect() {
	var contextNode *evaluationContextNode
	p.popNodeInto(&contextNode)
	var predicate *predicateNode
	p.popNodeInto(&predicate)
	var list *expressionList
	p.popNodeInto(&list)
	p.command = &SelectCommand{
		predicate:   predicate,
		expressions: list.expressions,
		context:     contextNode,
	}
}

func (p *Parser) addNullMatchClause() {
	p.pushNode(&matcherClause{regex: regexp.MustCompile("")})
}

func (p *Parser) addMatchClause() {
	compiled := p.popRegex()
	p.pushNode(&matcherClause{regex: compiled})
}

func (p *Parser) makeDescribeAll() {
	var matcher *matcherClause
	p.popNodeInto(&matcher)
	p.command = &DescribeAllCommand{matcher: matcher}
}

func (p *Parser) makeDescribeMetrics() {
	// Pop off the value.
	literal := p.popStringLiteral()
	// Pop of the tag name.
	var tagLiteral *tagLiteral
	p.popNodeInto(&tagLiteral)
	p.command = &DescribeMetricsCommand{tagKey: tagLiteral.tag, tagValue: literal}
}

func (p *Parser) addOperatorLiteral(operator string) {
	p.pushNode(&operatorLiteral{operator})
}

func (p *Parser) addOperatorFunction() {
	var right function.Expression

	p.popNodeInto(&right)
	var operatorNode *operatorLiteral
	p.popNodeInto(&operatorNode)
	var left function.Expression
	p.popNodeInto(&left)
	p.pushNode(&functionExpression{
		functionName: operatorNode.operator,
		arguments:    []function.Expression{left, right},
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
	var valueNode *evaluationContextValue
	p.popNodeInto(&valueNode)
	var keyNode *evaluationContextKey
	p.popNodeInto(&keyNode)
	var contextNode *evaluationContextNode
	p.popNodeInto(&contextNode)

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
	case "from", "to":
		var unix int64
		var err error
		now := time.Now()
		if unix, err = parseDate(value, now); err != nil {
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
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			contextNode.Resolution = intValue
		} else if duration, err := function.StringToDuration(value); err == nil {
			contextNode.Resolution = int64(duration / time.Millisecond)
		} else {
			p.flagSyntaxError(SyntaxError{
				token:   value,
				message: fmt.Sprintf("Expected number but parse failed; %s", err.Error()),
			})
		}
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
	var contextNode *evaluationContextNode
	p.popNodeInto(&contextNode)
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

func (p *Parser) addPipeExpression() {
	var groupBy *groupByList
	p.popNodeInto(&groupBy)
	var expressionList *expressionList
	p.popNodeInto(&expressionList)
	literal := p.popStringLiteral()
	var expressionNode function.Expression
	p.popNodeInto(&expressionNode)
	p.pushNode(&functionExpression{
		functionName:     literal,
		arguments:        append([]function.Expression{expressionNode}, expressionList.expressions...),
		groupBy:          groupBy.list,
		groupByCollapses: groupBy.collapses,
	})
}

func (p *Parser) addFunctionInvocation() {
	var groupBy *groupByList
	p.popNodeInto(&groupBy)
	var expressionList *expressionList
	p.popNodeInto(&expressionList)
	literal := p.popStringLiteral()
	// user-level error generation here.
	p.pushNode(&functionExpression{
		functionName:     literal,
		arguments:        expressionList.expressions,
		groupBy:          groupBy.list,
		groupByCollapses: groupBy.collapses,
	})
}

func (p *Parser) addAnnotationExpression(annotation string) {
	var content function.Expression
	p.popNodeInto(&content)
	p.pushNode(&annotationExpression{
		content:    content,
		annotation: annotation,
	})
}

func (p *Parser) addMetricExpression() {
	var predicateNode *predicateNode
	p.popNodeInto(&predicateNode)
	literal := p.popStringLiteral()
	p.pushNode(&metricFetchExpression{
		metricName: literal,
		predicate:  predicateNode,
	})
}

func (p *Parser) addExpressionList() {
	p.pushNode(&expressionList{
		make([]function.Expression, 0),
	})
}

func (p *Parser) appendExpression() {
	var expressionNode function.Expression
	p.popNodeInto(&expressionNode)
	var listNode *expressionList
	p.peekNodeInto(&listNode)
	listNode.expressions = append(listNode.expressions, expressionNode)
}

func (p *Parser) addLiteralMatcher() {
	literal := p.popStringLiteral()
	var tagLiteral *tagLiteral

	p.popNodeInto(&tagLiteral)
	p.pushNode(&predicateNode{
		predicate.ListMatcher{
			Tag:    tagLiteral.tag,
			Values: []string{literal},
		},
	})
}

func (p *Parser) addListMatcher() {
	var stringLiteral *stringLiteralList

	p.popNodeInto(&stringLiteral)
	var tagLiteral *tagLiteral
	p.popNodeInto(&tagLiteral)
	p.pushNode(&predicateNode{
		predicate.ListMatcher{
			Tag:    tagLiteral.tag,
			Values: stringLiteral.literals,
		},
	})
}

func (p *Parser) addRegexMatcher() {
	compiled := p.popRegex()
	var tagLiteral *tagLiteral
	p.popNodeInto(&tagLiteral)
	p.pushNode(&predicateNode{
		predicate.RegexMatcher{
			Tag:   tagLiteral.tag,
			Regex: compiled,
		},
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
	var listNode *stringLiteralList
	p.peekNodeInto(&listNode)
	listNode.literals = append(listNode.literals, literal)
}

func (p *Parser) addGroupBy() {
	p.pushNode(&groupByList{make([]string, 0), false})
}

func (p *Parser) appendGroupBy(literal string) {
	var listNode *groupByList
	p.peekNodeInto(&listNode)
	listNode.list = append(listNode.list, literal)
}

func (p *Parser) appendCollapseBy(literal string) {
	var listNode *groupByList
	p.peekNodeInto(&listNode)
	listNode.collapses = true // Switch to collapsing mode
	listNode.list = append(listNode.list, literal)
}

func (p *Parser) addNotPredicate() {
	var original *predicateNode
	p.popNodeInto(&original)
	p.pushNode(&predicateNode{
		predicate.NotPredicate{original},
	})
}

func (p *Parser) addOrPredicate() {
	var rightPredicate *predicateNode
	p.popNodeInto(&rightPredicate)
	var leftPredicate *predicateNode
	p.popNodeInto(&leftPredicate)
	p.pushNode(&predicateNode{
		predicate.Any(leftPredicate, rightPredicate),
	})
}

func (p *Parser) addNullPredicate() {
	p.pushNode(&predicateNode{
		predicate.All(),
	})
}

func (p *Parser) addAndPredicate() {
	var rightPredicate *predicateNode
	p.popNodeInto(&rightPredicate)
	var leftPredicate *predicateNode
	p.popNodeInto(&leftPredicate)
	p.pushNode(&predicateNode{
		predicate.All(leftPredicate, rightPredicate),
	})
}

func (p *Parser) addDurationNode(value string) {
	duration, err := function.StringToDuration(value)
	p.pushNode(&durationExpression{value, duration})
	if err != nil {
		p.flagSyntaxError(SyntaxError{
			token:   value,
			message: fmt.Sprintf("'%s' is not a valid duration: %s", value, err.Error()),
		})
	}
}

func (p *Parser) addNumberNode(value string) {
	parsedValue, err := strconv.ParseFloat(value, 64)
	p.pushNode(&scalarExpression{parsedValue})
	if err != nil || math.IsNaN(parsedValue) {
		p.flagSyntaxError(SyntaxError{
			token:   value,
			message: fmt.Sprintf("Cannot parse the number: %s", value),
		})
	}
}

func (p *Parser) addStringNode(value string) {
	p.pushNode(&stringExpression{value})
}

// Utility Stack Operations
func (p *Parser) popRegex() *regexp.Regexp {
	literal := p.popStringLiteral()
	compiled, err := regexp.Compile(literal)
	if err != nil {
		p.flagSyntaxError(SyntaxError{
			token:   literal,
			message: fmt.Sprintf("Cannot parse the regex: %s", err.Error()),
		})
		return nil
	}
	return compiled
}

func (p *Parser) popStringLiteral() string {
	var stringLiteral *stringLiteral
	p.popNodeInto(&stringLiteral)
	return stringLiteral.literal
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

// modified version of (*parseError).Error() so that it is not ANSI-colored.
func customParseError(parser *Parser) string {
	type pair struct {
		start int
		end   int
	}
	tokens, error := parser.tokenTree.Error(), ""
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(parser.buffer, positions)
	printedPositions := make(map[pair]bool)
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		if token.pegRule == ruleUnknown {
			continue // skip the unknown rule.
		} else if printedPositions[pair{begin, end}] {
			continue // already printed error for this position
		}
		printedPositions[pair{begin, end}] = true
		line, underline := makePrettyLine(parser, token, translations)
		error += fmt.Sprintf("parse error near [%v] (line %v symbol %v - line %v symbol %v):\n%s\n%s\n",
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			line,
			underline,
		)
	}

	return error
}

func makePrettyLine(parser *Parser, token token32, translations textPositionMap) (string, string) {
	N := len(parser.Buffer)
	begin, end := int(token.begin), int(token.end)
	lineStart := begin - translations[begin].symbol + 1
	lineEnd := lineStart
	for i := lineStart; i < N && parser.Buffer[i] != '\n'; i++ {
		lineEnd = i
	}
	if lineEnd < N && parser.Buffer[lineEnd] != '\n' {
		lineEnd = lineEnd + 1
	}
	lineStart = max(0, min(N-1, lineStart))
	lineEnd = max(0, min(N-1, lineEnd))
	line := parser.Buffer[lineStart:lineEnd]
	symbolBegin := translations[begin].symbol - 1
	if symbolBegin < 0 {
		symbolBegin = 0
	}
	if translations[begin].line == translations[end].line {
		// single-line error - print the entire line and draw carets under the token.
		length := translations[end].symbol - translations[begin].symbol
		if length <= 0 {
			length = 1
		}
		underline := strings.Repeat(" ", symbolBegin) + strings.Repeat("^", length)
		return line, underline
	} else {
		// multi-line error - print the firsst line and draw carets under the token until the line finishes.
		length := lineEnd - lineStart - translations[begin].symbol - 1
		if length <= 0 {
			length = 1
		}
		underline := strings.Repeat(" ", symbolBegin) + strings.Repeat("^", length)
		return line, underline
	}
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
