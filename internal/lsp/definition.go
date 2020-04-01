package lsp

import (
	"context"
	"fmt"
	"time"

	"github.com/nokia/ntt/internal/lsp/protocol"
)

func (s *Server) definition(ctx context.Context, params *protocol.DefinitionParams) (protocol.Definition, error) {
	start := time.Now()
	id, _ := s.suite.IdentifierAt(params.TextDocument.URI, int(params.Position.Line)+1, int(params.Position.Character)+1)
	elapsed := time.Since(start)
	s.Log(ctx, fmt.Sprintf("Goto definition took %s. IdentifierInfo: %#v", elapsed, id))

	if id != nil && id.Def != nil {
		return []protocol.Location{
			{
				URI: string(params.TextDocument.URI),
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      float64(id.Line(id.Def.Pos()) - 1),
						Character: float64(id.Column(id.Def.Pos()) - 1),
					},
					End: protocol.Position{
						Line:      float64(id.Line(id.Def.End()) - 1),
						Character: float64(id.Column(id.Def.End()) - 1),
					},
				},
			},
		}, nil
	}
	return nil, nil
}
