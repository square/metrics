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

func (t *tokens32) Expand(index int) {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
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
	tokens32
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
	p.tokens32.PrintSyntaxTree(p.Buffer)
}

func (p *Parser) Highlighter() {
	p.PrintSyntax()
}

func (p *Parser) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for token := range p.Tokens() {
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

	tree := tokens32{tree: make([]token32, math.MaxInt16)}
	var max token32
	position, depth, tokenIndex, buffer, _rules := uint32(0), uint32(0), 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule pegRule, begin uint32) {
		tree.Expand(tokenIndex)
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
		/* 8 propertyClause <- <(Action7 ((_ PROPERTY_KEY Action8 ((_ PROPERTY_VALUE Action9) / &{ p.errorHere(position, `expected value to follow key '%s'`, p.contents(tree, tokenIndex-2)) }) Action10) / (_ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY &{ p.errorHere(position, `encountered "where" after property clause; "where" blocks must go BEFORE 'from' and 'to' specifiers`) }) / (_ !!. &{ p.errorHere(position, `expected key (one of 'from', 'to', 'resolution', or 'sample by') or end of input but got %q following a completed expression`, p.after(position)) }))* Action11)> */
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
		/* 18 expression_annotation_required <- <(_ '{' <(!'}' .)*> ('}' / &{ p.errorHere(position, `expected "$CLOSEBRACE$" to close "$OPENBRACE$" opened for annotation`) }) Action29)> */
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
							if !(p.errorHere(position, `expected "$CLOSEBRACE$" to close "$OPENBRACE$" opened for annotation`)) {
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
		/* 20 optionalGroupBy <- <(groupByClause / collapseByClause / Action30)?> */
		func() bool {
			{
				position348 := position
				depth++
				{
					position349, tokenIndex349, depth349 := position, tokenIndex, depth
					{
						position351, tokenIndex351, depth351 := position, tokenIndex, depth
						{
							position353 := position
							depth++
							if !_rules[rule_]() {
								goto l352
							}
							{
								position354, tokenIndex354, depth354 := position, tokenIndex, depth
								if buffer[position] != rune('g') {
									goto l355
								}
								position++
								goto l354
							l355:
								position, tokenIndex, depth = position354, tokenIndex354, depth354
								if buffer[position] != rune('G') {
									goto l352
								}
								position++
							}
						l354:
							{
								position356, tokenIndex356, depth356 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l357
								}
								position++
								goto l356
							l357:
								position, tokenIndex, depth = position356, tokenIndex356, depth356
								if buffer[position] != rune('R') {
									goto l352
								}
								position++
							}
						l356:
							{
								position358, tokenIndex358, depth358 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l359
								}
								position++
								goto l358
							l359:
								position, tokenIndex, depth = position358, tokenIndex358, depth358
								if buffer[position] != rune('O') {
									goto l352
								}
								position++
							}
						l358:
							{
								position360, tokenIndex360, depth360 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l361
								}
								position++
								goto l360
							l361:
								position, tokenIndex, depth = position360, tokenIndex360, depth360
								if buffer[position] != rune('U') {
									goto l352
								}
								position++
							}
						l360:
							{
								position362, tokenIndex362, depth362 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l363
								}
								position++
								goto l362
							l363:
								position, tokenIndex, depth = position362, tokenIndex362, depth362
								if buffer[position] != rune('P') {
									goto l352
								}
								position++
							}
						l362:
							if !_rules[ruleKEY]() {
								goto l352
							}
							{
								position364, tokenIndex364, depth364 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l365
								}
								{
									position366, tokenIndex366, depth366 := position, tokenIndex, depth
									if buffer[position] != rune('b') {
										goto l367
									}
									position++
									goto l366
								l367:
									position, tokenIndex, depth = position366, tokenIndex366, depth366
									if buffer[position] != rune('B') {
										goto l365
									}
									position++
								}
							l366:
								{
									position368, tokenIndex368, depth368 := position, tokenIndex, depth
									if buffer[position] != rune('y') {
										goto l369
									}
									position++
									goto l368
								l369:
									position, tokenIndex, depth = position368, tokenIndex368, depth368
									if buffer[position] != rune('Y') {
										goto l365
									}
									position++
								}
							l368:
								if !_rules[ruleKEY]() {
									goto l365
								}
								goto l364
							l365:
								position, tokenIndex, depth = position364, tokenIndex364, depth364
								if !(p.errorHere(position, `expected keyword "by" to follow keyword "group" in "group by" clause`)) {
									goto l352
								}
							}
						l364:
							{
								position370, tokenIndex370, depth370 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l371
								}
								{
									position372 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l371
									}
									depth--
									add(rulePegText, position372)
								}
								goto l370
							l371:
								position, tokenIndex, depth = position370, tokenIndex370, depth370
								if !(p.errorHere(position, `expected tag key identifier to follow "group by" keywords in "group by" clause`)) {
									goto l352
								}
							}
						l370:
							{
								add(ruleAction36, position)
							}
							{
								add(ruleAction37, position)
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
									add(ruleAction38, position)
								}
								goto l375
							l376:
								position, tokenIndex, depth = position376, tokenIndex376, depth376
							}
							depth--
							add(rulegroupByClause, position353)
						}
						goto l351
					l352:
						position, tokenIndex, depth = position351, tokenIndex351, depth351
						{
							position382 := position
							depth++
							if !_rules[rule_]() {
								goto l381
							}
							{
								position383, tokenIndex383, depth383 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l384
								}
								position++
								goto l383
							l384:
								position, tokenIndex, depth = position383, tokenIndex383, depth383
								if buffer[position] != rune('C') {
									goto l381
								}
								position++
							}
						l383:
							{
								position385, tokenIndex385, depth385 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l386
								}
								position++
								goto l385
							l386:
								position, tokenIndex, depth = position385, tokenIndex385, depth385
								if buffer[position] != rune('O') {
									goto l381
								}
								position++
							}
						l385:
							{
								position387, tokenIndex387, depth387 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l388
								}
								position++
								goto l387
							l388:
								position, tokenIndex, depth = position387, tokenIndex387, depth387
								if buffer[position] != rune('L') {
									goto l381
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
									goto l381
								}
								position++
							}
						l389:
							{
								position391, tokenIndex391, depth391 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l392
								}
								position++
								goto l391
							l392:
								position, tokenIndex, depth = position391, tokenIndex391, depth391
								if buffer[position] != rune('A') {
									goto l381
								}
								position++
							}
						l391:
							{
								position393, tokenIndex393, depth393 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l394
								}
								position++
								goto l393
							l394:
								position, tokenIndex, depth = position393, tokenIndex393, depth393
								if buffer[position] != rune('P') {
									goto l381
								}
								position++
							}
						l393:
							{
								position395, tokenIndex395, depth395 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l396
								}
								position++
								goto l395
							l396:
								position, tokenIndex, depth = position395, tokenIndex395, depth395
								if buffer[position] != rune('S') {
									goto l381
								}
								position++
							}
						l395:
							{
								position397, tokenIndex397, depth397 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l398
								}
								position++
								goto l397
							l398:
								position, tokenIndex, depth = position397, tokenIndex397, depth397
								if buffer[position] != rune('E') {
									goto l381
								}
								position++
							}
						l397:
							if !_rules[ruleKEY]() {
								goto l381
							}
							{
								position399, tokenIndex399, depth399 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l400
								}
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
										goto l400
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
										goto l400
									}
									position++
								}
							l403:
								if !_rules[ruleKEY]() {
									goto l400
								}
								goto l399
							l400:
								position, tokenIndex, depth = position399, tokenIndex399, depth399
								if !(p.errorHere(position, `expected keyword "by" to follow keyword "collapse" in "collapse by" clause`)) {
									goto l381
								}
							}
						l399:
							{
								position405, tokenIndex405, depth405 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l406
								}
								{
									position407 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l406
									}
									depth--
									add(rulePegText, position407)
								}
								goto l405
							l406:
								position, tokenIndex, depth = position405, tokenIndex405, depth405
								if !(p.errorHere(position, `expected tag key identifier to follow "collapse by" keywords in "collapse by" clause`)) {
									goto l381
								}
							}
						l405:
							{
								add(ruleAction39, position)
							}
							{
								add(ruleAction40, position)
							}
						l410:
							{
								position411, tokenIndex411, depth411 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l411
								}
								if !_rules[ruleCOMMA]() {
									goto l411
								}
								{
									position412, tokenIndex412, depth412 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l413
									}
									{
										position414 := position
										depth++
										if !_rules[ruleCOLUMN_NAME]() {
											goto l413
										}
										depth--
										add(rulePegText, position414)
									}
									goto l412
								l413:
									position, tokenIndex, depth = position412, tokenIndex412, depth412
									if !(p.errorHere(position, `expected tag key identifier to follow "," in "collapse by" clause`)) {
										goto l411
									}
								}
							l412:
								{
									add(ruleAction41, position)
								}
								goto l410
							l411:
								position, tokenIndex, depth = position411, tokenIndex411, depth411
							}
							depth--
							add(rulecollapseByClause, position382)
						}
						goto l351
					l381:
						position, tokenIndex, depth = position351, tokenIndex351, depth351
						{
							add(ruleAction30, position)
						}
					}
				l351:
					goto l350

					position, tokenIndex, depth = position349, tokenIndex349, depth349
				}
			l350:
				depth--
				add(ruleoptionalGroupBy, position348)
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
			position422, tokenIndex422, depth422 := position, tokenIndex, depth
			{
				position423 := position
				depth++
				{
					position424, tokenIndex424, depth424 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l425
					}
					if !_rules[rule_]() {
						goto l425
					}
					{
						position426 := position
						depth++
						{
							position427, tokenIndex427, depth427 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l428
							}
							position++
							goto l427
						l428:
							position, tokenIndex, depth = position427, tokenIndex427, depth427
							if buffer[position] != rune('O') {
								goto l425
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
								goto l425
							}
							position++
						}
					l429:
						if !_rules[ruleKEY]() {
							goto l425
						}
						depth--
						add(ruleOP_OR, position426)
					}
					{
						position431, tokenIndex431, depth431 := position, tokenIndex, depth
						if !_rules[rulepredicate_1]() {
							goto l432
						}
						goto l431
					l432:
						position, tokenIndex, depth = position431, tokenIndex431, depth431
						if !(p.errorHere(position, `expected predicate to follow "or" operator`)) {
							goto l425
						}
					}
				l431:
					{
						add(ruleAction42, position)
					}
					goto l424
				l425:
					position, tokenIndex, depth = position424, tokenIndex424, depth424
					if !_rules[rulepredicate_2]() {
						goto l422
					}
				}
			l424:
				depth--
				add(rulepredicate_1, position423)
			}
			return true
		l422:
			position, tokenIndex, depth = position422, tokenIndex422, depth422
			return false
		},
		/* 27 predicate_2 <- <((predicate_3 _ OP_AND (predicate_2 / &{ p.errorHere(position, `expected predicate to follow "and" operator`) }) Action43) / predicate_3)> */
		func() bool {
			position434, tokenIndex434, depth434 := position, tokenIndex, depth
			{
				position435 := position
				depth++
				{
					position436, tokenIndex436, depth436 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l437
					}
					if !_rules[rule_]() {
						goto l437
					}
					{
						position438 := position
						depth++
						{
							position439, tokenIndex439, depth439 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l440
							}
							position++
							goto l439
						l440:
							position, tokenIndex, depth = position439, tokenIndex439, depth439
							if buffer[position] != rune('A') {
								goto l437
							}
							position++
						}
					l439:
						{
							position441, tokenIndex441, depth441 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l442
							}
							position++
							goto l441
						l442:
							position, tokenIndex, depth = position441, tokenIndex441, depth441
							if buffer[position] != rune('N') {
								goto l437
							}
							position++
						}
					l441:
						{
							position443, tokenIndex443, depth443 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l444
							}
							position++
							goto l443
						l444:
							position, tokenIndex, depth = position443, tokenIndex443, depth443
							if buffer[position] != rune('D') {
								goto l437
							}
							position++
						}
					l443:
						if !_rules[ruleKEY]() {
							goto l437
						}
						depth--
						add(ruleOP_AND, position438)
					}
					{
						position445, tokenIndex445, depth445 := position, tokenIndex, depth
						if !_rules[rulepredicate_2]() {
							goto l446
						}
						goto l445
					l446:
						position, tokenIndex, depth = position445, tokenIndex445, depth445
						if !(p.errorHere(position, `expected predicate to follow "and" operator`)) {
							goto l437
						}
					}
				l445:
					{
						add(ruleAction43, position)
					}
					goto l436
				l437:
					position, tokenIndex, depth = position436, tokenIndex436, depth436
					if !_rules[rulepredicate_3]() {
						goto l434
					}
				}
			l436:
				depth--
				add(rulepredicate_2, position435)
			}
			return true
		l434:
			position, tokenIndex, depth = position434, tokenIndex434, depth434
			return false
		},
		/* 28 predicate_3 <- <((_ OP_NOT (predicate_3 / &{ p.errorHere(position, `expected predicate to follow "not" operator`) }) Action44) / (_ PAREN_OPEN (predicate_1 / &{ p.errorHere(position, `expected predicate to follow "("`) }) ((_ PAREN_CLOSE) / &{ p.errorHere(position, `expected ")" to close "(" opened in predicate`) })) / tagMatcher)> */
		func() bool {
			position448, tokenIndex448, depth448 := position, tokenIndex, depth
			{
				position449 := position
				depth++
				{
					position450, tokenIndex450, depth450 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l451
					}
					{
						position452 := position
						depth++
						{
							position453, tokenIndex453, depth453 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l454
							}
							position++
							goto l453
						l454:
							position, tokenIndex, depth = position453, tokenIndex453, depth453
							if buffer[position] != rune('N') {
								goto l451
							}
							position++
						}
					l453:
						{
							position455, tokenIndex455, depth455 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l456
							}
							position++
							goto l455
						l456:
							position, tokenIndex, depth = position455, tokenIndex455, depth455
							if buffer[position] != rune('O') {
								goto l451
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
								goto l451
							}
							position++
						}
					l457:
						if !_rules[ruleKEY]() {
							goto l451
						}
						depth--
						add(ruleOP_NOT, position452)
					}
					{
						position459, tokenIndex459, depth459 := position, tokenIndex, depth
						if !_rules[rulepredicate_3]() {
							goto l460
						}
						goto l459
					l460:
						position, tokenIndex, depth = position459, tokenIndex459, depth459
						if !(p.errorHere(position, `expected predicate to follow "not" operator`)) {
							goto l451
						}
					}
				l459:
					{
						add(ruleAction44, position)
					}
					goto l450
				l451:
					position, tokenIndex, depth = position450, tokenIndex450, depth450
					if !_rules[rule_]() {
						goto l462
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l462
					}
					{
						position463, tokenIndex463, depth463 := position, tokenIndex, depth
						if !_rules[rulepredicate_1]() {
							goto l464
						}
						goto l463
					l464:
						position, tokenIndex, depth = position463, tokenIndex463, depth463
						if !(p.errorHere(position, `expected predicate to follow "("`)) {
							goto l462
						}
					}
				l463:
					{
						position465, tokenIndex465, depth465 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l466
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l466
						}
						goto l465
					l466:
						position, tokenIndex, depth = position465, tokenIndex465, depth465
						if !(p.errorHere(position, `expected ")" to close "(" opened in predicate`)) {
							goto l462
						}
					}
				l465:
					goto l450
				l462:
					position, tokenIndex, depth = position450, tokenIndex450, depth450
					{
						position467 := position
						depth++
						if !_rules[ruletagName]() {
							goto l448
						}
						{
							position468, tokenIndex468, depth468 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l469
							}
							if buffer[position] != rune('=') {
								goto l469
							}
							position++
							{
								position470, tokenIndex470, depth470 := position, tokenIndex, depth
								if !_rules[ruleliteralString]() {
									goto l471
								}
								goto l470
							l471:
								position, tokenIndex, depth = position470, tokenIndex470, depth470
								if !(p.errorHere(position, `expected string literal to follow "="`)) {
									goto l469
								}
							}
						l470:
							{
								add(ruleAction45, position)
							}
							goto l468
						l469:
							position, tokenIndex, depth = position468, tokenIndex468, depth468
							if !_rules[rule_]() {
								goto l473
							}
							if buffer[position] != rune('!') {
								goto l473
							}
							position++
							if buffer[position] != rune('=') {
								goto l473
							}
							position++
							{
								position474, tokenIndex474, depth474 := position, tokenIndex, depth
								if !_rules[ruleliteralString]() {
									goto l475
								}
								goto l474
							l475:
								position, tokenIndex, depth = position474, tokenIndex474, depth474
								if !(p.errorHere(position, `expected string literal to follow "!="`)) {
									goto l473
								}
							}
						l474:
							{
								add(ruleAction46, position)
							}
							{
								add(ruleAction47, position)
							}
							goto l468
						l473:
							position, tokenIndex, depth = position468, tokenIndex468, depth468
							if !_rules[rule_]() {
								goto l478
							}
							{
								position479, tokenIndex479, depth479 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l480
								}
								position++
								goto l479
							l480:
								position, tokenIndex, depth = position479, tokenIndex479, depth479
								if buffer[position] != rune('M') {
									goto l478
								}
								position++
							}
						l479:
							{
								position481, tokenIndex481, depth481 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l482
								}
								position++
								goto l481
							l482:
								position, tokenIndex, depth = position481, tokenIndex481, depth481
								if buffer[position] != rune('A') {
									goto l478
								}
								position++
							}
						l481:
							{
								position483, tokenIndex483, depth483 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l484
								}
								position++
								goto l483
							l484:
								position, tokenIndex, depth = position483, tokenIndex483, depth483
								if buffer[position] != rune('T') {
									goto l478
								}
								position++
							}
						l483:
							{
								position485, tokenIndex485, depth485 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l486
								}
								position++
								goto l485
							l486:
								position, tokenIndex, depth = position485, tokenIndex485, depth485
								if buffer[position] != rune('C') {
									goto l478
								}
								position++
							}
						l485:
							{
								position487, tokenIndex487, depth487 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l488
								}
								position++
								goto l487
							l488:
								position, tokenIndex, depth = position487, tokenIndex487, depth487
								if buffer[position] != rune('H') {
									goto l478
								}
								position++
							}
						l487:
							if !_rules[ruleKEY]() {
								goto l478
							}
							{
								position489, tokenIndex489, depth489 := position, tokenIndex, depth
								if !_rules[ruleliteralString]() {
									goto l490
								}
								goto l489
							l490:
								position, tokenIndex, depth = position489, tokenIndex489, depth489
								if !(p.errorHere(position, `expected regex string literal to follow "match"`)) {
									goto l478
								}
							}
						l489:
							{
								add(ruleAction48, position)
							}
							goto l468
						l478:
							position, tokenIndex, depth = position468, tokenIndex468, depth468
							if !_rules[rule_]() {
								goto l492
							}
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
									goto l492
								}
								position++
							}
						l493:
							{
								position495, tokenIndex495, depth495 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l496
								}
								position++
								goto l495
							l496:
								position, tokenIndex, depth = position495, tokenIndex495, depth495
								if buffer[position] != rune('N') {
									goto l492
								}
								position++
							}
						l495:
							if !_rules[ruleKEY]() {
								goto l492
							}
							{
								position497, tokenIndex497, depth497 := position, tokenIndex, depth
								{
									position499 := position
									depth++
									{
										add(ruleAction51, position)
									}
									if !_rules[rule_]() {
										goto l498
									}
									if !_rules[rulePAREN_OPEN]() {
										goto l498
									}
									{
										position501, tokenIndex501, depth501 := position, tokenIndex, depth
										if !_rules[ruleliteralListString]() {
											goto l502
										}
										goto l501
									l502:
										position, tokenIndex, depth = position501, tokenIndex501, depth501
										if !(p.errorHere(position, `expected string literal to follow "(" in literal list`)) {
											goto l498
										}
									}
								l501:
								l503:
									{
										position504, tokenIndex504, depth504 := position, tokenIndex, depth
										if !_rules[rule_]() {
											goto l504
										}
										if !_rules[ruleCOMMA]() {
											goto l504
										}
										{
											position505, tokenIndex505, depth505 := position, tokenIndex, depth
											if !_rules[ruleliteralListString]() {
												goto l506
											}
											goto l505
										l506:
											position, tokenIndex, depth = position505, tokenIndex505, depth505
											if !(p.errorHere(position, `expected string literal to follow "," in literal list`)) {
												goto l504
											}
										}
									l505:
										goto l503
									l504:
										position, tokenIndex, depth = position504, tokenIndex504, depth504
									}
									{
										position507, tokenIndex507, depth507 := position, tokenIndex, depth
										if !_rules[rule_]() {
											goto l508
										}
										if !_rules[rulePAREN_CLOSE]() {
											goto l508
										}
										goto l507
									l508:
										position, tokenIndex, depth = position507, tokenIndex507, depth507
										if !(p.errorHere(position, `expected ")" to close "(" for literal list`)) {
											goto l498
										}
									}
								l507:
									depth--
									add(ruleliteralList, position499)
								}
								goto l497
							l498:
								position, tokenIndex, depth = position497, tokenIndex497, depth497
								if !(p.errorHere(position, `expected string literal list to follow "in" keyword`)) {
									goto l492
								}
							}
						l497:
							{
								add(ruleAction49, position)
							}
							goto l468
						l492:
							position, tokenIndex, depth = position468, tokenIndex468, depth468
							if !(p.errorHere(position, `expected "=", "!=", "match", or "in" to follow tag key in predicate`)) {
								goto l448
							}
						}
					l468:
						depth--
						add(ruletagMatcher, position467)
					}
				}
			l450:
				depth--
				add(rulepredicate_3, position449)
			}
			return true
		l448:
			position, tokenIndex, depth = position448, tokenIndex448, depth448
			return false
		},
		/* 29 tagMatcher <- <(tagName ((_ '=' (literalString / &{ p.errorHere(position, `expected string literal to follow "="`) }) Action45) / (_ ('!' '=') (literalString / &{ p.errorHere(position, `expected string literal to follow "!="`) }) Action46 Action47) / (_ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H')) KEY (literalString / &{ p.errorHere(position, `expected regex string literal to follow "match"`) }) Action48) / (_ (('i' / 'I') ('n' / 'N')) KEY (literalList / &{ p.errorHere(position, `expected string literal list to follow "in" keyword`) }) Action49) / &{ p.errorHere(position, `expected "=", "!=", "match", or "in" to follow tag key in predicate`) }))> */
		nil,
		/* 30 literalString <- <(_ STRING Action50)> */
		func() bool {
			position511, tokenIndex511, depth511 := position, tokenIndex, depth
			{
				position512 := position
				depth++
				if !_rules[rule_]() {
					goto l511
				}
				if !_rules[ruleSTRING]() {
					goto l511
				}
				{
					add(ruleAction50, position)
				}
				depth--
				add(ruleliteralString, position512)
			}
			return true
		l511:
			position, tokenIndex, depth = position511, tokenIndex511, depth511
			return false
		},
		/* 31 literalList <- <(Action51 _ PAREN_OPEN (literalListString / &{ p.errorHere(position, `expected string literal to follow "(" in literal list`) }) (_ COMMA (literalListString / &{ p.errorHere(position, `expected string literal to follow "," in literal list`) }))* ((_ PAREN_CLOSE) / &{ p.errorHere(position, `expected ")" to close "(" for literal list`) }))> */
		nil,
		/* 32 literalListString <- <(_ STRING Action52)> */
		func() bool {
			position515, tokenIndex515, depth515 := position, tokenIndex, depth
			{
				position516 := position
				depth++
				if !_rules[rule_]() {
					goto l515
				}
				if !_rules[ruleSTRING]() {
					goto l515
				}
				{
					add(ruleAction52, position)
				}
				depth--
				add(ruleliteralListString, position516)
			}
			return true
		l515:
			position, tokenIndex, depth = position515, tokenIndex515, depth515
			return false
		},
		/* 33 tagName <- <(_ <TAG_NAME> Action53)> */
		func() bool {
			position518, tokenIndex518, depth518 := position, tokenIndex, depth
			{
				position519 := position
				depth++
				if !_rules[rule_]() {
					goto l518
				}
				{
					position520 := position
					depth++
					{
						position521 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l518
						}
						depth--
						add(ruleTAG_NAME, position521)
					}
					depth--
					add(rulePegText, position520)
				}
				{
					add(ruleAction53, position)
				}
				depth--
				add(ruletagName, position519)
			}
			return true
		l518:
			position, tokenIndex, depth = position518, tokenIndex518, depth518
			return false
		},
		/* 34 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position523, tokenIndex523, depth523 := position, tokenIndex, depth
			{
				position524 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l523
				}
				depth--
				add(ruleCOLUMN_NAME, position524)
			}
			return true
		l523:
			position, tokenIndex, depth = position523, tokenIndex523, depth523
			return false
		},
		/* 35 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 36 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 37 IDENTIFIER <- <(('`' CHAR* ('`' / &{ p.errorHere(position, "expected \"`\" to end identifier") })) / (!(KEYWORD KEY) ID_SEGMENT ('.' (ID_SEGMENT / &{ p.errorHere(position, `expected identifier segment to follow "."`) }))*))> */
		func() bool {
			position527, tokenIndex527, depth527 := position, tokenIndex, depth
			{
				position528 := position
				depth++
				{
					position529, tokenIndex529, depth529 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l530
					}
					position++
				l531:
					{
						position532, tokenIndex532, depth532 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l532
						}
						goto l531
					l532:
						position, tokenIndex, depth = position532, tokenIndex532, depth532
					}
					{
						position533, tokenIndex533, depth533 := position, tokenIndex, depth
						if buffer[position] != rune('`') {
							goto l534
						}
						position++
						goto l533
					l534:
						position, tokenIndex, depth = position533, tokenIndex533, depth533
						if !(p.errorHere(position, "expected \"`\" to end identifier")) {
							goto l530
						}
					}
				l533:
					goto l529
				l530:
					position, tokenIndex, depth = position529, tokenIndex529, depth529
					{
						position535, tokenIndex535, depth535 := position, tokenIndex, depth
						{
							position536 := position
							depth++
							{
								position537, tokenIndex537, depth537 := position, tokenIndex, depth
								{
									position539, tokenIndex539, depth539 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l540
									}
									position++
									goto l539
								l540:
									position, tokenIndex, depth = position539, tokenIndex539, depth539
									if buffer[position] != rune('A') {
										goto l538
									}
									position++
								}
							l539:
								{
									position541, tokenIndex541, depth541 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l542
									}
									position++
									goto l541
								l542:
									position, tokenIndex, depth = position541, tokenIndex541, depth541
									if buffer[position] != rune('L') {
										goto l538
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
										goto l538
									}
									position++
								}
							l543:
								goto l537
							l538:
								position, tokenIndex, depth = position537, tokenIndex537, depth537
								{
									position546, tokenIndex546, depth546 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l547
									}
									position++
									goto l546
								l547:
									position, tokenIndex, depth = position546, tokenIndex546, depth546
									if buffer[position] != rune('A') {
										goto l545
									}
									position++
								}
							l546:
								{
									position548, tokenIndex548, depth548 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l549
									}
									position++
									goto l548
								l549:
									position, tokenIndex, depth = position548, tokenIndex548, depth548
									if buffer[position] != rune('N') {
										goto l545
									}
									position++
								}
							l548:
								{
									position550, tokenIndex550, depth550 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l551
									}
									position++
									goto l550
								l551:
									position, tokenIndex, depth = position550, tokenIndex550, depth550
									if buffer[position] != rune('D') {
										goto l545
									}
									position++
								}
							l550:
								goto l537
							l545:
								position, tokenIndex, depth = position537, tokenIndex537, depth537
								{
									position553, tokenIndex553, depth553 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l554
									}
									position++
									goto l553
								l554:
									position, tokenIndex, depth = position553, tokenIndex553, depth553
									if buffer[position] != rune('M') {
										goto l552
									}
									position++
								}
							l553:
								{
									position555, tokenIndex555, depth555 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l556
									}
									position++
									goto l555
								l556:
									position, tokenIndex, depth = position555, tokenIndex555, depth555
									if buffer[position] != rune('A') {
										goto l552
									}
									position++
								}
							l555:
								{
									position557, tokenIndex557, depth557 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l558
									}
									position++
									goto l557
								l558:
									position, tokenIndex, depth = position557, tokenIndex557, depth557
									if buffer[position] != rune('T') {
										goto l552
									}
									position++
								}
							l557:
								{
									position559, tokenIndex559, depth559 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l560
									}
									position++
									goto l559
								l560:
									position, tokenIndex, depth = position559, tokenIndex559, depth559
									if buffer[position] != rune('C') {
										goto l552
									}
									position++
								}
							l559:
								{
									position561, tokenIndex561, depth561 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l562
									}
									position++
									goto l561
								l562:
									position, tokenIndex, depth = position561, tokenIndex561, depth561
									if buffer[position] != rune('H') {
										goto l552
									}
									position++
								}
							l561:
								goto l537
							l552:
								position, tokenIndex, depth = position537, tokenIndex537, depth537
								{
									position564, tokenIndex564, depth564 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l565
									}
									position++
									goto l564
								l565:
									position, tokenIndex, depth = position564, tokenIndex564, depth564
									if buffer[position] != rune('S') {
										goto l563
									}
									position++
								}
							l564:
								{
									position566, tokenIndex566, depth566 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l567
									}
									position++
									goto l566
								l567:
									position, tokenIndex, depth = position566, tokenIndex566, depth566
									if buffer[position] != rune('E') {
										goto l563
									}
									position++
								}
							l566:
								{
									position568, tokenIndex568, depth568 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l569
									}
									position++
									goto l568
								l569:
									position, tokenIndex, depth = position568, tokenIndex568, depth568
									if buffer[position] != rune('L') {
										goto l563
									}
									position++
								}
							l568:
								{
									position570, tokenIndex570, depth570 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l571
									}
									position++
									goto l570
								l571:
									position, tokenIndex, depth = position570, tokenIndex570, depth570
									if buffer[position] != rune('E') {
										goto l563
									}
									position++
								}
							l570:
								{
									position572, tokenIndex572, depth572 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l573
									}
									position++
									goto l572
								l573:
									position, tokenIndex, depth = position572, tokenIndex572, depth572
									if buffer[position] != rune('C') {
										goto l563
									}
									position++
								}
							l572:
								{
									position574, tokenIndex574, depth574 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l575
									}
									position++
									goto l574
								l575:
									position, tokenIndex, depth = position574, tokenIndex574, depth574
									if buffer[position] != rune('T') {
										goto l563
									}
									position++
								}
							l574:
								goto l537
							l563:
								position, tokenIndex, depth = position537, tokenIndex537, depth537
								{
									switch buffer[position] {
									case 'S', 's':
										{
											position577, tokenIndex577, depth577 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l578
											}
											position++
											goto l577
										l578:
											position, tokenIndex, depth = position577, tokenIndex577, depth577
											if buffer[position] != rune('S') {
												goto l535
											}
											position++
										}
									l577:
										{
											position579, tokenIndex579, depth579 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l580
											}
											position++
											goto l579
										l580:
											position, tokenIndex, depth = position579, tokenIndex579, depth579
											if buffer[position] != rune('A') {
												goto l535
											}
											position++
										}
									l579:
										{
											position581, tokenIndex581, depth581 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l582
											}
											position++
											goto l581
										l582:
											position, tokenIndex, depth = position581, tokenIndex581, depth581
											if buffer[position] != rune('M') {
												goto l535
											}
											position++
										}
									l581:
										{
											position583, tokenIndex583, depth583 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l584
											}
											position++
											goto l583
										l584:
											position, tokenIndex, depth = position583, tokenIndex583, depth583
											if buffer[position] != rune('P') {
												goto l535
											}
											position++
										}
									l583:
										{
											position585, tokenIndex585, depth585 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l586
											}
											position++
											goto l585
										l586:
											position, tokenIndex, depth = position585, tokenIndex585, depth585
											if buffer[position] != rune('L') {
												goto l535
											}
											position++
										}
									l585:
										{
											position587, tokenIndex587, depth587 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l588
											}
											position++
											goto l587
										l588:
											position, tokenIndex, depth = position587, tokenIndex587, depth587
											if buffer[position] != rune('E') {
												goto l535
											}
											position++
										}
									l587:
										break
									case 'R', 'r':
										{
											position589, tokenIndex589, depth589 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l590
											}
											position++
											goto l589
										l590:
											position, tokenIndex, depth = position589, tokenIndex589, depth589
											if buffer[position] != rune('R') {
												goto l535
											}
											position++
										}
									l589:
										{
											position591, tokenIndex591, depth591 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l592
											}
											position++
											goto l591
										l592:
											position, tokenIndex, depth = position591, tokenIndex591, depth591
											if buffer[position] != rune('E') {
												goto l535
											}
											position++
										}
									l591:
										{
											position593, tokenIndex593, depth593 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l594
											}
											position++
											goto l593
										l594:
											position, tokenIndex, depth = position593, tokenIndex593, depth593
											if buffer[position] != rune('S') {
												goto l535
											}
											position++
										}
									l593:
										{
											position595, tokenIndex595, depth595 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l596
											}
											position++
											goto l595
										l596:
											position, tokenIndex, depth = position595, tokenIndex595, depth595
											if buffer[position] != rune('O') {
												goto l535
											}
											position++
										}
									l595:
										{
											position597, tokenIndex597, depth597 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l598
											}
											position++
											goto l597
										l598:
											position, tokenIndex, depth = position597, tokenIndex597, depth597
											if buffer[position] != rune('L') {
												goto l535
											}
											position++
										}
									l597:
										{
											position599, tokenIndex599, depth599 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l600
											}
											position++
											goto l599
										l600:
											position, tokenIndex, depth = position599, tokenIndex599, depth599
											if buffer[position] != rune('U') {
												goto l535
											}
											position++
										}
									l599:
										{
											position601, tokenIndex601, depth601 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l602
											}
											position++
											goto l601
										l602:
											position, tokenIndex, depth = position601, tokenIndex601, depth601
											if buffer[position] != rune('T') {
												goto l535
											}
											position++
										}
									l601:
										{
											position603, tokenIndex603, depth603 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l604
											}
											position++
											goto l603
										l604:
											position, tokenIndex, depth = position603, tokenIndex603, depth603
											if buffer[position] != rune('I') {
												goto l535
											}
											position++
										}
									l603:
										{
											position605, tokenIndex605, depth605 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l606
											}
											position++
											goto l605
										l606:
											position, tokenIndex, depth = position605, tokenIndex605, depth605
											if buffer[position] != rune('O') {
												goto l535
											}
											position++
										}
									l605:
										{
											position607, tokenIndex607, depth607 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l608
											}
											position++
											goto l607
										l608:
											position, tokenIndex, depth = position607, tokenIndex607, depth607
											if buffer[position] != rune('N') {
												goto l535
											}
											position++
										}
									l607:
										break
									case 'T', 't':
										{
											position609, tokenIndex609, depth609 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l610
											}
											position++
											goto l609
										l610:
											position, tokenIndex, depth = position609, tokenIndex609, depth609
											if buffer[position] != rune('T') {
												goto l535
											}
											position++
										}
									l609:
										{
											position611, tokenIndex611, depth611 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l612
											}
											position++
											goto l611
										l612:
											position, tokenIndex, depth = position611, tokenIndex611, depth611
											if buffer[position] != rune('O') {
												goto l535
											}
											position++
										}
									l611:
										break
									case 'F', 'f':
										{
											position613, tokenIndex613, depth613 := position, tokenIndex, depth
											if buffer[position] != rune('f') {
												goto l614
											}
											position++
											goto l613
										l614:
											position, tokenIndex, depth = position613, tokenIndex613, depth613
											if buffer[position] != rune('F') {
												goto l535
											}
											position++
										}
									l613:
										{
											position615, tokenIndex615, depth615 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l616
											}
											position++
											goto l615
										l616:
											position, tokenIndex, depth = position615, tokenIndex615, depth615
											if buffer[position] != rune('R') {
												goto l535
											}
											position++
										}
									l615:
										{
											position617, tokenIndex617, depth617 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l618
											}
											position++
											goto l617
										l618:
											position, tokenIndex, depth = position617, tokenIndex617, depth617
											if buffer[position] != rune('O') {
												goto l535
											}
											position++
										}
									l617:
										{
											position619, tokenIndex619, depth619 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l620
											}
											position++
											goto l619
										l620:
											position, tokenIndex, depth = position619, tokenIndex619, depth619
											if buffer[position] != rune('M') {
												goto l535
											}
											position++
										}
									l619:
										break
									case 'M', 'm':
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
												goto l535
											}
											position++
										}
									l621:
										{
											position623, tokenIndex623, depth623 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l624
											}
											position++
											goto l623
										l624:
											position, tokenIndex, depth = position623, tokenIndex623, depth623
											if buffer[position] != rune('E') {
												goto l535
											}
											position++
										}
									l623:
										{
											position625, tokenIndex625, depth625 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l626
											}
											position++
											goto l625
										l626:
											position, tokenIndex, depth = position625, tokenIndex625, depth625
											if buffer[position] != rune('T') {
												goto l535
											}
											position++
										}
									l625:
										{
											position627, tokenIndex627, depth627 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l628
											}
											position++
											goto l627
										l628:
											position, tokenIndex, depth = position627, tokenIndex627, depth627
											if buffer[position] != rune('R') {
												goto l535
											}
											position++
										}
									l627:
										{
											position629, tokenIndex629, depth629 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l630
											}
											position++
											goto l629
										l630:
											position, tokenIndex, depth = position629, tokenIndex629, depth629
											if buffer[position] != rune('I') {
												goto l535
											}
											position++
										}
									l629:
										{
											position631, tokenIndex631, depth631 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l632
											}
											position++
											goto l631
										l632:
											position, tokenIndex, depth = position631, tokenIndex631, depth631
											if buffer[position] != rune('C') {
												goto l535
											}
											position++
										}
									l631:
										{
											position633, tokenIndex633, depth633 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l634
											}
											position++
											goto l633
										l634:
											position, tokenIndex, depth = position633, tokenIndex633, depth633
											if buffer[position] != rune('S') {
												goto l535
											}
											position++
										}
									l633:
										break
									case 'W', 'w':
										{
											position635, tokenIndex635, depth635 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l636
											}
											position++
											goto l635
										l636:
											position, tokenIndex, depth = position635, tokenIndex635, depth635
											if buffer[position] != rune('W') {
												goto l535
											}
											position++
										}
									l635:
										{
											position637, tokenIndex637, depth637 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l638
											}
											position++
											goto l637
										l638:
											position, tokenIndex, depth = position637, tokenIndex637, depth637
											if buffer[position] != rune('H') {
												goto l535
											}
											position++
										}
									l637:
										{
											position639, tokenIndex639, depth639 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l640
											}
											position++
											goto l639
										l640:
											position, tokenIndex, depth = position639, tokenIndex639, depth639
											if buffer[position] != rune('E') {
												goto l535
											}
											position++
										}
									l639:
										{
											position641, tokenIndex641, depth641 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l642
											}
											position++
											goto l641
										l642:
											position, tokenIndex, depth = position641, tokenIndex641, depth641
											if buffer[position] != rune('R') {
												goto l535
											}
											position++
										}
									l641:
										{
											position643, tokenIndex643, depth643 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l644
											}
											position++
											goto l643
										l644:
											position, tokenIndex, depth = position643, tokenIndex643, depth643
											if buffer[position] != rune('E') {
												goto l535
											}
											position++
										}
									l643:
										break
									case 'O', 'o':
										{
											position645, tokenIndex645, depth645 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l646
											}
											position++
											goto l645
										l646:
											position, tokenIndex, depth = position645, tokenIndex645, depth645
											if buffer[position] != rune('O') {
												goto l535
											}
											position++
										}
									l645:
										{
											position647, tokenIndex647, depth647 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l648
											}
											position++
											goto l647
										l648:
											position, tokenIndex, depth = position647, tokenIndex647, depth647
											if buffer[position] != rune('R') {
												goto l535
											}
											position++
										}
									l647:
										break
									case 'N', 'n':
										{
											position649, tokenIndex649, depth649 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l650
											}
											position++
											goto l649
										l650:
											position, tokenIndex, depth = position649, tokenIndex649, depth649
											if buffer[position] != rune('N') {
												goto l535
											}
											position++
										}
									l649:
										{
											position651, tokenIndex651, depth651 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l652
											}
											position++
											goto l651
										l652:
											position, tokenIndex, depth = position651, tokenIndex651, depth651
											if buffer[position] != rune('O') {
												goto l535
											}
											position++
										}
									l651:
										{
											position653, tokenIndex653, depth653 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l654
											}
											position++
											goto l653
										l654:
											position, tokenIndex, depth = position653, tokenIndex653, depth653
											if buffer[position] != rune('T') {
												goto l535
											}
											position++
										}
									l653:
										break
									case 'I', 'i':
										{
											position655, tokenIndex655, depth655 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l656
											}
											position++
											goto l655
										l656:
											position, tokenIndex, depth = position655, tokenIndex655, depth655
											if buffer[position] != rune('I') {
												goto l535
											}
											position++
										}
									l655:
										{
											position657, tokenIndex657, depth657 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l658
											}
											position++
											goto l657
										l658:
											position, tokenIndex, depth = position657, tokenIndex657, depth657
											if buffer[position] != rune('N') {
												goto l535
											}
											position++
										}
									l657:
										break
									case 'C', 'c':
										{
											position659, tokenIndex659, depth659 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l660
											}
											position++
											goto l659
										l660:
											position, tokenIndex, depth = position659, tokenIndex659, depth659
											if buffer[position] != rune('C') {
												goto l535
											}
											position++
										}
									l659:
										{
											position661, tokenIndex661, depth661 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l662
											}
											position++
											goto l661
										l662:
											position, tokenIndex, depth = position661, tokenIndex661, depth661
											if buffer[position] != rune('O') {
												goto l535
											}
											position++
										}
									l661:
										{
											position663, tokenIndex663, depth663 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l664
											}
											position++
											goto l663
										l664:
											position, tokenIndex, depth = position663, tokenIndex663, depth663
											if buffer[position] != rune('L') {
												goto l535
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
												goto l535
											}
											position++
										}
									l665:
										{
											position667, tokenIndex667, depth667 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l668
											}
											position++
											goto l667
										l668:
											position, tokenIndex, depth = position667, tokenIndex667, depth667
											if buffer[position] != rune('A') {
												goto l535
											}
											position++
										}
									l667:
										{
											position669, tokenIndex669, depth669 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l670
											}
											position++
											goto l669
										l670:
											position, tokenIndex, depth = position669, tokenIndex669, depth669
											if buffer[position] != rune('P') {
												goto l535
											}
											position++
										}
									l669:
										{
											position671, tokenIndex671, depth671 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l672
											}
											position++
											goto l671
										l672:
											position, tokenIndex, depth = position671, tokenIndex671, depth671
											if buffer[position] != rune('S') {
												goto l535
											}
											position++
										}
									l671:
										{
											position673, tokenIndex673, depth673 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l674
											}
											position++
											goto l673
										l674:
											position, tokenIndex, depth = position673, tokenIndex673, depth673
											if buffer[position] != rune('E') {
												goto l535
											}
											position++
										}
									l673:
										break
									case 'G', 'g':
										{
											position675, tokenIndex675, depth675 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l676
											}
											position++
											goto l675
										l676:
											position, tokenIndex, depth = position675, tokenIndex675, depth675
											if buffer[position] != rune('G') {
												goto l535
											}
											position++
										}
									l675:
										{
											position677, tokenIndex677, depth677 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l678
											}
											position++
											goto l677
										l678:
											position, tokenIndex, depth = position677, tokenIndex677, depth677
											if buffer[position] != rune('R') {
												goto l535
											}
											position++
										}
									l677:
										{
											position679, tokenIndex679, depth679 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l680
											}
											position++
											goto l679
										l680:
											position, tokenIndex, depth = position679, tokenIndex679, depth679
											if buffer[position] != rune('O') {
												goto l535
											}
											position++
										}
									l679:
										{
											position681, tokenIndex681, depth681 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l682
											}
											position++
											goto l681
										l682:
											position, tokenIndex, depth = position681, tokenIndex681, depth681
											if buffer[position] != rune('U') {
												goto l535
											}
											position++
										}
									l681:
										{
											position683, tokenIndex683, depth683 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l684
											}
											position++
											goto l683
										l684:
											position, tokenIndex, depth = position683, tokenIndex683, depth683
											if buffer[position] != rune('P') {
												goto l535
											}
											position++
										}
									l683:
										break
									case 'D', 'd':
										{
											position685, tokenIndex685, depth685 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l686
											}
											position++
											goto l685
										l686:
											position, tokenIndex, depth = position685, tokenIndex685, depth685
											if buffer[position] != rune('D') {
												goto l535
											}
											position++
										}
									l685:
										{
											position687, tokenIndex687, depth687 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l688
											}
											position++
											goto l687
										l688:
											position, tokenIndex, depth = position687, tokenIndex687, depth687
											if buffer[position] != rune('E') {
												goto l535
											}
											position++
										}
									l687:
										{
											position689, tokenIndex689, depth689 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l690
											}
											position++
											goto l689
										l690:
											position, tokenIndex, depth = position689, tokenIndex689, depth689
											if buffer[position] != rune('S') {
												goto l535
											}
											position++
										}
									l689:
										{
											position691, tokenIndex691, depth691 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l692
											}
											position++
											goto l691
										l692:
											position, tokenIndex, depth = position691, tokenIndex691, depth691
											if buffer[position] != rune('C') {
												goto l535
											}
											position++
										}
									l691:
										{
											position693, tokenIndex693, depth693 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l694
											}
											position++
											goto l693
										l694:
											position, tokenIndex, depth = position693, tokenIndex693, depth693
											if buffer[position] != rune('R') {
												goto l535
											}
											position++
										}
									l693:
										{
											position695, tokenIndex695, depth695 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l696
											}
											position++
											goto l695
										l696:
											position, tokenIndex, depth = position695, tokenIndex695, depth695
											if buffer[position] != rune('I') {
												goto l535
											}
											position++
										}
									l695:
										{
											position697, tokenIndex697, depth697 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l698
											}
											position++
											goto l697
										l698:
											position, tokenIndex, depth = position697, tokenIndex697, depth697
											if buffer[position] != rune('B') {
												goto l535
											}
											position++
										}
									l697:
										{
											position699, tokenIndex699, depth699 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l700
											}
											position++
											goto l699
										l700:
											position, tokenIndex, depth = position699, tokenIndex699, depth699
											if buffer[position] != rune('E') {
												goto l535
											}
											position++
										}
									l699:
										break
									case 'B', 'b':
										{
											position701, tokenIndex701, depth701 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l702
											}
											position++
											goto l701
										l702:
											position, tokenIndex, depth = position701, tokenIndex701, depth701
											if buffer[position] != rune('B') {
												goto l535
											}
											position++
										}
									l701:
										{
											position703, tokenIndex703, depth703 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l704
											}
											position++
											goto l703
										l704:
											position, tokenIndex, depth = position703, tokenIndex703, depth703
											if buffer[position] != rune('Y') {
												goto l535
											}
											position++
										}
									l703:
										break
									default:
										{
											position705, tokenIndex705, depth705 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l706
											}
											position++
											goto l705
										l706:
											position, tokenIndex, depth = position705, tokenIndex705, depth705
											if buffer[position] != rune('A') {
												goto l535
											}
											position++
										}
									l705:
										{
											position707, tokenIndex707, depth707 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l708
											}
											position++
											goto l707
										l708:
											position, tokenIndex, depth = position707, tokenIndex707, depth707
											if buffer[position] != rune('S') {
												goto l535
											}
											position++
										}
									l707:
										break
									}
								}

							}
						l537:
							depth--
							add(ruleKEYWORD, position536)
						}
						if !_rules[ruleKEY]() {
							goto l535
						}
						goto l527
					l535:
						position, tokenIndex, depth = position535, tokenIndex535, depth535
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l527
					}
				l709:
					{
						position710, tokenIndex710, depth710 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l710
						}
						position++
						{
							position711, tokenIndex711, depth711 := position, tokenIndex, depth
							if !_rules[ruleID_SEGMENT]() {
								goto l712
							}
							goto l711
						l712:
							position, tokenIndex, depth = position711, tokenIndex711, depth711
							if !(p.errorHere(position, `expected identifier segment to follow "."`)) {
								goto l710
							}
						}
					l711:
						goto l709
					l710:
						position, tokenIndex, depth = position710, tokenIndex710, depth710
					}
				}
			l529:
				depth--
				add(ruleIDENTIFIER, position528)
			}
			return true
		l527:
			position, tokenIndex, depth = position527, tokenIndex527, depth527
			return false
		},
		/* 38 TIMESTAMP <- <((_ <(NUMBER ([a-z] / [A-Z])*)>) / (_ STRING) / (_ <(('n' / 'N') ('o' / 'O') ('w' / 'W'))> KEY))> */
		nil,
		/* 39 ID_SEGMENT <- <(ID_START ID_CONT*)> */
		func() bool {
			position714, tokenIndex714, depth714 := position, tokenIndex, depth
			{
				position715 := position
				depth++
				if !_rules[ruleID_START]() {
					goto l714
				}
			l716:
				{
					position717, tokenIndex717, depth717 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l717
					}
					goto l716
				l717:
					position, tokenIndex, depth = position717, tokenIndex717, depth717
				}
				depth--
				add(ruleID_SEGMENT, position715)
			}
			return true
		l714:
			position, tokenIndex, depth = position714, tokenIndex714, depth714
			return false
		},
		/* 40 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position718, tokenIndex718, depth718 := position, tokenIndex, depth
			{
				position719 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l718
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l718
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l718
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position719)
			}
			return true
		l718:
			position, tokenIndex, depth = position718, tokenIndex718, depth718
			return false
		},
		/* 41 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position721, tokenIndex721, depth721 := position, tokenIndex, depth
			{
				position722 := position
				depth++
				{
					position723, tokenIndex723, depth723 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l724
					}
					goto l723
				l724:
					position, tokenIndex, depth = position723, tokenIndex723, depth723
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l721
					}
					position++
				}
			l723:
				depth--
				add(ruleID_CONT, position722)
			}
			return true
		l721:
			position, tokenIndex, depth = position721, tokenIndex721, depth721
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
			position736, tokenIndex736, depth736 := position, tokenIndex, depth
			{
				position737 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l736
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position737)
			}
			return true
		l736:
			position, tokenIndex, depth = position736, tokenIndex736, depth736
			return false
		},
		/* 54 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position738, tokenIndex738, depth738 := position, tokenIndex, depth
			{
				position739 := position
				depth++
				if buffer[position] != rune('"') {
					goto l738
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position739)
			}
			return true
		l738:
			position, tokenIndex, depth = position738, tokenIndex738, depth738
			return false
		},
		/* 55 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> (QUOTE_SINGLE / &{ p.errorHere(position, `expected "'" to close string`) })) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> (QUOTE_DOUBLE / &{ p.errorHere(position, `expected '"' to close string`) })))> */
		func() bool {
			position740, tokenIndex740, depth740 := position, tokenIndex, depth
			{
				position741 := position
				depth++
				{
					position742, tokenIndex742, depth742 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l743
					}
					{
						position744 := position
						depth++
					l745:
						{
							position746, tokenIndex746, depth746 := position, tokenIndex, depth
							{
								position747, tokenIndex747, depth747 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l747
								}
								goto l746
							l747:
								position, tokenIndex, depth = position747, tokenIndex747, depth747
							}
							if !_rules[ruleCHAR]() {
								goto l746
							}
							goto l745
						l746:
							position, tokenIndex, depth = position746, tokenIndex746, depth746
						}
						depth--
						add(rulePegText, position744)
					}
					{
						position748, tokenIndex748, depth748 := position, tokenIndex, depth
						if !_rules[ruleQUOTE_SINGLE]() {
							goto l749
						}
						goto l748
					l749:
						position, tokenIndex, depth = position748, tokenIndex748, depth748
						if !(p.errorHere(position, `expected "'" to close string`)) {
							goto l743
						}
					}
				l748:
					goto l742
				l743:
					position, tokenIndex, depth = position742, tokenIndex742, depth742
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l740
					}
					{
						position750 := position
						depth++
					l751:
						{
							position752, tokenIndex752, depth752 := position, tokenIndex, depth
							{
								position753, tokenIndex753, depth753 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l753
								}
								goto l752
							l753:
								position, tokenIndex, depth = position753, tokenIndex753, depth753
							}
							if !_rules[ruleCHAR]() {
								goto l752
							}
							goto l751
						l752:
							position, tokenIndex, depth = position752, tokenIndex752, depth752
						}
						depth--
						add(rulePegText, position750)
					}
					{
						position754, tokenIndex754, depth754 := position, tokenIndex, depth
						if !_rules[ruleQUOTE_DOUBLE]() {
							goto l755
						}
						goto l754
					l755:
						position, tokenIndex, depth = position754, tokenIndex754, depth754
						if !(p.errorHere(position, `expected '"' to close string`)) {
							goto l740
						}
					}
				l754:
				}
			l742:
				depth--
				add(ruleSTRING, position741)
			}
			return true
		l740:
			position, tokenIndex, depth = position740, tokenIndex740, depth740
			return false
		},
		/* 56 CHAR <- <(('\\' ((&('"') (QUOTE_DOUBLE / &{ p.errorHere(position, "expected \"\\\", \"'\", \"`\", or '\"' to follow \"\\\" in string literal") })) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position756, tokenIndex756, depth756 := position, tokenIndex, depth
			{
				position757 := position
				depth++
				{
					position758, tokenIndex758, depth758 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l759
					}
					position++
					{
						switch buffer[position] {
						case '"':
							{
								position761, tokenIndex761, depth761 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l762
								}
								goto l761
							l762:
								position, tokenIndex, depth = position761, tokenIndex761, depth761
								if !(p.errorHere(position, "expected \"\\\", \"'\", \"`\", or '\"' to follow \"\\\" in string literal")) {
									goto l759
								}
							}
						l761:
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l759
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l759
							}
							break
						}
					}

					goto l758
				l759:
					position, tokenIndex, depth = position758, tokenIndex758, depth758
					{
						position763, tokenIndex763, depth763 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l763
						}
						goto l756
					l763:
						position, tokenIndex, depth = position763, tokenIndex763, depth763
					}
					if !matchDot() {
						goto l756
					}
				}
			l758:
				depth--
				add(ruleCHAR, position757)
			}
			return true
		l756:
			position, tokenIndex, depth = position756, tokenIndex756, depth756
			return false
		},
		/* 57 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position764, tokenIndex764, depth764 := position, tokenIndex, depth
			{
				position765 := position
				depth++
				{
					position766, tokenIndex766, depth766 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l767
					}
					position++
					goto l766
				l767:
					position, tokenIndex, depth = position766, tokenIndex766, depth766
					if buffer[position] != rune('\\') {
						goto l764
					}
					position++
				}
			l766:
				depth--
				add(ruleESCAPE_CLASS, position765)
			}
			return true
		l764:
			position, tokenIndex, depth = position764, tokenIndex764, depth764
			return false
		},
		/* 58 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position768, tokenIndex768, depth768 := position, tokenIndex, depth
			{
				position769 := position
				depth++
				{
					position770 := position
					depth++
					{
						position771, tokenIndex771, depth771 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l771
						}
						position++
						goto l772
					l771:
						position, tokenIndex, depth = position771, tokenIndex771, depth771
					}
				l772:
					{
						position773 := position
						depth++
						{
							position774, tokenIndex774, depth774 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l775
							}
							position++
							goto l774
						l775:
							position, tokenIndex, depth = position774, tokenIndex774, depth774
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l768
							}
							position++
						l776:
							{
								position777, tokenIndex777, depth777 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l777
								}
								position++
								goto l776
							l777:
								position, tokenIndex, depth = position777, tokenIndex777, depth777
							}
						}
					l774:
						depth--
						add(ruleNUMBER_NATURAL, position773)
					}
					depth--
					add(ruleNUMBER_INTEGER, position770)
				}
				{
					position778, tokenIndex778, depth778 := position, tokenIndex, depth
					{
						position780 := position
						depth++
						if buffer[position] != rune('.') {
							goto l778
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l778
						}
						position++
					l781:
						{
							position782, tokenIndex782, depth782 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l782
							}
							position++
							goto l781
						l782:
							position, tokenIndex, depth = position782, tokenIndex782, depth782
						}
						depth--
						add(ruleNUMBER_FRACTION, position780)
					}
					goto l779
				l778:
					position, tokenIndex, depth = position778, tokenIndex778, depth778
				}
			l779:
				{
					position783, tokenIndex783, depth783 := position, tokenIndex, depth
					{
						position785 := position
						depth++
						{
							position786, tokenIndex786, depth786 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l787
							}
							position++
							goto l786
						l787:
							position, tokenIndex, depth = position786, tokenIndex786, depth786
							if buffer[position] != rune('E') {
								goto l783
							}
							position++
						}
					l786:
						{
							position788, tokenIndex788, depth788 := position, tokenIndex, depth
							{
								position790, tokenIndex790, depth790 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l791
								}
								position++
								goto l790
							l791:
								position, tokenIndex, depth = position790, tokenIndex790, depth790
								if buffer[position] != rune('-') {
									goto l788
								}
								position++
							}
						l790:
							goto l789
						l788:
							position, tokenIndex, depth = position788, tokenIndex788, depth788
						}
					l789:
						{
							position792, tokenIndex792, depth792 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l793
							}
							position++
						l794:
							{
								position795, tokenIndex795, depth795 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l795
								}
								position++
								goto l794
							l795:
								position, tokenIndex, depth = position795, tokenIndex795, depth795
							}
							goto l792
						l793:
							position, tokenIndex, depth = position792, tokenIndex792, depth792
							if !(p.errorHere(position, `expected exponent`)) {
								goto l783
							}
						}
					l792:
						depth--
						add(ruleNUMBER_EXP, position785)
					}
					goto l784
				l783:
					position, tokenIndex, depth = position783, tokenIndex783, depth783
				}
			l784:
				depth--
				add(ruleNUMBER, position769)
			}
			return true
		l768:
			position, tokenIndex, depth = position768, tokenIndex768, depth768
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
			position801, tokenIndex801, depth801 := position, tokenIndex, depth
			{
				position802 := position
				depth++
				if buffer[position] != rune('(') {
					goto l801
				}
				position++
				depth--
				add(rulePAREN_OPEN, position802)
			}
			return true
		l801:
			position, tokenIndex, depth = position801, tokenIndex801, depth801
			return false
		},
		/* 65 PAREN_CLOSE <- <')'> */
		func() bool {
			position803, tokenIndex803, depth803 := position, tokenIndex, depth
			{
				position804 := position
				depth++
				if buffer[position] != rune(')') {
					goto l803
				}
				position++
				depth--
				add(rulePAREN_CLOSE, position804)
			}
			return true
		l803:
			position, tokenIndex, depth = position803, tokenIndex803, depth803
			return false
		},
		/* 66 COMMA <- <','> */
		func() bool {
			position805, tokenIndex805, depth805 := position, tokenIndex, depth
			{
				position806 := position
				depth++
				if buffer[position] != rune(',') {
					goto l805
				}
				position++
				depth--
				add(ruleCOMMA, position806)
			}
			return true
		l805:
			position, tokenIndex, depth = position805, tokenIndex805, depth805
			return false
		},
		/* 67 _ <- <((&('/') COMMENT_BLOCK) | (&('-') COMMENT_TRAIL) | (&('\t' | '\n' | ' ') SPACE))*> */
		func() bool {
			{
				position808 := position
				depth++
			l809:
				{
					position810, tokenIndex810, depth810 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '/':
							{
								position812 := position
								depth++
								if buffer[position] != rune('/') {
									goto l810
								}
								position++
								if buffer[position] != rune('*') {
									goto l810
								}
								position++
							l813:
								{
									position814, tokenIndex814, depth814 := position, tokenIndex, depth
									{
										position815, tokenIndex815, depth815 := position, tokenIndex, depth
										if buffer[position] != rune('*') {
											goto l815
										}
										position++
										if buffer[position] != rune('/') {
											goto l815
										}
										position++
										goto l814
									l815:
										position, tokenIndex, depth = position815, tokenIndex815, depth815
									}
									if !matchDot() {
										goto l814
									}
									goto l813
								l814:
									position, tokenIndex, depth = position814, tokenIndex814, depth814
								}
								if buffer[position] != rune('*') {
									goto l810
								}
								position++
								if buffer[position] != rune('/') {
									goto l810
								}
								position++
								depth--
								add(ruleCOMMENT_BLOCK, position812)
							}
							break
						case '-':
							{
								position816 := position
								depth++
								if buffer[position] != rune('-') {
									goto l810
								}
								position++
								if buffer[position] != rune('-') {
									goto l810
								}
								position++
							l817:
								{
									position818, tokenIndex818, depth818 := position, tokenIndex, depth
									{
										position819, tokenIndex819, depth819 := position, tokenIndex, depth
										if buffer[position] != rune('\n') {
											goto l819
										}
										position++
										goto l818
									l819:
										position, tokenIndex, depth = position819, tokenIndex819, depth819
									}
									if !matchDot() {
										goto l818
									}
									goto l817
								l818:
									position, tokenIndex, depth = position818, tokenIndex818, depth818
								}
								depth--
								add(ruleCOMMENT_TRAIL, position816)
							}
							break
						default:
							{
								position820 := position
								depth++
								{
									switch buffer[position] {
									case '\t':
										if buffer[position] != rune('\t') {
											goto l810
										}
										position++
										break
									case '\n':
										if buffer[position] != rune('\n') {
											goto l810
										}
										position++
										break
									default:
										if buffer[position] != rune(' ') {
											goto l810
										}
										position++
										break
									}
								}

								depth--
								add(ruleSPACE, position820)
							}
							break
						}
					}

					goto l809
				l810:
					position, tokenIndex, depth = position810, tokenIndex810, depth810
				}
				depth--
				add(rule_, position808)
			}
			return true
		},
		/* 68 COMMENT_TRAIL <- <('-' '-' (!'\n' .)*)> */
		nil,
		/* 69 COMMENT_BLOCK <- <('/' '*' (!('*' '/') .)* ('*' '/'))> */
		nil,
		/* 70 KEY <- <!ID_CONT> */
		func() bool {
			position824, tokenIndex824, depth824 := position, tokenIndex, depth
			{
				position825 := position
				depth++
				{
					position826, tokenIndex826, depth826 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l826
					}
					goto l824
				l826:
					position, tokenIndex, depth = position826, tokenIndex826, depth826
				}
				depth--
				add(ruleKEY, position825)
			}
			return true
		l824:
			position, tokenIndex, depth = position824, tokenIndex824, depth824
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
