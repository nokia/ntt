package project_test

import (
	"os"
	"strings"
	"testing"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/yaml"
	"github.com/nokia/ntt/project"
	"github.com/stretchr/testify/assert"
)

// Verify unknown variable lead to an error
func TestSuiteUnknownVar(t *testing.T) {
	clearEnv()
	p := &project.Project{}
	s, err := p.Getenv("NTT_FNORD")
	if _, ok := err.(*project.NoSuchVariableError); !ok {
		assert.Fail(t, "Expected NoSuchVariableError")
	}
	assert.Equal(t, "", s)
}

// Verify known variables do not lead to an error
func TestSuiteKnownVar(t *testing.T) {
	clearEnv()
	p := &project.Project{}
	s, err := p.Getenv("NTT_NAME")
	assert.Nil(t, err)
	assert.Equal(t, "", s)
}

// Verify empty environment variables do not lead to an error.
func TestSuiteEmpty(t *testing.T) {
	clearEnv()
	p := &project.Project{}
	os.Setenv("K3_FNORD", "")
	s, err := p.Getenv("NTT_FNORD")
	assert.Nil(t, err)
	assert.Equal(t, "", s)
}

type Vars map[string]string

// Verify simple access.
func TestVarsSimple(t *testing.T) {
	p := projectWithVars(Vars{
		"NTT_SIMPLE": "simple",
	})
	v, err := p.Getenv("NTT_SIMPLE")
	assert.Nil(t, err)
	assert.Equal(t, "simple", v)
}

// Verify environment overwrites variables
func TestVarsEnvOverwrites(t *testing.T) {
	p := projectWithVars(Vars{
		"NTT_SIMPLE": "simple",
	})
	os.Setenv("NTT_SIMPLE", "fromEnv")
	v, err := p.Getenv("NTT_SIMPLE")
	assert.Nil(t, err)
	assert.Equal(t, "fromEnv", v)
}

// Verify variables expand transitively (with ErrUnknownVariable)
func TestExpandTransitive(t *testing.T) {
	p := projectWithVars(Vars{
		"NTT_A": "$NTT_B",
		"NTT_B": "$NTT_C",
	})
	v, err := p.Getenv("NTT_A")
	assert.NotNil(t, err)
	assert.Equal(t, "", v)
}

// Verify variables expand transitively (from environment)
func TestExpandTransitive2(t *testing.T) {
	p := projectWithVars(Vars{
		"NTT_A": "$NTT_B",
		"NTT_B": "$NTT_C",
	})
	os.Setenv("NTT_C", "fromEnv")
	v, err := p.Getenv("NTT_A")
	assert.Nil(t, err)
	assert.Equal(t, "fromEnv", v)
}

// Verify environment variables do not expand.
func TestExpandEnv(t *testing.T) {
	p := projectWithVars(Vars{
		"NTT_A": "$NTT_B",
		"NTT_B": "fromVars",
	})
	os.Setenv("NTT_A", "$NTT_B")
	v, err := p.Getenv("NTT_A")
	assert.Nil(t, err)
	assert.Equal(t, "$NTT_B", v)
}

// Verify recursion does not cause trouble
func TestExpandRecursive(t *testing.T) {
	p := projectWithVars(Vars{
		"NTT_A": "$NTT_A",
	})
	v, err := p.Getenv("NTT_A")
	assert.Nil(t, err)
	assert.Equal(t, "", v)
}

// Verify indirect recursion does not cause trouble
func TestExpandRecursive2(t *testing.T) {
	p := projectWithVars(Vars{
		"NTT_A": "$NTT_B",
		"NTT_B": "$NTT_A",
	})
	v, err := p.Getenv("NTT_A")
	assert.Nil(t, err)
	assert.Equal(t, "", v)
}

// The purpose of this test is to clarify what happens, when undeclared variables are used inside manifest.
func TestExpandManifest(t *testing.T) {
	var manifest struct {
		Sources   []string
		Variables Vars
	}
	manifest.Variables = Vars{"NTT_A": "."}
	manifest.Sources = []string{
		"$NTT_A/project.go",
		"$NTT_X/project.go",
		"${NTT_X}./project.go",
		"${NTT_ENV}/project.go",
	}
	b, err := yaml.Marshal(manifest)
	if err != nil {
		panic(err)
	}
	clearEnv()
	os.Setenv("NTT_ENV", "fromEnv")
	fs.SetContent("package.yml", b)
	p, err := project.Open(".")
	if err != nil {
		panic(err)
	}
	srcs, err := p.Sources()
	assert.NotNil(t, err)
	assert.Equal(t, []string{"project.go", "$NTT_X/project.go", "${NTT_X}./project.go", "fromEnv/project.go"}, srcs)
}

func projectWithVars(vars map[string]string) *project.Project {
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
	p, err := project.Open(".")
	if err != nil {
		panic(err)
	}
	return p
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
