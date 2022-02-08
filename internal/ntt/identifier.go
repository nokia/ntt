package ntt

import (
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/ttcn3/ast"
)

//IdentifierAt returns token of identifier at position line:column.
func (suite *Suite) IdentifierAt(mod *ParseInfo, line int, column int) *ast.Ident {
	return findIdentifier(mod.Module, mod.Pos(line, column))
}

func findIdentifier(n ast.Node, pos loc.Pos) *ast.Ident {
	var (
		found bool
		id    *ast.Ident = nil
	)

	ast.Inspect(n, func(n ast.Node) bool {
		if found || n == nil {
			return false
		}

		// We don't need to descend any deeper if we're not near desired
		// position.
		if n.End() < pos || pos < n.Pos() {
			return false
		}

		if n, ok := n.(*ast.Ident); ok {
			if n.Pos() <= pos && pos <= n.End() {
				found = true
				id = n
			}
		}

		return !found
	})

	return id
}
