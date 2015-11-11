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
	ruleAction51

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
	"Action51",

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
	rules  [122]func() bool
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
			p.addIndexClause()
		case ruleAction7:
			p.noIndexClause()
		case ruleAction8:
			p.makeDescribe()
		case ruleAction9:
			p.addEvaluationContext()
		case ruleAction10:
			p.addPropertyKey(buffer[begin:end])
		case ruleAction11:
			p.addPropertyValue(buffer[begin:end])
		case ruleAction12:
			p.insertPropertyKeyValue()
		case ruleAction13:
			p.checkPropertyClause()
		case ruleAction14:
			p.addNullPredicate()
		case ruleAction15:
			p.addExpressionList()
		case ruleAction16:
			p.appendExpression()
		case ruleAction17:
			p.appendExpression()
		case ruleAction18:
			p.addOperatorLiteral("+")
		case ruleAction19:
			p.addOperatorLiteral("-")
		case ruleAction20:
			p.addOperatorFunction()
		case ruleAction21:
			p.addOperatorLiteral("/")
		case ruleAction22:
			p.addOperatorLiteral("*")
		case ruleAction23:
			p.addOperatorFunction()
		case ruleAction24:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction25:
			p.addExpressionList()
		case ruleAction26:

			p.addExpressionList()
			p.addGroupBy()

		case ruleAction27:

			p.addPipeExpression()

		case ruleAction28:
			p.addDurationNode(text)
		case ruleAction29:
			p.addNumberNode(buffer[begin:end])
		case ruleAction30:
			p.addStringNode(unescapeLiteral(buffer[begin:end]))
		case ruleAction31:
			p.addGroupBy()
		case ruleAction32:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction33:

			p.addFunctionInvocation()

		case ruleAction34:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction35:
			p.addNullPredicate()
		case ruleAction36:

			p.addMetricExpression()

		case ruleAction37:

			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))

		case ruleAction38:

			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))

		case ruleAction39:

			p.appendCollapseBy(unescapeLiteral(text))

		case ruleAction40:
			p.appendCollapseBy(unescapeLiteral(text))
		case ruleAction41:
			p.addOrPredicate()
		case ruleAction42:
			p.addAndPredicate()
		case ruleAction43:
			p.addNotPredicate()
		case ruleAction44:

			p.addLiteralMatcher()

		case ruleAction45:

			p.addLiteralMatcher()
			p.addNotPredicate()

		case ruleAction46:

			p.addRegexMatcher()

		case ruleAction47:

			p.addListMatcher()

		case ruleAction48:

			p.addStringLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction49:
			p.addLiteralList()
		case ruleAction50:

			p.appendLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction51:
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
								add(ruleAction9, position)
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
									add(ruleAction10, position)
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
									add(ruleAction11, position)
								}
								{
									add(ruleAction12, position)
								}
								goto l21
							l22:
								position, tokenIndex, depth = position22, tokenIndex22, depth22
							}
							{
								add(ruleAction13, position)
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
								{
									position119, tokenIndex119, depth119 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l120
									}
									{
										position121, tokenIndex121, depth121 := position, tokenIndex, depth
										if buffer[position] != rune('i') {
											goto l122
										}
										position++
										goto l121
									l122:
										position, tokenIndex, depth = position121, tokenIndex121, depth121
										if buffer[position] != rune('I') {
											goto l120
										}
										position++
									}
								l121:
									{
										position123, tokenIndex123, depth123 := position, tokenIndex, depth
										if buffer[position] != rune('n') {
											goto l124
										}
										position++
										goto l123
									l124:
										position, tokenIndex, depth = position123, tokenIndex123, depth123
										if buffer[position] != rune('N') {
											goto l120
										}
										position++
									}
								l123:
									{
										position125, tokenIndex125, depth125 := position, tokenIndex, depth
										if buffer[position] != rune('d') {
											goto l126
										}
										position++
										goto l125
									l126:
										position, tokenIndex, depth = position125, tokenIndex125, depth125
										if buffer[position] != rune('D') {
											goto l120
										}
										position++
									}
								l125:
									{
										position127, tokenIndex127, depth127 := position, tokenIndex, depth
										if buffer[position] != rune('e') {
											goto l128
										}
										position++
										goto l127
									l128:
										position, tokenIndex, depth = position127, tokenIndex127, depth127
										if buffer[position] != rune('E') {
											goto l120
										}
										position++
									}
								l127:
									{
										position129, tokenIndex129, depth129 := position, tokenIndex, depth
										if buffer[position] != rune('x') {
											goto l130
										}
										position++
										goto l129
									l130:
										position, tokenIndex, depth = position129, tokenIndex129, depth129
										if buffer[position] != rune('X') {
											goto l120
										}
										position++
									}
								l129:
									if !_rules[ruleKEY]() {
										goto l120
									}
									{
										add(ruleAction6, position)
									}
									goto l119
								l120:
									position, tokenIndex, depth = position119, tokenIndex119, depth119
									{
										add(ruleAction7, position)
									}
								}
							l119:
								if !_rules[ruleoptionalPredicateClause]() {
									goto l0
								}
								{
									add(ruleAction8, position)
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
					position134, tokenIndex134, depth134 := position, tokenIndex, depth
					if !matchDot() {
						goto l134
					}
					goto l0
				l134:
					position, tokenIndex, depth = position134, tokenIndex134, depth134
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
		/* 7 describeSingleStmt <- <(_ <METRIC_NAME> Action5 ((_ (('i' / 'I') ('n' / 'N') ('d' / 'D') ('e' / 'E') ('x' / 'X')) KEY Action6) / Action7) optionalPredicateClause Action8)> */
		nil,
		/* 8 propertyClause <- <(Action9 (_ PROPERTY_KEY Action10 _ PROPERTY_VALUE Action11 Action12)* Action13)> */
		nil,
		/* 9 optionalPredicateClause <- <(predicateClause / Action14)> */
		func() bool {
			{
				position144 := position
				depth++
				{
					position145, tokenIndex145, depth145 := position, tokenIndex, depth
					{
						position147 := position
						depth++
						if !_rules[rule_]() {
							goto l146
						}
						{
							position148, tokenIndex148, depth148 := position, tokenIndex, depth
							if buffer[position] != rune('w') {
								goto l149
							}
							position++
							goto l148
						l149:
							position, tokenIndex, depth = position148, tokenIndex148, depth148
							if buffer[position] != rune('W') {
								goto l146
							}
							position++
						}
					l148:
						{
							position150, tokenIndex150, depth150 := position, tokenIndex, depth
							if buffer[position] != rune('h') {
								goto l151
							}
							position++
							goto l150
						l151:
							position, tokenIndex, depth = position150, tokenIndex150, depth150
							if buffer[position] != rune('H') {
								goto l146
							}
							position++
						}
					l150:
						{
							position152, tokenIndex152, depth152 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l153
							}
							position++
							goto l152
						l153:
							position, tokenIndex, depth = position152, tokenIndex152, depth152
							if buffer[position] != rune('E') {
								goto l146
							}
							position++
						}
					l152:
						{
							position154, tokenIndex154, depth154 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l155
							}
							position++
							goto l154
						l155:
							position, tokenIndex, depth = position154, tokenIndex154, depth154
							if buffer[position] != rune('R') {
								goto l146
							}
							position++
						}
					l154:
						{
							position156, tokenIndex156, depth156 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l157
							}
							position++
							goto l156
						l157:
							position, tokenIndex, depth = position156, tokenIndex156, depth156
							if buffer[position] != rune('E') {
								goto l146
							}
							position++
						}
					l156:
						if !_rules[ruleKEY]() {
							goto l146
						}
						if !_rules[rule_]() {
							goto l146
						}
						if !_rules[rulepredicate_1]() {
							goto l146
						}
						depth--
						add(rulepredicateClause, position147)
					}
					goto l145
				l146:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					{
						add(ruleAction14, position)
					}
				}
			l145:
				depth--
				add(ruleoptionalPredicateClause, position144)
			}
			return true
		},
		/* 10 expressionList <- <(Action15 expression_start Action16 (_ COMMA expression_start Action17)*)> */
		func() bool {
			position159, tokenIndex159, depth159 := position, tokenIndex, depth
			{
				position160 := position
				depth++
				{
					add(ruleAction15, position)
				}
				if !_rules[ruleexpression_start]() {
					goto l159
				}
				{
					add(ruleAction16, position)
				}
			l163:
				{
					position164, tokenIndex164, depth164 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l164
					}
					if !_rules[ruleCOMMA]() {
						goto l164
					}
					if !_rules[ruleexpression_start]() {
						goto l164
					}
					{
						add(ruleAction17, position)
					}
					goto l163
				l164:
					position, tokenIndex, depth = position164, tokenIndex164, depth164
				}
				depth--
				add(ruleexpressionList, position160)
			}
			return true
		l159:
			position, tokenIndex, depth = position159, tokenIndex159, depth159
			return false
		},
		/* 11 expression_start <- <(expression_sum add_pipe)> */
		func() bool {
			position166, tokenIndex166, depth166 := position, tokenIndex, depth
			{
				position167 := position
				depth++
				{
					position168 := position
					depth++
					if !_rules[ruleexpression_product]() {
						goto l166
					}
				l169:
					{
						position170, tokenIndex170, depth170 := position, tokenIndex, depth
						if !_rules[ruleadd_pipe]() {
							goto l170
						}
						{
							position171, tokenIndex171, depth171 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l172
							}
							{
								position173 := position
								depth++
								if buffer[position] != rune('+') {
									goto l172
								}
								position++
								depth--
								add(ruleOP_ADD, position173)
							}
							{
								add(ruleAction18, position)
							}
							goto l171
						l172:
							position, tokenIndex, depth = position171, tokenIndex171, depth171
							if !_rules[rule_]() {
								goto l170
							}
							{
								position175 := position
								depth++
								if buffer[position] != rune('-') {
									goto l170
								}
								position++
								depth--
								add(ruleOP_SUB, position175)
							}
							{
								add(ruleAction19, position)
							}
						}
					l171:
						if !_rules[ruleexpression_product]() {
							goto l170
						}
						{
							add(ruleAction20, position)
						}
						goto l169
					l170:
						position, tokenIndex, depth = position170, tokenIndex170, depth170
					}
					depth--
					add(ruleexpression_sum, position168)
				}
				if !_rules[ruleadd_pipe]() {
					goto l166
				}
				depth--
				add(ruleexpression_start, position167)
			}
			return true
		l166:
			position, tokenIndex, depth = position166, tokenIndex166, depth166
			return false
		},
		/* 12 expression_sum <- <(expression_product (add_pipe ((_ OP_ADD Action18) / (_ OP_SUB Action19)) expression_product Action20)*)> */
		nil,
		/* 13 expression_product <- <(expression_atom (add_pipe ((_ OP_DIV Action21) / (_ OP_MULT Action22)) expression_atom Action23)*)> */
		func() bool {
			position179, tokenIndex179, depth179 := position, tokenIndex, depth
			{
				position180 := position
				depth++
				if !_rules[ruleexpression_atom]() {
					goto l179
				}
			l181:
				{
					position182, tokenIndex182, depth182 := position, tokenIndex, depth
					if !_rules[ruleadd_pipe]() {
						goto l182
					}
					{
						position183, tokenIndex183, depth183 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l184
						}
						{
							position185 := position
							depth++
							if buffer[position] != rune('/') {
								goto l184
							}
							position++
							depth--
							add(ruleOP_DIV, position185)
						}
						{
							add(ruleAction21, position)
						}
						goto l183
					l184:
						position, tokenIndex, depth = position183, tokenIndex183, depth183
						if !_rules[rule_]() {
							goto l182
						}
						{
							position187 := position
							depth++
							if buffer[position] != rune('*') {
								goto l182
							}
							position++
							depth--
							add(ruleOP_MULT, position187)
						}
						{
							add(ruleAction22, position)
						}
					}
				l183:
					if !_rules[ruleexpression_atom]() {
						goto l182
					}
					{
						add(ruleAction23, position)
					}
					goto l181
				l182:
					position, tokenIndex, depth = position182, tokenIndex182, depth182
				}
				depth--
				add(ruleexpression_product, position180)
			}
			return true
		l179:
			position, tokenIndex, depth = position179, tokenIndex179, depth179
			return false
		},
		/* 14 add_pipe <- <(_ OP_PIPE _ <IDENTIFIER> Action24 ((_ PAREN_OPEN (expressionList / Action25) optionalGroupBy _ PAREN_CLOSE) / Action26) Action27)*> */
		func() bool {
			{
				position191 := position
				depth++
			l192:
				{
					position193, tokenIndex193, depth193 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l193
					}
					{
						position194 := position
						depth++
						if buffer[position] != rune('|') {
							goto l193
						}
						position++
						depth--
						add(ruleOP_PIPE, position194)
					}
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
						add(ruleAction24, position)
					}
					{
						position197, tokenIndex197, depth197 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l198
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l198
						}
						{
							position199, tokenIndex199, depth199 := position, tokenIndex, depth
							if !_rules[ruleexpressionList]() {
								goto l200
							}
							goto l199
						l200:
							position, tokenIndex, depth = position199, tokenIndex199, depth199
							{
								add(ruleAction25, position)
							}
						}
					l199:
						if !_rules[ruleoptionalGroupBy]() {
							goto l198
						}
						if !_rules[rule_]() {
							goto l198
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l198
						}
						goto l197
					l198:
						position, tokenIndex, depth = position197, tokenIndex197, depth197
						{
							add(ruleAction26, position)
						}
					}
				l197:
					{
						add(ruleAction27, position)
					}
					goto l192
				l193:
					position, tokenIndex, depth = position193, tokenIndex193, depth193
				}
				depth--
				add(ruleadd_pipe, position191)
			}
			return true
		},
		/* 15 expression_atom <- <(expression_function / expression_metric / (_ PAREN_OPEN expression_start _ PAREN_CLOSE) / (_ <DURATION> Action28) / (_ <NUMBER> Action29) / (_ STRING Action30))> */
		func() bool {
			position204, tokenIndex204, depth204 := position, tokenIndex, depth
			{
				position205 := position
				depth++
				{
					position206, tokenIndex206, depth206 := position, tokenIndex, depth
					{
						position208 := position
						depth++
						if !_rules[rule_]() {
							goto l207
						}
						{
							position209 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l207
							}
							depth--
							add(rulePegText, position209)
						}
						{
							add(ruleAction32, position)
						}
						if !_rules[rule_]() {
							goto l207
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l207
						}
						if !_rules[ruleexpressionList]() {
							goto l207
						}
						if !_rules[ruleoptionalGroupBy]() {
							goto l207
						}
						if !_rules[rule_]() {
							goto l207
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l207
						}
						{
							add(ruleAction33, position)
						}
						depth--
						add(ruleexpression_function, position208)
					}
					goto l206
				l207:
					position, tokenIndex, depth = position206, tokenIndex206, depth206
					{
						position213 := position
						depth++
						if !_rules[rule_]() {
							goto l212
						}
						{
							position214 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l212
							}
							depth--
							add(rulePegText, position214)
						}
						{
							add(ruleAction34, position)
						}
						{
							position216, tokenIndex216, depth216 := position, tokenIndex, depth
							{
								position218, tokenIndex218, depth218 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l219
								}
								if buffer[position] != rune('[') {
									goto l219
								}
								position++
								if !_rules[rulepredicate_1]() {
									goto l219
								}
								if !_rules[rule_]() {
									goto l219
								}
								if buffer[position] != rune(']') {
									goto l219
								}
								position++
								goto l218
							l219:
								position, tokenIndex, depth = position218, tokenIndex218, depth218
								{
									add(ruleAction35, position)
								}
							}
						l218:
							goto l217

							position, tokenIndex, depth = position216, tokenIndex216, depth216
						}
					l217:
						{
							add(ruleAction36, position)
						}
						depth--
						add(ruleexpression_metric, position213)
					}
					goto l206
				l212:
					position, tokenIndex, depth = position206, tokenIndex206, depth206
					if !_rules[rule_]() {
						goto l222
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l222
					}
					if !_rules[ruleexpression_start]() {
						goto l222
					}
					if !_rules[rule_]() {
						goto l222
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l222
					}
					goto l206
				l222:
					position, tokenIndex, depth = position206, tokenIndex206, depth206
					if !_rules[rule_]() {
						goto l223
					}
					{
						position224 := position
						depth++
						{
							position225 := position
							depth++
							if !_rules[ruleNUMBER]() {
								goto l223
							}
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l223
							}
							position++
						l226:
							{
								position227, tokenIndex227, depth227 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l227
								}
								position++
								goto l226
							l227:
								position, tokenIndex, depth = position227, tokenIndex227, depth227
							}
							if !_rules[ruleKEY]() {
								goto l223
							}
							depth--
							add(ruleDURATION, position225)
						}
						depth--
						add(rulePegText, position224)
					}
					{
						add(ruleAction28, position)
					}
					goto l206
				l223:
					position, tokenIndex, depth = position206, tokenIndex206, depth206
					if !_rules[rule_]() {
						goto l229
					}
					{
						position230 := position
						depth++
						if !_rules[ruleNUMBER]() {
							goto l229
						}
						depth--
						add(rulePegText, position230)
					}
					{
						add(ruleAction29, position)
					}
					goto l206
				l229:
					position, tokenIndex, depth = position206, tokenIndex206, depth206
					if !_rules[rule_]() {
						goto l204
					}
					if !_rules[ruleSTRING]() {
						goto l204
					}
					{
						add(ruleAction30, position)
					}
				}
			l206:
				depth--
				add(ruleexpression_atom, position205)
			}
			return true
		l204:
			position, tokenIndex, depth = position204, tokenIndex204, depth204
			return false
		},
		/* 16 optionalGroupBy <- <(Action31 (groupByClause / collapseByClause)?)> */
		func() bool {
			{
				position234 := position
				depth++
				{
					add(ruleAction31, position)
				}
				{
					position236, tokenIndex236, depth236 := position, tokenIndex, depth
					{
						position238, tokenIndex238, depth238 := position, tokenIndex, depth
						{
							position240 := position
							depth++
							if !_rules[rule_]() {
								goto l239
							}
							{
								position241, tokenIndex241, depth241 := position, tokenIndex, depth
								if buffer[position] != rune('g') {
									goto l242
								}
								position++
								goto l241
							l242:
								position, tokenIndex, depth = position241, tokenIndex241, depth241
								if buffer[position] != rune('G') {
									goto l239
								}
								position++
							}
						l241:
							{
								position243, tokenIndex243, depth243 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l244
								}
								position++
								goto l243
							l244:
								position, tokenIndex, depth = position243, tokenIndex243, depth243
								if buffer[position] != rune('R') {
									goto l239
								}
								position++
							}
						l243:
							{
								position245, tokenIndex245, depth245 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l246
								}
								position++
								goto l245
							l246:
								position, tokenIndex, depth = position245, tokenIndex245, depth245
								if buffer[position] != rune('O') {
									goto l239
								}
								position++
							}
						l245:
							{
								position247, tokenIndex247, depth247 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l248
								}
								position++
								goto l247
							l248:
								position, tokenIndex, depth = position247, tokenIndex247, depth247
								if buffer[position] != rune('U') {
									goto l239
								}
								position++
							}
						l247:
							{
								position249, tokenIndex249, depth249 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l250
								}
								position++
								goto l249
							l250:
								position, tokenIndex, depth = position249, tokenIndex249, depth249
								if buffer[position] != rune('P') {
									goto l239
								}
								position++
							}
						l249:
							if !_rules[ruleKEY]() {
								goto l239
							}
							if !_rules[rule_]() {
								goto l239
							}
							{
								position251, tokenIndex251, depth251 := position, tokenIndex, depth
								if buffer[position] != rune('b') {
									goto l252
								}
								position++
								goto l251
							l252:
								position, tokenIndex, depth = position251, tokenIndex251, depth251
								if buffer[position] != rune('B') {
									goto l239
								}
								position++
							}
						l251:
							{
								position253, tokenIndex253, depth253 := position, tokenIndex, depth
								if buffer[position] != rune('y') {
									goto l254
								}
								position++
								goto l253
							l254:
								position, tokenIndex, depth = position253, tokenIndex253, depth253
								if buffer[position] != rune('Y') {
									goto l239
								}
								position++
							}
						l253:
							if !_rules[ruleKEY]() {
								goto l239
							}
							if !_rules[rule_]() {
								goto l239
							}
							{
								position255 := position
								depth++
								if !_rules[ruleCOLUMN_NAME]() {
									goto l239
								}
								depth--
								add(rulePegText, position255)
							}
							{
								add(ruleAction37, position)
							}
						l257:
							{
								position258, tokenIndex258, depth258 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l258
								}
								if !_rules[ruleCOMMA]() {
									goto l258
								}
								if !_rules[rule_]() {
									goto l258
								}
								{
									position259 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l258
									}
									depth--
									add(rulePegText, position259)
								}
								{
									add(ruleAction38, position)
								}
								goto l257
							l258:
								position, tokenIndex, depth = position258, tokenIndex258, depth258
							}
							depth--
							add(rulegroupByClause, position240)
						}
						goto l238
					l239:
						position, tokenIndex, depth = position238, tokenIndex238, depth238
						{
							position261 := position
							depth++
							if !_rules[rule_]() {
								goto l236
							}
							{
								position262, tokenIndex262, depth262 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l263
								}
								position++
								goto l262
							l263:
								position, tokenIndex, depth = position262, tokenIndex262, depth262
								if buffer[position] != rune('C') {
									goto l236
								}
								position++
							}
						l262:
							{
								position264, tokenIndex264, depth264 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l265
								}
								position++
								goto l264
							l265:
								position, tokenIndex, depth = position264, tokenIndex264, depth264
								if buffer[position] != rune('O') {
									goto l236
								}
								position++
							}
						l264:
							{
								position266, tokenIndex266, depth266 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l267
								}
								position++
								goto l266
							l267:
								position, tokenIndex, depth = position266, tokenIndex266, depth266
								if buffer[position] != rune('L') {
									goto l236
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
									goto l236
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
									goto l236
								}
								position++
							}
						l270:
							{
								position272, tokenIndex272, depth272 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l273
								}
								position++
								goto l272
							l273:
								position, tokenIndex, depth = position272, tokenIndex272, depth272
								if buffer[position] != rune('P') {
									goto l236
								}
								position++
							}
						l272:
							{
								position274, tokenIndex274, depth274 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l275
								}
								position++
								goto l274
							l275:
								position, tokenIndex, depth = position274, tokenIndex274, depth274
								if buffer[position] != rune('S') {
									goto l236
								}
								position++
							}
						l274:
							{
								position276, tokenIndex276, depth276 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l277
								}
								position++
								goto l276
							l277:
								position, tokenIndex, depth = position276, tokenIndex276, depth276
								if buffer[position] != rune('E') {
									goto l236
								}
								position++
							}
						l276:
							if !_rules[ruleKEY]() {
								goto l236
							}
							if !_rules[rule_]() {
								goto l236
							}
							{
								position278, tokenIndex278, depth278 := position, tokenIndex, depth
								if buffer[position] != rune('b') {
									goto l279
								}
								position++
								goto l278
							l279:
								position, tokenIndex, depth = position278, tokenIndex278, depth278
								if buffer[position] != rune('B') {
									goto l236
								}
								position++
							}
						l278:
							{
								position280, tokenIndex280, depth280 := position, tokenIndex, depth
								if buffer[position] != rune('y') {
									goto l281
								}
								position++
								goto l280
							l281:
								position, tokenIndex, depth = position280, tokenIndex280, depth280
								if buffer[position] != rune('Y') {
									goto l236
								}
								position++
							}
						l280:
							if !_rules[ruleKEY]() {
								goto l236
							}
							if !_rules[rule_]() {
								goto l236
							}
							{
								position282 := position
								depth++
								if !_rules[ruleCOLUMN_NAME]() {
									goto l236
								}
								depth--
								add(rulePegText, position282)
							}
							{
								add(ruleAction39, position)
							}
						l284:
							{
								position285, tokenIndex285, depth285 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l285
								}
								if !_rules[ruleCOMMA]() {
									goto l285
								}
								if !_rules[rule_]() {
									goto l285
								}
								{
									position286 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l285
									}
									depth--
									add(rulePegText, position286)
								}
								{
									add(ruleAction40, position)
								}
								goto l284
							l285:
								position, tokenIndex, depth = position285, tokenIndex285, depth285
							}
							depth--
							add(rulecollapseByClause, position261)
						}
					}
				l238:
					goto l237
				l236:
					position, tokenIndex, depth = position236, tokenIndex236, depth236
				}
			l237:
				depth--
				add(ruleoptionalGroupBy, position234)
			}
			return true
		},
		/* 17 expression_function <- <(_ <IDENTIFIER> Action32 _ PAREN_OPEN expressionList optionalGroupBy _ PAREN_CLOSE Action33)> */
		nil,
		/* 18 expression_metric <- <(_ <IDENTIFIER> Action34 ((_ '[' predicate_1 _ ']') / Action35)? Action36)> */
		nil,
		/* 19 groupByClause <- <(_ (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P')) KEY _ (('b' / 'B') ('y' / 'Y')) KEY _ <COLUMN_NAME> Action37 (_ COMMA _ <COLUMN_NAME> Action38)*)> */
		nil,
		/* 20 collapseByClause <- <(_ (('c' / 'C') ('o' / 'O') ('l' / 'L') ('l' / 'L') ('a' / 'A') ('p' / 'P') ('s' / 'S') ('e' / 'E')) KEY _ (('b' / 'B') ('y' / 'Y')) KEY _ <COLUMN_NAME> Action39 (_ COMMA _ <COLUMN_NAME> Action40)*)> */
		nil,
		/* 21 predicateClause <- <(_ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY _ predicate_1)> */
		nil,
		/* 22 predicate_1 <- <((predicate_2 _ OP_OR predicate_1 Action41) / predicate_2)> */
		func() bool {
			position293, tokenIndex293, depth293 := position, tokenIndex, depth
			{
				position294 := position
				depth++
				{
					position295, tokenIndex295, depth295 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
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
							if buffer[position] != rune('o') {
								goto l299
							}
							position++
							goto l298
						l299:
							position, tokenIndex, depth = position298, tokenIndex298, depth298
							if buffer[position] != rune('O') {
								goto l296
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
								goto l296
							}
							position++
						}
					l300:
						if !_rules[ruleKEY]() {
							goto l296
						}
						depth--
						add(ruleOP_OR, position297)
					}
					if !_rules[rulepredicate_1]() {
						goto l296
					}
					{
						add(ruleAction41, position)
					}
					goto l295
				l296:
					position, tokenIndex, depth = position295, tokenIndex295, depth295
					if !_rules[rulepredicate_2]() {
						goto l293
					}
				}
			l295:
				depth--
				add(rulepredicate_1, position294)
			}
			return true
		l293:
			position, tokenIndex, depth = position293, tokenIndex293, depth293
			return false
		},
		/* 23 predicate_2 <- <((predicate_3 _ OP_AND predicate_2 Action42) / predicate_3)> */
		func() bool {
			position303, tokenIndex303, depth303 := position, tokenIndex, depth
			{
				position304 := position
				depth++
				{
					position305, tokenIndex305, depth305 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l306
					}
					if !_rules[rule_]() {
						goto l306
					}
					{
						position307 := position
						depth++
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
								goto l306
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
								goto l306
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
								goto l306
							}
							position++
						}
					l312:
						if !_rules[ruleKEY]() {
							goto l306
						}
						depth--
						add(ruleOP_AND, position307)
					}
					if !_rules[rulepredicate_2]() {
						goto l306
					}
					{
						add(ruleAction42, position)
					}
					goto l305
				l306:
					position, tokenIndex, depth = position305, tokenIndex305, depth305
					if !_rules[rulepredicate_3]() {
						goto l303
					}
				}
			l305:
				depth--
				add(rulepredicate_2, position304)
			}
			return true
		l303:
			position, tokenIndex, depth = position303, tokenIndex303, depth303
			return false
		},
		/* 24 predicate_3 <- <((_ OP_NOT predicate_3 Action43) / (_ PAREN_OPEN predicate_1 _ PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position315, tokenIndex315, depth315 := position, tokenIndex, depth
			{
				position316 := position
				depth++
				{
					position317, tokenIndex317, depth317 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l318
					}
					{
						position319 := position
						depth++
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
								goto l318
							}
							position++
						}
					l320:
						{
							position322, tokenIndex322, depth322 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l323
							}
							position++
							goto l322
						l323:
							position, tokenIndex, depth = position322, tokenIndex322, depth322
							if buffer[position] != rune('O') {
								goto l318
							}
							position++
						}
					l322:
						{
							position324, tokenIndex324, depth324 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l325
							}
							position++
							goto l324
						l325:
							position, tokenIndex, depth = position324, tokenIndex324, depth324
							if buffer[position] != rune('T') {
								goto l318
							}
							position++
						}
					l324:
						if !_rules[ruleKEY]() {
							goto l318
						}
						depth--
						add(ruleOP_NOT, position319)
					}
					if !_rules[rulepredicate_3]() {
						goto l318
					}
					{
						add(ruleAction43, position)
					}
					goto l317
				l318:
					position, tokenIndex, depth = position317, tokenIndex317, depth317
					if !_rules[rule_]() {
						goto l327
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l327
					}
					if !_rules[rulepredicate_1]() {
						goto l327
					}
					if !_rules[rule_]() {
						goto l327
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l327
					}
					goto l317
				l327:
					position, tokenIndex, depth = position317, tokenIndex317, depth317
					{
						position328 := position
						depth++
						{
							position329, tokenIndex329, depth329 := position, tokenIndex, depth
							if !_rules[ruletagName]() {
								goto l330
							}
							if !_rules[rule_]() {
								goto l330
							}
							if buffer[position] != rune('=') {
								goto l330
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l330
							}
							{
								add(ruleAction44, position)
							}
							goto l329
						l330:
							position, tokenIndex, depth = position329, tokenIndex329, depth329
							if !_rules[ruletagName]() {
								goto l332
							}
							if !_rules[rule_]() {
								goto l332
							}
							if buffer[position] != rune('!') {
								goto l332
							}
							position++
							if buffer[position] != rune('=') {
								goto l332
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l332
							}
							{
								add(ruleAction45, position)
							}
							goto l329
						l332:
							position, tokenIndex, depth = position329, tokenIndex329, depth329
							if !_rules[ruletagName]() {
								goto l334
							}
							if !_rules[rule_]() {
								goto l334
							}
							{
								position335, tokenIndex335, depth335 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l336
								}
								position++
								goto l335
							l336:
								position, tokenIndex, depth = position335, tokenIndex335, depth335
								if buffer[position] != rune('M') {
									goto l334
								}
								position++
							}
						l335:
							{
								position337, tokenIndex337, depth337 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l338
								}
								position++
								goto l337
							l338:
								position, tokenIndex, depth = position337, tokenIndex337, depth337
								if buffer[position] != rune('A') {
									goto l334
								}
								position++
							}
						l337:
							{
								position339, tokenIndex339, depth339 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l340
								}
								position++
								goto l339
							l340:
								position, tokenIndex, depth = position339, tokenIndex339, depth339
								if buffer[position] != rune('T') {
									goto l334
								}
								position++
							}
						l339:
							{
								position341, tokenIndex341, depth341 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l342
								}
								position++
								goto l341
							l342:
								position, tokenIndex, depth = position341, tokenIndex341, depth341
								if buffer[position] != rune('C') {
									goto l334
								}
								position++
							}
						l341:
							{
								position343, tokenIndex343, depth343 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l344
								}
								position++
								goto l343
							l344:
								position, tokenIndex, depth = position343, tokenIndex343, depth343
								if buffer[position] != rune('H') {
									goto l334
								}
								position++
							}
						l343:
							if !_rules[ruleKEY]() {
								goto l334
							}
							if !_rules[ruleliteralString]() {
								goto l334
							}
							{
								add(ruleAction46, position)
							}
							goto l329
						l334:
							position, tokenIndex, depth = position329, tokenIndex329, depth329
							if !_rules[ruletagName]() {
								goto l315
							}
							if !_rules[rule_]() {
								goto l315
							}
							{
								position346, tokenIndex346, depth346 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l347
								}
								position++
								goto l346
							l347:
								position, tokenIndex, depth = position346, tokenIndex346, depth346
								if buffer[position] != rune('I') {
									goto l315
								}
								position++
							}
						l346:
							{
								position348, tokenIndex348, depth348 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l349
								}
								position++
								goto l348
							l349:
								position, tokenIndex, depth = position348, tokenIndex348, depth348
								if buffer[position] != rune('N') {
									goto l315
								}
								position++
							}
						l348:
							if !_rules[ruleKEY]() {
								goto l315
							}
							{
								position350 := position
								depth++
								{
									add(ruleAction49, position)
								}
								if !_rules[rule_]() {
									goto l315
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l315
								}
								if !_rules[ruleliteralListString]() {
									goto l315
								}
							l352:
								{
									position353, tokenIndex353, depth353 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l353
									}
									if !_rules[ruleCOMMA]() {
										goto l353
									}
									if !_rules[ruleliteralListString]() {
										goto l353
									}
									goto l352
								l353:
									position, tokenIndex, depth = position353, tokenIndex353, depth353
								}
								if !_rules[rule_]() {
									goto l315
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l315
								}
								depth--
								add(ruleliteralList, position350)
							}
							{
								add(ruleAction47, position)
							}
						}
					l329:
						depth--
						add(ruletagMatcher, position328)
					}
				}
			l317:
				depth--
				add(rulepredicate_3, position316)
			}
			return true
		l315:
			position, tokenIndex, depth = position315, tokenIndex315, depth315
			return false
		},
		/* 25 tagMatcher <- <((tagName _ '=' literalString Action44) / (tagName _ ('!' '=') literalString Action45) / (tagName _ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H')) KEY literalString Action46) / (tagName _ (('i' / 'I') ('n' / 'N')) KEY literalList Action47))> */
		nil,
		/* 26 literalString <- <(_ STRING Action48)> */
		func() bool {
			position356, tokenIndex356, depth356 := position, tokenIndex, depth
			{
				position357 := position
				depth++
				if !_rules[rule_]() {
					goto l356
				}
				if !_rules[ruleSTRING]() {
					goto l356
				}
				{
					add(ruleAction48, position)
				}
				depth--
				add(ruleliteralString, position357)
			}
			return true
		l356:
			position, tokenIndex, depth = position356, tokenIndex356, depth356
			return false
		},
		/* 27 literalList <- <(Action49 _ PAREN_OPEN literalListString (_ COMMA literalListString)* _ PAREN_CLOSE)> */
		nil,
		/* 28 literalListString <- <(_ STRING Action50)> */
		func() bool {
			position360, tokenIndex360, depth360 := position, tokenIndex, depth
			{
				position361 := position
				depth++
				if !_rules[rule_]() {
					goto l360
				}
				if !_rules[ruleSTRING]() {
					goto l360
				}
				{
					add(ruleAction50, position)
				}
				depth--
				add(ruleliteralListString, position361)
			}
			return true
		l360:
			position, tokenIndex, depth = position360, tokenIndex360, depth360
			return false
		},
		/* 29 tagName <- <(_ <TAG_NAME> Action51)> */
		func() bool {
			position363, tokenIndex363, depth363 := position, tokenIndex, depth
			{
				position364 := position
				depth++
				if !_rules[rule_]() {
					goto l363
				}
				{
					position365 := position
					depth++
					{
						position366 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l363
						}
						depth--
						add(ruleTAG_NAME, position366)
					}
					depth--
					add(rulePegText, position365)
				}
				{
					add(ruleAction51, position)
				}
				depth--
				add(ruletagName, position364)
			}
			return true
		l363:
			position, tokenIndex, depth = position363, tokenIndex363, depth363
			return false
		},
		/* 30 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position368, tokenIndex368, depth368 := position, tokenIndex, depth
			{
				position369 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l368
				}
				depth--
				add(ruleCOLUMN_NAME, position369)
			}
			return true
		l368:
			position, tokenIndex, depth = position368, tokenIndex368, depth368
			return false
		},
		/* 31 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 32 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 33 IDENTIFIER <- <(('`' CHAR* '`') / (_ !(KEYWORD KEY) ID_SEGMENT ('.' ID_SEGMENT)*))> */
		func() bool {
			position372, tokenIndex372, depth372 := position, tokenIndex, depth
			{
				position373 := position
				depth++
				{
					position374, tokenIndex374, depth374 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l375
					}
					position++
				l376:
					{
						position377, tokenIndex377, depth377 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l377
						}
						goto l376
					l377:
						position, tokenIndex, depth = position377, tokenIndex377, depth377
					}
					if buffer[position] != rune('`') {
						goto l375
					}
					position++
					goto l374
				l375:
					position, tokenIndex, depth = position374, tokenIndex374, depth374
					if !_rules[rule_]() {
						goto l372
					}
					{
						position378, tokenIndex378, depth378 := position, tokenIndex, depth
						{
							position379 := position
							depth++
							{
								position380, tokenIndex380, depth380 := position, tokenIndex, depth
								{
									position382, tokenIndex382, depth382 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l383
									}
									position++
									goto l382
								l383:
									position, tokenIndex, depth = position382, tokenIndex382, depth382
									if buffer[position] != rune('A') {
										goto l381
									}
									position++
								}
							l382:
								{
									position384, tokenIndex384, depth384 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l385
									}
									position++
									goto l384
								l385:
									position, tokenIndex, depth = position384, tokenIndex384, depth384
									if buffer[position] != rune('L') {
										goto l381
									}
									position++
								}
							l384:
								{
									position386, tokenIndex386, depth386 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l387
									}
									position++
									goto l386
								l387:
									position, tokenIndex, depth = position386, tokenIndex386, depth386
									if buffer[position] != rune('L') {
										goto l381
									}
									position++
								}
							l386:
								goto l380
							l381:
								position, tokenIndex, depth = position380, tokenIndex380, depth380
								{
									position389, tokenIndex389, depth389 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l390
									}
									position++
									goto l389
								l390:
									position, tokenIndex, depth = position389, tokenIndex389, depth389
									if buffer[position] != rune('A') {
										goto l388
									}
									position++
								}
							l389:
								{
									position391, tokenIndex391, depth391 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l392
									}
									position++
									goto l391
								l392:
									position, tokenIndex, depth = position391, tokenIndex391, depth391
									if buffer[position] != rune('N') {
										goto l388
									}
									position++
								}
							l391:
								{
									position393, tokenIndex393, depth393 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l394
									}
									position++
									goto l393
								l394:
									position, tokenIndex, depth = position393, tokenIndex393, depth393
									if buffer[position] != rune('D') {
										goto l388
									}
									position++
								}
							l393:
								goto l380
							l388:
								position, tokenIndex, depth = position380, tokenIndex380, depth380
								{
									position396, tokenIndex396, depth396 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l397
									}
									position++
									goto l396
								l397:
									position, tokenIndex, depth = position396, tokenIndex396, depth396
									if buffer[position] != rune('M') {
										goto l395
									}
									position++
								}
							l396:
								{
									position398, tokenIndex398, depth398 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l399
									}
									position++
									goto l398
								l399:
									position, tokenIndex, depth = position398, tokenIndex398, depth398
									if buffer[position] != rune('A') {
										goto l395
									}
									position++
								}
							l398:
								{
									position400, tokenIndex400, depth400 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l401
									}
									position++
									goto l400
								l401:
									position, tokenIndex, depth = position400, tokenIndex400, depth400
									if buffer[position] != rune('T') {
										goto l395
									}
									position++
								}
							l400:
								{
									position402, tokenIndex402, depth402 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l403
									}
									position++
									goto l402
								l403:
									position, tokenIndex, depth = position402, tokenIndex402, depth402
									if buffer[position] != rune('C') {
										goto l395
									}
									position++
								}
							l402:
								{
									position404, tokenIndex404, depth404 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l405
									}
									position++
									goto l404
								l405:
									position, tokenIndex, depth = position404, tokenIndex404, depth404
									if buffer[position] != rune('H') {
										goto l395
									}
									position++
								}
							l404:
								goto l380
							l395:
								position, tokenIndex, depth = position380, tokenIndex380, depth380
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
										goto l406
									}
									position++
								}
							l407:
								{
									position409, tokenIndex409, depth409 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l410
									}
									position++
									goto l409
								l410:
									position, tokenIndex, depth = position409, tokenIndex409, depth409
									if buffer[position] != rune('E') {
										goto l406
									}
									position++
								}
							l409:
								{
									position411, tokenIndex411, depth411 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l412
									}
									position++
									goto l411
								l412:
									position, tokenIndex, depth = position411, tokenIndex411, depth411
									if buffer[position] != rune('L') {
										goto l406
									}
									position++
								}
							l411:
								{
									position413, tokenIndex413, depth413 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l414
									}
									position++
									goto l413
								l414:
									position, tokenIndex, depth = position413, tokenIndex413, depth413
									if buffer[position] != rune('E') {
										goto l406
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
										goto l406
									}
									position++
								}
							l415:
								{
									position417, tokenIndex417, depth417 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l418
									}
									position++
									goto l417
								l418:
									position, tokenIndex, depth = position417, tokenIndex417, depth417
									if buffer[position] != rune('T') {
										goto l406
									}
									position++
								}
							l417:
								goto l380
							l406:
								position, tokenIndex, depth = position380, tokenIndex380, depth380
								{
									switch buffer[position] {
									case 'M', 'm':
										{
											position420, tokenIndex420, depth420 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l421
											}
											position++
											goto l420
										l421:
											position, tokenIndex, depth = position420, tokenIndex420, depth420
											if buffer[position] != rune('M') {
												goto l378
											}
											position++
										}
									l420:
										{
											position422, tokenIndex422, depth422 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l423
											}
											position++
											goto l422
										l423:
											position, tokenIndex, depth = position422, tokenIndex422, depth422
											if buffer[position] != rune('E') {
												goto l378
											}
											position++
										}
									l422:
										{
											position424, tokenIndex424, depth424 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l425
											}
											position++
											goto l424
										l425:
											position, tokenIndex, depth = position424, tokenIndex424, depth424
											if buffer[position] != rune('T') {
												goto l378
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
												goto l378
											}
											position++
										}
									l426:
										{
											position428, tokenIndex428, depth428 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l429
											}
											position++
											goto l428
										l429:
											position, tokenIndex, depth = position428, tokenIndex428, depth428
											if buffer[position] != rune('I') {
												goto l378
											}
											position++
										}
									l428:
										{
											position430, tokenIndex430, depth430 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l431
											}
											position++
											goto l430
										l431:
											position, tokenIndex, depth = position430, tokenIndex430, depth430
											if buffer[position] != rune('C') {
												goto l378
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
												goto l378
											}
											position++
										}
									l432:
										break
									case 'W', 'w':
										{
											position434, tokenIndex434, depth434 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l435
											}
											position++
											goto l434
										l435:
											position, tokenIndex, depth = position434, tokenIndex434, depth434
											if buffer[position] != rune('W') {
												goto l378
											}
											position++
										}
									l434:
										{
											position436, tokenIndex436, depth436 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l437
											}
											position++
											goto l436
										l437:
											position, tokenIndex, depth = position436, tokenIndex436, depth436
											if buffer[position] != rune('H') {
												goto l378
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
												goto l378
											}
											position++
										}
									l438:
										{
											position440, tokenIndex440, depth440 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l441
											}
											position++
											goto l440
										l441:
											position, tokenIndex, depth = position440, tokenIndex440, depth440
											if buffer[position] != rune('R') {
												goto l378
											}
											position++
										}
									l440:
										{
											position442, tokenIndex442, depth442 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l443
											}
											position++
											goto l442
										l443:
											position, tokenIndex, depth = position442, tokenIndex442, depth442
											if buffer[position] != rune('E') {
												goto l378
											}
											position++
										}
									l442:
										break
									case 'O', 'o':
										{
											position444, tokenIndex444, depth444 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l445
											}
											position++
											goto l444
										l445:
											position, tokenIndex, depth = position444, tokenIndex444, depth444
											if buffer[position] != rune('O') {
												goto l378
											}
											position++
										}
									l444:
										{
											position446, tokenIndex446, depth446 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l447
											}
											position++
											goto l446
										l447:
											position, tokenIndex, depth = position446, tokenIndex446, depth446
											if buffer[position] != rune('R') {
												goto l378
											}
											position++
										}
									l446:
										break
									case 'N', 'n':
										{
											position448, tokenIndex448, depth448 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l449
											}
											position++
											goto l448
										l449:
											position, tokenIndex, depth = position448, tokenIndex448, depth448
											if buffer[position] != rune('N') {
												goto l378
											}
											position++
										}
									l448:
										{
											position450, tokenIndex450, depth450 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l451
											}
											position++
											goto l450
										l451:
											position, tokenIndex, depth = position450, tokenIndex450, depth450
											if buffer[position] != rune('O') {
												goto l378
											}
											position++
										}
									l450:
										{
											position452, tokenIndex452, depth452 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l453
											}
											position++
											goto l452
										l453:
											position, tokenIndex, depth = position452, tokenIndex452, depth452
											if buffer[position] != rune('T') {
												goto l378
											}
											position++
										}
									l452:
										break
									case 'I', 'i':
										{
											position454, tokenIndex454, depth454 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l455
											}
											position++
											goto l454
										l455:
											position, tokenIndex, depth = position454, tokenIndex454, depth454
											if buffer[position] != rune('I') {
												goto l378
											}
											position++
										}
									l454:
										{
											position456, tokenIndex456, depth456 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l457
											}
											position++
											goto l456
										l457:
											position, tokenIndex, depth = position456, tokenIndex456, depth456
											if buffer[position] != rune('N') {
												goto l378
											}
											position++
										}
									l456:
										break
									case 'C', 'c':
										{
											position458, tokenIndex458, depth458 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l459
											}
											position++
											goto l458
										l459:
											position, tokenIndex, depth = position458, tokenIndex458, depth458
											if buffer[position] != rune('C') {
												goto l378
											}
											position++
										}
									l458:
										{
											position460, tokenIndex460, depth460 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l461
											}
											position++
											goto l460
										l461:
											position, tokenIndex, depth = position460, tokenIndex460, depth460
											if buffer[position] != rune('O') {
												goto l378
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
												goto l378
											}
											position++
										}
									l462:
										{
											position464, tokenIndex464, depth464 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l465
											}
											position++
											goto l464
										l465:
											position, tokenIndex, depth = position464, tokenIndex464, depth464
											if buffer[position] != rune('L') {
												goto l378
											}
											position++
										}
									l464:
										{
											position466, tokenIndex466, depth466 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l467
											}
											position++
											goto l466
										l467:
											position, tokenIndex, depth = position466, tokenIndex466, depth466
											if buffer[position] != rune('A') {
												goto l378
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
												goto l378
											}
											position++
										}
									l468:
										{
											position470, tokenIndex470, depth470 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l471
											}
											position++
											goto l470
										l471:
											position, tokenIndex, depth = position470, tokenIndex470, depth470
											if buffer[position] != rune('S') {
												goto l378
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
												goto l378
											}
											position++
										}
									l472:
										break
									case 'G', 'g':
										{
											position474, tokenIndex474, depth474 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l475
											}
											position++
											goto l474
										l475:
											position, tokenIndex, depth = position474, tokenIndex474, depth474
											if buffer[position] != rune('G') {
												goto l378
											}
											position++
										}
									l474:
										{
											position476, tokenIndex476, depth476 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l477
											}
											position++
											goto l476
										l477:
											position, tokenIndex, depth = position476, tokenIndex476, depth476
											if buffer[position] != rune('R') {
												goto l378
											}
											position++
										}
									l476:
										{
											position478, tokenIndex478, depth478 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l479
											}
											position++
											goto l478
										l479:
											position, tokenIndex, depth = position478, tokenIndex478, depth478
											if buffer[position] != rune('O') {
												goto l378
											}
											position++
										}
									l478:
										{
											position480, tokenIndex480, depth480 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l481
											}
											position++
											goto l480
										l481:
											position, tokenIndex, depth = position480, tokenIndex480, depth480
											if buffer[position] != rune('U') {
												goto l378
											}
											position++
										}
									l480:
										{
											position482, tokenIndex482, depth482 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l483
											}
											position++
											goto l482
										l483:
											position, tokenIndex, depth = position482, tokenIndex482, depth482
											if buffer[position] != rune('P') {
												goto l378
											}
											position++
										}
									l482:
										break
									case 'D', 'd':
										{
											position484, tokenIndex484, depth484 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l485
											}
											position++
											goto l484
										l485:
											position, tokenIndex, depth = position484, tokenIndex484, depth484
											if buffer[position] != rune('D') {
												goto l378
											}
											position++
										}
									l484:
										{
											position486, tokenIndex486, depth486 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l487
											}
											position++
											goto l486
										l487:
											position, tokenIndex, depth = position486, tokenIndex486, depth486
											if buffer[position] != rune('E') {
												goto l378
											}
											position++
										}
									l486:
										{
											position488, tokenIndex488, depth488 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l489
											}
											position++
											goto l488
										l489:
											position, tokenIndex, depth = position488, tokenIndex488, depth488
											if buffer[position] != rune('S') {
												goto l378
											}
											position++
										}
									l488:
										{
											position490, tokenIndex490, depth490 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l491
											}
											position++
											goto l490
										l491:
											position, tokenIndex, depth = position490, tokenIndex490, depth490
											if buffer[position] != rune('C') {
												goto l378
											}
											position++
										}
									l490:
										{
											position492, tokenIndex492, depth492 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l493
											}
											position++
											goto l492
										l493:
											position, tokenIndex, depth = position492, tokenIndex492, depth492
											if buffer[position] != rune('R') {
												goto l378
											}
											position++
										}
									l492:
										{
											position494, tokenIndex494, depth494 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l495
											}
											position++
											goto l494
										l495:
											position, tokenIndex, depth = position494, tokenIndex494, depth494
											if buffer[position] != rune('I') {
												goto l378
											}
											position++
										}
									l494:
										{
											position496, tokenIndex496, depth496 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l497
											}
											position++
											goto l496
										l497:
											position, tokenIndex, depth = position496, tokenIndex496, depth496
											if buffer[position] != rune('B') {
												goto l378
											}
											position++
										}
									l496:
										{
											position498, tokenIndex498, depth498 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l499
											}
											position++
											goto l498
										l499:
											position, tokenIndex, depth = position498, tokenIndex498, depth498
											if buffer[position] != rune('E') {
												goto l378
											}
											position++
										}
									l498:
										break
									case 'B', 'b':
										{
											position500, tokenIndex500, depth500 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l501
											}
											position++
											goto l500
										l501:
											position, tokenIndex, depth = position500, tokenIndex500, depth500
											if buffer[position] != rune('B') {
												goto l378
											}
											position++
										}
									l500:
										{
											position502, tokenIndex502, depth502 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l503
											}
											position++
											goto l502
										l503:
											position, tokenIndex, depth = position502, tokenIndex502, depth502
											if buffer[position] != rune('Y') {
												goto l378
											}
											position++
										}
									l502:
										break
									case 'A', 'a':
										{
											position504, tokenIndex504, depth504 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l505
											}
											position++
											goto l504
										l505:
											position, tokenIndex, depth = position504, tokenIndex504, depth504
											if buffer[position] != rune('A') {
												goto l378
											}
											position++
										}
									l504:
										{
											position506, tokenIndex506, depth506 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l507
											}
											position++
											goto l506
										l507:
											position, tokenIndex, depth = position506, tokenIndex506, depth506
											if buffer[position] != rune('S') {
												goto l378
											}
											position++
										}
									l506:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l378
										}
										break
									}
								}

							}
						l380:
							depth--
							add(ruleKEYWORD, position379)
						}
						if !_rules[ruleKEY]() {
							goto l378
						}
						goto l372
					l378:
						position, tokenIndex, depth = position378, tokenIndex378, depth378
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l372
					}
				l508:
					{
						position509, tokenIndex509, depth509 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l509
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l509
						}
						goto l508
					l509:
						position, tokenIndex, depth = position509, tokenIndex509, depth509
					}
				}
			l374:
				depth--
				add(ruleIDENTIFIER, position373)
			}
			return true
		l372:
			position, tokenIndex, depth = position372, tokenIndex372, depth372
			return false
		},
		/* 34 TIMESTAMP <- <((_ <(NUMBER ([a-z] / [A-Z])*)>) / (_ STRING) / (_ <(('n' / 'N') ('o' / 'O') ('w' / 'W'))>))> */
		nil,
		/* 35 ID_SEGMENT <- <(_ ID_START ID_CONT*)> */
		func() bool {
			position511, tokenIndex511, depth511 := position, tokenIndex, depth
			{
				position512 := position
				depth++
				if !_rules[rule_]() {
					goto l511
				}
				if !_rules[ruleID_START]() {
					goto l511
				}
			l513:
				{
					position514, tokenIndex514, depth514 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l514
					}
					goto l513
				l514:
					position, tokenIndex, depth = position514, tokenIndex514, depth514
				}
				depth--
				add(ruleID_SEGMENT, position512)
			}
			return true
		l511:
			position, tokenIndex, depth = position511, tokenIndex511, depth511
			return false
		},
		/* 36 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position515, tokenIndex515, depth515 := position, tokenIndex, depth
			{
				position516 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l515
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l515
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l515
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position516)
			}
			return true
		l515:
			position, tokenIndex, depth = position515, tokenIndex515, depth515
			return false
		},
		/* 37 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position518, tokenIndex518, depth518 := position, tokenIndex, depth
			{
				position519 := position
				depth++
				{
					position520, tokenIndex520, depth520 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l521
					}
					goto l520
				l521:
					position, tokenIndex, depth = position520, tokenIndex520, depth520
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l518
					}
					position++
				}
			l520:
				depth--
				add(ruleID_CONT, position519)
			}
			return true
		l518:
			position, tokenIndex, depth = position518, tokenIndex518, depth518
			return false
		},
		/* 38 PROPERTY_KEY <- <(((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> KEY _ (('b' / 'B') ('y' / 'Y')))) | (&('R' | 'r') <(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))>) | (&('T' | 't') <(('t' / 'T') ('o' / 'O'))>) | (&('F' | 'f') <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))>)) KEY)> */
		func() bool {
			position522, tokenIndex522, depth522 := position, tokenIndex, depth
			{
				position523 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position525 := position
							depth++
							{
								position526, tokenIndex526, depth526 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l527
								}
								position++
								goto l526
							l527:
								position, tokenIndex, depth = position526, tokenIndex526, depth526
								if buffer[position] != rune('S') {
									goto l522
								}
								position++
							}
						l526:
							{
								position528, tokenIndex528, depth528 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l529
								}
								position++
								goto l528
							l529:
								position, tokenIndex, depth = position528, tokenIndex528, depth528
								if buffer[position] != rune('A') {
									goto l522
								}
								position++
							}
						l528:
							{
								position530, tokenIndex530, depth530 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l531
								}
								position++
								goto l530
							l531:
								position, tokenIndex, depth = position530, tokenIndex530, depth530
								if buffer[position] != rune('M') {
									goto l522
								}
								position++
							}
						l530:
							{
								position532, tokenIndex532, depth532 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l533
								}
								position++
								goto l532
							l533:
								position, tokenIndex, depth = position532, tokenIndex532, depth532
								if buffer[position] != rune('P') {
									goto l522
								}
								position++
							}
						l532:
							{
								position534, tokenIndex534, depth534 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l535
								}
								position++
								goto l534
							l535:
								position, tokenIndex, depth = position534, tokenIndex534, depth534
								if buffer[position] != rune('L') {
									goto l522
								}
								position++
							}
						l534:
							{
								position536, tokenIndex536, depth536 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l537
								}
								position++
								goto l536
							l537:
								position, tokenIndex, depth = position536, tokenIndex536, depth536
								if buffer[position] != rune('E') {
									goto l522
								}
								position++
							}
						l536:
							depth--
							add(rulePegText, position525)
						}
						if !_rules[ruleKEY]() {
							goto l522
						}
						if !_rules[rule_]() {
							goto l522
						}
						{
							position538, tokenIndex538, depth538 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l539
							}
							position++
							goto l538
						l539:
							position, tokenIndex, depth = position538, tokenIndex538, depth538
							if buffer[position] != rune('B') {
								goto l522
							}
							position++
						}
					l538:
						{
							position540, tokenIndex540, depth540 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l541
							}
							position++
							goto l540
						l541:
							position, tokenIndex, depth = position540, tokenIndex540, depth540
							if buffer[position] != rune('Y') {
								goto l522
							}
							position++
						}
					l540:
						break
					case 'R', 'r':
						{
							position542 := position
							depth++
							{
								position543, tokenIndex543, depth543 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l544
								}
								position++
								goto l543
							l544:
								position, tokenIndex, depth = position543, tokenIndex543, depth543
								if buffer[position] != rune('R') {
									goto l522
								}
								position++
							}
						l543:
							{
								position545, tokenIndex545, depth545 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l546
								}
								position++
								goto l545
							l546:
								position, tokenIndex, depth = position545, tokenIndex545, depth545
								if buffer[position] != rune('E') {
									goto l522
								}
								position++
							}
						l545:
							{
								position547, tokenIndex547, depth547 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l548
								}
								position++
								goto l547
							l548:
								position, tokenIndex, depth = position547, tokenIndex547, depth547
								if buffer[position] != rune('S') {
									goto l522
								}
								position++
							}
						l547:
							{
								position549, tokenIndex549, depth549 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l550
								}
								position++
								goto l549
							l550:
								position, tokenIndex, depth = position549, tokenIndex549, depth549
								if buffer[position] != rune('O') {
									goto l522
								}
								position++
							}
						l549:
							{
								position551, tokenIndex551, depth551 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l552
								}
								position++
								goto l551
							l552:
								position, tokenIndex, depth = position551, tokenIndex551, depth551
								if buffer[position] != rune('L') {
									goto l522
								}
								position++
							}
						l551:
							{
								position553, tokenIndex553, depth553 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l554
								}
								position++
								goto l553
							l554:
								position, tokenIndex, depth = position553, tokenIndex553, depth553
								if buffer[position] != rune('U') {
									goto l522
								}
								position++
							}
						l553:
							{
								position555, tokenIndex555, depth555 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l556
								}
								position++
								goto l555
							l556:
								position, tokenIndex, depth = position555, tokenIndex555, depth555
								if buffer[position] != rune('T') {
									goto l522
								}
								position++
							}
						l555:
							{
								position557, tokenIndex557, depth557 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l558
								}
								position++
								goto l557
							l558:
								position, tokenIndex, depth = position557, tokenIndex557, depth557
								if buffer[position] != rune('I') {
									goto l522
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
									goto l522
								}
								position++
							}
						l559:
							{
								position561, tokenIndex561, depth561 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l562
								}
								position++
								goto l561
							l562:
								position, tokenIndex, depth = position561, tokenIndex561, depth561
								if buffer[position] != rune('N') {
									goto l522
								}
								position++
							}
						l561:
							depth--
							add(rulePegText, position542)
						}
						break
					case 'T', 't':
						{
							position563 := position
							depth++
							{
								position564, tokenIndex564, depth564 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l565
								}
								position++
								goto l564
							l565:
								position, tokenIndex, depth = position564, tokenIndex564, depth564
								if buffer[position] != rune('T') {
									goto l522
								}
								position++
							}
						l564:
							{
								position566, tokenIndex566, depth566 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l567
								}
								position++
								goto l566
							l567:
								position, tokenIndex, depth = position566, tokenIndex566, depth566
								if buffer[position] != rune('O') {
									goto l522
								}
								position++
							}
						l566:
							depth--
							add(rulePegText, position563)
						}
						break
					default:
						{
							position568 := position
							depth++
							{
								position569, tokenIndex569, depth569 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l570
								}
								position++
								goto l569
							l570:
								position, tokenIndex, depth = position569, tokenIndex569, depth569
								if buffer[position] != rune('F') {
									goto l522
								}
								position++
							}
						l569:
							{
								position571, tokenIndex571, depth571 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l572
								}
								position++
								goto l571
							l572:
								position, tokenIndex, depth = position571, tokenIndex571, depth571
								if buffer[position] != rune('R') {
									goto l522
								}
								position++
							}
						l571:
							{
								position573, tokenIndex573, depth573 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l574
								}
								position++
								goto l573
							l574:
								position, tokenIndex, depth = position573, tokenIndex573, depth573
								if buffer[position] != rune('O') {
									goto l522
								}
								position++
							}
						l573:
							{
								position575, tokenIndex575, depth575 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l576
								}
								position++
								goto l575
							l576:
								position, tokenIndex, depth = position575, tokenIndex575, depth575
								if buffer[position] != rune('M') {
									goto l522
								}
								position++
							}
						l575:
							depth--
							add(rulePegText, position568)
						}
						break
					}
				}

				if !_rules[ruleKEY]() {
					goto l522
				}
				depth--
				add(rulePROPERTY_KEY, position523)
			}
			return true
		l522:
			position, tokenIndex, depth = position522, tokenIndex522, depth522
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
			position587, tokenIndex587, depth587 := position, tokenIndex, depth
			{
				position588 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l587
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position588)
			}
			return true
		l587:
			position, tokenIndex, depth = position587, tokenIndex587, depth587
			return false
		},
		/* 50 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position589, tokenIndex589, depth589 := position, tokenIndex, depth
			{
				position590 := position
				depth++
				if buffer[position] != rune('"') {
					goto l589
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position590)
			}
			return true
		l589:
			position, tokenIndex, depth = position589, tokenIndex589, depth589
			return false
		},
		/* 51 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> QUOTE_SINGLE) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> QUOTE_DOUBLE))> */
		func() bool {
			position591, tokenIndex591, depth591 := position, tokenIndex, depth
			{
				position592 := position
				depth++
				{
					position593, tokenIndex593, depth593 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l594
					}
					{
						position595 := position
						depth++
					l596:
						{
							position597, tokenIndex597, depth597 := position, tokenIndex, depth
							{
								position598, tokenIndex598, depth598 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l598
								}
								goto l597
							l598:
								position, tokenIndex, depth = position598, tokenIndex598, depth598
							}
							if !_rules[ruleCHAR]() {
								goto l597
							}
							goto l596
						l597:
							position, tokenIndex, depth = position597, tokenIndex597, depth597
						}
						depth--
						add(rulePegText, position595)
					}
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l594
					}
					goto l593
				l594:
					position, tokenIndex, depth = position593, tokenIndex593, depth593
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l591
					}
					{
						position599 := position
						depth++
					l600:
						{
							position601, tokenIndex601, depth601 := position, tokenIndex, depth
							{
								position602, tokenIndex602, depth602 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l602
								}
								goto l601
							l602:
								position, tokenIndex, depth = position602, tokenIndex602, depth602
							}
							if !_rules[ruleCHAR]() {
								goto l601
							}
							goto l600
						l601:
							position, tokenIndex, depth = position601, tokenIndex601, depth601
						}
						depth--
						add(rulePegText, position599)
					}
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l591
					}
				}
			l593:
				depth--
				add(ruleSTRING, position592)
			}
			return true
		l591:
			position, tokenIndex, depth = position591, tokenIndex591, depth591
			return false
		},
		/* 52 CHAR <- <(('\\' ((&('"') QUOTE_DOUBLE) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position603, tokenIndex603, depth603 := position, tokenIndex, depth
			{
				position604 := position
				depth++
				{
					position605, tokenIndex605, depth605 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l606
					}
					position++
					{
						switch buffer[position] {
						case '"':
							if !_rules[ruleQUOTE_DOUBLE]() {
								goto l606
							}
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l606
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l606
							}
							break
						}
					}

					goto l605
				l606:
					position, tokenIndex, depth = position605, tokenIndex605, depth605
					{
						position608, tokenIndex608, depth608 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l608
						}
						goto l603
					l608:
						position, tokenIndex, depth = position608, tokenIndex608, depth608
					}
					if !matchDot() {
						goto l603
					}
				}
			l605:
				depth--
				add(ruleCHAR, position604)
			}
			return true
		l603:
			position, tokenIndex, depth = position603, tokenIndex603, depth603
			return false
		},
		/* 53 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position609, tokenIndex609, depth609 := position, tokenIndex, depth
			{
				position610 := position
				depth++
				{
					position611, tokenIndex611, depth611 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l612
					}
					position++
					goto l611
				l612:
					position, tokenIndex, depth = position611, tokenIndex611, depth611
					if buffer[position] != rune('\\') {
						goto l609
					}
					position++
				}
			l611:
				depth--
				add(ruleESCAPE_CLASS, position610)
			}
			return true
		l609:
			position, tokenIndex, depth = position609, tokenIndex609, depth609
			return false
		},
		/* 54 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position613, tokenIndex613, depth613 := position, tokenIndex, depth
			{
				position614 := position
				depth++
				{
					position615 := position
					depth++
					{
						position616, tokenIndex616, depth616 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l616
						}
						position++
						goto l617
					l616:
						position, tokenIndex, depth = position616, tokenIndex616, depth616
					}
				l617:
					{
						position618 := position
						depth++
						{
							position619, tokenIndex619, depth619 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l620
							}
							position++
							goto l619
						l620:
							position, tokenIndex, depth = position619, tokenIndex619, depth619
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l613
							}
							position++
						l621:
							{
								position622, tokenIndex622, depth622 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l622
								}
								position++
								goto l621
							l622:
								position, tokenIndex, depth = position622, tokenIndex622, depth622
							}
						}
					l619:
						depth--
						add(ruleNUMBER_NATURAL, position618)
					}
					depth--
					add(ruleNUMBER_INTEGER, position615)
				}
				{
					position623, tokenIndex623, depth623 := position, tokenIndex, depth
					{
						position625 := position
						depth++
						if buffer[position] != rune('.') {
							goto l623
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l623
						}
						position++
					l626:
						{
							position627, tokenIndex627, depth627 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l627
							}
							position++
							goto l626
						l627:
							position, tokenIndex, depth = position627, tokenIndex627, depth627
						}
						depth--
						add(ruleNUMBER_FRACTION, position625)
					}
					goto l624
				l623:
					position, tokenIndex, depth = position623, tokenIndex623, depth623
				}
			l624:
				{
					position628, tokenIndex628, depth628 := position, tokenIndex, depth
					{
						position630 := position
						depth++
						{
							position631, tokenIndex631, depth631 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l632
							}
							position++
							goto l631
						l632:
							position, tokenIndex, depth = position631, tokenIndex631, depth631
							if buffer[position] != rune('E') {
								goto l628
							}
							position++
						}
					l631:
						{
							position633, tokenIndex633, depth633 := position, tokenIndex, depth
							{
								position635, tokenIndex635, depth635 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l636
								}
								position++
								goto l635
							l636:
								position, tokenIndex, depth = position635, tokenIndex635, depth635
								if buffer[position] != rune('-') {
									goto l633
								}
								position++
							}
						l635:
							goto l634
						l633:
							position, tokenIndex, depth = position633, tokenIndex633, depth633
						}
					l634:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l628
						}
						position++
					l637:
						{
							position638, tokenIndex638, depth638 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l638
							}
							position++
							goto l637
						l638:
							position, tokenIndex, depth = position638, tokenIndex638, depth638
						}
						depth--
						add(ruleNUMBER_EXP, position630)
					}
					goto l629
				l628:
					position, tokenIndex, depth = position628, tokenIndex628, depth628
				}
			l629:
				depth--
				add(ruleNUMBER, position614)
			}
			return true
		l613:
			position, tokenIndex, depth = position613, tokenIndex613, depth613
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
			position644, tokenIndex644, depth644 := position, tokenIndex, depth
			{
				position645 := position
				depth++
				if buffer[position] != rune('(') {
					goto l644
				}
				position++
				depth--
				add(rulePAREN_OPEN, position645)
			}
			return true
		l644:
			position, tokenIndex, depth = position644, tokenIndex644, depth644
			return false
		},
		/* 61 PAREN_CLOSE <- <')'> */
		func() bool {
			position646, tokenIndex646, depth646 := position, tokenIndex, depth
			{
				position647 := position
				depth++
				if buffer[position] != rune(')') {
					goto l646
				}
				position++
				depth--
				add(rulePAREN_CLOSE, position647)
			}
			return true
		l646:
			position, tokenIndex, depth = position646, tokenIndex646, depth646
			return false
		},
		/* 62 COMMA <- <','> */
		func() bool {
			position648, tokenIndex648, depth648 := position, tokenIndex, depth
			{
				position649 := position
				depth++
				if buffer[position] != rune(',') {
					goto l648
				}
				position++
				depth--
				add(ruleCOMMA, position649)
			}
			return true
		l648:
			position, tokenIndex, depth = position648, tokenIndex648, depth648
			return false
		},
		/* 63 _ <- <((&('/') COMMENT_BLOCK) | (&('-') COMMENT_TRAIL) | (&('\t' | '\n' | ' ') SPACE))*> */
		func() bool {
			{
				position651 := position
				depth++
			l652:
				{
					position653, tokenIndex653, depth653 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '/':
							{
								position655 := position
								depth++
								if buffer[position] != rune('/') {
									goto l653
								}
								position++
								if buffer[position] != rune('*') {
									goto l653
								}
								position++
							l656:
								{
									position657, tokenIndex657, depth657 := position, tokenIndex, depth
									{
										position658, tokenIndex658, depth658 := position, tokenIndex, depth
										if buffer[position] != rune('*') {
											goto l658
										}
										position++
										if buffer[position] != rune('/') {
											goto l658
										}
										position++
										goto l657
									l658:
										position, tokenIndex, depth = position658, tokenIndex658, depth658
									}
									if !matchDot() {
										goto l657
									}
									goto l656
								l657:
									position, tokenIndex, depth = position657, tokenIndex657, depth657
								}
								if buffer[position] != rune('*') {
									goto l653
								}
								position++
								if buffer[position] != rune('/') {
									goto l653
								}
								position++
								depth--
								add(ruleCOMMENT_BLOCK, position655)
							}
							break
						case '-':
							{
								position659 := position
								depth++
								if buffer[position] != rune('-') {
									goto l653
								}
								position++
								if buffer[position] != rune('-') {
									goto l653
								}
								position++
							l660:
								{
									position661, tokenIndex661, depth661 := position, tokenIndex, depth
									{
										position662, tokenIndex662, depth662 := position, tokenIndex, depth
										if buffer[position] != rune('\n') {
											goto l662
										}
										position++
										goto l661
									l662:
										position, tokenIndex, depth = position662, tokenIndex662, depth662
									}
									if !matchDot() {
										goto l661
									}
									goto l660
								l661:
									position, tokenIndex, depth = position661, tokenIndex661, depth661
								}
								depth--
								add(ruleCOMMENT_TRAIL, position659)
							}
							break
						default:
							{
								position663 := position
								depth++
								{
									switch buffer[position] {
									case '\t':
										if buffer[position] != rune('\t') {
											goto l653
										}
										position++
										break
									case '\n':
										if buffer[position] != rune('\n') {
											goto l653
										}
										position++
										break
									default:
										if buffer[position] != rune(' ') {
											goto l653
										}
										position++
										break
									}
								}

								depth--
								add(ruleSPACE, position663)
							}
							break
						}
					}

					goto l652
				l653:
					position, tokenIndex, depth = position653, tokenIndex653, depth653
				}
				depth--
				add(rule_, position651)
			}
			return true
		},
		/* 64 COMMENT_TRAIL <- <('-' '-' (!'\n' .)*)> */
		nil,
		/* 65 COMMENT_BLOCK <- <('/' '*' (!('*' '/') .)* ('*' '/'))> */
		nil,
		/* 66 KEY <- <!ID_CONT> */
		func() bool {
			position667, tokenIndex667, depth667 := position, tokenIndex, depth
			{
				position668 := position
				depth++
				{
					position669, tokenIndex669, depth669 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l669
					}
					goto l667
				l669:
					position, tokenIndex, depth = position669, tokenIndex669, depth669
				}
				depth--
				add(ruleKEY, position668)
			}
			return true
		l667:
			position, tokenIndex, depth = position667, tokenIndex667, depth667
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
		/* 76 Action6 <- <{ p.addIndexClause() }> */
		nil,
		/* 77 Action7 <- <{ p.noIndexClause() }> */
		nil,
		/* 78 Action8 <- <{ p.makeDescribe() }> */
		nil,
		/* 79 Action9 <- <{ p.addEvaluationContext() }> */
		nil,
		/* 80 Action10 <- <{ p.addPropertyKey(buffer[begin:end])   }> */
		nil,
		/* 81 Action11 <- <{ p.addPropertyValue(buffer[begin:end]) }> */
		nil,
		/* 82 Action12 <- <{ p.insertPropertyKeyValue() }> */
		nil,
		/* 83 Action13 <- <{ p.checkPropertyClause() }> */
		nil,
		/* 84 Action14 <- <{ p.addNullPredicate() }> */
		nil,
		/* 85 Action15 <- <{ p.addExpressionList() }> */
		nil,
		/* 86 Action16 <- <{ p.appendExpression() }> */
		nil,
		/* 87 Action17 <- <{ p.appendExpression() }> */
		nil,
		/* 88 Action18 <- <{ p.addOperatorLiteral("+") }> */
		nil,
		/* 89 Action19 <- <{ p.addOperatorLiteral("-") }> */
		nil,
		/* 90 Action20 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 91 Action21 <- <{ p.addOperatorLiteral("/") }> */
		nil,
		/* 92 Action22 <- <{ p.addOperatorLiteral("*") }> */
		nil,
		/* 93 Action23 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 94 Action24 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 95 Action25 <- <{p.addExpressionList()}> */
		nil,
		/* 96 Action26 <- <{
		   p.addExpressionList()
		   p.addGroupBy()
		 }> */
		nil,
		/* 97 Action27 <- <{
		   p.addPipeExpression()
		 }> */
		nil,
		/* 98 Action28 <- <{ p.addDurationNode(text) }> */
		nil,
		/* 99 Action29 <- <{ p.addNumberNode(buffer[begin:end]) }> */
		nil,
		/* 100 Action30 <- <{ p.addStringNode(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 101 Action31 <- <{ p.addGroupBy() }> */
		nil,
		/* 102 Action32 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 103 Action33 <- <{
		   p.addFunctionInvocation()
		 }> */
		nil,
		/* 104 Action34 <- <{
		   p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 105 Action35 <- <{ p.addNullPredicate() }> */
		nil,
		/* 106 Action36 <- <{
		   p.addMetricExpression()
		 }> */
		nil,
		/* 107 Action37 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 108 Action38 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 109 Action39 <- <{
		   p.appendCollapseBy(unescapeLiteral(text))
		 }> */
		nil,
		/* 110 Action40 <- <{p.appendCollapseBy(unescapeLiteral(text))}> */
		nil,
		/* 111 Action41 <- <{ p.addOrPredicate() }> */
		nil,
		/* 112 Action42 <- <{ p.addAndPredicate() }> */
		nil,
		/* 113 Action43 <- <{ p.addNotPredicate() }> */
		nil,
		/* 114 Action44 <- <{
		   p.addLiteralMatcher()
		 }> */
		nil,
		/* 115 Action45 <- <{
		   p.addLiteralMatcher()
		   p.addNotPredicate()
		 }> */
		nil,
		/* 116 Action46 <- <{
		   p.addRegexMatcher()
		 }> */
		nil,
		/* 117 Action47 <- <{
		   p.addListMatcher()
		 }> */
		nil,
		/* 118 Action48 <- <{
		  p.addStringLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 119 Action49 <- <{ p.addLiteralList() }> */
		nil,
		/* 120 Action50 <- <{
		  p.appendLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 121 Action51 <- <{ p.addTagLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
	}
	p.rules = _rules
}
