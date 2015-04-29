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

type tagRefNode struct {
	tag string
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
func (node *tagRefNode) Print(buffer bytes.Buffer, indent int) {
	printHelper(buffer, indent, "tagRefNode")
	printHelper(buffer, indent+1, fmt.Sprintf("%s", node.tag))
}
