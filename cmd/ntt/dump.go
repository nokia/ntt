package main

import (
	"encoding/json"
	"fmt"
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

	srcs, err := suite.Files()
	if err != nil {
		fatal(err)
	}

	var (
		asts = make([]*ntt.ParseInfo, len(srcs))
		wg   sync.WaitGroup
	)

	wg.Add(len(srcs))
	for i, src := range srcs {
		go func(i int, src string) {
			asts[i] = suite.Parse(src)
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
