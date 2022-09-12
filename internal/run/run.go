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
	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/k3/k3r"
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/tests"
)

// SplitQualifiedName splits a qualified name into module and test name.
func SplitQualifiedName(name string) (string, string) {
	parts := strings.Split(name, ".")
	if len(parts) == 1 {
		return "", name
	}
	return parts[0], strings.Join(parts[1:], ".")
}

// NewSuite creates a new suite from the given files. It expects either
// a single directory as argument or a list of regular .ttcn3 files.
//
// Calling NewSuite with an empty argument list will create a suite from
// current working directory or, if set, from NTT_SOURCE_DIR.
//
// NewSuite will read manifest (package.yml) if any.
func NewSuite(p *project.Config) (*Suite, error) {
	if err := project.Build(p); err != nil {
		return nil, fmt.Errorf("building test suite failed: %w", err)
	}

	var paths []string
	if s := env.Getenv("NTT_CACHE"); s != "" {
		paths = append(paths, strings.Split(s, string(os.PathListSeparator))...)
	}
	paths = append(paths, p.Manifest.Imports...)
	paths = append(paths, k3.Plugins()...)
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, cwd)
	}

	return &Suite{
		Config:       p,
		RuntimePaths: paths,
	}, nil

}

// Suite represents a test suite.
type Suite struct {
	*project.Config
	RuntimePaths []string
}

// Job represents a single job to be executed.
type Job struct {
	// Full qualified name of the test or control function to be executed.
	Name string

	// Working directory for the job.
	Dir string

	// Test suite the job belongs to.
	Suite *Suite

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

func (r *Runner) NewJob(name string, suite *Suite) *Job {
	r.Lock()
	defer r.Unlock()

	job := Job{
		id:    fmt.Sprintf("%s-%d", name, r.names[name]),
		Name:  name,
		Suite: suite,
	}
	r.names[name]++
	r.jobs[job.id] = &job

	log.Debugf("new job: name=%s, suite=%p, id=%s\n", name, suite, job.id)
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
			log.Debugf("Worker %d started.\n", i)
			defer log.Debugf("Worker %d finished.\n", i)

			for job := range jobs {
				r.run(ctx, job, results)
			}
		}(i)
	}

	// Wait for all workers to finish.
	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

// execute runs a single test and sends the results to the channel.
func (r *Runner) run(ctx context.Context, job *Job, results chan<- Result) {

	defer r.Done(job)
	var (
		workingDir string
		logFile    string
	)

	if job.Dir == "" {
		logFile = fmt.Sprintf("%s.log", strings.TrimSuffix(job.ID(), "-0"))
	} else {
		workingDir = filepath.Join(job.Dir, job.ID())
		if err := os.MkdirAll(workingDir, 0755); err != nil {
			results <- Result{Job: job, Event: tests.NewErrorEvent(err)}
			return
		}
	}

	t3xf := job.Suite.K3.T3XF
	if workingDir != "" {
		absT3xf, err := filepath.Abs(t3xf)
		if err != nil {
			results <- Result{Job: job, Event: tests.NewErrorEvent(err)}
			return
		}
		absDir, err := filepath.Abs(workingDir)
		if err != nil {
			results <- Result{Job: job, Event: tests.NewErrorEvent(err)}
			return
		}
		t3xf, err = filepath.Rel(absDir, absT3xf)
		if err != nil {
			results <- Result{Job: job, Event: tests.NewErrorEvent(err)}
			return
		}
	}

	t := k3r.NewTest(t3xf, job.Name)
	t.Config = job.Suite.Config

	var (
		pars    map[string]string
		timeout time.Duration
		err     error
	)
	if err != nil {
		results <- Result{Job: job, Event: tests.NewErrorEvent(err)}
		return
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	t.ModulePars = pars
	t.Dir = workingDir
	t.LogFile = logFile
	t.Env = append(t.Env, job.Suite.Variables.Slice()...)
	t.Env = append(t.Env, env.Environ()...)
	t.Env = append(t.Env, fmt.Sprintf("K3R_PATH=%s:%s", strings.Join(job.Suite.RuntimePaths, ":"), os.Getenv("K3R_PATH")))
	t.Env = append(t.Env, fmt.Sprintf("LD_LIBRARY_PATH=%s:%s", strings.Join(job.Suite.RuntimePaths, ":"), os.Getenv("LD_LIBRARY_PATH")))
	for event := range t.RunWithContext(ctx) {
		results <- Result{
			Job:   job,
			Test:  *t,
			Event: event,
		}
	}
}
