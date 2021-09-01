package project_test

import (
	"testing"

	"github.com/nokia/ntt/project"
	"github.com/stretchr/testify/assert"
)

func TestDiscoverManifests(t *testing.T) {
	actual := project.Discover("testdata/manifests/suite1/testcases")
	expected := []string{"testdata/manifests/suite1", "testdata/manifests/obsolete", "testdata/manifests/some-dir"}
	assert.Equal(t, expected, actual)

}

func TestDiscoverBuildDirectories(t *testing.T) {
	actual := project.Discover("testdata/builds/suite1/testcases")
	expected := []string{"testdata/builds/suite1/build", "testdata/builds/build/obsolete", "testdata/builds/build/some-dir", "testdata/builds/suite2"}
	assert.Equal(t, expected, actual)
}
