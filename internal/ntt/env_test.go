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
	s, _ := suite.Getenv("NTT_FNORD")
	assert.Equal(t, "", s)
}

func TestEnvSimple(t *testing.T) {
	defer os.Unsetenv("NTT_FNORD")
	os.Setenv("NTT_FNORD", "23.5")

	suite := &ntt.Suite{}
	s, _ := suite.Getenv("NTT_FNORD")
	assert.Equal(t, "23.5", s)
}

func TestEnvK3(t *testing.T) {
	defer os.Unsetenv("K3FNORD")
	os.Setenv("K3FNORD", "23.5")

	suite := &ntt.Suite{}
	s, _ := suite.Getenv("NTTFNORD")
	assert.Equal(t, "23.5", s)
}

func TestEnvFile(t *testing.T) {
	suite := &ntt.Suite{}
	suite.AddEnvFiles("ntt.env", "k3.env")

	// Basic tests if prefix mapping works with environment files.

	// K3 prefix is _not_ replaced with NTT prefix.
	suite.File("ntt.env").SetBytes([]byte(`NTT_FNORD="var1"`))
	s, err := suite.Getenv("NTT_FNORD")
	assert.Nil(t, err)
	assert.Equal(t, "var1", s)
	s, _ = suite.Getenv("K3_FNORD")
	assert.Equal(t, "", s)

	// NTT prefix is replaced with K3 prefix.
	suite.File("ntt.env").SetBytes([]byte(`K3_FNORD="var2"`))
	s, _ = suite.Getenv("NTT_FNORD")
	assert.Equal(t, "var2", s)
	s, _ = suite.Getenv("K3_FNORD")
	assert.Equal(t, "var2", s)

	suite.File("ntt.env").SetBytes([]byte(`NTT_FNORD="var1"
	K3_FNORD="var2"`))
	s, _ = suite.Getenv("NTT_FNORD")
	assert.Equal(t, "var1", s)
	s, _ = suite.Getenv("K3_FNORD")
	assert.Equal(t, "var2", s)

	suite.File("k3.env").SetBytes([]byte(`NTT_FNORD="var3"`))
	s, _ = suite.Getenv("NTT_FNORD")
	assert.Equal(t, "var3", s)
	s, _ = suite.Getenv("K3_FNORD")
	assert.Equal(t, "var2", s)

	suite.File("k3.env").SetBytes([]byte(`K3_FNORD="var3"`))
	s, _ = suite.Getenv("NTT_FNORD")
	assert.Equal(t, "var3", s)
	s, _ = suite.Getenv("K3_FNORD")
	assert.Equal(t, "var3", s)

	// Test if os environment overwrites environment files.
	suite = &ntt.Suite{}
	suite.AddEnvFiles("ntt.env")
	suite.File("ntt.env").SetBytes([]byte(`NTT_FNORD="var1"`))
	os.Setenv("K3_FNORD", "var2")
	s, err = suite.Getenv("NTT_FNORD")
	os.Unsetenv("K3_FNORD")
	assert.Nil(t, err)
	assert.Equal(t, "var2", s)

	// Test if types are converted to strings nicely.
	suite = &ntt.Suite{}
	suite.AddEnvFiles("ntt.env")
	suite.File("ntt.env").SetBytes([]byte(`NTT_FLOAT=23.5`))
	s, err = suite.Getenv("NTT_FLOAT")
	assert.Nil(t, err)
	assert.Equal(t, "23.5", s)
	// TODO(5nord) Also test collections

	// Various expansion tests.
	suite = &ntt.Suite{}
	suite.AddEnvFiles("ntt.env")
	suite.File("ntt.env").SetBytes([]byte(`
		# Undefined reference is okay.
		NTT_A="${NTT_UNDEFINED}"

		# Keys may be defined later.
		NTT_B="${NTT_C}"
		NTT_C=23.5

		# Direct recursion should not break ntt.
		NTT_D="${NTT_D}"

		# Indirect recursion shouldn't either.
		NTT_E="${NTT_F}"
		NTT_F="${NTT_E}"
	`))

	s, err = suite.Getenv("NTT_A")
	assert.NotNil(t, err)
	assert.Equal(t, "", s)

	s, err = suite.Getenv("NTT_B")
	assert.Nil(t, err)
	assert.Equal(t, "23.5", s)

	s, err = suite.Getenv("NTT_D")
	assert.NotNil(t, err)
	assert.Equal(t, "", s)
}
