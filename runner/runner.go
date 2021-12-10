package runner

import (
	"io"
)

type Runner interface {
	Run(w io.Writer, testID string) error
	LogDir(testID string) string
	Dir() string
}
