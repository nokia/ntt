package ntt

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

// Sources returns the list of sources required to compile a Suite.
// The error will be != nil if input sources could not be determined correctly. For
// example, when `package.yml` had syntax errors.
func (s *Suite) Sources() ([]*File, error) {
	var ret []*File

	// Environment variable overwrite everything.
	if env := getenv("sources"); env != "" {
		for _, x := range strings.Fields(env) {
			ret = append(ret, s.File(x))
		}
		return ret, nil
	}

	// If there's a parseable package.yml, try that one.
	m, err := s.parseManifest()
	if err != nil {
		return nil, err
	}
	if m != nil && len(m.Sources) > 0 && s.root != nil {
		for i := range m.Sources {
			// Substitute environment variables
			src := expand(m.Sources[i])

			// Make paths which are relative to manifest, relative to CWD.
			if !filepath.IsAbs(src) {
				src = filepath.Clean(filepath.Join(s.root.Path(), src))
			}

			// Directories need expansion into single files.
			info, err := os.Stat(src)
			if err != nil {
				return nil, err
			}
			switch {
			case info.IsDir():
				files, err := TTCN3Files(src)
				if err != nil {
					return nil, err
				}
				if len(files) == 0 {
					return nil, fmt.Errorf("Could not find ttcn3 source files in directory %q", src)
				}
				for i := range files {
					ret = append(ret, s.File(files[i]))
				}

			case info.Mode().IsRegular() && hasTTCN3Extension(src):
				ret = append(ret, s.File(src))

			default:
				return nil, fmt.Errorf("Cannot handle %q. Expecting directory or ttcn3 source file", src)
			}

		}
		return append(ret, s.sources...), nil
	}

	// If there's only a root folder, look for .ttcn3 files
	if s.root != nil {
		files, err := TTCN3Files(s.root.Path())
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			ret = append(ret, s.File(f))
		}
		return append(ret, s.sources...), nil

	}

	// Last resort is sources list, explicitly curated by AddSources-calls.
	return s.sources, nil
}

func (s *Suite) AddSources(files ...string) {
	for i := range files {
		s.sources = append(s.sources, s.File(files[i]))
	}
}

type manifest struct {
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
