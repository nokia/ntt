package ntt

import (
	"bytes"
	"fmt"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

// Type represents a type in TTCN-3
type Type interface {
	// Underlying  returns the underlying type of a type.
	Underlying() Type

	// String returns a string representation of a type.
	String() string
}

type Kind int

const (
	Invalid Kind = iota
	Bitstring
	Boolean
	Charstring
	Component
	Float
	Hexstring
	Integer
	Octetstring
	Omit
	Template
	Timer
	UniversalCharstring
	Verdict

	// TODO(5nord) Merge strings types into String and merge integer, float into
	// Numerical. Or make them sort of untyped types like unused, omit and
	// template?
	String
	Numerical
)

type BasicType struct {
	kind Kind
	name string
}

func (b *BasicType) Kind() Kind       { return b.kind }
func (b *BasicType) Underlying() Type { return b }
func (b *BasicType) String() string   { return b.name }

// Struct can be an union, a record or a set structure.
type Struct struct {
	scope
	pos, end loc.Pos
	parent   Scope
	Fields   []*Var
}

func NewStruct(rng Range, parent Scope) *Struct {
	return &Struct{
		pos:    rng.Pos(),
		end:    rng.End(),
		parent: parent,
	}
}

func (s *Struct) Pos() loc.Pos  { return s.pos }
func (s *Struct) End() loc.Pos  { return s.end }
func (s *Struct) Parent() Scope { return s.parent }

func (s *Struct) Insert(obj Object) Object {
	if alt := s.scope.Insert(obj); alt != nil {
		return alt
	}

	s.Fields = append(s.Fields, obj.(*Var))
	return nil
}

func (s *Struct) Underlying() Type { return s }
func (s *Struct) String() string {
	var buf bytes.Buffer
	fmt.Fprint(&buf, "struct{}")
	return buf.String()
}

type Port struct {
	object
}

func NewPort() *Port {
	return &Port{}
}

type ComponentType struct {
	scope
	object
	Vars  []*Var
	Ports []*Port
}

func NewComponentType(n *ast.ComponentTypeDecl, name string) *ComponentType {
	return &ComponentType{
		object: object{
			node: n,
			name: name,
		},
	}
}

func (c *ComponentType) Insert(obj Object) Object {
	if alt := c.scope.Insert(obj); alt != nil {
		return alt
	}

	switch obj := obj.(type) {
	case *Var:
		c.Vars = append(c.Vars, obj)
	case *Port:
		c.Ports = append(c.Ports, obj)
	default:
		// TODO(5nord) Add error
	}
	return nil
}

func (c *ComponentType) Underlying() Type { return c }
func (c *ComponentType) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "component %s", c.name)
	return buf.String()
}
