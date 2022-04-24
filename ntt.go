package ntt

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/k3/run"
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/doc"
	"github.com/spf13/pflag"
)

// BasketFlags returns a flagset with all flags for filtering objects.
//
// BasketFlags are regular expressions to filter objects. If you pass multiple regular
// expressions, all of them must match (AND). Example:
//
// 	$ cat example.ttcn3
// 	testcase foo() ...
// 	testcase bar() ...
// 	testcase foobar() ...
// 	...
//
// 	$ ntt list --regex=foo --regex=bar
// 	example.foobar
//
// 	$ ntt list --regex='foo|bar'
// 	example.foo
// 	example.bar
// 	example.foobar
//
//
// Similarly, you can also specify regular expressions for documentation tags.
// Example:
//
// 	$ cat example.ttcn3
// 	// @one
// 	// @two some-value
// 	testcase foo() ...
//
// 	// @two: some-other-value
// 	testcase bar() ...
// 	...
//
// 	$ ntt list --tags-regex=@one --tags-regex=@two
// 	example.foo
//
// 	$ ntt list --tags-regex='@two: some'
// 	example.foo
// 	example.bar
//
func BasketFlags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("basket", pflag.ContinueOnError)
	fs.StringSliceP("regex", "r", nil, "list objects matching regular * expression.")
	fs.StringSliceP("exclude", "x", nil, "exclude objects matching regular * expresion.")
	fs.StringSliceP("tags-regex", "R", nil, "list objects with tags matching regular * expression")
	fs.StringSliceP("tags-exclude", "X", nil, "exclude objects with tags matching * regular expression")
	return fs
}

// A Basket is a filter for objects. It can be used to filter objects by name
// and tags.
//
// Baskets are also filters defined by environment variables of the form:
//
//         NTT_LIST_BASKETS_<name> = <filters>
//
// For example, to define a basket "stable" which excludes all objects with @wip
// or @flaky tags:
//
// 	export NTT_LIST_BASKETS_stable="-X @wip|@flaky"
//
// Baskets become active when they are listed in colon separated environment
// variable NTT_LIST_BASKETS. If you specify multiple baskets, at least of them
// must match (OR).
//
// Rule of thumb: all baskets are ORed, all explicit filter options are ANDed.
// Example:
//
// 	$ export NTT_LIST_BASKETS_stable="--tags-exclude @wip|@flaky"
// 	$ export NTT_LIST_BASKETS_ipv6="--tags-regex @ipv6"
// 	$ NTT_LIST_BASKETS=stable:ipv6 ntt list -R @flaky
//
//
// Above example will output all tests with a @flaky tag and either @wip or @ipv6 tag.
//
// If a basket is not defined by an environment variable, it's equivalent to a
// "--tags-regex" filter. For example, to lists all tests, which have either a
// @flaky or a @wip tag:
//
// 	# Note, flaky and wip baskets are not specified explicitly.
// 	$ NTT_LIST_BASKETS=flaky:wip ntt list
//
// 	# This does the same:
// 	$ ntt list --tags-regex="@wip|@flaky"
//
type Basket struct {
	// Name is the name of the basket. The basket is used to filter objects
	// by tag, if no explicit filters are given.
	Name string

	// Regular expressions the object name must match.
	NameRegex []string

	// Regular expressions the object name must not match.
	NameExclude []string

	// Regular expressions the object tags must match.
	TagsRegex []string

	// Regular expressions the object tags must not match.
	TagsExclude []string

	// Baskets are sub-baskets to be ORed.
	Baskets []Basket
}

// NewBasket creates a new basket and parses the given arguments.
func NewBasket(name string, args ...string) (Basket, error) {

	fs := pflag.NewFlagSet(name, pflag.ContinueOnError)
	fs.AddFlagSet(BasketFlags())
	if err := fs.Parse(args); err != nil {
		return Basket{}, err
	}
	return NewBasketWithFlags(name, fs)
}

func NewBasketWithFlags(name string, fs *pflag.FlagSet) (Basket, error) {
	b := Basket{Name: name}

	var err error

	b.NameRegex, err = fs.GetStringSlice("regex")
	if err != nil {
		return b, err
	}

	b.NameExclude, err = fs.GetStringSlice("exclude")
	if err != nil {
		return b, err
	}
	b.TagsRegex, err = fs.GetStringSlice("tags-regex")
	if err != nil {
		return b, err
	}
	b.TagsExclude, err = fs.GetStringSlice("tags-exclude")
	if err != nil {
		return b, err
	}
	return b, nil
}

