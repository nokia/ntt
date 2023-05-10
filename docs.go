package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var DocsCommand = &cobra.Command{
	Use:               "docs",
	Hidden:            true,
	DisableAutoGenTag: true,
	Args:              cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := "./"
		if len(args) > 0 {
			out = args[0]
		}
		if err := os.MkdirAll(out, os.ModePerm); err != nil {
			return err
		}
		return doc.GenMarkdownTree(RootCommand, out)
	},
}

func init() {
	RootCommand.AddCommand(DocsCommand)
}
