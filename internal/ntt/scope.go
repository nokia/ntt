package ntt

import "github.com/nokia/ntt/internal/loc"

type Scope interface {
	// Parent returns the parent Scope or nil if there isn't any.
	Parent() Scope

	// Insert attemps to insert an object obj into the Scope.
	// If the scope already contains an alternative object alt with the same name, Insert
	// leaves the scope unchanged and returns altnative object. Otherwise it inserts obj, sets the
	// object's parent scope, if not already set, and returns nil.
	Insert(obj Object) Object

	Lookup(name string) Object

	Pos() loc.Pos
	End() loc.Pos
}

// scope is a default implementation for common scopes.
type scope struct {
	pos, end loc.Pos
	parent   Scope
	elems    map[string]Object
}

func newScope(pos, end loc.Pos, parent Scope) *scope {
	return &scope{
		pos:    pos,
		end:    end,
		parent: parent,
	}
}
func (s *scope) Parent() Scope {
	return s.parent
}

func (s *scope) Insert(obj Object) Object {
	name := obj.Name()
	if alt := s.elems[name]; alt != nil {
		return alt
	}
	if s.elems == nil {
		s.elems = make(map[string]Object)
	}
	s.elems[name] = obj
	if obj.Parent() == nil {
		obj.SetParent(s)
	}
	return nil
}

func (s *scope) Lookup(name string) Object {
	if obj := s.elems[name]; obj != nil {
		return obj
	}

	// Ascend into parent scopes.
	if enclosing := s.Parent(); enclosing != nil {
		return enclosing.Lookup(name)
	}

	return nil
}

func (s *scope) Pos() loc.Pos { return s.pos }
func (s *scope) End() loc.Pos { return s.end }
