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
	rules  [109]func() bool
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
			p.addNumberNode(buffer[begin:end])
		case ruleAction26:
			p.addStringNode(unescapeLiteral(buffer[begin:end]))
		case ruleAction27:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction28:
			p.addGroupBy()
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
			p.addOrPredicate()
		case ruleAction36:
			p.addAndPredicate()
		case ruleAction37:
			p.addNotPredicate()
		case ruleAction38:

			p.addLiteralMatcher()

		case ruleAction39:

			p.addLiteralMatcher()
			p.addNotPredicate()

		case ruleAction40:

			p.addRegexMatcher()

		case ruleAction41:

			p.addListMatcher()

		case ruleAction42:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction43:
			p.addLiteralList()
		case ruleAction44:

			p.appendLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction45:
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
		/* 12 add_pipe <- <(_ OP_PIPE _ <IDENTIFIER> Action20 ((_ PAREN_OPEN Action21 Action22 groupByClause? _ PAREN_CLOSE) / Action23) Action24)*> */
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
							add(ruleAction21, position)
						}
						{
							add(ruleAction22, position)
						}
						{
							position169, tokenIndex169, depth169 := position, tokenIndex, depth
							if !_rules[rulegroupByClause]() {
								goto l169
							}
							goto l170
						l169:
							position, tokenIndex, depth = position169, tokenIndex169, depth169
						}
					l170:
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
		/* 13 expression_atom <- <(expression_function / expression_metric / (_ PAREN_OPEN expression_start _ PAREN_CLOSE) / (_ <NUMBER> Action25) / (_ STRING Action26))> */
		func() bool {
			position173, tokenIndex173, depth173 := position, tokenIndex, depth
			{
				position174 := position
				depth++
				{
					position175, tokenIndex175, depth175 := position, tokenIndex, depth
					{
						position177 := position
						depth++
						if !_rules[rule_]() {
							goto l176
						}
						{
							position178 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l176
							}
							depth--
							add(rulePegText, position178)
						}
						{
							add(ruleAction27, position)
						}
						if !_rules[rule_]() {
							goto l176
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l176
						}
						if !_rules[ruleexpressionList]() {
							goto l176
						}
						{
							add(ruleAction28, position)
						}
						{
							position181, tokenIndex181, depth181 := position, tokenIndex, depth
							if !_rules[rulegroupByClause]() {
								goto l181
							}
							goto l182
						l181:
							position, tokenIndex, depth = position181, tokenIndex181, depth181
						}
					l182:
						if !_rules[rule_]() {
							goto l176
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l176
						}
						{
							add(ruleAction29, position)
						}
						depth--
						add(ruleexpression_function, position177)
					}
					goto l175
				l176:
					position, tokenIndex, depth = position175, tokenIndex175, depth175
					{
						position185 := position
						depth++
						if !_rules[rule_]() {
							goto l184
						}
						{
							position186 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l184
							}
							depth--
							add(rulePegText, position186)
						}
						{
							add(ruleAction30, position)
						}
						{
							position188, tokenIndex188, depth188 := position, tokenIndex, depth
							{
								position190, tokenIndex190, depth190 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l191
								}
								if buffer[position] != rune('[') {
									goto l191
								}
								position++
								if !_rules[rulepredicate_1]() {
									goto l191
								}
								if !_rules[rule_]() {
									goto l191
								}
								if buffer[position] != rune(']') {
									goto l191
								}
								position++
								goto l190
							l191:
								position, tokenIndex, depth = position190, tokenIndex190, depth190
								{
									add(ruleAction31, position)
								}
							}
						l190:
							goto l189

							position, tokenIndex, depth = position188, tokenIndex188, depth188
						}
					l189:
						{
							add(ruleAction32, position)
						}
						depth--
						add(ruleexpression_metric, position185)
					}
					goto l175
				l184:
					position, tokenIndex, depth = position175, tokenIndex175, depth175
					if !_rules[rule_]() {
						goto l194
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l194
					}
					if !_rules[ruleexpression_start]() {
						goto l194
					}
					if !_rules[rule_]() {
						goto l194
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l194
					}
					goto l175
				l194:
					position, tokenIndex, depth = position175, tokenIndex175, depth175
					if !_rules[rule_]() {
						goto l195
					}
					{
						position196 := position
						depth++
						if !_rules[ruleNUMBER]() {
							goto l195
						}
						depth--
						add(rulePegText, position196)
					}
					{
						add(ruleAction25, position)
					}
					goto l175
				l195:
					position, tokenIndex, depth = position175, tokenIndex175, depth175
					if !_rules[rule_]() {
						goto l173
					}
					if !_rules[ruleSTRING]() {
						goto l173
					}
					{
						add(ruleAction26, position)
					}
				}
			l175:
				depth--
				add(ruleexpression_atom, position174)
			}
			return true
		l173:
			position, tokenIndex, depth = position173, tokenIndex173, depth173
			return false
		},
		/* 14 expression_function <- <(_ <IDENTIFIER> Action27 _ PAREN_OPEN expressionList Action28 groupByClause? _ PAREN_CLOSE Action29)> */
		nil,
		/* 15 expression_metric <- <(_ <IDENTIFIER> Action30 ((_ '[' predicate_1 _ ']') / Action31)? Action32)> */
		nil,
		/* 16 groupByClause <- <(_ (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P')) KEY _ (('b' / 'B') ('y' / 'Y')) KEY _ <COLUMN_NAME> Action33 (_ COMMA _ <COLUMN_NAME> Action34)*)> */
		func() bool {
			position201, tokenIndex201, depth201 := position, tokenIndex, depth
			{
				position202 := position
				depth++
				if !_rules[rule_]() {
					goto l201
				}
				{
					position203, tokenIndex203, depth203 := position, tokenIndex, depth
					if buffer[position] != rune('g') {
						goto l204
					}
					position++
					goto l203
				l204:
					position, tokenIndex, depth = position203, tokenIndex203, depth203
					if buffer[position] != rune('G') {
						goto l201
					}
					position++
				}
			l203:
				{
					position205, tokenIndex205, depth205 := position, tokenIndex, depth
					if buffer[position] != rune('r') {
						goto l206
					}
					position++
					goto l205
				l206:
					position, tokenIndex, depth = position205, tokenIndex205, depth205
					if buffer[position] != rune('R') {
						goto l201
					}
					position++
				}
			l205:
				{
					position207, tokenIndex207, depth207 := position, tokenIndex, depth
					if buffer[position] != rune('o') {
						goto l208
					}
					position++
					goto l207
				l208:
					position, tokenIndex, depth = position207, tokenIndex207, depth207
					if buffer[position] != rune('O') {
						goto l201
					}
					position++
				}
			l207:
				{
					position209, tokenIndex209, depth209 := position, tokenIndex, depth
					if buffer[position] != rune('u') {
						goto l210
					}
					position++
					goto l209
				l210:
					position, tokenIndex, depth = position209, tokenIndex209, depth209
					if buffer[position] != rune('U') {
						goto l201
					}
					position++
				}
			l209:
				{
					position211, tokenIndex211, depth211 := position, tokenIndex, depth
					if buffer[position] != rune('p') {
						goto l212
					}
					position++
					goto l211
				l212:
					position, tokenIndex, depth = position211, tokenIndex211, depth211
					if buffer[position] != rune('P') {
						goto l201
					}
					position++
				}
			l211:
				if !_rules[ruleKEY]() {
					goto l201
				}
				if !_rules[rule_]() {
					goto l201
				}
				{
					position213, tokenIndex213, depth213 := position, tokenIndex, depth
					if buffer[position] != rune('b') {
						goto l214
					}
					position++
					goto l213
				l214:
					position, tokenIndex, depth = position213, tokenIndex213, depth213
					if buffer[position] != rune('B') {
						goto l201
					}
					position++
				}
			l213:
				{
					position215, tokenIndex215, depth215 := position, tokenIndex, depth
					if buffer[position] != rune('y') {
						goto l216
					}
					position++
					goto l215
				l216:
					position, tokenIndex, depth = position215, tokenIndex215, depth215
					if buffer[position] != rune('Y') {
						goto l201
					}
					position++
				}
			l215:
				if !_rules[ruleKEY]() {
					goto l201
				}
				if !_rules[rule_]() {
					goto l201
				}
				{
					position217 := position
					depth++
					if !_rules[ruleCOLUMN_NAME]() {
						goto l201
					}
					depth--
					add(rulePegText, position217)
				}
				{
					add(ruleAction33, position)
				}
			l219:
				{
					position220, tokenIndex220, depth220 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l220
					}
					if !_rules[ruleCOMMA]() {
						goto l220
					}
					if !_rules[rule_]() {
						goto l220
					}
					{
						position221 := position
						depth++
						if !_rules[ruleCOLUMN_NAME]() {
							goto l220
						}
						depth--
						add(rulePegText, position221)
					}
					{
						add(ruleAction34, position)
					}
					goto l219
				l220:
					position, tokenIndex, depth = position220, tokenIndex220, depth220
				}
				depth--
				add(rulegroupByClause, position202)
			}
			return true
		l201:
			position, tokenIndex, depth = position201, tokenIndex201, depth201
			return false
		},
		/* 17 predicateClause <- <(_ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY _ predicate_1)> */
		nil,
		/* 18 predicate_1 <- <((predicate_2 _ OP_OR predicate_1 Action35) / predicate_2)> */
		func() bool {
			position224, tokenIndex224, depth224 := position, tokenIndex, depth
			{
				position225 := position
				depth++
				{
					position226, tokenIndex226, depth226 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l227
					}
					if !_rules[rule_]() {
						goto l227
					}
					{
						position228 := position
						depth++
						{
							position229, tokenIndex229, depth229 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l230
							}
							position++
							goto l229
						l230:
							position, tokenIndex, depth = position229, tokenIndex229, depth229
							if buffer[position] != rune('O') {
								goto l227
							}
							position++
						}
					l229:
						{
							position231, tokenIndex231, depth231 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l232
							}
							position++
							goto l231
						l232:
							position, tokenIndex, depth = position231, tokenIndex231, depth231
							if buffer[position] != rune('R') {
								goto l227
							}
							position++
						}
					l231:
						if !_rules[ruleKEY]() {
							goto l227
						}
						depth--
						add(ruleOP_OR, position228)
					}
					if !_rules[rulepredicate_1]() {
						goto l227
					}
					{
						add(ruleAction35, position)
					}
					goto l226
				l227:
					position, tokenIndex, depth = position226, tokenIndex226, depth226
					if !_rules[rulepredicate_2]() {
						goto l224
					}
				}
			l226:
				depth--
				add(rulepredicate_1, position225)
			}
			return true
		l224:
			position, tokenIndex, depth = position224, tokenIndex224, depth224
			return false
		},
		/* 19 predicate_2 <- <((predicate_3 _ OP_AND predicate_2 Action36) / predicate_3)> */
		func() bool {
			position234, tokenIndex234, depth234 := position, tokenIndex, depth
			{
				position235 := position
				depth++
				{
					position236, tokenIndex236, depth236 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l237
					}
					if !_rules[rule_]() {
						goto l237
					}
					{
						position238 := position
						depth++
						{
							position239, tokenIndex239, depth239 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l240
							}
							position++
							goto l239
						l240:
							position, tokenIndex, depth = position239, tokenIndex239, depth239
							if buffer[position] != rune('A') {
								goto l237
							}
							position++
						}
					l239:
						{
							position241, tokenIndex241, depth241 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l242
							}
							position++
							goto l241
						l242:
							position, tokenIndex, depth = position241, tokenIndex241, depth241
							if buffer[position] != rune('N') {
								goto l237
							}
							position++
						}
					l241:
						{
							position243, tokenIndex243, depth243 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l244
							}
							position++
							goto l243
						l244:
							position, tokenIndex, depth = position243, tokenIndex243, depth243
							if buffer[position] != rune('D') {
								goto l237
							}
							position++
						}
					l243:
						if !_rules[ruleKEY]() {
							goto l237
						}
						depth--
						add(ruleOP_AND, position238)
					}
					if !_rules[rulepredicate_2]() {
						goto l237
					}
					{
						add(ruleAction36, position)
					}
					goto l236
				l237:
					position, tokenIndex, depth = position236, tokenIndex236, depth236
					if !_rules[rulepredicate_3]() {
						goto l234
					}
				}
			l236:
				depth--
				add(rulepredicate_2, position235)
			}
			return true
		l234:
			position, tokenIndex, depth = position234, tokenIndex234, depth234
			return false
		},
		/* 20 predicate_3 <- <((_ OP_NOT predicate_3 Action37) / (_ PAREN_OPEN predicate_1 _ PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position246, tokenIndex246, depth246 := position, tokenIndex, depth
			{
				position247 := position
				depth++
				{
					position248, tokenIndex248, depth248 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l249
					}
					{
						position250 := position
						depth++
						{
							position251, tokenIndex251, depth251 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l252
							}
							position++
							goto l251
						l252:
							position, tokenIndex, depth = position251, tokenIndex251, depth251
							if buffer[position] != rune('N') {
								goto l249
							}
							position++
						}
					l251:
						{
							position253, tokenIndex253, depth253 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l254
							}
							position++
							goto l253
						l254:
							position, tokenIndex, depth = position253, tokenIndex253, depth253
							if buffer[position] != rune('O') {
								goto l249
							}
							position++
						}
					l253:
						{
							position255, tokenIndex255, depth255 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l256
							}
							position++
							goto l255
						l256:
							position, tokenIndex, depth = position255, tokenIndex255, depth255
							if buffer[position] != rune('T') {
								goto l249
							}
							position++
						}
					l255:
						if !_rules[ruleKEY]() {
							goto l249
						}
						depth--
						add(ruleOP_NOT, position250)
					}
					if !_rules[rulepredicate_3]() {
						goto l249
					}
					{
						add(ruleAction37, position)
					}
					goto l248
				l249:
					position, tokenIndex, depth = position248, tokenIndex248, depth248
					if !_rules[rule_]() {
						goto l258
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l258
					}
					if !_rules[rulepredicate_1]() {
						goto l258
					}
					if !_rules[rule_]() {
						goto l258
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l258
					}
					goto l248
				l258:
					position, tokenIndex, depth = position248, tokenIndex248, depth248
					{
						position259 := position
						depth++
						{
							position260, tokenIndex260, depth260 := position, tokenIndex, depth
							if !_rules[ruletagName]() {
								goto l261
							}
							if !_rules[rule_]() {
								goto l261
							}
							if buffer[position] != rune('=') {
								goto l261
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l261
							}
							{
								add(ruleAction38, position)
							}
							goto l260
						l261:
							position, tokenIndex, depth = position260, tokenIndex260, depth260
							if !_rules[ruletagName]() {
								goto l263
							}
							if !_rules[rule_]() {
								goto l263
							}
							if buffer[position] != rune('!') {
								goto l263
							}
							position++
							if buffer[position] != rune('=') {
								goto l263
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l263
							}
							{
								add(ruleAction39, position)
							}
							goto l260
						l263:
							position, tokenIndex, depth = position260, tokenIndex260, depth260
							if !_rules[ruletagName]() {
								goto l265
							}
							if !_rules[rule_]() {
								goto l265
							}
							{
								position266, tokenIndex266, depth266 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l267
								}
								position++
								goto l266
							l267:
								position, tokenIndex, depth = position266, tokenIndex266, depth266
								if buffer[position] != rune('M') {
									goto l265
								}
								position++
							}
						l266:
							{
								position268, tokenIndex268, depth268 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l269
								}
								position++
								goto l268
							l269:
								position, tokenIndex, depth = position268, tokenIndex268, depth268
								if buffer[position] != rune('A') {
									goto l265
								}
								position++
							}
						l268:
							{
								position270, tokenIndex270, depth270 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l271
								}
								position++
								goto l270
							l271:
								position, tokenIndex, depth = position270, tokenIndex270, depth270
								if buffer[position] != rune('T') {
									goto l265
								}
								position++
							}
						l270:
							{
								position272, tokenIndex272, depth272 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l273
								}
								position++
								goto l272
							l273:
								position, tokenIndex, depth = position272, tokenIndex272, depth272
								if buffer[position] != rune('C') {
									goto l265
								}
								position++
							}
						l272:
							{
								position274, tokenIndex274, depth274 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l275
								}
								position++
								goto l274
							l275:
								position, tokenIndex, depth = position274, tokenIndex274, depth274
								if buffer[position] != rune('H') {
									goto l265
								}
								position++
							}
						l274:
							{
								position276, tokenIndex276, depth276 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l277
								}
								position++
								goto l276
							l277:
								position, tokenIndex, depth = position276, tokenIndex276, depth276
								if buffer[position] != rune('E') {
									goto l265
								}
								position++
							}
						l276:
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
									goto l265
								}
								position++
							}
						l278:
							if !_rules[ruleKEY]() {
								goto l265
							}
							if !_rules[ruleliteralString]() {
								goto l265
							}
							{
								add(ruleAction40, position)
							}
							goto l260
						l265:
							position, tokenIndex, depth = position260, tokenIndex260, depth260
							if !_rules[ruletagName]() {
								goto l246
							}
							if !_rules[rule_]() {
								goto l246
							}
							{
								position281, tokenIndex281, depth281 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l282
								}
								position++
								goto l281
							l282:
								position, tokenIndex, depth = position281, tokenIndex281, depth281
								if buffer[position] != rune('I') {
									goto l246
								}
								position++
							}
						l281:
							{
								position283, tokenIndex283, depth283 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l284
								}
								position++
								goto l283
							l284:
								position, tokenIndex, depth = position283, tokenIndex283, depth283
								if buffer[position] != rune('N') {
									goto l246
								}
								position++
							}
						l283:
							if !_rules[ruleKEY]() {
								goto l246
							}
							{
								position285 := position
								depth++
								{
									add(ruleAction43, position)
								}
								if !_rules[rule_]() {
									goto l246
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l246
								}
								if !_rules[ruleliteralListString]() {
									goto l246
								}
							l287:
								{
									position288, tokenIndex288, depth288 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l288
									}
									if !_rules[ruleCOMMA]() {
										goto l288
									}
									if !_rules[ruleliteralListString]() {
										goto l288
									}
									goto l287
								l288:
									position, tokenIndex, depth = position288, tokenIndex288, depth288
								}
								if !_rules[rule_]() {
									goto l246
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l246
								}
								depth--
								add(ruleliteralList, position285)
							}
							{
								add(ruleAction41, position)
							}
						}
					l260:
						depth--
						add(ruletagMatcher, position259)
					}
				}
			l248:
				depth--
				add(rulepredicate_3, position247)
			}
			return true
		l246:
			position, tokenIndex, depth = position246, tokenIndex246, depth246
			return false
		},
		/* 21 tagMatcher <- <((tagName _ '=' literalString Action38) / (tagName _ ('!' '=') literalString Action39) / (tagName _ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) KEY literalString Action40) / (tagName _ (('i' / 'I') ('n' / 'N')) KEY literalList Action41))> */
		nil,
		/* 22 literalString <- <(_ STRING Action42)> */
		func() bool {
			position291, tokenIndex291, depth291 := position, tokenIndex, depth
			{
				position292 := position
				depth++
				if !_rules[rule_]() {
					goto l291
				}
				if !_rules[ruleSTRING]() {
					goto l291
				}
				{
					add(ruleAction42, position)
				}
				depth--
				add(ruleliteralString, position292)
			}
			return true
		l291:
			position, tokenIndex, depth = position291, tokenIndex291, depth291
			return false
		},
		/* 23 literalList <- <(Action43 _ PAREN_OPEN literalListString (_ COMMA literalListString)* _ PAREN_CLOSE)> */
		nil,
		/* 24 literalListString <- <(_ STRING Action44)> */
		func() bool {
			position295, tokenIndex295, depth295 := position, tokenIndex, depth
			{
				position296 := position
				depth++
				if !_rules[rule_]() {
					goto l295
				}
				if !_rules[ruleSTRING]() {
					goto l295
				}
				{
					add(ruleAction44, position)
				}
				depth--
				add(ruleliteralListString, position296)
			}
			return true
		l295:
			position, tokenIndex, depth = position295, tokenIndex295, depth295
			return false
		},
		/* 25 tagName <- <(_ <TAG_NAME> Action45)> */
		func() bool {
			position298, tokenIndex298, depth298 := position, tokenIndex, depth
			{
				position299 := position
				depth++
				if !_rules[rule_]() {
					goto l298
				}
				{
					position300 := position
					depth++
					{
						position301 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l298
						}
						depth--
						add(ruleTAG_NAME, position301)
					}
					depth--
					add(rulePegText, position300)
				}
				{
					add(ruleAction45, position)
				}
				depth--
				add(ruletagName, position299)
			}
			return true
		l298:
			position, tokenIndex, depth = position298, tokenIndex298, depth298
			return false
		},
		/* 26 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position303, tokenIndex303, depth303 := position, tokenIndex, depth
			{
				position304 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l303
				}
				depth--
				add(ruleCOLUMN_NAME, position304)
			}
			return true
		l303:
			position, tokenIndex, depth = position303, tokenIndex303, depth303
			return false
		},
		/* 27 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 28 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 29 IDENTIFIER <- <(('`' CHAR* '`') / (_ !(KEYWORD KEY) ID_SEGMENT ('.' ID_SEGMENT)*))> */
		func() bool {
			position307, tokenIndex307, depth307 := position, tokenIndex, depth
			{
				position308 := position
				depth++
				{
					position309, tokenIndex309, depth309 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l310
					}
					position++
				l311:
					{
						position312, tokenIndex312, depth312 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l312
						}
						goto l311
					l312:
						position, tokenIndex, depth = position312, tokenIndex312, depth312
					}
					if buffer[position] != rune('`') {
						goto l310
					}
					position++
					goto l309
				l310:
					position, tokenIndex, depth = position309, tokenIndex309, depth309
					if !_rules[rule_]() {
						goto l307
					}
					{
						position313, tokenIndex313, depth313 := position, tokenIndex, depth
						{
							position314 := position
							depth++
							{
								position315, tokenIndex315, depth315 := position, tokenIndex, depth
								{
									position317, tokenIndex317, depth317 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l318
									}
									position++
									goto l317
								l318:
									position, tokenIndex, depth = position317, tokenIndex317, depth317
									if buffer[position] != rune('A') {
										goto l316
									}
									position++
								}
							l317:
								{
									position319, tokenIndex319, depth319 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l320
									}
									position++
									goto l319
								l320:
									position, tokenIndex, depth = position319, tokenIndex319, depth319
									if buffer[position] != rune('L') {
										goto l316
									}
									position++
								}
							l319:
								{
									position321, tokenIndex321, depth321 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l322
									}
									position++
									goto l321
								l322:
									position, tokenIndex, depth = position321, tokenIndex321, depth321
									if buffer[position] != rune('L') {
										goto l316
									}
									position++
								}
							l321:
								goto l315
							l316:
								position, tokenIndex, depth = position315, tokenIndex315, depth315
								{
									position324, tokenIndex324, depth324 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l325
									}
									position++
									goto l324
								l325:
									position, tokenIndex, depth = position324, tokenIndex324, depth324
									if buffer[position] != rune('A') {
										goto l323
									}
									position++
								}
							l324:
								{
									position326, tokenIndex326, depth326 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l327
									}
									position++
									goto l326
								l327:
									position, tokenIndex, depth = position326, tokenIndex326, depth326
									if buffer[position] != rune('N') {
										goto l323
									}
									position++
								}
							l326:
								{
									position328, tokenIndex328, depth328 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l329
									}
									position++
									goto l328
								l329:
									position, tokenIndex, depth = position328, tokenIndex328, depth328
									if buffer[position] != rune('D') {
										goto l323
									}
									position++
								}
							l328:
								goto l315
							l323:
								position, tokenIndex, depth = position315, tokenIndex315, depth315
								{
									position331, tokenIndex331, depth331 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l332
									}
									position++
									goto l331
								l332:
									position, tokenIndex, depth = position331, tokenIndex331, depth331
									if buffer[position] != rune('M') {
										goto l330
									}
									position++
								}
							l331:
								{
									position333, tokenIndex333, depth333 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l334
									}
									position++
									goto l333
								l334:
									position, tokenIndex, depth = position333, tokenIndex333, depth333
									if buffer[position] != rune('A') {
										goto l330
									}
									position++
								}
							l333:
								{
									position335, tokenIndex335, depth335 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l336
									}
									position++
									goto l335
								l336:
									position, tokenIndex, depth = position335, tokenIndex335, depth335
									if buffer[position] != rune('T') {
										goto l330
									}
									position++
								}
							l335:
								{
									position337, tokenIndex337, depth337 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l338
									}
									position++
									goto l337
								l338:
									position, tokenIndex, depth = position337, tokenIndex337, depth337
									if buffer[position] != rune('C') {
										goto l330
									}
									position++
								}
							l337:
								{
									position339, tokenIndex339, depth339 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l340
									}
									position++
									goto l339
								l340:
									position, tokenIndex, depth = position339, tokenIndex339, depth339
									if buffer[position] != rune('H') {
										goto l330
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
										goto l330
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
										goto l330
									}
									position++
								}
							l343:
								goto l315
							l330:
								position, tokenIndex, depth = position315, tokenIndex315, depth315
								{
									position346, tokenIndex346, depth346 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l347
									}
									position++
									goto l346
								l347:
									position, tokenIndex, depth = position346, tokenIndex346, depth346
									if buffer[position] != rune('S') {
										goto l345
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
										goto l345
									}
									position++
								}
							l348:
								{
									position350, tokenIndex350, depth350 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l351
									}
									position++
									goto l350
								l351:
									position, tokenIndex, depth = position350, tokenIndex350, depth350
									if buffer[position] != rune('L') {
										goto l345
									}
									position++
								}
							l350:
								{
									position352, tokenIndex352, depth352 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l353
									}
									position++
									goto l352
								l353:
									position, tokenIndex, depth = position352, tokenIndex352, depth352
									if buffer[position] != rune('E') {
										goto l345
									}
									position++
								}
							l352:
								{
									position354, tokenIndex354, depth354 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l355
									}
									position++
									goto l354
								l355:
									position, tokenIndex, depth = position354, tokenIndex354, depth354
									if buffer[position] != rune('C') {
										goto l345
									}
									position++
								}
							l354:
								{
									position356, tokenIndex356, depth356 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l357
									}
									position++
									goto l356
								l357:
									position, tokenIndex, depth = position356, tokenIndex356, depth356
									if buffer[position] != rune('T') {
										goto l345
									}
									position++
								}
							l356:
								goto l315
							l345:
								position, tokenIndex, depth = position315, tokenIndex315, depth315
								{
									switch buffer[position] {
									case 'M', 'm':
										{
											position359, tokenIndex359, depth359 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l360
											}
											position++
											goto l359
										l360:
											position, tokenIndex, depth = position359, tokenIndex359, depth359
											if buffer[position] != rune('M') {
												goto l313
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
												goto l313
											}
											position++
										}
									l361:
										{
											position363, tokenIndex363, depth363 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l364
											}
											position++
											goto l363
										l364:
											position, tokenIndex, depth = position363, tokenIndex363, depth363
											if buffer[position] != rune('T') {
												goto l313
											}
											position++
										}
									l363:
										{
											position365, tokenIndex365, depth365 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l366
											}
											position++
											goto l365
										l366:
											position, tokenIndex, depth = position365, tokenIndex365, depth365
											if buffer[position] != rune('R') {
												goto l313
											}
											position++
										}
									l365:
										{
											position367, tokenIndex367, depth367 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l368
											}
											position++
											goto l367
										l368:
											position, tokenIndex, depth = position367, tokenIndex367, depth367
											if buffer[position] != rune('I') {
												goto l313
											}
											position++
										}
									l367:
										{
											position369, tokenIndex369, depth369 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l370
											}
											position++
											goto l369
										l370:
											position, tokenIndex, depth = position369, tokenIndex369, depth369
											if buffer[position] != rune('C') {
												goto l313
											}
											position++
										}
									l369:
										{
											position371, tokenIndex371, depth371 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l372
											}
											position++
											goto l371
										l372:
											position, tokenIndex, depth = position371, tokenIndex371, depth371
											if buffer[position] != rune('S') {
												goto l313
											}
											position++
										}
									l371:
										break
									case 'W', 'w':
										{
											position373, tokenIndex373, depth373 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l374
											}
											position++
											goto l373
										l374:
											position, tokenIndex, depth = position373, tokenIndex373, depth373
											if buffer[position] != rune('W') {
												goto l313
											}
											position++
										}
									l373:
										{
											position375, tokenIndex375, depth375 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l376
											}
											position++
											goto l375
										l376:
											position, tokenIndex, depth = position375, tokenIndex375, depth375
											if buffer[position] != rune('H') {
												goto l313
											}
											position++
										}
									l375:
										{
											position377, tokenIndex377, depth377 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l378
											}
											position++
											goto l377
										l378:
											position, tokenIndex, depth = position377, tokenIndex377, depth377
											if buffer[position] != rune('E') {
												goto l313
											}
											position++
										}
									l377:
										{
											position379, tokenIndex379, depth379 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l380
											}
											position++
											goto l379
										l380:
											position, tokenIndex, depth = position379, tokenIndex379, depth379
											if buffer[position] != rune('R') {
												goto l313
											}
											position++
										}
									l379:
										{
											position381, tokenIndex381, depth381 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l382
											}
											position++
											goto l381
										l382:
											position, tokenIndex, depth = position381, tokenIndex381, depth381
											if buffer[position] != rune('E') {
												goto l313
											}
											position++
										}
									l381:
										break
									case 'O', 'o':
										{
											position383, tokenIndex383, depth383 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l384
											}
											position++
											goto l383
										l384:
											position, tokenIndex, depth = position383, tokenIndex383, depth383
											if buffer[position] != rune('O') {
												goto l313
											}
											position++
										}
									l383:
										{
											position385, tokenIndex385, depth385 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l386
											}
											position++
											goto l385
										l386:
											position, tokenIndex, depth = position385, tokenIndex385, depth385
											if buffer[position] != rune('R') {
												goto l313
											}
											position++
										}
									l385:
										break
									case 'N', 'n':
										{
											position387, tokenIndex387, depth387 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l388
											}
											position++
											goto l387
										l388:
											position, tokenIndex, depth = position387, tokenIndex387, depth387
											if buffer[position] != rune('N') {
												goto l313
											}
											position++
										}
									l387:
										{
											position389, tokenIndex389, depth389 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l390
											}
											position++
											goto l389
										l390:
											position, tokenIndex, depth = position389, tokenIndex389, depth389
											if buffer[position] != rune('O') {
												goto l313
											}
											position++
										}
									l389:
										{
											position391, tokenIndex391, depth391 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l392
											}
											position++
											goto l391
										l392:
											position, tokenIndex, depth = position391, tokenIndex391, depth391
											if buffer[position] != rune('T') {
												goto l313
											}
											position++
										}
									l391:
										break
									case 'I', 'i':
										{
											position393, tokenIndex393, depth393 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l394
											}
											position++
											goto l393
										l394:
											position, tokenIndex, depth = position393, tokenIndex393, depth393
											if buffer[position] != rune('I') {
												goto l313
											}
											position++
										}
									l393:
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
												goto l313
											}
											position++
										}
									l395:
										break
									case 'G', 'g':
										{
											position397, tokenIndex397, depth397 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l398
											}
											position++
											goto l397
										l398:
											position, tokenIndex, depth = position397, tokenIndex397, depth397
											if buffer[position] != rune('G') {
												goto l313
											}
											position++
										}
									l397:
										{
											position399, tokenIndex399, depth399 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l400
											}
											position++
											goto l399
										l400:
											position, tokenIndex, depth = position399, tokenIndex399, depth399
											if buffer[position] != rune('R') {
												goto l313
											}
											position++
										}
									l399:
										{
											position401, tokenIndex401, depth401 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l402
											}
											position++
											goto l401
										l402:
											position, tokenIndex, depth = position401, tokenIndex401, depth401
											if buffer[position] != rune('O') {
												goto l313
											}
											position++
										}
									l401:
										{
											position403, tokenIndex403, depth403 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l404
											}
											position++
											goto l403
										l404:
											position, tokenIndex, depth = position403, tokenIndex403, depth403
											if buffer[position] != rune('U') {
												goto l313
											}
											position++
										}
									l403:
										{
											position405, tokenIndex405, depth405 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l406
											}
											position++
											goto l405
										l406:
											position, tokenIndex, depth = position405, tokenIndex405, depth405
											if buffer[position] != rune('P') {
												goto l313
											}
											position++
										}
									l405:
										break
									case 'D', 'd':
										{
											position407, tokenIndex407, depth407 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l408
											}
											position++
											goto l407
										l408:
											position, tokenIndex, depth = position407, tokenIndex407, depth407
											if buffer[position] != rune('D') {
												goto l313
											}
											position++
										}
									l407:
										{
											position409, tokenIndex409, depth409 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l410
											}
											position++
											goto l409
										l410:
											position, tokenIndex, depth = position409, tokenIndex409, depth409
											if buffer[position] != rune('E') {
												goto l313
											}
											position++
										}
									l409:
										{
											position411, tokenIndex411, depth411 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l412
											}
											position++
											goto l411
										l412:
											position, tokenIndex, depth = position411, tokenIndex411, depth411
											if buffer[position] != rune('S') {
												goto l313
											}
											position++
										}
									l411:
										{
											position413, tokenIndex413, depth413 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l414
											}
											position++
											goto l413
										l414:
											position, tokenIndex, depth = position413, tokenIndex413, depth413
											if buffer[position] != rune('C') {
												goto l313
											}
											position++
										}
									l413:
										{
											position415, tokenIndex415, depth415 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l416
											}
											position++
											goto l415
										l416:
											position, tokenIndex, depth = position415, tokenIndex415, depth415
											if buffer[position] != rune('R') {
												goto l313
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
												goto l313
											}
											position++
										}
									l417:
										{
											position419, tokenIndex419, depth419 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l420
											}
											position++
											goto l419
										l420:
											position, tokenIndex, depth = position419, tokenIndex419, depth419
											if buffer[position] != rune('B') {
												goto l313
											}
											position++
										}
									l419:
										{
											position421, tokenIndex421, depth421 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l422
											}
											position++
											goto l421
										l422:
											position, tokenIndex, depth = position421, tokenIndex421, depth421
											if buffer[position] != rune('E') {
												goto l313
											}
											position++
										}
									l421:
										break
									case 'B', 'b':
										{
											position423, tokenIndex423, depth423 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l424
											}
											position++
											goto l423
										l424:
											position, tokenIndex, depth = position423, tokenIndex423, depth423
											if buffer[position] != rune('B') {
												goto l313
											}
											position++
										}
									l423:
										{
											position425, tokenIndex425, depth425 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l426
											}
											position++
											goto l425
										l426:
											position, tokenIndex, depth = position425, tokenIndex425, depth425
											if buffer[position] != rune('Y') {
												goto l313
											}
											position++
										}
									l425:
										break
									case 'A', 'a':
										{
											position427, tokenIndex427, depth427 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l428
											}
											position++
											goto l427
										l428:
											position, tokenIndex, depth = position427, tokenIndex427, depth427
											if buffer[position] != rune('A') {
												goto l313
											}
											position++
										}
									l427:
										{
											position429, tokenIndex429, depth429 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l430
											}
											position++
											goto l429
										l430:
											position, tokenIndex, depth = position429, tokenIndex429, depth429
											if buffer[position] != rune('S') {
												goto l313
											}
											position++
										}
									l429:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l313
										}
										break
									}
								}

							}
						l315:
							depth--
							add(ruleKEYWORD, position314)
						}
						if !_rules[ruleKEY]() {
							goto l313
						}
						goto l307
					l313:
						position, tokenIndex, depth = position313, tokenIndex313, depth313
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l307
					}
				l431:
					{
						position432, tokenIndex432, depth432 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l432
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l432
						}
						goto l431
					l432:
						position, tokenIndex, depth = position432, tokenIndex432, depth432
					}
				}
			l309:
				depth--
				add(ruleIDENTIFIER, position308)
			}
			return true
		l307:
			position, tokenIndex, depth = position307, tokenIndex307, depth307
			return false
		},
		/* 30 TIMESTAMP <- <((_ <(NUMBER ([a-z] / [A-Z])*)>) / (_ STRING) / (_ <(('n' / 'N') ('o' / 'O') ('w' / 'W'))>))> */
		nil,
		/* 31 ID_SEGMENT <- <(_ ID_START ID_CONT*)> */
		func() bool {
			position434, tokenIndex434, depth434 := position, tokenIndex, depth
			{
				position435 := position
				depth++
				if !_rules[rule_]() {
					goto l434
				}
				if !_rules[ruleID_START]() {
					goto l434
				}
			l436:
				{
					position437, tokenIndex437, depth437 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l437
					}
					goto l436
				l437:
					position, tokenIndex, depth = position437, tokenIndex437, depth437
				}
				depth--
				add(ruleID_SEGMENT, position435)
			}
			return true
		l434:
			position, tokenIndex, depth = position434, tokenIndex434, depth434
			return false
		},
		/* 32 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position438, tokenIndex438, depth438 := position, tokenIndex, depth
			{
				position439 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l438
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l438
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l438
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position439)
			}
			return true
		l438:
			position, tokenIndex, depth = position438, tokenIndex438, depth438
			return false
		},
		/* 33 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position441, tokenIndex441, depth441 := position, tokenIndex, depth
			{
				position442 := position
				depth++
				{
					position443, tokenIndex443, depth443 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l444
					}
					goto l443
				l444:
					position, tokenIndex, depth = position443, tokenIndex443, depth443
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l441
					}
					position++
				}
			l443:
				depth--
				add(ruleID_CONT, position442)
			}
			return true
		l441:
			position, tokenIndex, depth = position441, tokenIndex441, depth441
			return false
		},
		/* 34 PROPERTY_KEY <- <(((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> KEY _ (('b' / 'B') ('y' / 'Y')))) | (&('R' | 'r') <(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))>) | (&('T' | 't') <(('t' / 'T') ('o' / 'O'))>) | (&('F' | 'f') <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))>)) KEY)> */
		func() bool {
			position445, tokenIndex445, depth445 := position, tokenIndex, depth
			{
				position446 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position448 := position
							depth++
							{
								position449, tokenIndex449, depth449 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l450
								}
								position++
								goto l449
							l450:
								position, tokenIndex, depth = position449, tokenIndex449, depth449
								if buffer[position] != rune('S') {
									goto l445
								}
								position++
							}
						l449:
							{
								position451, tokenIndex451, depth451 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l452
								}
								position++
								goto l451
							l452:
								position, tokenIndex, depth = position451, tokenIndex451, depth451
								if buffer[position] != rune('A') {
									goto l445
								}
								position++
							}
						l451:
							{
								position453, tokenIndex453, depth453 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l454
								}
								position++
								goto l453
							l454:
								position, tokenIndex, depth = position453, tokenIndex453, depth453
								if buffer[position] != rune('M') {
									goto l445
								}
								position++
							}
						l453:
							{
								position455, tokenIndex455, depth455 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l456
								}
								position++
								goto l455
							l456:
								position, tokenIndex, depth = position455, tokenIndex455, depth455
								if buffer[position] != rune('P') {
									goto l445
								}
								position++
							}
						l455:
							{
								position457, tokenIndex457, depth457 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l458
								}
								position++
								goto l457
							l458:
								position, tokenIndex, depth = position457, tokenIndex457, depth457
								if buffer[position] != rune('L') {
									goto l445
								}
								position++
							}
						l457:
							{
								position459, tokenIndex459, depth459 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l460
								}
								position++
								goto l459
							l460:
								position, tokenIndex, depth = position459, tokenIndex459, depth459
								if buffer[position] != rune('E') {
									goto l445
								}
								position++
							}
						l459:
							depth--
							add(rulePegText, position448)
						}
						if !_rules[ruleKEY]() {
							goto l445
						}
						if !_rules[rule_]() {
							goto l445
						}
						{
							position461, tokenIndex461, depth461 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l462
							}
							position++
							goto l461
						l462:
							position, tokenIndex, depth = position461, tokenIndex461, depth461
							if buffer[position] != rune('B') {
								goto l445
							}
							position++
						}
					l461:
						{
							position463, tokenIndex463, depth463 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l464
							}
							position++
							goto l463
						l464:
							position, tokenIndex, depth = position463, tokenIndex463, depth463
							if buffer[position] != rune('Y') {
								goto l445
							}
							position++
						}
					l463:
						break
					case 'R', 'r':
						{
							position465 := position
							depth++
							{
								position466, tokenIndex466, depth466 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l467
								}
								position++
								goto l466
							l467:
								position, tokenIndex, depth = position466, tokenIndex466, depth466
								if buffer[position] != rune('R') {
									goto l445
								}
								position++
							}
						l466:
							{
								position468, tokenIndex468, depth468 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l469
								}
								position++
								goto l468
							l469:
								position, tokenIndex, depth = position468, tokenIndex468, depth468
								if buffer[position] != rune('E') {
									goto l445
								}
								position++
							}
						l468:
							{
								position470, tokenIndex470, depth470 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l471
								}
								position++
								goto l470
							l471:
								position, tokenIndex, depth = position470, tokenIndex470, depth470
								if buffer[position] != rune('S') {
									goto l445
								}
								position++
							}
						l470:
							{
								position472, tokenIndex472, depth472 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l473
								}
								position++
								goto l472
							l473:
								position, tokenIndex, depth = position472, tokenIndex472, depth472
								if buffer[position] != rune('O') {
									goto l445
								}
								position++
							}
						l472:
							{
								position474, tokenIndex474, depth474 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l475
								}
								position++
								goto l474
							l475:
								position, tokenIndex, depth = position474, tokenIndex474, depth474
								if buffer[position] != rune('L') {
									goto l445
								}
								position++
							}
						l474:
							{
								position476, tokenIndex476, depth476 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l477
								}
								position++
								goto l476
							l477:
								position, tokenIndex, depth = position476, tokenIndex476, depth476
								if buffer[position] != rune('U') {
									goto l445
								}
								position++
							}
						l476:
							{
								position478, tokenIndex478, depth478 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l479
								}
								position++
								goto l478
							l479:
								position, tokenIndex, depth = position478, tokenIndex478, depth478
								if buffer[position] != rune('T') {
									goto l445
								}
								position++
							}
						l478:
							{
								position480, tokenIndex480, depth480 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l481
								}
								position++
								goto l480
							l481:
								position, tokenIndex, depth = position480, tokenIndex480, depth480
								if buffer[position] != rune('I') {
									goto l445
								}
								position++
							}
						l480:
							{
								position482, tokenIndex482, depth482 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l483
								}
								position++
								goto l482
							l483:
								position, tokenIndex, depth = position482, tokenIndex482, depth482
								if buffer[position] != rune('O') {
									goto l445
								}
								position++
							}
						l482:
							{
								position484, tokenIndex484, depth484 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l485
								}
								position++
								goto l484
							l485:
								position, tokenIndex, depth = position484, tokenIndex484, depth484
								if buffer[position] != rune('N') {
									goto l445
								}
								position++
							}
						l484:
							depth--
							add(rulePegText, position465)
						}
						break
					case 'T', 't':
						{
							position486 := position
							depth++
							{
								position487, tokenIndex487, depth487 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l488
								}
								position++
								goto l487
							l488:
								position, tokenIndex, depth = position487, tokenIndex487, depth487
								if buffer[position] != rune('T') {
									goto l445
								}
								position++
							}
						l487:
							{
								position489, tokenIndex489, depth489 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l490
								}
								position++
								goto l489
							l490:
								position, tokenIndex, depth = position489, tokenIndex489, depth489
								if buffer[position] != rune('O') {
									goto l445
								}
								position++
							}
						l489:
							depth--
							add(rulePegText, position486)
						}
						break
					default:
						{
							position491 := position
							depth++
							{
								position492, tokenIndex492, depth492 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l493
								}
								position++
								goto l492
							l493:
								position, tokenIndex, depth = position492, tokenIndex492, depth492
								if buffer[position] != rune('F') {
									goto l445
								}
								position++
							}
						l492:
							{
								position494, tokenIndex494, depth494 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l495
								}
								position++
								goto l494
							l495:
								position, tokenIndex, depth = position494, tokenIndex494, depth494
								if buffer[position] != rune('R') {
									goto l445
								}
								position++
							}
						l494:
							{
								position496, tokenIndex496, depth496 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l497
								}
								position++
								goto l496
							l497:
								position, tokenIndex, depth = position496, tokenIndex496, depth496
								if buffer[position] != rune('O') {
									goto l445
								}
								position++
							}
						l496:
							{
								position498, tokenIndex498, depth498 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l499
								}
								position++
								goto l498
							l499:
								position, tokenIndex, depth = position498, tokenIndex498, depth498
								if buffer[position] != rune('M') {
									goto l445
								}
								position++
							}
						l498:
							depth--
							add(rulePegText, position491)
						}
						break
					}
				}

				if !_rules[ruleKEY]() {
					goto l445
				}
				depth--
				add(rulePROPERTY_KEY, position446)
			}
			return true
		l445:
			position, tokenIndex, depth = position445, tokenIndex445, depth445
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
			position510, tokenIndex510, depth510 := position, tokenIndex, depth
			{
				position511 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l510
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position511)
			}
			return true
		l510:
			position, tokenIndex, depth = position510, tokenIndex510, depth510
			return false
		},
		/* 46 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position512, tokenIndex512, depth512 := position, tokenIndex, depth
			{
				position513 := position
				depth++
				if buffer[position] != rune('"') {
					goto l512
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position513)
			}
			return true
		l512:
			position, tokenIndex, depth = position512, tokenIndex512, depth512
			return false
		},
		/* 47 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> QUOTE_SINGLE) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> QUOTE_DOUBLE))> */
		func() bool {
			position514, tokenIndex514, depth514 := position, tokenIndex, depth
			{
				position515 := position
				depth++
				{
					position516, tokenIndex516, depth516 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l517
					}
					{
						position518 := position
						depth++
					l519:
						{
							position520, tokenIndex520, depth520 := position, tokenIndex, depth
							{
								position521, tokenIndex521, depth521 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l521
								}
								goto l520
							l521:
								position, tokenIndex, depth = position521, tokenIndex521, depth521
							}
							if !_rules[ruleCHAR]() {
								goto l520
							}
							goto l519
						l520:
							position, tokenIndex, depth = position520, tokenIndex520, depth520
						}
						depth--
						add(rulePegText, position518)
					}
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l517
					}
					goto l516
				l517:
					position, tokenIndex, depth = position516, tokenIndex516, depth516
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l514
					}
					{
						position522 := position
						depth++
					l523:
						{
							position524, tokenIndex524, depth524 := position, tokenIndex, depth
							{
								position525, tokenIndex525, depth525 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l525
								}
								goto l524
							l525:
								position, tokenIndex, depth = position525, tokenIndex525, depth525
							}
							if !_rules[ruleCHAR]() {
								goto l524
							}
							goto l523
						l524:
							position, tokenIndex, depth = position524, tokenIndex524, depth524
						}
						depth--
						add(rulePegText, position522)
					}
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l514
					}
				}
			l516:
				depth--
				add(ruleSTRING, position515)
			}
			return true
		l514:
			position, tokenIndex, depth = position514, tokenIndex514, depth514
			return false
		},
		/* 48 CHAR <- <(('\\' ((&('"') QUOTE_DOUBLE) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position526, tokenIndex526, depth526 := position, tokenIndex, depth
			{
				position527 := position
				depth++
				{
					position528, tokenIndex528, depth528 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l529
					}
					position++
					{
						switch buffer[position] {
						case '"':
							if !_rules[ruleQUOTE_DOUBLE]() {
								goto l529
							}
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l529
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l529
							}
							break
						}
					}

					goto l528
				l529:
					position, tokenIndex, depth = position528, tokenIndex528, depth528
					{
						position531, tokenIndex531, depth531 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l531
						}
						goto l526
					l531:
						position, tokenIndex, depth = position531, tokenIndex531, depth531
					}
					if !matchDot() {
						goto l526
					}
				}
			l528:
				depth--
				add(ruleCHAR, position527)
			}
			return true
		l526:
			position, tokenIndex, depth = position526, tokenIndex526, depth526
			return false
		},
		/* 49 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position532, tokenIndex532, depth532 := position, tokenIndex, depth
			{
				position533 := position
				depth++
				{
					position534, tokenIndex534, depth534 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l535
					}
					position++
					goto l534
				l535:
					position, tokenIndex, depth = position534, tokenIndex534, depth534
					if buffer[position] != rune('\\') {
						goto l532
					}
					position++
				}
			l534:
				depth--
				add(ruleESCAPE_CLASS, position533)
			}
			return true
		l532:
			position, tokenIndex, depth = position532, tokenIndex532, depth532
			return false
		},
		/* 50 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position536, tokenIndex536, depth536 := position, tokenIndex, depth
			{
				position537 := position
				depth++
				{
					position538 := position
					depth++
					{
						position539, tokenIndex539, depth539 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l539
						}
						position++
						goto l540
					l539:
						position, tokenIndex, depth = position539, tokenIndex539, depth539
					}
				l540:
					{
						position541 := position
						depth++
						{
							position542, tokenIndex542, depth542 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l543
							}
							position++
							goto l542
						l543:
							position, tokenIndex, depth = position542, tokenIndex542, depth542
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l536
							}
							position++
						l544:
							{
								position545, tokenIndex545, depth545 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l545
								}
								position++
								goto l544
							l545:
								position, tokenIndex, depth = position545, tokenIndex545, depth545
							}
						}
					l542:
						depth--
						add(ruleNUMBER_NATURAL, position541)
					}
					depth--
					add(ruleNUMBER_INTEGER, position538)
				}
				{
					position546, tokenIndex546, depth546 := position, tokenIndex, depth
					{
						position548 := position
						depth++
						if buffer[position] != rune('.') {
							goto l546
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l546
						}
						position++
					l549:
						{
							position550, tokenIndex550, depth550 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l550
							}
							position++
							goto l549
						l550:
							position, tokenIndex, depth = position550, tokenIndex550, depth550
						}
						depth--
						add(ruleNUMBER_FRACTION, position548)
					}
					goto l547
				l546:
					position, tokenIndex, depth = position546, tokenIndex546, depth546
				}
			l547:
				{
					position551, tokenIndex551, depth551 := position, tokenIndex, depth
					{
						position553 := position
						depth++
						{
							position554, tokenIndex554, depth554 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l555
							}
							position++
							goto l554
						l555:
							position, tokenIndex, depth = position554, tokenIndex554, depth554
							if buffer[position] != rune('E') {
								goto l551
							}
							position++
						}
					l554:
						{
							position556, tokenIndex556, depth556 := position, tokenIndex, depth
							{
								position558, tokenIndex558, depth558 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l559
								}
								position++
								goto l558
							l559:
								position, tokenIndex, depth = position558, tokenIndex558, depth558
								if buffer[position] != rune('-') {
									goto l556
								}
								position++
							}
						l558:
							goto l557
						l556:
							position, tokenIndex, depth = position556, tokenIndex556, depth556
						}
					l557:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l551
						}
						position++
					l560:
						{
							position561, tokenIndex561, depth561 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l561
							}
							position++
							goto l560
						l561:
							position, tokenIndex, depth = position561, tokenIndex561, depth561
						}
						depth--
						add(ruleNUMBER_EXP, position553)
					}
					goto l552
				l551:
					position, tokenIndex, depth = position551, tokenIndex551, depth551
				}
			l552:
				depth--
				add(ruleNUMBER, position537)
			}
			return true
		l536:
			position, tokenIndex, depth = position536, tokenIndex536, depth536
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
		/* 55 PAREN_OPEN <- <'('> */
		func() bool {
			position566, tokenIndex566, depth566 := position, tokenIndex, depth
			{
				position567 := position
				depth++
				if buffer[position] != rune('(') {
					goto l566
				}
				position++
				depth--
				add(rulePAREN_OPEN, position567)
			}
			return true
		l566:
			position, tokenIndex, depth = position566, tokenIndex566, depth566
			return false
		},
		/* 56 PAREN_CLOSE <- <')'> */
		func() bool {
			position568, tokenIndex568, depth568 := position, tokenIndex, depth
			{
				position569 := position
				depth++
				if buffer[position] != rune(')') {
					goto l568
				}
				position++
				depth--
				add(rulePAREN_CLOSE, position569)
			}
			return true
		l568:
			position, tokenIndex, depth = position568, tokenIndex568, depth568
			return false
		},
		/* 57 COMMA <- <','> */
		func() bool {
			position570, tokenIndex570, depth570 := position, tokenIndex, depth
			{
				position571 := position
				depth++
				if buffer[position] != rune(',') {
					goto l570
				}
				position++
				depth--
				add(ruleCOMMA, position571)
			}
			return true
		l570:
			position, tokenIndex, depth = position570, tokenIndex570, depth570
			return false
		},
		/* 58 _ <- <SPACE*> */
		func() bool {
			{
				position573 := position
				depth++
			l574:
				{
					position575, tokenIndex575, depth575 := position, tokenIndex, depth
					{
						position576 := position
						depth++
						{
							switch buffer[position] {
							case '\t':
								if buffer[position] != rune('\t') {
									goto l575
								}
								position++
								break
							case '\n':
								if buffer[position] != rune('\n') {
									goto l575
								}
								position++
								break
							default:
								if buffer[position] != rune(' ') {
									goto l575
								}
								position++
								break
							}
						}

						depth--
						add(ruleSPACE, position576)
					}
					goto l574
				l575:
					position, tokenIndex, depth = position575, tokenIndex575, depth575
				}
				depth--
				add(rule_, position573)
			}
			return true
		},
		/* 59 KEY <- <!ID_CONT> */
		func() bool {
			position578, tokenIndex578, depth578 := position, tokenIndex, depth
			{
				position579 := position
				depth++
				{
					position580, tokenIndex580, depth580 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l580
					}
					goto l578
				l580:
					position, tokenIndex, depth = position580, tokenIndex580, depth580
				}
				depth--
				add(ruleKEY, position579)
			}
			return true
		l578:
			position, tokenIndex, depth = position578, tokenIndex578, depth578
			return false
		},
		/* 60 SPACE <- <((&('\t') '\t') | (&('\n') '\n') | (&(' ') ' '))> */
		nil,
		/* 62 Action0 <- <{
		   p.makeSelect()
		 }> */
		nil,
		/* 63 Action1 <- <{ p.makeDescribeAll() }> */
		nil,
		/* 64 Action2 <- <{ p.makeDescribeMetrics() }> */
		nil,
		nil,
		/* 66 Action3 <- <{ p.addStringLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 67 Action4 <- <{ p.makeDescribe() }> */
		nil,
		/* 68 Action5 <- <{ p.addEvaluationContext() }> */
		nil,
		/* 69 Action6 <- <{ p.addPropertyKey(buffer[begin:end])   }> */
		nil,
		/* 70 Action7 <- <{ p.addPropertyValue(buffer[begin:end]) }> */
		nil,
		/* 71 Action8 <- <{ p.insertPropertyKeyValue() }> */
		nil,
		/* 72 Action9 <- <{ p.checkPropertyClause() }> */
		nil,
		/* 73 Action10 <- <{ p.addNullPredicate() }> */
		nil,
		/* 74 Action11 <- <{ p.addExpressionList() }> */
		nil,
		/* 75 Action12 <- <{ p.appendExpression() }> */
		nil,
		/* 76 Action13 <- <{ p.appendExpression() }> */
		nil,
		/* 77 Action14 <- <{ p.addOperatorLiteral("+") }> */
		nil,
		/* 78 Action15 <- <{ p.addOperatorLiteral("-") }> */
		nil,
		/* 79 Action16 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 80 Action17 <- <{ p.addOperatorLiteral("/") }> */
		nil,
		/* 81 Action18 <- <{ p.addOperatorLiteral("*") }> */
		nil,
		/* 82 Action19 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 83 Action20 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 84 Action21 <- <{p.addExpressionList()}> */
		nil,
		/* 85 Action22 <- <{ p.addGroupBy() }> */
		nil,
		/* 86 Action23 <- <{
		   p.addExpressionList()
		   p.addGroupBy()
		 }> */
		nil,
		/* 87 Action24 <- <{
		   p.addPipeExpression()
		 }> */
		nil,
		/* 88 Action25 <- <{ p.addNumberNode(buffer[begin:end]) }> */
		nil,
		/* 89 Action26 <- <{ p.addStringNode(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 90 Action27 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 91 Action28 <- <{ p.addGroupBy() }> */
		nil,
		/* 92 Action29 <- <{
		   p.addFunctionInvocation()
		 }> */
		nil,
		/* 93 Action30 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 94 Action31 <- <{ p.addNullPredicate() }> */
		nil,
		/* 95 Action32 <- <{
		   p.addMetricExpression()
		 }> */
		nil,
		/* 96 Action33 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 97 Action34 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		   }> */
		nil,
		/* 98 Action35 <- <{ p.addOrPredicate() }> */
		nil,
		/* 99 Action36 <- <{ p.addAndPredicate() }> */
		nil,
		/* 100 Action37 <- <{ p.addNotPredicate() }> */
		nil,
		/* 101 Action38 <- <{
		   p.addLiteralMatcher()
		 }> */
		nil,
		/* 102 Action39 <- <{
		   p.addLiteralMatcher()
		   p.addNotPredicate()
		 }> */
		nil,
		/* 103 Action40 <- <{
		   p.addRegexMatcher()
		 }> */
		nil,
		/* 104 Action41 <- <{
		   p.addListMatcher()
		 }> */
		nil,
		/* 105 Action42 <- <{
		  p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 106 Action43 <- <{ p.addLiteralList() }> */
		nil,
		/* 107 Action44 <- <{
		  p.appendLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 108 Action45 <- <{ p.addTagLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
	}
	p.rules = _rules
}
