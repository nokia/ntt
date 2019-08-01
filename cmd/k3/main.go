package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/ttcn3/syntax"
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
	env.SetPrefix("k3")
	env.AddPath("${PWD}")
	env.AddPath("${HOME}/.config/k3")
	env.AddPath("${HOME}/.k3/config")

	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.AddCommand(listCmd, showCmd)
}

func main() {
	if err := env.ReadEnvFiles(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// Bash Workaround: If a custom command is a shell script, array variables
	// like K3_IMPORTS cannot be exported. Such scripts export _K3_SOURCES, ...
	// instead.
	copyEnv("_K3_SOURCES", "K3_SOURCES")
	copyEnv("_K3_IMPORTS", "K3_IMPORTS")
	copyEnv("_K3_TTCN3_FILES", "K3_TTCN3_FILES")

	if err := rootCmd.Execute(); err != nil {
		switch err := err.(type) {
		case *exec.ExitError:
			waitStatus := err.Sys().(syscall.WaitStatus)
			os.Exit(waitStatus.ExitStatus())
		}
		syntax.PrintError(os.Stderr, err)
		os.Exit(1)
	}
}

func copyEnv(from, to string) {
	if s := os.Getenv(from); s != "" {
		os.Setenv(to, s)
	}
}