// Load baskets from given environment variable.
func (b *Basket) LoadFromEnv(key string) error {
	s := env.Getenv(key)
	if s == "" {
		return nil
	}

	for _, name := range strings.Split(s, ":") {
		// Ignore empty fields
		if name == "" {
			continue
		}
		args := strings.Fields(env.Getenv(fmt.Sprintf("%s_%s", key, name)))
		if len(args) == 0 {
			args = []string{"-R", "@" + name}
		}

		sb, err := NewBasket(name, args...)
		if err != nil {
			return err
		}
		b.Baskets = append(b.Baskets, sb)
	}
	return nil
}

// Match returns true if the given name and tags match the basket or sub-basket filters.
func (b *Basket) Match(name string, tags [][]string) bool {
	ok := b.match(name, tags)
	if len(b.Baskets) == 0 {
		return ok
	}

	for _, basket := range b.Baskets {
		if basket.Match(name, tags) && ok {
			return true
		}
	}
	return false
}

// match returns true if the given name and tags match the basket filters.
func (b *Basket) match(name string, tags [][]string) bool {
	if !b.matchAll(b.NameRegex, name) {
		return false
	}
	if len(b.NameExclude) > 0 && b.matchAll(b.NameExclude, name) {
		return false
	}

	if len(b.TagsRegex) > 0 {
		if len(tags) == 0 {
			return false
		}
		if !b.matchAllTags(b.TagsRegex, tags) {
			return false
		}
	}

	if len(b.TagsExclude) > 0 && b.matchAllTags(b.TagsExclude, tags) {
		return false
	}

	return true
}

// matchAll returns true if all regular expressions match the given string.
func (b *Basket) matchAll(regexes []string, s string) bool {
	for _, r := range regexes {
		if ok, _ := regexp.Match(r, []byte(s)); !ok {
			return false
		}
	}
	return true
}

// matchAllTags returns true if all regular expressions match the all given tags.
func (b *Basket) matchAllTags(regexes []string, tags [][]string) bool {
next:
	for _, r := range regexes {
		f := strings.SplitN(r, ":", 2)
		for i := range f {
			f[i] = strings.TrimSpace(f[i])
		}
		for _, tag := range tags {
			if ok, _ := regexp.Match(f[0], []byte(tag[0])); !ok {
				continue
			}
			if len(f) > 1 {
				if ok, _ := regexp.Match(f[1], []byte(tag[1])); !ok {
					continue
				}
			}
			continue next
		}
		return false
	}
	return true
}

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

// GenerateIDs emits given IDs to a channel.
func GenerateIDs(ids ...string) <-chan string {
	return GenerateIDsWithContext(context.Background(), ids...)
}

// GenerateIDs emits given IDs to a channel.
func GenerateIDsWithContext(ctx context.Context, ids ...string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, id := range ids {
			out <- id
		}
	}()
	return out
}

// GenerateTests emits all test ids from given TTCN-3 files to a channel.
func GenerateTests(files ...string) <-chan string {
	return GenerateTestsWithContext(context.Background(), Basket{}, files...)
}

// GenerateTestsWithContext emits all test ids from given TTCN-3 files to a channel.
func GenerateTestsWithContext(ctx context.Context, b Basket, files ...string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, src := range files {
			tree := ttcn3.ParseFile(src)
			for _, def := range tree.Funcs() {
				if n := def.Node.(*ast.FuncDecl); n.IsTest() {
					if !generate(ctx, def, b, out) {
						return
					}
				}
			}
		}
	}()
	return out
}

// GenerateControls emits all control function ids from given TTCN-3 files to a channel.
func GenerateControls(files ...string) <-chan string {
	return GenerateControlsWithContext(context.Background(), Basket{}, files...)
}

// GenerateControls emits all control function ids from given TTCN-3 files to a channel.
func GenerateControlsWithContext(ctx context.Context, b Basket, files ...string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, src := range files {
			tree := ttcn3.ParseFile(src)
			for _, n := range tree.Controls() {
				if !generate(ctx, n, b, out) {
					return
				}
			}

		}
	}()
	return out
}

