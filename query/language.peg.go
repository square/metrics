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
	ruleAction39

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
	"Action39",

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
	rules  [99]func() bool
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
			p.addStringNode(unescapeLiteral(buffer[begin:end]))
		case ruleAction21:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction22:
			p.addGroupBy()
		case ruleAction23:

			p.addFunctionInvocation()

		case ruleAction24:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction25:
			p.addNullPredicate()
		case ruleAction26:

			p.addMetricExpression()

		case ruleAction27:

			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))

		case ruleAction28:

			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))

		case ruleAction29:
			p.addOrPredicate()
		case ruleAction30:
			p.addAndPredicate()
		case ruleAction31:
			p.addNotPredicate()
		case ruleAction32:

			p.addLiteralMatcher()

		case ruleAction33:

			p.addLiteralMatcher()
			p.addNotPredicate()

		case ruleAction34:

			p.addRegexMatcher()

		case ruleAction35:

			p.addListMatcher()

		case ruleAction36:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction37:
			p.addLiteralList()
		case ruleAction38:

			p.appendLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction39:
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
		/* 10 expression_3 <- <(expression_function / ((&('"' | '\'') (STRING Action20)) | (&('-' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') (<NUMBER> Action19)) | (&('\t' | '\n' | ' ' | '(') (PAREN_OPEN expression_1 PAREN_CLOSE)) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') expression_metric)))> */
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
							add(ruleAction21, position)
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l117
						}
						if !_rules[ruleexpressionList]() {
							goto l117
						}
						{
							add(ruleAction22, position)
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
									add(ruleAction27, position)
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
										add(ruleAction28, position)
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
							add(ruleAction23, position)
						}
						depth--
						add(ruleexpression_function, position118)
					}
					goto l116
				l117:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
					{
						switch buffer[position] {
						case '"', '\'':
							if !_rules[ruleSTRING]() {
								goto l114
							}
							{
								add(ruleAction20, position)
							}
							break
						case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							{
								position148 := position
								depth++
								if !_rules[ruleNUMBER]() {
									goto l114
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
								position150 := position
								depth++
								{
									position151 := position
									depth++
									if !_rules[ruleIDENTIFIER]() {
										goto l114
									}
									depth--
									add(rulePegText, position151)
								}
								{
									add(ruleAction24, position)
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
											add(ruleAction25, position)
										}
									}
								l155:
									goto l154

									position, tokenIndex, depth = position153, tokenIndex153, depth153
								}
							l154:
								{
									add(ruleAction26, position)
								}
								depth--
								add(ruleexpression_metric, position150)
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
		/* 11 expression_function <- <(<IDENTIFIER> Action21 PAREN_OPEN expressionList Action22 (__ groupByClause)? PAREN_CLOSE Action23)> */
		nil,
		/* 12 expression_metric <- <(<IDENTIFIER> Action24 (('[' _ predicate_1 _ ']') / Action25)? Action26)> */
		nil,
		/* 13 groupByClause <- <(('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P') __ (('b' / 'B') ('y' / 'Y')) __ <COLUMN_NAME> Action27 (COMMA <COLUMN_NAME> Action28)*)> */
		nil,
		/* 14 predicateClause <- <(('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E') __ predicate_1)> */
		nil,
		/* 15 predicate_1 <- <((predicate_2 OP_OR predicate_1 Action29) / predicate_2 / )> */
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
						add(ruleAction29, position)
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
		/* 16 predicate_2 <- <((predicate_3 OP_AND predicate_2 Action30) / predicate_3)> */
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
						add(ruleAction30, position)
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
		/* 17 predicate_3 <- <((OP_NOT predicate_3 Action31) / (PAREN_OPEN predicate_1 PAREN_CLOSE) / tagMatcher)> */
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
						add(ruleAction31, position)
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
								add(ruleAction32, position)
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
								add(ruleAction33, position)
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
								add(ruleAction34, position)
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
									add(ruleAction37, position)
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
								add(ruleAction35, position)
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
		/* 18 tagMatcher <- <((tagName _ '=' _ literalString Action32) / (tagName _ ('!' '=') _ literalString Action33) / (tagName __ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) __ literalString Action34) / (tagName __ (('i' / 'I') ('n' / 'N')) __ literalList Action35))> */
		nil,
		/* 19 literalString <- <(STRING Action36)> */
		func() bool {
			position231, tokenIndex231, depth231 := position, tokenIndex, depth
			{
				position232 := position
				depth++
				if !_rules[ruleSTRING]() {
					goto l231
				}
				{
					add(ruleAction36, position)
				}
				depth--
				add(ruleliteralString, position232)
			}
			return true
		l231:
			position, tokenIndex, depth = position231, tokenIndex231, depth231
			return false
		},
		/* 20 literalList <- <(Action37 PAREN_OPEN literalListString (COMMA literalListString)* PAREN_CLOSE)> */
		nil,
		/* 21 literalListString <- <(STRING Action38)> */
		func() bool {
			position235, tokenIndex235, depth235 := position, tokenIndex, depth
			{
				position236 := position
				depth++
				if !_rules[ruleSTRING]() {
					goto l235
				}
				{
					add(ruleAction38, position)
				}
				depth--
				add(ruleliteralListString, position236)
			}
			return true
		l235:
			position, tokenIndex, depth = position235, tokenIndex235, depth235
			return false
		},
		/* 22 tagName <- <(<TAG_NAME> Action39)> */
		func() bool {
			position238, tokenIndex238, depth238 := position, tokenIndex, depth
			{
				position239 := position
				depth++
				{
					position240 := position
					depth++
					{
						position241 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l238
						}
						depth--
						add(ruleTAG_NAME, position241)
					}
					depth--
					add(rulePegText, position240)
				}
				{
					add(ruleAction39, position)
				}
				depth--
				add(ruletagName, position239)
			}
			return true
		l238:
			position, tokenIndex, depth = position238, tokenIndex238, depth238
			return false
		},
		/* 23 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position243, tokenIndex243, depth243 := position, tokenIndex, depth
			{
				position244 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l243
				}
				depth--
				add(ruleCOLUMN_NAME, position244)
			}
			return true
		l243:
			position, tokenIndex, depth = position243, tokenIndex243, depth243
			return false
		},
		/* 24 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 25 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 26 IDENTIFIER <- <(('`' CHAR* '`') / (!(KEYWORD !ID_CONT) ID_SEGMENT ('.' ID_SEGMENT)*))> */
		func() bool {
			position247, tokenIndex247, depth247 := position, tokenIndex, depth
			{
				position248 := position
				depth++
				{
					position249, tokenIndex249, depth249 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l250
					}
					position++
				l251:
					{
						position252, tokenIndex252, depth252 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l252
						}
						goto l251
					l252:
						position, tokenIndex, depth = position252, tokenIndex252, depth252
					}
					if buffer[position] != rune('`') {
						goto l250
					}
					position++
					goto l249
				l250:
					position, tokenIndex, depth = position249, tokenIndex249, depth249
					{
						position253, tokenIndex253, depth253 := position, tokenIndex, depth
						{
							position254 := position
							depth++
							{
								position255, tokenIndex255, depth255 := position, tokenIndex, depth
								{
									position257, tokenIndex257, depth257 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l258
									}
									position++
									goto l257
								l258:
									position, tokenIndex, depth = position257, tokenIndex257, depth257
									if buffer[position] != rune('A') {
										goto l256
									}
									position++
								}
							l257:
								{
									position259, tokenIndex259, depth259 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l260
									}
									position++
									goto l259
								l260:
									position, tokenIndex, depth = position259, tokenIndex259, depth259
									if buffer[position] != rune('L') {
										goto l256
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
										goto l256
									}
									position++
								}
							l261:
								goto l255
							l256:
								position, tokenIndex, depth = position255, tokenIndex255, depth255
								{
									position264, tokenIndex264, depth264 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l265
									}
									position++
									goto l264
								l265:
									position, tokenIndex, depth = position264, tokenIndex264, depth264
									if buffer[position] != rune('A') {
										goto l263
									}
									position++
								}
							l264:
								{
									position266, tokenIndex266, depth266 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l267
									}
									position++
									goto l266
								l267:
									position, tokenIndex, depth = position266, tokenIndex266, depth266
									if buffer[position] != rune('N') {
										goto l263
									}
									position++
								}
							l266:
								{
									position268, tokenIndex268, depth268 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l269
									}
									position++
									goto l268
								l269:
									position, tokenIndex, depth = position268, tokenIndex268, depth268
									if buffer[position] != rune('D') {
										goto l263
									}
									position++
								}
							l268:
								goto l255
							l263:
								position, tokenIndex, depth = position255, tokenIndex255, depth255
								{
									position271, tokenIndex271, depth271 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l272
									}
									position++
									goto l271
								l272:
									position, tokenIndex, depth = position271, tokenIndex271, depth271
									if buffer[position] != rune('S') {
										goto l270
									}
									position++
								}
							l271:
								{
									position273, tokenIndex273, depth273 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l274
									}
									position++
									goto l273
								l274:
									position, tokenIndex, depth = position273, tokenIndex273, depth273
									if buffer[position] != rune('E') {
										goto l270
									}
									position++
								}
							l273:
								{
									position275, tokenIndex275, depth275 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l276
									}
									position++
									goto l275
								l276:
									position, tokenIndex, depth = position275, tokenIndex275, depth275
									if buffer[position] != rune('L') {
										goto l270
									}
									position++
								}
							l275:
								{
									position277, tokenIndex277, depth277 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l278
									}
									position++
									goto l277
								l278:
									position, tokenIndex, depth = position277, tokenIndex277, depth277
									if buffer[position] != rune('E') {
										goto l270
									}
									position++
								}
							l277:
								{
									position279, tokenIndex279, depth279 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l280
									}
									position++
									goto l279
								l280:
									position, tokenIndex, depth = position279, tokenIndex279, depth279
									if buffer[position] != rune('C') {
										goto l270
									}
									position++
								}
							l279:
								{
									position281, tokenIndex281, depth281 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l282
									}
									position++
									goto l281
								l282:
									position, tokenIndex, depth = position281, tokenIndex281, depth281
									if buffer[position] != rune('T') {
										goto l270
									}
									position++
								}
							l281:
								goto l255
							l270:
								position, tokenIndex, depth = position255, tokenIndex255, depth255
								{
									switch buffer[position] {
									case 'W', 'w':
										{
											position284, tokenIndex284, depth284 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l285
											}
											position++
											goto l284
										l285:
											position, tokenIndex, depth = position284, tokenIndex284, depth284
											if buffer[position] != rune('W') {
												goto l253
											}
											position++
										}
									l284:
										{
											position286, tokenIndex286, depth286 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l287
											}
											position++
											goto l286
										l287:
											position, tokenIndex, depth = position286, tokenIndex286, depth286
											if buffer[position] != rune('H') {
												goto l253
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
												goto l253
											}
											position++
										}
									l288:
										{
											position290, tokenIndex290, depth290 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l291
											}
											position++
											goto l290
										l291:
											position, tokenIndex, depth = position290, tokenIndex290, depth290
											if buffer[position] != rune('R') {
												goto l253
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
												goto l253
											}
											position++
										}
									l292:
										break
									case 'O', 'o':
										{
											position294, tokenIndex294, depth294 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l295
											}
											position++
											goto l294
										l295:
											position, tokenIndex, depth = position294, tokenIndex294, depth294
											if buffer[position] != rune('O') {
												goto l253
											}
											position++
										}
									l294:
										{
											position296, tokenIndex296, depth296 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l297
											}
											position++
											goto l296
										l297:
											position, tokenIndex, depth = position296, tokenIndex296, depth296
											if buffer[position] != rune('R') {
												goto l253
											}
											position++
										}
									l296:
										break
									case 'N', 'n':
										{
											position298, tokenIndex298, depth298 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l299
											}
											position++
											goto l298
										l299:
											position, tokenIndex, depth = position298, tokenIndex298, depth298
											if buffer[position] != rune('N') {
												goto l253
											}
											position++
										}
									l298:
										{
											position300, tokenIndex300, depth300 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l301
											}
											position++
											goto l300
										l301:
											position, tokenIndex, depth = position300, tokenIndex300, depth300
											if buffer[position] != rune('O') {
												goto l253
											}
											position++
										}
									l300:
										{
											position302, tokenIndex302, depth302 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l303
											}
											position++
											goto l302
										l303:
											position, tokenIndex, depth = position302, tokenIndex302, depth302
											if buffer[position] != rune('T') {
												goto l253
											}
											position++
										}
									l302:
										break
									case 'M', 'm':
										{
											position304, tokenIndex304, depth304 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l305
											}
											position++
											goto l304
										l305:
											position, tokenIndex, depth = position304, tokenIndex304, depth304
											if buffer[position] != rune('M') {
												goto l253
											}
											position++
										}
									l304:
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
												goto l253
											}
											position++
										}
									l306:
										{
											position308, tokenIndex308, depth308 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l309
											}
											position++
											goto l308
										l309:
											position, tokenIndex, depth = position308, tokenIndex308, depth308
											if buffer[position] != rune('T') {
												goto l253
											}
											position++
										}
									l308:
										{
											position310, tokenIndex310, depth310 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l311
											}
											position++
											goto l310
										l311:
											position, tokenIndex, depth = position310, tokenIndex310, depth310
											if buffer[position] != rune('C') {
												goto l253
											}
											position++
										}
									l310:
										{
											position312, tokenIndex312, depth312 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l313
											}
											position++
											goto l312
										l313:
											position, tokenIndex, depth = position312, tokenIndex312, depth312
											if buffer[position] != rune('H') {
												goto l253
											}
											position++
										}
									l312:
										{
											position314, tokenIndex314, depth314 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l315
											}
											position++
											goto l314
										l315:
											position, tokenIndex, depth = position314, tokenIndex314, depth314
											if buffer[position] != rune('E') {
												goto l253
											}
											position++
										}
									l314:
										{
											position316, tokenIndex316, depth316 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l317
											}
											position++
											goto l316
										l317:
											position, tokenIndex, depth = position316, tokenIndex316, depth316
											if buffer[position] != rune('S') {
												goto l253
											}
											position++
										}
									l316:
										break
									case 'I', 'i':
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
												goto l253
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
												goto l253
											}
											position++
										}
									l320:
										break
									case 'G', 'g':
										{
											position322, tokenIndex322, depth322 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l323
											}
											position++
											goto l322
										l323:
											position, tokenIndex, depth = position322, tokenIndex322, depth322
											if buffer[position] != rune('G') {
												goto l253
											}
											position++
										}
									l322:
										{
											position324, tokenIndex324, depth324 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l325
											}
											position++
											goto l324
										l325:
											position, tokenIndex, depth = position324, tokenIndex324, depth324
											if buffer[position] != rune('R') {
												goto l253
											}
											position++
										}
									l324:
										{
											position326, tokenIndex326, depth326 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l327
											}
											position++
											goto l326
										l327:
											position, tokenIndex, depth = position326, tokenIndex326, depth326
											if buffer[position] != rune('O') {
												goto l253
											}
											position++
										}
									l326:
										{
											position328, tokenIndex328, depth328 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l329
											}
											position++
											goto l328
										l329:
											position, tokenIndex, depth = position328, tokenIndex328, depth328
											if buffer[position] != rune('U') {
												goto l253
											}
											position++
										}
									l328:
										{
											position330, tokenIndex330, depth330 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l331
											}
											position++
											goto l330
										l331:
											position, tokenIndex, depth = position330, tokenIndex330, depth330
											if buffer[position] != rune('P') {
												goto l253
											}
											position++
										}
									l330:
										break
									case 'D', 'd':
										{
											position332, tokenIndex332, depth332 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l333
											}
											position++
											goto l332
										l333:
											position, tokenIndex, depth = position332, tokenIndex332, depth332
											if buffer[position] != rune('D') {
												goto l253
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
												goto l253
											}
											position++
										}
									l334:
										{
											position336, tokenIndex336, depth336 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l337
											}
											position++
											goto l336
										l337:
											position, tokenIndex, depth = position336, tokenIndex336, depth336
											if buffer[position] != rune('S') {
												goto l253
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
												goto l253
											}
											position++
										}
									l338:
										{
											position340, tokenIndex340, depth340 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l341
											}
											position++
											goto l340
										l341:
											position, tokenIndex, depth = position340, tokenIndex340, depth340
											if buffer[position] != rune('R') {
												goto l253
											}
											position++
										}
									l340:
										{
											position342, tokenIndex342, depth342 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l343
											}
											position++
											goto l342
										l343:
											position, tokenIndex, depth = position342, tokenIndex342, depth342
											if buffer[position] != rune('I') {
												goto l253
											}
											position++
										}
									l342:
										{
											position344, tokenIndex344, depth344 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l345
											}
											position++
											goto l344
										l345:
											position, tokenIndex, depth = position344, tokenIndex344, depth344
											if buffer[position] != rune('B') {
												goto l253
											}
											position++
										}
									l344:
										{
											position346, tokenIndex346, depth346 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l347
											}
											position++
											goto l346
										l347:
											position, tokenIndex, depth = position346, tokenIndex346, depth346
											if buffer[position] != rune('E') {
												goto l253
											}
											position++
										}
									l346:
										break
									case 'B', 'b':
										{
											position348, tokenIndex348, depth348 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l349
											}
											position++
											goto l348
										l349:
											position, tokenIndex, depth = position348, tokenIndex348, depth348
											if buffer[position] != rune('B') {
												goto l253
											}
											position++
										}
									l348:
										{
											position350, tokenIndex350, depth350 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l351
											}
											position++
											goto l350
										l351:
											position, tokenIndex, depth = position350, tokenIndex350, depth350
											if buffer[position] != rune('Y') {
												goto l253
											}
											position++
										}
									l350:
										break
									case 'A', 'a':
										{
											position352, tokenIndex352, depth352 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l353
											}
											position++
											goto l352
										l353:
											position, tokenIndex, depth = position352, tokenIndex352, depth352
											if buffer[position] != rune('A') {
												goto l253
											}
											position++
										}
									l352:
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
												goto l253
											}
											position++
										}
									l354:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l253
										}
										break
									}
								}

							}
						l255:
							depth--
							add(ruleKEYWORD, position254)
						}
						{
							position356, tokenIndex356, depth356 := position, tokenIndex, depth
							if !_rules[ruleID_CONT]() {
								goto l356
							}
							goto l253
						l356:
							position, tokenIndex, depth = position356, tokenIndex356, depth356
						}
						goto l247
					l253:
						position, tokenIndex, depth = position253, tokenIndex253, depth253
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l247
					}
				l357:
					{
						position358, tokenIndex358, depth358 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l358
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l358
						}
						goto l357
					l358:
						position, tokenIndex, depth = position358, tokenIndex358, depth358
					}
				}
			l249:
				depth--
				add(ruleIDENTIFIER, position248)
			}
			return true
		l247:
			position, tokenIndex, depth = position247, tokenIndex247, depth247
			return false
		},
		/* 27 TIMESTAMP <- <(<NUMBER> / STRING)> */
		nil,
		/* 28 ID_SEGMENT <- <(ID_START ID_CONT*)> */
		func() bool {
			position360, tokenIndex360, depth360 := position, tokenIndex, depth
			{
				position361 := position
				depth++
				if !_rules[ruleID_START]() {
					goto l360
				}
			l362:
				{
					position363, tokenIndex363, depth363 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l363
					}
					goto l362
				l363:
					position, tokenIndex, depth = position363, tokenIndex363, depth363
				}
				depth--
				add(ruleID_SEGMENT, position361)
			}
			return true
		l360:
			position, tokenIndex, depth = position360, tokenIndex360, depth360
			return false
		},
		/* 29 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position364, tokenIndex364, depth364 := position, tokenIndex, depth
			{
				position365 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l364
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l364
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l364
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position365)
			}
			return true
		l364:
			position, tokenIndex, depth = position364, tokenIndex364, depth364
			return false
		},
		/* 30 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position367, tokenIndex367, depth367 := position, tokenIndex, depth
			{
				position368 := position
				depth++
				{
					position369, tokenIndex369, depth369 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l370
					}
					goto l369
				l370:
					position, tokenIndex, depth = position369, tokenIndex369, depth369
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l367
					}
					position++
				}
			l369:
				depth--
				add(ruleID_CONT, position368)
			}
			return true
		l367:
			position, tokenIndex, depth = position367, tokenIndex367, depth367
			return false
		},
		/* 31 PROPERTY_KEY <- <((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> __ (('b' / 'B') ('y' / 'Y')))) | (&('R' | 'r') <(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))>) | (&('T' | 't') <(('t' / 'T') ('o' / 'O'))>) | (&('F' | 'f') <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))>))> */
		func() bool {
			position371, tokenIndex371, depth371 := position, tokenIndex, depth
			{
				position372 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position374 := position
							depth++
							{
								position375, tokenIndex375, depth375 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l376
								}
								position++
								goto l375
							l376:
								position, tokenIndex, depth = position375, tokenIndex375, depth375
								if buffer[position] != rune('S') {
									goto l371
								}
								position++
							}
						l375:
							{
								position377, tokenIndex377, depth377 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l378
								}
								position++
								goto l377
							l378:
								position, tokenIndex, depth = position377, tokenIndex377, depth377
								if buffer[position] != rune('A') {
									goto l371
								}
								position++
							}
						l377:
							{
								position379, tokenIndex379, depth379 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l380
								}
								position++
								goto l379
							l380:
								position, tokenIndex, depth = position379, tokenIndex379, depth379
								if buffer[position] != rune('M') {
									goto l371
								}
								position++
							}
						l379:
							{
								position381, tokenIndex381, depth381 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l382
								}
								position++
								goto l381
							l382:
								position, tokenIndex, depth = position381, tokenIndex381, depth381
								if buffer[position] != rune('P') {
									goto l371
								}
								position++
							}
						l381:
							{
								position383, tokenIndex383, depth383 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l384
								}
								position++
								goto l383
							l384:
								position, tokenIndex, depth = position383, tokenIndex383, depth383
								if buffer[position] != rune('L') {
									goto l371
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
									goto l371
								}
								position++
							}
						l385:
							depth--
							add(rulePegText, position374)
						}
						if !_rules[rule__]() {
							goto l371
						}
						{
							position387, tokenIndex387, depth387 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l388
							}
							position++
							goto l387
						l388:
							position, tokenIndex, depth = position387, tokenIndex387, depth387
							if buffer[position] != rune('B') {
								goto l371
							}
							position++
						}
					l387:
						{
							position389, tokenIndex389, depth389 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l390
							}
							position++
							goto l389
						l390:
							position, tokenIndex, depth = position389, tokenIndex389, depth389
							if buffer[position] != rune('Y') {
								goto l371
							}
							position++
						}
					l389:
						break
					case 'R', 'r':
						{
							position391 := position
							depth++
							{
								position392, tokenIndex392, depth392 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l393
								}
								position++
								goto l392
							l393:
								position, tokenIndex, depth = position392, tokenIndex392, depth392
								if buffer[position] != rune('R') {
									goto l371
								}
								position++
							}
						l392:
							{
								position394, tokenIndex394, depth394 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l395
								}
								position++
								goto l394
							l395:
								position, tokenIndex, depth = position394, tokenIndex394, depth394
								if buffer[position] != rune('E') {
									goto l371
								}
								position++
							}
						l394:
							{
								position396, tokenIndex396, depth396 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l397
								}
								position++
								goto l396
							l397:
								position, tokenIndex, depth = position396, tokenIndex396, depth396
								if buffer[position] != rune('S') {
									goto l371
								}
								position++
							}
						l396:
							{
								position398, tokenIndex398, depth398 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l399
								}
								position++
								goto l398
							l399:
								position, tokenIndex, depth = position398, tokenIndex398, depth398
								if buffer[position] != rune('O') {
									goto l371
								}
								position++
							}
						l398:
							{
								position400, tokenIndex400, depth400 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l401
								}
								position++
								goto l400
							l401:
								position, tokenIndex, depth = position400, tokenIndex400, depth400
								if buffer[position] != rune('L') {
									goto l371
								}
								position++
							}
						l400:
							{
								position402, tokenIndex402, depth402 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l403
								}
								position++
								goto l402
							l403:
								position, tokenIndex, depth = position402, tokenIndex402, depth402
								if buffer[position] != rune('U') {
									goto l371
								}
								position++
							}
						l402:
							{
								position404, tokenIndex404, depth404 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l405
								}
								position++
								goto l404
							l405:
								position, tokenIndex, depth = position404, tokenIndex404, depth404
								if buffer[position] != rune('T') {
									goto l371
								}
								position++
							}
						l404:
							{
								position406, tokenIndex406, depth406 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l407
								}
								position++
								goto l406
							l407:
								position, tokenIndex, depth = position406, tokenIndex406, depth406
								if buffer[position] != rune('I') {
									goto l371
								}
								position++
							}
						l406:
							{
								position408, tokenIndex408, depth408 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l409
								}
								position++
								goto l408
							l409:
								position, tokenIndex, depth = position408, tokenIndex408, depth408
								if buffer[position] != rune('O') {
									goto l371
								}
								position++
							}
						l408:
							{
								position410, tokenIndex410, depth410 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l411
								}
								position++
								goto l410
							l411:
								position, tokenIndex, depth = position410, tokenIndex410, depth410
								if buffer[position] != rune('N') {
									goto l371
								}
								position++
							}
						l410:
							depth--
							add(rulePegText, position391)
						}
						break
					case 'T', 't':
						{
							position412 := position
							depth++
							{
								position413, tokenIndex413, depth413 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l414
								}
								position++
								goto l413
							l414:
								position, tokenIndex, depth = position413, tokenIndex413, depth413
								if buffer[position] != rune('T') {
									goto l371
								}
								position++
							}
						l413:
							{
								position415, tokenIndex415, depth415 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l416
								}
								position++
								goto l415
							l416:
								position, tokenIndex, depth = position415, tokenIndex415, depth415
								if buffer[position] != rune('O') {
									goto l371
								}
								position++
							}
						l415:
							depth--
							add(rulePegText, position412)
						}
						break
					default:
						{
							position417 := position
							depth++
							{
								position418, tokenIndex418, depth418 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l419
								}
								position++
								goto l418
							l419:
								position, tokenIndex, depth = position418, tokenIndex418, depth418
								if buffer[position] != rune('F') {
									goto l371
								}
								position++
							}
						l418:
							{
								position420, tokenIndex420, depth420 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l421
								}
								position++
								goto l420
							l421:
								position, tokenIndex, depth = position420, tokenIndex420, depth420
								if buffer[position] != rune('R') {
									goto l371
								}
								position++
							}
						l420:
							{
								position422, tokenIndex422, depth422 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l423
								}
								position++
								goto l422
							l423:
								position, tokenIndex, depth = position422, tokenIndex422, depth422
								if buffer[position] != rune('O') {
									goto l371
								}
								position++
							}
						l422:
							{
								position424, tokenIndex424, depth424 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l425
								}
								position++
								goto l424
							l425:
								position, tokenIndex, depth = position424, tokenIndex424, depth424
								if buffer[position] != rune('M') {
									goto l371
								}
								position++
							}
						l424:
							depth--
							add(rulePegText, position417)
						}
						break
					}
				}

				depth--
				add(rulePROPERTY_KEY, position372)
			}
			return true
		l371:
			position, tokenIndex, depth = position371, tokenIndex371, depth371
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
			position435, tokenIndex435, depth435 := position, tokenIndex, depth
			{
				position436 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l435
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position436)
			}
			return true
		l435:
			position, tokenIndex, depth = position435, tokenIndex435, depth435
			return false
		},
		/* 42 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position437, tokenIndex437, depth437 := position, tokenIndex, depth
			{
				position438 := position
				depth++
				if buffer[position] != rune('"') {
					goto l437
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position438)
			}
			return true
		l437:
			position, tokenIndex, depth = position437, tokenIndex437, depth437
			return false
		},
		/* 43 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> QUOTE_SINGLE) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> QUOTE_DOUBLE))> */
		func() bool {
			position439, tokenIndex439, depth439 := position, tokenIndex, depth
			{
				position440 := position
				depth++
				{
					position441, tokenIndex441, depth441 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l442
					}
					{
						position443 := position
						depth++
					l444:
						{
							position445, tokenIndex445, depth445 := position, tokenIndex, depth
							{
								position446, tokenIndex446, depth446 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l446
								}
								goto l445
							l446:
								position, tokenIndex, depth = position446, tokenIndex446, depth446
							}
							if !_rules[ruleCHAR]() {
								goto l445
							}
							goto l444
						l445:
							position, tokenIndex, depth = position445, tokenIndex445, depth445
						}
						depth--
						add(rulePegText, position443)
					}
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l442
					}
					goto l441
				l442:
					position, tokenIndex, depth = position441, tokenIndex441, depth441
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l439
					}
					{
						position447 := position
						depth++
					l448:
						{
							position449, tokenIndex449, depth449 := position, tokenIndex, depth
							{
								position450, tokenIndex450, depth450 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l450
								}
								goto l449
							l450:
								position, tokenIndex, depth = position450, tokenIndex450, depth450
							}
							if !_rules[ruleCHAR]() {
								goto l449
							}
							goto l448
						l449:
							position, tokenIndex, depth = position449, tokenIndex449, depth449
						}
						depth--
						add(rulePegText, position447)
					}
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l439
					}
				}
			l441:
				depth--
				add(ruleSTRING, position440)
			}
			return true
		l439:
			position, tokenIndex, depth = position439, tokenIndex439, depth439
			return false
		},
		/* 44 CHAR <- <(('\\' ((&('"') QUOTE_DOUBLE) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position451, tokenIndex451, depth451 := position, tokenIndex, depth
			{
				position452 := position
				depth++
				{
					position453, tokenIndex453, depth453 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l454
					}
					position++
					{
						switch buffer[position] {
						case '"':
							if !_rules[ruleQUOTE_DOUBLE]() {
								goto l454
							}
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l454
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l454
							}
							break
						}
					}

					goto l453
				l454:
					position, tokenIndex, depth = position453, tokenIndex453, depth453
					{
						position456, tokenIndex456, depth456 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l456
						}
						goto l451
					l456:
						position, tokenIndex, depth = position456, tokenIndex456, depth456
					}
					if !matchDot() {
						goto l451
					}
				}
			l453:
				depth--
				add(ruleCHAR, position452)
			}
			return true
		l451:
			position, tokenIndex, depth = position451, tokenIndex451, depth451
			return false
		},
		/* 45 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position457, tokenIndex457, depth457 := position, tokenIndex, depth
			{
				position458 := position
				depth++
				{
					position459, tokenIndex459, depth459 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l460
					}
					position++
					goto l459
				l460:
					position, tokenIndex, depth = position459, tokenIndex459, depth459
					if buffer[position] != rune('\\') {
						goto l457
					}
					position++
				}
			l459:
				depth--
				add(ruleESCAPE_CLASS, position458)
			}
			return true
		l457:
			position, tokenIndex, depth = position457, tokenIndex457, depth457
			return false
		},
		/* 46 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position461, tokenIndex461, depth461 := position, tokenIndex, depth
			{
				position462 := position
				depth++
				{
					position463 := position
					depth++
					{
						position464, tokenIndex464, depth464 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l464
						}
						position++
						goto l465
					l464:
						position, tokenIndex, depth = position464, tokenIndex464, depth464
					}
				l465:
					{
						position466 := position
						depth++
						{
							position467, tokenIndex467, depth467 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l468
							}
							position++
							goto l467
						l468:
							position, tokenIndex, depth = position467, tokenIndex467, depth467
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l461
							}
							position++
						l469:
							{
								position470, tokenIndex470, depth470 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l470
								}
								position++
								goto l469
							l470:
								position, tokenIndex, depth = position470, tokenIndex470, depth470
							}
						}
					l467:
						depth--
						add(ruleNUMBER_NATURAL, position466)
					}
					depth--
					add(ruleNUMBER_INTEGER, position463)
				}
				{
					position471, tokenIndex471, depth471 := position, tokenIndex, depth
					{
						position473 := position
						depth++
						if buffer[position] != rune('.') {
							goto l471
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l471
						}
						position++
					l474:
						{
							position475, tokenIndex475, depth475 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l475
							}
							position++
							goto l474
						l475:
							position, tokenIndex, depth = position475, tokenIndex475, depth475
						}
						depth--
						add(ruleNUMBER_FRACTION, position473)
					}
					goto l472
				l471:
					position, tokenIndex, depth = position471, tokenIndex471, depth471
				}
			l472:
				{
					position476, tokenIndex476, depth476 := position, tokenIndex, depth
					{
						position478 := position
						depth++
						{
							position479, tokenIndex479, depth479 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l480
							}
							position++
							goto l479
						l480:
							position, tokenIndex, depth = position479, tokenIndex479, depth479
							if buffer[position] != rune('E') {
								goto l476
							}
							position++
						}
					l479:
						{
							position481, tokenIndex481, depth481 := position, tokenIndex, depth
							{
								position483, tokenIndex483, depth483 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l484
								}
								position++
								goto l483
							l484:
								position, tokenIndex, depth = position483, tokenIndex483, depth483
								if buffer[position] != rune('-') {
									goto l481
								}
								position++
							}
						l483:
							goto l482
						l481:
							position, tokenIndex, depth = position481, tokenIndex481, depth481
						}
					l482:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l476
						}
						position++
					l485:
						{
							position486, tokenIndex486, depth486 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l486
							}
							position++
							goto l485
						l486:
							position, tokenIndex, depth = position486, tokenIndex486, depth486
						}
						depth--
						add(ruleNUMBER_EXP, position478)
					}
					goto l477
				l476:
					position, tokenIndex, depth = position476, tokenIndex476, depth476
				}
			l477:
				depth--
				add(ruleNUMBER, position462)
			}
			return true
		l461:
			position, tokenIndex, depth = position461, tokenIndex461, depth461
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
			position491, tokenIndex491, depth491 := position, tokenIndex, depth
			{
				position492 := position
				depth++
				if !_rules[rule_]() {
					goto l491
				}
				if buffer[position] != rune('(') {
					goto l491
				}
				position++
				if !_rules[rule_]() {
					goto l491
				}
				depth--
				add(rulePAREN_OPEN, position492)
			}
			return true
		l491:
			position, tokenIndex, depth = position491, tokenIndex491, depth491
			return false
		},
		/* 52 PAREN_CLOSE <- <(_ ')' _)> */
		func() bool {
			position493, tokenIndex493, depth493 := position, tokenIndex, depth
			{
				position494 := position
				depth++
				if !_rules[rule_]() {
					goto l493
				}
				if buffer[position] != rune(')') {
					goto l493
				}
				position++
				if !_rules[rule_]() {
					goto l493
				}
				depth--
				add(rulePAREN_CLOSE, position494)
			}
			return true
		l493:
			position, tokenIndex, depth = position493, tokenIndex493, depth493
			return false
		},
		/* 53 COMMA <- <(_ ',' _)> */
		func() bool {
			position495, tokenIndex495, depth495 := position, tokenIndex, depth
			{
				position496 := position
				depth++
				if !_rules[rule_]() {
					goto l495
				}
				if buffer[position] != rune(',') {
					goto l495
				}
				position++
				if !_rules[rule_]() {
					goto l495
				}
				depth--
				add(ruleCOMMA, position496)
			}
			return true
		l495:
			position, tokenIndex, depth = position495, tokenIndex495, depth495
			return false
		},
		/* 54 _ <- <SPACE*> */
		func() bool {
			{
				position498 := position
				depth++
			l499:
				{
					position500, tokenIndex500, depth500 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l500
					}
					goto l499
				l500:
					position, tokenIndex, depth = position500, tokenIndex500, depth500
				}
				depth--
				add(rule_, position498)
			}
			return true
		},
		/* 55 __ <- <SPACE+> */
		func() bool {
			position501, tokenIndex501, depth501 := position, tokenIndex, depth
			{
				position502 := position
				depth++
				if !_rules[ruleSPACE]() {
					goto l501
				}
			l503:
				{
					position504, tokenIndex504, depth504 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l504
					}
					goto l503
				l504:
					position, tokenIndex, depth = position504, tokenIndex504, depth504
				}
				depth--
				add(rule__, position502)
			}
			return true
		l501:
			position, tokenIndex, depth = position501, tokenIndex501, depth501
			return false
		},
		/* 56 SPACE <- <((&('\t') '\t') | (&('\n') '\n') | (&(' ') ' '))> */
		func() bool {
			position505, tokenIndex505, depth505 := position, tokenIndex, depth
			{
				position506 := position
				depth++
				{
					switch buffer[position] {
					case '\t':
						if buffer[position] != rune('\t') {
							goto l505
						}
						position++
						break
					case '\n':
						if buffer[position] != rune('\n') {
							goto l505
						}
						position++
						break
					default:
						if buffer[position] != rune(' ') {
							goto l505
						}
						position++
						break
					}
				}

				depth--
				add(ruleSPACE, position506)
			}
			return true
		l505:
			position, tokenIndex, depth = position505, tokenIndex505, depth505
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
		/* 79 Action20 <- <{ p.addStringNode(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 80 Action21 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 81 Action22 <- <{ p.addGroupBy() }> */
		nil,
		/* 82 Action23 <- <{
		   p.addFunctionInvocation()
		 }> */
		nil,
		/* 83 Action24 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 84 Action25 <- <{ p.addNullPredicate() }> */
		nil,
		/* 85 Action26 <- <{
		   p.addMetricExpression()
		 }> */
		nil,
		/* 86 Action27 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 87 Action28 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		   }> */
		nil,
		/* 88 Action29 <- <{ p.addOrPredicate() }> */
		nil,
		/* 89 Action30 <- <{ p.addAndPredicate() }> */
		nil,
		/* 90 Action31 <- <{ p.addNotPredicate() }> */
		nil,
		/* 91 Action32 <- <{
		   p.addLiteralMatcher()
		 }> */
		nil,
		/* 92 Action33 <- <{
		   p.addLiteralMatcher()
		   p.addNotPredicate()
		 }> */
		nil,
		/* 93 Action34 <- <{
		   p.addRegexMatcher()
		 }> */
		nil,
		/* 94 Action35 <- <{
		   p.addListMatcher()
		 }> */
		nil,
		/* 95 Action36 <- <{
		  p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 96 Action37 <- <{ p.addLiteralList() }> */
		nil,
		/* 97 Action38 <- <{
		  p.appendLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 98 Action39 <- <{ p.addTagLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
	}
	p.rules = _rules
}
