package k3r

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/proc"
	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/tests"
)

// A Test is a test case instance.
type Test struct {
	*tests.Job

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

func (t *Test) Run(ctx context.Context) <-chan tests.Event {

	events := make(chan tests.Event)

	go func() {
		defer close(events)

		if !strings.Contains(t.Name, ".") {
			events <- tests.NewErrorEvent(ErrNotQualified)
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
			events <- tests.NewErrorEvent(err)
			return
		}
		cmd.Env = env
		cmd.Stdin = strings.NewReader(t.request())
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			events <- tests.NewErrorEvent(&Error{Err: err})
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			events <- tests.NewErrorEvent(&Error{Err: err})
			return
		}
		log.Debugf("+ %s\n", cmd.String())
		err = cmd.Start()
		if err != nil {
			events <- tests.NewErrorEvent(&Error{Err: err})
			return
		}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stderr)
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				events <- tests.NewLogEvent(t.Job, "k3r: "+scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				events <- tests.NewErrorEvent(&Error{Err: err})
			}
		}()

		// k3r does not have a ControlStarted event, so we'll fake it.
		if strings.HasSuffix(t.Name, ".control") {
			events <- tests.NewStartEvent(t.Job, t.Name)
		}

		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanLines)
		var name string
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			log.Traceln(">", line)
			switch v := strings.Fields(line); v[0] {
			case "tciTestCaseStarted":
				name = v[1][1 : len(v[1])-1]
				events <- tests.NewStartEvent(t.Job, name)
			case "tciTestCaseTerminated":
				verdict := v[1]
				events <- tests.NewStopEvent(t.Job, name, verdict)
			case "tciControlTerminated":
				// k3r does not send a verdict. I just assume it's pass.
				events <- tests.NewStopEvent(t.Job, t.Name, "pass")
			case "tciError":
				switch v[1] {
				case "E101:":
					events <- tests.NewErrorEvent(ErrNoSuchModule)
				case "E102:":
					events <- tests.NewErrorEvent(ErrNoSuchTest)
				case "E103:":
					events <- tests.NewErrorEvent(ErrNoSuchControl)
				case "E200:":
					events <- tests.NewErrorEvent(ErrRuntimeNotReady)
				case "E201:":
					// ErrModuleNotReady happens, when tciRootModule was not called.
					// This is a spurious error, so we'll ignore it.
					//events <- tests.NewErrorEvent(ErrModuleNotReady)
				case "E202:":
					events <- tests.NewErrorEvent(ErrTestNotReady)
				case "E203:":
					events <- tests.NewErrorEvent(ErrControlNotReady)
				case "E999:":
					events <- tests.NewErrorEvent(fmt.Errorf("%w: %s", ErrNotImplemented, strings.Join(v[2:], " ")))
				default:
					events <- tests.NewErrorEvent(fmt.Errorf("%w: %s", ErrUnknown, strings.Join(v[1:], " ")))
				}
			default:
				events <- tests.NewErrorEvent(fmt.Errorf("%w: %s", ErrInvalidMessage, line))
			}
		}
		if err := scanner.Err(); err != nil {
			events <- tests.NewErrorEvent(&Error{Err: err})
		}
		wg.Wait()
		err = waitGracefully(t, cmd)
		if ctx.Err() == context.DeadlineExceeded {
			events <- tests.NewErrorEvent(ErrTimeout)
		} else if err != nil {
			events <- tests.NewErrorEvent(&Error{Err: err})
		}
	}()

	return events
}

func buildEnv(t *Test) ([]string, error) {
	var ret []string

	paths, err := project.LibraryPaths(t.Config)
	if err != nil {
		return nil, err
	}

	// Prepend current working directory and NTT_CACHE envirnoment variable
	// to library paths.
	paths = append([]string{"."}, paths...)
	if s, ok := env.LookupEnv("NTT_CACHE"); ok {
		paths = append([]string{s}, paths...)
	}

	for _, p := range []string{"K3R_PATH", "LD_LIBRARY_PATH", "PATH"} {
		if s := buildEnvPaths(p, paths); s != "" {
			ret = append(ret, s)
		}
	}

	if t.Config != nil {
		for k, v := range t.Config.Variables {
			if _, ok := env.LookupEnv(k); !ok {
				ret = append(ret, fmt.Sprintf("%s=%s", k, v))
			}
		}
	}

	ret = append(ret, t.Env...)

	for k, v := range t.ModulePars {
		ret = append(ret, fmt.Sprintf("%s=%s", k, v))
	}

	ret = append(ret, "K3_SERVER=pipe,/dev/fd/0,/dev/fd/1")
	return ret, nil
}

// buildEnvPaths returns environment variable s with given search paths prepended.
func buildEnvPaths(s string, paths []string) string {
	if env, ok := env.LookupEnv(s); ok {
		paths = append(paths, env)
	}
	if len(paths) != 0 {
		return fmt.Sprintf("%s=%s", s, strings.Join(paths, string(os.PathListSeparator)))
	}
	return ""
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
	if ofs, err := f.Seek(-1, os.SEEK_END); err == nil {
		b := make([]byte, 1)
		if _, err = f.ReadAt(b, ofs); err != nil {
			return err
		}
		if b[0] != '\n' {
			f.Seek(0, os.SEEK_END)
			f.Write([]byte{'\n'})
		}
	}
	f.Seek(0, os.SEEK_END)
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
	log.Traceln("<", s)
	return s
}

// NewTest creates a new test instance ready to run.
func NewTest(t3xf string, name string) *Test {
	return &Test{
		Job: &tests.Job{
			Name: name,
		},
		T3XF:    t3xf,
		Runtime: k3.Runtime(),
	}
}
