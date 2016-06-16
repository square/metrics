// Copyright 2015 - 2016 Square Inc.
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

// Package parser contains all the logic to parse
package parser

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
	"github.com/square/metrics/query/command"
	"github.com/square/metrics/query/expression"
	"github.com/square/metrics/query/predicate"
	"github.com/square/metrics/timeseries"
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
// Time zones are omitted where only the date is given (perhaps this should be changed)
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
	// @@ leaking param: date

	if date == "now" {
		return now.Unix() * 1000, nil
	}
	// @@ inlining call to time.Time.Unix

	// Millisecond epoch timestamp.
	if epoch, err := strconv.ParseInt(date, 10, 0); err == nil {
		return epoch, nil
	}

	relativeTime, err := function.StringToDuration(date)
	if err == nil {
		// A relative date.
		return now.Add(relativeTime).Unix() * 1000, nil
	}
	// @@ inlining call to time.Time.Add
	// @@ inlining call to time.Time.Unix

	errorMessage := fmt.Sprintf("Expected formatted date or relative time but got '%s'", date)
	for _, format := range dateFormats {
		// @@ date escapes to heap
		t, err := time.Parse(format, date)
		if err == nil {
			return t.Unix()*1000 + int64(t.Nanosecond()/1000000), nil
		}
		// @@ inlining call to time.Time.Unix
		// @@ inlining call to time.Time.Nanosecond
	}
	return -1, errors.New(errorMessage)
}

// @@ inlining call to errors.New
// @@ &errors.errorString literal escapes to heap
// @@ &errors.errorString literal escapes to heap

type ParserAssert struct {
	error
}

func Parse(query string) (commandResult command.Command, finalErr error) {
	// @@ leaking param: query
	p := Parser{Buffer: query}
	p.Init()
	// @@ moved to heap: p
	defer func() {
		// @@ p escapes to heap
		r := recover()
		if r == nil {
			return
		}
		if message, ok := r.(ParserAssert); ok {
			finalErr = message
		}
		// @@ message escapes to heap
	}()
	if err := p.Parse(); err != nil {
		// Parsing error - invalid syntax.
		// TODO - return the token where the error is occurring.
		if _, ok := err.(*parseError); ok {
			return nil, SyntaxErrors([]SyntaxError{{
				token: "",
				// @@ []SyntaxError literal escapes to heap
				message: customParseError(&p),
			}})
		} else {
			// @@ SyntaxErrors([]SyntaxError literal) escapes to heap
			// generic error (should not occur).
			return nil, AssertionError{"Non-parse error raised"}
		}
		// @@ AssertionError literal escapes to heap
	}
	p.Execute()
	if len(p.nodeStack) > 0 {
		return nil, AssertionError{"Node stack is not empty"}
	}
	// @@ AssertionError literal escapes to heap
	if len(p.errors) > 0 {
		// user error - an invalid query is provided.
		return nil, SyntaxErrors(p.errors)
	}
	// @@ SyntaxErrors(p.errors) escapes to heap
	if p.command == nil {
		// after parsing has finished, there should be a command available.
		return nil, AssertionError{"No command"}
	}
	// @@ AssertionError literal escapes to heap
	return p.command, nil
}

// Error functions
// ===============
// these functions are called to mark that an error has occurred
// while parsing or constructing command.

func (p *Parser) flagSyntaxError(err SyntaxError) {
	// @@ leaking param: err
	// @@ leaking param content: p
	// @@ leaking param content: p
	p.errors = append(p.errors, err)
	// @@ can inline (*Parser).flagSyntaxError
}

