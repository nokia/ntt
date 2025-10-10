package lsp

import (
	"context"
	"errors"

	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/ttcn3/syntax"
)

// NewCommand returns a CodeLens command.
func NewCommand(pos syntax.Position, title string, command string, args ...interface{}) (protocol.CodeLens, error) {
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

func (s *Server) executeCommand(ctx context.Context, params *protocol.ExecuteCommandParams) (interface{}, error) {
	switch params.Command {
	case "ntt.debug.toggle":
		return s.toggleDebug(ctx)
	case "ntt.status":
		return s.status(ctx)
	case "ntt.test":
		return nil, errors.New("command ntt.test: not implemented")
	}
	return nil, nil
}
