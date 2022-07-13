// Package project provides a tool-independent interface for working with
// various TTCN-3 project layouts and configurations. The interface is
// intended to be uniform across all project layouts. Supported layouts will
// be extended over time.
//
// Here is a simple example, opening a project configuration:
//
// 	conf, err := project.Open("/path/to/project")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
//
package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/nokia/ntt/internal/cache"
	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/yaml"
	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
)

var (
	ErrNoSources = errors.New("no sources")
)

// Config describes a single project configuration. It aggregates various
// sources, such as manifest files, parameters files, environment files, ...
type Config struct {
	// The Manifest pulls in most of the project configuration.
	Manifest `json:",inline"`

	// ManifestFile is the path to the manifest file.
	ManifestFile string `json:"manifest_file"`

	// Root is the root directory of the project. Usually this is the
	// directory of the manifest file.
	Root string

	// SourceDir is the directory containing source files, data files or
	// additional configuration files. SourceDir and Root are usually
	// identical, unless manifest was generated into a dedicated build
	// directory.
	SourceDir string `json:"source_dir"`

	// EnvFile is the path to the environment file. The content of the
	// environment file will overwrite variables defined in the manifest.
	// Default:
	//
	// 	${NTT_CACHE}/ntt.env
	EnvFile string `json:"env_file"`

	K3 struct {
		// K3 root folder
		Root string `json:",omitempty"`

		// Path to the compiler.
		Compiler string

		// Path to the runtime.
		Runtime string

		// Path to $PREFIX/share/k3
		DataDir string `json:"data_dir"`

		// Path to additional TTCN-3 files
		Includes []string

		// Path to additional plugins
		Plugins []string

		// Path to OSS Nokalva installation.
		OssInfo string

		// T3XF is the path to the T3XF file.
		T3XF string `json:"t3xf"`
	}

	// what toolchain to use.
	toolchain string
}

// The Manifest file (package.yml).
type Manifest struct {
	// Name is a short name of the project. It is recommended to use TTCN-3
	// identifiers (A-Za-z0-9_).
	Name string

	Version     string
	Description string
	Author      string
	License     string
	Homepage    string
	Repository  string
	Bugs        string
	Keywords    []string

	// Sources is a list of source files which provide the TTCN-3
	// test-cases.
	Sources []string

	// Imports is list of dependencies, which are required to run TTCN-3
	// test-cases. E.g. common code, adapters, codecs, ...
	Imports []string

	// BeforeBuild is a list of shell commands to be executed before
	// building. An exit code unequal to 0 will cancel any further
	// execution.
	BeforeBuild []string `json:"before_build"`

	// AfterBuild is a list of shell commands to be executed after
	// building. An exit code unequal to 0 will cancel any further
	// execution.
	AfterBuild []string `json:"after_build"`

	// BeforeRun is execute before any tests are run. An exit code unequal
	// to 0 will cancel any further test execution.
	BeforeRun []string `json:"before_run"`

	// AfterRun is executed after all tests are run.
	AfterRun []string `json:"after_run"`

	// BeforeTest is executed before each test. An exit code unequal to 0
	// will prevent test execution.
	BeforeTest []string `json:"before_test"`

	// AfterTest is executed after each test. An exit code unequal to 0
	// will result in an error verdict.
	AfterTest []string `json:"after_test"`

	// Variables is a list of variables that can be used in the
	// configuration files. Environment variables and variables defined in
	// the environment file (ntt.env) will overwrite variables defined in
	// this variables section.
	Variables env.Env

	// Parameters is an embedded parameters file.
	Parameters `json:",inline"`

	// ParametersFile is the path to the parameters file. Default:
	//
	// 	${NTT_SOURCE_DIR}/${NTT_NAME}.parameters
	ParametersFile string `json:"parameters_file"`

	// HooksFile is the path to the hooks file. Default:
	//
	// 	${NTT_SOURCE_DIR}/${NTT_NAME}.hooks
	HooksFile string `json:"hooks_file"`

	// LintFile is the path to the lint configuration file. Default:
	//
	// ${NTT_SOURCE_DIR}/ntt-lint.yml
	LintFile string `json:"lint_file"`
}

// The Parameters file provide runtime configuration for a project (e.g. parameters files)
type Parameters struct {
	// Global test configuration
	TestConfig `json:",inline"`

	// Presets provides configuration presets, which can be used for global
	// and test specific configuration.
	Presets map[string]TestConfig

	// Execute provides a list of test specific configuration. Each entry
	// specifies how and when a test should be executed.
	Execute []TestConfig
}

