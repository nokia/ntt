package cache

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/nokia/ntt/internal/lsp/span"
)

// Lookup works similar to GNU Makes VPATH functionality: Paths without directory portion will be looked up alternate directory specified by NTT_CACHE environment variable.
func Lookup(path string) string {
	// Skip URLs
	path = string(span.URINormalizeAuthority(path))
	if u, _ := url.Parse(path); u != nil && u.Scheme != "" {
		return path
	}

	// Skip existing paths
	if _, err := os.Stat(path); err == nil {
		return path
	}

	// Skip paths with directory portion
	if dir, _ := filepath.Split(path); dir != "" {
		return path
	}

	if cache := fromEnv(); cache != "" {
		for _, dir := range strings.Split(cache, ":") {
			path := filepath.Join(dir, path)
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	return path
}

func fromEnv() string {
	if cache := os.Getenv("NTT_CACHE"); cache != "" {
		return cache
	}
	return os.Getenv("K3_CACHE")
}
