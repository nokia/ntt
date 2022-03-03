package main

import (
	"fmt"

	"github.com/nokia/ntt/internal/fs"
	"github.com/spf13/cobra"
)

var (
	LocateFileCommand = &cobra.Command{
		Use:    "locate-file",
		Short:  "Locate a file using NTT_CACHE",
		Hidden: true,
		RunE:   locate,
	}
)

func locate(cmd *cobra.Command, args []string) error {
	_, files := splitArgs(args, cmd.ArgsLenAtDash())

	for _, path := range files {
		if f := fs.Open(path); f != nil {
			fmt.Println(f.Path())
		}
	}

	return nil
}
