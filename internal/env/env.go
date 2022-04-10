// Package env provides functionality to dynamically load the environment variables
//
// Note: This is just a modified copy and wrapper around the gotenv package by suboisto
package env

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/nokia/ntt/internal/cache"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
)

var ErrUnknownVariable = fmt.Errorf("unknown variable")
var ErrCyclicVariable = fmt.Errorf("cyclic variable")

var knownVars = map[string]bool{
	"PATH":                true,
	"CFLAGS":              true,
	"CPPFLAGS":            true,
	"CXXFLAGS":            true,
	"K3CFLAGS":            true,
	"K3RFLAGS":            true,
	"LDFLAGS":             true,
	"LD_LIBRARY_PATH":     true,
	"NTT_DATADIR":         true,
	"NTT_IMPORTS":         true,
	"NTT_NAME":            true,
	"NTT_PARAMETERS_DIR":  true,
	"NTT_PARAMETERS_FILE": true,
	"NTT_SOURCES":         true,
	"NTT_SOURCE_DIR":      true,
	"NTT_TEST_HOOK":       true,
	"NTT_TIMEOUT":         true,
	"NTT_VARIABLES":       true,
}

// Slice returns a sorted string slice of the variables.
func (env Env) Slice() []string {
	var s []string
	for k, env := range env {
		s = append(s, fmt.Sprintf("%s=%s", k, env))
	}
	sort.Strings(s)
	return s
}

// Expand variable references recursively. Environment variables overwrite
// variables defined in environment files. Undefined variables won't
// be expanded and will return an error.
func (env Env) Expand() error {
	var (
		err     error
		expand  func(string) string
		visited map[string]bool
	)

	expand = func(name string) string {
		if s, ok := LookupEnv(name); ok {
			return s
		}

		v, ok := env[name]
		if !ok {
			if knownVars[name] {
				return ""
			}
			if err == nil {
				err = fmt.Errorf("%w: %s", ErrUnknownVariable, name)
			}
			return fmt.Sprintf("${%s}", name)
		}
		if !visited[name] {
			visited[name] = true
			v = os.Expand(v, expand)
			if err == nil {
				env[name] = v
			}
			return v

		}
		if err == nil {
			err = fmt.Errorf("%w: %s", ErrCyclicVariable, name)
		}
		return fmt.Sprintf("${%s}", name)
	}

	for name := range env {
		visited = make(map[string]bool)
		v := expand(name)
		if err == nil {
			env[name] = v
		}
	}
	return err
}

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
