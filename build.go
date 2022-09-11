package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nokia/ntt/internal/compdb"
	"github.com/nokia/ntt/k3"
	"github.com/nokia/ntt/project"
	"github.com/spf13/cobra"
)

var (
	BuildCommand = &cobra.Command{
		Use:   "build",
		Short: "Build test suite and its dependencies",
		RunE:  Build,
	}

	ErrNoSources = fmt.Errorf("no sources available")

	CompDB bool
)

func init() {
	BuildCommand.Flags().BoolVar(&CompDB, "compdb", false, "generate compilation database")
}

func Build(cmd *cobra.Command, args []string) error {
	tasks, err := project.BuildTasks(Project)
	if err != nil {
		return err
	}
	if CompDB {
		var db []compdb.Command
		for _, t := range tasks {
			if _, ok := t.(*k3.TTCN3Library); ok {
				continue
			}
			cmd := t.String()
			for _, in := range t.Inputs() {
				for _, out := range t.Outputs() {
					db = append(db, compdb.Command{
						Command: cmd,
						File:    in,
						Output:  out,
					})
				}
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
	for _, t := range tasks {
		if err := t.Run(); err != nil {
			return err
		}
	}
	return nil
}
