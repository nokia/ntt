package lsp

import (
	"bytes"
	"context"
	"fmt"

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
	if params.Options.InsertSpaces {
		p.Indent = fmt.Sprintf("%*s", params.Options.TabSize, " ")
	}

	if err := p.Fprint(b); err != nil {
		log.Debug("formatting:", err.Error())
		return nil, nil
	}
	fset := loc.NewFileSet()
	f := fset.AddFile(params.TextDocument.URI.SpanURI().Filename(), 1, len(b))
	f.SetLinesForContent(b)
	endPos := f.Position(loc.Pos(1 + len(b)))
	// vscode includes a standalone '\n' in a range if the line following it is addressed
	// e.g. \n in line 2 => range(0:0, 3:0) will include it, range(0:0, 2:1) won't
	if b[len(b)-1] == '\n' {
		endPos.Line++
		endPos.Column = 1
	}
	return []protocol.TextEdit{{Range: setProtocolRange(f.Position(1),
		endPos), NewText: out.String()}}, nil
}
