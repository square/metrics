package parser

import (
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/square/metrics/query/command"
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
	ruleAction52
	ruleAction53

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
	"Action52",
	"Action53",

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
	rules  [128]func() bool
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
			p.addGroupBy()
		case ruleAction37:
			p.appendGroupTag(unescapeLiteral(text))
		case ruleAction38:
			p.appendGroupTag(unescapeLiteral(text))
		case ruleAction39:
			p.addCollapseBy()
		case ruleAction40:
			p.appendGroupTag(unescapeLiteral(text))
		case ruleAction41:
			p.appendGroupTag(unescapeLiteral(text))
		case ruleAction42:
			p.addOrPredicate()
		case ruleAction43:
			p.addAndPredicate()
		case ruleAction44:
			p.addNotPredicate()
		case ruleAction45:
			p.addLiteralMatcher()
		case ruleAction46:
			p.addLiteralMatcher()
		case ruleAction47:
			p.addNotPredicate()
		case ruleAction48:
			p.addRegexMatcher()
		case ruleAction49:
			p.addListMatcher()
		case ruleAction50:
			p.pushString(unescapeLiteral(text))
		case ruleAction51:
			p.addLiteralList()
		case ruleAction52:
			p.appendLiteral(unescapeLiteral(text))
		case ruleAction53:
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
		/* 0 root <- <((describeStmt / selectStmt) _ !.)> */
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
							if buffer[position] != rune('d') {
								goto l6
							}
							position++
							goto l5
						l6:
							position, tokenIndex, depth = position5, tokenIndex5, depth5
							if buffer[position] != rune('D') {
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
							if buffer[position] != rune('s') {
								goto l10
							}
							position++
							goto l9
						l10:
							position, tokenIndex, depth = position9, tokenIndex9, depth9
							if buffer[position] != rune('S') {
								goto l3
							}
							position++
						}
					l9:
						{
							position11, tokenIndex11, depth11 := position, tokenIndex, depth
							if buffer[position] != rune('c') {
								goto l12
							}
							position++
							goto l11
						l12:
							position, tokenIndex, depth = position11, tokenIndex11, depth11
							if buffer[position] != rune('C') {
								goto l3
							}
							position++
						}
					l11:
						{
							position13, tokenIndex13, depth13 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l14
							}
							position++
							goto l13
						l14:
							position, tokenIndex, depth = position13, tokenIndex13, depth13
							if buffer[position] != rune('R') {
								goto l3
							}
							position++
						}
					l13:
						{
							position15, tokenIndex15, depth15 := position, tokenIndex, depth
							if buffer[position] != rune('i') {
								goto l16
							}
							position++
							goto l15
						l16:
							position, tokenIndex, depth = position15, tokenIndex15, depth15
							if buffer[position] != rune('I') {
								goto l3
							}
							position++
						}
					l15:
						{
							position17, tokenIndex17, depth17 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l18
							}
							position++
							goto l17
						l18:
							position, tokenIndex, depth = position17, tokenIndex17, depth17
							if buffer[position] != rune('B') {
								goto l3
							}
							position++
						}
					l17:
						{
							position19, tokenIndex19, depth19 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l20
							}
							position++
							goto l19
						l20:
							position, tokenIndex, depth = position19, tokenIndex19, depth19
							if buffer[position] != rune('E') {
								goto l3
							}
							position++
						}
					l19:
						if !_rules[ruleKEY]() {
							goto l3
						}
						{
							position21, tokenIndex21, depth21 := position, tokenIndex, depth
							{
								position23 := position
								depth++
								if !_rules[rule_]() {
									goto l22
								}
								{
									position24, tokenIndex24, depth24 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l25
									}
									position++
									goto l24
								l25:
									position, tokenIndex, depth = position24, tokenIndex24, depth24
									if buffer[position] != rune('A') {
										goto l22
									}
									position++
								}
							l24:
								{
									position26, tokenIndex26, depth26 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l27
									}
									position++
									goto l26
								l27:
									position, tokenIndex, depth = position26, tokenIndex26, depth26
									if buffer[position] != rune('L') {
										goto l22
									}
									position++
								}
							l26:
								{
									position28, tokenIndex28, depth28 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l29
									}
									position++
									goto l28
								l29:
									position, tokenIndex, depth = position28, tokenIndex28, depth28
									if buffer[position] != rune('L') {
										goto l22
									}
									position++
								}
							l28:
								if !_rules[ruleKEY]() {
									goto l22
								}
								{
									position30 := position
									depth++
									{
										position31, tokenIndex31, depth31 := position, tokenIndex, depth
										{
											position33 := position
											depth++
											if !_rules[rule_]() {
												goto l32
											}
											{
												position34, tokenIndex34, depth34 := position, tokenIndex, depth
												if buffer[position] != rune('m') {
													goto l35
												}
												position++
												goto l34
											l35:
												position, tokenIndex, depth = position34, tokenIndex34, depth34
												if buffer[position] != rune('M') {
													goto l32
												}
												position++
											}
										l34:
											{
												position36, tokenIndex36, depth36 := position, tokenIndex, depth
												if buffer[position] != rune('a') {
													goto l37
												}
												position++
												goto l36
											l37:
												position, tokenIndex, depth = position36, tokenIndex36, depth36
												if buffer[position] != rune('A') {
													goto l32
												}
												position++
											}
										l36:
											{
												position38, tokenIndex38, depth38 := position, tokenIndex, depth
												if buffer[position] != rune('t') {
													goto l39
												}
												position++
												goto l38
											l39:
												position, tokenIndex, depth = position38, tokenIndex38, depth38
												if buffer[position] != rune('T') {
													goto l32
												}
												position++
											}
										l38:
											{
												position40, tokenIndex40, depth40 := position, tokenIndex, depth
												if buffer[position] != rune('c') {
													goto l41
												}
												position++
												goto l40
											l41:
												position, tokenIndex, depth = position40, tokenIndex40, depth40
												if buffer[position] != rune('C') {
													goto l32
												}
												position++
											}
										l40:
											{
												position42, tokenIndex42, depth42 := position, tokenIndex, depth
												if buffer[position] != rune('h') {
													goto l43
												}
												position++
												goto l42
											l43:
												position, tokenIndex, depth = position42, tokenIndex42, depth42
												if buffer[position] != rune('H') {
													goto l32
												}
												position++
											}
										l42:
											if !_rules[ruleKEY]() {
												goto l32
											}
											{
												position44, tokenIndex44, depth44 := position, tokenIndex, depth
												if !_rules[ruleliteralString]() {
													goto l45
												}
												goto l44
											l45:
												position, tokenIndex, depth = position44, tokenIndex44, depth44
												if !(p.errorHere(position, `expected string literal to follow keyword "match"`)) {
													goto l32
												}
											}
										l44:
											{
												add(ruleAction3, position)
											}
											depth--
											add(rulematchClause, position33)
										}
										goto l31
									l32:
										position, tokenIndex, depth = position31, tokenIndex31, depth31
										{
											add(ruleAction2, position)
										}
									}
								l31:
									depth--
									add(ruleoptionalMatchClause, position30)
								}
								{
									add(ruleAction1, position)
								}
								{
									position49, tokenIndex49, depth49 := position, tokenIndex, depth
									{
										position50, tokenIndex50, depth50 := position, tokenIndex, depth
										if !_rules[rule_]() {
											goto l51
										}
										{
											position52, tokenIndex52, depth52 := position, tokenIndex, depth
											if !matchDot() {
												goto l52
											}
											goto l51
										l52:
											position, tokenIndex, depth = position52, tokenIndex52, depth52
										}
										goto l50
									l51:
										position, tokenIndex, depth = position50, tokenIndex50, depth50
										if !_rules[rule_]() {
											goto l22
										}
										if !(p.errorHere(position, `expected end of input after 'describe all' and optional match clause but got %q`, p.after(position))) {
											goto l22
										}
									}
								l50:
									position, tokenIndex, depth = position49, tokenIndex49, depth49
								}
								depth--
								add(ruledescribeAllStmt, position23)
							}
							goto l21
						l22:
							position, tokenIndex, depth = position21, tokenIndex21, depth21
							{
								position54 := position
								depth++
								if !_rules[rule_]() {
									goto l53
								}
								{
									position55, tokenIndex55, depth55 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l56
									}
									position++
									goto l55
								l56:
									position, tokenIndex, depth = position55, tokenIndex55, depth55
									if buffer[position] != rune('M') {
										goto l53
									}
									position++
								}
							l55:
								{
									position57, tokenIndex57, depth57 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l58
									}
									position++
									goto l57
								l58:
									position, tokenIndex, depth = position57, tokenIndex57, depth57
									if buffer[position] != rune('E') {
										goto l53
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
										goto l53
									}
									position++
								}
							l59:
								{
									position61, tokenIndex61, depth61 := position, tokenIndex, depth
									if buffer[position] != rune('r') {
										goto l62
									}
									position++
									goto l61
								l62:
									position, tokenIndex, depth = position61, tokenIndex61, depth61
									if buffer[position] != rune('R') {
										goto l53
									}
									position++
								}
							l61:
								{
									position63, tokenIndex63, depth63 := position, tokenIndex, depth
									if buffer[position] != rune('i') {
										goto l64
									}
									position++
									goto l63
								l64:
									position, tokenIndex, depth = position63, tokenIndex63, depth63
									if buffer[position] != rune('I') {
										goto l53
									}
									position++
								}
							l63:
								{
									position65, tokenIndex65, depth65 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l66
									}
									position++
									goto l65
								l66:
									position, tokenIndex, depth = position65, tokenIndex65, depth65
									if buffer[position] != rune('C') {
										goto l53
									}
									position++
								}
							l65:
								{
									position67, tokenIndex67, depth67 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l68
									}
									position++
									goto l67
								l68:
									position, tokenIndex, depth = position67, tokenIndex67, depth67
									if buffer[position] != rune('S') {
										goto l53
									}
									position++
								}
							l67:
								if !_rules[ruleKEY]() {
									goto l53
								}
								{
									position69, tokenIndex69, depth69 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l70
									}
									{
										position71, tokenIndex71, depth71 := position, tokenIndex, depth
										if buffer[position] != rune('w') {
											goto l72
										}
										position++
										goto l71
									l72:
										position, tokenIndex, depth = position71, tokenIndex71, depth71
										if buffer[position] != rune('W') {
											goto l70
										}
										position++
									}
								l71:
									{
										position73, tokenIndex73, depth73 := position, tokenIndex, depth
										if buffer[position] != rune('h') {
											goto l74
										}
										position++
										goto l73
									l74:
										position, tokenIndex, depth = position73, tokenIndex73, depth73
										if buffer[position] != rune('H') {
											goto l70
										}
										position++
									}
								l73:
									{
										position75, tokenIndex75, depth75 := position, tokenIndex, depth
										if buffer[position] != rune('e') {
											goto l76
										}
										position++
										goto l75
									l76:
										position, tokenIndex, depth = position75, tokenIndex75, depth75
										if buffer[position] != rune('E') {
											goto l70
										}
										position++
									}
								l75:
									{
										position77, tokenIndex77, depth77 := position, tokenIndex, depth
										if buffer[position] != rune('r') {
											goto l78
										}
										position++
										goto l77
									l78:
										position, tokenIndex, depth = position77, tokenIndex77, depth77
										if buffer[position] != rune('R') {
											goto l70
										}
										position++
									}
								l77:
									{
										position79, tokenIndex79, depth79 := position, tokenIndex, depth
										if buffer[position] != rune('e') {
											goto l80
										}
										position++
										goto l79
									l80:
										position, tokenIndex, depth = position79, tokenIndex79, depth79
										if buffer[position] != rune('E') {
											goto l70
										}
										position++
									}
								l79:
									if !_rules[ruleKEY]() {
										goto l70
									}
									goto l69
								l70:
									position, tokenIndex, depth = position69, tokenIndex69, depth69
									if !(p.errorHere(position, `expected "where" to follow keyword "metrics" in "describe metrics" command`)) {
										goto l53
									}
								}
							l69:
								{
									position81, tokenIndex81, depth81 := position, tokenIndex, depth
									if !_rules[ruletagName]() {
										goto l82
									}
									goto l81
								l82:
									position, tokenIndex, depth = position81, tokenIndex81, depth81
									if !(p.errorHere(position, `expected tag key to follow keyword "where" in "describe metrics" command`)) {
										goto l53
									}
								}
							l81:
								{
									position83, tokenIndex83, depth83 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l84
									}
									if buffer[position] != rune('=') {
										goto l84
									}
									position++
									goto l83
								l84:
									position, tokenIndex, depth = position83, tokenIndex83, depth83
									if !(p.errorHere(position, `expected "=" to follow keyword "where" in "describe metrics" command`)) {
										goto l53
									}
								}
							l83:
								{
									position85, tokenIndex85, depth85 := position, tokenIndex, depth
									if !_rules[ruleliteralString]() {
										goto l86
									}
									goto l85
								l86:
									position, tokenIndex, depth = position85, tokenIndex85, depth85
									if !(p.errorHere(position, `expected string literal to follow "=" in "describe metrics" command`)) {
										goto l53
									}
								}
							l85:
								{
									add(ruleAction4, position)
								}
								depth--
								add(ruledescribeMetrics, position54)
							}
							goto l21
						l53:
							position, tokenIndex, depth = position21, tokenIndex21, depth21
							{
								position88 := position
								depth++
								{
									position89, tokenIndex89, depth89 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l90
									}
									{
										position91 := position
										depth++
										{
											position92 := position
											depth++
											if !_rules[ruleIDENTIFIER]() {
												goto l90
											}
											depth--
											add(ruleMETRIC_NAME, position92)
										}
										depth--
										add(rulePegText, position91)
									}
									{
										add(ruleAction5, position)
									}
									goto l89
								l90:
									position, tokenIndex, depth = position89, tokenIndex89, depth89
									if !(p.errorHere(position, `expected metric name to follow "describe" in "describe" command`)) {
										goto l3
									}
								}
							l89:
								if !_rules[ruleoptionalPredicateClause]() {
									goto l3
								}
								{
									add(ruleAction6, position)
								}
								depth--
								add(ruledescribeSingleStmt, position88)
							}
						}
					l21:
						depth--
						add(ruledescribeStmt, position4)
					}
					goto l2
				l3:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
					{
						position95 := position
						depth++
						if !_rules[rule_]() {
							goto l0
						}
						{
							position96, tokenIndex96, depth96 := position, tokenIndex, depth
							{
								position98, tokenIndex98, depth98 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l99
								}
								position++
								goto l98
							l99:
								position, tokenIndex, depth = position98, tokenIndex98, depth98
								if buffer[position] != rune('S') {
									goto l96
								}
								position++
							}
						l98:
							{
								position100, tokenIndex100, depth100 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l101
								}
								position++
								goto l100
							l101:
								position, tokenIndex, depth = position100, tokenIndex100, depth100
								if buffer[position] != rune('E') {
									goto l96
								}
								position++
							}
						l100:
							{
								position102, tokenIndex102, depth102 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l103
								}
								position++
								goto l102
							l103:
								position, tokenIndex, depth = position102, tokenIndex102, depth102
								if buffer[position] != rune('L') {
									goto l96
								}
								position++
							}
						l102:
							{
								position104, tokenIndex104, depth104 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l105
								}
								position++
								goto l104
							l105:
								position, tokenIndex, depth = position104, tokenIndex104, depth104
								if buffer[position] != rune('E') {
									goto l96
								}
								position++
							}
						l104:
							{
								position106, tokenIndex106, depth106 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l107
								}
								position++
								goto l106
							l107:
								position, tokenIndex, depth = position106, tokenIndex106, depth106
								if buffer[position] != rune('C') {
									goto l96
								}
								position++
							}
						l106:
							{
								position108, tokenIndex108, depth108 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l109
								}
								position++
								goto l108
							l109:
								position, tokenIndex, depth = position108, tokenIndex108, depth108
								if buffer[position] != rune('T') {
									goto l96
								}
								position++
							}
						l108:
							if !_rules[ruleKEY]() {
								goto l96
							}
							goto l97
						l96:
							position, tokenIndex, depth = position96, tokenIndex96, depth96
						}
					l97:
						{
							position110, tokenIndex110, depth110 := position, tokenIndex, depth
							if !_rules[ruleexpressionList]() {
								goto l111
							}
							goto l110
						l111:
							position, tokenIndex, depth = position110, tokenIndex110, depth110
							if !(p.errorHere(position, "expected expression to start 'select' statement")) {
								goto l0
							}
						}
					l110:
						if !(p.setContext("after expression of select statement")) {
							goto l0
						}
						if !_rules[ruleoptionalPredicateClause]() {
							goto l0
						}
						if !(p.setContext("")) {
							goto l0
						}
						{
							position112 := position
							depth++
							{
								add(ruleAction7, position)
							}
						l114:
							{
								position115, tokenIndex115, depth115 := position, tokenIndex, depth
								{
									position116, tokenIndex116, depth116 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l117
									}
									{
										position118 := position
										depth++
										{
											switch buffer[position] {
											case 'S', 's':
												{
													position120 := position
													depth++
													{
														position121, tokenIndex121, depth121 := position, tokenIndex, depth
														if buffer[position] != rune('s') {
															goto l122
														}
														position++
														goto l121
													l122:
														position, tokenIndex, depth = position121, tokenIndex121, depth121
														if buffer[position] != rune('S') {
															goto l117
														}
														position++
													}
												l121:
													{
														position123, tokenIndex123, depth123 := position, tokenIndex, depth
														if buffer[position] != rune('a') {
															goto l124
														}
														position++
														goto l123
													l124:
														position, tokenIndex, depth = position123, tokenIndex123, depth123
														if buffer[position] != rune('A') {
															goto l117
														}
														position++
													}
												l123:
													{
														position125, tokenIndex125, depth125 := position, tokenIndex, depth
														if buffer[position] != rune('m') {
															goto l126
														}
														position++
														goto l125
													l126:
														position, tokenIndex, depth = position125, tokenIndex125, depth125
														if buffer[position] != rune('M') {
															goto l117
														}
														position++
													}
												l125:
													{
														position127, tokenIndex127, depth127 := position, tokenIndex, depth
														if buffer[position] != rune('p') {
															goto l128
														}
														position++
														goto l127
													l128:
														position, tokenIndex, depth = position127, tokenIndex127, depth127
														if buffer[position] != rune('P') {
															goto l117
														}
														position++
													}
												l127:
													{
														position129, tokenIndex129, depth129 := position, tokenIndex, depth
														if buffer[position] != rune('l') {
															goto l130
														}
														position++
														goto l129
													l130:
														position, tokenIndex, depth = position129, tokenIndex129, depth129
														if buffer[position] != rune('L') {
															goto l117
														}
														position++
													}
												l129:
													{
														position131, tokenIndex131, depth131 := position, tokenIndex, depth
														if buffer[position] != rune('e') {
															goto l132
														}
														position++
														goto l131
													l132:
														position, tokenIndex, depth = position131, tokenIndex131, depth131
														if buffer[position] != rune('E') {
															goto l117
														}
														position++
													}
												l131:
													depth--
													add(rulePegText, position120)
												}
												if !_rules[ruleKEY]() {
													goto l117
												}
												{
													position133, tokenIndex133, depth133 := position, tokenIndex, depth
													if !_rules[rule_]() {
														goto l134
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
															goto l134
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
															goto l134
														}
														position++
													}
												l137:
													if !_rules[ruleKEY]() {
														goto l134
													}
													goto l133
												l134:
													position, tokenIndex, depth = position133, tokenIndex133, depth133
													if !(p.errorHere(position, `expected keyword "by" to follow keyword "sample"`)) {
														goto l117
													}
												}
											l133:
												break
											case 'R', 'r':
												{
													position139 := position
													depth++
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
															goto l117
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
															goto l117
														}
														position++
													}
												l142:
													{
														position144, tokenIndex144, depth144 := position, tokenIndex, depth
														if buffer[position] != rune('s') {
															goto l145
														}
														position++
														goto l144
													l145:
														position, tokenIndex, depth = position144, tokenIndex144, depth144
														if buffer[position] != rune('S') {
															goto l117
														}
														position++
													}
												l144:
													{
														position146, tokenIndex146, depth146 := position, tokenIndex, depth
														if buffer[position] != rune('o') {
															goto l147
														}
														position++
														goto l146
													l147:
														position, tokenIndex, depth = position146, tokenIndex146, depth146
														if buffer[position] != rune('O') {
															goto l117
														}
														position++
													}
												l146:
													{
														position148, tokenIndex148, depth148 := position, tokenIndex, depth
														if buffer[position] != rune('l') {
															goto l149
														}
														position++
														goto l148
													l149:
														position, tokenIndex, depth = position148, tokenIndex148, depth148
														if buffer[position] != rune('L') {
															goto l117
														}
														position++
													}
												l148:
													{
														position150, tokenIndex150, depth150 := position, tokenIndex, depth
														if buffer[position] != rune('u') {
															goto l151
														}
														position++
														goto l150
													l151:
														position, tokenIndex, depth = position150, tokenIndex150, depth150
														if buffer[position] != rune('U') {
															goto l117
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
															goto l117
														}
														position++
													}
												l152:
													{
														position154, tokenIndex154, depth154 := position, tokenIndex, depth
														if buffer[position] != rune('i') {
															goto l155
														}
														position++
														goto l154
													l155:
														position, tokenIndex, depth = position154, tokenIndex154, depth154
														if buffer[position] != rune('I') {
															goto l117
														}
														position++
													}
												l154:
													{
														position156, tokenIndex156, depth156 := position, tokenIndex, depth
														if buffer[position] != rune('o') {
															goto l157
														}
														position++
														goto l156
													l157:
														position, tokenIndex, depth = position156, tokenIndex156, depth156
														if buffer[position] != rune('O') {
															goto l117
														}
														position++
													}
												l156:
													{
														position158, tokenIndex158, depth158 := position, tokenIndex, depth
														if buffer[position] != rune('n') {
															goto l159
														}
														position++
														goto l158
													l159:
														position, tokenIndex, depth = position158, tokenIndex158, depth158
														if buffer[position] != rune('N') {
															goto l117
														}
														position++
													}
												l158:
													depth--
													add(rulePegText, position139)
												}
												if !_rules[ruleKEY]() {
													goto l117
												}
												break
											case 'T', 't':
												{
													position160 := position
													depth++
													{
														position161, tokenIndex161, depth161 := position, tokenIndex, depth
														if buffer[position] != rune('t') {
															goto l162
														}
														position++
														goto l161
													l162:
														position, tokenIndex, depth = position161, tokenIndex161, depth161
														if buffer[position] != rune('T') {
															goto l117
														}
														position++
													}
												l161:
													{
														position163, tokenIndex163, depth163 := position, tokenIndex, depth
														if buffer[position] != rune('o') {
															goto l164
														}
														position++
														goto l163
													l164:
														position, tokenIndex, depth = position163, tokenIndex163, depth163
														if buffer[position] != rune('O') {
															goto l117
														}
														position++
													}
												l163:
													depth--
													add(rulePegText, position160)
												}
												if !_rules[ruleKEY]() {
													goto l117
												}
												break
											default:
												{
													position165 := position
													depth++
													{
														position166, tokenIndex166, depth166 := position, tokenIndex, depth
														if buffer[position] != rune('f') {
															goto l167
														}
														position++
														goto l166
													l167:
														position, tokenIndex, depth = position166, tokenIndex166, depth166
														if buffer[position] != rune('F') {
															goto l117
														}
														position++
													}
												l166:
													{
														position168, tokenIndex168, depth168 := position, tokenIndex, depth
														if buffer[position] != rune('r') {
															goto l169
														}
														position++
														goto l168
													l169:
														position, tokenIndex, depth = position168, tokenIndex168, depth168
														if buffer[position] != rune('R') {
															goto l117
														}
														position++
													}
												l168:
													{
														position170, tokenIndex170, depth170 := position, tokenIndex, depth
														if buffer[position] != rune('o') {
															goto l171
														}
														position++
														goto l170
													l171:
														position, tokenIndex, depth = position170, tokenIndex170, depth170
														if buffer[position] != rune('O') {
															goto l117
														}
														position++
													}
												l170:
													{
														position172, tokenIndex172, depth172 := position, tokenIndex, depth
														if buffer[position] != rune('m') {
															goto l173
														}
														position++
														goto l172
													l173:
														position, tokenIndex, depth = position172, tokenIndex172, depth172
														if buffer[position] != rune('M') {
															goto l117
														}
														position++
													}
												l172:
													depth--
													add(rulePegText, position165)
												}
												if !_rules[ruleKEY]() {
													goto l117
												}
												break
											}
										}

										depth--
										add(rulePROPERTY_KEY, position118)
									}
									{
										add(ruleAction8, position)
									}
									{
										position175, tokenIndex175, depth175 := position, tokenIndex, depth
										if !_rules[rule_]() {
											goto l176
										}
										{
											position177 := position
											depth++
											{
												position178 := position
												depth++
												{
													position179, tokenIndex179, depth179 := position, tokenIndex, depth
													if !_rules[rule_]() {
														goto l180
													}
													{
														position181 := position
														depth++
														if !_rules[ruleNUMBER]() {
															goto l180
														}
													l182:
														{
															position183, tokenIndex183, depth183 := position, tokenIndex, depth
															{
																position184, tokenIndex184, depth184 := position, tokenIndex, depth
																if c := buffer[position]; c < rune('a') || c > rune('z') {
																	goto l185
																}
																position++
																goto l184
															l185:
																position, tokenIndex, depth = position184, tokenIndex184, depth184
																if c := buffer[position]; c < rune('A') || c > rune('Z') {
																	goto l183
																}
																position++
															}
														l184:
															goto l182
														l183:
															position, tokenIndex, depth = position183, tokenIndex183, depth183
														}
														depth--
														add(rulePegText, position181)
													}
													goto l179
												l180:
													position, tokenIndex, depth = position179, tokenIndex179, depth179
													if !_rules[rule_]() {
														goto l186
													}
													if !_rules[ruleSTRING]() {
														goto l186
													}
													goto l179
												l186:
													position, tokenIndex, depth = position179, tokenIndex179, depth179
													if !_rules[rule_]() {
														goto l176
													}
													{
														position187 := position
														depth++
														{
															position188, tokenIndex188, depth188 := position, tokenIndex, depth
															if buffer[position] != rune('n') {
																goto l189
															}
															position++
															goto l188
														l189:
															position, tokenIndex, depth = position188, tokenIndex188, depth188
															if buffer[position] != rune('N') {
																goto l176
															}
															position++
														}
													l188:
														{
															position190, tokenIndex190, depth190 := position, tokenIndex, depth
															if buffer[position] != rune('o') {
																goto l191
															}
															position++
															goto l190
														l191:
															position, tokenIndex, depth = position190, tokenIndex190, depth190
															if buffer[position] != rune('O') {
																goto l176
															}
															position++
														}
													l190:
														{
															position192, tokenIndex192, depth192 := position, tokenIndex, depth
															if buffer[position] != rune('w') {
																goto l193
															}
															position++
															goto l192
														l193:
															position, tokenIndex, depth = position192, tokenIndex192, depth192
															if buffer[position] != rune('W') {
																goto l176
															}
															position++
														}
													l192:
														depth--
														add(rulePegText, position187)
													}
													if !_rules[ruleKEY]() {
														goto l176
													}
												}
											l179:
												depth--
												add(ruleTIMESTAMP, position178)
											}
											depth--
											add(rulePROPERTY_VALUE, position177)
										}
										{
											add(ruleAction9, position)
										}
										goto l175
									l176:
										position, tokenIndex, depth = position175, tokenIndex175, depth175
										if !(p.errorHere(position, `expected value to follow key '%s'`, p.contents(tree, tokenIndex-2))) {
											goto l117
										}
									}
								l175:
									{
										add(ruleAction10, position)
									}
									goto l116
								l117:
									position, tokenIndex, depth = position116, tokenIndex116, depth116
									if !_rules[rule_]() {
										goto l196
									}
									{
										position197, tokenIndex197, depth197 := position, tokenIndex, depth
										if buffer[position] != rune('w') {
											goto l198
										}
										position++
										goto l197
									l198:
										position, tokenIndex, depth = position197, tokenIndex197, depth197
										if buffer[position] != rune('W') {
											goto l196
										}
										position++
									}
								l197:
									{
										position199, tokenIndex199, depth199 := position, tokenIndex, depth
										if buffer[position] != rune('h') {
											goto l200
										}
										position++
										goto l199
									l200:
										position, tokenIndex, depth = position199, tokenIndex199, depth199
										if buffer[position] != rune('H') {
											goto l196
										}
										position++
									}
								l199:
									{
										position201, tokenIndex201, depth201 := position, tokenIndex, depth
										if buffer[position] != rune('e') {
											goto l202
										}
										position++
										goto l201
									l202:
										position, tokenIndex, depth = position201, tokenIndex201, depth201
										if buffer[position] != rune('E') {
											goto l196
										}
										position++
									}
								l201:
									{
										position203, tokenIndex203, depth203 := position, tokenIndex, depth
										if buffer[position] != rune('r') {
											goto l204
										}
										position++
										goto l203
									l204:
										position, tokenIndex, depth = position203, tokenIndex203, depth203
										if buffer[position] != rune('R') {
											goto l196
										}
										position++
									}
								l203:
									{
										position205, tokenIndex205, depth205 := position, tokenIndex, depth
										if buffer[position] != rune('e') {
											goto l206
										}
										position++
										goto l205
									l206:
										position, tokenIndex, depth = position205, tokenIndex205, depth205
										if buffer[position] != rune('E') {
											goto l196
										}
										position++
									}
								l205:
									if !_rules[ruleKEY]() {
										goto l196
									}
									if !(p.errorHere(position, `encountered "where" after property clause; "where" blocks must go BEFORE 'from' and 'to' specifiers`)) {
										goto l196
									}
									goto l116
								l196:
									position, tokenIndex, depth = position116, tokenIndex116, depth116
									if !_rules[rule_]() {
										goto l115
									}
									{
										position207, tokenIndex207, depth207 := position, tokenIndex, depth
										{
											position208, tokenIndex208, depth208 := position, tokenIndex, depth
											if !matchDot() {
												goto l208
											}
											goto l207
										l208:
											position, tokenIndex, depth = position208, tokenIndex208, depth208
										}
										goto l115
									l207:
										position, tokenIndex, depth = position207, tokenIndex207, depth207
									}
									if !(p.errorHere(position, `expected key (one of 'from', 'to', 'resolution', or 'sample by') or end of input but got %q following a completed expression`, p.after(position))) {
										goto l115
									}
								}
							l116:
								goto l114
							l115:
								position, tokenIndex, depth = position115, tokenIndex115, depth115
							}
							{
								add(ruleAction11, position)
							}
							depth--
							add(rulepropertyClause, position112)
						}
						{
							add(ruleAction0, position)
						}
						depth--
						add(ruleselectStmt, position95)
					}
				}
			l2:
				if !_rules[rule_]() {
					goto l0
				}
				{
					position211, tokenIndex211, depth211 := position, tokenIndex, depth
					if !matchDot() {
						goto l211
					}
					goto l0
				l211:
					position, tokenIndex, depth = position211, tokenIndex211, depth211
				}
				depth--
				add(ruleroot, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 selectStmt <- <(_ (('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T') KEY)? (expressionList / &{ p.errorHere(position, "expected expression to start 'select' statement") }) &{ p.setContext("after expression of select statement") } optionalPredicateClause &{ p.setContext("") } propertyClause Action0)> */
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
		/* 8 propertyClause <- <(Action7 ((_ PROPERTY_KEY Action8 ((_ PROPERTY_VALUE Action9) / &{ p.errorHere(position, `expected value to follow key '%s'`, p.contents(tree, tokenIndex-2)) }) Action10) / (_ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY &{ p.errorHere(position, `encountered "where" after property clause; "where" blocks must go BEFORE 'from' and 'to' specifiers`) }) / (_ !!. &{ p.errorHere(position, `expected key (one of 'from', 'to', 'resolution', or 'sample by') or end of input but got %q following a completed expression`, p.after(position)) }))* Action11)> */
		nil,
		/* 9 optionalPredicateClause <- <(predicateClause / Action12)> */
		func() bool {
			{
				position221 := position
				depth++
				{
					position222, tokenIndex222, depth222 := position, tokenIndex, depth
					{
						position224 := position
						depth++
						if !_rules[rule_]() {
							goto l223
						}
						{
							position225, tokenIndex225, depth225 := position, tokenIndex, depth
							if buffer[position] != rune('w') {
								goto l226
							}
							position++
							goto l225
						l226:
							position, tokenIndex, depth = position225, tokenIndex225, depth225
							if buffer[position] != rune('W') {
								goto l223
							}
							position++
						}
					l225:
						{
							position227, tokenIndex227, depth227 := position, tokenIndex, depth
							if buffer[position] != rune('h') {
								goto l228
							}
							position++
							goto l227
						l228:
							position, tokenIndex, depth = position227, tokenIndex227, depth227
							if buffer[position] != rune('H') {
								goto l223
							}
							position++
						}
					l227:
						{
							position229, tokenIndex229, depth229 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l230
							}
							position++
							goto l229
						l230:
							position, tokenIndex, depth = position229, tokenIndex229, depth229
							if buffer[position] != rune('E') {
								goto l223
							}
							position++
						}
					l229:
						{
							position231, tokenIndex231, depth231 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l232
							}
							position++
							goto l231
						l232:
							position, tokenIndex, depth = position231, tokenIndex231, depth231
							if buffer[position] != rune('R') {
								goto l223
							}
							position++
						}
					l231:
						{
							position233, tokenIndex233, depth233 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l234
							}
							position++
							goto l233
						l234:
							position, tokenIndex, depth = position233, tokenIndex233, depth233
							if buffer[position] != rune('E') {
								goto l223
							}
							position++
						}
					l233:
						if !_rules[ruleKEY]() {
							goto l223
						}
						{
							position235, tokenIndex235, depth235 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l236
							}
							if !_rules[rulepredicate_1]() {
								goto l236
							}
							goto l235
						l236:
							position, tokenIndex, depth = position235, tokenIndex235, depth235
							if !(p.errorHere(position, `expected predicate to follow "where" keyword`)) {
								goto l223
							}
						}
					l235:
						depth--
						add(rulepredicateClause, position224)
					}
					goto l222
				l223:
					position, tokenIndex, depth = position222, tokenIndex222, depth222
					{
						add(ruleAction12, position)
					}
				}
			l222:
				depth--
				add(ruleoptionalPredicateClause, position221)
			}
			return true
		},
		/* 10 expressionList <- <(Action13 expression_start Action14 (_ COMMA (expression_start / &{ p.errorHere(position, `expected expression to follow ","`) }) Action15)*)> */
		func() bool {
			position238, tokenIndex238, depth238 := position, tokenIndex, depth
			{
				position239 := position
				depth++
				{
					add(ruleAction13, position)
				}
				if !_rules[ruleexpression_start]() {
					goto l238
				}
				{
					add(ruleAction14, position)
				}
			l242:
				{
					position243, tokenIndex243, depth243 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l243
					}
					if !_rules[ruleCOMMA]() {
						goto l243
					}
					{
						position244, tokenIndex244, depth244 := position, tokenIndex, depth
						if !_rules[ruleexpression_start]() {
							goto l245
						}
						goto l244
					l245:
						position, tokenIndex, depth = position244, tokenIndex244, depth244
						if !(p.errorHere(position, `expected expression to follow ","`)) {
							goto l243
						}
					}
				l244:
					{
						add(ruleAction15, position)
					}
					goto l242
				l243:
					position, tokenIndex, depth = position243, tokenIndex243, depth243
				}
				depth--
				add(ruleexpressionList, position239)
			}
			return true
		l238:
			position, tokenIndex, depth = position238, tokenIndex238, depth238
			return false
		},
		/* 11 expression_start <- <(expression_sum add_pipe)> */
		func() bool {
			position247, tokenIndex247, depth247 := position, tokenIndex, depth
			{
				position248 := position
				depth++
				{
					position249 := position
					depth++
					if !_rules[ruleexpression_product]() {
						goto l247
					}
				l250:
					{
						position251, tokenIndex251, depth251 := position, tokenIndex, depth
						if !_rules[ruleadd_pipe]() {
							goto l251
						}
						{
							position252, tokenIndex252, depth252 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l253
							}
							{
								position254 := position
								depth++
								if buffer[position] != rune('+') {
									goto l253
								}
								position++
								depth--
								add(ruleOP_ADD, position254)
							}
							{
								add(ruleAction16, position)
							}
							goto l252
						l253:
							position, tokenIndex, depth = position252, tokenIndex252, depth252
							if !_rules[rule_]() {
								goto l251
							}
							{
								position256 := position
								depth++
								if buffer[position] != rune('-') {
									goto l251
								}
								position++
								depth--
								add(ruleOP_SUB, position256)
							}
							{
								add(ruleAction17, position)
							}
						}
					l252:
						{
							position258, tokenIndex258, depth258 := position, tokenIndex, depth
							if !_rules[ruleexpression_product]() {
								goto l259
							}
							goto l258
						l259:
							position, tokenIndex, depth = position258, tokenIndex258, depth258
							if !(p.errorHere(position, `expected expression to follow operator "+" or "-"`)) {
								goto l251
							}
						}
					l258:
						{
							add(ruleAction18, position)
						}
						goto l250
					l251:
						position, tokenIndex, depth = position251, tokenIndex251, depth251
					}
					depth--
					add(ruleexpression_sum, position249)
				}
				if !_rules[ruleadd_pipe]() {
					goto l247
				}
				depth--
				add(ruleexpression_start, position248)
			}
			return true
		l247:
			position, tokenIndex, depth = position247, tokenIndex247, depth247
			return false
		},
		/* 12 expression_sum <- <(expression_product (add_pipe ((_ OP_ADD Action16) / (_ OP_SUB Action17)) (expression_product / &{ p.errorHere(position, `expected expression to follow operator "+" or "-"`) }) Action18)*)> */
		nil,
		/* 13 expression_product <- <(expression_atom (add_pipe ((_ OP_DIV Action19) / (_ OP_MULT Action20)) (expression_atom / &{ p.errorHere(position, `expected expression to follow operator "*" or "/"`) }) Action21)*)> */
		func() bool {
			position262, tokenIndex262, depth262 := position, tokenIndex, depth
			{
				position263 := position
				depth++
				if !_rules[ruleexpression_atom]() {
					goto l262
				}
			l264:
				{
					position265, tokenIndex265, depth265 := position, tokenIndex, depth
					if !_rules[ruleadd_pipe]() {
						goto l265
					}
					{
						position266, tokenIndex266, depth266 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l267
						}
						{
							position268 := position
							depth++
							if buffer[position] != rune('/') {
								goto l267
							}
							position++
							depth--
							add(ruleOP_DIV, position268)
						}
						{
							add(ruleAction19, position)
						}
						goto l266
					l267:
						position, tokenIndex, depth = position266, tokenIndex266, depth266
						if !_rules[rule_]() {
							goto l265
						}
						{
							position270 := position
							depth++
							if buffer[position] != rune('*') {
								goto l265
							}
							position++
							depth--
							add(ruleOP_MULT, position270)
						}
						{
							add(ruleAction20, position)
						}
					}
				l266:
					{
						position272, tokenIndex272, depth272 := position, tokenIndex, depth
						if !_rules[ruleexpression_atom]() {
							goto l273
						}
						goto l272
					l273:
						position, tokenIndex, depth = position272, tokenIndex272, depth272
						if !(p.errorHere(position, `expected expression to follow operator "*" or "/"`)) {
							goto l265
						}
					}
				l272:
					{
						add(ruleAction21, position)
					}
					goto l264
				l265:
					position, tokenIndex, depth = position265, tokenIndex265, depth265
				}
				depth--
				add(ruleexpression_product, position263)
			}
			return true
		l262:
			position, tokenIndex, depth = position262, tokenIndex262, depth262
			return false
		},
		/* 14 add_one_pipe <- <(_ OP_PIPE ((_ <IDENTIFIER>) / &{ p.errorHere(position, `expected function name to follow pipe "|"`) }) Action22 ((_ PAREN_OPEN (expressionList / Action23) optionalGroupBy ((_ PAREN_CLOSE) / &{ p.errorHere(position, `expected ")" to close "(" opened in pipe function call`) })) / Action24) Action25 expression_annotation)> */
		nil,
		/* 15 add_pipe <- <add_one_pipe*> */
		func() bool {
			{
				position277 := position
				depth++
			l278:
				{
					position279, tokenIndex279, depth279 := position, tokenIndex, depth
					{
						position280 := position
						depth++
						if !_rules[rule_]() {
							goto l279
						}
						{
							position281 := position
							depth++
							if buffer[position] != rune('|') {
								goto l279
							}
							position++
							depth--
							add(ruleOP_PIPE, position281)
						}
						{
							position282, tokenIndex282, depth282 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l283
							}
							{
								position284 := position
								depth++
								if !_rules[ruleIDENTIFIER]() {
									goto l283
								}
								depth--
								add(rulePegText, position284)
							}
							goto l282
						l283:
							position, tokenIndex, depth = position282, tokenIndex282, depth282
							if !(p.errorHere(position, `expected function name to follow pipe "|"`)) {
								goto l279
							}
						}
					l282:
						{
							add(ruleAction22, position)
						}
						{
							position286, tokenIndex286, depth286 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l287
							}
							if !_rules[rulePAREN_OPEN]() {
								goto l287
							}
							{
								position288, tokenIndex288, depth288 := position, tokenIndex, depth
								if !_rules[ruleexpressionList]() {
									goto l289
								}
								goto l288
							l289:
								position, tokenIndex, depth = position288, tokenIndex288, depth288
								{
									add(ruleAction23, position)
								}
							}
						l288:
							if !_rules[ruleoptionalGroupBy]() {
								goto l287
							}
							{
								position291, tokenIndex291, depth291 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l292
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l292
								}
								goto l291
							l292:
								position, tokenIndex, depth = position291, tokenIndex291, depth291
								if !(p.errorHere(position, `expected ")" to close "(" opened in pipe function call`)) {
									goto l287
								}
							}
						l291:
							goto l286
						l287:
							position, tokenIndex, depth = position286, tokenIndex286, depth286
							{
								add(ruleAction24, position)
							}
						}
					l286:
						{
							add(ruleAction25, position)
						}
						if !_rules[ruleexpression_annotation]() {
							goto l279
						}
						depth--
						add(ruleadd_one_pipe, position280)
					}
					goto l278
				l279:
					position, tokenIndex, depth = position279, tokenIndex279, depth279
				}
				depth--
				add(ruleadd_pipe, position277)
			}
			return true
		},
		/* 16 expression_atom <- <(expression_atom_raw expression_annotation)> */
		func() bool {
			position295, tokenIndex295, depth295 := position, tokenIndex, depth
			{
				position296 := position
				depth++
				{
					position297 := position
					depth++
					{
						position298, tokenIndex298, depth298 := position, tokenIndex, depth
						{
							position300 := position
							depth++
							if !_rules[rule_]() {
								goto l299
							}
							{
								position301 := position
								depth++
								if !_rules[ruleIDENTIFIER]() {
									goto l299
								}
								depth--
								add(rulePegText, position301)
							}
							{
								add(ruleAction31, position)
							}
							if !_rules[rule_]() {
								goto l299
							}
							if !_rules[rulePAREN_OPEN]() {
								goto l299
							}
							{
								position303, tokenIndex303, depth303 := position, tokenIndex, depth
								if !_rules[ruleexpressionList]() {
									goto l304
								}
								goto l303
							l304:
								position, tokenIndex, depth = position303, tokenIndex303, depth303
								if !(p.errorHere(position, `expected expression list to follow "(" in function call`)) {
									goto l299
								}
							}
						l303:
							if !_rules[ruleoptionalGroupBy]() {
								goto l299
							}
							{
								position305, tokenIndex305, depth305 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l306
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l306
								}
								goto l305
							l306:
								position, tokenIndex, depth = position305, tokenIndex305, depth305
								if !(p.errorHere(position, `expected ")" to close "(" opened by function call`)) {
									goto l299
								}
							}
						l305:
							{
								add(ruleAction32, position)
							}
							depth--
							add(ruleexpression_function, position300)
						}
						goto l298
					l299:
						position, tokenIndex, depth = position298, tokenIndex298, depth298
						{
							position309 := position
							depth++
							if !_rules[rule_]() {
								goto l308
							}
							{
								position310 := position
								depth++
								if !_rules[ruleIDENTIFIER]() {
									goto l308
								}
								depth--
								add(rulePegText, position310)
							}
							{
								add(ruleAction33, position)
							}
							{
								position312, tokenIndex312, depth312 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l313
								}
								if buffer[position] != rune('[') {
									goto l313
								}
								position++
								{
									position314, tokenIndex314, depth314 := position, tokenIndex, depth
									if !_rules[rulepredicate_1]() {
										goto l315
									}
									goto l314
								l315:
									position, tokenIndex, depth = position314, tokenIndex314, depth314
									if !(p.errorHere(position, `expected predicate to follow "[" after metric`)) {
										goto l313
									}
								}
							l314:
								{
									position316, tokenIndex316, depth316 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l317
									}
									if buffer[position] != rune(']') {
										goto l317
									}
									position++
									goto l316
								l317:
									position, tokenIndex, depth = position316, tokenIndex316, depth316
									if !(p.errorHere(position, `expected "]" to close "[" opened to apply predicate`)) {
										goto l313
									}
								}
							l316:
								goto l312
							l313:
								position, tokenIndex, depth = position312, tokenIndex312, depth312
								{
									add(ruleAction34, position)
								}
							}
						l312:
							{
								add(ruleAction35, position)
							}
							depth--
							add(ruleexpression_metric, position309)
						}
						goto l298
					l308:
						position, tokenIndex, depth = position298, tokenIndex298, depth298
						if !_rules[rule_]() {
							goto l320
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l320
						}
						{
							position321, tokenIndex321, depth321 := position, tokenIndex, depth
							if !_rules[ruleexpression_start]() {
								goto l322
							}
							goto l321
						l322:
							position, tokenIndex, depth = position321, tokenIndex321, depth321
							if !(p.errorHere(position, `expected expression to follow "("`)) {
								goto l320
							}
						}
					l321:
						{
							position323, tokenIndex323, depth323 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l324
							}
							if !_rules[rulePAREN_CLOSE]() {
								goto l324
							}
							goto l323
						l324:
							position, tokenIndex, depth = position323, tokenIndex323, depth323
							if !(p.errorHere(position, `expected ")" to close "("`)) {
								goto l320
							}
						}
					l323:
						goto l298
					l320:
						position, tokenIndex, depth = position298, tokenIndex298, depth298
						if !_rules[rule_]() {
							goto l325
						}
						{
							position326 := position
							depth++
							{
								position327 := position
								depth++
								if !_rules[ruleNUMBER]() {
									goto l325
								}
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l325
								}
								position++
							l328:
								{
									position329, tokenIndex329, depth329 := position, tokenIndex, depth
									if c := buffer[position]; c < rune('a') || c > rune('z') {
										goto l329
									}
									position++
									goto l328
								l329:
									position, tokenIndex, depth = position329, tokenIndex329, depth329
								}
								if !_rules[ruleKEY]() {
									goto l325
								}
								depth--
								add(ruleDURATION, position327)
							}
							depth--
							add(rulePegText, position326)
						}
						{
							add(ruleAction26, position)
						}
						goto l298
					l325:
						position, tokenIndex, depth = position298, tokenIndex298, depth298
						if !_rules[rule_]() {
							goto l331
						}
						{
							position332 := position
							depth++
							if !_rules[ruleNUMBER]() {
								goto l331
							}
							depth--
							add(rulePegText, position332)
						}
						{
							add(ruleAction27, position)
						}
						goto l298
					l331:
						position, tokenIndex, depth = position298, tokenIndex298, depth298
						if !_rules[rule_]() {
							goto l295
						}
						if !_rules[ruleSTRING]() {
							goto l295
						}
						{
							add(ruleAction28, position)
						}
					}
				l298:
					depth--
					add(ruleexpression_atom_raw, position297)
				}
				if !_rules[ruleexpression_annotation]() {
					goto l295
				}
				depth--
				add(ruleexpression_atom, position296)
			}
			return true
		l295:
			position, tokenIndex, depth = position295, tokenIndex295, depth295
			return false
		},
		/* 17 expression_atom_raw <- <(expression_function / expression_metric / (_ PAREN_OPEN (expression_start / &{ p.errorHere(position, `expected expression to follow "("`) }) ((_ PAREN_CLOSE) / &{ p.errorHere(position, `expected ")" to close "("`) })) / (_ <DURATION> Action26) / (_ <NUMBER> Action27) / (_ STRING Action28))> */
		nil,
		/* 18 expression_annotation_required <- <(_ '{' <(!'}' .)*> ('}' / &{ p.errorHere(position, `expected "$CLOSEBRACE$" to close "$OPENBRACE$" opened for annotation`) }) Action29)> */
		nil,
		/* 19 expression_annotation <- <expression_annotation_required?> */
		func() bool {
			{
				position338 := position
				depth++
				{
					position339, tokenIndex339, depth339 := position, tokenIndex, depth
					{
						position341 := position
						depth++
						if !_rules[rule_]() {
							goto l339
						}
						if buffer[position] != rune('{') {
							goto l339
						}
						position++
						{
							position342 := position
							depth++
						l343:
							{
								position344, tokenIndex344, depth344 := position, tokenIndex, depth
								{
									position345, tokenIndex345, depth345 := position, tokenIndex, depth
									if buffer[position] != rune('}') {
										goto l345
									}
									position++
									goto l344
								l345:
									position, tokenIndex, depth = position345, tokenIndex345, depth345
								}
								if !matchDot() {
									goto l344
								}
								goto l343
							l344:
								position, tokenIndex, depth = position344, tokenIndex344, depth344
							}
							depth--
							add(rulePegText, position342)
						}
						{
							position346, tokenIndex346, depth346 := position, tokenIndex, depth
							if buffer[position] != rune('}') {
								goto l347
							}
							position++
							goto l346
						l347:
							position, tokenIndex, depth = position346, tokenIndex346, depth346
							if !(p.errorHere(position, `expected "$CLOSEBRACE$" to close "$OPENBRACE$" opened for annotation`)) {
								goto l339
							}
						}
					l346:
						{
							add(ruleAction29, position)
						}
						depth--
						add(ruleexpression_annotation_required, position341)
					}
					goto l340
				l339:
					position, tokenIndex, depth = position339, tokenIndex339, depth339
				}
			l340:
				depth--
				add(ruleexpression_annotation, position338)
			}
			return true
		},
		/* 20 optionalGroupBy <- <(groupByClause / collapseByClause / Action30)?> */
		func() bool {
			{
				position350 := position
				depth++
				{
					position351, tokenIndex351, depth351 := position, tokenIndex, depth
					{
						position353, tokenIndex353, depth353 := position, tokenIndex, depth
						{
							position355 := position
							depth++
							if !_rules[rule_]() {
								goto l354
							}
							{
								position356, tokenIndex356, depth356 := position, tokenIndex, depth
								if buffer[position] != rune('g') {
									goto l357
								}
								position++
								goto l356
							l357:
								position, tokenIndex, depth = position356, tokenIndex356, depth356
								if buffer[position] != rune('G') {
									goto l354
								}
								position++
							}
						l356:
							{
								position358, tokenIndex358, depth358 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l359
								}
								position++
								goto l358
							l359:
								position, tokenIndex, depth = position358, tokenIndex358, depth358
								if buffer[position] != rune('R') {
									goto l354
								}
								position++
							}
						l358:
							{
								position360, tokenIndex360, depth360 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l361
								}
								position++
								goto l360
							l361:
								position, tokenIndex, depth = position360, tokenIndex360, depth360
								if buffer[position] != rune('O') {
									goto l354
								}
								position++
							}
						l360:
							{
								position362, tokenIndex362, depth362 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l363
								}
								position++
								goto l362
							l363:
								position, tokenIndex, depth = position362, tokenIndex362, depth362
								if buffer[position] != rune('U') {
									goto l354
								}
								position++
							}
						l362:
							{
								position364, tokenIndex364, depth364 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l365
								}
								position++
								goto l364
							l365:
								position, tokenIndex, depth = position364, tokenIndex364, depth364
								if buffer[position] != rune('P') {
									goto l354
								}
								position++
							}
						l364:
							if !_rules[ruleKEY]() {
								goto l354
							}
							{
								position366, tokenIndex366, depth366 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l367
								}
								{
									position368, tokenIndex368, depth368 := position, tokenIndex, depth
									if buffer[position] != rune('b') {
										goto l369
									}
									position++
									goto l368
								l369:
									position, tokenIndex, depth = position368, tokenIndex368, depth368
									if buffer[position] != rune('B') {
										goto l367
									}
									position++
								}
							l368:
								{
									position370, tokenIndex370, depth370 := position, tokenIndex, depth
									if buffer[position] != rune('y') {
										goto l371
									}
									position++
									goto l370
								l371:
									position, tokenIndex, depth = position370, tokenIndex370, depth370
									if buffer[position] != rune('Y') {
										goto l367
									}
									position++
								}
							l370:
								if !_rules[ruleKEY]() {
									goto l367
								}
								goto l366
							l367:
								position, tokenIndex, depth = position366, tokenIndex366, depth366
								if !(p.errorHere(position, `expected keyword "by" to follow keyword "group" in "group by" clause`)) {
									goto l354
								}
							}
						l366:
							{
								position372, tokenIndex372, depth372 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l373
								}
								{
									position374 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l373
									}
									depth--
									add(rulePegText, position374)
								}
								goto l372
							l373:
								position, tokenIndex, depth = position372, tokenIndex372, depth372
								if !(p.errorHere(position, `expected tag key identifier to follow "group by" keywords in "group by" clause`)) {
									goto l354
								}
							}
						l372:
							{
								add(ruleAction36, position)
							}
							{
								add(ruleAction37, position)
							}
						l377:
							{
								position378, tokenIndex378, depth378 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l378
								}
								if !_rules[ruleCOMMA]() {
									goto l378
								}
								{
									position379, tokenIndex379, depth379 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l380
									}
									{
										position381 := position
										depth++
										if !_rules[ruleCOLUMN_NAME]() {
											goto l380
										}
										depth--
										add(rulePegText, position381)
									}
									goto l379
								l380:
									position, tokenIndex, depth = position379, tokenIndex379, depth379
									if !(p.errorHere(position, `expected tag key identifier to follow "," in "group by" clause`)) {
										goto l378
									}
								}
							l379:
								{
									add(ruleAction38, position)
								}
								goto l377
							l378:
								position, tokenIndex, depth = position378, tokenIndex378, depth378
							}
							depth--
							add(rulegroupByClause, position355)
						}
						goto l353
					l354:
						position, tokenIndex, depth = position353, tokenIndex353, depth353
						{
							position384 := position
							depth++
							if !_rules[rule_]() {
								goto l383
							}
							{
								position385, tokenIndex385, depth385 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l386
								}
								position++
								goto l385
							l386:
								position, tokenIndex, depth = position385, tokenIndex385, depth385
								if buffer[position] != rune('C') {
									goto l383
								}
								position++
							}
						l385:
							{
								position387, tokenIndex387, depth387 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l388
								}
								position++
								goto l387
							l388:
								position, tokenIndex, depth = position387, tokenIndex387, depth387
								if buffer[position] != rune('O') {
									goto l383
								}
								position++
							}
						l387:
							{
								position389, tokenIndex389, depth389 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l390
								}
								position++
								goto l389
							l390:
								position, tokenIndex, depth = position389, tokenIndex389, depth389
								if buffer[position] != rune('L') {
									goto l383
								}
								position++
							}
						l389:
							{
								position391, tokenIndex391, depth391 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l392
								}
								position++
								goto l391
							l392:
								position, tokenIndex, depth = position391, tokenIndex391, depth391
								if buffer[position] != rune('L') {
									goto l383
								}
								position++
							}
						l391:
							{
								position393, tokenIndex393, depth393 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l394
								}
								position++
								goto l393
							l394:
								position, tokenIndex, depth = position393, tokenIndex393, depth393
								if buffer[position] != rune('A') {
									goto l383
								}
								position++
							}
						l393:
							{
								position395, tokenIndex395, depth395 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l396
								}
								position++
								goto l395
							l396:
								position, tokenIndex, depth = position395, tokenIndex395, depth395
								if buffer[position] != rune('P') {
									goto l383
								}
								position++
							}
						l395:
							{
								position397, tokenIndex397, depth397 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l398
								}
								position++
								goto l397
							l398:
								position, tokenIndex, depth = position397, tokenIndex397, depth397
								if buffer[position] != rune('S') {
									goto l383
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
									goto l383
								}
								position++
							}
						l399:
							if !_rules[ruleKEY]() {
								goto l383
							}
							{
								position401, tokenIndex401, depth401 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l402
								}
								{
									position403, tokenIndex403, depth403 := position, tokenIndex, depth
									if buffer[position] != rune('b') {
										goto l404
									}
									position++
									goto l403
								l404:
									position, tokenIndex, depth = position403, tokenIndex403, depth403
									if buffer[position] != rune('B') {
										goto l402
									}
									position++
								}
							l403:
								{
									position405, tokenIndex405, depth405 := position, tokenIndex, depth
									if buffer[position] != rune('y') {
										goto l406
									}
									position++
									goto l405
								l406:
									position, tokenIndex, depth = position405, tokenIndex405, depth405
									if buffer[position] != rune('Y') {
										goto l402
									}
									position++
								}
							l405:
								if !_rules[ruleKEY]() {
									goto l402
								}
								goto l401
							l402:
								position, tokenIndex, depth = position401, tokenIndex401, depth401
								if !(p.errorHere(position, `expected keyword "by" to follow keyword "collapse" in "collapse by" clause`)) {
									goto l383
								}
							}
						l401:
							{
								position407, tokenIndex407, depth407 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l408
								}
								{
									position409 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l408
									}
									depth--
									add(rulePegText, position409)
								}
								goto l407
							l408:
								position, tokenIndex, depth = position407, tokenIndex407, depth407
								if !(p.errorHere(position, `expected tag key identifier to follow "collapse by" keywords in "collapse by" clause`)) {
									goto l383
								}
							}
						l407:
							{
								add(ruleAction39, position)
							}
							{
								add(ruleAction40, position)
							}
						l412:
							{
								position413, tokenIndex413, depth413 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l413
								}
								if !_rules[ruleCOMMA]() {
									goto l413
								}
								{
									position414, tokenIndex414, depth414 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l415
									}
									{
										position416 := position
										depth++
										if !_rules[ruleCOLUMN_NAME]() {
											goto l415
										}
										depth--
										add(rulePegText, position416)
									}
									goto l414
								l415:
									position, tokenIndex, depth = position414, tokenIndex414, depth414
									if !(p.errorHere(position, `expected tag key identifier to follow "," in "collapse by" clause`)) {
										goto l413
									}
								}
							l414:
								{
									add(ruleAction41, position)
								}
								goto l412
							l413:
								position, tokenIndex, depth = position413, tokenIndex413, depth413
							}
							depth--
							add(rulecollapseByClause, position384)
						}
						goto l353
					l383:
						position, tokenIndex, depth = position353, tokenIndex353, depth353
						{
							add(ruleAction30, position)
						}
					}
				l353:
					goto l352

					position, tokenIndex, depth = position351, tokenIndex351, depth351
				}
			l352:
				depth--
				add(ruleoptionalGroupBy, position350)
			}
			return true
		},
		/* 21 expression_function <- <(_ <IDENTIFIER> Action31 _ PAREN_OPEN (expressionList / &{ p.errorHere(position, `expected expression list to follow "(" in function call`) }) optionalGroupBy ((_ PAREN_CLOSE) / &{ p.errorHere(position, `expected ")" to close "(" opened by function call`) }) Action32)> */
		nil,
		/* 22 expression_metric <- <(_ <IDENTIFIER> Action33 ((_ '[' (predicate_1 / &{ p.errorHere(position, `expected predicate to follow "[" after metric`) }) ((_ ']') / &{ p.errorHere(position, `expected "]" to close "[" opened to apply predicate`) })) / Action34) Action35)> */
		nil,
		/* 23 groupByClause <- <(_ (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P')) KEY ((_ (('b' / 'B') ('y' / 'Y')) KEY) / &{ p.errorHere(position, `expected keyword "by" to follow keyword "group" in "group by" clause`) }) ((_ <COLUMN_NAME>) / &{ p.errorHere(position, `expected tag key identifier to follow "group by" keywords in "group by" clause`) }) Action36 Action37 (_ COMMA ((_ <COLUMN_NAME>) / &{ p.errorHere(position, `expected tag key identifier to follow "," in "group by" clause`) }) Action38)*)> */
		nil,
		/* 24 collapseByClause <- <(_ (('c' / 'C') ('o' / 'O') ('l' / 'L') ('l' / 'L') ('a' / 'A') ('p' / 'P') ('s' / 'S') ('e' / 'E')) KEY ((_ (('b' / 'B') ('y' / 'Y')) KEY) / &{ p.errorHere(position, `expected keyword "by" to follow keyword "collapse" in "collapse by" clause`) }) ((_ <COLUMN_NAME>) / &{ p.errorHere(position, `expected tag key identifier to follow "collapse by" keywords in "collapse by" clause`) }) Action39 Action40 (_ COMMA ((_ <COLUMN_NAME>) / &{ p.errorHere(position, `expected tag key identifier to follow "," in "collapse by" clause`) }) Action41)*)> */
		nil,
		/* 25 predicateClause <- <(_ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY ((_ predicate_1) / &{ p.errorHere(position, `expected predicate to follow "where" keyword`) }))> */
		nil,
		/* 26 predicate_1 <- <((predicate_2 _ OP_OR (predicate_1 / &{ p.errorHere(position, `expected predicate to follow "or" operator`) }) Action42) / predicate_2)> */
		func() bool {
			position424, tokenIndex424, depth424 := position, tokenIndex, depth
			{
				position425 := position
				depth++
				{
					position426, tokenIndex426, depth426 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l427
					}
					if !_rules[rule_]() {
						goto l427
					}
					{
						position428 := position
						depth++
						{
							position429, tokenIndex429, depth429 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l430
							}
							position++
							goto l429
						l430:
							position, tokenIndex, depth = position429, tokenIndex429, depth429
							if buffer[position] != rune('O') {
								goto l427
							}
							position++
						}
					l429:
						{
							position431, tokenIndex431, depth431 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l432
							}
							position++
							goto l431
						l432:
							position, tokenIndex, depth = position431, tokenIndex431, depth431
							if buffer[position] != rune('R') {
								goto l427
							}
							position++
						}
					l431:
						if !_rules[ruleKEY]() {
							goto l427
						}
						depth--
						add(ruleOP_OR, position428)
					}
					{
						position433, tokenIndex433, depth433 := position, tokenIndex, depth
						if !_rules[rulepredicate_1]() {
							goto l434
						}
						goto l433
					l434:
						position, tokenIndex, depth = position433, tokenIndex433, depth433
						if !(p.errorHere(position, `expected predicate to follow "or" operator`)) {
							goto l427
						}
					}
				l433:
					{
						add(ruleAction42, position)
					}
					goto l426
				l427:
					position, tokenIndex, depth = position426, tokenIndex426, depth426
					if !_rules[rulepredicate_2]() {
						goto l424
					}
				}
			l426:
				depth--
				add(rulepredicate_1, position425)
			}
			return true
		l424:
			position, tokenIndex, depth = position424, tokenIndex424, depth424
			return false
		},
		/* 27 predicate_2 <- <((predicate_3 _ OP_AND (predicate_2 / &{ p.errorHere(position, `expected predicate to follow "and" operator`) }) Action43) / predicate_3)> */
		func() bool {
			position436, tokenIndex436, depth436 := position, tokenIndex, depth
			{
				position437 := position
				depth++
				{
					position438, tokenIndex438, depth438 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l439
					}
					if !_rules[rule_]() {
						goto l439
					}
					{
						position440 := position
						depth++
						{
							position441, tokenIndex441, depth441 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l442
							}
							position++
							goto l441
						l442:
							position, tokenIndex, depth = position441, tokenIndex441, depth441
							if buffer[position] != rune('A') {
								goto l439
							}
							position++
						}
					l441:
						{
							position443, tokenIndex443, depth443 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l444
							}
							position++
							goto l443
						l444:
							position, tokenIndex, depth = position443, tokenIndex443, depth443
							if buffer[position] != rune('N') {
								goto l439
							}
							position++
						}
					l443:
						{
							position445, tokenIndex445, depth445 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l446
							}
							position++
							goto l445
						l446:
							position, tokenIndex, depth = position445, tokenIndex445, depth445
							if buffer[position] != rune('D') {
								goto l439
							}
							position++
						}
					l445:
						if !_rules[ruleKEY]() {
							goto l439
						}
						depth--
						add(ruleOP_AND, position440)
					}
					{
						position447, tokenIndex447, depth447 := position, tokenIndex, depth
						if !_rules[rulepredicate_2]() {
							goto l448
						}
						goto l447
					l448:
						position, tokenIndex, depth = position447, tokenIndex447, depth447
						if !(p.errorHere(position, `expected predicate to follow "and" operator`)) {
							goto l439
						}
					}
				l447:
					{
						add(ruleAction43, position)
					}
					goto l438
				l439:
					position, tokenIndex, depth = position438, tokenIndex438, depth438
					if !_rules[rulepredicate_3]() {
						goto l436
					}
				}
			l438:
				depth--
				add(rulepredicate_2, position437)
			}
			return true
		l436:
			position, tokenIndex, depth = position436, tokenIndex436, depth436
			return false
		},
		/* 28 predicate_3 <- <((_ OP_NOT (predicate_3 / &{ p.errorHere(position, `expected predicate to follow "not" operator`) }) Action44) / (_ PAREN_OPEN (predicate_1 / &{ p.errorHere(position, `expected predicate to follow "("`) }) ((_ PAREN_CLOSE) / &{ p.errorHere(position, `expected ")" to close "(" opened in predicate`) })) / tagMatcher)> */
		func() bool {
			position450, tokenIndex450, depth450 := position, tokenIndex, depth
			{
				position451 := position
				depth++
				{
					position452, tokenIndex452, depth452 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l453
					}
					{
						position454 := position
						depth++
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
								goto l453
							}
							position++
						}
					l455:
						{
							position457, tokenIndex457, depth457 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l458
							}
							position++
							goto l457
						l458:
							position, tokenIndex, depth = position457, tokenIndex457, depth457
							if buffer[position] != rune('O') {
								goto l453
							}
							position++
						}
					l457:
						{
							position459, tokenIndex459, depth459 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l460
							}
							position++
							goto l459
						l460:
							position, tokenIndex, depth = position459, tokenIndex459, depth459
							if buffer[position] != rune('T') {
								goto l453
							}
							position++
						}
					l459:
						if !_rules[ruleKEY]() {
							goto l453
						}
						depth--
						add(ruleOP_NOT, position454)
					}
					{
						position461, tokenIndex461, depth461 := position, tokenIndex, depth
						if !_rules[rulepredicate_3]() {
							goto l462
						}
						goto l461
					l462:
						position, tokenIndex, depth = position461, tokenIndex461, depth461
						if !(p.errorHere(position, `expected predicate to follow "not" operator`)) {
							goto l453
						}
					}
				l461:
					{
						add(ruleAction44, position)
					}
					goto l452
				l453:
					position, tokenIndex, depth = position452, tokenIndex452, depth452
					if !_rules[rule_]() {
						goto l464
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l464
					}
					{
						position465, tokenIndex465, depth465 := position, tokenIndex, depth
						if !_rules[rulepredicate_1]() {
							goto l466
						}
						goto l465
					l466:
						position, tokenIndex, depth = position465, tokenIndex465, depth465
						if !(p.errorHere(position, `expected predicate to follow "("`)) {
							goto l464
						}
					}
				l465:
					{
						position467, tokenIndex467, depth467 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l468
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l468
						}
						goto l467
					l468:
						position, tokenIndex, depth = position467, tokenIndex467, depth467
						if !(p.errorHere(position, `expected ")" to close "(" opened in predicate`)) {
							goto l464
						}
					}
				l467:
					goto l452
				l464:
					position, tokenIndex, depth = position452, tokenIndex452, depth452
					{
						position469 := position
						depth++
						if !_rules[ruletagName]() {
							goto l450
						}
						{
							position470, tokenIndex470, depth470 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l471
							}
							if buffer[position] != rune('=') {
								goto l471
							}
							position++
							{
								position472, tokenIndex472, depth472 := position, tokenIndex, depth
								if !_rules[ruleliteralString]() {
									goto l473
								}
								goto l472
							l473:
								position, tokenIndex, depth = position472, tokenIndex472, depth472
								if !(p.errorHere(position, `expected string literal to follow "="`)) {
									goto l471
								}
							}
						l472:
							{
								add(ruleAction45, position)
							}
							goto l470
						l471:
							position, tokenIndex, depth = position470, tokenIndex470, depth470
							if !_rules[rule_]() {
								goto l475
							}
							if buffer[position] != rune('!') {
								goto l475
							}
							position++
							if buffer[position] != rune('=') {
								goto l475
							}
							position++
							{
								position476, tokenIndex476, depth476 := position, tokenIndex, depth
								if !_rules[ruleliteralString]() {
									goto l477
								}
								goto l476
							l477:
								position, tokenIndex, depth = position476, tokenIndex476, depth476
								if !(p.errorHere(position, `expected string literal to follow "!="`)) {
									goto l475
								}
							}
						l476:
							{
								add(ruleAction46, position)
							}
							{
								add(ruleAction47, position)
							}
							goto l470
						l475:
							position, tokenIndex, depth = position470, tokenIndex470, depth470
							if !_rules[rule_]() {
								goto l480
							}
							{
								position481, tokenIndex481, depth481 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l482
								}
								position++
								goto l481
							l482:
								position, tokenIndex, depth = position481, tokenIndex481, depth481
								if buffer[position] != rune('M') {
									goto l480
								}
								position++
							}
						l481:
							{
								position483, tokenIndex483, depth483 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l484
								}
								position++
								goto l483
							l484:
								position, tokenIndex, depth = position483, tokenIndex483, depth483
								if buffer[position] != rune('A') {
									goto l480
								}
								position++
							}
						l483:
							{
								position485, tokenIndex485, depth485 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l486
								}
								position++
								goto l485
							l486:
								position, tokenIndex, depth = position485, tokenIndex485, depth485
								if buffer[position] != rune('T') {
									goto l480
								}
								position++
							}
						l485:
							{
								position487, tokenIndex487, depth487 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l488
								}
								position++
								goto l487
							l488:
								position, tokenIndex, depth = position487, tokenIndex487, depth487
								if buffer[position] != rune('C') {
									goto l480
								}
								position++
							}
						l487:
							{
								position489, tokenIndex489, depth489 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l490
								}
								position++
								goto l489
							l490:
								position, tokenIndex, depth = position489, tokenIndex489, depth489
								if buffer[position] != rune('H') {
									goto l480
								}
								position++
							}
						l489:
							if !_rules[ruleKEY]() {
								goto l480
							}
							{
								position491, tokenIndex491, depth491 := position, tokenIndex, depth
								if !_rules[ruleliteralString]() {
									goto l492
								}
								goto l491
							l492:
								position, tokenIndex, depth = position491, tokenIndex491, depth491
								if !(p.errorHere(position, `expected regex string literal to follow "match"`)) {
									goto l480
								}
							}
						l491:
							{
								add(ruleAction48, position)
							}
							goto l470
						l480:
							position, tokenIndex, depth = position470, tokenIndex470, depth470
							if !_rules[rule_]() {
								goto l494
							}
							{
								position495, tokenIndex495, depth495 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l496
								}
								position++
								goto l495
							l496:
								position, tokenIndex, depth = position495, tokenIndex495, depth495
								if buffer[position] != rune('I') {
									goto l494
								}
								position++
							}
						l495:
							{
								position497, tokenIndex497, depth497 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l498
								}
								position++
								goto l497
							l498:
								position, tokenIndex, depth = position497, tokenIndex497, depth497
								if buffer[position] != rune('N') {
									goto l494
								}
								position++
							}
						l497:
							if !_rules[ruleKEY]() {
								goto l494
							}
							{
								position499, tokenIndex499, depth499 := position, tokenIndex, depth
								{
									position501 := position
									depth++
									{
										add(ruleAction51, position)
									}
									if !_rules[rule_]() {
										goto l500
									}
									if !_rules[rulePAREN_OPEN]() {
										goto l500
									}
									{
										position503, tokenIndex503, depth503 := position, tokenIndex, depth
										if !_rules[ruleliteralListString]() {
											goto l504
										}
										goto l503
									l504:
										position, tokenIndex, depth = position503, tokenIndex503, depth503
										if !(p.errorHere(position, `expected string literal to follow "(" in literal list`)) {
											goto l500
										}
									}
								l503:
								l505:
									{
										position506, tokenIndex506, depth506 := position, tokenIndex, depth
										if !_rules[rule_]() {
											goto l506
										}
										if !_rules[ruleCOMMA]() {
											goto l506
										}
										{
											position507, tokenIndex507, depth507 := position, tokenIndex, depth
											if !_rules[ruleliteralListString]() {
												goto l508
											}
											goto l507
										l508:
											position, tokenIndex, depth = position507, tokenIndex507, depth507
											if !(p.errorHere(position, `expected string literal to follow "," in literal list`)) {
												goto l506
											}
										}
									l507:
										goto l505
									l506:
										position, tokenIndex, depth = position506, tokenIndex506, depth506
									}
									{
										position509, tokenIndex509, depth509 := position, tokenIndex, depth
										if !_rules[rule_]() {
											goto l510
										}
										if !_rules[rulePAREN_CLOSE]() {
											goto l510
										}
										goto l509
									l510:
										position, tokenIndex, depth = position509, tokenIndex509, depth509
										if !(p.errorHere(position, `expected ")" to close "(" for literal list`)) {
											goto l500
										}
									}
								l509:
									depth--
									add(ruleliteralList, position501)
								}
								goto l499
							l500:
								position, tokenIndex, depth = position499, tokenIndex499, depth499
								if !(p.errorHere(position, `expected string literal list to follow "in" keyword`)) {
									goto l494
								}
							}
						l499:
							{
								add(ruleAction49, position)
							}
							goto l470
						l494:
							position, tokenIndex, depth = position470, tokenIndex470, depth470
							if !(p.errorHere(position, `expected "=", "!=", "match", or "in" to follow tag key in predicate`)) {
								goto l450
							}
						}
					l470:
						depth--
						add(ruletagMatcher, position469)
					}
				}
			l452:
				depth--
				add(rulepredicate_3, position451)
			}
			return true
		l450:
			position, tokenIndex, depth = position450, tokenIndex450, depth450
			return false
		},
		/* 29 tagMatcher <- <(tagName ((_ '=' (literalString / &{ p.errorHere(position, `expected string literal to follow "="`) }) Action45) / (_ ('!' '=') (literalString / &{ p.errorHere(position, `expected string literal to follow "!="`) }) Action46 Action47) / (_ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H')) KEY (literalString / &{ p.errorHere(position, `expected regex string literal to follow "match"`) }) Action48) / (_ (('i' / 'I') ('n' / 'N')) KEY (literalList / &{ p.errorHere(position, `expected string literal list to follow "in" keyword`) }) Action49) / &{ p.errorHere(position, `expected "=", "!=", "match", or "in" to follow tag key in predicate`) }))> */
		nil,
		/* 30 literalString <- <(_ STRING Action50)> */
		func() bool {
			position513, tokenIndex513, depth513 := position, tokenIndex, depth
			{
				position514 := position
				depth++
				if !_rules[rule_]() {
					goto l513
				}
				if !_rules[ruleSTRING]() {
					goto l513
				}
				{
					add(ruleAction50, position)
				}
				depth--
				add(ruleliteralString, position514)
			}
			return true
		l513:
			position, tokenIndex, depth = position513, tokenIndex513, depth513
			return false
		},
		/* 31 literalList <- <(Action51 _ PAREN_OPEN (literalListString / &{ p.errorHere(position, `expected string literal to follow "(" in literal list`) }) (_ COMMA (literalListString / &{ p.errorHere(position, `expected string literal to follow "," in literal list`) }))* ((_ PAREN_CLOSE) / &{ p.errorHere(position, `expected ")" to close "(" for literal list`) }))> */
		nil,
		/* 32 literalListString <- <(_ STRING Action52)> */
		func() bool {
			position517, tokenIndex517, depth517 := position, tokenIndex, depth
			{
				position518 := position
				depth++
				if !_rules[rule_]() {
					goto l517
				}
				if !_rules[ruleSTRING]() {
					goto l517
				}
				{
					add(ruleAction52, position)
				}
				depth--
				add(ruleliteralListString, position518)
			}
			return true
		l517:
			position, tokenIndex, depth = position517, tokenIndex517, depth517
			return false
		},
		/* 33 tagName <- <(_ <TAG_NAME> Action53)> */
		func() bool {
			position520, tokenIndex520, depth520 := position, tokenIndex, depth
			{
				position521 := position
				depth++
				if !_rules[rule_]() {
					goto l520
				}
				{
					position522 := position
					depth++
					{
						position523 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l520
						}
						depth--
						add(ruleTAG_NAME, position523)
					}
					depth--
					add(rulePegText, position522)
				}
				{
					add(ruleAction53, position)
				}
				depth--
				add(ruletagName, position521)
			}
			return true
		l520:
			position, tokenIndex, depth = position520, tokenIndex520, depth520
			return false
		},
		/* 34 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position525, tokenIndex525, depth525 := position, tokenIndex, depth
			{
				position526 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l525
				}
				depth--
				add(ruleCOLUMN_NAME, position526)
			}
			return true
		l525:
			position, tokenIndex, depth = position525, tokenIndex525, depth525
			return false
		},
		/* 35 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 36 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 37 IDENTIFIER <- <(('`' CHAR* ('`' / &{ p.errorHere(position, "expected \"`\" to end identifier") })) / (!(KEYWORD KEY) ID_SEGMENT ('.' (ID_SEGMENT / &{ p.errorHere(position, `expected identifier segment to follow "."`) }))*))> */
		func() bool {
			position529, tokenIndex529, depth529 := position, tokenIndex, depth
			{
				position530 := position
				depth++
				{
					position531, tokenIndex531, depth531 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l532
					}
					position++
				l533:
					{
						position534, tokenIndex534, depth534 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l534
						}
						goto l533
					l534:
						position, tokenIndex, depth = position534, tokenIndex534, depth534
					}
					{
						position535, tokenIndex535, depth535 := position, tokenIndex, depth
						if buffer[position] != rune('`') {
							goto l536
						}
						position++
						goto l535
					l536:
						position, tokenIndex, depth = position535, tokenIndex535, depth535
						if !(p.errorHere(position, "expected \"`\" to end identifier")) {
							goto l532
						}
					}
				l535:
					goto l531
				l532:
					position, tokenIndex, depth = position531, tokenIndex531, depth531
					{
						position537, tokenIndex537, depth537 := position, tokenIndex, depth
						{
							position538 := position
							depth++
							{
								position539, tokenIndex539, depth539 := position, tokenIndex, depth
								{
									position541, tokenIndex541, depth541 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l542
									}
									position++
									goto l541
								l542:
									position, tokenIndex, depth = position541, tokenIndex541, depth541
									if buffer[position] != rune('A') {
										goto l540
									}
									position++
								}
							l541:
								{
									position543, tokenIndex543, depth543 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l544
									}
									position++
									goto l543
								l544:
									position, tokenIndex, depth = position543, tokenIndex543, depth543
									if buffer[position] != rune('L') {
										goto l540
									}
									position++
								}
							l543:
								{
									position545, tokenIndex545, depth545 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l546
									}
									position++
									goto l545
								l546:
									position, tokenIndex, depth = position545, tokenIndex545, depth545
									if buffer[position] != rune('L') {
										goto l540
									}
									position++
								}
							l545:
								goto l539
							l540:
								position, tokenIndex, depth = position539, tokenIndex539, depth539
								{
									position548, tokenIndex548, depth548 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l549
									}
									position++
									goto l548
								l549:
									position, tokenIndex, depth = position548, tokenIndex548, depth548
									if buffer[position] != rune('A') {
										goto l547
									}
									position++
								}
							l548:
								{
									position550, tokenIndex550, depth550 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l551
									}
									position++
									goto l550
								l551:
									position, tokenIndex, depth = position550, tokenIndex550, depth550
									if buffer[position] != rune('N') {
										goto l547
									}
									position++
								}
							l550:
								{
									position552, tokenIndex552, depth552 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l553
									}
									position++
									goto l552
								l553:
									position, tokenIndex, depth = position552, tokenIndex552, depth552
									if buffer[position] != rune('D') {
										goto l547
									}
									position++
								}
							l552:
								goto l539
							l547:
								position, tokenIndex, depth = position539, tokenIndex539, depth539
								{
									position555, tokenIndex555, depth555 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l556
									}
									position++
									goto l555
								l556:
									position, tokenIndex, depth = position555, tokenIndex555, depth555
									if buffer[position] != rune('M') {
										goto l554
									}
									position++
								}
							l555:
								{
									position557, tokenIndex557, depth557 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l558
									}
									position++
									goto l557
								l558:
									position, tokenIndex, depth = position557, tokenIndex557, depth557
									if buffer[position] != rune('A') {
										goto l554
									}
									position++
								}
							l557:
								{
									position559, tokenIndex559, depth559 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l560
									}
									position++
									goto l559
								l560:
									position, tokenIndex, depth = position559, tokenIndex559, depth559
									if buffer[position] != rune('T') {
										goto l554
									}
									position++
								}
							l559:
								{
									position561, tokenIndex561, depth561 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l562
									}
									position++
									goto l561
								l562:
									position, tokenIndex, depth = position561, tokenIndex561, depth561
									if buffer[position] != rune('C') {
										goto l554
									}
									position++
								}
							l561:
								{
									position563, tokenIndex563, depth563 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l564
									}
									position++
									goto l563
								l564:
									position, tokenIndex, depth = position563, tokenIndex563, depth563
									if buffer[position] != rune('H') {
										goto l554
									}
									position++
								}
							l563:
								goto l539
							l554:
								position, tokenIndex, depth = position539, tokenIndex539, depth539
								{
									position566, tokenIndex566, depth566 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l567
									}
									position++
									goto l566
								l567:
									position, tokenIndex, depth = position566, tokenIndex566, depth566
									if buffer[position] != rune('S') {
										goto l565
									}
									position++
								}
							l566:
								{
									position568, tokenIndex568, depth568 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l569
									}
									position++
									goto l568
								l569:
									position, tokenIndex, depth = position568, tokenIndex568, depth568
									if buffer[position] != rune('E') {
										goto l565
									}
									position++
								}
							l568:
								{
									position570, tokenIndex570, depth570 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l571
									}
									position++
									goto l570
								l571:
									position, tokenIndex, depth = position570, tokenIndex570, depth570
									if buffer[position] != rune('L') {
										goto l565
									}
									position++
								}
							l570:
								{
									position572, tokenIndex572, depth572 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l573
									}
									position++
									goto l572
								l573:
									position, tokenIndex, depth = position572, tokenIndex572, depth572
									if buffer[position] != rune('E') {
										goto l565
									}
									position++
								}
							l572:
								{
									position574, tokenIndex574, depth574 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l575
									}
									position++
									goto l574
								l575:
									position, tokenIndex, depth = position574, tokenIndex574, depth574
									if buffer[position] != rune('C') {
										goto l565
									}
									position++
								}
							l574:
								{
									position576, tokenIndex576, depth576 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l577
									}
									position++
									goto l576
								l577:
									position, tokenIndex, depth = position576, tokenIndex576, depth576
									if buffer[position] != rune('T') {
										goto l565
									}
									position++
								}
							l576:
								goto l539
							l565:
								position, tokenIndex, depth = position539, tokenIndex539, depth539
								{
									switch buffer[position] {
									case 'S', 's':
										{
											position579, tokenIndex579, depth579 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l580
											}
											position++
											goto l579
										l580:
											position, tokenIndex, depth = position579, tokenIndex579, depth579
											if buffer[position] != rune('S') {
												goto l537
											}
											position++
										}
									l579:
										{
											position581, tokenIndex581, depth581 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l582
											}
											position++
											goto l581
										l582:
											position, tokenIndex, depth = position581, tokenIndex581, depth581
											if buffer[position] != rune('A') {
												goto l537
											}
											position++
										}
									l581:
										{
											position583, tokenIndex583, depth583 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l584
											}
											position++
											goto l583
										l584:
											position, tokenIndex, depth = position583, tokenIndex583, depth583
											if buffer[position] != rune('M') {
												goto l537
											}
											position++
										}
									l583:
										{
											position585, tokenIndex585, depth585 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l586
											}
											position++
											goto l585
										l586:
											position, tokenIndex, depth = position585, tokenIndex585, depth585
											if buffer[position] != rune('P') {
												goto l537
											}
											position++
										}
									l585:
										{
											position587, tokenIndex587, depth587 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l588
											}
											position++
											goto l587
										l588:
											position, tokenIndex, depth = position587, tokenIndex587, depth587
											if buffer[position] != rune('L') {
												goto l537
											}
											position++
										}
									l587:
										{
											position589, tokenIndex589, depth589 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l590
											}
											position++
											goto l589
										l590:
											position, tokenIndex, depth = position589, tokenIndex589, depth589
											if buffer[position] != rune('E') {
												goto l537
											}
											position++
										}
									l589:
										break
									case 'R', 'r':
										{
											position591, tokenIndex591, depth591 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l592
											}
											position++
											goto l591
										l592:
											position, tokenIndex, depth = position591, tokenIndex591, depth591
											if buffer[position] != rune('R') {
												goto l537
											}
											position++
										}
									l591:
										{
											position593, tokenIndex593, depth593 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l594
											}
											position++
											goto l593
										l594:
											position, tokenIndex, depth = position593, tokenIndex593, depth593
											if buffer[position] != rune('E') {
												goto l537
											}
											position++
										}
									l593:
										{
											position595, tokenIndex595, depth595 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l596
											}
											position++
											goto l595
										l596:
											position, tokenIndex, depth = position595, tokenIndex595, depth595
											if buffer[position] != rune('S') {
												goto l537
											}
											position++
										}
									l595:
										{
											position597, tokenIndex597, depth597 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l598
											}
											position++
											goto l597
										l598:
											position, tokenIndex, depth = position597, tokenIndex597, depth597
											if buffer[position] != rune('O') {
												goto l537
											}
											position++
										}
									l597:
										{
											position599, tokenIndex599, depth599 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l600
											}
											position++
											goto l599
										l600:
											position, tokenIndex, depth = position599, tokenIndex599, depth599
											if buffer[position] != rune('L') {
												goto l537
											}
											position++
										}
									l599:
										{
											position601, tokenIndex601, depth601 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l602
											}
											position++
											goto l601
										l602:
											position, tokenIndex, depth = position601, tokenIndex601, depth601
											if buffer[position] != rune('U') {
												goto l537
											}
											position++
										}
									l601:
										{
											position603, tokenIndex603, depth603 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l604
											}
											position++
											goto l603
										l604:
											position, tokenIndex, depth = position603, tokenIndex603, depth603
											if buffer[position] != rune('T') {
												goto l537
											}
											position++
										}
									l603:
										{
											position605, tokenIndex605, depth605 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l606
											}
											position++
											goto l605
										l606:
											position, tokenIndex, depth = position605, tokenIndex605, depth605
											if buffer[position] != rune('I') {
												goto l537
											}
											position++
										}
									l605:
										{
											position607, tokenIndex607, depth607 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l608
											}
											position++
											goto l607
										l608:
											position, tokenIndex, depth = position607, tokenIndex607, depth607
											if buffer[position] != rune('O') {
												goto l537
											}
											position++
										}
									l607:
										{
											position609, tokenIndex609, depth609 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l610
											}
											position++
											goto l609
										l610:
											position, tokenIndex, depth = position609, tokenIndex609, depth609
											if buffer[position] != rune('N') {
												goto l537
											}
											position++
										}
									l609:
										break
									case 'T', 't':
										{
											position611, tokenIndex611, depth611 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l612
											}
											position++
											goto l611
										l612:
											position, tokenIndex, depth = position611, tokenIndex611, depth611
											if buffer[position] != rune('T') {
												goto l537
											}
											position++
										}
									l611:
										{
											position613, tokenIndex613, depth613 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l614
											}
											position++
											goto l613
										l614:
											position, tokenIndex, depth = position613, tokenIndex613, depth613
											if buffer[position] != rune('O') {
												goto l537
											}
											position++
										}
									l613:
										break
									case 'F', 'f':
										{
											position615, tokenIndex615, depth615 := position, tokenIndex, depth
											if buffer[position] != rune('f') {
												goto l616
											}
											position++
											goto l615
										l616:
											position, tokenIndex, depth = position615, tokenIndex615, depth615
											if buffer[position] != rune('F') {
												goto l537
											}
											position++
										}
									l615:
										{
											position617, tokenIndex617, depth617 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l618
											}
											position++
											goto l617
										l618:
											position, tokenIndex, depth = position617, tokenIndex617, depth617
											if buffer[position] != rune('R') {
												goto l537
											}
											position++
										}
									l617:
										{
											position619, tokenIndex619, depth619 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l620
											}
											position++
											goto l619
										l620:
											position, tokenIndex, depth = position619, tokenIndex619, depth619
											if buffer[position] != rune('O') {
												goto l537
											}
											position++
										}
									l619:
										{
											position621, tokenIndex621, depth621 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l622
											}
											position++
											goto l621
										l622:
											position, tokenIndex, depth = position621, tokenIndex621, depth621
											if buffer[position] != rune('M') {
												goto l537
											}
											position++
										}
									l621:
										break
									case 'M', 'm':
										{
											position623, tokenIndex623, depth623 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l624
											}
											position++
											goto l623
										l624:
											position, tokenIndex, depth = position623, tokenIndex623, depth623
											if buffer[position] != rune('M') {
												goto l537
											}
											position++
										}
									l623:
										{
											position625, tokenIndex625, depth625 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l626
											}
											position++
											goto l625
										l626:
											position, tokenIndex, depth = position625, tokenIndex625, depth625
											if buffer[position] != rune('E') {
												goto l537
											}
											position++
										}
									l625:
										{
											position627, tokenIndex627, depth627 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l628
											}
											position++
											goto l627
										l628:
											position, tokenIndex, depth = position627, tokenIndex627, depth627
											if buffer[position] != rune('T') {
												goto l537
											}
											position++
										}
									l627:
										{
											position629, tokenIndex629, depth629 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l630
											}
											position++
											goto l629
										l630:
											position, tokenIndex, depth = position629, tokenIndex629, depth629
											if buffer[position] != rune('R') {
												goto l537
											}
											position++
										}
									l629:
										{
											position631, tokenIndex631, depth631 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l632
											}
											position++
											goto l631
										l632:
											position, tokenIndex, depth = position631, tokenIndex631, depth631
											if buffer[position] != rune('I') {
												goto l537
											}
											position++
										}
									l631:
										{
											position633, tokenIndex633, depth633 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l634
											}
											position++
											goto l633
										l634:
											position, tokenIndex, depth = position633, tokenIndex633, depth633
											if buffer[position] != rune('C') {
												goto l537
											}
											position++
										}
									l633:
										{
											position635, tokenIndex635, depth635 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l636
											}
											position++
											goto l635
										l636:
											position, tokenIndex, depth = position635, tokenIndex635, depth635
											if buffer[position] != rune('S') {
												goto l537
											}
											position++
										}
									l635:
										break
									case 'W', 'w':
										{
											position637, tokenIndex637, depth637 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l638
											}
											position++
											goto l637
										l638:
											position, tokenIndex, depth = position637, tokenIndex637, depth637
											if buffer[position] != rune('W') {
												goto l537
											}
											position++
										}
									l637:
										{
											position639, tokenIndex639, depth639 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l640
											}
											position++
											goto l639
										l640:
											position, tokenIndex, depth = position639, tokenIndex639, depth639
											if buffer[position] != rune('H') {
												goto l537
											}
											position++
										}
									l639:
										{
											position641, tokenIndex641, depth641 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l642
											}
											position++
											goto l641
										l642:
											position, tokenIndex, depth = position641, tokenIndex641, depth641
											if buffer[position] != rune('E') {
												goto l537
											}
											position++
										}
									l641:
										{
											position643, tokenIndex643, depth643 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l644
											}
											position++
											goto l643
										l644:
											position, tokenIndex, depth = position643, tokenIndex643, depth643
											if buffer[position] != rune('R') {
												goto l537
											}
											position++
										}
									l643:
										{
											position645, tokenIndex645, depth645 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l646
											}
											position++
											goto l645
										l646:
											position, tokenIndex, depth = position645, tokenIndex645, depth645
											if buffer[position] != rune('E') {
												goto l537
											}
											position++
										}
									l645:
										break
									case 'O', 'o':
										{
											position647, tokenIndex647, depth647 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l648
											}
											position++
											goto l647
										l648:
											position, tokenIndex, depth = position647, tokenIndex647, depth647
											if buffer[position] != rune('O') {
												goto l537
											}
											position++
										}
									l647:
										{
											position649, tokenIndex649, depth649 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l650
											}
											position++
											goto l649
										l650:
											position, tokenIndex, depth = position649, tokenIndex649, depth649
											if buffer[position] != rune('R') {
												goto l537
											}
											position++
										}
									l649:
										break
									case 'N', 'n':
										{
											position651, tokenIndex651, depth651 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l652
											}
											position++
											goto l651
										l652:
											position, tokenIndex, depth = position651, tokenIndex651, depth651
											if buffer[position] != rune('N') {
												goto l537
											}
											position++
										}
									l651:
										{
											position653, tokenIndex653, depth653 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l654
											}
											position++
											goto l653
										l654:
											position, tokenIndex, depth = position653, tokenIndex653, depth653
											if buffer[position] != rune('O') {
												goto l537
											}
											position++
										}
									l653:
										{
											position655, tokenIndex655, depth655 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l656
											}
											position++
											goto l655
										l656:
											position, tokenIndex, depth = position655, tokenIndex655, depth655
											if buffer[position] != rune('T') {
												goto l537
											}
											position++
										}
									l655:
										break
									case 'I', 'i':
										{
											position657, tokenIndex657, depth657 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l658
											}
											position++
											goto l657
										l658:
											position, tokenIndex, depth = position657, tokenIndex657, depth657
											if buffer[position] != rune('I') {
												goto l537
											}
											position++
										}
									l657:
										{
											position659, tokenIndex659, depth659 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l660
											}
											position++
											goto l659
										l660:
											position, tokenIndex, depth = position659, tokenIndex659, depth659
											if buffer[position] != rune('N') {
												goto l537
											}
											position++
										}
									l659:
										break
									case 'C', 'c':
										{
											position661, tokenIndex661, depth661 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l662
											}
											position++
											goto l661
										l662:
											position, tokenIndex, depth = position661, tokenIndex661, depth661
											if buffer[position] != rune('C') {
												goto l537
											}
											position++
										}
									l661:
										{
											position663, tokenIndex663, depth663 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l664
											}
											position++
											goto l663
										l664:
											position, tokenIndex, depth = position663, tokenIndex663, depth663
											if buffer[position] != rune('O') {
												goto l537
											}
											position++
										}
									l663:
										{
											position665, tokenIndex665, depth665 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l666
											}
											position++
											goto l665
										l666:
											position, tokenIndex, depth = position665, tokenIndex665, depth665
											if buffer[position] != rune('L') {
												goto l537
											}
											position++
										}
									l665:
										{
											position667, tokenIndex667, depth667 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l668
											}
											position++
											goto l667
										l668:
											position, tokenIndex, depth = position667, tokenIndex667, depth667
											if buffer[position] != rune('L') {
												goto l537
											}
											position++
										}
									l667:
										{
											position669, tokenIndex669, depth669 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l670
											}
											position++
											goto l669
										l670:
											position, tokenIndex, depth = position669, tokenIndex669, depth669
											if buffer[position] != rune('A') {
												goto l537
											}
											position++
										}
									l669:
										{
											position671, tokenIndex671, depth671 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l672
											}
											position++
											goto l671
										l672:
											position, tokenIndex, depth = position671, tokenIndex671, depth671
											if buffer[position] != rune('P') {
												goto l537
											}
											position++
										}
									l671:
										{
											position673, tokenIndex673, depth673 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l674
											}
											position++
											goto l673
										l674:
											position, tokenIndex, depth = position673, tokenIndex673, depth673
											if buffer[position] != rune('S') {
												goto l537
											}
											position++
										}
									l673:
										{
											position675, tokenIndex675, depth675 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l676
											}
											position++
											goto l675
										l676:
											position, tokenIndex, depth = position675, tokenIndex675, depth675
											if buffer[position] != rune('E') {
												goto l537
											}
											position++
										}
									l675:
										break
									case 'G', 'g':
										{
											position677, tokenIndex677, depth677 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l678
											}
											position++
											goto l677
										l678:
											position, tokenIndex, depth = position677, tokenIndex677, depth677
											if buffer[position] != rune('G') {
												goto l537
											}
											position++
										}
									l677:
										{
											position679, tokenIndex679, depth679 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l680
											}
											position++
											goto l679
										l680:
											position, tokenIndex, depth = position679, tokenIndex679, depth679
											if buffer[position] != rune('R') {
												goto l537
											}
											position++
										}
									l679:
										{
											position681, tokenIndex681, depth681 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l682
											}
											position++
											goto l681
										l682:
											position, tokenIndex, depth = position681, tokenIndex681, depth681
											if buffer[position] != rune('O') {
												goto l537
											}
											position++
										}
									l681:
										{
											position683, tokenIndex683, depth683 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l684
											}
											position++
											goto l683
										l684:
											position, tokenIndex, depth = position683, tokenIndex683, depth683
											if buffer[position] != rune('U') {
												goto l537
											}
											position++
										}
									l683:
										{
											position685, tokenIndex685, depth685 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l686
											}
											position++
											goto l685
										l686:
											position, tokenIndex, depth = position685, tokenIndex685, depth685
											if buffer[position] != rune('P') {
												goto l537
											}
											position++
										}
									l685:
										break
									case 'D', 'd':
										{
											position687, tokenIndex687, depth687 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l688
											}
											position++
											goto l687
										l688:
											position, tokenIndex, depth = position687, tokenIndex687, depth687
											if buffer[position] != rune('D') {
												goto l537
											}
											position++
										}
									l687:
										{
											position689, tokenIndex689, depth689 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l690
											}
											position++
											goto l689
										l690:
											position, tokenIndex, depth = position689, tokenIndex689, depth689
											if buffer[position] != rune('E') {
												goto l537
											}
											position++
										}
									l689:
										{
											position691, tokenIndex691, depth691 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l692
											}
											position++
											goto l691
										l692:
											position, tokenIndex, depth = position691, tokenIndex691, depth691
											if buffer[position] != rune('S') {
												goto l537
											}
											position++
										}
									l691:
										{
											position693, tokenIndex693, depth693 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l694
											}
											position++
											goto l693
										l694:
											position, tokenIndex, depth = position693, tokenIndex693, depth693
											if buffer[position] != rune('C') {
												goto l537
											}
											position++
										}
									l693:
										{
											position695, tokenIndex695, depth695 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l696
											}
											position++
											goto l695
										l696:
											position, tokenIndex, depth = position695, tokenIndex695, depth695
											if buffer[position] != rune('R') {
												goto l537
											}
											position++
										}
									l695:
										{
											position697, tokenIndex697, depth697 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l698
											}
											position++
											goto l697
										l698:
											position, tokenIndex, depth = position697, tokenIndex697, depth697
											if buffer[position] != rune('I') {
												goto l537
											}
											position++
										}
									l697:
										{
											position699, tokenIndex699, depth699 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l700
											}
											position++
											goto l699
										l700:
											position, tokenIndex, depth = position699, tokenIndex699, depth699
											if buffer[position] != rune('B') {
												goto l537
											}
											position++
										}
									l699:
										{
											position701, tokenIndex701, depth701 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l702
											}
											position++
											goto l701
										l702:
											position, tokenIndex, depth = position701, tokenIndex701, depth701
											if buffer[position] != rune('E') {
												goto l537
											}
											position++
										}
									l701:
										break
									case 'B', 'b':
										{
											position703, tokenIndex703, depth703 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l704
											}
											position++
											goto l703
										l704:
											position, tokenIndex, depth = position703, tokenIndex703, depth703
											if buffer[position] != rune('B') {
												goto l537
											}
											position++
										}
									l703:
										{
											position705, tokenIndex705, depth705 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l706
											}
											position++
											goto l705
										l706:
											position, tokenIndex, depth = position705, tokenIndex705, depth705
											if buffer[position] != rune('Y') {
												goto l537
											}
											position++
										}
									l705:
										break
									default:
										{
											position707, tokenIndex707, depth707 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l708
											}
											position++
											goto l707
										l708:
											position, tokenIndex, depth = position707, tokenIndex707, depth707
											if buffer[position] != rune('A') {
												goto l537
											}
											position++
										}
									l707:
										{
											position709, tokenIndex709, depth709 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l710
											}
											position++
											goto l709
										l710:
											position, tokenIndex, depth = position709, tokenIndex709, depth709
											if buffer[position] != rune('S') {
												goto l537
											}
											position++
										}
									l709:
										break
									}
								}

							}
						l539:
							depth--
							add(ruleKEYWORD, position538)
						}
						if !_rules[ruleKEY]() {
							goto l537
						}
						goto l529
					l537:
						position, tokenIndex, depth = position537, tokenIndex537, depth537
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l529
					}
				l711:
					{
						position712, tokenIndex712, depth712 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l712
						}
						position++
						{
							position713, tokenIndex713, depth713 := position, tokenIndex, depth
							if !_rules[ruleID_SEGMENT]() {
								goto l714
							}
							goto l713
						l714:
							position, tokenIndex, depth = position713, tokenIndex713, depth713
							if !(p.errorHere(position, `expected identifier segment to follow "."`)) {
								goto l712
							}
						}
					l713:
						goto l711
					l712:
						position, tokenIndex, depth = position712, tokenIndex712, depth712
					}
				}
			l531:
				depth--
				add(ruleIDENTIFIER, position530)
			}
			return true
		l529:
			position, tokenIndex, depth = position529, tokenIndex529, depth529
			return false
		},
		/* 38 TIMESTAMP <- <((_ <(NUMBER ([a-z] / [A-Z])*)>) / (_ STRING) / (_ <(('n' / 'N') ('o' / 'O') ('w' / 'W'))> KEY))> */
		nil,
		/* 39 ID_SEGMENT <- <(ID_START ID_CONT*)> */
		func() bool {
			position716, tokenIndex716, depth716 := position, tokenIndex, depth
			{
				position717 := position
				depth++
				if !_rules[ruleID_START]() {
					goto l716
				}
			l718:
				{
					position719, tokenIndex719, depth719 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l719
					}
					goto l718
				l719:
					position, tokenIndex, depth = position719, tokenIndex719, depth719
				}
				depth--
				add(ruleID_SEGMENT, position717)
			}
			return true
		l716:
			position, tokenIndex, depth = position716, tokenIndex716, depth716
			return false
		},
		/* 40 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position720, tokenIndex720, depth720 := position, tokenIndex, depth
			{
				position721 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l720
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l720
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l720
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position721)
			}
			return true
		l720:
			position, tokenIndex, depth = position720, tokenIndex720, depth720
			return false
		},
		/* 41 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position723, tokenIndex723, depth723 := position, tokenIndex, depth
			{
				position724 := position
				depth++
				{
					position725, tokenIndex725, depth725 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l726
					}
					goto l725
				l726:
					position, tokenIndex, depth = position725, tokenIndex725, depth725
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l723
					}
					position++
				}
			l725:
				depth--
				add(ruleID_CONT, position724)
			}
			return true
		l723:
			position, tokenIndex, depth = position723, tokenIndex723, depth723
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
			position738, tokenIndex738, depth738 := position, tokenIndex, depth
			{
				position739 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l738
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position739)
			}
			return true
		l738:
			position, tokenIndex, depth = position738, tokenIndex738, depth738
			return false
		},
		/* 54 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position740, tokenIndex740, depth740 := position, tokenIndex, depth
			{
				position741 := position
				depth++
				if buffer[position] != rune('"') {
					goto l740
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position741)
			}
			return true
		l740:
			position, tokenIndex, depth = position740, tokenIndex740, depth740
			return false
		},
		/* 55 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> (QUOTE_SINGLE / &{ p.errorHere(position, `expected "'" to close string`) })) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> (QUOTE_DOUBLE / &{ p.errorHere(position, `expected '"' to close string`) })))> */
		func() bool {
			position742, tokenIndex742, depth742 := position, tokenIndex, depth
			{
				position743 := position
				depth++
				{
					position744, tokenIndex744, depth744 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l745
					}
					{
						position746 := position
						depth++
					l747:
						{
							position748, tokenIndex748, depth748 := position, tokenIndex, depth
							{
								position749, tokenIndex749, depth749 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l749
								}
								goto l748
							l749:
								position, tokenIndex, depth = position749, tokenIndex749, depth749
							}
							if !_rules[ruleCHAR]() {
								goto l748
							}
							goto l747
						l748:
							position, tokenIndex, depth = position748, tokenIndex748, depth748
						}
						depth--
						add(rulePegText, position746)
					}
					{
						position750, tokenIndex750, depth750 := position, tokenIndex, depth
						if !_rules[ruleQUOTE_SINGLE]() {
							goto l751
						}
						goto l750
					l751:
						position, tokenIndex, depth = position750, tokenIndex750, depth750
						if !(p.errorHere(position, `expected "'" to close string`)) {
							goto l745
						}
					}
				l750:
					goto l744
				l745:
					position, tokenIndex, depth = position744, tokenIndex744, depth744
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l742
					}
					{
						position752 := position
						depth++
					l753:
						{
							position754, tokenIndex754, depth754 := position, tokenIndex, depth
							{
								position755, tokenIndex755, depth755 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l755
								}
								goto l754
							l755:
								position, tokenIndex, depth = position755, tokenIndex755, depth755
							}
							if !_rules[ruleCHAR]() {
								goto l754
							}
							goto l753
						l754:
							position, tokenIndex, depth = position754, tokenIndex754, depth754
						}
						depth--
						add(rulePegText, position752)
					}
					{
						position756, tokenIndex756, depth756 := position, tokenIndex, depth
						if !_rules[ruleQUOTE_DOUBLE]() {
							goto l757
						}
						goto l756
					l757:
						position, tokenIndex, depth = position756, tokenIndex756, depth756
						if !(p.errorHere(position, `expected '"' to close string`)) {
							goto l742
						}
					}
				l756:
				}
			l744:
				depth--
				add(ruleSTRING, position743)
			}
			return true
		l742:
			position, tokenIndex, depth = position742, tokenIndex742, depth742
			return false
		},
		/* 56 CHAR <- <(('\\' ((&('"') (QUOTE_DOUBLE / &{ p.errorHere(position, "expected \"\\\", \"'\", \"`\", or '\"' to follow \"\\\" in string literal") })) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position758, tokenIndex758, depth758 := position, tokenIndex, depth
			{
				position759 := position
				depth++
				{
					position760, tokenIndex760, depth760 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l761
					}
					position++
					{
						switch buffer[position] {
						case '"':
							{
								position763, tokenIndex763, depth763 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l764
								}
								goto l763
							l764:
								position, tokenIndex, depth = position763, tokenIndex763, depth763
								if !(p.errorHere(position, "expected \"\\\", \"'\", \"`\", or '\"' to follow \"\\\" in string literal")) {
									goto l761
								}
							}
						l763:
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l761
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l761
							}
							break
						}
					}

					goto l760
				l761:
					position, tokenIndex, depth = position760, tokenIndex760, depth760
					{
						position765, tokenIndex765, depth765 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l765
						}
						goto l758
					l765:
						position, tokenIndex, depth = position765, tokenIndex765, depth765
					}
					if !matchDot() {
						goto l758
					}
				}
			l760:
				depth--
				add(ruleCHAR, position759)
			}
			return true
		l758:
			position, tokenIndex, depth = position758, tokenIndex758, depth758
			return false
		},
		/* 57 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position766, tokenIndex766, depth766 := position, tokenIndex, depth
			{
				position767 := position
				depth++
				{
					position768, tokenIndex768, depth768 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l769
					}
					position++
					goto l768
				l769:
					position, tokenIndex, depth = position768, tokenIndex768, depth768
					if buffer[position] != rune('\\') {
						goto l766
					}
					position++
				}
			l768:
				depth--
				add(ruleESCAPE_CLASS, position767)
			}
			return true
		l766:
			position, tokenIndex, depth = position766, tokenIndex766, depth766
			return false
		},
		/* 58 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position770, tokenIndex770, depth770 := position, tokenIndex, depth
			{
				position771 := position
				depth++
				{
					position772 := position
					depth++
					{
						position773, tokenIndex773, depth773 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l773
						}
						position++
						goto l774
					l773:
						position, tokenIndex, depth = position773, tokenIndex773, depth773
					}
				l774:
					{
						position775 := position
						depth++
						{
							position776, tokenIndex776, depth776 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l777
							}
							position++
							goto l776
						l777:
							position, tokenIndex, depth = position776, tokenIndex776, depth776
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l770
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
						}
					l776:
						depth--
						add(ruleNUMBER_NATURAL, position775)
					}
					depth--
					add(ruleNUMBER_INTEGER, position772)
				}
				{
					position780, tokenIndex780, depth780 := position, tokenIndex, depth
					{
						position782 := position
						depth++
						if buffer[position] != rune('.') {
							goto l780
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l780
						}
						position++
					l783:
						{
							position784, tokenIndex784, depth784 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l784
							}
							position++
							goto l783
						l784:
							position, tokenIndex, depth = position784, tokenIndex784, depth784
						}
						depth--
						add(ruleNUMBER_FRACTION, position782)
					}
					goto l781
				l780:
					position, tokenIndex, depth = position780, tokenIndex780, depth780
				}
			l781:
				{
					position785, tokenIndex785, depth785 := position, tokenIndex, depth
					{
						position787 := position
						depth++
						{
							position788, tokenIndex788, depth788 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l789
							}
							position++
							goto l788
						l789:
							position, tokenIndex, depth = position788, tokenIndex788, depth788
							if buffer[position] != rune('E') {
								goto l785
							}
							position++
						}
					l788:
						{
							position790, tokenIndex790, depth790 := position, tokenIndex, depth
							{
								position792, tokenIndex792, depth792 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l793
								}
								position++
								goto l792
							l793:
								position, tokenIndex, depth = position792, tokenIndex792, depth792
								if buffer[position] != rune('-') {
									goto l790
								}
								position++
							}
						l792:
							goto l791
						l790:
							position, tokenIndex, depth = position790, tokenIndex790, depth790
						}
					l791:
						{
							position794, tokenIndex794, depth794 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l795
							}
							position++
						l796:
							{
								position797, tokenIndex797, depth797 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l797
								}
								position++
								goto l796
							l797:
								position, tokenIndex, depth = position797, tokenIndex797, depth797
							}
							goto l794
						l795:
							position, tokenIndex, depth = position794, tokenIndex794, depth794
							if !(p.errorHere(position, `expected exponent`)) {
								goto l785
							}
						}
					l794:
						depth--
						add(ruleNUMBER_EXP, position787)
					}
					goto l786
				l785:
					position, tokenIndex, depth = position785, tokenIndex785, depth785
				}
			l786:
				depth--
				add(ruleNUMBER, position771)
			}
			return true
		l770:
			position, tokenIndex, depth = position770, tokenIndex770, depth770
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
			position803, tokenIndex803, depth803 := position, tokenIndex, depth
			{
				position804 := position
				depth++
				if buffer[position] != rune('(') {
					goto l803
				}
				position++
				depth--
				add(rulePAREN_OPEN, position804)
			}
			return true
		l803:
			position, tokenIndex, depth = position803, tokenIndex803, depth803
			return false
		},
		/* 65 PAREN_CLOSE <- <')'> */
		func() bool {
			position805, tokenIndex805, depth805 := position, tokenIndex, depth
			{
				position806 := position
				depth++
				if buffer[position] != rune(')') {
					goto l805
				}
				position++
				depth--
				add(rulePAREN_CLOSE, position806)
			}
			return true
		l805:
			position, tokenIndex, depth = position805, tokenIndex805, depth805
			return false
		},
		/* 66 COMMA <- <','> */
		func() bool {
			position807, tokenIndex807, depth807 := position, tokenIndex, depth
			{
				position808 := position
				depth++
				if buffer[position] != rune(',') {
					goto l807
				}
				position++
				depth--
				add(ruleCOMMA, position808)
			}
			return true
		l807:
			position, tokenIndex, depth = position807, tokenIndex807, depth807
			return false
		},
		/* 67 _ <- <((&('/') COMMENT_BLOCK) | (&('-') COMMENT_TRAIL) | (&('\t' | '\n' | ' ') SPACE))*> */
		func() bool {
			{
				position810 := position
				depth++
			l811:
				{
					position812, tokenIndex812, depth812 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '/':
							{
								position814 := position
								depth++
								if buffer[position] != rune('/') {
									goto l812
								}
								position++
								if buffer[position] != rune('*') {
									goto l812
								}
								position++
							l815:
								{
									position816, tokenIndex816, depth816 := position, tokenIndex, depth
									{
										position817, tokenIndex817, depth817 := position, tokenIndex, depth
										if buffer[position] != rune('*') {
											goto l817
										}
										position++
										if buffer[position] != rune('/') {
											goto l817
										}
										position++
										goto l816
									l817:
										position, tokenIndex, depth = position817, tokenIndex817, depth817
									}
									if !matchDot() {
										goto l816
									}
									goto l815
								l816:
									position, tokenIndex, depth = position816, tokenIndex816, depth816
								}
								if buffer[position] != rune('*') {
									goto l812
								}
								position++
								if buffer[position] != rune('/') {
									goto l812
								}
								position++
								depth--
								add(ruleCOMMENT_BLOCK, position814)
							}
							break
						case '-':
							{
								position818 := position
								depth++
								if buffer[position] != rune('-') {
									goto l812
								}
								position++
								if buffer[position] != rune('-') {
									goto l812
								}
								position++
							l819:
								{
									position820, tokenIndex820, depth820 := position, tokenIndex, depth
									{
										position821, tokenIndex821, depth821 := position, tokenIndex, depth
										if buffer[position] != rune('\n') {
											goto l821
										}
										position++
										goto l820
									l821:
										position, tokenIndex, depth = position821, tokenIndex821, depth821
									}
									if !matchDot() {
										goto l820
									}
									goto l819
								l820:
									position, tokenIndex, depth = position820, tokenIndex820, depth820
								}
								depth--
								add(ruleCOMMENT_TRAIL, position818)
							}
							break
						default:
							{
								position822 := position
								depth++
								{
									switch buffer[position] {
									case '\t':
										if buffer[position] != rune('\t') {
											goto l812
										}
										position++
										break
									case '\n':
										if buffer[position] != rune('\n') {
											goto l812
										}
										position++
										break
									default:
										if buffer[position] != rune(' ') {
											goto l812
										}
										position++
										break
									}
								}

								depth--
								add(ruleSPACE, position822)
							}
							break
						}
					}

					goto l811
				l812:
					position, tokenIndex, depth = position812, tokenIndex812, depth812
				}
				depth--
				add(rule_, position810)
			}
			return true
		},
		/* 68 COMMENT_TRAIL <- <('-' '-' (!'\n' .)*)> */
		nil,
		/* 69 COMMENT_BLOCK <- <('/' '*' (!('*' '/') .)* ('*' '/'))> */
		nil,
		/* 70 KEY <- <!ID_CONT> */
		func() bool {
			position826, tokenIndex826, depth826 := position, tokenIndex, depth
			{
				position827 := position
				depth++
				{
					position828, tokenIndex828, depth828 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l828
					}
					goto l826
				l828:
					position, tokenIndex, depth = position828, tokenIndex828, depth828
				}
				depth--
				add(ruleKEY, position827)
			}
			return true
		l826:
			position, tokenIndex, depth = position826, tokenIndex826, depth826
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
		/* 110 Action36 <- <{ p.addGroupBy() }> */
		nil,
		/* 111 Action37 <- <{ p.appendGroupTag(unescapeLiteral(text)) }> */
		nil,
		/* 112 Action38 <- <{ p.appendGroupTag(unescapeLiteral(text)) }> */
		nil,
		/* 113 Action39 <- <{ p.addCollapseBy() }> */
		nil,
		/* 114 Action40 <- <{ p.appendGroupTag(unescapeLiteral(text)) }> */
		nil,
		/* 115 Action41 <- <{ p.appendGroupTag(unescapeLiteral(text)) }> */
		nil,
		/* 116 Action42 <- <{ p.addOrPredicate() }> */
		nil,
		/* 117 Action43 <- <{ p.addAndPredicate() }> */
		nil,
		/* 118 Action44 <- <{ p.addNotPredicate() }> */
		nil,
		/* 119 Action45 <- <{ p.addLiteralMatcher() }> */
		nil,
		/* 120 Action46 <- <{ p.addLiteralMatcher() }> */
		nil,
		/* 121 Action47 <- <{ p.addNotPredicate() }> */
		nil,
		/* 122 Action48 <- <{ p.addRegexMatcher() }> */
		nil,
		/* 123 Action49 <- <{ p.addListMatcher() }> */
		nil,
		/* 124 Action50 <- <{ p.pushString(unescapeLiteral(text)) }> */
		nil,
		/* 125 Action51 <- <{ p.addLiteralList() }> */
		nil,
		/* 126 Action52 <- <{ p.appendLiteral(unescapeLiteral(text)) }> */
		nil,
		/* 127 Action53 <- <{ p.addTagLiteral(unescapeLiteral(text)) }> */
		nil,
	}
	p.rules = _rules
}
