package cache_test

import (
	"os"
	"testing"

	"github.com/nokia/ntt/internal/cache"
	"github.com/stretchr/testify/assert"
)

func init() {
	os.Unsetenv("NTT_CACHE")
	os.Unsetenv("K3_CACHE")
}

func TestLookup(t *testing.T) {
	os.Setenv("NTT_CACHE", "testdata/cache")
	assert.Equal(t, "./file", cache.Lookup("./file"))
	assert.Equal(t, "./cache.go", cache.Lookup("./cache.go"))

	assert.Equal(t, "file", cache.Lookup("file"))
	assert.Equal(t, "file://cache.go", cache.Lookup("file://cache.go"))
	assert.Equal(t, ".", cache.Lookup("."))
	assert.Equal(t, "..", cache.Lookup(".."))
	assert.Equal(t, "cache.go", cache.Lookup("cache.go"))
	assert.Equal(t, "testdata/cache/other.go", cache.Lookup("other.go"))
}
