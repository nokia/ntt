// Package ast provides TTCN-3 syntax tree nodes and functions for tree
// traversal.
package ast

import (
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/token"
)

// All node types implement the Node interface.
type Node interface {
	Pos() loc.Pos
	End() loc.Pos
	LastTok() *Token
}

// All expression nodes implement the Expr interface.
type Expr interface {
	Node
	exprNode()
}

// All statement nodes implement the Stmt interface.
type Stmt interface {
	Node
	stmtNode()
}

// All declaration nodes implement the Decl interface.
type Decl interface {
	Node
	declNode()
}

type ErrorNode struct {
	From, To Token
}

func (x ErrorNode) Pos() loc.Pos    { return x.From.Pos() }
func (x ErrorNode) End() loc.Pos    { return x.To.End() }
func (x ErrorNode) LastTok() *Token { return x.To.LastTok() }
func (x ErrorNode) exprNode()       {}
func (x ErrorNode) stmtNode()       {}
func (x ErrorNode) declNode()       {}
func (x ErrorNode) typeSpecNode()   {}

// token functionality shared by Token and Trivia
type Terminal struct {
	pos  loc.Pos
	Kind token.Kind // Token kind like TESTCASE, SEMICOLON, COMMENT, ...
	Lit  string     // Token values for non-operator tokens
}

func (x Terminal) Pos() loc.Pos { return x.pos }
func (x Terminal) End() loc.Pos {
	return loc.Pos(int(x.pos) + len(x.String()))
}

func (x *Terminal) String() string {
	if x.Kind.IsLiteral() {
		return x.Lit
	}
	return x.Kind.String()
}

func (t *Terminal) IsValid() bool {
	return t.pos.IsValid()
}

// A Token represents a TTCN-3 token and implements the Node interface. Tokens are leave-nodes.
type Token struct {
	Terminal
	LeadingTriv  []Trivia
	TrailingTriv []Trivia
}

func NewToken(pos loc.Pos, kind token.Kind, val string) Token {
	return Token{Terminal{pos, kind, val}, nil, nil}
}

// Comments concatenates all leading comments into one single string.
func (t *Token) Comments() string {
	return joinComments(t.LeadingTriv)
}

func (t *Token) LastTok() *Token { return t }

// Trivia represent the parts of the source text that are largely insignificant
// for normal understanding of the code, such as whitespace, comments, and
// preprocessor directives.
type Trivia struct {
	Terminal
}

func NewTrivia(pos loc.Pos, kind token.Kind, val string) Trivia {
	return Trivia{Terminal{pos, kind, val}}
}

type (
	// Ident represents an identifier.
	Ident struct {
		Tok  Token // first identifier token
		Tok2 Token // optional second identifier token, e.g. for "any port"
	}

	// ParametrizedIdent represents a paremetrized identifier, e.g. "f<charstring>".
	ParametrizedIdent struct {
		Ident  *Ident     // Identifier
		Params *ParenExpr // Parameter list
	}

	// A ValueLiteral represents simple literals, like integers, charstrings, ...
	ValueLiteral struct {
		Tok Token
	}

	// A CompositeLiteral represents composite literals, e.g. "{x:=23, y:=5}".
	CompositeLiteral struct {
		LBrace Token  // Position of "{"
		List   []Expr // Expression list
		RBrace Token  // Position of "{"
	}

	// A UnaryExpr represents a unary expresions.
	UnaryExpr struct {
		Op Token // Operator token, like "+", "-", "!", ...
		X  Expr
	}

	// A BinaryExpr represents a binary expression.
	// Possible operands are all tokens with a precedence value, TO and FROM.
	BinaryExpr struct {
		X  Expr  // First operand
		Op Token // Operator token
		Y  Expr  // Second operand
	}

	// A ParenExpr represents parenthized expression lists.
	ParenExpr struct {
		LParen Token  // Position of "(", "<"
		List   []Expr // Expression list
		RParen Token  // Position of ")", ">"
	}

	// A SelectorExpr represents an expression followed by a selector.
	SelectorExpr struct {
		X   Expr  // Preceding expression (might be nil)
		Dot Token // Position of "."
		Sel Expr  // Literal, identifier or reference.
	}

	// A IndexExpr represents an expression followed by an index.
	IndexExpr struct {
		X      Expr  // Preceding expression (might be nil)
		LBrack Token // Position of "["
		Index  Expr  // Actuall index expression (might be "-")
		RBrack Token // Position of "]"
	}

	// A CallExpr represents a regular function call.
	CallExpr struct {
		Fun  Expr       // Function expression
		Args *ParenExpr // Function arguments
	}

	// A LengthExpr represents a length expression.
	LengthExpr struct {
		X    Expr       // Preceding expression
		Len  Token      // Position of "length" keyword
		Size *ParenExpr // Size expression
	}

	// A RedirectExpr represents various redirect expressions
	RedirectExpr struct {
		X             Expr   // Preceding redirected expression
		Tok           Token  // Position of "->"
		ValueTok      Token  // Position of "value" or nil
		Value         []Expr // Value expression
		ParamTok      Token  // Position of "param" or nil
		Param         []Expr // Param expression
		SenderTok     Token  // Position of "sender" or nil
		Sender        Expr   // Sender expression
		IndexTok      Token  // Position of "@index" or nil
		IndexValueTok Token  // Position of "value" or nil
		Index         Expr   // Index expression
		TimestampTok  Token  // Position of "timestamp" or nil
		Timestamp     Expr   // Timestamp expression
	}

	// A ValueExpr represents the return value used by signature based communication.
	ValueExpr struct {
		X   Expr  // Preceding template expression
		Tok Token // Position of "value"
		Y   Expr  // Value expression
	}

	// A ParamExpr represents parametrized map and unmap statements.
	ParamExpr struct {
		X   Expr  // map or unmap statement
		Tok Token // Position "param"
		Y   Expr  // Additional arguments for map/unmap
	}

	// A FromExpr represents a "from" expression, like "any from a".
	FromExpr struct {
		Kind    Token // ANY or ALL
		FromTok Token // Position of "from"
		X       Expr  // Expression
	}

	// A ModifiesExpr represents a "modifies" expression.
	ModifiesExpr struct {
		Tok    Token // Position of "modifies"
		X      Expr  // Base template expression
		Assign Token // Position of ":="
		Y      Expr  // Modifying expression
	}

	// A RegexExpr represents a "regexp" expression.
	RegexpExpr struct {
		Tok    Token // Position of "regexp"
		NoCase Token // Position of "@nocase" or nil
		X      Expr  // Regex expression
	}

	// A PatternExpr represents a "pattern" expression.
	PatternExpr struct {
		Tok    Token // Position of "pattern"
		NoCase Token // Position of "@nocase" of nil
		X      Expr  // Pattern expression
	}

	// A DecmatchExpr represents a "decmatch" expression.
	DecmatchExpr struct {
		Tok    Token // Position of "decmatch"
		Params Expr  // Parameter list or nil
		X      Expr  // Template expression
	}

	// A DecodedExpr represents a "@decoded" expression.
	DecodedExpr struct {
		Tok    Token // Position of "decoded"
		Params Expr  // Parameter list or nil
		X      Expr  // Template expression
	}

	// A DefKindExpr represents a definition kind expression, used by imports
	// and with-attributes.
	DefKindExpr struct {
		Kind Token  // Definition kind, "type", "group", ...
		List []Expr // List of identifiers or except-expressions
	}

	// A ExceptExpr is used by DefKindExpr to express exlusion of specific
	// defintions.
	ExceptExpr struct {
		X         Expr   // (Qualified) identifier or "all"
		ExceptTok Token  // Position of "except"
		LBrace    Token  // Position of "{" or nil
		List      []Expr // List of identifiers or DefKindExprs to exclude
		RBrace    Token  // Position of "}" or nil
	}
)

