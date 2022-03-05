package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
	ntt2 "github.com/nokia/ntt"
	"github.com/nokia/ntt/internal/cache"
	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/internal/proc"
	"github.com/nokia/ntt/internal/results"
	"github.com/nokia/ntt/k3"
	k3r "github.com/nokia/ntt/k3/run"
	"github.com/nokia/ntt/project"
	"github.com/spf13/cobra"
)

var (
	RunCommand = &cobra.Command{
		Use:   "run [ <path>... ] [ -- <test id>... ]",
		Short: "Build and run test suite",
		Long: `Run tests from a TTCN-3 test suite.

The ntt run command first builds a test executable using the files or
directories passed as first argument list.
The test executable is then run with the tests specified as second argument
list. If no ids are specified, ntt run will run all tests in the test suite.

Running control functions is supported. For example:

        ntt run -- test.A test.control test.B


Test baskets are also supported (see "ntt help list"). Bellow example will run
all tests with @stable-tag:

	NTT_LIST_BASKETS=stable ntt run


Environment variables:

* SCT_K3_SERVER=on	use k3s as backend.
* K3_40_RUN_POLICY=old	if no ids are specified, run all control parts.

`,

		PreRunE: func(cmd *cobra.Command, args []string) error {
			files, _ := splitArgs(args, cmd.ArgsLenAtDash())
			return BuildCommand.RunE(cmd, files)
		},
		RunE: run,
	}

	MaxWorkers int
	MaxFail    int
	errorCount uint64
	OutputDir  string

	ColorFatal   = color.New(color.FgRed, color.Bold)
	ColorFailure = color.New(color.FgRed, color.Bold)
	ColorWarning = color.New(color.FgYellow, color.Bold)
	ColorSuccess = color.New()
	ColorStart   = color.New()
	ColorRunning = color.New(color.Faint)
	Colors       = func(v string) *color.Color {
		switch v {
		case "pass":
			return ColorSuccess
		case "inconc":
			return ColorWarning
		case "none":
			return ColorWarning
		default:
			return ColorFailure
		}
	}

	ErrCommandFailed = fmt.Errorf("command failed")

	Runs      []results.Run
	ledger    = NewLedger()
	stamps    = make(map[string]time.Time)
	Basket, _ = ntt2.NewBasket("default")

	ResultsFile = cache.Lookup("test_results.json")
	TickerTime  = time.Second * 30
)

type Job struct {
	id        string
	Name      string
	Iteration int
	Suite     *Suite
}

func (j *Job) ID() string {
	return j.id
}

type Ledger struct {
	sync.Mutex
	names map[string]int
	jobs  map[string]*Job
}

