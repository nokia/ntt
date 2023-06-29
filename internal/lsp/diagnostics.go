package lsp

import (
	"context"
	"errors"

	"github.com/hashicorp/go-multierror"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/syntax"
)

// Diagnose runs various checks over a ttcn3 test suite.
//
// From LSP spec:
//
//	Diagnostics are "owned" by the server so it is the server's
//	responsibility to clear them if necessary.
//
//	If a language has a project system (for example C#) diagnostics are not
//	cleared when a file closes.  When a project is opened all diagnostics
//	for all files are recomputed (or read from a cache).
//
//	When a file changes it is the server’s responsibility to re-compute
//	diagnostics and push them to the client. If the computed set is empty it
//	has to push the empty array to clear former diagnostics. Newly pushed
//	diagnostics always replace previously pushed diagnostics. There is no
//	merging that happens on the client side.
func (s *Server) Diagnose(uris ...protocol.DocumentURI) {
	s.diagsMu.Lock()
	defer s.diagsMu.Unlock()

	s.diags = make(map[string][]protocol.Diagnostic)
	defer s.syncDiagnostics()

	// TODO(5nord): Run linter against uris
	for _, uri := range uris {
		s.client.PublishDiagnostics(context.TODO(), &protocol.PublishDiagnosticsParams{
			Diagnostics: make([]protocol.Diagnostic, 0),
			URI:         uri,
		})
		tree := ttcn3.ParseFile(string(uri))
		if err := tree.Err; err != nil {
			s.reportError(err)
		}
	}
}

func (s *Server) reportError(err error) {
	var (
		serr syntax.Error
		merr *multierror.Error
	)

	switch {

	// Unpack multierrors
	case errors.As(err, &merr):
		for _, e := range merr.Errors {
			s.reportError(e)
		}

	// Errors with a location will become diagnostics
	case errors.As(err, &serr):
		span := syntax.SpanOf(serr.Node)
		uri := string(fs.Open(span.Filename).URI())
		diag := protocol.Diagnostic{
			Severity: protocol.SeverityError,
			Source:   string(fs.URI(span.Filename)),
			Range:    setProtocolRange(span.Begin, span.End),
			Message:  serr.Msg,
		}
		s.diags[uri] = append(s.diags[uri], diag)

	// Unknown errors and errors without location will become error notification.
	default:
		s.Fatal(context.TODO(), err.Error())

	}
}

func (s *Server) syncDiagnostics() {
	for k, v := range s.diags {
		s.client.PublishDiagnostics(context.TODO(), &protocol.PublishDiagnosticsParams{
			Diagnostics: v,
			URI:         protocol.DocumentURI(k),
		})
	}
}
