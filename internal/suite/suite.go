/*
Package Suite manages test suite configuration. A test suite configuration is
everything you need for building or running tests: TTCN-3, C/C++ and ASN.1 source
files, codecs adapters, hooks, ...

Please note, the TTCN-3 and ASN.1 specs do not have the notion of packages or
even files; only modules in a global scope. This becomes hard to organize and
maintain when the number of files grows. Hence this package introduces the
concept of packages similar to Go:

A package is a directory containing TTCN-3, ASN.1 or C/C++ source files. A test
suite is a special package with a configuration file (`package.yml`). This
configuration file specifies dependencies to other packages, as well as runtime
specific specific configurations like test_hooks, timeout, module parameters,
... . The name of the package is the base-name of the directory. So choose the
name carefully. Everything which belongs together should be put together.
Package names should describe their purpose, they should be short and easy to
understand. Avoid catchall packages like: `utils`, `helpers`, `functions`, ...

You may use environment variables and variables from `k3.env` at any place, also
in `package.yml`. They become expanded when the suite is loaded.

Package YAML

The main package (aka suite) may contain a YAML configuration file `package.yml` to specify
dependencies to other package and configure custom runtime configuration. Below
keys are supported:

        name                 Name of the package. Default is the directory name.

        sources              Source files.

        imports              Directories the suite depends on.

        timeout              Default timeout for tests in seconds.

        test_hook            Path to test_hook script. Default is `$K3_NAME.control`.

        parameters_file      Path to module parameters file. Default is `$K3_NAME.parameters`


All keys can be overwritten by environment variables. For instance, to overwrite
`test_hook` just set environment variable `K3_TEST_HOOK`.  Is it not
recommended to overwrite list values like `K3_SOURCES` and `K3_IMPORTS`,
because you might run into quoting issues with white-spaces.


Ad-hoc Packages

It's is possible to create a suite just from a directory without any
`package.yml`. The Config.Sources will contain the paths of all .ttcn3 sources
files in that directory. Alternatively you may specify a list of .ttcn3 files
directory. The associated will then be the current working directory.

Like with the configuration file, most values can be overwritten by environment
variables.

*/
package suite

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/fs"
	yaml "gopkg.in/yaml.v2"
)

// A Suite is a main Package with additional configuration.
type Suite struct {
	Package `yaml:",inline"`

	// Runtime configuration
	TestHook       string  `yaml:test_hook`       // Path for test hook.
	ParametersFile string  `yaml:parameters_file` // Path for module parameters file.
	Timeout        float64 `yaml:timeout`         // Global timeout for tests.

	env        *env.Env
	configFile string
}

// NewFromArgs creates a suite from command line arguments. It expects either
// one directory as argument or a list of regular .ttcn3 files.
//
// Calling NewFromArgs with an empty argument list will create a suite from
// current working directory or, if set, from K3_SOURCE_DIR.
func NewFromArgs(args []string) (*Suite, error) {
	switch len(args) {
	case 0:
		if source_dir := env.DefaultEnv.Get("source_dir"); source_dir != "" {
			return NewFromDirectory(source_dir)
		}
		return NewFromDirectory(".")
	case 1:
		if fs.IsDirectory(args[0]) {
			return NewFromDirectory(args[0])
		}
	}

	return NewFromFiles(args) // Multiple arguments are interpreted as file list.

}

// NewFromFiles creates a suite from a TTCN-3 sources list The current working
// directory will be associated with the suite, however `package.yml` will be
// ignored.
//
// Except for `K3_SOURCES` You may use the usual environment variables to
// overwrite default configuration.
//
// Allowing only ttcn3 files is a compromise between keeping k3 tool simple and
// allowing user to create ad-hoc packages for proto-typing. This constraint
// might be lifted in the future.
func NewFromFiles(files []string) (*Suite, error) {
	for _, file := range files {
		if !fs.IsRegular(file) || !fs.IsTTCN3File(file) {
			return nil, fmt.Errorf("named files must be regular .ttcn3 files: %s", file)
		}
	}

	suite := New(".")
	if err := suite.readEnv(); err != nil {
		return nil, err
	}
	suite.Sources = files
	return suite, nil
}

