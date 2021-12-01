package types

import (
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

// Var represents a variable.
type Var struct {
	Name  string
	Type  Type
	Scope Scope

	begin, end loc.Position
}

func (v *Var) Begin() loc.Position {
	return v.begin
}

func (v *Var) End() loc.Position {
	return v.begin
}

func (v *Var) EnclosingScope() Scope {
	return v.Scope
}

// Basic represents a basic TTCN-3 type, such as integer, boolean, ...
type Basic struct {
	Kind Kind
}

func (b *Basic) EnclosingScope() Scope {
	return nil
}

func (b *Basic) CompatibleTo(other Type) bool {
	if other, ok := other.(*Basic); ok {
		return b.Kind == other.Kind
	}
	return false
}

func (b *Basic) Underlying() Type {
	return b
}

func (b *Basic) String() string {
	return string(b.Kind)
}

// Ref is a reference to an object.
type Ref struct {
	Expr ast.Expr // The expression that refers to the object.

	Scp Scope  // The context ( == scope) of the reference.
	Obj Object // The object referenced by the reference.
}

func (r *Ref) EnclosingScope() Scope { // NOTE: bad name for this interface
	return r.Scp
}

func (r *Ref) CompatibleTo(other Type) bool {
	panic("not implemented")
}

func (r *Ref) Underlying() Type {
	panic("not implemented")
}