// Generic Stack Operation
// =======================
func (p *Parser) popNodeInto(target interface{}) {
	// @@ leaking param: target
	// @@ leaking param content: p
	targetValue := reflect.ValueOf(target)
	if targetValue.Type().Kind() != reflect.Ptr {
		panic(ParserAssert{ // Will unwind until it comes to "p.Parse()" which has a recover.
			fmt.Errorf("[%s] popNodeInto() given a non-pointer target", functionName(1)),
		})
		// @@ functionName(1) escapes to heap
	}
	l := len(p.nodeStack)
	if l == 0 {
		panic(ParserAssert{fmt.Errorf("[%s] popNodeInto() on an empty stack", functionName(1))})
	}
	// @@ functionName(1) escapes to heap
	node := p.nodeStack[l-1]
	p.nodeStack = p.nodeStack[:l-1]

	// @@ (*Parser).popNodeInto ignoring self-assignment to p.nodeStack
	nodeValue := reflect.ValueOf(node)
	actualType := nodeValue.Type()
	expectedType := targetValue.Elem().Type()
	if !actualType.ConvertibleTo(expectedType) {
		panic(ParserAssert{fmt.Errorf("[%s] popNodeInto() - expected %s, got off the stack %s",
			functionName(1),
			expectedType.String(),
			// @@ functionName(1) escapes to heap
			actualType.String()),
		// @@ expectedType.String() escapes to heap
		})
		// @@ actualType.String() escapes to heap
	}
	targetValue.Elem().Set(nodeValue)
}

func (p *Parser) peekNodeInto(target interface{}) {
	// @@ leaking param: target
	// @@ leaking param content: p
	targetValue := reflect.ValueOf(target)
	if targetValue.Type().Kind() != reflect.Ptr {
		panic(ParserAssert{
			fmt.Errorf("[%s] peekNodeInto() given a non-pointer target", functionName(1)),
		})
		// @@ functionName(1) escapes to heap
	}
	l := len(p.nodeStack)
	if l == 0 {
		panic(ParserAssert{fmt.Errorf("[%s] peekNodeInto() on an empty stack", functionName(1))})
	}
	// @@ functionName(1) escapes to heap

	nodeValue := reflect.ValueOf(p.nodeStack[l-1])

	expectedType := targetValue.Elem().Type()
	actualType := nodeValue.Type()
	if !actualType.ConvertibleTo(expectedType) {
		panic(ParserAssert{fmt.Errorf("[%s] peekNodeInto() - expected %s, got off the stack %s",
			functionName(1),
			expectedType.String(),
			// @@ functionName(1) escapes to heap
			actualType.String()),
		// @@ expectedType.String() escapes to heap
		})
		// @@ actualType.String() escapes to heap
	}
	targetValue.Elem().Set(nodeValue)
}

func (p *Parser) pushNode(node interface{}) {
	// @@ leaking param: node
	// @@ leaking param content: p
	// @@ leaking param content: p
	p.nodeStack = append(p.nodeStack, node)
	// @@ can inline (*Parser).pushNode
}

// Modification Operations
// =======================
// These operations are used by the embedded code snippets in language.peg
func (p *Parser) makeDescribe() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	var condition predicate.Predicate
	p.popNodeInto(&condition)
	// @@ moved to heap: condition
	var literal string
	// @@ &condition escapes to heap
	// @@ &condition escapes to heap
	p.popNodeInto(&literal)
	// @@ moved to heap: literal
	p.command = &command.DescribeCommand{
		// @@ &literal escapes to heap
		// @@ &literal escapes to heap
		MetricName: api.MetricKey(literal),
		Predicate:  condition,
	}
	// @@ &command.DescribeCommand literal escapes to heap
}

// @@ &command.DescribeCommand literal escapes to heap