func (x *Ident) String() string {
	if x.Tok2.IsValid() {
		return x.Tok.String() + " " + x.Tok2.String()
	}
	return x.Tok.String()
}

func (x *Ident) LastTok() *Token {
	if x.Tok2.LastTok().IsValid() {
		return x.Tok2.LastTok()
	}
	return x.Tok.LastTok()
}

func (x *ParametrizedIdent) LastTok() *Token { return x.Params.LastTok() }
func (x *ValueLiteral) LastTok() *Token      { return x.Tok.LastTok() }
func (x *CompositeLiteral) LastTok() *Token  { return x.RBrace.LastTok() }
func (x *UnaryExpr) LastTok() *Token         { return x.X.LastTok() }
func (x *BinaryExpr) LastTok() *Token        { return x.Y.LastTok() }
func (x *ParenExpr) LastTok() *Token         { return x.RParen.LastTok() }
func (x *SelectorExpr) LastTok() *Token      { return x.Sel.LastTok() }
func (x *IndexExpr) LastTok() *Token         { return x.RBrack.LastTok() }
func (x *CallExpr) LastTok() *Token          { return x.Args.LastTok() }
func (x *LengthExpr) LastTok() *Token        { return x.Size.LastTok() }
func (x *RedirectExpr) LastTok() *Token {
	if x.Timestamp != nil {
		return x.Timestamp.LastTok()
	}
	if x.Index != nil {
		return x.Index.LastTok()
	}
	if x.Sender != nil {
		return x.Sender.LastTok()
	}
	if x.Param != nil {
		return x.Param[len(x.Param)-1].LastTok()
	}
	if x.Value != nil {
		return x.Value[len(x.Value)-1].LastTok()
	}
	return x.Tok.LastTok()
}

func (x *ValueExpr) LastTok() *Token    { return x.Y.LastTok() }
func (x *ParamExpr) LastTok() *Token    { return x.Y.LastTok() }
func (x *FromExpr) LastTok() *Token     { return x.X.LastTok() }
func (x *ModifiesExpr) LastTok() *Token { return x.Y.LastTok() }
func (x *RegexpExpr) LastTok() *Token   { return x.X.LastTok() }
func (x *PatternExpr) LastTok() *Token  { return x.X.LastTok() }
func (x *DecmatchExpr) LastTok() *Token { return x.X.LastTok() }
func (x *DecodedExpr) LastTok() *Token  { return x.X.LastTok() }

func (x *DefKindExpr) LastTok() *Token {
	return x.List[len(x.List)-1].LastTok()
}

func (x *ExceptExpr) LastTok() *Token {
	if x.RBrace.LastTok().IsValid() {
		return x.RBrace.LastTok()
	}
	return x.List[len(x.List)-1].LastTok()
}

