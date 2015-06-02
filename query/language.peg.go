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
	rules  [96]func() bool
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
			p.addOperatorLiteral("*")
		case ruleAction14:
			p.addOperatorLiteral("-")
		case ruleAction15:
			p.addOperatorFunction()
		case ruleAction16:
			p.addOperatorLiteral("*")
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
											{
												position27 := position
												depth++
												if !_rules[ruleSTRING]() {
													goto l20
												}
												depth--
												add(rulePegText, position27)
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
						depth--
						add(ruleselectStmt, position4)
					}
					goto l2
				l3:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
					{
						position32 := position
						depth++
						{
							position33, tokenIndex33, depth33 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l34
							}
							position++
							goto l33
						l34:
							position, tokenIndex, depth = position33, tokenIndex33, depth33
							if buffer[position] != rune('D') {
								goto l0
							}
							position++
						}
					l33:
						{
							position35, tokenIndex35, depth35 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l36
							}
							position++
							goto l35
						l36:
							position, tokenIndex, depth = position35, tokenIndex35, depth35
							if buffer[position] != rune('E') {
								goto l0
							}
							position++
						}
					l35:
						{
							position37, tokenIndex37, depth37 := position, tokenIndex, depth
							if buffer[position] != rune('s') {
								goto l38
							}
							position++
							goto l37
						l38:
							position, tokenIndex, depth = position37, tokenIndex37, depth37
							if buffer[position] != rune('S') {
								goto l0
							}
							position++
						}
					l37:
						{
							position39, tokenIndex39, depth39 := position, tokenIndex, depth
							if buffer[position] != rune('c') {
								goto l40
							}
							position++
							goto l39
						l40:
							position, tokenIndex, depth = position39, tokenIndex39, depth39
							if buffer[position] != rune('C') {
								goto l0
							}
							position++
						}
					l39:
						{
							position41, tokenIndex41, depth41 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l42
							}
							position++
							goto l41
						l42:
							position, tokenIndex, depth = position41, tokenIndex41, depth41
							if buffer[position] != rune('R') {
								goto l0
							}
							position++
						}
					l41:
						{
							position43, tokenIndex43, depth43 := position, tokenIndex, depth
							if buffer[position] != rune('i') {
								goto l44
							}
							position++
							goto l43
						l44:
							position, tokenIndex, depth = position43, tokenIndex43, depth43
							if buffer[position] != rune('I') {
								goto l0
							}
							position++
						}
					l43:
						{
							position45, tokenIndex45, depth45 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l46
							}
							position++
							goto l45
						l46:
							position, tokenIndex, depth = position45, tokenIndex45, depth45
							if buffer[position] != rune('B') {
								goto l0
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
								goto l0
							}
							position++
						}
					l47:
						if !_rules[rule__]() {
							goto l0
						}
						{
							position49, tokenIndex49, depth49 := position, tokenIndex, depth
							{
								position51 := position
								depth++
								{
									position52, tokenIndex52, depth52 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l53
									}
									position++
									goto l52
								l53:
									position, tokenIndex, depth = position52, tokenIndex52, depth52
									if buffer[position] != rune('A') {
										goto l50
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
										goto l50
									}
									position++
								}
							l54:
								{
									position56, tokenIndex56, depth56 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l57
									}
									position++
									goto l56
								l57:
									position, tokenIndex, depth = position56, tokenIndex56, depth56
									if buffer[position] != rune('L') {
										goto l50
									}
									position++
								}
							l56:
								{
									add(ruleAction1, position)
								}
								depth--
								add(ruledescribeAllStmt, position51)
							}
							goto l49
						l50:
							position, tokenIndex, depth = position49, tokenIndex49, depth49
							{
								position59 := position
								depth++
								{
									position60 := position
									depth++
									{
										position61 := position
										depth++
										if !_rules[ruleIDENTIFIER]() {
											goto l0
										}
										depth--
										add(ruleMETRIC_NAME, position61)
									}
									depth--
									add(rulePegText, position60)
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
								add(ruledescribeSingleStmt, position59)
							}
						}
					l49:
						depth--
						add(ruledescribeStmt, position32)
					}
				}
			l2:
				{
					position64, tokenIndex64, depth64 := position, tokenIndex, depth
					if !matchDot() {
						goto l64
					}
					goto l0
				l64:
					position, tokenIndex, depth = position64, tokenIndex64, depth64
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
		/* 5 propertyClause <- <(Action4 (_ PROPERTY_KEY Action5 __ PROPERTY_VALUE Action6 Action7)* Action8)> */
		nil,
		/* 6 optionalPredicateClause <- <((__ predicateClause) / Action9)> */
		func() bool {
			{
				position71 := position
				depth++
				{
					position72, tokenIndex72, depth72 := position, tokenIndex, depth
					if !_rules[rule__]() {
						goto l73
					}
					{
						position74 := position
						depth++
						{
							position75, tokenIndex75, depth75 := position, tokenIndex, depth
							if buffer[position] != rune('w') {
								goto l76
							}
							position++
							goto l75
						l76:
							position, tokenIndex, depth = position75, tokenIndex75, depth75
							if buffer[position] != rune('W') {
								goto l73
							}
							position++
						}
					l75:
						{
							position77, tokenIndex77, depth77 := position, tokenIndex, depth
							if buffer[position] != rune('h') {
								goto l78
							}
							position++
							goto l77
						l78:
							position, tokenIndex, depth = position77, tokenIndex77, depth77
							if buffer[position] != rune('H') {
								goto l73
							}
							position++
						}
					l77:
						{
							position79, tokenIndex79, depth79 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l80
							}
							position++
							goto l79
						l80:
							position, tokenIndex, depth = position79, tokenIndex79, depth79
							if buffer[position] != rune('E') {
								goto l73
							}
							position++
						}
					l79:
						{
							position81, tokenIndex81, depth81 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l82
							}
							position++
							goto l81
						l82:
							position, tokenIndex, depth = position81, tokenIndex81, depth81
							if buffer[position] != rune('R') {
								goto l73
							}
							position++
						}
					l81:
						{
							position83, tokenIndex83, depth83 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l84
							}
							position++
							goto l83
						l84:
							position, tokenIndex, depth = position83, tokenIndex83, depth83
							if buffer[position] != rune('E') {
								goto l73
							}
							position++
						}
					l83:
						if !_rules[rule__]() {
							goto l73
						}
						if !_rules[rulepredicate_1]() {
							goto l73
						}
						depth--
						add(rulepredicateClause, position74)
					}
					goto l72
				l73:
					position, tokenIndex, depth = position72, tokenIndex72, depth72
					{
						add(ruleAction9, position)
					}
				}
			l72:
				depth--
				add(ruleoptionalPredicateClause, position71)
			}
			return true
		},
		/* 7 expressionList <- <(Action10 expression_1 Action11 (COMMA expression_1 Action12)*)> */
		func() bool {
			position86, tokenIndex86, depth86 := position, tokenIndex, depth
			{
				position87 := position
				depth++
				{
					add(ruleAction10, position)
				}
				if !_rules[ruleexpression_1]() {
					goto l86
				}
				{
					add(ruleAction11, position)
				}
			l90:
				{
					position91, tokenIndex91, depth91 := position, tokenIndex, depth
					if !_rules[ruleCOMMA]() {
						goto l91
					}
					if !_rules[ruleexpression_1]() {
						goto l91
					}
					{
						add(ruleAction12, position)
					}
					goto l90
				l91:
					position, tokenIndex, depth = position91, tokenIndex91, depth91
				}
				depth--
				add(ruleexpressionList, position87)
			}
			return true
		l86:
			position, tokenIndex, depth = position86, tokenIndex86, depth86
			return false
		},
		/* 8 expression_1 <- <(expression_2 (((OP_ADD Action13) / (OP_SUB Action14)) expression_2 Action15)*)> */
		func() bool {
			position93, tokenIndex93, depth93 := position, tokenIndex, depth
			{
				position94 := position
				depth++
				if !_rules[ruleexpression_2]() {
					goto l93
				}
			l95:
				{
					position96, tokenIndex96, depth96 := position, tokenIndex, depth
					{
						position97, tokenIndex97, depth97 := position, tokenIndex, depth
						{
							position99 := position
							depth++
							if !_rules[rule_]() {
								goto l98
							}
							if buffer[position] != rune('+') {
								goto l98
							}
							position++
							if !_rules[rule_]() {
								goto l98
							}
							depth--
							add(ruleOP_ADD, position99)
						}
						{
							add(ruleAction13, position)
						}
						goto l97
					l98:
						position, tokenIndex, depth = position97, tokenIndex97, depth97
						{
							position101 := position
							depth++
							if !_rules[rule_]() {
								goto l96
							}
							if buffer[position] != rune('-') {
								goto l96
							}
							position++
							if !_rules[rule_]() {
								goto l96
							}
							depth--
							add(ruleOP_SUB, position101)
						}
						{
							add(ruleAction14, position)
						}
					}
				l97:
					if !_rules[ruleexpression_2]() {
						goto l96
					}
					{
						add(ruleAction15, position)
					}
					goto l95
				l96:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
				}
				depth--
				add(ruleexpression_1, position94)
			}
			return true
		l93:
			position, tokenIndex, depth = position93, tokenIndex93, depth93
			return false
		},
		/* 9 expression_2 <- <(expression_3 (((OP_DIV Action16) / (OP_MULT Action17)) expression_3 Action18)*)> */
		func() bool {
			position104, tokenIndex104, depth104 := position, tokenIndex, depth
			{
				position105 := position
				depth++
				if !_rules[ruleexpression_3]() {
					goto l104
				}
			l106:
				{
					position107, tokenIndex107, depth107 := position, tokenIndex, depth
					{
						position108, tokenIndex108, depth108 := position, tokenIndex, depth
						{
							position110 := position
							depth++
							if !_rules[rule_]() {
								goto l109
							}
							if buffer[position] != rune('/') {
								goto l109
							}
							position++
							if !_rules[rule_]() {
								goto l109
							}
							depth--
							add(ruleOP_DIV, position110)
						}
						{
							add(ruleAction16, position)
						}
						goto l108
					l109:
						position, tokenIndex, depth = position108, tokenIndex108, depth108
						{
							position112 := position
							depth++
							if !_rules[rule_]() {
								goto l107
							}
							if buffer[position] != rune('*') {
								goto l107
							}
							position++
							if !_rules[rule_]() {
								goto l107
							}
							depth--
							add(ruleOP_MULT, position112)
						}
						{
							add(ruleAction17, position)
						}
					}
				l108:
					if !_rules[ruleexpression_3]() {
						goto l107
					}
					{
						add(ruleAction18, position)
					}
					goto l106
				l107:
					position, tokenIndex, depth = position107, tokenIndex107, depth107
				}
				depth--
				add(ruleexpression_2, position105)
			}
			return true
		l104:
			position, tokenIndex, depth = position104, tokenIndex104, depth104
			return false
		},
		/* 10 expression_3 <- <(expression_function / ((&('-' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') (<NUMBER> Action19)) | (&('\t' | '\n' | ' ' | '(') (PAREN_OPEN expression_1 PAREN_CLOSE)) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') expression_metric)))> */
		func() bool {
			position115, tokenIndex115, depth115 := position, tokenIndex, depth
			{
				position116 := position
				depth++
				{
					position117, tokenIndex117, depth117 := position, tokenIndex, depth
					{
						position119 := position
						depth++
						{
							position120 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l118
							}
							depth--
							add(rulePegText, position120)
						}
						{
							add(ruleAction20, position)
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l118
						}
						if !_rules[ruleexpressionList]() {
							goto l118
						}
						{
							add(ruleAction21, position)
						}
						{
							position123, tokenIndex123, depth123 := position, tokenIndex, depth
							if !_rules[rule__]() {
								goto l123
							}
							{
								position125 := position
								depth++
								{
									position126, tokenIndex126, depth126 := position, tokenIndex, depth
									if buffer[position] != rune('g') {
										goto l127
									}
									position++
									goto l126
								l127:
									position, tokenIndex, depth = position126, tokenIndex126, depth126
									if buffer[position] != rune('G') {
										goto l123
									}
									position++
								}
							l126:
								{
									position128, tokenIndex128, depth128 := position, tokenIndex, depth
									if buffer[position] != rune('r') {
										goto l129
									}
									position++
									goto l128
								l129:
									position, tokenIndex, depth = position128, tokenIndex128, depth128
									if buffer[position] != rune('R') {
										goto l123
									}
									position++
								}
							l128:
								{
									position130, tokenIndex130, depth130 := position, tokenIndex, depth
									if buffer[position] != rune('o') {
										goto l131
									}
									position++
									goto l130
								l131:
									position, tokenIndex, depth = position130, tokenIndex130, depth130
									if buffer[position] != rune('O') {
										goto l123
									}
									position++
								}
							l130:
								{
									position132, tokenIndex132, depth132 := position, tokenIndex, depth
									if buffer[position] != rune('u') {
										goto l133
									}
									position++
									goto l132
								l133:
									position, tokenIndex, depth = position132, tokenIndex132, depth132
									if buffer[position] != rune('U') {
										goto l123
									}
									position++
								}
							l132:
								{
									position134, tokenIndex134, depth134 := position, tokenIndex, depth
									if buffer[position] != rune('p') {
										goto l135
									}
									position++
									goto l134
								l135:
									position, tokenIndex, depth = position134, tokenIndex134, depth134
									if buffer[position] != rune('P') {
										goto l123
									}
									position++
								}
							l134:
								if !_rules[rule__]() {
									goto l123
								}
								{
									position136, tokenIndex136, depth136 := position, tokenIndex, depth
									if buffer[position] != rune('b') {
										goto l137
									}
									position++
									goto l136
								l137:
									position, tokenIndex, depth = position136, tokenIndex136, depth136
									if buffer[position] != rune('B') {
										goto l123
									}
									position++
								}
							l136:
								{
									position138, tokenIndex138, depth138 := position, tokenIndex, depth
									if buffer[position] != rune('y') {
										goto l139
									}
									position++
									goto l138
								l139:
									position, tokenIndex, depth = position138, tokenIndex138, depth138
									if buffer[position] != rune('Y') {
										goto l123
									}
									position++
								}
							l138:
								if !_rules[rule__]() {
									goto l123
								}
								{
									position140 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l123
									}
									depth--
									add(rulePegText, position140)
								}
								{
									add(ruleAction26, position)
								}
							l142:
								{
									position143, tokenIndex143, depth143 := position, tokenIndex, depth
									if !_rules[ruleCOMMA]() {
										goto l143
									}
									{
										position144 := position
										depth++
										if !_rules[ruleCOLUMN_NAME]() {
											goto l143
										}
										depth--
										add(rulePegText, position144)
									}
									{
										add(ruleAction27, position)
									}
									goto l142
								l143:
									position, tokenIndex, depth = position143, tokenIndex143, depth143
								}
								depth--
								add(rulegroupByClause, position125)
							}
							goto l124
						l123:
							position, tokenIndex, depth = position123, tokenIndex123, depth123
						}
					l124:
						if !_rules[rulePAREN_CLOSE]() {
							goto l118
						}
						{
							add(ruleAction22, position)
						}
						depth--
						add(ruleexpression_function, position119)
					}
					goto l117
				l118:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
					{
						switch buffer[position] {
						case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							{
								position148 := position
								depth++
								if !_rules[ruleNUMBER]() {
									goto l115
								}
								depth--
								add(rulePegText, position148)
							}
							{
								add(ruleAction19, position)
							}
							break
						case '\t', '\n', ' ', '(':
							if !_rules[rulePAREN_OPEN]() {
								goto l115
							}
							if !_rules[ruleexpression_1]() {
								goto l115
							}
							if !_rules[rulePAREN_CLOSE]() {
								goto l115
							}
							break
						default:
							{
								position150 := position
								depth++
								{
									position151 := position
									depth++
									if !_rules[ruleIDENTIFIER]() {
										goto l115
									}
									depth--
									add(rulePegText, position151)
								}
								{
									add(ruleAction23, position)
								}
								{
									position153, tokenIndex153, depth153 := position, tokenIndex, depth
									{
										position155, tokenIndex155, depth155 := position, tokenIndex, depth
										if buffer[position] != rune('[') {
											goto l156
										}
										position++
										if !_rules[rule_]() {
											goto l156
										}
										if !_rules[rulepredicate_1]() {
											goto l156
										}
										if !_rules[rule_]() {
											goto l156
										}
										if buffer[position] != rune(']') {
											goto l156
										}
										position++
										goto l155
									l156:
										position, tokenIndex, depth = position155, tokenIndex155, depth155
										{
											add(ruleAction24, position)
										}
									}
								l155:
									goto l154

									position, tokenIndex, depth = position153, tokenIndex153, depth153
								}
							l154:
								{
									add(ruleAction25, position)
								}
								depth--
								add(ruleexpression_metric, position150)
							}
							break
						}
					}

				}
			l117:
				depth--
				add(ruleexpression_3, position116)
			}
			return true
		l115:
			position, tokenIndex, depth = position115, tokenIndex115, depth115
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
				position164 := position
				depth++
				{
					position165, tokenIndex165, depth165 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l166
					}
					{
						position167 := position
						depth++
						if !_rules[rule_]() {
							goto l166
						}
						{
							position168, tokenIndex168, depth168 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l169
							}
							position++
							goto l168
						l169:
							position, tokenIndex, depth = position168, tokenIndex168, depth168
							if buffer[position] != rune('O') {
								goto l166
							}
							position++
						}
					l168:
						{
							position170, tokenIndex170, depth170 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l171
							}
							position++
							goto l170
						l171:
							position, tokenIndex, depth = position170, tokenIndex170, depth170
							if buffer[position] != rune('R') {
								goto l166
							}
							position++
						}
					l170:
						if !_rules[rule_]() {
							goto l166
						}
						depth--
						add(ruleOP_OR, position167)
					}
					if !_rules[rulepredicate_1]() {
						goto l166
					}
					{
						add(ruleAction28, position)
					}
					goto l165
				l166:
					position, tokenIndex, depth = position165, tokenIndex165, depth165
					if !_rules[rulepredicate_2]() {
						goto l173
					}
					goto l165
				l173:
					position, tokenIndex, depth = position165, tokenIndex165, depth165
				}
			l165:
				depth--
				add(rulepredicate_1, position164)
			}
			return true
		},
		/* 16 predicate_2 <- <((predicate_3 OP_AND predicate_2 Action29) / predicate_3)> */
		func() bool {
			position174, tokenIndex174, depth174 := position, tokenIndex, depth
			{
				position175 := position
				depth++
				{
					position176, tokenIndex176, depth176 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l177
					}
					{
						position178 := position
						depth++
						if !_rules[rule_]() {
							goto l177
						}
						{
							position179, tokenIndex179, depth179 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l180
							}
							position++
							goto l179
						l180:
							position, tokenIndex, depth = position179, tokenIndex179, depth179
							if buffer[position] != rune('A') {
								goto l177
							}
							position++
						}
					l179:
						{
							position181, tokenIndex181, depth181 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l182
							}
							position++
							goto l181
						l182:
							position, tokenIndex, depth = position181, tokenIndex181, depth181
							if buffer[position] != rune('N') {
								goto l177
							}
							position++
						}
					l181:
						{
							position183, tokenIndex183, depth183 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l184
							}
							position++
							goto l183
						l184:
							position, tokenIndex, depth = position183, tokenIndex183, depth183
							if buffer[position] != rune('D') {
								goto l177
							}
							position++
						}
					l183:
						if !_rules[rule_]() {
							goto l177
						}
						depth--
						add(ruleOP_AND, position178)
					}
					if !_rules[rulepredicate_2]() {
						goto l177
					}
					{
						add(ruleAction29, position)
					}
					goto l176
				l177:
					position, tokenIndex, depth = position176, tokenIndex176, depth176
					if !_rules[rulepredicate_3]() {
						goto l174
					}
				}
			l176:
				depth--
				add(rulepredicate_2, position175)
			}
			return true
		l174:
			position, tokenIndex, depth = position174, tokenIndex174, depth174
			return false
		},
		/* 17 predicate_3 <- <((OP_NOT predicate_3 Action30) / (PAREN_OPEN predicate_1 PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position186, tokenIndex186, depth186 := position, tokenIndex, depth
			{
				position187 := position
				depth++
				{
					position188, tokenIndex188, depth188 := position, tokenIndex, depth
					{
						position190 := position
						depth++
						{
							position191, tokenIndex191, depth191 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l192
							}
							position++
							goto l191
						l192:
							position, tokenIndex, depth = position191, tokenIndex191, depth191
							if buffer[position] != rune('N') {
								goto l189
							}
							position++
						}
					l191:
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
								goto l189
							}
							position++
						}
					l193:
						{
							position195, tokenIndex195, depth195 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l196
							}
							position++
							goto l195
						l196:
							position, tokenIndex, depth = position195, tokenIndex195, depth195
							if buffer[position] != rune('T') {
								goto l189
							}
							position++
						}
					l195:
						if !_rules[rule__]() {
							goto l189
						}
						depth--
						add(ruleOP_NOT, position190)
					}
					if !_rules[rulepredicate_3]() {
						goto l189
					}
					{
						add(ruleAction30, position)
					}
					goto l188
				l189:
					position, tokenIndex, depth = position188, tokenIndex188, depth188
					if !_rules[rulePAREN_OPEN]() {
						goto l198
					}
					if !_rules[rulepredicate_1]() {
						goto l198
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l198
					}
					goto l188
				l198:
					position, tokenIndex, depth = position188, tokenIndex188, depth188
					{
						position199 := position
						depth++
						{
							position200, tokenIndex200, depth200 := position, tokenIndex, depth
							if !_rules[ruletagName]() {
								goto l201
							}
							if !_rules[rule_]() {
								goto l201
							}
							if buffer[position] != rune('=') {
								goto l201
							}
							position++
							if !_rules[rule_]() {
								goto l201
							}
							if !_rules[ruleliteralString]() {
								goto l201
							}
							{
								add(ruleAction31, position)
							}
							goto l200
						l201:
							position, tokenIndex, depth = position200, tokenIndex200, depth200
							if !_rules[ruletagName]() {
								goto l203
							}
							if !_rules[rule_]() {
								goto l203
							}
							if buffer[position] != rune('!') {
								goto l203
							}
							position++
							if buffer[position] != rune('=') {
								goto l203
							}
							position++
							if !_rules[rule_]() {
								goto l203
							}
							if !_rules[ruleliteralString]() {
								goto l203
							}
							{
								add(ruleAction32, position)
							}
							goto l200
						l203:
							position, tokenIndex, depth = position200, tokenIndex200, depth200
							if !_rules[ruletagName]() {
								goto l205
							}
							if !_rules[rule__]() {
								goto l205
							}
							{
								position206, tokenIndex206, depth206 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l207
								}
								position++
								goto l206
							l207:
								position, tokenIndex, depth = position206, tokenIndex206, depth206
								if buffer[position] != rune('M') {
									goto l205
								}
								position++
							}
						l206:
							{
								position208, tokenIndex208, depth208 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l209
								}
								position++
								goto l208
							l209:
								position, tokenIndex, depth = position208, tokenIndex208, depth208
								if buffer[position] != rune('A') {
									goto l205
								}
								position++
							}
						l208:
							{
								position210, tokenIndex210, depth210 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l211
								}
								position++
								goto l210
							l211:
								position, tokenIndex, depth = position210, tokenIndex210, depth210
								if buffer[position] != rune('T') {
									goto l205
								}
								position++
							}
						l210:
							{
								position212, tokenIndex212, depth212 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l213
								}
								position++
								goto l212
							l213:
								position, tokenIndex, depth = position212, tokenIndex212, depth212
								if buffer[position] != rune('C') {
									goto l205
								}
								position++
							}
						l212:
							{
								position214, tokenIndex214, depth214 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l215
								}
								position++
								goto l214
							l215:
								position, tokenIndex, depth = position214, tokenIndex214, depth214
								if buffer[position] != rune('H') {
									goto l205
								}
								position++
							}
						l214:
							{
								position216, tokenIndex216, depth216 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l217
								}
								position++
								goto l216
							l217:
								position, tokenIndex, depth = position216, tokenIndex216, depth216
								if buffer[position] != rune('E') {
									goto l205
								}
								position++
							}
						l216:
							{
								position218, tokenIndex218, depth218 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l219
								}
								position++
								goto l218
							l219:
								position, tokenIndex, depth = position218, tokenIndex218, depth218
								if buffer[position] != rune('S') {
									goto l205
								}
								position++
							}
						l218:
							if !_rules[rule__]() {
								goto l205
							}
							if !_rules[ruleliteralString]() {
								goto l205
							}
							{
								add(ruleAction33, position)
							}
							goto l200
						l205:
							position, tokenIndex, depth = position200, tokenIndex200, depth200
							if !_rules[ruletagName]() {
								goto l186
							}
							if !_rules[rule__]() {
								goto l186
							}
							{
								position221, tokenIndex221, depth221 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l222
								}
								position++
								goto l221
							l222:
								position, tokenIndex, depth = position221, tokenIndex221, depth221
								if buffer[position] != rune('I') {
									goto l186
								}
								position++
							}
						l221:
							{
								position223, tokenIndex223, depth223 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l224
								}
								position++
								goto l223
							l224:
								position, tokenIndex, depth = position223, tokenIndex223, depth223
								if buffer[position] != rune('N') {
									goto l186
								}
								position++
							}
						l223:
							if !_rules[rule__]() {
								goto l186
							}
							{
								position225 := position
								depth++
								{
									add(ruleAction36, position)
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l186
								}
								if !_rules[ruleliteralListString]() {
									goto l186
								}
							l227:
								{
									position228, tokenIndex228, depth228 := position, tokenIndex, depth
									if !_rules[ruleCOMMA]() {
										goto l228
									}
									if !_rules[ruleliteralListString]() {
										goto l228
									}
									goto l227
								l228:
									position, tokenIndex, depth = position228, tokenIndex228, depth228
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l186
								}
								depth--
								add(ruleliteralList, position225)
							}
							{
								add(ruleAction34, position)
							}
						}
					l200:
						depth--
						add(ruletagMatcher, position199)
					}
				}
			l188:
				depth--
				add(rulepredicate_3, position187)
			}
			return true
		l186:
			position, tokenIndex, depth = position186, tokenIndex186, depth186
			return false
		},
		/* 18 tagMatcher <- <((tagName _ '=' _ literalString Action31) / (tagName _ ('!' '=') _ literalString Action32) / (tagName __ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) __ literalString Action33) / (tagName __ (('i' / 'I') ('n' / 'N')) __ literalList Action34))> */
		nil,
		/* 19 literalString <- <(<STRING> Action35)> */
		func() bool {
			position231, tokenIndex231, depth231 := position, tokenIndex, depth
			{
				position232 := position
				depth++
				{
					position233 := position
					depth++
					if !_rules[ruleSTRING]() {
						goto l231
					}
					depth--
					add(rulePegText, position233)
				}
				{
					add(ruleAction35, position)
				}
				depth--
				add(ruleliteralString, position232)
			}
			return true
		l231:
			position, tokenIndex, depth = position231, tokenIndex231, depth231
			return false
		},
		/* 20 literalList <- <(Action36 PAREN_OPEN literalListString (COMMA literalListString)* PAREN_CLOSE)> */
		nil,
		/* 21 literalListString <- <(STRING Action37)> */
		func() bool {
			position236, tokenIndex236, depth236 := position, tokenIndex, depth
			{
				position237 := position
				depth++
				if !_rules[ruleSTRING]() {
					goto l236
				}
				{
					add(ruleAction37, position)
				}
				depth--
				add(ruleliteralListString, position237)
			}
			return true
		l236:
			position, tokenIndex, depth = position236, tokenIndex236, depth236
			return false
		},
		/* 22 tagName <- <(<TAG_NAME> Action38)> */
		func() bool {
			position239, tokenIndex239, depth239 := position, tokenIndex, depth
			{
				position240 := position
				depth++
				{
					position241 := position
					depth++
					{
						position242 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l239
						}
						depth--
						add(ruleTAG_NAME, position242)
					}
					depth--
					add(rulePegText, position241)
				}
				{
					add(ruleAction38, position)
				}
				depth--
				add(ruletagName, position240)
			}
			return true
		l239:
			position, tokenIndex, depth = position239, tokenIndex239, depth239
			return false
		},
		/* 23 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position244, tokenIndex244, depth244 := position, tokenIndex, depth
			{
				position245 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l244
				}
				depth--
				add(ruleCOLUMN_NAME, position245)
			}
			return true
		l244:
			position, tokenIndex, depth = position244, tokenIndex244, depth244
			return false
		},
		/* 24 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 25 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 26 TIMESTAMP <- <(<NUMBER> / <STRING>)> */
		nil,
		/* 27 IDENTIFIER <- <(('`' CHAR* '`') / (!(KEYWORD !ID_CONT) ID_SEGMENT ('.' ID_SEGMENT)*))> */
		func() bool {
			position249, tokenIndex249, depth249 := position, tokenIndex, depth
			{
				position250 := position
				depth++
				{
					position251, tokenIndex251, depth251 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l252
					}
					position++
				l253:
					{
						position254, tokenIndex254, depth254 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l254
						}
						goto l253
					l254:
						position, tokenIndex, depth = position254, tokenIndex254, depth254
					}
					if buffer[position] != rune('`') {
						goto l252
					}
					position++
					goto l251
				l252:
					position, tokenIndex, depth = position251, tokenIndex251, depth251
					{
						position255, tokenIndex255, depth255 := position, tokenIndex, depth
						{
							position256 := position
							depth++
							{
								position257, tokenIndex257, depth257 := position, tokenIndex, depth
								{
									position259, tokenIndex259, depth259 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l260
									}
									position++
									goto l259
								l260:
									position, tokenIndex, depth = position259, tokenIndex259, depth259
									if buffer[position] != rune('A') {
										goto l258
									}
									position++
								}
							l259:
								{
									position261, tokenIndex261, depth261 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l262
									}
									position++
									goto l261
								l262:
									position, tokenIndex, depth = position261, tokenIndex261, depth261
									if buffer[position] != rune('L') {
										goto l258
									}
									position++
								}
							l261:
								{
									position263, tokenIndex263, depth263 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l264
									}
									position++
									goto l263
								l264:
									position, tokenIndex, depth = position263, tokenIndex263, depth263
									if buffer[position] != rune('L') {
										goto l258
									}
									position++
								}
							l263:
								goto l257
							l258:
								position, tokenIndex, depth = position257, tokenIndex257, depth257
								{
									position266, tokenIndex266, depth266 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l267
									}
									position++
									goto l266
								l267:
									position, tokenIndex, depth = position266, tokenIndex266, depth266
									if buffer[position] != rune('A') {
										goto l265
									}
									position++
								}
							l266:
								{
									position268, tokenIndex268, depth268 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l269
									}
									position++
									goto l268
								l269:
									position, tokenIndex, depth = position268, tokenIndex268, depth268
									if buffer[position] != rune('N') {
										goto l265
									}
									position++
								}
							l268:
								{
									position270, tokenIndex270, depth270 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l271
									}
									position++
									goto l270
								l271:
									position, tokenIndex, depth = position270, tokenIndex270, depth270
									if buffer[position] != rune('D') {
										goto l265
									}
									position++
								}
							l270:
								goto l257
							l265:
								position, tokenIndex, depth = position257, tokenIndex257, depth257
								{
									position273, tokenIndex273, depth273 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l274
									}
									position++
									goto l273
								l274:
									position, tokenIndex, depth = position273, tokenIndex273, depth273
									if buffer[position] != rune('S') {
										goto l272
									}
									position++
								}
							l273:
								{
									position275, tokenIndex275, depth275 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l276
									}
									position++
									goto l275
								l276:
									position, tokenIndex, depth = position275, tokenIndex275, depth275
									if buffer[position] != rune('E') {
										goto l272
									}
									position++
								}
							l275:
								{
									position277, tokenIndex277, depth277 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l278
									}
									position++
									goto l277
								l278:
									position, tokenIndex, depth = position277, tokenIndex277, depth277
									if buffer[position] != rune('L') {
										goto l272
									}
									position++
								}
							l277:
								{
									position279, tokenIndex279, depth279 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l280
									}
									position++
									goto l279
								l280:
									position, tokenIndex, depth = position279, tokenIndex279, depth279
									if buffer[position] != rune('E') {
										goto l272
									}
									position++
								}
							l279:
								{
									position281, tokenIndex281, depth281 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l282
									}
									position++
									goto l281
								l282:
									position, tokenIndex, depth = position281, tokenIndex281, depth281
									if buffer[position] != rune('C') {
										goto l272
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
										goto l272
									}
									position++
								}
							l283:
								goto l257
							l272:
								position, tokenIndex, depth = position257, tokenIndex257, depth257
								{
									switch buffer[position] {
									case 'W', 'w':
										{
											position286, tokenIndex286, depth286 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l287
											}
											position++
											goto l286
										l287:
											position, tokenIndex, depth = position286, tokenIndex286, depth286
											if buffer[position] != rune('W') {
												goto l255
											}
											position++
										}
									l286:
										{
											position288, tokenIndex288, depth288 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l289
											}
											position++
											goto l288
										l289:
											position, tokenIndex, depth = position288, tokenIndex288, depth288
											if buffer[position] != rune('H') {
												goto l255
											}
											position++
										}
									l288:
										{
											position290, tokenIndex290, depth290 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l291
											}
											position++
											goto l290
										l291:
											position, tokenIndex, depth = position290, tokenIndex290, depth290
											if buffer[position] != rune('E') {
												goto l255
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
												goto l255
											}
											position++
										}
									l292:
										{
											position294, tokenIndex294, depth294 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l295
											}
											position++
											goto l294
										l295:
											position, tokenIndex, depth = position294, tokenIndex294, depth294
											if buffer[position] != rune('E') {
												goto l255
											}
											position++
										}
									l294:
										break
									case 'O', 'o':
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
												goto l255
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
												goto l255
											}
											position++
										}
									l298:
										break
									case 'N', 'n':
										{
											position300, tokenIndex300, depth300 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l301
											}
											position++
											goto l300
										l301:
											position, tokenIndex, depth = position300, tokenIndex300, depth300
											if buffer[position] != rune('N') {
												goto l255
											}
											position++
										}
									l300:
										{
											position302, tokenIndex302, depth302 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l303
											}
											position++
											goto l302
										l303:
											position, tokenIndex, depth = position302, tokenIndex302, depth302
											if buffer[position] != rune('O') {
												goto l255
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
												goto l255
											}
											position++
										}
									l304:
										break
									case 'M', 'm':
										{
											position306, tokenIndex306, depth306 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l307
											}
											position++
											goto l306
										l307:
											position, tokenIndex, depth = position306, tokenIndex306, depth306
											if buffer[position] != rune('M') {
												goto l255
											}
											position++
										}
									l306:
										{
											position308, tokenIndex308, depth308 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l309
											}
											position++
											goto l308
										l309:
											position, tokenIndex, depth = position308, tokenIndex308, depth308
											if buffer[position] != rune('A') {
												goto l255
											}
											position++
										}
									l308:
										{
											position310, tokenIndex310, depth310 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l311
											}
											position++
											goto l310
										l311:
											position, tokenIndex, depth = position310, tokenIndex310, depth310
											if buffer[position] != rune('T') {
												goto l255
											}
											position++
										}
									l310:
										{
											position312, tokenIndex312, depth312 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l313
											}
											position++
											goto l312
										l313:
											position, tokenIndex, depth = position312, tokenIndex312, depth312
											if buffer[position] != rune('C') {
												goto l255
											}
											position++
										}
									l312:
										{
											position314, tokenIndex314, depth314 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l315
											}
											position++
											goto l314
										l315:
											position, tokenIndex, depth = position314, tokenIndex314, depth314
											if buffer[position] != rune('H') {
												goto l255
											}
											position++
										}
									l314:
										{
											position316, tokenIndex316, depth316 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l317
											}
											position++
											goto l316
										l317:
											position, tokenIndex, depth = position316, tokenIndex316, depth316
											if buffer[position] != rune('E') {
												goto l255
											}
											position++
										}
									l316:
										{
											position318, tokenIndex318, depth318 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l319
											}
											position++
											goto l318
										l319:
											position, tokenIndex, depth = position318, tokenIndex318, depth318
											if buffer[position] != rune('S') {
												goto l255
											}
											position++
										}
									l318:
										break
									case 'I', 'i':
										{
											position320, tokenIndex320, depth320 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l321
											}
											position++
											goto l320
										l321:
											position, tokenIndex, depth = position320, tokenIndex320, depth320
											if buffer[position] != rune('I') {
												goto l255
											}
											position++
										}
									l320:
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
												goto l255
											}
											position++
										}
									l322:
										break
									case 'G', 'g':
										{
											position324, tokenIndex324, depth324 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l325
											}
											position++
											goto l324
										l325:
											position, tokenIndex, depth = position324, tokenIndex324, depth324
											if buffer[position] != rune('G') {
												goto l255
											}
											position++
										}
									l324:
										{
											position326, tokenIndex326, depth326 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l327
											}
											position++
											goto l326
										l327:
											position, tokenIndex, depth = position326, tokenIndex326, depth326
											if buffer[position] != rune('R') {
												goto l255
											}
											position++
										}
									l326:
										{
											position328, tokenIndex328, depth328 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l329
											}
											position++
											goto l328
										l329:
											position, tokenIndex, depth = position328, tokenIndex328, depth328
											if buffer[position] != rune('O') {
												goto l255
											}
											position++
										}
									l328:
										{
											position330, tokenIndex330, depth330 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l331
											}
											position++
											goto l330
										l331:
											position, tokenIndex, depth = position330, tokenIndex330, depth330
											if buffer[position] != rune('U') {
												goto l255
											}
											position++
										}
									l330:
										{
											position332, tokenIndex332, depth332 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l333
											}
											position++
											goto l332
										l333:
											position, tokenIndex, depth = position332, tokenIndex332, depth332
											if buffer[position] != rune('P') {
												goto l255
											}
											position++
										}
									l332:
										break
									case 'D', 'd':
										{
											position334, tokenIndex334, depth334 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l335
											}
											position++
											goto l334
										l335:
											position, tokenIndex, depth = position334, tokenIndex334, depth334
											if buffer[position] != rune('D') {
												goto l255
											}
											position++
										}
									l334:
										{
											position336, tokenIndex336, depth336 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l337
											}
											position++
											goto l336
										l337:
											position, tokenIndex, depth = position336, tokenIndex336, depth336
											if buffer[position] != rune('E') {
												goto l255
											}
											position++
										}
									l336:
										{
											position338, tokenIndex338, depth338 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l339
											}
											position++
											goto l338
										l339:
											position, tokenIndex, depth = position338, tokenIndex338, depth338
											if buffer[position] != rune('S') {
												goto l255
											}
											position++
										}
									l338:
										{
											position340, tokenIndex340, depth340 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l341
											}
											position++
											goto l340
										l341:
											position, tokenIndex, depth = position340, tokenIndex340, depth340
											if buffer[position] != rune('C') {
												goto l255
											}
											position++
										}
									l340:
										{
											position342, tokenIndex342, depth342 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l343
											}
											position++
											goto l342
										l343:
											position, tokenIndex, depth = position342, tokenIndex342, depth342
											if buffer[position] != rune('R') {
												goto l255
											}
											position++
										}
									l342:
										{
											position344, tokenIndex344, depth344 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l345
											}
											position++
											goto l344
										l345:
											position, tokenIndex, depth = position344, tokenIndex344, depth344
											if buffer[position] != rune('I') {
												goto l255
											}
											position++
										}
									l344:
										{
											position346, tokenIndex346, depth346 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l347
											}
											position++
											goto l346
										l347:
											position, tokenIndex, depth = position346, tokenIndex346, depth346
											if buffer[position] != rune('B') {
												goto l255
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
												goto l255
											}
											position++
										}
									l348:
										break
									case 'B', 'b':
										{
											position350, tokenIndex350, depth350 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l351
											}
											position++
											goto l350
										l351:
											position, tokenIndex, depth = position350, tokenIndex350, depth350
											if buffer[position] != rune('B') {
												goto l255
											}
											position++
										}
									l350:
										{
											position352, tokenIndex352, depth352 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l353
											}
											position++
											goto l352
										l353:
											position, tokenIndex, depth = position352, tokenIndex352, depth352
											if buffer[position] != rune('Y') {
												goto l255
											}
											position++
										}
									l352:
										break
									case 'A', 'a':
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
												goto l255
											}
											position++
										}
									l354:
										{
											position356, tokenIndex356, depth356 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l357
											}
											position++
											goto l356
										l357:
											position, tokenIndex, depth = position356, tokenIndex356, depth356
											if buffer[position] != rune('S') {
												goto l255
											}
											position++
										}
									l356:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l255
										}
										break
									}
								}

							}
						l257:
							depth--
							add(ruleKEYWORD, position256)
						}
						{
							position358, tokenIndex358, depth358 := position, tokenIndex, depth
							if !_rules[ruleID_CONT]() {
								goto l358
							}
							goto l255
						l358:
							position, tokenIndex, depth = position358, tokenIndex358, depth358
						}
						goto l249
					l255:
						position, tokenIndex, depth = position255, tokenIndex255, depth255
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l249
					}
				l359:
					{
						position360, tokenIndex360, depth360 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l360
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l360
						}
						goto l359
					l360:
						position, tokenIndex, depth = position360, tokenIndex360, depth360
					}
				}
			l251:
				depth--
				add(ruleIDENTIFIER, position250)
			}
			return true
		l249:
			position, tokenIndex, depth = position249, tokenIndex249, depth249
			return false
		},
		/* 28 ID_SEGMENT <- <(ID_START ID_CONT*)> */
		func() bool {
			position361, tokenIndex361, depth361 := position, tokenIndex, depth
			{
				position362 := position
				depth++
				if !_rules[ruleID_START]() {
					goto l361
				}
			l363:
				{
					position364, tokenIndex364, depth364 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l364
					}
					goto l363
				l364:
					position, tokenIndex, depth = position364, tokenIndex364, depth364
				}
				depth--
				add(ruleID_SEGMENT, position362)
			}
			return true
		l361:
			position, tokenIndex, depth = position361, tokenIndex361, depth361
			return false
		},
		/* 29 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position365, tokenIndex365, depth365 := position, tokenIndex, depth
			{
				position366 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l365
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l365
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l365
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position366)
			}
			return true
		l365:
			position, tokenIndex, depth = position365, tokenIndex365, depth365
			return false
		},
		/* 30 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position368, tokenIndex368, depth368 := position, tokenIndex, depth
			{
				position369 := position
				depth++
				{
					position370, tokenIndex370, depth370 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l371
					}
					goto l370
				l371:
					position, tokenIndex, depth = position370, tokenIndex370, depth370
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l368
					}
					position++
				}
			l370:
				depth--
				add(ruleID_CONT, position369)
			}
			return true
		l368:
			position, tokenIndex, depth = position368, tokenIndex368, depth368
			return false
		},
		/* 31 PROPERTY_KEY <- <((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> __ (('b' / 'B') ('y' / 'Y')))) | (&('R' | 'r') <(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))>) | (&('T' | 't') <(('t' / 'T') ('o' / 'O'))>) | (&('F' | 'f') <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))>))> */
		func() bool {
			position372, tokenIndex372, depth372 := position, tokenIndex, depth
			{
				position373 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position375 := position
							depth++
							{
								position376, tokenIndex376, depth376 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l377
								}
								position++
								goto l376
							l377:
								position, tokenIndex, depth = position376, tokenIndex376, depth376
								if buffer[position] != rune('S') {
									goto l372
								}
								position++
							}
						l376:
							{
								position378, tokenIndex378, depth378 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l379
								}
								position++
								goto l378
							l379:
								position, tokenIndex, depth = position378, tokenIndex378, depth378
								if buffer[position] != rune('A') {
									goto l372
								}
								position++
							}
						l378:
							{
								position380, tokenIndex380, depth380 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l381
								}
								position++
								goto l380
							l381:
								position, tokenIndex, depth = position380, tokenIndex380, depth380
								if buffer[position] != rune('M') {
									goto l372
								}
								position++
							}
						l380:
							{
								position382, tokenIndex382, depth382 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l383
								}
								position++
								goto l382
							l383:
								position, tokenIndex, depth = position382, tokenIndex382, depth382
								if buffer[position] != rune('P') {
									goto l372
								}
								position++
							}
						l382:
							{
								position384, tokenIndex384, depth384 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l385
								}
								position++
								goto l384
							l385:
								position, tokenIndex, depth = position384, tokenIndex384, depth384
								if buffer[position] != rune('L') {
									goto l372
								}
								position++
							}
						l384:
							{
								position386, tokenIndex386, depth386 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l387
								}
								position++
								goto l386
							l387:
								position, tokenIndex, depth = position386, tokenIndex386, depth386
								if buffer[position] != rune('E') {
									goto l372
								}
								position++
							}
						l386:
							depth--
							add(rulePegText, position375)
						}
						if !_rules[rule__]() {
							goto l372
						}
						{
							position388, tokenIndex388, depth388 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l389
							}
							position++
							goto l388
						l389:
							position, tokenIndex, depth = position388, tokenIndex388, depth388
							if buffer[position] != rune('B') {
								goto l372
							}
							position++
						}
					l388:
						{
							position390, tokenIndex390, depth390 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l391
							}
							position++
							goto l390
						l391:
							position, tokenIndex, depth = position390, tokenIndex390, depth390
							if buffer[position] != rune('Y') {
								goto l372
							}
							position++
						}
					l390:
						break
					case 'R', 'r':
						{
							position392 := position
							depth++
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
									goto l372
								}
								position++
							}
						l393:
							{
								position395, tokenIndex395, depth395 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l396
								}
								position++
								goto l395
							l396:
								position, tokenIndex, depth = position395, tokenIndex395, depth395
								if buffer[position] != rune('E') {
									goto l372
								}
								position++
							}
						l395:
							{
								position397, tokenIndex397, depth397 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l398
								}
								position++
								goto l397
							l398:
								position, tokenIndex, depth = position397, tokenIndex397, depth397
								if buffer[position] != rune('S') {
									goto l372
								}
								position++
							}
						l397:
							{
								position399, tokenIndex399, depth399 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l400
								}
								position++
								goto l399
							l400:
								position, tokenIndex, depth = position399, tokenIndex399, depth399
								if buffer[position] != rune('O') {
									goto l372
								}
								position++
							}
						l399:
							{
								position401, tokenIndex401, depth401 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l402
								}
								position++
								goto l401
							l402:
								position, tokenIndex, depth = position401, tokenIndex401, depth401
								if buffer[position] != rune('L') {
									goto l372
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
									goto l372
								}
								position++
							}
						l403:
							{
								position405, tokenIndex405, depth405 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l406
								}
								position++
								goto l405
							l406:
								position, tokenIndex, depth = position405, tokenIndex405, depth405
								if buffer[position] != rune('T') {
									goto l372
								}
								position++
							}
						l405:
							{
								position407, tokenIndex407, depth407 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l408
								}
								position++
								goto l407
							l408:
								position, tokenIndex, depth = position407, tokenIndex407, depth407
								if buffer[position] != rune('I') {
									goto l372
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
									goto l372
								}
								position++
							}
						l409:
							{
								position411, tokenIndex411, depth411 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l412
								}
								position++
								goto l411
							l412:
								position, tokenIndex, depth = position411, tokenIndex411, depth411
								if buffer[position] != rune('N') {
									goto l372
								}
								position++
							}
						l411:
							depth--
							add(rulePegText, position392)
						}
						break
					case 'T', 't':
						{
							position413 := position
							depth++
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
									goto l372
								}
								position++
							}
						l414:
							{
								position416, tokenIndex416, depth416 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l417
								}
								position++
								goto l416
							l417:
								position, tokenIndex, depth = position416, tokenIndex416, depth416
								if buffer[position] != rune('O') {
									goto l372
								}
								position++
							}
						l416:
							depth--
							add(rulePegText, position413)
						}
						break
					default:
						{
							position418 := position
							depth++
							{
								position419, tokenIndex419, depth419 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l420
								}
								position++
								goto l419
							l420:
								position, tokenIndex, depth = position419, tokenIndex419, depth419
								if buffer[position] != rune('F') {
									goto l372
								}
								position++
							}
						l419:
							{
								position421, tokenIndex421, depth421 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l422
								}
								position++
								goto l421
							l422:
								position, tokenIndex, depth = position421, tokenIndex421, depth421
								if buffer[position] != rune('R') {
									goto l372
								}
								position++
							}
						l421:
							{
								position423, tokenIndex423, depth423 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l424
								}
								position++
								goto l423
							l424:
								position, tokenIndex, depth = position423, tokenIndex423, depth423
								if buffer[position] != rune('O') {
									goto l372
								}
								position++
							}
						l423:
							{
								position425, tokenIndex425, depth425 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l426
								}
								position++
								goto l425
							l426:
								position, tokenIndex, depth = position425, tokenIndex425, depth425
								if buffer[position] != rune('M') {
									goto l372
								}
								position++
							}
						l425:
							depth--
							add(rulePegText, position418)
						}
						break
					}
				}

				depth--
				add(rulePROPERTY_KEY, position373)
			}
			return true
		l372:
			position, tokenIndex, depth = position372, tokenIndex372, depth372
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
			position436, tokenIndex436, depth436 := position, tokenIndex, depth
			{
				position437 := position
				depth++
				{
					position438, tokenIndex438, depth438 := position, tokenIndex, depth
					if buffer[position] != rune('\'') {
						goto l439
					}
					position++
				l440:
					{
						position441, tokenIndex441, depth441 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l441
						}
						goto l440
					l441:
						position, tokenIndex, depth = position441, tokenIndex441, depth441
					}
					if buffer[position] != rune('\'') {
						goto l439
					}
					position++
					goto l438
				l439:
					position, tokenIndex, depth = position438, tokenIndex438, depth438
					if buffer[position] != rune('"') {
						goto l436
					}
					position++
				l442:
					{
						position443, tokenIndex443, depth443 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l443
						}
						goto l442
					l443:
						position, tokenIndex, depth = position443, tokenIndex443, depth443
					}
					if buffer[position] != rune('"') {
						goto l436
					}
					position++
				}
			l438:
				depth--
				add(ruleSTRING, position437)
			}
			return true
		l436:
			position, tokenIndex, depth = position436, tokenIndex436, depth436
			return false
		},
		/* 42 CHAR <- <(('\\' ESCAPE_CLASS) / (!ESCAPE_CLASS .))> */
		func() bool {
			position444, tokenIndex444, depth444 := position, tokenIndex, depth
			{
				position445 := position
				depth++
				{
					position446, tokenIndex446, depth446 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l447
					}
					position++
					if !_rules[ruleESCAPE_CLASS]() {
						goto l447
					}
					goto l446
				l447:
					position, tokenIndex, depth = position446, tokenIndex446, depth446
					{
						position448, tokenIndex448, depth448 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l448
						}
						goto l444
					l448:
						position, tokenIndex, depth = position448, tokenIndex448, depth448
					}
					if !matchDot() {
						goto l444
					}
				}
			l446:
				depth--
				add(ruleCHAR, position445)
			}
			return true
		l444:
			position, tokenIndex, depth = position444, tokenIndex444, depth444
			return false
		},
		/* 43 ESCAPE_CLASS <- <((&('\\') '\\') | (&('"') '"') | (&('`') '`') | (&('\'') '\''))> */
		func() bool {
			position449, tokenIndex449, depth449 := position, tokenIndex, depth
			{
				position450 := position
				depth++
				{
					switch buffer[position] {
					case '\\':
						if buffer[position] != rune('\\') {
							goto l449
						}
						position++
						break
					case '"':
						if buffer[position] != rune('"') {
							goto l449
						}
						position++
						break
					case '`':
						if buffer[position] != rune('`') {
							goto l449
						}
						position++
						break
					default:
						if buffer[position] != rune('\'') {
							goto l449
						}
						position++
						break
					}
				}

				depth--
				add(ruleESCAPE_CLASS, position450)
			}
			return true
		l449:
			position, tokenIndex, depth = position449, tokenIndex449, depth449
			return false
		},
		/* 44 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position452, tokenIndex452, depth452 := position, tokenIndex, depth
			{
				position453 := position
				depth++
				{
					position454 := position
					depth++
					{
						position455, tokenIndex455, depth455 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l455
						}
						position++
						goto l456
					l455:
						position, tokenIndex, depth = position455, tokenIndex455, depth455
					}
				l456:
					{
						position457 := position
						depth++
						{
							position458, tokenIndex458, depth458 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l459
							}
							position++
							goto l458
						l459:
							position, tokenIndex, depth = position458, tokenIndex458, depth458
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l452
							}
							position++
						l460:
							{
								position461, tokenIndex461, depth461 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l461
								}
								position++
								goto l460
							l461:
								position, tokenIndex, depth = position461, tokenIndex461, depth461
							}
						}
					l458:
						depth--
						add(ruleNUMBER_NATURAL, position457)
					}
					depth--
					add(ruleNUMBER_INTEGER, position454)
				}
				{
					position462, tokenIndex462, depth462 := position, tokenIndex, depth
					{
						position464 := position
						depth++
						if buffer[position] != rune('.') {
							goto l462
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l462
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
						depth--
						add(ruleNUMBER_FRACTION, position464)
					}
					goto l463
				l462:
					position, tokenIndex, depth = position462, tokenIndex462, depth462
				}
			l463:
				{
					position467, tokenIndex467, depth467 := position, tokenIndex, depth
					{
						position469 := position
						depth++
						{
							position470, tokenIndex470, depth470 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l471
							}
							position++
							goto l470
						l471:
							position, tokenIndex, depth = position470, tokenIndex470, depth470
							if buffer[position] != rune('E') {
								goto l467
							}
							position++
						}
					l470:
						{
							position472, tokenIndex472, depth472 := position, tokenIndex, depth
							{
								position474, tokenIndex474, depth474 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l475
								}
								position++
								goto l474
							l475:
								position, tokenIndex, depth = position474, tokenIndex474, depth474
								if buffer[position] != rune('-') {
									goto l472
								}
								position++
							}
						l474:
							goto l473
						l472:
							position, tokenIndex, depth = position472, tokenIndex472, depth472
						}
					l473:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l467
						}
						position++
					l476:
						{
							position477, tokenIndex477, depth477 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l477
							}
							position++
							goto l476
						l477:
							position, tokenIndex, depth = position477, tokenIndex477, depth477
						}
						depth--
						add(ruleNUMBER_EXP, position469)
					}
					goto l468
				l467:
					position, tokenIndex, depth = position467, tokenIndex467, depth467
				}
			l468:
				depth--
				add(ruleNUMBER, position453)
			}
			return true
		l452:
			position, tokenIndex, depth = position452, tokenIndex452, depth452
			return false
		},
		/* 45 NUMBER_NATURAL <- <('0' / ([1-9] [0-9]*))> */
		nil,
		/* 46 NUMBER_FRACTION <- <('.' [0-9]+)> */
		nil,
		/* 47 NUMBER_INTEGER <- <('-'? NUMBER_NATURAL)> */
		nil,
		/* 48 NUMBER_EXP <- <(('e' / 'E') ('+' / '-')? [0-9]+)> */
		nil,
		/* 49 PAREN_OPEN <- <(_ '(' _)> */
		func() bool {
			position482, tokenIndex482, depth482 := position, tokenIndex, depth
			{
				position483 := position
				depth++
				if !_rules[rule_]() {
					goto l482
				}
				if buffer[position] != rune('(') {
					goto l482
				}
				position++
				if !_rules[rule_]() {
					goto l482
				}
				depth--
				add(rulePAREN_OPEN, position483)
			}
			return true
		l482:
			position, tokenIndex, depth = position482, tokenIndex482, depth482
			return false
		},
		/* 50 PAREN_CLOSE <- <(_ ')' _)> */
		func() bool {
			position484, tokenIndex484, depth484 := position, tokenIndex, depth
			{
				position485 := position
				depth++
				if !_rules[rule_]() {
					goto l484
				}
				if buffer[position] != rune(')') {
					goto l484
				}
				position++
				if !_rules[rule_]() {
					goto l484
				}
				depth--
				add(rulePAREN_CLOSE, position485)
			}
			return true
		l484:
			position, tokenIndex, depth = position484, tokenIndex484, depth484
			return false
		},
		/* 51 COMMA <- <(_ ',' _)> */
		func() bool {
			position486, tokenIndex486, depth486 := position, tokenIndex, depth
			{
				position487 := position
				depth++
				if !_rules[rule_]() {
					goto l486
				}
				if buffer[position] != rune(',') {
					goto l486
				}
				position++
				if !_rules[rule_]() {
					goto l486
				}
				depth--
				add(ruleCOMMA, position487)
			}
			return true
		l486:
			position, tokenIndex, depth = position486, tokenIndex486, depth486
			return false
		},
		/* 52 _ <- <SPACE*> */
		func() bool {
			{
				position489 := position
				depth++
			l490:
				{
					position491, tokenIndex491, depth491 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l491
					}
					goto l490
				l491:
					position, tokenIndex, depth = position491, tokenIndex491, depth491
				}
				depth--
				add(rule_, position489)
			}
			return true
		},
		/* 53 __ <- <SPACE+> */
		func() bool {
			position492, tokenIndex492, depth492 := position, tokenIndex, depth
			{
				position493 := position
				depth++
				if !_rules[ruleSPACE]() {
					goto l492
				}
			l494:
				{
					position495, tokenIndex495, depth495 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l495
					}
					goto l494
				l495:
					position, tokenIndex, depth = position495, tokenIndex495, depth495
				}
				depth--
				add(rule__, position493)
			}
			return true
		l492:
			position, tokenIndex, depth = position492, tokenIndex492, depth492
			return false
		},
		/* 54 SPACE <- <((&('\t') '\t') | (&('\n') '\n') | (&(' ') ' '))> */
		func() bool {
			position496, tokenIndex496, depth496 := position, tokenIndex, depth
			{
				position497 := position
				depth++
				{
					switch buffer[position] {
					case '\t':
						if buffer[position] != rune('\t') {
							goto l496
						}
						position++
						break
					case '\n':
						if buffer[position] != rune('\n') {
							goto l496
						}
						position++
						break
					default:
						if buffer[position] != rune(' ') {
							goto l496
						}
						position++
						break
					}
				}

				depth--
				add(ruleSPACE, position497)
			}
			return true
		l496:
			position, tokenIndex, depth = position496, tokenIndex496, depth496
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
		/* 61 Action4 <- <{ p.addEvaluationContext() }> */
		nil,
		/* 62 Action5 <- <{ p.addPropertyKey(buffer[begin:end])   }> */
		nil,
		/* 63 Action6 <- <{ p.addPropertyValue(buffer[begin:end]) }> */
		nil,
		/* 64 Action7 <- <{ p.insertPropertyKeyValue() }> */
		nil,
		/* 65 Action8 <- <{ p.checkPropertyClause() }> */
		nil,
		/* 66 Action9 <- <{ p.addNullPredicate() }> */
		nil,
		/* 67 Action10 <- <{ p.addExpressionList() }> */
		nil,
		/* 68 Action11 <- <{ p.appendExpression() }> */
		nil,
		/* 69 Action12 <- <{ p.appendExpression() }> */
		nil,
		/* 70 Action13 <- <{ p.addOperatorLiteral("*") }> */
		nil,
		/* 71 Action14 <- <{ p.addOperatorLiteral("-") }> */
		nil,
		/* 72 Action15 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 73 Action16 <- <{ p.addOperatorLiteral("*") }> */
		nil,
		/* 74 Action17 <- <{ p.addOperatorLiteral("*") }> */
		nil,
		/* 75 Action18 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 76 Action19 <- <{ p.addNumberNode(buffer[begin:end]) }> */
		nil,
		/* 77 Action20 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 78 Action21 <- <{ p.addGroupBy() }> */
		nil,
		/* 79 Action22 <- <{
		   p.addFunctionInvocation()
		 }> */
		nil,
		/* 80 Action23 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 81 Action24 <- <{ p.addNullPredicate() }> */
		nil,
		/* 82 Action25 <- <{
		   p.addMetricExpression()
		 }> */
		nil,
		/* 83 Action26 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 84 Action27 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		   }> */
		nil,
		/* 85 Action28 <- <{ p.addOrPredicate() }> */
		nil,
		/* 86 Action29 <- <{ p.addAndPredicate() }> */
		nil,
		/* 87 Action30 <- <{ p.addNotPredicate() }> */
		nil,
		/* 88 Action31 <- <{
		   p.addLiteralMatcher()
		 }> */
		nil,
		/* 89 Action32 <- <{
		   p.addLiteralMatcher()
		   p.addNotPredicate()
		 }> */
		nil,
		/* 90 Action33 <- <{
		   p.addRegexMatcher()
		 }> */
		nil,
		/* 91 Action34 <- <{
		   p.addListMatcher()
		 }> */
		nil,
		/* 92 Action35 <- <{
		  p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 93 Action36 <- <{ p.addLiteralList() }> */
		nil,
		/* 94 Action37 <- <{
		  p.appendLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 95 Action38 <- <{ p.addTagLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
	}
	p.rules = _rules
}