func NewLedger() *Ledger {
	return &Ledger{
		names: make(map[string]int),
		jobs:  make(map[string]*Job),
	}
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

type Result struct {
	*Job
	k3r.Test
	k3r.Event
}

func (r *Result) ID() string {
	return fmt.Sprintf("%s-%s", r.Job.ID(), r.Event.Name)
}

func init() {
	flags := RunCommand.Flags()
	flags.AddFlagSet(ntt2.BasketFlags())
	flags.IntVarP(&MaxWorkers, "jobs", "j", runtime.NumCPU(), "Allow N test in parallel (default: number of CPU cores")
	flags.IntVar(&MaxFail, "max-fail", 0, "Stop after N failures")
	flags.StringVarP(&OutputDir, "output-dir", "o", "", "store test artefacts in DIR/ID")
	flags.StringP("tests-file", "t", "", "Read tests from file (use '-' for stdin)")
}

type Suite struct {
	Suite        *ntt.Suite
	Name         string
	Sources      []string
	RuntimePaths []string
}

func NewSuite(files ...string) (*Suite, error) {
	suite, err := ntt.NewFromArgs(files...)
	if err != nil {
		return nil, fmt.Errorf("loading test suite failed: %w", err)
	}

	name, err := suite.Name()
	if err != nil {
		return nil, fmt.Errorf("retrieving test suite name failed: %w", err)
	}

	srcs, err := suite.Sources()
	if err != nil {
		return nil, fmt.Errorf("retrieving TTCN-3 sources failed: %w", err)
	}

	paths, err := runtimePaths(suite)
	if err != nil {
		return nil, fmt.Errorf("retrieving runtime paths failed: %w", err)
	}

	return &Suite{
		Suite:        suite,
		Name:         name,
		Sources:      srcs,
		RuntimePaths: paths,
	}, nil

}

func (s *Suite) Timeout() time.Duration {
	t, err := s.Suite.Timeout()
	if err != nil {
		log.Verbosef("retrieving test suite timeout failed: %v\n", err)
	}
	d, err := time.ParseDuration(fmt.Sprintf("%fs", t))
	if err != nil {
		log.Verbosef("retrieving test suite timeout failed: %v\n", err)
	}
	return d
}

func (s *Suite) ParametersFile() string {
	f, err := s.Suite.ParametersFile()
	if err != nil {
		log.Verbosef("retrieving parameters file failed: %v", err)
		return ""
	}
	if f != nil {
		return f.Path()
	}
	return ""
}

func (s *Suite) ParametersDir() string {
	dir, err := s.Suite.ParametersDir()
	if err != nil {
		log.Verbosef("retrieving parameters file failed: %v", err)
	}
	return dir
}

// Run runs the given jobs in parallel.
func run(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()
	go func() {
		select {
		case <-signalChan: // first signal, cancel context
			cancel()
		case <-ctx.Done():
		}
		<-signalChan // second signal, hard exit
		os.Exit(2)
	}()
	if OutputDir != "" {
		if err := os.MkdirAll(OutputDir, os.ModePerm); err != nil {
			return fmt.Errorf("creating logs directory failed: %w", err)
		}
	}

	var err error
	Basket, err = ntt2.NewBasketWithFlags("list", cmd.Flags())
	Basket.LoadFromEnv("NTT_LIST_BASKETS")
	if err != nil {
		return err
	}

	files, ids := splitArgs(args, cmd.ArgsLenAtDash())
	suite, err := NewSuite(files...)
	if err != nil {
		return err
	}

	if path := cmd.Flag("tests-file").Value.String(); path != "" {
		tests, err := readTestsFromFile(path)
		if err != nil {
			return err
		}
		ids = append(tests, ids...)
	}

	jobs := GenerateJobs(ctx, suite, ids, MaxWorkers)

	if s, ok := os.LookupEnv("SCT_K3_SERVER"); ok && s != "ntt" && strings.ToLower(s) != "off" {
		return k3sRun(ctx, files, jobs)
	}
	return nttRun(ctx, jobs)
}

func nttRun(ctx context.Context, jobs <-chan *Job) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	os.Remove(cache.Lookup("test_results.json"))
	defer FlushTestResults()

	// Execute the jobs in parallel and collect the results.
	ticker := time.NewTicker(TickerTime)
	ressults := ExecuteJobs(ctx, jobs, MaxWorkers)
L:
	for {
		select {
		case res, ok := <-ressults:
			if !ok {
				break L
			}
			ticker.Reset(TickerTime)
			log.Debugf("result: jobID=%s, event=%#v\n", res.Job.ID(), res.Event)

			// Track errors for early exit (aka --max-fail).
			if res.Event.IsError() {
				errorCount++
			}

			// Track begin timestamps for calculating durations.
			begin, end := stamps[res.ID()], res.Event.Time
			if begin.IsZero() {
				stamps[res.ID()] = time.Now()
				begin = end
			}
			duration := end.Sub(begin)

			displayName := res.Job.Name
			if n := res.Event.Name; n != "" {
				displayName = n
			}

			displayVerdict := res.Event.Verdict
			if displayVerdict == "" {
				displayVerdict = "none"
			}
			if res.Type == k3r.Error {
				displayVerdict = "fatal"
			}

			run := results.Run{
				Name:       displayName,
				Verdict:    displayVerdict,
				Begin:      results.Timestamp{Time: begin},
				End:        results.Timestamp{Time: end},
				WorkingDir: res.Test.Dir,
			}
			if res.Err != nil {
				run.Reason = res.Err.Error()
			}

			if !res.IsStartEvent() {
				Runs = append(Runs, run)
			}

			switch Format() {
			case "quiet":
				// No output
			case "plain":
				if !res.IsStartEvent() {
					c := Colors(displayVerdict)
					c.Printf("%s\t%s\t%.4f\n", displayVerdict, displayName, float64(duration.Seconds()))
				}
			case "json":
				if !res.IsStartEvent() {
					b, err := json.Marshal(run)
					if err != nil {
						panic(fmt.Sprintf("cannot marshal run: %v", run))
					}
					fmt.Println(string(b))
				}
			case "text":
				switch {
				case res.IsStartEvent():
					ColorStart.Printf("=== RUN %s\n", displayName)
				case res.Type == k3r.Error:
					ColorFatal.Printf("+++ fatal %s\t(%s)\n", displayName, run.Reason)
				default:
					c := Colors(displayVerdict)
					c.Printf("--- %s %s\t(duration=%.2fs)\n", displayVerdict, displayName, float64(duration.Seconds()))
				}
			default:
				panic(fmt.Sprintf("unknown format: %s", Format()))
			}

			if MaxFail > 0 && errorCount >= uint64(MaxFail) {
				ColorFatal.Print("+++ fatal too many errors. Exiting.\n")
				cancel()
				break L
			}
		case <-ticker.C:
			for _, job := range ledger.Jobs() {
				ColorRunning.Printf("... running %s\n", job.Name)
			}
		}
	}
	if errorCount > 0 {
		return fmt.Errorf("%w: %d error(s) occurred", ErrCommandFailed, errorCount)
	}
	return nil
}

