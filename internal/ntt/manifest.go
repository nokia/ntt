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
	if s := s.Getenv("NTT_TIMEOUT"); s != "" {
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
	if env := s.Getenv("NTT_SOURCES"); env != "" {
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
			src := s.expand(m.Sources[i])

			// Make paths which are relative to manifest, relative to CWD.
			if !filepath.IsAbs(src) && src[0] != '$' {
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

func (s *Suite) Imports() ([]*File, error) {
	var ret []*File

	// Environment variable overwrite everything.
	if env := s.Getenv("NTT_IMPORTS"); env != "" {
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
	if m != nil && len(m.Imports) > 0 && s.root != nil {
		for i := range m.Imports {
			// Substitute environment variables
			path := s.expand(m.Imports[i])

			// Make paths which are relative to manifest, relative to CWD.
			if !filepath.IsAbs(path) && path[0] != '$' {
				path = filepath.Clean(filepath.Join(s.root.Path(), path))
			}

			ret = append(ret, s.File(path))

		}
		return append(ret, s.sources...), nil
	}

	// Last resort is imports list, explicitly curated by AddImports-calls.
	return s.imports, nil
}

func (s *Suite) AddSources(files ...string) {
	for i := range files {
		s.sources = append(s.sources, s.File(files[i]))
	}
}

func (s *Suite) AddImports(folders ...string) {
	for i := range folders {
		s.imports = append(s.imports, s.File(folders[i]))
	}
}

func (s *Suite) Name() (string, error) {
	if env := s.Getenv("NTT_NAME"); env != "" {
		return env, nil
	}

	// TODO(5nord) Should have SetName a higher priority than package.yml?
	if s.name != "" {
		return s.name, nil
	}

	// If there's a parseable package.yml, try that one.
	m, err := s.parseManifest()
	if err != nil {
		return "", err
	}
	if m != nil && m.Name != "" {
		return s.expand(m.Name), nil
	}

	// If there's a root dir, use its name.
	if s.root != nil {
		return filepath.Base(s.root.URI().Filename()), nil
	}

	// As last resort, try to find a name in source files.
	srcs, err := s.Sources()
	if err != nil {
		return "", err
	}
	if len(srcs) > 0 {
		n, err := filepath.Abs(srcs[0].Path())
		if err != nil {
			return "", err
		}
		n = filepath.Base(n)
		n = strings.TrimSuffix(n, filepath.Ext(n))
		return n, nil
	}

	return "", fmt.Errorf("Could not determine a suite name")
}

func (s *Suite) SetName(name string) {
	s.name = name
}

// TestHook return the File object to the test hook. If not hook was found, it
// will return nil. If an error occurred, like a parse error, then error is set
// appropriately.
func (s *Suite) TestHook() (*File, error) {
	if env := s.Getenv("NTT_TEST_HOOK"); env != "" {
		return s.File(env), nil
	}

	// If there's a parseable package.yml, try that one.
	m, err := s.parseManifest()
	if err != nil {
		return nil, err
	}
	if m != nil && m.TestHook != "" {
		path := s.expand(m.TestHook)
		if !filepath.IsAbs(path) && path[0] != '$' {
			path = filepath.Clean(filepath.Join(s.root.Path(), path))
		}

		return s.File(path), nil
	}

	// Construct default name
	filename, err := s.Name()
	if err != nil || filename == "" {
		return nil, err
	}
	filename = filename + ".control"

	// Look for hook in root folder
	if s.root != nil {
		hook := filepath.Join(s.root.Path(), filename)
		ok, err := fileExists(hook)
		if err != nil {
			return nil, err
		}
		if ok {
			return s.File(hook), nil
		}
	}

	// As last resort use directory of first source file
	srcs, err := s.Sources()
	if err != nil {
		return nil, err
	}
	if len(srcs) > 0 {
		hook := filepath.Join(filepath.Dir(srcs[0].Path()), filename)
		ok, err := fileExists(hook)
		if err != nil {
			return nil, err
		}
		if ok {
			return s.File(hook), nil
		}

	}

	return nil, nil
}

// ParametersFile return the File object to the parameter file. If no file was found, it
// will return nil. If an error occurred, like a parse error, then error is set
// appropriately.
func (s *Suite) ParametersFile() (*File, error) {
	if env := s.Getenv("NTT_PARAMETERS_FILE"); env != "" {
		return s.File(env), nil
	}

	// If there's a parseable package.yml, try that one.
	m, err := s.parseManifest()
	if err != nil {
		return nil, err
	}
	if m != nil && m.ParametersFile != "" {
		path := s.expand(m.ParametersFile)
		if !filepath.IsAbs(path) && path[0] != '$' {
			path = filepath.Clean(filepath.Join(s.root.Path(), path))
		}

		return s.File(path), nil
	}

	// Construct default name
	filename, err := s.Name()
	if err != nil || filename == "" {
		return nil, err
	}
	filename = filename + ".parameters"

	// Look for hook in root folder
	if s.root != nil {
		path := filepath.Join(s.root.Path(), filename)
		ok, err := fileExists(path)
		if err != nil {
			return nil, err
		}
		if ok {
			return s.File(path), nil
		}
	}

	// As last resort use directory of first source file
	srcs, err := s.Sources()
	if err != nil {
		return nil, err
	}
	if len(srcs) > 0 {
		path := filepath.Join(filepath.Dir(srcs[0].Path()), filename)
		ok, err := fileExists(path)
		if err != nil {
			return nil, err
		}
		if ok {
			return s.File(path), nil
		}

	}

	return nil, nil
}
func fileExists(path string) (bool, error) {

	info, err := os.Stat(path)
	if err != nil {
		// It's okay if a file does not exist.
		if os.IsNotExist(err) {
			return false, nil
		}
		// But report any other bad errors.
		return false, err
	}

	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("%q is not a regular file", path)
	}

	return true, nil
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
