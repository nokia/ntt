package lsp

import (
	"bytes"
	"context"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/format"
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
	if len(b) < 1 {
		log.Debugln("formatting: zero length file")
		return nil, nil
	}

	var out bytes.Buffer
	p := format.NewCanonicalPrinter(&out)
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
	if tree.Err != nil {
		log.Debug("skip formatting: ", tree.Err.Error())
		return nil, nil
	}
	begin := tree.Position(0)
	end := tree.Position(len(b)) // The end position is exclusive.

	if b[len(b)-1] == '\n' {
		end.Line++
		end.Column = 1
	}

	return []protocol.TextEdit{{
		Range:   setProtocolRange(begin, end),
		NewText: out.String(),
	}}, nil
}
