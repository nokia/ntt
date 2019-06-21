/*
Package env provides high-level access to environment variables.
  * Expansion
  * Config files
*/
package env

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	toml "github.com/pelletier/go-toml"
)

var e = New("ntt")

type Env struct {
	prefix string
	dirs   []string
	used   []string
	data   map[string]string
}

// New returns a new and empty instance of Env
func New(prefix string) *Env {
	e := &Env{
		prefix: prefix,
		data:   make(map[string]string, 0),
	}
	return e
}

// SetPrefix sets a new prefix. Note, previouly synced variables are not
// removed.
func (e *Env) SetPrefix(prefix string) {
	e.prefix = prefix
}

// Reset clears all variables. Note, environment variables won't be unset.
func (e *Env) Reset() {
	e.data = make(map[string]string, 0)
	e.used = nil
}

// Set adds a value with key.
func (e *Env) Set(key string, value string) {
	e.data[strings.ToLower(key)] = value
}

// Get returns the value of a given key; or "" if key does not exist.
// Variable references (e.g. ${SOMETHING}) are expanded, if that reference
// exists.
func (e *Env) Get(key string) string {
	return e.expand(e.get(key))
}

func (e *Env) get(key string) string {
	if v := os.Getenv(e.varToEnv(key)); v != "" {
		return v
	}

	if v, ok := e.data[strings.ToLower(key)]; ok {
		return v
	}

	return ""
}

func (e *Env) expand(s string) string {
	mapper := func(name string) string {
		// Always expand regular environment variables (PATH, CFLAGS, ...)
		if v := os.Getenv(name); v != "" {
			return v
		}

		// Try expanding unexported variables
		if v := e.get(e.envToVar(name)); v != "" {
			return v
		}

		// Don't expand
		return fmt.Sprintf("${%s}", name)
	}

	return os.Expand(s, mapper)
}

// Keys returns a sorted list of keys, including those from environment.
func (e *Env) Keys() []string {
	e.Sync()
	keys := make([]string, 0, len(e.data))
	for k, _ := range e.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Sync copies all environment variables starting with prefix into Env.
func (e *Env) Sync() {
	for _, v := range os.Environ() {
		if e.hasPrefix(v) {
			f := strings.Split(v, "=")
			if v := e.envToVar(f[0]); v != "" {
				e.Set(v, strings.Join(f[1:], "="))
			}
		}
	}
}

// ToMap exports all data from Env as string-map, including environment.
// variables starting with prefix.
func (e *Env) ToMap() map[string]string {
	m := make(map[string]string, len(e.data))
	for _, k := range e.Keys() {
		m[k] = e.Get(k)
	}
	return m
}

// AddConfigPath adds a path to search for configuration files in.
func (e *Env) AddPath(path string) {
	e.dirs = append(e.dirs, path)
}

// Will read a configuration file from a io.Reader adding existing variables.
func (e *Env) ReadEnv(r io.Reader) error {
	tree, err := toml.LoadReader(r)
	if err != nil {
		return err
	}

	for k, v := range tree.ToMap() {
		e.Set(k, fmt.Sprint(v))
	}

	return nil
}

// ReadEnvFiles discovers and reads environment variables
func (e *Env) ReadEnvFiles() error {
	for _, d := range e.dirs {
		path := filepath.Join(d, e.prefix+".env")
		if file, err := os.Open(path); err == nil {
			if err := e.ReadEnv(file); err != nil {
				return err
			}
			e.used = append(e.used, path)
		}
	}
	return nil
}

// EnvFileUsed returns a slice with files used to fill Env
func (e *Env) EnvFilesUsed() []string {
	return e.used
}

// varToEnv converts a variable name into a environment variable name.
// Example: "foo" becomes "K3_FOO" .
func (e *Env) varToEnv(s string) string {
	return strings.ToUpper(e.prefix + "_" + s)
}

// envToVar converts a environment variable name into variable name.
// Example: "K3_FOO" becomes "foo"
func (e *Env) envToVar(s string) string {
	return strings.TrimPrefix(strings.ToLower(s), strings.ToLower(e.prefix+"_"))
}

// hasPrefix returns true is string s begins with upper-case prefix and
// underscore.
func (e *Env) hasPrefix(s string) bool {
	return strings.HasPrefix(s, strings.ToUpper(e.prefix+"_"))
}

// SetPrefix sets a new prefix. Note, previouly synced variables are not
// removed.
func SetPrefix(prefix string) { e.SetPrefix(prefix) }

// Reset clears all variables. Note, environment variables won't be unset.
func Reset() { e.Reset() }

// Set adds a value with key.
func Set(key string, value string) { e.Set(key, value) }

// Get returns the value of a given key; or "" if key does not exist.
// Variable references (e.g. ${SOMETHING}) are expanded, if that reference
// exists.
func Get(key string) string { return e.Get(key) }

// Keys returns a sorted list of keys, including those from environment.
func Keys() []string { return e.Keys() }

// Sync copies all environment variables starting with prefix into Env.
func Sync() { e.Sync() }

// ToMap exports all data from Env as string-map, including environment.
// variables starting with prefix.
func ToMap() map[string]string { return e.ToMap() }

// AddConfigPath adds a path to search for configuration files in.
func AddPath(path string) { e.AddPath(path) }

// Will read a configuration file from a io.Reader adding existing variables.
func ReadEnv(r io.Reader) error { return e.ReadEnv(r) }

// ReadEnvFiles discovers and reads environment variables
func ReadEnvFiles() error { return e.ReadEnvFiles() }

// EnvFileUsed returns a slice with files used to fill Env
func EnvFilesUsed() []string { return e.EnvFilesUsed() }
