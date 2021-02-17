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
		Short: "show configuration variables used by ntt",
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
			s, err := suite.Getenv("NTT_DATADIR")
			if err != nil {
				fatal(err)
			}
			return s
		},

		"session_id": func(suite *ntt.Suite) string {
			s, err := suite.Getenv("NTT_SESSION_ID")
			if err != nil {
				fatal(err)
			}
			return s
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

	sources, keys := splitArgs(args, cmd.ArgsLenAtDash())

	suite, err := ntt.NewFromArgs(sources...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	switch {
	case ShSetup:
		return printShellScript(suite, keys)
	case len(keys) != 0:
		return printUserKeys(suite, keys)
	default:
		return printDefaultKeys(suite)
	}
}

func printShellScript(suite *ntt.Suite, keys []string) error {
	if len(keys) != 0 {
		return fmt.Errorf("command line option --sh-setup does not accept additional command line arguments")
	}

	for _, key := range []string{
		"name",
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

	fmt.Print(`
# k3-cached-file searches for file $1 in $K3_CACHE and $NTT_CACHE and returns its first
# occurence.
function k3-cached-file()
{
    local IFS=:
    read -r -a dirs <<<"$K3_CACHE:$NTT_CACHE"
    for dir in "${dirs[@]}"; do
        local cached_file="$dir/$1"
        if [ -e "$cached_file" ]; then
            echo "$cached_file"
            return
        fi
    done
    echo "$1"
}

`)

	env, err := suite.Environ()
	if err != nil {
		fatal(err)
	}
	sort.Strings(env)
	for _, e := range env {
		fmt.Printf("export %s\n", e)
	}
	fmt.Println()

	fmt.Println("K3_SOURCES=(")
	srcs, err := suite.Sources()
	for _, src := range ntt.PathSlice(srcs...) {
		fmt.Printf("\t%s\n", src)
	}
	fmt.Print(")\n\n")

	fmt.Println("K3_IMPORTS=(")
	imps, err := suite.Imports()
	for _, imp := range ntt.PathSlice(imps...) {
		fmt.Printf("\t%s\n", imp)
	}
	fmt.Print(")\n\n")

	fmt.Println("K3_TTCN3_FILES=(")
	files, err := suite.Files()
	for _, f := range append(files, ntt.FindAuxiliaryTTCN3Files()...) {
		fmt.Printf("\t%s\n", f)
	}
	fmt.Print(")\n\n")

	return nil
}

func printUserKeys(suite *ntt.Suite, keys []string) error {
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

func printDefaultKeys(suite *ntt.Suite) error {
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

// splitArgs splits an argument list at pos. Pos is usually the position of '--'
// (see cobra.Command.ArgsLenAtDash).
//
// Is pos < 0, the second list will be empty
func splitArgs(args []string, pos int) ([]string, []string) {
	if pos < 0 {
		return args, []string{}
	}
	return args[:pos], args[pos:]
}
