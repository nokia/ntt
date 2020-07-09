package types

import (
	"bytes"
	"fmt"

	"github.com/nokia/ntt/internal/loc"
)

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
