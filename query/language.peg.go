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
	rulepropertySource
	ruleCOLUMN_NAME
	ruleMETRIC_NAME
	ruleTAG_NAME
	ruleTIMESTAMP
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
	"propertySource",
	"COLUMN_NAME",
	"METRIC_NAME",
	"TAG_NAME",
	"TIMESTAMP",
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
	rules  [67]func() bool
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
								if !_rules[ruleINTEGER]() {
									goto l104
								}
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
								position128, tokenIndex128, depth128 := position, tokenIndex, depth
								if buffer[position] != rune('[') {
									goto l128
								}
								position++
								if !_rules[rule_]() {
									goto l128
								}
								if !_rules[rulepredicate_1]() {
									goto l128
								}
								if !_rules[rule_]() {
									goto l128
								}
								if buffer[position] != rune(']') {
									goto l128
								}
								position++
								goto l129
							l128:
								position, tokenIndex, depth = position128, tokenIndex128, depth128
							}
						l129:
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
				position133 := position
				depth++
				{
					position134, tokenIndex134, depth134 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l135
					}
					{
						position136 := position
						depth++
						if !_rules[rule__]() {
							goto l135
						}
						{
							position137, tokenIndex137, depth137 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l138
							}
							position++
							goto l137
						l138:
							position, tokenIndex, depth = position137, tokenIndex137, depth137
							if buffer[position] != rune('A') {
								goto l135
							}
							position++
						}
					l137:
						{
							position139, tokenIndex139, depth139 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l140
							}
							position++
							goto l139
						l140:
							position, tokenIndex, depth = position139, tokenIndex139, depth139
							if buffer[position] != rune('N') {
								goto l135
							}
							position++
						}
					l139:
						{
							position141, tokenIndex141, depth141 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l142
							}
							position++
							goto l141
						l142:
							position, tokenIndex, depth = position141, tokenIndex141, depth141
							if buffer[position] != rune('D') {
								goto l135
							}
							position++
						}
					l141:
						if !_rules[rule__]() {
							goto l135
						}
						depth--
						add(ruleOP_AND, position136)
					}
					if !_rules[rulepredicate_1]() {
						goto l135
					}
					{
						add(ruleAction4, position)
					}
					goto l134
				l135:
					position, tokenIndex, depth = position134, tokenIndex134, depth134
					if !_rules[rulepredicate_2]() {
						goto l144
					}
					goto l134
				l144:
					position, tokenIndex, depth = position134, tokenIndex134, depth134
				}
			l134:
				depth--
				add(rulepredicate_1, position133)
			}
			return true
		},
		/* 14 predicate_2 <- <((predicate_3 OP_OR predicate_2 Action5) / predicate_3)> */
		func() bool {
			position145, tokenIndex145, depth145 := position, tokenIndex, depth
			{
				position146 := position
				depth++
				{
					position147, tokenIndex147, depth147 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
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
							if buffer[position] != rune('o') {
								goto l151
							}
							position++
							goto l150
						l151:
							position, tokenIndex, depth = position150, tokenIndex150, depth150
							if buffer[position] != rune('O') {
								goto l148
							}
							position++
						}
					l150:
						{
							position152, tokenIndex152, depth152 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l153
							}
							position++
							goto l152
						l153:
							position, tokenIndex, depth = position152, tokenIndex152, depth152
							if buffer[position] != rune('R') {
								goto l148
							}
							position++
						}
					l152:
						if !_rules[rule__]() {
							goto l148
						}
						depth--
						add(ruleOP_OR, position149)
					}
					if !_rules[rulepredicate_2]() {
						goto l148
					}
					{
						add(ruleAction5, position)
					}
					goto l147
				l148:
					position, tokenIndex, depth = position147, tokenIndex147, depth147
					if !_rules[rulepredicate_3]() {
						goto l145
					}
				}
			l147:
				depth--
				add(rulepredicate_2, position146)
			}
			return true
		l145:
			position, tokenIndex, depth = position145, tokenIndex145, depth145
			return false
		},
		/* 15 predicate_3 <- <((OP_NOT predicate_3 Action6) / (PAREN_OPEN predicate_1 PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position155, tokenIndex155, depth155 := position, tokenIndex, depth
			{
				position156 := position
				depth++
				{
					position157, tokenIndex157, depth157 := position, tokenIndex, depth
					{
						position159 := position
						depth++
						{
							position160, tokenIndex160, depth160 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l161
							}
							position++
							goto l160
						l161:
							position, tokenIndex, depth = position160, tokenIndex160, depth160
							if buffer[position] != rune('N') {
								goto l158
							}
							position++
						}
					l160:
						{
							position162, tokenIndex162, depth162 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l163
							}
							position++
							goto l162
						l163:
							position, tokenIndex, depth = position162, tokenIndex162, depth162
							if buffer[position] != rune('O') {
								goto l158
							}
							position++
						}
					l162:
						{
							position164, tokenIndex164, depth164 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l165
							}
							position++
							goto l164
						l165:
							position, tokenIndex, depth = position164, tokenIndex164, depth164
							if buffer[position] != rune('T') {
								goto l158
							}
							position++
						}
					l164:
						if !_rules[rule__]() {
							goto l158
						}
						depth--
						add(ruleOP_NOT, position159)
					}
					if !_rules[rulepredicate_3]() {
						goto l158
					}
					{
						add(ruleAction6, position)
					}
					goto l157
				l158:
					position, tokenIndex, depth = position157, tokenIndex157, depth157
					if !_rules[rulePAREN_OPEN]() {
						goto l167
					}
					if !_rules[rulepredicate_1]() {
						goto l167
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l167
					}
					goto l157
				l167:
					position, tokenIndex, depth = position157, tokenIndex157, depth157
					{
						position168 := position
						depth++
						{
							position169, tokenIndex169, depth169 := position, tokenIndex, depth
							if !_rules[rulepropertySource]() {
								goto l170
							}
							if !_rules[rule_]() {
								goto l170
							}
							if buffer[position] != rune('=') {
								goto l170
							}
							position++
							if !_rules[rule_]() {
								goto l170
							}
							if !_rules[ruleliteralString]() {
								goto l170
							}
							{
								add(ruleAction7, position)
							}
							goto l169
						l170:
							position, tokenIndex, depth = position169, tokenIndex169, depth169
							if !_rules[rulepropertySource]() {
								goto l172
							}
							if !_rules[rule_]() {
								goto l172
							}
							if buffer[position] != rune('!') {
								goto l172
							}
							position++
							if buffer[position] != rune('=') {
								goto l172
							}
							position++
							if !_rules[rule_]() {
								goto l172
							}
							if !_rules[ruleliteralString]() {
								goto l172
							}
							{
								add(ruleAction8, position)
							}
							goto l169
						l172:
							position, tokenIndex, depth = position169, tokenIndex169, depth169
							if !_rules[rulepropertySource]() {
								goto l174
							}
							if !_rules[rule__]() {
								goto l174
							}
							{
								position175, tokenIndex175, depth175 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l176
								}
								position++
								goto l175
							l176:
								position, tokenIndex, depth = position175, tokenIndex175, depth175
								if buffer[position] != rune('M') {
									goto l174
								}
								position++
							}
						l175:
							{
								position177, tokenIndex177, depth177 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l178
								}
								position++
								goto l177
							l178:
								position, tokenIndex, depth = position177, tokenIndex177, depth177
								if buffer[position] != rune('A') {
									goto l174
								}
								position++
							}
						l177:
							{
								position179, tokenIndex179, depth179 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l180
								}
								position++
								goto l179
							l180:
								position, tokenIndex, depth = position179, tokenIndex179, depth179
								if buffer[position] != rune('T') {
									goto l174
								}
								position++
							}
						l179:
							{
								position181, tokenIndex181, depth181 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l182
								}
								position++
								goto l181
							l182:
								position, tokenIndex, depth = position181, tokenIndex181, depth181
								if buffer[position] != rune('C') {
									goto l174
								}
								position++
							}
						l181:
							{
								position183, tokenIndex183, depth183 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l184
								}
								position++
								goto l183
							l184:
								position, tokenIndex, depth = position183, tokenIndex183, depth183
								if buffer[position] != rune('H') {
									goto l174
								}
								position++
							}
						l183:
							{
								position185, tokenIndex185, depth185 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l186
								}
								position++
								goto l185
							l186:
								position, tokenIndex, depth = position185, tokenIndex185, depth185
								if buffer[position] != rune('E') {
									goto l174
								}
								position++
							}
						l185:
							{
								position187, tokenIndex187, depth187 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l188
								}
								position++
								goto l187
							l188:
								position, tokenIndex, depth = position187, tokenIndex187, depth187
								if buffer[position] != rune('S') {
									goto l174
								}
								position++
							}
						l187:
							if !_rules[rule__]() {
								goto l174
							}
							if !_rules[ruleliteralString]() {
								goto l174
							}
							{
								add(ruleAction9, position)
							}
							goto l169
						l174:
							position, tokenIndex, depth = position169, tokenIndex169, depth169
							if !_rules[rulepropertySource]() {
								goto l155
							}
							if !_rules[rule__]() {
								goto l155
							}
							{
								position190, tokenIndex190, depth190 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l191
								}
								position++
								goto l190
							l191:
								position, tokenIndex, depth = position190, tokenIndex190, depth190
								if buffer[position] != rune('I') {
									goto l155
								}
								position++
							}
						l190:
							{
								position192, tokenIndex192, depth192 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l193
								}
								position++
								goto l192
							l193:
								position, tokenIndex, depth = position192, tokenIndex192, depth192
								if buffer[position] != rune('N') {
									goto l155
								}
								position++
							}
						l192:
							if !_rules[rule__]() {
								goto l155
							}
							{
								position194 := position
								depth++
								{
									add(ruleAction12, position)
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l155
								}
								if !_rules[ruleliteralListString]() {
									goto l155
								}
							l196:
								{
									position197, tokenIndex197, depth197 := position, tokenIndex, depth
									if !_rules[ruleCOMMA]() {
										goto l197
									}
									if !_rules[ruleliteralListString]() {
										goto l197
									}
									goto l196
								l197:
									position, tokenIndex, depth = position197, tokenIndex197, depth197
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l155
								}
								depth--
								add(ruleliteralList, position194)
							}
							{
								add(ruleAction10, position)
							}
						}
					l169:
						depth--
						add(ruletagMatcher, position168)
					}
				}
			l157:
				depth--
				add(rulepredicate_3, position156)
			}
			return true
		l155:
			position, tokenIndex, depth = position155, tokenIndex155, depth155
			return false
		},
		/* 16 tagMatcher <- <((propertySource _ '=' _ literalString Action7) / (propertySource _ ('!' '=') _ literalString Action8) / (propertySource __ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) __ literalString Action9) / (propertySource __ (('i' / 'I') ('n' / 'N')) __ literalList Action10))> */
		nil,
		/* 17 literalString <- <(<STRING> Action11)> */
		func() bool {
			position200, tokenIndex200, depth200 := position, tokenIndex, depth
			{
				position201 := position
				depth++
				{
					position202 := position
					depth++
					if !_rules[ruleSTRING]() {
						goto l200
					}
					depth--
					add(rulePegText, position202)
				}
				{
					add(ruleAction11, position)
				}
				depth--
				add(ruleliteralString, position201)
			}
			return true
		l200:
			position, tokenIndex, depth = position200, tokenIndex200, depth200
			return false
		},
		/* 18 literalList <- <(Action12 PAREN_OPEN literalListString (COMMA literalListString)* PAREN_CLOSE)> */
		nil,
		/* 19 literalListString <- <(STRING Action13)> */
		func() bool {
			position205, tokenIndex205, depth205 := position, tokenIndex, depth
			{
				position206 := position
				depth++
				if !_rules[ruleSTRING]() {
					goto l205
				}
				{
					add(ruleAction13, position)
				}
				depth--
				add(ruleliteralListString, position206)
			}
			return true
		l205:
			position, tokenIndex, depth = position205, tokenIndex205, depth205
			return false
		},
		/* 20 propertySource <- <(Action14 (<IDENTIFIER> Action15 COLON)? <TAG_NAME> Action16)> */
		func() bool {
			position208, tokenIndex208, depth208 := position, tokenIndex, depth
			{
				position209 := position
				depth++
				{
					add(ruleAction14, position)
				}
				{
					position211, tokenIndex211, depth211 := position, tokenIndex, depth
					{
						position213 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l211
						}
						depth--
						add(rulePegText, position213)
					}
					{
						add(ruleAction15, position)
					}
					{
						position215 := position
						depth++
						if !_rules[rule_]() {
							goto l211
						}
						if buffer[position] != rune(':') {
							goto l211
						}
						position++
						if !_rules[rule_]() {
							goto l211
						}
						depth--
						add(ruleCOLON, position215)
					}
					goto l212
				l211:
					position, tokenIndex, depth = position211, tokenIndex211, depth211
				}
			l212:
				{
					position216 := position
					depth++
					{
						position217 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l208
						}
						depth--
						add(ruleTAG_NAME, position217)
					}
					depth--
					add(rulePegText, position216)
				}
				{
					add(ruleAction16, position)
				}
				depth--
				add(rulepropertySource, position209)
			}
			return true
		l208:
			position, tokenIndex, depth = position208, tokenIndex208, depth208
			return false
		},
		/* 21 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position219, tokenIndex219, depth219 := position, tokenIndex, depth
			{
				position220 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l219
				}
				depth--
				add(ruleCOLUMN_NAME, position220)
			}
			return true
		l219:
			position, tokenIndex, depth = position219, tokenIndex219, depth219
			return false
		},
		/* 22 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 23 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 24 TIMESTAMP <- <(INTEGER / STRING)> */
		func() bool {
			position223, tokenIndex223, depth223 := position, tokenIndex, depth
			{
				position224 := position
				depth++
				{
					position225, tokenIndex225, depth225 := position, tokenIndex, depth
					if !_rules[ruleINTEGER]() {
						goto l226
					}
					goto l225
				l226:
					position, tokenIndex, depth = position225, tokenIndex225, depth225
					if !_rules[ruleSTRING]() {
						goto l223
					}
				}
			l225:
				depth--
				add(ruleTIMESTAMP, position224)
			}
			return true
		l223:
			position, tokenIndex, depth = position223, tokenIndex223, depth223
			return false
		},
		/* 25 IDENTIFIER <- <((!KEYWORD ID_SEGMENT ('.' ID_SEGMENT)*) / ('`' CHAR* '`'))> */
		func() bool {
			position227, tokenIndex227, depth227 := position, tokenIndex, depth
			{
				position228 := position
				depth++
				{
					position229, tokenIndex229, depth229 := position, tokenIndex, depth
					{
						position231, tokenIndex231, depth231 := position, tokenIndex, depth
						{
							position232 := position
							depth++
							{
								position233, tokenIndex233, depth233 := position, tokenIndex, depth
								{
									position235, tokenIndex235, depth235 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l236
									}
									position++
									goto l235
								l236:
									position, tokenIndex, depth = position235, tokenIndex235, depth235
									if buffer[position] != rune('A') {
										goto l234
									}
									position++
								}
							l235:
								{
									position237, tokenIndex237, depth237 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l238
									}
									position++
									goto l237
								l238:
									position, tokenIndex, depth = position237, tokenIndex237, depth237
									if buffer[position] != rune('L') {
										goto l234
									}
									position++
								}
							l237:
								{
									position239, tokenIndex239, depth239 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l240
									}
									position++
									goto l239
								l240:
									position, tokenIndex, depth = position239, tokenIndex239, depth239
									if buffer[position] != rune('L') {
										goto l234
									}
									position++
								}
							l239:
								goto l233
							l234:
								position, tokenIndex, depth = position233, tokenIndex233, depth233
								{
									position242, tokenIndex242, depth242 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l243
									}
									position++
									goto l242
								l243:
									position, tokenIndex, depth = position242, tokenIndex242, depth242
									if buffer[position] != rune('A') {
										goto l241
									}
									position++
								}
							l242:
								{
									position244, tokenIndex244, depth244 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l245
									}
									position++
									goto l244
								l245:
									position, tokenIndex, depth = position244, tokenIndex244, depth244
									if buffer[position] != rune('N') {
										goto l241
									}
									position++
								}
							l244:
								{
									position246, tokenIndex246, depth246 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l247
									}
									position++
									goto l246
								l247:
									position, tokenIndex, depth = position246, tokenIndex246, depth246
									if buffer[position] != rune('D') {
										goto l241
									}
									position++
								}
							l246:
								goto l233
							l241:
								position, tokenIndex, depth = position233, tokenIndex233, depth233
								{
									switch buffer[position] {
									case 'W', 'w':
										{
											position249, tokenIndex249, depth249 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l250
											}
											position++
											goto l249
										l250:
											position, tokenIndex, depth = position249, tokenIndex249, depth249
											if buffer[position] != rune('W') {
												goto l231
											}
											position++
										}
									l249:
										{
											position251, tokenIndex251, depth251 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l252
											}
											position++
											goto l251
										l252:
											position, tokenIndex, depth = position251, tokenIndex251, depth251
											if buffer[position] != rune('H') {
												goto l231
											}
											position++
										}
									l251:
										{
											position253, tokenIndex253, depth253 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l254
											}
											position++
											goto l253
										l254:
											position, tokenIndex, depth = position253, tokenIndex253, depth253
											if buffer[position] != rune('E') {
												goto l231
											}
											position++
										}
									l253:
										{
											position255, tokenIndex255, depth255 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l256
											}
											position++
											goto l255
										l256:
											position, tokenIndex, depth = position255, tokenIndex255, depth255
											if buffer[position] != rune('R') {
												goto l231
											}
											position++
										}
									l255:
										{
											position257, tokenIndex257, depth257 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l258
											}
											position++
											goto l257
										l258:
											position, tokenIndex, depth = position257, tokenIndex257, depth257
											if buffer[position] != rune('E') {
												goto l231
											}
											position++
										}
									l257:
										break
									case 'T', 't':
										{
											position259, tokenIndex259, depth259 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l260
											}
											position++
											goto l259
										l260:
											position, tokenIndex, depth = position259, tokenIndex259, depth259
											if buffer[position] != rune('T') {
												goto l231
											}
											position++
										}
									l259:
										{
											position261, tokenIndex261, depth261 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l262
											}
											position++
											goto l261
										l262:
											position, tokenIndex, depth = position261, tokenIndex261, depth261
											if buffer[position] != rune('O') {
												goto l231
											}
											position++
										}
									l261:
										break
									case 'S', 's':
										{
											position263, tokenIndex263, depth263 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l264
											}
											position++
											goto l263
										l264:
											position, tokenIndex, depth = position263, tokenIndex263, depth263
											if buffer[position] != rune('S') {
												goto l231
											}
											position++
										}
									l263:
										{
											position265, tokenIndex265, depth265 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l266
											}
											position++
											goto l265
										l266:
											position, tokenIndex, depth = position265, tokenIndex265, depth265
											if buffer[position] != rune('E') {
												goto l231
											}
											position++
										}
									l265:
										{
											position267, tokenIndex267, depth267 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l268
											}
											position++
											goto l267
										l268:
											position, tokenIndex, depth = position267, tokenIndex267, depth267
											if buffer[position] != rune('L') {
												goto l231
											}
											position++
										}
									l267:
										{
											position269, tokenIndex269, depth269 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l270
											}
											position++
											goto l269
										l270:
											position, tokenIndex, depth = position269, tokenIndex269, depth269
											if buffer[position] != rune('E') {
												goto l231
											}
											position++
										}
									l269:
										{
											position271, tokenIndex271, depth271 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l272
											}
											position++
											goto l271
										l272:
											position, tokenIndex, depth = position271, tokenIndex271, depth271
											if buffer[position] != rune('C') {
												goto l231
											}
											position++
										}
									l271:
										{
											position273, tokenIndex273, depth273 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l274
											}
											position++
											goto l273
										l274:
											position, tokenIndex, depth = position273, tokenIndex273, depth273
											if buffer[position] != rune('T') {
												goto l231
											}
											position++
										}
									l273:
										break
									case 'O', 'o':
										{
											position275, tokenIndex275, depth275 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l276
											}
											position++
											goto l275
										l276:
											position, tokenIndex, depth = position275, tokenIndex275, depth275
											if buffer[position] != rune('O') {
												goto l231
											}
											position++
										}
									l275:
										{
											position277, tokenIndex277, depth277 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l278
											}
											position++
											goto l277
										l278:
											position, tokenIndex, depth = position277, tokenIndex277, depth277
											if buffer[position] != rune('R') {
												goto l231
											}
											position++
										}
									l277:
										break
									case 'N', 'n':
										{
											position279, tokenIndex279, depth279 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l280
											}
											position++
											goto l279
										l280:
											position, tokenIndex, depth = position279, tokenIndex279, depth279
											if buffer[position] != rune('N') {
												goto l231
											}
											position++
										}
									l279:
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
												goto l231
											}
											position++
										}
									l281:
										{
											position283, tokenIndex283, depth283 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l284
											}
											position++
											goto l283
										l284:
											position, tokenIndex, depth = position283, tokenIndex283, depth283
											if buffer[position] != rune('T') {
												goto l231
											}
											position++
										}
									l283:
										break
									case 'M', 'm':
										{
											position285, tokenIndex285, depth285 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l286
											}
											position++
											goto l285
										l286:
											position, tokenIndex, depth = position285, tokenIndex285, depth285
											if buffer[position] != rune('M') {
												goto l231
											}
											position++
										}
									l285:
										{
											position287, tokenIndex287, depth287 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l288
											}
											position++
											goto l287
										l288:
											position, tokenIndex, depth = position287, tokenIndex287, depth287
											if buffer[position] != rune('A') {
												goto l231
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
												goto l231
											}
											position++
										}
									l289:
										{
											position291, tokenIndex291, depth291 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l292
											}
											position++
											goto l291
										l292:
											position, tokenIndex, depth = position291, tokenIndex291, depth291
											if buffer[position] != rune('C') {
												goto l231
											}
											position++
										}
									l291:
										{
											position293, tokenIndex293, depth293 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l294
											}
											position++
											goto l293
										l294:
											position, tokenIndex, depth = position293, tokenIndex293, depth293
											if buffer[position] != rune('H') {
												goto l231
											}
											position++
										}
									l293:
										{
											position295, tokenIndex295, depth295 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l296
											}
											position++
											goto l295
										l296:
											position, tokenIndex, depth = position295, tokenIndex295, depth295
											if buffer[position] != rune('E') {
												goto l231
											}
											position++
										}
									l295:
										{
											position297, tokenIndex297, depth297 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l298
											}
											position++
											goto l297
										l298:
											position, tokenIndex, depth = position297, tokenIndex297, depth297
											if buffer[position] != rune('S') {
												goto l231
											}
											position++
										}
									l297:
										break
									case 'I', 'i':
										{
											position299, tokenIndex299, depth299 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l300
											}
											position++
											goto l299
										l300:
											position, tokenIndex, depth = position299, tokenIndex299, depth299
											if buffer[position] != rune('I') {
												goto l231
											}
											position++
										}
									l299:
										{
											position301, tokenIndex301, depth301 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l302
											}
											position++
											goto l301
										l302:
											position, tokenIndex, depth = position301, tokenIndex301, depth301
											if buffer[position] != rune('N') {
												goto l231
											}
											position++
										}
									l301:
										break
									case 'G', 'g':
										{
											position303, tokenIndex303, depth303 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l304
											}
											position++
											goto l303
										l304:
											position, tokenIndex, depth = position303, tokenIndex303, depth303
											if buffer[position] != rune('G') {
												goto l231
											}
											position++
										}
									l303:
										{
											position305, tokenIndex305, depth305 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l306
											}
											position++
											goto l305
										l306:
											position, tokenIndex, depth = position305, tokenIndex305, depth305
											if buffer[position] != rune('R') {
												goto l231
											}
											position++
										}
									l305:
										{
											position307, tokenIndex307, depth307 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l308
											}
											position++
											goto l307
										l308:
											position, tokenIndex, depth = position307, tokenIndex307, depth307
											if buffer[position] != rune('O') {
												goto l231
											}
											position++
										}
									l307:
										{
											position309, tokenIndex309, depth309 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l310
											}
											position++
											goto l309
										l310:
											position, tokenIndex, depth = position309, tokenIndex309, depth309
											if buffer[position] != rune('U') {
												goto l231
											}
											position++
										}
									l309:
										{
											position311, tokenIndex311, depth311 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l312
											}
											position++
											goto l311
										l312:
											position, tokenIndex, depth = position311, tokenIndex311, depth311
											if buffer[position] != rune('P') {
												goto l231
											}
											position++
										}
									l311:
										break
									case 'F', 'f':
										{
											position313, tokenIndex313, depth313 := position, tokenIndex, depth
											if buffer[position] != rune('f') {
												goto l314
											}
											position++
											goto l313
										l314:
											position, tokenIndex, depth = position313, tokenIndex313, depth313
											if buffer[position] != rune('F') {
												goto l231
											}
											position++
										}
									l313:
										{
											position315, tokenIndex315, depth315 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l316
											}
											position++
											goto l315
										l316:
											position, tokenIndex, depth = position315, tokenIndex315, depth315
											if buffer[position] != rune('R') {
												goto l231
											}
											position++
										}
									l315:
										{
											position317, tokenIndex317, depth317 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l318
											}
											position++
											goto l317
										l318:
											position, tokenIndex, depth = position317, tokenIndex317, depth317
											if buffer[position] != rune('O') {
												goto l231
											}
											position++
										}
									l317:
										{
											position319, tokenIndex319, depth319 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l320
											}
											position++
											goto l319
										l320:
											position, tokenIndex, depth = position319, tokenIndex319, depth319
											if buffer[position] != rune('M') {
												goto l231
											}
											position++
										}
									l319:
										break
									case 'D', 'd':
										{
											position321, tokenIndex321, depth321 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l322
											}
											position++
											goto l321
										l322:
											position, tokenIndex, depth = position321, tokenIndex321, depth321
											if buffer[position] != rune('D') {
												goto l231
											}
											position++
										}
									l321:
										{
											position323, tokenIndex323, depth323 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l324
											}
											position++
											goto l323
										l324:
											position, tokenIndex, depth = position323, tokenIndex323, depth323
											if buffer[position] != rune('E') {
												goto l231
											}
											position++
										}
									l323:
										{
											position325, tokenIndex325, depth325 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l326
											}
											position++
											goto l325
										l326:
											position, tokenIndex, depth = position325, tokenIndex325, depth325
											if buffer[position] != rune('S') {
												goto l231
											}
											position++
										}
									l325:
										{
											position327, tokenIndex327, depth327 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l328
											}
											position++
											goto l327
										l328:
											position, tokenIndex, depth = position327, tokenIndex327, depth327
											if buffer[position] != rune('C') {
												goto l231
											}
											position++
										}
									l327:
										{
											position329, tokenIndex329, depth329 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l330
											}
											position++
											goto l329
										l330:
											position, tokenIndex, depth = position329, tokenIndex329, depth329
											if buffer[position] != rune('R') {
												goto l231
											}
											position++
										}
									l329:
										{
											position331, tokenIndex331, depth331 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l332
											}
											position++
											goto l331
										l332:
											position, tokenIndex, depth = position331, tokenIndex331, depth331
											if buffer[position] != rune('I') {
												goto l231
											}
											position++
										}
									l331:
										{
											position333, tokenIndex333, depth333 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l334
											}
											position++
											goto l333
										l334:
											position, tokenIndex, depth = position333, tokenIndex333, depth333
											if buffer[position] != rune('B') {
												goto l231
											}
											position++
										}
									l333:
										{
											position335, tokenIndex335, depth335 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l336
											}
											position++
											goto l335
										l336:
											position, tokenIndex, depth = position335, tokenIndex335, depth335
											if buffer[position] != rune('E') {
												goto l231
											}
											position++
										}
									l335:
										break
									case 'B', 'b':
										{
											position337, tokenIndex337, depth337 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l338
											}
											position++
											goto l337
										l338:
											position, tokenIndex, depth = position337, tokenIndex337, depth337
											if buffer[position] != rune('B') {
												goto l231
											}
											position++
										}
									l337:
										{
											position339, tokenIndex339, depth339 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l340
											}
											position++
											goto l339
										l340:
											position, tokenIndex, depth = position339, tokenIndex339, depth339
											if buffer[position] != rune('Y') {
												goto l231
											}
											position++
										}
									l339:
										break
									default:
										{
											position341, tokenIndex341, depth341 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l342
											}
											position++
											goto l341
										l342:
											position, tokenIndex, depth = position341, tokenIndex341, depth341
											if buffer[position] != rune('A') {
												goto l231
											}
											position++
										}
									l341:
										{
											position343, tokenIndex343, depth343 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l344
											}
											position++
											goto l343
										l344:
											position, tokenIndex, depth = position343, tokenIndex343, depth343
											if buffer[position] != rune('S') {
												goto l231
											}
											position++
										}
									l343:
										break
									}
								}

							}
						l233:
							depth--
							add(ruleKEYWORD, position232)
						}
						goto l230
					l231:
						position, tokenIndex, depth = position231, tokenIndex231, depth231
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l230
					}
				l345:
					{
						position346, tokenIndex346, depth346 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l346
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l346
						}
						goto l345
					l346:
						position, tokenIndex, depth = position346, tokenIndex346, depth346
					}
					goto l229
				l230:
					position, tokenIndex, depth = position229, tokenIndex229, depth229
					if buffer[position] != rune('`') {
						goto l227
					}
					position++
				l347:
					{
						position348, tokenIndex348, depth348 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l348
						}
						goto l347
					l348:
						position, tokenIndex, depth = position348, tokenIndex348, depth348
					}
					if buffer[position] != rune('`') {
						goto l227
					}
					position++
				}
			l229:
				depth--
				add(ruleIDENTIFIER, position228)
			}
			return true
		l227:
			position, tokenIndex, depth = position227, tokenIndex227, depth227
			return false
		},
		/* 26 ID_SEGMENT <- <(ID_START ID_CONT*)> */
		func() bool {
			position349, tokenIndex349, depth349 := position, tokenIndex, depth
			{
				position350 := position
				depth++
				if !_rules[ruleID_START]() {
					goto l349
				}
			l351:
				{
					position352, tokenIndex352, depth352 := position, tokenIndex, depth
					{
						position353 := position
						depth++
						{
							position354, tokenIndex354, depth354 := position, tokenIndex, depth
							if !_rules[ruleID_START]() {
								goto l355
							}
							goto l354
						l355:
							position, tokenIndex, depth = position354, tokenIndex354, depth354
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l352
							}
							position++
						}
					l354:
						depth--
						add(ruleID_CONT, position353)
					}
					goto l351
				l352:
					position, tokenIndex, depth = position352, tokenIndex352, depth352
				}
				depth--
				add(ruleID_SEGMENT, position350)
			}
			return true
		l349:
			position, tokenIndex, depth = position349, tokenIndex349, depth349
			return false
		},
		/* 27 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position356, tokenIndex356, depth356 := position, tokenIndex, depth
			{
				position357 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l356
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l356
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l356
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position357)
			}
			return true
		l356:
			position, tokenIndex, depth = position356, tokenIndex356, depth356
			return false
		},
		/* 28 ID_CONT <- <(ID_START / [0-9])> */
		nil,
		/* 29 KEYWORD <- <((('a' / 'A') ('l' / 'L') ('l' / 'L')) / (('a' / 'A') ('n' / 'N') ('d' / 'D')) / ((&('W' | 'w') (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E'))) | (&('T' | 't') (('t' / 'T') ('o' / 'O'))) | (&('S' | 's') (('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T'))) | (&('O' | 'o') (('o' / 'O') ('r' / 'R'))) | (&('N' | 'n') (('n' / 'N') ('o' / 'O') ('t' / 'T'))) | (&('M' | 'm') (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S'))) | (&('I' | 'i') (('i' / 'I') ('n' / 'N'))) | (&('G' | 'g') (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P'))) | (&('F' | 'f') (('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))) | (&('D' | 'd') (('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E'))) | (&('B' | 'b') (('b' / 'B') ('y' / 'Y'))) | (&('A' | 'a') (('a' / 'A') ('s' / 'S')))))> */
		nil,
		/* 30 NUMBER <- <INTEGER> */
		nil,
		/* 31 INTEGER <- <('0' / ('-'? [1-9] [0-9]*))> */
		func() bool {
			position362, tokenIndex362, depth362 := position, tokenIndex, depth
			{
				position363 := position
				depth++
				{
					position364, tokenIndex364, depth364 := position, tokenIndex, depth
					if buffer[position] != rune('0') {
						goto l365
					}
					position++
					goto l364
				l365:
					position, tokenIndex, depth = position364, tokenIndex364, depth364
					{
						position366, tokenIndex366, depth366 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l366
						}
						position++
						goto l367
					l366:
						position, tokenIndex, depth = position366, tokenIndex366, depth366
					}
				l367:
					if c := buffer[position]; c < rune('1') || c > rune('9') {
						goto l362
					}
					position++
				l368:
					{
						position369, tokenIndex369, depth369 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l369
						}
						position++
						goto l368
					l369:
						position, tokenIndex, depth = position369, tokenIndex369, depth369
					}
				}
			l364:
				depth--
				add(ruleINTEGER, position363)
			}
			return true
		l362:
			position, tokenIndex, depth = position362, tokenIndex362, depth362
			return false
		},
		/* 32 OP_ADD <- <(_ '+' _)> */
		nil,
		/* 33 OP_SUB <- <(_ '-' _)> */
		nil,
		/* 34 OP_MULT <- <(_ '*' _)> */
		nil,
		/* 35 OP_DIV <- <(_ '/' _)> */
		nil,
		/* 36 OP_AND <- <(__ (('a' / 'A') ('n' / 'N') ('d' / 'D')) __)> */
		nil,
		/* 37 OP_OR <- <(__ (('o' / 'O') ('r' / 'R')) __)> */
		nil,
		/* 38 OP_NOT <- <(('n' / 'N') ('o' / 'O') ('t' / 'T') __)> */
		nil,
		/* 39 STRING <- <(('\'' CHAR* '\'') / ('"' CHAR* '"'))> */
		func() bool {
			position377, tokenIndex377, depth377 := position, tokenIndex, depth
			{
				position378 := position
				depth++
				{
					position379, tokenIndex379, depth379 := position, tokenIndex, depth
					if buffer[position] != rune('\'') {
						goto l380
					}
					position++
				l381:
					{
						position382, tokenIndex382, depth382 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l382
						}
						goto l381
					l382:
						position, tokenIndex, depth = position382, tokenIndex382, depth382
					}
					if buffer[position] != rune('\'') {
						goto l380
					}
					position++
					goto l379
				l380:
					position, tokenIndex, depth = position379, tokenIndex379, depth379
					if buffer[position] != rune('"') {
						goto l377
					}
					position++
				l383:
					{
						position384, tokenIndex384, depth384 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l384
						}
						goto l383
					l384:
						position, tokenIndex, depth = position384, tokenIndex384, depth384
					}
					if buffer[position] != rune('"') {
						goto l377
					}
					position++
				}
			l379:
				depth--
				add(ruleSTRING, position378)
			}
			return true
		l377:
			position, tokenIndex, depth = position377, tokenIndex377, depth377
			return false
		},
		/* 40 CHAR <- <(('\\' ((&('\\') '\\') | (&('"') '"') | (&('`') '`') | (&('\'') '\''))) / (!'\'' .))> */
		func() bool {
			position385, tokenIndex385, depth385 := position, tokenIndex, depth
			{
				position386 := position
				depth++
				{
					position387, tokenIndex387, depth387 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l388
					}
					position++
					{
						switch buffer[position] {
						case '\\':
							if buffer[position] != rune('\\') {
								goto l388
							}
							position++
							break
						case '"':
							if buffer[position] != rune('"') {
								goto l388
							}
							position++
							break
						case '`':
							if buffer[position] != rune('`') {
								goto l388
							}
							position++
							break
						default:
							if buffer[position] != rune('\'') {
								goto l388
							}
							position++
							break
						}
					}

					goto l387
				l388:
					position, tokenIndex, depth = position387, tokenIndex387, depth387
					{
						position390, tokenIndex390, depth390 := position, tokenIndex, depth
						if buffer[position] != rune('\'') {
							goto l390
						}
						position++
						goto l385
					l390:
						position, tokenIndex, depth = position390, tokenIndex390, depth390
					}
					if !matchDot() {
						goto l385
					}
				}
			l387:
				depth--
				add(ruleCHAR, position386)
			}
			return true
		l385:
			position, tokenIndex, depth = position385, tokenIndex385, depth385
			return false
		},
		/* 41 PAREN_OPEN <- <(_ '(' _)> */
		func() bool {
			position391, tokenIndex391, depth391 := position, tokenIndex, depth
			{
				position392 := position
				depth++
				if !_rules[rule_]() {
					goto l391
				}
				if buffer[position] != rune('(') {
					goto l391
				}
				position++
				if !_rules[rule_]() {
					goto l391
				}
				depth--
				add(rulePAREN_OPEN, position392)
			}
			return true
		l391:
			position, tokenIndex, depth = position391, tokenIndex391, depth391
			return false
		},
		/* 42 PAREN_CLOSE <- <(_ ')' _)> */
		func() bool {
			position393, tokenIndex393, depth393 := position, tokenIndex, depth
			{
				position394 := position
				depth++
				if !_rules[rule_]() {
					goto l393
				}
				if buffer[position] != rune(')') {
					goto l393
				}
				position++
				if !_rules[rule_]() {
					goto l393
				}
				depth--
				add(rulePAREN_CLOSE, position394)
			}
			return true
		l393:
			position, tokenIndex, depth = position393, tokenIndex393, depth393
			return false
		},
		/* 43 COMMA <- <(_ ',' _)> */
		func() bool {
			position395, tokenIndex395, depth395 := position, tokenIndex, depth
			{
				position396 := position
				depth++
				if !_rules[rule_]() {
					goto l395
				}
				if buffer[position] != rune(',') {
					goto l395
				}
				position++
				if !_rules[rule_]() {
					goto l395
				}
				depth--
				add(ruleCOMMA, position396)
			}
			return true
		l395:
			position, tokenIndex, depth = position395, tokenIndex395, depth395
			return false
		},
		/* 44 COLON <- <(_ ':' _)> */
		nil,
		/* 45 _ <- <SPACE*> */
		func() bool {
			{
				position399 := position
				depth++
			l400:
				{
					position401, tokenIndex401, depth401 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l401
					}
					goto l400
				l401:
					position, tokenIndex, depth = position401, tokenIndex401, depth401
				}
				depth--
				add(rule_, position399)
			}
			return true
		},
		/* 46 __ <- <SPACE+> */
		func() bool {
			position402, tokenIndex402, depth402 := position, tokenIndex, depth
			{
				position403 := position
				depth++
				if !_rules[ruleSPACE]() {
					goto l402
				}
			l404:
				{
					position405, tokenIndex405, depth405 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l405
					}
					goto l404
				l405:
					position, tokenIndex, depth = position405, tokenIndex405, depth405
				}
				depth--
				add(rule__, position403)
			}
			return true
		l402:
			position, tokenIndex, depth = position402, tokenIndex402, depth402
			return false
		},
		/* 47 SPACE <- <((&('\t') '\t') | (&('\n') '\n') | (&(' ') ' '))> */
		func() bool {
			position406, tokenIndex406, depth406 := position, tokenIndex, depth
			{
				position407 := position
				depth++
				{
					switch buffer[position] {
					case '\t':
						if buffer[position] != rune('\t') {
							goto l406
						}
						position++
						break
					case '\n':
						if buffer[position] != rune('\n') {
							goto l406
						}
						position++
						break
					default:
						if buffer[position] != rune(' ') {
							goto l406
						}
						position++
						break
					}
				}

				depth--
				add(ruleSPACE, position407)
			}
			return true
		l406:
			position, tokenIndex, depth = position406, tokenIndex406, depth406
			return false
		},
		/* 49 Action0 <- <{ p.makeDescribeAll() }> */
		nil,
		nil,
		/* 51 Action1 <- <{ p.addLiteralNode(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 52 Action2 <- <{ p.makeDescribe() }> */
		nil,
		/* 53 Action3 <- <{ p.addNullPredicate() }> */
		nil,
		/* 54 Action4 <- <{ p.addAndMatcher() }> */
		nil,
		/* 55 Action5 <- <{ p.addOrMatcher() }> */
		nil,
		/* 56 Action6 <- <{ p.addNotMatcher() }> */
		nil,
		/* 57 Action7 <- <{
		   p.addLiteralMatcher()
		 }> */
		nil,
		/* 58 Action8 <- <{
		   p.addLiteralMatcher()
		   p.addNotMatcher()
		 }> */
		nil,
		/* 59 Action9 <- <{
		   p.addRegexMatcher()
		 }> */
		nil,
		/* 60 Action10 <- <{
		   p.addListMatcher()
		 }> */
		nil,
		/* 61 Action11 <- <{
		  p.addLiteralNode(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 62 Action12 <- <{ p.addLiteralListNode() }> */
		nil,
		/* 63 Action13 <- <{
		  p.appendLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 64 Action14 <- <{ p.addTagRefNode() }> */
		nil,
		/* 65 Action15 <- <{ p.setAlias(buffer[begin:end]) }> */
		nil,
		/* 66 Action16 <- <{ p.setTag(buffer[begin:end]) }> */
		nil,
	}
	p.rules = _rules
}
