package types

import (
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

type pair struct {
	name string
	obj  Object
}

// Module represents a TTCN-3 module.
type Module struct {
	Name  string
	Scope Scope
	pairs []pair
	names map[string]pair
}

// EnclsosingScope returns the parent (== global) scope of the module
func (m *Module) EnclosingScope() Scope {
	return m.Scope
}

// Insert inserts an object into the scope.
func (m *Module) Insert(name string, obj Object) Object {
	if m.names == nil {
		m.names = make(map[string]pair)
	}
	if alt, ok := m.names[name]; ok {
		return alt.obj
	}

	m.names[name] = pair{name, obj}
	m.pairs = append(m.pairs, pair{name, obj})
	return obj
}

// Lookup returns the object with the given name in the scope.
func (m *Module) Lookup(name string) Object {
	if p, ok := m.names[name]; ok {
		return p.obj
	}
	return nil
}

// Names returns the names of all objects in the scope using the order of insertion.
func (m *Module) Names() []string {
	names := make([]string, len(m.pairs))
	for i, p := range m.pairs {
		names[i] = p.name
	}
	return names
}

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
