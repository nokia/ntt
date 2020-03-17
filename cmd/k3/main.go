package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/nokia/ntt/internal/errors"
	"github.com/nokia/ntt/internal/session"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "k3",
		Short: "k3 is a tool for managing TTCN-3 source code and tests",

		// Support for custom commands
		SilenceErrors:         true,
		SilenceUsage:          true,
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,
		Args:                  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 || strings.HasPrefix(args[0], "-") {
				cmd.Help()
				return nil
			}

			if path, err := exec.LookPath("k3-" + args[0]); err == nil {
				cmd := exec.Command(path, args[1:]...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				return cmd.Run()
			}

			err := fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
			cmd.Println("Error:", err.Error())
			cmd.Printf("Run '%v --help' for usage.\n", cmd.CommandPath())
			return err
		},
	}

	Verbose = false
)

func init() {
	session.SharedDir = "/tmp/k3"
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.AddCommand(showCmd)
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

	if err := rootCmd.Execute(); err != nil {
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
