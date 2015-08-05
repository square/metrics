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
	ruleoptionalMatchClause
	rulematchClause
	ruledescribeMetrics
	ruledescribeSingleStmt
	rulepropertyClause
	ruleoptionalPredicateClause
	ruleexpressionList
	ruleexpression_start
	ruleexpression_sum
	ruleexpression_product
	ruleadd_pipe
	ruleexpression_atom
	ruleoptionalGroupBy
	ruleexpression_function
	ruleexpression_metric
	rulegroupByClause
	rulecollapseByClause
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
	ruleDURATION
	rulePAREN_OPEN
	rulePAREN_CLOSE
	ruleCOMMA
	rule_
	ruleCOMMENT_TRAIL
	ruleCOMMENT_BLOCK
	ruleKEY
	ruleSPACE
	ruleAction0
	ruleAction1
	ruleAction2
	ruleAction3
	ruleAction4
	rulePegText
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
	ruleAction46
	ruleAction47
	ruleAction48
	ruleAction49

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
	"optionalMatchClause",
	"matchClause",
	"describeMetrics",
	"describeSingleStmt",
	"propertyClause",
	"optionalPredicateClause",
	"expressionList",
	"expression_start",
	"expression_sum",
	"expression_product",
	"add_pipe",
	"expression_atom",
	"optionalGroupBy",
	"expression_function",
	"expression_metric",
	"groupByClause",
	"collapseByClause",
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
	"DURATION",
	"PAREN_OPEN",
	"PAREN_CLOSE",
	"COMMA",
	"_",
	"COMMENT_TRAIL",
	"COMMENT_BLOCK",
	"KEY",
	"SPACE",
	"Action0",
	"Action1",
	"Action2",
	"Action3",
	"Action4",
	"PegText",
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
	"Action46",
	"Action47",
	"Action48",
	"Action49",

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
	rules  [120]func() bool
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
	for i, c := range []rune(buffer) {
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
			p.addNullMatchClause()
		case ruleAction3:
			p.addMatchClause()
		case ruleAction4:
			p.makeDescribeMetrics()
		case ruleAction5:
			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		case ruleAction6:
			p.makeDescribe()
		case ruleAction7:
			p.addEvaluationContext()
		case ruleAction8:
			p.addPropertyKey(buffer[begin:end])
		case ruleAction9:
			p.addPropertyValue(buffer[begin:end])
		case ruleAction10:
			p.insertPropertyKeyValue()
		case ruleAction11:
			p.checkPropertyClause()
		case ruleAction12:
			p.addNullPredicate()
		case ruleAction13:
			p.addExpressionList()
		case ruleAction14:
			p.appendExpression()
		case ruleAction15:
			p.appendExpression()
		case ruleAction16:
			p.addOperatorLiteral("+")
		case ruleAction17:
			p.addOperatorLiteral("-")
		case ruleAction18:
			p.addOperatorFunction()
		case ruleAction19:
			p.addOperatorLiteral("/")
		case ruleAction20:
			p.addOperatorLiteral("*")
		case ruleAction21:
			p.addOperatorFunction()
		case ruleAction22:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction23:
			p.addExpressionList()
		case ruleAction24:

			p.addExpressionList()
			p.addGroupBy()

		case ruleAction25:

			p.addPipeExpression()

		case ruleAction26:
			p.addDurationNode(text)
		case ruleAction27:
			p.addNumberNode(buffer[begin:end])
		case ruleAction28:
			p.addStringNode(unescapeLiteral(buffer[begin:end]))
		case ruleAction29:
			p.addGroupBy()
		case ruleAction30:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction31:

			p.addFunctionInvocation()

		case ruleAction32:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction33:
			p.addNullPredicate()
		case ruleAction34:

			p.addMetricExpression()

		case ruleAction35:

			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))

		case ruleAction36:

			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))

		case ruleAction37:

			p.appendCollapseBy(unescapeLiteral(text))

		case ruleAction38:
			p.appendCollapseBy(unescapeLiteral(text))
		case ruleAction39:
			p.addOrPredicate()
		case ruleAction40:
			p.addAndPredicate()
		case ruleAction41:
			p.addNotPredicate()
		case ruleAction42:

			p.addLiteralMatcher()

		case ruleAction43:

			p.addLiteralMatcher()
			p.addNotPredicate()

		case ruleAction44:

			p.addRegexMatcher()

		case ruleAction45:

			p.addListMatcher()

		case ruleAction46:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction47:
			p.addLiteralList()
		case ruleAction48:

			p.appendLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction49:
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
								add(ruleAction7, position)
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
									add(ruleAction8, position)
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
									add(ruleAction9, position)
								}
								{
									add(ruleAction10, position)
								}
								goto l21
							l22:
								position, tokenIndex, depth = position22, tokenIndex22, depth22
							}
							{
								add(ruleAction11, position)
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
									position71 := position
									depth++
									{
										position72, tokenIndex72, depth72 := position, tokenIndex, depth
										{
											position74 := position
											depth++
											if !_rules[rule_]() {
												goto l73
											}
											{
												position75, tokenIndex75, depth75 := position, tokenIndex, depth
												if buffer[position] != rune('m') {
													goto l76
												}
												position++
												goto l75
											l76:
												position, tokenIndex, depth = position75, tokenIndex75, depth75
												if buffer[position] != rune('M') {
													goto l73
												}
												position++
											}
										l75:
											{
												position77, tokenIndex77, depth77 := position, tokenIndex, depth
												if buffer[position] != rune('a') {
													goto l78
												}
												position++
												goto l77
											l78:
												position, tokenIndex, depth = position77, tokenIndex77, depth77
												if buffer[position] != rune('A') {
													goto l73
												}
												position++
											}
										l77:
											{
												position79, tokenIndex79, depth79 := position, tokenIndex, depth
												if buffer[position] != rune('t') {
													goto l80
												}
												position++
												goto l79
											l80:
												position, tokenIndex, depth = position79, tokenIndex79, depth79
												if buffer[position] != rune('T') {
													goto l73
												}
												position++
											}
										l79:
											{
												position81, tokenIndex81, depth81 := position, tokenIndex, depth
												if buffer[position] != rune('c') {
													goto l82
												}
												position++
												goto l81
											l82:
												position, tokenIndex, depth = position81, tokenIndex81, depth81
												if buffer[position] != rune('C') {
													goto l73
												}
												position++
											}
										l81:
											{
												position83, tokenIndex83, depth83 := position, tokenIndex, depth
												if buffer[position] != rune('h') {
													goto l84
												}
												position++
												goto l83
											l84:
												position, tokenIndex, depth = position83, tokenIndex83, depth83
												if buffer[position] != rune('H') {
													goto l73
												}
												position++
											}
										l83:
											if !_rules[ruleKEY]() {
												goto l73
											}
											if !_rules[ruleliteralString]() {
												goto l73
											}
											{
												add(ruleAction3, position)
											}
											depth--
											add(rulematchClause, position74)
										}
										goto l72
									l73:
										position, tokenIndex, depth = position72, tokenIndex72, depth72
										{
											add(ruleAction2, position)
										}
									}
								l72:
									depth--
									add(ruleoptionalMatchClause, position71)
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
								position89 := position
								depth++
								if !_rules[rule_]() {
									goto l88
								}
								{
									position90, tokenIndex90, depth90 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l91
									}
									position++
									goto l90
								l91:
									position, tokenIndex, depth = position90, tokenIndex90, depth90
									if buffer[position] != rune('M') {
										goto l88
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
										goto l88
									}
									position++
								}
							l92:
								{
									position94, tokenIndex94, depth94 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l95
									}
									position++
									goto l94
								l95:
									position, tokenIndex, depth = position94, tokenIndex94, depth94
									if buffer[position] != rune('T') {
										goto l88
									}
									position++
								}
							l94:
								{
									position96, tokenIndex96, depth96 := position, tokenIndex, depth
									if buffer[position] != rune('r') {
										goto l97
									}
									position++
									goto l96
								l97:
									position, tokenIndex, depth = position96, tokenIndex96, depth96
									if buffer[position] != rune('R') {
										goto l88
									}
									position++
								}
							l96:
								{
									position98, tokenIndex98, depth98 := position, tokenIndex, depth
									if buffer[position] != rune('i') {
										goto l99
									}
									position++
									goto l98
								l99:
									position, tokenIndex, depth = position98, tokenIndex98, depth98
									if buffer[position] != rune('I') {
										goto l88
									}
									position++
								}
							l98:
								{
									position100, tokenIndex100, depth100 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l101
									}
									position++
									goto l100
								l101:
									position, tokenIndex, depth = position100, tokenIndex100, depth100
									if buffer[position] != rune('C') {
										goto l88
									}
									position++
								}
							l100:
								{
									position102, tokenIndex102, depth102 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l103
									}
									position++
									goto l102
								l103:
									position, tokenIndex, depth = position102, tokenIndex102, depth102
									if buffer[position] != rune('S') {
										goto l88
									}
									position++
								}
							l102:
								if !_rules[ruleKEY]() {
									goto l88
								}
								if !_rules[rule_]() {
									goto l88
								}
								{
									position104, tokenIndex104, depth104 := position, tokenIndex, depth
									if buffer[position] != rune('w') {
										goto l105
									}
									position++
									goto l104
								l105:
									position, tokenIndex, depth = position104, tokenIndex104, depth104
									if buffer[position] != rune('W') {
										goto l88
									}
									position++
								}
							l104:
								{
									position106, tokenIndex106, depth106 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l107
									}
									position++
									goto l106
								l107:
									position, tokenIndex, depth = position106, tokenIndex106, depth106
									if buffer[position] != rune('H') {
										goto l88
									}
									position++
								}
							l106:
								{
									position108, tokenIndex108, depth108 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l109
									}
									position++
									goto l108
								l109:
									position, tokenIndex, depth = position108, tokenIndex108, depth108
									if buffer[position] != rune('E') {
										goto l88
									}
									position++
								}
							l108:
								{
									position110, tokenIndex110, depth110 := position, tokenIndex, depth
									if buffer[position] != rune('r') {
										goto l111
									}
									position++
									goto l110
								l111:
									position, tokenIndex, depth = position110, tokenIndex110, depth110
									if buffer[position] != rune('R') {
										goto l88
									}
									position++
								}
							l110:
								{
									position112, tokenIndex112, depth112 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l113
									}
									position++
									goto l112
								l113:
									position, tokenIndex, depth = position112, tokenIndex112, depth112
									if buffer[position] != rune('E') {
										goto l88
									}
									position++
								}
							l112:
								if !_rules[ruleKEY]() {
									goto l88
								}
								if !_rules[ruletagName]() {
									goto l88
								}
								if !_rules[rule_]() {
									goto l88
								}
								if buffer[position] != rune('=') {
									goto l88
								}
								position++
								if !_rules[ruleliteralString]() {
									goto l88
								}
								{
									add(ruleAction4, position)
								}
								depth--
								add(ruledescribeMetrics, position89)
							}
							goto l62
						l88:
							position, tokenIndex, depth = position62, tokenIndex62, depth62
							{
								position115 := position
								depth++
								if !_rules[rule_]() {
									goto l0
								}
								{
									position116 := position
									depth++
									{
										position117 := position
										depth++
										if !_rules[ruleIDENTIFIER]() {
											goto l0
										}
										depth--
										add(ruleMETRIC_NAME, position117)
									}
									depth--
									add(rulePegText, position116)
								}
								{
									add(ruleAction5, position)
								}
								if !_rules[ruleoptionalPredicateClause]() {
									goto l0
								}
								{
									add(ruleAction6, position)
								}
								depth--
								add(ruledescribeSingleStmt, position115)
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
					position120, tokenIndex120, depth120 := position, tokenIndex, depth
					if !matchDot() {
						goto l120
					}
					goto l0
				l120:
					position, tokenIndex, depth = position120, tokenIndex120, depth120
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
		/* 3 describeAllStmt <- <(_ (('a' / 'A') ('l' / 'L') ('l' / 'L')) KEY optionalMatchClause Action1)> */
		nil,
		/* 4 optionalMatchClause <- <(matchClause / Action2)> */
		nil,
		/* 5 matchClause <- <(_ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H')) KEY literalString Action3)> */
		nil,
		/* 6 describeMetrics <- <(_ (('m' / 'M') ('e' / 'E') ('t' / 'T') ('r' / 'R') ('i' / 'I') ('c' / 'C') ('s' / 'S')) KEY _ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY tagName _ '=' literalString Action4)> */
		nil,
		/* 7 describeSingleStmt <- <(_ <METRIC_NAME> Action5 optionalPredicateClause Action6)> */
		nil,
		/* 8 propertyClause <- <(Action7 (_ PROPERTY_KEY Action8 _ PROPERTY_VALUE Action9 Action10)* Action11)> */
		nil,
		/* 9 optionalPredicateClause <- <(predicateClause / Action12)> */
		func() bool {
			{
				position130 := position
				depth++
				{
					position131, tokenIndex131, depth131 := position, tokenIndex, depth
					{
						position133 := position
						depth++
						if !_rules[rule_]() {
							goto l132
						}
						{
							position134, tokenIndex134, depth134 := position, tokenIndex, depth
							if buffer[position] != rune('w') {
								goto l135
							}
							position++
							goto l134
						l135:
							position, tokenIndex, depth = position134, tokenIndex134, depth134
							if buffer[position] != rune('W') {
								goto l132
							}
							position++
						}
					l134:
						{
							position136, tokenIndex136, depth136 := position, tokenIndex, depth
							if buffer[position] != rune('h') {
								goto l137
							}
							position++
							goto l136
						l137:
							position, tokenIndex, depth = position136, tokenIndex136, depth136
							if buffer[position] != rune('H') {
								goto l132
							}
							position++
						}
					l136:
						{
							position138, tokenIndex138, depth138 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l139
							}
							position++
							goto l138
						l139:
							position, tokenIndex, depth = position138, tokenIndex138, depth138
							if buffer[position] != rune('E') {
								goto l132
							}
							position++
						}
					l138:
						{
							position140, tokenIndex140, depth140 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l141
							}
							position++
							goto l140
						l141:
							position, tokenIndex, depth = position140, tokenIndex140, depth140
							if buffer[position] != rune('R') {
								goto l132
							}
							position++
						}
					l140:
						{
							position142, tokenIndex142, depth142 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l143
							}
							position++
							goto l142
						l143:
							position, tokenIndex, depth = position142, tokenIndex142, depth142
							if buffer[position] != rune('E') {
								goto l132
							}
							position++
						}
					l142:
						if !_rules[ruleKEY]() {
							goto l132
						}
						if !_rules[rule_]() {
							goto l132
						}
						if !_rules[rulepredicate_1]() {
							goto l132
						}
						depth--
						add(rulepredicateClause, position133)
					}
					goto l131
				l132:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
					{
						add(ruleAction12, position)
					}
				}
			l131:
				depth--
				add(ruleoptionalPredicateClause, position130)
			}
			return true
		},
		/* 10 expressionList <- <(Action13 expression_start Action14 (_ COMMA expression_start Action15)*)> */
		func() bool {
			position145, tokenIndex145, depth145 := position, tokenIndex, depth
			{
				position146 := position
				depth++
				{
					add(ruleAction13, position)
				}
				if !_rules[ruleexpression_start]() {
					goto l145
				}
				{
					add(ruleAction14, position)
				}
			l149:
				{
					position150, tokenIndex150, depth150 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l150
					}
					if !_rules[ruleCOMMA]() {
						goto l150
					}
					if !_rules[ruleexpression_start]() {
						goto l150
					}
					{
						add(ruleAction15, position)
					}
					goto l149
				l150:
					position, tokenIndex, depth = position150, tokenIndex150, depth150
				}
				depth--
				add(ruleexpressionList, position146)
			}
			return true
		l145:
			position, tokenIndex, depth = position145, tokenIndex145, depth145
			return false
		},
		/* 11 expression_start <- <(expression_sum add_pipe)> */
		func() bool {
			position152, tokenIndex152, depth152 := position, tokenIndex, depth
			{
				position153 := position
				depth++
				{
					position154 := position
					depth++
					if !_rules[ruleexpression_product]() {
						goto l152
					}
				l155:
					{
						position156, tokenIndex156, depth156 := position, tokenIndex, depth
						if !_rules[ruleadd_pipe]() {
							goto l156
						}
						{
							position157, tokenIndex157, depth157 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l158
							}
							{
								position159 := position
								depth++
								if buffer[position] != rune('+') {
									goto l158
								}
								position++
								depth--
								add(ruleOP_ADD, position159)
							}
							{
								add(ruleAction16, position)
							}
							goto l157
						l158:
							position, tokenIndex, depth = position157, tokenIndex157, depth157
							if !_rules[rule_]() {
								goto l156
							}
							{
								position161 := position
								depth++
								if buffer[position] != rune('-') {
									goto l156
								}
								position++
								depth--
								add(ruleOP_SUB, position161)
							}
							{
								add(ruleAction17, position)
							}
						}
					l157:
						if !_rules[ruleexpression_product]() {
							goto l156
						}
						{
							add(ruleAction18, position)
						}
						goto l155
					l156:
						position, tokenIndex, depth = position156, tokenIndex156, depth156
					}
					depth--
					add(ruleexpression_sum, position154)
				}
				if !_rules[ruleadd_pipe]() {
					goto l152
				}
				depth--
				add(ruleexpression_start, position153)
			}
			return true
		l152:
			position, tokenIndex, depth = position152, tokenIndex152, depth152
			return false
		},
		/* 12 expression_sum <- <(expression_product (add_pipe ((_ OP_ADD Action16) / (_ OP_SUB Action17)) expression_product Action18)*)> */
		nil,
		/* 13 expression_product <- <(expression_atom (add_pipe ((_ OP_DIV Action19) / (_ OP_MULT Action20)) expression_atom Action21)*)> */
		func() bool {
			position165, tokenIndex165, depth165 := position, tokenIndex, depth
			{
				position166 := position
				depth++
				if !_rules[ruleexpression_atom]() {
					goto l165
				}
			l167:
				{
					position168, tokenIndex168, depth168 := position, tokenIndex, depth
					if !_rules[ruleadd_pipe]() {
						goto l168
					}
					{
						position169, tokenIndex169, depth169 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l170
						}
						{
							position171 := position
							depth++
							if buffer[position] != rune('/') {
								goto l170
							}
							position++
							depth--
							add(ruleOP_DIV, position171)
						}
						{
							add(ruleAction19, position)
						}
						goto l169
					l170:
						position, tokenIndex, depth = position169, tokenIndex169, depth169
						if !_rules[rule_]() {
							goto l168
						}
						{
							position173 := position
							depth++
							if buffer[position] != rune('*') {
								goto l168
							}
							position++
							depth--
							add(ruleOP_MULT, position173)
						}
						{
							add(ruleAction20, position)
						}
					}
				l169:
					if !_rules[ruleexpression_atom]() {
						goto l168
					}
					{
						add(ruleAction21, position)
					}
					goto l167
				l168:
					position, tokenIndex, depth = position168, tokenIndex168, depth168
				}
				depth--
				add(ruleexpression_product, position166)
			}
			return true
		l165:
			position, tokenIndex, depth = position165, tokenIndex165, depth165
			return false
		},
		/* 14 add_pipe <- <(_ OP_PIPE _ <IDENTIFIER> Action22 ((_ PAREN_OPEN (expressionList / Action23) optionalGroupBy _ PAREN_CLOSE) / Action24) Action25)*> */
		func() bool {
			{
				position177 := position
				depth++
			l178:
				{
					position179, tokenIndex179, depth179 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l179
					}
					{
						position180 := position
						depth++
						if buffer[position] != rune('|') {
							goto l179
						}
						position++
						depth--
						add(ruleOP_PIPE, position180)
					}
					if !_rules[rule_]() {
						goto l179
					}
					{
						position181 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l179
						}
						depth--
						add(rulePegText, position181)
					}
					{
						add(ruleAction22, position)
					}
					{
						position183, tokenIndex183, depth183 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l184
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l184
						}
						{
							position185, tokenIndex185, depth185 := position, tokenIndex, depth
							if !_rules[ruleexpressionList]() {
								goto l186
							}
							goto l185
						l186:
							position, tokenIndex, depth = position185, tokenIndex185, depth185
							{
								add(ruleAction23, position)
							}
						}
					l185:
						if !_rules[ruleoptionalGroupBy]() {
							goto l184
						}
						if !_rules[rule_]() {
							goto l184
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l184
						}
						goto l183
					l184:
						position, tokenIndex, depth = position183, tokenIndex183, depth183
						{
							add(ruleAction24, position)
						}
					}
				l183:
					{
						add(ruleAction25, position)
					}
					goto l178
				l179:
					position, tokenIndex, depth = position179, tokenIndex179, depth179
				}
				depth--
				add(ruleadd_pipe, position177)
			}
			return true
		},
		/* 15 expression_atom <- <(expression_function / expression_metric / (_ PAREN_OPEN expression_start _ PAREN_CLOSE) / (_ <DURATION> Action26) / (_ <NUMBER> Action27) / (_ STRING Action28))> */
		func() bool {
			position190, tokenIndex190, depth190 := position, tokenIndex, depth
			{
				position191 := position
				depth++
				{
					position192, tokenIndex192, depth192 := position, tokenIndex, depth
					{
						position194 := position
						depth++
						if !_rules[rule_]() {
							goto l193
						}
						{
							position195 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l193
							}
							depth--
							add(rulePegText, position195)
						}
						{
							add(ruleAction30, position)
						}
						if !_rules[rule_]() {
							goto l193
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l193
						}
						if !_rules[ruleexpressionList]() {
							goto l193
						}
						if !_rules[ruleoptionalGroupBy]() {
							goto l193
						}
						if !_rules[rule_]() {
							goto l193
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l193
						}
						{
							add(ruleAction31, position)
						}
						depth--
						add(ruleexpression_function, position194)
					}
					goto l192
				l193:
					position, tokenIndex, depth = position192, tokenIndex192, depth192
					{
						position199 := position
						depth++
						if !_rules[rule_]() {
							goto l198
						}
						{
							position200 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l198
							}
							depth--
							add(rulePegText, position200)
						}
						{
							add(ruleAction32, position)
						}
						{
							position202, tokenIndex202, depth202 := position, tokenIndex, depth
							{
								position204, tokenIndex204, depth204 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l205
								}
								if buffer[position] != rune('[') {
									goto l205
								}
								position++
								if !_rules[rulepredicate_1]() {
									goto l205
								}
								if !_rules[rule_]() {
									goto l205
								}
								if buffer[position] != rune(']') {
									goto l205
								}
								position++
								goto l204
							l205:
								position, tokenIndex, depth = position204, tokenIndex204, depth204
								{
									add(ruleAction33, position)
								}
							}
						l204:
							goto l203

							position, tokenIndex, depth = position202, tokenIndex202, depth202
						}
					l203:
						{
							add(ruleAction34, position)
						}
						depth--
						add(ruleexpression_metric, position199)
					}
					goto l192
				l198:
					position, tokenIndex, depth = position192, tokenIndex192, depth192
					if !_rules[rule_]() {
						goto l208
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l208
					}
					if !_rules[ruleexpression_start]() {
						goto l208
					}
					if !_rules[rule_]() {
						goto l208
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l208
					}
					goto l192
				l208:
					position, tokenIndex, depth = position192, tokenIndex192, depth192
					if !_rules[rule_]() {
						goto l209
					}
					{
						position210 := position
						depth++
						{
							position211 := position
							depth++
							if !_rules[ruleNUMBER]() {
								goto l209
							}
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l209
							}
							position++
						l212:
							{
								position213, tokenIndex213, depth213 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l213
								}
								position++
								goto l212
							l213:
								position, tokenIndex, depth = position213, tokenIndex213, depth213
							}
							if !_rules[ruleKEY]() {
								goto l209
							}
							depth--
							add(ruleDURATION, position211)
						}
						depth--
						add(rulePegText, position210)
					}
					{
						add(ruleAction26, position)
					}
					goto l192
				l209:
					position, tokenIndex, depth = position192, tokenIndex192, depth192
					if !_rules[rule_]() {
						goto l215
					}
					{
						position216 := position
						depth++
						if !_rules[ruleNUMBER]() {
							goto l215
						}
						depth--
						add(rulePegText, position216)
					}
					{
						add(ruleAction27, position)
					}
					goto l192
				l215:
					position, tokenIndex, depth = position192, tokenIndex192, depth192
					if !_rules[rule_]() {
						goto l190
					}
					if !_rules[ruleSTRING]() {
						goto l190
					}
					{
						add(ruleAction28, position)
					}
				}
			l192:
				depth--
				add(ruleexpression_atom, position191)
			}
			return true
		l190:
			position, tokenIndex, depth = position190, tokenIndex190, depth190
			return false
		},
		/* 16 optionalGroupBy <- <(Action29 (groupByClause / collapseByClause)?)> */
		func() bool {
			{
				position220 := position
				depth++
				{
					add(ruleAction29, position)
				}
				{
					position222, tokenIndex222, depth222 := position, tokenIndex, depth
					{
						position224, tokenIndex224, depth224 := position, tokenIndex, depth
						{
							position226 := position
							depth++
							if !_rules[rule_]() {
								goto l225
							}
							{
								position227, tokenIndex227, depth227 := position, tokenIndex, depth
								if buffer[position] != rune('g') {
									goto l228
								}
								position++
								goto l227
							l228:
								position, tokenIndex, depth = position227, tokenIndex227, depth227
								if buffer[position] != rune('G') {
									goto l225
								}
								position++
							}
						l227:
							{
								position229, tokenIndex229, depth229 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l230
								}
								position++
								goto l229
							l230:
								position, tokenIndex, depth = position229, tokenIndex229, depth229
								if buffer[position] != rune('R') {
									goto l225
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
									goto l225
								}
								position++
							}
						l231:
							{
								position233, tokenIndex233, depth233 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l234
								}
								position++
								goto l233
							l234:
								position, tokenIndex, depth = position233, tokenIndex233, depth233
								if buffer[position] != rune('U') {
									goto l225
								}
								position++
							}
						l233:
							{
								position235, tokenIndex235, depth235 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l236
								}
								position++
								goto l235
							l236:
								position, tokenIndex, depth = position235, tokenIndex235, depth235
								if buffer[position] != rune('P') {
									goto l225
								}
								position++
							}
						l235:
							if !_rules[ruleKEY]() {
								goto l225
							}
							if !_rules[rule_]() {
								goto l225
							}
							{
								position237, tokenIndex237, depth237 := position, tokenIndex, depth
								if buffer[position] != rune('b') {
									goto l238
								}
								position++
								goto l237
							l238:
								position, tokenIndex, depth = position237, tokenIndex237, depth237
								if buffer[position] != rune('B') {
									goto l225
								}
								position++
							}
						l237:
							{
								position239, tokenIndex239, depth239 := position, tokenIndex, depth
								if buffer[position] != rune('y') {
									goto l240
								}
								position++
								goto l239
							l240:
								position, tokenIndex, depth = position239, tokenIndex239, depth239
								if buffer[position] != rune('Y') {
									goto l225
								}
								position++
							}
						l239:
							if !_rules[ruleKEY]() {
								goto l225
							}
							if !_rules[rule_]() {
								goto l225
							}
							{
								position241 := position
								depth++
								if !_rules[ruleCOLUMN_NAME]() {
									goto l225
								}
								depth--
								add(rulePegText, position241)
							}
							{
								add(ruleAction35, position)
							}
						l243:
							{
								position244, tokenIndex244, depth244 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l244
								}
								if !_rules[ruleCOMMA]() {
									goto l244
								}
								if !_rules[rule_]() {
									goto l244
								}
								{
									position245 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l244
									}
									depth--
									add(rulePegText, position245)
								}
								{
									add(ruleAction36, position)
								}
								goto l243
							l244:
								position, tokenIndex, depth = position244, tokenIndex244, depth244
							}
							depth--
							add(rulegroupByClause, position226)
						}
						goto l224
					l225:
						position, tokenIndex, depth = position224, tokenIndex224, depth224
						{
							position247 := position
							depth++
							if !_rules[rule_]() {
								goto l222
							}
							{
								position248, tokenIndex248, depth248 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l249
								}
								position++
								goto l248
							l249:
								position, tokenIndex, depth = position248, tokenIndex248, depth248
								if buffer[position] != rune('C') {
									goto l222
								}
								position++
							}
						l248:
							{
								position250, tokenIndex250, depth250 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l251
								}
								position++
								goto l250
							l251:
								position, tokenIndex, depth = position250, tokenIndex250, depth250
								if buffer[position] != rune('O') {
									goto l222
								}
								position++
							}
						l250:
							{
								position252, tokenIndex252, depth252 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l253
								}
								position++
								goto l252
							l253:
								position, tokenIndex, depth = position252, tokenIndex252, depth252
								if buffer[position] != rune('L') {
									goto l222
								}
								position++
							}
						l252:
							{
								position254, tokenIndex254, depth254 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l255
								}
								position++
								goto l254
							l255:
								position, tokenIndex, depth = position254, tokenIndex254, depth254
								if buffer[position] != rune('L') {
									goto l222
								}
								position++
							}
						l254:
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
									goto l222
								}
								position++
							}
						l256:
							{
								position258, tokenIndex258, depth258 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l259
								}
								position++
								goto l258
							l259:
								position, tokenIndex, depth = position258, tokenIndex258, depth258
								if buffer[position] != rune('P') {
									goto l222
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
									goto l222
								}
								position++
							}
						l260:
							{
								position262, tokenIndex262, depth262 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l263
								}
								position++
								goto l262
							l263:
								position, tokenIndex, depth = position262, tokenIndex262, depth262
								if buffer[position] != rune('E') {
									goto l222
								}
								position++
							}
						l262:
							if !_rules[ruleKEY]() {
								goto l222
							}
							if !_rules[rule_]() {
								goto l222
							}
							{
								position264, tokenIndex264, depth264 := position, tokenIndex, depth
								if buffer[position] != rune('b') {
									goto l265
								}
								position++
								goto l264
							l265:
								position, tokenIndex, depth = position264, tokenIndex264, depth264
								if buffer[position] != rune('B') {
									goto l222
								}
								position++
							}
						l264:
							{
								position266, tokenIndex266, depth266 := position, tokenIndex, depth
								if buffer[position] != rune('y') {
									goto l267
								}
								position++
								goto l266
							l267:
								position, tokenIndex, depth = position266, tokenIndex266, depth266
								if buffer[position] != rune('Y') {
									goto l222
								}
								position++
							}
						l266:
							if !_rules[ruleKEY]() {
								goto l222
							}
							if !_rules[rule_]() {
								goto l222
							}
							{
								position268 := position
								depth++
								if !_rules[ruleCOLUMN_NAME]() {
									goto l222
								}
								depth--
								add(rulePegText, position268)
							}
							{
								add(ruleAction37, position)
							}
						l270:
							{
								position271, tokenIndex271, depth271 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l271
								}
								if !_rules[ruleCOMMA]() {
									goto l271
								}
								if !_rules[rule_]() {
									goto l271
								}
								{
									position272 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l271
									}
									depth--
									add(rulePegText, position272)
								}
								{
									add(ruleAction38, position)
								}
								goto l270
							l271:
								position, tokenIndex, depth = position271, tokenIndex271, depth271
							}
							depth--
							add(rulecollapseByClause, position247)
						}
					}
				l224:
					goto l223
				l222:
					position, tokenIndex, depth = position222, tokenIndex222, depth222
				}
			l223:
				depth--
				add(ruleoptionalGroupBy, position220)
			}
			return true
		},
		/* 17 expression_function <- <(_ <IDENTIFIER> Action30 _ PAREN_OPEN expressionList optionalGroupBy _ PAREN_CLOSE Action31)> */
		nil,
		/* 18 expression_metric <- <(_ <IDENTIFIER> Action32 ((_ '[' predicate_1 _ ']') / Action33)? Action34)> */
		nil,
		/* 19 groupByClause <- <(_ (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P')) KEY _ (('b' / 'B') ('y' / 'Y')) KEY _ <COLUMN_NAME> Action35 (_ COMMA _ <COLUMN_NAME> Action36)*)> */
		nil,
		/* 20 collapseByClause <- <(_ (('c' / 'C') ('o' / 'O') ('l' / 'L') ('l' / 'L') ('a' / 'A') ('p' / 'P') ('s' / 'S') ('e' / 'E')) KEY _ (('b' / 'B') ('y' / 'Y')) KEY _ <COLUMN_NAME> Action37 (_ COMMA _ <COLUMN_NAME> Action38)*)> */
		nil,
		/* 21 predicateClause <- <(_ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY _ predicate_1)> */
		nil,
		/* 22 predicate_1 <- <((predicate_2 _ OP_OR predicate_1 Action39) / predicate_2)> */
		func() bool {
			position279, tokenIndex279, depth279 := position, tokenIndex, depth
			{
				position280 := position
				depth++
				{
					position281, tokenIndex281, depth281 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l282
					}
					if !_rules[rule_]() {
						goto l282
					}
					{
						position283 := position
						depth++
						{
							position284, tokenIndex284, depth284 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l285
							}
							position++
							goto l284
						l285:
							position, tokenIndex, depth = position284, tokenIndex284, depth284
							if buffer[position] != rune('O') {
								goto l282
							}
							position++
						}
					l284:
						{
							position286, tokenIndex286, depth286 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l287
							}
							position++
							goto l286
						l287:
							position, tokenIndex, depth = position286, tokenIndex286, depth286
							if buffer[position] != rune('R') {
								goto l282
							}
							position++
						}
					l286:
						if !_rules[ruleKEY]() {
							goto l282
						}
						depth--
						add(ruleOP_OR, position283)
					}
					if !_rules[rulepredicate_1]() {
						goto l282
					}
					{
						add(ruleAction39, position)
					}
					goto l281
				l282:
					position, tokenIndex, depth = position281, tokenIndex281, depth281
					if !_rules[rulepredicate_2]() {
						goto l279
					}
				}
			l281:
				depth--
				add(rulepredicate_1, position280)
			}
			return true
		l279:
			position, tokenIndex, depth = position279, tokenIndex279, depth279
			return false
		},
		/* 23 predicate_2 <- <((predicate_3 _ OP_AND predicate_2 Action40) / predicate_3)> */
		func() bool {
			position289, tokenIndex289, depth289 := position, tokenIndex, depth
			{
				position290 := position
				depth++
				{
					position291, tokenIndex291, depth291 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l292
					}
					if !_rules[rule_]() {
						goto l292
					}
					{
						position293 := position
						depth++
						{
							position294, tokenIndex294, depth294 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l295
							}
							position++
							goto l294
						l295:
							position, tokenIndex, depth = position294, tokenIndex294, depth294
							if buffer[position] != rune('A') {
								goto l292
							}
							position++
						}
					l294:
						{
							position296, tokenIndex296, depth296 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l297
							}
							position++
							goto l296
						l297:
							position, tokenIndex, depth = position296, tokenIndex296, depth296
							if buffer[position] != rune('N') {
								goto l292
							}
							position++
						}
					l296:
						{
							position298, tokenIndex298, depth298 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l299
							}
							position++
							goto l298
						l299:
							position, tokenIndex, depth = position298, tokenIndex298, depth298
							if buffer[position] != rune('D') {
								goto l292
							}
							position++
						}
					l298:
						if !_rules[ruleKEY]() {
							goto l292
						}
						depth--
						add(ruleOP_AND, position293)
					}
					if !_rules[rulepredicate_2]() {
						goto l292
					}
					{
						add(ruleAction40, position)
					}
					goto l291
				l292:
					position, tokenIndex, depth = position291, tokenIndex291, depth291
					if !_rules[rulepredicate_3]() {
						goto l289
					}
				}
			l291:
				depth--
				add(rulepredicate_2, position290)
			}
			return true
		l289:
			position, tokenIndex, depth = position289, tokenIndex289, depth289
			return false
		},
		/* 24 predicate_3 <- <((_ OP_NOT predicate_3 Action41) / (_ PAREN_OPEN predicate_1 _ PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position301, tokenIndex301, depth301 := position, tokenIndex, depth
			{
				position302 := position
				depth++
				{
					position303, tokenIndex303, depth303 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l304
					}
					{
						position305 := position
						depth++
						{
							position306, tokenIndex306, depth306 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l307
							}
							position++
							goto l306
						l307:
							position, tokenIndex, depth = position306, tokenIndex306, depth306
							if buffer[position] != rune('N') {
								goto l304
							}
							position++
						}
					l306:
						{
							position308, tokenIndex308, depth308 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l309
							}
							position++
							goto l308
						l309:
							position, tokenIndex, depth = position308, tokenIndex308, depth308
							if buffer[position] != rune('O') {
								goto l304
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
								goto l304
							}
							position++
						}
					l310:
						if !_rules[ruleKEY]() {
							goto l304
						}
						depth--
						add(ruleOP_NOT, position305)
					}
					if !_rules[rulepredicate_3]() {
						goto l304
					}
					{
						add(ruleAction41, position)
					}
					goto l303
				l304:
					position, tokenIndex, depth = position303, tokenIndex303, depth303
					if !_rules[rule_]() {
						goto l313
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l313
					}
					if !_rules[rulepredicate_1]() {
						goto l313
					}
					if !_rules[rule_]() {
						goto l313
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l313
					}
					goto l303
				l313:
					position, tokenIndex, depth = position303, tokenIndex303, depth303
					{
						position314 := position
						depth++
						{
							position315, tokenIndex315, depth315 := position, tokenIndex, depth
							if !_rules[ruletagName]() {
								goto l316
							}
							if !_rules[rule_]() {
								goto l316
							}
							if buffer[position] != rune('=') {
								goto l316
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l316
							}
							{
								add(ruleAction42, position)
							}
							goto l315
						l316:
							position, tokenIndex, depth = position315, tokenIndex315, depth315
							if !_rules[ruletagName]() {
								goto l318
							}
							if !_rules[rule_]() {
								goto l318
							}
							if buffer[position] != rune('!') {
								goto l318
							}
							position++
							if buffer[position] != rune('=') {
								goto l318
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l318
							}
							{
								add(ruleAction43, position)
							}
							goto l315
						l318:
							position, tokenIndex, depth = position315, tokenIndex315, depth315
							if !_rules[ruletagName]() {
								goto l320
							}
							if !_rules[rule_]() {
								goto l320
							}
							{
								position321, tokenIndex321, depth321 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l322
								}
								position++
								goto l321
							l322:
								position, tokenIndex, depth = position321, tokenIndex321, depth321
								if buffer[position] != rune('M') {
									goto l320
								}
								position++
							}
						l321:
							{
								position323, tokenIndex323, depth323 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l324
								}
								position++
								goto l323
							l324:
								position, tokenIndex, depth = position323, tokenIndex323, depth323
								if buffer[position] != rune('A') {
									goto l320
								}
								position++
							}
						l323:
							{
								position325, tokenIndex325, depth325 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l326
								}
								position++
								goto l325
							l326:
								position, tokenIndex, depth = position325, tokenIndex325, depth325
								if buffer[position] != rune('T') {
									goto l320
								}
								position++
							}
						l325:
							{
								position327, tokenIndex327, depth327 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l328
								}
								position++
								goto l327
							l328:
								position, tokenIndex, depth = position327, tokenIndex327, depth327
								if buffer[position] != rune('C') {
									goto l320
								}
								position++
							}
						l327:
							{
								position329, tokenIndex329, depth329 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l330
								}
								position++
								goto l329
							l330:
								position, tokenIndex, depth = position329, tokenIndex329, depth329
								if buffer[position] != rune('H') {
									goto l320
								}
								position++
							}
						l329:
							if !_rules[ruleKEY]() {
								goto l320
							}
							if !_rules[ruleliteralString]() {
								goto l320
							}
							{
								add(ruleAction44, position)
							}
							goto l315
						l320:
							position, tokenIndex, depth = position315, tokenIndex315, depth315
							if !_rules[ruletagName]() {
								goto l301
							}
							if !_rules[rule_]() {
								goto l301
							}
							{
								position332, tokenIndex332, depth332 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l333
								}
								position++
								goto l332
							l333:
								position, tokenIndex, depth = position332, tokenIndex332, depth332
								if buffer[position] != rune('I') {
									goto l301
								}
								position++
							}
						l332:
							{
								position334, tokenIndex334, depth334 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l335
								}
								position++
								goto l334
							l335:
								position, tokenIndex, depth = position334, tokenIndex334, depth334
								if buffer[position] != rune('N') {
									goto l301
								}
								position++
							}
						l334:
							if !_rules[ruleKEY]() {
								goto l301
							}
							{
								position336 := position
								depth++
								{
									add(ruleAction47, position)
								}
								if !_rules[rule_]() {
									goto l301
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l301
								}
								if !_rules[ruleliteralListString]() {
									goto l301
								}
							l338:
								{
									position339, tokenIndex339, depth339 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l339
									}
									if !_rules[ruleCOMMA]() {
										goto l339
									}
									if !_rules[ruleliteralListString]() {
										goto l339
									}
									goto l338
								l339:
									position, tokenIndex, depth = position339, tokenIndex339, depth339
								}
								if !_rules[rule_]() {
									goto l301
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l301
								}
								depth--
								add(ruleliteralList, position336)
							}
							{
								add(ruleAction45, position)
							}
						}
					l315:
						depth--
						add(ruletagMatcher, position314)
					}
				}
			l303:
				depth--
				add(rulepredicate_3, position302)
			}
			return true
		l301:
			position, tokenIndex, depth = position301, tokenIndex301, depth301
			return false
		},
		/* 25 tagMatcher <- <((tagName _ '=' literalString Action42) / (tagName _ ('!' '=') literalString Action43) / (tagName _ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H')) KEY literalString Action44) / (tagName _ (('i' / 'I') ('n' / 'N')) KEY literalList Action45))> */
		nil,
		/* 26 literalString <- <(_ STRING Action46)> */
		func() bool {
			position342, tokenIndex342, depth342 := position, tokenIndex, depth
			{
				position343 := position
				depth++
				if !_rules[rule_]() {
					goto l342
				}
				if !_rules[ruleSTRING]() {
					goto l342
				}
				{
					add(ruleAction46, position)
				}
				depth--
				add(ruleliteralString, position343)
			}
			return true
		l342:
			position, tokenIndex, depth = position342, tokenIndex342, depth342
			return false
		},
		/* 27 literalList <- <(Action47 _ PAREN_OPEN literalListString (_ COMMA literalListString)* _ PAREN_CLOSE)> */
		nil,
		/* 28 literalListString <- <(_ STRING Action48)> */
		func() bool {
			position346, tokenIndex346, depth346 := position, tokenIndex, depth
			{
				position347 := position
				depth++
				if !_rules[rule_]() {
					goto l346
				}
				if !_rules[ruleSTRING]() {
					goto l346
				}
				{
					add(ruleAction48, position)
				}
				depth--
				add(ruleliteralListString, position347)
			}
			return true
		l346:
			position, tokenIndex, depth = position346, tokenIndex346, depth346
			return false
		},
		/* 29 tagName <- <(_ <TAG_NAME> Action49)> */
		func() bool {
			position349, tokenIndex349, depth349 := position, tokenIndex, depth
			{
				position350 := position
				depth++
				if !_rules[rule_]() {
					goto l349
				}
				{
					position351 := position
					depth++
					{
						position352 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l349
						}
						depth--
						add(ruleTAG_NAME, position352)
					}
					depth--
					add(rulePegText, position351)
				}
				{
					add(ruleAction49, position)
				}
				depth--
				add(ruletagName, position350)
			}
			return true
		l349:
			position, tokenIndex, depth = position349, tokenIndex349, depth349
			return false
		},
		/* 30 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position354, tokenIndex354, depth354 := position, tokenIndex, depth
			{
				position355 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l354
				}
				depth--
				add(ruleCOLUMN_NAME, position355)
			}
			return true
		l354:
			position, tokenIndex, depth = position354, tokenIndex354, depth354
			return false
		},
		/* 31 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 32 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 33 IDENTIFIER <- <(('`' CHAR* '`') / (_ !(KEYWORD KEY) ID_SEGMENT ('.' ID_SEGMENT)*))> */
		func() bool {
			position358, tokenIndex358, depth358 := position, tokenIndex, depth
			{
				position359 := position
				depth++
				{
					position360, tokenIndex360, depth360 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l361
					}
					position++
				l362:
					{
						position363, tokenIndex363, depth363 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l363
						}
						goto l362
					l363:
						position, tokenIndex, depth = position363, tokenIndex363, depth363
					}
					if buffer[position] != rune('`') {
						goto l361
					}
					position++
					goto l360
				l361:
					position, tokenIndex, depth = position360, tokenIndex360, depth360
					if !_rules[rule_]() {
						goto l358
					}
					{
						position364, tokenIndex364, depth364 := position, tokenIndex, depth
						{
							position365 := position
							depth++
							{
								position366, tokenIndex366, depth366 := position, tokenIndex, depth
								{
									position368, tokenIndex368, depth368 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l369
									}
									position++
									goto l368
								l369:
									position, tokenIndex, depth = position368, tokenIndex368, depth368
									if buffer[position] != rune('A') {
										goto l367
									}
									position++
								}
							l368:
								{
									position370, tokenIndex370, depth370 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l371
									}
									position++
									goto l370
								l371:
									position, tokenIndex, depth = position370, tokenIndex370, depth370
									if buffer[position] != rune('L') {
										goto l367
									}
									position++
								}
							l370:
								{
									position372, tokenIndex372, depth372 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l373
									}
									position++
									goto l372
								l373:
									position, tokenIndex, depth = position372, tokenIndex372, depth372
									if buffer[position] != rune('L') {
										goto l367
									}
									position++
								}
							l372:
								goto l366
							l367:
								position, tokenIndex, depth = position366, tokenIndex366, depth366
								{
									position375, tokenIndex375, depth375 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l376
									}
									position++
									goto l375
								l376:
									position, tokenIndex, depth = position375, tokenIndex375, depth375
									if buffer[position] != rune('A') {
										goto l374
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
										goto l374
									}
									position++
								}
							l377:
								{
									position379, tokenIndex379, depth379 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l380
									}
									position++
									goto l379
								l380:
									position, tokenIndex, depth = position379, tokenIndex379, depth379
									if buffer[position] != rune('D') {
										goto l374
									}
									position++
								}
							l379:
								goto l366
							l374:
								position, tokenIndex, depth = position366, tokenIndex366, depth366
								{
									position382, tokenIndex382, depth382 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l383
									}
									position++
									goto l382
								l383:
									position, tokenIndex, depth = position382, tokenIndex382, depth382
									if buffer[position] != rune('M') {
										goto l381
									}
									position++
								}
							l382:
								{
									position384, tokenIndex384, depth384 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l385
									}
									position++
									goto l384
								l385:
									position, tokenIndex, depth = position384, tokenIndex384, depth384
									if buffer[position] != rune('A') {
										goto l381
									}
									position++
								}
							l384:
								{
									position386, tokenIndex386, depth386 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l387
									}
									position++
									goto l386
								l387:
									position, tokenIndex, depth = position386, tokenIndex386, depth386
									if buffer[position] != rune('T') {
										goto l381
									}
									position++
								}
							l386:
								{
									position388, tokenIndex388, depth388 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l389
									}
									position++
									goto l388
								l389:
									position, tokenIndex, depth = position388, tokenIndex388, depth388
									if buffer[position] != rune('C') {
										goto l381
									}
									position++
								}
							l388:
								{
									position390, tokenIndex390, depth390 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l391
									}
									position++
									goto l390
								l391:
									position, tokenIndex, depth = position390, tokenIndex390, depth390
									if buffer[position] != rune('H') {
										goto l381
									}
									position++
								}
							l390:
								goto l366
							l381:
								position, tokenIndex, depth = position366, tokenIndex366, depth366
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
										goto l392
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
										goto l392
									}
									position++
								}
							l395:
								{
									position397, tokenIndex397, depth397 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l398
									}
									position++
									goto l397
								l398:
									position, tokenIndex, depth = position397, tokenIndex397, depth397
									if buffer[position] != rune('L') {
										goto l392
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
										goto l392
									}
									position++
								}
							l399:
								{
									position401, tokenIndex401, depth401 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l402
									}
									position++
									goto l401
								l402:
									position, tokenIndex, depth = position401, tokenIndex401, depth401
									if buffer[position] != rune('C') {
										goto l392
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
										goto l392
									}
									position++
								}
							l403:
								goto l366
							l392:
								position, tokenIndex, depth = position366, tokenIndex366, depth366
								{
									switch buffer[position] {
									case 'M', 'm':
										{
											position406, tokenIndex406, depth406 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l407
											}
											position++
											goto l406
										l407:
											position, tokenIndex, depth = position406, tokenIndex406, depth406
											if buffer[position] != rune('M') {
												goto l364
											}
											position++
										}
									l406:
										{
											position408, tokenIndex408, depth408 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l409
											}
											position++
											goto l408
										l409:
											position, tokenIndex, depth = position408, tokenIndex408, depth408
											if buffer[position] != rune('E') {
												goto l364
											}
											position++
										}
									l408:
										{
											position410, tokenIndex410, depth410 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l411
											}
											position++
											goto l410
										l411:
											position, tokenIndex, depth = position410, tokenIndex410, depth410
											if buffer[position] != rune('T') {
												goto l364
											}
											position++
										}
									l410:
										{
											position412, tokenIndex412, depth412 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l413
											}
											position++
											goto l412
										l413:
											position, tokenIndex, depth = position412, tokenIndex412, depth412
											if buffer[position] != rune('R') {
												goto l364
											}
											position++
										}
									l412:
										{
											position414, tokenIndex414, depth414 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l415
											}
											position++
											goto l414
										l415:
											position, tokenIndex, depth = position414, tokenIndex414, depth414
											if buffer[position] != rune('I') {
												goto l364
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
												goto l364
											}
											position++
										}
									l416:
										{
											position418, tokenIndex418, depth418 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l419
											}
											position++
											goto l418
										l419:
											position, tokenIndex, depth = position418, tokenIndex418, depth418
											if buffer[position] != rune('S') {
												goto l364
											}
											position++
										}
									l418:
										break
									case 'W', 'w':
										{
											position420, tokenIndex420, depth420 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l421
											}
											position++
											goto l420
										l421:
											position, tokenIndex, depth = position420, tokenIndex420, depth420
											if buffer[position] != rune('W') {
												goto l364
											}
											position++
										}
									l420:
										{
											position422, tokenIndex422, depth422 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l423
											}
											position++
											goto l422
										l423:
											position, tokenIndex, depth = position422, tokenIndex422, depth422
											if buffer[position] != rune('H') {
												goto l364
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
												goto l364
											}
											position++
										}
									l424:
										{
											position426, tokenIndex426, depth426 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l427
											}
											position++
											goto l426
										l427:
											position, tokenIndex, depth = position426, tokenIndex426, depth426
											if buffer[position] != rune('R') {
												goto l364
											}
											position++
										}
									l426:
										{
											position428, tokenIndex428, depth428 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l429
											}
											position++
											goto l428
										l429:
											position, tokenIndex, depth = position428, tokenIndex428, depth428
											if buffer[position] != rune('E') {
												goto l364
											}
											position++
										}
									l428:
										break
									case 'O', 'o':
										{
											position430, tokenIndex430, depth430 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l431
											}
											position++
											goto l430
										l431:
											position, tokenIndex, depth = position430, tokenIndex430, depth430
											if buffer[position] != rune('O') {
												goto l364
											}
											position++
										}
									l430:
										{
											position432, tokenIndex432, depth432 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l433
											}
											position++
											goto l432
										l433:
											position, tokenIndex, depth = position432, tokenIndex432, depth432
											if buffer[position] != rune('R') {
												goto l364
											}
											position++
										}
									l432:
										break
									case 'N', 'n':
										{
											position434, tokenIndex434, depth434 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l435
											}
											position++
											goto l434
										l435:
											position, tokenIndex, depth = position434, tokenIndex434, depth434
											if buffer[position] != rune('N') {
												goto l364
											}
											position++
										}
									l434:
										{
											position436, tokenIndex436, depth436 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l437
											}
											position++
											goto l436
										l437:
											position, tokenIndex, depth = position436, tokenIndex436, depth436
											if buffer[position] != rune('O') {
												goto l364
											}
											position++
										}
									l436:
										{
											position438, tokenIndex438, depth438 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l439
											}
											position++
											goto l438
										l439:
											position, tokenIndex, depth = position438, tokenIndex438, depth438
											if buffer[position] != rune('T') {
												goto l364
											}
											position++
										}
									l438:
										break
									case 'I', 'i':
										{
											position440, tokenIndex440, depth440 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l441
											}
											position++
											goto l440
										l441:
											position, tokenIndex, depth = position440, tokenIndex440, depth440
											if buffer[position] != rune('I') {
												goto l364
											}
											position++
										}
									l440:
										{
											position442, tokenIndex442, depth442 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l443
											}
											position++
											goto l442
										l443:
											position, tokenIndex, depth = position442, tokenIndex442, depth442
											if buffer[position] != rune('N') {
												goto l364
											}
											position++
										}
									l442:
										break
									case 'C', 'c':
										{
											position444, tokenIndex444, depth444 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l445
											}
											position++
											goto l444
										l445:
											position, tokenIndex, depth = position444, tokenIndex444, depth444
											if buffer[position] != rune('C') {
												goto l364
											}
											position++
										}
									l444:
										{
											position446, tokenIndex446, depth446 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l447
											}
											position++
											goto l446
										l447:
											position, tokenIndex, depth = position446, tokenIndex446, depth446
											if buffer[position] != rune('O') {
												goto l364
											}
											position++
										}
									l446:
										{
											position448, tokenIndex448, depth448 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l449
											}
											position++
											goto l448
										l449:
											position, tokenIndex, depth = position448, tokenIndex448, depth448
											if buffer[position] != rune('L') {
												goto l364
											}
											position++
										}
									l448:
										{
											position450, tokenIndex450, depth450 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l451
											}
											position++
											goto l450
										l451:
											position, tokenIndex, depth = position450, tokenIndex450, depth450
											if buffer[position] != rune('L') {
												goto l364
											}
											position++
										}
									l450:
										{
											position452, tokenIndex452, depth452 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l453
											}
											position++
											goto l452
										l453:
											position, tokenIndex, depth = position452, tokenIndex452, depth452
											if buffer[position] != rune('A') {
												goto l364
											}
											position++
										}
									l452:
										{
											position454, tokenIndex454, depth454 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l455
											}
											position++
											goto l454
										l455:
											position, tokenIndex, depth = position454, tokenIndex454, depth454
											if buffer[position] != rune('P') {
												goto l364
											}
											position++
										}
									l454:
										{
											position456, tokenIndex456, depth456 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l457
											}
											position++
											goto l456
										l457:
											position, tokenIndex, depth = position456, tokenIndex456, depth456
											if buffer[position] != rune('S') {
												goto l364
											}
											position++
										}
									l456:
										{
											position458, tokenIndex458, depth458 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l459
											}
											position++
											goto l458
										l459:
											position, tokenIndex, depth = position458, tokenIndex458, depth458
											if buffer[position] != rune('E') {
												goto l364
											}
											position++
										}
									l458:
										break
									case 'G', 'g':
										{
											position460, tokenIndex460, depth460 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l461
											}
											position++
											goto l460
										l461:
											position, tokenIndex, depth = position460, tokenIndex460, depth460
											if buffer[position] != rune('G') {
												goto l364
											}
											position++
										}
									l460:
										{
											position462, tokenIndex462, depth462 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l463
											}
											position++
											goto l462
										l463:
											position, tokenIndex, depth = position462, tokenIndex462, depth462
											if buffer[position] != rune('R') {
												goto l364
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
												goto l364
											}
											position++
										}
									l464:
										{
											position466, tokenIndex466, depth466 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l467
											}
											position++
											goto l466
										l467:
											position, tokenIndex, depth = position466, tokenIndex466, depth466
											if buffer[position] != rune('U') {
												goto l364
											}
											position++
										}
									l466:
										{
											position468, tokenIndex468, depth468 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l469
											}
											position++
											goto l468
										l469:
											position, tokenIndex, depth = position468, tokenIndex468, depth468
											if buffer[position] != rune('P') {
												goto l364
											}
											position++
										}
									l468:
										break
									case 'D', 'd':
										{
											position470, tokenIndex470, depth470 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l471
											}
											position++
											goto l470
										l471:
											position, tokenIndex, depth = position470, tokenIndex470, depth470
											if buffer[position] != rune('D') {
												goto l364
											}
											position++
										}
									l470:
										{
											position472, tokenIndex472, depth472 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l473
											}
											position++
											goto l472
										l473:
											position, tokenIndex, depth = position472, tokenIndex472, depth472
											if buffer[position] != rune('E') {
												goto l364
											}
											position++
										}
									l472:
										{
											position474, tokenIndex474, depth474 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l475
											}
											position++
											goto l474
										l475:
											position, tokenIndex, depth = position474, tokenIndex474, depth474
											if buffer[position] != rune('S') {
												goto l364
											}
											position++
										}
									l474:
										{
											position476, tokenIndex476, depth476 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l477
											}
											position++
											goto l476
										l477:
											position, tokenIndex, depth = position476, tokenIndex476, depth476
											if buffer[position] != rune('C') {
												goto l364
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
												goto l364
											}
											position++
										}
									l478:
										{
											position480, tokenIndex480, depth480 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l481
											}
											position++
											goto l480
										l481:
											position, tokenIndex, depth = position480, tokenIndex480, depth480
											if buffer[position] != rune('I') {
												goto l364
											}
											position++
										}
									l480:
										{
											position482, tokenIndex482, depth482 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l483
											}
											position++
											goto l482
										l483:
											position, tokenIndex, depth = position482, tokenIndex482, depth482
											if buffer[position] != rune('B') {
												goto l364
											}
											position++
										}
									l482:
										{
											position484, tokenIndex484, depth484 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l485
											}
											position++
											goto l484
										l485:
											position, tokenIndex, depth = position484, tokenIndex484, depth484
											if buffer[position] != rune('E') {
												goto l364
											}
											position++
										}
									l484:
										break
									case 'B', 'b':
										{
											position486, tokenIndex486, depth486 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l487
											}
											position++
											goto l486
										l487:
											position, tokenIndex, depth = position486, tokenIndex486, depth486
											if buffer[position] != rune('B') {
												goto l364
											}
											position++
										}
									l486:
										{
											position488, tokenIndex488, depth488 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l489
											}
											position++
											goto l488
										l489:
											position, tokenIndex, depth = position488, tokenIndex488, depth488
											if buffer[position] != rune('Y') {
												goto l364
											}
											position++
										}
									l488:
										break
									case 'A', 'a':
										{
											position490, tokenIndex490, depth490 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l491
											}
											position++
											goto l490
										l491:
											position, tokenIndex, depth = position490, tokenIndex490, depth490
											if buffer[position] != rune('A') {
												goto l364
											}
											position++
										}
									l490:
										{
											position492, tokenIndex492, depth492 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l493
											}
											position++
											goto l492
										l493:
											position, tokenIndex, depth = position492, tokenIndex492, depth492
											if buffer[position] != rune('S') {
												goto l364
											}
											position++
										}
									l492:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l364
										}
										break
									}
								}

							}
						l366:
							depth--
							add(ruleKEYWORD, position365)
						}
						if !_rules[ruleKEY]() {
							goto l364
						}
						goto l358
					l364:
						position, tokenIndex, depth = position364, tokenIndex364, depth364
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l358
					}
				l494:
					{
						position495, tokenIndex495, depth495 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l495
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l495
						}
						goto l494
					l495:
						position, tokenIndex, depth = position495, tokenIndex495, depth495
					}
				}
			l360:
				depth--
				add(ruleIDENTIFIER, position359)
			}
			return true
		l358:
			position, tokenIndex, depth = position358, tokenIndex358, depth358
			return false
		},
		/* 34 TIMESTAMP <- <((_ <(NUMBER ([a-z] / [A-Z])*)>) / (_ STRING) / (_ <(('n' / 'N') ('o' / 'O') ('w' / 'W'))>))> */
		nil,
		/* 35 ID_SEGMENT <- <(_ ID_START ID_CONT*)> */
		func() bool {
			position497, tokenIndex497, depth497 := position, tokenIndex, depth
			{
				position498 := position
				depth++
				if !_rules[rule_]() {
					goto l497
				}
				if !_rules[ruleID_START]() {
					goto l497
				}
			l499:
				{
					position500, tokenIndex500, depth500 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l500
					}
					goto l499
				l500:
					position, tokenIndex, depth = position500, tokenIndex500, depth500
				}
				depth--
				add(ruleID_SEGMENT, position498)
			}
			return true
		l497:
			position, tokenIndex, depth = position497, tokenIndex497, depth497
			return false
		},
		/* 36 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position501, tokenIndex501, depth501 := position, tokenIndex, depth
			{
				position502 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l501
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l501
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l501
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position502)
			}
			return true
		l501:
			position, tokenIndex, depth = position501, tokenIndex501, depth501
			return false
		},
		/* 37 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position504, tokenIndex504, depth504 := position, tokenIndex, depth
			{
				position505 := position
				depth++
				{
					position506, tokenIndex506, depth506 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l507
					}
					goto l506
				l507:
					position, tokenIndex, depth = position506, tokenIndex506, depth506
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l504
					}
					position++
				}
			l506:
				depth--
				add(ruleID_CONT, position505)
			}
			return true
		l504:
			position, tokenIndex, depth = position504, tokenIndex504, depth504
			return false
		},
		/* 38 PROPERTY_KEY <- <(((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> KEY _ (('b' / 'B') ('y' / 'Y')))) | (&('R' | 'r') <(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))>) | (&('T' | 't') <(('t' / 'T') ('o' / 'O'))>) | (&('F' | 'f') <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))>)) KEY)> */
		func() bool {
			position508, tokenIndex508, depth508 := position, tokenIndex, depth
			{
				position509 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position511 := position
							depth++
							{
								position512, tokenIndex512, depth512 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l513
								}
								position++
								goto l512
							l513:
								position, tokenIndex, depth = position512, tokenIndex512, depth512
								if buffer[position] != rune('S') {
									goto l508
								}
								position++
							}
						l512:
							{
								position514, tokenIndex514, depth514 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l515
								}
								position++
								goto l514
							l515:
								position, tokenIndex, depth = position514, tokenIndex514, depth514
								if buffer[position] != rune('A') {
									goto l508
								}
								position++
							}
						l514:
							{
								position516, tokenIndex516, depth516 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l517
								}
								position++
								goto l516
							l517:
								position, tokenIndex, depth = position516, tokenIndex516, depth516
								if buffer[position] != rune('M') {
									goto l508
								}
								position++
							}
						l516:
							{
								position518, tokenIndex518, depth518 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l519
								}
								position++
								goto l518
							l519:
								position, tokenIndex, depth = position518, tokenIndex518, depth518
								if buffer[position] != rune('P') {
									goto l508
								}
								position++
							}
						l518:
							{
								position520, tokenIndex520, depth520 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l521
								}
								position++
								goto l520
							l521:
								position, tokenIndex, depth = position520, tokenIndex520, depth520
								if buffer[position] != rune('L') {
									goto l508
								}
								position++
							}
						l520:
							{
								position522, tokenIndex522, depth522 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l523
								}
								position++
								goto l522
							l523:
								position, tokenIndex, depth = position522, tokenIndex522, depth522
								if buffer[position] != rune('E') {
									goto l508
								}
								position++
							}
						l522:
							depth--
							add(rulePegText, position511)
						}
						if !_rules[ruleKEY]() {
							goto l508
						}
						if !_rules[rule_]() {
							goto l508
						}
						{
							position524, tokenIndex524, depth524 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l525
							}
							position++
							goto l524
						l525:
							position, tokenIndex, depth = position524, tokenIndex524, depth524
							if buffer[position] != rune('B') {
								goto l508
							}
							position++
						}
					l524:
						{
							position526, tokenIndex526, depth526 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l527
							}
							position++
							goto l526
						l527:
							position, tokenIndex, depth = position526, tokenIndex526, depth526
							if buffer[position] != rune('Y') {
								goto l508
							}
							position++
						}
					l526:
						break
					case 'R', 'r':
						{
							position528 := position
							depth++
							{
								position529, tokenIndex529, depth529 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l530
								}
								position++
								goto l529
							l530:
								position, tokenIndex, depth = position529, tokenIndex529, depth529
								if buffer[position] != rune('R') {
									goto l508
								}
								position++
							}
						l529:
							{
								position531, tokenIndex531, depth531 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l532
								}
								position++
								goto l531
							l532:
								position, tokenIndex, depth = position531, tokenIndex531, depth531
								if buffer[position] != rune('E') {
									goto l508
								}
								position++
							}
						l531:
							{
								position533, tokenIndex533, depth533 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l534
								}
								position++
								goto l533
							l534:
								position, tokenIndex, depth = position533, tokenIndex533, depth533
								if buffer[position] != rune('S') {
									goto l508
								}
								position++
							}
						l533:
							{
								position535, tokenIndex535, depth535 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l536
								}
								position++
								goto l535
							l536:
								position, tokenIndex, depth = position535, tokenIndex535, depth535
								if buffer[position] != rune('O') {
									goto l508
								}
								position++
							}
						l535:
							{
								position537, tokenIndex537, depth537 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l538
								}
								position++
								goto l537
							l538:
								position, tokenIndex, depth = position537, tokenIndex537, depth537
								if buffer[position] != rune('L') {
									goto l508
								}
								position++
							}
						l537:
							{
								position539, tokenIndex539, depth539 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l540
								}
								position++
								goto l539
							l540:
								position, tokenIndex, depth = position539, tokenIndex539, depth539
								if buffer[position] != rune('U') {
									goto l508
								}
								position++
							}
						l539:
							{
								position541, tokenIndex541, depth541 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l542
								}
								position++
								goto l541
							l542:
								position, tokenIndex, depth = position541, tokenIndex541, depth541
								if buffer[position] != rune('T') {
									goto l508
								}
								position++
							}
						l541:
							{
								position543, tokenIndex543, depth543 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l544
								}
								position++
								goto l543
							l544:
								position, tokenIndex, depth = position543, tokenIndex543, depth543
								if buffer[position] != rune('I') {
									goto l508
								}
								position++
							}
						l543:
							{
								position545, tokenIndex545, depth545 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l546
								}
								position++
								goto l545
							l546:
								position, tokenIndex, depth = position545, tokenIndex545, depth545
								if buffer[position] != rune('O') {
									goto l508
								}
								position++
							}
						l545:
							{
								position547, tokenIndex547, depth547 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l548
								}
								position++
								goto l547
							l548:
								position, tokenIndex, depth = position547, tokenIndex547, depth547
								if buffer[position] != rune('N') {
									goto l508
								}
								position++
							}
						l547:
							depth--
							add(rulePegText, position528)
						}
						break
					case 'T', 't':
						{
							position549 := position
							depth++
							{
								position550, tokenIndex550, depth550 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l551
								}
								position++
								goto l550
							l551:
								position, tokenIndex, depth = position550, tokenIndex550, depth550
								if buffer[position] != rune('T') {
									goto l508
								}
								position++
							}
						l550:
							{
								position552, tokenIndex552, depth552 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l553
								}
								position++
								goto l552
							l553:
								position, tokenIndex, depth = position552, tokenIndex552, depth552
								if buffer[position] != rune('O') {
									goto l508
								}
								position++
							}
						l552:
							depth--
							add(rulePegText, position549)
						}
						break
					default:
						{
							position554 := position
							depth++
							{
								position555, tokenIndex555, depth555 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l556
								}
								position++
								goto l555
							l556:
								position, tokenIndex, depth = position555, tokenIndex555, depth555
								if buffer[position] != rune('F') {
									goto l508
								}
								position++
							}
						l555:
							{
								position557, tokenIndex557, depth557 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l558
								}
								position++
								goto l557
							l558:
								position, tokenIndex, depth = position557, tokenIndex557, depth557
								if buffer[position] != rune('R') {
									goto l508
								}
								position++
							}
						l557:
							{
								position559, tokenIndex559, depth559 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l560
								}
								position++
								goto l559
							l560:
								position, tokenIndex, depth = position559, tokenIndex559, depth559
								if buffer[position] != rune('O') {
									goto l508
								}
								position++
							}
						l559:
							{
								position561, tokenIndex561, depth561 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l562
								}
								position++
								goto l561
							l562:
								position, tokenIndex, depth = position561, tokenIndex561, depth561
								if buffer[position] != rune('M') {
									goto l508
								}
								position++
							}
						l561:
							depth--
							add(rulePegText, position554)
						}
						break
					}
				}

				if !_rules[ruleKEY]() {
					goto l508
				}
				depth--
				add(rulePROPERTY_KEY, position509)
			}
			return true
		l508:
			position, tokenIndex, depth = position508, tokenIndex508, depth508
			return false
		},
		/* 39 PROPERTY_VALUE <- <TIMESTAMP> */
		nil,
		/* 40 KEYWORD <- <((('a' / 'A') ('l' / 'L') ('l' / 'L')) / (('a' / 'A') ('n' / 'N') ('d' / 'D')) / (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H')) / (('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T')) / ((&('M' | 'm') (('m' / 'M') ('e' / 'E') ('t' / 'T') ('r' / 'R') ('i' / 'I') ('c' / 'C') ('s' / 'S'))) | (&('W' | 'w') (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E'))) | (&('O' | 'o') (('o' / 'O') ('r' / 'R'))) | (&('N' | 'n') (('n' / 'N') ('o' / 'O') ('t' / 'T'))) | (&('I' | 'i') (('i' / 'I') ('n' / 'N'))) | (&('C' | 'c') (('c' / 'C') ('o' / 'O') ('l' / 'L') ('l' / 'L') ('a' / 'A') ('p' / 'P') ('s' / 'S') ('e' / 'E'))) | (&('G' | 'g') (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P'))) | (&('D' | 'd') (('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E'))) | (&('B' | 'b') (('b' / 'B') ('y' / 'Y'))) | (&('A' | 'a') (('a' / 'A') ('s' / 'S'))) | (&('F' | 'R' | 'S' | 'T' | 'f' | 'r' | 's' | 't') PROPERTY_KEY)))> */
		nil,
		/* 41 OP_PIPE <- <'|'> */
		nil,
		/* 42 OP_ADD <- <'+'> */
		nil,
		/* 43 OP_SUB <- <'-'> */
		nil,
		/* 44 OP_MULT <- <'*'> */
		nil,
		/* 45 OP_DIV <- <'/'> */
		nil,
		/* 46 OP_AND <- <(('a' / 'A') ('n' / 'N') ('d' / 'D') KEY)> */
		nil,
		/* 47 OP_OR <- <(('o' / 'O') ('r' / 'R') KEY)> */
		nil,
		/* 48 OP_NOT <- <(('n' / 'N') ('o' / 'O') ('t' / 'T') KEY)> */
		nil,
		/* 49 QUOTE_SINGLE <- <'\''> */
		func() bool {
			position573, tokenIndex573, depth573 := position, tokenIndex, depth
			{
				position574 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l573
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position574)
			}
			return true
		l573:
			position, tokenIndex, depth = position573, tokenIndex573, depth573
			return false
		},
		/* 50 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position575, tokenIndex575, depth575 := position, tokenIndex, depth
			{
				position576 := position
				depth++
				if buffer[position] != rune('"') {
					goto l575
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position576)
			}
			return true
		l575:
			position, tokenIndex, depth = position575, tokenIndex575, depth575
			return false
		},
		/* 51 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> QUOTE_SINGLE) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> QUOTE_DOUBLE))> */
		func() bool {
			position577, tokenIndex577, depth577 := position, tokenIndex, depth
			{
				position578 := position
				depth++
				{
					position579, tokenIndex579, depth579 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l580
					}
					{
						position581 := position
						depth++
					l582:
						{
							position583, tokenIndex583, depth583 := position, tokenIndex, depth
							{
								position584, tokenIndex584, depth584 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l584
								}
								goto l583
							l584:
								position, tokenIndex, depth = position584, tokenIndex584, depth584
							}
							if !_rules[ruleCHAR]() {
								goto l583
							}
							goto l582
						l583:
							position, tokenIndex, depth = position583, tokenIndex583, depth583
						}
						depth--
						add(rulePegText, position581)
					}
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l580
					}
					goto l579
				l580:
					position, tokenIndex, depth = position579, tokenIndex579, depth579
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l577
					}
					{
						position585 := position
						depth++
					l586:
						{
							position587, tokenIndex587, depth587 := position, tokenIndex, depth
							{
								position588, tokenIndex588, depth588 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l588
								}
								goto l587
							l588:
								position, tokenIndex, depth = position588, tokenIndex588, depth588
							}
							if !_rules[ruleCHAR]() {
								goto l587
							}
							goto l586
						l587:
							position, tokenIndex, depth = position587, tokenIndex587, depth587
						}
						depth--
						add(rulePegText, position585)
					}
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l577
					}
				}
			l579:
				depth--
				add(ruleSTRING, position578)
			}
			return true
		l577:
			position, tokenIndex, depth = position577, tokenIndex577, depth577
			return false
		},
		/* 52 CHAR <- <(('\\' ((&('"') QUOTE_DOUBLE) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position589, tokenIndex589, depth589 := position, tokenIndex, depth
			{
				position590 := position
				depth++
				{
					position591, tokenIndex591, depth591 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l592
					}
					position++
					{
						switch buffer[position] {
						case '"':
							if !_rules[ruleQUOTE_DOUBLE]() {
								goto l592
							}
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l592
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l592
							}
							break
						}
					}

					goto l591
				l592:
					position, tokenIndex, depth = position591, tokenIndex591, depth591
					{
						position594, tokenIndex594, depth594 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l594
						}
						goto l589
					l594:
						position, tokenIndex, depth = position594, tokenIndex594, depth594
					}
					if !matchDot() {
						goto l589
					}
				}
			l591:
				depth--
				add(ruleCHAR, position590)
			}
			return true
		l589:
			position, tokenIndex, depth = position589, tokenIndex589, depth589
			return false
		},
		/* 53 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position595, tokenIndex595, depth595 := position, tokenIndex, depth
			{
				position596 := position
				depth++
				{
					position597, tokenIndex597, depth597 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l598
					}
					position++
					goto l597
				l598:
					position, tokenIndex, depth = position597, tokenIndex597, depth597
					if buffer[position] != rune('\\') {
						goto l595
					}
					position++
				}
			l597:
				depth--
				add(ruleESCAPE_CLASS, position596)
			}
			return true
		l595:
			position, tokenIndex, depth = position595, tokenIndex595, depth595
			return false
		},
		/* 54 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position599, tokenIndex599, depth599 := position, tokenIndex, depth
			{
				position600 := position
				depth++
				{
					position601 := position
					depth++
					{
						position602, tokenIndex602, depth602 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l602
						}
						position++
						goto l603
					l602:
						position, tokenIndex, depth = position602, tokenIndex602, depth602
					}
				l603:
					{
						position604 := position
						depth++
						{
							position605, tokenIndex605, depth605 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l606
							}
							position++
							goto l605
						l606:
							position, tokenIndex, depth = position605, tokenIndex605, depth605
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l599
							}
							position++
						l607:
							{
								position608, tokenIndex608, depth608 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l608
								}
								position++
								goto l607
							l608:
								position, tokenIndex, depth = position608, tokenIndex608, depth608
							}
						}
					l605:
						depth--
						add(ruleNUMBER_NATURAL, position604)
					}
					depth--
					add(ruleNUMBER_INTEGER, position601)
				}
				{
					position609, tokenIndex609, depth609 := position, tokenIndex, depth
					{
						position611 := position
						depth++
						if buffer[position] != rune('.') {
							goto l609
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l609
						}
						position++
					l612:
						{
							position613, tokenIndex613, depth613 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l613
							}
							position++
							goto l612
						l613:
							position, tokenIndex, depth = position613, tokenIndex613, depth613
						}
						depth--
						add(ruleNUMBER_FRACTION, position611)
					}
					goto l610
				l609:
					position, tokenIndex, depth = position609, tokenIndex609, depth609
				}
			l610:
				{
					position614, tokenIndex614, depth614 := position, tokenIndex, depth
					{
						position616 := position
						depth++
						{
							position617, tokenIndex617, depth617 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l618
							}
							position++
							goto l617
						l618:
							position, tokenIndex, depth = position617, tokenIndex617, depth617
							if buffer[position] != rune('E') {
								goto l614
							}
							position++
						}
					l617:
						{
							position619, tokenIndex619, depth619 := position, tokenIndex, depth
							{
								position621, tokenIndex621, depth621 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l622
								}
								position++
								goto l621
							l622:
								position, tokenIndex, depth = position621, tokenIndex621, depth621
								if buffer[position] != rune('-') {
									goto l619
								}
								position++
							}
						l621:
							goto l620
						l619:
							position, tokenIndex, depth = position619, tokenIndex619, depth619
						}
					l620:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l614
						}
						position++
					l623:
						{
							position624, tokenIndex624, depth624 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l624
							}
							position++
							goto l623
						l624:
							position, tokenIndex, depth = position624, tokenIndex624, depth624
						}
						depth--
						add(ruleNUMBER_EXP, position616)
					}
					goto l615
				l614:
					position, tokenIndex, depth = position614, tokenIndex614, depth614
				}
			l615:
				depth--
				add(ruleNUMBER, position600)
			}
			return true
		l599:
			position, tokenIndex, depth = position599, tokenIndex599, depth599
			return false
		},
		/* 55 NUMBER_NATURAL <- <('0' / ([1-9] [0-9]*))> */
		nil,
		/* 56 NUMBER_FRACTION <- <('.' [0-9]+)> */
		nil,
		/* 57 NUMBER_INTEGER <- <('-'? NUMBER_NATURAL)> */
		nil,
		/* 58 NUMBER_EXP <- <(('e' / 'E') ('+' / '-')? [0-9]+)> */
		nil,
		/* 59 DURATION <- <(NUMBER [a-z]+ KEY)> */
		nil,
		/* 60 PAREN_OPEN <- <'('> */
		func() bool {
			position630, tokenIndex630, depth630 := position, tokenIndex, depth
			{
				position631 := position
				depth++
				if buffer[position] != rune('(') {
					goto l630
				}
				position++
				depth--
				add(rulePAREN_OPEN, position631)
			}
			return true
		l630:
			position, tokenIndex, depth = position630, tokenIndex630, depth630
			return false
		},
		/* 61 PAREN_CLOSE <- <')'> */
		func() bool {
			position632, tokenIndex632, depth632 := position, tokenIndex, depth
			{
				position633 := position
				depth++
				if buffer[position] != rune(')') {
					goto l632
				}
				position++
				depth--
				add(rulePAREN_CLOSE, position633)
			}
			return true
		l632:
			position, tokenIndex, depth = position632, tokenIndex632, depth632
			return false
		},
		/* 62 COMMA <- <','> */
		func() bool {
			position634, tokenIndex634, depth634 := position, tokenIndex, depth
			{
				position635 := position
				depth++
				if buffer[position] != rune(',') {
					goto l634
				}
				position++
				depth--
				add(ruleCOMMA, position635)
			}
			return true
		l634:
			position, tokenIndex, depth = position634, tokenIndex634, depth634
			return false
		},
		/* 63 _ <- <((&('/') COMMENT_BLOCK) | (&('-') COMMENT_TRAIL) | (&('\t' | '\n' | ' ') SPACE))*> */
		func() bool {
			{
				position637 := position
				depth++
			l638:
				{
					position639, tokenIndex639, depth639 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '/':
							{
								position641 := position
								depth++
								if buffer[position] != rune('/') {
									goto l639
								}
								position++
								if buffer[position] != rune('*') {
									goto l639
								}
								position++
							l642:
								{
									position643, tokenIndex643, depth643 := position, tokenIndex, depth
									{
										position644, tokenIndex644, depth644 := position, tokenIndex, depth
										if buffer[position] != rune('*') {
											goto l644
										}
										position++
										if buffer[position] != rune('/') {
											goto l644
										}
										position++
										goto l643
									l644:
										position, tokenIndex, depth = position644, tokenIndex644, depth644
									}
									if !matchDot() {
										goto l643
									}
									goto l642
								l643:
									position, tokenIndex, depth = position643, tokenIndex643, depth643
								}
								if buffer[position] != rune('*') {
									goto l639
								}
								position++
								if buffer[position] != rune('/') {
									goto l639
								}
								position++
								depth--
								add(ruleCOMMENT_BLOCK, position641)
							}
							break
						case '-':
							{
								position645 := position
								depth++
								if buffer[position] != rune('-') {
									goto l639
								}
								position++
								if buffer[position] != rune('-') {
									goto l639
								}
								position++
							l646:
								{
									position647, tokenIndex647, depth647 := position, tokenIndex, depth
									{
										position648, tokenIndex648, depth648 := position, tokenIndex, depth
										if buffer[position] != rune('\n') {
											goto l648
										}
										position++
										goto l647
									l648:
										position, tokenIndex, depth = position648, tokenIndex648, depth648
									}
									if !matchDot() {
										goto l647
									}
									goto l646
								l647:
									position, tokenIndex, depth = position647, tokenIndex647, depth647
								}
								depth--
								add(ruleCOMMENT_TRAIL, position645)
							}
							break
						default:
							{
								position649 := position
								depth++
								{
									switch buffer[position] {
									case '\t':
										if buffer[position] != rune('\t') {
											goto l639
										}
										position++
										break
									case '\n':
										if buffer[position] != rune('\n') {
											goto l639
										}
										position++
										break
									default:
										if buffer[position] != rune(' ') {
											goto l639
										}
										position++
										break
									}
								}

								depth--
								add(ruleSPACE, position649)
							}
							break
						}
					}

					goto l638
				l639:
					position, tokenIndex, depth = position639, tokenIndex639, depth639
				}
				depth--
				add(rule_, position637)
			}
			return true
		},
		/* 64 COMMENT_TRAIL <- <('-' '-' (!'\n' .)*)> */
		nil,
		/* 65 COMMENT_BLOCK <- <('/' '*' (!('*' '/') .)* ('*' '/'))> */
		nil,
		/* 66 KEY <- <!ID_CONT> */
		func() bool {
			position653, tokenIndex653, depth653 := position, tokenIndex, depth
			{
				position654 := position
				depth++
				{
					position655, tokenIndex655, depth655 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l655
					}
					goto l653
				l655:
					position, tokenIndex, depth = position655, tokenIndex655, depth655
				}
				depth--
				add(ruleKEY, position654)
			}
			return true
		l653:
			position, tokenIndex, depth = position653, tokenIndex653, depth653
			return false
		},
		/* 67 SPACE <- <((&('\t') '\t') | (&('\n') '\n') | (&(' ') ' '))> */
		nil,
		/* 69 Action0 <- <{
		   p.makeSelect()
		 }> */
		nil,
		/* 70 Action1 <- <{ p.makeDescribeAll() }> */
		nil,
		/* 71 Action2 <- <{ p.addNullMatchClause() }> */
		nil,
		/* 72 Action3 <- <{ p.addMatchClause() }> */
		nil,
		/* 73 Action4 <- <{ p.makeDescribeMetrics() }> */
		nil,
		nil,
		/* 75 Action5 <- <{ p.addStringLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 76 Action6 <- <{ p.makeDescribe() }> */
		nil,
		/* 77 Action7 <- <{ p.addEvaluationContext() }> */
		nil,
		/* 78 Action8 <- <{ p.addPropertyKey(buffer[begin:end])   }> */
		nil,
		/* 79 Action9 <- <{ p.addPropertyValue(buffer[begin:end]) }> */
		nil,
		/* 80 Action10 <- <{ p.insertPropertyKeyValue() }> */
		nil,
		/* 81 Action11 <- <{ p.checkPropertyClause() }> */
		nil,
		/* 82 Action12 <- <{ p.addNullPredicate() }> */
		nil,
		/* 83 Action13 <- <{ p.addExpressionList() }> */
		nil,
		/* 84 Action14 <- <{ p.appendExpression() }> */
		nil,
		/* 85 Action15 <- <{ p.appendExpression() }> */
		nil,
		/* 86 Action16 <- <{ p.addOperatorLiteral("+") }> */
		nil,
		/* 87 Action17 <- <{ p.addOperatorLiteral("-") }> */
		nil,
		/* 88 Action18 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 89 Action19 <- <{ p.addOperatorLiteral("/") }> */
		nil,
		/* 90 Action20 <- <{ p.addOperatorLiteral("*") }> */
		nil,
		/* 91 Action21 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 92 Action22 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 93 Action23 <- <{p.addExpressionList()}> */
		nil,
		/* 94 Action24 <- <{
		   p.addExpressionList()
		   p.addGroupBy()
		 }> */
		nil,
		/* 95 Action25 <- <{
		   p.addPipeExpression()
		 }> */
		nil,
		/* 96 Action26 <- <{ p.addDurationNode(text) }> */
		nil,
		/* 97 Action27 <- <{ p.addNumberNode(buffer[begin:end]) }> */
		nil,
		/* 98 Action28 <- <{ p.addStringNode(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 99 Action29 <- <{ p.addGroupBy() }> */
		nil,
		/* 100 Action30 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 101 Action31 <- <{
		   p.addFunctionInvocation()
		 }> */
		nil,
		/* 102 Action32 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 103 Action33 <- <{ p.addNullPredicate() }> */
		nil,
		/* 104 Action34 <- <{
		   p.addMetricExpression()
		 }> */
		nil,
		/* 105 Action35 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 106 Action36 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 107 Action37 <- <{
		   p.appendCollapseBy(unescapeLiteral(text))
		 }> */
		nil,
		/* 108 Action38 <- <{p.appendCollapseBy(unescapeLiteral(text))}> */
		nil,
		/* 109 Action39 <- <{ p.addOrPredicate() }> */
		nil,
		/* 110 Action40 <- <{ p.addAndPredicate() }> */
		nil,
		/* 111 Action41 <- <{ p.addNotPredicate() }> */
		nil,
		/* 112 Action42 <- <{
		   p.addLiteralMatcher()
		 }> */
		nil,
		/* 113 Action43 <- <{
		   p.addLiteralMatcher()
		   p.addNotPredicate()
		 }> */
		nil,
		/* 114 Action44 <- <{
		   p.addRegexMatcher()
		 }> */
		nil,
		/* 115 Action45 <- <{
		   p.addListMatcher()
		 }> */
		nil,
		/* 116 Action46 <- <{
		  p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 117 Action47 <- <{ p.addLiteralList() }> */
		nil,
		/* 118 Action48 <- <{
		  p.appendLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 119 Action49 <- <{ p.addTagLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
	}
	p.rules = _rules
}
