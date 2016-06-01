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
	rulePegText
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
	"PegText",
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
	rules  [171]func() bool
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
			p.errorHere(position, `expected string literal to follow keyword "match"`)
		case ruleAction4:
			p.addMatchClause()
		case ruleAction5:
			p.errorHere(position, `expected "where" to follow keyword "metrics" in "describe metrics" command`)
		case ruleAction6:
			p.errorHere(position, `expected "=" to follow keyword "where" in "describe metrics" command`)
		case ruleAction7:
			p.errorHere(position, `expected string literal to follow "=" in "describe metrics" command`)
		case ruleAction8:
			p.makeDescribeMetrics()
		case ruleAction9:
			p.pushString(unescapeLiteral(buffer[begin:end]))
		case ruleAction10:
			p.errorHere(position, `expected metric name to follow "describe" in "describe" command`)
		case ruleAction11:
			p.makeDescribe()
		case ruleAction12:
			p.addEvaluationContext()
		case ruleAction13:
			p.addPropertyKey(buffer[begin:end])
		case ruleAction14:
			p.addPropertyValue(buffer[begin:end])
		case ruleAction15:
			p.errorHere(position, `expected property value to follow property key`)
		case ruleAction16:
			p.insertPropertyKeyValue()
		case ruleAction17:
			p.checkPropertyClause()
		case ruleAction18:
			p.addNullPredicate()
		case ruleAction19:
			p.addExpressionList()
		case ruleAction20:
			p.appendExpression()
		case ruleAction21:
			p.errorHere(position, `expected expression to follow ","`)
		case ruleAction22:
			p.appendExpression()
		case ruleAction23:
			p.addOperatorLiteral("+")
		case ruleAction24:
			p.addOperatorLiteral("-")
		case ruleAction25:
			p.errorHere(position, `expected expression to follow operator "+" or "-"`)
		case ruleAction26:
			p.addOperatorFunction()
		case ruleAction27:
			p.addOperatorLiteral("/")
		case ruleAction28:
			p.addOperatorLiteral("*")
		case ruleAction29:
			p.errorHere(position, `expected expression to follow operator "*" or "/"`)
		case ruleAction30:
			p.addOperatorFunction()
		case ruleAction31:
			p.errorHere(position, `expected function name to follow pipe "|"`)
		case ruleAction32:
			p.pushString(unescapeLiteral(buffer[begin:end]))
		case ruleAction33:
			p.addExpressionList()
		case ruleAction34:
			p.errorHere(position, `expected ")" to close "(" opened in pipe function call`)
		case ruleAction35:

			p.addExpressionList()
			p.addGroupBy()

		case ruleAction36:
			p.addPipeExpression()
		case ruleAction37:
			p.errorHere(position, `expected expression to follow "("`)
		case ruleAction38:
			p.errorHere(position, `expected ")" to close "("`)
		case ruleAction39:
			p.addDurationNode(text)
		case ruleAction40:
			p.addNumberNode(buffer[begin:end])
		case ruleAction41:
			p.addStringNode(unescapeLiteral(buffer[begin:end]))
		case ruleAction42:
			p.errorHere(position, `expected "
		case ruleAction43:
			" opened for annotation`)
		case ruleAction44:
			p.addAnnotationExpression(buffer[begin:end])
		case ruleAction45:
			p.addGroupBy()
		case ruleAction46:
			p.pushString(unescapeLiteral(buffer[begin:end]))
		case ruleAction47:
			p.errorHere(position, `expected ")" to close "(" opened by function call`)
		case ruleAction48:
			p.addFunctionInvocation()
		case ruleAction49:
			p.pushString(unescapeLiteral(buffer[begin:end]))
		case ruleAction50:
			p.errorHere(position, `expected predicate to follow "[" after metric`)
		case ruleAction51:
			p.errorHere(position, `expected "]" to close "[" opened to apply predicate`)
		case ruleAction52:
			p.addNullPredicate()
		case ruleAction53:
			p.addMetricExpression()
		case ruleAction54:
			p.errorHere(position, `expected keyword "by" to follow keyword "group" in "group by" clause`)
		case ruleAction55:
			p.errorHere(position, `expected tag key identifier to follow "group by" keywords in "group by" clause`)
		case ruleAction56:
			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		case ruleAction57:
			p.errorHere(position, `expected tag key identifier to follow "," in "group by" clause`)
		case ruleAction58:
			p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		case ruleAction59:
			p.errorHere(position, `expected keyword "by" to follow keyword "collapse" in "collapse by" clause`)
		case ruleAction60:
			p.errorHere(position, `expected tag key identifier to follow "collapse by" keywords in "collapse by" clause`)
		case ruleAction61:
			p.appendCollapseBy(unescapeLiteral(text))
		case ruleAction62:
			p.errorHere(position, `expected tag key identifier to follow "," in "collapse by" clause`)
		case ruleAction63:
			p.appendCollapseBy(unescapeLiteral(text))
		case ruleAction64:
			p.errorHere(position, `expected predicate to follow "where" keyword`)
		case ruleAction65:
			p.errorHere(position, `expected predicate to follow "or" operator`)
		case ruleAction66:
			p.addOrPredicate()
		case ruleAction67:
			p.errorHere(position, `expected predicate to follow "and" operator`)
		case ruleAction68:
			p.addAndPredicate()
		case ruleAction69:
			p.errorHere(position, `expected predicate to follow "not" operator`)
		case ruleAction70:
			p.addNotPredicate()
		case ruleAction71:
			p.errorHere(position, `expected predicate to follow "("`)
		case ruleAction72:
			p.errorHere(position, `expected ")" to close "(" opened in predicate`)
		case ruleAction73:
			p.errorHere(position, `expected string literal to follow "="`)
		case ruleAction74:
			p.addLiteralMatcher()
		case ruleAction75:
			p.errorHere(position, `expected string literal to follow "!="`)
		case ruleAction76:
			p.addLiteralMatcher()
		case ruleAction77:
			p.addNotPredicate()
		case ruleAction78:
			p.errorHere(position, `expected regex string literal to follow "match"`)
		case ruleAction79:
			p.addRegexMatcher()
		case ruleAction80:
			p.errorHere(position, `expected string literal list to follow "in" keyword`)
		case ruleAction81:
			p.addListMatcher()
		case ruleAction82:
			p.errorHere(position, `expected "=", "!=", "match", or "in" to follow tag key in predicate`)
		case ruleAction83:
			p.pushString(unescapeLiteral(buffer[begin:end]))
		case ruleAction84:
			p.addLiteralList()
		case ruleAction85:
			p.errorHere(position, `expected string literal to follow "(" in literal list`)
		case ruleAction86:
			p.errorHere(position, `expected string literal to follow "," in literal list`)
		case ruleAction87:
			p.errorHere(position, `expected ")" to close "(" for literal list`)
		case ruleAction88:
			p.appendLiteral(unescapeLiteral(buffer[begin:end]))
		case ruleAction89:
			p.addTagLiteral(unescapeLiteral(buffer[begin:end]))
		case ruleAction90:
			p.errorHere(position, "expected \"`\" to end identifier")
		case ruleAction91:
			p.errorHere(position, `expected identifier segment to follow "."`)
		case ruleAction92:
			p.errorHere(position, `expected keyword "by" to follow keyword "sample"`)
		case ruleAction93:
			p.errorHere(position, `expected "'" to close string`)
		case ruleAction94:
			p.errorHere(position, `expected '"' to close string`)
		case ruleAction95:
			p.errorHere(position, "expected \"\\\" or \"`\" to follow escaping backslash")
		case ruleAction96:
			p.errorHere(position, `expected exponent`)

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
								add(ruleAction12, position)
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
									add(ruleAction13, position)
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
										add(ruleAction14, position)
									}
									goto l24
								l25:
									position, tokenIndex, depth = position24, tokenIndex24, depth24
									{
										add(ruleAction15, position)
									}
								}
							l24:
								{
									add(ruleAction16, position)
								}
								goto l21
							l22:
								position, tokenIndex, depth = position22, tokenIndex22, depth22
							}
							{
								add(ruleAction17, position)
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
									if !_rules[ruletagName]() {
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
									if !_rules[rule_]() {
										goto l124
									}
									if buffer[position] != rune('=') {
										goto l124
									}
									position++
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
									if !_rules[ruleliteralString]() {
										goto l127
									}
									goto l126
								l127:
									position, tokenIndex, depth = position126, tokenIndex126, depth126
									{
										add(ruleAction7, position)
									}
								}
							l126:
								{
									add(ruleAction8, position)
								}
								depth--
								add(ruledescribeMetrics, position95)
							}
							goto l65
						l94:
							position, tokenIndex, depth = position65, tokenIndex65, depth65
							{
								position130 := position
								depth++
								{
									position131, tokenIndex131, depth131 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l132
									}
									{
										position133 := position
										depth++
										{
											position134 := position
											depth++
											if !_rules[ruleIDENTIFIER]() {
												goto l132
											}
											depth--
											add(ruleMETRIC_NAME, position134)
										}
										depth--
										add(rulePegText, position133)
									}
									{
										add(ruleAction9, position)
									}
									goto l131
								l132:
									position, tokenIndex, depth = position131, tokenIndex131, depth131
									{
										add(ruleAction10, position)
									}
								}
							l131:
								if !_rules[ruleoptionalPredicateClause]() {
									goto l0
								}
								{
									add(ruleAction11, position)
								}
								depth--
								add(ruledescribeSingleStmt, position130)
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
					position138, tokenIndex138, depth138 := position, tokenIndex, depth
					if !matchDot() {
						goto l138
					}
					goto l0
				l138:
					position, tokenIndex, depth = position138, tokenIndex138, depth138
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
		/* 6 describeMetrics <- <(_ (('m' / 'M') ('e' / 'E') ('t' / 'T') ('r' / 'R') ('i' / 'I') ('c' / 'C') ('s' / 'S')) KEY ((_ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY tagName) / Action5) ((_ '=') / Action6) (literalString / Action7) Action8)> */
		nil,
		/* 7 describeSingleStmt <- <(((_ <METRIC_NAME> Action9) / Action10) optionalPredicateClause Action11)> */
		nil,
		/* 8 propertyClause <- <(Action12 (_ PROPERTY_KEY Action13 ((_ PROPERTY_VALUE Action14) / Action15) Action16)* Action17)> */
		nil,
		/* 9 optionalPredicateClause <- <(predicateClause / Action18)> */
		func() bool {
			{
				position148 := position
				depth++
				{
					position149, tokenIndex149, depth149 := position, tokenIndex, depth
					{
						position151 := position
						depth++
						if !_rules[rule_]() {
							goto l150
						}
						{
							position152, tokenIndex152, depth152 := position, tokenIndex, depth
							if buffer[position] != rune('w') {
								goto l153
							}
							position++
							goto l152
						l153:
							position, tokenIndex, depth = position152, tokenIndex152, depth152
							if buffer[position] != rune('W') {
								goto l150
							}
							position++
						}
					l152:
						{
							position154, tokenIndex154, depth154 := position, tokenIndex, depth
							if buffer[position] != rune('h') {
								goto l155
							}
							position++
							goto l154
						l155:
							position, tokenIndex, depth = position154, tokenIndex154, depth154
							if buffer[position] != rune('H') {
								goto l150
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
								goto l150
							}
							position++
						}
					l156:
						{
							position158, tokenIndex158, depth158 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l159
							}
							position++
							goto l158
						l159:
							position, tokenIndex, depth = position158, tokenIndex158, depth158
							if buffer[position] != rune('R') {
								goto l150
							}
							position++
						}
					l158:
						{
							position160, tokenIndex160, depth160 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l161
							}
							position++
							goto l160
						l161:
							position, tokenIndex, depth = position160, tokenIndex160, depth160
							if buffer[position] != rune('E') {
								goto l150
							}
							position++
						}
					l160:
						if !_rules[ruleKEY]() {
							goto l150
						}
						{
							position162, tokenIndex162, depth162 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l163
							}
							if !_rules[rulepredicate_1]() {
								goto l163
							}
							goto l162
						l163:
							position, tokenIndex, depth = position162, tokenIndex162, depth162
							{
								add(ruleAction64, position)
							}
						}
					l162:
						depth--
						add(rulepredicateClause, position151)
					}
					goto l149
				l150:
					position, tokenIndex, depth = position149, tokenIndex149, depth149
					{
						add(ruleAction18, position)
					}
				}
			l149:
				depth--
				add(ruleoptionalPredicateClause, position148)
			}
			return true
		},
		/* 10 expressionList <- <(Action19 expression_start Action20 (_ COMMA (expression_start / Action21) Action22)*)> */
		func() bool {
			position166, tokenIndex166, depth166 := position, tokenIndex, depth
			{
				position167 := position
				depth++
				{
					add(ruleAction19, position)
				}
				if !_rules[ruleexpression_start]() {
					goto l166
				}
				{
					add(ruleAction20, position)
				}
			l170:
				{
					position171, tokenIndex171, depth171 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l171
					}
					if !_rules[ruleCOMMA]() {
						goto l171
					}
					{
						position172, tokenIndex172, depth172 := position, tokenIndex, depth
						if !_rules[ruleexpression_start]() {
							goto l173
						}
						goto l172
					l173:
						position, tokenIndex, depth = position172, tokenIndex172, depth172
						{
							add(ruleAction21, position)
						}
					}
				l172:
					{
						add(ruleAction22, position)
					}
					goto l170
				l171:
					position, tokenIndex, depth = position171, tokenIndex171, depth171
				}
				depth--
				add(ruleexpressionList, position167)
			}
			return true
		l166:
			position, tokenIndex, depth = position166, tokenIndex166, depth166
			return false
		},
		/* 11 expression_start <- <(expression_sum add_pipe)> */
		func() bool {
			position176, tokenIndex176, depth176 := position, tokenIndex, depth
			{
				position177 := position
				depth++
				{
					position178 := position
					depth++
					if !_rules[ruleexpression_product]() {
						goto l176
					}
				l179:
					{
						position180, tokenIndex180, depth180 := position, tokenIndex, depth
						if !_rules[ruleadd_pipe]() {
							goto l180
						}
						{
							position181, tokenIndex181, depth181 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l182
							}
							{
								position183 := position
								depth++
								if buffer[position] != rune('+') {
									goto l182
								}
								position++
								depth--
								add(ruleOP_ADD, position183)
							}
							{
								add(ruleAction23, position)
							}
							goto l181
						l182:
							position, tokenIndex, depth = position181, tokenIndex181, depth181
							if !_rules[rule_]() {
								goto l180
							}
							{
								position185 := position
								depth++
								if buffer[position] != rune('-') {
									goto l180
								}
								position++
								depth--
								add(ruleOP_SUB, position185)
							}
							{
								add(ruleAction24, position)
							}
						}
					l181:
						{
							position187, tokenIndex187, depth187 := position, tokenIndex, depth
							if !_rules[ruleexpression_product]() {
								goto l188
							}
							goto l187
						l188:
							position, tokenIndex, depth = position187, tokenIndex187, depth187
							{
								add(ruleAction25, position)
							}
						}
					l187:
						{
							add(ruleAction26, position)
						}
						goto l179
					l180:
						position, tokenIndex, depth = position180, tokenIndex180, depth180
					}
					depth--
					add(ruleexpression_sum, position178)
				}
				if !_rules[ruleadd_pipe]() {
					goto l176
				}
				depth--
				add(ruleexpression_start, position177)
			}
			return true
		l176:
			position, tokenIndex, depth = position176, tokenIndex176, depth176
			return false
		},
		/* 12 expression_sum <- <(expression_product (add_pipe ((_ OP_ADD Action23) / (_ OP_SUB Action24)) (expression_product / Action25) Action26)*)> */
		nil,
		/* 13 expression_product <- <(expression_atom (add_pipe ((_ OP_DIV Action27) / (_ OP_MULT Action28)) (expression_atom / Action29) Action30)*)> */
		func() bool {
			position192, tokenIndex192, depth192 := position, tokenIndex, depth
			{
				position193 := position
				depth++
				if !_rules[ruleexpression_atom]() {
					goto l192
				}
			l194:
				{
					position195, tokenIndex195, depth195 := position, tokenIndex, depth
					if !_rules[ruleadd_pipe]() {
						goto l195
					}
					{
						position196, tokenIndex196, depth196 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l197
						}
						{
							position198 := position
							depth++
							if buffer[position] != rune('/') {
								goto l197
							}
							position++
							depth--
							add(ruleOP_DIV, position198)
						}
						{
							add(ruleAction27, position)
						}
						goto l196
					l197:
						position, tokenIndex, depth = position196, tokenIndex196, depth196
						if !_rules[rule_]() {
							goto l195
						}
						{
							position200 := position
							depth++
							if buffer[position] != rune('*') {
								goto l195
							}
							position++
							depth--
							add(ruleOP_MULT, position200)
						}
						{
							add(ruleAction28, position)
						}
					}
				l196:
					{
						position202, tokenIndex202, depth202 := position, tokenIndex, depth
						if !_rules[ruleexpression_atom]() {
							goto l203
						}
						goto l202
					l203:
						position, tokenIndex, depth = position202, tokenIndex202, depth202
						{
							add(ruleAction29, position)
						}
					}
				l202:
					{
						add(ruleAction30, position)
					}
					goto l194
				l195:
					position, tokenIndex, depth = position195, tokenIndex195, depth195
				}
				depth--
				add(ruleexpression_product, position193)
			}
			return true
		l192:
			position, tokenIndex, depth = position192, tokenIndex192, depth192
			return false
		},
		/* 14 add_one_pipe <- <(_ OP_PIPE ((_ <IDENTIFIER>) / Action31) Action32 ((_ PAREN_OPEN (expressionList / Action33) optionalGroupBy ((_ PAREN_CLOSE) / Action34)) / Action35) Action36 expression_annotation)> */
		nil,
		/* 15 add_pipe <- <add_one_pipe*> */
		func() bool {
			{
				position208 := position
				depth++
			l209:
				{
					position210, tokenIndex210, depth210 := position, tokenIndex, depth
					{
						position211 := position
						depth++
						if !_rules[rule_]() {
							goto l210
						}
						{
							position212 := position
							depth++
							if buffer[position] != rune('|') {
								goto l210
							}
							position++
							depth--
							add(ruleOP_PIPE, position212)
						}
						{
							position213, tokenIndex213, depth213 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l214
							}
							{
								position215 := position
								depth++
								if !_rules[ruleIDENTIFIER]() {
									goto l214
								}
								depth--
								add(rulePegText, position215)
							}
							goto l213
						l214:
							position, tokenIndex, depth = position213, tokenIndex213, depth213
							{
								add(ruleAction31, position)
							}
						}
					l213:
						{
							add(ruleAction32, position)
						}
						{
							position218, tokenIndex218, depth218 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l219
							}
							if !_rules[rulePAREN_OPEN]() {
								goto l219
							}
							{
								position220, tokenIndex220, depth220 := position, tokenIndex, depth
								if !_rules[ruleexpressionList]() {
									goto l221
								}
								goto l220
							l221:
								position, tokenIndex, depth = position220, tokenIndex220, depth220
								{
									add(ruleAction33, position)
								}
							}
						l220:
							if !_rules[ruleoptionalGroupBy]() {
								goto l219
							}
							{
								position223, tokenIndex223, depth223 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l224
								}
								if !_rules[rulePAREN_CLOSE]() {
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
							goto l218
						l219:
							position, tokenIndex, depth = position218, tokenIndex218, depth218
							{
								add(ruleAction35, position)
							}
						}
					l218:
						{
							add(ruleAction36, position)
						}
						if !_rules[ruleexpression_annotation]() {
							goto l210
						}
						depth--
						add(ruleadd_one_pipe, position211)
					}
					goto l209
				l210:
					position, tokenIndex, depth = position210, tokenIndex210, depth210
				}
				depth--
				add(ruleadd_pipe, position208)
			}
			return true
		},
		/* 16 expression_atom <- <(expression_atom_raw expression_annotation)> */
		func() bool {
			position228, tokenIndex228, depth228 := position, tokenIndex, depth
			{
				position229 := position
				depth++
				{
					position230 := position
					depth++
					{
						position231, tokenIndex231, depth231 := position, tokenIndex, depth
						{
							position233 := position
							depth++
							if !_rules[rule_]() {
								goto l232
							}
							{
								position234 := position
								depth++
								if !_rules[ruleIDENTIFIER]() {
									goto l232
								}
								depth--
								add(rulePegText, position234)
							}
							{
								add(ruleAction46, position)
							}
							if !_rules[rule_]() {
								goto l232
							}
							if !_rules[rulePAREN_OPEN]() {
								goto l232
							}
							if !_rules[ruleexpressionList]() {
								goto l232
							}
							if !_rules[ruleoptionalGroupBy]() {
								goto l232
							}
							{
								position236, tokenIndex236, depth236 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l237
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l237
								}
								goto l236
							l237:
								position, tokenIndex, depth = position236, tokenIndex236, depth236
								{
									add(ruleAction47, position)
								}
							}
						l236:
							{
								add(ruleAction48, position)
							}
							depth--
							add(ruleexpression_function, position233)
						}
						goto l231
					l232:
						position, tokenIndex, depth = position231, tokenIndex231, depth231
						{
							position241 := position
							depth++
							if !_rules[rule_]() {
								goto l240
							}
							{
								position242 := position
								depth++
								if !_rules[ruleIDENTIFIER]() {
									goto l240
								}
								depth--
								add(rulePegText, position242)
							}
							{
								add(ruleAction49, position)
							}
							{
								position244, tokenIndex244, depth244 := position, tokenIndex, depth
								{
									position246, tokenIndex246, depth246 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l247
									}
									if buffer[position] != rune('[') {
										goto l247
									}
									position++
									{
										position248, tokenIndex248, depth248 := position, tokenIndex, depth
										if !_rules[rulepredicate_1]() {
											goto l249
										}
										goto l248
									l249:
										position, tokenIndex, depth = position248, tokenIndex248, depth248
										{
											add(ruleAction50, position)
										}
									}
								l248:
									{
										position251, tokenIndex251, depth251 := position, tokenIndex, depth
										if !_rules[rule_]() {
											goto l252
										}
										if buffer[position] != rune(']') {
											goto l252
										}
										position++
										goto l251
									l252:
										position, tokenIndex, depth = position251, tokenIndex251, depth251
										{
											add(ruleAction51, position)
										}
									}
								l251:
									goto l246
								l247:
									position, tokenIndex, depth = position246, tokenIndex246, depth246
									{
										add(ruleAction52, position)
									}
								}
							l246:
								goto l245

								position, tokenIndex, depth = position244, tokenIndex244, depth244
							}
						l245:
							{
								add(ruleAction53, position)
							}
							depth--
							add(ruleexpression_metric, position241)
						}
						goto l231
					l240:
						position, tokenIndex, depth = position231, tokenIndex231, depth231
						if !_rules[rule_]() {
							goto l256
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l256
						}
						{
							position257, tokenIndex257, depth257 := position, tokenIndex, depth
							if !_rules[ruleexpression_start]() {
								goto l258
							}
							goto l257
						l258:
							position, tokenIndex, depth = position257, tokenIndex257, depth257
							{
								add(ruleAction37, position)
							}
						}
					l257:
						{
							position260, tokenIndex260, depth260 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l261
							}
							if !_rules[rulePAREN_CLOSE]() {
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
						goto l231
					l256:
						position, tokenIndex, depth = position231, tokenIndex231, depth231
						if !_rules[rule_]() {
							goto l263
						}
						{
							position264 := position
							depth++
							{
								position265 := position
								depth++
								if !_rules[ruleNUMBER]() {
									goto l263
								}
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l263
								}
								position++
							l266:
								{
									position267, tokenIndex267, depth267 := position, tokenIndex, depth
									if c := buffer[position]; c < rune('a') || c > rune('z') {
										goto l267
									}
									position++
									goto l266
								l267:
									position, tokenIndex, depth = position267, tokenIndex267, depth267
								}
								if !_rules[ruleKEY]() {
									goto l263
								}
								depth--
								add(ruleDURATION, position265)
							}
							depth--
							add(rulePegText, position264)
						}
						{
							add(ruleAction39, position)
						}
						goto l231
					l263:
						position, tokenIndex, depth = position231, tokenIndex231, depth231
						if !_rules[rule_]() {
							goto l269
						}
						{
							position270 := position
							depth++
							if !_rules[ruleNUMBER]() {
								goto l269
							}
							depth--
							add(rulePegText, position270)
						}
						{
							add(ruleAction40, position)
						}
						goto l231
					l269:
						position, tokenIndex, depth = position231, tokenIndex231, depth231
						if !_rules[rule_]() {
							goto l228
						}
						if !_rules[ruleSTRING]() {
							goto l228
						}
						{
							add(ruleAction41, position)
						}
					}
				l231:
					depth--
					add(ruleexpression_atom_raw, position230)
				}
				if !_rules[ruleexpression_annotation]() {
					goto l228
				}
				depth--
				add(ruleexpression_atom, position229)
			}
			return true
		l228:
			position, tokenIndex, depth = position228, tokenIndex228, depth228
			return false
		},
		/* 17 expression_atom_raw <- <(expression_function / expression_metric / (_ PAREN_OPEN (expression_start / Action37) ((_ PAREN_CLOSE) / Action38)) / (_ <DURATION> Action39) / (_ <NUMBER> Action40) / (_ STRING Action41))> */
		nil,
		/* 18 expression_annotation_required <- <(_ '{' <(!'}' .)*> ('}' / (Action42 (' ' ('t' / 'T') ('o' / 'O') ' ' ('c' / 'C') ('l' / 'L') ('o' / 'O') ('s' / 'S') ('e' / 'E') ' ') Action43)) Action44)> */
		nil,
		/* 19 expression_annotation <- <expression_annotation_required?> */
		func() bool {
			{
				position276 := position
				depth++
				{
					position277, tokenIndex277, depth277 := position, tokenIndex, depth
					{
						position279 := position
						depth++
						if !_rules[rule_]() {
							goto l277
						}
						if buffer[position] != rune('{') {
							goto l277
						}
						position++
						{
							position280 := position
							depth++
						l281:
							{
								position282, tokenIndex282, depth282 := position, tokenIndex, depth
								{
									position283, tokenIndex283, depth283 := position, tokenIndex, depth
									if buffer[position] != rune('}') {
										goto l283
									}
									position++
									goto l282
								l283:
									position, tokenIndex, depth = position283, tokenIndex283, depth283
								}
								if !matchDot() {
									goto l282
								}
								goto l281
							l282:
								position, tokenIndex, depth = position282, tokenIndex282, depth282
							}
							depth--
							add(rulePegText, position280)
						}
						{
							position284, tokenIndex284, depth284 := position, tokenIndex, depth
							if buffer[position] != rune('}') {
								goto l285
							}
							position++
							goto l284
						l285:
							position, tokenIndex, depth = position284, tokenIndex284, depth284
							{
								add(ruleAction42, position)
							}
							if buffer[position] != rune(' ') {
								goto l277
							}
							position++
							{
								position287, tokenIndex287, depth287 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l288
								}
								position++
								goto l287
							l288:
								position, tokenIndex, depth = position287, tokenIndex287, depth287
								if buffer[position] != rune('T') {
									goto l277
								}
								position++
							}
						l287:
							{
								position289, tokenIndex289, depth289 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l290
								}
								position++
								goto l289
							l290:
								position, tokenIndex, depth = position289, tokenIndex289, depth289
								if buffer[position] != rune('O') {
									goto l277
								}
								position++
							}
						l289:
							if buffer[position] != rune(' ') {
								goto l277
							}
							position++
							{
								position291, tokenIndex291, depth291 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l292
								}
								position++
								goto l291
							l292:
								position, tokenIndex, depth = position291, tokenIndex291, depth291
								if buffer[position] != rune('C') {
									goto l277
								}
								position++
							}
						l291:
							{
								position293, tokenIndex293, depth293 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l294
								}
								position++
								goto l293
							l294:
								position, tokenIndex, depth = position293, tokenIndex293, depth293
								if buffer[position] != rune('L') {
									goto l277
								}
								position++
							}
						l293:
							{
								position295, tokenIndex295, depth295 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l296
								}
								position++
								goto l295
							l296:
								position, tokenIndex, depth = position295, tokenIndex295, depth295
								if buffer[position] != rune('O') {
									goto l277
								}
								position++
							}
						l295:
							{
								position297, tokenIndex297, depth297 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l298
								}
								position++
								goto l297
							l298:
								position, tokenIndex, depth = position297, tokenIndex297, depth297
								if buffer[position] != rune('S') {
									goto l277
								}
								position++
							}
						l297:
							{
								position299, tokenIndex299, depth299 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l300
								}
								position++
								goto l299
							l300:
								position, tokenIndex, depth = position299, tokenIndex299, depth299
								if buffer[position] != rune('E') {
									goto l277
								}
								position++
							}
						l299:
							if buffer[position] != rune(' ') {
								goto l277
							}
							position++
							{
								add(ruleAction43, position)
							}
						}
					l284:
						{
							add(ruleAction44, position)
						}
						depth--
						add(ruleexpression_annotation_required, position279)
					}
					goto l278
				l277:
					position, tokenIndex, depth = position277, tokenIndex277, depth277
				}
			l278:
				depth--
				add(ruleexpression_annotation, position276)
			}
			return true
		},
		/* 20 optionalGroupBy <- <(Action45 (groupByClause / collapseByClause)?)> */
		func() bool {
			{
				position304 := position
				depth++
				{
					add(ruleAction45, position)
				}
				{
					position306, tokenIndex306, depth306 := position, tokenIndex, depth
					{
						position308, tokenIndex308, depth308 := position, tokenIndex, depth
						{
							position310 := position
							depth++
							if !_rules[rule_]() {
								goto l309
							}
							{
								position311, tokenIndex311, depth311 := position, tokenIndex, depth
								if buffer[position] != rune('g') {
									goto l312
								}
								position++
								goto l311
							l312:
								position, tokenIndex, depth = position311, tokenIndex311, depth311
								if buffer[position] != rune('G') {
									goto l309
								}
								position++
							}
						l311:
							{
								position313, tokenIndex313, depth313 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l314
								}
								position++
								goto l313
							l314:
								position, tokenIndex, depth = position313, tokenIndex313, depth313
								if buffer[position] != rune('R') {
									goto l309
								}
								position++
							}
						l313:
							{
								position315, tokenIndex315, depth315 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l316
								}
								position++
								goto l315
							l316:
								position, tokenIndex, depth = position315, tokenIndex315, depth315
								if buffer[position] != rune('O') {
									goto l309
								}
								position++
							}
						l315:
							{
								position317, tokenIndex317, depth317 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l318
								}
								position++
								goto l317
							l318:
								position, tokenIndex, depth = position317, tokenIndex317, depth317
								if buffer[position] != rune('U') {
									goto l309
								}
								position++
							}
						l317:
							{
								position319, tokenIndex319, depth319 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l320
								}
								position++
								goto l319
							l320:
								position, tokenIndex, depth = position319, tokenIndex319, depth319
								if buffer[position] != rune('P') {
									goto l309
								}
								position++
							}
						l319:
							if !_rules[ruleKEY]() {
								goto l309
							}
							{
								position321, tokenIndex321, depth321 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l322
								}
								{
									position323, tokenIndex323, depth323 := position, tokenIndex, depth
									if buffer[position] != rune('b') {
										goto l324
									}
									position++
									goto l323
								l324:
									position, tokenIndex, depth = position323, tokenIndex323, depth323
									if buffer[position] != rune('B') {
										goto l322
									}
									position++
								}
							l323:
								{
									position325, tokenIndex325, depth325 := position, tokenIndex, depth
									if buffer[position] != rune('y') {
										goto l326
									}
									position++
									goto l325
								l326:
									position, tokenIndex, depth = position325, tokenIndex325, depth325
									if buffer[position] != rune('Y') {
										goto l322
									}
									position++
								}
							l325:
								if !_rules[ruleKEY]() {
									goto l322
								}
								goto l321
							l322:
								position, tokenIndex, depth = position321, tokenIndex321, depth321
								{
									add(ruleAction54, position)
								}
							}
						l321:
							{
								position328, tokenIndex328, depth328 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l329
								}
								{
									position330 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l329
									}
									depth--
									add(rulePegText, position330)
								}
								goto l328
							l329:
								position, tokenIndex, depth = position328, tokenIndex328, depth328
								{
									add(ruleAction55, position)
								}
							}
						l328:
							{
								add(ruleAction56, position)
							}
						l333:
							{
								position334, tokenIndex334, depth334 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l334
								}
								if !_rules[ruleCOMMA]() {
									goto l334
								}
								{
									position335, tokenIndex335, depth335 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l336
									}
									{
										position337 := position
										depth++
										if !_rules[ruleCOLUMN_NAME]() {
											goto l336
										}
										depth--
										add(rulePegText, position337)
									}
									goto l335
								l336:
									position, tokenIndex, depth = position335, tokenIndex335, depth335
									{
										add(ruleAction57, position)
									}
								}
							l335:
								{
									add(ruleAction58, position)
								}
								goto l333
							l334:
								position, tokenIndex, depth = position334, tokenIndex334, depth334
							}
							depth--
							add(rulegroupByClause, position310)
						}
						goto l308
					l309:
						position, tokenIndex, depth = position308, tokenIndex308, depth308
						{
							position340 := position
							depth++
							if !_rules[rule_]() {
								goto l306
							}
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
									goto l306
								}
								position++
							}
						l341:
							{
								position343, tokenIndex343, depth343 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l344
								}
								position++
								goto l343
							l344:
								position, tokenIndex, depth = position343, tokenIndex343, depth343
								if buffer[position] != rune('O') {
									goto l306
								}
								position++
							}
						l343:
							{
								position345, tokenIndex345, depth345 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l346
								}
								position++
								goto l345
							l346:
								position, tokenIndex, depth = position345, tokenIndex345, depth345
								if buffer[position] != rune('L') {
									goto l306
								}
								position++
							}
						l345:
							{
								position347, tokenIndex347, depth347 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l348
								}
								position++
								goto l347
							l348:
								position, tokenIndex, depth = position347, tokenIndex347, depth347
								if buffer[position] != rune('L') {
									goto l306
								}
								position++
							}
						l347:
							{
								position349, tokenIndex349, depth349 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l350
								}
								position++
								goto l349
							l350:
								position, tokenIndex, depth = position349, tokenIndex349, depth349
								if buffer[position] != rune('A') {
									goto l306
								}
								position++
							}
						l349:
							{
								position351, tokenIndex351, depth351 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l352
								}
								position++
								goto l351
							l352:
								position, tokenIndex, depth = position351, tokenIndex351, depth351
								if buffer[position] != rune('P') {
									goto l306
								}
								position++
							}
						l351:
							{
								position353, tokenIndex353, depth353 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l354
								}
								position++
								goto l353
							l354:
								position, tokenIndex, depth = position353, tokenIndex353, depth353
								if buffer[position] != rune('S') {
									goto l306
								}
								position++
							}
						l353:
							{
								position355, tokenIndex355, depth355 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l356
								}
								position++
								goto l355
							l356:
								position, tokenIndex, depth = position355, tokenIndex355, depth355
								if buffer[position] != rune('E') {
									goto l306
								}
								position++
							}
						l355:
							if !_rules[ruleKEY]() {
								goto l306
							}
							{
								position357, tokenIndex357, depth357 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l358
								}
								{
									position359, tokenIndex359, depth359 := position, tokenIndex, depth
									if buffer[position] != rune('b') {
										goto l360
									}
									position++
									goto l359
								l360:
									position, tokenIndex, depth = position359, tokenIndex359, depth359
									if buffer[position] != rune('B') {
										goto l358
									}
									position++
								}
							l359:
								{
									position361, tokenIndex361, depth361 := position, tokenIndex, depth
									if buffer[position] != rune('y') {
										goto l362
									}
									position++
									goto l361
								l362:
									position, tokenIndex, depth = position361, tokenIndex361, depth361
									if buffer[position] != rune('Y') {
										goto l358
									}
									position++
								}
							l361:
								if !_rules[ruleKEY]() {
									goto l358
								}
								goto l357
							l358:
								position, tokenIndex, depth = position357, tokenIndex357, depth357
								{
									add(ruleAction59, position)
								}
							}
						l357:
							{
								position364, tokenIndex364, depth364 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l365
								}
								{
									position366 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l365
									}
									depth--
									add(rulePegText, position366)
								}
								goto l364
							l365:
								position, tokenIndex, depth = position364, tokenIndex364, depth364
								{
									add(ruleAction60, position)
								}
							}
						l364:
							{
								add(ruleAction61, position)
							}
						l369:
							{
								position370, tokenIndex370, depth370 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l370
								}
								if !_rules[ruleCOMMA]() {
									goto l370
								}
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
									{
										add(ruleAction62, position)
									}
								}
							l371:
								{
									add(ruleAction63, position)
								}
								goto l369
							l370:
								position, tokenIndex, depth = position370, tokenIndex370, depth370
							}
							depth--
							add(rulecollapseByClause, position340)
						}
					}
				l308:
					goto l307
				l306:
					position, tokenIndex, depth = position306, tokenIndex306, depth306
				}
			l307:
				depth--
				add(ruleoptionalGroupBy, position304)
			}
			return true
		},
		/* 21 expression_function <- <(_ <IDENTIFIER> Action46 _ PAREN_OPEN expressionList optionalGroupBy ((_ PAREN_CLOSE) / Action47) Action48)> */
		nil,
		/* 22 expression_metric <- <(_ <IDENTIFIER> Action49 ((_ '[' (predicate_1 / Action50) ((_ ']') / Action51)) / Action52)? Action53)> */
		nil,
		/* 23 groupByClause <- <(_ (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P')) KEY ((_ (('b' / 'B') ('y' / 'Y')) KEY) / Action54) ((_ <COLUMN_NAME>) / Action55) Action56 (_ COMMA ((_ <COLUMN_NAME>) / Action57) Action58)*)> */
		nil,
		/* 24 collapseByClause <- <(_ (('c' / 'C') ('o' / 'O') ('l' / 'L') ('l' / 'L') ('a' / 'A') ('p' / 'P') ('s' / 'S') ('e' / 'E')) KEY ((_ (('b' / 'B') ('y' / 'Y')) KEY) / Action59) ((_ <COLUMN_NAME>) / Action60) Action61 (_ COMMA ((_ <COLUMN_NAME>) / Action62) Action63)*)> */
		nil,
		/* 25 predicateClause <- <(_ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY ((_ predicate_1) / Action64))> */
		nil,
		/* 26 predicate_1 <- <((predicate_2 _ OP_OR (predicate_1 / Action65) Action66) / predicate_2)> */
		func() bool {
			position381, tokenIndex381, depth381 := position, tokenIndex, depth
			{
				position382 := position
				depth++
				{
					position383, tokenIndex383, depth383 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l384
					}
					if !_rules[rule_]() {
						goto l384
					}
					{
						position385 := position
						depth++
						{
							position386, tokenIndex386, depth386 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l387
							}
							position++
							goto l386
						l387:
							position, tokenIndex, depth = position386, tokenIndex386, depth386
							if buffer[position] != rune('O') {
								goto l384
							}
							position++
						}
					l386:
						{
							position388, tokenIndex388, depth388 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l389
							}
							position++
							goto l388
						l389:
							position, tokenIndex, depth = position388, tokenIndex388, depth388
							if buffer[position] != rune('R') {
								goto l384
							}
							position++
						}
					l388:
						if !_rules[ruleKEY]() {
							goto l384
						}
						depth--
						add(ruleOP_OR, position385)
					}
					{
						position390, tokenIndex390, depth390 := position, tokenIndex, depth
						if !_rules[rulepredicate_1]() {
							goto l391
						}
						goto l390
					l391:
						position, tokenIndex, depth = position390, tokenIndex390, depth390
						{
							add(ruleAction65, position)
						}
					}
				l390:
					{
						add(ruleAction66, position)
					}
					goto l383
				l384:
					position, tokenIndex, depth = position383, tokenIndex383, depth383
					if !_rules[rulepredicate_2]() {
						goto l381
					}
				}
			l383:
				depth--
				add(rulepredicate_1, position382)
			}
			return true
		l381:
			position, tokenIndex, depth = position381, tokenIndex381, depth381
			return false
		},
		/* 27 predicate_2 <- <((predicate_3 _ OP_AND (predicate_2 / Action67) Action68) / predicate_3)> */
		func() bool {
			position394, tokenIndex394, depth394 := position, tokenIndex, depth
			{
				position395 := position
				depth++
				{
					position396, tokenIndex396, depth396 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l397
					}
					if !_rules[rule_]() {
						goto l397
					}
					{
						position398 := position
						depth++
						{
							position399, tokenIndex399, depth399 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l400
							}
							position++
							goto l399
						l400:
							position, tokenIndex, depth = position399, tokenIndex399, depth399
							if buffer[position] != rune('A') {
								goto l397
							}
							position++
						}
					l399:
						{
							position401, tokenIndex401, depth401 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l402
							}
							position++
							goto l401
						l402:
							position, tokenIndex, depth = position401, tokenIndex401, depth401
							if buffer[position] != rune('N') {
								goto l397
							}
							position++
						}
					l401:
						{
							position403, tokenIndex403, depth403 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l404
							}
							position++
							goto l403
						l404:
							position, tokenIndex, depth = position403, tokenIndex403, depth403
							if buffer[position] != rune('D') {
								goto l397
							}
							position++
						}
					l403:
						if !_rules[ruleKEY]() {
							goto l397
						}
						depth--
						add(ruleOP_AND, position398)
					}
					{
						position405, tokenIndex405, depth405 := position, tokenIndex, depth
						if !_rules[rulepredicate_2]() {
							goto l406
						}
						goto l405
					l406:
						position, tokenIndex, depth = position405, tokenIndex405, depth405
						{
							add(ruleAction67, position)
						}
					}
				l405:
					{
						add(ruleAction68, position)
					}
					goto l396
				l397:
					position, tokenIndex, depth = position396, tokenIndex396, depth396
					if !_rules[rulepredicate_3]() {
						goto l394
					}
				}
			l396:
				depth--
				add(rulepredicate_2, position395)
			}
			return true
		l394:
			position, tokenIndex, depth = position394, tokenIndex394, depth394
			return false
		},
		/* 28 predicate_3 <- <((_ OP_NOT (predicate_3 / Action69) Action70) / (_ PAREN_OPEN (predicate_1 / Action71) ((_ PAREN_CLOSE) / Action72)) / tagMatcher)> */
		func() bool {
			position409, tokenIndex409, depth409 := position, tokenIndex, depth
			{
				position410 := position
				depth++
				{
					position411, tokenIndex411, depth411 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l412
					}
					{
						position413 := position
						depth++
						{
							position414, tokenIndex414, depth414 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l415
							}
							position++
							goto l414
						l415:
							position, tokenIndex, depth = position414, tokenIndex414, depth414
							if buffer[position] != rune('N') {
								goto l412
							}
							position++
						}
					l414:
						{
							position416, tokenIndex416, depth416 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l417
							}
							position++
							goto l416
						l417:
							position, tokenIndex, depth = position416, tokenIndex416, depth416
							if buffer[position] != rune('O') {
								goto l412
							}
							position++
						}
					l416:
						{
							position418, tokenIndex418, depth418 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l419
							}
							position++
							goto l418
						l419:
							position, tokenIndex, depth = position418, tokenIndex418, depth418
							if buffer[position] != rune('T') {
								goto l412
							}
							position++
						}
					l418:
						if !_rules[ruleKEY]() {
							goto l412
						}
						depth--
						add(ruleOP_NOT, position413)
					}
					{
						position420, tokenIndex420, depth420 := position, tokenIndex, depth
						if !_rules[rulepredicate_3]() {
							goto l421
						}
						goto l420
					l421:
						position, tokenIndex, depth = position420, tokenIndex420, depth420
						{
							add(ruleAction69, position)
						}
					}
				l420:
					{
						add(ruleAction70, position)
					}
					goto l411
				l412:
					position, tokenIndex, depth = position411, tokenIndex411, depth411
					if !_rules[rule_]() {
						goto l424
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l424
					}
					{
						position425, tokenIndex425, depth425 := position, tokenIndex, depth
						if !_rules[rulepredicate_1]() {
							goto l426
						}
						goto l425
					l426:
						position, tokenIndex, depth = position425, tokenIndex425, depth425
						{
							add(ruleAction71, position)
						}
					}
				l425:
					{
						position428, tokenIndex428, depth428 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l429
						}
						if !_rules[rulePAREN_CLOSE]() {
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
					goto l411
				l424:
					position, tokenIndex, depth = position411, tokenIndex411, depth411
					{
						position431 := position
						depth++
						if !_rules[ruletagName]() {
							goto l409
						}
						{
							position432, tokenIndex432, depth432 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l433
							}
							if buffer[position] != rune('=') {
								goto l433
							}
							position++
							{
								position434, tokenIndex434, depth434 := position, tokenIndex, depth
								if !_rules[ruleliteralString]() {
									goto l435
								}
								goto l434
							l435:
								position, tokenIndex, depth = position434, tokenIndex434, depth434
								{
									add(ruleAction73, position)
								}
							}
						l434:
							{
								add(ruleAction74, position)
							}
							goto l432
						l433:
							position, tokenIndex, depth = position432, tokenIndex432, depth432
							if !_rules[rule_]() {
								goto l438
							}
							if buffer[position] != rune('!') {
								goto l438
							}
							position++
							if buffer[position] != rune('=') {
								goto l438
							}
							position++
							{
								position439, tokenIndex439, depth439 := position, tokenIndex, depth
								if !_rules[ruleliteralString]() {
									goto l440
								}
								goto l439
							l440:
								position, tokenIndex, depth = position439, tokenIndex439, depth439
								{
									add(ruleAction75, position)
								}
							}
						l439:
							{
								add(ruleAction76, position)
							}
							{
								add(ruleAction77, position)
							}
							goto l432
						l438:
							position, tokenIndex, depth = position432, tokenIndex432, depth432
							if !_rules[rule_]() {
								goto l444
							}
							{
								position445, tokenIndex445, depth445 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l446
								}
								position++
								goto l445
							l446:
								position, tokenIndex, depth = position445, tokenIndex445, depth445
								if buffer[position] != rune('M') {
									goto l444
								}
								position++
							}
						l445:
							{
								position447, tokenIndex447, depth447 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l448
								}
								position++
								goto l447
							l448:
								position, tokenIndex, depth = position447, tokenIndex447, depth447
								if buffer[position] != rune('A') {
									goto l444
								}
								position++
							}
						l447:
							{
								position449, tokenIndex449, depth449 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l450
								}
								position++
								goto l449
							l450:
								position, tokenIndex, depth = position449, tokenIndex449, depth449
								if buffer[position] != rune('T') {
									goto l444
								}
								position++
							}
						l449:
							{
								position451, tokenIndex451, depth451 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l452
								}
								position++
								goto l451
							l452:
								position, tokenIndex, depth = position451, tokenIndex451, depth451
								if buffer[position] != rune('C') {
									goto l444
								}
								position++
							}
						l451:
							{
								position453, tokenIndex453, depth453 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l454
								}
								position++
								goto l453
							l454:
								position, tokenIndex, depth = position453, tokenIndex453, depth453
								if buffer[position] != rune('H') {
									goto l444
								}
								position++
							}
						l453:
							if !_rules[ruleKEY]() {
								goto l444
							}
							{
								position455, tokenIndex455, depth455 := position, tokenIndex, depth
								if !_rules[ruleliteralString]() {
									goto l456
								}
								goto l455
							l456:
								position, tokenIndex, depth = position455, tokenIndex455, depth455
								{
									add(ruleAction78, position)
								}
							}
						l455:
							{
								add(ruleAction79, position)
							}
							goto l432
						l444:
							position, tokenIndex, depth = position432, tokenIndex432, depth432
							if !_rules[rule_]() {
								goto l459
							}
							{
								position460, tokenIndex460, depth460 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l461
								}
								position++
								goto l460
							l461:
								position, tokenIndex, depth = position460, tokenIndex460, depth460
								if buffer[position] != rune('I') {
									goto l459
								}
								position++
							}
						l460:
							{
								position462, tokenIndex462, depth462 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l463
								}
								position++
								goto l462
							l463:
								position, tokenIndex, depth = position462, tokenIndex462, depth462
								if buffer[position] != rune('N') {
									goto l459
								}
								position++
							}
						l462:
							if !_rules[ruleKEY]() {
								goto l459
							}
							{
								position464, tokenIndex464, depth464 := position, tokenIndex, depth
								{
									position466 := position
									depth++
									{
										add(ruleAction84, position)
									}
									if !_rules[rule_]() {
										goto l465
									}
									if !_rules[rulePAREN_OPEN]() {
										goto l465
									}
									{
										position468, tokenIndex468, depth468 := position, tokenIndex, depth
										if !_rules[ruleliteralListString]() {
											goto l469
										}
										goto l468
									l469:
										position, tokenIndex, depth = position468, tokenIndex468, depth468
										{
											add(ruleAction85, position)
										}
									}
								l468:
								l471:
									{
										position472, tokenIndex472, depth472 := position, tokenIndex, depth
										if !_rules[rule_]() {
											goto l472
										}
										if !_rules[ruleCOMMA]() {
											goto l472
										}
										{
											position473, tokenIndex473, depth473 := position, tokenIndex, depth
											if !_rules[ruleliteralListString]() {
												goto l474
											}
											goto l473
										l474:
											position, tokenIndex, depth = position473, tokenIndex473, depth473
											{
												add(ruleAction86, position)
											}
										}
									l473:
										goto l471
									l472:
										position, tokenIndex, depth = position472, tokenIndex472, depth472
									}
									{
										position476, tokenIndex476, depth476 := position, tokenIndex, depth
										if !_rules[rule_]() {
											goto l477
										}
										if !_rules[rulePAREN_CLOSE]() {
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
									depth--
									add(ruleliteralList, position466)
								}
								goto l464
							l465:
								position, tokenIndex, depth = position464, tokenIndex464, depth464
								{
									add(ruleAction80, position)
								}
							}
						l464:
							{
								add(ruleAction81, position)
							}
							goto l432
						l459:
							position, tokenIndex, depth = position432, tokenIndex432, depth432
							{
								add(ruleAction82, position)
							}
						}
					l432:
						depth--
						add(ruletagMatcher, position431)
					}
				}
			l411:
				depth--
				add(rulepredicate_3, position410)
			}
			return true
		l409:
			position, tokenIndex, depth = position409, tokenIndex409, depth409
			return false
		},
		/* 29 tagMatcher <- <(tagName ((_ '=' (literalString / Action73) Action74) / (_ ('!' '=') (literalString / Action75) Action76 Action77) / (_ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H')) KEY (literalString / Action78) Action79) / (_ (('i' / 'I') ('n' / 'N')) KEY (literalList / Action80) Action81) / Action82))> */
		nil,
		/* 30 literalString <- <(_ STRING Action83)> */
		func() bool {
			position483, tokenIndex483, depth483 := position, tokenIndex, depth
			{
				position484 := position
				depth++
				if !_rules[rule_]() {
					goto l483
				}
				if !_rules[ruleSTRING]() {
					goto l483
				}
				{
					add(ruleAction83, position)
				}
				depth--
				add(ruleliteralString, position484)
			}
			return true
		l483:
			position, tokenIndex, depth = position483, tokenIndex483, depth483
			return false
		},
		/* 31 literalList <- <(Action84 _ PAREN_OPEN (literalListString / Action85) (_ COMMA (literalListString / Action86))* ((_ PAREN_CLOSE) / Action87))> */
		nil,
		/* 32 literalListString <- <(_ STRING Action88)> */
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
					add(ruleAction88, position)
				}
				depth--
				add(ruleliteralListString, position488)
			}
			return true
		l487:
			position, tokenIndex, depth = position487, tokenIndex487, depth487
			return false
		},
		/* 33 tagName <- <(_ <TAG_NAME> Action89)> */
		func() bool {
			position490, tokenIndex490, depth490 := position, tokenIndex, depth
			{
				position491 := position
				depth++
				if !_rules[rule_]() {
					goto l490
				}
				{
					position492 := position
					depth++
					{
						position493 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l490
						}
						depth--
						add(ruleTAG_NAME, position493)
					}
					depth--
					add(rulePegText, position492)
				}
				{
					add(ruleAction89, position)
				}
				depth--
				add(ruletagName, position491)
			}
			return true
		l490:
			position, tokenIndex, depth = position490, tokenIndex490, depth490
			return false
		},
		/* 34 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position495, tokenIndex495, depth495 := position, tokenIndex, depth
			{
				position496 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l495
				}
				depth--
				add(ruleCOLUMN_NAME, position496)
			}
			return true
		l495:
			position, tokenIndex, depth = position495, tokenIndex495, depth495
			return false
		},
		/* 35 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 36 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 37 IDENTIFIER <- <(('`' CHAR* ('`' / Action90)) / (_ !(KEYWORD KEY) ID_SEGMENT ('.' (ID_SEGMENT / Action91))*))> */
		func() bool {
			position499, tokenIndex499, depth499 := position, tokenIndex, depth
			{
				position500 := position
				depth++
				{
					position501, tokenIndex501, depth501 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l502
					}
					position++
				l503:
					{
						position504, tokenIndex504, depth504 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l504
						}
						goto l503
					l504:
						position, tokenIndex, depth = position504, tokenIndex504, depth504
					}
					{
						position505, tokenIndex505, depth505 := position, tokenIndex, depth
						if buffer[position] != rune('`') {
							goto l506
						}
						position++
						goto l505
					l506:
						position, tokenIndex, depth = position505, tokenIndex505, depth505
						{
							add(ruleAction90, position)
						}
					}
				l505:
					goto l501
				l502:
					position, tokenIndex, depth = position501, tokenIndex501, depth501
					if !_rules[rule_]() {
						goto l499
					}
					{
						position508, tokenIndex508, depth508 := position, tokenIndex, depth
						{
							position509 := position
							depth++
							{
								position510, tokenIndex510, depth510 := position, tokenIndex, depth
								{
									position512, tokenIndex512, depth512 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l513
									}
									position++
									goto l512
								l513:
									position, tokenIndex, depth = position512, tokenIndex512, depth512
									if buffer[position] != rune('A') {
										goto l511
									}
									position++
								}
							l512:
								{
									position514, tokenIndex514, depth514 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l515
									}
									position++
									goto l514
								l515:
									position, tokenIndex, depth = position514, tokenIndex514, depth514
									if buffer[position] != rune('L') {
										goto l511
									}
									position++
								}
							l514:
								{
									position516, tokenIndex516, depth516 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l517
									}
									position++
									goto l516
								l517:
									position, tokenIndex, depth = position516, tokenIndex516, depth516
									if buffer[position] != rune('L') {
										goto l511
									}
									position++
								}
							l516:
								goto l510
							l511:
								position, tokenIndex, depth = position510, tokenIndex510, depth510
								{
									position519, tokenIndex519, depth519 := position, tokenIndex, depth
									if buffer[position] != rune('a') {
										goto l520
									}
									position++
									goto l519
								l520:
									position, tokenIndex, depth = position519, tokenIndex519, depth519
									if buffer[position] != rune('A') {
										goto l518
									}
									position++
								}
							l519:
								{
									position521, tokenIndex521, depth521 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l522
									}
									position++
									goto l521
								l522:
									position, tokenIndex, depth = position521, tokenIndex521, depth521
									if buffer[position] != rune('N') {
										goto l518
									}
									position++
								}
							l521:
								{
									position523, tokenIndex523, depth523 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l524
									}
									position++
									goto l523
								l524:
									position, tokenIndex, depth = position523, tokenIndex523, depth523
									if buffer[position] != rune('D') {
										goto l518
									}
									position++
								}
							l523:
								goto l510
							l518:
								position, tokenIndex, depth = position510, tokenIndex510, depth510
								{
									position526, tokenIndex526, depth526 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l527
									}
									position++
									goto l526
								l527:
									position, tokenIndex, depth = position526, tokenIndex526, depth526
									if buffer[position] != rune('M') {
										goto l525
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
										goto l525
									}
									position++
								}
							l528:
								{
									position530, tokenIndex530, depth530 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l531
									}
									position++
									goto l530
								l531:
									position, tokenIndex, depth = position530, tokenIndex530, depth530
									if buffer[position] != rune('T') {
										goto l525
									}
									position++
								}
							l530:
								{
									position532, tokenIndex532, depth532 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l533
									}
									position++
									goto l532
								l533:
									position, tokenIndex, depth = position532, tokenIndex532, depth532
									if buffer[position] != rune('C') {
										goto l525
									}
									position++
								}
							l532:
								{
									position534, tokenIndex534, depth534 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l535
									}
									position++
									goto l534
								l535:
									position, tokenIndex, depth = position534, tokenIndex534, depth534
									if buffer[position] != rune('H') {
										goto l525
									}
									position++
								}
							l534:
								goto l510
							l525:
								position, tokenIndex, depth = position510, tokenIndex510, depth510
								{
									position537, tokenIndex537, depth537 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l538
									}
									position++
									goto l537
								l538:
									position, tokenIndex, depth = position537, tokenIndex537, depth537
									if buffer[position] != rune('S') {
										goto l536
									}
									position++
								}
							l537:
								{
									position539, tokenIndex539, depth539 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l540
									}
									position++
									goto l539
								l540:
									position, tokenIndex, depth = position539, tokenIndex539, depth539
									if buffer[position] != rune('E') {
										goto l536
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
										goto l536
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
										goto l536
									}
									position++
								}
							l543:
								{
									position545, tokenIndex545, depth545 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l546
									}
									position++
									goto l545
								l546:
									position, tokenIndex, depth = position545, tokenIndex545, depth545
									if buffer[position] != rune('C') {
										goto l536
									}
									position++
								}
							l545:
								{
									position547, tokenIndex547, depth547 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l548
									}
									position++
									goto l547
								l548:
									position, tokenIndex, depth = position547, tokenIndex547, depth547
									if buffer[position] != rune('T') {
										goto l536
									}
									position++
								}
							l547:
								goto l510
							l536:
								position, tokenIndex, depth = position510, tokenIndex510, depth510
								{
									switch buffer[position] {
									case 'M', 'm':
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
												goto l508
											}
											position++
										}
									l550:
										{
											position552, tokenIndex552, depth552 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l553
											}
											position++
											goto l552
										l553:
											position, tokenIndex, depth = position552, tokenIndex552, depth552
											if buffer[position] != rune('E') {
												goto l508
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
												goto l508
											}
											position++
										}
									l554:
										{
											position556, tokenIndex556, depth556 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l557
											}
											position++
											goto l556
										l557:
											position, tokenIndex, depth = position556, tokenIndex556, depth556
											if buffer[position] != rune('R') {
												goto l508
											}
											position++
										}
									l556:
										{
											position558, tokenIndex558, depth558 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l559
											}
											position++
											goto l558
										l559:
											position, tokenIndex, depth = position558, tokenIndex558, depth558
											if buffer[position] != rune('I') {
												goto l508
											}
											position++
										}
									l558:
										{
											position560, tokenIndex560, depth560 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l561
											}
											position++
											goto l560
										l561:
											position, tokenIndex, depth = position560, tokenIndex560, depth560
											if buffer[position] != rune('C') {
												goto l508
											}
											position++
										}
									l560:
										{
											position562, tokenIndex562, depth562 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l563
											}
											position++
											goto l562
										l563:
											position, tokenIndex, depth = position562, tokenIndex562, depth562
											if buffer[position] != rune('S') {
												goto l508
											}
											position++
										}
									l562:
										break
									case 'W', 'w':
										{
											position564, tokenIndex564, depth564 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l565
											}
											position++
											goto l564
										l565:
											position, tokenIndex, depth = position564, tokenIndex564, depth564
											if buffer[position] != rune('W') {
												goto l508
											}
											position++
										}
									l564:
										{
											position566, tokenIndex566, depth566 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l567
											}
											position++
											goto l566
										l567:
											position, tokenIndex, depth = position566, tokenIndex566, depth566
											if buffer[position] != rune('H') {
												goto l508
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
												goto l508
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
												goto l508
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
												goto l508
											}
											position++
										}
									l572:
										break
									case 'O', 'o':
										{
											position574, tokenIndex574, depth574 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l575
											}
											position++
											goto l574
										l575:
											position, tokenIndex, depth = position574, tokenIndex574, depth574
											if buffer[position] != rune('O') {
												goto l508
											}
											position++
										}
									l574:
										{
											position576, tokenIndex576, depth576 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l577
											}
											position++
											goto l576
										l577:
											position, tokenIndex, depth = position576, tokenIndex576, depth576
											if buffer[position] != rune('R') {
												goto l508
											}
											position++
										}
									l576:
										break
									case 'N', 'n':
										{
											position578, tokenIndex578, depth578 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l579
											}
											position++
											goto l578
										l579:
											position, tokenIndex, depth = position578, tokenIndex578, depth578
											if buffer[position] != rune('N') {
												goto l508
											}
											position++
										}
									l578:
										{
											position580, tokenIndex580, depth580 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l581
											}
											position++
											goto l580
										l581:
											position, tokenIndex, depth = position580, tokenIndex580, depth580
											if buffer[position] != rune('O') {
												goto l508
											}
											position++
										}
									l580:
										{
											position582, tokenIndex582, depth582 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l583
											}
											position++
											goto l582
										l583:
											position, tokenIndex, depth = position582, tokenIndex582, depth582
											if buffer[position] != rune('T') {
												goto l508
											}
											position++
										}
									l582:
										break
									case 'I', 'i':
										{
											position584, tokenIndex584, depth584 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l585
											}
											position++
											goto l584
										l585:
											position, tokenIndex, depth = position584, tokenIndex584, depth584
											if buffer[position] != rune('I') {
												goto l508
											}
											position++
										}
									l584:
										{
											position586, tokenIndex586, depth586 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l587
											}
											position++
											goto l586
										l587:
											position, tokenIndex, depth = position586, tokenIndex586, depth586
											if buffer[position] != rune('N') {
												goto l508
											}
											position++
										}
									l586:
										break
									case 'C', 'c':
										{
											position588, tokenIndex588, depth588 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l589
											}
											position++
											goto l588
										l589:
											position, tokenIndex, depth = position588, tokenIndex588, depth588
											if buffer[position] != rune('C') {
												goto l508
											}
											position++
										}
									l588:
										{
											position590, tokenIndex590, depth590 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l591
											}
											position++
											goto l590
										l591:
											position, tokenIndex, depth = position590, tokenIndex590, depth590
											if buffer[position] != rune('O') {
												goto l508
											}
											position++
										}
									l590:
										{
											position592, tokenIndex592, depth592 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l593
											}
											position++
											goto l592
										l593:
											position, tokenIndex, depth = position592, tokenIndex592, depth592
											if buffer[position] != rune('L') {
												goto l508
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
												goto l508
											}
											position++
										}
									l594:
										{
											position596, tokenIndex596, depth596 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l597
											}
											position++
											goto l596
										l597:
											position, tokenIndex, depth = position596, tokenIndex596, depth596
											if buffer[position] != rune('A') {
												goto l508
											}
											position++
										}
									l596:
										{
											position598, tokenIndex598, depth598 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l599
											}
											position++
											goto l598
										l599:
											position, tokenIndex, depth = position598, tokenIndex598, depth598
											if buffer[position] != rune('P') {
												goto l508
											}
											position++
										}
									l598:
										{
											position600, tokenIndex600, depth600 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l601
											}
											position++
											goto l600
										l601:
											position, tokenIndex, depth = position600, tokenIndex600, depth600
											if buffer[position] != rune('S') {
												goto l508
											}
											position++
										}
									l600:
										{
											position602, tokenIndex602, depth602 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l603
											}
											position++
											goto l602
										l603:
											position, tokenIndex, depth = position602, tokenIndex602, depth602
											if buffer[position] != rune('E') {
												goto l508
											}
											position++
										}
									l602:
										break
									case 'G', 'g':
										{
											position604, tokenIndex604, depth604 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l605
											}
											position++
											goto l604
										l605:
											position, tokenIndex, depth = position604, tokenIndex604, depth604
											if buffer[position] != rune('G') {
												goto l508
											}
											position++
										}
									l604:
										{
											position606, tokenIndex606, depth606 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l607
											}
											position++
											goto l606
										l607:
											position, tokenIndex, depth = position606, tokenIndex606, depth606
											if buffer[position] != rune('R') {
												goto l508
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
												goto l508
											}
											position++
										}
									l608:
										{
											position610, tokenIndex610, depth610 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l611
											}
											position++
											goto l610
										l611:
											position, tokenIndex, depth = position610, tokenIndex610, depth610
											if buffer[position] != rune('U') {
												goto l508
											}
											position++
										}
									l610:
										{
											position612, tokenIndex612, depth612 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l613
											}
											position++
											goto l612
										l613:
											position, tokenIndex, depth = position612, tokenIndex612, depth612
											if buffer[position] != rune('P') {
												goto l508
											}
											position++
										}
									l612:
										break
									case 'D', 'd':
										{
											position614, tokenIndex614, depth614 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l615
											}
											position++
											goto l614
										l615:
											position, tokenIndex, depth = position614, tokenIndex614, depth614
											if buffer[position] != rune('D') {
												goto l508
											}
											position++
										}
									l614:
										{
											position616, tokenIndex616, depth616 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l617
											}
											position++
											goto l616
										l617:
											position, tokenIndex, depth = position616, tokenIndex616, depth616
											if buffer[position] != rune('E') {
												goto l508
											}
											position++
										}
									l616:
										{
											position618, tokenIndex618, depth618 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l619
											}
											position++
											goto l618
										l619:
											position, tokenIndex, depth = position618, tokenIndex618, depth618
											if buffer[position] != rune('S') {
												goto l508
											}
											position++
										}
									l618:
										{
											position620, tokenIndex620, depth620 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l621
											}
											position++
											goto l620
										l621:
											position, tokenIndex, depth = position620, tokenIndex620, depth620
											if buffer[position] != rune('C') {
												goto l508
											}
											position++
										}
									l620:
										{
											position622, tokenIndex622, depth622 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l623
											}
											position++
											goto l622
										l623:
											position, tokenIndex, depth = position622, tokenIndex622, depth622
											if buffer[position] != rune('R') {
												goto l508
											}
											position++
										}
									l622:
										{
											position624, tokenIndex624, depth624 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l625
											}
											position++
											goto l624
										l625:
											position, tokenIndex, depth = position624, tokenIndex624, depth624
											if buffer[position] != rune('I') {
												goto l508
											}
											position++
										}
									l624:
										{
											position626, tokenIndex626, depth626 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l627
											}
											position++
											goto l626
										l627:
											position, tokenIndex, depth = position626, tokenIndex626, depth626
											if buffer[position] != rune('B') {
												goto l508
											}
											position++
										}
									l626:
										{
											position628, tokenIndex628, depth628 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l629
											}
											position++
											goto l628
										l629:
											position, tokenIndex, depth = position628, tokenIndex628, depth628
											if buffer[position] != rune('E') {
												goto l508
											}
											position++
										}
									l628:
										break
									case 'B', 'b':
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
												goto l508
											}
											position++
										}
									l630:
										{
											position632, tokenIndex632, depth632 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l633
											}
											position++
											goto l632
										l633:
											position, tokenIndex, depth = position632, tokenIndex632, depth632
											if buffer[position] != rune('Y') {
												goto l508
											}
											position++
										}
									l632:
										break
									case 'A', 'a':
										{
											position634, tokenIndex634, depth634 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l635
											}
											position++
											goto l634
										l635:
											position, tokenIndex, depth = position634, tokenIndex634, depth634
											if buffer[position] != rune('A') {
												goto l508
											}
											position++
										}
									l634:
										{
											position636, tokenIndex636, depth636 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l637
											}
											position++
											goto l636
										l637:
											position, tokenIndex, depth = position636, tokenIndex636, depth636
											if buffer[position] != rune('S') {
												goto l508
											}
											position++
										}
									l636:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l508
										}
										break
									}
								}

							}
						l510:
							depth--
							add(ruleKEYWORD, position509)
						}
						if !_rules[ruleKEY]() {
							goto l508
						}
						goto l499
					l508:
						position, tokenIndex, depth = position508, tokenIndex508, depth508
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l499
					}
				l638:
					{
						position639, tokenIndex639, depth639 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l639
						}
						position++
						{
							position640, tokenIndex640, depth640 := position, tokenIndex, depth
							if !_rules[ruleID_SEGMENT]() {
								goto l641
							}
							goto l640
						l641:
							position, tokenIndex, depth = position640, tokenIndex640, depth640
							{
								add(ruleAction91, position)
							}
						}
					l640:
						goto l638
					l639:
						position, tokenIndex, depth = position639, tokenIndex639, depth639
					}
				}
			l501:
				depth--
				add(ruleIDENTIFIER, position500)
			}
			return true
		l499:
			position, tokenIndex, depth = position499, tokenIndex499, depth499
			return false
		},
		/* 38 TIMESTAMP <- <((_ <(NUMBER ([a-z] / [A-Z])*)>) / (_ STRING) / (_ <(('n' / 'N') ('o' / 'O') ('w' / 'W'))> KEY))> */
		nil,
		/* 39 ID_SEGMENT <- <(_ ID_START ID_CONT*)> */
		func() bool {
			position644, tokenIndex644, depth644 := position, tokenIndex, depth
			{
				position645 := position
				depth++
				if !_rules[rule_]() {
					goto l644
				}
				if !_rules[ruleID_START]() {
					goto l644
				}
			l646:
				{
					position647, tokenIndex647, depth647 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l647
					}
					goto l646
				l647:
					position, tokenIndex, depth = position647, tokenIndex647, depth647
				}
				depth--
				add(ruleID_SEGMENT, position645)
			}
			return true
		l644:
			position, tokenIndex, depth = position644, tokenIndex644, depth644
			return false
		},
		/* 40 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position648, tokenIndex648, depth648 := position, tokenIndex, depth
			{
				position649 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l648
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l648
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l648
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position649)
			}
			return true
		l648:
			position, tokenIndex, depth = position648, tokenIndex648, depth648
			return false
		},
		/* 41 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position651, tokenIndex651, depth651 := position, tokenIndex, depth
			{
				position652 := position
				depth++
				{
					position653, tokenIndex653, depth653 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l654
					}
					goto l653
				l654:
					position, tokenIndex, depth = position653, tokenIndex653, depth653
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l651
					}
					position++
				}
			l653:
				depth--
				add(ruleID_CONT, position652)
			}
			return true
		l651:
			position, tokenIndex, depth = position651, tokenIndex651, depth651
			return false
		},
		/* 42 PROPERTY_KEY <- <((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> KEY ((_ (('b' / 'B') ('y' / 'Y')) KEY) / Action92))) | (&('R' | 'r') (<(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))> KEY)) | (&('T' | 't') (<(('t' / 'T') ('o' / 'O'))> KEY)) | (&('F' | 'f') (<(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))> KEY)))> */
		func() bool {
			position655, tokenIndex655, depth655 := position, tokenIndex, depth
			{
				position656 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position658 := position
							depth++
							{
								position659, tokenIndex659, depth659 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l660
								}
								position++
								goto l659
							l660:
								position, tokenIndex, depth = position659, tokenIndex659, depth659
								if buffer[position] != rune('S') {
									goto l655
								}
								position++
							}
						l659:
							{
								position661, tokenIndex661, depth661 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l662
								}
								position++
								goto l661
							l662:
								position, tokenIndex, depth = position661, tokenIndex661, depth661
								if buffer[position] != rune('A') {
									goto l655
								}
								position++
							}
						l661:
							{
								position663, tokenIndex663, depth663 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l664
								}
								position++
								goto l663
							l664:
								position, tokenIndex, depth = position663, tokenIndex663, depth663
								if buffer[position] != rune('M') {
									goto l655
								}
								position++
							}
						l663:
							{
								position665, tokenIndex665, depth665 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l666
								}
								position++
								goto l665
							l666:
								position, tokenIndex, depth = position665, tokenIndex665, depth665
								if buffer[position] != rune('P') {
									goto l655
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
									goto l655
								}
								position++
							}
						l667:
							{
								position669, tokenIndex669, depth669 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l670
								}
								position++
								goto l669
							l670:
								position, tokenIndex, depth = position669, tokenIndex669, depth669
								if buffer[position] != rune('E') {
									goto l655
								}
								position++
							}
						l669:
							depth--
							add(rulePegText, position658)
						}
						if !_rules[ruleKEY]() {
							goto l655
						}
						{
							position671, tokenIndex671, depth671 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l672
							}
							{
								position673, tokenIndex673, depth673 := position, tokenIndex, depth
								if buffer[position] != rune('b') {
									goto l674
								}
								position++
								goto l673
							l674:
								position, tokenIndex, depth = position673, tokenIndex673, depth673
								if buffer[position] != rune('B') {
									goto l672
								}
								position++
							}
						l673:
							{
								position675, tokenIndex675, depth675 := position, tokenIndex, depth
								if buffer[position] != rune('y') {
									goto l676
								}
								position++
								goto l675
							l676:
								position, tokenIndex, depth = position675, tokenIndex675, depth675
								if buffer[position] != rune('Y') {
									goto l672
								}
								position++
							}
						l675:
							if !_rules[ruleKEY]() {
								goto l672
							}
							goto l671
						l672:
							position, tokenIndex, depth = position671, tokenIndex671, depth671
							{
								add(ruleAction92, position)
							}
						}
					l671:
						break
					case 'R', 'r':
						{
							position678 := position
							depth++
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
									goto l655
								}
								position++
							}
						l679:
							{
								position681, tokenIndex681, depth681 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l682
								}
								position++
								goto l681
							l682:
								position, tokenIndex, depth = position681, tokenIndex681, depth681
								if buffer[position] != rune('E') {
									goto l655
								}
								position++
							}
						l681:
							{
								position683, tokenIndex683, depth683 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l684
								}
								position++
								goto l683
							l684:
								position, tokenIndex, depth = position683, tokenIndex683, depth683
								if buffer[position] != rune('S') {
									goto l655
								}
								position++
							}
						l683:
							{
								position685, tokenIndex685, depth685 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l686
								}
								position++
								goto l685
							l686:
								position, tokenIndex, depth = position685, tokenIndex685, depth685
								if buffer[position] != rune('O') {
									goto l655
								}
								position++
							}
						l685:
							{
								position687, tokenIndex687, depth687 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l688
								}
								position++
								goto l687
							l688:
								position, tokenIndex, depth = position687, tokenIndex687, depth687
								if buffer[position] != rune('L') {
									goto l655
								}
								position++
							}
						l687:
							{
								position689, tokenIndex689, depth689 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l690
								}
								position++
								goto l689
							l690:
								position, tokenIndex, depth = position689, tokenIndex689, depth689
								if buffer[position] != rune('U') {
									goto l655
								}
								position++
							}
						l689:
							{
								position691, tokenIndex691, depth691 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l692
								}
								position++
								goto l691
							l692:
								position, tokenIndex, depth = position691, tokenIndex691, depth691
								if buffer[position] != rune('T') {
									goto l655
								}
								position++
							}
						l691:
							{
								position693, tokenIndex693, depth693 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l694
								}
								position++
								goto l693
							l694:
								position, tokenIndex, depth = position693, tokenIndex693, depth693
								if buffer[position] != rune('I') {
									goto l655
								}
								position++
							}
						l693:
							{
								position695, tokenIndex695, depth695 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l696
								}
								position++
								goto l695
							l696:
								position, tokenIndex, depth = position695, tokenIndex695, depth695
								if buffer[position] != rune('O') {
									goto l655
								}
								position++
							}
						l695:
							{
								position697, tokenIndex697, depth697 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l698
								}
								position++
								goto l697
							l698:
								position, tokenIndex, depth = position697, tokenIndex697, depth697
								if buffer[position] != rune('N') {
									goto l655
								}
								position++
							}
						l697:
							depth--
							add(rulePegText, position678)
						}
						if !_rules[ruleKEY]() {
							goto l655
						}
						break
					case 'T', 't':
						{
							position699 := position
							depth++
							{
								position700, tokenIndex700, depth700 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l701
								}
								position++
								goto l700
							l701:
								position, tokenIndex, depth = position700, tokenIndex700, depth700
								if buffer[position] != rune('T') {
									goto l655
								}
								position++
							}
						l700:
							{
								position702, tokenIndex702, depth702 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l703
								}
								position++
								goto l702
							l703:
								position, tokenIndex, depth = position702, tokenIndex702, depth702
								if buffer[position] != rune('O') {
									goto l655
								}
								position++
							}
						l702:
							depth--
							add(rulePegText, position699)
						}
						if !_rules[ruleKEY]() {
							goto l655
						}
						break
					default:
						{
							position704 := position
							depth++
							{
								position705, tokenIndex705, depth705 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l706
								}
								position++
								goto l705
							l706:
								position, tokenIndex, depth = position705, tokenIndex705, depth705
								if buffer[position] != rune('F') {
									goto l655
								}
								position++
							}
						l705:
							{
								position707, tokenIndex707, depth707 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l708
								}
								position++
								goto l707
							l708:
								position, tokenIndex, depth = position707, tokenIndex707, depth707
								if buffer[position] != rune('R') {
									goto l655
								}
								position++
							}
						l707:
							{
								position709, tokenIndex709, depth709 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l710
								}
								position++
								goto l709
							l710:
								position, tokenIndex, depth = position709, tokenIndex709, depth709
								if buffer[position] != rune('O') {
									goto l655
								}
								position++
							}
						l709:
							{
								position711, tokenIndex711, depth711 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l712
								}
								position++
								goto l711
							l712:
								position, tokenIndex, depth = position711, tokenIndex711, depth711
								if buffer[position] != rune('M') {
									goto l655
								}
								position++
							}
						l711:
							depth--
							add(rulePegText, position704)
						}
						if !_rules[ruleKEY]() {
							goto l655
						}
						break
					}
				}

				depth--
				add(rulePROPERTY_KEY, position656)
			}
			return true
		l655:
			position, tokenIndex, depth = position655, tokenIndex655, depth655
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
			position723, tokenIndex723, depth723 := position, tokenIndex, depth
			{
				position724 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l723
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position724)
			}
			return true
		l723:
			position, tokenIndex, depth = position723, tokenIndex723, depth723
			return false
		},
		/* 54 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position725, tokenIndex725, depth725 := position, tokenIndex, depth
			{
				position726 := position
				depth++
				if buffer[position] != rune('"') {
					goto l725
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position726)
			}
			return true
		l725:
			position, tokenIndex, depth = position725, tokenIndex725, depth725
			return false
		},
		/* 55 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> (QUOTE_SINGLE / Action93)) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> (QUOTE_DOUBLE / Action94)))> */
		func() bool {
			position727, tokenIndex727, depth727 := position, tokenIndex, depth
			{
				position728 := position
				depth++
				{
					position729, tokenIndex729, depth729 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l730
					}
					{
						position731 := position
						depth++
					l732:
						{
							position733, tokenIndex733, depth733 := position, tokenIndex, depth
							{
								position734, tokenIndex734, depth734 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l734
								}
								goto l733
							l734:
								position, tokenIndex, depth = position734, tokenIndex734, depth734
							}
							if !_rules[ruleCHAR]() {
								goto l733
							}
							goto l732
						l733:
							position, tokenIndex, depth = position733, tokenIndex733, depth733
						}
						depth--
						add(rulePegText, position731)
					}
					{
						position735, tokenIndex735, depth735 := position, tokenIndex, depth
						if !_rules[ruleQUOTE_SINGLE]() {
							goto l736
						}
						goto l735
					l736:
						position, tokenIndex, depth = position735, tokenIndex735, depth735
						{
							add(ruleAction93, position)
						}
					}
				l735:
					goto l729
				l730:
					position, tokenIndex, depth = position729, tokenIndex729, depth729
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l727
					}
					{
						position738 := position
						depth++
					l739:
						{
							position740, tokenIndex740, depth740 := position, tokenIndex, depth
							{
								position741, tokenIndex741, depth741 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l741
								}
								goto l740
							l741:
								position, tokenIndex, depth = position741, tokenIndex741, depth741
							}
							if !_rules[ruleCHAR]() {
								goto l740
							}
							goto l739
						l740:
							position, tokenIndex, depth = position740, tokenIndex740, depth740
						}
						depth--
						add(rulePegText, position738)
					}
					{
						position742, tokenIndex742, depth742 := position, tokenIndex, depth
						if !_rules[ruleQUOTE_DOUBLE]() {
							goto l743
						}
						goto l742
					l743:
						position, tokenIndex, depth = position742, tokenIndex742, depth742
						{
							add(ruleAction94, position)
						}
					}
				l742:
				}
			l729:
				depth--
				add(ruleSTRING, position728)
			}
			return true
		l727:
			position, tokenIndex, depth = position727, tokenIndex727, depth727
			return false
		},
		/* 56 CHAR <- <(('\\' ((&('"') QUOTE_DOUBLE) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position745, tokenIndex745, depth745 := position, tokenIndex, depth
			{
				position746 := position
				depth++
				{
					position747, tokenIndex747, depth747 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l748
					}
					position++
					{
						switch buffer[position] {
						case '"':
							if !_rules[ruleQUOTE_DOUBLE]() {
								goto l748
							}
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l748
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l748
							}
							break
						}
					}

					goto l747
				l748:
					position, tokenIndex, depth = position747, tokenIndex747, depth747
					{
						position750, tokenIndex750, depth750 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l750
						}
						goto l745
					l750:
						position, tokenIndex, depth = position750, tokenIndex750, depth750
					}
					if !matchDot() {
						goto l745
					}
				}
			l747:
				depth--
				add(ruleCHAR, position746)
			}
			return true
		l745:
			position, tokenIndex, depth = position745, tokenIndex745, depth745
			return false
		},
		/* 57 ESCAPE_CLASS <- <('`' / ('\\' / Action95))> */
		func() bool {
			{
				position752 := position
				depth++
				{
					position753, tokenIndex753, depth753 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l754
					}
					position++
					goto l753
				l754:
					position, tokenIndex, depth = position753, tokenIndex753, depth753
					{
						position755, tokenIndex755, depth755 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l756
						}
						position++
						goto l755
					l756:
						position, tokenIndex, depth = position755, tokenIndex755, depth755
						{
							add(ruleAction95, position)
						}
					}
				l755:
				}
			l753:
				depth--
				add(ruleESCAPE_CLASS, position752)
			}
			return true
		},
		/* 58 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position758, tokenIndex758, depth758 := position, tokenIndex, depth
			{
				position759 := position
				depth++
				{
					position760 := position
					depth++
					{
						position761, tokenIndex761, depth761 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l761
						}
						position++
						goto l762
					l761:
						position, tokenIndex, depth = position761, tokenIndex761, depth761
					}
				l762:
					{
						position763 := position
						depth++
						{
							position764, tokenIndex764, depth764 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l765
							}
							position++
							goto l764
						l765:
							position, tokenIndex, depth = position764, tokenIndex764, depth764
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l758
							}
							position++
						l766:
							{
								position767, tokenIndex767, depth767 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l767
								}
								position++
								goto l766
							l767:
								position, tokenIndex, depth = position767, tokenIndex767, depth767
							}
						}
					l764:
						depth--
						add(ruleNUMBER_NATURAL, position763)
					}
					depth--
					add(ruleNUMBER_INTEGER, position760)
				}
				{
					position768, tokenIndex768, depth768 := position, tokenIndex, depth
					{
						position770 := position
						depth++
						if buffer[position] != rune('.') {
							goto l768
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l768
						}
						position++
					l771:
						{
							position772, tokenIndex772, depth772 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l772
							}
							position++
							goto l771
						l772:
							position, tokenIndex, depth = position772, tokenIndex772, depth772
						}
						depth--
						add(ruleNUMBER_FRACTION, position770)
					}
					goto l769
				l768:
					position, tokenIndex, depth = position768, tokenIndex768, depth768
				}
			l769:
				{
					position773, tokenIndex773, depth773 := position, tokenIndex, depth
					{
						position775 := position
						depth++
						{
							position776, tokenIndex776, depth776 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l777
							}
							position++
							goto l776
						l777:
							position, tokenIndex, depth = position776, tokenIndex776, depth776
							if buffer[position] != rune('E') {
								goto l773
							}
							position++
						}
					l776:
						{
							position778, tokenIndex778, depth778 := position, tokenIndex, depth
							{
								position780, tokenIndex780, depth780 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l781
								}
								position++
								goto l780
							l781:
								position, tokenIndex, depth = position780, tokenIndex780, depth780
								if buffer[position] != rune('-') {
									goto l778
								}
								position++
							}
						l780:
							goto l779
						l778:
							position, tokenIndex, depth = position778, tokenIndex778, depth778
						}
					l779:
						{
							position782, tokenIndex782, depth782 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l783
							}
							position++
						l784:
							{
								position785, tokenIndex785, depth785 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l785
								}
								position++
								goto l784
							l785:
								position, tokenIndex, depth = position785, tokenIndex785, depth785
							}
							goto l782
						l783:
							position, tokenIndex, depth = position782, tokenIndex782, depth782
							{
								add(ruleAction96, position)
							}
						}
					l782:
						depth--
						add(ruleNUMBER_EXP, position775)
					}
					goto l774
				l773:
					position, tokenIndex, depth = position773, tokenIndex773, depth773
				}
			l774:
				depth--
				add(ruleNUMBER, position759)
			}
			return true
		l758:
			position, tokenIndex, depth = position758, tokenIndex758, depth758
			return false
		},
		/* 59 NUMBER_NATURAL <- <('0' / ([1-9] [0-9]*))> */
		nil,
		/* 60 NUMBER_FRACTION <- <('.' [0-9]+)> */
		nil,
		/* 61 NUMBER_INTEGER <- <('-'? NUMBER_NATURAL)> */
		nil,
		/* 62 NUMBER_EXP <- <(('e' / 'E') ('+' / '-')? ([0-9]+ / Action96))> */
		nil,
		/* 63 DURATION <- <(NUMBER [a-z]+ KEY)> */
		nil,
		/* 64 PAREN_OPEN <- <'('> */
		func() bool {
			position792, tokenIndex792, depth792 := position, tokenIndex, depth
			{
				position793 := position
				depth++
				if buffer[position] != rune('(') {
					goto l792
				}
				position++
				depth--
				add(rulePAREN_OPEN, position793)
			}
			return true
		l792:
			position, tokenIndex, depth = position792, tokenIndex792, depth792
			return false
		},
		/* 65 PAREN_CLOSE <- <')'> */
		func() bool {
			position794, tokenIndex794, depth794 := position, tokenIndex, depth
			{
				position795 := position
				depth++
				if buffer[position] != rune(')') {
					goto l794
				}
				position++
				depth--
				add(rulePAREN_CLOSE, position795)
			}
			return true
		l794:
			position, tokenIndex, depth = position794, tokenIndex794, depth794
			return false
		},
		/* 66 COMMA <- <','> */
		func() bool {
			position796, tokenIndex796, depth796 := position, tokenIndex, depth
			{
				position797 := position
				depth++
				if buffer[position] != rune(',') {
					goto l796
				}
				position++
				depth--
				add(ruleCOMMA, position797)
			}
			return true
		l796:
			position, tokenIndex, depth = position796, tokenIndex796, depth796
			return false
		},
		/* 67 _ <- <((&('/') COMMENT_BLOCK) | (&('-') COMMENT_TRAIL) | (&('\t' | '\n' | ' ') SPACE))*> */
		func() bool {
			{
				position799 := position
				depth++
			l800:
				{
					position801, tokenIndex801, depth801 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '/':
							{
								position803 := position
								depth++
								if buffer[position] != rune('/') {
									goto l801
								}
								position++
								if buffer[position] != rune('*') {
									goto l801
								}
								position++
							l804:
								{
									position805, tokenIndex805, depth805 := position, tokenIndex, depth
									{
										position806, tokenIndex806, depth806 := position, tokenIndex, depth
										if buffer[position] != rune('*') {
											goto l806
										}
										position++
										if buffer[position] != rune('/') {
											goto l806
										}
										position++
										goto l805
									l806:
										position, tokenIndex, depth = position806, tokenIndex806, depth806
									}
									if !matchDot() {
										goto l805
									}
									goto l804
								l805:
									position, tokenIndex, depth = position805, tokenIndex805, depth805
								}
								if buffer[position] != rune('*') {
									goto l801
								}
								position++
								if buffer[position] != rune('/') {
									goto l801
								}
								position++
								depth--
								add(ruleCOMMENT_BLOCK, position803)
							}
							break
						case '-':
							{
								position807 := position
								depth++
								if buffer[position] != rune('-') {
									goto l801
								}
								position++
								if buffer[position] != rune('-') {
									goto l801
								}
								position++
							l808:
								{
									position809, tokenIndex809, depth809 := position, tokenIndex, depth
									{
										position810, tokenIndex810, depth810 := position, tokenIndex, depth
										if buffer[position] != rune('\n') {
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
								depth--
								add(ruleCOMMENT_TRAIL, position807)
							}
							break
						default:
							{
								position811 := position
								depth++
								{
									switch buffer[position] {
									case '\t':
										if buffer[position] != rune('\t') {
											goto l801
										}
										position++
										break
									case '\n':
										if buffer[position] != rune('\n') {
											goto l801
										}
										position++
										break
									default:
										if buffer[position] != rune(' ') {
											goto l801
										}
										position++
										break
									}
								}

								depth--
								add(ruleSPACE, position811)
							}
							break
						}
					}

					goto l800
				l801:
					position, tokenIndex, depth = position801, tokenIndex801, depth801
				}
				depth--
				add(rule_, position799)
			}
			return true
		},
		/* 68 COMMENT_TRAIL <- <('-' '-' (!'\n' .)*)> */
		nil,
		/* 69 COMMENT_BLOCK <- <('/' '*' (!('*' '/') .)* ('*' '/'))> */
		nil,
		/* 70 KEY <- <!ID_CONT> */
		func() bool {
			position815, tokenIndex815, depth815 := position, tokenIndex, depth
			{
				position816 := position
				depth++
				{
					position817, tokenIndex817, depth817 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l817
					}
					goto l815
				l817:
					position, tokenIndex, depth = position817, tokenIndex817, depth817
				}
				depth--
				add(ruleKEY, position816)
			}
			return true
		l815:
			position, tokenIndex, depth = position815, tokenIndex815, depth815
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
		/* 76 Action3 <- <{ p.errorHere(position, `expected string literal to follow keyword "match"`) }> */
		nil,
		/* 77 Action4 <- <{ p.addMatchClause() }> */
		nil,
		/* 78 Action5 <- <{ p.errorHere(position, `expected "where" to follow keyword "metrics" in "describe metrics" command`) }> */
		nil,
		/* 79 Action6 <- <{ p.errorHere(position, `expected "=" to follow keyword "where" in "describe metrics" command`) }> */
		nil,
		/* 80 Action7 <- <{ p.errorHere(position, `expected string literal to follow "=" in "describe metrics" command`) }> */
		nil,
		/* 81 Action8 <- <{ p.makeDescribeMetrics() }> */
		nil,
		nil,
		/* 83 Action9 <- <{ p.pushString(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 84 Action10 <- <{ p.errorHere(position, `expected metric name to follow "describe" in "describe" command`) }> */
		nil,
		/* 85 Action11 <- <{ p.makeDescribe() }> */
		nil,
		/* 86 Action12 <- <{ p.addEvaluationContext() }> */
		nil,
		/* 87 Action13 <- <{ p.addPropertyKey(buffer[begin:end])   }> */
		nil,
		/* 88 Action14 <- <{ p.addPropertyValue(buffer[begin:end]) }> */
		nil,
		/* 89 Action15 <- <{ p.errorHere(position, `expected property value to follow property key`) }> */
		nil,
		/* 90 Action16 <- <{ p.insertPropertyKeyValue() }> */
		nil,
		/* 91 Action17 <- <{ p.checkPropertyClause() }> */
		nil,
		/* 92 Action18 <- <{ p.addNullPredicate() }> */
		nil,
		/* 93 Action19 <- <{ p.addExpressionList() }> */
		nil,
		/* 94 Action20 <- <{ p.appendExpression() }> */
		nil,
		/* 95 Action21 <- <{ p.errorHere(position, `expected expression to follow ","`) }> */
		nil,
		/* 96 Action22 <- <{ p.appendExpression() }> */
		nil,
		/* 97 Action23 <- <{ p.addOperatorLiteral("+") }> */
		nil,
		/* 98 Action24 <- <{ p.addOperatorLiteral("-") }> */
		nil,
		/* 99 Action25 <- <{ p.errorHere(position, `expected expression to follow operator "+" or "-"`) }> */
		nil,
		/* 100 Action26 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 101 Action27 <- <{ p.addOperatorLiteral("/") }> */
		nil,
		/* 102 Action28 <- <{ p.addOperatorLiteral("*") }> */
		nil,
		/* 103 Action29 <- <{ p.errorHere(position, `expected expression to follow operator "*" or "/"`) }> */
		nil,
		/* 104 Action30 <- <{ p.addOperatorFunction() }> */
		nil,
		/* 105 Action31 <- <{ p.errorHere(position, `expected function name to follow pipe "|"`) }> */
		nil,
		/* 106 Action32 <- <{ p.pushString(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 107 Action33 <- <{p.addExpressionList()}> */
		nil,
		/* 108 Action34 <- <{ p.errorHere(position, `expected ")" to close "(" opened in pipe function call`) }> */
		nil,
		/* 109 Action35 <- <{
		   p.addExpressionList()
		   p.addGroupBy()
		 }> */
		nil,
		/* 110 Action36 <- <{ p.addPipeExpression() }> */
		nil,
		/* 111 Action37 <- <{ p.errorHere(position, `expected expression to follow "("`) }> */
		nil,
		/* 112 Action38 <- <{ p.errorHere(position, `expected ")" to close "("`) }> */
		nil,
		/* 113 Action39 <- <{ p.addDurationNode(text) }> */
		nil,
		/* 114 Action40 <- <{ p.addNumberNode(buffer[begin:end]) }> */
		nil,
		/* 115 Action41 <- <{ p.addStringNode(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 116 Action42 <- <{ p.errorHere(position, `expected "}> */
		nil,
		/* 117 Action43 <- <{" opened for annotation`) }> */
		nil,
		/* 118 Action44 <- <{ p.addAnnotationExpression(buffer[begin:end]) }> */
		nil,
		/* 119 Action45 <- <{ p.addGroupBy() }> */
		nil,
		/* 120 Action46 <- <{ p.pushString(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 121 Action47 <- <{ p.errorHere(position, `expected ")" to close "(" opened by function call`) }> */
		nil,
		/* 122 Action48 <- <{ p.addFunctionInvocation() }> */
		nil,
		/* 123 Action49 <- <{ p.pushString(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 124 Action50 <- <{ p.errorHere(position, `expected predicate to follow "[" after metric`) }> */
		nil,
		/* 125 Action51 <- <{ p.errorHere(position, `expected "]" to close "[" opened to apply predicate`) }> */
		nil,
		/* 126 Action52 <- <{ p.addNullPredicate() }> */
		nil,
		/* 127 Action53 <- <{ p.addMetricExpression() }> */
		nil,
		/* 128 Action54 <- <{ p.errorHere(position, `expected keyword "by" to follow keyword "group" in "group by" clause`) }> */
		nil,
		/* 129 Action55 <- <{ p.errorHere(position, `expected tag key identifier to follow "group by" keywords in "group by" clause`) }> */
		nil,
		/* 130 Action56 <- <{ p.appendGroupBy(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 131 Action57 <- <{ p.errorHere(position, `expected tag key identifier to follow "," in "group by" clause`) }> */
		nil,
		/* 132 Action58 <- <{ p.appendGroupBy(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 133 Action59 <- <{ p.errorHere(position, `expected keyword "by" to follow keyword "collapse" in "collapse by" clause`) }> */
		nil,
		/* 134 Action60 <- <{ p.errorHere(position, `expected tag key identifier to follow "collapse by" keywords in "collapse by" clause`) }> */
		nil,
		/* 135 Action61 <- <{ p.appendCollapseBy(unescapeLiteral(text)) }> */
		nil,
		/* 136 Action62 <- <{ p.errorHere(position, `expected tag key identifier to follow "," in "collapse by" clause`) }> */
		nil,
		/* 137 Action63 <- <{ p.appendCollapseBy(unescapeLiteral(text)) }> */
		nil,
		/* 138 Action64 <- <{ p.errorHere(position, `expected predicate to follow "where" keyword`) }> */
		nil,
		/* 139 Action65 <- <{ p.errorHere(position, `expected predicate to follow "or" operator`) }> */
		nil,
		/* 140 Action66 <- <{ p.addOrPredicate() }> */
		nil,
		/* 141 Action67 <- <{ p.errorHere(position, `expected predicate to follow "and" operator`) }> */
		nil,
		/* 142 Action68 <- <{ p.addAndPredicate() }> */
		nil,
		/* 143 Action69 <- <{ p.errorHere(position, `expected predicate to follow "not" operator`) }> */
		nil,
		/* 144 Action70 <- <{ p.addNotPredicate() }> */
		nil,
		/* 145 Action71 <- <{ p.errorHere(position, `expected predicate to follow "("`) }> */
		nil,
		/* 146 Action72 <- <{ p.errorHere(position, `expected ")" to close "(" opened in predicate`) }> */
		nil,
		/* 147 Action73 <- <{ p.errorHere(position, `expected string literal to follow "="`) }> */
		nil,
		/* 148 Action74 <- <{ p.addLiteralMatcher() }> */
		nil,
		/* 149 Action75 <- <{ p.errorHere(position, `expected string literal to follow "!="`) }> */
		nil,
		/* 150 Action76 <- <{ p.addLiteralMatcher() }> */
		nil,
		/* 151 Action77 <- <{ p.addNotPredicate() }> */
		nil,
		/* 152 Action78 <- <{ p.errorHere(position, `expected regex string literal to follow "match"`) }> */
		nil,
		/* 153 Action79 <- <{ p.addRegexMatcher() }> */
		nil,
		/* 154 Action80 <- <{ p.errorHere(position, `expected string literal list to follow "in" keyword`) }> */
		nil,
		/* 155 Action81 <- <{ p.addListMatcher() }> */
		nil,
		/* 156 Action82 <- <{ p.errorHere(position, `expected "=", "!=", "match", or "in" to follow tag key in predicate`) }> */
		nil,
		/* 157 Action83 <- <{ p.pushString(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 158 Action84 <- <{ p.addLiteralList() }> */
		nil,
		/* 159 Action85 <- <{ p.errorHere(position, `expected string literal to follow "(" in literal list`) }> */
		nil,
		/* 160 Action86 <- <{ p.errorHere(position, `expected string literal to follow "," in literal list`) }> */
		nil,
		/* 161 Action87 <- <{ p.errorHere(position, `expected ")" to close "(" for literal list`) }> */
		nil,
		/* 162 Action88 <- <{ p.appendLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 163 Action89 <- <{ p.addTagLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 164 Action90 <- <{ p.errorHere(position, "expected \"`\" to end identifier") }> */
		nil,
		/* 165 Action91 <- <{ p.errorHere(position, `expected identifier segment to follow "."`) }> */
		nil,
		/* 166 Action92 <- <{ p.errorHere(position, `expected keyword "by" to follow keyword "sample"`) }> */
		nil,
		/* 167 Action93 <- <{ p.errorHere(position, `expected "'" to close string`) }> */
		nil,
		/* 168 Action94 <- <{ p.errorHere(position, `expected '"' to close string`) }> */
		nil,
		/* 169 Action95 <- <{ p.errorHere(position, "expected \"\\\" or \"`\" to follow escaping backslash") }> */
		nil,
		/* 170 Action96 <- <{ p.errorHere(position, `expected exponent`) }> */
		nil,
	}
	p.rules = _rules
}
