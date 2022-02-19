package lsp

import (
	"context"
	"fmt"
	"time"

	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
)

func (s *Server) definition(ctx context.Context, params *protocol.DefinitionParams) (protocol.Definition, error) {
	var (
		locs []protocol.Location
		file = string(params.TextDocument.URI.SpanURI())
		line = int(params.Position.Line) + 1
		col  = int(params.Position.Character) + 1
	)

	start := time.Now()
	defer func() {
		log.Debug(fmt.Sprintf("DefintionRequest took %s.\n", time.Since(start)))
	}()

	tree := ttcn3.ParseFile(file)
	x := tree.ExprAt(tree.Pos(line, col))
	if x != nil {
		if defs := tree.LookupWithDB(x, &s.db); len(defs) > 0 {
			for _, def := range defs {
				pos := def.Tree.Position(def.Ident.Pos())
				log.Debugf("Definition found at %s\n", pos)
				locs = append(locs, location(pos))
			}
			return unifyLocs(locs), nil
		}
	}

	// Fallback to cTags
	for _, suite := range s.Owners(params.TextDocument.URI) {
		locs = append(locs, cTags(suite, file, line, col)...)
	}
	return unifyLocs(locs), nil

}

func cTags(suite *ntt.Suite, file string, line int, col int) []protocol.Location {
	stast := time.Now()
	defer func() {
		log.Debug(fmt.Sprintf("cTags fallback took %s\n", time.Since(stast)))
	}()

	var ret []protocol.Location

	tree := suite.Parse(file)
	if tree == nil {
		log.Debug(fmt.Sprintf("Parsing %q failed.\n", file))
		return nil
	}

	// If Module is nil, there's no need to continue, because there's no
	// AST to iterate over.
	if tree.Module == nil {
		return nil
	}
	id := suite.IdentifierAt(tree, line, col)
	if id == nil {
		return nil
	}

	for _, mod := range tree.ImportedModules() {
		file, _ := suite.FindModule(mod)
		if file == "" {
			continue
		}
		tree, tags := suite.Tags(file)
		for _, t := range tags {
			if ast.Name(t) == ast.Name(id) {
				ret = append(ret, location(tree.Position(t.Pos())))
			}
		}
	}
	return ret
}

func unifyLocs(locs []protocol.Location) []protocol.Location {
	m := make(map[protocol.Location]bool)
	for _, loc := range locs {
		m[loc] = true
	}

	ret := make([]protocol.Location, 0, len(m))
	for loc := range m {
		ret = append(ret, loc)
	}
	return ret
}
