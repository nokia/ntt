package env

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/nokia/ntt/internal/log"
	"github.com/subosito/gotenv"
)

var Files = []string{"ntt.env", "k3.env"}

type Env gotenv.Env

// Load environment files ntt.env and k3.env
func Load(files ...string) {
	for k, v := range Parse(files...) {
		if _, ok := os.LookupEnv(k); !ok {
			os.Setenv(k, v)
		}
	}
}

// Parse environment files ntt.env and k3.env
func Parse(files ...string) Env {
	if len(files) == 0 {
		files = Files
	}

	var env Env
	for _, path := range Files {
		f, err := os.Open(FromCache(path))
		if err != nil {
			log.Debugf("open env: %s\n", err.Error())
		}
		e, err := gotenv.StrictParse(f)
		if err != nil {
			log.Verbosef("error parsing %q: %s\n", path, err.Error())
		}
		for k, v := range e {
			if _, ok := env[k]; !ok {
				if env == nil {
					env = make(map[string]string)
				}
				env[k] = v
			}
		}
	}
	return env
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
