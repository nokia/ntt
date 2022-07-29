package k3r_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/k3/k3r"
	"github.com/nokia/ntt/project"
	tsts "github.com/nokia/ntt/tests"
	"github.com/stretchr/testify/assert"
)

func TestEvents(t *testing.T) {
	old, _ := initStage(t)

	t3xf := testBuild(t, filepath.Join(old, "testdata/suite"))

	tests := []struct {
		input   string
		timeout time.Duration
		events  []string
	}{
		{
			input: "test.A",
			events: []string{
				`StartEvent test.A`,
				`StopEvent test.A pass`,
			}},
		{
			input: "test.control",
			events: []string{
				`StartEvent test.control`,
				`StartEvent test.B`,
				`StopEvent test.B fail`,
				`StartEvent test.A`,
				`StopEvent test.A pass`,
				`StopEvent test.control pass`,
			}},
		{
			input: "test2.control",
			events: []string{
				`StartEvent test2.control`,
				`StartEvent test2.A`,
				`StopEvent test2.A pass`,
				`StopEvent test2.control pass`,
			}},
		{
			input: "test3.control",
			events: []string{
				`StartEvent test3.control`,
				`StopEvent test3.control pass`, // no error message when control does not exist
			}},
		{
			input: "test3.X",
			events: []string{
				"ErrorEvent error (no such test case)",
			}},
		{
			input: "X.X",
			events: []string{
				"ErrorEvent error (no such module)",
			}},
		{
			input: "test3.test3",
			events: []string{
				"ErrorEvent error (exit status 2)", // Exit 2, due to exception.
			}},
		{
			input: "asd",
			events: []string{
				"ErrorEvent error (id not fully qualified)",
			}},
		{
			input: "test3",
			events: []string{
				"ErrorEvent error (id not fully qualified)",
			}},
		{
			input: "control",
			events: []string{
				"ErrorEvent error (id not fully qualified)",
			}},
		{
			input:   "test.D",
			timeout: 1 * time.Second,
			events: []string{
				`StartEvent test.D`,
				`ErrorEvent error (timeout)`,
			}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			ctx := context.Background()
			var cancel context.CancelFunc
			if tt.timeout > 0 {
				ctx, cancel = context.WithTimeout(ctx, time.Duration(tt.timeout))
			}

			var actual []string
			tst := k3r.NewTest(t3xf, tt.input)
			tst.Env = append(tst.Env, env.Environ()...)
			for e := range tst.RunWithContext(ctx) {
				s := strings.TrimPrefix(fmt.Sprintf("%T", e), "tests.")
				switch e := e.(type) {
				case tsts.StartEvent:
					s += fmt.Sprintf(" %s", e.Name)
				case tsts.StopEvent:
					s += fmt.Sprintf(" %s %s", e.Name, e.Verdict)
				case tsts.ErrorEvent:
					s += fmt.Sprintf(" error (%s)", e.Error())
				}
				actual = append(actual, s)
			}
			assert.Equal(t, tt.events, actual)
			if cancel != nil {
				cancel()
			}
		})
	}

}

func initStage(t *testing.T) (string, string) {
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	old, err = filepath.Rel(dir, old)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Chdir(old)
	})

	return old, dir
}

func testBuild(t *testing.T, args ...string) string {
	t.Helper()
	if k3r := k3.Runtime(); k3r == "k3r" {
		t.Skip("no k3 runtime found")
	}

	p, err := project.Open(args...)
	if err != nil {
		t.Fatal(err)
	}
	if err := project.Build(p); err != nil {
		t.Fatal(err)
	}
	return p.K3.T3XF
}
