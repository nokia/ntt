package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/nokia/ntt/internal/ntt"
	"github.com/spf13/cobra"
)

var (
	showCmd = &cobra.Command{
		Use:   "show [ <file>...] [-- var...]",
		Short: "show configuration variables used by k3",
		RunE:  show,
	}

	stringers = map[string]func(suite *ntt.Suite) string{
		"name": func(suite *ntt.Suite) string {
			s, err := suite.Name()
			if err != nil {
				fatal(err)
			}
			return s
		},

		"sources": func(suite *ntt.Suite) string {
			srcs, err := suite.Sources()
			if err != nil {
				fatal(err)
			}
			return strings.Join(ntt.PathSlice(srcs...), "\n")
		},

		"imports": func(suite *ntt.Suite) string {
			imps, err := suite.Imports()
			if err != nil {
				fatal(err)
			}
			return strings.Join(ntt.PathSlice(imps...), "\n")
		},

		"timeout": func(suite *ntt.Suite) string {
			t, err := suite.Timeout()
			if err != nil {
				fatal(err)
			}
			if t > 0 {
				return fmt.Sprint(t)
			}
			return ""

		},

		"parameters_file": func(suite *ntt.Suite) string {
			f, err := suite.ParametersFile()
			if err != nil {
				fatal(err)
			}
			if f != nil {
				return f.Path()
			}
			return ""
		},

		"test_hook": func(suite *ntt.Suite) string {
			f, err := suite.TestHook()
			if err != nil {
				fatal(err)
			}
			if f != nil {
				return f.Path()
			}
			return ""
		},

		"source_dir": func(suite *ntt.Suite) string {
			if root := suite.Root(); root != nil {
				return root.Path()
			}
			return ""
		},

		"datadir": func(suite *ntt.Suite) string {
			return os.Getenv("K3_DATADIR")
		},

		"session_id": func(suite *ntt.Suite) string {
			return os.Getenv("K3_SESSION_ID")
		},

		"env": func(suite *ntt.Suite) string {
			env, err := suite.Environ()
			if err != nil {
				fatal(err)
			}
			sort.Strings(env)
			return strings.Join(env, "\n")
		},
	}
)

func show(cmd *cobra.Command, args []string) error {

	sources, keys := ntt.SplitArgs(args, cmd.ArgsLenAtDash())

	suite, err := ntt.NewFromArgs(sources...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if len(keys) == 0 {
		for _, key := range []string{
			"name",
			"sources",
			"imports",
			"timeout",
			"parameters_file",
			"test_hook",
			"source_dir",
			"datadir",
			"session_id",
		} {
			if s := stringers[key](suite); s != "" {
				fmt.Printf("K3_%s=\"%s\"\n", strings.ToUpper(key), strings.Replace(s, "\n", " ", -1))
			}
		}

		return nil
	}

	for _, key := range keys {
		if fun, found := stringers[key]; found {
			if s := fun(suite); s != "" {
				fmt.Println(s)
			}
			continue
		}

		s, err := suite.Getenv(key)
		if err != nil {
			return err
		}

		fmt.Println(s)
	}
	return nil
}
