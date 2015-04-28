package query

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const end_symbol rune = 4

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleroot
	ruleselectStmt
	ruledescribeStmt
	ruledescribeAllStmt
	ruledescribeSingleStmt
	ruleexpressionList
	ruleexpression_1
	ruleexpression_2
	ruleexpression_3
	rulegroupByClause
	ruleoptionalFromClause
	rulefromClause
	rulealiasList
	rulealiasDeclaration
	ruleoptionalPredicateClause
	rulepredicateClause
	rulepredicate_1
	rulepredicate_2
	rulepredicate_3
	ruletagMatcher
	ruleliteralString
	ruleliteralList
	ruleliteralListString
	rulepropertySource
	ruleCOLUMN_NAME
	ruleMETRIC_NAME
	ruleTAG_NAME
	ruleIDENTIFIER
	ruleID_SEGMENT
	ruleID_START
	ruleID_CONT
	ruleKEYWORD
	ruleNUMBER
	ruleINTEGER
	ruleOP_ADD
	ruleOP_SUB
	ruleOP_MULT
	ruleOP_DIV
	ruleOP_AND
	ruleOP_OR
	ruleOP_NOT
	ruleSTRING
	ruleCHAR
	rulePAREN_OPEN
	rulePAREN_CLOSE
	ruleCOMMA
	ruleCOLON
	rule_
	rule__
	ruleSPACE
	ruleAction0
	rulePegText
	ruleAction1
	ruleAction2
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
	ruleAction9
	ruleAction10
	ruleAction11
	ruleAction12
	ruleAction13
	ruleAction14
	ruleAction15
	ruleAction16
	ruleAction17

	rulePre_
	rule_In_
	rule_Suf
)

var rul3s = [...]string{
	"Unknown",
	"root",
	"selectStmt",
	"describeStmt",
	"describeAllStmt",
	"describeSingleStmt",
	"expressionList",
	"expression_1",
	"expression_2",
	"expression_3",
	"groupByClause",
	"optionalFromClause",
	"fromClause",
	"aliasList",
	"aliasDeclaration",
	"optionalPredicateClause",
	"predicateClause",
	"predicate_1",
	"predicate_2",
	"predicate_3",
	"tagMatcher",
	"literalString",
	"literalList",
	"literalListString",
	"propertySource",
	"COLUMN_NAME",
	"METRIC_NAME",
	"TAG_NAME",
	"IDENTIFIER",
	"ID_SEGMENT",
	"ID_START",
	"ID_CONT",
	"KEYWORD",
	"NUMBER",
	"INTEGER",
	"OP_ADD",
	"OP_SUB",
	"OP_MULT",
	"OP_DIV",
	"OP_AND",
	"OP_OR",
	"OP_NOT",
	"STRING",
	"CHAR",
	"PAREN_OPEN",
	"PAREN_CLOSE",
	"COMMA",
	"COLON",
	"_",
	"__",
	"SPACE",
	"Action0",
	"PegText",
	"Action1",
	"Action2",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",
	"Action10",
	"Action11",
	"Action12",
	"Action13",
	"Action14",
	"Action15",
	"Action16",
	"Action17",

	"Pre_",
	"_In_",
	"_Suf",
}

type tokenTree interface {
	Print()
	PrintSyntax()
	PrintSyntaxTree(buffer string)
	Add(rule pegRule, begin, end, next, depth int)
	Expand(index int) tokenTree
	Tokens() <-chan token32
	AST() *node32
	Error() []token32
	trim(length int)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(depth int, buffer string) {
	for node != nil {
		for c := 0; c < depth; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(string(([]rune(buffer)[node.begin:node.end]))))
		if node.up != nil {
			node.up.print(depth+1, buffer)
		}
		node = node.next
	}
}

func (ast *node32) Print(buffer string) {
	ast.print(0, buffer)
}

type element struct {
	node *node32
	down *element
}

/* ${@} bit structure for abstract syntax tree */
type token16 struct {
	pegRule
	begin, end, next int16
}

func (t *token16) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token16) isParentOf(u token16) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token16) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token16) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens16 struct {
	tree    []token16
	ordered [][]token16
}

func (t *tokens16) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens16) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens16) Order() [][]token16 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int16, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token16, len(depths)), make([]token16, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int16(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state16 struct {
	token16
	depths []int16
	leaf   bool
}

func (t *tokens16) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens16) PreOrder() (<-chan state16, [][]token16) {
	s, ordered := make(chan state16, 6), t.Order()
	go func() {
		var states [8]state16
		for i, _ := range states {
			states[i].depths = make([]int16, len(ordered))
		}
		depths, state, depth := make([]int16, len(ordered)), 0, 1
		write := func(t token16, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, int16(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token16 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token16{pegRule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token16{pegRule: rulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token16{pegRule: rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens16) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens16) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(string(([]rune(buffer)[token.begin:token.end]))))
	}
}

func (t *tokens16) Add(rule pegRule, begin, end, depth, index int) {
	t.tree[index] = token16{pegRule: rule, begin: int16(begin), end: int16(end), next: int16(depth)}
}

func (t *tokens16) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens16) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	pegRule
	begin, end, next int32
}

func (t *token32) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens32) PreOrder() (<-chan state32, [][]token32) {
	s, ordered := make(chan state32, 6), t.Order()
	go func() {
		var states [8]state32
		for i, _ := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, int32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{pegRule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{pegRule: rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(string(([]rune(buffer)[token.begin:token.end]))))
	}
}

func (t *tokens32) Add(rule pegRule, begin, end, depth, index int) {
	t.tree[index] = token32{pegRule: rule, begin: int32(begin), end: int32(end), next: int32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

func (t *tokens16) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		for i, v := range tree {
			expanded[i] = v.getToken32()
		}
		return &tokens32{tree: expanded}
	}
	return nil
}

func (t *tokens32) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	return nil
}

type Parser struct {

	// temporary variables
	nodeStack []Node  // stack of nodes used during the AST traversal. should be empty at finish.
	errors    []error // errors accumulated during the AST traversal. should be empty at finish.

	// final result
	command Command

	Buffer string
	buffer []rune
	rules  [70]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	tokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer string, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer[0:] {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p *Parser
}

func (e *parseError) Error() string {
	tokens, error := e.p.tokenTree.Error(), "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.Buffer, positions)
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf("parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n",
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			/*strconv.Quote(*/ e.p.Buffer[begin:end] /*)*/)
	}

	return error
}

func (p *Parser) PrintSyntaxTree() {
	p.tokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *Parser) Highlighter() {
	p.tokenTree.PrintSyntax()
}

func (p *Parser) Execute() {
	buffer, begin, end := p.Buffer, 0, 0
	for token := range p.tokenTree.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)

		case ruleAction0:
			p.makeDescribeAll()
		case ruleAction1:
			p.addLiteralNode(unescapeLiteral(buffer[begin:end]))
		case ruleAction2:
			p.makeDescribe()
		case ruleAction3:

		case ruleAction4:
			p.addNullPredicate()
		case ruleAction5:
			p.addAndMatcher()
		case ruleAction6:
			p.addOrMatcher()
		case ruleAction7:
			p.addNotMatcher()
		case ruleAction8:

			p.addLiteralMatcher()

		case ruleAction9:

			p.addLiteralMatcher()
			p.addNotMatcher()

		case ruleAction10:

			p.addRegexMatcher()

		case ruleAction11:

			p.addListMatcher()

		case ruleAction12:

			p.addLiteralNode(unescapeLiteral(buffer[begin:end]))

		case ruleAction13:
			p.addLiteralListNode()
		case ruleAction14:

			p.appendLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction15:
			p.addTagRefNode()
		case ruleAction16:
			p.setAlias(buffer[begin:end])
		case ruleAction17:
			p.setTag(buffer[begin:end])

		}
	}
	_, _, _ = buffer, begin, end
}

