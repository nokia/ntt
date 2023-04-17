package syntax

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nokia/ntt/internal/errors"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/ttcn3/ast"
	tokn "github.com/nokia/ntt/ttcn3/token"
)

type tokenNode struct {
	*tree
	idx int
}

func (n *tokenNode) Kind() tokn.Kind {
	return n.tokens[n.idx].Kind
}

func (n *tokenNode) Pos() loc.Pos {
	return n.tokens[n.idx].Pos
}

func (n *tokenNode) End() loc.Pos {
	tok := n.tokens[n.idx]
	return loc.Pos(int(tok.Pos) + len(tok.String()))
}

func (n *tokenNode) LastTok() ast.Token   { return n }
func (n *tokenNode) FirstTok() ast.Token  { return n }
func (n *tokenNode) Children() []ast.Node { return nil }
func (n *tokenNode) PrevTok() ast.Token {
	if n.idx <= 0 {
		return nil
	}
	return &tokenNode{idx: n.idx - 1, tree: n.tree}
}

func (n *tokenNode) NextTok() ast.Token {
	if n.idx >= len(n.tree.tokens)-1 {
		return nil
	}
	return &tokenNode{idx: n.idx + 1, tree: n.tree}
}

func (n *tokenNode) String() string {
	return n.tokens[n.idx].String()
}

func (n *tokenNode) Inspect(fn func(ast.Node) bool) {
	fn(n)
}

type tree struct {
	tokens []token
}

type token struct {
	Kind tokn.Kind
	Lit  string
	Pos  loc.Pos
}

func (tok token) String() string {
	if tok.Kind.IsLiteral() {
		return tok.Lit
	}
	return tok.Kind.String()
}

// The parser structure holds the parser's internal state.
type parser struct {
	file    *loc.File
	errors  errors.ErrorList
	scanner Scanner

	// Tracing/debugging
	mode   Mode // parsing mode
	trace  bool // == (mode & Trace != 0)
	semi   bool // == (mode & PedanticSemicolon != 0)
	indent int  // indentation used for tracing output

	*tree

	// Tokens/Backtracking
	cursor  int
	queue   []ast.Token
	markers []int
	tok     tokn.Kind // for convenience (p.tok is used frequently)

	// Semicolon helper
	seenBrace bool

	// Pre-processor Handling
	ppLvl  int
	ppCnt  int
	ppSkip bool
	ppDefs map[string]bool

	names map[string]bool
	uses  map[string]bool

	// Error recovery
	// (used to limit the number of calls to advance
	// w/o making scanning progress - avoids potential endless
	// loops across multiple parser functions during error recovery)
	syncPos loc.Pos // last synchronization position
	syncCnt int     // number of advance calls without progress
}

func (p *parser) init(fset *loc.FileSet, filename string, src []byte, mode Mode) {
	if s := os.Getenv("NTT_DEBUG"); s == "trace" {
		mode |= Trace
	}

	p.file = fset.AddFile(filename, -1, len(src))

	eh := func(pos loc.Position, msg string) {
		p.errors.Add(pos, msg)
	}
	p.scanner.Init(p.file, src, eh)

	p.mode = mode
	p.trace = mode&Trace != 0 // for convenience (p.trace is used frequently)
	p.semi = mode&PedanticSemicolon != 0

	p.ppDefs = make(map[string]bool)
	p.ppDefs["0"] = false
	p.ppDefs["1"] = true

	p.names = make(map[string]bool)
	p.uses = make(map[string]bool)

	// fetch first token
	p.tree = &tree{}
	tok := p.peek(1)
	p.tok = tok.Kind()
}

// Usage pattern: defer un(trace(p, "..."))
func un(p *parser) {
	p.indent--
	p.printTrace(")")
}

func (p *parser) handlePreproc(s string) {
	f := strings.Fields(s)
	switch b := false; f[0] {
	case "#ifndef":
		b = true
		fallthrough
	case "#ifdef", "#if":
		if len(f) < 2 {
			p.error(p.pos(1), "missing condition in preprocessor directive")
			break
		}
		if p.ppSkip == false && p.ppLvl == p.ppCnt {
			p.ppLvl++
			p.ppSkip = (p.ppDefs[f[1]] == b)
		}
		p.ppCnt++
	case "#else":
		if p.ppLvl == p.ppCnt {
			p.ppSkip = !p.ppSkip
		}
	case "#endif":
		if p.ppLvl == p.ppCnt {
			p.ppLvl--
			p.ppSkip = false
		}
		p.ppCnt--

	case "#define":
		switch len(f) {
		case 2:
			p.ppDefs[f[1]] = true
		case 3:
			if v, err := strconv.ParseBool(f[2]); err != nil {
				p.ppDefs[f[1]] = v
				break
			}
			p.error(p.pos(1), "not a boolean expression")
		default:
			p.error(p.pos(1), "malformed 'define' directive")
		}
	default:
		if !strings.HasPrefix(s, "#!") {
			p.error(p.pos(1), "unknown preprocessor directive")
		}
	}
}

// Advance to the next token
func (p *parser) consume() ast.Token {
	tok := p.queue[p.cursor]
	if p.trace {
		p.printTrace(tok.String())
	}

	// Track curly braces for TTCN-3 semicolon rules
	p.seenBrace = false
	if p.tok == tokn.RBRACE {
		p.seenBrace = true
	}

	p.cursor++
	if p.cursor == len(p.queue) && !p.speculating() {
		p.cursor = 0
		p.queue = p.queue[:0]
	}
	p.peek(1)
	p.tok = p.queue[p.cursor].Kind()
	return tok
}

func (p *parser) ignoreToken(tok tokn.Kind) bool {
	switch {
	case tok == tokn.COMMENT:
		return true
	case tok == tokn.PREPROC:
		return true
	default:
		if p.ppSkip && tok != tokn.EOF {
			return true
		}
		return false
	}
}

func (p *parser) grow(n int) {
	for n > 0 {
		pos, kind, lit := p.scanner.Scan()
		tok := token{Pos: pos, Kind: kind, Lit: lit}
		p.tokens = append(p.tokens, tok)
		if !p.ignoreToken(kind) {
			p.queue = append(p.queue, &tokenNode{
				idx:  len(p.tokens) - 1,
				tree: p.tree,
			})
			n--
		}
	}
}

func (p *parser) peek(i int) ast.Token {
	idx := p.cursor + i - 1
	last := len(p.queue) - 1
	if idx > last {
		p.grow(idx - last)
	}
	return p.queue[idx]
}

func (p *parser) pos(i int) loc.Pos {
	tok := p.peek(i)
	return tok.Pos()
}

func (p *parser) lit(i int) string {
	return p.peek(i).String()
}

func (p *parser) mark() {
	p.markers = append(p.markers, p.cursor)
}

func (p *parser) commit() {
	last := len(p.markers) - 1
	p.markers = p.markers[0:last]
}

func (p *parser) reset() {
	last := len(p.markers) - 1
	marker := p.markers[last]
	p.markers = p.markers[0:last]
	p.cursor = marker
	p.tok = p.queue[p.cursor].Kind()
}

func (p *parser) speculating() bool {
	return len(p.markers) > 0
}

// ----------------------------------------------------------------------------
// Tracing support

func (p *parser) printTrace(a ...interface{}) {
	const dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . "
	const n = len(dots)
	pos := p.file.Position(p.pos(1))
	fmt.Printf("%5d:%3d: ", pos.Line, pos.Column)
	i := 2 * p.indent
	for i > n {
		fmt.Print(dots)
		i -= n
	}
	// i <= n
	fmt.Print(dots[0:i])
	fmt.Println(a...)
}

func trace(p *parser, msg string) *parser {
	p.printTrace(msg, "(")
	p.indent++
	return p
}

// A bailout panic is raised to indicate early termination.
type bailout struct{}

func (p *parser) error(pos loc.Pos, msg string) {
	epos := p.file.Position(pos)

	// If AllErrors is not set, discard errors reported on the same line
	// as the last recorded error and stop parsing if there are more than
	// 10 errors.
	if p.mode&AllErrors == 0 {
		n := len(p.errors)
		if n > 0 && p.errors[n-1].Pos.Line == epos.Line {
			return // discard - likely a spurious error
		}
		if n > 10 {
			panic(bailout{})
		}
	}

	p.errors.Add(epos, msg)
}

func (p *parser) errorExpected(pos loc.Pos, msg string) {
	msg = "expected " + msg
	tok := p.peek(1)
	if pos == tok.Pos() {
		// the error happened at the current position;
		// make the error message more specific
		msg += ", found " + tok.String()
	}
	p.error(pos, msg)
}

func (p *parser) expect(k tokn.Kind) ast.Token {
	if p.tok != k {
		tok := p.peek(1)
		p.errorExpected(tok.Pos(), "'"+k.String()+"'")
	}
	return p.consume() // make progress
}

func (p *parser) expectSemi(tok ast.Token) {
	if p.tok == tokn.SEMICOLON {
		p.consume()
		return
	}

	// pedantic semicolon
	if p.semi {
		// semicolon is optional before a closing '}'
		if !p.seenBrace && p.tok == tokn.RBRACE && p.tok != tokn.EOF {
			p.errorExpected(p.pos(1), "';'")
			p.advance(stmtStart)
		}
	}
}

