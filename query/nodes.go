package query

import (
	"regexp"
)

// Node is a processed AST node generated during the AST traversal.
// During Execute(), nodes are repeatedly pushed and popped
// on top of nodeStack.
type Node interface {
	Print() string
}

type andPred struct {
	predicates []Predicate
}

type orPred struct {
	predicates []Predicate
}

type notPred struct {
	predicate Predicate
}

type listMatcher struct {
	alias   string
	tag     string
	matches []string
}

type regexMatcher struct {
	alias string
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
	tag   string
	alias string
}

func (node *andPred) Print() string {
	return ""
}
func (node *orPred) Print() string {
	return ""
}
func (node *notPred) Print() string {
	return ""
}
func (node *listMatcher) Print() string {
	return ""
}
func (node *regexMatcher) Print() string {
	return ""
}
func (node *literalNode) Print() string {
	return ""
}
func (node *literalListNode) Print() string {
	return ""
}
func (node *tagRefNode) Print() string {
	return ""
}
