package lsp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
)

type param struct {
	ID  string // Fully qualified testcase identifier
	URI string // URL points to the ttcn3 source file containing the testcase
}

const separator string = "==============================================================================="

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

// getenvFromSuite invokes ntt show command to retrieve
// environment variables from the provided suite
func getenvFromSuite(nttCache string, pathToManifest string, evName string) string {
	cmd := exec.Command("ntt", "show", pathToManifest, "--", evName)
	cmd.Env = os.Environ()
	if nttCache != "" {
		cmd.Env = append(cmd.Env, "NTT_CACHE="+nttCache)
	}
	out, err := cmd.CombinedOutput()
	v := strings.TrimSuffix(string(out), "\n")
	log.Debug(fmt.Sprintf("%s=%q", evName, v))
	if err != nil {
		log.Debug(err.Error())
	}
	return v
}

func cmdTest(s *Server, testId string, fileUri string) error {
	var nttCache, nttDebug, pathToManifest string
	var cmd *exec.Cmd = nil
	log.Debug(fmt.Sprintf("testcase file uri: %q", fileUri))

	suites := s.Owners(protocol.DocumentURI(fileUri))
	if len(suites) > 0 {
		pathToManifest = suites[0].Root().Path()

		if k3EnvPath := fs.FindK3EnvInCurrPath(pathToManifest); len(k3EnvPath) > 0 {
			nttCache = k3EnvPath
			os.Mkdir(pathToManifest+"/ntt.test", 0744)
			if err := os.Chdir(pathToManifest + "/ntt.test"); err != nil {
				s.Log(context.TODO(), fmt.Sprintf("Could not change Current working directory: %q: %q", pathToManifest+"/ntt.test", err))
			} else {
				nttCache = nttCache + ":" + pathToManifest + "/ntt.test"
			}
			if path, err := suites[0].ParametersDir(); path != "" {
				nttCache = nttCache + ":" + path
			} else if err != nil {
				s.Log(context.TODO(), fmt.Sprintf("Error while extracting parameters_dir from manifest: %q", err))
			}
		}
	}
	nttDebug = getenvFromSuite(nttCache, pathToManifest, "NTT_DEBUG")
	log.Debug(fmt.Sprintf("NTT_CACHE=%q\nNTT_DEBUG=%q", nttCache, nttDebug))

	var opts = []string{"run", pathToManifest, "-j1", "--results-file=test_results.json", "--no-summary"}
	if nttDebug == "all" {
		opts = append(opts, "--debug")
	}
	cmd = exec.Command("ntt", opts...)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "SCT_K3_SERVER=ON")
	if nttCache != "" {
		cmd.Env = append(cmd.Env, "NTT_CACHE="+nttCache)
	}
	cmd.Stdin = strings.NewReader(testId + "\n")

	s.Log(context.TODO(), fmt.Sprintf("%s\nExecuting test : %q\nwith command   : %s\ncwd            : %s\n%s",
		separator, testId, cmd.String(), pathToManifest+"/ntt.test", separator))
	out, err := cmd.CombinedOutput()
	s.Log(context.TODO(), string(out))
	if err == nil {
		s.Log(context.TODO(), cmd.ProcessState.String())
		if cmd.ProcessState.ExitCode() >= 0 {
			// run ntt report
			cmd := exec.Command("ntt", "report", pathToManifest)
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, "NTT_COLORS=never")
			if nttCache != "" {
				cmd.Env = append(cmd.Env, "NTT_CACHE="+nttCache)
			}
			out, err := cmd.CombinedOutput()
			s.Log(context.TODO(), string(out))
			if err != nil {
				s.Log(context.TODO(), err.Error())
			}
			logDir := "./logs/" + testId + "-0"
			s.Log(context.TODO(), fmt.Sprintf("Content of the log directory %q:\n%s", logDir, strings.Join(fs.FindFilesRecursive(logDir), "\n")+"\n"+separator))
		}
	} else {
		s.Log(context.TODO(), err.Error())
	}
	return err
}
