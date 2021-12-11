package types

import (
	"github.com/nokia/ntt/ttcn3/ast"
)

// Func describes testcases, altsteps, functions and external functions.
type Func struct {
	object
	external bool
}

func NewFunc(n *ast.FuncDecl, name string) *Func {
	return &Func{
		object: object{
			node: n,
			name: name,
		},
	}
}
