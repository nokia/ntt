package run

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/nokia/ntt/internal/fs"
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

	// Timeout is the maximum run time for the test in seconds.
	Timeout time.Duration

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

	// Begin is the time the test started.
	Begin time.Time

	// End is the time the test ended.
	End time.Time

	// Verdict is the test verdict.
	Verdict string

	// Reason is the reason for the verdict.
	Reason string
}

func (t *Test) Run() error {
	ctx := context.Background()
	if t.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, t.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, t.Runtime, t.T3XF, "-o", t.LogFile)
	cmd.Dir = t.Dir
	cmd.Env = append(t.Env, "K3_SERVER=pipe,/dev/fd/0,/dev/fd/1")
	cmd.Env = append(cmd.Env, t.ModulePars...)
	cmd.Stdin = strings.NewReader(t.request())

	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	t.Begin = time.Now()
	err := cmd.Run()
	t.End = time.Now()

	t.handleResponse(&stdout)

	if ctx.Err() == context.DeadlineExceeded {
		t.Verdict = "error"
		t.Reason = "timeout"
		return ctx.Err()
	}

	return err
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
	return req.String()
}

// handleResponse parses the response from the K3 runtime, grepping the verdict
func (t *Test) handleResponse(resp io.Reader) {
	scanner := bufio.NewScanner(resp)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		switch v := strings.Split(line, "\t"); v[0] {
		case "tciTestCaseTerminated":
			t.Verdict = v[1]
		case "tciError":
			t.Verdict = "error"
			t.Reason = strings.Join(v[1:], "\t")
		case "tciControlTerminated":
		}
	}
}

// NewTest creates a new test instance ready to run.
func NewTest(t3xf string, name string) *Test {
	return &Test{
		Name:    name,
		T3XF:    t3xf,
		Env:     os.Environ(),
		LogFile: fs.ReplaceExt(t3xf, "log"),
		Runtime: k3.Runtime(),
		Verdict: "none",
	}
}