func k3sRun(ctx context.Context, files []string, jobs <-chan *Job) error {
	args := []string{
		"--no-summary",
		fmt.Sprintf("--results-file=%s", ResultsFile),
		fmt.Sprintf("-j%d", MaxWorkers),
	}
	args = append(args, files...)
	k3s := proc.CommandContext(ctx, "k3s", args...)
	k3s.Stdin = k3sJobs(jobs)
	k3s.Stdout = os.Stdout
	k3s.Stderr = os.Stderr
	log.Verboseln("+", k3s.String())
	return k3s.Run()
}

func k3sJobs(jobs <-chan *Job) io.Reader {
	var ids []string
	for j := range jobs {
		ids = append(ids, j.Name)
	}
	return strings.NewReader(strings.Join(ids, "\n"))
}

// GenerateIDs emits test IDs based on given file and and id list to a channel.
func GenerateIDs(ctx context.Context, ids []string, files []string, policy string, b ntt2.Basket) <-chan string {
	policy = strings.ToLower(policy)
	policy = strings.TrimSpace(policy)
	switch {
	case len(ids) > 0:
		return ntt2.GenerateIDsWithContext(ctx, ids...)
	case policy == "old":
		return ntt2.GenerateControlsWithContext(ctx, b, files...)
	default:
		return ntt2.GenerateTestsWithContext(ctx, b, files...)

	}
}

// GenerateJobs emits jobs from the given suite and ids to a job channel.
func GenerateJobs(ctx context.Context, suite *Suite, ids []string, size int) chan *Job {
	out := make(chan *Job, size)
	go func() {
		defer close(out)
		i := 0
		for id := range GenerateIDs(ctx, ids, suite.Sources, env.Getenv("K3_40_RUN_POLICY"), Basket) {
			i++
			out <- ledger.NewJob(id, suite)
		}
		log.Debugf("Generating %d jobs done.\n", i)
	}()
	return out
}

