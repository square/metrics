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
		/* 10 expression_3 <- <(expression_function / ((&('"' | '\'') (STRING Action20)) | (&('-' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') (<NUMBER> Action19)) | (&('\t' | '\n' | ' ' | '(') (PAREN_OPEN expression_1 PAREN_CLOSE)) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | '`' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') expression_metric)))> */
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
							add(ruleAction21, position)
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l127
						}
						if !_rules[ruleexpressionList]() {
							goto l127
						}
						{
							add(ruleAction22, position)
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
									add(ruleAction27, position)
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
										add(ruleAction28, position)
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
							add(ruleAction23, position)
						}
						depth--
						add(ruleexpression_function, position128)
					}
					goto l126
				l127:
					position, tokenIndex, depth = position126, tokenIndex126, depth126
					{
						switch buffer[position] {
						case '"', '\'':
							if !_rules[ruleSTRING]() {
								goto l124
							}
							{
								add(ruleAction20, position)
							}
							break
						case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							{
								position158 := position
								depth++
								if !_rules[ruleNUMBER]() {
									goto l124
								}
								depth--
								add(rulePegText, position158)
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
								position160 := position
								depth++
								{
									position161 := position
									depth++
									if !_rules[ruleIDENTIFIER]() {
										goto l124
									}
									depth--
									add(rulePegText, position161)
								}
								{
									add(ruleAction24, position)
								}
								{
									position163, tokenIndex163, depth163 := position, tokenIndex, depth
									{
										position165, tokenIndex165, depth165 := position, tokenIndex, depth
										if buffer[position] != rune('[') {
											goto l166
										}
										position++
										if !_rules[rule_]() {
											goto l166
										}
										if !_rules[rulepredicate_1]() {
											goto l166
										}
										if !_rules[rule_]() {
											goto l166
										}
										if buffer[position] != rune(']') {
											goto l166
										}
										position++
										goto l165
									l166:
										position, tokenIndex, depth = position165, tokenIndex165, depth165
										{
											add(ruleAction25, position)
										}
									}
								l165:
									goto l164

									position, tokenIndex, depth = position163, tokenIndex163, depth163
								}
							l164:
								{
									add(ruleAction26, position)
								}
								depth--
								add(ruleexpression_metric, position160)
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
				position174 := position
				depth++
				{
					position175, tokenIndex175, depth175 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
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
							if buffer[position] != rune('o') {
								goto l179
							}
							position++
							goto l178
						l179:
							position, tokenIndex, depth = position178, tokenIndex178, depth178
							if buffer[position] != rune('O') {
								goto l176
							}
							position++
						}
					l178:
						{
							position180, tokenIndex180, depth180 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l181
							}
							position++
							goto l180
						l181:
							position, tokenIndex, depth = position180, tokenIndex180, depth180
							if buffer[position] != rune('R') {
								goto l176
							}
							position++
						}
					l180:
						if !_rules[rule_]() {
							goto l176
						}
						depth--
						add(ruleOP_OR, position177)
					}
					if !_rules[rulepredicate_1]() {
						goto l176
					}
					{
						add(ruleAction29, position)
					}
					goto l175
				l176:
					position, tokenIndex, depth = position175, tokenIndex175, depth175
					if !_rules[rulepredicate_2]() {
						goto l183
					}
					goto l175
				l183:
					position, tokenIndex, depth = position175, tokenIndex175, depth175
				}
			l175:
				depth--
				add(rulepredicate_1, position174)
			}
			return true
		},
		/* 16 predicate_2 <- <((predicate_3 OP_AND predicate_2 Action30) / predicate_3)> */
		func() bool {
			position184, tokenIndex184, depth184 := position, tokenIndex, depth
			{
				position185 := position
				depth++
				{
					position186, tokenIndex186, depth186 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l187
					}
					{
						position188 := position
						depth++
						if !_rules[rule_]() {
							goto l187
						}
						{
							position189, tokenIndex189, depth189 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l190
							}
							position++
							goto l189
						l190:
							position, tokenIndex, depth = position189, tokenIndex189, depth189
							if buffer[position] != rune('A') {
								goto l187
							}
							position++
						}
					l189:
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
								goto l187
							}
							position++
						}
					l191:
						{
							position193, tokenIndex193, depth193 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l194
							}
							position++
							goto l193
						l194:
							position, tokenIndex, depth = position193, tokenIndex193, depth193
							if buffer[position] != rune('D') {
								goto l187
							}
							position++
						}
					l193:
						if !_rules[rule_]() {
							goto l187
						}
						depth--
						add(ruleOP_AND, position188)
					}
					if !_rules[rulepredicate_2]() {
						goto l187
					}
					{
						add(ruleAction30, position)
					}
					goto l186
				l187:
					position, tokenIndex, depth = position186, tokenIndex186, depth186
					if !_rules[rulepredicate_3]() {
						goto l184
					}
				}
			l186:
				depth--
				add(rulepredicate_2, position185)
			}
			return true
		l184:
			position, tokenIndex, depth = position184, tokenIndex184, depth184
			return false
		},
		/* 17 predicate_3 <- <((OP_NOT predicate_3 Action31) / (PAREN_OPEN predicate_1 PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position196, tokenIndex196, depth196 := position, tokenIndex, depth
			{
				position197 := position
				depth++
				{
					position198, tokenIndex198, depth198 := position, tokenIndex, depth
					{
						position200 := position
						depth++
						{
							position201, tokenIndex201, depth201 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l202
							}
							position++
							goto l201
						l202:
							position, tokenIndex, depth = position201, tokenIndex201, depth201
							if buffer[position] != rune('N') {
								goto l199
							}
							position++
						}
					l201:
						{
							position203, tokenIndex203, depth203 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l204
							}
							position++
							goto l203
						l204:
							position, tokenIndex, depth = position203, tokenIndex203, depth203
							if buffer[position] != rune('O') {
								goto l199
							}
							position++
						}
					l203:
						{
							position205, tokenIndex205, depth205 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l206
							}
							position++
							goto l205
						l206:
							position, tokenIndex, depth = position205, tokenIndex205, depth205
							if buffer[position] != rune('T') {
								goto l199
							}
							position++
						}
					l205:
						if !_rules[rule__]() {
							goto l199
						}
						depth--
						add(ruleOP_NOT, position200)
					}
					if !_rules[rulepredicate_3]() {
						goto l199
					}
					{
						add(ruleAction31, position)
					}
					goto l198
				l199:
					position, tokenIndex, depth = position198, tokenIndex198, depth198
					if !_rules[rulePAREN_OPEN]() {
						goto l208
					}
					if !_rules[rulepredicate_1]() {
						goto l208
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l208
					}
					goto l198
				l208:
					position, tokenIndex, depth = position198, tokenIndex198, depth198
					{
						position209 := position
						depth++
						{
							position210, tokenIndex210, depth210 := position, tokenIndex, depth
							if !_rules[ruletagName]() {
								goto l211
							}
							if !_rules[rule_]() {
								goto l211
							}
							if buffer[position] != rune('=') {
								goto l211
							}
							position++
							if !_rules[rule_]() {
								goto l211
							}
							if !_rules[ruleliteralString]() {
								goto l211
							}
							{
								add(ruleAction32, position)
							}
							goto l210
						l211:
							position, tokenIndex, depth = position210, tokenIndex210, depth210
							if !_rules[ruletagName]() {
								goto l213
							}
							if !_rules[rule_]() {
								goto l213
							}
							if buffer[position] != rune('!') {
								goto l213
							}
							position++
							if buffer[position] != rune('=') {
								goto l213
							}
							position++
							if !_rules[rule_]() {
								goto l213
							}
							if !_rules[ruleliteralString]() {
								goto l213
							}
							{
								add(ruleAction33, position)
							}
							goto l210
						l213:
							position, tokenIndex, depth = position210, tokenIndex210, depth210
							if !_rules[ruletagName]() {
								goto l215
							}
							if !_rules[rule__]() {
								goto l215
							}
							{
								position216, tokenIndex216, depth216 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l217
								}
								position++
								goto l216
							l217:
								position, tokenIndex, depth = position216, tokenIndex216, depth216
								if buffer[position] != rune('M') {
									goto l215
								}
								position++
							}
						l216:
							{
								position218, tokenIndex218, depth218 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l219
								}
								position++
								goto l218
							l219:
								position, tokenIndex, depth = position218, tokenIndex218, depth218
								if buffer[position] != rune('A') {
									goto l215
								}
								position++
							}
						l218:
							{
								position220, tokenIndex220, depth220 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l221
								}
								position++
								goto l220
							l221:
								position, tokenIndex, depth = position220, tokenIndex220, depth220
								if buffer[position] != rune('T') {
									goto l215
								}
								position++
							}
						l220:
							{
								position222, tokenIndex222, depth222 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l223
								}
								position++
								goto l222
							l223:
								position, tokenIndex, depth = position222, tokenIndex222, depth222
								if buffer[position] != rune('C') {
									goto l215
								}
								position++
							}
						l222:
							{
								position224, tokenIndex224, depth224 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l225
								}
								position++
								goto l224
							l225:
								position, tokenIndex, depth = position224, tokenIndex224, depth224
								if buffer[position] != rune('H') {
									goto l215
								}
								position++
							}
						l224:
							{
								position226, tokenIndex226, depth226 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l227
								}
								position++
								goto l226
							l227:
								position, tokenIndex, depth = position226, tokenIndex226, depth226
								if buffer[position] != rune('E') {
									goto l215
								}
								position++
							}
						l226:
							{
								position228, tokenIndex228, depth228 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l229
								}
								position++
								goto l228
							l229:
								position, tokenIndex, depth = position228, tokenIndex228, depth228
								if buffer[position] != rune('S') {
									goto l215
								}
								position++
							}
						l228:
							if !_rules[rule__]() {
								goto l215
							}
							if !_rules[ruleliteralString]() {
								goto l215
							}
							{
								add(ruleAction34, position)
							}
							goto l210
						l215:
							position, tokenIndex, depth = position210, tokenIndex210, depth210
							if !_rules[ruletagName]() {
								goto l196
							}
							if !_rules[rule__]() {
								goto l196
							}
							{
								position231, tokenIndex231, depth231 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l232
								}
								position++
								goto l231
							l232:
								position, tokenIndex, depth = position231, tokenIndex231, depth231
								if buffer[position] != rune('I') {
									goto l196
								}
								position++
							}
						l231:
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
									goto l196
								}
								position++
							}
						l233:
							if !_rules[rule__]() {
								goto l196
							}
							{
								position235 := position
								depth++
								{
									add(ruleAction37, position)
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l196
								}
								if !_rules[ruleliteralListString]() {
									goto l196
								}
							l237:
								{
									position238, tokenIndex238, depth238 := position, tokenIndex, depth
									if !_rules[ruleCOMMA]() {
										goto l238
									}
									if !_rules[ruleliteralListString]() {
										goto l238
									}
									goto l237
								l238:
									position, tokenIndex, depth = position238, tokenIndex238, depth238
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l196
								}
								depth--
								add(ruleliteralList, position235)
							}
							{
								add(ruleAction35, position)
							}
						}
					l210:
						depth--
						add(ruletagMatcher, position209)
					}
				}
			l198:
				depth--
				add(rulepredicate_3, position197)
			}
			return true
		l196:
			position, tokenIndex, depth = position196, tokenIndex196, depth196
			return false
		},
		/* 18 tagMatcher <- <((tagName _ '=' _ literalString Action32) / (tagName _ ('!' '=') _ literalString Action33) / (tagName __ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) __ literalString Action34) / (tagName __ (('i' / 'I') ('n' / 'N')) __ literalList Action35))> */
		nil,
		/* 19 literalString <- <(STRING Action36)> */
		func() bool {
			position241, tokenIndex241, depth241 := position, tokenIndex, depth
			{
				position242 := position
				depth++
				if !_rules[ruleSTRING]() {
					goto l241
				}
				{
					add(ruleAction36, position)
				}
				depth--
				add(ruleliteralString, position242)
			}
			return true
		l241:
			position, tokenIndex, depth = position241, tokenIndex241, depth241
			return false
		},
		/* 20 literalList <- <(Action37 PAREN_OPEN literalListString (COMMA literalListString)* PAREN_CLOSE)> */
		nil,
		/* 21 literalListString <- <(STRING Action38)> */
		func() bool {
			position245, tokenIndex245, depth245 := position, tokenIndex, depth
			{
				position246 := position
				depth++
				if !_rules[ruleSTRING]() {
					goto l245
				}
				{
					add(ruleAction38, position)
				}
				depth--
				add(ruleliteralListString, position246)
			}
			return true
		l245:
			position, tokenIndex, depth = position245, tokenIndex245, depth245
			return false
		},
		/* 22 tagName <- <(<TAG_NAME> Action39)> */
		func() bool {
			position248, tokenIndex248, depth248 := position, tokenIndex, depth
			{
				position249 := position
				depth++
				{
					position250 := position
					depth++
					{
						position251 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l248
						}
						depth--
						add(ruleTAG_NAME, position251)
					}
					depth--
					add(rulePegText, position250)
				}
				{
					add(ruleAction39, position)
				}
				depth--
				add(ruletagName, position249)
			}
			return true
		l248:
			position, tokenIndex, depth = position248, tokenIndex248, depth248
			return false
		},
		/* 23 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position253, tokenIndex253, depth253 := position, tokenIndex, depth
			{
				position254 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l253
				}
				depth--
				add(ruleCOLUMN_NAME, position254)
			}
			return true
		l253:
			position, tokenIndex, depth = position253, tokenIndex253, depth253
			return false
		},
		/* 24 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 25 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 26 IDENTIFIER <- <(('`' CHAR* '`') / (!(KEYWORD !ID_CONT) ID_SEGMENT ('.' ID_SEGMENT)*))> */
		func() bool {
			position257, tokenIndex257, depth257 := position, tokenIndex, depth
			{
				position258 := position
				depth++
				{
					position259, tokenIndex259, depth259 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l260
					}
					position++
				l261:
					{
						position262, tokenIndex262, depth262 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l262
						}
						goto l261
					l262:
						position, tokenIndex, depth = position262, tokenIndex262, depth262
					}
					if buffer[position] != rune('`') {
						goto l260
					}
					position++
					goto l259
				l260:
					position, tokenIndex, depth = position259, tokenIndex259, depth259
					{
						position263, tokenIndex263, depth263 := position, tokenIndex, depth
						{
							position264 := position
							depth++
							{
								position265, tokenIndex265, depth265 := position, tokenIndex, depth
								{
									position267, tokenIndex267, depth267 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l268
									}
									position++
									goto l267
								l268:
									position, tokenIndex, depth = position267, tokenIndex267, depth267
									if buffer[position] != rune('A') {
										goto l266
									}
									position++
								}
							l267:
								{
									position269, tokenIndex269, depth269 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l270
									}
									position++
									goto l269
								l270:
									position, tokenIndex, depth = position269, tokenIndex269, depth269
									if buffer[position] != rune('L') {
										goto l266
									}
									position++
								}
							l269:
								{
									position271, tokenIndex271, depth271 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l272
									}
									position++
									goto l271
								l272:
									position, tokenIndex, depth = position271, tokenIndex271, depth271
									if buffer[position] != rune('L') {
										goto l266
									}
									position++
								}
							l271:
								goto l265
							l266:
								position, tokenIndex, depth = position265, tokenIndex265, depth265
								{
									position274, tokenIndex274, depth274 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l275
									}
									position++
									goto l274
								l275:
									position, tokenIndex, depth = position274, tokenIndex274, depth274
									if buffer[position] != rune('A') {
										goto l273
									}
									position++
								}
							l274:
								{
									position276, tokenIndex276, depth276 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l277
									}
									position++
									goto l276
								l277:
									position, tokenIndex, depth = position276, tokenIndex276, depth276
									if buffer[position] != rune('N') {
										goto l273
									}
									position++
								}
							l276:
								{
									position278, tokenIndex278, depth278 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l279
									}
									position++
									goto l278
								l279:
									position, tokenIndex, depth = position278, tokenIndex278, depth278
									if buffer[position] != rune('D') {
										goto l273
									}
									position++
								}
							l278:
								goto l265
							l273:
								position, tokenIndex, depth = position265, tokenIndex265, depth265
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
										goto l280
									}
									position++
								}
							l281:
								{
									position283, tokenIndex283, depth283 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l284
									}
									position++
									goto l283
								l284:
									position, tokenIndex, depth = position283, tokenIndex283, depth283
									if buffer[position] != rune('E') {
										goto l280
									}
									position++
								}
							l283:
								{
									position285, tokenIndex285, depth285 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l286
									}
									position++
									goto l285
								l286:
									position, tokenIndex, depth = position285, tokenIndex285, depth285
									if buffer[position] != rune('L') {
										goto l280
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
										goto l280
									}
									position++
								}
							l287:
								{
									position289, tokenIndex289, depth289 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l290
									}
									position++
									goto l289
								l290:
									position, tokenIndex, depth = position289, tokenIndex289, depth289
									if buffer[position] != rune('C') {
										goto l280
									}
									position++
								}
							l289:
								{
									position291, tokenIndex291, depth291 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l292
									}
									position++
									goto l291
								l292:
									position, tokenIndex, depth = position291, tokenIndex291, depth291
									if buffer[position] != rune('T') {
										goto l280
									}
									position++
								}
							l291:
								goto l265
							l280:
								position, tokenIndex, depth = position265, tokenIndex265, depth265
								{
									switch buffer[position] {
									case 'W', 'w':
										{
											position294, tokenIndex294, depth294 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l295
											}
											position++
											goto l294
										l295:
											position, tokenIndex, depth = position294, tokenIndex294, depth294
											if buffer[position] != rune('W') {
												goto l263
											}
											position++
										}
									l294:
										{
											position296, tokenIndex296, depth296 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l297
											}
											position++
											goto l296
										l297:
											position, tokenIndex, depth = position296, tokenIndex296, depth296
											if buffer[position] != rune('H') {
												goto l263
											}
											position++
										}
									l296:
										{
											position298, tokenIndex298, depth298 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l299
											}
											position++
											goto l298
										l299:
											position, tokenIndex, depth = position298, tokenIndex298, depth298
											if buffer[position] != rune('E') {
												goto l263
											}
											position++
										}
									l298:
										{
											position300, tokenIndex300, depth300 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l301
											}
											position++
											goto l300
										l301:
											position, tokenIndex, depth = position300, tokenIndex300, depth300
											if buffer[position] != rune('R') {
												goto l263
											}
											position++
										}
									l300:
										{
											position302, tokenIndex302, depth302 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l303
											}
											position++
											goto l302
										l303:
											position, tokenIndex, depth = position302, tokenIndex302, depth302
											if buffer[position] != rune('E') {
												goto l263
											}
											position++
										}
									l302:
										break
									case 'O', 'o':
										{
											position304, tokenIndex304, depth304 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l305
											}
											position++
											goto l304
										l305:
											position, tokenIndex, depth = position304, tokenIndex304, depth304
											if buffer[position] != rune('O') {
												goto l263
											}
											position++
										}
									l304:
										{
											position306, tokenIndex306, depth306 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l307
											}
											position++
											goto l306
										l307:
											position, tokenIndex, depth = position306, tokenIndex306, depth306
											if buffer[position] != rune('R') {
												goto l263
											}
											position++
										}
									l306:
										break
									case 'N', 'n':
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
												goto l263
											}
											position++
										}
									l308:
										{
											position310, tokenIndex310, depth310 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l311
											}
											position++
											goto l310
										l311:
											position, tokenIndex, depth = position310, tokenIndex310, depth310
											if buffer[position] != rune('O') {
												goto l263
											}
											position++
										}
									l310:
										{
											position312, tokenIndex312, depth312 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l313
											}
											position++
											goto l312
										l313:
											position, tokenIndex, depth = position312, tokenIndex312, depth312
											if buffer[position] != rune('T') {
												goto l263
											}
											position++
										}
									l312:
										break
									case 'M', 'm':
										{
											position314, tokenIndex314, depth314 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l315
											}
											position++
											goto l314
										l315:
											position, tokenIndex, depth = position314, tokenIndex314, depth314
											if buffer[position] != rune('M') {
												goto l263
											}
											position++
										}
									l314:
										{
											position316, tokenIndex316, depth316 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l317
											}
											position++
											goto l316
										l317:
											position, tokenIndex, depth = position316, tokenIndex316, depth316
											if buffer[position] != rune('A') {
												goto l263
											}
											position++
										}
									l316:
										{
											position318, tokenIndex318, depth318 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l319
											}
											position++
											goto l318
										l319:
											position, tokenIndex, depth = position318, tokenIndex318, depth318
											if buffer[position] != rune('T') {
												goto l263
											}
											position++
										}
									l318:
										{
											position320, tokenIndex320, depth320 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l321
											}
											position++
											goto l320
										l321:
											position, tokenIndex, depth = position320, tokenIndex320, depth320
											if buffer[position] != rune('C') {
												goto l263
											}
											position++
										}
									l320:
										{
											position322, tokenIndex322, depth322 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l323
											}
											position++
											goto l322
										l323:
											position, tokenIndex, depth = position322, tokenIndex322, depth322
											if buffer[position] != rune('H') {
												goto l263
											}
											position++
										}
									l322:
										{
											position324, tokenIndex324, depth324 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l325
											}
											position++
											goto l324
										l325:
											position, tokenIndex, depth = position324, tokenIndex324, depth324
											if buffer[position] != rune('E') {
												goto l263
											}
											position++
										}
									l324:
										{
											position326, tokenIndex326, depth326 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l327
											}
											position++
											goto l326
										l327:
											position, tokenIndex, depth = position326, tokenIndex326, depth326
											if buffer[position] != rune('S') {
												goto l263
											}
											position++
										}
									l326:
										break
									case 'I', 'i':
										{
											position328, tokenIndex328, depth328 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l329
											}
											position++
											goto l328
										l329:
											position, tokenIndex, depth = position328, tokenIndex328, depth328
											if buffer[position] != rune('I') {
												goto l263
											}
											position++
										}
									l328:
										{
											position330, tokenIndex330, depth330 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l331
											}
											position++
											goto l330
										l331:
											position, tokenIndex, depth = position330, tokenIndex330, depth330
											if buffer[position] != rune('N') {
												goto l263
											}
											position++
										}
									l330:
										break
									case 'G', 'g':
										{
											position332, tokenIndex332, depth332 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l333
											}
											position++
											goto l332
										l333:
											position, tokenIndex, depth = position332, tokenIndex332, depth332
											if buffer[position] != rune('G') {
												goto l263
											}
											position++
										}
									l332:
										{
											position334, tokenIndex334, depth334 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l335
											}
											position++
											goto l334
										l335:
											position, tokenIndex, depth = position334, tokenIndex334, depth334
											if buffer[position] != rune('R') {
												goto l263
											}
											position++
										}
									l334:
										{
											position336, tokenIndex336, depth336 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l337
											}
											position++
											goto l336
										l337:
											position, tokenIndex, depth = position336, tokenIndex336, depth336
											if buffer[position] != rune('O') {
												goto l263
											}
											position++
										}
									l336:
										{
											position338, tokenIndex338, depth338 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l339
											}
											position++
											goto l338
										l339:
											position, tokenIndex, depth = position338, tokenIndex338, depth338
											if buffer[position] != rune('U') {
												goto l263
											}
											position++
										}
									l338:
										{
											position340, tokenIndex340, depth340 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l341
											}
											position++
											goto l340
										l341:
											position, tokenIndex, depth = position340, tokenIndex340, depth340
											if buffer[position] != rune('P') {
												goto l263
											}
											position++
										}
									l340:
										break
									case 'D', 'd':
										{
											position342, tokenIndex342, depth342 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l343
											}
											position++
											goto l342
										l343:
											position, tokenIndex, depth = position342, tokenIndex342, depth342
											if buffer[position] != rune('D') {
												goto l263
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
												goto l263
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
												goto l263
											}
											position++
										}
									l346:
										{
											position348, tokenIndex348, depth348 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l349
											}
											position++
											goto l348
										l349:
											position, tokenIndex, depth = position348, tokenIndex348, depth348
											if buffer[position] != rune('C') {
												goto l263
											}
											position++
										}
									l348:
										{
											position350, tokenIndex350, depth350 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l351
											}
											position++
											goto l350
										l351:
											position, tokenIndex, depth = position350, tokenIndex350, depth350
											if buffer[position] != rune('R') {
												goto l263
											}
											position++
										}
									l350:
										{
											position352, tokenIndex352, depth352 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l353
											}
											position++
											goto l352
										l353:
											position, tokenIndex, depth = position352, tokenIndex352, depth352
											if buffer[position] != rune('I') {
												goto l263
											}
											position++
										}
									l352:
										{
											position354, tokenIndex354, depth354 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l355
											}
											position++
											goto l354
										l355:
											position, tokenIndex, depth = position354, tokenIndex354, depth354
											if buffer[position] != rune('B') {
												goto l263
											}
											position++
										}
									l354:
										{
											position356, tokenIndex356, depth356 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l357
											}
											position++
											goto l356
										l357:
											position, tokenIndex, depth = position356, tokenIndex356, depth356
											if buffer[position] != rune('E') {
												goto l263
											}
											position++
										}
									l356:
										break
									case 'B', 'b':
										{
											position358, tokenIndex358, depth358 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l359
											}
											position++
											goto l358
										l359:
											position, tokenIndex, depth = position358, tokenIndex358, depth358
											if buffer[position] != rune('B') {
												goto l263
											}
											position++
										}
									l358:
										{
											position360, tokenIndex360, depth360 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l361
											}
											position++
											goto l360
										l361:
											position, tokenIndex, depth = position360, tokenIndex360, depth360
											if buffer[position] != rune('Y') {
												goto l263
											}
											position++
										}
									l360:
										break
									case 'A', 'a':
										{
											position362, tokenIndex362, depth362 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l363
											}
											position++
											goto l362
										l363:
											position, tokenIndex, depth = position362, tokenIndex362, depth362
											if buffer[position] != rune('A') {
												goto l263
											}
											position++
										}
									l362:
										{
											position364, tokenIndex364, depth364 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l365
											}
											position++
											goto l364
										l365:
											position, tokenIndex, depth = position364, tokenIndex364, depth364
											if buffer[position] != rune('S') {
												goto l263
											}
											position++
										}
									l364:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l263
										}
										break
									}
								}

							}
						l265:
							depth--
							add(ruleKEYWORD, position264)
						}
						{
							position366, tokenIndex366, depth366 := position, tokenIndex, depth
							if !_rules[ruleID_CONT]() {
								goto l366
							}
							goto l263
						l366:
							position, tokenIndex, depth = position366, tokenIndex366, depth366
						}
						goto l257
					l263:
						position, tokenIndex, depth = position263, tokenIndex263, depth263
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l257
					}
				l367:
					{
						position368, tokenIndex368, depth368 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l368
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l368
						}
						goto l367
					l368:
						position, tokenIndex, depth = position368, tokenIndex368, depth368
					}
				}
			l259:
				depth--
				add(ruleIDENTIFIER, position258)
			}
			return true
		l257:
			position, tokenIndex, depth = position257, tokenIndex257, depth257
			return false
		},
		/* 27 TIMESTAMP <- <((&('N' | 'n') <(('n' / 'N') ('o' / 'O') ('w' / 'W'))>) | (&('"' | '\'') STRING) | (&('-' | '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') <(NUMBER ([a-z] / [A-Z])?)>))> */
		nil,
		/* 28 ID_SEGMENT <- <(ID_START ID_CONT*)> */
		func() bool {
			position370, tokenIndex370, depth370 := position, tokenIndex, depth
			{
				position371 := position
				depth++
				if !_rules[ruleID_START]() {
					goto l370
				}
			l372:
				{
					position373, tokenIndex373, depth373 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l373
					}
					goto l372
				l373:
					position, tokenIndex, depth = position373, tokenIndex373, depth373
				}
				depth--
				add(ruleID_SEGMENT, position371)
			}
			return true
		l370:
			position, tokenIndex, depth = position370, tokenIndex370, depth370
			return false
		},
		/* 29 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position374, tokenIndex374, depth374 := position, tokenIndex, depth
			{
				position375 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l374
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l374
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l374
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position375)
			}
			return true
		l374:
			position, tokenIndex, depth = position374, tokenIndex374, depth374
			return false
		},
		/* 30 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position377, tokenIndex377, depth377 := position, tokenIndex, depth
			{
				position378 := position
				depth++
				{
					position379, tokenIndex379, depth379 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l380
					}
					goto l379
				l380:
					position, tokenIndex, depth = position379, tokenIndex379, depth379
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l377
					}
					position++
				}
			l379:
				depth--
				add(ruleID_CONT, position378)
			}
			return true
		l377:
			position, tokenIndex, depth = position377, tokenIndex377, depth377
			return false
		},
		/* 31 PROPERTY_KEY <- <((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> __ (('b' / 'B') ('y' / 'Y')))) | (&('R' | 'r') <(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))>) | (&('T' | 't') <(('t' / 'T') ('o' / 'O'))>) | (&('F' | 'f') <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))>))> */
		func() bool {
			position381, tokenIndex381, depth381 := position, tokenIndex, depth
			{
				position382 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position384 := position
							depth++
							{
								position385, tokenIndex385, depth385 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l386
								}
								position++
								goto l385
							l386:
								position, tokenIndex, depth = position385, tokenIndex385, depth385
								if buffer[position] != rune('S') {
									goto l381
								}
								position++
							}
						l385:
							{
								position387, tokenIndex387, depth387 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l388
								}
								position++
								goto l387
							l388:
								position, tokenIndex, depth = position387, tokenIndex387, depth387
								if buffer[position] != rune('A') {
									goto l381
								}
								position++
							}
						l387:
							{
								position389, tokenIndex389, depth389 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l390
								}
								position++
								goto l389
							l390:
								position, tokenIndex, depth = position389, tokenIndex389, depth389
								if buffer[position] != rune('M') {
									goto l381
								}
								position++
							}
						l389:
							{
								position391, tokenIndex391, depth391 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l392
								}
								position++
								goto l391
							l392:
								position, tokenIndex, depth = position391, tokenIndex391, depth391
								if buffer[position] != rune('P') {
									goto l381
								}
								position++
							}
						l391:
							{
								position393, tokenIndex393, depth393 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l394
								}
								position++
								goto l393
							l394:
								position, tokenIndex, depth = position393, tokenIndex393, depth393
								if buffer[position] != rune('L') {
									goto l381
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
									goto l381
								}
								position++
							}
						l395:
							depth--
							add(rulePegText, position384)
						}
						if !_rules[rule__]() {
							goto l381
						}
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
								goto l381
							}
							position++
						}
					l397:
						{
							position399, tokenIndex399, depth399 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l400
							}
							position++
							goto l399
						l400:
							position, tokenIndex, depth = position399, tokenIndex399, depth399
							if buffer[position] != rune('Y') {
								goto l381
							}
							position++
						}
					l399:
						break
					case 'R', 'r':
						{
							position401 := position
							depth++
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
									goto l381
								}
								position++
							}
						l402:
							{
								position404, tokenIndex404, depth404 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l405
								}
								position++
								goto l404
							l405:
								position, tokenIndex, depth = position404, tokenIndex404, depth404
								if buffer[position] != rune('E') {
									goto l381
								}
								position++
							}
						l404:
							{
								position406, tokenIndex406, depth406 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l407
								}
								position++
								goto l406
							l407:
								position, tokenIndex, depth = position406, tokenIndex406, depth406
								if buffer[position] != rune('S') {
									goto l381
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
									goto l381
								}
								position++
							}
						l408:
							{
								position410, tokenIndex410, depth410 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l411
								}
								position++
								goto l410
							l411:
								position, tokenIndex, depth = position410, tokenIndex410, depth410
								if buffer[position] != rune('L') {
									goto l381
								}
								position++
							}
						l410:
							{
								position412, tokenIndex412, depth412 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l413
								}
								position++
								goto l412
							l413:
								position, tokenIndex, depth = position412, tokenIndex412, depth412
								if buffer[position] != rune('U') {
									goto l381
								}
								position++
							}
						l412:
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
									goto l381
								}
								position++
							}
						l414:
							{
								position416, tokenIndex416, depth416 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l417
								}
								position++
								goto l416
							l417:
								position, tokenIndex, depth = position416, tokenIndex416, depth416
								if buffer[position] != rune('I') {
									goto l381
								}
								position++
							}
						l416:
							{
								position418, tokenIndex418, depth418 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l419
								}
								position++
								goto l418
							l419:
								position, tokenIndex, depth = position418, tokenIndex418, depth418
								if buffer[position] != rune('O') {
									goto l381
								}
								position++
							}
						l418:
							{
								position420, tokenIndex420, depth420 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l421
								}
								position++
								goto l420
							l421:
								position, tokenIndex, depth = position420, tokenIndex420, depth420
								if buffer[position] != rune('N') {
									goto l381
								}
								position++
							}
						l420:
							depth--
							add(rulePegText, position401)
						}
						break
					case 'T', 't':
						{
							position422 := position
							depth++
							{
								position423, tokenIndex423, depth423 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l424
								}
								position++
								goto l423
							l424:
								position, tokenIndex, depth = position423, tokenIndex423, depth423
								if buffer[position] != rune('T') {
									goto l381
								}
								position++
							}
						l423:
							{
								position425, tokenIndex425, depth425 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l426
								}
								position++
								goto l425
							l426:
								position, tokenIndex, depth = position425, tokenIndex425, depth425
								if buffer[position] != rune('O') {
									goto l381
								}
								position++
							}
						l425:
							depth--
							add(rulePegText, position422)
						}
						break
					default:
						{
							position427 := position
							depth++
							{
								position428, tokenIndex428, depth428 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l429
								}
								position++
								goto l428
							l429:
								position, tokenIndex, depth = position428, tokenIndex428, depth428
								if buffer[position] != rune('F') {
									goto l381
								}
								position++
							}
						l428:
							{
								position430, tokenIndex430, depth430 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l431
								}
								position++
								goto l430
							l431:
								position, tokenIndex, depth = position430, tokenIndex430, depth430
								if buffer[position] != rune('R') {
									goto l381
								}
								position++
							}
						l430:
							{
								position432, tokenIndex432, depth432 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l433
								}
								position++
								goto l432
							l433:
								position, tokenIndex, depth = position432, tokenIndex432, depth432
								if buffer[position] != rune('O') {
									goto l381
								}
								position++
							}
						l432:
							{
								position434, tokenIndex434, depth434 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l435
								}
								position++
								goto l434
							l435:
								position, tokenIndex, depth = position434, tokenIndex434, depth434
								if buffer[position] != rune('M') {
									goto l381
								}
								position++
							}
						l434:
							depth--
							add(rulePegText, position427)
						}
						break
					}
				}

				depth--
				add(rulePROPERTY_KEY, position382)
			}
			return true
		l381:
			position, tokenIndex, depth = position381, tokenIndex381, depth381
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
			position445, tokenIndex445, depth445 := position, tokenIndex, depth
			{
				position446 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l445
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position446)
			}
			return true
		l445:
			position, tokenIndex, depth = position445, tokenIndex445, depth445
			return false
		},
		/* 42 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position447, tokenIndex447, depth447 := position, tokenIndex, depth
			{
				position448 := position
				depth++
				if buffer[position] != rune('"') {
					goto l447
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position448)
			}
			return true
		l447:
			position, tokenIndex, depth = position447, tokenIndex447, depth447
			return false
		},
		/* 43 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> QUOTE_SINGLE) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> QUOTE_DOUBLE))> */
		func() bool {
			position449, tokenIndex449, depth449 := position, tokenIndex, depth
			{
				position450 := position
				depth++
				{
					position451, tokenIndex451, depth451 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l452
					}
					{
						position453 := position
						depth++
					l454:
						{
							position455, tokenIndex455, depth455 := position, tokenIndex, depth
							{
								position456, tokenIndex456, depth456 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l456
								}
								goto l455
							l456:
								position, tokenIndex, depth = position456, tokenIndex456, depth456
							}
							if !_rules[ruleCHAR]() {
								goto l455
							}
							goto l454
						l455:
							position, tokenIndex, depth = position455, tokenIndex455, depth455
						}
						depth--
						add(rulePegText, position453)
					}
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l452
					}
					goto l451
				l452:
					position, tokenIndex, depth = position451, tokenIndex451, depth451
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l449
					}
					{
						position457 := position
						depth++
					l458:
						{
							position459, tokenIndex459, depth459 := position, tokenIndex, depth
							{
								position460, tokenIndex460, depth460 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l460
								}
								goto l459
							l460:
								position, tokenIndex, depth = position460, tokenIndex460, depth460
							}
							if !_rules[ruleCHAR]() {
								goto l459
							}
							goto l458
						l459:
							position, tokenIndex, depth = position459, tokenIndex459, depth459
						}
						depth--
						add(rulePegText, position457)
					}
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l449
					}
				}
			l451:
				depth--
				add(ruleSTRING, position450)
			}
			return true
		l449:
			position, tokenIndex, depth = position449, tokenIndex449, depth449
			return false
		},
		/* 44 CHAR <- <(('\\' ((&('"') QUOTE_DOUBLE) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position461, tokenIndex461, depth461 := position, tokenIndex, depth
			{
				position462 := position
				depth++
				{
					position463, tokenIndex463, depth463 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l464
					}
					position++
					{
						switch buffer[position] {
						case '"':
							if !_rules[ruleQUOTE_DOUBLE]() {
								goto l464
							}
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l464
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l464
							}
							break
						}
					}

					goto l463
				l464:
					position, tokenIndex, depth = position463, tokenIndex463, depth463
					{
						position466, tokenIndex466, depth466 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l466
						}
						goto l461
					l466:
						position, tokenIndex, depth = position466, tokenIndex466, depth466
					}
					if !matchDot() {
						goto l461
					}
				}
			l463:
				depth--
				add(ruleCHAR, position462)
			}
			return true
		l461:
			position, tokenIndex, depth = position461, tokenIndex461, depth461
			return false
		},
		/* 45 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position467, tokenIndex467, depth467 := position, tokenIndex, depth
			{
				position468 := position
				depth++
				{
					position469, tokenIndex469, depth469 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l470
					}
					position++
					goto l469
				l470:
					position, tokenIndex, depth = position469, tokenIndex469, depth469
					if buffer[position] != rune('\\') {
						goto l467
					}
					position++
				}
			l469:
				depth--
				add(ruleESCAPE_CLASS, position468)
			}
			return true
		l467:
			position, tokenIndex, depth = position467, tokenIndex467, depth467
			return false
		},
		/* 46 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position471, tokenIndex471, depth471 := position, tokenIndex, depth
			{
				position472 := position
				depth++
				{
					position473 := position
					depth++
					{
						position474, tokenIndex474, depth474 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l474
						}
						position++
						goto l475
					l474:
						position, tokenIndex, depth = position474, tokenIndex474, depth474
					}
				l475:
					{
						position476 := position
						depth++
						{
							position477, tokenIndex477, depth477 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l478
							}
							position++
							goto l477
						l478:
							position, tokenIndex, depth = position477, tokenIndex477, depth477
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l471
							}
							position++
						l479:
							{
								position480, tokenIndex480, depth480 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l480
								}
								position++
								goto l479
							l480:
								position, tokenIndex, depth = position480, tokenIndex480, depth480
							}
						}
					l477:
						depth--
						add(ruleNUMBER_NATURAL, position476)
					}
					depth--
					add(ruleNUMBER_INTEGER, position473)
				}
				{
					position481, tokenIndex481, depth481 := position, tokenIndex, depth
					{
						position483 := position
						depth++
						if buffer[position] != rune('.') {
							goto l481
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l481
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
						add(ruleNUMBER_FRACTION, position483)
					}
					goto l482
				l481:
					position, tokenIndex, depth = position481, tokenIndex481, depth481
				}
			l482:
				{
					position486, tokenIndex486, depth486 := position, tokenIndex, depth
					{
						position488 := position
						depth++
						{
							position489, tokenIndex489, depth489 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l490
							}
							position++
							goto l489
						l490:
							position, tokenIndex, depth = position489, tokenIndex489, depth489
							if buffer[position] != rune('E') {
								goto l486
							}
							position++
						}
					l489:
						{
							position491, tokenIndex491, depth491 := position, tokenIndex, depth
							{
								position493, tokenIndex493, depth493 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l494
								}
								position++
								goto l493
							l494:
								position, tokenIndex, depth = position493, tokenIndex493, depth493
								if buffer[position] != rune('-') {
									goto l491
								}
								position++
							}
						l493:
							goto l492
						l491:
							position, tokenIndex, depth = position491, tokenIndex491, depth491
						}
					l492:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l486
						}
						position++
					l495:
						{
							position496, tokenIndex496, depth496 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l496
							}
							position++
							goto l495
						l496:
							position, tokenIndex, depth = position496, tokenIndex496, depth496
						}
						depth--
						add(ruleNUMBER_EXP, position488)
					}
					goto l487
				l486:
					position, tokenIndex, depth = position486, tokenIndex486, depth486
				}
			l487:
				depth--
				add(ruleNUMBER, position472)
			}
			return true
		l471:
			position, tokenIndex, depth = position471, tokenIndex471, depth471
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
			position501, tokenIndex501, depth501 := position, tokenIndex, depth
			{
				position502 := position
				depth++
				if !_rules[rule_]() {
					goto l501
				}
				if buffer[position] != rune('(') {
					goto l501
				}
				position++
				if !_rules[rule_]() {
					goto l501
				}
				depth--
				add(rulePAREN_OPEN, position502)
			}
			return true
		l501:
			position, tokenIndex, depth = position501, tokenIndex501, depth501
			return false
		},
		/* 52 PAREN_CLOSE <- <(_ ')' _)> */
		func() bool {
			position503, tokenIndex503, depth503 := position, tokenIndex, depth
			{
				position504 := position
				depth++
				if !_rules[rule_]() {
					goto l503
				}
				if buffer[position] != rune(')') {
					goto l503
				}
				position++
				if !_rules[rule_]() {
					goto l503
				}
				depth--
				add(rulePAREN_CLOSE, position504)
			}
			return true
		l503:
			position, tokenIndex, depth = position503, tokenIndex503, depth503
			return false
		},
		/* 53 COMMA <- <(_ ',' _)> */
		func() bool {
			position505, tokenIndex505, depth505 := position, tokenIndex, depth
			{
				position506 := position
				depth++
				if !_rules[rule_]() {
					goto l505
				}
				if buffer[position] != rune(',') {
					goto l505
				}
				position++
				if !_rules[rule_]() {
					goto l505
				}
				depth--
				add(ruleCOMMA, position506)
			}
			return true
		l505:
			position, tokenIndex, depth = position505, tokenIndex505, depth505
			return false
		},
		/* 54 _ <- <SPACE*> */
		func() bool {
			{
				position508 := position
				depth++
			l509:
				{
					position510, tokenIndex510, depth510 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l510
					}
					goto l509
				l510:
					position, tokenIndex, depth = position510, tokenIndex510, depth510
				}
				depth--
				add(rule_, position508)
			}
			return true
		},
		/* 55 __ <- <SPACE+> */
		func() bool {
			position511, tokenIndex511, depth511 := position, tokenIndex, depth
			{
				position512 := position
				depth++
				if !_rules[ruleSPACE]() {
					goto l511
				}
			l513:
				{
					position514, tokenIndex514, depth514 := position, tokenIndex, depth
					if !_rules[ruleSPACE]() {
						goto l514
					}
					goto l513
				l514:
					position, tokenIndex, depth = position514, tokenIndex514, depth514
				}
				depth--
				add(rule__, position512)
			}
			return true
		l511:
			position, tokenIndex, depth = position511, tokenIndex511, depth511
			return false
		},
		/* 56 SPACE <- <((&('\t') '\t') | (&('\n') '\n') | (&(' ') ' '))> */
		func() bool {
			position515, tokenIndex515, depth515 := position, tokenIndex, depth
			{
				position516 := position
				depth++
				{
					switch buffer[position] {
					case '\t':
						if buffer[position] != rune('\t') {
							goto l515
						}
						position++
						break
					case '\n':
						if buffer[position] != rune('\n') {
							goto l515
						}
						position++
						break
					default:
						if buffer[position] != rune(' ') {
							goto l515
						}
						position++
						break
					}
				}

				depth--
				add(ruleSPACE, position516)
			}
			return true
		l515:
			position, tokenIndex, depth = position515, tokenIndex515, depth515
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
