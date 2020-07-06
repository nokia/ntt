package types

import (
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

type TypeName struct {
	object
}

func NewTypeName(n ast.Node, name string, typ Type) *TypeName {
	return &TypeName{
		object: object{
			node: n,
			name: name,
			typ:  typ,
		},
	}
}
