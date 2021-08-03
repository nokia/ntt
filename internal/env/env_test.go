package env_test

import (
	"os"
	"testing"

	"github.com/nokia/ntt/internal/env"
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
	env.Load()
	if env := env.Getenv("NTT_FOO"); env != "" {
		t.Errorf("Environment variable NTT_FOO should be empty")
	}

	os.Setenv("NTT_CACHE", "testdata/cache")
	env.Load()
	if env := env.Getenv("NTT_FOO"); env != "bar" {
		t.Errorf("Want: NTT_FOO=%q, Got: NTT_FOO=%q", "bar", env)
	}

}

func TestFromCache(t *testing.T) {
	os.Setenv("NTT_CACHE", "testdata/cache")
	assert.Equal(t, "./file", env.FromCache("./file"))
	assert.Equal(t, "./env.go", env.FromCache("./env.go"))

	assert.Equal(t, "file", env.FromCache("file"))
	assert.Equal(t, "file://env.go", env.FromCache("file://env.go"))
	assert.Equal(t, ".", env.FromCache("."))
	assert.Equal(t, "..", env.FromCache(".."))
	assert.Equal(t, "env.go", env.FromCache("env.go"))
	assert.Equal(t, "testdata/cache/other.go", env.FromCache("other.go"))
}
