// Package control provides the test control and management interface.
package control

import (
	"time"

	"github.com/nokia/ntt/project"
)

// Job describes the test or control function to be executed.
type Job struct {
	// ID is the unique identifier of the job.
	ID string

	// Name is the fully qualified name of the test or control function.
	Name string

	// Args is the list of arguments to pass to the test.
	Args []string

	// Timeout is the duration after which the job will be stopped.
	Timeout time.Duration

	// Module Parameters
	ModulePars map[string]string

	// Dir specifies the working directory for the job.
	Dir string

	// Env specifies the environment variables to pass to the job.
	Env []string

	// Config provides the project configuration
	*project.Config
}

func NewJob(name string, conf *project.Config) *Job {
	return &Job{
		Name:   name,
		Config: conf,
	}
}

// JobError describes an error that occurred during the execution of a test or control function.
type JobError struct {
	*Job
	Err error
}

func (e *JobError) Error() string {
	return e.Err.Error()
}

func (e *JobError) Unwrap() error {
	return e.Err
}
