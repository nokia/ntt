package syntax

// ----------------------------------------------------------------------------
// Interfaces
//

// All node types implement the Node interface.
type Node interface {
	Pos() Pos
	//End() Pos
}

// All expression nodes implement the Expr interface.
type Expr interface {
	Node
}

// All statement nodes implement the Stmt interface.
type Stmt interface {
	Node
}

// All declaration nodes implement the Decl interface.
type Decl interface {
	Node
}

// Tokens
// ------------------------------------------------------------------------

type Token struct {
	pos  Pos
	Kind Kind
	Lit  string
}

func (x *Token) Pos() Pos { return x.pos }

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
		Value         Expr
		ParamTok      Token
		Param         Expr
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
		RBRace    Token
	}
)

func (x *Ident) Pos() Pos             { return x.Tok.Pos() }
func (x *ParametrizedIdent) Pos() Pos { return x.Ident.Pos() }
func (x *ValueLiteral) Pos() Pos      { return x.Tok.Pos() }
func (x *CompositeLiteral) Pos() Pos  { return x.LBrace.Pos() }
func (x *UnaryExpr) Pos() Pos         { return x.Op.Pos() }
func (x *BinaryExpr) Pos() Pos        { return x.Pos() }
func (x *ParenExpr) Pos() Pos         { return x.LParen.Pos() }
func (x *SelectorExpr) Pos() Pos      { return x.Pos() }
func (x *IndexExpr) Pos() Pos         { return x.Pos() }
func (x *CallExpr) Pos() Pos          { return x.Fun.Pos() }
func (x *LengthExpr) Pos() Pos        { return x.Pos() }
func (x *RedirectExpr) Pos() Pos      { return x.Pos() }
func (x *ValueExpr) Pos() Pos         { return x.Pos() }
func (x *ParamExpr) Pos() Pos         { return x.Pos() }
func (x *FromExpr) Pos() Pos          { return x.Kind.Pos() }
func (x *ModifiesExpr) Pos() Pos      { return x.Tok.Pos() }
func (x *RegexpExpr) Pos() Pos        { return x.Tok.Pos() }
func (x *PatternExpr) Pos() Pos       { return x.Tok.Pos() }
func (x *DecmatchExpr) Pos() Pos      { return x.Tok.Pos() }
func (x *DecodedExpr) Pos() Pos       { return x.Tok.Pos() }
func (x *DefSelectorExpr) Pos() Pos   { return x.Kind.Pos() }

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
		Tok   Token
		Block *BlockStmt
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

// ------------------------------------------------------------------------
// Declarations and Types

type TypeSpec interface {
	Node
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
		Params   Expr
		RunsOn   Expr
		Mtc      Expr
		System   Expr
		Return   *ReturnSpec
		Body     *BlockStmt
		With     *WithSpec
	}

	SignatureDecl struct {
		Tok          Token
		Name         *Ident
		Params       Expr
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
