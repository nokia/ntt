package lsp

import (
	"bytes"
	"context"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/ttcn3/v2/printer"
)

func (s *Server) formatting(ctx context.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	b, err := fs.Content(params.TextDocument.URI.SpanURI().Filename())
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

	// We don't want module defintions to be indented. By using a negative
	// indentation level we can achieve this for most module definitions.
	p.Indent = -1

	if err := p.Fprint(b); err != nil {
		log.Debug("formatting:", err.Error())
		return nil, nil
	}
	fset := loc.NewFileSet()
	f := fset.AddFile(params.TextDocument.URI.SpanURI().Filename(), 1, len(b))
	f.SetLinesForContent(b)
	return []protocol.TextEdit{{Range: setProtocolRange(f.Position(1),
		f.Position(loc.Pos(1+len(b)))), NewText: out.String()}}, nil
}
