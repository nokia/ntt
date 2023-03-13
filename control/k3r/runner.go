package k3r

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/control"
)

// Runner executes tests using k3r.
type Runner struct {
	jobs <-chan *control.Job
}

func Factory(jobs <-chan *control.Job) control.RunnerFactory {
	return func() (control.Runner, error) { return NewRunner(jobs), nil }
}

func NewRunner(jobs <-chan *control.Job) *Runner {
	return &Runner{jobs: jobs}
}

func (r *Runner) Run(ctx context.Context) <-chan control.Event {
	var (
		workingDir string
		logFile    string
	)

	results := make(chan control.Event)
	go func() {
		defer close(results)
		for job := range r.jobs {
			if job.Dir != "" {
				workingDir = filepath.Join(job.Dir, job.ID)
				if err := os.MkdirAll(workingDir, 0755); err != nil {
					results <- control.NewErrorEvent(err)
					continue
				}
			} else {
				logFile = fmt.Sprintf("%s.log", strings.TrimSuffix(job.ID, "-0"))
			}

			t3xf := job.Config.K3.T3XF
			if workingDir != "" {
				absT3xf, err := filepath.Abs(t3xf)
				if err != nil {
					results <- control.NewErrorEvent(err)
					continue
				}
				absDir, err := filepath.Abs(workingDir)
				if err != nil {
					results <- control.NewErrorEvent(err)
					continue
				}
				t3xf, err = filepath.Rel(absDir, absT3xf)
				if err != nil {
					results <- control.NewErrorEvent(err)
					continue
				}
			}

			t := NewTest(t3xf, job)

			var (
				timeout time.Duration
				err     error
			)
			if err != nil {
				results <- control.NewErrorEvent(err)
				continue
			}
			if timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}

			// TODO(5nord) implement module parameters
			t.Dir = workingDir
			t.LogFile = logFile
			t.Env = env.Environ()
			if s := env.Getenv("NTT_CACHE"); s != "" {
				t.Env = append(t.Env, strings.Split(s, string(os.PathListSeparator))...)
			}
			for e := range t.Run(ctx) {
				results <- e
			}
		}
	}()
	return results
}
