package types

import (
	"github.com/hashicorp/go-multierror"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

type Info struct {
	fset   *loc.FileSet
	Types  map[ast.Expr]Type
	Scopes map[ast.Node]Scope
}

func (info *Info) Position(pos loc.Pos) loc.Position {
	if info.fset == nil {
		return loc.Position{}
	}
	return info.fset.Position(pos)
}

// InsertTree inserts the given syntax tree n into the given parent scope scp.
func (info *Info) InsertTree(n ast.Node, scp Scope) error {
	if scp == nil {
		panic("scp is nil")
	}

	if n == nil {
		return nil
	}

	switch n := n.(type) {
	case *ast.Module:
		return insertModuleDefs(n.Defs, scp, info).ErrorOrNil()

	case *ast.GroupDecl:
		return insertModuleDefs(n.Defs, scp, info).ErrorOrNil()

	case *ast.ModuleDef:
		return info.InsertTree(n.Def, scp)

	case *ast.ValueDecl:
		return insertValueDecl(n, scp, info).ErrorOrNil()

	case ast.NodeList:
		return insertNodes(n, scp, info).ErrorOrNil()

	case *ast.ExprStmt:
		return info.InsertTree(n.Expr, scp)

	case *ast.DeclStmt:
		return info.InsertTree(n.Decl, scp)

	case *ast.ErrorNode:
		return nil

	}

	return &NodeNotImplementedError{Node: n}
}

func (info *Info) TypeOf(n ast.Expr, scp Scope) Type {
	if typ, ok := info.Types[n]; ok {
		return typ
	}

	switch n := n.(type) {
	case ast.Expr:
		// We shortcut resolving for predefined types.
		if typ, ok := predefinedTypes[ast.Name(n)]; ok {
			return typ
		}
		return &Ref{
			Expr: n,
			Scp:  scp,
		}
	}

	info.trackScopes(n, scp)
	return nil
}

func (info *Info) trackScopes(n ast.Node, scp Scope) {
	if info.Scopes == nil {
		info.Scopes = make(map[ast.Node]Scope)
	}
	ast.Inspect(n, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			info.Scopes[id] = scp
		}
		return true
	})
}

func insertNodes(n []ast.Node, scp Scope, info *Info) *multierror.Error {
	var errs *multierror.Error
	for _, n := range n {
		errs = multierror.Append(errs, info.InsertTree(n, scp))
	}
	return errs
}

func insertModuleDefs(n []*ast.ModuleDef, scp Scope, info *Info) *multierror.Error {
	var errs *multierror.Error
	for _, def := range n {
		errs = multierror.Append(errs, info.InsertTree(def, scp))
	}
	return errs
}

func insertValueDecl(n *ast.ValueDecl, scp Scope, info *Info) *multierror.Error {
	var errs *multierror.Error
	typ := info.TypeOf(n.Type, scp)

	for _, decl := range n.Decls {
		errs = multierror.Append(errs, insertDeclarator(decl, typ, scp, info))
	}
	return errs
}

func insertDeclarator(n *ast.Declarator, typ Type, scp Scope, info *Info) error {
	name := n.Name.String()
	if n.ArrayDef != nil {
		typ = makeArray(n.ArrayDef, typ, scp, info)
	}

	obj := &Var{
		Name:  name,
		Type:  typ,
		Scope: scp,
		begin: info.Position(n.Name.Pos()),
		end:   info.Position(n.Name.End()),
	}

	return insert(name, obj, scp)
}

func insert(name string, obj Object, scp Scope) error {
	if alt := scp.Insert(name, obj); alt != nil {
		return &RedefinitionError{Name: name, OldPos: begin(alt), NewPos: begin(obj)}
	}
	return nil
}

func begin(obj Object) loc.Position {
	if rng, ok := obj.(Range); ok {
		return rng.Begin()
	}
	return loc.Position{}
}

func makeArray(n []*ast.ParenExpr, typ Type, scp Scope, info *Info) Type {
	// TODO(5nord) implement list types
	return typ
}

var (
	predefinedTypes = map[string]Type{
		"integer": Integer,
	}
)
