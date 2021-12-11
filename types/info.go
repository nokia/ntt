package types

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

type Info struct {
	fset   *loc.FileSet
	Types  map[ast.Node]Type
	Scopes map[ast.Node]Scope
}

// TypeOf returns the type of the given expression/typespec. The given scope is stored
// to resolve type references later.
func (info *Info) TypeOf(n ast.Node, scp Scope) Type {
	if typ, ok := info.Types[n]; ok {
		return typ
	}

	switch n := n.(type) {
	case *ast.RefSpec:
		return info.TypeOf(n.X, scp)

	case *ast.StructSpec:
		obj := &Struct{
			Scope: scp,
			kind:  structKind(n.Kind.Kind),
			begin: info.position(ast.FirstToken(n).Pos()),
			end:   info.position(n.End()),
		}
		for _, fld := range n.Fields {
			insertNamedType(fld, obj, info)
		}
		return obj

	case *ast.EnumSpec:
		obj := &Struct{
			kind:  EnumeratedType,
			Scope: scp,
			begin: info.position(ast.FirstToken(n).Pos()),
			end:   info.position(n.End()),
		}
		for _, e := range n.Enums {
			insertEnum(e, obj, info)
		}

	case *ast.ListSpec:
		if n.Length != nil {
			info.trackScopes(n.Length, scp)
		}
		return &List{
			ElemType: info.TypeOf(n.ElemType, scp),
			Scope:    scp,
			kind:     listKind(n.Kind.Kind),
			begin:    info.position(ast.FirstToken(n).Pos()),
			end:      info.position(n.End()),
		}

	case ast.Expr:
		// We shortcut resolving for predefined types.
		if typ, ok := predefinedTypes[ast.Name(n)]; ok {
			return typ
		}
		info.trackScopes(n, scp)
		return &Ref{
			Expr: n,
			Scp:  scp,
		}
	}

	panic(fmt.Sprintf("unhandled type %T", n))
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
		var (
			mod *Module
			err error
		)

		// Check if the module is already defined.
		if obj := scp.Lookup(n.Name.String()); obj != nil {
			switch obj := obj.(type) {
			case *Module:
				mod = obj
			default:
				return &RedefinitionError{Name: n.Name.String(), OldPos: begin(obj), NewPos: begin(mod)}
			}
		} else {

			mod = &Module{
				Name: n.Name.String(),
			}
			err = insert(mod.Name, mod, scp)
		}

		return multierror.Append(err, insertModuleDefs(n.Defs, mod, info)).ErrorOrNil()

	case *ast.GroupDecl:
		return insertModuleDefs(n.Defs, scp, info).ErrorOrNil()

	case *ast.ModuleDef:
		return info.InsertTree(n.Def, scp)

	case *ast.ValueDecl:
		return insertValueDecl(n, scp, info).ErrorOrNil()

	case *ast.TemplateDecl:
		return insertTemplateDecl(n, scp, info)

	case *ast.SubTypeDecl:
		return insertNamedType(n.Field, scp, info)

	case *ast.StructTypeDecl:
		return insertStructTypeDecl(n, scp, info)

	case *ast.EnumTypeDecl:
		return insertEnumTypeDecl(n, scp, info)

	case *ast.ComponentTypeDecl:
		return insertComponentTypeDecl(n, scp, info)

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
	info.trackScopes(n.Value, scp)
	name := n.Name.String()
	obj := &Var{
		Name:  name,
		Type:  wrapArray(n.ArrayDef, typ, scp, info),
		Scope: scp,
		begin: info.position(n.Name.Pos()),
		end:   info.position(n.Name.End()),
	}

	return insert(name, obj, scp)
}

func insertTemplateDecl(n *ast.TemplateDecl, scp Scope, info *Info) error {
	info.trackScopes(n.Value, scp)
	name := n.Name.String()
	obj := &Var{
		Name:  name,
		Type:  info.TypeOf(n.Type, scp),
		Scope: scp,
		begin: info.position(n.Name.Pos()),
		end:   info.position(n.Name.End()),
	}

	return insert(name, obj, scp)
}

