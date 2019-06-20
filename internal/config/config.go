/*
Package config manages test suite configurations. A "test suite configuration"
is everything you need for building or running tests: TTCN-3 and ASN.1 files,
codecs, adapters, hooks, ...

Please note, the TTCN-3 and ASN.1 specs do not have the notion of packages or
even files; only modules in a global scope. This becomes hard to organize and
maintain when the number of files grows. Hence we try to introduce the concept
of packages similar to Go:

A test suite is built of "packages"; a main package and packages the suite
depends on, like codecs, adapters and libraries.

Usually packages and directories are the same thing, choose their name
carefully. Everything which belongs together should be put together. Package
names should describe their purpose, they should be short and easy to
understand. Avoid catchall packages like: `utils`, `helpers`, `functions`, ...


Auto-configuration

When loading a package with `config.FromArgs`


Package YAML

The main package may contain a YAML configuration file `package.yml` to specify
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
because you might run into quoting issue with white-spaces.

*/
package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// A Config describes how to build and configure a test suite.
type Config struct {
	Dir     string   // Absolute directory of suite
	Name    string   // Name of suite
	Sources []string // Source files of suite
	Imports []string // Requires libraries, adapters, ... of suite

	ParametersFile string  `yaml:parameters_file` // Path for module parameters
	TestHook       string  `yaml:test_hook`       // Path for test hook
	Timeout        float64 // Global timeout for tests
}

// Load configuration from `package.yml`, if available.
func (conf *Config) loadConfig() error {
	file := filepath.Join(conf.Dir, "package.yml")
	if _, err := os.Stat(file); err != nil {
		return nil
	}

	b, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	return yaml.Unmarshal([]byte(b), conf)
}

// loadEnv assigns environment variables to Config fields.
func (conf *Config) loadEnv() {

	if env := os.Getenv("K3_NAME"); env != "" {
		conf.Name = env
	}

	if env := os.Getenv("K3_PARAMETERS_FILE"); env != "" {
		conf.ParametersFile = env
	}

	if env := os.Getenv("K3_TEST_HOOK"); env != "" {
		conf.TestHook = env
	}

	if env := os.Getenv("K3_TIMEOUT"); env != "" {
		conf.Timeout, _ = strconv.ParseFloat(env, 64)
	}

	if env := os.Getenv("K3_SOURCES"); env != "" {
		conf.Sources = strings.Fields(env)
	}

	if env := os.Getenv("K3_IMPORTS"); env != "" {
		conf.Imports = strings.Fields(env)
	}

}

// loadDefaults assigns default values to empty Config fields.
func (conf *Config) loadDefaults() {

	if conf.Name == "" {
		n, _ := filepath.Abs(conf.Dir)
		conf.Name = filepath.Base(n)
	}

	if conf.ParametersFile == "" {
		parameters_file := filepath.Join(conf.Dir, conf.Name, ".parameters")
		if b, _ := isRegular(parameters_file); b {
			conf.ParametersFile = parameters_file
		}
	}

	if conf.TestHook == "" {
		test_hook := filepath.Join(conf.Dir, conf.Name, ".control")
		if b, _ := isRegular(test_hook); b {
			conf.TestHook = test_hook
		}
	}
}

func (conf *Config) expandPaths() {
	for i, _ := range conf.Sources {
		conf.Sources[i] = conf.expandPath(conf.Sources[i])
	}
	for i, _ := range conf.Imports {
		conf.Imports[i] = conf.expandPath(conf.Imports[i])
	}

	if conf.ParametersFile != "" {
		conf.ParametersFile = conf.expandPath(conf.ParametersFile)
	}
	if conf.TestHook != "" {
		conf.TestHook = conf.expandPath(conf.TestHook)
	}
}

func (conf *Config) expandPath(path string) string {
	path = expandEnv(path)
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(conf.Dir, path)
}

// FromArgs returns a Config struct based on command line arguments. It expects
// either one directory as argument or a list of regular .ttcn3 files.
//
// Calling FromArgs with an empty argument list is equivalent to FromDirectory(".").
func FromArgs(args []string) (*Config, error) {

	// Directory arguments
	switch len(args) {
	case 0:
		return fromDirectory(".")
	case 1:
		if isDir, _ := isDirectory(args[0]); isDir {
			return fromDirectory(args[0])
		}
	}

	// Regular file list
	for _, arg := range args {
		if isReg, err := isRegular(arg); !isReg || !isTTCN3File(arg) {
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("named files must be regular .ttcn3 files: %s", arg)
		}
	}
	return fromFiles(args)
}

// fromDirectory returns a Config struct.
func fromDirectory(dir string) (*Config, error) {
	conf := &Config{Dir: dir}

	if err := conf.loadConfig(); err != nil {
		return nil, err
	}
	if conf.Sources == nil {
		files, err := FindTTCN3Files(conf.Dir)
		if err != nil {
			return nil, err
		}
		conf.Sources = files
	}
	conf.loadEnv()
	conf.loadDefaults()
	conf.expandPaths()
	return conf, nil
}

// fromFiles returns a ad-hoc Config struct with files as sources.
func fromFiles(files []string) (*Config, error) {
	conf := &Config{
		Dir:     ".",
		Sources: files,
	}

	conf.loadEnv()
	conf.loadDefaults()
	conf.expandPaths()
	return conf, nil
}

// FindTTCN3Files returns a .ttcn3 (or .ttcn) source files from directory dir.
func FindTTCN3Files(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	sources := make([]string, 0, len(files))
	for _, file := range files {
		if file.Mode().IsRegular() && isTTCN3File(file.Name()) {
			sources = append(sources, file.Name())
		}
	}
	return sources, nil
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), err
}

func isRegular(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.Mode().IsRegular(), err
}

func isTTCN3File(name string) bool {
	return strings.HasSuffix(name, ".ttcn3") || strings.HasSuffix(name, ".ttcn")
}

func expandEnv(s string) string {
	mapper := func(name string) string {
		if env := os.Getenv(name); env != "" {
			return env
		}
		return fmt.Sprintf("${%s}", name)
	}

	return os.Expand(s, mapper)
}
