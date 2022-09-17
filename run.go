package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/nokia/ntt/internal/cache"
	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/proc"
	"github.com/nokia/ntt/internal/results"
	"github.com/nokia/ntt/project"
	"github.com/nokia/ntt/tests"
	"github.com/nokia/ntt/tests/runner"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/doc"
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

		RunE: runTests,
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

	DefaultBasket, _ = NewBasket("default")

	ResultsFile = cache.Lookup("test_results.json")
	Start       = time.Now()
)

func init() {
	flags := RunCommand.Flags()
	flags.AddFlagSet(BasketFlags())
	flags.IntVarP(&MaxWorkers, "jobs", "j", runtime.NumCPU(), "Allow N test in parallel (default: number of CPU cores")
	flags.IntVar(&MaxFail, "max-fail", 0, "Stop after N failures")
	flags.StringVarP(&OutputDir, "output-dir", "o", "", "store test artefacts in DIR/ID")
	flags.BoolVarP(&outputProgress, "progress", "P", false, "show progress")
	flags.StringVarP(&testsFile, "tests-file", "t", "", "read tests from FILE. When FILE is '-', read standard input")
}

// Run runs the given jobs in parallel.
func runTests(cmd *cobra.Command, args []string) error {

	ctx, cancel := WithSignalHandler(context.Background())
	defer cancel()

	if err := project.Build(Project); err != nil {
		return fmt.Errorf("building test suite failed: %w", err)
	}

	var err error
	DefaultBasket, err = NewBasketWithFlags("list", cmd.Flags())
	DefaultBasket.LoadFromEnvOrConfig(Project, "NTT_LIST_BASKETS")
	if err != nil {
		return err
	}

	files, ids := splitArgs(args, cmd.ArgsLenAtDash())
	if testsFile != "" {
		tests, err := readTestsFromFile(testsFile)
		if err != nil {
			return err
		}
		ids = append(tests, ids...)
	}

	srcs, err := fs.TTCN3Files(Project.Sources...)
	if err != nil {
		return err
	}

	jobs := make(chan *tests.Job, MaxWorkers)
	go func() {
		defer close(jobs)

		i := 0
		for id := range GenerateIDs(ctx, ids, srcs, env.Getenv("K3_40_RUN_POLICY"), DefaultBasket) {
			i++
			job := &tests.Job{
				Name:   id,
				Config: Project,
				Dir:    OutputDir,
			}
			log.Debugf("new job: name=%s\n", id)
			jobs <- job
		}
		log.Debugf("Generating %d jobs done.\n", i)
	}()

	if err != nil {
		return err
	}

	if s, ok := os.LookupEnv("SCT_K3_SERVER"); ok && s != "ntt" && strings.ToLower(s) != "off" {
		return k3sRun(ctx, files, jobs)
	}
	return nttRun(ctx, jobs)
}

func nttRun(ctx context.Context, jobs <-chan *tests.Job) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var runs []results.Run
	os.Remove(cache.Lookup("test_results.json"))
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
		err = ioutil.WriteFile(cache.Lookup("test_results.json"), b, 0644)
	}()

	running := make(map[*tests.Job]time.Time)
	for e := range runner.New(MaxWorkers, jobs).Run(ctx) {
		log.Debugf("result: event=%#v\n", e)

		switch e := e.(type) {

		case tests.StartEvent:
			running[e.Job] = e.Time()
			if Format() == "text" {
				ColorStart.Printf("=== RUN %s\n", e.Name)
			}

		case tests.TickerEvent:
			if Format() == "text" {
				for job := range running {
					ColorRunning.Printf("... active %s\n", job.Name)
				}
			}

		case tests.StopEvent:
			if e.Verdict != "pass" {
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

		case tests.ErrorEvent:
			errorCount++
			job := tests.UnwrapJob(e)
			if job == nil {
				ColorFatal.Print("+++ fatal error: " + e.Err.Error())
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
			ColorFatal.Print("+++ fatal too many errors. Exiting.\n")
			cancel()
			break
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("%w: %d error(s) occurred", ErrCommandFailed, errorCount)
	}

	return nil
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

func k3sRun(ctx context.Context, files []string, jobs <-chan *tests.Job) error {
	args := []string{
		"--no-summary",
		fmt.Sprintf("--results-file=%s", ResultsFile),
		fmt.Sprintf("-j%d", MaxWorkers),
	}
	if s := env.Getenv("K3SFLAGS"); s != "" {
		args = append(args, strings.Fields(s)...)
	}
	args = append(args, files...)
	k3s := proc.CommandContext(ctx, "k3s", args...)
	k3s.Stdin = k3sJobs(jobs)
	k3s.Stdout = os.Stdout
	k3s.Stderr = os.Stderr
	log.Verboseln("+", k3s.String())
	return k3s.Run()
}

func k3sJobs(jobs <-chan *tests.Job) io.Reader {
	var ids []string
	for j := range jobs {
		ids = append(ids, j.Name)
	}
	return strings.NewReader(strings.Join(ids, "\n"))
}

// GenerateIDs emits test IDs based on given file and and id list to a channel.
func GenerateIDs(ctx context.Context, ids []string, files []string, policy string, b Basket) <-chan string {
	policy = strings.ToLower(policy)
	policy = strings.TrimSpace(policy)
	switch {
	case len(ids) > 0:
		return GenerateIDsWithContext(ctx, ids...)
	case policy == "old":
		return GenerateControlsWithContext(ctx, b, files...)
	default:
		return GenerateTestsWithContext(ctx, b, files...)

	}
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
