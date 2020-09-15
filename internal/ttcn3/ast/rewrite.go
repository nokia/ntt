package ast

import (
	"fmt"
	"reflect"
)

// An ApplyFunc is invoked by Apply for each node n, even if n is nil, before
// and/or after the node's children, using a Cursor describing the current node
// and providing operations on it.
//
// The return value of ApplyFunc controls the syntax tree traversal.  See Apply
// for details.
type ApplyFunc func(*Cursor) bool

// Apply traverses a syntax tree recursively, starting with root, and calling
// pre and post for each node as described below.  Apply returns the syntax
// tree, possibly modified.
//
// If pre is not nil, it is called for each node before the node's children are
// traversed (pre-order). If pre returns false, no children are traversed, and
// post is not called for that node.
//
// If post is not nil, and a prior call of pre didn't return false, post is
// called for each node after its children are traversed (post-order). If post
// returns false, traversal is terminated and Apply returns immediately.
//
// Only fields that refer to AST nodes are considered children; i.e., token.Pos,
// Scopes, Objects, and fields of basic types (strings, etc.) are ignored.
//
// Children are traversed in the order in which they appear in the respective
// node's struct definition. A package's files are traversed in the filenames'
// alphabetical order.
//
func Apply(root Node, pre, post ApplyFunc) (result Node) {
	parent := &struct{ Node }{root}
	defer func() {
		if r := recover(); r != nil && r != abort {
			panic(r)
		}
		result = parent.Node
	}()
	a := &application{pre: pre, post: post}
	a.apply(parent, "Node", nil, root)
	return
}

var abort = new(int) // singleton, to signal termination of Apply

// A Cursor describes a node encountered during Apply.  Information about the
// node and its parent is available from the Node, Parent, Name, and Index
// methods.
//
// If p is a variable of type and value of the current parent node c.Parent(),
// and f is the field identifier with name c.Name(), the following invariants
// hold:
//
//   p.f            == c.Node()  if c.Index() <  0 p.f[c.Index()] == c.Node()
//   if c.Index() >= 0
//
// The methods Replace, Delete, InsertBefore, and InsertAfter can be used to
// change the AST without disrupting Apply.
type Cursor struct {
	parent Node
	name   string
	iter   *iterator // valid if non-nil
	node   Node
}

// Node returns the current Node.
func (c *Cursor) Node() Node { return c.node }

// Parent returns the parent of the current Node.
func (c *Cursor) Parent() Node { return c.parent }

// Name returns the name of the parent Node field that contains the current
// Node.  If the parent is a *Package and the current Node is a *File,
// Name returns the filename for the current Node.
func (c *Cursor) Name() string { return c.name }

// Index reports the index >= 0 of the current Node in the slice of Nodes that
// contains it, or a value < 0 if the current Node is not part of a slice.  The
// index of the current node changes if InsertBefore is called while processing
// the current node.
func (c *Cursor) Index() int {
	if c.iter != nil {
		return c.iter.index
	}
	return -1
}

// field returns the current node's parent field value.
func (c *Cursor) field() reflect.Value {
	return reflect.Indirect(reflect.ValueOf(c.parent)).FieldByName(c.name)
}

// Replace replaces the current Node with n.
// The replacement node is not walked by Apply.
func (c *Cursor) Replace(n Node) {
	v := c.field()
	if i := c.Index(); i >= 0 {
		v = v.Index(i)
	}
	v.Set(reflect.ValueOf(n))
}

// Delete deletes the current Node from its containing slice.  If the current
// Node is not part of a slice, Delete panics.  As a special case, if the
// current node is a package file, Delete removes it from the package's Files
// map.
func (c *Cursor) Delete() {
	i := c.Index()
	if i < 0 {
		panic("Delete node not contained in slice")
	}
	v := c.field()
	l := v.Len()
	reflect.Copy(v.Slice(i, l), v.Slice(i+1, l))
	v.Index(l - 1).Set(reflect.Zero(v.Type().Elem()))
	v.SetLen(l - 1)
	c.iter.step--
}

// InsertAfter inserts n after the current Node in its containing slice.  If the
// current Node is not part of a slice, InsertAfter panics.  Apply does not walk
// n.
func (c *Cursor) InsertAfter(n Node) {
	i := c.Index()
	if i < 0 {
		panic("InsertAfter node not contained in slice")
	}
	v := c.field()
	v.Set(reflect.Append(v, reflect.Zero(v.Type().Elem())))
	l := v.Len()
	reflect.Copy(v.Slice(i+2, l), v.Slice(i+1, l))
	v.Index(i + 1).Set(reflect.ValueOf(n))
	c.iter.step++
}

