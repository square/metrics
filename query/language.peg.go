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
	ruleoptionalPredicateClause
	ruleexpressionList
	ruleexpression_1
	ruleexpression_2
	ruleexpression_3
	rulegroupByClause
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
	"optionalPredicateClause",
	"expressionList",
	"expression_1",
	"expression_2",
	"expression_3",
	"groupByClause",
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
	rules  [65]func() bool
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
			p.addNullPredicate()
		case ruleAction4:
			p.addAndMatcher()
		case ruleAction5:
			p.addOrMatcher()
		case ruleAction6:
			p.addNotMatcher()
		case ruleAction7:

			p.addLiteralMatcher()

		case ruleAction8:

			p.addLiteralMatcher()
			p.addNotMatcher()

		case ruleAction9:

			p.addRegexMatcher()

		case ruleAction10:

			p.addListMatcher()

		case ruleAction11:

			p.addLiteralNode(unescapeLiteral(buffer[begin:end]))

		case ruleAction12:
			p.addLiteralListNode()
		case ruleAction13:

			p.appendLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction14:
			p.addTagRefNode()
		case ruleAction15:
			p.setAlias(buffer[begin:end])
		case ruleAction16:
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
						if !_rules[ruleoptionalPredicateClause]() {
							goto l3
						}
						depth--
						add(ruleselectStmt, position4)
					}
					goto l2
				l3:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
					{
						position17 := position
						depth++
						{
							position18, tokenIndex18, depth18 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l19
							}
							position++
							goto l18
						l19:
							position, tokenIndex, depth = position18, tokenIndex18, depth18
							if buffer[position] != rune('D') {
								goto l0
							}
							position++
						}
					l18:
						{
							position20, tokenIndex20, depth20 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l21
							}
							position++
							goto l20
						l21:
							position, tokenIndex, depth = position20, tokenIndex20, depth20
							if buffer[position] != rune('E') {
								goto l0
							}
							position++
						}
					l20:
						{
							position22, tokenIndex22, depth22 := position, tokenIndex, depth
							if buffer[position] != rune('s') {
								goto l23
							}
							position++
							goto l22
						l23:
							position, tokenIndex, depth = position22, tokenIndex22, depth22
							if buffer[position] != rune('S') {
								goto l0
							}
							position++
						}
					l22:
						{
							position24, tokenIndex24, depth24 := position, tokenIndex, depth
							if buffer[position] != rune('c') {
								goto l25
							}
							position++
							goto l24
						l25:
							position, tokenIndex, depth = position24, tokenIndex24, depth24
							if buffer[position] != rune('C') {
								goto l0
							}
							position++
						}
					l24:
						{
							position26, tokenIndex26, depth26 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l27
							}
							position++
							goto l26
						l27:
							position, tokenIndex, depth = position26, tokenIndex26, depth26
							if buffer[position] != rune('R') {
								goto l0
							}
							position++
						}
					l26:
						{
							position28, tokenIndex28, depth28 := position, tokenIndex, depth
							if buffer[position] != rune('i') {
								goto l29
							}
							position++
							goto l28
						l29:
							position, tokenIndex, depth = position28, tokenIndex28, depth28
							if buffer[position] != rune('I') {
								goto l0
							}
							position++
						}
					l28:
						{
							position30, tokenIndex30, depth30 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l31
							}
							position++
							goto l30
						l31:
							position, tokenIndex, depth = position30, tokenIndex30, depth30
							if buffer[position] != rune('B') {
								goto l0
							}
							position++
						}
					l30:
						{
							position32, tokenIndex32, depth32 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l33
							}
							position++
							goto l32
						l33:
							position, tokenIndex, depth = position32, tokenIndex32, depth32
							if buffer[position] != rune('E') {
								goto l0
							}
							position++
						}
					l32:
						if !_rules[rule__]() {
							goto l0
						}
						{
							position34, tokenIndex34, depth34 := position, tokenIndex, depth
							{
								position36 := position
								depth++
								{
									position37, tokenIndex37, depth37 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l38
									}
									position++
									goto l37
								l38:
									position, tokenIndex, depth = position37, tokenIndex37, depth37
									if buffer[position] != rune('A') {
										goto l35
									}
									position++
								}
							l37:
								{
									position39, tokenIndex39, depth39 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l40
									}
									position++
									goto l39
								l40:
									position, tokenIndex, depth = position39, tokenIndex39, depth39
									if buffer[position] != rune('L') {
										goto l35
									}
									position++
								}
							l39:
								{
									position41, tokenIndex41, depth41 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l42
									}
									position++
									goto l41
								l42:
									position, tokenIndex, depth = position41, tokenIndex41, depth41
									if buffer[position] != rune('L') {
										goto l35
									}
									position++
								}
							l41:
								{
									add(ruleAction0, position)
								}
								depth--
								add(ruledescribeAllStmt, position36)
							}
							goto l34
						l35:
							position, tokenIndex, depth = position34, tokenIndex34, depth34
							{
								position44 := position
								depth++
								{
									position45 := position
									depth++
									{
										position46 := position
										depth++
										if !_rules[ruleIDENTIFIER]() {
											goto l0
										}
										depth--
										add(ruleMETRIC_NAME, position46)
									}
									depth--
									add(rulePegText, position45)
								}
								{
									add(ruleAction1, position)
								}
								if !_rules[ruleoptionalPredicateClause]() {
									goto l0
								}
								{
									add(ruleAction2, position)
								}
								depth--
								add(ruledescribeSingleStmt, position44)
							}
						}
					l34:
						depth--
						add(ruledescribeStmt, position17)
					}
				}
			l2:
				{
					position49, tokenIndex49, depth49 := position, tokenIndex, depth
					if !matchDot() {
						goto l49
					}
					goto l0
				l49:
					position, tokenIndex, depth = position49, tokenIndex49, depth49
				}
				depth--
				add(ruleroot, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 selectStmt <- <(('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T') __ expressionList optionalPredicateClause)> */
		nil,
		/* 2 describeStmt <- <(('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E') __ (describeAllStmt / describeSingleStmt))> */
		nil,
		/* 3 describeAllStmt <- <(('a' / 'A') ('l' / 'L') ('l' / 'L') Action0)> */
		nil,
		/* 4 describeSingleStmt <- <(<METRIC_NAME> Action1 optionalPredicateClause Action2)> */
		nil,
		/* 5 optionalPredicateClause <- <((__ predicateClause) / Action3)> */
		func() bool {
			{
				position55 := position
				depth++
				{
					position56, tokenIndex56, depth56 := position, tokenIndex, depth
					if !_rules[rule__]() {
						goto l57
					}
					{
						position58 := position
						depth++
						{
							position59, tokenIndex59, depth59 := position, tokenIndex, depth
							if buffer[position] != rune('w') {
								goto l60
							}
							position++
							goto l59
						l60:
							position, tokenIndex, depth = position59, tokenIndex59, depth59
							if buffer[position] != rune('W') {
								goto l57
							}
							position++
						}
					l59:
						{
							position61, tokenIndex61, depth61 := position, tokenIndex, depth
							if buffer[position] != rune('h') {
								goto l62
							}
							position++
							goto l61
						l62:
							position, tokenIndex, depth = position61, tokenIndex61, depth61
							if buffer[position] != rune('H') {
								goto l57
							}
							position++
						}
					l61:
						{
							position63, tokenIndex63, depth63 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l64
							}
							position++
							goto l63
						l64:
							position, tokenIndex, depth = position63, tokenIndex63, depth63
							if buffer[position] != rune('E') {
								goto l57
							}
							position++
						}
					l63:
						{
							position65, tokenIndex65, depth65 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l66
							}
							position++
							goto l65
						l66:
							position, tokenIndex, depth = position65, tokenIndex65, depth65
							if buffer[position] != rune('R') {
								goto l57
							}
							position++
						}
					l65:
						{
							position67, tokenIndex67, depth67 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l68
							}
							position++
							goto l67
						l68:
							position, tokenIndex, depth = position67, tokenIndex67, depth67
							if buffer[position] != rune('E') {
								goto l57
							}
							position++
						}
					l67:
						if !_rules[rule__]() {
							goto l57
						}
						if !_rules[rulepredicate_1]() {
							goto l57
						}
						depth--
						add(rulepredicateClause, position58)
					}
					goto l56
				l57:
					position, tokenIndex, depth = position56, tokenIndex56, depth56
					{
						add(ruleAction3, position)
					}
				}
			l56:
				depth--
				add(ruleoptionalPredicateClause, position55)
			}
			return true
		},
		/* 6 expressionList <- <(expression_1 (COMMA expression_1)*)> */
		func() bool {
			position70, tokenIndex70, depth70 := position, tokenIndex, depth
			{
				position71 := position
				depth++
				if !_rules[ruleexpression_1]() {
					goto l70
				}
			l72:
				{
					position73, tokenIndex73, depth73 := position, tokenIndex, depth
					if !_rules[ruleCOMMA]() {
						goto l73
					}
					if !_rules[ruleexpression_1]() {
						goto l73
					}
					goto l72
				l73:
					position, tokenIndex, depth = position73, tokenIndex73, depth73
				}
				depth--
				add(ruleexpressionList, position71)
			}
			return true
		l70:
			position, tokenIndex, depth = position70, tokenIndex70, depth70
			return false
		},
		/* 7 expression_1 <- <(expression_2 ((OP_ADD / OP_SUB) expression_2)*)> */
		func() bool {
			position74, tokenIndex74, depth74 := position, tokenIndex, depth
			{
				position75 := position
				depth++
				if !_rules[ruleexpression_2]() {
					goto l74
				}
			l76:
				{
					position77, tokenIndex77, depth77 := position, tokenIndex, depth
					{
						position78, tokenIndex78, depth78 := position, tokenIndex, depth
						{
							position80 := position
							depth++
							if !_rules[rule_]() {
								goto l79
							}
							if buffer[position] != rune('+') {
								goto l79
							}
							position++
							if !_rules[rule_]() {
								goto l79
							}
							depth--
							add(ruleOP_ADD, position80)
						}
						goto l78
					l79:
						position, tokenIndex, depth = position78, tokenIndex78, depth78
						{
							position81 := position
							depth++
							if !_rules[rule_]() {
								goto l77
							}
							if buffer[position] != rune('-') {
								goto l77
							}
							position++
							if !_rules[rule_]() {
								goto l77
							}
							depth--
							add(ruleOP_SUB, position81)
						}
					}
				l78:
					if !_rules[ruleexpression_2]() {
						goto l77
					}
					goto l76
				l77:
					position, tokenIndex, depth = position77, tokenIndex77, depth77
				}
				depth--
				add(ruleexpression_1, position75)
			}
			return true
		l74:
			position, tokenIndex, depth = position74, tokenIndex74, depth74
			return false
		},
		/* 8 expression_2 <- <(expression_3 ((OP_DIV / OP_MULT) expression_3)*)> */
		func() bool {
			position82, tokenIndex82, depth82 := position, tokenIndex, depth
			{
				position83 := position
				depth++
				if !_rules[ruleexpression_3]() {
					goto l82
				}
			l84:
				{
					position85, tokenIndex85, depth85 := position, tokenIndex, depth
					{
						position86, tokenIndex86, depth86 := position, tokenIndex, depth
						{
							position88 := position
							depth++
							if !_rules[rule_]() {
								goto l87
							}
							if buffer[position] != rune('/') {
								goto l87
							}
							position++
							if !_rules[rule_]() {
								goto l87
							}
							depth--
							add(ruleOP_DIV, position88)
						}
						goto l86
					l87:
						position, tokenIndex, depth = position86, tokenIndex86, depth86
						{
							position89 := position
							depth++
							if !_rules[rule_]() {
								goto l85
							}
							if buffer[position] != rune('*') {
								goto l85
							}
							position++
							if !_rules[rule_]() {
								goto l85
							}
							depth--
							add(ruleOP_MULT, position89)
						}
					}
				l86:
					if !_rules[ruleexpression_3]() {
						goto l85
					}
					goto l84
				l85:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
				}
				depth--
				add(ruleexpression_2, position83)
			}
			return true
		l82:
			position, tokenIndex, depth = position82, tokenIndex82, depth82
			return false
		},
		/* 9 expression_3 <- <((IDENTIFIER PAREN_OPEN expression_1 __ groupByClause PAREN_CLOSE) / (IDENTIFIER PAREN_OPEN expressionList PAREN_CLOSE) / ((&('-' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') NUMBER) | (&('\t' | '\n' | ' ' | '(') (PAREN_OPEN expression_1 PAREN_CLOSE)) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') (IDENTIFIER ('[' _ predicate_1 _ ']')?))))> */
		func() bool {
			position90, tokenIndex90, depth90 := position, tokenIndex, depth
			{
				position91 := position
				depth++
				{
					position92, tokenIndex92, depth92 := position, tokenIndex, depth
					if !_rules[ruleIDENTIFIER]() {
						goto l93
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l93
					}
					if !_rules[ruleexpression_1]() {
						goto l93
					}
					if !_rules[rule__]() {
						goto l93
					}
					{
						position94 := position
						depth++
						{
							position95, tokenIndex95, depth95 := position, tokenIndex, depth
							if buffer[position] != rune('g') {
								goto l96
							}
							position++
							goto l95
						l96:
							position, tokenIndex, depth = position95, tokenIndex95, depth95
							if buffer[position] != rune('G') {
								goto l93
							}
							position++
						}
					l95:
						{
							position97, tokenIndex97, depth97 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l98
							}
							position++
							goto l97
						l98:
							position, tokenIndex, depth = position97, tokenIndex97, depth97
							if buffer[position] != rune('R') {
								goto l93
							}
							position++
						}
					l97:
						{
							position99, tokenIndex99, depth99 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l100
							}
							position++
							goto l99
						l100:
							position, tokenIndex, depth = position99, tokenIndex99, depth99
							if buffer[position] != rune('O') {
								goto l93
							}
							position++
						}
					l99:
						{
							position101, tokenIndex101, depth101 := position, tokenIndex, depth
							if buffer[position] != rune('u') {
								goto l102
							}
							position++
							goto l101
						l102:
							position, tokenIndex, depth = position101, tokenIndex101, depth101
							if buffer[position] != rune('U') {
								goto l93
							}
							position++
						}
					l101:
						{
							position103, tokenIndex103, depth103 := position, tokenIndex, depth
							if buffer[position] != rune('p') {
								goto l104
							}
							position++
							goto l103
						l104:
							position, tokenIndex, depth = position103, tokenIndex103, depth103
							if buffer[position] != rune('P') {
								goto l93
							}
							position++
						}
					l103:
						if !_rules[rule__]() {
							goto l93
						}
						{
							position105, tokenIndex105, depth105 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l106
							}
							position++
							goto l105
						l106:
							position, tokenIndex, depth = position105, tokenIndex105, depth105
							if buffer[position] != rune('B') {
								goto l93
							}
							position++
						}
					l105:
						{
							position107, tokenIndex107, depth107 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l108
							}
							position++
							goto l107
						l108:
							position, tokenIndex, depth = position107, tokenIndex107, depth107
							if buffer[position] != rune('Y') {
								goto l93
							}
							position++
						}
					l107:
						if !_rules[rule__]() {
							goto l93
						}
						if !_rules[ruleCOLUMN_NAME]() {
							goto l93
						}
					l109:
						{
							position110, tokenIndex110, depth110 := position, tokenIndex, depth
							if !_rules[ruleCOMMA]() {
								goto l110
							}
							if !_rules[ruleCOLUMN_NAME]() {
								goto l110
							}
							goto l109
						l110:
							position, tokenIndex, depth = position110, tokenIndex110, depth110
						}
						depth--
						add(rulegroupByClause, position94)
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l93
					}
					goto l92
				l93:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if !_rules[ruleIDENTIFIER]() {
						goto l111
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l111
					}
					if !_rules[ruleexpressionList]() {
						goto l111
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l111
					}
					goto l92
				l111:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					{
						switch buffer[position] {
						case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							{
								position113 := position
								depth++
								{
									position114 := position
									depth++
									{
										position115, tokenIndex115, depth115 := position, tokenIndex, depth
										if buffer[position] != rune('0') {
											goto l116
										}
										position++
										goto l115
									l116:
										position, tokenIndex, depth = position115, tokenIndex115, depth115
										{
											position117, tokenIndex117, depth117 := position, tokenIndex, depth
											if buffer[position] != rune('-') {
												goto l117
											}
											position++
											goto l118
										l117:
											position, tokenIndex, depth = position117, tokenIndex117, depth117
										}
									l118:
										if c := buffer[position]; c < rune('1') || c > rune('9') {
											goto l90
										}
										position++
									l119:
										{
											position120, tokenIndex120, depth120 := position, tokenIndex, depth
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l120
											}
											position++
											goto l119
										l120:
											position, tokenIndex, depth = position120, tokenIndex120, depth120
										}
									}
								l115:
									depth--
									add(ruleINTEGER, position114)
								}
								depth--
								add(ruleNUMBER, position113)
							}
							break
						case '\t', '\n', ' ', '(':
							if !_rules[rulePAREN_OPEN]() {
								goto l90
							}
							if !_rules[ruleexpression_1]() {
								goto l90
							}
							if !_rules[rulePAREN_CLOSE]() {
								goto l90
							}
							break
						default:
							if !_rules[ruleIDENTIFIER]() {
								goto l90
							}
							{
								position121, tokenIndex121, depth121 := position, tokenIndex, depth
								if buffer[position] != rune('[') {
									goto l121
								}
								position++
								if !_rules[rule_]() {
									goto l121
								}
								if !_rules[rulepredicate_1]() {
									goto l121
								}
								if !_rules[rule_]() {
									goto l121
								}
								if buffer[position] != rune(']') {
									goto l121
								}
								position++
								goto l122
							l121:
								position, tokenIndex, depth = position121, tokenIndex121, depth121
							}
						l122:
							break
						}
					}

				}
			l92:
				depth--
				add(ruleexpression_3, position91)
			}
			return true
		l90:
			position, tokenIndex, depth = position90, tokenIndex90, depth90
			return false
		},
		/* 10 groupByClause <- <(('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P') __ (('b' / 'B') ('y' / 'Y')) __ COLUMN_NAME (COMMA COLUMN_NAME)*)> */
		nil,
		/* 11 predicateClause <- <(('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E') __ predicate_1)> */
		nil,
		/* 12 predicate_1 <- <((predicate_2 OP_AND predicate_1 Action4) / predicate_2 / )> */
		func() bool {
			{
				position126 := position
				depth++
				{
					position127, tokenIndex127, depth127 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l128
					}
					{
						position129 := position
						depth++
						if !_rules[rule__]() {
							goto l128
						}
						{
							position130, tokenIndex130, depth130 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l131
							}
							position++
							goto l130
						l131:
							position, tokenIndex, depth = position130, tokenIndex130, depth130
							if buffer[position] != rune('A') {
								goto l128
							}
							position++
						}
					l130:
						{
							position132, tokenIndex132, depth132 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l133
							}
							position++
							goto l132
						l133:
							position, tokenIndex, depth = position132, tokenIndex132, depth132
							if buffer[position] != rune('N') {
								goto l128
							}
							position++
						}
					l132:
						{
							position134, tokenIndex134, depth134 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l135
							}
							position++
							goto l134
						l135:
							position, tokenIndex, depth = position134, tokenIndex134, depth134
							if buffer[position] != rune('D') {
								goto l128
							}
							position++
						}
					l134:
						if !_rules[rule__]() {
							goto l128
						}
						depth--
						add(ruleOP_AND, position129)
					}
					if !_rules[rulepredicate_1]() {
						goto l128
					}
					{
						add(ruleAction4, position)
					}
					goto l127
				l128:
					position, tokenIndex, depth = position127, tokenIndex127, depth127
					if !_rules[rulepredicate_2]() {
						goto l137
					}
					goto l127
				l137:
					position, tokenIndex, depth = position127, tokenIndex127, depth127
				}
			l127:
				depth--
				add(rulepredicate_1, position126)
			}
			return true
		},
		/* 13 predicate_2 <- <((predicate_3 OP_OR predicate_2 Action5) / predicate_3)> */
		func() bool {
			position138, tokenIndex138, depth138 := position, tokenIndex, depth
			{
				position139 := position
				depth++
				{
					position140, tokenIndex140, depth140 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l141
					}
					{
						position142 := position
						depth++
						if !_rules[rule__]() {
							goto l141
						}
						{
							position143, tokenIndex143, depth143 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l144
							}
							position++
							goto l143
						l144:
							position, tokenIndex, depth = position143, tokenIndex143, depth143
							if buffer[position] != rune('O') {
								goto l141
							}
							position++
						}
					l143:
						{
							position145, tokenIndex145, depth145 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l146
							}
							position++
							goto l145
						l146:
							position, tokenIndex, depth = position145, tokenIndex145, depth145
							if buffer[position] != rune('R') {
								goto l141
							}
							position++
						}
					l145:
						if !_rules[rule__]() {
							goto l141
						}
						depth--
						add(ruleOP_OR, position142)
					}
					if !_rules[rulepredicate_2]() {
						goto l141
					}
					{
						add(ruleAction5, position)
					}
					goto l140
				l141:
					position, tokenIndex, depth = position140, tokenIndex140, depth140
					if !_rules[rulepredicate_3]() {
						goto l138
					}
				}
			l140:
				depth--
				add(rulepredicate_2, position139)
			}
			return true
		l138:
			position, tokenIndex, depth = position138, tokenIndex138, depth138
			return false
		},
		/* 14 predicate_3 <- <((OP_NOT predicate_3 Action6) / (PAREN_OPEN predicate_1 PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position148, tokenIndex148, depth148 := position, tokenIndex, depth
			{
				position149 := position
				depth++
				{
					position150, tokenIndex150, depth150 := position, tokenIndex, depth
					{
						position152 := position
						depth++
						{
							position153, tokenIndex153, depth153 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l154
							}
							position++
							goto l153
						l154:
							position, tokenIndex, depth = position153, tokenIndex153, depth153
							if buffer[position] != rune('N') {
								goto l151
							}
							position++
						}
					l153:
						{
							position155, tokenIndex155, depth155 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l156
							}
							position++
							goto l155
						l156:
							position, tokenIndex, depth = position155, tokenIndex155, depth155
							if buffer[position] != rune('O') {
								goto l151
							}
							position++
						}
					l155:
						{
							position157, tokenIndex157, depth157 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l158
							}
							position++
							goto l157
						l158:
							position, tokenIndex, depth = position157, tokenIndex157, depth157
							if buffer[position] != rune('T') {
								goto l151
							}
							position++
						}
					l157:
						if !_rules[rule__]() {
							goto l151
						}
						depth--
						add(ruleOP_NOT, position152)
					}
					if !_rules[rulepredicate_3]() {
						goto l151
					}
					{
						add(ruleAction6, position)
					}
					goto l150
				l151:
					position, tokenIndex, depth = position150, tokenIndex150, depth150
					if !_rules[rulePAREN_OPEN]() {
						goto l160
					}
					if !_rules[rulepredicate_1]() {
						goto l160
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l160
					}
					goto l150
				l160:
					position, tokenIndex, depth = position150, tokenIndex150, depth150
					{
						position161 := position
						depth++
						{
							position162, tokenIndex162, depth162 := position, tokenIndex, depth
							if !_rules[rulepropertySource]() {
								goto l163
							}
							if !_rules[rule_]() {
								goto l163
							}
							if buffer[position] != rune('=') {
								goto l163
							}
							position++
							if !_rules[rule_]() {
								goto l163
							}
							if !_rules[ruleliteralString]() {
								goto l163
							}
							{
								add(ruleAction7, position)
							}
							goto l162
						l163:
							position, tokenIndex, depth = position162, tokenIndex162, depth162
							if !_rules[rulepropertySource]() {
								goto l165
							}
							if !_rules[rule_]() {
								goto l165
							}
							if buffer[position] != rune('!') {
								goto l165
							}
							position++
							if buffer[position] != rune('=') {
								goto l165
							}
							position++
							if !_rules[rule_]() {
								goto l165
							}
							if !_rules[ruleliteralString]() {
								goto l165
							}
							{
								add(ruleAction8, position)
							}
							goto l162
						l165:
							position, tokenIndex, depth = position162, tokenIndex162, depth162
							if !_rules[rulepropertySource]() {
								goto l167
							}
							if !_rules[rule__]() {
								goto l167
							}
							{
								position168, tokenIndex168, depth168 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l169
								}
								position++
								goto l168
							l169:
								position, tokenIndex, depth = position168, tokenIndex168, depth168
								if buffer[position] != rune('M') {
									goto l167
								}
								position++
							}
						l168:
							{
								position170, tokenIndex170, depth170 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l171
								}
								position++
								goto l170
							l171:
								position, tokenIndex, depth = position170, tokenIndex170, depth170
								if buffer[position] != rune('A') {
									goto l167
								}
								position++
							}
						l170:
							{
								position172, tokenIndex172, depth172 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l173
								}
								position++
								goto l172
							l173:
								position, tokenIndex, depth = position172, tokenIndex172, depth172
								if buffer[position] != rune('T') {
									goto l167
								}
								position++
							}
						l172:
							{
								position174, tokenIndex174, depth174 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l175
								}
								position++
								goto l174
							l175:
								position, tokenIndex, depth = position174, tokenIndex174, depth174
								if buffer[position] != rune('C') {
									goto l167
								}
								position++
							}
						l174:
							{
								position176, tokenIndex176, depth176 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l177
								}
								position++
								goto l176
							l177:
								position, tokenIndex, depth = position176, tokenIndex176, depth176
								if buffer[position] != rune('H') {
									goto l167
								}
								position++
							}
						l176:
							{
								position178, tokenIndex178, depth178 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l179
								}
								position++
								goto l178
							l179:
								position, tokenIndex, depth = position178, tokenIndex178, depth178
								if buffer[position] != rune('E') {
									goto l167
								}
								position++
							}
						l178:
							{
								position180, tokenIndex180, depth180 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l181
								}
								position++
								goto l180
							l181:
								position, tokenIndex, depth = position180, tokenIndex180, depth180
								if buffer[position] != rune('S') {
									goto l167
								}
								position++
							}
						l180:
							if !_rules[rule__]() {
								goto l167
							}
							if !_rules[ruleliteralString]() {
								goto l167
							}
							{
								add(ruleAction9, position)
							}
							goto l162
						l167:
							position, tokenIndex, depth = position162, tokenIndex162, depth162
							if !_rules[rulepropertySource]() {
								goto l148
							}
							if !_rules[rule__]() {
								goto l148
							}
							{
								position183, tokenIndex183, depth183 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l184
								}
								position++
								goto l183
							l184:
								position, tokenIndex, depth = position183, tokenIndex183, depth183
								if buffer[position] != rune('I') {
									goto l148
								}
								position++
							}
						l183:
							{
								position185, tokenIndex185, depth185 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l186
								}
								position++
								goto l185
							l186:
								position, tokenIndex, depth = position185, tokenIndex185, depth185
								if buffer[position] != rune('N') {
									goto l148
								}
								position++
							}
						l185:
							if !_rules[rule__]() {
								goto l148
							}
							{
								position187 := position
								depth++
								{
									add(ruleAction12, position)
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l148
								}
								if !_rules[ruleliteralListString]() {
									goto l148
								}
							l189:
								{
									position190, tokenIndex190, depth190 := position, tokenIndex, depth
									if !_rules[ruleCOMMA]() {
										goto l190
									}
									if !_rules[ruleliteralListString]() {
										goto l190
									}
									goto l189
								l190:
									position, tokenIndex, depth = position190, tokenIndex190, depth190
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l148
								}
								depth--
								add(ruleliteralList, position187)
							}
							{
								add(ruleAction10, position)
							}
						}
					l162:
						depth--
						add(ruletagMatcher, position161)
					}
				}
			l150:
				depth--
				add(rulepredicate_3, position149)
			}
			return true
		l148:
			position, tokenIndex, depth = position148, tokenIndex148, depth148
			return false
		},
		/* 15 tagMatcher <- <((propertySource _ '=' _ literalString Action7) / (propertySource _ ('!' '=') _ literalString Action8) / (propertySource __ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) __ literalString Action9) / (propertySource __ (('i' / 'I') ('n' / 'N')) __ literalList Action10))> */
		nil,
		/* 16 literalString <- <(<STRING> Action11)> */
		func() bool {
			position193, tokenIndex193, depth193 := position, tokenIndex, depth
			{
				position194 := position
				depth++
				{
					position195 := position
					depth++
					if !_rules[ruleSTRING]() {
						goto l193
					}
					depth--
					add(rulePegText, position195)
				}
				{
					add(ruleAction11, position)
				}
				depth--
				add(ruleliteralString, position194)
			}
			return true
		l193:
			position, tokenIndex, depth = position193, tokenIndex193, depth193
			return false
		},
		/* 17 literalList <- <(Action12 PAREN_OPEN literalListString (COMMA literalListString)* PAREN_CLOSE)> */
		nil,
		/* 18 literalListString <- <(STRING Action13)> */
		func() bool {
			position198, tokenIndex198, depth198 := position, tokenIndex, depth
			{
				position199 := position
				depth++
				if !_rules[ruleSTRING]() {
					goto l198
				}
				{
					add(ruleAction13, position)
				}
				depth--
				add(ruleliteralListString, position199)
			}
			return true
		l198:
			position, tokenIndex, depth = position198, tokenIndex198, depth198
			return false
		},
		/* 19 propertySource <- <(Action14 (<IDENTIFIER> Action15 COLON)? <TAG_NAME> Action16)> */
		func() bool {
			position201, tokenIndex201, depth201 := position, tokenIndex, depth
			{
				position202 := position
				depth++
				{
					add(ruleAction14, position)
				}
				{
					position204, tokenIndex204, depth204 := position, tokenIndex, depth
					{
						position206 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l204
						}
						depth--
						add(rulePegText, position206)
					}
					{
						add(ruleAction15, position)
					}
					{
						position208 := position
						depth++
						if !_rules[rule_]() {
							goto l204
						}
						if buffer[position] != rune(':') {
							goto l204
						}
						position++
						if !_rules[rule_]() {
							goto l204
						}
						depth--
						add(ruleCOLON, position208)
					}
					goto l205
				l204:
					position, tokenIndex, depth = position204, tokenIndex204, depth204
				}
			l205:
				{
					position209 := position
					depth++
					{
						position210 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l201
						}
						depth--
						add(ruleTAG_NAME, position210)
					}
					depth--
					add(rulePegText, position209)
				}
				{
					add(ruleAction16, position)
				}
				depth--
				add(rulepropertySource, position202)
			}
			return true
		l201:
			position, tokenIndex, depth = position201, tokenIndex201, depth201
			return false
		},
		/* 20 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position212, tokenIndex212, depth212 := position, tokenIndex, depth
			{
				position213 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l212
				}
				depth--
				add(ruleCOLUMN_NAME, position213)
			}
			return true
		l212:
			position, tokenIndex, depth = position212, tokenIndex212, depth212
			return false
		},
		/* 21 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 22 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 23 IDENTIFIER <- <((!KEYWORD ID_SEGMENT ('.' ID_SEGMENT)*) / ('`' CHAR* '`'))> */
		func() bool {
			position216, tokenIndex216, depth216 := position, tokenIndex, depth
			{
				position217 := position
				depth++
				{
					position218, tokenIndex218, depth218 := position, tokenIndex, depth
					{
						position220, tokenIndex220, depth220 := position, tokenIndex, depth
						{
							position221 := position
							depth++
							{
								position222, tokenIndex222, depth222 := position, tokenIndex, depth
								{
									position224, tokenIndex224, depth224 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l225
									}
									position++
									goto l224
								l225:
									position, tokenIndex, depth = position224, tokenIndex224, depth224
									if buffer[position] != rune('A') {
										goto l223
									}
									position++
								}
							l224:
								{
									position226, tokenIndex226, depth226 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l227
									}
									position++
									goto l226
								l227:
									position, tokenIndex, depth = position226, tokenIndex226, depth226
									if buffer[position] != rune('L') {
										goto l223
									}
									position++
								}
							l226:
								{
									position228, tokenIndex228, depth228 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l229
									}
									position++
									goto l228
								l229:
									position, tokenIndex, depth = position228, tokenIndex228, depth228
									if buffer[position] != rune('L') {
										goto l223
									}
									position++
								}
							l228:
								goto l222
							l223:
								position, tokenIndex, depth = position222, tokenIndex222, depth222
								{
									position231, tokenIndex231, depth231 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l232
									}
									position++
									goto l231
								l232:
									position, tokenIndex, depth = position231, tokenIndex231, depth231
									if buffer[position] != rune('A') {
										goto l230
									}
									position++
								}
							l231:
								{
									position233, tokenIndex233, depth233 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l234
									}
									position++
									goto l233
								l234:
									position, tokenIndex, depth = position233, tokenIndex233, depth233
									if buffer[position] != rune('N') {
										goto l230
									}
									position++
								}
							l233:
								{
									position235, tokenIndex235, depth235 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l236
									}
									position++
									goto l235
								l236:
									position, tokenIndex, depth = position235, tokenIndex235, depth235
									if buffer[position] != rune('D') {
										goto l230
									}
									position++
								}
							l235:
								goto l222
							l230:
								position, tokenIndex, depth = position222, tokenIndex222, depth222
								{
									switch buffer[position] {
									case 'W', 'w':
										{
											position238, tokenIndex238, depth238 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l239
											}
											position++
											goto l238
										l239:
											position, tokenIndex, depth = position238, tokenIndex238, depth238
											if buffer[position] != rune('W') {
												goto l220
											}
											position++
										}
									l238:
										{
											position240, tokenIndex240, depth240 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l241
											}
											position++
											goto l240
										l241:
											position, tokenIndex, depth = position240, tokenIndex240, depth240
											if buffer[position] != rune('H') {
												goto l220
											}
											position++
										}
									l240:
										{
											position242, tokenIndex242, depth242 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l243
											}
											position++
											goto l242
										l243:
											position, tokenIndex, depth = position242, tokenIndex242, depth242
											if buffer[position] != rune('E') {
												goto l220
											}
											position++
										}
									l242:
										{
											position244, tokenIndex244, depth244 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l245
											}
											position++
											goto l244
										l245:
											position, tokenIndex, depth = position244, tokenIndex244, depth244
											if buffer[position] != rune('R') {
												goto l220
											}
											position++
										}
									l244:
										{
											position246, tokenIndex246, depth246 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l247
											}
											position++
											goto l246
										l247:
											position, tokenIndex, depth = position246, tokenIndex246, depth246
											if buffer[position] != rune('E') {
												goto l220
											}
											position++
										}
									l246:
										break
									case 'S', 's':
										{
											position248, tokenIndex248, depth248 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l249
											}
											position++
											goto l248
										l249:
											position, tokenIndex, depth = position248, tokenIndex248, depth248
											if buffer[position] != rune('S') {
												goto l220
											}
											position++
										}
									l248:
										{
											position250, tokenIndex250, depth250 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l251
											}
											position++
											goto l250
										l251:
											position, tokenIndex, depth = position250, tokenIndex250, depth250
											if buffer[position] != rune('E') {
												goto l220
											}
											position++
										}
									l250:
										{
											position252, tokenIndex252, depth252 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l253
											}
											position++
											goto l252
										l253:
											position, tokenIndex, depth = position252, tokenIndex252, depth252
											if buffer[position] != rune('L') {
												goto l220
											}
											position++
										}
									l252:
										{
											position254, tokenIndex254, depth254 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l255
											}
											position++
											goto l254
										l255:
											position, tokenIndex, depth = position254, tokenIndex254, depth254
											if buffer[position] != rune('E') {
												goto l220
											}
											position++
										}
									l254:
										{
											position256, tokenIndex256, depth256 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l257
											}
											position++
											goto l256
										l257:
											position, tokenIndex, depth = position256, tokenIndex256, depth256
											if buffer[position] != rune('C') {
												goto l220
											}
											position++
										}
									l256:
										{
											position258, tokenIndex258, depth258 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l259
											}
											position++
											goto l258
										l259:
											position, tokenIndex, depth = position258, tokenIndex258, depth258
											if buffer[position] != rune('T') {
												goto l220
											}
											position++
										}
									l258:
										break
									case 'O', 'o':
										{
											position260, tokenIndex260, depth260 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l261
											}
											position++
											goto l260
										l261:
											position, tokenIndex, depth = position260, tokenIndex260, depth260
											if buffer[position] != rune('O') {
												goto l220
											}
											position++
										}
									l260:
										{
											position262, tokenIndex262, depth262 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l263
											}
											position++
											goto l262
										l263:
											position, tokenIndex, depth = position262, tokenIndex262, depth262
											if buffer[position] != rune('R') {
												goto l220
											}
											position++
										}
									l262:
										break
									case 'N', 'n':
										{
											position264, tokenIndex264, depth264 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l265
											}
											position++
											goto l264
										l265:
											position, tokenIndex, depth = position264, tokenIndex264, depth264
											if buffer[position] != rune('N') {
												goto l220
											}
											position++
										}
									l264:
										{
											position266, tokenIndex266, depth266 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l267
											}
											position++
											goto l266
										l267:
											position, tokenIndex, depth = position266, tokenIndex266, depth266
											if buffer[position] != rune('O') {
												goto l220
											}
											position++
										}
									l266:
										{
											position268, tokenIndex268, depth268 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l269
											}
											position++
											goto l268
										l269:
											position, tokenIndex, depth = position268, tokenIndex268, depth268
											if buffer[position] != rune('T') {
												goto l220
											}
											position++
										}
									l268:
										break
									case 'M', 'm':
										{
											position270, tokenIndex270, depth270 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l271
											}
											position++
											goto l270
										l271:
											position, tokenIndex, depth = position270, tokenIndex270, depth270
											if buffer[position] != rune('M') {
												goto l220
											}
											position++
										}
									l270:
										{
											position272, tokenIndex272, depth272 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l273
											}
											position++
											goto l272
										l273:
											position, tokenIndex, depth = position272, tokenIndex272, depth272
											if buffer[position] != rune('A') {
												goto l220
											}
											position++
										}
									l272:
										{
											position274, tokenIndex274, depth274 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l275
											}
											position++
											goto l274
										l275:
											position, tokenIndex, depth = position274, tokenIndex274, depth274
											if buffer[position] != rune('T') {
												goto l220
											}
											position++
										}
									l274:
										{
											position276, tokenIndex276, depth276 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l277
											}
											position++
											goto l276
										l277:
											position, tokenIndex, depth = position276, tokenIndex276, depth276
											if buffer[position] != rune('C') {
												goto l220
											}
											position++
										}
									l276:
										{
											position278, tokenIndex278, depth278 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l279
											}
											position++
											goto l278
										l279:
											position, tokenIndex, depth = position278, tokenIndex278, depth278
											if buffer[position] != rune('H') {
												goto l220
											}
											position++
										}
									l278:
										{
											position280, tokenIndex280, depth280 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l281
											}
											position++
											goto l280
										l281:
											position, tokenIndex, depth = position280, tokenIndex280, depth280
											if buffer[position] != rune('E') {
												goto l220
											}
											position++
										}
									l280:
										{
											position282, tokenIndex282, depth282 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l283
											}
											position++
											goto l282
										l283:
											position, tokenIndex, depth = position282, tokenIndex282, depth282
											if buffer[position] != rune('S') {
												goto l220
											}
											position++
										}
									l282:
										break
									case 'I', 'i':
										{
											position284, tokenIndex284, depth284 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l285
											}
											position++
											goto l284
										l285:
											position, tokenIndex, depth = position284, tokenIndex284, depth284
											if buffer[position] != rune('I') {
												goto l220
											}
											position++
										}
									l284:
										{
											position286, tokenIndex286, depth286 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l287
											}
											position++
											goto l286
										l287:
											position, tokenIndex, depth = position286, tokenIndex286, depth286
											if buffer[position] != rune('N') {
												goto l220
											}
											position++
										}
									l286:
										break
									case 'G', 'g':
										{
											position288, tokenIndex288, depth288 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l289
											}
											position++
											goto l288
										l289:
											position, tokenIndex, depth = position288, tokenIndex288, depth288
											if buffer[position] != rune('G') {
												goto l220
											}
											position++
										}
									l288:
										{
											position290, tokenIndex290, depth290 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l291
											}
											position++
											goto l290
										l291:
											position, tokenIndex, depth = position290, tokenIndex290, depth290
											if buffer[position] != rune('R') {
												goto l220
											}
											position++
										}
									l290:
										{
											position292, tokenIndex292, depth292 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l293
											}
											position++
											goto l292
										l293:
											position, tokenIndex, depth = position292, tokenIndex292, depth292
											if buffer[position] != rune('O') {
												goto l220
											}
											position++
										}
									l292:
										{
											position294, tokenIndex294, depth294 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l295
											}
											position++
											goto l294
										l295:
											position, tokenIndex, depth = position294, tokenIndex294, depth294
											if buffer[position] != rune('U') {
												goto l220
											}
											position++
										}
									l294:
										{
											position296, tokenIndex296, depth296 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l297
											}
											position++
											goto l296
										l297:
											position, tokenIndex, depth = position296, tokenIndex296, depth296
											if buffer[position] != rune('P') {
												goto l220
											}
											position++
										}
									l296:
										break
									case 'F', 'f':
										{
											position298, tokenIndex298, depth298 := position, tokenIndex, depth
											if buffer[position] != rune('f') {
												goto l299
											}
											position++
											goto l298
										l299:
											position, tokenIndex, depth = position298, tokenIndex298, depth298
											if buffer[position] != rune('F') {
												goto l220
											}
											position++
										}
									l298:
										{
											position300, tokenIndex300, depth300 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l301
											}
											position++
											goto l300
										l301:
											position, tokenIndex, depth = position300, tokenIndex300, depth300
											if buffer[position] != rune('R') {
												goto l220
											}
											position++
										}
									l300:
										{
											position302, tokenIndex302, depth302 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l303
											}
											position++
											goto l302
										l303:
											position, tokenIndex, depth = position302, tokenIndex302, depth302
											if buffer[position] != rune('O') {
												goto l220
											}
											position++
										}
									l302:
										{
											position304, tokenIndex304, depth304 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l305
											}
											position++
											goto l304
										l305:
											position, tokenIndex, depth = position304, tokenIndex304, depth304
											if buffer[position] != rune('M') {
												goto l220
											}
											position++
										}
									l304:
										break
									case 'D', 'd':
										{
											position306, tokenIndex306, depth306 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l307
											}
											position++
											goto l306
										l307:
											position, tokenIndex, depth = position306, tokenIndex306, depth306
											if buffer[position] != rune('D') {
												goto l220
											}
											position++
										}
									l306:
										{
											position308, tokenIndex308, depth308 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l309
											}
											position++
											goto l308
										l309:
											position, tokenIndex, depth = position308, tokenIndex308, depth308
											if buffer[position] != rune('E') {
												goto l220
											}
											position++
										}
									l308:
										{
											position310, tokenIndex310, depth310 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l311
											}
											position++
											goto l310
										l311:
											position, tokenIndex, depth = position310, tokenIndex310, depth310
											if buffer[position] != rune('S') {
												goto l220
											}
											position++
										}
									l310:
										{
											position312, tokenIndex312, depth312 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l313
											}
											position++
											goto l312
										l313:
											position, tokenIndex, depth = position312, tokenIndex312, depth312
											if buffer[position] != rune('C') {
												goto l220
											}
											position++
										}
									l312:
										{
											position314, tokenIndex314, depth314 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l315
											}
											position++
											goto l314
										l315:
											position, tokenIndex, depth = position314, tokenIndex314, depth314
											if buffer[position] != rune('R') {
												goto l220
											}
											position++
										}
									l314:
										{
											position316, tokenIndex316, depth316 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l317
											}
											position++
											goto l316
										l317:
											position, tokenIndex, depth = position316, tokenIndex316, depth316
											if buffer[position] != rune('I') {
												goto l220
											}
											position++
										}
									l316:
										{
											position318, tokenIndex318, depth318 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l319
											}
											position++
											goto l318
										l319:
											position, tokenIndex, depth = position318, tokenIndex318, depth318
											if buffer[position] != rune('B') {
												goto l220
											}
											position++
										}
									l318:
										{
											position320, tokenIndex320, depth320 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l321
											}
											position++
											goto l320
										l321:
											position, tokenIndex, depth = position320, tokenIndex320, depth320
											if buffer[position] != rune('E') {
												goto l220
											}
											position++
										}
									l320:
										break
									case 'B', 'b':
										{
											position322, tokenIndex322, depth322 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l323
											}
											position++
											goto l322
										l323:
											position, tokenIndex, depth = position322, tokenIndex322, depth322
											if buffer[position] != rune('B') {
												goto l220
											}
											position++
										}
									l322:
										{
											position324, tokenIndex324, depth324 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l325
											}
											position++
											goto l324
										l325:
											position, tokenIndex, depth = position324, tokenIndex324, depth324
											if buffer[position] != rune('Y') {
												goto l220
											}
											position++
										}
									l324:
										break
									default:
										{
											position326, tokenIndex326, depth326 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l327
											}
											position++
											goto l326
										l327:
											position, tokenIndex, depth = position326, tokenIndex326, depth326
											if buffer[position] != rune('A') {
												goto l220
											}
											position++
										}
									l326:
										{
											position328, tokenIndex328, depth328 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l329
											}
											position++
											goto l328
										l329:
											position, tokenIndex, depth = position328, tokenIndex328, depth328
											if buffer[position] != rune('S') {
												goto l220
											}
											position++
										}
									l328:
										break
									}
								}

							}
						l222:
							depth--
							add(ruleKEYWORD, position221)
						}
						goto l219
					l220:
						position, tokenIndex, depth = position220, tokenIndex220, depth220
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l219
					}
				l330:
					{
						position331, tokenIndex331, depth331 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l331
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l331
						}
						goto l330
					l331:
						position, tokenIndex, depth = position331, tokenIndex331, depth331
					}
					goto l218
				l219:
					position, tokenIndex, depth = position218, tokenIndex218, depth218
					if buffer[position] != rune('`') {
						goto l216
					}
					position++
				l332:
					{
						position333, tokenIndex333, depth333 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l333
						}
						goto l332
					l333:
						position, tokenIndex, depth = position333, tokenIndex333, depth333
					}
					if buffer[position] != rune('`') {
						goto l216
					}
					position++
				}
			l218:
				depth--
				add(ruleIDENTIFIER, position217)
			}
			return true
		l216:
			position, tokenIndex, depth = position216, tokenIndex216, depth216
			return false
		},
		/* 24 ID_SEGMENT <- <(ID_START ID_CONT*)> */
		func() bool {
			position334, tokenIndex334, depth334 := position, tokenIndex, depth
			{
				position335 := position
				depth++
				if !_rules[ruleID_START]() {
					goto l334
				}
			l336:
				{
					position337, tokenIndex337, depth337 := position, tokenIndex, depth
					{
						position338 := position
						depth++
						{
							position339, tokenIndex339, depth339 := position, tokenIndex, depth
							if !_rules[ruleID_START]() {
								goto l340
							}
							goto l339
						l340:
							position, tokenIndex, depth = position339, tokenIndex339, depth339
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l337
							}
							position++
						}
					l339:
						depth--
						add(ruleID_CONT, position338)
					}
					goto l336
				l337:
					position, tokenIndex, depth = position337, tokenIndex337, depth337
				}
				depth--
				add(ruleID_SEGMENT, position335)
			}
			return true
		l334:
			position, tokenIndex, depth = position334, tokenIndex334, depth334
			return false
		},
		/* 25 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position341, tokenIndex341, depth341 := position, tokenIndex, depth
			{
				position342 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l341
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l341
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l341
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position342)
			}
			return true
		l341:
			position, tokenIndex, depth = position341, tokenIndex341, depth341
			return false
		},
		/* 26 ID_CONT <- <(ID_START / [0-9])> */
		nil,
		/* 27 KEYWORD <- <((('a' / 'A') ('l' / 'L') ('l' / 'L')) / (('a' / 'A') ('n' / 'N') ('d' / 'D')) / ((&('W' | 'w') (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E'))) | (&('S' | 's') (('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T'))) | (&('O' | 'o') (('o' / 'O') ('r' / 'R'))) | (&('N' | 'n') (('n' / 'N') ('o' / 'O') ('t' / 'T'))) | (&('M' | 'm') (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S'))) | (&('I' | 'i') (('i' / 'I') ('n' / 'N'))) | (&('G' | 'g') (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P'))) | (&('F' | 'f') (('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))) | (&('D' | 'd') (('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E'))) | (&('B' | 'b') (('b' / 'B') ('y' / 'Y'))) | (&('A' | 'a') (('a' / 'A') ('s' / 'S')))))> */
		nil,
		/* 28 NUMBER <- <INTEGER> */
		nil,
		/* 29 INTEGER <- <('0' / ('-'? [1-9] [0-9]*))> */
		nil,
		/* 30 OP_ADD <- <(_ '+' _)> */
		nil,
		/* 31 OP_SUB <- <(_ '-' _)> */
		nil,
		/* 32 OP_MULT <- <(_ '*' _)> */
		nil,
		/* 33 OP_DIV <- <(_ '/' _)> */
		nil,
		/* 34 OP_AND <- <(__ (('a' / 'A') ('n' / 'N') ('d' / 'D')) __)> */
		nil,
		/* 35 OP_OR <- <(__ (('o' / 'O') ('r' / 'R')) __)> */
		nil,
		/* 36 OP_NOT <- <(('n' / 'N') ('o' / 'O') ('t' / 'T') __)> */
		nil,
		/* 37 STRING <- <(('\'' CHAR* '\'') / ('"' CHAR* '"'))> */
		func() bool {
			position355, tokenIndex355, depth355 := position, tokenIndex, depth
			{
				position356 := position
				depth++
				{
					position357, tokenIndex357, depth357 := position, tokenIndex, depth
					if buffer[position] != rune('\'') {
						goto l358
					}
					position++
				l359:
					{
						position360, tokenIndex360, depth360 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l360
						}
						goto l359
					l360:
						position, tokenIndex, depth = position360, tokenIndex360, depth360
					}
					if buffer[position] != rune('\'') {
						goto l358
					}
					position++
					goto l357
				l358:
					position, tokenIndex, depth = position357, tokenIndex357, depth357
					if buffer[position] != rune('"') {
						goto l355
					}
					position++
				l361:
					{
						position362, tokenIndex362, depth362 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l362
						}
						goto l361
					l362:
						position, tokenIndex, depth = position362, tokenIndex362, depth362
					}
					if buffer[position] != rune('"') {
						goto l355
					}
					position++
				}
			l357:
				depth--
				add(ruleSTRING, position356)
			}
			return true
		l355:
			position, tokenIndex, depth = position355, tokenIndex355, depth355
			return false
		},
		/* 38 CHAR <- <(('\\' ((&('\\') '\\') | (&('"') '"') | (&('`') '`') | (&('\'') '\''))) / (!'\'' .))> */
		func() bool {
			position363, tokenIndex363, depth363 := position, tokenIndex, depth
			{
				position364 := position
				depth++
				{
					position365, tokenIndex365, depth365 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l366
					}
					position++
					{
						switch buffer[position] {
						case '\\':
							if buffer[position] != rune('\\') {
								goto l366
							}
							position++
							break
						case '"':
							if buffer[position] != rune('"') {
								goto l366
							}
							position++
							break
						case '`':
							if buffer[position] != rune('`') {
								goto l366
							}
							position++
							break
						default:
							if buffer[position] != rune('\'') {
								goto l366
							}
							position++
							break
						}
					}

					goto l365
				l366:
					position, tokenIndex, depth = position365, tokenIndex365, depth365
					{
						position368, tokenIndex368, depth368 := position, tokenIndex, depth
						if buffer[position] != rune('\'') {
							goto l368
						}
						position++
						goto l363
					l368:
						position, tokenIndex, depth = position368, tokenIndex368, depth368
					}
					if !matchDot() {
						goto l363
					}
				}
			l365:
				depth--
				add(ruleCHAR, position364)
			}
			return true
		l363:
			position, tokenIndex, depth = position363, tokenIndex363, depth363
			return false
		},
		/* 39 PAREN_OPEN <- <(_ '(' _)> */
		func() bool {
			position369, tokenIndex369, depth369 := position, tokenIndex, depth
			{
				position370 := position
				depth++
				if !_rules[rule_]() {
					goto l369
				}
				if buffer[position] != rune('(') {
					goto l369
				}
				position++
				if !_rules[rule_]() {
					goto l369
				}
				depth--
				add(rulePAREN_OPEN, position370)
			}
			return true
		l369:
			position, tokenIndex, depth = position369, tokenIndex369, depth369
			return false
		},
		/* 40 PAREN_CLOSE <- <(_ ')' _)> */
		func() bool {
			position371, tokenIndex371, depth371 := position, tokenIndex, depth
			{
				position372 := position
				depth++
				if !_rules[rule_]() {
					goto l371
				}
				if buffer[position] != rune(')') {
					goto l371
				}
				position++
				if !_rules[rule_]() {
					goto l371
				}
				depth--
				add(rulePAREN_CLOSE, position372)
			}
			return true
		l371:
			position, tokenIndex, depth = position371, tokenIndex371, depth371
			return false
		},
		/* 41 COMMA <- <(_ ',' _)> */
		func() bool {
			position373, tokenIndex373, depth373 := position, tokenIndex, depth
			{
				position374 := position
				depth++
				if !_rules[rule_]() {
					goto l373
				}
				if buffer[position] != rune(',') {
					goto l373
				}
				position++
				if !_rules[rule_]() {
					goto l373
				}
				depth--
				add(ruleCOMMA, position374)
			}
			return true
		l373:
			position, tokenIndex, depth = position373, tokenIndex373, depth373
			return false
		},
		/* 42 COLON <- <(_ ':' _)> */
		nil,
		/* 43 _ <- <SPACE*> */
		func() bool {
			{
				position377 := position
				depth++
			l378:
				{
					position379, tokenIndex379, depth379 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l379
					}
					goto l378
				l379:
					position, tokenIndex, depth = position379, tokenIndex379, depth379
				}
				depth--
				add(rule_, position377)
			}
			return true
		},
		/* 44 __ <- <SPACE+> */
		func() bool {
			position380, tokenIndex380, depth380 := position, tokenIndex, depth
			{
				position381 := position
				depth++
				if !_rules[ruleSPACE]() {
					goto l380
				}
			l382:
				{
					position383, tokenIndex383, depth383 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l383
					}
					goto l382
				l383:
					position, tokenIndex, depth = position383, tokenIndex383, depth383
				}
				depth--
				add(rule__, position381)
			}
			return true
		l380:
			position, tokenIndex, depth = position380, tokenIndex380, depth380
			return false
		},
		/* 45 SPACE <- <((&('\t') '\t') | (&('\n') '\n') | (&(' ') ' '))> */
		func() bool {
			position384, tokenIndex384, depth384 := position, tokenIndex, depth
			{
				position385 := position
				depth++
				{
					switch buffer[position] {
					case '\t':
						if buffer[position] != rune('\t') {
							goto l384
						}
						position++
						break
					case '\n':
						if buffer[position] != rune('\n') {
							goto l384
						}
						position++
						break
					default:
						if buffer[position] != rune(' ') {
							goto l384
						}
						position++
						break
					}
				}

				depth--
				add(ruleSPACE, position385)
			}
			return true
		l384:
			position, tokenIndex, depth = position384, tokenIndex384, depth384
			return false
		},
		/* 47 Action0 <- <{ p.makeDescribeAll() }> */
		nil,
		nil,
		/* 49 Action1 <- <{ p.addLiteralNode(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 50 Action2 <- <{ p.makeDescribe() }> */
		nil,
		/* 51 Action3 <- <{ p.addNullPredicate() }> */
		nil,
		/* 52 Action4 <- <{ p.addAndMatcher() }> */
		nil,
		/* 53 Action5 <- <{ p.addOrMatcher() }> */
		nil,
		/* 54 Action6 <- <{ p.addNotMatcher() }> */
		nil,
		/* 55 Action7 <- <{
		   p.addLiteralMatcher()
		 }> */
		nil,
		/* 56 Action8 <- <{
		   p.addLiteralMatcher()
		   p.addNotMatcher()
		 }> */
		nil,
		/* 57 Action9 <- <{
		   p.addRegexMatcher()
		 }> */
		nil,
		/* 58 Action10 <- <{
		   p.addListMatcher()
		 }> */
		nil,
		/* 59 Action11 <- <{
		  p.addLiteralNode(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 60 Action12 <- <{ p.addLiteralListNode() }> */
		nil,
		/* 61 Action13 <- <{
		  p.appendLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 62 Action14 <- <{ p.addTagRefNode() }> */
		nil,
		/* 63 Action15 <- <{ p.setAlias(buffer[begin:end]) }> */
		nil,
		/* 64 Action16 <- <{ p.setTag(buffer[begin:end]) }> */
		nil,
	}
	p.rules = _rules
}
