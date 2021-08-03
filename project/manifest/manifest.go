package manifest

const Name = "package.yml"

type Config struct {
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
