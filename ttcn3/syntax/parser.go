package syntax

import (
	"fmt"
	"strings"
)

// A Mode value is a set of flags (or 0).
// They control the amount of source code parsed and other optional
// parser functionality.
//
type Mode uint

const (
	ImportsOnly       = 1 << iota      // stop parsing after import declarations
	PedanticSemicolon                  // expect semicolons pedantically
	ParseComments                      // parse comments and add them to AST
	Trace                              // print a trace of parsed productions
	DeclarationErrors                  // report declaration errors
	SpuriousErrors                     // same as AllErrors, for backward-compatibility
	AllErrors         = SpuriousErrors // report all errors (not just the first 10 on different lines)
)

// The parser structure holds the parser's internal state.
type parser struct {
	file    *File
	errors  ErrorList
	scanner Scanner

	// Tracing/debugging
	mode   Mode // parsing mode
	trace  bool // == (mode & Trace != 0)
	semi   bool // == (mode & PedanticSemicolon != 0)
	indent int  // indentation used for tracing output

	// Tokens/Backtracking
	cursor  int
	tokens  []Token
	markers []int
	tok     Kind // for convenience (p.tok is used frequently)

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
	syncPos Pos // last synchronization position
	syncCnt int // number of advance calls without progress
}

func (p *parser) init(fset *FileSet, filename string, src []byte, mode Mode, eh ErrorHandler) {
	p.file = fset.AddFile(filename, -1, len(src))

	eh2 := func(pos Position, msg string) {
		if eh != nil {
			eh(pos, msg)
		}
		p.errors.Add(pos, msg)
	}
	p.scanner.Init(p.file, src, eh2)

	p.mode = mode
	p.trace = mode&Trace != 0 // for convenience (p.trace is used frequently)
	p.semi = mode&PedanticSemicolon != 0

	p.tokens = make([]Token, 0, 200)
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
		p.error(p.pos(1), "'#define' is not supported")
	default:
		p.error(p.pos(1), "unknown preprocessor directive")
	}
}

// Read the next token from input-stream
func (p *parser) readKind() Token {
redo:
	pos, tok, lit := p.scanner.Scan()

	if tok == COMMENT {
		goto redo
	}

	if tok == PREPROC {
		p.handlePreproc(lit)
		goto redo
	}

	if p.ppSkip && tok != EOF {
		goto redo
	}

	return Token{pos, tok, lit}
}

// Advance to the next token
func (p *parser) consume() Token {
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
	if p.tok == RBRACE {
		p.seenBrace = true
	}

	p.cursor++
	if p.cursor == len(p.tokens) && !p.speculating() {
		p.cursor = 0
		p.tokens = p.tokens[:0]
	}

	p.grow(1)
	p.tok = p.tokens[p.cursor].Kind
	return tok
}

func (p *parser) grow(i int) {
	idx := p.cursor + i - 1
	last := len(p.tokens) - 1
	if idx > last {
		n := idx - last
		for i := 0; i < n; i++ {
			p.tokens = append(p.tokens, p.readKind())
		}
	}
}

func (p *parser) peek(i int) Token {
	p.grow(i)
	return p.tokens[p.cursor+i-1]
}

