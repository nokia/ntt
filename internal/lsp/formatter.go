package lsp

import (
	"bytes"
	"context"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/syntax"
	"github.com/nokia/ntt/ttcn3/v2/printer"
)

func (s *Server) formatting(ctx context.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	if !s.serverConfig.FormatEnabled {
		log.Verbose("formatting: disabled")
		return nil, nil
	}

	uri := string(params.TextDocument.URI)
	b, err := fs.Content(uri)
	if err != nil {
		log.Debug("formatting: ", err.Error())
		return nil, nil
	}
	var out bytes.Buffer
	p := printer.NewCanonicalPrinter(&out)
	p.TabWidth = int(params.Options.TabSize)
	if params.Options.InsertSpaces {
		p.UseSpaces = true
	}

	// We don't want module definitions to be indented. By using a negative
	// indentation level we can achieve this for most module definitions.
	p.Indent = -1

	if err := p.Fprint(b); err != nil {
		log.Debug("formatting:", err.Error())
		return nil, nil
	}

	tree := ttcn3.ParseFile(uri)
	begin := syntax.Begin(tree.Root)
	end := syntax.End(tree.Root)

	// TODO(5nord) Test this workaround with other language protocol
	// clients (e.g. emacs, vim, ...)
	if b[len(b)-1] == '\n' {
		end.Line++
		end.Column = 1
	}

	return []protocol.TextEdit{{
		Range:   setProtocolRange(begin, end),
		NewText: out.String(),
	}}, nil
}
