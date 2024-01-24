package main

import (
	"github.com/spf13/cobra"
)

var (
	ObjdumpCommand = &cobra.Command{
		Use:   "objdump",
		Short: "Display information from T3XF object",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	//bold  = color.New(color.Bold)
	//faint = color.New(color.Faint)
	//token = color.New(color.FgMagenta)
)