func (p *parser) pos(i int) Pos {
	return p.peek(i).Pos
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

func (p *parser) error(pos Pos, msg string) {
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

	if p.scanner.Err != nil {
		p.scanner.Err(epos, msg)
	}
	p.errors.Add(epos, msg)
}

func (p *parser) errorExpected(pos Pos, msg string) {
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

func (p *parser) expect(k Kind) Token {
	if p.tok != k {
		pos := p.peek(1).Pos
		p.errorExpected(pos, "'"+k.String()+"'")
	}
	return p.consume() // make progress
}

func (p *parser) expectSemi() {
	if p.tok == SEMICOLON {
		p.consume()
		return
	}

	// pedantic semicolon
	if p.semi {
		// semicolon is optional before a closing '}'
		if !p.seenBrace && p.tok == RBRACE && p.tok != EOF {
			p.errorExpected(p.pos(1), "';'")
			p.advance(stmtStart)
		}
	}
}

// advance consumes tokens until the current token p.tok
// is in the 'to' set, or EOF. For error recovery.
func (p *parser) advance(to map[Kind]bool) {
	for ; p.tok != EOF; p.consume() {
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

var stmtStart = map[Kind]bool{
	ALT:        true,
	ALTSTEP:    true,
	BREAK:      true,
	CASE:       true,
	CONST:      true,
	CONTINUE:   true,
	CONTROL:    true,
	DISPLAY:    true,
	DO:         true,
	ELSE:       true,
	ENCODE:     true,
	EXTENSION:  true,
	FOR:        true,
	FRIEND:     true,
	FUNCTION:   true,
	GOTO:       true,
	GROUP:      true,
	IF:         true,
	IMPORT:     true,
	INTERLEAVE: true,
	LABEL:      true,
	MAP:        true,
	MODULE:     true,
	MODULEPAR:  true,
	PORT:       true,
	PRIVATE:    true,
	PUBLIC:     true,
	REPEAT:     true,
	RETURN:     true,
	SELECT:     true,
	SIGNATURE:  true,
	TEMPLATE:   true,
	TESTCASE:   true,
	TIMER:      true,
	TYPE:       true,
	UNMAP:      true,
	VAR:        true,
	VARIANT:    true,
	WHILE:      true,
}

var operandStart = map[Kind]bool{
	ADDRESS:    true,
	ALL:        true,
	ANY:        true,
	ANYKW:      true,
	BSTRING:    true,
	CHARSTRING: true,
	ERROR:      true,
	FAIL:       true,
	FALSE:      true,
	FLOAT:      true,
	//IDENT: true, TODO(5nord) fix conflict, see failing parser tests
	INCONC:    true,
	INT:       true,
	MAP:       true,
	MTC:       true,
	MUL:       true,
	NAN:       true,
	NONE:      true,
	NULL:      true,
	OMIT:      true,
	PASS:      true,
	STRING:    true,
	SYSTEM:    true,
	TESTCASE:  true,
	TIMER:     true,
	TRUE:      true,
	UNIVERSAL: true,
	UNMAP:     true,
}

/*************************************************************************
 * Expressions
 *************************************************************************/

// ExprList ::= Expr { "," Expr }
func (p *parser) parseExprList() (list []Expr) {
	if p.trace {
		defer un(trace(p, "ExprList"))
	}

	list = append(list, p.parseExpr())
	for p.tok == COMMA {
		p.consume()
		list = append(list, p.parseExpr())
	}
	return list
}

// Expr ::= BinaryExpr
func (p *parser) parseExpr() Expr {
	if p.trace {
		defer un(trace(p, "Expr"))
	}

	return p.parseBinaryExpr(LowestPrec + 1)
}

// BinaryExpr ::= UnaryExpr
//              | BinaryExpr OP BinaryExpr
//
func (p *parser) parseBinaryExpr(prec1 int) Expr {
	x := p.parseUnaryExpr()
	for {
		prec := p.tok.Precedence()
		if prec < prec1 {
			return x
		}
		op := p.consume()
		y := p.parseBinaryExpr(prec + 1)
		x = &BinaryExpr{X: x, Op: op, Y: y}
	}
}

// UnaryExpr ::= "-"
//             | ("-"|"+"|"!"|"not"|"not4b") UnaryExpr
//             | PrimaryExpr
//
func (p *parser) parseUnaryExpr() Expr {
	switch p.tok {
	case ADD, EXCL, NOT, NOT4B, SUB:
		tok := p.consume()
		// handle unused expr '-'
		if tok.Kind == SUB {
			switch p.tok {
			case COMMA, SEMICOLON, RBRACE, RBRACK, RPAREN, EOF:
				return &ValueLiteral{Tok: tok}
			}
		}
		return &UnaryExpr{Op: tok, X: p.parseUnaryExpr()}
	}

	return p.parsePrimaryExpr()
}

// PrimaryExpr ::= Operand [{ExtFieldRef}] [Stuff]
//
// ExtFieldRef ::= "." ID
//               | "[" Expr "]"
//               | "(" ExprList ")"
//
// Stuff       ::= ["length"      "(" ExprList ")"]
//                 ["ifpresent"                   ]
//                 [("to"|"from") Expr            ]
//                 ["->"          Redirect        ]

// Redirect    ::= ["value"            ExprList   ]
//                 ["param"            ExprList   ]
//                 ["sender"           PrimaryExpr]
//                 ["@index" ["value"] PrimaryExpr]
//                 ["timestamp"        PrimaryExpr]
//
func (p *parser) parsePrimaryExpr() Expr {
	x := p.parseOperand()
L:
	for {
		switch p.tok {
		case DOT:
			x = p.parseSelectorExpr(x)
		case LBRACK:
			x = p.parseIndexExpr(x)
		case LPAREN:
			x = p.parseCallExpr(x)
		default:
			break L
		}
	}

	if p.tok == LENGTH {
		x = p.parseLength(x)
	}

	if p.tok == IFPRESENT {
		x = &UnaryExpr{Op: p.consume(), X: x}
	}

	if p.tok == TO || p.tok == FROM {
		x = &BinaryExpr{X: x, Op: p.consume(), Y: p.parseExpr()}
	}

	if p.tok == REDIR {
		x = p.parseRedirect(x)
	}

	if p.tok == VALUE {
		x = &ValueExpr{X: x, Tok: p.consume(), Y: p.parseExpr()}
	}

	if p.tok == PARAM {
		x = &ParamExpr{X: x, Tok: p.consume(), Y: p.parseParenExpr()}
	}

	if p.tok == ALIVE {
		x = &UnaryExpr{Op: p.consume(), X: x}
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
func (p *parser) parseOperand() Expr {
	switch p.tok {
	case ANYKW, ALL:
		tok := p.consume()
		switch p.tok {
		case COMPONENT, PORT, TIMER:
			// TODO(5nord): make this expression identifier?
			p.consume()
			return nil
		case FROM:
			p.consume() // TODO(5nord) move 'from' into AST
			return &FromExpr{Kind: tok, X: p.parsePrimaryExpr()}
		}

		// Workaround for deprecated port-attribute 'all'
		if tok.Kind == ALL {
			return &Ident{Tok: tok}
		}

		p.errorExpected(p.pos(1), "'component', 'port', 'timer' or 'from'")

	case UNIVERSAL:
		return p.parseUniversalCharstring()

	case ADDRESS,
		CHARSTRING,
		MAP,
		MTC,
		NULL,
		OMIT,
		SYSTEM,
		TESTCASE,
		TIMER,
		UNMAP:
		return &Ident{Tok: p.consume()}

	case IDENT:
		return p.parseRef()

	case INT,
		ANY,
		BSTRING,
		ERROR,
		FAIL,
		FALSE,
		FLOAT,
		INCONC,
		MUL,
		NAN,
		NONE,
		PASS,
		STRING,
		TRUE:
		return &ValueLiteral{Tok: p.consume()}

	case LPAREN:
		// can be template `x := (1,2,3)`, but also artihmetic expression: `1*(2+3)`
		return p.parseParenExpr()

	case LBRACK:
		return p.parseIndexExpr(nil)

	case LBRACE:
		return p.parseCompositeLiteral()

	case MODIFIES:
		return &ModifiesExpr{
			Tok:    p.consume(),
			X:      p.parsePrimaryExpr(),
			Assign: p.expect(ASSIGN),
			Y:      p.parseExpr(),
		}

	case REGEXP:
		return p.parseCallRegexp()

	case PATTERN:
		return p.parseCallPattern()

	case DECMATCH:
		return p.parseCallDecmatch()

	case MODIF:
		return p.parseDecodedModifier()

	default:
		p.errorExpected(p.pos(1), "operand")
	}

	return nil
}

func (p *parser) parseRef() Expr {
	id := p.parseIdent()
	if p.tok != LT {
		return id
	}

	p.mark()
	if x := p.tryTypeParameters(); x != nil && !operandStart[p.tok] {
		p.commit()
		return &ParametrizedIdent{Ident: id, Params: x}
	}
	p.reset()
	return id
}

func (p *parser) parseParenExpr() *ParenExpr {
	return &ParenExpr{
		LParen: p.expect(LPAREN),
		List:   p.parseExprList(),
		RParen: p.expect(RPAREN),
	}
}

func (p *parser) parseUniversalCharstring() *Ident {
	id := &Ident{Tok: p.expect(UNIVERSAL)}
	p.expect(CHARSTRING) // TODO(5nord) add this token to AST
	return id
}

func (p *parser) parseCompositeLiteral() *CompositeLiteral {
	c := new(CompositeLiteral)
	c.LBrace = p.expect(LBRACE)
	if p.tok != RBRACE {
		c.List = p.parseExprList()
	}
	c.RBrace = p.expect(RBRACE)
	return c
}

func (p *parser) parseCallRegexp() *RegexpExpr {
	c := new(RegexpExpr)
	c.Tok = p.expect(REGEXP)
	if p.tok == MODIF {
		c.NoCase = p.consume()
	}
	c.X = p.parseParenExpr()
	return c
}

func (p *parser) parseCallPattern() *PatternExpr {
	c := new(PatternExpr)
	c.Tok = p.expect(PATTERN)
	if p.tok == MODIF {
		c.NoCase = p.consume()
	}
	c.X = p.expect(STRING)
	return c
}

func (p *parser) parseCallDecmatch() *DecmatchExpr {
	c := new(DecmatchExpr)
	c.Tok = p.expect(DECMATCH)
	if p.tok == LPAREN {
		c.Params = p.parseParenExpr()
	}
	c.X = p.parseExpr()
	return c
}

func (p *parser) parseDecodedModifier() *DecodedExpr {
	d := new(DecodedExpr)
	d.Tok = p.expect(MODIF) // TODO(5nord) @decoded check
	if p.tok == LPAREN {
		d.Params = p.parseParenExpr()
	}
	d.X = p.parsePrimaryExpr()
	return d
}

func (p *parser) parseSelectorExpr(x Expr) *SelectorExpr {
	return &SelectorExpr{X: x, Dot: p.consume(), Sel: p.parseRef()}
}

func (p *parser) parseIndexExpr(x Expr) *IndexExpr {
	return &IndexExpr{
		X:      x,
		LBrack: p.expect(LBRACK),
		Index:  p.parseExpr(),
		RBrack: p.expect(RBRACK),
	}
}

//TODO(5nord) implement plz
func (p *parser) parseCallExpr(x Expr) *CallExpr {
	p.consume()

	switch p.tok {
	case FROM, TO:
		p.consume()
		p.parseExpr()
		if p.tok == REDIR {
			p.parseRedirect(nil)
		}
		p.expect(RPAREN)
		return nil
	case REDIR:
		p.parseRedirect(nil)
		p.expect(RPAREN)
		return nil
	default:
		var list []Expr
		if p.tok != RPAREN {
			list = p.parseExprList()
		}
		p.expect(RPAREN)
		return &CallExpr{Fun: x, Args: list}
	}
}

func (p *parser) parseLength(x Expr) *LengthExpr {
	return &LengthExpr{
		X:    x,
		Len:  p.expect(LENGTH),
		Size: p.parseParenExpr(),
	}
}

func (p *parser) parseRedirect(x Expr) *RedirectExpr {

	r := &RedirectExpr{
		X:   x,
		Tok: p.expect(REDIR),
	}

	if p.tok == VALUE {
		r.ValueTok = p.expect(VALUE)
		r.Value = p.parseExprList()
	}

	if p.tok == PARAM {
		r.ParamTok = p.expect(PARAM)
		r.Param = p.parseExprList()
	}

	if p.tok == SENDER {
		r.SenderTok = p.expect(SENDER)
		r.Sender = p.parsePrimaryExpr()
	}

	if p.tok == MODIF {
		if p.lit(1) != "@index" {
			p.errorExpected(p.pos(1), "@index")
		}

		tok := p.consume()
		if p.tok == VALUE {
			// just silently discard optional 'value' token
			p.consume()
		}
		r.IndexTok = tok
		r.Index = p.parsePrimaryExpr()
	}

	if p.tok == TIMESTAMP {
		r.TimestampTok = p.expect(TIMESTAMP)
		r.Timestamp = p.parsePrimaryExpr()
	}

	return r
}

func (p *parser) parseIdent() *Ident {
	switch p.tok {
	case UNIVERSAL:
		return p.parseUniversalCharstring()
	case IDENT, ADDRESS, ALIVE, CHARSTRING:
		return &Ident{Tok: p.consume()}
	default:
		p.expect(IDENT) // use expect() error handling
		return nil
	}
}

func (p *parser) parseRefList() []Expr {
	l := make([]Expr, 1)
	for {
		l = append(l, p.parseTypeRef())
		if p.tok != COMMA {
			break
		}
		p.consume()
	}
	return l
}

func (p *parser) parseTypeRef() Expr {
	if p.trace {
		defer un(trace(p, "TypeRef"))
	}
	return p.parsePrimaryExpr()
}

func (p *parser) tryTypeParameters() Expr {
	if p.trace {
		defer un(trace(p, "tryTypeParameters"))
	}
	x := &ParenExpr{
		LParen: p.consume(),
	}
	for p.tok != GT {
		y := p.tryTypeParameter()
		if y == nil {
			return nil
		}
		x.List = append(x.List, y)
		if p.tok != COMMA {
			break
		}
		p.consume() // consume ','
	}

	if p.tok != GT {
		return nil
	}
	x.RParen = p.consume()
	return x
}

func (p *parser) tryTypeParameter() Expr {
	if p.trace {
		defer un(trace(p, "tryTypeParameter"))
	}
	x := p.tryTypeIdent()
L:
	for {
		switch p.tok {
		case DOT:
			x = &SelectorExpr{
				X:   x,
				Dot: p.consume(),
				Sel: p.tryTypeIdent(),
			}
			if x.(*SelectorExpr).Sel == nil {
				return nil
			}
		case LBRACK:
			lbrack := p.consume()
			dash := p.consume()
			rbrack := p.consume()
			if dash.Kind != SUB || rbrack.Kind != RBRACK {
				return nil
			}
			x = &IndexExpr{
				X:      x,
				LBrack: lbrack,
				Index:  dash,
				RBrack: rbrack,
			}

		default:
			break L
		}
	}
	return x
}

func (p *parser) tryTypeIdent() Expr {
	if p.trace {
		defer un(trace(p, "tryTypeIdent"))
	}

	if p.tok != IDENT && p.tok != ADDRESS && p.tok != CHARSTRING && p.tok != UNIVERSAL {
		return nil
	}

	x := &Ident{Tok: p.consume()}

	if p.tok == LT {
		if y := p.tryTypeParameters(); y == nil {
			return &ParametrizedIdent{
				Ident:  x,
				Params: y,
			}
		}
	}
	return x
}

/*************************************************************************
 * Module
 *************************************************************************/

func (p *parser) parseModule() *Module {
	if p.trace {
		defer un(trace(p, "Module"))
	}

	m := new(Module)
	m.Tok = p.expect(MODULE)
	m.Name = p.parseIdent()

	if p.tok == LANGUAGE {
		m.Language = p.parseLanguageSpec()
	}

	m.LBrace = p.expect(LBRACE)

	var decls []Decl
	for p.tok != RBRACE && p.tok != EOF {
		decls = append(decls, p.parseModuleDef())
		p.expectSemi()
	}
	m.RBrace = p.expect(RBRACE)
	m.With = p.parseWith()
	return m
}

func (p *parser) parseLanguageSpec() *LanguageSpec {
	l := new(LanguageSpec)
	l.Tok = p.consume()
	for {
		l.List = append(l.List, p.expect(STRING))
		if p.tok != COMMA {
			break
		}
		p.consume()
	}
	return l
}

func (p *parser) parseModuleDef() *ModuleDef {
	m := new(ModuleDef)
	switch p.tok {
	case PRIVATE, PUBLIC:
		m.Visibility = p.consume()
	case FRIEND:
		if p.peek(2).Kind != MODULE {
			m.Visibility = p.consume()
		}
	}

	switch p.tok {
	case IMPORT:
		m.Def = p.parseImport()
	case GROUP:
		m.Def = p.parseGroup()
	case FRIEND:
		m.Def = p.parseFriend()
	case TYPE:
		m.Def = p.parseTypeDecl()
	case TEMPLATE:
		m.Def = p.parseTemplateDecl()
	case MODULEPAR:
		m.Def = p.parseModulePar()
	case VAR, CONST:
		m.Def = p.parseValueDecl()
	case SIGNATURE:
		m.Def = p.parseSignatureDecl()
	case FUNCTION, TESTCASE, ALTSTEP:
		m.Def = p.parseFuncDecl()
	case CONTROL:
		m.Def = &ControlPart{Tok: p.consume(), Body: p.parseBlockStmt()}
	case EXTERNAL:
		switch p.peek(2).Kind {
		case FUNCTION:
			m.Def = p.parseExtFuncDecl()
		case CONST:
			p.error(p.pos(1), "external constants not suppored")
			p.consume()
			m.Def = p.parseValueDecl()
		default:
			p.errorExpected(p.pos(1), "'function'")
			p.advance(stmtStart)
		}
	default:
		p.errorExpected(p.pos(1), "module definition")
		p.advance(stmtStart)
	}
	return m
}

/*************************************************************************
 * Import Definition
 *************************************************************************/

func (p *parser) parseImport() *ImportDecl {
	if p.trace {
		defer un(trace(p, "Import"))
	}

	x := new(ImportDecl)
	x.ImportTok = p.consume()
	x.FromTok = p.expect(FROM)
	x.Module = p.parseIdent()

	if p.tok == LANGUAGE {
		x.Language = p.parseLanguageSpec()
	}

	switch p.tok {
	case ALL:
		p.consume()
		if p.tok == EXCEPT {
			p.parseExceptSpec()
		}
	case LBRACE:
		p.parseImportSpec()
	default:
		p.errorExpected(p.pos(1), "'all' or import spec")
	}

	x.With = p.parseWith()

	return x
}

func (p *parser) parseImportSpec() {
	p.expect(LBRACE)
	for p.tok != RBRACE && p.tok != EOF {
		p.parseImportStmt()
		p.expectSemi()
	}
	p.expect(RBRACE)
}

func (p *parser) parseImportStmt() {
	switch p.tok {
	case ALTSTEP, CONST, FUNCTION, MODULEPAR,
		SIGNATURE, TEMPLATE, TESTCASE, TYPE:
		p.consume()
		if p.tok == ALL {
			p.consume()
			if p.tok == EXCEPT {
				p.consume()
				p.parseRefList()
			}
		} else {
			p.parseRefList()
		}
	case GROUP:
		p.consume()
		for {
			p.parseTypeRef()
			if p.tok == EXCEPT {
				p.parseExceptSpec()
			}
			if p.tok != COMMA {
				break
			}
			p.consume()
		}
	case IMPORT:
		p.consume()
		p.expect(ALL)
	default:
		p.errorExpected(p.pos(1), "definition qualifier")
	}

}

func (p *parser) parseExceptSpec() {
	p.consume()
	p.expect(LBRACE)
	for p.tok != RBRACE && p.tok != EOF {
		p.parseExceptStmt()
		p.expectSemi()
	}
	p.expect(RBRACE)
}

func (p *parser) parseExceptStmt() {
	switch p.tok {
	case ALTSTEP, CONST, FUNCTION, GROUP,
		IMPORT, MODULEPAR, SIGNATURE, TEMPLATE,
		TESTCASE, TYPE:
		p.consume()
	default:
		p.errorExpected(p.pos(1), "definition qualifier")
	}

	if p.tok == ALL {
		p.consume()
	} else {
		for {
			p.parseTypeRef()
			if p.tok != COMMA {
				break
			}
			p.consume()
		}
	}
}

/*************************************************************************
 * Group Definition
 *************************************************************************/

func (p *parser) parseGroup() *GroupDecl {
	x := new(GroupDecl)
	x.Tok = p.consume()
	x.Name = p.parseIdent()
	x.LBrace = p.expect(LBRACE)

	for p.tok != RBRACE && p.tok != EOF {
		x.Defs = append(x.Defs, p.parseModuleDef())
		p.expectSemi()
	}
	x.RBrace = p.expect(RBRACE)
	x.With = p.parseWith()
	return x
}

func (p *parser) parseFriend() *FriendDecl {
	return &FriendDecl{
		FriendTok: p.expect(FRIEND),
		ModuleTok: p.expect(MODULE),
		Module:    p.parseIdent(),
		With:      p.parseWith(),
	}
}

/*************************************************************************
 * With Attributes
 *************************************************************************/

func (p *parser) parseWith() *WithSpec {
	if p.tok != WITH {
		return nil
	}
	x := new(WithSpec)
	x.Tok = p.consume()
	x.LBrace = p.expect(LBRACE)
	for p.tok != RBRACE && p.tok != EOF {
		x.List = append(x.List, p.parseWithStmt())
		p.expectSemi()
	}
	x.RBrace = p.expect(RBRACE)
	return x
}

func (p *parser) parseWithStmt() *WithStmt {
	if p.trace {
		defer un(trace(p, "WithStmt"))
	}

	x := new(WithStmt)

	switch p.tok {
	case ENCODE,
		VARIANT,
		DISPLAY,
		EXTENSION,
		OPTIONAL,
		STEPSIZE,
		OVERRIDE:
		x.Kind = p.consume()
	default:
		p.errorExpected(p.pos(1), "with-attribute")
		p.advance(stmtStart)
	}

	switch p.tok {
	case OVERRIDE:
		x.Override = p.consume()
	case MODIF:
		if p.lit(1) != "@local" {
			p.errorExpected(p.pos(1), "@local")
		}
		x.Override = p.consume()
	}

	if p.tok == LPAREN {
		x.LParen = p.consume()
		for {
			x.List = append(x.List, p.parseWithQualifier())
			if p.tok != COMMA {
				break
			}
			p.consume()
		}
		x.RParen = p.expect(RPAREN)
	}

	var v Expr = &ValueLiteral{Tok: p.expect(STRING)}
	if p.tok == DOT {
		v = &SelectorExpr{
			X:   v,
			Dot: p.consume(),
			Sel: &ValueLiteral{Tok: p.expect(STRING)},
		}
	}
	x.Value = v

	return x
}

func (p *parser) parseWithQualifier() Node {
	switch p.tok {
	case IDENT:
		p.parseTypeRef()
	case LBRACK:
		p.parseIndexExpr(nil)
	case TYPE, TEMPLATE, CONST, ALTSTEP, TESTCASE, FUNCTION, SIGNATURE, MODULEPAR, GROUP:
		p.consume()
		p.expect(ALL)
		if p.tok == EXCEPT {
			p.consume()
			p.expect(LBRACE)
			p.parseRefList()
			p.expect(RBRACE)
		}
	default:
		p.errorExpected(p.pos(1), "with-qualifier")
	}
	return nil
}

/*************************************************************************
 * Type Definitions
 *************************************************************************/

func (p *parser) parseTypeDecl() Decl {
	switch p.peek(2).Kind {
	case IDENT, ADDRESS, CHARSTRING, NULL, UNIVERSAL:
		return p.parseSubTypeDecl()
	case PORT:
		return p.parsePortTypeDecl()
	case COMPONENT:
		return p.parseComponentTypeDecl()
	case UNION:
		return p.parseStructTypeDecl()
	case SET, RECORD:
		if p.peek(3).Kind == IDENT {
			return p.parseStructTypeDecl()
		}
		// lists are also parsed by parseSubTypeDecl
		return p.parseSubTypeDecl()
	case ENUMERATED:
		return p.parseEnumTypeDecl()
	case FUNCTION, ALTSTEP, TESTCASE:
		return p.parseBehaviourTypeDecl()
	default:
		p.errorExpected(p.pos(1), "type definition")
		p.advance(stmtStart)
		return nil
	}
}

/*************************************************************************
 * Port Type
 *************************************************************************/

func (p *parser) parsePortTypeDecl() *PortTypeDecl {
	if p.trace {
		defer un(trace(p, "PortTypeDecl"))
	}
	x := new(PortTypeDecl)
	x.TypeTok = p.consume()
	x.PortTok = p.consume()
	x.Name = p.parseIdent()
	if p.tok == LT {
		p.parseTypeFormalPars()
	}

	switch p.tok {
	case MIXED, MESSAGE, PROCEDURE:
		x.Kind = p.consume()
	default:
		p.errorExpected(p.pos(1), "'message' or 'procedure'")
	}

	if p.tok == REALTIME {
		x.Realtime = p.consume()
	}

	x.LBrace = p.expect(LBRACE)
	for p.tok != RBRACE && p.tok != EOF {
		x.Attrs = append(x.Attrs, p.parsePortAttribute())
		p.expectSemi()
	}
	x.RBrace = p.expect(RBRACE)
	x.With = p.parseWith()
	return x
}

func (p *parser) parsePortAttribute() Node {
	if p.trace {
		defer un(trace(p, "PortAttribute"))
	}
	switch p.tok {
	case IN, OUT, INOUT, ADDRESS:
		return &PortAttribute{
			Kind:  p.consume(),
			Types: p.parseRefList(),
		}
	case MAP, UNMAP:
		return &PortMapAttribute{
			MapTok:   p.consume(),
			ParamTok: p.expect(PARAM),
			Params:   p.parseFormalPars(),
		}
	default:
		p.errorExpected(p.pos(1), "port attribute")
		p.advance(stmtStart)
		return nil
	}
}

/*************************************************************************
 * Component Type
 *************************************************************************/

func (p *parser) parseComponentTypeDecl() *ComponentTypeDecl {
	if p.trace {
		defer un(trace(p, "ComponentTypeDecl"))
	}
	x := new(ComponentTypeDecl)
	x.TypeTok = p.consume()
	x.CompTok = p.consume()
	x.Name = p.parseIdent()
	if p.tok == LT {
		p.parseTypeFormalPars()
	}
	if p.tok == EXTENDS {
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

func (p *parser) parseStructTypeDecl() *StructTypeDecl {
	if p.trace {
		defer un(trace(p, "StructTypeDecl"))
	}
	x := new(StructTypeDecl)
	x.TypeTok = p.consume()
	x.Kind = p.consume()
	x.Name = p.parseIdent()
	if p.tok == LT {
		p.parseTypeFormalPars()
	}
	x.LBrace = p.expect(LBRACE)
	for p.tok != RBRACE && p.tok != EOF {
		x.Fields = append(x.Fields, p.parseField())
		if p.tok != COMMA {
			break
		}
		p.consume()
	}
	x.RBrace = p.expect(RBRACE)
	x.With = p.parseWith()
	return x
}

/*************************************************************************
 * Enumeration Type Declaration
 *************************************************************************/

func (p *parser) parseEnumTypeDecl() *EnumTypeDecl {
	if p.trace {
		defer un(trace(p, "EnumTypeDecl"))
	}

	x := new(EnumTypeDecl)
	x.TypeTok = p.consume()
	x.EnumTok = p.consume()
	x.Name = p.parseIdent()
	if p.tok == LT {
		p.parseTypeFormalPars()
	}
	x.LBrace = p.expect(LBRACE)
	for p.tok != RBRACE && p.tok != EOF {
		x.Enums = append(x.Enums, p.parseExpr())
		if p.tok != COMMA {
			break
		}
		p.consume()
	}
	x.RBrace = p.expect(RBRACE)
	x.With = p.parseWith()
	return x
}

/*************************************************************************
 * Behaviour Type Declaration
 *************************************************************************/

func (p *parser) parseBehaviourTypeDecl() *BehaviourTypeDecl {
	if p.trace {
		defer un(trace(p, "BehaviourTypeDecl"))
	}
	x := new(BehaviourTypeDecl)
	x.TypeTok = p.consume()
	x.Kind = p.consume()
	x.Name = p.parseIdent()
	if p.tok == LT {
		p.parseTypeFormalPars()
	}

	x.Params = p.parseFormalPars()
	if p.tok == RUNS {
		x.RunsOn = p.parseRunsOn()
	}

	if p.tok == SYSTEM {
		x.System = p.parseSystem()
	}

	if p.tok == RETURN {
		x.Return = p.parseReturn()
	}
	x.With = p.parseWith()
	return x
}

/*************************************************************************
 * Subtype
 *************************************************************************/

func (p *parser) parseSubTypeDecl() *SubTypeDecl {
	if p.trace {
		defer un(trace(p, "SubTypeDecl"))
	}
	x := new(SubTypeDecl)
	x.TypeTok = p.consume()
	x.Field = p.parseField()
	x.With = p.parseWith()
	return x
}

func (p *parser) parseField() *Field {
	if p.trace {
		defer un(trace(p, "Field"))
	}
	x := new(Field)

	if p.tok == MODIF {
		if p.lit(1) != "@default" {
			p.errorExpected(p.pos(1), "@default")
		}
		x.DefaultTok = p.consume()
	}
	x.Type = p.parseTypeSpec()
	x.Name = p.parsePrimaryExpr()
	if p.tok == LT {
		p.parseTypeFormalPars()
	}

	// TODO(mef) fix constraints consumed by previous PrimaryExpr
	if p.tok == LPAREN {
		x.ValueConstraint = p.parseParenExpr()
	}
	if p.tok == LENGTH {
		x.LengthConstraint = p.parseLength(nil)
	}

	if p.tok == OPTIONAL {
		x.Optional = p.consume()
	}
	return x
}

func (p *parser) parseTypeSpec() TypeSpec {
	if p.trace {
		defer un(trace(p, "TypeSpec"))
	}
	switch p.tok {
	case ADDRESS, CHARSTRING, IDENT, NULL, UNIVERSAL:
		return &RefSpec{X: p.parseTypeRef()}
	case UNION:
		return p.parseStructSpec()
	case SET, RECORD:
		if p.peek(2).Kind == LBRACE {
			return p.parseStructSpec()
		}
		return p.parseListSpec()
	case ENUMERATED:
		return p.parseEnumSpec()
	case FUNCTION, ALTSTEP, TESTCASE:
		return p.parseBehaviourSpec()
	default:
		p.errorExpected(p.pos(1), "type definition")
		return nil
	}
}

func (p *parser) parseStructSpec() *StructSpec {
	if p.trace {
		defer un(trace(p, "StructSpec"))
	}
	x := new(StructSpec)
	x.Kind = p.consume()
	x.LBrace = p.expect(LBRACE)
	for p.tok != RBRACE && p.tok != EOF {
		x.Fields = append(x.Fields, p.parseField())
		if p.tok != COMMA {
			break
		}
		p.consume()
	}
	x.RBrace = p.expect(RBRACE)
	return x
}

func (p *parser) parseEnumSpec() *EnumSpec {
	if p.trace {
		defer un(trace(p, "ListSpec"))
	}
	x := new(EnumSpec)
	x.Tok = p.consume()
	x.LBrace = p.expect(LBRACE)
	for p.tok != RBRACE && p.tok != EOF {
		x.Enums = append(x.Enums, p.parseExpr())
		if p.tok != COMMA {
			break
		}
		p.consume()
	}
	x.RBrace = p.expect(RBRACE)
	return x
}

func (p *parser) parseListSpec() *ListSpec {
	if p.trace {
		defer un(trace(p, "ListSpec"))
	}
	x := new(ListSpec)
	x.Kind = p.consume()
	if p.tok == LENGTH {
		x.Length = p.parseLength(nil)
	}
	x.OfTok = p.expect(OF)
	x.ElemType = p.parseTypeSpec()
	return x
}

func (p *parser) parseBehaviourSpec() *BehaviourSpec {
	if p.trace {
		defer un(trace(p, "BehaviourSpec"))
	}

	x := new(BehaviourSpec)
	x.Kind = p.consume()
	x.Params = p.parseFormalPars()

	if p.tok == RUNS {
		x.RunsOn = p.parseRunsOn()
	}

	if p.tok == SYSTEM {
		x.System = p.parseSystem()
	}

	if p.tok == RETURN {
		x.Return = p.parseReturn()
	}
	return x
}

/*************************************************************************
 * Template Declaration
 *************************************************************************/

func (p *parser) parseTemplateDecl() *ValueDecl {
	if p.trace {
		defer un(trace(p, "TemplateDecl"))
	}

	x := &ValueDecl{Kind: p.consume()}

	if p.tok == LPAREN {
		p.consume() // consume '('
		p.consume() // consume omit/value/...
		p.expect(RPAREN)
	}

	if p.tok == MODIF {
		p.consume()
	}

	x.Type = p.parseTypeRef()
	p.parseIdent()
	if p.tok == LT {
		p.parseTypeFormalPars()
	}
	if p.tok == LPAREN {
		p.parseFormalPars()
	}
	if p.tok == MODIFIES {
		p.consume()
		p.parsePrimaryExpr()
	}
	p.expect(ASSIGN)
	p.parseExpr()

	p.parseWith()
	return x
}

/*************************************************************************
 * Module FormalPar
 *************************************************************************/

func (p *parser) parseModulePar() *ValueDecl {
	if p.trace {
		defer un(trace(p, "ModulePar"))
	}

	x := &ValueDecl{Kind: p.consume()}

	if p.tok == LBRACE {
		p.consume()
		for p.tok != RBRACE && p.tok != EOF {
			p.parseRestrictionSpec()
			p.parseTypeRef()
			p.parseExprList()
			p.expectSemi()
		}
		p.expect(RBRACE)
	} else {
		p.parseRestrictionSpec()
		p.parseTypeRef()
		p.parseExprList()
	}

	p.parseWith()
	return x
}

/*************************************************************************
 * Value Declaration
 *************************************************************************/

func (p *parser) parseValueDecl() *ValueDecl {
	if p.trace {
		defer un(trace(p, "ValueDecl"))
	}
	x := &ValueDecl{Kind: p.consume()}
	p.parseRestrictionSpec()

	if p.tok == MODIF {
		p.consume()
	}

	if x.Kind.Kind != TIMER {
		x.Type = p.parseTypeRef()
	}
	x.Decls = p.parseExprList()
	p.parseWith()
	return x
}

func (p *parser) parseRestrictionSpec() *RestrictionSpec {
	switch p.tok {
	case TEMPLATE:
		x := &RestrictionSpec{Tok: p.consume()}
		if p.tok != LPAREN {
			return x
		}

		p.consume()
		x.Tok = p.consume()
		p.expect(RPAREN)

	case OMIT, VALUE, PRESENT:
		x := &RestrictionSpec{Tok: p.consume()}
		return x
	}
	return nil
}

/*************************************************************************
 * Behaviour Declaration
 *************************************************************************/

func (p *parser) parseFuncDecl() *FuncDecl {
	if p.trace {
		defer un(trace(p, "FuncDecl"))
	}

	x := &FuncDecl{Kind: p.consume()}
	x.Name = p.parseIdent()
	if p.tok == LT {
		p.parseTypeFormalPars()
	}

	if p.tok == MODIF {
		p.consume()
	}

	x.Params = p.parseFormalPars()

	if p.tok == RUNS {
		p.parseRunsOn()
	}

	if p.tok == MTC {
		p.parseMtc()
	}

	if p.tok == SYSTEM {
		p.parseSystem()
	}

	if p.tok == RETURN {
		x.Return = p.parseReturn()
	}

	if p.tok == LBRACE {
		x.Body = p.parseBlockStmt()
	}

	p.parseWith()
	return x
}

/*************************************************************************
 * External Function Declaration
 *************************************************************************/

func (p *parser) parseExtFuncDecl() *FuncDecl {
	if p.trace {
		defer un(trace(p, "ExtFuncDecl"))
	}
	p.consume() // consume 'external'
	x := &FuncDecl{Kind: p.consume()}
	x.Name = p.parseIdent()

	if p.tok == MODIF {
		p.consume()
	}

	x.Params = p.parseFormalPars()

	if p.tok == RUNS {
		p.parseRunsOn()
	}

	if p.tok == MTC {
		p.parseMtc()
	}

	if p.tok == SYSTEM {
		p.parseSystem()
	}

	if p.tok == RETURN {
		x.Return = p.parseReturn()
	}
	p.parseWith()
	return x
}

/*************************************************************************
 * Signature Declaration
 *************************************************************************/

func (p *parser) parseSignatureDecl() Decl {
	if p.trace {
		defer un(trace(p, "SignatureDecl"))
	}

	p.consume()
	p.parseIdent()
	if p.tok == LT {
		p.parseTypeFormalPars()
	}

	p.parseFormalPars()

	if p.tok == NOBLOCK {
		p.consume()
	}

	if p.tok == RETURN {
		p.parseReturn()
	}

	if p.tok == EXCEPTION {
		p.consume()
		p.parseParenExpr()
	}
	p.parseWith()
	return nil
}

func (p *parser) parseRunsOn() *RunsOnSpec {
	return &RunsOnSpec{
		RunsTok: p.expect(RUNS),
		OnTok:   p.expect(ON),
		Comp:    p.parseTypeRef(),
	}
}

func (p *parser) parseSystem() *SystemSpec {
	return &SystemSpec{
		Tok:  p.expect(SYSTEM),
		Comp: p.parseTypeRef(),
	}
}

func (p *parser) parseMtc() *MtcSpec {
	return &MtcSpec{
		Tok:  p.expect(MTC),
		Comp: p.parseTypeRef(),
	}
}

func (p *parser) parseReturn() *ReturnSpec {
	x := new(ReturnSpec)
	x.Tok = p.consume()
	x.Restriction = p.parseRestrictionSpec()
	if p.tok == MODIF {
		x.Modif = p.consume()
	}
	x.Type = p.parseTypeRef()
	return x
}

func (p *parser) parseFormalPars() *FormalPars {
	if p.trace {
		defer un(trace(p, "FormalPars"))
	}
	x := &FormalPars{}
	p.expect(LPAREN)
	for p.tok != RPAREN {
		x.List = append(x.List, p.parseFormalPar())
		if p.tok != COMMA {
			break
		}
		p.consume()
	}
	p.expect(RPAREN)
	return x
}

func (p *parser) parseFormalPar() *FormalPar {
	if p.trace {
		defer un(trace(p, "FormalPar"))
	}
	x := &FormalPar{}

	switch p.tok {
	case IN:
		p.consume()
	case OUT:
		p.consume()
	case INOUT:
		p.consume()
	}

	p.parseRestrictionSpec()
	if p.tok == MODIF {
		p.consume()
	}
	x.Type = p.parseTypeRef()
	x.Name = p.parseExpr()

	return x
}

func (p *parser) parseTypeFormalPars() {
	if p.trace {
		defer un(trace(p, "TypeFormalPars"))
	}
	p.expect(LT)
	for p.tok != GT {
		p.parseTypeFormalPar()
		if p.tok != COMMA {
			break
		}
		p.consume()
	}
	p.expect(GT)
}

func (p *parser) parseTypeFormalPar() {
	if p.trace {
		defer un(trace(p, "TypeFormalPar"))
	}
	if p.tok == IN {
		p.consume()
	}

	switch p.tok {
	case TYPE:
		p.consume()
	case SIGNATURE:
		p.consume()
	default:
		p.parseTypeRef()
	}
	p.expect(IDENT)
	if p.tok == ASSIGN {
		p.consume()
		p.parseTypeRef()
	}
}

/*************************************************************************
 * Statements
 *************************************************************************/

func (p *parser) parseBlockStmt() *BlockStmt {
	if p.trace {
		defer un(trace(p, "BlockStmt"))
	}

	x := &BlockStmt{LBrace: p.expect(LBRACE)}
	for p.tok != RBRACE && p.tok != EOF {
		x.Stmts = append(x.Stmts, p.parseStmt())
		p.expectSemi()
	}
	x.RBrace = p.expect(RBRACE)
	return x
}

func (p *parser) parseStmt() Stmt {
	if p.trace {
		defer un(trace(p, "Stmt"))
	}

	switch p.tok {
	case TEMPLATE:
		return p.parseTemplateDecl()
	case VAR, CONST, TIMER, PORT:
		return p.parseValueDecl()
	case REPEAT, BREAK, CONTINUE:
		return &BranchStmt{Tok: p.consume()}
	case LABEL:
		return &BranchStmt{Tok: p.consume(), Label: p.expect(IDENT)}
	case GOTO:
		return &BranchStmt{Tok: p.consume(), Label: p.expect(IDENT)}
	case RETURN:
		x := &ReturnStmt{Tok: p.consume()}
		if !stmtStart[p.tok] && p.tok != SEMICOLON && p.tok != RBRACE {
			x.Result = p.parseExpr()
		}
		return x
	case SELECT:
		return p.parseSelect()
	case ALT, INTERLEAVE:
		return &AltStmt{Tok: p.consume(), Block: p.parseBlockStmt()}
	case LBRACK:
		return p.parseAltGuard()
	case FOR:
		return p.parseForLoop()
	case WHILE:
		return p.parseWhileLoop()
	case DO:
		return p.parseDoWhileLoop()
	case IF:
		return p.parseIfStmt()
	case LBRACE:
		return p.parseBlockStmt()
	case IDENT, TESTCASE, ANYKW, ALL, MAP, UNMAP, MTC:
		x := &ExprStmt{Expr: p.parseSimpleStmt()}

		// call-statement block
		if p.tok == LBRACE {
			// TODO(5nord) check if x is call-expression
			p.parseBlockStmt()
		}
		return x
	default:
		p.errorExpected(p.pos(1), "statement")
		p.advance(stmtStart)
		return nil
	}
}

func (p *parser) parseForLoop() *ForStmt {
	x := new(ForStmt)
	x.Tok = p.consume()
	x.LParen = p.expect(LPAREN)
	if p.tok == VAR {
		x.Init = &DeclStmt{Decl: p.parseValueDecl()}
	} else {
		x.Init = &ExprStmt{Expr: p.parseExpr()}
	}
	x.InitSemi = p.expect(SEMICOLON)
	x.Cond = p.parseExpr()
	x.CondSemi = p.expect(SEMICOLON)
	x.Post = p.parseExpr()
	x.LParen = p.expect(RPAREN)
	x.Body = p.parseBlockStmt()
	return x
}

func (p *parser) parseWhileLoop() *WhileStmt {
	return &WhileStmt{
		Tok:  p.consume(),
		Cond: p.parseParenExpr(),
		Body: p.parseBlockStmt(),
	}
}

func (p *parser) parseDoWhileLoop() *DoWhileStmt {
	return &DoWhileStmt{
		DoTok:    p.consume(),
		Body:     p.parseBlockStmt(),
		WhileTok: p.expect(WHILE),
		Cond:     p.parseParenExpr(),
	}
}

func (p *parser) parseIfStmt() *IfStmt {
	x := &IfStmt{
		Tok:  p.consume(),
		Cond: p.parseParenExpr(),
		Then: p.parseBlockStmt(),
	}
	if p.tok == ELSE {
		x.ElseTok = p.consume()
		if p.tok == IF {
			x.Else = p.parseIfStmt()
		} else {
			x.Else = p.parseBlockStmt()
		}
	}
	return x
}

func (p *parser) parseSelect() *SelectStmt {
	x := new(SelectStmt)
	x.Tok = p.expect(SELECT)
	if p.tok == UNION {
		x.Union = p.consume()
	}
	x.Tag = p.parseParenExpr()
	x.LBrace = p.expect(LBRACE)
	for p.tok == CASE {
		x.Body = append(x.Body, p.parseCaseStmt())
	}
	x.RBrace = p.expect(RBRACE)
	return x
}

func (p *parser) parseCaseStmt() *CaseClause {
	x := new(CaseClause)
	x.Tok = p.expect(CASE)
	if p.tok == ELSE {
		p.consume() //TODO(5nord) move token into AST
	} else {
		x.Case = p.parseParenExpr()
	}
	x.Body = p.parseBlockStmt()
	return x
}

func (p *parser) parseAltGuard() *CommClause {
	x := new(CommClause)
	x.LBrack = p.expect(LBRACK)
	if p.tok == ELSE {
		x.Else = p.consume()
		x.RBrack = p.expect(RBRACK)
		x.Body = p.parseBlockStmt()
		return x
	}

	if p.tok != RBRACK {
		x.X = p.parseExpr()
	}
	x.RBrack = p.expect(RBRACK)
	x.Comm = p.parseSimpleStmt()
	if p.tok == LBRACE {
		x.Body = p.parseBlockStmt()
	}
	return x
}

func (p *parser) parseSimpleStmt() Stmt {
	if p.trace {
		defer un(trace(p, "SimpleStmt"))
	}

	return &ExprStmt{Expr: p.parseExpr()}
}
