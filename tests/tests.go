// Package test provides interfaces for running TTCN-3 test cases.
package tests

import (
	"errors"
	"time"

	"github.com/nokia/ntt/project"
)

// Job describes the test or control function to be executed.
type Job struct {
	// Name is the fully qualified name of the test or control function
	Name string

	// Args is the list of arguments to pass to the test.
	Args []string

	// Timeout is the duration after which the test will be stopped.
	Timeout time.Duration

	// Module Parameters
	ModulePars map[string]string

	// Env specifies the environment variables to pass to the test.
	Env []string

	// Config provides the project configuration
	*project.Config
}

// JobError describes an error that occurred during the execution of a test or control function.
type JobError struct {
	*Job
	Err error
}

func (e JobError) Error() string {
	return e.Err.Error()
}
func (e JobError) Unwrap() error {
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
	event
	*Job
	Name string
}

// NewStartEvent creates a new StartEvent.
func NewStartEvent(job *Job, name string) StartEvent {
	return StartEvent{event{t: time.Now()}, job, name}
}

// StopEvent is an event that is emitted when the test is stopped.
type StopEvent struct {
	event
	*Job
	Name    string
	Verdict string
}

// NewStopEvent creates a new StopEvent.
func NewStopEvent(job *Job, name string, verdict string) StopEvent {
	return StopEvent{event{t: time.Now()}, job, name, verdict}
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

// IsFail returns true if the event is an error event or has an non-pass verdict.
func IsFail(e Event) bool {
	switch e := e.(type) {
	case ErrorEvent:
		return true
	case StopEvent:
		return e.Verdict != "pass"
	}
	return false
}
