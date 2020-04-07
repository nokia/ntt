package ntt

import (
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

type IdentInfo struct {
	Syntax *ast.Ident
	Def    ast.Node
	Type   Type

	ParseInfo *ParseInfo
}

func (info *IdentInfo) Line(pos loc.Pos) int {
	return info.ParseInfo.Position(pos).Line
}

func (info *IdentInfo) Column(pos loc.Pos) int {
	return info.ParseInfo.Position(pos).Column
}

// IdentifierAt parses file and returns IdentInfo about identifier at position
// line:column.
func (suite *Suite) IdentifierAt(file string, line int, column int) (*IdentInfo, error) {
	syntax := suite.Parse(file)
	if syntax.Module == nil {
		return nil, syntax.Err
	}

	id := findIdentifier(syntax.Module, syntax.Pos(line, column))
	if id == nil {
		return nil, ErrNoIdentFound
	}

	info := IdentInfo{
		Syntax:    id,
		ParseInfo: syntax,
	}

	// build symbol table for lookup
	mod := suite.symbols(syntax)

	// Fill info struct with everything we have.
	if scp := mod.Scopes[id]; scp != nil {
		if obj := scp.Lookup(id.String()); obj != nil {
			info.Def = obj.Node()
			info.Type = obj.Type()
			return &info, nil
		}

		// Try our luck in import imported.
		for i := range mod.Imports {
			if file, _ := suite.FindModule(mod.Imports[i]); file != "" {
				if syntax := suite.Parse(file); syntax.Module != nil {
					imp := suite.symbols(syntax)
					if obj := imp.Lookup(id.String()); obj != nil {
						info.Def = obj.Node()
						info.Type = obj.Type()
						return &info, nil
					}
				}
			}
		}

	}
	return &info, nil
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
