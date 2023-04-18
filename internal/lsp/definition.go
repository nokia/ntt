package lsp

import (
	"context"
	"fmt"
	"time"

	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/syntax"
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
	if x == nil {
		log.Debug(fmt.Sprintf("No expression at %s:%d:%d\n", file, line, col))
	}

	for _, def := range tree.LookupWithDB(x, &s.db) {
		pos := syntax.Begin(def.Ident)
		log.Debugf("Definition found at %s\n", pos)
		locs = append(locs, location(pos))
	}

	return unifyLocs(locs), nil
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
