package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nokia/ntt/internal/errors"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/scanner"
	"github.com/nokia/ntt/internal/ttcn3/token"
)

// The parser structure holds the parser's internal state.
type parser struct {
	file    *loc.File
	errors  errors.ErrorList
	scanner scanner.Scanner

	// Tracing/debugging
	mode           Mode // parsing mode
	ignoreComments bool
	trace          bool // == (mode & Trace != 0)
	semi           bool // == (mode & PedanticSemicolon != 0)
	indent         int  // indentation used for tracing output

	// Tokens/Backtracking
	cursor  int
	tokens  []ast.Token
	trivia  []ast.Trivia
	markers []int
	tok     token.Kind // for convenience (p.tok is used frequently)

	// Required for unscanning ast.Token
	lastKind token.Kind
	lastPos  loc.Pos
	lastLit  string

	// Semicolon helper
	seenBrace bool

	// Pre-processor Handling
	ppLvl  int
	ppCnt  int
	ppSkip bool
	ppDefs map[string]bool

	// Error recovery
	// (used to limit the number of calls to advance
	// w/o making scanning progress - avoids potential endless
	// loops across multiple parser functions during error recovery)
	syncPos loc.Pos // last synchronization position
	syncCnt int     // number of advance calls without progress
}

func (p *parser) init(fset *loc.FileSet, filename string, src []byte, mode Mode) {
	p.file = fset.AddFile(filename, -1, len(src))

	eh := func(pos loc.Position, msg string) {
		p.errors.Add(pos, msg)
	}
	p.scanner.Init(p.file, src, eh)

	p.mode = mode
	p.trace = mode&Trace != 0 // for convenience (p.trace is used frequently)
	p.ignoreComments = mode&IgnoreComments != 0
	p.semi = mode&PedanticSemicolon != 0

	p.tokens = make([]ast.Token, 0, 200)
	p.markers = make([]int, 0, 200)
	p.ppDefs = make(map[string]bool)
	p.ppDefs["0"] = false
	p.ppDefs["1"] = true

	// fetch first token
	p.peek(1)
	p.tok = p.tokens[p.cursor].Kind
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
		p.error(p.pos(1), "unknown preprocessor directive")
	}
}

// Read the next token from input-stream
func (p *parser) scanToken() ast.Token {
	tok := p.lastKind
	pos := p.lastPos
	lit := p.lastLit

	if p.lastKind == token.Kind(0) {
	redo:
		pos, tok, lit = p.scanner.Scan()
		if tok == token.COMMENT && p.ignoreComments {
			goto redo
		}
	}
	p.lastKind = token.Kind(0)

	return ast.NewToken(pos, tok, lit)
}

// Unread the token
func (p *parser) unscanToken(tok ast.Token) {
	p.lastKind = tok.Kind
	p.lastPos = tok.Pos()
	p.lastLit = tok.Lit
}

func asTrivia(tok ast.Token) ast.Trivia {
	return ast.NewTrivia(tok.Pos(), tok.Kind, tok.Lit)
}

func (p *parser) scan() ast.Token {
	tok := p.scanToken()

	for ; p.isTrivia(tok); tok = p.scanToken() {
		if tok.Kind == token.PREPROC {
			p.handlePreproc(tok.Lit)
		}
		p.trivia = append(p.trivia, asTrivia(tok))
	}

	if tok.Kind == token.EOF {
		return tok
	}

	if len(p.trivia) != 0 {
		tok.LeadingTriv = p.trivia
		p.trivia = nil
	}

	// Try consume trailing trivia
	line := p.file.Line(tok.End())
	for {
		trail := p.scanToken()

		if !p.isTrivia(trail) {
			p.unscanToken(trail)
			break
		}

		if line != p.file.Line(trail.Pos()) {
			p.trivia = append(p.trivia, asTrivia(trail))
			break
		}

		tok.TrailingTriv = append(tok.TrailingTriv, asTrivia(trail))
	}

	return tok
}

func (p *parser) isTrivia(tok ast.Token) bool {
	switch {
	case tok.Kind == token.COMMENT:
		return true
	case tok.Kind == token.PREPROC:
		return true
	default:
		if p.ppSkip && tok.Kind != token.EOF {
			return true
		}
		return false
	}
}

// Advance to the next token
func (p *parser) consume() ast.Token {
	tok := p.tokens[p.cursor]
	if p.trace {
		s := tok.Kind.String()
		switch {
		case tok.Kind.IsLiteral():
			p.printTrace(s, tok.Lit)
		case tok.Kind.IsOperator(), tok.Kind.IsKeyword():
			p.printTrace("\"" + s + "\"")
		default:
			p.printTrace(s)
		}
	}

	// Track curly braces for TTCN-3 semicolon rules
	p.seenBrace = false
	if p.tok == token.RBRACE {
		p.seenBrace = true
	}

	p.cursor++
	if p.cursor == len(p.tokens) && !p.speculating() {
		p.cursor = 0
		p.tokens = p.tokens[:0]
	}

	p.peek(1)
	p.tok = p.tokens[p.cursor].Kind
	return tok
}

// Append current token to trailing trivia and advance to next token.
func (p *parser) consumeTrivia(tok *ast.Token) {
	triv := p.consume()

	// TODO(5nord) Remove when parser is ready
	if tok == nil {
		return
	}

	var trivs []ast.Trivia
	copy(trivs, triv.LeadingTriv)
	trivs = append(trivs, asTrivia(triv))
	trivs = append(trivs, triv.TrailingTriv...)
	tok.TrailingTriv = append(tok.TrailingTriv, trivs...)
}

func (p *parser) peek(i int) ast.Token {
	idx := p.cursor + i - 1
	last := len(p.tokens) - 1
	if idx > last {
		n := idx - last
		for i := 0; i < n; i++ {
			p.tokens = append(p.tokens, p.scan())
		}
	}
	return p.tokens[idx]
}

func (p *parser) pos(i int) loc.Pos {
	tok := p.peek(i)
	return tok.Pos()
}

func (p *parser) lit(i int) string {
	return p.peek(i).Lit
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
	p.tok = p.tokens[p.cursor].Kind
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
	if pos == p.pos(1) {
		// the error happened at the current position;
		// make the error message more specific
		switch {
		case p.tok.IsLiteral():
			// print 123 rather than 'INT', etc.
			msg += ", found " + p.lit(1)
		default:
			msg += ", found '" + p.tok.String() + "'"
		}
	}
	p.error(pos, msg)
}

