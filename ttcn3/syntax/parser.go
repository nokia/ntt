package syntax

import (
	"fmt"
)

// A Mode value is a set of flags (or 0).
// They control the amount of source code parsed and other optional
// parser functionality.
//
type Mode uint

const (
	ImportsOnly       = 1 << iota      // stop parsing after import declarations
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
	indent int  // indentation used for tracing output

	// Tokens/Backtracking
	cursor  int
	tokens  []Token
	markers []int

	// Semicolon helper
	seenBrace bool

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

	p.tokens = make([]Token, 0, 200)
	p.markers = make([]int, 0, 200)
}

// Usage pattern: defer un(trace(p, "..."))
func un(p *parser) {
	p.indent--
	p.printTrace(")")
}

// Read the next token from input-stream
func (p *parser) readKind() Token {
redo:
	pos, tok, lit := p.scanner.Scan()
	if tok == COMMENT || tok == PREPROC {
		goto redo
	}
	return Token{pos, tok, lit}
}

// Advance to the next token
func (p *parser) next() {

	if p.trace {
		tok := p.tokens[p.cursor].Kind
		lit := p.tokens[p.cursor].Lit
		s := tok.String()
		switch {
		case tok.IsLiteral():
			p.printTrace(s, lit)
		case tok.IsOperator(), tok.IsKeyword():
			p.printTrace("\"" + s + "\"")
		default:
			p.printTrace(s)
		}
	}

	// Track curly braces for TTCN-3 semicolon rules
	p.seenBrace = false
	if p.tok(1) == RBRACE {
		p.seenBrace = true
	}

	p.cursor++
	if p.cursor == len(p.tokens) && !p.speculating() {
		p.cursor = 0
		p.tokens = p.tokens[:0]
	}

	p.grow(1)
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

func (p *parser) tok(i int) Kind {
	return p.peek(i).Kind
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

func (p *parser) release() {
	last := len(p.markers) - 1
	marker := p.markers[last]
	p.markers = p.markers[0:last]
	p.cursor = marker
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
		case p.tok(1).IsLiteral():
			// print 123 rather than 'INT', etc.
			msg += ", found " + p.lit(1)
		default:
			msg += ", found '" + p.tok(1).String() + "'"
		}
	}
	p.error(pos, msg)
}

func (p *parser) expect(tok Kind) Pos {
	pos := p.pos(1)
	if p.tok(1) != tok {
		p.errorExpected(pos, "'"+tok.String()+"'")
	}
	p.next() // make progress
	return pos
}

func (p *parser) expectSemi() {
	switch p.tok(1) {
	case SEMICOLON:
		p.next()
	case RBRACE, EOF:
		// semicolon is optional before a closing '}'
	default:
		if !p.seenBrace {
			p.errorExpected(p.pos(1), "';'")
			p.advance(stmtStart)
		}
	}
}

// advance consumes tokens until the current token p.tok(1)
// is in the 'to' set, or EOF. For error recovery.
func (p *parser) advance(to map[Kind]bool) {
	for ; p.tok(1) != EOF; p.next() {
		if to[p.tok(1)] {
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
	CONST:     true,
	VAR:       true,
	MODULEPAR: true,
	FUNCTION:  true,
	TESTCASE:  true,
	ALTSTEP:   true,
}

/*************************************************************************
 * Expressions
 *************************************************************************/

// ExprList ::= Expr { "," Expr }
func (p *parser) parseExprList() (list []Expr) {
	list = append(list, p.parseExpr())
	for p.tok(1) == COMMA {
		p.next()
		list = append(list, p.parseExpr())
	}
	return list
}

// Expr ::= BinaryExpr [ ":=" Expr ]
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
		prec := p.tok(1).Precedence()
		if prec < prec1 {
			return x
		}
		pos := p.pos(1)
		op := p.tok(1)
		p.next()

		y := p.parseBinaryExpr(prec + 1)

		x = &BinaryExpr{X: x, Op: op, OpPos: pos, Y: y}
	}
}

// UnaryExpr ::= "-"
//             | ("-"|"+"|"!"|"not"|"not4b") UnaryExpr
//             | "modifies" PrimaryExpr ":=" Expr
//             | PrimaryExpr
//
func (p *parser) parseUnaryExpr() Expr {
	switch p.tok(1) {
	case ADD, EXCL, NOT, NOT4B, SUB:
		op, pos := p.tok(1), p.pos(1)
		p.next()
		// handle unused expr '-'
		if op == SUB {
			switch p.tok(1) {
			case COMMA, SEMICOLON, RBRACE, RBRACK, RPAREN, EOF:
				return &ValueLiteral{Kind: op, ValuePos: pos, Value: "-"}
			}
		}
		return &UnaryExpr{Op: op, OpPos: pos, X: p.parseUnaryExpr()}
	}

	return p.parsePrimaryExpr()
}