// A TestConfig specifies how and when a testcase should be executed.
type TestConfig struct {
	// A pattern describing a test. Optional testcase parameters are allowed.
	Test string `json:",omitempty"`

	// Presets is a list of preset configurations to be used to execute
	// the test.
	Preset []string `json:",omitempty"`

	// Timeout in seconds.
	Timeout yaml.Duration `json:",omitempty"`

	// Module parameters
	Parameters map[string]string `json:",omitempty"`

	// Only execute testcase if the given conditions are met.
	Only *ExecuteCondition `json:",omitempty"`

	// Do not execute testcase if the given conditions are met.
	Except *ExecuteCondition `json:",omitempty"`
}

// ExecuteCondition specifies conditions for executing a testcase.
type ExecuteCondition struct {
	Presets []string
}

// Index specifies where to look for project configuration files.
type Index struct {
	// SourceDir is the top level source directory.
	SourceDir string `json:"source_dir"`

	// BinaryDir is the top level binary directory.
	BinaryDir string `json:"binary_dir"`

	// Suite is a list of directories to search for project configuration files.
	Suites []Suite `json:"suites"`
}

// Suite specifies the root and source directories of a project.
type Suite struct {
	// RootDir is the root directory of the project. This is usually where the package.yml is located.
	RootDir string `json:"root_dir"`

	// SourceDir is the directory where the test suite source files are located.
	SourceDir string `json:"source_dir"`
}

var (
	ManifestFile = "package.yml"
	IndexFile    = "ttcn3_suites.json"
)

// Discover walks towards the file system root and collects
// known test suite layouts.
//
// Discover returns a list of potential test suite root directories.
func Discover(path string) []Suite {

	// Convert possible URIs to proper file system paths.
	path = fs.Path(path)

	var list []Suite

	// Return index, ignoring errors.
	readIndices := func(file string) []Suite {
		b, err := fs.Content(file)
		if err != nil {
			log.Debugf("Failed to read %s: %s", file, err.Error())
			return nil
		}
		var idx Index
		if err := json.Unmarshal(b, &idx); err != nil {
			log.Debugf("%s: %s", file, err.Error())
		}

		var list []Suite
		for _, s := range idx.Suites {
			if s.RootDir != "" {
				root := fs.Real(filepath.Dir(file), s.RootDir)
				log.Debugf("using root_dir: %q\n", root)
				list = append(list, s)
			}
		}
		return list
	}

	fs.WalkUp(path, func(path string) bool {
		// Check source directories
		if file := fs.JoinPath(path, ManifestFile); fs.IsRegular(file) {
			log.Debugf("discovered manifest: %q\n", file)
			list = append(list, Suite{RootDir: path, SourceDir: path})
		}
		list = append(list, readIndices(fs.JoinPath(path, IndexFile))...)

		// Check build directories
		for _, file := range fs.Glob(path + "/*build*/" + IndexFile) {
			list = append(list, readIndices(file)...)
		}
		for _, file := range fs.Glob(path + "/build/native/*/sct/" + IndexFile) {
			list = append(list, readIndices(file)...)
		}
		return true
	})

	// If we could not find any manifest, try guess a root directory based on known naming schemes.
	if len(list) == 0 {
		fs.WalkUp(path, func(path string) bool {
			if tests := fs.Glob(path + "/testcases/*"); len(tests) > 0 {
				log.Debugf("discovered testcases folder in %q\n", path)
				list = append(list, Suite{RootDir: path, SourceDir: path})
				return false
			}
			return true
		})
	}

	// Remove duplicate entries
	result := make([]Suite, 0, len(list))
	visited := make(map[Suite]bool)
	for _, v := range list {
		if !visited[v] {
			visited[v] = true
			result = append(result, v)
		}
	}
	return result
}

// Task is a build task.
type Task interface {
	// Inputs returns the list of input files.
	Inputs() []string

	// Outputs returns the list of output files.
	Outputs() []string

	// Run executes the task.
	Run() error

	// String returns a string representation of the task.
	String() string
}

// Build builds a project.
func Build(c *Config) error {
	since := time.Now()
	defer func() {
		log.Debugf("build took %s", time.Since(since))
	}()

	tasks, err := BuildTasks(c)
	if err != nil {
		return err
	}
	for _, tsk := range tasks {
		if err := tsk.Run(); err != nil {
			return err
		}
	}
	return nil
}

