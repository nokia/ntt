package main

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Output shell completion code",
	Long: `To load completion run

       . <(ntt completion)

To configure your bash shell to load completions for each session add to your bashrc

        # ~/.bashrc or ~/.profile
        . <(ntt completion)

Note, if bash-completion is not installed on Linux, please install the
'bash-completion' package via your distribution's package manager.

`,
	Run: func(cmd *cobra.Command, args []string) {
		rootCmd.GenBashCompletion(os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
