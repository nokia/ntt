package k3r

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nokia/ntt/control"
	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/proc"
	"github.com/nokia/ntt/internal/session"
)

// A Test is a test case instance.
type Test struct {
	*control.Job

	// Path to the T3XF file.
	T3XF string

	// Path to the K3 runtime
	Runtime string

	// LogFile is the path to the log file.
	LogFile string
}

var (
	ErrUnknown        = fmt.Errorf("unknown error")
	ErrNotImplemented = fmt.Errorf("not implemented")
	ErrTimeout        = fmt.Errorf("timeout")
	ErrInvalidMessage = fmt.Errorf("invalid message")

	ErrNoSuch        = fmt.Errorf("no such thing")
	ErrNoSuchModule  = fmt.Errorf("no such module")
	ErrNoSuchTest    = fmt.Errorf("no such test case")
	ErrNoSuchControl = fmt.Errorf("no such control")

	ErrRuntimeNotReady = fmt.Errorf("runtime not ready")
	ErrModuleNotReady  = fmt.Errorf("module not ready")
	ErrTestNotReady    = fmt.Errorf("test case not ready")
	ErrControlNotReady = fmt.Errorf("control not ready")
	ErrNotQualified    = fmt.Errorf("id not fully qualified")
)

type Error struct {
	Err error
}

func (r *Error) Error() string {
	return r.Unwrap().Error()
}

func (r *Error) Unwrap() error {
	return r.Err
}

func (t *Test) Run(ctx context.Context) <-chan control.Event {

	events := make(chan control.Event)

	go func() {
		defer close(events)

		if !strings.Contains(t.Name, ".") {
			events <- control.NewErrorEvent(ErrNotQualified)
			return
		}

		if t.LogFile == "" {
			t.LogFile = fmt.Sprintf("%s.log", fs.Stem(t.T3XF))
		}

		get := func(name string) string {
			if s, ok := env.LookupEnv(name); ok {
				return s
			}
			if t.Config != nil {
				return t.Config.Variables[name]
			}
			return ""
		}
		cmd := proc.CommandContext(ctx, t.Runtime, t.T3XF, "-o", t.LogFile)
		if addr := os.Getenv("GDB_SERVER"); addr != "" {
			gdb := proc.CommandContext(ctx, "gdbserver", "--once", addr)
			gdb.Args = append(gdb.Args, cmd.Args...)
			cmd = gdb
		}
		if s := get("K3RFLAGS"); s != "" {
			cmd.Args = append(cmd.Args, strings.Fields(s)...)
		}
		cmd.Dir = t.Dir
		env, err := buildEnv(t)
		if err != nil {
			events <- control.NewErrorEvent(err)
			return
		}
		cmd.Env = env

		sid, err := session.Get()
		if err != nil {
			events <- control.NewErrorEvent(err)
			return
		}
		defer session.Release(sid)
		cmd.Env = append(cmd.Env, fmt.Sprintf("NTT_SESSION_ID=%d", sid))

		cmd.Stdin = strings.NewReader(t.request())
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			events <- control.NewErrorEvent(&Error{Err: err})
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			events <- control.NewErrorEvent(&Error{Err: err})
			return
		}
		log.Debugf("env:\n")
		for _, e := range cmd.Env {
			log.Debugf("  %s\n", e)
		}
		log.Debugf("+ %s\n", cmd.String())
		err = cmd.Start()
		if err != nil {
			events <- control.NewErrorEvent(&Error{Err: err})
			return
		}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stderr)
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				events <- control.NewLogEvent(t.Job, "k3r: "+scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				events <- control.NewErrorEvent(&Error{Err: err})
			}
		}()

		// k3r does not have a ControlStarted event, so we'll fake it.
		if strings.HasSuffix(t.Name, ".control") {
			events <- control.NewStartEvent(t.Job, t.Name)
		}

		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanLines)
		var name string
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			log.Traceln(context.TODO(), ">", line)
			switch v := strings.Fields(line); v[0] {
			case "tciTestCaseStarted":
				name = v[1][1 : len(v[1])-1]
				events <- control.NewStartEvent(t.Job, name)
			case "tciTestCaseTerminated":
				verdict := v[1]
				events <- control.NewStopEvent(t.Job, name, verdict)
			case "tciControlTerminated":
				// k3r does not send a verdict. I just assume it's pass.
				events <- control.NewStopEvent(t.Job, t.Name, "pass")
			case "tciError":
				switch v[1] {
				case "E101:":
					events <- control.NewErrorEvent(ErrNoSuchModule)
				case "E102:":
					events <- control.NewErrorEvent(ErrNoSuchTest)
				case "E103:":
					events <- control.NewErrorEvent(ErrNoSuchControl)
				case "E200:":
					events <- control.NewErrorEvent(ErrRuntimeNotReady)
				case "E201:":
					// ErrModuleNotReady happens, when tciRootModule was not called.
					// This is a spurious error, so we'll ignore it.
					//events <- control.NewErrorEvent(ErrModuleNotReady)
				case "E202:":
					events <- control.NewErrorEvent(ErrTestNotReady)
				case "E203:":
					events <- control.NewErrorEvent(ErrControlNotReady)
				case "E999:":
					events <- control.NewErrorEvent(fmt.Errorf("%w: %s", ErrNotImplemented, strings.Join(v[2:], " ")))
				default:
					events <- control.NewErrorEvent(fmt.Errorf("%w: %s", ErrUnknown, strings.Join(v[1:], " ")))
				}
			default:
				events <- control.NewErrorEvent(fmt.Errorf("%w: %s", ErrInvalidMessage, line))
			}
		}
		if err := scanner.Err(); err != nil {
			events <- control.NewErrorEvent(&Error{Err: err})
		}
		wg.Wait()
		err = waitGracefully(t, cmd)
		if ctx.Err() == context.DeadlineExceeded {
			events <- control.NewErrorEvent(ErrTimeout)
		} else if err != nil {
			events <- control.NewErrorEvent(&Error{Err: err})
		}
	}()

	return events
}

