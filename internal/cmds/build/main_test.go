package build

import (
	"errors"
	"os"
	"testing"
)

func TestImportLib(t *testing.T) {
	tests := []struct {
		path   string
		result []string
		err    error
	}{
		{path: "./testdata/invalid", err: os.ErrNotExist},
		{path: "./testdata/other", err: ErrNoSources},
		{path: "./testdata/lib", result: []string{
			"testdata/lib/a.ttcn3",
			"testdata/lib/b.ttcn3",
		}},
	}

	for _, tt := range tests {
		result, err := buildImport(tt.path)
		if !errors.Is(err, tt.err) {
			t.Errorf("%v: %v, want %v", tt.path, err, tt.err)
		}
		if !equal(result, tt.result) {
			t.Errorf("%v: %v, want %v", tt.path, result, tt.result)
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
