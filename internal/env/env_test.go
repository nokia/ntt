package env_test

import (
	"os"
	"strings"
	"testing"

	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/fs"
	"github.com/stretchr/testify/assert"
)

func init() {
	os.Unsetenv("NTT_CACHE")
	os.Unsetenv("K3_CACHE")
	os.Unsetenv("NTT_FOO")
	os.Unsetenv("K3_FOO")
}

func TestCache(t *testing.T) {

	// First load should not find anything, because CACHE is not set.
	env.LoadFiles()
	if env := env.Getenv("NTT_FOO"); env != "" {
		t.Errorf("Environment variable NTT_FOO should be empty")
	}

	os.Setenv("NTT_CACHE", "testdata/cache")
	env.LoadFiles()
	if env := env.Getenv("NTT_FOO"); env != "bar" {
		t.Errorf("Want: NTT_FOO=%q, Got: NTT_FOO=%q", "bar", env)
	}

}

func TestEnvEmpty(t *testing.T) {
	clearEnv()
	assert.Equal(t, "", env.Getenv("NTT_FNORD"))
}

func TestEnvSimple(t *testing.T) {
	clearEnv()
	os.Setenv("NTT_FNORD", "23.5")
	assert.Equal(t, "23.5", env.Getenv("NTT_FNORD"))
}

func TestEnvK3(t *testing.T) {
	clearEnv()
	os.Setenv("K3FNORD", "23.5")
	assert.Equal(t, "23.5", env.Getenv("NTTFNORD"))
}

// Test that NTT prefix is replaced with K3 prefix.
//
// When, for example, asking for NTT_CACHE, we should also look for any
// K3_CACHE variables of legacy systems.
func TestPrefixReplaceNTT(t *testing.T) {
	clearEnv()
	setContent("ntt.env", `K3_FNORD="var2"`)
	env.LoadFiles("ntt.env")
	assert.Equal(t, "var2", env.Getenv("NTT_FNORD"))
	assert.Equal(t, "var2", env.Getenv("K3_FNORD"))
}

// Verify that K3 prefix is _not_ replaced with NTT prefix.
func TestPrefixKeepK3(t *testing.T) {
	clearEnv()
	setContent("ntt.env", `NTT_FNORD="var1"`)
	env.LoadFiles("ntt.env")
	assert.Equal(t, "var1", env.Getenv("NTT_FNORD"))
	assert.Equal(t, "", env.Getenv("K3_FNORD"))
}

// Verify that prefix substituion does not overwrite existing variables
func TestPrefixHasBoth(t *testing.T) {
	clearEnv()
	setContent("ntt.env", `NTT_FNORD="var1"
	K3_FNORD="var2"`)
	env.LoadFiles()
	assert.Equal(t, "var1", env.Getenv("NTT_FNORD"))
	assert.Equal(t, "var2", env.Getenv("K3_FNORD"))
}

// Verify that ntt.env is loaded before k3.env
func TestEnvK3BeforeNTT(t *testing.T) {
	clearEnv()
	setContent("ntt.env", `NTT_FNORD="fromNTT"`)
	setContent("k3.env", `NTT_FNORD="fromK3"
	K3_FNORD="fromK3"`)
	env.LoadFiles()
	assert.Equal(t, "fromNTT", env.Getenv("NTT_FNORD"))
	assert.Equal(t, "fromK3", env.Getenv("K3_FNORD"))
}

// Verify that ntt.env is loaded before k3.env, like before, but with difference prefixes
func TestEnvK3BeforeNTTWithSubstitution(t *testing.T) {
	clearEnv()
	setContent("ntt.env", `K3_FNORD="fromNTT"`)
	setContent("k3.env", `NTT_FNORD="fromK3"
	K3_FNORD="fromK3"`)
	env.LoadFiles()
	assert.Equal(t, "fromK3", env.Getenv("NTT_FNORD"))
	assert.Equal(t, "fromNTT", env.Getenv("K3_FNORD"))
}

// Test if os environment overwrites environment files.
//
// This behaviour has been removed, because variables from environment files
// have been promoted to real os environment variables
//
//func TestPrecedence(t *testing.T) {
//	clearEnv()
//	setContent("ntt.env", `NTT_FNORD="fromNTT"`)
//	env.Load()
//	os.Setenv("K3_FNORD", "fromEnv")
//	assert.Equal(t, "fromEnv", env.Getenv("NTT_FNORD"))
//}

// Test if types are converted to strings nicely.
func TestEnvConversion(t *testing.T) {
	clearEnv()
	setContent("ntt.env", `NTT_FLOAT=23.5`)
	env.LoadFiles()
	assert.Equal(t, "23.5", env.Getenv("NTT_FLOAT"))
}

// Ensure that variables are substituted
func TestEnvExpansion(t *testing.T) {
	clearEnv()
	setContent("ntt.env", `
		NTT_A=a
		NTT_B=$NTT_A
	`)
	env.LoadFiles()
	assert.Equal(t, "a", env.Getenv("NTT_B"))
}

func TestEnvExpansion2(t *testing.T) {
	clearEnv()
	setContent("ntt.env", `
		NTT_C=23.5
		NTT_B=${NTT_C} $NTT_C
	`)
	env.LoadFiles()
	assert.Equal(t, "23.5 23.5", env.Getenv("NTT_B"))
}

// Environment files are evaluted line by line. Line order matters.
func TestEnvExpansionEnv(t *testing.T) {
	clearEnv()
	os.Setenv("NTT_A", "fromEnv")
	setContent("ntt.env", `
		NTT_B=$NTT_A
		NTT_A=a
	`)
	env.LoadFiles()
	assert.Equal(t, "fromEnv", env.Getenv("NTT_B"))
}

// Environment files are evaluted line by line. Line order matters.
func TestEnvExpansionEnv2(t *testing.T) {
	clearEnv()
	os.Setenv("NTT_A", "fromEnv")
	setContent("ntt.env", `
		NTT_A=a
		NTT_B=$NTT_A
	`)
	env.LoadFiles()
	assert.Equal(t, "fromEnv", env.Getenv("NTT_B"))
}

// Unknown variables are substituted with empty string.
func TestEnvExpansionUnknown(t *testing.T) {
	clearEnv()
	setContent("ntt.env", `
		NTT_B=$NTT_A
		NTT_A=a
	`)
	env.LoadFiles()
	assert.Equal(t, "", env.Getenv("NTT_B"))
}

func setContent(file string, content string) {
	fs.SetContent(file, []byte(content))
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
