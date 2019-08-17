/*
Package env is a poor man's Go configuration library and provides simplistic
access to environment variables and configuration files.

This package provides map-like access to environment variables, shell-like
variable expansion and loading from TOML configuration files. Example:


	// Set prefix for environment variables and config files.
	env.SetPrefix("foo")

	// Add search directories. Values in $HOME override values in /etc
	env.AddPath("$HOME")
	env.AddPath("/etc")

	// Slurp in
	env.ReadEnvFiles()

	os.Setenv("FOO_VAR1", "original")
	env.Set("Var1", "value1")
	env.Set("vaR2", "value2")

	fmt.Println(env.Get("var1")) // Prints "original"
	fmt.Println(env.Get("VAR2")) // Prints "value2"

	// Note, all names are case-insensitive.

This packages was written to smooth migration from Nokia's internal tools to
ntt. It is not intended to be used by third-parties and may be removed in the
future.
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

const DefaultPrefix = "ntt"

var DefaultEnv = New(DefaultPrefix)

type Env struct {
	prefix string
	dirs   []string
	used   []string
	data   map[string]string
}

// New returns a new and empty instance of Env
func New(prefix string) *Env {
	DefaultEnv := &Env{
		prefix: prefix,
		data:   make(map[string]string, 0),
	}
	return DefaultEnv
}

// SetPrefix sets a new prefix. Note, previouly synced variables are not
// removed.
func (env *Env) SetPrefix(prefix string) {
	env.prefix = prefix
}

// Reset clears all variables. Note, environment variables won't be unset.
func (env *Env) Reset() {
	env.data = make(map[string]string, 0)
	env.used = nil
}

// Set adds a value with key.
func (env *Env) Set(key string, value string) {
	env.data[strings.ToLower(key)] = value
}

// Get returns the value of a given key; or "" if key does not exist.
// Variable references (env.g. ${SOMETHING}) are expanded, if that reference
// exists.
func (env *Env) Get(key string) string {
	return env.Expand(env.get(key))
}

func (env *Env) get(key string) string {
	if v := os.Getenv(env.varToEnv(key)); v != "" {
		return v
	}

	if v, ok := env.data[strings.ToLower(key)]; ok {
		return v
	}

	return ""
}

// Expand expands s using first getenv and then Env.
func (env *Env) Expand(s string) string {
	mapper := func(name string) string {
		// Always expand regular environment variables (PATH, CFLAGS, ...)
		if v := os.Getenv(name); v != "" {
			return v
		}

		// Try expanding unexported variables
		if v := env.get(env.envToVar(name)); v != "" {
			return v
		}

		// Don't expand
		return fmt.Sprintf("${%s}", name)
	}

	return os.Expand(s, mapper)
}

// Keys returns a sorted list of keys, including those from environment.
func (env *Env) Keys() []string {
	env.Sync()
	keys := make([]string, 0, len(env.data))
	for k, _ := range env.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Sync copies all environment variables starting with prefix into Env.
func (env *Env) Sync() {
	for _, v := range os.Environ() {
		if env.hasPrefix(v) {
			f := strings.Split(v, "=")
			if v := env.envToVar(f[0]); v != "" {
				env.Set(v, strings.Join(f[1:], "="))
			}
		}
	}
}

// ToMap exports all data from Env as string-map, including environment.
// variables starting with prefix.
func (env *Env) ToMap() map[string]string {
	m := make(map[string]string, len(env.data))
	for _, k := range env.Keys() {
		m[k] = env.Get(k)
	}
	return m
}

// AddConfigPath adds a path to search for configuration files in.
func (env *Env) AddPath(path string) {
	env.dirs = append(env.dirs, path)
}

// Will read a configuration file from a io.Reader adding existing variables.
func (env *Env) ReadEnv(r io.Reader) error {
	tree, err := toml.LoadReader(r)
	if err != nil {
		return err
	}

	for k, v := range tree.ToMap() {
		os.Setenv(k, env.Expand(fmt.Sprint(v)))
	}

	return nil
}

// ReadEnvFiles discovers and reads environment variables
func (env *Env) ReadEnvFiles() error {
	for _, d := range env.dirs {
		path := filepath.Join(env.Expand(d), env.prefix+".env")
		if file, err := os.Open(path); err == nil {
			if err := env.ReadEnv(file); err != nil {
				return err
			}
			env.used = append(env.used, path)
		}
	}
	return nil
}

// EnvFileUsed returns a slice with files used to fill Env
func (env *Env) EnvFilesUsed() []string {
	return env.used
}

// varToEnv converts a variable name into a environment variable name.
// Example: "foo" becomes "K3_FOO" .
func (env *Env) varToEnv(s string) string {
	return strings.ToUpper(env.prefix + "_" + s)
}

// envToVar converts a environment variable name into variable name.
// Example: "K3_FOO" becomes "foo"
func (env *Env) envToVar(s string) string {
	return strings.TrimPrefix(strings.ToLower(s), strings.ToLower(env.prefix+"_"))
}

// hasPrefix returns true is string s begins with upper-case prefix and
// underscore.
func (env *Env) hasPrefix(s string) bool {
	return strings.HasPrefix(s, strings.ToUpper(env.prefix+"_"))
}

// SetPrefix sets a new prefix. Note, previouly synced variables are not
// removed.
func SetPrefix(prefix string) { DefaultEnv.SetPrefix(prefix) }

// Reset clears all variables. Note, environment variables won't be unset.
func Reset() { DefaultEnv.Reset() }

// Set adds a value with key.
func Set(key string, value string) { DefaultEnv.Set(key, value) }

// Get returns the value of a given key; or "" if key does not exist.
// Variable references (DefaultEnv.g. ${SOMETHING}) are expanded, if that reference
// exists.
func Get(key string) string { return DefaultEnv.Get(key) }

// Expand expands s using first getenv and then Env.
func Expand(s string) string { return DefaultEnv.Expand(s) }

// Keys returns a sorted list of keys, including those from environment.
func Keys() []string { return DefaultEnv.Keys() }

// Sync copies all environment variables starting with prefix into Env.
func Sync() { DefaultEnv.Sync() }

// ToMap exports all data from Env as string-map, including environment.
// variables starting with prefix.
func ToMap() map[string]string { return DefaultEnv.ToMap() }

// AddConfigPath adds a path to search for configuration files in.
func AddPath(path string) { DefaultEnv.AddPath(path) }

// Will read a configuration file from a io.Reader adding existing variables.
func ReadEnv(r io.Reader) error { return DefaultEnv.ReadEnv(r) }

// ReadEnvFiles discovers and reads environment variables
func ReadEnvFiles() error { return DefaultEnv.ReadEnvFiles() }

// EnvFileUsed returns a slice with files used to fill Env
func EnvFilesUsed() []string { return DefaultEnv.EnvFilesUsed() }
