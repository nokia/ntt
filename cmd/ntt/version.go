package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "show version",
		Run:   versionInfo,
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

func versionInfo(cmd *cobra.Command, args []string) {
	fmt.Printf("ntt %v, commit %v, built at %v\n", version, commit, date)
}
