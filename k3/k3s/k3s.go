package k3s

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/project"
)

type runner struct {
	// Project.Interface provides files and root directory.
	p *project.Config

	// Working directory
	Dir string
}

// Build calls ntt build to regenerate or rebuild changed TTCN-3 source files.
func Build(w io.Writer, p *project.Config) error {
	// Find a nice working directory to put logs and other artifacts in it.
	dir, err := nttWorkingDir(p)
	if err != nil {
		return err
	}

	// Rebuild the test executable and required adapters first.
	cmd := nttCommand(p, "build")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	w.Write([]byte(cmd.String()))
	w.Write(out)
	if err != nil {
		return &BuildError{Cmd: cmd, Err: err}
	}
	return nil
}

// Run a test using k3s
func Run(w io.Writer, p *project.Config, testID string) (string, error) {
	// Find a nice working directory to put logs and other artifacts in it.
	dir, err := nttWorkingDir(p)
	if err != nil {
		return "", err
	}

	r := &runner{
		p:   p,
		Dir: dir,
	}
	return dir, r.Run(w, testID)
}

func (r *runner) Run(w io.Writer, testID string) error {

	// Clear any previous artifacts.
	r.clean(testID)

	// Execute test (k3s backend)
	cmd := nttCommand(r.p, "run", "--", testID)
	cmd.Dir = r.Dir
	cmd.Env = append(cmd.Env, "SCT_K3_SERVER=ON")
	out, err := cmd.CombinedOutput()
	w.Write(out)
	return multierror.Append(err, r.report(w, testID)).ErrorOrNil()
}

func (r *runner) report(w io.Writer, testID string) error {

	// Display a nice summary
	cmd := nttCommand(r.p, "report")
	cmd.Dir = r.Dir
	out, err := cmd.CombinedOutput()
	w.Write(out)
	return err
}

// clean removes all artifacts of testID from the working directory.
func (r *runner) clean(testID string) {
	files, _ := filepath.Glob(filepath.Join(r.Dir, "logs", testID+"-*"))
	for _, f := range files {
		if err := os.RemoveAll(f); err != nil {
			log.Debugf("Removing %q failed: %s", f, err)
		}
	}
}

func (r *runner) LogDir(testID string) string {
	return filepath.Join(r.Dir, "logs", testID+"-0")
}

// nttWorkingDir returns a working directory for ntt artifacts.
func nttWorkingDir(p *project.Config) (string, error) {

	dir := filepath.Join(fs.Path(p.Root), "ntt.test")
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		log.Debugf("Creating directory %q failed: %s", dir, err.Error())
		dir, err = ioutil.TempDir("", "ntt-run-")
	}

	log.Debugf("Using working directory %q", dir)
	return dir, err
}

// nttCommand will return a exec.Cmd for executing a ntt sub-command.
// This convencience function sets up common configuration:
// It will set path to ntt-binary, sets proper test suite arguments, environ
// variables and working directory.
func nttCommand(p *project.Config, cmdName string, opts ...string) *exec.Cmd {
	cmd := exec.Command(nttExecutable())
	cmd.Env = nttEnv(p)

	// ntt commands have a common format:
	//     ntt <cmdName> [<args>] [<opts>]
	cmd.Args = append(cmd.Args, cmdName)
	cmd.Args = append(cmd.Args, nttArgs(p)...)
	cmd.Args = append(cmd.Args, opts...)

	if debug := env.Getenv("NTT_DEBUG"); debug != "" {
		if strings.TrimSpace(strings.ToLower(debug)) == "all" {
			cmd.Args = append(cmd.Args, "-vvvvv")
		} else {
			cmd.Args = append(cmd.Args, "-v")
		}
	}

	return cmd
}

// nttExecutable returns the path to the ntt executable.
//
// If no ntt binary is available nttExecutable will return the path of the
// current executable. This situation happens (for example) when users use a
// automatically installed ntt binary with no prefix root environment loaded.
func nttExecutable() string {
	if exe, err := exec.LookPath("ntt"); err == nil {
		return exe
	}
	if exe, err := os.Executable(); err == nil {
		return exe
	}
	return "ntt"
}

// nttArgs returns the project root directory. If the project has no root
// directory nttArgs will return a list of all project source files.
func nttArgs(p *project.Config) []string {
	if root := p.Root; root != "" {
		return []string{root}
	}

	srcs, _ := project.Files(p)
	return srcs
}

// nttEnv returns the magic environment variables required to use ntt with
// Nokia component tests.
// nttEnv also copies os.Environ for variables like PATH and LD_LIBRARY_PATH
// required by various scripts and C++ applications.
func nttEnv(p *project.Config) []string {

	// SCTs use environment variable `NTT_CACHE` to find various required
	// files. For example file `k3.env`, which is required for the SCTs to
	// function correctly.
	return append(os.Environ(), "NTT_CACHE="+strings.Join(func(p *project.Config) []string {
		dirs := strings.Split(env.Getenv("NTT_CACHE"), ":")
		if path := fs.FindK3EnvInCurrPath(fs.Path(p.Root)); path != "" {
			dirs = append(dirs, path)
		}
		return dirs
	}(p), ":"))
}

type BuildError struct {
	Cmd *exec.Cmd
	Err error
}

func (e *BuildError) Error() string {
	return fmt.Sprintf("%s: %s", e.Cmd.String(), e.Err)
}

func (e *BuildError) Unwrap() error {
	return e.Err
}
