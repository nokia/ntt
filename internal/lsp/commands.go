package lsp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/lsp/protocol"
)

type param struct {
	ID  string // Fully qualified testcase identifier
	URI string // URL points to the ttcn3 source file containing the testcase
}

func (s *Server) executeCommand(ctx context.Context, params *protocol.ExecuteCommandParams) (interface{}, error) {
	switch params.Command {
	case "ntt.status":
		return s.status(ctx)
	case "ntt.test":
		var decParam param
		if err := unmarshalRaw(params.Arguments, &decParam); err != nil {
			return nil, err
		}
		return nil, cmdTest(s, decParam.ID, decParam.URI)
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

func cmdTest(s *Server, testId string, fileUri string) error {
	var nttCache, pathToManifest string
	s.Log(context.TODO(), fmt.Sprintf("testcase file uri: %q", fileUri))
	if cwd, err := os.Getwd(); err == nil {
		s.Log(context.TODO(), fmt.Sprintf("Current working directory: %q", cwd))
	}
	suites := s.Owners(protocol.DocumentURI(fileUri))
	if len(suites) > 0 {
		pathToManifest = suites[0].Root().Path()

		if k3EnvPath := fs.FindK3EnvInCurrPath(pathToManifest); len(k3EnvPath) > 0 {
			nttCache = k3EnvPath
			os.Mkdir(pathToManifest+"/ntt.test", 0744)
			if err := os.Chdir(pathToManifest + "/ntt.test"); err != nil {
				s.Log(context.TODO(), fmt.Sprintf("Could not change Current working directory: %q: %q", pathToManifest+"/ntt.test", err))
			} else {
				s.Log(context.TODO(), fmt.Sprintf("Changed Current working directory: %q", pathToManifest+"/ntt.test"))
				nttCache = nttCache + ":" + pathToManifest + "/ntt.test"
			}
			if path, err := suites[0].ParametersDir(); path != "" {
				nttCache = nttCache + ":" + path
			} else if err != nil {
				s.Log(context.TODO(), fmt.Sprintf("Error while extracting parameters_dir from manifest: %q", err))
			}
		}

		s.Log(context.TODO(), fmt.Sprintf(" NTT_CACHE: %v", nttCache))
	}
	cmd := exec.Command("ntt", "run", pathToManifest, "-j1", "--debug", "--", testId)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "SCT_K3_SERVER=ON")
	if nttCache != "" {
		cmd.Env = append(cmd.Env, "NTT_CACHE="+nttCache)
	}
	cmd.Stdin = strings.NewReader(testId + "\n")
	s.Log(context.TODO(), fmt.Sprint("Executing: ", cmd.String()))
	out, err := cmd.CombinedOutput()
	s.Log(context.TODO(), string(out))
	if err != nil {
		s.Log(context.TODO(), err.Error())
	}
	return err
}
