package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/nokia/ntt/internal/ntt"
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
		asts = make([]*ntt.ParseInfo, len(srcs))
		wg   sync.WaitGroup
	)

	wg.Add(len(srcs))
	for i, src := range srcs {
		go func(i int, src *ntt.File) {
			asts[i] = suite.Parse(src.Path())
			wg.Done()
		}(i, src)
	}
	wg.Wait()

	for i := range asts {
		b, err := json.MarshalIndent(asts[i].Module, "", "  ")
		if err != nil {
			fatal(err)
		}
		fmt.Println(string(b))
	}

}
