package lsp

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/project"
)

func (s *Server) didOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	if !strings.HasPrefix(strings.ToLower(params.TextDocument.LanguageID), "ttcn") {
		return nil
	}

	// Register file for diagnostics and set content
	s.registerFile(params.TextDocument)

	uri := params.TextDocument.URI

	// Every file should be owned by at least one suite to provide proper
	// language support.
	if len(s.Owners(uri)) == 0 {
		log.Printf("File %q does not belong to any known test suite\n", uri)
		dir := filepath.Dir(fs.Open(string(uri.SpanURI())).Path())
		log.Printf("Scanning %q for possible TTCN-3 suites\n", dir)
		if roots := project.Discover(dir); len(roots) > 0 {
			for _, root := range roots {
				s.AddSuite(root)
			}
		} else {
			log.Printf("Could not find a good candidate. Trying %q recursively and hope for the best.\n", dir)
			s.AddSuite(dir)
		}

	}
	s.Diagnose(uri)
	return nil
}

func (s *Server) didChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	f := fs.Open(string(params.TextDocument.URI.SpanURI()))
	for _, ch := range params.ContentChanges {
		f.SetBytes([]byte(ch.Text))
	}
	return nil
}

func (s *Server) didSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) error {
	return nil
}

func (s *Server) didClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	s.unregisterFile(params.TextDocument)
	return nil
}

func (s *Server) registerFile(doc protocol.TextDocumentItem) {
	s.filesMu.Lock()
	defer s.filesMu.Unlock()

	f := fs.Open(string(doc.URI.SpanURI()))
	f.SetBytes([]byte(doc.Text))
	if !s.files[f] {
		s.files[f] = true
	}
}

func (s *Server) unregisterFile(doc protocol.TextDocumentIdentifier) {
	s.filesMu.Lock()
	defer s.filesMu.Unlock()

	f := fs.Open(string(doc.URI.SpanURI()))
	f.Close()
	delete(s.files, f)
}
