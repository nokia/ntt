package lsp

import (
	"context"

	"github.com/nokia/ntt/internal/lsp/protocol"
)

func (s *Server) didOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	path := params.TextDocument.URI.SpanURI().Filename()
	// TODO(5nord) Sources might added multiple times.
	if params.TextDocument.LanguageID == "ttcn3" {
		s.suite.AddSources(path)
	}

	f := s.suite.File(path)
	f.SetBytes([]byte(params.TextDocument.Text))
	s.Diagnose()
	return nil
}

func (s *Server) didChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	f := s.suite.File(params.TextDocument.URI)
	for _, ch := range params.ContentChanges {
		f.SetBytes([]byte(ch.Text))
	}
	return nil
}

func (s *Server) didSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) error {
	return nil
}

func (s *Server) didClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	f := s.suite.File(params.TextDocument.URI)
	f.Reset()
	return nil
}
