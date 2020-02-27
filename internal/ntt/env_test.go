package ntt_test

import (
	"os"
	"testing"

	"github.com/nokia/ntt/internal/ntt"
	"github.com/stretchr/testify/assert"
)

func TestEnvEmpty(t *testing.T) {
	defer os.Unsetenv("NTT_FNORD")

	suite := &ntt.Suite{}
	s := suite.Getenv("NTT_FNORD")
	assert.Equal(t, "", s)
}

func TestEnvSimple(t *testing.T) {
	defer os.Unsetenv("NTT_FNORD")
	os.Setenv("NTT_FNORD", "23.5")

	suite := &ntt.Suite{}
	s := suite.Getenv("NTT_FNORD")
	assert.Equal(t, "23.5", s)
}

func TestEnvK3(t *testing.T) {
	os.Unsetenv("NTTFNORD")

	defer os.Unsetenv("K3FNORD")
	os.Setenv("K3FNORD", "23.5")

	suite := &ntt.Suite{}
	s := suite.Getenv("NTTFNORD")
	assert.Equal(t, "23.5", s)
}
