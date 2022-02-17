package build

import (
	"github.com/nokia/ntt/internal/ntt"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "build",
		Short: "Builds compiles TTCN-3 source and imports specified by the import paths.",
		RunE:  build,
	}
)

func build(cmd *cobra.Command, args []string) error {

	suite, err := ntt.NewFromArgs(args...)
	if err != nil {
		return err
	}

	srcs, err := suite.Sources()
	if err != nil {
		return err
	}

	imports, err := suite.Imports()
	if err != nil {
		return err
	}

	for _, imp := range imports {
		//mg.Deps(mg.F(importTask, imp))
	}
	return nil
}
