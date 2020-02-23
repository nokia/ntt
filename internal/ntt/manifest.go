package ntt

import (
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v2"
)

// Timeout returns the configured timeout of the Suite. If there's no timeout,
// the function will return 0.
//
// The error will be != nil if timeout could not be determined correctly. For
// example, when `package.yml` had syntax errors.
func (s *Suite) Timeout() (float64, error) {
	if s := getenv("timeout"); s != "" {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, err
		}
		return f, nil
	}
	m, err := s.parseManifest()
	if err != nil {
		return 0, err
	}
	if m != nil {
		return m.Timeout, nil
	}
	return 0, nil
}

type manifest struct {
	root string

	// Static configuration
	Name    string
	Sources []string
	Imports []string

	// Runtime configuration
	TestHook       string  `yaml:test_hook`       // Path for test hook.
	ParametersFile string  `yaml:parameters_file` // Path for module parameters file.
	Timeout        float64 `yaml:timeout`         // Global timeout for tests.
}

// parseManifest tries to parse an (optional) manifest file.
func (s *Suite) parseManifest() (*manifest, error) {
	// Without root folder, there's no manifest to parse. This is ok.
	if s.root == nil {
		return nil, nil
	}

	f := s.File(filepath.Join(s.root.Path(), "package.yml"))
	b, err := f.Bytes()
	if err != nil {
		// If there's not package.yml, it's okay, too.
		if os.IsNotExist(err) {
			return nil, nil
		}
		// all other errors should be reported.
		return nil, err
	}

	var m manifest
	return &m, yaml.UnmarshalStrict(b, &m)
}