// advance consumes tokens until the current token p.tok
// is in the 'to' set, or EOF. For error recovery.
func (p *parser) advance(to map[tokn.Kind]bool) {
	for ; p.tok != tokn.EOF; p.consume() {
		if to[p.tok] {
			// Return only if parser made some progress since last
			// sync or if it has not reached 10 advance calls without
			// progress. Otherwise consume at least one token to
			// avoid an endless parser loop (it is possible that
			// both parseOperand and parseStmt call advance and
			// correctly do not advance, thus the need for the
			// invocation limit p.syncCnt).
			if p.pos(1) == p.syncPos && p.syncCnt < 10 {
				p.syncCnt++
				return
			}
			if p.pos(1) > p.syncPos {
				p.syncPos = p.pos(1)
				p.syncCnt = 0
				return
			}
			// Reaching here indicates a parser bug, likely an
			// incorrect token list in this function, but it only
			// leads to skipping of possibly correct code if a
			// previous error is present, and thus is preferred
			// over a non-terminating parse.
		}
	}
}

var stmtStart = map[tokn.Kind]bool{
	tokn.ALT:        true,
	tokn.ALTSTEP:    true,
	tokn.BREAK:      true,
	tokn.CASE:       true,
	tokn.CONST:      true,
	tokn.CONTINUE:   true,
	tokn.CONTROL:    true,
	tokn.DISPLAY:    true,
	tokn.DO:         true,
	tokn.ELSE:       true,
	tokn.ENCODE:     true,
	tokn.EXTENSION:  true,
	tokn.FOR:        true,
	tokn.FRIEND:     true,
	tokn.FUNCTION:   true,
	tokn.GOTO:       true,
	tokn.GROUP:      true,
	tokn.IF:         true,
	tokn.IMPORT:     true,
	tokn.INTERLEAVE: true,
	tokn.LABEL:      true,
	tokn.MAP:        true,
	tokn.MODULE:     true,
	tokn.MODULEPAR:  true,
	tokn.PORT:       true,
	tokn.PRIVATE:    true,
	tokn.PUBLIC:     true,
	tokn.RBRACE:     true,
	tokn.REPEAT:     true,
	tokn.RETURN:     true,
	tokn.SELECT:     true,
	tokn.SEMICOLON:  true,
	tokn.SIGNATURE:  true,
	tokn.TEMPLATE:   true,
	tokn.TESTCASE:   true,
	tokn.TIMER:      true,
	tokn.TYPE:       true,
	tokn.UNMAP:      true,
	tokn.VAR:        true,
	tokn.VARIANT:    true,
	tokn.WHILE:      true,
}

var operandStart = map[tokn.Kind]bool{
	tokn.ADDRESS:    true,
	tokn.ALL:        true,
	tokn.ANY:        true,
	tokn.ANYKW:      true,
	tokn.BSTRING:    true,
	tokn.CHARSTRING: true,
	tokn.ERROR:      true,
	tokn.FAIL:       true,
	tokn.FALSE:      true,
	tokn.FLOAT:      true,
	//tokn.IDENT: true, TODO(5nord) fix conflict, see failing parser tests
	tokn.INCONC:    true,
	tokn.INT:       true,
	tokn.MAP:       true,
	tokn.MTC:       true,
	tokn.MUL:       true,
	tokn.NAN:       true,
	tokn.NONE:      true,
	tokn.NULL:      true,
	tokn.OMIT:      true,
	tokn.PASS:      true,
	tokn.STRING:    true,
	tokn.SYSTEM:    true,
	tokn.TESTCASE:  true,
	tokn.TIMER:     true,
	tokn.TRUE:      true,
	tokn.UNIVERSAL: true,
	tokn.UNMAP:     true,
}

var topLevelTokens = map[tokn.Kind]bool{
	tokn.COMMA:      true,
	tokn.SEMICOLON:  true,
	tokn.MODULE:     true,
	tokn.CONTROL:    true,
	tokn.EXTERNAL:   true,
	tokn.FRIEND:     true,
	tokn.FUNCTION:   true,
	tokn.GROUP:      true,
	tokn.IMPORT:     true,
	tokn.MODULEPAR:  true,
	tokn.SIGNATURE:  true,
	tokn.TEMPLATE:   true,
	tokn.TYPE:       true,
	tokn.VAR:        true,
	tokn.ALTSTEP:    true,
	tokn.CONST:      true,
	tokn.PRIVATE:    true,
	tokn.PUBLIC:     true,
	tokn.TIMER:      true,
	tokn.PORT:       true,
	tokn.REPEAT:     true,
	tokn.BREAK:      true,
	tokn.CONTINUE:   true,
	tokn.LABEL:      true,
	tokn.GOTO:       true,
	tokn.RETURN:     true,
	tokn.SELECT:     true,
	tokn.ALT:        true,
	tokn.INTERLEAVE: true,
	tokn.LBRACK:     true,
	tokn.FOR:        true,
	tokn.WHILE:      true,
	tokn.DO:         true,
	tokn.IF:         true,
	tokn.LBRACE:     true,
	tokn.IDENT:      true,
	tokn.ANYKW:      true,
	tokn.ALL:        true,
	tokn.MAP:        true,
	tokn.UNMAP:      true,
	tokn.MTC:        true,
	tokn.TESTCASE:   true,
}

// parse is a generic entry point
func (p *parser) parse() ast.Node {
	switch p.tok {
	case tokn.MODULE:
		return p.parseModule()

	case tokn.CONTROL,
		tokn.EXTERNAL,
		tokn.FRIEND,
		tokn.FUNCTION,
		tokn.GROUP,
		tokn.IMPORT,
		tokn.MODULEPAR,
		tokn.SIGNATURE,
		tokn.TEMPLATE,
		tokn.TYPE,
		tokn.VAR,
		tokn.ALTSTEP,
		tokn.CONST,
		tokn.PRIVATE,
		tokn.PUBLIC:
		return p.parseModuleDef()

	case tokn.TIMER, tokn.PORT,
		tokn.REPEAT, tokn.BREAK, tokn.CONTINUE,
		tokn.LABEL,
		tokn.GOTO,
		tokn.RETURN,
		tokn.SELECT,
		tokn.ALT, tokn.INTERLEAVE,
		tokn.LBRACK,
		tokn.FOR,
		tokn.WHILE,
		tokn.DO,
		tokn.IF,
		tokn.LBRACE,
		tokn.IDENT, tokn.ANYKW, tokn.ALL, tokn.MAP, tokn.UNMAP, tokn.MTC:
		return p.parseStmt()

	case tokn.TESTCASE:
		if p.peek(1).Kind() == tokn.DOT {
			return p.parseStmt()
		}
		return p.parseModuleDef()
	default:
		return p.parseExpr()
	}
}

/*************************************************************************
 * Expressions
 *************************************************************************/

// ExprList ::= ast.Expr { "," ast.Expr }
func (p *parser) parseExprList() (list []ast.Expr) {
	if p.trace {
		defer un(trace(p, "ExprList"))
	}

	list = append(list, p.parseExpr())
	for p.tok == tokn.COMMA {
		p.consume()
		list = append(list, p.parseExpr())
	}
	return list
}

// ast.Expr ::= BinaryExpr
func (p *parser) parseExpr() ast.Expr {
	if p.trace {
		defer un(trace(p, "ast.Expr"))
	}

	return p.parseBinaryExpr(tokn.LowestPrec + 1)
}

// BinaryExpr ::= UnaryExpr
//
//	| BinaryExpr OP BinaryExpr
func (p *parser) parseBinaryExpr(prec1 int) ast.Expr {
	x := p.parseUnaryExpr()
	for {
		prec := p.tok.Precedence()
		if prec < prec1 {
			return x
		}
		op := p.consume()
		y := p.parseBinaryExpr(prec + 1)
		x = &ast.BinaryExpr{X: x, Op: op, Y: y}
	}
}

// UnaryExpr ::= "-"
//
//	| ("-"|"+"|"!"|"not"|"not4b") UnaryExpr
//	| PrimaryExpr
func (p *parser) parseUnaryExpr() ast.Expr {
	switch p.tok {
	case tokn.ADD, tokn.EXCL, tokn.NOT, tokn.NOT4B, tokn.SUB:
		tok := p.consume()
		// handle unused expr '-'
		if tok.Kind() == tokn.SUB {
			switch p.tok {
			case tokn.COMMA, tokn.SEMICOLON, tokn.RBRACE, tokn.RBRACK, tokn.RPAREN, tokn.EOF:
				return &ast.ValueLiteral{Tok: tok}
			}
		}
		return &ast.UnaryExpr{Op: tok, X: p.parseUnaryExpr()}
	case tokn.COLONCOLON:
		tok := p.consume()
		return &ast.BinaryExpr{Op: tok, Y: p.parseExpr()}
	}

	return p.parsePrimaryExpr()
}

// PrimaryExpr ::= Operand [{ExtFieldRef}] [Stuff]
//
// ExtFieldRef ::= "." ID
//               | "[" ast.Expr "]"
//               | "(" ExprList ")"
//
// Stuff       ::= ["length"      "(" ExprList ")"]
//                 ["ifpresent"                   ]
//                 [("to"|"from") ast.Expr            ]
//                 ["->"          Redirect        ]