func ExecuteJobs(ctx context.Context, jobs <-chan *Job, n int) <-chan Result {
	wg := sync.WaitGroup{}
	results := make(chan Result, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			log.Debugf("Worker %d started.\n", i)
			defer log.Debugf("Worker %d finished.\n", i)

			for job := range jobs {
				Execute(ctx, job, results)
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

// Execute runs a single test and sends the results to the channel.
func Execute(ctx context.Context, job *Job, results chan<- Result) {

	defer ledger.Done(job)
	var (
		workingDir string
		logFile    string
	)

	if OutputDir == "" {
		logFile = fmt.Sprintf("%s.log", strings.TrimSuffix(job.ID(), "-0"))
	} else {
		workingDir = filepath.Join(OutputDir, job.ID())
		if err := os.MkdirAll(workingDir, 0755); err != nil {
			results <- Result{Job: job, Event: k3r.NewErrorEvent(err)}
			return
		}
	}

	t3xf := cache.Lookup(fmt.Sprintf("%s.t3xf", job.Suite.Name))
	if workingDir != "" {
		absT3xf, err := filepath.Abs(t3xf)
		if err != nil {
			results <- Result{Job: job, Event: k3r.NewErrorEvent(err)}
			return
		}
		absDir, err := filepath.Abs(workingDir)
		if err != nil {
			results <- Result{Job: job, Event: k3r.NewErrorEvent(err)}
			return
		}
		t3xf, err = filepath.Rel(absDir, absT3xf)
		if err != nil {
			results <- Result{Job: job, Event: k3r.NewErrorEvent(err)}
			return
		}
	}

	test := k3r.NewTest(t3xf, job.Name)

	pars, err := ModulePars(job.Name, job.Suite)
	if err != nil {
		results <- Result{Job: job, Event: k3r.NewErrorEvent(err)}
		return
	}
	test.ModulePars = pars
	test.Dir = workingDir
	test.LogFile = logFile
	test.Env = append(test.Env, fmt.Sprintf("K3R_PATH=%s", strings.Join(job.Suite.RuntimePaths, ":")))
	test.Env = append(test.Env, fmt.Sprintf("LD_LIBRARY_PATH=%s", strings.Join(job.Suite.RuntimePaths, ":")))
	for event := range test.RunWithContext(ctx) {
		results <- Result{
			Job:   job,
			Test:  *test,
			Event: event,
		}
	}
}

// Default Timeout: package.yml
// Suite Paremters: $K3_PARAMETERS_FILE
// Test parameters: $K3_PARAMETERS_DIR/$MODULE/$TEST.parameters
// $K3_TIMEOUT
func ModulePars(name string, suite *Suite) (map[string]string, error) {
	m := make(map[string]string)

	if suitePars := suite.ParametersFile(); suitePars != "" {
		if b, err := fs.Content(suitePars); err == nil {
			if _, err := toml.Decode(string(b), &m); err != nil {
				return nil, err
			}
		}
	}
	if mod, test := SplitQualifiedName(name); mod != "" {
		testPars := filepath.Join(suite.ParametersDir(), mod, test+".parameters")
		if b, err := fs.Content(testPars); err == nil {
			if _, err := toml.Decode(string(b), &m); err != nil {
				return nil, err
			}
		}
	}
	return m, nil
}

func SplitQualifiedName(name string) (string, string) {
	parts := strings.Split(name, ".")
	if len(parts) == 1 {
		return "", name
	}
	return parts[0], strings.Join(parts[1:], ".")
}

func FlushTestResults() error {
	db := &results.DB{
		Version: "1",
		Sessions: []results.Session{
			{
				Id:              "1",
				MaxJobs:         MaxWorkers,
				ExpectedVerdict: "pass",
				Runs:            Runs,
			},
		},
	}
	b, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(cache.Lookup("test_results.json"), b, 0644)
}

// runtimePaths returns the paths to the adapters and runtime libraries for the given test suite.
func runtimePaths(p project.Interface) ([]string, error) {
	imports, err := p.Imports()
	if err != nil {
		return nil, fmt.Errorf("suite imports: %w", err)
	}

	var paths []string
	if s := env.Getenv("NTT_CACHE"); s != "" {
		paths = append(paths, strings.Split(s, ":")...)
	}
	paths = append(imports, k3.FindAuxiliaryDirectories()...)
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, cwd)
	}
	return paths, nil
}

func readTestsFromFile(path string) ([]string, error) {
	var (
		lines []byte
		err   error
	)
	if path == "-" {
		lines, err = ioutil.ReadAll(os.Stdin)
	} else {
		f, ferr := os.Open(path)
		if ferr != nil {
			return nil, ferr
		}
		lines, err = ioutil.ReadAll(f)

	}
	var tests []string
	for _, line := range strings.Split(string(lines), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			tests = append(tests, line)
		}
	}
	return tests, err
}
