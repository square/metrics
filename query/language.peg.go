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
			p.addOrPredicate()
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
		/* 15 predicate_1 <- <((predicate_2 OP_AND predicate_1 Action23) / predicate_2 / )> */
		func() bool {
			{
				position177 := position
				depth++
				{
					position178, tokenIndex178, depth178 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l179
					}
					{
						position180 := position
						depth++
						if !_rules[rule__]() {
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
						if !_rules[rule__]() {
							goto l179
						}
						depth--
						add(ruleOP_AND, position180)
					}
					if !_rules[rulepredicate_1]() {
						goto l179
					}
					{
						add(ruleAction23, position)
					}
					goto l178
				l179:
					position, tokenIndex, depth = position178, tokenIndex178, depth178
					if !_rules[rulepredicate_2]() {
						goto l188
					}
					goto l178
				l188:
					position, tokenIndex, depth = position178, tokenIndex178, depth178
				}
			l178:
				depth--
				add(rulepredicate_1, position177)
			}
			return true
		},
		/* 16 predicate_2 <- <((predicate_3 OP_OR predicate_2 Action24) / predicate_3)> */
		func() bool {
			position189, tokenIndex189, depth189 := position, tokenIndex, depth
			{
				position190 := position
				depth++
				{
					position191, tokenIndex191, depth191 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l192
					}
					{
						position193 := position
						depth++
						if !_rules[rule__]() {
							goto l192
						}
						{
							position194, tokenIndex194, depth194 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l195
							}
							position++
							goto l194
						l195:
							position, tokenIndex, depth = position194, tokenIndex194, depth194
							if buffer[position] != rune('O') {
								goto l192
							}
							position++
						}
					l194:
						{
							position196, tokenIndex196, depth196 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l197
							}
							position++
							goto l196
						l197:
							position, tokenIndex, depth = position196, tokenIndex196, depth196
							if buffer[position] != rune('R') {
								goto l192
							}
							position++
						}
					l196:
						if !_rules[rule__]() {
							goto l192
						}
						depth--
						add(ruleOP_OR, position193)
					}
					if !_rules[rulepredicate_2]() {
						goto l192
					}
					{
						add(ruleAction24, position)
					}
					goto l191
				l192:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
					if !_rules[rulepredicate_3]() {
						goto l189
					}
				}
			l191:
				depth--
				add(rulepredicate_2, position190)
			}
			return true
		l189:
			position, tokenIndex, depth = position189, tokenIndex189, depth189
			return false
		},
		/* 17 predicate_3 <- <((OP_NOT predicate_3 Action25) / (PAREN_OPEN predicate_1 PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position199, tokenIndex199, depth199 := position, tokenIndex, depth
			{
				position200 := position
				depth++
				{
					position201, tokenIndex201, depth201 := position, tokenIndex, depth
					{
						position203 := position
						depth++
						{
							position204, tokenIndex204, depth204 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l205
							}
							position++
							goto l204
						l205:
							position, tokenIndex, depth = position204, tokenIndex204, depth204
							if buffer[position] != rune('N') {
								goto l202
							}
							position++
						}
					l204:
						{
							position206, tokenIndex206, depth206 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l207
							}
							position++
							goto l206
						l207:
							position, tokenIndex, depth = position206, tokenIndex206, depth206
							if buffer[position] != rune('O') {
								goto l202
							}
							position++
						}
					l206:
						{
							position208, tokenIndex208, depth208 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l209
							}
							position++
							goto l208
						l209:
							position, tokenIndex, depth = position208, tokenIndex208, depth208
							if buffer[position] != rune('T') {
								goto l202
							}
							position++
						}
					l208:
						if !_rules[rule__]() {
							goto l202
						}
						depth--
						add(ruleOP_NOT, position203)
					}
					if !_rules[rulepredicate_3]() {
						goto l202
					}
					{
						add(ruleAction25, position)
					}
					goto l201
				l202:
					position, tokenIndex, depth = position201, tokenIndex201, depth201
					if !_rules[rulePAREN_OPEN]() {
						goto l211
					}
					if !_rules[rulepredicate_1]() {
						goto l211
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l211
					}
					goto l201
				l211:
					position, tokenIndex, depth = position201, tokenIndex201, depth201
					{
						position212 := position
						depth++
						{
							position213, tokenIndex213, depth213 := position, tokenIndex, depth
							if !_rules[ruletagName]() {
								goto l214
							}
							if !_rules[rule_]() {
								goto l214
							}
							if buffer[position] != rune('=') {
								goto l214
							}
							position++
							if !_rules[rule_]() {
								goto l214
							}
							if !_rules[ruleliteralString]() {
								goto l214
							}
							{
								add(ruleAction26, position)
							}
							goto l213
						l214:
							position, tokenIndex, depth = position213, tokenIndex213, depth213
							if !_rules[ruletagName]() {
								goto l216
							}
							if !_rules[rule_]() {
								goto l216
							}
							if buffer[position] != rune('!') {
								goto l216
							}
							position++
							if buffer[position] != rune('=') {
								goto l216
							}
							position++
							if !_rules[rule_]() {
								goto l216
							}
							if !_rules[ruleliteralString]() {
								goto l216
							}
							{
								add(ruleAction27, position)
							}
							goto l213
						l216:
							position, tokenIndex, depth = position213, tokenIndex213, depth213
							if !_rules[ruletagName]() {
								goto l218
							}
							if !_rules[rule__]() {
								goto l218
							}
							{
								position219, tokenIndex219, depth219 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l220
								}
								position++
								goto l219
							l220:
								position, tokenIndex, depth = position219, tokenIndex219, depth219
								if buffer[position] != rune('M') {
									goto l218
								}
								position++
							}
						l219:
							{
								position221, tokenIndex221, depth221 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l222
								}
								position++
								goto l221
							l222:
								position, tokenIndex, depth = position221, tokenIndex221, depth221
								if buffer[position] != rune('A') {
									goto l218
								}
								position++
							}
						l221:
							{
								position223, tokenIndex223, depth223 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l224
								}
								position++
								goto l223
							l224:
								position, tokenIndex, depth = position223, tokenIndex223, depth223
								if buffer[position] != rune('T') {
									goto l218
								}
								position++
							}
						l223:
							{
								position225, tokenIndex225, depth225 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l226
								}
								position++
								goto l225
							l226:
								position, tokenIndex, depth = position225, tokenIndex225, depth225
								if buffer[position] != rune('C') {
									goto l218
								}
								position++
							}
						l225:
							{
								position227, tokenIndex227, depth227 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l228
								}
								position++
								goto l227
							l228:
								position, tokenIndex, depth = position227, tokenIndex227, depth227
								if buffer[position] != rune('H') {
									goto l218
								}
								position++
							}
						l227:
							{
								position229, tokenIndex229, depth229 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l230
								}
								position++
								goto l229
							l230:
								position, tokenIndex, depth = position229, tokenIndex229, depth229
								if buffer[position] != rune('E') {
									goto l218
								}
								position++
							}
						l229:
							{
								position231, tokenIndex231, depth231 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l232
								}
								position++
								goto l231
							l232:
								position, tokenIndex, depth = position231, tokenIndex231, depth231
								if buffer[position] != rune('S') {
									goto l218
								}
								position++
							}
						l231:
							if !_rules[rule__]() {
								goto l218
							}
							if !_rules[ruleliteralString]() {
								goto l218
							}
							{
								add(ruleAction28, position)
							}
							goto l213
						l218:
							position, tokenIndex, depth = position213, tokenIndex213, depth213
							if !_rules[ruletagName]() {
								goto l199
							}
							if !_rules[rule__]() {
								goto l199
							}
							{
								position234, tokenIndex234, depth234 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l235
								}
								position++
								goto l234
							l235:
								position, tokenIndex, depth = position234, tokenIndex234, depth234
								if buffer[position] != rune('I') {
									goto l199
								}
								position++
							}
						l234:
							{
								position236, tokenIndex236, depth236 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l237
								}
								position++
								goto l236
							l237:
								position, tokenIndex, depth = position236, tokenIndex236, depth236
								if buffer[position] != rune('N') {
									goto l199
								}
								position++
							}
						l236:
							if !_rules[rule__]() {
								goto l199
							}
							{
								position238 := position
								depth++
								{
									add(ruleAction31, position)
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l199
								}
								if !_rules[ruleliteralListString]() {
									goto l199
								}
							l240:
								{
									position241, tokenIndex241, depth241 := position, tokenIndex, depth
									if !_rules[ruleCOMMA]() {
										goto l241
									}
									if !_rules[ruleliteralListString]() {
										goto l241
									}
									goto l240
								l241:
									position, tokenIndex, depth = position241, tokenIndex241, depth241
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l199
								}
								depth--
								add(ruleliteralList, position238)
							}
							{
								add(ruleAction29, position)
							}
						}
					l213:
						depth--
						add(ruletagMatcher, position212)
					}
				}
			l201:
				depth--
				add(rulepredicate_3, position200)
			}
			return true
		l199:
			position, tokenIndex, depth = position199, tokenIndex199, depth199
			return false
		},
		/* 18 tagMatcher <- <((tagName _ '=' _ literalString Action26) / (tagName _ ('!' '=') _ literalString Action27) / (tagName __ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) __ literalString Action28) / (tagName __ (('i' / 'I') ('n' / 'N')) __ literalList Action29))> */
		nil,
		/* 19 literalString <- <(<STRING> Action30)> */
		func() bool {
			position244, tokenIndex244, depth244 := position, tokenIndex, depth
			{
				position245 := position
				depth++
				{
					position246 := position
					depth++
					if !_rules[ruleSTRING]() {
						goto l244
					}
					depth--
					add(rulePegText, position246)
				}
				{
					add(ruleAction30, position)
				}
				depth--
				add(ruleliteralString, position245)
			}
			return true
		l244:
			position, tokenIndex, depth = position244, tokenIndex244, depth244
			return false
		},
		/* 20 literalList <- <(Action31 PAREN_OPEN literalListString (COMMA literalListString)* PAREN_CLOSE)> */
		nil,
		/* 21 literalListString <- <(STRING Action32)> */
		func() bool {
			position249, tokenIndex249, depth249 := position, tokenIndex, depth
			{
				position250 := position
				depth++
				if !_rules[ruleSTRING]() {
					goto l249
				}
				{
					add(ruleAction32, position)
				}
				depth--
				add(ruleliteralListString, position250)
			}
			return true
		l249:
			position, tokenIndex, depth = position249, tokenIndex249, depth249
			return false
		},
		/* 22 tagName <- <(<TAG_NAME> Action33)> */
		func() bool {
			position252, tokenIndex252, depth252 := position, tokenIndex, depth
			{
				position253 := position
				depth++
				{
					position254 := position
					depth++
					{
						position255 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l252
						}
						depth--
						add(ruleTAG_NAME, position255)
					}
					depth--
					add(rulePegText, position254)
				}
				{
					add(ruleAction33, position)
				}
				depth--
				add(ruletagName, position253)
			}
			return true
		l252:
			position, tokenIndex, depth = position252, tokenIndex252, depth252
			return false
		},
		/* 23 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position257, tokenIndex257, depth257 := position, tokenIndex, depth
			{
				position258 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l257
				}
				depth--
				add(ruleCOLUMN_NAME, position258)
			}
			return true
		l257:
			position, tokenIndex, depth = position257, tokenIndex257, depth257
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
			position262, tokenIndex262, depth262 := position, tokenIndex, depth
			{
				position263 := position
				depth++
				{
					position264, tokenIndex264, depth264 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l265
					}
					position++
				l266:
					{
						position267, tokenIndex267, depth267 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l267
						}
						goto l266
					l267:
						position, tokenIndex, depth = position267, tokenIndex267, depth267
					}
					if buffer[position] != rune('`') {
						goto l265
					}
					position++
					goto l264
				l265:
					position, tokenIndex, depth = position264, tokenIndex264, depth264
					{
						position268, tokenIndex268, depth268 := position, tokenIndex, depth
						{
							position269 := position
							depth++
							{
								position270, tokenIndex270, depth270 := position, tokenIndex, depth
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
										goto l271
									}
									position++
								}
							l272:
								{
									position274, tokenIndex274, depth274 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l275
									}
									position++
									goto l274
								l275:
									position, tokenIndex, depth = position274, tokenIndex274, depth274
									if buffer[position] != rune('L') {
										goto l271
									}
									position++
								}
							l274:
								{
									position276, tokenIndex276, depth276 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l277
									}
									position++
									goto l276
								l277:
									position, tokenIndex, depth = position276, tokenIndex276, depth276
									if buffer[position] != rune('L') {
										goto l271
									}
									position++
								}
							l276:
								goto l270
							l271:
								position, tokenIndex, depth = position270, tokenIndex270, depth270
								{
									position279, tokenIndex279, depth279 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l280
									}
									position++
									goto l279
								l280:
									position, tokenIndex, depth = position279, tokenIndex279, depth279
									if buffer[position] != rune('A') {
										goto l278
									}
									position++
								}
							l279:
								{
									position281, tokenIndex281, depth281 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l282
									}
									position++
									goto l281
								l282:
									position, tokenIndex, depth = position281, tokenIndex281, depth281
									if buffer[position] != rune('N') {
										goto l278
									}
									position++
								}
							l281:
								{
									position283, tokenIndex283, depth283 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l284
									}
									position++
									goto l283
								l284:
									position, tokenIndex, depth = position283, tokenIndex283, depth283
									if buffer[position] != rune('D') {
										goto l278
									}
									position++
								}
							l283:
								goto l270
							l278:
								position, tokenIndex, depth = position270, tokenIndex270, depth270
								{
									position286, tokenIndex286, depth286 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l287
									}
									position++
									goto l286
								l287:
									position, tokenIndex, depth = position286, tokenIndex286, depth286
									if buffer[position] != rune('S') {
										goto l285
									}
									position++
								}
							l286:
								{
									position288, tokenIndex288, depth288 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l289
									}
									position++
									goto l288
								l289:
									position, tokenIndex, depth = position288, tokenIndex288, depth288
									if buffer[position] != rune('E') {
										goto l285
									}
									position++
								}
							l288:
								{
									position290, tokenIndex290, depth290 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l291
									}
									position++
									goto l290
								l291:
									position, tokenIndex, depth = position290, tokenIndex290, depth290
									if buffer[position] != rune('L') {
										goto l285
									}
									position++
								}
							l290:
								{
									position292, tokenIndex292, depth292 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l293
									}
									position++
									goto l292
								l293:
									position, tokenIndex, depth = position292, tokenIndex292, depth292
									if buffer[position] != rune('E') {
										goto l285
									}
									position++
								}
							l292:
								{
									position294, tokenIndex294, depth294 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l295
									}
									position++
									goto l294
								l295:
									position, tokenIndex, depth = position294, tokenIndex294, depth294
									if buffer[position] != rune('C') {
										goto l285
									}
									position++
								}
							l294:
								{
									position296, tokenIndex296, depth296 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l297
									}
									position++
									goto l296
								l297:
									position, tokenIndex, depth = position296, tokenIndex296, depth296
									if buffer[position] != rune('T') {
										goto l285
									}
									position++
								}
							l296:
								goto l270
							l285:
								position, tokenIndex, depth = position270, tokenIndex270, depth270
								{
									switch buffer[position] {
									case 'W', 'w':
										{
											position299, tokenIndex299, depth299 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l300
											}
											position++
											goto l299
										l300:
											position, tokenIndex, depth = position299, tokenIndex299, depth299
											if buffer[position] != rune('W') {
												goto l268
											}
											position++
										}
									l299:
										{
											position301, tokenIndex301, depth301 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l302
											}
											position++
											goto l301
										l302:
											position, tokenIndex, depth = position301, tokenIndex301, depth301
											if buffer[position] != rune('H') {
												goto l268
											}
											position++
										}
									l301:
										{
											position303, tokenIndex303, depth303 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l304
											}
											position++
											goto l303
										l304:
											position, tokenIndex, depth = position303, tokenIndex303, depth303
											if buffer[position] != rune('E') {
												goto l268
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
												goto l268
											}
											position++
										}
									l305:
										{
											position307, tokenIndex307, depth307 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l308
											}
											position++
											goto l307
										l308:
											position, tokenIndex, depth = position307, tokenIndex307, depth307
											if buffer[position] != rune('E') {
												goto l268
											}
											position++
										}
									l307:
										break
									case 'O', 'o':
										{
											position309, tokenIndex309, depth309 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l310
											}
											position++
											goto l309
										l310:
											position, tokenIndex, depth = position309, tokenIndex309, depth309
											if buffer[position] != rune('O') {
												goto l268
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
												goto l268
											}
											position++
										}
									l311:
										break
									case 'N', 'n':
										{
											position313, tokenIndex313, depth313 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l314
											}
											position++
											goto l313
										l314:
											position, tokenIndex, depth = position313, tokenIndex313, depth313
											if buffer[position] != rune('N') {
												goto l268
											}
											position++
										}
									l313:
										{
											position315, tokenIndex315, depth315 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l316
											}
											position++
											goto l315
										l316:
											position, tokenIndex, depth = position315, tokenIndex315, depth315
											if buffer[position] != rune('O') {
												goto l268
											}
											position++
										}
									l315:
										{
											position317, tokenIndex317, depth317 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l318
											}
											position++
											goto l317
										l318:
											position, tokenIndex, depth = position317, tokenIndex317, depth317
											if buffer[position] != rune('T') {
												goto l268
											}
											position++
										}
									l317:
										break
									case 'M', 'm':
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
												goto l268
											}
											position++
										}
									l319:
										{
											position321, tokenIndex321, depth321 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l322
											}
											position++
											goto l321
										l322:
											position, tokenIndex, depth = position321, tokenIndex321, depth321
											if buffer[position] != rune('A') {
												goto l268
											}
											position++
										}
									l321:
										{
											position323, tokenIndex323, depth323 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l324
											}
											position++
											goto l323
										l324:
											position, tokenIndex, depth = position323, tokenIndex323, depth323
											if buffer[position] != rune('T') {
												goto l268
											}
											position++
										}
									l323:
										{
											position325, tokenIndex325, depth325 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l326
											}
											position++
											goto l325
										l326:
											position, tokenIndex, depth = position325, tokenIndex325, depth325
											if buffer[position] != rune('C') {
												goto l268
											}
											position++
										}
									l325:
										{
											position327, tokenIndex327, depth327 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l328
											}
											position++
											goto l327
										l328:
											position, tokenIndex, depth = position327, tokenIndex327, depth327
											if buffer[position] != rune('H') {
												goto l268
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
												goto l268
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
												goto l268
											}
											position++
										}
									l331:
										break
									case 'I', 'i':
										{
											position333, tokenIndex333, depth333 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l334
											}
											position++
											goto l333
										l334:
											position, tokenIndex, depth = position333, tokenIndex333, depth333
											if buffer[position] != rune('I') {
												goto l268
											}
											position++
										}
									l333:
										{
											position335, tokenIndex335, depth335 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l336
											}
											position++
											goto l335
										l336:
											position, tokenIndex, depth = position335, tokenIndex335, depth335
											if buffer[position] != rune('N') {
												goto l268
											}
											position++
										}
									l335:
										break
									case 'G', 'g':
										{
											position337, tokenIndex337, depth337 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l338
											}
											position++
											goto l337
										l338:
											position, tokenIndex, depth = position337, tokenIndex337, depth337
											if buffer[position] != rune('G') {
												goto l268
											}
											position++
										}
									l337:
										{
											position339, tokenIndex339, depth339 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l340
											}
											position++
											goto l339
										l340:
											position, tokenIndex, depth = position339, tokenIndex339, depth339
											if buffer[position] != rune('R') {
												goto l268
											}
											position++
										}
									l339:
										{
											position341, tokenIndex341, depth341 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l342
											}
											position++
											goto l341
										l342:
											position, tokenIndex, depth = position341, tokenIndex341, depth341
											if buffer[position] != rune('O') {
												goto l268
											}
											position++
										}
									l341:
										{
											position343, tokenIndex343, depth343 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l344
											}
											position++
											goto l343
										l344:
											position, tokenIndex, depth = position343, tokenIndex343, depth343
											if buffer[position] != rune('U') {
												goto l268
											}
											position++
										}
									l343:
										{
											position345, tokenIndex345, depth345 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l346
											}
											position++
											goto l345
										l346:
											position, tokenIndex, depth = position345, tokenIndex345, depth345
											if buffer[position] != rune('P') {
												goto l268
											}
											position++
										}
									l345:
										break
									case 'D', 'd':
										{
											position347, tokenIndex347, depth347 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l348
											}
											position++
											goto l347
										l348:
											position, tokenIndex, depth = position347, tokenIndex347, depth347
											if buffer[position] != rune('D') {
												goto l268
											}
											position++
										}
									l347:
										{
											position349, tokenIndex349, depth349 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l350
											}
											position++
											goto l349
										l350:
											position, tokenIndex, depth = position349, tokenIndex349, depth349
											if buffer[position] != rune('E') {
												goto l268
											}
											position++
										}
									l349:
										{
											position351, tokenIndex351, depth351 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l352
											}
											position++
											goto l351
										l352:
											position, tokenIndex, depth = position351, tokenIndex351, depth351
											if buffer[position] != rune('S') {
												goto l268
											}
											position++
										}
									l351:
										{
											position353, tokenIndex353, depth353 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l354
											}
											position++
											goto l353
										l354:
											position, tokenIndex, depth = position353, tokenIndex353, depth353
											if buffer[position] != rune('C') {
												goto l268
											}
											position++
										}
									l353:
										{
											position355, tokenIndex355, depth355 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l356
											}
											position++
											goto l355
										l356:
											position, tokenIndex, depth = position355, tokenIndex355, depth355
											if buffer[position] != rune('R') {
												goto l268
											}
											position++
										}
									l355:
										{
											position357, tokenIndex357, depth357 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l358
											}
											position++
											goto l357
										l358:
											position, tokenIndex, depth = position357, tokenIndex357, depth357
											if buffer[position] != rune('I') {
												goto l268
											}
											position++
										}
									l357:
										{
											position359, tokenIndex359, depth359 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l360
											}
											position++
											goto l359
										l360:
											position, tokenIndex, depth = position359, tokenIndex359, depth359
											if buffer[position] != rune('B') {
												goto l268
											}
											position++
										}
									l359:
										{
											position361, tokenIndex361, depth361 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l362
											}
											position++
											goto l361
										l362:
											position, tokenIndex, depth = position361, tokenIndex361, depth361
											if buffer[position] != rune('E') {
												goto l268
											}
											position++
										}
									l361:
										break
									case 'B', 'b':
										{
											position363, tokenIndex363, depth363 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l364
											}
											position++
											goto l363
										l364:
											position, tokenIndex, depth = position363, tokenIndex363, depth363
											if buffer[position] != rune('B') {
												goto l268
											}
											position++
										}
									l363:
										{
											position365, tokenIndex365, depth365 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l366
											}
											position++
											goto l365
										l366:
											position, tokenIndex, depth = position365, tokenIndex365, depth365
											if buffer[position] != rune('Y') {
												goto l268
											}
											position++
										}
									l365:
										break
									case 'A', 'a':
										{
											position367, tokenIndex367, depth367 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l368
											}
											position++
											goto l367
										l368:
											position, tokenIndex, depth = position367, tokenIndex367, depth367
											if buffer[position] != rune('A') {
												goto l268
											}
											position++
										}
									l367:
										{
											position369, tokenIndex369, depth369 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l370
											}
											position++
											goto l369
										l370:
											position, tokenIndex, depth = position369, tokenIndex369, depth369
											if buffer[position] != rune('S') {
												goto l268
											}
											position++
										}
									l369:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l268
										}
										break
									}
								}

							}
						l270:
							depth--
							add(ruleKEYWORD, position269)
						}
						goto l262
					l268:
						position, tokenIndex, depth = position268, tokenIndex268, depth268
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l262
					}
				l371:
					{
						position372, tokenIndex372, depth372 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l372
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l372
						}
						goto l371
					l372:
						position, tokenIndex, depth = position372, tokenIndex372, depth372
					}
				}
			l264:
				depth--
				add(ruleIDENTIFIER, position263)
			}
			return true
		l262:
			position, tokenIndex, depth = position262, tokenIndex262, depth262
			return false
		},
		/* 28 ID_SEGMENT <- <(ID_START ID_CONT*)> */
		func() bool {
			position373, tokenIndex373, depth373 := position, tokenIndex, depth
			{
				position374 := position
				depth++
				if !_rules[ruleID_START]() {
					goto l373
				}
			l375:
				{
					position376, tokenIndex376, depth376 := position, tokenIndex, depth
					{
						position377 := position
						depth++
						{
							position378, tokenIndex378, depth378 := position, tokenIndex, depth
							if !_rules[ruleID_START]() {
								goto l379
							}
							goto l378
						l379:
							position, tokenIndex, depth = position378, tokenIndex378, depth378
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l376
							}
							position++
						}
					l378:
						depth--
						add(ruleID_CONT, position377)
					}
					goto l375
				l376:
					position, tokenIndex, depth = position376, tokenIndex376, depth376
				}
				depth--
				add(ruleID_SEGMENT, position374)
			}
			return true
		l373:
			position, tokenIndex, depth = position373, tokenIndex373, depth373
			return false
		},
		/* 29 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position380, tokenIndex380, depth380 := position, tokenIndex, depth
			{
				position381 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l380
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l380
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l380
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position381)
			}
			return true
		l380:
			position, tokenIndex, depth = position380, tokenIndex380, depth380
			return false
		},
		/* 30 ID_CONT <- <(ID_START / [0-9])> */
		nil,
		/* 31 PROPERTY_KEY <- <((&('S' | 's') (('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E') __ (('b' / 'B') ('y' / 'Y')))) | (&('R' | 'r') (('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))) | (&('T' | 't') (('t' / 'T') ('o' / 'O'))) | (&('F' | 'f') (('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))))> */
		func() bool {
			position384, tokenIndex384, depth384 := position, tokenIndex, depth
			{
				position385 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position387, tokenIndex387, depth387 := position, tokenIndex, depth
							if buffer[position] != rune('s') {
								goto l388
							}
							position++
							goto l387
						l388:
							position, tokenIndex, depth = position387, tokenIndex387, depth387
							if buffer[position] != rune('S') {
								goto l384
							}
							position++
						}
					l387:
						{
							position389, tokenIndex389, depth389 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l390
							}
							position++
							goto l389
						l390:
							position, tokenIndex, depth = position389, tokenIndex389, depth389
							if buffer[position] != rune('A') {
								goto l384
							}
							position++
						}
					l389:
						{
							position391, tokenIndex391, depth391 := position, tokenIndex, depth
							if buffer[position] != rune('m') {
								goto l392
							}
							position++
							goto l391
						l392:
							position, tokenIndex, depth = position391, tokenIndex391, depth391
							if buffer[position] != rune('M') {
								goto l384
							}
							position++
						}
					l391:
						{
							position393, tokenIndex393, depth393 := position, tokenIndex, depth
							if buffer[position] != rune('p') {
								goto l394
							}
							position++
							goto l393
						l394:
							position, tokenIndex, depth = position393, tokenIndex393, depth393
							if buffer[position] != rune('P') {
								goto l384
							}
							position++
						}
					l393:
						{
							position395, tokenIndex395, depth395 := position, tokenIndex, depth
							if buffer[position] != rune('l') {
								goto l396
							}
							position++
							goto l395
						l396:
							position, tokenIndex, depth = position395, tokenIndex395, depth395
							if buffer[position] != rune('L') {
								goto l384
							}
							position++
						}
					l395:
						{
							position397, tokenIndex397, depth397 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l398
							}
							position++
							goto l397
						l398:
							position, tokenIndex, depth = position397, tokenIndex397, depth397
							if buffer[position] != rune('E') {
								goto l384
							}
							position++
						}
					l397:
						if !_rules[rule__]() {
							goto l384
						}
						{
							position399, tokenIndex399, depth399 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l400
							}
							position++
							goto l399
						l400:
							position, tokenIndex, depth = position399, tokenIndex399, depth399
							if buffer[position] != rune('B') {
								goto l384
							}
							position++
						}
					l399:
						{
							position401, tokenIndex401, depth401 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l402
							}
							position++
							goto l401
						l402:
							position, tokenIndex, depth = position401, tokenIndex401, depth401
							if buffer[position] != rune('Y') {
								goto l384
							}
							position++
						}
					l401:
						break
					case 'R', 'r':
						{
							position403, tokenIndex403, depth403 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l404
							}
							position++
							goto l403
						l404:
							position, tokenIndex, depth = position403, tokenIndex403, depth403
							if buffer[position] != rune('R') {
								goto l384
							}
							position++
						}
					l403:
						{
							position405, tokenIndex405, depth405 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l406
							}
							position++
							goto l405
						l406:
							position, tokenIndex, depth = position405, tokenIndex405, depth405
							if buffer[position] != rune('E') {
								goto l384
							}
							position++
						}
					l405:
						{
							position407, tokenIndex407, depth407 := position, tokenIndex, depth
							if buffer[position] != rune('s') {
								goto l408
							}
							position++
							goto l407
						l408:
							position, tokenIndex, depth = position407, tokenIndex407, depth407
							if buffer[position] != rune('S') {
								goto l384
							}
							position++
						}
					l407:
						{
							position409, tokenIndex409, depth409 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l410
							}
							position++
							goto l409
						l410:
							position, tokenIndex, depth = position409, tokenIndex409, depth409
							if buffer[position] != rune('O') {
								goto l384
							}
							position++
						}
					l409:
						{
							position411, tokenIndex411, depth411 := position, tokenIndex, depth
							if buffer[position] != rune('l') {
								goto l412
							}
							position++
							goto l411
						l412:
							position, tokenIndex, depth = position411, tokenIndex411, depth411
							if buffer[position] != rune('L') {
								goto l384
							}
							position++
						}
					l411:
						{
							position413, tokenIndex413, depth413 := position, tokenIndex, depth
							if buffer[position] != rune('u') {
								goto l414
							}
							position++
							goto l413
						l414:
							position, tokenIndex, depth = position413, tokenIndex413, depth413
							if buffer[position] != rune('U') {
								goto l384
							}
							position++
						}
					l413:
						{
							position415, tokenIndex415, depth415 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l416
							}
							position++
							goto l415
						l416:
							position, tokenIndex, depth = position415, tokenIndex415, depth415
							if buffer[position] != rune('T') {
								goto l384
							}
							position++
						}
					l415:
						{
							position417, tokenIndex417, depth417 := position, tokenIndex, depth
							if buffer[position] != rune('i') {
								goto l418
							}
							position++
							goto l417
						l418:
							position, tokenIndex, depth = position417, tokenIndex417, depth417
							if buffer[position] != rune('I') {
								goto l384
							}
							position++
						}
					l417:
						{
							position419, tokenIndex419, depth419 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l420
							}
							position++
							goto l419
						l420:
							position, tokenIndex, depth = position419, tokenIndex419, depth419
							if buffer[position] != rune('O') {
								goto l384
							}
							position++
						}
					l419:
						{
							position421, tokenIndex421, depth421 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l422
							}
							position++
							goto l421
						l422:
							position, tokenIndex, depth = position421, tokenIndex421, depth421
							if buffer[position] != rune('N') {
								goto l384
							}
							position++
						}
					l421:
						break
					case 'T', 't':
						{
							position423, tokenIndex423, depth423 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l424
							}
							position++
							goto l423
						l424:
							position, tokenIndex, depth = position423, tokenIndex423, depth423
							if buffer[position] != rune('T') {
								goto l384
							}
							position++
						}
					l423:
						{
							position425, tokenIndex425, depth425 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l426
							}
							position++
							goto l425
						l426:
							position, tokenIndex, depth = position425, tokenIndex425, depth425
							if buffer[position] != rune('O') {
								goto l384
							}
							position++
						}
					l425:
						break
					default:
						{
							position427, tokenIndex427, depth427 := position, tokenIndex, depth
							if buffer[position] != rune('f') {
								goto l428
							}
							position++
							goto l427
						l428:
							position, tokenIndex, depth = position427, tokenIndex427, depth427
							if buffer[position] != rune('F') {
								goto l384
							}
							position++
						}
					l427:
						{
							position429, tokenIndex429, depth429 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l430
							}
							position++
							goto l429
						l430:
							position, tokenIndex, depth = position429, tokenIndex429, depth429
							if buffer[position] != rune('R') {
								goto l384
							}
							position++
						}
					l429:
						{
							position431, tokenIndex431, depth431 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l432
							}
							position++
							goto l431
						l432:
							position, tokenIndex, depth = position431, tokenIndex431, depth431
							if buffer[position] != rune('O') {
								goto l384
							}
							position++
						}
					l431:
						{
							position433, tokenIndex433, depth433 := position, tokenIndex, depth
							if buffer[position] != rune('m') {
								goto l434
							}
							position++
							goto l433
						l434:
							position, tokenIndex, depth = position433, tokenIndex433, depth433
							if buffer[position] != rune('M') {
								goto l384
							}
							position++
						}
					l433:
						break
					}
				}

				depth--
				add(rulePROPERTY_KEY, position385)
			}
			return true
		l384:
			position, tokenIndex, depth = position384, tokenIndex384, depth384
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
		/* 38 OP_AND <- <(__ (('a' / 'A') ('n' / 'N') ('d' / 'D')) __)> */
		nil,
		/* 39 OP_OR <- <(__ (('o' / 'O') ('r' / 'R')) __)> */
		nil,
		/* 40 OP_NOT <- <(('n' / 'N') ('o' / 'O') ('t' / 'T') __)> */
		nil,
		/* 41 STRING <- <(('\'' CHAR* '\'') / ('"' CHAR* '"'))> */
		func() bool {
			position444, tokenIndex444, depth444 := position, tokenIndex, depth
			{
				position445 := position
				depth++
				{
					position446, tokenIndex446, depth446 := position, tokenIndex, depth
					if buffer[position] != rune('\'') {
						goto l447
					}
					position++
				l448:
					{
						position449, tokenIndex449, depth449 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l449
						}
						goto l448
					l449:
						position, tokenIndex, depth = position449, tokenIndex449, depth449
					}
					if buffer[position] != rune('\'') {
						goto l447
					}
					position++
					goto l446
				l447:
					position, tokenIndex, depth = position446, tokenIndex446, depth446
					if buffer[position] != rune('"') {
						goto l444
					}
					position++
				l450:
					{
						position451, tokenIndex451, depth451 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l451
						}
						goto l450
					l451:
						position, tokenIndex, depth = position451, tokenIndex451, depth451
					}
					if buffer[position] != rune('"') {
						goto l444
					}
					position++
				}
			l446:
				depth--
				add(ruleSTRING, position445)
			}
			return true
		l444:
			position, tokenIndex, depth = position444, tokenIndex444, depth444
			return false
		},
		/* 42 CHAR <- <(('\\' ESCAPE_CLASS) / (!ESCAPE_CLASS .))> */
		func() bool {
			position452, tokenIndex452, depth452 := position, tokenIndex, depth
			{
				position453 := position
				depth++
				{
					position454, tokenIndex454, depth454 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l455
					}
					position++
					if !_rules[ruleESCAPE_CLASS]() {
						goto l455
					}
					goto l454
				l455:
					position, tokenIndex, depth = position454, tokenIndex454, depth454
					{
						position456, tokenIndex456, depth456 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l456
						}
						goto l452
					l456:
						position, tokenIndex, depth = position456, tokenIndex456, depth456
					}
					if !matchDot() {
						goto l452
					}
				}
			l454:
				depth--
				add(ruleCHAR, position453)
			}
			return true
		l452:
			position, tokenIndex, depth = position452, tokenIndex452, depth452
			return false
		},
		/* 43 ESCAPE_CLASS <- <((&('\\') '\\') | (&('"') '"') | (&('`') '`') | (&('\'') '\''))> */
		func() bool {
			position457, tokenIndex457, depth457 := position, tokenIndex, depth
			{
				position458 := position
				depth++
				{
					switch buffer[position] {
					case '\\':
						if buffer[position] != rune('\\') {
							goto l457
						}
						position++
						break
					case '"':
						if buffer[position] != rune('"') {
							goto l457
						}
						position++
						break
					case '`':
						if buffer[position] != rune('`') {
							goto l457
						}
						position++
						break
					default:
						if buffer[position] != rune('\'') {
							goto l457
						}
						position++
						break
					}
				}

				depth--
				add(ruleESCAPE_CLASS, position458)
			}
			return true
		l457:
			position, tokenIndex, depth = position457, tokenIndex457, depth457
			return false
		},
		/* 44 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		nil,
		/* 45 NUMBER_NATURAL <- <('0' / ([1-9] [0-9]*))> */
		func() bool {
			position461, tokenIndex461, depth461 := position, tokenIndex, depth
			{
				position462 := position
				depth++
				{
					position463, tokenIndex463, depth463 := position, tokenIndex, depth
					if buffer[position] != rune('0') {
						goto l464
					}
					position++
					goto l463
				l464:
					position, tokenIndex, depth = position463, tokenIndex463, depth463
					if c := buffer[position]; c < rune('1') || c > rune('9') {
						goto l461
					}
					position++
				l465:
					{
						position466, tokenIndex466, depth466 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l466
						}
						position++
						goto l465
					l466:
						position, tokenIndex, depth = position466, tokenIndex466, depth466
					}
				}
			l463:
				depth--
				add(ruleNUMBER_NATURAL, position462)
			}
			return true
		l461:
			position, tokenIndex, depth = position461, tokenIndex461, depth461
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
			position470, tokenIndex470, depth470 := position, tokenIndex, depth
			{
				position471 := position
				depth++
				if !_rules[rule_]() {
					goto l470
				}
				if buffer[position] != rune('(') {
					goto l470
				}
				position++
				if !_rules[rule_]() {
					goto l470
				}
				depth--
				add(rulePAREN_OPEN, position471)
			}
			return true
		l470:
			position, tokenIndex, depth = position470, tokenIndex470, depth470
			return false
		},
		/* 50 PAREN_CLOSE <- <(_ ')' _)> */
		func() bool {
			position472, tokenIndex472, depth472 := position, tokenIndex, depth
			{
				position473 := position
				depth++
				if !_rules[rule_]() {
					goto l472
				}
				if buffer[position] != rune(')') {
					goto l472
				}
				position++
				if !_rules[rule_]() {
					goto l472
				}
				depth--
				add(rulePAREN_CLOSE, position473)
			}
			return true
		l472:
			position, tokenIndex, depth = position472, tokenIndex472, depth472
			return false
		},
		/* 51 COMMA <- <(_ ',' _)> */
		func() bool {
			position474, tokenIndex474, depth474 := position, tokenIndex, depth
			{
				position475 := position
				depth++
				if !_rules[rule_]() {
					goto l474
				}
				if buffer[position] != rune(',') {
					goto l474
				}
				position++
				if !_rules[rule_]() {
					goto l474
				}
				depth--
				add(ruleCOMMA, position475)
			}
			return true
		l474:
			position, tokenIndex, depth = position474, tokenIndex474, depth474
			return false
		},
		/* 52 _ <- <SPACE*> */
		func() bool {
			{
				position477 := position
				depth++
			l478:
				{
					position479, tokenIndex479, depth479 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l479
					}
					goto l478
				l479:
					position, tokenIndex, depth = position479, tokenIndex479, depth479
				}
				depth--
				add(rule_, position477)
			}
			return true
		},
		/* 53 __ <- <SPACE+> */
		func() bool {
			position480, tokenIndex480, depth480 := position, tokenIndex, depth
			{
				position481 := position
				depth++
				if !_rules[ruleSPACE]() {
					goto l480
				}
			l482:
				{
					position483, tokenIndex483, depth483 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l483
					}
					goto l482
				l483:
					position, tokenIndex, depth = position483, tokenIndex483, depth483
				}
				depth--
				add(rule__, position481)
			}
			return true
		l480:
			position, tokenIndex, depth = position480, tokenIndex480, depth480
			return false
		},
		/* 54 SPACE <- <((&('\t') '\t') | (&('\n') '\n') | (&(' ') ' '))> */
		func() bool {
			position484, tokenIndex484, depth484 := position, tokenIndex, depth
			{
				position485 := position
				depth++
				{
					switch buffer[position] {
					case '\t':
						if buffer[position] != rune('\t') {
							goto l484
						}
						position++
						break
					case '\n':
						if buffer[position] != rune('\n') {
							goto l484
						}
						position++
						break
					default:
						if buffer[position] != rune(' ') {
							goto l484
						}
						position++
						break
					}
				}

				depth--
				add(ruleSPACE, position485)
			}
			return true
		l484:
			position, tokenIndex, depth = position484, tokenIndex484, depth484
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
		/* 81 Action24 <- <{ p.addOrPredicate() }> */
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