// Redirect    ::= ["value"            ExprList   ]
//
//	["param"            ExprList   ]
//	["sender"           PrimaryExpr]
//	["@index" ["value"] PrimaryExpr]
//	["timestamp"        PrimaryExpr]
func (p *parser) parsePrimaryExpr() ast.Expr {
	x := p.parseOperand()
L:
	for {
		switch p.tok {
		case tokn.DOT:
			x = p.parseSelectorExpr(x)
		case tokn.LBRACK:
			x = p.parseIndexExpr(x)
		case tokn.LPAREN:
			x = p.parseCallExpr(x)
			// Not supporting chained function calls like 'get().x'
			// eleminates conflicts with alt-guards.
			break
		default:
			break L
		}
	}

	if p.tok == tokn.LENGTH {
		x = p.parseLength(x)
	}

	if p.tok == tokn.IFPRESENT {
		x = &ast.UnaryExpr{Op: p.consume(), X: x}
	}

	if p.tok == tokn.TO || p.tok == tokn.FROM {
		x = &ast.BinaryExpr{X: x, Op: p.consume(), Y: p.parseExpr()}
	}

	if p.tok == tokn.REDIR {
		x = p.parseRedirect(x)
	}

	if p.tok == tokn.VALUE {
		x = &ast.ValueExpr{X: x, Tok: p.consume(), Y: p.parseExpr()}
	}

	if p.tok == tokn.PARAM {
		x = &ast.ParamExpr{X: x, Tok: p.consume(), Y: p.parseParenExpr()}
	}

	if p.tok == tokn.ALIVE {
		x = &ast.UnaryExpr{Op: p.consume(), X: x}
	}

	return x
}

// Operand ::= ("any"|"all") ("component"|"port"|"timer"|"from" PrimaryExpr)
//
//	| Literal
//	| Reference
//
// Literal ::= INT | STRING | BSTRING | FLOAT
//
//	| "?" | "*"
//	| "none" | "inconc" | "pass" | "fail" | "error"
//	| "true" | "false"
//	| "not_a_number"
//
// Reference ::= ID
//
//	| "address" | ["unviersal"] "charstring" | "timer"
//	| "null" | "omit"
//	| "mtc" | "system" | "testcase"
//	| "map" | "unmap"
func (p *parser) parseOperand() ast.Expr {
	etok := p.peek(1)
	switch p.tok {
	case tokn.ANYKW, tokn.ALL:
		tok := p.consume()
		switch p.tok {
		case tokn.COMPONENT, tokn.PORT, tokn.TIMER:
			return p.make_use(tok, p.consume())
		case tokn.FROM:
			return &ast.FromExpr{
				Kind:    tok,
				FromTok: p.consume(),
				X:       p.parsePrimaryExpr(),
			}
		}

		return p.make_use(tok)

	case tokn.UNIVERSAL:
		return p.parseUniversalCharstring()

	case tokn.ADDRESS,
		tokn.CHARSTRING,
		tokn.MAP,
		tokn.MTC,
		tokn.SYSTEM,
		tokn.TESTCASE,
		tokn.TIMER,
		tokn.UNMAP:
		return p.make_use(p.consume())

	case tokn.IDENT:
		return p.parseRef()

	case tokn.INT,
		tokn.ANY,
		tokn.BSTRING,
		tokn.ERROR,
		tokn.NULL,
		tokn.OMIT,
		tokn.FAIL,
		tokn.FALSE,
		tokn.FLOAT,
		tokn.INCONC,
		tokn.MUL,
		tokn.NAN,
		tokn.NONE,
		tokn.PASS,
		tokn.STRING,
		tokn.TRUE:
		return &ast.ValueLiteral{Tok: p.consume()}

	case tokn.LPAREN:
		// can be template `x := (1,2,3)`, but also artihmetic expression: `1*(2+3)`
		return p.parseParenExpr()

	case tokn.LBRACK:
		return p.parseIndexExpr(nil)

	case tokn.LBRACE:
		return p.parseCompositeLiteral()

	case tokn.MODIFIES:
		return &ast.ModifiesExpr{
			Tok:    p.consume(),
			X:      p.parsePrimaryExpr(),
			Assign: p.expect(tokn.ASSIGN),
			Y:      p.parseExpr(),
		}

	case tokn.REGEXP:
		return p.parseCallRegexp()

	case tokn.PATTERN:
		return p.parseCallPattern()

	case tokn.DECMATCH:
		return p.parseCallDecmatch()

	case tokn.MODIF:
		return p.parseDecodedModifier()

	default:
		p.errorExpected(p.pos(1), "operand")
	}

	return &ast.ErrorNode{From: etok, To: p.peek(1)}
}

func (p *parser) parseRef() ast.Expr {
	id := p.parseIdent()
	if id == nil {
		return nil
	}

	if p.tok != tokn.LT {
		return id
	}

	p.mark()
	if x := p.tryTypeParameters(); x != nil && !operandStart[p.tok] {
		p.commit()
		return &ast.ParametrizedIdent{Ident: id, Params: x}
	}
	p.reset()
	return id
}

func (p *parser) parseParenExpr() *ast.ParenExpr {
	return &ast.ParenExpr{
		LParen: p.expect(tokn.LPAREN),
		List:   p.parseExprList(),
		RParen: p.expect(tokn.RPAREN),
	}
}

func (p *parser) parseUniversalCharstring() *ast.Ident {
	return p.make_use(p.expect(tokn.UNIVERSAL), p.expect(tokn.CHARSTRING))
}

func (p *parser) parseCompositeLiteral() *ast.CompositeLiteral {
	c := new(ast.CompositeLiteral)
	c.LBrace = p.expect(tokn.LBRACE)
	if p.tok != tokn.RBRACE {
		c.List = p.parseExprList()
	}
	c.RBrace = p.expect(tokn.RBRACE)
	return c
}

func (p *parser) parseCallRegexp() *ast.RegexpExpr {
	c := new(ast.RegexpExpr)
	c.Tok = p.expect(tokn.REGEXP)
	if p.tok == tokn.MODIF {
		c.NoCase = p.consume()
	}
	c.X = p.parseParenExpr()
	return c
}

func (p *parser) parseCallPattern() *ast.PatternExpr {
	c := new(ast.PatternExpr)
	c.Tok = p.expect(tokn.PATTERN)
	if p.tok == tokn.MODIF {
		c.NoCase = p.consume()
	}
	c.X = p.parseExpr()
	return c
}

func (p *parser) parseCallDecmatch() *ast.DecmatchExpr {
	c := new(ast.DecmatchExpr)
	c.Tok = p.expect(tokn.DECMATCH)
	if p.tok == tokn.LPAREN {
		c.Params = p.parseParenExpr()
	}
	c.X = p.parseExpr()
	return c
}

func (p *parser) parseDecodedModifier() *ast.DecodedExpr {
	d := new(ast.DecodedExpr)
	d.Tok = p.expect(tokn.MODIF)
	if d.Tok.String() != "@decoded" {
		p.errorExpected(d.Tok.Pos(), "@decoded")
	}

	if p.tok == tokn.LPAREN {
		d.Params = p.parseParenExpr()
	}
	d.X = p.parsePrimaryExpr()
	return d
}

func (p *parser) parseSelectorExpr(x ast.Expr) *ast.SelectorExpr {
	return &ast.SelectorExpr{X: x, Dot: p.consume(), Sel: p.parseRef()}
}

func (p *parser) parseIndexExpr(x ast.Expr) *ast.IndexExpr {
	return &ast.IndexExpr{
		X:      x,
		LBrack: p.expect(tokn.LBRACK),
		Index:  p.parseExpr(),
		RBrack: p.expect(tokn.RBRACK),
	}
}

func (p *parser) parseCallExpr(x ast.Expr) *ast.CallExpr {
	c := new(ast.CallExpr)
	c.Fun = x
	c.Args = new(ast.ParenExpr)
	c.Args.LParen = p.expect(tokn.LPAREN)
	if p.tok != tokn.RPAREN {
		switch p.tok {
		case tokn.TO, tokn.FROM, tokn.REDIR:
			var x ast.Expr
			if p.tok == tokn.TO || p.tok == tokn.FROM {
				// TODO: Shouldn't this be a FromExpr?
				x = &ast.BinaryExpr{X: x, Op: p.consume(), Y: p.parseExpr()}
			}
			if p.tok == tokn.REDIR {
				x = p.parseRedirect(x)
			}
			c.Args.List = []ast.Expr{x}
		default:
			c.Args.List = append(c.Args.List, p.parseExprList()...)
		}
	}
	c.Args.RParen = p.expect(tokn.RPAREN)
	return c
}

func (p *parser) parseLength(x ast.Expr) *ast.LengthExpr {
	return &ast.LengthExpr{
		X:    x,
		Len:  p.expect(tokn.LENGTH),
		Size: p.parseParenExpr(),
	}
}

func (p *parser) parseRedirect(x ast.Expr) *ast.RedirectExpr {

	r := &ast.RedirectExpr{
		X:   x,
		Tok: p.expect(tokn.REDIR),
	}

	if p.tok == tokn.VALUE {
		r.ValueTok = p.expect(tokn.VALUE)
		r.Value = p.parseExprList()
	}

	if p.tok == tokn.PARAM {
		r.ParamTok = p.expect(tokn.PARAM)
		r.Param = p.parseExprList()
	}

	if p.tok == tokn.SENDER {
		r.SenderTok = p.expect(tokn.SENDER)
		r.Sender = p.parsePrimaryExpr()
	}

	if p.tok == tokn.MODIF {
		if p.lit(1) != "@index" {
			p.errorExpected(p.pos(1), "@index")
		}

		r.IndexTok = p.consume()
		if p.tok == tokn.VALUE {
			r.IndexValueTok = p.consume()
		}
		r.Index = p.parsePrimaryExpr()
	}

	if p.tok == tokn.TIMESTAMP {
		r.TimestampTok = p.expect(tokn.TIMESTAMP)
		r.Timestamp = p.parsePrimaryExpr()
	}

	return r
}

