package parser

import (
	"github.com/square/metrics/query/command"
	"fmt"
	"math"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

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
	ruleadd_one_pipe
	ruleadd_pipe
	ruleexpression_atom
	ruleexpression_atom_raw
	ruleexpression_annotation_required
	ruleexpression_annotation
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

	rulePre
	ruleIn
	ruleSuf
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
	"add_one_pipe",
	"add_pipe",
	"expression_atom",
	"expression_atom_raw",
	"expression_annotation_required",
	"expression_annotation",
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

func (node *node32) Print(buffer string) {
	node.print(0, buffer)
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
		for i := range states {
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
							write(token32{pegRule: ruleIn, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre, begin: a.begin, end: b.begin}, true)
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
					write(token32{pegRule: ruleSuf, begin: b.end, end: a.end}, true)
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
	for i := range tokens {
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
	nodeStack []any

	// user errors accumulated during the AST traversal.
	// a non-empty list at the finish time means an invalid query is provided.
	errors []SyntaxError

	// errorContext describes contexts used to build error messages
	fixedContext string
	errorContext []string

	// programming errors accumulated during the AST traversal.
	// a non-empty list at the finish time implies a programming error.

	// final result
	command command.Command

	Buffer string
	buffer []rune
	rules  [126]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	Pretty bool
	tokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
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
	p   *Parser
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
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
			p.pushString(unescapeLiteral(text))
		case ruleAction6:
			p.makeDescribe()
		case ruleAction7:
			p.addEvaluationContext()
		case ruleAction8:
			p.addPropertyKey(text)
		case ruleAction9:

			p.addPropertyValue(text)
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
			p.pushString(unescapeLiteral(text))
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
			p.addNumberNode(text)
		case ruleAction28:
			p.addStringNode(unescapeLiteral(text))
		case ruleAction29:
			p.addAnnotationExpression(text)
		case ruleAction30:
			p.addGroupBy()
		case ruleAction31:
			p.pushString(unescapeLiteral(text))
		case ruleAction32:
			p.addFunctionInvocation()
		case ruleAction33:
			p.pushString(unescapeLiteral(text))
		case ruleAction34:
			p.addNullPredicate()
		case ruleAction35:
			p.addMetricExpression()
		case ruleAction36:
			p.appendGroupBy(unescapeLiteral(text))
		case ruleAction37:
			p.appendGroupBy(unescapeLiteral(text))
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
		case ruleAction45:
			p.addNotPredicate()
		case ruleAction46:
			p.addRegexMatcher()
		case ruleAction47:
			p.addListMatcher()
		case ruleAction48:
			p.pushString(unescapeLiteral(text))
		case ruleAction49:
			p.addLiteralList()
		case ruleAction50:
			p.appendLiteral(unescapeLiteral(text))
		case ruleAction51:
			p.addTagLiteral(unescapeLiteral(text))

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *Parser) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
		p.buffer = append(p.buffer, endSymbol)
	}

	var tree tokenTree = &tokens32{tree: make([]token32, math.MaxInt16)}
	var max token32
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
		return &parseError{p, max}
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
		if begin != position && position > max.end {
			max = token32{rule, begin, position, depth}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
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
						if !(p.setContext("after expression of select statement")) {
							goto l3
						}
						if !_rules[ruleoptionalPredicateClause]() {
							goto l3
						}
						if !(p.setContext("")) {
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
								{
									position23, tokenIndex23, depth23 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l24
									}
									{
										position25 := position
										depth++
										{
											switch buffer[position] {
											case 'S', 's':
												{
													position27 := position
													depth++
													{
														position28, tokenIndex28, depth28 := position, tokenIndex, depth
														if buffer[position] != rune('s') {
															goto l29
														}
														position++
														goto l28
													l29:
														position, tokenIndex, depth = position28, tokenIndex28, depth28
														if buffer[position] != rune('S') {
															goto l24
														}
														position++
													}
												l28:
													{
														position30, tokenIndex30, depth30 := position, tokenIndex, depth
														if buffer[position] != rune('a') {
															goto l31
														}
														position++
														goto l30
													l31:
														position, tokenIndex, depth = position30, tokenIndex30, depth30
														if buffer[position] != rune('A') {
															goto l24
														}
														position++
													}
												l30:
													{
														position32, tokenIndex32, depth32 := position, tokenIndex, depth
														if buffer[position] != rune('m') {
															goto l33
														}
														position++
														goto l32
													l33:
														position, tokenIndex, depth = position32, tokenIndex32, depth32
														if buffer[position] != rune('M') {
															goto l24
														}
														position++
													}
												l32:
													{
														position34, tokenIndex34, depth34 := position, tokenIndex, depth
														if buffer[position] != rune('p') {
															goto l35
														}
														position++
														goto l34
													l35:
														position, tokenIndex, depth = position34, tokenIndex34, depth34
														if buffer[position] != rune('P') {
															goto l24
														}
														position++
													}
												l34:
													{
														position36, tokenIndex36, depth36 := position, tokenIndex, depth
														if buffer[position] != rune('l') {
															goto l37
														}
														position++
														goto l36
													l37:
														position, tokenIndex, depth = position36, tokenIndex36, depth36
														if buffer[position] != rune('L') {
															goto l24
														}
														position++
													}
												l36:
													{
														position38, tokenIndex38, depth38 := position, tokenIndex, depth
														if buffer[position] != rune('e') {
															goto l39
														}
														position++
														goto l38
													l39:
														position, tokenIndex, depth = position38, tokenIndex38, depth38
														if buffer[position] != rune('E') {
															goto l24
														}
														position++
													}
												l38:
													depth--
													add(rulePegText, position27)
												}
												if !_rules[ruleKEY]() {
													goto l24
												}
												{
													position40, tokenIndex40, depth40 := position, tokenIndex, depth
													if !_rules[rule_]() {
														goto l41
													}
													{
														position42, tokenIndex42, depth42 := position, tokenIndex, depth
														if buffer[position] != rune('b') {
															goto l43
														}
														position++
														goto l42
													l43:
														position, tokenIndex, depth = position42, tokenIndex42, depth42
														if buffer[position] != rune('B') {
															goto l41
														}
														position++
													}
												l42:
													{
														position44, tokenIndex44, depth44 := position, tokenIndex, depth
														if buffer[position] != rune('y') {
															goto l45
														}
														position++
														goto l44
													l45:
														position, tokenIndex, depth = position44, tokenIndex44, depth44
														if buffer[position] != rune('Y') {
															goto l41
														}
														position++
													}
												l44:
													if !_rules[ruleKEY]() {
														goto l41
													}
													goto l40
												l41:
													position, tokenIndex, depth = position40, tokenIndex40, depth40
													if !(p.errorHere(position, `expected keyword "by" to follow keyword "sample"`)) {
														goto l24
													}
												}
											l40:
												break
											case 'R', 'r':
												{
													position46 := position
													depth++
													{
														position47, tokenIndex47, depth47 := position, tokenIndex, depth
														if buffer[position] != rune('r') {
															goto l48
														}
														position++
														goto l47
													l48:
														position, tokenIndex, depth = position47, tokenIndex47, depth47
														if buffer[position] != rune('R') {
															goto l24
														}
														position++
													}
												l47:
													{
														position49, tokenIndex49, depth49 := position, tokenIndex, depth
														if buffer[position] != rune('e') {
															goto l50
														}
														position++
														goto l49
													l50:
														position, tokenIndex, depth = position49, tokenIndex49, depth49
														if buffer[position] != rune('E') {
															goto l24
														}
														position++
													}
												l49:
													{
														position51, tokenIndex51, depth51 := position, tokenIndex, depth
														if buffer[position] != rune('s') {
															goto l52
														}
														position++
														goto l51
													l52:
														position, tokenIndex, depth = position51, tokenIndex51, depth51
														if buffer[position] != rune('S') {
															goto l24
														}
														position++
													}
												l51:
													{
														position53, tokenIndex53, depth53 := position, tokenIndex, depth
														if buffer[position] != rune('o') {
															goto l54
														}
														position++
														goto l53
													l54:
														position, tokenIndex, depth = position53, tokenIndex53, depth53
														if buffer[position] != rune('O') {
															goto l24
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
															goto l24
														}
														position++
													}
												l55:
													{
														position57, tokenIndex57, depth57 := position, tokenIndex, depth
														if buffer[position] != rune('u') {
															goto l58
														}
														position++
														goto l57
													l58:
														position, tokenIndex, depth = position57, tokenIndex57, depth57
														if buffer[position] != rune('U') {
															goto l24
														}
														position++
													}
												l57:
													{
														position59, tokenIndex59, depth59 := position, tokenIndex, depth
														if buffer[position] != rune('t') {
															goto l60
														}
														position++
														goto l59
													l60:
														position, tokenIndex, depth = position59, tokenIndex59, depth59
														if buffer[position] != rune('T') {
															goto l24
														}
														position++
													}
												l59:
													{
														position61, tokenIndex61, depth61 := position, tokenIndex, depth
														if buffer[position] != rune('i') {
															goto l62
														}
														position++
														goto l61
													l62:
														position, tokenIndex, depth = position61, tokenIndex61, depth61
														if buffer[position] != rune('I') {
															goto l24
														}
														position++
													}
												l61:
													{
														position63, tokenIndex63, depth63 := position, tokenIndex, depth
														if buffer[position] != rune('o') {
															goto l64
														}
														position++
														goto l63
													l64:
														position, tokenIndex, depth = position63, tokenIndex63, depth63
														if buffer[position] != rune('O') {
															goto l24
														}
														position++
													}
												l63:
													{
														position65, tokenIndex65, depth65 := position, tokenIndex, depth
														if buffer[position] != rune('n') {
															goto l66
														}
														position++
														goto l65
													l66:
														position, tokenIndex, depth = position65, tokenIndex65, depth65
														if buffer[position] != rune('N') {
															goto l24
														}
														position++
													}
												l65:
													depth--
													add(rulePegText, position46)
												}
												if !_rules[ruleKEY]() {
													goto l24
												}
												break
											case 'T', 't':
												{
													position67 := position
													depth++
													{
														position68, tokenIndex68, depth68 := position, tokenIndex, depth
														if buffer[position] != rune('t') {
															goto l69
														}
														position++
														goto l68
													l69:
														position, tokenIndex, depth = position68, tokenIndex68, depth68
														if buffer[position] != rune('T') {
															goto l24
														}
														position++
													}
												l68:
													{
														position70, tokenIndex70, depth70 := position, tokenIndex, depth
														if buffer[position] != rune('o') {
															goto l71
														}
														position++
														goto l70
													l71:
														position, tokenIndex, depth = position70, tokenIndex70, depth70
														if buffer[position] != rune('O') {
															goto l24
														}
														position++
													}
												l70:
													depth--
													add(rulePegText, position67)
												}
												if !_rules[ruleKEY]() {
													goto l24
												}
												break
											default:
												{
													position72 := position
													depth++
													{
														position73, tokenIndex73, depth73 := position, tokenIndex, depth
														if buffer[position] != rune('f') {
															goto l74
														}
														position++
														goto l73
													l74:
														position, tokenIndex, depth = position73, tokenIndex73, depth73
														if buffer[position] != rune('F') {
															goto l24
														}
														position++
													}
												l73:
													{
														position75, tokenIndex75, depth75 := position, tokenIndex, depth
														if buffer[position] != rune('r') {
															goto l76
														}
														position++
														goto l75
													l76:
														position, tokenIndex, depth = position75, tokenIndex75, depth75
														if buffer[position] != rune('R') {
															goto l24
														}
														position++
													}
												l75:
													{
														position77, tokenIndex77, depth77 := position, tokenIndex, depth
														if buffer[position] != rune('o') {
															goto l78
														}
														position++
														goto l77
													l78:
														position, tokenIndex, depth = position77, tokenIndex77, depth77
														if buffer[position] != rune('O') {
															goto l24
														}
														position++
													}
												l77:
													{
														position79, tokenIndex79, depth79 := position, tokenIndex, depth
														if buffer[position] != rune('m') {
															goto l80
														}
														position++
														goto l79
													l80:
														position, tokenIndex, depth = position79, tokenIndex79, depth79
														if buffer[position] != rune('M') {
															goto l24
														}
														position++
													}
												l79:
													depth--
													add(rulePegText, position72)
												}
												if !_rules[ruleKEY]() {
													goto l24
												}
												break
											}
										}

										depth--
										add(rulePROPERTY_KEY, position25)
									}
									{
										add(ruleAction8, position)
									}
									{
										position82, tokenIndex82, depth82 := position, tokenIndex, depth
										if !_rules[rule_]() {
											goto l83
										}
										{
											position84 := position
											depth++
											{
												position85 := position
												depth++
												{
													position86, tokenIndex86, depth86 := position, tokenIndex, depth
													if !_rules[rule_]() {
														goto l87
													}
													{
														position88 := position
														depth++
														if !_rules[ruleNUMBER]() {
															goto l87
														}
													l89:
														{
															position90, tokenIndex90, depth90 := position, tokenIndex, depth
															{
																position91, tokenIndex91, depth91 := position, tokenIndex, depth
																if c := buffer[position]; c < rune('a') || c > rune('z') {
																	goto l92
																}
																position++
																goto l91
															l92:
																position, tokenIndex, depth = position91, tokenIndex91, depth91
																if c := buffer[position]; c < rune('A') || c > rune('Z') {
																	goto l90
																}
																position++
															}
														l91:
															goto l89
														l90:
															position, tokenIndex, depth = position90, tokenIndex90, depth90
														}
														depth--
														add(rulePegText, position88)
													}
													goto l86
												l87:
													position, tokenIndex, depth = position86, tokenIndex86, depth86
													if !_rules[rule_]() {
														goto l93
													}
													if !_rules[ruleSTRING]() {
														goto l93
													}
													goto l86
												l93:
													position, tokenIndex, depth = position86, tokenIndex86, depth86
													if !_rules[rule_]() {
														goto l83
													}
													{
														position94 := position
														depth++
														{
															position95, tokenIndex95, depth95 := position, tokenIndex, depth
															if buffer[position] != rune('n') {
																goto l96
															}
															position++
															goto l95
														l96:
															position, tokenIndex, depth = position95, tokenIndex95, depth95
															if buffer[position] != rune('N') {
																goto l83
															}
															position++
														}
													l95:
														{
															position97, tokenIndex97, depth97 := position, tokenIndex, depth
															if buffer[position] != rune('o') {
																goto l98
															}
															position++
															goto l97
														l98:
															position, tokenIndex, depth = position97, tokenIndex97, depth97
															if buffer[position] != rune('O') {
																goto l83
															}
															position++
														}
													l97:
														{
															position99, tokenIndex99, depth99 := position, tokenIndex, depth
															if buffer[position] != rune('w') {
																goto l100
															}
															position++
															goto l99
														l100:
															position, tokenIndex, depth = position99, tokenIndex99, depth99
															if buffer[position] != rune('W') {
																goto l83
															}
															position++
														}
													l99:
														depth--
														add(rulePegText, position94)
													}
													if !_rules[ruleKEY]() {
														goto l83
													}
												}
											l86:
												depth--
												add(ruleTIMESTAMP, position85)
											}
											depth--
											add(rulePROPERTY_VALUE, position84)
										}
										{
											add(ruleAction9, position)
										}
										goto l82
									l83:
										position, tokenIndex, depth = position82, tokenIndex82, depth82
										if !(p.errorHere(position, `expected value to follow key '%s'`, p.contents(tree, tokenIndex-2))) {
											goto l24
										}
									}
								l82:
									{
										add(ruleAction10, position)
									}
									goto l23
								l24:
									position, tokenIndex, depth = position23, tokenIndex23, depth23
									if !_rules[rule_]() {
										goto l103
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
											goto l103
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
											goto l103
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
											goto l103
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
											goto l103
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
											goto l103
										}
										position++
									}
								l112:
									if !_rules[ruleKEY]() {
										goto l103
									}
									if !(p.errorHere(position, `encountered "where" after property clause; "where" blocks must go BEFORE 'from' and 'to' specifiers`)) {
										goto l103
									}
									goto l23
								l103:
									position, tokenIndex, depth = position23, tokenIndex23, depth23
									if !_rules[rule_]() {
										goto l22
									}
									{
										position114, tokenIndex114, depth114 := position, tokenIndex, depth
										{
											position115, tokenIndex115, depth115 := position, tokenIndex, depth
											if !matchDot() {
												goto l115
											}
											goto l114
										l115:
											position, tokenIndex, depth = position115, tokenIndex115, depth115
										}
										goto l22
									l114:
										position, tokenIndex, depth = position114, tokenIndex114, depth114
									}
									if !(p.errorHere(position, `expected key (one of 'from', 'to', 'resolution', or 'sample by') or end of input but got %q following a completed expression`, p.after(position))) {
										goto l22
									}
								}
							l23:
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
						position118 := position
						depth++
						if !_rules[rule_]() {
							goto l0
						}
						{
							position119, tokenIndex119, depth119 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l120
							}
							position++
							goto l119
						l120:
							position, tokenIndex, depth = position119, tokenIndex119, depth119
							if buffer[position] != rune('D') {
								goto l0
							}
							position++
						}
					l119:
						{
							position121, tokenIndex121, depth121 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l122
							}
							position++
							goto l121
						l122:
							position, tokenIndex, depth = position121, tokenIndex121, depth121
							if buffer[position] != rune('E') {
								goto l0
							}
							position++
						}
					l121:
						{
							position123, tokenIndex123, depth123 := position, tokenIndex, depth
							if buffer[position] != rune('s') {
								goto l124
							}
							position++
							goto l123
						l124:
							position, tokenIndex, depth = position123, tokenIndex123, depth123
							if buffer[position] != rune('S') {
								goto l0
							}
							position++
						}
					l123:
						{
							position125, tokenIndex125, depth125 := position, tokenIndex, depth
							if buffer[position] != rune('c') {
								goto l126
							}
							position++
							goto l125
						l126:
							position, tokenIndex, depth = position125, tokenIndex125, depth125
							if buffer[position] != rune('C') {
								goto l0
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
								goto l0
							}
							position++
						}
					l127:
						{
							position129, tokenIndex129, depth129 := position, tokenIndex, depth
							if buffer[position] != rune('i') {
								goto l130
							}
							position++
							goto l129
						l130:
							position, tokenIndex, depth = position129, tokenIndex129, depth129
							if buffer[position] != rune('I') {
								goto l0
							}
							position++
						}
					l129:
						{
							position131, tokenIndex131, depth131 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l132
							}
							position++
							goto l131
						l132:
							position, tokenIndex, depth = position131, tokenIndex131, depth131
							if buffer[position] != rune('B') {
								goto l0
							}
							position++
						}
					l131:
						{
							position133, tokenIndex133, depth133 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l134
							}
							position++
							goto l133
						l134:
							position, tokenIndex, depth = position133, tokenIndex133, depth133
							if buffer[position] != rune('E') {
								goto l0
							}
							position++
						}
					l133:
						if !_rules[ruleKEY]() {
							goto l0
						}
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
									if buffer[position] != rune('a') {
										goto l139
									}
									position++
									goto l138
								l139:
									position, tokenIndex, depth = position138, tokenIndex138, depth138
									if buffer[position] != rune('A') {
										goto l136
									}
									position++
								}
							l138:
								{
									position140, tokenIndex140, depth140 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l141
									}
									position++
									goto l140
								l141:
									position, tokenIndex, depth = position140, tokenIndex140, depth140
									if buffer[position] != rune('L') {
										goto l136
									}
									position++
								}
							l140:
								{
									position142, tokenIndex142, depth142 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l143
									}
									position++
									goto l142
								l143:
									position, tokenIndex, depth = position142, tokenIndex142, depth142
									if buffer[position] != rune('L') {
										goto l136
									}
									position++
								}
							l142:
								if !_rules[ruleKEY]() {
									goto l136
								}
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
												if buffer[position] != rune('m') {
													goto l149
												}
												position++
												goto l148
											l149:
												position, tokenIndex, depth = position148, tokenIndex148, depth148
												if buffer[position] != rune('M') {
													goto l146
												}
												position++
											}
										l148:
											{
												position150, tokenIndex150, depth150 := position, tokenIndex, depth
												if buffer[position] != rune('a') {
													goto l151
												}
												position++
												goto l150
											l151:
												position, tokenIndex, depth = position150, tokenIndex150, depth150
												if buffer[position] != rune('A') {
													goto l146
												}
												position++
											}
										l150:
											{
												position152, tokenIndex152, depth152 := position, tokenIndex, depth
												if buffer[position] != rune('t') {
													goto l153
												}
												position++
												goto l152
											l153:
												position, tokenIndex, depth = position152, tokenIndex152, depth152
												if buffer[position] != rune('T') {
													goto l146
												}
												position++
											}
										l152:
											{
												position154, tokenIndex154, depth154 := position, tokenIndex, depth
												if buffer[position] != rune('c') {
													goto l155
												}
												position++
												goto l154
											l155:
												position, tokenIndex, depth = position154, tokenIndex154, depth154
												if buffer[position] != rune('C') {
													goto l146
												}
												position++
											}
										l154:
											{
												position156, tokenIndex156, depth156 := position, tokenIndex, depth
												if buffer[position] != rune('h') {
													goto l157
												}
												position++
												goto l156
											l157:
												position, tokenIndex, depth = position156, tokenIndex156, depth156
												if buffer[position] != rune('H') {
													goto l146
												}
												position++
											}
										l156:
											if !_rules[ruleKEY]() {
												goto l146
											}
											{
												position158, tokenIndex158, depth158 := position, tokenIndex, depth
												if !_rules[ruleliteralString]() {
													goto l159
												}
												goto l158
											l159:
												position, tokenIndex, depth = position158, tokenIndex158, depth158
												if !(p.errorHere(position, `expected string literal to follow keyword "match"`)) {
													goto l146
												}
											}
										l158:
											{
												add(ruleAction3, position)
											}
											depth--
											add(rulematchClause, position147)
										}
										goto l145
									l146:
										position, tokenIndex, depth = position145, tokenIndex145, depth145
										{
											add(ruleAction2, position)
										}
									}
								l145:
									depth--
									add(ruleoptionalMatchClause, position144)
								}
								{
									add(ruleAction1, position)
								}
								{
									position163, tokenIndex163, depth163 := position, tokenIndex, depth
									{
										position164, tokenIndex164, depth164 := position, tokenIndex, depth
										if !_rules[rule_]() {
											goto l165
										}
										{
											position166, tokenIndex166, depth166 := position, tokenIndex, depth
											if !matchDot() {
												goto l166
											}
											goto l165
										l166:
											position, tokenIndex, depth = position166, tokenIndex166, depth166
										}
										goto l164
									l165:
										position, tokenIndex, depth = position164, tokenIndex164, depth164
										if !_rules[rule_]() {
											goto l136
										}
										if !(p.errorHere(position, `expected end of input after 'describe all' and optional match clause but got %q`, p.after(position))) {
											goto l136
										}
									}
								l164:
									position, tokenIndex, depth = position163, tokenIndex163, depth163
								}
								depth--
								add(ruledescribeAllStmt, position137)
							}
							goto l135
						l136:
							position, tokenIndex, depth = position135, tokenIndex135, depth135
							{
								position168 := position
								depth++
								if !_rules[rule_]() {
									goto l167
								}
								{
									position169, tokenIndex169, depth169 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l170
									}
									position++
									goto l169
								l170:
									position, tokenIndex, depth = position169, tokenIndex169, depth169
									if buffer[position] != rune('M') {
										goto l167
									}
									position++
								}
							l169:
								{
									position171, tokenIndex171, depth171 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l172
									}
									position++
									goto l171
								l172:
									position, tokenIndex, depth = position171, tokenIndex171, depth171
									if buffer[position] != rune('E') {
										goto l167
									}
									position++
								}
							l171:
								{
									position173, tokenIndex173, depth173 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l174
									}
									position++
									goto l173
								l174:
									position, tokenIndex, depth = position173, tokenIndex173, depth173
									if buffer[position] != rune('T') {
										goto l167
									}
									position++
								}
							l173:
								{
									position175, tokenIndex175, depth175 := position, tokenIndex, depth
									if buffer[position] != rune('r') {
										goto l176
									}
									position++
									goto l175
								l176:
									position, tokenIndex, depth = position175, tokenIndex175, depth175
									if buffer[position] != rune('R') {
										goto l167
									}
									position++
								}
							l175:
								{
									position177, tokenIndex177, depth177 := position, tokenIndex, depth
									if buffer[position] != rune('i') {
										goto l178
									}
									position++
									goto l177
								l178:
									position, tokenIndex, depth = position177, tokenIndex177, depth177
									if buffer[position] != rune('I') {
										goto l167
									}
									position++
								}
							l177:
								{
									position179, tokenIndex179, depth179 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l180
									}
									position++
									goto l179
								l180:
									position, tokenIndex, depth = position179, tokenIndex179, depth179
									if buffer[position] != rune('C') {
										goto l167
									}
									position++
								}
							l179:
								{
									position181, tokenIndex181, depth181 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l182
									}
									position++
									goto l181
								l182:
									position, tokenIndex, depth = position181, tokenIndex181, depth181
									if buffer[position] != rune('S') {
										goto l167
									}
									position++
								}
							l181:
								if !_rules[ruleKEY]() {
									goto l167
								}
								{
									position183, tokenIndex183, depth183 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l184
									}
									{
										position185, tokenIndex185, depth185 := position, tokenIndex, depth
										if buffer[position] != rune('w') {
											goto l186
										}
										position++
										goto l185
									l186:
										position, tokenIndex, depth = position185, tokenIndex185, depth185
										if buffer[position] != rune('W') {
											goto l184
										}
										position++
									}
								l185:
									{
										position187, tokenIndex187, depth187 := position, tokenIndex, depth
										if buffer[position] != rune('h') {
											goto l188
										}
										position++
										goto l187
									l188:
										position, tokenIndex, depth = position187, tokenIndex187, depth187
										if buffer[position] != rune('H') {
											goto l184
										}
										position++
									}
								l187:
									{
										position189, tokenIndex189, depth189 := position, tokenIndex, depth
										if buffer[position] != rune('e') {
											goto l190
										}
										position++
										goto l189
									l190:
										position, tokenIndex, depth = position189, tokenIndex189, depth189
										if buffer[position] != rune('E') {
											goto l184
										}
										position++
									}
								l189:
									{
										position191, tokenIndex191, depth191 := position, tokenIndex, depth
										if buffer[position] != rune('r') {
											goto l192
										}
										position++
										goto l191
									l192:
										position, tokenIndex, depth = position191, tokenIndex191, depth191
										if buffer[position] != rune('R') {
											goto l184
										}
										position++
									}
								l191:
									{
										position193, tokenIndex193, depth193 := position, tokenIndex, depth
										if buffer[position] != rune('e') {
											goto l194
										}
										position++
										goto l193
									l194:
										position, tokenIndex, depth = position193, tokenIndex193, depth193
										if buffer[position] != rune('E') {
											goto l184
										}
										position++
									}
								l193:
									if !_rules[ruleKEY]() {
										goto l184
									}
									goto l183
								l184:
									position, tokenIndex, depth = position183, tokenIndex183, depth183
									if !(p.errorHere(position, `expected "where" to follow keyword "metrics" in "describe metrics" command`)) {
										goto l167
									}
								}
							l183:
								{
									position195, tokenIndex195, depth195 := position, tokenIndex, depth
									if !_rules[ruletagName]() {
										goto l196
									}
									goto l195
								l196:
									position, tokenIndex, depth = position195, tokenIndex195, depth195
									if !(p.errorHere(position, `expected tag key to follow keyword "where" in "describe metrics" command`)) {
										goto l167
									}
								}
							l195:
								{
									position197, tokenIndex197, depth197 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l198
									}
									if buffer[position] != rune('=') {
										goto l198
									}
									position++
									goto l197
								l198:
									position, tokenIndex, depth = position197, tokenIndex197, depth197
									if !(p.errorHere(position, `expected "=" to follow keyword "where" in "describe metrics" command`)) {
										goto l167
									}
								}
							l197:
								{
									position199, tokenIndex199, depth199 := position, tokenIndex, depth
									if !_rules[ruleliteralString]() {
										goto l200
									}
									goto l199
								l200:
									position, tokenIndex, depth = position199, tokenIndex199, depth199
									if !(p.errorHere(position, `expected string literal to follow "=" in "describe metrics" command`)) {
										goto l167
									}
								}
							l199:
								{
									add(ruleAction4, position)
								}
								depth--
								add(ruledescribeMetrics, position168)
							}
							goto l135
						l167:
							position, tokenIndex, depth = position135, tokenIndex135, depth135
							{
								position202 := position
								depth++
								{
									position203, tokenIndex203, depth203 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l204
									}
									{
										position205 := position
										depth++
										{
											position206 := position
											depth++
											if !_rules[ruleIDENTIFIER]() {
												goto l204
											}
											depth--
											add(ruleMETRIC_NAME, position206)
										}
										depth--
										add(rulePegText, position205)
									}
									{
										add(ruleAction5, position)
									}
									goto l203
								l204:
									position, tokenIndex, depth = position203, tokenIndex203, depth203
									if !(p.errorHere(position, `expected metric name to follow "describe" in "describe" command`)) {
										goto l0
									}
								}
							l203:
								if !_rules[ruleoptionalPredicateClause]() {
									goto l0
								}
								{
									add(ruleAction6, position)
								}
								depth--
								add(ruledescribeSingleStmt, position202)
							}
						}
					l135:
						depth--
						add(ruledescribeStmt, position118)
					}
				}
			l2:
				if !_rules[rule_]() {
					goto l0
				}
				{
					position209, tokenIndex209, depth209 := position, tokenIndex, depth
					if !matchDot() {
						goto l209
					}
					goto l0
				l209:
					position, tokenIndex, depth = position209, tokenIndex209, depth209
				}
				depth--
				add(ruleroot, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 selectStmt <- <(_ (('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T') KEY)? expressionList &{ p.setContext("after expression of select statement") } optionalPredicateClause &{ p.setContext("") } propertyClause Action0)> */
		nil,
		/* 2 describeStmt <- <(_ (('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E')) KEY (describeAllStmt / describeMetrics / describeSingleStmt))> */
		nil,
		/* 3 describeAllStmt <- <(_ (('a' / 'A') ('l' / 'L') ('l' / 'L')) KEY optionalMatchClause Action1 &((_ !.) / (_ &{p.errorHere(position, `expected end of input after 'describe all' and optional match clause but got %q`, p.after(position) )})))> */
		nil,
		/* 4 optionalMatchClause <- <(matchClause / Action2)> */
		nil,
		/* 5 matchClause <- <(_ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H')) KEY (literalString / &{ p.errorHere(position, `expected string literal to follow keyword "match"`) }) Action3)> */
		nil,
		/* 6 describeMetrics <- <(_ (('m' / 'M') ('e' / 'E') ('t' / 'T') ('r' / 'R') ('i' / 'I') ('c' / 'C') ('s' / 'S')) KEY ((_ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY) / &{ p.errorHere(position, `expected "where" to follow keyword "metrics" in "describe metrics" command`) }) (tagName / &{ p.errorHere(position, `expected tag key to follow keyword "where" in "describe metrics" command`) }) ((_ '=') / &{ p.errorHere(position, `expected "=" to follow keyword "where" in "describe metrics" command`) }) (literalString / &{ p.errorHere(position, `expected string literal to follow "=" in "describe metrics" command`) }) Action4)> */
		nil,
		/* 7 describeSingleStmt <- <(((_ <METRIC_NAME> Action5) / &{ p.errorHere(position, `expected metric name to follow "describe" in "describe" command`) }) optionalPredicateClause Action6)> */
		nil,
		/* 8 propertyClause <- <(Action7 ((_ PROPERTY_KEY Action8 ((_ PROPERTY_VALUE Action9) / &{ p.errorHere(position, `expected value to follow key '%s'`, p.contents(tree, tokenIndex-2)   ) }) Action10) / (_ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY &{ p.errorHere(position, `encountered "where" after property clause; "where" blocks must go BEFORE 'from' and 'to' specifiers`) }) / (_ !!. &{ p.errorHere(position, `expected key (one of 'from', 'to', 'resolution', or 'sample by') or end of input but got %q following a completed expression`, p.after(position)) }))* Action11)> */
		nil,
		/* 9 optionalPredicateClause <- <(predicateClause / Action12)> */
		func() bool {
			{
				position219 := position
				depth++
				{
					position220, tokenIndex220, depth220 := position, tokenIndex, depth
					{
						position222 := position
						depth++
						if !_rules[rule_]() {
							goto l221
						}
						{
							position223, tokenIndex223, depth223 := position, tokenIndex, depth
							if buffer[position] != rune('w') {
								goto l224
							}
							position++
							goto l223
						l224:
							position, tokenIndex, depth = position223, tokenIndex223, depth223
							if buffer[position] != rune('W') {
								goto l221
							}
							position++
						}
					l223:
						{
							position225, tokenIndex225, depth225 := position, tokenIndex, depth
							if buffer[position] != rune('h') {
								goto l226
							}
							position++
							goto l225
						l226:
							position, tokenIndex, depth = position225, tokenIndex225, depth225
							if buffer[position] != rune('H') {
								goto l221
							}
							position++
						}
					l225:
						{
							position227, tokenIndex227, depth227 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l228
							}
							position++
							goto l227
						l228:
							position, tokenIndex, depth = position227, tokenIndex227, depth227
							if buffer[position] != rune('E') {
								goto l221
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
								goto l221
							}
							position++
						}
					l229:
						{
							position231, tokenIndex231, depth231 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l232
							}
							position++
							goto l231
						l232:
							position, tokenIndex, depth = position231, tokenIndex231, depth231
							if buffer[position] != rune('E') {
								goto l221
							}
							position++
						}
					l231:
						if !_rules[ruleKEY]() {
							goto l221
						}
						{
							position233, tokenIndex233, depth233 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l234
							}
							if !_rules[rulepredicate_1]() {
								goto l234
							}
							goto l233
						l234:
							position, tokenIndex, depth = position233, tokenIndex233, depth233
							if !(p.errorHere(position, `expected predicate to follow "where" keyword`)) {
								goto l221
							}
						}
					l233:
						depth--
						add(rulepredicateClause, position222)
					}
					goto l220
				l221:
					position, tokenIndex, depth = position220, tokenIndex220, depth220
					{
						add(ruleAction12, position)
					}
				}
			l220:
				depth--
				add(ruleoptionalPredicateClause, position219)
			}
			return true
		},
		/* 10 expressionList <- <(Action13 expression_start Action14 (_ COMMA (expression_start / &{ p.errorHere(position, `expected expression to follow ","`) }) Action15)*)> */
		func() bool {
			position236, tokenIndex236, depth236 := position, tokenIndex, depth
			{
				position237 := position
				depth++
				{
					add(ruleAction13, position)
				}
				if !_rules[ruleexpression_start]() {
					goto l236
				}
				{
					add(ruleAction14, position)
				}
			l240:
				{
					position241, tokenIndex241, depth241 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l241
					}
					if !_rules[ruleCOMMA]() {
						goto l241
					}
					{
						position242, tokenIndex242, depth242 := position, tokenIndex, depth
						if !_rules[ruleexpression_start]() {
							goto l243
						}
						goto l242
					l243:
						position, tokenIndex, depth = position242, tokenIndex242, depth242
						if !(p.errorHere(position, `expected expression to follow ","`)) {
							goto l241
						}
					}
				l242:
					{
						add(ruleAction15, position)
					}
					goto l240
				l241:
					position, tokenIndex, depth = position241, tokenIndex241, depth241
				}
				depth--
				add(ruleexpressionList, position237)
			}
			return true
		l236:
			position, tokenIndex, depth = position236, tokenIndex236, depth236
			return false
		},
		/* 11 expression_start <- <(expression_sum add_pipe)> */
		func() bool {
			position245, tokenIndex245, depth245 := position, tokenIndex, depth
			{
				position246 := position
				depth++
				{
					position247 := position
					depth++
					if !_rules[ruleexpression_product]() {
						goto l245
					}
				l248:
					{
						position249, tokenIndex249, depth249 := position, tokenIndex, depth
						if !_rules[ruleadd_pipe]() {
							goto l249
						}
						{
							position250, tokenIndex250, depth250 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l251
							}
							{
								position252 := position
								depth++
								if buffer[position] != rune('+') {
									goto l251
								}
								position++
								depth--
								add(ruleOP_ADD, position252)
							}
							{
								add(ruleAction16, position)
							}
							goto l250
						l251:
							position, tokenIndex, depth = position250, tokenIndex250, depth250
							if !_rules[rule_]() {
								goto l249
							}
							{
								position254 := position
								depth++
								if buffer[position] != rune('-') {
									goto l249
								}
								position++
								depth--
								add(ruleOP_SUB, position254)
							}
							{
								add(ruleAction17, position)
							}
						}
					l250:
						{
							position256, tokenIndex256, depth256 := position, tokenIndex, depth
							if !_rules[ruleexpression_product]() {
								goto l257
							}
							goto l256
						l257:
							position, tokenIndex, depth = position256, tokenIndex256, depth256
							if !(p.errorHere(position, `expected expression to follow operator "+" or "-"`)) {
								goto l249
							}
						}
					l256:
						{
							add(ruleAction18, position)
						}
						goto l248
					l249:
						position, tokenIndex, depth = position249, tokenIndex249, depth249
					}
					depth--
					add(ruleexpression_sum, position247)
				}
				if !_rules[ruleadd_pipe]() {
					goto l245
				}
				depth--
				add(ruleexpression_start, position246)
			}
			return true
		l245:
			position, tokenIndex, depth = position245, tokenIndex245, depth245
			return false
		},
		/* 12 expression_sum <- <(expression_product (add_pipe ((_ OP_ADD Action16) / (_ OP_SUB Action17)) (expression_product / &{ p.errorHere(position, `expected expression to follow operator "+" or "-"`) }) Action18)*)> */
		nil,
		/* 13 expression_product <- <(expression_atom (add_pipe ((_ OP_DIV Action19) / (_ OP_MULT Action20)) (expression_atom / &{ p.errorHere(position, `expected expression to follow operator "*" or "/"`) }) Action21)*)> */
		func() bool {
			position260, tokenIndex260, depth260 := position, tokenIndex, depth
			{
				position261 := position
				depth++
				if !_rules[ruleexpression_atom]() {
					goto l260
				}
			l262:
				{
					position263, tokenIndex263, depth263 := position, tokenIndex, depth
					if !_rules[ruleadd_pipe]() {
						goto l263
					}
					{
						position264, tokenIndex264, depth264 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l265
						}
						{
							position266 := position
							depth++
							if buffer[position] != rune('/') {
								goto l265
							}
							position++
							depth--
							add(ruleOP_DIV, position266)
						}
						{
							add(ruleAction19, position)
						}
						goto l264
					l265:
						position, tokenIndex, depth = position264, tokenIndex264, depth264
						if !_rules[rule_]() {
							goto l263
						}
						{
							position268 := position
							depth++
							if buffer[position] != rune('*') {
								goto l263
							}
							position++
							depth--
							add(ruleOP_MULT, position268)
						}
						{
							add(ruleAction20, position)
						}
					}
				l264:
					{
						position270, tokenIndex270, depth270 := position, tokenIndex, depth
						if !_rules[ruleexpression_atom]() {
							goto l271
						}
						goto l270
					l271:
						position, tokenIndex, depth = position270, tokenIndex270, depth270
						if !(p.errorHere(position, `expected expression to follow operator "*" or "/"`)) {
							goto l263
						}
					}
				l270:
					{
						add(ruleAction21, position)
					}
					goto l262
				l263:
					position, tokenIndex, depth = position263, tokenIndex263, depth263
				}
				depth--
				add(ruleexpression_product, position261)
			}
			return true
		l260:
			position, tokenIndex, depth = position260, tokenIndex260, depth260
			return false
		},
		/* 14 add_one_pipe <- <(_ OP_PIPE ((_ <IDENTIFIER>) / &{ p.errorHere(position, `expected function name to follow pipe "|"`) }) Action22 ((_ PAREN_OPEN (expressionList / Action23) optionalGroupBy ((_ PAREN_CLOSE) / &{ p.errorHere(position, `expected ")" to close "(" opened in pipe function call`) })) / Action24) Action25 expression_annotation)> */
		nil,
		/* 15 add_pipe <- <add_one_pipe*> */
		func() bool {
			{
				position275 := position
				depth++
			l276:
				{
					position277, tokenIndex277, depth277 := position, tokenIndex, depth
					{
						position278 := position
						depth++
						if !_rules[rule_]() {
							goto l277
						}
						{
							position279 := position
							depth++
							if buffer[position] != rune('|') {
								goto l277
							}
							position++
							depth--
							add(ruleOP_PIPE, position279)
						}
						{
							position280, tokenIndex280, depth280 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l281
							}
							{
								position282 := position
								depth++
								if !_rules[ruleIDENTIFIER]() {
									goto l281
								}
								depth--
								add(rulePegText, position282)
							}
							goto l280
						l281:
							position, tokenIndex, depth = position280, tokenIndex280, depth280
							if !(p.errorHere(position, `expected function name to follow pipe "|"`)) {
								goto l277
							}
						}
					l280:
						{
							add(ruleAction22, position)
						}
						{
							position284, tokenIndex284, depth284 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l285
							}
							if !_rules[rulePAREN_OPEN]() {
								goto l285
							}
							{
								position286, tokenIndex286, depth286 := position, tokenIndex, depth
								if !_rules[ruleexpressionList]() {
									goto l287
								}
								goto l286
							l287:
								position, tokenIndex, depth = position286, tokenIndex286, depth286
								{
									add(ruleAction23, position)
								}
							}
						l286:
							if !_rules[ruleoptionalGroupBy]() {
								goto l285
							}
							{
								position289, tokenIndex289, depth289 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l290
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l290
								}
								goto l289
							l290:
								position, tokenIndex, depth = position289, tokenIndex289, depth289
								if !(p.errorHere(position, `expected ")" to close "(" opened in pipe function call`)) {
									goto l285
								}
							}
						l289:
							goto l284
						l285:
							position, tokenIndex, depth = position284, tokenIndex284, depth284
							{
								add(ruleAction24, position)
							}
						}
					l284:
						{
							add(ruleAction25, position)
						}
						if !_rules[ruleexpression_annotation]() {
							goto l277
						}
						depth--
						add(ruleadd_one_pipe, position278)
					}
					goto l276
				l277:
					position, tokenIndex, depth = position277, tokenIndex277, depth277
				}
				depth--
				add(ruleadd_pipe, position275)
			}
			return true
		},
		/* 16 expression_atom <- <(expression_atom_raw expression_annotation)> */
		func() bool {
			position293, tokenIndex293, depth293 := position, tokenIndex, depth
			{
				position294 := position
				depth++
				{
					position295 := position
					depth++
					{
						position296, tokenIndex296, depth296 := position, tokenIndex, depth
						{
							position298 := position
							depth++
							if !_rules[rule_]() {
								goto l297
							}
							{
								position299 := position
								depth++
								if !_rules[ruleIDENTIFIER]() {
									goto l297
								}
								depth--
								add(rulePegText, position299)
							}
							{
								add(ruleAction31, position)
							}
							if !_rules[rule_]() {
								goto l297
							}
							if !_rules[rulePAREN_OPEN]() {
								goto l297
							}
							{
								position301, tokenIndex301, depth301 := position, tokenIndex, depth
								if !_rules[ruleexpressionList]() {
									goto l302
								}
								goto l301
							l302:
								position, tokenIndex, depth = position301, tokenIndex301, depth301
								if !(p.errorHere(position, `expected expression list to follow "(" in function call`)) {
									goto l297
								}
							}
						l301:
							if !_rules[ruleoptionalGroupBy]() {
								goto l297
							}
							{
								position303, tokenIndex303, depth303 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l304
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l304
								}
								goto l303
							l304:
								position, tokenIndex, depth = position303, tokenIndex303, depth303
								if !(p.errorHere(position, `expected ")" to close "(" opened by function call`)) {
									goto l297
								}
							}
						l303:
							{
								add(ruleAction32, position)
							}
							depth--
							add(ruleexpression_function, position298)
						}
						goto l296
					l297:
						position, tokenIndex, depth = position296, tokenIndex296, depth296
						{
							position307 := position
							depth++
							if !_rules[rule_]() {
								goto l306
							}
							{
								position308 := position
								depth++
								if !_rules[ruleIDENTIFIER]() {
									goto l306
								}
								depth--
								add(rulePegText, position308)
							}
							{
								add(ruleAction33, position)
							}
							{
								position310, tokenIndex310, depth310 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l311
								}
								if buffer[position] != rune('[') {
									goto l311
								}
								position++
								{
									position312, tokenIndex312, depth312 := position, tokenIndex, depth
									if !_rules[rulepredicate_1]() {
										goto l313
									}
									goto l312
								l313:
									position, tokenIndex, depth = position312, tokenIndex312, depth312
									if !(p.errorHere(position, `expected predicate to follow "[" after metric`)) {
										goto l311
									}
								}
							l312:
								{
									position314, tokenIndex314, depth314 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l315
									}
									if buffer[position] != rune(']') {
										goto l315
									}
									position++
									goto l314
								l315:
									position, tokenIndex, depth = position314, tokenIndex314, depth314
									if !(p.errorHere(position, `expected "]" to close "[" opened to apply predicate`)) {
										goto l311
									}
								}
							l314:
								goto l310
							l311:
								position, tokenIndex, depth = position310, tokenIndex310, depth310
								{
									add(ruleAction34, position)
								}
							}
						l310:
							{
								add(ruleAction35, position)
							}
							depth--
							add(ruleexpression_metric, position307)
						}
						goto l296
					l306:
						position, tokenIndex, depth = position296, tokenIndex296, depth296
						if !_rules[rule_]() {
							goto l318
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l318
						}
						{
							position319, tokenIndex319, depth319 := position, tokenIndex, depth
							if !_rules[ruleexpression_start]() {
								goto l320
							}
							goto l319
						l320:
							position, tokenIndex, depth = position319, tokenIndex319, depth319
							if !(p.errorHere(position, `expected expression to follow "("`)) {
								goto l318
							}
						}
					l319:
						{
							position321, tokenIndex321, depth321 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l322
							}
							if !_rules[rulePAREN_CLOSE]() {
								goto l322
							}
							goto l321
						l322:
							position, tokenIndex, depth = position321, tokenIndex321, depth321
							if !(p.errorHere(position, `expected ")" to close "("`)) {
								goto l318
							}
						}
					l321:
						goto l296
					l318:
						position, tokenIndex, depth = position296, tokenIndex296, depth296
						if !_rules[rule_]() {
							goto l323
						}
						{
							position324 := position
							depth++
							{
								position325 := position
								depth++
								if !_rules[ruleNUMBER]() {
									goto l323
								}
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l323
								}
								position++
							l326:
								{
									position327, tokenIndex327, depth327 := position, tokenIndex, depth
									if c := buffer[position]; c < rune('a') || c > rune('z') {
										goto l327
									}
									position++
									goto l326
								l327:
									position, tokenIndex, depth = position327, tokenIndex327, depth327
								}
								if !_rules[ruleKEY]() {
									goto l323
								}
								depth--
								add(ruleDURATION, position325)
							}
							depth--
							add(rulePegText, position324)
						}
						{
							add(ruleAction26, position)
						}
						goto l296
					l323:
						position, tokenIndex, depth = position296, tokenIndex296, depth296
						if !_rules[rule_]() {
							goto l329
						}
						{
							position330 := position
							depth++
							if !_rules[ruleNUMBER]() {
								goto l329
							}
							depth--
							add(rulePegText, position330)
						}
						{
							add(ruleAction27, position)
						}
						goto l296
					l329:
						position, tokenIndex, depth = position296, tokenIndex296, depth296
						if !_rules[rule_]() {
							goto l293
						}
						if !_rules[ruleSTRING]() {
							goto l293
						}
						{
							add(ruleAction28, position)
						}
					}
				l296:
					depth--
					add(ruleexpression_atom_raw, position295)
				}
				if !_rules[ruleexpression_annotation]() {
					goto l293
				}
				depth--
				add(ruleexpression_atom, position294)
			}
			return true
		l293:
			position, tokenIndex, depth = position293, tokenIndex293, depth293
			return false
		},
		/* 17 expression_atom_raw <- <(expression_function / expression_metric / (_ PAREN_OPEN (expression_start / &{ p.errorHere(position, `expected expression to follow "("`) }) ((_ PAREN_CLOSE) / &{ p.errorHere(position, `expected ")" to close "("`) })) / (_ <DURATION> Action26) / (_ <NUMBER> Action27) / (_ STRING Action28))> */
		nil,
		/* 18 expression_annotation_required <- <(_ '{' <(!'}' .)*> ('}' / &{ p.errorHere(position, `expected "%CLOSEBRACE%" to close "%OPENBRACE%" opened for annotation`) }) Action29)> */
		nil,
		/* 19 expression_annotation <- <expression_annotation_required?> */
		func() bool {
			{
				position336 := position
				depth++
				{
					position337, tokenIndex337, depth337 := position, tokenIndex, depth
					{
						position339 := position
						depth++
						if !_rules[rule_]() {
							goto l337
						}
						if buffer[position] != rune('{') {
							goto l337
						}
						position++
						{
							position340 := position
							depth++
						l341:
							{
								position342, tokenIndex342, depth342 := position, tokenIndex, depth
								{
									position343, tokenIndex343, depth343 := position, tokenIndex, depth
									if buffer[position] != rune('}') {
										goto l343
									}
									position++
									goto l342
								l343:
									position, tokenIndex, depth = position343, tokenIndex343, depth343
								}
								if !matchDot() {
									goto l342
								}
								goto l341
							l342:
								position, tokenIndex, depth = position342, tokenIndex342, depth342
							}
							depth--
							add(rulePegText, position340)
						}
						{
							position344, tokenIndex344, depth344 := position, tokenIndex, depth
							if buffer[position] != rune('}') {
								goto l345
							}
							position++
							goto l344
						l345:
							position, tokenIndex, depth = position344, tokenIndex344, depth344
							if !(p.errorHere(position, `expected "%CLOSEBRACE%" to close "%OPENBRACE%" opened for annotation`)) {
								goto l337
							}
						}
					l344:
						{
							add(ruleAction29, position)
						}
						depth--
						add(ruleexpression_annotation_required, position339)
					}
					goto l338
				l337:
					position, tokenIndex, depth = position337, tokenIndex337, depth337
				}
			l338:
				depth--
				add(ruleexpression_annotation, position336)
			}
			return true
		},
		/* 20 optionalGroupBy <- <(Action30 (groupByClause / collapseByClause)?)> */
		func() bool {
			{
				position348 := position
				depth++
				{
					add(ruleAction30, position)
				}
				{
					position350, tokenIndex350, depth350 := position, tokenIndex, depth
					{
						position352, tokenIndex352, depth352 := position, tokenIndex, depth
						{
							position354 := position
							depth++
							if !_rules[rule_]() {
								goto l353
							}
							{
								position355, tokenIndex355, depth355 := position, tokenIndex, depth
								if buffer[position] != rune('g') {
									goto l356
								}
								position++
								goto l355
							l356:
								position, tokenIndex, depth = position355, tokenIndex355, depth355
								if buffer[position] != rune('G') {
									goto l353
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
									goto l353
								}
								position++
							}
						l357:
							{
								position359, tokenIndex359, depth359 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l360
								}
								position++
								goto l359
							l360:
								position, tokenIndex, depth = position359, tokenIndex359, depth359
								if buffer[position] != rune('O') {
									goto l353
								}
								position++
							}
						l359:
							{
								position361, tokenIndex361, depth361 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l362
								}
								position++
								goto l361
							l362:
								position, tokenIndex, depth = position361, tokenIndex361, depth361
								if buffer[position] != rune('U') {
									goto l353
								}
								position++
							}
						l361:
							{
								position363, tokenIndex363, depth363 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l364
								}
								position++
								goto l363
							l364:
								position, tokenIndex, depth = position363, tokenIndex363, depth363
								if buffer[position] != rune('P') {
									goto l353
								}
								position++
							}
						l363:
							if !_rules[ruleKEY]() {
								goto l353
							}
							{
								position365, tokenIndex365, depth365 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l366
								}
								{
									position367, tokenIndex367, depth367 := position, tokenIndex, depth
									if buffer[position] != rune('b') {
										goto l368
									}
									position++
									goto l367
								l368:
									position, tokenIndex, depth = position367, tokenIndex367, depth367
									if buffer[position] != rune('B') {
										goto l366
									}
									position++
								}
							l367:
								{
									position369, tokenIndex369, depth369 := position, tokenIndex, depth
									if buffer[position] != rune('y') {
										goto l370
									}
									position++
									goto l369
								l370:
									position, tokenIndex, depth = position369, tokenIndex369, depth369
									if buffer[position] != rune('Y') {
										goto l366
									}
									position++
								}
							l369:
								if !_rules[ruleKEY]() {
									goto l366
								}
								goto l365
							l366:
								position, tokenIndex, depth = position365, tokenIndex365, depth365
								if !(p.errorHere(position, `expected keyword "by" to follow keyword "group" in "group by" clause`)) {
									goto l353
								}
							}
						l365:
							{
								position371, tokenIndex371, depth371 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l372
								}
								{
									position373 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l372
									}
									depth--
									add(rulePegText, position373)
								}
								goto l371
							l372:
								position, tokenIndex, depth = position371, tokenIndex371, depth371
								if !(p.errorHere(position, `expected tag key identifier to follow "group by" keywords in "group by" clause`)) {
									goto l353
								}
							}
						l371:
							{
								add(ruleAction36, position)
							}
						l375:
							{
								position376, tokenIndex376, depth376 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l376
								}
								if !_rules[ruleCOMMA]() {
									goto l376
								}
								{
									position377, tokenIndex377, depth377 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l378
									}
									{
										position379 := position
										depth++
										if !_rules[ruleCOLUMN_NAME]() {
											goto l378
										}
										depth--
										add(rulePegText, position379)
									}
									goto l377
								l378:
									position, tokenIndex, depth = position377, tokenIndex377, depth377
									if !(p.errorHere(position, `expected tag key identifier to follow "," in "group by" clause`)) {
										goto l376
									}
								}
							l377:
								{
									add(ruleAction37, position)
								}
								goto l375
							l376:
								position, tokenIndex, depth = position376, tokenIndex376, depth376
							}
							depth--
							add(rulegroupByClause, position354)
						}
						goto l352
					l353:
						position, tokenIndex, depth = position352, tokenIndex352, depth352
						{
							position381 := position
							depth++
							if !_rules[rule_]() {
								goto l350
							}
							{
								position382, tokenIndex382, depth382 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l383
								}
								position++
								goto l382
							l383:
								position, tokenIndex, depth = position382, tokenIndex382, depth382
								if buffer[position] != rune('C') {
									goto l350
								}
								position++
							}
						l382:
							{
								position384, tokenIndex384, depth384 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l385
								}
								position++
								goto l384
							l385:
								position, tokenIndex, depth = position384, tokenIndex384, depth384
								if buffer[position] != rune('O') {
									goto l350
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
									goto l350
								}
								position++
							}
						l386:
							{
								position388, tokenIndex388, depth388 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l389
								}
								position++
								goto l388
							l389:
								position, tokenIndex, depth = position388, tokenIndex388, depth388
								if buffer[position] != rune('L') {
									goto l350
								}
								position++
							}
						l388:
							{
								position390, tokenIndex390, depth390 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l391
								}
								position++
								goto l390
							l391:
								position, tokenIndex, depth = position390, tokenIndex390, depth390
								if buffer[position] != rune('A') {
									goto l350
								}
								position++
							}
						l390:
							{
								position392, tokenIndex392, depth392 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l393
								}
								position++
								goto l392
							l393:
								position, tokenIndex, depth = position392, tokenIndex392, depth392
								if buffer[position] != rune('P') {
									goto l350
								}
								position++
							}
						l392:
							{
								position394, tokenIndex394, depth394 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l395
								}
								position++
								goto l394
							l395:
								position, tokenIndex, depth = position394, tokenIndex394, depth394
								if buffer[position] != rune('S') {
									goto l350
								}
								position++
							}
						l394:
							{
								position396, tokenIndex396, depth396 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l397
								}
								position++
								goto l396
							l397:
								position, tokenIndex, depth = position396, tokenIndex396, depth396
								if buffer[position] != rune('E') {
									goto l350
								}
								position++
							}
						l396:
							if !_rules[ruleKEY]() {
								goto l350
							}
							{
								position398, tokenIndex398, depth398 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l399
								}
								{
									position400, tokenIndex400, depth400 := position, tokenIndex, depth
									if buffer[position] != rune('b') {
										goto l401
									}
									position++
									goto l400
								l401:
									position, tokenIndex, depth = position400, tokenIndex400, depth400
									if buffer[position] != rune('B') {
										goto l399
									}
									position++
								}
							l400:
								{
									position402, tokenIndex402, depth402 := position, tokenIndex, depth
									if buffer[position] != rune('y') {
										goto l403
									}
									position++
									goto l402
								l403:
									position, tokenIndex, depth = position402, tokenIndex402, depth402
									if buffer[position] != rune('Y') {
										goto l399
									}
									position++
								}
							l402:
								if !_rules[ruleKEY]() {
									goto l399
								}
								goto l398
							l399:
								position, tokenIndex, depth = position398, tokenIndex398, depth398
								if !(p.errorHere(position, `expected keyword "by" to follow keyword "collapse" in "collapse by" clause`)) {
									goto l350
								}
							}
						l398:
							{
								position404, tokenIndex404, depth404 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l405
								}
								{
									position406 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l405
									}
									depth--
									add(rulePegText, position406)
								}
								goto l404
							l405:
								position, tokenIndex, depth = position404, tokenIndex404, depth404
								if !(p.errorHere(position, `expected tag key identifier to follow "collapse by" keywords in "collapse by" clause`)) {
									goto l350
								}
							}
						l404:
							{
								add(ruleAction38, position)
							}
						l408:
							{
								position409, tokenIndex409, depth409 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l409
								}
								if !_rules[ruleCOMMA]() {
									goto l409
								}
								{
									position410, tokenIndex410, depth410 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l411
									}
									{
										position412 := position
										depth++
										if !_rules[ruleCOLUMN_NAME]() {
											goto l411
										}
										depth--
										add(rulePegText, position412)
									}
									goto l410
								l411:
									position, tokenIndex, depth = position410, tokenIndex410, depth410
									if !(p.errorHere(position, `expected tag key identifier to follow "," in "collapse by" clause`)) {
										goto l409
									}
								}
							l410:
								{
									add(ruleAction39, position)
								}
								goto l408
							l409:
								position, tokenIndex, depth = position409, tokenIndex409, depth409
							}
							depth--
							add(rulecollapseByClause, position381)
						}
					}
				l352:
					goto l351
				l350:
					position, tokenIndex, depth = position350, tokenIndex350, depth350
				}
			l351:
				depth--
				add(ruleoptionalGroupBy, position348)
			}
			return true
		},
		/* 21 expression_function <- <(_ <IDENTIFIER> Action31 _ PAREN_OPEN (expressionList / &{ p.errorHere(position, `expected expression list to follow "(" in function call`) }) optionalGroupBy ((_ PAREN_CLOSE) / &{ p.errorHere(position, `expected ")" to close "(" opened by function call`) }) Action32)> */
		nil,
		/* 22 expression_metric <- <(_ <IDENTIFIER> Action33 ((_ '[' (predicate_1 / &{ p.errorHere(position, `expected predicate to follow "[" after metric`) }) ((_ ']') / &{ p.errorHere(position, `expected "]" to close "[" opened to apply predicate`) })) / Action34) Action35)> */
		nil,
		/* 23 groupByClause <- <(_ (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P')) KEY ((_ (('b' / 'B') ('y' / 'Y')) KEY) / &{ p.errorHere(position, `expected keyword "by" to follow keyword "group" in "group by" clause`) }) ((_ <COLUMN_NAME>) / &{ p.errorHere(position, `expected tag key identifier to follow "group by" keywords in "group by" clause`) }) Action36 (_ COMMA ((_ <COLUMN_NAME>) / &{ p.errorHere(position, `expected tag key identifier to follow "," in "group by" clause`) }) Action37)*)> */
		nil,
		/* 24 collapseByClause <- <(_ (('c' / 'C') ('o' / 'O') ('l' / 'L') ('l' / 'L') ('a' / 'A') ('p' / 'P') ('s' / 'S') ('e' / 'E')) KEY ((_ (('b' / 'B') ('y' / 'Y')) KEY) / &{ p.errorHere(position, `expected keyword "by" to follow keyword "collapse" in "collapse by" clause`) }) ((_ <COLUMN_NAME>) / &{ p.errorHere(position, `expected tag key identifier to follow "collapse by" keywords in "collapse by" clause`) }) Action38 (_ COMMA ((_ <COLUMN_NAME>) / &{ p.errorHere(position, `expected tag key identifier to follow "," in "collapse by" clause`) }) Action39)*)> */
		nil,
		/* 25 predicateClause <- <(_ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY ((_ predicate_1) / &{ p.errorHere(position, `expected predicate to follow "where" keyword`) }))> */
		nil,
		/* 26 predicate_1 <- <((predicate_2 _ OP_OR (predicate_1 / &{ p.errorHere(position, `expected predicate to follow "or" operator`) }) Action40) / predicate_2)> */
		func() bool {
			position419, tokenIndex419, depth419 := position, tokenIndex, depth
			{
				position420 := position
				depth++
				{
					position421, tokenIndex421, depth421 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l422
					}
					if !_rules[rule_]() {
						goto l422
					}
					{
						position423 := position
						depth++
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
								goto l422
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
								goto l422
							}
							position++
						}
					l426:
						if !_rules[ruleKEY]() {
							goto l422
						}
						depth--
						add(ruleOP_OR, position423)
					}
					{
						position428, tokenIndex428, depth428 := position, tokenIndex, depth
						if !_rules[rulepredicate_1]() {
							goto l429
						}
						goto l428
					l429:
						position, tokenIndex, depth = position428, tokenIndex428, depth428
						if !(p.errorHere(position, `expected predicate to follow "or" operator`)) {
							goto l422
						}
					}
				l428:
					{
						add(ruleAction40, position)
					}
					goto l421
				l422:
					position, tokenIndex, depth = position421, tokenIndex421, depth421
					if !_rules[rulepredicate_2]() {
						goto l419
					}
				}
			l421:
				depth--
				add(rulepredicate_1, position420)
			}
			return true
		l419:
			position, tokenIndex, depth = position419, tokenIndex419, depth419
			return false
		},
		/* 27 predicate_2 <- <((predicate_3 _ OP_AND (predicate_2 / &{ p.errorHere(position, `expected predicate to follow "and" operator`) }) Action41) / predicate_3)> */
		func() bool {
			position431, tokenIndex431, depth431 := position, tokenIndex, depth
			{
				position432 := position
				depth++
				{
					position433, tokenIndex433, depth433 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l434
					}
					if !_rules[rule_]() {
						goto l434
					}
					{
						position435 := position
						depth++
						{
							position436, tokenIndex436, depth436 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l437
							}
							position++
							goto l436
						l437:
							position, tokenIndex, depth = position436, tokenIndex436, depth436
							if buffer[position] != rune('A') {
								goto l434
							}
							position++
						}
					l436:
						{
							position438, tokenIndex438, depth438 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l439
							}
							position++
							goto l438
						l439:
							position, tokenIndex, depth = position438, tokenIndex438, depth438
							if buffer[position] != rune('N') {
								goto l434
							}
							position++
						}
					l438:
						{
							position440, tokenIndex440, depth440 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l441
							}
							position++
							goto l440
						l441:
							position, tokenIndex, depth = position440, tokenIndex440, depth440
							if buffer[position] != rune('D') {
								goto l434
							}
							position++
						}
					l440:
						if !_rules[ruleKEY]() {
							goto l434
						}
						depth--
						add(ruleOP_AND, position435)
					}
					{
						position442, tokenIndex442, depth442 := position, tokenIndex, depth
						if !_rules[rulepredicate_2]() {
							goto l443
						}
						goto l442
					l443:
						position, tokenIndex, depth = position442, tokenIndex442, depth442
						if !(p.errorHere(position, `expected predicate to follow "and" operator`)) {
							goto l434
						}
					}
				l442:
					{
						add(ruleAction41, position)
					}
					goto l433
				l434:
					position, tokenIndex, depth = position433, tokenIndex433, depth433
					if !_rules[rulepredicate_3]() {
						goto l431
					}
				}
			l433:
				depth--
				add(rulepredicate_2, position432)
			}
			return true
		l431:
			position, tokenIndex, depth = position431, tokenIndex431, depth431
			return false
		},
		/* 28 predicate_3 <- <((_ OP_NOT (predicate_3 / &{ p.errorHere(position, `expected predicate to follow "not" operator`) }) Action42) / (_ PAREN_OPEN (predicate_1 / &{ p.errorHere(position, `expected predicate to follow "("`) }) ((_ PAREN_CLOSE) / &{ p.errorHere(position, `expected ")" to close "(" opened in predicate`) })) / tagMatcher)> */
		func() bool {
			position445, tokenIndex445, depth445 := position, tokenIndex, depth
			{
				position446 := position
				depth++
				{
					position447, tokenIndex447, depth447 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l448
					}
					{
						position449 := position
						depth++
						{
							position450, tokenIndex450, depth450 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l451
							}
							position++
							goto l450
						l451:
							position, tokenIndex, depth = position450, tokenIndex450, depth450
							if buffer[position] != rune('N') {
								goto l448
							}
							position++
						}
					l450:
						{
							position452, tokenIndex452, depth452 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l453
							}
							position++
							goto l452
						l453:
							position, tokenIndex, depth = position452, tokenIndex452, depth452
							if buffer[position] != rune('O') {
								goto l448
							}
							position++
						}
					l452:
						{
							position454, tokenIndex454, depth454 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l455
							}
							position++
							goto l454
						l455:
							position, tokenIndex, depth = position454, tokenIndex454, depth454
							if buffer[position] != rune('T') {
								goto l448
							}
							position++
						}
					l454:
						if !_rules[ruleKEY]() {
							goto l448
						}
						depth--
						add(ruleOP_NOT, position449)
					}
					{
						position456, tokenIndex456, depth456 := position, tokenIndex, depth
						if !_rules[rulepredicate_3]() {
							goto l457
						}
						goto l456
					l457:
						position, tokenIndex, depth = position456, tokenIndex456, depth456
						if !(p.errorHere(position, `expected predicate to follow "not" operator`)) {
							goto l448
						}
					}
				l456:
					{
						add(ruleAction42, position)
					}
					goto l447
				l448:
					position, tokenIndex, depth = position447, tokenIndex447, depth447
					if !_rules[rule_]() {
						goto l459
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l459
					}
					{
						position460, tokenIndex460, depth460 := position, tokenIndex, depth
						if !_rules[rulepredicate_1]() {
							goto l461
						}
						goto l460
					l461:
						position, tokenIndex, depth = position460, tokenIndex460, depth460
						if !(p.errorHere(position, `expected predicate to follow "("`)) {
							goto l459
						}
					}
				l460:
					{
						position462, tokenIndex462, depth462 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l463
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l463
						}
						goto l462
					l463:
						position, tokenIndex, depth = position462, tokenIndex462, depth462
						if !(p.errorHere(position, `expected ")" to close "(" opened in predicate`)) {
							goto l459
						}
					}
				l462:
					goto l447
				l459:
					position, tokenIndex, depth = position447, tokenIndex447, depth447
					{
						position464 := position
						depth++
						if !_rules[ruletagName]() {
							goto l445
						}
						{
							position465, tokenIndex465, depth465 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l466
							}
							if buffer[position] != rune('=') {
								goto l466
							}
							position++
							{
								position467, tokenIndex467, depth467 := position, tokenIndex, depth
								if !_rules[ruleliteralString]() {
									goto l468
								}
								goto l467
							l468:
								position, tokenIndex, depth = position467, tokenIndex467, depth467
								if !(p.errorHere(position, `expected string literal to follow "="`)) {
									goto l466
								}
							}
						l467:
							{
								add(ruleAction43, position)
							}
							goto l465
						l466:
							position, tokenIndex, depth = position465, tokenIndex465, depth465
							if !_rules[rule_]() {
								goto l470
							}
							if buffer[position] != rune('!') {
								goto l470
							}
							position++
							if buffer[position] != rune('=') {
								goto l470
							}
							position++
							{
								position471, tokenIndex471, depth471 := position, tokenIndex, depth
								if !_rules[ruleliteralString]() {
									goto l472
								}
								goto l471
							l472:
								position, tokenIndex, depth = position471, tokenIndex471, depth471
								if !(p.errorHere(position, `expected string literal to follow "!="`)) {
									goto l470
								}
							}
						l471:
							{
								add(ruleAction44, position)
							}
							{
								add(ruleAction45, position)
							}
							goto l465
						l470:
							position, tokenIndex, depth = position465, tokenIndex465, depth465
							if !_rules[rule_]() {
								goto l475
							}
							{
								position476, tokenIndex476, depth476 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l477
								}
								position++
								goto l476
							l477:
								position, tokenIndex, depth = position476, tokenIndex476, depth476
								if buffer[position] != rune('M') {
									goto l475
								}
								position++
							}
						l476:
							{
								position478, tokenIndex478, depth478 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l479
								}
								position++
								goto l478
							l479:
								position, tokenIndex, depth = position478, tokenIndex478, depth478
								if buffer[position] != rune('A') {
									goto l475
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
									goto l475
								}
								position++
							}
						l480:
							{
								position482, tokenIndex482, depth482 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l483
								}
								position++
								goto l482
							l483:
								position, tokenIndex, depth = position482, tokenIndex482, depth482
								if buffer[position] != rune('C') {
									goto l475
								}
								position++
							}
						l482:
							{
								position484, tokenIndex484, depth484 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l485
								}
								position++
								goto l484
							l485:
								position, tokenIndex, depth = position484, tokenIndex484, depth484
								if buffer[position] != rune('H') {
									goto l475
								}
								position++
							}
						l484:
							if !_rules[ruleKEY]() {
								goto l475
							}
							{
								position486, tokenIndex486, depth486 := position, tokenIndex, depth
								if !_rules[ruleliteralString]() {
									goto l487
								}
								goto l486
							l487:
								position, tokenIndex, depth = position486, tokenIndex486, depth486
								if !(p.errorHere(position, `expected regex string literal to follow "match"`)) {
									goto l475
								}
							}
						l486:
							{
								add(ruleAction46, position)
							}
							goto l465
						l475:
							position, tokenIndex, depth = position465, tokenIndex465, depth465
							if !_rules[rule_]() {
								goto l489
							}
							{
								position490, tokenIndex490, depth490 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l491
								}
								position++
								goto l490
							l491:
								position, tokenIndex, depth = position490, tokenIndex490, depth490
								if buffer[position] != rune('I') {
									goto l489
								}
								position++
							}
						l490:
							{
								position492, tokenIndex492, depth492 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l493
								}
								position++
								goto l492
							l493:
								position, tokenIndex, depth = position492, tokenIndex492, depth492
								if buffer[position] != rune('N') {
									goto l489
								}
								position++
							}
						l492:
							if !_rules[ruleKEY]() {
								goto l489
							}
							{
								position494, tokenIndex494, depth494 := position, tokenIndex, depth
								{
									position496 := position
									depth++
									{
										add(ruleAction49, position)
									}
									if !_rules[rule_]() {
										goto l495
									}
									if !_rules[rulePAREN_OPEN]() {
										goto l495
									}
									{
										position498, tokenIndex498, depth498 := position, tokenIndex, depth
										if !_rules[ruleliteralListString]() {
											goto l499
										}
										goto l498
									l499:
										position, tokenIndex, depth = position498, tokenIndex498, depth498
										if !(p.errorHere(position, `expected string literal to follow "(" in literal list`)) {
											goto l495
										}
									}
								l498:
								l500:
									{
										position501, tokenIndex501, depth501 := position, tokenIndex, depth
										if !_rules[rule_]() {
											goto l501
										}
										if !_rules[ruleCOMMA]() {
											goto l501
										}
										{
											position502, tokenIndex502, depth502 := position, tokenIndex, depth
											if !_rules[ruleliteralListString]() {
												goto l503
											}
											goto l502
										l503:
											position, tokenIndex, depth = position502, tokenIndex502, depth502
											if !(p.errorHere(position, `expected string literal to follow "," in literal list`)) {
												goto l501
											}
										}
									l502:
										goto l500
									l501:
										position, tokenIndex, depth = position501, tokenIndex501, depth501
									}
									{
										position504, tokenIndex504, depth504 := position, tokenIndex, depth
										if !_rules[rule_]() {
											goto l505
										}
										if !_rules[rulePAREN_CLOSE]() {
											goto l505
										}
										goto l504
									l505:
										position, tokenIndex, depth = position504, tokenIndex504, depth504
										if !(p.errorHere(position, `expected ")" to close "(" for literal list`)) {
											goto l495
										}
									}
								l504:
									depth--
									add(ruleliteralList, position496)
								}
								goto l494
							l495:
								position, tokenIndex, depth = position494, tokenIndex494, depth494
								if !(p.errorHere(position, `expected string literal list to follow "in" keyword`)) {
									goto l489
								}
							}
						l494:
							{
								add(ruleAction47, position)
							}
							goto l465
						l489:
							position, tokenIndex, depth = position465, tokenIndex465, depth465
							if !(p.errorHere(position, `expected "=", "!=", "match", or "in" to follow tag key in predicate`)) {
								goto l445
							}
						}
					l465:
						depth--
						add(ruletagMatcher, position464)
					}
				}
			l447:
				depth--
				add(rulepredicate_3, position446)
			}
			return true
		l445:
			position, tokenIndex, depth = position445, tokenIndex445, depth445
			return false
		},
		/* 29 tagMatcher <- <(tagName ((_ '=' (literalString / &{ p.errorHere(position, `expected string literal to follow "="`) }) Action43) / (_ ('!' '=') (literalString / &{ p.errorHere(position, `expected string literal to follow "!="`) }) Action44 Action45) / (_ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H')) KEY (literalString / &{ p.errorHere(position, `expected regex string literal to follow "match"`) }) Action46) / (_ (('i' / 'I') ('n' / 'N')) KEY (literalList / &{ p.errorHere(position, `expected string literal list to follow "in" keyword`) }) Action47) / &{ p.errorHere(position, `expected "=", "!=", "match", or "in" to follow tag key in predicate`) }))> */
		nil,
		/* 30 literalString <- <(_ STRING Action48)> */
		func() bool {
			position508, tokenIndex508, depth508 := position, tokenIndex, depth
			{
				position509 := position
				depth++
				if !_rules[rule_]() {
					goto l508
				}
				if !_rules[ruleSTRING]() {
					goto l508
				}
				{
					add(ruleAction48, position)
				}
				depth--
				add(ruleliteralString, position509)
			}
			return true
		l508:
			position, tokenIndex, depth = position508, tokenIndex508, depth508
			return false
		},
		/* 31 literalList <- <(Action49 _ PAREN_OPEN (literalListString / &{ p.errorHere(position, `expected string literal to follow "(" in literal list`) }) (_ COMMA (literalListString / &{ p.errorHere(position, `expected string literal to follow "," in literal list`) }))* ((_ PAREN_CLOSE) / &{ p.errorHere(position, `expected ")" to close "(" for literal list`) }))> */
		nil,
		/* 32 literalListString <- <(_ STRING Action50)> */
		func() bool {
			position512, tokenIndex512, depth512 := position, tokenIndex, depth
			{
				position513 := position
				depth++
				if !_rules[rule_]() {
					goto l512
				}
				if !_rules[ruleSTRING]() {
					goto l512
				}
				{
					add(ruleAction50, position)
				}
				depth--
				add(ruleliteralListString, position513)
			}
			return true
		l512:
			position, tokenIndex, depth = position512, tokenIndex512, depth512
			return false
		},
		/* 33 tagName <- <(_ <TAG_NAME> Action51)> */
		func() bool {
			position515, tokenIndex515, depth515 := position, tokenIndex, depth
			{
				position516 := position
				depth++
				if !_rules[rule_]() {
					goto l515
				}
				{
					position517 := position
					depth++
					{
						position518 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l515
						}
						depth--
						add(ruleTAG_NAME, position518)
					}
					depth--
					add(rulePegText, position517)
				}
				{
					add(ruleAction51, position)
				}
				depth--
				add(ruletagName, position516)
			}
			return true
		l515:
			position, tokenIndex, depth = position515, tokenIndex515, depth515
			return false
		},
		/* 34 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position520, tokenIndex520, depth520 := position, tokenIndex, depth
			{
				position521 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l520
				}
				depth--
				add(ruleCOLUMN_NAME, position521)
			}
			return true
		l520:
			position, tokenIndex, depth = position520, tokenIndex520, depth520
			return false
		},
		/* 35 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 36 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 37 IDENTIFIER <- <(('`' CHAR* ('`' / &{ p.errorHere(position, "expected \"`\" to end identifier") })) / (!(KEYWORD KEY) ID_SEGMENT ('.' (ID_SEGMENT / &{ p.errorHere(position, `expected identifier segment to follow "."`) }))*))> */
		func() bool {
			position524, tokenIndex524, depth524 := position, tokenIndex, depth
			{
				position525 := position
				depth++
				{
					position526, tokenIndex526, depth526 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l527
					}
					position++
				l528:
					{
						position529, tokenIndex529, depth529 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l529
						}
						goto l528
					l529:
						position, tokenIndex, depth = position529, tokenIndex529, depth529
					}
					{
						position530, tokenIndex530, depth530 := position, tokenIndex, depth
						if buffer[position] != rune('`') {
							goto l531
						}
						position++
						goto l530
					l531:
						position, tokenIndex, depth = position530, tokenIndex530, depth530
						if !(p.errorHere(position, "expected \"`\" to end identifier")) {
							goto l527
						}
					}
				l530:
					goto l526
				l527:
					position, tokenIndex, depth = position526, tokenIndex526, depth526
					{
						position532, tokenIndex532, depth532 := position, tokenIndex, depth
						{
							position533 := position
							depth++
							{
								position534, tokenIndex534, depth534 := position, tokenIndex, depth
								{
									position536, tokenIndex536, depth536 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l537
									}
									position++
									goto l536
								l537:
									position, tokenIndex, depth = position536, tokenIndex536, depth536
									if buffer[position] != rune('A') {
										goto l535
									}
									position++
								}
							l536:
								{
									position538, tokenIndex538, depth538 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l539
									}
									position++
									goto l538
								l539:
									position, tokenIndex, depth = position538, tokenIndex538, depth538
									if buffer[position] != rune('L') {
										goto l535
									}
									position++
								}
							l538:
								{
									position540, tokenIndex540, depth540 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l541
									}
									position++
									goto l540
								l541:
									position, tokenIndex, depth = position540, tokenIndex540, depth540
									if buffer[position] != rune('L') {
										goto l535
									}
									position++
								}
							l540:
								goto l534
							l535:
								position, tokenIndex, depth = position534, tokenIndex534, depth534
								{
									position543, tokenIndex543, depth543 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l544
									}
									position++
									goto l543
								l544:
									position, tokenIndex, depth = position543, tokenIndex543, depth543
									if buffer[position] != rune('A') {
										goto l542
									}
									position++
								}
							l543:
								{
									position545, tokenIndex545, depth545 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l546
									}
									position++
									goto l545
								l546:
									position, tokenIndex, depth = position545, tokenIndex545, depth545
									if buffer[position] != rune('N') {
										goto l542
									}
									position++
								}
							l545:
								{
									position547, tokenIndex547, depth547 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l548
									}
									position++
									goto l547
								l548:
									position, tokenIndex, depth = position547, tokenIndex547, depth547
									if buffer[position] != rune('D') {
										goto l542
									}
									position++
								}
							l547:
								goto l534
							l542:
								position, tokenIndex, depth = position534, tokenIndex534, depth534
								{
									position550, tokenIndex550, depth550 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l551
									}
									position++
									goto l550
								l551:
									position, tokenIndex, depth = position550, tokenIndex550, depth550
									if buffer[position] != rune('M') {
										goto l549
									}
									position++
								}
							l550:
								{
									position552, tokenIndex552, depth552 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l553
									}
									position++
									goto l552
								l553:
									position, tokenIndex, depth = position552, tokenIndex552, depth552
									if buffer[position] != rune('A') {
										goto l549
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
										goto l549
									}
									position++
								}
							l554:
								{
									position556, tokenIndex556, depth556 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l557
									}
									position++
									goto l556
								l557:
									position, tokenIndex, depth = position556, tokenIndex556, depth556
									if buffer[position] != rune('C') {
										goto l549
									}
									position++
								}
							l556:
								{
									position558, tokenIndex558, depth558 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l559
									}
									position++
									goto l558
								l559:
									position, tokenIndex, depth = position558, tokenIndex558, depth558
									if buffer[position] != rune('H') {
										goto l549
									}
									position++
								}
							l558:
								goto l534
							l549:
								position, tokenIndex, depth = position534, tokenIndex534, depth534
								{
									position561, tokenIndex561, depth561 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l562
									}
									position++
									goto l561
								l562:
									position, tokenIndex, depth = position561, tokenIndex561, depth561
									if buffer[position] != rune('S') {
										goto l560
									}
									position++
								}
							l561:
								{
									position563, tokenIndex563, depth563 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l564
									}
									position++
									goto l563
								l564:
									position, tokenIndex, depth = position563, tokenIndex563, depth563
									if buffer[position] != rune('E') {
										goto l560
									}
									position++
								}
							l563:
								{
									position565, tokenIndex565, depth565 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l566
									}
									position++
									goto l565
								l566:
									position, tokenIndex, depth = position565, tokenIndex565, depth565
									if buffer[position] != rune('L') {
										goto l560
									}
									position++
								}
							l565:
								{
									position567, tokenIndex567, depth567 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l568
									}
									position++
									goto l567
								l568:
									position, tokenIndex, depth = position567, tokenIndex567, depth567
									if buffer[position] != rune('E') {
										goto l560
									}
									position++
								}
							l567:
								{
									position569, tokenIndex569, depth569 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l570
									}
									position++
									goto l569
								l570:
									position, tokenIndex, depth = position569, tokenIndex569, depth569
									if buffer[position] != rune('C') {
										goto l560
									}
									position++
								}
							l569:
								{
									position571, tokenIndex571, depth571 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l572
									}
									position++
									goto l571
								l572:
									position, tokenIndex, depth = position571, tokenIndex571, depth571
									if buffer[position] != rune('T') {
										goto l560
									}
									position++
								}
							l571:
								goto l534
							l560:
								position, tokenIndex, depth = position534, tokenIndex534, depth534
								{
									switch buffer[position] {
									case 'S', 's':
										{
											position574, tokenIndex574, depth574 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l575
											}
											position++
											goto l574
										l575:
											position, tokenIndex, depth = position574, tokenIndex574, depth574
											if buffer[position] != rune('S') {
												goto l532
											}
											position++
										}
									l574:
										{
											position576, tokenIndex576, depth576 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l577
											}
											position++
											goto l576
										l577:
											position, tokenIndex, depth = position576, tokenIndex576, depth576
											if buffer[position] != rune('A') {
												goto l532
											}
											position++
										}
									l576:
										{
											position578, tokenIndex578, depth578 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l579
											}
											position++
											goto l578
										l579:
											position, tokenIndex, depth = position578, tokenIndex578, depth578
											if buffer[position] != rune('M') {
												goto l532
											}
											position++
										}
									l578:
										{
											position580, tokenIndex580, depth580 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l581
											}
											position++
											goto l580
										l581:
											position, tokenIndex, depth = position580, tokenIndex580, depth580
											if buffer[position] != rune('P') {
												goto l532
											}
											position++
										}
									l580:
										{
											position582, tokenIndex582, depth582 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l583
											}
											position++
											goto l582
										l583:
											position, tokenIndex, depth = position582, tokenIndex582, depth582
											if buffer[position] != rune('L') {
												goto l532
											}
											position++
										}
									l582:
										{
											position584, tokenIndex584, depth584 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l585
											}
											position++
											goto l584
										l585:
											position, tokenIndex, depth = position584, tokenIndex584, depth584
											if buffer[position] != rune('E') {
												goto l532
											}
											position++
										}
									l584:
										break
									case 'R', 'r':
										{
											position586, tokenIndex586, depth586 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l587
											}
											position++
											goto l586
										l587:
											position, tokenIndex, depth = position586, tokenIndex586, depth586
											if buffer[position] != rune('R') {
												goto l532
											}
											position++
										}
									l586:
										{
											position588, tokenIndex588, depth588 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l589
											}
											position++
											goto l588
										l589:
											position, tokenIndex, depth = position588, tokenIndex588, depth588
											if buffer[position] != rune('E') {
												goto l532
											}
											position++
										}
									l588:
										{
											position590, tokenIndex590, depth590 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l591
											}
											position++
											goto l590
										l591:
											position, tokenIndex, depth = position590, tokenIndex590, depth590
											if buffer[position] != rune('S') {
												goto l532
											}
											position++
										}
									l590:
										{
											position592, tokenIndex592, depth592 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l593
											}
											position++
											goto l592
										l593:
											position, tokenIndex, depth = position592, tokenIndex592, depth592
											if buffer[position] != rune('O') {
												goto l532
											}
											position++
										}
									l592:
										{
											position594, tokenIndex594, depth594 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l595
											}
											position++
											goto l594
										l595:
											position, tokenIndex, depth = position594, tokenIndex594, depth594
											if buffer[position] != rune('L') {
												goto l532
											}
											position++
										}
									l594:
										{
											position596, tokenIndex596, depth596 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l597
											}
											position++
											goto l596
										l597:
											position, tokenIndex, depth = position596, tokenIndex596, depth596
											if buffer[position] != rune('U') {
												goto l532
											}
											position++
										}
									l596:
										{
											position598, tokenIndex598, depth598 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l599
											}
											position++
											goto l598
										l599:
											position, tokenIndex, depth = position598, tokenIndex598, depth598
											if buffer[position] != rune('T') {
												goto l532
											}
											position++
										}
									l598:
										{
											position600, tokenIndex600, depth600 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l601
											}
											position++
											goto l600
										l601:
											position, tokenIndex, depth = position600, tokenIndex600, depth600
											if buffer[position] != rune('I') {
												goto l532
											}
											position++
										}
									l600:
										{
											position602, tokenIndex602, depth602 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l603
											}
											position++
											goto l602
										l603:
											position, tokenIndex, depth = position602, tokenIndex602, depth602
											if buffer[position] != rune('O') {
												goto l532
											}
											position++
										}
									l602:
										{
											position604, tokenIndex604, depth604 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l605
											}
											position++
											goto l604
										l605:
											position, tokenIndex, depth = position604, tokenIndex604, depth604
											if buffer[position] != rune('N') {
												goto l532
											}
											position++
										}
									l604:
										break
									case 'T', 't':
										{
											position606, tokenIndex606, depth606 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l607
											}
											position++
											goto l606
										l607:
											position, tokenIndex, depth = position606, tokenIndex606, depth606
											if buffer[position] != rune('T') {
												goto l532
											}
											position++
										}
									l606:
										{
											position608, tokenIndex608, depth608 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l609
											}
											position++
											goto l608
										l609:
											position, tokenIndex, depth = position608, tokenIndex608, depth608
											if buffer[position] != rune('O') {
												goto l532
											}
											position++
										}
									l608:
										break
									case 'F', 'f':
										{
											position610, tokenIndex610, depth610 := position, tokenIndex, depth
											if buffer[position] != rune('f') {
												goto l611
											}
											position++
											goto l610
										l611:
											position, tokenIndex, depth = position610, tokenIndex610, depth610
											if buffer[position] != rune('F') {
												goto l532
											}
											position++
										}
									l610:
										{
											position612, tokenIndex612, depth612 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l613
											}
											position++
											goto l612
										l613:
											position, tokenIndex, depth = position612, tokenIndex612, depth612
											if buffer[position] != rune('R') {
												goto l532
											}
											position++
										}
									l612:
										{
											position614, tokenIndex614, depth614 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l615
											}
											position++
											goto l614
										l615:
											position, tokenIndex, depth = position614, tokenIndex614, depth614
											if buffer[position] != rune('O') {
												goto l532
											}
											position++
										}
									l614:
										{
											position616, tokenIndex616, depth616 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l617
											}
											position++
											goto l616
										l617:
											position, tokenIndex, depth = position616, tokenIndex616, depth616
											if buffer[position] != rune('M') {
												goto l532
											}
											position++
										}
									l616:
										break
									case 'M', 'm':
										{
											position618, tokenIndex618, depth618 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l619
											}
											position++
											goto l618
										l619:
											position, tokenIndex, depth = position618, tokenIndex618, depth618
											if buffer[position] != rune('M') {
												goto l532
											}
											position++
										}
									l618:
										{
											position620, tokenIndex620, depth620 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l621
											}
											position++
											goto l620
										l621:
											position, tokenIndex, depth = position620, tokenIndex620, depth620
											if buffer[position] != rune('E') {
												goto l532
											}
											position++
										}
									l620:
										{
											position622, tokenIndex622, depth622 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l623
											}
											position++
											goto l622
										l623:
											position, tokenIndex, depth = position622, tokenIndex622, depth622
											if buffer[position] != rune('T') {
												goto l532
											}
											position++
										}
									l622:
										{
											position624, tokenIndex624, depth624 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l625
											}
											position++
											goto l624
										l625:
											position, tokenIndex, depth = position624, tokenIndex624, depth624
											if buffer[position] != rune('R') {
												goto l532
											}
											position++
										}
									l624:
										{
											position626, tokenIndex626, depth626 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l627
											}
											position++
											goto l626
										l627:
											position, tokenIndex, depth = position626, tokenIndex626, depth626
											if buffer[position] != rune('I') {
												goto l532
											}
											position++
										}
									l626:
										{
											position628, tokenIndex628, depth628 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l629
											}
											position++
											goto l628
										l629:
											position, tokenIndex, depth = position628, tokenIndex628, depth628
											if buffer[position] != rune('C') {
												goto l532
											}
											position++
										}
									l628:
										{
											position630, tokenIndex630, depth630 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l631
											}
											position++
											goto l630
										l631:
											position, tokenIndex, depth = position630, tokenIndex630, depth630
											if buffer[position] != rune('S') {
												goto l532
											}
											position++
										}
									l630:
										break
									case 'W', 'w':
										{
											position632, tokenIndex632, depth632 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l633
											}
											position++
											goto l632
										l633:
											position, tokenIndex, depth = position632, tokenIndex632, depth632
											if buffer[position] != rune('W') {
												goto l532
											}
											position++
										}
									l632:
										{
											position634, tokenIndex634, depth634 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l635
											}
											position++
											goto l634
										l635:
											position, tokenIndex, depth = position634, tokenIndex634, depth634
											if buffer[position] != rune('H') {
												goto l532
											}
											position++
										}
									l634:
										{
											position636, tokenIndex636, depth636 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l637
											}
											position++
											goto l636
										l637:
											position, tokenIndex, depth = position636, tokenIndex636, depth636
											if buffer[position] != rune('E') {
												goto l532
											}
											position++
										}
									l636:
										{
											position638, tokenIndex638, depth638 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l639
											}
											position++
											goto l638
										l639:
											position, tokenIndex, depth = position638, tokenIndex638, depth638
											if buffer[position] != rune('R') {
												goto l532
											}
											position++
										}
									l638:
										{
											position640, tokenIndex640, depth640 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l641
											}
											position++
											goto l640
										l641:
											position, tokenIndex, depth = position640, tokenIndex640, depth640
											if buffer[position] != rune('E') {
												goto l532
											}
											position++
										}
									l640:
										break
									case 'O', 'o':
										{
											position642, tokenIndex642, depth642 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l643
											}
											position++
											goto l642
										l643:
											position, tokenIndex, depth = position642, tokenIndex642, depth642
											if buffer[position] != rune('O') {
												goto l532
											}
											position++
										}
									l642:
										{
											position644, tokenIndex644, depth644 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l645
											}
											position++
											goto l644
										l645:
											position, tokenIndex, depth = position644, tokenIndex644, depth644
											if buffer[position] != rune('R') {
												goto l532
											}
											position++
										}
									l644:
										break
									case 'N', 'n':
										{
											position646, tokenIndex646, depth646 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l647
											}
											position++
											goto l646
										l647:
											position, tokenIndex, depth = position646, tokenIndex646, depth646
											if buffer[position] != rune('N') {
												goto l532
											}
											position++
										}
									l646:
										{
											position648, tokenIndex648, depth648 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l649
											}
											position++
											goto l648
										l649:
											position, tokenIndex, depth = position648, tokenIndex648, depth648
											if buffer[position] != rune('O') {
												goto l532
											}
											position++
										}
									l648:
										{
											position650, tokenIndex650, depth650 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l651
											}
											position++
											goto l650
										l651:
											position, tokenIndex, depth = position650, tokenIndex650, depth650
											if buffer[position] != rune('T') {
												goto l532
											}
											position++
										}
									l650:
										break
									case 'I', 'i':
										{
											position652, tokenIndex652, depth652 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l653
											}
											position++
											goto l652
										l653:
											position, tokenIndex, depth = position652, tokenIndex652, depth652
											if buffer[position] != rune('I') {
												goto l532
											}
											position++
										}
									l652:
										{
											position654, tokenIndex654, depth654 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l655
											}
											position++
											goto l654
										l655:
											position, tokenIndex, depth = position654, tokenIndex654, depth654
											if buffer[position] != rune('N') {
												goto l532
											}
											position++
										}
									l654:
										break
									case 'C', 'c':
										{
											position656, tokenIndex656, depth656 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l657
											}
											position++
											goto l656
										l657:
											position, tokenIndex, depth = position656, tokenIndex656, depth656
											if buffer[position] != rune('C') {
												goto l532
											}
											position++
										}
									l656:
										{
											position658, tokenIndex658, depth658 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l659
											}
											position++
											goto l658
										l659:
											position, tokenIndex, depth = position658, tokenIndex658, depth658
											if buffer[position] != rune('O') {
												goto l532
											}
											position++
										}
									l658:
										{
											position660, tokenIndex660, depth660 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l661
											}
											position++
											goto l660
										l661:
											position, tokenIndex, depth = position660, tokenIndex660, depth660
											if buffer[position] != rune('L') {
												goto l532
											}
											position++
										}
									l660:
										{
											position662, tokenIndex662, depth662 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l663
											}
											position++
											goto l662
										l663:
											position, tokenIndex, depth = position662, tokenIndex662, depth662
											if buffer[position] != rune('L') {
												goto l532
											}
											position++
										}
									l662:
										{
											position664, tokenIndex664, depth664 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l665
											}
											position++
											goto l664
										l665:
											position, tokenIndex, depth = position664, tokenIndex664, depth664
											if buffer[position] != rune('A') {
												goto l532
											}
											position++
										}
									l664:
										{
											position666, tokenIndex666, depth666 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l667
											}
											position++
											goto l666
										l667:
											position, tokenIndex, depth = position666, tokenIndex666, depth666
											if buffer[position] != rune('P') {
												goto l532
											}
											position++
										}
									l666:
										{
											position668, tokenIndex668, depth668 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l669
											}
											position++
											goto l668
										l669:
											position, tokenIndex, depth = position668, tokenIndex668, depth668
											if buffer[position] != rune('S') {
												goto l532
											}
											position++
										}
									l668:
										{
											position670, tokenIndex670, depth670 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l671
											}
											position++
											goto l670
										l671:
											position, tokenIndex, depth = position670, tokenIndex670, depth670
											if buffer[position] != rune('E') {
												goto l532
											}
											position++
										}
									l670:
										break
									case 'G', 'g':
										{
											position672, tokenIndex672, depth672 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l673
											}
											position++
											goto l672
										l673:
											position, tokenIndex, depth = position672, tokenIndex672, depth672
											if buffer[position] != rune('G') {
												goto l532
											}
											position++
										}
									l672:
										{
											position674, tokenIndex674, depth674 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l675
											}
											position++
											goto l674
										l675:
											position, tokenIndex, depth = position674, tokenIndex674, depth674
											if buffer[position] != rune('R') {
												goto l532
											}
											position++
										}
									l674:
										{
											position676, tokenIndex676, depth676 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l677
											}
											position++
											goto l676
										l677:
											position, tokenIndex, depth = position676, tokenIndex676, depth676
											if buffer[position] != rune('O') {
												goto l532
											}
											position++
										}
									l676:
										{
											position678, tokenIndex678, depth678 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l679
											}
											position++
											goto l678
										l679:
											position, tokenIndex, depth = position678, tokenIndex678, depth678
											if buffer[position] != rune('U') {
												goto l532
											}
											position++
										}
									l678:
										{
											position680, tokenIndex680, depth680 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l681
											}
											position++
											goto l680
										l681:
											position, tokenIndex, depth = position680, tokenIndex680, depth680
											if buffer[position] != rune('P') {
												goto l532
											}
											position++
										}
									l680:
										break
									case 'D', 'd':
										{
											position682, tokenIndex682, depth682 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l683
											}
											position++
											goto l682
										l683:
											position, tokenIndex, depth = position682, tokenIndex682, depth682
											if buffer[position] != rune('D') {
												goto l532
											}
											position++
										}
									l682:
										{
											position684, tokenIndex684, depth684 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l685
											}
											position++
											goto l684
										l685:
											position, tokenIndex, depth = position684, tokenIndex684, depth684
											if buffer[position] != rune('E') {
												goto l532
											}
											position++
										}
									l684:
										{
											position686, tokenIndex686, depth686 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l687
											}
											position++
											goto l686
										l687:
											position, tokenIndex, depth = position686, tokenIndex686, depth686
											if buffer[position] != rune('S') {
												goto l532
											}
											position++
										}
									l686:
										{
											position688, tokenIndex688, depth688 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l689
											}
											position++
											goto l688
										l689:
											position, tokenIndex, depth = position688, tokenIndex688, depth688
											if buffer[position] != rune('C') {
												goto l532
											}
											position++
										}
									l688:
										{
											position690, tokenIndex690, depth690 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l691
											}
											position++
											goto l690
										l691:
											position, tokenIndex, depth = position690, tokenIndex690, depth690
											if buffer[position] != rune('R') {
												goto l532
											}
											position++
										}
									l690:
										{
											position692, tokenIndex692, depth692 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l693
											}
											position++
											goto l692
										l693:
											position, tokenIndex, depth = position692, tokenIndex692, depth692
											if buffer[position] != rune('I') {
												goto l532
											}
											position++
										}
									l692:
										{
											position694, tokenIndex694, depth694 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l695
											}
											position++
											goto l694
										l695:
											position, tokenIndex, depth = position694, tokenIndex694, depth694
											if buffer[position] != rune('B') {
												goto l532
											}
											position++
										}
									l694:
										{
											position696, tokenIndex696, depth696 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l697
											}
											position++
											goto l696
										l697:
											position, tokenIndex, depth = position696, tokenIndex696, depth696
											if buffer[position] != rune('E') {
												goto l532
											}
											position++
										}
									l696:
										break
									case 'B', 'b':
										{
											position698, tokenIndex698, depth698 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l699
											}
											position++
											goto l698
										l699:
											position, tokenIndex, depth = position698, tokenIndex698, depth698
											if buffer[position] != rune('B') {
												goto l532
											}
											position++
										}
									l698:
										{
											position700, tokenIndex700, depth700 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l701
											}
											position++
											goto l700
										l701:
											position, tokenIndex, depth = position700, tokenIndex700, depth700
											if buffer[position] != rune('Y') {
												goto l532
											}
											position++
										}
									l700:
										break
									default:
										{
											position702, tokenIndex702, depth702 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l703
											}
											position++
											goto l702
										l703:
											position, tokenIndex, depth = position702, tokenIndex702, depth702
											if buffer[position] != rune('A') {
												goto l532
											}
											position++
										}
									l702:
										{
											position704, tokenIndex704, depth704 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l705
											}
											position++
											goto l704
										l705:
											position, tokenIndex, depth = position704, tokenIndex704, depth704
											if buffer[position] != rune('S') {
												goto l532
											}
											position++
										}
									l704:
										break
									}
								}

							}
						l534:
							depth--
							add(ruleKEYWORD, position533)
						}
						if !_rules[ruleKEY]() {
							goto l532
						}
						goto l524
					l532:
						position, tokenIndex, depth = position532, tokenIndex532, depth532
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l524
					}
				l706:
					{
						position707, tokenIndex707, depth707 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l707
						}
						position++
						{
							position708, tokenIndex708, depth708 := position, tokenIndex, depth
							if !_rules[ruleID_SEGMENT]() {
								goto l709
							}
							goto l708
						l709:
							position, tokenIndex, depth = position708, tokenIndex708, depth708
							if !(p.errorHere(position, `expected identifier segment to follow "."`)) {
								goto l707
							}
						}
					l708:
						goto l706
					l707:
						position, tokenIndex, depth = position707, tokenIndex707, depth707
					}
				}
			l526:
				depth--
				add(ruleIDENTIFIER, position525)
			}
			return true
		l524:
			position, tokenIndex, depth = position524, tokenIndex524, depth524
			return false
		},
		/* 38 TIMESTAMP <- <((_ <(NUMBER ([a-z] / [A-Z])*)>) / (_ STRING) / (_ <(('n' / 'N') ('o' / 'O') ('w' / 'W'))> KEY))> */
		nil,
		/* 39 ID_SEGMENT <- <(ID_START ID_CONT*)> */
		func() bool {
			position711, tokenIndex711, depth711 := position, tokenIndex, depth
			{
				position712 := position
				depth++
				if !_rules[ruleID_START]() {
					goto l711
				}
			l713:
				{
					position714, tokenIndex714, depth714 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l714
					}
					goto l713
				l714:
					position, tokenIndex, depth = position714, tokenIndex714, depth714
				}
				depth--
				add(ruleID_SEGMENT, position712)
			}
			return true
		l711:
			position, tokenIndex, depth = position711, tokenIndex711, depth711
			return false
		},
		/* 40 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position715, tokenIndex715, depth715 := position, tokenIndex, depth
			{
				position716 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l715
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l715
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l715
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position716)
			}
			return true
		l715:
			position, tokenIndex, depth = position715, tokenIndex715, depth715
			return false
		},
		/* 41 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position718, tokenIndex718, depth718 := position, tokenIndex, depth
			{
				position719 := position
				depth++
				{
					position720, tokenIndex720, depth720 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l721
					}
					goto l720
				l721:
					position, tokenIndex, depth = position720, tokenIndex720, depth720
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l718
					}
					position++
				}
			l720:
				depth--
				add(ruleID_CONT, position719)
			}
			return true
		l718:
			position, tokenIndex, depth = position718, tokenIndex718, depth718
			return false
		},
		/* 42 PROPERTY_KEY <- <((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> KEY ((_ (('b' / 'B') ('y' / 'Y')) KEY) / &{ p.errorHere(position, `expected keyword "by" to follow keyword "sample"`) }))) | (&('R' | 'r') (<(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))> KEY)) | (&('T' | 't') (<(('t' / 'T') ('o' / 'O'))> KEY)) | (&('F' | 'f') (<(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))> KEY)))> */
		nil,
		/* 43 PROPERTY_VALUE <- <TIMESTAMP> */
		nil,
		/* 44 KEYWORD <- <((('a' / 'A') ('l' / 'L') ('l' / 'L')) / (('a' / 'A') ('n' / 'N') ('d' / 'D')) / (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H')) / (('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T')) / ((&('S' | 's') (('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))) | (&('R' | 'r') (('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))) | (&('T' | 't') (('t' / 'T') ('o' / 'O'))) | (&('F' | 'f') (('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))) | (&('M' | 'm') (('m' / 'M') ('e' / 'E') ('t' / 'T') ('r' / 'R') ('i' / 'I') ('c' / 'C') ('s' / 'S'))) | (&('W' | 'w') (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E'))) | (&('O' | 'o') (('o' / 'O') ('r' / 'R'))) | (&('N' | 'n') (('n' / 'N') ('o' / 'O') ('t' / 'T'))) | (&('I' | 'i') (('i' / 'I') ('n' / 'N'))) | (&('C' | 'c') (('c' / 'C') ('o' / 'O') ('l' / 'L') ('l' / 'L') ('a' / 'A') ('p' / 'P') ('s' / 'S') ('e' / 'E'))) | (&('G' | 'g') (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P'))) | (&('D' | 'd') (('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E'))) | (&('B' | 'b') (('b' / 'B') ('y' / 'Y'))) | (&('A' | 'a') (('a' / 'A') ('s' / 'S')))))> */
		nil,
		/* 45 OP_PIPE <- <'|'> */
		nil,
		/* 46 OP_ADD <- <'+'> */
		nil,
		/* 47 OP_SUB <- <'-'> */
		nil,
		/* 48 OP_MULT <- <'*'> */
		nil,
		/* 49 OP_DIV <- <'/'> */
		nil,
		/* 50 OP_AND <- <(('a' / 'A') ('n' / 'N') ('d' / 'D') KEY)> */
		nil,
		/* 51 OP_OR <- <(('o' / 'O') ('r' / 'R') KEY)> */
		nil,
		/* 52 OP_NOT <- <(('n' / 'N') ('o' / 'O') ('t' / 'T') KEY)> */
		nil,
		/* 53 QUOTE_SINGLE <- <'\''> */
		func() bool {
			position733, tokenIndex733, depth733 := position, tokenIndex, depth
			{
				position734 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l733
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position734)
			}
			return true
		l733:
			position, tokenIndex, depth = position733, tokenIndex733, depth733
			return false
		},
		/* 54 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position735, tokenIndex735, depth735 := position, tokenIndex, depth
			{
				position736 := position
				depth++
				if buffer[position] != rune('"') {
					goto l735
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position736)
			}
			return true
		l735:
			position, tokenIndex, depth = position735, tokenIndex735, depth735
			return false
		},
		/* 55 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> (QUOTE_SINGLE / &{ p.errorHere(position, `expected "'" to close string`) })) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> (QUOTE_DOUBLE / &{ p.errorHere(position, `expected '"' to close string`) })))> */
		func() bool {
			position737, tokenIndex737, depth737 := position, tokenIndex, depth
			{
				position738 := position
				depth++
				{
					position739, tokenIndex739, depth739 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l740
					}
					{
						position741 := position
						depth++
					l742:
						{
							position743, tokenIndex743, depth743 := position, tokenIndex, depth
							{
								position744, tokenIndex744, depth744 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l744
								}
								goto l743
							l744:
								position, tokenIndex, depth = position744, tokenIndex744, depth744
							}
							if !_rules[ruleCHAR]() {
								goto l743
							}
							goto l742
						l743:
							position, tokenIndex, depth = position743, tokenIndex743, depth743
						}
						depth--
						add(rulePegText, position741)
					}
					{
						position745, tokenIndex745, depth745 := position, tokenIndex, depth
						if !_rules[ruleQUOTE_SINGLE]() {
							goto l746
						}
						goto l745
					l746:
						position, tokenIndex, depth = position745, tokenIndex745, depth745
						if !(p.errorHere(position, `expected "'" to close string`)) {
							goto l740
						}
					}
				l745:
					goto l739
				l740:
					position, tokenIndex, depth = position739, tokenIndex739, depth739
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l737
					}
					{
						position747 := position
						depth++
					l748:
						{
							position749, tokenIndex749, depth749 := position, tokenIndex, depth
							{
								position750, tokenIndex750, depth750 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l750
								}
								goto l749
							l750:
								position, tokenIndex, depth = position750, tokenIndex750, depth750
							}
							if !_rules[ruleCHAR]() {
								goto l749
							}
							goto l748
						l749:
							position, tokenIndex, depth = position749, tokenIndex749, depth749
						}
						depth--
						add(rulePegText, position747)
					}
					{
						position751, tokenIndex751, depth751 := position, tokenIndex, depth
						if !_rules[ruleQUOTE_DOUBLE]() {
							goto l752
						}
						goto l751
					l752:
						position, tokenIndex, depth = position751, tokenIndex751, depth751
						if !(p.errorHere(position, `expected '"' to close string`)) {
							goto l737
						}
					}
				l751:
				}
			l739:
				depth--
				add(ruleSTRING, position738)
			}
			return true
		l737:
			position, tokenIndex, depth = position737, tokenIndex737, depth737
			return false
		},
		/* 56 CHAR <- <(('\\' ((&('"') (QUOTE_DOUBLE / &{ p.errorHere(position, "expected \"\\\", \"'\", \"`\", or '\"' to follow \"\\\" in string literal") })) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position753, tokenIndex753, depth753 := position, tokenIndex, depth
			{
				position754 := position
				depth++
				{
					position755, tokenIndex755, depth755 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l756
					}
					position++
					{
						switch buffer[position] {
						case '"':
							{
								position758, tokenIndex758, depth758 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l759
								}
								goto l758
							l759:
								position, tokenIndex, depth = position758, tokenIndex758, depth758
								if !(p.errorHere(position, "expected \"\\\", \"'\", \"`\", or '\"' to follow \"\\\" in string literal")) {
									goto l756
								}
							}
						l758:
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l756
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l756
							}
							break
						}
					}

					goto l755
				l756:
					position, tokenIndex, depth = position755, tokenIndex755, depth755
					{
						position760, tokenIndex760, depth760 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l760
						}
						goto l753
					l760:
						position, tokenIndex, depth = position760, tokenIndex760, depth760
					}
					if !matchDot() {
						goto l753
					}
				}
			l755:
				depth--
				add(ruleCHAR, position754)
			}
			return true
		l753:
			position, tokenIndex, depth = position753, tokenIndex753, depth753
			return false
		},
		/* 57 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position761, tokenIndex761, depth761 := position, tokenIndex, depth
			{
				position762 := position
				depth++
				{
					position763, tokenIndex763, depth763 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l764
					}
					position++
					goto l763
				l764:
					position, tokenIndex, depth = position763, tokenIndex763, depth763
					if buffer[position] != rune('\\') {
						goto l761
					}
					position++
				}
			l763:
				depth--
				add(ruleESCAPE_CLASS, position762)
			}
			return true
		l761:
			position, tokenIndex, depth = position761, tokenIndex761, depth761
			return false
		},
		/* 58 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position765, tokenIndex765, depth765 := position, tokenIndex, depth
			{
				position766 := position
				depth++
				{
					position767 := position
					depth++
					{
						position768, tokenIndex768, depth768 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l768
						}
						position++
						goto l769
					l768:
						position, tokenIndex, depth = position768, tokenIndex768, depth768
					}
				l769:
					{
						position770 := position
						depth++
						{
							position771, tokenIndex771, depth771 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l772
							}
							position++
							goto l771
						l772:
							position, tokenIndex, depth = position771, tokenIndex771, depth771
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l765
							}
							position++
						l773:
							{
								position774, tokenIndex774, depth774 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l774
								}
								position++
								goto l773
							l774:
								position, tokenIndex, depth = position774, tokenIndex774, depth774
							}
						}
					l771:
						depth--
						add(ruleNUMBER_NATURAL, position770)
					}
					depth--
					add(ruleNUMBER_INTEGER, position767)
				}
				{
					position775, tokenIndex775, depth775 := position, tokenIndex, depth
					{
						position777 := position
						depth++
						if buffer[position] != rune('.') {
							goto l775
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l775
						}
						position++
					l778:
						{
							position779, tokenIndex779, depth779 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l779
							}
							position++
							goto l778
						l779:
							position, tokenIndex, depth = position779, tokenIndex779, depth779
						}
						depth--
						add(ruleNUMBER_FRACTION, position777)
					}
					goto l776
				l775:
					position, tokenIndex, depth = position775, tokenIndex775, depth775
				}
			l776:
				{
					position780, tokenIndex780, depth780 := position, tokenIndex, depth
					{
						position782 := position
						depth++
						{
							position783, tokenIndex783, depth783 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l784
							}
							position++
							goto l783
						l784:
							position, tokenIndex, depth = position783, tokenIndex783, depth783
							if buffer[position] != rune('E') {
								goto l780
							}
							position++
						}
					l783:
						{
							position785, tokenIndex785, depth785 := position, tokenIndex, depth
							{
								position787, tokenIndex787, depth787 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l788
								}
								position++
								goto l787
							l788:
								position, tokenIndex, depth = position787, tokenIndex787, depth787
								if buffer[position] != rune('-') {
									goto l785
								}
								position++
							}
						l787:
							goto l786
						l785:
							position, tokenIndex, depth = position785, tokenIndex785, depth785
						}
					l786:
						{
							position789, tokenIndex789, depth789 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l790
							}
							position++
						l791:
							{
								position792, tokenIndex792, depth792 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l792
								}
								position++
								goto l791
							l792:
								position, tokenIndex, depth = position792, tokenIndex792, depth792
							}
							goto l789
						l790:
							position, tokenIndex, depth = position789, tokenIndex789, depth789
							if !(p.errorHere(position, `expected exponent`)) {
								goto l780
							}
						}
					l789:
						depth--
						add(ruleNUMBER_EXP, position782)
					}
					goto l781
				l780:
					position, tokenIndex, depth = position780, tokenIndex780, depth780
				}
			l781:
				depth--
				add(ruleNUMBER, position766)
			}
			return true
		l765:
			position, tokenIndex, depth = position765, tokenIndex765, depth765
			return false
		},
		/* 59 NUMBER_NATURAL <- <('0' / ([1-9] [0-9]*))> */
		nil,
		/* 60 NUMBER_FRACTION <- <('.' [0-9]+)> */
		nil,
		/* 61 NUMBER_INTEGER <- <('-'? NUMBER_NATURAL)> */
		nil,
		/* 62 NUMBER_EXP <- <(('e' / 'E') ('+' / '-')? ([0-9]+ / &{ p.errorHere(position, `expected exponent`) }))> */
		nil,
		/* 63 DURATION <- <(NUMBER [a-z]+ KEY)> */
		nil,
		/* 64 PAREN_OPEN <- <'('> */
		func() bool {
			position798, tokenIndex798, depth798 := position, tokenIndex, depth
			{
				position799 := position
				depth++
				if buffer[position] != rune('(') {
					goto l798
				}
				position++
				depth--
				add(rulePAREN_OPEN, position799)
			}
			return true
		l798:
			position, tokenIndex, depth = position798, tokenIndex798, depth798
			return false
		},
		/* 65 PAREN_CLOSE <- <')'> */
		func() bool {
			position800, tokenIndex800, depth800 := position, tokenIndex, depth
			{
				position801 := position
				depth++
				if buffer[position] != rune(')') {
					goto l800
				}
				position++
				depth--
				add(rulePAREN_CLOSE, position801)
			}
			return true
		l800:
			position, tokenIndex, depth = position800, tokenIndex800, depth800
			return false
		},
		/* 66 COMMA <- <','> */
		func() bool {
			position802, tokenIndex802, depth802 := position, tokenIndex, depth
			{
				position803 := position
				depth++
				if buffer[position] != rune(',') {
					goto l802
				}
				position++
				depth--
				add(ruleCOMMA, position803)
			}
			return true
		l802:
			position, tokenIndex, depth = position802, tokenIndex802, depth802
			return false
		},
		/* 67 _ <- <((&('/') COMMENT_BLOCK) | (&('-') COMMENT_TRAIL) | (&('\t' | '\n' | ' ') SPACE))*> */
		func() bool {
			{
				position805 := position
				depth++
			l806:
				{
					position807, tokenIndex807, depth807 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '/':
							{
								position809 := position
								depth++
								if buffer[position] != rune('/') {
									goto l807
								}
								position++
								if buffer[position] != rune('*') {
									goto l807
								}
								position++
							l810:
								{
									position811, tokenIndex811, depth811 := position, tokenIndex, depth
									{
										position812, tokenIndex812, depth812 := position, tokenIndex, depth
										if buffer[position] != rune('*') {
											goto l812
										}
										position++
										if buffer[position] != rune('/') {
											goto l812
										}
										position++
										goto l811
									l812:
										position, tokenIndex, depth = position812, tokenIndex812, depth812
									}
									if !matchDot() {
										goto l811
									}
									goto l810
								l811:
									position, tokenIndex, depth = position811, tokenIndex811, depth811
								}
								if buffer[position] != rune('*') {
									goto l807
								}
								position++
								if buffer[position] != rune('/') {
									goto l807
								}
								position++
								depth--
								add(ruleCOMMENT_BLOCK, position809)
							}
							break
						case '-':
							{
								position813 := position
								depth++
								if buffer[position] != rune('-') {
									goto l807
								}
								position++
								if buffer[position] != rune('-') {
									goto l807
								}
								position++
							l814:
								{
									position815, tokenIndex815, depth815 := position, tokenIndex, depth
									{
										position816, tokenIndex816, depth816 := position, tokenIndex, depth
										if buffer[position] != rune('\n') {
											goto l816
										}
										position++
										goto l815
									l816:
										position, tokenIndex, depth = position816, tokenIndex816, depth816
									}
									if !matchDot() {
										goto l815
									}
									goto l814
								l815:
									position, tokenIndex, depth = position815, tokenIndex815, depth815
								}
								depth--
								add(ruleCOMMENT_TRAIL, position813)
							}
							break
						default:
							{
								position817 := position
								depth++
								{
									switch buffer[position] {
									case '\t':
										if buffer[position] != rune('\t') {
											goto l807
										}
										position++
										break
									case '\n':
										if buffer[position] != rune('\n') {
											goto l807
										}
										position++
										break
									default:
										if buffer[position] != rune(' ') {
											goto l807
										}
										position++
										break
									}
								}

								depth--
								add(ruleSPACE, position817)
							}
							break
						}
					}

					goto l806
				l807:
					position, tokenIndex, depth = position807, tokenIndex807, depth807
				}
				depth--
				add(rule_, position805)
			}
			return true
		},
		/* 68 COMMENT_TRAIL <- <('-' '-' (!'\n' .)*)> */
		nil,
		/* 69 COMMENT_BLOCK <- <('/' '*' (!('*' '/') .)* ('*' '/'))> */
		nil,
		/* 70 KEY <- <!ID_CONT> */
		func() bool {
			position821, tokenIndex821, depth821 := position, tokenIndex, depth
			{
				position822 := position
				depth++
				{
					position823, tokenIndex823, depth823 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l823
					}
					goto l821
				l823:
					position, tokenIndex, depth = position823, tokenIndex823, depth823
				}
				depth--
				add(ruleKEY, position822)
			}
			return true
		l821:
			position, tokenIndex, depth = position821, tokenIndex821, depth821
			return false
		},
		/* 71 SPACE <- <((&('\t') '\t') | (&('\n') '\n') | (&(' ') ' '))> */
		nil,
		/* 73 Action0 <- <{ p.makeSelect() }> */
		nil,
		/* 74 Action1 <- <{ p.makeDescribeAll() }> */
		nil,
		/* 75 Action2 <- <{ p.addNullMatchClause() }> */
		nil,
		/* 76 Action3 <- <{ p.addMatchClause() }> */
		nil,
		/* 77 Action4 <- <{ p.makeDescribeMetrics() }> */
		nil,
		nil,
		/* 79 Action5 <- <{ p.pushString(unescapeLiteral(text)) }> */
		nil,
		/* 80 Action6 <- <{ p.makeDescribe() }> */
		nil,
		/* 81 Action7 <- <{ p.addEvaluationContext() }> */
		nil,
		/* 82 Action8 <- <{ p.addPropertyKey(text) }> */
		nil,
		/* 83 Action9 <- <{
		   p.addPropertyValue(text) }> */
		nil,
		/* 84 Action10 <- <{ p.insertPropertyKeyValue() }> */
		nil,
		/* 85 Action11 <- <{ p.checkPropertyClause() }> */
		nil,
		/* 86 Action12 <- <{ p.addNullPredicate() }> */
		nil,
		/* 87 Action13 <- <{ p.addExpressionList() }> */
		nil,
		/* 88 Action14 <- <{ p.appendExpression() }> */
		nil,
		/* 89 Action15 <- <{ p.appendExpression() }> */
		nil,
		/* 90 Action16 <- <{ p.addOperatorLiteral("+") }> */
		nil,
		/* 91 Action17 <- <{ p.addOperatorLiteral("-") }> */
		nil,
		/* 92 Action18 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 93 Action19 <- <{ p.addOperatorLiteral("/") }> */
		nil,
		/* 94 Action20 <- <{ p.addOperatorLiteral("*") }> */
		nil,
		/* 95 Action21 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 96 Action22 <- <{ p.pushString(unescapeLiteral(text)) }> */
		nil,
		/* 97 Action23 <- <{p.addExpressionList()}> */
		nil,
		/* 98 Action24 <- <{
		   p.addExpressionList()
		   p.addGroupBy()
		 }> */
		nil,
		/* 99 Action25 <- <{ p.addPipeExpression() }> */
		nil,
		/* 100 Action26 <- <{ p.addDurationNode(text) }> */
		nil,
		/* 101 Action27 <- <{ p.addNumberNode(text) }> */
		nil,
		/* 102 Action28 <- <{ p.addStringNode(unescapeLiteral(text)) }> */
		nil,
		/* 103 Action29 <- <{ p.addAnnotationExpression(text) }> */
		nil,
		/* 104 Action30 <- <{ p.addGroupBy() }> */
		nil,
		/* 105 Action31 <- <{ p.pushString(unescapeLiteral(text)) }> */
		nil,
		/* 106 Action32 <- <{ p.addFunctionInvocation() }> */
		nil,
		/* 107 Action33 <- <{ p.pushString(unescapeLiteral(text)) }> */
		nil,
		/* 108 Action34 <- <{ p.addNullPredicate() }> */
		nil,
		/* 109 Action35 <- <{ p.addMetricExpression() }> */
		nil,
		/* 110 Action36 <- <{ p.appendGroupBy(unescapeLiteral(text)) }> */
		nil,
		/* 111 Action37 <- <{ p.appendGroupBy(unescapeLiteral(text)) }> */
		nil,
		/* 112 Action38 <- <{ p.appendCollapseBy(unescapeLiteral(text)) }> */
		nil,
		/* 113 Action39 <- <{ p.appendCollapseBy(unescapeLiteral(text)) }> */
		nil,
		/* 114 Action40 <- <{ p.addOrPredicate() }> */
		nil,
		/* 115 Action41 <- <{ p.addAndPredicate() }> */
		nil,
		/* 116 Action42 <- <{ p.addNotPredicate() }> */
		nil,
		/* 117 Action43 <- <{ p.addLiteralMatcher() }> */
		nil,
		/* 118 Action44 <- <{ p.addLiteralMatcher() }> */
		nil,
		/* 119 Action45 <- <{ p.addNotPredicate() }> */
		nil,
		/* 120 Action46 <- <{ p.addRegexMatcher() }> */
		nil,
		/* 121 Action47 <- <{ p.addListMatcher() }> */
		nil,
		/* 122 Action48 <- <{ p.pushString(unescapeLiteral(text)) }> */
		nil,
		/* 123 Action49 <- <{ p.addLiteralList() }> */
		nil,
		/* 124 Action50 <- <{ p.appendLiteral(unescapeLiteral(text)) }> */
		nil,
		/* 125 Action51 <- <{ p.addTagLiteral(unescapeLiteral(text)) }> */
		nil,
	}
	p.rules = _rules
}
