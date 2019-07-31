package main

import (
	"fmt"
	"ntt/internal/loader"
	"os"

	"github.com/spf13/cobra"
)

var (
	listCmd = &cobra.Command{
		Use:   "list [tests|modules|imports] [files...]",
		Short: "List various types of objects",
		Long: `List various types of objects.

Default output shows the testcase names in current directory.`,

		RunE:      list,
		ValidArgs: []string{"tests", "imports", "modules"},
		ArgAliases: []string{
			"test", "tc", "tcs", "testcase", "testcases",
			"import", "dep", "deps", "dependency", "dependencies",
			"mod", "mods", "module",
		},
	}
)

var nouns = map[string]string{
	"tests":   "tests",
	"imports": "imports",
	"modules": "imports",

	"test":      "tests",
	"tc":        "tests",
	"tcs":       "tests",
	"testcase":  "tests",
	"testcases": "tests",

	"import":       "imports",
	"dep":          "imports",
	"deps":         "imports",
	"dependency":   "imports",
	"dependencies": "imports",

	"mod":    "modules",
	"mods":   "modules",
	"module": "modules",
}

func list(cmd *cobra.Command, args []string) error {
	noun := "tests"

	if len(args) > 0 {
		if n, ok := nouns[args[0]]; ok {
			noun = n
			args = args[1:]
		}
	}

	suite, err := loader.NewFromArgs(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if err := suite.Load(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	fmt.Println(noun)
	for _, m := range suite.Modules {
		fmt.Println(m)
	}
	return nil
}
