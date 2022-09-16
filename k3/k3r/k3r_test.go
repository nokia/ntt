package k3r

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
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/tests"
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
				"ErrorEvent error (not a test case)",
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
			tst := NewTest(t3xf, tt.input)
			tst.Env = append(tst.Env, env.Environ()...)
			for e := range tst.Run(ctx) {
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

func TestBuildEnv(t *testing.T) {
	clearEnv := func() func() {
		a, okA := os.LookupEnv("PATH")
		b, okB := os.LookupEnv("K3R_PATH")
		c, okC := os.LookupEnv("LD_LIBRARY_PATH")
		os.Unsetenv("PATH")
		os.Unsetenv("K3R_PATH")
		os.Unsetenv("LD_LIBRARY_PATH")
		return func() {
			if okA {
				os.Setenv("PATH", a)
			}
			if okB {
				os.Setenv("K3R_PATH", b)
			}
			if okC {
				os.Setenv("LD_LIBRARY_PATH", c)
			}
		}

	}

	t.Run("empty", func(t *testing.T) {
		want := []string{"K3_SERVER=pipe,/dev/fd/0,/dev/fd/1"}
		reset := clearEnv()
		defer reset()
		test := newTest()
		got, err := buildEnv(test)
		assert.Nil(t, err)
		assert.Equal(t, want, got, "without K3_SERVER k3r wont communicate")

	})

	t.Run("library paths", func(t *testing.T) {
		want := []string{
			"K3R_PATH=import1:import2:k3-plugins",
			"LD_LIBRARY_PATH=import1:import2:k3-plugins",
			"PATH=import1:import2:k3-plugins",
			"K3_SERVER=pipe,/dev/fd/0,/dev/fd/1",
		}
		reset := clearEnv()
		defer reset()
		test := newTest()
		test.Config.Imports = []string{"import1", "import2"}
		test.Config.K3.Plugins = []string{"k3-plugins"}
		got, err := buildEnv(test)
		assert.Nil(t, err)
		assert.Equal(t, want, got, "library path need to be exported to k3r")
	})

	t.Run("variables", func(t *testing.T) {
		want := []string{
			"K3R_PATH=import1:import2",
			"LD_LIBRARY_PATH=import1:import2",
			"PATH=import1:import2:path",
			"TEST_VAR_A=vartest",
			"K3_SERVER=pipe,/dev/fd/0,/dev/fd/1",
		}
		reset := clearEnv()
		defer reset()
		os.Setenv("PATH", "path")
		test := newTest()
		test.Config.Variables = map[string]string{
			"PATH":       "varpath",
			"TEST_VAR_A": "vartest",
		}
		test.Config.Imports = []string{"import1", "import2"}
		got, err := buildEnv(test)
		assert.Nil(t, err)
		assert.Equal(t, want, got, "library path need to be exported to k3r")
	})

	t.Run("test env", func(t *testing.T) {
		want := []string{"FOO=fromVar", "FOO=fromEnv", "K3_SERVER=pipe,/dev/fd/0,/dev/fd/1"}
		reset := clearEnv()
		defer reset()
		test := newTest()
		test.Config.Variables = map[string]string{
			"FOO": "fromVar",
		}
		test.Env = []string{"FOO=fromEnv"}

		got, err := buildEnv(test)
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	})
}

func newTest() *Test {
	return &Test{
		Job: &tests.Job{
			Config: &project.Config{},
		},
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
