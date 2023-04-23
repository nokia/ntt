package lsp

import (
	"context"
	"sort"

	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/syntax"
)

func (s *Server) codeLens(ctx context.Context, params *protocol.CodeLensParams) ([]protocol.CodeLens, error) {

	if enabled, _ := s.Config("ttcn3.server.codelens").(bool); !enabled {
		return nil, nil
	}

	var (
		result []protocol.CodeLens
		file   = string(params.TextDocument.URI.SpanURI())
	)

	tree := ttcn3.ParseFile(file)
	if tree == nil || tree.Root == nil {
		return nil, nil
	}
	tree.Inspect(func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.NodeList, *syntax.Module, *syntax.ModuleDef:
			return true
		case *syntax.FuncDecl:
			if !n.IsTest() {
				return false
			}
			title := "run test"
			params := nttTestParams{
				TestID: TestID{
					URI:  file,
					Name: tree.QualifiedName(n),
					Pos:  n.Pos(),
				},
			}
			if s.testCtrl.IsRunning(params.TestID) {
				title = "test running..."
				params.Stop = true
			}
			if cmd, err := NewCommand(syntax.Begin(n), title, "ntt.test", params); err == nil {
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
