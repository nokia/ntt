// Package cli provides helper functions for a consistent CLI. This package is
// very small. The only function implemented is SplitArgs, which can be used to
// split between file names and variables names. See k3-show for example:
//
//         source, keys := cli.SplitArgs(args, cmd.ArgsLenAtDash())
//
// The slice source contains all files required to build a test-suite and the
// optional slice keys contains all explicit variable names.
package cli

// SplitArgs splits an argument list at pos. Pos is usually the position of '--'
// (see cobra.Command.ArgsLenAtDash).
//
// Is pos < 0, the second list will be empty
func SplitArgs(args []string, pos int) ([]string, []string) {
	if pos < 0 {
		return args, []string{}
	}
	return args[:pos], args[pos:]
}
