package parser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/scanner"
	"github.com/nokia/ntt/ttcn3/token"
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
// the corresponding ast.Module node. The source code may be provided via
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
// errors were found, the result is a partial AST (with ast.Bad* nodes
// representing the fragments of erroneous source code). Multiple errors
// are returned via a scanner.ErrorList which is sorted by file position.
//
func ParseModule(fset *token.FileSet, filename string, src interface{}, mode Mode, eh scanner.ErrorHandler) (f *ast.Module, err error) {
	if fset == nil {
		panic("parser.ParseModule: no token.FileSet provided (fset == nil)")
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
			// *ast.Module
			f = &ast.Module{
				Name:  new(ast.Ident),
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
	file    *token.File
	errors  scanner.ErrorList
	scanner scanner.Scanner

	// Tracing/debugging
	mode   Mode // parsing mode
	trace  bool // == (mode & Trace != 0)
	indent int  // indentation used for tracing output

	// Comments
	comments    []*ast.CommentGroup
	leadComment *ast.CommentGroup // last lead comment
	lineComment *ast.CommentGroup // last line comment

	// Next token
	pos token.Pos   // token position
	tok token.Token // one token look-ahead
	lit string      // token literal

	// Semicolon helper
	seenBrace bool

	// Error recovery
	// (used to limit the number of calls to parser.advance
	// w/o making scanning progress - avoids potential endless
	// loops across multiple parser functions during error recovery)
	syncPos token.Pos // last synchronization position
	syncCnt int       // number of parser.advance calls without progress

	// Non-syntactic parser control
	exprLev int  // < 0: in control clause, >= 0: in expression
	inRhs   bool // if set, the parser is parsing a rhs expression

	// Ordinary identifier scopes
	modScope   *ast.Scope   // modScope.Outer == nil
	topScope   *ast.Scope   // top-most scope; may be modScope
	unresolved []*ast.Ident // unresolved identifiers
	//imports    []*ast.ImportSpec // list of imports

	// Label scopes
	// (maintained by open/close LabelScope)
	labelScope  *ast.Scope     // label scope for current function
	targetStack [][]*ast.Ident // stack of unresolved labels
}

func (p *parser) init(fset *token.FileSet, filename string, src []byte, mode Mode, eh scanner.ErrorHandler) {
	p.file = fset.AddFile(filename, -1, len(src))
	var m scanner.Mode
	if mode&ParseComments != 0 {
		m = scanner.ScanComments
	}

	eh2 := func(pos token.Position, msg string) {
		if eh != nil {
			eh(pos, msg)
		}
		p.errors.Add(pos, msg)
	}
	p.scanner.Init(p.file, src, eh2, m)

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

// Advance to the next token.
func (p *parser) next0() {
	// Because of one-token look-ahead, print the previous token
	// when tracing as it provides a more readable output. The
	// very first token (!p.pos.IsValid()) is not initialized
	// (it is token.ILLEGAL), so don't print it .
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
func (p *parser) consumeComment() (comment *ast.Comment, endline int) {
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

	comment = &ast.Comment{Slash: p.pos, Text: p.lit}
	p.next0()

	return
}

// Consume a group of adjacent comments, add it to the parser's
// comments list, and return it together with the line at which
// the last comment in the group ends. A non-comment token or n
// empty lines terminate a comment group.
//
func (p *parser) consumeCommentGroup(n int) (comments *ast.CommentGroup, endline int) {
	var list []*ast.Comment
	endline = p.file.Line(p.pos)
	for p.tok == token.COMMENT && p.file.Line(p.pos) <= endline+n {
		var comment *ast.Comment
		comment, endline = p.consumeComment()
		list = append(list, comment)
	}

	// add comment group to the comments list
	comments = &ast.CommentGroup{List: list}
	p.comments = append(p.comments, comments)

	return
}

// Advance to the next non-comment token. In the process, collect
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

	if p.tok == token.RBRACE {
		p.seenBrace = true
	}

	p.next0()

	if p.tok == token.COMMENT {
		var comment *ast.CommentGroup
		var endline int

		if p.file.Line(p.pos) == p.file.Line(prev) {
			// The comment is on same line as the previous token; it
			// cannot be a lead comment but may be a line comment.
			comment, endline = p.consumeCommentGroup(0)
			if p.file.Line(p.pos) != endline || p.tok == token.EOF {
				// The next token is on a different line, thus
				// the last comment group is a line comment.
				p.lineComment = comment
			}
		}

		// consume successor comments, if any
		endline = -1
		for p.tok == token.COMMENT {
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

func (p *parser) error(pos token.Pos, msg string) {
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

func (p *parser) errorExpected(pos token.Pos, msg string) {
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

func (p *parser) expect(tok token.Token) token.Pos {
	pos := p.pos
	if p.tok != tok {
		p.errorExpected(pos, "'"+tok.String()+"'")
	}
	p.next() // make progress
	return pos
}

func (p *parser) expectSemi() {
	switch p.tok {
	case token.SEMICOLON:
		p.next()
	case token.RBRACE, token.EOF:
		// semicolon is optional before a closing '}'
	default:
		if !p.seenBrace {
			p.errorExpected(p.pos, "';'")
			p.advance(stmtStart)
		}
	}
}

// advance consumes tokens until the current token p.tok
// is in the 'to' set, or token.EOF. For error recovery.
func (p *parser) advance(to map[token.Token]bool) {
	for ; p.tok != token.EOF; p.next() {
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

var stmtStart = map[token.Token]bool{
	token.CONST:     true,
	token.VAR:       true,
	token.MODULEPAR: true,
	token.FUNCTION:  true,
	token.TESTCASE:  true,
	token.ALTSTEP:   true,
}


/*************************************************************************
 * Expressions
 *************************************************************************/

func (p *parser) parseExprList() (list []ast.Expr) {
	list = append(list, p.parseExpr())
	for p.tok == token.COMMA {
		p.next()
		list = append(list, p.parseExpr())
	}
	return list
}

func (p *parser) parseExpr() ast.Expr {
	if p.trace {
		defer un(trace(p, "Expr"))
	}

	x := p.parseBinaryExpr(token.LowestPrec + 1)

	if p.tok == token.ASSIGN {
		p.next()
		p.parseExpr()
	}

	return x
}

func (p *parser) parseBinaryExpr(prec1 int) ast.Expr {
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

		x = &ast.BinaryExpr{X: x, Op: op, OpPos: pos, Y: y}
	}
}

func (p *parser) parseUnaryExpr() ast.Expr {
	switch p.tok {
	case token.SUB, token.ADD, token.NOT, token.NOT4B, token.EXCL:
		op, pos := p.tok, p.pos
		p.next()
		// handle unused expr '-'
		if op == token.SUB && (p.tok == token.COMMA || p.tok == token.SEMICOLON || p.tok == token.RBRACE || p.tok == token.RBRACK || p.tok == token.RPAREN || p.tok == token.EOF) {
			return nil
		}
		return &ast.UnaryExpr{Op: op, OpPos: pos, X: p.parseUnaryExpr()}

	case token.MODIFIES:
		p.next()
		p.parsePrimaryExpr()
		p.expect(token.ASSIGN)
		p.parseExpr()
		return nil
	}
	return p.parsePrimaryExpr()
}

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
		case token.COLON:
			p.next()
			p.parseExpr()
		default:
			break L
		}
	}

	if p.tok == token.LENGTH {
		p.next()
		p.expect(token.LPAREN)
		p.parseExpr()
		p.expect(token.RPAREN)
	}

	if p.tok == token.IFPRESENT {
		p.next()
	}

	if p.tok == token.TO || p.tok == token.FROM {
		p.next()
		p.parseExpr()
	}

	if p.tok == token.REDIR {
		p.parseRedirect()
	}

	if p.tok == token.VALUE {
		p.next()
		p.parseExpr()
	}

	if p.tok == token.PARAM {
		p.next()
		p.expect(token.LPAREN)
		p.parseExprList()
		p.expect(token.RPAREN)
	}

	if p.tok == token.ALIVE {
		p.next()
	}

	return x
}

func (p *parser) parseOperand() ast.Expr {
	switch p.tok {
	case token.ANYKW, token.ALL:
		k := p.tok
		p.next()
		switch p.tok {
		case token.COMPONENT, token.PORT, token.TIMER:
			p.next()
			return nil
		case token.FROM:
			p.next()
			p.parsePrimaryExpr()
			return nil
		}

		// Workaround for deprecated port-attribute 'all'
		if k == token.ALL {
			return nil
		}

		p.errorExpected(p.pos, "'component', 'port', 'timer' or 'from'")

	case token.UNIVERSAL:
		p.next()
		p.expect(token.CHARSTRING)
		id := &ast.Ident{NamePos: p.pos, Name: p.lit}
		return id
	case token.IDENT, token.TIMER, token.TESTCASE, token.SYSTEM, token.MTC, token.ADDRESS, token.NULL, token.OMIT,
		token.CHARSTRING,
		token.MAP, token.UNMAP:
		id := &ast.Ident{NamePos: p.pos, Name: p.lit}
		p.next()
		return id
	case token.LBRACK:
		p.parseIndexExpr(nil)
	case token.LBRACE:
		p.next()
		if p.tok != token.RBRACE {
			p.parseExprList()
		}
		p.expect(token.RBRACE)
		return nil
	case token.LPAREN:
		p.next()
		// can be template `x := (1,2,3)`, but also artihmetic expression: `1*(2+3)`
		set := &ast.SetExpr{List: p.parseExprList()}
		p.expect(token.RPAREN)
		return set
	case token.INT, token.FLOAT, token.STRING, token.BSTRING,
		token.ANY, token.MUL,
		token.TRUE, token.FALSE,
		token.PASS, token.FAIL, token.NONE, token.INCONC, token.ERROR,
		token.NAN:
		lit := &ast.ValueLiteral{Kind: p.tok, ValuePos: p.pos, Value: p.lit}
		p.next()
		return lit
	case token.REGEXP:
		p.next()
		if p.tok == token.MODIF {
			p.next()
		}
		p.expect(token.LPAREN)
		p.parseExprList()
		p.expect(token.RPAREN)
	case token.PATTERN:
		p.next()
		if p.tok == token.MODIF {
			p.next()
		}
		p.expect(token.STRING)
	case token.DECMATCH:
		p.next()
		if p.tok == token.LPAREN {
			p.next()
			p.parseExprList()
			p.expect(token.RPAREN)
		}
		p.parseExpr()
	case token.MODIF: // @decoded
		p.next()
		if p.tok == token.LPAREN {
			p.next()
			p.parseExprList()
			p.expect(token.RPAREN)
		}
		p.parseExpr()
	default:
		p.errorExpected(p.pos, "operand")
	}

	return nil
}

func (p *parser) parseSelectorExpr(x ast.Expr) ast.Expr {
	p.next()
	return &ast.SelectorExpr{X: x, Sel: p.parseIdent()}
}

func (p *parser) parseIndexExpr(x ast.Expr) ast.Expr {
	p.next()
	x = &ast.IndexExpr{X: x, Index: p.parseExpr()}
	p.expect(token.RBRACK)
	return x
}

func (p *parser) parseCallExpr(x ast.Expr) ast.Expr {
	p.next()

	switch p.tok {
	case token.FROM, token.TO:
		p.next()
		p.parseExpr()
		if p.tok == token.REDIR {
			p.parseRedirect()
		}
		p.expect(token.RPAREN)
		return nil
	case token.REDIR:
		p.parseRedirect()
		p.expect(token.RPAREN)
		return nil
	default:
		var list []ast.Expr
		if p.tok != token.RPAREN {
			list = p.parseExprList()
		}
		p.expect(token.RPAREN)
		return &ast.CallExpr{Fun: x, Args: list}
	}

}

func (p *parser) parseRedirect() ast.Expr {
	p.next()

	if p.tok == token.VALUE {
		p.next()
		p.parseExprList()
	}

	if p.tok == token.PARAM {
		p.next()
		p.parseExprList()
	}

	if p.tok == token.SENDER {
		p.next()
		p.parsePrimaryExpr()
	}

	if p.tok == token.MODIF {
		p.next()
		if p.tok == token.VALUE {
			p.next()
		}
		p.parsePrimaryExpr()
	}

	if p.tok == token.TIMESTAMP {
		p.next()
		p.parsePrimaryExpr()
	}

	return nil
}

func (p *parser) parseIdent() *ast.Ident {
	pos := p.pos
	name := "_"
	switch p.tok {
	case token.UNIVERSAL:
		p.next()
		p.expect(token.CHARSTRING)
		name = p.lit
	case token.IDENT, token.ADDRESS, token.ALIVE, token.CHARSTRING:
		name = p.lit
		p.next()
	default:
		p.expect(token.IDENT) // use expect() error handling
	}
	return &ast.Ident{NamePos: pos, Name: name}
}

func (p *parser) parseRefList() {
	for {
		p.parseTypeRef()
		if p.tok != token.COMMA {
			break
		}
		p.next()
	}
}

func (p *parser) parseTypeRef() ast.Expr {
	if p.trace {
		defer un(trace(p, "TypeRef"))
	}
	x := p.parsePrimaryExpr()
	return x
}

/*************************************************************************
 * Module
 *************************************************************************/

func (p *parser) parseModule() *ast.Module {
	if p.trace {
		defer un(trace(p, "Module"))
	}

	pos := p.expect(token.MODULE)
	name := p.parseIdent()

	if p.tok == token.LANGUAGE {
		p.parseLanguageSpec()
	}

	p.expect(token.LBRACE)

	var decls []ast.Decl
	for p.tok != token.RBRACE && p.tok != token.EOF {
		decls = append(decls, p.parseModuleDef())
	}
	p.expect(token.RBRACE)

	return &ast.Module{
		Module:   pos,
		Name:     name,
		Decls:    decls,
		Comments: p.comments,
	}
}

func (p *parser) parseLanguageSpec() {
	p.next()
	for {
		p.expect(token.STRING)
		if p.tok != token.COMMA {
			break
		}
		p.next()
	}
}

func (p *parser) parseModuleDef() ast.Decl {
	switch p.tok {
	case token.PRIVATE, token.PUBLIC:
		p.next()
	case token.FRIEND:
		p.next()
		if p.tok == token.MODULE {
			p.parseFriend()
			p.expectSemi()
			return nil
		}
	}

	switch p.tok {
	case token.IMPORT:
		p.parseImport()
	case token.GROUP:
		p.parseGroup()
	case token.FRIEND:
		p.next()
		p.parseFriend()
	case token.TYPE:
		p.parseType()
	case token.TEMPLATE:
		p.parseTemplateDecl()
	case token.MODULEPAR:
		p.parseModulePar()
	case token.VAR, token.CONST:
		p.parseValueDecl()
	case token.SIGNATURE:
		p.parseSignatureDecl()
	case token.FUNCTION, token.TESTCASE, token.ALTSTEP:
		p.parseFuncDecl()
	case token.CONTROL:
		p.next()
		p.parseBlockStmt()
	case token.EXTERNAL:
		p.next()
		switch p.tok {
		case token.FUNCTION:
			p.parseExtFuncDecl()
		case token.CONST:
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

func (p *parser) parseImport() ast.Decl {
	if p.trace {
		defer un(trace(p, "Import"))
	}

	pos := p.pos
	p.next()
	p.expect(token.FROM)

	name := p.parseIdent()

	if p.tok == token.LANGUAGE {
		p.parseLanguageSpec()
	}

	var specs []ast.ImportSpec
	switch p.tok {
	case token.ALL:
		p.next()
		if p.tok == token.EXCEPT {
			p.parseExceptSpec()
		}
	case token.LBRACE:
		p.parseImportSpec()
	default:
		p.errorExpected(p.pos, "'all' or import spec")
	}

	p.parseWith()

	return &ast.ImportDecl{
		ImportPos:   pos,
		Module:      name,
		ImportSpecs: specs,
	}
}

func (p *parser) parseImportSpec() {
	p.expect(token.LBRACE)
	for p.tok != token.RBRACE && p.tok != token.EOF {
		p.parseImportStmt()
	}
	p.expect(token.RBRACE)
}

func (p *parser) parseImportStmt() {
	switch p.tok {
	case token.ALTSTEP, token.CONST, token.FUNCTION, token.MODULEPAR,
		token.SIGNATURE, token.TEMPLATE, token.TESTCASE, token.TYPE:
		p.next()
		if p.tok == token.ALL {
			p.next()
			if p.tok == token.EXCEPT {
				p.next()
				p.parseRefList()
			}
		} else {
			p.parseRefList()
		}
	case token.GROUP:
		p.next()
		for {
			p.parseTypeRef()
			if p.tok == token.EXCEPT {
				p.parseExceptSpec()
			}
			if p.tok != token.COMMA {
				break
			}
			p.next()
		}
	case token.IMPORT:
		p.next()
		p.expect(token.ALL)
	default:
		p.errorExpected(p.pos, "definition qualifier")
	}

	p.expectSemi()
}

func (p *parser) parseExceptSpec() {
	p.next()
	p.expect(token.LBRACE)
	for p.tok != token.RBRACE && p.tok != token.EOF {
		p.parseExceptStmt()
	}
	p.expect(token.RBRACE)
}

func (p *parser) parseExceptStmt() {
	switch p.tok {
	case token.ALTSTEP, token.CONST, token.FUNCTION, token.GROUP,
		token.IMPORT, token.MODULEPAR, token.SIGNATURE, token.TEMPLATE,
		token.TESTCASE, token.TYPE:
		p.next()
	default:
		p.errorExpected(p.pos, "definition qualifier")
	}

	if p.tok == token.ALL {
		p.next()
	} else {
		for {
			p.parseTypeRef()
			if p.tok != token.COMMA {
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
	p.expect(token.LBRACE)

	var decls []ast.Decl
	for p.tok != token.RBRACE && p.tok != token.EOF {
		decls = append(decls, p.parseModuleDef())
	}
	p.expect(token.RBRACE)
	p.parseWith()
}

func (p *parser) parseFriend() {
	p.expect(token.MODULE)
	p.parseIdent()
	p.parseWith()
}


/*************************************************************************
 * With Attributes
 *************************************************************************/

func (p *parser) parseWith() ast.Node {
	if p.tok != token.WITH {
		return nil
	}

	p.expect(token.WITH)
	p.expect(token.LBRACE)
	for p.tok != token.RBRACE && p.tok != token.EOF {
		p.parseWithStmt()
	}
	p.expect(token.RBRACE)
	return nil
}

func (p *parser) parseWithStmt() ast.Node {
	if p.trace {
		defer un(trace(p, "WithStmt"))
	}
	switch p.tok {
	case token.ENCODE,
		token.VARIANT,
		token.DISPLAY,
		token.EXTENSION,
		token.OPTIONAL,
		token.STEPSIZE,
		token.OVERRIDE:
		p.next()
	default:
		p.errorExpected(p.pos, "with-attribute")
		p.next()
	}

	switch p.tok {
	case token.OVERRIDE:
		p.next()
	case token.MODIF:
		p.next() // consume '@local'
	}

	if p.tok == token.LPAREN {
		p.next()
		for {
			p.parseWithQualifier()
			if p.tok != token.COMMA {
				break
			}
			p.next()
		}
		p.expect(token.RPAREN)
	}

	p.expect(token.STRING)

	if p.tok == token.DOT {
		p.next()
		p.expect(token.STRING)
	}

	p.expectSemi()
	return nil
}

func (p *parser) parseWithQualifier() {
	switch p.tok {
	case token.IDENT:
		p.parseTypeRef()
	case token.LBRACK:
		p.parseIndexExpr(nil)
	case token.TYPE, token.TEMPLATE, token.CONST, token.ALTSTEP, token.TESTCASE, token.FUNCTION, token.SIGNATURE, token.MODULEPAR, token.GROUP:
		p.next()
		p.expect(token.ALL)
		if p.tok == token.EXCEPT {
			p.next()
			p.expect(token.LBRACE)
			p.parseRefList()
			p.expect(token.RBRACE)
		}
	default:
		p.errorExpected(p.pos, "with-qualifier")
	}
}


/*************************************************************************
 * Type Definitions
 *************************************************************************/

func (p *parser) parseType() ast.Decl {
	if p.trace {
		defer un(trace(p, "Type"))
	}
	p.next()
	switch p.tok {
	case token.IDENT, token.UNIVERSAL, token.CHARSTRING, token.ADDRESS:
		p.parseSubType()
	case token.UNION:
		p.next()
		p.parseStructType()
	case token.SET, token.RECORD:
		p.next()
		if p.tok == token.IDENT {
			p.parseStructType()
			break
		}
		p.parseListType()
	case token.ENUMERATED:
		p.parseEnumType()
	case token.PORT:
		p.parsePortType()
	case token.COMPONENT:
		p.parseComponentType()
	case token.FUNCTION, token.ALTSTEP, token.TESTCASE:
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
	case token.IDENT, token.ADDRESS, token.NULL, token.CHARSTRING, token.UNIVERSAL:
		p.parseTypeRef()
	case token.UNION:
		p.next()
		p.parseNestedStructType()
	case token.SET, token.RECORD:
		p.next()
		if p.tok == token.LBRACE {
			p.parseNestedStructType()
			break
		}
		p.parseNestedListType()
	case token.ENUMERATED:
		p.parseNestedEnumType()
	default:
		p.errorExpected(p.pos, "type definition")
	}
}


/*************************************************************************
 * Struct Types
 *************************************************************************/

func (p *parser) parseNestedStructType() {
	if p.trace {
		defer un(trace(p, "NestedStructType"))
	}
	p.expect(token.LBRACE)
	for p.tok != token.RBRACE && p.tok != token.EOF {
		p.parseStructField()
		if p.tok != token.COMMA {
			break
		}
		p.next()
	}
	p.expect(token.RBRACE)
}

func (p *parser) parseStructType() {
	if p.trace {
		defer un(trace(p, "StructType"))
	}
	p.parseIdent()
	p.expect(token.LBRACE)
	for p.tok != token.RBRACE && p.tok != token.EOF {
		p.parseStructField()
		if p.tok != token.COMMA {
			break
		}
		p.next()
	}
	p.expect(token.RBRACE)
	p.parseWith()
}

func (p *parser) parseStructField() {
	if p.trace {
		defer un(trace(p, "StructField"))
	}
	if p.tok == token.MODIF {
		p.next() // @default
	}
	p.parseNestedType()
	p.parsePrimaryExpr()
	p.parseConstraint()

	if p.tok == token.OPTIONAL {
		p.next()
	}
}


/*************************************************************************
 * List Type
 *************************************************************************/

func (p *parser) parseNestedListType() {
	if p.trace {
		defer un(trace(p, "NestedListType"))
	}
	p.parseLength()
	p.expect(token.OF)
	p.parseNestedType()
}

func (p *parser) parseListType() {
	if p.trace {
		defer un(trace(p, "ListType"))
	}
	p.parseLength()

	p.expect(token.OF)
	p.parseNestedType()
	p.parsePrimaryExpr()
	p.parseConstraint()
	p.parseWith()
}



/*************************************************************************
 * Enumeration Type
 *************************************************************************/

func (p *parser) parseNestedEnumType() {
	if p.trace {
		defer un(trace(p, "NestedEnumType"))
	}
	p.next()
	p.expect(token.LBRACE)
	for p.tok != token.RBRACE && p.tok != token.EOF {
		p.parseExpr()
		if p.tok != token.COMMA {
			break
		}
		p.next()
	}
	p.expect(token.RBRACE)
}

func (p *parser) parseEnumType() {
	if p.trace {
		defer un(trace(p, "EnumType"))
	}
	p.next()
	p.parseIdent()
	p.expect(token.LBRACE)
	for p.tok != token.RBRACE && p.tok != token.EOF {
		p.parseExpr()
		if p.tok != token.COMMA {
			break
		}
		p.next()
	}
	p.expect(token.RBRACE)
	p.parseWith()
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
	switch p.tok {
	case token.MIXED, token.MESSAGE, token.PROCEDURE:
		p.next()
	default:
		p.errorExpected(p.pos, "'message' or 'procedure'")
	}

	p.expect(token.LBRACE)
	for p.tok != token.RBRACE && p.tok != token.EOF {
		p.parsePortAttribute()
		p.expectSemi()
	}
	p.expect(token.RBRACE)
	p.parseWith()
}

func (p *parser) parsePortAttribute() {
	if p.trace {
		defer un(trace(p, "PortAttribute"))
	}
	switch p.tok {
	case token.IN, token.OUT, token.INOUT:
		p.next()
		p.parseRefList()
	case token.ADDRESS:
		p.next()
		p.parseRefList()
	case token.MAP, token.UNMAP:
		p.next()
		p.expect(token.PARAM)
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
	if p.tok == token.EXTENDS {
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
	if p.tok == token.RUNS {
		p.next()
		p.expect(token.ON)
		p.parseTypeRef()
	}
	if p.tok == token.SYSTEM {
		p.next()
		p.parseTypeRef()
	}
	if p.tok == token.RETURN {
		p.parseReturn()
	}
	p.parseWith()

}


/*************************************************************************
 * Subtype
 *************************************************************************/

func (p *parser) parseSubType() *ast.SubType {
	if p.trace {
		defer un(trace(p, "SubType"))
	}

	p.parseNestedType()
	p.parsePrimaryExpr()
	p.parseConstraint()

	p.parseWith()
	return nil
}

func (p *parser) parseConstraint() {
	// TODO(mef) fix constraints consumed by previous PrimaryExpr

	if p.tok == token.LPAREN {
		p.next()
		p.parseExprList()
		p.expect(token.RPAREN)
	}

	p.parseLength()
}

func (p *parser) parseLength() {
	if p.tok == token.LENGTH {
		p.next()
		p.expect(token.LPAREN)
		p.parseExpr()
		p.expect(token.RPAREN)
	}
}


/*************************************************************************
 * Declarations
 *************************************************************************/

func (p *parser) parseDecl() ast.Decl {
	switch p.tok {
	case token.TEMPLATE:
		return p.parseTemplateDecl()
	case token.MODULEPAR:
		return p.parseModulePar()
	case token.VAR, token.CONST, token.TIMER, token.PORT:
		return p.parseValueDecl()
	case token.FUNCTION, token.TESTCASE, token.ALTSTEP:
		return p.parseFuncDecl()
	case token.EXTERNAL:
		p.next()
		return p.parseExtFuncDecl()
	case token.SIGNATURE:
		return p.parseSignatureDecl()
	default:
		p.errorExpected(p.pos, "declaration")
	}
	return nil
}


/*************************************************************************
 * Template Declaration
 *************************************************************************/

func (p *parser) parseTemplateDecl() *ast.ValueDecl {
	if p.trace {
		defer un(trace(p, "TemplateDecl"))
	}

	x := &ast.ValueDecl{DeclPos: p.pos, Kind: p.tok}
	p.next()

	if p.tok == token.LPAREN {
		p.next() // consume '('
		p.next() // consume omit/value/...
		p.expect(token.RPAREN)
	}

	if p.tok == token.MODIF {
		p.next()
	}

	x.Type = p.parseTypeRef()
	p.parseIdent()
	if p.tok == token.LPAREN {
		p.parseParameters()
	}
	if p.tok == token.MODIFIES {
		p.next()
		p.parseIdent()
	}
	p.expect(token.ASSIGN)
	p.parseExpr()

	p.parseWith()
	return x
}


/*************************************************************************
 * Module Parameter
 *************************************************************************/

func (p *parser) parseModulePar() *ast.ValueDecl {
	if p.trace {
		defer un(trace(p, "ModulePar"))
	}

	x := &ast.ValueDecl{DeclPos: p.pos, Kind: p.tok}
	p.next()

	if p.tok == token.LBRACE {
		p.next()
		for p.tok != token.RBRACE && p.tok != token.EOF {
			p.parseRestrictionSpec()
			p.parseTypeRef()
			p.parseExprList()
			p.expectSemi()
		}
		p.expect(token.RBRACE)
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

func (p *parser) parseValueDecl() *ast.ValueDecl {
	if p.trace {
		defer un(trace(p, "ValueDecl"))
	}

	x := &ast.ValueDecl{DeclPos: p.pos, Kind: p.tok}
	p.next()
	p.parseRestrictionSpec()

	if p.tok == token.MODIF {
		p.next()
	}

	if x.Kind != token.TIMER {
		x.Type = p.parseTypeRef()
	}
	x.Decls = p.parseExprList()
	p.parseWith()
	return x
}

func (p *parser) parseRestrictionSpec() *ast.RestrictionSpec {
	switch p.tok {
	case token.TEMPLATE:
		x := &ast.RestrictionSpec{Kind: p.tok, KindPos: p.pos}
		p.next()
		if p.tok != token.LPAREN {
			return x
		}

		p.next()
		x.Kind = p.tok
		x.KindPos = p.pos
		p.next()
		p.expect(token.RPAREN)

	case token.OMIT, token.VALUE, token.PRESENT:
		x := &ast.RestrictionSpec{Kind: p.tok, KindPos: p.pos}
		p.next()
		return x
	}
	return nil
}


/*************************************************************************
 * Behaviour Declaration
 *************************************************************************/

func (p *parser) parseFuncDecl() *ast.FuncDecl {
	if p.trace {
		defer un(trace(p, "FuncDecl"))
	}

	x := &ast.FuncDecl{FuncPos: p.pos, Kind: p.tok}
	p.next()
	x.Name = p.parseIdent()

	if p.tok == token.MODIF {
		p.next()
	}

	x.Params = p.parseParameters()
	if p.tok == token.RUNS {
		p.next()
		p.expect(token.ON)
		x.RunsOn = p.parseTypeRef()
	}
	if p.tok == token.MTC {
		p.next()
		x.Mtc = p.parseTypeRef()
	}
	if p.tok == token.SYSTEM {
		p.next()
		x.System = p.parseTypeRef()
	}
	if p.tok == token.RETURN {
		x.Return = p.parseReturn()
	}
	if p.tok == token.LBRACE {
		x.Body = p.parseBlockStmt()
	}

	p.parseWith()
	return x
}


/*************************************************************************
 * External Function Declaration
 *************************************************************************/

func (p *parser) parseExtFuncDecl() *ast.FuncDecl {
	if p.trace {
		defer un(trace(p, "ExtFuncDecl"))
	}

	x := &ast.FuncDecl{FuncPos: p.pos, Kind: p.tok}
	p.next()
	x.Name = p.parseIdent()

	if p.tok == token.MODIF {
		p.next()
	}

	x.Params = p.parseParameters()
	if p.tok == token.RUNS {
		p.next()
		p.expect(token.ON)
		x.RunsOn = p.parseTypeRef()
	}
	if p.tok == token.MTC {
		p.next()
		x.Mtc = p.parseTypeRef()
	}
	if p.tok == token.SYSTEM {
		p.next()
		x.System = p.parseTypeRef()
	}
	if p.tok == token.RETURN {
		x.Return = p.parseReturn()
	}
	p.parseWith()
	return x
}


/*************************************************************************
 * Signature Declaration
 *************************************************************************/

func (p *parser) parseSignatureDecl() ast.Decl {
	if p.trace {
		defer un(trace(p, "SignatureDecl"))
	}

	p.next()
	p.parseIdent()

	p.parseParameters()

	if p.tok == token.NOBLOCK {
		p.next()
	}

	if p.tok == token.RETURN {
		p.parseReturn()
	}

	if p.tok == token.EXCEPTION {
		p.next()
		p.expect(token.LPAREN)
		p.parseRefList()
		p.expect(token.RPAREN)
	}
	p.parseWith()
	return nil
}

func (p *parser) parseReturn() ast.Expr {
	p.next()
	p.parseRestrictionSpec()
	if p.tok == token.MODIF {
		p.next()
	}
	return p.parseTypeRef()
}

func (p *parser) parseParameters() *ast.FieldList {
	x := &ast.FieldList{From: p.pos}
	p.expect(token.LPAREN)
	for p.tok != token.RPAREN {
		x.Fields = append(x.Fields, p.parseParameter())
		if p.tok != token.COMMA {
			break
		}
		p.next()
	}
	p.expect(token.RPAREN)
	return x
}

func (p *parser) parseParameter() *ast.Field {
	x := &ast.Field{}

	switch p.tok {
	case token.IN:
		p.next()
	case token.OUT:
		p.next()
	case token.INOUT:
		p.next()
	}

	p.parseRestrictionSpec()
	if p.tok == token.MODIF {
		p.next()
	}
	x.Type = p.parseTypeRef()
	x.Name = p.parseExpr()

	return x
}


/*************************************************************************
 * Statements
 *************************************************************************/

func (p *parser) parseBlockStmt() *ast.BlockStmt {
	if p.trace {
		defer un(trace(p, "BlockStmt"))
	}

	x := &ast.BlockStmt{LBrace: p.pos}
	p.expect(token.LBRACE)
	for p.tok != token.RBRACE && p.tok != token.EOF {
		x.Stmts = append(x.Stmts, p.parseStmt())
	}
	p.expect(token.RBRACE)
	return x
}

func (p *parser) parseStmt() ast.Stmt {
	if p.trace {
		defer un(trace(p, "Stmt"))
	}

	switch p.tok {
	case token.TEMPLATE, token.VAR, token.CONST, token.TIMER, token.PORT:
		p.parseDecl()
	case token.REPEAT, token.BREAK, token.CONTINUE:
		p.next()
	case token.LABEL:
		p.next()
		p.expect(token.IDENT)
	case token.GOTO:
		p.next()
		p.expect(token.IDENT)
	case token.RETURN:
		p.next()
		if p.tok != token.SEMICOLON && p.tok != token.RBRACE {
			p.parseExpr()
		}
	case token.SELECT:
		p.parseSelect()
	case token.ALT, token.INTERLEAVE:
		p.next()
		p.parseBlockStmt()
	case token.LBRACK:
		p.parseAltGuard()
	case token.FOR:
		p.parseForLoop()
	case token.WHILE:
		p.parseWhileLoop()
	case token.DO:
		p.parseDoWhileLoop()
	case token.IF:
		p.parseIfStmt()
	default:
		p.parseSimpleStmt()
	}
	p.expectSemi()
	return nil
}

func (p *parser) parseForLoop() {
	p.next()
	p.expect(token.LPAREN)
	if p.tok == token.VAR {
		p.parseValueDecl()
	} else {
		p.parseExpr()
	}
	p.expect(token.SEMICOLON)
	p.parseExpr()
	p.expect(token.SEMICOLON)
	p.parseExpr()
	p.expect(token.RPAREN)
	p.parseBlockStmt()
}

func (p *parser) parseWhileLoop() {
	p.next()
	p.expect(token.LPAREN)
	p.parseExpr()
	p.expect(token.RPAREN)
	p.parseBlockStmt()
}

func (p *parser) parseDoWhileLoop() {
	p.next()
	p.parseBlockStmt()
	p.expect(token.WHILE)
	p.expect(token.LPAREN)
	p.parseExpr()
	p.expect(token.RPAREN)
}

func (p *parser) parseIfStmt() {
	p.next()
	p.expect(token.LPAREN)
	p.parseExpr()
	p.expect(token.RPAREN)
	p.parseBlockStmt()
	if p.tok == token.ELSE {
		p.next()
		if p.tok == token.IF {
			p.parseIfStmt()
		} else {
			p.parseBlockStmt()
		}
	}
}

func (p *parser) parseSelect() {
	p.next()
	if p.tok == token.UNION {
		p.next()
	}

	p.expect(token.LPAREN)
	p.parseExpr()
	p.expect(token.RPAREN)
	p.expect(token.LBRACE)
	for p.tok == token.CASE {
		p.parseCaseStmt()
	}
	p.expect(token.RBRACE)
}

func (p *parser) parseCaseStmt() {
	p.expect(token.CASE)
	if p.tok == token.ELSE {
		p.next()
	} else {
		p.expect(token.LPAREN)
		p.parseExprList()
		p.expect(token.RPAREN)
	}
	p.parseBlockStmt()
}

func (p *parser) parseAltGuard() {
	p.next()
	if p.tok == token.ELSE {
		p.next()
		p.expect(token.RBRACK)
		p.parseBlockStmt()
		return
	}

	if p.tok != token.RBRACK {
		p.parseExpr()
	}
	p.expect(token.RBRACK)
	p.parseSimpleStmt()
	if p.tok == token.LBRACE {
		p.parseBlockStmt()
	}
}

func (p *parser) parseSimpleStmt() ast.Stmt {
	if p.trace {
		defer un(trace(p, "SimpleStmt"))
	}

	p.parseExpr()

	// for call-statement, note conflict in alt-context.
	if p.tok == token.LBRACE {
		p.parseBlockStmt()
	}

	return nil
}
