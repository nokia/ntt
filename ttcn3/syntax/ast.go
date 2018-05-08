package syntax

import (
	"strings"
)

// ----------------------------------------------------------------------------
// Interfaces
//
// There are 3 main classes of nodes: Expressions and type nodes,
// statement nodes, and declaration nodes. The node names usually
// match the corresponding Go spec production names to which they
// correspond. The node fields correspond to the individual parts
// of the respective productions.
//
// All nodes contain position information marking the beginning of
// the corresponding source text segment; it is accessible via the
// Pos accessor method. Nodes may contain additional position info
// for language constructs where comments may be found between parts
// of the construct (typically any larger, parenthesized subpart).
// That position information is needed to properly position comments
// when printing the construct.

// All node types implement the Node interface.
type Node interface {
	Pos() Pos // position of first character belonging to the node
	End() Pos // position of first character immediately after the node
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

// ----------------------------------------------------------------------------
// Comments

// A Comment node represents a single //-style or /*-style comment.
type Comment struct {
	Slash Pos    // position of "/" starting the comment
	Text  string // comment text (excluding '\n' for //-style comments)
}

func (c *Comment) Pos() Pos { return c.Slash }
func (c *Comment) End() Pos { return Pos(int(c.Slash) + len(c.Text)) }

// A CommentGroup represents a sequence of comments
// with no other tokens and no empty lines between.
//
type CommentGroup struct {
	List []*Comment // len(List) > 0
}

func (g *CommentGroup) Pos() Pos { return g.List[0].Pos() }
func (g *CommentGroup) End() Pos { return g.List[len(g.List)-1].End() }

func isWhitespace(ch byte) bool { return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' }

func stripTrailingWhitespace(s string) string {
	i := len(s)
	for i > 0 && isWhitespace(s[i-1]) {
		i--
	}
	return s[0:i]
}

// Text returns the text of the comment.
// Comment markers (//, /*, and */), the first space of a line comment, and
// leading and trailing empty lines are removed. Multiple empty lines are
// reduced to one, and trailing space on lines is trimmed. Unless the result
// is empty, it is newline-terminated.
//
func (g *CommentGroup) Text() string {
	if g == nil {
		return ""
	}
	comments := make([]string, len(g.List))
	for i, c := range g.List {
		comments[i] = c.Text
	}

	lines := make([]string, 0, 10) // most comments are less than 10 lines
	for _, c := range comments {
		// Remove comment markers.
		// The parser has given us exactly the comment text.
		switch c[1] {
		case '/':
			//-style comment (no newline at the end)
			c = c[2:]
			// strip first space - required for Example tests
			if len(c) > 0 && c[0] == ' ' {
				c = c[1:]
			}
		case '*':
			/*-style comment */
			c = c[2 : len(c)-2]
		}

		// Split on newlines.
		cl := strings.Split(c, "\n")

		// Walk lines, stripping trailing white space and adding to list.
		for _, l := range cl {
			lines = append(lines, stripTrailingWhitespace(l))
		}
	}

	// Remove leading blank lines; convert runs of
	// interior blank lines to a single blank line.
	n := 0
	for _, line := range lines {
		if line != "" || n > 0 && lines[n-1] != "" {
			lines[n] = line
			n++
		}
	}
	lines = lines[0:n]

	// Add final "" entry to get trailing newline from Join.
	if n > 0 && lines[n-1] != "" {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// ----------------------------------------------------------------------------
// Miscellaneous and types

type (
	// A RestrictionSpec represents template restrictions
	RestrictionSpec struct {
		Kind    Token
		KindPos Pos
	}
)

// ----------------------------------------------------------------------------
// Expressions and types

// An expression is represented by a tree consisting of one
// or more of the following concrete expression nodes.
//
type (
	// A BadExpr node is a placeholder for expressions containing
	// syntax errors for which no correct expression nodes can be
	// created.
	//
	BadExpr struct {
		From, To Pos // position range of bad expression
	}

	// An Ident node represents an identifier.
	Ident struct {
		NamePos Pos    // identifier position
		Name    string // identifier name
	}

	ValueLiteral struct {
		Kind     Token
		ValuePos Pos
		Value    string
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

	SetExpr struct {
		List []Expr
	}

	SelectorExpr struct {
		X   Expr
		Sel *Ident
	}

	IndexExpr struct {
		X     Expr
		Index Expr
	}

	CallExpr struct {
		Fun  Expr
		Args []Expr
	}
)

// Pos and End implementations for expression/type nodes.

func (x *BadExpr) Pos() Pos      { return x.From }
func (x *Ident) Pos() Pos        { return x.NamePos }
func (x *ValueLiteral) Pos() Pos { return x.ValuePos }
func (x *UnaryExpr) Pos() Pos    { return x.X.Pos() }
func (x *BinaryExpr) Pos() Pos   { return x.X.Pos() }
func (x *SetExpr) Pos() Pos      { return x.List[0].Pos() }
func (x *SelectorExpr) Pos() Pos { return x.X.Pos() }
func (x *IndexExpr) Pos() Pos    { return x.X.Pos() }
func (x *CallExpr) Pos() Pos     { return x.Fun.Pos() }

func (x *BadExpr) End() Pos      { return x.To }
func (x *Ident) End() Pos        { return Pos(int(x.NamePos) + len(x.Name)) }
func (x *ValueLiteral) End() Pos { return x.ValuePos }
func (x *UnaryExpr) End() Pos    { return x.X.End() }
func (x *BinaryExpr) End() Pos   { return x.X.End() }
func (x *SetExpr) End() Pos      { return x.List[0].End() }
func (x *SelectorExpr) End() Pos { return x.X.End() }
func (x *IndexExpr) End() Pos    { return x.X.End() }
func (x *CallExpr) End() Pos     { return x.Fun.End() }

// exprNode() ensures that only expression/type nodes can be
// assigned to an Expr.
//
func (*BadExpr) exprNode()        {}
func (*Ident) exprNode()          {}
func (x *ValueLiteral) exprNode() {}
func (x *UnaryExpr) exprNode()    {}
func (x *BinaryExpr) exprNode()   {}
func (x *SetExpr) exprNode()      {}
func (x *SelectorExpr) exprNode() {}
func (x *IndexExpr) exprNode()    {}
func (x *CallExpr) exprNode()     {}

// ----------------------------------------------------------------------------
// Convenience functions for Idents

// NewIdent creates a new Ident without position.
// Useful for ASTs generated by code other than the Go
//
func NewIdent(name string) *Ident { return &Ident{NoPos, name} }

func (id *Ident) String() string {
	if id != nil {
		return id.Name
	}
	return "<nil>"
}

// -----------------------------------------------------------------------
// Types

type Field struct {
	Type     Expr
	Name     Expr
	Optional bool
	In       bool
	Out      bool
}

type FieldList struct {
	From   Pos
	Fields []*Field
}

func (x *Field) Pos() Pos     { return x.Type.Pos() }
func (x *FieldList) Pos() Pos { return x.From }

func (x *Field) End() Pos     { return x.Type.End() }
func (x *FieldList) End() Pos { return x.Fields[len(x.Fields)-1].End() }

func (x *Field) exprNode()     {}
func (x *FieldList) exprNode() {}

// ----------------------------------------------------------------------------
// Declarations

type (
	ImportDecl struct {
		ImportPos   Pos
		Module      *Ident
		ImportSpecs []ImportSpec
	}

	ImportSpec struct {
	}

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
		Params  *FieldList
		Return  Expr
		RunsOn  Expr
		Mtc     Expr
		System  Expr
		Extern  bool
		Body    *BlockStmt
	}

	SubType struct {
		TypePos Pos
		Name    *Ident
		Type    Expr
	}
)

func (x *ImportDecl) Pos() Pos { return x.ImportPos }
func (x *ValueDecl) Pos() Pos  { return x.DeclPos }
func (x *FuncDecl) Pos() Pos   { return x.FuncPos }
func (x *SubType) Pos() Pos    { return x.TypePos }

func (x *ImportDecl) End() Pos { return x.Module.End() }
func (x *ValueDecl) End() Pos  { return x.Decls[len(x.Decls)-1].End() }
func (x *FuncDecl) End() Pos   { return x.Body.End() }
func (x *SubType) End() Pos    { return x.Name.End() }

func (x *ImportDecl) declNode() {}
func (x *ValueDecl) declNode()  {}
func (x *FuncDecl) declNode()   {}
func (x *SubType) declNode()    {}

// -----------------------------------------------------------------------
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
)

func (x *BlockStmt) Pos() Pos { return x.LBrace }
func (x *DeclStmt) Pos() Pos  { return x.Decl.Pos() }
func (x *ExprStmt) Pos() Pos  { return x.Expr.Pos() }

func (x *BlockStmt) End() Pos { return x.RBrace }
func (x *DeclStmt) End() Pos  { return x.Decl.End() }
func (x *ExprStmt) End() Pos  { return x.Expr.End() }

func (x *BlockStmt) stmtNode() {}
func (x *DeclStmt) stmtNode()  {}
func (x *ExprStmt) stmtNode()  {}

// ----------------------------------------------------------------------------
// Modules

type Module struct {
	Doc      *CommentGroup   // associated documentation; or nil
	Module   Pos             // position of "module" keyword
	Name     *Ident          // module name
	Decls    []Decl          // top-level declarations; or nil
	Comments []*CommentGroup // list of all comments in the source file
}

func (f *Module) Pos() Pos { return f.Module }
func (f *Module) End() Pos {
	if n := len(f.Decls); n > 0 {
		return f.Decls[n-1].End()
	}
	return f.Name.End()
}
