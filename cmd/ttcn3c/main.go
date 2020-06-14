package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/plugin"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

var (
	rootCmd = &cobra.Command{
		Use:   "ttcn3c",
		Short: "ttcn3c parses TTCN-3 files and generates output based on the options given",
		RunE:  run,
	}

	format = "t3xf"
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&format, "--generator", "G", "t3xf", "generator to use")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}

func run(cmd *cobra.Command, args []string) error {
	name := fmt.Sprintf("ttcn3c-gen-%s", format)
	generator, err := exec.LookPath(name)
	if err != nil {
		fatal(fmt.Errorf("could not find generator %q", name))
	}

	_, err = ntt.NewFromArgs(args...)
	if err != nil {
		fatal(err)
	}

	req := &plugin.GeneratorRequest{
		Version: &plugin.Version{
			Major: 3,
			Minor: 6,
		},
	}

	out, err := proto.Marshal(req)
	if err != nil {
		fatal(err)
	}

	proc := exec.Command(generator)
	proc.Stdin = bytes.NewBuffer(out)
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr

	if err := proc.Start(); err != nil {
		fatal(err)
	}

	proc.Wait()
	return nil
}
