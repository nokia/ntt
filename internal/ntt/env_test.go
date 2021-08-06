package ntt_test

import (
	"os"
	"strings"
	"testing"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

// Verify unknown variable lead to an error
func TestSuiteUnknownVar(t *testing.T) {
	clearEnv()
	suite := &ntt.Suite{}
	s, err := suite.Getenv("NTT_FNORD")
	if _, ok := err.(*ntt.NoSuchVariableError); !ok {
		assert.Fail(t, "Expected NoSuchVariableError")
	}
	assert.Equal(t, "", s)
}

// Verify known variables do not lead to an error
func TestSuiteKnownVar(t *testing.T) {
	clearEnv()
	suite := &ntt.Suite{}
	s, err := suite.Getenv("NTT_NAME")
	assert.Nil(t, err)
	assert.Equal(t, "", s)
}

// Verify empty environment variables do not lead to an error.
func TestSuiteEmpty(t *testing.T) {
	clearEnv()
	suite := &ntt.Suite{}
	os.Setenv("K3_FNORD", "")
	s, err := suite.Getenv("NTT_FNORD")
	assert.Nil(t, err)
	assert.Equal(t, "", s)
}

type Vars map[string]string

// Verify simple access.
func TestVarsSimple(t *testing.T) {
	suite := suiteWithVars(Vars{
		"NTT_SIMPLE": "simple",
	})
	v, err := suite.Getenv("NTT_SIMPLE")
	assert.Nil(t, err)
	assert.Equal(t, "simple", v)
}

// Verify environment overwrites variables
func TestVarsEnvOverwrites(t *testing.T) {
	suite := suiteWithVars(Vars{
		"NTT_SIMPLE": "simple",
	})
	os.Setenv("NTT_SIMPLE", "fromEnv")
	v, err := suite.Getenv("NTT_SIMPLE")
	assert.Nil(t, err)
	assert.Equal(t, "fromEnv", v)
}

// Verify variables expand transitively (with ErrUnknownVariable)
func TestExpandTransitive(t *testing.T) {
	suite := suiteWithVars(Vars{
		"NTT_A": "$NTT_B",
		"NTT_B": "$NTT_C",
	})
	v, err := suite.Getenv("NTT_A")
	assert.NotNil(t, err)
	assert.Equal(t, "", v)
}

// Verify variables expand transitively (from environment)
func TestExpandTransitive2(t *testing.T) {
	suite := suiteWithVars(Vars{
		"NTT_A": "$NTT_B",
		"NTT_B": "$NTT_C",
	})
	os.Setenv("NTT_C", "fromEnv")
	v, err := suite.Getenv("NTT_A")
	assert.Nil(t, err)
	assert.Equal(t, "fromEnv", v)
}

// Verify environment variables do not expand.
func TestExpandEnv(t *testing.T) {
	suite := suiteWithVars(Vars{
		"NTT_A": "$NTT_B",
		"NTT_B": "fromVars",
	})
	os.Setenv("NTT_A", "$NTT_B")
	v, err := suite.Getenv("NTT_A")
	assert.Nil(t, err)
	assert.Equal(t, "$NTT_B", v)
}

// Verify recursion does not cause trouble
func TestExpandRecursive(t *testing.T) {
	suite := suiteWithVars(Vars{
		"NTT_A": "$NTT_A",
	})
	v, err := suite.Getenv("NTT_A")
	assert.Nil(t, err)
	assert.Equal(t, "", v)
}

// Verify indirect recursion does not cause trouble
func TestExpandRecursive2(t *testing.T) {
	suite := suiteWithVars(Vars{
		"NTT_A": "$NTT_B",
		"NTT_B": "$NTT_A",
	})
	v, err := suite.Getenv("NTT_A")
	assert.Nil(t, err)
	assert.Equal(t, "", v)
}

func suiteWithVars(vars map[string]string) *ntt.Suite {
	var manifest struct {
		Variables map[string]string
	}
	manifest.Variables = vars

	b, err := yaml.Marshal(manifest)
	if err != nil {
		panic(err)
	}
	clearEnv()
	fs.SetContent("package.yml", b)
	suite := &ntt.Suite{}
	suite.SetRoot(".")
	return suite
}

func clearEnv(files ...string) {
	if len(files) == 0 {
		files = []string{"ntt.env", "k3.env"}
	}
	for _, file := range files {
		fs.Open(file).SetBytes(nil)
	}

	for _, e := range os.Environ() {
		if fields := strings.Split(e, "="); len(fields) > 0 {
			key := fields[0]
			if strings.HasPrefix(key, "K3") || strings.HasPrefix(key, "NTT") {
				os.Unsetenv(key)
			}
		}
	}
}
