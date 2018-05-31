package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List various types of objects",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("List command executed with ", args)
		},
	}
)