func (p *Parser) makeSelect() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	var contextNode *evaluationContextNode
	p.popNodeInto(&contextNode)
	// @@ moved to heap: contextNode
	var predicate predicate.Predicate
	// @@ &contextNode escapes to heap
	// @@ &contextNode escapes to heap
	p.popNodeInto(&predicate)
	// @@ moved to heap: predicate
	var list []function.Expression
	// @@ &predicate escapes to heap
	// @@ &predicate escapes to heap
	p.popNodeInto(&list)
	// @@ moved to heap: list
	p.command = &command.SelectCommand{
		// @@ &list escapes to heap
		// @@ &list escapes to heap
		Predicate:   predicate,
		Expressions: list,
		Context: command.SelectContext{
			Start:        contextNode.Start,
			End:          contextNode.End,
			Resolution:   contextNode.Resolution,
			SampleMethod: contextNode.SampleMethod,
		},
	}
	// @@ &command.SelectCommand literal escapes to heap
}

// @@ &command.SelectCommand literal escapes to heap

func (p *Parser) addNullMatchClause() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	p.pushNode(regexp.MustCompile(""))
}

// @@ inlining call to (*Parser).pushNode
// @@ regexp.MustCompile("") escapes to heap

func (p *Parser) addMatchClause() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	compiled := p.popRegex()
	p.pushNode(compiled)
}

// @@ inlining call to (*Parser).pushNode
// @@ compiled escapes to heap

func (p *Parser) makeDescribeAll() {
	// @@ leaking param content: p
	var matcher *regexp.Regexp
	p.popNodeInto(&matcher)
	// @@ moved to heap: matcher
	p.command = &command.DescribeAllCommand{Matcher: matcher}
	// @@ &matcher escapes to heap
	// @@ &matcher escapes to heap
}

// @@ &command.DescribeAllCommand literal escapes to heap
// @@ &command.DescribeAllCommand literal escapes to heap

func (p *Parser) makeDescribeMetrics() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	// Pop off the value.
	var literal string
	p.popNodeInto(&literal)
	// @@ moved to heap: literal
	// Pop of the tag name.
	// @@ &literal escapes to heap
	// @@ &literal escapes to heap
	var tagLiteral *tagLiteral
	p.popNodeInto(&tagLiteral)
	// @@ moved to heap: tagLiteral
	p.command = &command.DescribeMetricsCommand{
		// @@ &tagLiteral escapes to heap
		// @@ &tagLiteral escapes to heap
		TagKey:   tagLiteral.tag,
		TagValue: literal,
	}
	// @@ &command.DescribeMetricsCommand literal escapes to heap
}

// @@ &command.DescribeMetricsCommand literal escapes to heap

func (p *Parser) addOperatorLiteral(operator string) {
	// @@ leaking param: operator
	// @@ leaking param content: p
	// @@ leaking param content: p
	p.pushNode(&operatorLiteral{operator})
	// @@ can inline (*Parser).addOperatorLiteral
}

// @@ inlining call to (*Parser).pushNode
// @@ &operatorLiteral literal escapes to heap
// @@ &operatorLiteral literal escapes to heap

func (p *Parser) addOperatorFunction() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	var right function.Expression

	// @@ moved to heap: right
	p.popNodeInto(&right)
	var operatorNode *operatorLiteral
	// @@ &right escapes to heap
	// @@ &right escapes to heap
	p.popNodeInto(&operatorNode)
	// @@ moved to heap: operatorNode
	var left function.Expression
	// @@ &operatorNode escapes to heap
	// @@ &operatorNode escapes to heap
	p.popNodeInto(&left)
	// @@ moved to heap: left
	p.pushNode(&expression.FunctionExpression{
		// @@ &left escapes to heap
		// @@ &left escapes to heap
		FunctionName: operatorNode.operator,
		Arguments:    []function.Expression{left, right},
	})
	// @@ &expression.FunctionExpression literal escapes to heap
	// @@ &expression.FunctionExpression literal escapes to heap
	// @@ []function.Expression literal escapes to heap
}

// @@ inlining call to (*Parser).pushNode

func (p *Parser) addPropertyKey(key string) {
	// @@ leaking param: key
	// @@ leaking param content: p
	// @@ leaking param content: p
	p.pushNode(&evaluationContextKey{key})
	// @@ can inline (*Parser).addPropertyKey
}

