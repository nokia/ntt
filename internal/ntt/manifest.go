package ntt

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/k3"
	"gopkg.in/yaml.v2"
)

// Timeout returns the configured timeout of the Suite. If there's no timeout,
// the function will return 0.
//
// The error will be != nil if timeout could not be determined correctly. For
// example, when `package.yml` had syntax errors.
func (suite *Suite) Timeout() (float64, error) {
	s, err := suite.Getenv("NTT_TIMEOUT")
	if err != nil {
		return 0, err
	}
	if s != "" {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, err
		}
		return f, nil
	}
	m, err := suite.parseManifest()
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
func (suite *Suite) Sources() ([]*fs.File, error) {
	var ret []*fs.File

	// Environment variable overwrite everything.
	env, err := suite.Getenv("NTT_SOURCES")
	if err != nil {
		return nil, err
	}
	if env != "" {
		for _, x := range strings.Fields(env) {
			ret = append(ret, suite.File(x))
		}
		return ret, nil
	}

	// If there's a parseable package.yml, try that one.
	m, err := suite.parseManifest()
	if err != nil {
		return nil, err
	}
	if m != nil && len(m.Sources) > 0 && suite.root != nil {
		for i := range m.Sources {
			// Substitute environment variables
			src, err := suite.Expand(m.Sources[i])
			if err != nil {
				return nil, err
			}

			// Make paths which are relative to manifest, relative to CWD.
			if !filepath.IsAbs(src) && src[0] != '$' {
				src = filepath.Clean(filepath.Join(suite.root.Path(), src))
			}

			// Directories need expansion into single files.
			info, err := os.Stat(src)
			if err != nil {
				return nil, err
			}
			switch {
			case info.IsDir():
				files := fs.FindTTCN3Files(src)
				if len(files) == 0 {
					return nil, fmt.Errorf("Could not find ttcn3 source files in directory %q", src)
				}
				for i := range files {
					ret = append(ret, suite.File(files[i]))
				}

			case info.Mode().IsRegular() && fs.HasTTCN3Extension(src):
				ret = append(ret, suite.File(src))

			default:
				return nil, fmt.Errorf("Cannot handle %q. Expecting directory or ttcn3 source file", src)
			}

		}
		return append(ret, suite.sources...), nil
	}

	// If there's only a root folder, look for .ttcn3 files
	if suite.root != nil {
		files := fs.FindTTCN3Files(suite.root.Path())
		for _, f := range files {
			ret = append(ret, suite.File(f))
		}
		return append(ret, suite.sources...), nil

	}

	// Last resort is sources list, explicitly curated by AddSources-calls.
	return suite.sources, nil
}

// Imports returns the list of imported packages required to compile a Suite.
// The error will be != nil if imports could not be determined correctly. For
// example, when `package.yml` had syntax errors.
func (suite *Suite) Imports() ([]*fs.File, error) {
	var ret []*fs.File

	// Environment variable overwrite everything.
	env, err := suite.Getenv("NTT_IMPORTS")
	if err != nil {
		return nil, err
	}
	if env != "" {
		for _, x := range strings.Fields(env) {
			ret = append(ret, suite.File(x))
		}
		return ret, nil
	}

	// If there's a parseable package.yml, try that one.
	m, err := suite.parseManifest()
	if err != nil {
		return nil, err
	}
	if m != nil && len(m.Imports) > 0 && suite.root != nil {
		for i := range m.Imports {
			// Substitute environment variables
			path, err := suite.Expand(m.Imports[i])
			if err != nil {
				return nil, err
			}

			// Make paths which are relative to manifest, relative to CWD.
			if !filepath.IsAbs(path) && path[0] != '$' {
				path = filepath.Clean(filepath.Join(suite.root.Path(), path))
			}

			ret = append(ret, suite.File(path))

		}
		return append(ret, suite.imports...), nil
	}

	// Last resort is imports list, explicitly curated by AddImports-calls.
	return suite.imports, nil
}

// AddSources appends files... to the known sources list.
func (suite *Suite) AddSources(files ...string) {
	for i := range files {
		suite.sources = append(suite.sources, suite.File(files[i]))
	}
}

// AddImports appends folders.. to the known imports list.
func (suite *Suite) AddImports(folders ...string) {
	for i := range folders {
		suite.imports = append(suite.imports, suite.File(folders[i]))
	}
}

