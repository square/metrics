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
		/* 18 predicate_1 <- <((predicate_2 _ OP_OR predicate_1 Action35) / predicate_2 / )> */
		func() bool {
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
						goto l236
					}
					goto l228
				l236:
					position, tokenIndex, depth = position228, tokenIndex228, depth228
				}
			l228:
				depth--
				add(rulepredicate_1, position227)
			}
			return true
		},
		/* 19 predicate_2 <- <((predicate_3 _ OP_AND predicate_2 Action36) / predicate_3)> */
		func() bool {
			position237, tokenIndex237, depth237 := position, tokenIndex, depth
			{
				position238 := position
				depth++
				{
					position239, tokenIndex239, depth239 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l240
					}
					if !_rules[rule_]() {
						goto l240
					}
					{
						position241 := position
						depth++
						{
							position242, tokenIndex242, depth242 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l243
							}
							position++
							goto l242
						l243:
							position, tokenIndex, depth = position242, tokenIndex242, depth242
							if buffer[position] != rune('A') {
								goto l240
							}
							position++
						}
					l242:
						{
							position244, tokenIndex244, depth244 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l245
							}
							position++
							goto l244
						l245:
							position, tokenIndex, depth = position244, tokenIndex244, depth244
							if buffer[position] != rune('N') {
								goto l240
							}
							position++
						}
					l244:
						{
							position246, tokenIndex246, depth246 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l247
							}
							position++
							goto l246
						l247:
							position, tokenIndex, depth = position246, tokenIndex246, depth246
							if buffer[position] != rune('D') {
								goto l240
							}
							position++
						}
					l246:
						if !_rules[ruleKEY]() {
							goto l240
						}
						depth--
						add(ruleOP_AND, position241)
					}
					if !_rules[rulepredicate_2]() {
						goto l240
					}
					{
						add(ruleAction36, position)
					}
					goto l239
				l240:
					position, tokenIndex, depth = position239, tokenIndex239, depth239
					if !_rules[rulepredicate_3]() {
						goto l237
					}
				}
			l239:
				depth--
				add(rulepredicate_2, position238)
			}
			return true
		l237:
			position, tokenIndex, depth = position237, tokenIndex237, depth237
			return false
		},
		/* 20 predicate_3 <- <((_ OP_NOT predicate_3 Action37) / (_ PAREN_OPEN predicate_1 _ PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position249, tokenIndex249, depth249 := position, tokenIndex, depth
			{
				position250 := position
				depth++
				{
					position251, tokenIndex251, depth251 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l252
					}
					{
						position253 := position
						depth++
						{
							position254, tokenIndex254, depth254 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l255
							}
							position++
							goto l254
						l255:
							position, tokenIndex, depth = position254, tokenIndex254, depth254
							if buffer[position] != rune('N') {
								goto l252
							}
							position++
						}
					l254:
						{
							position256, tokenIndex256, depth256 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l257
							}
							position++
							goto l256
						l257:
							position, tokenIndex, depth = position256, tokenIndex256, depth256
							if buffer[position] != rune('O') {
								goto l252
							}
							position++
						}
					l256:
						{
							position258, tokenIndex258, depth258 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l259
							}
							position++
							goto l258
						l259:
							position, tokenIndex, depth = position258, tokenIndex258, depth258
							if buffer[position] != rune('T') {
								goto l252
							}
							position++
						}
					l258:
						if !_rules[ruleKEY]() {
							goto l252
						}
						depth--
						add(ruleOP_NOT, position253)
					}
					if !_rules[rulepredicate_3]() {
						goto l252
					}
					{
						add(ruleAction37, position)
					}
					goto l251
				l252:
					position, tokenIndex, depth = position251, tokenIndex251, depth251
					if !_rules[rule_]() {
						goto l261
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l261
					}
					if !_rules[rulepredicate_1]() {
						goto l261
					}
					if !_rules[rule_]() {
						goto l261
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l261
					}
					goto l251
				l261:
					position, tokenIndex, depth = position251, tokenIndex251, depth251
					{
						position262 := position
						depth++
						{
							position263, tokenIndex263, depth263 := position, tokenIndex, depth
							if !_rules[ruletagName]() {
								goto l264
							}
							if !_rules[rule_]() {
								goto l264
							}
							if buffer[position] != rune('=') {
								goto l264
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l264
							}
							{
								add(ruleAction38, position)
							}
							goto l263
						l264:
							position, tokenIndex, depth = position263, tokenIndex263, depth263
							if !_rules[ruletagName]() {
								goto l266
							}
							if !_rules[rule_]() {
								goto l266
							}
							if buffer[position] != rune('!') {
								goto l266
							}
							position++
							if buffer[position] != rune('=') {
								goto l266
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l266
							}
							{
								add(ruleAction39, position)
							}
							goto l263
						l266:
							position, tokenIndex, depth = position263, tokenIndex263, depth263
							if !_rules[ruletagName]() {
								goto l268
							}
							if !_rules[rule_]() {
								goto l268
							}
							{
								position269, tokenIndex269, depth269 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l270
								}
								position++
								goto l269
							l270:
								position, tokenIndex, depth = position269, tokenIndex269, depth269
								if buffer[position] != rune('M') {
									goto l268
								}
								position++
							}
						l269:
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
									goto l268
								}
								position++
							}
						l271:
							{
								position273, tokenIndex273, depth273 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l274
								}
								position++
								goto l273
							l274:
								position, tokenIndex, depth = position273, tokenIndex273, depth273
								if buffer[position] != rune('T') {
									goto l268
								}
								position++
							}
						l273:
							{
								position275, tokenIndex275, depth275 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l276
								}
								position++
								goto l275
							l276:
								position, tokenIndex, depth = position275, tokenIndex275, depth275
								if buffer[position] != rune('C') {
									goto l268
								}
								position++
							}
						l275:
							{
								position277, tokenIndex277, depth277 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l278
								}
								position++
								goto l277
							l278:
								position, tokenIndex, depth = position277, tokenIndex277, depth277
								if buffer[position] != rune('H') {
									goto l268
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
									goto l268
								}
								position++
							}
						l279:
							{
								position281, tokenIndex281, depth281 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l282
								}
								position++
								goto l281
							l282:
								position, tokenIndex, depth = position281, tokenIndex281, depth281
								if buffer[position] != rune('S') {
									goto l268
								}
								position++
							}
						l281:
							if !_rules[ruleKEY]() {
								goto l268
							}
							if !_rules[ruleliteralString]() {
								goto l268
							}
							{
								add(ruleAction40, position)
							}
							goto l263
						l268:
							position, tokenIndex, depth = position263, tokenIndex263, depth263
							if !_rules[ruletagName]() {
								goto l249
							}
							if !_rules[rule_]() {
								goto l249
							}
							{
								position284, tokenIndex284, depth284 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l285
								}
								position++
								goto l284
							l285:
								position, tokenIndex, depth = position284, tokenIndex284, depth284
								if buffer[position] != rune('I') {
									goto l249
								}
								position++
							}
						l284:
							{
								position286, tokenIndex286, depth286 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l287
								}
								position++
								goto l286
							l287:
								position, tokenIndex, depth = position286, tokenIndex286, depth286
								if buffer[position] != rune('N') {
									goto l249
								}
								position++
							}
						l286:
							if !_rules[ruleKEY]() {
								goto l249
							}
							{
								position288 := position
								depth++
								{
									add(ruleAction43, position)
								}
								if !_rules[rule_]() {
									goto l249
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l249
								}
								if !_rules[ruleliteralListString]() {
									goto l249
								}
							l290:
								{
									position291, tokenIndex291, depth291 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l291
									}
									if !_rules[ruleCOMMA]() {
										goto l291
									}
									if !_rules[ruleliteralListString]() {
										goto l291
									}
									goto l290
								l291:
									position, tokenIndex, depth = position291, tokenIndex291, depth291
								}
								if !_rules[rule_]() {
									goto l249
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l249
								}
								depth--
								add(ruleliteralList, position288)
							}
							{
								add(ruleAction41, position)
							}
						}
					l263:
						depth--
						add(ruletagMatcher, position262)
					}
				}
			l251:
				depth--
				add(rulepredicate_3, position250)
			}
			return true
		l249:
			position, tokenIndex, depth = position249, tokenIndex249, depth249
			return false
		},
		/* 21 tagMatcher <- <((tagName _ '=' literalString Action38) / (tagName _ ('!' '=') literalString Action39) / (tagName _ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) KEY literalString Action40) / (tagName _ (('i' / 'I') ('n' / 'N')) KEY literalList Action41))> */
		nil,
		/* 22 literalString <- <(_ STRING Action42)> */
		func() bool {
			position294, tokenIndex294, depth294 := position, tokenIndex, depth
			{
				position295 := position
				depth++
				if !_rules[rule_]() {
					goto l294
				}
				if !_rules[ruleSTRING]() {
					goto l294
				}
				{
					add(ruleAction42, position)
				}
				depth--
				add(ruleliteralString, position295)
			}
			return true
		l294:
			position, tokenIndex, depth = position294, tokenIndex294, depth294
			return false
		},
		/* 23 literalList <- <(Action43 _ PAREN_OPEN literalListString (_ COMMA literalListString)* _ PAREN_CLOSE)> */
		nil,
		/* 24 literalListString <- <(_ STRING Action44)> */
		func() bool {
			position298, tokenIndex298, depth298 := position, tokenIndex, depth
			{
				position299 := position
				depth++
				if !_rules[rule_]() {
					goto l298
				}
				if !_rules[ruleSTRING]() {
					goto l298
				}
				{
					add(ruleAction44, position)
				}
				depth--
				add(ruleliteralListString, position299)
			}
			return true
		l298:
			position, tokenIndex, depth = position298, tokenIndex298, depth298
			return false
		},
		/* 25 tagName <- <(_ <TAG_NAME> Action45)> */
		func() bool {
			position301, tokenIndex301, depth301 := position, tokenIndex, depth
			{
				position302 := position
				depth++
				if !_rules[rule_]() {
					goto l301
				}
				{
					position303 := position
					depth++
					{
						position304 := position
						depth++
						if !_rules[rule_]() {
							goto l301
						}
						if !_rules[ruleIDENTIFIER]() {
							goto l301
						}
						depth--
						add(ruleTAG_NAME, position304)
					}
					depth--
					add(rulePegText, position303)
				}
				{
					add(ruleAction45, position)
				}
				depth--
				add(ruletagName, position302)
			}
			return true
		l301:
			position, tokenIndex, depth = position301, tokenIndex301, depth301
			return false
		},
		/* 26 COLUMN_NAME <- <(_ IDENTIFIER)> */
		func() bool {
			position306, tokenIndex306, depth306 := position, tokenIndex, depth
			{
				position307 := position
				depth++
				if !_rules[rule_]() {
					goto l306
				}
				if !_rules[ruleIDENTIFIER]() {
					goto l306
				}
				depth--
				add(ruleCOLUMN_NAME, position307)
			}
			return true
		l306:
			position, tokenIndex, depth = position306, tokenIndex306, depth306
			return false
		},
		/* 27 METRIC_NAME <- <(_ IDENTIFIER)> */
		nil,
		/* 28 TAG_NAME <- <(_ IDENTIFIER)> */
		nil,
		/* 29 IDENTIFIER <- <((_ '`' CHAR* '`') / (_ !(KEYWORD KEY) ID_SEGMENT ('.' ID_SEGMENT)*))> */
		func() bool {
			position310, tokenIndex310, depth310 := position, tokenIndex, depth
			{
				position311 := position
				depth++
				{
					position312, tokenIndex312, depth312 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l313
					}
					if buffer[position] != rune('`') {
						goto l313
					}
					position++
				l314:
					{
						position315, tokenIndex315, depth315 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l315
						}
						goto l314
					l315:
						position, tokenIndex, depth = position315, tokenIndex315, depth315
					}
					if buffer[position] != rune('`') {
						goto l313
					}
					position++
					goto l312
				l313:
					position, tokenIndex, depth = position312, tokenIndex312, depth312
					if !_rules[rule_]() {
						goto l310
					}
					{
						position316, tokenIndex316, depth316 := position, tokenIndex, depth
						{
							position317 := position
							depth++
							{
								position318, tokenIndex318, depth318 := position, tokenIndex, depth
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
										goto l319
									}
									position++
								}
							l320:
								{
									position322, tokenIndex322, depth322 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l323
									}
									position++
									goto l322
								l323:
									position, tokenIndex, depth = position322, tokenIndex322, depth322
									if buffer[position] != rune('L') {
										goto l319
									}
									position++
								}
							l322:
								{
									position324, tokenIndex324, depth324 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l325
									}
									position++
									goto l324
								l325:
									position, tokenIndex, depth = position324, tokenIndex324, depth324
									if buffer[position] != rune('L') {
										goto l319
									}
									position++
								}
							l324:
								goto l318
							l319:
								position, tokenIndex, depth = position318, tokenIndex318, depth318
								{
									position327, tokenIndex327, depth327 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l328
									}
									position++
									goto l327
								l328:
									position, tokenIndex, depth = position327, tokenIndex327, depth327
									if buffer[position] != rune('A') {
										goto l326
									}
									position++
								}
							l327:
								{
									position329, tokenIndex329, depth329 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l330
									}
									position++
									goto l329
								l330:
									position, tokenIndex, depth = position329, tokenIndex329, depth329
									if buffer[position] != rune('N') {
										goto l326
									}
									position++
								}
							l329:
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
										goto l326
									}
									position++
								}
							l331:
								goto l318
							l326:
								position, tokenIndex, depth = position318, tokenIndex318, depth318
								{
									position334, tokenIndex334, depth334 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l335
									}
									position++
									goto l334
								l335:
									position, tokenIndex, depth = position334, tokenIndex334, depth334
									if buffer[position] != rune('M') {
										goto l333
									}
									position++
								}
							l334:
								{
									position336, tokenIndex336, depth336 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l337
									}
									position++
									goto l336
								l337:
									position, tokenIndex, depth = position336, tokenIndex336, depth336
									if buffer[position] != rune('A') {
										goto l333
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
										goto l333
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
										goto l333
									}
									position++
								}
							l340:
								{
									position342, tokenIndex342, depth342 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l343
									}
									position++
									goto l342
								l343:
									position, tokenIndex, depth = position342, tokenIndex342, depth342
									if buffer[position] != rune('H') {
										goto l333
									}
									position++
								}
							l342:
								{
									position344, tokenIndex344, depth344 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l345
									}
									position++
									goto l344
								l345:
									position, tokenIndex, depth = position344, tokenIndex344, depth344
									if buffer[position] != rune('E') {
										goto l333
									}
									position++
								}
							l344:
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
										goto l333
									}
									position++
								}
							l346:
								goto l318
							l333:
								position, tokenIndex, depth = position318, tokenIndex318, depth318
								{
									position349, tokenIndex349, depth349 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l350
									}
									position++
									goto l349
								l350:
									position, tokenIndex, depth = position349, tokenIndex349, depth349
									if buffer[position] != rune('S') {
										goto l348
									}
									position++
								}
							l349:
								{
									position351, tokenIndex351, depth351 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l352
									}
									position++
									goto l351
								l352:
									position, tokenIndex, depth = position351, tokenIndex351, depth351
									if buffer[position] != rune('E') {
										goto l348
									}
									position++
								}
							l351:
								{
									position353, tokenIndex353, depth353 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l354
									}
									position++
									goto l353
								l354:
									position, tokenIndex, depth = position353, tokenIndex353, depth353
									if buffer[position] != rune('L') {
										goto l348
									}
									position++
								}
							l353:
								{
									position355, tokenIndex355, depth355 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l356
									}
									position++
									goto l355
								l356:
									position, tokenIndex, depth = position355, tokenIndex355, depth355
									if buffer[position] != rune('E') {
										goto l348
									}
									position++
								}
							l355:
								{
									position357, tokenIndex357, depth357 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l358
									}
									position++
									goto l357
								l358:
									position, tokenIndex, depth = position357, tokenIndex357, depth357
									if buffer[position] != rune('C') {
										goto l348
									}
									position++
								}
							l357:
								{
									position359, tokenIndex359, depth359 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l360
									}
									position++
									goto l359
								l360:
									position, tokenIndex, depth = position359, tokenIndex359, depth359
									if buffer[position] != rune('T') {
										goto l348
									}
									position++
								}
							l359:
								goto l318
							l348:
								position, tokenIndex, depth = position318, tokenIndex318, depth318
								{
									switch buffer[position] {
									case 'M', 'm':
										{
											position362, tokenIndex362, depth362 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l363
											}
											position++
											goto l362
										l363:
											position, tokenIndex, depth = position362, tokenIndex362, depth362
											if buffer[position] != rune('M') {
												goto l316
											}
											position++
										}
									l362:
										{
											position364, tokenIndex364, depth364 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l365
											}
											position++
											goto l364
										l365:
											position, tokenIndex, depth = position364, tokenIndex364, depth364
											if buffer[position] != rune('E') {
												goto l316
											}
											position++
										}
									l364:
										{
											position366, tokenIndex366, depth366 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l367
											}
											position++
											goto l366
										l367:
											position, tokenIndex, depth = position366, tokenIndex366, depth366
											if buffer[position] != rune('T') {
												goto l316
											}
											position++
										}
									l366:
										{
											position368, tokenIndex368, depth368 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l369
											}
											position++
											goto l368
										l369:
											position, tokenIndex, depth = position368, tokenIndex368, depth368
											if buffer[position] != rune('R') {
												goto l316
											}
											position++
										}
									l368:
										{
											position370, tokenIndex370, depth370 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l371
											}
											position++
											goto l370
										l371:
											position, tokenIndex, depth = position370, tokenIndex370, depth370
											if buffer[position] != rune('I') {
												goto l316
											}
											position++
										}
									l370:
										{
											position372, tokenIndex372, depth372 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l373
											}
											position++
											goto l372
										l373:
											position, tokenIndex, depth = position372, tokenIndex372, depth372
											if buffer[position] != rune('C') {
												goto l316
											}
											position++
										}
									l372:
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
												goto l316
											}
											position++
										}
									l374:
										break
									case 'W', 'w':
										{
											position376, tokenIndex376, depth376 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l377
											}
											position++
											goto l376
										l377:
											position, tokenIndex, depth = position376, tokenIndex376, depth376
											if buffer[position] != rune('W') {
												goto l316
											}
											position++
										}
									l376:
										{
											position378, tokenIndex378, depth378 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l379
											}
											position++
											goto l378
										l379:
											position, tokenIndex, depth = position378, tokenIndex378, depth378
											if buffer[position] != rune('H') {
												goto l316
											}
											position++
										}
									l378:
										{
											position380, tokenIndex380, depth380 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l381
											}
											position++
											goto l380
										l381:
											position, tokenIndex, depth = position380, tokenIndex380, depth380
											if buffer[position] != rune('E') {
												goto l316
											}
											position++
										}
									l380:
										{
											position382, tokenIndex382, depth382 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l383
											}
											position++
											goto l382
										l383:
											position, tokenIndex, depth = position382, tokenIndex382, depth382
											if buffer[position] != rune('R') {
												goto l316
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
												goto l316
											}
											position++
										}
									l384:
										break
									case 'O', 'o':
										{
											position386, tokenIndex386, depth386 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l387
											}
											position++
											goto l386
										l387:
											position, tokenIndex, depth = position386, tokenIndex386, depth386
											if buffer[position] != rune('O') {
												goto l316
											}
											position++
										}
									l386:
										{
											position388, tokenIndex388, depth388 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l389
											}
											position++
											goto l388
										l389:
											position, tokenIndex, depth = position388, tokenIndex388, depth388
											if buffer[position] != rune('R') {
												goto l316
											}
											position++
										}
									l388:
										break
									case 'N', 'n':
										{
											position390, tokenIndex390, depth390 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l391
											}
											position++
											goto l390
										l391:
											position, tokenIndex, depth = position390, tokenIndex390, depth390
											if buffer[position] != rune('N') {
												goto l316
											}
											position++
										}
									l390:
										{
											position392, tokenIndex392, depth392 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l393
											}
											position++
											goto l392
										l393:
											position, tokenIndex, depth = position392, tokenIndex392, depth392
											if buffer[position] != rune('O') {
												goto l316
											}
											position++
										}
									l392:
										{
											position394, tokenIndex394, depth394 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l395
											}
											position++
											goto l394
										l395:
											position, tokenIndex, depth = position394, tokenIndex394, depth394
											if buffer[position] != rune('T') {
												goto l316
											}
											position++
										}
									l394:
										break
									case 'I', 'i':
										{
											position396, tokenIndex396, depth396 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l397
											}
											position++
											goto l396
										l397:
											position, tokenIndex, depth = position396, tokenIndex396, depth396
											if buffer[position] != rune('I') {
												goto l316
											}
											position++
										}
									l396:
										{
											position398, tokenIndex398, depth398 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l399
											}
											position++
											goto l398
										l399:
											position, tokenIndex, depth = position398, tokenIndex398, depth398
											if buffer[position] != rune('N') {
												goto l316
											}
											position++
										}
									l398:
										break
									case 'G', 'g':
										{
											position400, tokenIndex400, depth400 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l401
											}
											position++
											goto l400
										l401:
											position, tokenIndex, depth = position400, tokenIndex400, depth400
											if buffer[position] != rune('G') {
												goto l316
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
												goto l316
											}
											position++
										}
									l402:
										{
											position404, tokenIndex404, depth404 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l405
											}
											position++
											goto l404
										l405:
											position, tokenIndex, depth = position404, tokenIndex404, depth404
											if buffer[position] != rune('O') {
												goto l316
											}
											position++
										}
									l404:
										{
											position406, tokenIndex406, depth406 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l407
											}
											position++
											goto l406
										l407:
											position, tokenIndex, depth = position406, tokenIndex406, depth406
											if buffer[position] != rune('U') {
												goto l316
											}
											position++
										}
									l406:
										{
											position408, tokenIndex408, depth408 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l409
											}
											position++
											goto l408
										l409:
											position, tokenIndex, depth = position408, tokenIndex408, depth408
											if buffer[position] != rune('P') {
												goto l316
											}
											position++
										}
									l408:
										break
									case 'D', 'd':
										{
											position410, tokenIndex410, depth410 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l411
											}
											position++
											goto l410
										l411:
											position, tokenIndex, depth = position410, tokenIndex410, depth410
											if buffer[position] != rune('D') {
												goto l316
											}
											position++
										}
									l410:
										{
											position412, tokenIndex412, depth412 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l413
											}
											position++
											goto l412
										l413:
											position, tokenIndex, depth = position412, tokenIndex412, depth412
											if buffer[position] != rune('E') {
												goto l316
											}
											position++
										}
									l412:
										{
											position414, tokenIndex414, depth414 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l415
											}
											position++
											goto l414
										l415:
											position, tokenIndex, depth = position414, tokenIndex414, depth414
											if buffer[position] != rune('S') {
												goto l316
											}
											position++
										}
									l414:
										{
											position416, tokenIndex416, depth416 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l417
											}
											position++
											goto l416
										l417:
											position, tokenIndex, depth = position416, tokenIndex416, depth416
											if buffer[position] != rune('C') {
												goto l316
											}
											position++
										}
									l416:
										{
											position418, tokenIndex418, depth418 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l419
											}
											position++
											goto l418
										l419:
											position, tokenIndex, depth = position418, tokenIndex418, depth418
											if buffer[position] != rune('R') {
												goto l316
											}
											position++
										}
									l418:
										{
											position420, tokenIndex420, depth420 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l421
											}
											position++
											goto l420
										l421:
											position, tokenIndex, depth = position420, tokenIndex420, depth420
											if buffer[position] != rune('I') {
												goto l316
											}
											position++
										}
									l420:
										{
											position422, tokenIndex422, depth422 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l423
											}
											position++
											goto l422
										l423:
											position, tokenIndex, depth = position422, tokenIndex422, depth422
											if buffer[position] != rune('B') {
												goto l316
											}
											position++
										}
									l422:
										{
											position424, tokenIndex424, depth424 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l425
											}
											position++
											goto l424
										l425:
											position, tokenIndex, depth = position424, tokenIndex424, depth424
											if buffer[position] != rune('E') {
												goto l316
											}
											position++
										}
									l424:
										break
									case 'B', 'b':
										{
											position426, tokenIndex426, depth426 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l427
											}
											position++
											goto l426
										l427:
											position, tokenIndex, depth = position426, tokenIndex426, depth426
											if buffer[position] != rune('B') {
												goto l316
											}
											position++
										}
									l426:
										{
											position428, tokenIndex428, depth428 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l429
											}
											position++
											goto l428
										l429:
											position, tokenIndex, depth = position428, tokenIndex428, depth428
											if buffer[position] != rune('Y') {
												goto l316
											}
											position++
										}
									l428:
										break
									case 'A', 'a':
										{
											position430, tokenIndex430, depth430 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l431
											}
											position++
											goto l430
										l431:
											position, tokenIndex, depth = position430, tokenIndex430, depth430
											if buffer[position] != rune('A') {
												goto l316
											}
											position++
										}
									l430:
										{
											position432, tokenIndex432, depth432 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l433
											}
											position++
											goto l432
										l433:
											position, tokenIndex, depth = position432, tokenIndex432, depth432
											if buffer[position] != rune('S') {
												goto l316
											}
											position++
										}
									l432:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l316
										}
										break
									}
								}

							}
						l318:
							depth--
							add(ruleKEYWORD, position317)
						}
						if !_rules[ruleKEY]() {
							goto l316
						}
						goto l310
					l316:
						position, tokenIndex, depth = position316, tokenIndex316, depth316
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l310
					}
				l434:
					{
						position435, tokenIndex435, depth435 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l435
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l435
						}
						goto l434
					l435:
						position, tokenIndex, depth = position435, tokenIndex435, depth435
					}
				}
			l312:
				depth--
				add(ruleIDENTIFIER, position311)
			}
			return true
		l310:
			position, tokenIndex, depth = position310, tokenIndex310, depth310
			return false
		},
		/* 30 TIMESTAMP <- <((_ <(NUMBER ([a-z] / [A-Z])*)>) / (_ STRING) / (_ <(('n' / 'N') ('o' / 'O') ('w' / 'W'))>))> */
		nil,
		/* 31 ID_SEGMENT <- <(_ ID_START ID_CONT*)> */
		func() bool {
			position437, tokenIndex437, depth437 := position, tokenIndex, depth
			{
				position438 := position
				depth++
				if !_rules[rule_]() {
					goto l437
				}
				if !_rules[ruleID_START]() {
					goto l437
				}
			l439:
				{
					position440, tokenIndex440, depth440 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l440
					}
					goto l439
				l440:
					position, tokenIndex, depth = position440, tokenIndex440, depth440
				}
				depth--
				add(ruleID_SEGMENT, position438)
			}
			return true
		l437:
			position, tokenIndex, depth = position437, tokenIndex437, depth437
			return false
		},
		/* 32 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position441, tokenIndex441, depth441 := position, tokenIndex, depth
			{
				position442 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l441
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l441
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l441
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position442)
			}
			return true
		l441:
			position, tokenIndex, depth = position441, tokenIndex441, depth441
			return false
		},
		/* 33 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position444, tokenIndex444, depth444 := position, tokenIndex, depth
			{
				position445 := position
				depth++
				{
					position446, tokenIndex446, depth446 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l447
					}
					goto l446
				l447:
					position, tokenIndex, depth = position446, tokenIndex446, depth446
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l444
					}
					position++
				}
			l446:
				depth--
				add(ruleID_CONT, position445)
			}
			return true
		l444:
			position, tokenIndex, depth = position444, tokenIndex444, depth444
			return false
		},
		/* 34 PROPERTY_KEY <- <(((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> KEY _ (('b' / 'B') ('y' / 'Y')))) | (&('R' | 'r') <(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))>) | (&('T' | 't') <(('t' / 'T') ('o' / 'O'))>) | (&('F' | 'f') <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))>)) KEY)> */
		func() bool {
			position448, tokenIndex448, depth448 := position, tokenIndex, depth
			{
				position449 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position451 := position
							depth++
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
									goto l448
								}
								position++
							}
						l452:
							{
								position454, tokenIndex454, depth454 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l455
								}
								position++
								goto l454
							l455:
								position, tokenIndex, depth = position454, tokenIndex454, depth454
								if buffer[position] != rune('A') {
									goto l448
								}
								position++
							}
						l454:
							{
								position456, tokenIndex456, depth456 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l457
								}
								position++
								goto l456
							l457:
								position, tokenIndex, depth = position456, tokenIndex456, depth456
								if buffer[position] != rune('M') {
									goto l448
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
									goto l448
								}
								position++
							}
						l458:
							{
								position460, tokenIndex460, depth460 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l461
								}
								position++
								goto l460
							l461:
								position, tokenIndex, depth = position460, tokenIndex460, depth460
								if buffer[position] != rune('L') {
									goto l448
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
									goto l448
								}
								position++
							}
						l462:
							depth--
							add(rulePegText, position451)
						}
						if !_rules[ruleKEY]() {
							goto l448
						}
						if !_rules[rule_]() {
							goto l448
						}
						{
							position464, tokenIndex464, depth464 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l465
							}
							position++
							goto l464
						l465:
							position, tokenIndex, depth = position464, tokenIndex464, depth464
							if buffer[position] != rune('B') {
								goto l448
							}
							position++
						}
					l464:
						{
							position466, tokenIndex466, depth466 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l467
							}
							position++
							goto l466
						l467:
							position, tokenIndex, depth = position466, tokenIndex466, depth466
							if buffer[position] != rune('Y') {
								goto l448
							}
							position++
						}
					l466:
						break
					case 'R', 'r':
						{
							position468 := position
							depth++
							{
								position469, tokenIndex469, depth469 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l470
								}
								position++
								goto l469
							l470:
								position, tokenIndex, depth = position469, tokenIndex469, depth469
								if buffer[position] != rune('R') {
									goto l448
								}
								position++
							}
						l469:
							{
								position471, tokenIndex471, depth471 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l472
								}
								position++
								goto l471
							l472:
								position, tokenIndex, depth = position471, tokenIndex471, depth471
								if buffer[position] != rune('E') {
									goto l448
								}
								position++
							}
						l471:
							{
								position473, tokenIndex473, depth473 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l474
								}
								position++
								goto l473
							l474:
								position, tokenIndex, depth = position473, tokenIndex473, depth473
								if buffer[position] != rune('S') {
									goto l448
								}
								position++
							}
						l473:
							{
								position475, tokenIndex475, depth475 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l476
								}
								position++
								goto l475
							l476:
								position, tokenIndex, depth = position475, tokenIndex475, depth475
								if buffer[position] != rune('O') {
									goto l448
								}
								position++
							}
						l475:
							{
								position477, tokenIndex477, depth477 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l478
								}
								position++
								goto l477
							l478:
								position, tokenIndex, depth = position477, tokenIndex477, depth477
								if buffer[position] != rune('L') {
									goto l448
								}
								position++
							}
						l477:
							{
								position479, tokenIndex479, depth479 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l480
								}
								position++
								goto l479
							l480:
								position, tokenIndex, depth = position479, tokenIndex479, depth479
								if buffer[position] != rune('U') {
									goto l448
								}
								position++
							}
						l479:
							{
								position481, tokenIndex481, depth481 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l482
								}
								position++
								goto l481
							l482:
								position, tokenIndex, depth = position481, tokenIndex481, depth481
								if buffer[position] != rune('T') {
									goto l448
								}
								position++
							}
						l481:
							{
								position483, tokenIndex483, depth483 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l484
								}
								position++
								goto l483
							l484:
								position, tokenIndex, depth = position483, tokenIndex483, depth483
								if buffer[position] != rune('I') {
									goto l448
								}
								position++
							}
						l483:
							{
								position485, tokenIndex485, depth485 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l486
								}
								position++
								goto l485
							l486:
								position, tokenIndex, depth = position485, tokenIndex485, depth485
								if buffer[position] != rune('O') {
									goto l448
								}
								position++
							}
						l485:
							{
								position487, tokenIndex487, depth487 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l488
								}
								position++
								goto l487
							l488:
								position, tokenIndex, depth = position487, tokenIndex487, depth487
								if buffer[position] != rune('N') {
									goto l448
								}
								position++
							}
						l487:
							depth--
							add(rulePegText, position468)
						}
						break
					case 'T', 't':
						{
							position489 := position
							depth++
							{
								position490, tokenIndex490, depth490 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l491
								}
								position++
								goto l490
							l491:
								position, tokenIndex, depth = position490, tokenIndex490, depth490
								if buffer[position] != rune('T') {
									goto l448
								}
								position++
							}
						l490:
							{
								position492, tokenIndex492, depth492 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l493
								}
								position++
								goto l492
							l493:
								position, tokenIndex, depth = position492, tokenIndex492, depth492
								if buffer[position] != rune('O') {
									goto l448
								}
								position++
							}
						l492:
							depth--
							add(rulePegText, position489)
						}
						break
					default:
						{
							position494 := position
							depth++
							{
								position495, tokenIndex495, depth495 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l496
								}
								position++
								goto l495
							l496:
								position, tokenIndex, depth = position495, tokenIndex495, depth495
								if buffer[position] != rune('F') {
									goto l448
								}
								position++
							}
						l495:
							{
								position497, tokenIndex497, depth497 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l498
								}
								position++
								goto l497
							l498:
								position, tokenIndex, depth = position497, tokenIndex497, depth497
								if buffer[position] != rune('R') {
									goto l448
								}
								position++
							}
						l497:
							{
								position499, tokenIndex499, depth499 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l500
								}
								position++
								goto l499
							l500:
								position, tokenIndex, depth = position499, tokenIndex499, depth499
								if buffer[position] != rune('O') {
									goto l448
								}
								position++
							}
						l499:
							{
								position501, tokenIndex501, depth501 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l502
								}
								position++
								goto l501
							l502:
								position, tokenIndex, depth = position501, tokenIndex501, depth501
								if buffer[position] != rune('M') {
									goto l448
								}
								position++
							}
						l501:
							depth--
							add(rulePegText, position494)
						}
						break
					}
				}

				if !_rules[ruleKEY]() {
					goto l448
				}
				depth--
				add(rulePROPERTY_KEY, position449)
			}
			return true
		l448:
			position, tokenIndex, depth = position448, tokenIndex448, depth448
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
			position513, tokenIndex513, depth513 := position, tokenIndex, depth
			{
				position514 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l513
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position514)
			}
			return true
		l513:
			position, tokenIndex, depth = position513, tokenIndex513, depth513
			return false
		},
		/* 46 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position515, tokenIndex515, depth515 := position, tokenIndex, depth
			{
				position516 := position
				depth++
				if buffer[position] != rune('"') {
					goto l515
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position516)
			}
			return true
		l515:
			position, tokenIndex, depth = position515, tokenIndex515, depth515
			return false
		},
		/* 47 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> QUOTE_SINGLE) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> QUOTE_DOUBLE))> */
		func() bool {
			position517, tokenIndex517, depth517 := position, tokenIndex, depth
			{
				position518 := position
				depth++
				{
					position519, tokenIndex519, depth519 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l520
					}
					{
						position521 := position
						depth++
					l522:
						{
							position523, tokenIndex523, depth523 := position, tokenIndex, depth
							{
								position524, tokenIndex524, depth524 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l524
								}
								goto l523
							l524:
								position, tokenIndex, depth = position524, tokenIndex524, depth524
							}
							if !_rules[ruleCHAR]() {
								goto l523
							}
							goto l522
						l523:
							position, tokenIndex, depth = position523, tokenIndex523, depth523
						}
						depth--
						add(rulePegText, position521)
					}
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l520
					}
					goto l519
				l520:
					position, tokenIndex, depth = position519, tokenIndex519, depth519
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l517
					}
					{
						position525 := position
						depth++
					l526:
						{
							position527, tokenIndex527, depth527 := position, tokenIndex, depth
							{
								position528, tokenIndex528, depth528 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l528
								}
								goto l527
							l528:
								position, tokenIndex, depth = position528, tokenIndex528, depth528
							}
							if !_rules[ruleCHAR]() {
								goto l527
							}
							goto l526
						l527:
							position, tokenIndex, depth = position527, tokenIndex527, depth527
						}
						depth--
						add(rulePegText, position525)
					}
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l517
					}
				}
			l519:
				depth--
				add(ruleSTRING, position518)
			}
			return true
		l517:
			position, tokenIndex, depth = position517, tokenIndex517, depth517
			return false
		},
		/* 48 CHAR <- <(('\\' ((&('"') QUOTE_DOUBLE) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position529, tokenIndex529, depth529 := position, tokenIndex, depth
			{
				position530 := position
				depth++
				{
					position531, tokenIndex531, depth531 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l532
					}
					position++
					{
						switch buffer[position] {
						case '"':
							if !_rules[ruleQUOTE_DOUBLE]() {
								goto l532
							}
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l532
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l532
							}
							break
						}
					}

					goto l531
				l532:
					position, tokenIndex, depth = position531, tokenIndex531, depth531
					{
						position534, tokenIndex534, depth534 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l534
						}
						goto l529
					l534:
						position, tokenIndex, depth = position534, tokenIndex534, depth534
					}
					if !matchDot() {
						goto l529
					}
				}
			l531:
				depth--
				add(ruleCHAR, position530)
			}
			return true
		l529:
			position, tokenIndex, depth = position529, tokenIndex529, depth529
			return false
		},
		/* 49 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position535, tokenIndex535, depth535 := position, tokenIndex, depth
			{
				position536 := position
				depth++
				{
					position537, tokenIndex537, depth537 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l538
					}
					position++
					goto l537
				l538:
					position, tokenIndex, depth = position537, tokenIndex537, depth537
					if buffer[position] != rune('\\') {
						goto l535
					}
					position++
				}
			l537:
				depth--
				add(ruleESCAPE_CLASS, position536)
			}
			return true
		l535:
			position, tokenIndex, depth = position535, tokenIndex535, depth535
			return false
		},
		/* 50 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position539, tokenIndex539, depth539 := position, tokenIndex, depth
			{
				position540 := position
				depth++
				{
					position541 := position
					depth++
					{
						position542, tokenIndex542, depth542 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l542
						}
						position++
						goto l543
					l542:
						position, tokenIndex, depth = position542, tokenIndex542, depth542
					}
				l543:
					{
						position544 := position
						depth++
						{
							position545, tokenIndex545, depth545 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l546
							}
							position++
							goto l545
						l546:
							position, tokenIndex, depth = position545, tokenIndex545, depth545
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l539
							}
							position++
						l547:
							{
								position548, tokenIndex548, depth548 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l548
								}
								position++
								goto l547
							l548:
								position, tokenIndex, depth = position548, tokenIndex548, depth548
							}
						}
					l545:
						depth--
						add(ruleNUMBER_NATURAL, position544)
					}
					depth--
					add(ruleNUMBER_INTEGER, position541)
				}
				{
					position549, tokenIndex549, depth549 := position, tokenIndex, depth
					{
						position551 := position
						depth++
						if buffer[position] != rune('.') {
							goto l549
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l549
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
						depth--
						add(ruleNUMBER_FRACTION, position551)
					}
					goto l550
				l549:
					position, tokenIndex, depth = position549, tokenIndex549, depth549
				}
			l550:
				{
					position554, tokenIndex554, depth554 := position, tokenIndex, depth
					{
						position556 := position
						depth++
						{
							position557, tokenIndex557, depth557 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l558
							}
							position++
							goto l557
						l558:
							position, tokenIndex, depth = position557, tokenIndex557, depth557
							if buffer[position] != rune('E') {
								goto l554
							}
							position++
						}
					l557:
						{
							position559, tokenIndex559, depth559 := position, tokenIndex, depth
							{
								position561, tokenIndex561, depth561 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l562
								}
								position++
								goto l561
							l562:
								position, tokenIndex, depth = position561, tokenIndex561, depth561
								if buffer[position] != rune('-') {
									goto l559
								}
								position++
							}
						l561:
							goto l560
						l559:
							position, tokenIndex, depth = position559, tokenIndex559, depth559
						}
					l560:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l554
						}
						position++
					l563:
						{
							position564, tokenIndex564, depth564 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l564
							}
							position++
							goto l563
						l564:
							position, tokenIndex, depth = position564, tokenIndex564, depth564
						}
						depth--
						add(ruleNUMBER_EXP, position556)
					}
					goto l555
				l554:
					position, tokenIndex, depth = position554, tokenIndex554, depth554
				}
			l555:
				depth--
				add(ruleNUMBER, position540)
			}
			return true
		l539:
			position, tokenIndex, depth = position539, tokenIndex539, depth539
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
			position569, tokenIndex569, depth569 := position, tokenIndex, depth
			{
				position570 := position
				depth++
				if buffer[position] != rune('(') {
					goto l569
				}
				position++
				depth--
				add(rulePAREN_OPEN, position570)
			}
			return true
		l569:
			position, tokenIndex, depth = position569, tokenIndex569, depth569
			return false
		},
		/* 56 PAREN_CLOSE <- <')'> */
		func() bool {
			position571, tokenIndex571, depth571 := position, tokenIndex, depth
			{
				position572 := position
				depth++
				if buffer[position] != rune(')') {
					goto l571
				}
				position++
				depth--
				add(rulePAREN_CLOSE, position572)
			}
			return true
		l571:
			position, tokenIndex, depth = position571, tokenIndex571, depth571
			return false
		},
		/* 57 COMMA <- <','> */
		func() bool {
			position573, tokenIndex573, depth573 := position, tokenIndex, depth
			{
				position574 := position
				depth++
				if buffer[position] != rune(',') {
					goto l573
				}
				position++
				depth--
				add(ruleCOMMA, position574)
			}
			return true
		l573:
			position, tokenIndex, depth = position573, tokenIndex573, depth573
			return false
		},
		/* 58 _ <- <SPACE*> */
		func() bool {
			{
				position576 := position
				depth++
			l577:
				{
					position578, tokenIndex578, depth578 := position, tokenIndex, depth
					{
						position579 := position
						depth++
						{
							switch buffer[position] {
							case '\t':
								if buffer[position] != rune('\t') {
									goto l578
								}
								position++
								break
							case '\n':
								if buffer[position] != rune('\n') {
									goto l578
								}
								position++
								break
							default:
								if buffer[position] != rune(' ') {
									goto l578
								}
								position++
								break
							}
						}

						depth--
						add(ruleSPACE, position579)
					}
					goto l577
				l578:
					position, tokenIndex, depth = position578, tokenIndex578, depth578
				}
				depth--
				add(rule_, position576)
			}
			return true
		},
		/* 59 KEY <- <!ID_CONT> */
		func() bool {
			position581, tokenIndex581, depth581 := position, tokenIndex, depth
			{
				position582 := position
				depth++
				{
					position583, tokenIndex583, depth583 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l583
					}
					goto l581
				l583:
					position, tokenIndex, depth = position583, tokenIndex583, depth583
				}
				depth--
				add(ruleKEY, position582)
			}
			return true
		l581:
			position, tokenIndex, depth = position581, tokenIndex581, depth581
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
