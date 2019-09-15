package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:           "k3objdump",
		Short:         "k3objdump displays information from T3XF files.",
		Args:          cobra.ExactArgs(1),
		RunE:          displayInfo,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	Debug  bool
	Legacy bool
)

func main() {
	rootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, "verbose debug output")
	rootCmd.PersistentFlags().BoolVarP(&Legacy, "legacy", "", false, "legacy output with stripped line numbers")
	if err := rootCmd.Execute(); err != nil {
		fatal(err)
	}
}

func displayInfo(cmd *cobra.Command, args []string) error {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	p, err := NewPrinter(args[0], w)
	if err != nil {
		return fileError(args[0], err)
	}

	switch {
	case Debug:
		p.printAddrs = true
		p.printRaw = true
		p.printLiteralInstrs = true

	case Legacy:
		p.printLines = true

	default:
		p.Info()
		return nil
	}

	return p.Print()
}

func fileError(file string, err error) error {
	return errors.New(fmt.Sprintf("%s: %s", file, err.Error()))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}
