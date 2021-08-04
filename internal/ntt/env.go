package ntt

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-multierror"
	"github.com/nokia/ntt/internal/env"
)

type NoSuchVariableError struct {
	Name string
}

func (e *NoSuchVariableError) Error() string {
	return e.Name + ": variable not defined"
}

var knownVars = map[string]bool{
	"CXXFLAGS":            true,
	"K3CFLAGS":            true,
	"K3RFLAGS":            true,
	"LDFLAGS":             true,
	"LD_LIBRARY_PATH":     true,
	"PATH":                true,
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

// Environ returns a copy of strings representing the environment, in the form "key=value".
func (suite *Suite) Environ() ([]string, error) {
	var errs error

	allKeys := make(map[string]struct{})

	vars, err := suite.Variables()
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	for k := range vars {
		allKeys[k] = struct{}{}
	}

	for k := range env.Parse() {
		allKeys[k] = struct{}{}
	}

	ret := make([]string, 0, len(allKeys))
	for k := range allKeys {
		v, err := suite.Getenv(k)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
		ret = append(ret, fmt.Sprintf("%s=%s", k, v))
	}
	return ret, nil
}

// Expand expands string v using Suite.Getenv
func (suite *Suite) Expand(v string) (string, error) {
	return suite.expand(v, make(map[string]string))
}

func (suite *Suite) Getenv(v string) (string, error) {
	return suite.getenv(v, make(map[string]string))
}

func (suite *Suite) expand(v string, visited map[string]string) (string, error) {
	var errs error
	mapper := func(name string) string {
		v, err := suite.getenv(name, visited)
		if err != nil {
			errs = multierror.Append(errs, &NoSuchVariableError{Name: name})
		}
		return v
	}
	return os.Expand(v, mapper), errs
}

func (suite *Suite) getenv(key string, visited map[string]string) (string, error) {
	if v, ok := visited[key]; ok {
		return v, nil
	}
	visited[key] = ""

	if v, ok := env.LookupEnv(key); ok {
		visited[key] = v
		return v, nil
	}
	vars, err := suite.Variables()
	if err != nil {
		return "", err
	}

	// We must not look for NTT_CACHE in variables sections of package.yml,
	// because this would create an endless loop.
	if key != "NTT_CACHE" && key != "K3_CACHE" {
		if v, ok := vars[key]; ok {
			v, err := suite.expand(v, visited)
			visited[key] = v
			return v, err
		}
	}

	if knownVars[key] {
		return "", nil
	}

	return "", &NoSuchVariableError{Name: key}
}
