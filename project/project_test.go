package project_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nokia/ntt/project"
	"github.com/stretchr/testify/assert"
)

func TestProjects(t *testing.T) {
	// Regular test suite with package.yml
	assert.Equal(t, []string{"testdata/project/a.ttcn3", "testdata/project/c.ttcnpp"}, files(open("testdata/project/suite1")))

	// Regular test suite with variables
	assert.Equal(t, []string{"testdata/project/a.ttcn3"}, files(open("testdata/project/suite2")))

	// Regular test suite with variables
	assert.Equal(t, []string{"testdata/project/suite3/foobar/a.ttcn3"}, files(open("testdata/project/suite3")))

	// Absolute from environment
	abs, _ := filepath.Abs("project_test.go")
	os.Setenv("NTT_SOURCES", abs)
	assert.Equal(t, []string{abs}, files(open("testdata/project/suite1")))

	// Path with dollar sign from environment
	os.Setenv("NTT_SOURCES", "$PATH")
	assert.Equal(t, []string{"$PATH"}, files(open("testdata/project/suite1")))
	os.Unsetenv("NTT_SOURCES")

	// Invalid package.yml
	assert.Nil(t, files(open("testdata/project/suite4")))
}

func open(root string) project.Interface {
	p, _ := project.Open(root)
	return p
}

func files(p project.Interface) []string {
	files, _ := project.Files(p)
	return files
}