func (x *Ident) Pos() loc.Pos             { return x.Tok.Pos() }
func (x *ParametrizedIdent) Pos() loc.Pos { return x.Ident.Pos() }
func (x *ValueLiteral) Pos() loc.Pos      { return x.Tok.Pos() }
func (x *CompositeLiteral) Pos() loc.Pos  { return x.LBrace.Pos() }
func (x *UnaryExpr) Pos() loc.Pos         { return x.Op.Pos() }
func (x *BinaryExpr) Pos() loc.Pos        { return x.X.Pos() }
func (x *ParenExpr) Pos() loc.Pos         { return x.LParen.Pos() }
func (x *SelectorExpr) Pos() loc.Pos      { return x.X.Pos() }
func (x *IndexExpr) Pos() loc.Pos         { return x.X.Pos() }
func (x *CallExpr) Pos() loc.Pos          { return x.Fun.Pos() }
func (x *LengthExpr) Pos() loc.Pos        { return x.X.Pos() }
func (x *RedirectExpr) Pos() loc.Pos      { return x.X.Pos() }
func (x *ValueExpr) Pos() loc.Pos         { return x.X.Pos() }
func (x *ParamExpr) Pos() loc.Pos         { return x.X.Pos() }
func (x *FromExpr) Pos() loc.Pos          { return x.Kind.Pos() }
func (x *ModifiesExpr) Pos() loc.Pos      { return x.Tok.Pos() }
func (x *RegexpExpr) Pos() loc.Pos        { return x.Tok.Pos() }
func (x *PatternExpr) Pos() loc.Pos       { return x.Tok.Pos() }
func (x *DecmatchExpr) Pos() loc.Pos      { return x.Tok.Pos() }
func (x *DecodedExpr) Pos() loc.Pos       { return x.Tok.Pos() }
func (x *DefKindExpr) Pos() loc.Pos       { return x.Kind.Pos() }
func (x *ExceptExpr) Pos() loc.Pos        { return x.X.Pos() }

func (x *Ident) End() loc.Pos             { return x.LastTok().End() }
func (x *ParametrizedIdent) End() loc.Pos { return x.LastTok().End() }
func (x *ValueLiteral) End() loc.Pos      { return x.LastTok().End() }
func (x *CompositeLiteral) End() loc.Pos  { return x.LastTok().End() }
func (x *UnaryExpr) End() loc.Pos         { return x.LastTok().End() }
func (x *BinaryExpr) End() loc.Pos        { return x.LastTok().End() }
func (x *ParenExpr) End() loc.Pos         { return x.LastTok().End() }
func (x *SelectorExpr) End() loc.Pos      { return x.LastTok().End() }
func (x *IndexExpr) End() loc.Pos         { return x.LastTok().End() }
func (x *CallExpr) End() loc.Pos          { return x.LastTok().End() }
func (x *LengthExpr) End() loc.Pos        { return x.LastTok().End() }
func (x *RedirectExpr) End() loc.Pos      { return x.LastTok().End() }
func (x *ValueExpr) End() loc.Pos         { return x.LastTok().End() }
func (x *ParamExpr) End() loc.Pos         { return x.LastTok().End() }
func (x *FromExpr) End() loc.Pos          { return x.LastTok().End() }
func (x *ModifiesExpr) End() loc.Pos      { return x.LastTok().End() }
func (x *RegexpExpr) End() loc.Pos        { return x.LastTok().End() }
func (x *PatternExpr) End() loc.Pos       { return x.LastTok().End() }
func (x *DecmatchExpr) End() loc.Pos      { return x.LastTok().End() }
func (x *DecodedExpr) End() loc.Pos       { return x.LastTok().End() }
func (x *DefKindExpr) End() loc.Pos       { return x.LastTok().End() }
func (x *ExceptExpr) End() loc.Pos        { return x.LastTok().End() }

func (x *Ident) exprNode()             {}
func (x *ParametrizedIdent) exprNode() {}
func (x *ValueLiteral) exprNode()      {}
func (x *CompositeLiteral) exprNode()  {}
func (x *UnaryExpr) exprNode()         {}
func (x *BinaryExpr) exprNode()        {}
func (x *ParenExpr) exprNode()         {}
func (x *SelectorExpr) exprNode()      {}
func (x *IndexExpr) exprNode()         {}
func (x *CallExpr) exprNode()          {}
func (x *LengthExpr) exprNode()        {}
func (x *RedirectExpr) exprNode()      {}
func (x *ValueExpr) exprNode()         {}
func (x *ParamExpr) exprNode()         {}
func (x *FromExpr) exprNode()          {}
func (x *ModifiesExpr) exprNode()      {}
func (x *RegexpExpr) exprNode()        {}
func (x *PatternExpr) exprNode()       {}
func (x *DecmatchExpr) exprNode()      {}
func (x *DecodedExpr) exprNode()       {}
func (x *DefKindExpr) exprNode()       {}
func (x *ExceptExpr) exprNode()        {}

