package types

import "github.com/nokia/ntt/ttcn3/ast"

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
