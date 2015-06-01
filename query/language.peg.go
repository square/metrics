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
	rulepropertyClause
	ruleoptionalPredicateClause
	ruleexpressionList
	ruleexpression_1
	ruleexpression_2
	ruleexpression_3
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
	rulePROPERTY_KEY
	rulePROPERTY_VALUE
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
	ruleAction31
	ruleAction32
	ruleAction33

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
	"propertyClause",
	"optionalPredicateClause",
	"expressionList",
	"expression_1",
	"expression_2",
	"expression_3",
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
	"PROPERTY_KEY",
	"PROPERTY_VALUE",
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
	"Action31",
	"Action32",
	"Action33",

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
	rules  [91]func() bool
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
			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
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
			p.addOperatorLiteral("*")
		case ruleAction9:
			p.addOperatorLiteral("-")
		case ruleAction10:
			p.addOperatorFunction()
		case ruleAction11:
			p.addOperatorLiteral("*")
		case ruleAction12:
			p.addOperatorLiteral("*")
		case ruleAction13:
			p.addOperatorFunction()
		case ruleAction14:
			p.addNumberNode(buffer[begin:end])
		case ruleAction15:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction16:
			p.addGroupBy()
		case ruleAction17:

			p.addFunctionInvocation()

		case ruleAction18:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction19:
			p.addNullPredicate()
		case ruleAction20:

			p.addMetricExpression()

		case ruleAction21:

			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))

		case ruleAction22:

			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))

		case ruleAction23:
			p.addAndPredicate()
		case ruleAction24:
			p.addAndPredicate()
		case ruleAction25:
			p.addNotPredicate()
		case ruleAction26:

			p.addLiteralMatcher()

		case ruleAction27:

			p.addLiteralMatcher()
			p.addNotPredicate()

		case ruleAction28:

			p.addRegexMatcher()

		case ruleAction29:

			p.addListMatcher()

		case ruleAction30:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction31:
			p.addLiteralList()
		case ruleAction32:

			p.appendLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction33:
			p.addTagLiteral(unescapeLiteral(buffer[begin:end]))

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
						l18:
							{
								position19, tokenIndex19, depth19 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l19
								}
								if !_rules[rulePROPERTY_KEY]() {
									goto l19
								}
								if !_rules[rule__]() {
									goto l19
								}
								{
									position20 := position
									depth++
									{
										position21 := position
										depth++
										{
											position22, tokenIndex22, depth22 := position, tokenIndex, depth
											if !_rules[ruleNUMBER_NATURAL]() {
												goto l23
											}
											goto l22
										l23:
											position, tokenIndex, depth = position22, tokenIndex22, depth22
											if !_rules[ruleSTRING]() {
												goto l19
											}
										}
									l22:
										depth--
										add(ruleTIMESTAMP, position21)
									}
									depth--
									add(rulePROPERTY_VALUE, position20)
								}
								goto l18
							l19:
								position, tokenIndex, depth = position19, tokenIndex19, depth19
							}
							depth--
							add(rulepropertyClause, position17)
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
						position25 := position
						depth++
						{
							position26, tokenIndex26, depth26 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l27
							}
							position++
							goto l26
						l27:
							position, tokenIndex, depth = position26, tokenIndex26, depth26
							if buffer[position] != rune('D') {
								goto l0
							}
							position++
						}
					l26:
						{
							position28, tokenIndex28, depth28 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l29
							}
							position++
							goto l28
						l29:
							position, tokenIndex, depth = position28, tokenIndex28, depth28
							if buffer[position] != rune('E') {
								goto l0
							}
							position++
						}
					l28:
						{
							position30, tokenIndex30, depth30 := position, tokenIndex, depth
							if buffer[position] != rune('s') {
								goto l31
							}
							position++
							goto l30
						l31:
							position, tokenIndex, depth = position30, tokenIndex30, depth30
							if buffer[position] != rune('S') {
								goto l0
							}
							position++
						}
					l30:
						{
							position32, tokenIndex32, depth32 := position, tokenIndex, depth
							if buffer[position] != rune('c') {
								goto l33
							}
							position++
							goto l32
						l33:
							position, tokenIndex, depth = position32, tokenIndex32, depth32
							if buffer[position] != rune('C') {
								goto l0
							}
							position++
						}
					l32:
						{
							position34, tokenIndex34, depth34 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l35
							}
							position++
							goto l34
						l35:
							position, tokenIndex, depth = position34, tokenIndex34, depth34
							if buffer[position] != rune('R') {
								goto l0
							}
							position++
						}
					l34:
						{
							position36, tokenIndex36, depth36 := position, tokenIndex, depth
							if buffer[position] != rune('i') {
								goto l37
							}
							position++
							goto l36
						l37:
							position, tokenIndex, depth = position36, tokenIndex36, depth36
							if buffer[position] != rune('I') {
								goto l0
							}
							position++
						}
					l36:
						{
							position38, tokenIndex38, depth38 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l39
							}
							position++
							goto l38
						l39:
							position, tokenIndex, depth = position38, tokenIndex38, depth38
							if buffer[position] != rune('B') {
								goto l0
							}
							position++
						}
					l38:
						{
							position40, tokenIndex40, depth40 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l41
							}
							position++
							goto l40
						l41:
							position, tokenIndex, depth = position40, tokenIndex40, depth40
							if buffer[position] != rune('E') {
								goto l0
							}
							position++
						}
					l40:
						if !_rules[rule__]() {
							goto l0
						}
						{
							position42, tokenIndex42, depth42 := position, tokenIndex, depth
							{
								position44 := position
								depth++
								{
									position45, tokenIndex45, depth45 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l46
									}
									position++
									goto l45
								l46:
									position, tokenIndex, depth = position45, tokenIndex45, depth45
									if buffer[position] != rune('A') {
										goto l43
									}
									position++
								}
							l45:
								{
									position47, tokenIndex47, depth47 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l48
									}
									position++
									goto l47
								l48:
									position, tokenIndex, depth = position47, tokenIndex47, depth47
									if buffer[position] != rune('L') {
										goto l43
									}
									position++
								}
							l47:
								{
									position49, tokenIndex49, depth49 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l50
									}
									position++
									goto l49
								l50:
									position, tokenIndex, depth = position49, tokenIndex49, depth49
									if buffer[position] != rune('L') {
										goto l43
									}
									position++
								}
							l49:
								{
									add(ruleAction1, position)
								}
								depth--
								add(ruledescribeAllStmt, position44)
							}
							goto l42
						l43:
							position, tokenIndex, depth = position42, tokenIndex42, depth42
							{
								position52 := position
								depth++
								{
									position53 := position
									depth++
									{
										position54 := position
										depth++
										if !_rules[ruleIDENTIFIER]() {
											goto l0
										}
										depth--
										add(ruleMETRIC_NAME, position54)
									}
									depth--
									add(rulePegText, position53)
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
								add(ruledescribeSingleStmt, position52)
							}
						}
					l42:
						depth--
						add(ruledescribeStmt, position25)
					}
				}
			l2:
				{
					position57, tokenIndex57, depth57 := position, tokenIndex, depth
					if !matchDot() {
						goto l57
					}
					goto l0
				l57:
					position, tokenIndex, depth = position57, tokenIndex57, depth57
				}
				depth--
				add(ruleroot, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 selectStmt <- <(('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T') __ expressionList optionalPredicateClause propertyClause Action0)> */
		nil,
		/* 2 describeStmt <- <(('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E') __ (describeAllStmt / describeSingleStmt))> */
		nil,
		/* 3 describeAllStmt <- <(('a' / 'A') ('l' / 'L') ('l' / 'L') Action1)> */
		nil,
		/* 4 describeSingleStmt <- <(<METRIC_NAME> Action2 optionalPredicateClause Action3)> */
		nil,
		/* 5 propertyClause <- <(_ PROPERTY_KEY __ PROPERTY_VALUE)*> */
		nil,
		/* 6 optionalPredicateClause <- <((__ predicateClause) / Action4)> */
		func() bool {
			{
				position64 := position
				depth++
				{
					position65, tokenIndex65, depth65 := position, tokenIndex, depth
					if !_rules[rule__]() {
						goto l66
					}
					{
						position67 := position
						depth++
						{
							position68, tokenIndex68, depth68 := position, tokenIndex, depth
							if buffer[position] != rune('w') {
								goto l69
							}
							position++
							goto l68
						l69:
							position, tokenIndex, depth = position68, tokenIndex68, depth68
							if buffer[position] != rune('W') {
								goto l66
							}
							position++
						}
					l68:
						{
							position70, tokenIndex70, depth70 := position, tokenIndex, depth
							if buffer[position] != rune('h') {
								goto l71
							}
							position++
							goto l70
						l71:
							position, tokenIndex, depth = position70, tokenIndex70, depth70
							if buffer[position] != rune('H') {
								goto l66
							}
							position++
						}
					l70:
						{
							position72, tokenIndex72, depth72 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l73
							}
							position++
							goto l72
						l73:
							position, tokenIndex, depth = position72, tokenIndex72, depth72
							if buffer[position] != rune('E') {
								goto l66
							}
							position++
						}
					l72:
						{
							position74, tokenIndex74, depth74 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l75
							}
							position++
							goto l74
						l75:
							position, tokenIndex, depth = position74, tokenIndex74, depth74
							if buffer[position] != rune('R') {
								goto l66
							}
							position++
						}
					l74:
						{
							position76, tokenIndex76, depth76 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l77
							}
							position++
							goto l76
						l77:
							position, tokenIndex, depth = position76, tokenIndex76, depth76
							if buffer[position] != rune('E') {
								goto l66
							}
							position++
						}
					l76:
						if !_rules[rule__]() {
							goto l66
						}
						if !_rules[rulepredicate_1]() {
							goto l66
						}
						depth--
						add(rulepredicateClause, position67)
					}
					goto l65
				l66:
					position, tokenIndex, depth = position65, tokenIndex65, depth65
					{
						add(ruleAction4, position)
					}
				}
			l65:
				depth--
				add(ruleoptionalPredicateClause, position64)
			}
			return true
		},
		/* 7 expressionList <- <(Action5 expression_1 Action6 (COMMA expression_1 Action7)*)> */
		func() bool {
			position79, tokenIndex79, depth79 := position, tokenIndex, depth
			{
				position80 := position
				depth++
				{
					add(ruleAction5, position)
				}
				if !_rules[ruleexpression_1]() {
					goto l79
				}
				{
					add(ruleAction6, position)
				}
			l83:
				{
					position84, tokenIndex84, depth84 := position, tokenIndex, depth
					if !_rules[ruleCOMMA]() {
						goto l84
					}
					if !_rules[ruleexpression_1]() {
						goto l84
					}
					{
						add(ruleAction7, position)
					}
					goto l83
				l84:
					position, tokenIndex, depth = position84, tokenIndex84, depth84
				}
				depth--
				add(ruleexpressionList, position80)
			}
			return true
		l79:
			position, tokenIndex, depth = position79, tokenIndex79, depth79
			return false
		},
		/* 8 expression_1 <- <(expression_2 (((OP_ADD Action8) / (OP_SUB Action9)) expression_2 Action10)*)> */
		func() bool {
			position86, tokenIndex86, depth86 := position, tokenIndex, depth
			{
				position87 := position
				depth++
				if !_rules[ruleexpression_2]() {
					goto l86
				}
			l88:
				{
					position89, tokenIndex89, depth89 := position, tokenIndex, depth
					{
						position90, tokenIndex90, depth90 := position, tokenIndex, depth
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
						{
							add(ruleAction8, position)
						}
						goto l90
					l91:
						position, tokenIndex, depth = position90, tokenIndex90, depth90
						{
							position94 := position
							depth++
							if !_rules[rule_]() {
								goto l89
							}
							if buffer[position] != rune('-') {
								goto l89
							}
							position++
							if !_rules[rule_]() {
								goto l89
							}
							depth--
							add(ruleOP_SUB, position94)
						}
						{
							add(ruleAction9, position)
						}
					}
				l90:
					if !_rules[ruleexpression_2]() {
						goto l89
					}
					{
						add(ruleAction10, position)
					}
					goto l88
				l89:
					position, tokenIndex, depth = position89, tokenIndex89, depth89
				}
				depth--
				add(ruleexpression_1, position87)
			}
			return true
		l86:
			position, tokenIndex, depth = position86, tokenIndex86, depth86
			return false
		},
		/* 9 expression_2 <- <(expression_3 (((OP_DIV Action11) / (OP_MULT Action12)) expression_3 Action13)*)> */
		func() bool {
			position97, tokenIndex97, depth97 := position, tokenIndex, depth
			{
				position98 := position
				depth++
				if !_rules[ruleexpression_3]() {
					goto l97
				}
			l99:
				{
					position100, tokenIndex100, depth100 := position, tokenIndex, depth
					{
						position101, tokenIndex101, depth101 := position, tokenIndex, depth
						{
							position103 := position
							depth++
							if !_rules[rule_]() {
								goto l102
							}
							if buffer[position] != rune('/') {
								goto l102
							}
							position++
							if !_rules[rule_]() {
								goto l102
							}
							depth--
							add(ruleOP_DIV, position103)
						}
						{
							add(ruleAction11, position)
						}
						goto l101
					l102:
						position, tokenIndex, depth = position101, tokenIndex101, depth101
						{
							position105 := position
							depth++
							if !_rules[rule_]() {
								goto l100
							}
							if buffer[position] != rune('*') {
								goto l100
							}
							position++
							if !_rules[rule_]() {
								goto l100
							}
							depth--
							add(ruleOP_MULT, position105)
						}
						{
							add(ruleAction12, position)
						}
					}
				l101:
					if !_rules[ruleexpression_3]() {
						goto l100
					}
					{
						add(ruleAction13, position)
					}
					goto l99
				l100:
					position, tokenIndex, depth = position100, tokenIndex100, depth100
				}
				depth--
				add(ruleexpression_2, position98)
			}
			return true
		l97:
			position, tokenIndex, depth = position97, tokenIndex97, depth97
			return false
		},
		/* 10 expression_3 <- <(expression_function / ((&('-' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') (<NUMBER> Action14)) | (&('\t' | '\n' | ' ' | '(') (PAREN_OPEN expression_1 PAREN_CLOSE)) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') expression_metric)))> */
		func() bool {
			position108, tokenIndex108, depth108 := position, tokenIndex, depth
			{
				position109 := position
				depth++
				{
					position110, tokenIndex110, depth110 := position, tokenIndex, depth
					{
						position112 := position
						depth++
						{
							position113 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l111
							}
							depth--
							add(rulePegText, position113)
						}
						{
							add(ruleAction15, position)
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l111
						}
						if !_rules[ruleexpressionList]() {
							goto l111
						}
						{
							add(ruleAction16, position)
						}
						{
							position116, tokenIndex116, depth116 := position, tokenIndex, depth
							if !_rules[rule__]() {
								goto l116
							}
							{
								position118 := position
								depth++
								{
									position119, tokenIndex119, depth119 := position, tokenIndex, depth
									if buffer[position] != rune('g') {
										goto l120
									}
									position++
									goto l119
								l120:
									position, tokenIndex, depth = position119, tokenIndex119, depth119
									if buffer[position] != rune('G') {
										goto l116
									}
									position++
								}
							l119:
								{
									position121, tokenIndex121, depth121 := position, tokenIndex, depth
									if buffer[position] != rune('r') {
										goto l122
									}
									position++
									goto l121
								l122:
									position, tokenIndex, depth = position121, tokenIndex121, depth121
									if buffer[position] != rune('R') {
										goto l116
									}
									position++
								}
							l121:
								{
									position123, tokenIndex123, depth123 := position, tokenIndex, depth
									if buffer[position] != rune('o') {
										goto l124
									}
									position++
									goto l123
								l124:
									position, tokenIndex, depth = position123, tokenIndex123, depth123
									if buffer[position] != rune('O') {
										goto l116
									}
									position++
								}
							l123:
								{
									position125, tokenIndex125, depth125 := position, tokenIndex, depth
									if buffer[position] != rune('u') {
										goto l126
									}
									position++
									goto l125
								l126:
									position, tokenIndex, depth = position125, tokenIndex125, depth125
									if buffer[position] != rune('U') {
										goto l116
									}
									position++
								}
							l125:
								{
									position127, tokenIndex127, depth127 := position, tokenIndex, depth
									if buffer[position] != rune('p') {
										goto l128
									}
									position++
									goto l127
								l128:
									position, tokenIndex, depth = position127, tokenIndex127, depth127
									if buffer[position] != rune('P') {
										goto l116
									}
									position++
								}
							l127:
								if !_rules[rule__]() {
									goto l116
								}
								{
									position129, tokenIndex129, depth129 := position, tokenIndex, depth
									if buffer[position] != rune('b') {
										goto l130
									}
									position++
									goto l129
								l130:
									position, tokenIndex, depth = position129, tokenIndex129, depth129
									if buffer[position] != rune('B') {
										goto l116
									}
									position++
								}
							l129:
								{
									position131, tokenIndex131, depth131 := position, tokenIndex, depth
									if buffer[position] != rune('y') {
										goto l132
									}
									position++
									goto l131
								l132:
									position, tokenIndex, depth = position131, tokenIndex131, depth131
									if buffer[position] != rune('Y') {
										goto l116
									}
									position++
								}
							l131:
								if !_rules[rule__]() {
									goto l116
								}
								{
									position133 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l116
									}
									depth--
									add(rulePegText, position133)
								}
								{
									add(ruleAction21, position)
								}
							l135:
								{
									position136, tokenIndex136, depth136 := position, tokenIndex, depth
									if !_rules[ruleCOMMA]() {
										goto l136
									}
									{
										position137 := position
										depth++
										if !_rules[ruleCOLUMN_NAME]() {
											goto l136
										}
										depth--
										add(rulePegText, position137)
									}
									{
										add(ruleAction22, position)
									}
									goto l135
								l136:
									position, tokenIndex, depth = position136, tokenIndex136, depth136
								}
								depth--
								add(rulegroupByClause, position118)
							}
							goto l117
						l116:
							position, tokenIndex, depth = position116, tokenIndex116, depth116
						}
					l117:
						if !_rules[rulePAREN_CLOSE]() {
							goto l111
						}
						{
							add(ruleAction17, position)
						}
						depth--
						add(ruleexpression_function, position112)
					}
					goto l110
				l111:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					{
						switch buffer[position] {
						case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							{
								position141 := position
								depth++
								{
									position142 := position
									depth++
									{
										position143 := position
										depth++
										{
											position144, tokenIndex144, depth144 := position, tokenIndex, depth
											if buffer[position] != rune('-') {
												goto l144
											}
											position++
											goto l145
										l144:
											position, tokenIndex, depth = position144, tokenIndex144, depth144
										}
									l145:
										if !_rules[ruleNUMBER_NATURAL]() {
											goto l108
										}
										depth--
										add(ruleNUMBER_INTEGER, position143)
									}
									{
										position146, tokenIndex146, depth146 := position, tokenIndex, depth
										{
											position148 := position
											depth++
											if buffer[position] != rune('.') {
												goto l146
											}
											position++
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l146
											}
											position++
										l149:
											{
												position150, tokenIndex150, depth150 := position, tokenIndex, depth
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l150
												}
												position++
												goto l149
											l150:
												position, tokenIndex, depth = position150, tokenIndex150, depth150
											}
											depth--
											add(ruleNUMBER_FRACTION, position148)
										}
										goto l147
									l146:
										position, tokenIndex, depth = position146, tokenIndex146, depth146
									}
								l147:
									{
										position151, tokenIndex151, depth151 := position, tokenIndex, depth
										{
											position153 := position
											depth++
											{
												position154, tokenIndex154, depth154 := position, tokenIndex, depth
												if buffer[position] != rune('e') {
													goto l155
												}
												position++
												goto l154
											l155:
												position, tokenIndex, depth = position154, tokenIndex154, depth154
												if buffer[position] != rune('E') {
													goto l151
												}
												position++
											}
										l154:
											{
												position156, tokenIndex156, depth156 := position, tokenIndex, depth
												{
													position158, tokenIndex158, depth158 := position, tokenIndex, depth
													if buffer[position] != rune('+') {
														goto l159
													}
													position++
													goto l158
												l159:
													position, tokenIndex, depth = position158, tokenIndex158, depth158
													if buffer[position] != rune('-') {
														goto l156
													}
													position++
												}
											l158:
												goto l157
											l156:
												position, tokenIndex, depth = position156, tokenIndex156, depth156
											}
										l157:
											if c := buffer[position]; c < rune('0') || c > rune('9') {
												goto l151
											}
											position++
										l160:
											{
												position161, tokenIndex161, depth161 := position, tokenIndex, depth
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l161
												}
												position++
												goto l160
											l161:
												position, tokenIndex, depth = position161, tokenIndex161, depth161
											}
											depth--
											add(ruleNUMBER_EXP, position153)
										}
										goto l152
									l151:
										position, tokenIndex, depth = position151, tokenIndex151, depth151
									}
								l152:
									depth--
									add(ruleNUMBER, position142)
								}
								depth--
								add(rulePegText, position141)
							}
							{
								add(ruleAction14, position)
							}
							break
						case '\t', '\n', ' ', '(':
							if !_rules[rulePAREN_OPEN]() {
								goto l108
							}
							if !_rules[ruleexpression_1]() {
								goto l108
							}
							if !_rules[rulePAREN_CLOSE]() {
								goto l108
							}
							break
						default:
							{
								position163 := position
								depth++
								{
									position164 := position
									depth++
									if !_rules[ruleIDENTIFIER]() {
										goto l108
									}
									depth--
									add(rulePegText, position164)
								}
								{
									add(ruleAction18, position)
								}
								{
									position166, tokenIndex166, depth166 := position, tokenIndex, depth
									{
										position168, tokenIndex168, depth168 := position, tokenIndex, depth
										if buffer[position] != rune('[') {
											goto l169
										}
										position++
										if !_rules[rule_]() {
											goto l169
										}
										if !_rules[rulepredicate_1]() {
											goto l169
										}
										if !_rules[rule_]() {
											goto l169
										}
										if buffer[position] != rune(']') {
											goto l169
										}
										position++
										goto l168
									l169:
										position, tokenIndex, depth = position168, tokenIndex168, depth168
										{
											add(ruleAction19, position)
										}
									}
								l168:
									goto l167

									position, tokenIndex, depth = position166, tokenIndex166, depth166
								}
							l167:
								{
									add(ruleAction20, position)
								}
								depth--
								add(ruleexpression_metric, position163)
							}
							break
						}
					}

				}
			l110:
				depth--
				add(ruleexpression_3, position109)
			}
			return true
		l108:
			position, tokenIndex, depth = position108, tokenIndex108, depth108
			return false
		},
		/* 11 expression_function <- <(<IDENTIFIER> Action15 PAREN_OPEN expressionList Action16 (__ groupByClause)? PAREN_CLOSE Action17)> */
		nil,
		/* 12 expression_metric <- <(<IDENTIFIER> Action18 (('[' _ predicate_1 _ ']') / Action19)? Action20)> */
		nil,
		/* 13 groupByClause <- <(('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P') __ (('b' / 'B') ('y' / 'Y')) __ <COLUMN_NAME> Action21 (COMMA <COLUMN_NAME> Action22)*)> */
		nil,
		/* 14 predicateClause <- <(('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E') __ predicate_1)> */
		nil,
		/* 15 predicate_1 <- <(predicate_2 (OP_AND predicate_2 Action23)*)> */
		func() bool {
			position176, tokenIndex176, depth176 := position, tokenIndex, depth
			{
				position177 := position
				depth++
				if !_rules[rulepredicate_2]() {
					goto l176
				}
			l178:
				{
					position179, tokenIndex179, depth179 := position, tokenIndex, depth
					{
						position180 := position
						depth++
						if !_rules[rule_]() {
							goto l179
						}
						{
							position181, tokenIndex181, depth181 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l182
							}
							position++
							goto l181
						l182:
							position, tokenIndex, depth = position181, tokenIndex181, depth181
							if buffer[position] != rune('A') {
								goto l179
							}
							position++
						}
					l181:
						{
							position183, tokenIndex183, depth183 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l184
							}
							position++
							goto l183
						l184:
							position, tokenIndex, depth = position183, tokenIndex183, depth183
							if buffer[position] != rune('N') {
								goto l179
							}
							position++
						}
					l183:
						{
							position185, tokenIndex185, depth185 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l186
							}
							position++
							goto l185
						l186:
							position, tokenIndex, depth = position185, tokenIndex185, depth185
							if buffer[position] != rune('D') {
								goto l179
							}
							position++
						}
					l185:
						if !_rules[rule_]() {
							goto l179
						}
						depth--
						add(ruleOP_AND, position180)
					}
					if !_rules[rulepredicate_2]() {
						goto l179
					}
					{
						add(ruleAction23, position)
					}
					goto l178
				l179:
					position, tokenIndex, depth = position179, tokenIndex179, depth179
				}
				depth--
				add(rulepredicate_1, position177)
			}
			return true
		l176:
			position, tokenIndex, depth = position176, tokenIndex176, depth176
			return false
		},
		/* 16 predicate_2 <- <(predicate_3 (OP_OR predicate_3 Action24)*)> */
		func() bool {
			position188, tokenIndex188, depth188 := position, tokenIndex, depth
			{
				position189 := position
				depth++
				if !_rules[rulepredicate_3]() {
					goto l188
				}
			l190:
				{
					position191, tokenIndex191, depth191 := position, tokenIndex, depth
					{
						position192 := position
						depth++
						if !_rules[rule_]() {
							goto l191
						}
						{
							position193, tokenIndex193, depth193 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l194
							}
							position++
							goto l193
						l194:
							position, tokenIndex, depth = position193, tokenIndex193, depth193
							if buffer[position] != rune('O') {
								goto l191
							}
							position++
						}
					l193:
						{
							position195, tokenIndex195, depth195 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l196
							}
							position++
							goto l195
						l196:
							position, tokenIndex, depth = position195, tokenIndex195, depth195
							if buffer[position] != rune('R') {
								goto l191
							}
							position++
						}
					l195:
						if !_rules[rule_]() {
							goto l191
						}
						depth--
						add(ruleOP_OR, position192)
					}
					if !_rules[rulepredicate_3]() {
						goto l191
					}
					{
						add(ruleAction24, position)
					}
					goto l190
				l191:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
				}
				depth--
				add(rulepredicate_2, position189)
			}
			return true
		l188:
			position, tokenIndex, depth = position188, tokenIndex188, depth188
			return false
		},
		/* 17 predicate_3 <- <((OP_NOT predicate_3 Action25) / (PAREN_OPEN predicate_1 PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position198, tokenIndex198, depth198 := position, tokenIndex, depth
			{
				position199 := position
				depth++
				{
					position200, tokenIndex200, depth200 := position, tokenIndex, depth
					{
						position202 := position
						depth++
						{
							position203, tokenIndex203, depth203 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l204
							}
							position++
							goto l203
						l204:
							position, tokenIndex, depth = position203, tokenIndex203, depth203
							if buffer[position] != rune('N') {
								goto l201
							}
							position++
						}
					l203:
						{
							position205, tokenIndex205, depth205 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l206
							}
							position++
							goto l205
						l206:
							position, tokenIndex, depth = position205, tokenIndex205, depth205
							if buffer[position] != rune('O') {
								goto l201
							}
							position++
						}
					l205:
						{
							position207, tokenIndex207, depth207 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l208
							}
							position++
							goto l207
						l208:
							position, tokenIndex, depth = position207, tokenIndex207, depth207
							if buffer[position] != rune('T') {
								goto l201
							}
							position++
						}
					l207:
						if !_rules[rule__]() {
							goto l201
						}
						depth--
						add(ruleOP_NOT, position202)
					}
					if !_rules[rulepredicate_3]() {
						goto l201
					}
					{
						add(ruleAction25, position)
					}
					goto l200
				l201:
					position, tokenIndex, depth = position200, tokenIndex200, depth200
					if !_rules[rulePAREN_OPEN]() {
						goto l210
					}
					if !_rules[rulepredicate_1]() {
						goto l210
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l210
					}
					goto l200
				l210:
					position, tokenIndex, depth = position200, tokenIndex200, depth200
					{
						position211 := position
						depth++
						{
							position212, tokenIndex212, depth212 := position, tokenIndex, depth
							if !_rules[ruletagName]() {
								goto l213
							}
							if !_rules[rule_]() {
								goto l213
							}
							if buffer[position] != rune('=') {
								goto l213
							}
							position++
							if !_rules[rule_]() {
								goto l213
							}
							if !_rules[ruleliteralString]() {
								goto l213
							}
							{
								add(ruleAction26, position)
							}
							goto l212
						l213:
							position, tokenIndex, depth = position212, tokenIndex212, depth212
							if !_rules[ruletagName]() {
								goto l215
							}
							if !_rules[rule_]() {
								goto l215
							}
							if buffer[position] != rune('!') {
								goto l215
							}
							position++
							if buffer[position] != rune('=') {
								goto l215
							}
							position++
							if !_rules[rule_]() {
								goto l215
							}
							if !_rules[ruleliteralString]() {
								goto l215
							}
							{
								add(ruleAction27, position)
							}
							goto l212
						l215:
							position, tokenIndex, depth = position212, tokenIndex212, depth212
							if !_rules[ruletagName]() {
								goto l217
							}
							if !_rules[rule__]() {
								goto l217
							}
							{
								position218, tokenIndex218, depth218 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l219
								}
								position++
								goto l218
							l219:
								position, tokenIndex, depth = position218, tokenIndex218, depth218
								if buffer[position] != rune('M') {
									goto l217
								}
								position++
							}
						l218:
							{
								position220, tokenIndex220, depth220 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l221
								}
								position++
								goto l220
							l221:
								position, tokenIndex, depth = position220, tokenIndex220, depth220
								if buffer[position] != rune('A') {
									goto l217
								}
								position++
							}
						l220:
							{
								position222, tokenIndex222, depth222 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l223
								}
								position++
								goto l222
							l223:
								position, tokenIndex, depth = position222, tokenIndex222, depth222
								if buffer[position] != rune('T') {
									goto l217
								}
								position++
							}
						l222:
							{
								position224, tokenIndex224, depth224 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l225
								}
								position++
								goto l224
							l225:
								position, tokenIndex, depth = position224, tokenIndex224, depth224
								if buffer[position] != rune('C') {
									goto l217
								}
								position++
							}
						l224:
							{
								position226, tokenIndex226, depth226 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l227
								}
								position++
								goto l226
							l227:
								position, tokenIndex, depth = position226, tokenIndex226, depth226
								if buffer[position] != rune('H') {
									goto l217
								}
								position++
							}
						l226:
							{
								position228, tokenIndex228, depth228 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l229
								}
								position++
								goto l228
							l229:
								position, tokenIndex, depth = position228, tokenIndex228, depth228
								if buffer[position] != rune('E') {
									goto l217
								}
								position++
							}
						l228:
							{
								position230, tokenIndex230, depth230 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l231
								}
								position++
								goto l230
							l231:
								position, tokenIndex, depth = position230, tokenIndex230, depth230
								if buffer[position] != rune('S') {
									goto l217
								}
								position++
							}
						l230:
							if !_rules[rule__]() {
								goto l217
							}
							if !_rules[ruleliteralString]() {
								goto l217
							}
							{
								add(ruleAction28, position)
							}
							goto l212
						l217:
							position, tokenIndex, depth = position212, tokenIndex212, depth212
							if !_rules[ruletagName]() {
								goto l198
							}
							if !_rules[rule__]() {
								goto l198
							}
							{
								position233, tokenIndex233, depth233 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l234
								}
								position++
								goto l233
							l234:
								position, tokenIndex, depth = position233, tokenIndex233, depth233
								if buffer[position] != rune('I') {
									goto l198
								}
								position++
							}
						l233:
							{
								position235, tokenIndex235, depth235 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l236
								}
								position++
								goto l235
							l236:
								position, tokenIndex, depth = position235, tokenIndex235, depth235
								if buffer[position] != rune('N') {
									goto l198
								}
								position++
							}
						l235:
							if !_rules[rule__]() {
								goto l198
							}
							{
								position237 := position
								depth++
								{
									add(ruleAction31, position)
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l198
								}
								if !_rules[ruleliteralListString]() {
									goto l198
								}
							l239:
								{
									position240, tokenIndex240, depth240 := position, tokenIndex, depth
									if !_rules[ruleCOMMA]() {
										goto l240
									}
									if !_rules[ruleliteralListString]() {
										goto l240
									}
									goto l239
								l240:
									position, tokenIndex, depth = position240, tokenIndex240, depth240
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l198
								}
								depth--
								add(ruleliteralList, position237)
							}
							{
								add(ruleAction29, position)
							}
						}
					l212:
						depth--
						add(ruletagMatcher, position211)
					}
				}
			l200:
				depth--
				add(rulepredicate_3, position199)
			}
			return true
		l198:
			position, tokenIndex, depth = position198, tokenIndex198, depth198
			return false
		},
		/* 18 tagMatcher <- <((tagName _ '=' _ literalString Action26) / (tagName _ ('!' '=') _ literalString Action27) / (tagName __ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) __ literalString Action28) / (tagName __ (('i' / 'I') ('n' / 'N')) __ literalList Action29))> */
		nil,
		/* 19 literalString <- <(<STRING> Action30)> */
		func() bool {
			position243, tokenIndex243, depth243 := position, tokenIndex, depth
			{
				position244 := position
				depth++
				{
					position245 := position
					depth++
					if !_rules[ruleSTRING]() {
						goto l243
					}
					depth--
					add(rulePegText, position245)
				}
				{
					add(ruleAction30, position)
				}
				depth--
				add(ruleliteralString, position244)
			}
			return true
		l243:
			position, tokenIndex, depth = position243, tokenIndex243, depth243
			return false
		},
		/* 20 literalList <- <(Action31 PAREN_OPEN literalListString (COMMA literalListString)* PAREN_CLOSE)> */
		nil,
		/* 21 literalListString <- <(STRING Action32)> */
		func() bool {
			position248, tokenIndex248, depth248 := position, tokenIndex, depth
			{
				position249 := position
				depth++
				if !_rules[ruleSTRING]() {
					goto l248
				}
				{
					add(ruleAction32, position)
				}
				depth--
				add(ruleliteralListString, position249)
			}
			return true
		l248:
			position, tokenIndex, depth = position248, tokenIndex248, depth248
			return false
		},
		/* 22 tagName <- <(<TAG_NAME> Action33)> */
		func() bool {
			position251, tokenIndex251, depth251 := position, tokenIndex, depth
			{
				position252 := position
				depth++
				{
					position253 := position
					depth++
					{
						position254 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l251
						}
						depth--
						add(ruleTAG_NAME, position254)
					}
					depth--
					add(rulePegText, position253)
				}
				{
					add(ruleAction33, position)
				}
				depth--
				add(ruletagName, position252)
			}
			return true
		l251:
			position, tokenIndex, depth = position251, tokenIndex251, depth251
			return false
		},
		/* 23 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position256, tokenIndex256, depth256 := position, tokenIndex, depth
			{
				position257 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l256
				}
				depth--
				add(ruleCOLUMN_NAME, position257)
			}
			return true
		l256:
			position, tokenIndex, depth = position256, tokenIndex256, depth256
			return false
		},
		/* 24 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 25 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 26 TIMESTAMP <- <(NUMBER_NATURAL / STRING)> */
		nil,
		/* 27 IDENTIFIER <- <(('`' CHAR* '`') / (!KEYWORD ID_SEGMENT ('.' ID_SEGMENT)*))> */
		func() bool {
			position261, tokenIndex261, depth261 := position, tokenIndex, depth
			{
				position262 := position
				depth++
				{
					position263, tokenIndex263, depth263 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l264
					}
					position++
				l265:
					{
						position266, tokenIndex266, depth266 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l266
						}
						goto l265
					l266:
						position, tokenIndex, depth = position266, tokenIndex266, depth266
					}
					if buffer[position] != rune('`') {
						goto l264
					}
					position++
					goto l263
				l264:
					position, tokenIndex, depth = position263, tokenIndex263, depth263
					{
						position267, tokenIndex267, depth267 := position, tokenIndex, depth
						{
							position268 := position
							depth++
							{
								position269, tokenIndex269, depth269 := position, tokenIndex, depth
								{
									position271, tokenIndex271, depth271 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l272
									}
									position++
									goto l271
								l272:
									position, tokenIndex, depth = position271, tokenIndex271, depth271
									if buffer[position] != rune('A') {
										goto l270
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
										goto l270
									}
									position++
								}
							l273:
								{
									position275, tokenIndex275, depth275 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l276
									}
									position++
									goto l275
								l276:
									position, tokenIndex, depth = position275, tokenIndex275, depth275
									if buffer[position] != rune('L') {
										goto l270
									}
									position++
								}
							l275:
								goto l269
							l270:
								position, tokenIndex, depth = position269, tokenIndex269, depth269
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
									if buffer[position] != rune('n') {
										goto l281
									}
									position++
									goto l280
								l281:
									position, tokenIndex, depth = position280, tokenIndex280, depth280
									if buffer[position] != rune('N') {
										goto l277
									}
									position++
								}
							l280:
								{
									position282, tokenIndex282, depth282 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l283
									}
									position++
									goto l282
								l283:
									position, tokenIndex, depth = position282, tokenIndex282, depth282
									if buffer[position] != rune('D') {
										goto l277
									}
									position++
								}
							l282:
								goto l269
							l277:
								position, tokenIndex, depth = position269, tokenIndex269, depth269
								{
									position285, tokenIndex285, depth285 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l286
									}
									position++
									goto l285
								l286:
									position, tokenIndex, depth = position285, tokenIndex285, depth285
									if buffer[position] != rune('S') {
										goto l284
									}
									position++
								}
							l285:
								{
									position287, tokenIndex287, depth287 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l288
									}
									position++
									goto l287
								l288:
									position, tokenIndex, depth = position287, tokenIndex287, depth287
									if buffer[position] != rune('E') {
										goto l284
									}
									position++
								}
							l287:
								{
									position289, tokenIndex289, depth289 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l290
									}
									position++
									goto l289
								l290:
									position, tokenIndex, depth = position289, tokenIndex289, depth289
									if buffer[position] != rune('L') {
										goto l284
									}
									position++
								}
							l289:
								{
									position291, tokenIndex291, depth291 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l292
									}
									position++
									goto l291
								l292:
									position, tokenIndex, depth = position291, tokenIndex291, depth291
									if buffer[position] != rune('E') {
										goto l284
									}
									position++
								}
							l291:
								{
									position293, tokenIndex293, depth293 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l294
									}
									position++
									goto l293
								l294:
									position, tokenIndex, depth = position293, tokenIndex293, depth293
									if buffer[position] != rune('C') {
										goto l284
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
										goto l284
									}
									position++
								}
							l295:
								goto l269
							l284:
								position, tokenIndex, depth = position269, tokenIndex269, depth269
								{
									switch buffer[position] {
									case 'W', 'w':
										{
											position298, tokenIndex298, depth298 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l299
											}
											position++
											goto l298
										l299:
											position, tokenIndex, depth = position298, tokenIndex298, depth298
											if buffer[position] != rune('W') {
												goto l267
											}
											position++
										}
									l298:
										{
											position300, tokenIndex300, depth300 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l301
											}
											position++
											goto l300
										l301:
											position, tokenIndex, depth = position300, tokenIndex300, depth300
											if buffer[position] != rune('H') {
												goto l267
											}
											position++
										}
									l300:
										{
											position302, tokenIndex302, depth302 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l303
											}
											position++
											goto l302
										l303:
											position, tokenIndex, depth = position302, tokenIndex302, depth302
											if buffer[position] != rune('E') {
												goto l267
											}
											position++
										}
									l302:
										{
											position304, tokenIndex304, depth304 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l305
											}
											position++
											goto l304
										l305:
											position, tokenIndex, depth = position304, tokenIndex304, depth304
											if buffer[position] != rune('R') {
												goto l267
											}
											position++
										}
									l304:
										{
											position306, tokenIndex306, depth306 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l307
											}
											position++
											goto l306
										l307:
											position, tokenIndex, depth = position306, tokenIndex306, depth306
											if buffer[position] != rune('E') {
												goto l267
											}
											position++
										}
									l306:
										break
									case 'O', 'o':
										{
											position308, tokenIndex308, depth308 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l309
											}
											position++
											goto l308
										l309:
											position, tokenIndex, depth = position308, tokenIndex308, depth308
											if buffer[position] != rune('O') {
												goto l267
											}
											position++
										}
									l308:
										{
											position310, tokenIndex310, depth310 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l311
											}
											position++
											goto l310
										l311:
											position, tokenIndex, depth = position310, tokenIndex310, depth310
											if buffer[position] != rune('R') {
												goto l267
											}
											position++
										}
									l310:
										break
									case 'N', 'n':
										{
											position312, tokenIndex312, depth312 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l313
											}
											position++
											goto l312
										l313:
											position, tokenIndex, depth = position312, tokenIndex312, depth312
											if buffer[position] != rune('N') {
												goto l267
											}
											position++
										}
									l312:
										{
											position314, tokenIndex314, depth314 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l315
											}
											position++
											goto l314
										l315:
											position, tokenIndex, depth = position314, tokenIndex314, depth314
											if buffer[position] != rune('O') {
												goto l267
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
												goto l267
											}
											position++
										}
									l316:
										break
									case 'M', 'm':
										{
											position318, tokenIndex318, depth318 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l319
											}
											position++
											goto l318
										l319:
											position, tokenIndex, depth = position318, tokenIndex318, depth318
											if buffer[position] != rune('M') {
												goto l267
											}
											position++
										}
									l318:
										{
											position320, tokenIndex320, depth320 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l321
											}
											position++
											goto l320
										l321:
											position, tokenIndex, depth = position320, tokenIndex320, depth320
											if buffer[position] != rune('A') {
												goto l267
											}
											position++
										}
									l320:
										{
											position322, tokenIndex322, depth322 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l323
											}
											position++
											goto l322
										l323:
											position, tokenIndex, depth = position322, tokenIndex322, depth322
											if buffer[position] != rune('T') {
												goto l267
											}
											position++
										}
									l322:
										{
											position324, tokenIndex324, depth324 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l325
											}
											position++
											goto l324
										l325:
											position, tokenIndex, depth = position324, tokenIndex324, depth324
											if buffer[position] != rune('C') {
												goto l267
											}
											position++
										}
									l324:
										{
											position326, tokenIndex326, depth326 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l327
											}
											position++
											goto l326
										l327:
											position, tokenIndex, depth = position326, tokenIndex326, depth326
											if buffer[position] != rune('H') {
												goto l267
											}
											position++
										}
									l326:
										{
											position328, tokenIndex328, depth328 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l329
											}
											position++
											goto l328
										l329:
											position, tokenIndex, depth = position328, tokenIndex328, depth328
											if buffer[position] != rune('E') {
												goto l267
											}
											position++
										}
									l328:
										{
											position330, tokenIndex330, depth330 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l331
											}
											position++
											goto l330
										l331:
											position, tokenIndex, depth = position330, tokenIndex330, depth330
											if buffer[position] != rune('S') {
												goto l267
											}
											position++
										}
									l330:
										break
									case 'I', 'i':
										{
											position332, tokenIndex332, depth332 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l333
											}
											position++
											goto l332
										l333:
											position, tokenIndex, depth = position332, tokenIndex332, depth332
											if buffer[position] != rune('I') {
												goto l267
											}
											position++
										}
									l332:
										{
											position334, tokenIndex334, depth334 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l335
											}
											position++
											goto l334
										l335:
											position, tokenIndex, depth = position334, tokenIndex334, depth334
											if buffer[position] != rune('N') {
												goto l267
											}
											position++
										}
									l334:
										break
									case 'G', 'g':
										{
											position336, tokenIndex336, depth336 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l337
											}
											position++
											goto l336
										l337:
											position, tokenIndex, depth = position336, tokenIndex336, depth336
											if buffer[position] != rune('G') {
												goto l267
											}
											position++
										}
									l336:
										{
											position338, tokenIndex338, depth338 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l339
											}
											position++
											goto l338
										l339:
											position, tokenIndex, depth = position338, tokenIndex338, depth338
											if buffer[position] != rune('R') {
												goto l267
											}
											position++
										}
									l338:
										{
											position340, tokenIndex340, depth340 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l341
											}
											position++
											goto l340
										l341:
											position, tokenIndex, depth = position340, tokenIndex340, depth340
											if buffer[position] != rune('O') {
												goto l267
											}
											position++
										}
									l340:
										{
											position342, tokenIndex342, depth342 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l343
											}
											position++
											goto l342
										l343:
											position, tokenIndex, depth = position342, tokenIndex342, depth342
											if buffer[position] != rune('U') {
												goto l267
											}
											position++
										}
									l342:
										{
											position344, tokenIndex344, depth344 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l345
											}
											position++
											goto l344
										l345:
											position, tokenIndex, depth = position344, tokenIndex344, depth344
											if buffer[position] != rune('P') {
												goto l267
											}
											position++
										}
									l344:
										break
									case 'D', 'd':
										{
											position346, tokenIndex346, depth346 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l347
											}
											position++
											goto l346
										l347:
											position, tokenIndex, depth = position346, tokenIndex346, depth346
											if buffer[position] != rune('D') {
												goto l267
											}
											position++
										}
									l346:
										{
											position348, tokenIndex348, depth348 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l349
											}
											position++
											goto l348
										l349:
											position, tokenIndex, depth = position348, tokenIndex348, depth348
											if buffer[position] != rune('E') {
												goto l267
											}
											position++
										}
									l348:
										{
											position350, tokenIndex350, depth350 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l351
											}
											position++
											goto l350
										l351:
											position, tokenIndex, depth = position350, tokenIndex350, depth350
											if buffer[position] != rune('S') {
												goto l267
											}
											position++
										}
									l350:
										{
											position352, tokenIndex352, depth352 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l353
											}
											position++
											goto l352
										l353:
											position, tokenIndex, depth = position352, tokenIndex352, depth352
											if buffer[position] != rune('C') {
												goto l267
											}
											position++
										}
									l352:
										{
											position354, tokenIndex354, depth354 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l355
											}
											position++
											goto l354
										l355:
											position, tokenIndex, depth = position354, tokenIndex354, depth354
											if buffer[position] != rune('R') {
												goto l267
											}
											position++
										}
									l354:
										{
											position356, tokenIndex356, depth356 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l357
											}
											position++
											goto l356
										l357:
											position, tokenIndex, depth = position356, tokenIndex356, depth356
											if buffer[position] != rune('I') {
												goto l267
											}
											position++
										}
									l356:
										{
											position358, tokenIndex358, depth358 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l359
											}
											position++
											goto l358
										l359:
											position, tokenIndex, depth = position358, tokenIndex358, depth358
											if buffer[position] != rune('B') {
												goto l267
											}
											position++
										}
									l358:
										{
											position360, tokenIndex360, depth360 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l361
											}
											position++
											goto l360
										l361:
											position, tokenIndex, depth = position360, tokenIndex360, depth360
											if buffer[position] != rune('E') {
												goto l267
											}
											position++
										}
									l360:
										break
									case 'B', 'b':
										{
											position362, tokenIndex362, depth362 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l363
											}
											position++
											goto l362
										l363:
											position, tokenIndex, depth = position362, tokenIndex362, depth362
											if buffer[position] != rune('B') {
												goto l267
											}
											position++
										}
									l362:
										{
											position364, tokenIndex364, depth364 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l365
											}
											position++
											goto l364
										l365:
											position, tokenIndex, depth = position364, tokenIndex364, depth364
											if buffer[position] != rune('Y') {
												goto l267
											}
											position++
										}
									l364:
										break
									case 'A', 'a':
										{
											position366, tokenIndex366, depth366 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l367
											}
											position++
											goto l366
										l367:
											position, tokenIndex, depth = position366, tokenIndex366, depth366
											if buffer[position] != rune('A') {
												goto l267
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
												goto l267
											}
											position++
										}
									l368:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l267
										}
										break
									}
								}

							}
						l269:
							depth--
							add(ruleKEYWORD, position268)
						}
						goto l261
					l267:
						position, tokenIndex, depth = position267, tokenIndex267, depth267
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l261
					}
				l370:
					{
						position371, tokenIndex371, depth371 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l371
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l371
						}
						goto l370
					l371:
						position, tokenIndex, depth = position371, tokenIndex371, depth371
					}
				}
			l263:
				depth--
				add(ruleIDENTIFIER, position262)
			}
			return true
		l261:
			position, tokenIndex, depth = position261, tokenIndex261, depth261
			return false
		},
		/* 28 ID_SEGMENT <- <(ID_START ID_CONT*)> */
		func() bool {
			position372, tokenIndex372, depth372 := position, tokenIndex, depth
			{
				position373 := position
				depth++
				if !_rules[ruleID_START]() {
					goto l372
				}
			l374:
				{
					position375, tokenIndex375, depth375 := position, tokenIndex, depth
					{
						position376 := position
						depth++
						{
							position377, tokenIndex377, depth377 := position, tokenIndex, depth
							if !_rules[ruleID_START]() {
								goto l378
							}
							goto l377
						l378:
							position, tokenIndex, depth = position377, tokenIndex377, depth377
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l375
							}
							position++
						}
					l377:
						depth--
						add(ruleID_CONT, position376)
					}
					goto l374
				l375:
					position, tokenIndex, depth = position375, tokenIndex375, depth375
				}
				depth--
				add(ruleID_SEGMENT, position373)
			}
			return true
		l372:
			position, tokenIndex, depth = position372, tokenIndex372, depth372
			return false
		},
		/* 29 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position379, tokenIndex379, depth379 := position, tokenIndex, depth
			{
				position380 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l379
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l379
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l379
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position380)
			}
			return true
		l379:
			position, tokenIndex, depth = position379, tokenIndex379, depth379
			return false
		},
		/* 30 ID_CONT <- <(ID_START / [0-9])> */
		nil,
		/* 31 PROPERTY_KEY <- <((&('S' | 's') (('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E') __ (('b' / 'B') ('y' / 'Y')))) | (&('R' | 'r') (('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))) | (&('T' | 't') (('t' / 'T') ('o' / 'O'))) | (&('F' | 'f') (('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))))> */
		func() bool {
			position383, tokenIndex383, depth383 := position, tokenIndex, depth
			{
				position384 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
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
								goto l383
							}
							position++
						}
					l386:
						{
							position388, tokenIndex388, depth388 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l389
							}
							position++
							goto l388
						l389:
							position, tokenIndex, depth = position388, tokenIndex388, depth388
							if buffer[position] != rune('A') {
								goto l383
							}
							position++
						}
					l388:
						{
							position390, tokenIndex390, depth390 := position, tokenIndex, depth
							if buffer[position] != rune('m') {
								goto l391
							}
							position++
							goto l390
						l391:
							position, tokenIndex, depth = position390, tokenIndex390, depth390
							if buffer[position] != rune('M') {
								goto l383
							}
							position++
						}
					l390:
						{
							position392, tokenIndex392, depth392 := position, tokenIndex, depth
							if buffer[position] != rune('p') {
								goto l393
							}
							position++
							goto l392
						l393:
							position, tokenIndex, depth = position392, tokenIndex392, depth392
							if buffer[position] != rune('P') {
								goto l383
							}
							position++
						}
					l392:
						{
							position394, tokenIndex394, depth394 := position, tokenIndex, depth
							if buffer[position] != rune('l') {
								goto l395
							}
							position++
							goto l394
						l395:
							position, tokenIndex, depth = position394, tokenIndex394, depth394
							if buffer[position] != rune('L') {
								goto l383
							}
							position++
						}
					l394:
						{
							position396, tokenIndex396, depth396 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l397
							}
							position++
							goto l396
						l397:
							position, tokenIndex, depth = position396, tokenIndex396, depth396
							if buffer[position] != rune('E') {
								goto l383
							}
							position++
						}
					l396:
						if !_rules[rule__]() {
							goto l383
						}
						{
							position398, tokenIndex398, depth398 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l399
							}
							position++
							goto l398
						l399:
							position, tokenIndex, depth = position398, tokenIndex398, depth398
							if buffer[position] != rune('B') {
								goto l383
							}
							position++
						}
					l398:
						{
							position400, tokenIndex400, depth400 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l401
							}
							position++
							goto l400
						l401:
							position, tokenIndex, depth = position400, tokenIndex400, depth400
							if buffer[position] != rune('Y') {
								goto l383
							}
							position++
						}
					l400:
						break
					case 'R', 'r':
						{
							position402, tokenIndex402, depth402 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l403
							}
							position++
							goto l402
						l403:
							position, tokenIndex, depth = position402, tokenIndex402, depth402
							if buffer[position] != rune('R') {
								goto l383
							}
							position++
						}
					l402:
						{
							position404, tokenIndex404, depth404 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l405
							}
							position++
							goto l404
						l405:
							position, tokenIndex, depth = position404, tokenIndex404, depth404
							if buffer[position] != rune('E') {
								goto l383
							}
							position++
						}
					l404:
						{
							position406, tokenIndex406, depth406 := position, tokenIndex, depth
							if buffer[position] != rune('s') {
								goto l407
							}
							position++
							goto l406
						l407:
							position, tokenIndex, depth = position406, tokenIndex406, depth406
							if buffer[position] != rune('S') {
								goto l383
							}
							position++
						}
					l406:
						{
							position408, tokenIndex408, depth408 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l409
							}
							position++
							goto l408
						l409:
							position, tokenIndex, depth = position408, tokenIndex408, depth408
							if buffer[position] != rune('O') {
								goto l383
							}
							position++
						}
					l408:
						{
							position410, tokenIndex410, depth410 := position, tokenIndex, depth
							if buffer[position] != rune('l') {
								goto l411
							}
							position++
							goto l410
						l411:
							position, tokenIndex, depth = position410, tokenIndex410, depth410
							if buffer[position] != rune('L') {
								goto l383
							}
							position++
						}
					l410:
						{
							position412, tokenIndex412, depth412 := position, tokenIndex, depth
							if buffer[position] != rune('u') {
								goto l413
							}
							position++
							goto l412
						l413:
							position, tokenIndex, depth = position412, tokenIndex412, depth412
							if buffer[position] != rune('U') {
								goto l383
							}
							position++
						}
					l412:
						{
							position414, tokenIndex414, depth414 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l415
							}
							position++
							goto l414
						l415:
							position, tokenIndex, depth = position414, tokenIndex414, depth414
							if buffer[position] != rune('T') {
								goto l383
							}
							position++
						}
					l414:
						{
							position416, tokenIndex416, depth416 := position, tokenIndex, depth
							if buffer[position] != rune('i') {
								goto l417
							}
							position++
							goto l416
						l417:
							position, tokenIndex, depth = position416, tokenIndex416, depth416
							if buffer[position] != rune('I') {
								goto l383
							}
							position++
						}
					l416:
						{
							position418, tokenIndex418, depth418 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l419
							}
							position++
							goto l418
						l419:
							position, tokenIndex, depth = position418, tokenIndex418, depth418
							if buffer[position] != rune('O') {
								goto l383
							}
							position++
						}
					l418:
						{
							position420, tokenIndex420, depth420 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l421
							}
							position++
							goto l420
						l421:
							position, tokenIndex, depth = position420, tokenIndex420, depth420
							if buffer[position] != rune('N') {
								goto l383
							}
							position++
						}
					l420:
						break
					case 'T', 't':
						{
							position422, tokenIndex422, depth422 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l423
							}
							position++
							goto l422
						l423:
							position, tokenIndex, depth = position422, tokenIndex422, depth422
							if buffer[position] != rune('T') {
								goto l383
							}
							position++
						}
					l422:
						{
							position424, tokenIndex424, depth424 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l425
							}
							position++
							goto l424
						l425:
							position, tokenIndex, depth = position424, tokenIndex424, depth424
							if buffer[position] != rune('O') {
								goto l383
							}
							position++
						}
					l424:
						break
					default:
						{
							position426, tokenIndex426, depth426 := position, tokenIndex, depth
							if buffer[position] != rune('f') {
								goto l427
							}
							position++
							goto l426
						l427:
							position, tokenIndex, depth = position426, tokenIndex426, depth426
							if buffer[position] != rune('F') {
								goto l383
							}
							position++
						}
					l426:
						{
							position428, tokenIndex428, depth428 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l429
							}
							position++
							goto l428
						l429:
							position, tokenIndex, depth = position428, tokenIndex428, depth428
							if buffer[position] != rune('R') {
								goto l383
							}
							position++
						}
					l428:
						{
							position430, tokenIndex430, depth430 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l431
							}
							position++
							goto l430
						l431:
							position, tokenIndex, depth = position430, tokenIndex430, depth430
							if buffer[position] != rune('O') {
								goto l383
							}
							position++
						}
					l430:
						{
							position432, tokenIndex432, depth432 := position, tokenIndex, depth
							if buffer[position] != rune('m') {
								goto l433
							}
							position++
							goto l432
						l433:
							position, tokenIndex, depth = position432, tokenIndex432, depth432
							if buffer[position] != rune('M') {
								goto l383
							}
							position++
						}
					l432:
						break
					}
				}

				depth--
				add(rulePROPERTY_KEY, position384)
			}
			return true
		l383:
			position, tokenIndex, depth = position383, tokenIndex383, depth383
			return false
		},
		/* 32 PROPERTY_VALUE <- <TIMESTAMP> */
		nil,
		/* 33 KEYWORD <- <((('a' / 'A') ('l' / 'L') ('l' / 'L')) / (('a' / 'A') ('n' / 'N') ('d' / 'D')) / (('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T')) / ((&('W' | 'w') (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E'))) | (&('O' | 'o') (('o' / 'O') ('r' / 'R'))) | (&('N' | 'n') (('n' / 'N') ('o' / 'O') ('t' / 'T'))) | (&('M' | 'm') (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S'))) | (&('I' | 'i') (('i' / 'I') ('n' / 'N'))) | (&('G' | 'g') (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P'))) | (&('D' | 'd') (('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E'))) | (&('B' | 'b') (('b' / 'B') ('y' / 'Y'))) | (&('A' | 'a') (('a' / 'A') ('s' / 'S'))) | (&('F' | 'R' | 'S' | 'T' | 'f' | 'r' | 's' | 't') PROPERTY_KEY)))> */
		nil,
		/* 34 OP_ADD <- <(_ '+' _)> */
		nil,
		/* 35 OP_SUB <- <(_ '-' _)> */
		nil,
		/* 36 OP_MULT <- <(_ '*' _)> */
		nil,
		/* 37 OP_DIV <- <(_ '/' _)> */
		nil,
		/* 38 OP_AND <- <(_ (('a' / 'A') ('n' / 'N') ('d' / 'D')) _)> */
		nil,
		/* 39 OP_OR <- <(_ (('o' / 'O') ('r' / 'R')) _)> */
		nil,
		/* 40 OP_NOT <- <(('n' / 'N') ('o' / 'O') ('t' / 'T') __)> */
		nil,
		/* 41 STRING <- <(('\'' CHAR* '\'') / ('"' CHAR* '"'))> */
		func() bool {
			position443, tokenIndex443, depth443 := position, tokenIndex, depth
			{
				position444 := position
				depth++
				{
					position445, tokenIndex445, depth445 := position, tokenIndex, depth
					if buffer[position] != rune('\'') {
						goto l446
					}
					position++
				l447:
					{
						position448, tokenIndex448, depth448 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l448
						}
						goto l447
					l448:
						position, tokenIndex, depth = position448, tokenIndex448, depth448
					}
					if buffer[position] != rune('\'') {
						goto l446
					}
					position++
					goto l445
				l446:
					position, tokenIndex, depth = position445, tokenIndex445, depth445
					if buffer[position] != rune('"') {
						goto l443
					}
					position++
				l449:
					{
						position450, tokenIndex450, depth450 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l450
						}
						goto l449
					l450:
						position, tokenIndex, depth = position450, tokenIndex450, depth450
					}
					if buffer[position] != rune('"') {
						goto l443
					}
					position++
				}
			l445:
				depth--
				add(ruleSTRING, position444)
			}
			return true
		l443:
			position, tokenIndex, depth = position443, tokenIndex443, depth443
			return false
		},
		/* 42 CHAR <- <(('\\' ESCAPE_CLASS) / (!ESCAPE_CLASS .))> */
		func() bool {
			position451, tokenIndex451, depth451 := position, tokenIndex, depth
			{
				position452 := position
				depth++
				{
					position453, tokenIndex453, depth453 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l454
					}
					position++
					if !_rules[ruleESCAPE_CLASS]() {
						goto l454
					}
					goto l453
				l454:
					position, tokenIndex, depth = position453, tokenIndex453, depth453
					{
						position455, tokenIndex455, depth455 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l455
						}
						goto l451
					l455:
						position, tokenIndex, depth = position455, tokenIndex455, depth455
					}
					if !matchDot() {
						goto l451
					}
				}
			l453:
				depth--
				add(ruleCHAR, position452)
			}
			return true
		l451:
			position, tokenIndex, depth = position451, tokenIndex451, depth451
			return false
		},
		/* 43 ESCAPE_CLASS <- <((&('\\') '\\') | (&('"') '"') | (&('`') '`') | (&('\'') '\''))> */
		func() bool {
			position456, tokenIndex456, depth456 := position, tokenIndex, depth
			{
				position457 := position
				depth++
				{
					switch buffer[position] {
					case '\\':
						if buffer[position] != rune('\\') {
							goto l456
						}
						position++
						break
					case '"':
						if buffer[position] != rune('"') {
							goto l456
						}
						position++
						break
					case '`':
						if buffer[position] != rune('`') {
							goto l456
						}
						position++
						break
					default:
						if buffer[position] != rune('\'') {
							goto l456
						}
						position++
						break
					}
				}

				depth--
				add(ruleESCAPE_CLASS, position457)
			}
			return true
		l456:
			position, tokenIndex, depth = position456, tokenIndex456, depth456
			return false
		},
		/* 44 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		nil,
		/* 45 NUMBER_NATURAL <- <('0' / ([1-9] [0-9]*))> */
		func() bool {
			position460, tokenIndex460, depth460 := position, tokenIndex, depth
			{
				position461 := position
				depth++
				{
					position462, tokenIndex462, depth462 := position, tokenIndex, depth
					if buffer[position] != rune('0') {
						goto l463
					}
					position++
					goto l462
				l463:
					position, tokenIndex, depth = position462, tokenIndex462, depth462
					if c := buffer[position]; c < rune('1') || c > rune('9') {
						goto l460
					}
					position++
				l464:
					{
						position465, tokenIndex465, depth465 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l465
						}
						position++
						goto l464
					l465:
						position, tokenIndex, depth = position465, tokenIndex465, depth465
					}
				}
			l462:
				depth--
				add(ruleNUMBER_NATURAL, position461)
			}
			return true
		l460:
			position, tokenIndex, depth = position460, tokenIndex460, depth460
			return false
		},
		/* 46 NUMBER_FRACTION <- <('.' [0-9]+)> */
		nil,
		/* 47 NUMBER_INTEGER <- <('-'? NUMBER_NATURAL)> */
		nil,
		/* 48 NUMBER_EXP <- <(('e' / 'E') ('+' / '-')? [0-9]+)> */
		nil,
		/* 49 PAREN_OPEN <- <(_ '(' _)> */
		func() bool {
			position469, tokenIndex469, depth469 := position, tokenIndex, depth
			{
				position470 := position
				depth++
				if !_rules[rule_]() {
					goto l469
				}
				if buffer[position] != rune('(') {
					goto l469
				}
				position++
				if !_rules[rule_]() {
					goto l469
				}
				depth--
				add(rulePAREN_OPEN, position470)
			}
			return true
		l469:
			position, tokenIndex, depth = position469, tokenIndex469, depth469
			return false
		},
		/* 50 PAREN_CLOSE <- <(_ ')' _)> */
		func() bool {
			position471, tokenIndex471, depth471 := position, tokenIndex, depth
			{
				position472 := position
				depth++
				if !_rules[rule_]() {
					goto l471
				}
				if buffer[position] != rune(')') {
					goto l471
				}
				position++
				if !_rules[rule_]() {
					goto l471
				}
				depth--
				add(rulePAREN_CLOSE, position472)
			}
			return true
		l471:
			position, tokenIndex, depth = position471, tokenIndex471, depth471
			return false
		},
		/* 51 COMMA <- <(_ ',' _)> */
		func() bool {
			position473, tokenIndex473, depth473 := position, tokenIndex, depth
			{
				position474 := position
				depth++
				if !_rules[rule_]() {
					goto l473
				}
				if buffer[position] != rune(',') {
					goto l473
				}
				position++
				if !_rules[rule_]() {
					goto l473
				}
				depth--
				add(ruleCOMMA, position474)
			}
			return true
		l473:
			position, tokenIndex, depth = position473, tokenIndex473, depth473
			return false
		},
		/* 52 _ <- <SPACE*> */
		func() bool {
			{
				position476 := position
				depth++
			l477:
				{
					position478, tokenIndex478, depth478 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l478
					}
					goto l477
				l478:
					position, tokenIndex, depth = position478, tokenIndex478, depth478
				}
				depth--
				add(rule_, position476)
			}
			return true
		},
		/* 53 __ <- <SPACE+> */
		func() bool {
			position479, tokenIndex479, depth479 := position, tokenIndex, depth
			{
				position480 := position
				depth++
				if !_rules[ruleSPACE]() {
					goto l479
				}
			l481:
				{
					position482, tokenIndex482, depth482 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l482
					}
					goto l481
				l482:
					position, tokenIndex, depth = position482, tokenIndex482, depth482
				}
				depth--
				add(rule__, position480)
			}
			return true
		l479:
			position, tokenIndex, depth = position479, tokenIndex479, depth479
			return false
		},
		/* 54 SPACE <- <((&('\t') '\t') | (&('\n') '\n') | (&(' ') ' '))> */
		func() bool {
			position483, tokenIndex483, depth483 := position, tokenIndex, depth
			{
				position484 := position
				depth++
				{
					switch buffer[position] {
					case '\t':
						if buffer[position] != rune('\t') {
							goto l483
						}
						position++
						break
					case '\n':
						if buffer[position] != rune('\n') {
							goto l483
						}
						position++
						break
					default:
						if buffer[position] != rune(' ') {
							goto l483
						}
						position++
						break
					}
				}

				depth--
				add(ruleSPACE, position484)
			}
			return true
		l483:
			position, tokenIndex, depth = position483, tokenIndex483, depth483
			return false
		},
		/* 56 Action0 <- <{
		   p.makeSelect()
		 }> */
		nil,
		/* 57 Action1 <- <{ p.makeDescribeAll() }> */
		nil,
		nil,
		/* 59 Action2 <- <{ p.addStringLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 60 Action3 <- <{ p.makeDescribe() }> */
		nil,
		/* 61 Action4 <- <{ p.addNullPredicate() }> */
		nil,
		/* 62 Action5 <- <{ p.addExpressionList() }> */
		nil,
		/* 63 Action6 <- <{ p.appendExpression() }> */
		nil,
		/* 64 Action7 <- <{ p.appendExpression() }> */
		nil,
		/* 65 Action8 <- <{ p.addOperatorLiteral("*") }> */
		nil,
		/* 66 Action9 <- <{ p.addOperatorLiteral("-") }> */
		nil,
		/* 67 Action10 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 68 Action11 <- <{ p.addOperatorLiteral("*") }> */
		nil,
		/* 69 Action12 <- <{ p.addOperatorLiteral("*") }> */
		nil,
		/* 70 Action13 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 71 Action14 <- <{ p.addNumberNode(buffer[begin:end]) }> */
		nil,
		/* 72 Action15 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 73 Action16 <- <{ p.addGroupBy() }> */
		nil,
		/* 74 Action17 <- <{
		   p.addFunctionInvocation()
		 }> */
		nil,
		/* 75 Action18 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 76 Action19 <- <{ p.addNullPredicate() }> */
		nil,
		/* 77 Action20 <- <{
		   p.addMetricExpression()
		 }> */
		nil,
		/* 78 Action21 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 79 Action22 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		   }> */
		nil,
		/* 80 Action23 <- <{ p.addAndPredicate() }> */
		nil,
		/* 81 Action24 <- <{ p.addAndPredicate() }> */
		nil,
		/* 82 Action25 <- <{ p.addNotPredicate() }> */
		nil,
		/* 83 Action26 <- <{
		   p.addLiteralMatcher()
		 }> */
		nil,
		/* 84 Action27 <- <{
		   p.addLiteralMatcher()
		   p.addNotPredicate()
		 }> */
		nil,
		/* 85 Action28 <- <{
		   p.addRegexMatcher()
		 }> */
		nil,
		/* 86 Action29 <- <{
		   p.addListMatcher()
		 }> */
		nil,
		/* 87 Action30 <- <{
		  p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 88 Action31 <- <{ p.addLiteralList() }> */
		nil,
		/* 89 Action32 <- <{
		  p.appendLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 90 Action33 <- <{ p.addTagLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
	}
	p.rules = _rules
}