type (
	// A BlockStmt represents a curly braces enclosed list of statements.
	BlockStmt struct {
		LBrace Token  // Position of "{"
		Stmts  []Stmt // List of statements
		RBrace Token  // Position of "}"
	}

	// A DeclStmt represents a value declaration used as statement, lika a
	// local variable declaration.
	DeclStmt struct {
		Decl Decl
	}

	// An ExprStmt represents a expression used as statement, like an
	// assignment or function call.
	ExprStmt struct {
		Expr Expr
	}

	// A BranchStmt represents a branch statement.
	BranchStmt struct {
		Tok   Token // REPEAT, BREAK, CONTINUE, LABEL, GOTO
		Label Token // Label literal or nil
	}

	// A ReturnStmt represents a return statement.
	ReturnStmt struct {
		Tok    Token // Position of "return"
		Result Expr  // Resulting expression of nil
	}

	// A AltStmt represents an alternative statement.
	AltStmt struct {
		Tok  Token      // ALT or INTERLEAVE
		Body *BlockStmt // Block statement with alternations
	}

	// A CallStmt represents a "call" statement with communication-block.
	CallStmt struct {
		Stmt Stmt       // "call" statement
		Body *BlockStmt // Block statement with alternations
	}

	// A ForStmt represents a "for" statement.
	ForStmt struct {
		Tok      Token      // Position of "for"
		LParen   Token      // Position of "("
		Init     Stmt       // Initialization statement
		InitSemi Token      // Position of ";"
		Cond     Expr       // Conditional expression
		CondSemi Token      // Position of ";"
		Post     Stmt       // Post iteration statement
		RParen   Token      // Position of ")"
		Body     *BlockStmt // Loop-Body
	}

	// A WhilStmt represents a "while" statement.
	WhileStmt struct {
		Tok  Token      // Position of "while"
		Cond *ParenExpr // Conditional expression
		Body *BlockStmt // Loop-body
	}

	// A DoWhileStmt represents a do-while statement.
	DoWhileStmt struct {
		DoTok    Token      // Position of "do"
		Body     *BlockStmt // Loop-Body
		WhileTok Token      // Position of "while"
		Cond     *ParenExpr // Conditional expression
	}

	// A IfStmt represents a conditional statement.
	IfStmt struct {
		Tok     Token      // Position of "if"
		Cond    Expr       // Conditional expression
		Then    *BlockStmt // True branch
		ElseTok Token      // Position of "else" or nil
		Else    Stmt       // Else branch
	}

	// A SelectStmt represents a select statements.
	SelectStmt struct {
		Tok    Token         // Position of "select"
		Union  Token         // Position of "union" or nil
		Tag    *ParenExpr    // Tag expression
		LBrace Token         // Position of "{"
		Body   []*CaseClause // List of case clauses
		RBrace Token         // Position of "}"
	}

	// A CaseClause represents a case clause.
	CaseClause struct {
		Tok  Token      // Position of "case"
		Case *ParenExpr // nil means else-case
		Body *BlockStmt // Case body
	}

	// A CommClause represents communication clauses used by alt, interleave or check.
	CommClause struct {
		LBrack Token      // Position of '['
		X      Expr       // Conditional guard expression or nil
		Else   Token      // Else-clause of nil
		RBrack Token      // Position of ']'
		Comm   Stmt       // Communication statement
		Body   *BlockStmt // Body of nil
	}
)

func (x *BlockStmt) LastTok() *Token { return x.RBrace.LastTok() }
func (x *DeclStmt) LastTok() *Token  { return x.Decl.LastTok() }
func (x *ExprStmt) LastTok() *Token  { return x.Expr.LastTok() }

func (x *BranchStmt) LastTok() *Token {
	if x.Label.LastTok().IsValid() {
		return x.Label.LastTok()
	}
	return x.Tok.LastTok()
}

func (x *ReturnStmt) LastTok() *Token {
	if x.Result != nil {
		return x.Result.LastTok()
	}
	return x.Tok.LastTok()
}

func (x *CallStmt) LastTok() *Token    { return x.Body.LastTok() }
func (x *AltStmt) LastTok() *Token     { return x.Body.LastTok() }
func (x *ForStmt) LastTok() *Token     { return x.Body.LastTok() }
func (x *WhileStmt) LastTok() *Token   { return x.Body.LastTok() }
func (x *DoWhileStmt) LastTok() *Token { return x.Cond.LastTok() }

func (x *IfStmt) LastTok() *Token {
	if x.Else != nil {
		return x.Else.LastTok()
	}
	return x.Then.LastTok()
}

func (x *SelectStmt) LastTok() *Token { return x.RBrace.LastTok() }
func (x *CaseClause) LastTok() *Token { return x.Body.LastTok() }

func (x *CommClause) LastTok() *Token {
	if x.Body != nil {
		return x.Body.LastTok()
	}
	return x.Comm.LastTok()
}

func (x *BlockStmt) Pos() loc.Pos   { return x.LBrace.Pos() }
func (x *DeclStmt) Pos() loc.Pos    { return x.Decl.Pos() }
func (x *ExprStmt) Pos() loc.Pos    { return x.Expr.Pos() }
func (x *BranchStmt) Pos() loc.Pos  { return x.Tok.Pos() }
func (x *ReturnStmt) Pos() loc.Pos  { return x.Tok.Pos() }
func (x *CallStmt) Pos() loc.Pos    { return x.Stmt.Pos() }
func (x *AltStmt) Pos() loc.Pos     { return x.Tok.Pos() }
func (x *ForStmt) Pos() loc.Pos     { return x.Tok.Pos() }
func (x *WhileStmt) Pos() loc.Pos   { return x.Tok.Pos() }
func (x *DoWhileStmt) Pos() loc.Pos { return x.DoTok.Pos() }
func (x *IfStmt) Pos() loc.Pos      { return x.Tok.Pos() }
func (x *SelectStmt) Pos() loc.Pos  { return x.Tok.Pos() }
func (x *CaseClause) Pos() loc.Pos  { return x.Tok.Pos() }
func (x *CommClause) Pos() loc.Pos  { return x.LBrack.Pos() }

func (x *BlockStmt) End() loc.Pos   { return x.LastTok().End() }
func (x *DeclStmt) End() loc.Pos    { return x.LastTok().End() }
func (x *ExprStmt) End() loc.Pos    { return x.LastTok().End() }
func (x *BranchStmt) End() loc.Pos  { return x.LastTok().End() }
func (x *ReturnStmt) End() loc.Pos  { return x.LastTok().End() }
func (x *CallStmt) End() loc.Pos    { return x.LastTok().End() }
func (x *AltStmt) End() loc.Pos     { return x.LastTok().End() }
func (x *ForStmt) End() loc.Pos     { return x.LastTok().End() }
func (x *WhileStmt) End() loc.Pos   { return x.LastTok().End() }
func (x *DoWhileStmt) End() loc.Pos { return x.LastTok().End() }
func (x *IfStmt) End() loc.Pos      { return x.LastTok().End() }
func (x *SelectStmt) End() loc.Pos  { return x.LastTok().End() }
func (x *CaseClause) End() loc.Pos  { return x.LastTok().End() }
func (x *CommClause) End() loc.Pos  { return x.LastTok().End() }

