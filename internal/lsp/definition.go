package lsp

import (
	"context"
	"errors"
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
	elapsed := time.Since(start)
	if id != nil {
		s.Log(ctx, fmt.Sprintf("Found identifier %q in %s", id.Name, elapsed))
	}

	return nil, nil
}

type IdentifierInfo struct {
	Name string
}

var ErrNoIdentFound = errors.New("no identifier found")

func findIdentifier(ctx context.Context, mod *ast.Module, pos loc.Pos) (*IdentifierInfo, error) {
	path := pathEnclosingObjNode(mod, pos)
	if path == nil {
		return nil, ErrNoIdentFound
	}

	ident, _ := path[0].(*ast.Ident)
	if ident == nil {
		return nil, ErrNoIdentFound
	}

	id := IdentifierInfo{Name: ident.String()}

	return &id, nil
}

func pathEnclosingObjNode(mod *ast.Module, pos loc.Pos) []ast.Node {
	var (
		path  []ast.Node
		found bool
	)

	ast.Inspect(mod, func(n ast.Node) bool {
		if found {
			return false
		}

		path = append(path, n)

		switch n := n.(type) {
		case *ast.Ident:
			found = n.Pos() <= pos && pos <= n.End()
		}

		return !found
	})

	// Reverse the path.
	for i, l := 0, len(path); i < l/2; i++ {
		path[i], path[l-1-i] = path[l-1-i], path[i]
	}

	return path
}
