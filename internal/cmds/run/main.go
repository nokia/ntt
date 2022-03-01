package run

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
	ntt2 "github.com/nokia/ntt"
	"github.com/nokia/ntt/internal/cache"
	"github.com/nokia/ntt/internal/cmds/build"
	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/internal/results"
	"github.com/nokia/ntt/k3"
	k3r "github.com/nokia/ntt/k3/run"
	"github.com/nokia/ntt/project"
	"github.com/spf13/cobra"
)

type Job struct {
	Name      string
	Iteration int
	Suite     *Suite
}

type Result struct {
	*Job
	k3r.Test
	k3r.Event
}

var (
	Command = &cobra.Command{
		Use:   "run [ <path>... ] [ -- <test id>... ]",
		Short: "Run tests from a TTCN-3 test suite.",
		Long: `Run tests from a TTCN-3 test suite.

The ntt run first command builds a test executable using the files or
directories passed as first argument list.
The test executable is then run with the tests specified as second argument
list.
If no ids are specified, ntt run will run all tests in the test suite.
Running control functions is supported. For example:

        ntt run -- test.A test.control test.B



The ntt run command supports ntt list test baskets (see "ntt help list").
Bellow example will run all tests with @stable tag:

	NTT_LIST_BASKETS=stable ntt run


Environment variables:

* SCT_K3_SERVER=on	replace ntt process with k3s.
* K3_40_RUN_POLICY=old	if no ids are specified, run all control parts.

`,

		PreRunE: func(cmd *cobra.Command, args []string) error {
			files, _ := splitArgs(args, cmd.ArgsLenAtDash())
			return build.Command.RunE(cmd, files)
		},
		RunE: run,
	}

	MaxWorkers int
	OutputJSON bool
	errorCount uint64

	fatal   = color.New(color.FgRed).Add(color.Bold)
	failure = color.New(color.FgRed).Add(color.Bold)
	warning = color.New(color.FgYellow).Add(color.Bold)
	success = color.New(color.FgGreen)

	ErrCommandFailed = fmt.Errorf("command failed")

	Runs      []results.Run
	Ledger    = make(map[*Job]*Result)
	Basket, _ = ntt2.NewBasket("default")
)

func init() {
	flags := Command.Flags()
	flags.IntVarP(&MaxWorkers, "jobs", "j", runtime.NumCPU(), "number of parallel tests")
	flags.BoolVarP(&OutputJSON, "json", "", false, "output in JSON format")

	flags.Bool("build", false, "build test suite")
	flags.Bool("no-summary", false, "disable test summary")
	flags.StringP("engine", "", "syntax", "what engine to use (t3xf or syntax)")

	flags.MarkHidden("build")
	flags.MarkHidden("no-summary")

	Basket.LoadFromEnv("NTT_LIST_BASKETS")

	// When SCT_K3_SERVER is set, we hand execution over to k3s.
	if s, ok := os.LookupEnv("SCT_K3_SERVER"); ok && s != "ntt" && strings.ToLower(s) != "off" {
		Command.DisableFlagParsing = true
		Command.RunE = func(cmd *cobra.Command, args []string) error {
			log.Debugln("engine: k3s")
			exe, err := exec.LookPath("k3s")
			if err != nil {
				return fmt.Errorf("k3s: %w", err)
			}
			return Exec(exe, args...)
		}
	}
}

type Suite struct {
	Suite        *ntt.Suite
	Name         string
	Sources      []string
	RuntimePaths []string
}

