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

package query

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/square/metrics/api"
	"github.com/square/metrics/function"
)

// PrintNode prints the given node.
func PrintNode(node Node) string {
	var buffer bytes.Buffer
	node.Print(&buffer, 0)
	return buffer.String()
}

// Node is a processed AST node generated during the AST traversal.
// During Execute(), nodes are repeatedly pushed and popped
// on top of nodeStack.
type Node interface {
	Print(buffer *bytes.Buffer, indent int)
}

// Predicates
// ----------

type andPredicate struct {
	predicates []api.Predicate
}

type orPredicate struct {
	predicates []api.Predicate
}

type notPredicate struct {
	predicate api.Predicate
}

type listMatcher struct {
	tag    string
	values []string
}

type regexMatcher struct {
	tag   string
	regex *regexp.Regexp
}

// Expressions
// -----------
// nodes related to the query evaluation
// each of these nodes are implementations of Expression interface.

// durationExpression represents a duration (in ms).
type durationExpression struct {
	name     string
	duration time.Duration // milliseconds
}

func (d durationExpression) QueryString() string {
	return d.name
}

func (d durationExpression) Name() string {
	return d.QueryString()
}

// scalarExpression represents a scalar constant embedded within the expression.
type scalarExpression struct {
	value float64
}

func (s scalarExpression) QueryString() string {
	return fmt.Sprintf("%v", s.value)
}
func (s scalarExpression) Name() string {
	return s.QueryString()
}

// stringExpression represents a string literal used as an expression.
type stringExpression struct {
	value string
}

func (s stringExpression) QueryString() string {
	return fmt.Sprintf("%q", s.value)
}
func (s stringExpression) Name() string {
	return s.QueryString()
}

// metricFetchExpression represents a reference to a metric embedded within the expression.
type metricFetchExpression struct {
	metricName string
	predicate  api.Predicate
}

var OrdinaryIdentifierRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z_0-9]*$`)

func EscapeIdentifier(identifier string) string {
	if !OrdinaryIdentifierRegex.MatchString(identifier) {
		return fmt.Sprintf("`%s`", identifier)
	}
	return identifier
}

// TODO: QueryString should indicate the associated predicate
func (m metricFetchExpression) QueryString() string {
	query := EscapeIdentifier(m.metricName)
	if m.predicate != nil {
		predicateString := m.predicate.Query()
		if predicateString != "" {
			query = fmt.Sprintf("%s[%s]", query, predicateString)
		}
	}
	return query
}
func (m metricFetchExpression) Name() string {
	return m.QueryString()
}

// functionExpression represents a function call with subexpressions.
// This includes aggregate functions and arithmetic operators.
type functionExpression struct {
	functionName     string
	arguments        []function.Expression
	groupBy          []string
	groupByCollapses bool
}

func functionFormatString(argumentStrings []string, f functionExpression) string {
	switch f.functionName {
	case "+", "-", "*", "/":
		if len(f.arguments) != 2 {
			// Then it's not actually an operator.
			break
		}
		return fmt.Sprintf("(%s %s %s)", argumentStrings[0], f.functionName, argumentStrings[1])
	}
	argumentString := strings.Join(argumentStrings, ", ")
	groupString := ""
	if len(f.groupBy) != 0 {
		groupKeyword := "group by"
		if f.groupByCollapses {
			groupKeyword = "collapse by"
		}
		escapedGroupBy := []string{}
		for _, group := range f.groupBy {
			escapedGroupBy = append(escapedGroupBy, EscapeIdentifier(group))
		}
		groupString = fmt.Sprintf(" %s %s", groupKeyword, strings.Join(escapedGroupBy, ", "))
	}
	return fmt.Sprintf("%s(%s%s)", f.functionName, argumentString, groupString)
}

// QueryString does the heavy lifting so implementations don't have to.
func (f functionExpression) QueryString() string {
	argumentStrings := []string{}
	for i := range f.arguments {
		argumentStrings = append(argumentStrings, f.arguments[i].QueryString())
	}
	return functionFormatString(argumentStrings, f)
}
func (f functionExpression) Name() string {
	// TODO: deprecate (and remove) this behavior before it becomes permanent
	if f.functionName == "transform.alias" && len(f.arguments) == 2 {
		if alias, ok := f.arguments[1].(*stringExpression); ok {
			return alias.value
		}
	}
	argumentStrings := []string{}
	for i := range f.arguments {
		argumentStrings = append(argumentStrings, f.arguments[i].Name())
	}
	return functionFormatString(argumentStrings, f)
}

type annotationExpression struct {
	content    function.Expression
	annotation string
}

func (ae annotationExpression) Evaluate(context *function.EvaluationContext) (function.Value, error) {
	return ae.content.Evaluate(context)
}

func (ae annotationExpression) QueryString() string {
	return fmt.Sprintf("%s {%s}", ae.content.QueryString(), ae.annotation)
}
func (ae annotationExpression) Name() string {
	return ae.annotation
}

// etc nodes
// ---------

type matcherClause struct {
	regex *regexp.Regexp
}

// temporary nodes
// ---------------
// These nodes are only present during the parsing step and are not present
// in the resulting command.
// There are three types of temporary nodes:
// * literals (constants in the syntax tree).
// * lists
// * evaluation context nodes

type stringLiteral struct {
	literal string
}

// list of literals
type stringLiteralList struct {
	literals []string
}

// single tag
type tagLiteral struct {
	tag string
}

// a single operator
type operatorLiteral struct {
	operator string
}

type expressionList struct {
	expressions []function.Expression
}

type groupByList struct {
	list      []string
	collapses bool
}

// evaluationContextKey represents a key (from, to, sampleby) for the evaluation context.
type evaluationContextKey struct {
	key string
}

// evaluationContextValue represents a value (date, samplingmode, etc.) for the evaluation context.
type evaluationContextValue struct {
	value string
}

// evaluationContextMap represents a collection of key-value pairs that form the evaluation context.
type evaluationContextNode struct {
	Start        int64            // Start of data timerange
	End          int64            // End of data timerange
	Resolution   int64            // Resolution of data timerange
	SampleMethod api.SampleMethod // to use when up/downsampling to match requested resolution
	assigned     map[string]bool  // a map for knowing which elements of the context have been assigned
}

// Helper functions for printing
// =============================
func printHelper(buffer *bytes.Buffer, indent int, value string) {
	for i := 0; i < indent; i++ {
		buffer.WriteString("  ")
	}
	buffer.WriteString(value)
	buffer.WriteString("\n")
}

func printType(buffer *bytes.Buffer, indent int, node Node) {
	printHelper(buffer, indent, reflect.ValueOf(node).Type().String())
}

// printUnknown is used to print an item that may or may not be Node.
func printUnknown(buffer *bytes.Buffer, indent int, object interface{}) {
	if node, ok := object.(Node); ok {
		node.Print(buffer, indent)
	} else {
		printHelper(buffer, indent, fmt.Sprintf("%+v", object))
	}
}

// Predicates

func (node *andPredicate) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	for _, pred := range node.predicates {
		printUnknown(buffer, indent+1, pred)
	}
}

func (node *orPredicate) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	for _, pred := range node.predicates {
		printUnknown(buffer, indent+1, pred)
	}
}

func (node *notPredicate) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printUnknown(buffer, indent+1, node.predicate)
}

func (node *listMatcher) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, fmt.Sprintf("%s=%s",
		node.tag,
		strings.Join(node.values, ","),
	))
}

func (node *regexMatcher) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, fmt.Sprintf("%s=%s",
		node.tag,
		node.regex.String(),
	))
}

func (node *stringLiteral) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, node.literal)
}

func (node *stringLiteralList) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, strings.Join(node.literals, ","))
}

func (node *groupByList) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, strings.Join(node.list, ","))
}

func (node *tagLiteral) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, fmt.Sprintf("%s", node.tag))
}

func (node *matcherClause) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, node.regex.String())
}

// Expressions

func (node *durationExpression) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, fmt.Sprintf("%d ms", node.duration))
}

func (node *scalarExpression) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, fmt.Sprintf("%f", node.value))
}

func (node *stringExpression) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, node.value)
}

func (node *operatorLiteral) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, node.operator)
}

func (node *expressionList) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	for _, expression := range node.expressions {
		printUnknown(buffer, indent+1, expression)
	}
}

func (node *functionExpression) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, node.functionName)
	for _, expression := range node.arguments {
		printUnknown(buffer, indent+1, expression)
	}
}

func (node *annotationExpression) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, node.annotation)
	printUnknown(buffer, indent+1, node.content)
}

func (node *metricFetchExpression) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, node.metricName)
	printUnknown(buffer, indent+1, node.predicate)
}

func (node *evaluationContextKey) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printUnknown(buffer, indent+1, node.key)
}

func (node *evaluationContextValue) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printUnknown(buffer, indent+1, node.value)
}

func (node *evaluationContextNode) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printUnknown(buffer, indent+1, node.Start)
	printUnknown(buffer, indent+1, node.End)
	printUnknown(buffer, indent+1, node.Resolution)
	printUnknown(buffer, indent+1, node.SampleMethod)
	printUnknown(buffer, indent+1, node.assigned)
}

// Commands

func (node *DescribeCommand) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	indent++
	printUnknown(buffer, indent, node.metricName)
	printUnknown(buffer, indent, node.predicate)
}

func (node *DescribeAllCommand) Print(buffer *bytes.Buffer, indent int) {
	buffer.WriteString("describe all\n")
}

func (node *SelectCommand) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)

	indent++
	printUnknown(buffer, indent, node.context)
	printUnknown(buffer, indent, node.predicate)
	for _, expr := range node.expressions {
		printUnknown(buffer, indent, expr)
	}
}
