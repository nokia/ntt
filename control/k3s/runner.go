package k3s

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/nokia/ntt/control"
	"github.com/nokia/ntt/internal/fs"
)

type Runner struct {
	jobs <-chan *control.Job
	w    io.Writer
}

func NewRunner(jobs <-chan *control.Job, w io.Writer) *Runner {
	return &Runner{jobs: jobs, w: w}
}

func (r *Runner) Run(ctx context.Context) <-chan control.Event {
	results := make(chan control.Event)
	go func() {
		for job := range r.jobs {
			results <- control.NewStartEvent(job, job.Name)

			fmt.Fprintf(r.w, `
===============================================================================
Compiling test %s in %q`, job.Name, job.Config.Root)

			if err := Build(r.w, job.Config); err != nil {
				fmt.Fprintln(r.w, err.Error())
				results <- control.NewErrorEvent(&control.JobError{Job: job, Err: err})
				continue
			}

			fmt.Fprintf(r.w, `
===============================================================================
Running test %s in %q`, job.Name, job.Config.Root)

			logDir, _ := Run(r.w, job.Config, job.Name)

			if files := fs.Abs(excludeFromListIfPresent("mtc_workspace", fs.FindFilesRecursive(logDir))...); len(files) > 0 {
				fmt.Fprintf(r.w, `
Content of log directory %q:
===============================================================================
%s
`,
					logDir, strings.Join(files, "\n"))
			}
			results <- control.NewStopEvent(job, job.Name, "")

		}
	}()
	return results
}

func excludeFromListIfPresent(str string, input []string) []string {
	ret := make([]string, 0, len(input))
	for _, elem := range input {
		if !strings.Contains(elem, str) {
			ret = append(ret, elem)
		}
	}
	return ret
}
