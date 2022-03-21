package project

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/nokia/ntt/internal/env"
)

var (
	ErrNoSuchName = errors.New("no such name")
	ErrConversion = errors.New("conversion error")
)

const ManifestFile = "package.yml"

// Config describes everything required to build and run test suites
type Config struct {
	// Root folder of the project.
	Root string `json:"root,omitempty" yaml:"-"`

	// Name of the project.
	Name string `json:"name,omitempty"`

	// Sources is the TTCN-3 test case source files.
	Sources []string `json:"sources,omitempty"`

	// Imports are the dependencies of the project (e.g. adapters, codecs, etc.).
	Imports []string `json:"imports,omitempty"`

	Variables map[string]string `json:"variables,omitempty"`

	HooksFile string `json:"test_hook,omitempty"`
	Hooks     `json:",inline" yaml:",inline"`

	ParametersFile string `json:"parameters_file,omitempty" yaml:"parameters_file"` // Path for module parameters file.
	ParametersDir  string `json:"parameters_dir,omitempty" yaml:"parameters_dir"`   // Optional path for parameters_file.
	Parameters     `json:"parameters" yaml:",inline"`

	LintFile string `json:"lint_file,omitempty" yaml:"lint_file"` // Path for lint file.
	Lint     `json:"lint,omitempty"`
	K3       `json:"k3,omitempty" yaml:"-"`

	EnvFile string `json:"env_file,omitempty" yaml:"-"`
}

type Hooks struct {
	BeforeBuild []string `json:"before_build,omitempty" yaml:"before_build"`
	AfterBuild  []string `json:"after_build,omitempty" yaml:"after_build"`
	BeforeRun   []string `json:"before_run,omitempty" yaml:"before_run"`
	AfterRun    []string `json:"after_run,omitempty" yaml:"after_run"`
	BeforeTest  []string `json:"before_test,omitempty" yaml:"before_test"`
	AfterTest   []string `json:"after_test,omitempty" yaml:"after_test"`
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

type Lint struct {
	MaxLines        int  `json:"max_lines,omitempty" yaml:"max_lines"`
	AlignedBraces   bool `json:"aligned_braces,omitempty" yaml:"aligned_braces"`
	RequireCaseElse bool `json:"require_case_else,omitempty" yaml:"require_case_else"`
	Complexity      struct {
		Max          int  `json:"max,omitempty" yaml:"max"`
		IgnoreGuards bool `json:"ignore_guards,omitempty" yaml:"ignore_guards"`
	} `json:"complexity,omitempty" yaml:"complexity"`
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
	Tags struct {
		Tests map[string]string `json:"tests,omitempty" yaml:"tests"`
	} `json:"tags,omitempty" yaml:"tags"`
	Ignore struct {
		Modules []string `json:"modules,omitempty" yaml:"modules"`
		Files   []string `json:"files,omitempty" yaml:"files"`
	} `json:"ignore,omitempty" yaml:"ignore"`
	Usage map[string]*struct {
		Text  string `json:"text,omitempty" yaml:"text"`
		Limit int    `json:"limit,omitempty" yaml:"limit"`
		count int
	} `json:"usage,omitempty" yaml:"usage"`
	Unused struct {
		Modules bool `json:"modules,omitempty" yaml:"modules"`
	} `json:"unused,omitempty" yaml:"unused"`
}

type K3 struct {
	SessionID string   `json:"session_id"`
	Compiler  string   `json:"compiler"`
	Runtime   string   `json:"runtime"`
	Builtins  []string `json:"builtins,omitempty"`
	OssInfo   string   `json:"ossinfo,omitempty"`
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
