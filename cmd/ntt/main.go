package main

import (
	stderrors "errors"
	"fmt"
	"os"
	"os/exec"
	"runtime/pprof"
	"strconv"
	"strings"
	"syscall"

	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/errors"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/session"
	"github.com/nokia/ntt/k3"
	"github.com/spf13/cobra"

	"github.com/nokia/ntt/internal/cmds/build"
	"github.com/nokia/ntt/internal/cmds/dump"
	"github.com/nokia/ntt/internal/cmds/langserver"
	"github.com/nokia/ntt/internal/cmds/lint"
	"github.com/nokia/ntt/internal/cmds/list"
	"github.com/nokia/ntt/internal/cmds/locate_file"
	"github.com/nokia/ntt/internal/cmds/report"
	"github.com/nokia/ntt/internal/cmds/run"
	"github.com/nokia/ntt/internal/cmds/tags"
)

var (
	Command = &cobra.Command{
		Use:   "ntt",
		Short: "ntt is a tool for managing TTCN-3 source code and tests",

		// Support for custom commands
		SilenceErrors:         true,
		SilenceUsage:          true,
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,

		Args: cobra.ArbitraryArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			log.SetGlobalLevel(Verbosity())

			if cpuprofile != "" {
				f, err := os.Create(cpuprofile)
				if err != nil {
					return err
				}
				if err := pprof.StartCPUProfile(f); err != nil {
					return err
				}
			}
			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 || strings.HasPrefix(args[0], "-") {
				if _, ok := os.LookupEnv("NTT_ENABLE_REPL"); ok {
					return repl()
				}
				cmd.Help()
				return nil
			}

			if path, err := exec.LookPath("ntt-" + args[0]); err == nil {
				return Execute(path, args[1:]...)
			}

			if path, err := exec.LookPath("k3-" + args[0]); err == nil {
				return Execute(path, args[1:]...)
			}

			err := fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
			cmd.Println("Error:", err.Error())
			cmd.Printf("Run '%v --help' for usage.\n", cmd.CommandPath())
			return err
		},
	}

	verbose     int
	quiet       bool
	ShSetup     bool
	outputJSON  bool
	outputPlain bool

	version = "dev"
	commit  = "none"
	date    = "unknown"

	cpuprofile = ""
)

func init() {
	session.SharedDir = "/tmp/k3"
	Command.PersistentFlags().CountVarP(&verbose, "verbose", "v", "verbose output")
	Command.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet output")
	Command.PersistentFlags().BoolVarP(&outputJSON, "json", "", false, "output in JSON format")
	Command.PersistentFlags().BoolVarP(&outputPlain, "plain", "", false, "output in plain format (for grep and awk)")
	Command.PersistentFlags().StringVarP(&cpuprofile, "cpuprofile", "", "", "write cpu profile to `file`")
	Command.AddCommand(showCmd)

	showCmd.PersistentFlags().BoolVarP(&ShSetup, "sh", "", false, "output test suite data for shell consumption")
	Command.AddCommand(dump.Command)
	Command.AddCommand(locate_file.Command)
	Command.AddCommand(langserver.Command)
	Command.AddCommand(lint.Command)
	Command.AddCommand(list.Command)
	Command.AddCommand(tags.Command)
	Command.AddCommand(report.Command)
	Command.AddCommand(build.Command)
	Command.AddCommand(run.Command)

}

func Format() string {
	switch {
	case outputPlain:
		return "plain"
	case outputJSON:
		return "json"
	default:
		return "text"
	}
}

func Verbosity() log.Level {
	switch {
	case env.Getenv("NTT_TRACE") != "":
		return log.TraceLevel
	case env.Getenv("NTT_DEBUG") != "":
		return log.DebugLevel
	case quiet:
		return log.DisabledLevel
	default:
		lvl := log.PrintLevel + log.Level(verbose)
		if lvl > log.TraceLevel {
			lvl = log.TraceLevel
		}
		return lvl
	}
}

func main() {
	defer log.Close()

	if s := k3.DataDir(); s != "" {
		os.Setenv("K3_DATADIR", s)
	}
	if s := os.Getenv("K3_SESSION_ID"); s == "" {
		sid, err := session.Get()
		if err != nil {
			fatal(err)
		}
		os.Setenv("K3_SESSION_ID", strconv.Itoa(sid))
	}

	err := Command.Execute()
	if cpuprofile != "" {
		pprof.StopCPUProfile()
	}

	if err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	switch err := err.(type) {
	case *exec.ExitError:
		waitStatus := err.Sys().(syscall.WaitStatus)
		os.Exit(waitStatus.ExitStatus())
	case errors.ErrorList:
		errors.PrintError(os.Stderr, err)
	default:
		// Run command has its own error logging.
		if !stderrors.Is(err, run.ErrCommandFailed) {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}

	os.Exit(1)
}