func (p *parser) parseName() *ast.Ident {
	switch p.tok {
	case tokn.IDENT, tokn.ADDRESS, tokn.CONTROL:
		id := &ast.Ident{Tok: p.consume(), IsName: true}
		p.names[id.String()] = true
		return id
	}
	p.expect(tokn.IDENT)
	return nil

}

func (p *parser) parseIdent() *ast.Ident {
	switch p.tok {
	case tokn.UNIVERSAL:
		return p.parseUniversalCharstring()
	case tokn.IDENT, tokn.ADDRESS, tokn.ALIVE, tokn.CHARSTRING, tokn.CONTROL:
		return p.make_use(p.consume())
	default:
		p.expect(tokn.IDENT) // use expect() error handling
		return nil
	}
}

func (p *parser) parseArrayDefs() []*ast.ParenExpr {
	var l []*ast.ParenExpr
	for p.tok == tokn.LBRACK {
		l = append(l, p.parseArrayDef())
	}
	return l
}

func (p *parser) parseArrayDef() *ast.ParenExpr {
	return &ast.ParenExpr{
		LParen: p.expect(tokn.LBRACK),
		List:   p.parseExprList(),
		RParen: p.expect(tokn.RBRACK),
	}
}

func (p *parser) parseRefList() []ast.Expr {
	l := make([]ast.Expr, 0, 1)
	for {
		l = append(l, p.parseTypeRef())
		if p.tok != tokn.COMMA {
			break
		}
		p.consume()
	}
	return l
}

func (p *parser) parseTypeRef() ast.Expr {
	if p.trace {
		defer un(trace(p, "TypeRef"))
	}
	return p.parsePrimaryExpr()
}

func (p *parser) tryTypeParameters() *ast.ParenExpr {
	if p.trace {
		defer un(trace(p, "tryTypeParameters"))
	}
	x := &ast.ParenExpr{
		LParen: p.consume(),
	}
	for p.tok != tokn.GT {
		y := p.tryTypeParameter()
		if y == nil {
			return nil
		}
		x.List = append(x.List, y)
		if p.tok != tokn.COMMA {
			break
		}
		p.consume()
	}

	if p.tok != tokn.GT {
		return nil
	}
	x.RParen = p.consume()
	return x
}

func (p *parser) tryTypeParameter() ast.Expr {
	if p.trace {
		defer un(trace(p, "tryTypeParameter"))
	}
	x := p.tryTypeIdent()
L:
	for {
		switch p.tok {
		case tokn.DOT:
			x = &ast.SelectorExpr{
				X:   x,
				Dot: p.consume(),
				Sel: p.tryTypeIdent(),
			}
			if x.(*ast.SelectorExpr).Sel == nil {
				return nil
			}
		case tokn.LBRACK:
			lbrack := p.consume()
			dash := p.consume()
			rbrack := p.consume()
			if dash.Kind() != tokn.SUB || rbrack.Kind() != tokn.RBRACK {
				return nil
			}
			x = &ast.IndexExpr{
				X:      x,
				LBrack: lbrack,
				Index:  &ast.ValueLiteral{Tok: dash},
				RBrack: rbrack,
			}

		default:
			break L
		}
	}
	return x
}

func (p *parser) tryTypeIdent() ast.Expr {
	if p.trace {
		defer un(trace(p, "tryTypeIdent"))
	}

	var x *ast.Ident
	switch p.tok {
	case tokn.IDENT, tokn.ADDRESS, tokn.CHARSTRING:
		x = p.make_use(p.consume())
	case tokn.UNIVERSAL:
		x = p.parseUniversalCharstring()
	default:
		return nil
	}

	if p.tok == tokn.LT {
		if y := p.tryTypeParameters(); y == nil {
			return &ast.ParametrizedIdent{
				Ident:  x,
				Params: y,
			}
		}
	}
	return x
}

/*************************************************************************
 * ast.Module
 *************************************************************************/

func (p *parser) parseModuleList() []*ast.Module {
	var list []*ast.Module
	if m := p.parseModule(); m != nil {
		list = append(list, m)
		p.expectSemi(m.LastTok())
	}
	for p.tok == tokn.MODULE {
		if m := p.parseModule(); m != nil {
			list = append(list, m)
			p.expectSemi(m.LastTok())
		}
	}
	p.expect(tokn.EOF)
	return list
}

func (p *parser) parseModule() *ast.Module {
	if p.trace {
		defer un(trace(p, "ast.Module"))
	}

	m := new(ast.Module)
	m.Tok = p.expect(tokn.MODULE)
	m.Name = p.parseName()

	if p.tok == tokn.LANGUAGE {
		m.Language = p.parseLanguageSpec()
	}

	m.LBrace = p.expect(tokn.LBRACE)

	for p.tok != tokn.RBRACE && p.tok != tokn.EOF {
		m.Defs = append(m.Defs, p.parseModuleDef())
		p.expectSemi(m.Defs[len(m.Defs)-1].LastTok())
	}
	m.RBrace = p.expect(tokn.RBRACE)
	m.With = p.parseWith()
	return m
}

func (p *parser) parseLanguageSpec() *ast.LanguageSpec {
	l := new(ast.LanguageSpec)
	l.Tok = p.consume()
	for {
		l.List = append(l.List, p.expect(tokn.STRING))
		if p.tok != tokn.COMMA {
			break
		}
		p.consume()
	}
	return l
}

func (p *parser) parseModuleDef() *ast.ModuleDef {
	m := new(ast.ModuleDef)
	switch p.tok {
	case tokn.PRIVATE, tokn.PUBLIC:
		m.Visibility = p.consume()
	case tokn.FRIEND:
		if p.peek(2).Kind() != tokn.MODULE {
			m.Visibility = p.consume()
		}
	}

	etok := p.peek(1)
	switch p.tok {
	case tokn.IMPORT:
		m.Def = p.parseImport()
	case tokn.GROUP:
		m.Def = p.parseGroup()
	case tokn.FRIEND:
		m.Def = p.parseFriend()
	case tokn.TYPE:
		m.Def = p.parseTypeDecl()
	case tokn.TEMPLATE:
		m.Def = p.parseTemplateDecl()
	case tokn.MODULEPAR:
		m.Def = p.parseModulePar()
	case tokn.VAR, tokn.CONST:
		m.Def = p.parseValueDecl()
	case tokn.SIGNATURE:
		m.Def = p.parseSignatureDecl()
	case tokn.FUNCTION, tokn.TESTCASE, tokn.ALTSTEP:
		m.Def = p.parseFuncDecl()
	case tokn.CONTROL:
		m.Def = &ast.ControlPart{Name: p.parseIdent(), Body: p.parseBlockStmt(), With: p.parseWith()}
	case tokn.EXTERNAL:
		switch p.peek(2).Kind() {
		case tokn.FUNCTION:
			m.Def = p.parseExtFuncDecl()
		case tokn.CONST:
			p.error(p.pos(1), "external constants not suppored")
			p.consume()
			m.Def = p.parseValueDecl()
		default:
			p.errorExpected(p.pos(1), "'function'")
			p.advance(stmtStart)
			m.Def = &ast.ErrorNode{From: etok, To: p.peek(1)}
		}
	default:
		p.errorExpected(p.pos(1), "module definition")
		p.advance(stmtStart)
		m.Def = &ast.ErrorNode{From: etok, To: p.peek(1)}
	}

	if m.Def != nil {
		p.addName(m.Def)
	}
	return m
}

func (p *parser) addName(n ast.Node) {
	switch n := n.(type) {
	case *ast.ValueDecl:
		for _, n := range n.Decls {
			p.addName(n)
		}
	default:
		if name := ast.Name(n); name != "" {
			p.names[name] = true
		}
	}
}

/*************************************************************************
 * Import Definition
 *************************************************************************/

func (p *parser) make_use(toks ...ast.Token) *ast.Ident {
	if len(toks) != 1 && len(toks) != 2 {
		panic("No support for multi-token identifiers.")
	}
	id := &ast.Ident{Tok: toks[0]}
	p.uses[toks[0].String()] = true
	if len(toks) == 2 {
		id.Tok2 = toks[1]
		p.uses[toks[1].String()] = true
	}
	return id
}

