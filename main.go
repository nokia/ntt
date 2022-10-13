package main

import (
	stderrors "errors"
	"fmt"
	"os"
	"os/exec"
	"runtime/pprof"
	"strconv"
	"syscall"

	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/errors"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/internal/proc"
	"github.com/nokia/ntt/internal/session"
	"github.com/nokia/ntt/project"
	"github.com/spf13/cobra"
)

var (
	RootCommand = &cobra.Command{
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

			if chdir != "" {
				if err := os.Chdir(chdir); err != nil {
					return fmt.Errorf("chdir: %w", err)
				}
				log.Debugf("chdir: %s", chdir)
			}
			if cpuprofile != "" {
				f, err := os.Create(cpuprofile)
				if err != nil {
					return err
				}
				if err := pprof.StartCPUProfile(f); err != nil {
					return err
				}
			}

			// Skip opening the project if we're running a custom command or version.
			if cmd.Use == "ntt" || cmd.Use == "version" || cmd.Use == "stdout" {
				// first arg is either an external subkommand of the form
				// k3-Arg[0] or ntt-Arg[0] or unknown
				return nil
			}

			files, _ := splitArgs(args, cmd.ArgsLenAtDash())
			p, err := project.Open(files...)
			if err != nil {
				return err
			}
			Project = p
			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && args[0][0] != '-' {
				if path, err := exec.LookPath("ntt-" + args[0]); err == nil {
					return proc.Exec(path, args[1:]...)
				}
				if path, err := exec.LookPath("k3-" + args[0]); err == nil {
					return proc.Exec(path, args[1:]...)
				}
				return fmt.Errorf("unknown command: %s", args[0])
			}

			if err := cmd.Flags().Parse(args); err != nil {
				return err
			}

			if interactive, _ := cmd.Flags().GetBool("interactive"); interactive {
				return repl()
			}

			return cmd.Help()
		},
	}

	verbose        int
	ShSetup        bool
	outputQuiet    bool
	outputJSON     bool
	outputPlain    bool
	outputProgress bool
	testsFiles     []string
	chdir          string

	version = "dev"
	commit  = "none"
	date    = "unknown"

	cpuprofile = ""

	Project *project.Config
)

func init() {
	if s := os.Getenv("K3_SESSION_ID"); s == "" {
		sid, err := session.Get()
		if err != nil {
			fatal(err)
		}
		os.Setenv("K3_SESSION_ID", strconv.Itoa(sid))
	}
	root := RootCommand
	flags := root.PersistentFlags()
	flags.CountVarP(&verbose, "verbose", "v", "verbose output")
	flags.BoolVarP(&outputQuiet, "quiet", "q", false, "quiet output")
	flags.BoolVarP(&outputJSON, "json", "", false, "output in JSON format")
	flags.BoolVarP(&outputPlain, "plain", "", false, "output in plain format (for grep and awk)")
	flags.StringVarP(&cpuprofile, "cpuprofile", "", "", "write cpu profile to `file`")
	flags.StringVarP(&chdir, "chdir", "C", "", "change to DIR before doing anything else")

	RootCommand.Flags().BoolP("interactive", "i", false, "run in interactive mode")

	root.AddCommand(BuildCommand)
	root.AddCommand(CompileCommand)
	root.AddCommand(DumpCommand)
	root.AddCommand(FormatCommand)
	root.AddCommand(LangserverCommand)
	root.AddCommand(LintCommand)
	root.AddCommand(ListCommand)
	root.AddCommand(LocateFileCommand)
	root.AddCommand(ReportCommand)
	root.AddCommand(RunCommand)
	root.AddCommand(ShowCommand)
	root.AddCommand(TagsCommand)

	ShowCommand.PersistentFlags().BoolVarP(&ShSetup, "sh", "", false, "output test suite data for shell consumption")
}

func Format() string {
	switch {
	case outputQuiet:
		return "quiet"
	case outputPlain:
		return "plain"
	case outputJSON:
		return "json"
	case outputProgress:
		return "progress"
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
	case outputQuiet:
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
	err := RootCommand.Execute()
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
		if !stderrors.Is(err, ErrCommandFailed) {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}

	os.Exit(1)
}
