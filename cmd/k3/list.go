package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/nokia/ntt/internal/loader"
	"github.com/nokia/ntt/internal/ttcn3/doc"
	"github.com/nokia/ntt/internal/ttcn3/syntax"

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

	w = bufio.NewWriter(os.Stdout)

	printers = map[string]func(string, *syntax.Module, syntax.Node){
		"tests":     printTests,
		"test":      printTests,
		"tc":        printTests,
		"tcs":       printTests,
		"testcase":  printTests,
		"testcases": printTests,

		"imports":      printImports,
		"import":       printImports,
		"dep":          printImports,
		"deps":         printImports,
		"dependency":   printImports,
		"dependencies": printImports,

		"modules": printModules,
		"module":  printModules,
		"mod":     printModules,
		"mods":    printModules,
	}

	Tags = false
)

func init() {
	listCmd.Flags().BoolVarP(&Tags, "tags", "t", false, "enable output of testcase documentation tags")
	rootCmd.AddCommand(listCmd)
}

func list(cmd *cobra.Command, args []string) error {
	printer := printTests

	if len(args) > 0 {
		if n, ok := printers[args[0]]; ok {
			printer = n
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

	for _, mod := range suite.Modules {
		var file string
		if Verbose {
			file = suite.Fset.File(mod.Pos()).Name()
		}
		printer(file, mod, mod)
	}
	w.Flush()

	return nil
}

func printTests(file string, m *syntax.Module, n syntax.Node) {
	for _, def := range m.Defs {
		switch x := def.Def.(type) {
		case *syntax.GroupDecl:
		case *syntax.FuncDecl:
			if !x.IsTest() {
				break
			}

			if Tags {
				if comments := x.Kind.Comments(); comments != "" {
					tags := doc.FindAllTags(comments)
					if len(tags) > 0 {
						for _, tag := range tags {

							if Verbose {
								fmt.Fprint(w, file, "\t")
							}

							fmt.Fprintf(w, "%s.%s\t%s\t%s\n",
								m.Name.Tok.Lit,
								x.Name.Tok.Lit,
								tag[0],
								tag[1])
						}
						break
					}
				}
			}

			if Verbose {
				fmt.Fprint(w, file, "\t")
			}

			fmt.Fprintln(w, m.Name.Tok.Lit+"."+x.Name.Tok.Lit)
		}
	}
}

func printModules(file string, m *syntax.Module, n syntax.Node) {
	if Verbose {
		fmt.Fprint(w, file, "\t")
	}
	fmt.Fprintln(w, m.Name.Tok.Lit)
}

func printImports(file string, m *syntax.Module, n syntax.Node) {
	for _, def := range m.Defs {
		switch x := def.Def.(type) {
		case *syntax.GroupDecl:
		case *syntax.ImportDecl:
			if Verbose {
				fmt.Fprint(w, file, "\t")
			}
			fmt.Fprintln(w, m.Name.Tok.Lit, "<-", x.Module.Tok.Lit)
		}
	}
}