// PrimaryExpr ::= Operand [{ExtFieldRef}]
//                         ["length" "(" ExprList ")"] ["ifpresent"]
//                         [("to"|"from") Expr]        ["->" Redirect]
//
// ExtFieldRef ::= "." ID
//               | "[" Expr "]"
//               | "(" ExprList ")"
//               | ":" Expr
//
// Redirect    ::= ["value"            ExprList]
//                 ["param"            ExprList"]
//                 ["sender"           PrimaryExpr]
//                 ["@index" ["value"] PrimaryExpr]
//                 ["timestamp"        PrimaryExpr]
//
func (p *parser) parsePrimaryExpr() Expr {
	x := p.parseOperand()
L:
	for {
		switch p.tok(1) {
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

	if p.tok(1) == LENGTH {
		x = p.parseLength(x)
	}

	if p.tok(1) == IFPRESENT {
		op, pos := p.tok(1), p.pos(1)
		p.next()
		x = &UnaryExpr{Op: op, OpPos: pos, X: x}
	}

	if p.tok(1) == TO || p.tok(1) == FROM {
		op, pos := p.tok(1), p.pos(1)
		p.next()
		x = &BinaryExpr{X: x, Op: op, OpPos: pos, Y: p.parseExpr()}
	}

	if p.tok(1) == REDIR {
		x = p.parseRedirect(x)
	}

	if p.tok(1) == VALUE {
		//TODO(5nord) maybe merge with redirect
		p.next()
		p.parseExpr()
	}

	if p.tok(1) == PARAM {
		//TODO(5nord) maybe merge with redirect
		p.next()
		p.parseParenExpr()
	}

	if p.tok(1) == ALIVE {
		op, pos := p.tok(1), p.pos(1)
		p.next()
		x = &UnaryExpr{Op: op, OpPos: pos, X: x}
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
	switch p.tok(1) {
	case ANYKW, ALL:
		k := p.tok(1)
		p.next()
		switch p.tok(1) {
		case COMPONENT, PORT, TIMER:
			p.next()
			return nil
		case FROM:
			p.next()
			p.parsePrimaryExpr()
			return nil
		}

		// Workaround for deprecated port-attribute 'all'
		if k == ALL {
			return nil
		}

		p.errorExpected(p.pos(1), "'component', 'port', 'timer' or 'from'")

	case UNIVERSAL:
		p.parseUniversalCharstring()
		id := &Ident{NamePos: p.pos(1), Name: p.lit(1)}
		return id

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
		id := &Ident{NamePos: p.pos(1), Name: p.lit(1)}
		p.next()
		return id

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
		lit := &ValueLiteral{Kind: p.tok(1), ValuePos: p.pos(1), Value: p.lit(1)}
		p.next()
		return lit

	case LPAREN:
		// can be template `x := (1,2,3)`, but also artihmetic expression: `1*(2+3)`
		p.parseParenExpr()

	case LBRACK:
		p.parseIndexExpr(nil)

	case LBRACE:
		p.parseCompositeLiteral()

	case MODIFIES:
		p.next()
		p.parsePrimaryExpr()
		p.expect(ASSIGN)
		p.parseExpr()
		return nil

	case REGEXP:
		p.parseCallRegexp()

	case PATTERN:
		p.parseCallPattern()

	case DECMATCH:
		p.parseCallDecMatch()

	case MODIF:
		p.parseDecodedModifier()

	default:
		p.errorExpected(p.pos(1), "operand")
	}

	return nil
}

func isOperand(tok Kind) bool {
	switch tok {
	case ADDRESS,
		ALL,
		ANY,
		ANYKW,
		BSTRING,
		CHARSTRING,
		ERROR,
		FAIL,
		FALSE,
		FLOAT,
		IDENT,
		INCONC,
		INT,
		MAP,
		MTC,
		MUL,
		NAN,
		NONE,
		NULL,
		OMIT,
		PASS,
		STRING,
		SYSTEM,
		TESTCASE,
		TIMER,
		TRUE,
		UNIVERSAL,
		UNMAP:
		return true
	default:
		return false
	}
}

func (p *parser) parseRef() Expr {
	id := p.parseIdent()
	if p.tok(1) != LT {
		return id
	}

	p.mark()
	if x := p.tryTypeFormalPars(); x != nil && !isOperand(p.tok(1)) {
		p.commit()
		return id
	}
	p.release()
	return id
}

func (p *parser) parseParenExpr() Expr {
	p.expect(LPAREN)
	p.parseExprList()
	p.expect(RPAREN)
	return nil
}

func (p *parser) parseUniversalCharstring() {
	p.expect(UNIVERSAL)
	p.expect(CHARSTRING)
}

func (p *parser) parseCompositeLiteral() {
	p.expect(LBRACE)
	if p.tok(1) != RBRACE {
		p.parseExprList()
	}
	p.expect(RBRACE)
}

func (p *parser) parseCallRegexp() {
	p.expect(REGEXP)
	if p.tok(1) == MODIF {
		p.next()
	}
	p.parseParenExpr()
}

func (p *parser) parseCallPattern() {
	p.expect(PATTERN)
	if p.tok(1) == MODIF {
		p.next()
	}
	p.expect(STRING)
}

func (p *parser) parseCallDecMatch() {
	p.expect(DECMATCH)
	if p.tok(1) == LPAREN {
		p.parseParenExpr()
	}
	p.parseExpr()
}

func (p *parser) parseDecodedModifier() {
	p.expect(MODIF) // @decoded
	if p.tok(1) == LPAREN {
		p.parseParenExpr()
	}
	p.parsePrimaryExpr()
}

func (p *parser) parseSelectorExpr(x Expr) Expr {
	p.expect(DOT)
	return &SelectorExpr{X: x, Sel: p.parseRef()}
}

func (p *parser) parseIndexExpr(x Expr) Expr {
	p.expect(LBRACK)
	x = &IndexExpr{X: x, Index: p.parseExpr()}
	p.expect(RBRACK)
	return x
}

func (p *parser) parseCallExpr(x Expr) Expr {
	p.next()

	switch p.tok(1) {
	case FROM, TO:
		p.next()
		p.parseExpr()
		if p.tok(1) == REDIR {
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
		if p.tok(1) != RPAREN {
			list = p.parseExprList()
		}
		p.expect(RPAREN)
		return &CallExpr{Fun: x, Args: list}
	}
}

func (p *parser) parseLength(x Expr) *LengthExpr {
	return &LengthExpr{
		Pos: p.expect(LENGTH),
		X:   x,
		Len: p.parseParenExpr(),
	}
}

func (p *parser) parseRedirect(x Expr) *RedirectExpr {

	r := &RedirectExpr{
		X:   x,
		Pos: p.expect(REDIR),
	}

	if p.tok(1) == VALUE {
		r.ValuePos = p.expect(VALUE)
		r.Value = p.parseExprList()
	}

	if p.tok(1) == PARAM {
		r.ParamPos = p.expect(PARAM)
		r.Param = p.parseExprList()
	}

	if p.tok(1) == SENDER {
		r.SenderPos = p.expect(SENDER)
		r.Sender = p.parsePrimaryExpr()
	}

	if p.tok(1) == MODIF {
		if p.lit(1) != "@index" {
			p.errorExpected(p.pos(1), "@index")
		}

		pos := p.pos(1)
		p.next()
		if p.tok(1) == VALUE {
			// just silently discard optional 'value' token
			p.next()
		}
		r.IndexPos = pos
		r.Index = p.parsePrimaryExpr()
	}

	if p.tok(1) == TIMESTAMP {
		r.TimestampPos = p.expect(TIMESTAMP)
		r.Timestamp = p.parsePrimaryExpr()
	}

	return r
}

func (p *parser) parseIdent() *Ident {
	pos := p.pos(1)
	name := "_"
	switch p.tok(1) {
	case UNIVERSAL:
		p.parseUniversalCharstring()
	case IDENT, ADDRESS, ALIVE, CHARSTRING:
		name = p.lit(1)
		p.next()
	default:
		p.expect(IDENT) // use expect() error handling
	}
	return &Ident{NamePos: pos, Name: name}
}

func (p *parser) parseRefList() {
	for {
		p.parseTypeRef()
		if p.tok(1) != COMMA {
			break
		}
		p.next()
	}
}

func (p *parser) parseTypeRef() Expr {
	if p.trace {
		defer un(trace(p, "TypeRef"))
	}
	x := p.parsePrimaryExpr()
	return x
}

func (p *parser) tryTypeFormalPars() Expr {
	if p.trace {
		defer un(trace(p, "tryTypeFormalPars"))
	}
	x := &Ident{Name: "dummy"}
	p.next() // consume '<'
	for p.tok(1) != GT {
		y := p.tryTypeFormalPar()
		if y == nil {
			return nil
		}
		if p.tok(1) != COMMA {
			break
		}
		p.next()
	}

	if p.tok(1) != GT {
		return nil
	}
	p.next()
	return x
}

func (p *parser) tryTypeFormalPar() Expr {
	if p.trace {
		defer un(trace(p, "tryTypeFormalPar"))
	}
	x := p.tryTypeIdent()
L:
	for {
		switch p.tok(1) {
		case DOT:
			p.next() // consume '.'
			p.tryTypeIdent()
		case LBRACK:
			p.next() // consume '['

			if p.tok(1) != SUB {
				return nil
			}
			p.next() // consume '-'

			if p.tok(1) != RBRACK {
				return nil
			}
			p.next() // consume ']'

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
	switch p.tok(1) {
	case UNIVERSAL:
		p.next()
		fallthrough
	case IDENT, CHARSTRING:
		p.next()
	default:
		return nil
	}
	if p.tok(1) == LT {
		if x := p.tryTypeFormalPars(); x == nil {
			return nil
		}
	}
	return &Ident{Name: "todo"}
}

/*************************************************************************
 * Module
 *************************************************************************/

func (p *parser) parseModule() *Module {
	if p.trace {
		defer un(trace(p, "Module"))
	}

	pos := p.expect(MODULE)
	name := p.parseIdent()

	if p.tok(1) == LANGUAGE {
		p.parseLanguageSpec()
	}

	p.expect(LBRACE)

	var decls []Decl
	for p.tok(1) != RBRACE && p.tok(1) != EOF {
		decls = append(decls, p.parseModuleDef())
	}
	p.expect(RBRACE)

	return &Module{
		Module: pos,
		Name:   name,
		Decls:  decls,
	}
}

func (p *parser) parseLanguageSpec() {
	p.next()
	for {
		p.expect(STRING)
		if p.tok(1) != COMMA {
			break
		}
		p.next()
	}
}

func (p *parser) parseModuleDef() Decl {
	switch p.tok(1) {
	case PRIVATE, PUBLIC:
		p.next()
	case FRIEND:
		p.next()
		if p.tok(1) == MODULE {
			p.parseFriend()
			p.expectSemi()
			return nil
		}
	}

	switch p.tok(1) {
	case IMPORT:
		p.parseImport()
	case GROUP:
		p.parseGroup()
	case FRIEND:
		p.next()
		p.parseFriend()
	case TYPE:
		p.parseTypeDecl()
	case TEMPLATE:
		p.parseTemplateDecl()
	case MODULEPAR:
		p.parseModulePar()
	case VAR, CONST:
		p.parseValueDecl()
	case SIGNATURE:
		p.parseSignatureDecl()
	case FUNCTION, TESTCASE, ALTSTEP:
		p.parseFuncDecl()
	case CONTROL:
		p.next()
		p.parseBlockStmt()
	case EXTERNAL:
		p.next()
		switch p.tok(1) {
		case FUNCTION:
			p.parseExtFuncDecl()
		case CONST:
			p.parseValueDecl()
		default:
			p.errorExpected(p.pos(1), "'function'")
		}
	default:
		p.errorExpected(p.pos(1), "module definition")
		p.next()
	}
	p.expectSemi()
	return nil
}

/*************************************************************************
 * Import Definition
 *************************************************************************/

func (p *parser) parseImport() Decl {
	if p.trace {
		defer un(trace(p, "Import"))
	}

	pos := p.pos(1)
	p.next()
	p.expect(FROM)

	name := p.parseIdent()

	if p.tok(1) == LANGUAGE {
		p.parseLanguageSpec()
	}

	var specs []ImportSpec
	switch p.tok(1) {
	case ALL:
		p.next()
		if p.tok(1) == EXCEPT {
			p.parseExceptSpec()
		}
	case LBRACE:
		p.parseImportSpec()
	default:
		p.errorExpected(p.pos(1), "'all' or import spec")
	}

	p.parseWith()

	return &ImportDecl{
		ImportPos:   pos,
		Module:      name,
		ImportSpecs: specs,
	}
}

func (p *parser) parseImportSpec() {
	p.expect(LBRACE)
	for p.tok(1) != RBRACE && p.tok(1) != EOF {
		p.parseImportStmt()
	}
	p.expect(RBRACE)
}

func (p *parser) parseImportStmt() {
	switch p.tok(1) {
	case ALTSTEP, CONST, FUNCTION, MODULEPAR,
		SIGNATURE, TEMPLATE, TESTCASE, TYPE:
		p.next()
		if p.tok(1) == ALL {
			p.next()
			if p.tok(1) == EXCEPT {
				p.next()
				p.parseRefList()
			}
		} else {
			p.parseRefList()
		}
	case GROUP:
		p.next()
		for {
			p.parseTypeRef()
			if p.tok(1) == EXCEPT {
				p.parseExceptSpec()
			}
			if p.tok(1) != COMMA {
				break
			}
			p.next()
		}
	case IMPORT:
		p.next()
		p.expect(ALL)
	default:
		p.errorExpected(p.pos(1), "definition qualifier")
	}

	p.expectSemi()
}

func (p *parser) parseExceptSpec() {
	p.next()
	p.expect(LBRACE)
	for p.tok(1) != RBRACE && p.tok(1) != EOF {
		p.parseExceptStmt()
	}
	p.expect(RBRACE)
}

func (p *parser) parseExceptStmt() {
	switch p.tok(1) {
	case ALTSTEP, CONST, FUNCTION, GROUP,
		IMPORT, MODULEPAR, SIGNATURE, TEMPLATE,
		TESTCASE, TYPE:
		p.next()
	default:
		p.errorExpected(p.pos(1), "definition qualifier")
	}

	if p.tok(1) == ALL {
		p.next()
	} else {
		for {
			p.parseTypeRef()
			if p.tok(1) != COMMA {
				break
			}
			p.next()
		}
	}
	p.expectSemi()
}

/*************************************************************************
 * Group Definition
 *************************************************************************/

func (p *parser) parseGroup() {
	p.next()
	p.parseIdent()
	p.expect(LBRACE)

	var decls []Decl
	for p.tok(1) != RBRACE && p.tok(1) != EOF {
		decls = append(decls, p.parseModuleDef())
	}
	p.expect(RBRACE)
	p.parseWith()
}

func (p *parser) parseFriend() {
	p.expect(MODULE)
	p.parseIdent()
	p.parseWith()
}

/*************************************************************************
 * With Attributes
 *************************************************************************/

func (p *parser) parseWith() Node {
	if p.tok(1) != WITH {
		return nil
	}

	p.expect(WITH)
	p.expect(LBRACE)
	for p.tok(1) != RBRACE && p.tok(1) != EOF {
		p.parseWithStmt()
	}
	p.expect(RBRACE)
	return nil
}

func (p *parser) parseWithStmt() Node {
	if p.trace {
		defer un(trace(p, "WithStmt"))
	}
	switch p.tok(1) {
	case ENCODE,
		VARIANT,
		DISPLAY,
		EXTENSION,
		OPTIONAL,
		STEPSIZE,
		OVERRIDE:
		p.next()
	default:
		p.errorExpected(p.pos(1), "with-attribute")
		p.next()
	}

	switch p.tok(1) {
	case OVERRIDE:
		p.next()
	case MODIF:
		p.next() // consume '@local'
	}

	if p.tok(1) == LPAREN {
		p.next()
		for {
			p.parseWithQualifier()
			if p.tok(1) != COMMA {
				break
			}
			p.next()
		}
		p.expect(RPAREN)
	}

	p.expect(STRING)

	if p.tok(1) == DOT {
		p.next()
		p.expect(STRING)
	}

	p.expectSemi()
	return nil
}

func (p *parser) parseWithQualifier() {
	switch p.tok(1) {
	case IDENT:
		p.parseTypeRef()
	case LBRACK:
		p.parseIndexExpr(nil)
	case TYPE, TEMPLATE, CONST, ALTSTEP, TESTCASE, FUNCTION, SIGNATURE, MODULEPAR, GROUP:
		p.next()
		p.expect(ALL)
		if p.tok(1) == EXCEPT {
			p.next()
			p.expect(LBRACE)
			p.parseRefList()
			p.expect(RBRACE)
		}
	default:
		p.errorExpected(p.pos(1), "with-qualifier")
	}
}

/*************************************************************************
 * Type Definitions
 *************************************************************************/

func (p *parser) parseTypeDecl() Decl {
	if p.trace {
		defer un(trace(p, "Type"))
	}
	p.next()
	switch p.tok(1) {
	case ADDRESS, CHARSTRING, IDENT, NULL, UNIVERSAL:
		p.parseSubType()
	case UNION:
		p.next()
		p.parseStructType()
	case SET, RECORD:
		p.next()
		if p.tok(1) == IDENT {
			p.parseStructType()
			break
		}
		p.parseListType()
	case ENUMERATED:
		p.parseEnumType()
	case PORT:
		p.parsePortType()
	case COMPONENT:
		p.parseComponentType()
	case FUNCTION, ALTSTEP, TESTCASE:
		p.parseBehaviourType()
	default:
		p.errorExpected(p.pos(1), "type definition")
	}
	return nil
}

func (p *parser) parseType() {
	if p.trace {
		defer un(trace(p, "NestedType"))
	}
	switch p.tok(1) {
	case ADDRESS, CHARSTRING, IDENT, NULL, UNIVERSAL:
		p.parseTypeRef()
	case UNION:
		p.next()
		p.parseStructBody()
	case SET, RECORD:
		p.next()
		if p.tok(1) == LBRACE {
			p.parseStructBody()
			break
		}
		p.parseListBody()
	case ENUMERATED:
		p.next()
		p.parseEnumBody()

	case FUNCTION, ALTSTEP, TESTCASE:
		p.next()
		p.parseBehaviourTypeBody()
	default:
		p.errorExpected(p.pos(1), "type definition")
	}
}

/*************************************************************************
 * Struct Types
 *************************************************************************/

func (p *parser) parseStructType() {
	if p.trace {
		defer un(trace(p, "StructType"))
	}
	p.parseIdent()
	if p.tok(1) == LT {
		p.parseTypeFormalPars()
	}
	p.parseStructBody()
	p.parseWith()
}

func (p *parser) parseStructBody() {
	if p.trace {
		defer un(trace(p, "StructBody"))
	}
	p.expect(LBRACE)
	for p.tok(1) != RBRACE && p.tok(1) != EOF {
		p.parseStructField()
		if p.tok(1) != COMMA {
			break
		}
		p.next()
	}
	p.expect(RBRACE)
}

func (p *parser) parseStructField() {
	if p.trace {
		defer un(trace(p, "StructField"))
	}
	if p.tok(1) == MODIF {
		p.next() // @default
	}
	p.parseType()
	p.parsePrimaryExpr()

	if p.tok(1) == LPAREN {
		p.parseParenExpr()
	}
	if p.tok(1) == LENGTH {
		p.parseLength(nil)
	}

	if p.tok(1) == OPTIONAL {
		p.next()
	}
}

/*************************************************************************
 * List Type
 *************************************************************************/

func (p *parser) parseListType() {
	if p.trace {
		defer un(trace(p, "ListType"))
	}
	p.parseListBody()
	p.parsePrimaryExpr()
	if p.tok(1) == LT {
		p.parseTypeFormalPars()
	}

	if p.tok(1) == LPAREN {
		p.parseParenExpr()
	}

	if p.tok(1) == LENGTH {
		p.parseLength(nil)
	}

	p.parseWith()
}

func (p *parser) parseListBody() {
	if p.trace {
		defer un(trace(p, "ListBody"))
	}

	if p.tok(1) == LENGTH {
		p.parseLength(nil)
	}

	p.expect(OF)
	p.parseType()
}

/*************************************************************************
 * Enumeration Type
 *************************************************************************/

func (p *parser) parseEnumType() {
	if p.trace {
		defer un(trace(p, "EnumType"))
	}
	p.next()
	p.parseIdent()
	if p.tok(1) == LT {
		p.parseTypeFormalPars()
	}
	p.parseEnumBody()
	p.parseWith()
}

func (p *parser) parseEnumBody() {
	if p.trace {
		defer un(trace(p, "EnumBody"))
	}
	p.expect(LBRACE)
	for p.tok(1) != RBRACE && p.tok(1) != EOF {
		p.parseExpr()
		if p.tok(1) != COMMA {
			break
		}
		p.next()
	}
	p.expect(RBRACE)
}

/*************************************************************************
 * Port Type
 *************************************************************************/

func (p *parser) parsePortType() {
	if p.trace {
		defer un(trace(p, "PortType"))
	}
	p.next()
	p.parseIdent()
	if p.tok(1) == LT {
		p.parseTypeFormalPars()
	}

	switch p.tok(1) {
	case MIXED, MESSAGE, PROCEDURE:
		p.next()
	default:
		p.errorExpected(p.pos(1), "'message' or 'procedure'")
	}

	if p.tok(1) == REALTIME {
		p.next()
	}

	p.expect(LBRACE)
	for p.tok(1) != RBRACE && p.tok(1) != EOF {
		p.parsePortAttribute()
		p.expectSemi()
	}
	p.expect(RBRACE)
	p.parseWith()
}

func (p *parser) parsePortAttribute() {
	if p.trace {
		defer un(trace(p, "PortAttribute"))
	}
	switch p.tok(1) {
	case IN, OUT, INOUT:
		p.next()
		p.parseRefList()
	case ADDRESS:
		p.next()
		p.parseRefList()
	case MAP, UNMAP:
		p.next()
		p.expect(PARAM)
		p.parseFormalPars()
	}
}

/*************************************************************************
 * Component Type
 *************************************************************************/

func (p *parser) parseComponentType() {
	if p.trace {
		defer un(trace(p, "ComponentType"))
	}
	p.next()
	p.parseIdent()
	if p.tok(1) == LT {
		p.parseTypeFormalPars()
	}
	if p.tok(1) == EXTENDS {
		p.next()
		p.parseRefList()
	}
	p.parseBlockStmt()
	p.parseWith()
}

/*************************************************************************
 * Behaviour Types
 *************************************************************************/

func (p *parser) parseBehaviourTypeBody() {
	if p.trace {
		defer un(trace(p, "BehaviourTypeBody"))
	}
	p.parseFormalPars()

	if p.tok(1) == RUNS {
		p.parseRunsOn()
	}

	if p.tok(1) == SYSTEM {
		p.parseSystem()
	}

	if p.tok(1) == RETURN {
		p.parseReturn()
	}
}

func (p *parser) parseBehaviourType() {
	if p.trace {
		defer un(trace(p, "BehaviourType"))
	}
	p.next()
	p.next()
	if p.tok(1) == LT {
		p.parseTypeFormalPars()
	}
	p.parseBehaviourTypeBody()
	p.parseWith()

}

/*************************************************************************
 * Subtype
 *************************************************************************/

func (p *parser) parseSubType() *SubType {
	if p.trace {
		defer un(trace(p, "SubType"))
	}

	p.parseType()
	p.parsePrimaryExpr()
	if p.tok(1) == LT {
		p.parseTypeFormalPars()
	}
	// TODO(mef) fix constraints consumed by previous PrimaryExpr

	if p.tok(1) == LPAREN {
		p.parseParenExpr()
	}
	if p.tok(1) == LENGTH {
		p.parseLength(nil)
	}

	p.parseWith()
	return nil
}

/*************************************************************************
 * Template Declaration
 *************************************************************************/

func (p *parser) parseTemplateDecl() *ValueDecl {
	if p.trace {
		defer un(trace(p, "TemplateDecl"))
	}

	x := &ValueDecl{DeclPos: p.pos(1), Kind: p.tok(1)}
	p.next()

	if p.tok(1) == LPAREN {
		p.next() // consume '('
		p.next() // consume omit/value/...
		p.expect(RPAREN)
	}

	if p.tok(1) == MODIF {
		p.next()
	}

	x.Type = p.parseTypeRef()
	p.parseIdent()
	if p.tok(1) == LT {
		p.parseTypeFormalPars()
	}
	if p.tok(1) == LPAREN {
		p.parseFormalPars()
	}
	if p.tok(1) == MODIFIES {
		p.next()
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

	x := &ValueDecl{DeclPos: p.pos(1), Kind: p.tok(1)}
	p.next()

	if p.tok(1) == LBRACE {
		p.next()
		for p.tok(1) != RBRACE && p.tok(1) != EOF {
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

	x := &ValueDecl{DeclPos: p.pos(1), Kind: p.tok(1)}
	p.next()
	p.parseRestrictionSpec()

	if p.tok(1) == MODIF {
		p.next()
	}

	if x.Kind != TIMER {
		x.Type = p.parseTypeRef()
	}
	x.Decls = p.parseExprList()
	p.parseWith()
	return x
}

func (p *parser) parseRestrictionSpec() *RestrictionSpec {
	switch p.tok(1) {
	case TEMPLATE:
		x := &RestrictionSpec{Kind: p.tok(1), KindPos: p.pos(1)}
		p.next()
		if p.tok(1) != LPAREN {
			return x
		}

		p.next()
		x.Kind = p.tok(1)
		x.KindPos = p.pos(1)
		p.next()
		p.expect(RPAREN)

	case OMIT, VALUE, PRESENT:
		x := &RestrictionSpec{Kind: p.tok(1), KindPos: p.pos(1)}
		p.next()
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

	x := &FuncDecl{FuncPos: p.pos(1), Kind: p.tok(1)}
	p.next()
	x.Name = p.parseIdent()
	if p.tok(1) == LT {
		p.parseTypeFormalPars()
	}

	if p.tok(1) == MODIF {
		p.next()
	}

	x.Params = p.parseFormalPars()

	if p.tok(1) == RUNS {
		p.parseRunsOn()
	}

	if p.tok(1) == MTC {
		p.parseMtc()
	}

	if p.tok(1) == SYSTEM {
		p.parseSystem()
	}

	if p.tok(1) == RETURN {
		x.Return = p.parseReturn()
	}

	if p.tok(1) == LBRACE {
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

	x := &FuncDecl{FuncPos: p.pos(1), Kind: p.tok(1)}
	p.next()
	x.Name = p.parseIdent()

	if p.tok(1) == MODIF {
		p.next()
	}

	x.Params = p.parseFormalPars()

	if p.tok(1) == RUNS {
		p.parseRunsOn()
	}

	if p.tok(1) == MTC {
		p.parseMtc()
	}

	if p.tok(1) == SYSTEM {
		p.parseSystem()
	}

	if p.tok(1) == RETURN {
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

	p.next()
	p.parseIdent()
	if p.tok(1) == LT {
		p.parseTypeFormalPars()
	}

	p.parseFormalPars()

	if p.tok(1) == NOBLOCK {
		p.next()
	}

	if p.tok(1) == RETURN {
		p.parseReturn()
	}

	if p.tok(1) == EXCEPTION {
		p.next()
		p.parseParenExpr()
	}
	p.parseWith()
	return nil
}

func (p *parser) parseRunsOn() {
	p.expect(RUNS)
	p.expect(ON)
	p.parseTypeRef()
}

func (p *parser) parseSystem() {
	p.expect(SYSTEM)
	p.parseTypeRef()
}

func (p *parser) parseMtc() {
	p.expect(MTC)
	p.parseTypeRef()
}

func (p *parser) parseReturn() Expr {
	p.next()
	p.parseRestrictionSpec()
	if p.tok(1) == MODIF {
		p.next()
	}
	return p.parseTypeRef()
}

func (p *parser) parseFormalPars() *FormalPars {
	if p.trace {
		defer un(trace(p, "FormalPars"))
	}
	x := &FormalPars{}
	p.expect(LPAREN)
	for p.tok(1) != RPAREN {
		x.List = append(x.List, p.parseFormalPar())
		if p.tok(1) != COMMA {
			break
		}
		p.next()
	}
	p.expect(RPAREN)
	return x
}

func (p *parser) parseFormalPar() *FormalPar {
	if p.trace {
		defer un(trace(p, "FormalPar"))
	}
	x := &FormalPar{}

	switch p.tok(1) {
	case IN:
		p.next()
	case OUT:
		p.next()
	case INOUT:
		p.next()
	}

	p.parseRestrictionSpec()
	if p.tok(1) == MODIF {
		p.next()
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
	for p.tok(1) != GT {
		p.parseTypeFormalPar()
		if p.tok(1) != COMMA {
			break
		}
		p.next()
	}
	p.expect(GT)
}

func (p *parser) parseTypeFormalPar() {
	if p.trace {
		defer un(trace(p, "TypeFormalPar"))
	}
	if p.tok(1) == IN {
		p.next()
	}

	switch p.tok(1) {
	case TYPE:
		p.next()
	case SIGNATURE:
		p.next()
	default:
		p.parseTypeRef()
	}
	p.expect(IDENT)
	if p.tok(1) == ASSIGN {
		p.next()
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

	x := &BlockStmt{LBrace: p.pos(1)}
	p.expect(LBRACE)
	for p.tok(1) != RBRACE && p.tok(1) != EOF {
		x.Stmts = append(x.Stmts, p.parseStmt())
	}
	p.expect(RBRACE)
	return x
}

func (p *parser) parseStmt() Stmt {
	if p.trace {
		defer un(trace(p, "Stmt"))
	}

	switch p.tok(1) {
	case TEMPLATE:
		p.parseTemplateDecl()
	case VAR, CONST, TIMER, PORT:
		p.parseValueDecl()
	case REPEAT, BREAK, CONTINUE:
		p.next()
	case LABEL:
		p.next()
		p.expect(IDENT)
	case GOTO:
		p.next()
		p.expect(IDENT)
	case RETURN:
		p.next()
		if p.tok(1) != SEMICOLON && p.tok(1) != RBRACE {
			p.parseExpr()
		}
	case SELECT:
		p.parseSelect()
	case ALT, INTERLEAVE:
		p.next()
		p.parseBlockStmt()
	case LBRACK:
		p.parseAltGuard()
	case FOR:
		p.parseForLoop()
	case WHILE:
		p.parseWhileLoop()
	case DO:
		p.parseDoWhileLoop()
	case IF:
		p.parseIfStmt()
	default:
		if p.tok(1) == LBRACE {
			p.parseBlockStmt()
			break
		}

		p.parseSimpleStmt()

		// call-statement block
		if p.tok(1) == LBRACE {
			p.parseBlockStmt()
		}
	}
	p.expectSemi()
	return nil
}

func (p *parser) parseForLoop() {
	p.next()
	p.expect(LPAREN)
	if p.tok(1) == VAR {
		p.parseValueDecl()
	} else {
		p.parseExpr()
	}
	p.expect(SEMICOLON)
	p.parseExpr()
	p.expect(SEMICOLON)
	p.parseExpr()
	p.expect(RPAREN)
	p.parseBlockStmt()
}

func (p *parser) parseWhileLoop() {
	p.next()
	p.parseParenExpr()
	p.parseBlockStmt()
}

func (p *parser) parseDoWhileLoop() {
	p.next()
	p.parseBlockStmt()
	p.expect(WHILE)
	p.parseParenExpr()
}

func (p *parser) parseIfStmt() {
	p.next()
	p.parseParenExpr()
	p.parseBlockStmt()
	if p.tok(1) == ELSE {
		p.next()
		if p.tok(1) == IF {
			p.parseIfStmt()
		} else {
			p.parseBlockStmt()
		}
	}
}

func (p *parser) parseSelect() {
	p.expect(SELECT)
	if p.tok(1) == UNION {
		p.next()
	}
	p.parseParenExpr()
	p.expect(LBRACE)
	for p.tok(1) == CASE {
		p.parseCaseStmt()
	}
	p.expect(RBRACE)
}

func (p *parser) parseCaseStmt() {
	p.expect(CASE)
	if p.tok(1) == ELSE {
		p.next()
	} else {
		p.parseParenExpr()
	}
	p.parseBlockStmt()
}

func (p *parser) parseAltGuard() {
	p.next()
	if p.tok(1) == ELSE {
		p.next()
		p.expect(RBRACK)
		p.parseBlockStmt()
		return
	}

	if p.tok(1) != RBRACK {
		p.parseExpr()
	}
	p.expect(RBRACK)
	p.parseSimpleStmt()
	if p.tok(1) == LBRACE {
		p.parseBlockStmt()
	}
}

func (p *parser) parseSimpleStmt() Stmt {
	if p.trace {
		defer un(trace(p, "SimpleStmt"))
	}

	p.parseExpr()

	return nil
}
