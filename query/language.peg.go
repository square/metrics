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
	rules  [101]func() bool
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
			p.addNumberNode(buffer[begin:end])
		case ruleAction21:
			p.addStringNode(unescapeLiteral(buffer[begin:end]))
		case ruleAction22:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction23:
			p.addGroupBy()
		case ruleAction24:

			p.addFunctionInvocation()

		case ruleAction25:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction26:
			p.addNullPredicate()
		case ruleAction27:

			p.addMetricExpression()

		case ruleAction28:

			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))

		case ruleAction29:

			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))

		case ruleAction30:
			p.addOrPredicate()
		case ruleAction31:
			p.addAndPredicate()
		case ruleAction32:
			p.addNotPredicate()
		case ruleAction33:

			p.addLiteralMatcher()

		case ruleAction34:

			p.addLiteralMatcher()
			p.addNotPredicate()

		case ruleAction35:

			p.addRegexMatcher()

		case ruleAction36:

			p.addListMatcher()

		case ruleAction37:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction38:
			p.addLiteralList()
		case ruleAction39:

			p.appendLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction40:
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
						if !_rules[ruleKEY]() {
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
								add(ruleAction5, position)
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
									add(ruleAction6, position)
								}
								if !_rules[rule_]() {
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
											if !_rules[rule_]() {
												goto l25
											}
											{
												position26 := position
												depth++
												if !_rules[ruleNUMBER]() {
													goto l25
												}
											l27:
												{
													position28, tokenIndex28, depth28 := position, tokenIndex, depth
													{
														position29, tokenIndex29, depth29 := position, tokenIndex, depth
														if c := buffer[position]; c < rune('a') || c > rune('z') {
															goto l30
														}
														position++
														goto l29
													l30:
														position, tokenIndex, depth = position29, tokenIndex29, depth29
														if c := buffer[position]; c < rune('A') || c > rune('Z') {
															goto l28
														}
														position++
													}
												l29:
													goto l27
												l28:
													position, tokenIndex, depth = position28, tokenIndex28, depth28
												}
												depth--
												add(rulePegText, position26)
											}
											goto l24
										l25:
											position, tokenIndex, depth = position24, tokenIndex24, depth24
											if !_rules[rule_]() {
												goto l31
											}
											if !_rules[ruleSTRING]() {
												goto l31
											}
											goto l24
										l31:
											position, tokenIndex, depth = position24, tokenIndex24, depth24
											if !_rules[rule_]() {
												goto l20
											}
											{
												position32 := position
												depth++
												{
													position33, tokenIndex33, depth33 := position, tokenIndex, depth
													if buffer[position] != rune('n') {
														goto l34
													}
													position++
													goto l33
												l34:
													position, tokenIndex, depth = position33, tokenIndex33, depth33
													if buffer[position] != rune('N') {
														goto l20
													}
													position++
												}
											l33:
												{
													position35, tokenIndex35, depth35 := position, tokenIndex, depth
													if buffer[position] != rune('o') {
														goto l36
													}
													position++
													goto l35
												l36:
													position, tokenIndex, depth = position35, tokenIndex35, depth35
													if buffer[position] != rune('O') {
														goto l20
													}
													position++
												}
											l35:
												{
													position37, tokenIndex37, depth37 := position, tokenIndex, depth
													if buffer[position] != rune('w') {
														goto l38
													}
													position++
													goto l37
												l38:
													position, tokenIndex, depth = position37, tokenIndex37, depth37
													if buffer[position] != rune('W') {
														goto l20
													}
													position++
												}
											l37:
												depth--
												add(rulePegText, position32)
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
									add(ruleAction7, position)
								}
								{
									add(ruleAction8, position)
								}
								goto l19
							l20:
								position, tokenIndex, depth = position20, tokenIndex20, depth20
							}
							{
								add(ruleAction9, position)
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
						position43 := position
						depth++
						if !_rules[rule_]() {
							goto l0
						}
						{
							position44, tokenIndex44, depth44 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l45
							}
							position++
							goto l44
						l45:
							position, tokenIndex, depth = position44, tokenIndex44, depth44
							if buffer[position] != rune('D') {
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
						{
							position48, tokenIndex48, depth48 := position, tokenIndex, depth
							if buffer[position] != rune('s') {
								goto l49
							}
							position++
							goto l48
						l49:
							position, tokenIndex, depth = position48, tokenIndex48, depth48
							if buffer[position] != rune('S') {
								goto l0
							}
							position++
						}
					l48:
						{
							position50, tokenIndex50, depth50 := position, tokenIndex, depth
							if buffer[position] != rune('c') {
								goto l51
							}
							position++
							goto l50
						l51:
							position, tokenIndex, depth = position50, tokenIndex50, depth50
							if buffer[position] != rune('C') {
								goto l0
							}
							position++
						}
					l50:
						{
							position52, tokenIndex52, depth52 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l53
							}
							position++
							goto l52
						l53:
							position, tokenIndex, depth = position52, tokenIndex52, depth52
							if buffer[position] != rune('R') {
								goto l0
							}
							position++
						}
					l52:
						{
							position54, tokenIndex54, depth54 := position, tokenIndex, depth
							if buffer[position] != rune('i') {
								goto l55
							}
							position++
							goto l54
						l55:
							position, tokenIndex, depth = position54, tokenIndex54, depth54
							if buffer[position] != rune('I') {
								goto l0
							}
							position++
						}
					l54:
						{
							position56, tokenIndex56, depth56 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l57
							}
							position++
							goto l56
						l57:
							position, tokenIndex, depth = position56, tokenIndex56, depth56
							if buffer[position] != rune('B') {
								goto l0
							}
							position++
						}
					l56:
						{
							position58, tokenIndex58, depth58 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l59
							}
							position++
							goto l58
						l59:
							position, tokenIndex, depth = position58, tokenIndex58, depth58
							if buffer[position] != rune('E') {
								goto l0
							}
							position++
						}
					l58:
						if !_rules[ruleKEY]() {
							goto l0
						}
						{
							position60, tokenIndex60, depth60 := position, tokenIndex, depth
							{
								position62 := position
								depth++
								if !_rules[rule_]() {
									goto l61
								}
								{
									position63, tokenIndex63, depth63 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l64
									}
									position++
									goto l63
								l64:
									position, tokenIndex, depth = position63, tokenIndex63, depth63
									if buffer[position] != rune('A') {
										goto l61
									}
									position++
								}
							l63:
								{
									position65, tokenIndex65, depth65 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l66
									}
									position++
									goto l65
								l66:
									position, tokenIndex, depth = position65, tokenIndex65, depth65
									if buffer[position] != rune('L') {
										goto l61
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
										goto l61
									}
									position++
								}
							l67:
								if !_rules[ruleKEY]() {
									goto l61
								}
								{
									add(ruleAction1, position)
								}
								depth--
								add(ruledescribeAllStmt, position62)
							}
							goto l60
						l61:
							position, tokenIndex, depth = position60, tokenIndex60, depth60
							{
								position71 := position
								depth++
								if !_rules[rule_]() {
									goto l70
								}
								{
									position72, tokenIndex72, depth72 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l73
									}
									position++
									goto l72
								l73:
									position, tokenIndex, depth = position72, tokenIndex72, depth72
									if buffer[position] != rune('M') {
										goto l70
									}
									position++
								}
							l72:
								{
									position74, tokenIndex74, depth74 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l75
									}
									position++
									goto l74
								l75:
									position, tokenIndex, depth = position74, tokenIndex74, depth74
									if buffer[position] != rune('E') {
										goto l70
									}
									position++
								}
							l74:
								{
									position76, tokenIndex76, depth76 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l77
									}
									position++
									goto l76
								l77:
									position, tokenIndex, depth = position76, tokenIndex76, depth76
									if buffer[position] != rune('T') {
										goto l70
									}
									position++
								}
							l76:
								{
									position78, tokenIndex78, depth78 := position, tokenIndex, depth
									if buffer[position] != rune('r') {
										goto l79
									}
									position++
									goto l78
								l79:
									position, tokenIndex, depth = position78, tokenIndex78, depth78
									if buffer[position] != rune('R') {
										goto l70
									}
									position++
								}
							l78:
								{
									position80, tokenIndex80, depth80 := position, tokenIndex, depth
									if buffer[position] != rune('i') {
										goto l81
									}
									position++
									goto l80
								l81:
									position, tokenIndex, depth = position80, tokenIndex80, depth80
									if buffer[position] != rune('I') {
										goto l70
									}
									position++
								}
							l80:
								{
									position82, tokenIndex82, depth82 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l83
									}
									position++
									goto l82
								l83:
									position, tokenIndex, depth = position82, tokenIndex82, depth82
									if buffer[position] != rune('C') {
										goto l70
									}
									position++
								}
							l82:
								{
									position84, tokenIndex84, depth84 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l85
									}
									position++
									goto l84
								l85:
									position, tokenIndex, depth = position84, tokenIndex84, depth84
									if buffer[position] != rune('S') {
										goto l70
									}
									position++
								}
							l84:
								if !_rules[ruleKEY]() {
									goto l70
								}
								if !_rules[rule_]() {
									goto l70
								}
								{
									position86, tokenIndex86, depth86 := position, tokenIndex, depth
									if buffer[position] != rune('w') {
										goto l87
									}
									position++
									goto l86
								l87:
									position, tokenIndex, depth = position86, tokenIndex86, depth86
									if buffer[position] != rune('W') {
										goto l70
									}
									position++
								}
							l86:
								{
									position88, tokenIndex88, depth88 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l89
									}
									position++
									goto l88
								l89:
									position, tokenIndex, depth = position88, tokenIndex88, depth88
									if buffer[position] != rune('H') {
										goto l70
									}
									position++
								}
							l88:
								{
									position90, tokenIndex90, depth90 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l91
									}
									position++
									goto l90
								l91:
									position, tokenIndex, depth = position90, tokenIndex90, depth90
									if buffer[position] != rune('E') {
										goto l70
									}
									position++
								}
							l90:
								{
									position92, tokenIndex92, depth92 := position, tokenIndex, depth
									if buffer[position] != rune('r') {
										goto l93
									}
									position++
									goto l92
								l93:
									position, tokenIndex, depth = position92, tokenIndex92, depth92
									if buffer[position] != rune('R') {
										goto l70
									}
									position++
								}
							l92:
								{
									position94, tokenIndex94, depth94 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l95
									}
									position++
									goto l94
								l95:
									position, tokenIndex, depth = position94, tokenIndex94, depth94
									if buffer[position] != rune('E') {
										goto l70
									}
									position++
								}
							l94:
								if !_rules[ruleKEY]() {
									goto l70
								}
								if !_rules[ruletagName]() {
									goto l70
								}
								if !_rules[rule_]() {
									goto l70
								}
								if buffer[position] != rune('=') {
									goto l70
								}
								position++
								if !_rules[ruleliteralString]() {
									goto l70
								}
								{
									add(ruleAction2, position)
								}
								depth--
								add(ruledescribeMetrics, position71)
							}
							goto l60
						l70:
							position, tokenIndex, depth = position60, tokenIndex60, depth60
							{
								position97 := position
								depth++
								if !_rules[rule_]() {
									goto l0
								}
								{
									position98 := position
									depth++
									{
										position99 := position
										depth++
										if !_rules[rule_]() {
											goto l0
										}
										if !_rules[ruleIDENTIFIER]() {
											goto l0
										}
										depth--
										add(ruleMETRIC_NAME, position99)
									}
									depth--
									add(rulePegText, position98)
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
								add(ruledescribeSingleStmt, position97)
							}
						}
					l60:
						depth--
						add(ruledescribeStmt, position43)
					}
				}
			l2:
				if !_rules[rule_]() {
					goto l0
				}
				{
					position102, tokenIndex102, depth102 := position, tokenIndex, depth
					if !matchDot() {
						goto l102
					}
					goto l0
				l102:
					position, tokenIndex, depth = position102, tokenIndex102, depth102
				}
				depth--
				add(ruleroot, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 selectStmt <- <(_ (('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T')) KEY expressionList optionalPredicateClause propertyClause Action0)> */
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
				position110 := position
				depth++
				{
					position111, tokenIndex111, depth111 := position, tokenIndex, depth
					{
						position113 := position
						depth++
						if !_rules[rule_]() {
							goto l112
						}
						{
							position114, tokenIndex114, depth114 := position, tokenIndex, depth
							if buffer[position] != rune('w') {
								goto l115
							}
							position++
							goto l114
						l115:
							position, tokenIndex, depth = position114, tokenIndex114, depth114
							if buffer[position] != rune('W') {
								goto l112
							}
							position++
						}
					l114:
						{
							position116, tokenIndex116, depth116 := position, tokenIndex, depth
							if buffer[position] != rune('h') {
								goto l117
							}
							position++
							goto l116
						l117:
							position, tokenIndex, depth = position116, tokenIndex116, depth116
							if buffer[position] != rune('H') {
								goto l112
							}
							position++
						}
					l116:
						{
							position118, tokenIndex118, depth118 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l119
							}
							position++
							goto l118
						l119:
							position, tokenIndex, depth = position118, tokenIndex118, depth118
							if buffer[position] != rune('E') {
								goto l112
							}
							position++
						}
					l118:
						{
							position120, tokenIndex120, depth120 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l121
							}
							position++
							goto l120
						l121:
							position, tokenIndex, depth = position120, tokenIndex120, depth120
							if buffer[position] != rune('R') {
								goto l112
							}
							position++
						}
					l120:
						{
							position122, tokenIndex122, depth122 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l123
							}
							position++
							goto l122
						l123:
							position, tokenIndex, depth = position122, tokenIndex122, depth122
							if buffer[position] != rune('E') {
								goto l112
							}
							position++
						}
					l122:
						if !_rules[ruleKEY]() {
							goto l112
						}
						if !_rules[rule_]() {
							goto l112
						}
						if !_rules[rulepredicate_1]() {
							goto l112
						}
						depth--
						add(rulepredicateClause, position113)
					}
					goto l111
				l112:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					{
						add(ruleAction10, position)
					}
				}
			l111:
				depth--
				add(ruleoptionalPredicateClause, position110)
			}
			return true
		},
		/* 8 expressionList <- <(Action11 expression_1 Action12 (_ COMMA expression_1 Action13)*)> */
		func() bool {
			position125, tokenIndex125, depth125 := position, tokenIndex, depth
			{
				position126 := position
				depth++
				{
					add(ruleAction11, position)
				}
				if !_rules[ruleexpression_1]() {
					goto l125
				}
				{
					add(ruleAction12, position)
				}
			l129:
				{
					position130, tokenIndex130, depth130 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l130
					}
					if !_rules[ruleCOMMA]() {
						goto l130
					}
					if !_rules[ruleexpression_1]() {
						goto l130
					}
					{
						add(ruleAction13, position)
					}
					goto l129
				l130:
					position, tokenIndex, depth = position130, tokenIndex130, depth130
				}
				depth--
				add(ruleexpressionList, position126)
			}
			return true
		l125:
			position, tokenIndex, depth = position125, tokenIndex125, depth125
			return false
		},
		/* 9 expression_1 <- <(expression_2 (((_ OP_ADD Action14) / (_ OP_SUB Action15)) expression_2 Action16)*)> */
		func() bool {
			position132, tokenIndex132, depth132 := position, tokenIndex, depth
			{
				position133 := position
				depth++
				if !_rules[ruleexpression_2]() {
					goto l132
				}
			l134:
				{
					position135, tokenIndex135, depth135 := position, tokenIndex, depth
					{
						position136, tokenIndex136, depth136 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l137
						}
						{
							position138 := position
							depth++
							if buffer[position] != rune('+') {
								goto l137
							}
							position++
							depth--
							add(ruleOP_ADD, position138)
						}
						{
							add(ruleAction14, position)
						}
						goto l136
					l137:
						position, tokenIndex, depth = position136, tokenIndex136, depth136
						if !_rules[rule_]() {
							goto l135
						}
						{
							position140 := position
							depth++
							if buffer[position] != rune('-') {
								goto l135
							}
							position++
							depth--
							add(ruleOP_SUB, position140)
						}
						{
							add(ruleAction15, position)
						}
					}
				l136:
					if !_rules[ruleexpression_2]() {
						goto l135
					}
					{
						add(ruleAction16, position)
					}
					goto l134
				l135:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
				}
				depth--
				add(ruleexpression_1, position133)
			}
			return true
		l132:
			position, tokenIndex, depth = position132, tokenIndex132, depth132
			return false
		},
		/* 10 expression_2 <- <(expression_3 (((_ OP_DIV Action17) / (_ OP_MULT Action18)) expression_3 Action19)*)> */
		func() bool {
			position143, tokenIndex143, depth143 := position, tokenIndex, depth
			{
				position144 := position
				depth++
				if !_rules[ruleexpression_3]() {
					goto l143
				}
			l145:
				{
					position146, tokenIndex146, depth146 := position, tokenIndex, depth
					{
						position147, tokenIndex147, depth147 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l148
						}
						{
							position149 := position
							depth++
							if buffer[position] != rune('/') {
								goto l148
							}
							position++
							depth--
							add(ruleOP_DIV, position149)
						}
						{
							add(ruleAction17, position)
						}
						goto l147
					l148:
						position, tokenIndex, depth = position147, tokenIndex147, depth147
						if !_rules[rule_]() {
							goto l146
						}
						{
							position151 := position
							depth++
							if buffer[position] != rune('*') {
								goto l146
							}
							position++
							depth--
							add(ruleOP_MULT, position151)
						}
						{
							add(ruleAction18, position)
						}
					}
				l147:
					if !_rules[ruleexpression_3]() {
						goto l146
					}
					{
						add(ruleAction19, position)
					}
					goto l145
				l146:
					position, tokenIndex, depth = position146, tokenIndex146, depth146
				}
				depth--
				add(ruleexpression_2, position144)
			}
			return true
		l143:
			position, tokenIndex, depth = position143, tokenIndex143, depth143
			return false
		},
		/* 11 expression_3 <- <(expression_function / expression_metric / (_ PAREN_OPEN expression_1 _ PAREN_CLOSE) / (_ <NUMBER> Action20) / (_ STRING Action21))> */
		func() bool {
			position154, tokenIndex154, depth154 := position, tokenIndex, depth
			{
				position155 := position
				depth++
				{
					position156, tokenIndex156, depth156 := position, tokenIndex, depth
					{
						position158 := position
						depth++
						if !_rules[rule_]() {
							goto l157
						}
						{
							position159 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l157
							}
							depth--
							add(rulePegText, position159)
						}
						{
							add(ruleAction22, position)
						}
						if !_rules[rule_]() {
							goto l157
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l157
						}
						if !_rules[ruleexpressionList]() {
							goto l157
						}
						{
							add(ruleAction23, position)
						}
						{
							position162, tokenIndex162, depth162 := position, tokenIndex, depth
							{
								position164 := position
								depth++
								if !_rules[rule_]() {
									goto l162
								}
								{
									position165, tokenIndex165, depth165 := position, tokenIndex, depth
									if buffer[position] != rune('g') {
										goto l166
									}
									position++
									goto l165
								l166:
									position, tokenIndex, depth = position165, tokenIndex165, depth165
									if buffer[position] != rune('G') {
										goto l162
									}
									position++
								}
							l165:
								{
									position167, tokenIndex167, depth167 := position, tokenIndex, depth
									if buffer[position] != rune('r') {
										goto l168
									}
									position++
									goto l167
								l168:
									position, tokenIndex, depth = position167, tokenIndex167, depth167
									if buffer[position] != rune('R') {
										goto l162
									}
									position++
								}
							l167:
								{
									position169, tokenIndex169, depth169 := position, tokenIndex, depth
									if buffer[position] != rune('o') {
										goto l170
									}
									position++
									goto l169
								l170:
									position, tokenIndex, depth = position169, tokenIndex169, depth169
									if buffer[position] != rune('O') {
										goto l162
									}
									position++
								}
							l169:
								{
									position171, tokenIndex171, depth171 := position, tokenIndex, depth
									if buffer[position] != rune('u') {
										goto l172
									}
									position++
									goto l171
								l172:
									position, tokenIndex, depth = position171, tokenIndex171, depth171
									if buffer[position] != rune('U') {
										goto l162
									}
									position++
								}
							l171:
								{
									position173, tokenIndex173, depth173 := position, tokenIndex, depth
									if buffer[position] != rune('p') {
										goto l174
									}
									position++
									goto l173
								l174:
									position, tokenIndex, depth = position173, tokenIndex173, depth173
									if buffer[position] != rune('P') {
										goto l162
									}
									position++
								}
							l173:
								if !_rules[ruleKEY]() {
									goto l162
								}
								if !_rules[rule_]() {
									goto l162
								}
								{
									position175, tokenIndex175, depth175 := position, tokenIndex, depth
									if buffer[position] != rune('b') {
										goto l176
									}
									position++
									goto l175
								l176:
									position, tokenIndex, depth = position175, tokenIndex175, depth175
									if buffer[position] != rune('B') {
										goto l162
									}
									position++
								}
							l175:
								{
									position177, tokenIndex177, depth177 := position, tokenIndex, depth
									if buffer[position] != rune('y') {
										goto l178
									}
									position++
									goto l177
								l178:
									position, tokenIndex, depth = position177, tokenIndex177, depth177
									if buffer[position] != rune('Y') {
										goto l162
									}
									position++
								}
							l177:
								if !_rules[ruleKEY]() {
									goto l162
								}
								if !_rules[rule_]() {
									goto l162
								}
								{
									position179 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l162
									}
									depth--
									add(rulePegText, position179)
								}
								{
									add(ruleAction28, position)
								}
							l181:
								{
									position182, tokenIndex182, depth182 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l182
									}
									if !_rules[ruleCOMMA]() {
										goto l182
									}
									if !_rules[rule_]() {
										goto l182
									}
									{
										position183 := position
										depth++
										if !_rules[ruleCOLUMN_NAME]() {
											goto l182
										}
										depth--
										add(rulePegText, position183)
									}
									{
										add(ruleAction29, position)
									}
									goto l181
								l182:
									position, tokenIndex, depth = position182, tokenIndex182, depth182
								}
								depth--
								add(rulegroupByClause, position164)
							}
							goto l163
						l162:
							position, tokenIndex, depth = position162, tokenIndex162, depth162
						}
					l163:
						if !_rules[rule_]() {
							goto l157
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l157
						}
						{
							add(ruleAction24, position)
						}
						depth--
						add(ruleexpression_function, position158)
					}
					goto l156
				l157:
					position, tokenIndex, depth = position156, tokenIndex156, depth156
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
							add(ruleAction25, position)
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
									add(ruleAction26, position)
								}
							}
						l192:
							goto l191

							position, tokenIndex, depth = position190, tokenIndex190, depth190
						}
					l191:
						{
							add(ruleAction27, position)
						}
						depth--
						add(ruleexpression_metric, position187)
					}
					goto l156
				l186:
					position, tokenIndex, depth = position156, tokenIndex156, depth156
					if !_rules[rule_]() {
						goto l196
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l196
					}
					if !_rules[ruleexpression_1]() {
						goto l196
					}
					if !_rules[rule_]() {
						goto l196
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l196
					}
					goto l156
				l196:
					position, tokenIndex, depth = position156, tokenIndex156, depth156
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
						add(ruleAction20, position)
					}
					goto l156
				l197:
					position, tokenIndex, depth = position156, tokenIndex156, depth156
					if !_rules[rule_]() {
						goto l154
					}
					if !_rules[ruleSTRING]() {
						goto l154
					}
					{
						add(ruleAction21, position)
					}
				}
			l156:
				depth--
				add(ruleexpression_3, position155)
			}
			return true
		l154:
			position, tokenIndex, depth = position154, tokenIndex154, depth154
			return false
		},
		/* 12 expression_function <- <(_ <IDENTIFIER> Action22 _ PAREN_OPEN expressionList Action23 groupByClause? _ PAREN_CLOSE Action24)> */
		nil,
		/* 13 expression_metric <- <(_ <IDENTIFIER> Action25 ((_ '[' predicate_1 _ ']') / Action26)? Action27)> */
		nil,
		/* 14 groupByClause <- <(_ (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P')) KEY _ (('b' / 'B') ('y' / 'Y')) KEY _ <COLUMN_NAME> Action28 (_ COMMA _ <COLUMN_NAME> Action29)*)> */
		nil,
		/* 15 predicateClause <- <(_ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY _ predicate_1)> */
		nil,
		/* 16 predicate_1 <- <((predicate_2 _ OP_OR predicate_1 Action30) / predicate_2 / )> */
		func() bool {
			{
				position206 := position
				depth++
				{
					position207, tokenIndex207, depth207 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l208
					}
					if !_rules[rule_]() {
						goto l208
					}
					{
						position209 := position
						depth++
						{
							position210, tokenIndex210, depth210 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l211
							}
							position++
							goto l210
						l211:
							position, tokenIndex, depth = position210, tokenIndex210, depth210
							if buffer[position] != rune('O') {
								goto l208
							}
							position++
						}
					l210:
						{
							position212, tokenIndex212, depth212 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l213
							}
							position++
							goto l212
						l213:
							position, tokenIndex, depth = position212, tokenIndex212, depth212
							if buffer[position] != rune('R') {
								goto l208
							}
							position++
						}
					l212:
						if !_rules[ruleKEY]() {
							goto l208
						}
						depth--
						add(ruleOP_OR, position209)
					}
					if !_rules[rulepredicate_1]() {
						goto l208
					}
					{
						add(ruleAction30, position)
					}
					goto l207
				l208:
					position, tokenIndex, depth = position207, tokenIndex207, depth207
					if !_rules[rulepredicate_2]() {
						goto l215
					}
					goto l207
				l215:
					position, tokenIndex, depth = position207, tokenIndex207, depth207
				}
			l207:
				depth--
				add(rulepredicate_1, position206)
			}
			return true
		},
		/* 17 predicate_2 <- <((predicate_3 _ OP_AND predicate_2 Action31) / predicate_3)> */
		func() bool {
			position216, tokenIndex216, depth216 := position, tokenIndex, depth
			{
				position217 := position
				depth++
				{
					position218, tokenIndex218, depth218 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l219
					}
					if !_rules[rule_]() {
						goto l219
					}
					{
						position220 := position
						depth++
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
								goto l219
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
								goto l219
							}
							position++
						}
					l223:
						{
							position225, tokenIndex225, depth225 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l226
							}
							position++
							goto l225
						l226:
							position, tokenIndex, depth = position225, tokenIndex225, depth225
							if buffer[position] != rune('D') {
								goto l219
							}
							position++
						}
					l225:
						if !_rules[ruleKEY]() {
							goto l219
						}
						depth--
						add(ruleOP_AND, position220)
					}
					if !_rules[rulepredicate_2]() {
						goto l219
					}
					{
						add(ruleAction31, position)
					}
					goto l218
				l219:
					position, tokenIndex, depth = position218, tokenIndex218, depth218
					if !_rules[rulepredicate_3]() {
						goto l216
					}
				}
			l218:
				depth--
				add(rulepredicate_2, position217)
			}
			return true
		l216:
			position, tokenIndex, depth = position216, tokenIndex216, depth216
			return false
		},
		/* 18 predicate_3 <- <((_ OP_NOT predicate_3 Action32) / (_ PAREN_OPEN predicate_1 _ PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position228, tokenIndex228, depth228 := position, tokenIndex, depth
			{
				position229 := position
				depth++
				{
					position230, tokenIndex230, depth230 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l231
					}
					{
						position232 := position
						depth++
						{
							position233, tokenIndex233, depth233 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l234
							}
							position++
							goto l233
						l234:
							position, tokenIndex, depth = position233, tokenIndex233, depth233
							if buffer[position] != rune('N') {
								goto l231
							}
							position++
						}
					l233:
						{
							position235, tokenIndex235, depth235 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l236
							}
							position++
							goto l235
						l236:
							position, tokenIndex, depth = position235, tokenIndex235, depth235
							if buffer[position] != rune('O') {
								goto l231
							}
							position++
						}
					l235:
						{
							position237, tokenIndex237, depth237 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l238
							}
							position++
							goto l237
						l238:
							position, tokenIndex, depth = position237, tokenIndex237, depth237
							if buffer[position] != rune('T') {
								goto l231
							}
							position++
						}
					l237:
						if !_rules[ruleKEY]() {
							goto l231
						}
						depth--
						add(ruleOP_NOT, position232)
					}
					if !_rules[rulepredicate_3]() {
						goto l231
					}
					{
						add(ruleAction32, position)
					}
					goto l230
				l231:
					position, tokenIndex, depth = position230, tokenIndex230, depth230
					if !_rules[rule_]() {
						goto l240
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l240
					}
					if !_rules[rulepredicate_1]() {
						goto l240
					}
					if !_rules[rule_]() {
						goto l240
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l240
					}
					goto l230
				l240:
					position, tokenIndex, depth = position230, tokenIndex230, depth230
					{
						position241 := position
						depth++
						{
							position242, tokenIndex242, depth242 := position, tokenIndex, depth
							if !_rules[ruletagName]() {
								goto l243
							}
							if !_rules[rule_]() {
								goto l243
							}
							if buffer[position] != rune('=') {
								goto l243
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l243
							}
							{
								add(ruleAction33, position)
							}
							goto l242
						l243:
							position, tokenIndex, depth = position242, tokenIndex242, depth242
							if !_rules[ruletagName]() {
								goto l245
							}
							if !_rules[rule_]() {
								goto l245
							}
							if buffer[position] != rune('!') {
								goto l245
							}
							position++
							if buffer[position] != rune('=') {
								goto l245
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l245
							}
							{
								add(ruleAction34, position)
							}
							goto l242
						l245:
							position, tokenIndex, depth = position242, tokenIndex242, depth242
							if !_rules[ruletagName]() {
								goto l247
							}
							if !_rules[rule_]() {
								goto l247
							}
							{
								position248, tokenIndex248, depth248 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l249
								}
								position++
								goto l248
							l249:
								position, tokenIndex, depth = position248, tokenIndex248, depth248
								if buffer[position] != rune('M') {
									goto l247
								}
								position++
							}
						l248:
							{
								position250, tokenIndex250, depth250 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l251
								}
								position++
								goto l250
							l251:
								position, tokenIndex, depth = position250, tokenIndex250, depth250
								if buffer[position] != rune('A') {
									goto l247
								}
								position++
							}
						l250:
							{
								position252, tokenIndex252, depth252 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l253
								}
								position++
								goto l252
							l253:
								position, tokenIndex, depth = position252, tokenIndex252, depth252
								if buffer[position] != rune('T') {
									goto l247
								}
								position++
							}
						l252:
							{
								position254, tokenIndex254, depth254 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l255
								}
								position++
								goto l254
							l255:
								position, tokenIndex, depth = position254, tokenIndex254, depth254
								if buffer[position] != rune('C') {
									goto l247
								}
								position++
							}
						l254:
							{
								position256, tokenIndex256, depth256 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l257
								}
								position++
								goto l256
							l257:
								position, tokenIndex, depth = position256, tokenIndex256, depth256
								if buffer[position] != rune('H') {
									goto l247
								}
								position++
							}
						l256:
							{
								position258, tokenIndex258, depth258 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l259
								}
								position++
								goto l258
							l259:
								position, tokenIndex, depth = position258, tokenIndex258, depth258
								if buffer[position] != rune('E') {
									goto l247
								}
								position++
							}
						l258:
							{
								position260, tokenIndex260, depth260 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l261
								}
								position++
								goto l260
							l261:
								position, tokenIndex, depth = position260, tokenIndex260, depth260
								if buffer[position] != rune('S') {
									goto l247
								}
								position++
							}
						l260:
							if !_rules[ruleKEY]() {
								goto l247
							}
							if !_rules[ruleliteralString]() {
								goto l247
							}
							{
								add(ruleAction35, position)
							}
							goto l242
						l247:
							position, tokenIndex, depth = position242, tokenIndex242, depth242
							if !_rules[ruletagName]() {
								goto l228
							}
							if !_rules[rule_]() {
								goto l228
							}
							{
								position263, tokenIndex263, depth263 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l264
								}
								position++
								goto l263
							l264:
								position, tokenIndex, depth = position263, tokenIndex263, depth263
								if buffer[position] != rune('I') {
									goto l228
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
									goto l228
								}
								position++
							}
						l265:
							if !_rules[ruleKEY]() {
								goto l228
							}
							{
								position267 := position
								depth++
								{
									add(ruleAction38, position)
								}
								if !_rules[rule_]() {
									goto l228
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l228
								}
								if !_rules[ruleliteralListString]() {
									goto l228
								}
							l269:
								{
									position270, tokenIndex270, depth270 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l270
									}
									if !_rules[ruleCOMMA]() {
										goto l270
									}
									if !_rules[ruleliteralListString]() {
										goto l270
									}
									goto l269
								l270:
									position, tokenIndex, depth = position270, tokenIndex270, depth270
								}
								if !_rules[rule_]() {
									goto l228
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l228
								}
								depth--
								add(ruleliteralList, position267)
							}
							{
								add(ruleAction36, position)
							}
						}
					l242:
						depth--
						add(ruletagMatcher, position241)
					}
				}
			l230:
				depth--
				add(rulepredicate_3, position229)
			}
			return true
		l228:
			position, tokenIndex, depth = position228, tokenIndex228, depth228
			return false
		},
		/* 19 tagMatcher <- <((tagName _ '=' literalString Action33) / (tagName _ ('!' '=') literalString Action34) / (tagName _ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) KEY literalString Action35) / (tagName _ (('i' / 'I') ('n' / 'N')) KEY literalList Action36))> */
		nil,
		/* 20 literalString <- <(_ STRING Action37)> */
		func() bool {
			position273, tokenIndex273, depth273 := position, tokenIndex, depth
			{
				position274 := position
				depth++
				if !_rules[rule_]() {
					goto l273
				}
				if !_rules[ruleSTRING]() {
					goto l273
				}
				{
					add(ruleAction37, position)
				}
				depth--
				add(ruleliteralString, position274)
			}
			return true
		l273:
			position, tokenIndex, depth = position273, tokenIndex273, depth273
			return false
		},
		/* 21 literalList <- <(Action38 _ PAREN_OPEN literalListString (_ COMMA literalListString)* _ PAREN_CLOSE)> */
		nil,
		/* 22 literalListString <- <(_ STRING Action39)> */
		func() bool {
			position277, tokenIndex277, depth277 := position, tokenIndex, depth
			{
				position278 := position
				depth++
				if !_rules[rule_]() {
					goto l277
				}
				if !_rules[ruleSTRING]() {
					goto l277
				}
				{
					add(ruleAction39, position)
				}
				depth--
				add(ruleliteralListString, position278)
			}
			return true
		l277:
			position, tokenIndex, depth = position277, tokenIndex277, depth277
			return false
		},
		/* 23 tagName <- <(_ <TAG_NAME> Action40)> */
		func() bool {
			position280, tokenIndex280, depth280 := position, tokenIndex, depth
			{
				position281 := position
				depth++
				if !_rules[rule_]() {
					goto l280
				}
				{
					position282 := position
					depth++
					{
						position283 := position
						depth++
						if !_rules[rule_]() {
							goto l280
						}
						if !_rules[ruleIDENTIFIER]() {
							goto l280
						}
						depth--
						add(ruleTAG_NAME, position283)
					}
					depth--
					add(rulePegText, position282)
				}
				{
					add(ruleAction40, position)
				}
				depth--
				add(ruletagName, position281)
			}
			return true
		l280:
			position, tokenIndex, depth = position280, tokenIndex280, depth280
			return false
		},
		/* 24 COLUMN_NAME <- <(_ IDENTIFIER)> */
		func() bool {
			position285, tokenIndex285, depth285 := position, tokenIndex, depth
			{
				position286 := position
				depth++
				if !_rules[rule_]() {
					goto l285
				}
				if !_rules[ruleIDENTIFIER]() {
					goto l285
				}
				depth--
				add(ruleCOLUMN_NAME, position286)
			}
			return true
		l285:
			position, tokenIndex, depth = position285, tokenIndex285, depth285
			return false
		},
		/* 25 METRIC_NAME <- <(_ IDENTIFIER)> */
		nil,
		/* 26 TAG_NAME <- <(_ IDENTIFIER)> */
		nil,
		/* 27 IDENTIFIER <- <((_ '`' CHAR* '`') / (_ !(KEYWORD KEY) ID_SEGMENT ('.' ID_SEGMENT)*))> */
		func() bool {
			position289, tokenIndex289, depth289 := position, tokenIndex, depth
			{
				position290 := position
				depth++
				{
					position291, tokenIndex291, depth291 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l292
					}
					if buffer[position] != rune('`') {
						goto l292
					}
					position++
				l293:
					{
						position294, tokenIndex294, depth294 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l294
						}
						goto l293
					l294:
						position, tokenIndex, depth = position294, tokenIndex294, depth294
					}
					if buffer[position] != rune('`') {
						goto l292
					}
					position++
					goto l291
				l292:
					position, tokenIndex, depth = position291, tokenIndex291, depth291
					if !_rules[rule_]() {
						goto l289
					}
					{
						position295, tokenIndex295, depth295 := position, tokenIndex, depth
						{
							position296 := position
							depth++
							{
								position297, tokenIndex297, depth297 := position, tokenIndex, depth
								{
									position299, tokenIndex299, depth299 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l300
									}
									position++
									goto l299
								l300:
									position, tokenIndex, depth = position299, tokenIndex299, depth299
									if buffer[position] != rune('A') {
										goto l298
									}
									position++
								}
							l299:
								{
									position301, tokenIndex301, depth301 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l302
									}
									position++
									goto l301
								l302:
									position, tokenIndex, depth = position301, tokenIndex301, depth301
									if buffer[position] != rune('L') {
										goto l298
									}
									position++
								}
							l301:
								{
									position303, tokenIndex303, depth303 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l304
									}
									position++
									goto l303
								l304:
									position, tokenIndex, depth = position303, tokenIndex303, depth303
									if buffer[position] != rune('L') {
										goto l298
									}
									position++
								}
							l303:
								goto l297
							l298:
								position, tokenIndex, depth = position297, tokenIndex297, depth297
								{
									position306, tokenIndex306, depth306 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l307
									}
									position++
									goto l306
								l307:
									position, tokenIndex, depth = position306, tokenIndex306, depth306
									if buffer[position] != rune('A') {
										goto l305
									}
									position++
								}
							l306:
								{
									position308, tokenIndex308, depth308 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l309
									}
									position++
									goto l308
								l309:
									position, tokenIndex, depth = position308, tokenIndex308, depth308
									if buffer[position] != rune('N') {
										goto l305
									}
									position++
								}
							l308:
								{
									position310, tokenIndex310, depth310 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l311
									}
									position++
									goto l310
								l311:
									position, tokenIndex, depth = position310, tokenIndex310, depth310
									if buffer[position] != rune('D') {
										goto l305
									}
									position++
								}
							l310:
								goto l297
							l305:
								position, tokenIndex, depth = position297, tokenIndex297, depth297
								{
									position313, tokenIndex313, depth313 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l314
									}
									position++
									goto l313
								l314:
									position, tokenIndex, depth = position313, tokenIndex313, depth313
									if buffer[position] != rune('M') {
										goto l312
									}
									position++
								}
							l313:
								{
									position315, tokenIndex315, depth315 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l316
									}
									position++
									goto l315
								l316:
									position, tokenIndex, depth = position315, tokenIndex315, depth315
									if buffer[position] != rune('A') {
										goto l312
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
										goto l312
									}
									position++
								}
							l317:
								{
									position319, tokenIndex319, depth319 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l320
									}
									position++
									goto l319
								l320:
									position, tokenIndex, depth = position319, tokenIndex319, depth319
									if buffer[position] != rune('C') {
										goto l312
									}
									position++
								}
							l319:
								{
									position321, tokenIndex321, depth321 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l322
									}
									position++
									goto l321
								l322:
									position, tokenIndex, depth = position321, tokenIndex321, depth321
									if buffer[position] != rune('H') {
										goto l312
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
										goto l312
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
										goto l312
									}
									position++
								}
							l325:
								goto l297
							l312:
								position, tokenIndex, depth = position297, tokenIndex297, depth297
								{
									position328, tokenIndex328, depth328 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l329
									}
									position++
									goto l328
								l329:
									position, tokenIndex, depth = position328, tokenIndex328, depth328
									if buffer[position] != rune('S') {
										goto l327
									}
									position++
								}
							l328:
								{
									position330, tokenIndex330, depth330 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l331
									}
									position++
									goto l330
								l331:
									position, tokenIndex, depth = position330, tokenIndex330, depth330
									if buffer[position] != rune('E') {
										goto l327
									}
									position++
								}
							l330:
								{
									position332, tokenIndex332, depth332 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l333
									}
									position++
									goto l332
								l333:
									position, tokenIndex, depth = position332, tokenIndex332, depth332
									if buffer[position] != rune('L') {
										goto l327
									}
									position++
								}
							l332:
								{
									position334, tokenIndex334, depth334 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l335
									}
									position++
									goto l334
								l335:
									position, tokenIndex, depth = position334, tokenIndex334, depth334
									if buffer[position] != rune('E') {
										goto l327
									}
									position++
								}
							l334:
								{
									position336, tokenIndex336, depth336 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l337
									}
									position++
									goto l336
								l337:
									position, tokenIndex, depth = position336, tokenIndex336, depth336
									if buffer[position] != rune('C') {
										goto l327
									}
									position++
								}
							l336:
								{
									position338, tokenIndex338, depth338 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l339
									}
									position++
									goto l338
								l339:
									position, tokenIndex, depth = position338, tokenIndex338, depth338
									if buffer[position] != rune('T') {
										goto l327
									}
									position++
								}
							l338:
								goto l297
							l327:
								position, tokenIndex, depth = position297, tokenIndex297, depth297
								{
									switch buffer[position] {
									case 'M', 'm':
										{
											position341, tokenIndex341, depth341 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l342
											}
											position++
											goto l341
										l342:
											position, tokenIndex, depth = position341, tokenIndex341, depth341
											if buffer[position] != rune('M') {
												goto l295
											}
											position++
										}
									l341:
										{
											position343, tokenIndex343, depth343 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l344
											}
											position++
											goto l343
										l344:
											position, tokenIndex, depth = position343, tokenIndex343, depth343
											if buffer[position] != rune('E') {
												goto l295
											}
											position++
										}
									l343:
										{
											position345, tokenIndex345, depth345 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l346
											}
											position++
											goto l345
										l346:
											position, tokenIndex, depth = position345, tokenIndex345, depth345
											if buffer[position] != rune('T') {
												goto l295
											}
											position++
										}
									l345:
										{
											position347, tokenIndex347, depth347 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l348
											}
											position++
											goto l347
										l348:
											position, tokenIndex, depth = position347, tokenIndex347, depth347
											if buffer[position] != rune('R') {
												goto l295
											}
											position++
										}
									l347:
										{
											position349, tokenIndex349, depth349 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l350
											}
											position++
											goto l349
										l350:
											position, tokenIndex, depth = position349, tokenIndex349, depth349
											if buffer[position] != rune('I') {
												goto l295
											}
											position++
										}
									l349:
										{
											position351, tokenIndex351, depth351 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l352
											}
											position++
											goto l351
										l352:
											position, tokenIndex, depth = position351, tokenIndex351, depth351
											if buffer[position] != rune('C') {
												goto l295
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
												goto l295
											}
											position++
										}
									l353:
										break
									case 'W', 'w':
										{
											position355, tokenIndex355, depth355 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l356
											}
											position++
											goto l355
										l356:
											position, tokenIndex, depth = position355, tokenIndex355, depth355
											if buffer[position] != rune('W') {
												goto l295
											}
											position++
										}
									l355:
										{
											position357, tokenIndex357, depth357 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l358
											}
											position++
											goto l357
										l358:
											position, tokenIndex, depth = position357, tokenIndex357, depth357
											if buffer[position] != rune('H') {
												goto l295
											}
											position++
										}
									l357:
										{
											position359, tokenIndex359, depth359 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l360
											}
											position++
											goto l359
										l360:
											position, tokenIndex, depth = position359, tokenIndex359, depth359
											if buffer[position] != rune('E') {
												goto l295
											}
											position++
										}
									l359:
										{
											position361, tokenIndex361, depth361 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l362
											}
											position++
											goto l361
										l362:
											position, tokenIndex, depth = position361, tokenIndex361, depth361
											if buffer[position] != rune('R') {
												goto l295
											}
											position++
										}
									l361:
										{
											position363, tokenIndex363, depth363 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l364
											}
											position++
											goto l363
										l364:
											position, tokenIndex, depth = position363, tokenIndex363, depth363
											if buffer[position] != rune('E') {
												goto l295
											}
											position++
										}
									l363:
										break
									case 'O', 'o':
										{
											position365, tokenIndex365, depth365 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l366
											}
											position++
											goto l365
										l366:
											position, tokenIndex, depth = position365, tokenIndex365, depth365
											if buffer[position] != rune('O') {
												goto l295
											}
											position++
										}
									l365:
										{
											position367, tokenIndex367, depth367 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l368
											}
											position++
											goto l367
										l368:
											position, tokenIndex, depth = position367, tokenIndex367, depth367
											if buffer[position] != rune('R') {
												goto l295
											}
											position++
										}
									l367:
										break
									case 'N', 'n':
										{
											position369, tokenIndex369, depth369 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l370
											}
											position++
											goto l369
										l370:
											position, tokenIndex, depth = position369, tokenIndex369, depth369
											if buffer[position] != rune('N') {
												goto l295
											}
											position++
										}
									l369:
										{
											position371, tokenIndex371, depth371 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l372
											}
											position++
											goto l371
										l372:
											position, tokenIndex, depth = position371, tokenIndex371, depth371
											if buffer[position] != rune('O') {
												goto l295
											}
											position++
										}
									l371:
										{
											position373, tokenIndex373, depth373 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l374
											}
											position++
											goto l373
										l374:
											position, tokenIndex, depth = position373, tokenIndex373, depth373
											if buffer[position] != rune('T') {
												goto l295
											}
											position++
										}
									l373:
										break
									case 'I', 'i':
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
												goto l295
											}
											position++
										}
									l375:
										{
											position377, tokenIndex377, depth377 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l378
											}
											position++
											goto l377
										l378:
											position, tokenIndex, depth = position377, tokenIndex377, depth377
											if buffer[position] != rune('N') {
												goto l295
											}
											position++
										}
									l377:
										break
									case 'G', 'g':
										{
											position379, tokenIndex379, depth379 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l380
											}
											position++
											goto l379
										l380:
											position, tokenIndex, depth = position379, tokenIndex379, depth379
											if buffer[position] != rune('G') {
												goto l295
											}
											position++
										}
									l379:
										{
											position381, tokenIndex381, depth381 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l382
											}
											position++
											goto l381
										l382:
											position, tokenIndex, depth = position381, tokenIndex381, depth381
											if buffer[position] != rune('R') {
												goto l295
											}
											position++
										}
									l381:
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
												goto l295
											}
											position++
										}
									l383:
										{
											position385, tokenIndex385, depth385 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l386
											}
											position++
											goto l385
										l386:
											position, tokenIndex, depth = position385, tokenIndex385, depth385
											if buffer[position] != rune('U') {
												goto l295
											}
											position++
										}
									l385:
										{
											position387, tokenIndex387, depth387 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l388
											}
											position++
											goto l387
										l388:
											position, tokenIndex, depth = position387, tokenIndex387, depth387
											if buffer[position] != rune('P') {
												goto l295
											}
											position++
										}
									l387:
										break
									case 'D', 'd':
										{
											position389, tokenIndex389, depth389 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l390
											}
											position++
											goto l389
										l390:
											position, tokenIndex, depth = position389, tokenIndex389, depth389
											if buffer[position] != rune('D') {
												goto l295
											}
											position++
										}
									l389:
										{
											position391, tokenIndex391, depth391 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l392
											}
											position++
											goto l391
										l392:
											position, tokenIndex, depth = position391, tokenIndex391, depth391
											if buffer[position] != rune('E') {
												goto l295
											}
											position++
										}
									l391:
										{
											position393, tokenIndex393, depth393 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l394
											}
											position++
											goto l393
										l394:
											position, tokenIndex, depth = position393, tokenIndex393, depth393
											if buffer[position] != rune('S') {
												goto l295
											}
											position++
										}
									l393:
										{
											position395, tokenIndex395, depth395 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l396
											}
											position++
											goto l395
										l396:
											position, tokenIndex, depth = position395, tokenIndex395, depth395
											if buffer[position] != rune('C') {
												goto l295
											}
											position++
										}
									l395:
										{
											position397, tokenIndex397, depth397 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l398
											}
											position++
											goto l397
										l398:
											position, tokenIndex, depth = position397, tokenIndex397, depth397
											if buffer[position] != rune('R') {
												goto l295
											}
											position++
										}
									l397:
										{
											position399, tokenIndex399, depth399 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l400
											}
											position++
											goto l399
										l400:
											position, tokenIndex, depth = position399, tokenIndex399, depth399
											if buffer[position] != rune('I') {
												goto l295
											}
											position++
										}
									l399:
										{
											position401, tokenIndex401, depth401 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l402
											}
											position++
											goto l401
										l402:
											position, tokenIndex, depth = position401, tokenIndex401, depth401
											if buffer[position] != rune('B') {
												goto l295
											}
											position++
										}
									l401:
										{
											position403, tokenIndex403, depth403 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l404
											}
											position++
											goto l403
										l404:
											position, tokenIndex, depth = position403, tokenIndex403, depth403
											if buffer[position] != rune('E') {
												goto l295
											}
											position++
										}
									l403:
										break
									case 'B', 'b':
										{
											position405, tokenIndex405, depth405 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l406
											}
											position++
											goto l405
										l406:
											position, tokenIndex, depth = position405, tokenIndex405, depth405
											if buffer[position] != rune('B') {
												goto l295
											}
											position++
										}
									l405:
										{
											position407, tokenIndex407, depth407 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l408
											}
											position++
											goto l407
										l408:
											position, tokenIndex, depth = position407, tokenIndex407, depth407
											if buffer[position] != rune('Y') {
												goto l295
											}
											position++
										}
									l407:
										break
									case 'A', 'a':
										{
											position409, tokenIndex409, depth409 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l410
											}
											position++
											goto l409
										l410:
											position, tokenIndex, depth = position409, tokenIndex409, depth409
											if buffer[position] != rune('A') {
												goto l295
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
												goto l295
											}
											position++
										}
									l411:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l295
										}
										break
									}
								}

							}
						l297:
							depth--
							add(ruleKEYWORD, position296)
						}
						if !_rules[ruleKEY]() {
							goto l295
						}
						goto l289
					l295:
						position, tokenIndex, depth = position295, tokenIndex295, depth295
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l289
					}
				l413:
					{
						position414, tokenIndex414, depth414 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l414
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l414
						}
						goto l413
					l414:
						position, tokenIndex, depth = position414, tokenIndex414, depth414
					}
				}
			l291:
				depth--
				add(ruleIDENTIFIER, position290)
			}
			return true
		l289:
			position, tokenIndex, depth = position289, tokenIndex289, depth289
			return false
		},
		/* 28 TIMESTAMP <- <((_ <(NUMBER ([a-z] / [A-Z])*)>) / (_ STRING) / (_ <(('n' / 'N') ('o' / 'O') ('w' / 'W'))>))> */
		nil,
		/* 29 ID_SEGMENT <- <(_ ID_START ID_CONT*)> */
		func() bool {
			position416, tokenIndex416, depth416 := position, tokenIndex, depth
			{
				position417 := position
				depth++
				if !_rules[rule_]() {
					goto l416
				}
				if !_rules[ruleID_START]() {
					goto l416
				}
			l418:
				{
					position419, tokenIndex419, depth419 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l419
					}
					goto l418
				l419:
					position, tokenIndex, depth = position419, tokenIndex419, depth419
				}
				depth--
				add(ruleID_SEGMENT, position417)
			}
			return true
		l416:
			position, tokenIndex, depth = position416, tokenIndex416, depth416
			return false
		},
		/* 30 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position420, tokenIndex420, depth420 := position, tokenIndex, depth
			{
				position421 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l420
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l420
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l420
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position421)
			}
			return true
		l420:
			position, tokenIndex, depth = position420, tokenIndex420, depth420
			return false
		},
		/* 31 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position423, tokenIndex423, depth423 := position, tokenIndex, depth
			{
				position424 := position
				depth++
				{
					position425, tokenIndex425, depth425 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l426
					}
					goto l425
				l426:
					position, tokenIndex, depth = position425, tokenIndex425, depth425
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l423
					}
					position++
				}
			l425:
				depth--
				add(ruleID_CONT, position424)
			}
			return true
		l423:
			position, tokenIndex, depth = position423, tokenIndex423, depth423
			return false
		},
		/* 32 PROPERTY_KEY <- <(((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> KEY _ (('b' / 'B') ('y' / 'Y')))) | (&('R' | 'r') <(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))>) | (&('T' | 't') <(('t' / 'T') ('o' / 'O'))>) | (&('F' | 'f') <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))>)) KEY)> */
		func() bool {
			position427, tokenIndex427, depth427 := position, tokenIndex, depth
			{
				position428 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position430 := position
							depth++
							{
								position431, tokenIndex431, depth431 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l432
								}
								position++
								goto l431
							l432:
								position, tokenIndex, depth = position431, tokenIndex431, depth431
								if buffer[position] != rune('S') {
									goto l427
								}
								position++
							}
						l431:
							{
								position433, tokenIndex433, depth433 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l434
								}
								position++
								goto l433
							l434:
								position, tokenIndex, depth = position433, tokenIndex433, depth433
								if buffer[position] != rune('A') {
									goto l427
								}
								position++
							}
						l433:
							{
								position435, tokenIndex435, depth435 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l436
								}
								position++
								goto l435
							l436:
								position, tokenIndex, depth = position435, tokenIndex435, depth435
								if buffer[position] != rune('M') {
									goto l427
								}
								position++
							}
						l435:
							{
								position437, tokenIndex437, depth437 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l438
								}
								position++
								goto l437
							l438:
								position, tokenIndex, depth = position437, tokenIndex437, depth437
								if buffer[position] != rune('P') {
									goto l427
								}
								position++
							}
						l437:
							{
								position439, tokenIndex439, depth439 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l440
								}
								position++
								goto l439
							l440:
								position, tokenIndex, depth = position439, tokenIndex439, depth439
								if buffer[position] != rune('L') {
									goto l427
								}
								position++
							}
						l439:
							{
								position441, tokenIndex441, depth441 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l442
								}
								position++
								goto l441
							l442:
								position, tokenIndex, depth = position441, tokenIndex441, depth441
								if buffer[position] != rune('E') {
									goto l427
								}
								position++
							}
						l441:
							depth--
							add(rulePegText, position430)
						}
						if !_rules[ruleKEY]() {
							goto l427
						}
						if !_rules[rule_]() {
							goto l427
						}
						{
							position443, tokenIndex443, depth443 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l444
							}
							position++
							goto l443
						l444:
							position, tokenIndex, depth = position443, tokenIndex443, depth443
							if buffer[position] != rune('B') {
								goto l427
							}
							position++
						}
					l443:
						{
							position445, tokenIndex445, depth445 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l446
							}
							position++
							goto l445
						l446:
							position, tokenIndex, depth = position445, tokenIndex445, depth445
							if buffer[position] != rune('Y') {
								goto l427
							}
							position++
						}
					l445:
						break
					case 'R', 'r':
						{
							position447 := position
							depth++
							{
								position448, tokenIndex448, depth448 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l449
								}
								position++
								goto l448
							l449:
								position, tokenIndex, depth = position448, tokenIndex448, depth448
								if buffer[position] != rune('R') {
									goto l427
								}
								position++
							}
						l448:
							{
								position450, tokenIndex450, depth450 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l451
								}
								position++
								goto l450
							l451:
								position, tokenIndex, depth = position450, tokenIndex450, depth450
								if buffer[position] != rune('E') {
									goto l427
								}
								position++
							}
						l450:
							{
								position452, tokenIndex452, depth452 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l453
								}
								position++
								goto l452
							l453:
								position, tokenIndex, depth = position452, tokenIndex452, depth452
								if buffer[position] != rune('S') {
									goto l427
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
									goto l427
								}
								position++
							}
						l454:
							{
								position456, tokenIndex456, depth456 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l457
								}
								position++
								goto l456
							l457:
								position, tokenIndex, depth = position456, tokenIndex456, depth456
								if buffer[position] != rune('L') {
									goto l427
								}
								position++
							}
						l456:
							{
								position458, tokenIndex458, depth458 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l459
								}
								position++
								goto l458
							l459:
								position, tokenIndex, depth = position458, tokenIndex458, depth458
								if buffer[position] != rune('U') {
									goto l427
								}
								position++
							}
						l458:
							{
								position460, tokenIndex460, depth460 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l461
								}
								position++
								goto l460
							l461:
								position, tokenIndex, depth = position460, tokenIndex460, depth460
								if buffer[position] != rune('T') {
									goto l427
								}
								position++
							}
						l460:
							{
								position462, tokenIndex462, depth462 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l463
								}
								position++
								goto l462
							l463:
								position, tokenIndex, depth = position462, tokenIndex462, depth462
								if buffer[position] != rune('I') {
									goto l427
								}
								position++
							}
						l462:
							{
								position464, tokenIndex464, depth464 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l465
								}
								position++
								goto l464
							l465:
								position, tokenIndex, depth = position464, tokenIndex464, depth464
								if buffer[position] != rune('O') {
									goto l427
								}
								position++
							}
						l464:
							{
								position466, tokenIndex466, depth466 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l467
								}
								position++
								goto l466
							l467:
								position, tokenIndex, depth = position466, tokenIndex466, depth466
								if buffer[position] != rune('N') {
									goto l427
								}
								position++
							}
						l466:
							depth--
							add(rulePegText, position447)
						}
						break
					case 'T', 't':
						{
							position468 := position
							depth++
							{
								position469, tokenIndex469, depth469 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l470
								}
								position++
								goto l469
							l470:
								position, tokenIndex, depth = position469, tokenIndex469, depth469
								if buffer[position] != rune('T') {
									goto l427
								}
								position++
							}
						l469:
							{
								position471, tokenIndex471, depth471 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l472
								}
								position++
								goto l471
							l472:
								position, tokenIndex, depth = position471, tokenIndex471, depth471
								if buffer[position] != rune('O') {
									goto l427
								}
								position++
							}
						l471:
							depth--
							add(rulePegText, position468)
						}
						break
					default:
						{
							position473 := position
							depth++
							{
								position474, tokenIndex474, depth474 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l475
								}
								position++
								goto l474
							l475:
								position, tokenIndex, depth = position474, tokenIndex474, depth474
								if buffer[position] != rune('F') {
									goto l427
								}
								position++
							}
						l474:
							{
								position476, tokenIndex476, depth476 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l477
								}
								position++
								goto l476
							l477:
								position, tokenIndex, depth = position476, tokenIndex476, depth476
								if buffer[position] != rune('R') {
									goto l427
								}
								position++
							}
						l476:
							{
								position478, tokenIndex478, depth478 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l479
								}
								position++
								goto l478
							l479:
								position, tokenIndex, depth = position478, tokenIndex478, depth478
								if buffer[position] != rune('O') {
									goto l427
								}
								position++
							}
						l478:
							{
								position480, tokenIndex480, depth480 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l481
								}
								position++
								goto l480
							l481:
								position, tokenIndex, depth = position480, tokenIndex480, depth480
								if buffer[position] != rune('M') {
									goto l427
								}
								position++
							}
						l480:
							depth--
							add(rulePegText, position473)
						}
						break
					}
				}

				if !_rules[ruleKEY]() {
					goto l427
				}
				depth--
				add(rulePROPERTY_KEY, position428)
			}
			return true
		l427:
			position, tokenIndex, depth = position427, tokenIndex427, depth427
			return false
		},
		/* 33 PROPERTY_VALUE <- <TIMESTAMP> */
		nil,
		/* 34 KEYWORD <- <((('a' / 'A') ('l' / 'L') ('l' / 'L')) / (('a' / 'A') ('n' / 'N') ('d' / 'D')) / (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) / (('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T')) / ((&('M' | 'm') (('m' / 'M') ('e' / 'E') ('t' / 'T') ('r' / 'R') ('i' / 'I') ('c' / 'C') ('s' / 'S'))) | (&('W' | 'w') (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E'))) | (&('O' | 'o') (('o' / 'O') ('r' / 'R'))) | (&('N' | 'n') (('n' / 'N') ('o' / 'O') ('t' / 'T'))) | (&('I' | 'i') (('i' / 'I') ('n' / 'N'))) | (&('G' | 'g') (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P'))) | (&('D' | 'd') (('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E'))) | (&('B' | 'b') (('b' / 'B') ('y' / 'Y'))) | (&('A' | 'a') (('a' / 'A') ('s' / 'S'))) | (&('F' | 'R' | 'S' | 'T' | 'f' | 'r' | 's' | 't') PROPERTY_KEY)))> */
		nil,
		/* 35 OP_ADD <- <'+'> */
		nil,
		/* 36 OP_SUB <- <'-'> */
		nil,
		/* 37 OP_MULT <- <'*'> */
		nil,
		/* 38 OP_DIV <- <'/'> */
		nil,
		/* 39 OP_AND <- <(('a' / 'A') ('n' / 'N') ('d' / 'D') KEY)> */
		nil,
		/* 40 OP_OR <- <(('o' / 'O') ('r' / 'R') KEY)> */
		nil,
		/* 41 OP_NOT <- <(('n' / 'N') ('o' / 'O') ('t' / 'T') KEY)> */
		nil,
		/* 42 QUOTE_SINGLE <- <'\''> */
		func() bool {
			position491, tokenIndex491, depth491 := position, tokenIndex, depth
			{
				position492 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l491
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position492)
			}
			return true
		l491:
			position, tokenIndex, depth = position491, tokenIndex491, depth491
			return false
		},
		/* 43 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position493, tokenIndex493, depth493 := position, tokenIndex, depth
			{
				position494 := position
				depth++
				if buffer[position] != rune('"') {
					goto l493
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position494)
			}
			return true
		l493:
			position, tokenIndex, depth = position493, tokenIndex493, depth493
			return false
		},
		/* 44 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> QUOTE_SINGLE) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> QUOTE_DOUBLE))> */
		func() bool {
			position495, tokenIndex495, depth495 := position, tokenIndex, depth
			{
				position496 := position
				depth++
				{
					position497, tokenIndex497, depth497 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l498
					}
					{
						position499 := position
						depth++
					l500:
						{
							position501, tokenIndex501, depth501 := position, tokenIndex, depth
							{
								position502, tokenIndex502, depth502 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l502
								}
								goto l501
							l502:
								position, tokenIndex, depth = position502, tokenIndex502, depth502
							}
							if !_rules[ruleCHAR]() {
								goto l501
							}
							goto l500
						l501:
							position, tokenIndex, depth = position501, tokenIndex501, depth501
						}
						depth--
						add(rulePegText, position499)
					}
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l498
					}
					goto l497
				l498:
					position, tokenIndex, depth = position497, tokenIndex497, depth497
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l495
					}
					{
						position503 := position
						depth++
					l504:
						{
							position505, tokenIndex505, depth505 := position, tokenIndex, depth
							{
								position506, tokenIndex506, depth506 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l506
								}
								goto l505
							l506:
								position, tokenIndex, depth = position506, tokenIndex506, depth506
							}
							if !_rules[ruleCHAR]() {
								goto l505
							}
							goto l504
						l505:
							position, tokenIndex, depth = position505, tokenIndex505, depth505
						}
						depth--
						add(rulePegText, position503)
					}
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l495
					}
				}
			l497:
				depth--
				add(ruleSTRING, position496)
			}
			return true
		l495:
			position, tokenIndex, depth = position495, tokenIndex495, depth495
			return false
		},
		/* 45 CHAR <- <(('\\' ((&('"') QUOTE_DOUBLE) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position507, tokenIndex507, depth507 := position, tokenIndex, depth
			{
				position508 := position
				depth++
				{
					position509, tokenIndex509, depth509 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l510
					}
					position++
					{
						switch buffer[position] {
						case '"':
							if !_rules[ruleQUOTE_DOUBLE]() {
								goto l510
							}
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l510
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l510
							}
							break
						}
					}

					goto l509
				l510:
					position, tokenIndex, depth = position509, tokenIndex509, depth509
					{
						position512, tokenIndex512, depth512 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l512
						}
						goto l507
					l512:
						position, tokenIndex, depth = position512, tokenIndex512, depth512
					}
					if !matchDot() {
						goto l507
					}
				}
			l509:
				depth--
				add(ruleCHAR, position508)
			}
			return true
		l507:
			position, tokenIndex, depth = position507, tokenIndex507, depth507
			return false
		},
		/* 46 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position513, tokenIndex513, depth513 := position, tokenIndex, depth
			{
				position514 := position
				depth++
				{
					position515, tokenIndex515, depth515 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l516
					}
					position++
					goto l515
				l516:
					position, tokenIndex, depth = position515, tokenIndex515, depth515
					if buffer[position] != rune('\\') {
						goto l513
					}
					position++
				}
			l515:
				depth--
				add(ruleESCAPE_CLASS, position514)
			}
			return true
		l513:
			position, tokenIndex, depth = position513, tokenIndex513, depth513
			return false
		},
		/* 47 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position517, tokenIndex517, depth517 := position, tokenIndex, depth
			{
				position518 := position
				depth++
				{
					position519 := position
					depth++
					{
						position520, tokenIndex520, depth520 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l520
						}
						position++
						goto l521
					l520:
						position, tokenIndex, depth = position520, tokenIndex520, depth520
					}
				l521:
					{
						position522 := position
						depth++
						{
							position523, tokenIndex523, depth523 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l524
							}
							position++
							goto l523
						l524:
							position, tokenIndex, depth = position523, tokenIndex523, depth523
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l517
							}
							position++
						l525:
							{
								position526, tokenIndex526, depth526 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l526
								}
								position++
								goto l525
							l526:
								position, tokenIndex, depth = position526, tokenIndex526, depth526
							}
						}
					l523:
						depth--
						add(ruleNUMBER_NATURAL, position522)
					}
					depth--
					add(ruleNUMBER_INTEGER, position519)
				}
				{
					position527, tokenIndex527, depth527 := position, tokenIndex, depth
					{
						position529 := position
						depth++
						if buffer[position] != rune('.') {
							goto l527
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l527
						}
						position++
					l530:
						{
							position531, tokenIndex531, depth531 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l531
							}
							position++
							goto l530
						l531:
							position, tokenIndex, depth = position531, tokenIndex531, depth531
						}
						depth--
						add(ruleNUMBER_FRACTION, position529)
					}
					goto l528
				l527:
					position, tokenIndex, depth = position527, tokenIndex527, depth527
				}
			l528:
				{
					position532, tokenIndex532, depth532 := position, tokenIndex, depth
					{
						position534 := position
						depth++
						{
							position535, tokenIndex535, depth535 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l536
							}
							position++
							goto l535
						l536:
							position, tokenIndex, depth = position535, tokenIndex535, depth535
							if buffer[position] != rune('E') {
								goto l532
							}
							position++
						}
					l535:
						{
							position537, tokenIndex537, depth537 := position, tokenIndex, depth
							{
								position539, tokenIndex539, depth539 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l540
								}
								position++
								goto l539
							l540:
								position, tokenIndex, depth = position539, tokenIndex539, depth539
								if buffer[position] != rune('-') {
									goto l537
								}
								position++
							}
						l539:
							goto l538
						l537:
							position, tokenIndex, depth = position537, tokenIndex537, depth537
						}
					l538:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l532
						}
						position++
					l541:
						{
							position542, tokenIndex542, depth542 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l542
							}
							position++
							goto l541
						l542:
							position, tokenIndex, depth = position542, tokenIndex542, depth542
						}
						depth--
						add(ruleNUMBER_EXP, position534)
					}
					goto l533
				l532:
					position, tokenIndex, depth = position532, tokenIndex532, depth532
				}
			l533:
				depth--
				add(ruleNUMBER, position518)
			}
			return true
		l517:
			position, tokenIndex, depth = position517, tokenIndex517, depth517
			return false
		},
		/* 48 NUMBER_NATURAL <- <('0' / ([1-9] [0-9]*))> */
		nil,
		/* 49 NUMBER_FRACTION <- <('.' [0-9]+)> */
		nil,
		/* 50 NUMBER_INTEGER <- <('-'? NUMBER_NATURAL)> */
		nil,
		/* 51 NUMBER_EXP <- <(('e' / 'E') ('+' / '-')? [0-9]+)> */
		nil,
		/* 52 PAREN_OPEN <- <'('> */
		func() bool {
			position547, tokenIndex547, depth547 := position, tokenIndex, depth
			{
				position548 := position
				depth++
				if buffer[position] != rune('(') {
					goto l547
				}
				position++
				depth--
				add(rulePAREN_OPEN, position548)
			}
			return true
		l547:
			position, tokenIndex, depth = position547, tokenIndex547, depth547
			return false
		},
		/* 53 PAREN_CLOSE <- <')'> */
		func() bool {
			position549, tokenIndex549, depth549 := position, tokenIndex, depth
			{
				position550 := position
				depth++
				if buffer[position] != rune(')') {
					goto l549
				}
				position++
				depth--
				add(rulePAREN_CLOSE, position550)
			}
			return true
		l549:
			position, tokenIndex, depth = position549, tokenIndex549, depth549
			return false
		},
		/* 54 COMMA <- <','> */
		func() bool {
			position551, tokenIndex551, depth551 := position, tokenIndex, depth
			{
				position552 := position
				depth++
				if buffer[position] != rune(',') {
					goto l551
				}
				position++
				depth--
				add(ruleCOMMA, position552)
			}
			return true
		l551:
			position, tokenIndex, depth = position551, tokenIndex551, depth551
			return false
		},
		/* 55 _ <- <SPACE*> */
		func() bool {
			{
				position554 := position
				depth++
			l555:
				{
					position556, tokenIndex556, depth556 := position, tokenIndex, depth
					{
						position557 := position
						depth++
						{
							switch buffer[position] {
							case '\t':
								if buffer[position] != rune('\t') {
									goto l556
								}
								position++
								break
							case '\n':
								if buffer[position] != rune('\n') {
									goto l556
								}
								position++
								break
							default:
								if buffer[position] != rune(' ') {
									goto l556
								}
								position++
								break
							}
						}

						depth--
						add(ruleSPACE, position557)
					}
					goto l555
				l556:
					position, tokenIndex, depth = position556, tokenIndex556, depth556
				}
				depth--
				add(rule_, position554)
			}
			return true
		},
		/* 56 KEY <- <!ID_CONT> */
		func() bool {
			position559, tokenIndex559, depth559 := position, tokenIndex, depth
			{
				position560 := position
				depth++
				{
					position561, tokenIndex561, depth561 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l561
					}
					goto l559
				l561:
					position, tokenIndex, depth = position561, tokenIndex561, depth561
				}
				depth--
				add(ruleKEY, position560)
			}
			return true
		l559:
			position, tokenIndex, depth = position559, tokenIndex559, depth559
			return false
		},
		/* 57 SPACE <- <((&('\t') '\t') | (&('\n') '\n') | (&(' ') ' '))> */
		nil,
		/* 59 Action0 <- <{
		   p.makeSelect()
		 }> */
		nil,
		/* 60 Action1 <- <{ p.makeDescribeAll() }> */
		nil,
		/* 61 Action2 <- <{ p.makeDescribeMetrics() }> */
		nil,
		nil,
		/* 63 Action3 <- <{ p.addStringLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 64 Action4 <- <{ p.makeDescribe() }> */
		nil,
		/* 65 Action5 <- <{ p.addEvaluationContext() }> */
		nil,
		/* 66 Action6 <- <{ p.addPropertyKey(buffer[begin:end])   }> */
		nil,
		/* 67 Action7 <- <{ p.addPropertyValue(buffer[begin:end]) }> */
		nil,
		/* 68 Action8 <- <{ p.insertPropertyKeyValue() }> */
		nil,
		/* 69 Action9 <- <{ p.checkPropertyClause() }> */
		nil,
		/* 70 Action10 <- <{ p.addNullPredicate() }> */
		nil,
		/* 71 Action11 <- <{ p.addExpressionList() }> */
		nil,
		/* 72 Action12 <- <{ p.appendExpression() }> */
		nil,
		/* 73 Action13 <- <{ p.appendExpression() }> */
		nil,
		/* 74 Action14 <- <{ p.addOperatorLiteral("+") }> */
		nil,
		/* 75 Action15 <- <{ p.addOperatorLiteral("-") }> */
		nil,
		/* 76 Action16 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 77 Action17 <- <{ p.addOperatorLiteral("/") }> */
		nil,
		/* 78 Action18 <- <{ p.addOperatorLiteral("*") }> */
		nil,
		/* 79 Action19 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 80 Action20 <- <{ p.addNumberNode(buffer[begin:end]) }> */
		nil,
		/* 81 Action21 <- <{ p.addStringNode(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 82 Action22 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 83 Action23 <- <{ p.addGroupBy() }> */
		nil,
		/* 84 Action24 <- <{
		   p.addFunctionInvocation()
		 }> */
		nil,
		/* 85 Action25 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 86 Action26 <- <{ p.addNullPredicate() }> */
		nil,
		/* 87 Action27 <- <{
		   p.addMetricExpression()
		 }> */
		nil,
		/* 88 Action28 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 89 Action29 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		   }> */
		nil,
		/* 90 Action30 <- <{ p.addOrPredicate() }> */
		nil,
		/* 91 Action31 <- <{ p.addAndPredicate() }> */
		nil,
		/* 92 Action32 <- <{ p.addNotPredicate() }> */
		nil,
		/* 93 Action33 <- <{
		   p.addLiteralMatcher()
		 }> */
		nil,
		/* 94 Action34 <- <{
		   p.addLiteralMatcher()
		   p.addNotPredicate()
		 }> */
		nil,
		/* 95 Action35 <- <{
		   p.addRegexMatcher()
		 }> */
		nil,
		/* 96 Action36 <- <{
		   p.addListMatcher()
		 }> */
		nil,
		/* 97 Action37 <- <{
		  p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 98 Action38 <- <{ p.addLiteralList() }> */
		nil,
		/* 99 Action39 <- <{
		  p.appendLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 100 Action40 <- <{ p.addTagLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
	}
	p.rules = _rules
}
