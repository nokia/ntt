package main

import (
	"bufio"
	"os"

	"github.com/spf13/cobra"
)

var (
	T3xfasmCommand = &cobra.Command{
		Use:   "t3xfasm <file>",
		Short: "Assemble text file with t3xf instructions to T3XF binary file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := os.Open(args[0])
			if err != nil {
				return err
			}

			s := bufio.NewScanner(f)
			for s.Scan() {
			}
			if err := s.Err(); err != nil {
				return err
			}

			return nil
		},
	}
)