func (x *BlockStmt) stmtNode()   {}
func (x *DeclStmt) stmtNode()    {}
func (x *ExprStmt) stmtNode()    {}
func (x *BranchStmt) stmtNode()  {}
func (x *ReturnStmt) stmtNode()  {}
func (x *CallStmt) stmtNode()    {}
func (x *AltStmt) stmtNode()     {}
func (x *ForStmt) stmtNode()     {}
func (x *WhileStmt) stmtNode()   {}
func (x *DoWhileStmt) stmtNode() {}
func (x *IfStmt) stmtNode()      {}
func (x *SelectStmt) stmtNode()  {}
func (x *CaseClause) stmtNode()  {}
func (x *CommClause) stmtNode()  {}

// All nested types implement TypeSpec interface.
type TypeSpec interface {
	Node
	typeSpecNode()
}

type (
	// A Field represents a named struct member or sub type definition
	Field struct {
		DefaultTok       Token    // Position of "@default" or nil
		Type             TypeSpec // Type
		Name             Expr     // Name
		TypePars         *FormalPars
		ValueConstraint  *ParenExpr  // Value constraint or nil
		LengthConstraint *LengthExpr // Length constraint or nil
		Optional         Token       // Position of "optional" or nil
	}

	// A RefSpec represents a type reference.
	RefSpec struct {
		X Expr
	}

	// A StructSpec represents a struct type specification.
	StructSpec struct {
		Kind   Token    // RECORD, SET, UNION
		LBrace Token    // Position of "{"
		Fields []*Field // Member list
		RBrace Token    // Position of "}"
	}

	// A ListSpec represents a list type specification.
	ListSpec struct {
		Kind     Token       // RECORD, SET
		Length   *LengthExpr // Length constraint or nil
		OfTok    Token       // Position of "of"
		ElemType TypeSpec    // Element type specification
	}

	// A EnumSpec represents a enumeration type specification.
	EnumSpec struct {
		Tok    Token  // Position of "enumerated"
		LBrace Token  // Position of "{"
		Enums  []Expr // Enum list
		RBrace Token  // Position of "}"
	}

	// A BehaviourSpec represents a behaviour type specification.
	BehaviourSpec struct {
		Kind   Token       // TESTCASE, FUNCTION, ALTSTEP
		Params *FormalPars // Parameter list or nil
		RunsOn *RunsOnSpec // runs on spec or nil
		System *SystemSpec // system spec or nil
		Return *ReturnSpec // return value spec or nil
	}
)

func (x *Field) LastTok() *Token {
	if x.Optional.LastTok().IsValid() {
		return x.Optional.LastTok()
	}
	if x.LengthConstraint != nil {
		return x.LengthConstraint.LastTok()
	}
	if x.ValueConstraint != nil {
		return x.ValueConstraint.LastTok()
	}
	return x.Name.LastTok()
}

func (x *RefSpec) LastTok() *Token    { return x.X.LastTok() }
func (x *StructSpec) LastTok() *Token { return x.RBrace.LastTok() }
func (x *ListSpec) LastTok() *Token   { return x.ElemType.LastTok() }
func (x *EnumSpec) LastTok() *Token   { return x.RBrace.LastTok() }

func (x *BehaviourSpec) LastTok() *Token {
	if x.Return != nil {
		return x.Return.LastTok()
	}
	if x.System != nil {
		return x.System.LastTok()
	}
	if x.RunsOn != nil {
		return x.RunsOn.LastTok()
	}
	if x.Params != nil {
		return x.Params.LastTok()
	}
	return x.Kind.LastTok()
}

func (x *Field) Pos() loc.Pos {
	if x.DefaultTok.Pos().IsValid() {
		return x.DefaultTok.Pos()
	}
	return x.Type.Pos()
}

func (x *RefSpec) Pos() loc.Pos       { return x.X.Pos() }
func (x *StructSpec) Pos() loc.Pos    { return x.Kind.Pos() }
func (x *ListSpec) Pos() loc.Pos      { return x.Kind.Pos() }
func (x *EnumSpec) Pos() loc.Pos      { return x.Tok.Pos() }
func (x *BehaviourSpec) Pos() loc.Pos { return x.Kind.Pos() }

func (x *Field) End() loc.Pos         { return x.LastTok().End() }
func (x *RefSpec) End() loc.Pos       { return x.LastTok().End() }
func (x *StructSpec) End() loc.Pos    { return x.LastTok().End() }
func (x *ListSpec) End() loc.Pos      { return x.LastTok().End() }
func (x *EnumSpec) End() loc.Pos      { return x.LastTok().End() }
func (x *BehaviourSpec) End() loc.Pos { return x.LastTok().End() }

func (x *Field) typeSpecNode()         {}
func (x *RefSpec) typeSpecNode()       {}
func (x *StructSpec) typeSpecNode()    {}
func (x *ListSpec) typeSpecNode()      {}
func (x *EnumSpec) typeSpecNode()      {}
func (x *BehaviourSpec) typeSpecNode() {}