func (p *parser) parseImport() *ast.ImportDecl {
	if p.trace {
		defer un(trace(p, "Import"))
	}

	x := new(ast.ImportDecl)
	x.ImportTok = p.consume()
	x.FromTok = p.expect(tokn.FROM)
	x.Module = p.parseIdent()

	if p.tok == tokn.LANGUAGE {
		x.Language = p.parseLanguageSpec()
	}

	switch p.tok {
	case tokn.ALL:
		y := &ast.DefKindExpr{}
		var z ast.Expr = p.make_use(p.consume())
		if p.tok == tokn.EXCEPT {
			z = &ast.ExceptExpr{
				X:         z,
				ExceptTok: p.consume(),
				LBrace:    p.expect(tokn.LBRACE),
				List:      p.parseExceptStmts(),
				RBrace:    p.expect(tokn.RBRACE),
			}
		}
		y.List = []ast.Expr{z}
		x.List = append(x.List, y)
	case tokn.LBRACE:
		x.LBrace = p.expect(tokn.LBRACE)
		for p.tok != tokn.RBRACE && p.tok != tokn.EOF {
			x.List = append(x.List, p.parseImportStmt())
			p.expectSemi(x.List[len(x.List)-1].LastTok())
		}
		x.RBrace = p.expect(tokn.RBRACE)
	default:
		p.errorExpected(p.pos(1), "'all' or import spec")
	}

	x.With = p.parseWith()

	return x
}

func (p *parser) parseImportStmt() *ast.DefKindExpr {
	x := new(ast.DefKindExpr)
	switch p.tok {
	case tokn.ALTSTEP, tokn.CONST, tokn.FUNCTION, tokn.MODULEPAR,
		tokn.SIGNATURE, tokn.TEMPLATE, tokn.TESTCASE, tokn.TYPE:
		x.Kind = p.consume()
		if p.tok == tokn.ALL {
			var y ast.Expr = p.make_use(p.consume())
			if p.tok == tokn.EXCEPT {
				y = &ast.ExceptExpr{
					X:         y,
					ExceptTok: p.consume(),
					List:      p.parseRefList(),
				}
			}
			x.List = []ast.Expr{y}
		} else {
			x.List = p.parseRefList()
		}
	case tokn.GROUP:
		x.Kind = p.consume()
		for {
			y := p.parseTypeRef()
			if p.tok == tokn.EXCEPT {
				y = &ast.ExceptExpr{
					X:         y,
					ExceptTok: p.consume(),
					LBrace:    p.expect(tokn.LBRACE),
					List:      p.parseExceptStmts(),
					RBrace:    p.expect(tokn.RBRACE),
				}
			}
			x.List = append(x.List, y)
			if p.tok != tokn.COMMA {
				break
			}
			p.consume()
		}
	case tokn.IMPORT:
		x.Kind = p.consume()
		x.List = []ast.Expr{p.make_use(p.expect(tokn.ALL))}
	default:
		p.errorExpected(p.pos(1), "import definition qualifier")
		p.advance(stmtStart)
	}
	return x
}

func (p *parser) parseExceptStmts() []ast.Expr {
	var list []ast.Expr
	for p.tok != tokn.RBRACE && p.tok != tokn.EOF {
		x := p.parseExceptStmt()
		p.expectSemi(x.LastTok())
		list = append(list, x)
	}
	return list
}

func (p *parser) parseExceptStmt() *ast.DefKindExpr {
	x := new(ast.DefKindExpr)
	switch p.tok {
	case tokn.ALTSTEP, tokn.CONST, tokn.FUNCTION, tokn.GROUP,
		tokn.IMPORT, tokn.MODULEPAR, tokn.SIGNATURE, tokn.TEMPLATE,
		tokn.TESTCASE, tokn.TYPE:
		x.Kind = p.consume()
	default:
		p.errorExpected(p.pos(1), "definition qualifier")
	}

	if p.tok == tokn.ALL {
		x.List = []ast.Expr{p.make_use(p.consume())}
	} else {
		x.List = p.parseRefList()
	}
	return x
}

/*************************************************************************
 * Group Definition
 *************************************************************************/

func (p *parser) parseGroup() *ast.GroupDecl {
	x := new(ast.GroupDecl)
	x.Tok = p.consume()
	x.Name = p.parseName()
	x.LBrace = p.expect(tokn.LBRACE)

	for p.tok != tokn.RBRACE && p.tok != tokn.EOF {
		x.Defs = append(x.Defs, p.parseModuleDef())
		p.expectSemi(x.Defs[len(x.Defs)-1].LastTok())
	}
	x.RBrace = p.expect(tokn.RBRACE)
	x.With = p.parseWith()
	return x
}

func (p *parser) parseFriend() *ast.FriendDecl {
	return &ast.FriendDecl{
		FriendTok: p.expect(tokn.FRIEND),
		ModuleTok: p.expect(tokn.MODULE),
		Module:    p.parseIdent(),
		With:      p.parseWith(),
	}
}

/*************************************************************************
 * With Attributes
 *************************************************************************/

func (p *parser) parseWith() *ast.WithSpec {
	if p.tok != tokn.WITH {
		return nil
	}
	x := new(ast.WithSpec)
	x.Tok = p.consume()
	x.LBrace = p.expect(tokn.LBRACE)
	for p.tok != tokn.RBRACE && p.tok != tokn.EOF {
		x.List = append(x.List, p.parseWithStmt())
		p.expectSemi(x.List[len(x.List)-1].LastTok())
	}
	x.RBrace = p.expect(tokn.RBRACE)
	return x
}

func (p *parser) parseWithStmt() *ast.WithStmt {
	if p.trace {
		defer un(trace(p, "WithStmt"))
	}

	x := new(ast.WithStmt)

	switch p.tok {
	case tokn.ENCODE,
		tokn.VARIANT,
		tokn.DISPLAY,
		tokn.EXTENSION,
		tokn.OPTIONAL,
		tokn.STEPSIZE,
		tokn.OVERRIDE:
		x.Kind = p.consume()
	default:
		p.errorExpected(p.pos(1), "with-attribute")
		p.advance(stmtStart)
	}

	switch p.tok {
	case tokn.OVERRIDE:
		x.Override = p.consume()
	case tokn.MODIF:
		if p.lit(1) != "@local" {
			p.errorExpected(p.pos(1), "@local")
		}
		x.Override = p.consume()
	}

	if p.tok == tokn.LPAREN {
		x.LParen = p.consume()
		for {
			x.List = append(x.List, p.parseWithQualifier())
			if p.tok != tokn.COMMA {
				break
			}
			p.consume()
		}
		x.RParen = p.expect(tokn.RPAREN)
	}

	var v ast.Expr = &ast.ValueLiteral{Tok: p.expect(tokn.STRING)}
	if p.tok == tokn.DOT {
		v = &ast.SelectorExpr{
			X:   v,
			Dot: p.consume(),
			Sel: &ast.ValueLiteral{Tok: p.expect(tokn.STRING)},
		}
	}
	x.Value = v

	return x
}

func (p *parser) parseWithQualifier() ast.Expr {
	etok := p.peek(1)
	switch p.tok {
	case tokn.IDENT:
		return p.parseTypeRef()
	case tokn.LBRACK:
		return p.parseIndexExpr(nil)
	case tokn.TYPE, tokn.TEMPLATE, tokn.CONST, tokn.ALTSTEP, tokn.TESTCASE, tokn.FUNCTION, tokn.SIGNATURE, tokn.MODULEPAR, tokn.GROUP:
		x := new(ast.DefKindExpr)
		x.Kind = p.consume()
		var y ast.Expr = p.make_use(p.expect(tokn.ALL))
		if p.tok == tokn.EXCEPT {
			y = &ast.ExceptExpr{
				X:         y,
				ExceptTok: p.consume(),
				LBrace:    p.expect(tokn.LBRACE),
				List:      p.parseRefList(),
				RBrace:    p.expect(tokn.RBRACE),
			}
		}
		x.List = []ast.Expr{y}
		return x
	default:
		p.errorExpected(p.pos(1), "with-qualifier")
		p.advance(stmtStart)
		return &ast.ErrorNode{From: etok, To: p.peek(1)}
	}
}

/*************************************************************************
 * Type Definitions
 *************************************************************************/

func (p *parser) parseTypeDecl() ast.Decl {
	etok := p.peek(1)
	switch p.peek(2).Kind() {
	case tokn.IDENT, tokn.ADDRESS, tokn.CHARSTRING, tokn.NULL, tokn.UNIVERSAL:
		return p.parseSubTypeDecl()
	case tokn.PORT:
		return p.parsePortTypeDecl()
	case tokn.COMPONENT:
		return p.parseComponentTypeDecl()
	case tokn.UNION:
		return p.parseStructTypeDecl()
	case tokn.SET, tokn.RECORD:
		if p.peek(3).Kind() == tokn.IDENT || p.peek(3).Kind() == tokn.ADDRESS {
			return p.parseStructTypeDecl()
		}
		// lists are also parsed by parseSubTypeDecl
		return p.parseSubTypeDecl()
	case tokn.ENUMERATED:
		return p.parseEnumTypeDecl()
	case tokn.FUNCTION, tokn.ALTSTEP, tokn.TESTCASE:
		return p.parseBehaviourTypeDecl()
	default:
		p.errorExpected(p.pos(1), "type definition")
		p.advance(stmtStart)
		return &ast.ErrorNode{From: etok, To: p.peek(1)}
	}
}

/*************************************************************************
 * Port Type
 *************************************************************************/

