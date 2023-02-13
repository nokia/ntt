package k3_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nokia/ntt/internal/proc"
	"github.com/nokia/ntt/k3"
)

func TestNewPlugin(t *testing.T) {
	if k3.Runtime() == "k3r" {
		t.Skip("Cannot locate k3 runtime. Skipping test.")
	}

	srcdir, _ := initStage(t)

	b := k3.NewPlugin(k3.DefaultEnv, "extfunc", filepath.Join(srcdir, "testdata/suite/extfunc/plugin.cc"))[0]
	err := b.Run()
	if err != nil {
		t.Fatalf("Run() = %v", err)
	}

	out := b.Outputs()
	if len(out) != 1 {
		t.Fatalf("Unexpected numberof build artifacts: %v", out)
	}

	abs, err := filepath.Abs(out[0])
	if err != nil {
		t.Fatal(err)
	}

	cmd := proc.Command(k3.Runtime(), "--plugin", abs)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Plugin did not load correcty: %s", err.Error())
	}
	expected := fmt.Sprintf("|%s|k3r-extfunc-plugin|extfunc", abs)
	if actual := strings.Split(string(stdout), "\n"); !sliceContains(actual, expected) {
		t.Errorf("Plugin did not load correctly:\n%s\n", string(stdout))
	}
}

func TestNewT3XF(t *testing.T) {
	if k3.Compiler() == "mtc" {
		t.Skip("Cannot locate k3 compiler. Skipping test.")
	}
	srcdir, _ := initStage(t)

	b := k3.NewT3XF(k3.DefaultEnv, "suite.t3xf", []string{filepath.Join(srcdir, "testdata/suite/test.ttcn3")})[0]
	err := b.Run()
	if err != nil {
		t.Errorf("Run() = %v", err)
	}

	out := b.Outputs()
	if len(out) != 1 {
		t.Fatalf("Unexpected numberof build artifacts: %v", out)
	}

	_, err = os.Stat(out[0])
	if err != nil {
		t.Error(err)
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

func sliceContains(s []string, x string) bool {
	for _, s := range s {
		if strings.Contains(s, x) {
			return true
		}
	}
	return false
}
