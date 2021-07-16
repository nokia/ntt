package lsp

import (
	"context"
	"sort"

	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/token"
)

func (s *Server) codeLens(ctx context.Context, params *protocol.CodeLensParams) ([]protocol.CodeLens, error) {

	if enabled, _ := s.Config("ttcn3.server.codelens").(bool); !enabled {
		return nil, nil
	}

	var (
		result []protocol.CodeLens
		file   = params.TextDocument.URI
	)

	for _, suite := range s.Owners(file) {
		tree := suite.ParseWithAllErrors(string(file.SpanURI()))
		if tree == nil || tree.Module == nil {
			return nil, nil
		}
		ast.Inspect(tree.Module, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.Module, *ast.ModuleDef:
				return true
			case *ast.FuncDecl:
				if n.Kind.Kind != token.TESTCASE {
					return false
				}
				id := ast.Name(tree.Module.Name) + "." + ast.Name(n.Name)
				if cmd, err := NewCommand(tree.Position(n.Pos()), "run test", "ntt.test", param{ID: id, URI: string(file.SpanURI())}); err == nil {
					result = append(result, cmd)
				}
				return false
			default:
				return false
			}
		})
	}

	sort.Slice(result, func(i, j int) bool {
		a, b := result[i], result[j]
		if protocol.CompareRange(a.Range, b.Range) == 0 {
			return a.Command.Command < b.Command.Command
		}
		return protocol.CompareRange(a.Range, b.Range) < 0
	})
	return result, nil
}
