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
	ruleAction50

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
	"Action50",

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
	rules  [121]func() bool
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
			p.addIndexClause()
		case ruleAction8:
			p.addEvaluationContext()
		case ruleAction9:
			p.addPropertyKey(buffer[begin:end])
		case ruleAction10:
			p.addPropertyValue(buffer[begin:end])
		case ruleAction11:
			p.insertPropertyKeyValue()
		case ruleAction12:
			p.checkPropertyClause()
		case ruleAction13:
			p.addNullPredicate()
		case ruleAction14:
			p.addExpressionList()
		case ruleAction15:
			p.appendExpression()
		case ruleAction16:
			p.appendExpression()
		case ruleAction17:
			p.addOperatorLiteral("+")
		case ruleAction18:
			p.addOperatorLiteral("-")
		case ruleAction19:
			p.addOperatorFunction()
		case ruleAction20:
			p.addOperatorLiteral("/")
		case ruleAction21:
			p.addOperatorLiteral("*")
		case ruleAction22:
			p.addOperatorFunction()
		case ruleAction23:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction24:
			p.addExpressionList()
		case ruleAction25:

			p.addExpressionList()
			p.addGroupBy()

		case ruleAction26:

			p.addPipeExpression()

		case ruleAction27:
			p.addDurationNode(text)
		case ruleAction28:
			p.addNumberNode(buffer[begin:end])
		case ruleAction29:
			p.addStringNode(unescapeLiteral(buffer[begin:end]))
		case ruleAction30:
			p.addGroupBy()
		case ruleAction31:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction32:

			p.addFunctionInvocation()

		case ruleAction33:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction34:
			p.addNullPredicate()
		case ruleAction35:

			p.addMetricExpression()

		case ruleAction36:

			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))

		case ruleAction37:

			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))

		case ruleAction38:

			p.appendCollapseBy(unescapeLiteral(text))

		case ruleAction39:
			p.appendCollapseBy(unescapeLiteral(text))
		case ruleAction40:
			p.addOrPredicate()
		case ruleAction41:
			p.addAndPredicate()
		case ruleAction42:
			p.addNotPredicate()
		case ruleAction43:

			p.addLiteralMatcher()

		case ruleAction44:

			p.addLiteralMatcher()
			p.addNotPredicate()

		case ruleAction45:

			p.addRegexMatcher()

		case ruleAction46:

			p.addListMatcher()

		case ruleAction47:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction48:
			p.addLiteralList()
		case ruleAction49:

			p.appendLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction50:
			p.addTagLiteral(unescapeLiteral(buffer[begin:end]))

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
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
								add(ruleAction8, position)
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
									add(ruleAction9, position)
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
									add(ruleAction10, position)
								}
								{
									add(ruleAction11, position)
								}
								goto l21
							l22:
								position, tokenIndex, depth = position22, tokenIndex22, depth22
							}
							{
								add(ruleAction12, position)
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
								{
									position120, tokenIndex120, depth120 := position, tokenIndex, depth
									{
										position122, tokenIndex122, depth122 := position, tokenIndex, depth
										if buffer[position] != rune('i') {
											goto l123
										}
										position++
										goto l122
									l123:
										position, tokenIndex, depth = position122, tokenIndex122, depth122
										if buffer[position] != rune('I') {
											goto l120
										}
										position++
									}
								l122:
									{
										position124, tokenIndex124, depth124 := position, tokenIndex, depth
										if buffer[position] != rune('n') {
											goto l125
										}
										position++
										goto l124
									l125:
										position, tokenIndex, depth = position124, tokenIndex124, depth124
										if buffer[position] != rune('N') {
											goto l120
										}
										position++
									}
								l124:
									{
										position126, tokenIndex126, depth126 := position, tokenIndex, depth
										if buffer[position] != rune('d') {
											goto l127
										}
										position++
										goto l126
									l127:
										position, tokenIndex, depth = position126, tokenIndex126, depth126
										if buffer[position] != rune('D') {
											goto l120
										}
										position++
									}
								l126:
									{
										position128, tokenIndex128, depth128 := position, tokenIndex, depth
										if buffer[position] != rune('e') {
											goto l129
										}
										position++
										goto l128
									l129:
										position, tokenIndex, depth = position128, tokenIndex128, depth128
										if buffer[position] != rune('E') {
											goto l120
										}
										position++
									}
								l128:
									{
										position130, tokenIndex130, depth130 := position, tokenIndex, depth
										if buffer[position] != rune('x') {
											goto l131
										}
										position++
										goto l130
									l131:
										position, tokenIndex, depth = position130, tokenIndex130, depth130
										if buffer[position] != rune('X') {
											goto l120
										}
										position++
									}
								l130:
									{
										add(ruleAction7, position)
									}
									goto l121
								l120:
									position, tokenIndex, depth = position120, tokenIndex120, depth120
								}
							l121:
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
					position133, tokenIndex133, depth133 := position, tokenIndex, depth
					if !matchDot() {
						goto l133
					}
					goto l0
				l133:
					position, tokenIndex, depth = position133, tokenIndex133, depth133
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
		/* 7 describeSingleStmt <- <(_ <METRIC_NAME> Action5 optionalPredicateClause Action6 (('i' / 'I') ('n' / 'N') ('d' / 'D') ('e' / 'E') ('x' / 'X') Action7)?)> */
		nil,
		/* 8 propertyClause <- <(Action8 (_ PROPERTY_KEY Action9 _ PROPERTY_VALUE Action10 Action11)* Action12)> */
		nil,
		/* 9 optionalPredicateClause <- <(predicateClause / Action13)> */
		func() bool {
			{
				position143 := position
				depth++
				{
					position144, tokenIndex144, depth144 := position, tokenIndex, depth
					{
						position146 := position
						depth++
						if !_rules[rule_]() {
							goto l145
						}
						{
							position147, tokenIndex147, depth147 := position, tokenIndex, depth
							if buffer[position] != rune('w') {
								goto l148
							}
							position++
							goto l147
						l148:
							position, tokenIndex, depth = position147, tokenIndex147, depth147
							if buffer[position] != rune('W') {
								goto l145
							}
							position++
						}
					l147:
						{
							position149, tokenIndex149, depth149 := position, tokenIndex, depth
							if buffer[position] != rune('h') {
								goto l150
							}
							position++
							goto l149
						l150:
							position, tokenIndex, depth = position149, tokenIndex149, depth149
							if buffer[position] != rune('H') {
								goto l145
							}
							position++
						}
					l149:
						{
							position151, tokenIndex151, depth151 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l152
							}
							position++
							goto l151
						l152:
							position, tokenIndex, depth = position151, tokenIndex151, depth151
							if buffer[position] != rune('E') {
								goto l145
							}
							position++
						}
					l151:
						{
							position153, tokenIndex153, depth153 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l154
							}
							position++
							goto l153
						l154:
							position, tokenIndex, depth = position153, tokenIndex153, depth153
							if buffer[position] != rune('R') {
								goto l145
							}
							position++
						}
					l153:
						{
							position155, tokenIndex155, depth155 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l156
							}
							position++
							goto l155
						l156:
							position, tokenIndex, depth = position155, tokenIndex155, depth155
							if buffer[position] != rune('E') {
								goto l145
							}
							position++
						}
					l155:
						if !_rules[ruleKEY]() {
							goto l145
						}
						if !_rules[rule_]() {
							goto l145
						}
						if !_rules[rulepredicate_1]() {
							goto l145
						}
						depth--
						add(rulepredicateClause, position146)
					}
					goto l144
				l145:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
					{
						add(ruleAction13, position)
					}
				}
			l144:
				depth--
				add(ruleoptionalPredicateClause, position143)
			}
			return true
		},
		/* 10 expressionList <- <(Action14 expression_start Action15 (_ COMMA expression_start Action16)*)> */
		func() bool {
			position158, tokenIndex158, depth158 := position, tokenIndex, depth
			{
				position159 := position
				depth++
				{
					add(ruleAction14, position)
				}
				if !_rules[ruleexpression_start]() {
					goto l158
				}
				{
					add(ruleAction15, position)
				}
			l162:
				{
					position163, tokenIndex163, depth163 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l163
					}
					if !_rules[ruleCOMMA]() {
						goto l163
					}
					if !_rules[ruleexpression_start]() {
						goto l163
					}
					{
						add(ruleAction16, position)
					}
					goto l162
				l163:
					position, tokenIndex, depth = position163, tokenIndex163, depth163
				}
				depth--
				add(ruleexpressionList, position159)
			}
			return true
		l158:
			position, tokenIndex, depth = position158, tokenIndex158, depth158
			return false
		},
		/* 11 expression_start <- <(expression_sum add_pipe)> */
		func() bool {
			position165, tokenIndex165, depth165 := position, tokenIndex, depth
			{
				position166 := position
				depth++
				{
					position167 := position
					depth++
					if !_rules[ruleexpression_product]() {
						goto l165
					}
				l168:
					{
						position169, tokenIndex169, depth169 := position, tokenIndex, depth
						if !_rules[ruleadd_pipe]() {
							goto l169
						}
						{
							position170, tokenIndex170, depth170 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l171
							}
							{
								position172 := position
								depth++
								if buffer[position] != rune('+') {
									goto l171
								}
								position++
								depth--
								add(ruleOP_ADD, position172)
							}
							{
								add(ruleAction17, position)
							}
							goto l170
						l171:
							position, tokenIndex, depth = position170, tokenIndex170, depth170
							if !_rules[rule_]() {
								goto l169
							}
							{
								position174 := position
								depth++
								if buffer[position] != rune('-') {
									goto l169
								}
								position++
								depth--
								add(ruleOP_SUB, position174)
							}
							{
								add(ruleAction18, position)
							}
						}
					l170:
						if !_rules[ruleexpression_product]() {
							goto l169
						}
						{
							add(ruleAction19, position)
						}
						goto l168
					l169:
						position, tokenIndex, depth = position169, tokenIndex169, depth169
					}
					depth--
					add(ruleexpression_sum, position167)
				}
				if !_rules[ruleadd_pipe]() {
					goto l165
				}
				depth--
				add(ruleexpression_start, position166)
			}
			return true
		l165:
			position, tokenIndex, depth = position165, tokenIndex165, depth165
			return false
		},
		/* 12 expression_sum <- <(expression_product (add_pipe ((_ OP_ADD Action17) / (_ OP_SUB Action18)) expression_product Action19)*)> */
		nil,
		/* 13 expression_product <- <(expression_atom (add_pipe ((_ OP_DIV Action20) / (_ OP_MULT Action21)) expression_atom Action22)*)> */
		func() bool {
			position178, tokenIndex178, depth178 := position, tokenIndex, depth
			{
				position179 := position
				depth++
				if !_rules[ruleexpression_atom]() {
					goto l178
				}
			l180:
				{
					position181, tokenIndex181, depth181 := position, tokenIndex, depth
					if !_rules[ruleadd_pipe]() {
						goto l181
					}
					{
						position182, tokenIndex182, depth182 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l183
						}
						{
							position184 := position
							depth++
							if buffer[position] != rune('/') {
								goto l183
							}
							position++
							depth--
							add(ruleOP_DIV, position184)
						}
						{
							add(ruleAction20, position)
						}
						goto l182
					l183:
						position, tokenIndex, depth = position182, tokenIndex182, depth182
						if !_rules[rule_]() {
							goto l181
						}
						{
							position186 := position
							depth++
							if buffer[position] != rune('*') {
								goto l181
							}
							position++
							depth--
							add(ruleOP_MULT, position186)
						}
						{
							add(ruleAction21, position)
						}
					}
				l182:
					if !_rules[ruleexpression_atom]() {
						goto l181
					}
					{
						add(ruleAction22, position)
					}
					goto l180
				l181:
					position, tokenIndex, depth = position181, tokenIndex181, depth181
				}
				depth--
				add(ruleexpression_product, position179)
			}
			return true
		l178:
			position, tokenIndex, depth = position178, tokenIndex178, depth178
			return false
		},
		/* 14 add_pipe <- <(_ OP_PIPE _ <IDENTIFIER> Action23 ((_ PAREN_OPEN (expressionList / Action24) optionalGroupBy _ PAREN_CLOSE) / Action25) Action26)*> */
		func() bool {
			{
				position190 := position
				depth++
			l191:
				{
					position192, tokenIndex192, depth192 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l192
					}
					{
						position193 := position
						depth++
						if buffer[position] != rune('|') {
							goto l192
						}
						position++
						depth--
						add(ruleOP_PIPE, position193)
					}
					if !_rules[rule_]() {
						goto l192
					}
					{
						position194 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l192
						}
						depth--
						add(rulePegText, position194)
					}
					{
						add(ruleAction23, position)
					}
					{
						position196, tokenIndex196, depth196 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l197
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l197
						}
						{
							position198, tokenIndex198, depth198 := position, tokenIndex, depth
							if !_rules[ruleexpressionList]() {
								goto l199
							}
							goto l198
						l199:
							position, tokenIndex, depth = position198, tokenIndex198, depth198
							{
								add(ruleAction24, position)
							}
						}
					l198:
						if !_rules[ruleoptionalGroupBy]() {
							goto l197
						}
						if !_rules[rule_]() {
							goto l197
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l197
						}
						goto l196
					l197:
						position, tokenIndex, depth = position196, tokenIndex196, depth196
						{
							add(ruleAction25, position)
						}
					}
				l196:
					{
						add(ruleAction26, position)
					}
					goto l191
				l192:
					position, tokenIndex, depth = position192, tokenIndex192, depth192
				}
				depth--
				add(ruleadd_pipe, position190)
			}
			return true
		},
		/* 15 expression_atom <- <(expression_function / expression_metric / (_ PAREN_OPEN expression_start _ PAREN_CLOSE) / (_ <DURATION> Action27) / (_ <NUMBER> Action28) / (_ STRING Action29))> */
		func() bool {
			position203, tokenIndex203, depth203 := position, tokenIndex, depth
			{
				position204 := position
				depth++
				{
					position205, tokenIndex205, depth205 := position, tokenIndex, depth
					{
						position207 := position
						depth++
						if !_rules[rule_]() {
							goto l206
						}
						{
							position208 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l206
							}
							depth--
							add(rulePegText, position208)
						}
						{
							add(ruleAction31, position)
						}
						if !_rules[rule_]() {
							goto l206
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l206
						}
						if !_rules[ruleexpressionList]() {
							goto l206
						}
						if !_rules[ruleoptionalGroupBy]() {
							goto l206
						}
						if !_rules[rule_]() {
							goto l206
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l206
						}
						{
							add(ruleAction32, position)
						}
						depth--
						add(ruleexpression_function, position207)
					}
					goto l205
				l206:
					position, tokenIndex, depth = position205, tokenIndex205, depth205
					{
						position212 := position
						depth++
						if !_rules[rule_]() {
							goto l211
						}
						{
							position213 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l211
							}
							depth--
							add(rulePegText, position213)
						}
						{
							add(ruleAction33, position)
						}
						{
							position215, tokenIndex215, depth215 := position, tokenIndex, depth
							{
								position217, tokenIndex217, depth217 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l218
								}
								if buffer[position] != rune('[') {
									goto l218
								}
								position++
								if !_rules[rulepredicate_1]() {
									goto l218
								}
								if !_rules[rule_]() {
									goto l218
								}
								if buffer[position] != rune(']') {
									goto l218
								}
								position++
								goto l217
							l218:
								position, tokenIndex, depth = position217, tokenIndex217, depth217
								{
									add(ruleAction34, position)
								}
							}
						l217:
							goto l216

							position, tokenIndex, depth = position215, tokenIndex215, depth215
						}
					l216:
						{
							add(ruleAction35, position)
						}
						depth--
						add(ruleexpression_metric, position212)
					}
					goto l205
				l211:
					position, tokenIndex, depth = position205, tokenIndex205, depth205
					if !_rules[rule_]() {
						goto l221
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l221
					}
					if !_rules[ruleexpression_start]() {
						goto l221
					}
					if !_rules[rule_]() {
						goto l221
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l221
					}
					goto l205
				l221:
					position, tokenIndex, depth = position205, tokenIndex205, depth205
					if !_rules[rule_]() {
						goto l222
					}
					{
						position223 := position
						depth++
						{
							position224 := position
							depth++
							if !_rules[ruleNUMBER]() {
								goto l222
							}
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l222
							}
							position++
						l225:
							{
								position226, tokenIndex226, depth226 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l226
								}
								position++
								goto l225
							l226:
								position, tokenIndex, depth = position226, tokenIndex226, depth226
							}
							if !_rules[ruleKEY]() {
								goto l222
							}
							depth--
							add(ruleDURATION, position224)
						}
						depth--
						add(rulePegText, position223)
					}
					{
						add(ruleAction27, position)
					}
					goto l205
				l222:
					position, tokenIndex, depth = position205, tokenIndex205, depth205
					if !_rules[rule_]() {
						goto l228
					}
					{
						position229 := position
						depth++
						if !_rules[ruleNUMBER]() {
							goto l228
						}
						depth--
						add(rulePegText, position229)
					}
					{
						add(ruleAction28, position)
					}
					goto l205
				l228:
					position, tokenIndex, depth = position205, tokenIndex205, depth205
					if !_rules[rule_]() {
						goto l203
					}
					if !_rules[ruleSTRING]() {
						goto l203
					}
					{
						add(ruleAction29, position)
					}
				}
			l205:
				depth--
				add(ruleexpression_atom, position204)
			}
			return true
		l203:
			position, tokenIndex, depth = position203, tokenIndex203, depth203
			return false
		},
		/* 16 optionalGroupBy <- <(Action30 (groupByClause / collapseByClause)?)> */
		func() bool {
			{
				position233 := position
				depth++
				{
					add(ruleAction30, position)
				}
				{
					position235, tokenIndex235, depth235 := position, tokenIndex, depth
					{
						position237, tokenIndex237, depth237 := position, tokenIndex, depth
						{
							position239 := position
							depth++
							if !_rules[rule_]() {
								goto l238
							}
							{
								position240, tokenIndex240, depth240 := position, tokenIndex, depth
								if buffer[position] != rune('g') {
									goto l241
								}
								position++
								goto l240
							l241:
								position, tokenIndex, depth = position240, tokenIndex240, depth240
								if buffer[position] != rune('G') {
									goto l238
								}
								position++
							}
						l240:
							{
								position242, tokenIndex242, depth242 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l243
								}
								position++
								goto l242
							l243:
								position, tokenIndex, depth = position242, tokenIndex242, depth242
								if buffer[position] != rune('R') {
									goto l238
								}
								position++
							}
						l242:
							{
								position244, tokenIndex244, depth244 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l245
								}
								position++
								goto l244
							l245:
								position, tokenIndex, depth = position244, tokenIndex244, depth244
								if buffer[position] != rune('O') {
									goto l238
								}
								position++
							}
						l244:
							{
								position246, tokenIndex246, depth246 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l247
								}
								position++
								goto l246
							l247:
								position, tokenIndex, depth = position246, tokenIndex246, depth246
								if buffer[position] != rune('U') {
									goto l238
								}
								position++
							}
						l246:
							{
								position248, tokenIndex248, depth248 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l249
								}
								position++
								goto l248
							l249:
								position, tokenIndex, depth = position248, tokenIndex248, depth248
								if buffer[position] != rune('P') {
									goto l238
								}
								position++
							}
						l248:
							if !_rules[ruleKEY]() {
								goto l238
							}
							if !_rules[rule_]() {
								goto l238
							}
							{
								position250, tokenIndex250, depth250 := position, tokenIndex, depth
								if buffer[position] != rune('b') {
									goto l251
								}
								position++
								goto l250
							l251:
								position, tokenIndex, depth = position250, tokenIndex250, depth250
								if buffer[position] != rune('B') {
									goto l238
								}
								position++
							}
						l250:
							{
								position252, tokenIndex252, depth252 := position, tokenIndex, depth
								if buffer[position] != rune('y') {
									goto l253
								}
								position++
								goto l252
							l253:
								position, tokenIndex, depth = position252, tokenIndex252, depth252
								if buffer[position] != rune('Y') {
									goto l238
								}
								position++
							}
						l252:
							if !_rules[ruleKEY]() {
								goto l238
							}
							if !_rules[rule_]() {
								goto l238
							}
							{
								position254 := position
								depth++
								if !_rules[ruleCOLUMN_NAME]() {
									goto l238
								}
								depth--
								add(rulePegText, position254)
							}
							{
								add(ruleAction36, position)
							}
						l256:
							{
								position257, tokenIndex257, depth257 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l257
								}
								if !_rules[ruleCOMMA]() {
									goto l257
								}
								if !_rules[rule_]() {
									goto l257
								}
								{
									position258 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l257
									}
									depth--
									add(rulePegText, position258)
								}
								{
									add(ruleAction37, position)
								}
								goto l256
							l257:
								position, tokenIndex, depth = position257, tokenIndex257, depth257
							}
							depth--
							add(rulegroupByClause, position239)
						}
						goto l237
					l238:
						position, tokenIndex, depth = position237, tokenIndex237, depth237
						{
							position260 := position
							depth++
							if !_rules[rule_]() {
								goto l235
							}
							{
								position261, tokenIndex261, depth261 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l262
								}
								position++
								goto l261
							l262:
								position, tokenIndex, depth = position261, tokenIndex261, depth261
								if buffer[position] != rune('C') {
									goto l235
								}
								position++
							}
						l261:
							{
								position263, tokenIndex263, depth263 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l264
								}
								position++
								goto l263
							l264:
								position, tokenIndex, depth = position263, tokenIndex263, depth263
								if buffer[position] != rune('O') {
									goto l235
								}
								position++
							}
						l263:
							{
								position265, tokenIndex265, depth265 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l266
								}
								position++
								goto l265
							l266:
								position, tokenIndex, depth = position265, tokenIndex265, depth265
								if buffer[position] != rune('L') {
									goto l235
								}
								position++
							}
						l265:
							{
								position267, tokenIndex267, depth267 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l268
								}
								position++
								goto l267
							l268:
								position, tokenIndex, depth = position267, tokenIndex267, depth267
								if buffer[position] != rune('L') {
									goto l235
								}
								position++
							}
						l267:
							{
								position269, tokenIndex269, depth269 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l270
								}
								position++
								goto l269
							l270:
								position, tokenIndex, depth = position269, tokenIndex269, depth269
								if buffer[position] != rune('A') {
									goto l235
								}
								position++
							}
						l269:
							{
								position271, tokenIndex271, depth271 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l272
								}
								position++
								goto l271
							l272:
								position, tokenIndex, depth = position271, tokenIndex271, depth271
								if buffer[position] != rune('P') {
									goto l235
								}
								position++
							}
						l271:
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
									goto l235
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
									goto l235
								}
								position++
							}
						l275:
							if !_rules[ruleKEY]() {
								goto l235
							}
							if !_rules[rule_]() {
								goto l235
							}
							{
								position277, tokenIndex277, depth277 := position, tokenIndex, depth
								if buffer[position] != rune('b') {
									goto l278
								}
								position++
								goto l277
							l278:
								position, tokenIndex, depth = position277, tokenIndex277, depth277
								if buffer[position] != rune('B') {
									goto l235
								}
								position++
							}
						l277:
							{
								position279, tokenIndex279, depth279 := position, tokenIndex, depth
								if buffer[position] != rune('y') {
									goto l280
								}
								position++
								goto l279
							l280:
								position, tokenIndex, depth = position279, tokenIndex279, depth279
								if buffer[position] != rune('Y') {
									goto l235
								}
								position++
							}
						l279:
							if !_rules[ruleKEY]() {
								goto l235
							}
							if !_rules[rule_]() {
								goto l235
							}
							{
								position281 := position
								depth++
								if !_rules[ruleCOLUMN_NAME]() {
									goto l235
								}
								depth--
								add(rulePegText, position281)
							}
							{
								add(ruleAction38, position)
							}
						l283:
							{
								position284, tokenIndex284, depth284 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l284
								}
								if !_rules[ruleCOMMA]() {
									goto l284
								}
								if !_rules[rule_]() {
									goto l284
								}
								{
									position285 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l284
									}
									depth--
									add(rulePegText, position285)
								}
								{
									add(ruleAction39, position)
								}
								goto l283
							l284:
								position, tokenIndex, depth = position284, tokenIndex284, depth284
							}
							depth--
							add(rulecollapseByClause, position260)
						}
					}
				l237:
					goto l236
				l235:
					position, tokenIndex, depth = position235, tokenIndex235, depth235
				}
			l236:
				depth--
				add(ruleoptionalGroupBy, position233)
			}
			return true
		},
		/* 17 expression_function <- <(_ <IDENTIFIER> Action31 _ PAREN_OPEN expressionList optionalGroupBy _ PAREN_CLOSE Action32)> */
		nil,
		/* 18 expression_metric <- <(_ <IDENTIFIER> Action33 ((_ '[' predicate_1 _ ']') / Action34)? Action35)> */
		nil,
		/* 19 groupByClause <- <(_ (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P')) KEY _ (('b' / 'B') ('y' / 'Y')) KEY _ <COLUMN_NAME> Action36 (_ COMMA _ <COLUMN_NAME> Action37)*)> */
		nil,
		/* 20 collapseByClause <- <(_ (('c' / 'C') ('o' / 'O') ('l' / 'L') ('l' / 'L') ('a' / 'A') ('p' / 'P') ('s' / 'S') ('e' / 'E')) KEY _ (('b' / 'B') ('y' / 'Y')) KEY _ <COLUMN_NAME> Action38 (_ COMMA _ <COLUMN_NAME> Action39)*)> */
		nil,
		/* 21 predicateClause <- <(_ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY _ predicate_1)> */
		nil,
		/* 22 predicate_1 <- <((predicate_2 _ OP_OR predicate_1 Action40) / predicate_2)> */
		func() bool {
			position292, tokenIndex292, depth292 := position, tokenIndex, depth
			{
				position293 := position
				depth++
				{
					position294, tokenIndex294, depth294 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l295
					}
					if !_rules[rule_]() {
						goto l295
					}
					{
						position296 := position
						depth++
						{
							position297, tokenIndex297, depth297 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l298
							}
							position++
							goto l297
						l298:
							position, tokenIndex, depth = position297, tokenIndex297, depth297
							if buffer[position] != rune('O') {
								goto l295
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
								goto l295
							}
							position++
						}
					l299:
						if !_rules[ruleKEY]() {
							goto l295
						}
						depth--
						add(ruleOP_OR, position296)
					}
					if !_rules[rulepredicate_1]() {
						goto l295
					}
					{
						add(ruleAction40, position)
					}
					goto l294
				l295:
					position, tokenIndex, depth = position294, tokenIndex294, depth294
					if !_rules[rulepredicate_2]() {
						goto l292
					}
				}
			l294:
				depth--
				add(rulepredicate_1, position293)
			}
			return true
		l292:
			position, tokenIndex, depth = position292, tokenIndex292, depth292
			return false
		},
		/* 23 predicate_2 <- <((predicate_3 _ OP_AND predicate_2 Action41) / predicate_3)> */
		func() bool {
			position302, tokenIndex302, depth302 := position, tokenIndex, depth
			{
				position303 := position
				depth++
				{
					position304, tokenIndex304, depth304 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l305
					}
					if !_rules[rule_]() {
						goto l305
					}
					{
						position306 := position
						depth++
						{
							position307, tokenIndex307, depth307 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l308
							}
							position++
							goto l307
						l308:
							position, tokenIndex, depth = position307, tokenIndex307, depth307
							if buffer[position] != rune('A') {
								goto l305
							}
							position++
						}
					l307:
						{
							position309, tokenIndex309, depth309 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l310
							}
							position++
							goto l309
						l310:
							position, tokenIndex, depth = position309, tokenIndex309, depth309
							if buffer[position] != rune('N') {
								goto l305
							}
							position++
						}
					l309:
						{
							position311, tokenIndex311, depth311 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l312
							}
							position++
							goto l311
						l312:
							position, tokenIndex, depth = position311, tokenIndex311, depth311
							if buffer[position] != rune('D') {
								goto l305
							}
							position++
						}
					l311:
						if !_rules[ruleKEY]() {
							goto l305
						}
						depth--
						add(ruleOP_AND, position306)
					}
					if !_rules[rulepredicate_2]() {
						goto l305
					}
					{
						add(ruleAction41, position)
					}
					goto l304
				l305:
					position, tokenIndex, depth = position304, tokenIndex304, depth304
					if !_rules[rulepredicate_3]() {
						goto l302
					}
				}
			l304:
				depth--
				add(rulepredicate_2, position303)
			}
			return true
		l302:
			position, tokenIndex, depth = position302, tokenIndex302, depth302
			return false
		},
		/* 24 predicate_3 <- <((_ OP_NOT predicate_3 Action42) / (_ PAREN_OPEN predicate_1 _ PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position314, tokenIndex314, depth314 := position, tokenIndex, depth
			{
				position315 := position
				depth++
				{
					position316, tokenIndex316, depth316 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l317
					}
					{
						position318 := position
						depth++
						{
							position319, tokenIndex319, depth319 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l320
							}
							position++
							goto l319
						l320:
							position, tokenIndex, depth = position319, tokenIndex319, depth319
							if buffer[position] != rune('N') {
								goto l317
							}
							position++
						}
					l319:
						{
							position321, tokenIndex321, depth321 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l322
							}
							position++
							goto l321
						l322:
							position, tokenIndex, depth = position321, tokenIndex321, depth321
							if buffer[position] != rune('O') {
								goto l317
							}
							position++
						}
					l321:
						{
							position323, tokenIndex323, depth323 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l324
							}
							position++
							goto l323
						l324:
							position, tokenIndex, depth = position323, tokenIndex323, depth323
							if buffer[position] != rune('T') {
								goto l317
							}
							position++
						}
					l323:
						if !_rules[ruleKEY]() {
							goto l317
						}
						depth--
						add(ruleOP_NOT, position318)
					}
					if !_rules[rulepredicate_3]() {
						goto l317
					}
					{
						add(ruleAction42, position)
					}
					goto l316
				l317:
					position, tokenIndex, depth = position316, tokenIndex316, depth316
					if !_rules[rule_]() {
						goto l326
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l326
					}
					if !_rules[rulepredicate_1]() {
						goto l326
					}
					if !_rules[rule_]() {
						goto l326
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l326
					}
					goto l316
				l326:
					position, tokenIndex, depth = position316, tokenIndex316, depth316
					{
						position327 := position
						depth++
						{
							position328, tokenIndex328, depth328 := position, tokenIndex, depth
							if !_rules[ruletagName]() {
								goto l329
							}
							if !_rules[rule_]() {
								goto l329
							}
							if buffer[position] != rune('=') {
								goto l329
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l329
							}
							{
								add(ruleAction43, position)
							}
							goto l328
						l329:
							position, tokenIndex, depth = position328, tokenIndex328, depth328
							if !_rules[ruletagName]() {
								goto l331
							}
							if !_rules[rule_]() {
								goto l331
							}
							if buffer[position] != rune('!') {
								goto l331
							}
							position++
							if buffer[position] != rune('=') {
								goto l331
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l331
							}
							{
								add(ruleAction44, position)
							}
							goto l328
						l331:
							position, tokenIndex, depth = position328, tokenIndex328, depth328
							if !_rules[ruletagName]() {
								goto l333
							}
							if !_rules[rule_]() {
								goto l333
							}
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
							if !_rules[ruleKEY]() {
								goto l333
							}
							if !_rules[ruleliteralString]() {
								goto l333
							}
							{
								add(ruleAction45, position)
							}
							goto l328
						l333:
							position, tokenIndex, depth = position328, tokenIndex328, depth328
							if !_rules[ruletagName]() {
								goto l314
							}
							if !_rules[rule_]() {
								goto l314
							}
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
									goto l314
								}
								position++
							}
						l345:
							{
								position347, tokenIndex347, depth347 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l348
								}
								position++
								goto l347
							l348:
								position, tokenIndex, depth = position347, tokenIndex347, depth347
								if buffer[position] != rune('N') {
									goto l314
								}
								position++
							}
						l347:
							if !_rules[ruleKEY]() {
								goto l314
							}
							{
								position349 := position
								depth++
								{
									add(ruleAction48, position)
								}
								if !_rules[rule_]() {
									goto l314
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l314
								}
								if !_rules[ruleliteralListString]() {
									goto l314
								}
							l351:
								{
									position352, tokenIndex352, depth352 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l352
									}
									if !_rules[ruleCOMMA]() {
										goto l352
									}
									if !_rules[ruleliteralListString]() {
										goto l352
									}
									goto l351
								l352:
									position, tokenIndex, depth = position352, tokenIndex352, depth352
								}
								if !_rules[rule_]() {
									goto l314
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l314
								}
								depth--
								add(ruleliteralList, position349)
							}
							{
								add(ruleAction46, position)
							}
						}
					l328:
						depth--
						add(ruletagMatcher, position327)
					}
				}
			l316:
				depth--
				add(rulepredicate_3, position315)
			}
			return true
		l314:
			position, tokenIndex, depth = position314, tokenIndex314, depth314
			return false
		},
		/* 25 tagMatcher <- <((tagName _ '=' literalString Action43) / (tagName _ ('!' '=') literalString Action44) / (tagName _ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H')) KEY literalString Action45) / (tagName _ (('i' / 'I') ('n' / 'N')) KEY literalList Action46))> */
		nil,
		/* 26 literalString <- <(_ STRING Action47)> */
		func() bool {
			position355, tokenIndex355, depth355 := position, tokenIndex, depth
			{
				position356 := position
				depth++
				if !_rules[rule_]() {
					goto l355
				}
				if !_rules[ruleSTRING]() {
					goto l355
				}
				{
					add(ruleAction47, position)
				}
				depth--
				add(ruleliteralString, position356)
			}
			return true
		l355:
			position, tokenIndex, depth = position355, tokenIndex355, depth355
			return false
		},
		/* 27 literalList <- <(Action48 _ PAREN_OPEN literalListString (_ COMMA literalListString)* _ PAREN_CLOSE)> */
		nil,
		/* 28 literalListString <- <(_ STRING Action49)> */
		func() bool {
			position359, tokenIndex359, depth359 := position, tokenIndex, depth
			{
				position360 := position
				depth++
				if !_rules[rule_]() {
					goto l359
				}
				if !_rules[ruleSTRING]() {
					goto l359
				}
				{
					add(ruleAction49, position)
				}
				depth--
				add(ruleliteralListString, position360)
			}
			return true
		l359:
			position, tokenIndex, depth = position359, tokenIndex359, depth359
			return false
		},
		/* 29 tagName <- <(_ <TAG_NAME> Action50)> */
		func() bool {
			position362, tokenIndex362, depth362 := position, tokenIndex, depth
			{
				position363 := position
				depth++
				if !_rules[rule_]() {
					goto l362
				}
				{
					position364 := position
					depth++
					{
						position365 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l362
						}
						depth--
						add(ruleTAG_NAME, position365)
					}
					depth--
					add(rulePegText, position364)
				}
				{
					add(ruleAction50, position)
				}
				depth--
				add(ruletagName, position363)
			}
			return true
		l362:
			position, tokenIndex, depth = position362, tokenIndex362, depth362
			return false
		},
		/* 30 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position367, tokenIndex367, depth367 := position, tokenIndex, depth
			{
				position368 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l367
				}
				depth--
				add(ruleCOLUMN_NAME, position368)
			}
			return true
		l367:
			position, tokenIndex, depth = position367, tokenIndex367, depth367
			return false
		},
		/* 31 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 32 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 33 IDENTIFIER <- <(('`' CHAR* '`') / (_ !(KEYWORD KEY) ID_SEGMENT ('.' ID_SEGMENT)*))> */
		func() bool {
			position371, tokenIndex371, depth371 := position, tokenIndex, depth
			{
				position372 := position
				depth++
				{
					position373, tokenIndex373, depth373 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l374
					}
					position++
				l375:
					{
						position376, tokenIndex376, depth376 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l376
						}
						goto l375
					l376:
						position, tokenIndex, depth = position376, tokenIndex376, depth376
					}
					if buffer[position] != rune('`') {
						goto l374
					}
					position++
					goto l373
				l374:
					position, tokenIndex, depth = position373, tokenIndex373, depth373
					if !_rules[rule_]() {
						goto l371
					}
					{
						position377, tokenIndex377, depth377 := position, tokenIndex, depth
						{
							position378 := position
							depth++
							{
								position379, tokenIndex379, depth379 := position, tokenIndex, depth
								{
									position381, tokenIndex381, depth381 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l382
									}
									position++
									goto l381
								l382:
									position, tokenIndex, depth = position381, tokenIndex381, depth381
									if buffer[position] != rune('A') {
										goto l380
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
										goto l380
									}
									position++
								}
							l383:
								{
									position385, tokenIndex385, depth385 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l386
									}
									position++
									goto l385
								l386:
									position, tokenIndex, depth = position385, tokenIndex385, depth385
									if buffer[position] != rune('L') {
										goto l380
									}
									position++
								}
							l385:
								goto l379
							l380:
								position, tokenIndex, depth = position379, tokenIndex379, depth379
								{
									position388, tokenIndex388, depth388 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l389
									}
									position++
									goto l388
								l389:
									position, tokenIndex, depth = position388, tokenIndex388, depth388
									if buffer[position] != rune('A') {
										goto l387
									}
									position++
								}
							l388:
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
										goto l387
									}
									position++
								}
							l390:
								{
									position392, tokenIndex392, depth392 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l393
									}
									position++
									goto l392
								l393:
									position, tokenIndex, depth = position392, tokenIndex392, depth392
									if buffer[position] != rune('D') {
										goto l387
									}
									position++
								}
							l392:
								goto l379
							l387:
								position, tokenIndex, depth = position379, tokenIndex379, depth379
								{
									position395, tokenIndex395, depth395 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l396
									}
									position++
									goto l395
								l396:
									position, tokenIndex, depth = position395, tokenIndex395, depth395
									if buffer[position] != rune('M') {
										goto l394
									}
									position++
								}
							l395:
								{
									position397, tokenIndex397, depth397 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l398
									}
									position++
									goto l397
								l398:
									position, tokenIndex, depth = position397, tokenIndex397, depth397
									if buffer[position] != rune('A') {
										goto l394
									}
									position++
								}
							l397:
								{
									position399, tokenIndex399, depth399 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l400
									}
									position++
									goto l399
								l400:
									position, tokenIndex, depth = position399, tokenIndex399, depth399
									if buffer[position] != rune('T') {
										goto l394
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
										goto l394
									}
									position++
								}
							l401:
								{
									position403, tokenIndex403, depth403 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l404
									}
									position++
									goto l403
								l404:
									position, tokenIndex, depth = position403, tokenIndex403, depth403
									if buffer[position] != rune('H') {
										goto l394
									}
									position++
								}
							l403:
								goto l379
							l394:
								position, tokenIndex, depth = position379, tokenIndex379, depth379
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
										goto l405
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
										goto l405
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
										goto l405
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
										goto l405
									}
									position++
								}
							l412:
								{
									position414, tokenIndex414, depth414 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l415
									}
									position++
									goto l414
								l415:
									position, tokenIndex, depth = position414, tokenIndex414, depth414
									if buffer[position] != rune('C') {
										goto l405
									}
									position++
								}
							l414:
								{
									position416, tokenIndex416, depth416 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l417
									}
									position++
									goto l416
								l417:
									position, tokenIndex, depth = position416, tokenIndex416, depth416
									if buffer[position] != rune('T') {
										goto l405
									}
									position++
								}
							l416:
								goto l379
							l405:
								position, tokenIndex, depth = position379, tokenIndex379, depth379
								{
									switch buffer[position] {
									case 'M', 'm':
										{
											position419, tokenIndex419, depth419 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l420
											}
											position++
											goto l419
										l420:
											position, tokenIndex, depth = position419, tokenIndex419, depth419
											if buffer[position] != rune('M') {
												goto l377
											}
											position++
										}
									l419:
										{
											position421, tokenIndex421, depth421 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l422
											}
											position++
											goto l421
										l422:
											position, tokenIndex, depth = position421, tokenIndex421, depth421
											if buffer[position] != rune('E') {
												goto l377
											}
											position++
										}
									l421:
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
												goto l377
											}
											position++
										}
									l423:
										{
											position425, tokenIndex425, depth425 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l426
											}
											position++
											goto l425
										l426:
											position, tokenIndex, depth = position425, tokenIndex425, depth425
											if buffer[position] != rune('R') {
												goto l377
											}
											position++
										}
									l425:
										{
											position427, tokenIndex427, depth427 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l428
											}
											position++
											goto l427
										l428:
											position, tokenIndex, depth = position427, tokenIndex427, depth427
											if buffer[position] != rune('I') {
												goto l377
											}
											position++
										}
									l427:
										{
											position429, tokenIndex429, depth429 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l430
											}
											position++
											goto l429
										l430:
											position, tokenIndex, depth = position429, tokenIndex429, depth429
											if buffer[position] != rune('C') {
												goto l377
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
												goto l377
											}
											position++
										}
									l431:
										break
									case 'W', 'w':
										{
											position433, tokenIndex433, depth433 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l434
											}
											position++
											goto l433
										l434:
											position, tokenIndex, depth = position433, tokenIndex433, depth433
											if buffer[position] != rune('W') {
												goto l377
											}
											position++
										}
									l433:
										{
											position435, tokenIndex435, depth435 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l436
											}
											position++
											goto l435
										l436:
											position, tokenIndex, depth = position435, tokenIndex435, depth435
											if buffer[position] != rune('H') {
												goto l377
											}
											position++
										}
									l435:
										{
											position437, tokenIndex437, depth437 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l438
											}
											position++
											goto l437
										l438:
											position, tokenIndex, depth = position437, tokenIndex437, depth437
											if buffer[position] != rune('E') {
												goto l377
											}
											position++
										}
									l437:
										{
											position439, tokenIndex439, depth439 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l440
											}
											position++
											goto l439
										l440:
											position, tokenIndex, depth = position439, tokenIndex439, depth439
											if buffer[position] != rune('R') {
												goto l377
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
												goto l377
											}
											position++
										}
									l441:
										break
									case 'O', 'o':
										{
											position443, tokenIndex443, depth443 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l444
											}
											position++
											goto l443
										l444:
											position, tokenIndex, depth = position443, tokenIndex443, depth443
											if buffer[position] != rune('O') {
												goto l377
											}
											position++
										}
									l443:
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
												goto l377
											}
											position++
										}
									l445:
										break
									case 'N', 'n':
										{
											position447, tokenIndex447, depth447 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l448
											}
											position++
											goto l447
										l448:
											position, tokenIndex, depth = position447, tokenIndex447, depth447
											if buffer[position] != rune('N') {
												goto l377
											}
											position++
										}
									l447:
										{
											position449, tokenIndex449, depth449 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l450
											}
											position++
											goto l449
										l450:
											position, tokenIndex, depth = position449, tokenIndex449, depth449
											if buffer[position] != rune('O') {
												goto l377
											}
											position++
										}
									l449:
										{
											position451, tokenIndex451, depth451 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l452
											}
											position++
											goto l451
										l452:
											position, tokenIndex, depth = position451, tokenIndex451, depth451
											if buffer[position] != rune('T') {
												goto l377
											}
											position++
										}
									l451:
										break
									case 'I', 'i':
										{
											position453, tokenIndex453, depth453 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l454
											}
											position++
											goto l453
										l454:
											position, tokenIndex, depth = position453, tokenIndex453, depth453
											if buffer[position] != rune('I') {
												goto l377
											}
											position++
										}
									l453:
										{
											position455, tokenIndex455, depth455 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l456
											}
											position++
											goto l455
										l456:
											position, tokenIndex, depth = position455, tokenIndex455, depth455
											if buffer[position] != rune('N') {
												goto l377
											}
											position++
										}
									l455:
										break
									case 'C', 'c':
										{
											position457, tokenIndex457, depth457 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l458
											}
											position++
											goto l457
										l458:
											position, tokenIndex, depth = position457, tokenIndex457, depth457
											if buffer[position] != rune('C') {
												goto l377
											}
											position++
										}
									l457:
										{
											position459, tokenIndex459, depth459 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l460
											}
											position++
											goto l459
										l460:
											position, tokenIndex, depth = position459, tokenIndex459, depth459
											if buffer[position] != rune('O') {
												goto l377
											}
											position++
										}
									l459:
										{
											position461, tokenIndex461, depth461 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l462
											}
											position++
											goto l461
										l462:
											position, tokenIndex, depth = position461, tokenIndex461, depth461
											if buffer[position] != rune('L') {
												goto l377
											}
											position++
										}
									l461:
										{
											position463, tokenIndex463, depth463 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l464
											}
											position++
											goto l463
										l464:
											position, tokenIndex, depth = position463, tokenIndex463, depth463
											if buffer[position] != rune('L') {
												goto l377
											}
											position++
										}
									l463:
										{
											position465, tokenIndex465, depth465 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l466
											}
											position++
											goto l465
										l466:
											position, tokenIndex, depth = position465, tokenIndex465, depth465
											if buffer[position] != rune('A') {
												goto l377
											}
											position++
										}
									l465:
										{
											position467, tokenIndex467, depth467 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l468
											}
											position++
											goto l467
										l468:
											position, tokenIndex, depth = position467, tokenIndex467, depth467
											if buffer[position] != rune('P') {
												goto l377
											}
											position++
										}
									l467:
										{
											position469, tokenIndex469, depth469 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l470
											}
											position++
											goto l469
										l470:
											position, tokenIndex, depth = position469, tokenIndex469, depth469
											if buffer[position] != rune('S') {
												goto l377
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
												goto l377
											}
											position++
										}
									l471:
										break
									case 'G', 'g':
										{
											position473, tokenIndex473, depth473 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l474
											}
											position++
											goto l473
										l474:
											position, tokenIndex, depth = position473, tokenIndex473, depth473
											if buffer[position] != rune('G') {
												goto l377
											}
											position++
										}
									l473:
										{
											position475, tokenIndex475, depth475 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l476
											}
											position++
											goto l475
										l476:
											position, tokenIndex, depth = position475, tokenIndex475, depth475
											if buffer[position] != rune('R') {
												goto l377
											}
											position++
										}
									l475:
										{
											position477, tokenIndex477, depth477 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l478
											}
											position++
											goto l477
										l478:
											position, tokenIndex, depth = position477, tokenIndex477, depth477
											if buffer[position] != rune('O') {
												goto l377
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
												goto l377
											}
											position++
										}
									l479:
										{
											position481, tokenIndex481, depth481 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l482
											}
											position++
											goto l481
										l482:
											position, tokenIndex, depth = position481, tokenIndex481, depth481
											if buffer[position] != rune('P') {
												goto l377
											}
											position++
										}
									l481:
										break
									case 'D', 'd':
										{
											position483, tokenIndex483, depth483 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l484
											}
											position++
											goto l483
										l484:
											position, tokenIndex, depth = position483, tokenIndex483, depth483
											if buffer[position] != rune('D') {
												goto l377
											}
											position++
										}
									l483:
										{
											position485, tokenIndex485, depth485 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l486
											}
											position++
											goto l485
										l486:
											position, tokenIndex, depth = position485, tokenIndex485, depth485
											if buffer[position] != rune('E') {
												goto l377
											}
											position++
										}
									l485:
										{
											position487, tokenIndex487, depth487 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l488
											}
											position++
											goto l487
										l488:
											position, tokenIndex, depth = position487, tokenIndex487, depth487
											if buffer[position] != rune('S') {
												goto l377
											}
											position++
										}
									l487:
										{
											position489, tokenIndex489, depth489 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l490
											}
											position++
											goto l489
										l490:
											position, tokenIndex, depth = position489, tokenIndex489, depth489
											if buffer[position] != rune('C') {
												goto l377
											}
											position++
										}
									l489:
										{
											position491, tokenIndex491, depth491 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l492
											}
											position++
											goto l491
										l492:
											position, tokenIndex, depth = position491, tokenIndex491, depth491
											if buffer[position] != rune('R') {
												goto l377
											}
											position++
										}
									l491:
										{
											position493, tokenIndex493, depth493 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l494
											}
											position++
											goto l493
										l494:
											position, tokenIndex, depth = position493, tokenIndex493, depth493
											if buffer[position] != rune('I') {
												goto l377
											}
											position++
										}
									l493:
										{
											position495, tokenIndex495, depth495 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l496
											}
											position++
											goto l495
										l496:
											position, tokenIndex, depth = position495, tokenIndex495, depth495
											if buffer[position] != rune('B') {
												goto l377
											}
											position++
										}
									l495:
										{
											position497, tokenIndex497, depth497 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l498
											}
											position++
											goto l497
										l498:
											position, tokenIndex, depth = position497, tokenIndex497, depth497
											if buffer[position] != rune('E') {
												goto l377
											}
											position++
										}
									l497:
										break
									case 'B', 'b':
										{
											position499, tokenIndex499, depth499 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l500
											}
											position++
											goto l499
										l500:
											position, tokenIndex, depth = position499, tokenIndex499, depth499
											if buffer[position] != rune('B') {
												goto l377
											}
											position++
										}
									l499:
										{
											position501, tokenIndex501, depth501 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l502
											}
											position++
											goto l501
										l502:
											position, tokenIndex, depth = position501, tokenIndex501, depth501
											if buffer[position] != rune('Y') {
												goto l377
											}
											position++
										}
									l501:
										break
									case 'A', 'a':
										{
											position503, tokenIndex503, depth503 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l504
											}
											position++
											goto l503
										l504:
											position, tokenIndex, depth = position503, tokenIndex503, depth503
											if buffer[position] != rune('A') {
												goto l377
											}
											position++
										}
									l503:
										{
											position505, tokenIndex505, depth505 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l506
											}
											position++
											goto l505
										l506:
											position, tokenIndex, depth = position505, tokenIndex505, depth505
											if buffer[position] != rune('S') {
												goto l377
											}
											position++
										}
									l505:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l377
										}
										break
									}
								}

							}
						l379:
							depth--
							add(ruleKEYWORD, position378)
						}
						if !_rules[ruleKEY]() {
							goto l377
						}
						goto l371
					l377:
						position, tokenIndex, depth = position377, tokenIndex377, depth377
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l371
					}
				l507:
					{
						position508, tokenIndex508, depth508 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l508
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l508
						}
						goto l507
					l508:
						position, tokenIndex, depth = position508, tokenIndex508, depth508
					}
				}
			l373:
				depth--
				add(ruleIDENTIFIER, position372)
			}
			return true
		l371:
			position, tokenIndex, depth = position371, tokenIndex371, depth371
			return false
		},
		/* 34 TIMESTAMP <- <((_ <(NUMBER ([a-z] / [A-Z])*)>) / (_ STRING) / (_ <(('n' / 'N') ('o' / 'O') ('w' / 'W'))>))> */
		nil,
		/* 35 ID_SEGMENT <- <(_ ID_START ID_CONT*)> */
		func() bool {
			position510, tokenIndex510, depth510 := position, tokenIndex, depth
			{
				position511 := position
				depth++
				if !_rules[rule_]() {
					goto l510
				}
				if !_rules[ruleID_START]() {
					goto l510
				}
			l512:
				{
					position513, tokenIndex513, depth513 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l513
					}
					goto l512
				l513:
					position, tokenIndex, depth = position513, tokenIndex513, depth513
				}
				depth--
				add(ruleID_SEGMENT, position511)
			}
			return true
		l510:
			position, tokenIndex, depth = position510, tokenIndex510, depth510
			return false
		},
		/* 36 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position514, tokenIndex514, depth514 := position, tokenIndex, depth
			{
				position515 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l514
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l514
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l514
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position515)
			}
			return true
		l514:
			position, tokenIndex, depth = position514, tokenIndex514, depth514
			return false
		},
		/* 37 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position517, tokenIndex517, depth517 := position, tokenIndex, depth
			{
				position518 := position
				depth++
				{
					position519, tokenIndex519, depth519 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l520
					}
					goto l519
				l520:
					position, tokenIndex, depth = position519, tokenIndex519, depth519
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l517
					}
					position++
				}
			l519:
				depth--
				add(ruleID_CONT, position518)
			}
			return true
		l517:
			position, tokenIndex, depth = position517, tokenIndex517, depth517
			return false
		},
		/* 38 PROPERTY_KEY <- <(((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> KEY _ (('b' / 'B') ('y' / 'Y')))) | (&('R' | 'r') <(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))>) | (&('T' | 't') <(('t' / 'T') ('o' / 'O'))>) | (&('F' | 'f') <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))>)) KEY)> */
		func() bool {
			position521, tokenIndex521, depth521 := position, tokenIndex, depth
			{
				position522 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position524 := position
							depth++
							{
								position525, tokenIndex525, depth525 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l526
								}
								position++
								goto l525
							l526:
								position, tokenIndex, depth = position525, tokenIndex525, depth525
								if buffer[position] != rune('S') {
									goto l521
								}
								position++
							}
						l525:
							{
								position527, tokenIndex527, depth527 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l528
								}
								position++
								goto l527
							l528:
								position, tokenIndex, depth = position527, tokenIndex527, depth527
								if buffer[position] != rune('A') {
									goto l521
								}
								position++
							}
						l527:
							{
								position529, tokenIndex529, depth529 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l530
								}
								position++
								goto l529
							l530:
								position, tokenIndex, depth = position529, tokenIndex529, depth529
								if buffer[position] != rune('M') {
									goto l521
								}
								position++
							}
						l529:
							{
								position531, tokenIndex531, depth531 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l532
								}
								position++
								goto l531
							l532:
								position, tokenIndex, depth = position531, tokenIndex531, depth531
								if buffer[position] != rune('P') {
									goto l521
								}
								position++
							}
						l531:
							{
								position533, tokenIndex533, depth533 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l534
								}
								position++
								goto l533
							l534:
								position, tokenIndex, depth = position533, tokenIndex533, depth533
								if buffer[position] != rune('L') {
									goto l521
								}
								position++
							}
						l533:
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
									goto l521
								}
								position++
							}
						l535:
							depth--
							add(rulePegText, position524)
						}
						if !_rules[ruleKEY]() {
							goto l521
						}
						if !_rules[rule_]() {
							goto l521
						}
						{
							position537, tokenIndex537, depth537 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l538
							}
							position++
							goto l537
						l538:
							position, tokenIndex, depth = position537, tokenIndex537, depth537
							if buffer[position] != rune('B') {
								goto l521
							}
							position++
						}
					l537:
						{
							position539, tokenIndex539, depth539 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l540
							}
							position++
							goto l539
						l540:
							position, tokenIndex, depth = position539, tokenIndex539, depth539
							if buffer[position] != rune('Y') {
								goto l521
							}
							position++
						}
					l539:
						break
					case 'R', 'r':
						{
							position541 := position
							depth++
							{
								position542, tokenIndex542, depth542 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l543
								}
								position++
								goto l542
							l543:
								position, tokenIndex, depth = position542, tokenIndex542, depth542
								if buffer[position] != rune('R') {
									goto l521
								}
								position++
							}
						l542:
							{
								position544, tokenIndex544, depth544 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l545
								}
								position++
								goto l544
							l545:
								position, tokenIndex, depth = position544, tokenIndex544, depth544
								if buffer[position] != rune('E') {
									goto l521
								}
								position++
							}
						l544:
							{
								position546, tokenIndex546, depth546 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l547
								}
								position++
								goto l546
							l547:
								position, tokenIndex, depth = position546, tokenIndex546, depth546
								if buffer[position] != rune('S') {
									goto l521
								}
								position++
							}
						l546:
							{
								position548, tokenIndex548, depth548 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l549
								}
								position++
								goto l548
							l549:
								position, tokenIndex, depth = position548, tokenIndex548, depth548
								if buffer[position] != rune('O') {
									goto l521
								}
								position++
							}
						l548:
							{
								position550, tokenIndex550, depth550 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l551
								}
								position++
								goto l550
							l551:
								position, tokenIndex, depth = position550, tokenIndex550, depth550
								if buffer[position] != rune('L') {
									goto l521
								}
								position++
							}
						l550:
							{
								position552, tokenIndex552, depth552 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l553
								}
								position++
								goto l552
							l553:
								position, tokenIndex, depth = position552, tokenIndex552, depth552
								if buffer[position] != rune('U') {
									goto l521
								}
								position++
							}
						l552:
							{
								position554, tokenIndex554, depth554 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l555
								}
								position++
								goto l554
							l555:
								position, tokenIndex, depth = position554, tokenIndex554, depth554
								if buffer[position] != rune('T') {
									goto l521
								}
								position++
							}
						l554:
							{
								position556, tokenIndex556, depth556 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l557
								}
								position++
								goto l556
							l557:
								position, tokenIndex, depth = position556, tokenIndex556, depth556
								if buffer[position] != rune('I') {
									goto l521
								}
								position++
							}
						l556:
							{
								position558, tokenIndex558, depth558 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l559
								}
								position++
								goto l558
							l559:
								position, tokenIndex, depth = position558, tokenIndex558, depth558
								if buffer[position] != rune('O') {
									goto l521
								}
								position++
							}
						l558:
							{
								position560, tokenIndex560, depth560 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l561
								}
								position++
								goto l560
							l561:
								position, tokenIndex, depth = position560, tokenIndex560, depth560
								if buffer[position] != rune('N') {
									goto l521
								}
								position++
							}
						l560:
							depth--
							add(rulePegText, position541)
						}
						break
					case 'T', 't':
						{
							position562 := position
							depth++
							{
								position563, tokenIndex563, depth563 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l564
								}
								position++
								goto l563
							l564:
								position, tokenIndex, depth = position563, tokenIndex563, depth563
								if buffer[position] != rune('T') {
									goto l521
								}
								position++
							}
						l563:
							{
								position565, tokenIndex565, depth565 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l566
								}
								position++
								goto l565
							l566:
								position, tokenIndex, depth = position565, tokenIndex565, depth565
								if buffer[position] != rune('O') {
									goto l521
								}
								position++
							}
						l565:
							depth--
							add(rulePegText, position562)
						}
						break
					default:
						{
							position567 := position
							depth++
							{
								position568, tokenIndex568, depth568 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l569
								}
								position++
								goto l568
							l569:
								position, tokenIndex, depth = position568, tokenIndex568, depth568
								if buffer[position] != rune('F') {
									goto l521
								}
								position++
							}
						l568:
							{
								position570, tokenIndex570, depth570 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l571
								}
								position++
								goto l570
							l571:
								position, tokenIndex, depth = position570, tokenIndex570, depth570
								if buffer[position] != rune('R') {
									goto l521
								}
								position++
							}
						l570:
							{
								position572, tokenIndex572, depth572 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l573
								}
								position++
								goto l572
							l573:
								position, tokenIndex, depth = position572, tokenIndex572, depth572
								if buffer[position] != rune('O') {
									goto l521
								}
								position++
							}
						l572:
							{
								position574, tokenIndex574, depth574 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l575
								}
								position++
								goto l574
							l575:
								position, tokenIndex, depth = position574, tokenIndex574, depth574
								if buffer[position] != rune('M') {
									goto l521
								}
								position++
							}
						l574:
							depth--
							add(rulePegText, position567)
						}
						break
					}
				}

				if !_rules[ruleKEY]() {
					goto l521
				}
				depth--
				add(rulePROPERTY_KEY, position522)
			}
			return true
		l521:
			position, tokenIndex, depth = position521, tokenIndex521, depth521
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
			position586, tokenIndex586, depth586 := position, tokenIndex, depth
			{
				position587 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l586
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position587)
			}
			return true
		l586:
			position, tokenIndex, depth = position586, tokenIndex586, depth586
			return false
		},
		/* 50 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position588, tokenIndex588, depth588 := position, tokenIndex, depth
			{
				position589 := position
				depth++
				if buffer[position] != rune('"') {
					goto l588
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position589)
			}
			return true
		l588:
			position, tokenIndex, depth = position588, tokenIndex588, depth588
			return false
		},
		/* 51 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> QUOTE_SINGLE) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> QUOTE_DOUBLE))> */
		func() bool {
			position590, tokenIndex590, depth590 := position, tokenIndex, depth
			{
				position591 := position
				depth++
				{
					position592, tokenIndex592, depth592 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l593
					}
					{
						position594 := position
						depth++
					l595:
						{
							position596, tokenIndex596, depth596 := position, tokenIndex, depth
							{
								position597, tokenIndex597, depth597 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l597
								}
								goto l596
							l597:
								position, tokenIndex, depth = position597, tokenIndex597, depth597
							}
							if !_rules[ruleCHAR]() {
								goto l596
							}
							goto l595
						l596:
							position, tokenIndex, depth = position596, tokenIndex596, depth596
						}
						depth--
						add(rulePegText, position594)
					}
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l593
					}
					goto l592
				l593:
					position, tokenIndex, depth = position592, tokenIndex592, depth592
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l590
					}
					{
						position598 := position
						depth++
					l599:
						{
							position600, tokenIndex600, depth600 := position, tokenIndex, depth
							{
								position601, tokenIndex601, depth601 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l601
								}
								goto l600
							l601:
								position, tokenIndex, depth = position601, tokenIndex601, depth601
							}
							if !_rules[ruleCHAR]() {
								goto l600
							}
							goto l599
						l600:
							position, tokenIndex, depth = position600, tokenIndex600, depth600
						}
						depth--
						add(rulePegText, position598)
					}
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l590
					}
				}
			l592:
				depth--
				add(ruleSTRING, position591)
			}
			return true
		l590:
			position, tokenIndex, depth = position590, tokenIndex590, depth590
			return false
		},
		/* 52 CHAR <- <(('\\' ((&('"') QUOTE_DOUBLE) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position602, tokenIndex602, depth602 := position, tokenIndex, depth
			{
				position603 := position
				depth++
				{
					position604, tokenIndex604, depth604 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l605
					}
					position++
					{
						switch buffer[position] {
						case '"':
							if !_rules[ruleQUOTE_DOUBLE]() {
								goto l605
							}
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l605
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l605
							}
							break
						}
					}

					goto l604
				l605:
					position, tokenIndex, depth = position604, tokenIndex604, depth604
					{
						position607, tokenIndex607, depth607 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l607
						}
						goto l602
					l607:
						position, tokenIndex, depth = position607, tokenIndex607, depth607
					}
					if !matchDot() {
						goto l602
					}
				}
			l604:
				depth--
				add(ruleCHAR, position603)
			}
			return true
		l602:
			position, tokenIndex, depth = position602, tokenIndex602, depth602
			return false
		},
		/* 53 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position608, tokenIndex608, depth608 := position, tokenIndex, depth
			{
				position609 := position
				depth++
				{
					position610, tokenIndex610, depth610 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l611
					}
					position++
					goto l610
				l611:
					position, tokenIndex, depth = position610, tokenIndex610, depth610
					if buffer[position] != rune('\\') {
						goto l608
					}
					position++
				}
			l610:
				depth--
				add(ruleESCAPE_CLASS, position609)
			}
			return true
		l608:
			position, tokenIndex, depth = position608, tokenIndex608, depth608
			return false
		},
		/* 54 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position612, tokenIndex612, depth612 := position, tokenIndex, depth
			{
				position613 := position
				depth++
				{
					position614 := position
					depth++
					{
						position615, tokenIndex615, depth615 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l615
						}
						position++
						goto l616
					l615:
						position, tokenIndex, depth = position615, tokenIndex615, depth615
					}
				l616:
					{
						position617 := position
						depth++
						{
							position618, tokenIndex618, depth618 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l619
							}
							position++
							goto l618
						l619:
							position, tokenIndex, depth = position618, tokenIndex618, depth618
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l612
							}
							position++
						l620:
							{
								position621, tokenIndex621, depth621 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l621
								}
								position++
								goto l620
							l621:
								position, tokenIndex, depth = position621, tokenIndex621, depth621
							}
						}
					l618:
						depth--
						add(ruleNUMBER_NATURAL, position617)
					}
					depth--
					add(ruleNUMBER_INTEGER, position614)
				}
				{
					position622, tokenIndex622, depth622 := position, tokenIndex, depth
					{
						position624 := position
						depth++
						if buffer[position] != rune('.') {
							goto l622
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l622
						}
						position++
					l625:
						{
							position626, tokenIndex626, depth626 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l626
							}
							position++
							goto l625
						l626:
							position, tokenIndex, depth = position626, tokenIndex626, depth626
						}
						depth--
						add(ruleNUMBER_FRACTION, position624)
					}
					goto l623
				l622:
					position, tokenIndex, depth = position622, tokenIndex622, depth622
				}
			l623:
				{
					position627, tokenIndex627, depth627 := position, tokenIndex, depth
					{
						position629 := position
						depth++
						{
							position630, tokenIndex630, depth630 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l631
							}
							position++
							goto l630
						l631:
							position, tokenIndex, depth = position630, tokenIndex630, depth630
							if buffer[position] != rune('E') {
								goto l627
							}
							position++
						}
					l630:
						{
							position632, tokenIndex632, depth632 := position, tokenIndex, depth
							{
								position634, tokenIndex634, depth634 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l635
								}
								position++
								goto l634
							l635:
								position, tokenIndex, depth = position634, tokenIndex634, depth634
								if buffer[position] != rune('-') {
									goto l632
								}
								position++
							}
						l634:
							goto l633
						l632:
							position, tokenIndex, depth = position632, tokenIndex632, depth632
						}
					l633:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l627
						}
						position++
					l636:
						{
							position637, tokenIndex637, depth637 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l637
							}
							position++
							goto l636
						l637:
							position, tokenIndex, depth = position637, tokenIndex637, depth637
						}
						depth--
						add(ruleNUMBER_EXP, position629)
					}
					goto l628
				l627:
					position, tokenIndex, depth = position627, tokenIndex627, depth627
				}
			l628:
				depth--
				add(ruleNUMBER, position613)
			}
			return true
		l612:
			position, tokenIndex, depth = position612, tokenIndex612, depth612
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
			position643, tokenIndex643, depth643 := position, tokenIndex, depth
			{
				position644 := position
				depth++
				if buffer[position] != rune('(') {
					goto l643
				}
				position++
				depth--
				add(rulePAREN_OPEN, position644)
			}
			return true
		l643:
			position, tokenIndex, depth = position643, tokenIndex643, depth643
			return false
		},
		/* 61 PAREN_CLOSE <- <')'> */
		func() bool {
			position645, tokenIndex645, depth645 := position, tokenIndex, depth
			{
				position646 := position
				depth++
				if buffer[position] != rune(')') {
					goto l645
				}
				position++
				depth--
				add(rulePAREN_CLOSE, position646)
			}
			return true
		l645:
			position, tokenIndex, depth = position645, tokenIndex645, depth645
			return false
		},
		/* 62 COMMA <- <','> */
		func() bool {
			position647, tokenIndex647, depth647 := position, tokenIndex, depth
			{
				position648 := position
				depth++
				if buffer[position] != rune(',') {
					goto l647
				}
				position++
				depth--
				add(ruleCOMMA, position648)
			}
			return true
		l647:
			position, tokenIndex, depth = position647, tokenIndex647, depth647
			return false
		},
		/* 63 _ <- <((&('/') COMMENT_BLOCK) | (&('-') COMMENT_TRAIL) | (&('\t' | '\n' | ' ') SPACE))*> */
		func() bool {
			{
				position650 := position
				depth++
			l651:
				{
					position652, tokenIndex652, depth652 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '/':
							{
								position654 := position
								depth++
								if buffer[position] != rune('/') {
									goto l652
								}
								position++
								if buffer[position] != rune('*') {
									goto l652
								}
								position++
							l655:
								{
									position656, tokenIndex656, depth656 := position, tokenIndex, depth
									{
										position657, tokenIndex657, depth657 := position, tokenIndex, depth
										if buffer[position] != rune('*') {
											goto l657
										}
										position++
										if buffer[position] != rune('/') {
											goto l657
										}
										position++
										goto l656
									l657:
										position, tokenIndex, depth = position657, tokenIndex657, depth657
									}
									if !matchDot() {
										goto l656
									}
									goto l655
								l656:
									position, tokenIndex, depth = position656, tokenIndex656, depth656
								}
								if buffer[position] != rune('*') {
									goto l652
								}
								position++
								if buffer[position] != rune('/') {
									goto l652
								}
								position++
								depth--
								add(ruleCOMMENT_BLOCK, position654)
							}
							break
						case '-':
							{
								position658 := position
								depth++
								if buffer[position] != rune('-') {
									goto l652
								}
								position++
								if buffer[position] != rune('-') {
									goto l652
								}
								position++
							l659:
								{
									position660, tokenIndex660, depth660 := position, tokenIndex, depth
									{
										position661, tokenIndex661, depth661 := position, tokenIndex, depth
										if buffer[position] != rune('\n') {
											goto l661
										}
										position++
										goto l660
									l661:
										position, tokenIndex, depth = position661, tokenIndex661, depth661
									}
									if !matchDot() {
										goto l660
									}
									goto l659
								l660:
									position, tokenIndex, depth = position660, tokenIndex660, depth660
								}
								depth--
								add(ruleCOMMENT_TRAIL, position658)
							}
							break
						default:
							{
								position662 := position
								depth++
								{
									switch buffer[position] {
									case '\t':
										if buffer[position] != rune('\t') {
											goto l652
										}
										position++
										break
									case '\n':
										if buffer[position] != rune('\n') {
											goto l652
										}
										position++
										break
									default:
										if buffer[position] != rune(' ') {
											goto l652
										}
										position++
										break
									}
								}

								depth--
								add(ruleSPACE, position662)
							}
							break
						}
					}

					goto l651
				l652:
					position, tokenIndex, depth = position652, tokenIndex652, depth652
				}
				depth--
				add(rule_, position650)
			}
			return true
		},
		/* 64 COMMENT_TRAIL <- <('-' '-' (!'\n' .)*)> */
		nil,
		/* 65 COMMENT_BLOCK <- <('/' '*' (!('*' '/') .)* ('*' '/'))> */
		nil,
		/* 66 KEY <- <!ID_CONT> */
		func() bool {
			position666, tokenIndex666, depth666 := position, tokenIndex, depth
			{
				position667 := position
				depth++
				{
					position668, tokenIndex668, depth668 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l668
					}
					goto l666
				l668:
					position, tokenIndex, depth = position668, tokenIndex668, depth668
				}
				depth--
				add(ruleKEY, position667)
			}
			return true
		l666:
			position, tokenIndex, depth = position666, tokenIndex666, depth666
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
		/* 77 Action7 <- <{ p.addIndexClause() }> */
		nil,
		/* 78 Action8 <- <{ p.addEvaluationContext() }> */
		nil,
		/* 79 Action9 <- <{ p.addPropertyKey(buffer[begin:end])   }> */
		nil,
		/* 80 Action10 <- <{ p.addPropertyValue(buffer[begin:end]) }> */
		nil,
		/* 81 Action11 <- <{ p.insertPropertyKeyValue() }> */
		nil,
		/* 82 Action12 <- <{ p.checkPropertyClause() }> */
		nil,
		/* 83 Action13 <- <{ p.addNullPredicate() }> */
		nil,
		/* 84 Action14 <- <{ p.addExpressionList() }> */
		nil,
		/* 85 Action15 <- <{ p.appendExpression() }> */
		nil,
		/* 86 Action16 <- <{ p.appendExpression() }> */
		nil,
		/* 87 Action17 <- <{ p.addOperatorLiteral("+") }> */
		nil,
		/* 88 Action18 <- <{ p.addOperatorLiteral("-") }> */
		nil,
		/* 89 Action19 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 90 Action20 <- <{ p.addOperatorLiteral("/") }> */
		nil,
		/* 91 Action21 <- <{ p.addOperatorLiteral("*") }> */
		nil,
		/* 92 Action22 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 93 Action23 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 94 Action24 <- <{p.addExpressionList()}> */
		nil,
		/* 95 Action25 <- <{
		   p.addExpressionList()
		   p.addGroupBy()
		 }> */
		nil,
		/* 96 Action26 <- <{
		   p.addPipeExpression()
		 }> */
		nil,
		/* 97 Action27 <- <{ p.addDurationNode(text) }> */
		nil,
		/* 98 Action28 <- <{ p.addNumberNode(buffer[begin:end]) }> */
		nil,
		/* 99 Action29 <- <{ p.addStringNode(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 100 Action30 <- <{ p.addGroupBy() }> */
		nil,
		/* 101 Action31 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 102 Action32 <- <{
		   p.addFunctionInvocation()
		 }> */
		nil,
		/* 103 Action33 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 104 Action34 <- <{ p.addNullPredicate() }> */
		nil,
		/* 105 Action35 <- <{
		   p.addMetricExpression()
		 }> */
		nil,
		/* 106 Action36 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 107 Action37 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 108 Action38 <- <{
		   p.appendCollapseBy(unescapeLiteral(text))
		 }> */
		nil,
		/* 109 Action39 <- <{p.appendCollapseBy(unescapeLiteral(text))}> */
		nil,
		/* 110 Action40 <- <{ p.addOrPredicate() }> */
		nil,
		/* 111 Action41 <- <{ p.addAndPredicate() }> */
		nil,
		/* 112 Action42 <- <{ p.addNotPredicate() }> */
		nil,
		/* 113 Action43 <- <{
		   p.addLiteralMatcher()
		 }> */
		nil,
		/* 114 Action44 <- <{
		   p.addLiteralMatcher()
		   p.addNotPredicate()
		 }> */
		nil,
		/* 115 Action45 <- <{
		   p.addRegexMatcher()
		 }> */
		nil,
		/* 116 Action46 <- <{
		   p.addListMatcher()
		 }> */
		nil,
		/* 117 Action47 <- <{
		  p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 118 Action48 <- <{ p.addLiteralList() }> */
		nil,
		/* 119 Action49 <- <{
		  p.appendLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 120 Action50 <- <{ p.addTagLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
	}
	p.rules = _rules
}
