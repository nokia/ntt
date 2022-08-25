package lsp

import (
	"context"
	"strings"

	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
)

func (s *Server) hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	var (
		file    = string(params.TextDocument.URI.SpanURI())
		line    = int(params.Position.Line) + 1
		col     = int(params.Position.Character) + 1
		comment string
	)

	tree := ttcn3.ParseFile(file)
	x := tree.ExprAt(tree.Pos(line, col))
	if x == nil {
		return nil, nil
	}
	for _, def := range tree.LookupWithDB(x, &s.db) {
		if firstTok := ast.FirstToken(def.Node); firstTok == nil {
			continue
		} else {
			// make line breaks conform to markdown spec
			comment = strings.ReplaceAll(firstTok.Comments(), "\n", "  \n")
		}
	}

	hoverContents := protocol.MarkupContent{Kind: "markdown", Value: comment}
	hover := &protocol.Hover{Contents: hoverContents}

	return hover, nil
}
