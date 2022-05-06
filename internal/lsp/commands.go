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
	TestID      // TestID identifies a testcase in a ttch3 source file.
	Stop   bool // Stop the test
}

func (s *Server) executeCommand(ctx context.Context, params *protocol.ExecuteCommandParams) (interface{}, error) {
	switch params.Command {
	case "ntt.debug.toggle":
		return s.toggleDebug(ctx)
	case "ntt.status":
		return s.status(ctx)
	case "ntt.test":
		var tp nttTestParams
		if err := unmarshalRaw(params.Arguments, &tp); err != nil {
			return nil, err
		}

		return nil, s.testCtrl.RunTest(tp.TestID)
	}
	return nil, nil
}
