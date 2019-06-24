package main

import (
	"fmt"
	"os"

	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/ttcn3/syntax"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "k3",
		Short: "k3 is a tool for managing TTCN-3 source code and tests",
		Long:  "",
	}

	Verbose = false
)

func init() {
	env.SetPrefix("k3")
	env.AddPath("${PWD}")
	env.AddPath("${HOME}/.config/k3")

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
