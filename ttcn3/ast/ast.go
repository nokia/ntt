package ast

import (
	"github.com/nokia/ntt/ttcn3/token"
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
	Pos() token.Pos // position of first character belonging to the node
	End() token.Pos // position of first character immediately after the node
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
	Slash token.Pos // position of "/" starting the comment
	Text  string    // comment text (excluding '\n' for //-style comments)
}

func (c *Comment) Pos() token.Pos { return c.Slash }
func (c *Comment) End() token.Pos { return token.Pos(int(c.Slash) + len(c.Text)) }

// A CommentGroup represents a sequence of comments
// with no other tokens and no empty lines between.
//
type CommentGroup struct {
	List []*Comment // len(List) > 0
}

func (g *CommentGroup) Pos() token.Pos { return g.List[0].Pos() }
func (g *CommentGroup) End() token.Pos { return g.List[len(g.List)-1].End() }

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
		From, To token.Pos // position range of bad expression
	}

	// An Ident node represents an identifier.
	Ident struct {
		NamePos token.Pos // identifier position
		Name    string    // identifier name
		Obj     *Object   // denoted object; or nil
	}

	ValueLiteral struct {
		Kind     token.Token
		ValuePos token.Pos
		Value    string
	}

	UnaryExpr struct {
		Op    token.Token
		OpPos token.Pos
		X     Expr
	}

	BinaryExpr struct {
		X     Expr
		Op    token.Token
		OpPos token.Pos
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

func (x *BadExpr) Pos() token.Pos      { return x.From }
func (x *Ident) Pos() token.Pos        { return x.NamePos }
func (x *ValueLiteral) Pos() token.Pos { return x.ValuePos }
func (x *UnaryExpr) Pos() token.Pos    { return x.X.Pos() }
func (x *BinaryExpr) Pos() token.Pos   { return x.X.Pos() }
func (x *SetExpr) Pos() token.Pos      { return x.List[0].Pos() }
func (x *SelectorExpr) Pos() token.Pos { return x.X.Pos() }
func (x *IndexExpr) Pos() token.Pos    { return x.X.Pos() }
func (x *CallExpr) Pos() token.Pos     { return x.Fun.Pos() }

func (x *BadExpr) End() token.Pos      { return x.To }
func (x *Ident) End() token.Pos        { return token.Pos(int(x.NamePos) + len(x.Name)) }
func (x *ValueLiteral) End() token.Pos { return x.ValuePos }
func (x *UnaryExpr) End() token.Pos    { return x.X.End() }
func (x *BinaryExpr) End() token.Pos   { return x.X.End() }
func (x *SetExpr) End() token.Pos      { return x.List[0].End() }
func (x *SelectorExpr) End() token.Pos { return x.X.End() }
func (x *IndexExpr) End() token.Pos    { return x.X.End() }
func (x *CallExpr) End() token.Pos     { return x.Fun.End() }

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
// Useful for ASTs generated by code other than the Go parser.
//
func NewIdent(name string) *Ident { return &Ident{token.NoPos, name, nil} }

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
	From   token.Pos
	Fields []*Field
}

type SubType struct {
	TypePos token.Pos
	Name    *Ident
	Type    Expr
}

func (x *Field) Pos() token.Pos     { return x.Type.Pos() }
func (x *FieldList) Pos() token.Pos { return x.From }
func (x *SubType) Pos() token.Pos   { return x.TypePos }

func (x *Field) End() token.Pos     { return x.Type.End() }
func (x *FieldList) End() token.Pos { return x.Fields[len(x.Fields)-1].End() }
func (x *SubType) End() token.Pos   { return x.Name.End() }

func (x *Field) exprNode()     {}
func (x *FieldList) exprNode() {}
func (x *SubType) exprNode()   {}

// ----------------------------------------------------------------------------
// Declarations

type (
	ValueDecl struct {
		DeclPos token.Pos
		Kind    token.Token // VAR, CONST, MODULEPAT, TIMER, ...
		Type    *Ident
		Decls   []Expr
	}

	FuncDecl struct {
		FuncPos token.Pos
		Kind    token.Token
		Name    *Ident
		Params  *FieldList
		Return  Expr
		RunsOn  Expr
		Mtc     Expr
		System  Expr
		Extern  bool
		Body    *BlockStmt
	}
)

func (x *ValueDecl) Pos() token.Pos { return x.DeclPos }
func (x *FuncDecl) Pos() token.Pos  { return x.FuncPos }

func (x *ValueDecl) End() token.Pos { return x.Decls[len(x.Decls)-1].End() }
func (x *FuncDecl) End() token.Pos  { return x.Body.End() }

func (x *ValueDecl) declNode() {}
func (x *FuncDecl) declNode()  {}

// -----------------------------------------------------------------------
// Statements

type (
	BlockStmt struct {
		LBrace token.Pos
		Stmts  []Stmt
		RBrace token.Pos
	}

	DeclStmt struct {
		Decl Decl
	}

	ExprStmt struct {
		Expr Expr
	}
)

func (x *BlockStmt) Pos() token.Pos { return x.LBrace }
func (x *DeclStmt) Pos() token.Pos  { return x.Decl.Pos() }
func (x *ExprStmt) Pos() token.Pos  { return x.Expr.Pos() }

func (x *BlockStmt) End() token.Pos { return x.RBrace }
func (x *DeclStmt) End() token.Pos  { return x.Decl.End() }
func (x *ExprStmt) End() token.Pos  { return x.Expr.End() }

func (x *BlockStmt) stmtNode() {}
func (x *DeclStmt) stmtNode()  {}
func (x *ExprStmt) stmtNode()  {}

// ----------------------------------------------------------------------------
// Modules

type Module struct {
	Doc    *CommentGroup // associated documentation; or nil
	Module token.Pos     // position of "module" keyword
	Name   *Ident        // module name
	Decls  []Decl        // top-level declarations; or nil
	Scope  *Scope        // module scope
	//Imports    []*ImportSpec   // imports in this file
	Unresolved []*Ident        // unresolved identifiers in this file
	Comments   []*CommentGroup // list of all comments in the source file
}

func (f *Module) Pos() token.Pos { return f.Module }
func (f *Module) End() token.Pos {
	if n := len(f.Decls); n > 0 {
		return f.Decls[n-1].End()
	}
	return f.Name.End()
}