// BuildTasks returns the build tasks required to generate and build the test
// executable and its dependencies.
func BuildTasks(c *Config) ([]Task, error) {
	var (
		ret  []Task
		merr *multierror.Error
	)

	for _, imp := range c.Imports {
		tasks, err := ImportTasks(c, imp)
		if err != nil {
			merr = multierror.Append(merr, err)
			continue
		}
		ret = append(ret, tasks...)
	}

	srcs, err := fs.TTCN3Files(c.Sources...)
	if err != nil {
		merr = multierror.Append(merr, err)
	}

	var imports []string
	for _, t := range ret {
		for _, output := range t.Outputs() {
			if fs.HasTTCN3Extension(output) {
				imports = append(imports, output)
			}
		}
	}

	for _, t := range k3.NewT3XF(c.Variables, c.K3.T3XF, srcs, imports...) {
		ret = append(ret, t)
	}

	return ret, merr.ErrorOrNil()
}

// ImportTasks returns the build tasks required to generate and build a given
// test suite dependency.
func ImportTasks(c *Config, uri string) ([]Task, error) {
	dir := fs.Path(uri)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var (
		tasks                         []Task
		asn1Files, ttcn3Files, cFiles []string
		processed                     int
	)

	// Collect and categorize all the files!
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		switch path := filepath.Join(dir, f.Name()); filepath.Ext(path) {
		case ".asn1", ".asn":
			asn1Files = append(asn1Files, path)
			processed++
		case ".ttcn3", ".ttcn":
			ttcn3Files = append(ttcn3Files, path)
			processed++
		case ".c", ".cxx", ".cpp", ".cc":
			// Skip ASN1 codecs
			if strings.HasSuffix(path, ".enc.c") {
				continue
			}
			cFiles = append(cFiles, path)
			processed++
		}
	}
	if processed == 0 {
		return nil, fmt.Errorf("%s: %w", dir, ErrNoSources)
	}

	name, err := NameFromURI(dir)
	if err != nil {
		return nil, err
	}

	if len(asn1Files) > 0 {
		encoding, err := EncodingFromURI(dir)
		if err != nil {
			return nil, err
		}
		for _, t := range k3.NewASN1Codec(c.Variables, name, encoding, asn1Files...) {
			tasks = append(tasks, t)
		}
	}
	if len(cFiles) > 0 {
		for _, t := range k3.NewPlugin(c.Variables, name, cFiles...) {
			tasks = append(tasks, t)
		}
	}
	if len(ttcn3Files) > 0 {
		for _, t := range k3.NewTTCN3Library(c.Variables, name, ttcn3Files...) {
			tasks = append(tasks, t)
		}
	}
	return tasks, nil
}

// NameFromURI derives a TTCN-3 compatible name from a path or URI.
func NameFromURI(uri string) (string, error) {
	uri = fs.Path(uri)
	if strings.HasPrefix(uri, ".") {
		abs, err := filepath.Abs(uri)
		if err != nil {
			return "", fmt.Errorf("%s: %w", uri, err)
		}
		uri = abs
	}
	return fs.Slugify(fs.Stem(uri)), nil
}

// Encoding returns the ASN.1 encoding for a given URI. Current implementation
// always returns "per", unless the URI contains the string "rrc"
func EncodingFromURI(uri string) (string, error) {
	name, err := NameFromURI(uri)
	if err != nil {
		return "", err
	}
	if strings.Contains(strings.ToLower(name), "rrc") {
		return "uper", nil
	}
	return "per", nil
}

// Files returns the list of TTCN-3 source files. Genereated files are
// excluded for now, but might be added in the future.
func Files(c *Config) ([]string, error) {
	var files []string
	files = append(files, c.Sources...)
	files = append(files, c.Imports...)
	files = append(files, c.K3.Includes...)
	return fs.TTCN3Files(files...)
}

