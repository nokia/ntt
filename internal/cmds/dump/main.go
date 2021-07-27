package dump

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/nokia/ntt/internal/ntt"
	"github.com/nokia/ntt/project"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:    "dump",
		Hidden: true,
		RunE:   dump,
	}
)

func dump(cmd *cobra.Command, args []string) error {
	suite, err := ntt.NewFromArgs(args...)
	if err != nil {
		return err
	}

	srcs, err := project.Files(suite)
	if err != nil {
		return err
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
			return err
		}
		fmt.Println(string(b))
	}

	return nil
}
