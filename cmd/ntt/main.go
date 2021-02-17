package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/pprof"
	"strconv"
	"strings"
	"syscall"

	"github.com/nokia/ntt/internal/errors"
	"github.com/nokia/ntt/internal/session"
	"github.com/spf13/cobra"

	"github.com/nokia/ntt/internal/cmds/dump"
	"github.com/nokia/ntt/internal/cmds/langserver"
	"github.com/nokia/ntt/internal/cmds/lint"
	"github.com/nokia/ntt/internal/cmds/list"
	"github.com/nokia/ntt/internal/cmds/tags"
)

var (
	rootCmd = &cobra.Command{
		Use:   "ntt",
		Short: "ntt is a tool for managing TTCN-3 source code and tests",

		// Support for custom commands
		SilenceErrors:         true,
		SilenceUsage:          true,
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,

		Args: cobra.ArbitraryArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
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

	Verbose = false
	ShSetup = false

	version = "dev"
	commit  = "none"
	date    = "unknown"

	cpuprofile = ""
)

func init() {
	session.SharedDir = "/tmp/k3"
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&ShSetup, "sh-setup", "", false, "output test suite data for shell consumption")
	rootCmd.PersistentFlags().StringVarP(&cpuprofile, "cpuprofile", "", "", "write cpu profile to `file`")
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(dump.Command)
	rootCmd.AddCommand(langserver.Command)
	rootCmd.AddCommand(lint.Command)
	rootCmd.AddCommand(list.Command)
	rootCmd.AddCommand(tags.Command)
}

func main() {
	if s := os.Getenv("K3_DATADIR"); s == "" {
		os.Setenv("K3_DATADIR", filepath.Join(k3rootdir(), "share/k3"))
	}

	if s := os.Getenv("K3_SESSION_ID"); s == "" {
		sid, err := session.Get()
		if err != nil {
			fatal(err)
		}
		os.Setenv("K3_SESSION_ID", strconv.Itoa(sid))
	}

	err := rootCmd.Execute()
	if cpuprofile != "" {
		pprof.StopCPUProfile()
	}

	if err != nil {
		fatal(err)
	}
}

func k3rootdir() string {
	if s := os.Getenv("K3ROOT"); s != "" {
		return s
	}
	exe, _ := os.Executable()
	dir, _ := filepath.Abs(filepath.Join(filepath.Dir(exe), ".."))
	return dir
}

func fatal(err error) {
	switch err := err.(type) {
	case *exec.ExitError:
		waitStatus := err.Sys().(syscall.WaitStatus)
		os.Exit(waitStatus.ExitStatus())
	case errors.ErrorList:
		errors.PrintError(os.Stderr, err)
	default:
		fmt.Fprintln(os.Stderr, err.Error())
	}

	os.Exit(1)
}
