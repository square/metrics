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
	rulerangeClause
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
	ruletagName
	ruleCOLUMN_NAME
	ruleMETRIC_NAME
	ruleTAG_NAME
	ruleTIMESTAMP
	ruleIDENTIFIER
	ruleID_SEGMENT
	ruleID_START
	ruleID_CONT
	ruleKEYWORD
	ruleOP_ADD
	ruleOP_SUB
	ruleOP_MULT
	ruleOP_DIV
	ruleOP_AND
	ruleOP_OR
	ruleOP_NOT
	ruleSTRING
	ruleCHAR
	ruleESCAPE_CLASS
	ruleNUMBER
	ruleNUMBER_NATURAL
	ruleNUMBER_FRACTION
	ruleNUMBER_INTEGER
	ruleNUMBER_EXP
	rulePAREN_OPEN
	rulePAREN_CLOSE
	ruleCOMMA
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
	"rangeClause",
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
	"tagName",
	"COLUMN_NAME",
	"METRIC_NAME",
	"TAG_NAME",
	"TIMESTAMP",
	"IDENTIFIER",
	"ID_SEGMENT",
	"ID_START",
	"ID_CONT",
	"KEYWORD",
	"OP_ADD",
	"OP_SUB",
	"OP_MULT",
	"OP_DIV",
	"OP_AND",
	"OP_OR",
	"OP_NOT",
	"STRING",
	"CHAR",
	"ESCAPE_CLASS",
	"NUMBER",
	"NUMBER_NATURAL",
	"NUMBER_FRACTION",
	"NUMBER_INTEGER",
	"NUMBER_EXP",
	"PAREN_OPEN",
	"PAREN_CLOSE",
	"COMMA",
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
	rules  [68]func() bool
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
			p.addAndPredicate()
		case ruleAction5:
			p.addOrPredicate()
		case ruleAction6:
			p.addNotPredicate()
		case ruleAction7:

			p.addLiteralMatcher()

		case ruleAction8:

			p.addLiteralMatcher()
			p.addNotPredicate()

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
			p.addTag(buffer[begin:end])

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
						{
							position17 := position
							depth++
							if !_rules[rule_]() {
								goto l3
							}
							{
								position18, tokenIndex18, depth18 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l19
								}
								position++
								goto l18
							l19:
								position, tokenIndex, depth = position18, tokenIndex18, depth18
								if buffer[position] != rune('F') {
									goto l3
								}
								position++
							}
						l18:
							{
								position20, tokenIndex20, depth20 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l21
								}
								position++
								goto l20
							l21:
								position, tokenIndex, depth = position20, tokenIndex20, depth20
								if buffer[position] != rune('R') {
									goto l3
								}
								position++
							}
						l20:
							{
								position22, tokenIndex22, depth22 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l23
								}
								position++
								goto l22
							l23:
								position, tokenIndex, depth = position22, tokenIndex22, depth22
								if buffer[position] != rune('O') {
									goto l3
								}
								position++
							}
						l22:
							{
								position24, tokenIndex24, depth24 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l25
								}
								position++
								goto l24
							l25:
								position, tokenIndex, depth = position24, tokenIndex24, depth24
								if buffer[position] != rune('M') {
									goto l3
								}
								position++
							}
						l24:
							if !_rules[rule__]() {
								goto l3
							}
							if !_rules[ruleTIMESTAMP]() {
								goto l3
							}
							if !_rules[rule__]() {
								goto l3
							}
							{
								position26, tokenIndex26, depth26 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l27
								}
								position++
								goto l26
							l27:
								position, tokenIndex, depth = position26, tokenIndex26, depth26
								if buffer[position] != rune('T') {
									goto l3
								}
								position++
							}
						l26:
							{
								position28, tokenIndex28, depth28 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l29
								}
								position++
								goto l28
							l29:
								position, tokenIndex, depth = position28, tokenIndex28, depth28
								if buffer[position] != rune('O') {
									goto l3
								}
								position++
							}
						l28:
							if !_rules[rule__]() {
								goto l3
							}
							if !_rules[ruleTIMESTAMP]() {
								goto l3
							}
							depth--
							add(rulerangeClause, position17)
						}
						depth--
						add(ruleselectStmt, position4)
					}
					goto l2
				l3:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
					{
						position30 := position
						depth++
						{
							position31, tokenIndex31, depth31 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l32
							}
							position++
							goto l31
						l32:
							position, tokenIndex, depth = position31, tokenIndex31, depth31
							if buffer[position] != rune('D') {
								goto l0
							}
							position++
						}
					l31:
						{
							position33, tokenIndex33, depth33 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l34
							}
							position++
							goto l33
						l34:
							position, tokenIndex, depth = position33, tokenIndex33, depth33
							if buffer[position] != rune('E') {
								goto l0
							}
							position++
						}
					l33:
						{
							position35, tokenIndex35, depth35 := position, tokenIndex, depth
							if buffer[position] != rune('s') {
								goto l36
							}
							position++
							goto l35
						l36:
							position, tokenIndex, depth = position35, tokenIndex35, depth35
							if buffer[position] != rune('S') {
								goto l0
							}
							position++
						}
					l35:
						{
							position37, tokenIndex37, depth37 := position, tokenIndex, depth
							if buffer[position] != rune('c') {
								goto l38
							}
							position++
							goto l37
						l38:
							position, tokenIndex, depth = position37, tokenIndex37, depth37
							if buffer[position] != rune('C') {
								goto l0
							}
							position++
						}
					l37:
						{
							position39, tokenIndex39, depth39 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l40
							}
							position++
							goto l39
						l40:
							position, tokenIndex, depth = position39, tokenIndex39, depth39
							if buffer[position] != rune('R') {
								goto l0
							}
							position++
						}
					l39:
						{
							position41, tokenIndex41, depth41 := position, tokenIndex, depth
							if buffer[position] != rune('i') {
								goto l42
							}
							position++
							goto l41
						l42:
							position, tokenIndex, depth = position41, tokenIndex41, depth41
							if buffer[position] != rune('I') {
								goto l0
							}
							position++
						}
					l41:
						{
							position43, tokenIndex43, depth43 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l44
							}
							position++
							goto l43
						l44:
							position, tokenIndex, depth = position43, tokenIndex43, depth43
							if buffer[position] != rune('B') {
								goto l0
							}
							position++
						}
					l43:
						{
							position45, tokenIndex45, depth45 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l46
							}
							position++
							goto l45
						l46:
							position, tokenIndex, depth = position45, tokenIndex45, depth45
							if buffer[position] != rune('E') {
								goto l0
							}
							position++
						}
					l45:
						if !_rules[rule__]() {
							goto l0
						}
						{
							position47, tokenIndex47, depth47 := position, tokenIndex, depth
							{
								position49 := position
								depth++
								{
									position50, tokenIndex50, depth50 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l51
									}
									position++
									goto l50
								l51:
									position, tokenIndex, depth = position50, tokenIndex50, depth50
									if buffer[position] != rune('A') {
										goto l48
									}
									position++
								}
							l50:
								{
									position52, tokenIndex52, depth52 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l53
									}
									position++
									goto l52
								l53:
									position, tokenIndex, depth = position52, tokenIndex52, depth52
									if buffer[position] != rune('L') {
										goto l48
									}
									position++
								}
							l52:
								{
									position54, tokenIndex54, depth54 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l55
									}
									position++
									goto l54
								l55:
									position, tokenIndex, depth = position54, tokenIndex54, depth54
									if buffer[position] != rune('L') {
										goto l48
									}
									position++
								}
							l54:
								{
									add(ruleAction0, position)
								}
								depth--
								add(ruledescribeAllStmt, position49)
							}
							goto l47
						l48:
							position, tokenIndex, depth = position47, tokenIndex47, depth47
							{
								position57 := position
								depth++
								{
									position58 := position
									depth++
									{
										position59 := position
										depth++
										if !_rules[ruleIDENTIFIER]() {
											goto l0
										}
										depth--
										add(ruleMETRIC_NAME, position59)
									}
									depth--
									add(rulePegText, position58)
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
								add(ruledescribeSingleStmt, position57)
							}
						}
					l47:
						depth--
						add(ruledescribeStmt, position30)
					}
				}
			l2:
				{
					position62, tokenIndex62, depth62 := position, tokenIndex, depth
					if !matchDot() {
						goto l62
					}
					goto l0
				l62:
					position, tokenIndex, depth = position62, tokenIndex62, depth62
				}
				depth--
				add(ruleroot, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 selectStmt <- <(('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T') __ expressionList optionalPredicateClause rangeClause)> */
		nil,
		/* 2 describeStmt <- <(('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E') __ (describeAllStmt / describeSingleStmt))> */
		nil,
		/* 3 describeAllStmt <- <(('a' / 'A') ('l' / 'L') ('l' / 'L') Action0)> */
		nil,
		/* 4 describeSingleStmt <- <(<METRIC_NAME> Action1 optionalPredicateClause Action2)> */
		nil,
		/* 5 rangeClause <- <(_ (('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M')) __ TIMESTAMP __ (('t' / 'T') ('o' / 'O')) __ TIMESTAMP)> */
		nil,
		/* 6 optionalPredicateClause <- <((__ predicateClause) / Action3)> */
		func() bool {
			{
				position69 := position
				depth++
				{
					position70, tokenIndex70, depth70 := position, tokenIndex, depth
					if !_rules[rule__]() {
						goto l71
					}
					{
						position72 := position
						depth++
						{
							position73, tokenIndex73, depth73 := position, tokenIndex, depth
							if buffer[position] != rune('w') {
								goto l74
							}
							position++
							goto l73
						l74:
							position, tokenIndex, depth = position73, tokenIndex73, depth73
							if buffer[position] != rune('W') {
								goto l71
							}
							position++
						}
					l73:
						{
							position75, tokenIndex75, depth75 := position, tokenIndex, depth
							if buffer[position] != rune('h') {
								goto l76
							}
							position++
							goto l75
						l76:
							position, tokenIndex, depth = position75, tokenIndex75, depth75
							if buffer[position] != rune('H') {
								goto l71
							}
							position++
						}
					l75:
						{
							position77, tokenIndex77, depth77 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l78
							}
							position++
							goto l77
						l78:
							position, tokenIndex, depth = position77, tokenIndex77, depth77
							if buffer[position] != rune('E') {
								goto l71
							}
							position++
						}
					l77:
						{
							position79, tokenIndex79, depth79 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l80
							}
							position++
							goto l79
						l80:
							position, tokenIndex, depth = position79, tokenIndex79, depth79
							if buffer[position] != rune('R') {
								goto l71
							}
							position++
						}
					l79:
						{
							position81, tokenIndex81, depth81 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l82
							}
							position++
							goto l81
						l82:
							position, tokenIndex, depth = position81, tokenIndex81, depth81
							if buffer[position] != rune('E') {
								goto l71
							}
							position++
						}
					l81:
						if !_rules[rule__]() {
							goto l71
						}
						if !_rules[rulepredicate_1]() {
							goto l71
						}
						depth--
						add(rulepredicateClause, position72)
					}
					goto l70
				l71:
					position, tokenIndex, depth = position70, tokenIndex70, depth70
					{
						add(ruleAction3, position)
					}
				}
			l70:
				depth--
				add(ruleoptionalPredicateClause, position69)
			}
			return true
		},
		/* 7 expressionList <- <(expression_1 (COMMA expression_1)*)> */
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
		/* 8 expression_1 <- <(expression_2 ((OP_ADD / OP_SUB) expression_2)*)> */
		func() bool {
			position88, tokenIndex88, depth88 := position, tokenIndex, depth
			{
				position89 := position
				depth++
				if !_rules[ruleexpression_2]() {
					goto l88
				}
			l90:
				{
					position91, tokenIndex91, depth91 := position, tokenIndex, depth
					{
						position92, tokenIndex92, depth92 := position, tokenIndex, depth
						{
							position94 := position
							depth++
							if !_rules[rule_]() {
								goto l93
							}
							if buffer[position] != rune('+') {
								goto l93
							}
							position++
							if !_rules[rule_]() {
								goto l93
							}
							depth--
							add(ruleOP_ADD, position94)
						}
						goto l92
					l93:
						position, tokenIndex, depth = position92, tokenIndex92, depth92
						{
							position95 := position
							depth++
							if !_rules[rule_]() {
								goto l91
							}
							if buffer[position] != rune('-') {
								goto l91
							}
							position++
							if !_rules[rule_]() {
								goto l91
							}
							depth--
							add(ruleOP_SUB, position95)
						}
					}
				l92:
					if !_rules[ruleexpression_2]() {
						goto l91
					}
					goto l90
				l91:
					position, tokenIndex, depth = position91, tokenIndex91, depth91
				}
				depth--
				add(ruleexpression_1, position89)
			}
			return true
		l88:
			position, tokenIndex, depth = position88, tokenIndex88, depth88
			return false
		},
		/* 9 expression_2 <- <(expression_3 ((OP_DIV / OP_MULT) expression_3)*)> */
		func() bool {
			position96, tokenIndex96, depth96 := position, tokenIndex, depth
			{
				position97 := position
				depth++
				if !_rules[ruleexpression_3]() {
					goto l96
				}
			l98:
				{
					position99, tokenIndex99, depth99 := position, tokenIndex, depth
					{
						position100, tokenIndex100, depth100 := position, tokenIndex, depth
						{
							position102 := position
							depth++
							if !_rules[rule_]() {
								goto l101
							}
							if buffer[position] != rune('/') {
								goto l101
							}
							position++
							if !_rules[rule_]() {
								goto l101
							}
							depth--
							add(ruleOP_DIV, position102)
						}
						goto l100
					l101:
						position, tokenIndex, depth = position100, tokenIndex100, depth100
						{
							position103 := position
							depth++
							if !_rules[rule_]() {
								goto l99
							}
							if buffer[position] != rune('*') {
								goto l99
							}
							position++
							if !_rules[rule_]() {
								goto l99
							}
							depth--
							add(ruleOP_MULT, position103)
						}
					}
				l100:
					if !_rules[ruleexpression_3]() {
						goto l99
					}
					goto l98
				l99:
					position, tokenIndex, depth = position99, tokenIndex99, depth99
				}
				depth--
				add(ruleexpression_2, position97)
			}
			return true
		l96:
			position, tokenIndex, depth = position96, tokenIndex96, depth96
			return false
		},
		/* 10 expression_3 <- <((IDENTIFIER PAREN_OPEN expression_1 __ groupByClause PAREN_CLOSE) / (IDENTIFIER PAREN_OPEN expressionList PAREN_CLOSE) / ((&('-' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') NUMBER) | (&('\t' | '\n' | ' ' | '(') (PAREN_OPEN expression_1 PAREN_CLOSE)) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') (IDENTIFIER ('[' _ predicate_1 _ ']')?))))> */
		func() bool {
			position104, tokenIndex104, depth104 := position, tokenIndex, depth
			{
				position105 := position
				depth++
				{
					position106, tokenIndex106, depth106 := position, tokenIndex, depth
					if !_rules[ruleIDENTIFIER]() {
						goto l107
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l107
					}
					if !_rules[ruleexpression_1]() {
						goto l107
					}
					if !_rules[rule__]() {
						goto l107
					}
					{
						position108 := position
						depth++
						{
							position109, tokenIndex109, depth109 := position, tokenIndex, depth
							if buffer[position] != rune('g') {
								goto l110
							}
							position++
							goto l109
						l110:
							position, tokenIndex, depth = position109, tokenIndex109, depth109
							if buffer[position] != rune('G') {
								goto l107
							}
							position++
						}
					l109:
						{
							position111, tokenIndex111, depth111 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l112
							}
							position++
							goto l111
						l112:
							position, tokenIndex, depth = position111, tokenIndex111, depth111
							if buffer[position] != rune('R') {
								goto l107
							}
							position++
						}
					l111:
						{
							position113, tokenIndex113, depth113 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l114
							}
							position++
							goto l113
						l114:
							position, tokenIndex, depth = position113, tokenIndex113, depth113
							if buffer[position] != rune('O') {
								goto l107
							}
							position++
						}
					l113:
						{
							position115, tokenIndex115, depth115 := position, tokenIndex, depth
							if buffer[position] != rune('u') {
								goto l116
							}
							position++
							goto l115
						l116:
							position, tokenIndex, depth = position115, tokenIndex115, depth115
							if buffer[position] != rune('U') {
								goto l107
							}
							position++
						}
					l115:
						{
							position117, tokenIndex117, depth117 := position, tokenIndex, depth
							if buffer[position] != rune('p') {
								goto l118
							}
							position++
							goto l117
						l118:
							position, tokenIndex, depth = position117, tokenIndex117, depth117
							if buffer[position] != rune('P') {
								goto l107
							}
							position++
						}
					l117:
						if !_rules[rule__]() {
							goto l107
						}
						{
							position119, tokenIndex119, depth119 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l120
							}
							position++
							goto l119
						l120:
							position, tokenIndex, depth = position119, tokenIndex119, depth119
							if buffer[position] != rune('B') {
								goto l107
							}
							position++
						}
					l119:
						{
							position121, tokenIndex121, depth121 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l122
							}
							position++
							goto l121
						l122:
							position, tokenIndex, depth = position121, tokenIndex121, depth121
							if buffer[position] != rune('Y') {
								goto l107
							}
							position++
						}
					l121:
						if !_rules[rule__]() {
							goto l107
						}
						if !_rules[ruleCOLUMN_NAME]() {
							goto l107
						}
					l123:
						{
							position124, tokenIndex124, depth124 := position, tokenIndex, depth
							if !_rules[ruleCOMMA]() {
								goto l124
							}
							if !_rules[ruleCOLUMN_NAME]() {
								goto l124
							}
							goto l123
						l124:
							position, tokenIndex, depth = position124, tokenIndex124, depth124
						}
						depth--
						add(rulegroupByClause, position108)
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l107
					}
					goto l106
				l107:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if !_rules[ruleIDENTIFIER]() {
						goto l125
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l125
					}
					if !_rules[ruleexpressionList]() {
						goto l125
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l125
					}
					goto l106
				l125:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					{
						switch buffer[position] {
						case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							{
								position127 := position
								depth++
								{
									position128 := position
									depth++
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
									if !_rules[ruleNUMBER_NATURAL]() {
										goto l104
									}
									depth--
									add(ruleNUMBER_INTEGER, position128)
								}
								{
									position131, tokenIndex131, depth131 := position, tokenIndex, depth
									{
										position133 := position
										depth++
										if buffer[position] != rune('.') {
											goto l131
										}
										position++
										if c := buffer[position]; c < rune('0') || c > rune('9') {
											goto l131
										}
										position++
									l134:
										{
											position135, tokenIndex135, depth135 := position, tokenIndex, depth
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l135
											}
											position++
											goto l134
										l135:
											position, tokenIndex, depth = position135, tokenIndex135, depth135
										}
										depth--
										add(ruleNUMBER_FRACTION, position133)
									}
									goto l132
								l131:
									position, tokenIndex, depth = position131, tokenIndex131, depth131
								}
							l132:
								{
									position136, tokenIndex136, depth136 := position, tokenIndex, depth
									{
										position138 := position
										depth++
										{
											position139, tokenIndex139, depth139 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l140
											}
											position++
											goto l139
										l140:
											position, tokenIndex, depth = position139, tokenIndex139, depth139
											if buffer[position] != rune('E') {
												goto l136
											}
											position++
										}
									l139:
										{
											position141, tokenIndex141, depth141 := position, tokenIndex, depth
											{
												position143, tokenIndex143, depth143 := position, tokenIndex, depth
												if buffer[position] != rune('+') {
													goto l144
												}
												position++
												goto l143
											l144:
												position, tokenIndex, depth = position143, tokenIndex143, depth143
												if buffer[position] != rune('-') {
													goto l141
												}
												position++
											}
										l143:
											goto l142
										l141:
											position, tokenIndex, depth = position141, tokenIndex141, depth141
										}
									l142:
										if c := buffer[position]; c < rune('0') || c > rune('9') {
											goto l136
										}
										position++
									l145:
										{
											position146, tokenIndex146, depth146 := position, tokenIndex, depth
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l146
											}
											position++
											goto l145
										l146:
											position, tokenIndex, depth = position146, tokenIndex146, depth146
										}
										depth--
										add(ruleNUMBER_EXP, position138)
									}
									goto l137
								l136:
									position, tokenIndex, depth = position136, tokenIndex136, depth136
								}
							l137:
								depth--
								add(ruleNUMBER, position127)
							}
							break
						case '\t', '\n', ' ', '(':
							if !_rules[rulePAREN_OPEN]() {
								goto l104
							}
							if !_rules[ruleexpression_1]() {
								goto l104
							}
							if !_rules[rulePAREN_CLOSE]() {
								goto l104
							}
							break
						default:
							if !_rules[ruleIDENTIFIER]() {
								goto l104
							}
							{
								position147, tokenIndex147, depth147 := position, tokenIndex, depth
								if buffer[position] != rune('[') {
									goto l147
								}
								position++
								if !_rules[rule_]() {
									goto l147
								}
								if !_rules[rulepredicate_1]() {
									goto l147
								}
								if !_rules[rule_]() {
									goto l147
								}
								if buffer[position] != rune(']') {
									goto l147
								}
								position++
								goto l148
							l147:
								position, tokenIndex, depth = position147, tokenIndex147, depth147
							}
						l148:
							break
						}
					}

				}
			l106:
				depth--
				add(ruleexpression_3, position105)
			}
			return true
		l104:
			position, tokenIndex, depth = position104, tokenIndex104, depth104
			return false
		},
		/* 11 groupByClause <- <(('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P') __ (('b' / 'B') ('y' / 'Y')) __ COLUMN_NAME (COMMA COLUMN_NAME)*)> */
		nil,
		/* 12 predicateClause <- <(('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E') __ predicate_1)> */
		nil,
		/* 13 predicate_1 <- <((predicate_2 OP_AND predicate_1 Action4) / predicate_2 / )> */
		func() bool {
			{
				position152 := position
				depth++
				{
					position153, tokenIndex153, depth153 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l154
					}
					{
						position155 := position
						depth++
						if !_rules[rule__]() {
							goto l154
						}
						{
							position156, tokenIndex156, depth156 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l157
							}
							position++
							goto l156
						l157:
							position, tokenIndex, depth = position156, tokenIndex156, depth156
							if buffer[position] != rune('A') {
								goto l154
							}
							position++
						}
					l156:
						{
							position158, tokenIndex158, depth158 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l159
							}
							position++
							goto l158
						l159:
							position, tokenIndex, depth = position158, tokenIndex158, depth158
							if buffer[position] != rune('N') {
								goto l154
							}
							position++
						}
					l158:
						{
							position160, tokenIndex160, depth160 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l161
							}
							position++
							goto l160
						l161:
							position, tokenIndex, depth = position160, tokenIndex160, depth160
							if buffer[position] != rune('D') {
								goto l154
							}
							position++
						}
					l160:
						if !_rules[rule__]() {
							goto l154
						}
						depth--
						add(ruleOP_AND, position155)
					}
					if !_rules[rulepredicate_1]() {
						goto l154
					}
					{
						add(ruleAction4, position)
					}
					goto l153
				l154:
					position, tokenIndex, depth = position153, tokenIndex153, depth153
					if !_rules[rulepredicate_2]() {
						goto l163
					}
					goto l153
				l163:
					position, tokenIndex, depth = position153, tokenIndex153, depth153
				}
			l153:
				depth--
				add(rulepredicate_1, position152)
			}
			return true
		},
		/* 14 predicate_2 <- <((predicate_3 OP_OR predicate_2 Action5) / predicate_3)> */
		func() bool {
			position164, tokenIndex164, depth164 := position, tokenIndex, depth
			{
				position165 := position
				depth++
				{
					position166, tokenIndex166, depth166 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l167
					}
					{
						position168 := position
						depth++
						if !_rules[rule__]() {
							goto l167
						}
						{
							position169, tokenIndex169, depth169 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l170
							}
							position++
							goto l169
						l170:
							position, tokenIndex, depth = position169, tokenIndex169, depth169
							if buffer[position] != rune('O') {
								goto l167
							}
							position++
						}
					l169:
						{
							position171, tokenIndex171, depth171 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l172
							}
							position++
							goto l171
						l172:
							position, tokenIndex, depth = position171, tokenIndex171, depth171
							if buffer[position] != rune('R') {
								goto l167
							}
							position++
						}
					l171:
						if !_rules[rule__]() {
							goto l167
						}
						depth--
						add(ruleOP_OR, position168)
					}
					if !_rules[rulepredicate_2]() {
						goto l167
					}
					{
						add(ruleAction5, position)
					}
					goto l166
				l167:
					position, tokenIndex, depth = position166, tokenIndex166, depth166
					if !_rules[rulepredicate_3]() {
						goto l164
					}
				}
			l166:
				depth--
				add(rulepredicate_2, position165)
			}
			return true
		l164:
			position, tokenIndex, depth = position164, tokenIndex164, depth164
			return false
		},
		/* 15 predicate_3 <- <((OP_NOT predicate_3 Action6) / (PAREN_OPEN predicate_1 PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position174, tokenIndex174, depth174 := position, tokenIndex, depth
			{
				position175 := position
				depth++
				{
					position176, tokenIndex176, depth176 := position, tokenIndex, depth
					{
						position178 := position
						depth++
						{
							position179, tokenIndex179, depth179 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l180
							}
							position++
							goto l179
						l180:
							position, tokenIndex, depth = position179, tokenIndex179, depth179
							if buffer[position] != rune('N') {
								goto l177
							}
							position++
						}
					l179:
						{
							position181, tokenIndex181, depth181 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l182
							}
							position++
							goto l181
						l182:
							position, tokenIndex, depth = position181, tokenIndex181, depth181
							if buffer[position] != rune('O') {
								goto l177
							}
							position++
						}
					l181:
						{
							position183, tokenIndex183, depth183 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l184
							}
							position++
							goto l183
						l184:
							position, tokenIndex, depth = position183, tokenIndex183, depth183
							if buffer[position] != rune('T') {
								goto l177
							}
							position++
						}
					l183:
						if !_rules[rule__]() {
							goto l177
						}
						depth--
						add(ruleOP_NOT, position178)
					}
					if !_rules[rulepredicate_3]() {
						goto l177
					}
					{
						add(ruleAction6, position)
					}
					goto l176
				l177:
					position, tokenIndex, depth = position176, tokenIndex176, depth176
					if !_rules[rulePAREN_OPEN]() {
						goto l186
					}
					if !_rules[rulepredicate_1]() {
						goto l186
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l186
					}
					goto l176
				l186:
					position, tokenIndex, depth = position176, tokenIndex176, depth176
					{
						position187 := position
						depth++
						{
							position188, tokenIndex188, depth188 := position, tokenIndex, depth
							if !_rules[ruletagName]() {
								goto l189
							}
							if !_rules[rule_]() {
								goto l189
							}
							if buffer[position] != rune('=') {
								goto l189
							}
							position++
							if !_rules[rule_]() {
								goto l189
							}
							if !_rules[ruleliteralString]() {
								goto l189
							}
							{
								add(ruleAction7, position)
							}
							goto l188
						l189:
							position, tokenIndex, depth = position188, tokenIndex188, depth188
							if !_rules[ruletagName]() {
								goto l191
							}
							if !_rules[rule_]() {
								goto l191
							}
							if buffer[position] != rune('!') {
								goto l191
							}
							position++
							if buffer[position] != rune('=') {
								goto l191
							}
							position++
							if !_rules[rule_]() {
								goto l191
							}
							if !_rules[ruleliteralString]() {
								goto l191
							}
							{
								add(ruleAction8, position)
							}
							goto l188
						l191:
							position, tokenIndex, depth = position188, tokenIndex188, depth188
							if !_rules[ruletagName]() {
								goto l193
							}
							if !_rules[rule__]() {
								goto l193
							}
							{
								position194, tokenIndex194, depth194 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l195
								}
								position++
								goto l194
							l195:
								position, tokenIndex, depth = position194, tokenIndex194, depth194
								if buffer[position] != rune('M') {
									goto l193
								}
								position++
							}
						l194:
							{
								position196, tokenIndex196, depth196 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l197
								}
								position++
								goto l196
							l197:
								position, tokenIndex, depth = position196, tokenIndex196, depth196
								if buffer[position] != rune('A') {
									goto l193
								}
								position++
							}
						l196:
							{
								position198, tokenIndex198, depth198 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l199
								}
								position++
								goto l198
							l199:
								position, tokenIndex, depth = position198, tokenIndex198, depth198
								if buffer[position] != rune('T') {
									goto l193
								}
								position++
							}
						l198:
							{
								position200, tokenIndex200, depth200 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l201
								}
								position++
								goto l200
							l201:
								position, tokenIndex, depth = position200, tokenIndex200, depth200
								if buffer[position] != rune('C') {
									goto l193
								}
								position++
							}
						l200:
							{
								position202, tokenIndex202, depth202 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l203
								}
								position++
								goto l202
							l203:
								position, tokenIndex, depth = position202, tokenIndex202, depth202
								if buffer[position] != rune('H') {
									goto l193
								}
								position++
							}
						l202:
							{
								position204, tokenIndex204, depth204 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l205
								}
								position++
								goto l204
							l205:
								position, tokenIndex, depth = position204, tokenIndex204, depth204
								if buffer[position] != rune('E') {
									goto l193
								}
								position++
							}
						l204:
							{
								position206, tokenIndex206, depth206 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l207
								}
								position++
								goto l206
							l207:
								position, tokenIndex, depth = position206, tokenIndex206, depth206
								if buffer[position] != rune('S') {
									goto l193
								}
								position++
							}
						l206:
							if !_rules[rule__]() {
								goto l193
							}
							if !_rules[ruleliteralString]() {
								goto l193
							}
							{
								add(ruleAction9, position)
							}
							goto l188
						l193:
							position, tokenIndex, depth = position188, tokenIndex188, depth188
							if !_rules[ruletagName]() {
								goto l174
							}
							if !_rules[rule__]() {
								goto l174
							}
							{
								position209, tokenIndex209, depth209 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l210
								}
								position++
								goto l209
							l210:
								position, tokenIndex, depth = position209, tokenIndex209, depth209
								if buffer[position] != rune('I') {
									goto l174
								}
								position++
							}
						l209:
							{
								position211, tokenIndex211, depth211 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l212
								}
								position++
								goto l211
							l212:
								position, tokenIndex, depth = position211, tokenIndex211, depth211
								if buffer[position] != rune('N') {
									goto l174
								}
								position++
							}
						l211:
							if !_rules[rule__]() {
								goto l174
							}
							{
								position213 := position
								depth++
								{
									add(ruleAction12, position)
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l174
								}
								if !_rules[ruleliteralListString]() {
									goto l174
								}
							l215:
								{
									position216, tokenIndex216, depth216 := position, tokenIndex, depth
									if !_rules[ruleCOMMA]() {
										goto l216
									}
									if !_rules[ruleliteralListString]() {
										goto l216
									}
									goto l215
								l216:
									position, tokenIndex, depth = position216, tokenIndex216, depth216
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l174
								}
								depth--
								add(ruleliteralList, position213)
							}
							{
								add(ruleAction10, position)
							}
						}
					l188:
						depth--
						add(ruletagMatcher, position187)
					}
				}
			l176:
				depth--
				add(rulepredicate_3, position175)
			}
			return true
		l174:
			position, tokenIndex, depth = position174, tokenIndex174, depth174
			return false
		},
		/* 16 tagMatcher <- <((tagName _ '=' _ literalString Action7) / (tagName _ ('!' '=') _ literalString Action8) / (tagName __ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) __ literalString Action9) / (tagName __ (('i' / 'I') ('n' / 'N')) __ literalList Action10))> */
		nil,
		/* 17 literalString <- <(<STRING> Action11)> */
		func() bool {
			position219, tokenIndex219, depth219 := position, tokenIndex, depth
			{
				position220 := position
				depth++
				{
					position221 := position
					depth++
					if !_rules[ruleSTRING]() {
						goto l219
					}
					depth--
					add(rulePegText, position221)
				}
				{
					add(ruleAction11, position)
				}
				depth--
				add(ruleliteralString, position220)
			}
			return true
		l219:
			position, tokenIndex, depth = position219, tokenIndex219, depth219
			return false
		},
		/* 18 literalList <- <(Action12 PAREN_OPEN literalListString (COMMA literalListString)* PAREN_CLOSE)> */
		nil,
		/* 19 literalListString <- <(STRING Action13)> */
		func() bool {
			position224, tokenIndex224, depth224 := position, tokenIndex, depth
			{
				position225 := position
				depth++
				if !_rules[ruleSTRING]() {
					goto l224
				}
				{
					add(ruleAction13, position)
				}
				depth--
				add(ruleliteralListString, position225)
			}
			return true
		l224:
			position, tokenIndex, depth = position224, tokenIndex224, depth224
			return false
		},
		/* 20 tagName <- <(<TAG_NAME> Action14)> */
		func() bool {
			position227, tokenIndex227, depth227 := position, tokenIndex, depth
			{
				position228 := position
				depth++
				{
					position229 := position
					depth++
					{
						position230 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l227
						}
						depth--
						add(ruleTAG_NAME, position230)
					}
					depth--
					add(rulePegText, position229)
				}
				{
					add(ruleAction14, position)
				}
				depth--
				add(ruletagName, position228)
			}
			return true
		l227:
			position, tokenIndex, depth = position227, tokenIndex227, depth227
			return false
		},
		/* 21 COLUMN_NAME <- <IDENTIFIER> */
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
		/* 22 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 23 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 24 TIMESTAMP <- <(NUMBER_NATURAL / STRING)> */
		func() bool {
			position236, tokenIndex236, depth236 := position, tokenIndex, depth
			{
				position237 := position
				depth++
				{
					position238, tokenIndex238, depth238 := position, tokenIndex, depth
					if !_rules[ruleNUMBER_NATURAL]() {
						goto l239
					}
					goto l238
				l239:
					position, tokenIndex, depth = position238, tokenIndex238, depth238
					if !_rules[ruleSTRING]() {
						goto l236
					}
				}
			l238:
				depth--
				add(ruleTIMESTAMP, position237)
			}
			return true
		l236:
			position, tokenIndex, depth = position236, tokenIndex236, depth236
			return false
		},
		/* 25 IDENTIFIER <- <(('`' CHAR* '`') / (!KEYWORD ID_SEGMENT ('.' ID_SEGMENT)*))> */
		func() bool {
			position240, tokenIndex240, depth240 := position, tokenIndex, depth
			{
				position241 := position
				depth++
				{
					position242, tokenIndex242, depth242 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l243
					}
					position++
				l244:
					{
						position245, tokenIndex245, depth245 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l245
						}
						goto l244
					l245:
						position, tokenIndex, depth = position245, tokenIndex245, depth245
					}
					if buffer[position] != rune('`') {
						goto l243
					}
					position++
					goto l242
				l243:
					position, tokenIndex, depth = position242, tokenIndex242, depth242
					{
						position246, tokenIndex246, depth246 := position, tokenIndex, depth
						{
							position247 := position
							depth++
							{
								position248, tokenIndex248, depth248 := position, tokenIndex, depth
								{
									position250, tokenIndex250, depth250 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l251
									}
									position++
									goto l250
								l251:
									position, tokenIndex, depth = position250, tokenIndex250, depth250
									if buffer[position] != rune('A') {
										goto l249
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
										goto l249
									}
									position++
								}
							l252:
								{
									position254, tokenIndex254, depth254 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l255
									}
									position++
									goto l254
								l255:
									position, tokenIndex, depth = position254, tokenIndex254, depth254
									if buffer[position] != rune('L') {
										goto l249
									}
									position++
								}
							l254:
								goto l248
							l249:
								position, tokenIndex, depth = position248, tokenIndex248, depth248
								{
									position257, tokenIndex257, depth257 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l258
									}
									position++
									goto l257
								l258:
									position, tokenIndex, depth = position257, tokenIndex257, depth257
									if buffer[position] != rune('A') {
										goto l256
									}
									position++
								}
							l257:
								{
									position259, tokenIndex259, depth259 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l260
									}
									position++
									goto l259
								l260:
									position, tokenIndex, depth = position259, tokenIndex259, depth259
									if buffer[position] != rune('N') {
										goto l256
									}
									position++
								}
							l259:
								{
									position261, tokenIndex261, depth261 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l262
									}
									position++
									goto l261
								l262:
									position, tokenIndex, depth = position261, tokenIndex261, depth261
									if buffer[position] != rune('D') {
										goto l256
									}
									position++
								}
							l261:
								goto l248
							l256:
								position, tokenIndex, depth = position248, tokenIndex248, depth248
								{
									switch buffer[position] {
									case 'W', 'w':
										{
											position264, tokenIndex264, depth264 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l265
											}
											position++
											goto l264
										l265:
											position, tokenIndex, depth = position264, tokenIndex264, depth264
											if buffer[position] != rune('W') {
												goto l246
											}
											position++
										}
									l264:
										{
											position266, tokenIndex266, depth266 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l267
											}
											position++
											goto l266
										l267:
											position, tokenIndex, depth = position266, tokenIndex266, depth266
											if buffer[position] != rune('H') {
												goto l246
											}
											position++
										}
									l266:
										{
											position268, tokenIndex268, depth268 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l269
											}
											position++
											goto l268
										l269:
											position, tokenIndex, depth = position268, tokenIndex268, depth268
											if buffer[position] != rune('E') {
												goto l246
											}
											position++
										}
									l268:
										{
											position270, tokenIndex270, depth270 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l271
											}
											position++
											goto l270
										l271:
											position, tokenIndex, depth = position270, tokenIndex270, depth270
											if buffer[position] != rune('R') {
												goto l246
											}
											position++
										}
									l270:
										{
											position272, tokenIndex272, depth272 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l273
											}
											position++
											goto l272
										l273:
											position, tokenIndex, depth = position272, tokenIndex272, depth272
											if buffer[position] != rune('E') {
												goto l246
											}
											position++
										}
									l272:
										break
									case 'T', 't':
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
												goto l246
											}
											position++
										}
									l274:
										{
											position276, tokenIndex276, depth276 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l277
											}
											position++
											goto l276
										l277:
											position, tokenIndex, depth = position276, tokenIndex276, depth276
											if buffer[position] != rune('O') {
												goto l246
											}
											position++
										}
									l276:
										break
									case 'S', 's':
										{
											position278, tokenIndex278, depth278 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l279
											}
											position++
											goto l278
										l279:
											position, tokenIndex, depth = position278, tokenIndex278, depth278
											if buffer[position] != rune('S') {
												goto l246
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
												goto l246
											}
											position++
										}
									l280:
										{
											position282, tokenIndex282, depth282 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l283
											}
											position++
											goto l282
										l283:
											position, tokenIndex, depth = position282, tokenIndex282, depth282
											if buffer[position] != rune('L') {
												goto l246
											}
											position++
										}
									l282:
										{
											position284, tokenIndex284, depth284 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l285
											}
											position++
											goto l284
										l285:
											position, tokenIndex, depth = position284, tokenIndex284, depth284
											if buffer[position] != rune('E') {
												goto l246
											}
											position++
										}
									l284:
										{
											position286, tokenIndex286, depth286 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l287
											}
											position++
											goto l286
										l287:
											position, tokenIndex, depth = position286, tokenIndex286, depth286
											if buffer[position] != rune('C') {
												goto l246
											}
											position++
										}
									l286:
										{
											position288, tokenIndex288, depth288 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l289
											}
											position++
											goto l288
										l289:
											position, tokenIndex, depth = position288, tokenIndex288, depth288
											if buffer[position] != rune('T') {
												goto l246
											}
											position++
										}
									l288:
										break
									case 'O', 'o':
										{
											position290, tokenIndex290, depth290 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l291
											}
											position++
											goto l290
										l291:
											position, tokenIndex, depth = position290, tokenIndex290, depth290
											if buffer[position] != rune('O') {
												goto l246
											}
											position++
										}
									l290:
										{
											position292, tokenIndex292, depth292 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l293
											}
											position++
											goto l292
										l293:
											position, tokenIndex, depth = position292, tokenIndex292, depth292
											if buffer[position] != rune('R') {
												goto l246
											}
											position++
										}
									l292:
										break
									case 'N', 'n':
										{
											position294, tokenIndex294, depth294 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l295
											}
											position++
											goto l294
										l295:
											position, tokenIndex, depth = position294, tokenIndex294, depth294
											if buffer[position] != rune('N') {
												goto l246
											}
											position++
										}
									l294:
										{
											position296, tokenIndex296, depth296 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l297
											}
											position++
											goto l296
										l297:
											position, tokenIndex, depth = position296, tokenIndex296, depth296
											if buffer[position] != rune('O') {
												goto l246
											}
											position++
										}
									l296:
										{
											position298, tokenIndex298, depth298 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l299
											}
											position++
											goto l298
										l299:
											position, tokenIndex, depth = position298, tokenIndex298, depth298
											if buffer[position] != rune('T') {
												goto l246
											}
											position++
										}
									l298:
										break
									case 'M', 'm':
										{
											position300, tokenIndex300, depth300 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l301
											}
											position++
											goto l300
										l301:
											position, tokenIndex, depth = position300, tokenIndex300, depth300
											if buffer[position] != rune('M') {
												goto l246
											}
											position++
										}
									l300:
										{
											position302, tokenIndex302, depth302 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l303
											}
											position++
											goto l302
										l303:
											position, tokenIndex, depth = position302, tokenIndex302, depth302
											if buffer[position] != rune('A') {
												goto l246
											}
											position++
										}
									l302:
										{
											position304, tokenIndex304, depth304 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l305
											}
											position++
											goto l304
										l305:
											position, tokenIndex, depth = position304, tokenIndex304, depth304
											if buffer[position] != rune('T') {
												goto l246
											}
											position++
										}
									l304:
										{
											position306, tokenIndex306, depth306 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l307
											}
											position++
											goto l306
										l307:
											position, tokenIndex, depth = position306, tokenIndex306, depth306
											if buffer[position] != rune('C') {
												goto l246
											}
											position++
										}
									l306:
										{
											position308, tokenIndex308, depth308 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l309
											}
											position++
											goto l308
										l309:
											position, tokenIndex, depth = position308, tokenIndex308, depth308
											if buffer[position] != rune('H') {
												goto l246
											}
											position++
										}
									l308:
										{
											position310, tokenIndex310, depth310 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l311
											}
											position++
											goto l310
										l311:
											position, tokenIndex, depth = position310, tokenIndex310, depth310
											if buffer[position] != rune('E') {
												goto l246
											}
											position++
										}
									l310:
										{
											position312, tokenIndex312, depth312 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l313
											}
											position++
											goto l312
										l313:
											position, tokenIndex, depth = position312, tokenIndex312, depth312
											if buffer[position] != rune('S') {
												goto l246
											}
											position++
										}
									l312:
										break
									case 'I', 'i':
										{
											position314, tokenIndex314, depth314 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l315
											}
											position++
											goto l314
										l315:
											position, tokenIndex, depth = position314, tokenIndex314, depth314
											if buffer[position] != rune('I') {
												goto l246
											}
											position++
										}
									l314:
										{
											position316, tokenIndex316, depth316 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l317
											}
											position++
											goto l316
										l317:
											position, tokenIndex, depth = position316, tokenIndex316, depth316
											if buffer[position] != rune('N') {
												goto l246
											}
											position++
										}
									l316:
										break
									case 'G', 'g':
										{
											position318, tokenIndex318, depth318 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l319
											}
											position++
											goto l318
										l319:
											position, tokenIndex, depth = position318, tokenIndex318, depth318
											if buffer[position] != rune('G') {
												goto l246
											}
											position++
										}
									l318:
										{
											position320, tokenIndex320, depth320 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l321
											}
											position++
											goto l320
										l321:
											position, tokenIndex, depth = position320, tokenIndex320, depth320
											if buffer[position] != rune('R') {
												goto l246
											}
											position++
										}
									l320:
										{
											position322, tokenIndex322, depth322 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l323
											}
											position++
											goto l322
										l323:
											position, tokenIndex, depth = position322, tokenIndex322, depth322
											if buffer[position] != rune('O') {
												goto l246
											}
											position++
										}
									l322:
										{
											position324, tokenIndex324, depth324 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l325
											}
											position++
											goto l324
										l325:
											position, tokenIndex, depth = position324, tokenIndex324, depth324
											if buffer[position] != rune('U') {
												goto l246
											}
											position++
										}
									l324:
										{
											position326, tokenIndex326, depth326 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l327
											}
											position++
											goto l326
										l327:
											position, tokenIndex, depth = position326, tokenIndex326, depth326
											if buffer[position] != rune('P') {
												goto l246
											}
											position++
										}
									l326:
										break
									case 'F', 'f':
										{
											position328, tokenIndex328, depth328 := position, tokenIndex, depth
											if buffer[position] != rune('f') {
												goto l329
											}
											position++
											goto l328
										l329:
											position, tokenIndex, depth = position328, tokenIndex328, depth328
											if buffer[position] != rune('F') {
												goto l246
											}
											position++
										}
									l328:
										{
											position330, tokenIndex330, depth330 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l331
											}
											position++
											goto l330
										l331:
											position, tokenIndex, depth = position330, tokenIndex330, depth330
											if buffer[position] != rune('R') {
												goto l246
											}
											position++
										}
									l330:
										{
											position332, tokenIndex332, depth332 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l333
											}
											position++
											goto l332
										l333:
											position, tokenIndex, depth = position332, tokenIndex332, depth332
											if buffer[position] != rune('O') {
												goto l246
											}
											position++
										}
									l332:
										{
											position334, tokenIndex334, depth334 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l335
											}
											position++
											goto l334
										l335:
											position, tokenIndex, depth = position334, tokenIndex334, depth334
											if buffer[position] != rune('M') {
												goto l246
											}
											position++
										}
									l334:
										break
									case 'D', 'd':
										{
											position336, tokenIndex336, depth336 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l337
											}
											position++
											goto l336
										l337:
											position, tokenIndex, depth = position336, tokenIndex336, depth336
											if buffer[position] != rune('D') {
												goto l246
											}
											position++
										}
									l336:
										{
											position338, tokenIndex338, depth338 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l339
											}
											position++
											goto l338
										l339:
											position, tokenIndex, depth = position338, tokenIndex338, depth338
											if buffer[position] != rune('E') {
												goto l246
											}
											position++
										}
									l338:
										{
											position340, tokenIndex340, depth340 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l341
											}
											position++
											goto l340
										l341:
											position, tokenIndex, depth = position340, tokenIndex340, depth340
											if buffer[position] != rune('S') {
												goto l246
											}
											position++
										}
									l340:
										{
											position342, tokenIndex342, depth342 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l343
											}
											position++
											goto l342
										l343:
											position, tokenIndex, depth = position342, tokenIndex342, depth342
											if buffer[position] != rune('C') {
												goto l246
											}
											position++
										}
									l342:
										{
											position344, tokenIndex344, depth344 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l345
											}
											position++
											goto l344
										l345:
											position, tokenIndex, depth = position344, tokenIndex344, depth344
											if buffer[position] != rune('R') {
												goto l246
											}
											position++
										}
									l344:
										{
											position346, tokenIndex346, depth346 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l347
											}
											position++
											goto l346
										l347:
											position, tokenIndex, depth = position346, tokenIndex346, depth346
											if buffer[position] != rune('I') {
												goto l246
											}
											position++
										}
									l346:
										{
											position348, tokenIndex348, depth348 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l349
											}
											position++
											goto l348
										l349:
											position, tokenIndex, depth = position348, tokenIndex348, depth348
											if buffer[position] != rune('B') {
												goto l246
											}
											position++
										}
									l348:
										{
											position350, tokenIndex350, depth350 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l351
											}
											position++
											goto l350
										l351:
											position, tokenIndex, depth = position350, tokenIndex350, depth350
											if buffer[position] != rune('E') {
												goto l246
											}
											position++
										}
									l350:
										break
									case 'B', 'b':
										{
											position352, tokenIndex352, depth352 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l353
											}
											position++
											goto l352
										l353:
											position, tokenIndex, depth = position352, tokenIndex352, depth352
											if buffer[position] != rune('B') {
												goto l246
											}
											position++
										}
									l352:
										{
											position354, tokenIndex354, depth354 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l355
											}
											position++
											goto l354
										l355:
											position, tokenIndex, depth = position354, tokenIndex354, depth354
											if buffer[position] != rune('Y') {
												goto l246
											}
											position++
										}
									l354:
										break
									default:
										{
											position356, tokenIndex356, depth356 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l357
											}
											position++
											goto l356
										l357:
											position, tokenIndex, depth = position356, tokenIndex356, depth356
											if buffer[position] != rune('A') {
												goto l246
											}
											position++
										}
									l356:
										{
											position358, tokenIndex358, depth358 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l359
											}
											position++
											goto l358
										l359:
											position, tokenIndex, depth = position358, tokenIndex358, depth358
											if buffer[position] != rune('S') {
												goto l246
											}
											position++
										}
									l358:
										break
									}
								}

							}
						l248:
							depth--
							add(ruleKEYWORD, position247)
						}
						goto l240
					l246:
						position, tokenIndex, depth = position246, tokenIndex246, depth246
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l240
					}
				l360:
					{
						position361, tokenIndex361, depth361 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l361
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l361
						}
						goto l360
					l361:
						position, tokenIndex, depth = position361, tokenIndex361, depth361
					}
				}
			l242:
				depth--
				add(ruleIDENTIFIER, position241)
			}
			return true
		l240:
			position, tokenIndex, depth = position240, tokenIndex240, depth240
			return false
		},
		/* 26 ID_SEGMENT <- <(ID_START ID_CONT*)> */
		func() bool {
			position362, tokenIndex362, depth362 := position, tokenIndex, depth
			{
				position363 := position
				depth++
				if !_rules[ruleID_START]() {
					goto l362
				}
			l364:
				{
					position365, tokenIndex365, depth365 := position, tokenIndex, depth
					{
						position366 := position
						depth++
						{
							position367, tokenIndex367, depth367 := position, tokenIndex, depth
							if !_rules[ruleID_START]() {
								goto l368
							}
							goto l367
						l368:
							position, tokenIndex, depth = position367, tokenIndex367, depth367
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l365
							}
							position++
						}
					l367:
						depth--
						add(ruleID_CONT, position366)
					}
					goto l364
				l365:
					position, tokenIndex, depth = position365, tokenIndex365, depth365
				}
				depth--
				add(ruleID_SEGMENT, position363)
			}
			return true
		l362:
			position, tokenIndex, depth = position362, tokenIndex362, depth362
			return false
		},
		/* 27 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position369, tokenIndex369, depth369 := position, tokenIndex, depth
			{
				position370 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l369
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l369
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l369
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position370)
			}
			return true
		l369:
			position, tokenIndex, depth = position369, tokenIndex369, depth369
			return false
		},
		/* 28 ID_CONT <- <(ID_START / [0-9])> */
		nil,
		/* 29 KEYWORD <- <((('a' / 'A') ('l' / 'L') ('l' / 'L')) / (('a' / 'A') ('n' / 'N') ('d' / 'D')) / ((&('W' | 'w') (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E'))) | (&('T' | 't') (('t' / 'T') ('o' / 'O'))) | (&('S' | 's') (('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T'))) | (&('O' | 'o') (('o' / 'O') ('r' / 'R'))) | (&('N' | 'n') (('n' / 'N') ('o' / 'O') ('t' / 'T'))) | (&('M' | 'm') (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S'))) | (&('I' | 'i') (('i' / 'I') ('n' / 'N'))) | (&('G' | 'g') (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P'))) | (&('F' | 'f') (('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))) | (&('D' | 'd') (('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E'))) | (&('B' | 'b') (('b' / 'B') ('y' / 'Y'))) | (&('A' | 'a') (('a' / 'A') ('s' / 'S')))))> */
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
			position381, tokenIndex381, depth381 := position, tokenIndex, depth
			{
				position382 := position
				depth++
				{
					position383, tokenIndex383, depth383 := position, tokenIndex, depth
					if buffer[position] != rune('\'') {
						goto l384
					}
					position++
				l385:
					{
						position386, tokenIndex386, depth386 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l386
						}
						goto l385
					l386:
						position, tokenIndex, depth = position386, tokenIndex386, depth386
					}
					if buffer[position] != rune('\'') {
						goto l384
					}
					position++
					goto l383
				l384:
					position, tokenIndex, depth = position383, tokenIndex383, depth383
					if buffer[position] != rune('"') {
						goto l381
					}
					position++
				l387:
					{
						position388, tokenIndex388, depth388 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l388
						}
						goto l387
					l388:
						position, tokenIndex, depth = position388, tokenIndex388, depth388
					}
					if buffer[position] != rune('"') {
						goto l381
					}
					position++
				}
			l383:
				depth--
				add(ruleSTRING, position382)
			}
			return true
		l381:
			position, tokenIndex, depth = position381, tokenIndex381, depth381
			return false
		},
		/* 38 CHAR <- <(('\\' ESCAPE_CLASS) / (!ESCAPE_CLASS .))> */
		func() bool {
			position389, tokenIndex389, depth389 := position, tokenIndex, depth
			{
				position390 := position
				depth++
				{
					position391, tokenIndex391, depth391 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l392
					}
					position++
					if !_rules[ruleESCAPE_CLASS]() {
						goto l392
					}
					goto l391
				l392:
					position, tokenIndex, depth = position391, tokenIndex391, depth391
					{
						position393, tokenIndex393, depth393 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l393
						}
						goto l389
					l393:
						position, tokenIndex, depth = position393, tokenIndex393, depth393
					}
					if !matchDot() {
						goto l389
					}
				}
			l391:
				depth--
				add(ruleCHAR, position390)
			}
			return true
		l389:
			position, tokenIndex, depth = position389, tokenIndex389, depth389
			return false
		},
		/* 39 ESCAPE_CLASS <- <((&('\\') '\\') | (&('"') '"') | (&('`') '`') | (&('\'') '\''))> */
		func() bool {
			position394, tokenIndex394, depth394 := position, tokenIndex, depth
			{
				position395 := position
				depth++
				{
					switch buffer[position] {
					case '\\':
						if buffer[position] != rune('\\') {
							goto l394
						}
						position++
						break
					case '"':
						if buffer[position] != rune('"') {
							goto l394
						}
						position++
						break
					case '`':
						if buffer[position] != rune('`') {
							goto l394
						}
						position++
						break
					default:
						if buffer[position] != rune('\'') {
							goto l394
						}
						position++
						break
					}
				}

				depth--
				add(ruleESCAPE_CLASS, position395)
			}
			return true
		l394:
			position, tokenIndex, depth = position394, tokenIndex394, depth394
			return false
		},
		/* 40 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		nil,
		/* 41 NUMBER_NATURAL <- <('0' / ([1-9] [0-9]*))> */
		func() bool {
			position398, tokenIndex398, depth398 := position, tokenIndex, depth
			{
				position399 := position
				depth++
				{
					position400, tokenIndex400, depth400 := position, tokenIndex, depth
					if buffer[position] != rune('0') {
						goto l401
					}
					position++
					goto l400
				l401:
					position, tokenIndex, depth = position400, tokenIndex400, depth400
					if c := buffer[position]; c < rune('1') || c > rune('9') {
						goto l398
					}
					position++
				l402:
					{
						position403, tokenIndex403, depth403 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l403
						}
						position++
						goto l402
					l403:
						position, tokenIndex, depth = position403, tokenIndex403, depth403
					}
				}
			l400:
				depth--
				add(ruleNUMBER_NATURAL, position399)
			}
			return true
		l398:
			position, tokenIndex, depth = position398, tokenIndex398, depth398
			return false
		},
		/* 42 NUMBER_FRACTION <- <('.' [0-9]+)> */
		nil,
		/* 43 NUMBER_INTEGER <- <('-'? NUMBER_NATURAL)> */
		nil,
		/* 44 NUMBER_EXP <- <(('e' / 'E') ('+' / '-')? [0-9]+)> */
		nil,
		/* 45 PAREN_OPEN <- <(_ '(' _)> */
		func() bool {
			position407, tokenIndex407, depth407 := position, tokenIndex, depth
			{
				position408 := position
				depth++
				if !_rules[rule_]() {
					goto l407
				}
				if buffer[position] != rune('(') {
					goto l407
				}
				position++
				if !_rules[rule_]() {
					goto l407
				}
				depth--
				add(rulePAREN_OPEN, position408)
			}
			return true
		l407:
			position, tokenIndex, depth = position407, tokenIndex407, depth407
			return false
		},
		/* 46 PAREN_CLOSE <- <(_ ')' _)> */
		func() bool {
			position409, tokenIndex409, depth409 := position, tokenIndex, depth
			{
				position410 := position
				depth++
				if !_rules[rule_]() {
					goto l409
				}
				if buffer[position] != rune(')') {
					goto l409
				}
				position++
				if !_rules[rule_]() {
					goto l409
				}
				depth--
				add(rulePAREN_CLOSE, position410)
			}
			return true
		l409:
			position, tokenIndex, depth = position409, tokenIndex409, depth409
			return false
		},
		/* 47 COMMA <- <(_ ',' _)> */
		func() bool {
			position411, tokenIndex411, depth411 := position, tokenIndex, depth
			{
				position412 := position
				depth++
				if !_rules[rule_]() {
					goto l411
				}
				if buffer[position] != rune(',') {
					goto l411
				}
				position++
				if !_rules[rule_]() {
					goto l411
				}
				depth--
				add(ruleCOMMA, position412)
			}
			return true
		l411:
			position, tokenIndex, depth = position411, tokenIndex411, depth411
			return false
		},
		/* 48 _ <- <SPACE*> */
		func() bool {
			{
				position414 := position
				depth++
			l415:
				{
					position416, tokenIndex416, depth416 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l416
					}
					goto l415
				l416:
					position, tokenIndex, depth = position416, tokenIndex416, depth416
				}
				depth--
				add(rule_, position414)
			}
			return true
		},
		/* 49 __ <- <SPACE+> */
		func() bool {
			position417, tokenIndex417, depth417 := position, tokenIndex, depth
			{
				position418 := position
				depth++
				if !_rules[ruleSPACE]() {
					goto l417
				}
			l419:
				{
					position420, tokenIndex420, depth420 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l420
					}
					goto l419
				l420:
					position, tokenIndex, depth = position420, tokenIndex420, depth420
				}
				depth--
				add(rule__, position418)
			}
			return true
		l417:
			position, tokenIndex, depth = position417, tokenIndex417, depth417
			return false
		},
		/* 50 SPACE <- <((&('\t') '\t') | (&('\n') '\n') | (&(' ') ' '))> */
		func() bool {
			position421, tokenIndex421, depth421 := position, tokenIndex, depth
			{
				position422 := position
				depth++
				{
					switch buffer[position] {
					case '\t':
						if buffer[position] != rune('\t') {
							goto l421
						}
						position++
						break
					case '\n':
						if buffer[position] != rune('\n') {
							goto l421
						}
						position++
						break
					default:
						if buffer[position] != rune(' ') {
							goto l421
						}
						position++
						break
					}
				}

				depth--
				add(ruleSPACE, position422)
			}
			return true
		l421:
			position, tokenIndex, depth = position421, tokenIndex421, depth421
			return false
		},
		/* 52 Action0 <- <{ p.makeDescribeAll() }> */
		nil,
		nil,
		/* 54 Action1 <- <{ p.addLiteralNode(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 55 Action2 <- <{ p.makeDescribe() }> */
		nil,
		/* 56 Action3 <- <{ p.addNullPredicate() }> */
		nil,
		/* 57 Action4 <- <{ p.addAndPredicate() }> */
		nil,
		/* 58 Action5 <- <{ p.addOrPredicate() }> */
		nil,
		/* 59 Action6 <- <{ p.addNotPredicate() }> */
		nil,
		/* 60 Action7 <- <{
		   p.addLiteralMatcher()
		 }> */
		nil,
		/* 61 Action8 <- <{
		   p.addLiteralMatcher()
		   p.addNotPredicate()
		 }> */
		nil,
		/* 62 Action9 <- <{
		   p.addRegexMatcher()
		 }> */
		nil,
		/* 63 Action10 <- <{
		   p.addListMatcher()
		 }> */
		nil,
		/* 64 Action11 <- <{
		  p.addLiteralNode(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 65 Action12 <- <{ p.addLiteralListNode() }> */
		nil,
		/* 66 Action13 <- <{
		  p.appendLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 67 Action14 <- <{ p.addTag(buffer[begin:end]) }> */
		nil,
	}
	p.rules = _rules
}
