package locate_file

import (
	"fmt"

	"github.com/nokia/ntt/internal/fs"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:    "locate-file",
		Short:  "locate a file using NTT_CACHE",
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

// splitArgs splits an argument list at pos. Pos is usually the position of '--'
// (see cobra.Command.ArgsLenAtDash).
//
// Is pos < 0, the second list will be empty
func splitArgs(args []string, pos int) ([]string, []string) {
	if pos < 0 {
		return args, []string{}
	}
	return args[:pos], args[pos:]
}
