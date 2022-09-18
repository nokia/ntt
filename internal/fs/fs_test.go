package fs_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/nokia/ntt/internal/fs"
	"github.com/stretchr/testify/assert"
)

// TestBytesFromURL verifies, that files specified by an URL can be read.
func TestBytesFromURL(t *testing.T) {
	path, err := filepath.Abs("fs_test.go")
	if err != nil {
		panic(err)
	}

	expected, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	f := fs.Open("file://" + path)
	b, err := f.Bytes()
	assert.Nil(t, err)
	assert.Equal(t, expected, b)
}

func TestCaching(t *testing.T) {
	assert.Equal(t, "package.yml", fs.Open("package.yml").Path())

	os.Setenv("NTT_CACHE", "testdata/cache")
	assert.Equal(t, "testdata/cache/package.yml", fs.Open("package.yml").Path())
}

func TestJoinPath(t *testing.T) {
	tests := []struct {
		first, second string
		want          string
	}{
		{"", "", ""},
		{".", "", "."},
		{".", "a", "a"},
		{"/", "b", "/b"},
		{"//", "c", "/c"},
		{"/", "/d", "/d"},
		{"e", "f", "e/f"},
		{"/g", "h", "/g/h"},
		{"/i", "../j", "/j"},
		{"file://k", "l", "file://k/l"},
		{"file:///m", "n", "file:///m/n"},
		{"file:///o", "../p", "file:///p"},
	}

	for _, test := range tests {
		got := fs.JoinPath(test.first, test.second)
		assert.Equal(t, test.want, got)
	}

}

func TestTTCN3Files(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got, err := fs.TTCN3Files()
		assert.Nil(t, err)
		assert.Nil(t, got)
	})
	t.Run("dir", func(t *testing.T) {
		got, err := fs.TTCN3Files("testdata/TestTTCN3Files")
		assert.Nil(t, err)
		assert.Nil(t, got)
	})
	t.Run("dir", func(t *testing.T) {
		got, err := fs.TTCN3Files("testdata/TestTTCN3Files/some-dir")
		assert.Nil(t, err)
		assert.Nil(t, got)
	})
	t.Run("dir", func(t *testing.T) {
		want := []string{
			"testdata/TestTTCN3Files/ttcn3-dir/a.ttcn3",
			"testdata/TestTTCN3Files/ttcn3-dir/b.ttcn",
			"testdata/TestTTCN3Files/ttcn3-dir/c.ttcnpp",
		}
		got, err := fs.TTCN3Files("testdata/TestTTCN3Files/ttcn3-dir")
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	})
	t.Run("errors", func(t *testing.T) {
		want := []string{
			"testdata/TestTTCN3Files/xxx-dir/a.ttcn3",
		}
		got, err := fs.TTCN3Files("testdata/TestTTCN3Files/xxx-dir/a.ttcn3")
		assert.True(t, errors.Is(err, os.ErrNotExist))
		assert.Equal(t, want, got)
	})
	t.Run("file", func(t *testing.T) {
		want := []string{
			"testdata/TestTTCN3Files/ttcn3-dir/a.ttcn3",
			"testdata/TestTTCN3Files/ttcn3-dir/a.ttcn3",
		}
		got, err := fs.TTCN3Files(
			"testdata/TestTTCN3Files/ttcn3-dir/a.ttcn3",
			"testdata/TestTTCN3Files/ttcn3-dir/a.ttcn3",
		)
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	})
	t.Run("URI", func(t *testing.T) {
		want := []string{"foo://a.ttcn3"}
		got, err := fs.TTCN3Files("foo://a.ttcn3")
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	})
	t.Run("URI", func(t *testing.T) {
		want := []string{"foo://a.ttcn3?bar"}
		got, err := fs.TTCN3Files("foo://a.ttcn3?bar")
		assert.True(t, errors.Is(err, fs.ErrInvalidFileExtension))
		assert.Equal(t, want, got)
	})

}