// ApplyPresets returns a list of test case configurations with optional
// presets applied. The presets are applied in the order they are specified in
// the list.
//
// ApplyPresets reads test case configuration from environment variables, the
// parameters file, package.yml and from the TTCN-3 documentation tags.
func ApplyPresets(c *Config, presets ...string) (*Parameters, error) {

	// Global configuration
	gc := c.Parameters

	// Parameters file overrides/extends global test configuration
	if c.ParametersFile != "" {
		var pf Parameters
		b, err := fs.Content(c.ParametersFile)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(b, &pf); err != nil {
			return nil, err
		}
		gc = MergeParameters(gc, pf)
	}

	// Presets override/extend parameters files
	for _, preset := range presets {
		tc, ok := gc.Presets[preset]
		if !ok {
			return nil, fmt.Errorf("preset %q not found", preset)
		}
		gc.TestConfig = MergeTestConfig(gc.TestConfig, tc)
	}

	files, err := fs.TTCN3Files(c.Sources...)
	if err != nil {
		return nil, err
	}

	list := AcquireExecutables(&gc, files, presets)

	gc.Execute = list
	return &gc, nil
}

// AcquireExecutables depending on the provided presets and on the availability
// inside the ttcn-3 code, a list of executable ttcn-3 entities (i.e. testcases,
// control parts) is returned
func AcquireExecutables(gc *Parameters, files []string, presets []string) []TestConfig {
	var list []TestConfig
	add := func(name string, comments string) {
		// TODO(5nord) make this less quadratic.
		for _, tc := range gc.Execute {
			pattern, params := SplitTest(tc.Test)
			ok, err := filepath.Match(pattern, name)
			if err != nil {
				log.Verbosef("%s: %s\n", name, err.Error())
			}
			if ok {
				tc = MergeTestConfig(gc.TestConfig, tc)
				tc.Test = name
				if params != "" {
					tc.Test += "(" + params + ")"
				}
				if DoesTestcaseMatchPreset(tc, presets) {
					list = append(list, tc)
				}
			}
		}
	}

	for _, file := range files {
		tree := ttcn3.ParseFile(file)
		ast.Inspect(tree.Root, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.FuncDecl:
				if n.IsTest() {
					add(tree.QualifiedName(n), ast.FirstToken(n).Comments())
				}
				return false
			case *ast.ControlPart:
				add(tree.QualifiedName(n), ast.FirstToken(n).Comments())
				return false
			}
			return true
		})
	}
	return list
}

// SplitTest splits a test into its testcase name and parameters.
func SplitTest(name string) (string, string) {
	if i := strings.Index(name, "("); i > 0 {
		return name[:i], name[i+1 : len(name)-1]
	}
	return name, ""
}

// DoesTestcaseMatchPreset: check whether testcase instance shall
// be executed dependent on the specified presets
func DoesTestcaseMatchPreset(tc TestConfig, presets []string) bool {
	ret := false
	for _, p := range presets {
		if tc.Only != nil {
			if len(tc.Only.Presets) > 0 {
				for _, presetFromFile := range tc.Only.Presets {
					if presetFromFile == p {
						ret = true
						break
					}
				}
			}
		}
		if ret {
			break
		}
	}

	for _, p := range presets {
		if tc.Except != nil {
			if len(tc.Except.Presets) > 0 {
				for _, presetFromFile := range tc.Except.Presets {
					if presetFromFile != p {
						ret = true
						break
					}
				}
			}
		}
		if !ret {
			break
		}
	}
	if len(presets) == 0 && tc.Except != nil && len(tc.Except.Presets) > 0 {
		ret = true
	}
	return ret
}

// MergeParameters merges the given parameters. Scalar values from b override
// values from a. Maps are merged and arrays are appended.
func MergeParameters(a, b Parameters) Parameters {
	result := Parameters{}
	result.TestConfig = MergeTestConfig(a.TestConfig, b.TestConfig)
	result.Presets = make(map[string]TestConfig)
	for k, v := range a.Presets {
		result.Presets[k] = v
	}
	for k, v := range b.Presets {
		result.Presets[k] = MergeTestConfig(result.Presets[k], v)
	}
	if len(result.Presets) == 0 {
		result.Presets = nil
	}
	result.Execute = append(a.Execute, b.Execute...)
	return result
}

// MergeTestConfig merges to test configurations. Scalar values from b override
// values from a. Maps are merged. Arrays are appended.
func MergeTestConfig(a, b TestConfig) TestConfig {
	result := TestConfig{}
	result.Test = a.Test
	if b.Test != "" {
		result.Test = b.Test
	}
	result.Timeout = a.Timeout
	if b.Timeout.Duration > 0 {
		result.Timeout = b.Timeout
	}
	result.Preset = append(a.Preset, b.Preset...)
	result.Parameters = make(map[string]string)
	for k, v := range a.Parameters {
		result.Parameters[k] = v
	}
	for k, v := range b.Parameters {
		result.Parameters[k] = v
	}
	if len(result.Parameters) == 0 {
		result.Parameters = nil
	}
	// Should we return an error if a and b have conflicting execute conditions?
	result.Only = b.Only
	result.Except = b.Except
	return result
}

