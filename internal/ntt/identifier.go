package ntt

import (
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

// IdentInfo associates an identifier reference and its definition.
type IdentInfo struct {
	Syntax   ast.Node
	Position loc.Position
	Def      *IdentInfo
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
		Syntax:   id,
		Position: syntax.Position(id.Pos()),
	}

	// build symbol table for lookup
	syms := suite.symbols(syntax)

	// Fill info struct with everything we have.
	if scp := syms.Scopes[id]; scp != nil {
		if obj := scp.Lookup(id.String()); obj != nil {
			info.Def = &IdentInfo{
				Syntax:   obj.Node(),
				Position: syntax.Position(obj.Pos()),
			}
			return &info, nil
		}

		// Try our luck in imported modules.
		mod := syms.Modules[syntax.Module.Name.String()]
		for i := range mod.Imports {
			if file, _ := suite.FindModule(mod.Imports[i]); file != "" {
				if syntax := suite.Parse(file); syntax.Module != nil {
					// Check if we were looking for a module id
					if syntax.Module.Name.String() == id.String() {
						info.Def = &IdentInfo{
							Syntax:   syntax.Module,
							Position: syntax.Position(syntax.Module.Pos()),
						}
						return &info, nil
					}

					// Build symbol table and check module definitions
					syms := suite.symbols(syntax)
					imp := syms.Modules[mod.Imports[i]]
					if obj := imp.Lookup(id.String()); obj != nil {
						info.Def = &IdentInfo{
							Syntax:   obj.Node(),
							Position: syntax.Position(obj.Pos()),
						}
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
