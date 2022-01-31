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
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/runner/k3s"
)

type TestController struct {
	messages chan Message
	tests    map[pair]*Test
	testsMu  sync.Mutex
	client   protocol.Client
}

type Message struct {
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
	c.messages = make(chan Message)
	c.tests = make(map[pair]*Test)
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

	c.messages <- Message{Type: "tcst", Test: tst}

	go func() {
		fmt.Fprintf(logger, `
===============================================================================
Compiling test %s in %q`, name, p.Root())

		r, err := k3s.New(logger, p)
		if err != nil {
			fmt.Fprintln(logger, err.Error())
			c.messages <- Message{Type: "error", Test: tst}
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
%s\n\n`,
				logDir, strings.Join(files, "\n"))
		}

		c.messages <- Message{Type: "tcfi", Test: tst}
	}()

	return nil
}
