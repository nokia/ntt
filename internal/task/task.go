/*
Package task provides a simplistic interface for executing make-like targets.

Note, this package was implemented to provide a smooth migration from Nokia
internal test tool (k3-build) to ntt and maybe replaced in the future.
*/
package task

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Task struct {
	// cmd to execut by Task.Run(). All variables in cmd are expanded, before
	// execute by /bin/sh
	cmd string

	// Input files. Expandable by variable SCRS.
	Inputs []string

	// Output files. Expandable by variable DESTS.
	Outputs []string

	// VPath specifies alternate directories for Output files.
	VPath []string

	// Map for expanding variables
	Vars map[string]string

	Always bool // Always execute task, skip timestamp checking

	Stdout bytes.Buffer // Stdout is reset before Task is executed.
	Stderr bytes.Buffer // Stderr is reset before Task is executed.
}

// NewTask returns a new Task object.
func NewTask(cmd string) *Task {
	return &Task{cmd: cmd}
}

// AddInput adds a new input file.
func (tsk *Task) AddInput(path ...string) {
	for _, v := range path {
		tsk.Inputs = append(tsk.Inputs, v)
	}
}

// AddOutput adds a new output file. Output path may be overwritten is file was
// found in a directory specified by Task.VPath.
func (tsk *Task) AddOutput(path ...string) {
	for _, v := range path {
		tsk.Outputs = append(tsk.Outputs, tsk.locateCachedFile(v))
	}
}

func (tsk *Task) locateCachedFile(path string) string {
	name := filepath.Base(path)
	for _, dir := range tsk.VPath {
		file := filepath.Join(dir, name)
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			return file
		}
	}
	return path
}

// Run starts the specified task and waits for it to complete, by substituting
// all variables and calling exec.Command.Run().
//
// The returned error is of type *exec.ExitError. Other error types may be
// returned for other situations.
//
// Stderr and Stdout output is accessable through Task.Stderr and Task.Stdout
//
// If Task.Always is false (default), the task is only started if at least one
// input file is newer than the output files.
func (tsk *Task) Run() error {
	if !tsk.Always && !tsk.needsUpdate() {
		return nil
	}

	mapper := func(name string) string {
		switch name {
		case "SRCS":
			return strings.Join(tsk.Inputs, " ")
		case "DESTS":
			return strings.Join(tsk.Outputs, " ")
		}
		if env := os.Getenv(name); env != "" {
			return env
		}
		if v, ok := tsk.Vars[name]; ok {
			return v
		}
		return fmt.Sprintf("${%s}", name)
	}

	tsk.Stdout.Reset()
	tsk.Stderr.Reset()

	cmd := exec.Command("/bin/sh", "-c", os.Expand(tsk.cmd, mapper))
	cmd.Stdout = &tsk.Stdout
	cmd.Stderr = &tsk.Stderr
	return cmd.Run()
}

// TODO(5nord) Implement timestamp comparison.
func (tsk *Task) needsUpdate() bool {
	return true
}
