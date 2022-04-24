package k3_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/k3/t3xf"
)

func TestNewPlugin(t *testing.T) {
	if k3.Runtime() == "k3r" {
		t.Skip("Cannot locate k3 runtime. Skipping test.")
	}

	srcdir, _ := initStage(t)

	b := k3.NewPlugin(k3.DefaultEnv, "extfunc", filepath.Join(srcdir, "testdata/suite/extfunc/plugin.cc"))[0]
	err := b.Run()
	if err != nil {
		t.Errorf("Run() = %v", err)
	}

	out := b.Outputs()
	if len(out) != 1 {
		t.Fatalf("Unexpected numberof build artifacts: %v", out)
	}

	abs, err := filepath.Abs(out[0])
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(k3.Runtime(), "--plugin", abs)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Plugin did not load correcty: %s", err.Error())
	}
	expected := fmt.Sprintf("|plod|k3r|%s|k3r-extfunc-plugin|extfunc", abs)
	if actual := strings.Split(string(stdout), "\n"); !sliceContains(actual, expected) {
		t.Errorf("Plugin did not load correctly:\n%s\n", string(stdout))
	}

}

func TestNewT3XF(t *testing.T) {
	if k3.Compiler() == "mtc" {
		t.Skip("Cannot locate k3 compiler. Skipping test.")
	}
	srcdir, _ := initStage(t)

	b := k3.NewT3XF(k3.DefaultEnv, "suite.t3xf", filepath.Join(srcdir, "testdata/suite/test.ttcn3"))[0]
	err := b.Run()
	if err != nil {
		t.Errorf("Run() = %v", err)
	}

	out := b.Outputs()
	if len(out) != 1 {
		t.Fatalf("Unexpected numberof build artifacts: %v", out)
	}

	file, err := t3xf.ReadFile(out[0])
	if err != nil {
		t.Fatalf("t3xf.ReadFile() = %v", err)
	}
	s := t3xf.NewScanner(file.Sections.T3XF)
	for s.Scan() {
	}
	if err := s.Err(); err != nil {
		t.Errorf("t3xf.Scan() = %v", err)
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
