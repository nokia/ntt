package lsp

import (
	"context"
	"fmt"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
)

func (s *Server) hover(ctx context.Context, params *protocol.HoverParams) (protocol.Hover, error) {
	var (
		file = string(params.TextDocument.URI.SpanURI())
		line = int(params.Position.Line) + 1
		col  = int(params.Position.Character) + 1
	)

	tree := ttcn3.ParseFile(file)
	x := tree.ExprAt(tree.Pos(line, col))
	if x == nil {
		log.Debug(fmt.Sprintf("No expression at %s:%d:%d\n", file, line, col))
	}

	comment := ast.FirstToken(x).Comments()

	hoverContents := protocol.MarkupContent{Kind: "Plaintext", Value: comment}
	hover := protocol.Hover{Contents: hoverContents}

	return hover, nil
}
