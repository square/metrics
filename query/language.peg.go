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
	rule__
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
	"__",
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
											switch buffer[position] {
											case 'N', 'n':
												{
													position25 := position
													depth++
													{
														position26, tokenIndex26, depth26 := position, tokenIndex, depth
														if buffer[position] != rune('n') {
															goto l27
														}
														position++
														goto l26
													l27:
														position, tokenIndex, depth = position26, tokenIndex26, depth26
														if buffer[position] != rune('N') {
															goto l20
														}
														position++
													}
												l26:
													{
														position28, tokenIndex28, depth28 := position, tokenIndex, depth
														if buffer[position] != rune('o') {
															goto l29
														}
														position++
														goto l28
													l29:
														position, tokenIndex, depth = position28, tokenIndex28, depth28
														if buffer[position] != rune('O') {
															goto l20
														}
														position++
													}
												l28:
													{
														position30, tokenIndex30, depth30 := position, tokenIndex, depth
														if buffer[position] != rune('w') {
															goto l31
														}
														position++
														goto l30
													l31:
														position, tokenIndex, depth = position30, tokenIndex30, depth30
														if buffer[position] != rune('W') {
															goto l20
														}
														position++
													}
												l30:
													depth--
													add(rulePegText, position25)
												}
												break
											case '"', '\'':
												if !_rules[ruleSTRING]() {
													goto l20
												}
												break
											default:
												{
													position32 := position
													depth++
													if !_rules[ruleNUMBER]() {
														goto l20
													}
													{
														position33, tokenIndex33, depth33 := position, tokenIndex, depth
														{
															position35, tokenIndex35, depth35 := position, tokenIndex, depth
															if c := buffer[position]; c < rune('a') || c > rune('z') {
																goto l36
															}
															position++
															goto l35
														l36:
															position, tokenIndex, depth = position35, tokenIndex35, depth35
															if c := buffer[position]; c < rune('A') || c > rune('Z') {
																goto l33
															}
															position++
														}
													l35:
														goto l34
													l33:
														position, tokenIndex, depth = position33, tokenIndex33, depth33
													}
												l34:
													depth--
													add(rulePegText, position32)
												}
												break
											}
										}

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
						position41 := position
						depth++
						if !_rules[rule_]() {
							goto l0
						}
						{
							position42, tokenIndex42, depth42 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l43
							}
							position++
							goto l42
						l43:
							position, tokenIndex, depth = position42, tokenIndex42, depth42
							if buffer[position] != rune('D') {
								goto l0
							}
							position++
						}
					l42:
						{
							position44, tokenIndex44, depth44 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l45
							}
							position++
							goto l44
						l45:
							position, tokenIndex, depth = position44, tokenIndex44, depth44
							if buffer[position] != rune('E') {
								goto l0
							}
							position++
						}
					l44:
						{
							position46, tokenIndex46, depth46 := position, tokenIndex, depth
							if buffer[position] != rune('s') {
								goto l47
							}
							position++
							goto l46
						l47:
							position, tokenIndex, depth = position46, tokenIndex46, depth46
							if buffer[position] != rune('S') {
								goto l0
							}
							position++
						}
					l46:
						{
							position48, tokenIndex48, depth48 := position, tokenIndex, depth
							if buffer[position] != rune('c') {
								goto l49
							}
							position++
							goto l48
						l49:
							position, tokenIndex, depth = position48, tokenIndex48, depth48
							if buffer[position] != rune('C') {
								goto l0
							}
							position++
						}
					l48:
						{
							position50, tokenIndex50, depth50 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l51
							}
							position++
							goto l50
						l51:
							position, tokenIndex, depth = position50, tokenIndex50, depth50
							if buffer[position] != rune('R') {
								goto l0
							}
							position++
						}
					l50:
						{
							position52, tokenIndex52, depth52 := position, tokenIndex, depth
							if buffer[position] != rune('i') {
								goto l53
							}
							position++
							goto l52
						l53:
							position, tokenIndex, depth = position52, tokenIndex52, depth52
							if buffer[position] != rune('I') {
								goto l0
							}
							position++
						}
					l52:
						{
							position54, tokenIndex54, depth54 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l55
							}
							position++
							goto l54
						l55:
							position, tokenIndex, depth = position54, tokenIndex54, depth54
							if buffer[position] != rune('B') {
								goto l0
							}
							position++
						}
					l54:
						{
							position56, tokenIndex56, depth56 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l57
							}
							position++
							goto l56
						l57:
							position, tokenIndex, depth = position56, tokenIndex56, depth56
							if buffer[position] != rune('E') {
								goto l0
							}
							position++
						}
					l56:
						if !_rules[rule__]() {
							goto l0
						}
						{
							position58, tokenIndex58, depth58 := position, tokenIndex, depth
							{
								position60 := position
								depth++
								{
									position61, tokenIndex61, depth61 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l62
									}
									position++
									goto l61
								l62:
									position, tokenIndex, depth = position61, tokenIndex61, depth61
									if buffer[position] != rune('A') {
										goto l59
									}
									position++
								}
							l61:
								{
									position63, tokenIndex63, depth63 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l64
									}
									position++
									goto l63
								l64:
									position, tokenIndex, depth = position63, tokenIndex63, depth63
									if buffer[position] != rune('L') {
										goto l59
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
										goto l59
									}
									position++
								}
							l65:
								{
									add(ruleAction1, position)
								}
								depth--
								add(ruledescribeAllStmt, position60)
							}
							goto l58
						l59:
							position, tokenIndex, depth = position58, tokenIndex58, depth58
							{
								position69 := position
								depth++
								{
									position70, tokenIndex70, depth70 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l71
									}
									position++
									goto l70
								l71:
									position, tokenIndex, depth = position70, tokenIndex70, depth70
									if buffer[position] != rune('M') {
										goto l68
									}
									position++
								}
							l70:
								{
									position72, tokenIndex72, depth72 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l73
									}
									position++
									goto l72
								l73:
									position, tokenIndex, depth = position72, tokenIndex72, depth72
									if buffer[position] != rune('E') {
										goto l68
									}
									position++
								}
							l72:
								{
									position74, tokenIndex74, depth74 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l75
									}
									position++
									goto l74
								l75:
									position, tokenIndex, depth = position74, tokenIndex74, depth74
									if buffer[position] != rune('T') {
										goto l68
									}
									position++
								}
							l74:
								{
									position76, tokenIndex76, depth76 := position, tokenIndex, depth
									if buffer[position] != rune('r') {
										goto l77
									}
									position++
									goto l76
								l77:
									position, tokenIndex, depth = position76, tokenIndex76, depth76
									if buffer[position] != rune('R') {
										goto l68
									}
									position++
								}
							l76:
								{
									position78, tokenIndex78, depth78 := position, tokenIndex, depth
									if buffer[position] != rune('i') {
										goto l79
									}
									position++
									goto l78
								l79:
									position, tokenIndex, depth = position78, tokenIndex78, depth78
									if buffer[position] != rune('I') {
										goto l68
									}
									position++
								}
							l78:
								{
									position80, tokenIndex80, depth80 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l81
									}
									position++
									goto l80
								l81:
									position, tokenIndex, depth = position80, tokenIndex80, depth80
									if buffer[position] != rune('C') {
										goto l68
									}
									position++
								}
							l80:
								{
									position82, tokenIndex82, depth82 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l83
									}
									position++
									goto l82
								l83:
									position, tokenIndex, depth = position82, tokenIndex82, depth82
									if buffer[position] != rune('S') {
										goto l68
									}
									position++
								}
							l82:
								if !_rules[rule__]() {
									goto l68
								}
								{
									position84, tokenIndex84, depth84 := position, tokenIndex, depth
									if buffer[position] != rune('w') {
										goto l85
									}
									position++
									goto l84
								l85:
									position, tokenIndex, depth = position84, tokenIndex84, depth84
									if buffer[position] != rune('W') {
										goto l68
									}
									position++
								}
							l84:
								{
									position86, tokenIndex86, depth86 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l87
									}
									position++
									goto l86
								l87:
									position, tokenIndex, depth = position86, tokenIndex86, depth86
									if buffer[position] != rune('H') {
										goto l68
									}
									position++
								}
							l86:
								{
									position88, tokenIndex88, depth88 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l89
									}
									position++
									goto l88
								l89:
									position, tokenIndex, depth = position88, tokenIndex88, depth88
									if buffer[position] != rune('E') {
										goto l68
									}
									position++
								}
							l88:
								{
									position90, tokenIndex90, depth90 := position, tokenIndex, depth
									if buffer[position] != rune('r') {
										goto l91
									}
									position++
									goto l90
								l91:
									position, tokenIndex, depth = position90, tokenIndex90, depth90
									if buffer[position] != rune('R') {
										goto l68
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
										goto l68
									}
									position++
								}
							l92:
								if !_rules[rule__]() {
									goto l68
								}
								if !_rules[ruletagName]() {
									goto l68
								}
								if !_rules[rule_]() {
									goto l68
								}
								if buffer[position] != rune('=') {
									goto l68
								}
								position++
								if !_rules[rule_]() {
									goto l68
								}
								if !_rules[ruleliteralString]() {
									goto l68
								}
								{
									add(ruleAction2, position)
								}
								depth--
								add(ruledescribeMetrics, position69)
							}
							goto l58
						l68:
							position, tokenIndex, depth = position58, tokenIndex58, depth58
							{
								position95 := position
								depth++
								{
									position96 := position
									depth++
									{
										position97 := position
										depth++
										if !_rules[ruleIDENTIFIER]() {
											goto l0
										}
										depth--
										add(ruleMETRIC_NAME, position97)
									}
									depth--
									add(rulePegText, position96)
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
								add(ruledescribeSingleStmt, position95)
							}
						}
					l58:
						if !_rules[rule_]() {
							goto l0
						}
						depth--
						add(ruledescribeStmt, position41)
					}
				}
			l2:
				{
					position100, tokenIndex100, depth100 := position, tokenIndex, depth
					if !matchDot() {
						goto l100
					}
					goto l0
				l100:
					position, tokenIndex, depth = position100, tokenIndex100, depth100
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
		/* 2 describeStmt <- <(_ (('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E')) __ (describeAllStmt / describeMetrics / describeSingleStmt) _)> */
		nil,
		/* 3 describeAllStmt <- <(('a' / 'A') ('l' / 'L') ('l' / 'L') Action1)> */
		nil,
		/* 4 describeMetrics <- <(('m' / 'M') ('e' / 'E') ('t' / 'T') ('r' / 'R') ('i' / 'I') ('c' / 'C') ('s' / 'S') __ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) __ tagName _ '=' _ literalString Action2)> */
		nil,
		/* 5 describeSingleStmt <- <(<METRIC_NAME> Action3 optionalPredicateClause Action4)> */
		nil,
		/* 6 propertyClause <- <(Action5 (_ PROPERTY_KEY Action6 __ PROPERTY_VALUE Action7 Action8)* Action9)> */
		nil,
		/* 7 optionalPredicateClause <- <((_ predicateClause) / Action10)> */
		func() bool {
			{
				position108 := position
				depth++
				{
					position109, tokenIndex109, depth109 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l110
					}
					{
						position111 := position
						depth++
						{
							position112, tokenIndex112, depth112 := position, tokenIndex, depth
							if buffer[position] != rune('w') {
								goto l113
							}
							position++
							goto l112
						l113:
							position, tokenIndex, depth = position112, tokenIndex112, depth112
							if buffer[position] != rune('W') {
								goto l110
							}
							position++
						}
					l112:
						{
							position114, tokenIndex114, depth114 := position, tokenIndex, depth
							if buffer[position] != rune('h') {
								goto l115
							}
							position++
							goto l114
						l115:
							position, tokenIndex, depth = position114, tokenIndex114, depth114
							if buffer[position] != rune('H') {
								goto l110
							}
							position++
						}
					l114:
						{
							position116, tokenIndex116, depth116 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l117
							}
							position++
							goto l116
						l117:
							position, tokenIndex, depth = position116, tokenIndex116, depth116
							if buffer[position] != rune('E') {
								goto l110
							}
							position++
						}
					l116:
						{
							position118, tokenIndex118, depth118 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l119
							}
							position++
							goto l118
						l119:
							position, tokenIndex, depth = position118, tokenIndex118, depth118
							if buffer[position] != rune('R') {
								goto l110
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
								goto l110
							}
							position++
						}
					l120:
						if !_rules[rule__]() {
							goto l110
						}
						if !_rules[rulepredicate_1]() {
							goto l110
						}
						depth--
						add(rulepredicateClause, position111)
					}
					goto l109
				l110:
					position, tokenIndex, depth = position109, tokenIndex109, depth109
					{
						add(ruleAction10, position)
					}
				}
			l109:
				depth--
				add(ruleoptionalPredicateClause, position108)
			}
			return true
		},
		/* 8 expressionList <- <(Action11 expression_1 Action12 (COMMA expression_1 Action13)*)> */
		func() bool {
			position123, tokenIndex123, depth123 := position, tokenIndex, depth
			{
				position124 := position
				depth++
				{
					add(ruleAction11, position)
				}
				if !_rules[ruleexpression_1]() {
					goto l123
				}
				{
					add(ruleAction12, position)
				}
			l127:
				{
					position128, tokenIndex128, depth128 := position, tokenIndex, depth
					if !_rules[ruleCOMMA]() {
						goto l128
					}
					if !_rules[ruleexpression_1]() {
						goto l128
					}
					{
						add(ruleAction13, position)
					}
					goto l127
				l128:
					position, tokenIndex, depth = position128, tokenIndex128, depth128
				}
				depth--
				add(ruleexpressionList, position124)
			}
			return true
		l123:
			position, tokenIndex, depth = position123, tokenIndex123, depth123
			return false
		},
		/* 9 expression_1 <- <(expression_2 (((OP_ADD Action14) / (OP_SUB Action15)) expression_2 Action16)*)> */
		func() bool {
			position130, tokenIndex130, depth130 := position, tokenIndex, depth
			{
				position131 := position
				depth++
				if !_rules[ruleexpression_2]() {
					goto l130
				}
			l132:
				{
					position133, tokenIndex133, depth133 := position, tokenIndex, depth
					{
						position134, tokenIndex134, depth134 := position, tokenIndex, depth
						{
							position136 := position
							depth++
							if !_rules[rule_]() {
								goto l135
							}
							if buffer[position] != rune('+') {
								goto l135
							}
							position++
							if !_rules[rule_]() {
								goto l135
							}
							depth--
							add(ruleOP_ADD, position136)
						}
						{
							add(ruleAction14, position)
						}
						goto l134
					l135:
						position, tokenIndex, depth = position134, tokenIndex134, depth134
						{
							position138 := position
							depth++
							if !_rules[rule_]() {
								goto l133
							}
							if buffer[position] != rune('-') {
								goto l133
							}
							position++
							if !_rules[rule_]() {
								goto l133
							}
							depth--
							add(ruleOP_SUB, position138)
						}
						{
							add(ruleAction15, position)
						}
					}
				l134:
					if !_rules[ruleexpression_2]() {
						goto l133
					}
					{
						add(ruleAction16, position)
					}
					goto l132
				l133:
					position, tokenIndex, depth = position133, tokenIndex133, depth133
				}
				depth--
				add(ruleexpression_1, position131)
			}
			return true
		l130:
			position, tokenIndex, depth = position130, tokenIndex130, depth130
			return false
		},
		/* 10 expression_2 <- <(expression_3 (((OP_DIV Action17) / (OP_MULT Action18)) expression_3 Action19)*)> */
		func() bool {
			position141, tokenIndex141, depth141 := position, tokenIndex, depth
			{
				position142 := position
				depth++
				if !_rules[ruleexpression_3]() {
					goto l141
				}
			l143:
				{
					position144, tokenIndex144, depth144 := position, tokenIndex, depth
					{
						position145, tokenIndex145, depth145 := position, tokenIndex, depth
						{
							position147 := position
							depth++
							if !_rules[rule_]() {
								goto l146
							}
							if buffer[position] != rune('/') {
								goto l146
							}
							position++
							if !_rules[rule_]() {
								goto l146
							}
							depth--
							add(ruleOP_DIV, position147)
						}
						{
							add(ruleAction17, position)
						}
						goto l145
					l146:
						position, tokenIndex, depth = position145, tokenIndex145, depth145
						{
							position149 := position
							depth++
							if !_rules[rule_]() {
								goto l144
							}
							if buffer[position] != rune('*') {
								goto l144
							}
							position++
							if !_rules[rule_]() {
								goto l144
							}
							depth--
							add(ruleOP_MULT, position149)
						}
						{
							add(ruleAction18, position)
						}
					}
				l145:
					if !_rules[ruleexpression_3]() {
						goto l144
					}
					{
						add(ruleAction19, position)
					}
					goto l143
				l144:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
				}
				depth--
				add(ruleexpression_2, position142)
			}
			return true
		l141:
			position, tokenIndex, depth = position141, tokenIndex141, depth141
			return false
		},
		/* 11 expression_3 <- <(expression_function / ((&('"' | '\'') (STRING Action21)) | (&('-' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') (<NUMBER> Action20)) | (&('\t' | '\n' | ' ' | '(') (PAREN_OPEN expression_1 PAREN_CLOSE)) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') expression_metric)))> */
		func() bool {
			position152, tokenIndex152, depth152 := position, tokenIndex, depth
			{
				position153 := position
				depth++
				{
					position154, tokenIndex154, depth154 := position, tokenIndex, depth
					{
						position156 := position
						depth++
						{
							position157 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l155
							}
							depth--
							add(rulePegText, position157)
						}
						{
							add(ruleAction22, position)
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l155
						}
						if !_rules[ruleexpressionList]() {
							goto l155
						}
						{
							add(ruleAction23, position)
						}
						{
							position160, tokenIndex160, depth160 := position, tokenIndex, depth
							if !_rules[rule__]() {
								goto l160
							}
							{
								position162 := position
								depth++
								{
									position163, tokenIndex163, depth163 := position, tokenIndex, depth
									if buffer[position] != rune('g') {
										goto l164
									}
									position++
									goto l163
								l164:
									position, tokenIndex, depth = position163, tokenIndex163, depth163
									if buffer[position] != rune('G') {
										goto l160
									}
									position++
								}
							l163:
								{
									position165, tokenIndex165, depth165 := position, tokenIndex, depth
									if buffer[position] != rune('r') {
										goto l166
									}
									position++
									goto l165
								l166:
									position, tokenIndex, depth = position165, tokenIndex165, depth165
									if buffer[position] != rune('R') {
										goto l160
									}
									position++
								}
							l165:
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
										goto l160
									}
									position++
								}
							l167:
								{
									position169, tokenIndex169, depth169 := position, tokenIndex, depth
									if buffer[position] != rune('u') {
										goto l170
									}
									position++
									goto l169
								l170:
									position, tokenIndex, depth = position169, tokenIndex169, depth169
									if buffer[position] != rune('U') {
										goto l160
									}
									position++
								}
							l169:
								{
									position171, tokenIndex171, depth171 := position, tokenIndex, depth
									if buffer[position] != rune('p') {
										goto l172
									}
									position++
									goto l171
								l172:
									position, tokenIndex, depth = position171, tokenIndex171, depth171
									if buffer[position] != rune('P') {
										goto l160
									}
									position++
								}
							l171:
								if !_rules[rule__]() {
									goto l160
								}
								{
									position173, tokenIndex173, depth173 := position, tokenIndex, depth
									if buffer[position] != rune('b') {
										goto l174
									}
									position++
									goto l173
								l174:
									position, tokenIndex, depth = position173, tokenIndex173, depth173
									if buffer[position] != rune('B') {
										goto l160
									}
									position++
								}
							l173:
								{
									position175, tokenIndex175, depth175 := position, tokenIndex, depth
									if buffer[position] != rune('y') {
										goto l176
									}
									position++
									goto l175
								l176:
									position, tokenIndex, depth = position175, tokenIndex175, depth175
									if buffer[position] != rune('Y') {
										goto l160
									}
									position++
								}
							l175:
								if !_rules[rule__]() {
									goto l160
								}
								{
									position177 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l160
									}
									depth--
									add(rulePegText, position177)
								}
								{
									add(ruleAction28, position)
								}
							l179:
								{
									position180, tokenIndex180, depth180 := position, tokenIndex, depth
									if !_rules[ruleCOMMA]() {
										goto l180
									}
									{
										position181 := position
										depth++
										if !_rules[ruleCOLUMN_NAME]() {
											goto l180
										}
										depth--
										add(rulePegText, position181)
									}
									{
										add(ruleAction29, position)
									}
									goto l179
								l180:
									position, tokenIndex, depth = position180, tokenIndex180, depth180
								}
								depth--
								add(rulegroupByClause, position162)
							}
							goto l161
						l160:
							position, tokenIndex, depth = position160, tokenIndex160, depth160
						}
					l161:
						if !_rules[rulePAREN_CLOSE]() {
							goto l155
						}
						{
							add(ruleAction24, position)
						}
						depth--
						add(ruleexpression_function, position156)
					}
					goto l154
				l155:
					position, tokenIndex, depth = position154, tokenIndex154, depth154
					{
						switch buffer[position] {
						case '"', '\'':
							if !_rules[ruleSTRING]() {
								goto l152
							}
							{
								add(ruleAction21, position)
							}
							break
						case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							{
								position186 := position
								depth++
								if !_rules[ruleNUMBER]() {
									goto l152
								}
								depth--
								add(rulePegText, position186)
							}
							{
								add(ruleAction20, position)
							}
							break
						case '\t', '\n', ' ', '(':
							if !_rules[rulePAREN_OPEN]() {
								goto l152
							}
							if !_rules[ruleexpression_1]() {
								goto l152
							}
							if !_rules[rulePAREN_CLOSE]() {
								goto l152
							}
							break
						default:
							{
								position188 := position
								depth++
								{
									position189 := position
									depth++
									if !_rules[ruleIDENTIFIER]() {
										goto l152
									}
									depth--
									add(rulePegText, position189)
								}
								{
									add(ruleAction25, position)
								}
								{
									position191, tokenIndex191, depth191 := position, tokenIndex, depth
									{
										position193, tokenIndex193, depth193 := position, tokenIndex, depth
										if buffer[position] != rune('[') {
											goto l194
										}
										position++
										if !_rules[rule_]() {
											goto l194
										}
										if !_rules[rulepredicate_1]() {
											goto l194
										}
										if !_rules[rule_]() {
											goto l194
										}
										if buffer[position] != rune(']') {
											goto l194
										}
										position++
										goto l193
									l194:
										position, tokenIndex, depth = position193, tokenIndex193, depth193
										{
											add(ruleAction26, position)
										}
									}
								l193:
									goto l192

									position, tokenIndex, depth = position191, tokenIndex191, depth191
								}
							l192:
								{
									add(ruleAction27, position)
								}
								depth--
								add(ruleexpression_metric, position188)
							}
							break
						}
					}

				}
			l154:
				depth--
				add(ruleexpression_3, position153)
			}
			return true
		l152:
			position, tokenIndex, depth = position152, tokenIndex152, depth152
			return false
		},
		/* 12 expression_function <- <(<IDENTIFIER> Action22 PAREN_OPEN expressionList Action23 (__ groupByClause)? PAREN_CLOSE Action24)> */
		nil,
		/* 13 expression_metric <- <(<IDENTIFIER> Action25 (('[' _ predicate_1 _ ']') / Action26)? Action27)> */
		nil,
		/* 14 groupByClause <- <(('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P') __ (('b' / 'B') ('y' / 'Y')) __ <COLUMN_NAME> Action28 (COMMA <COLUMN_NAME> Action29)*)> */
		nil,
		/* 15 predicateClause <- <(('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E') __ predicate_1)> */
		nil,
		/* 16 predicate_1 <- <((predicate_2 OP_OR predicate_1 Action30) / predicate_2 / )> */
		func() bool {
			{
				position202 := position
				depth++
				{
					position203, tokenIndex203, depth203 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l204
					}
					{
						position205 := position
						depth++
						if !_rules[rule_]() {
							goto l204
						}
						{
							position206, tokenIndex206, depth206 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l207
							}
							position++
							goto l206
						l207:
							position, tokenIndex, depth = position206, tokenIndex206, depth206
							if buffer[position] != rune('O') {
								goto l204
							}
							position++
						}
					l206:
						{
							position208, tokenIndex208, depth208 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l209
							}
							position++
							goto l208
						l209:
							position, tokenIndex, depth = position208, tokenIndex208, depth208
							if buffer[position] != rune('R') {
								goto l204
							}
							position++
						}
					l208:
						if !_rules[rule_]() {
							goto l204
						}
						depth--
						add(ruleOP_OR, position205)
					}
					if !_rules[rulepredicate_1]() {
						goto l204
					}
					{
						add(ruleAction30, position)
					}
					goto l203
				l204:
					position, tokenIndex, depth = position203, tokenIndex203, depth203
					if !_rules[rulepredicate_2]() {
						goto l211
					}
					goto l203
				l211:
					position, tokenIndex, depth = position203, tokenIndex203, depth203
				}
			l203:
				depth--
				add(rulepredicate_1, position202)
			}
			return true
		},
		/* 17 predicate_2 <- <((predicate_3 OP_AND predicate_2 Action31) / predicate_3)> */
		func() bool {
			position212, tokenIndex212, depth212 := position, tokenIndex, depth
			{
				position213 := position
				depth++
				{
					position214, tokenIndex214, depth214 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l215
					}
					{
						position216 := position
						depth++
						if !_rules[rule_]() {
							goto l215
						}
						{
							position217, tokenIndex217, depth217 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l218
							}
							position++
							goto l217
						l218:
							position, tokenIndex, depth = position217, tokenIndex217, depth217
							if buffer[position] != rune('A') {
								goto l215
							}
							position++
						}
					l217:
						{
							position219, tokenIndex219, depth219 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l220
							}
							position++
							goto l219
						l220:
							position, tokenIndex, depth = position219, tokenIndex219, depth219
							if buffer[position] != rune('N') {
								goto l215
							}
							position++
						}
					l219:
						{
							position221, tokenIndex221, depth221 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l222
							}
							position++
							goto l221
						l222:
							position, tokenIndex, depth = position221, tokenIndex221, depth221
							if buffer[position] != rune('D') {
								goto l215
							}
							position++
						}
					l221:
						if !_rules[rule_]() {
							goto l215
						}
						depth--
						add(ruleOP_AND, position216)
					}
					if !_rules[rulepredicate_2]() {
						goto l215
					}
					{
						add(ruleAction31, position)
					}
					goto l214
				l215:
					position, tokenIndex, depth = position214, tokenIndex214, depth214
					if !_rules[rulepredicate_3]() {
						goto l212
					}
				}
			l214:
				depth--
				add(rulepredicate_2, position213)
			}
			return true
		l212:
			position, tokenIndex, depth = position212, tokenIndex212, depth212
			return false
		},
		/* 18 predicate_3 <- <((OP_NOT predicate_3 Action32) / (PAREN_OPEN predicate_1 PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position224, tokenIndex224, depth224 := position, tokenIndex, depth
			{
				position225 := position
				depth++
				{
					position226, tokenIndex226, depth226 := position, tokenIndex, depth
					{
						position228 := position
						depth++
						{
							position229, tokenIndex229, depth229 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l230
							}
							position++
							goto l229
						l230:
							position, tokenIndex, depth = position229, tokenIndex229, depth229
							if buffer[position] != rune('N') {
								goto l227
							}
							position++
						}
					l229:
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
								goto l227
							}
							position++
						}
					l231:
						{
							position233, tokenIndex233, depth233 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l234
							}
							position++
							goto l233
						l234:
							position, tokenIndex, depth = position233, tokenIndex233, depth233
							if buffer[position] != rune('T') {
								goto l227
							}
							position++
						}
					l233:
						if !_rules[rule__]() {
							goto l227
						}
						depth--
						add(ruleOP_NOT, position228)
					}
					if !_rules[rulepredicate_3]() {
						goto l227
					}
					{
						add(ruleAction32, position)
					}
					goto l226
				l227:
					position, tokenIndex, depth = position226, tokenIndex226, depth226
					if !_rules[rulePAREN_OPEN]() {
						goto l236
					}
					if !_rules[rulepredicate_1]() {
						goto l236
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l236
					}
					goto l226
				l236:
					position, tokenIndex, depth = position226, tokenIndex226, depth226
					{
						position237 := position
						depth++
						{
							position238, tokenIndex238, depth238 := position, tokenIndex, depth
							if !_rules[ruletagName]() {
								goto l239
							}
							if !_rules[rule_]() {
								goto l239
							}
							if buffer[position] != rune('=') {
								goto l239
							}
							position++
							if !_rules[rule_]() {
								goto l239
							}
							if !_rules[ruleliteralString]() {
								goto l239
							}
							{
								add(ruleAction33, position)
							}
							goto l238
						l239:
							position, tokenIndex, depth = position238, tokenIndex238, depth238
							if !_rules[ruletagName]() {
								goto l241
							}
							if !_rules[rule_]() {
								goto l241
							}
							if buffer[position] != rune('!') {
								goto l241
							}
							position++
							if buffer[position] != rune('=') {
								goto l241
							}
							position++
							if !_rules[rule_]() {
								goto l241
							}
							if !_rules[ruleliteralString]() {
								goto l241
							}
							{
								add(ruleAction34, position)
							}
							goto l238
						l241:
							position, tokenIndex, depth = position238, tokenIndex238, depth238
							if !_rules[ruletagName]() {
								goto l243
							}
							if !_rules[rule__]() {
								goto l243
							}
							{
								position244, tokenIndex244, depth244 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l245
								}
								position++
								goto l244
							l245:
								position, tokenIndex, depth = position244, tokenIndex244, depth244
								if buffer[position] != rune('M') {
									goto l243
								}
								position++
							}
						l244:
							{
								position246, tokenIndex246, depth246 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l247
								}
								position++
								goto l246
							l247:
								position, tokenIndex, depth = position246, tokenIndex246, depth246
								if buffer[position] != rune('A') {
									goto l243
								}
								position++
							}
						l246:
							{
								position248, tokenIndex248, depth248 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l249
								}
								position++
								goto l248
							l249:
								position, tokenIndex, depth = position248, tokenIndex248, depth248
								if buffer[position] != rune('T') {
									goto l243
								}
								position++
							}
						l248:
							{
								position250, tokenIndex250, depth250 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l251
								}
								position++
								goto l250
							l251:
								position, tokenIndex, depth = position250, tokenIndex250, depth250
								if buffer[position] != rune('C') {
									goto l243
								}
								position++
							}
						l250:
							{
								position252, tokenIndex252, depth252 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l253
								}
								position++
								goto l252
							l253:
								position, tokenIndex, depth = position252, tokenIndex252, depth252
								if buffer[position] != rune('H') {
									goto l243
								}
								position++
							}
						l252:
							{
								position254, tokenIndex254, depth254 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l255
								}
								position++
								goto l254
							l255:
								position, tokenIndex, depth = position254, tokenIndex254, depth254
								if buffer[position] != rune('E') {
									goto l243
								}
								position++
							}
						l254:
							{
								position256, tokenIndex256, depth256 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l257
								}
								position++
								goto l256
							l257:
								position, tokenIndex, depth = position256, tokenIndex256, depth256
								if buffer[position] != rune('S') {
									goto l243
								}
								position++
							}
						l256:
							if !_rules[rule__]() {
								goto l243
							}
							if !_rules[ruleliteralString]() {
								goto l243
							}
							{
								add(ruleAction35, position)
							}
							goto l238
						l243:
							position, tokenIndex, depth = position238, tokenIndex238, depth238
							if !_rules[ruletagName]() {
								goto l224
							}
							if !_rules[rule__]() {
								goto l224
							}
							{
								position259, tokenIndex259, depth259 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l260
								}
								position++
								goto l259
							l260:
								position, tokenIndex, depth = position259, tokenIndex259, depth259
								if buffer[position] != rune('I') {
									goto l224
								}
								position++
							}
						l259:
							{
								position261, tokenIndex261, depth261 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l262
								}
								position++
								goto l261
							l262:
								position, tokenIndex, depth = position261, tokenIndex261, depth261
								if buffer[position] != rune('N') {
									goto l224
								}
								position++
							}
						l261:
							if !_rules[rule__]() {
								goto l224
							}
							{
								position263 := position
								depth++
								{
									add(ruleAction38, position)
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l224
								}
								if !_rules[ruleliteralListString]() {
									goto l224
								}
							l265:
								{
									position266, tokenIndex266, depth266 := position, tokenIndex, depth
									if !_rules[ruleCOMMA]() {
										goto l266
									}
									if !_rules[ruleliteralListString]() {
										goto l266
									}
									goto l265
								l266:
									position, tokenIndex, depth = position266, tokenIndex266, depth266
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l224
								}
								depth--
								add(ruleliteralList, position263)
							}
							{
								add(ruleAction36, position)
							}
						}
					l238:
						depth--
						add(ruletagMatcher, position237)
					}
				}
			l226:
				depth--
				add(rulepredicate_3, position225)
			}
			return true
		l224:
			position, tokenIndex, depth = position224, tokenIndex224, depth224
			return false
		},
		/* 19 tagMatcher <- <((tagName _ '=' _ literalString Action33) / (tagName _ ('!' '=') _ literalString Action34) / (tagName __ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) __ literalString Action35) / (tagName __ (('i' / 'I') ('n' / 'N')) __ literalList Action36))> */
		nil,
		/* 20 literalString <- <(STRING Action37)> */
		func() bool {
			position269, tokenIndex269, depth269 := position, tokenIndex, depth
			{
				position270 := position
				depth++
				if !_rules[ruleSTRING]() {
					goto l269
				}
				{
					add(ruleAction37, position)
				}
				depth--
				add(ruleliteralString, position270)
			}
			return true
		l269:
			position, tokenIndex, depth = position269, tokenIndex269, depth269
			return false
		},
		/* 21 literalList <- <(Action38 PAREN_OPEN literalListString (COMMA literalListString)* PAREN_CLOSE)> */
		nil,
		/* 22 literalListString <- <(STRING Action39)> */
		func() bool {
			position273, tokenIndex273, depth273 := position, tokenIndex, depth
			{
				position274 := position
				depth++
				if !_rules[ruleSTRING]() {
					goto l273
				}
				{
					add(ruleAction39, position)
				}
				depth--
				add(ruleliteralListString, position274)
			}
			return true
		l273:
			position, tokenIndex, depth = position273, tokenIndex273, depth273
			return false
		},
		/* 23 tagName <- <(<TAG_NAME> Action40)> */
		func() bool {
			position276, tokenIndex276, depth276 := position, tokenIndex, depth
			{
				position277 := position
				depth++
				{
					position278 := position
					depth++
					{
						position279 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l276
						}
						depth--
						add(ruleTAG_NAME, position279)
					}
					depth--
					add(rulePegText, position278)
				}
				{
					add(ruleAction40, position)
				}
				depth--
				add(ruletagName, position277)
			}
			return true
		l276:
			position, tokenIndex, depth = position276, tokenIndex276, depth276
			return false
		},
		/* 24 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position281, tokenIndex281, depth281 := position, tokenIndex, depth
			{
				position282 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l281
				}
				depth--
				add(ruleCOLUMN_NAME, position282)
			}
			return true
		l281:
			position, tokenIndex, depth = position281, tokenIndex281, depth281
			return false
		},
		/* 25 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 26 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 27 IDENTIFIER <- <(('`' CHAR* '`') / (!(KEYWORD !ID_CONT) ID_SEGMENT ('.' ID_SEGMENT)*))> */
		func() bool {
			position285, tokenIndex285, depth285 := position, tokenIndex, depth
			{
				position286 := position
				depth++
				{
					position287, tokenIndex287, depth287 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l288
					}
					position++
				l289:
					{
						position290, tokenIndex290, depth290 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l290
						}
						goto l289
					l290:
						position, tokenIndex, depth = position290, tokenIndex290, depth290
					}
					if buffer[position] != rune('`') {
						goto l288
					}
					position++
					goto l287
				l288:
					position, tokenIndex, depth = position287, tokenIndex287, depth287
					{
						position291, tokenIndex291, depth291 := position, tokenIndex, depth
						{
							position292 := position
							depth++
							{
								position293, tokenIndex293, depth293 := position, tokenIndex, depth
								{
									position295, tokenIndex295, depth295 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l296
									}
									position++
									goto l295
								l296:
									position, tokenIndex, depth = position295, tokenIndex295, depth295
									if buffer[position] != rune('A') {
										goto l294
									}
									position++
								}
							l295:
								{
									position297, tokenIndex297, depth297 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l298
									}
									position++
									goto l297
								l298:
									position, tokenIndex, depth = position297, tokenIndex297, depth297
									if buffer[position] != rune('L') {
										goto l294
									}
									position++
								}
							l297:
								{
									position299, tokenIndex299, depth299 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l300
									}
									position++
									goto l299
								l300:
									position, tokenIndex, depth = position299, tokenIndex299, depth299
									if buffer[position] != rune('L') {
										goto l294
									}
									position++
								}
							l299:
								goto l293
							l294:
								position, tokenIndex, depth = position293, tokenIndex293, depth293
								{
									position302, tokenIndex302, depth302 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l303
									}
									position++
									goto l302
								l303:
									position, tokenIndex, depth = position302, tokenIndex302, depth302
									if buffer[position] != rune('A') {
										goto l301
									}
									position++
								}
							l302:
								{
									position304, tokenIndex304, depth304 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l305
									}
									position++
									goto l304
								l305:
									position, tokenIndex, depth = position304, tokenIndex304, depth304
									if buffer[position] != rune('N') {
										goto l301
									}
									position++
								}
							l304:
								{
									position306, tokenIndex306, depth306 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l307
									}
									position++
									goto l306
								l307:
									position, tokenIndex, depth = position306, tokenIndex306, depth306
									if buffer[position] != rune('D') {
										goto l301
									}
									position++
								}
							l306:
								goto l293
							l301:
								position, tokenIndex, depth = position293, tokenIndex293, depth293
								{
									position309, tokenIndex309, depth309 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l310
									}
									position++
									goto l309
								l310:
									position, tokenIndex, depth = position309, tokenIndex309, depth309
									if buffer[position] != rune('M') {
										goto l308
									}
									position++
								}
							l309:
								{
									position311, tokenIndex311, depth311 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l312
									}
									position++
									goto l311
								l312:
									position, tokenIndex, depth = position311, tokenIndex311, depth311
									if buffer[position] != rune('A') {
										goto l308
									}
									position++
								}
							l311:
								{
									position313, tokenIndex313, depth313 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l314
									}
									position++
									goto l313
								l314:
									position, tokenIndex, depth = position313, tokenIndex313, depth313
									if buffer[position] != rune('T') {
										goto l308
									}
									position++
								}
							l313:
								{
									position315, tokenIndex315, depth315 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l316
									}
									position++
									goto l315
								l316:
									position, tokenIndex, depth = position315, tokenIndex315, depth315
									if buffer[position] != rune('C') {
										goto l308
									}
									position++
								}
							l315:
								{
									position317, tokenIndex317, depth317 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l318
									}
									position++
									goto l317
								l318:
									position, tokenIndex, depth = position317, tokenIndex317, depth317
									if buffer[position] != rune('H') {
										goto l308
									}
									position++
								}
							l317:
								{
									position319, tokenIndex319, depth319 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l320
									}
									position++
									goto l319
								l320:
									position, tokenIndex, depth = position319, tokenIndex319, depth319
									if buffer[position] != rune('E') {
										goto l308
									}
									position++
								}
							l319:
								{
									position321, tokenIndex321, depth321 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l322
									}
									position++
									goto l321
								l322:
									position, tokenIndex, depth = position321, tokenIndex321, depth321
									if buffer[position] != rune('S') {
										goto l308
									}
									position++
								}
							l321:
								goto l293
							l308:
								position, tokenIndex, depth = position293, tokenIndex293, depth293
								{
									position324, tokenIndex324, depth324 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l325
									}
									position++
									goto l324
								l325:
									position, tokenIndex, depth = position324, tokenIndex324, depth324
									if buffer[position] != rune('S') {
										goto l323
									}
									position++
								}
							l324:
								{
									position326, tokenIndex326, depth326 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l327
									}
									position++
									goto l326
								l327:
									position, tokenIndex, depth = position326, tokenIndex326, depth326
									if buffer[position] != rune('E') {
										goto l323
									}
									position++
								}
							l326:
								{
									position328, tokenIndex328, depth328 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l329
									}
									position++
									goto l328
								l329:
									position, tokenIndex, depth = position328, tokenIndex328, depth328
									if buffer[position] != rune('L') {
										goto l323
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
										goto l323
									}
									position++
								}
							l330:
								{
									position332, tokenIndex332, depth332 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l333
									}
									position++
									goto l332
								l333:
									position, tokenIndex, depth = position332, tokenIndex332, depth332
									if buffer[position] != rune('C') {
										goto l323
									}
									position++
								}
							l332:
								{
									position334, tokenIndex334, depth334 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l335
									}
									position++
									goto l334
								l335:
									position, tokenIndex, depth = position334, tokenIndex334, depth334
									if buffer[position] != rune('T') {
										goto l323
									}
									position++
								}
							l334:
								goto l293
							l323:
								position, tokenIndex, depth = position293, tokenIndex293, depth293
								{
									switch buffer[position] {
									case 'M', 'm':
										{
											position337, tokenIndex337, depth337 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l338
											}
											position++
											goto l337
										l338:
											position, tokenIndex, depth = position337, tokenIndex337, depth337
											if buffer[position] != rune('M') {
												goto l291
											}
											position++
										}
									l337:
										{
											position339, tokenIndex339, depth339 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l340
											}
											position++
											goto l339
										l340:
											position, tokenIndex, depth = position339, tokenIndex339, depth339
											if buffer[position] != rune('E') {
												goto l291
											}
											position++
										}
									l339:
										{
											position341, tokenIndex341, depth341 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l342
											}
											position++
											goto l341
										l342:
											position, tokenIndex, depth = position341, tokenIndex341, depth341
											if buffer[position] != rune('T') {
												goto l291
											}
											position++
										}
									l341:
										{
											position343, tokenIndex343, depth343 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l344
											}
											position++
											goto l343
										l344:
											position, tokenIndex, depth = position343, tokenIndex343, depth343
											if buffer[position] != rune('R') {
												goto l291
											}
											position++
										}
									l343:
										{
											position345, tokenIndex345, depth345 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l346
											}
											position++
											goto l345
										l346:
											position, tokenIndex, depth = position345, tokenIndex345, depth345
											if buffer[position] != rune('I') {
												goto l291
											}
											position++
										}
									l345:
										{
											position347, tokenIndex347, depth347 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l348
											}
											position++
											goto l347
										l348:
											position, tokenIndex, depth = position347, tokenIndex347, depth347
											if buffer[position] != rune('C') {
												goto l291
											}
											position++
										}
									l347:
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
												goto l291
											}
											position++
										}
									l349:
										break
									case 'W', 'w':
										{
											position351, tokenIndex351, depth351 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l352
											}
											position++
											goto l351
										l352:
											position, tokenIndex, depth = position351, tokenIndex351, depth351
											if buffer[position] != rune('W') {
												goto l291
											}
											position++
										}
									l351:
										{
											position353, tokenIndex353, depth353 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l354
											}
											position++
											goto l353
										l354:
											position, tokenIndex, depth = position353, tokenIndex353, depth353
											if buffer[position] != rune('H') {
												goto l291
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
												goto l291
											}
											position++
										}
									l355:
										{
											position357, tokenIndex357, depth357 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l358
											}
											position++
											goto l357
										l358:
											position, tokenIndex, depth = position357, tokenIndex357, depth357
											if buffer[position] != rune('R') {
												goto l291
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
												goto l291
											}
											position++
										}
									l359:
										break
									case 'O', 'o':
										{
											position361, tokenIndex361, depth361 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l362
											}
											position++
											goto l361
										l362:
											position, tokenIndex, depth = position361, tokenIndex361, depth361
											if buffer[position] != rune('O') {
												goto l291
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
												goto l291
											}
											position++
										}
									l363:
										break
									case 'N', 'n':
										{
											position365, tokenIndex365, depth365 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l366
											}
											position++
											goto l365
										l366:
											position, tokenIndex, depth = position365, tokenIndex365, depth365
											if buffer[position] != rune('N') {
												goto l291
											}
											position++
										}
									l365:
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
												goto l291
											}
											position++
										}
									l367:
										{
											position369, tokenIndex369, depth369 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l370
											}
											position++
											goto l369
										l370:
											position, tokenIndex, depth = position369, tokenIndex369, depth369
											if buffer[position] != rune('T') {
												goto l291
											}
											position++
										}
									l369:
										break
									case 'I', 'i':
										{
											position371, tokenIndex371, depth371 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l372
											}
											position++
											goto l371
										l372:
											position, tokenIndex, depth = position371, tokenIndex371, depth371
											if buffer[position] != rune('I') {
												goto l291
											}
											position++
										}
									l371:
										{
											position373, tokenIndex373, depth373 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l374
											}
											position++
											goto l373
										l374:
											position, tokenIndex, depth = position373, tokenIndex373, depth373
											if buffer[position] != rune('N') {
												goto l291
											}
											position++
										}
									l373:
										break
									case 'G', 'g':
										{
											position375, tokenIndex375, depth375 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l376
											}
											position++
											goto l375
										l376:
											position, tokenIndex, depth = position375, tokenIndex375, depth375
											if buffer[position] != rune('G') {
												goto l291
											}
											position++
										}
									l375:
										{
											position377, tokenIndex377, depth377 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l378
											}
											position++
											goto l377
										l378:
											position, tokenIndex, depth = position377, tokenIndex377, depth377
											if buffer[position] != rune('R') {
												goto l291
											}
											position++
										}
									l377:
										{
											position379, tokenIndex379, depth379 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l380
											}
											position++
											goto l379
										l380:
											position, tokenIndex, depth = position379, tokenIndex379, depth379
											if buffer[position] != rune('O') {
												goto l291
											}
											position++
										}
									l379:
										{
											position381, tokenIndex381, depth381 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l382
											}
											position++
											goto l381
										l382:
											position, tokenIndex, depth = position381, tokenIndex381, depth381
											if buffer[position] != rune('U') {
												goto l291
											}
											position++
										}
									l381:
										{
											position383, tokenIndex383, depth383 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l384
											}
											position++
											goto l383
										l384:
											position, tokenIndex, depth = position383, tokenIndex383, depth383
											if buffer[position] != rune('P') {
												goto l291
											}
											position++
										}
									l383:
										break
									case 'D', 'd':
										{
											position385, tokenIndex385, depth385 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l386
											}
											position++
											goto l385
										l386:
											position, tokenIndex, depth = position385, tokenIndex385, depth385
											if buffer[position] != rune('D') {
												goto l291
											}
											position++
										}
									l385:
										{
											position387, tokenIndex387, depth387 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l388
											}
											position++
											goto l387
										l388:
											position, tokenIndex, depth = position387, tokenIndex387, depth387
											if buffer[position] != rune('E') {
												goto l291
											}
											position++
										}
									l387:
										{
											position389, tokenIndex389, depth389 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l390
											}
											position++
											goto l389
										l390:
											position, tokenIndex, depth = position389, tokenIndex389, depth389
											if buffer[position] != rune('S') {
												goto l291
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
												goto l291
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
												goto l291
											}
											position++
										}
									l393:
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
												goto l291
											}
											position++
										}
									l395:
										{
											position397, tokenIndex397, depth397 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l398
											}
											position++
											goto l397
										l398:
											position, tokenIndex, depth = position397, tokenIndex397, depth397
											if buffer[position] != rune('B') {
												goto l291
											}
											position++
										}
									l397:
										{
											position399, tokenIndex399, depth399 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l400
											}
											position++
											goto l399
										l400:
											position, tokenIndex, depth = position399, tokenIndex399, depth399
											if buffer[position] != rune('E') {
												goto l291
											}
											position++
										}
									l399:
										break
									case 'B', 'b':
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
												goto l291
											}
											position++
										}
									l401:
										{
											position403, tokenIndex403, depth403 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l404
											}
											position++
											goto l403
										l404:
											position, tokenIndex, depth = position403, tokenIndex403, depth403
											if buffer[position] != rune('Y') {
												goto l291
											}
											position++
										}
									l403:
										break
									case 'A', 'a':
										{
											position405, tokenIndex405, depth405 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l406
											}
											position++
											goto l405
										l406:
											position, tokenIndex, depth = position405, tokenIndex405, depth405
											if buffer[position] != rune('A') {
												goto l291
											}
											position++
										}
									l405:
										{
											position407, tokenIndex407, depth407 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l408
											}
											position++
											goto l407
										l408:
											position, tokenIndex, depth = position407, tokenIndex407, depth407
											if buffer[position] != rune('S') {
												goto l291
											}
											position++
										}
									l407:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l291
										}
										break
									}
								}

							}
						l293:
							depth--
							add(ruleKEYWORD, position292)
						}
						{
							position409, tokenIndex409, depth409 := position, tokenIndex, depth
							if !_rules[ruleID_CONT]() {
								goto l409
							}
							goto l291
						l409:
							position, tokenIndex, depth = position409, tokenIndex409, depth409
						}
						goto l285
					l291:
						position, tokenIndex, depth = position291, tokenIndex291, depth291
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l285
					}
				l410:
					{
						position411, tokenIndex411, depth411 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l411
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l411
						}
						goto l410
					l411:
						position, tokenIndex, depth = position411, tokenIndex411, depth411
					}
				}
			l287:
				depth--
				add(ruleIDENTIFIER, position286)
			}
			return true
		l285:
			position, tokenIndex, depth = position285, tokenIndex285, depth285
			return false
		},
		/* 28 TIMESTAMP <- <((&('N' | 'n') <(('n' / 'N') ('o' / 'O') ('w' / 'W'))>) | (&('"' | '\'') STRING) | (&('-' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') <(NUMBER ([a-z] / [A-Z])?)>))> */
		nil,
		/* 29 ID_SEGMENT <- <(ID_START ID_CONT*)> */
		func() bool {
			position413, tokenIndex413, depth413 := position, tokenIndex, depth
			{
				position414 := position
				depth++
				if !_rules[ruleID_START]() {
					goto l413
				}
			l415:
				{
					position416, tokenIndex416, depth416 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l416
					}
					goto l415
				l416:
					position, tokenIndex, depth = position416, tokenIndex416, depth416
				}
				depth--
				add(ruleID_SEGMENT, position414)
			}
			return true
		l413:
			position, tokenIndex, depth = position413, tokenIndex413, depth413
			return false
		},
		/* 30 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position417, tokenIndex417, depth417 := position, tokenIndex, depth
			{
				position418 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l417
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l417
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l417
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position418)
			}
			return true
		l417:
			position, tokenIndex, depth = position417, tokenIndex417, depth417
			return false
		},
		/* 31 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position420, tokenIndex420, depth420 := position, tokenIndex, depth
			{
				position421 := position
				depth++
				{
					position422, tokenIndex422, depth422 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l423
					}
					goto l422
				l423:
					position, tokenIndex, depth = position422, tokenIndex422, depth422
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l420
					}
					position++
				}
			l422:
				depth--
				add(ruleID_CONT, position421)
			}
			return true
		l420:
			position, tokenIndex, depth = position420, tokenIndex420, depth420
			return false
		},
		/* 32 PROPERTY_KEY <- <((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> __ (('b' / 'B') ('y' / 'Y')))) | (&('R' | 'r') <(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))>) | (&('T' | 't') <(('t' / 'T') ('o' / 'O'))>) | (&('F' | 'f') <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))>))> */
		func() bool {
			position424, tokenIndex424, depth424 := position, tokenIndex, depth
			{
				position425 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position427 := position
							depth++
							{
								position428, tokenIndex428, depth428 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l429
								}
								position++
								goto l428
							l429:
								position, tokenIndex, depth = position428, tokenIndex428, depth428
								if buffer[position] != rune('S') {
									goto l424
								}
								position++
							}
						l428:
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
									goto l424
								}
								position++
							}
						l430:
							{
								position432, tokenIndex432, depth432 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l433
								}
								position++
								goto l432
							l433:
								position, tokenIndex, depth = position432, tokenIndex432, depth432
								if buffer[position] != rune('M') {
									goto l424
								}
								position++
							}
						l432:
							{
								position434, tokenIndex434, depth434 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l435
								}
								position++
								goto l434
							l435:
								position, tokenIndex, depth = position434, tokenIndex434, depth434
								if buffer[position] != rune('P') {
									goto l424
								}
								position++
							}
						l434:
							{
								position436, tokenIndex436, depth436 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l437
								}
								position++
								goto l436
							l437:
								position, tokenIndex, depth = position436, tokenIndex436, depth436
								if buffer[position] != rune('L') {
									goto l424
								}
								position++
							}
						l436:
							{
								position438, tokenIndex438, depth438 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l439
								}
								position++
								goto l438
							l439:
								position, tokenIndex, depth = position438, tokenIndex438, depth438
								if buffer[position] != rune('E') {
									goto l424
								}
								position++
							}
						l438:
							depth--
							add(rulePegText, position427)
						}
						if !_rules[rule__]() {
							goto l424
						}
						{
							position440, tokenIndex440, depth440 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l441
							}
							position++
							goto l440
						l441:
							position, tokenIndex, depth = position440, tokenIndex440, depth440
							if buffer[position] != rune('B') {
								goto l424
							}
							position++
						}
					l440:
						{
							position442, tokenIndex442, depth442 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l443
							}
							position++
							goto l442
						l443:
							position, tokenIndex, depth = position442, tokenIndex442, depth442
							if buffer[position] != rune('Y') {
								goto l424
							}
							position++
						}
					l442:
						break
					case 'R', 'r':
						{
							position444 := position
							depth++
							{
								position445, tokenIndex445, depth445 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l446
								}
								position++
								goto l445
							l446:
								position, tokenIndex, depth = position445, tokenIndex445, depth445
								if buffer[position] != rune('R') {
									goto l424
								}
								position++
							}
						l445:
							{
								position447, tokenIndex447, depth447 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l448
								}
								position++
								goto l447
							l448:
								position, tokenIndex, depth = position447, tokenIndex447, depth447
								if buffer[position] != rune('E') {
									goto l424
								}
								position++
							}
						l447:
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
									goto l424
								}
								position++
							}
						l449:
							{
								position451, tokenIndex451, depth451 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l452
								}
								position++
								goto l451
							l452:
								position, tokenIndex, depth = position451, tokenIndex451, depth451
								if buffer[position] != rune('O') {
									goto l424
								}
								position++
							}
						l451:
							{
								position453, tokenIndex453, depth453 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l454
								}
								position++
								goto l453
							l454:
								position, tokenIndex, depth = position453, tokenIndex453, depth453
								if buffer[position] != rune('L') {
									goto l424
								}
								position++
							}
						l453:
							{
								position455, tokenIndex455, depth455 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l456
								}
								position++
								goto l455
							l456:
								position, tokenIndex, depth = position455, tokenIndex455, depth455
								if buffer[position] != rune('U') {
									goto l424
								}
								position++
							}
						l455:
							{
								position457, tokenIndex457, depth457 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l458
								}
								position++
								goto l457
							l458:
								position, tokenIndex, depth = position457, tokenIndex457, depth457
								if buffer[position] != rune('T') {
									goto l424
								}
								position++
							}
						l457:
							{
								position459, tokenIndex459, depth459 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l460
								}
								position++
								goto l459
							l460:
								position, tokenIndex, depth = position459, tokenIndex459, depth459
								if buffer[position] != rune('I') {
									goto l424
								}
								position++
							}
						l459:
							{
								position461, tokenIndex461, depth461 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l462
								}
								position++
								goto l461
							l462:
								position, tokenIndex, depth = position461, tokenIndex461, depth461
								if buffer[position] != rune('O') {
									goto l424
								}
								position++
							}
						l461:
							{
								position463, tokenIndex463, depth463 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l464
								}
								position++
								goto l463
							l464:
								position, tokenIndex, depth = position463, tokenIndex463, depth463
								if buffer[position] != rune('N') {
									goto l424
								}
								position++
							}
						l463:
							depth--
							add(rulePegText, position444)
						}
						break
					case 'T', 't':
						{
							position465 := position
							depth++
							{
								position466, tokenIndex466, depth466 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l467
								}
								position++
								goto l466
							l467:
								position, tokenIndex, depth = position466, tokenIndex466, depth466
								if buffer[position] != rune('T') {
									goto l424
								}
								position++
							}
						l466:
							{
								position468, tokenIndex468, depth468 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l469
								}
								position++
								goto l468
							l469:
								position, tokenIndex, depth = position468, tokenIndex468, depth468
								if buffer[position] != rune('O') {
									goto l424
								}
								position++
							}
						l468:
							depth--
							add(rulePegText, position465)
						}
						break
					default:
						{
							position470 := position
							depth++
							{
								position471, tokenIndex471, depth471 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l472
								}
								position++
								goto l471
							l472:
								position, tokenIndex, depth = position471, tokenIndex471, depth471
								if buffer[position] != rune('F') {
									goto l424
								}
								position++
							}
						l471:
							{
								position473, tokenIndex473, depth473 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l474
								}
								position++
								goto l473
							l474:
								position, tokenIndex, depth = position473, tokenIndex473, depth473
								if buffer[position] != rune('R') {
									goto l424
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
									goto l424
								}
								position++
							}
						l475:
							{
								position477, tokenIndex477, depth477 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l478
								}
								position++
								goto l477
							l478:
								position, tokenIndex, depth = position477, tokenIndex477, depth477
								if buffer[position] != rune('M') {
									goto l424
								}
								position++
							}
						l477:
							depth--
							add(rulePegText, position470)
						}
						break
					}
				}

				depth--
				add(rulePROPERTY_KEY, position425)
			}
			return true
		l424:
			position, tokenIndex, depth = position424, tokenIndex424, depth424
			return false
		},
		/* 33 PROPERTY_VALUE <- <TIMESTAMP> */
		nil,
		/* 34 KEYWORD <- <((('a' / 'A') ('l' / 'L') ('l' / 'L')) / (('a' / 'A') ('n' / 'N') ('d' / 'D')) / (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) / (('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T')) / ((&('M' | 'm') (('m' / 'M') ('e' / 'E') ('t' / 'T') ('r' / 'R') ('i' / 'I') ('c' / 'C') ('s' / 'S'))) | (&('W' | 'w') (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E'))) | (&('O' | 'o') (('o' / 'O') ('r' / 'R'))) | (&('N' | 'n') (('n' / 'N') ('o' / 'O') ('t' / 'T'))) | (&('I' | 'i') (('i' / 'I') ('n' / 'N'))) | (&('G' | 'g') (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P'))) | (&('D' | 'd') (('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E'))) | (&('B' | 'b') (('b' / 'B') ('y' / 'Y'))) | (&('A' | 'a') (('a' / 'A') ('s' / 'S'))) | (&('F' | 'R' | 'S' | 'T' | 'f' | 'r' | 's' | 't') PROPERTY_KEY)))> */
		nil,
		/* 35 OP_ADD <- <(_ '+' _)> */
		nil,
		/* 36 OP_SUB <- <(_ '-' _)> */
		nil,
		/* 37 OP_MULT <- <(_ '*' _)> */
		nil,
		/* 38 OP_DIV <- <(_ '/' _)> */
		nil,
		/* 39 OP_AND <- <(_ (('a' / 'A') ('n' / 'N') ('d' / 'D')) _)> */
		nil,
		/* 40 OP_OR <- <(_ (('o' / 'O') ('r' / 'R')) _)> */
		nil,
		/* 41 OP_NOT <- <(('n' / 'N') ('o' / 'O') ('t' / 'T') __)> */
		nil,
		/* 42 QUOTE_SINGLE <- <'\''> */
		func() bool {
			position488, tokenIndex488, depth488 := position, tokenIndex, depth
			{
				position489 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l488
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position489)
			}
			return true
		l488:
			position, tokenIndex, depth = position488, tokenIndex488, depth488
			return false
		},
		/* 43 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position490, tokenIndex490, depth490 := position, tokenIndex, depth
			{
				position491 := position
				depth++
				if buffer[position] != rune('"') {
					goto l490
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position491)
			}
			return true
		l490:
			position, tokenIndex, depth = position490, tokenIndex490, depth490
			return false
		},
		/* 44 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> QUOTE_SINGLE) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> QUOTE_DOUBLE))> */
		func() bool {
			position492, tokenIndex492, depth492 := position, tokenIndex, depth
			{
				position493 := position
				depth++
				{
					position494, tokenIndex494, depth494 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l495
					}
					{
						position496 := position
						depth++
					l497:
						{
							position498, tokenIndex498, depth498 := position, tokenIndex, depth
							{
								position499, tokenIndex499, depth499 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l499
								}
								goto l498
							l499:
								position, tokenIndex, depth = position499, tokenIndex499, depth499
							}
							if !_rules[ruleCHAR]() {
								goto l498
							}
							goto l497
						l498:
							position, tokenIndex, depth = position498, tokenIndex498, depth498
						}
						depth--
						add(rulePegText, position496)
					}
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l495
					}
					goto l494
				l495:
					position, tokenIndex, depth = position494, tokenIndex494, depth494
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l492
					}
					{
						position500 := position
						depth++
					l501:
						{
							position502, tokenIndex502, depth502 := position, tokenIndex, depth
							{
								position503, tokenIndex503, depth503 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l503
								}
								goto l502
							l503:
								position, tokenIndex, depth = position503, tokenIndex503, depth503
							}
							if !_rules[ruleCHAR]() {
								goto l502
							}
							goto l501
						l502:
							position, tokenIndex, depth = position502, tokenIndex502, depth502
						}
						depth--
						add(rulePegText, position500)
					}
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l492
					}
				}
			l494:
				depth--
				add(ruleSTRING, position493)
			}
			return true
		l492:
			position, tokenIndex, depth = position492, tokenIndex492, depth492
			return false
		},
		/* 45 CHAR <- <(('\\' ((&('"') QUOTE_DOUBLE) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position504, tokenIndex504, depth504 := position, tokenIndex, depth
			{
				position505 := position
				depth++
				{
					position506, tokenIndex506, depth506 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l507
					}
					position++
					{
						switch buffer[position] {
						case '"':
							if !_rules[ruleQUOTE_DOUBLE]() {
								goto l507
							}
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l507
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l507
							}
							break
						}
					}

					goto l506
				l507:
					position, tokenIndex, depth = position506, tokenIndex506, depth506
					{
						position509, tokenIndex509, depth509 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l509
						}
						goto l504
					l509:
						position, tokenIndex, depth = position509, tokenIndex509, depth509
					}
					if !matchDot() {
						goto l504
					}
				}
			l506:
				depth--
				add(ruleCHAR, position505)
			}
			return true
		l504:
			position, tokenIndex, depth = position504, tokenIndex504, depth504
			return false
		},
		/* 46 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position510, tokenIndex510, depth510 := position, tokenIndex, depth
			{
				position511 := position
				depth++
				{
					position512, tokenIndex512, depth512 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l513
					}
					position++
					goto l512
				l513:
					position, tokenIndex, depth = position512, tokenIndex512, depth512
					if buffer[position] != rune('\\') {
						goto l510
					}
					position++
				}
			l512:
				depth--
				add(ruleESCAPE_CLASS, position511)
			}
			return true
		l510:
			position, tokenIndex, depth = position510, tokenIndex510, depth510
			return false
		},
		/* 47 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position514, tokenIndex514, depth514 := position, tokenIndex, depth
			{
				position515 := position
				depth++
				{
					position516 := position
					depth++
					{
						position517, tokenIndex517, depth517 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l517
						}
						position++
						goto l518
					l517:
						position, tokenIndex, depth = position517, tokenIndex517, depth517
					}
				l518:
					{
						position519 := position
						depth++
						{
							position520, tokenIndex520, depth520 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l521
							}
							position++
							goto l520
						l521:
							position, tokenIndex, depth = position520, tokenIndex520, depth520
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l514
							}
							position++
						l522:
							{
								position523, tokenIndex523, depth523 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l523
								}
								position++
								goto l522
							l523:
								position, tokenIndex, depth = position523, tokenIndex523, depth523
							}
						}
					l520:
						depth--
						add(ruleNUMBER_NATURAL, position519)
					}
					depth--
					add(ruleNUMBER_INTEGER, position516)
				}
				{
					position524, tokenIndex524, depth524 := position, tokenIndex, depth
					{
						position526 := position
						depth++
						if buffer[position] != rune('.') {
							goto l524
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l524
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
						depth--
						add(ruleNUMBER_FRACTION, position526)
					}
					goto l525
				l524:
					position, tokenIndex, depth = position524, tokenIndex524, depth524
				}
			l525:
				{
					position529, tokenIndex529, depth529 := position, tokenIndex, depth
					{
						position531 := position
						depth++
						{
							position532, tokenIndex532, depth532 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l533
							}
							position++
							goto l532
						l533:
							position, tokenIndex, depth = position532, tokenIndex532, depth532
							if buffer[position] != rune('E') {
								goto l529
							}
							position++
						}
					l532:
						{
							position534, tokenIndex534, depth534 := position, tokenIndex, depth
							{
								position536, tokenIndex536, depth536 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l537
								}
								position++
								goto l536
							l537:
								position, tokenIndex, depth = position536, tokenIndex536, depth536
								if buffer[position] != rune('-') {
									goto l534
								}
								position++
							}
						l536:
							goto l535
						l534:
							position, tokenIndex, depth = position534, tokenIndex534, depth534
						}
					l535:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l529
						}
						position++
					l538:
						{
							position539, tokenIndex539, depth539 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l539
							}
							position++
							goto l538
						l539:
							position, tokenIndex, depth = position539, tokenIndex539, depth539
						}
						depth--
						add(ruleNUMBER_EXP, position531)
					}
					goto l530
				l529:
					position, tokenIndex, depth = position529, tokenIndex529, depth529
				}
			l530:
				depth--
				add(ruleNUMBER, position515)
			}
			return true
		l514:
			position, tokenIndex, depth = position514, tokenIndex514, depth514
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
		/* 52 PAREN_OPEN <- <(_ '(' _)> */
		func() bool {
			position544, tokenIndex544, depth544 := position, tokenIndex, depth
			{
				position545 := position
				depth++
				if !_rules[rule_]() {
					goto l544
				}
				if buffer[position] != rune('(') {
					goto l544
				}
				position++
				if !_rules[rule_]() {
					goto l544
				}
				depth--
				add(rulePAREN_OPEN, position545)
			}
			return true
		l544:
			position, tokenIndex, depth = position544, tokenIndex544, depth544
			return false
		},
		/* 53 PAREN_CLOSE <- <(_ ')' _)> */
		func() bool {
			position546, tokenIndex546, depth546 := position, tokenIndex, depth
			{
				position547 := position
				depth++
				if !_rules[rule_]() {
					goto l546
				}
				if buffer[position] != rune(')') {
					goto l546
				}
				position++
				if !_rules[rule_]() {
					goto l546
				}
				depth--
				add(rulePAREN_CLOSE, position547)
			}
			return true
		l546:
			position, tokenIndex, depth = position546, tokenIndex546, depth546
			return false
		},
		/* 54 COMMA <- <(_ ',' _)> */
		func() bool {
			position548, tokenIndex548, depth548 := position, tokenIndex, depth
			{
				position549 := position
				depth++
				if !_rules[rule_]() {
					goto l548
				}
				if buffer[position] != rune(',') {
					goto l548
				}
				position++
				if !_rules[rule_]() {
					goto l548
				}
				depth--
				add(ruleCOMMA, position549)
			}
			return true
		l548:
			position, tokenIndex, depth = position548, tokenIndex548, depth548
			return false
		},
		/* 55 _ <- <SPACE*> */
		func() bool {
			{
				position551 := position
				depth++
			l552:
				{
					position553, tokenIndex553, depth553 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l553
					}
					goto l552
				l553:
					position, tokenIndex, depth = position553, tokenIndex553, depth553
				}
				depth--
				add(rule_, position551)
			}
			return true
		},
		/* 56 __ <- <(!ID_CONT SPACE*)> */
		func() bool {
			position554, tokenIndex554, depth554 := position, tokenIndex, depth
			{
				position555 := position
				depth++
				{
					position556, tokenIndex556, depth556 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l556
					}
					goto l554
				l556:
					position, tokenIndex, depth = position556, tokenIndex556, depth556
				}
			l557:
				{
					position558, tokenIndex558, depth558 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l558
					}
					goto l557
				l558:
					position, tokenIndex, depth = position558, tokenIndex558, depth558
				}
				depth--
				add(rule__, position555)
			}
			return true
		l554:
			position, tokenIndex, depth = position554, tokenIndex554, depth554
			return false
		},
		/* 57 SPACE <- <((&('\t') '\t') | (&('\n') '\n') | (&(' ') ' '))> */
		func() bool {
			position559, tokenIndex559, depth559 := position, tokenIndex, depth
			{
				position560 := position
				depth++
				{
					switch buffer[position] {
					case '\t':
						if buffer[position] != rune('\t') {
							goto l559
						}
						position++
						break
					case '\n':
						if buffer[position] != rune('\n') {
							goto l559
						}
						position++
						break
					default:
						if buffer[position] != rune(' ') {
							goto l559
						}
						position++
						break
					}
				}

				depth--
				add(ruleSPACE, position560)
			}
			return true
		l559:
			position, tokenIndex, depth = position559, tokenIndex559, depth559
			return false
		},
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
