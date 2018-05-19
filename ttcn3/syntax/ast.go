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

// ----------------------------------------------------------------------------
// Expressions
//
type (
	Ident struct {
		NamePos Pos
		Name    string
	}

	ParametrizedIdent struct {
		Ident  *Ident
		Params Expr
	}

	ValueLiteral struct {
		Kind     Token
		ValuePos Pos
		Value    string
	}

	CompositeLiteral struct {
	}

	UnaryExpr struct {
		Op    Token
		OpPos Pos
		X     Expr
	}

	BinaryExpr struct {
		X     Expr
		Op    Token
		OpPos Pos
		Y     Expr
	}

	ParenExpr struct {
		List []Expr
	}

	SelectorExpr struct {
		X   Expr
		Sel Expr
	}

	IndexExpr struct {
		X     Expr
		Index Expr
	}

	CallExpr struct {
		Fun  Expr
		Args []Expr
	}

	ModifiesExpr struct {
	}

	LengthExpr struct {
		Pos Pos
		X   Expr
		Len Expr
	}

	RedirectExpr struct {
		X            Expr
		Pos          Pos
		ValuePos     Pos
		Value        Expr
		ParamPos     Pos
		Param        Expr
		SenderPos    Pos
		Sender       Expr
		IndexPos     Pos
		Index        Expr
		TimestampPos Pos
		Timestamp    Expr
	}

	RegexpExpr struct {
	}

	PatternExpr struct {
	}

	DecmatchExpr struct {
	}

	DecodedExpr struct {
	}
)

// ----------------------------------------------------------------------------
// Statements

type (
	BlockStmt struct {
		LBrace Pos
		Stmts  []Stmt
		RBrace Pos
	}

	DeclStmt struct {
		Decl Decl
	}

	ExprStmt struct {
		Expr Expr
	}

	ForStmt struct {
	}

	WhileStmt struct {
	}

	DoWhileStmt struct {
	}

	IfStmt struct {
	}

	SelectStmt struct {
	}

	CaseSpec struct {
	}

	AltStmt struct {
	}
)

// ----------------------------------------------------------------------------
// Declarations

type (
	ValueDecl struct {
		DeclPos Pos
		Kind    Token // VAR, CONST, MODULEPAT, TIMER, ...
		Type    Expr
		Decls   []Expr
	}

	FuncDecl struct {
		FuncPos Pos
		Kind    Token
		Name    *Ident
		Params  Expr
		Return  Expr
		RunsOn  Expr
		Mtc     Expr
		System  Expr
		Extern  bool
		Body    *BlockStmt
	}

	SignatureDecl struct {
	}
)

// ----------------------------------------------------------------------------
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

// ----------------------------------------------------------------------------
// Modules and Module Definitions

type (
	Module struct {
		Module Pos
		Name   *Ident
		Decls  []Decl
	}

	ModuleDef struct {
	}

	GroupDecl struct {
	}

	FriendDecl struct {
	}

	ImportDecl struct {
		ImportPos   Pos
		Module      *Ident
		ImportSpecs []ImportSpec
	}

	ImportSpec struct {
	}

	ImportStmt struct {
	}

	ExceptSpec struct {
	}

	ExceptStmt struct {
	}
)

// ----------------------------------------------------------------------------
// Miscellaneous

type (
	LanguageSpec struct {
	}

	RestrictionSpec struct {
		Kind    Token
		KindPos Pos
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
	}

	WithStmt struct {
	}
)
