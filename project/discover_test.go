package project_test

import (
	"testing"

	"github.com/nokia/ntt/project"
	"github.com/stretchr/testify/assert"
)

func TestDiscoverManifests(t *testing.T) {
	actual := project.Discover("testdata/manifests/suite1/testcases")
	expected := []string{"testdata/manifests/suite1", "testdata/manifests/some-dir"}
	assert.Equal(t, expected, actual)
}
