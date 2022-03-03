package build

import (
	"encoding/json"
	"fmt"
	"os"

	ntt2 "github.com/nokia/ntt"
	"github.com/nokia/ntt/internal/compdb"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "build",
		Short: "Build test suite and its dependencies",
		RunE:  Build,
	}

	ErrNoSources = fmt.Errorf("no sources available")

	CompDB bool
)

func init() {
	Command.Flags().BoolVar(&CompDB, "compdb", false, "generate compilation database")
}

func Build(cmd *cobra.Command, args []string) error {
	suite, err := ntt.NewFromArgs(args...)
	if err != nil {
		return err
	}
	name, err := suite.Name()
	if err != nil {
		return err
	}
	if CompDB {
		builders, err := ntt2.PlanProject(name, suite)
		if err != nil {
			return err
		}

		var db []compdb.Command
		for _, b := range builders {
			if c, ok := b.(compdb.Commander); ok {
				db = append(db, c.Commands()...)
			}
		}
		if len(db) > 0 {
			b, err := json.MarshalIndent(db, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal compdb: %w", err)
			}

			if err := os.WriteFile("compile_commands.json", b, 0644); err != nil {
				return fmt.Errorf("failed to write compile_commands.json: %w", err)
			}
		}
	}
	return ntt2.BuildProject(name, suite)
}
