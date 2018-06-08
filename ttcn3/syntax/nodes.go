package syntax

// ----------------------------------------------------------------------------
// Interfaces
//

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

// Tokens
// ------------------------------------------------------------------------

type Token struct {
	pos  Pos
	Kind Kind
	Lit  string
}

func (x Token) Pos() Pos { return x.pos }
func (x Token) End() Pos {
	if l := len(x.Lit); l != 0 {
		return x.pos + Pos(l)
	}
	if l := len(x.Kind.String()); l != 0 {
		return x.pos + Pos(l)
	}
	return NoPos
}

// ------------------------------------------------------------------------
// Expressions
//
type (
	Ident struct {
		Tok  Token
		Tok2 Token
	}

	ParametrizedIdent struct {
		Ident  *Ident
		Params Expr
	}

	ValueLiteral struct {
		Tok Token
	}

	CompositeLiteral struct {
		LBrace Token
		List   []Expr
		RBrace Token
	}

	UnaryExpr struct {
		Op Token
		X  Expr
	}

	BinaryExpr struct {
		X  Expr
		Op Token
		Y  Expr
	}

	ParenExpr struct {
		LParen Token
		List   []Expr
		RParen Token
	}

	SelectorExpr struct {
		X   Expr
		Dot Token
		Sel Expr
	}

	IndexExpr struct {
		X      Expr
		LBrack Token
		Index  Expr
		RBrack Token
	}

	CallExpr struct {
		Fun  Expr
		Args *ParenExpr
	}

	LengthExpr struct {
		X    Expr
		Len  Token
		Size Expr
	}

	RedirectExpr struct {
		X             Expr
		Tok           Token
		ValueTok      Token
		Value         []Expr
		ParamTok      Token
		Param         []Expr
		SenderTok     Token
		Sender        Expr
		IndexTok      Token
		IndexValueTok Token
		Index         Expr
		TimestampTok  Token
		Timestamp     Expr
	}

	// Required for Signatures
	ValueExpr struct {
		X   Expr
		Tok Token
		Y   Expr
	}

	// Required for map param, unmap param
	ParamExpr struct {
		X   Expr
		Tok Token
		Y   Expr
	}

	FromExpr struct {
		Kind    Token
		FromTok Token
		X       Expr
	}

	ModifiesExpr struct {
		Tok    Token
		X      Expr
		Assign Token
		Y      Expr
	}

	RegexpExpr struct {
		Tok    Token
		NoCase Token
		X      Expr
	}

	PatternExpr struct {
		Tok    Token
		NoCase Token
		X      Expr
	}

	DecmatchExpr struct {
		Tok    Token
		Params Expr
		X      Expr
	}

	DecodedExpr struct {
		Tok    Token
		Params Expr
		X      Expr
	}

	DefSelectorExpr struct {
		Kind      Token  // TYPE, TEMPLATE, CONST, ...
		Refs      []Expr // ALL, ids
		ExceptTok Token
		LBrace    Token
		Except    []Expr
		RBrace    Token
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

// ------------------------------------------------------------------------
// Statements

type (
	BlockStmt struct {
		LBrace Token
		Stmts  []Stmt
		RBrace Token
	}

	DeclStmt struct {
		Decl Decl
	}

	ExprStmt struct {
		Expr Expr
	}

	BranchStmt struct {
		Tok   Token
		Label Token
	}

	ReturnStmt struct {
		Tok    Token
		Result Expr
	}

	AltStmt struct {
		Tok  Token
		Body *BlockStmt
	}

	ForStmt struct {
		Tok      Token
		LParen   Token
		Init     Stmt
		InitSemi Token
		Cond     Expr
		CondSemi Token
		Post     Stmt
		RParen   Token
		Body     *BlockStmt
	}

	WhileStmt struct {
		Tok  Token
		Cond *ParenExpr
		Body *BlockStmt
	}

	DoWhileStmt struct {
		DoTok    Token
		Body     *BlockStmt
		WhileTok Token
		Cond     *ParenExpr
	}

	IfStmt struct {
		Tok     Token
		Cond    Expr
		Then    *BlockStmt
		ElseTok Token
		Else    Stmt
	}

	SelectStmt struct {
		Tok    Token
		Union  Token
		Tag    *ParenExpr
		LBrace Token
		Body   []*CaseClause
		RBrace Token
	}

	CaseClause struct {
		Tok  Token
		Case *ParenExpr // nil means else-case
		Body *BlockStmt
	}

	CommClause struct {
		LBrack Token
		X      Expr
		Else   Token
		RBrack Token
		Comm   Stmt
		Body   *BlockStmt
	}
)

func (x *BlockStmt) Pos() Pos   { return x.LBrace.Pos() }
func (x *DeclStmt) Pos() Pos    { return x.Decl.Pos() }
func (x *ExprStmt) Pos() Pos    { return x.Expr.Pos() }
func (x *BranchStmt) Pos() Pos  { return x.Tok.Pos() }
func (x *ReturnStmt) Pos() Pos  { return x.Tok.Pos() }
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
func (x *AltStmt) stmtNode()     {}
func (x *ForStmt) stmtNode()     {}
func (x *WhileStmt) stmtNode()   {}
func (x *DoWhileStmt) stmtNode() {}
func (x *IfStmt) stmtNode()      {}
func (x *SelectStmt) stmtNode()  {}
func (x *CaseClause) stmtNode()  {}
func (x *CommClause) stmtNode()  {}

// ------------------------------------------------------------------------
// Declarations and Types

type TypeSpec interface {
	Node
	typeSpecNode()
}

type (
	Field struct {
		DefaultTok       Token
		Type             TypeSpec
		Name             Expr
		ValueConstraint  *ParenExpr
		LengthConstraint *LengthExpr
		Optional         Token
	}

	RefSpec struct {
		X Expr
	}

	StructSpec struct {
		Kind   Token // RECORD, SET, UNION
		LBrace Token
		Fields []*Field
		RBrace Token
	}

	ListSpec struct {
		Kind     Token // RECORD, SET
		Length   *LengthExpr
		OfTok    Token
		ElemType TypeSpec
	}

	EnumSpec struct {
		Tok    Token
		LBrace Token
		Enums  []Expr
		RBrace Token
	}

	BehaviourSpec struct {
		Kind   Token
		Params *FormalPars
		RunsOn *RunsOnSpec
		System *SystemSpec
		Return *ReturnSpec
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
	ValueDecl struct {
		Kind                Token
		TemplateRestriction *RestrictionSpec
		Modif               Token
		Type                Expr
		Decls               []Expr
		With                *WithSpec
	}

	FuncDecl struct {
		External Token
		Kind     Token
		Name     *Ident
		Modif    Token
		Params   *FormalPars
		RunsOn   *RunsOnSpec
		Mtc      *MtcSpec
		System   *SystemSpec
		Return   *ReturnSpec
		Body     *BlockStmt
		With     *WithSpec
	}

	SignatureDecl struct {
		Tok          Token
		Name         *Ident
		Params       *FormalPars
		NoBlock      Token
		Return       *ReturnSpec
		ExceptionTok Token
		Exception    *ParenExpr
		With         *WithSpec
	}

	SubTypeDecl struct {
		TypeTok Token
		Field   *Field
		With    *WithSpec
	}

	StructTypeDecl struct {
		TypeTok Token
		Kind    Token // RECORD, SET, UNION
		Name    Expr
		LBrace  Token
		Fields  []*Field
		RBrace  Token
		With    *WithSpec
	}

	EnumTypeDecl struct {
		TypeTok Token
		EnumTok Token
		Name    Expr
		LBrace  Token
		Enums   []Expr
		RBrace  Token
		With    *WithSpec
	}

	BehaviourTypeDecl struct {
		TypeTok Token
		Kind    Token
		Name    Expr
		Params  *FormalPars
		RunsOn  *RunsOnSpec
		System  *SystemSpec
		Return  *ReturnSpec
		With    *WithSpec
	}

	PortTypeDecl struct {
		TypeTok  Token
		PortTok  Token
		Name     Expr
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
		ExtendsTok Token
		Extends    []Expr
		Body       *BlockStmt
		With       *WithSpec
	}
)

func (x *ValueDecl) Pos() Pos { return x.Kind.Pos() }

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

func (x *ValueDecl) declNode()         {}
func (x *FuncDecl) declNode()          {}
func (x *SignatureDecl) declNode()     {}
func (x *SubTypeDecl) declNode()       {}
func (x *StructTypeDecl) declNode()    {}
func (x *EnumTypeDecl) declNode()      {}
func (x *BehaviourTypeDecl) declNode() {}
func (x *PortTypeDecl) declNode()      {}
func (x *PortAttribute) declNode()     {}
func (x *PortMapAttribute) declNode()  {}
func (x *ComponentTypeDecl) declNode() {}

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
		Tok Token
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

func (x *LanguageSpec) Pos() Pos    { return x.Tok.Pos() }
func (x *RestrictionSpec) Pos() Pos { return x.Tok.Pos() }
func (x *RunsOnSpec) Pos() Pos      { return x.RunsTok.Pos() }
func (x *SystemSpec) Pos() Pos      { return x.Tok.Pos() }
func (x *MtcSpec) Pos() Pos         { return x.Tok.Pos() }
func (x *ReturnSpec) Pos() Pos      { return x.Tok.Pos() }
func (x *FormalPars) Pos() Pos      { return x.LParen.Pos() }

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

func (x *LanguageSpec) End() Pos    { return x.List[len(x.List)-1].End() }
func (x *RestrictionSpec) End() Pos { return x.Tok.End() }
func (x *RunsOnSpec) End() Pos      { return x.Comp.End() }
func (x *SystemSpec) End() Pos      { return x.Comp.End() }
func (x *MtcSpec) End() Pos         { return x.Comp.End() }
func (x *ReturnSpec) End() Pos      { return x.Type.End() }
func (x *FormalPars) End() Pos      { return x.RParen.End() }
func (x *FormalPar) End() Pos       { return x.Name.End() }
func (x *WithSpec) End() Pos        { return x.RBrace.End() }
func (x *WithStmt) End() Pos        { return x.Value.End() }