// Open returns the best possible configuration using the given arguments.
//
// Without any arguments Open will open the current working directory, unless
// environment variable NTT_SOURCE_DIR is set.
//
// If you pass a manifest file as single argument, Open will use it directly.
//
// If you pass a directory as single argument, Open will first look for a
// package.yml and use it.
//
// If no package.yml is found Open will look for TTCN-3 source files. It will
// look recursively if directory contains typical project root files (i.e.
// build.sh, testcases-folder, ...). Open will also recursively load TTCN-3
// source files from typical import-directories (i.e ../common).
func Open(args ...string) (*Config, error) {
	defaults := ConfigOptions(
		AutomaticEnv(),
		WithIndex(cache.Lookup(IndexFile)),
		WithDefaults(),
		WithK3(),
	)

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Use default arguments if none are given.
	if len(args) == 0 {
		args = []string{cwd}
		if source_dir := env.Getenv("NTT_SOURCE_DIR"); source_dir != "" {
			args = []string{source_dir}
		}
	}

	// Multiple arguments are always treated as sources.
	if len(args) > 1 {
		return NewConfig(WithSources(args...), defaults)
	}

	// Treat a single file argument as source, unless it is a directory or
	// manifest file.
	if file := args[0]; fs.IsRegular(file) {
		if filepath.Base(file) == ManifestFile {
			return NewConfig(WithManifest(file), defaults)
		}
		return NewConfig(WithSources(file), defaults)
	}

	return NewConfig(AutomaticRoot(args[0]), defaults)
}

func NewConfig(opts ...ConfigOption) (*Config, error) {
	c := &Config{}
	return c, ConfigOptions(opts...)(c)
}

type ConfigOption func(*Config) error

// ConfigOptions returns an options, which applies the given configuration options.
func ConfigOptions(opts ...ConfigOption) ConfigOption {
	return func(c *Config) error {
		var gerr *multierror.Error
		for _, opt := range opts {
			if err := opt(c); err != nil {
				gerr = multierror.Append(gerr, err)
			}
		}
		return gerr.ErrorOrNil()
	}
}

// WithIndex uses hints from given index file to configure source directories.
func WithIndex(file string) ConfigOption {
	return func(c *Config) error {
		if c.Root == "" {
			return nil
		}
		b, err := fs.Content(file)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		var idx Index
		if len(b) > 0 {
			if err := json.Unmarshal(b, &idx); err != nil {
				return fmt.Errorf("%s: %w", file, err)
			}
		}

		base := filepath.Dir(file)
		info, err := os.Stat(c.Root)
		if err != nil {
			return fmt.Errorf("%s: %w", c.Root, err)
		}
		for _, s := range idx.Suites {
			if s.SourceDir != "" && s.RootDir != "" {
				root := fs.Real(base, s.RootDir)
				info2, err := os.Stat(root)
				if err != nil {
					return fmt.Errorf("%s: %w", root, err)
				}
				if os.SameFile(info, info2) {
					log.Debugf("project: using source dir %s\n", s.SourceDir)
					c.SourceDir = fs.Real(base, s.SourceDir)
				}
			}
		}
		return nil
	}
}

// WithManifest reads a manifest file (package.yml).
//
// expand variable references recursively. Environment variables overwrite
// variables defined in environment files. Variables defined in environment
// files overwrite variables defined in the manifest. Undefined variables won't
// be expanded and will return an error.
func WithManifest(file string) ConfigOption {
	return func(c *Config) error {
		c.ManifestFile = file
		c.Root = filepath.Dir(file)
		b, err := fs.Content(file)
		if err != nil {
			return err
		}
		if len(b) > 0 {
			if err := yaml.Unmarshal(b, &c.Manifest); err != nil {
				return fmt.Errorf("%s: %w", file, err)
			}
		}
		c.updateVariables()
		if err := c.Variables.Expand(); err != nil {
			return err
		}
		env.ExpandAll(&c.Manifest, c.Variables)
		c.Manifest.expandPaths(c.Root)
		log.Debugf("project: using manifest %s\n", file)
		return nil
	}
}

