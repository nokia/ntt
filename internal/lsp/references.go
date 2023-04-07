package lsp

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
)

func newAllIdsWithSameNameFromFile(file string, idName string) []protocol.Location {
	list := make([]protocol.Location, 0, 10)
	syntax := ttcn3.ParseFile(file)
	ast.Inspect(syntax.Root, func(n ast.Node) bool {
		if n == nil {
			// called on node exit
			return false
		}

		switch node := n.(type) {
		case *ast.Ident:
			if idName == node.Tok.String() {
				list = append(list, location(syntax.FileSet.Position(node.Tok.Pos())))
			}
			if node.Tok2 != nil && idName == node.Tok2.String() {
				list = append(list, location(syntax.FileSet.Position(node.Tok2.Pos())))
			}
			return false
		default:
			return true
		}
	})
	return list
}

func NewAllIdsWithSameName(db *ttcn3.DB, name string) []protocol.Location {
	var (
		locs       []protocol.Location
		candidates []string
	)
	for file := range db.Uses[name] {
		candidates = append(candidates, file)
	}
	sort.Strings(candidates)
	for _, file := range candidates {
		locs = append(locs, newAllIdsWithSameNameFromFile(file, name)...)
	}
	return locs
}

func (s *Server) references(ctx context.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
	var (
		file = string(params.TextDocument.URI.SpanURI())
		line = int(params.Position.Line) + 1
		col  = int(params.Position.Character) + 1
	)

	start := time.Now()
	defer func() {
		log.Debug(fmt.Sprintf("References took %s.", time.Since(start)))
	}()

	tree := ttcn3.ParseFile(file)
	id, ok := tree.ExprAt(tree.Pos(line, col)).(*ast.Ident)
	if !ok || id == nil {
		return nil, errors.New("no identifier at cursor")
	}
	return NewAllIdsWithSameName(&s.db, id.String()), nil
}
