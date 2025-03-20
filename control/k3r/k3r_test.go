package k3r

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tsts "github.com/nokia/ntt/control"
	"github.com/nokia/ntt/project"
	"github.com/stretchr/testify/assert"
)

func TestEvents(t *testing.T) {
	old, _ := initStage(t)

	conf, err := project.Open(filepath.Join(old, "testdata/suite"))
	if err != nil {
		t.Fatal(err)
	}
	if k3r := conf.K3.Runtime; k3r == "k3r" || k3r == "" {
		t.Skip("no k3 runtime found")
	}

	if err := project.Build(conf); err != nil {
		t.Fatal(err)
	}

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
				`StopEvent test.control done`,
			}},
		{
			input: "test2.control",
			events: []string{
				`StartEvent test2.control`,
				`StartEvent test2.A`,
				`StopEvent test2.A pass`,
				`StopEvent test2.control done`,
			}},
		{
			input:  "test3.control",
			events: nil, // no error message when control does not exist
		},
		{
			input: "test3.X",
			events: []string{
				"ErrorEvent error (no such test case)",
			}},

		{
			input: "X.X",
			events: []string{
				"ErrorEvent error (no such module)",
				"ErrorEvent error (no such test case)",
			}},
		{
			input: "test3.test3",
			events: []string{
				"ErrorEvent error (no such test case)",
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
			tst := NewTest(conf.K3.T3XF, tsts.NewJob(tt.input, conf))
			for e := range tst.Run(ctx) {
				if _, ok := e.(tsts.LogEvent); ok {
					continue
				}

				s := strings.TrimPrefix(fmt.Sprintf("%T", e), "control.")
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
		want := []string{
			"K3_NAME=",
			"K3R_PATH=.",
			"LD_LIBRARY_PATH=.",
			"PATH=.",
			"K3_SERVER=pipe,/dev/fd/0,/dev/fd/1",
		}
		reset := clearEnv()
		defer reset()
		test := newTest()
		got, err := buildEnv(test)
		assert.Nil(t, err)
		assert.Equal(t, want, got, "without K3_SERVER k3r wont communicate")

	})

	t.Run("library paths", func(t *testing.T) {
		want := []string{
			"K3_NAME=",
			"K3R_PATH=.:import1:import2:k3-plugins",
			"LD_LIBRARY_PATH=.:import1:import2",
			"PATH=.:import1:import2",
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
			"K3_NAME=",
			"TEST_VAR_A=vartest",
			"K3R_PATH=.:import1:import2",
			"LD_LIBRARY_PATH=.:import1:import2",
			"PATH=.:import1:import2",
			"K3_SERVER=pipe,/dev/fd/0,/dev/fd/1",
		}
		reset := clearEnv()
		defer reset()
		test := newTest()
		test.Config.Variables = map[string]string{
			"TEST_VAR_A": "vartest",
		}
		test.Config.Imports = []string{"import1", "import2"}
		got, err := buildEnv(test)
		assert.Nil(t, err)
		assert.Equal(t, want, got, "library path need to be exported to k3r")
	})

	t.Run("test env", func(t *testing.T) {
		want := []string{
			"K3_NAME=",
			"FOO=fromVar",
			"FOO=fromEnv",
			"K3R_PATH=.",
			"LD_LIBRARY_PATH=.",
			"PATH=.",
			"K3_SERVER=pipe,/dev/fd/0,/dev/fd/1"}
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
	t.Run("test paths", func(t *testing.T) {
		want := []string{
			"K3_NAME=",
			"K3R_PATH=k3r_path:.:import1:plugins",
			"LD_LIBRARY_PATH=ld_library_path:.:import1:clib",
			"PATH=path:.:import1:libexec",
			"K3_SERVER=pipe,/dev/fd/0,/dev/fd/1"}
		reset := clearEnv()
		defer reset()
		test := newTest()
		test.K3.Runtime = "libexec/k3r"
		test.K3.CLibDirs = []string{"clib"}
		test.K3.Plugins = []string{"plugins"}
		test.Imports = []string{"import1"}
		os.Setenv("K3R_PATH", "k3r_path")
		os.Setenv("LD_LIBRARY_PATH", "ld_library_path")
		os.Setenv("PATH", "path")
		got, err := buildEnv(test)
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	})
}

func newTest() *Test {
	return &Test{
		Job: &tsts.Job{
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