func generate(ctx context.Context, def *ttcn3.Definition, b Basket, c chan<- string) bool {
	id := def.Tree.QualifiedName(def.Ident)
	tags := doc.FindAllTags(ast.FirstToken(def.Node).Comments())
	if b.Match(id, tags) {
		select {
		case c <- id:
		case <-ctx.Done():
			return false
		}
	}
	return true
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
	run.Test
	run.Event
}

func (r *Result) ID() string {
	return fmt.Sprintf("%s-%s", r.Job.ID(), r.Event.Name)
}

// Runner is a test runner.
type Runner interface {
	// Run the jobs in the given channel.
	Run(ctx context.Context, jobs <-chan *Job) <-chan Result
}

func NewLedger(n int) *Ledger {
	return &Ledger{
		maxWorkers: n,
		names:      make(map[string]int),
		jobs:       make(map[string]*Job),
	}
}

// Ledger is a worker pool for executing jobs.
type Ledger struct {
	sync.Mutex
	maxWorkers int
	names      map[string]int
	jobs       map[string]*Job
}

func (l *Ledger) NewJob(name string, suite *Suite) *Job {
	l.Lock()
	defer l.Unlock()

	job := Job{
		id:    fmt.Sprintf("%s-%d", name, l.names[name]),
		Name:  name,
		Suite: suite,
	}
	l.names[name]++
	l.jobs[job.id] = &job

	log.Debugf("new job: name=%s, suite=%p, id=%s\n", name, suite, job.id)
	return &job
}

func (l *Ledger) Done(job *Job) {
	l.Lock()
	defer l.Unlock()
	delete(l.jobs, job.id)
}

func (l *Ledger) Jobs() []*Job {
	l.Lock()
	defer l.Unlock()

	jobs := make([]*Job, 0, len(l.jobs))
	for _, job := range l.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

func (l *Ledger) Run(ctx context.Context, jobs <-chan *Job) <-chan Result {
	wg := sync.WaitGroup{}
	results := make(chan Result, l.maxWorkers)
	for i := 0; i < l.maxWorkers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			log.Debugf("Worker %d started.\n", i)
			defer log.Debugf("Worker %d finished.\n", i)

			for job := range jobs {
				l.run(ctx, job, results)
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
func (l *Ledger) run(ctx context.Context, job *Job, results chan<- Result) {

	defer l.Done(job)
	var (
		workingDir string
		logFile    string
	)

	if job.Dir == "" {
		logFile = fmt.Sprintf("%s.log", strings.TrimSuffix(job.ID(), "-0"))
	} else {
		workingDir = filepath.Join(job.Dir, job.ID())
		if err := os.MkdirAll(workingDir, 0755); err != nil {
			results <- Result{Job: job, Event: run.NewErrorEvent(err)}
			return
		}
	}

	t3xf := job.Suite.K3.T3XF
	if workingDir != "" {
		absT3xf, err := filepath.Abs(t3xf)
		if err != nil {
			results <- Result{Job: job, Event: run.NewErrorEvent(err)}
			return
		}
		absDir, err := filepath.Abs(workingDir)
		if err != nil {
			results <- Result{Job: job, Event: run.NewErrorEvent(err)}
			return
		}
		t3xf, err = filepath.Rel(absDir, absT3xf)
		if err != nil {
			results <- Result{Job: job, Event: run.NewErrorEvent(err)}
			return
		}
	}

	test := run.NewTest(t3xf, job.Name)

	var (
		pars    map[string]string
		timeout time.Duration
		err     error
	)
	if err != nil {
		results <- Result{Job: job, Event: run.NewErrorEvent(err)}
		return
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	test.ModulePars = pars
	test.Dir = workingDir
	test.LogFile = logFile
	test.Env = append(test.Env, job.Suite.Variables.Slice()...)
	test.Env = append(test.Env, fmt.Sprintf("K3R_PATH=%s:%s", strings.Join(job.Suite.RuntimePaths, ":"), os.Getenv("K3R_PATH")))
	test.Env = append(test.Env, fmt.Sprintf("LD_LIBRARY_PATH=%s:%s", strings.Join(job.Suite.RuntimePaths, ":"), os.Getenv("LD_LIBRARY_PATH")))
	for event := range test.RunWithContext(ctx) {
		results <- Result{
			Job:   job,
			Test:  *test,
			Event: event,
		}
	}
}