func buildEnv(t *Test) ([]string, error) {
	var (
		ret            []string
		k3rPaths       []string
		ldLibraryPaths []string
		binPaths       []string
	)

	if t.Config != nil {
		// K3_NAME is required by test hooks (aka. sut-control.sh)
		ret = append(ret, "K3_NAME="+t.Config.Name)

		// All declared (environment) variables are passed to k3r.
		for k, v := range t.Config.Variables {
			ret = append(ret, fmt.Sprintf("%s=%s", k, v))
		}

		var commonPaths []string

		// NTT_CACHE specifies the location of cached artifacts, such
		// as additional libraries. And should be looked up before the
		// current working directory.
		if s, ok := env.LookupEnv("NTT_CACHE"); ok {
			commonPaths = append(commonPaths, s)
		}

		// Artifacts without explicit path are put in the current
		// working directory by convention.
		//
		// Note: We use "." instead of os.Getwd() to make our tests
		// easier to write.
		commonPaths = append(commonPaths, ".")

		// Imports may provide additional artifacts.
		commonPaths = append(commonPaths, t.Config.Manifest.Imports...)

		k3rPaths = append(commonPaths, t.Config.K3.Plugins...)
		ldLibraryPaths = append(commonPaths, t.Config.K3.CLibDirs...)
		binPaths = commonPaths
		if t.Config.K3.Runtime != "k3r" && t.Config.K3.Runtime != "" {
			binPaths = append(binPaths, filepath.Dir(t.Config.K3.Runtime))
		}
	}

	ret = append(ret, t.Env...)
	ret = append(ret, buildEnvPaths("K3R_PATH", k3rPaths...))
	ret = append(ret, buildEnvPaths("LD_LIBRARY_PATH", ldLibraryPaths...))
	ret = append(ret, buildEnvPaths("PATH", binPaths...))

	for k, v := range t.ModulePars {
		ret = append(ret, fmt.Sprintf("%s=%s", k, v))
	}

	ret = append(ret, "K3_SERVER=pipe,/dev/fd/0,/dev/fd/1")
	return ret, nil
}

// buildEnvPaths returns environment variable 'name' with given search paths appended.
func buildEnvPaths(name string, paths ...string) string {
	if env, ok := env.LookupEnv(name); ok {
		paths = append([]string{env}, paths...)
	}
	return fmt.Sprintf("%s=%s", name, strings.Join(paths, string(os.PathListSeparator)))
}

// waitGracefully waits for the k3 runtime to exit, checks if the log file is
// truncated and stores the k3 exit code in the log-file for convenient retrieval.
func waitGracefully(t *Test, cmd *exec.Cmd) error {
	cmdErr := cmd.Wait()
	f, err := os.OpenFile(filepath.Join(t.Dir, t.LogFile), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if ofs, err := f.Seek(-1, io.SeekEnd); err == nil {
		b := make([]byte, 1)
		if _, err = f.ReadAt(b, ofs); err != nil {
			return err
		}
		if b[0] != '\n' {
			f.Seek(0, io.SeekEnd)
			f.Write([]byte{'\n'})
		}
	}
	f.Seek(0, io.SeekEnd)
	fmt.Fprintf(f, "%s.999999|exit|%d\n", time.Now().Format("20060102T150405"), cmd.ProcessState.ExitCode())
	return cmdErr
}

// request builds a request for running a test or control part.
func (t *Test) request() string {
	var req strings.Builder
	v := strings.SplitN(t.Name, ".", 2)
	if len(v) == 2 {
		fmt.Fprintln(&req, "tciRootModule", v[0])
	}
	if name := v[len(v)-1]; name == "control" {
		fmt.Fprintln(&req, "tciStartControl")
	} else {
		fmt.Fprintf(&req, "tciStartTestCase \"%s\" {%s}\n", t.Name, strings.Join(t.Args, ","))
	}
	s := req.String()
	log.Traceln(context.TODO(), "<", s)
	return s
}

// NewTest creates a new test instance from job ready to run.
func NewTest(t3xf string, job *control.Job) *Test {
	return &Test{
		Job:     job,
		T3XF:    t3xf,
		Runtime: job.Config.K3.Runtime,
	}
}
