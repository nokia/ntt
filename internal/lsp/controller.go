package lsp

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/k3/k3s"
	"github.com/nokia/ntt/tests"
)

type TestController struct {
	// messages is the channel where test events are sent.
	messages chan tests.Event

	// jobs maps id and test jobs.
	//
	// It assures that a test is not executed twice, because the CodeLens
	// UI can only display one "run test" button at a time.
	jobs map[TestID]*tests.Job
	mu   sync.Mutex

	// client is the LSP client. It is used to send notifications and
	// status updates to the IDE.
	client protocol.Client

	// console is used to display test output.
	console io.Writer

	// suites is required to find test suite configurations for a given test.
	suites *Suites
}

type TestID struct {
	URI  string
	Name string
	Pos  loc.Pos
}

// Start starts the test controller.
func (c *TestController) Start(client protocol.Client, logger io.Writer, suites *Suites) error {
	c.messages = make(chan tests.Event)
	c.jobs = make(map[TestID]*tests.Job)
	c.client = client
	c.console = logger
	c.suites = suites
	go func() {
		for event := range c.messages {
			log.Debugf("TestController: %+v\n", event)
			switch e := event.(type) {
			case tests.StartEvent:
			case tests.StopEvent, tests.ErrorEvent:
				if job := tests.UnwrapJob(e); job != nil {
					c.removeJob(job)
				}
			default:
				log.Debugf("TestController: unknown event: %T\n", e)
			}
			c.client.CodeLensRefresh(context.Background())
		}
	}()
	return nil
}

// Shutdown stops the test controller.
func (c *TestController) Shutdown() error {
	close(c.messages)
	return nil
}

// IsRunning returns true if a test is already scheduled for execution.
func (c *TestController) IsRunning(id TestID) bool {
	return c.lookupJob(id) != nil
}

// RunTest schedules a test for execution.
func (c *TestController) RunTest(id TestID) error {

	// Check if the test is already running.
	if job := c.lookupJob(id); job != nil {
		return fmt.Errorf("test %s already running", job.Name)
	}

	// First we need determine the runtime configuration.
	//
	// If we find multiple configurations, we don't ask the user to choose,
	// but simply use the last candidate. The assumption is: The latest
	// configuration has a higher probability to be the complete one.
	candidates := c.suites.Owners(protocol.DocumentURI(id.URI))
	if len(candidates) == 0 {
		return fmt.Errorf("cannot run %s: no configuration found", id.URI)
	}
	if len(candidates) > 1 {
		log.Printf("multiple configurations found for %s: %v\n", id.URI, candidates)
	}
	config := candidates[0].Config
	log.Printf("using configuration from %s", config.Root)

	// TODO(5nord): Use project.ApplyPreset to retrieve the configuration,
	// like expected verdict for the job.
	c.enqueueJob(id, &tests.Job{
		Name:   id.Name,
		Config: config,
	})
	return nil
}

// job returns the test job for the given id.
func (c *TestController) lookupJob(id TestID) *tests.Job {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.jobs[id]
}

// remove removes the test job from the controller.
func (c *TestController) removeJob(job *tests.Job) {
	for id, j := range c.jobs {
		if j == job {
			c.mu.Lock()
			delete(c.jobs, id)
			c.mu.Unlock()
			return
		}
	}
}

func (c *TestController) enqueueJob(id TestID, job *tests.Job) {
	c.mu.Lock()
	c.jobs[id] = job
	c.mu.Unlock()
	c.messages <- tests.NewStartEvent(job, job.Name)

	go func() {
		fmt.Fprintf(c.console, `
===============================================================================
Compiling test %s in %q`, job.Name, job.Config.Root)

		if err := k3s.Build(c.console, job.Config); err != nil {
			fmt.Fprintln(c.console, err.Error())
			c.messages <- tests.NewErrorEvent(&tests.JobError{Job: job, Err: err})
			return
		}

		fmt.Fprintf(c.console, `
===============================================================================
Running test %s in %q`, job.Name, job.Config.Root)

		logDir, _ := k3s.Run(c.console, job.Config, job.Name)

		if files := fs.Abs(fs.FindFilesRecursive(logDir)...); len(files) > 0 {
			fmt.Fprintf(c.console, `
Content of log directory %q:
===============================================================================
%s
`,
				logDir, strings.Join(files, "\n"))
		}
		c.messages <- tests.NewStopEvent(job, job.Name, "")
	}()

}
