package run

import (
	"fmt"

	"github.com/nokia/ntt/internal/k3r"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use: "run",
		Run: run,
	}

	testsFile  string
	buildSuite bool
)

func init() {
	Command.Flags().StringVarP(&testsFile, "tests-file", "t", "", "read tests to execute from `file`. Use `-` if you want to read from stdin")
	Command.Flags().BoolVarP(&buildSuite, "build", "", true, "ignored")
}

func run(cmd *cobra.Command, args []string) {
	// Separate test suite files from test cases
	args, tests := args[:cmd.ArgsLenAtDash()], args[cmd.ArgsLenAtDash():]

	suite, err := ntt.NewFromArgs(args...)
	if err != nil {
		fatal(err)
	}

	//if len(tests) == 0 {
	//	tests = suite.Tests()
	//}

	worker, err := k3r.New(suite)
	if err != nil {
		fatal(err)
	}

	for _, tst := range tests {
		verdict := worker.Run(tst)
		fmt.Println(verdict)
	}
}
