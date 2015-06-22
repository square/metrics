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
	ruleexpression_1
	ruleexpression_2
	ruleexpression_3
	ruleexpression_4
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
	"expression_1",
	"expression_2",
	"expression_3",
	"expression_4",
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
		/* 9 expression_start <- <expression_1> */
		func() bool {
			position134, tokenIndex134, depth134 := position, tokenIndex, depth
			{
				position135 := position
				depth++
				{
					position136 := position
					depth++
					if !_rules[ruleexpression_2]() {
						goto l134
					}
				l137:
					{
						position138, tokenIndex138, depth138 := position, tokenIndex, depth
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
						if !_rules[ruleexpression_2]() {
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
					add(ruleexpression_1, position136)
				}
				depth--
				add(ruleexpression_start, position135)
			}
			return true
		l134:
			position, tokenIndex, depth = position134, tokenIndex134, depth134
			return false
		},
		/* 10 expression_1 <- <(expression_2 (((_ OP_ADD Action14) / (_ OP_SUB Action15)) expression_2 Action16)*)> */
		nil,
		/* 11 expression_2 <- <(expression_3 (((_ OP_DIV Action17) / (_ OP_MULT Action18)) expression_3 Action19)*)> */
		func() bool {
			position147, tokenIndex147, depth147 := position, tokenIndex, depth
			{
				position148 := position
				depth++
				if !_rules[ruleexpression_3]() {
					goto l147
				}
			l149:
				{
					position150, tokenIndex150, depth150 := position, tokenIndex, depth
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
					if !_rules[ruleexpression_3]() {
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
				add(ruleexpression_2, position148)
			}
			return true
		l147:
			position, tokenIndex, depth = position147, tokenIndex147, depth147
			return false
		},
		/* 12 expression_3 <- <(expression_4 (_ OP_PIPE _ <IDENTIFIER> Action20 ((_ PAREN_OPEN (expressionList / Action21) Action22 groupByClause? _ PAREN_CLOSE) / Action23) Action24)*)> */
		func() bool {
			position158, tokenIndex158, depth158 := position, tokenIndex, depth
			{
				position159 := position
				depth++
				{
					position160 := position
					depth++
					{
						position161, tokenIndex161, depth161 := position, tokenIndex, depth
						{
							position163 := position
							depth++
							if !_rules[rule_]() {
								goto l162
							}
							{
								position164 := position
								depth++
								if !_rules[ruleIDENTIFIER]() {
									goto l162
								}
								depth--
								add(rulePegText, position164)
							}
							{
								add(ruleAction27, position)
							}
							if !_rules[rule_]() {
								goto l162
							}
							if !_rules[rulePAREN_OPEN]() {
								goto l162
							}
							if !_rules[ruleexpressionList]() {
								goto l162
							}
							{
								add(ruleAction28, position)
							}
							{
								position167, tokenIndex167, depth167 := position, tokenIndex, depth
								if !_rules[rulegroupByClause]() {
									goto l167
								}
								goto l168
							l167:
								position, tokenIndex, depth = position167, tokenIndex167, depth167
							}
						l168:
							if !_rules[rule_]() {
								goto l162
							}
							if !_rules[rulePAREN_CLOSE]() {
								goto l162
							}
							{
								add(ruleAction29, position)
							}
							depth--
							add(ruleexpression_function, position163)
						}
						goto l161
					l162:
						position, tokenIndex, depth = position161, tokenIndex161, depth161
						{
							position171 := position
							depth++
							if !_rules[rule_]() {
								goto l170
							}
							{
								position172 := position
								depth++
								if !_rules[ruleIDENTIFIER]() {
									goto l170
								}
								depth--
								add(rulePegText, position172)
							}
							{
								add(ruleAction30, position)
							}
							{
								position174, tokenIndex174, depth174 := position, tokenIndex, depth
								{
									position176, tokenIndex176, depth176 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l177
									}
									if buffer[position] != rune('[') {
										goto l177
									}
									position++
									if !_rules[rulepredicate_1]() {
										goto l177
									}
									if !_rules[rule_]() {
										goto l177
									}
									if buffer[position] != rune(']') {
										goto l177
									}
									position++
									goto l176
								l177:
									position, tokenIndex, depth = position176, tokenIndex176, depth176
									{
										add(ruleAction31, position)
									}
								}
							l176:
								goto l175

								position, tokenIndex, depth = position174, tokenIndex174, depth174
							}
						l175:
							{
								add(ruleAction32, position)
							}
							depth--
							add(ruleexpression_metric, position171)
						}
						goto l161
					l170:
						position, tokenIndex, depth = position161, tokenIndex161, depth161
						if !_rules[rule_]() {
							goto l180
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l180
						}
						if !_rules[ruleexpression_start]() {
							goto l180
						}
						if !_rules[rule_]() {
							goto l180
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l180
						}
						goto l161
					l180:
						position, tokenIndex, depth = position161, tokenIndex161, depth161
						if !_rules[rule_]() {
							goto l181
						}
						{
							position182 := position
							depth++
							if !_rules[ruleNUMBER]() {
								goto l181
							}
							depth--
							add(rulePegText, position182)
						}
						{
							add(ruleAction25, position)
						}
						goto l161
					l181:
						position, tokenIndex, depth = position161, tokenIndex161, depth161
						if !_rules[rule_]() {
							goto l158
						}
						if !_rules[ruleSTRING]() {
							goto l158
						}
						{
							add(ruleAction26, position)
						}
					}
				l161:
					depth--
					add(ruleexpression_4, position160)
				}
			l185:
				{
					position186, tokenIndex186, depth186 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l186
					}
					{
						position187 := position
						depth++
						if buffer[position] != rune('|') {
							goto l186
						}
						position++
						depth--
						add(ruleOP_PIPE, position187)
					}
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
						add(ruleAction20, position)
					}
					{
						position190, tokenIndex190, depth190 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l191
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l191
						}
						{
							position192, tokenIndex192, depth192 := position, tokenIndex, depth
							if !_rules[ruleexpressionList]() {
								goto l193
							}
							goto l192
						l193:
							position, tokenIndex, depth = position192, tokenIndex192, depth192
							{
								add(ruleAction21, position)
							}
						}
					l192:
						{
							add(ruleAction22, position)
						}
						{
							position196, tokenIndex196, depth196 := position, tokenIndex, depth
							if !_rules[rulegroupByClause]() {
								goto l196
							}
							goto l197
						l196:
							position, tokenIndex, depth = position196, tokenIndex196, depth196
						}
					l197:
						if !_rules[rule_]() {
							goto l191
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l191
						}
						goto l190
					l191:
						position, tokenIndex, depth = position190, tokenIndex190, depth190
						{
							add(ruleAction23, position)
						}
					}
				l190:
					{
						add(ruleAction24, position)
					}
					goto l185
				l186:
					position, tokenIndex, depth = position186, tokenIndex186, depth186
				}
				depth--
				add(ruleexpression_3, position159)
			}
			return true
		l158:
			position, tokenIndex, depth = position158, tokenIndex158, depth158
			return false
		},
		/* 13 expression_4 <- <(expression_function / expression_metric / (_ PAREN_OPEN expression_start _ PAREN_CLOSE) / (_ <NUMBER> Action25) / (_ STRING Action26))> */
		nil,
		/* 14 expression_function <- <(_ <IDENTIFIER> Action27 _ PAREN_OPEN expressionList Action28 groupByClause? _ PAREN_CLOSE Action29)> */
		nil,
		/* 15 expression_metric <- <(_ <IDENTIFIER> Action30 ((_ '[' predicate_1 _ ']') / Action31)? Action32)> */
		nil,
		/* 16 groupByClause <- <(_ (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P')) KEY _ (('b' / 'B') ('y' / 'Y')) KEY _ <COLUMN_NAME> Action33 (_ COMMA _ <COLUMN_NAME> Action34)*)> */
		func() bool {
			position203, tokenIndex203, depth203 := position, tokenIndex, depth
			{
				position204 := position
				depth++
				if !_rules[rule_]() {
					goto l203
				}
				{
					position205, tokenIndex205, depth205 := position, tokenIndex, depth
					if buffer[position] != rune('g') {
						goto l206
					}
					position++
					goto l205
				l206:
					position, tokenIndex, depth = position205, tokenIndex205, depth205
					if buffer[position] != rune('G') {
						goto l203
					}
					position++
				}
			l205:
				{
					position207, tokenIndex207, depth207 := position, tokenIndex, depth
					if buffer[position] != rune('r') {
						goto l208
					}
					position++
					goto l207
				l208:
					position, tokenIndex, depth = position207, tokenIndex207, depth207
					if buffer[position] != rune('R') {
						goto l203
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
						goto l203
					}
					position++
				}
			l209:
				{
					position211, tokenIndex211, depth211 := position, tokenIndex, depth
					if buffer[position] != rune('u') {
						goto l212
					}
					position++
					goto l211
				l212:
					position, tokenIndex, depth = position211, tokenIndex211, depth211
					if buffer[position] != rune('U') {
						goto l203
					}
					position++
				}
			l211:
				{
					position213, tokenIndex213, depth213 := position, tokenIndex, depth
					if buffer[position] != rune('p') {
						goto l214
					}
					position++
					goto l213
				l214:
					position, tokenIndex, depth = position213, tokenIndex213, depth213
					if buffer[position] != rune('P') {
						goto l203
					}
					position++
				}
			l213:
				if !_rules[ruleKEY]() {
					goto l203
				}
				if !_rules[rule_]() {
					goto l203
				}
				{
					position215, tokenIndex215, depth215 := position, tokenIndex, depth
					if buffer[position] != rune('b') {
						goto l216
					}
					position++
					goto l215
				l216:
					position, tokenIndex, depth = position215, tokenIndex215, depth215
					if buffer[position] != rune('B') {
						goto l203
					}
					position++
				}
			l215:
				{
					position217, tokenIndex217, depth217 := position, tokenIndex, depth
					if buffer[position] != rune('y') {
						goto l218
					}
					position++
					goto l217
				l218:
					position, tokenIndex, depth = position217, tokenIndex217, depth217
					if buffer[position] != rune('Y') {
						goto l203
					}
					position++
				}
			l217:
				if !_rules[ruleKEY]() {
					goto l203
				}
				if !_rules[rule_]() {
					goto l203
				}
				{
					position219 := position
					depth++
					if !_rules[ruleCOLUMN_NAME]() {
						goto l203
					}
					depth--
					add(rulePegText, position219)
				}
				{
					add(ruleAction33, position)
				}
			l221:
				{
					position222, tokenIndex222, depth222 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l222
					}
					if !_rules[ruleCOMMA]() {
						goto l222
					}
					if !_rules[rule_]() {
						goto l222
					}
					{
						position223 := position
						depth++
						if !_rules[ruleCOLUMN_NAME]() {
							goto l222
						}
						depth--
						add(rulePegText, position223)
					}
					{
						add(ruleAction34, position)
					}
					goto l221
				l222:
					position, tokenIndex, depth = position222, tokenIndex222, depth222
				}
				depth--
				add(rulegroupByClause, position204)
			}
			return true
		l203:
			position, tokenIndex, depth = position203, tokenIndex203, depth203
			return false
		},
		/* 17 predicateClause <- <(_ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY _ predicate_1)> */
		nil,
		/* 18 predicate_1 <- <((predicate_2 _ OP_OR predicate_1 Action35) / predicate_2)> */
		func() bool {
			position226, tokenIndex226, depth226 := position, tokenIndex, depth
			{
				position227 := position
				depth++
				{
					position228, tokenIndex228, depth228 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l229
					}
					if !_rules[rule_]() {
						goto l229
					}
					{
						position230 := position
						depth++
						{
							position231, tokenIndex231, depth231 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l232
							}
							position++
							goto l231
						l232:
							position, tokenIndex, depth = position231, tokenIndex231, depth231
							if buffer[position] != rune('O') {
								goto l229
							}
							position++
						}
					l231:
						{
							position233, tokenIndex233, depth233 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l234
							}
							position++
							goto l233
						l234:
							position, tokenIndex, depth = position233, tokenIndex233, depth233
							if buffer[position] != rune('R') {
								goto l229
							}
							position++
						}
					l233:
						if !_rules[ruleKEY]() {
							goto l229
						}
						depth--
						add(ruleOP_OR, position230)
					}
					if !_rules[rulepredicate_1]() {
						goto l229
					}
					{
						add(ruleAction35, position)
					}
					goto l228
				l229:
					position, tokenIndex, depth = position228, tokenIndex228, depth228
					if !_rules[rulepredicate_2]() {
						goto l226
					}
				}
			l228:
				depth--
				add(rulepredicate_1, position227)
			}
			return true
		l226:
			position, tokenIndex, depth = position226, tokenIndex226, depth226
			return false
		},
		/* 19 predicate_2 <- <((predicate_3 _ OP_AND predicate_2 Action36) / predicate_3)> */
		func() bool {
			position236, tokenIndex236, depth236 := position, tokenIndex, depth
			{
				position237 := position
				depth++
				{
					position238, tokenIndex238, depth238 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l239
					}
					if !_rules[rule_]() {
						goto l239
					}
					{
						position240 := position
						depth++
						{
							position241, tokenIndex241, depth241 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l242
							}
							position++
							goto l241
						l242:
							position, tokenIndex, depth = position241, tokenIndex241, depth241
							if buffer[position] != rune('A') {
								goto l239
							}
							position++
						}
					l241:
						{
							position243, tokenIndex243, depth243 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l244
							}
							position++
							goto l243
						l244:
							position, tokenIndex, depth = position243, tokenIndex243, depth243
							if buffer[position] != rune('N') {
								goto l239
							}
							position++
						}
					l243:
						{
							position245, tokenIndex245, depth245 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l246
							}
							position++
							goto l245
						l246:
							position, tokenIndex, depth = position245, tokenIndex245, depth245
							if buffer[position] != rune('D') {
								goto l239
							}
							position++
						}
					l245:
						if !_rules[ruleKEY]() {
							goto l239
						}
						depth--
						add(ruleOP_AND, position240)
					}
					if !_rules[rulepredicate_2]() {
						goto l239
					}
					{
						add(ruleAction36, position)
					}
					goto l238
				l239:
					position, tokenIndex, depth = position238, tokenIndex238, depth238
					if !_rules[rulepredicate_3]() {
						goto l236
					}
				}
			l238:
				depth--
				add(rulepredicate_2, position237)
			}
			return true
		l236:
			position, tokenIndex, depth = position236, tokenIndex236, depth236
			return false
		},
		/* 20 predicate_3 <- <((_ OP_NOT predicate_3 Action37) / (_ PAREN_OPEN predicate_1 _ PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position248, tokenIndex248, depth248 := position, tokenIndex, depth
			{
				position249 := position
				depth++
				{
					position250, tokenIndex250, depth250 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l251
					}
					{
						position252 := position
						depth++
						{
							position253, tokenIndex253, depth253 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l254
							}
							position++
							goto l253
						l254:
							position, tokenIndex, depth = position253, tokenIndex253, depth253
							if buffer[position] != rune('N') {
								goto l251
							}
							position++
						}
					l253:
						{
							position255, tokenIndex255, depth255 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l256
							}
							position++
							goto l255
						l256:
							position, tokenIndex, depth = position255, tokenIndex255, depth255
							if buffer[position] != rune('O') {
								goto l251
							}
							position++
						}
					l255:
						{
							position257, tokenIndex257, depth257 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l258
							}
							position++
							goto l257
						l258:
							position, tokenIndex, depth = position257, tokenIndex257, depth257
							if buffer[position] != rune('T') {
								goto l251
							}
							position++
						}
					l257:
						if !_rules[ruleKEY]() {
							goto l251
						}
						depth--
						add(ruleOP_NOT, position252)
					}
					if !_rules[rulepredicate_3]() {
						goto l251
					}
					{
						add(ruleAction37, position)
					}
					goto l250
				l251:
					position, tokenIndex, depth = position250, tokenIndex250, depth250
					if !_rules[rule_]() {
						goto l260
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l260
					}
					if !_rules[rulepredicate_1]() {
						goto l260
					}
					if !_rules[rule_]() {
						goto l260
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l260
					}
					goto l250
				l260:
					position, tokenIndex, depth = position250, tokenIndex250, depth250
					{
						position261 := position
						depth++
						{
							position262, tokenIndex262, depth262 := position, tokenIndex, depth
							if !_rules[ruletagName]() {
								goto l263
							}
							if !_rules[rule_]() {
								goto l263
							}
							if buffer[position] != rune('=') {
								goto l263
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l263
							}
							{
								add(ruleAction38, position)
							}
							goto l262
						l263:
							position, tokenIndex, depth = position262, tokenIndex262, depth262
							if !_rules[ruletagName]() {
								goto l265
							}
							if !_rules[rule_]() {
								goto l265
							}
							if buffer[position] != rune('!') {
								goto l265
							}
							position++
							if buffer[position] != rune('=') {
								goto l265
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l265
							}
							{
								add(ruleAction39, position)
							}
							goto l262
						l265:
							position, tokenIndex, depth = position262, tokenIndex262, depth262
							if !_rules[ruletagName]() {
								goto l267
							}
							if !_rules[rule_]() {
								goto l267
							}
							{
								position268, tokenIndex268, depth268 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l269
								}
								position++
								goto l268
							l269:
								position, tokenIndex, depth = position268, tokenIndex268, depth268
								if buffer[position] != rune('M') {
									goto l267
								}
								position++
							}
						l268:
							{
								position270, tokenIndex270, depth270 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l271
								}
								position++
								goto l270
							l271:
								position, tokenIndex, depth = position270, tokenIndex270, depth270
								if buffer[position] != rune('A') {
									goto l267
								}
								position++
							}
						l270:
							{
								position272, tokenIndex272, depth272 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l273
								}
								position++
								goto l272
							l273:
								position, tokenIndex, depth = position272, tokenIndex272, depth272
								if buffer[position] != rune('T') {
									goto l267
								}
								position++
							}
						l272:
							{
								position274, tokenIndex274, depth274 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l275
								}
								position++
								goto l274
							l275:
								position, tokenIndex, depth = position274, tokenIndex274, depth274
								if buffer[position] != rune('C') {
									goto l267
								}
								position++
							}
						l274:
							{
								position276, tokenIndex276, depth276 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l277
								}
								position++
								goto l276
							l277:
								position, tokenIndex, depth = position276, tokenIndex276, depth276
								if buffer[position] != rune('H') {
									goto l267
								}
								position++
							}
						l276:
							{
								position278, tokenIndex278, depth278 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l279
								}
								position++
								goto l278
							l279:
								position, tokenIndex, depth = position278, tokenIndex278, depth278
								if buffer[position] != rune('E') {
									goto l267
								}
								position++
							}
						l278:
							{
								position280, tokenIndex280, depth280 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l281
								}
								position++
								goto l280
							l281:
								position, tokenIndex, depth = position280, tokenIndex280, depth280
								if buffer[position] != rune('S') {
									goto l267
								}
								position++
							}
						l280:
							if !_rules[ruleKEY]() {
								goto l267
							}
							if !_rules[ruleliteralString]() {
								goto l267
							}
							{
								add(ruleAction40, position)
							}
							goto l262
						l267:
							position, tokenIndex, depth = position262, tokenIndex262, depth262
							if !_rules[ruletagName]() {
								goto l248
							}
							if !_rules[rule_]() {
								goto l248
							}
							{
								position283, tokenIndex283, depth283 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l284
								}
								position++
								goto l283
							l284:
								position, tokenIndex, depth = position283, tokenIndex283, depth283
								if buffer[position] != rune('I') {
									goto l248
								}
								position++
							}
						l283:
							{
								position285, tokenIndex285, depth285 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l286
								}
								position++
								goto l285
							l286:
								position, tokenIndex, depth = position285, tokenIndex285, depth285
								if buffer[position] != rune('N') {
									goto l248
								}
								position++
							}
						l285:
							if !_rules[ruleKEY]() {
								goto l248
							}
							{
								position287 := position
								depth++
								{
									add(ruleAction43, position)
								}
								if !_rules[rule_]() {
									goto l248
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l248
								}
								if !_rules[ruleliteralListString]() {
									goto l248
								}
							l289:
								{
									position290, tokenIndex290, depth290 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l290
									}
									if !_rules[ruleCOMMA]() {
										goto l290
									}
									if !_rules[ruleliteralListString]() {
										goto l290
									}
									goto l289
								l290:
									position, tokenIndex, depth = position290, tokenIndex290, depth290
								}
								if !_rules[rule_]() {
									goto l248
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l248
								}
								depth--
								add(ruleliteralList, position287)
							}
							{
								add(ruleAction41, position)
							}
						}
					l262:
						depth--
						add(ruletagMatcher, position261)
					}
				}
			l250:
				depth--
				add(rulepredicate_3, position249)
			}
			return true
		l248:
			position, tokenIndex, depth = position248, tokenIndex248, depth248
			return false
		},
		/* 21 tagMatcher <- <((tagName _ '=' literalString Action38) / (tagName _ ('!' '=') literalString Action39) / (tagName _ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) KEY literalString Action40) / (tagName _ (('i' / 'I') ('n' / 'N')) KEY literalList Action41))> */
		nil,
		/* 22 literalString <- <(_ STRING Action42)> */
		func() bool {
			position293, tokenIndex293, depth293 := position, tokenIndex, depth
			{
				position294 := position
				depth++
				if !_rules[rule_]() {
					goto l293
				}
				if !_rules[ruleSTRING]() {
					goto l293
				}
				{
					add(ruleAction42, position)
				}
				depth--
				add(ruleliteralString, position294)
			}
			return true
		l293:
			position, tokenIndex, depth = position293, tokenIndex293, depth293
			return false
		},
		/* 23 literalList <- <(Action43 _ PAREN_OPEN literalListString (_ COMMA literalListString)* _ PAREN_CLOSE)> */
		nil,
		/* 24 literalListString <- <(_ STRING Action44)> */
		func() bool {
			position297, tokenIndex297, depth297 := position, tokenIndex, depth
			{
				position298 := position
				depth++
				if !_rules[rule_]() {
					goto l297
				}
				if !_rules[ruleSTRING]() {
					goto l297
				}
				{
					add(ruleAction44, position)
				}
				depth--
				add(ruleliteralListString, position298)
			}
			return true
		l297:
			position, tokenIndex, depth = position297, tokenIndex297, depth297
			return false
		},
		/* 25 tagName <- <(_ <TAG_NAME> Action45)> */
		func() bool {
			position300, tokenIndex300, depth300 := position, tokenIndex, depth
			{
				position301 := position
				depth++
				if !_rules[rule_]() {
					goto l300
				}
				{
					position302 := position
					depth++
					{
						position303 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l300
						}
						depth--
						add(ruleTAG_NAME, position303)
					}
					depth--
					add(rulePegText, position302)
				}
				{
					add(ruleAction45, position)
				}
				depth--
				add(ruletagName, position301)
			}
			return true
		l300:
			position, tokenIndex, depth = position300, tokenIndex300, depth300
			return false
		},
		/* 26 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position305, tokenIndex305, depth305 := position, tokenIndex, depth
			{
				position306 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l305
				}
				depth--
				add(ruleCOLUMN_NAME, position306)
			}
			return true
		l305:
			position, tokenIndex, depth = position305, tokenIndex305, depth305
			return false
		},
		/* 27 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 28 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 29 IDENTIFIER <- <(('`' CHAR* '`') / (_ !(KEYWORD KEY) ID_SEGMENT ('.' ID_SEGMENT)*))> */
		func() bool {
			position309, tokenIndex309, depth309 := position, tokenIndex, depth
			{
				position310 := position
				depth++
				{
					position311, tokenIndex311, depth311 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l312
					}
					position++
				l313:
					{
						position314, tokenIndex314, depth314 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l314
						}
						goto l313
					l314:
						position, tokenIndex, depth = position314, tokenIndex314, depth314
					}
					if buffer[position] != rune('`') {
						goto l312
					}
					position++
					goto l311
				l312:
					position, tokenIndex, depth = position311, tokenIndex311, depth311
					if !_rules[rule_]() {
						goto l309
					}
					{
						position315, tokenIndex315, depth315 := position, tokenIndex, depth
						{
							position316 := position
							depth++
							{
								position317, tokenIndex317, depth317 := position, tokenIndex, depth
								{
									position319, tokenIndex319, depth319 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l320
									}
									position++
									goto l319
								l320:
									position, tokenIndex, depth = position319, tokenIndex319, depth319
									if buffer[position] != rune('A') {
										goto l318
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
										goto l318
									}
									position++
								}
							l321:
								{
									position323, tokenIndex323, depth323 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l324
									}
									position++
									goto l323
								l324:
									position, tokenIndex, depth = position323, tokenIndex323, depth323
									if buffer[position] != rune('L') {
										goto l318
									}
									position++
								}
							l323:
								goto l317
							l318:
								position, tokenIndex, depth = position317, tokenIndex317, depth317
								{
									position326, tokenIndex326, depth326 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l327
									}
									position++
									goto l326
								l327:
									position, tokenIndex, depth = position326, tokenIndex326, depth326
									if buffer[position] != rune('A') {
										goto l325
									}
									position++
								}
							l326:
								{
									position328, tokenIndex328, depth328 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l329
									}
									position++
									goto l328
								l329:
									position, tokenIndex, depth = position328, tokenIndex328, depth328
									if buffer[position] != rune('N') {
										goto l325
									}
									position++
								}
							l328:
								{
									position330, tokenIndex330, depth330 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l331
									}
									position++
									goto l330
								l331:
									position, tokenIndex, depth = position330, tokenIndex330, depth330
									if buffer[position] != rune('D') {
										goto l325
									}
									position++
								}
							l330:
								goto l317
							l325:
								position, tokenIndex, depth = position317, tokenIndex317, depth317
								{
									position333, tokenIndex333, depth333 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l334
									}
									position++
									goto l333
								l334:
									position, tokenIndex, depth = position333, tokenIndex333, depth333
									if buffer[position] != rune('M') {
										goto l332
									}
									position++
								}
							l333:
								{
									position335, tokenIndex335, depth335 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l336
									}
									position++
									goto l335
								l336:
									position, tokenIndex, depth = position335, tokenIndex335, depth335
									if buffer[position] != rune('A') {
										goto l332
									}
									position++
								}
							l335:
								{
									position337, tokenIndex337, depth337 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l338
									}
									position++
									goto l337
								l338:
									position, tokenIndex, depth = position337, tokenIndex337, depth337
									if buffer[position] != rune('T') {
										goto l332
									}
									position++
								}
							l337:
								{
									position339, tokenIndex339, depth339 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l340
									}
									position++
									goto l339
								l340:
									position, tokenIndex, depth = position339, tokenIndex339, depth339
									if buffer[position] != rune('C') {
										goto l332
									}
									position++
								}
							l339:
								{
									position341, tokenIndex341, depth341 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l342
									}
									position++
									goto l341
								l342:
									position, tokenIndex, depth = position341, tokenIndex341, depth341
									if buffer[position] != rune('H') {
										goto l332
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
										goto l332
									}
									position++
								}
							l343:
								{
									position345, tokenIndex345, depth345 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l346
									}
									position++
									goto l345
								l346:
									position, tokenIndex, depth = position345, tokenIndex345, depth345
									if buffer[position] != rune('S') {
										goto l332
									}
									position++
								}
							l345:
								goto l317
							l332:
								position, tokenIndex, depth = position317, tokenIndex317, depth317
								{
									position348, tokenIndex348, depth348 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l349
									}
									position++
									goto l348
								l349:
									position, tokenIndex, depth = position348, tokenIndex348, depth348
									if buffer[position] != rune('S') {
										goto l347
									}
									position++
								}
							l348:
								{
									position350, tokenIndex350, depth350 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l351
									}
									position++
									goto l350
								l351:
									position, tokenIndex, depth = position350, tokenIndex350, depth350
									if buffer[position] != rune('E') {
										goto l347
									}
									position++
								}
							l350:
								{
									position352, tokenIndex352, depth352 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l353
									}
									position++
									goto l352
								l353:
									position, tokenIndex, depth = position352, tokenIndex352, depth352
									if buffer[position] != rune('L') {
										goto l347
									}
									position++
								}
							l352:
								{
									position354, tokenIndex354, depth354 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l355
									}
									position++
									goto l354
								l355:
									position, tokenIndex, depth = position354, tokenIndex354, depth354
									if buffer[position] != rune('E') {
										goto l347
									}
									position++
								}
							l354:
								{
									position356, tokenIndex356, depth356 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l357
									}
									position++
									goto l356
								l357:
									position, tokenIndex, depth = position356, tokenIndex356, depth356
									if buffer[position] != rune('C') {
										goto l347
									}
									position++
								}
							l356:
								{
									position358, tokenIndex358, depth358 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l359
									}
									position++
									goto l358
								l359:
									position, tokenIndex, depth = position358, tokenIndex358, depth358
									if buffer[position] != rune('T') {
										goto l347
									}
									position++
								}
							l358:
								goto l317
							l347:
								position, tokenIndex, depth = position317, tokenIndex317, depth317
								{
									switch buffer[position] {
									case 'M', 'm':
										{
											position361, tokenIndex361, depth361 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l362
											}
											position++
											goto l361
										l362:
											position, tokenIndex, depth = position361, tokenIndex361, depth361
											if buffer[position] != rune('M') {
												goto l315
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
												goto l315
											}
											position++
										}
									l363:
										{
											position365, tokenIndex365, depth365 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l366
											}
											position++
											goto l365
										l366:
											position, tokenIndex, depth = position365, tokenIndex365, depth365
											if buffer[position] != rune('T') {
												goto l315
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
												goto l315
											}
											position++
										}
									l367:
										{
											position369, tokenIndex369, depth369 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l370
											}
											position++
											goto l369
										l370:
											position, tokenIndex, depth = position369, tokenIndex369, depth369
											if buffer[position] != rune('I') {
												goto l315
											}
											position++
										}
									l369:
										{
											position371, tokenIndex371, depth371 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l372
											}
											position++
											goto l371
										l372:
											position, tokenIndex, depth = position371, tokenIndex371, depth371
											if buffer[position] != rune('C') {
												goto l315
											}
											position++
										}
									l371:
										{
											position373, tokenIndex373, depth373 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l374
											}
											position++
											goto l373
										l374:
											position, tokenIndex, depth = position373, tokenIndex373, depth373
											if buffer[position] != rune('S') {
												goto l315
											}
											position++
										}
									l373:
										break
									case 'W', 'w':
										{
											position375, tokenIndex375, depth375 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l376
											}
											position++
											goto l375
										l376:
											position, tokenIndex, depth = position375, tokenIndex375, depth375
											if buffer[position] != rune('W') {
												goto l315
											}
											position++
										}
									l375:
										{
											position377, tokenIndex377, depth377 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l378
											}
											position++
											goto l377
										l378:
											position, tokenIndex, depth = position377, tokenIndex377, depth377
											if buffer[position] != rune('H') {
												goto l315
											}
											position++
										}
									l377:
										{
											position379, tokenIndex379, depth379 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l380
											}
											position++
											goto l379
										l380:
											position, tokenIndex, depth = position379, tokenIndex379, depth379
											if buffer[position] != rune('E') {
												goto l315
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
												goto l315
											}
											position++
										}
									l381:
										{
											position383, tokenIndex383, depth383 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l384
											}
											position++
											goto l383
										l384:
											position, tokenIndex, depth = position383, tokenIndex383, depth383
											if buffer[position] != rune('E') {
												goto l315
											}
											position++
										}
									l383:
										break
									case 'O', 'o':
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
												goto l315
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
												goto l315
											}
											position++
										}
									l387:
										break
									case 'N', 'n':
										{
											position389, tokenIndex389, depth389 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l390
											}
											position++
											goto l389
										l390:
											position, tokenIndex, depth = position389, tokenIndex389, depth389
											if buffer[position] != rune('N') {
												goto l315
											}
											position++
										}
									l389:
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
												goto l315
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
												goto l315
											}
											position++
										}
									l393:
										break
									case 'I', 'i':
										{
											position395, tokenIndex395, depth395 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l396
											}
											position++
											goto l395
										l396:
											position, tokenIndex, depth = position395, tokenIndex395, depth395
											if buffer[position] != rune('I') {
												goto l315
											}
											position++
										}
									l395:
										{
											position397, tokenIndex397, depth397 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l398
											}
											position++
											goto l397
										l398:
											position, tokenIndex, depth = position397, tokenIndex397, depth397
											if buffer[position] != rune('N') {
												goto l315
											}
											position++
										}
									l397:
										break
									case 'G', 'g':
										{
											position399, tokenIndex399, depth399 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l400
											}
											position++
											goto l399
										l400:
											position, tokenIndex, depth = position399, tokenIndex399, depth399
											if buffer[position] != rune('G') {
												goto l315
											}
											position++
										}
									l399:
										{
											position401, tokenIndex401, depth401 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l402
											}
											position++
											goto l401
										l402:
											position, tokenIndex, depth = position401, tokenIndex401, depth401
											if buffer[position] != rune('R') {
												goto l315
											}
											position++
										}
									l401:
										{
											position403, tokenIndex403, depth403 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l404
											}
											position++
											goto l403
										l404:
											position, tokenIndex, depth = position403, tokenIndex403, depth403
											if buffer[position] != rune('O') {
												goto l315
											}
											position++
										}
									l403:
										{
											position405, tokenIndex405, depth405 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l406
											}
											position++
											goto l405
										l406:
											position, tokenIndex, depth = position405, tokenIndex405, depth405
											if buffer[position] != rune('U') {
												goto l315
											}
											position++
										}
									l405:
										{
											position407, tokenIndex407, depth407 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l408
											}
											position++
											goto l407
										l408:
											position, tokenIndex, depth = position407, tokenIndex407, depth407
											if buffer[position] != rune('P') {
												goto l315
											}
											position++
										}
									l407:
										break
									case 'D', 'd':
										{
											position409, tokenIndex409, depth409 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l410
											}
											position++
											goto l409
										l410:
											position, tokenIndex, depth = position409, tokenIndex409, depth409
											if buffer[position] != rune('D') {
												goto l315
											}
											position++
										}
									l409:
										{
											position411, tokenIndex411, depth411 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l412
											}
											position++
											goto l411
										l412:
											position, tokenIndex, depth = position411, tokenIndex411, depth411
											if buffer[position] != rune('E') {
												goto l315
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
												goto l315
											}
											position++
										}
									l413:
										{
											position415, tokenIndex415, depth415 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l416
											}
											position++
											goto l415
										l416:
											position, tokenIndex, depth = position415, tokenIndex415, depth415
											if buffer[position] != rune('C') {
												goto l315
											}
											position++
										}
									l415:
										{
											position417, tokenIndex417, depth417 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l418
											}
											position++
											goto l417
										l418:
											position, tokenIndex, depth = position417, tokenIndex417, depth417
											if buffer[position] != rune('R') {
												goto l315
											}
											position++
										}
									l417:
										{
											position419, tokenIndex419, depth419 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l420
											}
											position++
											goto l419
										l420:
											position, tokenIndex, depth = position419, tokenIndex419, depth419
											if buffer[position] != rune('I') {
												goto l315
											}
											position++
										}
									l419:
										{
											position421, tokenIndex421, depth421 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l422
											}
											position++
											goto l421
										l422:
											position, tokenIndex, depth = position421, tokenIndex421, depth421
											if buffer[position] != rune('B') {
												goto l315
											}
											position++
										}
									l421:
										{
											position423, tokenIndex423, depth423 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l424
											}
											position++
											goto l423
										l424:
											position, tokenIndex, depth = position423, tokenIndex423, depth423
											if buffer[position] != rune('E') {
												goto l315
											}
											position++
										}
									l423:
										break
									case 'B', 'b':
										{
											position425, tokenIndex425, depth425 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l426
											}
											position++
											goto l425
										l426:
											position, tokenIndex, depth = position425, tokenIndex425, depth425
											if buffer[position] != rune('B') {
												goto l315
											}
											position++
										}
									l425:
										{
											position427, tokenIndex427, depth427 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l428
											}
											position++
											goto l427
										l428:
											position, tokenIndex, depth = position427, tokenIndex427, depth427
											if buffer[position] != rune('Y') {
												goto l315
											}
											position++
										}
									l427:
										break
									case 'A', 'a':
										{
											position429, tokenIndex429, depth429 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l430
											}
											position++
											goto l429
										l430:
											position, tokenIndex, depth = position429, tokenIndex429, depth429
											if buffer[position] != rune('A') {
												goto l315
											}
											position++
										}
									l429:
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
												goto l315
											}
											position++
										}
									l431:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l315
										}
										break
									}
								}

							}
						l317:
							depth--
							add(ruleKEYWORD, position316)
						}
						if !_rules[ruleKEY]() {
							goto l315
						}
						goto l309
					l315:
						position, tokenIndex, depth = position315, tokenIndex315, depth315
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l309
					}
				l433:
					{
						position434, tokenIndex434, depth434 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l434
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l434
						}
						goto l433
					l434:
						position, tokenIndex, depth = position434, tokenIndex434, depth434
					}
				}
			l311:
				depth--
				add(ruleIDENTIFIER, position310)
			}
			return true
		l309:
			position, tokenIndex, depth = position309, tokenIndex309, depth309
			return false
		},
		/* 30 TIMESTAMP <- <((_ <(NUMBER ([a-z] / [A-Z])*)>) / (_ STRING) / (_ <(('n' / 'N') ('o' / 'O') ('w' / 'W'))>))> */
		nil,
		/* 31 ID_SEGMENT <- <(_ ID_START ID_CONT*)> */
		func() bool {
			position436, tokenIndex436, depth436 := position, tokenIndex, depth
			{
				position437 := position
				depth++
				if !_rules[rule_]() {
					goto l436
				}
				if !_rules[ruleID_START]() {
					goto l436
				}
			l438:
				{
					position439, tokenIndex439, depth439 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l439
					}
					goto l438
				l439:
					position, tokenIndex, depth = position439, tokenIndex439, depth439
				}
				depth--
				add(ruleID_SEGMENT, position437)
			}
			return true
		l436:
			position, tokenIndex, depth = position436, tokenIndex436, depth436
			return false
		},
		/* 32 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position440, tokenIndex440, depth440 := position, tokenIndex, depth
			{
				position441 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l440
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l440
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l440
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position441)
			}
			return true
		l440:
			position, tokenIndex, depth = position440, tokenIndex440, depth440
			return false
		},
		/* 33 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position443, tokenIndex443, depth443 := position, tokenIndex, depth
			{
				position444 := position
				depth++
				{
					position445, tokenIndex445, depth445 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l446
					}
					goto l445
				l446:
					position, tokenIndex, depth = position445, tokenIndex445, depth445
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l443
					}
					position++
				}
			l445:
				depth--
				add(ruleID_CONT, position444)
			}
			return true
		l443:
			position, tokenIndex, depth = position443, tokenIndex443, depth443
			return false
		},
		/* 34 PROPERTY_KEY <- <(((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> KEY _ (('b' / 'B') ('y' / 'Y')))) | (&('R' | 'r') <(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))>) | (&('T' | 't') <(('t' / 'T') ('o' / 'O'))>) | (&('F' | 'f') <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))>)) KEY)> */
		func() bool {
			position447, tokenIndex447, depth447 := position, tokenIndex, depth
			{
				position448 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position450 := position
							depth++
							{
								position451, tokenIndex451, depth451 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l452
								}
								position++
								goto l451
							l452:
								position, tokenIndex, depth = position451, tokenIndex451, depth451
								if buffer[position] != rune('S') {
									goto l447
								}
								position++
							}
						l451:
							{
								position453, tokenIndex453, depth453 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l454
								}
								position++
								goto l453
							l454:
								position, tokenIndex, depth = position453, tokenIndex453, depth453
								if buffer[position] != rune('A') {
									goto l447
								}
								position++
							}
						l453:
							{
								position455, tokenIndex455, depth455 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l456
								}
								position++
								goto l455
							l456:
								position, tokenIndex, depth = position455, tokenIndex455, depth455
								if buffer[position] != rune('M') {
									goto l447
								}
								position++
							}
						l455:
							{
								position457, tokenIndex457, depth457 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l458
								}
								position++
								goto l457
							l458:
								position, tokenIndex, depth = position457, tokenIndex457, depth457
								if buffer[position] != rune('P') {
									goto l447
								}
								position++
							}
						l457:
							{
								position459, tokenIndex459, depth459 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l460
								}
								position++
								goto l459
							l460:
								position, tokenIndex, depth = position459, tokenIndex459, depth459
								if buffer[position] != rune('L') {
									goto l447
								}
								position++
							}
						l459:
							{
								position461, tokenIndex461, depth461 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l462
								}
								position++
								goto l461
							l462:
								position, tokenIndex, depth = position461, tokenIndex461, depth461
								if buffer[position] != rune('E') {
									goto l447
								}
								position++
							}
						l461:
							depth--
							add(rulePegText, position450)
						}
						if !_rules[ruleKEY]() {
							goto l447
						}
						if !_rules[rule_]() {
							goto l447
						}
						{
							position463, tokenIndex463, depth463 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l464
							}
							position++
							goto l463
						l464:
							position, tokenIndex, depth = position463, tokenIndex463, depth463
							if buffer[position] != rune('B') {
								goto l447
							}
							position++
						}
					l463:
						{
							position465, tokenIndex465, depth465 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l466
							}
							position++
							goto l465
						l466:
							position, tokenIndex, depth = position465, tokenIndex465, depth465
							if buffer[position] != rune('Y') {
								goto l447
							}
							position++
						}
					l465:
						break
					case 'R', 'r':
						{
							position467 := position
							depth++
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
									goto l447
								}
								position++
							}
						l468:
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
									goto l447
								}
								position++
							}
						l470:
							{
								position472, tokenIndex472, depth472 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l473
								}
								position++
								goto l472
							l473:
								position, tokenIndex, depth = position472, tokenIndex472, depth472
								if buffer[position] != rune('S') {
									goto l447
								}
								position++
							}
						l472:
							{
								position474, tokenIndex474, depth474 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l475
								}
								position++
								goto l474
							l475:
								position, tokenIndex, depth = position474, tokenIndex474, depth474
								if buffer[position] != rune('O') {
									goto l447
								}
								position++
							}
						l474:
							{
								position476, tokenIndex476, depth476 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l477
								}
								position++
								goto l476
							l477:
								position, tokenIndex, depth = position476, tokenIndex476, depth476
								if buffer[position] != rune('L') {
									goto l447
								}
								position++
							}
						l476:
							{
								position478, tokenIndex478, depth478 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l479
								}
								position++
								goto l478
							l479:
								position, tokenIndex, depth = position478, tokenIndex478, depth478
								if buffer[position] != rune('U') {
									goto l447
								}
								position++
							}
						l478:
							{
								position480, tokenIndex480, depth480 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l481
								}
								position++
								goto l480
							l481:
								position, tokenIndex, depth = position480, tokenIndex480, depth480
								if buffer[position] != rune('T') {
									goto l447
								}
								position++
							}
						l480:
							{
								position482, tokenIndex482, depth482 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l483
								}
								position++
								goto l482
							l483:
								position, tokenIndex, depth = position482, tokenIndex482, depth482
								if buffer[position] != rune('I') {
									goto l447
								}
								position++
							}
						l482:
							{
								position484, tokenIndex484, depth484 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l485
								}
								position++
								goto l484
							l485:
								position, tokenIndex, depth = position484, tokenIndex484, depth484
								if buffer[position] != rune('O') {
									goto l447
								}
								position++
							}
						l484:
							{
								position486, tokenIndex486, depth486 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l487
								}
								position++
								goto l486
							l487:
								position, tokenIndex, depth = position486, tokenIndex486, depth486
								if buffer[position] != rune('N') {
									goto l447
								}
								position++
							}
						l486:
							depth--
							add(rulePegText, position467)
						}
						break
					case 'T', 't':
						{
							position488 := position
							depth++
							{
								position489, tokenIndex489, depth489 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l490
								}
								position++
								goto l489
							l490:
								position, tokenIndex, depth = position489, tokenIndex489, depth489
								if buffer[position] != rune('T') {
									goto l447
								}
								position++
							}
						l489:
							{
								position491, tokenIndex491, depth491 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l492
								}
								position++
								goto l491
							l492:
								position, tokenIndex, depth = position491, tokenIndex491, depth491
								if buffer[position] != rune('O') {
									goto l447
								}
								position++
							}
						l491:
							depth--
							add(rulePegText, position488)
						}
						break
					default:
						{
							position493 := position
							depth++
							{
								position494, tokenIndex494, depth494 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l495
								}
								position++
								goto l494
							l495:
								position, tokenIndex, depth = position494, tokenIndex494, depth494
								if buffer[position] != rune('F') {
									goto l447
								}
								position++
							}
						l494:
							{
								position496, tokenIndex496, depth496 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l497
								}
								position++
								goto l496
							l497:
								position, tokenIndex, depth = position496, tokenIndex496, depth496
								if buffer[position] != rune('R') {
									goto l447
								}
								position++
							}
						l496:
							{
								position498, tokenIndex498, depth498 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l499
								}
								position++
								goto l498
							l499:
								position, tokenIndex, depth = position498, tokenIndex498, depth498
								if buffer[position] != rune('O') {
									goto l447
								}
								position++
							}
						l498:
							{
								position500, tokenIndex500, depth500 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l501
								}
								position++
								goto l500
							l501:
								position, tokenIndex, depth = position500, tokenIndex500, depth500
								if buffer[position] != rune('M') {
									goto l447
								}
								position++
							}
						l500:
							depth--
							add(rulePegText, position493)
						}
						break
					}
				}

				if !_rules[ruleKEY]() {
					goto l447
				}
				depth--
				add(rulePROPERTY_KEY, position448)
			}
			return true
		l447:
			position, tokenIndex, depth = position447, tokenIndex447, depth447
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
			position512, tokenIndex512, depth512 := position, tokenIndex, depth
			{
				position513 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l512
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position513)
			}
			return true
		l512:
			position, tokenIndex, depth = position512, tokenIndex512, depth512
			return false
		},
		/* 46 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position514, tokenIndex514, depth514 := position, tokenIndex, depth
			{
				position515 := position
				depth++
				if buffer[position] != rune('"') {
					goto l514
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position515)
			}
			return true
		l514:
			position, tokenIndex, depth = position514, tokenIndex514, depth514
			return false
		},
		/* 47 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> QUOTE_SINGLE) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> QUOTE_DOUBLE))> */
		func() bool {
			position516, tokenIndex516, depth516 := position, tokenIndex, depth
			{
				position517 := position
				depth++
				{
					position518, tokenIndex518, depth518 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l519
					}
					{
						position520 := position
						depth++
					l521:
						{
							position522, tokenIndex522, depth522 := position, tokenIndex, depth
							{
								position523, tokenIndex523, depth523 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l523
								}
								goto l522
							l523:
								position, tokenIndex, depth = position523, tokenIndex523, depth523
							}
							if !_rules[ruleCHAR]() {
								goto l522
							}
							goto l521
						l522:
							position, tokenIndex, depth = position522, tokenIndex522, depth522
						}
						depth--
						add(rulePegText, position520)
					}
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l519
					}
					goto l518
				l519:
					position, tokenIndex, depth = position518, tokenIndex518, depth518
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l516
					}
					{
						position524 := position
						depth++
					l525:
						{
							position526, tokenIndex526, depth526 := position, tokenIndex, depth
							{
								position527, tokenIndex527, depth527 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l527
								}
								goto l526
							l527:
								position, tokenIndex, depth = position527, tokenIndex527, depth527
							}
							if !_rules[ruleCHAR]() {
								goto l526
							}
							goto l525
						l526:
							position, tokenIndex, depth = position526, tokenIndex526, depth526
						}
						depth--
						add(rulePegText, position524)
					}
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l516
					}
				}
			l518:
				depth--
				add(ruleSTRING, position517)
			}
			return true
		l516:
			position, tokenIndex, depth = position516, tokenIndex516, depth516
			return false
		},
		/* 48 CHAR <- <(('\\' ((&('"') QUOTE_DOUBLE) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position528, tokenIndex528, depth528 := position, tokenIndex, depth
			{
				position529 := position
				depth++
				{
					position530, tokenIndex530, depth530 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l531
					}
					position++
					{
						switch buffer[position] {
						case '"':
							if !_rules[ruleQUOTE_DOUBLE]() {
								goto l531
							}
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l531
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l531
							}
							break
						}
					}

					goto l530
				l531:
					position, tokenIndex, depth = position530, tokenIndex530, depth530
					{
						position533, tokenIndex533, depth533 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l533
						}
						goto l528
					l533:
						position, tokenIndex, depth = position533, tokenIndex533, depth533
					}
					if !matchDot() {
						goto l528
					}
				}
			l530:
				depth--
				add(ruleCHAR, position529)
			}
			return true
		l528:
			position, tokenIndex, depth = position528, tokenIndex528, depth528
			return false
		},
		/* 49 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position534, tokenIndex534, depth534 := position, tokenIndex, depth
			{
				position535 := position
				depth++
				{
					position536, tokenIndex536, depth536 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l537
					}
					position++
					goto l536
				l537:
					position, tokenIndex, depth = position536, tokenIndex536, depth536
					if buffer[position] != rune('\\') {
						goto l534
					}
					position++
				}
			l536:
				depth--
				add(ruleESCAPE_CLASS, position535)
			}
			return true
		l534:
			position, tokenIndex, depth = position534, tokenIndex534, depth534
			return false
		},
		/* 50 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position538, tokenIndex538, depth538 := position, tokenIndex, depth
			{
				position539 := position
				depth++
				{
					position540 := position
					depth++
					{
						position541, tokenIndex541, depth541 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l541
						}
						position++
						goto l542
					l541:
						position, tokenIndex, depth = position541, tokenIndex541, depth541
					}
				l542:
					{
						position543 := position
						depth++
						{
							position544, tokenIndex544, depth544 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l545
							}
							position++
							goto l544
						l545:
							position, tokenIndex, depth = position544, tokenIndex544, depth544
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l538
							}
							position++
						l546:
							{
								position547, tokenIndex547, depth547 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l547
								}
								position++
								goto l546
							l547:
								position, tokenIndex, depth = position547, tokenIndex547, depth547
							}
						}
					l544:
						depth--
						add(ruleNUMBER_NATURAL, position543)
					}
					depth--
					add(ruleNUMBER_INTEGER, position540)
				}
				{
					position548, tokenIndex548, depth548 := position, tokenIndex, depth
					{
						position550 := position
						depth++
						if buffer[position] != rune('.') {
							goto l548
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l548
						}
						position++
					l551:
						{
							position552, tokenIndex552, depth552 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l552
							}
							position++
							goto l551
						l552:
							position, tokenIndex, depth = position552, tokenIndex552, depth552
						}
						depth--
						add(ruleNUMBER_FRACTION, position550)
					}
					goto l549
				l548:
					position, tokenIndex, depth = position548, tokenIndex548, depth548
				}
			l549:
				{
					position553, tokenIndex553, depth553 := position, tokenIndex, depth
					{
						position555 := position
						depth++
						{
							position556, tokenIndex556, depth556 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l557
							}
							position++
							goto l556
						l557:
							position, tokenIndex, depth = position556, tokenIndex556, depth556
							if buffer[position] != rune('E') {
								goto l553
							}
							position++
						}
					l556:
						{
							position558, tokenIndex558, depth558 := position, tokenIndex, depth
							{
								position560, tokenIndex560, depth560 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l561
								}
								position++
								goto l560
							l561:
								position, tokenIndex, depth = position560, tokenIndex560, depth560
								if buffer[position] != rune('-') {
									goto l558
								}
								position++
							}
						l560:
							goto l559
						l558:
							position, tokenIndex, depth = position558, tokenIndex558, depth558
						}
					l559:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l553
						}
						position++
					l562:
						{
							position563, tokenIndex563, depth563 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l563
							}
							position++
							goto l562
						l563:
							position, tokenIndex, depth = position563, tokenIndex563, depth563
						}
						depth--
						add(ruleNUMBER_EXP, position555)
					}
					goto l554
				l553:
					position, tokenIndex, depth = position553, tokenIndex553, depth553
				}
			l554:
				depth--
				add(ruleNUMBER, position539)
			}
			return true
		l538:
			position, tokenIndex, depth = position538, tokenIndex538, depth538
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
			position568, tokenIndex568, depth568 := position, tokenIndex, depth
			{
				position569 := position
				depth++
				if buffer[position] != rune('(') {
					goto l568
				}
				position++
				depth--
				add(rulePAREN_OPEN, position569)
			}
			return true
		l568:
			position, tokenIndex, depth = position568, tokenIndex568, depth568
			return false
		},
		/* 56 PAREN_CLOSE <- <')'> */
		func() bool {
			position570, tokenIndex570, depth570 := position, tokenIndex, depth
			{
				position571 := position
				depth++
				if buffer[position] != rune(')') {
					goto l570
				}
				position++
				depth--
				add(rulePAREN_CLOSE, position571)
			}
			return true
		l570:
			position, tokenIndex, depth = position570, tokenIndex570, depth570
			return false
		},
		/* 57 COMMA <- <','> */
		func() bool {
			position572, tokenIndex572, depth572 := position, tokenIndex, depth
			{
				position573 := position
				depth++
				if buffer[position] != rune(',') {
					goto l572
				}
				position++
				depth--
				add(ruleCOMMA, position573)
			}
			return true
		l572:
			position, tokenIndex, depth = position572, tokenIndex572, depth572
			return false
		},
		/* 58 _ <- <SPACE*> */
		func() bool {
			{
				position575 := position
				depth++
			l576:
				{
					position577, tokenIndex577, depth577 := position, tokenIndex, depth
					{
						position578 := position
						depth++
						{
							switch buffer[position] {
							case '\t':
								if buffer[position] != rune('\t') {
									goto l577
								}
								position++
								break
							case '\n':
								if buffer[position] != rune('\n') {
									goto l577
								}
								position++
								break
							default:
								if buffer[position] != rune(' ') {
									goto l577
								}
								position++
								break
							}
						}

						depth--
						add(ruleSPACE, position578)
					}
					goto l576
				l577:
					position, tokenIndex, depth = position577, tokenIndex577, depth577
				}
				depth--
				add(rule_, position575)
			}
			return true
		},
		/* 59 KEY <- <!ID_CONT> */
		func() bool {
			position580, tokenIndex580, depth580 := position, tokenIndex, depth
			{
				position581 := position
				depth++
				{
					position582, tokenIndex582, depth582 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l582
					}
					goto l580
				l582:
					position, tokenIndex, depth = position582, tokenIndex582, depth582
				}
				depth--
				add(ruleKEY, position581)
			}
			return true
		l580:
			position, tokenIndex, depth = position580, tokenIndex580, depth580
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
		/* 84 Action21 <- <{ p.addExpressionList() }> */
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
