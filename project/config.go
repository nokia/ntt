package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/fs"
)

var (
	ErrNoSuchName = errors.New("no such name")
	ErrConversion = errors.New("conversion error")
)

const (
	ManifestFile = "package.yml"
	IndexFile    = "ttcn3_suites.json"
)

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

// Config describes everything required to build and run test suites
type Config struct {
	// Root folder of the project.
	Root string `json:"root,omitempty"`

	// Source folder of the project.
	SourceDir string `json:"source_dir,omitempty"`

	// Manifest is the project manifest file.
	Manifest `json:,inline"`

	// K3 specific configuration.
	K3 `json:"k3,omitempty"`

	// Path to the environment file.
	EnvFile string `json:"env_file,omitempty"`
}

// Manifest describe the project manifest file (package.yml).
type Manifest struct {
	// Name of the project.
	Name string `json:"name,omitempty"`

	// Sources is the TTCN-3 test case source files.
	Sources []string `json:"sources,omitempty"`

	// Imports are the dependencies of the project (e.g. adapters, codecs, etc.).
	Imports []string `json:"imports,omitempty"`

	// Variables are project specific environment variables.
	Variables map[string]string `json:"variables,omitempty"`

	// Hooks are the commands that are executed by ntt to signal various events.
	Hooks `json:",inline" yaml:",inline"`

	// HooksFile is the path to an executable file that is executed by ntt to signal various events.
	HooksFile string `json:"test_hook,omitempty"`

	// Parameters are additional test suite parameters.
	Parameters `json:"parameters" yaml:",inline"`

	// ParametersFile is the path for additional test suite parameters.
	ParametersFile string `json:"parameters_file,omitempty" yaml:"parameters_file"`

	// ParametersDir is an additional search path for the parameters file.
	ParametersDir string `json:"parameters_dir,omitempty" yaml:"parameters_dir"`

	// Linter configuration.
	Lint `json:"lint,omitempty"`

	// Path to the an external linter configuration file.
	LintFile string `json:"lint_file,omitempty" yaml:"lint_file"` // Path for lint file.
}

// Hooks describes various hooks that can be executed by ntt. When a hook
// returns an error ntt will stop the execution of the test suite.
type Hooks struct {
	// BeforeBuild is executed before the build process.
	BeforeBuild []string `json:"before_build,omitempty" yaml:"before_build"`

	// AfterBuild is executed after the build process.
	AfterBuild []string `json:"after_build,omitempty" yaml:"after_build"`

	// BeforeRun is executed before any tests are executed.
	BeforeRun []string `json:"before_run,omitempty" yaml:"before_run"`

	// AfterRun is executed after all tests have been executed.
	AfterRun []string `json:"after_run,omitempty" yaml:"after_run"`

	// BeforeTest is executed before each test.
	BeforeTest []string `json:"before_test,omitempty" yaml:"before_test"`

	// AfterTest is executed after each test.
	AfterTest []string `json:"after_test,omitempty" yaml:"after_test"`
}

// Parameters describes the parameters file.
type Parameters struct {
	// Global test configuration.
	TestConfig `yaml:",inline"`

	// Testconfiguration presets providing alternative global configurations.
	Presets map[string]TestConfig `json:"presets,omitempty"`

	// Test specific configuration. Each entry is a test configuration and
	// specifies how and when a test should be executed.
	Execute []TestConfig `json:"execute,omitempty"`
}

