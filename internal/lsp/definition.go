package lsp

import (
	"context"
	"fmt"
	"time"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/span"
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

func (s *Server) definition(ctx context.Context, params *protocol.DefinitionParams) (protocol.Definition, error) {
	f := s.suite.File(params.TextDocument.URI)
	start := time.Now()
	mod, fset, _ := s.suite.Parse(f)
	elapsed := time.Since(start)

	// Without AST we don't need to continue any further.
	if mod == nil {
		return nil, nil
	}

	b, _ := f.Bytes()

	m := &protocol.ColumnMapper{
		URI:       f.URI(),
		Converter: span.NewTokenConverter(fset, fset.File(mod.Pos())),
		Content:   b,
	}
	spn, err := m.PointSpan(params.Position)
	if err != nil {
		return nil, err
	}
	rng, err := spn.Range(m.Converter)
	if err != nil {
		return nil, err
	}

	id, _ := findIdentifier(ctx, mod, rng.Start)

	//params.Position.Line
	//params.Position.Character

	s.Log(ctx, fmt.Sprintf("Found identifier %q in %s", id.Name, elapsed))
	return nil, nil
}

type IdentifierInfo struct {
	Name string
}

func findIdentifier(ctx context.Context, mod *ast.Module, pos loc.Pos) (*IdentifierInfo, error) {
	return nil, nil
}