func (p *parser) parsePortTypeDecl() *ast.PortTypeDecl {
	if p.trace {
		defer un(trace(p, "ast.PortTypeDecl"))
	}
	x := new(ast.PortTypeDecl)
	x.TypeTok = p.consume()
	x.PortTok = p.consume()
	x.Name = p.parseName()
	if p.tok == tokn.LT {
		x.TypePars = p.parseTypeFormalPars()
	}

	switch p.tok {
	case tokn.MIXED, tokn.MESSAGE, tokn.PROCEDURE:
		x.Kind = p.consume()
	default:
		p.errorExpected(p.pos(1), "'message' or 'procedure'")
	}

	if p.tok == tokn.REALTIME {
		x.Realtime = p.consume()
	}

	x.LBrace = p.expect(tokn.LBRACE)
	for p.tok != tokn.RBRACE && p.tok != tokn.EOF {
		x.Attrs = append(x.Attrs, p.parsePortAttribute())
		p.expectSemi(x.Attrs[len(x.Attrs)-1].LastTok())
	}
	x.RBrace = p.expect(tokn.RBRACE)
	x.With = p.parseWith()
	return x
}

func (p *parser) parsePortAttribute() ast.Node {
	if p.trace {
		defer un(trace(p, "ast.PortAttribute"))
	}
	etok := p.peek(1)
	switch p.tok {
	case tokn.IN, tokn.OUT, tokn.INOUT, tokn.ADDRESS:
		return &ast.PortAttribute{
			Kind:  p.consume(),
			Types: p.parseRefList(),
		}
	case tokn.MAP, tokn.UNMAP:
		return &ast.PortMapAttribute{
			MapTok:   p.consume(),
			ParamTok: p.expect(tokn.PARAM),
			Params:   p.parseFormalPars(),
		}
	default:
		p.errorExpected(p.pos(1), "port attribute")
		p.advance(stmtStart)
		return &ast.ErrorNode{From: etok, To: p.peek(1)}
	}
}

/*************************************************************************
 * Component Type
 *************************************************************************/

func (p *parser) parseComponentTypeDecl() *ast.ComponentTypeDecl {
	if p.trace {
		defer un(trace(p, "ast.ComponentTypeDecl"))
	}
	x := new(ast.ComponentTypeDecl)
	x.TypeTok = p.consume()
	x.CompTok = p.consume()
	x.Name = p.parseName()
	if p.tok == tokn.LT {
		x.TypePars = p.parseTypeFormalPars()
	}
	if p.tok == tokn.EXTENDS {
		x.ExtendsTok = p.consume()
		x.Extends = p.parseRefList()
	}
	x.Body = p.parseBlockStmt()
	x.With = p.parseWith()
	return x
}

/*************************************************************************
 * Struct Type Declaration
 *************************************************************************/

func (p *parser) parseStructTypeDecl() *ast.StructTypeDecl {
	if p.trace {
		defer un(trace(p, "ast.StructTypeDecl"))
	}
	x := new(ast.StructTypeDecl)
	x.TypeTok = p.consume()
	x.Kind = p.consume()
	x.Name = p.parseName()
	if p.tok == tokn.LT {
		x.TypePars = p.parseTypeFormalPars()
	}
	x.LBrace = p.expect(tokn.LBRACE)
	for p.tok != tokn.RBRACE && p.tok != tokn.EOF {
		x.Fields = append(x.Fields, p.parseField())
		if p.tok != tokn.COMMA {
			break
		}
		p.consume()
	}
	x.RBrace = p.expect(tokn.RBRACE)
	x.With = p.parseWith()
	return x
}

/*************************************************************************
 * Enumeration Type Declaration
 *************************************************************************/

func (p *parser) parseEnumTypeDecl() *ast.EnumTypeDecl {
	if p.trace {
		defer un(trace(p, "ast.EnumTypeDecl"))
	}

	x := new(ast.EnumTypeDecl)
	x.TypeTok = p.consume()
	x.EnumTok = p.consume()
	x.Name = p.parseName()
	if p.tok == tokn.LT {
		x.TypePars = p.parseTypeFormalPars()
	}
	x.LBrace = p.expect(tokn.LBRACE)
	for p.tok != tokn.RBRACE && p.tok != tokn.EOF {
		x.Enums = append(x.Enums, p.parseEnum())
		if p.tok != tokn.COMMA {
			break
		}
		p.consume()
	}
	x.RBrace = p.expect(tokn.RBRACE)
	x.With = p.parseWith()
	return x
}

func (p *parser) parseEnum() ast.Expr {
	var firstIdent func(n ast.Expr) *ast.Ident
	firstIdent = func(n ast.Expr) *ast.Ident {
		switch n := n.(type) {
		case *ast.CallExpr:
			return firstIdent(n.Fun)
		case *ast.SelectorExpr:
			return firstIdent(n.X)
		case *ast.Ident:
			return n
		default:
			return nil
		}
	}
	x := p.parseExpr()
	if id := firstIdent(x); id != nil {
		p.names[id.String()] = true
		id.IsName = true
	}
	return x
}

/*************************************************************************
 * Behaviour Type Declaration
 *************************************************************************/

func (p *parser) parseBehaviourTypeDecl() *ast.BehaviourTypeDecl {
	if p.trace {
		defer un(trace(p, "ast.BehaviourTypeDecl"))
	}
	x := new(ast.BehaviourTypeDecl)
	x.TypeTok = p.consume()
	x.Kind = p.consume()
	x.Name = p.parseName()
	if p.tok == tokn.LT {
		x.TypePars = p.parseTypeFormalPars()
	}

	x.Params = p.parseFormalPars()
	if p.tok == tokn.RUNS {
		x.RunsOn = p.parseRunsOn()
	}

	if p.tok == tokn.SYSTEM {
		x.System = p.parseSystem()
	}

	if p.tok == tokn.RETURN {
		x.Return = p.parseReturn()
	}
	x.With = p.parseWith()
	return x
}

/*************************************************************************
 * Subtype
 *************************************************************************/

func (p *parser) parseSubTypeDecl() *ast.SubTypeDecl {
	if p.trace {
		defer un(trace(p, "ast.SubTypeDecl"))
	}
	x := new(ast.SubTypeDecl)
	x.TypeTok = p.consume()
	x.Field = p.parseField()
	x.With = p.parseWith()
	return x
}

func (p *parser) parseField() *ast.Field {
	if p.trace {
		defer un(trace(p, "Field"))
	}
	x := new(ast.Field)

	if p.tok == tokn.MODIF {
		if p.lit(1) != "@default" {
			p.errorExpected(p.pos(1), "@default")
		}
		x.DefaultTok = p.consume()
	}
	x.Type = p.parseTypeSpec()
	x.Name = p.parseName()
	if p.tok == tokn.LT {
		x.TypePars = p.parseTypeFormalPars()
	}

	if p.tok == tokn.LBRACK {
		x.ArrayDef = p.parseArrayDefs()
	}

	if p.tok == tokn.LPAREN {
		x.ValueConstraint = p.parseParenExpr()
	}
	if p.tok == tokn.LENGTH {
		x.LengthConstraint = p.parseLength(nil)
	}

	if p.tok == tokn.OPTIONAL {
		x.Optional = p.consume()
	}
	return x
}

func (p *parser) parseTypeSpec() ast.TypeSpec {
	if p.trace {
		defer un(trace(p, "TypeSpec"))
	}
	etok := p.peek(1)
	switch p.tok {
	case tokn.ADDRESS, tokn.CHARSTRING, tokn.IDENT, tokn.NULL, tokn.UNIVERSAL:
		return &ast.RefSpec{X: p.parseTypeRef()}
	case tokn.UNION:
		return p.parseStructSpec()
	case tokn.SET, tokn.RECORD:
		if p.peek(2).Kind() == tokn.LBRACE {
			return p.parseStructSpec()
		}
		return p.parseListSpec()
	case tokn.ENUMERATED:
		return p.parseEnumSpec()
	case tokn.FUNCTION, tokn.ALTSTEP, tokn.TESTCASE:
		return p.parseBehaviourSpec()
	default:
		p.errorExpected(p.pos(1), "type definition")
		return &ast.ErrorNode{From: etok, To: p.peek(1)}
	}
}

func (p *parser) parseStructSpec() *ast.StructSpec {
	if p.trace {
		defer un(trace(p, "StructSpec"))
	}
	x := new(ast.StructSpec)
	x.Kind = p.consume()
	x.LBrace = p.expect(tokn.LBRACE)
	for p.tok != tokn.RBRACE && p.tok != tokn.EOF {
		x.Fields = append(x.Fields, p.parseField())
		if p.tok != tokn.COMMA {
			break
		}
		p.consume()
	}
	x.RBrace = p.expect(tokn.RBRACE)
	return x
}

func (p *parser) parseEnumSpec() *ast.EnumSpec {
	if p.trace {
		defer un(trace(p, "ListSpec"))
	}
	x := new(ast.EnumSpec)
	x.Tok = p.consume()
	x.LBrace = p.expect(tokn.LBRACE)
	for p.tok != tokn.RBRACE && p.tok != tokn.EOF {
		x.Enums = append(x.Enums, p.parseEnum())
		if p.tok != tokn.COMMA {
			break
		}
		p.consume()
	}
	x.RBrace = p.expect(tokn.RBRACE)
	return x
}

func (p *parser) parseListSpec() *ast.ListSpec {
	if p.trace {
		defer un(trace(p, "ListSpec"))
	}
	x := new(ast.ListSpec)
	x.Kind = p.consume()
	if p.tok == tokn.LENGTH {
		x.Length = p.parseLength(nil)
	}
	x.OfTok = p.expect(tokn.OF)
	x.ElemType = p.parseTypeSpec()
	return x
}

