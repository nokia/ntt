package lsp

import (
	"context"
	"fmt"
	"time"

	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/span"
)

func (s *Server) definition(ctx context.Context, params *protocol.DefinitionParams) (protocol.Definition, error) {
	start := time.Now()
	id, _ := s.suite.IdentifierAt(string(params.TextDocument.URI.SpanURI()), int(params.Position.Line)+1, int(params.Position.Character)+1)
	elapsed := time.Since(start)
	log.Debug(fmt.Sprintf("Goto Definition took %s. IdentifierInfo: %#v", elapsed, id))

	if id != nil && id.Def != nil {
		file := span.URIFromPath(id.Def.Position.Filename)
		line := id.Def.Position.Line - 1
		column := id.Def.Position.Column - 1
		return []protocol.Location{
			{
				URI: protocol.URIFromSpanURI(file),
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      float64(line),
						Character: float64(column),
					},
					End: protocol.Position{
						Line:      float64(line),
						Character: float64(column),
					},
				},
			},
		}, nil
	}
	return nil, nil
}
