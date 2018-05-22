package syntax

// ----------------------------------------------------------------------------
// Interfaces
//

// All node types implement the Node interface.
type Node interface {
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
	Pos  Pos
	Kind Kind
	Lit  string
}

// ------------------------------------------------------------------------
// Expressions
//
type (
	Ident struct {
		Tok Token
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
		Args []Expr
	}

	LengthExpr struct {
		X    Expr
		Len  Token
		Size Expr
	}

	RedirectExpr struct {
		X            Expr
		Tok          Token
		ValueTok     Token
		Value        Expr
		ParamTok     Token
		Param        Expr
		SenderTok    Token
		Sender       Expr
		IndexTok     Token
		Index        Expr
		TimestampTok Token
		Timestamp    Expr
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

	// ("any"|"all") "from" Expr
	FromExpr struct {
		Kind Token
		X    Expr
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
)

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
		Kind  Token
		Type  Expr
		Decls []Expr
		With  *WithSpec
	}

	FuncDecl struct {
		Kind   Token
		Name   *Ident
		Params Expr
		Return Expr
		RunsOn Expr
		Mtc    Expr
		System Expr
		Extern bool
		Body   *BlockStmt
		With   *WithSpec
	}

	SignatureDecl struct {
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
		ImportTok   Token
		FromTok     Token
		Module      *Ident
		Language    *LanguageSpec
		ImportSpecs []ImportSpec
		With        *WithSpec
	}

	ImportSpec struct {
	}

	ImportStmt struct {
	}

	ExceptSpec struct {
	}

	ExceptStmt struct {
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
		List []*FormalPar
	}

	FormalPar struct {
		Type Expr
		Name Expr
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