// @@ inlining call to (*Parser).pushNode
// @@ &evaluationContextKey literal escapes to heap
// @@ &evaluationContextKey literal escapes to heap

func (p *Parser) addPropertyValue(value string) {
	// @@ leaking param: value
	// @@ leaking param content: p
	// @@ leaking param content: p
	p.pushNode(&evaluationContextValue{value})
	// @@ can inline (*Parser).addPropertyValue
}

// @@ inlining call to (*Parser).pushNode
// @@ &evaluationContextValue literal escapes to heap
// @@ &evaluationContextValue literal escapes to heap

func (p *Parser) addEvaluationContext() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	p.pushNode(&evaluationContextNode{
		// @@ can inline (*Parser).addEvaluationContext
		0, 0, 30000,
		timeseries.SampleMean,
		make(map[string]bool),
	})
	// @@ &evaluationContextNode literal escapes to heap
	// @@ &evaluationContextNode literal escapes to heap
	// @@ make(map[string]bool) escapes to heap
}

// @@ inlining call to (*Parser).pushNode

func (p *Parser) insertPropertyKeyValue() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	var valueNode *evaluationContextValue
	p.popNodeInto(&valueNode)
	// @@ moved to heap: valueNode
	var keyNode *evaluationContextKey
	// @@ &valueNode escapes to heap
	// @@ &valueNode escapes to heap
	p.popNodeInto(&keyNode)
	// @@ moved to heap: keyNode
	var contextNode *evaluationContextNode
	// @@ &keyNode escapes to heap
	// @@ &keyNode escapes to heap
	p.popNodeInto(&contextNode)
	// @@ moved to heap: contextNode

	// @@ &contextNode escapes to heap
	// @@ &contextNode escapes to heap
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
		// @@ key escapes to heap
	}
	// @@ inlining call to (*Parser).flagSyntaxError
	contextNode.assigned[key] = true

	switch key {
	case "sample":
		// If the key is "sample", it means we're in a "sample by" declaration.
		// Only three possible sample methods are defined: min, max, or mean.
		switch value {
		case "max":
			contextNode.SampleMethod = timeseries.SampleMax
		case "min":
			contextNode.SampleMethod = timeseries.SampleMin
		case "mean":
			contextNode.SampleMethod = timeseries.SampleMean
		default:
			p.flagSyntaxError(SyntaxError{
				token:   value,
				message: fmt.Sprintf("Expected sampling method 'max', 'min', or 'mean' but got %s", value),
			})
			// @@ value escapes to heap
		}
		// @@ inlining call to (*Parser).flagSyntaxError
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
		// @@ inlining call to (*Parser).flagSyntaxError
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
			// @@ err.Error() escapes to heap
		}
		// @@ inlining call to (*Parser).flagSyntaxError
	default:
		p.flagSyntaxError(SyntaxError{
			token:   key,
			message: fmt.Sprintf("Unknown property key %s", key),
		})
		// @@ key escapes to heap
	}
	// @@ inlining call to (*Parser).flagSyntaxError
	p.pushNode(contextNode)
}

// @@ inlining call to (*Parser).pushNode
// @@ contextNode escapes to heap

// makePropertyClause verifies that all mandatory fields have been assigned in the evaluation context.
func (p *Parser) checkPropertyClause() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	var contextNode *evaluationContextNode
	p.popNodeInto(&contextNode)
	// @@ moved to heap: contextNode
	mandatoryFields := []string{"from", "to"} // Sample, resolution is optional (default to mean, 30s)
	// @@ &contextNode escapes to heap
	// @@ &contextNode escapes to heap
	for _, field := range mandatoryFields {
		if !contextNode.assigned[field] {
			p.flagSyntaxError(SyntaxError{
				token:   field,
				message: fmt.Sprintf("Field %s is never assigned in property clause", field),
			})
			// @@ field escapes to heap
		}
		// @@ inlining call to (*Parser).flagSyntaxError
	}
	p.pushNode(contextNode)
}

