package types

import "github.com/nokia/ntt/internal/ttcn3/ast"

// Import describes the view to an imported module.
type Import struct {
	object
	module string
}

func NewImport(n *ast.ImportDecl, name string, module string) *Import {
	return &Import{
		object: object{
			node: n,
			name: name,
		},
		module: module,
	}
}
