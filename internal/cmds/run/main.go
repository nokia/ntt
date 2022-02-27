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
	"github.com/nokia/ntt/internal/cache"
	"github.com/nokia/ntt/internal/cmds/build"
	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/internal/results"
	"github.com/nokia/ntt/k3"
	k3r "github.com/nokia/ntt/k3/run"
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/spf13/cobra"
)

type Job struct {
	Name      string
	Iteration int

	// "Precalculated" suite configuration.
	SuiteName    string
	RuntimePaths []string
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

	Runs   []results.Run
	Ledger = make(map[*Job]*Result)
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

// Run runs the given jobs in parallel.
func run(cmd *cobra.Command, args []string) error {
	defer FlushTestResults()

	ctx := context.Background()
	wg := sync.WaitGroup{}
	jobs := make(chan *Job, MaxWorkers)
	results := make(chan Result, MaxWorkers)

	// Enqueue jobs.
	go func() {
		defer close(jobs)
		files, ids := splitArgs(args, cmd.ArgsLenAtDash())
		if err := EnqueueJobs(files, ids, jobs); err != nil {
			results <- Result{Event: k3r.NewErrorEvent(err)}
		}
	}()

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

// EnqueueJobs creates the job load and sends it to the jobs channel.
func EnqueueJobs(files []string, ids []string, jobs chan *Job) error {
	suite, err := ntt.NewFromArgs(files...)
	if err != nil {
		return err
	}

	name, err := suite.Name()
	if err != nil {
		return fmt.Errorf("test suite name: %w", err)
	}

	paths, err := runtimePaths(suite)

	if len(ids) > 0 {
		for _, id := range ids {
			jobs <- &Job{Name: id, SuiteName: name, RuntimePaths: paths}
		}
		return nil
	}
	srcs, err := suite.Sources()
	if err != nil {
		return err
	}
	if policy := os.Getenv("K3_40_RUN_POLICY"); policy == "old" {
		enqueueControls(jobs, name, paths, srcs)
	} else {
		enqueueTests(jobs, name, paths, srcs)
	}
	return nil
}

func enqueueTests(jobs chan *Job, suiteName string, paths []string, srcs []string) {
	for _, src := range srcs {
		tree := ttcn3.ParseFile(src)
		for _, n := range tree.Funcs() {
			if n := n.Node.(*ast.FuncDecl); n.IsTest() {
				mod := ast.Name(tree.ModuleOf(n))
				jobs <- &Job{
					Name:         fmt.Sprintf("%s.%s", mod, n.Name.String()),
					SuiteName:    suiteName,
					RuntimePaths: paths,
				}
			}
		}

	}
}

func enqueueControls(jobs chan *Job, suiteName string, paths []string, srcs []string) {
	for _, src := range srcs {
		tree := ttcn3.ParseFile(src)
		for _, n := range tree.Controls() {
			mod := ast.Name(tree.ModuleOf(n.Node))
			jobs <- &Job{
				Name:         fmt.Sprintf("%s.control", mod),
				SuiteName:    suiteName,
				RuntimePaths: paths,
			}
		}

	}
}

// Execute runs a single test and sends the results to the channel.
func Execute(job *Job, results chan<- Result) {
	t3xf := cache.Lookup(fmt.Sprintf("%s.t3xf", job.SuiteName))
	test := k3r.NewTest(t3xf, job.Name)
	test.Env = append(test.Env, fmt.Sprintf("K3R_PATH=%s", strings.Join(job.RuntimePaths, ":")))
	test.Env = append(test.Env, fmt.Sprintf("LD_LIBRARY_PATH=%s", strings.Join(job.RuntimePaths, ":")))
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
