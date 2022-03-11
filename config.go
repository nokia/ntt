package ntt

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/k3"
)

// SplitQualifiedName splits a qualified name into module and test name.
func SplitQualifiedName(name string) (string, string) {
	parts := strings.Split(name, ".")
	if len(parts) == 1 {
		return "", name
	}
	return parts[0], strings.Join(parts[1:], ".")
}

// NewSuite creates a new suite from the given files. It expects either
// a single directory as argument or a list of regular .ttcn3 files.
//
// Calling NewSuite with an empty argument list will create a suite from
// current working directory or, if set, from NTT_SOURCE_DIR.
//
// NewSuite will read manifest (package.yml) if any.
func NewSuite(files ...string) (*Suite, error) {
	oldSuite, err := ntt.NewFromArgs(files...)
	if err != nil {
		return nil, fmt.Errorf("loading test suite failed: %w", err)
	}

	name, err := oldSuite.Name()
	if err != nil {
		return nil, fmt.Errorf("retrieving test suite name failed: %w", err)
	}

	srcs, err := oldSuite.Sources()
	if err != nil {
		return nil, fmt.Errorf("retrieving TTCN-3 sources failed: %w", err)
	}

	imports, err := oldSuite.Imports()
	if err != nil {
		return nil, fmt.Errorf("retrieving TTCN-3 imports failed: %w", err)
	}

	var paths []string
	if s := env.Getenv("NTT_CACHE"); s != "" {
		paths = append(paths, strings.Split(s, ":")...)
	}
	paths = append(imports, k3.FindAuxiliaryDirectories()...)
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, cwd)
	}

	if err != nil {
		return nil, fmt.Errorf("retrieving runtime paths failed: %w", err)
	}

	t, err := oldSuite.Timeout()
	if err != nil {
		return nil, fmt.Errorf("retrieving test suite timeout failed: %w\n", err)
	}

	d, err := time.ParseDuration(fmt.Sprintf("%fs", t))
	if err != nil {
		return nil, fmt.Errorf("retrieving test suite timeout failed: %w\n", err)
	}

	var parametersFile string
	f, err := oldSuite.ParametersFile()
	if err != nil {
		return nil, fmt.Errorf("retrieving parameters file failed: %w", err)
	}
	if f != nil {
		parametersFile = f.Path()
	}

	dir, err := oldSuite.ParametersDir()
	if err != nil {
		return nil, fmt.Errorf("retrieving parameters file failed: %w", err)
	}
	v, err := oldSuite.Variables()
	if err != nil {
		return nil, fmt.Errorf("retrieving test suite variables failed: %w", err)
	}
	return &Suite{
		Vars:           v,
		Name:           name,
		Sources:        srcs,
		RuntimePaths:   paths,
		timeout:        d,
		parametersFile: parametersFile,
		parametersDir:  dir,
	}, nil

}

// Suite represents a test suite.
type Suite struct {
	Name         string
	Sources      []string
	RuntimePaths []string
	Vars         map[string]string

	timeout        time.Duration
	parametersFile string
	parametersDir  string
}

// Parameters returns the module parameters and timeout to be used for execution.
func (s *Suite) Parameters() (map[string]string, time.Duration, error) {
	return s.parameters("")
}

// TestParameters returns module module parameters and timeout to be used for a single test case.
func (s *Suite) TestParameters(name string) (map[string]string, time.Duration, error) {
	return s.parameters(name)
}

func (s *Suite) parameters(name string) (map[string]string, time.Duration, error) {
	m := make(map[string]string)

	if path := s.parametersFile; path != "" {
		if err := readParametersFile(path, &m); err != nil {
			return nil, 0, fmt.Errorf("reading %q failed: %w", path, err)
		}
		log.Debugf("Read parameters from %q\n", path)
	}

	if mod, test := SplitQualifiedName(name); mod != "" {
		testPars := filepath.Join(s.parametersDir, mod, test+".parameters")
		if err := readParametersFile(testPars, &m); err != nil {
			return nil, 0, fmt.Errorf("reading %qw failed: %w", testPars, err)
		}
		log.Debugf("Read test specific parameters from %q\n", testPars)
	}

	d := s.timeout
	if t, ok := m["TestcaseExecutor.time_out"]; ok {
		delete(m, "TestcaseExecutor.time_out")
		d2, err := time.ParseDuration(fmt.Sprintf("%ss", t))
		if err != nil {
			return nil, 0, err
		}
		d = d2
	}
	return m, d, nil
}

func readParametersFile(path string, v interface{}) error {
	b, err := fs.Content(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("reading parameters file failed: %w", err)
	}
	_, err = toml.Decode(string(b), v)
	if err != nil {
		return fmt.Errorf("decoding parameters file failed: %w", err)
	}
	return nil
}
