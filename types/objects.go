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

type NamedType struct {
	Name  string
	Type  Type
	Scope Scope
}

func (n *NamedType) EnclosingScope() Scope {
	return n.Scope
}

func (n *NamedType) Underlying() Type {
	return n.Type.Underlying()
}

func (n *NamedType) CompatibleTo(other Type) bool {
	return n.Type.CompatibleTo(other)
}

// Struct represents a structured type, such as record, set, union or enumerated.
type Struct struct {
	Kind  Kind
	Scope Scope

	begin, end loc.Position
	fields     []pair
	names      map[string]pair
}

func (s *Struct) EnclosingScope() Scope {
	return s.Scope
}

// Insert inserts an object into the scope.
func (s *Struct) Insert(name string, obj Object) Object {
	if s.names == nil {
		s.names = make(map[string]pair)
	}
	if alt, ok := s.names[name]; ok {
		return alt.obj
	}

	s.names[name] = pair{name, obj}
	s.fields = append(s.fields, pair{name, obj})
	return obj
}

// Lookup returns the object with the given name in the scope.
func (s *Struct) Lookup(name string) Object {
	if p, ok := s.names[name]; ok {
		return p.obj
	}
	return nil
}

// Names returns the names of all objects in the scope using the order of insertion.
func (s *Struct) Names() []string {
	names := make([]string, len(s.fields))
	for i, p := range s.fields {
		names[i] = p.name
	}
	return names
}

func (s *Struct) Underlying() Type {
	return s
}

func (s *Struct) CompatibleTo(other Type) bool {
	panic("not implemented")
}

func (s *Struct) Begin() loc.Position {
	return s.begin
}

func (s *Struct) End() loc.Position {
	return s.end
}

type List struct {
	ElemType   Type
	Scope      Scope
	begin, end loc.Position
}

func (l *List) EnclosingScope() Scope {
	return l.Scope
}

func (l *List) Underlying() Type {
	return l
}

func (l *List) CompatibleTo(other Type) bool {
	panic("not implemented")
}

func (l *List) Begin() loc.Position {
	return l.begin
}

func (l *List) End() loc.Position {
	return l.end
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