func (p *Parser) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != end_symbol {
		p.buffer = append(p.buffer, end_symbol)
	}

	var tree tokenTree = &tokens16{tree: make([]token16, math.MaxInt16)}
	position, depth, tokenIndex, buffer, _rules := 0, 0, 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokenTree = tree
		if matches {
			p.tokenTree.trim(tokenIndex)
			return nil
		}
		return &parseError{p}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule pegRule, begin int) {
		if t := tree.Expand(tokenIndex); t != nil {
			tree = t
		}
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
	}

	matchDot := func() bool {
		if buffer[position] != end_symbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 root <- <((selectStmt / describeStmt) !.)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				{
					position2, tokenIndex2, depth2 := position, tokenIndex, depth
					{
						position4 := position
						depth++
						{
							position5, tokenIndex5, depth5 := position, tokenIndex, depth
							if buffer[position] != rune('s') {
								goto l6
							}
							position++
							goto l5
						l6:
							position, tokenIndex, depth = position5, tokenIndex5, depth5
							if buffer[position] != rune('S') {
								goto l3
							}
							position++
						}
					l5:
						{
							position7, tokenIndex7, depth7 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l8
							}
							position++
							goto l7
						l8:
							position, tokenIndex, depth = position7, tokenIndex7, depth7
							if buffer[position] != rune('E') {
								goto l3
							}
							position++
						}
					l7:
						{
							position9, tokenIndex9, depth9 := position, tokenIndex, depth
							if buffer[position] != rune('l') {
								goto l10
							}
							position++
							goto l9
						l10:
							position, tokenIndex, depth = position9, tokenIndex9, depth9
							if buffer[position] != rune('L') {
								goto l3
							}
							position++
						}
					l9:
						{
							position11, tokenIndex11, depth11 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l12
							}
							position++
							goto l11
						l12:
							position, tokenIndex, depth = position11, tokenIndex11, depth11
							if buffer[position] != rune('E') {
								goto l3
							}
							position++
						}
					l11:
						{
							position13, tokenIndex13, depth13 := position, tokenIndex, depth
							if buffer[position] != rune('c') {
								goto l14
							}
							position++
							goto l13
						l14:
							position, tokenIndex, depth = position13, tokenIndex13, depth13
							if buffer[position] != rune('C') {
								goto l3
							}
							position++
						}
					l13:
						{
							position15, tokenIndex15, depth15 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l16
							}
							position++
							goto l15
						l16:
							position, tokenIndex, depth = position15, tokenIndex15, depth15
							if buffer[position] != rune('T') {
								goto l3
							}
							position++
						}
					l15:
						if !_rules[rule__]() {
							goto l3
						}
						if !_rules[ruleexpressionList]() {
							goto l3
						}
						{
							position17 := position
							depth++
							{
								position18, tokenIndex18, depth18 := position, tokenIndex, depth
								if !_rules[rule__]() {
									goto l19
								}
								{
									position20 := position
									depth++
									{
										position21, tokenIndex21, depth21 := position, tokenIndex, depth
										if buffer[position] != rune('f') {
											goto l22
										}
										position++
										goto l21
									l22:
										position, tokenIndex, depth = position21, tokenIndex21, depth21
										if buffer[position] != rune('F') {
											goto l19
										}
										position++
									}
								l21:
									{
										position23, tokenIndex23, depth23 := position, tokenIndex, depth
										if buffer[position] != rune('r') {
											goto l24
										}
										position++
										goto l23
									l24:
										position, tokenIndex, depth = position23, tokenIndex23, depth23
										if buffer[position] != rune('R') {
											goto l19
										}
										position++
									}
								l23:
									{
										position25, tokenIndex25, depth25 := position, tokenIndex, depth
										if buffer[position] != rune('o') {
											goto l26
										}
										position++
										goto l25
									l26:
										position, tokenIndex, depth = position25, tokenIndex25, depth25
										if buffer[position] != rune('O') {
											goto l19
										}
										position++
									}
								l25:
									{
										position27, tokenIndex27, depth27 := position, tokenIndex, depth
										if buffer[position] != rune('m') {
											goto l28
										}
										position++
										goto l27
									l28:
										position, tokenIndex, depth = position27, tokenIndex27, depth27
										if buffer[position] != rune('M') {
											goto l19
										}
										position++
									}
								l27:
									if !_rules[rule__]() {
										goto l19
									}
									{
										position29 := position
										depth++
										if !_rules[rulealiasDeclaration]() {
											goto l19
										}
									l30:
										{
											position31, tokenIndex31, depth31 := position, tokenIndex, depth
											if !_rules[ruleCOMMA]() {
												goto l31
											}
											if !_rules[rulealiasDeclaration]() {
												goto l31
											}
											goto l30
										l31:
											position, tokenIndex, depth = position31, tokenIndex31, depth31
										}
										depth--
										add(rulealiasList, position29)
									}
									depth--
									add(rulefromClause, position20)
								}
								goto l18
							l19:
								position, tokenIndex, depth = position18, tokenIndex18, depth18
								{
									add(ruleAction3, position)
								}
							}
						l18:
							depth--
							add(ruleoptionalFromClause, position17)
						}
						depth--
						add(ruleselectStmt, position4)
					}
					goto l2
				l3:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
					{
						position33 := position
						depth++
						{
							position34, tokenIndex34, depth34 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l35
							}
							position++
							goto l34
						l35:
							position, tokenIndex, depth = position34, tokenIndex34, depth34
							if buffer[position] != rune('D') {
								goto l0
							}
							position++
						}
					l34:
						{
							position36, tokenIndex36, depth36 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l37
							}
							position++
							goto l36
						l37:
							position, tokenIndex, depth = position36, tokenIndex36, depth36
							if buffer[position] != rune('E') {
								goto l0
							}
							position++
						}
					l36:
						{
							position38, tokenIndex38, depth38 := position, tokenIndex, depth
							if buffer[position] != rune('s') {
								goto l39
							}
							position++
							goto l38
						l39:
							position, tokenIndex, depth = position38, tokenIndex38, depth38
							if buffer[position] != rune('S') {
								goto l0
							}
							position++
						}
					l38:
						{
							position40, tokenIndex40, depth40 := position, tokenIndex, depth
							if buffer[position] != rune('c') {
								goto l41
							}
							position++
							goto l40
						l41:
							position, tokenIndex, depth = position40, tokenIndex40, depth40
							if buffer[position] != rune('C') {
								goto l0
							}
							position++
						}
					l40:
						{
							position42, tokenIndex42, depth42 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l43
							}
							position++
							goto l42
						l43:
							position, tokenIndex, depth = position42, tokenIndex42, depth42
							if buffer[position] != rune('R') {
								goto l0
							}
							position++
						}
					l42:
						{
							position44, tokenIndex44, depth44 := position, tokenIndex, depth
							if buffer[position] != rune('i') {
								goto l45
							}
							position++
							goto l44
						l45:
							position, tokenIndex, depth = position44, tokenIndex44, depth44
							if buffer[position] != rune('I') {
								goto l0
							}
							position++
						}
					l44:
						{
							position46, tokenIndex46, depth46 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l47
							}
							position++
							goto l46
						l47:
							position, tokenIndex, depth = position46, tokenIndex46, depth46
							if buffer[position] != rune('B') {
								goto l0
							}
							position++
						}
					l46:
						{
							position48, tokenIndex48, depth48 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l49
							}
							position++
							goto l48
						l49:
							position, tokenIndex, depth = position48, tokenIndex48, depth48
							if buffer[position] != rune('E') {
								goto l0
							}
							position++
						}
					l48:
						if !_rules[rule__]() {
							goto l0
						}
						{
							position50, tokenIndex50, depth50 := position, tokenIndex, depth
							{
								position52 := position
								depth++
								{
									position53, tokenIndex53, depth53 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l54
									}
									position++
									goto l53
								l54:
									position, tokenIndex, depth = position53, tokenIndex53, depth53
									if buffer[position] != rune('A') {
										goto l51
									}
									position++
								}
							l53:
								{
									position55, tokenIndex55, depth55 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l56
									}
									position++
									goto l55
								l56:
									position, tokenIndex, depth = position55, tokenIndex55, depth55
									if buffer[position] != rune('L') {
										goto l51
									}
									position++
								}
							l55:
								{
									position57, tokenIndex57, depth57 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l58
									}
									position++
									goto l57
								l58:
									position, tokenIndex, depth = position57, tokenIndex57, depth57
									if buffer[position] != rune('L') {
										goto l51
									}
									position++
								}
							l57:
								{
									add(ruleAction0, position)
								}
								depth--
								add(ruledescribeAllStmt, position52)
							}
							goto l50
						l51:
							position, tokenIndex, depth = position50, tokenIndex50, depth50
							{
								position60 := position
								depth++
								{
									position61 := position
									depth++
									if !_rules[ruleMETRIC_NAME]() {
										goto l0
									}
									depth--
									add(rulePegText, position61)
								}
								{
									add(ruleAction1, position)
								}
								{
									position63 := position
									depth++
									{
										position64, tokenIndex64, depth64 := position, tokenIndex, depth
										if !_rules[rule__]() {
											goto l65
										}
										{
											position66 := position
											depth++
											{
												position67, tokenIndex67, depth67 := position, tokenIndex, depth
												if buffer[position] != rune('w') {
													goto l68
												}
												position++
												goto l67
											l68:
												position, tokenIndex, depth = position67, tokenIndex67, depth67
												if buffer[position] != rune('W') {
													goto l65
												}
												position++
											}
										l67:
											{
												position69, tokenIndex69, depth69 := position, tokenIndex, depth
												if buffer[position] != rune('h') {
													goto l70
												}
												position++
												goto l69
											l70:
												position, tokenIndex, depth = position69, tokenIndex69, depth69
												if buffer[position] != rune('H') {
													goto l65
												}
												position++
											}
										l69:
											{
												position71, tokenIndex71, depth71 := position, tokenIndex, depth
												if buffer[position] != rune('e') {
													goto l72
												}
												position++
												goto l71
											l72:
												position, tokenIndex, depth = position71, tokenIndex71, depth71
												if buffer[position] != rune('E') {
													goto l65
												}
												position++
											}
										l71:
											{
												position73, tokenIndex73, depth73 := position, tokenIndex, depth
												if buffer[position] != rune('r') {
													goto l74
												}
												position++
												goto l73
											l74:
												position, tokenIndex, depth = position73, tokenIndex73, depth73
												if buffer[position] != rune('R') {
													goto l65
												}
												position++
											}
										l73:
											{
												position75, tokenIndex75, depth75 := position, tokenIndex, depth
												if buffer[position] != rune('e') {
													goto l76
												}
												position++
												goto l75
											l76:
												position, tokenIndex, depth = position75, tokenIndex75, depth75
												if buffer[position] != rune('E') {
													goto l65
												}
												position++
											}
										l75:
											if !_rules[rule__]() {
												goto l65
											}
											if !_rules[rulepredicate_1]() {
												goto l65
											}
											depth--
											add(rulepredicateClause, position66)
										}
										goto l64
									l65:
										position, tokenIndex, depth = position64, tokenIndex64, depth64
										{
											add(ruleAction4, position)
										}
									}
								l64:
									depth--
									add(ruleoptionalPredicateClause, position63)
								}
								{
									add(ruleAction2, position)
								}
								depth--
								add(ruledescribeSingleStmt, position60)
							}
						}
					l50:
						depth--
						add(ruledescribeStmt, position33)
					}
				}
			l2:
				{
					position79, tokenIndex79, depth79 := position, tokenIndex, depth
					if !matchDot() {
						goto l79
					}
					goto l0
				l79:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
				}
				depth--
				add(ruleroot, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 selectStmt <- <(('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T') __ expressionList optionalFromClause)> */
		nil,
		/* 2 describeStmt <- <(('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E') __ (describeAllStmt / describeSingleStmt))> */
		nil,
		/* 3 describeAllStmt <- <(('a' / 'A') ('l' / 'L') ('l' / 'L') Action0)> */
		nil,
		/* 4 describeSingleStmt <- <(<METRIC_NAME> Action1 optionalPredicateClause Action2)> */
		nil,
		/* 5 expressionList <- <(expression_1 (COMMA expression_1)*)> */
		func() bool {
			position84, tokenIndex84, depth84 := position, tokenIndex, depth
			{
				position85 := position
				depth++
				if !_rules[ruleexpression_1]() {
					goto l84
				}
			l86:
				{
					position87, tokenIndex87, depth87 := position, tokenIndex, depth
					if !_rules[ruleCOMMA]() {
						goto l87
					}
					if !_rules[ruleexpression_1]() {
						goto l87
					}
					goto l86
				l87:
					position, tokenIndex, depth = position87, tokenIndex87, depth87
				}
				depth--
				add(ruleexpressionList, position85)
			}
			return true
		l84:
			position, tokenIndex, depth = position84, tokenIndex84, depth84
			return false
		},
		/* 6 expression_1 <- <((expression_2 OP_ADD expression_1) / (expression_2 OP_SUB expression_1) / expression_2)> */
		func() bool {
			position88, tokenIndex88, depth88 := position, tokenIndex, depth
			{
				position89 := position
				depth++
				{
					position90, tokenIndex90, depth90 := position, tokenIndex, depth
					if !_rules[ruleexpression_2]() {
						goto l91
					}
					{
						position92 := position
						depth++
						if !_rules[rule_]() {
							goto l91
						}
						if buffer[position] != rune('+') {
							goto l91
						}
						position++
						if !_rules[rule_]() {
							goto l91
						}
						depth--
						add(ruleOP_ADD, position92)
					}
					if !_rules[ruleexpression_1]() {
						goto l91
					}
					goto l90
				l91:
					position, tokenIndex, depth = position90, tokenIndex90, depth90
					if !_rules[ruleexpression_2]() {
						goto l93
					}
					{
						position94 := position
						depth++
						if !_rules[rule_]() {
							goto l93
						}
						if buffer[position] != rune('-') {
							goto l93
						}
						position++
						if !_rules[rule_]() {
							goto l93
						}
						depth--
						add(ruleOP_SUB, position94)
					}
					if !_rules[ruleexpression_1]() {
						goto l93
					}
					goto l90
				l93:
					position, tokenIndex, depth = position90, tokenIndex90, depth90
					if !_rules[ruleexpression_2]() {
						goto l88
					}
				}
			l90:
				depth--
				add(ruleexpression_1, position89)
			}
			return true
		l88:
			position, tokenIndex, depth = position88, tokenIndex88, depth88
			return false
		},
		/* 7 expression_2 <- <((expression_3 OP_MULT expression_2) / (expression_3 OP_DIV expression_2) / expression_3)> */
		func() bool {
			position95, tokenIndex95, depth95 := position, tokenIndex, depth
			{
				position96 := position
				depth++
				{
					position97, tokenIndex97, depth97 := position, tokenIndex, depth
					if !_rules[ruleexpression_3]() {
						goto l98
					}
					{
						position99 := position
						depth++
						if !_rules[rule_]() {
							goto l98
						}
						if buffer[position] != rune('*') {
							goto l98
						}
						position++
						if !_rules[rule_]() {
							goto l98
						}
						depth--
						add(ruleOP_MULT, position99)
					}
					if !_rules[ruleexpression_2]() {
						goto l98
					}
					goto l97
				l98:
					position, tokenIndex, depth = position97, tokenIndex97, depth97
					if !_rules[ruleexpression_3]() {
						goto l100
					}
					{
						position101 := position
						depth++
						if !_rules[rule_]() {
							goto l100
						}
						if buffer[position] != rune('/') {
							goto l100
						}
						position++
						if !_rules[rule_]() {
							goto l100
						}
						depth--
						add(ruleOP_DIV, position101)
					}
					if !_rules[ruleexpression_2]() {
						goto l100
					}
					goto l97
				l100:
					position, tokenIndex, depth = position97, tokenIndex97, depth97
					if !_rules[ruleexpression_3]() {
						goto l95
					}
				}
			l97:
				depth--
				add(ruleexpression_2, position96)
			}
			return true
		l95:
			position, tokenIndex, depth = position95, tokenIndex95, depth95
			return false
		},
		/* 8 expression_3 <- <((IDENTIFIER PAREN_OPEN expression_1 __ groupByClause PAREN_CLOSE) / (IDENTIFIER PAREN_OPEN expressionList PAREN_CLOSE) / ((&('-' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') NUMBER) | (&('\t' | '\n' | ' ' | '(') (PAREN_OPEN expression_1 PAREN_CLOSE)) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') IDENTIFIER)))> */
		func() bool {
			position102, tokenIndex102, depth102 := position, tokenIndex, depth
			{
				position103 := position
				depth++
				{
					position104, tokenIndex104, depth104 := position, tokenIndex, depth
					if !_rules[ruleIDENTIFIER]() {
						goto l105
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l105
					}
					if !_rules[ruleexpression_1]() {
						goto l105
					}
					if !_rules[rule__]() {
						goto l105
					}
					{
						position106 := position
						depth++
						{
							position107, tokenIndex107, depth107 := position, tokenIndex, depth
							if buffer[position] != rune('g') {
								goto l108
							}
							position++
							goto l107
						l108:
							position, tokenIndex, depth = position107, tokenIndex107, depth107
							if buffer[position] != rune('G') {
								goto l105
							}
							position++
						}
					l107:
						{
							position109, tokenIndex109, depth109 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l110
							}
							position++
							goto l109
						l110:
							position, tokenIndex, depth = position109, tokenIndex109, depth109
							if buffer[position] != rune('R') {
								goto l105
							}
							position++
						}
					l109:
						{
							position111, tokenIndex111, depth111 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l112
							}
							position++
							goto l111
						l112:
							position, tokenIndex, depth = position111, tokenIndex111, depth111
							if buffer[position] != rune('O') {
								goto l105
							}
							position++
						}
					l111:
						{
							position113, tokenIndex113, depth113 := position, tokenIndex, depth
							if buffer[position] != rune('u') {
								goto l114
							}
							position++
							goto l113
						l114:
							position, tokenIndex, depth = position113, tokenIndex113, depth113
							if buffer[position] != rune('U') {
								goto l105
							}
							position++
						}
					l113:
						{
							position115, tokenIndex115, depth115 := position, tokenIndex, depth
							if buffer[position] != rune('p') {
								goto l116
							}
							position++
							goto l115
						l116:
							position, tokenIndex, depth = position115, tokenIndex115, depth115
							if buffer[position] != rune('P') {
								goto l105
							}
							position++
						}
					l115:
						if !_rules[rule__]() {
							goto l105
						}
						{
							position117, tokenIndex117, depth117 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l118
							}
							position++
							goto l117
						l118:
							position, tokenIndex, depth = position117, tokenIndex117, depth117
							if buffer[position] != rune('B') {
								goto l105
							}
							position++
						}
					l117:
						{
							position119, tokenIndex119, depth119 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l120
							}
							position++
							goto l119
						l120:
							position, tokenIndex, depth = position119, tokenIndex119, depth119
							if buffer[position] != rune('Y') {
								goto l105
							}
							position++
						}
					l119:
						if !_rules[rule__]() {
							goto l105
						}
						if !_rules[ruleCOLUMN_NAME]() {
							goto l105
						}
					l121:
						{
							position122, tokenIndex122, depth122 := position, tokenIndex, depth
							if !_rules[ruleCOMMA]() {
								goto l122
							}
							if !_rules[ruleCOLUMN_NAME]() {
								goto l122
							}
							goto l121
						l122:
							position, tokenIndex, depth = position122, tokenIndex122, depth122
						}
						depth--
						add(rulegroupByClause, position106)
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l105
					}
					goto l104
				l105:
					position, tokenIndex, depth = position104, tokenIndex104, depth104
					if !_rules[ruleIDENTIFIER]() {
						goto l123
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l123
					}
					if !_rules[ruleexpressionList]() {
						goto l123
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l123
					}
					goto l104
				l123:
					position, tokenIndex, depth = position104, tokenIndex104, depth104
					{
						switch buffer[position] {
						case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							{
								position125 := position
								depth++
								{
									position126 := position
									depth++
									{
										position127, tokenIndex127, depth127 := position, tokenIndex, depth
										if buffer[position] != rune('0') {
											goto l128
										}
										position++
										goto l127
									l128:
										position, tokenIndex, depth = position127, tokenIndex127, depth127
										{
											position129, tokenIndex129, depth129 := position, tokenIndex, depth
											if buffer[position] != rune('-') {
												goto l129
											}
											position++
											goto l130
										l129:
											position, tokenIndex, depth = position129, tokenIndex129, depth129
										}
									l130:
										if c := buffer[position]; c < rune('1') || c > rune('9') {
											goto l102
										}
										position++
									l131:
										{
											position132, tokenIndex132, depth132 := position, tokenIndex, depth
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l132
											}
											position++
											goto l131
										l132:
											position, tokenIndex, depth = position132, tokenIndex132, depth132
										}
									}
								l127:
									depth--
									add(ruleINTEGER, position126)
								}
								depth--
								add(ruleNUMBER, position125)
							}
							break
						case '\t', '\n', ' ', '(':
							if !_rules[rulePAREN_OPEN]() {
								goto l102
							}
							if !_rules[ruleexpression_1]() {
								goto l102
							}
							if !_rules[rulePAREN_CLOSE]() {
								goto l102
							}
							break
						default:
							if !_rules[ruleIDENTIFIER]() {
								goto l102
							}
							break
						}
					}

				}
			l104:
				depth--
				add(ruleexpression_3, position103)
			}
			return true
		l102:
			position, tokenIndex, depth = position102, tokenIndex102, depth102
			return false
		},
		/* 9 groupByClause <- <(('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P') __ (('b' / 'B') ('y' / 'Y')) __ COLUMN_NAME (COMMA COLUMN_NAME)*)> */
		nil,
		/* 10 optionalFromClause <- <((__ fromClause) / Action3)> */
		nil,
		/* 11 fromClause <- <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M') __ aliasList)> */
		nil,
		/* 12 aliasList <- <(aliasDeclaration (COMMA aliasDeclaration)*)> */
		nil,
		/* 13 aliasDeclaration <- <(METRIC_NAME __ (('a' / 'A') ('s' / 'S')) __ IDENTIFIER)> */
		func() bool {
			position137, tokenIndex137, depth137 := position, tokenIndex, depth
			{
				position138 := position
				depth++
				if !_rules[ruleMETRIC_NAME]() {
					goto l137
				}
				if !_rules[rule__]() {
					goto l137
				}
				{
					position139, tokenIndex139, depth139 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l140
					}
					position++
					goto l139
				l140:
					position, tokenIndex, depth = position139, tokenIndex139, depth139
					if buffer[position] != rune('A') {
						goto l137
					}
					position++
				}
			l139:
				{
					position141, tokenIndex141, depth141 := position, tokenIndex, depth
					if buffer[position] != rune('s') {
						goto l142
					}
					position++
					goto l141
				l142:
					position, tokenIndex, depth = position141, tokenIndex141, depth141
					if buffer[position] != rune('S') {
						goto l137
					}
					position++
				}
			l141:
				if !_rules[rule__]() {
					goto l137
				}
				if !_rules[ruleIDENTIFIER]() {
					goto l137
				}
				depth--
				add(rulealiasDeclaration, position138)
			}
			return true
		l137:
			position, tokenIndex, depth = position137, tokenIndex137, depth137
			return false
		},
		/* 14 optionalPredicateClause <- <((__ predicateClause) / Action4)> */
		nil,
		/* 15 predicateClause <- <(('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E') __ predicate_1)> */
		nil,
		/* 16 predicate_1 <- <((predicate_2 OP_AND predicate_1 Action5) / predicate_2 / )> */
		func() bool {
			{
				position146 := position
				depth++
				{
					position147, tokenIndex147, depth147 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l148
					}
					{
						position149 := position
						depth++
						if !_rules[rule__]() {
							goto l148
						}
						{
							position150, tokenIndex150, depth150 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l151
							}
							position++
							goto l150
						l151:
							position, tokenIndex, depth = position150, tokenIndex150, depth150
							if buffer[position] != rune('A') {
								goto l148
							}
							position++
						}
					l150:
						{
							position152, tokenIndex152, depth152 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l153
							}
							position++
							goto l152
						l153:
							position, tokenIndex, depth = position152, tokenIndex152, depth152
							if buffer[position] != rune('N') {
								goto l148
							}
							position++
						}
					l152:
						{
							position154, tokenIndex154, depth154 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l155
							}
							position++
							goto l154
						l155:
							position, tokenIndex, depth = position154, tokenIndex154, depth154
							if buffer[position] != rune('D') {
								goto l148
							}
							position++
						}
					l154:
						if !_rules[rule__]() {
							goto l148
						}
						depth--
						add(ruleOP_AND, position149)
					}
					if !_rules[rulepredicate_1]() {
						goto l148
					}
					{
						add(ruleAction5, position)
					}
					goto l147
				l148:
					position, tokenIndex, depth = position147, tokenIndex147, depth147
					if !_rules[rulepredicate_2]() {
						goto l157
					}
					goto l147
				l157:
					position, tokenIndex, depth = position147, tokenIndex147, depth147
				}
			l147:
				depth--
				add(rulepredicate_1, position146)
			}
			return true
		},
		/* 17 predicate_2 <- <((predicate_3 OP_OR predicate_2 Action6) / predicate_3)> */
		func() bool {
			position158, tokenIndex158, depth158 := position, tokenIndex, depth
			{
				position159 := position
				depth++
				{
					position160, tokenIndex160, depth160 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l161
					}
					{
						position162 := position
						depth++
						if !_rules[rule__]() {
							goto l161
						}
						{
							position163, tokenIndex163, depth163 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l164
							}
							position++
							goto l163
						l164:
							position, tokenIndex, depth = position163, tokenIndex163, depth163
							if buffer[position] != rune('O') {
								goto l161
							}
							position++
						}
					l163:
						{
							position165, tokenIndex165, depth165 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l166
							}
							position++
							goto l165
						l166:
							position, tokenIndex, depth = position165, tokenIndex165, depth165
							if buffer[position] != rune('R') {
								goto l161
							}
							position++
						}
					l165:
						if !_rules[rule__]() {
							goto l161
						}
						depth--
						add(ruleOP_OR, position162)
					}
					if !_rules[rulepredicate_2]() {
						goto l161
					}
					{
						add(ruleAction6, position)
					}
					goto l160
				l161:
					position, tokenIndex, depth = position160, tokenIndex160, depth160
					if !_rules[rulepredicate_3]() {
						goto l158
					}
				}
			l160:
				depth--
				add(rulepredicate_2, position159)
			}
			return true
		l158:
			position, tokenIndex, depth = position158, tokenIndex158, depth158
			return false
		},
		/* 18 predicate_3 <- <((OP_NOT predicate_3 Action7) / (PAREN_OPEN predicate_1 PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position168, tokenIndex168, depth168 := position, tokenIndex, depth
			{
				position169 := position
				depth++
				{
					position170, tokenIndex170, depth170 := position, tokenIndex, depth
					{
						position172 := position
						depth++
						{
							position173, tokenIndex173, depth173 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l174
							}
							position++
							goto l173
						l174:
							position, tokenIndex, depth = position173, tokenIndex173, depth173
							if buffer[position] != rune('N') {
								goto l171
							}
							position++
						}
					l173:
						{
							position175, tokenIndex175, depth175 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l176
							}
							position++
							goto l175
						l176:
							position, tokenIndex, depth = position175, tokenIndex175, depth175
							if buffer[position] != rune('O') {
								goto l171
							}
							position++
						}
					l175:
						{
							position177, tokenIndex177, depth177 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l178
							}
							position++
							goto l177
						l178:
							position, tokenIndex, depth = position177, tokenIndex177, depth177
							if buffer[position] != rune('T') {
								goto l171
							}
							position++
						}
					l177:
						if !_rules[rule__]() {
							goto l171
						}
						depth--
						add(ruleOP_NOT, position172)
					}
					if !_rules[rulepredicate_3]() {
						goto l171
					}
					{
						add(ruleAction7, position)
					}
					goto l170
				l171:
					position, tokenIndex, depth = position170, tokenIndex170, depth170
					if !_rules[rulePAREN_OPEN]() {
						goto l180
					}
					if !_rules[rulepredicate_1]() {
						goto l180
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l180
					}
					goto l170
				l180:
					position, tokenIndex, depth = position170, tokenIndex170, depth170
					{
						position181 := position
						depth++
						{
							position182, tokenIndex182, depth182 := position, tokenIndex, depth
							if !_rules[rulepropertySource]() {
								goto l183
							}
							if !_rules[rule_]() {
								goto l183
							}
							if buffer[position] != rune('=') {
								goto l183
							}
							position++
							if !_rules[rule_]() {
								goto l183
							}
							if !_rules[ruleliteralString]() {
								goto l183
							}
							{
								add(ruleAction8, position)
							}
							goto l182
						l183:
							position, tokenIndex, depth = position182, tokenIndex182, depth182
							if !_rules[rulepropertySource]() {
								goto l185
							}
							if !_rules[rule_]() {
								goto l185
							}
							if buffer[position] != rune('!') {
								goto l185
							}
							position++
							if buffer[position] != rune('=') {
								goto l185
							}
							position++
							if !_rules[rule_]() {
								goto l185
							}
							if !_rules[ruleliteralString]() {
								goto l185
							}
							{
								add(ruleAction9, position)
							}
							goto l182
						l185:
							position, tokenIndex, depth = position182, tokenIndex182, depth182
							if !_rules[rulepropertySource]() {
								goto l187
							}
							if !_rules[rule__]() {
								goto l187
							}
							{
								position188, tokenIndex188, depth188 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l189
								}
								position++
								goto l188
							l189:
								position, tokenIndex, depth = position188, tokenIndex188, depth188
								if buffer[position] != rune('M') {
									goto l187
								}
								position++
							}
						l188:
							{
								position190, tokenIndex190, depth190 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l191
								}
								position++
								goto l190
							l191:
								position, tokenIndex, depth = position190, tokenIndex190, depth190
								if buffer[position] != rune('A') {
									goto l187
								}
								position++
							}
						l190:
							{
								position192, tokenIndex192, depth192 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l193
								}
								position++
								goto l192
							l193:
								position, tokenIndex, depth = position192, tokenIndex192, depth192
								if buffer[position] != rune('T') {
									goto l187
								}
								position++
							}
						l192:
							{
								position194, tokenIndex194, depth194 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l195
								}
								position++
								goto l194
							l195:
								position, tokenIndex, depth = position194, tokenIndex194, depth194
								if buffer[position] != rune('C') {
									goto l187
								}
								position++
							}
						l194:
							{
								position196, tokenIndex196, depth196 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l197
								}
								position++
								goto l196
							l197:
								position, tokenIndex, depth = position196, tokenIndex196, depth196
								if buffer[position] != rune('H') {
									goto l187
								}
								position++
							}
						l196:
							{
								position198, tokenIndex198, depth198 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l199
								}
								position++
								goto l198
							l199:
								position, tokenIndex, depth = position198, tokenIndex198, depth198
								if buffer[position] != rune('E') {
									goto l187
								}
								position++
							}
						l198:
							{
								position200, tokenIndex200, depth200 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l201
								}
								position++
								goto l200
							l201:
								position, tokenIndex, depth = position200, tokenIndex200, depth200
								if buffer[position] != rune('S') {
									goto l187
								}
								position++
							}
						l200:
							if !_rules[rule__]() {
								goto l187
							}
							if !_rules[ruleliteralString]() {
								goto l187
							}
							{
								add(ruleAction10, position)
							}
							goto l182
						l187:
							position, tokenIndex, depth = position182, tokenIndex182, depth182
							if !_rules[rulepropertySource]() {
								goto l168
							}
							if !_rules[rule__]() {
								goto l168
							}
							{
								position203, tokenIndex203, depth203 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l204
								}
								position++
								goto l203
							l204:
								position, tokenIndex, depth = position203, tokenIndex203, depth203
								if buffer[position] != rune('I') {
									goto l168
								}
								position++
							}
						l203:
							{
								position205, tokenIndex205, depth205 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l206
								}
								position++
								goto l205
							l206:
								position, tokenIndex, depth = position205, tokenIndex205, depth205
								if buffer[position] != rune('N') {
									goto l168
								}
								position++
							}
						l205:
							if !_rules[rule__]() {
								goto l168
							}
							{
								position207 := position
								depth++
								{
									add(ruleAction13, position)
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l168
								}
								if !_rules[ruleliteralListString]() {
									goto l168
								}
							l209:
								{
									position210, tokenIndex210, depth210 := position, tokenIndex, depth
									if !_rules[ruleCOMMA]() {
										goto l210
									}
									if !_rules[ruleliteralListString]() {
										goto l210
									}
									goto l209
								l210:
									position, tokenIndex, depth = position210, tokenIndex210, depth210
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l168
								}
								depth--
								add(ruleliteralList, position207)
							}
							{
								add(ruleAction11, position)
							}
						}
					l182:
						depth--
						add(ruletagMatcher, position181)
					}
				}
			l170:
				depth--
				add(rulepredicate_3, position169)
			}
			return true
		l168:
			position, tokenIndex, depth = position168, tokenIndex168, depth168
			return false
		},
		/* 19 tagMatcher <- <((propertySource _ '=' _ literalString Action8) / (propertySource _ ('!' '=') _ literalString Action9) / (propertySource __ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) __ literalString Action10) / (propertySource __ (('i' / 'I') ('n' / 'N')) __ literalList Action11))> */
		nil,
		/* 20 literalString <- <(<STRING> Action12)> */
		func() bool {
			position213, tokenIndex213, depth213 := position, tokenIndex, depth
			{
				position214 := position
				depth++
				{
					position215 := position
					depth++
					if !_rules[ruleSTRING]() {
						goto l213
					}
					depth--
					add(rulePegText, position215)
				}
				{
					add(ruleAction12, position)
				}
				depth--
				add(ruleliteralString, position214)
			}
			return true
		l213:
			position, tokenIndex, depth = position213, tokenIndex213, depth213
			return false
		},
		/* 21 literalList <- <(Action13 PAREN_OPEN literalListString (COMMA literalListString)* PAREN_CLOSE)> */
		nil,
		/* 22 literalListString <- <(STRING Action14)> */
		func() bool {
			position218, tokenIndex218, depth218 := position, tokenIndex, depth
			{
				position219 := position
				depth++
				if !_rules[ruleSTRING]() {
					goto l218
				}
				{
					add(ruleAction14, position)
				}
				depth--
				add(ruleliteralListString, position219)
			}
			return true
		l218:
			position, tokenIndex, depth = position218, tokenIndex218, depth218
			return false
		},
		/* 23 propertySource <- <(Action15 (<IDENTIFIER> Action16 COLON)? <TAG_NAME> Action17)> */
		func() bool {
			position221, tokenIndex221, depth221 := position, tokenIndex, depth
			{
				position222 := position
				depth++
				{
					add(ruleAction15, position)
				}
				{
					position224, tokenIndex224, depth224 := position, tokenIndex, depth
					{
						position226 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l224
						}
						depth--
						add(rulePegText, position226)
					}
					{
						add(ruleAction16, position)
					}
					{
						position228 := position
						depth++
						if !_rules[rule_]() {
							goto l224
						}
						if buffer[position] != rune(':') {
							goto l224
						}
						position++
						if !_rules[rule_]() {
							goto l224
						}
						depth--
						add(ruleCOLON, position228)
					}
					goto l225
				l224:
					position, tokenIndex, depth = position224, tokenIndex224, depth224
				}
			l225:
				{
					position229 := position
					depth++
					{
						position230 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l221
						}
						depth--
						add(ruleTAG_NAME, position230)
					}
					depth--
					add(rulePegText, position229)
				}
				{
					add(ruleAction17, position)
				}
				depth--
				add(rulepropertySource, position222)
			}
			return true
		l221:
			position, tokenIndex, depth = position221, tokenIndex221, depth221
			return false
		},
		/* 24 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position232, tokenIndex232, depth232 := position, tokenIndex, depth
			{
				position233 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l232
				}
				depth--
				add(ruleCOLUMN_NAME, position233)
			}
			return true
		l232:
			position, tokenIndex, depth = position232, tokenIndex232, depth232
			return false
		},
		/* 25 METRIC_NAME <- <IDENTIFIER> */
		func() bool {
			position234, tokenIndex234, depth234 := position, tokenIndex, depth
			{
				position235 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l234
				}
				depth--
				add(ruleMETRIC_NAME, position235)
			}
			return true
		l234:
			position, tokenIndex, depth = position234, tokenIndex234, depth234
			return false
		},
		/* 26 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 27 IDENTIFIER <- <((!KEYWORD ID_SEGMENT ('.' ID_SEGMENT)*) / ('`' CHAR* '`'))> */
		func() bool {
			position237, tokenIndex237, depth237 := position, tokenIndex, depth
			{
				position238 := position
				depth++
				{
					position239, tokenIndex239, depth239 := position, tokenIndex, depth
					{
						position241, tokenIndex241, depth241 := position, tokenIndex, depth
						{
							position242 := position
							depth++
							{
								position243, tokenIndex243, depth243 := position, tokenIndex, depth
								{
									position245, tokenIndex245, depth245 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l246
									}
									position++
									goto l245
								l246:
									position, tokenIndex, depth = position245, tokenIndex245, depth245
									if buffer[position] != rune('A') {
										goto l244
									}
									position++
								}
							l245:
								{
									position247, tokenIndex247, depth247 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l248
									}
									position++
									goto l247
								l248:
									position, tokenIndex, depth = position247, tokenIndex247, depth247
									if buffer[position] != rune('L') {
										goto l244
									}
									position++
								}
							l247:
								{
									position249, tokenIndex249, depth249 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l250
									}
									position++
									goto l249
								l250:
									position, tokenIndex, depth = position249, tokenIndex249, depth249
									if buffer[position] != rune('L') {
										goto l244
									}
									position++
								}
							l249:
								goto l243
							l244:
								position, tokenIndex, depth = position243, tokenIndex243, depth243
								{
									position252, tokenIndex252, depth252 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l253
									}
									position++
									goto l252
								l253:
									position, tokenIndex, depth = position252, tokenIndex252, depth252
									if buffer[position] != rune('A') {
										goto l251
									}
									position++
								}
							l252:
								{
									position254, tokenIndex254, depth254 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l255
									}
									position++
									goto l254
								l255:
									position, tokenIndex, depth = position254, tokenIndex254, depth254
									if buffer[position] != rune('N') {
										goto l251
									}
									position++
								}
							l254:
								{
									position256, tokenIndex256, depth256 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l257
									}
									position++
									goto l256
								l257:
									position, tokenIndex, depth = position256, tokenIndex256, depth256
									if buffer[position] != rune('D') {
										goto l251
									}
									position++
								}
							l256:
								goto l243
							l251:
								position, tokenIndex, depth = position243, tokenIndex243, depth243
								{
									switch buffer[position] {
									case 'W', 'w':
										{
											position259, tokenIndex259, depth259 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l260
											}
											position++
											goto l259
										l260:
											position, tokenIndex, depth = position259, tokenIndex259, depth259
											if buffer[position] != rune('W') {
												goto l241
											}
											position++
										}
									l259:
										{
											position261, tokenIndex261, depth261 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l262
											}
											position++
											goto l261
										l262:
											position, tokenIndex, depth = position261, tokenIndex261, depth261
											if buffer[position] != rune('H') {
												goto l241
											}
											position++
										}
									l261:
										{
											position263, tokenIndex263, depth263 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l264
											}
											position++
											goto l263
										l264:
											position, tokenIndex, depth = position263, tokenIndex263, depth263
											if buffer[position] != rune('E') {
												goto l241
											}
											position++
										}
									l263:
										{
											position265, tokenIndex265, depth265 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l266
											}
											position++
											goto l265
										l266:
											position, tokenIndex, depth = position265, tokenIndex265, depth265
											if buffer[position] != rune('R') {
												goto l241
											}
											position++
										}
									l265:
										{
											position267, tokenIndex267, depth267 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l268
											}
											position++
											goto l267
										l268:
											position, tokenIndex, depth = position267, tokenIndex267, depth267
											if buffer[position] != rune('E') {
												goto l241
											}
											position++
										}
									l267:
										break
									case 'S', 's':
										{
											position269, tokenIndex269, depth269 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l270
											}
											position++
											goto l269
										l270:
											position, tokenIndex, depth = position269, tokenIndex269, depth269
											if buffer[position] != rune('S') {
												goto l241
											}
											position++
										}
									l269:
										{
											position271, tokenIndex271, depth271 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l272
											}
											position++
											goto l271
										l272:
											position, tokenIndex, depth = position271, tokenIndex271, depth271
											if buffer[position] != rune('E') {
												goto l241
											}
											position++
										}
									l271:
										{
											position273, tokenIndex273, depth273 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l274
											}
											position++
											goto l273
										l274:
											position, tokenIndex, depth = position273, tokenIndex273, depth273
											if buffer[position] != rune('L') {
												goto l241
											}
											position++
										}
									l273:
										{
											position275, tokenIndex275, depth275 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l276
											}
											position++
											goto l275
										l276:
											position, tokenIndex, depth = position275, tokenIndex275, depth275
											if buffer[position] != rune('E') {
												goto l241
											}
											position++
										}
									l275:
										{
											position277, tokenIndex277, depth277 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l278
											}
											position++
											goto l277
										l278:
											position, tokenIndex, depth = position277, tokenIndex277, depth277
											if buffer[position] != rune('C') {
												goto l241
											}
											position++
										}
									l277:
										{
											position279, tokenIndex279, depth279 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l280
											}
											position++
											goto l279
										l280:
											position, tokenIndex, depth = position279, tokenIndex279, depth279
											if buffer[position] != rune('T') {
												goto l241
											}
											position++
										}
									l279:
										break
									case 'O', 'o':
										{
											position281, tokenIndex281, depth281 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l282
											}
											position++
											goto l281
										l282:
											position, tokenIndex, depth = position281, tokenIndex281, depth281
											if buffer[position] != rune('O') {
												goto l241
											}
											position++
										}
									l281:
										{
											position283, tokenIndex283, depth283 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l284
											}
											position++
											goto l283
										l284:
											position, tokenIndex, depth = position283, tokenIndex283, depth283
											if buffer[position] != rune('R') {
												goto l241
											}
											position++
										}
									l283:
										break
									case 'N', 'n':
										{
											position285, tokenIndex285, depth285 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l286
											}
											position++
											goto l285
										l286:
											position, tokenIndex, depth = position285, tokenIndex285, depth285
											if buffer[position] != rune('N') {
												goto l241
											}
											position++
										}
									l285:
										{
											position287, tokenIndex287, depth287 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l288
											}
											position++
											goto l287
										l288:
											position, tokenIndex, depth = position287, tokenIndex287, depth287
											if buffer[position] != rune('O') {
												goto l241
											}
											position++
										}
									l287:
										{
											position289, tokenIndex289, depth289 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l290
											}
											position++
											goto l289
										l290:
											position, tokenIndex, depth = position289, tokenIndex289, depth289
											if buffer[position] != rune('T') {
												goto l241
											}
											position++
										}
									l289:
										break
									case 'M', 'm':
										{
											position291, tokenIndex291, depth291 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l292
											}
											position++
											goto l291
										l292:
											position, tokenIndex, depth = position291, tokenIndex291, depth291
											if buffer[position] != rune('M') {
												goto l241
											}
											position++
										}
									l291:
										{
											position293, tokenIndex293, depth293 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l294
											}
											position++
											goto l293
										l294:
											position, tokenIndex, depth = position293, tokenIndex293, depth293
											if buffer[position] != rune('A') {
												goto l241
											}
											position++
										}
									l293:
										{
											position295, tokenIndex295, depth295 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l296
											}
											position++
											goto l295
										l296:
											position, tokenIndex, depth = position295, tokenIndex295, depth295
											if buffer[position] != rune('T') {
												goto l241
											}
											position++
										}
									l295:
										{
											position297, tokenIndex297, depth297 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l298
											}
											position++
											goto l297
										l298:
											position, tokenIndex, depth = position297, tokenIndex297, depth297
											if buffer[position] != rune('C') {
												goto l241
											}
											position++
										}
									l297:
										{
											position299, tokenIndex299, depth299 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l300
											}
											position++
											goto l299
										l300:
											position, tokenIndex, depth = position299, tokenIndex299, depth299
											if buffer[position] != rune('H') {
												goto l241
											}
											position++
										}
									l299:
										{
											position301, tokenIndex301, depth301 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l302
											}
											position++
											goto l301
										l302:
											position, tokenIndex, depth = position301, tokenIndex301, depth301
											if buffer[position] != rune('E') {
												goto l241
											}
											position++
										}
									l301:
										{
											position303, tokenIndex303, depth303 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l304
											}
											position++
											goto l303
										l304:
											position, tokenIndex, depth = position303, tokenIndex303, depth303
											if buffer[position] != rune('S') {
												goto l241
											}
											position++
										}
									l303:
										break
									case 'I', 'i':
										{
											position305, tokenIndex305, depth305 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l306
											}
											position++
											goto l305
										l306:
											position, tokenIndex, depth = position305, tokenIndex305, depth305
											if buffer[position] != rune('I') {
												goto l241
											}
											position++
										}
									l305:
										{
											position307, tokenIndex307, depth307 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l308
											}
											position++
											goto l307
										l308:
											position, tokenIndex, depth = position307, tokenIndex307, depth307
											if buffer[position] != rune('N') {
												goto l241
											}
											position++
										}
									l307:
										break
									case 'G', 'g':
										{
											position309, tokenIndex309, depth309 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l310
											}
											position++
											goto l309
										l310:
											position, tokenIndex, depth = position309, tokenIndex309, depth309
											if buffer[position] != rune('G') {
												goto l241
											}
											position++
										}
									l309:
										{
											position311, tokenIndex311, depth311 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l312
											}
											position++
											goto l311
										l312:
											position, tokenIndex, depth = position311, tokenIndex311, depth311
											if buffer[position] != rune('R') {
												goto l241
											}
											position++
										}
									l311:
										{
											position313, tokenIndex313, depth313 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l314
											}
											position++
											goto l313
										l314:
											position, tokenIndex, depth = position313, tokenIndex313, depth313
											if buffer[position] != rune('O') {
												goto l241
											}
											position++
										}
									l313:
										{
											position315, tokenIndex315, depth315 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l316
											}
											position++
											goto l315
										l316:
											position, tokenIndex, depth = position315, tokenIndex315, depth315
											if buffer[position] != rune('U') {
												goto l241
											}
											position++
										}
									l315:
										{
											position317, tokenIndex317, depth317 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l318
											}
											position++
											goto l317
										l318:
											position, tokenIndex, depth = position317, tokenIndex317, depth317
											if buffer[position] != rune('P') {
												goto l241
											}
											position++
										}
									l317:
										break
									case 'F', 'f':
										{
											position319, tokenIndex319, depth319 := position, tokenIndex, depth
											if buffer[position] != rune('f') {
												goto l320
											}
											position++
											goto l319
										l320:
											position, tokenIndex, depth = position319, tokenIndex319, depth319
											if buffer[position] != rune('F') {
												goto l241
											}
											position++
										}
									l319:
										{
											position321, tokenIndex321, depth321 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l322
											}
											position++
											goto l321
										l322:
											position, tokenIndex, depth = position321, tokenIndex321, depth321
											if buffer[position] != rune('R') {
												goto l241
											}
											position++
										}
									l321:
										{
											position323, tokenIndex323, depth323 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l324
											}
											position++
											goto l323
										l324:
											position, tokenIndex, depth = position323, tokenIndex323, depth323
											if buffer[position] != rune('O') {
												goto l241
											}
											position++
										}
									l323:
										{
											position325, tokenIndex325, depth325 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l326
											}
											position++
											goto l325
										l326:
											position, tokenIndex, depth = position325, tokenIndex325, depth325
											if buffer[position] != rune('M') {
												goto l241
											}
											position++
										}
									l325:
										break
									case 'D', 'd':
										{
											position327, tokenIndex327, depth327 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l328
											}
											position++
											goto l327
										l328:
											position, tokenIndex, depth = position327, tokenIndex327, depth327
											if buffer[position] != rune('D') {
												goto l241
											}
											position++
										}
									l327:
										{
											position329, tokenIndex329, depth329 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l330
											}
											position++
											goto l329
										l330:
											position, tokenIndex, depth = position329, tokenIndex329, depth329
											if buffer[position] != rune('E') {
												goto l241
											}
											position++
										}
									l329:
										{
											position331, tokenIndex331, depth331 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l332
											}
											position++
											goto l331
										l332:
											position, tokenIndex, depth = position331, tokenIndex331, depth331
											if buffer[position] != rune('S') {
												goto l241
											}
											position++
										}
									l331:
										{
											position333, tokenIndex333, depth333 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l334
											}
											position++
											goto l333
										l334:
											position, tokenIndex, depth = position333, tokenIndex333, depth333
											if buffer[position] != rune('C') {
												goto l241
											}
											position++
										}
									l333:
										{
											position335, tokenIndex335, depth335 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l336
											}
											position++
											goto l335
										l336:
											position, tokenIndex, depth = position335, tokenIndex335, depth335
											if buffer[position] != rune('R') {
												goto l241
											}
											position++
										}
									l335:
										{
											position337, tokenIndex337, depth337 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l338
											}
											position++
											goto l337
										l338:
											position, tokenIndex, depth = position337, tokenIndex337, depth337
											if buffer[position] != rune('I') {
												goto l241
											}
											position++
										}
									l337:
										{
											position339, tokenIndex339, depth339 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l340
											}
											position++
											goto l339
										l340:
											position, tokenIndex, depth = position339, tokenIndex339, depth339
											if buffer[position] != rune('B') {
												goto l241
											}
											position++
										}
									l339:
										{
											position341, tokenIndex341, depth341 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l342
											}
											position++
											goto l341
										l342:
											position, tokenIndex, depth = position341, tokenIndex341, depth341
											if buffer[position] != rune('E') {
												goto l241
											}
											position++
										}
									l341:
										break
									case 'B', 'b':
										{
											position343, tokenIndex343, depth343 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l344
											}
											position++
											goto l343
										l344:
											position, tokenIndex, depth = position343, tokenIndex343, depth343
											if buffer[position] != rune('B') {
												goto l241
											}
											position++
										}
									l343:
										{
											position345, tokenIndex345, depth345 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l346
											}
											position++
											goto l345
										l346:
											position, tokenIndex, depth = position345, tokenIndex345, depth345
											if buffer[position] != rune('Y') {
												goto l241
											}
											position++
										}
									l345:
										break
									default:
										{
											position347, tokenIndex347, depth347 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l348
											}
											position++
											goto l347
										l348:
											position, tokenIndex, depth = position347, tokenIndex347, depth347
											if buffer[position] != rune('A') {
												goto l241
											}
											position++
										}
									l347:
										{
											position349, tokenIndex349, depth349 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l350
											}
											position++
											goto l349
										l350:
											position, tokenIndex, depth = position349, tokenIndex349, depth349
											if buffer[position] != rune('S') {
												goto l241
											}
											position++
										}
									l349:
										break
									}
								}

							}
						l243:
							depth--
							add(ruleKEYWORD, position242)
						}
						goto l240
					l241:
						position, tokenIndex, depth = position241, tokenIndex241, depth241
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l240
					}
				l351:
					{
						position352, tokenIndex352, depth352 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l352
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l352
						}
						goto l351
					l352:
						position, tokenIndex, depth = position352, tokenIndex352, depth352
					}
					goto l239
				l240:
					position, tokenIndex, depth = position239, tokenIndex239, depth239
					if buffer[position] != rune('`') {
						goto l237
					}
					position++
				l353:
					{
						position354, tokenIndex354, depth354 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l354
						}
						goto l353
					l354:
						position, tokenIndex, depth = position354, tokenIndex354, depth354
					}
					if buffer[position] != rune('`') {
						goto l237
					}
					position++
				}
			l239:
				depth--
				add(ruleIDENTIFIER, position238)
			}
			return true
		l237:
			position, tokenIndex, depth = position237, tokenIndex237, depth237
			return false
		},
		/* 28 ID_SEGMENT <- <(ID_START ID_CONT*)> */
		func() bool {
			position355, tokenIndex355, depth355 := position, tokenIndex, depth
			{
				position356 := position
				depth++
				if !_rules[ruleID_START]() {
					goto l355
				}
			l357:
				{
					position358, tokenIndex358, depth358 := position, tokenIndex, depth
					{
						position359 := position
						depth++
						{
							position360, tokenIndex360, depth360 := position, tokenIndex, depth
							if !_rules[ruleID_START]() {
								goto l361
							}
							goto l360
						l361:
							position, tokenIndex, depth = position360, tokenIndex360, depth360
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l358
							}
							position++
						}
					l360:
						depth--
						add(ruleID_CONT, position359)
					}
					goto l357
				l358:
					position, tokenIndex, depth = position358, tokenIndex358, depth358
				}
				depth--
				add(ruleID_SEGMENT, position356)
			}
			return true
		l355:
			position, tokenIndex, depth = position355, tokenIndex355, depth355
			return false
		},
		/* 29 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position362, tokenIndex362, depth362 := position, tokenIndex, depth
			{
				position363 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l362
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l362
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l362
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position363)
			}
			return true
		l362:
			position, tokenIndex, depth = position362, tokenIndex362, depth362
			return false
		},
		/* 30 ID_CONT <- <(ID_START / [0-9])> */
		nil,
		/* 31 KEYWORD <- <((('a' / 'A') ('l' / 'L') ('l' / 'L')) / (('a' / 'A') ('n' / 'N') ('d' / 'D')) / ((&('W' | 'w') (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E'))) | (&('S' | 's') (('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T'))) | (&('O' | 'o') (('o' / 'O') ('r' / 'R'))) | (&('N' | 'n') (('n' / 'N') ('o' / 'O') ('t' / 'T'))) | (&('M' | 'm') (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S'))) | (&('I' | 'i') (('i' / 'I') ('n' / 'N'))) | (&('G' | 'g') (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P'))) | (&('F' | 'f') (('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))) | (&('D' | 'd') (('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E'))) | (&('B' | 'b') (('b' / 'B') ('y' / 'Y'))) | (&('A' | 'a') (('a' / 'A') ('s' / 'S')))))> */
		nil,
		/* 32 NUMBER <- <INTEGER> */
		nil,
		/* 33 INTEGER <- <('0' / ('-'? [1-9] [0-9]*))> */
		nil,
		/* 34 OP_ADD <- <(_ '+' _)> */
		nil,
		/* 35 OP_SUB <- <(_ '-' _)> */
		nil,
		/* 36 OP_MULT <- <(_ '*' _)> */
		nil,
		/* 37 OP_DIV <- <(_ '/' _)> */
		nil,
		/* 38 OP_AND <- <(__ (('a' / 'A') ('n' / 'N') ('d' / 'D')) __)> */
		nil,
		/* 39 OP_OR <- <(__ (('o' / 'O') ('r' / 'R')) __)> */
		nil,
		/* 40 OP_NOT <- <(('n' / 'N') ('o' / 'O') ('t' / 'T') __)> */
		nil,
		/* 41 STRING <- <(('\'' CHAR* '\'') / ('"' CHAR* '"'))> */
		func() bool {
			position376, tokenIndex376, depth376 := position, tokenIndex, depth
			{
				position377 := position
				depth++
				{
					position378, tokenIndex378, depth378 := position, tokenIndex, depth
					if buffer[position] != rune('\'') {
						goto l379
					}
					position++
				l380:
					{
						position381, tokenIndex381, depth381 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l381
						}
						goto l380
					l381:
						position, tokenIndex, depth = position381, tokenIndex381, depth381
					}
					if buffer[position] != rune('\'') {
						goto l379
					}
					position++
					goto l378
				l379:
					position, tokenIndex, depth = position378, tokenIndex378, depth378
					if buffer[position] != rune('"') {
						goto l376
					}
					position++
				l382:
					{
						position383, tokenIndex383, depth383 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l383
						}
						goto l382
					l383:
						position, tokenIndex, depth = position383, tokenIndex383, depth383
					}
					if buffer[position] != rune('"') {
						goto l376
					}
					position++
				}
			l378:
				depth--
				add(ruleSTRING, position377)
			}
			return true
		l376:
			position, tokenIndex, depth = position376, tokenIndex376, depth376
			return false
		},
		/* 42 CHAR <- <(('\\' ((&('\\') '\\') | (&('"') '"') | (&('`') '`') | (&('\'') '\''))) / (!'\'' .))> */
		func() bool {
			position384, tokenIndex384, depth384 := position, tokenIndex, depth
			{
				position385 := position
				depth++
				{
					position386, tokenIndex386, depth386 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l387
					}
					position++
					{
						switch buffer[position] {
						case '\\':
							if buffer[position] != rune('\\') {
								goto l387
							}
							position++
							break
						case '"':
							if buffer[position] != rune('"') {
								goto l387
							}
							position++
							break
						case '`':
							if buffer[position] != rune('`') {
								goto l387
							}
							position++
							break
						default:
							if buffer[position] != rune('\'') {
								goto l387
							}
							position++
							break
						}
					}

					goto l386
				l387:
					position, tokenIndex, depth = position386, tokenIndex386, depth386
					{
						position389, tokenIndex389, depth389 := position, tokenIndex, depth
						if buffer[position] != rune('\'') {
							goto l389
						}
						position++
						goto l384
					l389:
						position, tokenIndex, depth = position389, tokenIndex389, depth389
					}
					if !matchDot() {
						goto l384
					}
				}
			l386:
				depth--
				add(ruleCHAR, position385)
			}
			return true
		l384:
			position, tokenIndex, depth = position384, tokenIndex384, depth384
			return false
		},
		/* 43 PAREN_OPEN <- <(_ '(' _)> */
		func() bool {
			position390, tokenIndex390, depth390 := position, tokenIndex, depth
			{
				position391 := position
				depth++
				if !_rules[rule_]() {
					goto l390
				}
				if buffer[position] != rune('(') {
					goto l390
				}
				position++
				if !_rules[rule_]() {
					goto l390
				}
				depth--
				add(rulePAREN_OPEN, position391)
			}
			return true
		l390:
			position, tokenIndex, depth = position390, tokenIndex390, depth390
			return false
		},
		/* 44 PAREN_CLOSE <- <(_ ')' _)> */
		func() bool {
			position392, tokenIndex392, depth392 := position, tokenIndex, depth
			{
				position393 := position
				depth++
				if !_rules[rule_]() {
					goto l392
				}
				if buffer[position] != rune(')') {
					goto l392
				}
				position++
				if !_rules[rule_]() {
					goto l392
				}
				depth--
				add(rulePAREN_CLOSE, position393)
			}
			return true
		l392:
			position, tokenIndex, depth = position392, tokenIndex392, depth392
			return false
		},
		/* 45 COMMA <- <(_ ',' _)> */
		func() bool {
			position394, tokenIndex394, depth394 := position, tokenIndex, depth
			{
				position395 := position
				depth++
				if !_rules[rule_]() {
					goto l394
				}
				if buffer[position] != rune(',') {
					goto l394
				}
				position++
				if !_rules[rule_]() {
					goto l394
				}
				depth--
				add(ruleCOMMA, position395)
			}
			return true
		l394:
			position, tokenIndex, depth = position394, tokenIndex394, depth394
			return false
		},
		/* 46 COLON <- <(_ ':' _)> */
		nil,
		/* 47 _ <- <SPACE*> */
		func() bool {
			{
				position398 := position
				depth++
			l399:
				{
					position400, tokenIndex400, depth400 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l400
					}
					goto l399
				l400:
					position, tokenIndex, depth = position400, tokenIndex400, depth400
				}
				depth--
				add(rule_, position398)
			}
			return true
		},
		/* 48 __ <- <SPACE+> */
		func() bool {
			position401, tokenIndex401, depth401 := position, tokenIndex, depth
			{
				position402 := position
				depth++
				if !_rules[ruleSPACE]() {
					goto l401
				}
			l403:
				{
					position404, tokenIndex404, depth404 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l404
					}
					goto l403
				l404:
					position, tokenIndex, depth = position404, tokenIndex404, depth404
				}
				depth--
				add(rule__, position402)
			}
			return true
		l401:
			position, tokenIndex, depth = position401, tokenIndex401, depth401
			return false
		},
		/* 49 SPACE <- <((&('\t') '\t') | (&('\n') '\n') | (&(' ') ' '))> */
		func() bool {
			position405, tokenIndex405, depth405 := position, tokenIndex, depth
			{
				position406 := position
				depth++
				{
					switch buffer[position] {
					case '\t':
						if buffer[position] != rune('\t') {
							goto l405
						}
						position++
						break
					case '\n':
						if buffer[position] != rune('\n') {
							goto l405
						}
						position++
						break
					default:
						if buffer[position] != rune(' ') {
							goto l405
						}
						position++
						break
					}
				}

				depth--
				add(ruleSPACE, position406)
			}
			return true
		l405:
			position, tokenIndex, depth = position405, tokenIndex405, depth405
			return false
		},
		/* 51 Action0 <- <{ p.makeDescribeAll() }> */
		nil,
		nil,
		/* 53 Action1 <- <{ p.addLiteralNode(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 54 Action2 <- <{ p.makeDescribe() }> */
		nil,
		/* 55 Action3 <- <{ }> */
		nil,
		/* 56 Action4 <- <{ p.addNullPredicate() }> */
		nil,
		/* 57 Action5 <- <{ p.addAndMatcher() }> */
		nil,
		/* 58 Action6 <- <{ p.addOrMatcher() }> */
		nil,
		/* 59 Action7 <- <{ p.addNotMatcher() }> */
		nil,
		/* 60 Action8 <- <{
		   p.addLiteralMatcher()
		 }> */
		nil,
		/* 61 Action9 <- <{
		   p.addLiteralMatcher()
		   p.addNotMatcher()
		 }> */
		nil,
		/* 62 Action10 <- <{
		   p.addRegexMatcher()
		 }> */
		nil,
		/* 63 Action11 <- <{
		   p.addListMatcher()
		 }> */
		nil,
		/* 64 Action12 <- <{
		  p.addLiteralNode(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 65 Action13 <- <{ p.addLiteralListNode() }> */
		nil,
		/* 66 Action14 <- <{
		  p.appendLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 67 Action15 <- <{ p.addTagRefNode() }> */
		nil,
		/* 68 Action16 <- <{ p.setAlias(buffer[begin:end]) }> */
		nil,
		/* 69 Action17 <- <{ p.setTag(buffer[begin:end]) }> */
		nil,
	}
	p.rules = _rules
}
