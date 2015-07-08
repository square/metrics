package query

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const end_symbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleroot
	ruleselectStmt
	ruledescribeStmt
	ruledescribeAllStmt
	ruledescribeMetrics
	ruledescribeSingleStmt
	rulepropertyClause
	ruleoptionalPredicateClause
	ruleexpressionList
	ruleexpression_start
	ruleexpression_sum
	ruleexpression_product
	ruleadd_pipe
	ruleexpression_atom
	ruleoptionalGroupBy
	ruleexpression_function
	ruleexpression_metric
	rulegroupByClause
	rulecollapseByClause
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
	ruleIDENTIFIER
	ruleTIMESTAMP
	ruleID_SEGMENT
	ruleID_START
	ruleID_CONT
	rulePROPERTY_KEY
	rulePROPERTY_VALUE
	ruleKEYWORD
	ruleOP_PIPE
	ruleOP_ADD
	ruleOP_SUB
	ruleOP_MULT
	ruleOP_DIV
	ruleOP_AND
	ruleOP_OR
	ruleOP_NOT
	ruleQUOTE_SINGLE
	ruleQUOTE_DOUBLE
	ruleSTRING
	ruleCHAR
	ruleESCAPE_CLASS
	ruleNUMBER
	ruleNUMBER_NATURAL
	ruleNUMBER_FRACTION
	ruleNUMBER_INTEGER
	ruleNUMBER_EXP
	ruleDURATION
	rulePAREN_OPEN
	rulePAREN_CLOSE
	ruleCOMMA
	rule_
	ruleKEY
	ruleSPACE
	ruleAction0
	ruleAction1
	ruleAction2
	rulePegText
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
	ruleAction34
	ruleAction35
	ruleAction36
	ruleAction37
	ruleAction38
	ruleAction39
	ruleAction40
	ruleAction41
	ruleAction42
	ruleAction43
	ruleAction44
	ruleAction45
	ruleAction46
	ruleAction47

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
	"describeMetrics",
	"describeSingleStmt",
	"propertyClause",
	"optionalPredicateClause",
	"expressionList",
	"expression_start",
	"expression_sum",
	"expression_product",
	"add_pipe",
	"expression_atom",
	"optionalGroupBy",
	"expression_function",
	"expression_metric",
	"groupByClause",
	"collapseByClause",
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
	"IDENTIFIER",
	"TIMESTAMP",
	"ID_SEGMENT",
	"ID_START",
	"ID_CONT",
	"PROPERTY_KEY",
	"PROPERTY_VALUE",
	"KEYWORD",
	"OP_PIPE",
	"OP_ADD",
	"OP_SUB",
	"OP_MULT",
	"OP_DIV",
	"OP_AND",
	"OP_OR",
	"OP_NOT",
	"QUOTE_SINGLE",
	"QUOTE_DOUBLE",
	"STRING",
	"CHAR",
	"ESCAPE_CLASS",
	"NUMBER",
	"NUMBER_NATURAL",
	"NUMBER_FRACTION",
	"NUMBER_INTEGER",
	"NUMBER_EXP",
	"DURATION",
	"PAREN_OPEN",
	"PAREN_CLOSE",
	"COMMA",
	"_",
	"KEY",
	"SPACE",
	"Action0",
	"Action1",
	"Action2",
	"PegText",
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
	"Action34",
	"Action35",
	"Action36",
	"Action37",
	"Action38",
	"Action39",
	"Action40",
	"Action41",
	"Action42",
	"Action43",
	"Action44",
	"Action45",
	"Action46",
	"Action47",

	"Pre_",
	"_In_",
	"_Suf",
}

type tokenTree interface {
	Print()
	PrintSyntax()
	PrintSyntaxTree(buffer string)
	Add(rule pegRule, begin, end, next uint32, depth int)
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
type token32 struct {
	pegRule
	begin, end, next uint32
}

func (t *token32) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: uint32(t.begin), end: uint32(t.end), next: uint32(t.next)}
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
		token.next = uint32(i)
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
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, uint32(depth), leaf
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

func (t *tokens32) Add(rule pegRule, begin, end, depth uint32, index int) {
	t.tree[index] = token32{pegRule: rule, begin: uint32(begin), end: uint32(end), next: uint32(depth)}
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

/*func (t *tokens16) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2 * len(tree))
		for i, v := range tree {
			expanded[i] = v.getToken32()
		}
		return &tokens32{tree: expanded}
	}
	return nil
}*/

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
	rules  [114]func() bool
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
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for token := range p.tokenTree.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

		case ruleAction0:

			p.makeSelect()

		case ruleAction1:
			p.makeDescribeAll()
		case ruleAction2:
			p.makeDescribeMetrics()
		case ruleAction3:
			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		case ruleAction4:
			p.makeDescribe()
		case ruleAction5:
			p.addEvaluationContext()
		case ruleAction6:
			p.addPropertyKey(buffer[begin:end])
		case ruleAction7:
			p.addPropertyValue(buffer[begin:end])
		case ruleAction8:
			p.insertPropertyKeyValue()
		case ruleAction9:
			p.checkPropertyClause()
		case ruleAction10:
			p.addNullPredicate()
		case ruleAction11:
			p.addExpressionList()
		case ruleAction12:
			p.appendExpression()
		case ruleAction13:
			p.appendExpression()
		case ruleAction14:
			p.addOperatorLiteral("+")
		case ruleAction15:
			p.addOperatorLiteral("-")
		case ruleAction16:
			p.addOperatorFunction()
		case ruleAction17:
			p.addOperatorLiteral("/")
		case ruleAction18:
			p.addOperatorLiteral("*")
		case ruleAction19:
			p.addOperatorFunction()
		case ruleAction20:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction21:
			p.addExpressionList()
		case ruleAction22:

			p.addExpressionList()
			p.addGroupBy()

		case ruleAction23:

			p.addPipeExpression()

		case ruleAction24:
			p.addDurationNode(text)
		case ruleAction25:
			p.addNumberNode(buffer[begin:end])
		case ruleAction26:
			p.addStringNode(unescapeLiteral(buffer[begin:end]))
		case ruleAction27:
			p.addGroupBy()
		case ruleAction28:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction29:

			p.addFunctionInvocation()

		case ruleAction30:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction31:
			p.addNullPredicate()
		case ruleAction32:

			p.addMetricExpression()

		case ruleAction33:

			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))

		case ruleAction34:

			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))

		case ruleAction35:

			p.appendCollapseBy(unescapeLiteral(text))

		case ruleAction36:
			p.appendCollapseBy(unescapeLiteral(text))
		case ruleAction37:
			p.addOrPredicate()
		case ruleAction38:
			p.addAndPredicate()
		case ruleAction39:
			p.addNotPredicate()
		case ruleAction40:

			p.addLiteralMatcher()

		case ruleAction41:

			p.addLiteralMatcher()
			p.addNotPredicate()

		case ruleAction42:

			p.addRegexMatcher()

		case ruleAction43:

			p.addListMatcher()

		case ruleAction44:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction45:
			p.addLiteralList()
		case ruleAction46:

			p.appendLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction47:
			p.addTagLiteral(unescapeLiteral(buffer[begin:end]))

		}
	}
	_, _, _, _ = buffer, text, begin, end
}

