package syntax

// All node types implement the Node interface.
type Node interface {
	Pos() Pos
	End() Pos
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

// A Token represents a TTCN-3 token and implements the Node interface. Tokens are leave-nodes.
type Token struct {
	pos  Pos
	Kind Kind   // Token kind like TESTCASE, SEMICOLON, COMMENT, ...
	Lit  string // Token values for non-operator tokens
}

func (x Token) Pos() Pos { return x.pos }
func (x Token) End() Pos {
	if x.Kind.IsLiteral() {
		return Pos(int(x.pos) + len(x.Lit))
	}
	return Pos(int(x.pos) + len(x.Kind.String()))
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

	// A DefSelectorExpr represents a definition type selector expression,
	// used by imports and with attributes, like "all types except a, b"
	DefSelectorExpr struct {
		Kind      Token  // TYPE, TEMPLATE, CONST, ...
		Refs      []Expr // ALL, ids
		ExceptTok Token  // Position of "except" or nil
		LBrace    Token  // Position of "{" or nil
		Except    []Expr // Expression list
		RBrace    Token  // Position of "}" or nil
	}
)

func (x *Ident) Pos() Pos             { return x.Tok.Pos() }
func (x *ParametrizedIdent) Pos() Pos { return x.Ident.Pos() }
func (x *ValueLiteral) Pos() Pos      { return x.Tok.Pos() }
func (x *CompositeLiteral) Pos() Pos  { return x.LBrace.Pos() }
func (x *UnaryExpr) Pos() Pos         { return x.Op.Pos() }
func (x *BinaryExpr) Pos() Pos        { return x.X.Pos() }
func (x *ParenExpr) Pos() Pos         { return x.LParen.Pos() }
func (x *SelectorExpr) Pos() Pos      { return x.X.Pos() }
func (x *IndexExpr) Pos() Pos         { return x.X.Pos() }
func (x *CallExpr) Pos() Pos          { return x.Fun.Pos() }
func (x *LengthExpr) Pos() Pos        { return x.X.Pos() }
func (x *RedirectExpr) Pos() Pos      { return x.X.Pos() }
func (x *ValueExpr) Pos() Pos         { return x.X.Pos() }
func (x *ParamExpr) Pos() Pos         { return x.X.Pos() }
func (x *FromExpr) Pos() Pos          { return x.Kind.Pos() }
func (x *ModifiesExpr) Pos() Pos      { return x.Tok.Pos() }
func (x *RegexpExpr) Pos() Pos        { return x.Tok.Pos() }
func (x *PatternExpr) Pos() Pos       { return x.Tok.Pos() }
func (x *DecmatchExpr) Pos() Pos      { return x.Tok.Pos() }
func (x *DecodedExpr) Pos() Pos       { return x.Tok.Pos() }
func (x *DefSelectorExpr) Pos() Pos   { return x.Kind.Pos() }

func (x *Ident) End() Pos {
	if x.Tok2.End() != NoPos {
		return x.Tok2.End()
	}
	return x.Tok.End()
}

func (x *ParametrizedIdent) End() Pos { return x.Params.End() }
func (x *ValueLiteral) End() Pos      { return x.Tok.End() }
func (x *CompositeLiteral) End() Pos  { return x.RBrace.End() }
func (x *UnaryExpr) End() Pos         { return x.X.End() }
func (x *BinaryExpr) End() Pos        { return x.Y.End() }
func (x *ParenExpr) End() Pos         { return x.RParen.End() }
func (x *SelectorExpr) End() Pos      { return x.Sel.End() }
func (x *IndexExpr) End() Pos         { return x.RBrack.End() }
func (x *CallExpr) End() Pos          { return x.Args.End() }
func (x *LengthExpr) End() Pos        { return x.Size.End() }
func (x *RedirectExpr) End() Pos {
	if x.Timestamp != nil {
		return x.Timestamp.End()
	}
	if x.Index != nil {
		return x.Index.End()
	}
	if x.Sender != nil {
		return x.Sender.End()
	}
	if x.Param != nil {
		return x.Param[len(x.Param)-1].End()
	}
	if x.Value != nil {
		return x.Value[len(x.Value)-1].End()
	}
	return x.Tok.End()
}

func (x *ValueExpr) End() Pos    { return x.Y.End() }
func (x *ParamExpr) End() Pos    { return x.Y.End() }
func (x *FromExpr) End() Pos     { return x.X.End() }
func (x *ModifiesExpr) End() Pos { return x.Y.End() }
func (x *RegexpExpr) End() Pos   { return x.X.End() }
func (x *PatternExpr) End() Pos  { return x.X.End() }
func (x *DecmatchExpr) End() Pos { return x.X.End() }
func (x *DecodedExpr) End() Pos  { return x.X.End() }

func (x *DefSelectorExpr) End() Pos {
	if x.RBrace.End() != NoPos {
		return x.RBrace.End()
	}
	if x.Except != nil {
		return x.Except[len(x.Except)-1].End()
	}
	return x.Refs[len(x.Refs)-1].End()
}

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
func (x *DefSelectorExpr) exprNode()   {}

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

func (x *BlockStmt) Pos() Pos   { return x.LBrace.Pos() }
func (x *DeclStmt) Pos() Pos    { return x.Decl.Pos() }
func (x *ExprStmt) Pos() Pos    { return x.Expr.Pos() }
func (x *BranchStmt) Pos() Pos  { return x.Tok.Pos() }
func (x *ReturnStmt) Pos() Pos  { return x.Tok.Pos() }
func (x *CallStmt) Pos() Pos    { return x.Stmt.Pos() }
func (x *AltStmt) Pos() Pos     { return x.Tok.Pos() }
func (x *ForStmt) Pos() Pos     { return x.Tok.Pos() }
func (x *WhileStmt) Pos() Pos   { return x.Tok.Pos() }
func (x *DoWhileStmt) Pos() Pos { return x.DoTok.Pos() }
func (x *IfStmt) Pos() Pos      { return x.Tok.Pos() }
func (x *SelectStmt) Pos() Pos  { return x.Tok.Pos() }
func (x *CaseClause) Pos() Pos  { return x.Tok.Pos() }
func (x *CommClause) Pos() Pos  { return x.LBrack.Pos() }

func (x *BlockStmt) End() Pos { return x.RBrace.End() }
func (x *DeclStmt) End() Pos  { return x.Decl.End() }
func (x *ExprStmt) End() Pos  { return x.Expr.End() }

func (x *BranchStmt) End() Pos {
	if x.Label.End() != NoPos {
		return x.Label.End()
	}
	return x.Tok.End()
}

func (x *ReturnStmt) End() Pos {
	if x.Result != nil {
		return x.Result.End()
	}
	return x.Tok.End()
}

func (x *CallStmt) End() Pos    { return x.Body.End() }
func (x *AltStmt) End() Pos     { return x.Body.End() }
func (x *ForStmt) End() Pos     { return x.Body.End() }
func (x *WhileStmt) End() Pos   { return x.Body.End() }
func (x *DoWhileStmt) End() Pos { return x.Cond.End() }

func (x *IfStmt) End() Pos {
	if x.Else != nil {
		return x.Else.End()
	}
	return x.Then.End()
}

func (x *SelectStmt) End() Pos { return x.RBrace.End() }
func (x *CaseClause) End() Pos { return x.Body.End() }

func (x *CommClause) End() Pos {
	if x.Body != nil {
		return x.Body.End()
	}
	return x.Comm.End()
}

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

func (x *Field) Pos() Pos {
	if x.DefaultTok.Pos() != NoPos {
		return x.DefaultTok.Pos()
	}
	return x.Type.Pos()
}

func (x *RefSpec) Pos() Pos       { return x.X.Pos() }
func (x *StructSpec) Pos() Pos    { return x.Kind.Pos() }
func (x *ListSpec) Pos() Pos      { return x.Kind.Pos() }
func (x *EnumSpec) Pos() Pos      { return x.Tok.Pos() }
func (x *BehaviourSpec) Pos() Pos { return x.Kind.Pos() }

func (x *Field) End() Pos {
	if x.Optional.End() != NoPos {
		return x.Optional.End()
	}
	if x.LengthConstraint != nil {
		return x.LengthConstraint.End()
	}
	if x.ValueConstraint != nil {
		return x.ValueConstraint.End()
	}
	return x.Name.End()
}

func (x *RefSpec) End() Pos    { return x.X.End() }
func (x *StructSpec) End() Pos { return x.RBrace.End() }
func (x *ListSpec) End() Pos   { return x.ElemType.End() }
func (x *EnumSpec) End() Pos   { return x.RBrace.End() }

func (x *BehaviourSpec) End() Pos {
	if x.Return != nil {
		return x.Return.End()
	}
	if x.System != nil {
		return x.System.End()
	}
	if x.RunsOn != nil {
		return x.RunsOn.End()
	}
	if x.Params != nil {
		return x.Params.End()
	}
	return x.Kind.End()
}

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

func (x *ValueDecl) Pos() Pos {
	if x.Kind.Pos() != NoPos {
		return x.Kind.Pos()
	}
	return x.Type.Pos()
}

func (x *ModuleParameterGroup) Pos() Pos {
	return x.Tok.Pos()
}

func (x *FuncDecl) Pos() Pos {
	if x.External.Kind == EXTERNAL {
		return x.External.Pos()
	}
	return x.Kind.Pos()
}

func (x *SignatureDecl) Pos() Pos     { return x.Tok.Pos() }
func (x *SubTypeDecl) Pos() Pos       { return x.TypeTok.Pos() }
func (x *StructTypeDecl) Pos() Pos    { return x.TypeTok.Pos() }
func (x *EnumTypeDecl) Pos() Pos      { return x.TypeTok.Pos() }
func (x *BehaviourTypeDecl) Pos() Pos { return x.TypeTok.Pos() }
func (x *PortTypeDecl) Pos() Pos      { return x.TypeTok.Pos() }
func (x *PortAttribute) Pos() Pos     { return x.Kind.Pos() }
func (x *PortMapAttribute) Pos() Pos  { return x.MapTok.Pos() }
func (x *ComponentTypeDecl) Pos() Pos { return x.TypeTok.Pos() }

func (x *ValueDecl) End() Pos {
	if x.With != nil {
		return x.With.End()
	}
	return x.Decls[len(x.Decls)-1].End()
}

func (x *ModuleParameterGroup) End() Pos {
	if x.With != nil {
		return x.With.End()
	}
	return x.RBrace.End()
}

func (x *FuncDecl) End() Pos {
	if x.With != nil {
		return x.With.End()
	}
	if x.Body != nil {
		return x.Body.End()
	}
	if x.Return != nil {
		return x.Return.End()
	}
	if x.Params != nil {
		return x.Params.End()
	}
	return NoPos
}

func (x *SignatureDecl) End() Pos {
	if x.With != nil {
		return x.With.End()
	}
	if x.Exception != nil {
		return x.Exception.End()
	}
	if x.Return != nil {
		return x.Return.End()
	}
	if x.NoBlock.End() != NoPos {
		return x.NoBlock.End()
	}
	if x.Params != nil {
		return x.Params.End()
	}
	return x.Name.End()
}

func (x *SubTypeDecl) End() Pos {
	if x.With != nil {
		return x.With.End()
	}
	return x.Field.End()
}

func (x *StructTypeDecl) End() Pos {
	if x.With != nil {
		return x.With.End()
	}
	return x.RBrace.End()
}

func (x *EnumTypeDecl) End() Pos {
	if x.With != nil {
		return x.With.End()
	}
	return x.RBrace.End()
}

func (x *BehaviourTypeDecl) End() Pos {
	if x.With != nil {
		return x.With.End()
	}
	if x.Return != nil {
		return x.Return.End()
	}
	if x.System != nil {
		return x.System.End()
	}
	if x.RunsOn != nil {
		return x.RunsOn.End()
	}
	if x.Params != nil {
		return x.Params.End()
	}
	return x.Name.End()
}

func (x *PortTypeDecl) End() Pos {
	if x.With != nil {
		return x.With.End()
	}
	return x.RBrace.End()
}

func (x *PortAttribute) End() Pos {
	return x.Types[len(x.Types)-1].End()
}

func (x *PortMapAttribute) End() Pos {
	return x.Params.End()
}

func (x *ComponentTypeDecl) End() Pos {
	if x.With != nil {
		return x.With.End()
	}
	return x.Body.End()
}

func (x *ValueDecl) declNode()            {}
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
		Decls    []*ModuleDef
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
		List      []*DefSelectorExpr
		RBrace    Token
		With      *WithSpec
	}

	ExceptSpec struct {
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

func (x *Module) Pos() Pos { return x.Tok.Pos() }

func (x *ModuleDef) Pos() Pos {
	if x.Visibility.Pos() != NoPos {
		return x.Visibility.Pos()
	}
	return x.Def.Pos()
}

func (x *ControlPart) Pos() Pos { return x.Tok.Pos() }
func (x *ImportDecl) Pos() Pos  { return x.ImportTok.Pos() }
func (x *ExceptSpec) Pos() Pos  { return NoPos }
func (x *GroupDecl) Pos() Pos   { return x.Tok.Pos() }
func (x *FriendDecl) Pos() Pos  { return x.FriendTok.Pos() }

func (x *Module) End() Pos {
	if x.With != nil {
		return x.With.End()
	}
	return x.RBrace.End()
}

func (x *ModuleDef) End() Pos {
	return x.Def.End()
}

func (x *ControlPart) End() Pos {
	if x.With != nil {
		return x.With.End()
	}
	return x.Body.End()
}

func (x *ImportDecl) End() Pos {
	if x.With != nil {
		return x.With.End()
	}
	return x.RBrace.End()
}

func (x *ExceptSpec) End() Pos {
	return NoPos //TODO
}

func (x *GroupDecl) End() Pos {
	if x.With != nil {
		return x.With.End()
	}
	return x.RBrace.End()
}

func (x *FriendDecl) End() Pos {
	if x.With != nil {
		return x.With.End()
	}
	return x.Module.End()
}

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

func (x *LanguageSpec) Pos() Pos { return x.Tok.Pos() }
func (x *RestrictionSpec) Pos() Pos {
	if x.TemplateTok.Pos() != NoPos {
		return x.TemplateTok.Pos()
	}
	return x.Tok.Pos()
}
func (x *RunsOnSpec) Pos() Pos { return x.RunsTok.Pos() }
func (x *SystemSpec) Pos() Pos { return x.Tok.Pos() }
func (x *MtcSpec) Pos() Pos    { return x.Tok.Pos() }
func (x *ReturnSpec) Pos() Pos { return x.Tok.Pos() }
func (x *FormalPars) Pos() Pos { return x.LParen.Pos() }

func (x *FormalPar) Pos() Pos {
	if x.Direction.Pos() != NoPos {
		return x.Direction.Pos()
	}
	if x.TemplateRestriction.Pos() != NoPos {
		return x.TemplateRestriction.Pos()
	}
	if x.Modif.Pos() != NoPos {
		return x.Modif.Pos()
	}
	return x.Type.Pos()
}

func (x *WithSpec) Pos() Pos { return x.Tok.Pos() }
func (x *WithStmt) Pos() Pos { return x.Kind.Pos() }

func (x *LanguageSpec) End() Pos { return x.List[len(x.List)-1].End() }
func (x *RestrictionSpec) End() Pos {
	if x.RParen.End() != NoPos {
		return x.RParen.End()
	}
	return x.Tok.End()
}
func (x *RunsOnSpec) End() Pos { return x.Comp.End() }
func (x *SystemSpec) End() Pos { return x.Comp.End() }
func (x *MtcSpec) End() Pos    { return x.Comp.End() }
func (x *ReturnSpec) End() Pos { return x.Type.End() }
func (x *FormalPars) End() Pos { return x.RParen.End() }
func (x *FormalPar) End() Pos  { return x.Name.End() }
func (x *WithSpec) End() Pos   { return x.RBrace.End() }
func (x *WithStmt) End() Pos   { return x.Value.End() }
