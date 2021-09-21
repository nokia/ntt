package env

import (
	"bytes"
	"os"
	"strings"

	"github.com/moosq/gotenv"
	"github.com/nokia/ntt/internal/cache"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
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
		f := fs.Open(cache.Lookup(path))
		b, err := f.Bytes()
		if err != nil {
			log.Debugf("open env: %s\n", err.Error())
			continue
		}
		e, err := gotenv.StrictParse(bytes.NewReader(b))
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