func (p *Parser) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != end_symbol {
		p.buffer = append(p.buffer, end_symbol)
	}

	var tree tokenTree = &tokens32{tree: make([]token32, math.MaxInt16)}
	position, depth, tokenIndex, buffer, _rules := uint32(0), uint32(0), 0, p.buffer, p.rules

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

	add := func(rule pegRule, begin uint32) {
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
		/* 0 root <- <((selectStmt / describeStmt) _ !.)> */
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
						if !_rules[rule_]() {
							goto l3
						}
						{
							position5, tokenIndex5, depth5 := position, tokenIndex, depth
							{
								position7, tokenIndex7, depth7 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l8
								}
								position++
								goto l7
							l8:
								position, tokenIndex, depth = position7, tokenIndex7, depth7
								if buffer[position] != rune('S') {
									goto l5
								}
								position++
							}
						l7:
							{
								position9, tokenIndex9, depth9 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l10
								}
								position++
								goto l9
							l10:
								position, tokenIndex, depth = position9, tokenIndex9, depth9
								if buffer[position] != rune('E') {
									goto l5
								}
								position++
							}
						l9:
							{
								position11, tokenIndex11, depth11 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l12
								}
								position++
								goto l11
							l12:
								position, tokenIndex, depth = position11, tokenIndex11, depth11
								if buffer[position] != rune('L') {
									goto l5
								}
								position++
							}
						l11:
							{
								position13, tokenIndex13, depth13 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l14
								}
								position++
								goto l13
							l14:
								position, tokenIndex, depth = position13, tokenIndex13, depth13
								if buffer[position] != rune('E') {
									goto l5
								}
								position++
							}
						l13:
							{
								position15, tokenIndex15, depth15 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l16
								}
								position++
								goto l15
							l16:
								position, tokenIndex, depth = position15, tokenIndex15, depth15
								if buffer[position] != rune('C') {
									goto l5
								}
								position++
							}
						l15:
							{
								position17, tokenIndex17, depth17 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l18
								}
								position++
								goto l17
							l18:
								position, tokenIndex, depth = position17, tokenIndex17, depth17
								if buffer[position] != rune('T') {
									goto l5
								}
								position++
							}
						l17:
							if !_rules[ruleKEY]() {
								goto l5
							}
							goto l6
						l5:
							position, tokenIndex, depth = position5, tokenIndex5, depth5
						}
					l6:
						if !_rules[ruleexpressionList]() {
							goto l3
						}
						if !_rules[ruleoptionalPredicateClause]() {
							goto l3
						}
						{
							position19 := position
							depth++
							{
								add(ruleAction5, position)
							}
						l21:
							{
								position22, tokenIndex22, depth22 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l22
								}
								if !_rules[rulePROPERTY_KEY]() {
									goto l22
								}
								{
									add(ruleAction6, position)
								}
								if !_rules[rule_]() {
									goto l22
								}
								{
									position24 := position
									depth++
									{
										position25 := position
										depth++
										{
											position26, tokenIndex26, depth26 := position, tokenIndex, depth
											if !_rules[rule_]() {
												goto l27
											}
											{
												position28 := position
												depth++
												if !_rules[ruleNUMBER]() {
													goto l27
												}
											l29:
												{
													position30, tokenIndex30, depth30 := position, tokenIndex, depth
													{
														position31, tokenIndex31, depth31 := position, tokenIndex, depth
														if c := buffer[position]; c < rune('a') || c > rune('z') {
															goto l32
														}
														position++
														goto l31
													l32:
														position, tokenIndex, depth = position31, tokenIndex31, depth31
														if c := buffer[position]; c < rune('A') || c > rune('Z') {
															goto l30
														}
														position++
													}
												l31:
													goto l29
												l30:
													position, tokenIndex, depth = position30, tokenIndex30, depth30
												}
												depth--
												add(rulePegText, position28)
											}
											goto l26
										l27:
											position, tokenIndex, depth = position26, tokenIndex26, depth26
											if !_rules[rule_]() {
												goto l33
											}
											if !_rules[ruleSTRING]() {
												goto l33
											}
											goto l26
										l33:
											position, tokenIndex, depth = position26, tokenIndex26, depth26
											if !_rules[rule_]() {
												goto l22
											}
											{
												position34 := position
												depth++
												{
													position35, tokenIndex35, depth35 := position, tokenIndex, depth
													if buffer[position] != rune('n') {
														goto l36
													}
													position++
													goto l35
												l36:
													position, tokenIndex, depth = position35, tokenIndex35, depth35
													if buffer[position] != rune('N') {
														goto l22
													}
													position++
												}
											l35:
												{
													position37, tokenIndex37, depth37 := position, tokenIndex, depth
													if buffer[position] != rune('o') {
														goto l38
													}
													position++
													goto l37
												l38:
													position, tokenIndex, depth = position37, tokenIndex37, depth37
													if buffer[position] != rune('O') {
														goto l22
													}
													position++
												}
											l37:
												{
													position39, tokenIndex39, depth39 := position, tokenIndex, depth
													if buffer[position] != rune('w') {
														goto l40
													}
													position++
													goto l39
												l40:
													position, tokenIndex, depth = position39, tokenIndex39, depth39
													if buffer[position] != rune('W') {
														goto l22
													}
													position++
												}
											l39:
												depth--
												add(rulePegText, position34)
											}
										}
									l26:
										depth--
										add(ruleTIMESTAMP, position25)
									}
									depth--
									add(rulePROPERTY_VALUE, position24)
								}
								{
									add(ruleAction7, position)
								}
								{
									add(ruleAction8, position)
								}
								goto l21
							l22:
								position, tokenIndex, depth = position22, tokenIndex22, depth22
							}
							{
								add(ruleAction9, position)
							}
							depth--
							add(rulepropertyClause, position19)
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
						position45 := position
						depth++
						if !_rules[rule_]() {
							goto l0
						}
						{
							position46, tokenIndex46, depth46 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l47
							}
							position++
							goto l46
						l47:
							position, tokenIndex, depth = position46, tokenIndex46, depth46
							if buffer[position] != rune('D') {
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
						{
							position50, tokenIndex50, depth50 := position, tokenIndex, depth
							if buffer[position] != rune('s') {
								goto l51
							}
							position++
							goto l50
						l51:
							position, tokenIndex, depth = position50, tokenIndex50, depth50
							if buffer[position] != rune('S') {
								goto l0
							}
							position++
						}
					l50:
						{
							position52, tokenIndex52, depth52 := position, tokenIndex, depth
							if buffer[position] != rune('c') {
								goto l53
							}
							position++
							goto l52
						l53:
							position, tokenIndex, depth = position52, tokenIndex52, depth52
							if buffer[position] != rune('C') {
								goto l0
							}
							position++
						}
					l52:
						{
							position54, tokenIndex54, depth54 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l55
							}
							position++
							goto l54
						l55:
							position, tokenIndex, depth = position54, tokenIndex54, depth54
							if buffer[position] != rune('R') {
								goto l0
							}
							position++
						}
					l54:
						{
							position56, tokenIndex56, depth56 := position, tokenIndex, depth
							if buffer[position] != rune('i') {
								goto l57
							}
							position++
							goto l56
						l57:
							position, tokenIndex, depth = position56, tokenIndex56, depth56
							if buffer[position] != rune('I') {
								goto l0
							}
							position++
						}
					l56:
						{
							position58, tokenIndex58, depth58 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l59
							}
							position++
							goto l58
						l59:
							position, tokenIndex, depth = position58, tokenIndex58, depth58
							if buffer[position] != rune('B') {
								goto l0
							}
							position++
						}
					l58:
						{
							position60, tokenIndex60, depth60 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l61
							}
							position++
							goto l60
						l61:
							position, tokenIndex, depth = position60, tokenIndex60, depth60
							if buffer[position] != rune('E') {
								goto l0
							}
							position++
						}
					l60:
						if !_rules[ruleKEY]() {
							goto l0
						}
						{
							position62, tokenIndex62, depth62 := position, tokenIndex, depth
							{
								position64 := position
								depth++
								if !_rules[rule_]() {
									goto l63
								}
								{
									position65, tokenIndex65, depth65 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l66
									}
									position++
									goto l65
								l66:
									position, tokenIndex, depth = position65, tokenIndex65, depth65
									if buffer[position] != rune('A') {
										goto l63
									}
									position++
								}
							l65:
								{
									position67, tokenIndex67, depth67 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l68
									}
									position++
									goto l67
								l68:
									position, tokenIndex, depth = position67, tokenIndex67, depth67
									if buffer[position] != rune('L') {
										goto l63
									}
									position++
								}
							l67:
								{
									position69, tokenIndex69, depth69 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l70
									}
									position++
									goto l69
								l70:
									position, tokenIndex, depth = position69, tokenIndex69, depth69
									if buffer[position] != rune('L') {
										goto l63
									}
									position++
								}
							l69:
								if !_rules[ruleKEY]() {
									goto l63
								}
								{
									add(ruleAction1, position)
								}
								depth--
								add(ruledescribeAllStmt, position64)
							}
							goto l62
						l63:
							position, tokenIndex, depth = position62, tokenIndex62, depth62
							{
								position73 := position
								depth++
								if !_rules[rule_]() {
									goto l72
								}
								{
									position74, tokenIndex74, depth74 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l75
									}
									position++
									goto l74
								l75:
									position, tokenIndex, depth = position74, tokenIndex74, depth74
									if buffer[position] != rune('M') {
										goto l72
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
										goto l72
									}
									position++
								}
							l76:
								{
									position78, tokenIndex78, depth78 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l79
									}
									position++
									goto l78
								l79:
									position, tokenIndex, depth = position78, tokenIndex78, depth78
									if buffer[position] != rune('T') {
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
									if buffer[position] != rune('i') {
										goto l83
									}
									position++
									goto l82
								l83:
									position, tokenIndex, depth = position82, tokenIndex82, depth82
									if buffer[position] != rune('I') {
										goto l72
									}
									position++
								}
							l82:
								{
									position84, tokenIndex84, depth84 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l85
									}
									position++
									goto l84
								l85:
									position, tokenIndex, depth = position84, tokenIndex84, depth84
									if buffer[position] != rune('C') {
										goto l72
									}
									position++
								}
							l84:
								{
									position86, tokenIndex86, depth86 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l87
									}
									position++
									goto l86
								l87:
									position, tokenIndex, depth = position86, tokenIndex86, depth86
									if buffer[position] != rune('S') {
										goto l72
									}
									position++
								}
							l86:
								if !_rules[ruleKEY]() {
									goto l72
								}
								if !_rules[rule_]() {
									goto l72
								}
								{
									position88, tokenIndex88, depth88 := position, tokenIndex, depth
									if buffer[position] != rune('w') {
										goto l89
									}
									position++
									goto l88
								l89:
									position, tokenIndex, depth = position88, tokenIndex88, depth88
									if buffer[position] != rune('W') {
										goto l72
									}
									position++
								}
							l88:
								{
									position90, tokenIndex90, depth90 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l91
									}
									position++
									goto l90
								l91:
									position, tokenIndex, depth = position90, tokenIndex90, depth90
									if buffer[position] != rune('H') {
										goto l72
									}
									position++
								}
							l90:
								{
									position92, tokenIndex92, depth92 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l93
									}
									position++
									goto l92
								l93:
									position, tokenIndex, depth = position92, tokenIndex92, depth92
									if buffer[position] != rune('E') {
										goto l72
									}
									position++
								}
							l92:
								{
									position94, tokenIndex94, depth94 := position, tokenIndex, depth
									if buffer[position] != rune('r') {
										goto l95
									}
									position++
									goto l94
								l95:
									position, tokenIndex, depth = position94, tokenIndex94, depth94
									if buffer[position] != rune('R') {
										goto l72
									}
									position++
								}
							l94:
								{
									position96, tokenIndex96, depth96 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l97
									}
									position++
									goto l96
								l97:
									position, tokenIndex, depth = position96, tokenIndex96, depth96
									if buffer[position] != rune('E') {
										goto l72
									}
									position++
								}
							l96:
								if !_rules[ruleKEY]() {
									goto l72
								}
								if !_rules[ruletagName]() {
									goto l72
								}
								if !_rules[rule_]() {
									goto l72
								}
								if buffer[position] != rune('=') {
									goto l72
								}
								position++
								if !_rules[ruleliteralString]() {
									goto l72
								}
								{
									add(ruleAction2, position)
								}
								depth--
								add(ruledescribeMetrics, position73)
							}
							goto l62
						l72:
							position, tokenIndex, depth = position62, tokenIndex62, depth62
							{
								position99 := position
								depth++
								if !_rules[rule_]() {
									goto l0
								}
								{
									position100 := position
									depth++
									{
										position101 := position
										depth++
										if !_rules[ruleIDENTIFIER]() {
											goto l0
										}
										depth--
										add(ruleMETRIC_NAME, position101)
									}
									depth--
									add(rulePegText, position100)
								}
								{
									add(ruleAction3, position)
								}
								if !_rules[ruleoptionalPredicateClause]() {
									goto l0
								}
								{
									add(ruleAction4, position)
								}
								depth--
								add(ruledescribeSingleStmt, position99)
							}
						}
					l62:
						depth--
						add(ruledescribeStmt, position45)
					}
				}
			l2:
				if !_rules[rule_]() {
					goto l0
				}
				{
					position104, tokenIndex104, depth104 := position, tokenIndex, depth
					if !matchDot() {
						goto l104
					}
					goto l0
				l104:
					position, tokenIndex, depth = position104, tokenIndex104, depth104
				}
				depth--
				add(ruleroot, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 selectStmt <- <(_ (('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T') KEY)? expressionList optionalPredicateClause propertyClause Action0)> */
		nil,
		/* 2 describeStmt <- <(_ (('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E')) KEY (describeAllStmt / describeMetrics / describeSingleStmt))> */
		nil,
		/* 3 describeAllStmt <- <(_ (('a' / 'A') ('l' / 'L') ('l' / 'L')) KEY Action1)> */
		nil,
		/* 4 describeMetrics <- <(_ (('m' / 'M') ('e' / 'E') ('t' / 'T') ('r' / 'R') ('i' / 'I') ('c' / 'C') ('s' / 'S')) KEY _ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY tagName _ '=' literalString Action2)> */
		nil,
		/* 5 describeSingleStmt <- <(_ <METRIC_NAME> Action3 optionalPredicateClause Action4)> */
		nil,
		/* 6 propertyClause <- <(Action5 (_ PROPERTY_KEY Action6 _ PROPERTY_VALUE Action7 Action8)* Action9)> */
		nil,
		/* 7 optionalPredicateClause <- <(predicateClause / Action10)> */
		func() bool {
			{
				position112 := position
				depth++
				{
					position113, tokenIndex113, depth113 := position, tokenIndex, depth
					{
						position115 := position
						depth++
						if !_rules[rule_]() {
							goto l114
						}
						{
							position116, tokenIndex116, depth116 := position, tokenIndex, depth
							if buffer[position] != rune('w') {
								goto l117
							}
							position++
							goto l116
						l117:
							position, tokenIndex, depth = position116, tokenIndex116, depth116
							if buffer[position] != rune('W') {
								goto l114
							}
							position++
						}
					l116:
						{
							position118, tokenIndex118, depth118 := position, tokenIndex, depth
							if buffer[position] != rune('h') {
								goto l119
							}
							position++
							goto l118
						l119:
							position, tokenIndex, depth = position118, tokenIndex118, depth118
							if buffer[position] != rune('H') {
								goto l114
							}
							position++
						}
					l118:
						{
							position120, tokenIndex120, depth120 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l121
							}
							position++
							goto l120
						l121:
							position, tokenIndex, depth = position120, tokenIndex120, depth120
							if buffer[position] != rune('E') {
								goto l114
							}
							position++
						}
					l120:
						{
							position122, tokenIndex122, depth122 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l123
							}
							position++
							goto l122
						l123:
							position, tokenIndex, depth = position122, tokenIndex122, depth122
							if buffer[position] != rune('R') {
								goto l114
							}
							position++
						}
					l122:
						{
							position124, tokenIndex124, depth124 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l125
							}
							position++
							goto l124
						l125:
							position, tokenIndex, depth = position124, tokenIndex124, depth124
							if buffer[position] != rune('E') {
								goto l114
							}
							position++
						}
					l124:
						if !_rules[ruleKEY]() {
							goto l114
						}
						if !_rules[rule_]() {
							goto l114
						}
						if !_rules[rulepredicate_1]() {
							goto l114
						}
						depth--
						add(rulepredicateClause, position115)
					}
					goto l113
				l114:
					position, tokenIndex, depth = position113, tokenIndex113, depth113
					{
						add(ruleAction10, position)
					}
				}
			l113:
				depth--
				add(ruleoptionalPredicateClause, position112)
			}
			return true
		},
		/* 8 expressionList <- <(Action11 expression_start Action12 (_ COMMA expression_start Action13)*)> */
		func() bool {
			position127, tokenIndex127, depth127 := position, tokenIndex, depth
			{
				position128 := position
				depth++
				{
					add(ruleAction11, position)
				}
				if !_rules[ruleexpression_start]() {
					goto l127
				}
				{
					add(ruleAction12, position)
				}
			l131:
				{
					position132, tokenIndex132, depth132 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l132
					}
					if !_rules[ruleCOMMA]() {
						goto l132
					}
					if !_rules[ruleexpression_start]() {
						goto l132
					}
					{
						add(ruleAction13, position)
					}
					goto l131
				l132:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
				}
				depth--
				add(ruleexpressionList, position128)
			}
			return true
		l127:
			position, tokenIndex, depth = position127, tokenIndex127, depth127
			return false
		},
		/* 9 expression_start <- <(expression_sum add_pipe)> */
		func() bool {
			position134, tokenIndex134, depth134 := position, tokenIndex, depth
			{
				position135 := position
				depth++
				{
					position136 := position
					depth++
					if !_rules[ruleexpression_product]() {
						goto l134
					}
				l137:
					{
						position138, tokenIndex138, depth138 := position, tokenIndex, depth
						if !_rules[ruleadd_pipe]() {
							goto l138
						}
						{
							position139, tokenIndex139, depth139 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l140
							}
							{
								position141 := position
								depth++
								if buffer[position] != rune('+') {
									goto l140
								}
								position++
								depth--
								add(ruleOP_ADD, position141)
							}
							{
								add(ruleAction14, position)
							}
							goto l139
						l140:
							position, tokenIndex, depth = position139, tokenIndex139, depth139
							if !_rules[rule_]() {
								goto l138
							}
							{
								position143 := position
								depth++
								if buffer[position] != rune('-') {
									goto l138
								}
								position++
								depth--
								add(ruleOP_SUB, position143)
							}
							{
								add(ruleAction15, position)
							}
						}
					l139:
						if !_rules[ruleexpression_product]() {
							goto l138
						}
						{
							add(ruleAction16, position)
						}
						goto l137
					l138:
						position, tokenIndex, depth = position138, tokenIndex138, depth138
					}
					depth--
					add(ruleexpression_sum, position136)
				}
				if !_rules[ruleadd_pipe]() {
					goto l134
				}
				depth--
				add(ruleexpression_start, position135)
			}
			return true
		l134:
			position, tokenIndex, depth = position134, tokenIndex134, depth134
			return false
		},
		/* 10 expression_sum <- <(expression_product (add_pipe ((_ OP_ADD Action14) / (_ OP_SUB Action15)) expression_product Action16)*)> */
		nil,
		/* 11 expression_product <- <(expression_atom (add_pipe ((_ OP_DIV Action17) / (_ OP_MULT Action18)) expression_atom Action19)*)> */
		func() bool {
			position147, tokenIndex147, depth147 := position, tokenIndex, depth
			{
				position148 := position
				depth++
				if !_rules[ruleexpression_atom]() {
					goto l147
				}
			l149:
				{
					position150, tokenIndex150, depth150 := position, tokenIndex, depth
					if !_rules[ruleadd_pipe]() {
						goto l150
					}
					{
						position151, tokenIndex151, depth151 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l152
						}
						{
							position153 := position
							depth++
							if buffer[position] != rune('/') {
								goto l152
							}
							position++
							depth--
							add(ruleOP_DIV, position153)
						}
						{
							add(ruleAction17, position)
						}
						goto l151
					l152:
						position, tokenIndex, depth = position151, tokenIndex151, depth151
						if !_rules[rule_]() {
							goto l150
						}
						{
							position155 := position
							depth++
							if buffer[position] != rune('*') {
								goto l150
							}
							position++
							depth--
							add(ruleOP_MULT, position155)
						}
						{
							add(ruleAction18, position)
						}
					}
				l151:
					if !_rules[ruleexpression_atom]() {
						goto l150
					}
					{
						add(ruleAction19, position)
					}
					goto l149
				l150:
					position, tokenIndex, depth = position150, tokenIndex150, depth150
				}
				depth--
				add(ruleexpression_product, position148)
			}
			return true
		l147:
			position, tokenIndex, depth = position147, tokenIndex147, depth147
			return false
		},
		/* 12 add_pipe <- <(_ OP_PIPE _ <IDENTIFIER> Action20 ((_ PAREN_OPEN (expressionList / Action21) optionalGroupBy _ PAREN_CLOSE) / Action22) Action23)*> */
		func() bool {
			{
				position159 := position
				depth++
			l160:
				{
					position161, tokenIndex161, depth161 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l161
					}
					{
						position162 := position
						depth++
						if buffer[position] != rune('|') {
							goto l161
						}
						position++
						depth--
						add(ruleOP_PIPE, position162)
					}
					if !_rules[rule_]() {
						goto l161
					}
					{
						position163 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l161
						}
						depth--
						add(rulePegText, position163)
					}
					{
						add(ruleAction20, position)
					}
					{
						position165, tokenIndex165, depth165 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l166
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l166
						}
						{
							position167, tokenIndex167, depth167 := position, tokenIndex, depth
							if !_rules[ruleexpressionList]() {
								goto l168
							}
							goto l167
						l168:
							position, tokenIndex, depth = position167, tokenIndex167, depth167
							{
								add(ruleAction21, position)
							}
						}
					l167:
						if !_rules[ruleoptionalGroupBy]() {
							goto l166
						}
						if !_rules[rule_]() {
							goto l166
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l166
						}
						goto l165
					l166:
						position, tokenIndex, depth = position165, tokenIndex165, depth165
						{
							add(ruleAction22, position)
						}
					}
				l165:
					{
						add(ruleAction23, position)
					}
					goto l160
				l161:
					position, tokenIndex, depth = position161, tokenIndex161, depth161
				}
				depth--
				add(ruleadd_pipe, position159)
			}
			return true
		},
		/* 13 expression_atom <- <(expression_function / expression_metric / (_ PAREN_OPEN expression_start _ PAREN_CLOSE) / (_ <DURATION> Action24) / (_ <NUMBER> Action25) / (_ STRING Action26))> */
		func() bool {
			position172, tokenIndex172, depth172 := position, tokenIndex, depth
			{
				position173 := position
				depth++
				{
					position174, tokenIndex174, depth174 := position, tokenIndex, depth
					{
						position176 := position
						depth++
						if !_rules[rule_]() {
							goto l175
						}
						{
							position177 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l175
							}
							depth--
							add(rulePegText, position177)
						}
						{
							add(ruleAction28, position)
						}
						if !_rules[rule_]() {
							goto l175
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l175
						}
						if !_rules[ruleexpressionList]() {
							goto l175
						}
						if !_rules[ruleoptionalGroupBy]() {
							goto l175
						}
						if !_rules[rule_]() {
							goto l175
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l175
						}
						{
							add(ruleAction29, position)
						}
						depth--
						add(ruleexpression_function, position176)
					}
					goto l174
				l175:
					position, tokenIndex, depth = position174, tokenIndex174, depth174
					{
						position181 := position
						depth++
						if !_rules[rule_]() {
							goto l180
						}
						{
							position182 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l180
							}
							depth--
							add(rulePegText, position182)
						}
						{
							add(ruleAction30, position)
						}
						{
							position184, tokenIndex184, depth184 := position, tokenIndex, depth
							{
								position186, tokenIndex186, depth186 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l187
								}
								if buffer[position] != rune('[') {
									goto l187
								}
								position++
								if !_rules[rulepredicate_1]() {
									goto l187
								}
								if !_rules[rule_]() {
									goto l187
								}
								if buffer[position] != rune(']') {
									goto l187
								}
								position++
								goto l186
							l187:
								position, tokenIndex, depth = position186, tokenIndex186, depth186
								{
									add(ruleAction31, position)
								}
							}
						l186:
							goto l185

							position, tokenIndex, depth = position184, tokenIndex184, depth184
						}
					l185:
						{
							add(ruleAction32, position)
						}
						depth--
						add(ruleexpression_metric, position181)
					}
					goto l174
				l180:
					position, tokenIndex, depth = position174, tokenIndex174, depth174
					if !_rules[rule_]() {
						goto l190
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l190
					}
					if !_rules[ruleexpression_start]() {
						goto l190
					}
					if !_rules[rule_]() {
						goto l190
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l190
					}
					goto l174
				l190:
					position, tokenIndex, depth = position174, tokenIndex174, depth174
					if !_rules[rule_]() {
						goto l191
					}
					{
						position192 := position
						depth++
						{
							position193 := position
							depth++
							if !_rules[ruleNUMBER]() {
								goto l191
							}
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l191
							}
							position++
						l194:
							{
								position195, tokenIndex195, depth195 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l195
								}
								position++
								goto l194
							l195:
								position, tokenIndex, depth = position195, tokenIndex195, depth195
							}
							if !_rules[ruleKEY]() {
								goto l191
							}
							depth--
							add(ruleDURATION, position193)
						}
						depth--
						add(rulePegText, position192)
					}
					{
						add(ruleAction24, position)
					}
					goto l174
				l191:
					position, tokenIndex, depth = position174, tokenIndex174, depth174
					if !_rules[rule_]() {
						goto l197
					}
					{
						position198 := position
						depth++
						if !_rules[ruleNUMBER]() {
							goto l197
						}
						depth--
						add(rulePegText, position198)
					}
					{
						add(ruleAction25, position)
					}
					goto l174
				l197:
					position, tokenIndex, depth = position174, tokenIndex174, depth174
					if !_rules[rule_]() {
						goto l172
					}
					if !_rules[ruleSTRING]() {
						goto l172
					}
					{
						add(ruleAction26, position)
					}
				}
			l174:
				depth--
				add(ruleexpression_atom, position173)
			}
			return true
		l172:
			position, tokenIndex, depth = position172, tokenIndex172, depth172
			return false
		},
		/* 14 optionalGroupBy <- <(Action27 (groupByClause / collapseByClause)?)> */
		func() bool {
			{
				position202 := position
				depth++
				{
					add(ruleAction27, position)
				}
				{
					position204, tokenIndex204, depth204 := position, tokenIndex, depth
					{
						position206, tokenIndex206, depth206 := position, tokenIndex, depth
						{
							position208 := position
							depth++
							if !_rules[rule_]() {
								goto l207
							}
							{
								position209, tokenIndex209, depth209 := position, tokenIndex, depth
								if buffer[position] != rune('g') {
									goto l210
								}
								position++
								goto l209
							l210:
								position, tokenIndex, depth = position209, tokenIndex209, depth209
								if buffer[position] != rune('G') {
									goto l207
								}
								position++
							}
						l209:
							{
								position211, tokenIndex211, depth211 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l212
								}
								position++
								goto l211
							l212:
								position, tokenIndex, depth = position211, tokenIndex211, depth211
								if buffer[position] != rune('R') {
									goto l207
								}
								position++
							}
						l211:
							{
								position213, tokenIndex213, depth213 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l214
								}
								position++
								goto l213
							l214:
								position, tokenIndex, depth = position213, tokenIndex213, depth213
								if buffer[position] != rune('O') {
									goto l207
								}
								position++
							}
						l213:
							{
								position215, tokenIndex215, depth215 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l216
								}
								position++
								goto l215
							l216:
								position, tokenIndex, depth = position215, tokenIndex215, depth215
								if buffer[position] != rune('U') {
									goto l207
								}
								position++
							}
						l215:
							{
								position217, tokenIndex217, depth217 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l218
								}
								position++
								goto l217
							l218:
								position, tokenIndex, depth = position217, tokenIndex217, depth217
								if buffer[position] != rune('P') {
									goto l207
								}
								position++
							}
						l217:
							if !_rules[ruleKEY]() {
								goto l207
							}
							if !_rules[rule_]() {
								goto l207
							}
							{
								position219, tokenIndex219, depth219 := position, tokenIndex, depth
								if buffer[position] != rune('b') {
									goto l220
								}
								position++
								goto l219
							l220:
								position, tokenIndex, depth = position219, tokenIndex219, depth219
								if buffer[position] != rune('B') {
									goto l207
								}
								position++
							}
						l219:
							{
								position221, tokenIndex221, depth221 := position, tokenIndex, depth
								if buffer[position] != rune('y') {
									goto l222
								}
								position++
								goto l221
							l222:
								position, tokenIndex, depth = position221, tokenIndex221, depth221
								if buffer[position] != rune('Y') {
									goto l207
								}
								position++
							}
						l221:
							if !_rules[ruleKEY]() {
								goto l207
							}
							if !_rules[rule_]() {
								goto l207
							}
							{
								position223 := position
								depth++
								if !_rules[ruleCOLUMN_NAME]() {
									goto l207
								}
								depth--
								add(rulePegText, position223)
							}
							{
								add(ruleAction33, position)
							}
						l225:
							{
								position226, tokenIndex226, depth226 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l226
								}
								if !_rules[ruleCOMMA]() {
									goto l226
								}
								if !_rules[rule_]() {
									goto l226
								}
								{
									position227 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l226
									}
									depth--
									add(rulePegText, position227)
								}
								{
									add(ruleAction34, position)
								}
								goto l225
							l226:
								position, tokenIndex, depth = position226, tokenIndex226, depth226
							}
							depth--
							add(rulegroupByClause, position208)
						}
						goto l206
					l207:
						position, tokenIndex, depth = position206, tokenIndex206, depth206
						{
							position229 := position
							depth++
							if !_rules[rule_]() {
								goto l204
							}
							{
								position230, tokenIndex230, depth230 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l231
								}
								position++
								goto l230
							l231:
								position, tokenIndex, depth = position230, tokenIndex230, depth230
								if buffer[position] != rune('C') {
									goto l204
								}
								position++
							}
						l230:
							{
								position232, tokenIndex232, depth232 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l233
								}
								position++
								goto l232
							l233:
								position, tokenIndex, depth = position232, tokenIndex232, depth232
								if buffer[position] != rune('O') {
									goto l204
								}
								position++
							}
						l232:
							{
								position234, tokenIndex234, depth234 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l235
								}
								position++
								goto l234
							l235:
								position, tokenIndex, depth = position234, tokenIndex234, depth234
								if buffer[position] != rune('L') {
									goto l204
								}
								position++
							}
						l234:
							{
								position236, tokenIndex236, depth236 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l237
								}
								position++
								goto l236
							l237:
								position, tokenIndex, depth = position236, tokenIndex236, depth236
								if buffer[position] != rune('L') {
									goto l204
								}
								position++
							}
						l236:
							{
								position238, tokenIndex238, depth238 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l239
								}
								position++
								goto l238
							l239:
								position, tokenIndex, depth = position238, tokenIndex238, depth238
								if buffer[position] != rune('A') {
									goto l204
								}
								position++
							}
						l238:
							{
								position240, tokenIndex240, depth240 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l241
								}
								position++
								goto l240
							l241:
								position, tokenIndex, depth = position240, tokenIndex240, depth240
								if buffer[position] != rune('P') {
									goto l204
								}
								position++
							}
						l240:
							{
								position242, tokenIndex242, depth242 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l243
								}
								position++
								goto l242
							l243:
								position, tokenIndex, depth = position242, tokenIndex242, depth242
								if buffer[position] != rune('S') {
									goto l204
								}
								position++
							}
						l242:
							{
								position244, tokenIndex244, depth244 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l245
								}
								position++
								goto l244
							l245:
								position, tokenIndex, depth = position244, tokenIndex244, depth244
								if buffer[position] != rune('E') {
									goto l204
								}
								position++
							}
						l244:
							if !_rules[ruleKEY]() {
								goto l204
							}
							if !_rules[rule_]() {
								goto l204
							}
							{
								position246, tokenIndex246, depth246 := position, tokenIndex, depth
								if buffer[position] != rune('b') {
									goto l247
								}
								position++
								goto l246
							l247:
								position, tokenIndex, depth = position246, tokenIndex246, depth246
								if buffer[position] != rune('B') {
									goto l204
								}
								position++
							}
						l246:
							{
								position248, tokenIndex248, depth248 := position, tokenIndex, depth
								if buffer[position] != rune('y') {
									goto l249
								}
								position++
								goto l248
							l249:
								position, tokenIndex, depth = position248, tokenIndex248, depth248
								if buffer[position] != rune('Y') {
									goto l204
								}
								position++
							}
						l248:
							if !_rules[ruleKEY]() {
								goto l204
							}
							if !_rules[rule_]() {
								goto l204
							}
							{
								position250 := position
								depth++
								if !_rules[ruleCOLUMN_NAME]() {
									goto l204
								}
								depth--
								add(rulePegText, position250)
							}
							{
								add(ruleAction35, position)
							}
						l252:
							{
								position253, tokenIndex253, depth253 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l253
								}
								if !_rules[ruleCOMMA]() {
									goto l253
								}
								if !_rules[rule_]() {
									goto l253
								}
								{
									position254 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l253
									}
									depth--
									add(rulePegText, position254)
								}
								{
									add(ruleAction36, position)
								}
								goto l252
							l253:
								position, tokenIndex, depth = position253, tokenIndex253, depth253
							}
							depth--
							add(rulecollapseByClause, position229)
						}
					}
				l206:
					goto l205
				l204:
					position, tokenIndex, depth = position204, tokenIndex204, depth204
				}
			l205:
				depth--
				add(ruleoptionalGroupBy, position202)
			}
			return true
		},
		/* 15 expression_function <- <(_ <IDENTIFIER> Action28 _ PAREN_OPEN expressionList optionalGroupBy _ PAREN_CLOSE Action29)> */
		nil,
		/* 16 expression_metric <- <(_ <IDENTIFIER> Action30 ((_ '[' predicate_1 _ ']') / Action31)? Action32)> */
		nil,
		/* 17 groupByClause <- <(_ (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P')) KEY _ (('b' / 'B') ('y' / 'Y')) KEY _ <COLUMN_NAME> Action33 (_ COMMA _ <COLUMN_NAME> Action34)*)> */
		nil,
		/* 18 collapseByClause <- <(_ (('c' / 'C') ('o' / 'O') ('l' / 'L') ('l' / 'L') ('a' / 'A') ('p' / 'P') ('s' / 'S') ('e' / 'E')) KEY _ (('b' / 'B') ('y' / 'Y')) KEY _ <COLUMN_NAME> Action35 (_ COMMA _ <COLUMN_NAME> Action36)*)> */
		nil,
		/* 19 predicateClause <- <(_ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY _ predicate_1)> */
		nil,
		/* 20 predicate_1 <- <((predicate_2 _ OP_OR predicate_1 Action37) / predicate_2)> */
		func() bool {
			position261, tokenIndex261, depth261 := position, tokenIndex, depth
			{
				position262 := position
				depth++
				{
					position263, tokenIndex263, depth263 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l264
					}
					if !_rules[rule_]() {
						goto l264
					}
					{
						position265 := position
						depth++
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
								goto l264
							}
							position++
						}
					l266:
						{
							position268, tokenIndex268, depth268 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l269
							}
							position++
							goto l268
						l269:
							position, tokenIndex, depth = position268, tokenIndex268, depth268
							if buffer[position] != rune('R') {
								goto l264
							}
							position++
						}
					l268:
						if !_rules[ruleKEY]() {
							goto l264
						}
						depth--
						add(ruleOP_OR, position265)
					}
					if !_rules[rulepredicate_1]() {
						goto l264
					}
					{
						add(ruleAction37, position)
					}
					goto l263
				l264:
					position, tokenIndex, depth = position263, tokenIndex263, depth263
					if !_rules[rulepredicate_2]() {
						goto l261
					}
				}
			l263:
				depth--
				add(rulepredicate_1, position262)
			}
			return true
		l261:
			position, tokenIndex, depth = position261, tokenIndex261, depth261
			return false
		},
		/* 21 predicate_2 <- <((predicate_3 _ OP_AND predicate_2 Action38) / predicate_3)> */
		func() bool {
			position271, tokenIndex271, depth271 := position, tokenIndex, depth
			{
				position272 := position
				depth++
				{
					position273, tokenIndex273, depth273 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l274
					}
					if !_rules[rule_]() {
						goto l274
					}
					{
						position275 := position
						depth++
						{
							position276, tokenIndex276, depth276 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l277
							}
							position++
							goto l276
						l277:
							position, tokenIndex, depth = position276, tokenIndex276, depth276
							if buffer[position] != rune('A') {
								goto l274
							}
							position++
						}
					l276:
						{
							position278, tokenIndex278, depth278 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l279
							}
							position++
							goto l278
						l279:
							position, tokenIndex, depth = position278, tokenIndex278, depth278
							if buffer[position] != rune('N') {
								goto l274
							}
							position++
						}
					l278:
						{
							position280, tokenIndex280, depth280 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l281
							}
							position++
							goto l280
						l281:
							position, tokenIndex, depth = position280, tokenIndex280, depth280
							if buffer[position] != rune('D') {
								goto l274
							}
							position++
						}
					l280:
						if !_rules[ruleKEY]() {
							goto l274
						}
						depth--
						add(ruleOP_AND, position275)
					}
					if !_rules[rulepredicate_2]() {
						goto l274
					}
					{
						add(ruleAction38, position)
					}
					goto l273
				l274:
					position, tokenIndex, depth = position273, tokenIndex273, depth273
					if !_rules[rulepredicate_3]() {
						goto l271
					}
				}
			l273:
				depth--
				add(rulepredicate_2, position272)
			}
			return true
		l271:
			position, tokenIndex, depth = position271, tokenIndex271, depth271
			return false
		},
		/* 22 predicate_3 <- <((_ OP_NOT predicate_3 Action39) / (_ PAREN_OPEN predicate_1 _ PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position283, tokenIndex283, depth283 := position, tokenIndex, depth
			{
				position284 := position
				depth++
				{
					position285, tokenIndex285, depth285 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l286
					}
					{
						position287 := position
						depth++
						{
							position288, tokenIndex288, depth288 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l289
							}
							position++
							goto l288
						l289:
							position, tokenIndex, depth = position288, tokenIndex288, depth288
							if buffer[position] != rune('N') {
								goto l286
							}
							position++
						}
					l288:
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
								goto l286
							}
							position++
						}
					l290:
						{
							position292, tokenIndex292, depth292 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l293
							}
							position++
							goto l292
						l293:
							position, tokenIndex, depth = position292, tokenIndex292, depth292
							if buffer[position] != rune('T') {
								goto l286
							}
							position++
						}
					l292:
						if !_rules[ruleKEY]() {
							goto l286
						}
						depth--
						add(ruleOP_NOT, position287)
					}
					if !_rules[rulepredicate_3]() {
						goto l286
					}
					{
						add(ruleAction39, position)
					}
					goto l285
				l286:
					position, tokenIndex, depth = position285, tokenIndex285, depth285
					if !_rules[rule_]() {
						goto l295
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l295
					}
					if !_rules[rulepredicate_1]() {
						goto l295
					}
					if !_rules[rule_]() {
						goto l295
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l295
					}
					goto l285
				l295:
					position, tokenIndex, depth = position285, tokenIndex285, depth285
					{
						position296 := position
						depth++
						{
							position297, tokenIndex297, depth297 := position, tokenIndex, depth
							if !_rules[ruletagName]() {
								goto l298
							}
							if !_rules[rule_]() {
								goto l298
							}
							if buffer[position] != rune('=') {
								goto l298
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l298
							}
							{
								add(ruleAction40, position)
							}
							goto l297
						l298:
							position, tokenIndex, depth = position297, tokenIndex297, depth297
							if !_rules[ruletagName]() {
								goto l300
							}
							if !_rules[rule_]() {
								goto l300
							}
							if buffer[position] != rune('!') {
								goto l300
							}
							position++
							if buffer[position] != rune('=') {
								goto l300
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l300
							}
							{
								add(ruleAction41, position)
							}
							goto l297
						l300:
							position, tokenIndex, depth = position297, tokenIndex297, depth297
							if !_rules[ruletagName]() {
								goto l302
							}
							if !_rules[rule_]() {
								goto l302
							}
							{
								position303, tokenIndex303, depth303 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l304
								}
								position++
								goto l303
							l304:
								position, tokenIndex, depth = position303, tokenIndex303, depth303
								if buffer[position] != rune('M') {
									goto l302
								}
								position++
							}
						l303:
							{
								position305, tokenIndex305, depth305 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l306
								}
								position++
								goto l305
							l306:
								position, tokenIndex, depth = position305, tokenIndex305, depth305
								if buffer[position] != rune('A') {
									goto l302
								}
								position++
							}
						l305:
							{
								position307, tokenIndex307, depth307 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l308
								}
								position++
								goto l307
							l308:
								position, tokenIndex, depth = position307, tokenIndex307, depth307
								if buffer[position] != rune('T') {
									goto l302
								}
								position++
							}
						l307:
							{
								position309, tokenIndex309, depth309 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l310
								}
								position++
								goto l309
							l310:
								position, tokenIndex, depth = position309, tokenIndex309, depth309
								if buffer[position] != rune('C') {
									goto l302
								}
								position++
							}
						l309:
							{
								position311, tokenIndex311, depth311 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l312
								}
								position++
								goto l311
							l312:
								position, tokenIndex, depth = position311, tokenIndex311, depth311
								if buffer[position] != rune('H') {
									goto l302
								}
								position++
							}
						l311:
							{
								position313, tokenIndex313, depth313 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l314
								}
								position++
								goto l313
							l314:
								position, tokenIndex, depth = position313, tokenIndex313, depth313
								if buffer[position] != rune('E') {
									goto l302
								}
								position++
							}
						l313:
							{
								position315, tokenIndex315, depth315 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l316
								}
								position++
								goto l315
							l316:
								position, tokenIndex, depth = position315, tokenIndex315, depth315
								if buffer[position] != rune('S') {
									goto l302
								}
								position++
							}
						l315:
							if !_rules[ruleKEY]() {
								goto l302
							}
							if !_rules[ruleliteralString]() {
								goto l302
							}
							{
								add(ruleAction42, position)
							}
							goto l297
						l302:
							position, tokenIndex, depth = position297, tokenIndex297, depth297
							if !_rules[ruletagName]() {
								goto l283
							}
							if !_rules[rule_]() {
								goto l283
							}
							{
								position318, tokenIndex318, depth318 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l319
								}
								position++
								goto l318
							l319:
								position, tokenIndex, depth = position318, tokenIndex318, depth318
								if buffer[position] != rune('I') {
									goto l283
								}
								position++
							}
						l318:
							{
								position320, tokenIndex320, depth320 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l321
								}
								position++
								goto l320
							l321:
								position, tokenIndex, depth = position320, tokenIndex320, depth320
								if buffer[position] != rune('N') {
									goto l283
								}
								position++
							}
						l320:
							if !_rules[ruleKEY]() {
								goto l283
							}
							{
								position322 := position
								depth++
								{
									add(ruleAction45, position)
								}
								if !_rules[rule_]() {
									goto l283
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l283
								}
								if !_rules[ruleliteralListString]() {
									goto l283
								}
							l324:
								{
									position325, tokenIndex325, depth325 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l325
									}
									if !_rules[ruleCOMMA]() {
										goto l325
									}
									if !_rules[ruleliteralListString]() {
										goto l325
									}
									goto l324
								l325:
									position, tokenIndex, depth = position325, tokenIndex325, depth325
								}
								if !_rules[rule_]() {
									goto l283
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l283
								}
								depth--
								add(ruleliteralList, position322)
							}
							{
								add(ruleAction43, position)
							}
						}
					l297:
						depth--
						add(ruletagMatcher, position296)
					}
				}
			l285:
				depth--
				add(rulepredicate_3, position284)
			}
			return true
		l283:
			position, tokenIndex, depth = position283, tokenIndex283, depth283
			return false
		},
		/* 23 tagMatcher <- <((tagName _ '=' literalString Action40) / (tagName _ ('!' '=') literalString Action41) / (tagName _ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) KEY literalString Action42) / (tagName _ (('i' / 'I') ('n' / 'N')) KEY literalList Action43))> */
		nil,
		/* 24 literalString <- <(_ STRING Action44)> */
		func() bool {
			position328, tokenIndex328, depth328 := position, tokenIndex, depth
			{
				position329 := position
				depth++
				if !_rules[rule_]() {
					goto l328
				}
				if !_rules[ruleSTRING]() {
					goto l328
				}
				{
					add(ruleAction44, position)
				}
				depth--
				add(ruleliteralString, position329)
			}
			return true
		l328:
			position, tokenIndex, depth = position328, tokenIndex328, depth328
			return false
		},
		/* 25 literalList <- <(Action45 _ PAREN_OPEN literalListString (_ COMMA literalListString)* _ PAREN_CLOSE)> */
		nil,
		/* 26 literalListString <- <(_ STRING Action46)> */
		func() bool {
			position332, tokenIndex332, depth332 := position, tokenIndex, depth
			{
				position333 := position
				depth++
				if !_rules[rule_]() {
					goto l332
				}
				if !_rules[ruleSTRING]() {
					goto l332
				}
				{
					add(ruleAction46, position)
				}
				depth--
				add(ruleliteralListString, position333)
			}
			return true
		l332:
			position, tokenIndex, depth = position332, tokenIndex332, depth332
			return false
		},
		/* 27 tagName <- <(_ <TAG_NAME> Action47)> */
		func() bool {
			position335, tokenIndex335, depth335 := position, tokenIndex, depth
			{
				position336 := position
				depth++
				if !_rules[rule_]() {
					goto l335
				}
				{
					position337 := position
					depth++
					{
						position338 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l335
						}
						depth--
						add(ruleTAG_NAME, position338)
					}
					depth--
					add(rulePegText, position337)
				}
				{
					add(ruleAction47, position)
				}
				depth--
				add(ruletagName, position336)
			}
			return true
		l335:
			position, tokenIndex, depth = position335, tokenIndex335, depth335
			return false
		},
		/* 28 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position340, tokenIndex340, depth340 := position, tokenIndex, depth
			{
				position341 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l340
				}
				depth--
				add(ruleCOLUMN_NAME, position341)
			}
			return true
		l340:
			position, tokenIndex, depth = position340, tokenIndex340, depth340
			return false
		},
		/* 29 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 30 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 31 IDENTIFIER <- <(('`' CHAR* '`') / (_ !(KEYWORD KEY) ID_SEGMENT ('.' ID_SEGMENT)*))> */
		func() bool {
			position344, tokenIndex344, depth344 := position, tokenIndex, depth
			{
				position345 := position
				depth++
				{
					position346, tokenIndex346, depth346 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l347
					}
					position++
				l348:
					{
						position349, tokenIndex349, depth349 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l349
						}
						goto l348
					l349:
						position, tokenIndex, depth = position349, tokenIndex349, depth349
					}
					if buffer[position] != rune('`') {
						goto l347
					}
					position++
					goto l346
				l347:
					position, tokenIndex, depth = position346, tokenIndex346, depth346
					if !_rules[rule_]() {
						goto l344
					}
					{
						position350, tokenIndex350, depth350 := position, tokenIndex, depth
						{
							position351 := position
							depth++
							{
								position352, tokenIndex352, depth352 := position, tokenIndex, depth
								{
									position354, tokenIndex354, depth354 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l355
									}
									position++
									goto l354
								l355:
									position, tokenIndex, depth = position354, tokenIndex354, depth354
									if buffer[position] != rune('A') {
										goto l353
									}
									position++
								}
							l354:
								{
									position356, tokenIndex356, depth356 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l357
									}
									position++
									goto l356
								l357:
									position, tokenIndex, depth = position356, tokenIndex356, depth356
									if buffer[position] != rune('L') {
										goto l353
									}
									position++
								}
							l356:
								{
									position358, tokenIndex358, depth358 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l359
									}
									position++
									goto l358
								l359:
									position, tokenIndex, depth = position358, tokenIndex358, depth358
									if buffer[position] != rune('L') {
										goto l353
									}
									position++
								}
							l358:
								goto l352
							l353:
								position, tokenIndex, depth = position352, tokenIndex352, depth352
								{
									position361, tokenIndex361, depth361 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l362
									}
									position++
									goto l361
								l362:
									position, tokenIndex, depth = position361, tokenIndex361, depth361
									if buffer[position] != rune('A') {
										goto l360
									}
									position++
								}
							l361:
								{
									position363, tokenIndex363, depth363 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l364
									}
									position++
									goto l363
								l364:
									position, tokenIndex, depth = position363, tokenIndex363, depth363
									if buffer[position] != rune('N') {
										goto l360
									}
									position++
								}
							l363:
								{
									position365, tokenIndex365, depth365 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l366
									}
									position++
									goto l365
								l366:
									position, tokenIndex, depth = position365, tokenIndex365, depth365
									if buffer[position] != rune('D') {
										goto l360
									}
									position++
								}
							l365:
								goto l352
							l360:
								position, tokenIndex, depth = position352, tokenIndex352, depth352
								{
									position368, tokenIndex368, depth368 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l369
									}
									position++
									goto l368
								l369:
									position, tokenIndex, depth = position368, tokenIndex368, depth368
									if buffer[position] != rune('M') {
										goto l367
									}
									position++
								}
							l368:
								{
									position370, tokenIndex370, depth370 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l371
									}
									position++
									goto l370
								l371:
									position, tokenIndex, depth = position370, tokenIndex370, depth370
									if buffer[position] != rune('A') {
										goto l367
									}
									position++
								}
							l370:
								{
									position372, tokenIndex372, depth372 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l373
									}
									position++
									goto l372
								l373:
									position, tokenIndex, depth = position372, tokenIndex372, depth372
									if buffer[position] != rune('T') {
										goto l367
									}
									position++
								}
							l372:
								{
									position374, tokenIndex374, depth374 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l375
									}
									position++
									goto l374
								l375:
									position, tokenIndex, depth = position374, tokenIndex374, depth374
									if buffer[position] != rune('C') {
										goto l367
									}
									position++
								}
							l374:
								{
									position376, tokenIndex376, depth376 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l377
									}
									position++
									goto l376
								l377:
									position, tokenIndex, depth = position376, tokenIndex376, depth376
									if buffer[position] != rune('H') {
										goto l367
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
										goto l367
									}
									position++
								}
							l378:
								{
									position380, tokenIndex380, depth380 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l381
									}
									position++
									goto l380
								l381:
									position, tokenIndex, depth = position380, tokenIndex380, depth380
									if buffer[position] != rune('S') {
										goto l367
									}
									position++
								}
							l380:
								goto l352
							l367:
								position, tokenIndex, depth = position352, tokenIndex352, depth352
								{
									position383, tokenIndex383, depth383 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l384
									}
									position++
									goto l383
								l384:
									position, tokenIndex, depth = position383, tokenIndex383, depth383
									if buffer[position] != rune('S') {
										goto l382
									}
									position++
								}
							l383:
								{
									position385, tokenIndex385, depth385 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l386
									}
									position++
									goto l385
								l386:
									position, tokenIndex, depth = position385, tokenIndex385, depth385
									if buffer[position] != rune('E') {
										goto l382
									}
									position++
								}
							l385:
								{
									position387, tokenIndex387, depth387 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l388
									}
									position++
									goto l387
								l388:
									position, tokenIndex, depth = position387, tokenIndex387, depth387
									if buffer[position] != rune('L') {
										goto l382
									}
									position++
								}
							l387:
								{
									position389, tokenIndex389, depth389 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l390
									}
									position++
									goto l389
								l390:
									position, tokenIndex, depth = position389, tokenIndex389, depth389
									if buffer[position] != rune('E') {
										goto l382
									}
									position++
								}
							l389:
								{
									position391, tokenIndex391, depth391 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l392
									}
									position++
									goto l391
								l392:
									position, tokenIndex, depth = position391, tokenIndex391, depth391
									if buffer[position] != rune('C') {
										goto l382
									}
									position++
								}
							l391:
								{
									position393, tokenIndex393, depth393 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l394
									}
									position++
									goto l393
								l394:
									position, tokenIndex, depth = position393, tokenIndex393, depth393
									if buffer[position] != rune('T') {
										goto l382
									}
									position++
								}
							l393:
								goto l352
							l382:
								position, tokenIndex, depth = position352, tokenIndex352, depth352
								{
									switch buffer[position] {
									case 'M', 'm':
										{
											position396, tokenIndex396, depth396 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l397
											}
											position++
											goto l396
										l397:
											position, tokenIndex, depth = position396, tokenIndex396, depth396
											if buffer[position] != rune('M') {
												goto l350
											}
											position++
										}
									l396:
										{
											position398, tokenIndex398, depth398 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l399
											}
											position++
											goto l398
										l399:
											position, tokenIndex, depth = position398, tokenIndex398, depth398
											if buffer[position] != rune('E') {
												goto l350
											}
											position++
										}
									l398:
										{
											position400, tokenIndex400, depth400 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l401
											}
											position++
											goto l400
										l401:
											position, tokenIndex, depth = position400, tokenIndex400, depth400
											if buffer[position] != rune('T') {
												goto l350
											}
											position++
										}
									l400:
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
												goto l350
											}
											position++
										}
									l402:
										{
											position404, tokenIndex404, depth404 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l405
											}
											position++
											goto l404
										l405:
											position, tokenIndex, depth = position404, tokenIndex404, depth404
											if buffer[position] != rune('I') {
												goto l350
											}
											position++
										}
									l404:
										{
											position406, tokenIndex406, depth406 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l407
											}
											position++
											goto l406
										l407:
											position, tokenIndex, depth = position406, tokenIndex406, depth406
											if buffer[position] != rune('C') {
												goto l350
											}
											position++
										}
									l406:
										{
											position408, tokenIndex408, depth408 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l409
											}
											position++
											goto l408
										l409:
											position, tokenIndex, depth = position408, tokenIndex408, depth408
											if buffer[position] != rune('S') {
												goto l350
											}
											position++
										}
									l408:
										break
									case 'W', 'w':
										{
											position410, tokenIndex410, depth410 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l411
											}
											position++
											goto l410
										l411:
											position, tokenIndex, depth = position410, tokenIndex410, depth410
											if buffer[position] != rune('W') {
												goto l350
											}
											position++
										}
									l410:
										{
											position412, tokenIndex412, depth412 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l413
											}
											position++
											goto l412
										l413:
											position, tokenIndex, depth = position412, tokenIndex412, depth412
											if buffer[position] != rune('H') {
												goto l350
											}
											position++
										}
									l412:
										{
											position414, tokenIndex414, depth414 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l415
											}
											position++
											goto l414
										l415:
											position, tokenIndex, depth = position414, tokenIndex414, depth414
											if buffer[position] != rune('E') {
												goto l350
											}
											position++
										}
									l414:
										{
											position416, tokenIndex416, depth416 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l417
											}
											position++
											goto l416
										l417:
											position, tokenIndex, depth = position416, tokenIndex416, depth416
											if buffer[position] != rune('R') {
												goto l350
											}
											position++
										}
									l416:
										{
											position418, tokenIndex418, depth418 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l419
											}
											position++
											goto l418
										l419:
											position, tokenIndex, depth = position418, tokenIndex418, depth418
											if buffer[position] != rune('E') {
												goto l350
											}
											position++
										}
									l418:
										break
									case 'O', 'o':
										{
											position420, tokenIndex420, depth420 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l421
											}
											position++
											goto l420
										l421:
											position, tokenIndex, depth = position420, tokenIndex420, depth420
											if buffer[position] != rune('O') {
												goto l350
											}
											position++
										}
									l420:
										{
											position422, tokenIndex422, depth422 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l423
											}
											position++
											goto l422
										l423:
											position, tokenIndex, depth = position422, tokenIndex422, depth422
											if buffer[position] != rune('R') {
												goto l350
											}
											position++
										}
									l422:
										break
									case 'N', 'n':
										{
											position424, tokenIndex424, depth424 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l425
											}
											position++
											goto l424
										l425:
											position, tokenIndex, depth = position424, tokenIndex424, depth424
											if buffer[position] != rune('N') {
												goto l350
											}
											position++
										}
									l424:
										{
											position426, tokenIndex426, depth426 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l427
											}
											position++
											goto l426
										l427:
											position, tokenIndex, depth = position426, tokenIndex426, depth426
											if buffer[position] != rune('O') {
												goto l350
											}
											position++
										}
									l426:
										{
											position428, tokenIndex428, depth428 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l429
											}
											position++
											goto l428
										l429:
											position, tokenIndex, depth = position428, tokenIndex428, depth428
											if buffer[position] != rune('T') {
												goto l350
											}
											position++
										}
									l428:
										break
									case 'I', 'i':
										{
											position430, tokenIndex430, depth430 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l431
											}
											position++
											goto l430
										l431:
											position, tokenIndex, depth = position430, tokenIndex430, depth430
											if buffer[position] != rune('I') {
												goto l350
											}
											position++
										}
									l430:
										{
											position432, tokenIndex432, depth432 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l433
											}
											position++
											goto l432
										l433:
											position, tokenIndex, depth = position432, tokenIndex432, depth432
											if buffer[position] != rune('N') {
												goto l350
											}
											position++
										}
									l432:
										break
									case 'C', 'c':
										{
											position434, tokenIndex434, depth434 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l435
											}
											position++
											goto l434
										l435:
											position, tokenIndex, depth = position434, tokenIndex434, depth434
											if buffer[position] != rune('C') {
												goto l350
											}
											position++
										}
									l434:
										{
											position436, tokenIndex436, depth436 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l437
											}
											position++
											goto l436
										l437:
											position, tokenIndex, depth = position436, tokenIndex436, depth436
											if buffer[position] != rune('O') {
												goto l350
											}
											position++
										}
									l436:
										{
											position438, tokenIndex438, depth438 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l439
											}
											position++
											goto l438
										l439:
											position, tokenIndex, depth = position438, tokenIndex438, depth438
											if buffer[position] != rune('L') {
												goto l350
											}
											position++
										}
									l438:
										{
											position440, tokenIndex440, depth440 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l441
											}
											position++
											goto l440
										l441:
											position, tokenIndex, depth = position440, tokenIndex440, depth440
											if buffer[position] != rune('L') {
												goto l350
											}
											position++
										}
									l440:
										{
											position442, tokenIndex442, depth442 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l443
											}
											position++
											goto l442
										l443:
											position, tokenIndex, depth = position442, tokenIndex442, depth442
											if buffer[position] != rune('A') {
												goto l350
											}
											position++
										}
									l442:
										{
											position444, tokenIndex444, depth444 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l445
											}
											position++
											goto l444
										l445:
											position, tokenIndex, depth = position444, tokenIndex444, depth444
											if buffer[position] != rune('P') {
												goto l350
											}
											position++
										}
									l444:
										{
											position446, tokenIndex446, depth446 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l447
											}
											position++
											goto l446
										l447:
											position, tokenIndex, depth = position446, tokenIndex446, depth446
											if buffer[position] != rune('S') {
												goto l350
											}
											position++
										}
									l446:
										{
											position448, tokenIndex448, depth448 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l449
											}
											position++
											goto l448
										l449:
											position, tokenIndex, depth = position448, tokenIndex448, depth448
											if buffer[position] != rune('E') {
												goto l350
											}
											position++
										}
									l448:
										break
									case 'G', 'g':
										{
											position450, tokenIndex450, depth450 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l451
											}
											position++
											goto l450
										l451:
											position, tokenIndex, depth = position450, tokenIndex450, depth450
											if buffer[position] != rune('G') {
												goto l350
											}
											position++
										}
									l450:
										{
											position452, tokenIndex452, depth452 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l453
											}
											position++
											goto l452
										l453:
											position, tokenIndex, depth = position452, tokenIndex452, depth452
											if buffer[position] != rune('R') {
												goto l350
											}
											position++
										}
									l452:
										{
											position454, tokenIndex454, depth454 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l455
											}
											position++
											goto l454
										l455:
											position, tokenIndex, depth = position454, tokenIndex454, depth454
											if buffer[position] != rune('O') {
												goto l350
											}
											position++
										}
									l454:
										{
											position456, tokenIndex456, depth456 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l457
											}
											position++
											goto l456
										l457:
											position, tokenIndex, depth = position456, tokenIndex456, depth456
											if buffer[position] != rune('U') {
												goto l350
											}
											position++
										}
									l456:
										{
											position458, tokenIndex458, depth458 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l459
											}
											position++
											goto l458
										l459:
											position, tokenIndex, depth = position458, tokenIndex458, depth458
											if buffer[position] != rune('P') {
												goto l350
											}
											position++
										}
									l458:
										break
									case 'D', 'd':
										{
											position460, tokenIndex460, depth460 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l461
											}
											position++
											goto l460
										l461:
											position, tokenIndex, depth = position460, tokenIndex460, depth460
											if buffer[position] != rune('D') {
												goto l350
											}
											position++
										}
									l460:
										{
											position462, tokenIndex462, depth462 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l463
											}
											position++
											goto l462
										l463:
											position, tokenIndex, depth = position462, tokenIndex462, depth462
											if buffer[position] != rune('E') {
												goto l350
											}
											position++
										}
									l462:
										{
											position464, tokenIndex464, depth464 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l465
											}
											position++
											goto l464
										l465:
											position, tokenIndex, depth = position464, tokenIndex464, depth464
											if buffer[position] != rune('S') {
												goto l350
											}
											position++
										}
									l464:
										{
											position466, tokenIndex466, depth466 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l467
											}
											position++
											goto l466
										l467:
											position, tokenIndex, depth = position466, tokenIndex466, depth466
											if buffer[position] != rune('C') {
												goto l350
											}
											position++
										}
									l466:
										{
											position468, tokenIndex468, depth468 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l469
											}
											position++
											goto l468
										l469:
											position, tokenIndex, depth = position468, tokenIndex468, depth468
											if buffer[position] != rune('R') {
												goto l350
											}
											position++
										}
									l468:
										{
											position470, tokenIndex470, depth470 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l471
											}
											position++
											goto l470
										l471:
											position, tokenIndex, depth = position470, tokenIndex470, depth470
											if buffer[position] != rune('I') {
												goto l350
											}
											position++
										}
									l470:
										{
											position472, tokenIndex472, depth472 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l473
											}
											position++
											goto l472
										l473:
											position, tokenIndex, depth = position472, tokenIndex472, depth472
											if buffer[position] != rune('B') {
												goto l350
											}
											position++
										}
									l472:
										{
											position474, tokenIndex474, depth474 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l475
											}
											position++
											goto l474
										l475:
											position, tokenIndex, depth = position474, tokenIndex474, depth474
											if buffer[position] != rune('E') {
												goto l350
											}
											position++
										}
									l474:
										break
									case 'B', 'b':
										{
											position476, tokenIndex476, depth476 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l477
											}
											position++
											goto l476
										l477:
											position, tokenIndex, depth = position476, tokenIndex476, depth476
											if buffer[position] != rune('B') {
												goto l350
											}
											position++
										}
									l476:
										{
											position478, tokenIndex478, depth478 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l479
											}
											position++
											goto l478
										l479:
											position, tokenIndex, depth = position478, tokenIndex478, depth478
											if buffer[position] != rune('Y') {
												goto l350
											}
											position++
										}
									l478:
										break
									case 'A', 'a':
										{
											position480, tokenIndex480, depth480 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l481
											}
											position++
											goto l480
										l481:
											position, tokenIndex, depth = position480, tokenIndex480, depth480
											if buffer[position] != rune('A') {
												goto l350
											}
											position++
										}
									l480:
										{
											position482, tokenIndex482, depth482 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l483
											}
											position++
											goto l482
										l483:
											position, tokenIndex, depth = position482, tokenIndex482, depth482
											if buffer[position] != rune('S') {
												goto l350
											}
											position++
										}
									l482:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l350
										}
										break
									}
								}

							}
						l352:
							depth--
							add(ruleKEYWORD, position351)
						}
						if !_rules[ruleKEY]() {
							goto l350
						}
						goto l344
					l350:
						position, tokenIndex, depth = position350, tokenIndex350, depth350
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l344
					}
				l484:
					{
						position485, tokenIndex485, depth485 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l485
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l485
						}
						goto l484
					l485:
						position, tokenIndex, depth = position485, tokenIndex485, depth485
					}
				}
			l346:
				depth--
				add(ruleIDENTIFIER, position345)
			}
			return true
		l344:
			position, tokenIndex, depth = position344, tokenIndex344, depth344
			return false
		},
		/* 32 TIMESTAMP <- <((_ <(NUMBER ([a-z] / [A-Z])*)>) / (_ STRING) / (_ <(('n' / 'N') ('o' / 'O') ('w' / 'W'))>))> */
		nil,
		/* 33 ID_SEGMENT <- <(_ ID_START ID_CONT*)> */
		func() bool {
			position487, tokenIndex487, depth487 := position, tokenIndex, depth
			{
				position488 := position
				depth++
				if !_rules[rule_]() {
					goto l487
				}
				if !_rules[ruleID_START]() {
					goto l487
				}
			l489:
				{
					position490, tokenIndex490, depth490 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l490
					}
					goto l489
				l490:
					position, tokenIndex, depth = position490, tokenIndex490, depth490
				}
				depth--
				add(ruleID_SEGMENT, position488)
			}
			return true
		l487:
			position, tokenIndex, depth = position487, tokenIndex487, depth487
			return false
		},
		/* 34 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position491, tokenIndex491, depth491 := position, tokenIndex, depth
			{
				position492 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l491
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l491
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l491
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position492)
			}
			return true
		l491:
			position, tokenIndex, depth = position491, tokenIndex491, depth491
			return false
		},
		/* 35 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position494, tokenIndex494, depth494 := position, tokenIndex, depth
			{
				position495 := position
				depth++
				{
					position496, tokenIndex496, depth496 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l497
					}
					goto l496
				l497:
					position, tokenIndex, depth = position496, tokenIndex496, depth496
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l494
					}
					position++
				}
			l496:
				depth--
				add(ruleID_CONT, position495)
			}
			return true
		l494:
			position, tokenIndex, depth = position494, tokenIndex494, depth494
			return false
		},
		/* 36 PROPERTY_KEY <- <(((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> KEY _ (('b' / 'B') ('y' / 'Y')))) | (&('R' | 'r') <(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))>) | (&('T' | 't') <(('t' / 'T') ('o' / 'O'))>) | (&('F' | 'f') <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))>)) KEY)> */
		func() bool {
			position498, tokenIndex498, depth498 := position, tokenIndex, depth
			{
				position499 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position501 := position
							depth++
							{
								position502, tokenIndex502, depth502 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l503
								}
								position++
								goto l502
							l503:
								position, tokenIndex, depth = position502, tokenIndex502, depth502
								if buffer[position] != rune('S') {
									goto l498
								}
								position++
							}
						l502:
							{
								position504, tokenIndex504, depth504 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l505
								}
								position++
								goto l504
							l505:
								position, tokenIndex, depth = position504, tokenIndex504, depth504
								if buffer[position] != rune('A') {
									goto l498
								}
								position++
							}
						l504:
							{
								position506, tokenIndex506, depth506 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l507
								}
								position++
								goto l506
							l507:
								position, tokenIndex, depth = position506, tokenIndex506, depth506
								if buffer[position] != rune('M') {
									goto l498
								}
								position++
							}
						l506:
							{
								position508, tokenIndex508, depth508 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l509
								}
								position++
								goto l508
							l509:
								position, tokenIndex, depth = position508, tokenIndex508, depth508
								if buffer[position] != rune('P') {
									goto l498
								}
								position++
							}
						l508:
							{
								position510, tokenIndex510, depth510 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l511
								}
								position++
								goto l510
							l511:
								position, tokenIndex, depth = position510, tokenIndex510, depth510
								if buffer[position] != rune('L') {
									goto l498
								}
								position++
							}
						l510:
							{
								position512, tokenIndex512, depth512 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l513
								}
								position++
								goto l512
							l513:
								position, tokenIndex, depth = position512, tokenIndex512, depth512
								if buffer[position] != rune('E') {
									goto l498
								}
								position++
							}
						l512:
							depth--
							add(rulePegText, position501)
						}
						if !_rules[ruleKEY]() {
							goto l498
						}
						if !_rules[rule_]() {
							goto l498
						}
						{
							position514, tokenIndex514, depth514 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l515
							}
							position++
							goto l514
						l515:
							position, tokenIndex, depth = position514, tokenIndex514, depth514
							if buffer[position] != rune('B') {
								goto l498
							}
							position++
						}
					l514:
						{
							position516, tokenIndex516, depth516 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l517
							}
							position++
							goto l516
						l517:
							position, tokenIndex, depth = position516, tokenIndex516, depth516
							if buffer[position] != rune('Y') {
								goto l498
							}
							position++
						}
					l516:
						break
					case 'R', 'r':
						{
							position518 := position
							depth++
							{
								position519, tokenIndex519, depth519 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l520
								}
								position++
								goto l519
							l520:
								position, tokenIndex, depth = position519, tokenIndex519, depth519
								if buffer[position] != rune('R') {
									goto l498
								}
								position++
							}
						l519:
							{
								position521, tokenIndex521, depth521 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l522
								}
								position++
								goto l521
							l522:
								position, tokenIndex, depth = position521, tokenIndex521, depth521
								if buffer[position] != rune('E') {
									goto l498
								}
								position++
							}
						l521:
							{
								position523, tokenIndex523, depth523 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l524
								}
								position++
								goto l523
							l524:
								position, tokenIndex, depth = position523, tokenIndex523, depth523
								if buffer[position] != rune('S') {
									goto l498
								}
								position++
							}
						l523:
							{
								position525, tokenIndex525, depth525 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l526
								}
								position++
								goto l525
							l526:
								position, tokenIndex, depth = position525, tokenIndex525, depth525
								if buffer[position] != rune('O') {
									goto l498
								}
								position++
							}
						l525:
							{
								position527, tokenIndex527, depth527 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l528
								}
								position++
								goto l527
							l528:
								position, tokenIndex, depth = position527, tokenIndex527, depth527
								if buffer[position] != rune('L') {
									goto l498
								}
								position++
							}
						l527:
							{
								position529, tokenIndex529, depth529 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l530
								}
								position++
								goto l529
							l530:
								position, tokenIndex, depth = position529, tokenIndex529, depth529
								if buffer[position] != rune('U') {
									goto l498
								}
								position++
							}
						l529:
							{
								position531, tokenIndex531, depth531 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l532
								}
								position++
								goto l531
							l532:
								position, tokenIndex, depth = position531, tokenIndex531, depth531
								if buffer[position] != rune('T') {
									goto l498
								}
								position++
							}
						l531:
							{
								position533, tokenIndex533, depth533 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l534
								}
								position++
								goto l533
							l534:
								position, tokenIndex, depth = position533, tokenIndex533, depth533
								if buffer[position] != rune('I') {
									goto l498
								}
								position++
							}
						l533:
							{
								position535, tokenIndex535, depth535 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l536
								}
								position++
								goto l535
							l536:
								position, tokenIndex, depth = position535, tokenIndex535, depth535
								if buffer[position] != rune('O') {
									goto l498
								}
								position++
							}
						l535:
							{
								position537, tokenIndex537, depth537 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l538
								}
								position++
								goto l537
							l538:
								position, tokenIndex, depth = position537, tokenIndex537, depth537
								if buffer[position] != rune('N') {
									goto l498
								}
								position++
							}
						l537:
							depth--
							add(rulePegText, position518)
						}
						break
					case 'T', 't':
						{
							position539 := position
							depth++
							{
								position540, tokenIndex540, depth540 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l541
								}
								position++
								goto l540
							l541:
								position, tokenIndex, depth = position540, tokenIndex540, depth540
								if buffer[position] != rune('T') {
									goto l498
								}
								position++
							}
						l540:
							{
								position542, tokenIndex542, depth542 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l543
								}
								position++
								goto l542
							l543:
								position, tokenIndex, depth = position542, tokenIndex542, depth542
								if buffer[position] != rune('O') {
									goto l498
								}
								position++
							}
						l542:
							depth--
							add(rulePegText, position539)
						}
						break
					default:
						{
							position544 := position
							depth++
							{
								position545, tokenIndex545, depth545 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l546
								}
								position++
								goto l545
							l546:
								position, tokenIndex, depth = position545, tokenIndex545, depth545
								if buffer[position] != rune('F') {
									goto l498
								}
								position++
							}
						l545:
							{
								position547, tokenIndex547, depth547 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l548
								}
								position++
								goto l547
							l548:
								position, tokenIndex, depth = position547, tokenIndex547, depth547
								if buffer[position] != rune('R') {
									goto l498
								}
								position++
							}
						l547:
							{
								position549, tokenIndex549, depth549 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l550
								}
								position++
								goto l549
							l550:
								position, tokenIndex, depth = position549, tokenIndex549, depth549
								if buffer[position] != rune('O') {
									goto l498
								}
								position++
							}
						l549:
							{
								position551, tokenIndex551, depth551 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l552
								}
								position++
								goto l551
							l552:
								position, tokenIndex, depth = position551, tokenIndex551, depth551
								if buffer[position] != rune('M') {
									goto l498
								}
								position++
							}
						l551:
							depth--
							add(rulePegText, position544)
						}
						break
					}
				}

				if !_rules[ruleKEY]() {
					goto l498
				}
				depth--
				add(rulePROPERTY_KEY, position499)
			}
			return true
		l498:
			position, tokenIndex, depth = position498, tokenIndex498, depth498
			return false
		},
		/* 37 PROPERTY_VALUE <- <TIMESTAMP> */
		nil,
		/* 38 KEYWORD <- <((('a' / 'A') ('l' / 'L') ('l' / 'L')) / (('a' / 'A') ('n' / 'N') ('d' / 'D')) / (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) / (('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T')) / ((&('M' | 'm') (('m' / 'M') ('e' / 'E') ('t' / 'T') ('r' / 'R') ('i' / 'I') ('c' / 'C') ('s' / 'S'))) | (&('W' | 'w') (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E'))) | (&('O' | 'o') (('o' / 'O') ('r' / 'R'))) | (&('N' | 'n') (('n' / 'N') ('o' / 'O') ('t' / 'T'))) | (&('I' | 'i') (('i' / 'I') ('n' / 'N'))) | (&('C' | 'c') (('c' / 'C') ('o' / 'O') ('l' / 'L') ('l' / 'L') ('a' / 'A') ('p' / 'P') ('s' / 'S') ('e' / 'E'))) | (&('G' | 'g') (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P'))) | (&('D' | 'd') (('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E'))) | (&('B' | 'b') (('b' / 'B') ('y' / 'Y'))) | (&('A' | 'a') (('a' / 'A') ('s' / 'S'))) | (&('F' | 'R' | 'S' | 'T' | 'f' | 'r' | 's' | 't') PROPERTY_KEY)))> */
		nil,
		/* 39 OP_PIPE <- <'|'> */
		nil,
		/* 40 OP_ADD <- <'+'> */
		nil,
		/* 41 OP_SUB <- <'-'> */
		nil,
		/* 42 OP_MULT <- <'*'> */
		nil,
		/* 43 OP_DIV <- <'/'> */
		nil,
		/* 44 OP_AND <- <(('a' / 'A') ('n' / 'N') ('d' / 'D') KEY)> */
		nil,
		/* 45 OP_OR <- <(('o' / 'O') ('r' / 'R') KEY)> */
		nil,
		/* 46 OP_NOT <- <(('n' / 'N') ('o' / 'O') ('t' / 'T') KEY)> */
		nil,
		/* 47 QUOTE_SINGLE <- <'\''> */
		func() bool {
			position563, tokenIndex563, depth563 := position, tokenIndex, depth
			{
				position564 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l563
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position564)
			}
			return true
		l563:
			position, tokenIndex, depth = position563, tokenIndex563, depth563
			return false
		},
		/* 48 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position565, tokenIndex565, depth565 := position, tokenIndex, depth
			{
				position566 := position
				depth++
				if buffer[position] != rune('"') {
					goto l565
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position566)
			}
			return true
		l565:
			position, tokenIndex, depth = position565, tokenIndex565, depth565
			return false
		},
		/* 49 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> QUOTE_SINGLE) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> QUOTE_DOUBLE))> */
		func() bool {
			position567, tokenIndex567, depth567 := position, tokenIndex, depth
			{
				position568 := position
				depth++
				{
					position569, tokenIndex569, depth569 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l570
					}
					{
						position571 := position
						depth++
					l572:
						{
							position573, tokenIndex573, depth573 := position, tokenIndex, depth
							{
								position574, tokenIndex574, depth574 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l574
								}
								goto l573
							l574:
								position, tokenIndex, depth = position574, tokenIndex574, depth574
							}
							if !_rules[ruleCHAR]() {
								goto l573
							}
							goto l572
						l573:
							position, tokenIndex, depth = position573, tokenIndex573, depth573
						}
						depth--
						add(rulePegText, position571)
					}
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l570
					}
					goto l569
				l570:
					position, tokenIndex, depth = position569, tokenIndex569, depth569
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l567
					}
					{
						position575 := position
						depth++
					l576:
						{
							position577, tokenIndex577, depth577 := position, tokenIndex, depth
							{
								position578, tokenIndex578, depth578 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l578
								}
								goto l577
							l578:
								position, tokenIndex, depth = position578, tokenIndex578, depth578
							}
							if !_rules[ruleCHAR]() {
								goto l577
							}
							goto l576
						l577:
							position, tokenIndex, depth = position577, tokenIndex577, depth577
						}
						depth--
						add(rulePegText, position575)
					}
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l567
					}
				}
			l569:
				depth--
				add(ruleSTRING, position568)
			}
			return true
		l567:
			position, tokenIndex, depth = position567, tokenIndex567, depth567
			return false
		},
		/* 50 CHAR <- <(('\\' ((&('"') QUOTE_DOUBLE) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position579, tokenIndex579, depth579 := position, tokenIndex, depth
			{
				position580 := position
				depth++
				{
					position581, tokenIndex581, depth581 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l582
					}
					position++
					{
						switch buffer[position] {
						case '"':
							if !_rules[ruleQUOTE_DOUBLE]() {
								goto l582
							}
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l582
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l582
							}
							break
						}
					}

					goto l581
				l582:
					position, tokenIndex, depth = position581, tokenIndex581, depth581
					{
						position584, tokenIndex584, depth584 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l584
						}
						goto l579
					l584:
						position, tokenIndex, depth = position584, tokenIndex584, depth584
					}
					if !matchDot() {
						goto l579
					}
				}
			l581:
				depth--
				add(ruleCHAR, position580)
			}
			return true
		l579:
			position, tokenIndex, depth = position579, tokenIndex579, depth579
			return false
		},
		/* 51 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position585, tokenIndex585, depth585 := position, tokenIndex, depth
			{
				position586 := position
				depth++
				{
					position587, tokenIndex587, depth587 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l588
					}
					position++
					goto l587
				l588:
					position, tokenIndex, depth = position587, tokenIndex587, depth587
					if buffer[position] != rune('\\') {
						goto l585
					}
					position++
				}
			l587:
				depth--
				add(ruleESCAPE_CLASS, position586)
			}
			return true
		l585:
			position, tokenIndex, depth = position585, tokenIndex585, depth585
			return false
		},
		/* 52 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position589, tokenIndex589, depth589 := position, tokenIndex, depth
			{
				position590 := position
				depth++
				{
					position591 := position
					depth++
					{
						position592, tokenIndex592, depth592 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l592
						}
						position++
						goto l593
					l592:
						position, tokenIndex, depth = position592, tokenIndex592, depth592
					}
				l593:
					{
						position594 := position
						depth++
						{
							position595, tokenIndex595, depth595 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l596
							}
							position++
							goto l595
						l596:
							position, tokenIndex, depth = position595, tokenIndex595, depth595
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l589
							}
							position++
						l597:
							{
								position598, tokenIndex598, depth598 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l598
								}
								position++
								goto l597
							l598:
								position, tokenIndex, depth = position598, tokenIndex598, depth598
							}
						}
					l595:
						depth--
						add(ruleNUMBER_NATURAL, position594)
					}
					depth--
					add(ruleNUMBER_INTEGER, position591)
				}
				{
					position599, tokenIndex599, depth599 := position, tokenIndex, depth
					{
						position601 := position
						depth++
						if buffer[position] != rune('.') {
							goto l599
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l599
						}
						position++
					l602:
						{
							position603, tokenIndex603, depth603 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l603
							}
							position++
							goto l602
						l603:
							position, tokenIndex, depth = position603, tokenIndex603, depth603
						}
						depth--
						add(ruleNUMBER_FRACTION, position601)
					}
					goto l600
				l599:
					position, tokenIndex, depth = position599, tokenIndex599, depth599
				}
			l600:
				{
					position604, tokenIndex604, depth604 := position, tokenIndex, depth
					{
						position606 := position
						depth++
						{
							position607, tokenIndex607, depth607 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l608
							}
							position++
							goto l607
						l608:
							position, tokenIndex, depth = position607, tokenIndex607, depth607
							if buffer[position] != rune('E') {
								goto l604
							}
							position++
						}
					l607:
						{
							position609, tokenIndex609, depth609 := position, tokenIndex, depth
							{
								position611, tokenIndex611, depth611 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l612
								}
								position++
								goto l611
							l612:
								position, tokenIndex, depth = position611, tokenIndex611, depth611
								if buffer[position] != rune('-') {
									goto l609
								}
								position++
							}
						l611:
							goto l610
						l609:
							position, tokenIndex, depth = position609, tokenIndex609, depth609
						}
					l610:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l604
						}
						position++
					l613:
						{
							position614, tokenIndex614, depth614 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l614
							}
							position++
							goto l613
						l614:
							position, tokenIndex, depth = position614, tokenIndex614, depth614
						}
						depth--
						add(ruleNUMBER_EXP, position606)
					}
					goto l605
				l604:
					position, tokenIndex, depth = position604, tokenIndex604, depth604
				}
			l605:
				depth--
				add(ruleNUMBER, position590)
			}
			return true
		l589:
			position, tokenIndex, depth = position589, tokenIndex589, depth589
			return false
		},
		/* 53 NUMBER_NATURAL <- <('0' / ([1-9] [0-9]*))> */
		nil,
		/* 54 NUMBER_FRACTION <- <('.' [0-9]+)> */
		nil,
		/* 55 NUMBER_INTEGER <- <('-'? NUMBER_NATURAL)> */
		nil,
		/* 56 NUMBER_EXP <- <(('e' / 'E') ('+' / '-')? [0-9]+)> */
		nil,
		/* 57 DURATION <- <(NUMBER [a-z]+ KEY)> */
		nil,
		/* 58 PAREN_OPEN <- <'('> */
		func() bool {
			position620, tokenIndex620, depth620 := position, tokenIndex, depth
			{
				position621 := position
				depth++
				if buffer[position] != rune('(') {
					goto l620
				}
				position++
				depth--
				add(rulePAREN_OPEN, position621)
			}
			return true
		l620:
			position, tokenIndex, depth = position620, tokenIndex620, depth620
			return false
		},
		/* 59 PAREN_CLOSE <- <')'> */
		func() bool {
			position622, tokenIndex622, depth622 := position, tokenIndex, depth
			{
				position623 := position
				depth++
				if buffer[position] != rune(')') {
					goto l622
				}
				position++
				depth--
				add(rulePAREN_CLOSE, position623)
			}
			return true
		l622:
			position, tokenIndex, depth = position622, tokenIndex622, depth622
			return false
		},
		/* 60 COMMA <- <','> */
		func() bool {
			position624, tokenIndex624, depth624 := position, tokenIndex, depth
			{
				position625 := position
				depth++
				if buffer[position] != rune(',') {
					goto l624
				}
				position++
				depth--
				add(ruleCOMMA, position625)
			}
			return true
		l624:
			position, tokenIndex, depth = position624, tokenIndex624, depth624
			return false
		},
		/* 61 _ <- <SPACE*> */
		func() bool {
			{
				position627 := position
				depth++
			l628:
				{
					position629, tokenIndex629, depth629 := position, tokenIndex, depth
					{
						position630 := position
						depth++
						{
							switch buffer[position] {
							case '\t':
								if buffer[position] != rune('\t') {
									goto l629
								}
								position++
								break
							case '\n':
								if buffer[position] != rune('\n') {
									goto l629
								}
								position++
								break
							default:
								if buffer[position] != rune(' ') {
									goto l629
								}
								position++
								break
							}
						}

						depth--
						add(ruleSPACE, position630)
					}
					goto l628
				l629:
					position, tokenIndex, depth = position629, tokenIndex629, depth629
				}
				depth--
				add(rule_, position627)
			}
			return true
		},
		/* 62 KEY <- <!ID_CONT> */
		func() bool {
			position632, tokenIndex632, depth632 := position, tokenIndex, depth
			{
				position633 := position
				depth++
				{
					position634, tokenIndex634, depth634 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l634
					}
					goto l632
				l634:
					position, tokenIndex, depth = position634, tokenIndex634, depth634
				}
				depth--
				add(ruleKEY, position633)
			}
			return true
		l632:
			position, tokenIndex, depth = position632, tokenIndex632, depth632
			return false
		},
		/* 63 SPACE <- <((&('\t') '\t') | (&('\n') '\n') | (&(' ') ' '))> */
		nil,
		/* 65 Action0 <- <{
		   p.makeSelect()
		 }> */
		nil,
		/* 66 Action1 <- <{ p.makeDescribeAll() }> */
		nil,
		/* 67 Action2 <- <{ p.makeDescribeMetrics() }> */
		nil,
		nil,
		/* 69 Action3 <- <{ p.addStringLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 70 Action4 <- <{ p.makeDescribe() }> */
		nil,
		/* 71 Action5 <- <{ p.addEvaluationContext() }> */
		nil,
		/* 72 Action6 <- <{ p.addPropertyKey(buffer[begin:end])   }> */
		nil,
		/* 73 Action7 <- <{ p.addPropertyValue(buffer[begin:end]) }> */
		nil,
		/* 74 Action8 <- <{ p.insertPropertyKeyValue() }> */
		nil,
		/* 75 Action9 <- <{ p.checkPropertyClause() }> */
		nil,
		/* 76 Action10 <- <{ p.addNullPredicate() }> */
		nil,
		/* 77 Action11 <- <{ p.addExpressionList() }> */
		nil,
		/* 78 Action12 <- <{ p.appendExpression() }> */
		nil,
		/* 79 Action13 <- <{ p.appendExpression() }> */
		nil,
		/* 80 Action14 <- <{ p.addOperatorLiteral("+") }> */
		nil,
		/* 81 Action15 <- <{ p.addOperatorLiteral("-") }> */
		nil,
		/* 82 Action16 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 83 Action17 <- <{ p.addOperatorLiteral("/") }> */
		nil,
		/* 84 Action18 <- <{ p.addOperatorLiteral("*") }> */
		nil,
		/* 85 Action19 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 86 Action20 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 87 Action21 <- <{p.addExpressionList()}> */
		nil,
		/* 88 Action22 <- <{
		   p.addExpressionList()
		   p.addGroupBy()
		 }> */
		nil,
		/* 89 Action23 <- <{
		   p.addPipeExpression()
		 }> */
		nil,
		/* 90 Action24 <- <{ p.addDurationNode(text) }> */
		nil,
		/* 91 Action25 <- <{ p.addNumberNode(buffer[begin:end]) }> */
		nil,
		/* 92 Action26 <- <{ p.addStringNode(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 93 Action27 <- <{ p.addGroupBy() }> */
		nil,
		/* 94 Action28 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 95 Action29 <- <{
		   p.addFunctionInvocation()
		 }> */
		nil,
		/* 96 Action30 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 97 Action31 <- <{ p.addNullPredicate() }> */
		nil,
		/* 98 Action32 <- <{
		   p.addMetricExpression()
		 }> */
		nil,
		/* 99 Action33 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 100 Action34 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 101 Action35 <- <{
		   p.appendCollapseBy(unescapeLiteral(text))
		 }> */
		nil,
		/* 102 Action36 <- <{p.appendCollapseBy(unescapeLiteral(text))}> */
		nil,
		/* 103 Action37 <- <{ p.addOrPredicate() }> */
		nil,
		/* 104 Action38 <- <{ p.addAndPredicate() }> */
		nil,
		/* 105 Action39 <- <{ p.addNotPredicate() }> */
		nil,
		/* 106 Action40 <- <{
		   p.addLiteralMatcher()
		 }> */
		nil,
		/* 107 Action41 <- <{
		   p.addLiteralMatcher()
		   p.addNotPredicate()
		 }> */
		nil,
		/* 108 Action42 <- <{
		   p.addRegexMatcher()
		 }> */
		nil,
		/* 109 Action43 <- <{
		   p.addListMatcher()
		 }> */
		nil,
		/* 110 Action44 <- <{
		  p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 111 Action45 <- <{ p.addLiteralList() }> */
		nil,
		/* 112 Action46 <- <{
		  p.appendLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 113 Action47 <- <{ p.addTagLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
	}
	p.rules = _rules
}
