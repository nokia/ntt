package lsp

import (
	"context"

	"github.com/nokia/ntt/internal/errors"
	"github.com/nokia/ntt/internal/lsp/protocol"
)

// Diagnose runs various checks over a ttcn3 test suite.
//
// From LSP spec:
//
//     Diagnostics are "owned" by the server so it is the server's
//     responsibility to clear them if necessary.
//
//     If a language has a project system (for example C#) diagnostics are not
//     cleared when a file closes.  When a project is opened all diagnostics
//     for all files are recomputed (or read from a cache).
//
//     When a file changes it is the server’s responsibility to re-compute
//     diagnostics and push them to the client. If the computed set is empty it
//     has to push the empty array to clear former diagnostics. Newly pushed
//     diagnostics always replace previously pushed diagnostics. There is no
//     merging that happens on the client side.
func (s *Server) Diagnose() {
	s.diagsMu.Lock()
	defer s.diagsMu.Unlock()

	s.diags = make(map[string][]protocol.Diagnostic)
	defer s.syncDiagnostics()

	// Just do a very basic test if all .ttcn3 files are accessable.
	_, err := s.suite.Files()
	if err != nil {
		s.reportError(err)
		return
	}

}

func (s *Server) reportError(err error) {
	switch err := err.(type) {

	// Errors with a location will become diagnostics
	case errors.Error:
		uri := string(s.suite.File(err.Pos.Filename).URI())
		diag := protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      float64(err.Pos.Line - 1),
					Character: float64(err.Pos.Column - 1),
				},
				End: protocol.Position{
					Line:      float64(err.Pos.Line - 1),
					Character: float64(err.Pos.Column - 1),
				},
			},
			Severity: protocol.SeverityError,
			Source:   err.Pos.Filename,
			Message:  err.Msg,
		}
		s.diags[uri] = append(s.diags[uri], diag)

	// Expand error lists (like syntax error)
	case errors.ErrorList:
		for _, e := range err {
			s.reportError(e)
		}

	// Unknown errors and errors without location will become an error notification.
	default:
		s.Fatal(context.TODO(), err.Error())
	}
}

func (s *Server) syncDiagnostics() {
	for k, v := range s.diags {
		s.client.PublishDiagnostics(context.TODO(), &protocol.PublishDiagnosticsParams{
			Diagnostics: v,
			URI:         k,
		})
	}
}
