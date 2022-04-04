package fs_test

import (
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
		got, err := fs.JoinPath(test.first, test.second)
		if err != nil {
			t.Errorf("JoinPath(%q, %q) error: %v", test.first, test.second, err)
		}
		assert.Equal(t, test.want, got)
	}

}
