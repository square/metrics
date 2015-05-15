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

	"github.com/square/metrics/api"
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

// scalarExpression represents a scalar constant embedded within the expression.
type scalarExpression struct {
	value float64
}

// metricFetchExpression represents a reference to a metric embedded within the expression.
type metricFetchExpression struct {
	metricName string
	predicate  api.Predicate
}

// functionExpression represents a function call with subexpressions.
// This includes aggregate functions and arithmetic operators.
type functionExpression struct {
	functionName string
	arguments    []Expression
	groupBy      []string
}

// temporary nodes
// ---------------
// These nodes are only present during the parsing step and are not present
// in the resulting command.
// There are two types of temporary nodes:
// * literals (constants in the syntax tree).
// * lists

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
	expressions []Expression
}

type groupByList struct {
	list []string
}

// Helper functions for printing
// =============================
func printHelper(buffer *bytes.Buffer, indent int, value string) {
	for i := 0; i < indent; i++ {
		buffer.WriteString(" ")
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
		node.Print(buffer, indent+1)
	} else {
		printHelper(buffer, indent, fmt.Sprintf("%+v", object))
	}
}

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

func (node *scalarExpression) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, fmt.Sprintf("%f", node.value))
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
	for _, expression := range node.arguments {
		printUnknown(buffer, indent+1, expression)
	}
}

func (node *metricFetchExpression) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, node.metricName)
	printUnknown(buffer, indent+1, node.predicate)
}
