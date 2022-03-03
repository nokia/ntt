package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show version.",
		Run:   versionInfo,
	}
)

func init() {
	Command.AddCommand(versionCmd)
}

func versionInfo(cmd *cobra.Command, args []string) {
	fmt.Printf("ntt %v, commit %v, built at %v\n", version, commit, date)
}
