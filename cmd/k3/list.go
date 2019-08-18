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
		Use:   "list",
		Short: "List various types of objects",
		Long: `List various types of objects.

Default output shows the testcase names in current directory.`,

		Run: listTests,
	}

	listTestsCmd = &cobra.Command{
		Use:     `tests`,
		Aliases: []string{"test", "tc", "tcs", "testcase", "testcases"},
		Run:     listTests,
	}

	listModulesCmd = &cobra.Command{
		Use:     `modules`,
		Aliases: []string{"module", "mod", "mods"},
		Run:     listModules,
	}

	listImportsCmd = &cobra.Command{
		Use:     `imports`,
		Aliases: []string{"import", "dep", "deps", "dependency", "dependencies"},
		Run:     listImports,
	}

	w = bufio.NewWriter(os.Stdout)

	Tags = false
)

func init() {
	listCmd.PersistentFlags().BoolVarP(&Tags, "tags", "t", false, "enable output of testcase documentation tags")
	listCmd.AddCommand(listTestsCmd, listModulesCmd, listImportsCmd)
	rootCmd.AddCommand(listCmd)
}

func listTests(cmd *cobra.Command, args []string) {
	suite := loadSuite(args, loader.Config{
		IgnoreImports:  true,
		IgnoreTags:     !Tags,
		IgnoreComments: !Tags,
	})

	for _, test := range suite.Tests() {
		printItem(test.Module().File(), test.FullName(), test.Tags())
	}
	w.Flush()
}

func listModules(cmd *cobra.Command, args []string) {
	suite := loadSuite(args, loader.Config{
		IgnoreImports:  true,
		IgnoreTags:     !Tags,
		IgnoreComments: !Tags,
	})

	for _, mod := range suite.Modules() {
		printItem(mod.File(), mod.Name(), mod.Tags())
	}
	w.Flush()
}

func listImports(cmd *cobra.Command, args []string) {
	suite := loadSuite(args, loader.Config{
		IgnoreTags:     true,
		IgnoreComments: true,
	})

	for _, mod := range suite.Modules() {
		for _, imp := range mod.Imports {
			printItem(imp.Module().File(), mod.Name()+"\t"+imp.ImportedModule(), nil)
		}
	}
	w.Flush()
}

func loadSuite(args []string, conf loader.Config) runtime.Suite {
	if _, err := conf.FromArgs(args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	suite, err := conf.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	return suite
}

func printItem(file string, item string, tags [][]string) {
	file = file + "\t"
	if !Verbose {
		file = ""
	}

	if len(tags) != 0 {
		for _, tag := range tags {
			fmt.Fprintf(w, "%s%s\t%s\t%s\n", file, item, tag[0], tag[1])
		}
	}

	fmt.Fprintf(w, "%s%s\n", file, item)
}
