package main

import (
	"os"

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
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.AddCommand(listCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		syntax.PrintError(os.Stderr, err)
		os.Exit(1)
	}
}
