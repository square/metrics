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
			// @@ rul3s[node.token32.pegRule] escapes to heap
			// @@ strconv.Quote(string(([]rune)(buffer)[node.token32.begin:node.token32.end])) escapes to heap
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
	// @@ can inline (*token32).isZero
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
	// @@ can inline (*token32).isParentOf
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: uint32(t.begin), end: uint32(t.end), next: uint32(t.next)}
	// @@ can inline (*token32).getToken32
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

// @@ rul3s[t.pegRule] escapes to heap
// @@ t.begin escapes to heap
// @@ t.end escapes to heap
// @@ t.next escapes to heap

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
	// @@ can inline (*tokens32).trim
}

// @@ (*tokens32).trim ignoring self-assignment to t.tree

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
	// @@ token.String() escapes to heap
}

func (t *tokens32) Order() [][]token32 {
	// @@ leaking param: t to result ~r0 level=1
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		// @@ make([]int32, 1, math.MaxInt16) escapes to heap
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
			// @@ (*tokens32).Order ignoring self-assignment to t.tree
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
		// @@ make([][]token32, len(depths)) escapes to heap
		// @@ make([][]token32, len(depths)) escapes to heap
		// @@ make([]token32, len(t.tree) + len(depths)) escapes to heap
		// @@ make([]token32, len(t.tree) + len(depths)) escapes to heap
		// @@ make([][]token32, len(depths)) escapes to heap
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
	// @@ leaking param: t
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		// @@ &node32 literal escapes to heap
		// @@ &element literal escapes to heap
		// @@ &node32 literal escapes to heap
		// @@ &element literal escapes to heap
		// @@ &node32 literal escapes to heap
		// @@ &element literal escapes to heap
		// @@ &node32 literal escapes to heap
		// @@ &element literal escapes to heap
		// @@ &node32 literal escapes to heap
		// @@ &element literal escapes to heap
		// @@ &node32 literal escapes to heap
		// @@ &element literal escapes to heap
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			// @@ &node32 literal escapes to heap
			// @@ &node32 literal escapes to heap
			// @@ &node32 literal escapes to heap
			// @@ &node32 literal escapes to heap
			// @@ &node32 literal escapes to heap
			// @@ &node32 literal escapes to heap
			// @@ &node32 literal escapes to heap
			// @@ &node32 literal escapes to heap
			// @@ &node32 literal escapes to heap
			// @@ &node32 literal escapes to heap
			// @@ &node32 literal escapes to heap
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	// @@ &element literal escapes to heap
	// @@ &element literal escapes to heap
	// @@ &element literal escapes to heap
	// @@ &element literal escapes to heap
	// @@ &element literal escapes to heap
	// @@ &element literal escapes to heap
	// @@ &element literal escapes to heap
	// @@ &element literal escapes to heap
	// @@ &element literal escapes to heap
	// @@ &element literal escapes to heap
	// @@ &element literal escapes to heap
	// @@ &element literal escapes to heap
	return stack.node
}

func (t *tokens32) PreOrder() (<-chan state32, [][]token32) {
	// @@ leaking param content: t
	// @@ leaking param: t to result ~r1 level=1
	s, ordered := make(chan state32, 6), t.Order()
	// @@ mark escaped content: t
	go func() {
		// @@ make(chan state32, 6) escapes to heap
		// @@ make(chan state32, 6) escapes to heap
		var states [8]state32
		// @@ func literal escapes to heap
		// @@ func literal escapes to heap
		for i := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		// @@ make([]int32, len(ordered)) escapes to heap
		// @@ make([]int32, len(ordered)) escapes to heap
		// @@ make([]int32, len(ordered)) escapes to heap
		// @@ make([]int32, len(ordered)) escapes to heap
		// @@ make([]int32, len(ordered)) escapes to heap
		// @@ make([]int32, len(ordered)) escapes to heap
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			// @@ make([]int32, len(ordered)) escapes to heap
			S := states[state]
			// @@ can inline (*tokens32).PreOrder.func1.1
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, uint32(depth), leaf
			// @@ leaking closure reference states
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
						// @@ inlining call to (*token32).isParentOf
						if c.end != b.begin {
							// @@ inlining call to (*token32).isParentOf
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
				// @@ inlining call to (*token32).isParentOf
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
					// @@ inlining call to (*token32).isParentOf
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{pegRule: ruleSuf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
					// @@ inlining call to (*token32).isParentOf
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	// @@ leaking param content: t
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				// @@ token.token32.begin escapes to heap
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			// @@ rul3s[ordered[i][depths[i] - 1].pegRule] escapes to heap
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			// @@ rul3s[token.token32.pegRule] escapes to heap
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				// @@ token.token32.begin escapes to heap
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			// @@ rul3s[ordered[i][depths[i] - 1].pegRule] escapes to heap
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			// @@ rul3s[token.token32.pegRule] escapes to heap
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					// @@ j escapes to heap
					// @@ token.token32.String() escapes to heap
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
					// @@ j escapes to heap
					// @@ token.token32.String() escapes to heap
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					// @@ c escapes to heap
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				// @@ rul3s[ordered[i][depths[i] - 1].pegRule] escapes to heap
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			// @@ rul3s[token.token32.pegRule] escapes to heap
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	// @@ leaking param content: t
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(string(([]rune(buffer)[token.begin:token.end]))))
	}
	// @@ rul3s[token.token32.pegRule] escapes to heap
	// @@ strconv.Quote(string(([]rune)(buffer)[token.token32.begin:token.token32.end])) escapes to heap
}

func (t *tokens32) Add(rule pegRule, begin, end, depth uint32, index int) {
	t.tree[index] = token32{pegRule: rule, begin: uint32(begin), end: uint32(end), next: uint32(depth)}
	// @@ can inline (*tokens32).Add
}

func (t *tokens32) Tokens() <-chan token32 {
	// @@ leaking param: t
	s := make(chan token32, 16)
	go func() {
		// @@ make(chan token32, 16) escapes to heap
		// @@ make(chan token32, 16) escapes to heap
		for _, v := range t.tree {
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			s <- v.getToken32()
		}
		// @@ inlining call to (*token32).getToken32
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i := range tokens {
		// @@ make([]token32, length) escapes to heap
		// @@ make([]token32, length) escapes to heap
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
		// @@ inlining call to (*token32).getToken32
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
	// @@ leaking param content: t
	tree := t.tree
	// @@ can inline (*tokens32).Expand
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		// @@ make([]token32, 2 * len(tree)) escapes to heap
		// @@ make([]token32, 2 * len(tree)) escapes to heap
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
	rules  [125]func() bool
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
	// @@ leaking param: positions
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)
	// @@ make(textPositionMap, len(positions)) escapes to heap

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
		// @@ make([]int, 2 * len(tokens)) escapes to heap
		// @@ make([]int, 2 * len(tokens)) escapes to heap
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
			// @@ rul3s[token.pegRule] escapes to heap
			translations[end].line, translations[end].symbol,
			// @@ translations[begin].line escapes to heap
			// @@ translations[begin].symbol escapes to heap
			strconv.Quote(string(e.p.buffer[begin:end])))
		// @@ translations[end].line escapes to heap
		// @@ translations[end].symbol escapes to heap
	}
	// @@ strconv.Quote(string(e.p.buffer[begin:end])) escapes to heap

	return error
}

func (p *Parser) PrintSyntaxTree() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	p.tokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *Parser) Highlighter() {
	// @@ leaking param content: p
	p.tokenTree.PrintSyntax()
}

