package runner

import (
	"io"
)

// The runner interface is used to run a test.
type Runner interface {

	// Run the test.
	Run(w io.Writer, testID string) error

	// LogDir returns the directory where the test logs are stored.
	LogDir(testID string) string

	// Dir returns the working directory of the test artifacts.
	Dir() string
}
