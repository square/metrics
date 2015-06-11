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
	ruleIDENTIFIER
	ruleTIMESTAMP
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
	ruleAction34
	ruleAction35
	ruleAction36
	ruleAction37
	ruleAction38

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
	"IDENTIFIER",
	"TIMESTAMP",
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
	"Action34",
	"Action35",
	"Action36",
	"Action37",
	"Action38",

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
	rules  [98]func() bool
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
			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		case ruleAction3:
			p.makeDescribe()
		case ruleAction4:
			p.addEvaluationContext()
		case ruleAction5:
			p.addPropertyKey(buffer[begin:end])
		case ruleAction6:
			p.addPropertyValue(buffer[begin:end])
		case ruleAction7:
			p.insertPropertyKeyValue()
		case ruleAction8:
			p.checkPropertyClause()
		case ruleAction9:
			p.addNullPredicate()
		case ruleAction10:
			p.addExpressionList()
		case ruleAction11:
			p.appendExpression()
		case ruleAction12:
			p.appendExpression()
		case ruleAction13:
			p.addOperatorLiteral("+")
		case ruleAction14:
			p.addOperatorLiteral("-")
		case ruleAction15:
			p.addOperatorFunction()
		case ruleAction16:
			p.addOperatorLiteral("/")
		case ruleAction17:
			p.addOperatorLiteral("*")
		case ruleAction18:
			p.addOperatorFunction()
		case ruleAction19:
			p.addNumberNode(buffer[begin:end])
		case ruleAction20:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction21:
			p.addGroupBy()
		case ruleAction22:

			p.addFunctionInvocation()

		case ruleAction23:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction24:
			p.addNullPredicate()
		case ruleAction25:

			p.addMetricExpression()

		case ruleAction26:

			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))

		case ruleAction27:

			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))

		case ruleAction28:
			p.addOrPredicate()
		case ruleAction29:
			p.addAndPredicate()
		case ruleAction30:
			p.addNotPredicate()
		case ruleAction31:

			p.addLiteralMatcher()

		case ruleAction32:

			p.addLiteralMatcher()
			p.addNotPredicate()

		case ruleAction33:

			p.addRegexMatcher()

		case ruleAction34:

			p.addListMatcher()

		case ruleAction35:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction36:
			p.addLiteralList()
		case ruleAction37:

			p.appendLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction38:
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
						if !_rules[rule_]() {
							goto l3
						}
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
							{
								add(ruleAction4, position)
							}
						l19:
							{
								position20, tokenIndex20, depth20 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l20
								}
								if !_rules[rulePROPERTY_KEY]() {
									goto l20
								}
								{
									add(ruleAction5, position)
								}
								if !_rules[rule__]() {
									goto l20
								}
								{
									position22 := position
									depth++
									{
										position23 := position
										depth++
										{
											position24, tokenIndex24, depth24 := position, tokenIndex, depth
											{
												position26 := position
												depth++
												if !_rules[ruleNUMBER]() {
													goto l25
												}
												depth--
												add(rulePegText, position26)
											}
											goto l24
										l25:
											position, tokenIndex, depth = position24, tokenIndex24, depth24
											if !_rules[ruleSTRING]() {
												goto l20
											}
										}
									l24:
										depth--
										add(ruleTIMESTAMP, position23)
									}
									depth--
									add(rulePROPERTY_VALUE, position22)
								}
								{
									add(ruleAction6, position)
								}
								{
									add(ruleAction7, position)
								}
								goto l19
							l20:
								position, tokenIndex, depth = position20, tokenIndex20, depth20
							}
							{
								add(ruleAction8, position)
							}
							depth--
							add(rulepropertyClause, position17)
						}
						{
							add(ruleAction0, position)
						}
						if !_rules[rule_]() {
							goto l3
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
						if !_rules[rule_]() {
							goto l0
						}
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
						if !_rules[rule_]() {
							goto l0
						}
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
		/* 1 selectStmt <- <(_ (('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T')) __ expressionList optionalPredicateClause propertyClause Action0 _)> */
		nil,
		/* 2 describeStmt <- <(_ (('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E')) __ (describeAllStmt / describeSingleStmt) _)> */
		nil,
		/* 3 describeAllStmt <- <(('a' / 'A') ('l' / 'L') ('l' / 'L') Action1)> */
		nil,
		/* 4 describeSingleStmt <- <(<METRIC_NAME> Action2 optionalPredicateClause Action3)> */
		nil,
		/* 5 propertyClause <- <(Action4 (_ PROPERTY_KEY Action5 __ PROPERTY_VALUE Action6 Action7)* Action8)> */
		nil,
		/* 6 optionalPredicateClause <- <((_ predicateClause) / Action9)> */
		func() bool {
			{
				position70 := position
				depth++
				{
					position71, tokenIndex71, depth71 := position, tokenIndex, depth
					if !_rules[rule_]() {
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
						add(ruleAction9, position)
					}
				}
			l71:
				depth--
				add(ruleoptionalPredicateClause, position70)
			}
			return true
		},
		/* 7 expressionList <- <(Action10 expression_1 Action11 (COMMA expression_1 Action12)*)> */
		func() bool {
			position85, tokenIndex85, depth85 := position, tokenIndex, depth
			{
				position86 := position
				depth++
				{
					add(ruleAction10, position)
				}
				if !_rules[ruleexpression_1]() {
					goto l85
				}
				{
					add(ruleAction11, position)
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
						add(ruleAction12, position)
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
		/* 8 expression_1 <- <(expression_2 (((OP_ADD Action13) / (OP_SUB Action14)) expression_2 Action15)*)> */
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
							add(ruleAction13, position)
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
							add(ruleAction14, position)
						}
					}
				l96:
					if !_rules[ruleexpression_2]() {
						goto l95
					}
					{
						add(ruleAction15, position)
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
		/* 9 expression_2 <- <(expression_3 (((OP_DIV Action16) / (OP_MULT Action17)) expression_3 Action18)*)> */
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
							add(ruleAction16, position)
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
							add(ruleAction17, position)
						}
					}
				l107:
					if !_rules[ruleexpression_3]() {
						goto l106
					}
					{
						add(ruleAction18, position)
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
		/* 10 expression_3 <- <(expression_function / ((&('-' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') (<NUMBER> Action19)) | (&('\t' | '\n' | ' ' | '(') (PAREN_OPEN expression_1 PAREN_CLOSE)) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') expression_metric)))> */
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
							add(ruleAction20, position)
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l117
						}
						if !_rules[ruleexpressionList]() {
							goto l117
						}
						{
							add(ruleAction21, position)
						}
						{
							position122, tokenIndex122, depth122 := position, tokenIndex, depth
							if !_rules[rule__]() {
								goto l122
							}
							{
								position124 := position
								depth++
								{
									position125, tokenIndex125, depth125 := position, tokenIndex, depth
									if buffer[position] != rune('g') {
										goto l126
									}
									position++
									goto l125
								l126:
									position, tokenIndex, depth = position125, tokenIndex125, depth125
									if buffer[position] != rune('G') {
										goto l122
									}
									position++
								}
							l125:
								{
									position127, tokenIndex127, depth127 := position, tokenIndex, depth
									if buffer[position] != rune('r') {
										goto l128
									}
									position++
									goto l127
								l128:
									position, tokenIndex, depth = position127, tokenIndex127, depth127
									if buffer[position] != rune('R') {
										goto l122
									}
									position++
								}
							l127:
								{
									position129, tokenIndex129, depth129 := position, tokenIndex, depth
									if buffer[position] != rune('o') {
										goto l130
									}
									position++
									goto l129
								l130:
									position, tokenIndex, depth = position129, tokenIndex129, depth129
									if buffer[position] != rune('O') {
										goto l122
									}
									position++
								}
							l129:
								{
									position131, tokenIndex131, depth131 := position, tokenIndex, depth
									if buffer[position] != rune('u') {
										goto l132
									}
									position++
									goto l131
								l132:
									position, tokenIndex, depth = position131, tokenIndex131, depth131
									if buffer[position] != rune('U') {
										goto l122
									}
									position++
								}
							l131:
								{
									position133, tokenIndex133, depth133 := position, tokenIndex, depth
									if buffer[position] != rune('p') {
										goto l134
									}
									position++
									goto l133
								l134:
									position, tokenIndex, depth = position133, tokenIndex133, depth133
									if buffer[position] != rune('P') {
										goto l122
									}
									position++
								}
							l133:
								if !_rules[rule__]() {
									goto l122
								}
								{
									position135, tokenIndex135, depth135 := position, tokenIndex, depth
									if buffer[position] != rune('b') {
										goto l136
									}
									position++
									goto l135
								l136:
									position, tokenIndex, depth = position135, tokenIndex135, depth135
									if buffer[position] != rune('B') {
										goto l122
									}
									position++
								}
							l135:
								{
									position137, tokenIndex137, depth137 := position, tokenIndex, depth
									if buffer[position] != rune('y') {
										goto l138
									}
									position++
									goto l137
								l138:
									position, tokenIndex, depth = position137, tokenIndex137, depth137
									if buffer[position] != rune('Y') {
										goto l122
									}
									position++
								}
							l137:
								if !_rules[rule__]() {
									goto l122
								}
								{
									position139 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l122
									}
									depth--
									add(rulePegText, position139)
								}
								{
									add(ruleAction26, position)
								}
							l141:
								{
									position142, tokenIndex142, depth142 := position, tokenIndex, depth
									if !_rules[ruleCOMMA]() {
										goto l142
									}
									{
										position143 := position
										depth++
										if !_rules[ruleCOLUMN_NAME]() {
											goto l142
										}
										depth--
										add(rulePegText, position143)
									}
									{
										add(ruleAction27, position)
									}
									goto l141
								l142:
									position, tokenIndex, depth = position142, tokenIndex142, depth142
								}
								depth--
								add(rulegroupByClause, position124)
							}
							goto l123
						l122:
							position, tokenIndex, depth = position122, tokenIndex122, depth122
						}
					l123:
						if !_rules[rulePAREN_CLOSE]() {
							goto l117
						}
						{
							add(ruleAction22, position)
						}
						depth--
						add(ruleexpression_function, position118)
					}
					goto l116
				l117:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
					{
						switch buffer[position] {
						case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							{
								position147 := position
								depth++
								if !_rules[ruleNUMBER]() {
									goto l114
								}
								depth--
								add(rulePegText, position147)
							}
							{
								add(ruleAction19, position)
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
								position149 := position
								depth++
								{
									position150 := position
									depth++
									if !_rules[ruleIDENTIFIER]() {
										goto l114
									}
									depth--
									add(rulePegText, position150)
								}
								{
									add(ruleAction23, position)
								}
								{
									position152, tokenIndex152, depth152 := position, tokenIndex, depth
									{
										position154, tokenIndex154, depth154 := position, tokenIndex, depth
										if buffer[position] != rune('[') {
											goto l155
										}
										position++
										if !_rules[rule_]() {
											goto l155
										}
										if !_rules[rulepredicate_1]() {
											goto l155
										}
										if !_rules[rule_]() {
											goto l155
										}
										if buffer[position] != rune(']') {
											goto l155
										}
										position++
										goto l154
									l155:
										position, tokenIndex, depth = position154, tokenIndex154, depth154
										{
											add(ruleAction24, position)
										}
									}
								l154:
									goto l153

									position, tokenIndex, depth = position152, tokenIndex152, depth152
								}
							l153:
								{
									add(ruleAction25, position)
								}
								depth--
								add(ruleexpression_metric, position149)
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
		/* 11 expression_function <- <(<IDENTIFIER> Action20 PAREN_OPEN expressionList Action21 (__ groupByClause)? PAREN_CLOSE Action22)> */
		nil,
		/* 12 expression_metric <- <(<IDENTIFIER> Action23 (('[' _ predicate_1 _ ']') / Action24)? Action25)> */
		nil,
		/* 13 groupByClause <- <(('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P') __ (('b' / 'B') ('y' / 'Y')) __ <COLUMN_NAME> Action26 (COMMA <COLUMN_NAME> Action27)*)> */
		nil,
		/* 14 predicateClause <- <(('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E') __ predicate_1)> */
		nil,
		/* 15 predicate_1 <- <((predicate_2 OP_OR predicate_1 Action28) / predicate_2 / )> */
		func() bool {
			{
				position163 := position
				depth++
				{
					position164, tokenIndex164, depth164 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l165
					}
					{
						position166 := position
						depth++
						if !_rules[rule_]() {
							goto l165
						}
						{
							position167, tokenIndex167, depth167 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l168
							}
							position++
							goto l167
						l168:
							position, tokenIndex, depth = position167, tokenIndex167, depth167
							if buffer[position] != rune('O') {
								goto l165
							}
							position++
						}
					l167:
						{
							position169, tokenIndex169, depth169 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l170
							}
							position++
							goto l169
						l170:
							position, tokenIndex, depth = position169, tokenIndex169, depth169
							if buffer[position] != rune('R') {
								goto l165
							}
							position++
						}
					l169:
						if !_rules[rule_]() {
							goto l165
						}
						depth--
						add(ruleOP_OR, position166)
					}
					if !_rules[rulepredicate_1]() {
						goto l165
					}
					{
						add(ruleAction28, position)
					}
					goto l164
				l165:
					position, tokenIndex, depth = position164, tokenIndex164, depth164
					if !_rules[rulepredicate_2]() {
						goto l172
					}
					goto l164
				l172:
					position, tokenIndex, depth = position164, tokenIndex164, depth164
				}
			l164:
				depth--
				add(rulepredicate_1, position163)
			}
			return true
		},
		/* 16 predicate_2 <- <((predicate_3 OP_AND predicate_2 Action29) / predicate_3)> */
		func() bool {
			position173, tokenIndex173, depth173 := position, tokenIndex, depth
			{
				position174 := position
				depth++
				{
					position175, tokenIndex175, depth175 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l176
					}
					{
						position177 := position
						depth++
						if !_rules[rule_]() {
							goto l176
						}
						{
							position178, tokenIndex178, depth178 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l179
							}
							position++
							goto l178
						l179:
							position, tokenIndex, depth = position178, tokenIndex178, depth178
							if buffer[position] != rune('A') {
								goto l176
							}
							position++
						}
					l178:
						{
							position180, tokenIndex180, depth180 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l181
							}
							position++
							goto l180
						l181:
							position, tokenIndex, depth = position180, tokenIndex180, depth180
							if buffer[position] != rune('N') {
								goto l176
							}
							position++
						}
					l180:
						{
							position182, tokenIndex182, depth182 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l183
							}
							position++
							goto l182
						l183:
							position, tokenIndex, depth = position182, tokenIndex182, depth182
							if buffer[position] != rune('D') {
								goto l176
							}
							position++
						}
					l182:
						if !_rules[rule_]() {
							goto l176
						}
						depth--
						add(ruleOP_AND, position177)
					}
					if !_rules[rulepredicate_2]() {
						goto l176
					}
					{
						add(ruleAction29, position)
					}
					goto l175
				l176:
					position, tokenIndex, depth = position175, tokenIndex175, depth175
					if !_rules[rulepredicate_3]() {
						goto l173
					}
				}
			l175:
				depth--
				add(rulepredicate_2, position174)
			}
			return true
		l173:
			position, tokenIndex, depth = position173, tokenIndex173, depth173
			return false
		},
		/* 17 predicate_3 <- <((OP_NOT predicate_3 Action30) / (PAREN_OPEN predicate_1 PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position185, tokenIndex185, depth185 := position, tokenIndex, depth
			{
				position186 := position
				depth++
				{
					position187, tokenIndex187, depth187 := position, tokenIndex, depth
					{
						position189 := position
						depth++
						{
							position190, tokenIndex190, depth190 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l191
							}
							position++
							goto l190
						l191:
							position, tokenIndex, depth = position190, tokenIndex190, depth190
							if buffer[position] != rune('N') {
								goto l188
							}
							position++
						}
					l190:
						{
							position192, tokenIndex192, depth192 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l193
							}
							position++
							goto l192
						l193:
							position, tokenIndex, depth = position192, tokenIndex192, depth192
							if buffer[position] != rune('O') {
								goto l188
							}
							position++
						}
					l192:
						{
							position194, tokenIndex194, depth194 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l195
							}
							position++
							goto l194
						l195:
							position, tokenIndex, depth = position194, tokenIndex194, depth194
							if buffer[position] != rune('T') {
								goto l188
							}
							position++
						}
					l194:
						if !_rules[rule__]() {
							goto l188
						}
						depth--
						add(ruleOP_NOT, position189)
					}
					if !_rules[rulepredicate_3]() {
						goto l188
					}
					{
						add(ruleAction30, position)
					}
					goto l187
				l188:
					position, tokenIndex, depth = position187, tokenIndex187, depth187
					if !_rules[rulePAREN_OPEN]() {
						goto l197
					}
					if !_rules[rulepredicate_1]() {
						goto l197
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l197
					}
					goto l187
				l197:
					position, tokenIndex, depth = position187, tokenIndex187, depth187
					{
						position198 := position
						depth++
						{
							position199, tokenIndex199, depth199 := position, tokenIndex, depth
							if !_rules[ruletagName]() {
								goto l200
							}
							if !_rules[rule_]() {
								goto l200
							}
							if buffer[position] != rune('=') {
								goto l200
							}
							position++
							if !_rules[rule_]() {
								goto l200
							}
							if !_rules[ruleliteralString]() {
								goto l200
							}
							{
								add(ruleAction31, position)
							}
							goto l199
						l200:
							position, tokenIndex, depth = position199, tokenIndex199, depth199
							if !_rules[ruletagName]() {
								goto l202
							}
							if !_rules[rule_]() {
								goto l202
							}
							if buffer[position] != rune('!') {
								goto l202
							}
							position++
							if buffer[position] != rune('=') {
								goto l202
							}
							position++
							if !_rules[rule_]() {
								goto l202
							}
							if !_rules[ruleliteralString]() {
								goto l202
							}
							{
								add(ruleAction32, position)
							}
							goto l199
						l202:
							position, tokenIndex, depth = position199, tokenIndex199, depth199
							if !_rules[ruletagName]() {
								goto l204
							}
							if !_rules[rule__]() {
								goto l204
							}
							{
								position205, tokenIndex205, depth205 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l206
								}
								position++
								goto l205
							l206:
								position, tokenIndex, depth = position205, tokenIndex205, depth205
								if buffer[position] != rune('M') {
									goto l204
								}
								position++
							}
						l205:
							{
								position207, tokenIndex207, depth207 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l208
								}
								position++
								goto l207
							l208:
								position, tokenIndex, depth = position207, tokenIndex207, depth207
								if buffer[position] != rune('A') {
									goto l204
								}
								position++
							}
						l207:
							{
								position209, tokenIndex209, depth209 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l210
								}
								position++
								goto l209
							l210:
								position, tokenIndex, depth = position209, tokenIndex209, depth209
								if buffer[position] != rune('T') {
									goto l204
								}
								position++
							}
						l209:
							{
								position211, tokenIndex211, depth211 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l212
								}
								position++
								goto l211
							l212:
								position, tokenIndex, depth = position211, tokenIndex211, depth211
								if buffer[position] != rune('C') {
									goto l204
								}
								position++
							}
						l211:
							{
								position213, tokenIndex213, depth213 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l214
								}
								position++
								goto l213
							l214:
								position, tokenIndex, depth = position213, tokenIndex213, depth213
								if buffer[position] != rune('H') {
									goto l204
								}
								position++
							}
						l213:
							{
								position215, tokenIndex215, depth215 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l216
								}
								position++
								goto l215
							l216:
								position, tokenIndex, depth = position215, tokenIndex215, depth215
								if buffer[position] != rune('E') {
									goto l204
								}
								position++
							}
						l215:
							{
								position217, tokenIndex217, depth217 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l218
								}
								position++
								goto l217
							l218:
								position, tokenIndex, depth = position217, tokenIndex217, depth217
								if buffer[position] != rune('S') {
									goto l204
								}
								position++
							}
						l217:
							if !_rules[rule__]() {
								goto l204
							}
							if !_rules[ruleliteralString]() {
								goto l204
							}
							{
								add(ruleAction33, position)
							}
							goto l199
						l204:
							position, tokenIndex, depth = position199, tokenIndex199, depth199
							if !_rules[ruletagName]() {
								goto l185
							}
							if !_rules[rule__]() {
								goto l185
							}
							{
								position220, tokenIndex220, depth220 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l221
								}
								position++
								goto l220
							l221:
								position, tokenIndex, depth = position220, tokenIndex220, depth220
								if buffer[position] != rune('I') {
									goto l185
								}
								position++
							}
						l220:
							{
								position222, tokenIndex222, depth222 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l223
								}
								position++
								goto l222
							l223:
								position, tokenIndex, depth = position222, tokenIndex222, depth222
								if buffer[position] != rune('N') {
									goto l185
								}
								position++
							}
						l222:
							if !_rules[rule__]() {
								goto l185
							}
							{
								position224 := position
								depth++
								{
									add(ruleAction36, position)
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l185
								}
								if !_rules[ruleliteralListString]() {
									goto l185
								}
							l226:
								{
									position227, tokenIndex227, depth227 := position, tokenIndex, depth
									if !_rules[ruleCOMMA]() {
										goto l227
									}
									if !_rules[ruleliteralListString]() {
										goto l227
									}
									goto l226
								l227:
									position, tokenIndex, depth = position227, tokenIndex227, depth227
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l185
								}
								depth--
								add(ruleliteralList, position224)
							}
							{
								add(ruleAction34, position)
							}
						}
					l199:
						depth--
						add(ruletagMatcher, position198)
					}
				}
			l187:
				depth--
				add(rulepredicate_3, position186)
			}
			return true
		l185:
			position, tokenIndex, depth = position185, tokenIndex185, depth185
			return false
		},
		/* 18 tagMatcher <- <((tagName _ '=' _ literalString Action31) / (tagName _ ('!' '=') _ literalString Action32) / (tagName __ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) __ literalString Action33) / (tagName __ (('i' / 'I') ('n' / 'N')) __ literalList Action34))> */
		nil,
		/* 19 literalString <- <(STRING Action35)> */
		func() bool {
			position230, tokenIndex230, depth230 := position, tokenIndex, depth
			{
				position231 := position
				depth++
				if !_rules[ruleSTRING]() {
					goto l230
				}
				{
					add(ruleAction35, position)
				}
				depth--
				add(ruleliteralString, position231)
			}
			return true
		l230:
			position, tokenIndex, depth = position230, tokenIndex230, depth230
			return false
		},
		/* 20 literalList <- <(Action36 PAREN_OPEN literalListString (COMMA literalListString)* PAREN_CLOSE)> */
		nil,
		/* 21 literalListString <- <(STRING Action37)> */
		func() bool {
			position234, tokenIndex234, depth234 := position, tokenIndex, depth
			{
				position235 := position
				depth++
				if !_rules[ruleSTRING]() {
					goto l234
				}
				{
					add(ruleAction37, position)
				}
				depth--
				add(ruleliteralListString, position235)
			}
			return true
		l234:
			position, tokenIndex, depth = position234, tokenIndex234, depth234
			return false
		},
		/* 22 tagName <- <(<TAG_NAME> Action38)> */
		func() bool {
			position237, tokenIndex237, depth237 := position, tokenIndex, depth
			{
				position238 := position
				depth++
				{
					position239 := position
					depth++
					{
						position240 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l237
						}
						depth--
						add(ruleTAG_NAME, position240)
					}
					depth--
					add(rulePegText, position239)
				}
				{
					add(ruleAction38, position)
				}
				depth--
				add(ruletagName, position238)
			}
			return true
		l237:
			position, tokenIndex, depth = position237, tokenIndex237, depth237
			return false
		},
		/* 23 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position242, tokenIndex242, depth242 := position, tokenIndex, depth
			{
				position243 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l242
				}
				depth--
				add(ruleCOLUMN_NAME, position243)
			}
			return true
		l242:
			position, tokenIndex, depth = position242, tokenIndex242, depth242
			return false
		},
		/* 24 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 25 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 26 IDENTIFIER <- <(('`' CHAR* '`') / (!(KEYWORD !ID_CONT) ID_SEGMENT ('.' ID_SEGMENT)*))> */
		func() bool {
			position246, tokenIndex246, depth246 := position, tokenIndex, depth
			{
				position247 := position
				depth++
				{
					position248, tokenIndex248, depth248 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l249
					}
					position++
				l250:
					{
						position251, tokenIndex251, depth251 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l251
						}
						goto l250
					l251:
						position, tokenIndex, depth = position251, tokenIndex251, depth251
					}
					if buffer[position] != rune('`') {
						goto l249
					}
					position++
					goto l248
				l249:
					position, tokenIndex, depth = position248, tokenIndex248, depth248
					{
						position252, tokenIndex252, depth252 := position, tokenIndex, depth
						{
							position253 := position
							depth++
							{
								position254, tokenIndex254, depth254 := position, tokenIndex, depth
								{
									position256, tokenIndex256, depth256 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l257
									}
									position++
									goto l256
								l257:
									position, tokenIndex, depth = position256, tokenIndex256, depth256
									if buffer[position] != rune('A') {
										goto l255
									}
									position++
								}
							l256:
								{
									position258, tokenIndex258, depth258 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l259
									}
									position++
									goto l258
								l259:
									position, tokenIndex, depth = position258, tokenIndex258, depth258
									if buffer[position] != rune('L') {
										goto l255
									}
									position++
								}
							l258:
								{
									position260, tokenIndex260, depth260 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l261
									}
									position++
									goto l260
								l261:
									position, tokenIndex, depth = position260, tokenIndex260, depth260
									if buffer[position] != rune('L') {
										goto l255
									}
									position++
								}
							l260:
								goto l254
							l255:
								position, tokenIndex, depth = position254, tokenIndex254, depth254
								{
									position263, tokenIndex263, depth263 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l264
									}
									position++
									goto l263
								l264:
									position, tokenIndex, depth = position263, tokenIndex263, depth263
									if buffer[position] != rune('A') {
										goto l262
									}
									position++
								}
							l263:
								{
									position265, tokenIndex265, depth265 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l266
									}
									position++
									goto l265
								l266:
									position, tokenIndex, depth = position265, tokenIndex265, depth265
									if buffer[position] != rune('N') {
										goto l262
									}
									position++
								}
							l265:
								{
									position267, tokenIndex267, depth267 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l268
									}
									position++
									goto l267
								l268:
									position, tokenIndex, depth = position267, tokenIndex267, depth267
									if buffer[position] != rune('D') {
										goto l262
									}
									position++
								}
							l267:
								goto l254
							l262:
								position, tokenIndex, depth = position254, tokenIndex254, depth254
								{
									position270, tokenIndex270, depth270 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l271
									}
									position++
									goto l270
								l271:
									position, tokenIndex, depth = position270, tokenIndex270, depth270
									if buffer[position] != rune('S') {
										goto l269
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
										goto l269
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
										goto l269
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
										goto l269
									}
									position++
								}
							l276:
								{
									position278, tokenIndex278, depth278 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l279
									}
									position++
									goto l278
								l279:
									position, tokenIndex, depth = position278, tokenIndex278, depth278
									if buffer[position] != rune('C') {
										goto l269
									}
									position++
								}
							l278:
								{
									position280, tokenIndex280, depth280 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l281
									}
									position++
									goto l280
								l281:
									position, tokenIndex, depth = position280, tokenIndex280, depth280
									if buffer[position] != rune('T') {
										goto l269
									}
									position++
								}
							l280:
								goto l254
							l269:
								position, tokenIndex, depth = position254, tokenIndex254, depth254
								{
									switch buffer[position] {
									case 'W', 'w':
										{
											position283, tokenIndex283, depth283 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l284
											}
											position++
											goto l283
										l284:
											position, tokenIndex, depth = position283, tokenIndex283, depth283
											if buffer[position] != rune('W') {
												goto l252
											}
											position++
										}
									l283:
										{
											position285, tokenIndex285, depth285 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l286
											}
											position++
											goto l285
										l286:
											position, tokenIndex, depth = position285, tokenIndex285, depth285
											if buffer[position] != rune('H') {
												goto l252
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
												goto l252
											}
											position++
										}
									l287:
										{
											position289, tokenIndex289, depth289 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l290
											}
											position++
											goto l289
										l290:
											position, tokenIndex, depth = position289, tokenIndex289, depth289
											if buffer[position] != rune('R') {
												goto l252
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
												goto l252
											}
											position++
										}
									l291:
										break
									case 'O', 'o':
										{
											position293, tokenIndex293, depth293 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l294
											}
											position++
											goto l293
										l294:
											position, tokenIndex, depth = position293, tokenIndex293, depth293
											if buffer[position] != rune('O') {
												goto l252
											}
											position++
										}
									l293:
										{
											position295, tokenIndex295, depth295 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l296
											}
											position++
											goto l295
										l296:
											position, tokenIndex, depth = position295, tokenIndex295, depth295
											if buffer[position] != rune('R') {
												goto l252
											}
											position++
										}
									l295:
										break
									case 'N', 'n':
										{
											position297, tokenIndex297, depth297 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l298
											}
											position++
											goto l297
										l298:
											position, tokenIndex, depth = position297, tokenIndex297, depth297
											if buffer[position] != rune('N') {
												goto l252
											}
											position++
										}
									l297:
										{
											position299, tokenIndex299, depth299 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l300
											}
											position++
											goto l299
										l300:
											position, tokenIndex, depth = position299, tokenIndex299, depth299
											if buffer[position] != rune('O') {
												goto l252
											}
											position++
										}
									l299:
										{
											position301, tokenIndex301, depth301 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l302
											}
											position++
											goto l301
										l302:
											position, tokenIndex, depth = position301, tokenIndex301, depth301
											if buffer[position] != rune('T') {
												goto l252
											}
											position++
										}
									l301:
										break
									case 'M', 'm':
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
												goto l252
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
												goto l252
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
												goto l252
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
												goto l252
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
												goto l252
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
												goto l252
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
												goto l252
											}
											position++
										}
									l315:
										break
									case 'I', 'i':
										{
											position317, tokenIndex317, depth317 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l318
											}
											position++
											goto l317
										l318:
											position, tokenIndex, depth = position317, tokenIndex317, depth317
											if buffer[position] != rune('I') {
												goto l252
											}
											position++
										}
									l317:
										{
											position319, tokenIndex319, depth319 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l320
											}
											position++
											goto l319
										l320:
											position, tokenIndex, depth = position319, tokenIndex319, depth319
											if buffer[position] != rune('N') {
												goto l252
											}
											position++
										}
									l319:
										break
									case 'G', 'g':
										{
											position321, tokenIndex321, depth321 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l322
											}
											position++
											goto l321
										l322:
											position, tokenIndex, depth = position321, tokenIndex321, depth321
											if buffer[position] != rune('G') {
												goto l252
											}
											position++
										}
									l321:
										{
											position323, tokenIndex323, depth323 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l324
											}
											position++
											goto l323
										l324:
											position, tokenIndex, depth = position323, tokenIndex323, depth323
											if buffer[position] != rune('R') {
												goto l252
											}
											position++
										}
									l323:
										{
											position325, tokenIndex325, depth325 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l326
											}
											position++
											goto l325
										l326:
											position, tokenIndex, depth = position325, tokenIndex325, depth325
											if buffer[position] != rune('O') {
												goto l252
											}
											position++
										}
									l325:
										{
											position327, tokenIndex327, depth327 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l328
											}
											position++
											goto l327
										l328:
											position, tokenIndex, depth = position327, tokenIndex327, depth327
											if buffer[position] != rune('U') {
												goto l252
											}
											position++
										}
									l327:
										{
											position329, tokenIndex329, depth329 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l330
											}
											position++
											goto l329
										l330:
											position, tokenIndex, depth = position329, tokenIndex329, depth329
											if buffer[position] != rune('P') {
												goto l252
											}
											position++
										}
									l329:
										break
									case 'D', 'd':
										{
											position331, tokenIndex331, depth331 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l332
											}
											position++
											goto l331
										l332:
											position, tokenIndex, depth = position331, tokenIndex331, depth331
											if buffer[position] != rune('D') {
												goto l252
											}
											position++
										}
									l331:
										{
											position333, tokenIndex333, depth333 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l334
											}
											position++
											goto l333
										l334:
											position, tokenIndex, depth = position333, tokenIndex333, depth333
											if buffer[position] != rune('E') {
												goto l252
											}
											position++
										}
									l333:
										{
											position335, tokenIndex335, depth335 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l336
											}
											position++
											goto l335
										l336:
											position, tokenIndex, depth = position335, tokenIndex335, depth335
											if buffer[position] != rune('S') {
												goto l252
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
												goto l252
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
												goto l252
											}
											position++
										}
									l339:
										{
											position341, tokenIndex341, depth341 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l342
											}
											position++
											goto l341
										l342:
											position, tokenIndex, depth = position341, tokenIndex341, depth341
											if buffer[position] != rune('I') {
												goto l252
											}
											position++
										}
									l341:
										{
											position343, tokenIndex343, depth343 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l344
											}
											position++
											goto l343
										l344:
											position, tokenIndex, depth = position343, tokenIndex343, depth343
											if buffer[position] != rune('B') {
												goto l252
											}
											position++
										}
									l343:
										{
											position345, tokenIndex345, depth345 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l346
											}
											position++
											goto l345
										l346:
											position, tokenIndex, depth = position345, tokenIndex345, depth345
											if buffer[position] != rune('E') {
												goto l252
											}
											position++
										}
									l345:
										break
									case 'B', 'b':
										{
											position347, tokenIndex347, depth347 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l348
											}
											position++
											goto l347
										l348:
											position, tokenIndex, depth = position347, tokenIndex347, depth347
											if buffer[position] != rune('B') {
												goto l252
											}
											position++
										}
									l347:
										{
											position349, tokenIndex349, depth349 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l350
											}
											position++
											goto l349
										l350:
											position, tokenIndex, depth = position349, tokenIndex349, depth349
											if buffer[position] != rune('Y') {
												goto l252
											}
											position++
										}
									l349:
										break
									case 'A', 'a':
										{
											position351, tokenIndex351, depth351 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l352
											}
											position++
											goto l351
										l352:
											position, tokenIndex, depth = position351, tokenIndex351, depth351
											if buffer[position] != rune('A') {
												goto l252
											}
											position++
										}
									l351:
										{
											position353, tokenIndex353, depth353 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l354
											}
											position++
											goto l353
										l354:
											position, tokenIndex, depth = position353, tokenIndex353, depth353
											if buffer[position] != rune('S') {
												goto l252
											}
											position++
										}
									l353:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l252
										}
										break
									}
								}

							}
						l254:
							depth--
							add(ruleKEYWORD, position253)
						}
						{
							position355, tokenIndex355, depth355 := position, tokenIndex, depth
							if !_rules[ruleID_CONT]() {
								goto l355
							}
							goto l252
						l355:
							position, tokenIndex, depth = position355, tokenIndex355, depth355
						}
						goto l246
					l252:
						position, tokenIndex, depth = position252, tokenIndex252, depth252
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l246
					}
				l356:
					{
						position357, tokenIndex357, depth357 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l357
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l357
						}
						goto l356
					l357:
						position, tokenIndex, depth = position357, tokenIndex357, depth357
					}
				}
			l248:
				depth--
				add(ruleIDENTIFIER, position247)
			}
			return true
		l246:
			position, tokenIndex, depth = position246, tokenIndex246, depth246
			return false
		},
		/* 27 TIMESTAMP <- <(<NUMBER> / STRING)> */
		nil,
		/* 28 ID_SEGMENT <- <(ID_START ID_CONT*)> */
		func() bool {
			position359, tokenIndex359, depth359 := position, tokenIndex, depth
			{
				position360 := position
				depth++
				if !_rules[ruleID_START]() {
					goto l359
				}
			l361:
				{
					position362, tokenIndex362, depth362 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l362
					}
					goto l361
				l362:
					position, tokenIndex, depth = position362, tokenIndex362, depth362
				}
				depth--
				add(ruleID_SEGMENT, position360)
			}
			return true
		l359:
			position, tokenIndex, depth = position359, tokenIndex359, depth359
			return false
		},
		/* 29 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position363, tokenIndex363, depth363 := position, tokenIndex, depth
			{
				position364 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l363
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l363
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l363
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position364)
			}
			return true
		l363:
			position, tokenIndex, depth = position363, tokenIndex363, depth363
			return false
		},
		/* 30 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position366, tokenIndex366, depth366 := position, tokenIndex, depth
			{
				position367 := position
				depth++
				{
					position368, tokenIndex368, depth368 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l369
					}
					goto l368
				l369:
					position, tokenIndex, depth = position368, tokenIndex368, depth368
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l366
					}
					position++
				}
			l368:
				depth--
				add(ruleID_CONT, position367)
			}
			return true
		l366:
			position, tokenIndex, depth = position366, tokenIndex366, depth366
			return false
		},
		/* 31 PROPERTY_KEY <- <((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> __ (('b' / 'B') ('y' / 'Y')))) | (&('R' | 'r') <(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))>) | (&('T' | 't') <(('t' / 'T') ('o' / 'O'))>) | (&('F' | 'f') <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))>))> */
		func() bool {
			position370, tokenIndex370, depth370 := position, tokenIndex, depth
			{
				position371 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position373 := position
							depth++
							{
								position374, tokenIndex374, depth374 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l375
								}
								position++
								goto l374
							l375:
								position, tokenIndex, depth = position374, tokenIndex374, depth374
								if buffer[position] != rune('S') {
									goto l370
								}
								position++
							}
						l374:
							{
								position376, tokenIndex376, depth376 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l377
								}
								position++
								goto l376
							l377:
								position, tokenIndex, depth = position376, tokenIndex376, depth376
								if buffer[position] != rune('A') {
									goto l370
								}
								position++
							}
						l376:
							{
								position378, tokenIndex378, depth378 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l379
								}
								position++
								goto l378
							l379:
								position, tokenIndex, depth = position378, tokenIndex378, depth378
								if buffer[position] != rune('M') {
									goto l370
								}
								position++
							}
						l378:
							{
								position380, tokenIndex380, depth380 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l381
								}
								position++
								goto l380
							l381:
								position, tokenIndex, depth = position380, tokenIndex380, depth380
								if buffer[position] != rune('P') {
									goto l370
								}
								position++
							}
						l380:
							{
								position382, tokenIndex382, depth382 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l383
								}
								position++
								goto l382
							l383:
								position, tokenIndex, depth = position382, tokenIndex382, depth382
								if buffer[position] != rune('L') {
									goto l370
								}
								position++
							}
						l382:
							{
								position384, tokenIndex384, depth384 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l385
								}
								position++
								goto l384
							l385:
								position, tokenIndex, depth = position384, tokenIndex384, depth384
								if buffer[position] != rune('E') {
									goto l370
								}
								position++
							}
						l384:
							depth--
							add(rulePegText, position373)
						}
						if !_rules[rule__]() {
							goto l370
						}
						{
							position386, tokenIndex386, depth386 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l387
							}
							position++
							goto l386
						l387:
							position, tokenIndex, depth = position386, tokenIndex386, depth386
							if buffer[position] != rune('B') {
								goto l370
							}
							position++
						}
					l386:
						{
							position388, tokenIndex388, depth388 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l389
							}
							position++
							goto l388
						l389:
							position, tokenIndex, depth = position388, tokenIndex388, depth388
							if buffer[position] != rune('Y') {
								goto l370
							}
							position++
						}
					l388:
						break
					case 'R', 'r':
						{
							position390 := position
							depth++
							{
								position391, tokenIndex391, depth391 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l392
								}
								position++
								goto l391
							l392:
								position, tokenIndex, depth = position391, tokenIndex391, depth391
								if buffer[position] != rune('R') {
									goto l370
								}
								position++
							}
						l391:
							{
								position393, tokenIndex393, depth393 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l394
								}
								position++
								goto l393
							l394:
								position, tokenIndex, depth = position393, tokenIndex393, depth393
								if buffer[position] != rune('E') {
									goto l370
								}
								position++
							}
						l393:
							{
								position395, tokenIndex395, depth395 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l396
								}
								position++
								goto l395
							l396:
								position, tokenIndex, depth = position395, tokenIndex395, depth395
								if buffer[position] != rune('S') {
									goto l370
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
									goto l370
								}
								position++
							}
						l397:
							{
								position399, tokenIndex399, depth399 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l400
								}
								position++
								goto l399
							l400:
								position, tokenIndex, depth = position399, tokenIndex399, depth399
								if buffer[position] != rune('L') {
									goto l370
								}
								position++
							}
						l399:
							{
								position401, tokenIndex401, depth401 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l402
								}
								position++
								goto l401
							l402:
								position, tokenIndex, depth = position401, tokenIndex401, depth401
								if buffer[position] != rune('U') {
									goto l370
								}
								position++
							}
						l401:
							{
								position403, tokenIndex403, depth403 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l404
								}
								position++
								goto l403
							l404:
								position, tokenIndex, depth = position403, tokenIndex403, depth403
								if buffer[position] != rune('T') {
									goto l370
								}
								position++
							}
						l403:
							{
								position405, tokenIndex405, depth405 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l406
								}
								position++
								goto l405
							l406:
								position, tokenIndex, depth = position405, tokenIndex405, depth405
								if buffer[position] != rune('I') {
									goto l370
								}
								position++
							}
						l405:
							{
								position407, tokenIndex407, depth407 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l408
								}
								position++
								goto l407
							l408:
								position, tokenIndex, depth = position407, tokenIndex407, depth407
								if buffer[position] != rune('O') {
									goto l370
								}
								position++
							}
						l407:
							{
								position409, tokenIndex409, depth409 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l410
								}
								position++
								goto l409
							l410:
								position, tokenIndex, depth = position409, tokenIndex409, depth409
								if buffer[position] != rune('N') {
									goto l370
								}
								position++
							}
						l409:
							depth--
							add(rulePegText, position390)
						}
						break
					case 'T', 't':
						{
							position411 := position
							depth++
							{
								position412, tokenIndex412, depth412 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l413
								}
								position++
								goto l412
							l413:
								position, tokenIndex, depth = position412, tokenIndex412, depth412
								if buffer[position] != rune('T') {
									goto l370
								}
								position++
							}
						l412:
							{
								position414, tokenIndex414, depth414 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l415
								}
								position++
								goto l414
							l415:
								position, tokenIndex, depth = position414, tokenIndex414, depth414
								if buffer[position] != rune('O') {
									goto l370
								}
								position++
							}
						l414:
							depth--
							add(rulePegText, position411)
						}
						break
					default:
						{
							position416 := position
							depth++
							{
								position417, tokenIndex417, depth417 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l418
								}
								position++
								goto l417
							l418:
								position, tokenIndex, depth = position417, tokenIndex417, depth417
								if buffer[position] != rune('F') {
									goto l370
								}
								position++
							}
						l417:
							{
								position419, tokenIndex419, depth419 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l420
								}
								position++
								goto l419
							l420:
								position, tokenIndex, depth = position419, tokenIndex419, depth419
								if buffer[position] != rune('R') {
									goto l370
								}
								position++
							}
						l419:
							{
								position421, tokenIndex421, depth421 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l422
								}
								position++
								goto l421
							l422:
								position, tokenIndex, depth = position421, tokenIndex421, depth421
								if buffer[position] != rune('O') {
									goto l370
								}
								position++
							}
						l421:
							{
								position423, tokenIndex423, depth423 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l424
								}
								position++
								goto l423
							l424:
								position, tokenIndex, depth = position423, tokenIndex423, depth423
								if buffer[position] != rune('M') {
									goto l370
								}
								position++
							}
						l423:
							depth--
							add(rulePegText, position416)
						}
						break
					}
				}

				depth--
				add(rulePROPERTY_KEY, position371)
			}
			return true
		l370:
			position, tokenIndex, depth = position370, tokenIndex370, depth370
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
		/* 41 QUOTE_SINGLE <- <'\''> */
		func() bool {
			position434, tokenIndex434, depth434 := position, tokenIndex, depth
			{
				position435 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l434
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position435)
			}
			return true
		l434:
			position, tokenIndex, depth = position434, tokenIndex434, depth434
			return false
		},
		/* 42 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position436, tokenIndex436, depth436 := position, tokenIndex, depth
			{
				position437 := position
				depth++
				if buffer[position] != rune('"') {
					goto l436
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position437)
			}
			return true
		l436:
			position, tokenIndex, depth = position436, tokenIndex436, depth436
			return false
		},
		/* 43 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> QUOTE_SINGLE) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> QUOTE_DOUBLE))> */
		func() bool {
			position438, tokenIndex438, depth438 := position, tokenIndex, depth
			{
				position439 := position
				depth++
				{
					position440, tokenIndex440, depth440 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l441
					}
					{
						position442 := position
						depth++
					l443:
						{
							position444, tokenIndex444, depth444 := position, tokenIndex, depth
							{
								position445, tokenIndex445, depth445 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l445
								}
								goto l444
							l445:
								position, tokenIndex, depth = position445, tokenIndex445, depth445
							}
							if !_rules[ruleCHAR]() {
								goto l444
							}
							goto l443
						l444:
							position, tokenIndex, depth = position444, tokenIndex444, depth444
						}
						depth--
						add(rulePegText, position442)
					}
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l441
					}
					goto l440
				l441:
					position, tokenIndex, depth = position440, tokenIndex440, depth440
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l438
					}
					{
						position446 := position
						depth++
					l447:
						{
							position448, tokenIndex448, depth448 := position, tokenIndex, depth
							{
								position449, tokenIndex449, depth449 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l449
								}
								goto l448
							l449:
								position, tokenIndex, depth = position449, tokenIndex449, depth449
							}
							if !_rules[ruleCHAR]() {
								goto l448
							}
							goto l447
						l448:
							position, tokenIndex, depth = position448, tokenIndex448, depth448
						}
						depth--
						add(rulePegText, position446)
					}
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l438
					}
				}
			l440:
				depth--
				add(ruleSTRING, position439)
			}
			return true
		l438:
			position, tokenIndex, depth = position438, tokenIndex438, depth438
			return false
		},
		/* 44 CHAR <- <(('\\' ((&('"') QUOTE_DOUBLE) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position450, tokenIndex450, depth450 := position, tokenIndex, depth
			{
				position451 := position
				depth++
				{
					position452, tokenIndex452, depth452 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l453
					}
					position++
					{
						switch buffer[position] {
						case '"':
							if !_rules[ruleQUOTE_DOUBLE]() {
								goto l453
							}
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l453
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l453
							}
							break
						}
					}

					goto l452
				l453:
					position, tokenIndex, depth = position452, tokenIndex452, depth452
					{
						position455, tokenIndex455, depth455 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l455
						}
						goto l450
					l455:
						position, tokenIndex, depth = position455, tokenIndex455, depth455
					}
					if !matchDot() {
						goto l450
					}
				}
			l452:
				depth--
				add(ruleCHAR, position451)
			}
			return true
		l450:
			position, tokenIndex, depth = position450, tokenIndex450, depth450
			return false
		},
		/* 45 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position456, tokenIndex456, depth456 := position, tokenIndex, depth
			{
				position457 := position
				depth++
				{
					position458, tokenIndex458, depth458 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l459
					}
					position++
					goto l458
				l459:
					position, tokenIndex, depth = position458, tokenIndex458, depth458
					if buffer[position] != rune('\\') {
						goto l456
					}
					position++
				}
			l458:
				depth--
				add(ruleESCAPE_CLASS, position457)
			}
			return true
		l456:
			position, tokenIndex, depth = position456, tokenIndex456, depth456
			return false
		},
		/* 46 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position460, tokenIndex460, depth460 := position, tokenIndex, depth
			{
				position461 := position
				depth++
				{
					position462 := position
					depth++
					{
						position463, tokenIndex463, depth463 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l463
						}
						position++
						goto l464
					l463:
						position, tokenIndex, depth = position463, tokenIndex463, depth463
					}
				l464:
					{
						position465 := position
						depth++
						{
							position466, tokenIndex466, depth466 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l467
							}
							position++
							goto l466
						l467:
							position, tokenIndex, depth = position466, tokenIndex466, depth466
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l460
							}
							position++
						l468:
							{
								position469, tokenIndex469, depth469 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l469
								}
								position++
								goto l468
							l469:
								position, tokenIndex, depth = position469, tokenIndex469, depth469
							}
						}
					l466:
						depth--
						add(ruleNUMBER_NATURAL, position465)
					}
					depth--
					add(ruleNUMBER_INTEGER, position462)
				}
				{
					position470, tokenIndex470, depth470 := position, tokenIndex, depth
					{
						position472 := position
						depth++
						if buffer[position] != rune('.') {
							goto l470
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l470
						}
						position++
					l473:
						{
							position474, tokenIndex474, depth474 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l474
							}
							position++
							goto l473
						l474:
							position, tokenIndex, depth = position474, tokenIndex474, depth474
						}
						depth--
						add(ruleNUMBER_FRACTION, position472)
					}
					goto l471
				l470:
					position, tokenIndex, depth = position470, tokenIndex470, depth470
				}
			l471:
				{
					position475, tokenIndex475, depth475 := position, tokenIndex, depth
					{
						position477 := position
						depth++
						{
							position478, tokenIndex478, depth478 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l479
							}
							position++
							goto l478
						l479:
							position, tokenIndex, depth = position478, tokenIndex478, depth478
							if buffer[position] != rune('E') {
								goto l475
							}
							position++
						}
					l478:
						{
							position480, tokenIndex480, depth480 := position, tokenIndex, depth
							{
								position482, tokenIndex482, depth482 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l483
								}
								position++
								goto l482
							l483:
								position, tokenIndex, depth = position482, tokenIndex482, depth482
								if buffer[position] != rune('-') {
									goto l480
								}
								position++
							}
						l482:
							goto l481
						l480:
							position, tokenIndex, depth = position480, tokenIndex480, depth480
						}
					l481:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l475
						}
						position++
					l484:
						{
							position485, tokenIndex485, depth485 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l485
							}
							position++
							goto l484
						l485:
							position, tokenIndex, depth = position485, tokenIndex485, depth485
						}
						depth--
						add(ruleNUMBER_EXP, position477)
					}
					goto l476
				l475:
					position, tokenIndex, depth = position475, tokenIndex475, depth475
				}
			l476:
				depth--
				add(ruleNUMBER, position461)
			}
			return true
		l460:
			position, tokenIndex, depth = position460, tokenIndex460, depth460
			return false
		},
		/* 47 NUMBER_NATURAL <- <('0' / ([1-9] [0-9]*))> */
		nil,
		/* 48 NUMBER_FRACTION <- <('.' [0-9]+)> */
		nil,
		/* 49 NUMBER_INTEGER <- <('-'? NUMBER_NATURAL)> */
		nil,
		/* 50 NUMBER_EXP <- <(('e' / 'E') ('+' / '-')? [0-9]+)> */
		nil,
		/* 51 PAREN_OPEN <- <(_ '(' _)> */
		func() bool {
			position490, tokenIndex490, depth490 := position, tokenIndex, depth
			{
				position491 := position
				depth++
				if !_rules[rule_]() {
					goto l490
				}
				if buffer[position] != rune('(') {
					goto l490
				}
				position++
				if !_rules[rule_]() {
					goto l490
				}
				depth--
				add(rulePAREN_OPEN, position491)
			}
			return true
		l490:
			position, tokenIndex, depth = position490, tokenIndex490, depth490
			return false
		},
		/* 52 PAREN_CLOSE <- <(_ ')' _)> */
		func() bool {
			position492, tokenIndex492, depth492 := position, tokenIndex, depth
			{
				position493 := position
				depth++
				if !_rules[rule_]() {
					goto l492
				}
				if buffer[position] != rune(')') {
					goto l492
				}
				position++
				if !_rules[rule_]() {
					goto l492
				}
				depth--
				add(rulePAREN_CLOSE, position493)
			}
			return true
		l492:
			position, tokenIndex, depth = position492, tokenIndex492, depth492
			return false
		},
		/* 53 COMMA <- <(_ ',' _)> */
		func() bool {
			position494, tokenIndex494, depth494 := position, tokenIndex, depth
			{
				position495 := position
				depth++
				if !_rules[rule_]() {
					goto l494
				}
				if buffer[position] != rune(',') {
					goto l494
				}
				position++
				if !_rules[rule_]() {
					goto l494
				}
				depth--
				add(ruleCOMMA, position495)
			}
			return true
		l494:
			position, tokenIndex, depth = position494, tokenIndex494, depth494
			return false
		},
		/* 54 _ <- <SPACE*> */
		func() bool {
			{
				position497 := position
				depth++
			l498:
				{
					position499, tokenIndex499, depth499 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l499
					}
					goto l498
				l499:
					position, tokenIndex, depth = position499, tokenIndex499, depth499
				}
				depth--
				add(rule_, position497)
			}
			return true
		},
		/* 55 __ <- <SPACE+> */
		func() bool {
			position500, tokenIndex500, depth500 := position, tokenIndex, depth
			{
				position501 := position
				depth++
				if !_rules[ruleSPACE]() {
					goto l500
				}
			l502:
				{
					position503, tokenIndex503, depth503 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l503
					}
					goto l502
				l503:
					position, tokenIndex, depth = position503, tokenIndex503, depth503
				}
				depth--
				add(rule__, position501)
			}
			return true
		l500:
			position, tokenIndex, depth = position500, tokenIndex500, depth500
			return false
		},
		/* 56 SPACE <- <((&('\t') '\t') | (&('\n') '\n') | (&(' ') ' '))> */
		func() bool {
			position504, tokenIndex504, depth504 := position, tokenIndex, depth
			{
				position505 := position
				depth++
				{
					switch buffer[position] {
					case '\t':
						if buffer[position] != rune('\t') {
							goto l504
						}
						position++
						break
					case '\n':
						if buffer[position] != rune('\n') {
							goto l504
						}
						position++
						break
					default:
						if buffer[position] != rune(' ') {
							goto l504
						}
						position++
						break
					}
				}

				depth--
				add(ruleSPACE, position505)
			}
			return true
		l504:
			position, tokenIndex, depth = position504, tokenIndex504, depth504
			return false
		},
		/* 58 Action0 <- <{
		   p.makeSelect()
		 }> */
		nil,
		/* 59 Action1 <- <{ p.makeDescribeAll() }> */
		nil,
		nil,
		/* 61 Action2 <- <{ p.addStringLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 62 Action3 <- <{ p.makeDescribe() }> */
		nil,
		/* 63 Action4 <- <{ p.addEvaluationContext() }> */
		nil,
		/* 64 Action5 <- <{ p.addPropertyKey(buffer[begin:end])   }> */
		nil,
		/* 65 Action6 <- <{ p.addPropertyValue(buffer[begin:end]) }> */
		nil,
		/* 66 Action7 <- <{ p.insertPropertyKeyValue() }> */
		nil,
		/* 67 Action8 <- <{ p.checkPropertyClause() }> */
		nil,
		/* 68 Action9 <- <{ p.addNullPredicate() }> */
		nil,
		/* 69 Action10 <- <{ p.addExpressionList() }> */
		nil,
		/* 70 Action11 <- <{ p.appendExpression() }> */
		nil,
		/* 71 Action12 <- <{ p.appendExpression() }> */
		nil,
		/* 72 Action13 <- <{ p.addOperatorLiteral("+") }> */
		nil,
		/* 73 Action14 <- <{ p.addOperatorLiteral("-") }> */
		nil,
		/* 74 Action15 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 75 Action16 <- <{ p.addOperatorLiteral("/") }> */
		nil,
		/* 76 Action17 <- <{ p.addOperatorLiteral("*") }> */
		nil,
		/* 77 Action18 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 78 Action19 <- <{ p.addNumberNode(buffer[begin:end]) }> */
		nil,
		/* 79 Action20 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 80 Action21 <- <{ p.addGroupBy() }> */
		nil,
		/* 81 Action22 <- <{
		   p.addFunctionInvocation()
		 }> */
		nil,
		/* 82 Action23 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 83 Action24 <- <{ p.addNullPredicate() }> */
		nil,
		/* 84 Action25 <- <{
		   p.addMetricExpression()
		 }> */
		nil,
		/* 85 Action26 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 86 Action27 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		   }> */
		nil,
		/* 87 Action28 <- <{ p.addOrPredicate() }> */
		nil,
		/* 88 Action29 <- <{ p.addAndPredicate() }> */
		nil,
		/* 89 Action30 <- <{ p.addNotPredicate() }> */
		nil,
		/* 90 Action31 <- <{
		   p.addLiteralMatcher()
		 }> */
		nil,
		/* 91 Action32 <- <{
		   p.addLiteralMatcher()
		   p.addNotPredicate()
		 }> */
		nil,
		/* 92 Action33 <- <{
		   p.addRegexMatcher()
		 }> */
		nil,
		/* 93 Action34 <- <{
		   p.addListMatcher()
		 }> */
		nil,
		/* 94 Action35 <- <{
		  p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 95 Action36 <- <{ p.addLiteralList() }> */
		nil,
		/* 96 Action37 <- <{
		  p.appendLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 97 Action38 <- <{ p.addTagLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
	}
	p.rules = _rules
}
