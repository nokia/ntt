package lsp

import (
	"context"
	"sort"

	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
)

func (s *Server) codeLens(ctx context.Context, params *protocol.CodeLensParams) ([]protocol.CodeLens, error) {

	if enabled, _ := s.Config("ttcn3.server.codelens").(bool); !enabled {
		return nil, nil
	}

	var (
		result []protocol.CodeLens
		file   = params.TextDocument.URI
	)

	uri := string(file.SpanURI())

	suite, err := s.FirstSuite(uri)
	if err != nil {
		return nil, err
	}

	tree := ttcn3.ParseFile(uri)
	if tree == nil || tree.Root == nil {
		return nil, nil
	}
	ast.Inspect(tree.Root, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.NodeList, *ast.Module, *ast.ModuleDef:
			return true
		case *ast.FuncDecl:
			if !n.IsTest() {
				return false
			}
			title := "run test"
			params := nttTestParams{
				ID:  tree.QualifiedName(n),
				URI: uri,
			}
			if s.testCtrl.IsRunning(suite, params.ID) {
				title = "test running..."
				params.Stop = true
			}
			if cmd, err := NewCommand(tree.Position(n.Pos()), title, "ntt.test", params); err == nil {
				result = append(result, cmd)
			}
		}
		return false
	})

	sort.Slice(result, func(i, j int) bool {
		a, b := result[i], result[j]
		if protocol.CompareRange(a.Range, b.Range) == 0 {
			return a.Command.Command < b.Command.Command
		}
		return protocol.CompareRange(a.Range, b.Range) < 0
	})
	return result, nil
}
