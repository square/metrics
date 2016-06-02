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
	ruleAction98

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
	"Action98",

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
	rules  [173]func() bool
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
			p.errorHere(token.begin, `expected expression list to follow "(" in function call`)
		case ruleAction49:
			p.errorHere(token.begin, `expected ")" to close "(" opened by function call`)
		case ruleAction50:
			p.addFunctionInvocation()
		case ruleAction51:
			p.pushString(unescapeLiteral(buffer[begin:end]))
		case ruleAction52:
			p.errorHere(token.begin, `expected predicate to follow "[" after metric`)
		case ruleAction53:
			p.errorHere(token.begin, `expected "]" to close "[" opened to apply predicate`)
		case ruleAction54:
			p.addNullPredicate()
		case ruleAction55:
			p.addMetricExpression()
		case ruleAction56:
			p.errorHere(token.begin, `expected keyword "by" to follow keyword "group" in "group by" clause`)
		case ruleAction57:
			p.errorHere(token.begin, `expected tag key identifier to follow "group by" keywords in "group by" clause`)
		case ruleAction58:
			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		case ruleAction59:
			p.errorHere(token.begin, `expected tag key identifier to follow "," in "group by" clause`)
		case ruleAction60:
			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		case ruleAction61:
			p.errorHere(token.begin, `expected keyword "by" to follow keyword "collapse" in "collapse by" clause`)
		case ruleAction62:
			p.errorHere(token.begin, `expected tag key identifier to follow "collapse by" keywords in "collapse by" clause`)
		case ruleAction63:
			p.appendCollapseBy(unescapeLiteral(text))
		case ruleAction64:
			p.errorHere(token.begin, `expected tag key identifier to follow "," in "collapse by" clause`)
		case ruleAction65:
			p.appendCollapseBy(unescapeLiteral(text))
		case ruleAction66:
			p.errorHere(token.begin, `expected predicate to follow "where" keyword`)
		case ruleAction67:
			p.errorHere(token.begin, `expected predicate to follow "or" operator`)
		case ruleAction68:
			p.addOrPredicate()
		case ruleAction69:
			p.errorHere(token.begin, `expected predicate to follow "and" operator`)
		case ruleAction70:
			p.addAndPredicate()
		case ruleAction71:
			p.errorHere(token.begin, `expected predicate to follow "not" operator`)
		case ruleAction72:
			p.addNotPredicate()
		case ruleAction73:
			p.errorHere(token.begin, `expected predicate to follow "("`)
		case ruleAction74:
			p.errorHere(token.begin, `expected ")" to close "(" opened in predicate`)
		case ruleAction75:
			p.errorHere(token.begin, `expected string literal to follow "="`)
		case ruleAction76:
			p.addLiteralMatcher()
		case ruleAction77:
			p.errorHere(token.begin, `expected string literal to follow "!="`)
		case ruleAction78:
			p.addLiteralMatcher()
		case ruleAction79:
			p.addNotPredicate()
		case ruleAction80:
			p.errorHere(token.begin, `expected regex string literal to follow "match"`)
		case ruleAction81:
			p.addRegexMatcher()
		case ruleAction82:
			p.errorHere(token.begin, `expected string literal list to follow "in" keyword`)
		case ruleAction83:
			p.addListMatcher()
		case ruleAction84:
			p.errorHere(token.begin, `expected "=", "!=", "match", or "in" to follow tag key in predicate`)
		case ruleAction85:
			p.pushString(unescapeLiteral(buffer[begin:end]))
		case ruleAction86:
			p.addLiteralList()
		case ruleAction87:
			p.errorHere(token.begin, `expected string literal to follow "(" in literal list`)
		case ruleAction88:
			p.errorHere(token.begin, `expected string literal to follow "," in literal list`)
		case ruleAction89:
			p.errorHere(token.begin, `expected ")" to close "(" for literal list`)
		case ruleAction90:
			p.appendLiteral(unescapeLiteral(buffer[begin:end]))
		case ruleAction91:
			p.addTagLiteral(unescapeLiteral(buffer[begin:end]))
		case ruleAction92:
			p.errorHere(token.begin, "expected \"`\" to end identifier")
		case ruleAction93:
			p.errorHere(token.begin, `expected identifier segment to follow "."`)
		case ruleAction94:
			p.errorHere(token.begin, `expected keyword "by" to follow keyword "sample"`)
		case ruleAction95:
			p.errorHere(token.begin, `expected "'" to close string`)
		case ruleAction96:
			p.errorHere(token.begin, `expected '"' to close string`)
		case ruleAction97:
			p.errorHere(token.begin, "expected \"\\\", \"'\", \"`\", or '\"' to follow \"\\\" in string literal")
		case ruleAction98:
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
								add(ruleAction66, position)
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
							{
								position239, tokenIndex239, depth239 := position, tokenIndex, depth
								if !_rules[ruleexpressionList]() {
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
							if !_rules[ruleoptionalGroupBy]() {
								goto l235
							}
							{
								position242, tokenIndex242, depth242 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l243
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l243
								}
								goto l242
							l243:
								position, tokenIndex, depth = position242, tokenIndex242, depth242
								{
									add(ruleAction49, position)
								}
							}
						l242:
							{
								add(ruleAction50, position)
							}
							depth--
							add(ruleexpression_function, position236)
						}
						goto l234
					l235:
						position, tokenIndex, depth = position234, tokenIndex234, depth234
						{
							position247 := position
							depth++
							if !_rules[rule_]() {
								goto l246
							}
							{
								position248 := position
								depth++
								if !_rules[ruleIDENTIFIER]() {
									goto l246
								}
								depth--
								add(rulePegText, position248)
							}
							{
								add(ruleAction51, position)
							}
							{
								position250, tokenIndex250, depth250 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l251
								}
								if buffer[position] != rune('[') {
									goto l251
								}
								position++
								{
									position252, tokenIndex252, depth252 := position, tokenIndex, depth
									if !_rules[rulepredicate_1]() {
										goto l253
									}
									goto l252
								l253:
									position, tokenIndex, depth = position252, tokenIndex252, depth252
									{
										add(ruleAction52, position)
									}
								}
							l252:
								{
									position255, tokenIndex255, depth255 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l256
									}
									if buffer[position] != rune(']') {
										goto l256
									}
									position++
									goto l255
								l256:
									position, tokenIndex, depth = position255, tokenIndex255, depth255
									{
										add(ruleAction53, position)
									}
								}
							l255:
								goto l250
							l251:
								position, tokenIndex, depth = position250, tokenIndex250, depth250
								{
									add(ruleAction54, position)
								}
							}
						l250:
							{
								add(ruleAction55, position)
							}
							depth--
							add(ruleexpression_metric, position247)
						}
						goto l234
					l246:
						position, tokenIndex, depth = position234, tokenIndex234, depth234
						if !_rules[rule_]() {
							goto l260
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l260
						}
						{
							position261, tokenIndex261, depth261 := position, tokenIndex, depth
							if !_rules[ruleexpression_start]() {
								goto l262
							}
							goto l261
						l262:
							position, tokenIndex, depth = position261, tokenIndex261, depth261
							{
								add(ruleAction38, position)
							}
						}
					l261:
						{
							position264, tokenIndex264, depth264 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l265
							}
							if !_rules[rulePAREN_CLOSE]() {
								goto l265
							}
							goto l264
						l265:
							position, tokenIndex, depth = position264, tokenIndex264, depth264
							{
								add(ruleAction39, position)
							}
						}
					l264:
						goto l234
					l260:
						position, tokenIndex, depth = position234, tokenIndex234, depth234
						if !_rules[rule_]() {
							goto l267
						}
						{
							position268 := position
							depth++
							{
								position269 := position
								depth++
								if !_rules[ruleNUMBER]() {
									goto l267
								}
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l267
								}
								position++
							l270:
								{
									position271, tokenIndex271, depth271 := position, tokenIndex, depth
									if c := buffer[position]; c < rune('a') || c > rune('z') {
										goto l271
									}
									position++
									goto l270
								l271:
									position, tokenIndex, depth = position271, tokenIndex271, depth271
								}
								if !_rules[ruleKEY]() {
									goto l267
								}
								depth--
								add(ruleDURATION, position269)
							}
							depth--
							add(rulePegText, position268)
						}
						{
							add(ruleAction40, position)
						}
						goto l234
					l267:
						position, tokenIndex, depth = position234, tokenIndex234, depth234
						if !_rules[rule_]() {
							goto l273
						}
						{
							position274 := position
							depth++
							if !_rules[ruleNUMBER]() {
								goto l273
							}
							depth--
							add(rulePegText, position274)
						}
						{
							add(ruleAction41, position)
						}
						goto l234
					l273:
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
				position280 := position
				depth++
				{
					position281, tokenIndex281, depth281 := position, tokenIndex, depth
					{
						position283 := position
						depth++
						if !_rules[rule_]() {
							goto l281
						}
						if buffer[position] != rune('{') {
							goto l281
						}
						position++
						{
							position284 := position
							depth++
						l285:
							{
								position286, tokenIndex286, depth286 := position, tokenIndex, depth
								{
									position287, tokenIndex287, depth287 := position, tokenIndex, depth
									if buffer[position] != rune('}') {
										goto l287
									}
									position++
									goto l286
								l287:
									position, tokenIndex, depth = position287, tokenIndex287, depth287
								}
								if !matchDot() {
									goto l286
								}
								goto l285
							l286:
								position, tokenIndex, depth = position286, tokenIndex286, depth286
							}
							depth--
							add(rulePegText, position284)
						}
						{
							position288, tokenIndex288, depth288 := position, tokenIndex, depth
							if buffer[position] != rune('}') {
								goto l289
							}
							position++
							goto l288
						l289:
							position, tokenIndex, depth = position288, tokenIndex288, depth288
							{
								add(ruleAction43, position)
							}
							if buffer[position] != rune(' ') {
								goto l281
							}
							position++
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
									goto l281
								}
								position++
							}
						l291:
							{
								position293, tokenIndex293, depth293 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l294
								}
								position++
								goto l293
							l294:
								position, tokenIndex, depth = position293, tokenIndex293, depth293
								if buffer[position] != rune('O') {
									goto l281
								}
								position++
							}
						l293:
							if buffer[position] != rune(' ') {
								goto l281
							}
							position++
							{
								position295, tokenIndex295, depth295 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l296
								}
								position++
								goto l295
							l296:
								position, tokenIndex, depth = position295, tokenIndex295, depth295
								if buffer[position] != rune('C') {
									goto l281
								}
								position++
							}
						l295:
							{
								position297, tokenIndex297, depth297 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l298
								}
								position++
								goto l297
							l298:
								position, tokenIndex, depth = position297, tokenIndex297, depth297
								if buffer[position] != rune('L') {
									goto l281
								}
								position++
							}
						l297:
							{
								position299, tokenIndex299, depth299 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l300
								}
								position++
								goto l299
							l300:
								position, tokenIndex, depth = position299, tokenIndex299, depth299
								if buffer[position] != rune('O') {
									goto l281
								}
								position++
							}
						l299:
							{
								position301, tokenIndex301, depth301 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l302
								}
								position++
								goto l301
							l302:
								position, tokenIndex, depth = position301, tokenIndex301, depth301
								if buffer[position] != rune('S') {
									goto l281
								}
								position++
							}
						l301:
							{
								position303, tokenIndex303, depth303 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l304
								}
								position++
								goto l303
							l304:
								position, tokenIndex, depth = position303, tokenIndex303, depth303
								if buffer[position] != rune('E') {
									goto l281
								}
								position++
							}
						l303:
							if buffer[position] != rune(' ') {
								goto l281
							}
							position++
							{
								add(ruleAction44, position)
							}
						}
					l288:
						{
							add(ruleAction45, position)
						}
						depth--
						add(ruleexpression_annotation_required, position283)
					}
					goto l282
				l281:
					position, tokenIndex, depth = position281, tokenIndex281, depth281
				}
			l282:
				depth--
				add(ruleexpression_annotation, position280)
			}
			return true
		},
		/* 20 optionalGroupBy <- <(Action46 (groupByClause / collapseByClause)?)> */
		func() bool {
			{
				position308 := position
				depth++
				{
					add(ruleAction46, position)
				}
				{
					position310, tokenIndex310, depth310 := position, tokenIndex, depth
					{
						position312, tokenIndex312, depth312 := position, tokenIndex, depth
						{
							position314 := position
							depth++
							if !_rules[rule_]() {
								goto l313
							}
							{
								position315, tokenIndex315, depth315 := position, tokenIndex, depth
								if buffer[position] != rune('g') {
									goto l316
								}
								position++
								goto l315
							l316:
								position, tokenIndex, depth = position315, tokenIndex315, depth315
								if buffer[position] != rune('G') {
									goto l313
								}
								position++
							}
						l315:
							{
								position317, tokenIndex317, depth317 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l318
								}
								position++
								goto l317
							l318:
								position, tokenIndex, depth = position317, tokenIndex317, depth317
								if buffer[position] != rune('R') {
									goto l313
								}
								position++
							}
						l317:
							{
								position319, tokenIndex319, depth319 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l320
								}
								position++
								goto l319
							l320:
								position, tokenIndex, depth = position319, tokenIndex319, depth319
								if buffer[position] != rune('O') {
									goto l313
								}
								position++
							}
						l319:
							{
								position321, tokenIndex321, depth321 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l322
								}
								position++
								goto l321
							l322:
								position, tokenIndex, depth = position321, tokenIndex321, depth321
								if buffer[position] != rune('U') {
									goto l313
								}
								position++
							}
						l321:
							{
								position323, tokenIndex323, depth323 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l324
								}
								position++
								goto l323
							l324:
								position, tokenIndex, depth = position323, tokenIndex323, depth323
								if buffer[position] != rune('P') {
									goto l313
								}
								position++
							}
						l323:
							if !_rules[ruleKEY]() {
								goto l313
							}
							{
								position325, tokenIndex325, depth325 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l326
								}
								{
									position327, tokenIndex327, depth327 := position, tokenIndex, depth
									if buffer[position] != rune('b') {
										goto l328
									}
									position++
									goto l327
								l328:
									position, tokenIndex, depth = position327, tokenIndex327, depth327
									if buffer[position] != rune('B') {
										goto l326
									}
									position++
								}
							l327:
								{
									position329, tokenIndex329, depth329 := position, tokenIndex, depth
									if buffer[position] != rune('y') {
										goto l330
									}
									position++
									goto l329
								l330:
									position, tokenIndex, depth = position329, tokenIndex329, depth329
									if buffer[position] != rune('Y') {
										goto l326
									}
									position++
								}
							l329:
								if !_rules[ruleKEY]() {
									goto l326
								}
								goto l325
							l326:
								position, tokenIndex, depth = position325, tokenIndex325, depth325
								{
									add(ruleAction56, position)
								}
							}
						l325:
							{
								position332, tokenIndex332, depth332 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l333
								}
								{
									position334 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l333
									}
									depth--
									add(rulePegText, position334)
								}
								goto l332
							l333:
								position, tokenIndex, depth = position332, tokenIndex332, depth332
								{
									add(ruleAction57, position)
								}
							}
						l332:
							{
								add(ruleAction58, position)
							}
						l337:
							{
								position338, tokenIndex338, depth338 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l338
								}
								if !_rules[ruleCOMMA]() {
									goto l338
								}
								{
									position339, tokenIndex339, depth339 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l340
									}
									{
										position341 := position
										depth++
										if !_rules[ruleCOLUMN_NAME]() {
											goto l340
										}
										depth--
										add(rulePegText, position341)
									}
									goto l339
								l340:
									position, tokenIndex, depth = position339, tokenIndex339, depth339
									{
										add(ruleAction59, position)
									}
								}
							l339:
								{
									add(ruleAction60, position)
								}
								goto l337
							l338:
								position, tokenIndex, depth = position338, tokenIndex338, depth338
							}
							depth--
							add(rulegroupByClause, position314)
						}
						goto l312
					l313:
						position, tokenIndex, depth = position312, tokenIndex312, depth312
						{
							position344 := position
							depth++
							if !_rules[rule_]() {
								goto l310
							}
							{
								position345, tokenIndex345, depth345 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l346
								}
								position++
								goto l345
							l346:
								position, tokenIndex, depth = position345, tokenIndex345, depth345
								if buffer[position] != rune('C') {
									goto l310
								}
								position++
							}
						l345:
							{
								position347, tokenIndex347, depth347 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l348
								}
								position++
								goto l347
							l348:
								position, tokenIndex, depth = position347, tokenIndex347, depth347
								if buffer[position] != rune('O') {
									goto l310
								}
								position++
							}
						l347:
							{
								position349, tokenIndex349, depth349 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l350
								}
								position++
								goto l349
							l350:
								position, tokenIndex, depth = position349, tokenIndex349, depth349
								if buffer[position] != rune('L') {
									goto l310
								}
								position++
							}
						l349:
							{
								position351, tokenIndex351, depth351 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l352
								}
								position++
								goto l351
							l352:
								position, tokenIndex, depth = position351, tokenIndex351, depth351
								if buffer[position] != rune('L') {
									goto l310
								}
								position++
							}
						l351:
							{
								position353, tokenIndex353, depth353 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l354
								}
								position++
								goto l353
							l354:
								position, tokenIndex, depth = position353, tokenIndex353, depth353
								if buffer[position] != rune('A') {
									goto l310
								}
								position++
							}
						l353:
							{
								position355, tokenIndex355, depth355 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l356
								}
								position++
								goto l355
							l356:
								position, tokenIndex, depth = position355, tokenIndex355, depth355
								if buffer[position] != rune('P') {
									goto l310
								}
								position++
							}
						l355:
							{
								position357, tokenIndex357, depth357 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l358
								}
								position++
								goto l357
							l358:
								position, tokenIndex, depth = position357, tokenIndex357, depth357
								if buffer[position] != rune('S') {
									goto l310
								}
								position++
							}
						l357:
							{
								position359, tokenIndex359, depth359 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l360
								}
								position++
								goto l359
							l360:
								position, tokenIndex, depth = position359, tokenIndex359, depth359
								if buffer[position] != rune('E') {
									goto l310
								}
								position++
							}
						l359:
							if !_rules[ruleKEY]() {
								goto l310
							}
							{
								position361, tokenIndex361, depth361 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l362
								}
								{
									position363, tokenIndex363, depth363 := position, tokenIndex, depth
									if buffer[position] != rune('b') {
										goto l364
									}
									position++
									goto l363
								l364:
									position, tokenIndex, depth = position363, tokenIndex363, depth363
									if buffer[position] != rune('B') {
										goto l362
									}
									position++
								}
							l363:
								{
									position365, tokenIndex365, depth365 := position, tokenIndex, depth
									if buffer[position] != rune('y') {
										goto l366
									}
									position++
									goto l365
								l366:
									position, tokenIndex, depth = position365, tokenIndex365, depth365
									if buffer[position] != rune('Y') {
										goto l362
									}
									position++
								}
							l365:
								if !_rules[ruleKEY]() {
									goto l362
								}
								goto l361
							l362:
								position, tokenIndex, depth = position361, tokenIndex361, depth361
								{
									add(ruleAction61, position)
								}
							}
						l361:
							{
								position368, tokenIndex368, depth368 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l369
								}
								{
									position370 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l369
									}
									depth--
									add(rulePegText, position370)
								}
								goto l368
							l369:
								position, tokenIndex, depth = position368, tokenIndex368, depth368
								{
									add(ruleAction62, position)
								}
							}
						l368:
							{
								add(ruleAction63, position)
							}
						l373:
							{
								position374, tokenIndex374, depth374 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l374
								}
								if !_rules[ruleCOMMA]() {
									goto l374
								}
								{
									position375, tokenIndex375, depth375 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l376
									}
									{
										position377 := position
										depth++
										if !_rules[ruleCOLUMN_NAME]() {
											goto l376
										}
										depth--
										add(rulePegText, position377)
									}
									goto l375
								l376:
									position, tokenIndex, depth = position375, tokenIndex375, depth375
									{
										add(ruleAction64, position)
									}
								}
							l375:
								{
									add(ruleAction65, position)
								}
								goto l373
							l374:
								position, tokenIndex, depth = position374, tokenIndex374, depth374
							}
							depth--
							add(rulecollapseByClause, position344)
						}
					}
				l312:
					goto l311
				l310:
					position, tokenIndex, depth = position310, tokenIndex310, depth310
				}
			l311:
				depth--
				add(ruleoptionalGroupBy, position308)
			}
			return true
		},
		/* 21 expression_function <- <(_ <IDENTIFIER> Action47 _ PAREN_OPEN (expressionList / Action48) optionalGroupBy ((_ PAREN_CLOSE) / Action49) Action50)> */
		nil,
		/* 22 expression_metric <- <(_ <IDENTIFIER> Action51 ((_ '[' (predicate_1 / Action52) ((_ ']') / Action53)) / Action54) Action55)> */
		nil,
		/* 23 groupByClause <- <(_ (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P')) KEY ((_ (('b' / 'B') ('y' / 'Y')) KEY) / Action56) ((_ <COLUMN_NAME>) / Action57) Action58 (_ COMMA ((_ <COLUMN_NAME>) / Action59) Action60)*)> */
		nil,
		/* 24 collapseByClause <- <(_ (('c' / 'C') ('o' / 'O') ('l' / 'L') ('l' / 'L') ('a' / 'A') ('p' / 'P') ('s' / 'S') ('e' / 'E')) KEY ((_ (('b' / 'B') ('y' / 'Y')) KEY) / Action61) ((_ <COLUMN_NAME>) / Action62) Action63 (_ COMMA ((_ <COLUMN_NAME>) / Action64) Action65)*)> */
		nil,
		/* 25 predicateClause <- <(_ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY ((_ predicate_1) / Action66))> */
		nil,
		/* 26 predicate_1 <- <((predicate_2 _ OP_OR (predicate_1 / Action67) Action68) / predicate_2)> */
		func() bool {
			position385, tokenIndex385, depth385 := position, tokenIndex, depth
			{
				position386 := position
				depth++
				{
					position387, tokenIndex387, depth387 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l388
					}
					if !_rules[rule_]() {
						goto l388
					}
					{
						position389 := position
						depth++
						{
							position390, tokenIndex390, depth390 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l391
							}
							position++
							goto l390
						l391:
							position, tokenIndex, depth = position390, tokenIndex390, depth390
							if buffer[position] != rune('O') {
								goto l388
							}
							position++
						}
					l390:
						{
							position392, tokenIndex392, depth392 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l393
							}
							position++
							goto l392
						l393:
							position, tokenIndex, depth = position392, tokenIndex392, depth392
							if buffer[position] != rune('R') {
								goto l388
							}
							position++
						}
					l392:
						if !_rules[ruleKEY]() {
							goto l388
						}
						depth--
						add(ruleOP_OR, position389)
					}
					{
						position394, tokenIndex394, depth394 := position, tokenIndex, depth
						if !_rules[rulepredicate_1]() {
							goto l395
						}
						goto l394
					l395:
						position, tokenIndex, depth = position394, tokenIndex394, depth394
						{
							add(ruleAction67, position)
						}
					}
				l394:
					{
						add(ruleAction68, position)
					}
					goto l387
				l388:
					position, tokenIndex, depth = position387, tokenIndex387, depth387
					if !_rules[rulepredicate_2]() {
						goto l385
					}
				}
			l387:
				depth--
				add(rulepredicate_1, position386)
			}
			return true
		l385:
			position, tokenIndex, depth = position385, tokenIndex385, depth385
			return false
		},
		/* 27 predicate_2 <- <((predicate_3 _ OP_AND (predicate_2 / Action69) Action70) / predicate_3)> */
		func() bool {
			position398, tokenIndex398, depth398 := position, tokenIndex, depth
			{
				position399 := position
				depth++
				{
					position400, tokenIndex400, depth400 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l401
					}
					if !_rules[rule_]() {
						goto l401
					}
					{
						position402 := position
						depth++
						{
							position403, tokenIndex403, depth403 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l404
							}
							position++
							goto l403
						l404:
							position, tokenIndex, depth = position403, tokenIndex403, depth403
							if buffer[position] != rune('A') {
								goto l401
							}
							position++
						}
					l403:
						{
							position405, tokenIndex405, depth405 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l406
							}
							position++
							goto l405
						l406:
							position, tokenIndex, depth = position405, tokenIndex405, depth405
							if buffer[position] != rune('N') {
								goto l401
							}
							position++
						}
					l405:
						{
							position407, tokenIndex407, depth407 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l408
							}
							position++
							goto l407
						l408:
							position, tokenIndex, depth = position407, tokenIndex407, depth407
							if buffer[position] != rune('D') {
								goto l401
							}
							position++
						}
					l407:
						if !_rules[ruleKEY]() {
							goto l401
						}
						depth--
						add(ruleOP_AND, position402)
					}
					{
						position409, tokenIndex409, depth409 := position, tokenIndex, depth
						if !_rules[rulepredicate_2]() {
							goto l410
						}
						goto l409
					l410:
						position, tokenIndex, depth = position409, tokenIndex409, depth409
						{
							add(ruleAction69, position)
						}
					}
				l409:
					{
						add(ruleAction70, position)
					}
					goto l400
				l401:
					position, tokenIndex, depth = position400, tokenIndex400, depth400
					if !_rules[rulepredicate_3]() {
						goto l398
					}
				}
			l400:
				depth--
				add(rulepredicate_2, position399)
			}
			return true
		l398:
			position, tokenIndex, depth = position398, tokenIndex398, depth398
			return false
		},
		/* 28 predicate_3 <- <((_ OP_NOT (predicate_3 / Action71) Action72) / (_ PAREN_OPEN (predicate_1 / Action73) ((_ PAREN_CLOSE) / Action74)) / tagMatcher)> */
		func() bool {
			position413, tokenIndex413, depth413 := position, tokenIndex, depth
			{
				position414 := position
				depth++
				{
					position415, tokenIndex415, depth415 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l416
					}
					{
						position417 := position
						depth++
						{
							position418, tokenIndex418, depth418 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l419
							}
							position++
							goto l418
						l419:
							position, tokenIndex, depth = position418, tokenIndex418, depth418
							if buffer[position] != rune('N') {
								goto l416
							}
							position++
						}
					l418:
						{
							position420, tokenIndex420, depth420 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l421
							}
							position++
							goto l420
						l421:
							position, tokenIndex, depth = position420, tokenIndex420, depth420
							if buffer[position] != rune('O') {
								goto l416
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
								goto l416
							}
							position++
						}
					l422:
						if !_rules[ruleKEY]() {
							goto l416
						}
						depth--
						add(ruleOP_NOT, position417)
					}
					{
						position424, tokenIndex424, depth424 := position, tokenIndex, depth
						if !_rules[rulepredicate_3]() {
							goto l425
						}
						goto l424
					l425:
						position, tokenIndex, depth = position424, tokenIndex424, depth424
						{
							add(ruleAction71, position)
						}
					}
				l424:
					{
						add(ruleAction72, position)
					}
					goto l415
				l416:
					position, tokenIndex, depth = position415, tokenIndex415, depth415
					if !_rules[rule_]() {
						goto l428
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l428
					}
					{
						position429, tokenIndex429, depth429 := position, tokenIndex, depth
						if !_rules[rulepredicate_1]() {
							goto l430
						}
						goto l429
					l430:
						position, tokenIndex, depth = position429, tokenIndex429, depth429
						{
							add(ruleAction73, position)
						}
					}
				l429:
					{
						position432, tokenIndex432, depth432 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l433
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l433
						}
						goto l432
					l433:
						position, tokenIndex, depth = position432, tokenIndex432, depth432
						{
							add(ruleAction74, position)
						}
					}
				l432:
					goto l415
				l428:
					position, tokenIndex, depth = position415, tokenIndex415, depth415
					{
						position435 := position
						depth++
						if !_rules[ruletagName]() {
							goto l413
						}
						{
							position436, tokenIndex436, depth436 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l437
							}
							if buffer[position] != rune('=') {
								goto l437
							}
							position++
							{
								position438, tokenIndex438, depth438 := position, tokenIndex, depth
								if !_rules[ruleliteralString]() {
									goto l439
								}
								goto l438
							l439:
								position, tokenIndex, depth = position438, tokenIndex438, depth438
								{
									add(ruleAction75, position)
								}
							}
						l438:
							{
								add(ruleAction76, position)
							}
							goto l436
						l437:
							position, tokenIndex, depth = position436, tokenIndex436, depth436
							if !_rules[rule_]() {
								goto l442
							}
							if buffer[position] != rune('!') {
								goto l442
							}
							position++
							if buffer[position] != rune('=') {
								goto l442
							}
							position++
							{
								position443, tokenIndex443, depth443 := position, tokenIndex, depth
								if !_rules[ruleliteralString]() {
									goto l444
								}
								goto l443
							l444:
								position, tokenIndex, depth = position443, tokenIndex443, depth443
								{
									add(ruleAction77, position)
								}
							}
						l443:
							{
								add(ruleAction78, position)
							}
							{
								add(ruleAction79, position)
							}
							goto l436
						l442:
							position, tokenIndex, depth = position436, tokenIndex436, depth436
							if !_rules[rule_]() {
								goto l448
							}
							{
								position449, tokenIndex449, depth449 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l450
								}
								position++
								goto l449
							l450:
								position, tokenIndex, depth = position449, tokenIndex449, depth449
								if buffer[position] != rune('M') {
									goto l448
								}
								position++
							}
						l449:
							{
								position451, tokenIndex451, depth451 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l452
								}
								position++
								goto l451
							l452:
								position, tokenIndex, depth = position451, tokenIndex451, depth451
								if buffer[position] != rune('A') {
									goto l448
								}
								position++
							}
						l451:
							{
								position453, tokenIndex453, depth453 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l454
								}
								position++
								goto l453
							l454:
								position, tokenIndex, depth = position453, tokenIndex453, depth453
								if buffer[position] != rune('T') {
									goto l448
								}
								position++
							}
						l453:
							{
								position455, tokenIndex455, depth455 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l456
								}
								position++
								goto l455
							l456:
								position, tokenIndex, depth = position455, tokenIndex455, depth455
								if buffer[position] != rune('C') {
									goto l448
								}
								position++
							}
						l455:
							{
								position457, tokenIndex457, depth457 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l458
								}
								position++
								goto l457
							l458:
								position, tokenIndex, depth = position457, tokenIndex457, depth457
								if buffer[position] != rune('H') {
									goto l448
								}
								position++
							}
						l457:
							if !_rules[ruleKEY]() {
								goto l448
							}
							{
								position459, tokenIndex459, depth459 := position, tokenIndex, depth
								if !_rules[ruleliteralString]() {
									goto l460
								}
								goto l459
							l460:
								position, tokenIndex, depth = position459, tokenIndex459, depth459
								{
									add(ruleAction80, position)
								}
							}
						l459:
							{
								add(ruleAction81, position)
							}
							goto l436
						l448:
							position, tokenIndex, depth = position436, tokenIndex436, depth436
							if !_rules[rule_]() {
								goto l463
							}
							{
								position464, tokenIndex464, depth464 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l465
								}
								position++
								goto l464
							l465:
								position, tokenIndex, depth = position464, tokenIndex464, depth464
								if buffer[position] != rune('I') {
									goto l463
								}
								position++
							}
						l464:
							{
								position466, tokenIndex466, depth466 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l467
								}
								position++
								goto l466
							l467:
								position, tokenIndex, depth = position466, tokenIndex466, depth466
								if buffer[position] != rune('N') {
									goto l463
								}
								position++
							}
						l466:
							if !_rules[ruleKEY]() {
								goto l463
							}
							{
								position468, tokenIndex468, depth468 := position, tokenIndex, depth
								{
									position470 := position
									depth++
									{
										add(ruleAction86, position)
									}
									if !_rules[rule_]() {
										goto l469
									}
									if !_rules[rulePAREN_OPEN]() {
										goto l469
									}
									{
										position472, tokenIndex472, depth472 := position, tokenIndex, depth
										if !_rules[ruleliteralListString]() {
											goto l473
										}
										goto l472
									l473:
										position, tokenIndex, depth = position472, tokenIndex472, depth472
										{
											add(ruleAction87, position)
										}
									}
								l472:
								l475:
									{
										position476, tokenIndex476, depth476 := position, tokenIndex, depth
										if !_rules[rule_]() {
											goto l476
										}
										if !_rules[ruleCOMMA]() {
											goto l476
										}
										{
											position477, tokenIndex477, depth477 := position, tokenIndex, depth
											if !_rules[ruleliteralListString]() {
												goto l478
											}
											goto l477
										l478:
											position, tokenIndex, depth = position477, tokenIndex477, depth477
											{
												add(ruleAction88, position)
											}
										}
									l477:
										goto l475
									l476:
										position, tokenIndex, depth = position476, tokenIndex476, depth476
									}
									{
										position480, tokenIndex480, depth480 := position, tokenIndex, depth
										if !_rules[rule_]() {
											goto l481
										}
										if !_rules[rulePAREN_CLOSE]() {
											goto l481
										}
										goto l480
									l481:
										position, tokenIndex, depth = position480, tokenIndex480, depth480
										{
											add(ruleAction89, position)
										}
									}
								l480:
									depth--
									add(ruleliteralList, position470)
								}
								goto l468
							l469:
								position, tokenIndex, depth = position468, tokenIndex468, depth468
								{
									add(ruleAction82, position)
								}
							}
						l468:
							{
								add(ruleAction83, position)
							}
							goto l436
						l463:
							position, tokenIndex, depth = position436, tokenIndex436, depth436
							{
								add(ruleAction84, position)
							}
						}
					l436:
						depth--
						add(ruletagMatcher, position435)
					}
				}
			l415:
				depth--
				add(rulepredicate_3, position414)
			}
			return true
		l413:
			position, tokenIndex, depth = position413, tokenIndex413, depth413
			return false
		},
		/* 29 tagMatcher <- <(tagName ((_ '=' (literalString / Action75) Action76) / (_ ('!' '=') (literalString / Action77) Action78 Action79) / (_ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H')) KEY (literalString / Action80) Action81) / (_ (('i' / 'I') ('n' / 'N')) KEY (literalList / Action82) Action83) / Action84))> */
		nil,
		/* 30 literalString <- <(_ STRING Action85)> */
		func() bool {
			position487, tokenIndex487, depth487 := position, tokenIndex, depth
			{
				position488 := position
				depth++
				if !_rules[rule_]() {
					goto l487
				}
				if !_rules[ruleSTRING]() {
					goto l487
				}
				{
					add(ruleAction85, position)
				}
				depth--
				add(ruleliteralString, position488)
			}
			return true
		l487:
			position, tokenIndex, depth = position487, tokenIndex487, depth487
			return false
		},
		/* 31 literalList <- <(Action86 _ PAREN_OPEN (literalListString / Action87) (_ COMMA (literalListString / Action88))* ((_ PAREN_CLOSE) / Action89))> */
		nil,
		/* 32 literalListString <- <(_ STRING Action90)> */
		func() bool {
			position491, tokenIndex491, depth491 := position, tokenIndex, depth
			{
				position492 := position
				depth++
				if !_rules[rule_]() {
					goto l491
				}
				if !_rules[ruleSTRING]() {
					goto l491
				}
				{
					add(ruleAction90, position)
				}
				depth--
				add(ruleliteralListString, position492)
			}
			return true
		l491:
			position, tokenIndex, depth = position491, tokenIndex491, depth491
			return false
		},
		/* 33 tagName <- <(_ <TAG_NAME> Action91)> */
		func() bool {
			position494, tokenIndex494, depth494 := position, tokenIndex, depth
			{
				position495 := position
				depth++
				if !_rules[rule_]() {
					goto l494
				}
				{
					position496 := position
					depth++
					{
						position497 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l494
						}
						depth--
						add(ruleTAG_NAME, position497)
					}
					depth--
					add(rulePegText, position496)
				}
				{
					add(ruleAction91, position)
				}
				depth--
				add(ruletagName, position495)
			}
			return true
		l494:
			position, tokenIndex, depth = position494, tokenIndex494, depth494
			return false
		},
		/* 34 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position499, tokenIndex499, depth499 := position, tokenIndex, depth
			{
				position500 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l499
				}
				depth--
				add(ruleCOLUMN_NAME, position500)
			}
			return true
		l499:
			position, tokenIndex, depth = position499, tokenIndex499, depth499
			return false
		},
		/* 35 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 36 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 37 IDENTIFIER <- <(('`' CHAR* ('`' / Action92)) / (!(KEYWORD KEY) ID_SEGMENT ('.' (ID_SEGMENT / Action93))*))> */
		func() bool {
			position503, tokenIndex503, depth503 := position, tokenIndex, depth
			{
				position504 := position
				depth++
				{
					position505, tokenIndex505, depth505 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l506
					}
					position++
				l507:
					{
						position508, tokenIndex508, depth508 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l508
						}
						goto l507
					l508:
						position, tokenIndex, depth = position508, tokenIndex508, depth508
					}
					{
						position509, tokenIndex509, depth509 := position, tokenIndex, depth
						if buffer[position] != rune('`') {
							goto l510
						}
						position++
						goto l509
					l510:
						position, tokenIndex, depth = position509, tokenIndex509, depth509
						{
							add(ruleAction92, position)
						}
					}
				l509:
					goto l505
				l506:
					position, tokenIndex, depth = position505, tokenIndex505, depth505
					{
						position512, tokenIndex512, depth512 := position, tokenIndex, depth
						{
							position513 := position
							depth++
							{
								position514, tokenIndex514, depth514 := position, tokenIndex, depth
								{
									position516, tokenIndex516, depth516 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l517
									}
									position++
									goto l516
								l517:
									position, tokenIndex, depth = position516, tokenIndex516, depth516
									if buffer[position] != rune('A') {
										goto l515
									}
									position++
								}
							l516:
								{
									position518, tokenIndex518, depth518 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l519
									}
									position++
									goto l518
								l519:
									position, tokenIndex, depth = position518, tokenIndex518, depth518
									if buffer[position] != rune('L') {
										goto l515
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
										goto l515
									}
									position++
								}
							l520:
								goto l514
							l515:
								position, tokenIndex, depth = position514, tokenIndex514, depth514
								{
									position523, tokenIndex523, depth523 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l524
									}
									position++
									goto l523
								l524:
									position, tokenIndex, depth = position523, tokenIndex523, depth523
									if buffer[position] != rune('A') {
										goto l522
									}
									position++
								}
							l523:
								{
									position525, tokenIndex525, depth525 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l526
									}
									position++
									goto l525
								l526:
									position, tokenIndex, depth = position525, tokenIndex525, depth525
									if buffer[position] != rune('N') {
										goto l522
									}
									position++
								}
							l525:
								{
									position527, tokenIndex527, depth527 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l528
									}
									position++
									goto l527
								l528:
									position, tokenIndex, depth = position527, tokenIndex527, depth527
									if buffer[position] != rune('D') {
										goto l522
									}
									position++
								}
							l527:
								goto l514
							l522:
								position, tokenIndex, depth = position514, tokenIndex514, depth514
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
										goto l529
									}
									position++
								}
							l530:
								{
									position532, tokenIndex532, depth532 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l533
									}
									position++
									goto l532
								l533:
									position, tokenIndex, depth = position532, tokenIndex532, depth532
									if buffer[position] != rune('A') {
										goto l529
									}
									position++
								}
							l532:
								{
									position534, tokenIndex534, depth534 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l535
									}
									position++
									goto l534
								l535:
									position, tokenIndex, depth = position534, tokenIndex534, depth534
									if buffer[position] != rune('T') {
										goto l529
									}
									position++
								}
							l534:
								{
									position536, tokenIndex536, depth536 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l537
									}
									position++
									goto l536
								l537:
									position, tokenIndex, depth = position536, tokenIndex536, depth536
									if buffer[position] != rune('C') {
										goto l529
									}
									position++
								}
							l536:
								{
									position538, tokenIndex538, depth538 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l539
									}
									position++
									goto l538
								l539:
									position, tokenIndex, depth = position538, tokenIndex538, depth538
									if buffer[position] != rune('H') {
										goto l529
									}
									position++
								}
							l538:
								goto l514
							l529:
								position, tokenIndex, depth = position514, tokenIndex514, depth514
								{
									position541, tokenIndex541, depth541 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l542
									}
									position++
									goto l541
								l542:
									position, tokenIndex, depth = position541, tokenIndex541, depth541
									if buffer[position] != rune('S') {
										goto l540
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
								{
									position547, tokenIndex547, depth547 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l548
									}
									position++
									goto l547
								l548:
									position, tokenIndex, depth = position547, tokenIndex547, depth547
									if buffer[position] != rune('E') {
										goto l540
									}
									position++
								}
							l547:
								{
									position549, tokenIndex549, depth549 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l550
									}
									position++
									goto l549
								l550:
									position, tokenIndex, depth = position549, tokenIndex549, depth549
									if buffer[position] != rune('C') {
										goto l540
									}
									position++
								}
							l549:
								{
									position551, tokenIndex551, depth551 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l552
									}
									position++
									goto l551
								l552:
									position, tokenIndex, depth = position551, tokenIndex551, depth551
									if buffer[position] != rune('T') {
										goto l540
									}
									position++
								}
							l551:
								goto l514
							l540:
								position, tokenIndex, depth = position514, tokenIndex514, depth514
								{
									switch buffer[position] {
									case 'M', 'm':
										{
											position554, tokenIndex554, depth554 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l555
											}
											position++
											goto l554
										l555:
											position, tokenIndex, depth = position554, tokenIndex554, depth554
											if buffer[position] != rune('M') {
												goto l512
											}
											position++
										}
									l554:
										{
											position556, tokenIndex556, depth556 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l557
											}
											position++
											goto l556
										l557:
											position, tokenIndex, depth = position556, tokenIndex556, depth556
											if buffer[position] != rune('E') {
												goto l512
											}
											position++
										}
									l556:
										{
											position558, tokenIndex558, depth558 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l559
											}
											position++
											goto l558
										l559:
											position, tokenIndex, depth = position558, tokenIndex558, depth558
											if buffer[position] != rune('T') {
												goto l512
											}
											position++
										}
									l558:
										{
											position560, tokenIndex560, depth560 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l561
											}
											position++
											goto l560
										l561:
											position, tokenIndex, depth = position560, tokenIndex560, depth560
											if buffer[position] != rune('R') {
												goto l512
											}
											position++
										}
									l560:
										{
											position562, tokenIndex562, depth562 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l563
											}
											position++
											goto l562
										l563:
											position, tokenIndex, depth = position562, tokenIndex562, depth562
											if buffer[position] != rune('I') {
												goto l512
											}
											position++
										}
									l562:
										{
											position564, tokenIndex564, depth564 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l565
											}
											position++
											goto l564
										l565:
											position, tokenIndex, depth = position564, tokenIndex564, depth564
											if buffer[position] != rune('C') {
												goto l512
											}
											position++
										}
									l564:
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
												goto l512
											}
											position++
										}
									l566:
										break
									case 'W', 'w':
										{
											position568, tokenIndex568, depth568 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l569
											}
											position++
											goto l568
										l569:
											position, tokenIndex, depth = position568, tokenIndex568, depth568
											if buffer[position] != rune('W') {
												goto l512
											}
											position++
										}
									l568:
										{
											position570, tokenIndex570, depth570 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l571
											}
											position++
											goto l570
										l571:
											position, tokenIndex, depth = position570, tokenIndex570, depth570
											if buffer[position] != rune('H') {
												goto l512
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
												goto l512
											}
											position++
										}
									l572:
										{
											position574, tokenIndex574, depth574 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l575
											}
											position++
											goto l574
										l575:
											position, tokenIndex, depth = position574, tokenIndex574, depth574
											if buffer[position] != rune('R') {
												goto l512
											}
											position++
										}
									l574:
										{
											position576, tokenIndex576, depth576 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l577
											}
											position++
											goto l576
										l577:
											position, tokenIndex, depth = position576, tokenIndex576, depth576
											if buffer[position] != rune('E') {
												goto l512
											}
											position++
										}
									l576:
										break
									case 'O', 'o':
										{
											position578, tokenIndex578, depth578 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l579
											}
											position++
											goto l578
										l579:
											position, tokenIndex, depth = position578, tokenIndex578, depth578
											if buffer[position] != rune('O') {
												goto l512
											}
											position++
										}
									l578:
										{
											position580, tokenIndex580, depth580 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l581
											}
											position++
											goto l580
										l581:
											position, tokenIndex, depth = position580, tokenIndex580, depth580
											if buffer[position] != rune('R') {
												goto l512
											}
											position++
										}
									l580:
										break
									case 'N', 'n':
										{
											position582, tokenIndex582, depth582 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l583
											}
											position++
											goto l582
										l583:
											position, tokenIndex, depth = position582, tokenIndex582, depth582
											if buffer[position] != rune('N') {
												goto l512
											}
											position++
										}
									l582:
										{
											position584, tokenIndex584, depth584 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l585
											}
											position++
											goto l584
										l585:
											position, tokenIndex, depth = position584, tokenIndex584, depth584
											if buffer[position] != rune('O') {
												goto l512
											}
											position++
										}
									l584:
										{
											position586, tokenIndex586, depth586 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l587
											}
											position++
											goto l586
										l587:
											position, tokenIndex, depth = position586, tokenIndex586, depth586
											if buffer[position] != rune('T') {
												goto l512
											}
											position++
										}
									l586:
										break
									case 'I', 'i':
										{
											position588, tokenIndex588, depth588 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l589
											}
											position++
											goto l588
										l589:
											position, tokenIndex, depth = position588, tokenIndex588, depth588
											if buffer[position] != rune('I') {
												goto l512
											}
											position++
										}
									l588:
										{
											position590, tokenIndex590, depth590 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l591
											}
											position++
											goto l590
										l591:
											position, tokenIndex, depth = position590, tokenIndex590, depth590
											if buffer[position] != rune('N') {
												goto l512
											}
											position++
										}
									l590:
										break
									case 'C', 'c':
										{
											position592, tokenIndex592, depth592 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l593
											}
											position++
											goto l592
										l593:
											position, tokenIndex, depth = position592, tokenIndex592, depth592
											if buffer[position] != rune('C') {
												goto l512
											}
											position++
										}
									l592:
										{
											position594, tokenIndex594, depth594 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l595
											}
											position++
											goto l594
										l595:
											position, tokenIndex, depth = position594, tokenIndex594, depth594
											if buffer[position] != rune('O') {
												goto l512
											}
											position++
										}
									l594:
										{
											position596, tokenIndex596, depth596 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l597
											}
											position++
											goto l596
										l597:
											position, tokenIndex, depth = position596, tokenIndex596, depth596
											if buffer[position] != rune('L') {
												goto l512
											}
											position++
										}
									l596:
										{
											position598, tokenIndex598, depth598 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l599
											}
											position++
											goto l598
										l599:
											position, tokenIndex, depth = position598, tokenIndex598, depth598
											if buffer[position] != rune('L') {
												goto l512
											}
											position++
										}
									l598:
										{
											position600, tokenIndex600, depth600 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l601
											}
											position++
											goto l600
										l601:
											position, tokenIndex, depth = position600, tokenIndex600, depth600
											if buffer[position] != rune('A') {
												goto l512
											}
											position++
										}
									l600:
										{
											position602, tokenIndex602, depth602 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l603
											}
											position++
											goto l602
										l603:
											position, tokenIndex, depth = position602, tokenIndex602, depth602
											if buffer[position] != rune('P') {
												goto l512
											}
											position++
										}
									l602:
										{
											position604, tokenIndex604, depth604 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l605
											}
											position++
											goto l604
										l605:
											position, tokenIndex, depth = position604, tokenIndex604, depth604
											if buffer[position] != rune('S') {
												goto l512
											}
											position++
										}
									l604:
										{
											position606, tokenIndex606, depth606 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l607
											}
											position++
											goto l606
										l607:
											position, tokenIndex, depth = position606, tokenIndex606, depth606
											if buffer[position] != rune('E') {
												goto l512
											}
											position++
										}
									l606:
										break
									case 'G', 'g':
										{
											position608, tokenIndex608, depth608 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l609
											}
											position++
											goto l608
										l609:
											position, tokenIndex, depth = position608, tokenIndex608, depth608
											if buffer[position] != rune('G') {
												goto l512
											}
											position++
										}
									l608:
										{
											position610, tokenIndex610, depth610 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l611
											}
											position++
											goto l610
										l611:
											position, tokenIndex, depth = position610, tokenIndex610, depth610
											if buffer[position] != rune('R') {
												goto l512
											}
											position++
										}
									l610:
										{
											position612, tokenIndex612, depth612 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l613
											}
											position++
											goto l612
										l613:
											position, tokenIndex, depth = position612, tokenIndex612, depth612
											if buffer[position] != rune('O') {
												goto l512
											}
											position++
										}
									l612:
										{
											position614, tokenIndex614, depth614 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l615
											}
											position++
											goto l614
										l615:
											position, tokenIndex, depth = position614, tokenIndex614, depth614
											if buffer[position] != rune('U') {
												goto l512
											}
											position++
										}
									l614:
										{
											position616, tokenIndex616, depth616 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l617
											}
											position++
											goto l616
										l617:
											position, tokenIndex, depth = position616, tokenIndex616, depth616
											if buffer[position] != rune('P') {
												goto l512
											}
											position++
										}
									l616:
										break
									case 'D', 'd':
										{
											position618, tokenIndex618, depth618 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l619
											}
											position++
											goto l618
										l619:
											position, tokenIndex, depth = position618, tokenIndex618, depth618
											if buffer[position] != rune('D') {
												goto l512
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
												goto l512
											}
											position++
										}
									l620:
										{
											position622, tokenIndex622, depth622 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l623
											}
											position++
											goto l622
										l623:
											position, tokenIndex, depth = position622, tokenIndex622, depth622
											if buffer[position] != rune('S') {
												goto l512
											}
											position++
										}
									l622:
										{
											position624, tokenIndex624, depth624 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l625
											}
											position++
											goto l624
										l625:
											position, tokenIndex, depth = position624, tokenIndex624, depth624
											if buffer[position] != rune('C') {
												goto l512
											}
											position++
										}
									l624:
										{
											position626, tokenIndex626, depth626 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l627
											}
											position++
											goto l626
										l627:
											position, tokenIndex, depth = position626, tokenIndex626, depth626
											if buffer[position] != rune('R') {
												goto l512
											}
											position++
										}
									l626:
										{
											position628, tokenIndex628, depth628 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l629
											}
											position++
											goto l628
										l629:
											position, tokenIndex, depth = position628, tokenIndex628, depth628
											if buffer[position] != rune('I') {
												goto l512
											}
											position++
										}
									l628:
										{
											position630, tokenIndex630, depth630 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l631
											}
											position++
											goto l630
										l631:
											position, tokenIndex, depth = position630, tokenIndex630, depth630
											if buffer[position] != rune('B') {
												goto l512
											}
											position++
										}
									l630:
										{
											position632, tokenIndex632, depth632 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l633
											}
											position++
											goto l632
										l633:
											position, tokenIndex, depth = position632, tokenIndex632, depth632
											if buffer[position] != rune('E') {
												goto l512
											}
											position++
										}
									l632:
										break
									case 'B', 'b':
										{
											position634, tokenIndex634, depth634 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l635
											}
											position++
											goto l634
										l635:
											position, tokenIndex, depth = position634, tokenIndex634, depth634
											if buffer[position] != rune('B') {
												goto l512
											}
											position++
										}
									l634:
										{
											position636, tokenIndex636, depth636 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l637
											}
											position++
											goto l636
										l637:
											position, tokenIndex, depth = position636, tokenIndex636, depth636
											if buffer[position] != rune('Y') {
												goto l512
											}
											position++
										}
									l636:
										break
									case 'A', 'a':
										{
											position638, tokenIndex638, depth638 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l639
											}
											position++
											goto l638
										l639:
											position, tokenIndex, depth = position638, tokenIndex638, depth638
											if buffer[position] != rune('A') {
												goto l512
											}
											position++
										}
									l638:
										{
											position640, tokenIndex640, depth640 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l641
											}
											position++
											goto l640
										l641:
											position, tokenIndex, depth = position640, tokenIndex640, depth640
											if buffer[position] != rune('S') {
												goto l512
											}
											position++
										}
									l640:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l512
										}
										break
									}
								}

							}
						l514:
							depth--
							add(ruleKEYWORD, position513)
						}
						if !_rules[ruleKEY]() {
							goto l512
						}
						goto l503
					l512:
						position, tokenIndex, depth = position512, tokenIndex512, depth512
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l503
					}
				l642:
					{
						position643, tokenIndex643, depth643 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l643
						}
						position++
						{
							position644, tokenIndex644, depth644 := position, tokenIndex, depth
							if !_rules[ruleID_SEGMENT]() {
								goto l645
							}
							goto l644
						l645:
							position, tokenIndex, depth = position644, tokenIndex644, depth644
							{
								add(ruleAction93, position)
							}
						}
					l644:
						goto l642
					l643:
						position, tokenIndex, depth = position643, tokenIndex643, depth643
					}
				}
			l505:
				depth--
				add(ruleIDENTIFIER, position504)
			}
			return true
		l503:
			position, tokenIndex, depth = position503, tokenIndex503, depth503
			return false
		},
		/* 38 TIMESTAMP <- <((_ <(NUMBER ([a-z] / [A-Z])*)>) / (_ STRING) / (_ <(('n' / 'N') ('o' / 'O') ('w' / 'W'))> KEY))> */
		nil,
		/* 39 ID_SEGMENT <- <(ID_START ID_CONT*)> */
		func() bool {
			position648, tokenIndex648, depth648 := position, tokenIndex, depth
			{
				position649 := position
				depth++
				if !_rules[ruleID_START]() {
					goto l648
				}
			l650:
				{
					position651, tokenIndex651, depth651 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l651
					}
					goto l650
				l651:
					position, tokenIndex, depth = position651, tokenIndex651, depth651
				}
				depth--
				add(ruleID_SEGMENT, position649)
			}
			return true
		l648:
			position, tokenIndex, depth = position648, tokenIndex648, depth648
			return false
		},
		/* 40 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position652, tokenIndex652, depth652 := position, tokenIndex, depth
			{
				position653 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l652
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l652
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l652
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position653)
			}
			return true
		l652:
			position, tokenIndex, depth = position652, tokenIndex652, depth652
			return false
		},
		/* 41 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position655, tokenIndex655, depth655 := position, tokenIndex, depth
			{
				position656 := position
				depth++
				{
					position657, tokenIndex657, depth657 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l658
					}
					goto l657
				l658:
					position, tokenIndex, depth = position657, tokenIndex657, depth657
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l655
					}
					position++
				}
			l657:
				depth--
				add(ruleID_CONT, position656)
			}
			return true
		l655:
			position, tokenIndex, depth = position655, tokenIndex655, depth655
			return false
		},
		/* 42 PROPERTY_KEY <- <((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> KEY ((_ (('b' / 'B') ('y' / 'Y')) KEY) / Action94))) | (&('R' | 'r') (<(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))> KEY)) | (&('T' | 't') (<(('t' / 'T') ('o' / 'O'))> KEY)) | (&('F' | 'f') (<(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))> KEY)))> */
		func() bool {
			position659, tokenIndex659, depth659 := position, tokenIndex, depth
			{
				position660 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position662 := position
							depth++
							{
								position663, tokenIndex663, depth663 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l664
								}
								position++
								goto l663
							l664:
								position, tokenIndex, depth = position663, tokenIndex663, depth663
								if buffer[position] != rune('S') {
									goto l659
								}
								position++
							}
						l663:
							{
								position665, tokenIndex665, depth665 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l666
								}
								position++
								goto l665
							l666:
								position, tokenIndex, depth = position665, tokenIndex665, depth665
								if buffer[position] != rune('A') {
									goto l659
								}
								position++
							}
						l665:
							{
								position667, tokenIndex667, depth667 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l668
								}
								position++
								goto l667
							l668:
								position, tokenIndex, depth = position667, tokenIndex667, depth667
								if buffer[position] != rune('M') {
									goto l659
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
									goto l659
								}
								position++
							}
						l669:
							{
								position671, tokenIndex671, depth671 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l672
								}
								position++
								goto l671
							l672:
								position, tokenIndex, depth = position671, tokenIndex671, depth671
								if buffer[position] != rune('L') {
									goto l659
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
									goto l659
								}
								position++
							}
						l673:
							depth--
							add(rulePegText, position662)
						}
						if !_rules[ruleKEY]() {
							goto l659
						}
						{
							position675, tokenIndex675, depth675 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l676
							}
							{
								position677, tokenIndex677, depth677 := position, tokenIndex, depth
								if buffer[position] != rune('b') {
									goto l678
								}
								position++
								goto l677
							l678:
								position, tokenIndex, depth = position677, tokenIndex677, depth677
								if buffer[position] != rune('B') {
									goto l676
								}
								position++
							}
						l677:
							{
								position679, tokenIndex679, depth679 := position, tokenIndex, depth
								if buffer[position] != rune('y') {
									goto l680
								}
								position++
								goto l679
							l680:
								position, tokenIndex, depth = position679, tokenIndex679, depth679
								if buffer[position] != rune('Y') {
									goto l676
								}
								position++
							}
						l679:
							if !_rules[ruleKEY]() {
								goto l676
							}
							goto l675
						l676:
							position, tokenIndex, depth = position675, tokenIndex675, depth675
							{
								add(ruleAction94, position)
							}
						}
					l675:
						break
					case 'R', 'r':
						{
							position682 := position
							depth++
							{
								position683, tokenIndex683, depth683 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l684
								}
								position++
								goto l683
							l684:
								position, tokenIndex, depth = position683, tokenIndex683, depth683
								if buffer[position] != rune('R') {
									goto l659
								}
								position++
							}
						l683:
							{
								position685, tokenIndex685, depth685 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l686
								}
								position++
								goto l685
							l686:
								position, tokenIndex, depth = position685, tokenIndex685, depth685
								if buffer[position] != rune('E') {
									goto l659
								}
								position++
							}
						l685:
							{
								position687, tokenIndex687, depth687 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l688
								}
								position++
								goto l687
							l688:
								position, tokenIndex, depth = position687, tokenIndex687, depth687
								if buffer[position] != rune('S') {
									goto l659
								}
								position++
							}
						l687:
							{
								position689, tokenIndex689, depth689 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l690
								}
								position++
								goto l689
							l690:
								position, tokenIndex, depth = position689, tokenIndex689, depth689
								if buffer[position] != rune('O') {
									goto l659
								}
								position++
							}
						l689:
							{
								position691, tokenIndex691, depth691 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l692
								}
								position++
								goto l691
							l692:
								position, tokenIndex, depth = position691, tokenIndex691, depth691
								if buffer[position] != rune('L') {
									goto l659
								}
								position++
							}
						l691:
							{
								position693, tokenIndex693, depth693 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l694
								}
								position++
								goto l693
							l694:
								position, tokenIndex, depth = position693, tokenIndex693, depth693
								if buffer[position] != rune('U') {
									goto l659
								}
								position++
							}
						l693:
							{
								position695, tokenIndex695, depth695 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l696
								}
								position++
								goto l695
							l696:
								position, tokenIndex, depth = position695, tokenIndex695, depth695
								if buffer[position] != rune('T') {
									goto l659
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
									goto l659
								}
								position++
							}
						l697:
							{
								position699, tokenIndex699, depth699 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l700
								}
								position++
								goto l699
							l700:
								position, tokenIndex, depth = position699, tokenIndex699, depth699
								if buffer[position] != rune('O') {
									goto l659
								}
								position++
							}
						l699:
							{
								position701, tokenIndex701, depth701 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l702
								}
								position++
								goto l701
							l702:
								position, tokenIndex, depth = position701, tokenIndex701, depth701
								if buffer[position] != rune('N') {
									goto l659
								}
								position++
							}
						l701:
							depth--
							add(rulePegText, position682)
						}
						if !_rules[ruleKEY]() {
							goto l659
						}
						break
					case 'T', 't':
						{
							position703 := position
							depth++
							{
								position704, tokenIndex704, depth704 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l705
								}
								position++
								goto l704
							l705:
								position, tokenIndex, depth = position704, tokenIndex704, depth704
								if buffer[position] != rune('T') {
									goto l659
								}
								position++
							}
						l704:
							{
								position706, tokenIndex706, depth706 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l707
								}
								position++
								goto l706
							l707:
								position, tokenIndex, depth = position706, tokenIndex706, depth706
								if buffer[position] != rune('O') {
									goto l659
								}
								position++
							}
						l706:
							depth--
							add(rulePegText, position703)
						}
						if !_rules[ruleKEY]() {
							goto l659
						}
						break
					default:
						{
							position708 := position
							depth++
							{
								position709, tokenIndex709, depth709 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l710
								}
								position++
								goto l709
							l710:
								position, tokenIndex, depth = position709, tokenIndex709, depth709
								if buffer[position] != rune('F') {
									goto l659
								}
								position++
							}
						l709:
							{
								position711, tokenIndex711, depth711 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l712
								}
								position++
								goto l711
							l712:
								position, tokenIndex, depth = position711, tokenIndex711, depth711
								if buffer[position] != rune('R') {
									goto l659
								}
								position++
							}
						l711:
							{
								position713, tokenIndex713, depth713 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l714
								}
								position++
								goto l713
							l714:
								position, tokenIndex, depth = position713, tokenIndex713, depth713
								if buffer[position] != rune('O') {
									goto l659
								}
								position++
							}
						l713:
							{
								position715, tokenIndex715, depth715 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l716
								}
								position++
								goto l715
							l716:
								position, tokenIndex, depth = position715, tokenIndex715, depth715
								if buffer[position] != rune('M') {
									goto l659
								}
								position++
							}
						l715:
							depth--
							add(rulePegText, position708)
						}
						if !_rules[ruleKEY]() {
							goto l659
						}
						break
					}
				}

				depth--
				add(rulePROPERTY_KEY, position660)
			}
			return true
		l659:
			position, tokenIndex, depth = position659, tokenIndex659, depth659
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
			position727, tokenIndex727, depth727 := position, tokenIndex, depth
			{
				position728 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l727
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position728)
			}
			return true
		l727:
			position, tokenIndex, depth = position727, tokenIndex727, depth727
			return false
		},
		/* 54 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position729, tokenIndex729, depth729 := position, tokenIndex, depth
			{
				position730 := position
				depth++
				if buffer[position] != rune('"') {
					goto l729
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position730)
			}
			return true
		l729:
			position, tokenIndex, depth = position729, tokenIndex729, depth729
			return false
		},
		/* 55 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> (QUOTE_SINGLE / Action95)) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> (QUOTE_DOUBLE / Action96)))> */
		func() bool {
			position731, tokenIndex731, depth731 := position, tokenIndex, depth
			{
				position732 := position
				depth++
				{
					position733, tokenIndex733, depth733 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l734
					}
					{
						position735 := position
						depth++
					l736:
						{
							position737, tokenIndex737, depth737 := position, tokenIndex, depth
							{
								position738, tokenIndex738, depth738 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l738
								}
								goto l737
							l738:
								position, tokenIndex, depth = position738, tokenIndex738, depth738
							}
							if !_rules[ruleCHAR]() {
								goto l737
							}
							goto l736
						l737:
							position, tokenIndex, depth = position737, tokenIndex737, depth737
						}
						depth--
						add(rulePegText, position735)
					}
					{
						position739, tokenIndex739, depth739 := position, tokenIndex, depth
						if !_rules[ruleQUOTE_SINGLE]() {
							goto l740
						}
						goto l739
					l740:
						position, tokenIndex, depth = position739, tokenIndex739, depth739
						{
							add(ruleAction95, position)
						}
					}
				l739:
					goto l733
				l734:
					position, tokenIndex, depth = position733, tokenIndex733, depth733
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l731
					}
					{
						position742 := position
						depth++
					l743:
						{
							position744, tokenIndex744, depth744 := position, tokenIndex, depth
							{
								position745, tokenIndex745, depth745 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l745
								}
								goto l744
							l745:
								position, tokenIndex, depth = position745, tokenIndex745, depth745
							}
							if !_rules[ruleCHAR]() {
								goto l744
							}
							goto l743
						l744:
							position, tokenIndex, depth = position744, tokenIndex744, depth744
						}
						depth--
						add(rulePegText, position742)
					}
					{
						position746, tokenIndex746, depth746 := position, tokenIndex, depth
						if !_rules[ruleQUOTE_DOUBLE]() {
							goto l747
						}
						goto l746
					l747:
						position, tokenIndex, depth = position746, tokenIndex746, depth746
						{
							add(ruleAction96, position)
						}
					}
				l746:
				}
			l733:
				depth--
				add(ruleSTRING, position732)
			}
			return true
		l731:
			position, tokenIndex, depth = position731, tokenIndex731, depth731
			return false
		},
		/* 56 CHAR <- <(('\\' ((&('"') (QUOTE_DOUBLE / Action97)) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position749, tokenIndex749, depth749 := position, tokenIndex, depth
			{
				position750 := position
				depth++
				{
					position751, tokenIndex751, depth751 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l752
					}
					position++
					{
						switch buffer[position] {
						case '"':
							{
								position754, tokenIndex754, depth754 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l755
								}
								goto l754
							l755:
								position, tokenIndex, depth = position754, tokenIndex754, depth754
								{
									add(ruleAction97, position)
								}
							}
						l754:
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l752
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l752
							}
							break
						}
					}

					goto l751
				l752:
					position, tokenIndex, depth = position751, tokenIndex751, depth751
					{
						position757, tokenIndex757, depth757 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l757
						}
						goto l749
					l757:
						position, tokenIndex, depth = position757, tokenIndex757, depth757
					}
					if !matchDot() {
						goto l749
					}
				}
			l751:
				depth--
				add(ruleCHAR, position750)
			}
			return true
		l749:
			position, tokenIndex, depth = position749, tokenIndex749, depth749
			return false
		},
		/* 57 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position758, tokenIndex758, depth758 := position, tokenIndex, depth
			{
				position759 := position
				depth++
				{
					position760, tokenIndex760, depth760 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l761
					}
					position++
					goto l760
				l761:
					position, tokenIndex, depth = position760, tokenIndex760, depth760
					if buffer[position] != rune('\\') {
						goto l758
					}
					position++
				}
			l760:
				depth--
				add(ruleESCAPE_CLASS, position759)
			}
			return true
		l758:
			position, tokenIndex, depth = position758, tokenIndex758, depth758
			return false
		},
		/* 58 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position762, tokenIndex762, depth762 := position, tokenIndex, depth
			{
				position763 := position
				depth++
				{
					position764 := position
					depth++
					{
						position765, tokenIndex765, depth765 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l765
						}
						position++
						goto l766
					l765:
						position, tokenIndex, depth = position765, tokenIndex765, depth765
					}
				l766:
					{
						position767 := position
						depth++
						{
							position768, tokenIndex768, depth768 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l769
							}
							position++
							goto l768
						l769:
							position, tokenIndex, depth = position768, tokenIndex768, depth768
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l762
							}
							position++
						l770:
							{
								position771, tokenIndex771, depth771 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l771
								}
								position++
								goto l770
							l771:
								position, tokenIndex, depth = position771, tokenIndex771, depth771
							}
						}
					l768:
						depth--
						add(ruleNUMBER_NATURAL, position767)
					}
					depth--
					add(ruleNUMBER_INTEGER, position764)
				}
				{
					position772, tokenIndex772, depth772 := position, tokenIndex, depth
					{
						position774 := position
						depth++
						if buffer[position] != rune('.') {
							goto l772
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l772
						}
						position++
					l775:
						{
							position776, tokenIndex776, depth776 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l776
							}
							position++
							goto l775
						l776:
							position, tokenIndex, depth = position776, tokenIndex776, depth776
						}
						depth--
						add(ruleNUMBER_FRACTION, position774)
					}
					goto l773
				l772:
					position, tokenIndex, depth = position772, tokenIndex772, depth772
				}
			l773:
				{
					position777, tokenIndex777, depth777 := position, tokenIndex, depth
					{
						position779 := position
						depth++
						{
							position780, tokenIndex780, depth780 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l781
							}
							position++
							goto l780
						l781:
							position, tokenIndex, depth = position780, tokenIndex780, depth780
							if buffer[position] != rune('E') {
								goto l777
							}
							position++
						}
					l780:
						{
							position782, tokenIndex782, depth782 := position, tokenIndex, depth
							{
								position784, tokenIndex784, depth784 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l785
								}
								position++
								goto l784
							l785:
								position, tokenIndex, depth = position784, tokenIndex784, depth784
								if buffer[position] != rune('-') {
									goto l782
								}
								position++
							}
						l784:
							goto l783
						l782:
							position, tokenIndex, depth = position782, tokenIndex782, depth782
						}
					l783:
						{
							position786, tokenIndex786, depth786 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l787
							}
							position++
						l788:
							{
								position789, tokenIndex789, depth789 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l789
								}
								position++
								goto l788
							l789:
								position, tokenIndex, depth = position789, tokenIndex789, depth789
							}
							goto l786
						l787:
							position, tokenIndex, depth = position786, tokenIndex786, depth786
							{
								add(ruleAction98, position)
							}
						}
					l786:
						depth--
						add(ruleNUMBER_EXP, position779)
					}
					goto l778
				l777:
					position, tokenIndex, depth = position777, tokenIndex777, depth777
				}
			l778:
				depth--
				add(ruleNUMBER, position763)
			}
			return true
		l762:
			position, tokenIndex, depth = position762, tokenIndex762, depth762
			return false
		},
		/* 59 NUMBER_NATURAL <- <('0' / ([1-9] [0-9]*))> */
		nil,
		/* 60 NUMBER_FRACTION <- <('.' [0-9]+)> */
		nil,
		/* 61 NUMBER_INTEGER <- <('-'? NUMBER_NATURAL)> */
		nil,
		/* 62 NUMBER_EXP <- <(('e' / 'E') ('+' / '-')? ([0-9]+ / Action98))> */
		nil,
		/* 63 DURATION <- <(NUMBER [a-z]+ KEY)> */
		nil,
		/* 64 PAREN_OPEN <- <'('> */
		func() bool {
			position796, tokenIndex796, depth796 := position, tokenIndex, depth
			{
				position797 := position
				depth++
				if buffer[position] != rune('(') {
					goto l796
				}
				position++
				depth--
				add(rulePAREN_OPEN, position797)
			}
			return true
		l796:
			position, tokenIndex, depth = position796, tokenIndex796, depth796
			return false
		},
		/* 65 PAREN_CLOSE <- <')'> */
		func() bool {
			position798, tokenIndex798, depth798 := position, tokenIndex, depth
			{
				position799 := position
				depth++
				if buffer[position] != rune(')') {
					goto l798
				}
				position++
				depth--
				add(rulePAREN_CLOSE, position799)
			}
			return true
		l798:
			position, tokenIndex, depth = position798, tokenIndex798, depth798
			return false
		},
		/* 66 COMMA <- <','> */
		func() bool {
			position800, tokenIndex800, depth800 := position, tokenIndex, depth
			{
				position801 := position
				depth++
				if buffer[position] != rune(',') {
					goto l800
				}
				position++
				depth--
				add(ruleCOMMA, position801)
			}
			return true
		l800:
			position, tokenIndex, depth = position800, tokenIndex800, depth800
			return false
		},
		/* 67 _ <- <((&('/') COMMENT_BLOCK) | (&('-') COMMENT_TRAIL) | (&('\t' | '\n' | ' ') SPACE))*> */
		func() bool {
			{
				position803 := position
				depth++
			l804:
				{
					position805, tokenIndex805, depth805 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '/':
							{
								position807 := position
								depth++
								if buffer[position] != rune('/') {
									goto l805
								}
								position++
								if buffer[position] != rune('*') {
									goto l805
								}
								position++
							l808:
								{
									position809, tokenIndex809, depth809 := position, tokenIndex, depth
									{
										position810, tokenIndex810, depth810 := position, tokenIndex, depth
										if buffer[position] != rune('*') {
											goto l810
										}
										position++
										if buffer[position] != rune('/') {
											goto l810
										}
										position++
										goto l809
									l810:
										position, tokenIndex, depth = position810, tokenIndex810, depth810
									}
									if !matchDot() {
										goto l809
									}
									goto l808
								l809:
									position, tokenIndex, depth = position809, tokenIndex809, depth809
								}
								if buffer[position] != rune('*') {
									goto l805
								}
								position++
								if buffer[position] != rune('/') {
									goto l805
								}
								position++
								depth--
								add(ruleCOMMENT_BLOCK, position807)
							}
							break
						case '-':
							{
								position811 := position
								depth++
								if buffer[position] != rune('-') {
									goto l805
								}
								position++
								if buffer[position] != rune('-') {
									goto l805
								}
								position++
							l812:
								{
									position813, tokenIndex813, depth813 := position, tokenIndex, depth
									{
										position814, tokenIndex814, depth814 := position, tokenIndex, depth
										if buffer[position] != rune('\n') {
											goto l814
										}
										position++
										goto l813
									l814:
										position, tokenIndex, depth = position814, tokenIndex814, depth814
									}
									if !matchDot() {
										goto l813
									}
									goto l812
								l813:
									position, tokenIndex, depth = position813, tokenIndex813, depth813
								}
								depth--
								add(ruleCOMMENT_TRAIL, position811)
							}
							break
						default:
							{
								position815 := position
								depth++
								{
									switch buffer[position] {
									case '\t':
										if buffer[position] != rune('\t') {
											goto l805
										}
										position++
										break
									case '\n':
										if buffer[position] != rune('\n') {
											goto l805
										}
										position++
										break
									default:
										if buffer[position] != rune(' ') {
											goto l805
										}
										position++
										break
									}
								}

								depth--
								add(ruleSPACE, position815)
							}
							break
						}
					}

					goto l804
				l805:
					position, tokenIndex, depth = position805, tokenIndex805, depth805
				}
				depth--
				add(rule_, position803)
			}
			return true
		},
		/* 68 COMMENT_TRAIL <- <('-' '-' (!'\n' .)*)> */
		nil,
		/* 69 COMMENT_BLOCK <- <('/' '*' (!('*' '/') .)* ('*' '/'))> */
		nil,
		/* 70 KEY <- <!ID_CONT> */
		func() bool {
			position819, tokenIndex819, depth819 := position, tokenIndex, depth
			{
				position820 := position
				depth++
				{
					position821, tokenIndex821, depth821 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l821
					}
					goto l819
				l821:
					position, tokenIndex, depth = position821, tokenIndex821, depth821
				}
				depth--
				add(ruleKEY, position820)
			}
			return true
		l819:
			position, tokenIndex, depth = position819, tokenIndex819, depth819
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
		/* 76 Action3 <- <{ p.errorHere(token.begin, `expected string literal to follow keyword "match"`) }> */
		nil,
		/* 77 Action4 <- <{ p.addMatchClause() }> */
		nil,
		/* 78 Action5 <- <{ p.errorHere(token.begin, `expected "where" to follow keyword "metrics" in "describe metrics" command`) }> */
		nil,
		/* 79 Action6 <- <{ p.errorHere(token.begin, `expected tag key to follow keyword "where" in "describe metrics" command`) }> */
		nil,
		/* 80 Action7 <- <{ p.errorHere(token.begin, `expected "=" to follow keyword "where" in "describe metrics" command`) }> */
		nil,
		/* 81 Action8 <- <{ p.errorHere(token.begin, `expected string literal to follow "=" in "describe metrics" command`) }> */
		nil,
		/* 82 Action9 <- <{ p.makeDescribeMetrics() }> */
		nil,
		nil,
		/* 84 Action10 <- <{ p.pushString(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 85 Action11 <- <{ p.errorHere(token.begin, `expected metric name to follow "describe" in "describe" command`) }> */
		nil,
		/* 86 Action12 <- <{ p.makeDescribe() }> */
		nil,
		/* 87 Action13 <- <{ p.addEvaluationContext() }> */
		nil,
		/* 88 Action14 <- <{ p.addPropertyKey(buffer[begin:end]) }> */
		nil,
		/* 89 Action15 <- <{
		   p.addPropertyValue(buffer[begin:end]) }> */
		nil,
		/* 90 Action16 <- <{ p.errorHere(token.begin, `expected property value to follow property key`) }> */
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
		/* 96 Action22 <- <{ p.errorHere(token.begin, `expected expression to follow ","`) }> */
		nil,
		/* 97 Action23 <- <{ p.appendExpression() }> */
		nil,
		/* 98 Action24 <- <{ p.addOperatorLiteral("+") }> */
		nil,
		/* 99 Action25 <- <{ p.addOperatorLiteral("-") }> */
		nil,
		/* 100 Action26 <- <{ p.errorHere(token.begin, `expected expression to follow operator "+" or "-"`) }> */
		nil,
		/* 101 Action27 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 102 Action28 <- <{ p.addOperatorLiteral("/") }> */
		nil,
		/* 103 Action29 <- <{ p.addOperatorLiteral("*") }> */
		nil,
		/* 104 Action30 <- <{ p.errorHere(token.begin, `expected expression to follow operator "*" or "/"`) }> */
		nil,
		/* 105 Action31 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 106 Action32 <- <{ p.errorHere(token.begin, `expected function name to follow pipe "|"`) }> */
		nil,
		/* 107 Action33 <- <{ p.pushString(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 108 Action34 <- <{p.addExpressionList()}> */
		nil,
		/* 109 Action35 <- <{ p.errorHere(token.begin, `expected ")" to close "(" opened in pipe function call`) }> */
		nil,
		/* 110 Action36 <- <{
		   p.addExpressionList()
		   p.addGroupBy()
		 }> */
		nil,
		/* 111 Action37 <- <{ p.addPipeExpression() }> */
		nil,
		/* 112 Action38 <- <{ p.errorHere(token.begin, `expected expression to follow "("`) }> */
		nil,
		/* 113 Action39 <- <{ p.errorHere(token.begin, `expected ")" to close "("`) }> */
		nil,
		/* 114 Action40 <- <{ p.addDurationNode(text) }> */
		nil,
		/* 115 Action41 <- <{ p.addNumberNode(buffer[begin:end]) }> */
		nil,
		/* 116 Action42 <- <{ p.addStringNode(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 117 Action43 <- <{ p.errorHere(token.begin, `expected "}> */
		nil,
		/* 118 Action44 <- <{" opened for annotation`) }> */
		nil,
		/* 119 Action45 <- <{ p.addAnnotationExpression(buffer[begin:end]) }> */
		nil,
		/* 120 Action46 <- <{ p.addGroupBy() }> */
		nil,
		/* 121 Action47 <- <{ p.pushString(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 122 Action48 <- <{ p.errorHere(token.begin, `expected expression list to follow "(" in function call`) }> */
		nil,
		/* 123 Action49 <- <{ p.errorHere(token.begin, `expected ")" to close "(" opened by function call`) }> */
		nil,
		/* 124 Action50 <- <{ p.addFunctionInvocation() }> */
		nil,
		/* 125 Action51 <- <{ p.pushString(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 126 Action52 <- <{ p.errorHere(token.begin, `expected predicate to follow "[" after metric`) }> */
		nil,
		/* 127 Action53 <- <{ p.errorHere(token.begin, `expected "]" to close "[" opened to apply predicate`) }> */
		nil,
		/* 128 Action54 <- <{ p.addNullPredicate() }> */
		nil,
		/* 129 Action55 <- <{ p.addMetricExpression() }> */
		nil,
		/* 130 Action56 <- <{ p.errorHere(token.begin, `expected keyword "by" to follow keyword "group" in "group by" clause`) }> */
		nil,
		/* 131 Action57 <- <{ p.errorHere(token.begin, `expected tag key identifier to follow "group by" keywords in "group by" clause`) }> */
		nil,
		/* 132 Action58 <- <{ p.appendGroupBy(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 133 Action59 <- <{ p.errorHere(token.begin, `expected tag key identifier to follow "," in "group by" clause`) }> */
		nil,
		/* 134 Action60 <- <{ p.appendGroupBy(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 135 Action61 <- <{ p.errorHere(token.begin, `expected keyword "by" to follow keyword "collapse" in "collapse by" clause`) }> */
		nil,
		/* 136 Action62 <- <{ p.errorHere(token.begin, `expected tag key identifier to follow "collapse by" keywords in "collapse by" clause`) }> */
		nil,
		/* 137 Action63 <- <{ p.appendCollapseBy(unescapeLiteral(text)) }> */
		nil,
		/* 138 Action64 <- <{ p.errorHere(token.begin, `expected tag key identifier to follow "," in "collapse by" clause`) }> */
		nil,
		/* 139 Action65 <- <{ p.appendCollapseBy(unescapeLiteral(text)) }> */
		nil,
		/* 140 Action66 <- <{ p.errorHere(token.begin, `expected predicate to follow "where" keyword`) }> */
		nil,
		/* 141 Action67 <- <{ p.errorHere(token.begin, `expected predicate to follow "or" operator`) }> */
		nil,
		/* 142 Action68 <- <{ p.addOrPredicate() }> */
		nil,
		/* 143 Action69 <- <{ p.errorHere(token.begin, `expected predicate to follow "and" operator`) }> */
		nil,
		/* 144 Action70 <- <{ p.addAndPredicate() }> */
		nil,
		/* 145 Action71 <- <{ p.errorHere(token.begin, `expected predicate to follow "not" operator`) }> */
		nil,
		/* 146 Action72 <- <{ p.addNotPredicate() }> */
		nil,
		/* 147 Action73 <- <{ p.errorHere(token.begin, `expected predicate to follow "("`) }> */
		nil,
		/* 148 Action74 <- <{ p.errorHere(token.begin, `expected ")" to close "(" opened in predicate`) }> */
		nil,
		/* 149 Action75 <- <{ p.errorHere(token.begin, `expected string literal to follow "="`) }> */
		nil,
		/* 150 Action76 <- <{ p.addLiteralMatcher() }> */
		nil,
		/* 151 Action77 <- <{ p.errorHere(token.begin, `expected string literal to follow "!="`) }> */
		nil,
		/* 152 Action78 <- <{ p.addLiteralMatcher() }> */
		nil,
		/* 153 Action79 <- <{ p.addNotPredicate() }> */
		nil,
		/* 154 Action80 <- <{ p.errorHere(token.begin, `expected regex string literal to follow "match"`) }> */
		nil,
		/* 155 Action81 <- <{ p.addRegexMatcher() }> */
		nil,
		/* 156 Action82 <- <{ p.errorHere(token.begin, `expected string literal list to follow "in" keyword`) }> */
		nil,
		/* 157 Action83 <- <{ p.addListMatcher() }> */
		nil,
		/* 158 Action84 <- <{ p.errorHere(token.begin, `expected "=", "!=", "match", or "in" to follow tag key in predicate`) }> */
		nil,
		/* 159 Action85 <- <{ p.pushString(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 160 Action86 <- <{ p.addLiteralList() }> */
		nil,
		/* 161 Action87 <- <{ p.errorHere(token.begin, `expected string literal to follow "(" in literal list`) }> */
		nil,
		/* 162 Action88 <- <{ p.errorHere(token.begin, `expected string literal to follow "," in literal list`) }> */
		nil,
		/* 163 Action89 <- <{ p.errorHere(token.begin, `expected ")" to close "(" for literal list`) }> */
		nil,
		/* 164 Action90 <- <{ p.appendLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 165 Action91 <- <{ p.addTagLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 166 Action92 <- <{ p.errorHere(token.begin, "expected \"`\" to end identifier") }> */
		nil,
		/* 167 Action93 <- <{ p.errorHere(token.begin, `expected identifier segment to follow "."`) }> */
		nil,
		/* 168 Action94 <- <{ p.errorHere(token.begin, `expected keyword "by" to follow keyword "sample"`) }> */
		nil,
		/* 169 Action95 <- <{ p.errorHere(token.begin, `expected "'" to close string`) }> */
		nil,
		/* 170 Action96 <- <{ p.errorHere(token.begin, `expected '"' to close string`) }> */
		nil,
		/* 171 Action97 <- <{ p.errorHere(token.begin, "expected \"\\\", \"'\", \"`\", or '\"' to follow \"\\\" in string literal") }> */
		nil,
		/* 172 Action98 <- <{ p.errorHere(token.begin, `expected exponent`) }> */
		nil,
	}
	p.rules = _rules
}
