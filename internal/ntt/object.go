package ntt

import (
	"fmt"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/token"
)

type ObjFlags uint32

const (
	ObjVar ObjFlags = 1 << iota
	ObjConst
	ObjTemplate
	ObjModulepar
	ObjExternal
	ObjTest
	ObjAltstep
	//ObjTimer
	//ObjPort
)

type ObjKind uint32

const (
	ObjUnknownKind ObjKind = iota
	ObjDecl
	ObjFunc
	ObjModule
	ObjImport
)

// Object describes a named language entity, such as functions or consts.
type Object interface {
	Name() string // Object name.

	// Returns the scope the object is embedded in.
	Parent() Scope

	// SetParent sets the scope the object is embedded in.
	SetParent(s Scope)

	Pos() loc.Pos
	End() loc.Pos
}

type object struct {
	pos, end loc.Pos
	name     string
	kind     ObjKind
	flags    ObjFlags
	parent   Scope
}

func (obj object) Name() string      { return obj.name }
func (obj object) Parent() Scope     { return obj.parent }
func (obj object) SetParent(s Scope) { obj.parent = s }
func (obj object) Pos() loc.Pos      { return obj.pos }
func (obj object) End() loc.Pos      { return obj.end }

type Function struct {
	object
	syntax *ast.FuncDecl
}

func newFunction(node *ast.FuncDecl, parent Scope) *Function {
	f := Function{
		syntax: node,
		object: object{
			kind:   ObjFunc,
			name:   node.Name.String(),
			parent: parent,
		},
	}

	switch node.Kind.Kind {
	case token.TESTCASE:
		f.flags |= ObjTest
	case token.FUNCTION:
	case token.ALTSTEP:
		f.flags |= ObjAltstep
	default:
		panic(fmt.Sprintf("unexpected token %q", node.Kind.Kind.String()))
	}

	if node.External.IsValid() {
		f.flags |= ObjExternal
	}
	return &f
}

func (fun *Function) Pos() loc.Pos { return fun.syntax.Pos() }
func (fun *Function) End() loc.Pos { return fun.syntax.End() }

func (fun *Function) IsTest() bool     { return fun.flags&ObjTest != 0 }
func (fun *Function) IsAltstep() bool  { return fun.flags&ObjAltstep != 0 }
func (fun *Function) IsExternal() bool { return fun.flags&ObjExternal != 0 }

type Import struct {
	object
	syntax *ast.ImportDecl
}

func newImport(node *ast.ImportDecl, parent Scope) *Import {
	return &Import{
		syntax: node,
		object: object{
			kind:   ObjImport,
			name:   node.Module.String(),
			parent: parent,
		},
	}
}

func (imp *Import) Pos() loc.Pos { return imp.syntax.Pos() }
func (imp *Import) End() loc.Pos { return imp.syntax.End() }

type Module struct {
	object
	scope
	syntax *ast.Module
	links  map[*ast.Ident]Scope

	Imports   []*Import
	Tests     []*Function
	Functions []*Function
	Altsteps  []*Function
}

func newModule(node *ast.Module, parent Scope) *Module {
	return &Module{
		syntax: node,
		object: object{
			kind:   ObjModule,
			name:   node.Name.String(),
			parent: parent,
		},
	}
}

func (mod *Module) Parent() Scope          { return mod.scope.parent }
func (mod *Module) SetParent(parent Scope) { mod.scope.parent = parent }

func (mod *Module) Pos() loc.Pos { return mod.syntax.Pos() }
func (mod *Module) End() loc.Pos { return mod.syntax.End() }

type Decl struct {
	object
}

func newDecl(name string) *Decl {
	return &Decl{
		object: object{
			kind: ObjDecl,
			name: name,
		},
	}
}
