package syntax

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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

// ParseModule parses the source code of a single Go source file and returns
// the corresponding Module node. The source code may be provided via
// the filename of the source file, or via the src parameter.
//
// If src != nil, ParseModule parses the source from src and the filename is
// only used when recording position information. The type of the argument
// for the src parameter must be string, []byte, or io.Reader.
// If src == nil, ParseModule parses the file specified by filename.
//
// The mode parameter controls the amount of source text parsed and other
// optional parser functionality. Position information is recorded in the
// file set fset, which must not be nil.
//
// If the source couldn't be read, the returned AST is nil and the error
// indicates the specific failure. If the source was read but syntax
// errors were found, the result is a partial AST (with Bad* nodes
// representing the fragments of erroneous source code). Multiple errors
// are returned via a ErrorList which is sorted by file position.
//
func ParseModule(fset *FileSet, filename string, src interface{}, mode Mode, eh ErrorHandler) (f *Module, err error) {
	if fset == nil {
		panic("ParseModule: no FileSet provided (fset == nil)")
	}

	// get source
	text, err := readSource(filename, src)
	if err != nil {
		return nil, err
	}

	var p parser
	defer func() {
		if e := recover(); e != nil {
			// resume same panic if it's not a bailout
			if _, ok := e.(bailout); !ok {
				panic(e)
			}
		}

		// set result values
		if f == nil {
			// source is not a valid Go source file - satisfy
			// ParseModule API and return a valid (but) empty
			// *Module
			f = &Module{
				Name: new(Ident),
			}
		}

		p.errors.Sort()
		err = p.errors.Err()
	}()

	// parse source
	p.init(fset, filename, text, mode, eh)
	f = p.parseModule()

	return
}

// If src != nil, readSource converts src to a []byte if possible;
// otherwise it returns an error. If src == nil, readSource returns
// the result of reading the file specified by filename.
//
func readSource(filename string, src interface{}) ([]byte, error) {
	if src != nil {
		switch s := src.(type) {
		case string:
			return []byte(s), nil
		case []byte:
			return s, nil
		case *bytes.Buffer:
			// is io.Reader, but src is already available in []byte form
			if s != nil {
				return s.Bytes(), nil
			}
		case io.Reader:
			return ioutil.ReadAll(s)
		}
		return nil, errors.New("invalid source")
	}
	return ioutil.ReadFile(filename)
}

// The parser structure holds the parser's internal state.
type parser struct {
	file    *File
	errors  ErrorList
	scanner Scanner

	// Tracing/debugging
	mode   Mode // parsing mode
	trace  bool // == (mode & Trace != 0)
	indent int  // indentation used for tracing output

	// Comments
	comments    []*CommentGroup
	leadComment *CommentGroup // last lead comment
	lineComment *CommentGroup // last line comment

	// Next token
	pos Pos    // token position
	tok Token  // one token look-ahead
	lit string // token literal

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

	p.next()
}

// ----------------------------------------------------------------------------
// Parsing support

