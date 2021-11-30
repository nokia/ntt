package lsp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/project"
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

	// Execute ntt build backend (k3-build)
	build_cmd := nttCommand(suite, "build")
	if s := env.Getenv("NTT_DEBUG"); s != "" {
		if s == "all" {
			build_cmd.Args = append(build_cmd.Args, "-vvvvv")
		} else {
			build_cmd.Args = append(build_cmd.Args, "-v")
		}
	}

	// Execute ntt run backend (k3s)
	cmd := nttCommand(suite, "run", "-j1", "--results-file=test_results.json", "--no-summary", "--no-color")
	if s := env.Getenv("NTT_DEBUG"); s != "" {
		cmd.Args = append(cmd.Args, "--debug")
	}
	cmd.Stdin = strings.NewReader(testID + "\n")
	s.Log(context.TODO(), fmt.Sprintf(`
===============================================================================
Executing test  : %q
compile command : %s
run command     : %s
cwd             : %s
===============================================================================`,
		testID, build_cmd.String(), cmd.String(), cmd.Dir))

	// compile test
	s.Log(context.TODO(), "compiling ...\n")
	out, err := build_cmd.CombinedOutput()
	s.Log(context.TODO(), string(out))

	if err != nil {
		s.Log(context.TODO(), err.Error())
		return err
	}
	s.Log(context.TODO(), build_cmd.ProcessState.String())

	// clean logs from all previous runs of actual test
	removeLogsForTest(s, filepath.Join(cmd.Dir, "logs"), testID)
	// run test
	s.Log(context.TODO(), "running ...\n")
	out, err = cmd.CombinedOutput()
	s.Log(context.TODO(), string(out))
	if err != nil {
		s.Log(context.TODO(), err.Error())
	} else {
		s.Log(context.TODO(), cmd.ProcessState.String())
	}

	// Continue anyway to execute ntt report
	cmd = nttCommand(suite, "report")
	out, err = cmd.CombinedOutput()
	s.Log(context.TODO(), string(out))
	if err != nil {
		s.Log(context.TODO(), err.Error())
		return err
	}

	// Display nice artifact overview for convenient navigation
	logDir := filepath.Join(cmd.Dir, "logs", testID+"-0")
	s.Log(context.TODO(), fmt.Sprintf(`
Content of log directory %q:
===============================================================================
%s`,
		logDir, strings.Join(fs.FindFilesRecursive(logDir), "\n")))
	return nil
}

// nttCommand will return a exec.Cmd for executing a ntt sub-command.
// It will set path to ntt-binary, proper test suite arguments, environ
// variables and working directory.
func nttCommand(suite *ntt.Suite, name string, opts ...string) *exec.Cmd {
	cmd := exec.Command(findExecutable())
	cmd.Args = append(cmd.Args, name)
	cmd.Args = append(cmd.Args, suiteArgs(suite)...)
	cmd.Args = append(cmd.Args, opts...)
	cmd.Env = os.Environ()
	for k, v := range map[string]string{
		"SCT_K3_SERVER": "ON",
		"NTT_COLORS":    "never",
		"NTT_CACHE":     strings.Join(cacheDirs(suite), ":"),
		"K3CFLAGS":      k3cFlags(suite),
	} {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Dir = filepath.Join(fs.Path(suite.Root()), "ntt.test")
	os.Mkdir(cmd.Dir, 0755)
	return cmd
}

// remove log files from previous runs
func removeLogsForTest(s *Server, baseDir string, tcName string) {
	files, _ := filepath.Glob(filepath.Join(baseDir, tcName+"-*"))
	for _, f := range files {
		if err := os.RemoveAll(f); err != nil {
			s.Log(context.TODO(), fmt.Sprintf("error removing %q: %s", f, err.Error()))
		}
	}
}

// suiteArgs returns the suite arguments require to execute ntt commands. Which
// is either the test suite root folder or a list of TTCN-3 files.
func suiteArgs(suite *ntt.Suite) []string {
	if root := fs.Path(suite.Root()); root != "" {
		return []string{root}
	}
	return project.FindAllFiles(suite)
}

// cacheDirs returns directories of interest. Such as parameters dir or build dir.
func cacheDirs(suite *ntt.Suite) []string {
	dirs := strings.Split(env.Getenv("NTT_CACHE"), ":")

	// Large test suites usually require some environment setup to function correctly.
	if path := fs.FindK3EnvInCurrPath(fs.Path(suite.Root())); path != "" {
		dirs = append(dirs, path)
	}

	// Test suite might store additional configuration in their parameter files.
	if path, _ := suite.ParametersDir(); path != "" {
		dirs = append(dirs, path)
	}
	return dirs
}

// k3cFlags expand k3.env variable K3CFLAGS with a flag to
// disable color output.
//
// VSCode Ouput channel has only poor support for ANSI color sequences.
// It's best we disable colors all together.
func k3cFlags(suite *ntt.Suite) string {
	flags := "--diagnostics-color=never"
	if ext, _ := suite.Getenv("K3CFLAGS"); ext != "" {
		flags = ext + " " + flags
	}
	return flags
}

// findExecutable looks for ntt binary in various locations.
//
// If no ntt binary is available findExecutable will return the path of the
// current executable. This situation happens (for example) when users use a
// automatically installed ntt binary with no prefix root environment loaded.
func findExecutable() string {
	if exe, err := exec.LookPath("ntt"); err == nil {
		return exe
	}
	if exe, err := os.Executable(); err == nil {
		return exe
	}
	return "ntt"
}
