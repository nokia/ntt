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
	"github.com/nokia/ntt"
	"github.com/nokia/ntt/internal/cache"
	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/proc"
	"github.com/nokia/ntt/internal/results"
	k3r "github.com/nokia/ntt/k3/run"
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
	stamps    = make(map[string]time.Time)
	Basket, _ = ntt.NewBasket("default")

	ResultsFile = cache.Lookup("test_results.json")
	TickerTime  = time.Second * 30
	Start       = time.Now()
)

func init() {
	flags := RunCommand.Flags()
	flags.AddFlagSet(ntt.BasketFlags())
	flags.IntVarP(&MaxWorkers, "jobs", "j", runtime.NumCPU(), "Allow N test in parallel (default: number of CPU cores")
	flags.IntVar(&MaxFail, "max-fail", 0, "Stop after N failures")
	flags.StringVarP(&OutputDir, "output-dir", "o", "", "store test artefacts in DIR/ID")
	flags.BoolVarP(&outputProgress, "progress", "P", false, "show progress")
	flags.StringVarP(&testsFile, "tests-file", "t", "", "read tests from FILE. When FILE is '-', read standard input")
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
	Basket, err = ntt.NewBasketWithFlags("list", cmd.Flags())
	Basket.LoadFromEnv("NTT_LIST_BASKETS")
	if err != nil {
		return err
	}

	files, ids := splitArgs(args, cmd.ArgsLenAtDash())
	suite, err := ntt.NewSuite(files...)
	if err != nil {
		return err
	}

	if testsFile != "" {
		tests, err := readTestsFromFile(testsFile)
		if err != nil {
			return err
		}
		ids = append(tests, ids...)
	}

	ledger := ntt.NewLedger(MaxWorkers)
	jobs := GenerateJobs(ctx, suite, ids, MaxWorkers, ledger)

	if s, ok := os.LookupEnv("SCT_K3_SERVER"); ok && s != "ntt" && strings.ToLower(s) != "off" {
		return k3sRun(ctx, files, jobs)
	}
	return nttRun(ctx, jobs, ledger)
}

func nttRun(ctx context.Context, jobs <-chan *ntt.Job, ledger *ntt.Ledger) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	os.Remove(cache.Lookup("test_results.json"))
	defer FlushTestResults()

	// Execute the jobs in parallel and collect the results.
	ticker := time.NewTicker(TickerTime)
	ressults := ledger.Run(ctx, jobs)
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
			if Format() == "text" {
				for _, job := range ledger.Jobs() {
					ColorRunning.Printf("... active %s\n", job.Name)
				}
			}
		}
	}
	if errorCount > 0 {
		return fmt.Errorf("%w: %d error(s) occurred", ErrCommandFailed, errorCount)
	}
	return nil
}

func k3sRun(ctx context.Context, files []string, jobs <-chan *ntt.Job) error {
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

func k3sJobs(jobs <-chan *ntt.Job) io.Reader {
	var ids []string
	for j := range jobs {
		ids = append(ids, j.Name)
	}
	return strings.NewReader(strings.Join(ids, "\n"))
}

// GenerateIDs emits test IDs based on given file and and id list to a channel.
func GenerateIDs(ctx context.Context, ids []string, files []string, policy string, b ntt.Basket) <-chan string {
	policy = strings.ToLower(policy)
	policy = strings.TrimSpace(policy)
	switch {
	case len(ids) > 0:
		return ntt.GenerateIDsWithContext(ctx, ids...)
	case policy == "old":
		return ntt.GenerateControlsWithContext(ctx, b, files...)
	default:
		return ntt.GenerateTestsWithContext(ctx, b, files...)

	}
}

// GenerateJobs emits jobs from the given suite and ids to a job channel.
func GenerateJobs(ctx context.Context, suite *ntt.Suite, ids []string, size int, ledger *ntt.Ledger) chan *ntt.Job {
	out := make(chan *ntt.Job, size)
	go func() {
		defer close(out)
		i := 0
		for id := range GenerateIDs(ctx, ids, suite.Manifest.Sources, env.Getenv("K3_40_RUN_POLICY"), Basket) {
			i++
			out <- ledger.NewJob(id, suite)
		}
		log.Debugf("Generating %d jobs done.\n", i)
	}()
	return out
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
