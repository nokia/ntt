package run

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/proc"
	"github.com/nokia/ntt/k3"
)

// A Test is a test case instance.
type Test struct {

	// Name is the fully qualified name of the test.
	Name string

	// Args is the list of arguments to pass to the test.
	Args []string

	// ModulePars hold module parameters
	ModulePars []string

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

	ErrNoSuch        = fmt.Errorf("no such thing")
	ErrNoSuchModule  = fmt.Errorf("no such module")
	ErrNoSuchTest    = fmt.Errorf("no such test case")
	ErrNoSuchControl = fmt.Errorf("no such control")

	ErrRuntimeNotReady = fmt.Errorf("runtime not ready")
	ErrModuleNotReady  = fmt.Errorf("module not ready")
	ErrTestNotReady    = fmt.Errorf("test case not ready")
	ErrControlNotReady = fmt.Errorf("control not ready")
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

		if t.LogFile == "" {
			t.LogFile = fmt.Sprintf("%s.log", fs.Stem(t.T3XF))
		}
		cmd := proc.CommandContext(ctx, t.Runtime, t.T3XF, "-o", t.LogFile)
		cmd.Dir = t.Dir
		cmd.Env = append(t.Env, "K3_SERVER=pipe,/dev/fd/0,/dev/fd/1")
		cmd.Env = append(cmd.Env, t.ModulePars...)
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

		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanLines)
		var name string
		for scanner.Scan() {
			line := scanner.Text()
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
				events <- Event{Type: ControlTerminated, Time: time.Now()}
			case "tciError":
				switch v[1] {
				case "E101:":
					events <- NewErrorEvent(ErrNoSuchModule)
				case "E102:":
					events <- NewErrorEvent(ErrNoSuchTest)
				case "E200:":
					events <- NewErrorEvent(ErrRuntimeNotReady)
				case "E201:":
					events <- NewErrorEvent(ErrModuleNotReady)
				case "E202:":
					events <- NewErrorEvent(ErrTestNotReady)
				case "E203:":
					events <- NewErrorEvent(ErrControlNotReady)
				case "E999:":
					events <- NewErrorEvent(fmt.Errorf("%w: %s", ErrNotImplemented, strings.Join(v[2:], " ")))
				default:
					events <- NewErrorEvent(fmt.Errorf("%w: %s", ErrUnknown, strings.Join(v[1:], " ")))
				}

			}
		}
		err = cmd.Wait()
		if ctx.Err() == context.DeadlineExceeded {
			events <- NewErrorEvent(ErrTimeout)
		} else if err != nil {
			events <- NewErrorEvent(&RuntimeError{Err: err})
		}
	}()

	return events
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
	fmt.Fprintln(&req, "tcinonExitTE")
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
