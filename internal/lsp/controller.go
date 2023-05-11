package lsp

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/nokia/ntt/control"
	"github.com/nokia/ntt/control/k3s"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
)

type TestController struct {
	// running maps id and test running.
	//
	// It assures that a test is not executed twice, because the CodeLens
	// UI can only display one "run test" button at a time.
	running map[TestID]*control.Job
	mu      sync.Mutex

	// suites is required to find test suite configurations for a given test.
	suites *Suites

	// jobs is used to schedule test jobs.
	jobs chan *control.Job
}

type TestID struct {
	URI  string
	Name string
	Pos  int
}

// Start starts the test controller.
func (c *TestController) Start(client protocol.Client, logger io.Writer, suites *Suites) {
	c.suites = suites
	c.running = make(map[TestID]*control.Job)
	c.jobs = make(chan *control.Job, 1)
	go func() {
		runner := k3s.NewRunner(c.jobs, logger)
		for event := range runner.Run(context.Background()) {
			log.Debugf("TestController: %+v\n", event)
			switch e := event.(type) {
			case control.StartEvent:
			case control.StopEvent, control.ErrorEvent:
				if job := control.UnwrapJob(e); job != nil {
					c.removeJob(job)
				}
			default:
				log.Debugf("TestController: unknown event: %T\n", e)
			}
			client.CodeLensRefresh(context.Background())
		}
	}()
}

// Shutdown stops the test controller.
func (c *TestController) Shutdown() error {
	close(c.jobs)
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
	log.Printf("using configuration from %s\n", config.Root)

	// TODO(5nord): Use project.ApplyPreset to retrieve the configuration,
	// like expected verdict for the job.
	job := &control.Job{
		Name:   id.Name,
		Config: config,
	}

	c.mu.Lock()
	c.running[id] = job
	c.mu.Unlock()

	c.jobs <- job

	return nil
}

// job returns the test job for the given id.
func (c *TestController) lookupJob(id TestID) *control.Job {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.running[id]
}

// remove removes the test job from the controller.
func (c *TestController) removeJob(job *control.Job) {
	for id, j := range c.running {
		if j == job {
			c.mu.Lock()
			delete(c.running, id)
			c.mu.Unlock()
			return
		}
	}
}