type (
	// A ValueDecl represents a value declaration.
	ValueDecl struct {
		Kind                Token // VAR, CONST, TIMER, PORT, TEMPLATE, MODULEPAR
		TemplateRestriction *RestrictionSpec
		Modif               Token // "@lazy", "@fuzzy" or nil
		Type                Expr
		Decls               []Expr
		With                *WithSpec
	}

	TemplateDecl struct {
		RestrictionSpec
		Modif       Token // "@lazy", "@fuzzy" or nil
		Type        Expr
		Name        *Ident
		TypePars    *FormalPars
		Params      *FormalPars
		ModifiesTok Token
		Base        Expr
		AssignTok   Token
		Value       Expr
		With        *WithSpec
	}

	// A ModuleParameterGroup represents a deprecated module parameter list
	ModuleParameterGroup struct {
		Tok    Token        // Position of "modulepar"
		LBrace Token        // Position of "{"
		Decls  []*ValueDecl // Module parameter list
		RBrace Token        // Position of "}"
		With   *WithSpec
	}

	// A FuncDecl represents a behaviour definition.
	FuncDecl struct {
		External Token // Position of "external" or nil
		Kind     Token // TESTCASE, ALTSTEP, FUNCTION
		Name     *Ident
		Modif    Token // Position of "@deterministic" or nil
		TypePars *FormalPars
		Params   *FormalPars // Formal parameter list or nil
		RunsOn   *RunsOnSpec // Optional runs-on-spec
		Mtc      *MtcSpec    // Optional mtc-spec
		System   *SystemSpec // Optional system-spec
		Return   *ReturnSpec // Optional return-spec
		Body     *BlockStmt  // Body or nil
		With     *WithSpec
	}

	// A SignatureDecl represents a signature type for procedure based communication.
	SignatureDecl struct {
		Tok          Token // Position of "signature"
		Name         *Ident
		TypePars     *FormalPars
		Params       *FormalPars
		NoBlock      Token       // Optional "noblock"
		Return       *ReturnSpec // Optional return-spec
		ExceptionTok Token       // Position of "exeception" or nil
		Exception    *ParenExpr  // Exception list
		With         *WithSpec
	}

	// A SubTypeDecl represents a named sub type declaration
	SubTypeDecl struct {
		TypeTok Token  // Position of "type"
		Field   *Field // Field spec
		With    *WithSpec
	}

	// A StructTypeDecl represents a name struct type.
	StructTypeDecl struct {
		TypeTok  Token // Position of "type"
		Kind     Token // RECORD, SET, UNION
		Name     Expr  // Name
		TypePars *FormalPars
		LBrace   Token    // Position of "{"
		Fields   []*Field // Member list
		RBrace   Token    // Position of }"
		With     *WithSpec
	}

	// A EnumTypeDecl represents a named enum type.
	EnumTypeDecl struct {
		TypeTok  Token // Position of "type"
		EnumTok  Token // Position of "ENUMERATED"
		Name     Expr
		TypePars *FormalPars
		LBrace   Token  // Position of "{"
		Enums    []Expr // Enum list
		RBrace   Token  // Position of "}"
		With     *WithSpec
	}

	// A BehaviourTypeDecl represents a named behaviour type.
	BehaviourTypeDecl struct {
		TypeTok  Token // Position of "type"
		Kind     Token // TESTCASE, ALTSTEP, FUNCTION
		Name     Expr
		TypePars *FormalPars
		Params   *FormalPars // Formal parameter list
		RunsOn   *RunsOnSpec // Optional runs-on spec
		System   *SystemSpec // Optional system spec
		Return   *ReturnSpec // Optional return spec
		With     *WithSpec
	}

	PortTypeDecl struct {
		TypeTok  Token
		PortTok  Token
		Name     Expr
		TypePars *FormalPars
		Kind     Token // MIXED, MESSAGE, PROCEDURE
		Realtime Token
		LBrace   Token
		Attrs    []Node
		RBrace   Token
		With     *WithSpec
	}

	PortAttribute struct {
		Kind  Token // IN, OUT, INOUT, ADDRESS
		Types []Expr
	}

	PortMapAttribute struct {
		MapTok   Token // MAP, UNMAP
		ParamTok Token
		Params   *FormalPars
	}

	ComponentTypeDecl struct {
		TypeTok    Token
		CompTok    Token
		Name       Expr
		TypePars   *FormalPars
		ExtendsTok Token
		Extends    []Expr
		Body       *BlockStmt
		With       *WithSpec
	}
)

func (x *FuncDecl) IsTest() bool {
	return x.Kind.Kind == token.TESTCASE
}

func (x *ValueDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	return x.Decls[len(x.Decls)-1].LastTok()
}

func (x *TemplateDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	return x.Value.LastTok()
}

func (x *ModuleParameterGroup) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	return x.RBrace.LastTok()
}

func (x *FuncDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	if x.Body != nil {
		return x.Body.LastTok()
	}
	if x.Return != nil {
		return x.Return.LastTok()
	}
	if x.Params != nil {
		return x.Params.LastTok()
	}
	return x.Kind.LastTok()
}

func (x *SignatureDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	if x.Exception != nil {
		return x.Exception.LastTok()
	}
	if x.Return != nil {
		return x.Return.LastTok()
	}
	if x.NoBlock.LastTok().IsValid() {
		return x.NoBlock.LastTok()
	}
	if x.Params != nil {
		return x.Params.LastTok()
	}
	return x.Name.LastTok()
}

func (x *SubTypeDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	return x.Field.LastTok()
}

func (x *StructTypeDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	return x.RBrace.LastTok()
}

