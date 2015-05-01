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
	"regexp"
	"strings"
)

// Node is a processed AST node generated during the AST traversal.
// During Execute(), nodes are repeatedly pushed and popped
// on top of nodeStack.
type Node interface {
	Print(buffer bytes.Buffer, indent int)
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

type numberNode struct {
	value float64
}

func printHelper(buffer bytes.Buffer, indent int, value string) {
	for i := 0; i < indent; i++ {
		buffer.WriteString(" ")
	}
	buffer.WriteString(value)
	buffer.WriteString("\n")
}

func printUnknown(buffer bytes.Buffer, indent int) {
	printHelper(buffer, indent, "<?>")
}

func (node *andPredicate) Print(buffer bytes.Buffer, indent int) {
	printHelper(buffer, indent, "andPredicate")
	for _, pred := range node.predicates {
		if node, ok := pred.(Node); ok {
			node.Print(buffer, indent+1)
		} else {
			printUnknown(buffer, indent+1)
		}
	}
}

func (node *orPredicate) Print(buffer bytes.Buffer, indent int) {
	printHelper(buffer, indent, "orPredicate")
	for _, pred := range node.predicates {
		if node, ok := pred.(Node); ok {
			node.Print(buffer, indent+1)
		} else {
			printUnknown(buffer, indent+1)
		}
	}
}

func (node *notPredicate) Print(buffer bytes.Buffer, indent int) {
	printHelper(buffer, indent, "notPredicate")
	if node, ok := node.predicate.(Node); ok {
		node.Print(buffer, indent+1)
	} else {
		printUnknown(buffer, indent+1)
	}
}

func (node *listMatcher) Print(buffer bytes.Buffer, indent int) {
	printHelper(buffer, indent, "listMatcher")
	printHelper(buffer, indent+1, fmt.Sprintf("%s=%s",
		node.tag,
		strings.Join(node.matches, ","),
	))
}

func (node *regexMatcher) Print(buffer bytes.Buffer, indent int) {
	printHelper(buffer, indent, "regexMatcher")
	printHelper(buffer, indent+1, fmt.Sprintf("%s=%s",
		node.tag,
		node.regex.String(),
	))
}
func (node *literalNode) Print(buffer bytes.Buffer, indent int) {
	printHelper(buffer, indent, "literalNode")
	printHelper(buffer, indent+1, node.literal)
}
func (node *literalListNode) Print(buffer bytes.Buffer, indent int) {
	printHelper(buffer, indent, "literalNode")
	printHelper(buffer, indent+1, strings.Join(node.literals, ","))
}
func (node *tagNode) Print(buffer bytes.Buffer, indent int) {
	printHelper(buffer, indent, "tagNode")
	printHelper(buffer, indent+1, fmt.Sprintf("%s", node.tag))
}
func (node *numberNode) Print(buffer bytes.Buffer, indent int) {
	printHelper(buffer, indent, "numberNode")
	printHelper(buffer, indent+1, fmt.Sprintf("%f", node.value))
}
