package lsp

import (
	"context"

	"github.com/nokia/ntt/internal/lsp/protocol"
)

func (s *Server) didOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	// TODO(5nord) Sources might added multiple times.
	if params.TextDocument.LanguageID == "ttcn3" {
		s.suite.AddSources(params.TextDocument.URI)
	}

	f := s.suite.File(params.TextDocument.URI)
	f.SetBytes([]byte(params.TextDocument.Text))
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
