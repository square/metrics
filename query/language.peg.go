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
	ruleexpression_aggregate
	ruleexpression_function
	ruleexpression_metric
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
	ruleAction1
	rulePegText
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
	ruleAction18
	ruleAction19
	ruleAction20
	ruleAction21
	ruleAction22
	ruleAction23
	ruleAction24
	ruleAction25
	ruleAction26
	ruleAction27
	ruleAction28
	ruleAction29
	ruleAction30

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
	"expression_aggregate",
	"expression_function",
	"expression_metric",
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
	"Action1",
	"PegText",
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
	"Action18",
	"Action19",
	"Action20",
	"Action21",
	"Action22",
	"Action23",
	"Action24",
	"Action25",
	"Action26",
	"Action27",
	"Action28",
	"Action29",
	"Action30",

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
	// ===================

	// stack of nodes used during the AST traversal.
	// a non-empty stack at the finish implies a programming error.
	nodeStack []Node

	// user errors accumulated during the AST traversal.
	// a non-empty list at the finish time means an invalid query is provided.
	errors []SyntaxError

	// programming errors accumulated during the AST traversal.
	// a non-empty list at the finish time implies a programming error.
	assertions []error

	// final result
	command Command

	Buffer string
	buffer []rune
	rules  [87]func() bool
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

			p.makeSelect()

		case ruleAction1:
			p.makeDescribeAll()
		case ruleAction2:
			p.addLiteralNode(unescapeLiteral(buffer[begin:end]))
		case ruleAction3:
			p.makeDescribe()
		case ruleAction4:
			p.addNullPredicate()
		case ruleAction5:
			p.addExpressionList()
		case ruleAction6:
			p.appendExpression()
		case ruleAction7:
			p.appendExpression()
		case ruleAction8:
			p.addOperatorLiteralNode("*")
		case ruleAction9:
			p.addOperatorLiteralNode("-")
		case ruleAction10:
			p.addOperatorExpressionNode()
		case ruleAction11:
			p.addOperatorLiteralNode("*")
		case ruleAction12:
			p.addOperatorLiteralNode("*")
		case ruleAction13:
			p.addOperatorExpressionNode()
		case ruleAction14:
			p.addNumberNode(buffer[begin:end])
		case ruleAction15:

			p.addLiteralNode(unescapeLiteral(buffer[begin:end]))

		case ruleAction16:

			p.addLiteralNode(unescapeLiteral(buffer[begin:end]))

		case ruleAction17:

			p.addLiteralNode(unescapeLiteral(buffer[begin:end]))

		case ruleAction18:
			p.addNullPredicate()
		case ruleAction19:

			p.addMetricReferenceNode()

		case ruleAction20:
			p.addAndPredicate()
		case ruleAction21:
			p.addOrPredicate()
		case ruleAction22:
			p.addNotPredicate()
		case ruleAction23:

			p.addLiteralMatcher()

		case ruleAction24:

			p.addLiteralMatcher()
			p.addNotPredicate()

		case ruleAction25:

			p.addRegexMatcher()

		case ruleAction26:

			p.addListMatcher()

		case ruleAction27:

			p.addLiteralNode(unescapeLiteral(buffer[begin:end]))

		case ruleAction28:
			p.addLiteralListNode()
		case ruleAction29:

			p.appendLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction30:
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
						{
							add(ruleAction0, position)
						}
						depth--
						add(ruleselectStmt, position4)
					}
					goto l2
				l3:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
					{
						position31 := position
						depth++
						{
							position32, tokenIndex32, depth32 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l33
							}
							position++
							goto l32
						l33:
							position, tokenIndex, depth = position32, tokenIndex32, depth32
							if buffer[position] != rune('D') {
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
						{
							position36, tokenIndex36, depth36 := position, tokenIndex, depth
							if buffer[position] != rune('s') {
								goto l37
							}
							position++
							goto l36
						l37:
							position, tokenIndex, depth = position36, tokenIndex36, depth36
							if buffer[position] != rune('S') {
								goto l0
							}
							position++
						}
					l36:
						{
							position38, tokenIndex38, depth38 := position, tokenIndex, depth
							if buffer[position] != rune('c') {
								goto l39
							}
							position++
							goto l38
						l39:
							position, tokenIndex, depth = position38, tokenIndex38, depth38
							if buffer[position] != rune('C') {
								goto l0
							}
							position++
						}
					l38:
						{
							position40, tokenIndex40, depth40 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l41
							}
							position++
							goto l40
						l41:
							position, tokenIndex, depth = position40, tokenIndex40, depth40
							if buffer[position] != rune('R') {
								goto l0
							}
							position++
						}
					l40:
						{
							position42, tokenIndex42, depth42 := position, tokenIndex, depth
							if buffer[position] != rune('i') {
								goto l43
							}
							position++
							goto l42
						l43:
							position, tokenIndex, depth = position42, tokenIndex42, depth42
							if buffer[position] != rune('I') {
								goto l0
							}
							position++
						}
					l42:
						{
							position44, tokenIndex44, depth44 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l45
							}
							position++
							goto l44
						l45:
							position, tokenIndex, depth = position44, tokenIndex44, depth44
							if buffer[position] != rune('B') {
								goto l0
							}
							position++
						}
					l44:
						{
							position46, tokenIndex46, depth46 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l47
							}
							position++
							goto l46
						l47:
							position, tokenIndex, depth = position46, tokenIndex46, depth46
							if buffer[position] != rune('E') {
								goto l0
							}
							position++
						}
					l46:
						if !_rules[rule__]() {
							goto l0
						}
						{
							position48, tokenIndex48, depth48 := position, tokenIndex, depth
							{
								position50 := position
								depth++
								{
									position51, tokenIndex51, depth51 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l52
									}
									position++
									goto l51
								l52:
									position, tokenIndex, depth = position51, tokenIndex51, depth51
									if buffer[position] != rune('A') {
										goto l49
									}
									position++
								}
							l51:
								{
									position53, tokenIndex53, depth53 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l54
									}
									position++
									goto l53
								l54:
									position, tokenIndex, depth = position53, tokenIndex53, depth53
									if buffer[position] != rune('L') {
										goto l49
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
										goto l49
									}
									position++
								}
							l55:
								{
									add(ruleAction1, position)
								}
								depth--
								add(ruledescribeAllStmt, position50)
							}
							goto l48
						l49:
							position, tokenIndex, depth = position48, tokenIndex48, depth48
							{
								position58 := position
								depth++
								{
									position59 := position
									depth++
									{
										position60 := position
										depth++
										if !_rules[ruleIDENTIFIER]() {
											goto l0
										}
										depth--
										add(ruleMETRIC_NAME, position60)
									}
									depth--
									add(rulePegText, position59)
								}
								{
									add(ruleAction2, position)
								}
								if !_rules[ruleoptionalPredicateClause]() {
									goto l0
								}
								{
									add(ruleAction3, position)
								}
								depth--
								add(ruledescribeSingleStmt, position58)
							}
						}
					l48:
						depth--
						add(ruledescribeStmt, position31)
					}
				}
			l2:
				{
					position63, tokenIndex63, depth63 := position, tokenIndex, depth
					if !matchDot() {
						goto l63
					}
					goto l0
				l63:
					position, tokenIndex, depth = position63, tokenIndex63, depth63
				}
				depth--
				add(ruleroot, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 selectStmt <- <(('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T') __ expressionList optionalPredicateClause rangeClause Action0)> */
		nil,
		/* 2 describeStmt <- <(('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E') __ (describeAllStmt / describeSingleStmt))> */
		nil,
		/* 3 describeAllStmt <- <(('a' / 'A') ('l' / 'L') ('l' / 'L') Action1)> */
		nil,
		/* 4 describeSingleStmt <- <(<METRIC_NAME> Action2 optionalPredicateClause Action3)> */
		nil,
		/* 5 rangeClause <- <(_ (('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M')) __ TIMESTAMP __ (('t' / 'T') ('o' / 'O')) __ TIMESTAMP)> */
		nil,
		/* 6 optionalPredicateClause <- <((__ predicateClause) / Action4)> */
		func() bool {
			{
				position70 := position
				depth++
				{
					position71, tokenIndex71, depth71 := position, tokenIndex, depth
					if !_rules[rule__]() {
						goto l72
					}
					{
						position73 := position
						depth++
						{
							position74, tokenIndex74, depth74 := position, tokenIndex, depth
							if buffer[position] != rune('w') {
								goto l75
							}
							position++
							goto l74
						l75:
							position, tokenIndex, depth = position74, tokenIndex74, depth74
							if buffer[position] != rune('W') {
								goto l72
							}
							position++
						}
					l74:
						{
							position76, tokenIndex76, depth76 := position, tokenIndex, depth
							if buffer[position] != rune('h') {
								goto l77
							}
							position++
							goto l76
						l77:
							position, tokenIndex, depth = position76, tokenIndex76, depth76
							if buffer[position] != rune('H') {
								goto l72
							}
							position++
						}
					l76:
						{
							position78, tokenIndex78, depth78 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l79
							}
							position++
							goto l78
						l79:
							position, tokenIndex, depth = position78, tokenIndex78, depth78
							if buffer[position] != rune('E') {
								goto l72
							}
							position++
						}
					l78:
						{
							position80, tokenIndex80, depth80 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l81
							}
							position++
							goto l80
						l81:
							position, tokenIndex, depth = position80, tokenIndex80, depth80
							if buffer[position] != rune('R') {
								goto l72
							}
							position++
						}
					l80:
						{
							position82, tokenIndex82, depth82 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l83
							}
							position++
							goto l82
						l83:
							position, tokenIndex, depth = position82, tokenIndex82, depth82
							if buffer[position] != rune('E') {
								goto l72
							}
							position++
						}
					l82:
						if !_rules[rule__]() {
							goto l72
						}
						if !_rules[rulepredicate_1]() {
							goto l72
						}
						depth--
						add(rulepredicateClause, position73)
					}
					goto l71
				l72:
					position, tokenIndex, depth = position71, tokenIndex71, depth71
					{
						add(ruleAction4, position)
					}
				}
			l71:
				depth--
				add(ruleoptionalPredicateClause, position70)
			}
			return true
		},
		/* 7 expressionList <- <(Action5 expression_1 Action6 (COMMA expression_1 Action7)*)> */
		func() bool {
			position85, tokenIndex85, depth85 := position, tokenIndex, depth
			{
				position86 := position
				depth++
				{
					add(ruleAction5, position)
				}
				if !_rules[ruleexpression_1]() {
					goto l85
				}
				{
					add(ruleAction6, position)
				}
			l89:
				{
					position90, tokenIndex90, depth90 := position, tokenIndex, depth
					if !_rules[ruleCOMMA]() {
						goto l90
					}
					if !_rules[ruleexpression_1]() {
						goto l90
					}
					{
						add(ruleAction7, position)
					}
					goto l89
				l90:
					position, tokenIndex, depth = position90, tokenIndex90, depth90
				}
				depth--
				add(ruleexpressionList, position86)
			}
			return true
		l85:
			position, tokenIndex, depth = position85, tokenIndex85, depth85
			return false
		},
		/* 8 expression_1 <- <(expression_2 (((OP_ADD Action8) / (OP_SUB Action9)) expression_2 Action10)*)> */
		func() bool {
			position92, tokenIndex92, depth92 := position, tokenIndex, depth
			{
				position93 := position
				depth++
				if !_rules[ruleexpression_2]() {
					goto l92
				}
			l94:
				{
					position95, tokenIndex95, depth95 := position, tokenIndex, depth
					{
						position96, tokenIndex96, depth96 := position, tokenIndex, depth
						{
							position98 := position
							depth++
							if !_rules[rule_]() {
								goto l97
							}
							if buffer[position] != rune('+') {
								goto l97
							}
							position++
							if !_rules[rule_]() {
								goto l97
							}
							depth--
							add(ruleOP_ADD, position98)
						}
						{
							add(ruleAction8, position)
						}
						goto l96
					l97:
						position, tokenIndex, depth = position96, tokenIndex96, depth96
						{
							position100 := position
							depth++
							if !_rules[rule_]() {
								goto l95
							}
							if buffer[position] != rune('-') {
								goto l95
							}
							position++
							if !_rules[rule_]() {
								goto l95
							}
							depth--
							add(ruleOP_SUB, position100)
						}
						{
							add(ruleAction9, position)
						}
					}
				l96:
					if !_rules[ruleexpression_2]() {
						goto l95
					}
					{
						add(ruleAction10, position)
					}
					goto l94
				l95:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
				}
				depth--
				add(ruleexpression_1, position93)
			}
			return true
		l92:
			position, tokenIndex, depth = position92, tokenIndex92, depth92
			return false
		},
		/* 9 expression_2 <- <(expression_3 (((OP_DIV Action11) / (OP_MULT Action12)) expression_3 Action13)*)> */
		func() bool {
			position103, tokenIndex103, depth103 := position, tokenIndex, depth
			{
				position104 := position
				depth++
				if !_rules[ruleexpression_3]() {
					goto l103
				}
			l105:
				{
					position106, tokenIndex106, depth106 := position, tokenIndex, depth
					{
						position107, tokenIndex107, depth107 := position, tokenIndex, depth
						{
							position109 := position
							depth++
							if !_rules[rule_]() {
								goto l108
							}
							if buffer[position] != rune('/') {
								goto l108
							}
							position++
							if !_rules[rule_]() {
								goto l108
							}
							depth--
							add(ruleOP_DIV, position109)
						}
						{
							add(ruleAction11, position)
						}
						goto l107
					l108:
						position, tokenIndex, depth = position107, tokenIndex107, depth107
						{
							position111 := position
							depth++
							if !_rules[rule_]() {
								goto l106
							}
							if buffer[position] != rune('*') {
								goto l106
							}
							position++
							if !_rules[rule_]() {
								goto l106
							}
							depth--
							add(ruleOP_MULT, position111)
						}
						{
							add(ruleAction12, position)
						}
					}
				l107:
					if !_rules[ruleexpression_3]() {
						goto l106
					}
					{
						add(ruleAction13, position)
					}
					goto l105
				l106:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
				}
				depth--
				add(ruleexpression_2, position104)
			}
			return true
		l103:
			position, tokenIndex, depth = position103, tokenIndex103, depth103
			return false
		},
		/* 10 expression_3 <- <(expression_aggregate / expression_function / ((&('-' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') (<NUMBER> Action14)) | (&('\t' | '\n' | ' ' | '(') (PAREN_OPEN expression_1 PAREN_CLOSE)) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') expression_metric)))> */
		func() bool {
			position114, tokenIndex114, depth114 := position, tokenIndex, depth
			{
				position115 := position
				depth++
				{
					position116, tokenIndex116, depth116 := position, tokenIndex, depth
					{
						position118 := position
						depth++
						{
							position119 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l117
							}
							depth--
							add(rulePegText, position119)
						}
						{
							add(ruleAction15, position)
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l117
						}
						if !_rules[ruleexpression_1]() {
							goto l117
						}
						if !_rules[rule__]() {
							goto l117
						}
						{
							position121 := position
							depth++
							{
								position122, tokenIndex122, depth122 := position, tokenIndex, depth
								if buffer[position] != rune('g') {
									goto l123
								}
								position++
								goto l122
							l123:
								position, tokenIndex, depth = position122, tokenIndex122, depth122
								if buffer[position] != rune('G') {
									goto l117
								}
								position++
							}
						l122:
							{
								position124, tokenIndex124, depth124 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l125
								}
								position++
								goto l124
							l125:
								position, tokenIndex, depth = position124, tokenIndex124, depth124
								if buffer[position] != rune('R') {
									goto l117
								}
								position++
							}
						l124:
							{
								position126, tokenIndex126, depth126 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l127
								}
								position++
								goto l126
							l127:
								position, tokenIndex, depth = position126, tokenIndex126, depth126
								if buffer[position] != rune('O') {
									goto l117
								}
								position++
							}
						l126:
							{
								position128, tokenIndex128, depth128 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l129
								}
								position++
								goto l128
							l129:
								position, tokenIndex, depth = position128, tokenIndex128, depth128
								if buffer[position] != rune('U') {
									goto l117
								}
								position++
							}
						l128:
							{
								position130, tokenIndex130, depth130 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l131
								}
								position++
								goto l130
							l131:
								position, tokenIndex, depth = position130, tokenIndex130, depth130
								if buffer[position] != rune('P') {
									goto l117
								}
								position++
							}
						l130:
							if !_rules[rule__]() {
								goto l117
							}
							{
								position132, tokenIndex132, depth132 := position, tokenIndex, depth
								if buffer[position] != rune('b') {
									goto l133
								}
								position++
								goto l132
							l133:
								position, tokenIndex, depth = position132, tokenIndex132, depth132
								if buffer[position] != rune('B') {
									goto l117
								}
								position++
							}
						l132:
							{
								position134, tokenIndex134, depth134 := position, tokenIndex, depth
								if buffer[position] != rune('y') {
									goto l135
								}
								position++
								goto l134
							l135:
								position, tokenIndex, depth = position134, tokenIndex134, depth134
								if buffer[position] != rune('Y') {
									goto l117
								}
								position++
							}
						l134:
							if !_rules[rule__]() {
								goto l117
							}
							if !_rules[ruleCOLUMN_NAME]() {
								goto l117
							}
						l136:
							{
								position137, tokenIndex137, depth137 := position, tokenIndex, depth
								if !_rules[ruleCOMMA]() {
									goto l137
								}
								if !_rules[ruleCOLUMN_NAME]() {
									goto l137
								}
								goto l136
							l137:
								position, tokenIndex, depth = position137, tokenIndex137, depth137
							}
							depth--
							add(rulegroupByClause, position121)
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l117
						}
						depth--
						add(ruleexpression_aggregate, position118)
					}
					goto l116
				l117:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
					{
						position139 := position
						depth++
						{
							position140 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l138
							}
							depth--
							add(rulePegText, position140)
						}
						{
							add(ruleAction16, position)
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l138
						}
						if !_rules[ruleexpressionList]() {
							goto l138
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l138
						}
						depth--
						add(ruleexpression_function, position139)
					}
					goto l116
				l138:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
					{
						switch buffer[position] {
						case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							{
								position143 := position
								depth++
								{
									position144 := position
									depth++
									{
										position145 := position
										depth++
										{
											position146, tokenIndex146, depth146 := position, tokenIndex, depth
											if buffer[position] != rune('-') {
												goto l146
											}
											position++
											goto l147
										l146:
											position, tokenIndex, depth = position146, tokenIndex146, depth146
										}
									l147:
										if !_rules[ruleNUMBER_NATURAL]() {
											goto l114
										}
										depth--
										add(ruleNUMBER_INTEGER, position145)
									}
									{
										position148, tokenIndex148, depth148 := position, tokenIndex, depth
										{
											position150 := position
											depth++
											if buffer[position] != rune('.') {
												goto l148
											}
											position++
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l148
											}
											position++
										l151:
											{
												position152, tokenIndex152, depth152 := position, tokenIndex, depth
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l152
												}
												position++
												goto l151
											l152:
												position, tokenIndex, depth = position152, tokenIndex152, depth152
											}
											depth--
											add(ruleNUMBER_FRACTION, position150)
										}
										goto l149
									l148:
										position, tokenIndex, depth = position148, tokenIndex148, depth148
									}
								l149:
									{
										position153, tokenIndex153, depth153 := position, tokenIndex, depth
										{
											position155 := position
											depth++
											{
												position156, tokenIndex156, depth156 := position, tokenIndex, depth
												if buffer[position] != rune('e') {
													goto l157
												}
												position++
												goto l156
											l157:
												position, tokenIndex, depth = position156, tokenIndex156, depth156
												if buffer[position] != rune('E') {
													goto l153
												}
												position++
											}
										l156:
											{
												position158, tokenIndex158, depth158 := position, tokenIndex, depth
												{
													position160, tokenIndex160, depth160 := position, tokenIndex, depth
													if buffer[position] != rune('+') {
														goto l161
													}
													position++
													goto l160
												l161:
													position, tokenIndex, depth = position160, tokenIndex160, depth160
													if buffer[position] != rune('-') {
														goto l158
													}
													position++
												}
											l160:
												goto l159
											l158:
												position, tokenIndex, depth = position158, tokenIndex158, depth158
											}
										l159:
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l153
											}
											position++
										l162:
											{
												position163, tokenIndex163, depth163 := position, tokenIndex, depth
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l163
												}
												position++
												goto l162
											l163:
												position, tokenIndex, depth = position163, tokenIndex163, depth163
											}
											depth--
											add(ruleNUMBER_EXP, position155)
										}
										goto l154
									l153:
										position, tokenIndex, depth = position153, tokenIndex153, depth153
									}
								l154:
									depth--
									add(ruleNUMBER, position144)
								}
								depth--
								add(rulePegText, position143)
							}
							{
								add(ruleAction14, position)
							}
							break
						case '\t', '\n', ' ', '(':
							if !_rules[rulePAREN_OPEN]() {
								goto l114
							}
							if !_rules[ruleexpression_1]() {
								goto l114
							}
							if !_rules[rulePAREN_CLOSE]() {
								goto l114
							}
							break
						default:
							{
								position165 := position
								depth++
								{
									position166 := position
									depth++
									if !_rules[ruleIDENTIFIER]() {
										goto l114
									}
									depth--
									add(rulePegText, position166)
								}
								{
									add(ruleAction17, position)
								}
								{
									position168, tokenIndex168, depth168 := position, tokenIndex, depth
									{
										position170, tokenIndex170, depth170 := position, tokenIndex, depth
										if buffer[position] != rune('[') {
											goto l171
										}
										position++
										if !_rules[rule_]() {
											goto l171
										}
										if !_rules[rulepredicate_1]() {
											goto l171
										}
										if !_rules[rule_]() {
											goto l171
										}
										if buffer[position] != rune(']') {
											goto l171
										}
										position++
										goto l170
									l171:
										position, tokenIndex, depth = position170, tokenIndex170, depth170
										{
											add(ruleAction18, position)
										}
									}
								l170:
									goto l169

									position, tokenIndex, depth = position168, tokenIndex168, depth168
								}
							l169:
								{
									add(ruleAction19, position)
								}
								depth--
								add(ruleexpression_metric, position165)
							}
							break
						}
					}

				}
			l116:
				depth--
				add(ruleexpression_3, position115)
			}
			return true
		l114:
			position, tokenIndex, depth = position114, tokenIndex114, depth114
			return false
		},
		/* 11 expression_aggregate <- <(<IDENTIFIER> Action15 PAREN_OPEN expression_1 __ groupByClause PAREN_CLOSE)> */
		nil,
		/* 12 expression_function <- <(<IDENTIFIER> Action16 PAREN_OPEN expressionList PAREN_CLOSE)> */
		nil,
		/* 13 expression_metric <- <(<IDENTIFIER> Action17 (('[' _ predicate_1 _ ']') / Action18)? Action19)> */
		nil,
		/* 14 groupByClause <- <(('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P') __ (('b' / 'B') ('y' / 'Y')) __ COLUMN_NAME (COMMA COLUMN_NAME)*)> */
		nil,
		/* 15 predicateClause <- <(('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E') __ predicate_1)> */
		nil,
		/* 16 predicate_1 <- <((predicate_2 OP_AND predicate_1 Action20) / predicate_2 / )> */
		func() bool {
			{
				position180 := position
				depth++
				{
					position181, tokenIndex181, depth181 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l182
					}
					{
						position183 := position
						depth++
						if !_rules[rule__]() {
							goto l182
						}
						{
							position184, tokenIndex184, depth184 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l185
							}
							position++
							goto l184
						l185:
							position, tokenIndex, depth = position184, tokenIndex184, depth184
							if buffer[position] != rune('A') {
								goto l182
							}
							position++
						}
					l184:
						{
							position186, tokenIndex186, depth186 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l187
							}
							position++
							goto l186
						l187:
							position, tokenIndex, depth = position186, tokenIndex186, depth186
							if buffer[position] != rune('N') {
								goto l182
							}
							position++
						}
					l186:
						{
							position188, tokenIndex188, depth188 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l189
							}
							position++
							goto l188
						l189:
							position, tokenIndex, depth = position188, tokenIndex188, depth188
							if buffer[position] != rune('D') {
								goto l182
							}
							position++
						}
					l188:
						if !_rules[rule__]() {
							goto l182
						}
						depth--
						add(ruleOP_AND, position183)
					}
					if !_rules[rulepredicate_1]() {
						goto l182
					}
					{
						add(ruleAction20, position)
					}
					goto l181
				l182:
					position, tokenIndex, depth = position181, tokenIndex181, depth181
					if !_rules[rulepredicate_2]() {
						goto l191
					}
					goto l181
				l191:
					position, tokenIndex, depth = position181, tokenIndex181, depth181
				}
			l181:
				depth--
				add(rulepredicate_1, position180)
			}
			return true
		},
		/* 17 predicate_2 <- <((predicate_3 OP_OR predicate_2 Action21) / predicate_3)> */
		func() bool {
			position192, tokenIndex192, depth192 := position, tokenIndex, depth
			{
				position193 := position
				depth++
				{
					position194, tokenIndex194, depth194 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l195
					}
					{
						position196 := position
						depth++
						if !_rules[rule__]() {
							goto l195
						}
						{
							position197, tokenIndex197, depth197 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l198
							}
							position++
							goto l197
						l198:
							position, tokenIndex, depth = position197, tokenIndex197, depth197
							if buffer[position] != rune('O') {
								goto l195
							}
							position++
						}
					l197:
						{
							position199, tokenIndex199, depth199 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l200
							}
							position++
							goto l199
						l200:
							position, tokenIndex, depth = position199, tokenIndex199, depth199
							if buffer[position] != rune('R') {
								goto l195
							}
							position++
						}
					l199:
						if !_rules[rule__]() {
							goto l195
						}
						depth--
						add(ruleOP_OR, position196)
					}
					if !_rules[rulepredicate_2]() {
						goto l195
					}
					{
						add(ruleAction21, position)
					}
					goto l194
				l195:
					position, tokenIndex, depth = position194, tokenIndex194, depth194
					if !_rules[rulepredicate_3]() {
						goto l192
					}
				}
			l194:
				depth--
				add(rulepredicate_2, position193)
			}
			return true
		l192:
			position, tokenIndex, depth = position192, tokenIndex192, depth192
			return false
		},
		/* 18 predicate_3 <- <((OP_NOT predicate_3 Action22) / (PAREN_OPEN predicate_1 PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position202, tokenIndex202, depth202 := position, tokenIndex, depth
			{
				position203 := position
				depth++
				{
					position204, tokenIndex204, depth204 := position, tokenIndex, depth
					{
						position206 := position
						depth++
						{
							position207, tokenIndex207, depth207 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l208
							}
							position++
							goto l207
						l208:
							position, tokenIndex, depth = position207, tokenIndex207, depth207
							if buffer[position] != rune('N') {
								goto l205
							}
							position++
						}
					l207:
						{
							position209, tokenIndex209, depth209 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l210
							}
							position++
							goto l209
						l210:
							position, tokenIndex, depth = position209, tokenIndex209, depth209
							if buffer[position] != rune('O') {
								goto l205
							}
							position++
						}
					l209:
						{
							position211, tokenIndex211, depth211 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l212
							}
							position++
							goto l211
						l212:
							position, tokenIndex, depth = position211, tokenIndex211, depth211
							if buffer[position] != rune('T') {
								goto l205
							}
							position++
						}
					l211:
						if !_rules[rule__]() {
							goto l205
						}
						depth--
						add(ruleOP_NOT, position206)
					}
					if !_rules[rulepredicate_3]() {
						goto l205
					}
					{
						add(ruleAction22, position)
					}
					goto l204
				l205:
					position, tokenIndex, depth = position204, tokenIndex204, depth204
					if !_rules[rulePAREN_OPEN]() {
						goto l214
					}
					if !_rules[rulepredicate_1]() {
						goto l214
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l214
					}
					goto l204
				l214:
					position, tokenIndex, depth = position204, tokenIndex204, depth204
					{
						position215 := position
						depth++
						{
							position216, tokenIndex216, depth216 := position, tokenIndex, depth
							if !_rules[ruletagName]() {
								goto l217
							}
							if !_rules[rule_]() {
								goto l217
							}
							if buffer[position] != rune('=') {
								goto l217
							}
							position++
							if !_rules[rule_]() {
								goto l217
							}
							if !_rules[ruleliteralString]() {
								goto l217
							}
							{
								add(ruleAction23, position)
							}
							goto l216
						l217:
							position, tokenIndex, depth = position216, tokenIndex216, depth216
							if !_rules[ruletagName]() {
								goto l219
							}
							if !_rules[rule_]() {
								goto l219
							}
							if buffer[position] != rune('!') {
								goto l219
							}
							position++
							if buffer[position] != rune('=') {
								goto l219
							}
							position++
							if !_rules[rule_]() {
								goto l219
							}
							if !_rules[ruleliteralString]() {
								goto l219
							}
							{
								add(ruleAction24, position)
							}
							goto l216
						l219:
							position, tokenIndex, depth = position216, tokenIndex216, depth216
							if !_rules[ruletagName]() {
								goto l221
							}
							if !_rules[rule__]() {
								goto l221
							}
							{
								position222, tokenIndex222, depth222 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l223
								}
								position++
								goto l222
							l223:
								position, tokenIndex, depth = position222, tokenIndex222, depth222
								if buffer[position] != rune('M') {
									goto l221
								}
								position++
							}
						l222:
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
									goto l221
								}
								position++
							}
						l224:
							{
								position226, tokenIndex226, depth226 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l227
								}
								position++
								goto l226
							l227:
								position, tokenIndex, depth = position226, tokenIndex226, depth226
								if buffer[position] != rune('T') {
									goto l221
								}
								position++
							}
						l226:
							{
								position228, tokenIndex228, depth228 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l229
								}
								position++
								goto l228
							l229:
								position, tokenIndex, depth = position228, tokenIndex228, depth228
								if buffer[position] != rune('C') {
									goto l221
								}
								position++
							}
						l228:
							{
								position230, tokenIndex230, depth230 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l231
								}
								position++
								goto l230
							l231:
								position, tokenIndex, depth = position230, tokenIndex230, depth230
								if buffer[position] != rune('H') {
									goto l221
								}
								position++
							}
						l230:
							{
								position232, tokenIndex232, depth232 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l233
								}
								position++
								goto l232
							l233:
								position, tokenIndex, depth = position232, tokenIndex232, depth232
								if buffer[position] != rune('E') {
									goto l221
								}
								position++
							}
						l232:
							{
								position234, tokenIndex234, depth234 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l235
								}
								position++
								goto l234
							l235:
								position, tokenIndex, depth = position234, tokenIndex234, depth234
								if buffer[position] != rune('S') {
									goto l221
								}
								position++
							}
						l234:
							if !_rules[rule__]() {
								goto l221
							}
							if !_rules[ruleliteralString]() {
								goto l221
							}
							{
								add(ruleAction25, position)
							}
							goto l216
						l221:
							position, tokenIndex, depth = position216, tokenIndex216, depth216
							if !_rules[ruletagName]() {
								goto l202
							}
							if !_rules[rule__]() {
								goto l202
							}
							{
								position237, tokenIndex237, depth237 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l238
								}
								position++
								goto l237
							l238:
								position, tokenIndex, depth = position237, tokenIndex237, depth237
								if buffer[position] != rune('I') {
									goto l202
								}
								position++
							}
						l237:
							{
								position239, tokenIndex239, depth239 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l240
								}
								position++
								goto l239
							l240:
								position, tokenIndex, depth = position239, tokenIndex239, depth239
								if buffer[position] != rune('N') {
									goto l202
								}
								position++
							}
						l239:
							if !_rules[rule__]() {
								goto l202
							}
							{
								position241 := position
								depth++
								{
									add(ruleAction28, position)
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l202
								}
								if !_rules[ruleliteralListString]() {
									goto l202
								}
							l243:
								{
									position244, tokenIndex244, depth244 := position, tokenIndex, depth
									if !_rules[ruleCOMMA]() {
										goto l244
									}
									if !_rules[ruleliteralListString]() {
										goto l244
									}
									goto l243
								l244:
									position, tokenIndex, depth = position244, tokenIndex244, depth244
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l202
								}
								depth--
								add(ruleliteralList, position241)
							}
							{
								add(ruleAction26, position)
							}
						}
					l216:
						depth--
						add(ruletagMatcher, position215)
					}
				}
			l204:
				depth--
				add(rulepredicate_3, position203)
			}
			return true
		l202:
			position, tokenIndex, depth = position202, tokenIndex202, depth202
			return false
		},
		/* 19 tagMatcher <- <((tagName _ '=' _ literalString Action23) / (tagName _ ('!' '=') _ literalString Action24) / (tagName __ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) __ literalString Action25) / (tagName __ (('i' / 'I') ('n' / 'N')) __ literalList Action26))> */
		nil,
		/* 20 literalString <- <(<STRING> Action27)> */
		func() bool {
			position247, tokenIndex247, depth247 := position, tokenIndex, depth
			{
				position248 := position
				depth++
				{
					position249 := position
					depth++
					if !_rules[ruleSTRING]() {
						goto l247
					}
					depth--
					add(rulePegText, position249)
				}
				{
					add(ruleAction27, position)
				}
				depth--
				add(ruleliteralString, position248)
			}
			return true
		l247:
			position, tokenIndex, depth = position247, tokenIndex247, depth247
			return false
		},
		/* 21 literalList <- <(Action28 PAREN_OPEN literalListString (COMMA literalListString)* PAREN_CLOSE)> */
		nil,
		/* 22 literalListString <- <(STRING Action29)> */
		func() bool {
			position252, tokenIndex252, depth252 := position, tokenIndex, depth
			{
				position253 := position
				depth++
				if !_rules[ruleSTRING]() {
					goto l252
				}
				{
					add(ruleAction29, position)
				}
				depth--
				add(ruleliteralListString, position253)
			}
			return true
		l252:
			position, tokenIndex, depth = position252, tokenIndex252, depth252
			return false
		},
		/* 23 tagName <- <(<TAG_NAME> Action30)> */
		func() bool {
			position255, tokenIndex255, depth255 := position, tokenIndex, depth
			{
				position256 := position
				depth++
				{
					position257 := position
					depth++
					{
						position258 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l255
						}
						depth--
						add(ruleTAG_NAME, position258)
					}
					depth--
					add(rulePegText, position257)
				}
				{
					add(ruleAction30, position)
				}
				depth--
				add(ruletagName, position256)
			}
			return true
		l255:
			position, tokenIndex, depth = position255, tokenIndex255, depth255
			return false
		},
		/* 24 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position260, tokenIndex260, depth260 := position, tokenIndex, depth
			{
				position261 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l260
				}
				depth--
				add(ruleCOLUMN_NAME, position261)
			}
			return true
		l260:
			position, tokenIndex, depth = position260, tokenIndex260, depth260
			return false
		},
		/* 25 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 26 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 27 TIMESTAMP <- <(NUMBER_NATURAL / STRING)> */
		func() bool {
			position264, tokenIndex264, depth264 := position, tokenIndex, depth
			{
				position265 := position
				depth++
				{
					position266, tokenIndex266, depth266 := position, tokenIndex, depth
					if !_rules[ruleNUMBER_NATURAL]() {
						goto l267
					}
					goto l266
				l267:
					position, tokenIndex, depth = position266, tokenIndex266, depth266
					if !_rules[ruleSTRING]() {
						goto l264
					}
				}
			l266:
				depth--
				add(ruleTIMESTAMP, position265)
			}
			return true
		l264:
			position, tokenIndex, depth = position264, tokenIndex264, depth264
			return false
		},
		/* 28 IDENTIFIER <- <(('`' CHAR* '`') / (!KEYWORD ID_SEGMENT ('.' ID_SEGMENT)*))> */
		func() bool {
			position268, tokenIndex268, depth268 := position, tokenIndex, depth
			{
				position269 := position
				depth++
				{
					position270, tokenIndex270, depth270 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l271
					}
					position++
				l272:
					{
						position273, tokenIndex273, depth273 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l273
						}
						goto l272
					l273:
						position, tokenIndex, depth = position273, tokenIndex273, depth273
					}
					if buffer[position] != rune('`') {
						goto l271
					}
					position++
					goto l270
				l271:
					position, tokenIndex, depth = position270, tokenIndex270, depth270
					{
						position274, tokenIndex274, depth274 := position, tokenIndex, depth
						{
							position275 := position
							depth++
							{
								position276, tokenIndex276, depth276 := position, tokenIndex, depth
								{
									position278, tokenIndex278, depth278 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l279
									}
									position++
									goto l278
								l279:
									position, tokenIndex, depth = position278, tokenIndex278, depth278
									if buffer[position] != rune('A') {
										goto l277
									}
									position++
								}
							l278:
								{
									position280, tokenIndex280, depth280 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l281
									}
									position++
									goto l280
								l281:
									position, tokenIndex, depth = position280, tokenIndex280, depth280
									if buffer[position] != rune('L') {
										goto l277
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
										goto l277
									}
									position++
								}
							l282:
								goto l276
							l277:
								position, tokenIndex, depth = position276, tokenIndex276, depth276
								{
									position285, tokenIndex285, depth285 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l286
									}
									position++
									goto l285
								l286:
									position, tokenIndex, depth = position285, tokenIndex285, depth285
									if buffer[position] != rune('A') {
										goto l284
									}
									position++
								}
							l285:
								{
									position287, tokenIndex287, depth287 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l288
									}
									position++
									goto l287
								l288:
									position, tokenIndex, depth = position287, tokenIndex287, depth287
									if buffer[position] != rune('N') {
										goto l284
									}
									position++
								}
							l287:
								{
									position289, tokenIndex289, depth289 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l290
									}
									position++
									goto l289
								l290:
									position, tokenIndex, depth = position289, tokenIndex289, depth289
									if buffer[position] != rune('D') {
										goto l284
									}
									position++
								}
							l289:
								goto l276
							l284:
								position, tokenIndex, depth = position276, tokenIndex276, depth276
								{
									switch buffer[position] {
									case 'W', 'w':
										{
											position292, tokenIndex292, depth292 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l293
											}
											position++
											goto l292
										l293:
											position, tokenIndex, depth = position292, tokenIndex292, depth292
											if buffer[position] != rune('W') {
												goto l274
											}
											position++
										}
									l292:
										{
											position294, tokenIndex294, depth294 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l295
											}
											position++
											goto l294
										l295:
											position, tokenIndex, depth = position294, tokenIndex294, depth294
											if buffer[position] != rune('H') {
												goto l274
											}
											position++
										}
									l294:
										{
											position296, tokenIndex296, depth296 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l297
											}
											position++
											goto l296
										l297:
											position, tokenIndex, depth = position296, tokenIndex296, depth296
											if buffer[position] != rune('E') {
												goto l274
											}
											position++
										}
									l296:
										{
											position298, tokenIndex298, depth298 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l299
											}
											position++
											goto l298
										l299:
											position, tokenIndex, depth = position298, tokenIndex298, depth298
											if buffer[position] != rune('R') {
												goto l274
											}
											position++
										}
									l298:
										{
											position300, tokenIndex300, depth300 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l301
											}
											position++
											goto l300
										l301:
											position, tokenIndex, depth = position300, tokenIndex300, depth300
											if buffer[position] != rune('E') {
												goto l274
											}
											position++
										}
									l300:
										break
									case 'T', 't':
										{
											position302, tokenIndex302, depth302 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l303
											}
											position++
											goto l302
										l303:
											position, tokenIndex, depth = position302, tokenIndex302, depth302
											if buffer[position] != rune('T') {
												goto l274
											}
											position++
										}
									l302:
										{
											position304, tokenIndex304, depth304 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l305
											}
											position++
											goto l304
										l305:
											position, tokenIndex, depth = position304, tokenIndex304, depth304
											if buffer[position] != rune('O') {
												goto l274
											}
											position++
										}
									l304:
										break
									case 'S', 's':
										{
											position306, tokenIndex306, depth306 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l307
											}
											position++
											goto l306
										l307:
											position, tokenIndex, depth = position306, tokenIndex306, depth306
											if buffer[position] != rune('S') {
												goto l274
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
												goto l274
											}
											position++
										}
									l308:
										{
											position310, tokenIndex310, depth310 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l311
											}
											position++
											goto l310
										l311:
											position, tokenIndex, depth = position310, tokenIndex310, depth310
											if buffer[position] != rune('L') {
												goto l274
											}
											position++
										}
									l310:
										{
											position312, tokenIndex312, depth312 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l313
											}
											position++
											goto l312
										l313:
											position, tokenIndex, depth = position312, tokenIndex312, depth312
											if buffer[position] != rune('E') {
												goto l274
											}
											position++
										}
									l312:
										{
											position314, tokenIndex314, depth314 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l315
											}
											position++
											goto l314
										l315:
											position, tokenIndex, depth = position314, tokenIndex314, depth314
											if buffer[position] != rune('C') {
												goto l274
											}
											position++
										}
									l314:
										{
											position316, tokenIndex316, depth316 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l317
											}
											position++
											goto l316
										l317:
											position, tokenIndex, depth = position316, tokenIndex316, depth316
											if buffer[position] != rune('T') {
												goto l274
											}
											position++
										}
									l316:
										break
									case 'O', 'o':
										{
											position318, tokenIndex318, depth318 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l319
											}
											position++
											goto l318
										l319:
											position, tokenIndex, depth = position318, tokenIndex318, depth318
											if buffer[position] != rune('O') {
												goto l274
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
												goto l274
											}
											position++
										}
									l320:
										break
									case 'N', 'n':
										{
											position322, tokenIndex322, depth322 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l323
											}
											position++
											goto l322
										l323:
											position, tokenIndex, depth = position322, tokenIndex322, depth322
											if buffer[position] != rune('N') {
												goto l274
											}
											position++
										}
									l322:
										{
											position324, tokenIndex324, depth324 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l325
											}
											position++
											goto l324
										l325:
											position, tokenIndex, depth = position324, tokenIndex324, depth324
											if buffer[position] != rune('O') {
												goto l274
											}
											position++
										}
									l324:
										{
											position326, tokenIndex326, depth326 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l327
											}
											position++
											goto l326
										l327:
											position, tokenIndex, depth = position326, tokenIndex326, depth326
											if buffer[position] != rune('T') {
												goto l274
											}
											position++
										}
									l326:
										break
									case 'M', 'm':
										{
											position328, tokenIndex328, depth328 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l329
											}
											position++
											goto l328
										l329:
											position, tokenIndex, depth = position328, tokenIndex328, depth328
											if buffer[position] != rune('M') {
												goto l274
											}
											position++
										}
									l328:
										{
											position330, tokenIndex330, depth330 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l331
											}
											position++
											goto l330
										l331:
											position, tokenIndex, depth = position330, tokenIndex330, depth330
											if buffer[position] != rune('A') {
												goto l274
											}
											position++
										}
									l330:
										{
											position332, tokenIndex332, depth332 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l333
											}
											position++
											goto l332
										l333:
											position, tokenIndex, depth = position332, tokenIndex332, depth332
											if buffer[position] != rune('T') {
												goto l274
											}
											position++
										}
									l332:
										{
											position334, tokenIndex334, depth334 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l335
											}
											position++
											goto l334
										l335:
											position, tokenIndex, depth = position334, tokenIndex334, depth334
											if buffer[position] != rune('C') {
												goto l274
											}
											position++
										}
									l334:
										{
											position336, tokenIndex336, depth336 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l337
											}
											position++
											goto l336
										l337:
											position, tokenIndex, depth = position336, tokenIndex336, depth336
											if buffer[position] != rune('H') {
												goto l274
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
												goto l274
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
												goto l274
											}
											position++
										}
									l340:
										break
									case 'I', 'i':
										{
											position342, tokenIndex342, depth342 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l343
											}
											position++
											goto l342
										l343:
											position, tokenIndex, depth = position342, tokenIndex342, depth342
											if buffer[position] != rune('I') {
												goto l274
											}
											position++
										}
									l342:
										{
											position344, tokenIndex344, depth344 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l345
											}
											position++
											goto l344
										l345:
											position, tokenIndex, depth = position344, tokenIndex344, depth344
											if buffer[position] != rune('N') {
												goto l274
											}
											position++
										}
									l344:
										break
									case 'G', 'g':
										{
											position346, tokenIndex346, depth346 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l347
											}
											position++
											goto l346
										l347:
											position, tokenIndex, depth = position346, tokenIndex346, depth346
											if buffer[position] != rune('G') {
												goto l274
											}
											position++
										}
									l346:
										{
											position348, tokenIndex348, depth348 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l349
											}
											position++
											goto l348
										l349:
											position, tokenIndex, depth = position348, tokenIndex348, depth348
											if buffer[position] != rune('R') {
												goto l274
											}
											position++
										}
									l348:
										{
											position350, tokenIndex350, depth350 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l351
											}
											position++
											goto l350
										l351:
											position, tokenIndex, depth = position350, tokenIndex350, depth350
											if buffer[position] != rune('O') {
												goto l274
											}
											position++
										}
									l350:
										{
											position352, tokenIndex352, depth352 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l353
											}
											position++
											goto l352
										l353:
											position, tokenIndex, depth = position352, tokenIndex352, depth352
											if buffer[position] != rune('U') {
												goto l274
											}
											position++
										}
									l352:
										{
											position354, tokenIndex354, depth354 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l355
											}
											position++
											goto l354
										l355:
											position, tokenIndex, depth = position354, tokenIndex354, depth354
											if buffer[position] != rune('P') {
												goto l274
											}
											position++
										}
									l354:
										break
									case 'F', 'f':
										{
											position356, tokenIndex356, depth356 := position, tokenIndex, depth
											if buffer[position] != rune('f') {
												goto l357
											}
											position++
											goto l356
										l357:
											position, tokenIndex, depth = position356, tokenIndex356, depth356
											if buffer[position] != rune('F') {
												goto l274
											}
											position++
										}
									l356:
										{
											position358, tokenIndex358, depth358 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l359
											}
											position++
											goto l358
										l359:
											position, tokenIndex, depth = position358, tokenIndex358, depth358
											if buffer[position] != rune('R') {
												goto l274
											}
											position++
										}
									l358:
										{
											position360, tokenIndex360, depth360 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l361
											}
											position++
											goto l360
										l361:
											position, tokenIndex, depth = position360, tokenIndex360, depth360
											if buffer[position] != rune('O') {
												goto l274
											}
											position++
										}
									l360:
										{
											position362, tokenIndex362, depth362 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l363
											}
											position++
											goto l362
										l363:
											position, tokenIndex, depth = position362, tokenIndex362, depth362
											if buffer[position] != rune('M') {
												goto l274
											}
											position++
										}
									l362:
										break
									case 'D', 'd':
										{
											position364, tokenIndex364, depth364 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l365
											}
											position++
											goto l364
										l365:
											position, tokenIndex, depth = position364, tokenIndex364, depth364
											if buffer[position] != rune('D') {
												goto l274
											}
											position++
										}
									l364:
										{
											position366, tokenIndex366, depth366 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l367
											}
											position++
											goto l366
										l367:
											position, tokenIndex, depth = position366, tokenIndex366, depth366
											if buffer[position] != rune('E') {
												goto l274
											}
											position++
										}
									l366:
										{
											position368, tokenIndex368, depth368 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l369
											}
											position++
											goto l368
										l369:
											position, tokenIndex, depth = position368, tokenIndex368, depth368
											if buffer[position] != rune('S') {
												goto l274
											}
											position++
										}
									l368:
										{
											position370, tokenIndex370, depth370 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l371
											}
											position++
											goto l370
										l371:
											position, tokenIndex, depth = position370, tokenIndex370, depth370
											if buffer[position] != rune('C') {
												goto l274
											}
											position++
										}
									l370:
										{
											position372, tokenIndex372, depth372 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l373
											}
											position++
											goto l372
										l373:
											position, tokenIndex, depth = position372, tokenIndex372, depth372
											if buffer[position] != rune('R') {
												goto l274
											}
											position++
										}
									l372:
										{
											position374, tokenIndex374, depth374 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l375
											}
											position++
											goto l374
										l375:
											position, tokenIndex, depth = position374, tokenIndex374, depth374
											if buffer[position] != rune('I') {
												goto l274
											}
											position++
										}
									l374:
										{
											position376, tokenIndex376, depth376 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l377
											}
											position++
											goto l376
										l377:
											position, tokenIndex, depth = position376, tokenIndex376, depth376
											if buffer[position] != rune('B') {
												goto l274
											}
											position++
										}
									l376:
										{
											position378, tokenIndex378, depth378 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l379
											}
											position++
											goto l378
										l379:
											position, tokenIndex, depth = position378, tokenIndex378, depth378
											if buffer[position] != rune('E') {
												goto l274
											}
											position++
										}
									l378:
										break
									case 'B', 'b':
										{
											position380, tokenIndex380, depth380 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l381
											}
											position++
											goto l380
										l381:
											position, tokenIndex, depth = position380, tokenIndex380, depth380
											if buffer[position] != rune('B') {
												goto l274
											}
											position++
										}
									l380:
										{
											position382, tokenIndex382, depth382 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l383
											}
											position++
											goto l382
										l383:
											position, tokenIndex, depth = position382, tokenIndex382, depth382
											if buffer[position] != rune('Y') {
												goto l274
											}
											position++
										}
									l382:
										break
									default:
										{
											position384, tokenIndex384, depth384 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l385
											}
											position++
											goto l384
										l385:
											position, tokenIndex, depth = position384, tokenIndex384, depth384
											if buffer[position] != rune('A') {
												goto l274
											}
											position++
										}
									l384:
										{
											position386, tokenIndex386, depth386 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l387
											}
											position++
											goto l386
										l387:
											position, tokenIndex, depth = position386, tokenIndex386, depth386
											if buffer[position] != rune('S') {
												goto l274
											}
											position++
										}
									l386:
										break
									}
								}

							}
						l276:
							depth--
							add(ruleKEYWORD, position275)
						}
						goto l268
					l274:
						position, tokenIndex, depth = position274, tokenIndex274, depth274
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l268
					}
				l388:
					{
						position389, tokenIndex389, depth389 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l389
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l389
						}
						goto l388
					l389:
						position, tokenIndex, depth = position389, tokenIndex389, depth389
					}
				}
			l270:
				depth--
				add(ruleIDENTIFIER, position269)
			}
			return true
		l268:
			position, tokenIndex, depth = position268, tokenIndex268, depth268
			return false
		},
		/* 29 ID_SEGMENT <- <(ID_START ID_CONT*)> */
		func() bool {
			position390, tokenIndex390, depth390 := position, tokenIndex, depth
			{
				position391 := position
				depth++
				if !_rules[ruleID_START]() {
					goto l390
				}
			l392:
				{
					position393, tokenIndex393, depth393 := position, tokenIndex, depth
					{
						position394 := position
						depth++
						{
							position395, tokenIndex395, depth395 := position, tokenIndex, depth
							if !_rules[ruleID_START]() {
								goto l396
							}
							goto l395
						l396:
							position, tokenIndex, depth = position395, tokenIndex395, depth395
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l393
							}
							position++
						}
					l395:
						depth--
						add(ruleID_CONT, position394)
					}
					goto l392
				l393:
					position, tokenIndex, depth = position393, tokenIndex393, depth393
				}
				depth--
				add(ruleID_SEGMENT, position391)
			}
			return true
		l390:
			position, tokenIndex, depth = position390, tokenIndex390, depth390
			return false
		},
		/* 30 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position397, tokenIndex397, depth397 := position, tokenIndex, depth
			{
				position398 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l397
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l397
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l397
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position398)
			}
			return true
		l397:
			position, tokenIndex, depth = position397, tokenIndex397, depth397
			return false
		},
		/* 31 ID_CONT <- <(ID_START / [0-9])> */
		nil,
		/* 32 KEYWORD <- <((('a' / 'A') ('l' / 'L') ('l' / 'L')) / (('a' / 'A') ('n' / 'N') ('d' / 'D')) / ((&('W' | 'w') (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E'))) | (&('T' | 't') (('t' / 'T') ('o' / 'O'))) | (&('S' | 's') (('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T'))) | (&('O' | 'o') (('o' / 'O') ('r' / 'R'))) | (&('N' | 'n') (('n' / 'N') ('o' / 'O') ('t' / 'T'))) | (&('M' | 'm') (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S'))) | (&('I' | 'i') (('i' / 'I') ('n' / 'N'))) | (&('G' | 'g') (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P'))) | (&('F' | 'f') (('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))) | (&('D' | 'd') (('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E'))) | (&('B' | 'b') (('b' / 'B') ('y' / 'Y'))) | (&('A' | 'a') (('a' / 'A') ('s' / 'S')))))> */
		nil,
		/* 33 OP_ADD <- <(_ '+' _)> */
		nil,
		/* 34 OP_SUB <- <(_ '-' _)> */
		nil,
		/* 35 OP_MULT <- <(_ '*' _)> */
		nil,
		/* 36 OP_DIV <- <(_ '/' _)> */
		nil,
		/* 37 OP_AND <- <(__ (('a' / 'A') ('n' / 'N') ('d' / 'D')) __)> */
		nil,
		/* 38 OP_OR <- <(__ (('o' / 'O') ('r' / 'R')) __)> */
		nil,
		/* 39 OP_NOT <- <(('n' / 'N') ('o' / 'O') ('t' / 'T') __)> */
		nil,
		/* 40 STRING <- <(('\'' CHAR* '\'') / ('"' CHAR* '"'))> */
		func() bool {
			position409, tokenIndex409, depth409 := position, tokenIndex, depth
			{
				position410 := position
				depth++
				{
					position411, tokenIndex411, depth411 := position, tokenIndex, depth
					if buffer[position] != rune('\'') {
						goto l412
					}
					position++
				l413:
					{
						position414, tokenIndex414, depth414 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l414
						}
						goto l413
					l414:
						position, tokenIndex, depth = position414, tokenIndex414, depth414
					}
					if buffer[position] != rune('\'') {
						goto l412
					}
					position++
					goto l411
				l412:
					position, tokenIndex, depth = position411, tokenIndex411, depth411
					if buffer[position] != rune('"') {
						goto l409
					}
					position++
				l415:
					{
						position416, tokenIndex416, depth416 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l416
						}
						goto l415
					l416:
						position, tokenIndex, depth = position416, tokenIndex416, depth416
					}
					if buffer[position] != rune('"') {
						goto l409
					}
					position++
				}
			l411:
				depth--
				add(ruleSTRING, position410)
			}
			return true
		l409:
			position, tokenIndex, depth = position409, tokenIndex409, depth409
			return false
		},
		/* 41 CHAR <- <(('\\' ESCAPE_CLASS) / (!ESCAPE_CLASS .))> */
		func() bool {
			position417, tokenIndex417, depth417 := position, tokenIndex, depth
			{
				position418 := position
				depth++
				{
					position419, tokenIndex419, depth419 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l420
					}
					position++
					if !_rules[ruleESCAPE_CLASS]() {
						goto l420
					}
					goto l419
				l420:
					position, tokenIndex, depth = position419, tokenIndex419, depth419
					{
						position421, tokenIndex421, depth421 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l421
						}
						goto l417
					l421:
						position, tokenIndex, depth = position421, tokenIndex421, depth421
					}
					if !matchDot() {
						goto l417
					}
				}
			l419:
				depth--
				add(ruleCHAR, position418)
			}
			return true
		l417:
			position, tokenIndex, depth = position417, tokenIndex417, depth417
			return false
		},
		/* 42 ESCAPE_CLASS <- <((&('\\') '\\') | (&('"') '"') | (&('`') '`') | (&('\'') '\''))> */
		func() bool {
			position422, tokenIndex422, depth422 := position, tokenIndex, depth
			{
				position423 := position
				depth++
				{
					switch buffer[position] {
					case '\\':
						if buffer[position] != rune('\\') {
							goto l422
						}
						position++
						break
					case '"':
						if buffer[position] != rune('"') {
							goto l422
						}
						position++
						break
					case '`':
						if buffer[position] != rune('`') {
							goto l422
						}
						position++
						break
					default:
						if buffer[position] != rune('\'') {
							goto l422
						}
						position++
						break
					}
				}

				depth--
				add(ruleESCAPE_CLASS, position423)
			}
			return true
		l422:
			position, tokenIndex, depth = position422, tokenIndex422, depth422
			return false
		},
		/* 43 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		nil,
		/* 44 NUMBER_NATURAL <- <('0' / ([1-9] [0-9]*))> */
		func() bool {
			position426, tokenIndex426, depth426 := position, tokenIndex, depth
			{
				position427 := position
				depth++
				{
					position428, tokenIndex428, depth428 := position, tokenIndex, depth
					if buffer[position] != rune('0') {
						goto l429
					}
					position++
					goto l428
				l429:
					position, tokenIndex, depth = position428, tokenIndex428, depth428
					if c := buffer[position]; c < rune('1') || c > rune('9') {
						goto l426
					}
					position++
				l430:
					{
						position431, tokenIndex431, depth431 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l431
						}
						position++
						goto l430
					l431:
						position, tokenIndex, depth = position431, tokenIndex431, depth431
					}
				}
			l428:
				depth--
				add(ruleNUMBER_NATURAL, position427)
			}
			return true
		l426:
			position, tokenIndex, depth = position426, tokenIndex426, depth426
			return false
		},
		/* 45 NUMBER_FRACTION <- <('.' [0-9]+)> */
		nil,
		/* 46 NUMBER_INTEGER <- <('-'? NUMBER_NATURAL)> */
		nil,
		/* 47 NUMBER_EXP <- <(('e' / 'E') ('+' / '-')? [0-9]+)> */
		nil,
		/* 48 PAREN_OPEN <- <(_ '(' _)> */
		func() bool {
			position435, tokenIndex435, depth435 := position, tokenIndex, depth
			{
				position436 := position
				depth++
				if !_rules[rule_]() {
					goto l435
				}
				if buffer[position] != rune('(') {
					goto l435
				}
				position++
				if !_rules[rule_]() {
					goto l435
				}
				depth--
				add(rulePAREN_OPEN, position436)
			}
			return true
		l435:
			position, tokenIndex, depth = position435, tokenIndex435, depth435
			return false
		},
		/* 49 PAREN_CLOSE <- <(_ ')' _)> */
		func() bool {
			position437, tokenIndex437, depth437 := position, tokenIndex, depth
			{
				position438 := position
				depth++
				if !_rules[rule_]() {
					goto l437
				}
				if buffer[position] != rune(')') {
					goto l437
				}
				position++
				if !_rules[rule_]() {
					goto l437
				}
				depth--
				add(rulePAREN_CLOSE, position438)
			}
			return true
		l437:
			position, tokenIndex, depth = position437, tokenIndex437, depth437
			return false
		},
		/* 50 COMMA <- <(_ ',' _)> */
		func() bool {
			position439, tokenIndex439, depth439 := position, tokenIndex, depth
			{
				position440 := position
				depth++
				if !_rules[rule_]() {
					goto l439
				}
				if buffer[position] != rune(',') {
					goto l439
				}
				position++
				if !_rules[rule_]() {
					goto l439
				}
				depth--
				add(ruleCOMMA, position440)
			}
			return true
		l439:
			position, tokenIndex, depth = position439, tokenIndex439, depth439
			return false
		},
		/* 51 _ <- <SPACE*> */
		func() bool {
			{
				position442 := position
				depth++
			l443:
				{
					position444, tokenIndex444, depth444 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l444
					}
					goto l443
				l444:
					position, tokenIndex, depth = position444, tokenIndex444, depth444
				}
				depth--
				add(rule_, position442)
			}
			return true
		},
		/* 52 __ <- <SPACE+> */
		func() bool {
			position445, tokenIndex445, depth445 := position, tokenIndex, depth
			{
				position446 := position
				depth++
				if !_rules[ruleSPACE]() {
					goto l445
				}
			l447:
				{
					position448, tokenIndex448, depth448 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l448
					}
					goto l447
				l448:
					position, tokenIndex, depth = position448, tokenIndex448, depth448
				}
				depth--
				add(rule__, position446)
			}
			return true
		l445:
			position, tokenIndex, depth = position445, tokenIndex445, depth445
			return false
		},
		/* 53 SPACE <- <((&('\t') '\t') | (&('\n') '\n') | (&(' ') ' '))> */
		func() bool {
			position449, tokenIndex449, depth449 := position, tokenIndex, depth
			{
				position450 := position
				depth++
				{
					switch buffer[position] {
					case '\t':
						if buffer[position] != rune('\t') {
							goto l449
						}
						position++
						break
					case '\n':
						if buffer[position] != rune('\n') {
							goto l449
						}
						position++
						break
					default:
						if buffer[position] != rune(' ') {
							goto l449
						}
						position++
						break
					}
				}

				depth--
				add(ruleSPACE, position450)
			}
			return true
		l449:
			position, tokenIndex, depth = position449, tokenIndex449, depth449
			return false
		},
		/* 55 Action0 <- <{
		   p.makeSelect()
		 }> */
		nil,
		/* 56 Action1 <- <{ p.makeDescribeAll() }> */
		nil,
		nil,
		/* 58 Action2 <- <{ p.addLiteralNode(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 59 Action3 <- <{ p.makeDescribe() }> */
		nil,
		/* 60 Action4 <- <{ p.addNullPredicate() }> */
		nil,
		/* 61 Action5 <- <{ p.addExpressionList() }> */
		nil,
		/* 62 Action6 <- <{ p.appendExpression() }> */
		nil,
		/* 63 Action7 <- <{ p.appendExpression() }> */
		nil,
		/* 64 Action8 <- <{ p.addOperatorLiteralNode("*") }> */
		nil,
		/* 65 Action9 <- <{ p.addOperatorLiteralNode("-") }> */
		nil,
		/* 66 Action10 <- <{ p.addOperatorExpressionNode() }> */
		nil,
		/* 67 Action11 <- <{ p.addOperatorLiteralNode("*") }> */
		nil,
		/* 68 Action12 <- <{ p.addOperatorLiteralNode("*") }> */
		nil,
		/* 69 Action13 <- <{ p.addOperatorExpressionNode() }> */
		nil,
		/* 70 Action14 <- <{ p.addNumberNode(buffer[begin:end]) }> */
		nil,
		/* 71 Action15 <- <{
		   p.addLiteralNode(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 72 Action16 <- <{
		   p.addLiteralNode(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 73 Action17 <- <{
		   p.addLiteralNode(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 74 Action18 <- <{ p.addNullPredicate() }> */
		nil,
		/* 75 Action19 <- <{
		   p.addMetricReferenceNode()
		 }> */
		nil,
		/* 76 Action20 <- <{ p.addAndPredicate() }> */
		nil,
		/* 77 Action21 <- <{ p.addOrPredicate() }> */
		nil,
		/* 78 Action22 <- <{ p.addNotPredicate() }> */
		nil,
		/* 79 Action23 <- <{
		   p.addLiteralMatcher()
		 }> */
		nil,
		/* 80 Action24 <- <{
		   p.addLiteralMatcher()
		   p.addNotPredicate()
		 }> */
		nil,
		/* 81 Action25 <- <{
		   p.addRegexMatcher()
		 }> */
		nil,
		/* 82 Action26 <- <{
		   p.addListMatcher()
		 }> */
		nil,
		/* 83 Action27 <- <{
		  p.addLiteralNode(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 84 Action28 <- <{ p.addLiteralListNode() }> */
		nil,
		/* 85 Action29 <- <{
		  p.appendLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 86 Action30 <- <{ p.addTag(buffer[begin:end]) }> */
		nil,
	}
	p.rules = _rules
}
