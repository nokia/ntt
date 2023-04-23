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

	// The language protocol client requires the range of the original text.
	//
	// Our syntax tree does not have a nice interface for that. So we just
	// scan the whole file and calculate the range explicitly.
	//
	// This is inefficient, but sufficient for now.
	fset := loc.NewFileSet()
	f := fset.AddFile(params.TextDocument.URI.SpanURI().Filename(), 1, len(b))
	f.SetLinesForContent(b)
	begin, end := f.Position(1), f.Position(loc.Pos(1+len(b)))

	// f.Position includes the trailing line-break in line and column
	// calculation. For example the end of string "foo;\n" will be in line
	// 1 and column 5.
	//
	// However, this causes "phantom lines" to appear in vscode, because it
	// expects line-breaks to always start a new line. For example the end
	// of above string should be in line 2 and column 1 instead.
	//
	// As a workaround we adapt the position so the client handles it
	// properly, when a string ends with a line-break
	//
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