func (p *Parser) Execute() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for token := range p.tokenTree.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

			// @@ string(_buffer[begin:end]) escapes to heap
			// @@ string(_buffer[begin:end]) escapes to heap
			// @@ string(_buffer[begin:end]) escapes to heap
			// @@ string(_buffer[begin:end]) escapes to heap
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
			p.pushNode(unescapeLiteral(buffer[begin:end]))
		case ruleAction6:
			// @@ inlining call to (*Parser).pushNode
			// @@ unescapeLiteral(buffer[begin:end]) escapes to heap
			p.makeDescribe()
		case ruleAction7:
			p.addEvaluationContext()
		case ruleAction8:
			// @@ inlining call to (*Parser).addEvaluationContext
			// @@ inlining call to (*Parser).pushNode
			// @@ &evaluationContextNode literal escapes to heap
			// @@ &evaluationContextNode literal escapes to heap
			// @@ make(map[string]bool) escapes to heap
			p.addPropertyKey(buffer[begin:end])
		case ruleAction9:
			// @@ inlining call to (*Parser).addPropertyKey
			// @@ inlining call to (*Parser).pushNode
			// @@ &evaluationContextKey literal escapes to heap
			// @@ &evaluationContextKey literal escapes to heap
			p.addPropertyValue(buffer[begin:end])
		case ruleAction10:
			// @@ inlining call to (*Parser).addPropertyValue
			// @@ inlining call to (*Parser).pushNode
			// @@ &evaluationContextValue literal escapes to heap
			// @@ &evaluationContextValue literal escapes to heap
			p.insertPropertyKeyValue()
		case ruleAction11:
			p.checkPropertyClause()
		case ruleAction12:
			p.addNullPredicate()
		case ruleAction13:
			p.addExpressionList()
		case ruleAction14:
			// @@ inlining call to (*Parser).addExpressionList
			// @@ inlining call to (*Parser).pushNode
			// @@ []function.Expression literal escapes to heap
			// @@ []function.Expression literal escapes to heap
			p.appendExpression()
		case ruleAction15:
			p.appendExpression()
		case ruleAction16:
			p.addOperatorLiteral("+")
		case ruleAction17:
			// @@ inlining call to (*Parser).addOperatorLiteral
			// @@ inlining call to (*Parser).pushNode
			// @@ &operatorLiteral literal escapes to heap
			// @@ &operatorLiteral literal escapes to heap
			p.addOperatorLiteral("-")
		case ruleAction18:
			// @@ inlining call to (*Parser).addOperatorLiteral
			// @@ inlining call to (*Parser).pushNode
			// @@ &operatorLiteral literal escapes to heap
			// @@ &operatorLiteral literal escapes to heap
			p.addOperatorFunction()
		case ruleAction19:
			p.addOperatorLiteral("/")
		case ruleAction20:
			// @@ inlining call to (*Parser).addOperatorLiteral
			// @@ inlining call to (*Parser).pushNode
			// @@ &operatorLiteral literal escapes to heap
			// @@ &operatorLiteral literal escapes to heap
			p.addOperatorLiteral("*")
		case ruleAction21:
			// @@ inlining call to (*Parser).addOperatorLiteral
			// @@ inlining call to (*Parser).pushNode
			// @@ &operatorLiteral literal escapes to heap
			// @@ &operatorLiteral literal escapes to heap
			p.addOperatorFunction()
		case ruleAction22:
			p.pushNode(unescapeLiteral(buffer[begin:end]))
		case ruleAction23:
			// @@ inlining call to (*Parser).pushNode
			// @@ unescapeLiteral(buffer[begin:end]) escapes to heap
			p.addExpressionList()
		case ruleAction24:
			// @@ inlining call to (*Parser).addExpressionList
			// @@ inlining call to (*Parser).pushNode
			// @@ []function.Expression literal escapes to heap
			// @@ []function.Expression literal escapes to heap

			p.addExpressionList()
			p.addGroupBy()
			// @@ inlining call to (*Parser).addExpressionList
			// @@ inlining call to (*Parser).pushNode
			// @@ []function.Expression literal escapes to heap
			// @@ []function.Expression literal escapes to heap

			// @@ inlining call to (*Parser).addGroupBy
			// @@ inlining call to (*Parser).pushNode
			// @@ &groupByList literal escapes to heap
			// @@ &groupByList literal escapes to heap
			// @@ make([]string, 0) escapes to heap
		case ruleAction25:
			p.addPipeExpression()
		case ruleAction26:
			p.addDurationNode(text)
		case ruleAction27:
			p.addNumberNode(buffer[begin:end])
		case ruleAction28:
			p.addStringNode(unescapeLiteral(buffer[begin:end]))
		case ruleAction29:
			// @@ inlining call to (*Parser).addStringNode
			// @@ inlining call to (*Parser).pushNode
			// @@ expression.String literal escapes to heap
			p.addAnnotationExpression(buffer[begin:end])
		case ruleAction30:
			p.addGroupBy()
		case ruleAction31:
			// @@ inlining call to (*Parser).addGroupBy
			// @@ inlining call to (*Parser).pushNode
			// @@ &groupByList literal escapes to heap
			// @@ &groupByList literal escapes to heap
			// @@ make([]string, 0) escapes to heap

			p.pushNode(unescapeLiteral(buffer[begin:end]))

			// @@ inlining call to (*Parser).pushNode
			// @@ unescapeLiteral(buffer[begin:end]) escapes to heap
		case ruleAction32:

			p.addFunctionInvocation()

		case ruleAction33:

			p.pushNode(unescapeLiteral(buffer[begin:end]))

			// @@ inlining call to (*Parser).pushNode
			// @@ unescapeLiteral(buffer[begin:end]) escapes to heap
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

			p.pushNode(unescapeLiteral(buffer[begin:end]))

			// @@ inlining call to (*Parser).pushNode
			// @@ unescapeLiteral(buffer[begin:end]) escapes to heap
		case ruleAction48:
			p.addLiteralList()
		case ruleAction49:
			// @@ inlining call to (*Parser).addLiteralList
			// @@ inlining call to (*Parser).pushNode
			// @@ &stringLiteralList literal escapes to heap
			// @@ &stringLiteralList literal escapes to heap
			// @@ make([]string, 0) escapes to heap

			p.appendLiteral(unescapeLiteral(buffer[begin:end]))

		case ruleAction50:
			p.addTagLiteral(unescapeLiteral(buffer[begin:end]))

			// @@ inlining call to (*Parser).addTagLiteral
			// @@ inlining call to (*Parser).pushNode
			// @@ &tagLiteral literal escapes to heap
			// @@ &tagLiteral literal escapes to heap
		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *Parser) Init() {
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param content: p
	// @@ leaking param: p
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
		// @@ ([]rune)(p.Buffer) escapes to heap
		p.buffer = append(p.buffer, endSymbol)
	}

	var tree tokenTree = &tokens32{tree: make([]token32, math.MaxInt16)}
	var max token32
	// @@ &tokens32 literal escapes to heap
	// @@ &tokens32 literal escapes to heap
	// @@ make([]token32, math.MaxInt16) escapes to heap
	// @@ &tokens32 literal escapes to heap
	// @@ &tokens32 literal escapes to heap
	// @@ make([]token32, math.MaxInt16) escapes to heap
	// @@ &tokens32 literal escapes to heap
	// @@ &tokens32 literal escapes to heap
	// @@ make([]token32, math.MaxInt16) escapes to heap
	// @@ make([]token32, math.MaxInt16) escapes to heap
	// @@ moved to heap: tree
	// @@ &tokens32 literal escapes to heap
	// @@ &tokens32 literal escapes to heap
	// @@ make([]token32, math.MaxInt16) escapes to heap
	// @@ &tokens32 literal escapes to heap
	// @@ &tokens32 literal escapes to heap
	// @@ make([]token32, math.MaxInt16) escapes to heap
	position, depth, tokenIndex, buffer, _rules := uint32(0), uint32(0), 0, p.buffer, p.rules
	// @@ moved to heap: max

	// @@ moved to heap: tokenIndex
	// @@ moved to heap: position
	// @@ moved to heap: depth
	// @@ moved to heap: _rules
	p.Parse = func(rule ...int) error {
		r := 1
		// @@ func literal escapes to heap
		// @@ func literal escapes to heap
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokenTree = tree
		if matches {
			// @@ leaking closure reference tree
			// @@ &tree escapes to heap
			p.tokenTree.trim(tokenIndex)
			return nil
			// @@ &tokenIndex escapes to heap
		}
		return &parseError{p, max}
	}
	// @@ &max escapes to heap
	// @@ &parseError literal escapes to heap
	// @@ &parseError literal escapes to heap

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
		// @@ can inline (*Parser).Init.func2
		// @@ func literal escapes to heap
		// @@ func literal escapes to heap
	}
	// @@ &position escapes to heap
	// @@ &tokenIndex escapes to heap
	// @@ &depth escapes to heap

	add := func(rule pegRule, begin uint32) {
		if t := tree.Expand(tokenIndex); t != nil {
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			tree = t
			// @@ leaking closure reference tree
			// @@ leaking closure reference tree
			// @@ &tree escapes to heap
			// @@ &tokenIndex escapes to heap
		}
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
		// @@ &position escapes to heap
		// @@ &depth escapes to heap
		if begin != position && position > max.end {
			max = token32{rule, begin, position, depth}
			// @@ &max escapes to heap
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
			// @@ can inline (*Parser).Init.func4
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			position++
			// @@ &position escapes to heap
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
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position1 := position
				depth++
				{
					position2, tokenIndex2, depth2 := position, tokenIndex, depth
					{
						position4 := position
						depth++
						if !_rules[rule_]() {
							goto l3
							// @@ &_rules escapes to heap
							// @@ &_rules escapes to heap
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
					position120, tokenIndex120, depth120 := position, tokenIndex, depth
					if !matchDot() {
						goto l120
					}
					goto l0
				l120:
					position, tokenIndex, depth = position120, tokenIndex120, depth120
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
		/* 7 describeSingleStmt <- <(_ <METRIC_NAME> Action5 optionalPredicateClause Action6)> */
		nil,
		/* 8 propertyClause <- <(Action7 (_ PROPERTY_KEY Action8 _ PROPERTY_VALUE Action9 Action10)* Action11)> */
		nil,
		/* 9 optionalPredicateClause <- <(predicateClause / Action12)> */
		func() bool {
			{
				// @@ func literal escapes to heap
				// @@ func literal escapes to heap
				position130 := position
				depth++
				// @@ &position escapes to heap
				{
					// @@ &depth escapes to heap
					position131, tokenIndex131, depth131 := position, tokenIndex, depth
					{
						// @@ &tokenIndex escapes to heap
						position133 := position
						depth++
						if !_rules[rule_]() {
							goto l132
							// @@ &_rules escapes to heap
						}
						{
							position134, tokenIndex134, depth134 := position, tokenIndex, depth
							if buffer[position] != rune('w') {
								goto l135
							}
							position++
							goto l134
						l135:
							position, tokenIndex, depth = position134, tokenIndex134, depth134
							if buffer[position] != rune('W') {
								goto l132
							}
							position++
						}
					l134:
						{
							position136, tokenIndex136, depth136 := position, tokenIndex, depth
							if buffer[position] != rune('h') {
								goto l137
							}
							position++
							goto l136
						l137:
							position, tokenIndex, depth = position136, tokenIndex136, depth136
							if buffer[position] != rune('H') {
								goto l132
							}
							position++
						}
					l136:
						{
							position138, tokenIndex138, depth138 := position, tokenIndex, depth
							if buffer[position] != rune('e') {
								goto l139
							}
							position++
							goto l138
						l139:
							position, tokenIndex, depth = position138, tokenIndex138, depth138
							if buffer[position] != rune('E') {
								goto l132
							}
							position++
						}
					l138:
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
								goto l132
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
								goto l132
							}
							position++
						}
					l142:
						if !_rules[ruleKEY]() {
							goto l132
						}
						if !_rules[rule_]() {
							goto l132
						}
						if !_rules[rulepredicate_1]() {
							goto l132
						}
						depth--
						add(rulepredicateClause, position133)
					}
					goto l131
				l132:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
					{
						add(ruleAction12, position)
					}
				}
			l131:
				depth--
				add(ruleoptionalPredicateClause, position130)
			}
			return true
		},
		/* 10 expressionList <- <(Action13 expression_start Action14 (_ COMMA expression_start Action15)*)> */
		func() bool {
			position145, tokenIndex145, depth145 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position146 := position
				depth++
				{
					add(ruleAction13, position)
				}
				if !_rules[ruleexpression_start]() {
					goto l145
					// @@ &_rules escapes to heap
				}
				{
					add(ruleAction14, position)
				}
			l149:
				{
					position150, tokenIndex150, depth150 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l150
					}
					if !_rules[ruleCOMMA]() {
						goto l150
					}
					if !_rules[ruleexpression_start]() {
						goto l150
					}
					{
						add(ruleAction15, position)
					}
					goto l149
				l150:
					position, tokenIndex, depth = position150, tokenIndex150, depth150
				}
				depth--
				add(ruleexpressionList, position146)
			}
			return true
		l145:
			position, tokenIndex, depth = position145, tokenIndex145, depth145
			return false
		},
		/* 11 expression_start <- <(expression_sum add_pipe)> */
		func() bool {
			position152, tokenIndex152, depth152 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position153 := position
				depth++
				{
					position154 := position
					depth++
					if !_rules[ruleexpression_product]() {
						goto l152
						// @@ &_rules escapes to heap
					}
				l155:
					{
						position156, tokenIndex156, depth156 := position, tokenIndex, depth
						if !_rules[ruleadd_pipe]() {
							goto l156
						}
						{
							position157, tokenIndex157, depth157 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l158
							}
							{
								position159 := position
								depth++
								if buffer[position] != rune('+') {
									goto l158
								}
								position++
								depth--
								add(ruleOP_ADD, position159)
							}
							{
								add(ruleAction16, position)
							}
							goto l157
						l158:
							position, tokenIndex, depth = position157, tokenIndex157, depth157
							if !_rules[rule_]() {
								goto l156
							}
							{
								position161 := position
								depth++
								if buffer[position] != rune('-') {
									goto l156
								}
								position++
								depth--
								add(ruleOP_SUB, position161)
							}
							{
								add(ruleAction17, position)
							}
						}
					l157:
						if !_rules[ruleexpression_product]() {
							goto l156
						}
						{
							add(ruleAction18, position)
						}
						goto l155
					l156:
						position, tokenIndex, depth = position156, tokenIndex156, depth156
					}
					depth--
					add(ruleexpression_sum, position154)
				}
				if !_rules[ruleadd_pipe]() {
					goto l152
				}
				depth--
				add(ruleexpression_start, position153)
			}
			return true
		l152:
			position, tokenIndex, depth = position152, tokenIndex152, depth152
			return false
		},
		/* 12 expression_sum <- <(expression_product (add_pipe ((_ OP_ADD Action16) / (_ OP_SUB Action17)) expression_product Action18)*)> */
		nil,
		/* 13 expression_product <- <(expression_atom (add_pipe ((_ OP_DIV Action19) / (_ OP_MULT Action20)) expression_atom Action21)*)> */
		func() bool {
			position165, tokenIndex165, depth165 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position166 := position
				depth++
				if !_rules[ruleexpression_atom]() {
					goto l165
					// @@ &_rules escapes to heap
				}
			l167:
				{
					position168, tokenIndex168, depth168 := position, tokenIndex, depth
					if !_rules[ruleadd_pipe]() {
						goto l168
					}
					{
						position169, tokenIndex169, depth169 := position, tokenIndex, depth
						if !_rules[rule_]() {
							goto l170
						}
						{
							position171 := position
							depth++
							if buffer[position] != rune('/') {
								goto l170
							}
							position++
							depth--
							add(ruleOP_DIV, position171)
						}
						{
							add(ruleAction19, position)
						}
						goto l169
					l170:
						position, tokenIndex, depth = position169, tokenIndex169, depth169
						if !_rules[rule_]() {
							goto l168
						}
						{
							position173 := position
							depth++
							if buffer[position] != rune('*') {
								goto l168
							}
							position++
							depth--
							add(ruleOP_MULT, position173)
						}
						{
							add(ruleAction20, position)
						}
					}
				l169:
					if !_rules[ruleexpression_atom]() {
						goto l168
					}
					{
						add(ruleAction21, position)
					}
					goto l167
				l168:
					position, tokenIndex, depth = position168, tokenIndex168, depth168
				}
				depth--
				add(ruleexpression_product, position166)
			}
			return true
		l165:
			position, tokenIndex, depth = position165, tokenIndex165, depth165
			return false
		},
		/* 14 add_one_pipe <- <(_ OP_PIPE _ <IDENTIFIER> Action22 ((_ PAREN_OPEN (expressionList / Action23) optionalGroupBy _ PAREN_CLOSE) / Action24) Action25 expression_annotation)> */
		nil,
		/* 15 add_pipe <- <add_one_pipe*> */
		func() bool {
			{
				// @@ func literal escapes to heap
				// @@ func literal escapes to heap
				position178 := position
				depth++
				// @@ &position escapes to heap
			l179:
				// @@ &depth escapes to heap
				{
					position180, tokenIndex180, depth180 := position, tokenIndex, depth
					{
						// @@ &tokenIndex escapes to heap
						position181 := position
						depth++
						if !_rules[rule_]() {
							goto l180
							// @@ &_rules escapes to heap
						}
						{
							position182 := position
							depth++
							if buffer[position] != rune('|') {
								goto l180
							}
							position++
							depth--
							add(ruleOP_PIPE, position182)
						}
						if !_rules[rule_]() {
							goto l180
						}
						{
							position183 := position
							depth++
							if !_rules[ruleIDENTIFIER]() {
								goto l180
							}
							depth--
							add(rulePegText, position183)
						}
						{
							add(ruleAction22, position)
						}
						{
							position185, tokenIndex185, depth185 := position, tokenIndex, depth
							if !_rules[rule_]() {
								goto l186
							}
							if !_rules[rulePAREN_OPEN]() {
								goto l186
							}
							{
								position187, tokenIndex187, depth187 := position, tokenIndex, depth
								if !_rules[ruleexpressionList]() {
									goto l188
								}
								goto l187
							l188:
								position, tokenIndex, depth = position187, tokenIndex187, depth187
								{
									add(ruleAction23, position)
								}
							}
						l187:
							if !_rules[ruleoptionalGroupBy]() {
								goto l186
							}
							if !_rules[rule_]() {
								goto l186
							}
							if !_rules[rulePAREN_CLOSE]() {
								goto l186
							}
							goto l185
						l186:
							position, tokenIndex, depth = position185, tokenIndex185, depth185
							{
								add(ruleAction24, position)
							}
						}
					l185:
						{
							add(ruleAction25, position)
						}
						if !_rules[ruleexpression_annotation]() {
							goto l180
						}
						depth--
						add(ruleadd_one_pipe, position181)
					}
					goto l179
				l180:
					position, tokenIndex, depth = position180, tokenIndex180, depth180
				}
				depth--
				add(ruleadd_pipe, position178)
			}
			return true
		},
		/* 16 expression_atom <- <(expression_atom_raw expression_annotation)> */
		func() bool {
			position192, tokenIndex192, depth192 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position193 := position
				depth++
				{
					position194 := position
					depth++
					{
						position195, tokenIndex195, depth195 := position, tokenIndex, depth
						{
							position197 := position
							depth++
							if !_rules[rule_]() {
								goto l196
								// @@ &_rules escapes to heap
							}
							{
								position198 := position
								depth++
								if !_rules[ruleIDENTIFIER]() {
									goto l196
								}
								depth--
								add(rulePegText, position198)
							}
							{
								add(ruleAction31, position)
							}
							if !_rules[rule_]() {
								goto l196
							}
							if !_rules[rulePAREN_OPEN]() {
								goto l196
							}
							if !_rules[ruleexpressionList]() {
								goto l196
							}
							if !_rules[ruleoptionalGroupBy]() {
								goto l196
							}
							if !_rules[rule_]() {
								goto l196
							}
							if !_rules[rulePAREN_CLOSE]() {
								goto l196
							}
							{
								add(ruleAction32, position)
							}
							depth--
							add(ruleexpression_function, position197)
						}
						goto l195
					l196:
						position, tokenIndex, depth = position195, tokenIndex195, depth195
						{
							position202 := position
							depth++
							if !_rules[rule_]() {
								goto l201
							}
							{
								position203 := position
								depth++
								if !_rules[ruleIDENTIFIER]() {
									goto l201
								}
								depth--
								add(rulePegText, position203)
							}
							{
								add(ruleAction33, position)
							}
							{
								position205, tokenIndex205, depth205 := position, tokenIndex, depth
								{
									position207, tokenIndex207, depth207 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l208
									}
									if buffer[position] != rune('[') {
										goto l208
									}
									position++
									if !_rules[rulepredicate_1]() {
										goto l208
									}
									if !_rules[rule_]() {
										goto l208
									}
									if buffer[position] != rune(']') {
										goto l208
									}
									position++
									goto l207
								l208:
									position, tokenIndex, depth = position207, tokenIndex207, depth207
									{
										add(ruleAction34, position)
									}
								}
							l207:
								goto l206

								position, tokenIndex, depth = position205, tokenIndex205, depth205
							}
						l206:
							{
								add(ruleAction35, position)
							}
							depth--
							add(ruleexpression_metric, position202)
						}
						goto l195
					l201:
						position, tokenIndex, depth = position195, tokenIndex195, depth195
						if !_rules[rule_]() {
							goto l211
						}
						if !_rules[rulePAREN_OPEN]() {
							goto l211
						}
						if !_rules[ruleexpression_start]() {
							goto l211
						}
						if !_rules[rule_]() {
							goto l211
						}
						if !_rules[rulePAREN_CLOSE]() {
							goto l211
						}
						goto l195
					l211:
						position, tokenIndex, depth = position195, tokenIndex195, depth195
						if !_rules[rule_]() {
							goto l212
						}
						{
							position213 := position
							depth++
							{
								position214 := position
								depth++
								if !_rules[ruleNUMBER]() {
									goto l212
								}
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l212
								}
								position++
							l215:
								{
									position216, tokenIndex216, depth216 := position, tokenIndex, depth
									if c := buffer[position]; c < rune('a') || c > rune('z') {
										goto l216
									}
									position++
									goto l215
								l216:
									position, tokenIndex, depth = position216, tokenIndex216, depth216
								}
								if !_rules[ruleKEY]() {
									goto l212
								}
								depth--
								add(ruleDURATION, position214)
							}
							depth--
							add(rulePegText, position213)
						}
						{
							add(ruleAction26, position)
						}
						goto l195
					l212:
						position, tokenIndex, depth = position195, tokenIndex195, depth195
						if !_rules[rule_]() {
							goto l218
						}
						{
							position219 := position
							depth++
							if !_rules[ruleNUMBER]() {
								goto l218
							}
							depth--
							add(rulePegText, position219)
						}
						{
							add(ruleAction27, position)
						}
						goto l195
					l218:
						position, tokenIndex, depth = position195, tokenIndex195, depth195
						if !_rules[rule_]() {
							goto l192
						}
						if !_rules[ruleSTRING]() {
							goto l192
						}
						{
							add(ruleAction28, position)
						}
					}
				l195:
					depth--
					add(ruleexpression_atom_raw, position194)
				}
				if !_rules[ruleexpression_annotation]() {
					goto l192
				}
				depth--
				add(ruleexpression_atom, position193)
			}
			return true
		l192:
			position, tokenIndex, depth = position192, tokenIndex192, depth192
			return false
		},
		/* 17 expression_atom_raw <- <(expression_function / expression_metric / (_ PAREN_OPEN expression_start _ PAREN_CLOSE) / (_ <DURATION> Action26) / (_ <NUMBER> Action27) / (_ STRING Action28))> */
		nil,
		/* 18 expression_annotation_required <- <(_ '{' <(!'}' .)*> '}' Action29)> */
		nil,
		/* 19 expression_annotation <- <expression_annotation_required?> */
		func() bool {
			{
				// @@ func literal escapes to heap
				// @@ func literal escapes to heap
				position225 := position
				depth++
				// @@ &position escapes to heap
				{
					// @@ &depth escapes to heap
					position226, tokenIndex226, depth226 := position, tokenIndex, depth
					{
						// @@ &tokenIndex escapes to heap
						position228 := position
						depth++
						if !_rules[rule_]() {
							goto l226
							// @@ &_rules escapes to heap
						}
						if buffer[position] != rune('{') {
							goto l226
						}
						position++
						{
							position229 := position
							depth++
						l230:
							{
								position231, tokenIndex231, depth231 := position, tokenIndex, depth
								{
									position232, tokenIndex232, depth232 := position, tokenIndex, depth
									if buffer[position] != rune('}') {
										goto l232
									}
									position++
									goto l231
								l232:
									position, tokenIndex, depth = position232, tokenIndex232, depth232
								}
								if !matchDot() {
									goto l231
								}
								goto l230
							l231:
								position, tokenIndex, depth = position231, tokenIndex231, depth231
							}
							depth--
							add(rulePegText, position229)
						}
						if buffer[position] != rune('}') {
							goto l226
						}
						position++
						{
							add(ruleAction29, position)
						}
						depth--
						add(ruleexpression_annotation_required, position228)
					}
					goto l227
				l226:
					position, tokenIndex, depth = position226, tokenIndex226, depth226
				}
			l227:
				depth--
				add(ruleexpression_annotation, position225)
			}
			return true
		},
		/* 20 optionalGroupBy <- <(Action30 (groupByClause / collapseByClause)?)> */
		func() bool {
			{
				// @@ func literal escapes to heap
				// @@ func literal escapes to heap
				position235 := position
				depth++
				// @@ &position escapes to heap
				{
					// @@ &depth escapes to heap
					add(ruleAction30, position)
				}
				{
					position237, tokenIndex237, depth237 := position, tokenIndex, depth
					{
						// @@ &tokenIndex escapes to heap
						position239, tokenIndex239, depth239 := position, tokenIndex, depth
						{
							position241 := position
							depth++
							if !_rules[rule_]() {
								goto l240
								// @@ &_rules escapes to heap
							}
							{
								position242, tokenIndex242, depth242 := position, tokenIndex, depth
								if buffer[position] != rune('g') {
									goto l243
								}
								position++
								goto l242
							l243:
								position, tokenIndex, depth = position242, tokenIndex242, depth242
								if buffer[position] != rune('G') {
									goto l240
								}
								position++
							}
						l242:
							{
								position244, tokenIndex244, depth244 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l245
								}
								position++
								goto l244
							l245:
								position, tokenIndex, depth = position244, tokenIndex244, depth244
								if buffer[position] != rune('R') {
									goto l240
								}
								position++
							}
						l244:
							{
								position246, tokenIndex246, depth246 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l247
								}
								position++
								goto l246
							l247:
								position, tokenIndex, depth = position246, tokenIndex246, depth246
								if buffer[position] != rune('O') {
									goto l240
								}
								position++
							}
						l246:
							{
								position248, tokenIndex248, depth248 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l249
								}
								position++
								goto l248
							l249:
								position, tokenIndex, depth = position248, tokenIndex248, depth248
								if buffer[position] != rune('U') {
									goto l240
								}
								position++
							}
						l248:
							{
								position250, tokenIndex250, depth250 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l251
								}
								position++
								goto l250
							l251:
								position, tokenIndex, depth = position250, tokenIndex250, depth250
								if buffer[position] != rune('P') {
									goto l240
								}
								position++
							}
						l250:
							if !_rules[ruleKEY]() {
								goto l240
							}
							if !_rules[rule_]() {
								goto l240
							}
							{
								position252, tokenIndex252, depth252 := position, tokenIndex, depth
								if buffer[position] != rune('b') {
									goto l253
								}
								position++
								goto l252
							l253:
								position, tokenIndex, depth = position252, tokenIndex252, depth252
								if buffer[position] != rune('B') {
									goto l240
								}
								position++
							}
						l252:
							{
								position254, tokenIndex254, depth254 := position, tokenIndex, depth
								if buffer[position] != rune('y') {
									goto l255
								}
								position++
								goto l254
							l255:
								position, tokenIndex, depth = position254, tokenIndex254, depth254
								if buffer[position] != rune('Y') {
									goto l240
								}
								position++
							}
						l254:
							if !_rules[ruleKEY]() {
								goto l240
							}
							if !_rules[rule_]() {
								goto l240
							}
							{
								position256 := position
								depth++
								if !_rules[ruleCOLUMN_NAME]() {
									goto l240
								}
								depth--
								add(rulePegText, position256)
							}
							{
								add(ruleAction36, position)
							}
						l258:
							{
								position259, tokenIndex259, depth259 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l259
								}
								if !_rules[ruleCOMMA]() {
									goto l259
								}
								if !_rules[rule_]() {
									goto l259
								}
								{
									position260 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l259
									}
									depth--
									add(rulePegText, position260)
								}
								{
									add(ruleAction37, position)
								}
								goto l258
							l259:
								position, tokenIndex, depth = position259, tokenIndex259, depth259
							}
							depth--
							add(rulegroupByClause, position241)
						}
						goto l239
					l240:
						position, tokenIndex, depth = position239, tokenIndex239, depth239
						{
							position262 := position
							depth++
							if !_rules[rule_]() {
								goto l237
							}
							{
								position263, tokenIndex263, depth263 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l264
								}
								position++
								goto l263
							l264:
								position, tokenIndex, depth = position263, tokenIndex263, depth263
								if buffer[position] != rune('C') {
									goto l237
								}
								position++
							}
						l263:
							{
								position265, tokenIndex265, depth265 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l266
								}
								position++
								goto l265
							l266:
								position, tokenIndex, depth = position265, tokenIndex265, depth265
								if buffer[position] != rune('O') {
									goto l237
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
									goto l237
								}
								position++
							}
						l267:
							{
								position269, tokenIndex269, depth269 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l270
								}
								position++
								goto l269
							l270:
								position, tokenIndex, depth = position269, tokenIndex269, depth269
								if buffer[position] != rune('L') {
									goto l237
								}
								position++
							}
						l269:
							{
								position271, tokenIndex271, depth271 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l272
								}
								position++
								goto l271
							l272:
								position, tokenIndex, depth = position271, tokenIndex271, depth271
								if buffer[position] != rune('A') {
									goto l237
								}
								position++
							}
						l271:
							{
								position273, tokenIndex273, depth273 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l274
								}
								position++
								goto l273
							l274:
								position, tokenIndex, depth = position273, tokenIndex273, depth273
								if buffer[position] != rune('P') {
									goto l237
								}
								position++
							}
						l273:
							{
								position275, tokenIndex275, depth275 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l276
								}
								position++
								goto l275
							l276:
								position, tokenIndex, depth = position275, tokenIndex275, depth275
								if buffer[position] != rune('S') {
									goto l237
								}
								position++
							}
						l275:
							{
								position277, tokenIndex277, depth277 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l278
								}
								position++
								goto l277
							l278:
								position, tokenIndex, depth = position277, tokenIndex277, depth277
								if buffer[position] != rune('E') {
									goto l237
								}
								position++
							}
						l277:
							if !_rules[ruleKEY]() {
								goto l237
							}
							if !_rules[rule_]() {
								goto l237
							}
							{
								position279, tokenIndex279, depth279 := position, tokenIndex, depth
								if buffer[position] != rune('b') {
									goto l280
								}
								position++
								goto l279
							l280:
								position, tokenIndex, depth = position279, tokenIndex279, depth279
								if buffer[position] != rune('B') {
									goto l237
								}
								position++
							}
						l279:
							{
								position281, tokenIndex281, depth281 := position, tokenIndex, depth
								if buffer[position] != rune('y') {
									goto l282
								}
								position++
								goto l281
							l282:
								position, tokenIndex, depth = position281, tokenIndex281, depth281
								if buffer[position] != rune('Y') {
									goto l237
								}
								position++
							}
						l281:
							if !_rules[ruleKEY]() {
								goto l237
							}
							if !_rules[rule_]() {
								goto l237
							}
							{
								position283 := position
								depth++
								if !_rules[ruleCOLUMN_NAME]() {
									goto l237
								}
								depth--
								add(rulePegText, position283)
							}
							{
								add(ruleAction38, position)
							}
						l285:
							{
								position286, tokenIndex286, depth286 := position, tokenIndex, depth
								if !_rules[rule_]() {
									goto l286
								}
								if !_rules[ruleCOMMA]() {
									goto l286
								}
								if !_rules[rule_]() {
									goto l286
								}
								{
									position287 := position
									depth++
									if !_rules[ruleCOLUMN_NAME]() {
										goto l286
									}
									depth--
									add(rulePegText, position287)
								}
								{
									add(ruleAction39, position)
								}
								goto l285
							l286:
								position, tokenIndex, depth = position286, tokenIndex286, depth286
							}
							depth--
							add(rulecollapseByClause, position262)
						}
					}
				l239:
					goto l238
				l237:
					position, tokenIndex, depth = position237, tokenIndex237, depth237
				}
			l238:
				depth--
				add(ruleoptionalGroupBy, position235)
			}
			return true
		},
		/* 21 expression_function <- <(_ <IDENTIFIER> Action31 _ PAREN_OPEN expressionList optionalGroupBy _ PAREN_CLOSE Action32)> */
		nil,
		/* 22 expression_metric <- <(_ <IDENTIFIER> Action33 ((_ '[' predicate_1 _ ']') / Action34)? Action35)> */
		nil,
		/* 23 groupByClause <- <(_ (('g' / 'G') ('r' / 'R') ('o' / 'O') ('u' / 'U') ('p' / 'P')) KEY _ (('b' / 'B') ('y' / 'Y')) KEY _ <COLUMN_NAME> Action36 (_ COMMA _ <COLUMN_NAME> Action37)*)> */
		nil,
		/* 24 collapseByClause <- <(_ (('c' / 'C') ('o' / 'O') ('l' / 'L') ('l' / 'L') ('a' / 'A') ('p' / 'P') ('s' / 'S') ('e' / 'E')) KEY _ (('b' / 'B') ('y' / 'Y')) KEY _ <COLUMN_NAME> Action38 (_ COMMA _ <COLUMN_NAME> Action39)*)> */
		nil,
		/* 25 predicateClause <- <(_ (('w' / 'W') ('h' / 'H') ('e' / 'E') ('r' / 'R') ('e' / 'E')) KEY _ predicate_1)> */
		nil,
		/* 26 predicate_1 <- <((predicate_2 _ OP_OR predicate_1 Action40) / predicate_2)> */
		func() bool {
			position294, tokenIndex294, depth294 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position295 := position
				depth++
				{
					position296, tokenIndex296, depth296 := position, tokenIndex, depth
					if !_rules[rulepredicate_2]() {
						goto l297
						// @@ &_rules escapes to heap
					}
					if !_rules[rule_]() {
						goto l297
					}
					{
						position298 := position
						depth++
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
								goto l297
							}
							position++
						}
					l299:
						{
							position301, tokenIndex301, depth301 := position, tokenIndex, depth
							if buffer[position] != rune('r') {
								goto l302
							}
							position++
							goto l301
						l302:
							position, tokenIndex, depth = position301, tokenIndex301, depth301
							if buffer[position] != rune('R') {
								goto l297
							}
							position++
						}
					l301:
						if !_rules[ruleKEY]() {
							goto l297
						}
						depth--
						add(ruleOP_OR, position298)
					}
					if !_rules[rulepredicate_1]() {
						goto l297
					}
					{
						add(ruleAction40, position)
					}
					goto l296
				l297:
					position, tokenIndex, depth = position296, tokenIndex296, depth296
					if !_rules[rulepredicate_2]() {
						goto l294
					}
				}
			l296:
				depth--
				add(rulepredicate_1, position295)
			}
			return true
		l294:
			position, tokenIndex, depth = position294, tokenIndex294, depth294
			return false
		},
		/* 27 predicate_2 <- <((predicate_3 _ OP_AND predicate_2 Action41) / predicate_3)> */
		func() bool {
			position304, tokenIndex304, depth304 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position305 := position
				depth++
				{
					position306, tokenIndex306, depth306 := position, tokenIndex, depth
					if !_rules[rulepredicate_3]() {
						goto l307
						// @@ &_rules escapes to heap
					}
					if !_rules[rule_]() {
						goto l307
					}
					{
						position308 := position
						depth++
						{
							position309, tokenIndex309, depth309 := position, tokenIndex, depth
							if buffer[position] != rune('a') {
								goto l310
							}
							position++
							goto l309
						l310:
							position, tokenIndex, depth = position309, tokenIndex309, depth309
							if buffer[position] != rune('A') {
								goto l307
							}
							position++
						}
					l309:
						{
							position311, tokenIndex311, depth311 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l312
							}
							position++
							goto l311
						l312:
							position, tokenIndex, depth = position311, tokenIndex311, depth311
							if buffer[position] != rune('N') {
								goto l307
							}
							position++
						}
					l311:
						{
							position313, tokenIndex313, depth313 := position, tokenIndex, depth
							if buffer[position] != rune('d') {
								goto l314
							}
							position++
							goto l313
						l314:
							position, tokenIndex, depth = position313, tokenIndex313, depth313
							if buffer[position] != rune('D') {
								goto l307
							}
							position++
						}
					l313:
						if !_rules[ruleKEY]() {
							goto l307
						}
						depth--
						add(ruleOP_AND, position308)
					}
					if !_rules[rulepredicate_2]() {
						goto l307
					}
					{
						add(ruleAction41, position)
					}
					goto l306
				l307:
					position, tokenIndex, depth = position306, tokenIndex306, depth306
					if !_rules[rulepredicate_3]() {
						goto l304
					}
				}
			l306:
				depth--
				add(rulepredicate_2, position305)
			}
			return true
		l304:
			position, tokenIndex, depth = position304, tokenIndex304, depth304
			return false
		},
		/* 28 predicate_3 <- <((_ OP_NOT predicate_3 Action42) / (_ PAREN_OPEN predicate_1 _ PAREN_CLOSE) / tagMatcher)> */
		func() bool {
			position316, tokenIndex316, depth316 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position317 := position
				depth++
				{
					position318, tokenIndex318, depth318 := position, tokenIndex, depth
					if !_rules[rule_]() {
						goto l319
						// @@ &_rules escapes to heap
					}
					{
						position320 := position
						depth++
						{
							position321, tokenIndex321, depth321 := position, tokenIndex, depth
							if buffer[position] != rune('n') {
								goto l322
							}
							position++
							goto l321
						l322:
							position, tokenIndex, depth = position321, tokenIndex321, depth321
							if buffer[position] != rune('N') {
								goto l319
							}
							position++
						}
					l321:
						{
							position323, tokenIndex323, depth323 := position, tokenIndex, depth
							if buffer[position] != rune('o') {
								goto l324
							}
							position++
							goto l323
						l324:
							position, tokenIndex, depth = position323, tokenIndex323, depth323
							if buffer[position] != rune('O') {
								goto l319
							}
							position++
						}
					l323:
						{
							position325, tokenIndex325, depth325 := position, tokenIndex, depth
							if buffer[position] != rune('t') {
								goto l326
							}
							position++
							goto l325
						l326:
							position, tokenIndex, depth = position325, tokenIndex325, depth325
							if buffer[position] != rune('T') {
								goto l319
							}
							position++
						}
					l325:
						if !_rules[ruleKEY]() {
							goto l319
						}
						depth--
						add(ruleOP_NOT, position320)
					}
					if !_rules[rulepredicate_3]() {
						goto l319
					}
					{
						add(ruleAction42, position)
					}
					goto l318
				l319:
					position, tokenIndex, depth = position318, tokenIndex318, depth318
					if !_rules[rule_]() {
						goto l328
					}
					if !_rules[rulePAREN_OPEN]() {
						goto l328
					}
					if !_rules[rulepredicate_1]() {
						goto l328
					}
					if !_rules[rule_]() {
						goto l328
					}
					if !_rules[rulePAREN_CLOSE]() {
						goto l328
					}
					goto l318
				l328:
					position, tokenIndex, depth = position318, tokenIndex318, depth318
					{
						position329 := position
						depth++
						{
							position330, tokenIndex330, depth330 := position, tokenIndex, depth
							if !_rules[ruletagName]() {
								goto l331
							}
							if !_rules[rule_]() {
								goto l331
							}
							if buffer[position] != rune('=') {
								goto l331
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l331
							}
							{
								add(ruleAction43, position)
							}
							goto l330
						l331:
							position, tokenIndex, depth = position330, tokenIndex330, depth330
							if !_rules[ruletagName]() {
								goto l333
							}
							if !_rules[rule_]() {
								goto l333
							}
							if buffer[position] != rune('!') {
								goto l333
							}
							position++
							if buffer[position] != rune('=') {
								goto l333
							}
							position++
							if !_rules[ruleliteralString]() {
								goto l333
							}
							{
								add(ruleAction44, position)
							}
							goto l330
						l333:
							position, tokenIndex, depth = position330, tokenIndex330, depth330
							if !_rules[ruletagName]() {
								goto l335
							}
							if !_rules[rule_]() {
								goto l335
							}
							{
								position336, tokenIndex336, depth336 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l337
								}
								position++
								goto l336
							l337:
								position, tokenIndex, depth = position336, tokenIndex336, depth336
								if buffer[position] != rune('M') {
									goto l335
								}
								position++
							}
						l336:
							{
								position338, tokenIndex338, depth338 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l339
								}
								position++
								goto l338
							l339:
								position, tokenIndex, depth = position338, tokenIndex338, depth338
								if buffer[position] != rune('A') {
									goto l335
								}
								position++
							}
						l338:
							{
								position340, tokenIndex340, depth340 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l341
								}
								position++
								goto l340
							l341:
								position, tokenIndex, depth = position340, tokenIndex340, depth340
								if buffer[position] != rune('T') {
									goto l335
								}
								position++
							}
						l340:
							{
								position342, tokenIndex342, depth342 := position, tokenIndex, depth
								if buffer[position] != rune('c') {
									goto l343
								}
								position++
								goto l342
							l343:
								position, tokenIndex, depth = position342, tokenIndex342, depth342
								if buffer[position] != rune('C') {
									goto l335
								}
								position++
							}
						l342:
							{
								position344, tokenIndex344, depth344 := position, tokenIndex, depth
								if buffer[position] != rune('h') {
									goto l345
								}
								position++
								goto l344
							l345:
								position, tokenIndex, depth = position344, tokenIndex344, depth344
								if buffer[position] != rune('H') {
									goto l335
								}
								position++
							}
						l344:
							if !_rules[ruleKEY]() {
								goto l335
							}
							if !_rules[ruleliteralString]() {
								goto l335
							}
							{
								add(ruleAction45, position)
							}
							goto l330
						l335:
							position, tokenIndex, depth = position330, tokenIndex330, depth330
							if !_rules[ruletagName]() {
								goto l316
							}
							if !_rules[rule_]() {
								goto l316
							}
							{
								position347, tokenIndex347, depth347 := position, tokenIndex, depth
								if buffer[position] != rune('i') {
									goto l348
								}
								position++
								goto l347
							l348:
								position, tokenIndex, depth = position347, tokenIndex347, depth347
								if buffer[position] != rune('I') {
									goto l316
								}
								position++
							}
						l347:
							{
								position349, tokenIndex349, depth349 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l350
								}
								position++
								goto l349
							l350:
								position, tokenIndex, depth = position349, tokenIndex349, depth349
								if buffer[position] != rune('N') {
									goto l316
								}
								position++
							}
						l349:
							if !_rules[ruleKEY]() {
								goto l316
							}
							{
								position351 := position
								depth++
								{
									add(ruleAction48, position)
								}
								if !_rules[rule_]() {
									goto l316
								}
								if !_rules[rulePAREN_OPEN]() {
									goto l316
								}
								if !_rules[ruleliteralListString]() {
									goto l316
								}
							l353:
								{
									position354, tokenIndex354, depth354 := position, tokenIndex, depth
									if !_rules[rule_]() {
										goto l354
									}
									if !_rules[ruleCOMMA]() {
										goto l354
									}
									if !_rules[ruleliteralListString]() {
										goto l354
									}
									goto l353
								l354:
									position, tokenIndex, depth = position354, tokenIndex354, depth354
								}
								if !_rules[rule_]() {
									goto l316
								}
								if !_rules[rulePAREN_CLOSE]() {
									goto l316
								}
								depth--
								add(ruleliteralList, position351)
							}
							{
								add(ruleAction46, position)
							}
						}
					l330:
						depth--
						add(ruletagMatcher, position329)
					}
				}
			l318:
				depth--
				add(rulepredicate_3, position317)
			}
			return true
		l316:
			position, tokenIndex, depth = position316, tokenIndex316, depth316
			return false
		},
		/* 29 tagMatcher <- <((tagName _ '=' literalString Action43) / (tagName _ ('!' '=') literalString Action44) / (tagName _ (('m' / 'M') ('a' / 'A') ('t' / 'T') ('c' / 'C') ('h' / 'H')) KEY literalString Action45) / (tagName _ (('i' / 'I') ('n' / 'N')) KEY literalList Action46))> */
		nil,
		/* 30 literalString <- <(_ STRING Action47)> */
		func() bool {
			position357, tokenIndex357, depth357 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position358 := position
				depth++
				if !_rules[rule_]() {
					goto l357
					// @@ &_rules escapes to heap
				}
				if !_rules[ruleSTRING]() {
					goto l357
				}
				{
					add(ruleAction47, position)
				}
				depth--
				add(ruleliteralString, position358)
			}
			return true
		l357:
			position, tokenIndex, depth = position357, tokenIndex357, depth357
			return false
		},
		/* 31 literalList <- <(Action48 _ PAREN_OPEN literalListString (_ COMMA literalListString)* _ PAREN_CLOSE)> */
		nil,
		/* 32 literalListString <- <(_ STRING Action49)> */
		func() bool {
			position361, tokenIndex361, depth361 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position362 := position
				depth++
				if !_rules[rule_]() {
					goto l361
					// @@ &_rules escapes to heap
				}
				if !_rules[ruleSTRING]() {
					goto l361
				}
				{
					add(ruleAction49, position)
				}
				depth--
				add(ruleliteralListString, position362)
			}
			return true
		l361:
			position, tokenIndex, depth = position361, tokenIndex361, depth361
			return false
		},
		/* 33 tagName <- <(_ <TAG_NAME> Action50)> */
		func() bool {
			position364, tokenIndex364, depth364 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position365 := position
				depth++
				if !_rules[rule_]() {
					goto l364
					// @@ &_rules escapes to heap
				}
				{
					position366 := position
					depth++
					{
						position367 := position
						depth++
						if !_rules[ruleIDENTIFIER]() {
							goto l364
						}
						depth--
						add(ruleTAG_NAME, position367)
					}
					depth--
					add(rulePegText, position366)
				}
				{
					add(ruleAction50, position)
				}
				depth--
				add(ruletagName, position365)
			}
			return true
		l364:
			position, tokenIndex, depth = position364, tokenIndex364, depth364
			return false
		},
		/* 34 COLUMN_NAME <- <IDENTIFIER> */
		func() bool {
			position369, tokenIndex369, depth369 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position370 := position
				depth++
				if !_rules[ruleIDENTIFIER]() {
					goto l369
					// @@ &_rules escapes to heap
				}
				depth--
				add(ruleCOLUMN_NAME, position370)
			}
			return true
		l369:
			position, tokenIndex, depth = position369, tokenIndex369, depth369
			return false
		},
		/* 35 METRIC_NAME <- <IDENTIFIER> */
		nil,
		/* 36 TAG_NAME <- <IDENTIFIER> */
		nil,
		/* 37 IDENTIFIER <- <(('`' CHAR* '`') / (_ !(KEYWORD KEY) ID_SEGMENT ('.' ID_SEGMENT)*))> */
		func() bool {
			position373, tokenIndex373, depth373 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position374 := position
				depth++
				{
					position375, tokenIndex375, depth375 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l376
					}
					position++
				l377:
					{
						position378, tokenIndex378, depth378 := position, tokenIndex, depth
						if !_rules[ruleCHAR]() {
							goto l378
							// @@ &_rules escapes to heap
						}
						goto l377
					l378:
						position, tokenIndex, depth = position378, tokenIndex378, depth378
					}
					if buffer[position] != rune('`') {
						goto l376
					}
					position++
					goto l375
				l376:
					position, tokenIndex, depth = position375, tokenIndex375, depth375
					if !_rules[rule_]() {
						goto l373
					}
					{
						position379, tokenIndex379, depth379 := position, tokenIndex, depth
						{
							position380 := position
							depth++
							{
								position381, tokenIndex381, depth381 := position, tokenIndex, depth
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
									if buffer[position] != rune('l') {
										goto l386
									}
									position++
									goto l385
								l386:
									position, tokenIndex, depth = position385, tokenIndex385, depth385
									if buffer[position] != rune('L') {
										goto l382
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
										goto l382
									}
									position++
								}
							l387:
								goto l381
							l382:
								position, tokenIndex, depth = position381, tokenIndex381, depth381
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
										goto l389
									}
									position++
								}
							l390:
								{
									position392, tokenIndex392, depth392 := position, tokenIndex, depth
									if buffer[position] != rune('n') {
										goto l393
									}
									position++
									goto l392
								l393:
									position, tokenIndex, depth = position392, tokenIndex392, depth392
									if buffer[position] != rune('N') {
										goto l389
									}
									position++
								}
							l392:
								{
									position394, tokenIndex394, depth394 := position, tokenIndex, depth
									if buffer[position] != rune('d') {
										goto l395
									}
									position++
									goto l394
								l395:
									position, tokenIndex, depth = position394, tokenIndex394, depth394
									if buffer[position] != rune('D') {
										goto l389
									}
									position++
								}
							l394:
								goto l381
							l389:
								position, tokenIndex, depth = position381, tokenIndex381, depth381
								{
									position397, tokenIndex397, depth397 := position, tokenIndex, depth
									if buffer[position] != rune('m') {
										goto l398
									}
									position++
									goto l397
								l398:
									position, tokenIndex, depth = position397, tokenIndex397, depth397
									if buffer[position] != rune('M') {
										goto l396
									}
									position++
								}
							l397:
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
										goto l396
									}
									position++
								}
							l399:
								{
									position401, tokenIndex401, depth401 := position, tokenIndex, depth
									if buffer[position] != rune('t') {
										goto l402
									}
									position++
									goto l401
								l402:
									position, tokenIndex, depth = position401, tokenIndex401, depth401
									if buffer[position] != rune('T') {
										goto l396
									}
									position++
								}
							l401:
								{
									position403, tokenIndex403, depth403 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l404
									}
									position++
									goto l403
								l404:
									position, tokenIndex, depth = position403, tokenIndex403, depth403
									if buffer[position] != rune('C') {
										goto l396
									}
									position++
								}
							l403:
								{
									position405, tokenIndex405, depth405 := position, tokenIndex, depth
									if buffer[position] != rune('h') {
										goto l406
									}
									position++
									goto l405
								l406:
									position, tokenIndex, depth = position405, tokenIndex405, depth405
									if buffer[position] != rune('H') {
										goto l396
									}
									position++
								}
							l405:
								goto l381
							l396:
								position, tokenIndex, depth = position381, tokenIndex381, depth381
								{
									position408, tokenIndex408, depth408 := position, tokenIndex, depth
									if buffer[position] != rune('s') {
										goto l409
									}
									position++
									goto l408
								l409:
									position, tokenIndex, depth = position408, tokenIndex408, depth408
									if buffer[position] != rune('S') {
										goto l407
									}
									position++
								}
							l408:
								{
									position410, tokenIndex410, depth410 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l411
									}
									position++
									goto l410
								l411:
									position, tokenIndex, depth = position410, tokenIndex410, depth410
									if buffer[position] != rune('E') {
										goto l407
									}
									position++
								}
							l410:
								{
									position412, tokenIndex412, depth412 := position, tokenIndex, depth
									if buffer[position] != rune('l') {
										goto l413
									}
									position++
									goto l412
								l413:
									position, tokenIndex, depth = position412, tokenIndex412, depth412
									if buffer[position] != rune('L') {
										goto l407
									}
									position++
								}
							l412:
								{
									position414, tokenIndex414, depth414 := position, tokenIndex, depth
									if buffer[position] != rune('e') {
										goto l415
									}
									position++
									goto l414
								l415:
									position, tokenIndex, depth = position414, tokenIndex414, depth414
									if buffer[position] != rune('E') {
										goto l407
									}
									position++
								}
							l414:
								{
									position416, tokenIndex416, depth416 := position, tokenIndex, depth
									if buffer[position] != rune('c') {
										goto l417
									}
									position++
									goto l416
								l417:
									position, tokenIndex, depth = position416, tokenIndex416, depth416
									if buffer[position] != rune('C') {
										goto l407
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
										goto l407
									}
									position++
								}
							l418:
								goto l381
							l407:
								position, tokenIndex, depth = position381, tokenIndex381, depth381
								{
									switch buffer[position] {
									case 'M', 'm':
										{
											position421, tokenIndex421, depth421 := position, tokenIndex, depth
											if buffer[position] != rune('m') {
												goto l422
											}
											position++
											goto l421
										l422:
											position, tokenIndex, depth = position421, tokenIndex421, depth421
											if buffer[position] != rune('M') {
												goto l379
											}
											position++
										}
									l421:
										{
											position423, tokenIndex423, depth423 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l424
											}
											position++
											goto l423
										l424:
											position, tokenIndex, depth = position423, tokenIndex423, depth423
											if buffer[position] != rune('E') {
												goto l379
											}
											position++
										}
									l423:
										{
											position425, tokenIndex425, depth425 := position, tokenIndex, depth
											if buffer[position] != rune('t') {
												goto l426
											}
											position++
											goto l425
										l426:
											position, tokenIndex, depth = position425, tokenIndex425, depth425
											if buffer[position] != rune('T') {
												goto l379
											}
											position++
										}
									l425:
										{
											position427, tokenIndex427, depth427 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l428
											}
											position++
											goto l427
										l428:
											position, tokenIndex, depth = position427, tokenIndex427, depth427
											if buffer[position] != rune('R') {
												goto l379
											}
											position++
										}
									l427:
										{
											position429, tokenIndex429, depth429 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l430
											}
											position++
											goto l429
										l430:
											position, tokenIndex, depth = position429, tokenIndex429, depth429
											if buffer[position] != rune('I') {
												goto l379
											}
											position++
										}
									l429:
										{
											position431, tokenIndex431, depth431 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l432
											}
											position++
											goto l431
										l432:
											position, tokenIndex, depth = position431, tokenIndex431, depth431
											if buffer[position] != rune('C') {
												goto l379
											}
											position++
										}
									l431:
										{
											position433, tokenIndex433, depth433 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l434
											}
											position++
											goto l433
										l434:
											position, tokenIndex, depth = position433, tokenIndex433, depth433
											if buffer[position] != rune('S') {
												goto l379
											}
											position++
										}
									l433:
										break
									case 'W', 'w':
										{
											position435, tokenIndex435, depth435 := position, tokenIndex, depth
											if buffer[position] != rune('w') {
												goto l436
											}
											position++
											goto l435
										l436:
											position, tokenIndex, depth = position435, tokenIndex435, depth435
											if buffer[position] != rune('W') {
												goto l379
											}
											position++
										}
									l435:
										{
											position437, tokenIndex437, depth437 := position, tokenIndex, depth
											if buffer[position] != rune('h') {
												goto l438
											}
											position++
											goto l437
										l438:
											position, tokenIndex, depth = position437, tokenIndex437, depth437
											if buffer[position] != rune('H') {
												goto l379
											}
											position++
										}
									l437:
										{
											position439, tokenIndex439, depth439 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l440
											}
											position++
											goto l439
										l440:
											position, tokenIndex, depth = position439, tokenIndex439, depth439
											if buffer[position] != rune('E') {
												goto l379
											}
											position++
										}
									l439:
										{
											position441, tokenIndex441, depth441 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l442
											}
											position++
											goto l441
										l442:
											position, tokenIndex, depth = position441, tokenIndex441, depth441
											if buffer[position] != rune('R') {
												goto l379
											}
											position++
										}
									l441:
										{
											position443, tokenIndex443, depth443 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l444
											}
											position++
											goto l443
										l444:
											position, tokenIndex, depth = position443, tokenIndex443, depth443
											if buffer[position] != rune('E') {
												goto l379
											}
											position++
										}
									l443:
										break
									case 'O', 'o':
										{
											position445, tokenIndex445, depth445 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l446
											}
											position++
											goto l445
										l446:
											position, tokenIndex, depth = position445, tokenIndex445, depth445
											if buffer[position] != rune('O') {
												goto l379
											}
											position++
										}
									l445:
										{
											position447, tokenIndex447, depth447 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l448
											}
											position++
											goto l447
										l448:
											position, tokenIndex, depth = position447, tokenIndex447, depth447
											if buffer[position] != rune('R') {
												goto l379
											}
											position++
										}
									l447:
										break
									case 'N', 'n':
										{
											position449, tokenIndex449, depth449 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l450
											}
											position++
											goto l449
										l450:
											position, tokenIndex, depth = position449, tokenIndex449, depth449
											if buffer[position] != rune('N') {
												goto l379
											}
											position++
										}
									l449:
										{
											position451, tokenIndex451, depth451 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l452
											}
											position++
											goto l451
										l452:
											position, tokenIndex, depth = position451, tokenIndex451, depth451
											if buffer[position] != rune('O') {
												goto l379
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
												goto l379
											}
											position++
										}
									l453:
										break
									case 'I', 'i':
										{
											position455, tokenIndex455, depth455 := position, tokenIndex, depth
											if buffer[position] != rune('i') {
												goto l456
											}
											position++
											goto l455
										l456:
											position, tokenIndex, depth = position455, tokenIndex455, depth455
											if buffer[position] != rune('I') {
												goto l379
											}
											position++
										}
									l455:
										{
											position457, tokenIndex457, depth457 := position, tokenIndex, depth
											if buffer[position] != rune('n') {
												goto l458
											}
											position++
											goto l457
										l458:
											position, tokenIndex, depth = position457, tokenIndex457, depth457
											if buffer[position] != rune('N') {
												goto l379
											}
											position++
										}
									l457:
										break
									case 'C', 'c':
										{
											position459, tokenIndex459, depth459 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l460
											}
											position++
											goto l459
										l460:
											position, tokenIndex, depth = position459, tokenIndex459, depth459
											if buffer[position] != rune('C') {
												goto l379
											}
											position++
										}
									l459:
										{
											position461, tokenIndex461, depth461 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l462
											}
											position++
											goto l461
										l462:
											position, tokenIndex, depth = position461, tokenIndex461, depth461
											if buffer[position] != rune('O') {
												goto l379
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
												goto l379
											}
											position++
										}
									l463:
										{
											position465, tokenIndex465, depth465 := position, tokenIndex, depth
											if buffer[position] != rune('l') {
												goto l466
											}
											position++
											goto l465
										l466:
											position, tokenIndex, depth = position465, tokenIndex465, depth465
											if buffer[position] != rune('L') {
												goto l379
											}
											position++
										}
									l465:
										{
											position467, tokenIndex467, depth467 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l468
											}
											position++
											goto l467
										l468:
											position, tokenIndex, depth = position467, tokenIndex467, depth467
											if buffer[position] != rune('A') {
												goto l379
											}
											position++
										}
									l467:
										{
											position469, tokenIndex469, depth469 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l470
											}
											position++
											goto l469
										l470:
											position, tokenIndex, depth = position469, tokenIndex469, depth469
											if buffer[position] != rune('P') {
												goto l379
											}
											position++
										}
									l469:
										{
											position471, tokenIndex471, depth471 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l472
											}
											position++
											goto l471
										l472:
											position, tokenIndex, depth = position471, tokenIndex471, depth471
											if buffer[position] != rune('S') {
												goto l379
											}
											position++
										}
									l471:
										{
											position473, tokenIndex473, depth473 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l474
											}
											position++
											goto l473
										l474:
											position, tokenIndex, depth = position473, tokenIndex473, depth473
											if buffer[position] != rune('E') {
												goto l379
											}
											position++
										}
									l473:
										break
									case 'G', 'g':
										{
											position475, tokenIndex475, depth475 := position, tokenIndex, depth
											if buffer[position] != rune('g') {
												goto l476
											}
											position++
											goto l475
										l476:
											position, tokenIndex, depth = position475, tokenIndex475, depth475
											if buffer[position] != rune('G') {
												goto l379
											}
											position++
										}
									l475:
										{
											position477, tokenIndex477, depth477 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l478
											}
											position++
											goto l477
										l478:
											position, tokenIndex, depth = position477, tokenIndex477, depth477
											if buffer[position] != rune('R') {
												goto l379
											}
											position++
										}
									l477:
										{
											position479, tokenIndex479, depth479 := position, tokenIndex, depth
											if buffer[position] != rune('o') {
												goto l480
											}
											position++
											goto l479
										l480:
											position, tokenIndex, depth = position479, tokenIndex479, depth479
											if buffer[position] != rune('O') {
												goto l379
											}
											position++
										}
									l479:
										{
											position481, tokenIndex481, depth481 := position, tokenIndex, depth
											if buffer[position] != rune('u') {
												goto l482
											}
											position++
											goto l481
										l482:
											position, tokenIndex, depth = position481, tokenIndex481, depth481
											if buffer[position] != rune('U') {
												goto l379
											}
											position++
										}
									l481:
										{
											position483, tokenIndex483, depth483 := position, tokenIndex, depth
											if buffer[position] != rune('p') {
												goto l484
											}
											position++
											goto l483
										l484:
											position, tokenIndex, depth = position483, tokenIndex483, depth483
											if buffer[position] != rune('P') {
												goto l379
											}
											position++
										}
									l483:
										break
									case 'D', 'd':
										{
											position485, tokenIndex485, depth485 := position, tokenIndex, depth
											if buffer[position] != rune('d') {
												goto l486
											}
											position++
											goto l485
										l486:
											position, tokenIndex, depth = position485, tokenIndex485, depth485
											if buffer[position] != rune('D') {
												goto l379
											}
											position++
										}
									l485:
										{
											position487, tokenIndex487, depth487 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l488
											}
											position++
											goto l487
										l488:
											position, tokenIndex, depth = position487, tokenIndex487, depth487
											if buffer[position] != rune('E') {
												goto l379
											}
											position++
										}
									l487:
										{
											position489, tokenIndex489, depth489 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l490
											}
											position++
											goto l489
										l490:
											position, tokenIndex, depth = position489, tokenIndex489, depth489
											if buffer[position] != rune('S') {
												goto l379
											}
											position++
										}
									l489:
										{
											position491, tokenIndex491, depth491 := position, tokenIndex, depth
											if buffer[position] != rune('c') {
												goto l492
											}
											position++
											goto l491
										l492:
											position, tokenIndex, depth = position491, tokenIndex491, depth491
											if buffer[position] != rune('C') {
												goto l379
											}
											position++
										}
									l491:
										{
											position493, tokenIndex493, depth493 := position, tokenIndex, depth
											if buffer[position] != rune('r') {
												goto l494
											}
											position++
											goto l493
										l494:
											position, tokenIndex, depth = position493, tokenIndex493, depth493
											if buffer[position] != rune('R') {
												goto l379
											}
											position++
										}
									l493:
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
												goto l379
											}
											position++
										}
									l495:
										{
											position497, tokenIndex497, depth497 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l498
											}
											position++
											goto l497
										l498:
											position, tokenIndex, depth = position497, tokenIndex497, depth497
											if buffer[position] != rune('B') {
												goto l379
											}
											position++
										}
									l497:
										{
											position499, tokenIndex499, depth499 := position, tokenIndex, depth
											if buffer[position] != rune('e') {
												goto l500
											}
											position++
											goto l499
										l500:
											position, tokenIndex, depth = position499, tokenIndex499, depth499
											if buffer[position] != rune('E') {
												goto l379
											}
											position++
										}
									l499:
										break
									case 'B', 'b':
										{
											position501, tokenIndex501, depth501 := position, tokenIndex, depth
											if buffer[position] != rune('b') {
												goto l502
											}
											position++
											goto l501
										l502:
											position, tokenIndex, depth = position501, tokenIndex501, depth501
											if buffer[position] != rune('B') {
												goto l379
											}
											position++
										}
									l501:
										{
											position503, tokenIndex503, depth503 := position, tokenIndex, depth
											if buffer[position] != rune('y') {
												goto l504
											}
											position++
											goto l503
										l504:
											position, tokenIndex, depth = position503, tokenIndex503, depth503
											if buffer[position] != rune('Y') {
												goto l379
											}
											position++
										}
									l503:
										break
									case 'A', 'a':
										{
											position505, tokenIndex505, depth505 := position, tokenIndex, depth
											if buffer[position] != rune('a') {
												goto l506
											}
											position++
											goto l505
										l506:
											position, tokenIndex, depth = position505, tokenIndex505, depth505
											if buffer[position] != rune('A') {
												goto l379
											}
											position++
										}
									l505:
										{
											position507, tokenIndex507, depth507 := position, tokenIndex, depth
											if buffer[position] != rune('s') {
												goto l508
											}
											position++
											goto l507
										l508:
											position, tokenIndex, depth = position507, tokenIndex507, depth507
											if buffer[position] != rune('S') {
												goto l379
											}
											position++
										}
									l507:
										break
									default:
										if !_rules[rulePROPERTY_KEY]() {
											goto l379
										}
										break
									}
								}

							}
						l381:
							depth--
							add(ruleKEYWORD, position380)
						}
						if !_rules[ruleKEY]() {
							goto l379
						}
						goto l373
					l379:
						position, tokenIndex, depth = position379, tokenIndex379, depth379
					}
					if !_rules[ruleID_SEGMENT]() {
						goto l373
					}
				l509:
					{
						position510, tokenIndex510, depth510 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l510
						}
						position++
						if !_rules[ruleID_SEGMENT]() {
							goto l510
						}
						goto l509
					l510:
						position, tokenIndex, depth = position510, tokenIndex510, depth510
					}
				}
			l375:
				depth--
				add(ruleIDENTIFIER, position374)
			}
			return true
		l373:
			position, tokenIndex, depth = position373, tokenIndex373, depth373
			return false
		},
		/* 38 TIMESTAMP <- <((_ <(NUMBER ([a-z] / [A-Z])*)>) / (_ STRING) / (_ <(('n' / 'N') ('o' / 'O') ('w' / 'W'))>))> */
		nil,
		/* 39 ID_SEGMENT <- <(_ ID_START ID_CONT*)> */
		func() bool {
			position512, tokenIndex512, depth512 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position513 := position
				depth++
				if !_rules[rule_]() {
					goto l512
					// @@ &_rules escapes to heap
				}
				if !_rules[ruleID_START]() {
					goto l512
				}
			l514:
				{
					position515, tokenIndex515, depth515 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l515
					}
					goto l514
				l515:
					position, tokenIndex, depth = position515, tokenIndex515, depth515
				}
				depth--
				add(ruleID_SEGMENT, position513)
			}
			return true
		l512:
			position, tokenIndex, depth = position512, tokenIndex512, depth512
			return false
		},
		/* 40 ID_START <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position516, tokenIndex516, depth516 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position517 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l516
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l516
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l516
						}
						position++
						break
					}
				}

				depth--
				add(ruleID_START, position517)
			}
			return true
		l516:
			position, tokenIndex, depth = position516, tokenIndex516, depth516
			return false
		},
		/* 41 ID_CONT <- <(ID_START / [0-9])> */
		func() bool {
			position519, tokenIndex519, depth519 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position520 := position
				depth++
				{
					position521, tokenIndex521, depth521 := position, tokenIndex, depth
					if !_rules[ruleID_START]() {
						goto l522
						// @@ &_rules escapes to heap
					}
					goto l521
				l522:
					position, tokenIndex, depth = position521, tokenIndex521, depth521
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l519
					}
					position++
				}
			l521:
				depth--
				add(ruleID_CONT, position520)
			}
			return true
		l519:
			position, tokenIndex, depth = position519, tokenIndex519, depth519
			return false
		},
		/* 42 PROPERTY_KEY <- <(((&('S' | 's') (<(('s' / 'S') ('a' / 'A') ('m' / 'M') ('p' / 'P') ('l' / 'L') ('e' / 'E'))> KEY _ (('b' / 'B') ('y' / 'Y')))) | (&('R' | 'r') <(('r' / 'R') ('e' / 'E') ('s' / 'S') ('o' / 'O') ('l' / 'L') ('u' / 'U') ('t' / 'T') ('i' / 'I') ('o' / 'O') ('n' / 'N'))>) | (&('T' | 't') <(('t' / 'T') ('o' / 'O'))>) | (&('F' | 'f') <(('f' / 'F') ('r' / 'R') ('o' / 'O') ('m' / 'M'))>)) KEY)> */
		func() bool {
			position523, tokenIndex523, depth523 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position524 := position
				depth++
				{
					switch buffer[position] {
					case 'S', 's':
						{
							position526 := position
							depth++
							{
								position527, tokenIndex527, depth527 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l528
								}
								position++
								goto l527
							l528:
								position, tokenIndex, depth = position527, tokenIndex527, depth527
								if buffer[position] != rune('S') {
									goto l523
								}
								position++
							}
						l527:
							{
								position529, tokenIndex529, depth529 := position, tokenIndex, depth
								if buffer[position] != rune('a') {
									goto l530
								}
								position++
								goto l529
							l530:
								position, tokenIndex, depth = position529, tokenIndex529, depth529
								if buffer[position] != rune('A') {
									goto l523
								}
								position++
							}
						l529:
							{
								position531, tokenIndex531, depth531 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l532
								}
								position++
								goto l531
							l532:
								position, tokenIndex, depth = position531, tokenIndex531, depth531
								if buffer[position] != rune('M') {
									goto l523
								}
								position++
							}
						l531:
							{
								position533, tokenIndex533, depth533 := position, tokenIndex, depth
								if buffer[position] != rune('p') {
									goto l534
								}
								position++
								goto l533
							l534:
								position, tokenIndex, depth = position533, tokenIndex533, depth533
								if buffer[position] != rune('P') {
									goto l523
								}
								position++
							}
						l533:
							{
								position535, tokenIndex535, depth535 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l536
								}
								position++
								goto l535
							l536:
								position, tokenIndex, depth = position535, tokenIndex535, depth535
								if buffer[position] != rune('L') {
									goto l523
								}
								position++
							}
						l535:
							{
								position537, tokenIndex537, depth537 := position, tokenIndex, depth
								if buffer[position] != rune('e') {
									goto l538
								}
								position++
								goto l537
							l538:
								position, tokenIndex, depth = position537, tokenIndex537, depth537
								if buffer[position] != rune('E') {
									goto l523
								}
								position++
							}
						l537:
							depth--
							add(rulePegText, position526)
						}
						if !_rules[ruleKEY]() {
							goto l523
							// @@ &_rules escapes to heap
						}
						if !_rules[rule_]() {
							goto l523
						}
						{
							position539, tokenIndex539, depth539 := position, tokenIndex, depth
							if buffer[position] != rune('b') {
								goto l540
							}
							position++
							goto l539
						l540:
							position, tokenIndex, depth = position539, tokenIndex539, depth539
							if buffer[position] != rune('B') {
								goto l523
							}
							position++
						}
					l539:
						{
							position541, tokenIndex541, depth541 := position, tokenIndex, depth
							if buffer[position] != rune('y') {
								goto l542
							}
							position++
							goto l541
						l542:
							position, tokenIndex, depth = position541, tokenIndex541, depth541
							if buffer[position] != rune('Y') {
								goto l523
							}
							position++
						}
					l541:
						break
					case 'R', 'r':
						{
							position543 := position
							depth++
							{
								position544, tokenIndex544, depth544 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l545
								}
								position++
								goto l544
							l545:
								position, tokenIndex, depth = position544, tokenIndex544, depth544
								if buffer[position] != rune('R') {
									goto l523
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
									goto l523
								}
								position++
							}
						l546:
							{
								position548, tokenIndex548, depth548 := position, tokenIndex, depth
								if buffer[position] != rune('s') {
									goto l549
								}
								position++
								goto l548
							l549:
								position, tokenIndex, depth = position548, tokenIndex548, depth548
								if buffer[position] != rune('S') {
									goto l523
								}
								position++
							}
						l548:
							{
								position550, tokenIndex550, depth550 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l551
								}
								position++
								goto l550
							l551:
								position, tokenIndex, depth = position550, tokenIndex550, depth550
								if buffer[position] != rune('O') {
									goto l523
								}
								position++
							}
						l550:
							{
								position552, tokenIndex552, depth552 := position, tokenIndex, depth
								if buffer[position] != rune('l') {
									goto l553
								}
								position++
								goto l552
							l553:
								position, tokenIndex, depth = position552, tokenIndex552, depth552
								if buffer[position] != rune('L') {
									goto l523
								}
								position++
							}
						l552:
							{
								position554, tokenIndex554, depth554 := position, tokenIndex, depth
								if buffer[position] != rune('u') {
									goto l555
								}
								position++
								goto l554
							l555:
								position, tokenIndex, depth = position554, tokenIndex554, depth554
								if buffer[position] != rune('U') {
									goto l523
								}
								position++
							}
						l554:
							{
								position556, tokenIndex556, depth556 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l557
								}
								position++
								goto l556
							l557:
								position, tokenIndex, depth = position556, tokenIndex556, depth556
								if buffer[position] != rune('T') {
									goto l523
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
									goto l523
								}
								position++
							}
						l558:
							{
								position560, tokenIndex560, depth560 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l561
								}
								position++
								goto l560
							l561:
								position, tokenIndex, depth = position560, tokenIndex560, depth560
								if buffer[position] != rune('O') {
									goto l523
								}
								position++
							}
						l560:
							{
								position562, tokenIndex562, depth562 := position, tokenIndex, depth
								if buffer[position] != rune('n') {
									goto l563
								}
								position++
								goto l562
							l563:
								position, tokenIndex, depth = position562, tokenIndex562, depth562
								if buffer[position] != rune('N') {
									goto l523
								}
								position++
							}
						l562:
							depth--
							add(rulePegText, position543)
						}
						break
					case 'T', 't':
						{
							position564 := position
							depth++
							{
								position565, tokenIndex565, depth565 := position, tokenIndex, depth
								if buffer[position] != rune('t') {
									goto l566
								}
								position++
								goto l565
							l566:
								position, tokenIndex, depth = position565, tokenIndex565, depth565
								if buffer[position] != rune('T') {
									goto l523
								}
								position++
							}
						l565:
							{
								position567, tokenIndex567, depth567 := position, tokenIndex, depth
								if buffer[position] != rune('o') {
									goto l568
								}
								position++
								goto l567
							l568:
								position, tokenIndex, depth = position567, tokenIndex567, depth567
								if buffer[position] != rune('O') {
									goto l523
								}
								position++
							}
						l567:
							depth--
							add(rulePegText, position564)
						}
						break
					default:
						{
							position569 := position
							depth++
							{
								position570, tokenIndex570, depth570 := position, tokenIndex, depth
								if buffer[position] != rune('f') {
									goto l571
								}
								position++
								goto l570
							l571:
								position, tokenIndex, depth = position570, tokenIndex570, depth570
								if buffer[position] != rune('F') {
									goto l523
								}
								position++
							}
						l570:
							{
								position572, tokenIndex572, depth572 := position, tokenIndex, depth
								if buffer[position] != rune('r') {
									goto l573
								}
								position++
								goto l572
							l573:
								position, tokenIndex, depth = position572, tokenIndex572, depth572
								if buffer[position] != rune('R') {
									goto l523
								}
								position++
							}
						l572:
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
									goto l523
								}
								position++
							}
						l574:
							{
								position576, tokenIndex576, depth576 := position, tokenIndex, depth
								if buffer[position] != rune('m') {
									goto l577
								}
								position++
								goto l576
							l577:
								position, tokenIndex, depth = position576, tokenIndex576, depth576
								if buffer[position] != rune('M') {
									goto l523
								}
								position++
							}
						l576:
							depth--
							add(rulePegText, position569)
						}
						break
					}
				}

				if !_rules[ruleKEY]() {
					goto l523
				}
				depth--
				add(rulePROPERTY_KEY, position524)
			}
			return true
		l523:
			position, tokenIndex, depth = position523, tokenIndex523, depth523
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
			position588, tokenIndex588, depth588 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position589 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l588
				}
				position++
				depth--
				add(ruleQUOTE_SINGLE, position589)
			}
			return true
		l588:
			position, tokenIndex, depth = position588, tokenIndex588, depth588
			return false
		},
		/* 54 QUOTE_DOUBLE <- <'"'> */
		func() bool {
			position590, tokenIndex590, depth590 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position591 := position
				depth++
				if buffer[position] != rune('"') {
					goto l590
				}
				position++
				depth--
				add(ruleQUOTE_DOUBLE, position591)
			}
			return true
		l590:
			position, tokenIndex, depth = position590, tokenIndex590, depth590
			return false
		},
		/* 55 STRING <- <((QUOTE_SINGLE <(!QUOTE_SINGLE CHAR)*> QUOTE_SINGLE) / (QUOTE_DOUBLE <(!QUOTE_DOUBLE CHAR)*> QUOTE_DOUBLE))> */
		func() bool {
			position592, tokenIndex592, depth592 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position593 := position
				depth++
				{
					position594, tokenIndex594, depth594 := position, tokenIndex, depth
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l595
						// @@ &_rules escapes to heap
					}
					{
						position596 := position
						depth++
					l597:
						{
							position598, tokenIndex598, depth598 := position, tokenIndex, depth
							{
								position599, tokenIndex599, depth599 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_SINGLE]() {
									goto l599
								}
								goto l598
							l599:
								position, tokenIndex, depth = position599, tokenIndex599, depth599
							}
							if !_rules[ruleCHAR]() {
								goto l598
							}
							goto l597
						l598:
							position, tokenIndex, depth = position598, tokenIndex598, depth598
						}
						depth--
						add(rulePegText, position596)
					}
					if !_rules[ruleQUOTE_SINGLE]() {
						goto l595
					}
					goto l594
				l595:
					position, tokenIndex, depth = position594, tokenIndex594, depth594
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l592
					}
					{
						position600 := position
						depth++
					l601:
						{
							position602, tokenIndex602, depth602 := position, tokenIndex, depth
							{
								position603, tokenIndex603, depth603 := position, tokenIndex, depth
								if !_rules[ruleQUOTE_DOUBLE]() {
									goto l603
								}
								goto l602
							l603:
								position, tokenIndex, depth = position603, tokenIndex603, depth603
							}
							if !_rules[ruleCHAR]() {
								goto l602
							}
							goto l601
						l602:
							position, tokenIndex, depth = position602, tokenIndex602, depth602
						}
						depth--
						add(rulePegText, position600)
					}
					if !_rules[ruleQUOTE_DOUBLE]() {
						goto l592
					}
				}
			l594:
				depth--
				add(ruleSTRING, position593)
			}
			return true
		l592:
			position, tokenIndex, depth = position592, tokenIndex592, depth592
			return false
		},
		/* 56 CHAR <- <(('\\' ((&('"') QUOTE_DOUBLE) | (&('\'') QUOTE_SINGLE) | (&('\\' | '`') ESCAPE_CLASS))) / (!ESCAPE_CLASS .))> */
		func() bool {
			position604, tokenIndex604, depth604 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position605 := position
				depth++
				{
					position606, tokenIndex606, depth606 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l607
					}
					position++
					{
						switch buffer[position] {
						case '"':
							if !_rules[ruleQUOTE_DOUBLE]() {
								goto l607
								// @@ &_rules escapes to heap
							}
							break
						case '\'':
							if !_rules[ruleQUOTE_SINGLE]() {
								goto l607
							}
							break
						default:
							if !_rules[ruleESCAPE_CLASS]() {
								goto l607
							}
							break
						}
					}

					goto l606
				l607:
					position, tokenIndex, depth = position606, tokenIndex606, depth606
					{
						position609, tokenIndex609, depth609 := position, tokenIndex, depth
						if !_rules[ruleESCAPE_CLASS]() {
							goto l609
						}
						goto l604
					l609:
						position, tokenIndex, depth = position609, tokenIndex609, depth609
					}
					if !matchDot() {
						goto l604
					}
				}
			l606:
				depth--
				add(ruleCHAR, position605)
			}
			return true
		l604:
			position, tokenIndex, depth = position604, tokenIndex604, depth604
			return false
		},
		/* 57 ESCAPE_CLASS <- <('`' / '\\')> */
		func() bool {
			position610, tokenIndex610, depth610 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position611 := position
				depth++
				{
					position612, tokenIndex612, depth612 := position, tokenIndex, depth
					if buffer[position] != rune('`') {
						goto l613
					}
					position++
					goto l612
				l613:
					position, tokenIndex, depth = position612, tokenIndex612, depth612
					if buffer[position] != rune('\\') {
						goto l610
					}
					position++
				}
			l612:
				depth--
				add(ruleESCAPE_CLASS, position611)
			}
			return true
		l610:
			position, tokenIndex, depth = position610, tokenIndex610, depth610
			return false
		},
		/* 58 NUMBER <- <(NUMBER_INTEGER NUMBER_FRACTION? NUMBER_EXP?)> */
		func() bool {
			position614, tokenIndex614, depth614 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position615 := position
				depth++
				{
					position616 := position
					depth++
					{
						position617, tokenIndex617, depth617 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l617
						}
						position++
						goto l618
					l617:
						position, tokenIndex, depth = position617, tokenIndex617, depth617
					}
				l618:
					{
						position619 := position
						depth++
						{
							position620, tokenIndex620, depth620 := position, tokenIndex, depth
							if buffer[position] != rune('0') {
								goto l621
							}
							position++
							goto l620
						l621:
							position, tokenIndex, depth = position620, tokenIndex620, depth620
							if c := buffer[position]; c < rune('1') || c > rune('9') {
								goto l614
							}
							position++
						l622:
							{
								position623, tokenIndex623, depth623 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l623
								}
								position++
								goto l622
							l623:
								position, tokenIndex, depth = position623, tokenIndex623, depth623
							}
						}
					l620:
						depth--
						add(ruleNUMBER_NATURAL, position619)
					}
					depth--
					add(ruleNUMBER_INTEGER, position616)
				}
				{
					position624, tokenIndex624, depth624 := position, tokenIndex, depth
					{
						position626 := position
						depth++
						if buffer[position] != rune('.') {
							goto l624
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l624
						}
						position++
					l627:
						{
							position628, tokenIndex628, depth628 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l628
							}
							position++
							goto l627
						l628:
							position, tokenIndex, depth = position628, tokenIndex628, depth628
						}
						depth--
						add(ruleNUMBER_FRACTION, position626)
					}
					goto l625
				l624:
					position, tokenIndex, depth = position624, tokenIndex624, depth624
				}
			l625:
				{
					position629, tokenIndex629, depth629 := position, tokenIndex, depth
					{
						position631 := position
						depth++
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
								goto l629
							}
							position++
						}
					l632:
						{
							position634, tokenIndex634, depth634 := position, tokenIndex, depth
							{
								position636, tokenIndex636, depth636 := position, tokenIndex, depth
								if buffer[position] != rune('+') {
									goto l637
								}
								position++
								goto l636
							l637:
								position, tokenIndex, depth = position636, tokenIndex636, depth636
								if buffer[position] != rune('-') {
									goto l634
								}
								position++
							}
						l636:
							goto l635
						l634:
							position, tokenIndex, depth = position634, tokenIndex634, depth634
						}
					l635:
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l629
						}
						position++
					l638:
						{
							position639, tokenIndex639, depth639 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l639
							}
							position++
							goto l638
						l639:
							position, tokenIndex, depth = position639, tokenIndex639, depth639
						}
						depth--
						add(ruleNUMBER_EXP, position631)
					}
					goto l630
				l629:
					position, tokenIndex, depth = position629, tokenIndex629, depth629
				}
			l630:
				depth--
				add(ruleNUMBER, position615)
			}
			return true
		l614:
			position, tokenIndex, depth = position614, tokenIndex614, depth614
			return false
		},
		/* 59 NUMBER_NATURAL <- <('0' / ([1-9] [0-9]*))> */
		nil,
		/* 60 NUMBER_FRACTION <- <('.' [0-9]+)> */
		nil,
		/* 61 NUMBER_INTEGER <- <('-'? NUMBER_NATURAL)> */
		nil,
		/* 62 NUMBER_EXP <- <(('e' / 'E') ('+' / '-')? [0-9]+)> */
		nil,
		/* 63 DURATION <- <(NUMBER [a-z]+ KEY)> */
		nil,
		/* 64 PAREN_OPEN <- <'('> */
		func() bool {
			position645, tokenIndex645, depth645 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position646 := position
				depth++
				if buffer[position] != rune('(') {
					goto l645
				}
				position++
				depth--
				add(rulePAREN_OPEN, position646)
			}
			return true
		l645:
			position, tokenIndex, depth = position645, tokenIndex645, depth645
			return false
		},
		/* 65 PAREN_CLOSE <- <')'> */
		func() bool {
			position647, tokenIndex647, depth647 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position648 := position
				depth++
				if buffer[position] != rune(')') {
					goto l647
				}
				position++
				depth--
				add(rulePAREN_CLOSE, position648)
			}
			return true
		l647:
			position, tokenIndex, depth = position647, tokenIndex647, depth647
			return false
		},
		/* 66 COMMA <- <','> */
		func() bool {
			position649, tokenIndex649, depth649 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position650 := position
				depth++
				if buffer[position] != rune(',') {
					goto l649
				}
				position++
				depth--
				add(ruleCOMMA, position650)
			}
			return true
		l649:
			position, tokenIndex, depth = position649, tokenIndex649, depth649
			return false
		},
		/* 67 _ <- <((&('/') COMMENT_BLOCK) | (&('-') COMMENT_TRAIL) | (&('\t' | '\n' | ' ') SPACE))*> */
		func() bool {
			{
				// @@ func literal escapes to heap
				// @@ func literal escapes to heap
				position652 := position
				depth++
				// @@ &position escapes to heap
			l653:
				// @@ &depth escapes to heap
				{
					position654, tokenIndex654, depth654 := position, tokenIndex, depth
					{
						// @@ &tokenIndex escapes to heap
						switch buffer[position] {
						case '/':
							{
								position656 := position
								depth++
								if buffer[position] != rune('/') {
									goto l654
								}
								position++
								if buffer[position] != rune('*') {
									goto l654
								}
								position++
							l657:
								{
									position658, tokenIndex658, depth658 := position, tokenIndex, depth
									{
										position659, tokenIndex659, depth659 := position, tokenIndex, depth
										if buffer[position] != rune('*') {
											goto l659
										}
										position++
										if buffer[position] != rune('/') {
											goto l659
										}
										position++
										goto l658
									l659:
										position, tokenIndex, depth = position659, tokenIndex659, depth659
									}
									if !matchDot() {
										goto l658
									}
									goto l657
								l658:
									position, tokenIndex, depth = position658, tokenIndex658, depth658
								}
								if buffer[position] != rune('*') {
									goto l654
								}
								position++
								if buffer[position] != rune('/') {
									goto l654
								}
								position++
								depth--
								add(ruleCOMMENT_BLOCK, position656)
							}
							break
						case '-':
							{
								position660 := position
								depth++
								if buffer[position] != rune('-') {
									goto l654
								}
								position++
								if buffer[position] != rune('-') {
									goto l654
								}
								position++
							l661:
								{
									position662, tokenIndex662, depth662 := position, tokenIndex, depth
									{
										position663, tokenIndex663, depth663 := position, tokenIndex, depth
										if buffer[position] != rune('\n') {
											goto l663
										}
										position++
										goto l662
									l663:
										position, tokenIndex, depth = position663, tokenIndex663, depth663
									}
									if !matchDot() {
										goto l662
									}
									goto l661
								l662:
									position, tokenIndex, depth = position662, tokenIndex662, depth662
								}
								depth--
								add(ruleCOMMENT_TRAIL, position660)
							}
							break
						default:
							{
								position664 := position
								depth++
								{
									switch buffer[position] {
									case '\t':
										if buffer[position] != rune('\t') {
											goto l654
										}
										position++
										break
									case '\n':
										if buffer[position] != rune('\n') {
											goto l654
										}
										position++
										break
									default:
										if buffer[position] != rune(' ') {
											goto l654
										}
										position++
										break
									}
								}

								depth--
								add(ruleSPACE, position664)
							}
							break
						}
					}

					goto l653
				l654:
					position, tokenIndex, depth = position654, tokenIndex654, depth654
				}
				depth--
				add(rule_, position652)
			}
			return true
		},
		/* 68 COMMENT_TRAIL <- <('-' '-' (!'\n' .)*)> */
		nil,
		/* 69 COMMENT_BLOCK <- <('/' '*' (!('*' '/') .)* ('*' '/'))> */
		nil,
		/* 70 KEY <- <!ID_CONT> */
		func() bool {
			position668, tokenIndex668, depth668 := position, tokenIndex, depth
			// @@ func literal escapes to heap
			// @@ func literal escapes to heap
			{
				// @@ &position escapes to heap
				// @@ &tokenIndex escapes to heap
				// @@ &depth escapes to heap
				position669 := position
				depth++
				{
					position670, tokenIndex670, depth670 := position, tokenIndex, depth
					if !_rules[ruleID_CONT]() {
						goto l670
						// @@ &_rules escapes to heap
					}
					goto l668
				l670:
					position, tokenIndex, depth = position670, tokenIndex670, depth670
				}
				depth--
				add(ruleKEY, position669)
			}
			return true
		l668:
			position, tokenIndex, depth = position668, tokenIndex668, depth668
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
		/* 76 Action3 <- <{ p.addMatchClause() }> */
		nil,
		/* 77 Action4 <- <{ p.makeDescribeMetrics() }> */
		nil,
		nil,
		/* 79 Action5 <- <{ p.pushNode(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 80 Action6 <- <{ p.makeDescribe() }> */
		nil,
		/* 81 Action7 <- <{ p.addEvaluationContext() }> */
		nil,
		/* 82 Action8 <- <{ p.addPropertyKey(buffer[begin:end])   }> */
		nil,
		/* 83 Action9 <- <{ p.addPropertyValue(buffer[begin:end]) }> */
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
		/* 96 Action22 <- <{ p.pushNode(unescapeLiteral(buffer[begin:end])) }> */
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
		/* 101 Action27 <- <{ p.addNumberNode(buffer[begin:end]) }> */
		nil,
		/* 102 Action28 <- <{ p.addStringNode(unescapeLiteral(buffer[begin:end])) }> */
		nil,
		/* 103 Action29 <- <{ p.addAnnotationExpression(buffer[begin:end]) }> */
		nil,
		/* 104 Action30 <- <{ p.addGroupBy() }> */
		nil,
		/* 105 Action31 <- <{
		   p.pushNode(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 106 Action32 <- <{
		   p.addFunctionInvocation()
		 }> */
		nil,
		/* 107 Action33 <- <{
		   p.pushNode(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 108 Action34 <- <{ p.addNullPredicate() }> */
		nil,
		/* 109 Action35 <- <{
		   p.addMetricExpression()
		 }> */
		nil,
		/* 110 Action36 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 111 Action37 <- <{
		   p.appendGroupBy(unescapeLiteral(buffer[begin:end]))
		 }> */
		nil,
		/* 112 Action38 <- <{
		   p.appendCollapseBy(unescapeLiteral(text))
		 }> */
		nil,
		/* 113 Action39 <- <{p.appendCollapseBy(unescapeLiteral(text))}> */
		nil,
		/* 114 Action40 <- <{ p.addOrPredicate() }> */
		nil,
		/* 115 Action41 <- <{ p.addAndPredicate() }> */
		nil,
		/* 116 Action42 <- <{ p.addNotPredicate() }> */
		nil,
		/* 117 Action43 <- <{
		   p.addLiteralMatcher()
		 }> */
		nil,
		/* 118 Action44 <- <{
		   p.addLiteralMatcher()
		   p.addNotPredicate()
		 }> */
		nil,
		/* 119 Action45 <- <{
		   p.addRegexMatcher()
		 }> */
		nil,
		/* 120 Action46 <- <{
		   p.addListMatcher()
		 }> */
		nil,
		/* 121 Action47 <- <{
		  p.pushNode(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 122 Action48 <- <{ p.addLiteralList() }> */
		nil,
		/* 123 Action49 <- <{
		  p.appendLiteral(unescapeLiteral(buffer[begin:end]))
		}> */
		nil,
		/* 124 Action50 <- <{ p.addTagLiteral(unescapeLiteral(buffer[begin:end])) }> */
		nil,
	}
	p.rules = _rules
}
