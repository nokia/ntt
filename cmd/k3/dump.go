package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/spf13/cobra"
)

var (
	dumpCmd = &cobra.Command{
		Use:    "dump",
		Hidden: true,
		Run:    dump,
	}
)

func init() {
	rootCmd.AddCommand(dumpCmd)
}

func dump(cmd *cobra.Command, args []string) {
	suite, err := ntt.NewFromArgs(args...)
	if err != nil {
		fatal(err)
	}
	srcs, err := suite.Sources()
	if err != nil {
		fatal(err)
	}

	imps, err := suite.Imports()
	if err != nil {
		fatal(err)
	}

	for i := range imps {
		files, err := ntt.TTCN3Files(imps[i].Path())
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: warning: %s\n", imps[i].Path(), err.Error())
		}
		for _, f := range files {
			srcs = append(srcs, suite.File(f))
		}
	}

	var (
		modules  = make([]*ast.Module, len(srcs))
		errors   = make([]error, len(srcs))
		filesets = make([]*loc.FileSet, len(srcs))
		wg       sync.WaitGroup
	)

	wg.Add(len(srcs))
	for i, src := range srcs {
		go func(i int, src *ntt.File) {
			modules[i], filesets[i], errors[i] = suite.Parse(src)
			wg.Done()
		}(i, src)
	}
	wg.Wait()

	for i := range modules {
		b, err := json.MarshalIndent(modules[i], "", "  ")
		if err != nil {
			fatal(err)
		}
		fmt.Println(string(b))
	}

}
