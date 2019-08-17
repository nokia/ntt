package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/nokia/ntt/internal/loader"
	"github.com/nokia/ntt/internal/runtime"
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

	printers = map[string]func(runtime.Suite){
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

	conf := loader.Config{
		IgnoreImports: true,
	}

	if _, err := conf.FromArgs(args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	suite, err := conf.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	printer(suite)
	w.Flush()

	return nil
}

func printModules(suite runtime.Suite) {
	for _, mod := range suite.Modules() {
		printItem(mod.File(), mod.Name(), mod.Tags())
	}
}

func printTests(suite runtime.Suite) {
	for _, test := range suite.Tests() {
		printItem(test.Module().File(), test.FullName(), test.Tags())
	}
}

func printImports(suite runtime.Suite) {
	for _, mod := range suite.Modules() {
		for _, imp := range mod.Imports {
			printItem(imp.Module().File(), mod.Name()+"\t"+imp.ImportedModule(), nil)
		}
	}
}

func printItem(file string, item string, tags [][]string) {
	file = file + "\t"
	if !Verbose {
		file = ""
	}

	if Tags && len(tags) != 0 {
		for _, tag := range tags {
			fmt.Fprintf(w, "%s%s\t%s\t%s\n", file, item, tag[0], tag[1])
		}
	}

	fmt.Fprintf(w, "%s%s\n", file, item)
}
