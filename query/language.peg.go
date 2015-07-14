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
	ruleoptionalMatchesClause
	rulematchesClause
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
	"optionalMatchesClause",
	"matchesClause",
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
			p.addNullMatchesClause()
		case ruleAction3:
			p.addMatchesClause()
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
											{
												position85, tokenIndex85, depth85 := position, tokenIndex, depth
												if buffer[position] != rune('e') {
													goto l86
												}
												position++
												goto l85
											l86:
												position, tokenIndex, depth = position85, tokenIndex85, depth85
												if buffer[position] != rune('E') {
													goto l73
												}
												position++
											}
										l85:
											{
												position87, tokenIndex87, depth87 := position, tokenIndex, depth
												if buffer[position] != rune('s') {
													goto l88
												}
												position++
												goto l87
											l88:
												position, tokenIndex, depth = position87, tokenIndex87, depth87
												if buffer[position] != rune('S') {
													goto l73
												}
												position++
											}
										l87:
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
											add(rulematchesClause, position74)
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
									add(ruleoptionalMatchesClause, position71)
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
								position93 := position
								depth++
								if !_rules[rule_]() {
									goto l92
								}
								{
									position94, tokenIndex94, depth94 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l95
									}
									position++
									goto l94
								l95:
									position, tokenIndex, depth = position94, tokenIndex94, depth94
									if buffer[position] != rune('M') {
										goto l92
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
										goto l92
									}
									position++
								}
							l96:
								{
									position98, tokenIndex98, depth98 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l99
									}
									position++
									goto l98
								l99:
									position, tokenIndex, depth = position98, tokenIndex98, depth98
									if buffer[position] != rune('T') {
										goto l92
									}
									position++
								}
							l98:
								{
									position100, tokenIndex100, depth100 := position, tokenIndex, depth
									if buffer[position] != rune('r') {
										goto l101
									}
									position++
									goto l100
								l101:
									position, tokenIndex, depth = position100, tokenIndex100, depth100
									if buffer[position] != rune('R') {
										goto l92
									}
									position++
								}
							l100:
								{
									position102, tokenIndex102, depth102 := position, tokenIndex, depth
									if buffer[position] != rune('i') {
										goto l103
									}
									position++
									goto l102
								l103:
									position, tokenIndex, depth = position102, tokenIndex102, depth102
									if buffer[position] != rune('I') {
										goto l92
									}
									position++
								}
							l102:
								{
									position104, tokenIndex104, depth104 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l105
									}
									position++
									goto l104
								l105:
									position, tokenIndex, depth = position104, tokenIndex104, depth104
									if buffer[position] != rune('C') {
										goto l92
									}
									position++
								}
							l104:
								{
									position106, tokenIndex106, depth106 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l107
									}
									position++
									goto l106
								l107:
									position, tokenIndex, depth = position106, tokenIndex106, depth106
									if buffer[position] != rune('S') {
										goto l92
									}
									position++
								}
							l106:
								if !_rules[ruleKEY]() {
									goto l92
								}
								if !_rules[rule_]() {
									goto l92
								}
								{
									position108, tokenIndex108, depth108 := position, tokenIndex, depth
									if buffer[position] != rune('w') {
										goto l109
									}
									position++
									goto l108
								l109:
									position, tokenIndex, depth = position108, tokenIndex108, depth108
									if buffer[position] != rune('W') {
										goto l92
									}
									position++
								}
							l108:
								{
									position110, tokenIndex110, depth110 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l111
									}
									position++
									goto l110
								l111:
									position, tokenIndex, depth = position110, tokenIndex110, depth110
									if buffer[position] != rune('H') {
										goto l92
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
										goto l92
									}
									position++
								}
							l112:
								{
									position114, tokenIndex114, depth114 := position, tokenIndex, depth
									if buffer[position] != rune('r') {
										goto l115
									}
									position++
									goto l114
								l115:
									position, tokenIndex, depth = position114, tokenIndex114, depth114
									if buffer[position] != rune('R') {
										goto l92
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
										goto l92
									}
									position++
								}
							l116:
								if !_rules[ruleKEY]() {
									goto l92
								}
								if !_rules[ruletagName]() {
									goto l92
								}
								if !_rules[rule_]() {
									goto l92
								}
								if buffer[position] != rune('=') {
									goto l92
								}
								position++
								if !_rules[ruleliteralString]() {
									goto l92
								}
								{
									add(ruleAction4, position)
								}
								depth--
								add(ruledescribeMetrics, position93)
							}
							goto l62
						l92:
							position, tokenIndex, depth = position62, tokenIndex62, depth62
							{
								position119 := position
								depth++
								if !_rules[rule_]() {
									goto l0
								}
								{
									position120 := position
									depth++
									{
										position121 := position
										depth++
										if !_rules[ruleIDENTIFIER]() {
											goto l0
										}
										depth--
										add(ruleMETRIC_NAME, position121)
									}
									depth--
									add(rulePegText, position120)
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
								add(ruledescribeSingleStmt, position119)
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
					position124, tokenIndex124, depth124 := position, tokenIndex, depth
					if !matchDot() {
						goto l124
					}
					goto l0
				l124:
					position, tokenIndex, depth = position124, tokenIndex124, depth124
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
		/* 3 describeAllStmt <- <(_ (('a' / 'A') ('l' / 'L') ('l' / 'L')) KEY optionalMatchesClause Action1)> */
		nil,
		/* 4 optionalMatchesClause <- <(matchesClause / Action2)> */
		nil,
		/* 5 matchesClause <- <(_ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) KEY literalString Action3)> */
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
				position134 := position
				depth++
				{
					position135, tokenIndex135, depth135 := position, tokenIndex, depth
					{
						position137 := position
						depth++
						if !_rules[rule_]() {
							goto l136
						}
						{
							position138, tokenIndex138, depth138 := position, tokenIndex, depth
							if buffer[position] != rune('w') {
								goto l139
							}
							position++
							goto l138
						l139:
							position, tokenIndex, depth = position138, tokenIndex138, depth138
							if buffer[position] != rune('W') {
								goto l136
							}
							position++
						}
					l138:
						{
							position140, tokenIndex140, depth140 := position, tokenIndex, depth
							if buffer[position] != rune('h') {
								goto l141
							}
							position++
							goto l140
						l141:
							position, tokenIndex, depth = position140, tokenIndex140, depth140
							if buffer[position] != rune('H') {
								goto l136
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
								goto l136
							}
							position++
						}
					l142:
						{
							position144, tokenIndex144, depth144 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l145
							}
							position++
							goto l144
						l145:
							position, tokenIndex, depth = position144, tokenIndex144, depth144
							if buffer[position] != rune('R') {
								goto l136
							}
							position++
						}
					l144:
						{
							position146, tokenIndex146, depth146 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l147
							}
							position++
							goto l146
						l147:
							position, tokenIndex, depth = position146, tokenIndex146, depth146
							if buffer[position] != rune('E') {
								goto l136
							}
							position++
						}
					l146:
						if !_rules[ruleKEY]() {
							goto l136
						}
						if !_rules[rule_]() {
							goto l136
						}
						if !_rules[rulepredicate_1]() {
							goto l136
						}
						depth--
						add(rulepredicateClause, position137)
					}
					goto l135
				l136:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
					{
						add(ruleAction12, position)
					}
				}
			l135:
				depth--
				add(ruleoptionalPredicateClause, position134)
			}
			return true
		},
		/* 10 expressionList <- <(Action13 expression_start Action14 (_ COMMA expression_start Action15)*)> */
		func() bool {
			position149, tokenIndex149, depth149 := position, tokenIndex, depth
			{
				position150 := position
				depth++
				{
					add(ruleAction13, position)
				}
				if !_rules[ruleexpression_start]() {
					goto l149
				}
				{
					add(ruleAction14, position)
				}
			l153:
				{
					position154, tokenIndex154, depth154 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l154
					}
					if !_rules[ruleCOMMA]() {
						goto l154
					}
					if !_rules[ruleexpression_start]() {
						goto l154
					}
					{
						add(ruleAction15, position)
					}
					goto l153
				l154:
					position, tokenIndex, depth = position154, tokenIndex154, depth154
				}
				depth--
				add(ruleexpressionList, position150)
			}
			return true
		l149:
			position, tokenIndex, depth = position149, tokenIndex149, depth149
			return false
		},
		/* 11 expression_start <- <(expression_sum add_pipe)> */
		func() bool {
			position156, tokenIndex156, depth156 := position, tokenIndex, depth
			{
				position157 := position
				depth++
				{
					position158 := position
					depth++
					if !_rules[ruleexpression_product]() {
						goto l156
					}
				l159:
					{
						position160, tokenIndex160, depth160 := position, tokenIndex, depth
						if !_rules[ruleadd_pipe]() {
							goto l160
						}
						{
							position161, tokenIndex161, depth161 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l162
							}
							{
								position163 := position
								depth++
								if buffer[position] != rune('+') {
									goto l162
								}
								position++
								depth--
								add(ruleOP_ADD, position163)
							}
							{
								add(ruleAction16, position)
							}
							goto l161
						l162:
							position, tokenIndex, depth = position161, tokenIndex161, depth161
							if !_rules[rule_]() {
								goto l160
							}
							{
								position165 := position
								depth++
								if buffer[position] != rune('-') {
									goto l160
								}
								position++
								depth--
								add(ruleOP_SUB, position165)
							}
							{
								add(ruleAction17, position)
							}
						}
					l161:
						if !_rules[ruleexpression_product]() {
							goto l160
						}
						{
							add(ruleAction18, position)
						}
						goto l159
					l160:
						position, tokenIndex, depth = position160, tokenIndex160, depth160
					}
					depth--
					add(ruleexpression_sum, position158)
				}
				if !_rules[ruleadd_pipe]() {
					goto l156
				}
				depth--
				add(ruleexpression_start, position157)
			}
			return true
		l156:
			position, tokenIndex, depth = position156, tokenIndex156, depth156
			return false
		},
		/* 12 expression_sum <- <(expression_product (add_pipe ((_ OP_ADD Action16) / (_ OP_SUB Action17)) expression_product Action18)*)> */
		nil,
		/* 13 expression_product <- <(expression_atom (add_pipe ((_ OP_DIV Action19) / (_ OP_MULT Action20)) expression_atom Action21)*)> */
		func() bool {
			position169, tokenIndex169, depth169 := position, tokenIndex, depth
			{
				position170 := position
				depth++
				if !_rules[ruleexpression_atom]() {
					goto l169
				}
			l171:
				{
					position172, tokenIndex172, depth172 := position, tokenIndex, depth
					if !_rules[ruleadd_pipe]() {
						goto l172
					}
					{
						position173, tokenIndex173, depth173 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l174
						}
						{
							position175 := position
							depth++
							if buffer[position] != rune('/') {
								goto l174
							}
							position++
							depth--
							add(ruleOP_DIV, position175)
						}
						{
							add(ruleAction19, position)
						}
						goto l173
					l174:
						position, tokenIndex, depth = position173, tokenIndex173, depth173
						if !_rules[rule_]() {
							goto l172
						}
						{
							position177 := position
							depth++
							if buffer[position] != rune('*') {
								goto l172
							}
							position++
							depth--
							add(ruleOP_MULT, position177)
						}
						{
							add(ruleAction20, position)
						}
					}
				l173:
					if !_rules[ruleexpression_atom]() {
						goto l172
					}
					{
						add(ruleAction21, position)
					}
					goto l171
				l172:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
				}
				depth--
				add(ruleexpression_product, position170)
			}
			return true
		l169:
			position, tokenIndex, depth = position169, tokenIndex169, depth169
			return false
		},
		/* 14 add_pipe <- <(_ OP_PIPE _ <IDENTIFIER> Action22 ((_ PAREN_OPEN (expressionList / Action23) optionalGroupBy _ PAREN_CLOSE) / Action24) Action25)*> */
		func() bool {
			{
				position181 := position
				depth++
			l182:
				{
					position183, tokenIndex183, depth183 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l183
					}
					{
						position184 := position
						depth++
						if buffer[position] != rune('|') {
							goto l183
						}
						position++
						depth--
						add(ruleOP_PIPE, position184)
					}
					if !_rules[rule_]() {
						goto l183
					}
					{
						position185 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l183
						}
						depth--
						add(rulePegText, position185)
					}
					{
						add(ruleAction22, position)
					}
					{
						position187, tokenIndex187, depth187 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l188
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l188
						}
						{
							position189, tokenIndex189, depth189 := position, tokenIndex, depth
							if !_rules[ruleexpressionList]() {
								goto l190
							}
							goto l189
						l190:
							position, tokenIndex, depth = position189, tokenIndex189, depth189
							{
								add(ruleAction23, position)
							}
						}
					l189:
						if !_rules[ruleoptionalGroupBy]() {
							goto l188
						}
						if !_rules[rule_]() {
							goto l188
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l188
						}
						goto l187
					l188:
						position, tokenIndex, depth = position187, tokenIndex187, depth187
						{
							add(ruleAction24, position)
						}
					}
				l187:
					{
						add(ruleAction25, position)
					}
					goto l182
				l183:
					position, tokenIndex, depth = position183, tokenIndex183, depth183
				}
				depth--
				add(ruleadd_pipe, position181)
			}
			return true
		},
		/* 15 expression_atom <- <(expression_function / expression_metric / (_ PAREN_OPEN expression_start _ PAREN_CLOSE) / (_ <DURATION> Action26) / (_ <NUMBER> Action27) / (_ STRING Action28))> */
		func() bool {
			position194, tokenIndex194, depth194 := position, tokenIndex, depth
			{
				position195 := position
				depth++
				{
					position196, tokenIndex196, depth196 := position, tokenIndex, depth
					{
						position198 := position
						depth++
						if !_rules[rule_]() {
							goto l197
						}
						{
							position199 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l197
							}
							depth--
							add(rulePegText, position199)
						}
						{
							add(ruleAction30, position)
						}
						if !_rules[rule_]() {
							goto l197
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l197
						}
						if !_rules[ruleexpressionList]() {
							goto l197
						}
						if !_rules[ruleoptionalGroupBy]() {
							goto l197
						}
						if !_rules[rule_]() {
							goto l197
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l197
						}
						{
							add(ruleAction31, position)
						}
						depth--
						add(ruleexpression_function, position198)
					}
					goto l196
				l197:
					position, tokenIndex, depth = position196, tokenIndex196, depth196
					{
						position203 := position
						depth++
						if !_rules[rule_]() {
							goto l202
						}
						{
							position204 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l202
							}
							depth--
							add(rulePegText, position204)
						}
						{
							add(ruleAction32, position)
						}
						{
							position206, tokenIndex206, depth206 := position, tokenIndex, depth
							{
								position208, tokenIndex208, depth208 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l209
								}
								if buffer[position] != rune('[') {
									goto l209
								}
								position++
								if !_rules[rulepredicate_1]() {
									goto l209
								}
								if !_rules[rule_]() {
									goto l209
								}
								if buffer[position] != rune(']') {
									goto l209
								}
								position++
								goto l208
							l209:
								position, tokenIndex, depth = position208, tokenIndex208, depth208
								{
									add(ruleAction33, position)
								}
							}
						l208:
							goto l207

							position, tokenIndex, depth = position206, tokenIndex206, depth206
						}
					l207:
						{
							add(ruleAction34, position)
						}
						depth--
						add(ruleexpression_metric, position203)
					}
					goto l196
				l202:
					position, tokenIndex, depth = position196, tokenIndex196, depth196
					if !_rules[rule_]() {
						goto l212
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l212
					}
					if !_rules[ruleexpression_start]() {
						goto l212
					}
					if !_rules[rule_]() {
						goto l212
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l212
					}
					goto l196
				l212:
					position, tokenIndex, depth = position196, tokenIndex196, depth196
					if !_rules[rule_]() {
						goto l213
					}
					{
						position214 := position
						depth++
						{
							position215 := position
							depth++
							if !_rules[ruleNUMBER]() {
								goto l213
							}
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l213
							}
							position++
						l216:
							{
								position217, tokenIndex217, depth217 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l217
								}
								position++
								goto l216
							l217:
								position, tokenIndex, depth = position217, tokenIndex217, depth217
							}
							if !_rules[ruleKEY]() {
								goto l213
							}
							depth--
							add(ruleDURATION, position215)
						}
						depth--
						add(rulePegText, position214)
					}
					{
						add(ruleAction26, position)
					}
					goto l196
				l213:
					position, tokenIndex, depth = position196, tokenIndex196, depth196
					if !_rules[rule_]() {
						goto l219
					}
					{
						position220 := position
						depth++
						if !_rules[ruleNUMBER]() {
							goto l219
						}
						depth--
						add(rulePegText, position220)
					}
					{
						add(ruleAction27, position)
					}
					goto l196
				l219:
					position, tokenIndex, depth = position196, tokenIndex196, depth196
					if !_rules[rule_]() {
						goto l194
					}
					if !_rules[ruleSTRING]() {
						goto l194
					}
					{
						add(ruleAction28, position)
					}
				}
			l196:
				depth--
				add(ruleexpression_atom, position195)
			}
			return true
		l194:
			position, tokenIndex, depth = position194, tokenIndex194, depth194
			return false
		},
		/* 16 optionalGroupBy <- <(Action29 (groupByClause / collapseByClause)?)> */
		func() bool {
			{
				position224 := position
				depth++
				{
					add(ruleAction29, position)
				}
				{
					position226, tokenIndex226, depth226 := position, tokenIndex, depth
					{
						position228, tokenIndex228, depth228 := position, tokenIndex, depth
						{
							position230 := position
							depth++
							if !_rules[rule_]() {
								goto l229
							}
							{
								position231, tokenIndex231, depth231 := position, tokenIndex, depth
								if buffer[position] != rune('g') {
									goto l232
								}
								position++
								goto l231
							l232:
								position, tokenIndex, depth = position231, tokenIndex231, depth231
								if buffer[position] != rune('G') {
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
									goto l229
								}
								position++
							}
						l235:
							{
								position237, tokenIndex237, depth237 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l238
								}
								position++
								goto l237
							l238:
								position, tokenIndex, depth = position237, tokenIndex237, depth237
								if buffer[position] != rune('U') {
									goto l229
								}
								position++
							}
						l237:
							{
								position239, tokenIndex239, depth239 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l240
								}
								position++
								goto l239
							l240:
								position, tokenIndex, depth = position239, tokenIndex239, depth239
								if buffer[position] != rune('P') {
									goto l229
								}
								position++
							}
						l239:
							if !_rules[ruleKEY]() {
								goto l229
							}
							if !_rules[rule_]() {
								goto l229
							}
							{
								position241, tokenIndex241, depth241 := position, tokenIndex, depth
								if buffer[position] != rune('b') {
									goto l242
								}
								position++
								goto l241
							l242:
								position, tokenIndex, depth = position241, tokenIndex241, depth241
								if buffer[position] != rune('B') {
									goto l229
								}
								position++
							}
						l241:
							{
								position243, tokenIndex243, depth243 := position, tokenIndex, depth
								if buffer[position] != rune('y') {
									goto l244
								}
								position++
								goto l243
							l244:
								position, tokenIndex, depth = position243, tokenIndex243, depth243
								if buffer[position] != rune('Y') {
									goto l229
								}
								position++
							}
						l243:
							if !_rules[ruleKEY]() {
								goto l229
							}
							if !_rules[rule_]() {
								goto l229
							}
							{
								position245 := position
								depth++
								if !_rules[ruleCOLUMN_NAME]() {
									goto l229
								}
								depth--
								add(rulePegText, position245)
							}
							{
								add(ruleAction35, position)
							}
						l247:
							{
								position248, tokenIndex248, depth248 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l248
								}
								if !_rules[ruleCOMMA]() {
									goto l248
								}
								if !_rules[rule_]() {
									goto l248
								}
								{
									position249 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l248
									}
									depth--
									add(rulePegText, position249)
								}
								{
									add(ruleAction36, position)
								}
								goto l247
							l248:
								position, tokenIndex, depth = position248, tokenIndex248, depth248
							}
							depth--
							add(rulegroupByClause, position230)
						}
						goto l228
					l229:
						position, tokenIndex, depth = position228, tokenIndex228, depth228
						{
							position251 := position
							depth++
							if !_rules[rule_]() {
								goto l226
							}
							{
								position252, tokenIndex252, depth252 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l253
								}
								position++
								goto l252
							l253:
								position, tokenIndex, depth = position252, tokenIndex252, depth252
								if buffer[position] != rune('C') {
									goto l226
								}
								position++
							}
						l252:
							{
								position254, tokenIndex254, depth254 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l255
								}
								position++
								goto l254
							l255:
								position, tokenIndex, depth = position254, tokenIndex254, depth254
								if buffer[position] != rune('O') {
									goto l226
								}
								position++
							}
						l254:
							{
								position256, tokenIndex256, depth256 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l257
								}
								position++
								goto l256
							l257:
								position, tokenIndex, depth = position256, tokenIndex256, depth256
								if buffer[position] != rune('L') {
									goto l226
								}
								position++
							}
						l256:
							{
								position258, tokenIndex258, depth258 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l259
								}
								position++
								goto l258
							l259:
								position, tokenIndex, depth = position258, tokenIndex258, depth258
								if buffer[position] != rune('L') {
									goto l226
								}
								position++
							}
						l258:
							{
								position260, tokenIndex260, depth260 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l261
								}
								position++
								goto l260
							l261:
								position, tokenIndex, depth = position260, tokenIndex260, depth260
								if buffer[position] != rune('A') {
									goto l226
								}
								position++
							}
						l260:
							{
								position262, tokenIndex262, depth262 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l263
								}
								position++
								goto l262
							l263:
								position, tokenIndex, depth = position262, tokenIndex262, depth262
								if buffer[position] != rune('P') {
									goto l226
								}
								position++
							}
						l262:
							{
								position264, tokenIndex264, depth264 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l265
								}
								position++
								goto l264
							l265:
								position, tokenIndex, depth = position264, tokenIndex264, depth264
								if buffer[position] != rune('S') {
									goto l226
								}
								position++
							}
						l264:
							{
								position266, tokenIndex266, depth266 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l267
								}
								position++
								goto l266
							l267:
								position, tokenIndex, depth = position266, tokenIndex266, depth266
								if buffer[position] != rune('E') {
									goto l226
								}
								position++
							}
						l266:
							if !_rules[ruleKEY]() {
								goto l226
							}
							if !_rules[rule_]() {
								goto l226
							}
							{
								position268, tokenIndex268, depth268 := position, tokenIndex, depth
								if buffer[position] != rune('b') {
									goto l269
								}
								position++
								goto l268
							l269:
								position, tokenIndex, depth = position268, tokenIndex268, depth268
								if buffer[position] != rune('B') {
									goto l226
								}
								position++
							}
						l268:
							{
								position270, tokenIndex270, depth270 := position, tokenIndex, depth
								if buffer[position] != rune('y') {
									goto l271
								}
								position++
								goto l270
							l271:
								position, tokenIndex, depth = position270, tokenIndex270, depth270
								if buffer[position] != rune('Y') {
									goto l226
								}
								position++
							}
						l270:
							if !_rules[ruleKEY]() {
								goto l226
							}
							if !_rules[rule_]() {
								goto l226
							}
							{
								position272 := position
								depth++
								if !_rules[ruleCOLUMN_NAME]() {
									goto l226
								}
								depth--
								add(rulePegText, position272)
							}
							{
								add(ruleAction37, position)
							}
						l274:
							{
								position275, tokenIndex275, depth275 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l275
								}
								if !_rules[ruleCOMMA]() {
									goto l275
								}
								if !_rules[rule_]() {
									goto l275
								}
								{
									position276 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l275
									}
									depth--
									add(rulePegText, position276)
								}
								{
									add(ruleAction38, position)
								}
								goto l274
							l275:
								position, tokenIndex, depth = position275, tokenIndex275, depth275
							}
							depth--
							add(rulecollapseByClause, position251)
						}
					}
				l228:
					goto l227
				l226:
					position, tokenIndex, depth = position226, tokenIndex226, depth226
				}
			l227:
				depth--
				add(ruleoptionalGroupBy, position224)
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
			position283, tokenIndex283, depth283 := position, tokenIndex, depth
			{
				position284 := position
				depth++
				{
					position285, tokenIndex285, depth285 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l286
					}
					if !_rules[rule_]() {
						goto l286
					}
					{
						position287 := position
						depth++
						{
							position288, tokenIndex288, depth288 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l289
							}
							position++
							goto l288
						l289:
							position, tokenIndex, depth = position288, tokenIndex288, depth288
							if buffer[position] != rune('O') {
								goto l286
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
								goto l286
							}
							position++
						}
					l290:
						if !_rules[ruleKEY]() {
							goto l286
						}
						depth--
						add(ruleOP_OR, position287)
					}
					if !_rules[rulepredicate_1]() {
						goto l286
					}
					{
						add(ruleAction39, position)
					}
					goto l285
				l286:
					position, tokenIndex, depth = position285, tokenIndex285, depth285
					if !_rules[rulepredicate_2]() {
						goto l283
					}
				}
			l285:
				depth--
				add(rulepredicate_1, position284)
			}
			return true
		l283:
			position, tokenIndex, depth = position283, tokenIndex283, depth283
			return false
		},
		/* 23 predicate_2 <- <((predicate_3 _ OP_AND predicate_2 Action40) / predicate_3)> */
		func() bool {
			position293, tokenIndex293, depth293 := position, tokenIndex, depth
			{
				position294 := position
				depth++
				{
					position295, tokenIndex295, depth295 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l296
					}
					if !_rules[rule_]() {
						goto l296
					}
					{
						position297 := position
						depth++
						{
							position298, tokenIndex298, depth298 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l299
							}
							position++
							goto l298
						l299:
							position, tokenIndex, depth = position298, tokenIndex298, depth298
							if buffer[position] != rune('A') {
								goto l296
							}
							position++
						}
					l298:
						{
							position300, tokenIndex300, depth300 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l301
							}
							position++
							goto l300
						l301:
							position, tokenIndex, depth = position300, tokenIndex300, depth300
							if buffer[position] != rune('N') {
								goto l296
							}
							position++
						}
					l300:
						{
							position302, tokenIndex302, depth302 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l303
							}
							position++
							goto l302
						l303:
							position, tokenIndex, depth = position302, tokenIndex302, depth302
							if buffer[position] != rune('D') {
								goto l296
							}
							position++
						}
					l302:
						if !_rules[ruleKEY]() {
							goto l296
						}
						depth--
						add(ruleOP_AND, position297)
					}
					if !_rules[rulepredicate_2]() {
						goto l296
					}
					{
						add(ruleAction40, position)
					}
					goto l295
				l296:
					position, tokenIndex, depth = position295, tokenIndex295, depth295
					if !_rules[rulepredicate_3]() {
						goto l293
					}
				}
			l295:
				depth--
				add(rulepredicate_2, position294)
			}
			return true
		l293:
			position, tokenIndex, depth = position293, tokenIndex293, depth293
			return false
		},
		/* 24 predicate_3 <- <((_ OP_NOT predicate_3 Action41) / (_ PAREN_OPEN predicate_1 _ PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position305, tokenIndex305, depth305 := position, tokenIndex, depth
			{
				position306 := position
				depth++
				{
					position307, tokenIndex307, depth307 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l308
					}
					{
						position309 := position
						depth++
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
								goto l308
							}
							position++
						}
					l310:
						{
							position312, tokenIndex312, depth312 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l313
							}
							position++
							goto l312
						l313:
							position, tokenIndex, depth = position312, tokenIndex312, depth312
							if buffer[position] != rune('O') {
								goto l308
							}
							position++
						}
					l312:
						{
							position314, tokenIndex314, depth314 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l315
							}
							position++
							goto l314
						l315:
							position, tokenIndex, depth = position314, tokenIndex314, depth314
							if buffer[position] != rune('T') {
								goto l308
							}
							position++
						}
					l314:
						if !_rules[ruleKEY]() {
							goto l308
						}
						depth--
						add(ruleOP_NOT, position309)
					}
					if !_rules[rulepredicate_3]() {
						goto l308
					}
					{
						add(ruleAction41, position)
					}
					goto l307
				l308:
					position, tokenIndex, depth = position307, tokenIndex307, depth307
					if !_rules[rule_]() {
						goto l317
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l317
					}
					if !_rules[rulepredicate_1]() {
						goto l317
					}
					if !_rules[rule_]() {
						goto l317
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l317
					}
					goto l307
				l317:
					position, tokenIndex, depth = position307, tokenIndex307, depth307
					{
						position318 := position
						depth++
						{
							position319, tokenIndex319, depth319 := position, tokenIndex, depth
							if !_rules[ruletagName]() {
								goto l320
							}
							if !_rules[rule_]() {
								goto l320
							}
							if buffer[position] != rune('=') {
								goto l320
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l320
							}
							{
								add(ruleAction42, position)
							}
							goto l319
						l320:
							position, tokenIndex, depth = position319, tokenIndex319, depth319
							if !_rules[ruletagName]() {
								goto l322
							}
							if !_rules[rule_]() {
								goto l322
							}
							if buffer[position] != rune('!') {
								goto l322
							}
							position++
							if buffer[position] != rune('=') {
								goto l322
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l322
							}
							{
								add(ruleAction43, position)
							}
							goto l319
						l322:
							position, tokenIndex, depth = position319, tokenIndex319, depth319
							if !_rules[ruletagName]() {
								goto l324
							}
							if !_rules[rule_]() {
								goto l324
							}
							{
								position325, tokenIndex325, depth325 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l326
								}
								position++
								goto l325
							l326:
								position, tokenIndex, depth = position325, tokenIndex325, depth325
								if buffer[position] != rune('M') {
									goto l324
								}
								position++
							}
						l325:
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
									goto l324
								}
								position++
							}
						l327:
							{
								position329, tokenIndex329, depth329 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l330
								}
								position++
								goto l329
							l330:
								position, tokenIndex, depth = position329, tokenIndex329, depth329
								if buffer[position] != rune('T') {
									goto l324
								}
								position++
							}
						l329:
							{
								position331, tokenIndex331, depth331 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l332
								}
								position++
								goto l331
							l332:
								position, tokenIndex, depth = position331, tokenIndex331, depth331
								if buffer[position] != rune('C') {
									goto l324
								}
								position++
							}
						l331:
							{
								position333, tokenIndex333, depth333 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l334
								}
								position++
								goto l333
							l334:
								position, tokenIndex, depth = position333, tokenIndex333, depth333
								if buffer[position] != rune('H') {
									goto l324
								}
								position++
							}
						l333:
							{
								position335, tokenIndex335, depth335 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l336
								}
								position++
								goto l335
							l336:
								position, tokenIndex, depth = position335, tokenIndex335, depth335
								if buffer[position] != rune('E') {
									goto l324
								}
								position++
							}
						l335:
							{
								position337, tokenIndex337, depth337 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l338
								}
								position++
								goto l337
							l338:
								position, tokenIndex, depth = position337, tokenIndex337, depth337
								if buffer[position] != rune('S') {
									goto l324
								}
								position++
							}
						l337:
							if !_rules[ruleKEY]() {
								goto l324
							}
							if !_rules[ruleliteralString]() {
								goto l324
							}
							{
								add(ruleAction44, position)
							}
							goto l319
						l324:
							position, tokenIndex, depth = position319, tokenIndex319, depth319
							if !_rules[ruletagName]() {
								goto l305
							}
							if !_rules[rule_]() {
								goto l305
							}
							{
								position340, tokenIndex340, depth340 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l341
								}
								position++
								goto l340
							l341:
								position, tokenIndex, depth = position340, tokenIndex340, depth340
								if buffer[position] != rune('I') {
									goto l305
								}
								position++
							}
						l340:
							{
								position342, tokenIndex342, depth342 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l343
								}
								position++
								goto l342
							l343:
								position, tokenIndex, depth = position342, tokenIndex342, depth342
								if buffer[position] != rune('N') {
									goto l305
								}
								position++
							}
						l342:
							if !_rules[ruleKEY]() {
								goto l305
							}
							{
								position344 := position
								depth++
								{
									add(ruleAction47, position)
								}
								if !_rules[rule_]() {
									goto l305
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l305
								}
								if !_rules[ruleliteralListString]() {
									goto l305
								}
							l346:
								{
									position347, tokenIndex347, depth347 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l347
									}
									if !_rules[ruleCOMMA]() {
										goto l347
									}
									if !_rules[ruleliteralListString]() {
										goto l347
									}
									goto l346
								l347:
									position, tokenIndex, depth = position347, tokenIndex347, depth347
								}
								if !_rules[rule_]() {
									goto l305
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l305
								}
								depth--
								add(ruleliteralList, position344)
							}
							{
								add(ruleAction45, position)
							}
						}
					l319:
						depth--
						add(ruletagMatcher, position318)
					}
				}
			l307:
				depth--
				add(rulepredicate_3, position306)
			}
			return true
		l305:
			position, tokenIndex, depth = position305, tokenIndex305, depth305
			return false
		},
		/* 25 tagMatcher <- <((tagName _ '=' literalString Action42) / (tagName _ ('!' '=') literalString Action43) / (tagName _ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) KEY literalString Action44) / (tagName _ (('i' / 'I') ('n' / 'N')) KEY literalList Action45))> */
		nil,
		/* 26 literalString <- <(_ STRING Action46)> */
		func() bool {
			position350, tokenIndex350, depth350 := position, tokenIndex, depth
			{
				position351 := position
				depth++
				if !_rules[rule_]() {
					goto l350
				}
				if !_rules[ruleSTRING]() {
					goto l350
				}
				{
					add(ruleAction46, position)
				}
				depth--
				add(ruleliteralString, position351)
			}
			return true
		l350:
			position, tokenIndex, depth = position350, tokenIndex350, depth350
			return false
		},
		/* 27 literalList <- <(Action47 _ PAREN_OPEN literalListString (_ COMMA literalListString)* _ PAREN_CLOSE)> */
		nil,
		/* 28 literalListString <- <(_ STRING Action48)> */
		func() bool {
			position354, tokenIndex354, depth354 := position, tokenIndex, depth
			{
				position355 := position
				depth++
				if !_rules[rule_]() {
					goto l354
				}
				if !_rules[ruleSTRING]() {
					goto l354
				}
				{
					add(ruleAction48, position)
				}
				depth--
				add(ruleliteralListString, position355)
			}
			return true
		l354:
			position, tokenIndex, depth = position354, tokenIndex354, depth354
			return false
		},
		/* 29 tagName <- <(_ <TAG_NAME> Action49)> */
		func() bool {
			position357, tokenIndex357, depth357 := position, tokenIndex, depth
			{
				position358 := position
				depth++
				if !_rules[rule_]() {
					goto l357
				}
				{
					position359 := position
					depth++
					{
						position360 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l357
						}
						depth--
						add(ruleTAG_NAME, position360)
					}
					depth--
					add(rulePegText, position359)
				}
				{
					add(ruleAction49, position)
				}
				depth--
				add(ruletagName, position358)
			}
			return true
		l357:
			position, tokenIndex, depth = position357, tokenIndex357, depth357
			return false
		},
		/* 30 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position362, tokenIndex362, depth362 := position, tokenIndex, depth
			{
				position363 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l362
				}
				depth--
				add(ruleCOLUMN_NAME, position363)
			}
			return true
		l362:
			position, tokenIndex, depth = position362, tokenIndex362, depth362
			return false
		},
		/* 31 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 32 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 33 IDENTIFIER <- <(('`' CHAR* '`') / (_ !(KEYWORD KEY) ID_SEGMENT ('.' ID_SEGMENT)*))> */
		func() bool {
			position366, tokenIndex366, depth366 := position, tokenIndex, depth
			{
				position367 := position
				depth++
				{
					position368, tokenIndex368, depth368 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l369
					}
					position++
				l370:
					{
						position371, tokenIndex371, depth371 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l371
						}
						goto l370
					l371:
						position, tokenIndex, depth = position371, tokenIndex371, depth371
					}
					if buffer[position] != rune('`') {
						goto l369
					}
					position++
					goto l368
				l369:
					position, tokenIndex, depth = position368, tokenIndex368, depth368
					if !_rules[rule_]() {
						goto l366
					}
					{
						position372, tokenIndex372, depth372 := position, tokenIndex, depth
						{
							position373 := position
							depth++
							{
								position374, tokenIndex374, depth374 := position, tokenIndex, depth
								{
									position376, tokenIndex376, depth376 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l377
									}
									position++
									goto l376
								l377:
									position, tokenIndex, depth = position376, tokenIndex376, depth376
									if buffer[position] != rune('A') {
										goto l375
									}
									position++
								}
							l376:
								{
									position378, tokenIndex378, depth378 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l379
									}
									position++
									goto l378
								l379:
									position, tokenIndex, depth = position378, tokenIndex378, depth378
									if buffer[position] != rune('L') {
										goto l375
									}
									position++
								}
							l378:
								{
									position380, tokenIndex380, depth380 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l381
									}
									position++
									goto l380
								l381:
									position, tokenIndex, depth = position380, tokenIndex380, depth380
									if buffer[position] != rune('L') {
										goto l375
									}
									position++
								}
							l380:
								goto l374
							l375:
								position, tokenIndex, depth = position374, tokenIndex374, depth374
								{
									position383, tokenIndex383, depth383 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l384
									}
									position++
									goto l383
								l384:
									position, tokenIndex, depth = position383, tokenIndex383, depth383
									if buffer[position] != rune('A') {
										goto l382
									}
									position++
								}
							l383:
								{
									position385, tokenIndex385, depth385 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l386
									}
									position++
									goto l385
								l386:
									position, tokenIndex, depth = position385, tokenIndex385, depth385
									if buffer[position] != rune('N') {
										goto l382
									}
									position++
								}
							l385:
								{
									position387, tokenIndex387, depth387 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l388
									}
									position++
									goto l387
								l388:
									position, tokenIndex, depth = position387, tokenIndex387, depth387
									if buffer[position] != rune('D') {
										goto l382
									}
									position++
								}
							l387:
								goto l374
							l382:
								position, tokenIndex, depth = position374, tokenIndex374, depth374
								{
									position390, tokenIndex390, depth390 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l391
									}
									position++
									goto l390
								l391:
									position, tokenIndex, depth = position390, tokenIndex390, depth390
									if buffer[position] != rune('M') {
										goto l389
									}
									position++
								}
							l390:
								{
									position392, tokenIndex392, depth392 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l393
									}
									position++
									goto l392
								l393:
									position, tokenIndex, depth = position392, tokenIndex392, depth392
									if buffer[position] != rune('A') {
										goto l389
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
										goto l389
									}
									position++
								}
							l394:
								{
									position396, tokenIndex396, depth396 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l397
									}
									position++
									goto l396
								l397:
									position, tokenIndex, depth = position396, tokenIndex396, depth396
									if buffer[position] != rune('C') {
										goto l389
									}
									position++
								}
							l396:
								{
									position398, tokenIndex398, depth398 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l399
									}
									position++
									goto l398
								l399:
									position, tokenIndex, depth = position398, tokenIndex398, depth398
									if buffer[position] != rune('H') {
										goto l389
									}
									position++
								}
							l398:
								{
									position400, tokenIndex400, depth400 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l401
									}
									position++
									goto l400
								l401:
									position, tokenIndex, depth = position400, tokenIndex400, depth400
									if buffer[position] != rune('E') {
										goto l389
									}
									position++
								}
							l400:
								{
									position402, tokenIndex402, depth402 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l403
									}
									position++
									goto l402
								l403:
									position, tokenIndex, depth = position402, tokenIndex402, depth402
									if buffer[position] != rune('S') {
										goto l389
									}
									position++
								}
							l402:
								goto l374
							l389:
								position, tokenIndex, depth = position374, tokenIndex374, depth374
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
										goto l404
									}
									position++
								}
							l405:
								{
									position407, tokenIndex407, depth407 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l408
									}
									position++
									goto l407
								l408:
									position, tokenIndex, depth = position407, tokenIndex407, depth407
									if buffer[position] != rune('E') {
										goto l404
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
										goto l404
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
										goto l404
									}
									position++
								}
							l411:
								{
									position413, tokenIndex413, depth413 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l414
									}
									position++
									goto l413
								l414:
									position, tokenIndex, depth = position413, tokenIndex413, depth413
									if buffer[position] != rune('C') {
										goto l404
									}
									position++
								}
							l413:
								{
									position415, tokenIndex415, depth415 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l416
									}
									position++
									goto l415
								l416:
									position, tokenIndex, depth = position415, tokenIndex415, depth415
									if buffer[position] != rune('T') {
										goto l404
									}
									position++
								}
							l415:
								goto l374
							l404:
								position, tokenIndex, depth = position374, tokenIndex374, depth374
								{
									switch buffer[position] {
									case 'M', 'm':
										{
											position418, tokenIndex418, depth418 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l419
											}
											position++
											goto l418
										l419:
											position, tokenIndex, depth = position418, tokenIndex418, depth418
											if buffer[position] != rune('M') {
												goto l372
											}
											position++
										}
									l418:
										{
											position420, tokenIndex420, depth420 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l421
											}
											position++
											goto l420
										l421:
											position, tokenIndex, depth = position420, tokenIndex420, depth420
											if buffer[position] != rune('E') {
												goto l372
											}
											position++
										}
									l420:
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
												goto l372
											}
											position++
										}
									l422:
										{
											position424, tokenIndex424, depth424 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l425
											}
											position++
											goto l424
										l425:
											position, tokenIndex, depth = position424, tokenIndex424, depth424
											if buffer[position] != rune('R') {
												goto l372
											}
											position++
										}
									l424:
										{
											position426, tokenIndex426, depth426 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l427
											}
											position++
											goto l426
										l427:
											position, tokenIndex, depth = position426, tokenIndex426, depth426
											if buffer[position] != rune('I') {
												goto l372
											}
											position++
										}
									l426:
										{
											position428, tokenIndex428, depth428 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l429
											}
											position++
											goto l428
										l429:
											position, tokenIndex, depth = position428, tokenIndex428, depth428
											if buffer[position] != rune('C') {
												goto l372
											}
											position++
										}
									l428:
										{
											position430, tokenIndex430, depth430 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l431
											}
											position++
											goto l430
										l431:
											position, tokenIndex, depth = position430, tokenIndex430, depth430
											if buffer[position] != rune('S') {
												goto l372
											}
											position++
										}
									l430:
										break
									case 'W', 'w':
										{
											position432, tokenIndex432, depth432 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l433
											}
											position++
											goto l432
										l433:
											position, tokenIndex, depth = position432, tokenIndex432, depth432
											if buffer[position] != rune('W') {
												goto l372
											}
											position++
										}
									l432:
										{
											position434, tokenIndex434, depth434 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l435
											}
											position++
											goto l434
										l435:
											position, tokenIndex, depth = position434, tokenIndex434, depth434
											if buffer[position] != rune('H') {
												goto l372
											}
											position++
										}
									l434:
										{
											position436, tokenIndex436, depth436 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l437
											}
											position++
											goto l436
										l437:
											position, tokenIndex, depth = position436, tokenIndex436, depth436
											if buffer[position] != rune('E') {
												goto l372
											}
											position++
										}
									l436:
										{
											position438, tokenIndex438, depth438 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l439
											}
											position++
											goto l438
										l439:
											position, tokenIndex, depth = position438, tokenIndex438, depth438
											if buffer[position] != rune('R') {
												goto l372
											}
											position++
										}
									l438:
										{
											position440, tokenIndex440, depth440 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l441
											}
											position++
											goto l440
										l441:
											position, tokenIndex, depth = position440, tokenIndex440, depth440
											if buffer[position] != rune('E') {
												goto l372
											}
											position++
										}
									l440:
										break
									case 'O', 'o':
										{
											position442, tokenIndex442, depth442 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l443
											}
											position++
											goto l442
										l443:
											position, tokenIndex, depth = position442, tokenIndex442, depth442
											if buffer[position] != rune('O') {
												goto l372
											}
											position++
										}
									l442:
										{
											position444, tokenIndex444, depth444 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l445
											}
											position++
											goto l444
										l445:
											position, tokenIndex, depth = position444, tokenIndex444, depth444
											if buffer[position] != rune('R') {
												goto l372
											}
											position++
										}
									l444:
										break
									case 'N', 'n':
										{
											position446, tokenIndex446, depth446 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l447
											}
											position++
											goto l446
										l447:
											position, tokenIndex, depth = position446, tokenIndex446, depth446
											if buffer[position] != rune('N') {
												goto l372
											}
											position++
										}
									l446:
										{
											position448, tokenIndex448, depth448 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l449
											}
											position++
											goto l448
										l449:
											position, tokenIndex, depth = position448, tokenIndex448, depth448
											if buffer[position] != rune('O') {
												goto l372
											}
											position++
										}
									l448:
										{
											position450, tokenIndex450, depth450 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l451
											}
											position++
											goto l450
										l451:
											position, tokenIndex, depth = position450, tokenIndex450, depth450
											if buffer[position] != rune('T') {
												goto l372
											}
											position++
										}
									l450:
										break
									case 'I', 'i':
										{
											position452, tokenIndex452, depth452 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l453
											}
											position++
											goto l452
										l453:
											position, tokenIndex, depth = position452, tokenIndex452, depth452
											if buffer[position] != rune('I') {
												goto l372
											}
											position++
										}
									l452:
										{
											position454, tokenIndex454, depth454 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l455
											}
											position++
											goto l454
										l455:
											position, tokenIndex, depth = position454, tokenIndex454, depth454
											if buffer[position] != rune('N') {
												goto l372
											}
											position++
										}
									l454:
										break
									case 'C', 'c':
										{
											position456, tokenIndex456, depth456 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l457
											}
											position++
											goto l456
										l457:
											position, tokenIndex, depth = position456, tokenIndex456, depth456
											if buffer[position] != rune('C') {
												goto l372
											}
											position++
										}
									l456:
										{
											position458, tokenIndex458, depth458 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l459
											}
											position++
											goto l458
										l459:
											position, tokenIndex, depth = position458, tokenIndex458, depth458
											if buffer[position] != rune('O') {
												goto l372
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
												goto l372
											}
											position++
										}
									l460:
										{
											position462, tokenIndex462, depth462 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l463
											}
											position++
											goto l462
										l463:
											position, tokenIndex, depth = position462, tokenIndex462, depth462
											if buffer[position] != rune('L') {
												goto l372
											}
											position++
										}
									l462:
										{
											position464, tokenIndex464, depth464 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l465
											}
											position++
											goto l464
										l465:
											position, tokenIndex, depth = position464, tokenIndex464, depth464
											if buffer[position] != rune('A') {
												goto l372
											}
											position++
										}
									l464:
										{
											position466, tokenIndex466, depth466 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l467
											}
											position++
											goto l466
										l467:
											position, tokenIndex, depth = position466, tokenIndex466, depth466
											if buffer[position] != rune('P') {
												goto l372
											}
											position++
										}
									l466:
										{
											position468, tokenIndex468, depth468 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l469
											}
											position++
											goto l468
										l469:
											position, tokenIndex, depth = position468, tokenIndex468, depth468
											if buffer[position] != rune('S') {
												goto l372
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
												goto l372
											}
											position++
										}
									l470:
										break
									case 'G', 'g':
										{
											position472, tokenIndex472, depth472 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l473
											}
											position++
											goto l472
										l473:
											position, tokenIndex, depth = position472, tokenIndex472, depth472
											if buffer[position] != rune('G') {
												goto l372
											}
											position++
										}
									l472:
										{
											position474, tokenIndex474, depth474 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l475
											}
											position++
											goto l474
										l475:
											position, tokenIndex, depth = position474, tokenIndex474, depth474
											if buffer[position] != rune('R') {
												goto l372
											}
											position++
										}
									l474:
										{
											position476, tokenIndex476, depth476 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l477
											}
											position++
											goto l476
										l477:
											position, tokenIndex, depth = position476, tokenIndex476, depth476
											if buffer[position] != rune('O') {
												goto l372
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
												goto l372
											}
											position++
										}
									l478:
										{
											position480, tokenIndex480, depth480 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l481
											}
											position++
											goto l480
										l481:
											position, tokenIndex, depth = position480, tokenIndex480, depth480
											if buffer[position] != rune('P') {
												goto l372
											}
											position++
										}
									l480:
										break
									case 'D', 'd':
										{
											position482, tokenIndex482, depth482 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l483
											}
											position++
											goto l482
										l483:
											position, tokenIndex, depth = position482, tokenIndex482, depth482
											if buffer[position] != rune('D') {
												goto l372
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
												goto l372
											}
											position++
										}
									l484:
										{
											position486, tokenIndex486, depth486 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l487
											}
											position++
											goto l486
										l487:
											position, tokenIndex, depth = position486, tokenIndex486, depth486
											if buffer[position] != rune('S') {
												goto l372
											}
											position++
										}
									l486:
										{
											position488, tokenIndex488, depth488 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l489
											}
											position++
											goto l488
										l489:
											position, tokenIndex, depth = position488, tokenIndex488, depth488
											if buffer[position] != rune('C') {
												goto l372
											}
											position++
										}
									l488:
										{
											position490, tokenIndex490, depth490 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l491
											}
											position++
											goto l490
										l491:
											position, tokenIndex, depth = position490, tokenIndex490, depth490
											if buffer[position] != rune('R') {
												goto l372
											}
											position++
										}
									l490:
										{
											position492, tokenIndex492, depth492 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l493
											}
											position++
											goto l492
										l493:
											position, tokenIndex, depth = position492, tokenIndex492, depth492
											if buffer[position] != rune('I') {
												goto l372
											}
											position++
										}
									l492:
										{
											position494, tokenIndex494, depth494 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l495
											}
											position++
											goto l494
										l495:
											position, tokenIndex, depth = position494, tokenIndex494, depth494
											if buffer[position] != rune('B') {
												goto l372
											}
											position++
										}
									l494:
										{
											position496, tokenIndex496, depth496 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l497
											}
											position++
											goto l496
										l497:
											position, tokenIndex, depth = position496, tokenIndex496, depth496
											if buffer[position] != rune('E') {
												goto l372
											}
											position++
										}
									l496:
										break
									case 'B', 'b':
										{
											position498, tokenIndex498, depth498 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l499
											}
											position++
											goto l498
										l499:
											position, tokenIndex, depth = position498, tokenIndex498, depth498
											if buffer[position] != rune('B') {
												goto l372
											}
											position++
										}
									l498:
										{
											position500, tokenIndex500, depth500 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l501
											}
											position++
											goto l500
										l501:
											position, tokenIndex, depth = position500, tokenIndex500, depth500
											if buffer[position] != rune('Y') {
												goto l372
											}
											position++
										}
									l500:
										break
									case 'A', 'a':
										{
											position502, tokenIndex502, depth502 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l503
											}
											position++
											goto l502
										l503:
											position, tokenIndex, depth = position502, tokenIndex502, depth502
											if buffer[position] != rune('A') {
												goto l372
											}
											position++
										}
									l502:
										{
											position504, tokenIndex504, depth504 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l505
											}
											position++
											goto l504
										l505:
											position, tokenIndex, depth = position504, tokenIndex504, depth504
											if buffer[position] != rune('S') {
												goto l372
											}
											position++
										}
									l504:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l372
										}
										break
									}
								}

							}
						l374:
							depth--
							add(ruleKEYWORD, position373)
						}
						if !_rules[ruleKEY]() {
							goto l372
						}
						goto l366
					l372:
						position, tokenIndex, depth = position372, tokenIndex372, depth372
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l366
					}
				l506:
					{
						position507, tokenIndex507, depth507 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l507
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l507
						}
						goto l506
					l507:
						position, tokenIndex, depth = position507, tokenIndex507, depth507
					}
				}
			l368:
				depth--
				add(ruleIDENTIFIER, position367)
			}
			return true
		l366:
			position, tokenIndex, depth = position366, tokenIndex366, depth366
			return false
		},
		/* 34 TIMESTAMP <- <((_ <(NUMBER ([a-z] / [A-Z])*)>) / (_ STRING) / (_ <(('n' / 'N') ('o' / 'O') ('w' / 'W'))>))> */
		nil,
		/* 35 ID_SEGMENT <- <(_ ID_START ID_CONT*)> */
		func() bool {
			position509, tokenIndex509, depth509 := position, tokenIndex, depth
			{
				position510 := position
				depth++
				if !_rules[rule_]() {
					goto l509
				}
				if !_rules[ruleID_START]() {
					goto l509
				}
			l511:
				{
					position512, tokenIndex512, depth512 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l512
					}
					goto l511
				l512:
					position, tokenIndex, depth = position512, tokenIndex512, depth512
				}
				depth--
				add(ruleID_SEGMENT, position510)
			}
			return true
		l509:
			position, tokenIndex, depth = position509, tokenIndex509, depth509
			return false
		},
		/* 36 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position513, tokenIndex513, depth513 := position, tokenIndex, depth
			{
				position514 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l513
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l513
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l513
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position514)
			}
			return true
		l513:
			position, tokenIndex, depth = position513, tokenIndex513, depth513
			return false
		},
		/* 37 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position516, tokenIndex516, depth516 := position, tokenIndex, depth
			{
				position517 := position
				depth++
				{
					position518, tokenIndex518, depth518 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l519
					}
					goto l518
				l519:
					position, tokenIndex, depth = position518, tokenIndex518, depth518
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l516
					}
					position++
				}
			l518:
				depth--
				add(ruleID_CONT, position517)
			}
			return true
		l516:
			position, tokenIndex, depth = position516, tokenIndex516, depth516
			return false
		},
		/* 38 PROPERTY_KEY <- <(((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> KEY _ (('b' / 'B') ('y' / 'Y')))) | (&('R' | 'r') <(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))>) | (&('T' | 't') <(('t' / 'T') ('o' / 'O'))>) | (&('F' | 'f') <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))>)) KEY)> */
		func() bool {
			position520, tokenIndex520, depth520 := position, tokenIndex, depth
			{
				position521 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position523 := position
							depth++
							{
								position524, tokenIndex524, depth524 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l525
								}
								position++
								goto l524
							l525:
								position, tokenIndex, depth = position524, tokenIndex524, depth524
								if buffer[position] != rune('S') {
									goto l520
								}
								position++
							}
						l524:
							{
								position526, tokenIndex526, depth526 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l527
								}
								position++
								goto l526
							l527:
								position, tokenIndex, depth = position526, tokenIndex526, depth526
								if buffer[position] != rune('A') {
									goto l520
								}
								position++
							}
						l526:
							{
								position528, tokenIndex528, depth528 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l529
								}
								position++
								goto l528
							l529:
								position, tokenIndex, depth = position528, tokenIndex528, depth528
								if buffer[position] != rune('M') {
									goto l520
								}
								position++
							}
						l528:
							{
								position530, tokenIndex530, depth530 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l531
								}
								position++
								goto l530
							l531:
								position, tokenIndex, depth = position530, tokenIndex530, depth530
								if buffer[position] != rune('P') {
									goto l520
								}
								position++
							}
						l530:
							{
								position532, tokenIndex532, depth532 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l533
								}
								position++
								goto l532
							l533:
								position, tokenIndex, depth = position532, tokenIndex532, depth532
								if buffer[position] != rune('L') {
									goto l520
								}
								position++
							}
						l532:
							{
								position534, tokenIndex534, depth534 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l535
								}
								position++
								goto l534
							l535:
								position, tokenIndex, depth = position534, tokenIndex534, depth534
								if buffer[position] != rune('E') {
									goto l520
								}
								position++
							}
						l534:
							depth--
							add(rulePegText, position523)
						}
						if !_rules[ruleKEY]() {
							goto l520
						}
						if !_rules[rule_]() {
							goto l520
						}
						{
							position536, tokenIndex536, depth536 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l537
							}
							position++
							goto l536
						l537:
							position, tokenIndex, depth = position536, tokenIndex536, depth536
							if buffer[position] != rune('B') {
								goto l520
							}
							position++
						}
					l536:
						{
							position538, tokenIndex538, depth538 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l539
							}
							position++
							goto l538
						l539:
							position, tokenIndex, depth = position538, tokenIndex538, depth538
							if buffer[position] != rune('Y') {
								goto l520
							}
							position++
						}
					l538:
						break
					case 'R', 'r':
						{
							position540 := position
							depth++
							{
								position541, tokenIndex541, depth541 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l542
								}
								position++
								goto l541
							l542:
								position, tokenIndex, depth = position541, tokenIndex541, depth541
								if buffer[position] != rune('R') {
									goto l520
								}
								position++
							}
						l541:
							{
								position543, tokenIndex543, depth543 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l544
								}
								position++
								goto l543
							l544:
								position, tokenIndex, depth = position543, tokenIndex543, depth543
								if buffer[position] != rune('E') {
									goto l520
								}
								position++
							}
						l543:
							{
								position545, tokenIndex545, depth545 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l546
								}
								position++
								goto l545
							l546:
								position, tokenIndex, depth = position545, tokenIndex545, depth545
								if buffer[position] != rune('S') {
									goto l520
								}
								position++
							}
						l545:
							{
								position547, tokenIndex547, depth547 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l548
								}
								position++
								goto l547
							l548:
								position, tokenIndex, depth = position547, tokenIndex547, depth547
								if buffer[position] != rune('O') {
									goto l520
								}
								position++
							}
						l547:
							{
								position549, tokenIndex549, depth549 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l550
								}
								position++
								goto l549
							l550:
								position, tokenIndex, depth = position549, tokenIndex549, depth549
								if buffer[position] != rune('L') {
									goto l520
								}
								position++
							}
						l549:
							{
								position551, tokenIndex551, depth551 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l552
								}
								position++
								goto l551
							l552:
								position, tokenIndex, depth = position551, tokenIndex551, depth551
								if buffer[position] != rune('U') {
									goto l520
								}
								position++
							}
						l551:
							{
								position553, tokenIndex553, depth553 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l554
								}
								position++
								goto l553
							l554:
								position, tokenIndex, depth = position553, tokenIndex553, depth553
								if buffer[position] != rune('T') {
									goto l520
								}
								position++
							}
						l553:
							{
								position555, tokenIndex555, depth555 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l556
								}
								position++
								goto l555
							l556:
								position, tokenIndex, depth = position555, tokenIndex555, depth555
								if buffer[position] != rune('I') {
									goto l520
								}
								position++
							}
						l555:
							{
								position557, tokenIndex557, depth557 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l558
								}
								position++
								goto l557
							l558:
								position, tokenIndex, depth = position557, tokenIndex557, depth557
								if buffer[position] != rune('O') {
									goto l520
								}
								position++
							}
						l557:
							{
								position559, tokenIndex559, depth559 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l560
								}
								position++
								goto l559
							l560:
								position, tokenIndex, depth = position559, tokenIndex559, depth559
								if buffer[position] != rune('N') {
									goto l520
								}
								position++
							}
						l559:
							depth--
							add(rulePegText, position540)
						}
						break
					case 'T', 't':
						{
							position561 := position
							depth++
							{
								position562, tokenIndex562, depth562 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l563
								}
								position++
								goto l562
							l563:
								position, tokenIndex, depth = position562, tokenIndex562, depth562
								if buffer[position] != rune('T') {
									goto l520
								}
								position++
							}
						l562:
							{
								position564, tokenIndex564, depth564 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l565
								}
								position++
								goto l564
							l565:
								position, tokenIndex, depth = position564, tokenIndex564, depth564
								if buffer[position] != rune('O') {
									goto l520
								}
								position++
							}
						l564:
							depth--
							add(rulePegText, position561)
						}
						break
					default:
						{
							position566 := position
							depth++
							{
								position567, tokenIndex567, depth567 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l568
								}
								position++
								goto l567
							l568:
								position, tokenIndex, depth = position567, tokenIndex567, depth567
								if buffer[position] != rune('F') {
									goto l520
								}
								position++
							}
						l567:
							{
								position569, tokenIndex569, depth569 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l570
								}
								position++
								goto l569
							l570:
								position, tokenIndex, depth = position569, tokenIndex569, depth569
								if buffer[position] != rune('R') {
									goto l520
								}
								position++
							}
						l569:
							{
								position571, tokenIndex571, depth571 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l572
								}
								position++
								goto l571
							l572:
								position, tokenIndex, depth = position571, tokenIndex571, depth571
								if buffer[position] != rune('O') {
									goto l520
								}
								position++
							}
						l571:
							{
								position573, tokenIndex573, depth573 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l574
								}
								position++
								goto l573
							l574:
								position, tokenIndex, depth = position573, tokenIndex573, depth573
								if buffer[position] != rune('M') {
									goto l520
								}
								position++
							}
						l573:
							depth--
							add(rulePegText, position566)
						}
						break
					}
				}

				if !_rules[ruleKEY]() {
					goto l520
				}
				depth--
				add(rulePROPERTY_KEY, position521)
			}
			return true
		l520:
			position, tokenIndex, depth = position520, tokenIndex520, depth520
			return false
		},
		/* 39 PROPERTY_VALUE <- <TIMESTAMP> */
		nil,
		/* 40 KEYWORD <- <((('a' / 'A') ('l' / 'L') ('l' / 'L')) / (('a' / 'A') ('n' / 'N') ('d' / 'D')) / (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H') ('e' / 'E') ('s' / 'S')) / (('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T')) / ((&('M' | 'm') (('m' / 'M') ('e' / 'E') ('t' / 'T') ('r' / 'R') ('i' / 'I') ('c' / 'C') ('s' / 'S'))) | (&('W' | 'w') (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E'))) | (&('O' | 'o') (('o' / 'O') ('r' / 'R'))) | (&('N' | 'n') (('n' / 'N') ('o' / 'O') ('t' / 'T'))) | (&('I' | 'i') (('i' / 'I') ('n' / 'N'))) | (&('C' | 'c') (('c' / 'C') ('o' / 'O') ('l' / 'L') ('l' / 'L') ('a' / 'A') ('p' / 'P') ('s' / 'S') ('e' / 'E'))) | (&('G' | 'g') (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P'))) | (&('D' | 'd') (('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E'))) | (&('B' | 'b') (('b' / 'B') ('y' / 'Y'))) | (&('A' | 'a') (('a' / 'A') ('s' / 'S'))) | (&('F' | 'R' | 'S' | 'T' | 'f' | 'r' | 's' | 't') PROPERTY_KEY)))> */
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
			position585, tokenIndex585, depth585 := position, tokenIndex, depth
			{
				position586 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l585
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position586)
			}
			return true
		l585:
			position, tokenIndex, depth = position585, tokenIndex585, depth585
			return false
		},
		/* 50 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position587, tokenIndex587, depth587 := position, tokenIndex, depth
			{
				position588 := position
				depth++
				if buffer[position] != rune('"') {
					goto l587
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position588)
			}
			return true
		l587:
			position, tokenIndex, depth = position587, tokenIndex587, depth587
			return false
		},
		/* 51 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> QUOTE_SINGLE) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> QUOTE_DOUBLE))> */
		func() bool {
			position589, tokenIndex589, depth589 := position, tokenIndex, depth
			{
				position590 := position
				depth++
				{
					position591, tokenIndex591, depth591 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l592
					}
					{
						position593 := position
						depth++
					l594:
						{
							position595, tokenIndex595, depth595 := position, tokenIndex, depth
							{
								position596, tokenIndex596, depth596 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l596
								}
								goto l595
							l596:
								position, tokenIndex, depth = position596, tokenIndex596, depth596
							}
							if !_rules[ruleCHAR]() {
								goto l595
							}
							goto l594
						l595:
							position, tokenIndex, depth = position595, tokenIndex595, depth595
						}
						depth--
						add(rulePegText, position593)
					}
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l592
					}
					goto l591
				l592:
					position, tokenIndex, depth = position591, tokenIndex591, depth591
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l589
					}
					{
						position597 := position
						depth++
					l598:
						{
							position599, tokenIndex599, depth599 := position, tokenIndex, depth
							{
								position600, tokenIndex600, depth600 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l600
								}
								goto l599
							l600:
								position, tokenIndex, depth = position600, tokenIndex600, depth600
							}
							if !_rules[ruleCHAR]() {
								goto l599
							}
							goto l598
						l599:
							position, tokenIndex, depth = position599, tokenIndex599, depth599
						}
						depth--
						add(rulePegText, position597)
					}
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l589
					}
				}
			l591:
				depth--
				add(ruleSTRING, position590)
			}
			return true
		l589:
			position, tokenIndex, depth = position589, tokenIndex589, depth589
			return false
		},
		/* 52 CHAR <- <(('\\' ((&('"') QUOTE_DOUBLE) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position601, tokenIndex601, depth601 := position, tokenIndex, depth
			{
				position602 := position
				depth++
				{
					position603, tokenIndex603, depth603 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l604
					}
					position++
					{
						switch buffer[position] {
						case '"':
							if !_rules[ruleQUOTE_DOUBLE]() {
								goto l604
							}
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l604
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l604
							}
							break
						}
					}

					goto l603
				l604:
					position, tokenIndex, depth = position603, tokenIndex603, depth603
					{
						position606, tokenIndex606, depth606 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l606
						}
						goto l601
					l606:
						position, tokenIndex, depth = position606, tokenIndex606, depth606
					}
					if !matchDot() {
						goto l601
					}
				}
			l603:
				depth--
				add(ruleCHAR, position602)
			}
			return true
		l601:
			position, tokenIndex, depth = position601, tokenIndex601, depth601
			return false
		},
		/* 53 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position607, tokenIndex607, depth607 := position, tokenIndex, depth
			{
				position608 := position
				depth++
				{
					position609, tokenIndex609, depth609 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l610
					}
					position++
					goto l609
				l610:
					position, tokenIndex, depth = position609, tokenIndex609, depth609
					if buffer[position] != rune('\\') {
						goto l607
					}
					position++
				}
			l609:
				depth--
				add(ruleESCAPE_CLASS, position608)
			}
			return true
		l607:
			position, tokenIndex, depth = position607, tokenIndex607, depth607
			return false
		},
		/* 54 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position611, tokenIndex611, depth611 := position, tokenIndex, depth
			{
				position612 := position
				depth++
				{
					position613 := position
					depth++
					{
						position614, tokenIndex614, depth614 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l614
						}
						position++
						goto l615
					l614:
						position, tokenIndex, depth = position614, tokenIndex614, depth614
					}
				l615:
					{
						position616 := position
						depth++
						{
							position617, tokenIndex617, depth617 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l618
							}
							position++
							goto l617
						l618:
							position, tokenIndex, depth = position617, tokenIndex617, depth617
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l611
							}
							position++
						l619:
							{
								position620, tokenIndex620, depth620 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l620
								}
								position++
								goto l619
							l620:
								position, tokenIndex, depth = position620, tokenIndex620, depth620
							}
						}
					l617:
						depth--
						add(ruleNUMBER_NATURAL, position616)
					}
					depth--
					add(ruleNUMBER_INTEGER, position613)
				}
				{
					position621, tokenIndex621, depth621 := position, tokenIndex, depth
					{
						position623 := position
						depth++
						if buffer[position] != rune('.') {
							goto l621
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l621
						}
						position++
					l624:
						{
							position625, tokenIndex625, depth625 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l625
							}
							position++
							goto l624
						l625:
							position, tokenIndex, depth = position625, tokenIndex625, depth625
						}
						depth--
						add(ruleNUMBER_FRACTION, position623)
					}
					goto l622
				l621:
					position, tokenIndex, depth = position621, tokenIndex621, depth621
				}
			l622:
				{
					position626, tokenIndex626, depth626 := position, tokenIndex, depth
					{
						position628 := position
						depth++
						{
							position629, tokenIndex629, depth629 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l630
							}
							position++
							goto l629
						l630:
							position, tokenIndex, depth = position629, tokenIndex629, depth629
							if buffer[position] != rune('E') {
								goto l626
							}
							position++
						}
					l629:
						{
							position631, tokenIndex631, depth631 := position, tokenIndex, depth
							{
								position633, tokenIndex633, depth633 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l634
								}
								position++
								goto l633
							l634:
								position, tokenIndex, depth = position633, tokenIndex633, depth633
								if buffer[position] != rune('-') {
									goto l631
								}
								position++
							}
						l633:
							goto l632
						l631:
							position, tokenIndex, depth = position631, tokenIndex631, depth631
						}
					l632:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l626
						}
						position++
					l635:
						{
							position636, tokenIndex636, depth636 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l636
							}
							position++
							goto l635
						l636:
							position, tokenIndex, depth = position636, tokenIndex636, depth636
						}
						depth--
						add(ruleNUMBER_EXP, position628)
					}
					goto l627
				l626:
					position, tokenIndex, depth = position626, tokenIndex626, depth626
				}
			l627:
				depth--
				add(ruleNUMBER, position612)
			}
			return true
		l611:
			position, tokenIndex, depth = position611, tokenIndex611, depth611
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
			position642, tokenIndex642, depth642 := position, tokenIndex, depth
			{
				position643 := position
				depth++
				if buffer[position] != rune('(') {
					goto l642
				}
				position++
				depth--
				add(rulePAREN_OPEN, position643)
			}
			return true
		l642:
			position, tokenIndex, depth = position642, tokenIndex642, depth642
			return false
		},
		/* 61 PAREN_CLOSE <- <')'> */
		func() bool {
			position644, tokenIndex644, depth644 := position, tokenIndex, depth
			{
				position645 := position
				depth++
				if buffer[position] != rune(')') {
					goto l644
				}
				position++
				depth--
				add(rulePAREN_CLOSE, position645)
			}
			return true
		l644:
			position, tokenIndex, depth = position644, tokenIndex644, depth644
			return false
		},
		/* 62 COMMA <- <','> */
		func() bool {
			position646, tokenIndex646, depth646 := position, tokenIndex, depth
			{
				position647 := position
				depth++
				if buffer[position] != rune(',') {
					goto l646
				}
				position++
				depth--
				add(ruleCOMMA, position647)
			}
			return true
		l646:
			position, tokenIndex, depth = position646, tokenIndex646, depth646
			return false
		},
		/* 63 _ <- <((&('/') COMMENT_BLOCK) | (&('-') COMMENT_TRAIL) | (&('\t' | '\n' | ' ') SPACE))*> */
		func() bool {
			{
				position649 := position
				depth++
			l650:
				{
					position651, tokenIndex651, depth651 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '/':
							{
								position653 := position
								depth++
								if buffer[position] != rune('/') {
									goto l651
								}
								position++
								if buffer[position] != rune('*') {
									goto l651
								}
								position++
							l654:
								{
									position655, tokenIndex655, depth655 := position, tokenIndex, depth
									{
										position656, tokenIndex656, depth656 := position, tokenIndex, depth
										if buffer[position] != rune('*') {
											goto l656
										}
										position++
										if buffer[position] != rune('/') {
											goto l656
										}
										position++
										goto l655
									l656:
										position, tokenIndex, depth = position656, tokenIndex656, depth656
									}
									if !matchDot() {
										goto l655
									}
									goto l654
								l655:
									position, tokenIndex, depth = position655, tokenIndex655, depth655
								}
								if buffer[position] != rune('*') {
									goto l651
								}
								position++
								if buffer[position] != rune('/') {
									goto l651
								}
								position++
								depth--
								add(ruleCOMMENT_BLOCK, position653)
							}
							break
						case '-':
							{
								position657 := position
								depth++
								if buffer[position] != rune('-') {
									goto l651
								}
								position++
								if buffer[position] != rune('-') {
									goto l651
								}
								position++
							l658:
								{
									position659, tokenIndex659, depth659 := position, tokenIndex, depth
									{
										position660, tokenIndex660, depth660 := position, tokenIndex, depth
										if buffer[position] != rune('\n') {
											goto l660
										}
										position++
										goto l659
									l660:
										position, tokenIndex, depth = position660, tokenIndex660, depth660
									}
									if !matchDot() {
										goto l659
									}
									goto l658
								l659:
									position, tokenIndex, depth = position659, tokenIndex659, depth659
								}
								depth--
								add(ruleCOMMENT_TRAIL, position657)
							}
							break
						default:
							{
								position661 := position
								depth++
								{
									switch buffer[position] {
									case '\t':
										if buffer[position] != rune('\t') {
											goto l651
										}
										position++
										break
									case '\n':
										if buffer[position] != rune('\n') {
											goto l651
										}
										position++
										break
									default:
										if buffer[position] != rune(' ') {
											goto l651
										}
										position++
										break
									}
								}

								depth--
								add(ruleSPACE, position661)
							}
							break
						}
					}

					goto l650
				l651:
					position, tokenIndex, depth = position651, tokenIndex651, depth651
				}
				depth--
				add(rule_, position649)
			}
			return true
		},
		/* 64 COMMENT_TRAIL <- <('-' '-' (!'\n' .)*)> */
		nil,
		/* 65 COMMENT_BLOCK <- <('/' '*' (!('*' '/') .)* ('*' '/'))> */
		nil,
		/* 66 KEY <- <!ID_CONT> */
		func() bool {
			position665, tokenIndex665, depth665 := position, tokenIndex, depth
			{
				position666 := position
				depth++
				{
					position667, tokenIndex667, depth667 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l667
					}
					goto l665
				l667:
					position, tokenIndex, depth = position667, tokenIndex667, depth667
				}
				depth--
				add(ruleKEY, position666)
			}
			return true
		l665:
			position, tokenIndex, depth = position665, tokenIndex665, depth665
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
		/* 71 Action2 <- <{ p.addNullMatchesClause() }> */
		nil,
		/* 72 Action3 <- <{ p.addMatchesClause() }> */
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
