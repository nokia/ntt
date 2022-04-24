package run

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
)

// A Test is a test case instance.
type Test struct {

	// Name is the fully qualified name of the test.
	Name string

	// Args is the list of arguments to pass to the test.
	Args []string

	// ModulePars hold module parameters
	ModulePars map[string]string

	// Path to the T3XF file.
	T3XF string

	// Path to the K3 runtime
	Runtime string

	// Dir specifies the working directory for the test.
	Dir string

	// LogFile is the path to the log file.
	LogFile string

	// Env specifies the environment variables to pass to the test.
	Env []string

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

type RuntimeError struct {
	Err error
}

func (r *RuntimeError) Error() string {
	return r.Unwrap().Error()
}

func (r *RuntimeError) Unwrap() error {
	return r.Err
}

// EventType is the type of an event.
type EventType int

const (
	Error = EventType(iota)
	TestStarted
	TestTerminated
	ControlStarted
	ControlTerminated
)

func (t EventType) String() string {
	switch t {
	case Error:
		return "tciError"
	case TestStarted:
		return "tciTestCaseStarted"
	case TestTerminated:
		return "tciTestCaseTerminated"
	case ControlStarted:
		return "tciControlStarted"
	case ControlTerminated:
		return "tciControlTerminated"
	default:
		return "unknown"
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (t EventType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (t EventType) UnmarshalText(text []byte) error {
	switch string(text) {
	case "tciError":
		t = Error
	case "tciTestCaseStarted":
		t = TestStarted
	case "tciTestCaseTerminated":
		t = TestTerminated
	case "tciControlStarted":
		t = ControlStarted
	case "tciControlTerminated":
		t = ControlTerminated
	default:
		return fmt.Errorf("unknown event type: %s", text)
	}
	return nil
}

type Event struct {
	Time    time.Time `json:"time,omitempty"`
	Type    EventType `json:"type,omitempty"`
	Name    string    `json:"name,omitempty"`
	Verdict string    `json:"verdict,omitempty"`
	Err     error     `json:"error,omitempty"`
}

func (e Event) String() string {
	s := e.Type.String()
	if e.Name != "" {
		s += " " + e.Name
	}
	if e.Verdict != "" {
		s += " " + e.Verdict
	}
	if e.Err != nil {
		s += " (" + e.Err.Error() + ")"
	}
	return s
}

func (e Event) IsStartEvent() bool {
	return e.Type == TestStarted || e.Type == ControlStarted
}

func (e Event) IsEndEvent() bool {
	return e.Type == TestTerminated || e.Type == ControlTerminated
}

func (e Event) IsError() bool {
	if e.IsStartEvent() || e.Verdict == "pass" && e.Type != Error {
		return false
	}
	return true
}
func NewErrorEvent(err error) Event {
	return Event{
		Time:    time.Now(),
		Type:    Error,
		Verdict: "error",
		Err:     err,
	}
}

func (t *Test) Run() <-chan Event {
	return t.RunWithContext(context.Background())
}

func (t *Test) RunWithContext(ctx context.Context) <-chan Event {

	events := make(chan Event)

	go func() {
		defer close(events)

		if !strings.Contains(t.Name, ".") {
			events <- NewErrorEvent(ErrNotQualified)
			return
		}

		sid, err := session.Get()
		if err != nil {
			events <- NewErrorEvent(err)
			return
		}
		defer session.Release(sid)

		if t.LogFile == "" {
			t.LogFile = fmt.Sprintf("%s.log", fs.Stem(t.T3XF))
		}
		cmd := proc.CommandContext(ctx, t.Runtime, t.T3XF, "-o", t.LogFile)
		if s := env.Getenv("K3RFLAGS"); s != "" {
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
			events <- NewErrorEvent(&RuntimeError{Err: err})
			return
		}
		log.Debugf("+ %s\n", cmd.String())
		err = cmd.Start()
		if err != nil {
			events <- NewErrorEvent(&RuntimeError{Err: err})
			return
		}

		// k3r does not have a ControlStarted event, so we'll fake it.
		if strings.HasSuffix(t.Name, ".control") {
			events <- Event{
				Type: ControlStarted,
				Name: t.Name,
				Time: time.Now(),
			}
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
				events <- Event{
					Type: TestStarted,
					Name: name,
					Time: time.Now(),
				}
			case "tciTestCaseTerminated":
				events <- Event{
					Type:    TestTerminated,
					Name:    name,
					Verdict: v[1],
					Time:    time.Now(),
				}
			case "tciControlTerminated":
				// k3r does not send a verdict. I just assume it's pass.
				events <- Event{Type: ControlTerminated, Name: t.Name, Verdict: "pass", Time: time.Now()}
			case "tciError":
				switch v[1] {
				case "E101:":
					events <- NewErrorEvent(ErrNoSuchModule)
				case "E102:":
					events <- NewErrorEvent(ErrNoSuchTest)
				case "E200:":
					events <- NewErrorEvent(ErrRuntimeNotReady)
				case "E201:":
					// ErrModuleNotReady happens, when tciRootModule was not called.
					// This is a spurious error, so we'll ignore it.
					//events <- NewErrorEvent(ErrModuleNotReady)
				case "E202:":
					events <- NewErrorEvent(ErrTestNotReady)
				case "E203:":
					events <- NewErrorEvent(ErrControlNotReady)
				case "E999:":
					events <- NewErrorEvent(fmt.Errorf("%w: %s", ErrNotImplemented, strings.Join(v[2:], " ")))
				default:
					events <- NewErrorEvent(fmt.Errorf("%w: %s", ErrUnknown, strings.Join(v[1:], " ")))
				}
			default:
				events <- NewErrorEvent(fmt.Errorf("%w: %s", ErrInvalidMessage, line))
			}
		}
		err = waitGracefully(t, cmd)
		if ctx.Err() == context.DeadlineExceeded {
			events <- NewErrorEvent(ErrTimeout)
		} else if err != nil {
			events <- NewErrorEvent(&RuntimeError{Err: err})
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
		Name:    name,
		T3XF:    t3xf,
		Env:     os.Environ(),
		Runtime: k3.Runtime(),
	}
}