// @@ inlining call to (*Parser).pushNode
// @@ contextNode escapes to heap

func (p *Parser) addPipeExpression() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	var groupBy *groupByList
	p.popNodeInto(&groupBy)
	// @@ moved to heap: groupBy
	var expressionList []function.Expression
	// @@ &groupBy escapes to heap
	// @@ &groupBy escapes to heap
	p.popNodeInto(&expressionList)
	// @@ moved to heap: expressionList
	var literal string
	// @@ &expressionList escapes to heap
	// @@ &expressionList escapes to heap
	p.popNodeInto(&literal)
	// @@ moved to heap: literal
	var expressionNode function.Expression
	// @@ &literal escapes to heap
	// @@ &literal escapes to heap
	p.popNodeInto(&expressionNode)
	// @@ moved to heap: expressionNode
	p.pushNode(&expression.FunctionExpression{
		// @@ &expressionNode escapes to heap
		// @@ &expressionNode escapes to heap
		FunctionName: literal,
		Arguments:    append([]function.Expression{expressionNode}, expressionList...),
		GroupBy:      groupBy.list,
		// @@ []function.Expression literal escapes to heap
		GroupByCollapses: groupBy.collapses,
	})
	// @@ &expression.FunctionExpression literal escapes to heap
	// @@ &expression.FunctionExpression literal escapes to heap
}

// @@ inlining call to (*Parser).pushNode

func (p *Parser) addFunctionInvocation() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	var groupBy *groupByList
	p.popNodeInto(&groupBy)
	// @@ moved to heap: groupBy
	var expressionList []function.Expression
	// @@ &groupBy escapes to heap
	// @@ &groupBy escapes to heap
	p.popNodeInto(&expressionList)
	// @@ moved to heap: expressionList
	var literal string
	// @@ &expressionList escapes to heap
	// @@ &expressionList escapes to heap
	p.popNodeInto(&literal)
	// @@ moved to heap: literal
	// user-level error generation here.
	// @@ &literal escapes to heap
	// @@ &literal escapes to heap
	p.pushNode(&expression.FunctionExpression{
		FunctionName:     literal,
		Arguments:        expressionList,
		GroupBy:          groupBy.list,
		GroupByCollapses: groupBy.collapses,
	})
	// @@ &expression.FunctionExpression literal escapes to heap
	// @@ &expression.FunctionExpression literal escapes to heap
}

// @@ inlining call to (*Parser).pushNode

func (p *Parser) addAnnotationExpression(annotation string) {
	// @@ leaking param content: p
	// @@ leaking param: annotation
	// @@ leaking param content: p
	// @@ leaking param content: p
	var content function.Expression
	p.popNodeInto(&content)
	// @@ moved to heap: content
	p.pushNode(&expression.AnnotationExpression{
		// @@ &content escapes to heap
		// @@ &content escapes to heap
		Expression: content,
		Annotation: annotation,
	})
	// @@ &expression.AnnotationExpression literal escapes to heap
	// @@ &expression.AnnotationExpression literal escapes to heap
}

// @@ inlining call to (*Parser).pushNode

func (p *Parser) addMetricExpression() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	var predicateNode predicate.Predicate
	p.popNodeInto(&predicateNode)
	// @@ moved to heap: predicateNode
	var literal string
	// @@ &predicateNode escapes to heap
	// @@ &predicateNode escapes to heap
	p.popNodeInto(&literal)
	// @@ moved to heap: literal
	p.pushNode(&expression.MetricFetchExpression{
		// @@ &literal escapes to heap
		// @@ &literal escapes to heap
		MetricName: literal,
		Predicate:  predicateNode,
	})
	// @@ &expression.MetricFetchExpression literal escapes to heap
	// @@ &expression.MetricFetchExpression literal escapes to heap
}

