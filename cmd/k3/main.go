package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
	if err := rootCmd.Execute(); err != nil {
		syntax.PrintError(os.Stderr, err)
		os.Exit(1)
	}
}
