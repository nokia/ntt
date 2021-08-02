package ntt

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/project"
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
func (suite *Suite) Sources() ([]string, error) {
	if suite.p == nil {
		suite.p = &project.Project{}
	}
	return suite.p.Sources()
}

// Imports returns the list of imported packages required to compile a Suite.
// The error will be != nil if imports could not be determined correctly. For
// example, when `package.yml` had syntax errors.
func (suite *Suite) Imports() ([]string, error) {
	if suite.p == nil {
		suite.p = &project.Project{}
	}
	return suite.p.Imports()
}

// AddSources appends files... to the known sources list.
func (suite *Suite) AddSources(files ...string) {
	if suite.p == nil {
		suite.p = &project.Project{}
	}
	suite.p.Manifest.Sources = append(suite.p.Manifest.Sources, files...)
}

// Name returns the name of the test suite. Or err != nil if the name could not
// determined correctly.
//
// Name first checks for environment variable NTT_NAME, then if a name is
// specified in a manifest. Then if a name was specified explicitly. Next, if
// suite has a root folder, its base-name will be used. Last resort is the
// first source .ttcn3 without extension.
func (suite *Suite) Name() (string, error) {
	if suite.p == nil {
		suite.p = &project.Project{}
	}
	return suite.p.Name(), nil
}

// SetName will set the suites name.
func (suite *Suite) SetName(name string) {
	suite.p.Manifest.Name = name
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
		return fs.Open(env), nil
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
			path = filepath.Clean(filepath.Join(fs.Path(suite.root), path))
		}

		return fs.Open(path), nil
	}

	// Construct default name
	filename, err := suite.Name()
	if err != nil || filename == "" {
		return nil, err
	}
	filename = filename + ".control"

	// Look for hook in root folder
	if suite.root != "" {
		hook, _ := filepath.Abs(filepath.Join(fs.Path(suite.root), filename))
		ok, err := fileExists(hook)
		if err != nil {
			return nil, err
		}
		if ok {
			return fs.Open(hook), nil
		}
	}

	// As last resort use directory of first source file
	srcs, err := suite.Sources()
	if err != nil {
		return nil, err
	}
	if len(srcs) > 0 {
		hook, _ := filepath.Abs(filepath.Join(filepath.Dir(fs.Path(srcs[0])), filename))
		ok, err := fileExists(hook)
		if err != nil {
			return nil, err
		}
		if ok {
			return fs.Open(hook), nil
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
			paramDir, err = filepath.Abs(filepath.Join(fs.Path(suite.root), paramDir))
		}
		return paramDir, err
	}
	if suite.root != "" {
		return filepath.Abs(fs.Path(suite.root))
	}
	return "", err
}

// ParametersFile return the File object to the parameter file. If no file was found, it
// will return nil. If an error occurred, like a parse error, then error is set
// appropriately.
func (suite *Suite) ParametersFile() (*fs.File, error) {
	var pDir string = ""
	env, err := suite.Getenv("NTT_PARAMETERS_FILE")
	if err != nil {
		return nil, err
	}

	if env != "" && filepath.IsAbs(env) {
		return fs.Open(env), nil
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
			return fs.Open(path), nil
		}

	} else if env != "" {
		return fs.Open(env), nil
	}

	// If there's a parseable package.yml, try that one.
	m, err := suite.parseManifest()
	if err != nil {
		return nil, err
	}

	if m != nil && m.ParametersFile != "" {
		path, err := suite.Expand(m.ParametersFile)
		if err != nil {
			return nil, err
		}
		if !filepath.IsAbs(path) && path[0] != '$' {
			if pDir != "" {
				path = filepath.Clean(filepath.Join(pDir, path))
			} else {
				path = filepath.Clean(filepath.Join(fs.Path(suite.root), path))
			}
		}
		return fs.Open(path), nil
	}

	// Construct default name
	filename, err := suite.Name()
	if err != nil || filename == "" {
		return nil, err
	}
	filename = filename + ".parameters"

	// Look for hook in root folder
	if suite.root != "" {
		path := ""
		if pDir != "" {
			path = filepath.Clean(filepath.Join(pDir, filename))
		} else {
			path = filepath.Join(fs.Path(suite.root), filename)
		}
		ok, err := fileExists(path)
		if err != nil {
			return nil, err
		}
		if ok {
			return fs.Open(path), nil
		}
	}

	// As last resort use directory of first source file
	srcs, err := suite.Sources()
	if err != nil {
		return nil, err
	}
	if len(srcs) > 0 {
		path := filepath.Join(filepath.Dir(fs.Path(srcs[0])), filename)
		ok, err := fileExists(path)
		if err != nil {
			return nil, err
		}
		if ok {
			return fs.Open(path), nil
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
	if suite.root == "" {
		return nil, nil
	}

	f := fs.Open(filepath.Join(fs.Path(suite.root), "package.yml"))
	log.Debugf("Open manifest %q\n", f.Path())
	b, err := f.Bytes()
	if err != nil {
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

// FindAllFiles returns all .ttcn3 files including auxiliary files from
// k3 installation
func (suite *Suite) FindAllFiles() []string {
	files, _ := project.Files(suite)
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
	f := fs.Open(name + ".ttcn3")
	if _, err := f.Bytes(); err == nil {
		suite.modules[name] = f.Path()
		return f.Path(), nil
	}

	for _, file := range suite.FindAllFiles() {
		if filepath.Base(file) == name+".ttcn3" {
			suite.modules[name] = file
			return file, nil
		}
	}

	return "", fmt.Errorf("No such module %q", name)
}