// @@ inlining call to (*Parser).pushNode

func (p *Parser) addExpressionList() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	p.pushNode([]function.Expression{})
	// @@ can inline (*Parser).addExpressionList
}

// @@ inlining call to (*Parser).pushNode
// @@ []function.Expression literal escapes to heap
// @@ []function.Expression literal escapes to heap

func (p *Parser) appendExpression() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	var expression function.Expression
	p.popNodeInto(&expression)
	// @@ moved to heap: expression
	var list []function.Expression
	// @@ &expression escapes to heap
	// @@ &expression escapes to heap
	p.popNodeInto(&list)
	// @@ moved to heap: list
	p.pushNode(append(list, expression))
	// @@ &list escapes to heap
	// @@ &list escapes to heap
}

// @@ inlining call to (*Parser).pushNode
// @@ append(list, expression) escapes to heap

func (p *Parser) addLiteralMatcher() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	var literal string
	p.popNodeInto(&literal)
	// @@ moved to heap: literal
	var tagLiteral *tagLiteral
	// @@ &literal escapes to heap
	// @@ &literal escapes to heap

	// @@ moved to heap: tagLiteral
	p.popNodeInto(&tagLiteral)
	p.pushNode(predicate.ListMatcher{
		// @@ &tagLiteral escapes to heap
		// @@ &tagLiteral escapes to heap
		Tag: tagLiteral.tag,
		// @@ predicate.ListMatcher literal escapes to heap
		Values: []string{literal},
	})
	// @@ []string literal escapes to heap
}

// @@ inlining call to (*Parser).pushNode

func (p *Parser) addListMatcher() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	var stringLiteral *stringLiteralList

	// @@ moved to heap: stringLiteral
	p.popNodeInto(&stringLiteral)
	var tagLiteral *tagLiteral
	// @@ &stringLiteral escapes to heap
	// @@ &stringLiteral escapes to heap
	p.popNodeInto(&tagLiteral)
	// @@ moved to heap: tagLiteral
	p.pushNode(predicate.ListMatcher{
		// @@ &tagLiteral escapes to heap
		// @@ &tagLiteral escapes to heap
		Tag: tagLiteral.tag,
		// @@ predicate.ListMatcher literal escapes to heap
		Values: stringLiteral.literals,
	})
}

// @@ inlining call to (*Parser).pushNode

func (p *Parser) addRegexMatcher() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	compiled := p.popRegex()
	var tagLiteral *tagLiteral
	p.popNodeInto(&tagLiteral)
	// @@ moved to heap: tagLiteral
	p.pushNode(predicate.RegexMatcher{
		// @@ &tagLiteral escapes to heap
		// @@ &tagLiteral escapes to heap
		Tag: tagLiteral.tag,
		// @@ predicate.RegexMatcher literal escapes to heap
		Regex: compiled,
	})
}

// @@ inlining call to (*Parser).pushNode

func (p *Parser) addTagLiteral(tag string) {
	// @@ leaking param: tag
	// @@ leaking param content: p
	// @@ leaking param content: p
	p.pushNode(&tagLiteral{tag: tag})
	// @@ can inline (*Parser).addTagLiteral
}

// @@ inlining call to (*Parser).pushNode
// @@ &tagLiteral literal escapes to heap
// @@ &tagLiteral literal escapes to heap

func (p *Parser) addLiteralList() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	p.pushNode(&stringLiteralList{make([]string, 0)})
	// @@ can inline (*Parser).addLiteralList
}

// @@ inlining call to (*Parser).pushNode
// @@ &stringLiteralList literal escapes to heap
// @@ &stringLiteralList literal escapes to heap
// @@ make([]string, 0) escapes to heap

func (p *Parser) appendLiteral(literal string) {
	// @@ leaking param content: p
	// @@ leaking param: literal
	var listNode *stringLiteralList
	p.peekNodeInto(&listNode)
	// @@ moved to heap: listNode
	listNode.literals = append(listNode.literals, literal)
	// @@ &listNode escapes to heap
	// @@ &listNode escapes to heap
}

