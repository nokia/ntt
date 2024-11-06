package main

import (
	"fmt"
	"strings"

	"runtime/debug"

	"github.com/spf13/cobra"
)

var (
	VersionCommand = &cobra.Command{
		Use:   "version",
		Short: "Show version.",
		Run:   versionInfo,
	}
)

func init() {
	RootCommand.AddCommand(VersionCommand)
}

func versionInfo(cmd *cobra.Command, args []string) {
	if version == "dev" {
		info, ok := debug.ReadBuildInfo()
		if ok && strings.HasPrefix(info.Main.Version, "v") {
			version = info.Main.Version
		}
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.time":
				date = s.Value
			case "vcs.revision":
				commit = s.Value
			case "vcs.modified":
				if s.Value == "true" {
					commit += "-dirty"
				}
			}
		}
	}

	fmt.Printf("ntt %v, commit %s, built at %s\n", version, commit, date)
}
