package lsp

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/lsp/protocol"
)

func (s *Server) executeCommand(ctx context.Context, params *protocol.ExecuteCommandParams) (interface{}, error) {
	switch params.Command {
	case "ntt.status":
		return s.status(ctx)
	case "ntt.test":
		var testID string
		if err := unmarshalRaw(params.Arguments, &testID); err != nil {
			return nil, err
		}
		return nil, cmdTest(s, testID)
	}
	return nil, nil
}
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

func cmdTest(s *Server, testId string) error {
	if cwd, err := os.Getwd(); err == nil {
		s.Log(context.TODO(), fmt.Sprintf("Current working directory: %q", cwd))
	}
	cmd := exec.Command("ntt", "run", "--", testId)
	s.Log(context.TODO(), fmt.Sprint("Executing: ", cmd.String()))
	out, err := cmd.CombinedOutput()
	if err != nil {
		s.Log(context.TODO(), err.Error())
		return err
	}
	s.Log(context.TODO(), string(out))
	return nil
}
