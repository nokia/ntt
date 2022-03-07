package build_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nokia/ntt/build"
	"github.com/stretchr/testify/assert"
)

func TestPathf(t *testing.T) {
	os.Setenv("NTT_CACHE", "..")
	defer os.Unsetenv("NTT_CACHE")

	assert.Equal(t, "../README.md", build.Pathf("README.%s", "md"))
	assert.Equal(t, "foobar.md", build.Pathf("foobar.md"))
	assert.Equal(t, "build/README.md", build.Pathf("build/README.md"))
}

func TestFieldsExpand(t *testing.T) {
	os.Setenv("TEST_FOO", "xxx yyy")
	os.Unsetenv("TEST_BAR")
	defer os.Unsetenv("TEST_FOO")

	assert.Equal(t, []string{"a", "xxx", "yyy", "b", "c"},
		build.FieldsExpand("a $TEST_FOO b $TEST_BAR c"))
}

func TestCommand(t *testing.T) {
	os.Setenv("CC", "ccache -x gcc")
	os.Unsetenv("CXX")
	defer os.Unsetenv("CC")

	actual := build.Command("$CC", "-c", "foo.c")
	assert.Equal(t, []string{"ccache", "-x", "gcc", "-c", "foo.c"}, actual.Args)
	assert.Equal(t, []string{""}, build.Command("$CXX").Args)
}

func TestNeedsRebuild(t *testing.T) {
	t.Parallel()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	for _, n := range []string{"a", "b", "c", "d"} {
		path := filepath.Join(dir, n)
		if err := ioutil.WriteFile(path, []byte{}, 0644); err != nil {
			t.Fatal(err)
		}
		time.Sleep(200 * time.Millisecond)
	}

	os.Setenv("TEST_B", "b")
	defer os.Unsetenv("TEST_B")

	a := filepath.Join(dir, "a")
	b := filepath.Join(dir, "$TEST_B")
	c := filepath.Join(dir, "c")
	d := filepath.Join(dir, "d")
	x := filepath.Join(dir, "x")

	tests := []struct {
		paths []string
		want  bool
		err   error
	}{
		{paths: []string{x}, want: true},
		{paths: []string{d}, want: false},
		{paths: []string{d, x}, want: false, err: os.ErrNotExist},
		{paths: []string{d, d}, want: false},
		{paths: []string{d, a, c, b}, want: false},
		{paths: []string{d, a, a, a}, want: false},
		{paths: []string{c, a, d, b}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.paths[0], func(t *testing.T) {
			ok, err := build.NeedsRebuild(tt.paths[0], tt.paths[1:]...)
			if !errors.Is(err, tt.err) {
				t.Errorf("NeedsRebuild(%v) error = %v, want %v", tt.paths, err, tt.err)
			}
			assert.Equal(t, tt.want, ok)
		})
	}

}
