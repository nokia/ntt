package runtime

import "github.com/nokia/ntt/internal/ttcn3/ast"

type Suite interface {
	Modules() []*Module
	Tests() []*Test
	Syntax() []*ast.Module
}