func (p *Parser) addGroupBy() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	p.pushNode(&groupByList{make([]string, 0), false})
	// @@ can inline (*Parser).addGroupBy
}

// @@ inlining call to (*Parser).pushNode
// @@ &groupByList literal escapes to heap
// @@ &groupByList literal escapes to heap
// @@ make([]string, 0) escapes to heap

func (p *Parser) appendGroupBy(literal string) {
	// @@ leaking param content: p
	// @@ leaking param: literal
	var listNode *groupByList
	p.peekNodeInto(&listNode)
	// @@ moved to heap: listNode
	listNode.list = append(listNode.list, literal)
	// @@ &listNode escapes to heap
	// @@ &listNode escapes to heap
}

func (p *Parser) appendCollapseBy(literal string) {
	// @@ leaking param content: p
	// @@ leaking param: literal
	var listNode *groupByList
	p.peekNodeInto(&listNode)
	// @@ moved to heap: listNode
	listNode.collapses = true // Switch to collapsing mode
	// @@ &listNode escapes to heap
	// @@ &listNode escapes to heap
	listNode.list = append(listNode.list, literal)
}

func (p *Parser) addNotPredicate() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	var original predicate.Predicate
	p.popNodeInto(&original)
	// @@ moved to heap: original
	p.pushNode(predicate.NotPredicate{original})
	// @@ &original escapes to heap
	// @@ &original escapes to heap
}

// @@ inlining call to (*Parser).pushNode
// @@ predicate.NotPredicate literal escapes to heap

func (p *Parser) addOrPredicate() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	var rightPredicate predicate.Predicate
	p.popNodeInto(&rightPredicate)
	// @@ moved to heap: rightPredicate
	var leftPredicate predicate.Predicate
	// @@ &rightPredicate escapes to heap
	// @@ &rightPredicate escapes to heap
	p.popNodeInto(&leftPredicate)
	// @@ moved to heap: leftPredicate
	p.pushNode(predicate.Any(leftPredicate, rightPredicate))
	// @@ &leftPredicate escapes to heap
	// @@ &leftPredicate escapes to heap
}

// @@ inlining call to (*Parser).pushNode
// @@ predicate.Any(leftPredicate, rightPredicate) escapes to heap

func (p *Parser) addNullPredicate() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	p.pushNode(predicate.All())
}

// @@ inlining call to (*Parser).pushNode
// @@ predicate.All() escapes to heap

func (p *Parser) addAndPredicate() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	var rightPredicate predicate.Predicate
	p.popNodeInto(&rightPredicate)
	// @@ moved to heap: rightPredicate
	var leftPredicate predicate.Predicate
	// @@ &rightPredicate escapes to heap
	// @@ &rightPredicate escapes to heap
	p.popNodeInto(&leftPredicate)
	// @@ moved to heap: leftPredicate
	p.pushNode(predicate.All(leftPredicate, rightPredicate))
	// @@ &leftPredicate escapes to heap
	// @@ &leftPredicate escapes to heap
}

// @@ inlining call to (*Parser).pushNode
// @@ predicate.All(leftPredicate, rightPredicate) escapes to heap

func (p *Parser) addDurationNode(value string) {
	// @@ leaking param: value
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	duration, err := function.StringToDuration(value)
	p.pushNode(expression.Duration{value, duration})
	if err != nil {
		// @@ inlining call to (*Parser).pushNode
		// @@ expression.Duration literal escapes to heap
		p.flagSyntaxError(SyntaxError{
			token:   value,
			message: fmt.Sprintf("'%s' is not a valid duration: %s", value, err.Error()),
		})
		// @@ value escapes to heap
		// @@ err.Error() escapes to heap
	}
	// @@ inlining call to (*Parser).flagSyntaxError
}

