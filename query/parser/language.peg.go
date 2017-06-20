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
}

type token32 struct {
	pegRule
	begin, end uint32
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v", rul3s[t.pegRule], t.begin, t.end)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(pretty bool, buffer string) {
	var print func(node *node32, depth int)
	print = func(node *node32, depth int) {
		for node != nil {
			for c := 0; c < depth; c++ {
				fmt.Printf(" ")
			}
			rule := rul3s[node.pegRule]
			quote := strconv.Quote(string(([]rune(buffer)[node.begin:node.end])))
			if !pretty {
				fmt.Printf("%v %v\n", rule, quote)
			} else {
				fmt.Printf("\x1B[34m%v\x1B[m %v\n", rule, quote)
			}
			if node.up != nil {
				print(node.up, depth+1)
			}
			node = node.next
		}
	}
	print(node, 0)
}

func (node *node32) Print(buffer string) {
	node.print(false, buffer)
}

func (node *node32) PrettyPrint(buffer string) {
	node.print(true, buffer)
}

type tokens32 struct {
	tree []token32
}

func (t *tokens32) Trim(length uint32) {
	t.tree = t.tree[:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) AST() *node32 {
	type element struct {
		node *node32
		down *element
	}
	tokens := t.Tokens()
	var stack *element
	for _, token := range tokens {
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
	if stack != nil {
		return stack.node
	}
	return nil
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	t.AST().Print(buffer)
}

func (t *tokens32) PrettyPrintSyntaxTree(buffer string) {
	t.AST().PrettyPrint(buffer)
}

func (t *tokens32) Add(rule pegRule, begin, end, index uint32) {
	if tree := t.tree; int(index) >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	t.tree[index] = token32{
		pegRule: rule,
		begin:   begin,
		end:     end,
	}
}

func (t *tokens32) Tokens() []token32 {
	return t.tree
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
	parse  func(rule ...int) error
	reset  func()
	Pretty bool
	tokens32
}

func (p *Parser) Parse(rule ...int) error {
	return p.parse(rule...)
}

func (p *Parser) Reset() {
	p.reset()
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
	if p.Pretty {
		p.tokens32.PrettyPrintSyntaxTree(p.Buffer)
	} else {
		p.tokens32.PrintSyntaxTree(p.Buffer)
	}
}

func (p *Parser) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for _, token := range p.Tokens() {
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
	var (
		max                  token32
		position, tokenIndex uint32
		buffer               []rune
	)
	p.reset = func() {
		max = token32{}
		position, tokenIndex = 0, 0

		p.buffer = []rune(p.Buffer)
		if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
			p.buffer = append(p.buffer, endSymbol)
		}
		buffer = p.buffer
	}
	p.reset()

	_rules := p.rules
	tree := tokens32{tree: make([]token32, math.MaxInt16)}
	p.parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.Trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	add := func(rule pegRule, begin uint32) {
		tree.Add(rule, begin, position, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position}
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
			position0, tokenIndex0 := position, tokenIndex
			{
				position1 := position
				{
					position2, tokenIndex2 := position, tokenIndex
					{
						position4 := position
						if !_rules[rule_]() {
							goto l3
						}
						{
							position5, tokenIndex5 := position, tokenIndex
							{
								position7, tokenIndex7 := position, tokenIndex
								if buffer[position] != rune('s') {
									goto l8
								}
								position++
								goto l7
							l8:
								position, tokenIndex = position7, tokenIndex7
								if buffer[position] != rune('S') {
									goto l5
								}
								position++
							}
						l7:
							{
								position9, tokenIndex9 := position, tokenIndex
								if buffer[position] != rune('e') {
									goto l10
								}
								position++
								goto l9
							l10:
								position, tokenIndex = position9, tokenIndex9
								if buffer[position] != rune('E') {
									goto l5
								}
								position++
							}
						l9:
							{
								position11, tokenIndex11 := position, tokenIndex
								if buffer[position] != rune('l') {
									goto l12
								}
								position++
								goto l11
							l12:
								position, tokenIndex = position11, tokenIndex11
								if buffer[position] != rune('L') {
									goto l5
								}
								position++
							}
						l11:
							{
								position13, tokenIndex13 := position, tokenIndex
								if buffer[position] != rune('e') {
									goto l14
								}
								position++
								goto l13
							l14:
								position, tokenIndex = position13, tokenIndex13
								if buffer[position] != rune('E') {
									goto l5
								}
								position++
							}
						l13:
							{
								position15, tokenIndex15 := position, tokenIndex
								if buffer[position] != rune('c') {
									goto l16
								}
								position++
								goto l15
							l16:
								position, tokenIndex = position15, tokenIndex15
								if buffer[position] != rune('C') {
									goto l5
								}
								position++
							}
						l15:
							{
								position17, tokenIndex17 := position, tokenIndex
								if buffer[position] != rune('t') {
									goto l18
								}
								position++
								goto l17
							l18:
								position, tokenIndex = position17, tokenIndex17
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
							position, tokenIndex = position5, tokenIndex5
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
							{
								add(ruleAction7, position)
							}
						l21:
							{
								position22, tokenIndex22 := position, tokenIndex
								{
									position23, tokenIndex23 := position, tokenIndex
									if !_rules[rule_]() {
										goto l24
									}
									{
										position25 := position
										{
											switch buffer[position] {
											case 'S', 's':
												{
													position27 := position
													{
														position28, tokenIndex28 := position, tokenIndex
														if buffer[position] != rune('s') {
															goto l29
														}
														position++
														goto l28
													l29:
														position, tokenIndex = position28, tokenIndex28
														if buffer[position] != rune('S') {
															goto l24
														}
														position++
													}
												l28:
													{
														position30, tokenIndex30 := position, tokenIndex
														if buffer[position] != rune('a') {
															goto l31
														}
														position++
														goto l30
													l31:
														position, tokenIndex = position30, tokenIndex30
														if buffer[position] != rune('A') {
															goto l24
														}
														position++
													}
												l30:
													{
														position32, tokenIndex32 := position, tokenIndex
														if buffer[position] != rune('m') {
															goto l33
														}
														position++
														goto l32
													l33:
														position, tokenIndex = position32, tokenIndex32
														if buffer[position] != rune('M') {
															goto l24
														}
														position++
													}
												l32:
													{
														position34, tokenIndex34 := position, tokenIndex
														if buffer[position] != rune('p') {
															goto l35
														}
														position++
														goto l34
													l35:
														position, tokenIndex = position34, tokenIndex34
														if buffer[position] != rune('P') {
															goto l24
														}
														position++
													}
												l34:
													{
														position36, tokenIndex36 := position, tokenIndex
														if buffer[position] != rune('l') {
															goto l37
														}
														position++
														goto l36
													l37:
														position, tokenIndex = position36, tokenIndex36
														if buffer[position] != rune('L') {
															goto l24
														}
														position++
													}
												l36:
													{
														position38, tokenIndex38 := position, tokenIndex
														if buffer[position] != rune('e') {
															goto l39
														}
														position++
														goto l38
													l39:
														position, tokenIndex = position38, tokenIndex38
														if buffer[position] != rune('E') {
															goto l24
														}
														position++
													}
												l38:
													add(rulePegText, position27)
												}
												if !_rules[ruleKEY]() {
													goto l24
												}
												{
													position40, tokenIndex40 := position, tokenIndex
													if !_rules[rule_]() {
														goto l41
													}
													{
														position42, tokenIndex42 := position, tokenIndex
														if buffer[position] != rune('b') {
															goto l43
														}
														position++
														goto l42
													l43:
														position, tokenIndex = position42, tokenIndex42
														if buffer[position] != rune('B') {
															goto l41
														}
														position++
													}
												l42:
													{
														position44, tokenIndex44 := position, tokenIndex
														if buffer[position] != rune('y') {
															goto l45
														}
														position++
														goto l44
													l45:
														position, tokenIndex = position44, tokenIndex44
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
													position, tokenIndex = position40, tokenIndex40
													if !(p.errorHere(position, `expected keyword "by" to follow keyword "sample"`)) {
														goto l24
													}
												}
											l40:
												break
											case 'R', 'r':
												{
													position46 := position
													{
														position47, tokenIndex47 := position, tokenIndex
														if buffer[position] != rune('r') {
															goto l48
														}
														position++
														goto l47
													l48:
														position, tokenIndex = position47, tokenIndex47
														if buffer[position] != rune('R') {
															goto l24
														}
														position++
													}
												l47:
													{
														position49, tokenIndex49 := position, tokenIndex
														if buffer[position] != rune('e') {
															goto l50
														}
														position++
														goto l49
													l50:
														position, tokenIndex = position49, tokenIndex49
														if buffer[position] != rune('E') {
															goto l24
														}
														position++
													}
												l49:
													{
														position51, tokenIndex51 := position, tokenIndex
														if buffer[position] != rune('s') {
															goto l52
														}
														position++
														goto l51
													l52:
														position, tokenIndex = position51, tokenIndex51
														if buffer[position] != rune('S') {
															goto l24
														}
														position++
													}
												l51:
													{
														position53, tokenIndex53 := position, tokenIndex
														if buffer[position] != rune('o') {
															goto l54
														}
														position++
														goto l53
													l54:
														position, tokenIndex = position53, tokenIndex53
														if buffer[position] != rune('O') {
															goto l24
														}
														position++
													}
												l53:
													{
														position55, tokenIndex55 := position, tokenIndex
														if buffer[position] != rune('l') {
															goto l56
														}
														position++
														goto l55
													l56:
														position, tokenIndex = position55, tokenIndex55
														if buffer[position] != rune('L') {
															goto l24
														}
														position++
													}
												l55:
													{
														position57, tokenIndex57 := position, tokenIndex
														if buffer[position] != rune('u') {
															goto l58
														}
														position++
														goto l57
													l58:
														position, tokenIndex = position57, tokenIndex57
														if buffer[position] != rune('U') {
															goto l24
														}
														position++
													}
												l57:
													{
														position59, tokenIndex59 := position, tokenIndex
														if buffer[position] != rune('t') {
															goto l60
														}
														position++
														goto l59
													l60:
														position, tokenIndex = position59, tokenIndex59
														if buffer[position] != rune('T') {
															goto l24
														}
														position++
													}
												l59:
													{
														position61, tokenIndex61 := position, tokenIndex
														if buffer[position] != rune('i') {
															goto l62
														}
														position++
														goto l61
													l62:
														position, tokenIndex = position61, tokenIndex61
														if buffer[position] != rune('I') {
															goto l24
														}
														position++
													}
												l61:
													{
														position63, tokenIndex63 := position, tokenIndex
														if buffer[position] != rune('o') {
															goto l64
														}
														position++
														goto l63
													l64:
														position, tokenIndex = position63, tokenIndex63
														if buffer[position] != rune('O') {
															goto l24
														}
														position++
													}
												l63:
													{
														position65, tokenIndex65 := position, tokenIndex
														if buffer[position] != rune('n') {
															goto l66
														}
														position++
														goto l65
													l66:
														position, tokenIndex = position65, tokenIndex65
														if buffer[position] != rune('N') {
															goto l24
														}
														position++
													}
												l65:
													add(rulePegText, position46)
												}
												if !_rules[ruleKEY]() {
													goto l24
												}
												break
											case 'T', 't':
												{
													position67 := position
													{
														position68, tokenIndex68 := position, tokenIndex
														if buffer[position] != rune('t') {
															goto l69
														}
														position++
														goto l68
													l69:
														position, tokenIndex = position68, tokenIndex68
														if buffer[position] != rune('T') {
															goto l24
														}
														position++
													}
												l68:
													{
														position70, tokenIndex70 := position, tokenIndex
														if buffer[position] != rune('o') {
															goto l71
														}
														position++
														goto l70
													l71:
														position, tokenIndex = position70, tokenIndex70
														if buffer[position] != rune('O') {
															goto l24
														}
														position++
													}
												l70:
													add(rulePegText, position67)
												}
												if !_rules[ruleKEY]() {
													goto l24
												}
												break
											default:
												{
													position72 := position
													{
														position73, tokenIndex73 := position, tokenIndex
														if buffer[position] != rune('f') {
															goto l74
														}
														position++
														goto l73
													l74:
														position, tokenIndex = position73, tokenIndex73
														if buffer[position] != rune('F') {
															goto l24
														}
														position++
													}
												l73:
													{
														position75, tokenIndex75 := position, tokenIndex
														if buffer[position] != rune('r') {
															goto l76
														}
														position++
														goto l75
													l76:
														position, tokenIndex = position75, tokenIndex75
														if buffer[position] != rune('R') {
															goto l24
														}
														position++
													}
												l75:
													{
														position77, tokenIndex77 := position, tokenIndex
														if buffer[position] != rune('o') {
															goto l78
														}
														position++
														goto l77
													l78:
														position, tokenIndex = position77, tokenIndex77
														if buffer[position] != rune('O') {
															goto l24
														}
														position++
													}
												l77:
													{
														position79, tokenIndex79 := position, tokenIndex
														if buffer[position] != rune('m') {
															goto l80
														}
														position++
														goto l79
													l80:
														position, tokenIndex = position79, tokenIndex79
														if buffer[position] != rune('M') {
															goto l24
														}
														position++
													}
												l79:
													add(rulePegText, position72)
												}
												if !_rules[ruleKEY]() {
													goto l24
												}
												break
											}
										}

										add(rulePROPERTY_KEY, position25)
									}
									{
										add(ruleAction8, position)
									}
									{
										position82, tokenIndex82 := position, tokenIndex
										if !_rules[rule_]() {
											goto l83
										}
										{
											position84 := position
											{
												position85 := position
												{
													position86, tokenIndex86 := position, tokenIndex
													if !_rules[rule_]() {
														goto l87
													}
													{
														position88 := position
														if !_rules[ruleNUMBER]() {
															goto l87
														}
													l89:
														{
															position90, tokenIndex90 := position, tokenIndex
															{
																position91, tokenIndex91 := position, tokenIndex
																if c := buffer[position]; c < rune('a') || c > rune('z') {
																	goto l92
																}
																position++
																goto l91
															l92:
																position, tokenIndex = position91, tokenIndex91
																if c := buffer[position]; c < rune('A') || c > rune('Z') {
																	goto l90
																}
																position++
															}
														l91:
															goto l89
														l90:
															position, tokenIndex = position90, tokenIndex90
														}
														add(rulePegText, position88)
													}
													goto l86
												l87:
													position, tokenIndex = position86, tokenIndex86
													if !_rules[rule_]() {
														goto l93
													}
													if !_rules[ruleSTRING]() {
														goto l93
													}
													goto l86
												l93:
													position, tokenIndex = position86, tokenIndex86
													if !_rules[rule_]() {
														goto l83
													}
													{
														position94 := position
														{
															position95, tokenIndex95 := position, tokenIndex
															if buffer[position] != rune('n') {
																goto l96
															}
															position++
															goto l95
														l96:
															position, tokenIndex = position95, tokenIndex95
															if buffer[position] != rune('N') {
																goto l83
															}
															position++
														}
													l95:
														{
															position97, tokenIndex97 := position, tokenIndex
															if buffer[position] != rune('o') {
																goto l98
															}
															position++
															goto l97
														l98:
															position, tokenIndex = position97, tokenIndex97
															if buffer[position] != rune('O') {
																goto l83
															}
															position++
														}
													l97:
														{
															position99, tokenIndex99 := position, tokenIndex
															if buffer[position] != rune('w') {
																goto l100
															}
															position++
															goto l99
														l100:
															position, tokenIndex = position99, tokenIndex99
															if buffer[position] != rune('W') {
																goto l83
															}
															position++
														}
													l99:
														add(rulePegText, position94)
													}
													if !_rules[ruleKEY]() {
														goto l83
													}
												}
											l86:
												add(ruleTIMESTAMP, position85)
											}
											add(rulePROPERTY_VALUE, position84)
										}
										{
											add(ruleAction9, position)
										}
										goto l82
									l83:
										position, tokenIndex = position82, tokenIndex82
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
									position, tokenIndex = position23, tokenIndex23
									if !_rules[rule_]() {
										goto l103
									}
									{
										position104, tokenIndex104 := position, tokenIndex
										if buffer[position] != rune('w') {
											goto l105
										}
										position++
										goto l104
									l105:
										position, tokenIndex = position104, tokenIndex104
										if buffer[position] != rune('W') {
											goto l103
										}
										position++
									}
								l104:
									{
										position106, tokenIndex106 := position, tokenIndex
										if buffer[position] != rune('h') {
											goto l107
										}
										position++
										goto l106
									l107:
										position, tokenIndex = position106, tokenIndex106
										if buffer[position] != rune('H') {
											goto l103
										}
										position++
									}
								l106:
									{
										position108, tokenIndex108 := position, tokenIndex
										if buffer[position] != rune('e') {
											goto l109
										}
										position++
										goto l108
									l109:
										position, tokenIndex = position108, tokenIndex108
										if buffer[position] != rune('E') {
											goto l103
										}
										position++
									}
								l108:
									{
										position110, tokenIndex110 := position, tokenIndex
										if buffer[position] != rune('r') {
											goto l111
										}
										position++
										goto l110
									l111:
										position, tokenIndex = position110, tokenIndex110
										if buffer[position] != rune('R') {
											goto l103
										}
										position++
									}
								l110:
									{
										position112, tokenIndex112 := position, tokenIndex
										if buffer[position] != rune('e') {
											goto l113
										}
										position++
										goto l112
									l113:
										position, tokenIndex = position112, tokenIndex112
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
									position, tokenIndex = position23, tokenIndex23
									if !_rules[rule_]() {
										goto l22
									}
									{
										position114, tokenIndex114 := position, tokenIndex
										{
											position115, tokenIndex115 := position, tokenIndex
											if !matchDot() {
												goto l115
											}
											goto l114
										l115:
											position, tokenIndex = position115, tokenIndex115
										}
										goto l22
									l114:
										position, tokenIndex = position114, tokenIndex114
									}
									if !(p.errorHere(position, `expected key (one of 'from', 'to', 'resolution', or 'sample by') or end of input but got %q following a completed expression`, p.after(position))) {
										goto l22
									}
								}
							l23:
								goto l21
							l22:
								position, tokenIndex = position22, tokenIndex22
							}
							{
								add(ruleAction11, position)
							}
							add(rulepropertyClause, position19)
						}
						{
							add(ruleAction0, position)
						}
						add(ruleselectStmt, position4)
					}
					goto l2
				l3:
					position, tokenIndex = position2, tokenIndex2
					{
						position118 := position
						if !_rules[rule_]() {
							goto l0
						}
						{
							position119, tokenIndex119 := position, tokenIndex
							if buffer[position] != rune('d') {
								goto l120
							}
							position++
							goto l119
						l120:
							position, tokenIndex = position119, tokenIndex119
							if buffer[position] != rune('D') {
								goto l0
							}
							position++
						}
					l119:
						{
							position121, tokenIndex121 := position, tokenIndex
							if buffer[position] != rune('e') {
								goto l122
							}
							position++
							goto l121
						l122:
							position, tokenIndex = position121, tokenIndex121
							if buffer[position] != rune('E') {
								goto l0
							}
							position++
						}
					l121:
						{
							position123, tokenIndex123 := position, tokenIndex
							if buffer[position] != rune('s') {
								goto l124
							}
							position++
							goto l123
						l124:
							position, tokenIndex = position123, tokenIndex123
							if buffer[position] != rune('S') {
								goto l0
							}
							position++
						}
					l123:
						{
							position125, tokenIndex125 := position, tokenIndex
							if buffer[position] != rune('c') {
								goto l126
							}
							position++
							goto l125
						l126:
							position, tokenIndex = position125, tokenIndex125
							if buffer[position] != rune('C') {
								goto l0
							}
							position++
						}
					l125:
						{
							position127, tokenIndex127 := position, tokenIndex
							if buffer[position] != rune('r') {
								goto l128
							}
							position++
							goto l127
						l128:
							position, tokenIndex = position127, tokenIndex127
							if buffer[position] != rune('R') {
								goto l0
							}
							position++
						}
					l127:
						{
							position129, tokenIndex129 := position, tokenIndex
							if buffer[position] != rune('i') {
								goto l130
							}
							position++
							goto l129
						l130:
							position, tokenIndex = position129, tokenIndex129
							if buffer[position] != rune('I') {
								goto l0
							}
							position++
						}
					l129:
						{
							position131, tokenIndex131 := position, tokenIndex
							if buffer[position] != rune('b') {
								goto l132
							}
							position++
							goto l131
						l132:
							position, tokenIndex = position131, tokenIndex131
							if buffer[position] != rune('B') {
								goto l0
							}
							position++
						}
					l131:
						{
							position133, tokenIndex133 := position, tokenIndex
							if buffer[position] != rune('e') {
								goto l134
							}
							position++
							goto l133
						l134:
							position, tokenIndex = position133, tokenIndex133
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
							position135, tokenIndex135 := position, tokenIndex
							{
								position137 := position
								if !_rules[rule_]() {
									goto l136
								}
								{
									position138, tokenIndex138 := position, tokenIndex
									if buffer[position] != rune('a') {
										goto l139
									}
									position++
									goto l138
								l139:
									position, tokenIndex = position138, tokenIndex138
									if buffer[position] != rune('A') {
										goto l136
									}
									position++
								}
							l138:
								{
									position140, tokenIndex140 := position, tokenIndex
									if buffer[position] != rune('l') {
										goto l141
									}
									position++
									goto l140
								l141:
									position, tokenIndex = position140, tokenIndex140
									if buffer[position] != rune('L') {
										goto l136
									}
									position++
								}
							l140:
								{
									position142, tokenIndex142 := position, tokenIndex
									if buffer[position] != rune('l') {
										goto l143
									}
									position++
									goto l142
								l143:
									position, tokenIndex = position142, tokenIndex142
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
									{
										position145, tokenIndex145 := position, tokenIndex
										{
											position147 := position
											if !_rules[rule_]() {
												goto l146
											}
											{
												position148, tokenIndex148 := position, tokenIndex
												if buffer[position] != rune('m') {
													goto l149
												}
												position++
												goto l148
											l149:
												position, tokenIndex = position148, tokenIndex148
												if buffer[position] != rune('M') {
													goto l146
												}
												position++
											}
										l148:
											{
												position150, tokenIndex150 := position, tokenIndex
												if buffer[position] != rune('a') {
													goto l151
												}
												position++
												goto l150
											l151:
												position, tokenIndex = position150, tokenIndex150
												if buffer[position] != rune('A') {
													goto l146
												}
												position++
											}
										l150:
											{
												position152, tokenIndex152 := position, tokenIndex
												if buffer[position] != rune('t') {
													goto l153
												}
												position++
												goto l152
											l153:
												position, tokenIndex = position152, tokenIndex152
												if buffer[position] != rune('T') {
													goto l146
												}
												position++
											}
										l152:
											{
												position154, tokenIndex154 := position, tokenIndex
												if buffer[position] != rune('c') {
													goto l155
												}
												position++
												goto l154
											l155:
												position, tokenIndex = position154, tokenIndex154
												if buffer[position] != rune('C') {
													goto l146
												}
												position++
											}
										l154:
											{
												position156, tokenIndex156 := position, tokenIndex
												if buffer[position] != rune('h') {
													goto l157
												}
												position++
												goto l156
											l157:
												position, tokenIndex = position156, tokenIndex156
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
												position158, tokenIndex158 := position, tokenIndex
												if !_rules[ruleliteralString]() {
													goto l159
												}
												goto l158
											l159:
												position, tokenIndex = position158, tokenIndex158
												if !(p.errorHere(position, `expected string literal to follow keyword "match"`)) {
													goto l146
												}
											}
										l158:
											{
												add(ruleAction3, position)
											}
											add(rulematchClause, position147)
										}
										goto l145
									l146:
										position, tokenIndex = position145, tokenIndex145
										{
											add(ruleAction2, position)
										}
									}
								l145:
									add(ruleoptionalMatchClause, position144)
								}
								{
									add(ruleAction1, position)
								}
								{
									position163, tokenIndex163 := position, tokenIndex
									{
										position164, tokenIndex164 := position, tokenIndex
										if !_rules[rule_]() {
											goto l165
										}
										{
											position166, tokenIndex166 := position, tokenIndex
											if !matchDot() {
												goto l166
											}
											goto l165
										l166:
											position, tokenIndex = position166, tokenIndex166
										}
										goto l164
									l165:
										position, tokenIndex = position164, tokenIndex164
										if !_rules[rule_]() {
											goto l136
										}
										if !(p.errorHere(position, `expected end of input after 'describe all' and optional match clause but got %q`, p.after(position))) {
											goto l136
										}
									}
								l164:
									position, tokenIndex = position163, tokenIndex163
								}
								add(ruledescribeAllStmt, position137)
							}
							goto l135
						l136:
							position, tokenIndex = position135, tokenIndex135
							{
								position168 := position
								if !_rules[rule_]() {
									goto l167
								}
								{
									position169, tokenIndex169 := position, tokenIndex
									if buffer[position] != rune('m') {
										goto l170
									}
									position++
									goto l169
								l170:
									position, tokenIndex = position169, tokenIndex169
									if buffer[position] != rune('M') {
										goto l167
									}
									position++
								}
							l169:
								{
									position171, tokenIndex171 := position, tokenIndex
									if buffer[position] != rune('e') {
										goto l172
									}
									position++
									goto l171
								l172:
									position, tokenIndex = position171, tokenIndex171
									if buffer[position] != rune('E') {
										goto l167
									}
									position++
								}
							l171:
								{
									position173, tokenIndex173 := position, tokenIndex
									if buffer[position] != rune('t') {
										goto l174
									}
									position++
									goto l173
								l174:
									position, tokenIndex = position173, tokenIndex173
									if buffer[position] != rune('T') {
										goto l167
									}
									position++
								}
							l173:
								{
									position175, tokenIndex175 := position, tokenIndex
									if buffer[position] != rune('r') {
										goto l176
									}
									position++
									goto l175
								l176:
									position, tokenIndex = position175, tokenIndex175
									if buffer[position] != rune('R') {
										goto l167
									}
									position++
								}
							l175:
								{
									position177, tokenIndex177 := position, tokenIndex
									if buffer[position] != rune('i') {
										goto l178
									}
									position++
									goto l177
								l178:
									position, tokenIndex = position177, tokenIndex177
									if buffer[position] != rune('I') {
										goto l167
									}
									position++
								}
							l177:
								{
									position179, tokenIndex179 := position, tokenIndex
									if buffer[position] != rune('c') {
										goto l180
									}
									position++
									goto l179
								l180:
									position, tokenIndex = position179, tokenIndex179
									if buffer[position] != rune('C') {
										goto l167
									}
									position++
								}
							l179:
								{
									position181, tokenIndex181 := position, tokenIndex
									if buffer[position] != rune('s') {
										goto l182
									}
									position++
									goto l181
								l182:
									position, tokenIndex = position181, tokenIndex181
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
									position183, tokenIndex183 := position, tokenIndex
									if !_rules[rule_]() {
										goto l184
									}
									{
										position185, tokenIndex185 := position, tokenIndex
										if buffer[position] != rune('w') {
											goto l186
										}
										position++
										goto l185
									l186:
										position, tokenIndex = position185, tokenIndex185
										if buffer[position] != rune('W') {
											goto l184
										}
										position++
									}
								l185:
									{
										position187, tokenIndex187 := position, tokenIndex
										if buffer[position] != rune('h') {
											goto l188
										}
										position++
										goto l187
									l188:
										position, tokenIndex = position187, tokenIndex187
										if buffer[position] != rune('H') {
											goto l184
										}
										position++
									}
								l187:
									{
										position189, tokenIndex189 := position, tokenIndex
										if buffer[position] != rune('e') {
											goto l190
										}
										position++
										goto l189
									l190:
										position, tokenIndex = position189, tokenIndex189
										if buffer[position] != rune('E') {
											goto l184
										}
										position++
									}
								l189:
									{
										position191, tokenIndex191 := position, tokenIndex
										if buffer[position] != rune('r') {
											goto l192
										}
										position++
										goto l191
									l192:
										position, tokenIndex = position191, tokenIndex191
										if buffer[position] != rune('R') {
											goto l184
										}
										position++
									}
								l191:
									{
										position193, tokenIndex193 := position, tokenIndex
										if buffer[position] != rune('e') {
											goto l194
										}
										position++
										goto l193
									l194:
										position, tokenIndex = position193, tokenIndex193
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
									position, tokenIndex = position183, tokenIndex183
									if !(p.errorHere(position, `expected "where" to follow keyword "metrics" in "describe metrics" command`)) {
										goto l167
									}
								}
							l183:
								{
									position195, tokenIndex195 := position, tokenIndex
									if !_rules[ruletagName]() {
										goto l196
									}
									goto l195
								l196:
									position, tokenIndex = position195, tokenIndex195
									if !(p.errorHere(position, `expected tag key to follow keyword "where" in "describe metrics" command`)) {
										goto l167
									}
								}
							l195:
								{
									position197, tokenIndex197 := position, tokenIndex
									if !_rules[rule_]() {
										goto l198
									}
									if buffer[position] != rune('=') {
										goto l198
									}
									position++
									goto l197
								l198:
									position, tokenIndex = position197, tokenIndex197
									if !(p.errorHere(position, `expected "=" to follow keyword "where" in "describe metrics" command`)) {
										goto l167
									}
								}
							l197:
								{
									position199, tokenIndex199 := position, tokenIndex
									if !_rules[ruleliteralString]() {
										goto l200
									}
									goto l199
								l200:
									position, tokenIndex = position199, tokenIndex199
									if !(p.errorHere(position, `expected string literal to follow "=" in "describe metrics" command`)) {
										goto l167
									}
								}
							l199:
								{
									add(ruleAction4, position)
								}
								add(ruledescribeMetrics, position168)
							}
							goto l135
						l167:
							position, tokenIndex = position135, tokenIndex135
							{
								position202 := position
								{
									position203, tokenIndex203 := position, tokenIndex
									if !_rules[rule_]() {
										goto l204
									}
									{
										position205 := position
										{
											position206 := position
											if !_rules[ruleIDENTIFIER]() {
												goto l204
											}
											add(ruleMETRIC_NAME, position206)
										}
										add(rulePegText, position205)
									}
									{
										add(ruleAction5, position)
									}
									goto l203
								l204:
									position, tokenIndex = position203, tokenIndex203
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
								add(ruledescribeSingleStmt, position202)
							}
						}
					l135:
						add(ruledescribeStmt, position118)
					}
				}
			l2:
				if !_rules[rule_]() {
					goto l0
				}
				{
					position209, tokenIndex209 := position, tokenIndex
					if !matchDot() {
						goto l209
					}
					goto l0
				l209:
					position, tokenIndex = position209, tokenIndex209
				}
				add(ruleroot, position1)
			}
			return true
		l0:
			position, tokenIndex = position0, tokenIndex0
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
				{
					position220, tokenIndex220 := position, tokenIndex
					{
						position222 := position
						if !_rules[rule_]() {
							goto l221
						}
						{
							position223, tokenIndex223 := position, tokenIndex
							if buffer[position] != rune('w') {
								goto l224
							}
							position++
							goto l223
						l224:
							position, tokenIndex = position223, tokenIndex223
							if buffer[position] != rune('W') {
								goto l221
							}
							position++
						}
					l223:
						{
							position225, tokenIndex225 := position, tokenIndex
							if buffer[position] != rune('h') {
								goto l226
							}
							position++
							goto l225
						l226:
							position, tokenIndex = position225, tokenIndex225
							if buffer[position] != rune('H') {
								goto l221
							}
							position++
						}
					l225:
						{
							position227, tokenIndex227 := position, tokenIndex
							if buffer[position] != rune('e') {
								goto l228
							}
							position++
							goto l227
						l228:
							position, tokenIndex = position227, tokenIndex227
							if buffer[position] != rune('E') {
								goto l221
							}
							position++
						}
					l227:
						{
							position229, tokenIndex229 := position, tokenIndex
							if buffer[position] != rune('r') {
								goto l230
							}
							position++
							goto l229
						l230:
							position, tokenIndex = position229, tokenIndex229
							if buffer[position] != rune('R') {
								goto l221
							}
							position++
						}
					l229:
						{
							position231, tokenIndex231 := position, tokenIndex
							if buffer[position] != rune('e') {
								goto l232
							}
							position++
							goto l231
						l232:
							position, tokenIndex = position231, tokenIndex231
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
							position233, tokenIndex233 := position, tokenIndex
							if !_rules[rule_]() {
								goto l234
							}
							if !_rules[rulepredicate_1]() {
								goto l234
							}
							goto l233
						l234:
							position, tokenIndex = position233, tokenIndex233
							if !(p.errorHere(position, `expected predicate to follow "where" keyword`)) {
								goto l221
							}
						}
					l233:
						add(rulepredicateClause, position222)
					}
					goto l220
				l221:
					position, tokenIndex = position220, tokenIndex220
					{
						add(ruleAction12, position)
					}
				}
			l220:
				add(ruleoptionalPredicateClause, position219)
			}
			return true
		},
		/* 10 expressionList <- <(Action13 expression_start Action14 (_ COMMA (expression_start / &{ p.errorHere(position, `expected expression to follow ","`) }) Action15)*)> */
		func() bool {
			position236, tokenIndex236 := position, tokenIndex
			{
				position237 := position
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
					position241, tokenIndex241 := position, tokenIndex
					if !_rules[rule_]() {
						goto l241
					}
					if !_rules[ruleCOMMA]() {
						goto l241
					}
					{
						position242, tokenIndex242 := position, tokenIndex
						if !_rules[ruleexpression_start]() {
							goto l243
						}
						goto l242
					l243:
						position, tokenIndex = position242, tokenIndex242
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
					position, tokenIndex = position241, tokenIndex241
				}
				add(ruleexpressionList, position237)
			}
			return true
		l236:
			position, tokenIndex = position236, tokenIndex236
			return false
		},
		/* 11 expression_start <- <(expression_sum add_pipe)> */
		func() bool {
			position245, tokenIndex245 := position, tokenIndex
			{
				position246 := position
				{
					position247 := position
					if !_rules[ruleexpression_product]() {
						goto l245
					}
				l248:
					{
						position249, tokenIndex249 := position, tokenIndex
						if !_rules[ruleadd_pipe]() {
							goto l249
						}
						{
							position250, tokenIndex250 := position, tokenIndex
							if !_rules[rule_]() {
								goto l251
							}
							{
								position252 := position
								if buffer[position] != rune('+') {
									goto l251
								}
								position++
								add(ruleOP_ADD, position252)
							}
							{
								add(ruleAction16, position)
							}
							goto l250
						l251:
							position, tokenIndex = position250, tokenIndex250
							if !_rules[rule_]() {
								goto l249
							}
							{
								position254 := position
								if buffer[position] != rune('-') {
									goto l249
								}
								position++
								add(ruleOP_SUB, position254)
							}
							{
								add(ruleAction17, position)
							}
						}
					l250:
						{
							position256, tokenIndex256 := position, tokenIndex
							if !_rules[ruleexpression_product]() {
								goto l257
							}
							goto l256
						l257:
							position, tokenIndex = position256, tokenIndex256
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
						position, tokenIndex = position249, tokenIndex249
					}
					add(ruleexpression_sum, position247)
				}
				if !_rules[ruleadd_pipe]() {
					goto l245
				}
				add(ruleexpression_start, position246)
			}
			return true
		l245:
			position, tokenIndex = position245, tokenIndex245
			return false
		},
		/* 12 expression_sum <- <(expression_product (add_pipe ((_ OP_ADD Action16) / (_ OP_SUB Action17)) (expression_product / &{ p.errorHere(position, `expected expression to follow operator "+" or "-"`) }) Action18)*)> */
		nil,
		/* 13 expression_product <- <(expression_atom (add_pipe ((_ OP_DIV Action19) / (_ OP_MULT Action20)) (expression_atom / &{ p.errorHere(position, `expected expression to follow operator "*" or "/"`) }) Action21)*)> */
		func() bool {
			position260, tokenIndex260 := position, tokenIndex
			{
				position261 := position
				if !_rules[ruleexpression_atom]() {
					goto l260
				}
			l262:
				{
					position263, tokenIndex263 := position, tokenIndex
					if !_rules[ruleadd_pipe]() {
						goto l263
					}
					{
						position264, tokenIndex264 := position, tokenIndex
						if !_rules[rule_]() {
							goto l265
						}
						{
							position266 := position
							if buffer[position] != rune('/') {
								goto l265
							}
							position++
							add(ruleOP_DIV, position266)
						}
						{
							add(ruleAction19, position)
						}
						goto l264
					l265:
						position, tokenIndex = position264, tokenIndex264
						if !_rules[rule_]() {
							goto l263
						}
						{
							position268 := position
							if buffer[position] != rune('*') {
								goto l263
							}
							position++
							add(ruleOP_MULT, position268)
						}
						{
							add(ruleAction20, position)
						}
					}
				l264:
					{
						position270, tokenIndex270 := position, tokenIndex
						if !_rules[ruleexpression_atom]() {
							goto l271
						}
						goto l270
					l271:
						position, tokenIndex = position270, tokenIndex270
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
					position, tokenIndex = position263, tokenIndex263
				}
				add(ruleexpression_product, position261)
			}
			return true
		l260:
			position, tokenIndex = position260, tokenIndex260
			return false
		},
		/* 14 add_one_pipe <- <(_ OP_PIPE ((_ <IDENTIFIER>) / &{ p.errorHere(position, `expected function name to follow pipe "|"`) }) Action22 ((_ PAREN_OPEN (expressionList / Action23) optionalGroupBy ((_ PAREN_CLOSE) / &{ p.errorHere(position, `expected ")" to close "(" opened in pipe function call`) })) / Action24) Action25 expression_annotation)> */
		nil,
		/* 15 add_pipe <- <add_one_pipe*> */
		func() bool {
			{
				position275 := position
			l276:
				{
					position277, tokenIndex277 := position, tokenIndex
					{
						position278 := position
						if !_rules[rule_]() {
							goto l277
						}
						{
							position279 := position
							if buffer[position] != rune('|') {
								goto l277
							}
							position++
							add(ruleOP_PIPE, position279)
						}
						{
							position280, tokenIndex280 := position, tokenIndex
							if !_rules[rule_]() {
								goto l281
							}
							{
								position282 := position
								if !_rules[ruleIDENTIFIER]() {
									goto l281
								}
								add(rulePegText, position282)
							}
							goto l280
						l281:
							position, tokenIndex = position280, tokenIndex280
							if !(p.errorHere(position, `expected function name to follow pipe "|"`)) {
								goto l277
							}
						}
					l280:
						{
							add(ruleAction22, position)
						}
						{
							position284, tokenIndex284 := position, tokenIndex
							if !_rules[rule_]() {
								goto l285
							}
							if !_rules[rulePAREN_OPEN]() {
								goto l285
							}
							{
								position286, tokenIndex286 := position, tokenIndex
								if !_rules[ruleexpressionList]() {
									goto l287
								}
								goto l286
							l287:
								position, tokenIndex = position286, tokenIndex286
								{
									add(ruleAction23, position)
								}
							}
						l286:
							if !_rules[ruleoptionalGroupBy]() {
								goto l285
							}
							{
								position289, tokenIndex289 := position, tokenIndex
								if !_rules[rule_]() {
									goto l290
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l290
								}
								goto l289
							l290:
								position, tokenIndex = position289, tokenIndex289
								if !(p.errorHere(position, `expected ")" to close "(" opened in pipe function call`)) {
									goto l285
								}
							}
						l289:
							goto l284
						l285:
							position, tokenIndex = position284, tokenIndex284
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
						add(ruleadd_one_pipe, position278)
					}
					goto l276
				l277:
					position, tokenIndex = position277, tokenIndex277
				}
				add(ruleadd_pipe, position275)
			}
			return true
		},
		/* 16 expression_atom <- <(expression_atom_raw expression_annotation)> */
		func() bool {
			position293, tokenIndex293 := position, tokenIndex
			{
				position294 := position
				{
					position295 := position
					{
						position296, tokenIndex296 := position, tokenIndex
						{
							position298 := position
							if !_rules[rule_]() {
								goto l297
							}
							{
								position299 := position
								if !_rules[ruleIDENTIFIER]() {
									goto l297
								}
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
								position301, tokenIndex301 := position, tokenIndex
								if !_rules[ruleexpressionList]() {
									goto l302
								}
								goto l301
							l302:
								position, tokenIndex = position301, tokenIndex301
								if !(p.errorHere(position, `expected expression list to follow "(" in function call`)) {
									goto l297
								}
							}
						l301:
							if !_rules[ruleoptionalGroupBy]() {
								goto l297
							}
							{
								position303, tokenIndex303 := position, tokenIndex
								if !_rules[rule_]() {
									goto l304
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l304
								}
								goto l303
							l304:
								position, tokenIndex = position303, tokenIndex303
								if !(p.errorHere(position, `expected ")" to close "(" opened by function call`)) {
									goto l297
								}
							}
						l303:
							{
								add(ruleAction32, position)
							}
							add(ruleexpression_function, position298)
						}
						goto l296
					l297:
						position, tokenIndex = position296, tokenIndex296
						{
							position307 := position
							if !_rules[rule_]() {
								goto l306
							}
							{
								position308 := position
								if !_rules[ruleIDENTIFIER]() {
									goto l306
								}
								add(rulePegText, position308)
							}
							{
								add(ruleAction33, position)
							}
							{
								position310, tokenIndex310 := position, tokenIndex
								if !_rules[rule_]() {
									goto l311
								}
								if buffer[position] != rune('[') {
									goto l311
								}
								position++
								{
									position312, tokenIndex312 := position, tokenIndex
									if !_rules[rulepredicate_1]() {
										goto l313
									}
									goto l312
								l313:
									position, tokenIndex = position312, tokenIndex312
									if !(p.errorHere(position, `expected predicate to follow "[" after metric`)) {
										goto l311
									}
								}
							l312:
								{
									position314, tokenIndex314 := position, tokenIndex
									if !_rules[rule_]() {
										goto l315
									}
									if buffer[position] != rune(']') {
										goto l315
									}
									position++
									goto l314
								l315:
									position, tokenIndex = position314, tokenIndex314
									if !(p.errorHere(position, `expected "]" to close "[" opened to apply predicate`)) {
										goto l311
									}
								}
							l314:
								goto l310
							l311:
								position, tokenIndex = position310, tokenIndex310
								{
									add(ruleAction34, position)
								}
							}
						l310:
							{
								add(ruleAction35, position)
							}
							add(ruleexpression_metric, position307)
						}
						goto l296
					l306:
						position, tokenIndex = position296, tokenIndex296
						if !_rules[rule_]() {
							goto l318
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l318
						}
						{
							position319, tokenIndex319 := position, tokenIndex
							if !_rules[ruleexpression_start]() {
								goto l320
							}
							goto l319
						l320:
							position, tokenIndex = position319, tokenIndex319
							if !(p.errorHere(position, `expected expression to follow "("`)) {
								goto l318
							}
						}
					l319:
						{
							position321, tokenIndex321 := position, tokenIndex
							if !_rules[rule_]() {
								goto l322
							}
							if !_rules[rulePAREN_CLOSE]() {
								goto l322
							}
							goto l321
						l322:
							position, tokenIndex = position321, tokenIndex321
							if !(p.errorHere(position, `expected ")" to close "("`)) {
								goto l318
							}
						}
					l321:
						goto l296
					l318:
						position, tokenIndex = position296, tokenIndex296
						if !_rules[rule_]() {
							goto l323
						}
						{
							position324 := position
							{
								position325 := position
								if !_rules[ruleNUMBER]() {
									goto l323
								}
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l323
								}
								position++
							l326:
								{
									position327, tokenIndex327 := position, tokenIndex
									if c := buffer[position]; c < rune('a') || c > rune('z') {
										goto l327
									}
									position++
									goto l326
								l327:
									position, tokenIndex = position327, tokenIndex327
								}
								if !_rules[ruleKEY]() {
									goto l323
								}
								add(ruleDURATION, position325)
							}
							add(rulePegText, position324)
						}
						{
							add(ruleAction26, position)
						}
						goto l296
					l323:
						position, tokenIndex = position296, tokenIndex296
						if !_rules[rule_]() {
							goto l329
						}
						{
							position330 := position
							if !_rules[ruleNUMBER]() {
								goto l329
							}
							add(rulePegText, position330)
						}
						{
							add(ruleAction27, position)
						}
						goto l296
					l329:
						position, tokenIndex = position296, tokenIndex296
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
					add(ruleexpression_atom_raw, position295)
				}
				if !_rules[ruleexpression_annotation]() {
					goto l293
				}
				add(ruleexpression_atom, position294)
			}
			return true
		l293:
			position, tokenIndex = position293, tokenIndex293
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
				{
					position337, tokenIndex337 := position, tokenIndex
					{
						position339 := position
						if !_rules[rule_]() {
							goto l337
						}
						if buffer[position] != rune('{') {
							goto l337
						}
						position++
						{
							position340 := position
						l341:
							{
								position342, tokenIndex342 := position, tokenIndex
								{
									position343, tokenIndex343 := position, tokenIndex
									if buffer[position] != rune('}') {
										goto l343
									}
									position++
									goto l342
								l343:
									position, tokenIndex = position343, tokenIndex343
								}
								if !matchDot() {
									goto l342
								}
								goto l341
							l342:
								position, tokenIndex = position342, tokenIndex342
							}
							add(rulePegText, position340)
						}
						{
							position344, tokenIndex344 := position, tokenIndex
							if buffer[position] != rune('}') {
								goto l345
							}
							position++
							goto l344
						l345:
							position, tokenIndex = position344, tokenIndex344
							if !(p.errorHere(position, `expected "$CLOSEBRACE$" to close "$OPENBRACE$" opened for annotation`)) {
								goto l337
							}
						}
					l344:
						{
							add(ruleAction29, position)
						}
						add(ruleexpression_annotation_required, position339)
					}
					goto l338
				l337:
					position, tokenIndex = position337, tokenIndex337
				}
			l338:
				add(ruleexpression_annotation, position336)
			}
			return true
		},
		/* 20 optionalGroupBy <- <(groupByClause / collapseByClause / Action30)?> */
		func() bool {
			{
				position348 := position
				{
					position349, tokenIndex349 := position, tokenIndex
					{
						position351, tokenIndex351 := position, tokenIndex
						{
							position353 := position
							if !_rules[rule_]() {
								goto l352
							}
							{
								position354, tokenIndex354 := position, tokenIndex
								if buffer[position] != rune('g') {
									goto l355
								}
								position++
								goto l354
							l355:
								position, tokenIndex = position354, tokenIndex354
								if buffer[position] != rune('G') {
									goto l352
								}
								position++
							}
						l354:
							{
								position356, tokenIndex356 := position, tokenIndex
								if buffer[position] != rune('r') {
									goto l357
								}
								position++
								goto l356
							l357:
								position, tokenIndex = position356, tokenIndex356
								if buffer[position] != rune('R') {
									goto l352
								}
								position++
							}
						l356:
							{
								position358, tokenIndex358 := position, tokenIndex
								if buffer[position] != rune('o') {
									goto l359
								}
								position++
								goto l358
							l359:
								position, tokenIndex = position358, tokenIndex358
								if buffer[position] != rune('O') {
									goto l352
								}
								position++
							}
						l358:
							{
								position360, tokenIndex360 := position, tokenIndex
								if buffer[position] != rune('u') {
									goto l361
								}
								position++
								goto l360
							l361:
								position, tokenIndex = position360, tokenIndex360
								if buffer[position] != rune('U') {
									goto l352
								}
								position++
							}
						l360:
							{
								position362, tokenIndex362 := position, tokenIndex
								if buffer[position] != rune('p') {
									goto l363
								}
								position++
								goto l362
							l363:
								position, tokenIndex = position362, tokenIndex362
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
								position364, tokenIndex364 := position, tokenIndex
								if !_rules[rule_]() {
									goto l365
								}
								{
									position366, tokenIndex366 := position, tokenIndex
									if buffer[position] != rune('b') {
										goto l367
									}
									position++
									goto l366
								l367:
									position, tokenIndex = position366, tokenIndex366
									if buffer[position] != rune('B') {
										goto l365
									}
									position++
								}
							l366:
								{
									position368, tokenIndex368 := position, tokenIndex
									if buffer[position] != rune('y') {
										goto l369
									}
									position++
									goto l368
								l369:
									position, tokenIndex = position368, tokenIndex368
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
								position, tokenIndex = position364, tokenIndex364
								if !(p.errorHere(position, `expected keyword "by" to follow keyword "group" in "group by" clause`)) {
									goto l352
								}
							}
						l364:
							{
								position370, tokenIndex370 := position, tokenIndex
								if !_rules[rule_]() {
									goto l371
								}
								{
									position372 := position
									if !_rules[ruleCOLUMN_NAME]() {
										goto l371
									}
									add(rulePegText, position372)
								}
								goto l370
							l371:
								position, tokenIndex = position370, tokenIndex370
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
								position376, tokenIndex376 := position, tokenIndex
								if !_rules[rule_]() {
									goto l376
								}
								if !_rules[ruleCOMMA]() {
									goto l376
								}
								{
									position377, tokenIndex377 := position, tokenIndex
									if !_rules[rule_]() {
										goto l378
									}
									{
										position379 := position
										if !_rules[ruleCOLUMN_NAME]() {
											goto l378
										}
										add(rulePegText, position379)
									}
									goto l377
								l378:
									position, tokenIndex = position377, tokenIndex377
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
								position, tokenIndex = position376, tokenIndex376
							}
							add(rulegroupByClause, position353)
						}
						goto l351
					l352:
						position, tokenIndex = position351, tokenIndex351
						{
							position382 := position
							if !_rules[rule_]() {
								goto l381
							}
							{
								position383, tokenIndex383 := position, tokenIndex
								if buffer[position] != rune('c') {
									goto l384
								}
								position++
								goto l383
							l384:
								position, tokenIndex = position383, tokenIndex383
								if buffer[position] != rune('C') {
									goto l381
								}
								position++
							}
						l383:
							{
								position385, tokenIndex385 := position, tokenIndex
								if buffer[position] != rune('o') {
									goto l386
								}
								position++
								goto l385
							l386:
								position, tokenIndex = position385, tokenIndex385
								if buffer[position] != rune('O') {
									goto l381
								}
								position++
							}
						l385:
							{
								position387, tokenIndex387 := position, tokenIndex
								if buffer[position] != rune('l') {
									goto l388
								}
								position++
								goto l387
							l388:
								position, tokenIndex = position387, tokenIndex387
								if buffer[position] != rune('L') {
									goto l381
								}
								position++
							}
						l387:
							{
								position389, tokenIndex389 := position, tokenIndex
								if buffer[position] != rune('l') {
									goto l390
								}
								position++
								goto l389
							l390:
								position, tokenIndex = position389, tokenIndex389
								if buffer[position] != rune('L') {
									goto l381
								}
								position++
							}
						l389:
							{
								position391, tokenIndex391 := position, tokenIndex
								if buffer[position] != rune('a') {
									goto l392
								}
								position++
								goto l391
							l392:
								position, tokenIndex = position391, tokenIndex391
								if buffer[position] != rune('A') {
									goto l381
								}
								position++
							}
						l391:
							{
								position393, tokenIndex393 := position, tokenIndex
								if buffer[position] != rune('p') {
									goto l394
								}
								position++
								goto l393
							l394:
								position, tokenIndex = position393, tokenIndex393
								if buffer[position] != rune('P') {
									goto l381
								}
								position++
							}
						l393:
							{
								position395, tokenIndex395 := position, tokenIndex
								if buffer[position] != rune('s') {
									goto l396
								}
								position++
								goto l395
							l396:
								position, tokenIndex = position395, tokenIndex395
								if buffer[position] != rune('S') {
									goto l381
								}
								position++
							}
						l395:
							{
								position397, tokenIndex397 := position, tokenIndex
								if buffer[position] != rune('e') {
									goto l398
								}
								position++
								goto l397
							l398:
								position, tokenIndex = position397, tokenIndex397
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
								position399, tokenIndex399 := position, tokenIndex
								if !_rules[rule_]() {
									goto l400
								}
								{
									position401, tokenIndex401 := position, tokenIndex
									if buffer[position] != rune('b') {
										goto l402
									}
									position++
									goto l401
								l402:
									position, tokenIndex = position401, tokenIndex401
									if buffer[position] != rune('B') {
										goto l400
									}
									position++
								}
							l401:
								{
									position403, tokenIndex403 := position, tokenIndex
									if buffer[position] != rune('y') {
										goto l404
									}
									position++
									goto l403
								l404:
									position, tokenIndex = position403, tokenIndex403
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
								position, tokenIndex = position399, tokenIndex399
								if !(p.errorHere(position, `expected keyword "by" to follow keyword "collapse" in "collapse by" clause`)) {
									goto l381
								}
							}
						l399:
							{
								position405, tokenIndex405 := position, tokenIndex
								if !_rules[rule_]() {
									goto l406
								}
								{
									position407 := position
									if !_rules[ruleCOLUMN_NAME]() {
										goto l406
									}
									add(rulePegText, position407)
								}
								goto l405
							l406:
								position, tokenIndex = position405, tokenIndex405
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
								position411, tokenIndex411 := position, tokenIndex
								if !_rules[rule_]() {
									goto l411
								}
								if !_rules[ruleCOMMA]() {
									goto l411
								}
								{
									position412, tokenIndex412 := position, tokenIndex
									if !_rules[rule_]() {
										goto l413
									}
									{
										position414 := position
										if !_rules[ruleCOLUMN_NAME]() {
											goto l413
										}
										add(rulePegText, position414)
									}
									goto l412
								l413:
									position, tokenIndex = position412, tokenIndex412
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
								position, tokenIndex = position411, tokenIndex411
							}
							add(rulecollapseByClause, position382)
						}
						goto l351
					l381:
						position, tokenIndex = position351, tokenIndex351
						{
							add(ruleAction30, position)
						}
					}
				l351:
					goto l350

					position, tokenIndex = position349, tokenIndex349
				}
			l350:
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
			position422, tokenIndex422 := position, tokenIndex
			{
				position423 := position
				{
					position424, tokenIndex424 := position, tokenIndex
					if !_rules[rulepredicate_2]() {
						goto l425
					}
					if !_rules[rule_]() {
						goto l425
					}
					{
						position426 := position
						{
							position427, tokenIndex427 := position, tokenIndex
							if buffer[position] != rune('o') {
								goto l428
							}
							position++
							goto l427
						l428:
							position, tokenIndex = position427, tokenIndex427
							if buffer[position] != rune('O') {
								goto l425
							}
							position++
						}
					l427:
						{
							position429, tokenIndex429 := position, tokenIndex
							if buffer[position] != rune('r') {
								goto l430
							}
							position++
							goto l429
						l430:
							position, tokenIndex = position429, tokenIndex429
							if buffer[position] != rune('R') {
								goto l425
							}
							position++
						}
					l429:
						if !_rules[ruleKEY]() {
							goto l425
						}
						add(ruleOP_OR, position426)
					}
					{
						position431, tokenIndex431 := position, tokenIndex
						if !_rules[rulepredicate_1]() {
							goto l432
						}
						goto l431
					l432:
						position, tokenIndex = position431, tokenIndex431
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
					position, tokenIndex = position424, tokenIndex424
					if !_rules[rulepredicate_2]() {
						goto l422
					}
				}
			l424:
				add(rulepredicate_1, position423)
			}
			return true
		l422:
			position, tokenIndex = position422, tokenIndex422
			return false
		},
		/* 27 predicate_2 <- <((predicate_3 _ OP_AND (predicate_2 / &{ p.errorHere(position, `expected predicate to follow "and" operator`) }) Action43) / predicate_3)> */
		func() bool {
			position434, tokenIndex434 := position, tokenIndex
			{
				position435 := position
				{
					position436, tokenIndex436 := position, tokenIndex
					if !_rules[rulepredicate_3]() {
						goto l437
					}
					if !_rules[rule_]() {
						goto l437
					}
					{
						position438 := position
						{
							position439, tokenIndex439 := position, tokenIndex
							if buffer[position] != rune('a') {
								goto l440
							}
							position++
							goto l439
						l440:
							position, tokenIndex = position439, tokenIndex439
							if buffer[position] != rune('A') {
								goto l437
							}
							position++
						}
					l439:
						{
							position441, tokenIndex441 := position, tokenIndex
							if buffer[position] != rune('n') {
								goto l442
							}
							position++
							goto l441
						l442:
							position, tokenIndex = position441, tokenIndex441
							if buffer[position] != rune('N') {
								goto l437
							}
							position++
						}
					l441:
						{
							position443, tokenIndex443 := position, tokenIndex
							if buffer[position] != rune('d') {
								goto l444
							}
							position++
							goto l443
						l444:
							position, tokenIndex = position443, tokenIndex443
							if buffer[position] != rune('D') {
								goto l437
							}
							position++
						}
					l443:
						if !_rules[ruleKEY]() {
							goto l437
						}
						add(ruleOP_AND, position438)
					}
					{
						position445, tokenIndex445 := position, tokenIndex
						if !_rules[rulepredicate_2]() {
							goto l446
						}
						goto l445
					l446:
						position, tokenIndex = position445, tokenIndex445
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
					position, tokenIndex = position436, tokenIndex436
					if !_rules[rulepredicate_3]() {
						goto l434
					}
				}
			l436:
				add(rulepredicate_2, position435)
			}
			return true
		l434:
			position, tokenIndex = position434, tokenIndex434
			return false
		},
		/* 28 predicate_3 <- <((_ OP_NOT (predicate_3 / &{ p.errorHere(position, `expected predicate to follow "not" operator`) }) Action44) / (_ PAREN_OPEN (predicate_1 / &{ p.errorHere(position, `expected predicate to follow "("`) }) ((_ PAREN_CLOSE) / &{ p.errorHere(position, `expected ")" to close "(" opened in predicate`) })) / tagMatcher)> */
		func() bool {
			position448, tokenIndex448 := position, tokenIndex
			{
				position449 := position
				{
					position450, tokenIndex450 := position, tokenIndex
					if !_rules[rule_]() {
						goto l451
					}
					{
						position452 := position
						{
							position453, tokenIndex453 := position, tokenIndex
							if buffer[position] != rune('n') {
								goto l454
							}
							position++
							goto l453
						l454:
							position, tokenIndex = position453, tokenIndex453
							if buffer[position] != rune('N') {
								goto l451
							}
							position++
						}
					l453:
						{
							position455, tokenIndex455 := position, tokenIndex
							if buffer[position] != rune('o') {
								goto l456
							}
							position++
							goto l455
						l456:
							position, tokenIndex = position455, tokenIndex455
							if buffer[position] != rune('O') {
								goto l451
							}
							position++
						}
					l455:
						{
							position457, tokenIndex457 := position, tokenIndex
							if buffer[position] != rune('t') {
								goto l458
							}
							position++
							goto l457
						l458:
							position, tokenIndex = position457, tokenIndex457
							if buffer[position] != rune('T') {
								goto l451
							}
							position++
						}
					l457:
						if !_rules[ruleKEY]() {
							goto l451
						}
						add(ruleOP_NOT, position452)
					}
					{
						position459, tokenIndex459 := position, tokenIndex
						if !_rules[rulepredicate_3]() {
							goto l460
						}
						goto l459
					l460:
						position, tokenIndex = position459, tokenIndex459
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
					position, tokenIndex = position450, tokenIndex450
					if !_rules[rule_]() {
						goto l462
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l462
					}
					{
						position463, tokenIndex463 := position, tokenIndex
						if !_rules[rulepredicate_1]() {
							goto l464
						}
						goto l463
					l464:
						position, tokenIndex = position463, tokenIndex463
						if !(p.errorHere(position, `expected predicate to follow "("`)) {
							goto l462
						}
					}
				l463:
					{
						position465, tokenIndex465 := position, tokenIndex
						if !_rules[rule_]() {
							goto l466
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l466
						}
						goto l465
					l466:
						position, tokenIndex = position465, tokenIndex465
						if !(p.errorHere(position, `expected ")" to close "(" opened in predicate`)) {
							goto l462
						}
					}
				l465:
					goto l450
				l462:
					position, tokenIndex = position450, tokenIndex450
					{
						position467 := position
						if !_rules[ruletagName]() {
							goto l448
						}
						{
							position468, tokenIndex468 := position, tokenIndex
							if !_rules[rule_]() {
								goto l469
							}
							if buffer[position] != rune('=') {
								goto l469
							}
							position++
							{
								position470, tokenIndex470 := position, tokenIndex
								if !_rules[ruleliteralString]() {
									goto l471
								}
								goto l470
							l471:
								position, tokenIndex = position470, tokenIndex470
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
							position, tokenIndex = position468, tokenIndex468
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
								position474, tokenIndex474 := position, tokenIndex
								if !_rules[ruleliteralString]() {
									goto l475
								}
								goto l474
							l475:
								position, tokenIndex = position474, tokenIndex474
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
							position, tokenIndex = position468, tokenIndex468
							if !_rules[rule_]() {
								goto l478
							}
							{
								position479, tokenIndex479 := position, tokenIndex
								if buffer[position] != rune('m') {
									goto l480
								}
								position++
								goto l479
							l480:
								position, tokenIndex = position479, tokenIndex479
								if buffer[position] != rune('M') {
									goto l478
								}
								position++
							}
						l479:
							{
								position481, tokenIndex481 := position, tokenIndex
								if buffer[position] != rune('a') {
									goto l482
								}
								position++
								goto l481
							l482:
								position, tokenIndex = position481, tokenIndex481
								if buffer[position] != rune('A') {
									goto l478
								}
								position++
							}
						l481:
							{
								position483, tokenIndex483 := position, tokenIndex
								if buffer[position] != rune('t') {
									goto l484
								}
								position++
								goto l483
							l484:
								position, tokenIndex = position483, tokenIndex483
								if buffer[position] != rune('T') {
									goto l478
								}
								position++
							}
						l483:
							{
								position485, tokenIndex485 := position, tokenIndex
								if buffer[position] != rune('c') {
									goto l486
								}
								position++
								goto l485
							l486:
								position, tokenIndex = position485, tokenIndex485
								if buffer[position] != rune('C') {
									goto l478
								}
								position++
							}
						l485:
							{
								position487, tokenIndex487 := position, tokenIndex
								if buffer[position] != rune('h') {
									goto l488
								}
								position++
								goto l487
							l488:
								position, tokenIndex = position487, tokenIndex487
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
								position489, tokenIndex489 := position, tokenIndex
								if !_rules[ruleliteralString]() {
									goto l490
								}
								goto l489
							l490:
								position, tokenIndex = position489, tokenIndex489
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
							position, tokenIndex = position468, tokenIndex468
							if !_rules[rule_]() {
								goto l492
							}
							{
								position493, tokenIndex493 := position, tokenIndex
								if buffer[position] != rune('i') {
									goto l494
								}
								position++
								goto l493
							l494:
								position, tokenIndex = position493, tokenIndex493
								if buffer[position] != rune('I') {
									goto l492
								}
								position++
							}
						l493:
							{
								position495, tokenIndex495 := position, tokenIndex
								if buffer[position] != rune('n') {
									goto l496
								}
								position++
								goto l495
							l496:
								position, tokenIndex = position495, tokenIndex495
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
								position497, tokenIndex497 := position, tokenIndex
								{
									position499 := position
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
										position501, tokenIndex501 := position, tokenIndex
										if !_rules[ruleliteralListString]() {
											goto l502
										}
										goto l501
									l502:
										position, tokenIndex = position501, tokenIndex501
										if !(p.errorHere(position, `expected string literal to follow "(" in literal list`)) {
											goto l498
										}
									}
								l501:
								l503:
									{
										position504, tokenIndex504 := position, tokenIndex
										if !_rules[rule_]() {
											goto l504
										}
										if !_rules[ruleCOMMA]() {
											goto l504
										}
										{
											position505, tokenIndex505 := position, tokenIndex
											if !_rules[ruleliteralListString]() {
												goto l506
											}
											goto l505
										l506:
											position, tokenIndex = position505, tokenIndex505
											if !(p.errorHere(position, `expected string literal to follow "," in literal list`)) {
												goto l504
											}
										}
									l505:
										goto l503
									l504:
										position, tokenIndex = position504, tokenIndex504
									}
									{
										position507, tokenIndex507 := position, tokenIndex
										if !_rules[rule_]() {
											goto l508
										}
										if !_rules[rulePAREN_CLOSE]() {
											goto l508
										}
										goto l507
									l508:
										position, tokenIndex = position507, tokenIndex507
										if !(p.errorHere(position, `expected ")" to close "(" for literal list`)) {
											goto l498
										}
									}
								l507:
									add(ruleliteralList, position499)
								}
								goto l497
							l498:
								position, tokenIndex = position497, tokenIndex497
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
							position, tokenIndex = position468, tokenIndex468
							if !(p.errorHere(position, `expected "=", "!=", "match", or "in" to follow tag key in predicate`)) {
								goto l448
							}
						}
					l468:
						add(ruletagMatcher, position467)
					}
				}
			l450:
				add(rulepredicate_3, position449)
			}
			return true
		l448:
			position, tokenIndex = position448, tokenIndex448
			return false
		},
		/* 29 tagMatcher <- <(tagName ((_ '=' (literalString / &{ p.errorHere(position, `expected string literal to follow "="`) }) Action45) / (_ ('!' '=') (literalString / &{ p.errorHere(position, `expected string literal to follow "!="`) }) Action46 Action47) / (_ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H')) KEY (literalString / &{ p.errorHere(position, `expected regex string literal to follow "match"`) }) Action48) / (_ (('i' / 'I') ('n' / 'N')) KEY (literalList / &{ p.errorHere(position, `expected string literal list to follow "in" keyword`) }) Action49) / &{ p.errorHere(position, `expected "=", "!=", "match", or "in" to follow tag key in predicate`) }))> */
		nil,
		/* 30 literalString <- <(_ STRING Action50)> */
		func() bool {
			position511, tokenIndex511 := position, tokenIndex
			{
				position512 := position
				if !_rules[rule_]() {
					goto l511
				}
				if !_rules[ruleSTRING]() {
					goto l511
				}
				{
					add(ruleAction50, position)
				}
				add(ruleliteralString, position512)
			}
			return true
		l511:
			position, tokenIndex = position511, tokenIndex511
			return false
		},
		/* 31 literalList <- <(Action51 _ PAREN_OPEN (literalListString / &{ p.errorHere(position, `expected string literal to follow "(" in literal list`) }) (_ COMMA (literalListString / &{ p.errorHere(position, `expected string literal to follow "," in literal list`) }))* ((_ PAREN_CLOSE) / &{ p.errorHere(position, `expected ")" to close "(" for literal list`) }))> */
		nil,
		/* 32 literalListString <- <(_ STRING Action52)> */
		func() bool {
			position515, tokenIndex515 := position, tokenIndex
			{
				position516 := position
				if !_rules[rule_]() {
					goto l515
				}
				if !_rules[ruleSTRING]() {
					goto l515
				}
				{
					add(ruleAction52, position)
				}
				add(ruleliteralListString, position516)
			}
			return true
		l515:
			position, tokenIndex = position515, tokenIndex515
			return false
		},
		/* 33 tagName <- <(_ <TAG_NAME> Action53)> */
		func() bool {
			position518, tokenIndex518 := position, tokenIndex
			{
				position519 := position
				if !_rules[rule_]() {
					goto l518
				}
				{
					position520 := position
					{
						position521 := position
						if !_rules[ruleIDENTIFIER]() {
							goto l518
						}
						add(ruleTAG_NAME, position521)
					}
					add(rulePegText, position520)
				}
				{
					add(ruleAction53, position)
				}
				add(ruletagName, position519)
			}
			return true
		l518:
			position, tokenIndex = position518, tokenIndex518
			return false
		},
		/* 34 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position523, tokenIndex523 := position, tokenIndex
			{
				position524 := position
				if !_rules[ruleIDENTIFIER]() {
					goto l523
				}
				add(ruleCOLUMN_NAME, position524)
			}
			return true
		l523:
			position, tokenIndex = position523, tokenIndex523
			return false
		},
		/* 35 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 36 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 37 IDENTIFIER <- <(('`' CHAR* ('`' / &{ p.errorHere(position, "expected \"`\" to end identifier") })) / (!(KEYWORD KEY) ID_SEGMENT ('.' (ID_SEGMENT / &{ p.errorHere(position, `expected identifier segment to follow "."`) }))*))> */
		func() bool {
			position527, tokenIndex527 := position, tokenIndex
			{
				position528 := position
				{
					position529, tokenIndex529 := position, tokenIndex
					if buffer[position] != rune('`') {
						goto l530
					}
					position++
				l531:
					{
						position532, tokenIndex532 := position, tokenIndex
						if !_rules[ruleCHAR]() {
							goto l532
						}
						goto l531
					l532:
						position, tokenIndex = position532, tokenIndex532
					}
					{
						position533, tokenIndex533 := position, tokenIndex
						if buffer[position] != rune('`') {
							goto l534
						}
						position++
						goto l533
					l534:
						position, tokenIndex = position533, tokenIndex533
						if !(p.errorHere(position, "expected \"`\" to end identifier")) {
							goto l530
						}
					}
				l533:
					goto l529
				l530:
					position, tokenIndex = position529, tokenIndex529
					{
						position535, tokenIndex535 := position, tokenIndex
						{
							position536 := position
							{
								position537, tokenIndex537 := position, tokenIndex
								{
									position539, tokenIndex539 := position, tokenIndex
									if buffer[position] != rune('a') {
										goto l540
									}
									position++
									goto l539
								l540:
									position, tokenIndex = position539, tokenIndex539
									if buffer[position] != rune('A') {
										goto l538
									}
									position++
								}
							l539:
								{
									position541, tokenIndex541 := position, tokenIndex
									if buffer[position] != rune('l') {
										goto l542
									}
									position++
									goto l541
								l542:
									position, tokenIndex = position541, tokenIndex541
									if buffer[position] != rune('L') {
										goto l538
									}
									position++
								}
							l541:
								{
									position543, tokenIndex543 := position, tokenIndex
									if buffer[position] != rune('l') {
										goto l544
									}
									position++
									goto l543
								l544:
									position, tokenIndex = position543, tokenIndex543
									if buffer[position] != rune('L') {
										goto l538
									}
									position++
								}
							l543:
								goto l537
							l538:
								position, tokenIndex = position537, tokenIndex537
								{
									position546, tokenIndex546 := position, tokenIndex
									if buffer[position] != rune('a') {
										goto l547
									}
									position++
									goto l546
								l547:
									position, tokenIndex = position546, tokenIndex546
									if buffer[position] != rune('A') {
										goto l545
									}
									position++
								}
							l546:
								{
									position548, tokenIndex548 := position, tokenIndex
									if buffer[position] != rune('n') {
										goto l549
									}
									position++
									goto l548
								l549:
									position, tokenIndex = position548, tokenIndex548
									if buffer[position] != rune('N') {
										goto l545
									}
									position++
								}
							l548:
								{
									position550, tokenIndex550 := position, tokenIndex
									if buffer[position] != rune('d') {
										goto l551
									}
									position++
									goto l550
								l551:
									position, tokenIndex = position550, tokenIndex550
									if buffer[position] != rune('D') {
										goto l545
									}
									position++
								}
							l550:
								goto l537
							l545:
								position, tokenIndex = position537, tokenIndex537
								{
									position553, tokenIndex553 := position, tokenIndex
									if buffer[position] != rune('m') {
										goto l554
									}
									position++
									goto l553
								l554:
									position, tokenIndex = position553, tokenIndex553
									if buffer[position] != rune('M') {
										goto l552
									}
									position++
								}
							l553:
								{
									position555, tokenIndex555 := position, tokenIndex
									if buffer[position] != rune('a') {
										goto l556
									}
									position++
									goto l555
								l556:
									position, tokenIndex = position555, tokenIndex555
									if buffer[position] != rune('A') {
										goto l552
									}
									position++
								}
							l555:
								{
									position557, tokenIndex557 := position, tokenIndex
									if buffer[position] != rune('t') {
										goto l558
									}
									position++
									goto l557
								l558:
									position, tokenIndex = position557, tokenIndex557
									if buffer[position] != rune('T') {
										goto l552
									}
									position++
								}
							l557:
								{
									position559, tokenIndex559 := position, tokenIndex
									if buffer[position] != rune('c') {
										goto l560
									}
									position++
									goto l559
								l560:
									position, tokenIndex = position559, tokenIndex559
									if buffer[position] != rune('C') {
										goto l552
									}
									position++
								}
							l559:
								{
									position561, tokenIndex561 := position, tokenIndex
									if buffer[position] != rune('h') {
										goto l562
									}
									position++
									goto l561
								l562:
									position, tokenIndex = position561, tokenIndex561
									if buffer[position] != rune('H') {
										goto l552
									}
									position++
								}
							l561:
								goto l537
							l552:
								position, tokenIndex = position537, tokenIndex537
								{
									position564, tokenIndex564 := position, tokenIndex
									if buffer[position] != rune('s') {
										goto l565
									}
									position++
									goto l564
								l565:
									position, tokenIndex = position564, tokenIndex564
									if buffer[position] != rune('S') {
										goto l563
									}
									position++
								}
							l564:
								{
									position566, tokenIndex566 := position, tokenIndex
									if buffer[position] != rune('e') {
										goto l567
									}
									position++
									goto l566
								l567:
									position, tokenIndex = position566, tokenIndex566
									if buffer[position] != rune('E') {
										goto l563
									}
									position++
								}
							l566:
								{
									position568, tokenIndex568 := position, tokenIndex
									if buffer[position] != rune('l') {
										goto l569
									}
									position++
									goto l568
								l569:
									position, tokenIndex = position568, tokenIndex568
									if buffer[position] != rune('L') {
										goto l563
									}
									position++
								}
							l568:
								{
									position570, tokenIndex570 := position, tokenIndex
									if buffer[position] != rune('e') {
										goto l571
									}
									position++
									goto l570
								l571:
									position, tokenIndex = position570, tokenIndex570
									if buffer[position] != rune('E') {
										goto l563
									}
									position++
								}
							l570:
								{
									position572, tokenIndex572 := position, tokenIndex
									if buffer[position] != rune('c') {
										goto l573
									}
									position++
									goto l572
								l573:
									position, tokenIndex = position572, tokenIndex572
									if buffer[position] != rune('C') {
										goto l563
									}
									position++
								}
							l572:
								{
									position574, tokenIndex574 := position, tokenIndex
									if buffer[position] != rune('t') {
										goto l575
									}
									position++
									goto l574
								l575:
									position, tokenIndex = position574, tokenIndex574
									if buffer[position] != rune('T') {
										goto l563
									}
									position++
								}
							l574:
								goto l537
							l563:
								position, tokenIndex = position537, tokenIndex537
								{
									switch buffer[position] {
									case 'S', 's':
										{
											position577, tokenIndex577 := position, tokenIndex
											if buffer[position] != rune('s') {
												goto l578
											}
											position++
											goto l577
										l578:
											position, tokenIndex = position577, tokenIndex577
											if buffer[position] != rune('S') {
												goto l535
											}
											position++
										}
									l577:
										{
											position579, tokenIndex579 := position, tokenIndex
											if buffer[position] != rune('a') {
												goto l580
											}
											position++
											goto l579
										l580:
											position, tokenIndex = position579, tokenIndex579
											if buffer[position] != rune('A') {
												goto l535
											}
											position++
										}
									l579:
										{
											position581, tokenIndex581 := position, tokenIndex
											if buffer[position] != rune('m') {
												goto l582
											}
											position++
											goto l581
										l582:
											position, tokenIndex = position581, tokenIndex581
											if buffer[position] != rune('M') {
												goto l535
											}
											position++
										}
									l581:
										{
											position583, tokenIndex583 := position, tokenIndex
											if buffer[position] != rune('p') {
												goto l584
											}
											position++
											goto l583
										l584:
											position, tokenIndex = position583, tokenIndex583
											if buffer[position] != rune('P') {
												goto l535
											}
											position++
										}
									l583:
										{
											position585, tokenIndex585 := position, tokenIndex
											if buffer[position] != rune('l') {
												goto l586
											}
											position++
											goto l585
										l586:
											position, tokenIndex = position585, tokenIndex585
											if buffer[position] != rune('L') {
												goto l535
											}
											position++
										}
									l585:
										{
											position587, tokenIndex587 := position, tokenIndex
											if buffer[position] != rune('e') {
												goto l588
											}
											position++
											goto l587
										l588:
											position, tokenIndex = position587, tokenIndex587
											if buffer[position] != rune('E') {
												goto l535
											}
											position++
										}
									l587:
										break
									case 'R', 'r':
										{
											position589, tokenIndex589 := position, tokenIndex
											if buffer[position] != rune('r') {
												goto l590
											}
											position++
											goto l589
										l590:
											position, tokenIndex = position589, tokenIndex589
											if buffer[position] != rune('R') {
												goto l535
											}
											position++
										}
									l589:
										{
											position591, tokenIndex591 := position, tokenIndex
											if buffer[position] != rune('e') {
												goto l592
											}
											position++
											goto l591
										l592:
											position, tokenIndex = position591, tokenIndex591
											if buffer[position] != rune('E') {
												goto l535
											}
											position++
										}
									l591:
										{
											position593, tokenIndex593 := position, tokenIndex
											if buffer[position] != rune('s') {
												goto l594
											}
											position++
											goto l593
										l594:
											position, tokenIndex = position593, tokenIndex593
											if buffer[position] != rune('S') {
												goto l535
											}
											position++
										}
									l593:
										{
											position595, tokenIndex595 := position, tokenIndex
											if buffer[position] != rune('o') {
												goto l596
											}
											position++
											goto l595
										l596:
											position, tokenIndex = position595, tokenIndex595
											if buffer[position] != rune('O') {
												goto l535
											}
											position++
										}
									l595:
										{
											position597, tokenIndex597 := position, tokenIndex
											if buffer[position] != rune('l') {
												goto l598
											}
											position++
											goto l597
										l598:
											position, tokenIndex = position597, tokenIndex597
											if buffer[position] != rune('L') {
												goto l535
											}
											position++
										}
									l597:
										{
											position599, tokenIndex599 := position, tokenIndex
											if buffer[position] != rune('u') {
												goto l600
											}
											position++
											goto l599
										l600:
											position, tokenIndex = position599, tokenIndex599
											if buffer[position] != rune('U') {
												goto l535
											}
											position++
										}
									l599:
										{
											position601, tokenIndex601 := position, tokenIndex
											if buffer[position] != rune('t') {
												goto l602
											}
											position++
											goto l601
										l602:
											position, tokenIndex = position601, tokenIndex601
											if buffer[position] != rune('T') {
												goto l535
											}
											position++
										}
									l601:
										{
											position603, tokenIndex603 := position, tokenIndex
											if buffer[position] != rune('i') {
												goto l604
											}
											position++
											goto l603
										l604:
											position, tokenIndex = position603, tokenIndex603
											if buffer[position] != rune('I') {
												goto l535
											}
											position++
										}
									l603:
										{
											position605, tokenIndex605 := position, tokenIndex
											if buffer[position] != rune('o') {
												goto l606
											}
											position++
											goto l605
										l606:
											position, tokenIndex = position605, tokenIndex605
											if buffer[position] != rune('O') {
												goto l535
											}
											position++
										}
									l605:
										{
											position607, tokenIndex607 := position, tokenIndex
											if buffer[position] != rune('n') {
												goto l608
											}
											position++
											goto l607
										l608:
											position, tokenIndex = position607, tokenIndex607
											if buffer[position] != rune('N') {
												goto l535
											}
											position++
										}
									l607:
										break
									case 'T', 't':
										{
											position609, tokenIndex609 := position, tokenIndex
											if buffer[position] != rune('t') {
												goto l610
											}
											position++
											goto l609
										l610:
											position, tokenIndex = position609, tokenIndex609
											if buffer[position] != rune('T') {
												goto l535
											}
											position++
										}
									l609:
										{
											position611, tokenIndex611 := position, tokenIndex
											if buffer[position] != rune('o') {
												goto l612
											}
											position++
											goto l611
										l612:
											position, tokenIndex = position611, tokenIndex611
											if buffer[position] != rune('O') {
												goto l535
											}
											position++
										}
									l611:
										break
									case 'F', 'f':
										{
											position613, tokenIndex613 := position, tokenIndex
											if buffer[position] != rune('f') {
												goto l614
											}
											position++
											goto l613
										l614:
											position, tokenIndex = position613, tokenIndex613
											if buffer[position] != rune('F') {
												goto l535
											}
											position++
										}
									l613:
										{
											position615, tokenIndex615 := position, tokenIndex
											if buffer[position] != rune('r') {
												goto l616
											}
											position++
											goto l615
										l616:
											position, tokenIndex = position615, tokenIndex615
											if buffer[position] != rune('R') {
												goto l535
											}
											position++
										}
									l615:
										{
											position617, tokenIndex617 := position, tokenIndex
											if buffer[position] != rune('o') {
												goto l618
											}
											position++
											goto l617
										l618:
											position, tokenIndex = position617, tokenIndex617
											if buffer[position] != rune('O') {
												goto l535
											}
											position++
										}
									l617:
										{
											position619, tokenIndex619 := position, tokenIndex
											if buffer[position] != rune('m') {
												goto l620
											}
											position++
											goto l619
										l620:
											position, tokenIndex = position619, tokenIndex619
											if buffer[position] != rune('M') {
												goto l535
											}
											position++
										}
									l619:
										break
									case 'M', 'm':
										{
											position621, tokenIndex621 := position, tokenIndex
											if buffer[position] != rune('m') {
												goto l622
											}
											position++
											goto l621
										l622:
											position, tokenIndex = position621, tokenIndex621
											if buffer[position] != rune('M') {
												goto l535
											}
											position++
										}
									l621:
										{
											position623, tokenIndex623 := position, tokenIndex
											if buffer[position] != rune('e') {
												goto l624
											}
											position++
											goto l623
										l624:
											position, tokenIndex = position623, tokenIndex623
											if buffer[position] != rune('E') {
												goto l535
											}
											position++
										}
									l623:
										{
											position625, tokenIndex625 := position, tokenIndex
											if buffer[position] != rune('t') {
												goto l626
											}
											position++
											goto l625
										l626:
											position, tokenIndex = position625, tokenIndex625
											if buffer[position] != rune('T') {
												goto l535
											}
											position++
										}
									l625:
										{
											position627, tokenIndex627 := position, tokenIndex
											if buffer[position] != rune('r') {
												goto l628
											}
											position++
											goto l627
										l628:
											position, tokenIndex = position627, tokenIndex627
											if buffer[position] != rune('R') {
												goto l535
											}
											position++
										}
									l627:
										{
											position629, tokenIndex629 := position, tokenIndex
											if buffer[position] != rune('i') {
												goto l630
											}
											position++
											goto l629
										l630:
											position, tokenIndex = position629, tokenIndex629
											if buffer[position] != rune('I') {
												goto l535
											}
											position++
										}
									l629:
										{
											position631, tokenIndex631 := position, tokenIndex
											if buffer[position] != rune('c') {
												goto l632
											}
											position++
											goto l631
										l632:
											position, tokenIndex = position631, tokenIndex631
											if buffer[position] != rune('C') {
												goto l535
											}
											position++
										}
									l631:
										{
											position633, tokenIndex633 := position, tokenIndex
											if buffer[position] != rune('s') {
												goto l634
											}
											position++
											goto l633
										l634:
											position, tokenIndex = position633, tokenIndex633
											if buffer[position] != rune('S') {
												goto l535
											}
											position++
										}
									l633:
										break
									case 'W', 'w':
										{
											position635, tokenIndex635 := position, tokenIndex
											if buffer[position] != rune('w') {
												goto l636
											}
											position++
											goto l635
										l636:
											position, tokenIndex = position635, tokenIndex635
											if buffer[position] != rune('W') {
												goto l535
											}
											position++
										}
									l635:
										{
											position637, tokenIndex637 := position, tokenIndex
											if buffer[position] != rune('h') {
												goto l638
											}
											position++
											goto l637
										l638:
											position, tokenIndex = position637, tokenIndex637
											if buffer[position] != rune('H') {
												goto l535
											}
											position++
										}
									l637:
										{
											position639, tokenIndex639 := position, tokenIndex
											if buffer[position] != rune('e') {
												goto l640
											}
											position++
											goto l639
										l640:
											position, tokenIndex = position639, tokenIndex639
											if buffer[position] != rune('E') {
												goto l535
											}
											position++
										}
									l639:
										{
											position641, tokenIndex641 := position, tokenIndex
											if buffer[position] != rune('r') {
												goto l642
											}
											position++
											goto l641
										l642:
											position, tokenIndex = position641, tokenIndex641
											if buffer[position] != rune('R') {
												goto l535
											}
											position++
										}
									l641:
										{
											position643, tokenIndex643 := position, tokenIndex
											if buffer[position] != rune('e') {
												goto l644
											}
											position++
											goto l643
										l644:
											position, tokenIndex = position643, tokenIndex643
											if buffer[position] != rune('E') {
												goto l535
											}
											position++
										}
									l643:
										break
									case 'O', 'o':
										{
											position645, tokenIndex645 := position, tokenIndex
											if buffer[position] != rune('o') {
												goto l646
											}
											position++
											goto l645
										l646:
											position, tokenIndex = position645, tokenIndex645
											if buffer[position] != rune('O') {
												goto l535
											}
											position++
										}
									l645:
										{
											position647, tokenIndex647 := position, tokenIndex
											if buffer[position] != rune('r') {
												goto l648
											}
											position++
											goto l647
										l648:
											position, tokenIndex = position647, tokenIndex647
											if buffer[position] != rune('R') {
												goto l535
											}
											position++
										}
									l647:
										break
									case 'N', 'n':
										{
											position649, tokenIndex649 := position, tokenIndex
											if buffer[position] != rune('n') {
												goto l650
											}
											position++
											goto l649
										l650:
											position, tokenIndex = position649, tokenIndex649
											if buffer[position] != rune('N') {
												goto l535
											}
											position++
										}
									l649:
										{
											position651, tokenIndex651 := position, tokenIndex
											if buffer[position] != rune('o') {
												goto l652
											}
											position++
											goto l651
										l652:
											position, tokenIndex = position651, tokenIndex651
											if buffer[position] != rune('O') {
												goto l535
											}
											position++
										}
									l651:
										{
											position653, tokenIndex653 := position, tokenIndex
											if buffer[position] != rune('t') {
												goto l654
											}
											position++
											goto l653
										l654:
											position, tokenIndex = position653, tokenIndex653
											if buffer[position] != rune('T') {
												goto l535
											}
											position++
										}
									l653:
										break
									case 'I', 'i':
										{
											position655, tokenIndex655 := position, tokenIndex
											if buffer[position] != rune('i') {
												goto l656
											}
											position++
											goto l655
										l656:
											position, tokenIndex = position655, tokenIndex655
											if buffer[position] != rune('I') {
												goto l535
											}
											position++
										}
									l655:
										{
											position657, tokenIndex657 := position, tokenIndex
											if buffer[position] != rune('n') {
												goto l658
											}
											position++
											goto l657
										l658:
											position, tokenIndex = position657, tokenIndex657
											if buffer[position] != rune('N') {
												goto l535
											}
											position++
										}
									l657:
										break
									case 'C', 'c':
										{
											position659, tokenIndex659 := position, tokenIndex
											if buffer[position] != rune('c') {
												goto l660
											}
											position++
											goto l659
										l660:
											position, tokenIndex = position659, tokenIndex659
											if buffer[position] != rune('C') {
												goto l535
											}
											position++
										}
									l659:
										{
											position661, tokenIndex661 := position, tokenIndex
											if buffer[position] != rune('o') {
												goto l662
											}
											position++
											goto l661
										l662:
											position, tokenIndex = position661, tokenIndex661
											if buffer[position] != rune('O') {
												goto l535
											}
											position++
										}
									l661:
										{
											position663, tokenIndex663 := position, tokenIndex
											if buffer[position] != rune('l') {
												goto l664
											}
											position++
											goto l663
										l664:
											position, tokenIndex = position663, tokenIndex663
											if buffer[position] != rune('L') {
												goto l535
											}
											position++
										}
									l663:
										{
											position665, tokenIndex665 := position, tokenIndex
											if buffer[position] != rune('l') {
												goto l666
											}
											position++
											goto l665
										l666:
											position, tokenIndex = position665, tokenIndex665
											if buffer[position] != rune('L') {
												goto l535
											}
											position++
										}
									l665:
										{
											position667, tokenIndex667 := position, tokenIndex
											if buffer[position] != rune('a') {
												goto l668
											}
											position++
											goto l667
										l668:
											position, tokenIndex = position667, tokenIndex667
											if buffer[position] != rune('A') {
												goto l535
											}
											position++
										}
									l667:
										{
											position669, tokenIndex669 := position, tokenIndex
											if buffer[position] != rune('p') {
												goto l670
											}
											position++
											goto l669
										l670:
											position, tokenIndex = position669, tokenIndex669
											if buffer[position] != rune('P') {
												goto l535
											}
											position++
										}
									l669:
										{
											position671, tokenIndex671 := position, tokenIndex
											if buffer[position] != rune('s') {
												goto l672
											}
											position++
											goto l671
										l672:
											position, tokenIndex = position671, tokenIndex671
											if buffer[position] != rune('S') {
												goto l535
											}
											position++
										}
									l671:
										{
											position673, tokenIndex673 := position, tokenIndex
											if buffer[position] != rune('e') {
												goto l674
											}
											position++
											goto l673
										l674:
											position, tokenIndex = position673, tokenIndex673
											if buffer[position] != rune('E') {
												goto l535
											}
											position++
										}
									l673:
										break
									case 'G', 'g':
										{
											position675, tokenIndex675 := position, tokenIndex
											if buffer[position] != rune('g') {
												goto l676
											}
											position++
											goto l675
										l676:
											position, tokenIndex = position675, tokenIndex675
											if buffer[position] != rune('G') {
												goto l535
											}
											position++
										}
									l675:
										{
											position677, tokenIndex677 := position, tokenIndex
											if buffer[position] != rune('r') {
												goto l678
											}
											position++
											goto l677
										l678:
											position, tokenIndex = position677, tokenIndex677
											if buffer[position] != rune('R') {
												goto l535
											}
											position++
										}
									l677:
										{
											position679, tokenIndex679 := position, tokenIndex
											if buffer[position] != rune('o') {
												goto l680
											}
											position++
											goto l679
										l680:
											position, tokenIndex = position679, tokenIndex679
											if buffer[position] != rune('O') {
												goto l535
											}
											position++
										}
									l679:
										{
											position681, tokenIndex681 := position, tokenIndex
											if buffer[position] != rune('u') {
												goto l682
											}
											position++
											goto l681
										l682:
											position, tokenIndex = position681, tokenIndex681
											if buffer[position] != rune('U') {
												goto l535
											}
											position++
										}
									l681:
										{
											position683, tokenIndex683 := position, tokenIndex
											if buffer[position] != rune('p') {
												goto l684
											}
											position++
											goto l683
										l684:
											position, tokenIndex = position683, tokenIndex683
											if buffer[position] != rune('P') {
												goto l535
											}
											position++
										}
									l683:
										break
									case 'D', 'd':
										{
											position685, tokenIndex685 := position, tokenIndex
											if buffer[position] != rune('d') {
												goto l686
											}
											position++
											goto l685
										l686:
											position, tokenIndex = position685, tokenIndex685
											if buffer[position] != rune('D') {
												goto l535
											}
											position++
										}
									l685:
										{
											position687, tokenIndex687 := position, tokenIndex
											if buffer[position] != rune('e') {
												goto l688
											}
											position++
											goto l687
										l688:
											position, tokenIndex = position687, tokenIndex687
											if buffer[position] != rune('E') {
												goto l535
											}
											position++
										}
									l687:
										{
											position689, tokenIndex689 := position, tokenIndex
											if buffer[position] != rune('s') {
												goto l690
											}
											position++
											goto l689
										l690:
											position, tokenIndex = position689, tokenIndex689
											if buffer[position] != rune('S') {
												goto l535
											}
											position++
										}
									l689:
										{
											position691, tokenIndex691 := position, tokenIndex
											if buffer[position] != rune('c') {
												goto l692
											}
											position++
											goto l691
										l692:
											position, tokenIndex = position691, tokenIndex691
											if buffer[position] != rune('C') {
												goto l535
											}
											position++
										}
									l691:
										{
											position693, tokenIndex693 := position, tokenIndex
											if buffer[position] != rune('r') {
												goto l694
											}
											position++
											goto l693
										l694:
											position, tokenIndex = position693, tokenIndex693
											if buffer[position] != rune('R') {
												goto l535
											}
											position++
										}
									l693:
										{
											position695, tokenIndex695 := position, tokenIndex
											if buffer[position] != rune('i') {
												goto l696
											}
											position++
											goto l695
										l696:
											position, tokenIndex = position695, tokenIndex695
											if buffer[position] != rune('I') {
												goto l535
											}
											position++
										}
									l695:
										{
											position697, tokenIndex697 := position, tokenIndex
											if buffer[position] != rune('b') {
												goto l698
											}
											position++
											goto l697
										l698:
											position, tokenIndex = position697, tokenIndex697
											if buffer[position] != rune('B') {
												goto l535
											}
											position++
										}
									l697:
										{
											position699, tokenIndex699 := position, tokenIndex
											if buffer[position] != rune('e') {
												goto l700
											}
											position++
											goto l699
										l700:
											position, tokenIndex = position699, tokenIndex699
											if buffer[position] != rune('E') {
												goto l535
											}
											position++
										}
									l699:
										break
									case 'B', 'b':
										{
											position701, tokenIndex701 := position, tokenIndex
											if buffer[position] != rune('b') {
												goto l702
											}
											position++
											goto l701
										l702:
											position, tokenIndex = position701, tokenIndex701
											if buffer[position] != rune('B') {
												goto l535
											}
											position++
										}
									l701:
										{
											position703, tokenIndex703 := position, tokenIndex
											if buffer[position] != rune('y') {
												goto l704
											}
											position++
											goto l703
										l704:
											position, tokenIndex = position703, tokenIndex703
											if buffer[position] != rune('Y') {
												goto l535
											}
											position++
										}
									l703:
										break
									default:
										{
											position705, tokenIndex705 := position, tokenIndex
											if buffer[position] != rune('a') {
												goto l706
											}
											position++
											goto l705
										l706:
											position, tokenIndex = position705, tokenIndex705
											if buffer[position] != rune('A') {
												goto l535
											}
											position++
										}
									l705:
										{
											position707, tokenIndex707 := position, tokenIndex
											if buffer[position] != rune('s') {
												goto l708
											}
											position++
											goto l707
										l708:
											position, tokenIndex = position707, tokenIndex707
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
							add(ruleKEYWORD, position536)
						}
						if !_rules[ruleKEY]() {
							goto l535
						}
						goto l527
					l535:
						position, tokenIndex = position535, tokenIndex535
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l527
					}
				l709:
					{
						position710, tokenIndex710 := position, tokenIndex
						if buffer[position] != rune('.') {
							goto l710
						}
						position++
						{
							position711, tokenIndex711 := position, tokenIndex
							if !_rules[ruleID_SEGMENT]() {
								goto l712
							}
							goto l711
						l712:
							position, tokenIndex = position711, tokenIndex711
							if !(p.errorHere(position, `expected identifier segment to follow "."`)) {
								goto l710
							}
						}
					l711:
						goto l709
					l710:
						position, tokenIndex = position710, tokenIndex710
					}
				}
			l529:
				add(ruleIDENTIFIER, position528)
			}
			return true
		l527:
			position, tokenIndex = position527, tokenIndex527
			return false
		},
		/* 38 TIMESTAMP <- <((_ <(NUMBER ([a-z] / [A-Z])*)>) / (_ STRING) / (_ <(('n' / 'N') ('o' / 'O') ('w' / 'W'))> KEY))> */
		nil,
		/* 39 ID_SEGMENT <- <(ID_START ID_CONT*)> */
		func() bool {
			position714, tokenIndex714 := position, tokenIndex
			{
				position715 := position
				if !_rules[ruleID_START]() {
					goto l714
				}
			l716:
				{
					position717, tokenIndex717 := position, tokenIndex
					if !_rules[ruleID_CONT]() {
						goto l717
					}
					goto l716
				l717:
					position, tokenIndex = position717, tokenIndex717
				}
				add(ruleID_SEGMENT, position715)
			}
			return true
		l714:
			position, tokenIndex = position714, tokenIndex714
			return false
		},
		/* 40 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position718, tokenIndex718 := position, tokenIndex
			{
				position719 := position
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

				add(ruleID_START, position719)
			}
			return true
		l718:
			position, tokenIndex = position718, tokenIndex718
			return false
		},
		/* 41 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position721, tokenIndex721 := position, tokenIndex
			{
				position722 := position
				{
					position723, tokenIndex723 := position, tokenIndex
					if !_rules[ruleID_START]() {
						goto l724
					}
					goto l723
				l724:
					position, tokenIndex = position723, tokenIndex723
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l721
					}
					position++
				}
			l723:
				add(ruleID_CONT, position722)
			}
			return true
		l721:
			position, tokenIndex = position721, tokenIndex721
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
			position736, tokenIndex736 := position, tokenIndex
			{
				position737 := position
				if buffer[position] != rune('\'') {
					goto l736
				}
				position++
				add(ruleQUOTE_SINGLE, position737)
			}
			return true
		l736:
			position, tokenIndex = position736, tokenIndex736
			return false
		},
		/* 54 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position738, tokenIndex738 := position, tokenIndex
			{
				position739 := position
				if buffer[position] != rune('"') {
					goto l738
				}
				position++
				add(ruleQUOTE_DOUBLE, position739)
			}
			return true
		l738:
			position, tokenIndex = position738, tokenIndex738
			return false
		},
		/* 55 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> (QUOTE_SINGLE / &{ p.errorHere(position, `expected "'" to close string`) })) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> (QUOTE_DOUBLE / &{ p.errorHere(position, `expected '"' to close string`) })))> */
		func() bool {
			position740, tokenIndex740 := position, tokenIndex
			{
				position741 := position
				{
					position742, tokenIndex742 := position, tokenIndex
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l743
					}
					{
						position744 := position
					l745:
						{
							position746, tokenIndex746 := position, tokenIndex
							{
								position747, tokenIndex747 := position, tokenIndex
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l747
								}
								goto l746
							l747:
								position, tokenIndex = position747, tokenIndex747
							}
							if !_rules[ruleCHAR]() {
								goto l746
							}
							goto l745
						l746:
							position, tokenIndex = position746, tokenIndex746
						}
						add(rulePegText, position744)
					}
					{
						position748, tokenIndex748 := position, tokenIndex
						if !_rules[ruleQUOTE_SINGLE]() {
							goto l749
						}
						goto l748
					l749:
						position, tokenIndex = position748, tokenIndex748
						if !(p.errorHere(position, `expected "'" to close string`)) {
							goto l743
						}
					}
				l748:
					goto l742
				l743:
					position, tokenIndex = position742, tokenIndex742
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l740
					}
					{
						position750 := position
					l751:
						{
							position752, tokenIndex752 := position, tokenIndex
							{
								position753, tokenIndex753 := position, tokenIndex
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l753
								}
								goto l752
							l753:
								position, tokenIndex = position753, tokenIndex753
							}
							if !_rules[ruleCHAR]() {
								goto l752
							}
							goto l751
						l752:
							position, tokenIndex = position752, tokenIndex752
						}
						add(rulePegText, position750)
					}
					{
						position754, tokenIndex754 := position, tokenIndex
						if !_rules[ruleQUOTE_DOUBLE]() {
							goto l755
						}
						goto l754
					l755:
						position, tokenIndex = position754, tokenIndex754
						if !(p.errorHere(position, `expected '"' to close string`)) {
							goto l740
						}
					}
				l754:
				}
			l742:
				add(ruleSTRING, position741)
			}
			return true
		l740:
			position, tokenIndex = position740, tokenIndex740
			return false
		},
		/* 56 CHAR <- <(('\\' ((&('"') (QUOTE_DOUBLE / &{ p.errorHere(position, "expected \"\\\", \"'\", \"`\", or '\"' to follow \"\\\" in string literal") })) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position756, tokenIndex756 := position, tokenIndex
			{
				position757 := position
				{
					position758, tokenIndex758 := position, tokenIndex
					if buffer[position] != rune('\\') {
						goto l759
					}
					position++
					{
						switch buffer[position] {
						case '"':
							{
								position761, tokenIndex761 := position, tokenIndex
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l762
								}
								goto l761
							l762:
								position, tokenIndex = position761, tokenIndex761
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
					position, tokenIndex = position758, tokenIndex758
					{
						position763, tokenIndex763 := position, tokenIndex
						if !_rules[ruleESCAPE_CLASS]() {
							goto l763
						}
						goto l756
					l763:
						position, tokenIndex = position763, tokenIndex763
					}
					if !matchDot() {
						goto l756
					}
				}
			l758:
				add(ruleCHAR, position757)
			}
			return true
		l756:
			position, tokenIndex = position756, tokenIndex756
			return false
		},
		/* 57 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position764, tokenIndex764 := position, tokenIndex
			{
				position765 := position
				{
					position766, tokenIndex766 := position, tokenIndex
					if buffer[position] != rune('`') {
						goto l767
					}
					position++
					goto l766
				l767:
					position, tokenIndex = position766, tokenIndex766
					if buffer[position] != rune('\\') {
						goto l764
					}
					position++
				}
			l766:
				add(ruleESCAPE_CLASS, position765)
			}
			return true
		l764:
			position, tokenIndex = position764, tokenIndex764
			return false
		},
		/* 58 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position768, tokenIndex768 := position, tokenIndex
			{
				position769 := position
				{
					position770 := position
					{
						position771, tokenIndex771 := position, tokenIndex
						if buffer[position] != rune('-') {
							goto l771
						}
						position++
						goto l772
					l771:
						position, tokenIndex = position771, tokenIndex771
					}
				l772:
					{
						position773 := position
						{
							position774, tokenIndex774 := position, tokenIndex
							if buffer[position] != rune('0') {
								goto l775
							}
							position++
							goto l774
						l775:
							position, tokenIndex = position774, tokenIndex774
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l768
							}
							position++
						l776:
							{
								position777, tokenIndex777 := position, tokenIndex
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l777
								}
								position++
								goto l776
							l777:
								position, tokenIndex = position777, tokenIndex777
							}
						}
					l774:
						add(ruleNUMBER_NATURAL, position773)
					}
					add(ruleNUMBER_INTEGER, position770)
				}
				{
					position778, tokenIndex778 := position, tokenIndex
					{
						position780 := position
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
							position782, tokenIndex782 := position, tokenIndex
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l782
							}
							position++
							goto l781
						l782:
							position, tokenIndex = position782, tokenIndex782
						}
						add(ruleNUMBER_FRACTION, position780)
					}
					goto l779
				l778:
					position, tokenIndex = position778, tokenIndex778
				}
			l779:
				{
					position783, tokenIndex783 := position, tokenIndex
					{
						position785 := position
						{
							position786, tokenIndex786 := position, tokenIndex
							if buffer[position] != rune('e') {
								goto l787
							}
							position++
							goto l786
						l787:
							position, tokenIndex = position786, tokenIndex786
							if buffer[position] != rune('E') {
								goto l783
							}
							position++
						}
					l786:
						{
							position788, tokenIndex788 := position, tokenIndex
							{
								position790, tokenIndex790 := position, tokenIndex
								if buffer[position] != rune('+') {
									goto l791
								}
								position++
								goto l790
							l791:
								position, tokenIndex = position790, tokenIndex790
								if buffer[position] != rune('-') {
									goto l788
								}
								position++
							}
						l790:
							goto l789
						l788:
							position, tokenIndex = position788, tokenIndex788
						}
					l789:
						{
							position792, tokenIndex792 := position, tokenIndex
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l793
							}
							position++
						l794:
							{
								position795, tokenIndex795 := position, tokenIndex
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l795
								}
								position++
								goto l794
							l795:
								position, tokenIndex = position795, tokenIndex795
							}
							goto l792
						l793:
							position, tokenIndex = position792, tokenIndex792
							if !(p.errorHere(position, `expected exponent`)) {
								goto l783
							}
						}
					l792:
						add(ruleNUMBER_EXP, position785)
					}
					goto l784
				l783:
					position, tokenIndex = position783, tokenIndex783
				}
			l784:
				add(ruleNUMBER, position769)
			}
			return true
		l768:
			position, tokenIndex = position768, tokenIndex768
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
			position801, tokenIndex801 := position, tokenIndex
			{
				position802 := position
				if buffer[position] != rune('(') {
					goto l801
				}
				position++
				add(rulePAREN_OPEN, position802)
			}
			return true
		l801:
			position, tokenIndex = position801, tokenIndex801
			return false
		},
		/* 65 PAREN_CLOSE <- <')'> */
		func() bool {
			position803, tokenIndex803 := position, tokenIndex
			{
				position804 := position
				if buffer[position] != rune(')') {
					goto l803
				}
				position++
				add(rulePAREN_CLOSE, position804)
			}
			return true
		l803:
			position, tokenIndex = position803, tokenIndex803
			return false
		},
		/* 66 COMMA <- <','> */
		func() bool {
			position805, tokenIndex805 := position, tokenIndex
			{
				position806 := position
				if buffer[position] != rune(',') {
					goto l805
				}
				position++
				add(ruleCOMMA, position806)
			}
			return true
		l805:
			position, tokenIndex = position805, tokenIndex805
			return false
		},
		/* 67 _ <- <((&('/') COMMENT_BLOCK) | (&('-') COMMENT_TRAIL) | (&('\t' | '\n' | ' ') SPACE))*> */
		func() bool {
			{
				position808 := position
			l809:
				{
					position810, tokenIndex810 := position, tokenIndex
					{
						switch buffer[position] {
						case '/':
							{
								position812 := position
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
									position814, tokenIndex814 := position, tokenIndex
									{
										position815, tokenIndex815 := position, tokenIndex
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
										position, tokenIndex = position815, tokenIndex815
									}
									if !matchDot() {
										goto l814
									}
									goto l813
								l814:
									position, tokenIndex = position814, tokenIndex814
								}
								if buffer[position] != rune('*') {
									goto l810
								}
								position++
								if buffer[position] != rune('/') {
									goto l810
								}
								position++
								add(ruleCOMMENT_BLOCK, position812)
							}
							break
						case '-':
							{
								position816 := position
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
									position818, tokenIndex818 := position, tokenIndex
									{
										position819, tokenIndex819 := position, tokenIndex
										if buffer[position] != rune('\n') {
											goto l819
										}
										position++
										goto l818
									l819:
										position, tokenIndex = position819, tokenIndex819
									}
									if !matchDot() {
										goto l818
									}
									goto l817
								l818:
									position, tokenIndex = position818, tokenIndex818
								}
								add(ruleCOMMENT_TRAIL, position816)
							}
							break
						default:
							{
								position820 := position
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

								add(ruleSPACE, position820)
							}
							break
						}
					}

					goto l809
				l810:
					position, tokenIndex = position810, tokenIndex810
				}
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
			position824, tokenIndex824 := position, tokenIndex
			{
				position825 := position
				{
					position826, tokenIndex826 := position, tokenIndex
					if !_rules[ruleID_CONT]() {
						goto l826
					}
					goto l824
				l826:
					position, tokenIndex = position826, tokenIndex826
				}
				add(ruleKEY, position825)
			}
			return true
		l824:
			position, tokenIndex = position824, tokenIndex824
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
