package k3r

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/proc"
	"github.com/nokia/ntt/internal/session"
	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/tests"
)

// A Test is a test case instance.
type Test struct {
	*tests.Job

	// Path to the T3XF file.
	T3XF string

	// Path to the K3 runtime
	Runtime string

	// Dir specifies the working directory for the test.
	Dir string

	// LogFile is the path to the log file.
	LogFile string

	// Stderr is the stderr of the test.
	Stderr bytes.Buffer
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

func (t *Test) Run() <-chan tests.Event {
	return t.RunWithContext(context.Background())
}

func (t *Test) RunWithContext(ctx context.Context) <-chan tests.Event {

	events := make(chan tests.Event)

	go func() {
		defer close(events)

		if !strings.Contains(t.Name, ".") {
			events <- tests.NewErrorEvent(ErrNotQualified)
			return
		}

		sid, err := session.Get()
		if err != nil {
			events <- tests.NewErrorEvent(err)
			return
		}
		defer session.Release(sid)

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
		if s := get("K3RFLAGS"); s != "" {
			cmd.Args = append(cmd.Args, strings.Fields(s)...)
		}
		cmd.Dir = t.Dir
		cmd.Env = append(t.Env,
			"K3_SERVER=pipe,/dev/fd/0,/dev/fd/1",
			fmt.Sprintf("K3_SESSION_ID=%d", sid),
		)

		for k, v := range t.ModulePars {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Stdin = strings.NewReader(t.request())
		cmd.Stderr = &t.Stderr
		stdout, err := cmd.StdoutPipe()
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
		err = waitGracefully(t, cmd)
		if ctx.Err() == context.DeadlineExceeded {
			events <- tests.NewErrorEvent(ErrTimeout)
		} else if err != nil {
			events <- tests.NewErrorEvent(&Error{Err: err})
		}
	}()

	return events
}

// waitGracefully waits for the k3 runtime to exit, checks if the log file is
// truncated and stores the k3 exit code in the log-file for convenient retrieval.
func waitGracefully(t *Test, cmd *exec.Cmd) error {
	cmdErr := cmd.Wait()
	f, err := os.OpenFile(t.LogFile, os.O_RDWR|os.O_CREATE, 0644)
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
