package run

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/nokia/ntt/internal/cache"
	"github.com/nokia/ntt/internal/cmds/build"
	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/k3"
	k3r "github.com/nokia/ntt/k3/run"
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/spf13/cobra"
)

type Job struct {
	Name string
	*ntt.Suite
	SuiteName    string
	RuntimePaths []string
}

type Result struct {
	k3r.Test
	k3r.Event
}

func (r *Result) Log() {
	switch {
	case OutputJSON:

	}
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

	ErrNoSuchTest = fmt.Errorf("no such test")
)

func init() {
	flags := Command.Flags()
	flags.IntVarP(&MaxWorkers, "jobs", "j", 1, "number of parallel tests")
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
	ctx := context.Background()
	wg := sync.WaitGroup{}
	jobs := make(chan Job, MaxWorkers)
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
	return nil
}

// EnqueueJobs creates the job load and sends it to the jobs channel.
func EnqueueJobs(files []string, ids []string, jobs chan Job) error {
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
			jobs <- Job{Name: id, Suite: suite, SuiteName: name, RuntimePaths: paths}
		}
		return nil
	}
	srcs, err := suite.Sources()
	if err != nil {
		return err
	}
	for _, src := range srcs {
		tree := ttcn3.ParseFile(src)
		for _, n := range tree.Funcs() {
			if n := n.Node.(*ast.FuncDecl); n.IsTest() {
				mod := ast.Name(tree.ModuleOf(n))
				jobs <- Job{
					Name:         fmt.Sprintf("%s.%s", mod, n.Name.String()),
					Suite:        suite,
					SuiteName:    name,
					RuntimePaths: paths,
				}
			}
		}

	}
	return nil
}

// Execute runs a single test and sends the results to the channel.
func Execute(job Job, results chan<- Result) {
	t3xf := cache.Lookup(fmt.Sprintf("%s.t3xf", job.SuiteName))
	test := k3r.NewTest(t3xf, job.Name)
	test.Env = append(test.Env, fmt.Sprintf("K3R_PATH=%s", strings.Join(job.RuntimePaths, ":")))
	test.Env = append(test.Env, fmt.Sprintf("LD_LIBRARY_PATH=%s", strings.Join(job.RuntimePaths, ":")))
	for event := range test.Run() {
		results <- Result{Test: *test, Event: event}
	}
}

func HandleResult(res Result) {
	switch res.Type {
	case k3r.TestStarted:
		fmt.Printf("=== RUN %s\n", res.Event.Name)
	case k3r.TestTerminated:
		switch res.Event.Verdict {
		case "fail", "error":
			color.Set(color.FgRed, color.Bold)
		case "inconc", "none":
			color.Set(color.FgYellow, color.Bold)
		}
		fmt.Printf("--- %s %s\n", res.Event.Verdict, res.Event.Name)
		color.Unset()
	}
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
