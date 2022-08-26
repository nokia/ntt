package lsp

import (
	"bytes"
	"context"
	"strings"

	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/printer"
)

func getSignature(def *ttcn3.Definition) string {
	var sig bytes.Buffer
	switch node := def.Node.(type) {
	case *ast.FuncDecl:
		sig.WriteString(node.Kind.Lit + " " + node.Name.String())
		printer.Print(&sig, def.FileSet, node.Params)
		if node.RunsOn != nil {
			sig.WriteString("\n  ")
			printer.Print(&sig, def.FileSet, node.RunsOn)
		}
		if node.System != nil {
			sig.WriteString("\n  ")
			printer.Print(&sig, def.FileSet, node.System)
		}
		if node.Return != nil {
			sig.WriteString("\n  ")
			printer.Print(&sig, def.FileSet, node.Return)
		}
	}
	return sig.String()
}

func (s *Server) hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	var (
		file      = string(params.TextDocument.URI.SpanURI())
		line      = int(params.Position.Line) + 1
		col       = int(params.Position.Character) + 1
		comment   string
		signature string
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
			signature = getSignature(def) + "\n"
		}
	}

	hoverContents := protocol.MarkupContent{Kind: "markdown", Value: "```typescript\n" + string(signature) + "```\n - - -\n" + comment}
	hover := &protocol.Hover{Contents: hoverContents}

	return hover, nil
}
