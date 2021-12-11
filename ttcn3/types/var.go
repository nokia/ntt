package types

import (
	"github.com/nokia/ntt/ttcn3/ast"
)

// Var describes an object, which can hold an value. This could be a local
// variable, a constant, a module parameter or a template.
type Var struct {
	object
}

func NewVar(n ast.Node, name string) *Var {
	return &Var{
		object: object{
			node: n,
			name: name,
		},
	}
}
