package lsp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/internal/span"
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

func (s *Server) definition(ctx context.Context, params *protocol.DefinitionParams) (protocol.Definition, error) {
	f := s.suite.File(params.TextDocument.URI)
	mod, fset, _ := s.suite.Parse(f)

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

	start := time.Now()
	id, path, _ := findIdentifier(ctx, mod, offs)
	elapsed := time.Since(start)

	// Just for debugging
	if id != nil {
		s.Log(ctx, fmt.Sprintf("Found identifier %q in %s", id.Name, elapsed))
		var de string
		for i, n := range path {
			de = de + fmt.Sprintf("\t%d: %#v\n", i, n)
		}
		s.Log(context.TODO(), "Path:\n"+de)
	}

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

var ErrNoIdentFound = errors.New("no identifier found")

func findIdentifier(ctx context.Context, mod *ast.Module, pos loc.Pos) (*IdentifierInfo, []ast.Node, error) {
	path := pathEnclosingObjNode(mod, pos)

	if len(path) == 0 {
		return nil, nil, ErrNoIdentFound
	}

	ident, _ := path[0].(*ast.Ident)
	if ident == nil {
		return nil, nil, ErrNoIdentFound
	}

	id := IdentifierInfo{Name: ident.String()}

	return &id, path, nil
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

		if n == nil {
			path = path[:len(path)-1]
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