func WithK3() ConfigOption {
	return func(c *Config) error {
		c.toolchain = "k3"
		c.K3.Compiler = k3.Compiler()
		c.K3.Runtime = k3.Runtime()
		c.K3.Includes = k3.Includes()
		c.K3.Plugins = k3.Plugins()
		c.K3.OssInfo = filepath.Join(k3.DataDir(), "asn1")
		c.K3.T3XF = cache.Lookup(fmt.Sprintf("%s.t3xf", c.Name))
		log.Debugf("project: k3 compiler : %v\n", c.K3.Compiler)
		log.Debugf("project: k3 runtime  : %v\n", c.K3.Runtime)
		log.Debugf("project: k3 t3xf     : %v\n", c.K3.T3XF)
		log.Debugf("project: k3 includes : %v\n", c.K3.Includes)
		log.Debugf("project: k3 plugins  : %v\n", c.K3.Plugins)
		return nil
	}
}

// WithRoot sets the root directory of the project.
func WithRoot(root string) ConfigOption {
	return func(c *Config) error {
		c.Root = root
		return nil
	}
}

// WithSourceDir sets the source directory of the project.
func WithSourceDir(dir string) ConfigOption {
	return func(c *Config) error {
		c.SourceDir = dir
		return nil
	}
}

// WithSources sets the sources of the project.
func WithSources(srcs ...string) ConfigOption {
	return func(c *Config) error {
		c.Sources = srcs
		return nil
	}
}

// WithImports sets the imports of the project.
func WithImports(dirs ...string) ConfigOption {
	return func(c *Config) error {
		c.Imports = dirs
		return nil
	}
}

// AutomaticRoot sets the root directory and automatically sets the source and
// imports directories.
//
// If the root directory contains a package.yml, sources and imports are set
// from the manifest exclusively.
//
// AutomaticRoot will load TTCN-3 source files recursively, if the root
// directory contains typical project root files (e.g. build.sh, testcases/,
// ...).
//
// Additionally AutomaticRoot finds common import directories relative to the root folder
// (i.e ../library or ../common) and recursively adds all folders containing
// TTCN-3 source files.
func AutomaticRoot(root string) ConfigOption {
	return func(c *Config) error {
		c.Root = root
		log.Debugf("project: root %s\n", root)
		if manifest := fs.JoinPath(root, ManifestFile); fs.IsRegular(manifest) {
			return WithManifest(manifest)(c)
		}

		if isRoot(c.Root) {
			log.Debugln("project: scanning recursively...")
			c.Sources = fs.FindTTCN3FilesRecursive(c.Root)
		} else {
			c.Sources = fs.FindTTCN3Files(c.Root)
		}
		s := fmt.Sprintf("%v", c.Sources)
		if len(s) > 200 {
			s = s[:200] + fmt.Sprintf("...] (%d files/directories)", len(c.Sources))
		}
		log.Debugf("project: use sources: %s\n", s)

		commonDirs := []string{
			"../../../sct",
			"../../../../sct",
			"../common",
			"../Common",
			"../library",
			"../../../../Common",
		}
		for _, dir := range commonDirs {
			path := fs.JoinPath(c.Root, dir)
			if eval, err := filepath.EvalSymlinks(path); err == nil {
				path = eval
			}

			c.Imports = append(c.Imports, fs.FindTTCN3DirectoriesRecursive(path)...)
		}
		s = fmt.Sprintf("%v", c.Imports)
		if len(s) > 200 {
			s = s[:200] + fmt.Sprintf("...] (%d directories)", len(c.Imports))
		}
		log.Debugf("project: use imports: %s\n", s)

		return nil
	}
}

// AutomaticEnv let environment variables with NTT_ prefix overwrite
// configuration. Currently supported are: NTT_NAME, NTT_SOURCES, NTT_IMPORTS,
// NTT_PARAMETERS_FILE, NTT_HOOKS_FILE, NTT_LINT_FILE, NTT_TIMEOUT.
func AutomaticEnv() ConfigOption {
	return func(c *Config) error {
		for k, v := range env.EnvironMap() {
			used := true
			switch k {
			case "NTT_NAME":
				c.Name = v
			case "NTT_SOURCES":
				c.Sources = strings.Split(v, string(os.PathListSeparator))
			case "NTT_IMPORTS":
				c.Imports = strings.Split(v, string(os.PathListSeparator))
			case "NTT_PARAMETERS_FILE":
				c.ParametersFile = v
			case "NTT_HOOKS_FILE":
				c.HooksFile = v
			case "NTT_LINT_FILE":
				c.LintFile = v
			case "NTT_TIMEOUT":
				if err := c.Timeout.UnmarshalText([]byte(v)); err != nil {
					return fmt.Errorf("environment variable %s: %w", k, err)
				}
			default:
				used = false
			}
			if used {
				log.Debugf("project: using environment variable %s=%s\n", k, v)
			}
		}
		return nil
	}
}

