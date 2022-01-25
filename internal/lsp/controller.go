package lsp

import (
	"fmt"
	"io"
	"strings"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/runner/k3s"
)

type TestController struct {
	events chan Event
}

type Event struct {
	Type string
}

func (c *TestController) Start() error {
	c.events = make(chan Event)
	go c.handleEvents()
	return nil
}

func (c *TestController) Shutdown() error {
	close(c.events)
	return nil
}

func (c *TestController) handleEvents() {
	for event := range c.events {
		switch event.Type {
		case "tcst":
			log.Println("tcst")
		case "tcfi":
			log.Println("tcfi")
		case "error":
			log.Println("error")
		case "log":
			log.Println("log")
		}
	}
}

func (c *TestController) IsRunning(p project.Interface, name string) bool {
	return false
}

func (c *TestController) RunTest(p project.Interface, name string, logger io.Writer) error {
	fmt.Fprintf(logger, `
===============================================================================
Compiling test %s in %q`, name, p.Root())

	r, err := k3s.New(logger, p)
	if err != nil {
		fmt.Fprintln(logger, err.Error())
		return err
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

	return nil
}
