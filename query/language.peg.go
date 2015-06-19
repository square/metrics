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
										if !_rules[rule_]() {
											goto l0
										}
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
		/* 8 expressionList <- <(Action11 expression_1 Action12 (_ COMMA expression_1 Action13)*)> */
		func() bool {
			position127, tokenIndex127, depth127 := position, tokenIndex, depth
			{
				position128 := position
				depth++
				{
					add(ruleAction11, position)
				}
				if !_rules[ruleexpression_1]() {
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
					if !_rules[ruleexpression_1]() {
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
		/* 9 expression_1 <- <(expression_2 (((_ OP_ADD Action14) / (_ OP_SUB Action15)) expression_2 Action16)*)> */
		func() bool {
			position134, tokenIndex134, depth134 := position, tokenIndex, depth
			{
				position135 := position
				depth++
				if !_rules[ruleexpression_2]() {
					goto l134
				}
			l136:
				{
					position137, tokenIndex137, depth137 := position, tokenIndex, depth
					{
						position138, tokenIndex138, depth138 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l139
						}
						{
							position140 := position
							depth++
							if buffer[position] != rune('+') {
								goto l139
							}
							position++
							depth--
							add(ruleOP_ADD, position140)
						}
						{
							add(ruleAction14, position)
						}
						goto l138
					l139:
						position, tokenIndex, depth = position138, tokenIndex138, depth138
						if !_rules[rule_]() {
							goto l137
						}
						{
							position142 := position
							depth++
							if buffer[position] != rune('-') {
								goto l137
							}
							position++
							depth--
							add(ruleOP_SUB, position142)
						}
						{
							add(ruleAction15, position)
						}
					}
				l138:
					if !_rules[ruleexpression_2]() {
						goto l137
					}
					{
						add(ruleAction16, position)
					}
					goto l136
				l137:
					position, tokenIndex, depth = position137, tokenIndex137, depth137
				}
				depth--
				add(ruleexpression_1, position135)
			}
			return true
		l134:
			position, tokenIndex, depth = position134, tokenIndex134, depth134
			return false
		},
		/* 10 expression_2 <- <(expression_3 (((_ OP_DIV Action17) / (_ OP_MULT Action18)) expression_3 Action19)*)> */
		func() bool {
			position145, tokenIndex145, depth145 := position, tokenIndex, depth
			{
				position146 := position
				depth++
				if !_rules[ruleexpression_3]() {
					goto l145
				}
			l147:
				{
					position148, tokenIndex148, depth148 := position, tokenIndex, depth
					{
						position149, tokenIndex149, depth149 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l150
						}
						{
							position151 := position
							depth++
							if buffer[position] != rune('/') {
								goto l150
							}
							position++
							depth--
							add(ruleOP_DIV, position151)
						}
						{
							add(ruleAction17, position)
						}
						goto l149
					l150:
						position, tokenIndex, depth = position149, tokenIndex149, depth149
						if !_rules[rule_]() {
							goto l148
						}
						{
							position153 := position
							depth++
							if buffer[position] != rune('*') {
								goto l148
							}
							position++
							depth--
							add(ruleOP_MULT, position153)
						}
						{
							add(ruleAction18, position)
						}
					}
				l149:
					if !_rules[ruleexpression_3]() {
						goto l148
					}
					{
						add(ruleAction19, position)
					}
					goto l147
				l148:
					position, tokenIndex, depth = position148, tokenIndex148, depth148
				}
				depth--
				add(ruleexpression_2, position146)
			}
			return true
		l145:
			position, tokenIndex, depth = position145, tokenIndex145, depth145
			return false
		},
		/* 11 expression_3 <- <(expression_function / expression_metric / (_ PAREN_OPEN expression_1 _ PAREN_CLOSE) / (_ <NUMBER> Action20) / (_ STRING Action21))> */
		func() bool {
			position156, tokenIndex156, depth156 := position, tokenIndex, depth
			{
				position157 := position
				depth++
				{
					position158, tokenIndex158, depth158 := position, tokenIndex, depth
					{
						position160 := position
						depth++
						if !_rules[rule_]() {
							goto l159
						}
						{
							position161 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l159
							}
							depth--
							add(rulePegText, position161)
						}
						{
							add(ruleAction22, position)
						}
						if !_rules[rule_]() {
							goto l159
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l159
						}
						if !_rules[ruleexpressionList]() {
							goto l159
						}
						{
							add(ruleAction23, position)
						}
						{
							position164, tokenIndex164, depth164 := position, tokenIndex, depth
							{
								position166 := position
								depth++
								if !_rules[rule_]() {
									goto l164
								}
								{
									position167, tokenIndex167, depth167 := position, tokenIndex, depth
									if buffer[position] != rune('g') {
										goto l168
									}
									position++
									goto l167
								l168:
									position, tokenIndex, depth = position167, tokenIndex167, depth167
									if buffer[position] != rune('G') {
										goto l164
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
										goto l164
									}
									position++
								}
							l169:
								{
									position171, tokenIndex171, depth171 := position, tokenIndex, depth
									if buffer[position] != rune('o') {
										goto l172
									}
									position++
									goto l171
								l172:
									position, tokenIndex, depth = position171, tokenIndex171, depth171
									if buffer[position] != rune('O') {
										goto l164
									}
									position++
								}
							l171:
								{
									position173, tokenIndex173, depth173 := position, tokenIndex, depth
									if buffer[position] != rune('u') {
										goto l174
									}
									position++
									goto l173
								l174:
									position, tokenIndex, depth = position173, tokenIndex173, depth173
									if buffer[position] != rune('U') {
										goto l164
									}
									position++
								}
							l173:
								{
									position175, tokenIndex175, depth175 := position, tokenIndex, depth
									if buffer[position] != rune('p') {
										goto l176
									}
									position++
									goto l175
								l176:
									position, tokenIndex, depth = position175, tokenIndex175, depth175
									if buffer[position] != rune('P') {
										goto l164
									}
									position++
								}
							l175:
								if !_rules[ruleKEY]() {
									goto l164
								}
								if !_rules[rule_]() {
									goto l164
								}
								{
									position177, tokenIndex177, depth177 := position, tokenIndex, depth
									if buffer[position] != rune('b') {
										goto l178
									}
									position++
									goto l177
								l178:
									position, tokenIndex, depth = position177, tokenIndex177, depth177
									if buffer[position] != rune('B') {
										goto l164
									}
									position++
								}
							l177:
								{
									position179, tokenIndex179, depth179 := position, tokenIndex, depth
									if buffer[position] != rune('y') {
										goto l180
									}
									position++
									goto l179
								l180:
									position, tokenIndex, depth = position179, tokenIndex179, depth179
									if buffer[position] != rune('Y') {
										goto l164
									}
									position++
								}
							l179:
								if !_rules[ruleKEY]() {
									goto l164
								}
								if !_rules[rule_]() {
									goto l164
								}
								{
									position181 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l164
									}
									depth--
									add(rulePegText, position181)
								}
								{
									add(ruleAction28, position)
								}
							l183:
								{
									position184, tokenIndex184, depth184 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l184
									}
									if !_rules[ruleCOMMA]() {
										goto l184
									}
									if !_rules[rule_]() {
										goto l184
									}
									{
										position185 := position
										depth++
										if !_rules[ruleCOLUMN_NAME]() {
											goto l184
										}
										depth--
										add(rulePegText, position185)
									}
									{
										add(ruleAction29, position)
									}
									goto l183
								l184:
									position, tokenIndex, depth = position184, tokenIndex184, depth184
								}
								depth--
								add(rulegroupByClause, position166)
							}
							goto l165
						l164:
							position, tokenIndex, depth = position164, tokenIndex164, depth164
						}
					l165:
						if !_rules[rule_]() {
							goto l159
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l159
						}
						{
							add(ruleAction24, position)
						}
						depth--
						add(ruleexpression_function, position160)
					}
					goto l158
				l159:
					position, tokenIndex, depth = position158, tokenIndex158, depth158
					{
						position189 := position
						depth++
						if !_rules[rule_]() {
							goto l188
						}
						{
							position190 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l188
							}
							depth--
							add(rulePegText, position190)
						}
						{
							add(ruleAction25, position)
						}
						{
							position192, tokenIndex192, depth192 := position, tokenIndex, depth
							{
								position194, tokenIndex194, depth194 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l195
								}
								if buffer[position] != rune('[') {
									goto l195
								}
								position++
								if !_rules[rulepredicate_1]() {
									goto l195
								}
								if !_rules[rule_]() {
									goto l195
								}
								if buffer[position] != rune(']') {
									goto l195
								}
								position++
								goto l194
							l195:
								position, tokenIndex, depth = position194, tokenIndex194, depth194
								{
									add(ruleAction26, position)
								}
							}
						l194:
							goto l193

							position, tokenIndex, depth = position192, tokenIndex192, depth192
						}
					l193:
						{
							add(ruleAction27, position)
						}
						depth--
						add(ruleexpression_metric, position189)
					}
					goto l158
				l188:
					position, tokenIndex, depth = position158, tokenIndex158, depth158
					if !_rules[rule_]() {
						goto l198
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l198
					}
					if !_rules[ruleexpression_1]() {
						goto l198
					}
					if !_rules[rule_]() {
						goto l198
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l198
					}
					goto l158
				l198:
					position, tokenIndex, depth = position158, tokenIndex158, depth158
					if !_rules[rule_]() {
						goto l199
					}
					{
						position200 := position
						depth++
						if !_rules[ruleNUMBER]() {
							goto l199
						}
						depth--
						add(rulePegText, position200)
					}
					{
						add(ruleAction20, position)
					}
					goto l158
				l199:
					position, tokenIndex, depth = position158, tokenIndex158, depth158
					if !_rules[rule_]() {
						goto l156
					}
					if !_rules[ruleSTRING]() {
						goto l156
					}
					{
						add(ruleAction21, position)
					}
				}
			l158:
				depth--
				add(ruleexpression_3, position157)
			}
			return true
		l156:
			position, tokenIndex, depth = position156, tokenIndex156, depth156
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
				position208 := position
				depth++
				{
					position209, tokenIndex209, depth209 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l210
					}
					if !_rules[rule_]() {
						goto l210
					}
					{
						position211 := position
						depth++
						{
							position212, tokenIndex212, depth212 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l213
							}
							position++
							goto l212
						l213:
							position, tokenIndex, depth = position212, tokenIndex212, depth212
							if buffer[position] != rune('O') {
								goto l210
							}
							position++
						}
					l212:
						{
							position214, tokenIndex214, depth214 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l215
							}
							position++
							goto l214
						l215:
							position, tokenIndex, depth = position214, tokenIndex214, depth214
							if buffer[position] != rune('R') {
								goto l210
							}
							position++
						}
					l214:
						if !_rules[ruleKEY]() {
							goto l210
						}
						depth--
						add(ruleOP_OR, position211)
					}
					if !_rules[rulepredicate_1]() {
						goto l210
					}
					{
						add(ruleAction30, position)
					}
					goto l209
				l210:
					position, tokenIndex, depth = position209, tokenIndex209, depth209
					if !_rules[rulepredicate_2]() {
						goto l217
					}
					goto l209
				l217:
					position, tokenIndex, depth = position209, tokenIndex209, depth209
				}
			l209:
				depth--
				add(rulepredicate_1, position208)
			}
			return true
		},
		/* 17 predicate_2 <- <((predicate_3 _ OP_AND predicate_2 Action31) / predicate_3)> */
		func() bool {
			position218, tokenIndex218, depth218 := position, tokenIndex, depth
			{
				position219 := position
				depth++
				{
					position220, tokenIndex220, depth220 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l221
					}
					if !_rules[rule_]() {
						goto l221
					}
					{
						position222 := position
						depth++
						{
							position223, tokenIndex223, depth223 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l224
							}
							position++
							goto l223
						l224:
							position, tokenIndex, depth = position223, tokenIndex223, depth223
							if buffer[position] != rune('A') {
								goto l221
							}
							position++
						}
					l223:
						{
							position225, tokenIndex225, depth225 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l226
							}
							position++
							goto l225
						l226:
							position, tokenIndex, depth = position225, tokenIndex225, depth225
							if buffer[position] != rune('N') {
								goto l221
							}
							position++
						}
					l225:
						{
							position227, tokenIndex227, depth227 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l228
							}
							position++
							goto l227
						l228:
							position, tokenIndex, depth = position227, tokenIndex227, depth227
							if buffer[position] != rune('D') {
								goto l221
							}
							position++
						}
					l227:
						if !_rules[ruleKEY]() {
							goto l221
						}
						depth--
						add(ruleOP_AND, position222)
					}
					if !_rules[rulepredicate_2]() {
						goto l221
					}
					{
						add(ruleAction31, position)
					}
					goto l220
				l221:
					position, tokenIndex, depth = position220, tokenIndex220, depth220
					if !_rules[rulepredicate_3]() {
						goto l218
					}
				}
			l220:
				depth--
				add(rulepredicate_2, position219)
			}
			return true
		l218:
			position, tokenIndex, depth = position218, tokenIndex218, depth218
			return false
		},
		/* 18 predicate_3 <- <((_ OP_NOT predicate_3 Action32) / (_ PAREN_OPEN predicate_1 _ PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position230, tokenIndex230, depth230 := position, tokenIndex, depth
			{
				position231 := position
				depth++
				{
					position232, tokenIndex232, depth232 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l233
					}
					{
						position234 := position
						depth++
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
								goto l233
							}
							position++
						}
					l235:
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
								goto l233
							}
							position++
						}
					l237:
						{
							position239, tokenIndex239, depth239 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l240
							}
							position++
							goto l239
						l240:
							position, tokenIndex, depth = position239, tokenIndex239, depth239
							if buffer[position] != rune('T') {
								goto l233
							}
							position++
						}
					l239:
						if !_rules[ruleKEY]() {
							goto l233
						}
						depth--
						add(ruleOP_NOT, position234)
					}
					if !_rules[rulepredicate_3]() {
						goto l233
					}
					{
						add(ruleAction32, position)
					}
					goto l232
				l233:
					position, tokenIndex, depth = position232, tokenIndex232, depth232
					if !_rules[rule_]() {
						goto l242
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l242
					}
					if !_rules[rulepredicate_1]() {
						goto l242
					}
					if !_rules[rule_]() {
						goto l242
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l242
					}
					goto l232
				l242:
					position, tokenIndex, depth = position232, tokenIndex232, depth232
					{
						position243 := position
						depth++
						{
							position244, tokenIndex244, depth244 := position, tokenIndex, depth
							if !_rules[ruletagName]() {
								goto l245
							}
							if !_rules[rule_]() {
								goto l245
							}
							if buffer[position] != rune('=') {
								goto l245
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l245
							}
							{
								add(ruleAction33, position)
							}
							goto l244
						l245:
							position, tokenIndex, depth = position244, tokenIndex244, depth244
							if !_rules[ruletagName]() {
								goto l247
							}
							if !_rules[rule_]() {
								goto l247
							}
							if buffer[position] != rune('!') {
								goto l247
							}
							position++
							if buffer[position] != rune('=') {
								goto l247
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l247
							}
							{
								add(ruleAction34, position)
							}
							goto l244
						l247:
							position, tokenIndex, depth = position244, tokenIndex244, depth244
							if !_rules[ruletagName]() {
								goto l249
							}
							if !_rules[rule_]() {
								goto l249
							}
							{
								position250, tokenIndex250, depth250 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l251
								}
								position++
								goto l250
							l251:
								position, tokenIndex, depth = position250, tokenIndex250, depth250
								if buffer[position] != rune('M') {
									goto l249
								}
								position++
							}
						l250:
							{
								position252, tokenIndex252, depth252 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l253
								}
								position++
								goto l252
							l253:
								position, tokenIndex, depth = position252, tokenIndex252, depth252
								if buffer[position] != rune('A') {
									goto l249
								}
								position++
							}
						l252:
							{
								position254, tokenIndex254, depth254 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l255
								}
								position++
								goto l254
							l255:
								position, tokenIndex, depth = position254, tokenIndex254, depth254
								if buffer[position] != rune('T') {
									goto l249
								}
								position++
							}
						l254:
							{
								position256, tokenIndex256, depth256 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l257
								}
								position++
								goto l256
							l257:
								position, tokenIndex, depth = position256, tokenIndex256, depth256
								if buffer[position] != rune('C') {
									goto l249
								}
								position++
							}
						l256:
							{
								position258, tokenIndex258, depth258 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l259
								}
								position++
								goto l258
							l259:
								position, tokenIndex, depth = position258, tokenIndex258, depth258
								if buffer[position] != rune('H') {
									goto l249
								}
								position++
							}
						l258:
							{
								position260, tokenIndex260, depth260 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l261
								}
								position++
								goto l260
							l261:
								position, tokenIndex, depth = position260, tokenIndex260, depth260
								if buffer[position] != rune('E') {
									goto l249
								}
								position++
							}
						l260:
							{
								position262, tokenIndex262, depth262 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l263
								}
								position++
								goto l262
							l263:
								position, tokenIndex, depth = position262, tokenIndex262, depth262
								if buffer[position] != rune('S') {
									goto l249
								}
								position++
							}
						l262:
							if !_rules[ruleKEY]() {
								goto l249
							}
							if !_rules[ruleliteralString]() {
								goto l249
							}
							{
								add(ruleAction35, position)
							}
							goto l244
						l249:
							position, tokenIndex, depth = position244, tokenIndex244, depth244
							if !_rules[ruletagName]() {
								goto l230
							}
							if !_rules[rule_]() {
								goto l230
							}
							{
								position265, tokenIndex265, depth265 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l266
								}
								position++
								goto l265
							l266:
								position, tokenIndex, depth = position265, tokenIndex265, depth265
								if buffer[position] != rune('I') {
									goto l230
								}
								position++
							}
						l265:
							{
								position267, tokenIndex267, depth267 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l268
								}
								position++
								goto l267
							l268:
								position, tokenIndex, depth = position267, tokenIndex267, depth267
								if buffer[position] != rune('N') {
									goto l230
								}
								position++
							}
						l267:
							if !_rules[ruleKEY]() {
								goto l230
							}
							{
								position269 := position
								depth++
								{
									add(ruleAction38, position)
								}
								if !_rules[rule_]() {
									goto l230
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l230
								}
								if !_rules[ruleliteralListString]() {
									goto l230
								}
							l271:
								{
									position272, tokenIndex272, depth272 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l272
									}
									if !_rules[ruleCOMMA]() {
										goto l272
									}
									if !_rules[ruleliteralListString]() {
										goto l272
									}
									goto l271
								l272:
									position, tokenIndex, depth = position272, tokenIndex272, depth272
								}
								if !_rules[rule_]() {
									goto l230
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l230
								}
								depth--
								add(ruleliteralList, position269)
							}
							{
								add(ruleAction36, position)
							}
						}
					l244:
						depth--
						add(ruletagMatcher, position243)
					}
				}
			l232:
				depth--
				add(rulepredicate_3, position231)
			}
			return true
		l230:
			position, tokenIndex, depth = position230, tokenIndex230, depth230
			return false
		},
		/* 19 tagMatcher <- <((tagName _ '=' literalString Action33) / (tagName _ ('!' '=') literalString Action34) / (tagName _ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) KEY literalString Action35) / (tagName _ (('i' / 'I') ('n' / 'N')) KEY literalList Action36))> */
		nil,
		/* 20 literalString <- <(_ STRING Action37)> */
		func() bool {
			position275, tokenIndex275, depth275 := position, tokenIndex, depth
			{
				position276 := position
				depth++
				if !_rules[rule_]() {
					goto l275
				}
				if !_rules[ruleSTRING]() {
					goto l275
				}
				{
					add(ruleAction37, position)
				}
				depth--
				add(ruleliteralString, position276)
			}
			return true
		l275:
			position, tokenIndex, depth = position275, tokenIndex275, depth275
			return false
		},
		/* 21 literalList <- <(Action38 _ PAREN_OPEN literalListString (_ COMMA literalListString)* _ PAREN_CLOSE)> */
		nil,
		/* 22 literalListString <- <(_ STRING Action39)> */
		func() bool {
			position279, tokenIndex279, depth279 := position, tokenIndex, depth
			{
				position280 := position
				depth++
				if !_rules[rule_]() {
					goto l279
				}
				if !_rules[ruleSTRING]() {
					goto l279
				}
				{
					add(ruleAction39, position)
				}
				depth--
				add(ruleliteralListString, position280)
			}
			return true
		l279:
			position, tokenIndex, depth = position279, tokenIndex279, depth279
			return false
		},
		/* 23 tagName <- <(_ <TAG_NAME> Action40)> */
		func() bool {
			position282, tokenIndex282, depth282 := position, tokenIndex, depth
			{
				position283 := position
				depth++
				if !_rules[rule_]() {
					goto l282
				}
				{
					position284 := position
					depth++
					{
						position285 := position
						depth++
						if !_rules[rule_]() {
							goto l282
						}
						if !_rules[ruleIDENTIFIER]() {
							goto l282
						}
						depth--
						add(ruleTAG_NAME, position285)
					}
					depth--
					add(rulePegText, position284)
				}
				{
					add(ruleAction40, position)
				}
				depth--
				add(ruletagName, position283)
			}
			return true
		l282:
			position, tokenIndex, depth = position282, tokenIndex282, depth282
			return false
		},
		/* 24 COLUMN_NAME <- <(_ IDENTIFIER)> */
		func() bool {
			position287, tokenIndex287, depth287 := position, tokenIndex, depth
			{
				position288 := position
				depth++
				if !_rules[rule_]() {
					goto l287
				}
				if !_rules[ruleIDENTIFIER]() {
					goto l287
				}
				depth--
				add(ruleCOLUMN_NAME, position288)
			}
			return true
		l287:
			position, tokenIndex, depth = position287, tokenIndex287, depth287
			return false
		},
		/* 25 METRIC_NAME <- <(_ IDENTIFIER)> */
		nil,
		/* 26 TAG_NAME <- <(_ IDENTIFIER)> */
		nil,
		/* 27 IDENTIFIER <- <((_ '`' CHAR* '`') / (_ !(KEYWORD KEY) ID_SEGMENT ('.' ID_SEGMENT)*))> */
		func() bool {
			position291, tokenIndex291, depth291 := position, tokenIndex, depth
			{
				position292 := position
				depth++
				{
					position293, tokenIndex293, depth293 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l294
					}
					if buffer[position] != rune('`') {
						goto l294
					}
					position++
				l295:
					{
						position296, tokenIndex296, depth296 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l296
						}
						goto l295
					l296:
						position, tokenIndex, depth = position296, tokenIndex296, depth296
					}
					if buffer[position] != rune('`') {
						goto l294
					}
					position++
					goto l293
				l294:
					position, tokenIndex, depth = position293, tokenIndex293, depth293
					if !_rules[rule_]() {
						goto l291
					}
					{
						position297, tokenIndex297, depth297 := position, tokenIndex, depth
						{
							position298 := position
							depth++
							{
								position299, tokenIndex299, depth299 := position, tokenIndex, depth
								{
									position301, tokenIndex301, depth301 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l302
									}
									position++
									goto l301
								l302:
									position, tokenIndex, depth = position301, tokenIndex301, depth301
									if buffer[position] != rune('A') {
										goto l300
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
										goto l300
									}
									position++
								}
							l303:
								{
									position305, tokenIndex305, depth305 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l306
									}
									position++
									goto l305
								l306:
									position, tokenIndex, depth = position305, tokenIndex305, depth305
									if buffer[position] != rune('L') {
										goto l300
									}
									position++
								}
							l305:
								goto l299
							l300:
								position, tokenIndex, depth = position299, tokenIndex299, depth299
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
										goto l307
									}
									position++
								}
							l308:
								{
									position310, tokenIndex310, depth310 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l311
									}
									position++
									goto l310
								l311:
									position, tokenIndex, depth = position310, tokenIndex310, depth310
									if buffer[position] != rune('N') {
										goto l307
									}
									position++
								}
							l310:
								{
									position312, tokenIndex312, depth312 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l313
									}
									position++
									goto l312
								l313:
									position, tokenIndex, depth = position312, tokenIndex312, depth312
									if buffer[position] != rune('D') {
										goto l307
									}
									position++
								}
							l312:
								goto l299
							l307:
								position, tokenIndex, depth = position299, tokenIndex299, depth299
								{
									position315, tokenIndex315, depth315 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l316
									}
									position++
									goto l315
								l316:
									position, tokenIndex, depth = position315, tokenIndex315, depth315
									if buffer[position] != rune('M') {
										goto l314
									}
									position++
								}
							l315:
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
										goto l314
									}
									position++
								}
							l317:
								{
									position319, tokenIndex319, depth319 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l320
									}
									position++
									goto l319
								l320:
									position, tokenIndex, depth = position319, tokenIndex319, depth319
									if buffer[position] != rune('T') {
										goto l314
									}
									position++
								}
							l319:
								{
									position321, tokenIndex321, depth321 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l322
									}
									position++
									goto l321
								l322:
									position, tokenIndex, depth = position321, tokenIndex321, depth321
									if buffer[position] != rune('C') {
										goto l314
									}
									position++
								}
							l321:
								{
									position323, tokenIndex323, depth323 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l324
									}
									position++
									goto l323
								l324:
									position, tokenIndex, depth = position323, tokenIndex323, depth323
									if buffer[position] != rune('H') {
										goto l314
									}
									position++
								}
							l323:
								{
									position325, tokenIndex325, depth325 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l326
									}
									position++
									goto l325
								l326:
									position, tokenIndex, depth = position325, tokenIndex325, depth325
									if buffer[position] != rune('E') {
										goto l314
									}
									position++
								}
							l325:
								{
									position327, tokenIndex327, depth327 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l328
									}
									position++
									goto l327
								l328:
									position, tokenIndex, depth = position327, tokenIndex327, depth327
									if buffer[position] != rune('S') {
										goto l314
									}
									position++
								}
							l327:
								goto l299
							l314:
								position, tokenIndex, depth = position299, tokenIndex299, depth299
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
										goto l329
									}
									position++
								}
							l330:
								{
									position332, tokenIndex332, depth332 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l333
									}
									position++
									goto l332
								l333:
									position, tokenIndex, depth = position332, tokenIndex332, depth332
									if buffer[position] != rune('E') {
										goto l329
									}
									position++
								}
							l332:
								{
									position334, tokenIndex334, depth334 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l335
									}
									position++
									goto l334
								l335:
									position, tokenIndex, depth = position334, tokenIndex334, depth334
									if buffer[position] != rune('L') {
										goto l329
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
										goto l329
									}
									position++
								}
							l336:
								{
									position338, tokenIndex338, depth338 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l339
									}
									position++
									goto l338
								l339:
									position, tokenIndex, depth = position338, tokenIndex338, depth338
									if buffer[position] != rune('C') {
										goto l329
									}
									position++
								}
							l338:
								{
									position340, tokenIndex340, depth340 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l341
									}
									position++
									goto l340
								l341:
									position, tokenIndex, depth = position340, tokenIndex340, depth340
									if buffer[position] != rune('T') {
										goto l329
									}
									position++
								}
							l340:
								goto l299
							l329:
								position, tokenIndex, depth = position299, tokenIndex299, depth299
								{
									switch buffer[position] {
									case 'M', 'm':
										{
											position343, tokenIndex343, depth343 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l344
											}
											position++
											goto l343
										l344:
											position, tokenIndex, depth = position343, tokenIndex343, depth343
											if buffer[position] != rune('M') {
												goto l297
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
												goto l297
											}
											position++
										}
									l345:
										{
											position347, tokenIndex347, depth347 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l348
											}
											position++
											goto l347
										l348:
											position, tokenIndex, depth = position347, tokenIndex347, depth347
											if buffer[position] != rune('T') {
												goto l297
											}
											position++
										}
									l347:
										{
											position349, tokenIndex349, depth349 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l350
											}
											position++
											goto l349
										l350:
											position, tokenIndex, depth = position349, tokenIndex349, depth349
											if buffer[position] != rune('R') {
												goto l297
											}
											position++
										}
									l349:
										{
											position351, tokenIndex351, depth351 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l352
											}
											position++
											goto l351
										l352:
											position, tokenIndex, depth = position351, tokenIndex351, depth351
											if buffer[position] != rune('I') {
												goto l297
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
												goto l297
											}
											position++
										}
									l353:
										{
											position355, tokenIndex355, depth355 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l356
											}
											position++
											goto l355
										l356:
											position, tokenIndex, depth = position355, tokenIndex355, depth355
											if buffer[position] != rune('S') {
												goto l297
											}
											position++
										}
									l355:
										break
									case 'W', 'w':
										{
											position357, tokenIndex357, depth357 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l358
											}
											position++
											goto l357
										l358:
											position, tokenIndex, depth = position357, tokenIndex357, depth357
											if buffer[position] != rune('W') {
												goto l297
											}
											position++
										}
									l357:
										{
											position359, tokenIndex359, depth359 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l360
											}
											position++
											goto l359
										l360:
											position, tokenIndex, depth = position359, tokenIndex359, depth359
											if buffer[position] != rune('H') {
												goto l297
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
												goto l297
											}
											position++
										}
									l361:
										{
											position363, tokenIndex363, depth363 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l364
											}
											position++
											goto l363
										l364:
											position, tokenIndex, depth = position363, tokenIndex363, depth363
											if buffer[position] != rune('R') {
												goto l297
											}
											position++
										}
									l363:
										{
											position365, tokenIndex365, depth365 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l366
											}
											position++
											goto l365
										l366:
											position, tokenIndex, depth = position365, tokenIndex365, depth365
											if buffer[position] != rune('E') {
												goto l297
											}
											position++
										}
									l365:
										break
									case 'O', 'o':
										{
											position367, tokenIndex367, depth367 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l368
											}
											position++
											goto l367
										l368:
											position, tokenIndex, depth = position367, tokenIndex367, depth367
											if buffer[position] != rune('O') {
												goto l297
											}
											position++
										}
									l367:
										{
											position369, tokenIndex369, depth369 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l370
											}
											position++
											goto l369
										l370:
											position, tokenIndex, depth = position369, tokenIndex369, depth369
											if buffer[position] != rune('R') {
												goto l297
											}
											position++
										}
									l369:
										break
									case 'N', 'n':
										{
											position371, tokenIndex371, depth371 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l372
											}
											position++
											goto l371
										l372:
											position, tokenIndex, depth = position371, tokenIndex371, depth371
											if buffer[position] != rune('N') {
												goto l297
											}
											position++
										}
									l371:
										{
											position373, tokenIndex373, depth373 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l374
											}
											position++
											goto l373
										l374:
											position, tokenIndex, depth = position373, tokenIndex373, depth373
											if buffer[position] != rune('O') {
												goto l297
											}
											position++
										}
									l373:
										{
											position375, tokenIndex375, depth375 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l376
											}
											position++
											goto l375
										l376:
											position, tokenIndex, depth = position375, tokenIndex375, depth375
											if buffer[position] != rune('T') {
												goto l297
											}
											position++
										}
									l375:
										break
									case 'I', 'i':
										{
											position377, tokenIndex377, depth377 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l378
											}
											position++
											goto l377
										l378:
											position, tokenIndex, depth = position377, tokenIndex377, depth377
											if buffer[position] != rune('I') {
												goto l297
											}
											position++
										}
									l377:
										{
											position379, tokenIndex379, depth379 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l380
											}
											position++
											goto l379
										l380:
											position, tokenIndex, depth = position379, tokenIndex379, depth379
											if buffer[position] != rune('N') {
												goto l297
											}
											position++
										}
									l379:
										break
									case 'G', 'g':
										{
											position381, tokenIndex381, depth381 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l382
											}
											position++
											goto l381
										l382:
											position, tokenIndex, depth = position381, tokenIndex381, depth381
											if buffer[position] != rune('G') {
												goto l297
											}
											position++
										}
									l381:
										{
											position383, tokenIndex383, depth383 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l384
											}
											position++
											goto l383
										l384:
											position, tokenIndex, depth = position383, tokenIndex383, depth383
											if buffer[position] != rune('R') {
												goto l297
											}
											position++
										}
									l383:
										{
											position385, tokenIndex385, depth385 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l386
											}
											position++
											goto l385
										l386:
											position, tokenIndex, depth = position385, tokenIndex385, depth385
											if buffer[position] != rune('O') {
												goto l297
											}
											position++
										}
									l385:
										{
											position387, tokenIndex387, depth387 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l388
											}
											position++
											goto l387
										l388:
											position, tokenIndex, depth = position387, tokenIndex387, depth387
											if buffer[position] != rune('U') {
												goto l297
											}
											position++
										}
									l387:
										{
											position389, tokenIndex389, depth389 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l390
											}
											position++
											goto l389
										l390:
											position, tokenIndex, depth = position389, tokenIndex389, depth389
											if buffer[position] != rune('P') {
												goto l297
											}
											position++
										}
									l389:
										break
									case 'D', 'd':
										{
											position391, tokenIndex391, depth391 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l392
											}
											position++
											goto l391
										l392:
											position, tokenIndex, depth = position391, tokenIndex391, depth391
											if buffer[position] != rune('D') {
												goto l297
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
												goto l297
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
												goto l297
											}
											position++
										}
									l395:
										{
											position397, tokenIndex397, depth397 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l398
											}
											position++
											goto l397
										l398:
											position, tokenIndex, depth = position397, tokenIndex397, depth397
											if buffer[position] != rune('C') {
												goto l297
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
												goto l297
											}
											position++
										}
									l399:
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
												goto l297
											}
											position++
										}
									l401:
										{
											position403, tokenIndex403, depth403 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l404
											}
											position++
											goto l403
										l404:
											position, tokenIndex, depth = position403, tokenIndex403, depth403
											if buffer[position] != rune('B') {
												goto l297
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
												goto l297
											}
											position++
										}
									l405:
										break
									case 'B', 'b':
										{
											position407, tokenIndex407, depth407 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l408
											}
											position++
											goto l407
										l408:
											position, tokenIndex, depth = position407, tokenIndex407, depth407
											if buffer[position] != rune('B') {
												goto l297
											}
											position++
										}
									l407:
										{
											position409, tokenIndex409, depth409 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l410
											}
											position++
											goto l409
										l410:
											position, tokenIndex, depth = position409, tokenIndex409, depth409
											if buffer[position] != rune('Y') {
												goto l297
											}
											position++
										}
									l409:
										break
									case 'A', 'a':
										{
											position411, tokenIndex411, depth411 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l412
											}
											position++
											goto l411
										l412:
											position, tokenIndex, depth = position411, tokenIndex411, depth411
											if buffer[position] != rune('A') {
												goto l297
											}
											position++
										}
									l411:
										{
											position413, tokenIndex413, depth413 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l414
											}
											position++
											goto l413
										l414:
											position, tokenIndex, depth = position413, tokenIndex413, depth413
											if buffer[position] != rune('S') {
												goto l297
											}
											position++
										}
									l413:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l297
										}
										break
									}
								}

							}
						l299:
							depth--
							add(ruleKEYWORD, position298)
						}
						if !_rules[ruleKEY]() {
							goto l297
						}
						goto l291
					l297:
						position, tokenIndex, depth = position297, tokenIndex297, depth297
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l291
					}
				l415:
					{
						position416, tokenIndex416, depth416 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l416
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l416
						}
						goto l415
					l416:
						position, tokenIndex, depth = position416, tokenIndex416, depth416
					}
				}
			l293:
				depth--
				add(ruleIDENTIFIER, position292)
			}
			return true
		l291:
			position, tokenIndex, depth = position291, tokenIndex291, depth291
			return false
		},
		/* 28 TIMESTAMP <- <((_ <(NUMBER ([a-z] / [A-Z])*)>) / (_ STRING) / (_ <(('n' / 'N') ('o' / 'O') ('w' / 'W'))>))> */
		nil,
		/* 29 ID_SEGMENT <- <(_ ID_START ID_CONT*)> */
		func() bool {
			position418, tokenIndex418, depth418 := position, tokenIndex, depth
			{
				position419 := position
				depth++
				if !_rules[rule_]() {
					goto l418
				}
				if !_rules[ruleID_START]() {
					goto l418
				}
			l420:
				{
					position421, tokenIndex421, depth421 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l421
					}
					goto l420
				l421:
					position, tokenIndex, depth = position421, tokenIndex421, depth421
				}
				depth--
				add(ruleID_SEGMENT, position419)
			}
			return true
		l418:
			position, tokenIndex, depth = position418, tokenIndex418, depth418
			return false
		},
		/* 30 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position422, tokenIndex422, depth422 := position, tokenIndex, depth
			{
				position423 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l422
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l422
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l422
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position423)
			}
			return true
		l422:
			position, tokenIndex, depth = position422, tokenIndex422, depth422
			return false
		},
		/* 31 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position425, tokenIndex425, depth425 := position, tokenIndex, depth
			{
				position426 := position
				depth++
				{
					position427, tokenIndex427, depth427 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l428
					}
					goto l427
				l428:
					position, tokenIndex, depth = position427, tokenIndex427, depth427
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l425
					}
					position++
				}
			l427:
				depth--
				add(ruleID_CONT, position426)
			}
			return true
		l425:
			position, tokenIndex, depth = position425, tokenIndex425, depth425
			return false
		},
		/* 32 PROPERTY_KEY <- <(((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> KEY _ (('b' / 'B') ('y' / 'Y')))) | (&('R' | 'r') <(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))>) | (&('T' | 't') <(('t' / 'T') ('o' / 'O'))>) | (&('F' | 'f') <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))>)) KEY)> */
		func() bool {
			position429, tokenIndex429, depth429 := position, tokenIndex, depth
			{
				position430 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position432 := position
							depth++
							{
								position433, tokenIndex433, depth433 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l434
								}
								position++
								goto l433
							l434:
								position, tokenIndex, depth = position433, tokenIndex433, depth433
								if buffer[position] != rune('S') {
									goto l429
								}
								position++
							}
						l433:
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
									goto l429
								}
								position++
							}
						l435:
							{
								position437, tokenIndex437, depth437 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l438
								}
								position++
								goto l437
							l438:
								position, tokenIndex, depth = position437, tokenIndex437, depth437
								if buffer[position] != rune('M') {
									goto l429
								}
								position++
							}
						l437:
							{
								position439, tokenIndex439, depth439 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l440
								}
								position++
								goto l439
							l440:
								position, tokenIndex, depth = position439, tokenIndex439, depth439
								if buffer[position] != rune('P') {
									goto l429
								}
								position++
							}
						l439:
							{
								position441, tokenIndex441, depth441 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l442
								}
								position++
								goto l441
							l442:
								position, tokenIndex, depth = position441, tokenIndex441, depth441
								if buffer[position] != rune('L') {
									goto l429
								}
								position++
							}
						l441:
							{
								position443, tokenIndex443, depth443 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l444
								}
								position++
								goto l443
							l444:
								position, tokenIndex, depth = position443, tokenIndex443, depth443
								if buffer[position] != rune('E') {
									goto l429
								}
								position++
							}
						l443:
							depth--
							add(rulePegText, position432)
						}
						if !_rules[ruleKEY]() {
							goto l429
						}
						if !_rules[rule_]() {
							goto l429
						}
						{
							position445, tokenIndex445, depth445 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l446
							}
							position++
							goto l445
						l446:
							position, tokenIndex, depth = position445, tokenIndex445, depth445
							if buffer[position] != rune('B') {
								goto l429
							}
							position++
						}
					l445:
						{
							position447, tokenIndex447, depth447 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l448
							}
							position++
							goto l447
						l448:
							position, tokenIndex, depth = position447, tokenIndex447, depth447
							if buffer[position] != rune('Y') {
								goto l429
							}
							position++
						}
					l447:
						break
					case 'R', 'r':
						{
							position449 := position
							depth++
							{
								position450, tokenIndex450, depth450 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l451
								}
								position++
								goto l450
							l451:
								position, tokenIndex, depth = position450, tokenIndex450, depth450
								if buffer[position] != rune('R') {
									goto l429
								}
								position++
							}
						l450:
							{
								position452, tokenIndex452, depth452 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l453
								}
								position++
								goto l452
							l453:
								position, tokenIndex, depth = position452, tokenIndex452, depth452
								if buffer[position] != rune('E') {
									goto l429
								}
								position++
							}
						l452:
							{
								position454, tokenIndex454, depth454 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l455
								}
								position++
								goto l454
							l455:
								position, tokenIndex, depth = position454, tokenIndex454, depth454
								if buffer[position] != rune('S') {
									goto l429
								}
								position++
							}
						l454:
							{
								position456, tokenIndex456, depth456 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l457
								}
								position++
								goto l456
							l457:
								position, tokenIndex, depth = position456, tokenIndex456, depth456
								if buffer[position] != rune('O') {
									goto l429
								}
								position++
							}
						l456:
							{
								position458, tokenIndex458, depth458 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l459
								}
								position++
								goto l458
							l459:
								position, tokenIndex, depth = position458, tokenIndex458, depth458
								if buffer[position] != rune('L') {
									goto l429
								}
								position++
							}
						l458:
							{
								position460, tokenIndex460, depth460 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l461
								}
								position++
								goto l460
							l461:
								position, tokenIndex, depth = position460, tokenIndex460, depth460
								if buffer[position] != rune('U') {
									goto l429
								}
								position++
							}
						l460:
							{
								position462, tokenIndex462, depth462 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l463
								}
								position++
								goto l462
							l463:
								position, tokenIndex, depth = position462, tokenIndex462, depth462
								if buffer[position] != rune('T') {
									goto l429
								}
								position++
							}
						l462:
							{
								position464, tokenIndex464, depth464 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l465
								}
								position++
								goto l464
							l465:
								position, tokenIndex, depth = position464, tokenIndex464, depth464
								if buffer[position] != rune('I') {
									goto l429
								}
								position++
							}
						l464:
							{
								position466, tokenIndex466, depth466 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l467
								}
								position++
								goto l466
							l467:
								position, tokenIndex, depth = position466, tokenIndex466, depth466
								if buffer[position] != rune('O') {
									goto l429
								}
								position++
							}
						l466:
							{
								position468, tokenIndex468, depth468 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l469
								}
								position++
								goto l468
							l469:
								position, tokenIndex, depth = position468, tokenIndex468, depth468
								if buffer[position] != rune('N') {
									goto l429
								}
								position++
							}
						l468:
							depth--
							add(rulePegText, position449)
						}
						break
					case 'T', 't':
						{
							position470 := position
							depth++
							{
								position471, tokenIndex471, depth471 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l472
								}
								position++
								goto l471
							l472:
								position, tokenIndex, depth = position471, tokenIndex471, depth471
								if buffer[position] != rune('T') {
									goto l429
								}
								position++
							}
						l471:
							{
								position473, tokenIndex473, depth473 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l474
								}
								position++
								goto l473
							l474:
								position, tokenIndex, depth = position473, tokenIndex473, depth473
								if buffer[position] != rune('O') {
									goto l429
								}
								position++
							}
						l473:
							depth--
							add(rulePegText, position470)
						}
						break
					default:
						{
							position475 := position
							depth++
							{
								position476, tokenIndex476, depth476 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l477
								}
								position++
								goto l476
							l477:
								position, tokenIndex, depth = position476, tokenIndex476, depth476
								if buffer[position] != rune('F') {
									goto l429
								}
								position++
							}
						l476:
							{
								position478, tokenIndex478, depth478 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l479
								}
								position++
								goto l478
							l479:
								position, tokenIndex, depth = position478, tokenIndex478, depth478
								if buffer[position] != rune('R') {
									goto l429
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
									goto l429
								}
								position++
							}
						l480:
							{
								position482, tokenIndex482, depth482 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l483
								}
								position++
								goto l482
							l483:
								position, tokenIndex, depth = position482, tokenIndex482, depth482
								if buffer[position] != rune('M') {
									goto l429
								}
								position++
							}
						l482:
							depth--
							add(rulePegText, position475)
						}
						break
					}
				}

				if !_rules[ruleKEY]() {
					goto l429
				}
				depth--
				add(rulePROPERTY_KEY, position430)
			}
			return true
		l429:
			position, tokenIndex, depth = position429, tokenIndex429, depth429
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
			position493, tokenIndex493, depth493 := position, tokenIndex, depth
			{
				position494 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l493
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position494)
			}
			return true
		l493:
			position, tokenIndex, depth = position493, tokenIndex493, depth493
			return false
		},
		/* 43 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position495, tokenIndex495, depth495 := position, tokenIndex, depth
			{
				position496 := position
				depth++
				if buffer[position] != rune('"') {
					goto l495
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position496)
			}
			return true
		l495:
			position, tokenIndex, depth = position495, tokenIndex495, depth495
			return false
		},
		/* 44 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> QUOTE_SINGLE) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> QUOTE_DOUBLE))> */
		func() bool {
			position497, tokenIndex497, depth497 := position, tokenIndex, depth
			{
				position498 := position
				depth++
				{
					position499, tokenIndex499, depth499 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l500
					}
					{
						position501 := position
						depth++
					l502:
						{
							position503, tokenIndex503, depth503 := position, tokenIndex, depth
							{
								position504, tokenIndex504, depth504 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l504
								}
								goto l503
							l504:
								position, tokenIndex, depth = position504, tokenIndex504, depth504
							}
							if !_rules[ruleCHAR]() {
								goto l503
							}
							goto l502
						l503:
							position, tokenIndex, depth = position503, tokenIndex503, depth503
						}
						depth--
						add(rulePegText, position501)
					}
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l500
					}
					goto l499
				l500:
					position, tokenIndex, depth = position499, tokenIndex499, depth499
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l497
					}
					{
						position505 := position
						depth++
					l506:
						{
							position507, tokenIndex507, depth507 := position, tokenIndex, depth
							{
								position508, tokenIndex508, depth508 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l508
								}
								goto l507
							l508:
								position, tokenIndex, depth = position508, tokenIndex508, depth508
							}
							if !_rules[ruleCHAR]() {
								goto l507
							}
							goto l506
						l507:
							position, tokenIndex, depth = position507, tokenIndex507, depth507
						}
						depth--
						add(rulePegText, position505)
					}
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l497
					}
				}
			l499:
				depth--
				add(ruleSTRING, position498)
			}
			return true
		l497:
			position, tokenIndex, depth = position497, tokenIndex497, depth497
			return false
		},
		/* 45 CHAR <- <(('\\' ((&('"') QUOTE_DOUBLE) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position509, tokenIndex509, depth509 := position, tokenIndex, depth
			{
				position510 := position
				depth++
				{
					position511, tokenIndex511, depth511 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l512
					}
					position++
					{
						switch buffer[position] {
						case '"':
							if !_rules[ruleQUOTE_DOUBLE]() {
								goto l512
							}
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l512
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l512
							}
							break
						}
					}

					goto l511
				l512:
					position, tokenIndex, depth = position511, tokenIndex511, depth511
					{
						position514, tokenIndex514, depth514 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l514
						}
						goto l509
					l514:
						position, tokenIndex, depth = position514, tokenIndex514, depth514
					}
					if !matchDot() {
						goto l509
					}
				}
			l511:
				depth--
				add(ruleCHAR, position510)
			}
			return true
		l509:
			position, tokenIndex, depth = position509, tokenIndex509, depth509
			return false
		},
		/* 46 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position515, tokenIndex515, depth515 := position, tokenIndex, depth
			{
				position516 := position
				depth++
				{
					position517, tokenIndex517, depth517 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l518
					}
					position++
					goto l517
				l518:
					position, tokenIndex, depth = position517, tokenIndex517, depth517
					if buffer[position] != rune('\\') {
						goto l515
					}
					position++
				}
			l517:
				depth--
				add(ruleESCAPE_CLASS, position516)
			}
			return true
		l515:
			position, tokenIndex, depth = position515, tokenIndex515, depth515
			return false
		},
		/* 47 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position519, tokenIndex519, depth519 := position, tokenIndex, depth
			{
				position520 := position
				depth++
				{
					position521 := position
					depth++
					{
						position522, tokenIndex522, depth522 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l522
						}
						position++
						goto l523
					l522:
						position, tokenIndex, depth = position522, tokenIndex522, depth522
					}
				l523:
					{
						position524 := position
						depth++
						{
							position525, tokenIndex525, depth525 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l526
							}
							position++
							goto l525
						l526:
							position, tokenIndex, depth = position525, tokenIndex525, depth525
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l519
							}
							position++
						l527:
							{
								position528, tokenIndex528, depth528 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l528
								}
								position++
								goto l527
							l528:
								position, tokenIndex, depth = position528, tokenIndex528, depth528
							}
						}
					l525:
						depth--
						add(ruleNUMBER_NATURAL, position524)
					}
					depth--
					add(ruleNUMBER_INTEGER, position521)
				}
				{
					position529, tokenIndex529, depth529 := position, tokenIndex, depth
					{
						position531 := position
						depth++
						if buffer[position] != rune('.') {
							goto l529
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l529
						}
						position++
					l532:
						{
							position533, tokenIndex533, depth533 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l533
							}
							position++
							goto l532
						l533:
							position, tokenIndex, depth = position533, tokenIndex533, depth533
						}
						depth--
						add(ruleNUMBER_FRACTION, position531)
					}
					goto l530
				l529:
					position, tokenIndex, depth = position529, tokenIndex529, depth529
				}
			l530:
				{
					position534, tokenIndex534, depth534 := position, tokenIndex, depth
					{
						position536 := position
						depth++
						{
							position537, tokenIndex537, depth537 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l538
							}
							position++
							goto l537
						l538:
							position, tokenIndex, depth = position537, tokenIndex537, depth537
							if buffer[position] != rune('E') {
								goto l534
							}
							position++
						}
					l537:
						{
							position539, tokenIndex539, depth539 := position, tokenIndex, depth
							{
								position541, tokenIndex541, depth541 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l542
								}
								position++
								goto l541
							l542:
								position, tokenIndex, depth = position541, tokenIndex541, depth541
								if buffer[position] != rune('-') {
									goto l539
								}
								position++
							}
						l541:
							goto l540
						l539:
							position, tokenIndex, depth = position539, tokenIndex539, depth539
						}
					l540:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l534
						}
						position++
					l543:
						{
							position544, tokenIndex544, depth544 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l544
							}
							position++
							goto l543
						l544:
							position, tokenIndex, depth = position544, tokenIndex544, depth544
						}
						depth--
						add(ruleNUMBER_EXP, position536)
					}
					goto l535
				l534:
					position, tokenIndex, depth = position534, tokenIndex534, depth534
				}
			l535:
				depth--
				add(ruleNUMBER, position520)
			}
			return true
		l519:
			position, tokenIndex, depth = position519, tokenIndex519, depth519
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
			position549, tokenIndex549, depth549 := position, tokenIndex, depth
			{
				position550 := position
				depth++
				if buffer[position] != rune('(') {
					goto l549
				}
				position++
				depth--
				add(rulePAREN_OPEN, position550)
			}
			return true
		l549:
			position, tokenIndex, depth = position549, tokenIndex549, depth549
			return false
		},
		/* 53 PAREN_CLOSE <- <')'> */
		func() bool {
			position551, tokenIndex551, depth551 := position, tokenIndex, depth
			{
				position552 := position
				depth++
				if buffer[position] != rune(')') {
					goto l551
				}
				position++
				depth--
				add(rulePAREN_CLOSE, position552)
			}
			return true
		l551:
			position, tokenIndex, depth = position551, tokenIndex551, depth551
			return false
		},
		/* 54 COMMA <- <','> */
		func() bool {
			position553, tokenIndex553, depth553 := position, tokenIndex, depth
			{
				position554 := position
				depth++
				if buffer[position] != rune(',') {
					goto l553
				}
				position++
				depth--
				add(ruleCOMMA, position554)
			}
			return true
		l553:
			position, tokenIndex, depth = position553, tokenIndex553, depth553
			return false
		},
		/* 55 _ <- <SPACE*> */
		func() bool {
			{
				position556 := position
				depth++
			l557:
				{
					position558, tokenIndex558, depth558 := position, tokenIndex, depth
					{
						position559 := position
						depth++
						{
							switch buffer[position] {
							case '\t':
								if buffer[position] != rune('\t') {
									goto l558
								}
								position++
								break
							case '\n':
								if buffer[position] != rune('\n') {
									goto l558
								}
								position++
								break
							default:
								if buffer[position] != rune(' ') {
									goto l558
								}
								position++
								break
							}
						}

						depth--
						add(ruleSPACE, position559)
					}
					goto l557
				l558:
					position, tokenIndex, depth = position558, tokenIndex558, depth558
				}
				depth--
				add(rule_, position556)
			}
			return true
		},
		/* 56 KEY <- <!ID_CONT> */
		func() bool {
			position561, tokenIndex561, depth561 := position, tokenIndex, depth
			{
				position562 := position
				depth++
				{
					position563, tokenIndex563, depth563 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l563
					}
					goto l561
				l563:
					position, tokenIndex, depth = position563, tokenIndex563, depth563
				}
				depth--
				add(ruleKEY, position562)
			}
			return true
		l561:
			position, tokenIndex, depth = position561, tokenIndex561, depth561
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