// InsertBefore inserts n before the current Node in its containing slice.
// If the current Node is not part of a slice, InsertBefore panics.
// Apply will not walk n.
func (c *Cursor) InsertBefore(n Node) {
	i := c.Index()
	if i < 0 {
		panic("InsertBefore node not contained in slice")
	}
	v := c.field()
	v.Set(reflect.Append(v, reflect.Zero(v.Type().Elem())))
	l := v.Len()
	reflect.Copy(v.Slice(i+1, l), v.Slice(i, l))
	v.Index(i).Set(reflect.ValueOf(n))
	c.iter.index++
}

// application carries all the shared data so we can pass it around cheaply.
type application struct {
	pre, post ApplyFunc
	cursor    Cursor
	iter      iterator
}

func (a *application) apply(parent Node, name string, iter *iterator, n Node) {
	// convert typed nil into untyped nil
	if v := reflect.ValueOf(n); v.Kind() == reflect.Ptr && v.IsNil() {
		n = nil
	}

	// avoid heap-allocating a new cursor for each apply call; reuse a.cursor instead
	saved := a.cursor
	a.cursor.parent = parent
	a.cursor.name = name
	a.cursor.iter = iter
	a.cursor.node = n

	if a.pre != nil && !a.pre(&a.cursor) {
		a.cursor = saved
		return
	}

	// walk children
	// (the order of the cases matches the order of the corresponding node types)
	switch n := n.(type) {
	case nil:
		// nothing to do

	case *ErrorNode:
		// nothing to do

	case *Ident:
		// nothing to do

	case *ParametrizedIdent:
		a.apply(n, "Ident", nil, n.Ident)
		a.apply(n, "Params", nil, n.Params)

	case *ValueLiteral:

	case *CompositeLiteral:
		a.applyList(n, "List")

	case *UnaryExpr:
		a.apply(n, "X", nil, n.X)

	case *BinaryExpr:
		a.apply(n, "X", nil, n.X)
		a.apply(n, "Y", nil, n.Y)

	case *ParenExpr:
		a.applyList(n, "List")

	case *SelectorExpr:
		a.apply(n, "X", nil, n.X)
		a.apply(n, "Sel", nil, n.Sel)

	case *IndexExpr:
		a.apply(n, "X", nil, n.X)
		a.apply(n, "Index", nil, n.Index)

	case *CallExpr:
		a.apply(n, "Fun", nil, n.Fun)
		a.apply(n, "Args", nil, n.Args)

	case *LengthExpr:
		a.apply(n, "X", nil, n.X)
		a.apply(n, "Size", nil, n.Size)

	case *RedirectExpr:
		a.apply(n, "X", nil, n.X)
		a.applyList(n, "Value")
		a.applyList(n, "Param")
		a.apply(n, "Sender", nil, n.Sender)
		a.apply(n, "Index", nil, n.Index)
		a.apply(n, "Timestamp", nil, n.Timestamp)

	case *ValueExpr:
		a.apply(n, "X", nil, n.X)
		a.apply(n, "Y", nil, n.Y)

	case *ParamExpr:
		a.apply(n, "X", nil, n.X)
		a.apply(n, "Y", nil, n.Y)

	case *FromExpr:
		a.apply(n, "X", nil, n.X)

	case *ModifiesExpr:
		a.apply(n, "X", nil, n.X)
		a.apply(n, "Y", nil, n.Y)

	case *RegexpExpr:
		a.apply(n, "X", nil, n.X)

	case *PatternExpr:
		a.apply(n, "X", nil, n.X)

	case *DecmatchExpr:
		a.apply(n, "Params", nil, n.Params)
		a.apply(n, "X", nil, n.X)

	case *DecodedExpr:
		a.apply(n, "Params", nil, n.Params)
		a.apply(n, "X", nil, n.X)

	case *DefKindExpr:
		a.applyList(n, "List")

	case *ExceptExpr:
		a.apply(n, "X", nil, n.X)
		a.applyList(n, "List")

	// Statements
	// -------------------------

	case *BlockStmt:
		a.applyList(n, "Stmts")

	case *DeclStmt:
		a.apply(n, "Decl", nil, n.Decl)

	case *ExprStmt:
		a.apply(n, "Expr", nil, n.Expr)

	case *BranchStmt:

	case *ReturnStmt:
		a.apply(n, "Result", nil, n.Result)

	case *AltStmt:
		a.apply(n, "Body", nil, n.Body)

	case *CallStmt:
		a.apply(n, "Stmt", nil, n.Stmt)
		a.apply(n, "Body", nil, n.Body)

	case *ForStmt:
		a.apply(n, "Init", nil, n.Init)
		a.apply(n, "Cond", nil, n.Cond)
		a.apply(n, "Post", nil, n.Post)
		a.apply(n, "Body", nil, n.Body)

	case *WhileStmt:
		a.apply(n, "Cond", nil, n.Cond)
		a.apply(n, "Body", nil, n.Body)

	case *DoWhileStmt:
		a.apply(n, "Body", nil, n.Body)
		a.apply(n, "Cond", nil, n.Cond)

	case *IfStmt:
		a.apply(n, "Cond", nil, n.Cond)
		a.apply(n, "Then", nil, n.Then)
		a.apply(n, "Else", nil, n.Else)

	case *SelectStmt:
		a.apply(n, "Tag", nil, n.Tag)
		a.applyList(n, "Body")

	case *CaseClause:
		a.apply(n, "Case", nil, n.Case)
		a.apply(n, "Body", nil, n.Body)

	case *CommClause:
		a.apply(n, "X", nil, n.X)
		a.apply(n, "Comm", nil, n.Comm)
		a.apply(n, "Body", nil, n.Body)

	// TypeSpecs
	// ----------------

	case *Field:
		a.apply(n, "Type", nil, n.Type)
		a.apply(n, "Name", nil, n.Name)
		a.applyList(n, "ArrayDef")
		a.apply(n, "TypePars", nil, n.TypePars)
		a.apply(n, "ValueConstraint", nil, n.ValueConstraint)
		a.apply(n, "LengthConstraint", nil, n.LengthConstraint)

	case *RefSpec:
		a.apply(n, "X", nil, n.X)

	case *StructSpec:
		a.applyList(n, "Fields")

	case *ListSpec:
		a.apply(n, "Length", nil, n.Length)
		a.apply(n, "ElemType", nil, n.ElemType)

	case *EnumSpec:
		a.applyList(n, "Enums")

	case *BehaviourSpec:
		a.apply(n, "Params", nil, n.Params)
		a.apply(n, "RunsOn", nil, n.RunsOn)
		a.apply(n, "System", nil, n.System)
		a.apply(n, "Return", nil, n.Return)

	// Declarations
	// ----------------------
	case *ValueDecl:
		a.apply(n, "TemplateRestriction", nil, n.TemplateRestriction)
		a.apply(n, "Type", nil, n.Type)
		a.applyList(n, "Decls")
		a.apply(n, "With", nil, n.With)

	case *TemplateDecl:
		a.apply(n, "Type", nil, n.Type)
		a.apply(n, "Name", nil, n.Name)
		a.apply(n, "TypePars", nil, n.TypePars)
		a.apply(n, "Params", nil, n.Params)
		a.apply(n, "Base", nil, n.Base)
		a.apply(n, "Value", nil, n.Value)
		a.apply(n, "With", nil, n.With)

	case *ModuleParameterGroup:
		a.applyList(n, "Decls")
		a.apply(n, "With", nil, n.With)

	case *FuncDecl:
		a.apply(n, "Name", nil, n.Name)
		a.apply(n, "TypePars", nil, n.TypePars)
		a.apply(n, "Params", nil, n.Params)
		a.apply(n, "RunsOn", nil, n.RunsOn)
		a.apply(n, "Mtc", nil, n.Mtc)
		a.apply(n, "System", nil, n.System)
		a.apply(n, "Return", nil, n.Return)
		a.apply(n, "Body", nil, n.Body)
		a.apply(n, "With", nil, n.With)

	case *SignatureDecl:
		a.apply(n, "Name", nil, n.Name)
		a.apply(n, "TypePars", nil, n.TypePars)
		a.apply(n, "Params", nil, n.Params)
		a.apply(n, "Return", nil, n.Return)
		a.apply(n, "Exception", nil, n.Exception)
		a.apply(n, "With", nil, n.With)

	case *SubTypeDecl:
		a.apply(n, "Field", nil, n.Field)
		a.apply(n, "With", nil, n.With)

	case *StructTypeDecl:
		a.apply(n, "Name", nil, n.Name)
		a.apply(n, "TypePars", nil, n.TypePars)
		a.applyList(n, "Fields")
		a.apply(n, "With", nil, n.With)

	case *EnumTypeDecl:
		a.apply(n, "Name", nil, n.Name)
		a.apply(n, "TypePars", nil, n.TypePars)
		a.applyList(n, "Enums")
		a.apply(n, "With", nil, n.With)

	case *BehaviourTypeDecl:
		a.apply(n, "Name", nil, n.Name)
		a.apply(n, "TypePars", nil, n.TypePars)
		a.apply(n, "Params", nil, n.Params)
		a.apply(n, "RunsOn", nil, n.RunsOn)
		a.apply(n, "System", nil, n.System)
		a.apply(n, "Return", nil, n.Return)
		a.apply(n, "With", nil, n.With)

	case *PortTypeDecl:
		a.apply(n, "Name", nil, n.Name)
		a.apply(n, "TypePars", nil, n.TypePars)
		a.applyList(n, "Attrs")
		a.apply(n, "With", nil, n.With)

	case *PortAttribute:
		a.applyList(n, "Types")

	case *PortMapAttribute:
		a.apply(n, "Params", nil, n.Params)

	case *ComponentTypeDecl:
		a.apply(n, "Name", nil, n.Name)
		a.apply(n, "TypePars", nil, n.TypePars)
		a.applyList(n, "Extends")
		a.apply(n, "Body", nil, n.Body)
		a.apply(n, "With", nil, n.With)

	case *Module:
		a.apply(n, "Name", nil, n.Name)
		a.apply(n, "Language", nil, n.Language)
		a.applyList(n, "Defs")
		a.apply(n, "With", nil, n.With)

	case *ModuleDef:
		a.apply(n, "Def", nil, n.Def)

	case *ControlPart:
		a.apply(n, "Body", nil, n.Body)
		a.apply(n, "With", nil, n.With)

	case *ImportDecl:
		a.apply(n, "Module", nil, n.Module)
		a.apply(n, "Language", nil, n.Language)
		a.applyList(n, "List")
		a.apply(n, "With", nil, n.With)

	case *GroupDecl:
		a.apply(n, "Name", nil, n.Name)
		a.applyList(n, "Defs")
		a.apply(n, "With", nil, n.With)

	case *FriendDecl:
		a.apply(n, "Module", nil, n.Module)
		a.apply(n, "With", nil, n.With)

	// Misc
	// -------

	case *LanguageSpec:
		// nothing to do

	case *RestrictionSpec:

	case *RunsOnSpec:
		a.apply(n, "Comp", nil, n.Comp)

	case *SystemSpec:
		a.apply(n, "Comp", nil, n.Comp)

	case *MtcSpec:
		a.apply(n, "Comp", nil, n.Comp)

	case *ReturnSpec:
		a.apply(n, "Restriction", nil, n.Restriction)
		a.apply(n, "Type", nil, n.Type)

	case *FormalPars:
		a.applyList(n, "List")

	case *FormalPar:
		a.apply(n, "TemplateRestriction", nil, n.TemplateRestriction)
		a.apply(n, "Type", nil, n.Type)
		a.apply(n, "Name", nil, n.Name)

	case *WithSpec:
		a.applyList(n, "List")

	case *WithStmt:
		a.applyList(n, "List")
		a.apply(n, "Value", nil, n.Value)
	default:
		panic(fmt.Sprintf("Apply: unexpected node type %T", n))
	}

	if a.post != nil && !a.post(&a.cursor) {
		panic(abort)
	}

	a.cursor = saved
}

// An iterator controls iteration over a slice of nodes.
type iterator struct {
	index, step int
}

func (a *application) applyList(parent Node, name string) {
	// avoid heap-allocating a new iterator for each applyList call; reuse a.iter instead
	saved := a.iter
	a.iter.index = 0
	for {
		// must reload parent.name each time, since cursor modifications might change it
		v := reflect.Indirect(reflect.ValueOf(parent)).FieldByName(name)
		if a.iter.index >= v.Len() {
			break
		}

		// element x may be nil in a bad AST - be cautious
		var x Node

		if e := v.Index(a.iter.index); e.IsValid() {
			x, _ = e.Interface().(Node)
		}

		a.iter.step = 1
		a.apply(parent, name, &a.iter, x)
		a.iter.index += a.iter.step
	}
	a.iter = saved
}
