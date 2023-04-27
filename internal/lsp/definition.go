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
	x := tree.ExprAt(line, col)
	if x == nil {
		log.Debug(fmt.Sprintf("No expression at %s:%d:%d\n", file, line, col))
	}

	for _, def := range tree.LookupWithDB(x, &s.db) {
		span := syntax.SpanOf(def.Ident)
		log.Debugf("Definition found at %s\n", &span)
		locs = append(locs, location(span))
	}

	return unifyLocs(locs), nil
}
