package ntt

import (
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml"
)

// Environ returns a copy of strings representing the environment, in the form "key=value".
func (s *Suite) Environ() []string {
	return nil
}

func (s *Suite) AddEnvFiles(files ...string) {
	for i := range files {
		s.envFiles = append(s.envFiles, s.File(files[i]))
	}
}

// Expand expands string v using Suite.Getenv
func (s *Suite) Expand(v string) (string, error) {
	return s.expand(v, make(map[string]bool))
}

// Expand expands string v using Suite.Getenv
func (s *Suite) expand(v string, visited map[string]bool) (string, error) {
	var gerr error
	mapper := func(name string) string {
		v, err := s.getenv(name, visited)
		if err != nil && gerr == nil {
			gerr = err
		}
		return v
	}

	return os.Expand(v, mapper), gerr
}

// Getenv retrieves the value of the environment variable named by the key. It
// returns the value, which will be empty if the variable is not present.
// Getenv will return an error object for critical errors, like IO or syntax
// errors in env files.
//
//
// Getenv looks for the key in the process environment first. If not found, it
// searches for key in all env files and finally it tries the variables-section
// in the manifest, if any.
//
// If key starts with "NTT" and could not be found, Getenv will replace the
// prefix with "K3" to help with migrating old scripts. For example, when
// looking for "NTT_FOO" following lookups will happen:
//
//         os.Getenv("NTT_FOO") os.Getenv("K3_FOO")
//         suite.lookupEnvFile("$PWD/ntt.env", "NTT_FOO")
//         suite.lookupEnvFile("$PWD/ntt.env", "K3_FOO")
//         suite.lookupEnvFile("$PWD/k3.env", "NTT_FOO")
//         suite.lookupEnvFile("$PWD/k3.env", "K3_FOO")
//         suite.lookupEnvFile("$HOME/.config/ntt/ntt.env", "NTT_FOO")
//         suite.lookupEnvFile("$HOME/.config/ntt/ntt.env", "K3_FOO")
//         suite.lookupEnvFile("$HOME/.config/ntt/k3.env", "NTT_FOO")
//         suite.lookupEnvFile("$HOME/.config/ntt/k3.env", "K3_FOO") ...
//
//
func (s *Suite) Getenv(key string) (string, error) {
	return s.getenv(key, make(map[string]bool))
}

func (s *Suite) getenv(key string, visited map[string]bool) (string, error) {
	if visited[key] {
		return "", fmt.Errorf("recursion detected when expanding variable %q.", key)
	}
	visited[key] = true

	if env := s.lookupProcessEnv(key); env != "" {
		return env, nil
	}

	for i := len(s.envFiles); i > 0; i-- {
		v, err := s.lookupEnvFile(s.envFiles[i-1], key)
		if err != nil {
			return "", err
		}
		if v != "" {
			return s.expand(v, visited)
		}
	}
	return "", fmt.Errorf("variable %q not found.", key)
}

// Lookup key in process environment
func (s *Suite) lookupProcessEnv(key string) string {
	if env := os.Getenv(key); env != "" {
		return env
	}

	if len(key) >= 3 && key[:3] == "NTT" {
		key = "K3" + strings.TrimPrefix(key, "NTT")
		return os.Getenv(key)
	}

	return ""
}

// Lookup key in environment file
func (s *Suite) lookupEnvFile(file *File, key string) (string, error) {
	b, err := file.Bytes()
	if err != nil {
		// It's okay if this file does not exist.
		if os.IsNotExist(err) {
			return "", nil
		}
	}

	tree, err := toml.LoadBytes(b)
	if tree == nil || err != nil {
		return "", err
	}

	if v := tree.Get(key); v != nil {
		return fmt.Sprint(v), nil
	}

	// Try K3 prefix.
	if len(key) >= 3 && key[:3] == "NTT" {
		key = "K3" + strings.TrimPrefix(key, "NTT")
		if v := tree.Get(key); v != nil {
			return fmt.Sprint(v), nil
		}
	}

	return "", nil
}