func insertStructTypeDecl(n *ast.StructTypeDecl, scp Scope, info *Info) error {
	typ := &Struct{
		kind:  structKind(n.Kind.Kind),
		Scope: scp,
		begin: info.position(ast.FirstToken(n).Pos()),
		end:   info.position(n.End()),
	}
	for _, fld := range n.Fields {
		insertNamedType(fld, typ, info)
	}

	name := n.Name.String()
	obj := &NamedType{
		Name: name,
		Type: typ,
	}

	return insert(name, obj, scp)
}

func insertEnumTypeDecl(n *ast.EnumTypeDecl, scp Scope, info *Info) error {
	typ := &Struct{
		kind:  EnumeratedType,
		Scope: scp,
		begin: info.position(ast.FirstToken(n).Pos()),
		end:   info.position(n.End()),
	}
	for _, e := range n.Enums {
		insertEnum(e, typ, info)
	}

	name := n.Name.String()
	obj := &NamedType{
		Name: name,
		Type: typ,
	}

	return insert(name, obj, scp)
}

func insertComponentTypeDecl(n *ast.ComponentTypeDecl, scp Scope, info *Info) error {
	comp := &Component{
		Scope: scp,
		begin: info.position(ast.FirstToken(n).Pos()),
		end:   info.position(n.End()),
	}

	var errs *multierror.Error
	for _, stmt := range n.Body.Stmts {
		errs = multierror.Append(errs, info.InsertTree(stmt, comp))
	}

	name := n.Name.String()
	obj := &NamedType{
		Name: name,
		Type: comp,
	}

	return multierror.Append(errs, insert(name, obj, scp)).ErrorOrNil()
}

func insertNamedType(n *ast.Field, scp Scope, info *Info) error {
	if n.ValueConstraint != nil {
		info.trackScopes(n.ValueConstraint, scp)
	}
	if n.LengthConstraint != nil {
		info.trackScopes(n.LengthConstraint, scp)
	}

	name := n.Name.String()
	obj := &NamedType{
		Name:  name,
		Type:  wrapArray(n.ArrayDef, info.TypeOf(n.Type, scp), scp, info),
		Scope: scp,
	}
	return insert(name, obj, scp)
}

func insertEnum(n ast.Expr, s *Struct, info *Info) error {
	if c, ok := n.(*ast.CallExpr); ok {
		info.trackScopes(c.Args, s)
	}

	name := ast.Name(n)
	if name == "" {
		return fmt.Errorf("cannot use %T as enum name", n)
	}
	obj := &Var{
		Name:  name,
		Type:  s,
		Scope: s,
		begin: info.position(n.Pos()),
		end:   info.position(n.End()),
	}

	return insert(name, obj, s)
}

// trackScopes tracks the scopes of the given node and its children. The scope
// is will be used to resolve references at a later stage.
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

func (info *Info) position(pos loc.Pos) loc.Position {
	if info.fset == nil {
		return loc.Position{}
	}
	return info.fset.Position(pos)
}

// insert inserts the given object into the given scope. And returns an error
// if another object with the same name already exists.
func insert(name string, obj Object, scp Scope) error {
	if alt := scp.Insert(name, obj); alt != nil {
		return &RedefinitionError{Name: name, OldPos: begin(alt), NewPos: begin(obj)}
	}
	return nil
}

// wrapArray creates an array type from the given array definition and element type.
func wrapArray(n []*ast.ParenExpr, typ Type, scp Scope, info *Info) Type {
	if len(n) == 0 {
		return typ
	}

	return &List{
		kind:     ArrayType,
		ElemType: typ,
		Scope:    scp,
	}
}

func structKind(tok token.Kind) Kind {
	switch tok {
	case token.RECORD:
		return RecordType
	case token.SET:
		return SetType
	case token.UNION:
		return UnionType
	case token.ENUMERATED:
		return EnumeratedType
	default:
		return UnknownType
	}
}

func listKind(tok token.Kind) Kind {
	switch tok {
	case token.RECORD:
		return RecordOfType
	case token.SET:
		return SetOfType
	default:
		return UnknownType
	}
}

// begin returns the begin position of the given object.
func begin(obj Object) loc.Position {
	if rng, ok := obj.(Range); ok {
		return rng.Begin()
	}
	return loc.Position{}
}

var (
	predefinedTypes = map[string]Type{
		"integer": Integer,
	}
)