func NewSuite(files []string) (*Suite, error) {
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

// Run runs the given jobs in parallel.
func run(cmd *cobra.Command, args []string) error {
	defer FlushTestResults()

	ctx := context.Background()
	wg := sync.WaitGroup{}
	results := make(chan Result, MaxWorkers)

	files, ids := splitArgs(args, cmd.ArgsLenAtDash())

	suite, err := NewSuite(files)
	if err != nil {
		return err
	}

	jobs := GenerateJobs(suite, ids, MaxWorkers)

	// Start workers and process jobs in parallel.
	for i := 0; i < MaxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case job, ok := <-jobs:
					if !ok {
						return
					}
					Execute(job, results)

				case <-ctx.Done():
					results <- Result{Event: k3r.NewErrorEvent(ctx.Err())}
					return
				}
			}
		}()
	}

	// Wait for all workers to finish.
	go func() {
		wg.Wait()
		close(results)
	}()

	// Process results.
	for r := range results {
		HandleResult(r)
	}

	if errorCount > 0 {
		return fmt.Errorf("%w: %d error(s) occurred", ErrCommandFailed, errorCount)
	}
	return nil
}

// GenerateIDs emits test IDs based on given file and and id list to a channel.
func GenerateIDs(ids []string, files []string, policy string, b ntt2.Basket) <-chan string {
	policy = strings.ToLower(policy)
	policy = strings.TrimSpace(policy)
	switch {
	case len(ids) > 0:
		return ntt2.GenerateIDs(ids)
	case policy == "old":
		return ntt2.GenerateControlsWithBasket(files, b)
	default:
		return ntt2.GenerateTestsWithBasket(files, b)

	}
}

// GenerateJobs emits jobs from the given suite and ids to a job channel.
func GenerateJobs(suite *Suite, ids []string, size int) chan *Job {
	out := make(chan *Job, size)
	go func() {
		defer close(out)
		for id := range GenerateIDs(ids, suite.Sources, env.Getenv("K3_40_RUN_POLICY"), Basket) {
			out <- &Job{
				Name:  id,
				Suite: suite,
			}
		}
	}()
	return out
}

// Execute runs a single test and sends the results to the channel.
func Execute(job *Job, results chan<- Result) {
	t3xf := cache.Lookup(fmt.Sprintf("%s.t3xf", job.Suite.Name))
	test := k3r.NewTest(t3xf, job.Name)
	test.Env = append(test.Env, fmt.Sprintf("K3R_PATH=%s", strings.Join(job.Suite.RuntimePaths, ":")))
	test.Env = append(test.Env, fmt.Sprintf("LD_LIBRARY_PATH=%s", strings.Join(job.Suite.RuntimePaths, ":")))
	for event := range test.Run() {
		results <- Result{
			Job:   job,
			Test:  *test,
			Event: event,
		}
	}
}

func HandleResult(res Result) {
	switch res.Type {
	case k3r.TestStarted:
		Ledger[res.Job] = &res
		fmt.Printf("=== RUN %s\n", res.Event.Name)

	case k3r.TestTerminated:
		var d time.Duration
		if prev := Ledger[res.Job]; prev != nil {
			delete(Ledger, res.Job)
			Runs = append(Runs, results.Run{
				Name:    res.Event.Name,
				Verdict: res.Event.Verdict,
				Begin:   results.Timestamp{Time: prev.Event.Time},
				End:     results.Timestamp{Time: res.Event.Time},
			})
			d = res.Event.Time.Sub(prev.Event.Time)
		}
		line := fmt.Sprintf("--- %s %s\t(duration=%.3gs)", res.Event.Verdict, res.Event.Name, float64(d.Seconds()))
		switch res.Event.Verdict {
		case "pass":
			success.Println(line)

		case "fail", "error":
			failure.Println(line)
			atomic.AddUint64(&errorCount, 1)

		case "inconc", "none":
			warning.Println(line)
			atomic.AddUint64(&errorCount, 1)
		}

	case k3r.Error:
		fatal.Printf("+++ fatal ")
		if name := res.Event.Name; name != "" {
			fatal.Printf("%s: ", name)
		}
		fatal.Printf("%s\n", res.Event.Err.Error())
		atomic.AddUint64(&errorCount, 1)
	}
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

// splitArgs splits an argument list at pos. Pos is usually the position of '--'
// (see cobra.Command.ArgsLenAtDash).
//
// Is pos < 0, the second list will be empty
func splitArgs(args []string, pos int) ([]string, []string) {
	if pos < 0 {
		return args, []string{}
	}
	return args[:pos], args[pos:]
}