func (p *parser) parseBehaviourSpec() *ast.BehaviourSpec {
	if p.trace {
		defer un(trace(p, "BehaviourSpec"))
	}

	x := new(ast.BehaviourSpec)
	x.Kind = p.consume()
	x.Params = p.parseFormalPars()

	if p.tok == tokn.RUNS {
		x.RunsOn = p.parseRunsOn()
	}

	if p.tok == tokn.SYSTEM {
		x.System = p.parseSystem()
	}

	if p.tok == tokn.RETURN {
		x.Return = p.parseReturn()
	}
	return x
}

/*************************************************************************
 * Template Declaration
 *************************************************************************/

func (p *parser) parseTemplateDecl() *ast.TemplateDecl {
	if p.trace {
		defer un(trace(p, "TemplateDecl"))
	}

	x := &ast.TemplateDecl{RestrictionSpec: &ast.RestrictionSpec{}}
	x.TemplateTok = p.consume()

	if p.tok == tokn.LPAREN {
		x.LParen = p.consume() // consume '('
		x.Tok = p.consume()    // consume omit/value/...
		x.RParen = p.expect(tokn.RPAREN)
	}

	if p.tok == tokn.MODIF {
		x.Modif = p.consume()
	}

	x.Type = p.parseTypeRef()
	if _, ok := x.Type.(*ast.ErrorNode); ok {
		return x
	}
	x.Name = p.parseName()
	if p.tok == tokn.LT {
		x.TypePars = p.parseTypeFormalPars()
	}
	if p.tok == tokn.LPAREN {
		x.Params = p.parseFormalPars()
	}
	if p.tok == tokn.MODIFIES {
		x.ModifiesTok = p.consume()
		x.Base = p.parsePrimaryExpr()
	}
	x.AssignTok = p.expect(tokn.ASSIGN)
	x.Value = p.parseExpr()
	x.With = p.parseWith()

	return x
}

/*************************************************************************
 * ast.Module ast.FormalPar
 *************************************************************************/

func (p *parser) parseModulePar() ast.Decl {
	if p.trace {
		defer un(trace(p, "ModulePar"))
	}

	tok := p.consume()

	// parse deprecated module parameter group
	if p.tok == tokn.LBRACE {
		x := &ast.ModuleParameterGroup{Tok: tok}
		x.LBrace = p.consume()
		for p.tok != tokn.RBRACE && p.tok != tokn.EOF {
			d := new(ast.ValueDecl)
			d.TemplateRestriction = p.parseRestrictionSpec()
			d.Type = p.parseTypeRef()
			d.Decls = p.parseDeclList()
			p.expectSemi(d.Decls[len(d.Decls)-1].LastTok())
			x.Decls = append(x.Decls, d)
		}
		x.RBrace = p.expect(tokn.RBRACE)
		x.With = p.parseWith()
		return x
	}

	x := &ast.ValueDecl{Kind: tok}
	x.TemplateRestriction = p.parseRestrictionSpec()
	x.Type = p.parseTypeRef()
	x.Decls = p.parseDeclList()
	x.With = p.parseWith()
	return x
}

/*************************************************************************
 * Value Declaration
 *************************************************************************/

func (p *parser) parseValueDecl() *ast.ValueDecl {
	if p.trace {
		defer un(trace(p, "ValueDecl"))
	}
	x := &ast.ValueDecl{Kind: p.consume()}
	x.TemplateRestriction = p.parseRestrictionSpec()

	if p.tok == tokn.MODIF {
		x.Modif = p.consume()
	}

	if x.Kind.Kind() != tokn.TIMER {
		x.Type = p.parseTypeRef()
	}
	x.Decls = p.parseDeclList()
	x.With = p.parseWith()
	return x
}

func (p *parser) parseRestrictionSpec() *ast.RestrictionSpec {
	switch p.tok {
	case tokn.TEMPLATE:
		x := new(ast.RestrictionSpec)
		x.TemplateTok = p.consume()
		if p.tok == tokn.LPAREN {
			x.LParen = p.consume()
			x.Tok = p.consume()
			x.RParen = p.expect(tokn.RPAREN)
		}
		return x

	case tokn.OMIT, tokn.VALUE, tokn.PRESENT:
		x := new(ast.RestrictionSpec)
		x.Tok = p.consume()
		return x
	default:
		return nil
	}
}

func (p *parser) parseDeclList() (list []*ast.Declarator) {
	if p.trace {
		defer un(trace(p, "DeclList"))
	}

	list = append(list, p.parseDeclarator())
	for p.tok == tokn.COMMA {
		p.consume()
		list = append(list, p.parseDeclarator())
	}
	return
}

func (p *parser) parseDeclarator() *ast.Declarator {
	x := &ast.Declarator{}
	x.Name = p.parseName()
	if p.tok == tokn.LBRACK {
		x.ArrayDef = p.parseArrayDefs()
	}
	if p.tok == tokn.ASSIGN {
		x.AssignTok = p.consume()
		x.Value = p.parseExpr()
	}
	return x
}

/*************************************************************************
 * Behaviour Declaration
 *************************************************************************/

func (p *parser) parseFuncDecl() *ast.FuncDecl {
	if p.trace {
		defer un(trace(p, "FuncDecl"))
	}

	x := new(ast.FuncDecl)
	x.Kind = p.consume()
	if p.tok == tokn.MODIF {
		x.Modif = p.consume()
	}

	x.Name = p.parseName()
	if p.tok == tokn.LT {
		x.TypePars = p.parseTypeFormalPars()
	}

	x.Params = p.parseFormalPars()

	if p.tok == tokn.RUNS {
		x.RunsOn = p.parseRunsOn()
	}

	if p.tok == tokn.MTC {
		x.Mtc = p.parseMtc()
	}

	if p.tok == tokn.SYSTEM {
		x.System = p.parseSystem()
	}

	if p.tok == tokn.RETURN {
		x.Return = p.parseReturn()
	}

	if p.tok == tokn.LBRACE {
		x.Body = p.parseBlockStmt()
	}

	x.With = p.parseWith()
	return x
}

/*************************************************************************
 * External Function Declaration
 *************************************************************************/

func (p *parser) parseExtFuncDecl() *ast.FuncDecl {
	if p.trace {
		defer un(trace(p, "ExtFuncDecl"))
	}

	x := new(ast.FuncDecl)
	x.External = p.consume()
	x.Kind = p.consume()
	if p.tok == tokn.MODIF {
		x.Modif = p.consume()
	}

	x.Name = p.parseName()

	x.Params = p.parseFormalPars()

	if p.tok == tokn.RUNS {
		x.RunsOn = p.parseRunsOn()
	}

	if p.tok == tokn.MTC {
		x.Mtc = p.parseMtc()
	}

	if p.tok == tokn.SYSTEM {
		x.System = p.parseSystem()
	}

	if p.tok == tokn.RETURN {
		x.Return = p.parseReturn()
	}
	x.With = p.parseWith()
	return x
}

/*************************************************************************
 * Signature Declaration
 *************************************************************************/

func (p *parser) parseSignatureDecl() *ast.SignatureDecl {
	if p.trace {
		defer un(trace(p, "SignatureDecl"))
	}

	x := new(ast.SignatureDecl)
	x.Tok = p.consume()
	x.Name = p.parseName()
	if p.tok == tokn.LT {
		x.TypePars = p.parseTypeFormalPars()
	}

	x.Params = p.parseFormalPars()

	if p.tok == tokn.NOBLOCK {
		x.NoBlock = p.consume()
	}

	if p.tok == tokn.RETURN {
		x.Return = p.parseReturn()
	}

	if p.tok == tokn.EXCEPTION {
		x.ExceptionTok = p.consume()
		x.Exception = p.parseParenExpr()
	}
	x.With = p.parseWith()
	return x
}

func (p *parser) parseRunsOn() *ast.RunsOnSpec {
	return &ast.RunsOnSpec{
		RunsTok: p.expect(tokn.RUNS),
		OnTok:   p.expect(tokn.ON),
		Comp:    p.parseTypeRef(),
	}
}

func (p *parser) parseSystem() *ast.SystemSpec {
	return &ast.SystemSpec{
		Tok:  p.expect(tokn.SYSTEM),
		Comp: p.parseTypeRef(),
	}
}

func (p *parser) parseMtc() *ast.MtcSpec {
	return &ast.MtcSpec{
		Tok:  p.expect(tokn.MTC),
		Comp: p.parseTypeRef(),
	}
}

func (p *parser) parseReturn() *ast.ReturnSpec {
	x := new(ast.ReturnSpec)
	x.Tok = p.consume()
	x.Restriction = p.parseRestrictionSpec()
	if p.tok == tokn.MODIF {
		x.Modif = p.consume()
	}
	x.Type = p.parseTypeRef()
	return x
}

func (p *parser) parseFormalPars() *ast.FormalPars {
	if p.trace {
		defer un(trace(p, "FormalPars"))
	}
	x := new(ast.FormalPars)
	x.LParen = p.expect(tokn.LPAREN)
	for p.tok != tokn.RPAREN {
		x.List = append(x.List, p.parseFormalPar())
		if p.tok != tokn.COMMA {
			break
		}
		p.consume()
	}
	x.RParen = p.expect(tokn.RPAREN)
	return x
}

