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
	predicates []Predicate
}

type orPredicate struct {
	predicates []Predicate
}

type notPredicate struct {
	predicate Predicate
}

type listMatcher struct {
	tag     string
	matches []string
}

type regexMatcher struct {
	tag   string
	regex *regexp.Regexp
}

// list of literals
type literalNode struct {
	literal string
}

// list of literals
type literalListNode struct {
	literals []string
}

type tagNode struct {
	tag string
}

// nodes related to the query evaluation
// -------------------------------------
type numberNode struct {
	value float64
}

type metricReferenceNode struct {
	metricName string
	predicate  Predicate
}

type operatorLiteralNode struct {
	operator string
}

type expressionListNode struct {
	expressions []Expression
}

// functionNode represents a function call, including arithmetic operators.
type functionNode struct {
	functionName string
	arguments    []Expression
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
		printHelper(buffer, indent, "<?>")
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
		strings.Join(node.matches, ","),
	))
}

func (node *regexMatcher) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, fmt.Sprintf("%s=%s",
		node.tag,
		node.regex.String(),
	))
}

func (node *literalNode) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, node.literal)
}

func (node *literalListNode) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, strings.Join(node.literals, ","))
}

func (node *tagNode) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, fmt.Sprintf("%s", node.tag))
}

func (node *numberNode) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, fmt.Sprintf("%f", node.value))
}

func (node *operatorLiteralNode) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, node.operator)
}

func (node *expressionListNode) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	for _, expression := range node.expressions {
		printUnknown(buffer, indent+1, expression)
	}
}

func (node *functionNode) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	for _, expression := range node.arguments {
		printUnknown(buffer, indent+1, expression)
	}
}

func (node *metricReferenceNode) Print(buffer *bytes.Buffer, indent int) {
	printType(buffer, indent, node)
	printHelper(buffer, indent+1, node.metricName)
	printUnknown(buffer, indent+1, node.predicate)
}
