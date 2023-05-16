package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/nokia/ntt/control"
	"github.com/nokia/ntt/control/k3r"
	"github.com/nokia/ntt/control/pool"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/results"
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/doc"
	"github.com/nokia/ntt/ttcn3/syntax"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	RunCommand = &cobra.Command{
		Use:   "run [ <path>... ] [ -- <test id>... ]",
		Short: "Build and run test suite",
		Long: `Build and run a test suite.

The ntt run command first builds a test executable using the files or
directories passed as first argument list.
The test executable is then run with the tests specified as second argument
list. If no ids are specified, ntt run will run all tests in the test suite.

Running control functions is supported. For example:

        ntt run -- test.A test.control test.B


Test baskets are also supported (see "ntt help list"). Bellow example will run
all tests with @stable-tag:

	NTT_LIST_BASKETS=stable ntt run

`,

		RunE: runTests,
	}

	RunAllTests bool
	MaxWorkers  int
	MaxFail     int
	errorCount  uint64
	OutputDir   string

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
		case "done":
			return color.New()
		default:
			return ColorFailure
		}
	}

	ErrCommandFailed = fmt.Errorf("command failed")
)

func init() {
	flags := RunCommand.Flags()
	flags.AddFlagSet(BasketFlags())
	flags.IntVarP(&MaxWorkers, "jobs", "j", runtime.NumCPU(), "Allow N test in parallel (default: number of CPU cores")
	flags.IntVar(&MaxFail, "max-fail", 0, "Stop after N failures")
	flags.StringVarP(&OutputDir, "output-dir", "o", "", "store test artefacts in DIR/ID")
	flags.BoolVarP(&outputProgress, "progress", "P", false, "show progress")
	flags.BoolVarP(&RunAllTests, "all-tests", "a", false, "run all tests instead of control parts")
	flags.StringSliceVarP(&testsFiles, "tests-file", "t", nil, "read tests from FILE. If this option is used multiple times all contained tests will be executed in that order. When FILE is '-', read standard input")
}