// A TestConfig specifies how and when a testcase should be executed.
type TestConfig struct {
	// Fully qualified name. Optional testcase parameters are allowed.
	Test string `json:"test,omitempty"`

	// Timeout in seconds.
	Timeout float64 `json:"timeout,omitempty"`

	// Module parameters
	Parameters map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`

	// Only execute testcase if the given conditions are met.
	Only *ExecuteCondition `json:"only,omitempty"`

	// Do not execute testcase if the given conditions are met.
	Except *ExecuteCondition `json:"except,omitempty"`
}

// ExecuteCondition specifies conditions for executing a testcase.
type ExecuteCondition struct {
	Presets []string `json:"presets,omitempty"`
}

// Lint describes the lint configuration.
type Lint struct {

	// The maximum nummber of lines a function, test or altstep can have.
	MaxLines int `json:"max_lines,omitempty" yaml:"max_lines"`

	// AlignedBraces reports an error if braces are not aligned.
	AlignedBraces bool `json:"aligned_braces,omitempty" yaml:"aligned_braces"`

	// RequireCaseElse reports an error if a case statement does not have an else branch.
	RequireCaseElse bool `json:"require_case_else,omitempty" yaml:"require_case_else"`

	// Complexity configures the cyclomatic complexity settings.
	Complexity struct {
		// The maximum allowed cyclomatic complexity of a function.
		Max int `json:"max,omitempty" yaml:"max"`

		// IgnoreGuards ignores the cyclomatic complexity of alt guards.
		IgnoreGuards bool `json:"ignore_guards,omitempty" yaml:"ignore_guards"`
	} `json:"complexity,omitempty" yaml:"complexity"`

	// Naming configure the naming rules. Each key is a regular expression.
	// A leading ! negates the rule. The value is the error message. Example:
	//
	//     "^[A-Z0-9_]+$": "global constants must be UPPER_CASE"
	//
	Naming struct {
		Modules         map[string]string `json:"modules,omitempty" yaml:"modules"`
		Tests           map[string]string `json:"tests,omitempty" yaml:"tests"`
		Functions       map[string]string `json:"functions,omitempty" yaml:"functions"`
		Altsteps        map[string]string `json:"altsteps,omitempty" yaml:"altsteps"`
		Parameters      map[string]string `json:"parameters,omitempty" yaml:"parameters"`
		ComponentVars   map[string]string `json:"component_vars,omitempty" yaml:"component_vars"`
		VarTemplates    map[string]string `json:"var_templates,omitempty" yaml:"var_templates"`
		PortTypes       map[string]string `json:"port_types,omitempty" yaml:"port_types"`
		Ports           map[string]string `json:"ports,omitempty" yaml:"ports"`
		GlobalConsts    map[string]string `json:"global_consts,omitempty" yaml:"global_consts"`
		ComponentConsts map[string]string `json:"component_consts,omitempty" yaml:"component_consts"`
		Templates       map[string]string `json:"templates,omitempty" yaml:"templates"`
		Locals          map[string]string `json:"locals,omitempty" yaml:"locals"`
	} `json:"naming,omitempty" yaml:"naming"`

	// Tags configure the tags each test must have.
	Tags struct {
		Tests map[string]string `json:"tests,omitempty" yaml:"tests"`
	} `json:"tags,omitempty" yaml:"tags"`

	Ignore struct {
		// Modules to ignore.
		Modules []string `json:"modules,omitempty" yaml:"modules"`

		// File to ignore.
		Files []string `json:"files,omitempty" yaml:"files"`
	} `json:"ignore,omitempty" yaml:"ignore"`

	// Usage defines an upper limit for the number of times a symbol can be used.
	Usage map[string]*struct {
		// Text to display when the limit is reached.
		Text string `json:"text,omitempty" yaml:"text"`

		// Limit is the number of times a symbol can be used.
		Limit int `json:"limit,omitempty" yaml:"limit"`
		count int
	} `json:"usage,omitempty" yaml:"usage"`

	Unused struct {
		// Warn about unused modules.
		Modules bool `json:"modules,omitempty" yaml:"modules"`
	} `json:"unused,omitempty" yaml:"unused"`
}

// K3 specific configuration
type K3 struct {
	// A unique identifier for the test.
	SessionID string `json:"session_id"`

	// Path to the compiler.
	Compiler string `json:"compiler"`

	// Path to the runtime.
	Runtime string `json:"runtime"`

	// Path to additional TTCN-3 files
	Builtins []string `json:"builtins,omitempty"`

	// Path to OSS Nokalva installation.
	OssInfo string `json:"ossinfo,omitempty"`
}

// Get returns the configuration a given name. Environment variables with the
// capitalization of the name and NTT_ or K3_ prefix are also considered.
// Names are case-insensitive, underscores are removed. For example, "foo_bar" is the
// same as "FooBar".
//
// If the name is not be found or the conversion of environment variables failed, Get will return an error.
func (c *Config) Get(name string) (interface{}, error) {
	mapper := func(field string) bool {
		return strings.ToLower(field) == strings.ToLower(strings.ReplaceAll(name, "_", ""))
	}
	field := reflect.Indirect(reflect.ValueOf(c)).FieldByNameFunc(mapper)
	if s, ok := env.LookupEnv(fmt.Sprintf("NTT_%s", strings.ToUpper(name))); ok {
		return cast(s, field)
	}
	if field.IsValid() {
		return field.Interface(), nil
	}
	if s, ok := c.Variables[name]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("config: %w: %s", ErrNoSuchName, name)
}

func cast(s string, v reflect.Value) (interface{}, error) {
	if !v.IsValid() {
		return s, nil
	}

	switch v.Interface().(type) {
	case string:
		return s, nil
	case []string:
		return strings.Fields(s), nil
	case float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf("config: %w", err)
		}
		return f, nil
	case int:
		i, err := strconv.ParseInt(s, 0, 64)
		if err != nil {
			return nil, fmt.Errorf("config: %w", err)
		}
		return int(i), nil
	case bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return nil, fmt.Errorf("config: %w", err)
		}
		return b, nil
	}
	return s, nil
}

func ReadIndex(file string) (Index, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return Index{}, err
	}

	c := Index{}

	if err := json.Unmarshal(b, &c); err != nil {
		return Index{}, err
	}

	base := filepath.Dir(file)
	for i := range c.Suites {
		if c.Suites[i].RootDir != "" {
			c.Suites[i].RootDir = fs.Real(base, c.Suites[i].RootDir)
		}
		if c.Suites[i].SourceDir != "" {
			c.Suites[i].SourceDir = fs.Real(base, c.Suites[i].SourceDir)
		}
	}
	return c, nil
}
