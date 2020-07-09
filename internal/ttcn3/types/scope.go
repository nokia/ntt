package types

import (
	"sort"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

type Scope interface {

	// Insert attemps to insert an object obj into the Scope. If the scope
	// already contains an alternative object alt with the same name, Insert
	// leaves the scope unchanged and returns altnative object. Otherwise it
	// inserts obj, sets the object's parent scope, if not already set, and
	// returns nil.
	Insert(obj Object) Object

	// Lookup returns the object for a given name. Lookup may follow scope chains.
	Lookup(name string) Object

	// Names lists all names defined in this scope.
	Names() []string
}

// Object describes a named language entity, such as a function or const.
type Object interface {
	Name() string // Object name.

	// Parent returns the (lexical) scope the object is defined in.
	Parent() Scope

	// Type returns the type of the object.
	Type() Type

	// setParent sets the scope the object is defined in.
	setParent(s Scope)

	// Node returns the representing AST node of the object.
	Node() ast.Node

	Range
}

// Range interface is identical to ast.Node interface and helps handling source
// code locations.
type Range interface {
	Pos() loc.Pos
	End() loc.Pos
}

// object implements the common parts of an Object
type object struct {
	node   ast.Node
	name   string
	parent Scope
	typ    Type
}

// Object interface

func (obj *object) Name() string      { return obj.name }
func (obj *object) Parent() Scope     { return obj.parent }
func (obj *object) Type() Type        { return obj.typ }
func (obj *object) setParent(s Scope) { obj.parent = s }

func (obj *object) Node() ast.Node { return obj.node }

// Range interface

func (obj *object) Pos() loc.Pos { return obj.node.Pos() }
func (obj *object) End() loc.Pos { return obj.node.End() }

// scope implements the common parts of Scope
type scope struct {
	elems map[string]Object
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
	return nil
}

func (s *scope) Lookup(name string) Object {
	if obj := s.elems[name]; obj != nil {
		return obj
	}

	return nil
}

func (s *scope) Names() []string {
	names := make([]string, len(s.elems))
	i := 0
	for name := range s.elems {
		names[i] = name
		i++
	}
	sort.Strings(names)
	return names
}

type LocalScope struct {
	pos, end loc.Pos
	parent   Scope
	scope
}

func NewLocalScope(rng Range, parent Scope) *LocalScope {
	return &LocalScope{
		pos:    rng.Pos(),
		end:    rng.End(),
		parent: parent,
	}
}

func (ls *LocalScope) Lookup(name string) Object {
	if obj := ls.scope.Lookup(name); obj != nil {
		return obj
	}

	// Ascend into parent scopes.
	if ls.parent != nil {
		return ls.parent.Lookup(name)
	}

	return nil
}
