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
)

type TestController struct {
	events  chan Event
	tests   map[pair]*Test
	testsMu sync.Mutex
	client  protocol.Client
}

type Event struct {
	Type string
	Test *Test
}

type Test struct {
	pair
	logger io.Writer
	state  string
}

type pair struct {
	name string
	p    project.Interface
}

func (c *TestController) Start(client protocol.Client) error {
	c.events = make(chan Event)
	c.tests = make(map[pair]*Test)
	c.client = client
	go c.handleEvents()
	return nil
}

func (c *TestController) Shutdown() error {
	close(c.events)
	return nil
}

func (c *TestController) handleEvents() {
	for event := range c.events {
		log.Debugf("TestController: %+v", event)
		switch event.Type {
		case "tcst":
			c.testsMu.Lock()
			c.tests[event.Test.pair] = event.Test
			c.testsMu.Unlock()
			c.client.CodeLensRefresh(context.Background())

		case "tcfi", "error":
			c.testsMu.Lock()
			delete(c.tests, event.Test.pair)
			c.testsMu.Unlock()
			c.client.CodeLensRefresh(context.Background())
		}
	}
}

func (c *TestController) IsRunning(p project.Interface, name string) bool {
	c.testsMu.Lock()
	defer c.testsMu.Unlock()
	_, ok := c.tests[pair{name, p}]
	return ok
}

func (c *TestController) RunTest(p project.Interface, name string, logger io.Writer) error {

	if c.IsRunning(p, name) {
		return fmt.Errorf("test %s already running", name)
	}

	tst := &Test{
		state:  "pending",
		pair:   pair{name, p},
		logger: logger,
	}

	c.events <- Event{Type: "tcst", Test: tst}

	go func() {
		fmt.Fprintf(logger, `
===============================================================================
Compiling test %s in %q`, name, p.Root())

		r, err := k3s.New(logger, p)
		if err != nil {
			fmt.Fprintln(logger, err.Error())
			c.events <- Event{Type: "error", Test: tst}
			return
		}

		fmt.Fprintf(logger, `
===============================================================================
Running test %s in %q`, name, p.Root())

		err = r.Run(logger, name)

		// Show a directory listing of the artifacts (independently of any test errors)
		logDir := r.LogDir(name)
		if files := fs.Abs(fs.FindFilesRecursive(logDir)...); len(files) > 0 {
			fmt.Fprintf(logger, `
Content of log directory %q:
===============================================================================
%s
`,
				logDir, strings.Join(files, "\n"))
		}

		c.events <- Event{Type: "tcfi", Test: tst}
	}()

	return nil
}
