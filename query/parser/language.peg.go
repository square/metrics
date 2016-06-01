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
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
	ruleAction9
	rulePegText
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
	ruleAction54
	ruleAction55
	ruleAction56
	ruleAction57
	ruleAction58
	ruleAction59
	ruleAction60
	ruleAction61
	ruleAction62
	ruleAction63
	ruleAction64
	ruleAction65
	ruleAction66
	ruleAction67
	ruleAction68
	ruleAction69
	ruleAction70
	ruleAction71
	ruleAction72
	ruleAction73
	ruleAction74
	ruleAction75
	ruleAction76
	ruleAction77
	ruleAction78
	ruleAction79
	ruleAction80
	ruleAction81
	ruleAction82
	ruleAction83
	ruleAction84
	ruleAction85
	ruleAction86
	ruleAction87
	ruleAction88
	ruleAction89
	ruleAction90
	ruleAction91
	ruleAction92
	ruleAction93
	ruleAction94
	ruleAction95
	ruleAction96
	ruleAction97

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
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",
	"PegText",
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
	"Action54",
	"Action55",
	"Action56",
	"Action57",
	"Action58",
	"Action59",
	"Action60",
	"Action61",
	"Action62",
	"Action63",
	"Action64",
	"Action65",
	"Action66",
	"Action67",
	"Action68",
	"Action69",
	"Action70",
	"Action71",
	"Action72",
	"Action73",
	"Action74",
	"Action75",
	"Action76",
	"Action77",
	"Action78",
	"Action79",
	"Action80",
	"Action81",
	"Action82",
	"Action83",
	"Action84",
	"Action85",
	"Action86",
	"Action87",
	"Action88",
	"Action89",
	"Action90",
	"Action91",
	"Action92",
	"Action93",
	"Action94",
	"Action95",
	"Action96",
	"Action97",

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

	// programming errors accumulated during the AST traversal.
	// a non-empty list at the finish time implies a programming error.

	// final result
	command command.Command

	Buffer string
	buffer []rune
	rules  [172]func() bool
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
			p.errorHere(token.begin, `expected string literal to follow keyword "match"`)
		case ruleAction4:
			p.addMatchClause()
		case ruleAction5:
			p.errorHere(token.begin, `expected "where" to follow keyword "metrics" in "describe metrics" command`)
		case ruleAction6:
			p.errorHere(token.begin, `expected tag key to follow keyword "where" in "describe metrics" command`)
		case ruleAction7:
			p.errorHere(token.begin, `expected "=" to follow keyword "where" in "describe metrics" command`)
		case ruleAction8:
			p.errorHere(token.begin, `expected string literal to follow "=" in "describe metrics" command`)
		case ruleAction9:
			p.makeDescribeMetrics()
		case ruleAction10:
			p.pushString(unescapeLiteral(buffer[begin:end]))
		case ruleAction11:
			p.errorHere(token.begin, `expected metric name to follow "describe" in "describe" command`)
		case ruleAction12:
			p.makeDescribe()
		case ruleAction13:
			p.addEvaluationContext()
		case ruleAction14:
			p.addPropertyKey(buffer[begin:end])
		case ruleAction15:
			p.addPropertyValue(buffer[begin:end])
		case ruleAction16:
			p.errorHere(token.begin, `expected property value to follow property key`)
		case ruleAction17:
			p.insertPropertyKeyValue()
		case ruleAction18:
			p.checkPropertyClause()
		case ruleAction19:
			p.addNullPredicate()
		case ruleAction20:
			p.addExpressionList()
		case ruleAction21:
			p.appendExpression()
		case ruleAction22:
			p.errorHere(token.begin, `expected expression to follow ","`)
		case ruleAction23:
			p.appendExpression()
		case ruleAction24:
			p.addOperatorLiteral("+")
		case ruleAction25:
			p.addOperatorLiteral("-")
		case ruleAction26:
			p.errorHere(token.begin, `expected expression to follow operator "+" or "-"`)
		case ruleAction27:
			p.addOperatorFunction()
		case ruleAction28:
			p.addOperatorLiteral("/")
		case ruleAction29:
			p.addOperatorLiteral("*")
		case ruleAction30:
			p.errorHere(token.begin, `expected expression to follow operator "*" or "/"`)
		case ruleAction31:
			p.addOperatorFunction()
		case ruleAction32:
			p.errorHere(token.begin, `expected function name to follow pipe "|"`)
		case ruleAction33:
			p.pushString(unescapeLiteral(buffer[begin:end]))
		case ruleAction34:
			p.addExpressionList()
		case ruleAction35:
			p.errorHere(token.begin, `expected ")" to close "(" opened in pipe function call`)
		case ruleAction36:

			p.addExpressionList()
			p.addGroupBy()

		case ruleAction37:
			p.addPipeExpression()
		case ruleAction38:
			p.errorHere(token.begin, `expected expression to follow "("`)
		case ruleAction39:
			p.errorHere(token.begin, `expected ")" to close "("`)
		case ruleAction40:
			p.addDurationNode(text)
		case ruleAction41:
			p.addNumberNode(buffer[begin:end])
		case ruleAction42:
			p.addStringNode(unescapeLiteral(buffer[begin:end]))
		case ruleAction43:
			p.errorHere(token.begin, `expected "
		case ruleAction44:
			" opened for annotation`)
		case ruleAction45:
			p.addAnnotationExpression(buffer[begin:end])
		case ruleAction46:
			p.addGroupBy()
		case ruleAction47:
			p.pushString(unescapeLiteral(buffer[begin:end]))
		case ruleAction48:
			p.errorHere(token.begin, `expected ")" to close "(" opened by function call`)
		case ruleAction49:
			p.addFunctionInvocation()
		case ruleAction50:
			p.pushString(unescapeLiteral(buffer[begin:end]))
		case ruleAction51:
			p.errorHere(token.begin, `expected predicate to follow "[" after metric`)
		case ruleAction52:
			p.errorHere(token.begin, `expected "]" to close "[" opened to apply predicate`)
		case ruleAction53:
			p.addNullPredicate()
		case ruleAction54:
			p.addMetricExpression()
		case ruleAction55:
			p.errorHere(token.begin, `expected keyword "by" to follow keyword "group" in "group by" clause`)
		case ruleAction56:
			p.errorHere(token.begin, `expected tag key identifier to follow "group by" keywords in "group by" clause`)
		case ruleAction57:
			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		case ruleAction58:
			p.errorHere(token.begin, `expected tag key identifier to follow "," in "group by" clause`)
		case ruleAction59:
			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		case ruleAction60:
			p.errorHere(token.begin, `expected keyword "by" to follow keyword "collapse" in "collapse by" clause`)
		case ruleAction61:
			p.errorHere(token.begin, `expected tag key identifier to follow "collapse by" keywords in "collapse by" clause`)
		case ruleAction62:
			p.appendCollapseBy(unescapeLiteral(text))
		case ruleAction63:
			p.errorHere(token.begin, `expected tag key identifier to follow "," in "collapse by" clause`)
		case ruleAction64:
			p.appendCollapseBy(unescapeLiteral(text))
		case ruleAction65:
			p.errorHere(token.begin, `expected predicate to follow "where" keyword`)
		case ruleAction66:
			p.errorHere(token.begin, `expected predicate to follow "or" operator`)
		case ruleAction67:
			p.addOrPredicate()
		case ruleAction68:
			p.errorHere(token.begin, `expected predicate to follow "and" operator`)
		case ruleAction69:
			p.addAndPredicate()
		case ruleAction70:
			p.errorHere(token.begin, `expected predicate to follow "not" operator`)
		case ruleAction71:
			p.addNotPredicate()
		case ruleAction72:
			p.errorHere(token.begin, `expected predicate to follow "("`)
		case ruleAction73:
			p.errorHere(token.begin, `expected ")" to close "(" opened in predicate`)
		case ruleAction74:
			p.errorHere(token.begin, `expected string literal to follow "="`)
		case ruleAction75:
			p.addLiteralMatcher()
		case ruleAction76:
			p.errorHere(token.begin, `expected string literal to follow "!="`)
		case ruleAction77:
			p.addLiteralMatcher()
		case ruleAction78:
			p.addNotPredicate()
		case ruleAction79:
			p.errorHere(token.begin, `expected regex string literal to follow "match"`)
		case ruleAction80:
			p.addRegexMatcher()
		case ruleAction81:
			p.errorHere(token.begin, `expected string literal list to follow "in" keyword`)
		case ruleAction82:
			p.addListMatcher()
		case ruleAction83:
			p.errorHere(token.begin, `expected "=", "!=", "match", or "in" to follow tag key in predicate`)
		case ruleAction84:
			p.pushString(unescapeLiteral(buffer[begin:end]))
		case ruleAction85:
			p.addLiteralList()
		case ruleAction86:
			p.errorHere(token.begin, `expected string literal to follow "(" in literal list`)
		case ruleAction87:
			p.errorHere(token.begin, `expected string literal to follow "," in literal list`)
		case ruleAction88:
			p.errorHere(token.begin, `expected ")" to close "(" for literal list`)
		case ruleAction89:
			p.appendLiteral(unescapeLiteral(buffer[begin:end]))
		case ruleAction90:
			p.addTagLiteral(unescapeLiteral(buffer[begin:end]))
		case ruleAction91:
			p.errorHere(token.begin, "expected \"`\" to end identifier")
		case ruleAction92:
			p.errorHere(token.begin, `expected identifier segment to follow "."`)
		case ruleAction93:
			p.errorHere(token.begin, `expected keyword "by" to follow keyword "sample"`)
		case ruleAction94:
			p.errorHere(token.begin, `expected "'" to close string`)
		case ruleAction95:
			p.errorHere(token.begin, `expected '"' to close string`)
		case ruleAction96:
			p.errorHere(token.begin, "expected \"\\\", \"'\", \"`\", or '\"' to follow \"\\\" in string literal")
		case ruleAction97:
			p.errorHere(token.begin, `expected exponent`)

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
						if !_rules[ruleoptionalPredicateClause]() {
							goto l3
						}
						{
							position19 := position
							depth++
							{
								add(ruleAction13, position)
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
									add(ruleAction14, position)
								}
								{
									position24, tokenIndex24, depth24 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l25
									}
									{
										position26 := position
										depth++
										{
											position27 := position
											depth++
											{
												position28, tokenIndex28, depth28 := position, tokenIndex, depth
												if !_rules[rule_]() {
													goto l29
												}
												{
													position30 := position
													depth++
													if !_rules[ruleNUMBER]() {
														goto l29
													}
												l31:
													{
														position32, tokenIndex32, depth32 := position, tokenIndex, depth
														{
															position33, tokenIndex33, depth33 := position, tokenIndex, depth
															if c := buffer[position]; c < rune('a') || c > rune('z') {
																goto l34
															}
															position++
															goto l33
														l34:
															position, tokenIndex, depth = position33, tokenIndex33, depth33
															if c := buffer[position]; c < rune('A') || c > rune('Z') {
																goto l32
															}
															position++
														}
													l33:
														goto l31
													l32:
														position, tokenIndex, depth = position32, tokenIndex32, depth32
													}
													depth--
													add(rulePegText, position30)
												}
												goto l28
											l29:
												position, tokenIndex, depth = position28, tokenIndex28, depth28
												if !_rules[rule_]() {
													goto l35
												}
												if !_rules[ruleSTRING]() {
													goto l35
												}
												goto l28
											l35:
												position, tokenIndex, depth = position28, tokenIndex28, depth28
												if !_rules[rule_]() {
													goto l25
												}
												{
													position36 := position
													depth++
													{
														position37, tokenIndex37, depth37 := position, tokenIndex, depth
														if buffer[position] != rune('n') {
															goto l38
														}
														position++
														goto l37
													l38:
														position, tokenIndex, depth = position37, tokenIndex37, depth37
														if buffer[position] != rune('N') {
															goto l25
														}
														position++
													}
												l37:
													{
														position39, tokenIndex39, depth39 := position, tokenIndex, depth
														if buffer[position] != rune('o') {
															goto l40
														}
														position++
														goto l39
													l40:
														position, tokenIndex, depth = position39, tokenIndex39, depth39
														if buffer[position] != rune('O') {
															goto l25
														}
														position++
													}
												l39:
													{
														position41, tokenIndex41, depth41 := position, tokenIndex, depth
														if buffer[position] != rune('w') {
															goto l42
														}
														position++
														goto l41
													l42:
														position, tokenIndex, depth = position41, tokenIndex41, depth41
														if buffer[position] != rune('W') {
															goto l25
														}
														position++
													}
												l41:
													depth--
													add(rulePegText, position36)
												}
												if !_rules[ruleKEY]() {
													goto l25
												}
											}
										l28:
											depth--
											add(ruleTIMESTAMP, position27)
										}
										depth--
										add(rulePROPERTY_VALUE, position26)
									}
									{
										add(ruleAction15, position)
									}
									goto l24
								l25:
									position, tokenIndex, depth = position24, tokenIndex24, depth24
									{
										add(ruleAction16, position)
									}
								}
							l24:
								{
									add(ruleAction17, position)
								}
								goto l21
							l22:
								position, tokenIndex, depth = position22, tokenIndex22, depth22
							}
							{
								add(ruleAction18, position)
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
						position48 := position
						depth++
						if !_rules[rule_]() {
							goto l0
						}
						{
							position49, tokenIndex49, depth49 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l50
							}
							position++
							goto l49
						l50:
							position, tokenIndex, depth = position49, tokenIndex49, depth49
							if buffer[position] != rune('D') {
								goto l0
							}
							position++
						}
					l49:
						{
							position51, tokenIndex51, depth51 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l52
							}
							position++
							goto l51
						l52:
							position, tokenIndex, depth = position51, tokenIndex51, depth51
							if buffer[position] != rune('E') {
								goto l0
							}
							position++
						}
					l51:
						{
							position53, tokenIndex53, depth53 := position, tokenIndex, depth
							if buffer[position] != rune('s') {
								goto l54
							}
							position++
							goto l53
						l54:
							position, tokenIndex, depth = position53, tokenIndex53, depth53
							if buffer[position] != rune('S') {
								goto l0
							}
							position++
						}
					l53:
						{
							position55, tokenIndex55, depth55 := position, tokenIndex, depth
							if buffer[position] != rune('c') {
								goto l56
							}
							position++
							goto l55
						l56:
							position, tokenIndex, depth = position55, tokenIndex55, depth55
							if buffer[position] != rune('C') {
								goto l0
							}
							position++
						}
					l55:
						{
							position57, tokenIndex57, depth57 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l58
							}
							position++
							goto l57
						l58:
							position, tokenIndex, depth = position57, tokenIndex57, depth57
							if buffer[position] != rune('R') {
								goto l0
							}
							position++
						}
					l57:
						{
							position59, tokenIndex59, depth59 := position, tokenIndex, depth
							if buffer[position] != rune('i') {
								goto l60
							}
							position++
							goto l59
						l60:
							position, tokenIndex, depth = position59, tokenIndex59, depth59
							if buffer[position] != rune('I') {
								goto l0
							}
							position++
						}
					l59:
						{
							position61, tokenIndex61, depth61 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l62
							}
							position++
							goto l61
						l62:
							position, tokenIndex, depth = position61, tokenIndex61, depth61
							if buffer[position] != rune('B') {
								goto l0
							}
							position++
						}
					l61:
						{
							position63, tokenIndex63, depth63 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l64
							}
							position++
							goto l63
						l64:
							position, tokenIndex, depth = position63, tokenIndex63, depth63
							if buffer[position] != rune('E') {
								goto l0
							}
							position++
						}
					l63:
						if !_rules[ruleKEY]() {
							goto l0
						}
						{
							position65, tokenIndex65, depth65 := position, tokenIndex, depth
							{
								position67 := position
								depth++
								if !_rules[rule_]() {
									goto l66
								}
								{
									position68, tokenIndex68, depth68 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l69
									}
									position++
									goto l68
								l69:
									position, tokenIndex, depth = position68, tokenIndex68, depth68
									if buffer[position] != rune('A') {
										goto l66
									}
									position++
								}
							l68:
								{
									position70, tokenIndex70, depth70 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l71
									}
									position++
									goto l70
								l71:
									position, tokenIndex, depth = position70, tokenIndex70, depth70
									if buffer[position] != rune('L') {
										goto l66
									}
									position++
								}
							l70:
								{
									position72, tokenIndex72, depth72 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l73
									}
									position++
									goto l72
								l73:
									position, tokenIndex, depth = position72, tokenIndex72, depth72
									if buffer[position] != rune('L') {
										goto l66
									}
									position++
								}
							l72:
								if !_rules[ruleKEY]() {
									goto l66
								}
								{
									position74 := position
									depth++
									{
										position75, tokenIndex75, depth75 := position, tokenIndex, depth
										{
											position77 := position
											depth++
											if !_rules[rule_]() {
												goto l76
											}
											{
												position78, tokenIndex78, depth78 := position, tokenIndex, depth
												if buffer[position] != rune('m') {
													goto l79
												}
												position++
												goto l78
											l79:
												position, tokenIndex, depth = position78, tokenIndex78, depth78
												if buffer[position] != rune('M') {
													goto l76
												}
												position++
											}
										l78:
											{
												position80, tokenIndex80, depth80 := position, tokenIndex, depth
												if buffer[position] != rune('a') {
													goto l81
												}
												position++
												goto l80
											l81:
												position, tokenIndex, depth = position80, tokenIndex80, depth80
												if buffer[position] != rune('A') {
													goto l76
												}
												position++
											}
										l80:
											{
												position82, tokenIndex82, depth82 := position, tokenIndex, depth
												if buffer[position] != rune('t') {
													goto l83
												}
												position++
												goto l82
											l83:
												position, tokenIndex, depth = position82, tokenIndex82, depth82
												if buffer[position] != rune('T') {
													goto l76
												}
												position++
											}
										l82:
											{
												position84, tokenIndex84, depth84 := position, tokenIndex, depth
												if buffer[position] != rune('c') {
													goto l85
												}
												position++
												goto l84
											l85:
												position, tokenIndex, depth = position84, tokenIndex84, depth84
												if buffer[position] != rune('C') {
													goto l76
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
													goto l76
												}
												position++
											}
										l86:
											if !_rules[ruleKEY]() {
												goto l76
											}
											{
												position88, tokenIndex88, depth88 := position, tokenIndex, depth
												if !_rules[ruleliteralString]() {
													goto l89
												}
												goto l88
											l89:
												position, tokenIndex, depth = position88, tokenIndex88, depth88
												{
													add(ruleAction3, position)
												}
											}
										l88:
											{
												add(ruleAction4, position)
											}
											depth--
											add(rulematchClause, position77)
										}
										goto l75
									l76:
										position, tokenIndex, depth = position75, tokenIndex75, depth75
										{
											add(ruleAction2, position)
										}
									}
								l75:
									depth--
									add(ruleoptionalMatchClause, position74)
								}
								{
									add(ruleAction1, position)
								}
								depth--
								add(ruledescribeAllStmt, position67)
							}
							goto l65
						l66:
							position, tokenIndex, depth = position65, tokenIndex65, depth65
							{
								position95 := position
								depth++
								if !_rules[rule_]() {
									goto l94
								}
								{
									position96, tokenIndex96, depth96 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l97
									}
									position++
									goto l96
								l97:
									position, tokenIndex, depth = position96, tokenIndex96, depth96
									if buffer[position] != rune('M') {
										goto l94
									}
									position++
								}
							l96:
								{
									position98, tokenIndex98, depth98 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l99
									}
									position++
									goto l98
								l99:
									position, tokenIndex, depth = position98, tokenIndex98, depth98
									if buffer[position] != rune('E') {
										goto l94
									}
									position++
								}
							l98:
								{
									position100, tokenIndex100, depth100 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l101
									}
									position++
									goto l100
								l101:
									position, tokenIndex, depth = position100, tokenIndex100, depth100
									if buffer[position] != rune('T') {
										goto l94
									}
									position++
								}
							l100:
								{
									position102, tokenIndex102, depth102 := position, tokenIndex, depth
									if buffer[position] != rune('r') {
										goto l103
									}
									position++
									goto l102
								l103:
									position, tokenIndex, depth = position102, tokenIndex102, depth102
									if buffer[position] != rune('R') {
										goto l94
									}
									position++
								}
							l102:
								{
									position104, tokenIndex104, depth104 := position, tokenIndex, depth
									if buffer[position] != rune('i') {
										goto l105
									}
									position++
									goto l104
								l105:
									position, tokenIndex, depth = position104, tokenIndex104, depth104
									if buffer[position] != rune('I') {
										goto l94
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
										goto l94
									}
									position++
								}
							l106:
								{
									position108, tokenIndex108, depth108 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l109
									}
									position++
									goto l108
								l109:
									position, tokenIndex, depth = position108, tokenIndex108, depth108
									if buffer[position] != rune('S') {
										goto l94
									}
									position++
								}
							l108:
								if !_rules[ruleKEY]() {
									goto l94
								}
								{
									position110, tokenIndex110, depth110 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l111
									}
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
											goto l111
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
											goto l111
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
											goto l111
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
											goto l111
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
											goto l111
										}
										position++
									}
								l120:
									if !_rules[ruleKEY]() {
										goto l111
									}
									goto l110
								l111:
									position, tokenIndex, depth = position110, tokenIndex110, depth110
									{
										add(ruleAction5, position)
									}
								}
							l110:
								{
									position123, tokenIndex123, depth123 := position, tokenIndex, depth
									if !_rules[ruletagName]() {
										goto l124
									}
									goto l123
								l124:
									position, tokenIndex, depth = position123, tokenIndex123, depth123
									{
										add(ruleAction6, position)
									}
								}
							l123:
								{
									position126, tokenIndex126, depth126 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l127
									}
									if buffer[position] != rune('=') {
										goto l127
									}
									position++
									goto l126
								l127:
									position, tokenIndex, depth = position126, tokenIndex126, depth126
									{
										add(ruleAction7, position)
									}
								}
							l126:
								{
									position129, tokenIndex129, depth129 := position, tokenIndex, depth
									if !_rules[ruleliteralString]() {
										goto l130
									}
									goto l129
								l130:
									position, tokenIndex, depth = position129, tokenIndex129, depth129
									{
										add(ruleAction8, position)
									}
								}
							l129:
								{
									add(ruleAction9, position)
								}
								depth--
								add(ruledescribeMetrics, position95)
							}
							goto l65
						l94:
							position, tokenIndex, depth = position65, tokenIndex65, depth65
							{
								position133 := position
								depth++
								{
									position134, tokenIndex134, depth134 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l135
									}
									{
										position136 := position
										depth++
										{
											position137 := position
											depth++
											if !_rules[ruleIDENTIFIER]() {
												goto l135
											}
											depth--
											add(ruleMETRIC_NAME, position137)
										}
										depth--
										add(rulePegText, position136)
									}
									{
										add(ruleAction10, position)
									}
									goto l134
								l135:
									position, tokenIndex, depth = position134, tokenIndex134, depth134
									{
										add(ruleAction11, position)
									}
								}
							l134:
								if !_rules[ruleoptionalPredicateClause]() {
									goto l0
								}
								{
									add(ruleAction12, position)
								}
								depth--
								add(ruledescribeSingleStmt, position133)
							}
						}
					l65:
						depth--
						add(ruledescribeStmt, position48)
					}
				}
			l2:
				if !_rules[rule_]() {
					goto l0
				}
				{
					position141, tokenIndex141, depth141 := position, tokenIndex, depth
					if !matchDot() {
						goto l141
					}
					goto l0
				l141:
					position, tokenIndex, depth = position141, tokenIndex141, depth141
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
		/* 5 matchClause <- <(_ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H')) KEY (literalString / Action3) Action4)> */
		nil,
		/* 6 describeMetrics <- <(_ (('m' / 'M') ('e' / 'E') ('t' / 'T') ('r' / 'R') ('i' / 'I') ('c' / 'C') ('s' / 'S')) KEY ((_ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY) / Action5) (tagName / Action6) ((_ '=') / Action7) (literalString / Action8) Action9)> */
		nil,
		/* 7 describeSingleStmt <- <(((_ <METRIC_NAME> Action10) / Action11) optionalPredicateClause Action12)> */
		nil,
		/* 8 propertyClause <- <(Action13 (_ PROPERTY_KEY Action14 ((_ PROPERTY_VALUE Action15) / Action16) Action17)* Action18)> */
		nil,
		/* 9 optionalPredicateClause <- <(predicateClause / Action19)> */
		func() bool {
			{
				position151 := position
				depth++
				{
					position152, tokenIndex152, depth152 := position, tokenIndex, depth
					{
						position154 := position
						depth++
						if !_rules[rule_]() {
							goto l153
						}
						{
							position155, tokenIndex155, depth155 := position, tokenIndex, depth
							if buffer[position] != rune('w') {
								goto l156
							}
							position++
							goto l155
						l156:
							position, tokenIndex, depth = position155, tokenIndex155, depth155
							if buffer[position] != rune('W') {
								goto l153
							}
							position++
						}
					l155:
						{
							position157, tokenIndex157, depth157 := position, tokenIndex, depth
							if buffer[position] != rune('h') {
								goto l158
							}
							position++
							goto l157
						l158:
							position, tokenIndex, depth = position157, tokenIndex157, depth157
							if buffer[position] != rune('H') {
								goto l153
							}
							position++
						}
					l157:
						{
							position159, tokenIndex159, depth159 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l160
							}
							position++
							goto l159
						l160:
							position, tokenIndex, depth = position159, tokenIndex159, depth159
							if buffer[position] != rune('E') {
								goto l153
							}
							position++
						}
					l159:
						{
							position161, tokenIndex161, depth161 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l162
							}
							position++
							goto l161
						l162:
							position, tokenIndex, depth = position161, tokenIndex161, depth161
							if buffer[position] != rune('R') {
								goto l153
							}
							position++
						}
					l161:
						{
							position163, tokenIndex163, depth163 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l164
							}
							position++
							goto l163
						l164:
							position, tokenIndex, depth = position163, tokenIndex163, depth163
							if buffer[position] != rune('E') {
								goto l153
							}
							position++
						}
					l163:
						if !_rules[ruleKEY]() {
							goto l153
						}
						{
							position165, tokenIndex165, depth165 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l166
							}
							if !_rules[rulepredicate_1]() {
								goto l166
							}
							goto l165
						l166:
							position, tokenIndex, depth = position165, tokenIndex165, depth165
							{
								add(ruleAction65, position)
							}
						}
					l165:
						depth--
						add(rulepredicateClause, position154)
					}
					goto l152
				l153:
					position, tokenIndex, depth = position152, tokenIndex152, depth152
					{
						add(ruleAction19, position)
					}
				}
			l152:
				depth--
				add(ruleoptionalPredicateClause, position151)
			}
			return true
		},
		/* 10 expressionList <- <(Action20 expression_start Action21 (_ COMMA (expression_start / Action22) Action23)*)> */
		func() bool {
			position169, tokenIndex169, depth169 := position, tokenIndex, depth
			{
				position170 := position
				depth++
				{
					add(ruleAction20, position)
				}
				if !_rules[ruleexpression_start]() {
					goto l169
				}
				{
					add(ruleAction21, position)
				}
			l173:
				{
					position174, tokenIndex174, depth174 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l174
					}
					if !_rules[ruleCOMMA]() {
						goto l174
					}
					{
						position175, tokenIndex175, depth175 := position, tokenIndex, depth
						if !_rules[ruleexpression_start]() {
							goto l176
						}
						goto l175
					l176:
						position, tokenIndex, depth = position175, tokenIndex175, depth175
						{
							add(ruleAction22, position)
						}
					}
				l175:
					{
						add(ruleAction23, position)
					}
					goto l173
				l174:
					position, tokenIndex, depth = position174, tokenIndex174, depth174
				}
				depth--
				add(ruleexpressionList, position170)
			}
			return true
		l169:
			position, tokenIndex, depth = position169, tokenIndex169, depth169
			return false
		},
		/* 11 expression_start <- <(expression_sum add_pipe)> */
		func() bool {
			position179, tokenIndex179, depth179 := position, tokenIndex, depth
			{
				position180 := position
				depth++
				{
					position181 := position
					depth++
					if !_rules[ruleexpression_product]() {
						goto l179
					}
				l182:
					{
						position183, tokenIndex183, depth183 := position, tokenIndex, depth
						if !_rules[ruleadd_pipe]() {
							goto l183
						}
						{
							position184, tokenIndex184, depth184 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l185
							}
							{
								position186 := position
								depth++
								if buffer[position] != rune('+') {
									goto l185
								}
								position++
								depth--
								add(ruleOP_ADD, position186)
							}
							{
								add(ruleAction24, position)
							}
							goto l184
						l185:
							position, tokenIndex, depth = position184, tokenIndex184, depth184
							if !_rules[rule_]() {
								goto l183
							}
							{
								position188 := position
								depth++
								if buffer[position] != rune('-') {
									goto l183
								}
								position++
								depth--
								add(ruleOP_SUB, position188)
							}
							{
								add(ruleAction25, position)
							}
						}
					l184:
						{
							position190, tokenIndex190, depth190 := position, tokenIndex, depth
							if !_rules[ruleexpression_product]() {
								goto l191
							}
							goto l190
						l191:
							position, tokenIndex, depth = position190, tokenIndex190, depth190
							{
								add(ruleAction26, position)
							}
						}
					l190:
						{
							add(ruleAction27, position)
						}
						goto l182
					l183:
						position, tokenIndex, depth = position183, tokenIndex183, depth183
					}
					depth--
					add(ruleexpression_sum, position181)
				}
				if !_rules[ruleadd_pipe]() {
					goto l179
				}
				depth--
				add(ruleexpression_start, position180)
			}
			return true
		l179:
			position, tokenIndex, depth = position179, tokenIndex179, depth179
			return false
		},
		/* 12 expression_sum <- <(expression_product (add_pipe ((_ OP_ADD Action24) / (_ OP_SUB Action25)) (expression_product / Action26) Action27)*)> */
		nil,
		/* 13 expression_product <- <(expression_atom (add_pipe ((_ OP_DIV Action28) / (_ OP_MULT Action29)) (expression_atom / Action30) Action31)*)> */
		func() bool {
			position195, tokenIndex195, depth195 := position, tokenIndex, depth
			{
				position196 := position
				depth++
				if !_rules[ruleexpression_atom]() {
					goto l195
				}
			l197:
				{
					position198, tokenIndex198, depth198 := position, tokenIndex, depth
					if !_rules[ruleadd_pipe]() {
						goto l198
					}
					{
						position199, tokenIndex199, depth199 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l200
						}
						{
							position201 := position
							depth++
							if buffer[position] != rune('/') {
								goto l200
							}
							position++
							depth--
							add(ruleOP_DIV, position201)
						}
						{
							add(ruleAction28, position)
						}
						goto l199
					l200:
						position, tokenIndex, depth = position199, tokenIndex199, depth199
						if !_rules[rule_]() {
							goto l198
						}
						{
							position203 := position
							depth++
							if buffer[position] != rune('*') {
								goto l198
							}
							position++
							depth--
							add(ruleOP_MULT, position203)
						}
						{
							add(ruleAction29, position)
						}
					}
				l199:
					{
						position205, tokenIndex205, depth205 := position, tokenIndex, depth
						if !_rules[ruleexpression_atom]() {
							goto l206
						}
						goto l205
					l206:
						position, tokenIndex, depth = position205, tokenIndex205, depth205
						{
							add(ruleAction30, position)
						}
					}
				l205:
					{
						add(ruleAction31, position)
					}
					goto l197
				l198:
					position, tokenIndex, depth = position198, tokenIndex198, depth198
				}
				depth--
				add(ruleexpression_product, position196)
			}
			return true
		l195:
			position, tokenIndex, depth = position195, tokenIndex195, depth195
			return false
		},
		/* 14 add_one_pipe <- <(_ OP_PIPE ((_ <IDENTIFIER>) / Action32) Action33 ((_ PAREN_OPEN (expressionList / Action34) optionalGroupBy ((_ PAREN_CLOSE) / Action35)) / Action36) Action37 expression_annotation)> */
		nil,
		/* 15 add_pipe <- <add_one_pipe*> */
		func() bool {
			{
				position211 := position
				depth++
			l212:
				{
					position213, tokenIndex213, depth213 := position, tokenIndex, depth
					{
						position214 := position
						depth++
						if !_rules[rule_]() {
							goto l213
						}
						{
							position215 := position
							depth++
							if buffer[position] != rune('|') {
								goto l213
							}
							position++
							depth--
							add(ruleOP_PIPE, position215)
						}
						{
							position216, tokenIndex216, depth216 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l217
							}
							{
								position218 := position
								depth++
								if !_rules[ruleIDENTIFIER]() {
									goto l217
								}
								depth--
								add(rulePegText, position218)
							}
							goto l216
						l217:
							position, tokenIndex, depth = position216, tokenIndex216, depth216
							{
								add(ruleAction32, position)
							}
						}
					l216:
						{
							add(ruleAction33, position)
						}
						{
							position221, tokenIndex221, depth221 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l222
							}
							if !_rules[rulePAREN_OPEN]() {
								goto l222
							}
							{
								position223, tokenIndex223, depth223 := position, tokenIndex, depth
								if !_rules[ruleexpressionList]() {
									goto l224
								}
								goto l223
							l224:
								position, tokenIndex, depth = position223, tokenIndex223, depth223
								{
									add(ruleAction34, position)
								}
							}
						l223:
							if !_rules[ruleoptionalGroupBy]() {
								goto l222
							}
							{
								position226, tokenIndex226, depth226 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l227
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l227
								}
								goto l226
							l227:
								position, tokenIndex, depth = position226, tokenIndex226, depth226
								{
									add(ruleAction35, position)
								}
							}
						l226:
							goto l221
						l222:
							position, tokenIndex, depth = position221, tokenIndex221, depth221
							{
								add(ruleAction36, position)
							}
						}
					l221:
						{
							add(ruleAction37, position)
						}
						if !_rules[ruleexpression_annotation]() {
							goto l213
						}
						depth--
						add(ruleadd_one_pipe, position214)
					}
					goto l212
				l213:
					position, tokenIndex, depth = position213, tokenIndex213, depth213
				}
				depth--
				add(ruleadd_pipe, position211)
			}
			return true
		},
		/* 16 expression_atom <- <(expression_atom_raw expression_annotation)> */
		func() bool {
			position231, tokenIndex231, depth231 := position, tokenIndex, depth
			{
				position232 := position
				depth++
				{
					position233 := position
					depth++
					{
						position234, tokenIndex234, depth234 := position, tokenIndex, depth
						{
							position236 := position
							depth++
							if !_rules[rule_]() {
								goto l235
							}
							{
								position237 := position
								depth++
								if !_rules[ruleIDENTIFIER]() {
									goto l235
								}
								depth--
								add(rulePegText, position237)
							}
							{
								add(ruleAction47, position)
							}
							if !_rules[rule_]() {
								goto l235
							}
							if !_rules[rulePAREN_OPEN]() {
								goto l235
							}
							if !_rules[ruleexpressionList]() {
								goto l235
							}
							if !_rules[ruleoptionalGroupBy]() {
								goto l235
							}
							{
								position239, tokenIndex239, depth239 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l240
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l240
								}
								goto l239
							l240:
								position, tokenIndex, depth = position239, tokenIndex239, depth239
								{
									add(ruleAction48, position)
								}
							}
						l239:
							{
								add(ruleAction49, position)
							}
							depth--
							add(ruleexpression_function, position236)
						}
						goto l234
					l235:
						position, tokenIndex, depth = position234, tokenIndex234, depth234
						{
							position244 := position
							depth++
							if !_rules[rule_]() {
								goto l243
							}
							{
								position245 := position
								depth++
								if !_rules[ruleIDENTIFIER]() {
									goto l243
								}
								depth--
								add(rulePegText, position245)
							}
							{
								add(ruleAction50, position)
							}
							{
								position247, tokenIndex247, depth247 := position, tokenIndex, depth
								{
									position249, tokenIndex249, depth249 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l250
									}
									if buffer[position] != rune('[') {
										goto l250
									}
									position++
									{
										position251, tokenIndex251, depth251 := position, tokenIndex, depth
										if !_rules[rulepredicate_1]() {
											goto l252
										}
										goto l251
									l252:
										position, tokenIndex, depth = position251, tokenIndex251, depth251
										{
											add(ruleAction51, position)
										}
									}
								l251:
									{
										position254, tokenIndex254, depth254 := position, tokenIndex, depth
										if !_rules[rule_]() {
											goto l255
										}
										if buffer[position] != rune(']') {
											goto l255
										}
										position++
										goto l254
									l255:
										position, tokenIndex, depth = position254, tokenIndex254, depth254
										{
											add(ruleAction52, position)
										}
									}
								l254:
									goto l249
								l250:
									position, tokenIndex, depth = position249, tokenIndex249, depth249
									{
										add(ruleAction53, position)
									}
								}
							l249:
								goto l248

								position, tokenIndex, depth = position247, tokenIndex247, depth247
							}
						l248:
							{
								add(ruleAction54, position)
							}
							depth--
							add(ruleexpression_metric, position244)
						}
						goto l234
					l243:
						position, tokenIndex, depth = position234, tokenIndex234, depth234
						if !_rules[rule_]() {
							goto l259
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l259
						}
						{
							position260, tokenIndex260, depth260 := position, tokenIndex, depth
							if !_rules[ruleexpression_start]() {
								goto l261
							}
							goto l260
						l261:
							position, tokenIndex, depth = position260, tokenIndex260, depth260
							{
								add(ruleAction38, position)
							}
						}
					l260:
						{
							position263, tokenIndex263, depth263 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l264
							}
							if !_rules[rulePAREN_CLOSE]() {
								goto l264
							}
							goto l263
						l264:
							position, tokenIndex, depth = position263, tokenIndex263, depth263
							{
								add(ruleAction39, position)
							}
						}
					l263:
						goto l234
					l259:
						position, tokenIndex, depth = position234, tokenIndex234, depth234
						if !_rules[rule_]() {
							goto l266
						}
						{
							position267 := position
							depth++
							{
								position268 := position
								depth++
								if !_rules[ruleNUMBER]() {
									goto l266
								}
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l266
								}
								position++
							l269:
								{
									position270, tokenIndex270, depth270 := position, tokenIndex, depth
									if c := buffer[position]; c < rune('a') || c > rune('z') {
										goto l270
									}
									position++
									goto l269
								l270:
									position, tokenIndex, depth = position270, tokenIndex270, depth270
								}
								if !_rules[ruleKEY]() {
									goto l266
								}
								depth--
								add(ruleDURATION, position268)
							}
							depth--
							add(rulePegText, position267)
						}
						{
							add(ruleAction40, position)
						}
						goto l234
					l266:
						position, tokenIndex, depth = position234, tokenIndex234, depth234
						if !_rules[rule_]() {
							goto l272
						}
						{
							position273 := position
							depth++
							if !_rules[ruleNUMBER]() {
								goto l272
							}
							depth--
							add(rulePegText, position273)
						}
						{
							add(ruleAction41, position)
						}
						goto l234
					l272:
						position, tokenIndex, depth = position234, tokenIndex234, depth234
						if !_rules[rule_]() {
							goto l231
						}
						if !_rules[ruleSTRING]() {
							goto l231
						}
						{
							add(ruleAction42, position)
						}
					}
				l234:
					depth--
					add(ruleexpression_atom_raw, position233)
				}
				if !_rules[ruleexpression_annotation]() {
					goto l231
				}
				depth--
				add(ruleexpression_atom, position232)
			}
			return true
		l231:
			position, tokenIndex, depth = position231, tokenIndex231, depth231
			return false
		},
		/* 17 expression_atom_raw <- <(expression_function / expression_metric / (_ PAREN_OPEN (expression_start / Action38) ((_ PAREN_CLOSE) / Action39)) / (_ <DURATION> Action40) / (_ <NUMBER> Action41) / (_ STRING Action42))> */
		nil,
		/* 18 expression_annotation_required <- <(_ '{' <(!'}' .)*> ('}' / (Action43 (' ' ('t' / 'T') ('o' / 'O') ' ' ('c' / 'C') ('l' / 'L') ('o' / 'O') ('s' / 'S') ('e' / 'E') ' ') Action44)) Action45)> */
		nil,
		/* 19 expression_annotation <- <expression_annotation_required?> */
		func() bool {
			{
				position279 := position
				depth++
				{
					position280, tokenIndex280, depth280 := position, tokenIndex, depth
					{
						position282 := position
						depth++
						if !_rules[rule_]() {
							goto l280
						}
						if buffer[position] != rune('{') {
							goto l280
						}
						position++
						{
							position283 := position
							depth++
						l284:
							{
								position285, tokenIndex285, depth285 := position, tokenIndex, depth
								{
									position286, tokenIndex286, depth286 := position, tokenIndex, depth
									if buffer[position] != rune('}') {
										goto l286
									}
									position++
									goto l285
								l286:
									position, tokenIndex, depth = position286, tokenIndex286, depth286
								}
								if !matchDot() {
									goto l285
								}
								goto l284
							l285:
								position, tokenIndex, depth = position285, tokenIndex285, depth285
							}
							depth--
							add(rulePegText, position283)
						}
						{
							position287, tokenIndex287, depth287 := position, tokenIndex, depth
							if buffer[position] != rune('}') {
								goto l288
							}
							position++
							goto l287
						l288:
							position, tokenIndex, depth = position287, tokenIndex287, depth287
							{
								add(ruleAction43, position)
							}
							if buffer[position] != rune(' ') {
								goto l280
							}
							position++
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
									goto l280
								}
								position++
							}
						l290:
							{
								position292, tokenIndex292, depth292 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l293
								}
								position++
								goto l292
							l293:
								position, tokenIndex, depth = position292, tokenIndex292, depth292
								if buffer[position] != rune('O') {
									goto l280
								}
								position++
							}
						l292:
							if buffer[position] != rune(' ') {
								goto l280
							}
							position++
							{
								position294, tokenIndex294, depth294 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l295
								}
								position++
								goto l294
							l295:
								position, tokenIndex, depth = position294, tokenIndex294, depth294
								if buffer[position] != rune('C') {
									goto l280
								}
								position++
							}
						l294:
							{
								position296, tokenIndex296, depth296 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l297
								}
								position++
								goto l296
							l297:
								position, tokenIndex, depth = position296, tokenIndex296, depth296
								if buffer[position] != rune('L') {
									goto l280
								}
								position++
							}
						l296:
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
									goto l280
								}
								position++
							}
						l298:
							{
								position300, tokenIndex300, depth300 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l301
								}
								position++
								goto l300
							l301:
								position, tokenIndex, depth = position300, tokenIndex300, depth300
								if buffer[position] != rune('S') {
									goto l280
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
									goto l280
								}
								position++
							}
						l302:
							if buffer[position] != rune(' ') {
								goto l280
							}
							position++
							{
								add(ruleAction44, position)
							}
						}
					l287:
						{
							add(ruleAction45, position)
						}
						depth--
						add(ruleexpression_annotation_required, position282)
					}
					goto l281
				l280:
					position, tokenIndex, depth = position280, tokenIndex280, depth280
				}
			l281:
				depth--
				add(ruleexpression_annotation, position279)
			}
			return true
		},
		/* 20 optionalGroupBy <- <(Action46 (groupByClause / collapseByClause)?)> */
		func() bool {
			{
				position307 := position
				depth++
				{
					add(ruleAction46, position)
				}
				{
					position309, tokenIndex309, depth309 := position, tokenIndex, depth
					{
						position311, tokenIndex311, depth311 := position, tokenIndex, depth
						{
							position313 := position
							depth++
							if !_rules[rule_]() {
								goto l312
							}
							{
								position314, tokenIndex314, depth314 := position, tokenIndex, depth
								if buffer[position] != rune('g') {
									goto l315
								}
								position++
								goto l314
							l315:
								position, tokenIndex, depth = position314, tokenIndex314, depth314
								if buffer[position] != rune('G') {
									goto l312
								}
								position++
							}
						l314:
							{
								position316, tokenIndex316, depth316 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l317
								}
								position++
								goto l316
							l317:
								position, tokenIndex, depth = position316, tokenIndex316, depth316
								if buffer[position] != rune('R') {
									goto l312
								}
								position++
							}
						l316:
							{
								position318, tokenIndex318, depth318 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l319
								}
								position++
								goto l318
							l319:
								position, tokenIndex, depth = position318, tokenIndex318, depth318
								if buffer[position] != rune('O') {
									goto l312
								}
								position++
							}
						l318:
							{
								position320, tokenIndex320, depth320 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l321
								}
								position++
								goto l320
							l321:
								position, tokenIndex, depth = position320, tokenIndex320, depth320
								if buffer[position] != rune('U') {
									goto l312
								}
								position++
							}
						l320:
							{
								position322, tokenIndex322, depth322 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l323
								}
								position++
								goto l322
							l323:
								position, tokenIndex, depth = position322, tokenIndex322, depth322
								if buffer[position] != rune('P') {
									goto l312
								}
								position++
							}
						l322:
							if !_rules[ruleKEY]() {
								goto l312
							}
							{
								position324, tokenIndex324, depth324 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l325
								}
								{
									position326, tokenIndex326, depth326 := position, tokenIndex, depth
									if buffer[position] != rune('b') {
										goto l327
									}
									position++
									goto l326
								l327:
									position, tokenIndex, depth = position326, tokenIndex326, depth326
									if buffer[position] != rune('B') {
										goto l325
									}
									position++
								}
							l326:
								{
									position328, tokenIndex328, depth328 := position, tokenIndex, depth
									if buffer[position] != rune('y') {
										goto l329
									}
									position++
									goto l328
								l329:
									position, tokenIndex, depth = position328, tokenIndex328, depth328
									if buffer[position] != rune('Y') {
										goto l325
									}
									position++
								}
							l328:
								if !_rules[ruleKEY]() {
									goto l325
								}
								goto l324
							l325:
								position, tokenIndex, depth = position324, tokenIndex324, depth324
								{
									add(ruleAction55, position)
								}
							}
						l324:
							{
								position331, tokenIndex331, depth331 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l332
								}
								{
									position333 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l332
									}
									depth--
									add(rulePegText, position333)
								}
								goto l331
							l332:
								position, tokenIndex, depth = position331, tokenIndex331, depth331
								{
									add(ruleAction56, position)
								}
							}
						l331:
							{
								add(ruleAction57, position)
							}
						l336:
							{
								position337, tokenIndex337, depth337 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l337
								}
								if !_rules[ruleCOMMA]() {
									goto l337
								}
								{
									position338, tokenIndex338, depth338 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l339
									}
									{
										position340 := position
										depth++
										if !_rules[ruleCOLUMN_NAME]() {
											goto l339
										}
										depth--
										add(rulePegText, position340)
									}
									goto l338
								l339:
									position, tokenIndex, depth = position338, tokenIndex338, depth338
									{
										add(ruleAction58, position)
									}
								}
							l338:
								{
									add(ruleAction59, position)
								}
								goto l336
							l337:
								position, tokenIndex, depth = position337, tokenIndex337, depth337
							}
							depth--
							add(rulegroupByClause, position313)
						}
						goto l311
					l312:
						position, tokenIndex, depth = position311, tokenIndex311, depth311
						{
							position343 := position
							depth++
							if !_rules[rule_]() {
								goto l309
							}
							{
								position344, tokenIndex344, depth344 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l345
								}
								position++
								goto l344
							l345:
								position, tokenIndex, depth = position344, tokenIndex344, depth344
								if buffer[position] != rune('C') {
									goto l309
								}
								position++
							}
						l344:
							{
								position346, tokenIndex346, depth346 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l347
								}
								position++
								goto l346
							l347:
								position, tokenIndex, depth = position346, tokenIndex346, depth346
								if buffer[position] != rune('O') {
									goto l309
								}
								position++
							}
						l346:
							{
								position348, tokenIndex348, depth348 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l349
								}
								position++
								goto l348
							l349:
								position, tokenIndex, depth = position348, tokenIndex348, depth348
								if buffer[position] != rune('L') {
									goto l309
								}
								position++
							}
						l348:
							{
								position350, tokenIndex350, depth350 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l351
								}
								position++
								goto l350
							l351:
								position, tokenIndex, depth = position350, tokenIndex350, depth350
								if buffer[position] != rune('L') {
									goto l309
								}
								position++
							}
						l350:
							{
								position352, tokenIndex352, depth352 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l353
								}
								position++
								goto l352
							l353:
								position, tokenIndex, depth = position352, tokenIndex352, depth352
								if buffer[position] != rune('A') {
									goto l309
								}
								position++
							}
						l352:
							{
								position354, tokenIndex354, depth354 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l355
								}
								position++
								goto l354
							l355:
								position, tokenIndex, depth = position354, tokenIndex354, depth354
								if buffer[position] != rune('P') {
									goto l309
								}
								position++
							}
						l354:
							{
								position356, tokenIndex356, depth356 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l357
								}
								position++
								goto l356
							l357:
								position, tokenIndex, depth = position356, tokenIndex356, depth356
								if buffer[position] != rune('S') {
									goto l309
								}
								position++
							}
						l356:
							{
								position358, tokenIndex358, depth358 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l359
								}
								position++
								goto l358
							l359:
								position, tokenIndex, depth = position358, tokenIndex358, depth358
								if buffer[position] != rune('E') {
									goto l309
								}
								position++
							}
						l358:
							if !_rules[ruleKEY]() {
								goto l309
							}
							{
								position360, tokenIndex360, depth360 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l361
								}
								{
									position362, tokenIndex362, depth362 := position, tokenIndex, depth
									if buffer[position] != rune('b') {
										goto l363
									}
									position++
									goto l362
								l363:
									position, tokenIndex, depth = position362, tokenIndex362, depth362
									if buffer[position] != rune('B') {
										goto l361
									}
									position++
								}
							l362:
								{
									position364, tokenIndex364, depth364 := position, tokenIndex, depth
									if buffer[position] != rune('y') {
										goto l365
									}
									position++
									goto l364
								l365:
									position, tokenIndex, depth = position364, tokenIndex364, depth364
									if buffer[position] != rune('Y') {
										goto l361
									}
									position++
								}
							l364:
								if !_rules[ruleKEY]() {
									goto l361
								}
								goto l360
							l361:
								position, tokenIndex, depth = position360, tokenIndex360, depth360
								{
									add(ruleAction60, position)
								}
							}
						l360:
							{
								position367, tokenIndex367, depth367 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l368
								}
								{
									position369 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l368
									}
									depth--
									add(rulePegText, position369)
								}
								goto l367
							l368:
								position, tokenIndex, depth = position367, tokenIndex367, depth367
								{
									add(ruleAction61, position)
								}
							}
						l367:
							{
								add(ruleAction62, position)
							}
						l372:
							{
								position373, tokenIndex373, depth373 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l373
								}
								if !_rules[ruleCOMMA]() {
									goto l373
								}
								{
									position374, tokenIndex374, depth374 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l375
									}
									{
										position376 := position
										depth++
										if !_rules[ruleCOLUMN_NAME]() {
											goto l375
										}
										depth--
										add(rulePegText, position376)
									}
									goto l374
								l375:
									position, tokenIndex, depth = position374, tokenIndex374, depth374
									{
										add(ruleAction63, position)
									}
								}
							l374:
								{
									add(ruleAction64, position)
								}
								goto l372
							l373:
								position, tokenIndex, depth = position373, tokenIndex373, depth373
							}
							depth--
							add(rulecollapseByClause, position343)
						}
					}
				l311:
					goto l310
				l309:
					position, tokenIndex, depth = position309, tokenIndex309, depth309
				}
			l310:
				depth--
				add(ruleoptionalGroupBy, position307)
			}
			return true
		},
		/* 21 expression_function <- <(_ <IDENTIFIER> Action47 _ PAREN_OPEN expressionList optionalGroupBy ((_ PAREN_CLOSE) / Action48) Action49)> */
		nil,
		/* 22 expression_metric <- <(_ <IDENTIFIER> Action50 ((_ '[' (predicate_1 / Action51) ((_ ']') / Action52)) / Action53)? Action54)> */
		nil,
		/* 23 groupByClause <- <(_ (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P')) KEY ((_ (('b' / 'B') ('y' / 'Y')) KEY) / Action55) ((_ <COLUMN_NAME>) / Action56) Action57 (_ COMMA ((_ <COLUMN_NAME>) / Action58) Action59)*)> */
		nil,
		/* 24 collapseByClause <- <(_ (('c' / 'C') ('o' / 'O') ('l' / 'L') ('l' / 'L') ('a' / 'A') ('p' / 'P') ('s' / 'S') ('e' / 'E')) KEY ((_ (('b' / 'B') ('y' / 'Y')) KEY) / Action60) ((_ <COLUMN_NAME>) / Action61) Action62 (_ COMMA ((_ <COLUMN_NAME>) / Action63) Action64)*)> */
		nil,
		/* 25 predicateClause <- <(_ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY ((_ predicate_1) / Action65))> */
		nil,
		/* 26 predicate_1 <- <((predicate_2 _ OP_OR (predicate_1 / Action66) Action67) / predicate_2)> */
		func() bool {
			position384, tokenIndex384, depth384 := position, tokenIndex, depth
			{
				position385 := position
				depth++
				{
					position386, tokenIndex386, depth386 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l387
					}
					if !_rules[rule_]() {
						goto l387
					}
					{
						position388 := position
						depth++
						{
							position389, tokenIndex389, depth389 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l390
							}
							position++
							goto l389
						l390:
							position, tokenIndex, depth = position389, tokenIndex389, depth389
							if buffer[position] != rune('O') {
								goto l387
							}
							position++
						}
					l389:
						{
							position391, tokenIndex391, depth391 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l392
							}
							position++
							goto l391
						l392:
							position, tokenIndex, depth = position391, tokenIndex391, depth391
							if buffer[position] != rune('R') {
								goto l387
							}
							position++
						}
					l391:
						if !_rules[ruleKEY]() {
							goto l387
						}
						depth--
						add(ruleOP_OR, position388)
					}
					{
						position393, tokenIndex393, depth393 := position, tokenIndex, depth
						if !_rules[rulepredicate_1]() {
							goto l394
						}
						goto l393
					l394:
						position, tokenIndex, depth = position393, tokenIndex393, depth393
						{
							add(ruleAction66, position)
						}
					}
				l393:
					{
						add(ruleAction67, position)
					}
					goto l386
				l387:
					position, tokenIndex, depth = position386, tokenIndex386, depth386
					if !_rules[rulepredicate_2]() {
						goto l384
					}
				}
			l386:
				depth--
				add(rulepredicate_1, position385)
			}
			return true
		l384:
			position, tokenIndex, depth = position384, tokenIndex384, depth384
			return false
		},
		/* 27 predicate_2 <- <((predicate_3 _ OP_AND (predicate_2 / Action68) Action69) / predicate_3)> */
		func() bool {
			position397, tokenIndex397, depth397 := position, tokenIndex, depth
			{
				position398 := position
				depth++
				{
					position399, tokenIndex399, depth399 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l400
					}
					if !_rules[rule_]() {
						goto l400
					}
					{
						position401 := position
						depth++
						{
							position402, tokenIndex402, depth402 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l403
							}
							position++
							goto l402
						l403:
							position, tokenIndex, depth = position402, tokenIndex402, depth402
							if buffer[position] != rune('A') {
								goto l400
							}
							position++
						}
					l402:
						{
							position404, tokenIndex404, depth404 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l405
							}
							position++
							goto l404
						l405:
							position, tokenIndex, depth = position404, tokenIndex404, depth404
							if buffer[position] != rune('N') {
								goto l400
							}
							position++
						}
					l404:
						{
							position406, tokenIndex406, depth406 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l407
							}
							position++
							goto l406
						l407:
							position, tokenIndex, depth = position406, tokenIndex406, depth406
							if buffer[position] != rune('D') {
								goto l400
							}
							position++
						}
					l406:
						if !_rules[ruleKEY]() {
							goto l400
						}
						depth--
						add(ruleOP_AND, position401)
					}
					{
						position408, tokenIndex408, depth408 := position, tokenIndex, depth
						if !_rules[rulepredicate_2]() {
							goto l409
						}
						goto l408
					l409:
						position, tokenIndex, depth = position408, tokenIndex408, depth408
						{
							add(ruleAction68, position)
						}
					}
				l408:
					{
						add(ruleAction69, position)
					}
					goto l399
				l400:
					position, tokenIndex, depth = position399, tokenIndex399, depth399
					if !_rules[rulepredicate_3]() {
						goto l397
					}
				}
			l399:
				depth--
				add(rulepredicate_2, position398)
			}
			return true
		l397:
			position, tokenIndex, depth = position397, tokenIndex397, depth397
			return false
		},
		/* 28 predicate_3 <- <((_ OP_NOT (predicate_3 / Action70) Action71) / (_ PAREN_OPEN (predicate_1 / Action72) ((_ PAREN_CLOSE) / Action73)) / tagMatcher)> */
		func() bool {
			position412, tokenIndex412, depth412 := position, tokenIndex, depth
			{
				position413 := position
				depth++
				{
					position414, tokenIndex414, depth414 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l415
					}
					{
						position416 := position
						depth++
						{
							position417, tokenIndex417, depth417 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l418
							}
							position++
							goto l417
						l418:
							position, tokenIndex, depth = position417, tokenIndex417, depth417
							if buffer[position] != rune('N') {
								goto l415
							}
							position++
						}
					l417:
						{
							position419, tokenIndex419, depth419 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l420
							}
							position++
							goto l419
						l420:
							position, tokenIndex, depth = position419, tokenIndex419, depth419
							if buffer[position] != rune('O') {
								goto l415
							}
							position++
						}
					l419:
						{
							position421, tokenIndex421, depth421 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l422
							}
							position++
							goto l421
						l422:
							position, tokenIndex, depth = position421, tokenIndex421, depth421
							if buffer[position] != rune('T') {
								goto l415
							}
							position++
						}
					l421:
						if !_rules[ruleKEY]() {
							goto l415
						}
						depth--
						add(ruleOP_NOT, position416)
					}
					{
						position423, tokenIndex423, depth423 := position, tokenIndex, depth
						if !_rules[rulepredicate_3]() {
							goto l424
						}
						goto l423
					l424:
						position, tokenIndex, depth = position423, tokenIndex423, depth423
						{
							add(ruleAction70, position)
						}
					}
				l423:
					{
						add(ruleAction71, position)
					}
					goto l414
				l415:
					position, tokenIndex, depth = position414, tokenIndex414, depth414
					if !_rules[rule_]() {
						goto l427
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l427
					}
					{
						position428, tokenIndex428, depth428 := position, tokenIndex, depth
						if !_rules[rulepredicate_1]() {
							goto l429
						}
						goto l428
					l429:
						position, tokenIndex, depth = position428, tokenIndex428, depth428
						{
							add(ruleAction72, position)
						}
					}
				l428:
					{
						position431, tokenIndex431, depth431 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l432
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l432
						}
						goto l431
					l432:
						position, tokenIndex, depth = position431, tokenIndex431, depth431
						{
							add(ruleAction73, position)
						}
					}
				l431:
					goto l414
				l427:
					position, tokenIndex, depth = position414, tokenIndex414, depth414
					{
						position434 := position
						depth++
						if !_rules[ruletagName]() {
							goto l412
						}
						{
							position435, tokenIndex435, depth435 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l436
							}
							if buffer[position] != rune('=') {
								goto l436
							}
							position++
							{
								position437, tokenIndex437, depth437 := position, tokenIndex, depth
								if !_rules[ruleliteralString]() {
									goto l438
								}
								goto l437
							l438:
								position, tokenIndex, depth = position437, tokenIndex437, depth437
								{
									add(ruleAction74, position)
								}
							}
						l437:
							{
								add(ruleAction75, position)
							}
							goto l435
						l436:
							position, tokenIndex, depth = position435, tokenIndex435, depth435
							if !_rules[rule_]() {
								goto l441
							}
							if buffer[position] != rune('!') {
								goto l441
							}
							position++
							if buffer[position] != rune('=') {
								goto l441
							}
							position++
							{
								position442, tokenIndex442, depth442 := position, tokenIndex, depth
								if !_rules[ruleliteralString]() {
									goto l443
								}
								goto l442
							l443:
								position, tokenIndex, depth = position442, tokenIndex442, depth442
								{
									add(ruleAction76, position)
								}
							}
						l442:
							{
								add(ruleAction77, position)
							}
							{
								add(ruleAction78, position)
							}
							goto l435
						l441:
							position, tokenIndex, depth = position435, tokenIndex435, depth435
							if !_rules[rule_]() {
								goto l447
							}
							{
								position448, tokenIndex448, depth448 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l449
								}
								position++
								goto l448
							l449:
								position, tokenIndex, depth = position448, tokenIndex448, depth448
								if buffer[position] != rune('M') {
									goto l447
								}
								position++
							}
						l448:
							{
								position450, tokenIndex450, depth450 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l451
								}
								position++
								goto l450
							l451:
								position, tokenIndex, depth = position450, tokenIndex450, depth450
								if buffer[position] != rune('A') {
									goto l447
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
									goto l447
								}
								position++
							}
						l452:
							{
								position454, tokenIndex454, depth454 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l455
								}
								position++
								goto l454
							l455:
								position, tokenIndex, depth = position454, tokenIndex454, depth454
								if buffer[position] != rune('C') {
									goto l447
								}
								position++
							}
						l454:
							{
								position456, tokenIndex456, depth456 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l457
								}
								position++
								goto l456
							l457:
								position, tokenIndex, depth = position456, tokenIndex456, depth456
								if buffer[position] != rune('H') {
									goto l447
								}
								position++
							}
						l456:
							if !_rules[ruleKEY]() {
								goto l447
							}
							{
								position458, tokenIndex458, depth458 := position, tokenIndex, depth
								if !_rules[ruleliteralString]() {
									goto l459
								}
								goto l458
							l459:
								position, tokenIndex, depth = position458, tokenIndex458, depth458
								{
									add(ruleAction79, position)
								}
							}
						l458:
							{
								add(ruleAction80, position)
							}
							goto l435
						l447:
							position, tokenIndex, depth = position435, tokenIndex435, depth435
							if !_rules[rule_]() {
								goto l462
							}
							{
								position463, tokenIndex463, depth463 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l464
								}
								position++
								goto l463
							l464:
								position, tokenIndex, depth = position463, tokenIndex463, depth463
								if buffer[position] != rune('I') {
									goto l462
								}
								position++
							}
						l463:
							{
								position465, tokenIndex465, depth465 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l466
								}
								position++
								goto l465
							l466:
								position, tokenIndex, depth = position465, tokenIndex465, depth465
								if buffer[position] != rune('N') {
									goto l462
								}
								position++
							}
						l465:
							if !_rules[ruleKEY]() {
								goto l462
							}
							{
								position467, tokenIndex467, depth467 := position, tokenIndex, depth
								{
									position469 := position
									depth++
									{
										add(ruleAction85, position)
									}
									if !_rules[rule_]() {
										goto l468
									}
									if !_rules[rulePAREN_OPEN]() {
										goto l468
									}
									{
										position471, tokenIndex471, depth471 := position, tokenIndex, depth
										if !_rules[ruleliteralListString]() {
											goto l472
										}
										goto l471
									l472:
										position, tokenIndex, depth = position471, tokenIndex471, depth471
										{
											add(ruleAction86, position)
										}
									}
								l471:
								l474:
									{
										position475, tokenIndex475, depth475 := position, tokenIndex, depth
										if !_rules[rule_]() {
											goto l475
										}
										if !_rules[ruleCOMMA]() {
											goto l475
										}
										{
											position476, tokenIndex476, depth476 := position, tokenIndex, depth
											if !_rules[ruleliteralListString]() {
												goto l477
											}
											goto l476
										l477:
											position, tokenIndex, depth = position476, tokenIndex476, depth476
											{
												add(ruleAction87, position)
											}
										}
									l476:
										goto l474
									l475:
										position, tokenIndex, depth = position475, tokenIndex475, depth475
									}
									{
										position479, tokenIndex479, depth479 := position, tokenIndex, depth
										if !_rules[rule_]() {
											goto l480
										}
										if !_rules[rulePAREN_CLOSE]() {
											goto l480
										}
										goto l479
									l480:
										position, tokenIndex, depth = position479, tokenIndex479, depth479
										{
											add(ruleAction88, position)
										}
									}
								l479:
									depth--
									add(ruleliteralList, position469)
								}
								goto l467
							l468:
								position, tokenIndex, depth = position467, tokenIndex467, depth467
								{
									add(ruleAction81, position)
								}
							}
						l467:
							{
								add(ruleAction82, position)
							}
							goto l435
						l462:
							position, tokenIndex, depth = position435, tokenIndex435, depth435
							{
								add(ruleAction83, position)
							}
						}
					l435:
						depth--
						add(ruletagMatcher, position434)
					}
				}
			l414:
				depth--
				add(rulepredicate_3, position413)
			}
			return true
		l412:
			position, tokenIndex, depth = position412, tokenIndex412, depth412
			return false
		},
		/* 29 tagMatcher <- <(tagName ((_ '=' (literalString / Action74) Action75) / (_ ('!' '=') (literalString / Action76) Action77 Action78) / (_ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H')) KEY (literalString / Action79) Action80) / (_ (('i' / 'I') ('n' / 'N')) KEY (literalList / Action81) Action82) / Action83))> */
		nil,
		/* 30 literalString <- <(_ STRING Action84)> */
		func() bool {
			position486, tokenIndex486, depth486 := position, tokenIndex, depth
			{
				position487 := position
				depth++
				if !_rules[rule_]() {
					goto l486
				}
				if !_rules[ruleSTRING]() {
					goto l486
				}
				{
					add(ruleAction84, position)
				}
				depth--
				add(ruleliteralString, position487)
			}
			return true
		l486:
			position, tokenIndex, depth = position486, tokenIndex486, depth486
			return false
		},
		/* 31 literalList <- <(Action85 _ PAREN_OPEN (literalListString / Action86) (_ COMMA (literalListString / Action87))* ((_ PAREN_CLOSE) / Action88))> */
		nil,
		/* 32 literalListString <- <(_ STRING Action89)> */
		func() bool {
			position490, tokenIndex490, depth490 := position, tokenIndex, depth
			{
				position491 := position
				depth++
				if !_rules[rule_]() {
					goto l490
				}
				if !_rules[ruleSTRING]() {
					goto l490
				}
				{
					add(ruleAction89, position)
				}
				depth--
				add(ruleliteralListString, position491)
			}
			return true
		l490:
			position, tokenIndex, depth = position490, tokenIndex490, depth490
			return false
		},
		/* 33 tagName <- <(_ <TAG_NAME> Action90)> */
		func() bool {
			position493, tokenIndex493, depth493 := position, tokenIndex, depth
			{
				position494 := position
				depth++
				if !_rules[rule_]() {
					goto l493
				}
				{
					position495 := position
					depth++
					{
						position496 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l493
						}
						depth--
						add(ruleTAG_NAME, position496)
					}
					depth--
					add(rulePegText, position495)
				}
				{
					add(ruleAction90, position)
				}
				depth--
				add(ruletagName, position494)
			}
			return true
		l493:
			position, tokenIndex, depth = position493, tokenIndex493, depth493
			return false
		},
		/* 34 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position498, tokenIndex498, depth498 := position, tokenIndex, depth
			{
				position499 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l498
				}
				depth--
				add(ruleCOLUMN_NAME, position499)
			}
			return true
		l498:
			position, tokenIndex, depth = position498, tokenIndex498, depth498
			return false
		},
		/* 35 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 36 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 37 IDENTIFIER <- <(('`' CHAR* ('`' / Action91)) / (_ !(KEYWORD KEY) ID_SEGMENT ('.' (ID_SEGMENT / Action92))*))> */
		func() bool {
			position502, tokenIndex502, depth502 := position, tokenIndex, depth
			{
				position503 := position
				depth++
				{
					position504, tokenIndex504, depth504 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l505
					}
					position++
				l506:
					{
						position507, tokenIndex507, depth507 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l507
						}
						goto l506
					l507:
						position, tokenIndex, depth = position507, tokenIndex507, depth507
					}
					{
						position508, tokenIndex508, depth508 := position, tokenIndex, depth
						if buffer[position] != rune('`') {
							goto l509
						}
						position++
						goto l508
					l509:
						position, tokenIndex, depth = position508, tokenIndex508, depth508
						{
							add(ruleAction91, position)
						}
					}
				l508:
					goto l504
				l505:
					position, tokenIndex, depth = position504, tokenIndex504, depth504
					if !_rules[rule_]() {
						goto l502
					}
					{
						position511, tokenIndex511, depth511 := position, tokenIndex, depth
						{
							position512 := position
							depth++
							{
								position513, tokenIndex513, depth513 := position, tokenIndex, depth
								{
									position515, tokenIndex515, depth515 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l516
									}
									position++
									goto l515
								l516:
									position, tokenIndex, depth = position515, tokenIndex515, depth515
									if buffer[position] != rune('A') {
										goto l514
									}
									position++
								}
							l515:
								{
									position517, tokenIndex517, depth517 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l518
									}
									position++
									goto l517
								l518:
									position, tokenIndex, depth = position517, tokenIndex517, depth517
									if buffer[position] != rune('L') {
										goto l514
									}
									position++
								}
							l517:
								{
									position519, tokenIndex519, depth519 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l520
									}
									position++
									goto l519
								l520:
									position, tokenIndex, depth = position519, tokenIndex519, depth519
									if buffer[position] != rune('L') {
										goto l514
									}
									position++
								}
							l519:
								goto l513
							l514:
								position, tokenIndex, depth = position513, tokenIndex513, depth513
								{
									position522, tokenIndex522, depth522 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l523
									}
									position++
									goto l522
								l523:
									position, tokenIndex, depth = position522, tokenIndex522, depth522
									if buffer[position] != rune('A') {
										goto l521
									}
									position++
								}
							l522:
								{
									position524, tokenIndex524, depth524 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l525
									}
									position++
									goto l524
								l525:
									position, tokenIndex, depth = position524, tokenIndex524, depth524
									if buffer[position] != rune('N') {
										goto l521
									}
									position++
								}
							l524:
								{
									position526, tokenIndex526, depth526 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l527
									}
									position++
									goto l526
								l527:
									position, tokenIndex, depth = position526, tokenIndex526, depth526
									if buffer[position] != rune('D') {
										goto l521
									}
									position++
								}
							l526:
								goto l513
							l521:
								position, tokenIndex, depth = position513, tokenIndex513, depth513
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
										goto l528
									}
									position++
								}
							l529:
								{
									position531, tokenIndex531, depth531 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l532
									}
									position++
									goto l531
								l532:
									position, tokenIndex, depth = position531, tokenIndex531, depth531
									if buffer[position] != rune('A') {
										goto l528
									}
									position++
								}
							l531:
								{
									position533, tokenIndex533, depth533 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l534
									}
									position++
									goto l533
								l534:
									position, tokenIndex, depth = position533, tokenIndex533, depth533
									if buffer[position] != rune('T') {
										goto l528
									}
									position++
								}
							l533:
								{
									position535, tokenIndex535, depth535 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l536
									}
									position++
									goto l535
								l536:
									position, tokenIndex, depth = position535, tokenIndex535, depth535
									if buffer[position] != rune('C') {
										goto l528
									}
									position++
								}
							l535:
								{
									position537, tokenIndex537, depth537 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l538
									}
									position++
									goto l537
								l538:
									position, tokenIndex, depth = position537, tokenIndex537, depth537
									if buffer[position] != rune('H') {
										goto l528
									}
									position++
								}
							l537:
								goto l513
							l528:
								position, tokenIndex, depth = position513, tokenIndex513, depth513
								{
									position540, tokenIndex540, depth540 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l541
									}
									position++
									goto l540
								l541:
									position, tokenIndex, depth = position540, tokenIndex540, depth540
									if buffer[position] != rune('S') {
										goto l539
									}
									position++
								}
							l540:
								{
									position542, tokenIndex542, depth542 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l543
									}
									position++
									goto l542
								l543:
									position, tokenIndex, depth = position542, tokenIndex542, depth542
									if buffer[position] != rune('E') {
										goto l539
									}
									position++
								}
							l542:
								{
									position544, tokenIndex544, depth544 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l545
									}
									position++
									goto l544
								l545:
									position, tokenIndex, depth = position544, tokenIndex544, depth544
									if buffer[position] != rune('L') {
										goto l539
									}
									position++
								}
							l544:
								{
									position546, tokenIndex546, depth546 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l547
									}
									position++
									goto l546
								l547:
									position, tokenIndex, depth = position546, tokenIndex546, depth546
									if buffer[position] != rune('E') {
										goto l539
									}
									position++
								}
							l546:
								{
									position548, tokenIndex548, depth548 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l549
									}
									position++
									goto l548
								l549:
									position, tokenIndex, depth = position548, tokenIndex548, depth548
									if buffer[position] != rune('C') {
										goto l539
									}
									position++
								}
							l548:
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
										goto l539
									}
									position++
								}
							l550:
								goto l513
							l539:
								position, tokenIndex, depth = position513, tokenIndex513, depth513
								{
									switch buffer[position] {
									case 'M', 'm':
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
												goto l511
											}
											position++
										}
									l553:
										{
											position555, tokenIndex555, depth555 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l556
											}
											position++
											goto l555
										l556:
											position, tokenIndex, depth = position555, tokenIndex555, depth555
											if buffer[position] != rune('E') {
												goto l511
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
												goto l511
											}
											position++
										}
									l557:
										{
											position559, tokenIndex559, depth559 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l560
											}
											position++
											goto l559
										l560:
											position, tokenIndex, depth = position559, tokenIndex559, depth559
											if buffer[position] != rune('R') {
												goto l511
											}
											position++
										}
									l559:
										{
											position561, tokenIndex561, depth561 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l562
											}
											position++
											goto l561
										l562:
											position, tokenIndex, depth = position561, tokenIndex561, depth561
											if buffer[position] != rune('I') {
												goto l511
											}
											position++
										}
									l561:
										{
											position563, tokenIndex563, depth563 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l564
											}
											position++
											goto l563
										l564:
											position, tokenIndex, depth = position563, tokenIndex563, depth563
											if buffer[position] != rune('C') {
												goto l511
											}
											position++
										}
									l563:
										{
											position565, tokenIndex565, depth565 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l566
											}
											position++
											goto l565
										l566:
											position, tokenIndex, depth = position565, tokenIndex565, depth565
											if buffer[position] != rune('S') {
												goto l511
											}
											position++
										}
									l565:
										break
									case 'W', 'w':
										{
											position567, tokenIndex567, depth567 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l568
											}
											position++
											goto l567
										l568:
											position, tokenIndex, depth = position567, tokenIndex567, depth567
											if buffer[position] != rune('W') {
												goto l511
											}
											position++
										}
									l567:
										{
											position569, tokenIndex569, depth569 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l570
											}
											position++
											goto l569
										l570:
											position, tokenIndex, depth = position569, tokenIndex569, depth569
											if buffer[position] != rune('H') {
												goto l511
											}
											position++
										}
									l569:
										{
											position571, tokenIndex571, depth571 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l572
											}
											position++
											goto l571
										l572:
											position, tokenIndex, depth = position571, tokenIndex571, depth571
											if buffer[position] != rune('E') {
												goto l511
											}
											position++
										}
									l571:
										{
											position573, tokenIndex573, depth573 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l574
											}
											position++
											goto l573
										l574:
											position, tokenIndex, depth = position573, tokenIndex573, depth573
											if buffer[position] != rune('R') {
												goto l511
											}
											position++
										}
									l573:
										{
											position575, tokenIndex575, depth575 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l576
											}
											position++
											goto l575
										l576:
											position, tokenIndex, depth = position575, tokenIndex575, depth575
											if buffer[position] != rune('E') {
												goto l511
											}
											position++
										}
									l575:
										break
									case 'O', 'o':
										{
											position577, tokenIndex577, depth577 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l578
											}
											position++
											goto l577
										l578:
											position, tokenIndex, depth = position577, tokenIndex577, depth577
											if buffer[position] != rune('O') {
												goto l511
											}
											position++
										}
									l577:
										{
											position579, tokenIndex579, depth579 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l580
											}
											position++
											goto l579
										l580:
											position, tokenIndex, depth = position579, tokenIndex579, depth579
											if buffer[position] != rune('R') {
												goto l511
											}
											position++
										}
									l579:
										break
									case 'N', 'n':
										{
											position581, tokenIndex581, depth581 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l582
											}
											position++
											goto l581
										l582:
											position, tokenIndex, depth = position581, tokenIndex581, depth581
											if buffer[position] != rune('N') {
												goto l511
											}
											position++
										}
									l581:
										{
											position583, tokenIndex583, depth583 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l584
											}
											position++
											goto l583
										l584:
											position, tokenIndex, depth = position583, tokenIndex583, depth583
											if buffer[position] != rune('O') {
												goto l511
											}
											position++
										}
									l583:
										{
											position585, tokenIndex585, depth585 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l586
											}
											position++
											goto l585
										l586:
											position, tokenIndex, depth = position585, tokenIndex585, depth585
											if buffer[position] != rune('T') {
												goto l511
											}
											position++
										}
									l585:
										break
									case 'I', 'i':
										{
											position587, tokenIndex587, depth587 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l588
											}
											position++
											goto l587
										l588:
											position, tokenIndex, depth = position587, tokenIndex587, depth587
											if buffer[position] != rune('I') {
												goto l511
											}
											position++
										}
									l587:
										{
											position589, tokenIndex589, depth589 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l590
											}
											position++
											goto l589
										l590:
											position, tokenIndex, depth = position589, tokenIndex589, depth589
											if buffer[position] != rune('N') {
												goto l511
											}
											position++
										}
									l589:
										break
									case 'C', 'c':
										{
											position591, tokenIndex591, depth591 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l592
											}
											position++
											goto l591
										l592:
											position, tokenIndex, depth = position591, tokenIndex591, depth591
											if buffer[position] != rune('C') {
												goto l511
											}
											position++
										}
									l591:
										{
											position593, tokenIndex593, depth593 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l594
											}
											position++
											goto l593
										l594:
											position, tokenIndex, depth = position593, tokenIndex593, depth593
											if buffer[position] != rune('O') {
												goto l511
											}
											position++
										}
									l593:
										{
											position595, tokenIndex595, depth595 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l596
											}
											position++
											goto l595
										l596:
											position, tokenIndex, depth = position595, tokenIndex595, depth595
											if buffer[position] != rune('L') {
												goto l511
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
												goto l511
											}
											position++
										}
									l597:
										{
											position599, tokenIndex599, depth599 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l600
											}
											position++
											goto l599
										l600:
											position, tokenIndex, depth = position599, tokenIndex599, depth599
											if buffer[position] != rune('A') {
												goto l511
											}
											position++
										}
									l599:
										{
											position601, tokenIndex601, depth601 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l602
											}
											position++
											goto l601
										l602:
											position, tokenIndex, depth = position601, tokenIndex601, depth601
											if buffer[position] != rune('P') {
												goto l511
											}
											position++
										}
									l601:
										{
											position603, tokenIndex603, depth603 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l604
											}
											position++
											goto l603
										l604:
											position, tokenIndex, depth = position603, tokenIndex603, depth603
											if buffer[position] != rune('S') {
												goto l511
											}
											position++
										}
									l603:
										{
											position605, tokenIndex605, depth605 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l606
											}
											position++
											goto l605
										l606:
											position, tokenIndex, depth = position605, tokenIndex605, depth605
											if buffer[position] != rune('E') {
												goto l511
											}
											position++
										}
									l605:
										break
									case 'G', 'g':
										{
											position607, tokenIndex607, depth607 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l608
											}
											position++
											goto l607
										l608:
											position, tokenIndex, depth = position607, tokenIndex607, depth607
											if buffer[position] != rune('G') {
												goto l511
											}
											position++
										}
									l607:
										{
											position609, tokenIndex609, depth609 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l610
											}
											position++
											goto l609
										l610:
											position, tokenIndex, depth = position609, tokenIndex609, depth609
											if buffer[position] != rune('R') {
												goto l511
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
												goto l511
											}
											position++
										}
									l611:
										{
											position613, tokenIndex613, depth613 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l614
											}
											position++
											goto l613
										l614:
											position, tokenIndex, depth = position613, tokenIndex613, depth613
											if buffer[position] != rune('U') {
												goto l511
											}
											position++
										}
									l613:
										{
											position615, tokenIndex615, depth615 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l616
											}
											position++
											goto l615
										l616:
											position, tokenIndex, depth = position615, tokenIndex615, depth615
											if buffer[position] != rune('P') {
												goto l511
											}
											position++
										}
									l615:
										break
									case 'D', 'd':
										{
											position617, tokenIndex617, depth617 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l618
											}
											position++
											goto l617
										l618:
											position, tokenIndex, depth = position617, tokenIndex617, depth617
											if buffer[position] != rune('D') {
												goto l511
											}
											position++
										}
									l617:
										{
											position619, tokenIndex619, depth619 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l620
											}
											position++
											goto l619
										l620:
											position, tokenIndex, depth = position619, tokenIndex619, depth619
											if buffer[position] != rune('E') {
												goto l511
											}
											position++
										}
									l619:
										{
											position621, tokenIndex621, depth621 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l622
											}
											position++
											goto l621
										l622:
											position, tokenIndex, depth = position621, tokenIndex621, depth621
											if buffer[position] != rune('S') {
												goto l511
											}
											position++
										}
									l621:
										{
											position623, tokenIndex623, depth623 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l624
											}
											position++
											goto l623
										l624:
											position, tokenIndex, depth = position623, tokenIndex623, depth623
											if buffer[position] != rune('C') {
												goto l511
											}
											position++
										}
									l623:
										{
											position625, tokenIndex625, depth625 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l626
											}
											position++
											goto l625
										l626:
											position, tokenIndex, depth = position625, tokenIndex625, depth625
											if buffer[position] != rune('R') {
												goto l511
											}
											position++
										}
									l625:
										{
											position627, tokenIndex627, depth627 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l628
											}
											position++
											goto l627
										l628:
											position, tokenIndex, depth = position627, tokenIndex627, depth627
											if buffer[position] != rune('I') {
												goto l511
											}
											position++
										}
									l627:
										{
											position629, tokenIndex629, depth629 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l630
											}
											position++
											goto l629
										l630:
											position, tokenIndex, depth = position629, tokenIndex629, depth629
											if buffer[position] != rune('B') {
												goto l511
											}
											position++
										}
									l629:
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
												goto l511
											}
											position++
										}
									l631:
										break
									case 'B', 'b':
										{
											position633, tokenIndex633, depth633 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l634
											}
											position++
											goto l633
										l634:
											position, tokenIndex, depth = position633, tokenIndex633, depth633
											if buffer[position] != rune('B') {
												goto l511
											}
											position++
										}
									l633:
										{
											position635, tokenIndex635, depth635 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l636
											}
											position++
											goto l635
										l636:
											position, tokenIndex, depth = position635, tokenIndex635, depth635
											if buffer[position] != rune('Y') {
												goto l511
											}
											position++
										}
									l635:
										break
									case 'A', 'a':
										{
											position637, tokenIndex637, depth637 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l638
											}
											position++
											goto l637
										l638:
											position, tokenIndex, depth = position637, tokenIndex637, depth637
											if buffer[position] != rune('A') {
												goto l511
											}
											position++
										}
									l637:
										{
											position639, tokenIndex639, depth639 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l640
											}
											position++
											goto l639
										l640:
											position, tokenIndex, depth = position639, tokenIndex639, depth639
											if buffer[position] != rune('S') {
												goto l511
											}
											position++
										}
									l639:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l511
										}
										break
									}
								}

							}
						l513:
							depth--
							add(ruleKEYWORD, position512)
						}
						if !_rules[ruleKEY]() {
							goto l511
						}
						goto l502
					l511:
						position, tokenIndex, depth = position511, tokenIndex511, depth511
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l502
					}
				l641:
					{
						position642, tokenIndex642, depth642 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l642
						}
						position++
						{
							position643, tokenIndex643, depth643 := position, tokenIndex, depth
							if !_rules[ruleID_SEGMENT]() {
								goto l644
							}
							goto l643
						l644:
							position, tokenIndex, depth = position643, tokenIndex643, depth643
							{
								add(ruleAction92, position)
							}
						}
					l643:
						goto l641
					l642:
						position, tokenIndex, depth = position642, tokenIndex642, depth642
					}
				}
			l504:
				depth--
				add(ruleIDENTIFIER, position503)
			}
			return true
		l502:
			position, tokenIndex, depth = position502, tokenIndex502, depth502
			return false
		},
		/* 38 TIMESTAMP <- <((_ <(NUMBER ([a-z] / [A-Z])*)>) / (_ STRING) / (_ <(('n' / 'N') ('o' / 'O') ('w' / 'W'))> KEY))> */
		nil,
		/* 39 ID_SEGMENT <- <(ID_START ID_CONT*)> */
		func() bool {
			position647, tokenIndex647, depth647 := position, tokenIndex, depth
			{
				position648 := position
				depth++
				if !_rules[ruleID_START]() {
					goto l647
				}
			l649:
				{
					position650, tokenIndex650, depth650 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l650
					}
					goto l649
				l650:
					position, tokenIndex, depth = position650, tokenIndex650, depth650
				}
				depth--
				add(ruleID_SEGMENT, position648)
			}
			return true
		l647:
			position, tokenIndex, depth = position647, tokenIndex647, depth647
			return false
		},
		/* 40 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position651, tokenIndex651, depth651 := position, tokenIndex, depth
			{
				position652 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l651
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l651
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l651
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position652)
			}
			return true
		l651:
			position, tokenIndex, depth = position651, tokenIndex651, depth651
			return false
		},
		/* 41 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position654, tokenIndex654, depth654 := position, tokenIndex, depth
			{
				position655 := position
				depth++
				{
					position656, tokenIndex656, depth656 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l657
					}
					goto l656
				l657:
					position, tokenIndex, depth = position656, tokenIndex656, depth656
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l654
					}
					position++
				}
			l656:
				depth--
				add(ruleID_CONT, position655)
			}
			return true
		l654:
			position, tokenIndex, depth = position654, tokenIndex654, depth654
			return false
		},
		/* 42 PROPERTY_KEY <- <((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> KEY ((_ (('b' / 'B') ('y' / 'Y')) KEY) / Action93))) | (&('R' | 'r') (<(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))> KEY)) | (&('T' | 't') (<(('t' / 'T') ('o' / 'O'))> KEY)) | (&('F' | 'f') (<(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))> KEY)))> */
		func() bool {
			position658, tokenIndex658, depth658 := position, tokenIndex, depth
			{
				position659 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position661 := position
							depth++
							{
								position662, tokenIndex662, depth662 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l663
								}
								position++
								goto l662
							l663:
								position, tokenIndex, depth = position662, tokenIndex662, depth662
								if buffer[position] != rune('S') {
									goto l658
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
									goto l658
								}
								position++
							}
						l664:
							{
								position666, tokenIndex666, depth666 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l667
								}
								position++
								goto l666
							l667:
								position, tokenIndex, depth = position666, tokenIndex666, depth666
								if buffer[position] != rune('M') {
									goto l658
								}
								position++
							}
						l666:
							{
								position668, tokenIndex668, depth668 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l669
								}
								position++
								goto l668
							l669:
								position, tokenIndex, depth = position668, tokenIndex668, depth668
								if buffer[position] != rune('P') {
									goto l658
								}
								position++
							}
						l668:
							{
								position670, tokenIndex670, depth670 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l671
								}
								position++
								goto l670
							l671:
								position, tokenIndex, depth = position670, tokenIndex670, depth670
								if buffer[position] != rune('L') {
									goto l658
								}
								position++
							}
						l670:
							{
								position672, tokenIndex672, depth672 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l673
								}
								position++
								goto l672
							l673:
								position, tokenIndex, depth = position672, tokenIndex672, depth672
								if buffer[position] != rune('E') {
									goto l658
								}
								position++
							}
						l672:
							depth--
							add(rulePegText, position661)
						}
						if !_rules[ruleKEY]() {
							goto l658
						}
						{
							position674, tokenIndex674, depth674 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l675
							}
							{
								position676, tokenIndex676, depth676 := position, tokenIndex, depth
								if buffer[position] != rune('b') {
									goto l677
								}
								position++
								goto l676
							l677:
								position, tokenIndex, depth = position676, tokenIndex676, depth676
								if buffer[position] != rune('B') {
									goto l675
								}
								position++
							}
						l676:
							{
								position678, tokenIndex678, depth678 := position, tokenIndex, depth
								if buffer[position] != rune('y') {
									goto l679
								}
								position++
								goto l678
							l679:
								position, tokenIndex, depth = position678, tokenIndex678, depth678
								if buffer[position] != rune('Y') {
									goto l675
								}
								position++
							}
						l678:
							if !_rules[ruleKEY]() {
								goto l675
							}
							goto l674
						l675:
							position, tokenIndex, depth = position674, tokenIndex674, depth674
							{
								add(ruleAction93, position)
							}
						}
					l674:
						break
					case 'R', 'r':
						{
							position681 := position
							depth++
							{
								position682, tokenIndex682, depth682 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l683
								}
								position++
								goto l682
							l683:
								position, tokenIndex, depth = position682, tokenIndex682, depth682
								if buffer[position] != rune('R') {
									goto l658
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
									goto l658
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
									goto l658
								}
								position++
							}
						l686:
							{
								position688, tokenIndex688, depth688 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l689
								}
								position++
								goto l688
							l689:
								position, tokenIndex, depth = position688, tokenIndex688, depth688
								if buffer[position] != rune('O') {
									goto l658
								}
								position++
							}
						l688:
							{
								position690, tokenIndex690, depth690 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l691
								}
								position++
								goto l690
							l691:
								position, tokenIndex, depth = position690, tokenIndex690, depth690
								if buffer[position] != rune('L') {
									goto l658
								}
								position++
							}
						l690:
							{
								position692, tokenIndex692, depth692 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l693
								}
								position++
								goto l692
							l693:
								position, tokenIndex, depth = position692, tokenIndex692, depth692
								if buffer[position] != rune('U') {
									goto l658
								}
								position++
							}
						l692:
							{
								position694, tokenIndex694, depth694 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l695
								}
								position++
								goto l694
							l695:
								position, tokenIndex, depth = position694, tokenIndex694, depth694
								if buffer[position] != rune('T') {
									goto l658
								}
								position++
							}
						l694:
							{
								position696, tokenIndex696, depth696 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l697
								}
								position++
								goto l696
							l697:
								position, tokenIndex, depth = position696, tokenIndex696, depth696
								if buffer[position] != rune('I') {
									goto l658
								}
								position++
							}
						l696:
							{
								position698, tokenIndex698, depth698 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l699
								}
								position++
								goto l698
							l699:
								position, tokenIndex, depth = position698, tokenIndex698, depth698
								if buffer[position] != rune('O') {
									goto l658
								}
								position++
							}
						l698:
							{
								position700, tokenIndex700, depth700 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l701
								}
								position++
								goto l700
							l701:
								position, tokenIndex, depth = position700, tokenIndex700, depth700
								if buffer[position] != rune('N') {
									goto l658
								}
								position++
							}
						l700:
							depth--
							add(rulePegText, position681)
						}
						if !_rules[ruleKEY]() {
							goto l658
						}
						break
					case 'T', 't':
						{
							position702 := position
							depth++
							{
								position703, tokenIndex703, depth703 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l704
								}
								position++
								goto l703
							l704:
								position, tokenIndex, depth = position703, tokenIndex703, depth703
								if buffer[position] != rune('T') {
									goto l658
								}
								position++
							}
						l703:
							{
								position705, tokenIndex705, depth705 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l706
								}
								position++
								goto l705
							l706:
								position, tokenIndex, depth = position705, tokenIndex705, depth705
								if buffer[position] != rune('O') {
									goto l658
								}
								position++
							}
						l705:
							depth--
							add(rulePegText, position702)
						}
						if !_rules[ruleKEY]() {
							goto l658
						}
						break
					default:
						{
							position707 := position
							depth++
							{
								position708, tokenIndex708, depth708 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l709
								}
								position++
								goto l708
							l709:
								position, tokenIndex, depth = position708, tokenIndex708, depth708
								if buffer[position] != rune('F') {
									goto l658
								}
								position++
							}
						l708:
							{
								position710, tokenIndex710, depth710 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l711
								}
								position++
								goto l710
							l711:
								position, tokenIndex, depth = position710, tokenIndex710, depth710
								if buffer[position] != rune('R') {
									goto l658
								}
								position++
							}
						l710:
							{
								position712, tokenIndex712, depth712 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l713
								}
								position++
								goto l712
							l713:
								position, tokenIndex, depth = position712, tokenIndex712, depth712
								if buffer[position] != rune('O') {
									goto l658
								}
								position++
							}
						l712:
							{
								position714, tokenIndex714, depth714 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l715
								}
								position++
								goto l714
							l715:
								position, tokenIndex, depth = position714, tokenIndex714, depth714
								if buffer[position] != rune('M') {
									goto l658
								}
								position++
							}
						l714:
							depth--
							add(rulePegText, position707)
						}
						if !_rules[ruleKEY]() {
							goto l658
						}
						break
					}
				}

				depth--
				add(rulePROPERTY_KEY, position659)
			}
			return true
		l658:
			position, tokenIndex, depth = position658, tokenIndex658, depth658
			return false
		},
		/* 43 PROPERTY_VALUE <- <TIMESTAMP> */
		nil,
		/* 44 KEYWORD <- <((('a' / 'A') ('l' / 'L') ('l' / 'L')) / (('a' / 'A') ('n' / 'N') ('d' / 'D')) / (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H')) / (('s' / 'S') ('e' / 'E') ('l' / 'L') ('e' / 'E') ('c' / 'C') ('t' / 'T')) / ((&('M' | 'm') (('m' / 'M') ('e' / 'E') ('t' / 'T') ('r' / 'R') ('i' / 'I') ('c' / 'C') ('s' / 'S'))) | (&('W' | 'w') (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E'))) | (&('O' | 'o') (('o' / 'O') ('r' / 'R'))) | (&('N' | 'n') (('n' / 'N') ('o' / 'O') ('t' / 'T'))) | (&('I' | 'i') (('i' / 'I') ('n' / 'N'))) | (&('C' | 'c') (('c' / 'C') ('o' / 'O') ('l' / 'L') ('l' / 'L') ('a' / 'A') ('p' / 'P') ('s' / 'S') ('e' / 'E'))) | (&('G' | 'g') (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P'))) | (&('D' | 'd') (('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C') ('r' / 'R') ('i' / 'I') ('b' / 'B') ('e' / 'E'))) | (&('B' | 'b') (('b' / 'B') ('y' / 'Y'))) | (&('A' | 'a') (('a' / 'A') ('s' / 'S'))) | (&('F' | 'R' | 'S' | 'T' | 'f' | 'r' | 's' | 't') PROPERTY_KEY)))> */
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
			position726, tokenIndex726, depth726 := position, tokenIndex, depth
			{
				position727 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l726
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position727)
			}
			return true
		l726:
			position, tokenIndex, depth = position726, tokenIndex726, depth726
			return false
		},
		/* 54 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position728, tokenIndex728, depth728 := position, tokenIndex, depth
			{
				position729 := position
				depth++
				if buffer[position] != rune('"') {
					goto l728
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position729)
			}
			return true
		l728:
			position, tokenIndex, depth = position728, tokenIndex728, depth728
			return false
		},
		/* 55 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> (QUOTE_SINGLE / Action94)) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> (QUOTE_DOUBLE / Action95)))> */
		func() bool {
			position730, tokenIndex730, depth730 := position, tokenIndex, depth
			{
				position731 := position
				depth++
				{
					position732, tokenIndex732, depth732 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l733
					}
					{
						position734 := position
						depth++
					l735:
						{
							position736, tokenIndex736, depth736 := position, tokenIndex, depth
							{
								position737, tokenIndex737, depth737 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l737
								}
								goto l736
							l737:
								position, tokenIndex, depth = position737, tokenIndex737, depth737
							}
							if !_rules[ruleCHAR]() {
								goto l736
							}
							goto l735
						l736:
							position, tokenIndex, depth = position736, tokenIndex736, depth736
						}
						depth--
						add(rulePegText, position734)
					}
					{
						position738, tokenIndex738, depth738 := position, tokenIndex, depth
						if !_rules[ruleQUOTE_SINGLE]() {
							goto l739
						}
						goto l738
					l739:
						position, tokenIndex, depth = position738, tokenIndex738, depth738
						{
							add(ruleAction94, position)
						}
					}
				l738:
					goto l732
				l733:
					position, tokenIndex, depth = position732, tokenIndex732, depth732
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l730
					}
					{
						position741 := position
						depth++
					l742:
						{
							position743, tokenIndex743, depth743 := position, tokenIndex, depth
							{
								position744, tokenIndex744, depth744 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
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
						if !_rules[ruleQUOTE_DOUBLE]() {
							goto l746
						}
						goto l745
					l746:
						position, tokenIndex, depth = position745, tokenIndex745, depth745
						{
							add(ruleAction95, position)
						}
					}
				l745:
				}
			l732:
				depth--
				add(ruleSTRING, position731)
			}
			return true
		l730:
			position, tokenIndex, depth = position730, tokenIndex730, depth730
			return false
		},
		/* 56 CHAR <- <(('\\' ((&('"') (QUOTE_DOUBLE / Action96)) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position748, tokenIndex748, depth748 := position, tokenIndex, depth
			{
				position749 := position
				depth++
				{
					position750, tokenIndex750, depth750 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l751
					}
					position++
					{
						switch buffer[position] {
						case '"':
							{
								position753, tokenIndex753, depth753 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l754
								}
								goto l753
							l754:
								position, tokenIndex, depth = position753, tokenIndex753, depth753
								{
									add(ruleAction96, position)
								}
							}
						l753:
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l751
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l751
							}
							break
						}
					}

					goto l750
				l751:
					position, tokenIndex, depth = position750, tokenIndex750, depth750
					{
						position756, tokenIndex756, depth756 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l756
						}
						goto l748
					l756:
						position, tokenIndex, depth = position756, tokenIndex756, depth756
					}
					if !matchDot() {
						goto l748
					}
				}
			l750:
				depth--
				add(ruleCHAR, position749)
			}
			return true
		l748:
			position, tokenIndex, depth = position748, tokenIndex748, depth748
			return false
		},
		/* 57 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position757, tokenIndex757, depth757 := position, tokenIndex, depth
			{
				position758 := position
				depth++
				{
					position759, tokenIndex759, depth759 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l760
					}
					position++
					goto l759
				l760:
					position, tokenIndex, depth = position759, tokenIndex759, depth759
					if buffer[position] != rune('\\') {
						goto l757
					}
					position++
				}
			l759:
				depth--
				add(ruleESCAPE_CLASS, position758)
			}
			return true
		l757:
			position, tokenIndex, depth = position757, tokenIndex757, depth757
			return false
		},
		/* 58 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position761, tokenIndex761, depth761 := position, tokenIndex, depth
			{
				position762 := position
				depth++
				{
					position763 := position
					depth++
					{
						position764, tokenIndex764, depth764 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l764
						}
						position++
						goto l765
					l764:
						position, tokenIndex, depth = position764, tokenIndex764, depth764
					}
				l765:
					{
						position766 := position
						depth++
						{
							position767, tokenIndex767, depth767 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l768
							}
							position++
							goto l767
						l768:
							position, tokenIndex, depth = position767, tokenIndex767, depth767
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l761
							}
							position++
						l769:
							{
								position770, tokenIndex770, depth770 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l770
								}
								position++
								goto l769
							l770:
								position, tokenIndex, depth = position770, tokenIndex770, depth770
							}
						}
					l767:
						depth--
						add(ruleNUMBER_NATURAL, position766)
					}
					depth--
					add(ruleNUMBER_INTEGER, position763)
				}
				{
					position771, tokenIndex771, depth771 := position, tokenIndex, depth
					{
						position773 := position
						depth++
						if buffer[position] != rune('.') {
							goto l771
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l771
						}
						position++
					l774:
						{
							position775, tokenIndex775, depth775 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l775
							}
							position++
							goto l774
						l775:
							position, tokenIndex, depth = position775, tokenIndex775, depth775
						}
						depth--
						add(ruleNUMBER_FRACTION, position773)
					}
					goto l772
				l771:
					position, tokenIndex, depth = position771, tokenIndex771, depth771
				}
			l772:
				{
					position776, tokenIndex776, depth776 := position, tokenIndex, depth
					{
						position778 := position
						depth++
						{
							position779, tokenIndex779, depth779 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l780
							}
							position++
							goto l779
						l780:
							position, tokenIndex, depth = position779, tokenIndex779, depth779
							if buffer[position] != rune('E') {
								goto l776
							}
							position++
						}
					l779:
						{
							position781, tokenIndex781, depth781 := position, tokenIndex, depth
							{
								position783, tokenIndex783, depth783 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l784
								}
								position++
								goto l783
							l784:
								position, tokenIndex, depth = position783, tokenIndex783, depth783
								if buffer[position] != rune('-') {
									goto l781
								}
								position++
							}
						l783:
							goto l782
						l781:
							position, tokenIndex, depth = position781, tokenIndex781, depth781
						}
					l782:
						{
							position785, tokenIndex785, depth785 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l786
							}
							position++
						l787:
							{
								position788, tokenIndex788, depth788 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l788
								}
								position++
								goto l787
							l788:
								position, tokenIndex, depth = position788, tokenIndex788, depth788
							}
							goto l785
						l786:
							position, tokenIndex, depth = position785, tokenIndex785, depth785
							{
								add(ruleAction97, position)
							}
						}
					l785:
						depth--
						add(ruleNUMBER_EXP, position778)
					}
					goto l777
				l776:
					position, tokenIndex, depth = position776, tokenIndex776, depth776
				}
			l777:
				depth--
				add(ruleNUMBER, position762)
			}
			return true
		l761:
			position, tokenIndex, depth = position761, tokenIndex761, depth761
			return false
		},
		/* 59 NUMBER_NATURAL <- <('0' / ([1-9] [0-9]*))> */
		nil,
		/* 60 NUMBER_FRACTION <- <('.' [0-9]+)> */
		nil,
		/* 61 NUMBER_INTEGER <- <('-'? NUMBER_NATURAL)> */
		nil,
		/* 62 NUMBER_EXP <- <(('e' / 'E') ('+' / '-')? ([0-9]+ / Action97))> */
		nil,
		/* 63 DURATION <- <(NUMBER [a-z]+ KEY)> */
		nil,
		/* 64 PAREN_OPEN <- <'('> */
		func() bool {
			position795, tokenIndex795, depth795 := position, tokenIndex, depth
			{
				position796 := position
				depth++
				if buffer[position] != rune('(') {
					goto l795
				}
				position++
				depth--
				add(rulePAREN_OPEN, position796)
			}
			return true
		l795:
			position, tokenIndex, depth = position795, tokenIndex795, depth795
			return false
		},
		/* 65 PAREN_CLOSE <- <')'> */
		func() bool {
			position797, tokenIndex797, depth797 := position, tokenIndex, depth
			{
				position798 := position
				depth++
				if buffer[position] != rune(')') {
					goto l797
				}
				position++
				depth--
				add(rulePAREN_CLOSE, position798)
			}
			return true
		l797:
			position, tokenIndex, depth = position797, tokenIndex797, depth797
			return false
		},
		/* 66 COMMA <- <','> */
		func() bool {
			position799, tokenIndex799, depth799 := position, tokenIndex, depth
			{
				position800 := position
				depth++
				if buffer[position] != rune(',') {
					goto l799
				}
				position++
				depth--
				add(ruleCOMMA, position800)
			}
			return true
		l799:
			position, tokenIndex, depth = position799, tokenIndex799, depth799
			return false
		},
		/* 67 _ <- <((&('/') COMMENT_BLOCK) | (&('-') COMMENT_TRAIL) | (&('\t' | '\n' | ' ') SPACE))*> */
		func() bool {
			{
				position802 := position
				depth++
			l803:
				{
					position804, tokenIndex804, depth804 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '/':
							{
								position806 := position
								depth++
								if buffer[position] != rune('/') {
									goto l804
								}
								position++
								if buffer[position] != rune('*') {
									goto l804
								}
								position++
							l807:
								{
									position808, tokenIndex808, depth808 := position, tokenIndex, depth
									{
										position809, tokenIndex809, depth809 := position, tokenIndex, depth
										if buffer[position] != rune('*') {
											goto l809
										}
										position++
										if buffer[position] != rune('/') {
											goto l809
										}
										position++
										goto l808
									l809:
										position, tokenIndex, depth = position809, tokenIndex809, depth809
									}
									if !matchDot() {
										goto l808
									}
									goto l807
								l808:
									position, tokenIndex, depth = position808, tokenIndex808, depth808
								}
								if buffer[position] != rune('*') {
									goto l804
								}
								position++
								if buffer[position] != rune('/') {
									goto l804
								}
								position++
								depth--
								add(ruleCOMMENT_BLOCK, position806)
							}
							break
						case '-':
							{
								position810 := position
								depth++
								if buffer[position] != rune('-') {
									goto l804
								}
								position++
								if buffer[position] != rune('-') {
									goto l804
								}
								position++
							l811:
								{
									position812, tokenIndex812, depth812 := position, tokenIndex, depth
									{
										position813, tokenIndex813, depth813 := position, tokenIndex, depth
										if buffer[position] != rune('\n') {
											goto l813
										}
										position++
										goto l812
									l813:
										position, tokenIndex, depth = position813, tokenIndex813, depth813
									}
									if !matchDot() {
										goto l812
									}
									goto l811
								l812:
									position, tokenIndex, depth = position812, tokenIndex812, depth812
								}
								depth--
								add(ruleCOMMENT_TRAIL, position810)
							}
							break
						default:
							{
								position814 := position
								depth++
								{
									switch buffer[position] {
									case '\t':
										if buffer[position] != rune('\t') {
											goto l804
										}
										position++
										break
									case '\n':
										if buffer[position] != rune('\n') {
											goto l804
										}
										position++
										break
									default:
										if buffer[position] != rune(' ') {
											goto l804
										}
										position++
										break
									}
								}

								depth--
								add(ruleSPACE, position814)
							}
							break
						}
					}

					goto l803
				l804:
					position, tokenIndex, depth = position804, tokenIndex804, depth804
				}
				depth--
				add(rule_, position802)
			}
			return true
		},
		/* 68 COMMENT_TRAIL <- <('-' '-' (!'\n' .)*)> */
		nil,
		/* 69 COMMENT_BLOCK <- <('/' '*' (!('*' '/') .)* ('*' '/'))> */
		nil,
		/* 70 KEY <- <!ID_CONT> */
		func() bool {
			position818, tokenIndex818, depth818 := position, tokenIndex, depth
			{
				position819 := position
				depth++
				{
					position820, tokenIndex820, depth820 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l820
					}
					goto l818
				l820:
					position, tokenIndex, depth = position820, tokenIndex820, depth820
				}
				depth--
				add(ruleKEY, position819)
			}
			return true
		l818:
			position, tokenIndex, depth = position818, tokenIndex818, depth818
			return false
		},
		/* 71 SPACE <- <((&('\t') '\t') | (&('\n') '\n') | (&(' ') ' '))> */
		nil,
		/* 73 Action0 <- <{
		   p.makeSelect()
		 }> */
		nil,
		/* 74 Action1 <- <{ p.makeDescribeAll() }> */
		nil,
		/* 75 Action2 <- <{ p.addNullMatchClause() }> */
		nil,
		/* 76 Action3 <- <{ p.errorHere(token.begin,`expected string literal to follow keyword "match"`) }> */
		nil,
		/* 77 Action4 <- <{ p.addMatchClause() }> */
		nil,
		/* 78 Action5 <- <{ p.errorHere(token.begin,`expected "where" to follow keyword "metrics" in "describe metrics" command`) }> */
		nil,
		/* 79 Action6 <- <{ p.errorHere(token.begin,`expected tag key to follow keyword "where" in "describe metrics" command`) }> */
		nil,
		/* 80 Action7 <- <{ p.errorHere(token.begin,`expected "=" to follow keyword "where" in "describe metrics" command`) }> */
		nil,
		/* 81 Action8 <- <{ p.errorHere(token.begin,`expected string literal to follow "=" in "describe metrics" command`) }> */
		nil,
		/* 82 Action9 <- <{ p.makeDescribeMetrics() }> */
		nil,
		nil,
		/* 84 Action10 <- <{ p.pushString(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 85 Action11 <- <{ p.errorHere(token.begin,`expected metric name to follow "describe" in "describe" command`) }> */
		nil,
		/* 86 Action12 <- <{ p.makeDescribe() }> */
		nil,
		/* 87 Action13 <- <{ p.addEvaluationContext() }> */
		nil,
		/* 88 Action14 <- <{ p.addPropertyKey(buffer[begin:end])   }> */
		nil,
		/* 89 Action15 <- <{ p.addPropertyValue(buffer[begin:end]) }> */
		nil,
		/* 90 Action16 <- <{ p.errorHere(token.begin,`expected property value to follow property key`) }> */
		nil,
		/* 91 Action17 <- <{ p.insertPropertyKeyValue() }> */
		nil,
		/* 92 Action18 <- <{ p.checkPropertyClause() }> */
		nil,
		/* 93 Action19 <- <{ p.addNullPredicate() }> */
		nil,
		/* 94 Action20 <- <{ p.addExpressionList() }> */
		nil,
		/* 95 Action21 <- <{ p.appendExpression() }> */
		nil,
		/* 96 Action22 <- <{ p.errorHere(token.begin,`expected expression to follow ","`) }> */
		nil,
		/* 97 Action23 <- <{ p.appendExpression() }> */
		nil,
		/* 98 Action24 <- <{ p.addOperatorLiteral("+") }> */
		nil,
		/* 99 Action25 <- <{ p.addOperatorLiteral("-") }> */
		nil,
		/* 100 Action26 <- <{ p.errorHere(token.begin,`expected expression to follow operator "+" or "-"`) }> */
		nil,
		/* 101 Action27 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 102 Action28 <- <{ p.addOperatorLiteral("/") }> */
		nil,
		/* 103 Action29 <- <{ p.addOperatorLiteral("*") }> */
		nil,
		/* 104 Action30 <- <{ p.errorHere(token.begin,`expected expression to follow operator "*" or "/"`) }> */
		nil,
		/* 105 Action31 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 106 Action32 <- <{ p.errorHere(token.begin,`expected function name to follow pipe "|"`) }> */
		nil,
		/* 107 Action33 <- <{ p.pushString(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 108 Action34 <- <{p.addExpressionList()}> */
		nil,
		/* 109 Action35 <- <{ p.errorHere(token.begin,`expected ")" to close "(" opened in pipe function call`) }> */
		nil,
		/* 110 Action36 <- <{
		   p.addExpressionList()
		   p.addGroupBy()
		 }> */
		nil,
		/* 111 Action37 <- <{ p.addPipeExpression() }> */
		nil,
		/* 112 Action38 <- <{ p.errorHere(token.begin,`expected expression to follow "("`) }> */
		nil,
		/* 113 Action39 <- <{ p.errorHere(token.begin,`expected ")" to close "("`) }> */
		nil,
		/* 114 Action40 <- <{ p.addDurationNode(text) }> */
		nil,
		/* 115 Action41 <- <{ p.addNumberNode(buffer[begin:end]) }> */
		nil,
		/* 116 Action42 <- <{ p.addStringNode(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 117 Action43 <- <{ p.errorHere(token.begin,`expected "}> */
		nil,
		/* 118 Action44 <- <{" opened for annotation`) }> */
		nil,
		/* 119 Action45 <- <{ p.addAnnotationExpression(buffer[begin:end]) }> */
		nil,
		/* 120 Action46 <- <{ p.addGroupBy() }> */
		nil,
		/* 121 Action47 <- <{ p.pushString(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 122 Action48 <- <{ p.errorHere(token.begin,`expected ")" to close "(" opened by function call`) }> */
		nil,
		/* 123 Action49 <- <{ p.addFunctionInvocation() }> */
		nil,
		/* 124 Action50 <- <{ p.pushString(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 125 Action51 <- <{ p.errorHere(token.begin,`expected predicate to follow "[" after metric`) }> */
		nil,
		/* 126 Action52 <- <{ p.errorHere(token.begin,`expected "]" to close "[" opened to apply predicate`) }> */
		nil,
		/* 127 Action53 <- <{ p.addNullPredicate() }> */
		nil,
		/* 128 Action54 <- <{ p.addMetricExpression() }> */
		nil,
		/* 129 Action55 <- <{ p.errorHere(token.begin,`expected keyword "by" to follow keyword "group" in "group by" clause`) }> */
		nil,
		/* 130 Action56 <- <{ p.errorHere(token.begin,`expected tag key identifier to follow "group by" keywords in "group by" clause`) }> */
		nil,
		/* 131 Action57 <- <{ p.appendGroupBy(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 132 Action58 <- <{ p.errorHere(token.begin,`expected tag key identifier to follow "," in "group by" clause`) }> */
		nil,
		/* 133 Action59 <- <{ p.appendGroupBy(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 134 Action60 <- <{ p.errorHere(token.begin,`expected keyword "by" to follow keyword "collapse" in "collapse by" clause`) }> */
		nil,
		/* 135 Action61 <- <{ p.errorHere(token.begin,`expected tag key identifier to follow "collapse by" keywords in "collapse by" clause`) }> */
		nil,
		/* 136 Action62 <- <{ p.appendCollapseBy(unescapeLiteral(text)) }> */
		nil,
		/* 137 Action63 <- <{ p.errorHere(token.begin,`expected tag key identifier to follow "," in "collapse by" clause`) }> */
		nil,
		/* 138 Action64 <- <{ p.appendCollapseBy(unescapeLiteral(text)) }> */
		nil,
		/* 139 Action65 <- <{ p.errorHere(token.begin,`expected predicate to follow "where" keyword`) }> */
		nil,
		/* 140 Action66 <- <{ p.errorHere(token.begin,`expected predicate to follow "or" operator`) }> */
		nil,
		/* 141 Action67 <- <{ p.addOrPredicate() }> */
		nil,
		/* 142 Action68 <- <{ p.errorHere(token.begin,`expected predicate to follow "and" operator`) }> */
		nil,
		/* 143 Action69 <- <{ p.addAndPredicate() }> */
		nil,
		/* 144 Action70 <- <{ p.errorHere(token.begin,`expected predicate to follow "not" operator`) }> */
		nil,
		/* 145 Action71 <- <{ p.addNotPredicate() }> */
		nil,
		/* 146 Action72 <- <{ p.errorHere(token.begin,`expected predicate to follow "("`) }> */
		nil,
		/* 147 Action73 <- <{ p.errorHere(token.begin,`expected ")" to close "(" opened in predicate`) }> */
		nil,
		/* 148 Action74 <- <{ p.errorHere(token.begin,`expected string literal to follow "="`) }> */
		nil,
		/* 149 Action75 <- <{ p.addLiteralMatcher() }> */
		nil,
		/* 150 Action76 <- <{ p.errorHere(token.begin,`expected string literal to follow "!="`) }> */
		nil,
		/* 151 Action77 <- <{ p.addLiteralMatcher() }> */
		nil,
		/* 152 Action78 <- <{ p.addNotPredicate() }> */
		nil,
		/* 153 Action79 <- <{ p.errorHere(token.begin,`expected regex string literal to follow "match"`) }> */
		nil,
		/* 154 Action80 <- <{ p.addRegexMatcher() }> */
		nil,
		/* 155 Action81 <- <{ p.errorHere(token.begin,`expected string literal list to follow "in" keyword`) }> */
		nil,
		/* 156 Action82 <- <{ p.addListMatcher() }> */
		nil,
		/* 157 Action83 <- <{ p.errorHere(token.begin,`expected "=", "!=", "match", or "in" to follow tag key in predicate`) }> */
		nil,
		/* 158 Action84 <- <{ p.pushString(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 159 Action85 <- <{ p.addLiteralList() }> */
		nil,
		/* 160 Action86 <- <{ p.errorHere(token.begin,`expected string literal to follow "(" in literal list`) }> */
		nil,
		/* 161 Action87 <- <{ p.errorHere(token.begin,`expected string literal to follow "," in literal list`) }> */
		nil,
		/* 162 Action88 <- <{ p.errorHere(token.begin,`expected ")" to close "(" for literal list`) }> */
		nil,
		/* 163 Action89 <- <{ p.appendLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 164 Action90 <- <{ p.addTagLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 165 Action91 <- <{ p.errorHere(token.begin,"expected \"`\" to end identifier") }> */
		nil,
		/* 166 Action92 <- <{ p.errorHere(token.begin,`expected identifier segment to follow "."`) }> */
		nil,
		/* 167 Action93 <- <{ p.errorHere(token.begin,`expected keyword "by" to follow keyword "sample"`) }> */
		nil,
		/* 168 Action94 <- <{ p.errorHere(token.begin,`expected "'" to close string`) }> */
		nil,
		/* 169 Action95 <- <{ p.errorHere(token.begin,`expected '"' to close string`) }> */
		nil,
		/* 170 Action96 <- <{ p.errorHere(token.begin, "expected \"\\\", \"'\", \"`\", or '\"' to follow \"\\\" in string literal") }> */
		nil,
		/* 171 Action97 <- <{ p.errorHere(token.begin,`expected exponent`) }> */
		nil,
	}
	p.rules = _rules
}
