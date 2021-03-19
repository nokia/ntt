package lsp

import (
	"context"
	"fmt"
	"time"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/span"
)

func (s *Server) definition(ctx context.Context, params *protocol.DefinitionParams) (protocol.Definition, error) {
	var (
		locs []protocol.Location
		file = params.TextDocument.URI
		line = int(params.Position.Line) + 1
		col  = int(params.Position.Character) + 1
	)

	for _, suite := range s.Owners(file) {
		start := time.Now()
		id, _ := suite.IdentifierAt(string(file.SpanURI()), line, col)
		elapsed := time.Since(start)
		log.Debug(fmt.Sprintf("Goto Definition took %s. IdentifierInfo: %#v", elapsed, id))

		if id != nil && id.Def != nil {
			locs = append(locs, location(id.Def.Position))
		}
	}

	return locs, nil
}

func location(pos loc.Position) protocol.Location {
	return protocol.Location{
		URI: protocol.URIFromSpanURI(span.URIFromPath(pos.Filename)),
		Range: protocol.Range{
			Start: position(pos.Line, pos.Column),
			End:   position(pos.Line, pos.Column),
		},
	}
}

func position(line, column int) protocol.Position {
	return protocol.Position{
		Line:      float64(line - 1),
		Character: float64(column - 1),
	}
}