func (p *parser) parseFormalPar() *ast.FormalPar {
	if p.trace {
		defer un(trace(p, "ast.FormalPar"))
	}
	x := new(ast.FormalPar)

	switch p.tok {
	case tokn.IN:
		x.Direction = p.consume()
	case tokn.OUT:
		x.Direction = p.consume()
	case tokn.INOUT:
		x.Direction = p.consume()
	}

	x.TemplateRestriction = p.parseRestrictionSpec()
	if p.tok == tokn.MODIF {
		x.Modif = p.consume()
	}
	x.Type = p.parseTypeRef()
	x.Name = p.parseName()

	if p.tok == tokn.LBRACK {
		x.ArrayDef = p.parseArrayDefs()
	}
	if p.tok == tokn.ASSIGN {
		x.AssignTok = p.consume()
		x.Value = p.parseExpr()
	}
	return x
}

func (p *parser) parseTypeFormalPars() *ast.FormalPars {
	if p.trace {
		defer un(trace(p, "TypeFormalPars"))
	}
	x := new(ast.FormalPars)
	x.LParen = p.expect(tokn.LT)
	for p.tok != tokn.GT {
		x.List = append(x.List, p.parseTypeFormalPar())
		if p.tok != tokn.COMMA {
			break
		}
		p.consume()
	}
	x.RParen = p.expect(tokn.GT)
	return x
}

func (p *parser) parseTypeFormalPar() *ast.FormalPar {
	if p.trace {
		defer un(trace(p, "TypeFormalPar"))
	}

	x := new(ast.FormalPar)

	if p.tok == tokn.IN {
		x.Direction = p.consume()
	}

	switch p.tok {
	case tokn.TYPE:
		x.Type = p.make_use(p.consume())
	case tokn.SIGNATURE:
		x.Type = p.make_use(p.consume())
	default:
		x.Type = p.parseTypeRef()
	}
	x.Name = p.make_use(p.expect(tokn.IDENT))
	x.Name.IsName = true
	if p.tok == tokn.ASSIGN {
		x.AssignTok = p.consume()
		x.Value = p.parseTypeRef()
	}

	return x
}

/*************************************************************************
 * Statements
 *************************************************************************/

func (p *parser) parseBlockStmt() *ast.BlockStmt {
	if p.trace {
		defer un(trace(p, "BlockStmt"))
	}

	x := &ast.BlockStmt{LBrace: p.expect(tokn.LBRACE)}
	for p.tok != tokn.RBRACE && p.tok != tokn.EOF {
		x.Stmts = append(x.Stmts, p.parseStmt())
		p.expectSemi(x.Stmts[len(x.Stmts)-1].LastTok())
	}
	x.RBrace = p.expect(tokn.RBRACE)
	return x
}

func (p *parser) parseStmt() ast.Stmt {
	if p.trace {
		defer un(trace(p, "ast.Stmt"))
	}

	etok := p.peek(1)
	switch p.tok {
	case tokn.TEMPLATE:
		return &ast.DeclStmt{Decl: p.parseTemplateDecl()}
	case tokn.VAR, tokn.CONST, tokn.TIMER, tokn.PORT:
		return &ast.DeclStmt{Decl: p.parseValueDecl()}
	case tokn.REPEAT, tokn.BREAK, tokn.CONTINUE:
		return &ast.BranchStmt{Tok: p.consume()}
	case tokn.LABEL:
		return &ast.BranchStmt{Tok: p.consume(), Label: p.make_use(p.expect(tokn.IDENT))}
	case tokn.GOTO:
		return &ast.BranchStmt{Tok: p.consume(), Label: p.make_use(p.expect(tokn.IDENT))}
	case tokn.RETURN:
		x := &ast.ReturnStmt{Tok: p.consume()}
		if !stmtStart[p.tok] && p.tok != tokn.SEMICOLON && p.tok != tokn.RBRACE {
			x.Result = p.parseExpr()
		}
		return x
	case tokn.SELECT:
		return p.parseSelect()
	case tokn.ALT, tokn.INTERLEAVE:
		alt := &ast.AltStmt{Tok: p.consume()}
		if p.tok == tokn.MODIF {
			alt.NoDefault = p.consume()
		}
		alt.Body = p.parseBlockStmt()
		return alt
	case tokn.LBRACK:
		return p.parseAltGuard()
	case tokn.FOR:
		return p.parseForLoop()
	case tokn.WHILE:
		return p.parseWhileLoop()
	case tokn.DO:
		return p.parseDoWhileLoop()
	case tokn.IF:
		return p.parseIfStmt()
	case tokn.LBRACE:
		return p.parseBlockStmt()
	case tokn.IDENT, tokn.TESTCASE, tokn.ANYKW, tokn.ALL, tokn.MAP, tokn.UNMAP, tokn.MTC:
		x := p.parseSimpleStmt()

		// try call-statement block
		if p.tok == tokn.LBRACE {
			c, ok := x.Expr.(*ast.CallExpr)
			if !ok {
				return x
			}
			s, ok := c.Fun.(*ast.SelectorExpr)
			if !ok {
				return x
			}
			id, ok := s.Sel.(*ast.Ident)
			if !ok {
				return x
			}
			if id.Tok.String() != "call" {
				return x
			}

			call := new(ast.CallStmt)
			call.Stmt = x
			call.Body = p.parseBlockStmt()
			return call
		}
		return x
	// Interpret simple literal expressions like integers or strings as statement.
	// This exception was added to help implementing ast-evaluator code like this:
	//
	//       if (1 > 2) { 10 } else { 20 }
	//
	case tokn.INT, tokn.FLOAT, tokn.STRING, tokn.BSTRING, tokn.TRUE, tokn.FALSE, tokn.PASS, tokn.FAIL, tokn.NONE, tokn.INCONC, tokn.ERROR:
		return p.parseSimpleStmt()
	default:
		p.errorExpected(p.pos(1), "statement")
		p.advance(stmtStart)
		return &ast.ErrorNode{From: etok, To: p.peek(1)}
	}
}

func (p *parser) parseForLoop() *ast.ForStmt {
	x := new(ast.ForStmt)
	x.Tok = p.consume()
	x.LParen = p.expect(tokn.LPAREN)
	if p.tok == tokn.VAR {
		x.Init = &ast.DeclStmt{Decl: p.parseValueDecl()}
	} else {
		x.Init = &ast.ExprStmt{Expr: p.parseExpr()}
	}
	x.InitSemi = p.expect(tokn.SEMICOLON)
	x.Cond = p.parseExpr()
	x.CondSemi = p.expect(tokn.SEMICOLON)
	x.Post = p.parseSimpleStmt()
	x.LParen = p.expect(tokn.RPAREN)
	x.Body = p.parseBlockStmt()
	return x
}

func (p *parser) parseWhileLoop() *ast.WhileStmt {
	return &ast.WhileStmt{
		Tok:  p.consume(),
		Cond: p.parseParenExpr(),
		Body: p.parseBlockStmt(),
	}
}

func (p *parser) parseDoWhileLoop() *ast.DoWhileStmt {
	return &ast.DoWhileStmt{
		DoTok:    p.consume(),
		Body:     p.parseBlockStmt(),
		WhileTok: p.expect(tokn.WHILE),
		Cond:     p.parseParenExpr(),
	}
}

func (p *parser) parseIfStmt() *ast.IfStmt {
	x := &ast.IfStmt{
		Tok:  p.consume(),
		Cond: p.parseParenExpr(),
		Then: p.parseBlockStmt(),
	}
	if p.tok == tokn.ELSE {
		x.ElseTok = p.consume()
		if p.tok == tokn.IF {
			x.Else = p.parseIfStmt()
		} else {
			x.Else = p.parseBlockStmt()
		}
	}
	return x
}

func (p *parser) parseSelect() *ast.SelectStmt {
	x := new(ast.SelectStmt)
	x.Tok = p.expect(tokn.SELECT)
	if p.tok == tokn.UNION {
		x.Union = p.consume()
	}
	x.Tag = p.parseParenExpr()
	x.LBrace = p.expect(tokn.LBRACE)
	for p.tok == tokn.CASE {
		x.Body = append(x.Body, p.parseCaseStmt())
	}
	x.RBrace = p.expect(tokn.RBRACE)
	return x
}

func (p *parser) parseCaseStmt() *ast.CaseClause {
	x := new(ast.CaseClause)
	x.Tok = p.expect(tokn.CASE)
	if p.tok == tokn.ELSE {
		p.consume() // TODO(5nord) move token into AST
	} else {
		x.Case = p.parseParenExpr()
	}
	x.Body = p.parseBlockStmt()
	return x
}

func (p *parser) parseAltGuard() *ast.CommClause {
	x := new(ast.CommClause)
	x.LBrack = p.expect(tokn.LBRACK)
	if p.tok == tokn.ELSE {
		x.Else = p.consume()
		x.RBrack = p.expect(tokn.RBRACK)
		x.Body = p.parseBlockStmt()
		return x
	}

	if p.tok != tokn.RBRACK {
		x.X = p.parseExpr()
	}
	x.RBrack = p.expect(tokn.RBRACK)
	x.Comm = p.parseSimpleStmt()
	if p.tok == tokn.LBRACE {
		x.Body = p.parseBlockStmt()
	}
	return x
}

func (p *parser) parseSimpleStmt() *ast.ExprStmt {
	if p.trace {
		defer un(trace(p, "SimpleStmt"))
	}

	return &ast.ExprStmt{Expr: p.parseExpr()}
}