func (p *Parser) addNumberNode(value string) {
	// @@ leaking param: value
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	parsedValue, err := strconv.ParseFloat(value, 64)
	p.pushNode(expression.Scalar{parsedValue})
	if err != nil || math.IsNaN(parsedValue) {
		// @@ inlining call to (*Parser).pushNode
		// @@ expression.Scalar literal escapes to heap
		p.flagSyntaxError(SyntaxError{
			// @@ inlining call to math.IsNaN
			token:   value,
			message: fmt.Sprintf("Cannot parse the number: %s", value),
		})
		// @@ value escapes to heap
	}
	// @@ inlining call to (*Parser).flagSyntaxError
}

func (p *Parser) addStringNode(value string) {
	// @@ leaking param: value
	// @@ leaking param content: p
	// @@ leaking param content: p
	p.pushNode(expression.String{value})
	// @@ can inline (*Parser).addStringNode
}

// @@ inlining call to (*Parser).pushNode
// @@ expression.String literal escapes to heap

// Utility Stack Operations
func (p *Parser) popRegex() *regexp.Regexp {
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	var literal string
	p.popNodeInto(&literal)
	// @@ moved to heap: literal
	compiled, err := regexp.Compile(literal)
	// @@ &literal escapes to heap
	// @@ &literal escapes to heap
	if err != nil {
		p.flagSyntaxError(SyntaxError{
			token:   literal,
			message: fmt.Sprintf("Cannot parse the regex: %s", err.Error()),
		})
		// @@ err.Error() escapes to heap
		return nil
		// @@ inlining call to (*Parser).flagSyntaxError
	}
	return compiled
}

// Utility Functions
// =================

// used to unescape:
// - identifiers (no unescaping required).
// - quoted strings.
func unescapeLiteral(escaped string) string {
	// @@ leaking param: escaped to result ~r1 level=0
	// @@ leaking param: escaped to result ~r1 level=0
	// @@ leaking param: escaped
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
	// @@ leaking param content: parser
	// @@ leaking param content: parser
	type pair struct {
		start int
		end   int
	}
	tokens, error := parser.tokenTree.Error(), ""
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		// @@ make([]int, 2 * len(tokens)) escapes to heap
		// @@ make([]int, 2 * len(tokens)) escapes to heap
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
			// @@ rul3s[token.pegRule] escapes to heap
			translations[end].line, translations[end].symbol,
			// @@ translations[begin].line escapes to heap
			// @@ translations[begin].symbol escapes to heap
			line,
			// @@ translations[end].line escapes to heap
			// @@ translations[end].symbol escapes to heap
			// @@ line escapes to heap
			// @@ underline escapes to heap
			underline,
		)
	}

	return error
}

func makePrettyLine(parser *Parser, token token32, translations textPositionMap) (string, string) {
	// @@ leaking param: parser to result ~r3 level=1
	// @@ leaking param: parser to result ~r3 level=1
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
	// @@ inlining call to min
	// @@ inlining call to max
	line := parser.Buffer[lineStart:lineEnd]
	// @@ inlining call to min
	// @@ inlining call to max
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
		// @@ strings.Repeat(" ", symbolBegin) + strings.Repeat("^", length) escapes to heap
	} else {
		// multi-line error - print the firsst line and draw carets under the token until the line finishes.
		length := lineEnd - lineStart - translations[begin].symbol - 1
		if length <= 0 {
			length = 1
		}
		underline := strings.Repeat(" ", symbolBegin) + strings.Repeat("^", length)
		return line, underline
		// @@ strings.Repeat(" ", symbolBegin) + strings.Repeat("^", length) escapes to heap
	}
}

func min(x, y int) int {
	if x < y {
		// @@ can inline min
		return x
	}
	return y
}

func max(x, y int) int {
	if x > y {
		// @@ can inline max
		return x
	}
	return y
}
