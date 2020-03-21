package lsp

import (
	"context"
	"fmt"
	"time"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/internal/span"
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

func (s *Server) definition(ctx context.Context, params *protocol.DefinitionParams) (protocol.Definition, error) {
	start := time.Now()
	f := s.suite.File(params.TextDocument.URI)
	mod, fset, _ := s.suite.Parse(f)
	elapsed := time.Since(start)
	s.Log(ctx, fmt.Sprintf("Parsing %q took %s", f.URI(), elapsed))

	// Without AST we don't need to continue any further. We also don't return
	// any syntax error or such. The diagnostics are probably a better place for
	// that.
	if mod == nil {
		return nil, nil
	}

	offs, err := toOffset(f, fset, mod, params.Position)
	if err != nil {
		return nil, err
	}

	// Find identifier
	start = time.Now()
	id := findIdentifier(ctx, mod, offs)
	if id == nil {
		return nil, nil
	}
	elapsed = time.Since(start)
	s.Log(ctx, fmt.Sprintf("Found identifier %q in %s", id.String(), elapsed))

	// Build scope tree
	start = time.Now()
	m, err := s.suite.Symbols(params.TextDocument.URI)
	if err != nil {
		return nil, nil
	}
	elapsed = time.Since(start)
	s.Log(ctx, fmt.Sprintf("Symbols() took %s", elapsed))

	start = time.Now()
	_, obj := m.Lookup(id.String())
	elapsed = time.Since(start)

	if obj != nil {
		s.Log(ctx, fmt.Sprintf("Found %#v. Lookup took %s", obj, elapsed))
		return []protocol.Location{
			{
				URI: string(f.URI()),
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      float64(fset.Position(obj.Pos()).Line - 1),
						Character: float64(fset.Position(obj.Pos()).Column - 1),
					},
					End: protocol.Position{
						Line:      float64(fset.Position(obj.End()).Line - 1),
						Character: float64(fset.Position(obj.End()).Column - 1),
					},
				},
			},
		}, nil

	}

	s.Log(ctx, "Done.\n")
	return nil, nil
}

func toOffset(f *ntt.File, fset *loc.FileSet, node ast.Node, pos protocol.Position) (loc.Pos, error) {
	b, _ := f.Bytes()

	m := &protocol.ColumnMapper{
		URI:       f.URI(),
		Converter: span.NewTokenConverter(fset, fset.File(node.Pos())),
		Content:   b,
	}
	spn, err := m.PointSpan(pos)
	if err != nil {
		return loc.NoPos, err
	}
	rng, err := spn.Range(m.Converter)
	if err != nil {
		return loc.NoPos, err
	}

	return rng.Start, nil
}

type IdentifierInfo struct {
	Name string
}

func findIdentifier(ctx context.Context, mod *ast.Module, pos loc.Pos) *ast.Ident {
	var (
		found bool
		id    *ast.Ident = nil
	)

	ast.Inspect(mod, func(n ast.Node) bool {
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
