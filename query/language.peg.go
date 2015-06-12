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
								position68 := position
								depth++
								{
									position69 := position
									depth++
									{
										position70 := position
										depth++
										if !_rules[ruleIDENTIFIER]() {
											goto l0
										}
										depth--
										add(ruleMETRIC_NAME, position70)
									}
									depth--
									add(rulePegText, position69)
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
								add(ruledescribeSingleStmt, position68)
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
					position73, tokenIndex73, depth73 := position, tokenIndex, depth
					if !matchDot() {
						goto l73
					}
					goto l0
				l73:
					position, tokenIndex, depth = position73, tokenIndex73, depth73
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
				position80 := position
				depth++
				{
					position81, tokenIndex81, depth81 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l82
					}
					{
						position83 := position
						depth++
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
								goto l82
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
								goto l82
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
								goto l82
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
								goto l82
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
								goto l82
							}
							position++
						}
					l92:
						if !_rules[rule__]() {
							goto l82
						}
						if !_rules[rulepredicate_1]() {
							goto l82
						}
						depth--
						add(rulepredicateClause, position83)
					}
					goto l81
				l82:
					position, tokenIndex, depth = position81, tokenIndex81, depth81
					{
						add(ruleAction9, position)
					}
				}
			l81:
				depth--
				add(ruleoptionalPredicateClause, position80)
			}
			return true
		},
		/* 7 expressionList <- <(Action10 expression_1 Action11 (COMMA expression_1 Action12)*)> */
		func() bool {
			position95, tokenIndex95, depth95 := position, tokenIndex, depth
			{
				position96 := position
				depth++
				{
					add(ruleAction10, position)
				}
				if !_rules[ruleexpression_1]() {
					goto l95
				}
				{
					add(ruleAction11, position)
				}
			l99:
				{
					position100, tokenIndex100, depth100 := position, tokenIndex, depth
					if !_rules[ruleCOMMA]() {
						goto l100
					}
					if !_rules[ruleexpression_1]() {
						goto l100
					}
					{
						add(ruleAction12, position)
					}
					goto l99
				l100:
					position, tokenIndex, depth = position100, tokenIndex100, depth100
				}
				depth--
				add(ruleexpressionList, position96)
			}
			return true
		l95:
			position, tokenIndex, depth = position95, tokenIndex95, depth95
			return false
		},
		/* 8 expression_1 <- <(expression_2 (((OP_ADD Action13) / (OP_SUB Action14)) expression_2 Action15)*)> */
		func() bool {
			position102, tokenIndex102, depth102 := position, tokenIndex, depth
			{
				position103 := position
				depth++
				if !_rules[ruleexpression_2]() {
					goto l102
				}
			l104:
				{
					position105, tokenIndex105, depth105 := position, tokenIndex, depth
					{
						position106, tokenIndex106, depth106 := position, tokenIndex, depth
						{
							position108 := position
							depth++
							if !_rules[rule_]() {
								goto l107
							}
							if buffer[position] != rune('+') {
								goto l107
							}
							position++
							if !_rules[rule_]() {
								goto l107
							}
							depth--
							add(ruleOP_ADD, position108)
						}
						{
							add(ruleAction13, position)
						}
						goto l106
					l107:
						position, tokenIndex, depth = position106, tokenIndex106, depth106
						{
							position110 := position
							depth++
							if !_rules[rule_]() {
								goto l105
							}
							if buffer[position] != rune('-') {
								goto l105
							}
							position++
							if !_rules[rule_]() {
								goto l105
							}
							depth--
							add(ruleOP_SUB, position110)
						}
						{
							add(ruleAction14, position)
						}
					}
				l106:
					if !_rules[ruleexpression_2]() {
						goto l105
					}
					{
						add(ruleAction15, position)
					}
					goto l104
				l105:
					position, tokenIndex, depth = position105, tokenIndex105, depth105
				}
				depth--
				add(ruleexpression_1, position103)
			}
			return true
		l102:
			position, tokenIndex, depth = position102, tokenIndex102, depth102
			return false
		},
		/* 9 expression_2 <- <(expression_3 (((OP_DIV Action16) / (OP_MULT Action17)) expression_3 Action18)*)> */
		func() bool {
			position113, tokenIndex113, depth113 := position, tokenIndex, depth
			{
				position114 := position
				depth++
				if !_rules[ruleexpression_3]() {
					goto l113
				}
			l115:
				{
					position116, tokenIndex116, depth116 := position, tokenIndex, depth
					{
						position117, tokenIndex117, depth117 := position, tokenIndex, depth
						{
							position119 := position
							depth++
							if !_rules[rule_]() {
								goto l118
							}
							if buffer[position] != rune('/') {
								goto l118
							}
							position++
							if !_rules[rule_]() {
								goto l118
							}
							depth--
							add(ruleOP_DIV, position119)
						}
						{
							add(ruleAction16, position)
						}
						goto l117
					l118:
						position, tokenIndex, depth = position117, tokenIndex117, depth117
						{
							position121 := position
							depth++
							if !_rules[rule_]() {
								goto l116
							}
							if buffer[position] != rune('*') {
								goto l116
							}
							position++
							if !_rules[rule_]() {
								goto l116
							}
							depth--
							add(ruleOP_MULT, position121)
						}
						{
							add(ruleAction17, position)
						}
					}
				l117:
					if !_rules[ruleexpression_3]() {
						goto l116
					}
					{
						add(ruleAction18, position)
					}
					goto l115
				l116:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
				}
				depth--
				add(ruleexpression_2, position114)
			}
			return true
		l113:
			position, tokenIndex, depth = position113, tokenIndex113, depth113
			return false
		},
		/* 10 expression_3 <- <(expression_function / ((&('-' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') (<NUMBER> Action19)) | (&('\t' | '\n' | ' ' | '(') (PAREN_OPEN expression_1 PAREN_CLOSE)) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') expression_metric)))> */
		func() bool {
			position124, tokenIndex124, depth124 := position, tokenIndex, depth
			{
				position125 := position
				depth++
				{
					position126, tokenIndex126, depth126 := position, tokenIndex, depth
					{
						position128 := position
						depth++
						{
							position129 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l127
							}
							depth--
							add(rulePegText, position129)
						}
						{
							add(ruleAction20, position)
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l127
						}
						if !_rules[ruleexpressionList]() {
							goto l127
						}
						{
							add(ruleAction21, position)
						}
						{
							position132, tokenIndex132, depth132 := position, tokenIndex, depth
							if !_rules[rule__]() {
								goto l132
							}
							{
								position134 := position
								depth++
								{
									position135, tokenIndex135, depth135 := position, tokenIndex, depth
									if buffer[position] != rune('g') {
										goto l136
									}
									position++
									goto l135
								l136:
									position, tokenIndex, depth = position135, tokenIndex135, depth135
									if buffer[position] != rune('G') {
										goto l132
									}
									position++
								}
							l135:
								{
									position137, tokenIndex137, depth137 := position, tokenIndex, depth
									if buffer[position] != rune('r') {
										goto l138
									}
									position++
									goto l137
								l138:
									position, tokenIndex, depth = position137, tokenIndex137, depth137
									if buffer[position] != rune('R') {
										goto l132
									}
									position++
								}
							l137:
								{
									position139, tokenIndex139, depth139 := position, tokenIndex, depth
									if buffer[position] != rune('o') {
										goto l140
									}
									position++
									goto l139
								l140:
									position, tokenIndex, depth = position139, tokenIndex139, depth139
									if buffer[position] != rune('O') {
										goto l132
									}
									position++
								}
							l139:
								{
									position141, tokenIndex141, depth141 := position, tokenIndex, depth
									if buffer[position] != rune('u') {
										goto l142
									}
									position++
									goto l141
								l142:
									position, tokenIndex, depth = position141, tokenIndex141, depth141
									if buffer[position] != rune('U') {
										goto l132
									}
									position++
								}
							l141:
								{
									position143, tokenIndex143, depth143 := position, tokenIndex, depth
									if buffer[position] != rune('p') {
										goto l144
									}
									position++
									goto l143
								l144:
									position, tokenIndex, depth = position143, tokenIndex143, depth143
									if buffer[position] != rune('P') {
										goto l132
									}
									position++
								}
							l143:
								if !_rules[rule__]() {
									goto l132
								}
								{
									position145, tokenIndex145, depth145 := position, tokenIndex, depth
									if buffer[position] != rune('b') {
										goto l146
									}
									position++
									goto l145
								l146:
									position, tokenIndex, depth = position145, tokenIndex145, depth145
									if buffer[position] != rune('B') {
										goto l132
									}
									position++
								}
							l145:
								{
									position147, tokenIndex147, depth147 := position, tokenIndex, depth
									if buffer[position] != rune('y') {
										goto l148
									}
									position++
									goto l147
								l148:
									position, tokenIndex, depth = position147, tokenIndex147, depth147
									if buffer[position] != rune('Y') {
										goto l132
									}
									position++
								}
							l147:
								if !_rules[rule__]() {
									goto l132
								}
								{
									position149 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l132
									}
									depth--
									add(rulePegText, position149)
								}
								{
									add(ruleAction26, position)
								}
							l151:
								{
									position152, tokenIndex152, depth152 := position, tokenIndex, depth
									if !_rules[ruleCOMMA]() {
										goto l152
									}
									{
										position153 := position
										depth++
										if !_rules[ruleCOLUMN_NAME]() {
											goto l152
										}
										depth--
										add(rulePegText, position153)
									}
									{
										add(ruleAction27, position)
									}
									goto l151
								l152:
									position, tokenIndex, depth = position152, tokenIndex152, depth152
								}
								depth--
								add(rulegroupByClause, position134)
							}
							goto l133
						l132:
							position, tokenIndex, depth = position132, tokenIndex132, depth132
						}
					l133:
						if !_rules[rulePAREN_CLOSE]() {
							goto l127
						}
						{
							add(ruleAction22, position)
						}
						depth--
						add(ruleexpression_function, position128)
					}
					goto l126
				l127:
					position, tokenIndex, depth = position126, tokenIndex126, depth126
					{
						switch buffer[position] {
						case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							{
								position157 := position
								depth++
								if !_rules[ruleNUMBER]() {
									goto l124
								}
								depth--
								add(rulePegText, position157)
							}
							{
								add(ruleAction19, position)
							}
							break
						case '\t', '\n', ' ', '(':
							if !_rules[rulePAREN_OPEN]() {
								goto l124
							}
							if !_rules[ruleexpression_1]() {
								goto l124
							}
							if !_rules[rulePAREN_CLOSE]() {
								goto l124
							}
							break
						default:
							{
								position159 := position
								depth++
								{
									position160 := position
									depth++
									if !_rules[ruleIDENTIFIER]() {
										goto l124
									}
									depth--
									add(rulePegText, position160)
								}
								{
									add(ruleAction23, position)
								}
								{
									position162, tokenIndex162, depth162 := position, tokenIndex, depth
									{
										position164, tokenIndex164, depth164 := position, tokenIndex, depth
										if buffer[position] != rune('[') {
											goto l165
										}
										position++
										if !_rules[rule_]() {
											goto l165
										}
										if !_rules[rulepredicate_1]() {
											goto l165
										}
										if !_rules[rule_]() {
											goto l165
										}
										if buffer[position] != rune(']') {
											goto l165
										}
										position++
										goto l164
									l165:
										position, tokenIndex, depth = position164, tokenIndex164, depth164
										{
											add(ruleAction24, position)
										}
									}
								l164:
									goto l163

									position, tokenIndex, depth = position162, tokenIndex162, depth162
								}
							l163:
								{
									add(ruleAction25, position)
								}
								depth--
								add(ruleexpression_metric, position159)
							}
							break
						}
					}

				}
			l126:
				depth--
				add(ruleexpression_3, position125)
			}
			return true
		l124:
			position, tokenIndex, depth = position124, tokenIndex124, depth124
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
				position173 := position
				depth++
				{
					position174, tokenIndex174, depth174 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l175
					}
					{
						position176 := position
						depth++
						if !_rules[rule_]() {
							goto l175
						}
						{
							position177, tokenIndex177, depth177 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l178
							}
							position++
							goto l177
						l178:
							position, tokenIndex, depth = position177, tokenIndex177, depth177
							if buffer[position] != rune('O') {
								goto l175
							}
							position++
						}
					l177:
						{
							position179, tokenIndex179, depth179 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l180
							}
							position++
							goto l179
						l180:
							position, tokenIndex, depth = position179, tokenIndex179, depth179
							if buffer[position] != rune('R') {
								goto l175
							}
							position++
						}
					l179:
						if !_rules[rule_]() {
							goto l175
						}
						depth--
						add(ruleOP_OR, position176)
					}
					if !_rules[rulepredicate_1]() {
						goto l175
					}
					{
						add(ruleAction28, position)
					}
					goto l174
				l175:
					position, tokenIndex, depth = position174, tokenIndex174, depth174
					if !_rules[rulepredicate_2]() {
						goto l182
					}
					goto l174
				l182:
					position, tokenIndex, depth = position174, tokenIndex174, depth174
				}
			l174:
				depth--
				add(rulepredicate_1, position173)
			}
			return true
		},
		/* 16 predicate_2 <- <((predicate_3 OP_AND predicate_2 Action29) / predicate_3)> */
		func() bool {
			position183, tokenIndex183, depth183 := position, tokenIndex, depth
			{
				position184 := position
				depth++
				{
					position185, tokenIndex185, depth185 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l186
					}
					{
						position187 := position
						depth++
						if !_rules[rule_]() {
							goto l186
						}
						{
							position188, tokenIndex188, depth188 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l189
							}
							position++
							goto l188
						l189:
							position, tokenIndex, depth = position188, tokenIndex188, depth188
							if buffer[position] != rune('A') {
								goto l186
							}
							position++
						}
					l188:
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
								goto l186
							}
							position++
						}
					l190:
						{
							position192, tokenIndex192, depth192 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l193
							}
							position++
							goto l192
						l193:
							position, tokenIndex, depth = position192, tokenIndex192, depth192
							if buffer[position] != rune('D') {
								goto l186
							}
							position++
						}
					l192:
						if !_rules[rule_]() {
							goto l186
						}
						depth--
						add(ruleOP_AND, position187)
					}
					if !_rules[rulepredicate_2]() {
						goto l186
					}
					{
						add(ruleAction29, position)
					}
					goto l185
				l186:
					position, tokenIndex, depth = position185, tokenIndex185, depth185
					if !_rules[rulepredicate_3]() {
						goto l183
					}
				}
			l185:
				depth--
				add(rulepredicate_2, position184)
			}
			return true
		l183:
			position, tokenIndex, depth = position183, tokenIndex183, depth183
			return false
		},
		/* 17 predicate_3 <- <((OP_NOT predicate_3 Action30) / (PAREN_OPEN predicate_1 PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position195, tokenIndex195, depth195 := position, tokenIndex, depth
			{
				position196 := position
				depth++
				{
					position197, tokenIndex197, depth197 := position, tokenIndex, depth
					{
						position199 := position
						depth++
						{
							position200, tokenIndex200, depth200 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l201
							}
							position++
							goto l200
						l201:
							position, tokenIndex, depth = position200, tokenIndex200, depth200
							if buffer[position] != rune('N') {
								goto l198
							}
							position++
						}
					l200:
						{
							position202, tokenIndex202, depth202 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l203
							}
							position++
							goto l202
						l203:
							position, tokenIndex, depth = position202, tokenIndex202, depth202
							if buffer[position] != rune('O') {
								goto l198
							}
							position++
						}
					l202:
						{
							position204, tokenIndex204, depth204 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l205
							}
							position++
							goto l204
						l205:
							position, tokenIndex, depth = position204, tokenIndex204, depth204
							if buffer[position] != rune('T') {
								goto l198
							}
							position++
						}
					l204:
						if !_rules[rule__]() {
							goto l198
						}
						depth--
						add(ruleOP_NOT, position199)
					}
					if !_rules[rulepredicate_3]() {
						goto l198
					}
					{
						add(ruleAction30, position)
					}
					goto l197
				l198:
					position, tokenIndex, depth = position197, tokenIndex197, depth197
					if !_rules[rulePAREN_OPEN]() {
						goto l207
					}
					if !_rules[rulepredicate_1]() {
						goto l207
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l207
					}
					goto l197
				l207:
					position, tokenIndex, depth = position197, tokenIndex197, depth197
					{
						position208 := position
						depth++
						{
							position209, tokenIndex209, depth209 := position, tokenIndex, depth
							if !_rules[ruletagName]() {
								goto l210
							}
							if !_rules[rule_]() {
								goto l210
							}
							if buffer[position] != rune('=') {
								goto l210
							}
							position++
							if !_rules[rule_]() {
								goto l210
							}
							if !_rules[ruleliteralString]() {
								goto l210
							}
							{
								add(ruleAction31, position)
							}
							goto l209
						l210:
							position, tokenIndex, depth = position209, tokenIndex209, depth209
							if !_rules[ruletagName]() {
								goto l212
							}
							if !_rules[rule_]() {
								goto l212
							}
							if buffer[position] != rune('!') {
								goto l212
							}
							position++
							if buffer[position] != rune('=') {
								goto l212
							}
							position++
							if !_rules[rule_]() {
								goto l212
							}
							if !_rules[ruleliteralString]() {
								goto l212
							}
							{
								add(ruleAction32, position)
							}
							goto l209
						l212:
							position, tokenIndex, depth = position209, tokenIndex209, depth209
							if !_rules[ruletagName]() {
								goto l214
							}
							if !_rules[rule__]() {
								goto l214
							}
							{
								position215, tokenIndex215, depth215 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l216
								}
								position++
								goto l215
							l216:
								position, tokenIndex, depth = position215, tokenIndex215, depth215
								if buffer[position] != rune('M') {
									goto l214
								}
								position++
							}
						l215:
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
									goto l214
								}
								position++
							}
						l217:
							{
								position219, tokenIndex219, depth219 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l220
								}
								position++
								goto l219
							l220:
								position, tokenIndex, depth = position219, tokenIndex219, depth219
								if buffer[position] != rune('T') {
									goto l214
								}
								position++
							}
						l219:
							{
								position221, tokenIndex221, depth221 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l222
								}
								position++
								goto l221
							l222:
								position, tokenIndex, depth = position221, tokenIndex221, depth221
								if buffer[position] != rune('C') {
									goto l214
								}
								position++
							}
						l221:
							{
								position223, tokenIndex223, depth223 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l224
								}
								position++
								goto l223
							l224:
								position, tokenIndex, depth = position223, tokenIndex223, depth223
								if buffer[position] != rune('H') {
									goto l214
								}
								position++
							}
						l223:
							{
								position225, tokenIndex225, depth225 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l226
								}
								position++
								goto l225
							l226:
								position, tokenIndex, depth = position225, tokenIndex225, depth225
								if buffer[position] != rune('E') {
									goto l214
								}
								position++
							}
						l225:
							{
								position227, tokenIndex227, depth227 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l228
								}
								position++
								goto l227
							l228:
								position, tokenIndex, depth = position227, tokenIndex227, depth227
								if buffer[position] != rune('S') {
									goto l214
								}
								position++
							}
						l227:
							if !_rules[rule__]() {
								goto l214
							}
							if !_rules[ruleliteralString]() {
								goto l214
							}
							{
								add(ruleAction33, position)
							}
							goto l209
						l214:
							position, tokenIndex, depth = position209, tokenIndex209, depth209
							if !_rules[ruletagName]() {
								goto l195
							}
							if !_rules[rule__]() {
								goto l195
							}
							{
								position230, tokenIndex230, depth230 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l231
								}
								position++
								goto l230
							l231:
								position, tokenIndex, depth = position230, tokenIndex230, depth230
								if buffer[position] != rune('I') {
									goto l195
								}
								position++
							}
						l230:
							{
								position232, tokenIndex232, depth232 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l233
								}
								position++
								goto l232
							l233:
								position, tokenIndex, depth = position232, tokenIndex232, depth232
								if buffer[position] != rune('N') {
									goto l195
								}
								position++
							}
						l232:
							if !_rules[rule__]() {
								goto l195
							}
							{
								position234 := position
								depth++
								{
									add(ruleAction36, position)
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l195
								}
								if !_rules[ruleliteralListString]() {
									goto l195
								}
							l236:
								{
									position237, tokenIndex237, depth237 := position, tokenIndex, depth
									if !_rules[ruleCOMMA]() {
										goto l237
									}
									if !_rules[ruleliteralListString]() {
										goto l237
									}
									goto l236
								l237:
									position, tokenIndex, depth = position237, tokenIndex237, depth237
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l195
								}
								depth--
								add(ruleliteralList, position234)
							}
							{
								add(ruleAction34, position)
							}
						}
					l209:
						depth--
						add(ruletagMatcher, position208)
					}
				}
			l197:
				depth--
				add(rulepredicate_3, position196)
			}
			return true
		l195:
			position, tokenIndex, depth = position195, tokenIndex195, depth195
			return false
		},
		/* 18 tagMatcher <- <((tagName _ '=' _ literalString Action31) / (tagName _ ('!' '=') _ literalString Action32) / (tagName __ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) __ literalString Action33) / (tagName __ (('i' / 'I') ('n' / 'N')) __ literalList Action34))> */
		nil,
		/* 19 literalString <- <(STRING Action35)> */
		func() bool {
			position240, tokenIndex240, depth240 := position, tokenIndex, depth
			{
				position241 := position
				depth++
				if !_rules[ruleSTRING]() {
					goto l240
				}
				{
					add(ruleAction35, position)
				}
				depth--
				add(ruleliteralString, position241)
			}
			return true
		l240:
			position, tokenIndex, depth = position240, tokenIndex240, depth240
			return false
		},
		/* 20 literalList <- <(Action36 PAREN_OPEN literalListString (COMMA literalListString)* PAREN_CLOSE)> */
		nil,
		/* 21 literalListString <- <(STRING Action37)> */
		func() bool {
			position244, tokenIndex244, depth244 := position, tokenIndex, depth
			{
				position245 := position
				depth++
				if !_rules[ruleSTRING]() {
					goto l244
				}
				{
					add(ruleAction37, position)
				}
				depth--
				add(ruleliteralListString, position245)
			}
			return true
		l244:
			position, tokenIndex, depth = position244, tokenIndex244, depth244
			return false
		},
		/* 22 tagName <- <(<TAG_NAME> Action38)> */
		func() bool {
			position247, tokenIndex247, depth247 := position, tokenIndex, depth
			{
				position248 := position
				depth++
				{
					position249 := position
					depth++
					{
						position250 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l247
						}
						depth--
						add(ruleTAG_NAME, position250)
					}
					depth--
					add(rulePegText, position249)
				}
				{
					add(ruleAction38, position)
				}
				depth--
				add(ruletagName, position248)
			}
			return true
		l247:
			position, tokenIndex, depth = position247, tokenIndex247, depth247
			return false
		},
		/* 23 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position252, tokenIndex252, depth252 := position, tokenIndex, depth
			{
				position253 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l252
				}
				depth--
				add(ruleCOLUMN_NAME, position253)
			}
			return true
		l252:
			position, tokenIndex, depth = position252, tokenIndex252, depth252
			return false
		},
		/* 24 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 25 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 26 IDENTIFIER <- <(('`' CHAR* '`') / (!(KEYWORD !ID_CONT) ID_SEGMENT ('.' ID_SEGMENT)*))> */
		func() bool {
			position256, tokenIndex256, depth256 := position, tokenIndex, depth
			{
				position257 := position
				depth++
				{
					position258, tokenIndex258, depth258 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l259
					}
					position++
				l260:
					{
						position261, tokenIndex261, depth261 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l261
						}
						goto l260
					l261:
						position, tokenIndex, depth = position261, tokenIndex261, depth261
					}
					if buffer[position] != rune('`') {
						goto l259
					}
					position++
					goto l258
				l259:
					position, tokenIndex, depth = position258, tokenIndex258, depth258
					{
						position262, tokenIndex262, depth262 := position, tokenIndex, depth
						{
							position263 := position
							depth++
							{
								position264, tokenIndex264, depth264 := position, tokenIndex, depth
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
									if buffer[position] != rune('l') {
										goto l269
									}
									position++
									goto l268
								l269:
									position, tokenIndex, depth = position268, tokenIndex268, depth268
									if buffer[position] != rune('L') {
										goto l265
									}
									position++
								}
							l268:
								{
									position270, tokenIndex270, depth270 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l271
									}
									position++
									goto l270
								l271:
									position, tokenIndex, depth = position270, tokenIndex270, depth270
									if buffer[position] != rune('L') {
										goto l265
									}
									position++
								}
							l270:
								goto l264
							l265:
								position, tokenIndex, depth = position264, tokenIndex264, depth264
								{
									position273, tokenIndex273, depth273 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l274
									}
									position++
									goto l273
								l274:
									position, tokenIndex, depth = position273, tokenIndex273, depth273
									if buffer[position] != rune('A') {
										goto l272
									}
									position++
								}
							l273:
								{
									position275, tokenIndex275, depth275 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l276
									}
									position++
									goto l275
								l276:
									position, tokenIndex, depth = position275, tokenIndex275, depth275
									if buffer[position] != rune('N') {
										goto l272
									}
									position++
								}
							l275:
								{
									position277, tokenIndex277, depth277 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l278
									}
									position++
									goto l277
								l278:
									position, tokenIndex, depth = position277, tokenIndex277, depth277
									if buffer[position] != rune('D') {
										goto l272
									}
									position++
								}
							l277:
								goto l264
							l272:
								position, tokenIndex, depth = position264, tokenIndex264, depth264
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
										goto l279
									}
									position++
								}
							l280:
								{
									position282, tokenIndex282, depth282 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l283
									}
									position++
									goto l282
								l283:
									position, tokenIndex, depth = position282, tokenIndex282, depth282
									if buffer[position] != rune('E') {
										goto l279
									}
									position++
								}
							l282:
								{
									position284, tokenIndex284, depth284 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l285
									}
									position++
									goto l284
								l285:
									position, tokenIndex, depth = position284, tokenIndex284, depth284
									if buffer[position] != rune('L') {
										goto l279
									}
									position++
								}
							l284:
								{
									position286, tokenIndex286, depth286 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l287
									}
									position++
									goto l286
								l287:
									position, tokenIndex, depth = position286, tokenIndex286, depth286
									if buffer[position] != rune('E') {
										goto l279
									}
									position++
								}
							l286:
								{
									position288, tokenIndex288, depth288 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l289
									}
									position++
									goto l288
								l289:
									position, tokenIndex, depth = position288, tokenIndex288, depth288
									if buffer[position] != rune('C') {
										goto l279
									}
									position++
								}
							l288:
								{
									position290, tokenIndex290, depth290 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l291
									}
									position++
									goto l290
								l291:
									position, tokenIndex, depth = position290, tokenIndex290, depth290
									if buffer[position] != rune('T') {
										goto l279
									}
									position++
								}
							l290:
								goto l264
							l279:
								position, tokenIndex, depth = position264, tokenIndex264, depth264
								{
									switch buffer[position] {
									case 'W', 'w':
										{
											position293, tokenIndex293, depth293 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l294
											}
											position++
											goto l293
										l294:
											position, tokenIndex, depth = position293, tokenIndex293, depth293
											if buffer[position] != rune('W') {
												goto l262
											}
											position++
										}
									l293:
										{
											position295, tokenIndex295, depth295 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l296
											}
											position++
											goto l295
										l296:
											position, tokenIndex, depth = position295, tokenIndex295, depth295
											if buffer[position] != rune('H') {
												goto l262
											}
											position++
										}
									l295:
										{
											position297, tokenIndex297, depth297 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l298
											}
											position++
											goto l297
										l298:
											position, tokenIndex, depth = position297, tokenIndex297, depth297
											if buffer[position] != rune('E') {
												goto l262
											}
											position++
										}
									l297:
										{
											position299, tokenIndex299, depth299 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l300
											}
											position++
											goto l299
										l300:
											position, tokenIndex, depth = position299, tokenIndex299, depth299
											if buffer[position] != rune('R') {
												goto l262
											}
											position++
										}
									l299:
										{
											position301, tokenIndex301, depth301 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l302
											}
											position++
											goto l301
										l302:
											position, tokenIndex, depth = position301, tokenIndex301, depth301
											if buffer[position] != rune('E') {
												goto l262
											}
											position++
										}
									l301:
										break
									case 'O', 'o':
										{
											position303, tokenIndex303, depth303 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l304
											}
											position++
											goto l303
										l304:
											position, tokenIndex, depth = position303, tokenIndex303, depth303
											if buffer[position] != rune('O') {
												goto l262
											}
											position++
										}
									l303:
										{
											position305, tokenIndex305, depth305 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l306
											}
											position++
											goto l305
										l306:
											position, tokenIndex, depth = position305, tokenIndex305, depth305
											if buffer[position] != rune('R') {
												goto l262
											}
											position++
										}
									l305:
										break
									case 'N', 'n':
										{
											position307, tokenIndex307, depth307 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l308
											}
											position++
											goto l307
										l308:
											position, tokenIndex, depth = position307, tokenIndex307, depth307
											if buffer[position] != rune('N') {
												goto l262
											}
											position++
										}
									l307:
										{
											position309, tokenIndex309, depth309 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l310
											}
											position++
											goto l309
										l310:
											position, tokenIndex, depth = position309, tokenIndex309, depth309
											if buffer[position] != rune('O') {
												goto l262
											}
											position++
										}
									l309:
										{
											position311, tokenIndex311, depth311 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l312
											}
											position++
											goto l311
										l312:
											position, tokenIndex, depth = position311, tokenIndex311, depth311
											if buffer[position] != rune('T') {
												goto l262
											}
											position++
										}
									l311:
										break
									case 'M', 'm':
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
												goto l262
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
												goto l262
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
												goto l262
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
												goto l262
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
												goto l262
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
												goto l262
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
												goto l262
											}
											position++
										}
									l325:
										break
									case 'I', 'i':
										{
											position327, tokenIndex327, depth327 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l328
											}
											position++
											goto l327
										l328:
											position, tokenIndex, depth = position327, tokenIndex327, depth327
											if buffer[position] != rune('I') {
												goto l262
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
												goto l262
											}
											position++
										}
									l329:
										break
									case 'G', 'g':
										{
											position331, tokenIndex331, depth331 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l332
											}
											position++
											goto l331
										l332:
											position, tokenIndex, depth = position331, tokenIndex331, depth331
											if buffer[position] != rune('G') {
												goto l262
											}
											position++
										}
									l331:
										{
											position333, tokenIndex333, depth333 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l334
											}
											position++
											goto l333
										l334:
											position, tokenIndex, depth = position333, tokenIndex333, depth333
											if buffer[position] != rune('R') {
												goto l262
											}
											position++
										}
									l333:
										{
											position335, tokenIndex335, depth335 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l336
											}
											position++
											goto l335
										l336:
											position, tokenIndex, depth = position335, tokenIndex335, depth335
											if buffer[position] != rune('O') {
												goto l262
											}
											position++
										}
									l335:
										{
											position337, tokenIndex337, depth337 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l338
											}
											position++
											goto l337
										l338:
											position, tokenIndex, depth = position337, tokenIndex337, depth337
											if buffer[position] != rune('U') {
												goto l262
											}
											position++
										}
									l337:
										{
											position339, tokenIndex339, depth339 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l340
											}
											position++
											goto l339
										l340:
											position, tokenIndex, depth = position339, tokenIndex339, depth339
											if buffer[position] != rune('P') {
												goto l262
											}
											position++
										}
									l339:
										break
									case 'D', 'd':
										{
											position341, tokenIndex341, depth341 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l342
											}
											position++
											goto l341
										l342:
											position, tokenIndex, depth = position341, tokenIndex341, depth341
											if buffer[position] != rune('D') {
												goto l262
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
												goto l262
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
												goto l262
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
												goto l262
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
												goto l262
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
												goto l262
											}
											position++
										}
									l351:
										{
											position353, tokenIndex353, depth353 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l354
											}
											position++
											goto l353
										l354:
											position, tokenIndex, depth = position353, tokenIndex353, depth353
											if buffer[position] != rune('B') {
												goto l262
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
												goto l262
											}
											position++
										}
									l355:
										break
									case 'B', 'b':
										{
											position357, tokenIndex357, depth357 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l358
											}
											position++
											goto l357
										l358:
											position, tokenIndex, depth = position357, tokenIndex357, depth357
											if buffer[position] != rune('B') {
												goto l262
											}
											position++
										}
									l357:
										{
											position359, tokenIndex359, depth359 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l360
											}
											position++
											goto l359
										l360:
											position, tokenIndex, depth = position359, tokenIndex359, depth359
											if buffer[position] != rune('Y') {
												goto l262
											}
											position++
										}
									l359:
										break
									case 'A', 'a':
										{
											position361, tokenIndex361, depth361 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l362
											}
											position++
											goto l361
										l362:
											position, tokenIndex, depth = position361, tokenIndex361, depth361
											if buffer[position] != rune('A') {
												goto l262
											}
											position++
										}
									l361:
										{
											position363, tokenIndex363, depth363 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l364
											}
											position++
											goto l363
										l364:
											position, tokenIndex, depth = position363, tokenIndex363, depth363
											if buffer[position] != rune('S') {
												goto l262
											}
											position++
										}
									l363:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l262
										}
										break
									}
								}

							}
						l264:
							depth--
							add(ruleKEYWORD, position263)
						}
						{
							position365, tokenIndex365, depth365 := position, tokenIndex, depth
							if !_rules[ruleID_CONT]() {
								goto l365
							}
							goto l262
						l365:
							position, tokenIndex, depth = position365, tokenIndex365, depth365
						}
						goto l256
					l262:
						position, tokenIndex, depth = position262, tokenIndex262, depth262
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l256
					}
				l366:
					{
						position367, tokenIndex367, depth367 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l367
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l367
						}
						goto l366
					l367:
						position, tokenIndex, depth = position367, tokenIndex367, depth367
					}
				}
			l258:
				depth--
				add(ruleIDENTIFIER, position257)
			}
			return true
		l256:
			position, tokenIndex, depth = position256, tokenIndex256, depth256
			return false
		},
		/* 27 TIMESTAMP <- <((&('N' | 'n') <(('n' / 'N') ('o' / 'O') ('w' / 'W'))>) | (&('"' | '\'') STRING) | (&('-' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') <(NUMBER ([a-z] / [A-Z])?)>))> */
		nil,
		/* 28 ID_SEGMENT <- <(ID_START ID_CONT*)> */
		func() bool {
			position369, tokenIndex369, depth369 := position, tokenIndex, depth
			{
				position370 := position
				depth++
				if !_rules[ruleID_START]() {
					goto l369
				}
			l371:
				{
					position372, tokenIndex372, depth372 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l372
					}
					goto l371
				l372:
					position, tokenIndex, depth = position372, tokenIndex372, depth372
				}
				depth--
				add(ruleID_SEGMENT, position370)
			}
			return true
		l369:
			position, tokenIndex, depth = position369, tokenIndex369, depth369
			return false
		},
		/* 29 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position373, tokenIndex373, depth373 := position, tokenIndex, depth
			{
				position374 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l373
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l373
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l373
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position374)
			}
			return true
		l373:
			position, tokenIndex, depth = position373, tokenIndex373, depth373
			return false
		},
		/* 30 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position376, tokenIndex376, depth376 := position, tokenIndex, depth
			{
				position377 := position
				depth++
				{
					position378, tokenIndex378, depth378 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l379
					}
					goto l378
				l379:
					position, tokenIndex, depth = position378, tokenIndex378, depth378
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l376
					}
					position++
				}
			l378:
				depth--
				add(ruleID_CONT, position377)
			}
			return true
		l376:
			position, tokenIndex, depth = position376, tokenIndex376, depth376
			return false
		},
		/* 31 PROPERTY_KEY <- <((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> __ (('b' / 'B') ('y' / 'Y')))) | (&('R' | 'r') <(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))>) | (&('T' | 't') <(('t' / 'T') ('o' / 'O'))>) | (&('F' | 'f') <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))>))> */
		func() bool {
			position380, tokenIndex380, depth380 := position, tokenIndex, depth
			{
				position381 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position383 := position
							depth++
							{
								position384, tokenIndex384, depth384 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l385
								}
								position++
								goto l384
							l385:
								position, tokenIndex, depth = position384, tokenIndex384, depth384
								if buffer[position] != rune('S') {
									goto l380
								}
								position++
							}
						l384:
							{
								position386, tokenIndex386, depth386 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l387
								}
								position++
								goto l386
							l387:
								position, tokenIndex, depth = position386, tokenIndex386, depth386
								if buffer[position] != rune('A') {
									goto l380
								}
								position++
							}
						l386:
							{
								position388, tokenIndex388, depth388 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l389
								}
								position++
								goto l388
							l389:
								position, tokenIndex, depth = position388, tokenIndex388, depth388
								if buffer[position] != rune('M') {
									goto l380
								}
								position++
							}
						l388:
							{
								position390, tokenIndex390, depth390 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l391
								}
								position++
								goto l390
							l391:
								position, tokenIndex, depth = position390, tokenIndex390, depth390
								if buffer[position] != rune('P') {
									goto l380
								}
								position++
							}
						l390:
							{
								position392, tokenIndex392, depth392 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l393
								}
								position++
								goto l392
							l393:
								position, tokenIndex, depth = position392, tokenIndex392, depth392
								if buffer[position] != rune('L') {
									goto l380
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
									goto l380
								}
								position++
							}
						l394:
							depth--
							add(rulePegText, position383)
						}
						if !_rules[rule__]() {
							goto l380
						}
						{
							position396, tokenIndex396, depth396 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l397
							}
							position++
							goto l396
						l397:
							position, tokenIndex, depth = position396, tokenIndex396, depth396
							if buffer[position] != rune('B') {
								goto l380
							}
							position++
						}
					l396:
						{
							position398, tokenIndex398, depth398 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l399
							}
							position++
							goto l398
						l399:
							position, tokenIndex, depth = position398, tokenIndex398, depth398
							if buffer[position] != rune('Y') {
								goto l380
							}
							position++
						}
					l398:
						break
					case 'R', 'r':
						{
							position400 := position
							depth++
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
									goto l380
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
									goto l380
								}
								position++
							}
						l403:
							{
								position405, tokenIndex405, depth405 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l406
								}
								position++
								goto l405
							l406:
								position, tokenIndex, depth = position405, tokenIndex405, depth405
								if buffer[position] != rune('S') {
									goto l380
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
									goto l380
								}
								position++
							}
						l407:
							{
								position409, tokenIndex409, depth409 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l410
								}
								position++
								goto l409
							l410:
								position, tokenIndex, depth = position409, tokenIndex409, depth409
								if buffer[position] != rune('L') {
									goto l380
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
									goto l380
								}
								position++
							}
						l411:
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
									goto l380
								}
								position++
							}
						l413:
							{
								position415, tokenIndex415, depth415 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l416
								}
								position++
								goto l415
							l416:
								position, tokenIndex, depth = position415, tokenIndex415, depth415
								if buffer[position] != rune('I') {
									goto l380
								}
								position++
							}
						l415:
							{
								position417, tokenIndex417, depth417 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l418
								}
								position++
								goto l417
							l418:
								position, tokenIndex, depth = position417, tokenIndex417, depth417
								if buffer[position] != rune('O') {
									goto l380
								}
								position++
							}
						l417:
							{
								position419, tokenIndex419, depth419 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l420
								}
								position++
								goto l419
							l420:
								position, tokenIndex, depth = position419, tokenIndex419, depth419
								if buffer[position] != rune('N') {
									goto l380
								}
								position++
							}
						l419:
							depth--
							add(rulePegText, position400)
						}
						break
					case 'T', 't':
						{
							position421 := position
							depth++
							{
								position422, tokenIndex422, depth422 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l423
								}
								position++
								goto l422
							l423:
								position, tokenIndex, depth = position422, tokenIndex422, depth422
								if buffer[position] != rune('T') {
									goto l380
								}
								position++
							}
						l422:
							{
								position424, tokenIndex424, depth424 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l425
								}
								position++
								goto l424
							l425:
								position, tokenIndex, depth = position424, tokenIndex424, depth424
								if buffer[position] != rune('O') {
									goto l380
								}
								position++
							}
						l424:
							depth--
							add(rulePegText, position421)
						}
						break
					default:
						{
							position426 := position
							depth++
							{
								position427, tokenIndex427, depth427 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l428
								}
								position++
								goto l427
							l428:
								position, tokenIndex, depth = position427, tokenIndex427, depth427
								if buffer[position] != rune('F') {
									goto l380
								}
								position++
							}
						l427:
							{
								position429, tokenIndex429, depth429 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l430
								}
								position++
								goto l429
							l430:
								position, tokenIndex, depth = position429, tokenIndex429, depth429
								if buffer[position] != rune('R') {
									goto l380
								}
								position++
							}
						l429:
							{
								position431, tokenIndex431, depth431 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l432
								}
								position++
								goto l431
							l432:
								position, tokenIndex, depth = position431, tokenIndex431, depth431
								if buffer[position] != rune('O') {
									goto l380
								}
								position++
							}
						l431:
							{
								position433, tokenIndex433, depth433 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l434
								}
								position++
								goto l433
							l434:
								position, tokenIndex, depth = position433, tokenIndex433, depth433
								if buffer[position] != rune('M') {
									goto l380
								}
								position++
							}
						l433:
							depth--
							add(rulePegText, position426)
						}
						break
					}
				}

				depth--
				add(rulePROPERTY_KEY, position381)
			}
			return true
		l380:
			position, tokenIndex, depth = position380, tokenIndex380, depth380
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
			position444, tokenIndex444, depth444 := position, tokenIndex, depth
			{
				position445 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l444
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position445)
			}
			return true
		l444:
			position, tokenIndex, depth = position444, tokenIndex444, depth444
			return false
		},
		/* 42 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position446, tokenIndex446, depth446 := position, tokenIndex, depth
			{
				position447 := position
				depth++
				if buffer[position] != rune('"') {
					goto l446
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position447)
			}
			return true
		l446:
			position, tokenIndex, depth = position446, tokenIndex446, depth446
			return false
		},
		/* 43 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> QUOTE_SINGLE) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> QUOTE_DOUBLE))> */
		func() bool {
			position448, tokenIndex448, depth448 := position, tokenIndex, depth
			{
				position449 := position
				depth++
				{
					position450, tokenIndex450, depth450 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l451
					}
					{
						position452 := position
						depth++
					l453:
						{
							position454, tokenIndex454, depth454 := position, tokenIndex, depth
							{
								position455, tokenIndex455, depth455 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l455
								}
								goto l454
							l455:
								position, tokenIndex, depth = position455, tokenIndex455, depth455
							}
							if !_rules[ruleCHAR]() {
								goto l454
							}
							goto l453
						l454:
							position, tokenIndex, depth = position454, tokenIndex454, depth454
						}
						depth--
						add(rulePegText, position452)
					}
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l451
					}
					goto l450
				l451:
					position, tokenIndex, depth = position450, tokenIndex450, depth450
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l448
					}
					{
						position456 := position
						depth++
					l457:
						{
							position458, tokenIndex458, depth458 := position, tokenIndex, depth
							{
								position459, tokenIndex459, depth459 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l459
								}
								goto l458
							l459:
								position, tokenIndex, depth = position459, tokenIndex459, depth459
							}
							if !_rules[ruleCHAR]() {
								goto l458
							}
							goto l457
						l458:
							position, tokenIndex, depth = position458, tokenIndex458, depth458
						}
						depth--
						add(rulePegText, position456)
					}
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l448
					}
				}
			l450:
				depth--
				add(ruleSTRING, position449)
			}
			return true
		l448:
			position, tokenIndex, depth = position448, tokenIndex448, depth448
			return false
		},
		/* 44 CHAR <- <(('\\' ((&('"') QUOTE_DOUBLE) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position460, tokenIndex460, depth460 := position, tokenIndex, depth
			{
				position461 := position
				depth++
				{
					position462, tokenIndex462, depth462 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l463
					}
					position++
					{
						switch buffer[position] {
						case '"':
							if !_rules[ruleQUOTE_DOUBLE]() {
								goto l463
							}
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l463
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l463
							}
							break
						}
					}

					goto l462
				l463:
					position, tokenIndex, depth = position462, tokenIndex462, depth462
					{
						position465, tokenIndex465, depth465 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l465
						}
						goto l460
					l465:
						position, tokenIndex, depth = position465, tokenIndex465, depth465
					}
					if !matchDot() {
						goto l460
					}
				}
			l462:
				depth--
				add(ruleCHAR, position461)
			}
			return true
		l460:
			position, tokenIndex, depth = position460, tokenIndex460, depth460
			return false
		},
		/* 45 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position466, tokenIndex466, depth466 := position, tokenIndex, depth
			{
				position467 := position
				depth++
				{
					position468, tokenIndex468, depth468 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l469
					}
					position++
					goto l468
				l469:
					position, tokenIndex, depth = position468, tokenIndex468, depth468
					if buffer[position] != rune('\\') {
						goto l466
					}
					position++
				}
			l468:
				depth--
				add(ruleESCAPE_CLASS, position467)
			}
			return true
		l466:
			position, tokenIndex, depth = position466, tokenIndex466, depth466
			return false
		},
		/* 46 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position470, tokenIndex470, depth470 := position, tokenIndex, depth
			{
				position471 := position
				depth++
				{
					position472 := position
					depth++
					{
						position473, tokenIndex473, depth473 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l473
						}
						position++
						goto l474
					l473:
						position, tokenIndex, depth = position473, tokenIndex473, depth473
					}
				l474:
					{
						position475 := position
						depth++
						{
							position476, tokenIndex476, depth476 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l477
							}
							position++
							goto l476
						l477:
							position, tokenIndex, depth = position476, tokenIndex476, depth476
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l470
							}
							position++
						l478:
							{
								position479, tokenIndex479, depth479 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l479
								}
								position++
								goto l478
							l479:
								position, tokenIndex, depth = position479, tokenIndex479, depth479
							}
						}
					l476:
						depth--
						add(ruleNUMBER_NATURAL, position475)
					}
					depth--
					add(ruleNUMBER_INTEGER, position472)
				}
				{
					position480, tokenIndex480, depth480 := position, tokenIndex, depth
					{
						position482 := position
						depth++
						if buffer[position] != rune('.') {
							goto l480
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l480
						}
						position++
					l483:
						{
							position484, tokenIndex484, depth484 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l484
							}
							position++
							goto l483
						l484:
							position, tokenIndex, depth = position484, tokenIndex484, depth484
						}
						depth--
						add(ruleNUMBER_FRACTION, position482)
					}
					goto l481
				l480:
					position, tokenIndex, depth = position480, tokenIndex480, depth480
				}
			l481:
				{
					position485, tokenIndex485, depth485 := position, tokenIndex, depth
					{
						position487 := position
						depth++
						{
							position488, tokenIndex488, depth488 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l489
							}
							position++
							goto l488
						l489:
							position, tokenIndex, depth = position488, tokenIndex488, depth488
							if buffer[position] != rune('E') {
								goto l485
							}
							position++
						}
					l488:
						{
							position490, tokenIndex490, depth490 := position, tokenIndex, depth
							{
								position492, tokenIndex492, depth492 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l493
								}
								position++
								goto l492
							l493:
								position, tokenIndex, depth = position492, tokenIndex492, depth492
								if buffer[position] != rune('-') {
									goto l490
								}
								position++
							}
						l492:
							goto l491
						l490:
							position, tokenIndex, depth = position490, tokenIndex490, depth490
						}
					l491:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l485
						}
						position++
					l494:
						{
							position495, tokenIndex495, depth495 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l495
							}
							position++
							goto l494
						l495:
							position, tokenIndex, depth = position495, tokenIndex495, depth495
						}
						depth--
						add(ruleNUMBER_EXP, position487)
					}
					goto l486
				l485:
					position, tokenIndex, depth = position485, tokenIndex485, depth485
				}
			l486:
				depth--
				add(ruleNUMBER, position471)
			}
			return true
		l470:
			position, tokenIndex, depth = position470, tokenIndex470, depth470
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
			position500, tokenIndex500, depth500 := position, tokenIndex, depth
			{
				position501 := position
				depth++
				if !_rules[rule_]() {
					goto l500
				}
				if buffer[position] != rune('(') {
					goto l500
				}
				position++
				if !_rules[rule_]() {
					goto l500
				}
				depth--
				add(rulePAREN_OPEN, position501)
			}
			return true
		l500:
			position, tokenIndex, depth = position500, tokenIndex500, depth500
			return false
		},
		/* 52 PAREN_CLOSE <- <(_ ')' _)> */
		func() bool {
			position502, tokenIndex502, depth502 := position, tokenIndex, depth
			{
				position503 := position
				depth++
				if !_rules[rule_]() {
					goto l502
				}
				if buffer[position] != rune(')') {
					goto l502
				}
				position++
				if !_rules[rule_]() {
					goto l502
				}
				depth--
				add(rulePAREN_CLOSE, position503)
			}
			return true
		l502:
			position, tokenIndex, depth = position502, tokenIndex502, depth502
			return false
		},
		/* 53 COMMA <- <(_ ',' _)> */
		func() bool {
			position504, tokenIndex504, depth504 := position, tokenIndex, depth
			{
				position505 := position
				depth++
				if !_rules[rule_]() {
					goto l504
				}
				if buffer[position] != rune(',') {
					goto l504
				}
				position++
				if !_rules[rule_]() {
					goto l504
				}
				depth--
				add(ruleCOMMA, position505)
			}
			return true
		l504:
			position, tokenIndex, depth = position504, tokenIndex504, depth504
			return false
		},
		/* 54 _ <- <SPACE*> */
		func() bool {
			{
				position507 := position
				depth++
			l508:
				{
					position509, tokenIndex509, depth509 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l509
					}
					goto l508
				l509:
					position, tokenIndex, depth = position509, tokenIndex509, depth509
				}
				depth--
				add(rule_, position507)
			}
			return true
		},
		/* 55 __ <- <SPACE+> */
		func() bool {
			position510, tokenIndex510, depth510 := position, tokenIndex, depth
			{
				position511 := position
				depth++
				if !_rules[ruleSPACE]() {
					goto l510
				}
			l512:
				{
					position513, tokenIndex513, depth513 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l513
					}
					goto l512
				l513:
					position, tokenIndex, depth = position513, tokenIndex513, depth513
				}
				depth--
				add(rule__, position511)
			}
			return true
		l510:
			position, tokenIndex, depth = position510, tokenIndex510, depth510
			return false
		},
		/* 56 SPACE <- <((&('\t') '\t') | (&('\n') '\n') | (&(' ') ' '))> */
		func() bool {
			position514, tokenIndex514, depth514 := position, tokenIndex, depth
			{
				position515 := position
				depth++
				{
					switch buffer[position] {
					case '\t':
						if buffer[position] != rune('\t') {
							goto l514
						}
						position++
						break
					case '\n':
						if buffer[position] != rune('\n') {
							goto l514
						}
						position++
						break
					default:
						if buffer[position] != rune(' ') {
							goto l514
						}
						position++
						break
					}
				}

				depth--
				add(ruleSPACE, position515)
			}
			return true
		l514:
			position, tokenIndex, depth = position514, tokenIndex514, depth514
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