func (x *EnumTypeDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	return x.RBrace.LastTok()
}

func (x *BehaviourTypeDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	if x.Return != nil {
		return x.Return.LastTok()
	}
	if x.System != nil {
		return x.System.LastTok()
	}
	if x.RunsOn != nil {
		return x.RunsOn.LastTok()
	}
	if x.Params != nil {
		return x.Params.LastTok()
	}
	return x.Name.LastTok()
}

func (x *PortTypeDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	return x.RBrace.LastTok()
}

func (x *PortAttribute) LastTok() *Token {
	return x.Types[len(x.Types)-1].LastTok()
}

func (x *PortMapAttribute) LastTok() *Token {
	return x.Params.LastTok()
}

func (x *ComponentTypeDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	return x.Body.LastTok()
}

func (x *ValueDecl) Pos() loc.Pos {
	if x.Kind.Pos().IsValid() {
		return x.Kind.Pos()
	}
	return x.Type.Pos()
}

func (x *ModuleParameterGroup) Pos() loc.Pos {
	return x.Tok.Pos()
}

func (x *FuncDecl) Pos() loc.Pos {
	if x.External.Kind == token.EXTERNAL {
		return x.External.Pos()
	}
	return x.Kind.Pos()
}

func (x *SignatureDecl) Pos() loc.Pos     { return x.Tok.Pos() }
func (x *SubTypeDecl) Pos() loc.Pos       { return x.TypeTok.Pos() }
func (x *StructTypeDecl) Pos() loc.Pos    { return x.TypeTok.Pos() }
func (x *EnumTypeDecl) Pos() loc.Pos      { return x.TypeTok.Pos() }
func (x *BehaviourTypeDecl) Pos() loc.Pos { return x.TypeTok.Pos() }
func (x *PortTypeDecl) Pos() loc.Pos      { return x.TypeTok.Pos() }
func (x *PortAttribute) Pos() loc.Pos     { return x.Kind.Pos() }
func (x *PortMapAttribute) Pos() loc.Pos  { return x.MapTok.Pos() }
func (x *ComponentTypeDecl) Pos() loc.Pos { return x.TypeTok.Pos() }

func (x *ValueDecl) End() loc.Pos            { return x.LastTok().End() }
func (x *TemplateDecl) End() loc.Pos         { return x.LastTok().End() }
func (x *ModuleParameterGroup) End() loc.Pos { return x.LastTok().End() }
func (x *FuncDecl) End() loc.Pos             { return x.LastTok().End() }
func (x *SignatureDecl) End() loc.Pos        { return x.LastTok().End() }
func (x *SubTypeDecl) End() loc.Pos          { return x.LastTok().End() }
func (x *StructTypeDecl) End() loc.Pos       { return x.LastTok().End() }
func (x *EnumTypeDecl) End() loc.Pos         { return x.LastTok().End() }
func (x *BehaviourTypeDecl) End() loc.Pos    { return x.LastTok().End() }
func (x *PortTypeDecl) End() loc.Pos         { return x.LastTok().End() }
func (x *PortAttribute) End() loc.Pos        { return x.LastTok().End() }
func (x *PortMapAttribute) End() loc.Pos     { return x.LastTok().End() }
func (x *ComponentTypeDecl) End() loc.Pos    { return x.LastTok().End() }

func (x *ValueDecl) declNode()            {}
func (x *TemplateDecl) declNode()         {}
func (x *ModuleParameterGroup) declNode() {}
func (x *FuncDecl) declNode()             {}
func (x *SignatureDecl) declNode()        {}
func (x *SubTypeDecl) declNode()          {}
func (x *StructTypeDecl) declNode()       {}
func (x *EnumTypeDecl) declNode()         {}
func (x *BehaviourTypeDecl) declNode()    {}
func (x *PortTypeDecl) declNode()         {}
func (x *PortAttribute) declNode()        {}
func (x *PortMapAttribute) declNode()     {}
func (x *ComponentTypeDecl) declNode()    {}

// ------------------------------------------------------------------------
// Modules and Module Definitions

type (
	Module struct {
		Tok      Token
		Name     *Ident
		Language *LanguageSpec
		LBrace   Token
		Defs     []*ModuleDef
		RBrace   Token
		With     *WithSpec
	}

	ModuleDef struct {
		Visibility Token
		Def        Node
	}

	ControlPart struct {
		Tok  Token
		Body *BlockStmt
		With *WithSpec
	}

	ImportDecl struct {
		ImportTok Token
		FromTok   Token
		Module    *Ident
		Language  *LanguageSpec
		LBrace    Token
		List      []*DefKindExpr
		RBrace    Token
		With      *WithSpec
	}

	GroupDecl struct {
		Tok    Token
		Name   *Ident
		LBrace Token
		Defs   []*ModuleDef
		RBrace Token
		With   *WithSpec
	}

	FriendDecl struct {
		FriendTok Token
		ModuleTok Token
		Module    *Ident
		With      *WithSpec
	}
)

func (x *Module) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	if x.RBrace.IsValid() {
		return x.RBrace.LastTok()
	}
	if l := len(x.Defs); l > 0 {
		return x.Defs[l-1].LastTok()
	}
	if x.Language != nil {
		return x.Language.LastTok()
	}

	if x.Name != nil {
		return x.Name.LastTok()
	}

	return x.Tok.LastTok()
}

func (x *ModuleDef) LastTok() *Token {
	return x.Def.LastTok()
}

func (x *ControlPart) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	return x.Body.LastTok()
}

func (x *ImportDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	return x.RBrace.LastTok()
}

