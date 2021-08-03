package env

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/subosito/gotenv"
)

// Load environment files ntt.env and k3.env
func Load() {
	gotenv.Load(FromCache("ntt.env"))
	gotenv.Load(FromCache("k3.env"))
}

// Lookup key in process environment. Is key begins with "NTT_" also lookup key
// with "K3_" prefix.
func Getenv(key string) string {
	if env, ok := LookupEnv(key); ok {
		return env
	}
	return ""
}

// LookupEnv is like Getenv, but returns a true boolean when key exists
func LookupEnv(key string) (string, bool) {
	if env, ok := os.LookupEnv(key); ok {
		return env, ok
	}
	if strings.HasPrefix(key, "NTT") {
		return os.LookupEnv(strings.Replace(key, "NTT", "K3", 1))
	}
	return "", false
}

// FromCache works similar to GNU Makes VPATH functionality: Paths without directory portion will be looked up alternate directory specified by NTT_CACHE environment variable.
func FromCache(path string) string {
	// Skip URLs
	if u, _ := url.Parse(path); u.Scheme != "" {
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

	if cache := Getenv("NTT_CACHE"); cache != "" {
		for _, dir := range strings.Split(cache, ":") {
			path := filepath.Join(dir, path)
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	return path
}
