// Package types provides type inference and checking for TTCN-3.
package types

import (
	"fmt"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/ttcn3/ast"
)

var (
	Integer = &Basic{kind: IntegerType}
	Boolean = &Basic{kind: BooleanType}
)

const (
	UnknownType    Kind = "unknown type"
	IntegerType    Kind = "integer"
	UnionType      Kind = "union"
	EnumeratedType Kind = "enumerated"
	SetType        Kind = "set"
	RecordType     Kind = "record"
	BooleanType    Kind = "boolean"
	RecordOfType   Kind = "record of"
	SetOfType      Kind = "set of"
	ArrayType      Kind = "array of"
	ComponentType  Kind = "component"
	TypeReference  Kind = "type reference"
)

// Kind returns the kind of the object.
type Kind string

// Object represents a type system object; such a scope, a types, a variable, ...
type Object interface {
	// Scope returns the enclosing (== lexical) scope of an object.
	EnclosingScope() Scope
}

// Scope represents a TTCN-3 scope; such as a modules, a function, ...
type Scope interface {
	Object

	// Insert inserts an object into the scope.
	Insert(name string, obj Object) Object

	// Lookup returns the object with the given name in the scope.
	Lookup(name string) Object

	// Names returns the names of all objects in the scope.
	Names() []string
}

// Type represents a TTCN-3 type.
type Type interface {
	Object

	Kind() Kind

	// Compatible returns true if the type is compatible with the given type.
	CompatibleTo(other Type) bool
}

// Range represents a range of TTCN-3 source code.
type Range interface {
	Begin() loc.Position
	End() loc.Position
}

// Equal returns true if the two object are equal.
func Equal(a, b Object) bool {
	if a == nil || b == nil {
		return false
	}

	switch a := a.(type) {
	case *Var:
		b, ok := b.(*Var)
		if !ok {
			return false
		}
		return a.Name == b.Name && a.Type == b.Type
	}
	return true
}

// NodeNotImplementedError is returned when a syntax node is not implemented.
type NodeNotImplementedError struct {
	Node ast.Node
}

func (e *NodeNotImplementedError) Error() string {
	// TODO(5nord) Add position information.
	return fmt.Sprintf("syntax node not implemented: %T", e.Node)
}

type RedefinitionError struct {
	Name           string
	OldPos, NewPos loc.Position
}

func (e *RedefinitionError) Error() string {
	return fmt.Sprintf("redefinition of %s: %s", e.Name, e.OldPos)
}
