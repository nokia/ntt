package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/nokia/ntt/internal/cli"
	"github.com/nokia/ntt/internal/env"
	"github.com/nokia/ntt/internal/suite"
	"github.com/spf13/cobra"
)

var (
	showCmd = &cobra.Command{
		Use:   "show [ <file>...] [-- var...]",
		Short: "show configuration variables used by k3",
		RunE:  show,
	}
)

func show(cmd *cobra.Command, args []string) error {

	source, keys := cli.SplitArgs(args, cmd.ArgsLenAtDash())

	s, err := suite.NewFromArgs(source)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	s.SetEnv()

	if len(keys) == 0 {
		for _, k := range env.Keys() {
			if v := env.Get(k); v != "" {
				fmt.Printf("K3_%s=%s\n", strings.ToUpper(k), v)
			}
		}
		return nil
	}

	for _, k := range keys {
		if v := env.Get(k); v != "" {
			fmt.Println(v)
		}
	}
	return nil
}
