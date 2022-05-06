// Package test provides interfaces for running TTCN-3 test cases.
package tests

import (
	"errors"
	"time"

	"github.com/nokia/ntt/project"
)

// Job describes the test or control function to be executed.
type Job struct {
	// Name is the fully qualified name of the test or control function.
	Name string

	// Args is the list of arguments to pass to the test.
	Args []string

	// Timeout is the duration after which the job will be stopped.
	Timeout time.Duration

	// Module Parameters
	ModulePars map[string]string

	// Env specifies the environment variables to pass to the job.
	Env []string

	// Config provides the project configuration
	*project.Config
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
}

// NewTickerEvent creates a new TickerEvent.
func NewTickerEvent() TickerEvent {
	return TickerEvent{event{t: time.Now()}}
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
	case ErrorEvent:
		var err *JobError
		if errors.As(e.Err, &err) {
			return err.Job
		}
	}
	return nil
}

// IsPass returns true if the event is an StopEvent with a pass verdict. For
// all other verdicts or events types IsPass will return false.
func IsPass(e Event) bool {
	if se, ok := e.(StopEvent); ok {
		return se.Verdict == "pass"
	}
	return false
}

// Controller is responsible for running test cases. Essentially you put jobs
// in and events come out.
type Controller struct {
	events chan Event
}

// NewController creates a new Controller for executing tests using the given executor.
func (c *Controller) NewController(e Executor) *Controller {
	return nil
}

// EnqueueJob puts a job in the queue.
func (c *Controller) EnqueueJob(job *Job) error {
	return nil
}

// Events returns a channel that emits events.
func (c *Controller) Events() <-chan Event {
	return c.events
}

type Executor interface {
	// Execute runs the test job.
	Execute(*Job) <-chan Event
}