// Run runs the given jobs in parallel.
func runTests(cmd *cobra.Command, args []string) error {
	ctx, cancel := WithSignalHandler(context.Background())
	defer cancel()

	// Assure that that project binaries are up-to-date, before we execute the tests.
	if err := project.Build(Project); err != nil {
		return fmt.Errorf("building test suite failed: %w", err)
	}

	_, ids := splitArgs(args, cmd.ArgsLenAtDash())
	jobs, err := JobQueue(ctx, cmd.Flags(), Project, testsFiles, ids, RunAllTests)
	if err != nil {
		return err
	}

	var runs []results.Run
	os.Remove(Project.ResultsFile)
	defer func() {
		db := &results.DB{
			Version: "1",
			Sessions: []results.Session{
				{
					Id:              "1",
					MaxJobs:         MaxWorkers,
					ExpectedVerdict: "pass",
					Runs:            runs,
				},
			},
		}
		b, err := json.MarshalIndent(db, "", "  ")
		if err != nil {
			return
		}
		err = ioutil.WriteFile(Project.ResultsFile, b, 0644)
	}()

	running := make(map[*control.Job]time.Time)
	runner, err := pool.NewRunner(
		pool.MaxWorkers(MaxWorkers),
		pool.WithFactory(k3r.Factory(jobs)),
	)
	if err != nil {
		return err
	}

	for e := range runner.Run(ctx) {
		log.Debugf("result: event=%#v\n", e)

		switch e := e.(type) {
		case control.LogEvent:
			log.Debugln(e.Text)

		case control.StartEvent:
			running[e.Job] = e.Time()
			if Format() == "text" {
				ColorStart.Printf("=== RUN %s\n", e.Name)
			}

		case control.TickerEvent:
			if Format() == "text" {
				for job := range running {
					ColorRunning.Printf("... active %s\n", job.Name)
				}
			}

		case control.StopEvent:
			if e.Verdict != "pass" && e.Verdict != "done" {
				errorCount++
			}
			run := results.Run{
				Name:       e.Name,
				Verdict:    e.Verdict,
				Begin:      results.Timestamp{Time: running[e.Job]},
				End:        results.Timestamp{Time: e.Time()},
				WorkingDir: e.Job.Dir,
			}
			runs = append(runs, run)
			printRun(run)

		case control.ErrorEvent:
			errorCount++
			job := control.UnwrapJob(e)
			if job == nil {
				ColorFatal.Println("+++ fatal error: " + e.Err.Error())
				break
			}
			run := results.Run{
				Name:       job.Name,
				Verdict:    "fatal",
				Begin:      results.Timestamp{Time: running[job]},
				End:        results.Timestamp{Time: e.Time()},
				WorkingDir: job.Dir,
			}
			printRun(run)

		default:
			panic(fmt.Sprintf("event type %T not implemented", e))
		}

		if MaxFail > 0 && errorCount >= uint64(MaxFail) {
			ColorFatal.Println("+++ fatal too many errors. Exiting.")
			cancel()
			break
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("%w: %d error(s) occurred", ErrCommandFailed, errorCount)
	}

	return nil

}

func JobQueue(ctx context.Context, flags *pflag.FlagSet, conf *project.Config, testsFiles []string, tests []string, allTests bool) (<-chan *control.Job, error) {

	basket, err := NewBasketWithFlags("run", flags)
	if err != nil {
		return nil, fmt.Errorf("creating basket failed: %w", err)
	}
	if err := basket.LoadFromEnvOrConfig(conf, "NTT_LIST_BASKETS"); err != nil {
		return nil, fmt.Errorf("loading baskets failed: %w", err)
	}

	var tsts []string
	for _, f := range testsFiles {
		t, err := readTestsFromFile(f)
		if err != nil {
			return nil, fmt.Errorf("reading tests from file %s failed: %w", f, err)
		}
		tsts = append(tsts, t...)
	}
	srcs, err := fs.TTCN3Files(conf.Sources...)
	if err != nil {
		return nil, err
	}
	needTests := len(tests) == 0 && len(testsFiles) == 0
	m := sync.Map{}
	t := make([][]string, len(srcs))
	wg := sync.WaitGroup{}
	wg.Add(len(srcs))
	start := time.Now()
	for i, src := range srcs {
		go func(src string, i int) {
			defer wg.Done()
			var (
				mod         string
				modLvl, lvl int
			)
			root := ttcn3.ParseFile(src)
			root.Inspect(func(n syntax.Node) bool {
				if n == nil {
					if lvl == modLvl {
						mod = ""
						modLvl = 0
					}
					lvl--
				} else {
					lvl++
				}
				switch n := n.(type) {
				case *syntax.Module:
					mod = n.Name.String()
					modLvl = lvl
					return true
				case *syntax.FuncDecl:
					if !n.IsTest() && !n.IsControl() {
						return false
					}
					name := ttcn3.JoinNames(mod, n.Name.String())
					m.Store(name, n)
					if needTests {
						if n.IsTest() && allTests || n.IsControl() && !allTests {
							t[i] = append(t[i], name)
						}
					}
					return false
				case *syntax.ControlPart:
					name := ttcn3.JoinNames(mod, n.Name.String())
					m.Store(name, n)
					if needTests && !allTests {
						t[i] = append(t[i], name)
					}
					return false

				default:
					return true
				}
			})
		}(src, i)
	}
	wg.Wait()
	log.Debugf("Scanned all tests in %s.\n", time.Since(start))

	testPlan := append(tsts, tests...)
	if needTests {
		for _, tests := range t {
			testPlan = append(testPlan, tests...)
		}
	}

	out := make(chan *control.Job)
	go func() {
		defer close(out)
		names := make(map[string]int)
		for _, name := range testPlan {
			var tags [][]string
			if def, ok := m.Load(name); ok {
				tags = doc.FindAllTags(syntax.Doc(def.(syntax.Node)))
			}
			if !basket.Match(name, tags) {
				continue
			}
			configs, err := conf.TestConfigs(name)
			if err != nil {
				log.Verbose(err.Error())
				continue
			}
			if len(configs) == 0 {
				log.Verbosef("no config for %s", name)
				continue
			}

			for _, tc := range configs {
				id := fmt.Sprintf("%s-%d", name, names[name])
				names[name]++

				job := &control.Job{
					ID:         id,
					Name:       name,
					Config:     conf,
					Dir:        OutputDir,
					Timeout:    tc.Timeout.Duration,
					ModulePars: tc.Parameters,
				}

				select {
				case out <- job:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out, nil
}

// EntryPoints returns controls parts of the given TTCN-3 source file. When tests is true, it returns all testcases instead.
func EntryPoints(file string, tests bool) []*ttcn3.Node {
	tree := ttcn3.ParseFile(file)
	if tests {
		return tree.Tests()
	}
	return tree.Controls()
}

func printRun(r results.Run) {
	switch Format() {
	case "plain":
		c := Colors(r.Verdict)
		c.Printf("%s\t%s\t%.4f\n", r.Verdict, r.Name, float64(r.Duration().Seconds()))
	case "text":
		switch {
		case r.Verdict == "fatal":
			ColorFatal.Printf("+++ fatal %s\t(%s)\n", r.Name, r.Reason)
		default:
			c := Colors(r.Verdict)
			c.Printf("--- %s %s\t(duration=%.2fs)\n", r.Verdict, r.Name, float64(r.Duration().Seconds()))
		}
	case "json":
		b, err := json.Marshal(r)
		if err != nil {
			panic(fmt.Sprintf("cannot marshal run: %v", r))
		}
		fmt.Println(string(b))
	}
}

// WithSignalHandler adds a signal handler for ^C to the context.
func WithSignalHandler(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx2, cancel := context.WithCancel(context.Background())

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		select {
		case <-signalChan: // first signal, cancel context
			cancel()
		case <-ctx.Done():
		}
		<-signalChan // second signal, hard exit
		os.Exit(2)
	}()
	return ctx2, func() {
		signal.Stop(signalChan)
		cancel()
	}
}

func readTestsFromFile(path string) ([]string, error) {
	var (
		lines []byte
		err   error
	)
	if path == "-" {
		lines, err = ioutil.ReadAll(os.Stdin)
	} else {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		lines, err = ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}

	}
	var tests []string
	for _, line := range strings.Split(string(lines), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		tests = append(tests, line)
	}
	return tests, err
}