// Name returns the name of the test suite. Or err != nil if the name could not
// determined correctly.
//
// Name first checks for environment variable NTT_NAME, then if a name is
// specified in a manifest. Then if a name was specified explicitly. Next, if
// suite has a root folder, its base-name will be used. Last resort is the
// first source .ttcn3 without extension.
func (suite *Suite) Name() (string, error) {
	if suite.name != "" {
		return suite.name, nil
	}

	env, err := suite.Getenv("NTT_NAME")
	if err != nil {
		return "", err
	}
	if env != "" {
		return env, nil
	}

	// If there's a parseable package.yml, try that one.
	m, err := suite.parseManifest()
	if err != nil {
		return "", err
	}
	if m != nil && m.Name != "" {
		return suite.Expand(m.Name)
	}

	// If there's a root dir, use its name.
	if suite.root != nil {
		return filepath.Base(suite.root.URI().Filename()), nil
	}

	// As last resort, try to find a name in source files.
	srcs, err := suite.Sources()
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

// SetName will set the suites name.
func (suite *Suite) SetName(name string) {
	suite.name = name
}

// Variables will return a string map containing the variables-sections of the
// manifest files.
func (suite *Suite) Variables() (map[string]string, error) {
	m, err := suite.parseManifest()
	if err != nil {
		return nil, err
	}
	if m != nil && m.Variables != nil {
		return m.Variables, nil
	}
	return nil, nil
}

// TestHook return the File object to the test hook. If not hook was found, it
// will return nil. If an error occurred, like a parse error, then error is set
// appropriately.
func (suite *Suite) TestHook() (*fs.File, error) {
	env, err := suite.Getenv("NTT_TEST_HOOK")
	if err != nil {
		return nil, err
	}
	if env != "" {
		return suite.File(env), nil
	}

	// If there's a parseable package.yml, try that one.
	m, err := suite.parseManifest()
	if err != nil {
		return nil, err
	}
	if m != nil && m.TestHook != "" {
		path, err := suite.Expand(m.TestHook)
		if err != nil {
			return nil, err
		}
		if !filepath.IsAbs(path) && path[0] != '$' {
			path = filepath.Clean(filepath.Join(suite.root.Path(), path))
		}

		return suite.File(path), nil
	}

	// Construct default name
	filename, err := suite.Name()
	if err != nil || filename == "" {
		return nil, err
	}
	filename = filename + ".control"

	// Look for hook in root folder
	if suite.root != nil {
		hook, _ := filepath.Abs(filepath.Join(suite.root.Path(), filename))
		ok, err := fileExists(hook)
		if err != nil {
			return nil, err
		}
		if ok {
			return suite.File(hook), nil
		}
	}

	// As last resort use directory of first source file
	srcs, err := suite.Sources()
	if err != nil {
		return nil, err
	}
	if len(srcs) > 0 {
		hook, _ := filepath.Abs(filepath.Join(filepath.Dir(srcs[0].Path()), filename))
		ok, err := fileExists(hook)
		if err != nil {
			return nil, err
		}
		if ok {
			return suite.File(hook), nil
		}

	}

	return nil, nil
}

// ParametersDir return the absolute path retrieven from either
// * parameters_dir field from manifest
// * suites root path.
func (suite *Suite) ParametersDir() (string, error) {
	if env, err := suite.Getenv("NTT_PARAMETERS_DIR"); err == nil {
		if env != "" {
			return env, nil
		}
	} else {
		return "", err
	}

	// If there's a parseable package.yml, try that one.
	m, err := suite.parseManifest()
	if err != nil {
		return "", err
	}
	if m != nil && m.ParametersDir != "" {
		paramDir, err := suite.Expand(m.ParametersDir)
		if err != nil {
			return "", err
		}
		if !filepath.IsAbs(paramDir) && paramDir[0] != '$' {
			paramDir, err = filepath.Abs(filepath.Join(suite.root.Path(), paramDir))
		}
		return paramDir, err
	}
	if suite.root != nil {
		return filepath.Abs(suite.root.Path())
	}
	return "", err
}

// ParametersFile return the File object to the parameter file. If no file was found, it
// will return nil. If an error occurred, like a parse error, then error is set
// appropriately.
func (suite *Suite) ParametersFile() (*fs.File, error) {
	var pDir string = ""
	env, err := suite.Getenv("NTT_PARAMETERS_FILE")
	log.Debug(fmt.Sprintf("ParametersFile1: %q, error: %q", env, err))
	if err != nil {
		return nil, err
	}

	if env != "" && filepath.IsAbs(env) {
		return suite.File(env), nil
	}
	// first get the path to the root of the parameters file(s)
	if paramDir, err := suite.ParametersDir(); err == nil {
		pDir = paramDir
	} else {
		return nil, err
	}

	if pDir != "" {
		if env != "" {
			path := filepath.Clean(filepath.Join(pDir, env))
			log.Debug(fmt.Sprintf("ParametersFile2: %q, error: %q", path, err))
			return suite.File(path), nil
		}

	} else if env != "" {
		return suite.File(env), nil
	}

	// If there's a parseable package.yml, try that one.
	m, err := suite.parseManifest()
	if err != nil {
		return nil, err
	}

	if m != nil && m.ParametersFile != "" {
		path, err := suite.Expand(m.ParametersFile)
		log.Debug(fmt.Sprintf("ParametersFile3: %q, error: %q", path, err))
		if err != nil {
			return nil, err
		}
		if !filepath.IsAbs(path) && path[0] != '$' {
			if pDir != "" {
				path = filepath.Clean(filepath.Join(pDir, path))
			} else {
				path = filepath.Clean(filepath.Join(suite.root.Path(), path))
			}
		}
		log.Debug(fmt.Sprintf("ParametersFile4: %q, error: %q", path, err))
		return suite.File(path), nil
	}

	// Construct default name
	filename, err := suite.Name()
	if err != nil || filename == "" {
		return nil, err
	}
	filename = filename + ".parameters"

	// Look for hook in root folder
	if suite.root != nil {
		path := ""
		if pDir != "" {
			path = filepath.Clean(filepath.Join(pDir, filename))
		} else {
			path = filepath.Join(suite.root.Path(), filename)
		}
		log.Debug(fmt.Sprintf("ParametersFile5: %q, error: %q", path, err))
		ok, err := fileExists(path)
		if err != nil {
			return nil, err
		}
		if ok {
			return suite.File(path), nil
		}
	}

	// As last resort use directory of first source file
	srcs, err := suite.Sources()
	if err != nil {
		return nil, err
	}
	if len(srcs) > 0 {
		path := filepath.Join(filepath.Dir(srcs[0].Path()), filename)
		log.Debug(fmt.Sprintf("ParametersFile6: %q, error: %q", path, err))
		ok, err := fileExists(path)
		if err != nil {
			return nil, err
		}
		if ok {
			return suite.File(path), nil
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
	Name      string
	Sources   []string
	Imports   []string
	Variables map[string]string

	// Runtime configuration
	TestHook       string  `yaml:"test_hook"`       // Path for test hook.
	ParametersFile string  `yaml:"parameters_file"` // Path for module parameters file.
	ParametersDir  string  `yaml:"parameters_dir"`  // Optional path for parameters_file.
	Timeout        float64 `yaml:"timeout"`         // Global timeout for tests.
}

// parseManifest tries to parse an (optional) manifest file.
func (suite *Suite) parseManifest() (*manifest, error) {
	// Without root folder, there's no manifest to parse. This is ok.
	if suite.root == nil {
		return nil, nil
	}

	f := suite.File(filepath.Join(suite.root.Path(), "package.yml"))
	log.Debugf("Open manifest %q\n", f.Path())
	b, err := f.Bytes()
	if err != nil {
		log.Debugln(err.Error())
		// If there's not package.yml, it's okay, too.
		if os.IsNotExist(err) {
			return nil, nil
		}
		// all other errors should be reported.
		return nil, err
	}

	type manifestData struct {
		manifest manifest
		err      error
	}

	f.Handle = suite.store.Bind(f.ID(), func(ctx context.Context) interface{} {
		data := manifestData{}
		data.err = yaml.UnmarshalStrict(b, &data.manifest)
		return &data
	})

	v := f.Handle.Get(context.TODO())
	data := v.(*manifestData)

	return &data.manifest, data.err
}

// Files returns all .ttcn3 available. It will not return generated .ttcn3 files.
// On error Files will return an error.
func (suite *Suite) Files() ([]string, error) {
	srcs, err := suite.Sources()
	if err != nil {
		return nil, err
	}
	files := fs.PathSlice(srcs...)

	dirs, err := suite.Imports()
	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {
		f := fs.FindTTCN3Files(dir.Path())
		files = append(files, f...)
	}
	return files, err
}

// FindAllFiles returns all .ttcn3 files including auxiliary files from
// k3 installation
func (suite *Suite) FindAllFiles() []string {
	files, _ := suite.Files()
	// Use auxilliaryFiles from K3 to locate file
	for _, dir := range k3.FindAuxiliaryDirectories() {
		for _, file := range fs.FindTTCN3Files(dir) {
			files = append(files, file)
		}
	}
	return files
}

// FindModule tries to find a .ttcn3 based on its module name.
func (suite *Suite) FindModule(name string) (string, error) {

	suite.modulesMu.Lock()
	defer suite.modulesMu.Unlock()

	if suite.modules == nil {
		suite.modules = make(map[string]string)
	}
	if file, ok := suite.modules[name]; ok {
		return file, nil
	}

	// Use NTT_CACHE to locate file
	f := suite.File(name + ".ttcn3")
	if _, err := f.Bytes(); err == nil {
		suite.modules[name] = f.Path()
		return f.Path(), nil
	}
	if files, err := suite.Files(); err == nil {
		for _, file := range files {
			if filepath.Base(file) == name+".ttcn3" {
				suite.modules[name] = file
				return file, nil
			}
		}
	}

	// Use auxilliaryFiles from K3 to locate file
	for _, dir := range k3.FindAuxiliaryDirectories() {
		for _, file := range fs.FindTTCN3Files(dir) {
			if filepath.Base(file) == name+".ttcn3" {
				suite.modules[name] = file
				return file, nil
			}
		}
	}

	return "", fmt.Errorf("No such module %q", name)
}

// IsOwned returns true if path is part of this test suite
func (suite *Suite) IsOwned(uri string) bool {
	path := fs.Open(uri).URI().Filename()

	files, _ := suite.Files()
	for _, file := range files {
		if (file == path) ||
			(file == uri) {
			return true
		}
	}
	return false
}