func (x *GroupDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	return x.RBrace.LastTok()
}

func (x *FriendDecl) LastTok() *Token {
	if x.With != nil {
		return x.With.LastTok()
	}
	return x.Module.LastTok()
}

func (x *Module) Pos() loc.Pos { return x.Tok.Pos() }

func (x *ModuleDef) Pos() loc.Pos {
	if x.Visibility.Pos().IsValid() {
		return x.Visibility.Pos()
	}
	return x.Def.Pos()
}

func (x *ControlPart) Pos() loc.Pos { return x.Tok.Pos() }
func (x *ImportDecl) Pos() loc.Pos  { return x.ImportTok.Pos() }
func (x *GroupDecl) Pos() loc.Pos   { return x.Tok.Pos() }
func (x *FriendDecl) Pos() loc.Pos  { return x.FriendTok.Pos() }

func (x *Module) End() loc.Pos      { return x.LastTok().End() }
func (x *ModuleDef) End() loc.Pos   { return x.LastTok().End() }
func (x *ControlPart) End() loc.Pos { return x.LastTok().End() }
func (x *ImportDecl) End() loc.Pos  { return x.LastTok().End() }
func (x *GroupDecl) End() loc.Pos   { return x.LastTok().End() }
func (x *FriendDecl) End() loc.Pos  { return x.LastTok().End() }

// ------------------------------------------------------------------------
// Miscellaneous

type (
	LanguageSpec struct {
		Tok  Token
		List []Token
	}

	RestrictionSpec struct {
		TemplateTok Token
		LParen      Token
		Tok         Token
		RParen      Token
	}

	RunsOnSpec struct {
		RunsTok Token
		OnTok   Token
		Comp    Expr
	}

	SystemSpec struct {
		Tok  Token
		Comp Expr
	}

	MtcSpec struct {
		Tok  Token
		Comp Expr
	}

	ReturnSpec struct {
		Tok         Token
		Restriction *RestrictionSpec
		Modif       Token
		Type        Expr
	}

	FormalPars struct {
		LParen Token
		List   []*FormalPar
		RParen Token
	}

	FormalPar struct {
		Direction           Token
		TemplateRestriction *RestrictionSpec
		Modif               Token
		Type                Expr
		Name                Expr
	}

	WithSpec struct {
		Tok    Token
		LBrace Token
		List   []*WithStmt
		RBrace Token
	}

	WithStmt struct {
		Kind     Token
		Override Token
		LParen   Token
		List     []Node
		RParen   Token
		Value    Expr
	}
)

func (x *LanguageSpec) LastTok() *Token { return x.List[len(x.List)-1].LastTok() }
func (x *RestrictionSpec) LastTok() *Token {
	if x.RParen.LastTok().IsValid() {
		return x.RParen.LastTok()
	}
	return x.Tok.LastTok()
}
func (x *RunsOnSpec) LastTok() *Token { return x.Comp.LastTok() }
func (x *SystemSpec) LastTok() *Token { return x.Comp.LastTok() }
func (x *MtcSpec) LastTok() *Token    { return x.Comp.LastTok() }
func (x *ReturnSpec) LastTok() *Token { return x.Type.LastTok() }
func (x *FormalPars) LastTok() *Token { return x.RParen.LastTok() }
func (x *FormalPar) LastTok() *Token  { return x.Name.LastTok() }
func (x *WithSpec) LastTok() *Token   { return x.RBrace.LastTok() }
func (x *WithStmt) LastTok() *Token   { return x.Value.LastTok() }

func (x *LanguageSpec) Pos() loc.Pos { return x.Tok.Pos() }
func (x *RestrictionSpec) Pos() loc.Pos {
	if x.TemplateTok.Pos().IsValid() {
		return x.TemplateTok.Pos()
	}
	return x.Tok.Pos()
}
func (x *RunsOnSpec) Pos() loc.Pos { return x.RunsTok.Pos() }
func (x *SystemSpec) Pos() loc.Pos { return x.Tok.Pos() }
func (x *MtcSpec) Pos() loc.Pos    { return x.Tok.Pos() }
func (x *ReturnSpec) Pos() loc.Pos { return x.Tok.Pos() }
func (x *FormalPars) Pos() loc.Pos { return x.LParen.Pos() }

func (x *FormalPar) Pos() loc.Pos {
	if x.Direction.Pos().IsValid() {
		return x.Direction.Pos()
	}
	if x.TemplateRestriction.Pos().IsValid() {
		return x.TemplateRestriction.Pos()
	}
	if x.Modif.Pos().IsValid() {
		return x.Modif.Pos()
	}
	return x.Type.Pos()
}

func (x *WithSpec) Pos() loc.Pos        { return x.LastTok().End() }
func (x *WithStmt) Pos() loc.Pos        { return x.LastTok().End() }
func (x *LanguageSpec) End() loc.Pos    { return x.LastTok().End() }
func (x *RestrictionSpec) End() loc.Pos { return x.LastTok().End() }
func (x *RunsOnSpec) End() loc.Pos      { return x.LastTok().End() }
func (x *SystemSpec) End() loc.Pos      { return x.LastTok().End() }
func (x *MtcSpec) End() loc.Pos         { return x.LastTok().End() }
func (x *ReturnSpec) End() loc.Pos      { return x.LastTok().End() }
func (x *FormalPars) End() loc.Pos      { return x.LastTok().End() }
func (x *FormalPar) End() loc.Pos       { return x.LastTok().End() }
func (x *WithSpec) End() loc.Pos        { return x.LastTok().End() }
func (x *WithStmt) End() loc.Pos        { return x.LastTok().End() }
