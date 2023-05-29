// Package control provides the test control and management interface.
package control

import (
	"context"
	"errors"
	"time"

	"github.com/nokia/ntt/project"
)

// A Runner runs one or multiple jobs and emits Events
type Runner interface {
	Run(context.Context) <-chan Event
}

// RunnerFactory creates a new Runner.
type RunnerFactory func() (Runner, error)

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

// Event provides information regarding test execution.
type Event interface {
	Time() time.Time // Time when the event happened.
}

// ErrorEvent is an error event.
type ErrorEvent struct {
	event
	Err error
}

func (e ErrorEvent) Error() string { return e.Err.Error() }
func (e ErrorEvent) Unwrap() error { return e.Err }

// NewError wraps an error into an ErrorEvent. It uses time.Now() as the time when the event happened.
func NewErrorEvent(err error) ErrorEvent {
	return ErrorEvent{event{t: time.Now()}, err}
}

// LogEvent is an event that provided additional information about the test execution.
type LogEvent struct {
	Text string
	event
	*Job
}

func NewLogEvent(job *Job, text string) LogEvent {
	return LogEvent{event: event{t: time.Now()}, Job: job, Text: text}
}

// StartEvent is an event that is emitted when the test is started.
type StartEvent struct {
	Name string
	event
	*Job
}

// NewStartEvent creates a new StartEvent.
func NewStartEvent(job *Job, name string) StartEvent {
	return StartEvent{event: event{t: time.Now()}, Job: job, Name: name}
}

// StopEvent is an event that is emitted when the test is stopped.
type StopEvent struct {
	Name    string
	Verdict string
	Begin   time.Time
	event
	*Job
}

// NewStopEvent creates a new StopEvent.
func NewStopEvent(job *Job, name string, verdict string) StopEvent {
	return StopEvent{event: event{t: time.Now()}, Job: job, Name: name, Verdict: verdict}
}

// TickerEvent is an event that is emitted periodically during the test execution.
type TickerEvent struct {
	event
	*Job
}

// NewTickerEvent creates a new TickerEvent.
func NewTickerEvent(job *Job) TickerEvent {
	return TickerEvent{event{t: time.Now()}, job}
}

// event is the base type for all events.
type event struct {
	t time.Time
}

func (e event) Time() time.Time { return e.t }

// UnwrapJob returns the job that caused the event or nil if no such job is available.
func UnwrapJob(e Event) *Job {
	switch e := e.(type) {
	case StartEvent:
		return e.Job
	case StopEvent:
		return e.Job
	case LogEvent:
		return e.Job
	case ErrorEvent:
		var err *JobError
		if errors.As(e.Err, &err) {
			return err.Job
		}
	}
	return nil
}
