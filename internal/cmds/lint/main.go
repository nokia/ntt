package lint

import (
	"github.com/nokia/ntt/internal/ntt"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "lint",
		Short: "lint examines TTCN-3 source files and reports suspicious code",
		Long: `lint examines TTCN-3 source files and reports suspicious code.
It may find problems not caught by the compiler, but also constructs
considered "bad style".

Lint's exit code is non-zero for erroneous invocation of the tool or if a
problem was reported.

To list the available checks, run "ntt lint help":

    <none>

For details and flags of a particular check, run "ntt lint help <check>".

For information on writing a new check, see <TBD>.
`,

		RunE: lint,
	}
)

func lint(cmd *cobra.Command, args []string) error {
	suite, err := ntt.NewFromArgs(args...)
	if err != nil {
		return err
	}

	suite.SetErrorHandler(func(e error) { err = e })
	files := suite.Files()

	for i := range suite.Files() {
		info := suite.Parse(files[i])
		if info.Err != nil {
			return err
		}
	}
	return err
}
