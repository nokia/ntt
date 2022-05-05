package lsp

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/lsp/protocol"
	"github.com/nokia/ntt/k3/k3s"
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/tests"
)

type TestController struct {
	// messages is the channel where test events are sent.
	messages chan tests.Event

	running map[key]bool
	mu      sync.Mutex

	// client is the LSP client. It is used to send notifications and
	// status updates to the IDE.
	client protocol.Client
}

type key struct {
	name   string
	config *project.Config
}

func (c *TestController) Start(client protocol.Client) error {
	c.messages = make(chan tests.Event)
	c.running = make(map[key]bool)
	c.client = client
	go c.handleEvents()
	return nil
}

func (c *TestController) Shutdown() error {
	close(c.messages)
	return nil
}

func (c *TestController) handleEvents() {
	for event := range c.messages {
		log.Debugf("TestController: %+v", event)
		c.mu.Lock()
		switch e := event.(type) {
		case tests.StartEvent:
		case tests.StopEvent, tests.ErrorEvent:
			if job := tests.UnwrapJob(e); job != nil {
				delete(c.running, key{job.Name, job.Config})
			}
		default:
			log.Debugf("TestController: unknown event")
		}
		c.mu.Unlock()
		c.client.CodeLensRefresh(context.Background())
	}
}

func (c *TestController) IsRunning(config *project.Config, name string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.running[key{name, config}]
}

func (c *TestController) RunTest(config *project.Config, name string, logger io.Writer) error {
	if c.IsRunning(config, name) {
		return fmt.Errorf("test %s already running", name)
	}

	job := &tests.Job{
		Name:   name,
		Config: config,
	}

	c.mu.Lock()
	c.running[key{name, config}] = true
	c.mu.Unlock()
	c.messages <- tests.NewStartEvent(job, name)

	go func() {
		fmt.Fprintf(logger, `
===============================================================================
Compiling test %s in %q`, name, config.Root)

		if err := k3s.Build(logger, config); err != nil {
			fmt.Fprintln(logger, err.Error())
			c.messages <- tests.NewErrorEvent(&tests.JobError{Job: job, Err: err})
			return
		}

		fmt.Fprintf(logger, `
===============================================================================
Running test %s in %q`, name, config.Root)

		logDir, _ := k3s.Run(logger, config, name)

		if files := fs.Abs(fs.FindFilesRecursive(logDir)...); len(files) > 0 {
			fmt.Fprintf(logger, `
Content of log directory %q:
===============================================================================
%s
`,
				logDir, strings.Join(files, "\n"))
		}
		c.messages <- tests.NewStopEvent(job, name, "")
	}()

	return nil
}
