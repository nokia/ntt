package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/k3/k3r"
	"github.com/nokia/ntt/tests"
)

func New(n int, jobs <-chan *tests.Job) *Runner {
	return &Runner{
		maxWorkers: n,
		names:      make(map[string]int),
		jobs:       jobs,
	}
}

// Runner is a worker pool for executing jobs.
type Runner struct {
	sync.Mutex
	maxWorkers int
	names      map[string]int
	jobs       <-chan *tests.Job
}

func (r *Runner) Run(ctx context.Context) <-chan tests.Event {
	wg := sync.WaitGroup{}
	results := make(chan tests.Event, r.maxWorkers)
	for i := 0; i < r.maxWorkers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for job := range r.jobs {
				log.Debugf("worker #%02d: execute %s\n", i, job.Name)
				r.run(ctx, job, results)
			}
		}(i)
	}

	out := make(chan tests.Event, r.maxWorkers)
	go func() {
		const secs = time.Duration(30.0)
		ticker := time.NewTicker(secs * time.Second)
		for {
			select {
			case res, ok := <-results:
				if !ok {
					close(out)
					return
				}
				ticker.Reset(secs * time.Second)
				out <- res
			case <-ticker.C:
				out <- tests.NewTickerEvent()
			}
		}
	}()

	// Wait for all workers to finish.
	go func() {
		wg.Wait()
		close(results)
	}()

	return out
}

// execute runs a single test and sends the results to the channel.
func (r *Runner) run(ctx context.Context, job *tests.Job, results chan<- tests.Event) {
	var (
		workingDir string
		logFile    string
	)

	r.Lock()
	id := fmt.Sprintf("%s-%d", job.Name, r.names[job.Name])
	r.names[job.Name]++
	r.Unlock()

	if job.Dir != "" {
		workingDir = filepath.Join(job.Dir, id)
		if err := os.MkdirAll(workingDir, 0755); err != nil {
			results <- tests.NewErrorEvent(err)
			return
		}
	} else {
		logFile = fmt.Sprintf("%s.log", strings.TrimSuffix(id, "-0"))
	}

	t3xf := job.Config.K3.T3XF
	if workingDir != "" {
		absT3xf, err := filepath.Abs(t3xf)
		if err != nil {
			results <- tests.NewErrorEvent(err)
			return
		}
		absDir, err := filepath.Abs(workingDir)
		if err != nil {
			results <- tests.NewErrorEvent(err)
			return
		}
		t3xf, err = filepath.Rel(absDir, absT3xf)
		if err != nil {
			results <- tests.NewErrorEvent(err)
			return
		}
	}

	t := k3r.NewTest(t3xf, job.Name)
	t.Job = job

	var (
		timeout time.Duration
		err     error
	)
	if err != nil {
		results <- tests.NewErrorEvent(err)
		return
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

	for event := range t.Run(ctx) {
		results <- event
	}
}
