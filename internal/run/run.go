package run

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
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/tests"
)

// Job represents a single job to be executed.
type Job struct {
	// Full qualified name of the test or control function to be executed.
	Name string

	// Working directory for the job.
	Dir string

	// Test suite the job belongs to.
	*project.Config

	id string
}

// A unique job identifier.
func (j *Job) ID() string {
	return j.id
}

type Result struct {
	*Job
	k3r.Test
	tests.Event
}

func (r *Result) ID() string {
	name := ""
	if job := tests.UnwrapJob(r.Event); job != nil {
		name = job.Name
	}
	return fmt.Sprintf("%s-%s", r.Job.ID(), name)
}

func NewRunner(n int) *Runner {
	return &Runner{
		maxWorkers: n,
		names:      make(map[string]int),
		jobs:       make(map[string]*Job),
	}
}

// Runner is a worker pool for executing jobs.
type Runner struct {
	sync.Mutex
	maxWorkers int
	names      map[string]int
	jobs       map[string]*Job
}

func (r *Runner) NewJob(name string, conf *project.Config) *Job {
	r.Lock()
	defer r.Unlock()

	job := Job{
		id:     fmt.Sprintf("%s-%d", name, r.names[name]),
		Name:   name,
		Config: conf,
	}
	r.names[name]++
	r.jobs[job.id] = &job

	log.Debugf("new job: name=%s, conf=%p, id=%s\n", name, conf, job.id)
	return &job
}

func (r *Runner) Done(job *Job) {
	r.Lock()
	defer r.Unlock()
	delete(r.jobs, job.id)
}

func (r *Runner) Jobs() []*Job {
	r.Lock()
	defer r.Unlock()

	jobs := make([]*Job, 0, len(r.jobs))
	for _, job := range r.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

func (r *Runner) Run(ctx context.Context, jobs <-chan *Job) <-chan Result {
	wg := sync.WaitGroup{}
	results := make(chan Result, r.maxWorkers)
	for i := 0; i < r.maxWorkers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for job := range jobs {
				log.Debugf("worker #%02d: execute %s\n", i, job.Name)
				r.run(ctx, job, results)
			}
		}(i)
	}

	out := make(chan Result, r.maxWorkers)
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
				out <- Result{
					Event: tests.NewTickerEvent(),
				}
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

func jobError(j *Job, err error) Result {
	return Result{Job: j, Event: tests.NewErrorEvent(err)}
}

// execute runs a single test and sends the results to the channel.
func (r *Runner) run(ctx context.Context, job *Job, results chan<- Result) {
	defer r.Done(job)

	var (
		workingDir string
		logFile    string
	)

	if job.Dir != "" {
		workingDir = filepath.Join(job.Dir, job.ID())
		if err := os.MkdirAll(workingDir, 0755); err != nil {
			results <- jobError(job, err)
			return
		}
	} else {
		logFile = fmt.Sprintf("%s.log", strings.TrimSuffix(job.ID(), "-0"))
	}

	t3xf := job.Config.K3.T3XF
	if workingDir != "" {
		absT3xf, err := filepath.Abs(t3xf)
		if err != nil {
			results <- jobError(job, err)
			return
		}
		absDir, err := filepath.Abs(workingDir)
		if err != nil {
			results <- jobError(job, err)
			return
		}
		t3xf, err = filepath.Rel(absDir, absT3xf)
		if err != nil {
			results <- jobError(job, err)
			return
		}
	}

	t := k3r.NewTest(t3xf, job.Name)
	t.Config = job.Config

	var (
		timeout time.Duration
		err     error
	)
	if err != nil {
		results <- jobError(job, err)
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

	for event := range t.RunWithContext(ctx) {
		results <- Result{
			Job:   job,
			Test:  *t,
			Event: event,
		}
	}
}
