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

// All node types implement the Node interface.
type Type interface {
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
// Declarations

type (
	ValueDecl struct {
		Kind  Token
		Type  Expr
		Decls []Expr
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
	}

	SignatureDecl struct {
	}
)

// ------------------------------------------------------------------------
// Types

type (
	TypeDecl struct {
	}

	SubType struct {
	}

	ListType struct {
	}

	StructType struct {
	}

	Field struct {
	}

	EnumType struct {
	}

	BehaviourType struct {
	}

	ComponentType struct {
	}

	PortAttribute struct {
	}

	PortType struct {
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
	}

	SystemSpec struct {
	}

	MtcSpec struct {
	}

	ReturnSpec struct {
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