// WithDefaults initializes Root, SourceDir, Name, ParametersFile and HooksFile.
func WithDefaults() ConfigOption {
	return func(c *Config) error {
		if c.Root == "" {
			switch {
			case c.SourceDir != "":
				c.Root = c.SourceDir
			case len(c.Sources) > 0:
				c.Root = filepath.Dir(c.Sources[0])
				// When there's no root, but only source, we want the suite to be named after the source.
				n, err := NameFromURI(c.Sources[0])
				if err != nil {
					return err
				}
				c.Name = n
			default:
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				c.Root = cwd
			}
			log.Debugf("project: using default root %s\n", c.Root)
		}
		if c.SourceDir == "" {
			c.SourceDir = c.Root
			log.Debugf("project: using default source dir %s\n", c.SourceDir)
		}
		if c.Name == "" {
			n, err := NameFromURI(c.Root)
			if err != nil {
				return err
			}
			c.Name = n
			log.Debugf("project: using default name %s\n", c.Name)
		}
		defaultFile := func(name string) string {
			if path := cache.Lookup(name); fs.IsRegular(path) {
				return path
			}
			if path := fs.JoinPath(c.Root, name); fs.IsRegular(path) {
				return path
			}
			if path := fs.JoinPath(c.SourceDir, name); fs.IsRegular(path) {
				return path
			}
			return ""
		}
		if c.ParametersFile == "" {
			if path := defaultFile(fmt.Sprintf("%s.parameters", c.Name)); path != "" {
				c.ParametersFile = path
				log.Debugf("project: using parameters file %s\n", c.ParametersFile)
			}
		}
		if c.HooksFile == "" {
			if path := defaultFile(fmt.Sprintf("%s.hooks", c.Name)); path != "" {
				c.HooksFile = path
				log.Debugf("project: using hooks file %s\n", c.HooksFile)
			}
		}
		if c.LintFile == "" {
			if path := defaultFile("ntt-lint.yml"); path != "" {
				c.LintFile = path
				log.Debugf("project: using lint file %s\n", c.LintFile)
			}
		}

		c.updateVariables()
		return nil
	}
}

// isRoot returns true if the given path contains typical project root files.
func isRoot(root string) bool {
	return fs.IsRegular(fs.JoinPath(root, ManifestFile)) ||
		fs.IsRegular(fs.JoinPath(root, "build.sh")) ||
		fs.IsRegular(fs.JoinPath(root, "project.xml")) ||
		fs.IsDir(fs.JoinPath(root, "testcases")) ||
		len(fs.Glob(fs.JoinPath(root, "*.cfg"))) > 0 ||
		len(fs.Glob(fs.JoinPath(root, "*.parameters"))) > 0
}

// updateVariables updates the given variable with the variables from
// environment files. Environment variables override environment files.
// Environment files overwrite manifest variables.
func (m *Manifest) updateVariables() {
	if m.Variables == nil {
		m.Variables = make(map[string]string)
	}
	for k, v := range env.ParseFiles() {
		if s, ok := env.LookupEnv(k); ok {
			v = s
		}
		m.Variables[k] = v
	}
	for k, v := range k3.DefaultEnv {
		if _, ok := m.Variables[k]; !ok {
			m.Variables[k] = v
		}
	}
	if len(m.Variables) == 0 {
		m.Variables = nil
	}
}

func (m *Manifest) expandPaths(base string) {
	for i, src := range m.Sources {
		m.Sources[i] = fs.Real(base, src)
	}
	for i, imp := range m.Imports {
		m.Imports[i] = fs.Real(base, imp)
	}
	m.HooksFile = fs.Real(base, m.HooksFile)
	m.ParametersFile = fs.Real(base, m.ParametersFile)
	m.LintFile = fs.Real(base, m.LintFile)
}
