package ntt_test

import (
	"errors"
	"os"
	"syscall"
	"testing"

	"github.com/nokia/ntt"
	"github.com/nokia/ntt/k3"
)

func TestPlanImports(t *testing.T) {
	os.Setenv("NTT_WANT_HELPER_PROCESS", "1")
	defer os.Unsetenv("NTT_WANT_HELPER_PROCESS")

	tests := []struct {
		path   string
		result []string
		err    error
	}{
		{path: "./testdata/invalid/notexist", err: os.ErrNotExist},
		{path: "./testdata/invalid/file.ttcn3", err: syscall.ENOTDIR},
		{path: "./testdata/invalid/dirs", err: ntt.ErrNoSources},
		{path: "./testdata/other", err: ntt.ErrNoSources},
		{path: "./testdata/🤔", result: []string{"testdata/🤔/a.ttcn3"}},
		{path: "./testdata/lib", result: []string{"testdata/lib/a.ttcn3", "testdata/lib/b.ttcn3", "testdata/lib/🤔.ttcn3"}},
	}

	for _, tt := range tests {
		result, err := ntt.PlanImport(tt.path)
		if !errors.Is(err, tt.err) {
			t.Errorf("%v: %v, want %v", tt.path, err, tt.err)
		}
		if len(tt.result) == 0 {
			continue
		}
		if len(result) != 1 {
			t.Errorf("Unexpected result: %v", result)
			continue
		}

		b, ok := result[0].(*k3.T3XF)
		if !ok {
			t.Errorf("Expected T3XF builder, got %T", result[0])
			continue
		}
		actual := b.Sources()
		if !equal(actual, tt.result) {
			t.Errorf("%v: %v, want %v", tt.path, actual, tt.result)
		}
	}

}

// equal returns true if a and b are equal string slices, order is ignored.
func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]int, len(a))
	for i := range a {
		m[a[i]]++
		m[b[i]]--
	}
	for _, v := range m {
		if v != 0 {
			return false
		}
	}
	return true
}
