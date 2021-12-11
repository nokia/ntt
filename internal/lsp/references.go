package lsp

import (
	"context"
	"fmt"
	"time"

	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/project"
)

func newAllIdsWithSameNameFromFile(suite *ntt.Suite, file string, idName string) []protocol.Location {
	list := make([]protocol.Location, 0, 10)
	syntax := suite.Parse(file)
	ast.Inspect(syntax.Module, func(n ast.Node) bool {
		if n == nil {
			// called on node exit
			return false
		}

		switch node := n.(type) {
		case *ast.Ident:
			if idName == node.Tok.String() {
				list = append(list, location(syntax.FileSet.Position(node.Tok.Pos())))
			}
			if idName == node.Tok2.String() {
				list = append(list, location(syntax.FileSet.Position(node.Tok2.Pos())))
			}
			return false
		default:
			return true
		}
	})
	return list
}
func newAllIdsWithSameName(suite *ntt.Suite, idName string) []protocol.Location {
	var complList []protocol.Location = nil
	if files := project.FindAllFiles(suite); len(files) != 0 {
		complList = make([]protocol.Location, 0, len(files))
		for _, f := range files {
			complList = append(complList, newAllIdsWithSameNameFromFile(suite, f, idName)...)
		}
	}
	return complList
}

func (s *Server) references(ctx context.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
	var (
		locs []protocol.Location
		file = params.TextDocument.URI
		line = int(params.Position.Line) + 1
		col  = int(params.Position.Character) + 1
	)

	for _, suite := range s.Owners(file) {
		start := time.Now()
		syntax := suite.Parse(string(file.SpanURI()))
		if syntax.Module == nil {
			return nil, syntax.Err
		}
		id := suite.IdentifierAt(syntax, line, col)
		if id == nil {
			return nil, ntt.ErrNoIdentFound
		}
		locs = append(locs, newAllIdsWithSameName(suite, id.Tok.String())...)
		elapsed := time.Since(start)
		log.Debug(fmt.Sprintf("References took %s. IdentifierInfo: %#v", elapsed, id.String()))
	}
	return unifyLocs(locs), nil
}
