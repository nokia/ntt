package lsp

import (
	"context"
	"fmt"
	"strings"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/runner/k3s"
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
	ID  string // Fully qualified testcase identifier
	URI string // URL points to the ttcn3 source file containing the testcase
}

func (s *Server) executeCommand(ctx context.Context, params *protocol.ExecuteCommandParams) (interface{}, error) {
	switch params.Command {
	case "ntt.status":
		return s.status(ctx)
	case "ntt.test":
		var param nttTestParams
		if err := unmarshalRaw(params.Arguments, &param); err != nil {
			return nil, err
		}
		return nil, nttTest(s, param.URI, param.ID)
	}
	return nil, nil
}

func nttTest(s *Server, fileURI string, testID string) error {

	// A file might belong to multiple test suites, possibly. But
	// it's sufficient to just take the first test suite available,
	// practically.
	suite, err := s.FirstSuite(fileURI)
	if err != nil {
		return err
	}

	log.Printf(`
===============================================================================
Compiling test %s in %q`, testID, suite.Root())

	r, err := k3s.New(s, suite)
	if err != nil {
		s.Log(context.TODO(), err.Error())
		return err
	}

	log.Printf(`
===============================================================================
Running test %s in %q`, testID, suite.Root())

	err = r.Run(s, testID)

	// Show a directory listing of the artifacts (independently of any test errors)
	logDir := r.LogDir(testID)
	if files := fs.Abs(fs.FindFilesRecursive(logDir)...); len(files) > 0 {
		s.Log(context.Background(), fmt.Sprintf(`
Content of log directory %q:
===============================================================================
%s\n\n`,
			logDir, strings.Join(files, "\n")))
	}

	return err
}