func (p *parser) expect(k token.Kind) ast.Token {
	if p.tok != k {
		tok := p.peek(1)
		p.errorExpected(tok.Pos(), "'"+k.String()+"'")
	}
	return p.consume() // make progress
}

func (p *parser) expectSemi(tok *ast.Token) {
	if p.tok == token.SEMICOLON {
		p.consumeTrivia(tok)
		return
	}

	// pedantic semicolon
	if p.semi {
		// semicolon is optional before a closing '}'
		if !p.seenBrace && p.tok == token.RBRACE && p.tok != token.EOF {
			p.errorExpected(p.pos(1), "';'")
			p.advance(stmtStart)
		}
	}
}

// advance consumes tokens until the current token p.tok
// is in the 'to' set, or EOF. For error recovery.
func (p *parser) advance(to map[token.Kind]bool) {
	for ; p.tok != token.EOF; p.consume() {
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

var stmtStart = map[token.Kind]bool{
	token.ALT:        true,
	token.ALTSTEP:    true,
	token.BREAK:      true,
	token.CASE:       true,
	token.CONST:      true,
	token.CONTINUE:   true,
	token.CONTROL:    true,
	token.DISPLAY:    true,
	token.DO:         true,
	token.ELSE:       true,
	token.ENCODE:     true,
	token.EXTENSION:  true,
	token.FOR:        true,
	token.FRIEND:     true,
	token.FUNCTION:   true,
	token.GOTO:       true,
	token.GROUP:      true,
	token.IF:         true,
	token.IMPORT:     true,
	token.INTERLEAVE: true,
	token.LABEL:      true,
	token.MAP:        true,
	token.MODULE:     true,
	token.MODULEPAR:  true,
	token.PORT:       true,
	token.PRIVATE:    true,
	token.PUBLIC:     true,
	token.REPEAT:     true,
	token.RETURN:     true,
	token.SELECT:     true,
	token.SIGNATURE:  true,
	token.TEMPLATE:   true,
	token.TESTCASE:   true,
	token.TIMER:      true,
	token.TYPE:       true,
	token.UNMAP:      true,
	token.VAR:        true,
	token.VARIANT:    true,
	token.WHILE:      true,
}

var operandStart = map[token.Kind]bool{
	token.ADDRESS:    true,
	token.ALL:        true,
	token.ANY:        true,
	token.ANYKW:      true,
	token.BSTRING:    true,
	token.CHARSTRING: true,
	token.ERROR:      true,
	token.FAIL:       true,
	token.FALSE:      true,
	token.FLOAT:      true,
	//token.IDENT: true, TODO(5nord) fix conflict, see failing parser tests
	token.INCONC:    true,
	token.INT:       true,
	token.MAP:       true,
	token.MTC:       true,
	token.MUL:       true,
	token.NAN:       true,
	token.NONE:      true,
	token.NULL:      true,
	token.OMIT:      true,
	token.PASS:      true,
	token.STRING:    true,
	token.SYSTEM:    true,
	token.TESTCASE:  true,
	token.TIMER:     true,
	token.TRUE:      true,
	token.UNIVERSAL: true,
	token.UNMAP:     true,
}

// parse is a generic entry point
func (p *parser) parse() []ast.Node {
	switch p.tok {
	case token.MODULE:
		list := p.parseModuleList()
		nodes := make([]ast.Node, len(list))
		for i, d := range list {
			nodes[i] = d
		}
		return nodes
	case token.CONTROL,
		token.EXTERNAL,
		token.FRIEND,
		token.FUNCTION,
		token.GROUP,
		token.IMPORT,
		token.MODULEPAR,
		token.SIGNATURE,
		token.TEMPLATE,
		token.TYPE,
		token.VAR,
		token.ALTSTEP,
		token.CONST,
		token.PRIVATE,
		token.PUBLIC,
		token.TESTCASE:
		nodes := []ast.Node{p.parseModuleDef()}
		p.expect(token.EOF)
		return nodes
	default:
		list := p.parseExprList()
		nodes := make([]ast.Node, len(list))
		for i, d := range list {
			nodes[i] = d
		}
		return nodes
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
	for p.tok == token.COMMA {
		p.consumeTrivia(list[len(list)-1].LastTok())
		list = append(list, p.parseExpr())
	}
	return list
}

// ast.Expr ::= BinaryExpr
func (p *parser) parseExpr() ast.Expr {
	if p.trace {
		defer un(trace(p, "ast.Expr"))
	}

	return p.parseBinaryExpr(token.LowestPrec + 1)
}

// BinaryExpr ::= UnaryExpr
//              | BinaryExpr OP BinaryExpr
//
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
//             | ("-"|"+"|"!"|"not"|"not4b") UnaryExpr
//             | PrimaryExpr
//
func (p *parser) parseUnaryExpr() ast.Expr {
	switch p.tok {
	case token.ADD, token.EXCL, token.NOT, token.NOT4B, token.SUB:
		tok := p.consume()
		// handle unused expr '-'
		if tok.Kind == token.SUB {
			switch p.tok {
			case token.COMMA, token.SEMICOLON, token.RBRACE, token.RBRACK, token.RPAREN, token.EOF:
				return &ast.ValueLiteral{Tok: tok}
			}
		}
		return &ast.UnaryExpr{Op: tok, X: p.parseUnaryExpr()}
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
//                 ["param"            ExprList   ]
//                 ["sender"           PrimaryExpr]
//                 ["@index" ["value"] PrimaryExpr]
//                 ["timestamp"        PrimaryExpr]
//
func (p *parser) parsePrimaryExpr() ast.Expr {
	x := p.parseOperand()
L:
	for {
		switch p.tok {
		case token.DOT:
			x = p.parseSelectorExpr(x)
		case token.LBRACK:
			x = p.parseIndexExpr(x)
		case token.LPAREN:
			x = p.parseCallExpr(x)
			// Not supporting chained function calls like 'get().x'
			// eleminates conflicts with alt-guards.
			break
		default:
			break L
		}
	}

	if p.tok == token.LENGTH {
		x = p.parseLength(x)
	}

	if p.tok == token.IFPRESENT {
		x = &ast.UnaryExpr{Op: p.consume(), X: x}
	}

	if p.tok == token.TO || p.tok == token.FROM {
		x = &ast.BinaryExpr{X: x, Op: p.consume(), Y: p.parseExpr()}
	}

	if p.tok == token.REDIR {
		x = p.parseRedirect(x)
	}

	if p.tok == token.VALUE {
		x = &ast.ValueExpr{X: x, Tok: p.consume(), Y: p.parseExpr()}
	}

	if p.tok == token.PARAM {
		x = &ast.ParamExpr{X: x, Tok: p.consume(), Y: p.parseParenExpr()}
	}

	if p.tok == token.ALIVE {
		x = &ast.UnaryExpr{Op: p.consume(), X: x}
	}

	return x
}

// Operand ::= ("any"|"all") ("component"|"port"|"timer"|"from" PrimaryExpr)
//           | Literal
//           | Reference
//
// Literal ::= INT | STRING | BSTRING | FLOAT
//           | "?" | "*"
//           | "none" | "inconc" | "pass" | "fail" | "error"
//           | "true" | "false"
//           | "not_a_number"
//
// Reference ::= ID
//             | "address" | ["unviersal"] "charstring" | "timer"
//             | "null" | "omit"
//             | "mtc" | "system" | "testcase"
//             | "map" | "unmap"
//
func (p *parser) parseOperand() ast.Expr {
	etok := p.peek(1)
	switch p.tok {
	case token.ANYKW, token.ALL:
		tok := p.consume()
		switch p.tok {
		case token.COMPONENT, token.PORT, token.TIMER:
			return &ast.Ident{
				Tok:  tok,
				Tok2: p.consume(),
			}
		case token.FROM:
			return &ast.FromExpr{
				Kind:    tok,
				FromTok: p.consume(),
				X:       p.parsePrimaryExpr(),
			}
		}

		// Workaround for deprecated port-attribute 'all'
		if tok.Kind == token.ALL {
			return make_ident(tok)
		}

		p.errorExpected(p.pos(1), "'component', 'port', 'timer' or 'from'")

	case token.UNIVERSAL:
		return p.parseUniversalCharstring()

	case token.ADDRESS,
		token.CHARSTRING,
		token.MAP,
		token.MTC,
		token.SYSTEM,
		token.TESTCASE,
		token.TIMER,
		token.UNMAP:
		return make_ident(p.consume())

	case token.IDENT:
		return p.parseRef()

	case token.INT,
		token.ANY,
		token.BSTRING,
		token.ERROR,
		token.NULL,
		token.OMIT,
		token.FAIL,
		token.FALSE,
		token.FLOAT,
		token.INCONC,
		token.MUL,
		token.NAN,
		token.NONE,
		token.PASS,
		token.STRING,
		token.TRUE:
		return &ast.ValueLiteral{Tok: p.consume()}

	case token.LPAREN:
		// can be template `x := (1,2,3)`, but also artihmetic expression: `1*(2+3)`
		return p.parseParenExpr()

	case token.LBRACK:
		return p.parseIndexExpr(nil)

	case token.LBRACE:
		return p.parseCompositeLiteral()

	case token.MODIFIES:
		return &ast.ModifiesExpr{
			Tok:    p.consume(),
			X:      p.parsePrimaryExpr(),
			Assign: p.expect(token.ASSIGN),
			Y:      p.parseExpr(),
		}

	case token.REGEXP:
		return p.parseCallRegexp()

	case token.PATTERN:
		return p.parseCallPattern()

	case token.DECMATCH:
		return p.parseCallDecmatch()

	case token.MODIF:
		return p.parseDecodedModifier()

	default:
		p.errorExpected(p.pos(1), "operand")
	}

	return &ast.ErrorNode{From: etok, To: p.peek(1)}
}

func (p *parser) parseRef() ast.Expr {
	id := p.parseIdent()
	if p.tok != token.LT {
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
		LParen: p.expect(token.LPAREN),
		List:   p.parseExprList(),
		RParen: p.expect(token.RPAREN),
	}
}

func (p *parser) parseUniversalCharstring() *ast.Ident {
	return &ast.Ident{
		Tok:  p.expect(token.UNIVERSAL),
		Tok2: p.expect(token.CHARSTRING),
	}
}

func (p *parser) parseCompositeLiteral() *ast.CompositeLiteral {
	c := new(ast.CompositeLiteral)
	c.LBrace = p.expect(token.LBRACE)
	if p.tok != token.RBRACE {
		c.List = p.parseExprList()
	}
	c.RBrace = p.expect(token.RBRACE)
	return c
}

func (p *parser) parseCallRegexp() *ast.RegexpExpr {
	c := new(ast.RegexpExpr)
	c.Tok = p.expect(token.REGEXP)
	if p.tok == token.MODIF {
		c.NoCase = p.consume()
	}
	c.X = p.parseParenExpr()
	return c
}

func (p *parser) parseCallPattern() *ast.PatternExpr {
	c := new(ast.PatternExpr)
	c.Tok = p.expect(token.PATTERN)
	if p.tok == token.MODIF {
		c.NoCase = p.consume()
	}
	c.X = p.parseExpr()
	return c
}

func (p *parser) parseCallDecmatch() *ast.DecmatchExpr {
	c := new(ast.DecmatchExpr)
	c.Tok = p.expect(token.DECMATCH)
	if p.tok == token.LPAREN {
		c.Params = p.parseParenExpr()
	}
	c.X = p.parseExpr()
	return c
}

func (p *parser) parseDecodedModifier() *ast.DecodedExpr {
	d := new(ast.DecodedExpr)
	d.Tok = p.expect(token.MODIF)
	if d.Tok.Lit != "@decoded" {
		p.errorExpected(d.Tok.Pos(), "@decoded")
	}

	if p.tok == token.LPAREN {
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
		LBrack: p.expect(token.LBRACK),
		Index:  p.parseExpr(),
		RBrack: p.expect(token.RBRACK),
	}
}

func (p *parser) parseCallExpr(x ast.Expr) *ast.CallExpr {
	c := new(ast.CallExpr)
	c.Fun = x
	c.Args = new(ast.ParenExpr)
	c.Args.LParen = p.expect(token.LPAREN)
	if p.tok != token.RPAREN {
		switch p.tok {
		case token.TO, token.FROM, token.REDIR:
			var x ast.Expr
			if p.tok == token.TO || p.tok == token.FROM {
				// TODO: Shouldn't this be a FromExpr?
				x = &ast.BinaryExpr{X: x, Op: p.consume(), Y: p.parseExpr()}
			}
			if p.tok == token.REDIR {
				x = p.parseRedirect(x)
			}
			c.Args.List = []ast.Expr{x}
		default:
			c.Args.List = append(c.Args.List, p.parseExprList()...)
		}
	}
	c.Args.RParen = p.expect(token.RPAREN)
	return c
}

func (p *parser) parseLength(x ast.Expr) *ast.LengthExpr {
	return &ast.LengthExpr{
		X:    x,
		Len:  p.expect(token.LENGTH),
		Size: p.parseParenExpr(),
	}
}

func (p *parser) parseRedirect(x ast.Expr) *ast.RedirectExpr {

	r := &ast.RedirectExpr{
		X:   x,
		Tok: p.expect(token.REDIR),
	}

	if p.tok == token.VALUE {
		r.ValueTok = p.expect(token.VALUE)
		r.Value = p.parseExprList()
	}

	if p.tok == token.PARAM {
		r.ParamTok = p.expect(token.PARAM)
		r.Param = p.parseExprList()
	}

	if p.tok == token.SENDER {
		r.SenderTok = p.expect(token.SENDER)
		r.Sender = p.parsePrimaryExpr()
	}

	if p.tok == token.MODIF {
		if p.lit(1) != "@index" {
			p.errorExpected(p.pos(1), "@index")
		}

		r.IndexTok = p.consume()
		if p.tok == token.VALUE {
			r.IndexValueTok = p.consume()
		}
		r.Index = p.parsePrimaryExpr()
	}

	if p.tok == token.TIMESTAMP {
		r.TimestampTok = p.expect(token.TIMESTAMP)
		r.Timestamp = p.parsePrimaryExpr()
	}

	return r
}

func (p *parser) parseIdent() *ast.Ident {
	switch p.tok {
	case token.UNIVERSAL:
		return p.parseUniversalCharstring()
	case token.IDENT, token.ADDRESS, token.ALIVE, token.CHARSTRING:
		return &ast.Ident{Tok: p.consume()}
	default:
		p.expect(token.IDENT) // use expect() error handling
		return nil
	}
}

func (p *parser) parseRefList() []ast.Expr {
	l := make([]ast.Expr, 0, 1)
	for {
		l = append(l, p.parseTypeRef())
		if p.tok != token.COMMA {
			break
		}
		p.consumeTrivia(l[len(l)-1].LastTok()) // consume ','
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
	for p.tok != token.GT {
		y := p.tryTypeParameter()
		if y == nil {
			return nil
		}
		x.List = append(x.List, y)
		if p.tok != token.COMMA {
			break
		}
		p.consumeTrivia(x.List[len(x.List)-1].LastTok()) // consume ','
	}

	if p.tok != token.GT {
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
		case token.DOT:
			x = &ast.SelectorExpr{
				X:   x,
				Dot: p.consume(),
				Sel: p.tryTypeIdent(),
			}
			if x.(*ast.SelectorExpr).Sel == nil {
				return nil
			}
		case token.LBRACK:
			lbrack := p.consume()
			dash := p.consume()
			rbrack := p.consume()
			if dash.Kind != token.SUB || rbrack.Kind != token.RBRACK {
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
	case token.IDENT, token.ADDRESS, token.CHARSTRING:
		x = &ast.Ident{Tok: p.consume()}
	case token.UNIVERSAL:
		x = p.parseUniversalCharstring()
	default:
		return nil
	}

	if p.tok == token.LT {
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
	for p.tok == token.MODULE {
		if m := p.parseModule(); m != nil {
			list = append(list, m)
			p.expectSemi(m.LastTok())
		}
	}
	p.expect(token.EOF)
	return list
}

func (p *parser) parseModule() *ast.Module {
	if p.trace {
		defer un(trace(p, "ast.Module"))
	}

	m := new(ast.Module)
	m.Tok = p.expect(token.MODULE)
	m.Name = p.parseIdent()

	if p.tok == token.LANGUAGE {
		m.Language = p.parseLanguageSpec()
	}

	m.LBrace = p.expect(token.LBRACE)

	for p.tok != token.RBRACE && p.tok != token.EOF {
		m.Defs = append(m.Defs, p.parseModuleDef())
		p.expectSemi(m.Defs[len(m.Defs)-1].LastTok())
	}
	m.RBrace = p.expect(token.RBRACE)
	m.With = p.parseWith()
	return m
}

func (p *parser) parseLanguageSpec() *ast.LanguageSpec {
	l := new(ast.LanguageSpec)
	l.Tok = p.consume()
	for {
		l.List = append(l.List, p.expect(token.STRING))
		if p.tok != token.COMMA {
			break
		}
		p.consumeTrivia(l.List[len(l.List)-1].LastTok()) // consume ','
	}
	return l
}

func (p *parser) parseModuleDef() *ast.ModuleDef {
	m := new(ast.ModuleDef)
	switch p.tok {
	case token.PRIVATE, token.PUBLIC:
		m.Visibility = p.consume()
	case token.FRIEND:
		if p.peek(2).Kind != token.MODULE {
			m.Visibility = p.consume()
		}
	}

	etok := p.peek(1)
	switch p.tok {
	case token.IMPORT:
		m.Def = p.parseImport()
	case token.GROUP:
		m.Def = p.parseGroup()
	case token.FRIEND:
		m.Def = p.parseFriend()
	case token.TYPE:
		m.Def = p.parseTypeDecl()
	case token.TEMPLATE:
		m.Def = p.parseTemplateDecl()
	case token.MODULEPAR:
		m.Def = p.parseModulePar()
	case token.VAR, token.CONST:
		m.Def = p.parseValueDecl()
	case token.SIGNATURE:
		m.Def = p.parseSignatureDecl()
	case token.FUNCTION, token.TESTCASE, token.ALTSTEP:
		m.Def = p.parseFuncDecl()
	case token.CONTROL:
		m.Def = &ast.ControlPart{Tok: p.consume(), Body: p.parseBlockStmt(), With: p.parseWith()}
	case token.EXTERNAL:
		switch p.peek(2).Kind {
		case token.FUNCTION:
			m.Def = p.parseExtFuncDecl()
		case token.CONST:
			p.error(p.pos(1), "external constants not suppored")
			p.consumeTrivia(nil)
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
	return m
}

/*************************************************************************
 * Import Definition
 *************************************************************************/

func make_ident(tok ast.Token) ast.Expr {
	return &ast.Ident{Tok: tok}
}

func (p *parser) parseImport() *ast.ImportDecl {
	if p.trace {
		defer un(trace(p, "Import"))
	}

	x := new(ast.ImportDecl)
	x.ImportTok = p.consume()
	x.FromTok = p.expect(token.FROM)
	x.Module = p.parseIdent()

	if p.tok == token.LANGUAGE {
		x.Language = p.parseLanguageSpec()
	}

	switch p.tok {
	case token.ALL:
		y := &ast.DefKindExpr{}
		z := make_ident(p.consume())
		if p.tok == token.EXCEPT {
			z = &ast.ExceptExpr{
				X:         z,
				ExceptTok: p.consume(),
				LBrace:    p.expect(token.LBRACE),
				List:      p.parseExceptStmts(),
				RBrace:    p.expect(token.RBRACE),
			}
		}
		y.List = []ast.Expr{z}
		x.List = append(x.List, y)
	case token.LBRACE:
		x.LBrace = p.expect(token.LBRACE)
		for p.tok != token.RBRACE && p.tok != token.EOF {
			x.List = append(x.List, p.parseImportStmt())
			p.expectSemi(x.List[len(x.List)-1].LastTok())
		}
		x.RBrace = p.expect(token.RBRACE)
	default:
		p.errorExpected(p.pos(1), "'all' or import spec")
	}

	x.With = p.parseWith()

	return x
}

func (p *parser) parseImportStmt() *ast.DefKindExpr {
	x := new(ast.DefKindExpr)
	switch p.tok {
	case token.ALTSTEP, token.CONST, token.FUNCTION, token.MODULEPAR,
		token.SIGNATURE, token.TEMPLATE, token.TESTCASE, token.TYPE:
		x.Kind = p.consume()
		if p.tok == token.ALL {
			y := make_ident(p.consume())
			if p.tok == token.EXCEPT {
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
	case token.GROUP:
		x.Kind = p.consume()
		for {
			y := p.parseTypeRef()
			if p.tok == token.EXCEPT {
				y = &ast.ExceptExpr{
					X:         y,
					ExceptTok: p.consume(),
					LBrace:    p.expect(token.LBRACE),
					List:      p.parseExceptStmts(),
					RBrace:    p.expect(token.RBRACE),
				}
			}
			x.List = append(x.List, y)
			if p.tok != token.COMMA {
				break
			}
			p.consumeTrivia(y.LastTok()) // consume ','
		}
	case token.IMPORT:
		x.Kind = p.consume()
		x.List = []ast.Expr{make_ident(p.expect(token.ALL))}
	default:
		p.errorExpected(p.pos(1), "import definition qualifier")
		p.advance(stmtStart)
	}
	return x
}

func (p *parser) parseExceptStmts() []ast.Expr {
	var list []ast.Expr
	for p.tok != token.RBRACE && p.tok != token.EOF {
		x := p.parseExceptStmt()
		p.expectSemi(x.LastTok())
		list = append(list, x)
	}
	return list
}

func (p *parser) parseExceptStmt() *ast.DefKindExpr {
	x := new(ast.DefKindExpr)
	switch p.tok {
	case token.ALTSTEP, token.CONST, token.FUNCTION, token.GROUP,
		token.IMPORT, token.MODULEPAR, token.SIGNATURE, token.TEMPLATE,
		token.TESTCASE, token.TYPE:
		x.Kind = p.consume()
	default:
		p.errorExpected(p.pos(1), "definition qualifier")
	}

	if p.tok == token.ALL {
		x.List = []ast.Expr{make_ident(p.consume())}
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
	x.Name = p.parseIdent()
	x.LBrace = p.expect(token.LBRACE)

	for p.tok != token.RBRACE && p.tok != token.EOF {
		x.Defs = append(x.Defs, p.parseModuleDef())
		p.expectSemi(x.Defs[len(x.Defs)-1].LastTok())
	}
	x.RBrace = p.expect(token.RBRACE)
	x.With = p.parseWith()
	return x
}

func (p *parser) parseFriend() *ast.FriendDecl {
	return &ast.FriendDecl{
		FriendTok: p.expect(token.FRIEND),
		ModuleTok: p.expect(token.MODULE),
		Module:    p.parseIdent(),
		With:      p.parseWith(),
	}
}

/*************************************************************************
 * With Attributes
 *************************************************************************/

func (p *parser) parseWith() *ast.WithSpec {
	if p.tok != token.WITH {
		return nil
	}
	x := new(ast.WithSpec)
	x.Tok = p.consume()
	x.LBrace = p.expect(token.LBRACE)
	for p.tok != token.RBRACE && p.tok != token.EOF {
		x.List = append(x.List, p.parseWithStmt())
		p.expectSemi(x.List[len(x.List)-1].LastTok())
	}
	x.RBrace = p.expect(token.RBRACE)
	return x
}

func (p *parser) parseWithStmt() *ast.WithStmt {
	if p.trace {
		defer un(trace(p, "WithStmt"))
	}

	x := new(ast.WithStmt)

	switch p.tok {
	case token.ENCODE,
		token.VARIANT,
		token.DISPLAY,
		token.EXTENSION,
		token.OPTIONAL,
		token.STEPSIZE,
		token.OVERRIDE:
		x.Kind = p.consume()
	default:
		p.errorExpected(p.pos(1), "with-attribute")
		p.advance(stmtStart)
	}

	switch p.tok {
	case token.OVERRIDE:
		x.Override = p.consume()
	case token.MODIF:
		if p.lit(1) != "@local" {
			p.errorExpected(p.pos(1), "@local")
		}
		x.Override = p.consume()
	}

	if p.tok == token.LPAREN {
		x.LParen = p.consume()
		for {
			x.List = append(x.List, p.parseWithQualifier())
			if p.tok != token.COMMA {
				break
			}
			p.consumeTrivia(x.List[len(x.List)-1].LastTok())
		}
		x.RParen = p.expect(token.RPAREN)
	}

	var v ast.Expr = &ast.ValueLiteral{Tok: p.expect(token.STRING)}
	if p.tok == token.DOT {
		v = &ast.SelectorExpr{
			X:   v,
			Dot: p.consume(),
			Sel: &ast.ValueLiteral{Tok: p.expect(token.STRING)},
		}
	}
	x.Value = v

	return x
}

func (p *parser) parseWithQualifier() ast.Expr {
	etok := p.peek(1)
	switch p.tok {
	case token.IDENT:
		return p.parseTypeRef()
	case token.LBRACK:
		return p.parseIndexExpr(nil)
	case token.TYPE, token.TEMPLATE, token.CONST, token.ALTSTEP, token.TESTCASE, token.FUNCTION, token.SIGNATURE, token.MODULEPAR, token.GROUP:
		x := new(ast.DefKindExpr)
		x.Kind = p.consume()
		y := make_ident(p.expect(token.ALL))
		if p.tok == token.EXCEPT {
			y = &ast.ExceptExpr{
				X:         y,
				ExceptTok: p.consume(),
				LBrace:    p.expect(token.LBRACE),
				List:      p.parseRefList(),
				RBrace:    p.expect(token.RBRACE),
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
	switch p.peek(2).Kind {
	case token.IDENT, token.ADDRESS, token.CHARSTRING, token.NULL, token.UNIVERSAL:
		return p.parseSubTypeDecl()
	case token.PORT:
		return p.parsePortTypeDecl()
	case token.COMPONENT:
		return p.parseComponentTypeDecl()
	case token.UNION:
		return p.parseStructTypeDecl()
	case token.SET, token.RECORD:
		if p.peek(3).Kind == token.IDENT || p.peek(3).Kind == token.ADDRESS {
			return p.parseStructTypeDecl()
		}
		// lists are also parsed by parseSubTypeDecl
		return p.parseSubTypeDecl()
	case token.ENUMERATED:
		return p.parseEnumTypeDecl()
	case token.FUNCTION, token.ALTSTEP, token.TESTCASE:
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
	x.Name = p.parseIdent()
	if p.tok == token.LT {
		x.TypePars = p.parseTypeFormalPars()
	}

	switch p.tok {
	case token.MIXED, token.MESSAGE, token.PROCEDURE:
		x.Kind = p.consume()
	default:
		p.errorExpected(p.pos(1), "'message' or 'procedure'")
	}

	if p.tok == token.REALTIME {
		x.Realtime = p.consume()
	}

	x.LBrace = p.expect(token.LBRACE)
	for p.tok != token.RBRACE && p.tok != token.EOF {
		x.Attrs = append(x.Attrs, p.parsePortAttribute())
		p.expectSemi(x.Attrs[len(x.Attrs)-1].LastTok())
	}
	x.RBrace = p.expect(token.RBRACE)
	x.With = p.parseWith()
	return x
}

func (p *parser) parsePortAttribute() ast.Node {
	if p.trace {
		defer un(trace(p, "ast.PortAttribute"))
	}
	etok := p.peek(1)
	switch p.tok {
	case token.IN, token.OUT, token.INOUT, token.ADDRESS:
		return &ast.PortAttribute{
			Kind:  p.consume(),
			Types: p.parseRefList(),
		}
	case token.MAP, token.UNMAP:
		return &ast.PortMapAttribute{
			MapTok:   p.consume(),
			ParamTok: p.expect(token.PARAM),
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
	x.Name = p.parseIdent()
	if p.tok == token.LT {
		x.TypePars = p.parseTypeFormalPars()
	}
	if p.tok == token.EXTENDS {
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
	x.Name = p.parseIdent()
	if p.tok == token.LT {
		x.TypePars = p.parseTypeFormalPars()
	}
	x.LBrace = p.expect(token.LBRACE)
	for p.tok != token.RBRACE && p.tok != token.EOF {
		x.Fields = append(x.Fields, p.parseField())
		if p.tok != token.COMMA {
			break
		}
		p.consumeTrivia(x.Fields[len(x.Fields)-1].LastTok())
	}
	x.RBrace = p.expect(token.RBRACE)
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
	x.Name = p.parseIdent()
	if p.tok == token.LT {
		x.TypePars = p.parseTypeFormalPars()
	}
	x.LBrace = p.expect(token.LBRACE)
	for p.tok != token.RBRACE && p.tok != token.EOF {
		x.Enums = append(x.Enums, p.parseExpr())
		if p.tok != token.COMMA {
			break
		}
		p.consumeTrivia(x.Enums[len(x.Enums)-1].LastTok())
	}
	x.RBrace = p.expect(token.RBRACE)
	x.With = p.parseWith()
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
	x.Name = p.parseIdent()
	if p.tok == token.LT {
		x.TypePars = p.parseTypeFormalPars()
	}

	x.Params = p.parseFormalPars()
	if p.tok == token.RUNS {
		x.RunsOn = p.parseRunsOn()
	}

	if p.tok == token.SYSTEM {
		x.System = p.parseSystem()
	}

	if p.tok == token.RETURN {
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

	if p.tok == token.MODIF {
		if p.lit(1) != "@default" {
			p.errorExpected(p.pos(1), "@default")
		}
		x.DefaultTok = p.consume()
	}
	x.Type = p.parseTypeSpec()
	x.Name = p.parseIdent()
	if p.tok == token.LT {
		x.TypePars = p.parseTypeFormalPars()
	}

	// TODO(5nord) fix constraints consumed by previous PrimaryExpr
	if p.tok == token.LPAREN {
		x.ValueConstraint = p.parseParenExpr()
	}
	if p.tok == token.LENGTH {
		x.LengthConstraint = p.parseLength(nil)
	}

	if p.tok == token.OPTIONAL {
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
	case token.ADDRESS, token.CHARSTRING, token.IDENT, token.NULL, token.UNIVERSAL:
		return &ast.RefSpec{X: p.parseTypeRef()}
	case token.UNION:
		return p.parseStructSpec()
	case token.SET, token.RECORD:
		if p.peek(2).Kind == token.LBRACE {
			return p.parseStructSpec()
		}
		return p.parseListSpec()
	case token.ENUMERATED:
		return p.parseEnumSpec()
	case token.FUNCTION, token.ALTSTEP, token.TESTCASE:
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
	x.LBrace = p.expect(token.LBRACE)
	for p.tok != token.RBRACE && p.tok != token.EOF {
		x.Fields = append(x.Fields, p.parseField())
		if p.tok != token.COMMA {
			break
		}
		p.consumeTrivia(x.Fields[len(x.Fields)-1].LastTok())
	}
	x.RBrace = p.expect(token.RBRACE)
	return x
}

func (p *parser) parseEnumSpec() *ast.EnumSpec {
	if p.trace {
		defer un(trace(p, "ListSpec"))
	}
	x := new(ast.EnumSpec)
	x.Tok = p.consume()
	x.LBrace = p.expect(token.LBRACE)
	for p.tok != token.RBRACE && p.tok != token.EOF {
		x.Enums = append(x.Enums, p.parseExpr())
		if p.tok != token.COMMA {
			break
		}
		p.consumeTrivia(x.Enums[len(x.Enums)-1].LastTok())
	}
	x.RBrace = p.expect(token.RBRACE)
	return x
}

func (p *parser) parseListSpec() *ast.ListSpec {
	if p.trace {
		defer un(trace(p, "ListSpec"))
	}
	x := new(ast.ListSpec)
	x.Kind = p.consume()
	if p.tok == token.LENGTH {
		x.Length = p.parseLength(nil)
	}
	x.OfTok = p.expect(token.OF)
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

	if p.tok == token.RUNS {
		x.RunsOn = p.parseRunsOn()
	}

	if p.tok == token.SYSTEM {
		x.System = p.parseSystem()
	}

	if p.tok == token.RETURN {
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

	x := new(ast.TemplateDecl)
	x.TemplateTok = p.consume()

	if p.tok == token.LPAREN {
		x.LParen = p.consume() // consume '('
		x.Tok = p.consume()    // consume omit/value/...
		x.RParen = p.expect(token.RPAREN)
	}

	if p.tok == token.MODIF {
		x.Modif = p.consume()
	}

	x.Type = p.parseTypeRef()
	x.Name = p.parseIdent()
	if p.tok == token.LT {
		x.TypePars = p.parseTypeFormalPars()
	}
	if p.tok == token.LPAREN {
		x.Params = p.parseFormalPars()
	}
	if p.tok == token.MODIFIES {
		x.ModifiesTok = p.consume()
		x.Base = p.parsePrimaryExpr()
	}
	x.AssignTok = p.expect(token.ASSIGN)
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
	if p.tok == token.LBRACE {
		x := &ast.ModuleParameterGroup{Tok: tok}
		x.LBrace = p.consume()
		for p.tok != token.RBRACE && p.tok != token.EOF {
			d := new(ast.ValueDecl)
			d.TemplateRestriction = p.parseRestrictionSpec()
			d.Type = p.parseTypeRef()
			d.Decls = p.parseExprList()
			p.expectSemi(d.Decls[len(d.Decls)-1].LastTok())
			x.Decls = append(x.Decls, d)
		}
		x.RBrace = p.expect(token.RBRACE)
		x.With = p.parseWith()
		return x
	}

	x := &ast.ValueDecl{Kind: tok}
	x.TemplateRestriction = p.parseRestrictionSpec()
	x.Type = p.parseTypeRef()
	x.Decls = p.parseExprList()
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

	if p.tok == token.MODIF {
		x.Modif = p.consume()
	}

	if x.Kind.Kind != token.TIMER {
		x.Type = p.parseTypeRef()
	}
	x.Decls = p.parseExprList()
	x.With = p.parseWith()
	return x
}

func (p *parser) parseRestrictionSpec() *ast.RestrictionSpec {
	x := new(ast.RestrictionSpec)
	switch p.tok {
	case token.TEMPLATE:
		x.TemplateTok = p.consume()
		if p.tok != token.LPAREN {
			return nil
		}

		x.LParen = p.consume()
		x.Tok = p.consume()
		x.RParen = p.expect(token.RPAREN)

	case token.OMIT, token.VALUE, token.PRESENT:
		x.Tok = p.consume()
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
	if p.tok == token.MODIF {
		x.Modif = p.consume()
	}

	x.Name = p.parseIdent()
	if p.tok == token.LT {
		x.TypePars = p.parseTypeFormalPars()
	}

	x.Params = p.parseFormalPars()

	if p.tok == token.RUNS {
		x.RunsOn = p.parseRunsOn()
	}

	if p.tok == token.MTC {
		x.Mtc = p.parseMtc()
	}

	if p.tok == token.SYSTEM {
		x.System = p.parseSystem()
	}

	if p.tok == token.RETURN {
		x.Return = p.parseReturn()
	}

	if p.tok == token.LBRACE {
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
	if p.tok == token.MODIF {
		x.Modif = p.consume()
	}

	x.Name = p.parseIdent()

	x.Params = p.parseFormalPars()

	if p.tok == token.RUNS {
		x.RunsOn = p.parseRunsOn()
	}

	if p.tok == token.MTC {
		x.Mtc = p.parseMtc()
	}

	if p.tok == token.SYSTEM {
		x.System = p.parseSystem()
	}

	if p.tok == token.RETURN {
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
	x.Name = p.parseIdent()
	if p.tok == token.LT {
		x.TypePars = p.parseTypeFormalPars()
	}

	x.Params = p.parseFormalPars()

	if p.tok == token.NOBLOCK {
		x.NoBlock = p.consume()
	}

	if p.tok == token.RETURN {
		x.Return = p.parseReturn()
	}

	if p.tok == token.EXCEPTION {
		x.ExceptionTok = p.consume()
		x.Exception = p.parseParenExpr()
	}
	x.With = p.parseWith()
	return x
}

func (p *parser) parseRunsOn() *ast.RunsOnSpec {
	return &ast.RunsOnSpec{
		RunsTok: p.expect(token.RUNS),
		OnTok:   p.expect(token.ON),
		Comp:    p.parseTypeRef(),
	}
}

func (p *parser) parseSystem() *ast.SystemSpec {
	return &ast.SystemSpec{
		Tok:  p.expect(token.SYSTEM),
		Comp: p.parseTypeRef(),
	}
}

func (p *parser) parseMtc() *ast.MtcSpec {
	return &ast.MtcSpec{
		Tok:  p.expect(token.MTC),
		Comp: p.parseTypeRef(),
	}
}

func (p *parser) parseReturn() *ast.ReturnSpec {
	x := new(ast.ReturnSpec)
	x.Tok = p.consume()
	x.Restriction = p.parseRestrictionSpec()
	if p.tok == token.MODIF {
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
	x.LParen = p.expect(token.LPAREN)
	for p.tok != token.RPAREN {
		x.List = append(x.List, p.parseFormalPar())
		if p.tok != token.COMMA {
			break
		}
		p.consumeTrivia(x.List[len(x.List)-1].LastTok())
	}
	x.RParen = p.expect(token.RPAREN)
	return x
}

func (p *parser) parseFormalPar() *ast.FormalPar {
	if p.trace {
		defer un(trace(p, "ast.FormalPar"))
	}
	x := new(ast.FormalPar)

	switch p.tok {
	case token.IN:
		x.Direction = p.consume()
	case token.OUT:
		x.Direction = p.consume()
	case token.INOUT:
		x.Direction = p.consume()
	}

	x.TemplateRestriction = p.parseRestrictionSpec()
	if p.tok == token.MODIF {
		x.Modif = p.consume()
	}
	x.Type = p.parseTypeRef()
	x.Name = p.parseExpr()

	return x
}

func (p *parser) parseTypeFormalPars() *ast.FormalPars {
	if p.trace {
		defer un(trace(p, "TypeFormalPars"))
	}
	x := new(ast.FormalPars)
	x.LParen = p.expect(token.LT)
	for p.tok != token.GT {
		x.List = append(x.List, p.parseTypeFormalPar())
		if p.tok != token.COMMA {
			break
		}
		p.consumeTrivia(x.List[len(x.List)-1].LastTok())
	}
	x.RParen = p.expect(token.GT)
	return x
}

func (p *parser) parseTypeFormalPar() *ast.FormalPar {
	if p.trace {
		defer un(trace(p, "TypeFormalPar"))
	}

	x := new(ast.FormalPar)

	if p.tok == token.IN {
		x.Direction = p.consume()
	}

	switch p.tok {
	case token.TYPE:
		x.Type = make_ident(p.consume())
	case token.SIGNATURE:
		x.Type = make_ident(p.consume())
	default:
		x.Type = p.parseTypeRef()
	}
	x.Name = make_ident(p.expect(token.IDENT))
	if p.tok == token.ASSIGN {
		x.Name = &ast.BinaryExpr{
			X:  x.Name,
			Op: p.consume(),
			Y:  p.parseTypeRef(),
		}
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

	x := &ast.BlockStmt{LBrace: p.expect(token.LBRACE)}
	for p.tok != token.RBRACE && p.tok != token.EOF {
		x.Stmts = append(x.Stmts, p.parseStmt())
		p.expectSemi(x.Stmts[len(x.Stmts)-1].LastTok())
	}
	x.RBrace = p.expect(token.RBRACE)
	return x
}

func (p *parser) parseStmt() ast.Stmt {
	if p.trace {
		defer un(trace(p, "ast.Stmt"))
	}

	etok := p.peek(1)
	switch p.tok {
	case token.TEMPLATE:
		return &ast.DeclStmt{p.parseTemplateDecl()}
	case token.VAR, token.CONST, token.TIMER, token.PORT:
		return &ast.DeclStmt{p.parseValueDecl()}
	case token.REPEAT, token.BREAK, token.CONTINUE:
		return &ast.BranchStmt{Tok: p.consume()}
	case token.LABEL:
		return &ast.BranchStmt{Tok: p.consume(), Label: p.expect(token.IDENT)}
	case token.GOTO:
		return &ast.BranchStmt{Tok: p.consume(), Label: p.expect(token.IDENT)}
	case token.RETURN:
		x := &ast.ReturnStmt{Tok: p.consume()}
		if !stmtStart[p.tok] && p.tok != token.SEMICOLON && p.tok != token.RBRACE {
			x.Result = p.parseExpr()
		}
		return x
	case token.SELECT:
		return p.parseSelect()
	case token.ALT, token.INTERLEAVE:
		return &ast.AltStmt{Tok: p.consume(), Body: p.parseBlockStmt()}
	case token.LBRACK:
		return p.parseAltGuard()
	case token.FOR:
		return p.parseForLoop()
	case token.WHILE:
		return p.parseWhileLoop()
	case token.DO:
		return p.parseDoWhileLoop()
	case token.IF:
		return p.parseIfStmt()
	case token.LBRACE:
		return p.parseBlockStmt()
	case token.IDENT, token.TESTCASE, token.ANYKW, token.ALL, token.MAP, token.UNMAP, token.MTC:
		x := p.parseSimpleStmt()

		// try call-statement block
		if p.tok == token.LBRACE {
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
			if id.Tok.Lit != "call" {
				return x
			}

			call := new(ast.CallStmt)
			call.Stmt = x
			call.Body = p.parseBlockStmt()
			return call
		}
		return x
	default:
		p.errorExpected(p.pos(1), "statement")
		p.advance(stmtStart)
		return &ast.ErrorNode{From: etok, To: p.peek(1)}
	}
}

func (p *parser) parseForLoop() *ast.ForStmt {
	x := new(ast.ForStmt)
	x.Tok = p.consume()
	x.LParen = p.expect(token.LPAREN)
	if p.tok == token.VAR {
		x.Init = &ast.DeclStmt{Decl: p.parseValueDecl()}
	} else {
		x.Init = &ast.ExprStmt{Expr: p.parseExpr()}
	}
	x.InitSemi = p.expect(token.SEMICOLON)
	x.Cond = p.parseExpr()
	x.CondSemi = p.expect(token.SEMICOLON)
	x.Post = p.parseSimpleStmt()
	x.LParen = p.expect(token.RPAREN)
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
		WhileTok: p.expect(token.WHILE),
		Cond:     p.parseParenExpr(),
	}
}

func (p *parser) parseIfStmt() *ast.IfStmt {
	x := &ast.IfStmt{
		Tok:  p.consume(),
		Cond: p.parseParenExpr(),
		Then: p.parseBlockStmt(),
	}
	if p.tok == token.ELSE {
		x.ElseTok = p.consume()
		if p.tok == token.IF {
			x.Else = p.parseIfStmt()
		} else {
			x.Else = p.parseBlockStmt()
		}
	}
	return x
}

func (p *parser) parseSelect() *ast.SelectStmt {
	x := new(ast.SelectStmt)
	x.Tok = p.expect(token.SELECT)
	if p.tok == token.UNION {
		x.Union = p.consume()
	}
	x.Tag = p.parseParenExpr()
	x.LBrace = p.expect(token.LBRACE)
	for p.tok == token.CASE {
		x.Body = append(x.Body, p.parseCaseStmt())
	}
	x.RBrace = p.expect(token.RBRACE)
	return x
}

func (p *parser) parseCaseStmt() *ast.CaseClause {
	x := new(ast.CaseClause)
	x.Tok = p.expect(token.CASE)
	if p.tok == token.ELSE {
		p.consume() // TODO(5nord) move token into AST
	} else {
		x.Case = p.parseParenExpr()
	}
	x.Body = p.parseBlockStmt()
	return x
}

func (p *parser) parseAltGuard() *ast.CommClause {
	x := new(ast.CommClause)
	x.LBrack = p.expect(token.LBRACK)
	if p.tok == token.ELSE {
		x.Else = p.consume()
		x.RBrack = p.expect(token.RBRACK)
		x.Body = p.parseBlockStmt()
		return x
	}

	if p.tok != token.RBRACK {
		x.X = p.parseExpr()
	}
	x.RBrack = p.expect(token.RBRACK)
	x.Comm = p.parseSimpleStmt()
	if p.tok == token.LBRACE {
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
