package run

import (
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/nokia/ntt/interpreter"
	"github.com/nokia/ntt/runtime"
	"github.com/nokia/ntt/ttcn3"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:    "run",
		Short:  "Run tests from a TTCN-3 test suite.",
		RunE:   run,
		Hidden: true,
	}
)

func run(cmd *cobra.Command, args []string) error {

	files, err := parse(args...)
	if err != nil {
		return err
	}

	if err := check(files...); err != nil {
		return err
	}

	env := runtime.NewEnv(nil)
	if err := load(env, files...); err != nil {
		return err
	}

	// TODO(5nord) execute all control parts
	return nil
}

// parse TTCN-3 files and return a list of syntax trees.
func parse(files ...string) ([]*ttcn3.Tree, error) {
	result := make([]*ttcn3.Tree, len(files))
	var wg sync.WaitGroup
	wg.Add(len(files))
	for i, file := range files {
		go func(i int, file string) {
			defer wg.Done()
			result[i] = ttcn3.ParseFile(file)
		}(i, file)
	}
	wg.Wait()

	// collect all syntax errors
	var err error
	for _, file := range result {
		if file.Err != nil {
			err = multierror.Append(err, file.Err)
		}
	}

	return result, err
}

func check(files ...*ttcn3.Tree) error {
	return nil
}

// load TTCN-3 files by executing them.
func load(env *runtime.Env, files ...*ttcn3.Tree) error {
	var err error
	for _, file := range files {
		result := interpreter.Eval(file.Root, env)
		if rerr, ok := result.(*runtime.Error); ok {
			err = multierror.Append(err, rerr)
		}
	}
	return err
}
