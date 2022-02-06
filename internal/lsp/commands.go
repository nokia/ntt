package lsp

import (
	"context"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/lsp/protocol"
)

// NewCommand returns a CodeLens command.
func NewCommand(pos loc.Position, title string, command string, args ...interface{}) (protocol.CodeLens, error) {
	b, err := marshalRaw(args...)
	if err != nil {
		return protocol.CodeLens{}, err
	}
	return protocol.CodeLens{
		Range: protocol.Range{
			Start: position(pos.Line, pos.Column),
			End:   position(pos.Line, pos.Column),
		},
		Command: protocol.Command{
			Title:     "run test",
			Command:   "ntt.test",
			Arguments: b,
		},
	}, nil
}

type nttTestParams struct {
	ID   string // Fully qualified testcase identifier
	URI  string // URL points to the ttcn3 source file containing the testcase
	Stop bool   // Stop the test
}

func (s *Server) executeCommand(ctx context.Context, params *protocol.ExecuteCommandParams) (interface{}, error) {
	switch params.Command {
	case "ntt.debug.toggle":
		return s.toggleDebug(ctx)
	case "ntt.status":
		return s.status(ctx)
	case "ntt.test":
		var param nttTestParams
		if err := unmarshalRaw(params.Arguments, &param); err != nil {
			return nil, err
		}

		// A file might belong to multiple test suites, possibly. But
		// it's sufficient to just take the first test suite available,
		// practically.
		suite, err := s.FirstSuite(param.URI)
		if err != nil {
			return nil, err
		}

		return nil, s.testCtrl.RunTest(suite, param.ID, s)
	}
	return nil, nil
}
