package ntt

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml"
	errors "golang.org/x/xerrors"
)

var ErrNoSuchVariable = errors.New("no such variable")

var knownVars = map[string]bool{
	"CFLAGS":              true,
	"CXXFLAGS":            true,
	"K3CFLAGS":            true,
	"K3RFLAGS":            true,
	"LDFLAGS":             true,
	"LD_LIBRARY_PATH":     true,
	"PATH":                true,
	"NTT_DATADIR":         true,
	"NTT_IMPORTS":         true,
	"NTT_NAME":            true,
	"NTT_PARAMETERS_FILE": true,
	"NTT_SOURCES":         true,
	"NTT_SOURCE_DIR":      true,
	"NTT_TEST_HOOK":       true,
	"NTT_TIMEOUT":         true,
	"NTT_VARIABLES":       true,
}

// Environ returns a copy of strings representing the environment, in the form "key=value".
func (suite *Suite) Environ() ([]string, error) {
	allKeys := make(map[string]struct{})

	if vars, _ := suite.Variables(); vars != nil {
		for k := range vars {
			allKeys[k] = struct{}{}
		}
	}
	for i := range suite.envFiles {
		tree, err := suite.parseEnvFile(suite.envFiles[i])
		if err != nil {
			return nil, err
		}
		if tree == nil {
			continue
		}
		for _, k := range tree.Keys() {
			allKeys[k] = struct{}{}
		}
	}

	ret := make([]string, 0, len(allKeys))
	for k := range allKeys {
		v, err := suite.Getenv(k)
		if err != nil {
			return nil, err
		}
		ret = append(ret, fmt.Sprintf("%s=\"%s\"", k, v))
	}
	return ret, nil
}

func (suite *Suite) AddEnvFiles(files ...string) {
	for i := range files {
		suite.envFiles = append(suite.envFiles, suite.File(files[i]))
	}
}

// Expand expands string v using Suite.Getenv
func (suite *Suite) Expand(v string) (string, error) {
	return suite.expand(v, make(map[string]string))
}

// Expand expands string v using Suite.Getenv
func (suite *Suite) expand(v string, visited map[string]string) (string, error) {
	var gerr error
	mapper := func(name string) string {
		v, err := suite.getenv(name, visited)
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
//
//
func (suite *Suite) Getenv(key string) (string, error) {
	return suite.getenv(key, make(map[string]string))
}

func (suite *Suite) getenv(key string, visited map[string]string) (string, error) {
	if s, ok := visited[key]; ok {
		return s, nil
	}

	if env, ok := suite.lookupProcessEnv(key); ok {
		visited[key] = env
		return env, nil
	}

	visited[key] = ""

	for i := len(suite.envFiles); i > 0; i-- {
		v, err := suite.lookupEnvFile(suite.envFiles[i-1], key)
		if err == nil {
			s, err := suite.expand(v, visited)
			visited[key] = s
			return s, err
		} else if err != ErrNoSuchVariable {
			return "", err
		}
	}

	// We must not look for NTT_CACHE in variables sections of package.yml,
	// because this would create an endless loop.
	if key != "NTT_CACHE" && key != "K3_CACHE" {
		v, err := suite.Variables()
		if err != nil {
			return "", err
		}
		if v != nil {
			if s, ok := v[key]; ok {
				s, err := suite.expand(s, visited)
				visited[key] = s
				return s, err
			}
		}
	}

	if suite.isKnown(key) {
		return "", nil
	}

	return "", fmt.Errorf("variable %q not found.", key)
}

// Lookup key in process environment
func (suite *Suite) lookupProcessEnv(key string) (string, bool) {
	if env, ok := os.LookupEnv(key); ok {
		return env, true
	}

	if len(key) >= 3 && key[:3] == "NTT" {
		key = "K3" + strings.TrimPrefix(key, "NTT")
		return os.LookupEnv(key)
	}

	return "", false
}

// Lookup key in environment file
func (suite *Suite) lookupEnvFile(file *File, key string) (string, error) {
	tree, err := suite.parseEnvFile(file)
	if err != nil {
		return "", err
	}

	if tree == nil {
		return "", ErrNoSuchVariable
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

	return "", ErrNoSuchVariable
}

func (suite *Suite) parseEnvFile(f *File) (*toml.Tree, error) {

	type envData struct {
		tree *toml.Tree
		err  error
	}

	f.handle = suite.store.Bind(f.ID(), func(ctx context.Context) interface{} {
		data := envData{}

		b, err := f.Bytes()
		if err != nil {
			// It's okay if this file does not exist.
			if os.IsNotExist(err) {
				data.tree, data.err = nil, nil
				return &data
			}
		}

		data.tree, data.err = toml.LoadBytes(b)
		return &data
	})

	v := f.handle.Get(context.TODO())
	data := v.(*envData)

	return data.tree, data.err
}

func (suite *Suite) isKnown(key string) bool {
	if knownVars[key] {
		return true
	}

	if len(key) >= 3 && key[:3] == "NTT" {
		key = "K3" + strings.TrimPrefix(key, "NTT")
	}

	return knownVars[key]
}
