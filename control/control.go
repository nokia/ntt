// Package control provides the test control and management interface.
package control

import (
	"context"
)

// A Runner runs one or multiple jobs and emits Events
type Runner interface {
	Run(context.Context) <-chan Event
}

// RunnerFactory creates a new Runner.
type RunnerFactory func() (Runner, error)
