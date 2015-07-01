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
	rules  [111]func() bool
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
			p.addGroupBy()
		case ruleAction23:

			p.addExpressionList()
			p.addGroupBy()

		case ruleAction24:

			p.addPipeExpression()

		case ruleAction25:
			p.addDurationNode(text)
		case ruleAction26:
			p.addNumberNode(buffer[begin:end])
		case ruleAction27:
			p.addStringNode(unescapeLiteral(buffer[begin:end]))
		case ruleAction28:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction29:
			p.addGroupBy()
		case ruleAction30:

			p.addFunctionInvocation()

		case ruleAction31:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction32:
			p.addNullPredicate()
		case ruleAction33:

			p.addMetricExpression()

		case ruleAction34:

			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))

		case ruleAction35:

			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))

		case ruleAction36:
			p.addOrPredicate()
		case ruleAction37:
			p.addAndPredicate()
		case ruleAction38:
			p.addNotPredicate()
		case ruleAction39:

			p.addLiteralMatcher()

		case ruleAction40:

			p.addLiteralMatcher()
			p.addNotPredicate()

		case ruleAction41:

			p.addRegexMatcher()

		case ruleAction42:

			p.addListMatcher()

		case ruleAction43:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction44:
			p.addLiteralList()
		case ruleAction45:

			p.appendLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction46:
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
		/* 12 add_pipe <- <(_ OP_PIPE _ <IDENTIFIER> Action20 ((_ PAREN_OPEN (expressionList / Action21) Action22 groupByClause? _ PAREN_CLOSE) / Action23) Action24)*> */
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
						{
							add(ruleAction22, position)
						}
						{
							position171, tokenIndex171, depth171 := position, tokenIndex, depth
							if !_rules[rulegroupByClause]() {
								goto l171
							}
							goto l172
						l171:
							position, tokenIndex, depth = position171, tokenIndex171, depth171
						}
					l172:
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
							add(ruleAction23, position)
						}
					}
				l165:
					{
						add(ruleAction24, position)
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
		/* 13 expression_atom <- <(expression_function / expression_metric / (_ PAREN_OPEN expression_start _ PAREN_CLOSE) / (_ <DURATION> Action25) / (_ <NUMBER> Action26) / (_ STRING Action27))> */
		func() bool {
			position175, tokenIndex175, depth175 := position, tokenIndex, depth
			{
				position176 := position
				depth++
				{
					position177, tokenIndex177, depth177 := position, tokenIndex, depth
					{
						position179 := position
						depth++
						if !_rules[rule_]() {
							goto l178
						}
						{
							position180 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l178
							}
							depth--
							add(rulePegText, position180)
						}
						{
							add(ruleAction28, position)
						}
						if !_rules[rule_]() {
							goto l178
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l178
						}
						if !_rules[ruleexpressionList]() {
							goto l178
						}
						{
							add(ruleAction29, position)
						}
						{
							position183, tokenIndex183, depth183 := position, tokenIndex, depth
							if !_rules[rulegroupByClause]() {
								goto l183
							}
							goto l184
						l183:
							position, tokenIndex, depth = position183, tokenIndex183, depth183
						}
					l184:
						if !_rules[rule_]() {
							goto l178
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l178
						}
						{
							add(ruleAction30, position)
						}
						depth--
						add(ruleexpression_function, position179)
					}
					goto l177
				l178:
					position, tokenIndex, depth = position177, tokenIndex177, depth177
					{
						position187 := position
						depth++
						if !_rules[rule_]() {
							goto l186
						}
						{
							position188 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l186
							}
							depth--
							add(rulePegText, position188)
						}
						{
							add(ruleAction31, position)
						}
						{
							position190, tokenIndex190, depth190 := position, tokenIndex, depth
							{
								position192, tokenIndex192, depth192 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l193
								}
								if buffer[position] != rune('[') {
									goto l193
								}
								position++
								if !_rules[rulepredicate_1]() {
									goto l193
								}
								if !_rules[rule_]() {
									goto l193
								}
								if buffer[position] != rune(']') {
									goto l193
								}
								position++
								goto l192
							l193:
								position, tokenIndex, depth = position192, tokenIndex192, depth192
								{
									add(ruleAction32, position)
								}
							}
						l192:
							goto l191

							position, tokenIndex, depth = position190, tokenIndex190, depth190
						}
					l191:
						{
							add(ruleAction33, position)
						}
						depth--
						add(ruleexpression_metric, position187)
					}
					goto l177
				l186:
					position, tokenIndex, depth = position177, tokenIndex177, depth177
					if !_rules[rule_]() {
						goto l196
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l196
					}
					if !_rules[ruleexpression_start]() {
						goto l196
					}
					if !_rules[rule_]() {
						goto l196
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l196
					}
					goto l177
				l196:
					position, tokenIndex, depth = position177, tokenIndex177, depth177
					if !_rules[rule_]() {
						goto l197
					}
					{
						position198 := position
						depth++
						{
							position199 := position
							depth++
							if !_rules[ruleNUMBER]() {
								goto l197
							}
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l197
							}
							position++
						l200:
							{
								position201, tokenIndex201, depth201 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l201
								}
								position++
								goto l200
							l201:
								position, tokenIndex, depth = position201, tokenIndex201, depth201
							}
							if !_rules[ruleKEY]() {
								goto l197
							}
							depth--
							add(ruleDURATION, position199)
						}
						depth--
						add(rulePegText, position198)
					}
					{
						add(ruleAction25, position)
					}
					goto l177
				l197:
					position, tokenIndex, depth = position177, tokenIndex177, depth177
					if !_rules[rule_]() {
						goto l203
					}
					{
						position204 := position
						depth++
						if !_rules[ruleNUMBER]() {
							goto l203
						}
						depth--
						add(rulePegText, position204)
					}
					{
						add(ruleAction26, position)
					}
					goto l177
				l203:
					position, tokenIndex, depth = position177, tokenIndex177, depth177
					if !_rules[rule_]() {
						goto l175
					}
					if !_rules[ruleSTRING]() {
						goto l175
					}
					{
						add(ruleAction27, position)
					}
				}
			l177:
				depth--
				add(ruleexpression_atom, position176)
			}
			return true
		l175:
			position, tokenIndex, depth = position175, tokenIndex175, depth175
			return false
		},
		/* 14 expression_function <- <(_ <IDENTIFIER> Action28 _ PAREN_OPEN expressionList Action29 groupByClause? _ PAREN_CLOSE Action30)> */
		nil,
		/* 15 expression_metric <- <(_ <IDENTIFIER> Action31 ((_ '[' predicate_1 _ ']') / Action32)? Action33)> */
		nil,
		/* 16 groupByClause <- <(_ (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P')) KEY _ (('b' / 'B') ('y' / 'Y')) KEY _ <COLUMN_NAME> Action34 (_ COMMA _ <COLUMN_NAME> Action35)*)> */
		func() bool {
			position209, tokenIndex209, depth209 := position, tokenIndex, depth
			{
				position210 := position
				depth++
				if !_rules[rule_]() {
					goto l209
				}
				{
					position211, tokenIndex211, depth211 := position, tokenIndex, depth
					if buffer[position] != rune('g') {
						goto l212
					}
					position++
					goto l211
				l212:
					position, tokenIndex, depth = position211, tokenIndex211, depth211
					if buffer[position] != rune('G') {
						goto l209
					}
					position++
				}
			l211:
				{
					position213, tokenIndex213, depth213 := position, tokenIndex, depth
					if buffer[position] != rune('r') {
						goto l214
					}
					position++
					goto l213
				l214:
					position, tokenIndex, depth = position213, tokenIndex213, depth213
					if buffer[position] != rune('R') {
						goto l209
					}
					position++
				}
			l213:
				{
					position215, tokenIndex215, depth215 := position, tokenIndex, depth
					if buffer[position] != rune('o') {
						goto l216
					}
					position++
					goto l215
				l216:
					position, tokenIndex, depth = position215, tokenIndex215, depth215
					if buffer[position] != rune('O') {
						goto l209
					}
					position++
				}
			l215:
				{
					position217, tokenIndex217, depth217 := position, tokenIndex, depth
					if buffer[position] != rune('u') {
						goto l218
					}
					position++
					goto l217
				l218:
					position, tokenIndex, depth = position217, tokenIndex217, depth217
					if buffer[position] != rune('U') {
						goto l209
					}
					position++
				}
			l217:
				{
					position219, tokenIndex219, depth219 := position, tokenIndex, depth
					if buffer[position] != rune('p') {
						goto l220
					}
					position++
					goto l219
				l220:
					position, tokenIndex, depth = position219, tokenIndex219, depth219
					if buffer[position] != rune('P') {
						goto l209
					}
					position++
				}
			l219:
				if !_rules[ruleKEY]() {
					goto l209
				}
				if !_rules[rule_]() {
					goto l209
				}
				{
					position221, tokenIndex221, depth221 := position, tokenIndex, depth
					if buffer[position] != rune('b') {
						goto l222
					}
					position++
					goto l221
				l222:
					position, tokenIndex, depth = position221, tokenIndex221, depth221
					if buffer[position] != rune('B') {
						goto l209
					}
					position++
				}
			l221:
				{
					position223, tokenIndex223, depth223 := position, tokenIndex, depth
					if buffer[position] != rune('y') {
						goto l224
					}
					position++
					goto l223
				l224:
					position, tokenIndex, depth = position223, tokenIndex223, depth223
					if buffer[position] != rune('Y') {
						goto l209
					}
					position++
				}
			l223:
				if !_rules[ruleKEY]() {
					goto l209
				}
				if !_rules[rule_]() {
					goto l209
				}
				{
					position225 := position
					depth++
					if !_rules[ruleCOLUMN_NAME]() {
						goto l209
					}
					depth--
					add(rulePegText, position225)
				}
				{
					add(ruleAction34, position)
				}
			l227:
				{
					position228, tokenIndex228, depth228 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l228
					}
					if !_rules[ruleCOMMA]() {
						goto l228
					}
					if !_rules[rule_]() {
						goto l228
					}
					{
						position229 := position
						depth++
						if !_rules[ruleCOLUMN_NAME]() {
							goto l228
						}
						depth--
						add(rulePegText, position229)
					}
					{
						add(ruleAction35, position)
					}
					goto l227
				l228:
					position, tokenIndex, depth = position228, tokenIndex228, depth228
				}
				depth--
				add(rulegroupByClause, position210)
			}
			return true
		l209:
			position, tokenIndex, depth = position209, tokenIndex209, depth209
			return false
		},
		/* 17 predicateClause <- <(_ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY _ predicate_1)> */
		nil,
		/* 18 predicate_1 <- <((predicate_2 _ OP_OR predicate_1 Action36) / predicate_2)> */
		func() bool {
			position232, tokenIndex232, depth232 := position, tokenIndex, depth
			{
				position233 := position
				depth++
				{
					position234, tokenIndex234, depth234 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l235
					}
					if !_rules[rule_]() {
						goto l235
					}
					{
						position236 := position
						depth++
						{
							position237, tokenIndex237, depth237 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l238
							}
							position++
							goto l237
						l238:
							position, tokenIndex, depth = position237, tokenIndex237, depth237
							if buffer[position] != rune('O') {
								goto l235
							}
							position++
						}
					l237:
						{
							position239, tokenIndex239, depth239 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l240
							}
							position++
							goto l239
						l240:
							position, tokenIndex, depth = position239, tokenIndex239, depth239
							if buffer[position] != rune('R') {
								goto l235
							}
							position++
						}
					l239:
						if !_rules[ruleKEY]() {
							goto l235
						}
						depth--
						add(ruleOP_OR, position236)
					}
					if !_rules[rulepredicate_1]() {
						goto l235
					}
					{
						add(ruleAction36, position)
					}
					goto l234
				l235:
					position, tokenIndex, depth = position234, tokenIndex234, depth234
					if !_rules[rulepredicate_2]() {
						goto l232
					}
				}
			l234:
				depth--
				add(rulepredicate_1, position233)
			}
			return true
		l232:
			position, tokenIndex, depth = position232, tokenIndex232, depth232
			return false
		},
		/* 19 predicate_2 <- <((predicate_3 _ OP_AND predicate_2 Action37) / predicate_3)> */
		func() bool {
			position242, tokenIndex242, depth242 := position, tokenIndex, depth
			{
				position243 := position
				depth++
				{
					position244, tokenIndex244, depth244 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l245
					}
					if !_rules[rule_]() {
						goto l245
					}
					{
						position246 := position
						depth++
						{
							position247, tokenIndex247, depth247 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l248
							}
							position++
							goto l247
						l248:
							position, tokenIndex, depth = position247, tokenIndex247, depth247
							if buffer[position] != rune('A') {
								goto l245
							}
							position++
						}
					l247:
						{
							position249, tokenIndex249, depth249 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l250
							}
							position++
							goto l249
						l250:
							position, tokenIndex, depth = position249, tokenIndex249, depth249
							if buffer[position] != rune('N') {
								goto l245
							}
							position++
						}
					l249:
						{
							position251, tokenIndex251, depth251 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l252
							}
							position++
							goto l251
						l252:
							position, tokenIndex, depth = position251, tokenIndex251, depth251
							if buffer[position] != rune('D') {
								goto l245
							}
							position++
						}
					l251:
						if !_rules[ruleKEY]() {
							goto l245
						}
						depth--
						add(ruleOP_AND, position246)
					}
					if !_rules[rulepredicate_2]() {
						goto l245
					}
					{
						add(ruleAction37, position)
					}
					goto l244
				l245:
					position, tokenIndex, depth = position244, tokenIndex244, depth244
					if !_rules[rulepredicate_3]() {
						goto l242
					}
				}
			l244:
				depth--
				add(rulepredicate_2, position243)
			}
			return true
		l242:
			position, tokenIndex, depth = position242, tokenIndex242, depth242
			return false
		},
		/* 20 predicate_3 <- <((_ OP_NOT predicate_3 Action38) / (_ PAREN_OPEN predicate_1 _ PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position254, tokenIndex254, depth254 := position, tokenIndex, depth
			{
				position255 := position
				depth++
				{
					position256, tokenIndex256, depth256 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l257
					}
					{
						position258 := position
						depth++
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
								goto l257
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
								goto l257
							}
							position++
						}
					l261:
						{
							position263, tokenIndex263, depth263 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l264
							}
							position++
							goto l263
						l264:
							position, tokenIndex, depth = position263, tokenIndex263, depth263
							if buffer[position] != rune('T') {
								goto l257
							}
							position++
						}
					l263:
						if !_rules[ruleKEY]() {
							goto l257
						}
						depth--
						add(ruleOP_NOT, position258)
					}
					if !_rules[rulepredicate_3]() {
						goto l257
					}
					{
						add(ruleAction38, position)
					}
					goto l256
				l257:
					position, tokenIndex, depth = position256, tokenIndex256, depth256
					if !_rules[rule_]() {
						goto l266
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l266
					}
					if !_rules[rulepredicate_1]() {
						goto l266
					}
					if !_rules[rule_]() {
						goto l266
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l266
					}
					goto l256
				l266:
					position, tokenIndex, depth = position256, tokenIndex256, depth256
					{
						position267 := position
						depth++
						{
							position268, tokenIndex268, depth268 := position, tokenIndex, depth
							if !_rules[ruletagName]() {
								goto l269
							}
							if !_rules[rule_]() {
								goto l269
							}
							if buffer[position] != rune('=') {
								goto l269
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l269
							}
							{
								add(ruleAction39, position)
							}
							goto l268
						l269:
							position, tokenIndex, depth = position268, tokenIndex268, depth268
							if !_rules[ruletagName]() {
								goto l271
							}
							if !_rules[rule_]() {
								goto l271
							}
							if buffer[position] != rune('!') {
								goto l271
							}
							position++
							if buffer[position] != rune('=') {
								goto l271
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l271
							}
							{
								add(ruleAction40, position)
							}
							goto l268
						l271:
							position, tokenIndex, depth = position268, tokenIndex268, depth268
							if !_rules[ruletagName]() {
								goto l273
							}
							if !_rules[rule_]() {
								goto l273
							}
							{
								position274, tokenIndex274, depth274 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l275
								}
								position++
								goto l274
							l275:
								position, tokenIndex, depth = position274, tokenIndex274, depth274
								if buffer[position] != rune('M') {
									goto l273
								}
								position++
							}
						l274:
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
									goto l273
								}
								position++
							}
						l276:
							{
								position278, tokenIndex278, depth278 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l279
								}
								position++
								goto l278
							l279:
								position, tokenIndex, depth = position278, tokenIndex278, depth278
								if buffer[position] != rune('T') {
									goto l273
								}
								position++
							}
						l278:
							{
								position280, tokenIndex280, depth280 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l281
								}
								position++
								goto l280
							l281:
								position, tokenIndex, depth = position280, tokenIndex280, depth280
								if buffer[position] != rune('C') {
									goto l273
								}
								position++
							}
						l280:
							{
								position282, tokenIndex282, depth282 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l283
								}
								position++
								goto l282
							l283:
								position, tokenIndex, depth = position282, tokenIndex282, depth282
								if buffer[position] != rune('H') {
									goto l273
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
									goto l273
								}
								position++
							}
						l284:
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
									goto l273
								}
								position++
							}
						l286:
							if !_rules[ruleKEY]() {
								goto l273
							}
							if !_rules[ruleliteralString]() {
								goto l273
							}
							{
								add(ruleAction41, position)
							}
							goto l268
						l273:
							position, tokenIndex, depth = position268, tokenIndex268, depth268
							if !_rules[ruletagName]() {
								goto l254
							}
							if !_rules[rule_]() {
								goto l254
							}
							{
								position289, tokenIndex289, depth289 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l290
								}
								position++
								goto l289
							l290:
								position, tokenIndex, depth = position289, tokenIndex289, depth289
								if buffer[position] != rune('I') {
									goto l254
								}
								position++
							}
						l289:
							{
								position291, tokenIndex291, depth291 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l292
								}
								position++
								goto l291
							l292:
								position, tokenIndex, depth = position291, tokenIndex291, depth291
								if buffer[position] != rune('N') {
									goto l254
								}
								position++
							}
						l291:
							if !_rules[ruleKEY]() {
								goto l254
							}
							{
								position293 := position
								depth++
								{
									add(ruleAction44, position)
								}
								if !_rules[rule_]() {
									goto l254
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l254
								}
								if !_rules[ruleliteralListString]() {
									goto l254
								}
							l295:
								{
									position296, tokenIndex296, depth296 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l296
									}
									if !_rules[ruleCOMMA]() {
										goto l296
									}
									if !_rules[ruleliteralListString]() {
										goto l296
									}
									goto l295
								l296:
									position, tokenIndex, depth = position296, tokenIndex296, depth296
								}
								if !_rules[rule_]() {
									goto l254
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l254
								}
								depth--
								add(ruleliteralList, position293)
							}
							{
								add(ruleAction42, position)
							}
						}
					l268:
						depth--
						add(ruletagMatcher, position267)
					}
				}
			l256:
				depth--
				add(rulepredicate_3, position255)
			}
			return true
		l254:
			position, tokenIndex, depth = position254, tokenIndex254, depth254
			return false
		},
		/* 21 tagMatcher <- <((tagName _ '=' literalString Action39) / (tagName _ ('!' '=') literalString Action40) / (tagName _ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) KEY literalString Action41) / (tagName _ (('i' / 'I') ('n' / 'N')) KEY literalList Action42))> */
		nil,
		/* 22 literalString <- <(_ STRING Action43)> */
		func() bool {
			position299, tokenIndex299, depth299 := position, tokenIndex, depth
			{
				position300 := position
				depth++
				if !_rules[rule_]() {
					goto l299
				}
				if !_rules[ruleSTRING]() {
					goto l299
				}
				{
					add(ruleAction43, position)
				}
				depth--
				add(ruleliteralString, position300)
			}
			return true
		l299:
			position, tokenIndex, depth = position299, tokenIndex299, depth299
			return false
		},
		/* 23 literalList <- <(Action44 _ PAREN_OPEN literalListString (_ COMMA literalListString)* _ PAREN_CLOSE)> */
		nil,
		/* 24 literalListString <- <(_ STRING Action45)> */
		func() bool {
			position303, tokenIndex303, depth303 := position, tokenIndex, depth
			{
				position304 := position
				depth++
				if !_rules[rule_]() {
					goto l303
				}
				if !_rules[ruleSTRING]() {
					goto l303
				}
				{
					add(ruleAction45, position)
				}
				depth--
				add(ruleliteralListString, position304)
			}
			return true
		l303:
			position, tokenIndex, depth = position303, tokenIndex303, depth303
			return false
		},
		/* 25 tagName <- <(_ <TAG_NAME> Action46)> */
		func() bool {
			position306, tokenIndex306, depth306 := position, tokenIndex, depth
			{
				position307 := position
				depth++
				if !_rules[rule_]() {
					goto l306
				}
				{
					position308 := position
					depth++
					{
						position309 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l306
						}
						depth--
						add(ruleTAG_NAME, position309)
					}
					depth--
					add(rulePegText, position308)
				}
				{
					add(ruleAction46, position)
				}
				depth--
				add(ruletagName, position307)
			}
			return true
		l306:
			position, tokenIndex, depth = position306, tokenIndex306, depth306
			return false
		},
		/* 26 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position311, tokenIndex311, depth311 := position, tokenIndex, depth
			{
				position312 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l311
				}
				depth--
				add(ruleCOLUMN_NAME, position312)
			}
			return true
		l311:
			position, tokenIndex, depth = position311, tokenIndex311, depth311
			return false
		},
		/* 27 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 28 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 29 IDENTIFIER <- <(('`' CHAR* '`') / (_ !(KEYWORD KEY) ID_SEGMENT ('.' ID_SEGMENT)*))> */
		func() bool {
			position315, tokenIndex315, depth315 := position, tokenIndex, depth
			{
				position316 := position
				depth++
				{
					position317, tokenIndex317, depth317 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l318
					}
					position++
				l319:
					{
						position320, tokenIndex320, depth320 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l320
						}
						goto l319
					l320:
						position, tokenIndex, depth = position320, tokenIndex320, depth320
					}
					if buffer[position] != rune('`') {
						goto l318
					}
					position++
					goto l317
				l318:
					position, tokenIndex, depth = position317, tokenIndex317, depth317
					if !_rules[rule_]() {
						goto l315
					}
					{
						position321, tokenIndex321, depth321 := position, tokenIndex, depth
						{
							position322 := position
							depth++
							{
								position323, tokenIndex323, depth323 := position, tokenIndex, depth
								{
									position325, tokenIndex325, depth325 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l326
									}
									position++
									goto l325
								l326:
									position, tokenIndex, depth = position325, tokenIndex325, depth325
									if buffer[position] != rune('A') {
										goto l324
									}
									position++
								}
							l325:
								{
									position327, tokenIndex327, depth327 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l328
									}
									position++
									goto l327
								l328:
									position, tokenIndex, depth = position327, tokenIndex327, depth327
									if buffer[position] != rune('L') {
										goto l324
									}
									position++
								}
							l327:
								{
									position329, tokenIndex329, depth329 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l330
									}
									position++
									goto l329
								l330:
									position, tokenIndex, depth = position329, tokenIndex329, depth329
									if buffer[position] != rune('L') {
										goto l324
									}
									position++
								}
							l329:
								goto l323
							l324:
								position, tokenIndex, depth = position323, tokenIndex323, depth323
								{
									position332, tokenIndex332, depth332 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l333
									}
									position++
									goto l332
								l333:
									position, tokenIndex, depth = position332, tokenIndex332, depth332
									if buffer[position] != rune('A') {
										goto l331
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
										goto l331
									}
									position++
								}
							l334:
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
										goto l331
									}
									position++
								}
							l336:
								goto l323
							l331:
								position, tokenIndex, depth = position323, tokenIndex323, depth323
								{
									position339, tokenIndex339, depth339 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l340
									}
									position++
									goto l339
								l340:
									position, tokenIndex, depth = position339, tokenIndex339, depth339
									if buffer[position] != rune('M') {
										goto l338
									}
									position++
								}
							l339:
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
										goto l338
									}
									position++
								}
							l341:
								{
									position343, tokenIndex343, depth343 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l344
									}
									position++
									goto l343
								l344:
									position, tokenIndex, depth = position343, tokenIndex343, depth343
									if buffer[position] != rune('T') {
										goto l338
									}
									position++
								}
							l343:
								{
									position345, tokenIndex345, depth345 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l346
									}
									position++
									goto l345
								l346:
									position, tokenIndex, depth = position345, tokenIndex345, depth345
									if buffer[position] != rune('C') {
										goto l338
									}
									position++
								}
							l345:
								{
									position347, tokenIndex347, depth347 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l348
									}
									position++
									goto l347
								l348:
									position, tokenIndex, depth = position347, tokenIndex347, depth347
									if buffer[position] != rune('H') {
										goto l338
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
										goto l338
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
										goto l338
									}
									position++
								}
							l351:
								goto l323
							l338:
								position, tokenIndex, depth = position323, tokenIndex323, depth323
								{
									position354, tokenIndex354, depth354 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l355
									}
									position++
									goto l354
								l355:
									position, tokenIndex, depth = position354, tokenIndex354, depth354
									if buffer[position] != rune('S') {
										goto l353
									}
									position++
								}
							l354:
								{
									position356, tokenIndex356, depth356 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l357
									}
									position++
									goto l356
								l357:
									position, tokenIndex, depth = position356, tokenIndex356, depth356
									if buffer[position] != rune('E') {
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
										goto l353
									}
									position++
								}
							l360:
								{
									position362, tokenIndex362, depth362 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l363
									}
									position++
									goto l362
								l363:
									position, tokenIndex, depth = position362, tokenIndex362, depth362
									if buffer[position] != rune('C') {
										goto l353
									}
									position++
								}
							l362:
								{
									position364, tokenIndex364, depth364 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l365
									}
									position++
									goto l364
								l365:
									position, tokenIndex, depth = position364, tokenIndex364, depth364
									if buffer[position] != rune('T') {
										goto l353
									}
									position++
								}
							l364:
								goto l323
							l353:
								position, tokenIndex, depth = position323, tokenIndex323, depth323
								{
									switch buffer[position] {
									case 'M', 'm':
										{
											position367, tokenIndex367, depth367 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l368
											}
											position++
											goto l367
										l368:
											position, tokenIndex, depth = position367, tokenIndex367, depth367
											if buffer[position] != rune('M') {
												goto l321
											}
											position++
										}
									l367:
										{
											position369, tokenIndex369, depth369 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l370
											}
											position++
											goto l369
										l370:
											position, tokenIndex, depth = position369, tokenIndex369, depth369
											if buffer[position] != rune('E') {
												goto l321
											}
											position++
										}
									l369:
										{
											position371, tokenIndex371, depth371 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l372
											}
											position++
											goto l371
										l372:
											position, tokenIndex, depth = position371, tokenIndex371, depth371
											if buffer[position] != rune('T') {
												goto l321
											}
											position++
										}
									l371:
										{
											position373, tokenIndex373, depth373 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l374
											}
											position++
											goto l373
										l374:
											position, tokenIndex, depth = position373, tokenIndex373, depth373
											if buffer[position] != rune('R') {
												goto l321
											}
											position++
										}
									l373:
										{
											position375, tokenIndex375, depth375 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l376
											}
											position++
											goto l375
										l376:
											position, tokenIndex, depth = position375, tokenIndex375, depth375
											if buffer[position] != rune('I') {
												goto l321
											}
											position++
										}
									l375:
										{
											position377, tokenIndex377, depth377 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l378
											}
											position++
											goto l377
										l378:
											position, tokenIndex, depth = position377, tokenIndex377, depth377
											if buffer[position] != rune('C') {
												goto l321
											}
											position++
										}
									l377:
										{
											position379, tokenIndex379, depth379 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l380
											}
											position++
											goto l379
										l380:
											position, tokenIndex, depth = position379, tokenIndex379, depth379
											if buffer[position] != rune('S') {
												goto l321
											}
											position++
										}
									l379:
										break
									case 'W', 'w':
										{
											position381, tokenIndex381, depth381 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l382
											}
											position++
											goto l381
										l382:
											position, tokenIndex, depth = position381, tokenIndex381, depth381
											if buffer[position] != rune('W') {
												goto l321
											}
											position++
										}
									l381:
										{
											position383, tokenIndex383, depth383 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l384
											}
											position++
											goto l383
										l384:
											position, tokenIndex, depth = position383, tokenIndex383, depth383
											if buffer[position] != rune('H') {
												goto l321
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
												goto l321
											}
											position++
										}
									l385:
										{
											position387, tokenIndex387, depth387 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l388
											}
											position++
											goto l387
										l388:
											position, tokenIndex, depth = position387, tokenIndex387, depth387
											if buffer[position] != rune('R') {
												goto l321
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
												goto l321
											}
											position++
										}
									l389:
										break
									case 'O', 'o':
										{
											position391, tokenIndex391, depth391 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l392
											}
											position++
											goto l391
										l392:
											position, tokenIndex, depth = position391, tokenIndex391, depth391
											if buffer[position] != rune('O') {
												goto l321
											}
											position++
										}
									l391:
										{
											position393, tokenIndex393, depth393 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l394
											}
											position++
											goto l393
										l394:
											position, tokenIndex, depth = position393, tokenIndex393, depth393
											if buffer[position] != rune('R') {
												goto l321
											}
											position++
										}
									l393:
										break
									case 'N', 'n':
										{
											position395, tokenIndex395, depth395 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l396
											}
											position++
											goto l395
										l396:
											position, tokenIndex, depth = position395, tokenIndex395, depth395
											if buffer[position] != rune('N') {
												goto l321
											}
											position++
										}
									l395:
										{
											position397, tokenIndex397, depth397 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l398
											}
											position++
											goto l397
										l398:
											position, tokenIndex, depth = position397, tokenIndex397, depth397
											if buffer[position] != rune('O') {
												goto l321
											}
											position++
										}
									l397:
										{
											position399, tokenIndex399, depth399 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l400
											}
											position++
											goto l399
										l400:
											position, tokenIndex, depth = position399, tokenIndex399, depth399
											if buffer[position] != rune('T') {
												goto l321
											}
											position++
										}
									l399:
										break
									case 'I', 'i':
										{
											position401, tokenIndex401, depth401 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l402
											}
											position++
											goto l401
										l402:
											position, tokenIndex, depth = position401, tokenIndex401, depth401
											if buffer[position] != rune('I') {
												goto l321
											}
											position++
										}
									l401:
										{
											position403, tokenIndex403, depth403 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l404
											}
											position++
											goto l403
										l404:
											position, tokenIndex, depth = position403, tokenIndex403, depth403
											if buffer[position] != rune('N') {
												goto l321
											}
											position++
										}
									l403:
										break
									case 'G', 'g':
										{
											position405, tokenIndex405, depth405 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l406
											}
											position++
											goto l405
										l406:
											position, tokenIndex, depth = position405, tokenIndex405, depth405
											if buffer[position] != rune('G') {
												goto l321
											}
											position++
										}
									l405:
										{
											position407, tokenIndex407, depth407 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l408
											}
											position++
											goto l407
										l408:
											position, tokenIndex, depth = position407, tokenIndex407, depth407
											if buffer[position] != rune('R') {
												goto l321
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
												goto l321
											}
											position++
										}
									l409:
										{
											position411, tokenIndex411, depth411 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l412
											}
											position++
											goto l411
										l412:
											position, tokenIndex, depth = position411, tokenIndex411, depth411
											if buffer[position] != rune('U') {
												goto l321
											}
											position++
										}
									l411:
										{
											position413, tokenIndex413, depth413 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l414
											}
											position++
											goto l413
										l414:
											position, tokenIndex, depth = position413, tokenIndex413, depth413
											if buffer[position] != rune('P') {
												goto l321
											}
											position++
										}
									l413:
										break
									case 'D', 'd':
										{
											position415, tokenIndex415, depth415 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l416
											}
											position++
											goto l415
										l416:
											position, tokenIndex, depth = position415, tokenIndex415, depth415
											if buffer[position] != rune('D') {
												goto l321
											}
											position++
										}
									l415:
										{
											position417, tokenIndex417, depth417 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l418
											}
											position++
											goto l417
										l418:
											position, tokenIndex, depth = position417, tokenIndex417, depth417
											if buffer[position] != rune('E') {
												goto l321
											}
											position++
										}
									l417:
										{
											position419, tokenIndex419, depth419 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l420
											}
											position++
											goto l419
										l420:
											position, tokenIndex, depth = position419, tokenIndex419, depth419
											if buffer[position] != rune('S') {
												goto l321
											}
											position++
										}
									l419:
										{
											position421, tokenIndex421, depth421 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l422
											}
											position++
											goto l421
										l422:
											position, tokenIndex, depth = position421, tokenIndex421, depth421
											if buffer[position] != rune('C') {
												goto l321
											}
											position++
										}
									l421:
										{
											position423, tokenIndex423, depth423 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l424
											}
											position++
											goto l423
										l424:
											position, tokenIndex, depth = position423, tokenIndex423, depth423
											if buffer[position] != rune('R') {
												goto l321
											}
											position++
										}
									l423:
										{
											position425, tokenIndex425, depth425 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l426
											}
											position++
											goto l425
										l426:
											position, tokenIndex, depth = position425, tokenIndex425, depth425
											if buffer[position] != rune('I') {
												goto l321
											}
											position++
										}
									l425:
										{
											position427, tokenIndex427, depth427 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l428
											}
											position++
											goto l427
										l428:
											position, tokenIndex, depth = position427, tokenIndex427, depth427
											if buffer[position] != rune('B') {
												goto l321
											}
											position++
										}
									l427:
										{
											position429, tokenIndex429, depth429 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l430
											}
											position++
											goto l429
										l430:
											position, tokenIndex, depth = position429, tokenIndex429, depth429
											if buffer[position] != rune('E') {
												goto l321
											}
											position++
										}
									l429:
										break
									case 'B', 'b':
										{
											position431, tokenIndex431, depth431 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l432
											}
											position++
											goto l431
										l432:
											position, tokenIndex, depth = position431, tokenIndex431, depth431
											if buffer[position] != rune('B') {
												goto l321
											}
											position++
										}
									l431:
										{
											position433, tokenIndex433, depth433 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l434
											}
											position++
											goto l433
										l434:
											position, tokenIndex, depth = position433, tokenIndex433, depth433
											if buffer[position] != rune('Y') {
												goto l321
											}
											position++
										}
									l433:
										break
									case 'A', 'a':
										{
											position435, tokenIndex435, depth435 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l436
											}
											position++
											goto l435
										l436:
											position, tokenIndex, depth = position435, tokenIndex435, depth435
											if buffer[position] != rune('A') {
												goto l321
											}
											position++
										}
									l435:
										{
											position437, tokenIndex437, depth437 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l438
											}
											position++
											goto l437
										l438:
											position, tokenIndex, depth = position437, tokenIndex437, depth437
											if buffer[position] != rune('S') {
												goto l321
											}
											position++
										}
									l437:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l321
										}
										break
									}
								}

							}
						l323:
							depth--
							add(ruleKEYWORD, position322)
						}
						if !_rules[ruleKEY]() {
							goto l321
						}
						goto l315
					l321:
						position, tokenIndex, depth = position321, tokenIndex321, depth321
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l315
					}
				l439:
					{
						position440, tokenIndex440, depth440 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l440
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l440
						}
						goto l439
					l440:
						position, tokenIndex, depth = position440, tokenIndex440, depth440
					}
				}
			l317:
				depth--
				add(ruleIDENTIFIER, position316)
			}
			return true
		l315:
			position, tokenIndex, depth = position315, tokenIndex315, depth315
			return false
		},
		/* 30 TIMESTAMP <- <((_ <(NUMBER ([a-z] / [A-Z])*)>) / (_ STRING) / (_ <(('n' / 'N') ('o' / 'O') ('w' / 'W'))>))> */
		nil,
		/* 31 ID_SEGMENT <- <(_ ID_START ID_CONT*)> */
		func() bool {
			position442, tokenIndex442, depth442 := position, tokenIndex, depth
			{
				position443 := position
				depth++
				if !_rules[rule_]() {
					goto l442
				}
				if !_rules[ruleID_START]() {
					goto l442
				}
			l444:
				{
					position445, tokenIndex445, depth445 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l445
					}
					goto l444
				l445:
					position, tokenIndex, depth = position445, tokenIndex445, depth445
				}
				depth--
				add(ruleID_SEGMENT, position443)
			}
			return true
		l442:
			position, tokenIndex, depth = position442, tokenIndex442, depth442
			return false
		},
		/* 32 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position446, tokenIndex446, depth446 := position, tokenIndex, depth
			{
				position447 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l446
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l446
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l446
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position447)
			}
			return true
		l446:
			position, tokenIndex, depth = position446, tokenIndex446, depth446
			return false
		},
		/* 33 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position449, tokenIndex449, depth449 := position, tokenIndex, depth
			{
				position450 := position
				depth++
				{
					position451, tokenIndex451, depth451 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l452
					}
					goto l451
				l452:
					position, tokenIndex, depth = position451, tokenIndex451, depth451
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l449
					}
					position++
				}
			l451:
				depth--
				add(ruleID_CONT, position450)
			}
			return true
		l449:
			position, tokenIndex, depth = position449, tokenIndex449, depth449
			return false
		},
		/* 34 PROPERTY_KEY <- <(((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> KEY _ (('b' / 'B') ('y' / 'Y')))) | (&('R' | 'r') <(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))>) | (&('T' | 't') <(('t' / 'T') ('o' / 'O'))>) | (&('F' | 'f') <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))>)) KEY)> */
		func() bool {
			position453, tokenIndex453, depth453 := position, tokenIndex, depth
			{
				position454 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position456 := position
							depth++
							{
								position457, tokenIndex457, depth457 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l458
								}
								position++
								goto l457
							l458:
								position, tokenIndex, depth = position457, tokenIndex457, depth457
								if buffer[position] != rune('S') {
									goto l453
								}
								position++
							}
						l457:
							{
								position459, tokenIndex459, depth459 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l460
								}
								position++
								goto l459
							l460:
								position, tokenIndex, depth = position459, tokenIndex459, depth459
								if buffer[position] != rune('A') {
									goto l453
								}
								position++
							}
						l459:
							{
								position461, tokenIndex461, depth461 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l462
								}
								position++
								goto l461
							l462:
								position, tokenIndex, depth = position461, tokenIndex461, depth461
								if buffer[position] != rune('M') {
									goto l453
								}
								position++
							}
						l461:
							{
								position463, tokenIndex463, depth463 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l464
								}
								position++
								goto l463
							l464:
								position, tokenIndex, depth = position463, tokenIndex463, depth463
								if buffer[position] != rune('P') {
									goto l453
								}
								position++
							}
						l463:
							{
								position465, tokenIndex465, depth465 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l466
								}
								position++
								goto l465
							l466:
								position, tokenIndex, depth = position465, tokenIndex465, depth465
								if buffer[position] != rune('L') {
									goto l453
								}
								position++
							}
						l465:
							{
								position467, tokenIndex467, depth467 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l468
								}
								position++
								goto l467
							l468:
								position, tokenIndex, depth = position467, tokenIndex467, depth467
								if buffer[position] != rune('E') {
									goto l453
								}
								position++
							}
						l467:
							depth--
							add(rulePegText, position456)
						}
						if !_rules[ruleKEY]() {
							goto l453
						}
						if !_rules[rule_]() {
							goto l453
						}
						{
							position469, tokenIndex469, depth469 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l470
							}
							position++
							goto l469
						l470:
							position, tokenIndex, depth = position469, tokenIndex469, depth469
							if buffer[position] != rune('B') {
								goto l453
							}
							position++
						}
					l469:
						{
							position471, tokenIndex471, depth471 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l472
							}
							position++
							goto l471
						l472:
							position, tokenIndex, depth = position471, tokenIndex471, depth471
							if buffer[position] != rune('Y') {
								goto l453
							}
							position++
						}
					l471:
						break
					case 'R', 'r':
						{
							position473 := position
							depth++
							{
								position474, tokenIndex474, depth474 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l475
								}
								position++
								goto l474
							l475:
								position, tokenIndex, depth = position474, tokenIndex474, depth474
								if buffer[position] != rune('R') {
									goto l453
								}
								position++
							}
						l474:
							{
								position476, tokenIndex476, depth476 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l477
								}
								position++
								goto l476
							l477:
								position, tokenIndex, depth = position476, tokenIndex476, depth476
								if buffer[position] != rune('E') {
									goto l453
								}
								position++
							}
						l476:
							{
								position478, tokenIndex478, depth478 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l479
								}
								position++
								goto l478
							l479:
								position, tokenIndex, depth = position478, tokenIndex478, depth478
								if buffer[position] != rune('S') {
									goto l453
								}
								position++
							}
						l478:
							{
								position480, tokenIndex480, depth480 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l481
								}
								position++
								goto l480
							l481:
								position, tokenIndex, depth = position480, tokenIndex480, depth480
								if buffer[position] != rune('O') {
									goto l453
								}
								position++
							}
						l480:
							{
								position482, tokenIndex482, depth482 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l483
								}
								position++
								goto l482
							l483:
								position, tokenIndex, depth = position482, tokenIndex482, depth482
								if buffer[position] != rune('L') {
									goto l453
								}
								position++
							}
						l482:
							{
								position484, tokenIndex484, depth484 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l485
								}
								position++
								goto l484
							l485:
								position, tokenIndex, depth = position484, tokenIndex484, depth484
								if buffer[position] != rune('U') {
									goto l453
								}
								position++
							}
						l484:
							{
								position486, tokenIndex486, depth486 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l487
								}
								position++
								goto l486
							l487:
								position, tokenIndex, depth = position486, tokenIndex486, depth486
								if buffer[position] != rune('T') {
									goto l453
								}
								position++
							}
						l486:
							{
								position488, tokenIndex488, depth488 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l489
								}
								position++
								goto l488
							l489:
								position, tokenIndex, depth = position488, tokenIndex488, depth488
								if buffer[position] != rune('I') {
									goto l453
								}
								position++
							}
						l488:
							{
								position490, tokenIndex490, depth490 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l491
								}
								position++
								goto l490
							l491:
								position, tokenIndex, depth = position490, tokenIndex490, depth490
								if buffer[position] != rune('O') {
									goto l453
								}
								position++
							}
						l490:
							{
								position492, tokenIndex492, depth492 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l493
								}
								position++
								goto l492
							l493:
								position, tokenIndex, depth = position492, tokenIndex492, depth492
								if buffer[position] != rune('N') {
									goto l453
								}
								position++
							}
						l492:
							depth--
							add(rulePegText, position473)
						}
						break
					case 'T', 't':
						{
							position494 := position
							depth++
							{
								position495, tokenIndex495, depth495 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l496
								}
								position++
								goto l495
							l496:
								position, tokenIndex, depth = position495, tokenIndex495, depth495
								if buffer[position] != rune('T') {
									goto l453
								}
								position++
							}
						l495:
							{
								position497, tokenIndex497, depth497 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l498
								}
								position++
								goto l497
							l498:
								position, tokenIndex, depth = position497, tokenIndex497, depth497
								if buffer[position] != rune('O') {
									goto l453
								}
								position++
							}
						l497:
							depth--
							add(rulePegText, position494)
						}
						break
					default:
						{
							position499 := position
							depth++
							{
								position500, tokenIndex500, depth500 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l501
								}
								position++
								goto l500
							l501:
								position, tokenIndex, depth = position500, tokenIndex500, depth500
								if buffer[position] != rune('F') {
									goto l453
								}
								position++
							}
						l500:
							{
								position502, tokenIndex502, depth502 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l503
								}
								position++
								goto l502
							l503:
								position, tokenIndex, depth = position502, tokenIndex502, depth502
								if buffer[position] != rune('R') {
									goto l453
								}
								position++
							}
						l502:
							{
								position504, tokenIndex504, depth504 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l505
								}
								position++
								goto l504
							l505:
								position, tokenIndex, depth = position504, tokenIndex504, depth504
								if buffer[position] != rune('O') {
									goto l453
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
									goto l453
								}
								position++
							}
						l506:
							depth--
							add(rulePegText, position499)
						}
						break
					}
				}

				if !_rules[ruleKEY]() {
					goto l453
				}
				depth--
				add(rulePROPERTY_KEY, position454)
			}
			return true
		l453:
			position, tokenIndex, depth = position453, tokenIndex453, depth453
			return false
		},
		/* 35 PROPERTY_VALUE <- <TIMESTAMP> */
		nil,
		/* 36 KEYWORD <- <((('a' / 'A') ('l' / 'L') ('l' / 'L')) / (('a' / 'A') ('n' / 'N') ('d' / 'D')) / (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) / (('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T')) / ((&('M' | 'm') (('m' / 'M') ('e' / 'E') ('t' / 'T') ('r' / 'R') ('i' / 'I') ('c' / 'C') ('s' / 'S'))) | (&('W' | 'w') (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E'))) | (&('O' | 'o') (('o' / 'O') ('r' / 'R'))) | (&('N' | 'n') (('n' / 'N') ('o' / 'O') ('t' / 'T'))) | (&('I' | 'i') (('i' / 'I') ('n' / 'N'))) | (&('G' | 'g') (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P'))) | (&('D' | 'd') (('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E'))) | (&('B' | 'b') (('b' / 'B') ('y' / 'Y'))) | (&('A' | 'a') (('a' / 'A') ('s' / 'S'))) | (&('F' | 'R' | 'S' | 'T' | 'f' | 'r' | 's' | 't') PROPERTY_KEY)))> */
		nil,
		/* 37 OP_PIPE <- <'|'> */
		nil,
		/* 38 OP_ADD <- <'+'> */
		nil,
		/* 39 OP_SUB <- <'-'> */
		nil,
		/* 40 OP_MULT <- <'*'> */
		nil,
		/* 41 OP_DIV <- <'/'> */
		nil,
		/* 42 OP_AND <- <(('a' / 'A') ('n' / 'N') ('d' / 'D') KEY)> */
		nil,
		/* 43 OP_OR <- <(('o' / 'O') ('r' / 'R') KEY)> */
		nil,
		/* 44 OP_NOT <- <(('n' / 'N') ('o' / 'O') ('t' / 'T') KEY)> */
		nil,
		/* 45 QUOTE_SINGLE <- <'\''> */
		func() bool {
			position518, tokenIndex518, depth518 := position, tokenIndex, depth
			{
				position519 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l518
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position519)
			}
			return true
		l518:
			position, tokenIndex, depth = position518, tokenIndex518, depth518
			return false
		},
		/* 46 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position520, tokenIndex520, depth520 := position, tokenIndex, depth
			{
				position521 := position
				depth++
				if buffer[position] != rune('"') {
					goto l520
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position521)
			}
			return true
		l520:
			position, tokenIndex, depth = position520, tokenIndex520, depth520
			return false
		},
		/* 47 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> QUOTE_SINGLE) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> QUOTE_DOUBLE))> */
		func() bool {
			position522, tokenIndex522, depth522 := position, tokenIndex, depth
			{
				position523 := position
				depth++
				{
					position524, tokenIndex524, depth524 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l525
					}
					{
						position526 := position
						depth++
					l527:
						{
							position528, tokenIndex528, depth528 := position, tokenIndex, depth
							{
								position529, tokenIndex529, depth529 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l529
								}
								goto l528
							l529:
								position, tokenIndex, depth = position529, tokenIndex529, depth529
							}
							if !_rules[ruleCHAR]() {
								goto l528
							}
							goto l527
						l528:
							position, tokenIndex, depth = position528, tokenIndex528, depth528
						}
						depth--
						add(rulePegText, position526)
					}
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l525
					}
					goto l524
				l525:
					position, tokenIndex, depth = position524, tokenIndex524, depth524
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l522
					}
					{
						position530 := position
						depth++
					l531:
						{
							position532, tokenIndex532, depth532 := position, tokenIndex, depth
							{
								position533, tokenIndex533, depth533 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l533
								}
								goto l532
							l533:
								position, tokenIndex, depth = position533, tokenIndex533, depth533
							}
							if !_rules[ruleCHAR]() {
								goto l532
							}
							goto l531
						l532:
							position, tokenIndex, depth = position532, tokenIndex532, depth532
						}
						depth--
						add(rulePegText, position530)
					}
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l522
					}
				}
			l524:
				depth--
				add(ruleSTRING, position523)
			}
			return true
		l522:
			position, tokenIndex, depth = position522, tokenIndex522, depth522
			return false
		},
		/* 48 CHAR <- <(('\\' ((&('"') QUOTE_DOUBLE) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position534, tokenIndex534, depth534 := position, tokenIndex, depth
			{
				position535 := position
				depth++
				{
					position536, tokenIndex536, depth536 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l537
					}
					position++
					{
						switch buffer[position] {
						case '"':
							if !_rules[ruleQUOTE_DOUBLE]() {
								goto l537
							}
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l537
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l537
							}
							break
						}
					}

					goto l536
				l537:
					position, tokenIndex, depth = position536, tokenIndex536, depth536
					{
						position539, tokenIndex539, depth539 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l539
						}
						goto l534
					l539:
						position, tokenIndex, depth = position539, tokenIndex539, depth539
					}
					if !matchDot() {
						goto l534
					}
				}
			l536:
				depth--
				add(ruleCHAR, position535)
			}
			return true
		l534:
			position, tokenIndex, depth = position534, tokenIndex534, depth534
			return false
		},
		/* 49 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position540, tokenIndex540, depth540 := position, tokenIndex, depth
			{
				position541 := position
				depth++
				{
					position542, tokenIndex542, depth542 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l543
					}
					position++
					goto l542
				l543:
					position, tokenIndex, depth = position542, tokenIndex542, depth542
					if buffer[position] != rune('\\') {
						goto l540
					}
					position++
				}
			l542:
				depth--
				add(ruleESCAPE_CLASS, position541)
			}
			return true
		l540:
			position, tokenIndex, depth = position540, tokenIndex540, depth540
			return false
		},
		/* 50 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position544, tokenIndex544, depth544 := position, tokenIndex, depth
			{
				position545 := position
				depth++
				{
					position546 := position
					depth++
					{
						position547, tokenIndex547, depth547 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l547
						}
						position++
						goto l548
					l547:
						position, tokenIndex, depth = position547, tokenIndex547, depth547
					}
				l548:
					{
						position549 := position
						depth++
						{
							position550, tokenIndex550, depth550 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l551
							}
							position++
							goto l550
						l551:
							position, tokenIndex, depth = position550, tokenIndex550, depth550
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l544
							}
							position++
						l552:
							{
								position553, tokenIndex553, depth553 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l553
								}
								position++
								goto l552
							l553:
								position, tokenIndex, depth = position553, tokenIndex553, depth553
							}
						}
					l550:
						depth--
						add(ruleNUMBER_NATURAL, position549)
					}
					depth--
					add(ruleNUMBER_INTEGER, position546)
				}
				{
					position554, tokenIndex554, depth554 := position, tokenIndex, depth
					{
						position556 := position
						depth++
						if buffer[position] != rune('.') {
							goto l554
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l554
						}
						position++
					l557:
						{
							position558, tokenIndex558, depth558 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l558
							}
							position++
							goto l557
						l558:
							position, tokenIndex, depth = position558, tokenIndex558, depth558
						}
						depth--
						add(ruleNUMBER_FRACTION, position556)
					}
					goto l555
				l554:
					position, tokenIndex, depth = position554, tokenIndex554, depth554
				}
			l555:
				{
					position559, tokenIndex559, depth559 := position, tokenIndex, depth
					{
						position561 := position
						depth++
						{
							position562, tokenIndex562, depth562 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l563
							}
							position++
							goto l562
						l563:
							position, tokenIndex, depth = position562, tokenIndex562, depth562
							if buffer[position] != rune('E') {
								goto l559
							}
							position++
						}
					l562:
						{
							position564, tokenIndex564, depth564 := position, tokenIndex, depth
							{
								position566, tokenIndex566, depth566 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l567
								}
								position++
								goto l566
							l567:
								position, tokenIndex, depth = position566, tokenIndex566, depth566
								if buffer[position] != rune('-') {
									goto l564
								}
								position++
							}
						l566:
							goto l565
						l564:
							position, tokenIndex, depth = position564, tokenIndex564, depth564
						}
					l565:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l559
						}
						position++
					l568:
						{
							position569, tokenIndex569, depth569 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l569
							}
							position++
							goto l568
						l569:
							position, tokenIndex, depth = position569, tokenIndex569, depth569
						}
						depth--
						add(ruleNUMBER_EXP, position561)
					}
					goto l560
				l559:
					position, tokenIndex, depth = position559, tokenIndex559, depth559
				}
			l560:
				depth--
				add(ruleNUMBER, position545)
			}
			return true
		l544:
			position, tokenIndex, depth = position544, tokenIndex544, depth544
			return false
		},
		/* 51 NUMBER_NATURAL <- <('0' / ([1-9] [0-9]*))> */
		nil,
		/* 52 NUMBER_FRACTION <- <('.' [0-9]+)> */
		nil,
		/* 53 NUMBER_INTEGER <- <('-'? NUMBER_NATURAL)> */
		nil,
		/* 54 NUMBER_EXP <- <(('e' / 'E') ('+' / '-')? [0-9]+)> */
		nil,
		/* 55 DURATION <- <(NUMBER [a-z]+ KEY)> */
		nil,
		/* 56 PAREN_OPEN <- <'('> */
		func() bool {
			position575, tokenIndex575, depth575 := position, tokenIndex, depth
			{
				position576 := position
				depth++
				if buffer[position] != rune('(') {
					goto l575
				}
				position++
				depth--
				add(rulePAREN_OPEN, position576)
			}
			return true
		l575:
			position, tokenIndex, depth = position575, tokenIndex575, depth575
			return false
		},
		/* 57 PAREN_CLOSE <- <')'> */
		func() bool {
			position577, tokenIndex577, depth577 := position, tokenIndex, depth
			{
				position578 := position
				depth++
				if buffer[position] != rune(')') {
					goto l577
				}
				position++
				depth--
				add(rulePAREN_CLOSE, position578)
			}
			return true
		l577:
			position, tokenIndex, depth = position577, tokenIndex577, depth577
			return false
		},
		/* 58 COMMA <- <','> */
		func() bool {
			position579, tokenIndex579, depth579 := position, tokenIndex, depth
			{
				position580 := position
				depth++
				if buffer[position] != rune(',') {
					goto l579
				}
				position++
				depth--
				add(ruleCOMMA, position580)
			}
			return true
		l579:
			position, tokenIndex, depth = position579, tokenIndex579, depth579
			return false
		},
		/* 59 _ <- <SPACE*> */
		func() bool {
			{
				position582 := position
				depth++
			l583:
				{
					position584, tokenIndex584, depth584 := position, tokenIndex, depth
					{
						position585 := position
						depth++
						{
							switch buffer[position] {
							case '\t':
								if buffer[position] != rune('\t') {
									goto l584
								}
								position++
								break
							case '\n':
								if buffer[position] != rune('\n') {
									goto l584
								}
								position++
								break
							default:
								if buffer[position] != rune(' ') {
									goto l584
								}
								position++
								break
							}
						}

						depth--
						add(ruleSPACE, position585)
					}
					goto l583
				l584:
					position, tokenIndex, depth = position584, tokenIndex584, depth584
				}
				depth--
				add(rule_, position582)
			}
			return true
		},
		/* 60 KEY <- <!ID_CONT> */
		func() bool {
			position587, tokenIndex587, depth587 := position, tokenIndex, depth
			{
				position588 := position
				depth++
				{
					position589, tokenIndex589, depth589 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l589
					}
					goto l587
				l589:
					position, tokenIndex, depth = position589, tokenIndex589, depth589
				}
				depth--
				add(ruleKEY, position588)
			}
			return true
		l587:
			position, tokenIndex, depth = position587, tokenIndex587, depth587
			return false
		},
		/* 61 SPACE <- <((&('\t') '\t') | (&('\n') '\n') | (&(' ') ' '))> */
		nil,
		/* 63 Action0 <- <{
		   p.makeSelect()
		 }> */
		nil,
		/* 64 Action1 <- <{ p.makeDescribeAll() }> */
		nil,
		/* 65 Action2 <- <{ p.makeDescribeMetrics() }> */
		nil,
		nil,
		/* 67 Action3 <- <{ p.addStringLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 68 Action4 <- <{ p.makeDescribe() }> */
		nil,
		/* 69 Action5 <- <{ p.addEvaluationContext() }> */
		nil,
		/* 70 Action6 <- <{ p.addPropertyKey(buffer[begin:end])   }> */
		nil,
		/* 71 Action7 <- <{ p.addPropertyValue(buffer[begin:end]) }> */
		nil,
		/* 72 Action8 <- <{ p.insertPropertyKeyValue() }> */
		nil,
		/* 73 Action9 <- <{ p.checkPropertyClause() }> */
		nil,
		/* 74 Action10 <- <{ p.addNullPredicate() }> */
		nil,
		/* 75 Action11 <- <{ p.addExpressionList() }> */
		nil,
		/* 76 Action12 <- <{ p.appendExpression() }> */
		nil,
		/* 77 Action13 <- <{ p.appendExpression() }> */
		nil,
		/* 78 Action14 <- <{ p.addOperatorLiteral("+") }> */
		nil,
		/* 79 Action15 <- <{ p.addOperatorLiteral("-") }> */
		nil,
		/* 80 Action16 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 81 Action17 <- <{ p.addOperatorLiteral("/") }> */
		nil,
		/* 82 Action18 <- <{ p.addOperatorLiteral("*") }> */
		nil,
		/* 83 Action19 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 84 Action20 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 85 Action21 <- <{p.addExpressionList()}> */
		nil,
		/* 86 Action22 <- <{ p.addGroupBy() }> */
		nil,
		/* 87 Action23 <- <{
		   p.addExpressionList()
		   p.addGroupBy()
		 }> */
		nil,
		/* 88 Action24 <- <{
		   p.addPipeExpression()
		 }> */
		nil,
		/* 89 Action25 <- <{ p.addDurationNode(text) }> */
		nil,
		/* 90 Action26 <- <{ p.addNumberNode(buffer[begin:end]) }> */
		nil,
		/* 91 Action27 <- <{ p.addStringNode(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 92 Action28 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 93 Action29 <- <{ p.addGroupBy() }> */
		nil,
		/* 94 Action30 <- <{
		   p.addFunctionInvocation()
		 }> */
		nil,
		/* 95 Action31 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 96 Action32 <- <{ p.addNullPredicate() }> */
		nil,
		/* 97 Action33 <- <{
		   p.addMetricExpression()
		 }> */
		nil,
		/* 98 Action34 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 99 Action35 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		   }> */
		nil,
		/* 100 Action36 <- <{ p.addOrPredicate() }> */
		nil,
		/* 101 Action37 <- <{ p.addAndPredicate() }> */
		nil,
		/* 102 Action38 <- <{ p.addNotPredicate() }> */
		nil,
		/* 103 Action39 <- <{
		   p.addLiteralMatcher()
		 }> */
		nil,
		/* 104 Action40 <- <{
		   p.addLiteralMatcher()
		   p.addNotPredicate()
		 }> */
		nil,
		/* 105 Action41 <- <{
		   p.addRegexMatcher()
		 }> */
		nil,
		/* 106 Action42 <- <{
		   p.addListMatcher()
		 }> */
		nil,
		/* 107 Action43 <- <{
		  p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 108 Action44 <- <{ p.addLiteralList() }> */
		nil,
		/* 109 Action45 <- <{
		  p.appendLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 110 Action46 <- <{ p.addTagLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
	}
	p.rules = _rules
}
