package types

import "github.com/nokia/ntt/internal/ttcn3/ast"

// Struct can be an union, a record or a set structure.
type EnumeratedType struct {
	scope
	object
	BasicType
	Fields []*Var
}

func NewEnumeratedType(n ast.Node, name string, typ Type) *EnumeratedType {
	return &EnumeratedType{
		object: object{
			node: n,
			name: name,
			typ:  typ,
		},
	}
}
