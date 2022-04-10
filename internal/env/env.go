// Package env provides functionality to dynamically load the environment variables
//
// Note: This is just a modified copy and wrapper around the gotenv package by suboisto
package env

import (
	"bytes"
	"os"
	"strings"

	"github.com/nokia/ntt/internal/cache"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
)

var Files = []string{"ntt.env", "k3.env"}

// LoadFiles environment files ntt.env and k3.env
func LoadFiles(files ...string) {
	for k, v := range ParseFiles(files...) {
		if _, ok := os.LookupEnv(k); !ok {
			os.Setenv(k, v)
		}
	}
}

// ParseFiles environment files ntt.env and k3.env
func ParseFiles(files ...string) Env {
	if len(files) == 0 {
		files = Files
	}

	var env Env
	for _, path := range Files {
		f := fs.Open(cache.Lookup(path))
		b, err := f.Bytes()
		if err != nil {
			continue
		}
		log.Debugf("found environment file: %s\n", f.Path())
		e, err := StrictParse(bytes.NewReader(b))
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
