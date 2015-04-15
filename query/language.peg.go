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
	ruledescribeAllMetricsStmt
	ruledescribeSingleMetricStmt
	rulepredicateClause
	rulepredicate_1
	rulepredicate_2
	rulepredicate_3
	ruletagMatcher
	ruleliteralString
	ruleliteralList
	ruleliteralListString
	rulematcherPart
	ruleALIAS_NAME
	ruleMETRIC_NAME
	ruleTAG_NAME
	ruleIDENTIFIER
	ruleID_SEGMENT
	ruleID_START
	ruleID_CONT
	ruleSTRING
	ruleCHAR
	rule__
	rule_
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
	"describeAllMetricsStmt",
	"describeSingleMetricStmt",
	"predicateClause",
	"predicate_1",
	"predicate_2",
	"predicate_3",
	"tagMatcher",
	"literalString",
	"literalList",
	"literalListString",
	"matcherPart",
	"ALIAS_NAME",
	"METRIC_NAME",
	"TAG_NAME",
	"IDENTIFIER",
	"ID_SEGMENT",
	"ID_START",
	"ID_CONT",
	"STRING",
	"CHAR",
	"__",
	"_",
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
	rules  [44]func() bool
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
			p.addNullPredicate()
		case ruleAction3:
			p.makeDescribe()
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
					l17:
						{
							position18, tokenIndex18, depth18 := position, tokenIndex, depth
							if !matchDot() {
								goto l18
							}
							goto l17
						l18:
							position, tokenIndex, depth = position18, tokenIndex18, depth18
						}
						depth--
						add(ruleselectStmt, position4)
					}
					goto l2
				l3:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
					{
						position19 := position
						depth++
						{
							position20, tokenIndex20, depth20 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l21
							}
							position++
							goto l20
						l21:
							position, tokenIndex, depth = position20, tokenIndex20, depth20
							if buffer[position] != rune('D') {
								goto l0
							}
							position++
						}
					l20:
						{
							position22, tokenIndex22, depth22 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l23
							}
							position++
							goto l22
						l23:
							position, tokenIndex, depth = position22, tokenIndex22, depth22
							if buffer[position] != rune('E') {
								goto l0
							}
							position++
						}
					l22:
						{
							position24, tokenIndex24, depth24 := position, tokenIndex, depth
							if buffer[position] != rune('s') {
								goto l25
							}
							position++
							goto l24
						l25:
							position, tokenIndex, depth = position24, tokenIndex24, depth24
							if buffer[position] != rune('S') {
								goto l0
							}
							position++
						}
					l24:
						{
							position26, tokenIndex26, depth26 := position, tokenIndex, depth
							if buffer[position] != rune('c') {
								goto l27
							}
							position++
							goto l26
						l27:
							position, tokenIndex, depth = position26, tokenIndex26, depth26
							if buffer[position] != rune('C') {
								goto l0
							}
							position++
						}
					l26:
						{
							position28, tokenIndex28, depth28 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l29
							}
							position++
							goto l28
						l29:
							position, tokenIndex, depth = position28, tokenIndex28, depth28
							if buffer[position] != rune('R') {
								goto l0
							}
							position++
						}
					l28:
						{
							position30, tokenIndex30, depth30 := position, tokenIndex, depth
							if buffer[position] != rune('i') {
								goto l31
							}
							position++
							goto l30
						l31:
							position, tokenIndex, depth = position30, tokenIndex30, depth30
							if buffer[position] != rune('I') {
								goto l0
							}
							position++
						}
					l30:
						{
							position32, tokenIndex32, depth32 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l33
							}
							position++
							goto l32
						l33:
							position, tokenIndex, depth = position32, tokenIndex32, depth32
							if buffer[position] != rune('B') {
								goto l0
							}
							position++
						}
					l32:
						{
							position34, tokenIndex34, depth34 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l35
							}
							position++
							goto l34
						l35:
							position, tokenIndex, depth = position34, tokenIndex34, depth34
							if buffer[position] != rune('E') {
								goto l0
							}
							position++
						}
					l34:
						if !_rules[rule__]() {
							goto l0
						}
						{
							position36, tokenIndex36, depth36 := position, tokenIndex, depth
							{
								position38 := position
								depth++
								{
									position39, tokenIndex39, depth39 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l40
									}
									position++
									goto l39
								l40:
									position, tokenIndex, depth = position39, tokenIndex39, depth39
									if buffer[position] != rune('A') {
										goto l37
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
										goto l37
									}
									position++
								}
							l41:
								{
									position43, tokenIndex43, depth43 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l44
									}
									position++
									goto l43
								l44:
									position, tokenIndex, depth = position43, tokenIndex43, depth43
									if buffer[position] != rune('L') {
										goto l37
									}
									position++
								}
							l43:
								if !_rules[rule__]() {
									goto l37
								}
								{
									position45, tokenIndex45, depth45 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l46
									}
									position++
									goto l45
								l46:
									position, tokenIndex, depth = position45, tokenIndex45, depth45
									if buffer[position] != rune('M') {
										goto l37
									}
									position++
								}
							l45:
								{
									position47, tokenIndex47, depth47 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l48
									}
									position++
									goto l47
								l48:
									position, tokenIndex, depth = position47, tokenIndex47, depth47
									if buffer[position] != rune('E') {
										goto l37
									}
									position++
								}
							l47:
								{
									position49, tokenIndex49, depth49 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l50
									}
									position++
									goto l49
								l50:
									position, tokenIndex, depth = position49, tokenIndex49, depth49
									if buffer[position] != rune('T') {
										goto l37
									}
									position++
								}
							l49:
								{
									position51, tokenIndex51, depth51 := position, tokenIndex, depth
									if buffer[position] != rune('r') {
										goto l52
									}
									position++
									goto l51
								l52:
									position, tokenIndex, depth = position51, tokenIndex51, depth51
									if buffer[position] != rune('R') {
										goto l37
									}
									position++
								}
							l51:
								{
									position53, tokenIndex53, depth53 := position, tokenIndex, depth
									if buffer[position] != rune('i') {
										goto l54
									}
									position++
									goto l53
								l54:
									position, tokenIndex, depth = position53, tokenIndex53, depth53
									if buffer[position] != rune('I') {
										goto l37
									}
									position++
								}
							l53:
								{
									position55, tokenIndex55, depth55 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l56
									}
									position++
									goto l55
								l56:
									position, tokenIndex, depth = position55, tokenIndex55, depth55
									if buffer[position] != rune('C') {
										goto l37
									}
									position++
								}
							l55:
								{
									position57, tokenIndex57, depth57 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l58
									}
									position++
									goto l57
								l58:
									position, tokenIndex, depth = position57, tokenIndex57, depth57
									if buffer[position] != rune('S') {
										goto l37
									}
									position++
								}
							l57:
								{
									add(ruleAction0, position)
								}
								depth--
								add(ruledescribeAllMetricsStmt, position38)
							}
							goto l36
						l37:
							position, tokenIndex, depth = position36, tokenIndex36, depth36
							{
								position60 := position
								depth++
								{
									position61 := position
									depth++
									{
										position62 := position
										depth++
										if !_rules[ruleIDENTIFIER]() {
											goto l0
										}
										depth--
										add(ruleMETRIC_NAME, position62)
									}
									depth--
									add(rulePegText, position61)
								}
								{
									add(ruleAction1, position)
								}
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
										add(ruleAction2, position)
									}
								}
							l64:
								{
									add(ruleAction3, position)
								}
								depth--
								add(ruledescribeSingleMetricStmt, position60)
							}
						}
					l36:
						depth--
						add(ruledescribeStmt, position19)
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
		/* 1 selectStmt <- <(('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T') __ .*)> */
		nil,
		/* 2 describeStmt <- <(('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E') __ (describeAllMetricsStmt / describeSingleMetricStmt))> */
		nil,
		/* 3 describeAllMetricsStmt <- <(('a' / 'A') ('l' / 'L') ('l' / 'L') __ (('m' / 'M') ('e' / 'E') ('t' / 'T') ('r' / 'R') ('i' / 'I') ('c' / 'C') ('s' / 'S')) Action0)> */
		nil,
		/* 4 describeSingleMetricStmt <- <(<METRIC_NAME> Action1 ((__ predicateClause) / Action2) Action3)> */
		nil,
		/* 5 predicateClause <- <(('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E') __ predicate_1)> */
		nil,
		/* 6 predicate_1 <- <((predicate_2 __ (('a' / 'A') ('n' / 'N') ('d' / 'D')) __ predicate_1 Action4) / predicate_2 / )> */
		func() bool {
			{
				position86 := position
				depth++
				{
					position87, tokenIndex87, depth87 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l88
					}
					if !_rules[rule__]() {
						goto l88
					}
					{
						position89, tokenIndex89, depth89 := position, tokenIndex, depth
						if buffer[position] != rune('a') {
							goto l90
						}
						position++
						goto l89
					l90:
						position, tokenIndex, depth = position89, tokenIndex89, depth89
						if buffer[position] != rune('A') {
							goto l88
						}
						position++
					}
				l89:
					{
						position91, tokenIndex91, depth91 := position, tokenIndex, depth
						if buffer[position] != rune('n') {
							goto l92
						}
						position++
						goto l91
					l92:
						position, tokenIndex, depth = position91, tokenIndex91, depth91
						if buffer[position] != rune('N') {
							goto l88
						}
						position++
					}
				l91:
					{
						position93, tokenIndex93, depth93 := position, tokenIndex, depth
						if buffer[position] != rune('d') {
							goto l94
						}
						position++
						goto l93
					l94:
						position, tokenIndex, depth = position93, tokenIndex93, depth93
						if buffer[position] != rune('D') {
							goto l88
						}
						position++
					}
				l93:
					if !_rules[rule__]() {
						goto l88
					}
					if !_rules[rulepredicate_1]() {
						goto l88
					}
					{
						add(ruleAction4, position)
					}
					goto l87
				l88:
					position, tokenIndex, depth = position87, tokenIndex87, depth87
					if !_rules[rulepredicate_2]() {
						goto l96
					}
					goto l87
				l96:
					position, tokenIndex, depth = position87, tokenIndex87, depth87
				}
			l87:
				depth--
				add(rulepredicate_1, position86)
			}
			return true
		},
		/* 7 predicate_2 <- <((predicate_3 __ (('o' / 'O') ('r' / 'R')) __ predicate_2 Action5) / predicate_3)> */
		func() bool {
			position97, tokenIndex97, depth97 := position, tokenIndex, depth
			{
				position98 := position
				depth++
				{
					position99, tokenIndex99, depth99 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l100
					}
					if !_rules[rule__]() {
						goto l100
					}
					{
						position101, tokenIndex101, depth101 := position, tokenIndex, depth
						if buffer[position] != rune('o') {
							goto l102
						}
						position++
						goto l101
					l102:
						position, tokenIndex, depth = position101, tokenIndex101, depth101
						if buffer[position] != rune('O') {
							goto l100
						}
						position++
					}
				l101:
					{
						position103, tokenIndex103, depth103 := position, tokenIndex, depth
						if buffer[position] != rune('r') {
							goto l104
						}
						position++
						goto l103
					l104:
						position, tokenIndex, depth = position103, tokenIndex103, depth103
						if buffer[position] != rune('R') {
							goto l100
						}
						position++
					}
				l103:
					if !_rules[rule__]() {
						goto l100
					}
					if !_rules[rulepredicate_2]() {
						goto l100
					}
					{
						add(ruleAction5, position)
					}
					goto l99
				l100:
					position, tokenIndex, depth = position99, tokenIndex99, depth99
					if !_rules[rulepredicate_3]() {
						goto l97
					}
				}
			l99:
				depth--
				add(rulepredicate_2, position98)
			}
			return true
		l97:
			position, tokenIndex, depth = position97, tokenIndex97, depth97
			return false
		},
		/* 8 predicate_3 <- <((('n' / 'N') ('o' / 'O') ('t' / 'T') __ predicate_3 Action6) / ('(' _ predicate_1 _ ')') / tagMatcher)> */
		func() bool {
			position106, tokenIndex106, depth106 := position, tokenIndex, depth
			{
				position107 := position
				depth++
				{
					position108, tokenIndex108, depth108 := position, tokenIndex, depth
					{
						position110, tokenIndex110, depth110 := position, tokenIndex, depth
						if buffer[position] != rune('n') {
							goto l111
						}
						position++
						goto l110
					l111:
						position, tokenIndex, depth = position110, tokenIndex110, depth110
						if buffer[position] != rune('N') {
							goto l109
						}
						position++
					}
				l110:
					{
						position112, tokenIndex112, depth112 := position, tokenIndex, depth
						if buffer[position] != rune('o') {
							goto l113
						}
						position++
						goto l112
					l113:
						position, tokenIndex, depth = position112, tokenIndex112, depth112
						if buffer[position] != rune('O') {
							goto l109
						}
						position++
					}
				l112:
					{
						position114, tokenIndex114, depth114 := position, tokenIndex, depth
						if buffer[position] != rune('t') {
							goto l115
						}
						position++
						goto l114
					l115:
						position, tokenIndex, depth = position114, tokenIndex114, depth114
						if buffer[position] != rune('T') {
							goto l109
						}
						position++
					}
				l114:
					if !_rules[rule__]() {
						goto l109
					}
					if !_rules[rulepredicate_3]() {
						goto l109
					}
					{
						add(ruleAction6, position)
					}
					goto l108
				l109:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
					if buffer[position] != rune('(') {
						goto l117
					}
					position++
					if !_rules[rule_]() {
						goto l117
					}
					if !_rules[rulepredicate_1]() {
						goto l117
					}
					if !_rules[rule_]() {
						goto l117
					}
					if buffer[position] != rune(')') {
						goto l117
					}
					position++
					goto l108
				l117:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
					{
						position118 := position
						depth++
						{
							position119, tokenIndex119, depth119 := position, tokenIndex, depth
							if !_rules[rulematcherPart]() {
								goto l120
							}
							if !_rules[rule_]() {
								goto l120
							}
							if buffer[position] != rune('=') {
								goto l120
							}
							position++
							if !_rules[rule_]() {
								goto l120
							}
							if !_rules[ruleliteralString]() {
								goto l120
							}
							{
								add(ruleAction7, position)
							}
							goto l119
						l120:
							position, tokenIndex, depth = position119, tokenIndex119, depth119
							if !_rules[rulematcherPart]() {
								goto l122
							}
							if !_rules[rule_]() {
								goto l122
							}
							if buffer[position] != rune('!') {
								goto l122
							}
							position++
							if buffer[position] != rune('=') {
								goto l122
							}
							position++
							if !_rules[rule_]() {
								goto l122
							}
							if !_rules[ruleliteralString]() {
								goto l122
							}
							{
								add(ruleAction8, position)
							}
							goto l119
						l122:
							position, tokenIndex, depth = position119, tokenIndex119, depth119
							if !_rules[rulematcherPart]() {
								goto l124
							}
							if !_rules[rule_]() {
								goto l124
							}
							{
								position125, tokenIndex125, depth125 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l126
								}
								position++
								goto l125
							l126:
								position, tokenIndex, depth = position125, tokenIndex125, depth125
								if buffer[position] != rune('M') {
									goto l124
								}
								position++
							}
						l125:
							{
								position127, tokenIndex127, depth127 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l128
								}
								position++
								goto l127
							l128:
								position, tokenIndex, depth = position127, tokenIndex127, depth127
								if buffer[position] != rune('A') {
									goto l124
								}
								position++
							}
						l127:
							{
								position129, tokenIndex129, depth129 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l130
								}
								position++
								goto l129
							l130:
								position, tokenIndex, depth = position129, tokenIndex129, depth129
								if buffer[position] != rune('T') {
									goto l124
								}
								position++
							}
						l129:
							{
								position131, tokenIndex131, depth131 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l132
								}
								position++
								goto l131
							l132:
								position, tokenIndex, depth = position131, tokenIndex131, depth131
								if buffer[position] != rune('C') {
									goto l124
								}
								position++
							}
						l131:
							{
								position133, tokenIndex133, depth133 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l134
								}
								position++
								goto l133
							l134:
								position, tokenIndex, depth = position133, tokenIndex133, depth133
								if buffer[position] != rune('H') {
									goto l124
								}
								position++
							}
						l133:
							{
								position135, tokenIndex135, depth135 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l136
								}
								position++
								goto l135
							l136:
								position, tokenIndex, depth = position135, tokenIndex135, depth135
								if buffer[position] != rune('E') {
									goto l124
								}
								position++
							}
						l135:
							{
								position137, tokenIndex137, depth137 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l138
								}
								position++
								goto l137
							l138:
								position, tokenIndex, depth = position137, tokenIndex137, depth137
								if buffer[position] != rune('S') {
									goto l124
								}
								position++
							}
						l137:
							if !_rules[rule_]() {
								goto l124
							}
							if !_rules[ruleliteralString]() {
								goto l124
							}
							{
								add(ruleAction9, position)
							}
							goto l119
						l124:
							position, tokenIndex, depth = position119, tokenIndex119, depth119
							if !_rules[rulematcherPart]() {
								goto l106
							}
							if !_rules[rule_]() {
								goto l106
							}
							{
								position140, tokenIndex140, depth140 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l141
								}
								position++
								goto l140
							l141:
								position, tokenIndex, depth = position140, tokenIndex140, depth140
								if buffer[position] != rune('I') {
									goto l106
								}
								position++
							}
						l140:
							{
								position142, tokenIndex142, depth142 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l143
								}
								position++
								goto l142
							l143:
								position, tokenIndex, depth = position142, tokenIndex142, depth142
								if buffer[position] != rune('N') {
									goto l106
								}
								position++
							}
						l142:
							if !_rules[rule_]() {
								goto l106
							}
							{
								position144 := position
								depth++
								{
									add(ruleAction12, position)
								}
								if buffer[position] != rune('(') {
									goto l106
								}
								position++
								if !_rules[rule_]() {
									goto l106
								}
								{
									position146 := position
									depth++
									if !_rules[ruleliteralListString]() {
										goto l106
									}
									depth--
									add(rulePegText, position146)
								}
							l147:
								{
									position148, tokenIndex148, depth148 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l148
									}
									if buffer[position] != rune(',') {
										goto l148
									}
									position++
									if !_rules[rule_]() {
										goto l148
									}
									{
										position149 := position
										depth++
										if !_rules[ruleliteralListString]() {
											goto l148
										}
										depth--
										add(rulePegText, position149)
									}
									goto l147
								l148:
									position, tokenIndex, depth = position148, tokenIndex148, depth148
								}
								if !_rules[rule_]() {
									goto l106
								}
								if buffer[position] != rune(')') {
									goto l106
								}
								position++
								depth--
								add(ruleliteralList, position144)
							}
							{
								add(ruleAction10, position)
							}
						}
					l119:
						depth--
						add(ruletagMatcher, position118)
					}
				}
			l108:
				depth--
				add(rulepredicate_3, position107)
			}
			return true
		l106:
			position, tokenIndex, depth = position106, tokenIndex106, depth106
			return false
		},
		/* 9 tagMatcher <- <((matcherPart _ '=' _ literalString Action7) / (matcherPart _ ('!' '=') _ literalString Action8) / (matcherPart _ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) _ literalString Action9) / (matcherPart _ (('i' / 'I') ('n' / 'N')) _ literalList Action10))> */
		nil,
		/* 10 literalString <- <(<STRING> Action11)> */
		func() bool {
			position152, tokenIndex152, depth152 := position, tokenIndex, depth
			{
				position153 := position
				depth++
				{
					position154 := position
					depth++
					if !_rules[ruleSTRING]() {
						goto l152
					}
					depth--
					add(rulePegText, position154)
				}
				{
					add(ruleAction11, position)
				}
				depth--
				add(ruleliteralString, position153)
			}
			return true
		l152:
			position, tokenIndex, depth = position152, tokenIndex152, depth152
			return false
		},
		/* 11 literalList <- <(Action12 '(' _ <literalListString> (_ ',' _ <literalListString>)* _ ')')> */
		nil,
		/* 12 literalListString <- <(STRING Action13)> */
		func() bool {
			position157, tokenIndex157, depth157 := position, tokenIndex, depth
			{
				position158 := position
				depth++
				if !_rules[ruleSTRING]() {
					goto l157
				}
				{
					add(ruleAction13, position)
				}
				depth--
				add(ruleliteralListString, position158)
			}
			return true
		l157:
			position, tokenIndex, depth = position157, tokenIndex157, depth157
			return false
		},
		/* 13 matcherPart <- <(Action14 (<ALIAS_NAME> Action15 _ ':' _)? <TAG_NAME> Action16)> */
		func() bool {
			position160, tokenIndex160, depth160 := position, tokenIndex, depth
			{
				position161 := position
				depth++
				{
					add(ruleAction14, position)
				}
				{
					position163, tokenIndex163, depth163 := position, tokenIndex, depth
					{
						position165 := position
						depth++
						{
							position166 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l163
							}
							depth--
							add(ruleALIAS_NAME, position166)
						}
						depth--
						add(rulePegText, position165)
					}
					{
						add(ruleAction15, position)
					}
					if !_rules[rule_]() {
						goto l163
					}
					if buffer[position] != rune(':') {
						goto l163
					}
					position++
					if !_rules[rule_]() {
						goto l163
					}
					goto l164
				l163:
					position, tokenIndex, depth = position163, tokenIndex163, depth163
				}
			l164:
				{
					position168 := position
					depth++
					{
						position169 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l160
						}
						depth--
						add(ruleTAG_NAME, position169)
					}
					depth--
					add(rulePegText, position168)
				}
				{
					add(ruleAction16, position)
				}
				depth--
				add(rulematcherPart, position161)
			}
			return true
		l160:
			position, tokenIndex, depth = position160, tokenIndex160, depth160
			return false
		},
		/* 14 ALIAS_NAME <- <IDENTIFIER> */
		nil,
		/* 15 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 16 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 17 IDENTIFIER <- <((ID_SEGMENT ('.' ID_SEGMENT)*) / ('`' CHAR* '`'))> */
		func() bool {
			position174, tokenIndex174, depth174 := position, tokenIndex, depth
			{
				position175 := position
				depth++
				{
					position176, tokenIndex176, depth176 := position, tokenIndex, depth
					if !_rules[ruleID_SEGMENT]() {
						goto l177
					}
				l178:
					{
						position179, tokenIndex179, depth179 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l179
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l179
						}
						goto l178
					l179:
						position, tokenIndex, depth = position179, tokenIndex179, depth179
					}
					goto l176
				l177:
					position, tokenIndex, depth = position176, tokenIndex176, depth176
					if buffer[position] != rune('`') {
						goto l174
					}
					position++
				l180:
					{
						position181, tokenIndex181, depth181 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l181
						}
						goto l180
					l181:
						position, tokenIndex, depth = position181, tokenIndex181, depth181
					}
					if buffer[position] != rune('`') {
						goto l174
					}
					position++
				}
			l176:
				depth--
				add(ruleIDENTIFIER, position175)
			}
			return true
		l174:
			position, tokenIndex, depth = position174, tokenIndex174, depth174
			return false
		},
		/* 18 ID_SEGMENT <- <(ID_START ID_CONT*)> */
		func() bool {
			position182, tokenIndex182, depth182 := position, tokenIndex, depth
			{
				position183 := position
				depth++
				if !_rules[ruleID_START]() {
					goto l182
				}
			l184:
				{
					position185, tokenIndex185, depth185 := position, tokenIndex, depth
					{
						position186 := position
						depth++
						{
							position187, tokenIndex187, depth187 := position, tokenIndex, depth
							if !_rules[ruleID_START]() {
								goto l188
							}
							goto l187
						l188:
							position, tokenIndex, depth = position187, tokenIndex187, depth187
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l185
							}
							position++
						}
					l187:
						depth--
						add(ruleID_CONT, position186)
					}
					goto l184
				l185:
					position, tokenIndex, depth = position185, tokenIndex185, depth185
				}
				depth--
				add(ruleID_SEGMENT, position183)
			}
			return true
		l182:
			position, tokenIndex, depth = position182, tokenIndex182, depth182
			return false
		},
		/* 19 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position189, tokenIndex189, depth189 := position, tokenIndex, depth
			{
				position190 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l189
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l189
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l189
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position190)
			}
			return true
		l189:
			position, tokenIndex, depth = position189, tokenIndex189, depth189
			return false
		},
		/* 20 ID_CONT <- <(ID_START / [0-9])> */
		nil,
		/* 21 STRING <- <(('\'' CHAR* '\'') / ('"' CHAR* '"'))> */
		func() bool {
			position193, tokenIndex193, depth193 := position, tokenIndex, depth
			{
				position194 := position
				depth++
				{
					position195, tokenIndex195, depth195 := position, tokenIndex, depth
					if buffer[position] != rune('\'') {
						goto l196
					}
					position++
				l197:
					{
						position198, tokenIndex198, depth198 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l198
						}
						goto l197
					l198:
						position, tokenIndex, depth = position198, tokenIndex198, depth198
					}
					if buffer[position] != rune('\'') {
						goto l196
					}
					position++
					goto l195
				l196:
					position, tokenIndex, depth = position195, tokenIndex195, depth195
					if buffer[position] != rune('"') {
						goto l193
					}
					position++
				l199:
					{
						position200, tokenIndex200, depth200 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l200
						}
						goto l199
					l200:
						position, tokenIndex, depth = position200, tokenIndex200, depth200
					}
					if buffer[position] != rune('"') {
						goto l193
					}
					position++
				}
			l195:
				depth--
				add(ruleSTRING, position194)
			}
			return true
		l193:
			position, tokenIndex, depth = position193, tokenIndex193, depth193
			return false
		},
		/* 22 CHAR <- <(('\\' ((&('\\') '\\') | (&('"') '"') | (&('`') '`') | (&('\'') '\''))) / (!'\'' .))> */
		func() bool {
			position201, tokenIndex201, depth201 := position, tokenIndex, depth
			{
				position202 := position
				depth++
				{
					position203, tokenIndex203, depth203 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l204
					}
					position++
					{
						switch buffer[position] {
						case '\\':
							if buffer[position] != rune('\\') {
								goto l204
							}
							position++
							break
						case '"':
							if buffer[position] != rune('"') {
								goto l204
							}
							position++
							break
						case '`':
							if buffer[position] != rune('`') {
								goto l204
							}
							position++
							break
						default:
							if buffer[position] != rune('\'') {
								goto l204
							}
							position++
							break
						}
					}

					goto l203
				l204:
					position, tokenIndex, depth = position203, tokenIndex203, depth203
					{
						position206, tokenIndex206, depth206 := position, tokenIndex, depth
						if buffer[position] != rune('\'') {
							goto l206
						}
						position++
						goto l201
					l206:
						position, tokenIndex, depth = position206, tokenIndex206, depth206
					}
					if !matchDot() {
						goto l201
					}
				}
			l203:
				depth--
				add(ruleCHAR, position202)
			}
			return true
		l201:
			position, tokenIndex, depth = position201, tokenIndex201, depth201
			return false
		},
		/* 23 __ <- <' '+> */
		func() bool {
			position207, tokenIndex207, depth207 := position, tokenIndex, depth
			{
				position208 := position
				depth++
				if buffer[position] != rune(' ') {
					goto l207
				}
				position++
			l209:
				{
					position210, tokenIndex210, depth210 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l210
					}
					position++
					goto l209
				l210:
					position, tokenIndex, depth = position210, tokenIndex210, depth210
				}
				depth--
				add(rule__, position208)
			}
			return true
		l207:
			position, tokenIndex, depth = position207, tokenIndex207, depth207
			return false
		},
		/* 24 _ <- <' '*> */
		func() bool {
			{
				position212 := position
				depth++
			l213:
				{
					position214, tokenIndex214, depth214 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l214
					}
					position++
					goto l213
				l214:
					position, tokenIndex, depth = position214, tokenIndex214, depth214
				}
				depth--
				add(rule_, position212)
			}
			return true
		},
		/* 26 Action0 <- <{ p.makeDescribeAll() }> */
		nil,
		nil,
		/* 28 Action1 <- <{ p.addLiteralNode(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 29 Action2 <- <{ p.addNullPredicate() }> */
		nil,
		/* 30 Action3 <- <{ p.makeDescribe() }> */
		nil,
		/* 31 Action4 <- <{ p.addAndMatcher() }> */
		nil,
		/* 32 Action5 <- <{ p.addOrMatcher() }> */
		nil,
		/* 33 Action6 <- <{ p.addNotMatcher() }> */
		nil,
		/* 34 Action7 <- <{
		   p.addLiteralMatcher()
		 }> */
		nil,
		/* 35 Action8 <- <{
		   p.addLiteralMatcher()
		   p.addNotMatcher()
		 }> */
		nil,
		/* 36 Action9 <- <{
		   p.addRegexMatcher()
		 }> */
		nil,
		/* 37 Action10 <- <{
		   p.addListMatcher()
		 }> */
		nil,
		/* 38 Action11 <- <{
		  p.addLiteralNode(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 39 Action12 <- <{ p.addLiteralListNode() }> */
		nil,
		/* 40 Action13 <- <{
		  p.appendLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 41 Action14 <- <{ p.addTagRefNode() }> */
		nil,
		/* 42 Action15 <- <{ p.setAlias(buffer[begin:end]) }> */
		nil,
		/* 43 Action16 <- <{ p.setTag(buffer[begin:end]) }> */
		nil,
	}
	p.rules = _rules
}