func (p *parser) printTrace(a ...interface{}) {
	const dots = ". . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . . "
	const n = len(dots)
	pos := p.file.Position(p.pos)
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

// Usage pattern: defer un(trace(p, "..."))
func un(p *parser) {
	p.indent--
	p.printTrace(")")
}

// Advance to the next
func (p *parser) next0() {
	// Because of one-token look-ahead, print the previous token
	// when tracing as it provides a more readable output. The
	// very first token (!p.pos.IsValid()) is not initialized
	// (it is ILLEGAL), so don't print it .
	if p.trace && p.pos.IsValid() {
		s := p.tok.String()
		switch {
		case p.tok.IsLiteral():
			p.printTrace(s, p.lit)
		case p.tok.IsOperator(), p.tok.IsKeyword():
			p.printTrace("\"" + s + "\"")
		default:
			p.printTrace(s)
		}
	}

	p.pos, p.tok, p.lit = p.scanner.Scan()
}

// Consume a comment and return it and the line on which it ends.
func (p *parser) consumeComment() (comment *Comment, endline int) {
	// /*-style comments may end on a different line than where they start.
	// Scan the comment for '\n' chars and adjust endline accordingly.
	endline = p.file.Line(p.pos)
	if p.lit[1] == '*' {
		// don't use range here - no need to decode Unicode code points
		for i := 0; i < len(p.lit); i++ {
			if p.lit[i] == '\n' {
				endline++
			}
		}
	}

	comment = &Comment{Slash: p.pos, Text: p.lit}
	p.next0()

	return
}

// Consume a group of adjacent comments, add it to the parser's
// comments list, and return it together with the line at which
// the last comment in the group ends. A non-comment token or n
// empty lines terminate a comment group.
//
func (p *parser) consumeCommentGroup(n int) (comments *CommentGroup, endline int) {
	var list []*Comment
	endline = p.file.Line(p.pos)
	for p.tok == COMMENT && p.file.Line(p.pos) <= endline+n {
		var comment *Comment
		comment, endline = p.consumeComment()
		list = append(list, comment)
	}

	// add comment group to the comments list
	comments = &CommentGroup{List: list}
	p.comments = append(p.comments, comments)

	return
}

// Advance to the next non-comment  In the process, collect
// any comment groups encountered, and remember the last lead and
// and line comments.
//
// A lead comment is a comment group that starts and ends in a
// line without any other tokens and that is followed by a non-comment
// token on the line immediately after the comment group.
//
// A line comment is a comment group that follows a non-comment
// token on the same line, and that has no tokens after it on the line
// where it ends.
//
// Lead and line comments may be considered documentation that is
// stored in the AST.
//
func (p *parser) next() {
	p.leadComment = nil
	p.lineComment = nil
	p.seenBrace = false

	prev := p.pos

	if p.tok == RBRACE {
		p.seenBrace = true
	}

	p.next0()

	if p.tok == COMMENT {
		var comment *CommentGroup
		var endline int

		if p.file.Line(p.pos) == p.file.Line(prev) {
			// The comment is on same line as the previous token; it
			// cannot be a lead comment but may be a line comment.
			comment, endline = p.consumeCommentGroup(0)
			if p.file.Line(p.pos) != endline || p.tok == EOF {
				// The next token is on a different line, thus
				// the last comment group is a line comment.
				p.lineComment = comment
			}
		}

		// consume successor comments, if any
		endline = -1
		for p.tok == COMMENT {
			comment, endline = p.consumeCommentGroup(1)
		}

		if endline+1 == p.file.Line(p.pos) {
			// The next token is following on the line immediately after the
			// comment group, thus the last comment group is a lead comment.
			p.leadComment = comment
		}
	}
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
	if pos == p.pos {
		// the error happened at the current position;
		// make the error message more specific
		switch {
		case p.tok.IsLiteral():
			// print 123 rather than 'INT', etc.
			msg += ", found " + p.lit
		default:
			msg += ", found '" + p.tok.String() + "'"
		}
	}
	p.error(pos, msg)
}

func (p *parser) expect(tok Token) Pos {
	pos := p.pos
	if p.tok != tok {
		p.errorExpected(pos, "'"+tok.String()+"'")
	}
	p.next() // make progress
	return pos
}

func (p *parser) expectSemi() {
	switch p.tok {
	case SEMICOLON:
		p.next()
	case RBRACE, EOF:
		// semicolon is optional before a closing '}'
	default:
		if !p.seenBrace {
			p.errorExpected(p.pos, "';'")
			p.advance(stmtStart)
		}
	}
}

// advance consumes tokens until the current token p.tok
// is in the 'to' set, or EOF. For error recovery.
func (p *parser) advance(to map[Token]bool) {
	for ; p.tok != EOF; p.next() {
		if to[p.tok] {
			// Return only if parser made some progress since last
			// sync or if it has not reached 10 advance calls without
			// progress. Otherwise consume at least one token to
			// avoid an endless parser loop (it is possible that
			// both parseOperand and parseStmt call advance and
			// correctly do not advance, thus the need for the
			// invocation limit p.syncCnt).
			if p.pos == p.syncPos && p.syncCnt < 10 {
				p.syncCnt++
				return
			}
			if p.pos > p.syncPos {
				p.syncPos = p.pos
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

var stmtStart = map[Token]bool{
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

func (p *parser) parseExprList() (list []Expr) {
	list = append(list, p.parseExpr())
	for p.tok == COMMA {
		p.next()
		list = append(list, p.parseExpr())
	}
	return list
}

func (p *parser) parseExpr() Expr {
	if p.trace {
		defer un(trace(p, "Expr"))
	}

	x := p.parseBinaryExpr(LowestPrec + 1)

	if p.tok == ASSIGN {
		p.next()
		p.parseExpr()
	}

	return x
}

func (p *parser) parseBinaryExpr(prec1 int) Expr {
	x := p.parseUnaryExpr()
	for {
		prec := p.tok.Precedence()
		if prec < prec1 {
			return x
		}
		pos := p.pos
		op := p.tok
		p.next()

		y := p.parseBinaryExpr(prec + 1)

		x = &BinaryExpr{X: x, Op: op, OpPos: pos, Y: y}
	}
}

func (p *parser) parseUnaryExpr() Expr {
	switch p.tok {
	case ADD,
		EXCL,
		NOT,
		NOT4B,
		SUB:
		op, pos := p.tok, p.pos
		p.next()
		// handle unused expr '-'
		if op == SUB && (p.tok == COMMA || p.tok == SEMICOLON || p.tok == RBRACE || p.tok == RBRACK || p.tok == RPAREN || p.tok == EOF) {
			return nil
		}
		return &UnaryExpr{Op: op, OpPos: pos, X: p.parseUnaryExpr()}

	case MODIFIES:
		p.next()
		p.parsePrimaryExpr()
		p.expect(ASSIGN)
		p.parseExpr()
		return nil
	}
	return p.parsePrimaryExpr()
}

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
		case COLON:
			p.next()
			p.parseExpr()
		default:
			break L
		}
	}

	if p.tok == LENGTH {
		p.parseLength()
	}

	if p.tok == IFPRESENT {
		p.next()
	}

	if p.tok == TO || p.tok == FROM {
		p.next()
		p.parseExpr()
	}

	if p.tok == REDIR {
		p.parseRedirect()
	}

	if p.tok == VALUE {
		p.next()
		p.parseExpr()
	}

	if p.tok == PARAM {
		p.next()
		p.parseSetExpr()
	}

	if p.tok == ALIVE {
		p.next()
	}

	return x
}

func (p *parser) parseOperand() Expr {
	switch p.tok {
	case ANYKW, ALL:
		k := p.tok
		p.next()
		switch p.tok {
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

		p.errorExpected(p.pos, "'component', 'port', 'timer' or 'from'")

	case UNIVERSAL:
		p.parseUniversalCharstring()
		id := &Ident{NamePos: p.pos, Name: p.lit}
		return id
	case IDENT,
		ADDRESS,
		CHARSTRING,
		MAP,
		MTC,
		NULL,
		OMIT,
		SYSTEM,
		TESTCASE,
		TIMER,
		UNMAP:
		id := &Ident{NamePos: p.pos, Name: p.lit}
		p.next()
		return id

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
		lit := &ValueLiteral{Kind: p.tok, ValuePos: p.pos, Value: p.lit}
		p.next()
		return lit

	case LPAREN:
		// can be template `x := (1,2,3)`, but also artihmetic expression: `1*(2+3)`
		p.parseSetExpr()

	case LBRACK:
		p.parseIndexExpr(nil)

	case LBRACE:
		p.parseCompositeLiteral()

	case REGEXP:
		p.parseCallRegexp()

	case PATTERN:
		p.parseCallPattern()

	case DECMATCH:
		p.parseCallDecMatch()

	case MODIF:
		p.parseCallDecoded()

	default:
		p.errorExpected(p.pos, "operand")
	}

	return nil
}

func (p *parser) parseSetExpr() {
	p.expect(LPAREN)
	p.parseExprList()
	p.expect(RPAREN)
}

func (p *parser) parseUniversalCharstring() {
	p.expect(UNIVERSAL)
	p.expect(CHARSTRING)
}

func (p *parser) parseCompositeLiteral() {
	p.expect(LBRACE)
	if p.tok != RBRACE {
		p.parseExprList()
	}
	p.expect(RBRACE)
}

func (p *parser) parseCallRegexp() {
	p.expect(REGEXP)
	if p.tok == MODIF {
		p.next()
	}
	p.parseSetExpr()
}

func (p *parser) parseCallPattern() {
	p.expect(PATTERN)
	if p.tok == MODIF {
		p.next()
	}
	p.expect(STRING)
}

func (p *parser) parseCallDecMatch() {
	p.expect(DECMATCH)
	if p.tok == LPAREN {
		p.parseSetExpr()
	}
	p.parseExpr()
}

func (p *parser) parseCallDecoded() {
	p.expect(MODIF) // @decoded
	if p.tok == LPAREN {
		p.parseSetExpr()
	}
	p.parseExpr()
}

func (p *parser) parseSelectorExpr(x Expr) Expr {
	p.expect(DOT)
	return &SelectorExpr{X: x, Sel: p.parseIdent()}
}

func (p *parser) parseIndexExpr(x Expr) Expr {
	p.expect(LBRACK)
	x = &IndexExpr{X: x, Index: p.parseExpr()}
	p.expect(RBRACK)
	return x
}

func (p *parser) parseCallExpr(x Expr) Expr {
	p.next()

	switch p.tok {
	case FROM, TO:
		p.next()
		p.parseExpr()
		if p.tok == REDIR {
			p.parseRedirect()
		}
		p.expect(RPAREN)
		return nil
	case REDIR:
		p.parseRedirect()
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

func (p *parser) parseLength() {
	p.expect(LENGTH)
	p.parseSetExpr()
}

func (p *parser) parseRedirect() Expr {
	p.next()

	if p.tok == VALUE {
		p.next()
		p.parseExprList()
	}

	if p.tok == PARAM {
		p.next()
		p.parseExprList()
	}

	if p.tok == SENDER {
		p.next()
		p.parsePrimaryExpr()
	}

	if p.tok == MODIF {
		p.next() // @index

		if p.tok == VALUE {
			p.next() // optional
		}
		p.parsePrimaryExpr()
	}

	if p.tok == TIMESTAMP {
		p.next()
		p.parsePrimaryExpr()
	}

	return nil
}

func (p *parser) parseIdent() *Ident {
	pos := p.pos
	name := "_"
	switch p.tok {
	case UNIVERSAL:
		p.parseUniversalCharstring()
	case IDENT, ADDRESS, ALIVE, CHARSTRING:
		name = p.lit
		p.next()
	default:
		p.expect(IDENT) // use expect() error handling
	}
	return &Ident{NamePos: pos, Name: name}
}

func (p *parser) parseRefList() {
	for {
		p.parseTypeRef()
		if p.tok != COMMA {
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

/*************************************************************************
 * Module
 *************************************************************************/

func (p *parser) parseModule() *Module {
	if p.trace {
		defer un(trace(p, "Module"))
	}

	pos := p.expect(MODULE)
	name := p.parseIdent()

	if p.tok == LANGUAGE {
		p.parseLanguageSpec()
	}

	p.expect(LBRACE)

	var decls []Decl
	for p.tok != RBRACE && p.tok != EOF {
		decls = append(decls, p.parseModuleDef())
	}
	p.expect(RBRACE)

	return &Module{
		Module:   pos,
		Name:     name,
		Decls:    decls,
		Comments: p.comments,
	}
}

func (p *parser) parseLanguageSpec() {
	p.next()
	for {
		p.expect(STRING)
		if p.tok != COMMA {
			break
		}
		p.next()
	}
}

func (p *parser) parseModuleDef() Decl {
	switch p.tok {
	case PRIVATE, PUBLIC:
		p.next()
	case FRIEND:
		p.next()
		if p.tok == MODULE {
			p.parseFriend()
			p.expectSemi()
			return nil
		}
	}

	switch p.tok {
	case IMPORT:
		p.parseImport()
	case GROUP:
		p.parseGroup()
	case FRIEND:
		p.next()
		p.parseFriend()
	case TYPE:
		p.parseType()
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
		switch p.tok {
		case FUNCTION:
			p.parseExtFuncDecl()
		case CONST:
			p.parseValueDecl()
		default:
			p.errorExpected(p.pos, "'function'")
		}
	default:
		p.errorExpected(p.pos, "module definition")
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

	pos := p.pos
	p.next()
	p.expect(FROM)

	name := p.parseIdent()

	if p.tok == LANGUAGE {
		p.parseLanguageSpec()
	}

	var specs []ImportSpec
	switch p.tok {
	case ALL:
		p.next()
		if p.tok == EXCEPT {
			p.parseExceptSpec()
		}
	case LBRACE:
		p.parseImportSpec()
	default:
		p.errorExpected(p.pos, "'all' or import spec")
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
	for p.tok != RBRACE && p.tok != EOF {
		p.parseImportStmt()
	}
	p.expect(RBRACE)
}

func (p *parser) parseImportStmt() {
	switch p.tok {
	case ALTSTEP, CONST, FUNCTION, MODULEPAR,
		SIGNATURE, TEMPLATE, TESTCASE, TYPE:
		p.next()
		if p.tok == ALL {
			p.next()
			if p.tok == EXCEPT {
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
			if p.tok == EXCEPT {
				p.parseExceptSpec()
			}
			if p.tok != COMMA {
				break
			}
			p.next()
		}
	case IMPORT:
		p.next()
		p.expect(ALL)
	default:
		p.errorExpected(p.pos, "definition qualifier")
	}

	p.expectSemi()
}

func (p *parser) parseExceptSpec() {
	p.next()
	p.expect(LBRACE)
	for p.tok != RBRACE && p.tok != EOF {
		p.parseExceptStmt()
	}
	p.expect(RBRACE)
}

func (p *parser) parseExceptStmt() {
	switch p.tok {
	case ALTSTEP, CONST, FUNCTION, GROUP,
		IMPORT, MODULEPAR, SIGNATURE, TEMPLATE,
		TESTCASE, TYPE:
		p.next()
	default:
		p.errorExpected(p.pos, "definition qualifier")
	}

	if p.tok == ALL {
		p.next()
	} else {
		for {
			p.parseTypeRef()
			if p.tok != COMMA {
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
	for p.tok != RBRACE && p.tok != EOF {
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
	if p.tok != WITH {
		return nil
	}

	p.expect(WITH)
	p.expect(LBRACE)
	for p.tok != RBRACE && p.tok != EOF {
		p.parseWithStmt()
	}
	p.expect(RBRACE)
	return nil
}

func (p *parser) parseWithStmt() Node {
	if p.trace {
		defer un(trace(p, "WithStmt"))
	}
	switch p.tok {
	case ENCODE,
		VARIANT,
		DISPLAY,
		EXTENSION,
		OPTIONAL,
		STEPSIZE,
		OVERRIDE:
		p.next()
	default:
		p.errorExpected(p.pos, "with-attribute")
		p.next()
	}

	switch p.tok {
	case OVERRIDE:
		p.next()
	case MODIF:
		p.next() // consume '@local'
	}

	if p.tok == LPAREN {
		p.next()
		for {
			p.parseWithQualifier()
			if p.tok != COMMA {
				break
			}
			p.next()
		}
		p.expect(RPAREN)
	}

	p.expect(STRING)

	if p.tok == DOT {
		p.next()
		p.expect(STRING)
	}

	p.expectSemi()
	return nil
}

func (p *parser) parseWithQualifier() {
	switch p.tok {
	case IDENT:
		p.parseTypeRef()
	case LBRACK:
		p.parseIndexExpr(nil)
	case TYPE, TEMPLATE, CONST, ALTSTEP, TESTCASE, FUNCTION, SIGNATURE, MODULEPAR, GROUP:
		p.next()
		p.expect(ALL)
		if p.tok == EXCEPT {
			p.next()
			p.expect(LBRACE)
			p.parseRefList()
			p.expect(RBRACE)
		}
	default:
		p.errorExpected(p.pos, "with-qualifier")
	}
}

/*************************************************************************
 * Type Definitions
 *************************************************************************/

func (p *parser) parseType() Decl {
	if p.trace {
		defer un(trace(p, "Type"))
	}
	p.next()
	switch p.tok {
	case ADDRESS, CHARSTRING, IDENT, NULL, UNIVERSAL:
		p.parseSubType()
	case UNION:
		p.next()
		p.parseStructType()
	case SET, RECORD:
		p.next()
		if p.tok == IDENT {
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
		p.errorExpected(p.pos, "type definition")
	}
	return nil
}

func (p *parser) parseNestedType() {
	if p.trace {
		defer un(trace(p, "NestedType"))
	}
	switch p.tok {
	case ADDRESS, CHARSTRING, IDENT, NULL, UNIVERSAL:
		p.parseTypeRef()
	case UNION:
		p.next()
		p.parseStructBody()
	case SET, RECORD:
		p.next()
		if p.tok == LBRACE {
			p.parseStructBody()
			break
		}
		p.parseListBody()
	case ENUMERATED:
		p.parseEnumBody()
	default:
		p.errorExpected(p.pos, "type definition")
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
	if p.tok == LT {
		p.parseTypeParameters()
	}
	p.parseStructBody()
	p.parseWith()
}

func (p *parser) parseStructBody() {
	if p.trace {
		defer un(trace(p, "StructBody"))
	}
	p.expect(LBRACE)
	for p.tok != RBRACE && p.tok != EOF {
		p.parseStructField()
		if p.tok != COMMA {
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
	if p.tok == MODIF {
		p.next() // @default
	}
	p.parseNestedType()
	p.parsePrimaryExpr()

	if p.tok == LPAREN {
		p.parseSetExpr()
	}
	if p.tok == LENGTH {
		p.parseLength()
	}

	if p.tok == OPTIONAL {
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
	if p.tok == LT {
		p.parseTypeParameters()
	}

	if p.tok == LPAREN {
		p.parseSetExpr()
	}

	if p.tok == LENGTH {
		p.parseLength()
	}

	p.parseWith()
}

func (p *parser) parseListBody() {
	if p.trace {
		defer un(trace(p, "ListBody"))
	}

	if p.tok == LENGTH {
		p.parseLength()
	}

	p.expect(OF)
	p.parseNestedType()
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
	if p.tok == LT {
		p.parseTypeParameters()
	}
	p.parseEnumBody()
	p.parseWith()
}

func (p *parser) parseEnumBody() {
	if p.trace {
		defer un(trace(p, "EnumBody"))
	}
	p.expect(LBRACE)
	for p.tok != RBRACE && p.tok != EOF {
		p.parseExpr()
		if p.tok != COMMA {
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
	if p.tok == LT {
		p.parseTypeParameters()
	}

	switch p.tok {
	case MIXED, MESSAGE, PROCEDURE:
		p.next()
	default:
		p.errorExpected(p.pos, "'message' or 'procedure'")
	}

	p.expect(LBRACE)
	for p.tok != RBRACE && p.tok != EOF {
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
	switch p.tok {
	case IN, OUT, INOUT:
		p.next()
		p.parseRefList()
	case ADDRESS:
		p.next()
		p.parseRefList()
	case MAP, UNMAP:
		p.next()
		p.expect(PARAM)
		p.parseParameters()
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
	if p.tok == LT {
		p.parseTypeParameters()
	}
	if p.tok == EXTENDS {
		p.next()
		p.parseRefList()
	}
	p.parseBlockStmt()
	p.parseWith()
}

/*************************************************************************
 * Behaviour Types
 *************************************************************************/

func (p *parser) parseBehaviourType() {
	if p.trace {
		defer un(trace(p, "BehaviourType"))
	}
	p.next()
	p.next()
	p.parseParameters()

	if p.tok == RUNS {
		p.parseRunsOn()
	}

	if p.tok == SYSTEM {
		p.parseSystem()
	}

	if p.tok == RETURN {
		p.parseReturn()
	}
	p.parseWith()

}

/*************************************************************************
 * Subtype
 *************************************************************************/

func (p *parser) parseSubType() *SubType {
	if p.trace {
		defer un(trace(p, "SubType"))
	}

	p.parseNestedType()
	p.parsePrimaryExpr()
	if p.tok == LT {
		p.parseTypeParameters()
	}
	// TODO(mef) fix constraints consumed by previous PrimaryExpr

	if p.tok == LPAREN {
		p.parseSetExpr()
	}
	if p.tok == LENGTH {
		p.parseLength()
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

	x := &ValueDecl{DeclPos: p.pos, Kind: p.tok}
	p.next()

	if p.tok == LPAREN {
		p.next() // consume '('
		p.next() // consume omit/value/...
		p.expect(RPAREN)
	}

	if p.tok == MODIF {
		p.next()
	}

	x.Type = p.parseTypeRef()
	p.parseIdent()
	if p.tok == LT {
		p.parseTypeParameters()
	}
	if p.tok == LPAREN {
		p.parseParameters()
	}
	if p.tok == MODIFIES {
		p.next()
		p.parseIdent()
	}
	p.expect(ASSIGN)
	p.parseExpr()

	p.parseWith()
	return x
}

/*************************************************************************
 * Module Parameter
 *************************************************************************/

func (p *parser) parseModulePar() *ValueDecl {
	if p.trace {
		defer un(trace(p, "ModulePar"))
	}

	x := &ValueDecl{DeclPos: p.pos, Kind: p.tok}
	p.next()

	if p.tok == LBRACE {
		p.next()
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

	x := &ValueDecl{DeclPos: p.pos, Kind: p.tok}
	p.next()
	p.parseRestrictionSpec()

	if p.tok == MODIF {
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
	switch p.tok {
	case TEMPLATE:
		x := &RestrictionSpec{Kind: p.tok, KindPos: p.pos}
		p.next()
		if p.tok != LPAREN {
			return x
		}

		p.next()
		x.Kind = p.tok
		x.KindPos = p.pos
		p.next()
		p.expect(RPAREN)

	case OMIT, VALUE, PRESENT:
		x := &RestrictionSpec{Kind: p.tok, KindPos: p.pos}
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

	x := &FuncDecl{FuncPos: p.pos, Kind: p.tok}
	p.next()
	x.Name = p.parseIdent()
	if p.tok == LT {
		p.parseTypeParameters()
	}

	if p.tok == MODIF {
		p.next()
	}

	x.Params = p.parseParameters()

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

	x := &FuncDecl{FuncPos: p.pos, Kind: p.tok}
	p.next()
	x.Name = p.parseIdent()

	if p.tok == MODIF {
		p.next()
	}

	x.Params = p.parseParameters()

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

	p.next()
	p.parseIdent()
	if p.tok == LT {
		p.parseTypeParameters()
	}

	p.parseParameters()

	if p.tok == NOBLOCK {
		p.next()
	}

	if p.tok == RETURN {
		p.parseReturn()
	}

	if p.tok == EXCEPTION {
		p.next()
		p.parseSetExpr()
	}
	p.parseWith()
	return nil
}

func (p *parser) parseReturn() Expr {
	p.next()
	p.parseRestrictionSpec()
	if p.tok == MODIF {
		p.next()
	}
	return p.parseTypeRef()
}

func (p *parser) parseParameters() *FieldList {
	if p.trace {
		defer un(trace(p, "Parameters"))
	}
	x := &FieldList{From: p.pos}
	p.expect(LPAREN)
	for p.tok != RPAREN {
		x.Fields = append(x.Fields, p.parseParameter())
		if p.tok != COMMA {
			break
		}
		p.next()
	}
	p.expect(RPAREN)
	return x
}

func (p *parser) parseParameter() *Field {
	if p.trace {
		defer un(trace(p, "Parameter"))
	}
	x := &Field{}

	switch p.tok {
	case IN:
		p.next()
	case OUT:
		p.next()
	case INOUT:
		p.next()
	}

	p.parseRestrictionSpec()
	if p.tok == MODIF {
		p.next()
	}
	x.Type = p.parseTypeRef()
	x.Name = p.parseExpr()

	return x
}

func (p *parser) parseTypeParameters() {
	if p.trace {
		defer un(trace(p, "TypeParameters"))
	}
	p.expect(LT)
	for p.tok != GT {
		p.parseTypeParameter()
		if p.tok != COMMA {
			break
		}
		p.next()
	}
	p.expect(GT)
}

func (p *parser) parseTypeParameter() {
	if p.trace {
		defer un(trace(p, "TypeParameter"))
	}
	if p.tok == IN {
		p.next()
	}

	switch p.tok {
	case TYPE:
		p.next()
	case SIGNATURE:
		p.next()
	default:
		p.parseTypeRef()
	}
	p.parseExpr()
}

/*************************************************************************
 * Statements
 *************************************************************************/

func (p *parser) parseBlockStmt() *BlockStmt {
	if p.trace {
		defer un(trace(p, "BlockStmt"))
	}

	x := &BlockStmt{LBrace: p.pos}
	p.expect(LBRACE)
	for p.tok != RBRACE && p.tok != EOF {
		x.Stmts = append(x.Stmts, p.parseStmt())
	}
	p.expect(RBRACE)
	return x
}

func (p *parser) parseStmt() Stmt {
	if p.trace {
		defer un(trace(p, "Stmt"))
	}

	switch p.tok {
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
		if p.tok != SEMICOLON && p.tok != RBRACE {
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
		if p.tok == LBRACE {
			p.parseBlockStmt()
			break
		}

		p.parseSimpleStmt()

		// call-statement block
		if p.tok == LBRACE {
			p.parseBlockStmt()
		}
	}
	p.expectSemi()
	return nil
}

func (p *parser) parseForLoop() {
	p.next()
	p.expect(LPAREN)
	if p.tok == VAR {
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
	p.parseSetExpr()
	p.parseBlockStmt()
}

func (p *parser) parseDoWhileLoop() {
	p.next()
	p.parseBlockStmt()
	p.expect(WHILE)
	p.parseSetExpr()
}

func (p *parser) parseIfStmt() {
	p.next()
	p.parseSetExpr()
	p.parseBlockStmt()
	if p.tok == ELSE {
		p.next()
		if p.tok == IF {
			p.parseIfStmt()
		} else {
			p.parseBlockStmt()
		}
	}
}

func (p *parser) parseSelect() {
	p.expect(SELECT)
	if p.tok == UNION {
		p.next()
	}
	p.parseSetExpr()
	p.expect(LBRACE)
	for p.tok == CASE {
		p.parseCaseStmt()
	}
	p.expect(RBRACE)
}

func (p *parser) parseCaseStmt() {
	p.expect(CASE)
	if p.tok == ELSE {
		p.next()
	} else {
		p.parseSetExpr()
	}
	p.parseBlockStmt()
}

func (p *parser) parseAltGuard() {
	p.next()
	if p.tok == ELSE {
		p.next()
		p.expect(RBRACK)
		p.parseBlockStmt()
		return
	}

	if p.tok != RBRACK {
		p.parseExpr()
	}
	p.expect(RBRACK)
	p.parseSimpleStmt()
	if p.tok == LBRACE {
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
