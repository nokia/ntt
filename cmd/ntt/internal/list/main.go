package list

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/nokia/ntt/internal/loader"
	"github.com/nokia/ntt/internal/runtime"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "list",
		Short: "List various types of objects",
		Long: `List various types of objects.

List modules, imports or tests. The list command without any explicit
sub-commands will output tests.

List will not output objects from imported directories. If you need to list all
objects from a testsuite you currently have to pass .ttcn3 files as arguments.
Example:

    ntt list $(ntt show -- sources) $(find $(ntt show -- imports) -name \*.ttcn3)


You can use regular expressions to filter objects. If you pass multiple regular
expressions, all of them must match (AND). Example:

	$ cat example.ttcn3
	testcase foo() ...
	testcase bar() ...
	testcase foobar() ...
	...

	$ ntt list --regex=foo --regex=bar
	example.foobar

	$ ntt list --regex='foo|bar'
	example.foo
	example.bar
	example.foobar


Similarly, you can also specify regular expressions for documentation tags.
Example:

	$ cat example.ttcn3
	// @one
	// @two some-value
	testcase foo() ...

	// @two: some-other-value
	testcase bar() ...
	...

	$ ntt list --tags-regex=@one --tags-regex=@two
	example.foo

	$ ntt list --tags-regex='@two: some'
	example.foo
	example.bar

`,

		RunE: listTests,
	}

	listTestsCmd = &cobra.Command{
		Use:     `tests`,
		Aliases: []string{"test", "tc", "tcs", "testcase", "testcases"},
		RunE:    listTests,
	}

	listModulesCmd = &cobra.Command{
		Use:     `modules`,
		Aliases: []string{"module", "mod", "mods"},
		RunE:    listModules,
	}

	listImportsCmd = &cobra.Command{
		Use:     `imports`,
		Aliases: []string{"import", "dep", "deps", "dependency", "dependencies"},
		RunE:    listImports,
	}

	w = bufio.NewWriter(os.Stdout)

	Tags    = false
	verbose = false

	ItemsRegex   = []string{}
	ItemsExclude = []string{}
	TagsRegex    = []string{}
	TagsExclude  = []string{}
)

func init() {
	Command.PersistentFlags().BoolVarP(&Tags, "tags", "t", false, "enable output of testcase documentation tags")
	Command.PersistentFlags().StringSliceVarP(&ItemsRegex, "regex", "r", []string{}, "list objects matching regular * expression.")
	Command.PersistentFlags().StringSliceVarP(&ItemsExclude, "exclude", "x", []string{}, "exclude objects matching regular * expresion.")
	Command.PersistentFlags().StringSliceVarP(&TagsRegex, "tags-regex", "R", []string{}, "list objects with tags matching regular * expression")
	Command.PersistentFlags().StringSliceVarP(&TagsExclude, "tags-exclude", "X", []string{}, "exclude objects with tags matching * regular expression")
	Command.AddCommand(listTestsCmd, listModulesCmd, listImportsCmd)
}

func listTests(cmd *cobra.Command, args []string) error {
	verbose, _ = cmd.Flags().GetBool("verbose")

	suite, err := loadSuite(args, loader.Config{
		IgnoreImports:  true,
		IgnoreTags:     !needTags(),
		IgnoreComments: !needTags(),
	})

	if err != nil {
		return err
	}

	for _, test := range suite.Tests() {
		printItem(test.Module().File(), test.FullName(), test.Tags())
	}
	w.Flush()
	return nil
}

func listModules(cmd *cobra.Command, args []string) error {
	verbose, _ = cmd.Flags().GetBool("verbose")
	suite, err := loadSuite(args, loader.Config{
		IgnoreImports:  true,
		IgnoreTags:     !needTags(),
		IgnoreComments: !needTags(),
	})

	if err != nil {
		return err
	}

	for _, mod := range suite.Modules() {
		printItem(mod.File(), mod.Name(), mod.Tags())
	}
	w.Flush()
	return nil
}

func listImports(cmd *cobra.Command, args []string) error {
	verbose, _ = cmd.Flags().GetBool("verbose")
	suite, err := loadSuite(args, loader.Config{
		IgnoreTags:     !needTags(),
		IgnoreComments: !needTags(),
	})

	if err != nil {
		return err
	}

	for _, mod := range suite.Modules() {
		for _, imp := range mod.Imports {
			printItem(imp.Module().File(), mod.Name()+"\t"+imp.ImportedModule(), nil)
		}
	}
	w.Flush()
	return nil
}

func loadSuite(args []string, conf loader.Config) (runtime.Suite, error) {
	// Update configuration with TTCN-3 source files from args
	if _, err := conf.FromArgs(args); err != nil {
		return nil, err
	}

	// Load suite
	suite, err := conf.Load()
	if err != nil {
		return nil, err
	}

	return suite, nil
}

func printItem(file string, item string, tags [][]string) {

	if !matchAll(ItemsRegex, item) {
		return
	}

	if len(ItemsExclude) > 0 && matchAll(ItemsExclude, item) {
		return
	}

	if len(TagsRegex) > 0 {
		if len(tags) == 0 {
			return
		}
		if !matchAllTags(TagsRegex, tags) {
			return
		}
	}

	if len(TagsExclude) > 0 && matchAllTags(TagsExclude, tags) {
		return
	}

	file = file + "\t"
	if !verbose {
		file = ""
	}

	if Tags && len(tags) != 0 {
		for _, tag := range tags {
			fmt.Fprintf(w, "%s%s\t%s\t%s\n", file, item, tag[0], tag[1])
		}
		return
	}

	fmt.Fprintf(w, "%s%s\n", file, item)
}

func matchAll(regexes []string, s string) bool {
	for _, r := range regexes {
		if ok, _ := regexp.Match(r, []byte(s)); !ok {
			return false
		}
	}
	return true
}

func matchAllTags(regexes []string, tags [][]string) bool {
next:
	for _, r := range regexes {
		f := strings.SplitN(r, ":", 2)
		for i := range f {
			f[i] = strings.TrimSpace(f[i])
		}
		for _, tag := range tags {
			if ok, _ := regexp.Match(f[0], []byte(tag[0])); !ok {
				continue
			}
			if len(f) > 1 {
				if ok, _ := regexp.Match(f[1], []byte(tag[1])); !ok {
					continue
				}
			}
			continue next
		}
		return false
	}
	return true
}

func needTags() bool {
	return Tags || len(ItemsRegex) != 0 || len(ItemsExclude) != 0 || len(TagsRegex) != 0 || len(TagsExclude) != 0
}