// NewFromDirectory creates a suite from directory dir. If the directory
// contains a configuration file (`package.yml`) it will be loaded. You may use
// the usual environment variables to overwrite configuration. If no source
// files were specified, NewFromDirectory will assign all .ttcn3 it finds.
//
// NewFromDirectory returns an error if package.yml was not readable or hab
// syntax errors or if environment variables have from type (e.g. float64 for
// K3_TIMEOUT)
func NewFromDirectory(dir string) (*Suite, error) {
	suite := New(dir)
	if suite.configFile != "" {
		err := suite.readConfig()
		if err != nil {
			return nil, err
		}
	}

	if err := suite.readEnv(); err != nil {
		return nil, err
	}

	// If we still don't have sources specified, take .ttcn3 files from directory.
	if suite.Sources == nil {
		files, err := fs.TTCN3Files(suite.Dir())
		if err != nil {
			return nil, err
		}
		suite.Sources = files
	}

	return suite, nil
}

// New creates a new Suite with default values.
func New(dir string) *Suite {
	suite := &Suite{}
	suite.init(dir)
	suite.env = env.DefaultEnv
	suite.configFile = fs.FindFile(suite.dir, "package.yml")
	suite.TestHook = fs.FindFile(suite.dir, suite.Name+".control")
	suite.ParametersFile = fs.FindFile(suite.dir, suite.Name+".parameters")
	return suite
}

// SetEnv stores suites configuration in the environment.
func (suite *Suite) SetEnv() {
	suite.env.Set("name", suite.Name)
	suite.env.Set("sources", strings.Join(suite.Sources, " "))
	suite.env.Set("imports", strings.Join(suite.Imports, " "))
	suite.env.Set("test_hook", suite.TestHook)
	suite.env.Set("parameters_file", suite.ParametersFile)
	suite.env.Set("timeout", strconv.FormatFloat(suite.Timeout, 'G', -1, 64))

	if abs, err := filepath.Abs(suite.Dir()); err == nil {
		suite.env.Set("source_dir", abs)
	}

	// TODO(5nord) Add K3_TTCN3_FILES
}

// readConfig reads `package.yml` values into the suite
func (suite *Suite) readConfig() error {
	r, err := os.Open(suite.configFile)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)

	if err := yaml.UnmarshalStrict(buf.Bytes(), suite); err != nil {
		return err
	}

	if len(suite.Sources) != 0 {
		for i, path := range suite.Sources {
			suite.Sources[i] = suite.fixPath(path)
		}
	}

	if len(suite.Imports) != 0 {
		for i, path := range suite.Imports {
			suite.Imports[i] = suite.fixPath(path)
		}
	}

	if suite.TestHook != "" {
		suite.TestHook = suite.fixPath(suite.TestHook)
	}

	if suite.ParametersFile != "" {
		suite.ParametersFile = suite.fixPath(suite.ParametersFile)
	}

	return nil
}

// fixPath makes a relativ existing path, relative to CWD
func (suite *Suite) fixPath(path string) string {
	path = suite.env.Expand(path)
	if rel := filepath.Clean(filepath.Join(suite.Dir(), path)); fs.IsExist(rel) && !filepath.IsAbs(path) {
		return rel
	}
	return path
}

// readEnv reads the usual evironment variables and overwrites the suites
// configuration.
func (suite *Suite) readEnv() error {
	if env := suite.env.Get("name"); env != "" {
		suite.SetName(env)
	}
	if env := suite.env.Get("test_hook"); env != "" {
		suite.TestHook = env
	}
	if env := suite.env.Get("parameters_file"); env != "" {
		suite.ParametersFile = env
	}
	if env := suite.env.Get("sources"); env != "" {
		suite.Sources = strings.Split(env, " ")
	}
	if env := suite.env.Get("imports"); env != "" {
		suite.Imports = strings.Split(env, " ")
	}
	if env := suite.env.Get("timeout"); env != "" {
		f, err := strconv.ParseFloat(env, 64)
		if err != nil {
			return err
		}
		suite.Timeout = f
	}

	suite.SetDir(suite.env.Expand(suite.Dir()))
	suite.SetName(suite.env.Expand(suite.Name))

	return nil
}
